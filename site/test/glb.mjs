// glb.mjs — a minimal GLB reader, for tests only.
//
// xrfloor.js takes plain arrays precisely so it can be tested against the REAL
// shipped levels rather than a hand-made fixture, and this is the twenty lines
// that get them out of a .glb. It understands exactly what the Studio's own
// exporters emit: one buffer, no external URIs, no sparse accessors.

import { readFileSync } from 'node:fs';

const COMPONENT = {
  5120: Int8Array, 5121: Uint8Array, 5122: Int16Array,
  5123: Uint16Array, 5125: Uint32Array, 5126: Float32Array,
};
const NUM = { SCALAR: 1, VEC2: 2, VEC3: 3, VEC4: 4, MAT4: 16 };

export function readGLB(path) {
  const buf = readFileSync(path);
  if (buf.toString('utf8', 0, 4) !== 'glTF') throw new Error(`${path}: not a GLB`);
  let off = 12;
  let json = null, bin = null;
  while (off < buf.length) {
    const len = buf.readUInt32LE(off);
    const type = buf.readUInt32LE(off + 4);
    const body = buf.subarray(off + 8, off + 8 + len);
    if (type === 0x4e4f534a) json = JSON.parse(body.toString('utf8'));
    else if (type === 0x004e4942) bin = body;
    off += 8 + len + ((4 - ((off + 8 + len) % 4)) % 4);
  }
  return { json, bin };
}

function accessor(g, bin, i) {
  const acc = g.accessors[i];
  const Ctor = COMPONENT[acc.componentType];
  const n = NUM[acc.type];
  const bv = g.bufferViews[acc.bufferView];
  const base = (bv.byteOffset || 0) + (acc.byteOffset || 0);
  const stride = bv.byteStride || 0;
  // Interleaved accessors need a copy; tightly packed ones can be viewed, but a
  // copy is a rounding error at these sizes and the code stays one path.
  const out = new Ctor(acc.count * n);
  const view = new DataView(bin.buffer, bin.byteOffset);
  const size = Ctor.BYTES_PER_ELEMENT;
  const get = Ctor === Float32Array ? 'getFloat32'
    : Ctor === Uint32Array ? 'getUint32'
      : Ctor === Uint16Array ? 'getUint16'
        : Ctor === Int16Array ? 'getInt16'
          : Ctor === Int8Array ? 'getInt8' : 'getUint8';
  for (let e = 0; e < acc.count; e++) {
    const at = base + (stride ? e * stride : e * n * size);
    for (let c = 0; c < n; c++) out[e * n + c] = view[get](at + c * size, true);
  }
  return out;
}

// mul returns a*b for two column-major 16-arrays (three's Matrix4 convention).
function mul(a, b) {
  const o = new Float64Array(16);
  for (let c = 0; c < 4; c++) {
    for (let r = 0; r < 4; r++) {
      o[c * 4 + r] = a[r] * b[c * 4] + a[4 + r] * b[c * 4 + 1]
        + a[8 + r] * b[c * 4 + 2] + a[12 + r] * b[c * 4 + 3];
    }
  }
  return o;
}

const IDENTITY = new Float64Array([1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1]);

function nodeMatrix(node) {
  if (node.matrix) return Float64Array.from(node.matrix);
  const [x, y, z, w] = node.rotation || [0, 0, 0, 1];
  const [sx, sy, sz] = node.scale || [1, 1, 1];
  const [tx, ty, tz] = node.translation || [0, 0, 0];
  const x2 = x + x, y2 = y + y, z2 = z + z;
  const xx = x * x2, xy = x * y2, xz = x * z2;
  const yy = y * y2, yz = y * z2, zz = z * z2;
  const wx = w * x2, wy = w * y2, wz = w * z2;
  return Float64Array.from([
    (1 - (yy + zz)) * sx, (xy + wz) * sx, (xz - wy) * sx, 0,
    (xy - wz) * sy, (1 - (xx + zz)) * sy, (yz + wx) * sy, 0,
    (xz + wy) * sz, (yz - wx) * sz, (1 - (xx + yy)) * sz, 0,
    tx, ty, tz, 1,
  ]);
}

// meshesOf flattens a GLB into what buildFloor wants: one entry per primitive,
// with the node's world matrix already resolved.
export function meshesOf(path, sceneIndex = 0) {
  const { json: g, bin } = readGLB(path);
  const out = [];
  const walk = (ni, parent) => {
    const node = g.nodes[ni];
    const world = mul(parent, nodeMatrix(node));
    if (node.mesh != null) {
      for (const prim of g.meshes[node.mesh].primitives) {
        if (prim.attributes.POSITION == null) continue;
        out.push({
          positions: accessor(g, bin, prim.attributes.POSITION),
          indices: prim.indices != null ? accessor(g, bin, prim.indices) : null,
          matrix: world,
          material: prim.material,
        });
      }
    }
    for (const c of node.children || []) walk(c, world);
  };
  for (const ni of g.scenes[sceneIndex].nodes) walk(ni, IDENTITY);
  return out;
}
