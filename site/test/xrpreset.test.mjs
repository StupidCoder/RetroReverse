// xrpreset.test.mjs — the VR presets: defaults, overrides, and the errors.
//
//   node --test-reporter=tap --test 'site/test/*.test.mjs'
//
// The error messages matter as much as the parsing does. There is no console in
// a headset — the status strip on the menu panel is the whole diagnostic surface
// — so a preset with a typo has to say which file, which field and what it got.

import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import { normalise, DEFAULTS } from '../src/xrpreset.js';

const HERE = dirname(fileURLToPath(import.meta.url));
const VR = join(HERE, '..', 'vr');
const read = (p) => JSON.parse(readFileSync(join(VR, p), 'utf8'));
const params = (o) => new URLSearchParams(o);

test('the shipped presets load, and say what the plan said they would', () => {
  const uw = normalise(read('ultima-underworld-pc/level-1.json'), params({}), 'uw');
  // 2 m doors. The derivation is in the preset's own comment; this is the guard
  // that stops it drifting away from the number it was derived for.
  assert.ok(Math.abs(uw.metresPerUnit - 2.4615) < 1e-4);
  const doorTiles = 208 / 256;
  assert.ok(Math.abs(doorTiles * uw.metresPerUnit - 2.0) < 0.005, 'the door is no longer 2 m');
  // The start room: ceiling 4.0 tiles, floor 3.0, so a tile of head-room.
  assert.ok(Math.abs((4.0 - uw.spawn.pos[1]) * uw.metresPerUnit - 2.4615) < 0.01);
  assert.equal(uw.sky, false);
  assert.equal(uw.fog.far, 11);

  const nfs = normalise(read('need-for-speed-3do/cy1.json'), params({}), 'nfs');
  assert.equal(nfs.metresPerUnit, 1);
  assert.equal(nfs.sky, true);
  assert.equal(nfs.props.length, 2);
  // Lane centres are 5*lane - 17.5; the player stands in the -2.5 lane.
  for (const p of nfs.props) assert.ok(Math.abs(((p.pos[0] + 17.5) / 5) % 1) < 1e-6, `${p.object} is not in a lane`);
  assert.ok(nfs.props.every((p) => p.pos[0] !== nfs.spawn.pos[0]), 'a car is parked on the player');
});

test('every preset named in the index exists and parses', () => {
  const idx = read('index.json');
  assert.ok(idx.presets.length > 0);
  for (const id of idx.presets) {
    const cfg = normalise(read(`${id}.json`), params({}), id);
    assert.ok(cfg.metresPerUnit > 0, `${id}: bad scale`);
  }
});

test('a preset with only a scale and a spawn is a complete preset', () => {
  const cfg = normalise({ metresPerUnit: 2, spawn: { pos: [1, 2, 3] } }, params({}));
  assert.equal(cfg.metresPerUnit, 2);
  assert.deepEqual(cfg.spawn.pos, [1, 2, 3]);
  assert.deepEqual(cfg.spawn.dir, DEFAULTS.spawn.dir);
  assert.equal(cfg.fog, null);            // no fog is a choice, not an omission
  assert.equal(cfg.teleport.speed, DEFAULTS.teleport.speed);
  assert.deepEqual(cfg.props, []);
});

test('the ?vr knobs override the file, because a headset has no console', () => {
  const raw = read('ultima-underworld-pc/level-1.json');
  const cfg = normalise(raw, params({ vrscale: '3', vrfog: '25', vrspeed: '12', vrflicker: '0.5' }));
  assert.equal(cfg.metresPerUnit, 3);
  assert.equal(cfg.fog.far, 25);
  assert.equal(cfg.teleport.speed, 12);
  assert.equal(cfg.torch.flicker, 0.5);

  // vrtorch=0 is the "is the darkness the problem?" switch.
  const off = normalise(raw, params({ vrtorch: '0' }));
  assert.equal(off.torch.flicker, 0);
  assert.equal(off.torch.radial, false);

  // A knob with rubbish in it must not silently become NaN and scale the world
  // to nothing — it falls back to the file.
  const junk = normalise(raw, params({ vrscale: 'big' }));
  assert.equal(junk.metresPerUnit, raw.metresPerUnit);
});

test('what a mistyped preset says', () => {
  const bad = (raw, re) => assert.throws(() => normalise(raw, params({}), 'lvl'), re);
  bad(null, /not an object/);
  bad({ metresPerUnit: 0 }, /metresPerUnit must be positive/);
  bad({ metresPerUnit: -2 }, /metresPerUnit must be positive/);
  bad({ spawn: { pos: [1, 2] } }, /spawn\.pos: expected three numbers/);
  bad({ spawn: { pos: 'origin' } }, /spawn\.pos: expected three numbers/);
  bad({ fog: { near: 10, far: 5 } }, /fog\.far \(5\) must be beyond fog\.near \(10\)/);
  bad({ floor: { maxSlope: 120 } }, /maxSlope must be between 0 and 90/);
  bad({ props: [{ pos: [0, 0, 0] }] }, /props\[0\]: needs an "object"/);
  bad({ sounds: [{ file: 'a.mp3', kind: 'shout' }] }, /kind "shout" is not one of/);
  bad({ sounds: [{ kind: 'ambient' }] }, /sounds\[0\]: needs a "file"/);
  // Every message names the file and the field.
  assert.throws(() => normalise({ spawn: { pos: [1] } }, params({}), 'uw/level-1'), /^Error: uw\/level-1: spawn\.pos/);
});

test('the sounds block is validated now and played later', () => {
  // The slot is specified so a preset can be authored against it before the
  // playback pass exists; what it must NOT do is accept nonsense quietly.
  const cfg = normalise({
    metresPerUnit: 1,
    sounds: [
      { file: 'drip.mp3', kind: 'random', pos: [30, 3, -6], radius: 8, everyMin: 6, everyMax: 25, gain: 0.6 },
      { file: 'hum.mp3', kind: 'ambient', gain: 0.3 },
      { file: 'steps.mp3', kind: 'onTeleport' },
    ],
  }, params({}));
  assert.equal(cfg.sounds.length, 3);
  assert.deepEqual(cfg.sounds[0].pos, [30, 3, -6]);
  assert.equal(cfg.sounds[1].loop, true, 'an ambient bed loops by default');
  assert.equal(cfg.sounds[2].loop, false, 'a one-shot does not');
  assert.equal(cfg.sounds[2].gain, 1);
});
