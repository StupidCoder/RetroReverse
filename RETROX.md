# Retro-X — an open exchange format for reverse-engineered retro game assets

**Retro-X** ("Retro-Exchange") is a JSON + standard-media file format for publishing the
assets of a reverse-engineered retro game — levels, objects and characters, music, sound
effects, pictures and videos — in a form any web viewer or tool can consume. The **X**
stands for *exchange*, and for the many asset types the format carries.

This document is the complete, self-contained specification of **Retro-X version 1**. A
reader should be able to write an exporter or a viewer from this document alone. It
supersedes the repository's earlier `FORMAT2.md`.

Design goals:

- **One format, zero per-game code.** A viewer implements this spec once; every game is
  pure data. If a game's assets cannot be represented, the format — not the viewer — must
  grow.
- **Small and general.** One general mechanism is preferred over several special-cased
  ones (e.g. skyboxes, environment swaps and collision overlays are all just *layers*).
- **Standard payloads.** Geometry is glTF-Binary (GLB), images are PNG, audio is MP3,
  video is H.264 MP4. The JSON documents only *bind* these payloads together.

## Contents

1. [Versioning and the document header](#1-versioning-and-the-document-header)
2. [Global conventions](#2-global-conventions)
3. [The game tree and `index.json`](#3-the-game-tree-and-indexjson)
4. [`manifest.json`](#4-manifestjson)
5. [Level documents](#5-level-documents)
6. [Object documents](#6-object-documents)
7. [Cutscene scripts](#7-cutscene-scripts)
8. [The camera block](#8-the-camera-block)
9. [Media assets](#9-media-assets)
10. [Validation rules](#10-validation-rules)

---

## 1. Versioning and the document header

Every Retro-X JSON document begins with the same two fields:

```jsonc
{ "format": "retro-x", "version": 1, ... }
```

- `format` is always the literal string `"retro-x"`.
- `version` is an integer, incremented **only for compatibility-breaking changes**.
  Adding a new *optional* field does not bump the version. A consumer MUST reject a
  document whose `version` is greater than the highest version it implements, and SHOULD
  ignore unknown fields within a version it does implement (that is what makes additive
  evolution safe).

This applies to `index.json`, `manifest.json`, level documents, object documents,
cutscene scripts and camera-track files. Media payloads (PNG/MP3/MP4/GLB) carry no
header; their integrity is their own container's business.

## 2. Global conventions

**Identifiers.** Every asset, animation, layer, variant, route, room, pool, shot and
placement has an `id` — a string (or integer for placements/rooms) unique within its
scope. All cross-references are **by id**; `name` fields are display-only and carry no
semantics. Asset ids appear in URLs, so they SHOULD be short lowercase slugs
(`[a-z0-9-]+`).

**File references** are POSIX-style relative paths, resolved **relative to the directory
of the referencing document**. They MUST NOT contain `..` or be absolute. The manifest
references documents; a document references its own payloads (which by convention live in
a directory named after the document, §3).

**2-D space.** Coordinates are world **pixels**, x right, y **down**, origin at the
level's top-left. Grids are row-major. Times are **engine frames** counted at the
manifest's `display.tickHz` (50 for PAL machines, 60 for NTSC/handhelds).

**3-D space.** Right-handed, **Y-up**, in the game's own world units — an exporter
normalizes whatever axis/scale conventions the game engine uses so that one game is
internally consistent. A viewer never needs to know the absolute scale because every 3-D
asset carries a complete `camera` block (§8) with explicit distances, near/far planes and
movement speed in those same units.

**Rotations** are `[rx, ry, rz]` in **radians**, applied as intrinsic Euler X-then-Y-then-Z
(the glTF/three.js `'XYZ'` order). A placement MAY instead carry a full `matrix` of **16
floats, column-major** (the glTF convention); when present it overrides `pos`/`rot`/`scale`
entirely. Exporters convert engine-native units (degrees, 4096-per-turn, etc.) at export
time; raw engine values never appear in typed fields (they may survive in the freeform
`props` bag for display).

**3-D animation time.** Clips inside a GLB are in seconds (glTF native). Every animation
metadata entry declares the authored `fps` so a viewer can step-quantize playback to the
original frame rate. Cutscene scripts declare one `fps` and address everything in integer
frames of that clock.

**Colours** are `"#rrggbb"` (or `"#rrggbbaa"` where alpha is meaningful).

**Optionality.** Fields marked ✱ are required. Every other field is optional and is
**omitted** (not null, not empty) when it has nothing to say. Booleans default to
`false`, numbers to `0`, unless a different default is stated.

## 3. The game tree and `index.json`

A Retro-X publication is a directory of games plus one root index:

```
<root>/
  index.json                 # { "format":"retro-x", "version":1, "games":["sonic-gg", ...] }
  <game-id>/
    manifest.json
    logo.png                 # optional, referenced from the manifest
    docs/*.html              # optional technical write-up pages
    levels/<id>.json         # one document per level asset
    levels/<id>/...          # that level's payloads: atlas.png, geo/*.glb, rooms/*.glb,
                             #   cameras/*.json, <script>.json, shapes.json, col/*.glb
    objects/<id>.json        # one document per object asset
    objects/<id>.png|.glb    # the object's payload, sibling with the same basename
    objects/tex/*.png        # inspectable texture sheets (atlasPicture)
    music/*.mp3   sfx/*.mp3   pictures/*.png   videos/*.mp4
```

`index.json` lists game ids in display order; each id names a subdirectory containing a
`manifest.json`. Everything a viewer shows is reachable from the index — there are no
side-channel files.

## 4. `manifest.json`

The per-game entry point: game metadata plus a flat list of assets.

```jsonc
{
  "format": "retro-x", "version": 1,                       // ✱
  "id": "sonic-gg",                                        // ✱ matches the directory name
  "title": "Sonic the Hedgehog",                           // ✱ display title
  "platform": "Game Gear",                                 // display string
  "year": 1991,                                            // original release year
  "description": "One-paragraph editorial description.",   // landing-page blurb
  "logo": "logo.png",                                      // small logo bitmap

  "display": {                                             // ✱
    "native": { "w": 160, "h": 144 },                      // ✱ the machine's visible resolution
    "tickHz": 60,                                          // ✱ engine frame clock for all 2-D times
    "filter": "gg"                                         // preferred screen-filter profile id
  },                                                       //   (viewer-defined set, e.g. crt|gb|gg);
                                                           //   omit = no filter offered

  "docs": [                                                // technical write-up tabs (static pages)
    { "id": "engine", "title": "Game Engine", "file": "docs/engine.html" }
  ],

  "assets": [ ... ]                                        // ✱ see below
}
```

`platform`, `year`, `description` and `logo` are presentation niceties for the game
list — assets work without them, so they are optional. `display.native` doubles as the
screen filter's pixel grid. An individual asset whose
internal render raster differs from the platform (e.g. a 200-line 3-D race view on a
256-line machine) declares a `pixelGrid` override in its own document (§5.1).

### Asset entries

One flat array; `category` discriminates. Common fields:

| field | | meaning |
|---|---|---|
| `id` | ✱ | unique within the game; used in URLs |
| `category` | ✱ | `level` \| `object` \| `music` \| `sfx` \| `picture` \| `video` |
| `name` | ✱ | display name |
| `group` | | grouping label for menus ("Green Hill", "Cars") |
| `description` | | editorial text shown in the asset's info dialog |
| `related` | | array of asset ids to cross-link (e.g. a course's alternate representation) |

Category-specific fields:

```jsonc
// level / object — the entry points to an external document (§5, §6):
{ "id":"ghz1", "category":"level",  "name":"Green Hill Act 1", "group":"Green Hill",
  "file":"levels/ghz1.json", "description":"…", "related":["ghz1-map"] },
{ "id":"crab", "category":"object", "name":"Crabmeat", "file":"objects/crab.json" },

// media — leaf entries, no sub-document:
{ "id":"greenhills", "category":"music", "name":"Green Hills",
  "file":"music/greenhills.mp3", "loop":true, "duration":93.2 },
{ "id":"ring",  "category":"sfx",     "name":"Ring",        "file":"sfx/ring.mp3" },
{ "id":"title", "category":"picture", "name":"Title screen","file":"pictures/title.png",
  "w":160, "h":144 },
{ "id":"intro", "category":"video",   "name":"Intro",       "file":"videos/intro.mp4",
  "w":320, "h":240, "fps":15, "duration":42.0 }
```

`music.loop` marks a track that loops in-game (a player may loop it seamlessly).
`duration` is seconds. `w`/`h` are source pixel dimensions (used for letterboxing and
info display).

## 5. Level documents

A level is either a 2-D tilemap or a 3-D scene, plus everything instantiated inside it.

### 5.1 Envelope

```jsonc
{
  "format": "retro-x", "version": 1,       // ✱
  "type": "tilemap" | "scene3d",           // ✱ selects which body is present
  "music": "greenhills",                   // music asset id; auto-plays while viewing
  "pixelGrid": { "lines": 200 },           // override of display.native for this asset's raster
  "camera": { ... },                       // §8 — ✱ for scene3d; optional for tilemap
  "variants": [                            // alternate object sets over the same geometry
    { "id": "star1", "name": "Star 1 — First mission", "default": true }
  ],
  "tilemap": { ... } | "scene": { ... },   // ✱ exactly one body (§5.2 / §5.3)
  "collision": { ... },                    // 2-D collision (§5.4); 3-D collision is a layer
  "placements": [ ... ],                   // §5.5
  "pools": [ ... ],                        // §5.6
  "routes": [                              // shared polylines for route-following & drive cams
    { "id": "lap", "loop": true, "points": [[x,y,z], ...] }
  ],
  "scripts": [                             // cutscenes owned by this level (§7)
    { "id": "intro", "name": "Opening cutscene", "file": "ghz1/intro.json",
      "description": "Plays on boot; Luigi walks through the forest." }
  ]
}
```

**Variants** are the single mechanism for "same level, different object sets": per-star
object layers, per-mission masks, day/night sets, multiplayer trims. When `variants` is
present the viewer shows a picker; each placement/pool lists the variant ids it belongs to
(absent = present in all variants). Exactly one variant may be `default: true` (otherwise
the first is the default).

### 5.2 `tilemap` body (2-D)

```jsonc
"tilemap": {
  "tileSize": 8,                            // ✱ pixels per tile
  "width": 203, "height": 16,               // ✱ map size in CELLS, row-major
  "atlas": { "file": "ghz1/atlas.png",      // ✱ tile sheet; tiles packed left-to-right,
             "cols": 16,                    //   top-to-bottom, `cols` tiles per sheet row
             "gutter": 0 },                 //   1 = each tile extruded 1px (bleed guard)
  "cells": [ ... ],                         // ✱ width*height tile (or block) ids
  "hflipMask": 32768,                       // cell & mask ⇒ draw h-flipped; id = cell & ~mask
  "blocks": {                               // optional block indirection:
    "size": 4,                              //   cells index blocks; each block is size×size tiles
    "tiles": [ [16 tile ids], ... ],        //   (cell pixel size becomes tileSize*size)
    "shapes": [ shapeId, ... ]              //   per-block collision-profile id (§5.4 "profiles")
  },
  "wrap": "none" | "x",                     // "x": horizontal cylinder — the viewer tiles the
                                            //   map seamlessly and clamps zoom-out to one period
  "view":  { "x": 0, "y": 256, "w": 160, "h": 144 },   // initial framing, world px
  "spawn": { "x": 160, "y": 368, "object": "sonic", "anim": "stand", "tint": "#b8c76f" },
  "tileAnims": [                            // global tile animation (tile replacement):
    { "tiles": [252,253,254,255],           //   any cell showing tiles[i] cycles through
      "frames": [[...4 ids], ...],          //   frames[step][i]
      "periodFrames": 10 }
  ],
  "cellAnims": [                            // anchored strip animators with per-phase holds:
    { "tx": 36, "ty": 44, "tw": 2, "th": 4, //   a tw×th tile rectangle at cell (tx,ty)
      "phases": [ { "tiles": [tw*th ids], "frames": 240 }, ... ] }
  ],
  "paletteFx": {                            // palette effects:
    "palette": ["#...", ...],               //   the level's base palette
    "cycle":   { "slots": [10,11,12],                  // palette indices that cycle
                 "steps": [["#..","#..","#.."], ...],  // per step: one colour per slot
                 "periodFrames": 32, "tiles": [12,13] },  // tiles that use a cycling slot
    "regions": [                            // areas drawn with an alternate palette
      { "name": "Underwater",               //   (raster splits on the original hardware)
        "rect": { "x": 0, "y": 416, "w": 1952, "h": 1632 },   // world px
        "palette": ["#...", ...] }
    ]
  }
}
```

### 5.3 `scene` body (3-D)

```jsonc
"scene": {
  "background": "#000010",                  // clear colour
  "fog": { "color": "#334455", "near": 500, "far": 9000 },
  "layers": [                               // ✱ ≥1 unless `rooms` is present —
                                            //   THE general composition mechanism
    { "id": "terrain",                      // ✱
      "name": "Terrain",                    //   label for the layer-toggle UI
      "file": "ghz1/geo/terrain.glb",       // ✱ GLB payload
      "mode": "base",                       //   "base" (always on, no UI) |
                                            //   "toggle" (checkbox) |
                                            //   "exclusive:<group>" (radio among same group)
      "visible": true,                      //   initial state for toggle/exclusive (default true)
      "attach": "world",                    //   "world" | "camera" (follows the eye — skyboxes) |
                                            //   "cameraYaw" (follows position, not pitch/roll)
      "renderOrder": 0,                     //   explicit paint order (skyboxes negative,
                                            //   overlays positive)
      "transparent": false,                 //   force transparent materials
      "depthTest": true,                    //   set false + renderOrder for painter's-algorithm
                                            //   2-D-from-3-D scenes
      "polygonOffset": 0,                   //   push back to resolve coplanar decal layers
      "role": "collision"                   //   semantic hint; a viewer may style/label it
    }                                       //   (recognised roles: "collision", "sky", "water")
  ],
  "rooms": {                                // optional room-graph assembly:
    "areas": [                              // vertical/spatial areas — drive "peel"
      { "id": "basement", "name": "Basement" },        //   visibility toggles
      { "id": "ground", "name": "Ground floor" }
    ],
    "stream": true,                         // load rooms progressively, nearest first
    "list": [
      { "id": 5, "name": "Foyer", "file": "ghz1/rooms/r05.glb",
        "area": "ground",                   //   area id (ids everywhere, names display-only)
        "aabb": { "min": [x,y,z], "max": [x,y,z] } }
    ]
  }
}
```

Everything that used to be a special case is a layer: a camera-locked skybox is
`attach:"camera", renderOrder:-1000`; a streamed-vs-static environment swap is two layers
in one `exclusive:` group; a collision overlay is a hidden `toggle` layer with
`role:"collision"`; a painter's-ordered 2.5-D scene sets `transparent + depthTest:false` and
per-layer `renderOrder`. Material properties (single/double-sided, alpha mode, vertex
colours, unlit) live inside the GLB itself, as does skinning, morph targets and clip data.

Placements may carry a `room` id; a viewer streams a room's shell together with its
placements, and the `areas` list drives dollhouse-style peel toggles.

### 5.4 2-D collision

```jsonc
// per-tile solidity grid:
"collision": { "kind": "grid",
  "sub": 4,                                 // each tile splits into sub×sub cells (1 = whole tile)
  "solid": [ ... ],                         // tileId*sub*sub + r*sub + c → 0 (empty) or class byte
  "legend": { "1": "#ff3030", "127": "#33ddff" } }   // class byte → overlay colour

// per-column height profiles (slope engines):
"collision": { "kind": "profiles", "file": "ghz1/shapes.json" }
// shapes.json: { "format":"retro-x", "version":1, "count": N,
//   "profiles": [ [32 signed column heights; -128 = none], ... ],  // one per shape id
//   "angles":   [ one slope angle per shape ] }
// the profile for a cell comes from blocks.shapes[cellBlockId]
```

3-D collision is not a separate mechanism: level collision is a `role:"collision"` layer;
per-object collision is a placement field (§5.5).

### 5.5 Placements

One shape for 2-D and 3-D. A placement instantiates an **object asset** (§6) — always by
reference, never inline.

```jsonc
{ "id": 7,                                  // ✱ unique within the level
  "object": "crab",                         // ✱ object asset id
  "pos": [x,y] | [x,y,z],                   // ✱ (unless matrix is given)
  "rot": [rx,ry,rz],                        // radians, XYZ (3-D only)
  "scale": 1.0 | [sx,sy,sz],                // default 1
  "matrix": [ ...16 floats, column-major ], // full transform; overrides pos/rot/scale
  "anim": "walk",                           // initial animation id (default: the object's first)
  "hflip": true,                            // 2-D horizontal flip
  "tint": "#b8c76f",                        // 2-D multiply tint
  "hard": true,                             // 2-D "solid to the player" flag (info display)
  "layer": "objects",                       // scene layer this placement toggles with
  "room": 5,                                // room id (scene3d rooms)
  "variants": ["star1","star3"],            // level-variant membership; absent = all
  "collision": { "file": "ghz1/col/obj12.glb",   // per-object collision mesh +
                 "matrix": [ ...12 floats ] },   //   3×4 local→world (row-major rows of a 4×4)
  "route": { "id": "lap", "speed": 220,     // follow a level route (constant world speed;
             "mode": "loop" | "pingpong",   //   wrap or out-and-back)
             "face": true },                //   face travel direction
  "behavior": { ... },                      // parametric motion, below
  "onClick": { ... },                       // interaction, below
  "name": "Crabmeat #7",                    // instance display name
  "info": { "title": "…", "body": "…", "quote": "…" },  // extracted/editorial info-card text
  "props": { ... }                          // freeform provenance (raw engine ids, addresses,
}                                           //   flags) — displayed verbatim in the info card
```

**Behaviours** are a small closed vocabulary of *data-driven* motion (decoded from the
game, not reimplemented AI). Version 1 defines:

```jsonc
{ "kind": "spin",  "axis": [0,1,0], "rate": 1.2 }        // rad/s about a local axis
{ "kind": "flyer",                                        // scripted waypoint flight:
  "keys": [ { "pos": [x,y,z], "dur": 120, "hold": 0, "yaw": 1.57 }, ... ],
  "loop": true,                                           //   dur/hold in engine frames
  "spinPart": { "node": "rotor", "axis": [0,1,0], "rate": 20 } }  // named sub-node spins
```

A viewer MUST ignore placements' unknown `behavior.kind` values (render the object
static). Route-following is the `route` field, not a behaviour.

**onClick** defines what a click (or tap) on the object does. Version 1 defines two
actions; unknown `action` values are ignored — that is the extension point for future
actions (switch level, hide object, ...).

```jsonc
// play an animation:
{ "action": "animate",
  "target": "self",                          // or another placement id
  "clip": "open",                            // animation id on the target's object
  "holdAt": 0.5,                             // freeze at this normalized clip time
                                             //   (a door's half-open apex); absent = play out
  "toggle": true,                            // second click plays back from the held position
  "sfx": [ { "id": "door-creak", "at": 0.0 } ] }   // sfx asset ids, offset seconds into the clip

// show a text popup:
{ "action": "text", "title": "Sign", "body": "WELCOME TO THE FIRST LEVEL..." }
```

Double-click/long-press is reserved by the viewer for the object info popup (name, stats,
`info`, `props`, description, link to the object asset) and is not represented in data.

### 5.6 Pools (randomized placements)

A pool places `count` instances at randomly chosen candidate positions each time the
object layer is enabled. With `seedable`, a viewer seeds the RNG from a URL parameter so a
particular roll is shareable/reproducible.

```jsonc
"pools": [
  { "id": "prisoners",
    "count": 8,                              // ✱ how many to place
    "object": "prisoner",                    // ✱ object asset id
    "candidates": [ [x,y], ... ],            // ✱ possible positions (pick count of them)
    "seedable": true,                        // honour a user-supplied RNG seed
    "anim": "stand", "tint": "#352879",      // as on a placement (applied to every instance)
    "name": "Prisoner",                      // instance display name for info cards
    "info": { "title": "…", "body": "…" },   // instance info-card text
    "variants": ["day"]                      // level-variant membership (as placements)
  }
]
```

Pools are static in version 1: instances do not move, and every instance of a pool looks
the same (per-instance random variation would be an additive future field — no exported
game needs it).

## 6. Object documents

An object (or character) is a reusable, placeable thing with named animations. Four types.
Every animation entry carries `id` ✱, optional `name`/`description`, and
`loop`: `"once"` (play and hold first frame), `"loop"`, `"pingpong"`, or `"hold"` (play
and hold last frame).

### 6.1 `sprite2d`

The payload is **one PNG atlas per object**: each **row is one animation, each column one
frame**, on a uniform cell grid (frames padded/centred into the cell by the exporter).
This makes the sheet directly usable in other projects and legible to a human.

```jsonc
{ "format": "retro-x", "version": 1, "type": "sprite2d",
  "name": "Crabmeat",
  "atlas": {
    "file": "crab.png",                      // ✱ sibling atlas
    "cellW": 48, "cellH": 48,                // ✱ uniform cell size
    "anchor": [24, 48]                       // default draw origin within a cell, px
  },
  "animations": [                            // ✱ ≥1
    { "id": "walk", "name": "Walk",
      "row": 0,                              // ✱ atlas row
      "frames": 4,                           // ✱ frame count (columns used)
      "loop": "loop",                        // ✱
      "durations": [13,13,13,13],            // per-frame hold, engine frames …
      "steps": [[0,300],[1,16],[0,16]],      // …OR an explicit [frameIndex, hold] program
                                             //   (mutually exclusive with durations;
                                             //    default: every frame held 1)
      "path": [ [dx,dy], ... ],              // per-frame world-position offsets (moving
                                             //   platforms); loops with the animation
      "anchor": [24, 48],                    // per-animation anchor override
      "mirror": "walk-left",                 // id of the left-facing art; absent ⇒ the left
                                             //   facing is a horizontal flip of this row
      "events": [ { "frame": 2, "sfx": "step" } ],   // per-frame sound-cue hooks
      "description": "Plays while pacing its ledge." }
  ],
  "stats": { "HP": 1, "Score": 100 },        // freeform display table (info panel)
  "props": { ... }                           // freeform provenance
}
```

### 6.2 `model3d`

```jsonc
{ "format": "retro-x", "version": 1, "type": "model3d",
  "name": "Standard Kart",
  "model": "kart.glb",                       // ✱ sibling GLB (geometry, materials, skins,
                                             //   morph targets and clips all live inside)
  "variants": [                              // independent alternates of the model: one glTF
    { "id": "car",  "name": "Car",           //   SCENE each in the same GLB. Scene 0 is the
      "scene": "car" },                      //   default, so a viewer that knows nothing of
    { "id": "lod1", "name": "LOD 1",         //   variants shows only that one. The Studio
      "scene": "lod1",                       //   offers a picker and swaps scenes in place;
      "description": "mid-distance model" }  //   buffers and textures are shared in the file.
  ],                                         //   First entry must name scene 0. Use for LOD
                                             //   chains, shadow-caster proxies, liveries.
  "instanced": true,                         // geometry safe to share across placements
  "skinnedClone": false,                     // placements must deep-clone (skinned meshes)
  "billboard": "yaw",                        // rotate the whole model to face the camera
                                             //   about world-up (flat "tree quad" models);
                                             //   per-BONE billboarding is a GLB node extra
                                             //   ({"billboard": true}), not a document field
  "animations": [                            // metadata over the GLB's named clips
    { "id": "spin", "clip": "spin",          // ✱ clip = the GLB animation name
      "name": "Victory spin", "fps": 30,     // authored frame rate (step-quantized playback)
      "loop": "loop", "description": "…" }
  ],
  "uvAnims": [                               // material UV (texture-matrix) animation tracks:
    { "material": "water",                   //   bound by GLB material name
      "frames": 128,                         //   loop length, engine frames
      "scaleS": { "const": 1.0 },            //   each channel: {const: v} or
      "scaleT": { "const": 1.0 },            //   {samples: [...], step: N} — one sample every
      "rot":    { "const": 0.0 },            //   N frames, linearly interpolated
      "transS": { "samples": [...], "step": 2 },
      "transT": { "const": 0.0 } }
  ],
  "flipbooks": [                             // texture-pattern (flipbook) animation:
    { "material": "screen",
      "textures": ["tex/tv0.png","tex/tv1.png"],
      "sequence": [0,1,1,0],                 //   indices into textures
      "step": 4 }                            //   engine frames per sequence entry
  ],
  "atlasPicture": "tex/kart-atlas.png",      // the texture sheet, openable as a picture
  "stats": { ... }, "props": { ... }
}
```

### 6.3 `billboard3d`

A **view-angle-dependent sprite** in a 3-D scene: the art shown depends on the camera's
bearing to the object (8-view creatures and props in engines that draw flats). This is
the one behaviour glTF cannot express — there is no way to encode "select an atlas
region from the viewing angle" inside a GLB — and it keeps such art as an inspectable
sprite sheet. A flat model that merely turns to face the camera is NOT this type; that is
a `model3d` with `billboard:"yaw"`. First-class object — 100 billboard ghosts are one
object and 100 placements. The atlas convention matches `sprite2d` with one twist:
**rows are view directions**, and an animation is a run of columns.

```jsonc
{ "format": "retro-x", "version": 1, "type": "billboard3d",
  "name": "Ghost",
  "atlas": { "file": "ghost.png", "cellW": 32, "cellH": 48 },   // ✱ row = view direction
  "views": 8,                                // ✱ direction buckets (rows); 1 = always same art
  "heading": 0.0,                            // world yaw of view row 0, radians
  "mode": "camera" | "yaw",                  // full camera-facing vs rotate about world-up only
  "size": [1.0, 1.5],                        // ✱ world-unit quad [w,h]
  "anchorMode": "bottom" | "center",         // what a placement's pos means (default center)
  "blend": "opaque" | "alpha" | "additive",  // default opaque (alpha-tested cut-out)
  "animations": [                            // ✱ ≥1; columns [col, col+framesPerView)
    { "id": "drift", "col": 0, "framesPerView": 2, "fps": 8, "loop": "loop" }
  ],
  "stats": { ... }, "props": { ... }
}
```

The rendered view row is `quantize(angleFromObjectToCamera - heading, views)`; the frame is
`floor(t * fps) mod framesPerView` within the animation's column run.

### 6.4 `wireframe3d`

Vector-display objects (filled-vector and wireframe engines). Carries the model's own
hidden-surface data so a viewer can reproduce the original renderer: an edge is drawn when
either adjacent face points toward the eye.

```jsonc
{ "format": "retro-x", "version": 1, "type": "wireframe3d",
  "name": "Cobra Mk III",
  "wireframe": {
    "positions": [ x,y,z, ... ],             // ✱ flat vertex array
    "edges": [ [v0, v1, faceA, faceB], ... ],// ✱ vertex pair + the two adjacent face ids
    "faces": [ [nx,ny,nz], ... ],            // ✱ outward face normals
    "faceCenters": [ [x,y,z], ... ]          // ✱ a point on each face (for the visibility dot)
  },
  "stats": { ... }, "props": { ... }
}
```

Face `i` is visible when `dot(faces[i], eye - faceCenters[i]) > 0`. A face id of `-1` on an
edge means "no face on this side" — an open edge, drawn unconditionally (Elite's blueprint
sentinel `$F` maps here). Presentation effects (authentic CRT edge flicker, glow) are viewer
options, not data.

## 7. Cutscene scripts

A cutscene is an **animation script owned by a level** (`scripts[]`, §5.1): a sequence of
*shots* binding the level's layers and placements to a camera track, lights and sound cues
on one frame clock. It is data-only replay — no game logic.

```jsonc
{ "format": "retro-x", "version": 1,
  "name": "Opening cutscene",
  "fps": 30,                                 // ✱ the script clock; all frames below use it
  "shots": [                                 // ✱ played in order; a player shows chapters
    { "id": "opwf", "name": "The forest walk",
      "frames": 234,                         // ✱ shot length
      "layers": ["forest"],                  // scene layers visible during this shot
                                             //   (omit when the shot's sets are actors —
                                             //    an animated set-piece is a placement)
      "actors": [                            // placements animated during the shot:
        { "placement": 12,                   // ✱ placement id in the owning level
          "clip": "walk",                    //   animation id; absent = posed static
          "start": 0,                        //   frame at which the clip starts
          "matrix": [ ...16 floats ],        //   per-shot transform override
          "mirror": false }                  //   mirror the actor (X-flip)
      ],
      "camera": {                            // ✱ scripted camera:
        "near": 1, "far": 327680,
        "trackFile": "cameras/opwf.json"     //   baked per-frame track (below), or inline
        // "track": [ {…}, … ]               //   with the same element shape
      },
      "lights": [                            // optional keyed lights:
        { "id": "key", "type": "ambient" | "directional" | "point",
          "keys": [ { "frame": 0, "color": "#404060", "pos": [x,y,z], "dir": [x,y,z] } ] }
      ],
      "sounds": [                            // optional cue track:
        { "sfx": "thunder",                  //   sfx asset id
          "start": 30, "end": 90,            //   frames; end 0/absent = one-shot
          "volume": 1.0, "pan": 0.0,         //   pan -1..1
          "reverse": false }                 //   cue authored for reverse playback
      ] }
  ]
}
```

Placements referenced as actors are hidden outside the script and shown/posed by the
player during it. Sounds are optional by design — a game may ship a silent-capable
cutscene first and gain cues later without a format change.

**Camera-track file** (also usable standalone for fly-through tours):

```jsonc
{ "format": "retro-x", "version": 1,
  "frames": 234, "fps": 30,                  // ✱
  "near": 1, "far": 327680,                  // defaults for the consumer
  "track": [                                 // ✱ exactly one sample per frame:
    { "pos": [x,y,z], "target": [x,y,z], "roll": 0.0, "fov": 45.0 }, ...
  ]
}
```

## 8. The camera block

Standard interactive-camera description. Required for every 3-D asset (this is what makes
viewers scale-agnostic); optional for 2-D levels (default: fit the level).

```jsonc
"camera": {
  "mode": "map2d" | "orbit" | "fly" | "ortho",   // ✱
  "pos": [x,y,z], "target": [x,y,z],             // ✱ for 3-D modes: the opening shot
  "fov": 45, "near": 0.1, "far": 50000,          // omit for sensible defaults derived from
                                                 //   the pos↔target distance
  "map2d": { "minFitFactor": 1.0,                // never zoom out past fitting × this
             "maxNativeFactor": 4 },             // never zoom in past native pixel × this
  "orbit": { "minDist": 2, "maxDist": 400,
             "autoRotate": false, "autoRotateSpeed": 0.3 },
  "fly":   { "speed": 600 },                     // ✱ for fly: world units/second
  "ortho": { "dir": [1,-1,1],                    // fixed dimetric/isometric view direction
             "zoomMin": 0.5, "zoomMax": 8 },
  "drive": { "route": "lap", "eyeHeight": 120,   // optional auto-drive along a level route
             "speed": 900, "mode": "loop" | "pingpong" }
}
```

`orbit` is the default presentation for object assets (a viewer supplies its own defaults
there; object documents do not carry camera blocks).

## 9. Media assets

- **Music** — MP3, one file per track. `loop` and `duration` in the manifest entry.
- **Sound effects** — MP3, one file per effect. Referenced by asset id from animation
  `events`, `onClick.sfx` and cutscene `sounds`.
- **Pictures** — PNG. Title/loading screens, maps, montages — and object texture sheets
  (`atlasPicture`) open through the same picture presentation.
- **Videos** — MP4 (H.264 + AAC), progressive-download friendly (moov atom up front).
  `w`/`h`/`fps`/`duration` in the manifest entry.

## 10. Validation rules

A conforming tree satisfies all of the following (the reference validator enforces them):

1. Every JSON document parses, carries `format:"retro-x"`, and its `version` ≤ the
   validator's version.
2. `index.json` game ids ↔ directories with a `manifest.json`; manifest `id` matches its
   directory name.
3. Asset ids are unique per game; every `file` reference resolves to an existing file
   within the game directory (no `..`, no absolute paths).
4. Every id cross-reference resolves: `related`, level `music`, placement `object` /
   `anim` / `route` / `layer` / `room` / `variants` / onClick `clip` / onClick + events +
   cutscene `sfx`, pool `object`, script actor `placement`, drive `route`, room `area`.
5. `tilemap`: `cells.length == width*height`; every cell id (after masking `hflipMask`)
   indexes the atlas (or `blocks`); `blocks.tiles[i].length == size*size`.
6. `sprite2d`/`billboard3d`: the atlas PNG's dimensions are exact multiples of
   `cellW`/`cellH`; every animation's `row` (and column run) fits the sheet;
   `durations.length == frames`; `durations` and `steps` are not both present.
7. `model3d`: every `animations[].clip` names a clip present in the GLB; every
   `uvAnims[].material` / `flipbooks[].material` names a GLB material; every
   `variants[].scene` names a GLB scene, variant ids are unique, the first variant is the
   GLB's default scene (scene 0), and a one-entry `variants` list is rejected.
8. `wireframe3d`: edge vertex/face indices in range; `faces.length == faceCenters.length`.
9. Camera tracks: `track.length == frames`.
10. A `matrix` has 16 numbers; a placement `collision.matrix` has 12.
11. Every scene layer id referenced by placements/scripts exists; `exclusive:` groups have
    exactly one initially visible member; scripts' `layers` reference existing layer ids.
12. Media entries' `w`/`h` match the actual PNG dimensions (pictures).

---

*Retro-X version 1 — 2026. This specification is maintained in the RetroReverse
repository (`RETROX.md`); the reference implementation is the `tools/lib/retrox` Go
library, the `retroxlint` validator, and the Studio web viewer.*
