// tileatlas.js — the CPU half of the AR tilemap renderer.
//
// Turns a Retro-X `tilemap` document plus its decoded atlas into the plain
// typed arrays the GPU wants, and nothing else: no three.js, no PixiJS, no
// DOM beyond an ImageData-shaped input. That is deliberate — this is where
// every decision that can be wrong lives (which colour is the backdrop, which
// tile goes in which slot, which cells are empty), so it has to be testable
// headlessly across all 145 shipped levels.
//
// The shape of the answer:
//
//   index   one texel per CELL — logical id (LE16), palette variant, flags
//   blocks  one texel per (sub-tile, block) — the block indirection, if any
//   slot    one texel per (tile, variant) — tile -> array slot. ANIMATION
//           WRITES HERE, and nowhere else: a tile animation step or a palette
//           cycle step is a handful of texels, not a repainted canvas.
//   tiles   the art itself, one tile per array layer (P x P packed when a
//           level has more tiles than the driver allows layers)
//   far     one texel per cell, the cell's average colour — the only correct
//           thing to sample below one texel per tile, where the right average
//           is over neighbouring MAP CELLS and no atlas scheme can produce it
//
// Colour transforms reproduce level2d.js exactly, in its order: the atlas tile
// (through any tile-animation override), then the palette-region remap for a
// region variant, then the colour cycle — which in the flat view rewrites only
// the BASE block canvases, never the region twins, so it does not apply inside
// a palette region here either.

const BRICK = 16; // cells per brick quad

export function buildTileModel(opts) {
  const { tm, atlas, key: keyOpt = 'auto', maxLayers = 2048 } = opts;
  const ts = tm.tileSize;
  const cols = tm.atlas.cols || 16;
  const gutter = tm.atlas.gutter || 0;
  const pitch = ts + 2 * gutter;
  const hmask = tm.hflipMask || 0;
  const W = tm.width, H = tm.height;
  const cells = tm.cells;
  const sub = tm.blocks ? tm.blocks.size : 1;
  const blockTiles = tm.blocks ? tm.blocks.tiles : null;

  // ---- who uses what -----------------------------------------------------------
  // Tile use counts weighted by cells, so the modal colour is the modal colour
  // of the MAP, not of the atlas: a tile drawn once must not outvote the sky.
  const use = new Map();
  const bump = (t, n) => use.set(t, (use.get(t) || 0) + n);
  for (const raw of cells) {
    const id = hmask ? raw & ~hmask : raw;
    if (blockTiles) { for (const t of blockTiles[id] || []) bump(t, 1); } else bump(id, 1);
  }

  const px = readTile(atlas, cols, pitch, gutter, ts);

  // ---- the backdrop colour ------------------------------------------------------
  let key = null;
  let keyStats = { pixelPct: 0, cellPct: 0 };
  if (keyOpt && keyOpt !== 'none') {
    key = keyOpt === 'auto' ? modalColour(px, use, ts) : keyOpt;
    if (key) keyStats.pixelPct = colourShare(px, use, ts, key);
  }

  // A tile is "blank" when every one of its texels is the key.
  const blank = new Set();
  if (key) {
    for (const t of use.keys()) if (isUniform(px(t), ts, key)) blank.add(t);
  }
  // Cell emptiness is baked into the index once, so it may only consider tiles
  // whose content can never change. A tile an animation drives is blank in
  // some frames and not others; flagging its cell empty would delete the
  // animation, and giving it no slot would leave the lookup pointing at slot 0
  // — which is how Fort Apocalypse drew its purple border where the map has
  // black. Those tiles get a (transparent) slot like any other and are
  // discarded per frame by the alpha test instead.
  const animated = new Set();
  for (const ta of tm.tileAnims || []) {
    for (const t of ta.tiles) animated.add(t);
    for (const fr of ta.frames) for (const t of fr) animated.add(t);
  }
  const alwaysBlank = new Set([...blank].filter((t) => !animated.has(t)));

  // ---- palette variants ----------------------------------------------------------
  const regions = (tm.paletteFx?.regions || []).map((r) => ({
    rect: r.rect, map: remap(tm.paletteFx.palette, r.palette),
  }));
  const numVariants = 1 + regions.length;

  const cyc = tm.paletteFx?.cycle || null;
  const cycSteps = cyc ? cyc.steps.map((s) => s.map(hexRgb)) : null;
  const cycTiles = new Set(cyc?.tiles || []);
  const numSteps = cycSteps ? cycSteps.length : 1;

  // Every atlas tile that can ever be shown: the ones the map uses, plus every
  // frame a tile animation can swap in.
  const shown = new Set(use.keys());
  for (const ta of tm.tileAnims || []) for (const fr of ta.frames) for (const t of fr) shown.add(t);

  // variantPixels applies the same transform the bake does, so "is this blank?"
  // and "what is baked here?" are answered by one piece of code.
  const variantPixels = (tile, variant, step) => {
    const src = px(tile);
    const out = new Uint8Array(src);
    const rmap = variant > 0 ? regions[variant - 1].map : null;
    for (let i = 0; i < ts * ts; i++) {
      const o = i * 4;
      let r = out[o], g = out[o + 1], b = out[o + 2];
      if (rmap) {
        const t2 = rmap.get((r << 16) | (g << 8) | b);
        if (t2) { [r, g, b] = t2; }
      } else if (step > 0) {
        const from = cycSteps[0];
        for (let c = 0; c < from.length; c++) {
          const f = from[c];
          if (r === f[0] && g === f[1] && b === f[2]) { [r, g, b] = cycSteps[step][c]; break; }
        }
      }
      out[o] = r; out[o + 1] = g; out[o + 2] = b;
    }
    return out;
  };
  // Blankness is PER VARIANT. Inside a palette region the backdrop colour is
  // remapped to something visible, so a tile that is pure sky in the base
  // palette is real art there — keying it away deleted a third of Sonic's
  // act10.
  const blankIn = new Map(); // "tile|variant" -> bool
  const isBlank = (t, v) => {
    const k = t + '|' + v;
    let b = blankIn.get(k);
    if (b === undefined) { b = !!key && isUniform(variantPixels(t, v, 0), ts, key); blankIn.set(k, b); }
    return b;
  };

  // ---- slot allocation -------------------------------------------------------------
  // One slot per (atlas tile, variant, cycle step) that can occur. The cycle
  // only ever touches variant 0, so the other variants get one step each.
  const slotOf = new Map(); // key string -> slot index
  const slotSrc = [];       // slot index -> {tile, variant, step}
  const slotKey = (t, v, s) => `${t}|${v}|${s}`;
  const alloc = (t, v, s) => {
    const k = slotKey(t, v, s);
    let i = slotOf.get(k);
    if (i === undefined) { i = slotSrc.length; slotOf.set(k, i); slotSrc.push({ tile: t, variant: v, step: s }); }
    return i;
  };
  // One shared, entirely transparent slot for every tile that is pure
  // backdrop. A blank tile still needs somewhere to point: a cell is only
  // flagged empty when ALL of it is blank, and a Sonic block mixes blank
  // sub-tiles with solid ones, so those sub-tiles are looked up like any
  // other. Leaving them unallocated is what made block 27 draw a stray tile —
  // an absent LUT entry reads as slot 0, which is some other tile's art.
  const blankSlot = alloc(-1, 0, 0);
  for (const t of shown) {
    for (let v = 0; v < numVariants; v++) {
      if (isBlank(t, v)) continue; // resolves to blankSlot
      const steps = v === 0 && cycTiles.has(t) ? numSteps : 1;
      for (let s = 0; s < steps; s++) alloc(t, v, s);
    }
  }

  // P x P tiles per array layer when there are more slots than layers.
  let P = 1;
  while (Math.ceil(slotSrc.length / (P * P)) > maxLayers && P < 16) P *= 2;
  const layers = Math.max(1, Math.ceil(slotSrc.length / (P * P)));
  const lw = ts * P;

  // ---- bake the tile art -----------------------------------------------------------
  const tiles = new Uint8Array(lw * lw * 4 * layers);
  for (let i = 0; i < slotSrc.length; i++) {
    const { tile, variant, step } = slotSrc[i];
    if (tile < 0) continue; // the shared blank slot: zeros are already correct
    const src = variantPixels(tile, variant, step);
    const lay = Math.floor(i / (P * P));
    const k = i - lay * P * P;
    const ox = (k % P) * ts, oy = Math.floor(k / P) * ts;
    for (let y = 0; y < ts; y++) {
      for (let x = 0; x < ts; x++) {
        const s0 = (y * ts + x) * 4;
        let r = src[s0], g = src[s0 + 1], b = src[s0 + 2], a = src[s0 + 3];
        // Keyed texels are (0,0,0,0), never (sky,0): the mip chain averages
        // premultiplied values, so edges never blend the backdrop back in.
        if (key && r === key.r && g === key.g && b === key.b) { r = g = b = a = 0; }
        const d0 = (lay * lw * lw + (oy + y) * lw + (ox + x)) * 4;
        tiles[d0] = r; tiles[d0 + 1] = g; tiles[d0 + 2] = b; tiles[d0 + 3] = a;
      }
    }
  }

  // ---- the slot LUT -----------------------------------------------------------------
  const numIds = Math.max(1, 1 + Math.max(...shown, 0));
  const lut = new Uint8Array(numIds * numVariants * 4);
  // A tile animation's step 0 shows frames[0], not the tile's own art — the
  // flat view repaints the tile canvas the moment the animation starts, so
  // "no override" is a transient that only exists before the first tick and
  // is not part of the cycle. Seeding it here is what makes frame 0 of the two
  // renderers the same frame.
  const initialOverrides = new Map();
  for (const ta of tm.tileAnims || []) {
    const fr = ta.frames[0] || [];
    ta.tiles.forEach((t, i) => { if (fr[i] !== undefined) initialOverrides.set(t, fr[i]); });
  }

  const writeSlots = (overrides = null, step = 0) => {
    for (let v = 0; v < numVariants; v++) {
      for (let t = 0; t < numIds; t++) {
        const a = overrides?.get(t) ?? t;
        const s = v === 0 && cycTiles.has(a) ? step % numSteps : 0;
        // Anything without art of its own resolves to the transparent slot,
        // never to slot 0 by omission.
        const slot = slotOf.get(slotKey(a, v, s)) ?? blankSlot;
        const o = (v * numIds + t) * 4;
        lut[o] = slot & 255;
        lut[o + 1] = (slot >> 8) & 255;
      }
    }
    return lut;
  };
  writeSlots(initialOverrides);

  // ---- the block table ---------------------------------------------------------------
  let blocks = null;
  if (blockTiles) {
    const nb = blockTiles.length, per = sub * sub;
    const data = new Uint8Array(per * nb * 4);
    for (let b = 0; b < nb; b++) {
      const row = blockTiles[b] || [];
      for (let i = 0; i < per; i++) {
        const t = row[i] || 0;
        const o = (b * per + i) * 4;
        data[o] = t & 255; data[o + 1] = (t >> 8) & 255;
      }
    }
    blocks = { data, w: per, h: nb };
  }

  // ---- the cell index -----------------------------------------------------------------
  const index = new Uint8Array(W * H * 4);
  const blankCell = (id, variant) => (blockTiles
    ? (blockTiles[id] || []).every((t) => isBlank(t, variant) && !animated.has(t))
    : isBlank(id, variant) && !animated.has(id));
  let empties = 0;
  const cellPx = ts * sub;
  for (let r = 0; r < H; r++) {
    for (let c = 0; c < W; c++) {
      const raw = cells[r * W + c];
      const flip = hmask ? (raw & hmask) !== 0 : false;
      const id = hmask ? raw & ~hmask : raw;
      // The palette region a cell belongs to is decided by its CENTRE, exactly
      // as level2d does when it picks the region texture.
      let variant = 0;
      if (regions.length) {
        const cx = c * cellPx + cellPx / 2, cy = r * cellPx + cellPx / 2;
        const ri = regions.findIndex((rg) => cx >= rg.rect.x && cx < rg.rect.x + rg.rect.w
          && cy >= rg.rect.y && cy < rg.rect.y + rg.rect.h);
        if (ri >= 0) variant = ri + 1;
      }
      const empty = key ? blankCell(id, variant) : false;
      if (empty) empties++;
      const o = (r * W + c) * 4;
      index[o] = id & 255;
      index[o + 1] = (id >> 8) & 255;
      index[o + 2] = variant;
      index[o + 3] = (flip ? 1 : 0) | (empty ? 128 : 0);
    }
  }
  keyStats.cellPct = (100 * empties) / (W * H);

  // ---- the far-LOD texture --------------------------------------------------------------
  // Averaged in LINEAR light and re-encoded, because that is what the sampler
  // does; averaging sRGB bytes directly would come out visibly dark.
  const far = new Uint8Array(W * H * 4);
  const avg = new Map(); // logical id + variant -> [r,g,b,a]
  for (let r = 0; r < H; r++) {
    for (let c = 0; c < W; c++) {
      const o = (r * W + c) * 4;
      const id = index[o] | (index[o + 1] << 8);
      const variant = index[o + 2];
      if (index[o + 3] & 128) continue; // blank: leave transparent
      const ak = `${id}|${variant}`;
      let m = avg.get(ak);
      if (!m) { m = cellAverage(id, variant); avg.set(ak, m); }
      far[o] = m[0]; far[o + 1] = m[1]; far[o + 2] = m[2]; far[o + 3] = m[3];
    }
  }

  function cellAverage(id, variant) {
    const list = blockTiles ? (blockTiles[id] || []) : [id];
    let lr = 0, lg = 0, lb = 0, la = 0, n = 0;
    for (const t of list) {
      const slot = slotOf.get(slotKey(t, variant, 0)) ?? slotOf.get(slotKey(t, 0, 0));
      if (slot === undefined) { n += ts * ts; continue; } // blank tile: transparent
      const lay = Math.floor(slot / (P * P));
      const k = slot - lay * P * P;
      const ox = (k % P) * ts, oy = Math.floor(k / P) * ts;
      for (let y = 0; y < ts; y++) {
        for (let x = 0; x < ts; x++) {
          const d0 = (lay * lw * lw + (oy + y) * lw + (ox + x)) * 4;
          const a = tiles[d0 + 3] / 255;
          lr += srgbToLinear(tiles[d0]) * a;
          lg += srgbToLinear(tiles[d0 + 1]) * a;
          lb += srgbToLinear(tiles[d0 + 2]) * a;
          la += a;
          n++;
        }
      }
    }
    if (!n || !la) return [0, 0, 0, 0];
    // Un-premultiply back to straight alpha for storage.
    return [
      linearToSrgb(lr / la), linearToSrgb(lg / la), linearToSrgb(lb / la),
      Math.round((la / n) * 255),
    ];
  }

  // ---- brick quads ------------------------------------------------------------------------
  // One quad per BRICK rather than per cell: the corpus max is 294,240 cells,
  // which as quads would be sub-pixel triangles at overview distance — the
  // worst thing a GPU can be asked to rasterise. Bricks that are entirely
  // blank are dropped outright.
  const bricks = [];
  for (let by = 0; by < H; by += BRICK) {
    for (let bx = 0; bx < W; bx += BRICK) {
      const bw = Math.min(BRICK, W - bx), bh = Math.min(BRICK, H - by);
      let any = false;
      for (let y = by; y < by + bh && !any; y++) {
        for (let x = bx; x < bx + bw; x++) {
          if (!(index[(y * W + x) * 4 + 3] & 128)) { any = true; break; }
        }
      }
      if (any) bricks.push({ x: bx, y: by, w: bw, h: bh });
    }
  }

  return {
    tm, initialOverrides,
    ts, sub, P, layers, lw,
    grid: { w: W, h: H },
    mapPx: { w: W * cellPx, h: H * cellPx },
    key, keyStats,
    index: { data: index, w: W, h: H },
    slot: { data: lut, w: numIds, h: numVariants },
    blocks,
    tiles: { data: tiles, size: lw, layers },
    far: { data: far, w: W, h: H },
    bricks,
    numVariants,
    cycle: cyc ? { steps: numSteps, periodFrames: cyc.periodFrames || 10 } : null,
    writeSlots,
    stats: {
      cells: W * H, bricks: bricks.length, slots: slotSrc.length, layers, P,
      tiles: shown.size, blankTiles: blank.size, emptyCells: empties,
      indexKB: Math.round(index.length / 1024), farKB: Math.round(far.length / 1024),
      tilesKB: Math.round(tiles.length / 1024),
    },
  };
}

// readTile returns a function giving one atlas tile's RGBA, cached.
function readTile(atlas, cols, pitch, gutter, ts) {
  const cache = new Map();
  const { data, width } = atlas;
  return (t) => {
    let v = cache.get(t);
    if (v) return v;
    v = new Uint8Array(ts * ts * 4);
    const tx = (t % cols) * pitch + gutter, ty = Math.floor(t / cols) * pitch + gutter;
    for (let y = 0; y < ts; y++) {
      for (let x = 0; x < ts; x++) {
        const s = ((ty + y) * width + (tx + x)) * 4;
        const d = (y * ts + x) * 4;
        v[d] = data[s]; v[d + 1] = data[s + 1]; v[d + 2] = data[s + 2];
        v[d + 3] = data[s + 3] === undefined ? 255 : data[s + 3];
      }
    }
    cache.set(t, v);
    return v;
  };
}

// modalColour is the most common colour across the RENDERED map: each tile's
// histogram weighted by how many cells draw it. Equal to sampling the finished
// picture, at O(cells + tiles*ts^2) instead of O(map pixels).
function modalColour(px, use, ts) {
  const hist = new Map();
  for (const [t, n] of use) {
    const src = px(t);
    for (let i = 0; i < ts * ts; i++) {
      const o = i * 4;
      if (src[o + 3] < 128) continue;
      const k = (src[o] << 16) | (src[o + 1] << 8) | src[o + 2];
      hist.set(k, (hist.get(k) || 0) + n);
    }
  }
  let best = -1, bestN = 0;
  for (const [k, n] of hist) if (n > bestN) { bestN = n; best = k; }
  if (best < 0) return null;
  return { r: (best >> 16) & 255, g: (best >> 8) & 255, b: best & 255 };
}

function colourShare(px, use, ts, key) {
  let hit = 0, tot = 0;
  for (const [t, n] of use) {
    const src = px(t);
    for (let i = 0; i < ts * ts; i++) {
      const o = i * 4;
      if (src[o + 3] < 128) continue;
      tot += n;
      if (src[o] === key.r && src[o + 1] === key.g && src[o + 2] === key.b) hit += n;
    }
  }
  return tot ? (100 * hit) / tot : 0;
}

function isUniform(src, ts, key) {
  for (let i = 0; i < ts * ts; i++) {
    const o = i * 4;
    if (src[o] !== key.r || src[o + 1] !== key.g || src[o + 2] !== key.b) return false;
  }
  return true;
}

function remap(from, to) {
  const m = new Map();
  for (let i = 0; i < Math.min(from?.length || 0, to.length); i++) {
    const f = hexRgb(from[i]), t = hexRgb(to[i]);
    m.set((f[0] << 16) | (f[1] << 8) | f[2], t);
  }
  return m;
}

const hexRgb = (h) => {
  const n = parseInt(h.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
};

const srgbToLinear = (v) => {
  const c = v / 255;
  return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
};
const linearToSrgb = (c) => {
  const v = c <= 0.0031308 ? c * 12.92 : 1.055 * c ** (1 / 2.4) - 0.055;
  return Math.max(0, Math.min(255, Math.round(v * 255)));
};
