# The Studio's common level format (format 1)

Every 2-D tilemap game (Sonic GG, Fort Apocalypse, Turrican, Marble Madness' map view,
Super Mario Land) exports the same JSON shapes into `site/public/<game>/`, consumed by
the shared viewer core in `site/src/shared/`. The format is a **superset**: a game uses
the sections its hardware/engine has, and omits the rest. All data is produced by each
game's Go exporter (`<Game>/extract/cmd/webexport`) straight from the reverse-engineered
formats — nothing is hand-authored.

Conventions: **all coordinates are world pixels** unless a field says otherwise; all
times are **engine frames** at the game's `tickHz` (50 = PAL C64/Amiga, 60 = GG/GB);
colours are `#rrggbb` strings; grids are row-major.

## meta.json

```jsonc
{
  "format": 1,
  "game": "sonic",
  "native": { "w": 160, "h": 144 },     // the machine's visible resolution
  "tickHz": 60,
  "wrap": "x",                          // optional: the map is a horizontal cylinder (Fort)
  "levels": [
    { "name": "Green Hills Act 1",
      "section": "Green Hills",         // optional accordion grouping in the Studio menu
      "file": "act01.json",
      "atlas": "atlas_0.png" }
  ]
}
```

## Level file

```jsonc
{
  "format": 1,
  "name": "Green Hills Act 1",

  // --- the tilemap -------------------------------------------------------------
  "grid": {
    "tileSize": 8,                      // px per tile (8 or 32)
    "atlas": "atlas_0.png",             // tile sheet PNG
    "atlasCols": 16,                    // tiles per atlas row
    "atlasGutter": 0,                   // 1 = each tile extruded by 1px (bleed guard)
    "width": 203, "height": 16,         // map size in CELLS
    "cells": [ ... ],                   // width*height cell values
    "hflipMask": 32768                  // optional: cell & mask = draw h-flipped (Turrican)
  },

  // Optional block indirection (Sonic): cells are indices into `tiles`, each block a
  // size x size arrangement of atlas tiles. Cell px size = tileSize * size.
  "blocks": { "size": 4, "tiles": [[16 tile ids], ...], "shapes": [shape id, ...] },

  // --- framing / player ----------------------------------------------------------
  "view":  { "x": 0, "y": 256, "w": 160, "h": 144 },  // initial camera framing rect
  "spawn": { "x": 160, "y": 368, "sprite": "0/00", "tint": "#b8c76f" },  // sprite/tint optional

  // --- objects --------------------------------------------------------------------
  "objects": [
    { "type": 8, "name": "crab", "x": 1344, "y": 401,   // engine REST position
      "sprite": "0/08", "tint": "#ffffff", "hard": true }  // sprite/tint/hard optional
  ],
  // Randomized placements (Fort): the viewer picks `count` candidates per pool each
  // time the objects layer is enabled (Math.random, or mulberry32(?seed=N)).
  "objectPools": [
    { "count": 8,
      "candidates": [[x, y], ...],
      "variants": [                       // one variant per placement, picked at random
        { "stamps": [{ "dx": 0, "dy": -8, "tile": 73 }, ...],   // atlas tiles stamped at offsets
          "sprite": "chopper-side", "tint": "#352879" }         // or/and a sprite
      ] }
  ],

  // --- animation -------------------------------------------------------------------
  // Tile animation: any map cell showing tiles[i] cycles through frames[step][i].
  "tileAnims": [
    { "tiles": [252, 253, 254, 255], "frames": [[...4 ids], ...], "periodFrames": 10 }
  ],
  // Placement-anchored cell animators (Sonic's $50 objects): a 2x4-tile strip at
  // (tx,ty) [8px tile coords] repainted through phases with per-phase hold times.
  "cellAnims": [
    { "tx": 36, "ty": 44, "phases": [{ "tiles": [8 ids], "frames": 240 }, ...] }
  ],

  // --- collision --------------------------------------------------------------------
  // kind "grid": per-TILE solidity, sub x sub cells each (sub 1 = whole tile).
  //   solid[tile*sub*sub + r*sub + c] = 0 (empty) or a class byte; `legend` maps class
  //   bytes to overlay colours (default red).
  // kind "profiles": per-column height profiles shared via shapesFile (Sonic):
  //   shapes.json = { count, profiles[count][32] (signed; -128 = none), angles[count] },
  //   selected per block by blocks.shapes[].
  "collision": { "kind": "grid", "sub": 4, "solid": [...], "legend": { "1": "#ff3030" } },

  // --- palette effects (Sonic only) --------------------------------------------------
  "paletteFx": {
    "palette": ["#...", ...],                          // the 16 BG colours as exported
    "cycle": { "steps": [["#..",..], ...], "periodFrames": 32, "tiles": [12, 13] },
    "waterLine": { "y": 416, "palette": ["#...", ...] } // raster split: below y use this
  },

  "music": "greenhills"                  // optional Studio track name
}
```

## sprites/index.json

One entry per sprite key. Sonic scopes keys by zone (`"0/08"`); other games use flat
names (`"chopper-fwd"`, `"tank"`, SML type ids).

```jsonc
{
  "0/08": {
    "src": "sprites/0/08.png",
    "frames": [[0, 0, 48, 48], [48, 0, 48, 48], ...],  // rects into src
    "anchor": [0, 0],                                   // draw at (x-ax, y-ay); default 0,0
    "durations": [13, 13, 13, 13],                      // per-frame hold (engine frames)
    "steps": [[0, 300], [1, 16], ...],                  // optional explicit play program
    "path": [[dx, dy], ...]                             // optional per-frame movement offsets
  }
}
```

- A sprite with one frame is static. With `durations` the frames play in order; with
  `steps` the explicit program plays (frame index + hold); both loop.
- `path` loops too: the object's draw position each engine frame is
  `(x, y) + path[t % len]` (moving platforms; sampled from the engines' own tables).

## What each game uses

| Section | Sonic | Fort | Turrican | Marble | SML |
|---|---|---|---|---|---|
| grid | ✓ (via blocks) | ✓ | ✓ + hflipMask | ✓ | ✓ |
| blocks | ✓ | – | – | – | – |
| tileAnims | ✓ | ✓ | – | – | – |
| cellAnims | ✓ | – | – | – | – |
| objects | ✓ | – | ✓ | – | ✓ |
| objectPools | – | ✓ | – | – | – |
| sprites index | ✓ (anims, paths) | ✓ (tinted) | ✓ | – | ✓ (anchored) |
| collision | profiles | – | grid sub 4 | – | grid sub 1 |
| paletteFx | ✓ | – | – | – | – |
| wrap | – | ✓ | – | – | – |

Marble's 3-D "Slopes" view (`<key>.slope.json`: height field + patrol-route markers) is
outside this format — it stays a per-game three.js view, like Elite and Stunt Car Racer.
