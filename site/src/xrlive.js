// xrlive.js — the moving parts of a 2-D level, drawn over the baked AR map.
//
// The map itself is a still: xrmap.js reads it out of PixiJS in chunks, which
// means stopping the clock (chunks arrive one per frame, so a running
// animation would freeze at a different phase in each one). This puts the
// motion back — Sonic's idle loop, the moving platforms, the sparkling water —
// as real three.js geometry on top, so the static 95% stays baked and only the
// handful of things that actually move cost anything.
//
// Nothing here re-implements the 2-D runtime. sprite2d.js's SpriteInstances
// are still the ones ticking; this reads their state each XR frame and writes
// quads from it. Same for the tile animations: every one of them works by
// repainting a canvas, so the canvas becomes a texture on STATIC geometry and
// only the pixels change.

const VERTS = 6; // two triangles, non-indexed — the quad counts are tiny

// The live quads sit ON the baked map, which is at z = 0. Alpha-tested
// geometry writes depth, so exactly-coplanar quads would z-fight; these push
// them apart in MAP PIXELS, which at a typical fit (~0.6 mm per pixel) is far
// too small to see and far too large to fight.
const Z_SURFACES = 0.5;
const Z_CELLS = 0.75;
const Z_SPRITES = 1.0;

export class LiveLayer {
  // (THREE, live, { W, H, renderer }) — W/H are the map's pixel size; the
  // group's coordinate system is 1 unit = 1 map pixel, +Y up, origin centre.
  constructor(THREE, live, { W, H, renderer, key = null, aniso = 4, blend = false }) {
    this.THREE = THREE;
    this.live = live;
    this.W = W;
    this.H = H;
    this.renderer = renderer;
    // The backdrop colour the map bake keyed out. Tile art and cell strips are
    // whole opaque tiles — sky pixels and all — so without this they come back
    // as solid sky-coloured blocks over the keyed map. Sprites are cut-out art
    // with real alpha and are left alone.
    this.key = key;
    this.aniso = aniso;
    this.blend = blend;
    this.group = new THREE.Group();
    this._scratch = {};
    this._clip = live.clip?.() || null;

    this.sprites = (live.sprites || []).map((g) => this._buildSprites(g, renderer));
    this.surfaces = (live.surfaces || []).map((s) => this._buildSurface(s));
    this.cells = (live.cells || []).map((c) => this._buildCell(c));
  }

  // _keyed copies a canvas through a private one with the backdrop punched to
  // transparent. Cheap: these are single tiles or blocks (32x32 for Sonic), and
  // it only runs when the source is actually repainted.
  _keyed(src, dst) {
    const w = src.width, h = src.height;
    if (!dst) { dst = document.createElement('canvas'); dst.width = w; dst.height = h; }
    const c = dst.getContext('2d', { willReadFrequently: true });
    c.clearRect(0, 0, w, h);
    c.drawImage(src, 0, 0);
    if (!this.key) return dst;
    const img = c.getImageData(0, 0, w, h);
    const u32 = new Uint32Array(img.data.buffer);
    const want = ((0xff000000 | (this.key.b << 16) | (this.key.g << 8) | this.key.r) >>> 0);
    for (let i = 0; i < u32.length; i++) if (u32[i] === want) u32[i] = 0;
    c.putImageData(img, 0, 0);
    return dst;
  }

  get counts() {
    const q = this.sprites.reduce((n, g) => n + g.items.length, 0);
    const tiles = this.surfaces.reduce((n, s) => n + s.rects.length, 0);
    return { sprites: q, atlases: this.sprites.length, surfaces: this.surfaces.length, tiles, cells: this.cells.length };
  }

  // ---- construction -------------------------------------------------------------

  _texture(image, renderer) {
    const THREE = this.THREE;
    const t = new THREE.Texture(image);
    t.colorSpace = THREE.SRGBColorSpace; // without this the live art sits darker than the bake
    t.magFilter = THREE.NearestFilter;
    t.minFilter = THREE.LinearMipmapLinearFilter;
    t.generateMipmaps = true;
    t.wrapS = t.wrapT = THREE.ClampToEdgeWrapping;
    if (renderer && this.aniso > 1) t.anisotropy = Math.min(this.aniso, renderer.capabilities.getMaxAnisotropy());
    t.needsUpdate = true;
    return t;
  }

  _material(map) {
    const THREE = this.THREE;
    // Alpha is BINARY here — cut-out sprite art, and keyed tile art — so an
    // alpha test in the OPAQUE pass is both correct and much cheaper on a
    // tiler than blended geometry: it keeps depth writes, so overlapping quads
    // reject each other's fragments instead of every layer paying full fill.
    // ?xrblend=1 puts it back in the transparent pass.
    // premultipliedAlpha is false regardless, unlike the baked chunks: those
    // are premultiplied because the keyer writes (0,0,0,0), while a PNG or
    // canvas upload is straight alpha.
    return new THREE.MeshBasicMaterial({
      map, toneMapped: false, side: THREE.DoubleSide,
      alphaTest: 0.5, premultipliedAlpha: false,
      transparent: this.blend, depthWrite: !this.blend,
    });
  }

  _buildSprites(g, renderer) {
    const THREE = this.THREE;
    const n = g.items.length;
    const geo = new THREE.BufferGeometry();
    const pos = new Float32Array(n * VERTS * 3);
    const uv = new Float32Array(n * VERTS * 2);
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
    geo.setDrawRange(0, n * VERTS);
    const tex = this._texture(g.img, renderer);
    const mat = this._material(tex);
    if (g.tint) mat.color = new THREE.Color(g.tint);
    const mesh = new THREE.Mesh(geo, mat);
    mesh.frustumCulled = false; // the quads move; the group is small anyway
    mesh.renderOrder = 2;
    mesh.position.z = Z_SPRITES;
    this.group.add(mesh);
    return { ...g, geo, pos, uv, mesh, iw: g.img.width, ih: g.img.height };
  }

  // A surface is one repainted canvas plus every map cell drawing it. The
  // geometry never changes — only the canvas contents do — so it is written
  // once here and the per-frame cost is a texture upload when the canvas is
  // actually dirty.
  _buildSurface(s) {
    const THREE = this.THREE;
    // Draw from a keyed COPY, refreshed whenever the source is repainted —
    // the tile art is opaque and would otherwise punch sky-coloured blocks
    // through the keyed map.
    const own = this.key ? this._keyed(s.canvas) : s.canvas;
    const tex = new THREE.CanvasTexture(own);
    tex.colorSpace = THREE.SRGBColorSpace;
    tex.magFilter = THREE.NearestFilter;
    tex.minFilter = THREE.LinearMipmapLinearFilter;
    tex.generateMipmaps = true;
    if (this.renderer && this.aniso > 1) {
      tex.anisotropy = Math.min(this.aniso, this.renderer.capabilities.getMaxAnisotropy());
    }
    const geo = new THREE.BufferGeometry();
    const n = s.rects.length;
    const pos = new Float32Array(n * VERTS * 3);
    const uv = new Float32Array(n * VERTS * 2);
    s.rects.forEach((r, i) => {
      this._writeQuad(pos, uv, i, r.x, r.y, r.w, r.h, r.flip ? 1 : 0, r.flip ? 0 : 1, 0, 1);
    });
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
    const mesh = new THREE.Mesh(geo, this._material(tex));
    mesh.frustumCulled = false;
    mesh.renderOrder = 1; // under the sprites, over the baked chunks
    mesh.position.z = Z_SURFACES;
    this.group.add(mesh);
    return { ...s, tex, mesh, own: this.key ? own : null };
  }

  _buildCell(c) {
    const THREE = this.THREE;
    // Phase canvases are pre-baked and never change, so key them once here.
    const texs = c.canvases.map((cv) => {
      const t = new THREE.CanvasTexture(this.key ? this._keyed(cv) : cv);
      t.colorSpace = THREE.SRGBColorSpace;
      t.magFilter = THREE.NearestFilter;
      t.minFilter = THREE.LinearMipmapLinearFilter;
      return t;
    });
    const geo = new THREE.BufferGeometry();
    const pos = new Float32Array(VERTS * 3);
    const uv = new Float32Array(VERTS * 2);
    this._writeQuad(pos, uv, 0, c.x, c.y, c.w, c.h, 0, 1, 0, 1);
    geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
    const mesh = new THREE.Mesh(geo, this._material(texs[0]));
    mesh.frustumCulled = false;
    mesh.renderOrder = 1;
    mesh.position.z = Z_CELLS;
    this.group.add(mesh);
    return { ...c, texs, mesh, shown: -1 };
  }

  // _writeQuad lays one quad into the arrays. x/y/w/h are in MAP pixels with
  // +Y down and the origin top-left; the group's space is +Y up around the
  // centre, which is the same mapping the baked chunks use.
  _writeQuad(pos, uv, i, x, y, w, h, u0, u1, v0, v1) {
    const X0 = x - this.W / 2, X1 = x + w - this.W / 2;
    const Y0 = this.H / 2 - y, Y1 = this.H / 2 - y - h;
    // v is flipped: texture v=0 is the BOTTOM, map y=0 is the TOP
    const V0 = 1 - v0, V1 = 1 - v1;
    const p = i * VERTS * 3, t = i * VERTS * 2;
    const P = [X0, Y0, 0, X1, Y0, 0, X1, Y1, 0, X0, Y0, 0, X1, Y1, 0, X0, Y1, 0];
    const T = [u0, V0, u1, V0, u1, V1, u0, V0, u1, V1, u0, V1];
    for (let k = 0; k < 18; k++) pos[p + k] = P[k];
    for (let k = 0; k < 12; k++) uv[t + k] = T[k];
  }

  // ---- per-frame ----------------------------------------------------------------

  sync() {
    const o = this._scratch;
    for (const g of this.sprites) {
      let n = 0;
      for (const it of g.items) {
        it.read(o);
        if (o.hidden) continue;
        let { x, y } = o;
        let w = g.cw, h = g.ch;
        // Atlas rect for the current (row, col), in UV.
        let u0 = (o.col * g.cw) / g.iw, u1 = u0 + g.cw / g.iw;
        let v0 = (o.row * g.ch) / g.ih, v1 = v0 + g.ch / g.ih;
        if (o.flip) { const t = u0; u0 = u1; u1 = t; }
        // Clip to the playfield, matching pixi's objLayer mask: a path-animated
        // piece sweeping past the edge must be cut, not left hanging in the room.
        if (this._clip) {
          const cw = this._clip.w, ch = this._clip.h;
          const lx = Math.max(0, x), rx = Math.min(cw, x + w);
          const ty = Math.max(0, y), by = Math.min(ch, y + h);
          if (rx <= lx || by <= ty) continue;
          if (lx !== x || rx !== x + w || ty !== y || by !== y + h) {
            const fu0 = (lx - x) / w, fu1 = (rx - x) / w;
            const fv0 = (ty - y) / h, fv1 = (by - y) / h;
            const du = u1 - u0, dv = v1 - v0;
            u1 = u0 + du * fu1; u0 += du * fu0;
            v1 = v0 + dv * fv1; v0 += dv * fv0;
            x = lx; y = ty; w = rx - lx; h = by - ty;
          }
        }
        this._writeQuad(g.pos, g.uv, n++, x, y, w, h, u0, u1, v0, v1);
      }
      g.geo.setDrawRange(0, n * VERTS);
      g.geo.attributes.position.needsUpdate = true;
      g.geo.attributes.uv.needsUpdate = true;
      g.mesh.visible = n > 0;
    }
    for (const s of this.surfaces) {
      if (!s.dirty()) continue;
      if (s.own) this._keyed(s.canvas, s.own); // re-key the repainted tile
      s.tex.needsUpdate = true;
      s.clear();
    }
    for (const c of this.cells) {
      const i = c.index();
      if (i === c.shown) continue;
      c.shown = i;
      c.mesh.material.map = c.texs[i] || c.texs[0];
    }
  }

  dispose() {
    for (const m of [...this.group.children]) {
      this.group.remove(m);
      m.geometry.dispose();
      m.material.map?.dispose();
      m.material.dispose();
    }
    for (const c of this.cells) for (const t of c.texs) t.dispose();
  }
}
