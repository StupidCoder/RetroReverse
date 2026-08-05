// xrflame.test.mjs — the torch signal, judged without a headset.
//
// You cannot unit-test "looks like fire". You can test the four properties that
// the two-sine version failed, each of which is why it didn't:
//
//   it is bounded          — a torch never goes out and never blows the exposure
//   it is not periodic     — two sines repeat; a viewer learns the loop
//   it rises faster than it falls — flame flares and fades, it does not swing
//   it is broadband        — energy spread across 0.5-20 Hz, not at two spikes

import test from 'node:test';
import assert from 'node:assert/strict';
import { Flame, mixHex } from '../src/xrflame.js';

// Run a flame at a fixed rate and collect the intensity trace.
function trace(opts, seconds, hz = 90) {
  const f = new Flame(opts);
  const dt = 1 / hz;
  const out = [];
  for (let i = 0; i < seconds * hz; i++) out.push(f.step(dt).intensity);
  return out;
}

test('the light stays inside its depth, and never goes out', () => {
  const v = trace({ flicker: 0.35, gusts: 0.5, seed: 7 }, 120);
  const lo = Math.min(...v), hi = Math.max(...v);
  assert.ok(lo > 0.05, `went (nearly) dark: ${lo}`);
  assert.ok(hi <= 1.0000001, `brighter than unlit: ${hi}`);
  // The gust channel is allowed to dig below the nominal depth, but not by much.
  assert.ok(lo >= 1 - 0.95, `below the hard floor: ${lo}`);
  // And it has to actually move — a flame that sits at 1.0 is not a flame.
  assert.ok(hi - lo > 0.15, `barely moved: ${(hi - lo).toFixed(3)}`);
});

test('flicker 0 is a steady lamp, not a slow one', () => {
  const v = trace({ flicker: 0 }, 5);
  assert.ok(v.every((x) => x === 1), 'an unflickering torch must be exactly steady');
});

// maxAutocorr is the periodicity instrument: a signal that repeats with period
// P correlates with itself at lag P. Lags below `minLag` are skipped because a
// smooth signal is trivially correlated with its recent past — what is being
// asked is whether it comes BACK.
function maxAutocorr(v, minLag, maxLag) {
  const mean = v.reduce((a, b) => a + b, 0) / v.length;
  const x = v.map((a) => a - mean);
  const v0 = x.reduce((s, a) => s + a * a, 0);
  let best = 0, at = 0;
  for (let lag = minLag; lag <= maxLag; lag++) {
    let s = 0;
    for (let i = 0; i + lag < x.length; i++) s += x[i] * x[i + lag];
    const r = Math.abs(s / v0);
    if (r > best) { best = r; at = lag; }
  }
  return { r: best, lag: at };
}

test('it does not repeat — the thing two sines cannot do', () => {
  const hz = 90;
  // The control, first: two sines of different amplitude and frequency, which
  // is what this replaced. If the instrument cannot see the periodicity in
  // THIS, it cannot say anything about the flame either. (An earlier version of
  // this test compared windows for similarity; over thousands of candidate
  // windows a bounded smooth signal always finds a close one, so it flagged the
  // flame and would have flagged anything.)
  const sines = [];
  for (let i = 0; i < 300 * hz; i++) {
    const t = i / hz;
    sines.push(1 - 0.2 * (Math.sin(2 * Math.PI * 7 * t) * 0.6 + Math.sin(2 * Math.PI * 4.27 * t) * 0.4));
  }
  const ctrl = maxAutocorr(sines, 5 * hz, 20 * hz);
  assert.ok(ctrl.r > 0.9, `the control is not detected as periodic (r=${ctrl.r.toFixed(3)})`);

  // And the flame, measured the same way over five minutes.
  const v = trace({ flicker: 0.35, gusts: 0.5, seed: 3 }, 300, hz);
  const got = maxAutocorr(v, 5 * hz, 60 * hz);
  assert.ok(got.r < 0.5,
    `repeats at lag ${(got.lag / hz).toFixed(2)} s (r=${got.r.toFixed(3)})`);
});

test('it rises faster than it falls', () => {
  const v = trace({ flicker: 0.4, gusts: 0, seed: 11 }, 120);
  let up = 0, upN = 0, down = 0, downN = 0;
  for (let i = 1; i < v.length; i++) {
    const d = v[i] - v[i - 1];
    if (d > 0) { up += d; upN++; } else if (d < 0) { down -= d; downN++; }
  }
  const meanUp = up / upN, meanDown = down / downN;
  // The envelope's attack is 25 ms against a 90 ms release: the mean upward
  // step must be clearly the larger. (Symmetric shaping gives a ratio of ~1,
  // which is what a sine would score.)
  assert.ok(meanUp / meanDown > 1.5,
    `rise/fall ratio ${(meanUp / meanDown).toFixed(2)} — not asymmetric enough`);
});

test('the spectrum is broadband and falls with frequency', () => {
  // A coarse DFT over the trace: 1/f means energy decreasing with frequency and
  // present at every band, as opposed to two spikes and silence between them.
  const hz = 90, secs = 60;
  const v = trace({ flicker: 0.4, gusts: 0, seed: 5 }, secs, hz);
  const mean = v.reduce((a, b) => a + b, 0) / v.length;
  const power = (f) => {
    let re = 0, im = 0;
    for (let i = 0; i < v.length; i++) {
      const w = 2 * Math.PI * f * (i / hz);
      re += (v[i] - mean) * Math.cos(w);
      im += (v[i] - mean) * Math.sin(w);
    }
    return (re * re + im * im) / (v.length * v.length);
  };
  const bands = [0.5, 1, 2, 4, 8, 16].map((f) => ({ f, p: power(f) }));
  // Every band carries something: no silent gaps between tones.
  for (const b of bands) assert.ok(b.p > 0, `band ${b.f} Hz is empty`);
  // And the low end dominates the high end, as 1/f requires.
  assert.ok(bands[0].p > bands[5].p * 8,
    `0.5 Hz (${bands[0].p.toExponential(2)}) does not dominate 16 Hz (${bands[5].p.toExponential(2)})`);
  assert.ok(bands[1].p > bands[4].p, '1 Hz should carry more than 8 Hz');
});

test('the same seed is the same flame, a different seed is not', () => {
  const a = trace({ flicker: 0.3, seed: 42 }, 3);
  const b = trace({ flicker: 0.3, seed: 42 }, 3);
  const c = trace({ flicker: 0.3, seed: 43 }, 3);
  assert.deepEqual(a, b, 'not deterministic');
  assert.notDeepEqual(a, c, 'the seed does nothing');
});

test('reach moves less than brightness', () => {
  // The whole diagnosis of the first version: it modulated ONLY reach, so the
  // room breathed and nothing ever got brighter. Reach must now be the junior
  // partner.
  const f = new Flame({ flicker: 0.5, gusts: 0, seed: 9 });
  let bright = 0, reach = 0;
  for (let i = 0; i < 90 * 60; i++) {
    f.step(1 / 90);
    bright = Math.max(bright, 1 - f.intensity);
    reach = Math.max(reach, 1 - f.reach);
  }
  assert.ok(bright > 0.1, 'brightness never moved');
  assert.ok(reach < bright * 0.5, `reach (${reach.toFixed(3)}) rivals brightness (${bright.toFixed(3)})`);
});

test('a long frame does not lurch the light', () => {
  // A level finishing its stream, or a tab coming back, hands over a huge dt.
  const f = new Flame({ flicker: 0.5, seed: 2 });
  for (let i = 0; i < 200; i++) f.step(1 / 90);
  const before = f.intensity;
  f.step(4.0);
  assert.ok(Math.abs(f.intensity - before) <= 0.5,
    `a 4 s frame jumped the light from ${before.toFixed(3)} to ${f.intensity.toFixed(3)}`);
});

test('dimmer is redder', () => {
  assert.equal(mixHex('#ff6a1e', '#ffd9a0', 1), '#ffd9a0');
  assert.equal(mixHex('#ff6a1e', '#ffd9a0', 0), '#ff6a1e');
  const mid = mixHex('#000000', '#ffffff', 0.5);
  assert.equal(mid, '#808080');
  // Out-of-range mixes clamp rather than producing nonsense hex.
  assert.equal(mixHex('#000000', '#ffffff', 2), '#ffffff');
  assert.equal(mixHex('#000000', '#ffffff', -1), '#000000');
});
