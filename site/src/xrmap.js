// xrmap.js — the WHOLE 2-D map, wall-sized, hanging in your room.
//
// The screen-shaped mirror in xr2d.js shows a window into the map. This shows
// the map: every cell of it at once, a few metres across, so you walk up to
// read the pixels and step back for the overview. The backdrop colour is keyed
// out, so a Turrican level hangs in the air without its sky and Fort
// Apocalypse without its black.
//
// It does NOT reimplement the tilemap in three.js. level2d.js already builds
// the entire map as Pixi sprites up front — no culling, no windowing — so the
// picture is read straight out of the renderer that already draws it, which
// keeps block indirection, palette regions, tile flips, baked tile-animation
// canvases and every placement sprite exactly as the flat view has them.
//
// The map is read in CHUNKS, one per XR frame, because a single render target
// the size of a Turrican level may not fit the GPU's texture limit and reading
// it in one go would stall the compositor. Chunks that turn out to be entirely
// backdrop are dropped rather than uploaded — on a sky-heavy level that is most
// of them.

import { ARSession, arSupported } from './xr.js';

const CHUNK = 1024;     // px per chunk, before the texture-size clamp
const PAD = 2;          // px of real neighbouring content around each chunk
const BUDGET = 12e6;    // px above which the snapshot drops to half resolution
const SAMPLE = 16;      // histogram every Nth pixel when deriving the key

export class MapAR {
  // { el, params, onStatus, size: () => ({w,h}), snap: {begin, read, end} }
  constructor(opts) {
    this.opts = opts;
    this.supported = arSupported;
    this.onChange = null;
    this.keying = this._param('xrkey') !== 'none';
    this.live = this._param('xrlive') !== '0' && !!opts.live;
    this._ar = null;
    this._liveLayer = null;
    this._teardown = null;
    this._ready = false;
    this._note = '';
  }

  get active() { return !!this._ar?.active; }

  _param(k) {
    const p = this.opts.params;
    return p?.get?.(k) ?? p?.[k];
  }

  _num(k, d) {
    const n = parseFloat(this._param(k));
    return Number.isFinite(n) && n > 0 ? n : d;
  }

  status(t) {
    this.opts.onStatus?.(this._note ? `${t} · ${this._note}` : t);
  }

  // setKeying drops the built map so the next entry re-snapshots with the new
  // setting. Called from the display panel, which is reachable on the flat
  // screen before you put the headset on.
  setKeying(on) {
    if (on === this.keying) return;
    this.keying = on;
    this._dropChunks();
  }

  // Whether this level has anything that moves at all — Turrican pins one frame
  // per placement and has no tile animation, so it has none and the toggle is
  // not offered.
  get hasLive() { return !!this.opts.live; }

  // Same drop-and-rebake as setKeying: the bake genuinely differs, because live
  // mode hides the moving pieces so they are not baked in twice.
  setLive(on) {
    if (!this.opts.live || on === this.live) return;
    this.live = on;
    this._dropChunks();
  }

  async enter() {
    if (!this._ar) await this._build();
    if (!this._ready && this._param('xrsnap') === 'pre') {
      // Escape hatch: if the 2-D renderer turns out not to serve readPixels
      // while a session is presenting, take the snapshot on the flat page
      // first. Costs the user activation on slow maps, hence not the default.
      this._startPump();
      while (!this._ready) this._step();
    }
    await this._ar.enter();
    if (!this._ready) {
      this._startPump();
      this._pump = () => this._step();
      this._ar.stage.updaters.add(this._pump);
    } else if (this.live && this.opts.live && !this._liveLayer) {
      // Re-entering with the chunks still built: re-arm the live layer without
      // paying for another snapshot.
      this.opts.snap.begin();
      this.opts.live.begin();
      this._liveOn = true;
      this._liveStart();
    }
  }

  exit() { this._ar?.exit(); }

  dispose() {
    this._ar?.exit();
    this._abortPump();
    this._liveEnd();
    this._dropChunks();
    this._teardown?.();
    this._teardown = null;
    this._ar = null;
  }

  // _abortPump gives the 2-D view its clock back. snap.begin() stopped the
  // ticker; if the session ends (or the toggle flips) mid-snapshot, nothing
  // else would ever call end() and the flat view would stay frozen for good.
  _abortPump() {
    if (!this._plan) return;
    this._plan = null;
    if (this._pump) { this._ar?.stage.updaters.delete(this._pump); this._pump = null; }
    this._liveEnd();
  }

  // _liveEnd gives the 2-D view its clock back and is the ONLY place that does.
  // Every exit path routes through it — abort, session end, dispose, both
  // toggles — because the ticker now stays stopped for the whole session, and
  // a missed restore would follow the user out of the headset as a permanently
  // frozen flat view.
  _liveEnd() {
    try {
      if (this._liveLayer) {
        this._liveLayer.dispose();
        this._ar?.stage.scene.remove(this._liveLayer.group);
        this._liveLayer = null;
      }
      if (this._liveOn) { this.opts.live.end(); this._liveOn = false; }
    } catch { /* the pixi app can already be gone on a late sessionend */ }
    this.opts.snap.end();
  }

  _dropChunks() {
    this._abortPump();
    this._liveEnd();
    this._ready = false;
    if (!this._group) return;
    for (const m of [...this._group.children]) {
      this._group.remove(m);
      m.geometry.dispose();
      m.material.map?.dispose();
      m.material.dispose();
    }
  }

  async _build() {
    const { el } = this.opts;
    const { THREE, Stage } = await import('./engine3d.js');
    this.THREE = THREE;
    ({ LiveLayer: this.LiveLayer } = await import('./xrlive.js'));

    // An immersive session renders into its own framebuffer, so the canvas
    // backing this context never has to be on screen — and must not be, or it
    // would cover the 2-D view it is reading from.
    const mount = document.createElement('div');
    mount.style.cssText = 'position:absolute;left:0;top:0;width:0;height:0;overflow:hidden;pointer-events:none';
    el.appendChild(mount);
    const stage = new Stage(mount, { fov: 50, near: 0.01, far: 200 });
    this._group = new THREE.Group();
    stage.scene.add(this._group);

    const maxW = this._num('xrw', this._num('xrsize', 4.0));
    const maxH = this._num('xrh', 2.5);

    this._ar = new ARSession({
      stage,
      // The FULL map rect, not the built geometry: dropping the all-backdrop
      // chunks would otherwise shrink the box and re-centre the composition.
      contentBox: () => {
        const { w, h } = this.opts.size();
        return new THREE.Box3(new THREE.Vector3(-w / 2, -h / 2, 0), new THREE.Vector3(w / 2, h / 2, 0));
      },
      // A wall fits a box, not an axis.
      fitScale: (size) => Math.min(maxW / (size.x || 1), maxH / (size.y || 1)),
      ready: () => this._ready,
      dropBelowEye: 0,          // a picture hangs at eye level
      frontDir: new THREE.Vector3(0, 0, -1), // the quads' faces are +Z
      onStatus: (t) => this.opts.onStatus?.(t),
    });
    this._ar.onChange = (on) => {
      // Leaving — mid-snapshot or after one — must always give the flat view
      // its clock back; in live mode the ticker stayed stopped the whole time.
      if (!on) { this._abortPump(); this._liveEnd(); }
      this.onChange?.(on);
    };
    this._maxW = maxW;

    // Registered once, for the life of the stage.
    this._liveStep = (dt) => {
      if (!this._liveLayer || !this._ready) return; // still baking, or still mode
      // Load-bearing: this Stage's setAnimationLoop keeps running on its hidden
      // 0x0 mount forever once built, so without this guard the animations
      // would advance from BOTH clocks whenever the flat page is ticking —
      // everything at double speed.
      if (!stage.renderer.xr.isPresenting) return;
      this.opts.live.advance(Math.min(dt, 0.05));
      this._liveLayer.sync();
    };
    stage.updaters.add(this._liveStep);

    this._teardown = () => { stage.dispose(); mount.remove(); };
  }

  // _startPump plans the chunk grid and freezes the 2-D view. Nothing is read
  // yet — _step does one chunk per call.
  _startPump() {
    const { w, h } = this.opts.size();
    const cap = this._ar.stage.renderer.capabilities.maxTextureSize || 4096;
    const chunk = Math.min(this._num('xrchunk', CHUNK), cap - 2 * PAD);
    const budget = this._num('xrbudget', BUDGET);
    const res = w * h > budget ? 0.5 : 1;
    this._note = res !== 1 ? `half-res: ${(w * h / 1e6).toFixed(1)} Mpx over ${(budget / 1e6).toFixed(0)} Mpx budget` : '';

    const rects = [];
    for (let y = 0; y < h; y += chunk) {
      for (let x = 0; x < w; x += chunk) {
        rects.push({ x, y, w: Math.min(chunk, w - x), h: Math.min(chunk, h - y) });
      }
    }
    this._plan = { rects, i: 0, res, W: w, H: h, phase: 'read', bufs: [], hist: new Map(), dropped: 0, kept: 0 };
    this.opts.snap.begin();
    // Hide the movers BEFORE the first read: they are about to be drawn live,
    // and a baked copy underneath would show as a ghost through their alpha.
    if (this.live && this.opts.live) { this.opts.live.begin(); this._liveOn = true; }
    this.status(`AR: reading map 0/${rects.length}`);
  }

  _step() {
    const p = this._plan;
    if (!p) return;
    try {
      if (p.phase === 'read') this._stepRead(p);
      else this._stepKey(p);
    } catch (e) {
      this._finish(p, `AR: snapshot failed (${e.message || e.name}) — try ?xrsnap=pre`);
    }
  }

  _stepRead(p) {
    const r = p.rects[p.i];
    // PAD pulls in real neighbouring content so each chunk's mip chain does not
    // clamp at its own border and show a seam under minification. Off the map
    // it yields transparency; on a wrapping map it yields the wrapped content.
    // pixi returns { pixels, width, height } — and its width/height are the
    // render target's actual size, which is what the canvas must be built at;
    // recomputing them from the frame would drift on rounding.
    const out = this.opts.snap.read(r.x - PAD, r.y - PAD, r.w + 2 * PAD, r.h + 2 * PAD, p.res);
    const px = out?.pixels;
    if (!px || !px.length) {
      this._finish(p, 'AR: snapshot came back empty — try ?xrsnap=pre');
      return;
    }
    p.bufs.push({ px, cw: out.width, ch: out.height });
    if (this.keying && !this._forcedKey()) {
      const u32 = new Uint32Array(px.buffer);
      for (let i = 0; i < u32.length; i += SAMPLE) {
        const v = u32[i];
        if ((v & 0xff000000) === 0) continue; // fully transparent, not a colour
        p.hist.set(v, (p.hist.get(v) || 0) + 1);
      }
    }
    p.i++;
    if (p.i < p.rects.length) {
      this.status(`AR: reading map ${p.i}/${p.rects.length}`);
    } else {
      p.phase = 'key';
      p.i = 0;
      p.key = this.keying ? this._pickKey(p) : null;
      this.status(`AR: building ${0}/${p.rects.length}`);
    }
  }

  _forcedKey() {
    const v = this._param('xrkey');
    if (!v || v === 'none') return null;
    const m = /^#?([0-9a-f]{6})$/i.exec(v);
    if (!m) return null;
    const n = parseInt(m[1], 16);
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
  }

  _pickKey(p) {
    const forced = this._forcedKey();
    if (forced) {
      this._note = `key #${hex(forced)} (forced)`;
      return forced;
    }
    let best = 0, bestN = 0, total = 0;
    for (const [v, n] of p.hist) { total += n; if (n > bestN) { bestN = n; best = v; } }
    if (!total) return null;
    const key = { r: best & 255, g: (best >> 8) & 255, b: (best >> 16) & 255 };
    this._note = `key #${hex(key)} ${Math.round((100 * bestN) / total)}%`;
    return key;
  }

  _stepKey(p) {
    const THREE = this.THREE;
    const r = p.rects[p.i];
    const { px, cw, ch } = p.bufs[p.i];
    p.bufs[p.i] = null; // let the buffer go as soon as it is uploaded
    const tol = Math.max(0, Math.floor(this._num('xrtol', 0)));

    let kept = 0;
    if (p.key) {
      const u32 = new Uint32Array(px.buffer);
      // Store 0 on a match: that zeroes RGB *and* alpha in one write, which is
      // what keeps the buffer premultiplied-correct — a keyed pixel must be
      // (0,0,0,0), not (sky,0), or the mip chain averages sky into the edges.
      if (tol === 0) {
        // >>> 0: `|` yields a SIGNED int32, so 0xff000000|… is negative while
        // the Uint32Array elements are unsigned — the comparison would never
        // match and every pixel would survive the key.
        const want = ((0xff000000 | (p.key.b << 16) | (p.key.g << 8) | p.key.r) >>> 0);
        for (let i = 0; i < u32.length; i++) {
          if (u32[i] === want) u32[i] = 0; else if (u32[i] & 0xff000000) kept++;
        }
      } else {
        for (let i = 0; i < u32.length; i++) {
          const v = u32[i];
          if ((v & 0xff000000) === 0) continue;
          if (Math.abs((v & 255) - p.key.r) <= tol && Math.abs(((v >> 8) & 255) - p.key.g) <= tol
            && Math.abs(((v >> 16) & 255) - p.key.b) <= tol) u32[i] = 0;
          else kept++;
        }
      }
    } else {
      kept = 1;
    }

    if (kept === 0) {
      p.dropped++;
    } else {
      p.kept++;
      const cv = document.createElement('canvas');
      cv.width = cw; cv.height = ch;
      const img = new ImageData(px instanceof Uint8ClampedArray ? px : new Uint8ClampedArray(px.buffer), cw, ch);
      cv.getContext('2d').putImageData(img, 0, 0);

      const tex = new THREE.CanvasTexture(cv);
      tex.colorSpace = THREE.SRGBColorSpace;
      tex.generateMipmaps = true;
      tex.magFilter = THREE.NearestFilter;            // game pixels, up close
      tex.minFilter = THREE.LinearMipmapLinearFilter; // clean from across the room
      tex.wrapS = tex.wrapT = THREE.ClampToEdgeWrapping;
      tex.anisotropy = Math.min(8, this._ar.stage.renderer.capabilities.getMaxAnisotropy());
      tex.needsUpdate = true;

      // Geometry comes from the FRAME, never the canvas, so half-res chunks
      // still span the right area.
      const geo = new THREE.PlaneGeometry(r.w, r.h);
      // Inset the UVs past the padding, so the padded ring feeds the mip chain
      // but is never itself sampled.
      const uv = geo.attributes.uv;
      const iu = (PAD * p.res) / cw, iv = (PAD * p.res) / ch;
      for (let i = 0; i < uv.count; i++) {
        uv.setXY(i, iu + uv.getX(i) * (1 - 2 * iu), iv + uv.getY(i) * (1 - 2 * iv));
      }
      uv.needsUpdate = true;

      const mesh = new THREE.Mesh(geo, new THREE.MeshBasicMaterial({
        map: tex, toneMapped: false, side: THREE.DoubleSide,
        transparent: !!p.key, premultipliedAlpha: !!p.key, depthWrite: !p.key,
      }));
      // 1 world unit = 1 map pixel, origin at the map's centre, +Y up.
      mesh.position.set(r.x + r.w / 2 - p.W / 2, p.H / 2 - r.y - r.h / 2, 0);
      this._group.add(mesh);
    }

    p.i++;
    if (p.i < p.rects.length) {
      this.status(`AR: building ${p.i}/${p.rects.length}`);
    } else {
      const note = this._note ? `${this._note} · ` : '';
      this._note = `${note}${p.kept}/${p.rects.length} chunks${p.dropped ? `, ${p.dropped} all-backdrop` : ''}`;
      this._finish(p, null);
    }
  }

  _finish(p, err) {
    if (this._pump) { this._ar.stage.updaters.delete(this._pump); this._pump = null; }
    this._plan = null;
    p.bufs.length = 0;
    if (err) { this._liveEnd(); this.status(err); return; }
    this._W = p.W; this._H = p.H;
    if (this._liveOn) this._liveStart();
    else this.opts.snap.end(); // still mode: the 2-D view gets its clock back now
    // Stand back far enough to take the whole thing in.
    this._ar.distance = this._num('xrdist', Math.max(1.2, 0.62 * this._maxW));
    this._ready = true;
  }

  // In live mode the Pixi ticker stays stopped for the whole session and the
  // XR frame loop advances the same `anims` array instead — window rAF is not
  // reliably serviced while an immersive session is presenting, which is why
  // three swaps to the session's own rAF in setAnimationLoop.
  _liveStart() {
    try {
      const layer = new this.LiveLayer(this.THREE, this.opts.live, {
        W: this._W, H: this._H, renderer: this._ar.stage.renderer,
      });
      this._ar.stage.scene.add(layer.group);
      this._liveLayer = layer;
      const c = layer.counts;
      const bits = [];
      if (c.sprites) bits.push(`${c.sprites} sprites/${c.atlases}`);
      if (c.tiles) bits.push(`${c.tiles} anim tiles`);
      if (c.cells) bits.push(`${c.cells} strips`);
      if (bits.length) this._note = `${this._note ? `${this._note} · ` : ''}live ${bits.join(', ')}`;
    } catch (e) {
      // A broken live layer must not cost the user the map: fall back to the
      // still, which is what they had before.
      this._liveEnd();
      this._note = `${this._note ? `${this._note} · ` : ''}live failed: ${e.message || e.name}`;
    }
  }
}

function hex({ r, g, b }) {
  return ((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1);
}
