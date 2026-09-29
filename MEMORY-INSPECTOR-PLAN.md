# Memory inspector proposal

Status: initial review implementation. Shared workspace, GG/GB/GBA snapshots,
Sonic/Fort annotations, C64 pulse recording/replay, and Amiga ADF/DMA inspection
are implemented. GG also records CPU reads and RAM/video-port writes. See
[usage and coverage](site/README.md#memory-inspector).

Remaining extensions from this proposal: GB/GBA access recording, additional
core adapters, checkpoint-backed long histories, historical CPU-mapping changes,
automatic game-phase annotation activation, and adjustable pane dividers. The
current UI labels recording coverage and mapping-at-start explicitly. Historical
navigation reconstructs a bounded recording; it never changes the live machine.

## Experience

Add **Memory** beside **Play** and **Render**. A paused machine opens a synchronized hex view, bitmap atlas, and searchable region list. The atlas is the main navigation surface: RAM, video memory, cartridge RAM, every physical ROM bank, and loaded media appear in one scrollable view, separated by three blank raster rows and labeled headers. Unmapped ROM banks remain visible.

Use the existing light emulator styling: paper `#f4f6fa`, ink `#283246`, muted `#607087`, selection `#4d578f`. Add read blue `#2479ad` and write orange `#c66025`. Keep Avenir/system UI for controls and a fixed-width system font for addresses/bytes. Alignment is left; panes have resizable dividers. The bitmap is the distinctive element, with quiet surrounding controls.

## Desktop mock layout

```text
Game Gear · Sonic the Hedgehog                 Play  Render  [Memory]
Paused · current machine state                [Record & run] [Resume]

[Physical storage ▾]  [Go to address or label…                 ] [Go]
Activity: [Reads ✓] [Writes ✓] [Fetches ☐]  [Recent window ▾]

┌─ Hex · Work RAM ─────────────────┬─ Memory atlas ─────────────┬─ Regions ─────────────┐
│ CPU $D200 · RAM +$1200           │ [Fit all] [1 byte/pixel]   │ [Search regions…    ]│
│ Address  00 01 02 03 … 0F  ASCII│ Grayscale bytes + activity│ [All ▾] [ROM/RAM ▾]  │
│ $D200    .. .. .. .. … ..  ....│                           │                      │
│ $D210    .. .. .. .. … ..  ....│ Work RAM · 8 KiB          │ Work RAM             │
│ $D220    .. .. .. .. … ..  ....│ ░░▒▒░░░░░░░▒▒▒░░░░░░░░░░ │   Game state         │
│ $D230    .. .. .. .. … ..  ....│ ░░░░░┏━━━━━━━━┓░░░░░░░░░ │   Mapper shadows     │
│ $D240    .. .. .. .. … ..  ....│ ░░░░░┗━━━━━━━━┛░░░░░░░░░ │                      │
│   …                              │       ↑ hex viewport      │ Cartridge            │
│                                 │                           │   Special-stage map  │
│ Selected byte                   │ ROM bank 00 · 16 KiB      │   Collision tables   │
│ Value / region / CPU aliases    │ ▒▒░▒▒▒░░░░░▒▒░░▒▒▒░░░▒░░ │   Tiles / palettes   │
│ Last read / last write / PC      │                           │                      │
│ [Follow source] [Open reference]│ ROM bank 01 · 16 KiB      │ Selected region      │
│                                 │ ░▒▒▒░░░░▒▒▒░░░░░▒▒░░░▒░ │ Range, bank, purpose │
│                                 │   …all banks in this atlas│ Evidence / reference │
└─────────────────────────────────┴───────────────────────────┴──────────────────────┘

Recorded activity  [▶ Replay] [← Event] [Event →] [Speed ▾]
Reads             ▁▂▃▇▃▁▁▂▅▂▁▁▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
Writes            ▁▁▂▇▅▂▁▁▂▅▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
                  ├── selected interval ──┤         ↑ cursor
```

Byte placeholders and activity shapes above are illustrative, not captured Sonic data. Default desktop widths: hex 46%, atlas 32%, regions 22%. Give the workspace more width than the current 1,180px content cap on large screens. At medium widths move Regions into a drawer; on phones stack hex and atlas, preserving shared selection. Keep game display available through Play; an optional small preview must not reduce usable hex width.

### Interaction details

- Hex defaults to 16 bytes per row plus ASCII, grouped every four bytes; virtualize rows. Wheel/Page Up/Page Down scroll addresses, arrows select bytes, and Go accepts a CPU address or an explicit region/bank offset.
- The atlas uses byte-value grayscale by default. Optional 1bpp/2bpp/4bpp views can follow later; tile decoding already belongs to Render. Read/write overlays remain independent of byte values.
- Start with 256 bytes per raster row at byte zoom. At fit-all scale, aggregate contiguous bytes and state the scale in bytes/pixel. Preserve short access bursts using activity counts/maxima rather than averaging them away. Include minimum visible footprints for tiny regions, with their actual sizes labeled.
- The hex viewport is an outlined interval on the atlas, wrapping across raster rows when necessary. A separate caret marks the selected byte. The outline stays visible regardless of read/write overlays; a magnified inset/tooltip resolves subpixel intervals.
- Clicking or dragging inside a map segment navigates the hex view to that physical range. Fit-all clicks select a bucket and offer zoom for byte precision. Headers and blank separator rows never resolve to bytes. Hex navigation scrolls the atlas enough to reveal its marker without repeatedly resetting zoom.
- Region selections highlight their full ranges, navigate the hex view, and show the source reference. Overlapping labels are allowed. Search includes name, address, bank, and description. Offer keyboard navigation and text equivalents for map selection/activity; do not rely on color alone.
- Physical storage is the default. An optional CPU-address view answers “what does the CPU see here now?” and links to the backing physical region. Show inaccessible/I/O/open-bus locations explicitly rather than inventing bytes.

## Storage model

Represent each visible segment by a stable region ID, kind, physical offset, size, optional bank, CPU aliases, and read-only status. Address selection is `(region ID, offset)`, not an ambiguous integer. Map layout is a separate transform with reversible pixel/range mapping and non-addressable gaps.

| Machine | Atlas contents and addressing |
|---|---|
| GB | Work RAM, video/OAM memory, available cartridge RAM banks, and every actual cartridge ROM bank; expose mapper selection and aliases separately. Enumerate actual core/model capabilities. |
| GG | Work RAM, VRAM/CRAM, cartridge RAM when present, and all 16 KiB ROM banks. Mapper slot/fixed-window rules are supplied by the core, including mirrors. Sonic's documented image is 16 banks. |
| GBA | EWRAM, IWRAM, VRAM, palette/OAM, save storage, firmware where exposed, and one physical cartridge ROM divided into manageable display chunks. The existing bus maps ROM through 0x08/0x0A/0x0C windows with different wait states; these are aliases, not three copies or switched banks. Add true bank metadata only for a supported mapper that needs it. |
| C64 | Full underlying RAM, firmware and color RAM as distinct storage; separately expose current CPU mapping and I/O. Include TAP media underneath RAM. RAM under ROM must remain inspectable. |
| Amiga | Chip/other implemented RAM, firmware, and ADF sectors ordered by track/side/sector, with track boundaries. Include DMA as an activity source. |
| Other emulators | Same capability-based workspace; add adapters incrementally. An unsupported access-history capability is labeled explicitly; snapshot-only support must not imply tracing exists. |

### Tape and disk in the same atlas

Tape is a typed media segment. At overview scale encode pulse duration by intensity; show a consumption cursor and consumed range. At closer zoom show individual pulse widths. Selecting a pulse displays duration, pulse index, emulated cycle, and corresponding TAP byte offset; the hex pane shows those raw encoded TAP bytes. A pulse is not a decoded game byte. Maintain the variable-length TAP encoding index so selection round-trips correctly. Fort block labels are an optional verified decoder/annotation layer.

ADF segments show sector payload bytes. Activity must be tied to the actual emulated drive/track stream and DMA transfer, not host file loading. Preserve the distinction between raw encoded track data copied to RAM and decoded sector payload. Highlight a logical sector only when the track-to-sector mapping is established; otherwise show track activity with the limitation stated. Never infer a direct sector-to-final-RAM copy from timing alone.

## Recording and replay

A snapshot answers “what is here?”; an activity recording answers “how did it get here?”. Provide both:

1. **Pause / snapshot:** freeze at the next safe execution boundary and take side-effect-free memory peeks. Navigating Memory while running requests that pause. No hidden frame advancement to prepare memory.
2. **Record & run:** start at a paused boundary, advance the real machine with observation enabled, then stop automatically at a bounded duration/event budget or when requested. Preserve a checkpoint and input log. C64 additionally offers pulse and verified block boundaries.
3. **Replay recording:** while the live game remains paused, animate recorded activity and reconstructed historical bytes. Explicitly label historical position. Scrubbing must update bytes, mapping state, media position, and access overlays to the same cursor. Resume returns to the live paused endpoint; branching from history is outside the first release.

Use blue for reads, orange for writes, and split marks for both; fetches have a separate optional layer. The default recent window moves with the emulated cursor; cumulative mode shows activity over a selected interval. Freeze the overlay when paused rather than fading by wall-clock time. Show source/agent (CPU, DMA, video, media), PC when meaningful, logical address, physical address/bank resolved at event time, access width/value, cycle and stable sequence ID. Record writes even when they do not change the byte value. Snapshot diffs are not a substitute for access tracing.

For long recordings keep coarse summaries plus a bounded detailed window and checkpoint/replay support. Use native/WASM counters and compact binary batches; avoid per-access JavaScript calls or JSON. Fetch hex pages on demand; retain immutable ROM/media once; update only dirty RAM/map tiles. Detailed tracing starts disabled. Expose retained range, dropped events, and truncation; replay must never silently bridge a missing interval. Benchmark overhead before enabling any always-on history.

### Fort Apocalypse acceptance story

Start recording **before** tape playback (or load an existing verified lesson checkpoint). Replay the pulse cursor advancing while loader writes populate successive RAM blocks. Then inspect decompression: the documented level-0 compressed map is `$7000–$762A`, the decompressor is at `$8CDB`, and the expanded terrain occupies `$0503–$2D02`. Reads should highlight the compressed input while writes spread through the output. Distinguish these reference labels from dynamically verified event evidence.

For a selected write show the instruction and nearby reads/pulses. Strong source→destination claims require existing provenance or a verified decoder, not merely neighboring reads and writes. Reuse the Fort lesson chapters and evidence infrastructure where compatible rather than duplicating their game-specific knowledge inside the core.

## Reverse-engineering annotations

Create a versioned, game-specific region manifest derived from checked-in references/extractors. Match the exact local image by hash and size; do not attach precise labels solely by filename. A mismatched revision can remain unlabeled, with a clear explanation. Hash locally; no image upload.

Each label has: stable ID, name, region ID, physical start/length (or multiple spans), optional bank-relative/CPU aliases, kind, description, source document anchor, confidence, and optional phase/activation predicate. Separate symbols with unknown size from bounded regions; do not guess an extent. Keep historical/banked labels tied to the appropriate mapping and phase.

Seed packages:

- **Fort:** tape blocks, loader code, compressed terrain/scanner data, decompression outputs, charsets, sprite data, and documented game RAM. Many locations change purpose between loading and gameplay.
- **Sonic GG:** mapper shadows, game/player state, map streams, block/collision tables, tiles and palettes across banks. Extract exact spans from the existing extraction code and reference tables. For example the documented special-stage map starts at ROM offset `$1EE30` (bank 7, offset `$2E30`); its full extent must come from the decoder. Existing `sonic.annotations.txt` is explicitly a home-bank CPU configuration and cannot be imported as a flat all-bank symbol table.

Longer term the manifest can link directly to Studio assets and Render resources. Keep that outside the first release.

## Integration plan

Current integration points:

- `site/emulators/ui-shell.js` defines the workspace shell. Workspace navigation is currently instantiated inside `render-workspace.js`; move shared ownership to the app/shell so Memory is a peer.
- `site/emulators/app.js` and `worker.js` own session/request handling, execution, and capture. Today `pause` invokes `captureNext()`. Separate pause/snapshot from the Render capture action while retaining a clear Render capture path.
- Add `memory-workspace.js`, `memory-atlas.js`, `memory-hex.js`, and annotation loading with the existing plain ES-module conventions.
- Add versioned worker requests for capabilities/regions, paused snapshot/page reads, atlas summaries, activity record/stop/query, and historical seek. Scope replies to session, snapshot/recording, generation, and request; discard stale results after load/reset/seek. Use transferables and cancellable bounded work.
- Add narrow core exports/adapters for physical enumeration and side-effect-free peeks. Never implement inspection using normal hardware read handlers: registers and EEPROM can change state on reads. Read bytes without producing observation events.
- C64 `browser/core/core.cpp` already contains pulse positions, checkpoints and selected trace events; audit its coverage/filters and expose them through the shared worker. Existing Render observation hooks in other cores are useful infrastructure, not evidence that every CPU read/write is recorded.
- Change source core files and build/package through the existing browser tooling; do not hand-edit generated `site/emulators/cores/*/core.js` bundles. Respect generated-core source ownership for GB/GG/GBA.

## Delivery order and validation

| Stage | Deliverable | Acceptance |
|---|---|---|
| 1 | Snapshot UI + GG physical adapter; shared protocol designed for all machines | Pause is stable; all Sonic banks visible; click/hex/marker round-trip; bank aliases correct; reads leave machine state unchanged. |
| 2 | GB/GBA physical adapters + initial Sonic/Fort annotation manifests | Real sizes/mirrors/save banks; large-ROM fit/zoom; exact image matching; label spans validated against decoder output. |
| 3 | C64 tape atlas + bounded read/write recording + historical replay | Fort pulse→loader writes and compressed-input→output progression visible; bytes/cursor/mapping agree at each seek; native/WASM evidence comparison where available. |
| 4 | Amiga memory/ADF atlas + drive/DMA activity | Sector/track selection maps correctly; actual DMA activity distinguished from decoded payload and host loading. |
| 5 | Broaden core adapters and polish tracing performance | Per-core capability audit; honest unsupported states; measured budgets on representative large memories. |

Validate address transforms at segment edges/gaps, mirror/bank changes during capture, same-value writes, read-versus-fetch filters, trace overflow, cancellation, stale replies, state reloads, and side-effect-free peeks. Browser checks cover linked scrolling, zoom, keyboard access, narrow layouts, and multi-megabyte images. Performance goals: immediate cached selection, responsive/cancellable recording, bounded memory use, and no material normal-play regression with tracing disabled; measure before committing to numerical budgets.

First review decisions: the three-pane layout, physical-atlas default, explicit Record & run workflow, and this staged rollout. Memory editing, arbitrary reverse execution outside retained recordings, and automatic source/destination inference are future extensions.
