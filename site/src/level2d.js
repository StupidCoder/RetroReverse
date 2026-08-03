// level2d.js — the 2D tilemap engine (PixiJS): grid + block indirection,
// tile/cell animation, palette cycle + alternate-palette regions, placements
// of sprite2d objects, seeded pools, cylinder wrap, pan/zoom/pinch camera.
// Everything comes from the level document; there is no per-game code.

import { Application, Container, Graphics, Rectangle, Sprite, Texture } from 'pixi.js';
import { SpriteObject, loadImage } from './sprite2d.js';
import { mulberry32 } from './data.js';
import { PanInput } from './pancam.js';

const hexRgb = (h) => { const n = parseInt(h.slice(1), 16); return [(n >> 16) & 255, (n >> 8) & 255, n & 255]; };

export async function mount(ctx, doc) {
  const { stage, game, asset, params } = ctx;
  const tm = doc.tilemap;
  const docPath = asset.file;

  stage.classList.add('render2d');
  const app = new Application();
  // preserveDrawingBuffer: the screen-filter overlay captures this canvas
  // from its own RAF loop; without it the capture races pixi's render and
  // goes black whenever the filter loop runs first (filter enabled before
  // the view mounts).
  await app.init({ background: '#000000', resizeTo: stage, antialias: false, resolution: devicePixelRatio, autoDensity: true, preserveDrawingBuffer: true });
  app.canvas.classList.add('pixi', 'fill');
  stage.appendChild(app.canvas);

  const world = new Container();
  app.stage.addChild(world);

  const atlasImg = await loadImage(game.url(docPath, tm.atlas.file));
  const map = tm.blocks
    ? new BlockMap(tm, atlasImg, doc)
    : new TileMapPlain(tm, atlasImg);
  const copies = tm.wrap === 'x' ? 3 : 1; // cylinder: three map copies, camera wraps
  const period = map.widthPx;
  for (let i = 0; i < copies; i++) {
    const holder = new Container();
    holder.x = (i - 1) * period * (copies > 1 ? 1 : 0);
    holder.addChild(i === 0 ? map.container : map.mirror());
    world.addChild(holder);
  }

  // ---- animation clock (engine frames at tickHz) -------------------------------
  const tickHz = game.display.tickHz || 60;
  const anims = []; // objects with tick(dfFrames)
  app.ticker.add(() => {
    const df = (app.ticker.deltaMS * tickHz) / 1000;
    for (const a of anims) a.tick(df);
  });

  // tile animations: repaint tile ids through their frame lists
  for (const ta of tm.tileAnims || []) {
    let acc = 0, step = 0;
    anims.push({
      tick(df) {
        acc += df;
        const period = ta.periodFrames || 10;
        let changed = false;
        while (acc >= period) { acc -= period; step++; changed = true; }
        if (changed) {
          const fr = ta.frames[step % ta.frames.length];
          ta.tiles.forEach((t, i) => map.paintTile(t, fr[i]));
        }
      },
    });
  }

  // cell animators: tw×th strips at (tx,ty) with per-phase holds
  const objLayer = new Container();
  const cellAnimRecs = []; // kept for the AR live layer
  for (const ca of tm.cellAnims || []) {
    const ts = tm.tileSize;
    const canvases = [];
    const texs = ca.phases.map((ph) => {
      const cv = document.createElement('canvas');
      cv.width = ca.tw * ts; cv.height = ca.th * ts;
      const c2 = cv.getContext('2d');
      c2.imageSmoothingEnabled = false;
      ph.tiles.forEach((tile, i) => {
        const cols = tm.atlas.cols || 16;
        const cell = ts + 2 * (tm.atlas.gutter || 0);
        c2.drawImage(atlasImg,
          (tile % cols) * cell + (tm.atlas.gutter || 0),
          ((tile / cols) | 0) * cell + (tm.atlas.gutter || 0),
          ts, ts, (i % ca.tw) * ts, ((i / ca.tw) | 0) * ts, ts, ts);
      });
      const tex = Texture.from(cv);
      tex.source.scaleMode = 'nearest';
      canvases.push(cv);
      return tex;
    });
    const spr = new Sprite(texs[0]);
    spr.position.set(ca.tx * ts, ca.ty * ts);
    world.addChild(spr);
    let idx = 0, acc = 0;
    anims.push({
      tick(df) {
        acc += df;
        while (acc >= (ca.phases[idx].frames || 1)) {
          acc -= ca.phases[idx].frames || 1;
          idx = (idx + 1) % texs.length;
          spr.texture = texs[idx];
        }
      },
    });
    cellAnimRecs.push({
      spr, canvases,
      x: ca.tx * ts, y: ca.ty * ts, w: ca.tw * ts, h: ca.th * ts,
      index: () => idx,
    });
  }

  // palette cycle (block maps)
  if (tm.paletteFx?.cycle && map.cycle) {
    const fx = map.cycle(tm.paletteFx);
    if (fx) anims.push(fx);
  }

  // ---- camera ---------------------------------------------------------------------
  // Fitted BEFORE the placements load: their loading awaits many fetches and
  // the level must never render unzoomed in the meantime.
  const cam = new MapCamera(stage, app, world, {
    bounds: () => ({ w: map.widthPx, h: map.heightPx }),
    wrapX: () => (tm.wrap === 'x' ? map.widthPx : 0),
  });
  const caps = doc.camera?.map2d || {};
  cam.fitView(tm.view, {
    maxNativeFactor: caps.maxNativeFactor || 4,
    minFitFactor: caps.minFitFactor || 0.95,
  });
  cam.wirePointer();

  // Arrows pan, +/- zooms — through the inertial integrator, so holding a key
  // accelerates into a smooth scroll and releasing it glides to a stop. The
  // drag handlers feed their throw into the same velocity (see wirePointer).
  const input = new PanInput({
    onPan: (dx, dy) => cam.panBy(dx, dy),
    onZoom: (f) => cam.zoomAtCenter(f),
  });
  cam.input = input;
  app.ticker.add(() => input.step(app.ticker.deltaMS / 1000));

  // ---- placements, pools, spawn --------------------------------------------------
  world.addChild(objLayer);
  // The playfield clips sprites/blits at its edges (Marble Madness's blit_object
  // narrows any draw crossing the 288-px right edge), so a path-animated piece
  // sweeping into the border must clip against the map, not overhang into the
  // void. Wrap-scrolling maps are exempt: their pieces may straddle the seam.
  if (tm.wrap !== 'x') {
    const clip = new Graphics().rect(0, 0, map.widthPx, map.heightPx).fill(0xffffff);
    world.addChild(clip);
    objLayer.mask = clip;
  }
  // Named placement layers: each declared layer gets its own container inside
  // objLayer (so the playfield clip still applies) and a display-panel toggle.
  // Placements naming no layer land in the first one.
  const layerNodes = new Map();
  for (const ly of tm.layers || []) {
    const node = new Container();
    node.visible = ly.visible !== false;
    objLayer.addChild(node);
    layerNodes.set(ly.id, node);
    ctx.displayPanel?.toggle(ly.name || ly.id, node.visible, (on) => { node.visible = on; });
  }
  const layerFor = (id) =>
    layerNodes.get(id) || (layerNodes.size ? layerNodes.values().next().value : objLayer);

  const objects = new Map(); // id -> SpriteObject (loaded on demand)
  const getObj = async (id) => {
    if (!objects.has(id)) objects.set(id, SpriteObject.load(game, id));
    return objects.get(id);
  };
  const pickables = [];

  const place = async (pl, extra = {}) => {
    try {
      const obj = await getObj(pl.object);
      const inst = obj.makeInstance({ anim: pl.anim || extra.anim, tint: pl.tint || extra.tint, hflip: pl.hflip });
      inst.node.position.set(pl.pos[0], pl.pos[1]);
      layerFor(pl.layer).addChild(inst.node);
      anims.push(inst);
      pickables.push({ inst, pl, obj });
      return inst;
    } catch (e) {
      console.warn('placement failed', pl, e);
      return null;
    }
  };

  for (const pl of doc.placements || []) await place(pl);

  // pools: count random candidates each mount; ?seed=N reproduces a roll
  const seed = parseInt(params.get?.('seed') ?? params.seed ?? '', 10);
  const rng = Number.isFinite(seed) ? mulberry32(seed) : Math.random;
  for (const pool of doc.pools || []) {
    const cand = [...pool.candidates];
    for (let i = cand.length - 1; i > 0; i--) {
      const j = (rng() * (i + 1)) | 0;
      [cand[i], cand[j]] = [cand[j], cand[i]];
    }
    for (const pos of cand.slice(0, pool.count)) {
      await place({ object: pool.object, pos, anim: pool.anim, tint: pool.tint,
        name: pool.name, info: pool.info, props: { pool: pool.id } });
    }
  }

  if (tm.spawn?.object) {
    await place({ object: tm.spawn.object, pos: [tm.spawn.x, tm.spawn.y],
      anim: tm.spawn.anim, tint: tm.spawn.tint, name: 'Player spawn' });
  }

  // ---- picking ----------------------------------------------------------------------
  const hud = document.createElement('div');
  hud.className = 'hud';
  hud.textContent = `${tm.width}×${tm.height} cells · drag or arrows to pan · wheel/pinch to zoom`;
  stage.appendChild(hud);
  let card = null;
  const closeCard = () => { card?.remove(); card = null; };
  let downAt = null;
  app.canvas.addEventListener('pointerdown', (e) => { downAt = { x: e.clientX, y: e.clientY }; });
  app.canvas.addEventListener('pointerup', (e) => {
    if (!downAt || Math.hypot(e.clientX - downAt.x, e.clientY - downAt.y) > 6) return;
    const sp = cam.screenPt(e.clientX, e.clientY);
    const wx0 = (sp.x - world.position.x) / cam.zoom;
    const wy = (sp.y - world.position.y) / cam.zoom;
    const wx = tm.wrap === 'x' ? ((wx0 % map.widthPx) + map.widthPx) % map.widthPx : wx0;
    let hit = null;
    for (const p of pickables) {
      const s = p.inst.sprite;
      // The box tracks the sprite where it is DRAWN: path-animated objects
      // (the swinging platform) offset their sprite from the placement, and
      // the click must follow the visual, not the anchor.
      const x0 = p.pl.pos[0] + s.position.x - s.pivot.x;
      const y0 = p.pl.pos[1] + s.position.y - s.pivot.y;
      if (wx >= x0 && wx <= x0 + p.obj.cellW && wy >= y0 && wy <= y0 + p.obj.cellH) { hit = p; break; }
    }
    closeCard();
    if (!hit) return;
    if (hit.pl.onClick?.action === 'text') {
      showCard(e, hit.pl.onClick.title || hit.pl.name, hit.pl.onClick.body, hit);
      return;
    }
    if (hit.pl.onClick?.action === 'animate') {
      const target = hit; // 2D targets are self in every current game
      target.inst.setAnim(hit.pl.onClick.clip);
      return;
    }
    showCard(e, null, null, hit);
  });

  function showCard(e, title, body, p) {
    closeCard();
    card = document.createElement('div');
    card.className = 'infocard';
    const name = title || p.pl.info?.title || p.pl.name || p.obj.asset.name;
    const desc = body || p.pl.info?.body || p.obj.asset.description || '';
    const props = p.pl.props ? Object.entries(p.pl.props).map(([k, v]) => `${k} ${JSON.stringify(v)}`).join(' · ') : '';
    card.innerHTML = `<span class="x">×</span><b></b><div class="dim"></div>
      <div class="mono">${props}</div><a>Open object →</a>`;
    card.querySelector('b').textContent = name;
    card.querySelector('.dim').textContent = desc;
    card.querySelector('a').onclick = () => ctx.navigate(p.obj.asset.id);
    card.querySelector('.x').onclick = closeCard;
    stage.appendChild(card);
    placeCard(card, stage, e);
  }

  // placeCard positions an infocard near the click but always fully inside
  // the stage: measure the real card (its height depends on the text), then
  // clamp both axes.
  function placeCard(c, host, e) {
    const r = host.getBoundingClientRect();
    const x = e.clientX - r.left + 10, y = e.clientY - r.top + 10;
    c.style.left = `${Math.max(8, Math.min(x, r.width - c.offsetWidth - 8))}px`;
    c.style.top = `${Math.max(8, Math.min(y, r.height - c.offsetHeight - 8))}px`;
  }

  // Cylinder wrap: the visible window may span the seam, so each object hops
  // one period right whenever its base position falls off the left edge.
  if (tm.wrap === 'x') {
    app.ticker.add(() => {
      const P = map.widthPx;
      for (const p of pickables) {
        const base = p.pl.pos[0];
        const sx = base * cam.zoom + world.position.x;
        p.inst.node.x = sx < -64 * cam.zoom ? base + P : base;
      }
    });
  }

  // ---- AR: the WHOLE map hangs in the room -------------------------------------
  // Dynamic import: this drags three.js in, and a 2-D level should not pay for
  // that on load — MapAR builds nothing until the button is pressed.
  const { MapAR } = await import('./xrmap.js');
  const xrStatus = document.createElement('div');
  xrStatus.className = 'hud xr-status';
  xrStatus.hidden = true;
  stage.appendChild(xrStatus);
  const ar = new MapAR({
    el: stage,
    params,
    onStatus: (t) => { xrStatus.hidden = false; xrStatus.textContent = t; },
    size: () => ({ w: map.widthPx, h: map.heightPx }),
    live: liveModel({ tm, map, world, pickables, cellAnimRecs, anims, tickHz }),
    snap: {
      begin() {
        // Stop the clock: chunks are read one per frame, so a running tile /
        // cell / palette animation would freeze at a different phase in each
        // chunk and leave the difference visible on the chunk boundary.
        app.ticker.stop();
        // Undo the cylinder hop — objects must sit at their real map position,
        // not the one the visible window pushed them to.
        for (const p of pickables) p.inst.node.x = p.pl.pos[0];
      },
      read(x, y, w, h, resolution) {
        // The frame is in `world` local space. For a non-wrapping map the only
        // holder sits at x=0; for a cylinder the middle copy does, and that is
        // the one cellAnims and objLayer are aligned with.
        return app.renderer.extract.pixels({
          target: world, frame: new Rectangle(x, y, w, h), resolution,
        });
      },
      end() {
        cam.apply();
        app.ticker.start();
      },
    },
  });
  ctx.displayPanel?.section('AR');
  ctx.displayPanel?.toggle('Key out the backdrop', ar.keying, (on) => ar.setKeying(on));
  if (ar.hasLive) ctx.displayPanel?.toggle('Animate in AR', ar.live, (on) => ar.setLive(on));

  window.__rx = { app, world, objLayer, pickables, cam, map, ar }; // debug handle

  return {
    unmount() {
      input.dispose();
      ar.dispose();
      app.destroy(true, { children: true });
      closeCard();
      hud.remove();
      xrStatus.remove();
      stage.classList.remove('render2d');
    },
    xr: ar,
    sources: () => [app.canvas],
    // { cell, ox, oy, ref } — one game pixel in viewer px + the grid origin
    // (the shape screenfx._pixelCell consumes; it phase-locks the filter's
    // scanlines/LCD cells to the game pixels at any zoom/pan).
    pixelGrid: () => ({
      cell: cam.zoom, ox: world.position.x, oy: world.position.y,
      ref: app.screen.width,
    }),
    stats: () => ({
      Size: `${tm.width} × ${tm.height} cells (${map.widthPx} × ${map.heightPx} px)`,
      'Tile size': `${tm.tileSize} px${tm.blocks ? ` · ${tm.blocks.size}×${tm.blocks.size} blocks` : ''}`,
      Placements: (doc.placements || []).length,
      Pools: (doc.pools || []).map((p) => `${p.id} (${p.count} of ${p.candidates.length})`).join(', ') || '—',
      Wrap: tm.wrap || 'none',
    }),
  };
}

// ---- plain per-tile map --------------------------------------------------------

class TileMapPlain {
  constructor(tm, img) {
    this.tm = tm;
    this.img = img;
    this.container = new Container();
    this.baked = new Map(); // animated tile id -> {canvas, tex}
    const src = Texture.from(img).source;
    src.scaleMode = 'nearest';
    const ts = tm.tileSize, cols = tm.atlas.cols || 16, g = tm.atlas.gutter || 0;
    const cell = ts + 2 * g;
    const slice = [];
    const n = cols * Math.ceil(img.height / cell);
    for (let i = 0; i < n; i++) {
      slice.push(new Texture({ source: src, frame: new Rectangle((i % cols) * cell + g, ((i / cols) | 0) * cell + g, ts, ts) }));
    }
    const animated = new Set();
    for (const a of tm.tileAnims || []) for (const t of a.tiles) animated.add(t);
    const bake = (tile) => {
      const cv = document.createElement('canvas');
      cv.width = cv.height = ts;
      const tex = Texture.from(cv);
      tex.source.scaleMode = 'nearest';
      const rec = { cv, tex };
      this.baked.set(tile, rec);
      this._paint(rec, tile);
      return rec;
    };
    for (let r = 0; r < tm.height; r++) {
      for (let c = 0; c < tm.width; c++) {
        let v = tm.cells[r * tm.width + c];
        let flip = false;
        if (tm.hflipMask && (v & tm.hflipMask)) { v &= ~tm.hflipMask; flip = true; }
        const tex = animated.has(v) ? (this.baked.get(v) || bake(v)).tex : slice[v];
        if (!tex) continue;
        const s = new Sprite(tex);
        if (flip) { s.scale.x = -1; s.x = c * ts + ts; } else s.x = c * ts;
        s.y = r * ts;
        this.container.addChild(s);
      }
    }
    this.widthPx = tm.width * ts;
    this.heightPx = tm.height * ts;
  }

  _paint(rec, srcTile) {
    const tm = this.tm, ts = tm.tileSize, cols = tm.atlas.cols || 16, g = tm.atlas.gutter || 0;
    const cell = ts + 2 * g;
    const c2 = rec.cv.getContext('2d');
    c2.imageSmoothingEnabled = false;
    c2.clearRect(0, 0, ts, ts);
    c2.drawImage(this.img, (srcTile % cols) * cell + g, ((srcTile / cols) | 0) * cell + g, ts, ts, 0, 0, ts, ts);
    rec.tex.source.update();
  }

  paintTile(tileId, atlasTile) {
    const rec = this.baked.get(tileId);
    if (rec) this._paint(rec, atlasTile);
  }

  // dynamicSources lists the canvases that get repainted over time, paired
  // with the texture source that identifies them among the map's sprites.
  // `baked` holds exactly the animated tiles, so it is the whole answer here.
  dynamicSources() {
    return [...this.baked.values()].map((r) => ({ src: r.tex.source, canvas: r.cv }));
  }

  mirror() {
    // wrap copies share textures: rebuild sprite tree cheaply
    const c = new Container();
    for (const s of this.container.children) {
      const d = new Sprite(s.texture);
      d.position.copyFrom(s.position);
      d.scale.copyFrom(s.scale);
      c.addChild(d);
    }
    return c;
  }

  cycle() { return null; }
}

// ---- block-indirected map (Sonic) ------------------------------------------------

class BlockMap {
  constructor(tm, img, doc) {
    this.tm = tm;
    this.img = img;
    this.blocks = tm.blocks;
    this.blockPx = tm.tileSize * tm.blocks.size;
    this.container = new Container();
    this.tileOverride = new Map();
    this.blockCanvas = {};
    this.blockTex = {};
    this.regionTex = {};   // regionIdx -> { blockIdx -> tex } (alternate palettes)
    this.blocksWithTile = new Map();
    // kept for dynamicSources(): which tiles animate, and (filled by cycle())
    // which blocks the colour cycle rewrites
    this._animTiles = new Set();
    for (const a of tm.tileAnims || []) for (const t of a.tiles) this._animTiles.add(t);
    this._cycleBlocks = [];
    this.regions = (tm.paletteFx?.regions || []).map((r) => ({
      ...r,
      map: this._remap(tm.paletteFx.palette, r.palette),
    }));

    const B = this.blockPx;
    for (const idx of new Set(tm.cells)) {
      const mk = () => {
        const cv = document.createElement('canvas');
        cv.width = cv.height = B;
        const tex = Texture.from(cv);
        tex.source.autoGenerateMipmaps = true;
        tex.source.scaleMode = 'nearest';
        return { cv, tex };
      };
      const m = mk();
      this.blockCanvas[idx] = m.cv;
      this.blockTex[idx] = m.tex;
      this.regions.forEach((r, ri) => {
        (this.regionTex[ri] ||= {})[idx] = mk();
      });
      this.repaintBlock(idx);
      for (const t of this.blocks.tiles[idx]) {
        if (!this.blocksWithTile.has(t)) this.blocksWithTile.set(t, new Set());
        this.blocksWithTile.get(t).add(idx);
      }
    }
    for (let r = 0; r < tm.height; r++) {
      for (let c = 0; c < tm.width; c++) {
        const blk = tm.cells[r * tm.width + c];
        const cx = c * B + B / 2, cy = r * B + B / 2;
        const ri = this.regions.findIndex((rg) =>
          cx >= rg.rect.x && cx < rg.rect.x + rg.rect.w && cy >= rg.rect.y && cy < rg.rect.y + rg.rect.h);
        const tex = ri >= 0 ? this.regionTex[ri][blk].tex : this.blockTex[blk];
        if (!tex) continue;
        const s = new Sprite(tex);
        s.position.set(c * B, r * B);
        this.container.addChild(s);
      }
    }
    this.widthPx = tm.width * B;
    this.heightPx = tm.height * B;
  }

  _remap(from, to) {
    const m = new Map();
    for (let i = 0; i < Math.min(from?.length || 0, to.length); i++) {
      const f = hexRgb(from[i]), t = hexRgb(to[i]);
      m.set((f[0] << 16) | (f[1] << 8) | f[2], t);
    }
    return m;
  }

  _paint(ctx, idx) {
    ctx.imageSmoothingEnabled = false;
    const tm = this.tm, ts = tm.tileSize, n = this.blocks.size;
    const cols = tm.atlas.cols || 16, g = tm.atlas.gutter || 0;
    const cell = ts + 2 * g;
    this.blocks.tiles[idx].forEach((tile0, i) => {
      const tile = this.tileOverride.get(tile0) ?? tile0;
      ctx.drawImage(this.img, (tile % cols) * cell + g, ((tile / cols) | 0) * cell + g,
        ts, ts, (i % n) * ts, ((i / n) | 0) * ts, ts, ts);
    });
  }

  repaintBlock(idx) {
    const B = this.blockPx;
    this._paint(this.blockCanvas[idx].getContext('2d'), idx);
    this.blockTex[idx].source.update();
    this.regions.forEach((rg, ri) => {
      const rec = this.regionTex[ri][idx];
      const ctx = rec.cv.getContext('2d', { willReadFrequently: true });
      this._paint(ctx, idx);
      const im = ctx.getImageData(0, 0, B, B), d = im.data;
      for (let p = 0; p < d.length; p += 4) {
        const t = rg.map.get((d[p] << 16) | (d[p + 1] << 8) | d[p + 2]);
        if (t) { d[p] = t[0]; d[p + 1] = t[1]; d[p + 2] = t[2]; }
      }
      ctx.putImageData(im, 0, 0);
      rec.tex.source.update();
    });
  }

  paintTile(tileId, atlasTile) {
    this.tileOverride.set(tileId, atlasTile);
    for (const idx of this.blocksWithTile.get(tileId) || []) this.repaintBlock(idx);
  }

  // Which block canvases actually change: the blocks containing an animated
  // tile (repaintBlock touches the plain block AND its palette-region twins)
  // plus the blocks the colour cycle rewrites. Blocks that merely sit in a
  // palette region are static and must NOT be listed — that is why this is
  // enumerated rather than "every block".
  dynamicSources() {
    const idxs = new Set(this._cycleBlocks || []);
    for (const t of this._animTiles || []) for (const i of this.blocksWithTile.get(t) || []) idxs.add(i);
    const out = [];
    for (const idx of idxs) {
      if (this.blockTex[idx]) out.push({ src: this.blockTex[idx].source, canvas: this.blockCanvas[idx] });
      for (const ri of Object.keys(this.regionTex)) {
        const rec = this.regionTex[ri][idx];
        if (rec) out.push({ src: rec.tex.source, canvas: rec.cv });
      }
    }
    return out;
  }

  mirror() {
    const c = new Container();
    for (const s of this.container.children) {
      const d = new Sprite(s.texture);
      d.position.copyFrom(s.position);
      c.addChild(d);
    }
    return c;
  }

  // palette colour-cycle fx: remap the cycle's step-0 colours inside the tiles
  // that genuinely use a cycling slot.
  cycle(fx) {
    const cyc = fx.cycle;
    const steps = cyc.steps.map((s) => s.map(hexRgb));
    const from = steps[0];
    const cycleTiles = new Set(cyc.tiles || []);
    const B = this.blockPx, n = this.blocks.size, ts = this.tm.tileSize;
    const blocks = [];
    for (const idx of Object.keys(this.blockCanvas).map(Number)) {
      const cells = [];
      this.blocks.tiles[idx].forEach((t, cell) => { if (cycleTiles.has(t)) cells.push(cell); });
      if (!cells.length) continue;
      blocks.push({ idx, cells, clean: this.blockCanvas[idx].getContext('2d').getImageData(0, 0, B, B) });
    }
    if (!blocks.length) return null;
    this._cycleBlocks = blocks.map((b) => b.idx);
    const apply = (s) => {
      for (const b of blocks) {
        const ctx = this.blockCanvas[b.idx].getContext('2d');
        const out = ctx.createImageData(B, B);
        out.data.set(b.clean.data);
        const d = out.data;
        for (const cell of b.cells) {
          const cx = (cell % n) * ts, cy = ((cell / n) | 0) * ts;
          for (let y = 0; y < ts; y++) for (let x = 0; x < ts; x++) {
            const p = ((cy + y) * B + (cx + x)) * 4;
            for (let i = 0; i < from.length; i++) {
              const f = from[i];
              if (d[p] === f[0] && d[p + 1] === f[1] && d[p + 2] === f[2]) {
                const t = steps[s][i];
                d[p] = t[0]; d[p + 1] = t[1]; d[p + 2] = t[2];
                break;
              }
            }
          }
        }
        ctx.putImageData(out, 0, 0);
        this.blockTex[b.idx].source.update();
      }
    };
    let acc = 0, step = 0;
    return {
      tick: (df) => {
        acc += df;
        const period = cyc.periodFrames || 30;
        let changed = false;
        while (acc >= period) { acc -= period; step = (step + 1) % steps.length; changed = true; }
        if (changed) apply(step);
      },
    };
  }
}

// ---- 2D camera (port of shared/camera.js) -----------------------------------------

const ZOOM_STEP = Math.pow(1.15, 0.25);

class MapCamera {
  constructor(el, app, world, opts) {
    this.el = el;
    this.app = app;
    this.world = world;
    this.bounds = opts.bounds;
    this.wrapX = opts.wrapX || (() => 0);
    this.zoom = 1;
    this.minZoom = 0.05;
    this.maxZoom = 12;
  }

  fitView(view, { maxNativeFactor = 4, minFitFactor = 0.95 } = {}) {
    const W = this.app.screen.width, H = this.app.screen.height;
    const { w: lw, h: lh } = this.bounds();
    const v = view || { x: 0, y: 0, w: lw, h: lh };
    const z = H / v.h;
    const wrap = this.wrapX();
    const fitAll = wrap ? W / wrap : Math.min(W / lw, H / lh);
    this.minZoom = Math.min(fitAll * minFitFactor, z);
    this.maxZoom = Math.max((W / v.w) * maxNativeFactor, z);
    this.zoom = Math.max(this.minZoom, Math.min(this.maxZoom, z));
    this.panTo(v.x + v.w / 2, v.y + v.h / 2);
    this.apply();
  }

  panTo(wx, wy) {
    this.world.position.set(
      this.app.screen.width / 2 - wx * this.zoom,
      this.app.screen.height / 2 - wy * this.zoom,
    );
    this.apply();
  }

  // panBy scrolls the content and reports the delta it could actually apply:
  // an axis clamped at the edge of the map returns less than it was asked for,
  // which is how the inertial glide knows to stop. A wrapping axis is endless,
  // so it always applied the whole step.
  panBy(dx, dy) {
    if (!this.world.position) return { x: 0, y: 0 }; // destroyed mid-event
    const p = this.world.position;
    const x0 = p.x, y0 = p.y;
    p.x += dx;
    p.y += dy;
    this.clampPan();
    return { x: this.wrapX() ? dx : p.x - x0, y: p.y - y0 };
  }

  screenPt(cx, cy) {
    const r = this.el.getBoundingClientRect();
    return {
      x: (cx - r.left) * (this.app.screen.width / r.width),
      y: (cy - r.top) * (this.app.screen.height / r.height),
    };
  }

  zoomAt(px, py, f) {
    const wx = (px - this.world.position.x) / this.zoom;
    const wy = (py - this.world.position.y) / this.zoom;
    this.zoom = Math.min(this.maxZoom, Math.max(this.minZoom, this.zoom * f));
    this.world.position.set(px - wx * this.zoom, py - wy * this.zoom);
    this.apply();
  }

  zoomAtCenter(f) { this.zoomAt(this.app.screen.width / 2, this.app.screen.height / 2, f); }

  clampPan() {
    const sw = this.app.screen.width, sh = this.app.screen.height;
    const { w: lw, h: lh } = this.bounds();
    const zw = lw * this.zoom, zh = lh * this.zoom;
    let { x, y } = this.world.position;
    y = zh <= sh ? (sh - zh) / 2 : Math.min(0, Math.max(sh - zh, y));
    const period = this.wrapX() * this.zoom;
    if (period > 0) x = -(((-x % period) + period) % period);
    else x = zw <= sw ? (sw - zw) / 2 : Math.min(0, Math.max(sw - zw, x));
    this.world.position.set(x, y);
  }

  apply() {
    this.world.scale.set(this.zoom);
    this.clampPan();
  }

  wirePointer() {
    const c = this.el;
    const pts = new Map();
    let pinchDist = 0, pinchMid = null;
    c.addEventListener('pointerdown', (e) => {
      pts.set(e.pointerId, { x: e.clientX, y: e.clientY });
      this.input?.stop(); // grabbing the map arrests any glide in progress
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
        if (pinchMid) {
          this.world.position.x += mid.x - pinchMid.x;
          this.world.position.y += mid.y - pinchMid.y;
        }
        const sp = this.screenPt(mid.x, mid.y);
        this.zoomAt(sp.x, sp.y, pinchDist > 0 ? dist / pinchDist : 1);
        pinchDist = dist; pinchMid = mid;
      } else {
        this.panBy(dx, dy);
        this.input?.dragMove(dx, dy);
      }
    });
    const end = (e) => {
      pts.delete(e.pointerId);
      if (pts.size < 2) { pinchMid = null; pinchDist = 0; }
      if (pts.size === 0) {
        c.classList.remove('dragging');
        this.input?.dragEnd(); // a thrown map coasts on
      }
    };
    c.addEventListener('pointerup', end);
    c.addEventListener('pointercancel', end);
    c.addEventListener('wheel', (e) => {
      e.preventDefault();
      const sp = this.screenPt(e.clientX, e.clientY);
      this.zoomAt(sp.x, sp.y, e.deltaY < 0 ? ZOOM_STEP : 1 / ZOOM_STEP);
    }, { passive: false });
  }
}

// ---- AR live layer: what moves, described as plain data --------------------------
//
// The AR map is baked from a stopped clock, so the moving parts are drawn over
// it as three.js geometry instead. This side owns the PixiJS knowledge and
// hands xrlive.js a description with no three.js in it — and nothing here
// re-implements sprite2d.js. The SpriteInstances keep ticking; `read` just
// reports where they are.
//
// Returns null when a level has nothing that moves (Turrican pins one frame per
// placement, Minish Cap's markers are static), so those levels keep the still
// path exactly as it was.
function liveModel({ tm, map, world, pickables, cellAnimRecs, anims, tickHz }) {
  // A placement is live if its frame program has more than one step or it
  // carries a path. Turrican's `anim: "f9"` expands to steps [[9,1]] — one
  // step, no path — so its 226 placements stay in the bake where they belong.
  const visible = (node) => {
    for (let n = node; n && n !== world; n = n.parent) if (n.visible === false) return false;
    return true;
  };
  const liveSprites = pickables.filter(({ inst }) =>
    (inst.prog?.length > 1 || inst.anim?.path?.length) && visible(inst.node));

  const groups = new Map(); // atlas image + tint -> group
  for (const p of liveSprites) {
    const tint = p.inst.sprite.tint;
    const key = `${p.obj.img.src}|${tint}`;
    if (!groups.has(key)) {
      groups.set(key, {
        img: p.obj.img, cw: p.obj.cellW, ch: p.obj.cellH,
        tint: tint && tint !== 0xffffff ? tint : null,
        items: [],
      });
    }
    groups.get(key).items.push({
      // Read from the SpriteInstance's own fields rather than the Pixi display
      // objects: the display position also carries the cylinder-wrap hop, which
      // is a camera artefact and has no meaning on a map hanging in a room.
      read(o) {
        const inst = p.inst, s = inst.sprite;
        o.col = inst.prog[inst.idx][0];
        o.row = inst.anim.row || 0;
        o.flip = s.scale.x < 0;
        // A mirrored sprite is flipped about its own origin, so its left edge
        // moves to pos + pivot − cellW. (level2d's picking box at the top of
        // this file has the unflipped form only; nothing ships hflip today, so
        // it has never fired.)
        o.x = p.pl.pos[0] + s.position.x + (o.flip ? s.pivot.x - p.obj.cellW : -s.pivot.x);
        o.y = p.pl.pos[1] + s.position.y - s.pivot.y;
        o.hidden = false;
      },
    });
  }

  // Every tile animation, block repaint and palette cycle works by repainting a
  // canvas, so one rule covers all three: find the canvases, find the cells
  // drawing them, and let the texture upload do the rest. The dirty flag comes
  // free — pixi already calls source.update() at every repaint site.
  const surfaces = [];
  const dyn = new Map(map.dynamicSources ? map.dynamicSources().map((d) => [d.src, d]) : []);
  if (dyn.size) {
    const byCanvas = new Map();
    for (const s of map.container.children) {
      const d = dyn.get(s.texture?.source);
      if (!d) continue;
      if (!byCanvas.has(d.canvas)) byCanvas.set(d.canvas, { canvas: d.canvas, src: d.src, rects: [] });
      const f = s.texture.frame;
      const flip = s.scale.x < 0;
      byCanvas.get(d.canvas).rects.push({
        x: flip ? s.x - f.width : s.x, y: s.y, w: f.width, h: f.height, flip,
      });
    }
    for (const rec of byCanvas.values()) {
      if (!rec.rects.length) continue;
      let dirty = true; // upload once on the first frame
      rec.src.on('update', () => { dirty = true; });
      surfaces.push({ ...rec, dirty: () => dirty, clear: () => { dirty = false; } });
    }
  }

  const cells = (cellAnimRecs || []).map((c) => ({
    x: c.x, y: c.y, w: c.w, h: c.h, canvases: c.canvases, index: c.index,
  }));

  if (!groups.size && !surfaces.length && !cells.length) return null;

  return {
    tickHz,
    sprites: [...groups.values()],
    surfaces,
    cells,
    // Non-wrapping maps clip their object layer to the playfield; a cylinder
    // does not (level2d does the same).
    clip: () => (tm.wrap === 'x' ? null : { w: map.widthPx, h: map.heightPx }),
    // The XR frame loop drives this — the Pixi ticker is stopped for the whole
    // session, because window rAF is not reliably serviced while presenting.
    advance(dt) {
      const df = dt * (tickHz || 60);
      for (const a of anims) a.tick(df);
    },
    begin() {
      for (const p of liveSprites) p.inst.node.visible = false;
      for (const c of cellAnimRecs || []) c.spr.visible = false;
    },
    end() {
      for (const p of liveSprites) p.inst.node.visible = true;
      for (const c of cellAnimRecs || []) c.spr.visible = true;
    },
  };
}
