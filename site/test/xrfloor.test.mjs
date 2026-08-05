// xrfloor.test.mjs — the walkable-surface index, against the real shipped levels.
//
//   node --test site/test/
//
// The two levels here are deliberately opposite: Ultima Underworld's floors are
// exactly 1x1 tiles over a 64x64 map, Need for Speed's are road ribbons over ten
// kilometres. One cell-size rule has to fit both, and the assertions below are
// what says it does.

import test from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import { meshesOf } from './glb.mjs';
import { buildFloor, arcSamples, hitArc, floorUnder, blocked } from '../src/xrfloor.js';

// Heights come back through a Möller-Trumbore solve on Float32 vertices, so an
// exactly-flat quad at y = 0 answers -1.4e-17. Compare heights the way the
// viewer cares about them.
const near = (a, b, tol = 1e-4) => assert.ok(
  a !== null && Math.abs(a - b) <= tol, `expected ~${b}, got ${a}`);

const HERE = dirname(fileURLToPath(import.meta.url));
const PUB = join(HERE, '..', 'public');
const UW = join(PUB, 'ultima-underworld-pc', 'levels', 'level1.glb');
const NFS = join(PUB, 'need-for-speed-3do', 'levels', 'course-al1.glb');

// The two presets' own numbers, so the test fails if the geometry these were
// derived from ever changes under them.
const UW_SCALE = 2.4615;          // metres per tile — a 2 m door (see the plan)
const UW_SPAWN = [31.5, 3.0, -2.5];
const NFS_SPAWN = [-2.902, 0, -96];

let uw, nfs;

test('Underworld: 1x1 floor tiles index at a tile-sized cell', () => {
  uw = buildFloor(meshesOf(UW), { maxSlope: 45 });
  assert.ok(uw.n > 3000, `expected thousands of walkable triangles, got ${uw.n}`);
  // Floors are unit quads: the chosen cell must land in the same order of
  // magnitude, and the index must not blow up.
  assert.ok(uw.cell >= 0.5 && uw.cell <= 8, `cell ${uw.cell} is not tile-scale`);
  assert.ok(uw.items.length <= uw.n * 8, `${uw.items.length} insertions for ${uw.n} tris`);
});

test('Need for Speed: road ribbons index without exploding', () => {
  nfs = buildFloor(meshesOf(NFS), { maxSlope: 45 });
  assert.ok(nfs.n > 10000, `expected a big road, got ${nfs.n}`);
  // The measured trap: at a 1 m cell this course costs 9,704,985 insertions.
  // The rule has to walk away from that on its own.
  assert.ok(nfs.cell >= 4, `cell ${nfs.cell} is too fine for 8 u ribbons`);
  assert.ok(nfs.items.length <= nfs.n * 8, `${nfs.items.length} insertions for ${nfs.n} tris`);
  assert.ok(nfs.nx * nfs.nz <= 4e6, `${nfs.nx}x${nfs.nz} cells`);
});

test('the floor is where the exporter said the spawn was', () => {
  // level.go puts the level-1 spawn on the start-room floor and the camera
  // 0.55 above it; the index has to agree, or the preset's spawn height is a
  // guess rather than a measurement.
  const y = floorUnder(uw, UW_SPAWN[0], UW_SPAWN[1] + 2, UW_SPAWN[2]);
  assert.ok(y !== null, 'no floor under the Underworld spawn');
  assert.ok(Math.abs(y - UW_SPAWN[1]) < 0.01, `floor at ${y}, expected ${UW_SPAWN[1]}`);

  // The NFS grid camera was calibrated against the running game; the road under
  // it is y = 0.
  const r = floorUnder(nfs, NFS_SPAWN[0], NFS_SPAWN[1] + 5, NFS_SPAWN[2]);
  assert.ok(r !== null, 'no road under the Need for Speed grid position');
  assert.ok(Math.abs(r) < 0.5, `road at ${r}, expected ~0`);
});

test('an arc thrown level lands on the floor ahead', () => {
  // From eye height at the spawn, thrown horizontally down the corridor the
  // preset faces (-Z). In CONTENT units, as the module documents: a 7 m/s throw
  // under 9.81 m/s² is 7/k and 9.81/k here.
  const eye = [UW_SPAWN[0], UW_SPAWN[1] + 1.6 / UW_SCALE, UW_SPAWN[2]];
  const pts = arcSamples(eye, [0, 0, -1], 7 / UW_SCALE, 9.81 / UW_SCALE, 2.0, 24);
  const hit = hitArc(uw, pts);
  assert.ok(hit, 'arc found no floor');
  assert.ok(hit.normal[1] > 0.7, `landed on something ${hit.normal} — not a floor`);
  assert.ok(hit.point[2] < eye[2], 'landed behind the thrower');
  // It should be a walk away, not at your feet and not across the map.
  const d = Math.hypot(hit.point[0] - eye[0], hit.point[2] - eye[2]);
  assert.ok(d > 0.3 && d < 20, `landed ${d} tiles away`);
});

test('an arc thrown at the sky finds nothing', () => {
  const eye = [UW_SPAWN[0], UW_SPAWN[1] + 0.65, UW_SPAWN[2]];
  // Straight up, and only briefly: it must not come back down within tMax.
  const pts = arcSamples(eye, [0, 1, 0], 0.5, 0.0, 0.5, 24);
  assert.equal(hitArc(uw, pts), null);
});

// A platform over a longer floor, built by hand: ground y = 0 across x 0..10, a
// platform y = 2 across x 0..4 only. The overhang is the point — an arc dropping
// across it crosses BOTH surfaces, so which one is reported is a statement about
// ordering and not just about geometry. (A fixture where the platform covers the
// whole ground silently passes even if the arc is walked backwards: only one
// segment ever hits anything. That mistake was made here first.)
function terrace() {
  // Wound so the face normal comes out +Y — the same winding the exporters emit
  // for a floor, and the thing the grid filters on.
  const quad = (y, x0, x1) => [
    x0, y, 0, x0, y, 4, x1, y, 4,
    x0, y, 0, x1, y, 4, x1, y, 0,
  ];
  return buildFloor([{ positions: Float32Array.from([...quad(0, 0, 10), ...quad(2, 0, 4)]) }], {});
}

test('stacked floors resolve in arc order, not by depth', () => {
  const g = terrace();
  assert.equal(g.n, 4, 'both quads should be walkable');

  // Plumb lines first: over the platform, and past its edge.
  near(floorUnder(g, 2, 6, 2), 2);
  near(floorUnder(g, 2, 1, 2), 0);   // a probe below the platform must see the ground
  near(floorUnder(g, 6, 6, 2), 0);   // past the overhang there is only ground

  // The arc: thrown so it crosses y = 2 while still OVER the platform and would
  // cross y = 0 shortly after. Both are hits; the platform is the first in arc
  // order and is the correct answer. Walking the segments backwards, or sorting
  // the candidates by depth, answers 0 here.
  const onto = hitArc(g, arcSamples([0.5, 5, 2], [1, 0, 0], 3, 9.81, 1.2, 24));
  assert.ok(onto, 'arc toward the platform found nothing');
  near(onto.point[1], 2);            // landed under the platform it should be on
  assert.ok(onto.point[0] < 4, 'the reported hit is not on the platform footprint');

  // Thrown from past the overhang the same arc must reach the ground — same
  // code, opposite answer, so the test above cannot pass by always saying 2.
  const past = hitArc(g, arcSamples([5, 5, 2], [1, 0, 0], 3, 9.81, 1.2, 24));
  assert.ok(past, 'arc past the platform found nothing');
  near(past.point[1], 0);
});

test('the cell size responds to the target, and respects the cell cap', () => {
  // The doubling loop, exercised directly: a tighter budget must produce a
  // coarser grid, and the cap must bind whatever the triangles want.
  const loose = buildFloor(meshesOf(UW), { targetPerCell: 8 });
  const tight = buildFloor(meshesOf(UW), { targetPerCell: 1 });
  assert.ok(tight.cell >= loose.cell, `tight ${tight.cell} < loose ${loose.cell}`);
  assert.ok(tight.items.length <= tight.n * 1, 'the tight budget was not met');

  const capped = buildFloor(meshesOf(UW), { maxCells: 64 });
  assert.ok(capped.nx * capped.nz <= 64, `${capped.nx}x${capped.nz} exceeds the cap`);
});

test('the real dungeon has floors at more than one height in a column', () => {
  // Not the ordering property (that is the ledge test above) — this is the
  // claim that the property MATTERS for the level we are shipping a preset for.
  // Walk a lattice and count columns whose floor set has more than one member.
  const floorsAt = (x, z) => {
    const ys = [];
    let y = 8;
    for (let i = 0; i < 8; i++) {
      const h = floorUnder(uw, x, y, z);
      if (h === null) break;
      ys.push(h);
      y = h - 0.05;
    }
    return ys;
  };
  let stacked = 0;
  for (let tx = 0; tx < 64; tx += 1) {
    for (let ty = 0; ty < 64; ty += 1) {
      if (floorsAt(tx + 0.5, -(ty + 0.5)).length > 1) stacked++;
    }
  }
  assert.ok(stacked > 0, 'expected at least one stacked column in Underworld level 1');
});

test('the side grid sees the walls the floor grid ignores', () => {
  const walls = buildFloor(meshesOf(UW), { mode: 'side', minSlope: 70 });
  assert.ok(walls.n > 1000, `expected thousands of wall triangles, got ${walls.n}`);
  // Every kept triangle must actually be near-vertical.
  for (let t = 0; t < Math.min(walls.n, 500); t++) {
    assert.ok(Math.abs(walls.normals[t * 3 + 1]) <= Math.cos((70 * Math.PI) / 180) + 1e-6);
  }
  // A line right across the 64-tile map has to cross a wall somewhere; a line
  // from a point to itself cannot.
  assert.equal(blocked(walls, [1, 3.5, -32], [63, 3.5, -32]), true);
  assert.equal(blocked(walls, [31.5, 3.5, -2.5], [31.5, 3.5, -2.5]), false);
});

test('an empty input is answered, not thrown at', () => {
  const g = buildFloor([], {});
  assert.equal(g.n, 0);
  assert.equal(hitArc(g, arcSamples([0, 0, 0], [0, 0, -1], 1, 1, 1, 8)), null);
  assert.equal(floorUnder(g, 0, 0, 0), null);
});
