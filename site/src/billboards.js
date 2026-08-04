// billboards.js — one mesh, one draw call, and a vertex shader that turns
// four coincident points into a sprite.
//
// A billboard3d placement used to be its own Mesh with its own material and
// its own cloned texture. That is one draw call each — two, in fact, because
// three renders a transparent DoubleSide mesh in a back-face pass and a
// front-face pass unless told otherwise — and Ultima Underworld's level 8
// places 571 of them. On a Quest 3 that scene submitted 1212 draw calls per
// frame and spent 75 ms of CPU doing it, for 17,000 triangles.
//
// The sprites have everything a batch needs in common: they are quads, they
// are all cut out with the same alpha test, and none of the game's 259 sheets
// carries a single partially-transparent texel. So the sheets are packed into
// one atlas and the quads into one geometry.
//
// The geometry is STATIC. All four of a sprite's vertices sit at its origin,
// and each carries the corner it represents; the vertex shader builds the
// camera-facing axes and pushes the vertex out along them. It also picks the
// cell: the view row from the bearing to the eye, the animation column from a
// clock uniform. So a frame costs two uniform writes for the whole cast —
// against 1224 float writes and a vertex-buffer upload per frame when the CPU
// oriented them, on a buffer the driver may still be reading from.
//
// The CPU keeps one copy of the same quad arithmetic, in quadCorners(), for
// picking and for the bounds. The two must agree; the pick tests below are
// what says they do.
//
// The cutout batch draws OPAQUE (depth writes on), which is not a compromise
// but a correction: the per-sprite path had to switch depth writes off and
// lean on three's back-to-front sort, so a distant sprite could paint over a
// near one when the sort disagreed with the depth. A binary cutout in the
// depth buffer cannot.

import { THREE, billboardCell } from './engine3d.js';

const GUTTER = 1; // px between packed sheets, so NEAREST cannot sample a neighbour
const _dir = new THREE.Vector3();
const _ax = new THREE.Vector3();
const _ay = new THREE.Vector3();
const _v = new THREE.Vector3();
const _tri = [new THREE.Vector3(), new THREE.Vector3(), new THREE.Vector3(), new THREE.Vector3()];

export class BillboardBatch {
  // opts: { maxTextureSize } — the GL limit, from renderer.capabilities.
  constructor(opts = {}) {
    this.maxTextureSize = opts.maxTextureSize || 4096;
    this.entries = [];
    this.group = new THREE.Group();
    this.group.name = 'billboard-batch';
    this.buckets = [];
    this.atlas = null;
    this.stats = null;
    this._camLocal = new THREE.Vector3();
  }

  get size() { return this.entries.length; }

  // add registers one placement. node is the placement's transform group (its
  // position and scale are the sprite's); inst is the ObjectLibrary instance,
  // whose node is the Mesh this batch replaces.
  add(node, inst) {
    const doc = inst.doc;
    const map = inst.node.material?.map;
    if (!map?.image?.width) return false; // sheet not decoded — leave it alone
    const [w, h] = doc.size || [1, 1];
    const anim = (doc.animations || [])[0] || { col: 0, framesPerView: 1, fps: 0 };
    this.entries.push({
      node,
      inst,
      doc,
      image: map.image,
      w,
      h,
      bottom: (doc.anchorMode || 'center') === 'bottom',
      yaw: (doc.mode || 'camera') !== 'camera',
      cellW: doc.atlas.cellW,
      cellH: doc.atlas.cellH,
      col0: anim.col || 0,
      views: doc.views || 1,
      per: anim.framesPerView || 1,
      fps: anim.fps || 0,
      heading: doc.heading || 0,
      additive: doc.blend === 'additive',
      colorSpace: map.colorSpace,
      rect: null,  // filled by the pack
      slot: -1,    // vertex base within its bucket
      bucket: null,
      shown: true,
      px: 0, py: 0, pz: 0, // last position written into the buffer
    });
    return true;
  }

  // build packs the sheets, makes the meshes and takes the per-sprite meshes
  // out of the scene. Returns stats, or null if the batch could not be built —
  // in which case nothing has been changed and the caller keeps the old path.
  build() {
    if (!this.entries.length) return null;
    const images = [...new Set(this.entries.map((e) => e.image))];
    const pack = packSheets(images, this.maxTextureSize);
    if (!pack) return null; // does not fit this GPU's texture limit

    const cv = document.createElement('canvas');
    cv.width = pack.w;
    cv.height = pack.h;
    const g2 = cv.getContext('2d');
    g2.imageSmoothingEnabled = false;
    for (const [img, r] of pack.rects) {
      g2.drawImage(img, r.x, r.y);
      // A ring of the sheet's own edge texels. The sheets were sampled with
      // CLAMP_TO_EDGE, so a fetch that lands just outside a cell lying at the
      // sheet's edge used to return that edge texel; packed into an atlas it
      // would return whatever was placed next door — or the empty gutter,
      // which cost sprites the top row of pixels wherever the rounding of the
      // interpolated v fell just outside. Replicating the edge makes the
      // atlas answer what the separate sheet answered.
      const w = img.width, h = img.height;
      g2.drawImage(img, 0, 0, w, 1, r.x, r.y - 1, w, 1);
      g2.drawImage(img, 0, h - 1, w, 1, r.x, r.y + h, w, 1);
      g2.drawImage(img, 0, 0, 1, h, r.x - 1, r.y, 1, h);
      g2.drawImage(img, w - 1, 0, 1, h, r.x + w, r.y, 1, h);
      g2.drawImage(img, 0, 0, 1, 1, r.x - 1, r.y - 1, 1, 1);
      g2.drawImage(img, w - 1, 0, 1, 1, r.x + w, r.y - 1, 1, 1);
      g2.drawImage(img, 0, h - 1, 1, 1, r.x - 1, r.y + h, 1, 1);
      g2.drawImage(img, w - 1, h - 1, 1, 1, r.x + w, r.y + h, 1, 1);
    }
    const tex = new THREE.CanvasTexture(cv);
    // flipY false so a rect's y is measured the way it was packed: from the
    // top of the canvas, which is how the cells were laid out in the sheet.
    tex.flipY = false;
    tex.magFilter = tex.minFilter = THREE.NearestFilter;
    tex.generateMipmaps = false;
    tex.colorSpace = this.entries[0].colorSpace;
    tex.needsUpdate = true;
    this.atlas = tex;
    for (const e of this.entries) e.rect = pack.rects.get(e.image);

    // Two buckets at most: the cutouts, and the additive glows (which must
    // blend, so they cannot share a material with the depth-writing cutouts).
    const parts = [
      { additive: false, list: this.entries.filter((e) => !e.additive) },
      { additive: true, list: this.entries.filter((e) => e.additive) },
    ];
    const aw = pack.w, ah = pack.h;
    for (const p of parts) {
      if (!p.list.length) continue;
      const n = p.list.length;
      const pos = new Float32Array(n * 12);
      const uv = new Float32Array(n * 8);
      const corner = new Float32Array(n * 8);
      const size = new Float32Array(n * 8);
      const cell = new Float32Array(n * 16);
      const anim = new Float32Array(n * 16);
      const mode = new Float32Array(n * 4);
      const idx = new Uint32Array(n * 6);
      for (let i = 0; i < n; i++) {
        const e = p.list[i];
        e.slot = i;
        const v = i * 4;
        idx.set([v, v + 2, v + 1, v + 2, v + 3, v + 1], i * 6);
        // The four corners, as multiples of the sprite's size: x across, y up
        // from the anchor. Anchor 'bottom' puts the origin at the sprite's
        // feet, 'center' at its middle.
        const top = e.bottom ? 1 : 0.5, bot = e.bottom ? 0 : -0.5;
        corner.set([-0.5, top, 0.5, top, -0.5, bot, 0.5, bot], i * 8);
        const sx = e.node.scale.x, sy = e.node.scale.y;
        // The corner UVs are ABSOLUTE, and each is its own division. The
        // shader could have derived the far edge as near + size, but that
        // lands one ULP away from the division, and one ULP over a sprite
        // magnified to 400 screen pixels moves a texel boundary by a whole
        // pixel — a row of the sprite's top edge went missing that way.
        const u0 = (e.rect.x + e.col0 * e.cellW) / aw;
        const u1 = (e.rect.x + (e.col0 + 1) * e.cellW) / aw;
        const v0 = e.rect.y / ah;
        const v1 = (e.rect.y + e.cellH) / ah;
        uv.set([u0, v0, u1, v0, u0, v1, u1, v1], i * 8);
        const du = e.cellW / aw, dv = e.cellH / ah;
        for (let k = 0; k < 4; k++) {
          size[i * 8 + k * 2] = e.w * sx;
          size[i * 8 + k * 2 + 1] = e.h * sy;
          cell.set([du, dv, 0, 0], i * 16 + k * 4);
          anim.set([e.views, e.per, e.fps, e.heading], i * 16 + k * 4);
          mode[i * 4 + k] = e.yaw ? 1 : 0;
        }
        writePos(pos, i, e.node.position);
        e.px = e.node.position.x; e.py = e.node.position.y; e.pz = e.node.position.z;
      }
      const geo = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(pos, 3);
      const sizeAttr = new THREE.BufferAttribute(size, 2);
      geo.setAttribute('position', posAttr);
      geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
      geo.setAttribute('aCorner', new THREE.BufferAttribute(corner, 2));
      geo.setAttribute('aSize', sizeAttr);
      geo.setAttribute('aCell', new THREE.BufferAttribute(cell, 4));
      geo.setAttribute('aAnim', new THREE.BufferAttribute(anim, 4));
      geo.setAttribute('aMode', new THREE.BufferAttribute(mode, 1));
      geo.setIndex(new THREE.BufferAttribute(idx, 1));

      const uniforms = {
        uCamLocal: { value: new THREE.Vector3() },
        uTime: { value: 0 },
      };
      const matOpts = {
        // forceSinglePass for the same reason the per-sprite path sets it: a
        // transparent DoubleSide mesh is otherwise drawn back-faces-then-
        // front-faces, and this batch would have paid that on every sprite at
        // once — the additive bucket was quietly costing two calls, not one.
        map: tex, side: THREE.DoubleSide, forceSinglePass: true, toneMapped: false, alphaTest: 0.5,
      };
      if (p.additive) {
        Object.assign(matOpts, {
          transparent: true, blending: THREE.AdditiveBlending, depthWrite: false, alphaTest: 0,
        });
      }
      const mat = new THREE.MeshBasicMaterial(matOpts);
      // onBeforeCompile rather than a ShaderMaterial: three's own colour-space
      // and alpha-test handling stays in the fragment path untouched, so the
      // batch cannot drift from what the per-sprite materials did.
      mat.onBeforeCompile = (shader) => {
        Object.assign(shader.uniforms, uniforms);
        shader.vertexShader = shader.vertexShader
          .replace('#include <common>', `#include <common>
attribute vec2 aCorner;
attribute vec2 aSize;
attribute vec4 aCell;   // one cell's size in atlas UV (du,dv); zw unused
attribute vec4 aAnim;   // views, framesPerView, fps, heading
attribute float aMode;  // 0 = face the eye, 1 = yaw only
uniform vec3 uCamLocal; // the eye, in this mesh's own space
uniform float uTime;`)
          .replace('void main() {', `void main() {
  vec3 rxDir = uCamLocal - position;
  vec3 rxAx, rxAy;
  if (aMode > 0.5) {
    // Yaw billboard: upright however high the eye is.
    float l = max(length(rxDir.xz), 1e-6);
    rxAx = vec3(rxDir.z / l, 0.0, -rxDir.x / l);
    rxAy = vec3(0.0, 1.0, 0.0);
  } else {
    // The axes Object3D.lookAt would have built: +Z toward the eye, +X across
    // it (up x forward), +Y from those two.
    vec3 f = normalize(rxDir);
    rxAx = normalize(vec3(f.z, 0.0, -f.x));
    rxAy = cross(f, rxAx);
  }
  // Which of the eight views faces the eye, and where the idle cycle is.
  float rxRow = 0.0;
  if (aAnim.x > 1.0) {
    float st = 6.283185307179586 / aAnim.x;
    rxRow = mod(floor((atan(rxDir.x, rxDir.z) - aAnim.w) / st + 0.5), aAnim.x);
  }
  float rxFrame = aAnim.z > 0.0 ? mod(floor(uTime * aAnim.z), aAnim.y) : 0.0;`)
          // uv already holds this corner's exact atlas coordinate for the
          // first cell; the row and frame only step it along.
          .replace('#include <uv_vertex>', `vMapUv = uv + vec2(rxFrame * aCell.x, rxRow * aCell.y);`)
          .replace('#include <begin_vertex>',
            'vec3 transformed = position + rxAx * (aCorner.x * aSize.x) + rxAy * (aCorner.y * aSize.y);');
      };
      // Without a cache key of its own this material would share a compiled
      // program with any other unlit textured material in the scene.
      mat.customProgramCacheKey = () => `rx-billboard-${p.additive ? 'add' : 'cut'}`;

      const mesh = new THREE.Mesh(geo, mat);
      // The batch spans the level: culling it would only ever cull all of it.
      // (It would also cull it WRONGLY — the vertices in the buffer are points
      // at the sprites' origins, and the quads only exist after the shader.)
      mesh.frustumCulled = false;
      mesh.renderOrder = p.additive ? 1 : 0;
      this.group.add(mesh);
      const bucket = { mesh, geo, posAttr, sizeAttr, uniforms, list: p.list, maxSize: 0 };
      for (const e of p.list) { e.bucket = bucket; bucket.maxSize = Math.max(bucket.maxSize, e.w, e.h); }
      this.buckets.push(bucket);
    }
    this._bounds();

    // The per-sprite meshes go: their draw call is what this replaces. The
    // placement groups stay — they carry the transform this reads, the
    // visibility the level's layer/variant toggles write, and the pick record.
    for (const e of this.entries) {
      e.inst.node.removeFromParent();
      e.inst.node.geometry.dispose();
      e.inst.node.material.map?.dispose();
      e.inst.node.material.dispose();
    }
    this.stats = {
      sprites: this.entries.length,
      sheets: images.length,
      atlas: `${pack.w}x${pack.h}`,
      draws: this.buckets.length,
    };
    return this.stats;
  }

  // update hands the shader the frame's two facts. camPos must be in the batch
  // group's own space — the same space the placement nodes are in, which under
  // an AR rig is the scaled scene, not metres.
  //
  // It also carries across anything the level moved or hid since the last
  // frame. Nothing does, in a dungeon; the comparison is what makes that free
  // instead of a buffer upload.
  update(camPos, t) {
    let boundsDirty = false;
    for (const b of this.buckets) {
      b.uniforms.uCamLocal.value.copy(camPos);
      b.uniforms.uTime.value = t;
      let posDirty = false, sizeDirty = false;
      for (const e of b.list) {
        const p = e.node.position;
        if (p.x !== e.px || p.y !== e.py || p.z !== e.pz) {
          e.px = p.x; e.py = p.y; e.pz = p.z;
          writePos(b.posAttr.array, e.slot, p);
          posDirty = true;
          boundsDirty = true;
        }
        if (e.node.visible !== e.shown) {
          // A hidden sprite keeps its vertices where they are and loses its
          // size: the quad collapses to a point AT THE SPRITE, which draws
          // nothing and leaves the bounds alone (a point parked at the origin
          // would drag the box AR fits the diorama to).
          e.shown = e.node.visible;
          const s = e.shown ? 1 : 0;
          const sx = e.node.scale.x * s, sy = e.node.scale.y * s;
          for (let k = 0; k < 4; k++) {
            b.sizeAttr.array[e.slot * 8 + k * 2] = e.w * sx;
            b.sizeAttr.array[e.slot * 8 + k * 2 + 1] = e.h * sy;
          }
          sizeDirty = true;
        }
      }
      if (posDirty) b.posAttr.needsUpdate = true;
      if (sizeDirty) b.sizeAttr.needsUpdate = true;
    }
    if (boundsDirty) this._bounds();
  }

  // _bounds sets the bounds by hand. three cannot compute them: the buffer
  // holds points, and the quads only exist downstream of the vertex shader.
  _bounds() {
    for (const b of this.buckets) {
      const box = new THREE.Box3();
      for (const e of b.list) box.expandByPoint(_v.copy(e.node.position));
      box.expandByScalar(b.maxSize); // the widest a quad can reach from its origin
      b.geo.boundingBox = box;
      b.geo.boundingSphere = box.getBoundingSphere(new THREE.Sphere());
    }
  }

  // pick returns the nearest sprite the ray hits, as { distance, node }, where
  // node is the placement group the caller already knows how to read. The
  // quads live in the vertex shader, so the ray is tested against the same
  // quad built on the CPU — a click is not a hot path.
  pick(raycaster) {
    const cam = raycaster.ray.origin;
    let best = null;
    for (const b of this.buckets) {
      for (const e of b.list) {
        if (!e.node.visible) continue;
        quadCorners(e, cam, _tri);
        const d = rayQuad(raycaster.ray, _tri);
        if (d !== null && (!best || d < best.distance)) best = { distance: d, node: e.node };
      }
    }
    return best;
  }

  dispose() {
    for (const b of this.buckets) {
      b.geo.dispose();
      b.mesh.material.dispose();
    }
    this.atlas?.dispose();
    this.group.removeFromParent();
    this.buckets = [];
    this.entries = [];
  }
}

function writePos(arr, slot, p) {
  const o = slot * 12;
  for (let k = 0; k < 12; k += 3) { arr[o + k] = p.x; arr[o + k + 1] = p.y; arr[o + k + 2] = p.z; }
}

// quadCorners builds a sprite's four world corners the way the vertex shader
// does — TL, TR, BL, BR. This is the CPU's copy of that rule, and it exists so
// a click can be resolved without reading back the GPU.
function quadCorners(e, camPos, out) {
  const p = e.node.position;
  _dir.set(camPos.x - p.x, camPos.y - p.y, camPos.z - p.z);
  if (e.yaw) {
    const l = Math.max(Math.hypot(_dir.x, _dir.z), 1e-6);
    _ax.set(_dir.z / l, 0, -_dir.x / l);
    _ay.set(0, 1, 0);
  } else {
    const f = _dir.clone().normalize();
    _ax.set(f.z, 0, -f.x).normalize();
    _ay.crossVectors(f, _ax);
  }
  const hw = (e.w * e.node.scale.x) / 2, hh = e.h * e.node.scale.y;
  const top = e.bottom ? hh : hh / 2, bot = e.bottom ? 0 : -hh / 2;
  for (let k = 0; k < 4; k++) {
    const cx = (k === 0 || k === 2) ? -hw : hw;
    const cy = (k < 2) ? top : bot;
    out[k].set(p.x + _ax.x * cx + _ay.x * cy, p.y + _ax.y * cx + _ay.y * cy, p.z + _ax.z * cx + _ay.z * cy);
  }
}

// rayQuad returns the distance at which the ray crosses the quad (TL,TR,BL,BR
// as two triangles), or null.
function rayQuad(ray, q) {
  const a = rayTri(ray, q[0], q[2], q[1]);
  const b = rayTri(ray, q[2], q[3], q[1]);
  if (a === null) return b;
  if (b === null) return a;
  return Math.min(a, b);
}

const _e1 = new THREE.Vector3(), _e2 = new THREE.Vector3(), _pv = new THREE.Vector3();
const _tv = new THREE.Vector3(), _qv = new THREE.Vector3();

function rayTri(ray, a, b, c) {
  // Möller–Trumbore, two-sided: a billboard has no back.
  _e1.subVectors(b, a);
  _e2.subVectors(c, a);
  _pv.crossVectors(ray.direction, _e2);
  const det = _e1.dot(_pv);
  if (Math.abs(det) < 1e-12) return null;
  const inv = 1 / det;
  _tv.subVectors(ray.origin, a);
  const u = _tv.dot(_pv) * inv;
  if (u < 0 || u > 1) return null;
  _qv.crossVectors(_tv, _e1);
  const v = ray.direction.dot(_qv) * inv;
  if (v < 0 || u + v > 1) return null;
  const t = _e2.dot(_qv) * inv;
  return t > 0 ? t : null;
}

// packSheets shelf-packs the sheets into one atlas, tallest first. Sprite
// sheets are few and similar in height, so a shelf is within a few per cent of
// optimal here and costs nothing to reason about. Returns null if the result
// would exceed the GPU's texture limit.
function packSheets(images, maxSize) {
  const sorted = [...images].sort((a, b) => b.height - a.height || b.width - a.width);
  const pad = 2 * GUTTER; // the clamp ring, on both sides of each axis
  const widest = sorted.reduce((m, i) => Math.max(m, i.width + pad), 1);
  const area = sorted.reduce((s, i) => s + (i.width + pad) * (i.height + pad), 0);
  let w = Math.max(pow2(widest), pow2(Math.ceil(Math.sqrt(area))));
  for (; w <= maxSize; w *= 2) {
    const rects = shelf(sorted, w);
    if (rects && pow2(rects.h) <= maxSize) return { w, h: pow2(rects.h), rects: rects.map };
  }
  return null;
}

// shelf places each sheet with a GUTTER-wide margin on EVERY side — the ring
// build() fills with the sheet's own edge texels. The rect it records is where
// the image itself goes, inside that margin.
function shelf(images, w) {
  const map = new Map();
  const pad = 2 * GUTTER;
  let x = 0, y = 0, shelfH = 0;
  for (const im of images) {
    if (im.width + pad > w) return null;
    if (x + im.width + pad > w) { x = 0; y += shelfH; shelfH = 0; }
    map.set(im, { x: x + GUTTER, y: y + GUTTER });
    x += im.width + pad;
    shelfH = Math.max(shelfH, im.height + pad);
  }
  return { map, h: y + shelfH };
}

function pow2(n) {
  let p = 1;
  while (p < n) p *= 2;
  return p;
}

// billboardCell is imported so the CPU and the shader cannot disagree about
// which cell a sprite shows — it is the reference the shader's mod/floor pair
// reproduces, and the sprite tests compare the two paths' pixels directly.
export { billboardCell };
