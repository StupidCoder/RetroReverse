// Super Mario Land level viewer — a tilemap rebuilt from the cartridge: each level is
// a row-major grid of 8×8 background tile indices (decoded by extract/level from the
// ROM), drawn from the world's 256-tile atlas. Drag to pan, scroll to zoom. The
// collision and object layers will hang off this later.
import { Application, Container } from 'pixi.js';
import { composeTilemap } from '../tilemap-compose.js';

const DATA = 'public/sml/';
const TILE = 8;
const NATIVE_H = 144; // the Game Boy screen height — the default vertical framing
const ZOOM_STEP = Math.pow(1.15, 0.25);

export class SMLViewer {
  constructor(viewportEl, hudEl) {
    this.el = viewportEl;
    this.hud = hudEl;
    this.app = new Application();
    this.world = new Container();
    this.zoom = 1; this.minZoom = 0.05; this.maxZoom = 16;
    this.layer = null;
    this.src = null;
    this._texMode = 'nearest';
  }

  // No display layers yet (collision / objects come next).
  setLayer() {}

  async init() {
    await this.app.init({ background: 0x0a0e16, antialias: false, resizeTo: this.el, preserveDrawingBuffer: true });
    this.app.canvas.style.imageRendering = 'pixelated';
    this.el.appendChild(this.app.canvas);
    this.app.stage.addChild(this.world);
    this._wireCamera();
    const meta = await fetch(DATA + 'meta.json').then((r) => r.json());
    return meta.levels;
  }

  _loadImage(src) {
    return new Promise((res, rej) => {
      const i = new Image();
      i.onload = () => res(i); i.onerror = rej; i.src = src;
    });
  }

  async loadLevel(meta) {
    this.name = meta.name;
    const level = await fetch(DATA + meta.file).then((r) => r.json());
    const atlas = await this._loadImage(DATA + level.atlas);
    if (this.layer) { this.world.removeChild(this.layer); this.layer.destroy({ children: true }); }
    const { container, src } = composeTilemap(atlas, level.cells, level.width, level.height, { tileSize: TILE });
    this.layer = container;
    this.src = src;
    this.world.addChild(this.layer);
    this.levelW = level.width * TILE;
    this.levelH = level.height * TILE;
    this._fitDefault();
    if (this.hud) this.hud.textContent = `${this.name} · ${level.width}×${level.height} tiles`;
  }

  // --- camera (shared pattern with the Marble/Turrican viewers) -----------
  _fitDefault() {
    const W = this.app.screen.width, H = this.app.screen.height;
    // Frame the Game Boy's screen height (one screenful tall), start at the left edge.
    const z = H / NATIVE_H;
    this.minZoom = Math.min((W / this.levelW) * 0.9, z);
    this.maxZoom = Math.max((W / 160) * 6, z);
    this.zoom = Math.max(this.minZoom, Math.min(this.maxZoom, z));
    this.world.position.set(8, (H - this.levelH * this.zoom) / 2);
    this._apply();
  }
  _screenPt(cx, cy) {
    const r = this.el.getBoundingClientRect();
    return { x: (cx - r.left) * (this.app.screen.width / r.width), y: (cy - r.top) * (this.app.screen.height / r.height) };
  }
  _zoomAt(px, py, f) {
    const wx = (px - this.world.position.x) / this.zoom, wy = (py - this.world.position.y) / this.zoom;
    this.zoom = Math.min(this.maxZoom, Math.max(this.minZoom, this.zoom * f));
    this.world.position.set(px - wx * this.zoom, py - wy * this.zoom);
    this._apply();
  }
  _clampPan() {
    const sw = this.app.screen.width, sh = this.app.screen.height;
    const lw = this.levelW * this.zoom, lh = this.levelH * this.zoom;
    let { x, y } = this.world.position;
    x = lw <= sw ? (sw - lw) / 2 : Math.min(0, Math.max(sw - lw, x));
    y = lh <= sh ? (sh - lh) / 2 : Math.min(0, Math.max(sh - lh, y));
    this.world.position.set(x, y);
  }
  _apply() {
    this.world.scale.set(this.zoom);
    this._clampPan();
    const mode = this.zoom < 1 ? 'linear' : 'nearest';
    if (mode !== this._texMode && this.src) { this._texMode = mode; this.src.scaleMode = mode; }
  }
  _wireCamera() {
    const c = this.el;
    const pts = new Map();
    let pinchDist = 0, pinchMid = null;
    c.addEventListener('pointerdown', (e) => {
      try { c.setPointerCapture(e.pointerId); } catch {}
      pts.set(e.pointerId, { x: e.clientX, y: e.clientY });
      c.classList.add('dragging');
      if (pts.size === 2) {
        const [a, b] = [...pts.values()];
        pinchDist = Math.hypot(a.x - b.x, a.y - b.y);
        pinchMid = { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
      }
    });
    c.addEventListener('pointermove', (e) => {
      const p = pts.get(e.pointerId);
      if (!p) return;
      const dx = e.clientX - p.x, dy = e.clientY - p.y;
      p.x = e.clientX; p.y = e.clientY;
      if (pts.size >= 2) {
        const [a, b] = [...pts.values()];
        const dist = Math.hypot(a.x - b.x, a.y - b.y);
        const mid = { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
        if (pinchMid) { this.world.position.x += mid.x - pinchMid.x; this.world.position.y += mid.y - pinchMid.y; }
        const sp = this._screenPt(mid.x, mid.y);
        this._zoomAt(sp.x, sp.y, pinchDist > 0 ? dist / pinchDist : 1);
        pinchDist = dist; pinchMid = mid;
      } else {
        this.world.position.x += dx; this.world.position.y += dy;
        this._clampPan();
      }
    });
    const end = (e) => {
      pts.delete(e.pointerId);
      try { c.releasePointerCapture(e.pointerId); } catch {}
      if (pts.size < 2) { pinchMid = null; pinchDist = 0; }
      if (pts.size === 0) c.classList.remove('dragging');
    };
    c.addEventListener('pointerup', end);
    c.addEventListener('pointercancel', end);
    c.addEventListener('wheel', (e) => {
      e.preventDefault();
      const sp = this._screenPt(e.clientX, e.clientY);
      this._zoomAt(sp.x, sp.y, e.deltaY < 0 ? ZOOM_STEP : 1 / ZOOM_STEP);
    }, { passive: false });
  }
}
