// billboards.js — one mesh for all of a level's camera-facing sprites.
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
// one atlas, the quads into one geometry, and the whole cast is drawn in one
// call. Facing the camera is then a matter of writing four positions per
// sprite per frame — the same arithmetic Object3D.lookAt did, minus the
// matrix, the render-list entry and the state change.
//
// The cutout batch draws OPAQUE (depth writes on), which is not a compromise
// but a correction: the per-sprite path had to switch depth writes off and
// lean on three's back-to-front sort, so a distant sprite could paint over a
// near one when the sort disagreed with the depth. A binary cutout in the
// depth buffer cannot.

import { THREE, billboardCell } from './engine3d.js';

const GUTTER = 1; // px between packed sheets, so NEAREST cannot sample a neighbour

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
      additive: doc.blend === 'additive',
      colorSpace: map.colorSpace,
      cell: -1,
      rect: null, // filled by the pack
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
    for (const [img, r] of pack.rects) g2.drawImage(img, r.x, r.y);
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
    for (const p of parts) {
      if (!p.list.length) continue;
      const n = p.list.length;
      const pos = new THREE.BufferAttribute(new Float32Array(n * 12), 3);
      const uv = new THREE.BufferAttribute(new Float32Array(n * 8), 2);
      pos.setUsage(THREE.DynamicDrawUsage);
      uv.setUsage(THREE.DynamicDrawUsage);
      const idx = new Uint32Array(n * 6);
      for (let i = 0; i < n; i++) {
        const v = i * 4;
        idx.set([v, v + 2, v + 1, v + 2, v + 3, v + 1], i * 6);
      }
      const geo = new THREE.BufferGeometry();
      geo.setAttribute('position', pos);
      geo.setAttribute('uv', uv);
      geo.setIndex(new THREE.BufferAttribute(idx, 1));
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
      const mesh = new THREE.Mesh(geo, new THREE.MeshBasicMaterial(matOpts));
      // The batch spans the level: culling it would only ever cull all of it.
      mesh.frustumCulled = false;
      mesh.renderOrder = p.additive ? 1 : 0;
      this.group.add(mesh);
      this.buckets.push({ mesh, geo, pos, uv, list: p.list });
    }

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

  // update rewrites every sprite's four corners to face camPos. camPos must be
  // in the batch group's own space — the same space the placement nodes are
  // in, which under an AR rig is the scaled scene, not metres.
  update(camPos, t) {
    for (const b of this.buckets) {
      const pa = b.pos.array, ua = b.uv.array;
      let uvDirty = false;
      for (let i = 0; i < b.list.length; i++) {
        const e = b.list[i];
        const p = i * 12, q = i * 8;
        const node = e.node;
        const ox = node.position.x, oy = node.position.y, oz = node.position.z;
        if (!node.visible) {
          // Degenerate: four coincident points rasterise nothing, and cost
          // nothing beyond the vertex fetch. Cheaper than re-indexing. They
          // collapse ONTO THE SPRITE, not onto the origin — a hidden sprite
          // parked at (0,0,0) would still be inside the geometry's bounding
          // box, and AR fits the diorama to that box.
          for (let k = 0; k < 12; k += 3) { pa[p + k] = ox; pa[p + k + 1] = oy; pa[p + k + 2] = oz; }
          continue;
        }
        const dx = camPos.x - ox, dz = camPos.z - oz;
        const sx = node.scale.x, sy = node.scale.y;

        // The axes Object3D.lookAt would have built: +Z toward the eye, +X
        // across it, +Y from those two. For a yaw billboard the eye's height
        // is ignored, so the quad stays upright.
        let ax, ay, az, bx, by, bz;
        if (e.yaw) {
          const len = Math.hypot(dx, dz) || 1;
          ax = dz / len; ay = 0; az = -dx / len;
          bx = 0; by = 1; bz = 0;
        } else {
          const dy = camPos.y - oy;
          let fx = dx, fy = dy, fz = dz;
          const fl = Math.hypot(fx, fy, fz) || 1;
          fx /= fl; fy /= fl; fz /= fl;
          // x = up × f, with up = +Y: (1*fz - 0*fy, 0*fx - 0*fz, 0*fy - 1*fx)
          ax = fz; ay = 0; az = -fx;
          const al = Math.hypot(ax, az);
          if (al < 1e-6) { ax = 1; az = 0; } else { ax /= al; az /= al; }
          // y = f × x
          bx = fy * az - fz * ay; by = fz * ax - fx * az; bz = fx * ay - fy * ax;
        }

        const hw = (e.w * sx) / 2, hh = e.h * sy;
        const y0 = e.bottom ? 0 : -hh / 2;   // base, in the quad's own up axis
        const y1 = e.bottom ? hh : hh / 2;   // top
        const rx = ax * hw, ry = ay * hw, rz = az * hw;
        const l0x = ox + bx * y0, l0y = oy + by * y0, l0z = oz + bz * y0;
        const l1x = ox + bx * y1, l1y = oy + by * y1, l1z = oz + bz * y1;
        pa[p] = l1x - rx; pa[p + 1] = l1y - ry; pa[p + 2] = l1z - rz;       // top-left
        pa[p + 3] = l1x + rx; pa[p + 4] = l1y + ry; pa[p + 5] = l1z + rz;   // top-right
        pa[p + 6] = l0x - rx; pa[p + 7] = l0y - ry; pa[p + 8] = l0z - rz;   // bottom-left
        pa[p + 9] = l0x + rx; pa[p + 10] = l0y + ry; pa[p + 11] = l0z + rz; // bottom-right

        const { row, col } = billboardCell(e.doc, dx, dz, t);
        const key = row * 4096 + col;
        if (key !== e.cell) {
          e.cell = key;
          const r = e.rect, aw = this.atlas.image.width, ah = this.atlas.image.height;
          const u0 = (r.x + col * e.cellW) / aw, u1 = (r.x + (col + 1) * e.cellW) / aw;
          const v0 = (r.y + row * e.cellH) / ah, v1 = (r.y + (row + 1) * e.cellH) / ah;
          ua[q] = u0; ua[q + 1] = v0;
          ua[q + 2] = u1; ua[q + 3] = v0;
          ua[q + 4] = u0; ua[q + 5] = v1;
          ua[q + 6] = u1; ua[q + 7] = v1;
          uvDirty = true;
        }
      }
      b.pos.needsUpdate = true;
      if (uvDirty) b.uv.needsUpdate = true;
    }
  }

  // pick returns the nearest sprite the ray hits, as { distance, node }, where
  // node is the placement group the caller already knows how to read. The
  // quads move every frame, so the bounding sphere is rebuilt per query rather
  // than per frame — a click is not a hot path.
  pick(raycaster) {
    let best = null;
    for (const b of this.buckets) {
      b.geo.computeBoundingSphere();
      const hits = [];
      b.mesh.raycast(raycaster, hits);
      for (const h of hits) {
        const i = h.faceIndex >> 1; // two triangles per sprite
        const e = b.list[i];
        if (!e || !e.node.visible) continue;
        if (!best || h.distance < best.distance) best = { distance: h.distance, node: e.node };
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

// packSheets shelf-packs the sheets into one atlas, tallest first. Sprite
// sheets are few and similar in height, so a shelf is within a few per cent of
// optimal here and costs nothing to reason about. Returns null if the result
// would exceed the GPU's texture limit.
function packSheets(images, maxSize) {
  const sorted = [...images].sort((a, b) => b.height - a.height || b.width - a.width);
  const widest = sorted.reduce((m, i) => Math.max(m, i.width + GUTTER), 1);
  const area = sorted.reduce((s, i) => s + (i.width + GUTTER) * (i.height + GUTTER), 0);
  let w = Math.max(pow2(widest), pow2(Math.ceil(Math.sqrt(area))));
  for (; w <= maxSize; w *= 2) {
    const rects = shelf(sorted, w);
    if (rects && pow2(rects.h) <= maxSize) return { w, h: pow2(rects.h), rects: rects.map };
  }
  return null;
}

function shelf(images, w) {
  const map = new Map();
  let x = 0, y = 0, shelfH = 0;
  for (const im of images) {
    if (im.width + GUTTER > w) return null;
    if (x + im.width + GUTTER > w) { x = 0; y += shelfH; shelfH = 0; }
    map.set(im, { x, y });
    x += im.width + GUTTER;
    shelfH = Math.max(shelfH, im.height + GUTTER);
  }
  return { map, h: y + shelfH };
}

function pow2(n) {
  let p = 1;
  while (p < n) p *= 2;
  return p;
}
