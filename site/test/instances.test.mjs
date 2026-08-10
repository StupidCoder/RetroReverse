// instances.test.mjs — the static-instancing batch, against the real shipped
// OutRun level. The numbers here are the point: the batch exists because
// Sunny Beach's 411 scenery placements cost 663 draw calls on a Quest 3's
// main thread, and this file pins both the eligibility rules and the
// collapse those rules buy.
//
//   node --test site/test/

import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import { readGLB } from './glb.mjs';
import { instanceable, groupInstanceable, buildInstances } from '../src/instances.js';

const HERE = dirname(fileURLToPath(import.meta.url));
const GAME = join(HERE, '..', 'public', 'outrun-2006-xbox');

const levelDoc = JSON.parse(readFileSync(join(GAME, 'levels', 'stage-beac.json')));
const objDoc = (id) => JSON.parse(readFileSync(join(GAME, 'objects', id + '.json')));
const primsOf = (id) => {
  const doc = objDoc(id);
  const { json } = readGLB(join(GAME, 'objects', doc.model || id + '.glb'));
  const per = (json.meshes || []).map((m) => m.primitives.length);
  let prims = 0;
  for (const n of json.nodes || []) if (n.mesh != null) prims += per[n.mesh];
  return prims;
};

test('sunny beach: every scenery placement is eligible, the starter is not', () => {
  const pls = levelDoc.placements;
  const flag = pls.find((p) => p.object === 'flagman');
  assert.ok(flag, 'the starter is placed');
  assert.equal(instanceable(flag, objDoc('flagman')), false, 'a skinned, clip-carrying actor must keep his clone');
  const scenery = pls.filter((p) => p.object !== 'flagman');
  for (const pl of scenery) {
    assert.equal(instanceable(pl, objDoc(pl.object)), true, `${pl.object} #${pl.id}`);
  }
});

test('sunny beach: the batch collapses 663 scenery draw calls into a few dozen', () => {
  const records = levelDoc.placements.map((pl) => ({ pl, doc: objDoc(pl.object) }));
  const groups = groupInstanceable(records);
  let before = 0, after = 0, taken = 0;
  for (const [obj, rs] of groups) {
    const prims = primsOf(obj);
    before += prims * rs.length;
    after += prims;
    taken += rs.length;
  }
  // The exported level this was measured on: if a re-export moves these, the
  // budget conversation needs to happen again, not be silently absorbed.
  assert.ok(taken >= 390, `expected the bulk of the 410 scenery placements, got ${taken}`);
  assert.ok(before >= 600, `expected the measured ~663 calls, got ${before}`);
  assert.ok(after <= 40, `the collapsed count must stay a few dozen, got ${after}`);
});

test('eligibility refuses every behaviour a batch cannot express', () => {
  const doc = { type: 'model3d' };
  const base = { object: 'x' };
  assert.equal(instanceable(base, doc), true);
  assert.equal(instanceable({ ...base, onClick: {} }, doc), false);
  assert.equal(instanceable({ ...base, anim: 'idle' }, doc), false);
  assert.equal(instanceable({ ...base, tint: '#f00' }, doc), false);
  assert.equal(instanceable({ ...base, hflip: true }, doc), false);
  assert.equal(instanceable({ ...base, variants: ['a'] }, doc), false);
  assert.equal(instanceable({ ...base, room: 0 }, doc), false);
  assert.equal(instanceable({ ...base, layer: 'env' }, doc), false);
  assert.equal(instanceable(base, { ...doc, skinnedClone: true }), false);
  assert.equal(instanceable(base, { ...doc, animations: [{}] }), false);
  assert.equal(instanceable(base, { type: 'billboard3d' }), false);
});

test('assembly: one InstancedMesh per primitive, per-instance world matrices, clones hidden', () => {
  // A minimal stand-in for the three classes and nodes the assembly touches.
  class FakeGroup {
    constructor() { this.children = []; this.name = ''; }
    add(o) { this.children.push(o); }
    traverse(fn) { fn(this); for (const c of this.children) c.traverse?.(fn) ?? fn(c); }
  }
  class FakeInstanced {
    constructor(geometry, material, count) {
      Object.assign(this, { geometry, material, count, matrices: [], isInstancedMesh: true });
      this.instanceMatrix = { needsUpdate: false };
    }
    setMatrixAt(i, m) { this.matrices[i] = m; }
  }
  const mkNode = (id, matrix) => ({
    visible: true,
    updateMatrixWorld() {},
    traverse(fn) { fn(this); fn(this.mesh); },
    mesh: { isMesh: true, geometry: 'geo-' + id, material: 'mat-' + id, matrixWorld: matrix },
  });
  const records = [0, 1, 2, 3].map((i) => ({
    pl: { object: 'palm' }, doc: { type: 'model3d' },
    node: mkNode('palm', 'world-' + i),
  }));
  const out = buildInstances(records, { Group: FakeGroup, InstancedMesh: FakeInstanced });
  assert.ok(out);
  assert.equal(out.group.children.length, 1);
  const im = out.group.children[0];
  assert.equal(im.count, 4);
  assert.deepEqual(im.matrices, ['world-0', 'world-1', 'world-2', 'world-3']);
  assert.equal(im.instanceMatrix.needsUpdate, true);
  for (const r of records) {
    assert.equal(r.batched, true);
    assert.equal(r.node.visible, false, 'the clone must become a pick proxy');
  }
  assert.deepEqual(out.stats, { taken: 4, meshes: 1, callsBefore: 4 });
});
