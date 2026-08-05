// xrpreset.js — the per-level VR preset: what a walk-in session needs that no
// amount of measuring a GLB can tell you.
//
// Retro-X carries no units, and it is right not to: the exporters emit whatever
// the game's own coordinates were, and two of them disagree by a factor of two
// and a half. Need for Speed's GLBs are metres (16.16 fixed point, road segments
// 6 m apart). Ultima Underworld's are TILES. Nothing in either file says which,
// and nothing could — "how tall is a door meant to be" is a question about the
// fiction, not about the data.
//
// So it is written down, by hand, once per level, and this module reads it. The
// files live under site/vr/ rather than in the exported tree because they are
// authored rather than generated: you tune one from inside a headset, reload,
// and look again, without running a Go program.
//
// Imports NOTHING (see xrlayout.js, tileatlas.js, xrfloor.js). Everything here
// is plain data, so the defaults, the overrides and the error messages are all
// testable in node.

export const PRESET_ROOT = 'vr/';

// The defaults are the ones that make an unconfigured level SAFE rather than
// pretty: metre-scale, no fog, a gentle throw. A preset that only sets
// metresPerUnit and spawn is a complete preset.
export const DEFAULTS = {
  metresPerUnit: 1,
  spawn: { pos: [0, 0, 0], dir: [0, 0, -1], absolute: false },
  // sky: true | false | { show, scale, fog }. `scale: "auto"` is the one that
  // matters — see the note in xrworld.js on why a unit-radius horizon is fine on
  // a monitor and a wall in your face in stereo.
  sky: { show: true, scale: 'auto', fog: false },
  background: null,
  fog: null,
  // radial defaults to "on if it flickers": brightness is only reachable through
  // the fragment patch, and a torch whose brightness never moves is fog being
  // animated. warm/cool are the tint at full and at lowest intensity.
  torch: { flicker: 0, gusts: 0, radial: null, warm: '#ffd9a0', cool: '#ff6a1e', seed: 1 },
  floor: { source: 'visible', maxSlope: 45, twoSided: false },
  teleport: {
    speed: 7,          // m/s, thrown from the hand
    maxRange: 25,      // m; beyond this the arc is drawn but will not commit
    maxRise: 1.5,      // m you may arrive ABOVE the floor you are standing on
    markerRadius: 0.35,
    snapTurn: 30,      // degrees per flick; 0 disables
    blockers: true,    // reject a landing with a wall between you and it
    steps: 24,
    flightTime: 2.0,   // s of arc integrated
  },
  props: [],
  sounds: [],
};

// ---- loading -------------------------------------------------------------------

// index() lists which (game, asset) pairs have a preset. ONE fetch, rather than
// probing for each asset: probing spams a 404 per level browsed in the headset,
// and a static host with an SPA fallback answers 200 + HTML, so the "missing"
// case would arrive as a JSON parse error instead of a 404.
export async function loadIndex(root = PRESET_ROOT, fetchFn = fetch) {
  let doc;
  try {
    const res = await fetchFn(`${root}index.json`, { cache: 'no-cache' });
    if (!res.ok) return new Set();
    doc = await res.json();
  } catch {
    return new Set(); // no presets deployed: the VR button simply never appears
  }
  return new Set(doc?.presets || []);
}

export async function loadPreset(gameId, assetId, params, root = PRESET_ROOT, fetchFn = fetch) {
  const path = `${root}${gameId}/${assetId}.json`;
  const res = await fetchFn(path, { cache: 'no-cache' });
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  let raw;
  try {
    raw = await res.json();
  } catch (e) {
    // The headset has no console; this string is the whole diagnosis.
    throw new Error(`${path}: not JSON (${e.message})`);
  }
  return normalise(raw, params, `${gameId}/${assetId}`);
}

// ---- normalising ------------------------------------------------------------------

// normalise fills the defaults, applies the ?vr* overrides and rejects what it
// cannot make sense of. Pure: hand it an object and a URLSearchParams-alike.
export function normalise(raw, params, where = 'preset') {
  if (!raw || typeof raw !== 'object') throw new Error(`${where}: not an object`);
  const p = (k) => {
    const v = params?.get?.(k) ?? params?.[k];
    return v == null || v === '' ? null : v;
  };
  const num = (v, fallback) => {
    if (v == null) return fallback;
    const n = Number(v);
    return Number.isFinite(n) ? n : fallback;
  };
  // override -> file -> default, in that order, and a junk override falls
  // through to the FILE rather than to the default. `num(p('x') ?? raw.y, def)`
  // reads like it does this and does not: a present-but-unparseable override is
  // not null, so the ?? never fires and the file's value is skipped entirely.
  // A mistyped ?vrscale would have silently scaled the Abyss to 1 m per tile.
  const knob = (key, fileVal, def) => {
    const o = num(p(key), null);
    if (o != null) return o;
    const f = num(fileVal, null);
    return f != null ? f : def;
  };

  const cfg = {
    metresPerUnit: knob('vrscale', raw.metresPerUnit, DEFAULTS.metresPerUnit),
    spawn: {
      pos: vec3(raw.spawn?.pos, DEFAULTS.spawn.pos, `${where}: spawn.pos`),
      dir: vec3(raw.spawn?.dir, DEFAULTS.spawn.dir, `${where}: spawn.dir`),
      absolute: !!(raw.spawn?.absolute),
    },
    sky: sky(raw.sky),
    background: raw.background ?? DEFAULTS.background,
    torch: { ...DEFAULTS.torch, ...(raw.torch || {}) },
    floor: { ...DEFAULTS.floor, ...(raw.floor || {}) },
    teleport: { ...DEFAULTS.teleport, ...(raw.teleport || {}) },
    props: Array.isArray(raw.props) ? raw.props.map((o, i) => prop(o, `${where}: props[${i}]`)) : [],
    // Read and checked so an authoring mistake is caught now rather than in the
    // pass that starts playing them. Nothing here makes a sound yet.
    sounds: Array.isArray(raw.sounds) ? raw.sounds.map((o, i) => sound(o, `${where}: sounds[${i}]`)) : [],
  };

  // Fog is the torch: distance to black IS the "vision limited to ten metres"
  // the mode exists to give, and it is the far plane's policy as well. So it is
  // normalised even when absent — `null` means "no fog", which is a choice a
  // daylit racetrack makes and a dungeon does not.
  if (raw.fog) {
    cfg.fog = {
      color: raw.fog.color ?? '#000000',
      near: num(raw.fog.near, 1),
      far: knob('vrfog', raw.fog.far, 40),
    };
    if (!(cfg.fog.far > cfg.fog.near)) {
      throw new Error(`${where}: fog.far (${cfg.fog.far}) must be beyond fog.near (${cfg.fog.near})`);
    }
  } else {
    cfg.fog = p('vrfog') ? { color: '#000000', near: 1, far: num(p('vrfog'), 40) } : null;
  }

  cfg.teleport.speed = num(p('vrspeed'), cfg.teleport.speed);
  cfg.torch.flicker = num(p('vrflicker'), cfg.torch.flicker);
  // "on if it flickers" — resolved after the override, so ?vrflicker= turning a
  // steady lamp into a torch also turns on the thing that can carry it.
  if (cfg.torch.radial == null) cfg.torch.radial = cfg.torch.flicker > 0;
  if (p('vrradial') != null) cfg.torch.radial = p('vrradial') !== '0';
  if (p('vrsky') != null) cfg.sky.show = p('vrsky') !== '0';
  if (p('vrskyscale') != null) cfg.sky.scale = num(p('vrskyscale'), cfg.sky.scale);
  if (p('vrtorch') === '0') { cfg.torch.flicker = 0; cfg.torch.radial = false; }

  if (!(cfg.metresPerUnit > 0)) throw new Error(`${where}: metresPerUnit must be positive, got ${cfg.metresPerUnit}`);
  if (!(cfg.floor.maxSlope > 0 && cfg.floor.maxSlope < 90)) {
    throw new Error(`${where}: floor.maxSlope must be between 0 and 90, got ${cfg.floor.maxSlope}`);
  }
  if (!(cfg.teleport.speed > 0)) throw new Error(`${where}: teleport.speed must be positive`);
  return cfg;
}

// sky accepts the boolean it used to be as well as the object it now is, so a
// preset written before the horizon was ever looked at in stereo still loads.
function sky(v) {
  const d = DEFAULTS.sky;
  if (v == null) return { ...d };
  if (v === false) return { ...d, show: false };
  if (v === true) return { ...d };
  return {
    show: v.show !== false,
    scale: v.scale === 'auto' || v.scale == null ? 'auto' : Number(v.scale),
    fog: v.fog === true,
  };
}

function vec3(v, fallback, where) {
  if (v == null) return [...fallback];
  if (!Array.isArray(v) || v.length !== 3 || !v.every((n) => Number.isFinite(n))) {
    throw new Error(`${where}: expected three numbers, got ${JSON.stringify(v)}`);
  }
  return [v[0], v[1], v[2]];
}

function prop(o, where) {
  if (!o?.object) throw new Error(`${where}: needs an "object" (an object asset id in this game)`);
  return {
    object: o.object,
    pos: vec3(o.pos, [0, 0, 0], `${where}.pos`),
    // rotY is degrees, because every one of these is authored by looking at a
    // car on a road and deciding which way it faces.
    rotY: Number.isFinite(o.rotY) ? o.rotY : 0,
    scale: Number.isFinite(o.scale) ? o.scale : 1,
    variant: o.variant || null,
    anim: o.anim || null,
  };
}

const SOUND_KINDS = new Set(['ambient', 'random', 'onTeleport']);

function sound(o, where) {
  if (!o?.file) throw new Error(`${where}: needs a "file"`);
  const kind = o.kind || 'ambient';
  if (!SOUND_KINDS.has(kind)) {
    throw new Error(`${where}: kind "${kind}" is not one of ${[...SOUND_KINDS].join(', ')}`);
  }
  return {
    file: o.file,
    kind,
    // Positions are in CONTENT units, so they are authored off the same
    // coordinates as spawn — you find a spot by standing on it, not by
    // converting metres.
    pos: o.pos ? vec3(o.pos, [0, 0, 0], `${where}.pos`) : null,
    radius: Number.isFinite(o.radius) ? o.radius : 10,
    gain: Number.isFinite(o.gain) ? o.gain : 1,
    loop: o.loop !== false && kind === 'ambient',
    everyMin: Number.isFinite(o.everyMin) ? o.everyMin : 5,
    everyMax: Number.isFinite(o.everyMax) ? o.everyMax : 20,
  };
}
