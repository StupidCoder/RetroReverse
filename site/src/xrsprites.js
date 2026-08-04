// xrsprites.js — every placement in a 2-D level, in one draw call.
//
// The level's sprite sheets are packed into a single atlas at load, so the
// whole cast shares one texture and one buffer no matter how many distinct
// objects a level places (Marble's aerial has 21 sheets for 21 quads; Turrican
// has 11 for 226). Tint rides along as a vertex colour rather than splitting
// the draw.
//
// Nothing here re-implements sprite2d.js. Its SpriteInstances are the ones
// ticking, and this reads their state — the frame program's current column,
// the animation's row, the path offset, the anchor pivot — and writes quads.
// Porting the four loop modes, the mirror link and the stepped |0 path sample
// would buy nothing and would diverge the first time that file changed.
//
// The buffer is ordered [animating | static] and sync() rewrites only the
// animating prefix. Turrican's 226 placements all pin a single frame with no
// path, so its per-frame vertex write is zero bytes.

const VERTS = 6;
const GUTTER = 2; // transparent margin around each packed sheet

export class SpriteLayer {
  // (THREE, { placements, clip, W, H, renderer, blend })
  //   placements: [{ inst, pl, obj, visible() }]  — live SpriteInstances
  //   clip:       { w, h } | null — the playfield, for path-animated pieces
  constructor(THREE, opts) {
    this.THREE = THREE;
    const { placements, clip = null, W, H, renderer, blend = false } = opts;
    this.W = W; this.H = H;
    this.clip = clip;
    this.group = new THREE.Group();
    this._scratch = {};

    // ---- pack the sheets ---------------------------------------------------------
    // Keyed by resolved URL, not object id: Sonic has 32 objects sharing
    // marker.png, and the flat view loads it once per object.
    const sheets = new Map();
    for (const p of placements) {
      const img = p.obj.img;
      if (img && !sheets.has(img.src)) sheets.set(img.src, img);
    }
    const packed = pack([...sheets.values()]);
    this.atlas = packed;

    const tex = new THREE.CanvasTexture(packed.canvas);
    tex.colorSpace = THREE.SRGBColorSpace;
    tex.magFilter = THREE.NearestFilter;
    // No mip chain: the sheets are packed side by side, so a mip would average
    // one sprite into its neighbour. They are small on screen at wall scale
    // and there is no gutter wide enough to fix that at every level.
    tex.minFilter = THREE.LinearFilter;
    tex.generateMipmaps = false;
    tex.wrapS = tex.wrapT = THREE.ClampToEdgeWrapping;
    tex.needsUpdate = true;
    this.texture = tex;

    // ---- classify: only some placements ever change ---------------------------------
    const items = placements
      .filter((p) => p.obj?.img && p.visible?.() !== false)
      .map((p) => ({
        p,
        rect: packed.rects.get(p.obj.img.src),
        cw: p.obj.cellW, ch: p.obj.cellH,
        live: (p.inst.prog?.length > 1) || !!p.inst.anim?.path?.length,
        tint: p.inst.sprite?.tint,
      }));
    items.sort((a, b) => (b.live ? 1 : 0) - (a.live ? 1 : 0)); // animating first
    this.items = items;
    this.liveCount = items.filter((i) => i.live).length;

    const n = items.length;
    const pos = new Float32Array(n * VERTS * 3);
    const uv = new Float32Array(n * VERTS * 2);
    const col = new Float32Array(n * VERTS * 3);
    this.pos = pos; this.uv = uv;

    const c = new THREE.Color();
    items.forEach((it, i) => {
      // Tint as a vertex colour keeps every placement in the same draw call.
      const t = it.tint && it.tint !== 0xffffff ? it.tint : 0xffffff;
      c.setHex(t, THREE.SRGBColorSpace);
      for (let k = 0; k < VERTS; k++) {
        const o = (i * VERTS + k) * 3;
        col[o] = c.r; col[o + 1] = c.g; col[o + 2] = c.b;
      }
    });

    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
    geo.setAttribute('color', new THREE.BufferAttribute(col, 3));
    this.geometry = geo;

    this.material = new THREE.MeshBasicMaterial({
      map: tex, vertexColors: true, toneMapped: false, side: THREE.DoubleSide,
      alphaTest: 0.5, transparent: blend, depthWrite: !blend,
    });
    this.mesh = new THREE.Mesh(geo, this.material);
    this.mesh.frustumCulled = false;
    this.mesh.renderOrder = 3;   // over the map and its strips
    this.mesh.position.z = 1.5;  // and clear of them: everything writes depth
    this.group.add(this.mesh);

    this.sync(true); // static quads written once, here
  }

  get counts() {
    return {
      sprites: this.items.length, live: this.liveCount, sheets: this.atlas.rects.size,
      atlas: `${this.atlas.canvas.width}x${this.atlas.canvas.height}`,
      draws: 1, tris: this.items.length * 2,
    };
  }

  // sync(all) rewrites the animating prefix; `all` also does the static tail,
  // which only happens once at construction.
  sync(all = false) {
    const n = all ? this.items.length : this.liveCount;
    if (!n) return;
    const o = this._scratch;
    const aw = this.atlas.canvas.width, ah = this.atlas.canvas.height;
    let drawn = 0;
    for (let i = 0; i < n; i++) {
      const it = this.items[i];
      read(it, o);
      let { x, y } = o;
      let w = it.cw, h = it.ch;
      let u0 = (it.rect.x + o.col * it.cw) / aw;
      let u1 = u0 + it.cw / aw;
      let v0 = (it.rect.y + o.row * it.ch) / ah;
      let v1 = v0 + it.ch / ah;
      if (o.flip) { const t = u0; u0 = u1; u1 = t; }
      if (this.clip) {
        // The playfield clips its edges, so a path-animated piece sweeping past
        // one is cut rather than left hanging in the room (Marble).
        const lx = Math.max(0, x), rx = Math.min(this.clip.w, x + w);
        const ty = Math.max(0, y), by = Math.min(this.clip.h, y + h);
        if (rx <= lx || by <= ty) { degenerate(this.pos, i); continue; }
        if (lx !== x || rx !== x + w || ty !== y || by !== y + h) {
          const du = u1 - u0, dv = v1 - v0;
          const fu0 = (lx - x) / w, fu1 = (rx - x) / w;
          const fv0 = (ty - y) / h, fv1 = (by - y) / h;
          u1 = u0 + du * fu1; u0 += du * fu0;
          v1 = v0 + dv * fv1; v0 += dv * fv0;
          x = lx; y = ty; w = rx - lx; h = by - ty;
        }
      }
      this._quad(i, x, y, w, h, u0, u1, v0, v1);
      drawn++;
    }
    const posAttr = this.geometry.attributes.position;
    const uvAttr = this.geometry.attributes.uv;
    // r160's updateRange is getter-only; partial uploads go through
    // addUpdateRange/clearUpdateRanges. Only the animating prefix is ever
    // dirty, so a level whose placements all pin one frame uploads nothing.
    posAttr.clearUpdateRanges?.();
    uvAttr.clearUpdateRanges?.();
    if (!all && posAttr.addUpdateRange) {
      posAttr.addUpdateRange(0, n * VERTS * 3);
      uvAttr.addUpdateRange(0, n * VERTS * 2);
    }
    posAttr.needsUpdate = true;
    uvAttr.needsUpdate = true;
    if (all) this.geometry.attributes.color.needsUpdate = true;
    this.drawn = drawn;
  }

  _quad(i, x, y, w, h, u0, u1, v0, v1) {
    // Map pixels, +Y down, origin top-left -> group units, +Y up, centred:
    // the same mapping the tile bricks use.
    const X0 = x - this.W / 2, X1 = x + w - this.W / 2;
    const Y0 = this.H / 2 - y, Y1 = this.H / 2 - y - h;
    const p = i * VERTS * 3, t = i * VERTS * 2;
    const P = [X0, Y0, 0, X1, Y0, 0, X1, Y1, 0, X0, Y0, 0, X1, Y1, 0, X0, Y1, 0];
    const T = [u0, 1 - v0, u1, 1 - v0, u1, 1 - v1, u0, 1 - v0, u1, 1 - v1, u0, 1 - v1];
    for (let k = 0; k < 18; k++) this.pos[p + k] = P[k];
    for (let k = 0; k < 12; k++) this.uv[t + k] = T[k];
  }

  dispose() {
    this.geometry.dispose();
    this.material.dispose();
    this.texture.dispose();
  }
}

// read fills a scratch object from the SpriteInstance's own fields — not from
// the Pixi display objects, whose x also carries the cylinder-wrap hop, a
// camera artefact with no meaning on a map hanging in a room.
function read(it, o) {
  const inst = it.p.inst, s = inst.sprite, pl = it.p.pl;
  o.col = inst.prog[inst.idx][0];
  o.row = inst.anim.row || 0;
  o.flip = s.scale.x < 0;
  // A mirrored sprite flips about its own origin, so its left edge moves to
  // pos + pivot - cellW.
  o.x = pl.pos[0] + s.position.x + (o.flip ? s.pivot.x - it.cw : -s.pivot.x);
  o.y = pl.pos[1] + s.position.y - s.pivot.y;
}

function degenerate(pos, i) {
  const p = i * VERTS * 3;
  for (let k = 0; k < 18; k++) pos[p + k] = 0;
}

// pack lays the sheets out in shelves on one power-of-two-wide canvas, with a
// transparent gutter so neighbours cannot bleed into each other.
function pack(images) {
  const area = images.reduce((a, im) => a + (im.width + GUTTER * 2) * (im.height + GUTTER * 2), 0);
  let width = 64;
  const widest = images.reduce((m, im) => Math.max(m, im.width + GUTTER * 2), 1);
  while (width < widest || width * width < area) width *= 2;
  const sorted = [...images].sort((a, b) => b.height - a.height);
  const rects = new Map();
  let x = GUTTER, y = GUTTER, shelf = 0;
  for (const im of sorted) {
    if (x + im.width + GUTTER > width) { x = GUTTER; y += shelf + GUTTER * 2; shelf = 0; }
    rects.set(im.src, { x, y, w: im.width, h: im.height });
    x += im.width + GUTTER * 2;
    shelf = Math.max(shelf, im.height);
  }
  const height = nextPow2(y + shelf + GUTTER);
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = Math.max(1, height);
  const c = canvas.getContext('2d');
  c.imageSmoothingEnabled = false;
  for (const im of sorted) {
    const r = rects.get(im.src);
    c.drawImage(im, r.x, r.y);
  }
  return { canvas, rects };
}

const nextPow2 = (v) => { let n = 1; while (n < v) n *= 2; return n; };
