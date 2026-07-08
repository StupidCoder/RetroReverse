# The Studio's common asset format (format 2)

Format 2 is the single asset contract every game exports into `site/public/<slug>/`,
consumed by the shared viewer core in `site/src/shared/`. It supersedes the earlier format-1 spec: the 2-D tilemap body is unchanged, now carried inside a common
envelope that also admits 3-D levels, and every asset kind (bitmaps, music, sfx, models,
levels) is indexed by one `manifest.json`.

The format is a **superset**: a game emits the sections its hardware/engine has and omits
the rest. All data is produced by each game's Go exporter
(`games/<slug>/extract/cmd/webexport`) straight from the reverse-engineered formats —
nothing is hand-authored.

Conventions: 2-D coordinates are **world pixels**; 3-D coordinates are the game's own world
units. Times are **engine frames** at the game's `tickHz` (50 = PAL C64/Amiga, 60 = GG/GB/DS);
colours are `#rrggbb`; grids are row-major.

## manifest.json

The per-game index. Replaces the old `meta.json` + `models.json`.

```jsonc
{
  "format": 2,
  "game": "sonic-gg",
  "platform": "Game Gear",
  "native": { "w": 160, "h": 144 },     // the machine's visible resolution
  "tickHz": 60,
  "wrap": "x",                          // optional default wrap for levels (horizontal cylinder)

  "bitmaps": [ { "name": "Title", "file": "bitmaps/title.png" } ],
  "music":   [ { "name": "Green Hills", "file": "music/greenhills.mp3", "loop": true } ],
  "sfx":     [ { "name": "ring", "file": "sfx/ring.wav" } ],
  "models":  [ { "name": "Kart", "file": "models/kart.glb" } ],

  "levels": [
    { "name": "Green Hills Act 1",
      "section": "Green Hills",         // optional accordion grouping in the Studio menu
      "file": "levels/act01.json",
      "kind": "tilemap2d",              // or "mesh3d"
      "atlas": "levels/atlas_0.png",    // tilemap2d only
      "objects": "levels/act01.objects.json" }
  ],

  "views": [                            // escape hatch: bespoke per-game three.js views
    { "name": "Ship models", "file": "models/ships.json", "kind": "elite-wireframe" }
  ]
}
```

## Level file — common envelope

```jsonc
{
  "format": 2,
  "name": "Green Hills Act 1",
  "kind": "tilemap2d",                  // "tilemap2d" | "mesh3d"

  "extents": { "tileSize": 8, "width": 203, "height": 16 },   // tilemap2d: size in CELLS
  //         { "min": [x,y,z], "max": [x,y,z] }               // mesh3d: world-unit AABB

  "wrap": "x",                          // "none" | "x" | "xy"
  "view":  { "x": 0, "y": 256, "w": 160, "h": 144 },          // initial camera framing (2d)
  "spawn": { "x": 160, "y": 368, "sprite": "0/00", "tint": "#b8c76f" },
  //         { "pos": [x,y,z], "dir": [dx,dy,dz] }            // mesh3d spawn

  // exactly one body follows, per `kind`:
  "grid": { ... },   // tilemap2d — see below (identical to format 1)
  "mesh": { ... },   // mesh3d     — see below

  // shared, all optional:
  "objects": [ ... ],
  "objectPools": [ ... ],
  "objectsFile": "act01.objects.json",
  "tileAnims": [ ... ],
  "cellAnims": [ ... ],
  "collision": { ... },
  "paletteFx": { ... },
  "music": "greenhills"
}
```

## tilemap2d body — `grid`

Identical to format 1. `extents` carries the cell dimensions that format 1 kept inside
`grid.width`/`grid.height`; both are emitted for compatibility during migration.

```jsonc
"grid": {
  "tileSize": 8,                        // px per tile (8 or 32)
  "atlas": "atlas_0.png",               // tile sheet PNG; may carry a ?v=hash cache-buster
  "atlasCols": 16,                      // tiles per atlas row
  "atlasGutter": 0,                     // 1 = each tile extruded 1px (bleed guard)
  "width": 203, "height": 16,           // map size in CELLS (mirror of extents)
  "cells": [ ... ],                     // width*height cell values, row-major
  "hflipMask": 32768                    // optional: cell & mask => draw h-flipped
}
```

Optional block indirection (cells index into `blocks.tiles`, each block a `size x size`
arrangement; cell px = `tileSize * size`):

```jsonc
"blocks": { "size": 4, "tiles": [[16 tile ids], ...], "shapes": [shape id, ...] }
```

## mesh3d body — `mesh`

A self-contained textured mesh, or a reference to a GLB. Fields mirror the geometry a
first-person/3-D engine emits.

```jsonc
"mesh": {
  "glb": "models/level1.glb",           // option A: geometry lives in a GLB
  // option B: inline geometry (as Ultima Underworld's level export):
  "positions": [x,y,z, ...],            // flat vertex array
  "uvs": [u,v, ...],
  "groups": [ { "material": 0, "start": 0, "count": 1234 } ],  // draw ranges
  "textures": [ "data:image/png;base64,..." ]                  // indexed by group.material
}
```

Billboard/animated sprite entities that belong to the mesh view (doors, creatures) live in
`objects` / `objectsFile`, not in `mesh`.

**Single-sided materials.** A level GLB may mark any material single-sided (glTF
`material.doubleSided:false`), which three.js honours by back-face culling. The shared writer
`tools/lib/glb` emits this per triangle group via `WriteTrianglesMat` (`TriGroup.SingleSided`);
the default is double-sided (unchanged). Use it for one-way geometry — a dungeon **ceiling** is
authored single-sided so the camera sees into rooms from above but the ceiling still occludes
from below. This is a property of the GLB itself; no viewer flag is needed.

## objects (inline) and `<level>.objects.json`

`objects` is the lightweight inline placement list (as format 1). The full
machine-readable object database is the sibling `objectsFile`. Both share one object shape;
fields absent in a game's engine are omitted.

```jsonc
// inline, in the level file:
"objects": [
  { "type": 8, "name": "crab", "x": 1344, "y": 401,
    "sprite": "0/08", "tint": "#ffffff", "hard": true }
],

// full DB, <level>.objects.json:
{ "format": 2, "level": "Green Hills Act 1",
  "objects": [
    { "id": 0, "type": 8, "name": "crab",
      "pos": [1344, 401], "size": [16,16]?, "rot": [0,0,0]?,
      "model": "models/crab.glb"?, "actor": 33?, "hard": true,
      "props": { "sprite": "0/08", "tint": "#ffffff" } } ] }
```

`objectPools` (randomized placements with optional patrol metadata) is carried unchanged
from format 1 — see the objectPools section below.

**Click-to-inspect (object info cards).** Clicking an object/sprite shows an info card, the
same way in 2-D and 3-D. Its content comes from two sources, merged by the shared resolver
`site/src/shared/objinfo.js` → `resolveObjectInfo(obj, objectInfo)`:

- **Extracted** (per object, in `objects.json`): the object's own `name`, `type`, and
  `props.text` (its in-game words), or an optional pre-composed `info: { title?, body?,
  quote? }` — data the game itself supplies (e.g. Ultima Underworld's STRINGS.PAK names +
  writings).
- **Editorial** (per game): an `objectInfo` map `{ <key>: { title, text } }` of hand-written
  documentation, keyed by the object's `name` / `type` / `model` / `actor`. It lives in the
  game's viewer config, not the asset JSON, because it is authored prose, not extractable data.

3-D viewers wire the raycast/click plumbing with the shared `installPicker(stage, pickables,
{ objectInfo, resolve?, enabled? })`; the 2-D viewer keeps its own cursor-anchored card but
takes its content from the same `resolveObjectInfo`.

### 3-D object rendering — `model` and `sprite`

The shared object layer (`site/src/shared/renderers.js` → `placeObjects`) renders a 3-D level's
objects two ways, chosen per object:

- **`model`** — a `"models/x.glb"` path. The GLB is loaded and placed at `pos` (and `rot`, in
  **radians**, when present).
- **`sprite`** — a first-class **directional-billboard sprite spec** (below). It becomes a
  textured camera-facing quad placed at `pos`, updated every frame. This is the shared
  replacement for per-game billboard code (Doom-style 8-view things, Ultima Underworld creatures);
  a plain always-facing billboard is just `views: 1`.

```jsonc
"sprite": {
  "billboard": "camera",   // "camera" (full facing) | "yaw" (about world-up only) | "none"; default "camera"
  "views": 8,              // directional angle buckets; 1 = a plain non-directional billboard
  "heading": 0,            // object facing in RADIANS (base angle for view bucket 0); ignored when views==1
  "size": [1.0, 1.5],      // world-unit quad [w, h]
  "anchor": "bottom",      // what pos means: "center" (default) | "bottom" (feet at pos — standing things)
  "sheet": "sprites/creatures.png",  // atlas URL, resolved by the viewer as base + sheet
  "frames": [[0,0,32,48], ...],      // atlas rects in PIXELS; length == views*perView,
                                     //   laid out as `views` blocks of `perView` frames
  "perView": 1,            // animation frames per view
  "fps": 8,                // animation rate (0 = static: first frame of each view)
  "blend": "alpha"         // "opaque" | "alpha" | "additive"; default "opaque"
}
```

The view bucket is `quantize(angleToCamera - heading, views)` and the animation frame is
`floor(t*fps) % perView`, so `frames[bucket*perView + frame]` is drawn.

**`blend` values** (all unlit, `NearestFilter` for retro crispness):
- `opaque` — solid quad; transparent border pixels still cut out (`alphaTest`).
- `alpha` — alpha-masked cut-out; `depthWrite:false`.
- `additive` — `THREE.AdditiveBlending`, `depthWrite:false` (translucent bodies: ghosts, fire,
  wisps — the game brightens the background rather than drawing a colour).

## objectPools (randomized 2-D placements)

Randomized placements (Fort): the viewer picks `count` candidates per pool each time the
objects layer is enabled (`Math.random`, or `mulberry32(?seed=N)`).

```jsonc
"objectPools": [
  { "count": 8,
    "name": "prisoner",                 // optional object kind; keys config.objectInfo[name] on click
    "candidates": [[x, y], ...],        // [x, y, minDx, maxDx] when the pool patrols
    // Back-and-forth patrol (Fort's tanks/prisoners/mines): every `stepFrames` engine
    // frames one update fires; it advances the dirStamps phase, and every
    // `updatesPerStep`-th update (default 1) moves the placement `stepPx` in its current
    // direction. A move leaving the candidate's inclusive [minDx,maxDx] span flips the
    // direction (and facing) and moves the other way in the same update. minDx==maxDx==0
    // stands still. With `slide` the phase is the object's SUB-CELL POSITION: it walks up
    // with the direction and down after a reversal, the stepPx move commits on wrap, and a
    // span edge turns the walk around in place. `startPhase` is the rest phase. Spans are
    // precomputed against the static map; mover-vs-mover reversal isn't modelled.
    "patrol": { "stepPx": 8, "stepFrames": 8, "updatesPerStep": 1,
                "slide": false, "startPhase": 0, "start": "random" }, // "random"|"right"|"left"
    "variants": [                       // one variant per placement, picked at random
      { "stamps": [{ "dx": 0, "dy": -8, "tile": 73 }, ...],   // atlas tiles stamped at offsets
        "sprite": "chopper-side", "tint": "#352879" },        // and/or a sprite
      // Direction-dependent art: one stamps array per animation phase per facing
      // (right.length == left.length). Facing swaps when the patrol reverses; anim-off
      // shows right[0] (or left[0] if the placement started leftward).
      { "dirStamps": { "right": [[{ "dx": 0, "dy": 0, "tile": 91 }, ...], ...],
                       "left":  [[...], ...] } }
    ] }
]
```

## Animation, collision, palette effects (tilemap2d)

```jsonc
// Tile animation: any map cell showing tiles[i] cycles through frames[step][i].
"tileAnims": [ { "tiles": [252,253,254,255], "frames": [[...4 ids], ...], "periodFrames": 10 } ],

// Placement-anchored cell animators: a tw×th tile strip at (tx,ty) [8px tile coords]
// repainted through phases with per-phase hold times. tw/th default 2×4 (Sonic's $50
// objects); Marble's Ultimate screen-swap uses 36×30.
"cellAnims": [ { "tx": 36, "ty": 44, "tw": 2, "th": 4,
                 "phases": [ { "tiles": [tw*th ids], "frames": 240 }, ... ] } ],

// collision kind "grid": per-TILE solidity, sub×sub cells each (sub 1 = whole tile).
//   solid[tile*sub*sub + r*sub + c] = 0 (empty) or a class byte; legend maps class bytes to
//   overlay colours (default red).
// collision kind "profiles": per-column height profiles shared via shapesFile (Sonic):
//   shapes.json = { count, profiles[count][32] (signed; -128 = none), angles[count] },
//   selected per block by blocks.shapes[].
"collision": { "kind": "grid", "sub": 4, "solid": [...], "legend": { "1": "#ff3030" } },

// palette effects (Sonic): BG palette + colour-cycle steps + a water raster split.
"paletteFx": {
  "palette": ["#...", ...],
  "cycle": { "steps": [["#..",..], ...], "periodFrames": 32, "tiles": [12,13] },
  "waterLine": { "y": 416, "palette": ["#...", ...] }
}
```

## sprites/index.json

One entry per sprite key. Sonic scopes keys by zone (`"0/08"`); other games use flat names
(`"chopper-fwd"`, SML type ids).

```jsonc
{ "0/08": {
    "src": "sprites/0/08.png",
    "frames": [[0,0,48,48], [48,0,48,48], ...],  // rects into src
    "anchor": [0, 0],                             // draw at (x-ax, y-ay); default 0,0
    "durations": [13,13,13,13],                   // per-frame hold (engine frames)
    "steps": [[0,300],[1,16], ...],               // optional explicit play program
    "path": [[dx,dy], ...] } }                     // optional per-frame movement offsets
```

A one-frame sprite is static. With `durations` the frames play in order; with `steps` the
explicit program (frame index + hold) plays; both loop. `path` loops too: the draw position
each engine frame is `(x,y) + path[t % len]` (moving platforms).

## What each game emits

Every game emits `manifest.json`, then fills only the asset kinds and level sections its
engine has. Per-section support for the tilemap2d games:

| Section | Sonic | Fort | Turrican | Marble | SML |
|---|---|---|---|---|---|
| grid | ✓ (via blocks) | ✓ | ✓ + hflipMask | ✓ | ✓ |
| blocks | ✓ | – | – | – | – |
| tileAnims | ✓ | ✓ | – | ✓ | – |
| cellAnims | ✓ | – | – | ✓ | – |
| objects | ✓ | – | ✓ | ✓ | ✓ |
| objectPools | – | ✓ | – | – | – |
| sprites index | ✓ | ✓ | ✓ | ✓ | ✓ |
| collision | profiles | – | grid sub 4 | – | grid sub 1 |
| paletteFx | ✓ | – | – | – | – |
| wrap | – | ✓ | – | – | – |

The mesh3d games (Ultima Underworld, Super Mario 64 DS, Mario Kart DS) emit `mesh3d` levels
+ `objectsFile` + `models/` GLBs. Bespoke per-game 3-D views (Elite wireframes, Stunt Car
tracks, Marble slopes) are `manifest.views[]` / `models[]` with a game-specific `kind`,
rendered by that game's renderer (see STANDARDS §4.4).
