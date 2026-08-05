// xrflame.js — what a torch actually does to the light in a room.
//
// The first version of this was two sine waves. It read as a pulse, and the
// reason is worth writing down, because the fix is not "more sines".
//
// WHAT THE MEASUREMENTS SAY. Flame luminance is a 1/f random process — pink
// noise, rolling off about 20 dB per decade — which is the signature of the
// turbulence driving it, not of any oscillator. Flames carry content from
// roughly 0.5 to 20 Hz; in still air most of the energy sits below 4 Hz, and
// moving air pushes it up. There IS a coherent mode — "puffing", the toroidal
// vortex a fire sheds — and it scales as f ∝ D^-0.5, putting a pool a few
// centimetres across at about 10 Hz, which is the size of a torch head. So a
// torch is a broadband 1/f signal with a little more energy near 10 Hz, and any
// two-tone approximation of that will always be heard as two tones.
//
// So: five octaves at 0.7 / 1.6 / 3.7 / 8.5 / 19.6 Hz with amplitudes falling
// as 1/f. That IS a 1/f spectrum, by construction, and it is about twenty
// operations a frame.
//
// THREE THINGS BEYOND THE SPECTRUM, all of which matter more than they sound:
//
//   ASYMMETRY. Fire brightens faster than it fades. A signal whose rise and
//   fall are mirror images reads as mechanical however well-shaped it is, so
//   the noise goes through a one-pole envelope with a fast attack and a slower
//   release.
//
//   GUSTS. A separate, very slow channel that occasionally deepens the whole
//   modulation. Without it the light has a texture but never has a moment.
//
//   COLOUR. Dimmer fire is redder — the same reason a dying ember is. Tying the
//   tint to the intensity is most of what makes it read as combustion rather
//   than as a dimmer.
//
// Imports nothing, so all of that is testable in node without a browser or a
// headset: see site/test/xrflame.test.mjs.

// Octave frequencies in Hz and their 1/f amplitudes.
//
// These are ADVANCED, not sampled at a position, and that is the whole reason:
// a hashed lattice sampled at t*hz repeats as soon as every octave's lattice
// lines up again, and with frequencies that are all multiples of 0.1 Hz that is
// every TEN SECONDS. (The repetition test found it — a window matching one
// exactly twelve periods earlier.) Driving each octave from a seeded PRNG
// instead means there is no lattice to line up: the sequence does not come back
// until the generator does, which is 2^32 draws away.
const OCTAVES = [
  { hz: 0.7, amp: 1.0 },
  { hz: 1.6, amp: 0.44 },
  { hz: 3.7, amp: 0.19 },
  { hz: 8.5, amp: 0.082 },
  { hz: 19.6, amp: 0.036 },
];
const NORM = OCTAVES.reduce((s, o) => s + o.amp, 0);

// How much of the modulation reaches the torch's REACH as opposed to its
// brightness. Low on purpose: a torch flickers, a room does not breathe. The
// first version put 100% here and 0% on brightness, which is exactly why it
// looked like fog being animated.
const REACH_SHARE = 0.3;

const GUST_HZ = 0.15;
const GUST_GATE = 0.62;   // above this the gust channel starts to bite
const ATTACK = 0.02;      // s — flare
const RELEASE = 0.13;     // s — fade

export class Flame {
  // { flicker (0..1 depth), gusts (0..1), seed }
  constructor(opts = {}) {
    this.flicker = opts.flicker ?? 0;
    this.gusts = opts.gusts ?? 0;
    this.seed = opts.seed ?? 1;
    this.t = 0;
    this._rng = mulberry32(this.seed >>> 0);
    // One interpolator per octave, plus one for the gust channel at the end.
    this._ch = [...OCTAVES, { hz: GUST_HZ, amp: 0 }].map(() => ({
      prev: this._rng(), next: this._rng(), phase: 0,
    }));
    this.env = 0;          // the smoothed, asymmetric signal, in [-1, 1]
    this.intensity = 1;    // brightness multiplier
    this.reach = 1;        // fog-distance multiplier
    this.warmth = 1;       // 0 = coolest/dimmest tint, 1 = warmest/brightest
  }

  // step advances the flame and returns itself. dt in seconds.
  step(dt) {
    if (!(dt > 0)) return this;
    // A long frame (a tab coming back, a level streaming) must not teleport the
    // envelope — clamp rather than integrate a hole.
    this.t += Math.min(dt, 0.1);
    if (!this.flicker) {
      this.env = 0;
      this.intensity = this.reach = this.warmth = 1;
      return this;
    }

    // 1/f noise in [-1, 1].
    const step = Math.min(dt, 0.1);
    let n = 0;
    for (let i = 0; i < OCTAVES.length; i++) {
      n += this._advance(i, OCTAVES[i].hz, step) * OCTAVES[i].amp;
    }
    n = (n / NORM) * 2 - 1;

    // Asymmetric one-pole: fire flares quicker than it dies.
    const tau = n > this.env ? ATTACK : RELEASE;
    this.env += (n - this.env) * (1 - Math.exp(-Math.min(dt, 0.1) / tau));

    // The gust channel rides on top, deepening the modulation now and then.
    const g = this._advance(OCTAVES.length, GUST_HZ, step);
    let depth = this.flicker;
    if (this.gusts && g > GUST_GATE) {
      depth *= 1 + this.gusts * ((g - GUST_GATE) / (1 - GUST_GATE));
    }
    depth = Math.min(depth, 0.95); // never fully dark: a torch that goes out is a bug

    // env in [-1,1] -> [0,1], with 1 the brightest the flame gets.
    const lift = (this.env + 1) / 2;
    this.intensity = 1 - depth * (1 - lift);
    this.reach = 1 - depth * REACH_SHARE * (1 - lift);
    this.warmth = lift;
    return this;
  }
}

// _advance moves one octave's interpolator on by dt and returns its current
// value in [0,1]. Each octave holds the value it came from and the value it is
// heading to; when it arrives it draws a new target. Smoothstep between them, so
// the light has no corners.
Flame.prototype._advance = function _advance(i, hz, dt) {
  const c = this._ch[i];
  c.phase += dt * hz;
  // A while, not an if: at 0.15 Hz one long frame cannot outrun a channel, but
  // at 19.6 Hz a 100 ms frame crosses two whole intervals.
  while (c.phase >= 1) {
    c.phase -= 1;
    c.prev = c.next;
    c.next = this._rng();
  }
  const u = c.phase * c.phase * (3 - 2 * c.phase);
  return c.prev * (1 - u) + c.next * u;
};

// mulberry32: a small, fast, seeded PRNG. Seeded so a flame is reproducible —
// same seed and same frame timings give the same light, which is what makes the
// signal testable — while never repeating within any session.
function mulberry32(a) {
  return function rng() {
    a = (a + 0x6D2B79F5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// mixHex blends two "#rrggbb" strings and returns the same form. The torch tint
// is authored as colours because that is how anyone picking them thinks.
export function mixHex(a, b, t) {
  const pa = hex(a), pb = hex(b);
  const k = t < 0 ? 0 : t > 1 ? 1 : t;
  const c = [0, 1, 2].map((i) => Math.round(pa[i] + (pb[i] - pa[i]) * k));
  return `#${c.map((v) => v.toString(16).padStart(2, '0')).join('')}`;
}

function hex(s) {
  const v = parseInt(String(s).replace('#', ''), 16) || 0;
  return [(v >> 16) & 255, (v >> 8) & 255, v & 255];
}
