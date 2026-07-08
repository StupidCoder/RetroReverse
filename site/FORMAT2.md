# The Studio's common asset format (format 2)

Format 2 is the single asset contract every game exports into `site/public/<slug>/`,
consumed by the shared viewer core in `site/src/shared/`. It supersedes format 1
(`FORMAT.md`): the 2-D tilemap body is format 1 unchanged, now carried inside a common
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
from format 1; see `FORMAT.md` for its `patrol`/`variants`/`dirStamps` sub-schema.

## Animation, collision, palette effects

Carried unchanged from format 1: `tileAnims`, `cellAnims`, `collision`
(`kind:"grid"` per-tile solidity, or `kind:"profiles"` height columns via a shared
`shapes.json`), and `paletteFx` (colour cycling + water raster split). See `FORMAT.md` for
their field-level definitions until this document absorbs them at Phase D.

## sprites/index.json

Unchanged from format 1: one entry per sprite key (`src`, `frames` rects, `anchor`,
`durations`, optional `steps` program, optional `path` per-frame offsets). Sonic scopes
keys by zone (`"0/08"`); other games use flat names.

## What each game emits

Every game emits `manifest.json`. Beyond that it fills only the asset kinds and level
sections its engine has — e.g. Sonic: `tilemap2d` + `blocks` + `paletteFx` + profile
collision; Ultima Underworld: `mesh3d` + `objectsFile`; Super Mario 64 DS: `models/` (GLB)
+ `mesh3d` levels + `objectsFile`. The per-section support matrix from `FORMAT.md` still
applies to the tilemap2d games.
