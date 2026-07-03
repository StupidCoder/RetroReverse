// Overlay layers for the shared 2-D level viewer: collision geometry, object
// placements (sprites with anchors/tints, marker fallbacks, the player spawn) and
// randomized object pools. See site/FORMAT.md.

import { Container, Graphics, Sprite, Text } from 'pixi.js';
import { cellTile } from './tilemap.js';

const MARKER_COLORS = {
  enemy: 0xff3838, item: 0xff9020, platform: 0x29d46e, boss: 0xc83cff,
  ctrl: 0xffe000, default: 0xaaaaaa,
};

// Tiny deterministic PRNG for ?seed= runs (mulberry32); falls back to Math.random.
export function rng(seed) {
  if (seed == null || Number.isNaN(seed)) return Math.random;
  let a = seed >>> 0;
  return () => {
    a |= 0; a = (a + 0x6D2B79F5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// --- collision ---------------------------------------------------------------------

// kind "grid": per-tile solidity, sub x sub cells each. Solid classes are drawn as
// run-length rects per sub-row; `legend` maps class bytes to colours.
export function buildCollisionGrid(level) {
  const g = level.grid;
  const coll = level.collision;
  const sub = coll.sub || 1;
  const ts = g.tileSize, cellPx = ts / sub;
  const W = g.width, H = g.height;
  const layer = new Container();
  const gr = new Graphics();
  const legend = coll.legend || {};
  const colorOf = (v) => {
    const c = legend[String(v)];
    return c ? parseInt(c.slice(1), 16) : 0xff3030;
  };
  const bw = W * sub, bh = H * sub;
  for (let br = 0; br < bh; br++) {
    const row = (br / sub) | 0, sr = br % sub;
    let runStart = 0, runVal = 0;
    for (let bc = 0; bc <= bw; bc++) {
      let v = 0;
      if (bc < bw) {
        const col = (bc / sub) | 0;
        const { tile, flip } = cellTile(g, g.cells[row * W + col]);
        const sc = flip ? sub - 1 - (bc % sub) : bc % sub;
        v = coll.solid[tile * sub * sub + sr * sub + sc] || 0;
      }
      if (v !== runVal) {
        if (runVal !== 0) {
          gr.rect(runStart * cellPx, br * cellPx, (bc - runStart) * cellPx, cellPx)
            .fill({ color: colorOf(runVal), alpha: 0.5 });
        }
        runStart = bc; runVal = v;
      }
    }
  }
  layer.addChild(gr);
  return layer;
}

// kind "profiles": per-column height profiles (Sonic), selected per block by
// blocks.shapes. Solid regions filled as merged red rects. (Used from M4 on.)
export function buildCollisionProfiles(level, shapes) {
  const g = level.grid;
  const blocks = level.blocks;
  const blockPx = g.tileSize * blocks.size;
  const layer = new Container();
  const gr = new Graphics();
  const { width: W, height: H, cells } = g;
  const profiles = shapes.profiles;
  for (let r = 0; r < H; r++) {
    for (let c = 0; c < W; c++) {
      const prof = profiles[blocks.shapes[cells[r * W + c]]];
      if (!prof) continue;
      const ox = c * blockPx, oy = r * blockPx;
      let runStart = -1, runTop = 0;
      const flush = (xEnd) => {
        if (runStart >= 0) gr.rect(ox + runStart, oy + runTop, xEnd - runStart, blockPx - runTop);
        runStart = -1;
      };
      for (let x = 0; x < blockPx; x++) {
        const h = prof[x];
        let top = null;
        if (h !== -128) { const s = Math.max(0, Math.min(blockPx, h)); if (s < blockPx) top = s; }
        if (top === null) flush(x);
        else if (runStart < 0) { runStart = x; runTop = top; }
        else if (top !== runTop) { flush(x); runStart = x; runTop = top; }
      }
      flush(blockPx);
    }
  }
  gr.fill({ color: 0xff2020, alpha: 0.8 });
  layer.addChild(gr);
  return layer;
}

// --- objects -----------------------------------------------------------------------

function marker(layer, x, y, w, h, color, label) {
  const g = new Graphics();
  g.rect(x, y, w, h).stroke({ width: 2, color });
  layer.addChild(g);
  if (label) {
    const t = new Text({ text: label, style: { fontFamily: 'monospace', fontSize: 9, fill: color } });
    t.x = x; t.y = y - 11;
    layer.addChild(t);
  }
}

function spawnCross(layer, x, y) {
  const r = 13;
  const g = new Graphics();
  g.circle(x, y, r).stroke({ width: 3, color: 0x33ff66 });
  g.moveTo(x - r - 7, y).lineTo(x + r + 7, y)
    .moveTo(x, y - r - 7).lineTo(x, y + r + 7).stroke({ width: 2, color: 0x33ff66 });
  g.circle(x, y, 2.5).fill(0x33ff66);
  layer.addChild(g);
}

// Place one sprite (frame 0) at its anchored position; returns records the anim
// runner picks up for stepped animation and movement paths.
function placeSprite(layer, rec, x, y, tint) {
  const entry = rec.entry;
  const [ax, ay] = entry.anchor || [0, 0];
  const s = new Sprite(rec.frames[0]);
  s.x = x - ax;
  s.y = y - ay;
  if (tint) s.tint = parseInt(tint.slice(1), 16);
  layer.addChild(s);
  const out = { sprite: s };
  if (entry.steps) {
    out.anim = {
      sprite: s,
      steps: entry.steps.map(([f, d]) => ({ tex: rec.frames[f], frames: Math.max(1, d) })),
      idx: 0, acc: 0,
    };
  } else if (rec.frames.length > 1) {
    out.anim = {
      sprite: s,
      steps: rec.frames.map((tex, i) => ({ tex, frames: Math.max(1, (entry.durations || [])[i] || 1) })),
      idx: 0, acc: 0,
    };
  }
  if (entry.path) out.path = { sprite: s, baseX: s.x, baseY: s.y, path: entry.path, t: 0 };
  return out;
}

// Build the object layer: fixed placements + spawn. Returns { container, animObjs,
// pathObjs } — the anim/path records feed the shared AnimRunner (M3+).
export async function buildObjects(level, data, { markerCat = () => 'default' } = {}) {
  const container = new Container();
  const animObjs = [], pathObjs = [];
  const put = async (o) => {
    const rec = o.sprite ? await data.sprite(o.sprite) : null;
    if (rec) {
      const placed = placeSprite(container, rec, o.x, o.y, o.tint);
      if (placed.anim) animObjs.push(placed.anim);
      if (placed.path) pathObjs.push(placed.path);
    } else {
      const cat = markerCat(o);
      const size = level.blocks ? level.grid.tileSize * level.blocks.size : level.grid.tileSize * 2;
      const bx = Math.floor(o.x / size) * size, by = Math.floor(o.y / size) * size;
      marker(container, bx, by, size, size,
        MARKER_COLORS[cat] || MARKER_COLORS.default,
        o.name || (o.type != null ? '?' + o.type.toString(16) : ''));
    }
  };
  for (const o of level.objects || []) await put(o);
  if (level.spawn) {
    const { x, y, sprite, tint } = level.spawn;
    const rec = sprite ? await data.sprite(sprite) : null;
    if (rec) {
      const placed = placeSprite(container, rec, x, y, tint);
      if (placed.anim) animObjs.push(placed.anim);
    } else {
      spawnCross(container, x, y);
    }
  }
  return { container, animObjs, pathObjs };
}

// Build randomized pool placements (Fort): `count` picks from `candidates`, each a
// random variant of tile stamps and/or a sprite. Rebuilt every time the objects
// layer is switched on. `stampTex(tileId)` supplies the tile textures. Returns
// { container, patrols } — the patrol records feed AnimRunner.patrols.
export async function buildPools(level, data, { random = Math.random, stampTex } = {}) {
  const container = new Container();
  const patrols = [];
  const stampInto = (parent, stamps) => {
    for (const st of stamps || []) {
      const tex = stampTex && stampTex(st.tile);
      if (!tex) continue;
      const s = new Sprite(tex);
      s.x = st.dx; s.y = st.dy;
      parent.addChild(s);
    }
  };
  for (const pool of level.objectPools || []) {
    const picks = shuffle([...(pool.candidates || [])], random).slice(0, pool.count);
    for (const [x, y, minDx = 0, maxDx = 0] of picks) {
      const v = pool.variants[(random() * pool.variants.length) | 0];
      // Each placement is one group at its candidate position; the patrol moves
      // the group. Plain stamps go straight into the group; dirStamps become one
      // hidden sub-container per phase per facing, phase 0 of the start facing shown.
      const group = new Container();
      group.x = x; group.y = y;
      container.addChild(group);
      stampInto(group, v.stamps);
      if (v.sprite) {
        const rec = await data.sprite(v.sprite);
        if (rec) placeSprite(group, rec, 0, 0, v.tint);
      }
      const p = pool.patrol;
      const dir0 = !p || p.start === 'random' ? (random() < 0.5 ? 1 : -1)
        : (p.start === 'left' ? -1 : 1);
      let phases = null;
      if (v.dirStamps) {
        phases = { 1: [], '-1': [] };
        for (const [dir, key] of [[1, 'right'], [-1, 'left']]) {
          for (const stamps of v.dirStamps[key] || []) {
            const ph = new Container();
            ph.visible = false;
            stampInto(ph, stamps);
            group.addChild(ph);
            phases[dir].push(ph);
          }
        }
        (phases[dir0][0] || {}).visible = true;
      }
      if (p) {
        patrols.push({
          group, baseX: x, minDx, maxDx,
          stepPx: p.stepPx || 8,
          stepFrames: p.stepFrames || 1,
          updatesPerStep: p.updatesPerStep || 1,
          dir0, dir: dir0, dx: 0, acc: 0, upd: 0, phase: 0, phases,
        });
      }
    }
  }
  return { container, patrols };
}

function shuffle(a, random) {
  for (let i = a.length - 1; i > 0; i--) {
    const j = (random() * (i + 1)) | 0;
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}
