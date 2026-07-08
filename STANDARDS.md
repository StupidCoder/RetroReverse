# RetroReverse conventions

This document defines the coherent structure every game and tool in the repository
follows. It is the authoritative reference for layout, command-line interfaces, extracted
asset formats, and writeup style. New games and tools conform to it; existing ones are
migrated to it (see the restructuring rollout).

Nothing here overrides the two standing rules: derive everything from the game image via
our own tools (never external game-specific sources), and reimplement decode algorithms in
Go rather than scraping oracle output.

---

## 1. Directory hierarchy

```
games/<slug>/                    # one directory per game
  <slug>.md                      # the writeup; pins the image MD5 near the top
  image/                         # ROM/disk image(s) — .gitignored (copyright)
  extract/                       # Go module: retroreverse.com/games/<slug>/extract
    go.mod
    cmd/<tool>/main.go           # one command per tool
    <packages>
  disasm/                        # optional: text disassembly dumps
  figures/                       # images embedded in <slug>.md — tracked
  work/                          # regenerable dev/debug scratch — .gitignored
tools/
  cpu/       mos6502 m68k z80 sm83 arm arm60 mips x86
  platform/  psx nds dos gameboy gamegear threedo amiga c64
             (each platform owns its format/codec libs, e.g. amiga/iff, nds/nitro, c64/sid)
  lib/       reserved for genuinely cross-platform helpers (none yet)
  cmd/       dis<cpu> / codetrace<cpu> command binaries
site/
  public/<slug>/                 # the canonical extracted-asset tree (Standard 4)
```

**Slugs** are `lowercase-with-hyphens`, ending in the platform tag:
`elite-c64`, `fort-apocalypse-c64`, `marble-madness-amiga`, `mario-kart-ds`,
`need-for-speed-3do`, `ridge-racer-psx`, `sonic-gg`, `stunt-car-racer-amiga`,
`super-mario-64-ds`, `super-mario-land-gb`, `turrican-amiga`, `ultima-underworld-pc`.

The `site/public/<slug>/` asset key matches the game slug exactly.

The two ARM cores are distinct: `tools/cpu/arm` is little-endian (Nintendo DS), `tools/cpu/arm60`
is big-endian ARMv3 (3DO).

**Modules.** `tools` is one module (`retroreverse.com/tools`); each game's `extract/` is
its own module (`retroreverse.com/games/<slug>/extract`). `go.work` lists `./tools` and one
`./games/<slug>/extract` per game.

**Copyright.** Game ROM/disk images are never committed — `games/*/image/` and the image
extensions (`*.tap *.adf *.nds *.gb *.gg *.bin *.cue *.prg`) are `.gitignored`. Each image
is identified by an MD5 pinned in that game's writeup, so a reader can supply the exact
image themselves. Extracted assets under `site/public/<slug>/` are derived data (our own
output) and stay tracked.

**Per-game output dirs.** Three distinct roles, kept separate:

- `site/public/<slug>/` — curated web deliverables (the `webexport` output). Tracked.
- `games/<slug>/figures/` — **only** the images actually embedded in `<slug>.md`
  (`![...](figures/...)`). Tracked, so the writeup renders on GitHub. Nothing else belongs
  here — an image that the writeup does not embed is not a figure.
- `games/<slug>/work/` — regenerable dev/debug scratch (inspection renders, dumps). All of
  it is `.gitignored`. Anything a tool can re-emit lives here, never in `figures/`.

There is no `rendered/` or `extracted/` dir; regenerable output that no writeup embeds is
either written to `work/` or not kept at all.

---

## 2. CPU toolchain flags (`tools/cmd/dis<cpu>`, `tools/cmd/codetrace<cpu>`)

Every CPU has three components at parity: a disassembler command, a code-tracer command,
and an execution core with a validating test. One flag vocabulary across all of them.

**Disassemblers** (`dis6502`, `dis68k`, `disz80`, `dissm83`, `disarm`, `disarm60`, `dismips`, `disx86`):

| flag | meaning |
|---|---|
| `-base HEX`   | load address of the region (may auto-derive, e.g. from a `.prg` header) |
| `-start HEX`  | first address to disassemble (absolute) |
| `-end HEX`    | last address to disassemble (absolute) |
| `-skip INT`   | leading bytes of the file to drop before `-base` |

**Tracers** (`codetrace<cpu>`): the same `-base`/`-skip`, plus the shared set already in
use — `-entry HEX[,HEX...]`, `-table ADDR:N` (repeatable jump-table seed), `-annotate FILE`,
`-o FILE` (default stdout).

**Architecture extras** (kept, documented per tool): `-thumb` (arm), `-bank` (sm83 banked
select), `-slots` (z80 bank layout). Endianness is a property of the core, not a flag
(`arm` LE vs `arm60` BE).

**Execution-core validation.** Prefer an external per-instruction conformance suite,
env-gated so it is skipped when the data is absent: mips (`PSX_SST_DIR`, SingleStepTests/psx),
x86 (`HARTE_DIR`, SingleStepTests/8088). mos6502 gets SingleStepTests/6502. Cores without a
published suite (m68k, z80, sm83, arm, arm60) keep hand-written assembled-loop tests.

---

## 3. Boot-oracle contract (`games/<slug>/extract/cmd/bootoracle`)

Each game exposes one `bootoracle` that boots the real image on the shared platform machine
+ CPU core. Oracles use Go's `flag` package (never raw `os.Args`) and this vocabulary:

| flag | meaning |
|---|---|
| `-image PATH`        | ROM / disc / game-dir input |
| `-steps N`           | run budget (accepts hex or decimal) |
| `-trace`             | trace execution |
| `-tracen N`          | limit traced instructions |
| `-bp ADDR`           | breakpoint (repeatable) |
| `-watch ADDR[:LEN]`  | memory watch (repeatable) |
| `-keys FILE`         | input script (keyboard/mouse/pad) |
| `-shot DIR`          | write framebuffer screenshot(s) |
| `-o DIR`             | output directory |
| `-savestate FILE` / `-loadstate FILE` | state snapshot |

Address arguments are the platform's natural form (flat hex, or `SEG:OFF` for x86). A
shared helper (`tools/platform/oracleflags`) registers the common set so games do not
re-declare it. Platforms with several single-purpose probes (Game Boy, Game Gear) fold them
under one `bootoracle` (thin wrappers are fine). NDS oracles share the dual-CPU
`tools/platform/nds/dsmachine`.

---

## 4. Extracted assets

### 4.1 Extract CLI (`games/<slug>/extract/cmd/webexport`)

Each game exposes one primary export command named `webexport` that runs the whole
extraction internally (from the raw image) and writes the full asset tree below plus
`manifest.json`:

- `-in PATH` — input image / rom / game-dir. Games that stage a pre-extracted tree may add
  `-extracted DIR` as a secondary input.
- `-o DIR` — output root, default `../../site/public/<slug>` for every game.
- `-only bitmaps|music|sfx|models|levels|sprites|all` — selective export (default `all`),
  comma-separated. **`-only` gates which stages run**, so an expensive stage is skipped when
  its assets are not requested (e.g. `-only music` must not boot a model/actor oracle).

**Progress output (required).** Every `webexport` reports what it is doing on stderr as it
runs: a line per stage announcing the stage, and progress within long stages (e.g.
`[models]  142/463  castle_1f.glb`). The output is human-readable and ordered; it names the
stage, the item, and a running count so a slow run (oracle, GLB export) is legible. Keep it
concise — one line per item or a periodic counter, not a flood.

Inspection/dev tools use `-in` + `-o` and write to `games/<slug>/work/`.

### 4.2 Asset tree (`site/public/<slug>/`)

```
site/public/<slug>/
  manifest.json     # top-level index (subsumes the old meta.json + models.json)
  bitmaps/          # standalone images — title/loading screens (PNG)
  music/            # tracks (MP3, uniform encode) + index
  sfx/              # sound effects (WAV/MP3)
  models/           # animated 3D models (GLB) + textures
  sprites/          # sprite atlases + index.json
  levels/           # <level>.json (format 2) + atlas PNG + <level>.objects.json
```

`manifest.json`:

```jsonc
{ "format": 2, "game": "<slug>", "platform": "Game Gear",
  "native": { "w": 160, "h": 144 }, "tickHz": 60, "wrap": "x"?,
  "bitmaps": [ { "name": "...", "file": "bitmaps/title.png" } ],
  "music":   [ { "name": "...", "file": "music/greenhills.mp3" } ],
  "sfx":     [ ... ],
  "models":  [ { "name": "...", "file": "models/kart.glb" } ],
  "levels":  [ { "name": "Green Hills Act 1", "section": "Green Hills",
                 "file": "levels/act01.json", "kind": "tilemap2d",
                 "objects": "levels/act01.objects.json" } ],
  "views":   [ ... ]  // escape hatch for bespoke per-game three.js views (Elite/Stunt/Marble slopes)
}
```

### 4.3 Level format 2

A single superset envelope; `kind` selects the body. See `FORMAT2.md` for the full
field-by-field spec (supersedes the earlier format-1 spec, now folded into `FORMAT2.md`).

```jsonc
{ "format": 2, "name": "...", "kind": "tilemap2d" | "mesh3d",
  "extents": { ... },   // tilemap2d: { tileSize, width, height } in cells
                        // mesh3d:    AABB { min:[x,y,z], max:[x,y,z] }
  "wrap": "none" | "x" | "xy",
  "spawn": { ... }, "view": { ... },
  "grid": { ... },      // present when kind == tilemap2d (the former format-1 body)
  "mesh": { ... },      // present when kind == mesh3d  (positions/uvs/groups/textures, or a GLB ref)
  "objects": [ ... ],           // lightweight inline placements
  "objectsFile": "<level>.objects.json",   // full machine-readable object DB
  "tileAnims": [], "cellAnims": [], "collision": {}, "paletteFx": {}, "music": "" }
```

### 4.4 Machine-readable object DB (`<level>.objects.json`)

Unifies the three prior schemas (format-1 `objects`, Ultima Underworld `picks`/`creatures`,
SM64DS actor placement):

```jsonc
{ "format": 2, "level": "...",
  "objects": [
    { "id": 0, "type": 8, "name": "crab",
      "pos": [x, y, z], "size": [w, h, d]?, "rot": [rx, ry, rz]?,
      "model": "models/crab.glb"?, "actor": 0x21?, "hard": true?,
      "props": { ... } } ] }
```

`z`/`size`/`rot`/`model`/`actor` are omitted where a game's engine has no such concept
(a 2-D tilemap object is `pos: [x, y]`, `hard`, and a sprite key in `props`).

A 3-D object is rendered by the shared object layer (`site/src/shared/renderers.js` →
`placeObjects`, used by the builtin `mesh3d` renderer and game plugins) two ways:

- `model` — a `"models/x.glb"` path; loaded and placed at `pos`/`rot` (`rot` in **radians**).
- `sprite` — a first-class **directional-billboard sprite spec** (`site/src/shared/sprites3d.js`):
  `{ billboard, views, heading (radians), size:[w,h], anchor, sheet, frames:[[x,y,w,h]…], perView,
  fps, blend }`. It becomes a camera-facing textured quad; the view is
  `quantize(angleToCamera - heading, views)` and the frame is `floor(t*fps) % perView`. A plain
  always-facing billboard is `views: 1`. `anchor` selects what `pos` means — `"center"` (default)
  or `"bottom"` (the quad's feet at `pos`, for standing creatures/props). This shared path replaces
  per-game billboard code (Doom things, Ultima Underworld creatures).

`blend` is `opaque` | `alpha` (alpha-masked, depth-write off) | `additive`
(`AdditiveBlending`, for translucent bodies). Separately, a level **GLB** may mark a material
single-sided (glTF `doubleSided:false`) so three.js back-face culls it — emitted by
`tools/lib/glb` `WriteTrianglesMat` via `TriGroup.SingleSided` (default double-sided), used for
ceilings and other one-way geometry. See `FORMAT2.md` for the full sprite-spec fields.

---

## 5. Writeup style (`games/<slug>/<slug>.md`)

A writeup is a neutral technical manual describing the final understanding of the game.

- Declarative present tense. No first person, no meta-narration about the analysis process.
- **No ordinal or comparison framing** — never "the first PSX title", "the repository's
  first MIPS target", "unlike the C64 tape", "in the same spirit as tools/x86".
- **No reverse-engineering history** — never describe a wrong path and then correct it, and
  never call something a "red herring" or say a lead "turned out" to be wrong. State only
  the final result.
- Skeleton: `# <Game> (<Platform>) — technical reference`, an **Image** line (filename,
  byte size, MD5), a **Contents** list, then the numbered Parts (below). Preserve all
  technical depth (addresses, formats, algorithms).

### Canonical Parts

A writeup follows the same progression — from the raw image down to running mechanics —
regardless of platform. The default Parts are:

- **Part I — The image.** The container format (tape/disk/cartridge/disc), its filesystem
  or catalog, the machine's memory map, and the executable/CPU binaries the image holds.
- **Part II — Boot chain.** From power-on or load through the loader, decryptor and
  relocation to the game's entry point (`main`). Where the platform has its own CPU core
  and boot oracle, the toolchain — disassembler, code tracer, execution core, machine
  oracle — is documented here or as a dedicated toolchain Part.
- **Part III — Program architecture.** The runtime memory map, initialization, interrupt
  handling and the main loop — the engine scaffolding, pinned by the oracle where a static
  read fans out into indirect calls.
- **Part IV — Graphics and data formats.** Tiles, tilemaps, palettes and sprites; 3-D
  models; and the level format(s).
- **Part V — Game mechanics.** Player physics, the object and enemy types and their
  behaviour, the levels, scoring and game states.
- **Part VI — Sound and music.** The sound-effect player and the music sequencer/synth.
  Games with light audio may fold this into Part IV.

Games subdivide when the content warrants it — e.g. splitting model / level / collision /
message formats into their own Parts — but keep the same image → boot → architecture →
formats → mechanics → sound order, numbering the Parts sequentially.
