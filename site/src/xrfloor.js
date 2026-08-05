// xrfloor.js — the walkable-surface index, and the ballistic arc that queries it.
//
// Teleporting needs one answer, twenty times a frame: where does a thrown arc
// first meet a floor? A THREE.Raycaster per arc segment against a 33,000-
// triangle course cannot give it at 90 Hz, so the work is done once — at
// placement — and the per-frame query becomes a walk over a handful of grid
// cells.
//
// Three decisions are load-bearing:
//
//   THE ARC IS TRACED, NOT DROPPED. A plumb line from each sample would answer
//   a different question. Ultima Underworld's level 1 has 85 XZ cells with more
//   than one floor stacked in them (bridges, ledges, the pit rims); the arc
//   passes through some of them and lands on others, and only "first hit in ARC
//   ORDER" tells the two apart.
//
//   THE CELL SIZE COMES FROM THE DATA. Underworld's floor triangles are exactly
//   1x1 tile; Need for Speed's are road ribbons with a median XZ extent of 8.3
//   and a p90 of 81. A 1 m cell costs 4,405 insertions for the dungeon and
//   9,704,985 for the course. So the size is chosen by measuring the triangles
//   rather than by picking a number, and no caller has to know.
//
//   THE STORE IS CSR, NOT AN ARRAY PER CELL. At a few hundred thousand
//   insertions the Map-of-Arrays form is tens of thousands of JS arrays and
//   megabytes of garbage; two counting passes into a pair of typed arrays is
//   about a megabyte and no garbage at all.
//
// This module imports NOTHING — same reason as xrlayout.js and tileatlas.js: it
// is the half that can be tested off-headset, in node, against the real shipped
// GLBs. Callers hand it plain arrays, not three.js objects.
//
// Everything here is in CONTENT units (the GLB's own), never metres. The scene
// scale k changes when the viewer retunes it; a grid built in world units would
// have to be rebuilt each time, and `hold` in xr.js is content units anyway. The
// caller converts the hand ray once per frame instead (dividing gravity and
// speed by k) rather than converting twenty arc samples back.

const EPS = 1e-9;

// ---- building ----------------------------------------------------------------

// buildFloor collects the triangles worth keeping and indexes them.
//
//   meshes  [{ positions, indices|null, matrix|null }]
//           positions: flat xyz triples. matrix: 16 numbers in COLUMN-major
//           order — Object3D.matrixWorld.elements, straight off three.
//   opts    { mode: 'up' | 'side', maxSlope: degrees (up), minSlope: degrees
//             (side), targetPerCell, maxCells }
//
// 'up' keeps surfaces you could stand on; 'side' keeps near-vertical ones, which
// is the same code answering "is there a wall in the way".
export function buildFloor(meshes, opts = {}) {
  const mode = opts.mode === 'side' ? 'side' : 'up';
  // cos of the steepest floor still walkable; for walls, the largest |ny| still
  // counted as vertical.
  const upMin = Math.cos(((opts.maxSlope ?? 45) * Math.PI) / 180);
  const sideMax = Math.cos(((opts.minSlope ?? 70) * Math.PI) / 180);

  const tri = [];      // 9 floats per kept triangle
  const nrm = [];      // 3 floats per kept triangle (unit)
  let minX = Infinity, minZ = Infinity, maxX = -Infinity, maxZ = -Infinity;

  for (const m of meshes || []) {
    const p = m.positions;
    if (!p || !p.length) continue;
    const e = m.matrix || null;
    const idx = m.indices;
    const count = idx ? idx.length : p.length / 3;
    for (let i = 0; i + 2 < count; i += 3) {
      const a = (idx ? idx[i] : i) * 3;
      const b = (idx ? idx[i + 1] : i + 1) * 3;
      const c = (idx ? idx[i + 2] : i + 2) * 3;
      const ax = xf(e, p, a, 0), ay = xf(e, p, a, 1), az = xf(e, p, a, 2);
      const bx = xf(e, p, b, 0), by = xf(e, p, b, 1), bz = xf(e, p, b, 2);
      const cx = xf(e, p, c, 0), cy = xf(e, p, c, 1), cz = xf(e, p, c, 2);

      // The geometric face normal, computed from the positions — which is why
      // it does not matter that not one shipped level GLB carries a NORMAL
      // attribute (they are all KHR_materials_unlit).
      const ux = bx - ax, uy = by - ay, uz = bz - az;
      const vx = cx - ax, vy = cy - ay, vz = cz - az;
      let nx = uy * vz - uz * vy;
      let ny = uz * vx - ux * vz;
      let nz = ux * vy - uy * vx;
      const len = Math.hypot(nx, ny, nz);
      if (len < EPS) continue; // degenerate
      nx /= len; ny /= len; nz /= len;

      if (mode === 'up') {
        if (ny < upMin) continue;
      } else if (Math.abs(ny) > sideMax) continue;

      tri.push(ax, ay, az, bx, by, bz, cx, cy, cz);
      nrm.push(nx, ny, nz);
      minX = Math.min(minX, ax, bx, cx); maxX = Math.max(maxX, ax, bx, cx);
      minZ = Math.min(minZ, az, bz, cz); maxZ = Math.max(maxZ, az, bz, cz);
    }
  }

  const n = nrm.length / 3;
  const tris = Float32Array.from(tri);
  const normals = Float32Array.from(nrm);
  if (!n) {
    return { n: 0, tris, normals, cell: 1, minX: 0, minZ: 0, nx: 0, nz: 0,
      offsets: new Int32Array(1), items: new Int32Array(0), stamp: new Int32Array(0),
      stampAge: 0, mode, note: 'empty' };
  }

  const cell = chooseCell(tris, n, minX, minZ, maxX, maxZ, opts);
  const nx = Math.max(1, Math.floor((maxX - minX) / cell) + 1);
  const nz = Math.max(1, Math.floor((maxZ - minZ) / cell) + 1);

  // Two passes: count per cell, prefix-sum, then fill. `cursor` is the moving
  // write head per cell during the fill and is thrown away after.
  const counts = new Int32Array(nx * nz + 1);
  spanEach(tris, n, cell, minX, minZ, nx, nz, (c) => { counts[c]++; });
  let acc = 0;
  const offsets = new Int32Array(nx * nz + 1);
  for (let i = 0; i < nx * nz; i++) { offsets[i] = acc; acc += counts[i]; }
  offsets[nx * nz] = acc;
  const cursor = offsets.slice(0, nx * nz);
  const items = new Int32Array(acc);
  spanEach(tris, n, cell, minX, minZ, nx, nz, (c, t) => { items[cursor[c]++] = t; });

  return {
    n, tris, normals, cell, minX, minZ, nx, nz, offsets, items,
    // Visit stamps let one segment's cell walk skip a triangle it has already
    // tested — road ribbons span many cells, so without this the wide triangles
    // dominate the query.
    stamp: new Int32Array(n), stampAge: 0, mode,
    note: `${n} tris · cell ${round(cell)} · ${nx}x${nz} cells · ${acc} insertions`,
  };
}

function xf(e, p, o, axis) {
  const x = p[o], y = p[o + 1], z = p[o + 2];
  if (!e) return axis === 0 ? x : axis === 1 ? y : z;
  return e[axis] * x + e[4 + axis] * y + e[8 + axis] * z + e[12 + axis];
}

// chooseCell measures the triangles and picks a size that keeps the index small
// without making the cells useless. Starting from the median XZ extent is what
// makes one rule fit both a dungeon of 1x1 tiles and a 10 km road ribbon.
function chooseCell(tris, n, minX, minZ, maxX, maxZ, opts) {
  const target = opts.targetPerCell ?? 8;
  const maxCells = opts.maxCells ?? 4e6;
  const ext = new Float64Array(n);
  for (let t = 0; t < n; t++) {
    const o = t * 9;
    const dx = Math.max(tris[o], tris[o + 3], tris[o + 6]) - Math.min(tris[o], tris[o + 3], tris[o + 6]);
    const dz = Math.max(tris[o + 2], tris[o + 5], tris[o + 8]) - Math.min(tris[o + 2], tris[o + 5], tris[o + 8]);
    ext[t] = Math.max(dx, dz);
  }
  const sorted = ext.slice().sort();
  const median = sorted[Math.floor(n / 2)] || 1;
  const spanX = Math.max(maxX - minX, EPS), spanZ = Math.max(maxZ - minZ, EPS);

  let cell = Math.max(median * 2, Math.max(spanX, spanZ) / 4096, 1e-4);
  for (let guard = 0; guard < 40; guard++) {
    const nx = Math.floor(spanX / cell) + 1;
    const nz = Math.floor(spanZ / cell) + 1;
    if (nx * nz > maxCells) { cell *= 2; continue; }
    let ins = 0;
    for (let t = 0; t < n; t++) {
      const o = t * 9;
      const x0 = Math.floor((Math.min(tris[o], tris[o + 3], tris[o + 6]) - minX) / cell);
      const x1 = Math.floor((Math.max(tris[o], tris[o + 3], tris[o + 6]) - minX) / cell);
      const z0 = Math.floor((Math.min(tris[o + 2], tris[o + 5], tris[o + 8]) - minZ) / cell);
      const z1 = Math.floor((Math.max(tris[o + 2], tris[o + 5], tris[o + 8]) - minZ) / cell);
      ins += (x1 - x0 + 1) * (z1 - z0 + 1);
      if (ins > target * n) break;
    }
    if (ins <= target * n) return cell;
    cell *= 2;
  }
  return cell;
}

function spanEach(tris, n, cell, minX, minZ, nx, nz, fn) {
  for (let t = 0; t < n; t++) {
    const o = t * 9;
    const x0 = clampI(Math.floor((Math.min(tris[o], tris[o + 3], tris[o + 6]) - minX) / cell), 0, nx - 1);
    const x1 = clampI(Math.floor((Math.max(tris[o], tris[o + 3], tris[o + 6]) - minX) / cell), 0, nx - 1);
    const z0 = clampI(Math.floor((Math.min(tris[o + 2], tris[o + 5], tris[o + 8]) - minZ) / cell), 0, nz - 1);
    const z1 = clampI(Math.floor((Math.max(tris[o + 2], tris[o + 5], tris[o + 8]) - minZ) / cell), 0, nz - 1);
    for (let z = z0; z <= z1; z++) for (let x = x0; x <= x1; x++) fn(z * nx + x, t);
  }
}

const clampI = (v, lo, hi) => (v < lo ? lo : v > hi ? hi : v);
const round = (v) => (Math.abs(v) >= 10 ? v.toFixed(0) : v.toFixed(3));

// ---- the arc -----------------------------------------------------------------

// arcSamples integrates p(t) = o + v·t + ½g·t², writing (steps+1) points.
//
// Called with everything already divided by the scene scale, so the arc is in
// content units like the grid: a dungeon tile is 2.46 m, and an arc thrown at
// 7 m/s under 9.81 m/s² is thrown at 2.84 u/s under 3.99 u/s² there. Doing it
// this way round means one conversion per frame rather than twenty.
export function arcSamples(origin, dir, speed, gravity, tMax, steps, out) {
  const pts = out && out.length >= (steps + 1) * 3 ? out : new Float32Array((steps + 1) * 3);
  const dt = tMax / steps;
  for (let i = 0; i <= steps; i++) {
    const t = i * dt;
    pts[i * 3] = origin[0] + dir[0] * speed * t;
    pts[i * 3 + 1] = origin[1] + dir[1] * speed * t - 0.5 * gravity * t * t;
    pts[i * 3 + 2] = origin[2] + dir[2] * speed * t;
  }
  return pts;
}

// hitArc walks the polyline segment by segment and returns the FIRST surface it
// meets — in arc order, which is the whole reason the arc is traced rather than
// a ray dropped from each sample. Returns null if it meets nothing.
export function hitArc(grid, pts, count) {
  if (!grid || !grid.n) return null;
  const segs = (count ?? pts.length / 3) - 1;
  for (let s = 0; s < segs; s++) {
    const h = hitSegment(grid, pts, s * 3, (s + 1) * 3);
    if (h) { h.seg = s; return h; }
  }
  return null;
}

// hitSegment tests one segment against every triangle in the cells its XZ
// footprint touches, and answers with the nearest hit ALONG the segment (not
// the nearest triangle) so a cell holding two stacked floors resolves correctly.
export function hitSegment(grid, pts, a, b) {
  const { cell, minX, minZ, nx, nz, offsets, items, tris, normals, stamp } = grid;
  const ax = pts[a], ay = pts[a + 1], az = pts[a + 2];
  const bx = pts[b], by = pts[b + 1], bz = pts[b + 2];
  const x0 = clampI(Math.floor((Math.min(ax, bx) - minX) / cell), 0, nx - 1);
  const x1 = clampI(Math.floor((Math.max(ax, bx) - minX) / cell), 0, nx - 1);
  const z0 = clampI(Math.floor((Math.min(az, bz) - minZ) / cell), 0, nz - 1);
  const z1 = clampI(Math.floor((Math.max(az, bz) - minZ) / cell), 0, nz - 1);
  // Outside the index entirely — Math.floor of a NaN or a point far off the map
  // both land here.
  if (!(x1 >= 0 && z1 >= 0)) return null;

  const age = ++grid.stampAge;
  const dx = bx - ax, dy = by - ay, dz = bz - az;
  let best = -1, bestU = Infinity;

  for (let z = z0; z <= z1; z++) {
    for (let x = x0; x <= x1; x++) {
      const c = z * nx + x;
      for (let i = offsets[c]; i < offsets[c + 1]; i++) {
        const t = items[i];
        if (stamp[t] === age) continue;
        stamp[t] = age;
        const u = segTri(ax, ay, az, dx, dy, dz, tris, t * 9);
        if (u >= 0 && u < bestU) { bestU = u; best = t; }
      }
    }
  }
  if (best < 0) return null;
  return {
    point: [ax + dx * bestU, ay + dy * bestU, az + dz * bestU],
    normal: [normals[best * 3], normals[best * 3 + 1], normals[best * 3 + 2]],
    tri: best,
    u: bestU,
  };
}

// blocked answers "is there a wall between these two points" against a grid
// built with mode:'side'. Same walk, no interpolation wanted — just yes or no.
export function blocked(grid, from, to) {
  if (!grid || !grid.n) return false;
  const pts = [from[0], from[1], from[2], to[0], to[1], to[2]];
  return !!hitSegment(grid, pts, 0, 3);
}

// Möller-Trumbore, two-sided: an arc can come at a floor from underneath (off a
// ledge, through a gap) and rejecting that would silently make some ledges
// unreachable. Returns the parameter along the segment in [0,1], or -1.
function segTri(ox, oy, oz, dx, dy, dz, tris, o) {
  const ax = tris[o], ay = tris[o + 1], az = tris[o + 2];
  const e1x = tris[o + 3] - ax, e1y = tris[o + 4] - ay, e1z = tris[o + 5] - az;
  const e2x = tris[o + 6] - ax, e2y = tris[o + 7] - ay, e2z = tris[o + 8] - az;
  const px = dy * e2z - dz * e2y;
  const py = dz * e2x - dx * e2z;
  const pz = dx * e2y - dy * e2x;
  const det = e1x * px + e1y * py + e1z * pz;
  if (Math.abs(det) < EPS) return -1;
  const inv = 1 / det;
  const tx = ox - ax, ty = oy - ay, tz = oz - az;
  const u = (tx * px + ty * py + tz * pz) * inv;
  if (u < 0 || u > 1) return -1;
  const qx = ty * e1z - tz * e1y;
  const qy = tz * e1x - tx * e1z;
  const qz = tx * e1y - ty * e1x;
  const v = (dx * qx + dy * qy + dz * qz) * inv;
  if (v < 0 || u + v > 1) return -1;
  const s = (e2x * qx + e2y * qy + e2z * qz) * inv;
  return s >= 0 && s <= 1 ? s : -1;
}

// floorUnder is the one query the arc does not answer: how high is the ground
// directly below a point. Used to reject a landing that is a storey above where
// you stand (teleport.maxRise) and to sanity-check a spawn while authoring.
export function floorUnder(grid, x, y, z, reach = 1e6) {
  const pts = [x, y, z, x, y - reach, z];
  const h = hitSegment(grid, pts, 0, 3);
  return h ? h.point[1] : null;
}
