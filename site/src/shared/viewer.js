// The shared 2-D level viewer: consumes the common format (site/FORMAT2.md) and
// composes the shared pieces — MapCamera, Tilemap, overlay layers, AnimRunner.
// Per-game differences live in a small config (site/src/<game>/config.js):
//
//   { base: 'public/super-mario-land-gb/',           // data directory
//     strategy: 'sliced'|'baked',    // tile texture strategy (baked = tile anim games)
//     maxNativeFactor: 4,            // zoom-in cap, in multiples of the native view
//     markerCat(o) -> category,      // marker colour class for sprite-less objects
//     hud(level) -> string,          // HUD caption
//     objectInfo?: { <name>: { title, text } },  // click-to-inspect cards: keyed by an
//                                    // object's/pool's `name` (spawn = "player"). Present
//                                    // -> the viewer wires object picking (Fort).
//     hooks?: { loaded(viewer, level) }   // game-specific extension point
//   }

import { Application, Container, Sprite, Texture } from 'pixi.js';
import { MapCamera } from './camera.js';
import { GameData } from './data.js';
import { Tilemap, BlockTilemap } from './tilemap.js';
import { AnimRunner } from './anim.js';
import { buildWaterInfo, setupCycle } from './palettefx.js';
import { buildCollisionGrid, buildCollisionProfiles, buildObjects, buildPools, rng } from './layers.js';

// Compact identifier line for an info card: whichever of an object's raw ids are present,
// in a fixed order. `type` is shown in hex (our RE docs' convention); `handler` is an AI
// routine address string (Turrican); `sprite` is the atlas / frame-table key. Empty when
// the pick carries no ids (e.g. Fort's randomized pools, keyed only by name).
function idLine(pick) {
  const m = pick.meta || {};
  const parts = [];
  if (m.type != null) parts.push('type $' + Number(m.type).toString(16).toUpperCase());
  // Turrican: the placement's low byte (node+$1E) is the orientation/direction the AI
  // handler reads; `frame` is the frame it picked for that orientation (node+$C).
  if (m.orient != null) parts.push('orient $' + Number(m.orient).toString(16).toUpperCase());
  if (m.handler) parts.push('AI $' + m.handler);
  if (m.hard) parts.push('hard-mode');
  if (m.sprite) parts.push('sprite ' + m.sprite);
  return parts.join('   ·   ');
}

export class LevelViewer {
  constructor(viewportEl, hudEl, config) {
    this.el = viewportEl;
    this.hud = hudEl;
    this.config = config;
    this.data = new GameData(config.base);
    this.app = new Application();
    this.world = new Container();
    this.tileLayer = new Container();
    this.collisionLayer = new Container();
    this.objectLayer = new Container();
    this.poolLayer = new Container();      // randomized placements (rebuilt per toggle)
    this.collisionLayer.visible = false;
    this._texMode = 'nearest';
    this.level = null;
    this.tilemap = null;
    this.meta = null;
    // Clickable objects for the info card (only collected when the game's config
    // supplies `objectInfo`). `picks` are the fixed placements + spawn, rebuilt per
    // level; `poolPicks` the randomized pool placements, rebuilt on every re-roll.
    this.picks = [];
    this.poolPicks = [];
  }

  async init() {
    await this.app.init({ background: 0x000000, antialias: false, resizeTo: this.el, preserveDrawingBuffer: true });
    this.el.appendChild(this.app.canvas);
    this.world.addChild(this.tileLayer, this.collisionLayer, this.objectLayer, this.poolLayer);
    this.app.stage.addChild(this.world);
    this.meta = await this.data.loadMeta();
    this.anim = new AnimRunner(this.app, this.meta.tickHz);
    this.cam = new MapCamera(this, {
      enabled: () => (this.config.camEnabled ? this.config.camEnabled() : true),
      bounds: () => ({ w: this.levelW || 1, h: this.levelH || 1 }),
      wrapX: () => (this.meta.wrap === 'x' && this.levelW ? this.levelW : 0),
      onApply: () => {
        const mode = this.cam.zoom < 1 ? 'linear' : 'nearest';
        if (mode !== this._texMode) {
          this._texMode = mode;
          if (this.tilemap) this.tilemap.setScaleMode(mode);
        }
        if (this.hud && this.level && this.config.hud) this.hud.textContent = this.config.hud(this.level);
      },
    });
    this.cam.wirePointer();
    if (this.config.objectInfo) this._wireObjectPicking();
    return this.meta;
  }

  async loadLevel(entry) {
    const level = await this.data.loadLevel(entry);
    this.level = level;

    // Drop every consumer of the old tilemap's textures BEFORE destroying it —
    // the awaits below yield to the render loop, and a tick drawing a sprite
    // whose baked TextureSource is gone crashes the renderer (pool stamps) or
    // paints into a dead canvas (tile anims).
    this.anim.reset();
    this._hideCard();
    this.picks = [];
    this.poolPicks = [];
    this.poolLayer.removeChildren();
    this.objectLayer.removeChildren();
    this.collisionLayer.removeChildren();
    this.tileLayer.removeChildren();

    // tilemap (owns its textures; destroyed on reload). Games with block
    // indirection (level.blocks) use the block-baked strategy automatically.
    if (this.tilemap) this.tilemap.destroy();
    for (const m of this._extraMaps || []) m.destroy();
    this._extraMaps = [];
    const atlasImg = await this.data.atlasImage(level.grid.atlas);
    const mkMap = () => level.blocks
      ? new BlockTilemap(level, atlasImg, { water: buildWaterInfo(level) })
      : new Tilemap(level, atlasImg, { strategy: this.config.strategy || 'sliced' });
    this.tilemap = mkMap();
    const copies = this.meta.wrap === 'x' ? 3 : 1;
    this.tileLayer.addChild(this.tilemap.container);
    this.levelW = this.tilemap.widthPx;
    this.levelH = this.tilemap.heightPx;
    for (let i = 1; i < copies; i++) {
      const dup = mkMap();
      dup.container.x = i * this.levelW;
      this.tileLayer.addChild(dup.container);
      this._extraMaps.push(dup);
    }
    this._texMode = 'nearest';

    // animation subsystems
    for (const a of level.tileAnims || []) {
      this.anim.tileAnims.push({
        anim: a, acc: 0, step: 0,
        paint: (tileId, atlasTile) => {
          this.tilemap.paintTile(tileId, atlasTile);
          for (const m of this._extraMaps || []) m.paintTile(tileId, atlasTile);
          // cellAnim phase canvases showing this tile follow too (Ultimate's
          // screen-swap band contains shimmering gold tiles)
          for (const c of this.anim.cellAnims) c.paintTile && c.paintTile(tileId, atlasTile);
        },
      });
    }
    // placement-anchored cell animators: tw x th tile strips (default 2x4, the
    // Sonic $50 shape; Marble's screen-swap uses 36x30) baked from the atlas,
    // one sprite each in the tile layer, per-phase hold times
    for (const ca of level.cellAnims || []) {
      const ts = level.grid.tileSize;
      const tw = ca.tw ?? 2, th = ca.th ?? 4;
      const cols = level.grid.atlasCols ?? 16;
      const cell = ts + 2 * (level.grid.atlasGutter ?? 0);
      const drawTile = (ctx, tile, i) => ctx.drawImage(atlasImg,
        (tile % cols) * cell + (level.grid.atlasGutter ?? 0),
        ((tile / cols) | 0) * cell + (level.grid.atlasGutter ?? 0),
        ts, ts, (i % tw) * ts, ((i / tw) | 0) * ts, ts, ts);
      const ctxs = [];
      const texs = ca.phases.map((ph) => {
        const cv = document.createElement('canvas');
        cv.width = tw * ts; cv.height = th * ts;
        const ctx = cv.getContext('2d');
        ctx.imageSmoothingEnabled = false;
        ctxs.push(ctx);
        ph.tiles.forEach((tile, i) => drawTile(ctx, tile, i));
        const tx = Texture.from(cv);
        tx.source.scaleMode = 'nearest';
        return tx;
      });
      const sp = new Sprite(texs[0]);
      sp.x = ca.tx * ts;
      sp.y = ca.ty * ts;
      this.tileLayer.addChild(sp);
      this.anim.cellAnims.push({
        sprite: sp, texs,
        durs: ca.phases.map((p) => Math.max(1, p.frames)),
        idx: 0, acc: 0,
        paintTile: (tileId, atlasTile) => {
          let hit = false;
          ca.phases.forEach((ph, p) => ph.tiles.forEach((t, i) => {
            if (t === tileId) { drawTile(ctxs[p], atlasTile, i); hit = true; }
          }));
          if (hit) texs.forEach((t) => t.source.update());
        },
      });
    }
    // palette cycle (water/lava shimmer), an AnimRunner fx
    const cycle = setupCycle(level, this.tilemap);
    if (cycle) this.anim.fx.push(cycle);

    // collision overlay
    if (level.collision) {
      if (level.collision.kind === 'grid') {
        this.collisionLayer.addChild(buildCollisionGrid(level));
      } else if (level.collision.kind === 'profiles') {
        const shapes = await this.data.loadShapes(level.collision.shapesFile || 'shapes.json');
        this.collisionLayer.addChild(buildCollisionProfiles(level, shapes));
      }
    }

    // objects + spawn (fixed placements)
    const picks = this.config.objectInfo ? this.picks : undefined;
    const { container, animObjs, pathObjs } = await buildObjects(level, this.data, {
      markerCat: this.config.markerCat || (() => 'default'), picks,
    });
    for (let i = 0; i < copies; i++) {
      if (i === 0) { this.objectLayer.addChild(container); continue; }
      const { container: dup } = await buildObjects(level, this.data, {
        markerCat: this.config.markerCat || (() => 'default'), picks,
      });
      dup.x = i * this.levelW;
      this.objectLayer.addChild(dup);
    }
    this.anim.spriteSteps.push(...animObjs);
    this.anim.movePaths.push(...pathObjs);

    // randomized pools (rebuilt whenever the objects layer switches on)
    await this._rebuildPools();

    this.cam.fitView(level.view, {
      maxNativeFactor: this.config.maxNativeFactor ?? 4,
      minFitFactor: this.config.minFitFactor ?? 0.95,
    });
    if (this.config.hooks?.loaded) await this.config.hooks.loaded(this, level);
    return level;
  }

  async _rebuildPools() {
    // patrol records are pool-only: drop the stale ones before the re-roll
    // (pools rebuild on every objects-layer-on toggle, not just level loads)
    this.anim.patrols.length = 0;
    this.poolPicks = [];
    this._hideCard(); // its target node may be one of the pool placements we're dropping
    this.poolLayer.removeChildren();
    const level = this.level;
    if (!level || !(level.objectPools || []).length) return;
    // One seed per re-roll: every wrap copy replays the same rng stream, so
    // the three cylinder copies are identical — panning across the seam must
    // never show an object "teleporting". Without ?seed= a fresh random seed
    // is drawn per re-roll (placements still re-randomize per toggle).
    const fixed = typeof window !== 'undefined' ? window.__studioSeed : null;
    const seed = fixed != null && !Number.isNaN(fixed) ? fixed : (Math.random() * 0x7fffffff) | 0;
    const copies = this.meta.wrap === 'x' ? 3 : 1;
    const stampTex = (tileId) => this.tilemap.tileTexture(tileId);
    const picks = this.config.objectInfo ? this.poolPicks : undefined;
    for (let i = 0; i < copies; i++) {
      // group positions are container-local, so the copy offset composes
      const { container, patrols } = await buildPools(level, this.data,
        { random: rng(seed), stampTex, picks });
      container.x = i * this.levelW;
      this.poolLayer.addChild(container);
      this.anim.patrols.push(...patrols);
    }
  }

  setLayer(name, on) {
    if (name === 'collision') this.collisionLayer.visible = on;
    if (name === 'objects') {
      this.objectLayer.visible = on;
      this.poolLayer.visible = on;
      if (!on) this._hideCard(); // objects hidden -> drop any open card
      if (on && (this.level?.objectPools || []).length) this._rebuildPools();
    }
    if (name === 'animation') this.anim.setEnabled(on);
  }

  // --- object info cards (opt-in via config.objectInfo) ------------------------------
  // A click (not a drag) on a placed object or the spawn shows a small card naming it
  // and describing its behaviour — the 2-D analogue of the Super Mario 64 DS viewer's
  // actor cards. Listeners live on `this.el` (not the canvas) because the camera
  // setPointerCapture()s the pointer to `this.el` on pointerdown, so a captured
  // pointerup only fires on `this.el` and its ancestors, never a descendant canvas.
  _wireObjectPicking() {
    const el = this.el;
    let downAt = null;
    el.addEventListener('pointerdown', (e) => { downAt = { x: e.clientX, y: e.clientY }; });
    el.addEventListener('pointerup', (e) => {
      if (!downAt) return;
      const moved = Math.hypot(e.clientX - downAt.x, e.clientY - downAt.y);
      downAt = null;
      if (moved > 5) return; // a drag/pan, not a click
      this._pickAt(e.clientX, e.clientY);
    });
  }

  // Hit-test the visible object/pool placements under a client point; show the nearest
  // (smallest-area, i.e. most specific) match's card, or dismiss the card on a miss.
  _pickAt(clientX, clientY) {
    const p = this.cam.screenPt(clientX, clientY);
    const candidates = [];
    if (this.objectLayer.visible) candidates.push(...this.picks);
    if (this.poolLayer.visible) candidates.push(...this.poolPicks);
    let best = null, bestArea = Infinity;
    for (const pk of candidates) {
      const node = pk.node;
      if (!node || !node.parent || !node.visible) continue;
      const b = node.getBounds(); // pixi Bounds, in global (screen) px — matches screenPt
      if (p.x < b.minX || p.x > b.maxX || p.y < b.minY || p.y > b.maxY) continue;
      const area = (b.maxX - b.minX) * (b.maxY - b.minY);
      if (area < bestArea) { bestArea = area; best = pk; }
    }
    if (best) this._showCard(best, clientX, clientY);
    else this._hideCard();
  }

  _showCard(pick, clientX, clientY) {
    this._hideCard();
    const info = (this.config.objectInfo || {})[pick.name];
    const card = document.createElement('div');
    card.style.cssText = 'position:absolute;max-width:min(360px,74%);z-index:12;'
      + 'background:rgba(10,13,18,.94);border:1px solid #3a4a5c;border-radius:8px;'
      + 'padding:10px 12px 11px;font:12px/1.55 system-ui,sans-serif;color:#dfe6f0;'
      + 'box-shadow:0 8px 28px rgba(0,0,0,.5)';
    // Keep the card's own clicks off the camera/picker (they bubble to this.el).
    for (const ev of ['pointerdown', 'pointerup', 'wheel']) {
      card.addEventListener(ev, (e) => e.stopPropagation());
    }
    const close = document.createElement('button');
    close.textContent = '×';
    close.title = 'Close';
    close.style.cssText = 'position:absolute;top:3px;right:7px;border:0;background:none;'
      + 'color:#8895a8;font-size:18px;line-height:1;cursor:pointer;padding:2px';
    close.addEventListener('click', () => this._hideCard());
    const h = document.createElement('div');
    h.style.cssText = 'font-weight:600;margin:0 14px 4px 0;color:#ffd75e';
    h.textContent = info ? info.title : (pick.name === 'player' ? 'Player' : 'Object');
    // ID line: the object's raw identifiers (type / AI handler / sprite), always shown so
    // the user has a stable handle to ask about later — the point of these cards.
    const ids = idLine(pick);
    let idEl = null;
    if (ids) {
      idEl = document.createElement('div');
      idEl.style.cssText = 'font:11px/1.4 ui-monospace,Menlo,monospace;color:#7f8ea3;margin:0 0 6px';
      idEl.textContent = ids;
    }
    const body = document.createElement('div');
    if (info) { body.textContent = info.text; }
    else { body.style.color = '#9aa7b8'; body.textContent = 'No notes for this object yet — the id above is its handle.'; }
    card.append(close, h);
    if (idEl) card.append(idEl);
    card.append(body);

    // this.el is the studio's .mount (position:absolute, inset:0) — a positioning
    // context. Append first so we can measure, then clamp the card inside the viewport.
    this.el.appendChild(card);
    const rect = this.el.getBoundingClientRect();
    let x = clientX - rect.left + 14, y = clientY - rect.top + 14;
    x = Math.max(8, Math.min(x, rect.width - card.offsetWidth - 8));
    y = Math.max(8, Math.min(y, rect.height - card.offsetHeight - 8));
    card.style.left = x + 'px';
    card.style.top = y + 'px';
    this._card = card;
    clearTimeout(this._cardTimer);
    this._cardTimer = setTimeout(() => this._hideCard(), 15000);
  }

  _hideCard() {
    clearTimeout(this._cardTimer);
    if (this._card) { this._card.remove(); this._card = null; }
  }
}
