# U0 — shared browser contract and capability audit

Completed audit: 2026-09-28. Implementation target: U1–U3; save files, automatic provenance capture and rendering replay remain U4–U7.

## Existing behavior and boot gaps

| Platform | Media/boot now | Display boundary | Existing browser inspection/state | U2 action |
|---|---|---|---|---|
| C64 | TAP with three local ROMs; authentic PAL reset; Fort scripts optional | VIC complete raster frame; presented crop 392×272 | Cycle/fetch provenance, volatile checkpoints | Automatically fetch pinned ROMs; generic keyboard/tape workflow |
| PS1 | 128 MiB whole-disc upload; ISO9660 PS-X EXE loader; BIOS HLE; Ridge handler supplied by frontend | Synthetic field per 250k run-loop ticks | Volatile checkpoints, GPU command/write capture | Sector-backed larger discs; image-derived BIOS interrupt setup; CUE validation |
| N64 | z64/v64/n64 normalization; CIC-6102 IPL3 CRC; 4 MiB handoff, NTSC/Pilotwings frontend gate | Synthetic VI per 750k ticks | Core hooks only; Go adapter has replay/snapshots | Remove image gate; boot chip/region detection and explicit unsupported cases |
| 3DO | Opera raw 2352 or ISO 2048; LaunchMe AIF; Portfolio HLE; NFS mirror at 0x42734 | DisplayScreen/OnDisplay | Core hooks only; Go adapter has cel replay/snapshots | Generic executable discovery/default settings; NFS profile only for exact recognized image |

No full-hardware BIOS boot is being claimed for PS1/3DO. C64 drive/disk formats and compressed disc formats require new implementations. A successful import is distinct from successful game execution. All structurally supported unknown images must be attempted without a title whitelist.

## Ownership and protocol

One dedicated Worker owns the selected core and mounted File objects. The DOM owns presentation and physical input devices. One session generation increments on boot/reset/replacement; async replies carry that generation and are ignored if stale. Commands carry a monotonically increasing request ID. Loading creates a new Worker so a long-running old core cannot delay a new session. Normal Pause must yield cooperatively, not terminate the Worker.

Commands: `load(files, options)`, `run`, `pause`, `reset`, `input(buttons, axes, keys)`, `step`, `profile(enabled)`. Later commands: `save`, `restore`, `capture`, `pixel`, `seek`. Responses: ready/capabilities, state+frame, profiling window, progress, diagnostic, command acknowledgment. Unsupported later capabilities are not rendered as working buttons.

Capabilities declare media formats, firmware, input mapping, frame units, snapshot/capture/replay availability and profiler bucket meanings. Unknown-image loading never depends on lesson metadata. Image identity/configuration, firmware identity and future lesson identity are separate.

Input arrives as absolute state (not browser autorepeat), with a sequence number, then applies at the next execution boundary. Keyboard/touch/gamepad feed one aggregator so releasing one source does not release another held source. Blur/hidden tab/disconnect releases the relevant source. Reset clears all sources and queued inputs. Hidden tabs pause. Browsing statistics or file dialogs must not inject gameplay keys.

U3 Pause is a cooperative stop; U5 upgrades it to automatic complete-frame capture. The U5 definition in the root plan is authoritative: next complete display interval, display-buffer mapping, immutable evidence, scratch replay, no silent skipping of blank frames, explicit pre-existing content/overflow. U3 must not claim provenance readiness prematurely.

## Media and firmware

Retain Files in the Worker. Read disc ranges using File.slice with a bounded sector cache, without media upload or a whole-disc WASM allocation. Validate headers/geometry/limits and selected CUE companions. Firmware fetches use same-origin versioned assets under `/emulators/firmware/c64/`; hashes and sources come from the existing C64 ROM manifest. No game media or saved machine snapshots are bundled.

User-supplied state files later bind to exact media/configuration/firmware hashes, never raw pointers or a linear-memory dump. Full image identity can be computed incrementally; unknown games are not held behind a metadata lookup. Exact compatibility profiles must not be applied based solely on a filename or weak header signature.

## Scheduling and profiling

The acceptance target is <250 ms normal command servicing, immediate UI acknowledgement, and explicit measured longest call. 3DO currently has a ~28-second call; its per-Run `seen`, spin ring, progress counters and cadence must survive slices. Preserve emulated behavior when refactoring. N64 synchronous RSP tasks are a separate possible long-call boundary.

Profile actual host execution time. Use exclusive buckets for totals/percentages; CPU/rest can be derived from total minus measured subsystems, labeled as derived. If a parent includes child time, label it inclusive and never add both to the total. Exclude pause, UI pacing and time between execution calls. Show reporting-window duration, milliseconds, percentages, calls/counters where meaningful, and profiling overhead as a limitation. Keep presentation/copy and disc I/O separate from emulator CPU cost. PS1's geometry unit is GTE (the user's GCE reference is interpreted as this unit).

Initial buckets: C64 CPU, VIC-II, CIA, SID, integration/rest; PS1 CPU/rest, GTE, GPU rasterizer, CD/DMA/I/O as instrumentable; N64 CPU/rest, RSP, RDP; 3DO ARM60/rest, Portfolio HLE excluding cel work, cel rasterizer, file/disc I/O. Avoid timing every instruction with expensive host callbacks: aggregate or sample narrow C64/CPU boundaries, label estimates, and validate overhead. A count-only placeholder is not a timing profiler.

## Acceptance evidence and compatibility matrix

Existing reference observations are in each `tools/platform/*/browser/results` directory: Fort TAP, Ridge Racer data track, Pilotwings cartridge, NFS disc. Preserve those as regressions. Native/Go drivers remain local tools, not deployed services.

| Case | Current evidence | Required extension |
|---|---|---|
| Four reference games | Boot + gameplay/input verified in earlier prototypes | Same local-import path, static app, profiles on/off parity |
| Unrecognized valid images | No generalized compatibility claim | Inventory locally available titles; record import, boot, scene separately |
| Malformed media | Some existing core checks | Truncation, invalid magic, unsupported layout, missing CUE companion |
| New scheduler | Not implemented | Same observations across slice sizes and pause/resume points |
| Input | PS1 browser mapping exists; hardware not verified | Shared aggregation tests and actual device test if connected |
| Hosting | User confirms automatic Cloudflare Pages deployment on repo push | Preserve `site/`; static emulator subtree; no fixture/server calls |

An unavailable physical controller is recorded as untested, never invented as a pass. Additional commercial images are not downloaded without user direction. Synthetic test guests can establish generic loading/execution behavior but do not count as demonstrated commercial-title compatibility.

## U1 design direction

Use four large, consistently lit generated console illustrations as the focus, arranged in a two-column desktop gallery. Quiet cool-paper palette: paper #f4f6fa, artwork #e9edf3, ink #283246, secondary #607087, accent #4d578f. System humanist sans typography, large plain headings, no marketing badges. Reuse the restrained emulator display/transport language; do not redesign Studio. The emulator route has a dominant display, compact file/transport row, controller help and an expandable timing table. At narrow widths stack controls naturally. U6+ owns the richer inspection layout.

Deployment root is the existing `site/`. Commit built browser JS/WASM under `site/emulators/cores/` so the existing no-build Pages deployment can pick up each milestone without dashboard changes. Source/build tools remain under `tools/`; a packaging script checks manifests and size limits. Preserve existing asset data and routes. Only the simple Emulators link changes the Studio introduction.
