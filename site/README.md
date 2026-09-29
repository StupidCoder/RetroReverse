# site/

Interactive companion website for the reverse-engineering write-ups.

It is a **no-build static site**: plain ES modules with an [import map](index.html) that pulls
PixiJS (v8) from a CDN, so there is nothing to install or compile. 3D viewers use three.js
the same way (Elite ships, Stunt Car Racer tracks).

## Run locally

Serve the folder with any static server (the import map + `fetch()` need HTTP, not `file://`):

```sh
cd site
python3 -m http.server 8000
# open http://localhost:8000/
```

## Deep links

The Studio mirrors its state into the URL, so every view is a copyable link and any game/level
opens directly:

- `?game=<id>` — the game (`sonic`, `fort`, `turrican`, `marble`, `sml`, `stuntcar`, `elite`).
- `?level=<slug>` — the level/asset, by a stable readable slug shown in the address bar
  (e.g. `?game=sonic&level=sky-base-act-3`, `?game=elite&level=cobra-mk-iii`). A numeric index
  is also accepted (and rewritten to the slug); `?asset=<n>` is the legacy index-only alias.
- `?objects=0|1` — force the **objects & enemies** overlay off/on (default on for games that
  have it).
- `?crt=0|1` — force the CRT filter off/on (default on).

The address bar updates as you switch games, levels, and toggles, so you can just copy the
current URL to share exactly what you're looking at.

## Layout

- `index.html` — the **Studio**: a single full-screen front-end for all games. A floating menu
  picks the game and asset; the selected viewer renders full-bleed, with display-layer toggles,
  a music player, an optional CRT filter, and a technical-details panel. This is the whole site.
- `src/studio/` — the Studio shell (`main.js`, `info-content.js`, `crt.js`, `camera.js`).
- `src/shared/` — the **shared 2-D level viewer** every tilemap game runs on (Sonic, Fort,
  Turrican, Marble's map view, SML): camera, tilemap renderer (sliced / baked / block
  strategies), overlay layers, the animation runner and palette effects. It consumes the
  **common level format** specified in [FORMAT2.md](../FORMAT2.md); per-game specifics live in a
  small `src/<game>/config.js`.
- `src/elite/`, `src/stuntcar/`, `src/marble/slopes.js` — the three.js 3-D viewers (wireframe
  ships, track ribbons, the slope height-mesh), outside the tilemap format.
- `public/marble/` — per-course `<course>.png` + `meta.json`. Regenerate from the disk with:

  ```sh
  cd "Marble Madness (Amiga)/extract"
  go run ./cmd/webexport
  ```
- `public/fort/` — per-level JSON (`level0/1.json`, `meta.json`) + `atlas-L0/1.png`.
  Regenerate (after extracting `FORT-fast-7000.prg` from the tape) with:

  ```sh
  cd "Fort Apocalypse (C64)/extract"
  go run ./cmd/webexport
  ```
- `public/turrican/` — per-world `atlas<N>.png` tile sheets + per-scene `world<W>_scene<S>.json`
  (row-major tile-index `cells`, `ntiles` for the flip threshold) + `meta.json`. Regenerate
  from the disk image with:

  ```sh
  go run turrican/extract/cmd/webexport   # run from the repo root
  ```
- `public/elite/ships.json` — decoded ship blueprints (vertices, edges, face normals).
  Regenerate from the extracted engine block with:

  ```sh
  cd "Elite (C64)/extract"
  go run ./cmd/webexport
  ```
- `public/sonic/` — exported data: `meta.json`, `shapes.json`, the `atlas_*.png` tile
  atlases, and `act01.json … act18.json`. Regenerate from the ROM with:

  ```sh
  cd "Sonic (GG)/extract"
  go run ./cmd/webexport "../Sonic The Hedgehog (Japan, USA).gg" "../../site/public/sonic"
  ```

## Deploy

Hosted on Cloudflare Pages. Pushes to the repository are picked up automatically by the existing Pages integration. The emulator area lives in `emulators/`; package its prebuilt WASM artifacts with `python3 tools/browser/package.py` from the repository root. Game images are supplied locally by visitors. C64 firmware is included under `emulators/firmware/c64/`, with pinned provenance in its manifest.

## Sonic level viewer

The map is a real tilemap rebuilt from the cartridge — each 32×32 block is baked once from
the zone's tile atlas, then placed as one sprite per cell (so PixiJS batches the whole level
and zoom is cheap). Drag to pan, scroll to zoom. Toggle layers:

- **Animation** — the rings (6 frames) and Green Hills flowers (2 frames) cycle at the
  Game Gear's cadence by re-baking only the animated block textures.
- **Collision shapes** — each block's surface height-profile (red) over the real tiles;
  non-solid blocks tinted blue (where Sonic falls through).
- **Objects** — enemies/items/bosses and Sonic's spawn marker.

> Note: the frontend was written in an environment without a browser or JS runtime, so it is
> verified at the data-contract level (the exporter output is checked against the
> `cmd/levelmap` render pixel-for-pixel, and the JSON/atlas indices are validated) but not
> yet run in a browser — please try it and report anything that needs fixing.

## Browser emulators

The existing site links to `/emulators/`, containing C64, PS1, N64 and 3DO WASM
emulators. Game images are selected locally; C64 firmware is hosted with the
app. Build and validation instructions are in [tools/browser/README.md](../tools/browser/README.md).
Use that directory's preview server when testing the emulator profiling headers.

## Memory inspector

Load a local game and choose **Memory** to pause at the current execution boundary.
The hex pane and physical-storage bitmap stay linked: click a bank or RAM area to
jump, use arrow/Page Up/Page Down keys in the hex pane, or enter a physical address
or explicit `rom-5:3430` bank offset. ROM rows show file offsets; RAM rows show their
base addresses. CPU aliases are labeled separately, including partial GG windows.

| Core | Physical snapshots | Recorded activity |
|---|---|---|
| GG | RAM, VRAM, CRAM, every ROM bank | CPU reads (including fetches), RAM and video-port writes |
| GB | WRAM, VRAM, OAM, HRAM, cartridge RAM and ROM banks | Not yet available |
| GBA | EWRAM, IWRAM, VRAM, palette, OAM, EEPROM, ROM chunks | Not yet available |
| C64 | Underlying RAM, BASIC/KERNAL/character ROM, color RAM, tape | CPU memory accesses and consumed tape pulses; optional fetches |
| Amiga | Chip/slow RAM, Kickstart, every ADF track/side | CPU and chip DMA accesses; MFM payload consumption attributed to sectors |

Other cores show an explicit unavailable message. Hardware I/O inspection is not
implemented using live register reads. Display-chip reads are not generally traced;
the workspace describes the coverage of each supported recorder.

**Record & run** advances the machine for the selected maximum duration, stopping
earlier at 524,288 events. Replay/scrubbing reconstructs historical bytes from an
immutable initial snapshot and actual writes, without changing the paused machine.
**Resume game** continues from the live endpoint, not the historical cursor.
Truncation is explicit. Recordings are bounded windows, not an unlimited rewind
history; a new snapshot or recording replaces the prior recording. CPU mapping
details describe the start snapshot, while events resolve physical banks at access time.

For Fort, start the normal `LOAD` / tape playback flow in Play, then use Memory to
record an interval. The tape bitmap encodes pulse duration and highlights consumed
pulses; selecting it opens the corresponding raw TAP bytes. ADF rows are grouped by
cylinder/side, with 11 sectors per track. Sector highlights identify the underlying
payload bits being consumed as encoded MFM, not a fictitious decoded-byte DMA copy.

Exact-hash annotations cover the documented Sonic GG and Fort tape revisions.
Sonic map spans are read from its actual act descriptors and split at ROM bank
boundaries. Unknown revisions still expose hardware regions without guessed labels.
Game-phase descriptions are documentation, not automatic phase detection.

Pause now freezes the machine. Memory includes a small live output preview and Play, Pause, and Next frame controls
for every emulator. Memory-capable cores refresh physical storage at up to 5 Hz;
C64, GG, and Amiga also highlight accesses since the last update. Live tracing
stops when leaving Memory, uses the existing bounded event buffer, and reports
overflow. ROM data and its overview are cached. This live view is separate from
historical recording. The C64 view includes tape controls and pulse position.

Opening **Render** automatically captures the next complete display interval.
Returning to Render keeps the existing capture until the machine advances.

Run `node tools/browser/tests/memory.mjs` for media-free model/service checks and
`node tools/browser/tests/memory-wasm.mjs <platform> <local-image>` for optional
WASM validation. The native GG/C64/Amiga memory tests are included in
`python3 tools/browser/check.py` alongside the existing rendering checks.
