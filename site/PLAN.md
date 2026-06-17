# Website — implementation plan

An interactive companion site for the reverse-engineering work in this repo. It presents
findings across all four games with visualizations that markdown and static PNGs can't do:
a draggable, zoomable Sonic level viewer with toggleable layers (collision shapes, objects)
and live tile animation; later, 3D viewers for the Elite ship models and a Marble Madness
height map.

The centerpiece and first deliverable is the **Sonic level viewer**. The site shell is built
to host all four games from the start, but the other three get their pages later.

## Goals

- Render a Sonic act as an **actual tilemap** (individual tiles, not one baked PNG), at the
  Game Gear screen aspect ratio (160×144, 10:9), **drag to pan**, **wheel/pinch to zoom**
  from one screen down to a single pixel and out to the whole level.
- **Live tile animation** (rings, flowers) driven the way the game does it.
- A **collision layer** overlaying each block's height profile + solidity on the real map —
  the payoff of decoding both the terrain format (Part IV §4) and the collision shapes
  (Part V §2).
- An **object layer** (enemies/items/spawn) reusing the same layer machinery.
- Everything **static and reproducible**: all data is exported from the existing Go tools by
  re-running a command; the site has no backend.

## Stack & conventions (locked)

- **2D viewers → PixiJS** (`@pixi/tilemap` + `pixi-viewport`). **3D pages → three.js**
  (Elite wireframes, Marble terrain), added later. The two renderers share nothing at the
  engine level but share the data pipeline, site shell and UI conventions.
- **Static site**, built with **Vite**, deployed to **GitHub Pages** via a GitHub Actions
  workflow. (Project-pages base path must be set in `vite.config.js`.)
- Lives in a top-level **`site/`** (it spans all four games).
- **Exported assets are committed** under `site/public/<game>/…` (they're small — KBs — and
  deterministic), so the published site is purely static and the repo stays the source of
  truth. Re-running the exporter regenerates them.

## Why this is mostly presentation, not new RE

Everything the Sonic viewer needs is already decoded in the Go tools (`Sonic (GG)/extract`):
per-act block maps (`decomp.LoadMapRLE`), block→tile tables (bank 4 `$10000`), zone tilesets
(`decomp.Decompress`) + palettes (`romPalette`), animation sources (rings `$2F73D`, flowers
`$2FA3D`/`$2FABD`), object placements (`cmd/objprobe`), and the collision data (48 profiles at
`$3E7A`, per-zone block→shape at `$343D`, angles at `$3978`). The new work is an **exporter**
that serializes this and a **PixiJS frontend** that draws it.

## Repo layout

```
site/
  PLAN.md                 ← this file
  index.html              site shell / landing page
  package.json
  vite.config.js          base path = "/AIReverseEngineering/"
  src/
    main.js               shell + simple router (landing ↔ game pages)
    shared/
      data.js             fetch + parse meta/zone/act JSON and atlas textures
      viewport.js         pixi-viewport setup, GG-aspect framing, zoom clamps
      layers.js           layer container + visibility toggles
      controls.js         the control bar (checkboxes, act/zone selector)
      style.css
    sonic/
      app.js              wires the level viewer together
      tilemap.js          build the @pixi/tilemap mesh from a level
      collision.js        the collision overlay layer
      objects.js          object markers + spawn marker (reuse names/colours)
      anim.js             ring/flower animation ticker
    elite/                (later — three.js ship-model viewer)
    marble/               (later — three.js height map)
  public/
    sonic/
      meta.json           zone names, act index, animation timing
      shapes.json         the 48 collision profiles + angles (global)
      zone0.json … zone5.json + zone17.json   per-zone bundle (+ skybase act3 override)
      zone0.atlas.png …   per-zone tile atlas (256 tiles + anim frames)
      act01.json … act18.json                 per-act block map + objects + spawn
  .github/workflows/pages.yml   build with Vite, publish to Pages
```

The Go exporter is `Sonic (GG)/extract/cmd/webexport`, reusing the `cmd/levelmap` decode
path; it writes into `site/public/sonic/`.

## Data formats

Block-indexed throughout, so payloads are tiny (a 256×16 level is ~4 KB of block bytes; the
atlas is a 128×128-ish PNG). The client expands blocks → tiles into the tilemap mesh.

**`meta.json`**
```jsonc
{
  "zones": ["Green Hills","Bridge","Jungle","Labyrinth","Scrap Brain","Sky Base"],
  "acts": [ { "file": "act01.json", "zone": 0, "name": "Green Hills Act 1" }, … 18 ],
  "anim": { "framesPerTick": 10 }   // GG anim cadence
}
```

**`zone<N>.json`** (shared by the zone's 3 acts; Sky Base Act 3 gets its own as `zone17`)
```jsonc
{
  "zone": 0,
  "atlas": "zone0.atlas.png",
  "tileCount": 256,
  "tileSize": 8,
  "blockTiles": [ [t0..t15], … ],   // block → its 4×4 = 16 tile indices (row-major)
  "blockShape": [ s0, s1, … ],      // block → collision shape 0–47 (from $343D, bits 0-5)
  "palette": ["#rrggbb", … 16],
  "anim": {
    "rings":   { "tiles": [252,253,254,255], "frames": 6, "atlasRow": 16 },
    "flowers": { "tiles": [12,13,14,15],    "frames": 2, "atlasRow": 22 }
  }
}
```

**`act<N>.json`**
```jsonc
{
  "zone": 0, "act": 1, "name": "Green Hills Act 1",
  "stride": 256,            // $D232-derived column count
  "widthBlocks": 203,       // clipped to camera bound + one screen
  "heightBlocks": 16,       // 4096 / stride
  "order": "column",        // block-map layout (column-major, as the game streams it)
  "blocks": [ … ],          // block-index map, length widthBlocks*heightBlocks, 1 byte each
  "spawn": { "bx": 5, "by": 8 },
  "objects": [ { "type": 8, "bx": 18, "by": 12, "name": "crab" }, … ]
}
```

**`shapes.json`** (global — the bank-0 tables)
```jsonc
{
  "count": 48,
  "profiles": [ [h0..h31], … ],   // 48 × 32 signed heights; null column = no surface ($80)
  "angles":   [ a0, … ]           // 48 signed angles ($3978)
}
```

**`zone<N>.atlas.png`** — the 256 tiles rendered RGBA at the zone palette in a 16×16 grid of
8×8 (128×128), with the ring (6) and flower (2) animation frames appended in extra rows.

## The Sonic viewer (PixiJS)

- **Base map** — `@pixi/tilemap` `CompositeTilemap`: one mesh referencing the atlas, built by
  expanding `blocks` → `blockTiles` → tile placements. Few draw calls regardless of level
  size; this is what makes whole-map zoom cheap.
- **Camera** — `pixi-viewport`: `drag()`, `wheel()`, `pinch()`, `decelerate()`,
  `clamp({direction:'all'})` to the level, `clampZoom({minScale, maxScale})` from
  "whole level fits the GG frame" to ~8× pixel. The viewport DOM element is locked to 10:9;
  default zoom = one 160×144 screen at integer scale.
- **Layers** (`Container.visible`, driven by the control bar):
  1. base tilemap,
  2. **collision overlay**,
  3. **object markers**.
- **Animation** — a `Ticker` advancing a frame counter at the GG cadence; the ring/flower
  tile placements switch atlas frames. Toggleable (freeze on first frame).

### Collision overlay (`collision.js`)

Per block (32×32 px, aligned 1:1 with the block grid): look up `blockShape[block]` →
`shapes.profiles[shape]`, draw the surface polyline at `blockTop + height[col]` across the 32
columns (clipped to the cell, semi-transparent); **tint non-solid** shapes (`$00`, `$18`–`$26`)
a faint colour; optional colour-by-angle from `shapes.angles`. Drawn into a `Graphics`/mesh
over the visible region. This is the view that shows the collision *on the real tiles* — e.g.
the non-solid blocks that drop Sonic into the Green Hills Act 2 cave.

### Object layer (`objects.js`)

Markers at `objects[*].bx/by`, coloured/labelled by the type names already established
(crab, beetle, spring, boss, checkpoint, …), plus the 2×4-tile spawn marker. Reuses the
overlay/label machinery (same as `cmd/levelmap`'s overlays, now interactive).

## Site shell

A landing page introducing the project and the four games, linking to each game's page. The
Sonic page is the level viewer (zone/act selector + the control bar). Elite/Marble/Fort start
as stubs describing what's coming, filled in later.

## Build & deploy

- `npm create vite`, add `pixi.js`, `@pixi/tilemap`, `pixi-viewport`.
- `vite.config.js` `base: '/AIReverseEngineering/'` for project Pages.
- `.github/workflows/pages.yml`: on push to `main`, `npm ci && npm run build`, upload `dist/`,
  deploy to Pages.

## Verification

- The exporter's atlas + expanded block map must match the existing `cmd/levelmap` render
  pixel-for-pixel (reuse that as ground truth — it's already validated against the oracle).
- `shapes.json` must round-trip the `rendered/block_collision_profiles.png` figure.
- Spot-check a few acts in the viewer against the committed `rendered/level_*_objects.png`.

## Milestones

- **M0 — scaffold.** `site/` Vite project, shell page, Pages workflow, empty Sonic page.
- **M1 — exporter (Green Hills).** `cmd/webexport` emits `meta`/`shapes`/`zone0`/`act01-03`
  + `zone0.atlas.png`; verify vs `cmd/levelmap`.
- **M2 — viewer MVP.** Load one act; render the tilemap; pan + zoom in the GG frame.
- **M3 — animation.** Rings + flowers cycling at the GG cadence; freeze toggle.
- **M4 — collision overlay.** Profiles + non-solid tint + angle colour; toggle.
- **M5 — object layer.** Markers, labels, spawn; toggle.
- **M6 — all acts.** Export 18 acts (zone bundles + Sky Base Act 3 override); zone/act
  selector; handle stride/aspect variations (vertical levels).
- **M7 — shell polish.** Landing page, navigation, stubs for Elite/Marble/Fort.
- **Later — 3D.** three.js Elite ship-model viewer; Marble height-map terrain — reusing the
  data pipeline + shell + UI conventions.

## Open / deferred

- Exact GG animation timing per zone (rings are global; other zones may animate water/other —
  only rings + Green Hills flowers are decoded so far).
- Whether to add a minimap for very large/vertical levels (nice-to-have, post-M6).
- A "what is this?" info panel surfacing the relevant Sonic.md section per layer.
