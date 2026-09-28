# Browser Frame Observatory — unified Pages release

Implementation update (2026-09-28): U0/U1 are committed and pushed. U2 import
and boot paths are implemented; broader-game compatibility acceptance remains
open (Elite imports but renders incorrectly; no second PS1/N64/3DO title was
available). U3 provides shared controls, subsystem timings and resumable 3DO
execution. Physical-controller validation remains open because none is attached.
See `tools/browser/docs/U1-ACCEPTANCE.md`, `U2-ACCEPTANCE.md`, and
`U3-ACCEPTANCE.md` for measured checks and limitations. U4 now provides portable states with native/WASM continuation checks; see
`tools/browser/docs/U4-ACCEPTANCE.md`. U5 provides bounded complete-interval capture and cancellation; see
`tools/browser/docs/U5-ACCEPTANCE.md`. U6 is in progress.


Status: implementation started, 2026-09-28. U0 contracts are recorded in `tools/browser/docs/U0-CONTRACTS.md`; subsequent acceptance results are tracked per milestone.

This is the next release target after the four C++/WASM prototypes. Preserve the existing RetroReverse site in `site/`, including its extracted game assets, explanations, viewers and deep links. Add the emulator area alongside it; do not replace the current site or reorganize its content in this release. The emulator area supersedes only the separate emulator development pages. The earlier [educational-debugger plan](EDUCATIONAL-DEBUGGER-PLAN.md) remains the record of the Fort Apocalypse lessons and performance investigation; preserve its existing work and evidence.

## Product target

One static Cloudflare Pages deployment contains the existing RetroReverse site plus the Commodore 64, PlayStation, Nintendo 64 and 3DO emulators. The existing RetroReverse home page gains a simple **Emulators** link to `/emulators/`. That separate landing page presents four recognizable console/computer images with names and a return link to the main site. Selecting a system opens its emulator and prompts for a local game image and any companion files. C64 firmware is hosted with the application and loaded automatically; users do not need to locate those ROMs. User-selected game images stay on their computer. Existing hosted extracted assets and explanations remain available as before.

Every system uses the same basic flow: select media → run → pause and capture → click a pixel → step through how the frame was rendered → resume. Run, Pause, Reset, keyboard/on-screen/physical gamepad input, performance statistics, and downloadable/loadable save states are available independently of game recognition. A recognized image may gain a lesson later; an unrecognized image does not lose any generic debugger features.

The goal remains useful educational exploration, not realtime gameplay. Existing emulation approximations remain visible and compatibility is tested rather than presumed.

**Important scope distinction:** accept every structurally valid image in the supported media formats, without a title/hash whitelist. This is not a promise that the current cores correctly execute every game ever released. Unknown titles must be attempted through a generic boot path; unsupported hardware, formats and boot requirements must produce useful diagnostics. Broad compatibility is continuing emulator work, not something achieved by removing a frontend check.

## Current foundations and gaps

| System | Reuse | Work needed for this target |
|---|---|---|
| C64 | Local TAP/ROM import, actual reset/tape execution, volatile checkpoints, captured VIC fetches and pixel-to-memory/writer history | Shared app integration, persistent state format, broader tape testing, rendering timeline presentation |
| PS1 | Browser gamepad mapping, volatile checkpoints, GPU command/write capture and pixel history | Remove Ridge-specific boot assumptions, portable states, exact pause-to-display capture, deterministic rendering replay/scrubbing |
| N64 | VR4300/RSP/RDP WASM core, observation hooks, Go snapshot/replay adapter | Generic cartridge boot/CIC coverage, C++ state serialization, browser capture, pixel queries and replay |
| 3DO | ARM60/Portfolio HLE/cel renderer, sliced local-file reads, Go snapshot/provenance adapter | Remove NFS-specific boot/timing assumptions, resumable execution, C++ state serialization, capture including offscreen bitmaps, browser replay |

The existing PS1 inspector can search up to four fields for drawing and show a drawing buffer distinct from the live front buffer. That is useful prototype evidence, but does not satisfy this release's automatic next-frame contract. N64 and 3DO browser inspectors are not yet wired. C64 and PS1 user checkpoints are currently volatile, not downloadable state files.

The main uncertainties are generic PS1/3DO booting, 3DO execution latency, complete snapshots, and rendering history across buffer swaps and earlier frames. Resolve these before investing in a sophisticated inspector layout.

## Architecture and behavioral contracts

### One static app, four independent cores

Keep the existing site at `/` and retain its current asset paths and hash routes. Add `/emulators/`, `/emulators/c64/`, `/emulators/ps1/`, `/emulators/n64/` and `/emulators/3do/` under the same origin. Build one deployment directory starting with the existing site, then add the shared emulator UI, four WASM modules and worker entry points, console artwork, small metadata and C64 firmware. Use paths that also work in local static previews; do not point the published site at localhost development ports. Load only the selected core. Keep each machine owned by its browser Worker, with versioned messages and generation/request IDs so stale responses cannot affect a new session.

The shared adapter contract covers capabilities, local media/firmware mounting, reset, bounded execution, input, statistics, snapshots, capture, pixel queries and rendering replay. Platform adapters define frame boundaries and event meanings; the shared UI does not pretend that a C64 raster fetch is a PS1 GPU command.

No Python service, Go binary, R2 bucket, database or Cloudflare Function is required for this target. Development report-writing endpoints become local JSON downloads or test-harness features. Native builds and the Go oracles remain development/CI tools.

Build WASM with pinned Emscripten in CI or a reproducible local build, then deploy the static artifacts. Do not require visitors to compile anything. Pages currently limits individual assets to 25 MiB; check every packaged artifact against the current limit, and keep full game images, generated game snapshots and private emulator test fixtures out of the new emulator deployment output. Explicitly include the hosted C64 firmware. Preserve the existing site's extracted game assets and explanations; this exclusion is not a cleanup instruction for `site/public/`. [Cloudflare Pages limits](https://developers.cloudflare.com/pages/platform/limits/).

### Local media and recognition

Retain selected `File` objects in the Worker. Use offset/length reads and bounded caching for discs; a whole-disc buffer is not required. Companion tracks are selected together and resolved within that selected file set, never by arbitrary filesystem access. A page reload normally asks the user to reselect media; persistent file handles or copying images into browser storage are not requirements for this release.

Recognition is an optional, incremental identity check against a small metadata registry. It must not block an unknown image from booting. Separate three things: format/firmware validity, emulator compatibility configuration, and lesson availability. Necessary verified compatibility profiles can be keyed to an exact image, but never apply Fort/Ridge/Pilotwings/NFS settings to unknown titles. Lessons do not configure hidden game-memory patches or replace execution.

Firmware is its own dependency. Host the C64 BASIC, KERNAL and character ROMs as versioned static assets, using the existing pinned ROM identities and recording their source, hashes and accompanying notices. Fetch and validate them automatically when C64 is selected; provide clear retry/error behavior if a download fails. A local ROM override may remain an advanced option, not a prerequisite. Save states record the exact firmware hashes and reload the matching hosted version. PS1 and 3DO currently use HLE; any later need for additional firmware is a separate explicit compatibility decision. N64 boot/CIC support must be stated explicitly.

### What Pause means

The control changes immediately to **Capturing next frame…**. At the next safe execution boundary, freeze controller input for the capture, retain a start checkpoint and complete one whole platform-defined display interval with provenance enabled. If the request arrives mid-interval, finish that interval first and capture the next complete one. The machine stops at the captured interval's end. Show which frame/field was captured and the emulated time advanced.

Use platform boundaries: complete VIC raster frame for C64, the declared video field/frame boundary for PS1/N64, and display-buffer presentation for the current 3DO HLE. Preserve interlace/field metadata instead of calling every event a full progressive frame. A repeated or blank display still counts: do not silently skip forward until something interesting is drawn.

Record the displayed buffer and its mapping, not whichever buffer happens to receive writes. A pixel inherited from the capture's starting state is explicitly labeled as pre-existing; never invent a writer. A bounded earlier-frame history can recover its producer where available, with the actual history range shown. Complete ancestry back to boot is outside this release.

Captured evidence is immutable. Rendering steps run in a scratch replay machine; they do not change the live machine at the capture end. **Run** resumes from that end, regardless of the inspection cursor. **Reset**, loading a state or replacing media invalidates stale capture references. Branching live execution from an intermediate rendering event is later work.

No display progress, unsupported behavior and trace overflow have explicit outcomes, not infinite waits or fabricated completeness. Cancellation restores the retained pre-capture checkpoint and pauses without a new capture. A worker restart is an emergency escape, not the normal pause implementation.

## Milestones

| ID | Outcome | Depends on |
|---|---|---|
| U0 | Capability audit and precise platform contracts | Existing prototypes |
| U1 | Emulator area linked from the preserved RetroReverse site | U0 |
| U2 | Unrestricted local imports and generic boot paths | U0, U1 |
| U3 | Shared transport, gamepads, statistics and responsive execution | U1, U2 |
| U4 | Complete, versioned save/load states on all systems | U2, U3 |
| U5 | Automatic next-frame capture and common evidence model | U3, U4 |
| U6 | Pixel inspection on all four platforms | U5 |
| U7 | Rendering replay, stepping and scrubbing | U4–U6 |
| U8 | Pages release and cross-platform acceptance | U1–U7 |
| Later | GPU/PPU state, memory watches and authored explanations | Released exploration foundation |

### U0 — Audit capabilities and lock the contracts

Deliverables:

- Record actual accepted image formats, firmware dependencies, title-specific assumptions, frame boundaries, capture hooks and snapshot coverage for each core.
- Define adapter messages and capabilities, session generations, state identities, input ordering and the Pause/capture/replay behavior above.
- Establish repeatable boot-to-scene fixtures for the four current games, synthetic hardware tests and an initial unfamiliar-title compatibility matrix.
- Decide which PS1/3DO generic boot facilities are missing. If generic HLE requires substantial work or supplied firmware, record that explicitly as implementation work in U2; do not disguise a whitelist removal as compatibility.

Acceptance: every requested user action maps to an implemented capability or a named milestone gap. Capture boundaries and the supported-format table are written down before UI integration starts.

### U1 — Shared app and visual console launcher

Deliverables:

- Preserve the existing `site/` home page, asset viewers, explanations and navigation. Its only initial product integration change is a simple **Emulators** link in the landing-page introduction, opening `/emulators/`; add a return link from that page. Build the destination before exposing the link so there is no broken public route. Do not embed emulators in asset viewers or redesign the existing site.
- Four image-based cards on the separate emulator landing page: C64, PS1, N64, 3DO. Use consistent original illustrations or appropriately sourced local artwork with descriptive alternative text; no dependency on externally hosted images.
- Each card opens a stable emulator route with a local-file prompt, large display and a shared transport/status area. Direct URLs and browser navigation work.
- Lazy-load the selected WASM core. Replace localhost ports and root-level fixture URLs with deployment-safe platform paths.
- Produce a single static output directory combining the existing site and the emulator subtree, including the C64 firmware assets. Keep existing prototypes available while migrating. The user confirmed the existing repo deploys automatically to Cloudflare Pages on push. Preserve that setup and commit/push each completed milestone; `site/README.md`'s GitHub Pages description is stale.

Acceptance: the existing site and representative extracted-asset/explanation links still work unchanged. The new Emulators link and return navigation work. A clean static server serves all emulator routes; console selection, automatic C64 firmware loading and game-file prompts work without any emulator backend or game fixture request. Only the chosen core loads. Keyboard navigation and loading/error states work.

### U2 — Local media, firmware and generic boot

Implement as four explicit platform submilestones; do not mark U2 complete when only file picking is unified.

- **U2a, C64:** generic TAP import and automatic hosted BASIC/KERNAL/character ROM setup (with an optional local override); tape motor/transport and keyboard boot interaction must work without Fort's schedule. Document supported TAP versions and PAL/NTSC scope.
- **U2b, PS1:** raw data-track and supported ISO/BIN geometry; add CUE plus companion-track selection with explicit validation for unsupported layouts. Replace Ridge Racer's fixed interrupt configuration with generic guest-driven initialization or an explicitly supported firmware path. Do not present silent misinterpretation of CD layouts as successful loading.
- **U2c, N64:** normalize z64/v64/n64 byte orders, detect region/boot requirements, remove the Pilotwings hash gate and handle supported CIC variants. Unsupported boot hardware gets a precise error, not the Pilotwings seed applied to everything.
- **U2d, 3DO:** local random-access disc reads, supported raw-sector/ISO layouts and generic executable discovery. Remove global dependence on NFS's VBL mirror, startup input lock and fixed boot script. Develop generic OS/timing behavior; exact-image compatibility profiles remain optional and visibly identified where still necessary.

Start with the media families above. Additional families, such as C64 D64 disk images requiring a drive implementation or compressed disc formats requiring new decoders, are separate compatibility extensions. Unsupported formats receive explicit explanations rather than a known-title rejection.

Acceptance: known and unknown images take the same import path. At least one unfamiliar available title per platform is loaded and its boot outcome recorded; at least one additional title per platform must reach a useful scene before claiming demonstrated broader compatibility. Loader acceptance and execution failures are reported separately. Known reference games still boot, and large discs remain bounded random-access reads over a long session. No game bytes are sent over the network.

### U3 — Transport, controls, performance and execution scheduling

Deliverables:

- Consistent Run, Pause, Reset and platform-appropriate gamepad controls; keyboard and on-screen fallback plus physical browser gamepads, including analog sticks, dead zones, device disconnect and mapping visibility.
- Reset recreates the machine using the same mounted media and firmware/configuration, releases input, and invalidates history. It does not require file reselection during the same session.
- Route physical, keyboard, touch and scripted inputs through one ordered emulated-time queue. Release inputs on focus loss; block accidental game input while inspecting or typing into UI fields.
- Show per-subsystem wall time and percentages, following framedbg: N64 CPU/RSP/RDP, PS1 CPU/GTE/software GPU, 3DO ARM60/Portfolio/cel rasterizer/disc I/O, and C64 CPU/VIC/CIA/SID where measurable. Distinguish exclusive buckets from nested inclusive totals; never present instruction counts as measured time. Also show display updates/s, CPU throughput, emulated versus wall time where meaningful, allocated WASM memory and execution/capture mode. Label synthetic fields honestly. Separate running performance from capture/replay cost.
- Refactor long synchronous execution into resumable slices while preserving scheduler state. This is particularly important for the 3DO `Run` method and N64 RSP/RDP work: arbitrary calls with smaller budgets are not automatically equivalent.

Acceptance: all controls work on all adapters; physical gamepad and keyboard are actually exercised. Core state/output matches the reference under different slice sizes and interruption points. Target normal command servicing below 250 ms on the reference Mac, with immediate UI acknowledgment; measure worst-case loading calls and document any unresolved exception. The present 28-second 3DO call is not accepted as the final interaction design. Reset/reload/disconnect cannot leave held input behind.

### U4 — Save and load complete machine states

Deliverables:

- Shared **Save state** download and **Load state** file selection, plus an optional in-session quick slot. Save/load works at safe boundaries during loading as well as gameplay. Loading restores a paused machine; it does not blindly resume a held physical button.
- Explicit versioned serialization, not a dump of WASM linear memory or native pointers. Include platform, format/core compatibility version, exact media/firmware identities, configuration and integrity checks.
- Capture CPU pipelines, devices, interrupts, graphics state and transfers, clocks, input queues, tape/disc positions, save media, and platform-specific state. For 3DO this includes HLE tasks/items/signals and stream positions; for N64, RSP/RDP and pending work.
- Reconstruct references and callbacks on load. Keep immutable disc/ROM bytes outside the state file and rebind to the selected local game media and matching hosted firmware. Verify identities with bounded incremental hashing; require reselection if necessary game files or an explicitly chosen local firmware override are absent. Standard C64 firmware should be fetched automatically, including after a page reload.
- Load transactionally: malformed, truncated, oversized, incompatible or wrong-media states leave the current machine intact. Initially require an explicitly compatible core version; automatic migrations are not promised.

Acceptance: on every core, save at multiple loading/render/gameplay points, advance with recorded inputs, reload and rerun to identical state/frame/evidence fingerprints. Download, reload the page, reselect game media, automatically reload hosted C64 firmware, import the state and reproduce continuation. Cross-check native/WASM serialization where supported. Record storage size and restore latency. Existing C64/PS1 volatile snapshots alone do not pass this milestone.

### U5 — Pause automatically captures the next frame

Deliverables:

- Implement the shared Pause contract, a capture start checkpoint and end state, and an immutable evidence object identified by session and capture generation.
- Use typed platform events under a shared envelope: event order, emulated time, source, target buffer/address, old/new values and available origin evidence. Distinguish a command-submission PC from the instruction that constructed its data.
- Preserve display mapping, source memory versions and sufficient intermediate data to replay the capture without consulting later live memory.
- Retain a bounded history/checkpoint budget. On overflow, report incomplete evidence; allow targeted recapture/replay where possible. Pixel queries and capture progress remain cancellable.
- Keep continuous execution cheap: detailed per-pixel recording is activated for capture rather than indefinitely for all gameplay.

Acceptance: pauses requested at different points yield a complete declared display interval whose final pixels match execution without tracing. Blank/repeated frames, double buffering, loading stalls, cancellation and trace overflow are covered. Resume from the capture end matches uncaptured execution with equivalent inputs. No stale evidence survives a reset or state load.

### U6 — Click a pixel and explain its recorded contributors

Deliverables:

- A shared compact inspector: selected coordinate/color, ordered contributors, selected event and available source evidence. Correctly map clicks through scaling, borders, aspect ratio, resolution changes and field/display transforms.
- **C64:** raster-time character/bitmap/sprite fetches, palette/priority decisions and historical memory/writer links where available.
- **PS1:** primitives, transfers, fills and CPU/DMA VRAM writes; texture/CLUT sources, masking/blending and displayed-buffer identity.
- **N64:** RDP events, textures/palette, depth/coverage/blending decisions exposed by the model, and the path from RSP/RDP submission to display. Represent modeled scanout separately from framebuffer writes.
- **3DO:** CCB/cel events, source texels/PLUT, PIXC blending and direct framebuffer writes, including offscreen-to-onscreen composition such as the rear-view mirror.
- Include rejected candidates when the renderer can observe them. Distinguish them from successful writes and from unobserved history. Do not claim unsupported hardware effects or universal dataflow/taint tracing.

Acceptance: sample pixels across all reference games and synthetic cases for overwrite, blend, rejection/priority, texture/palette lookup, CPU-written pixels, offscreen composition and unchanged background. Reconstruct the modeled final color from the captured evidence. Missing ancestry is labeled rather than attributed to a later RAM value.

### U7 — Step through rendering without changing the paused game

Deliverables:

- First/previous/next/last event and a rendering progress slider. Select a pixel contributor to jump to its event; optionally restrict navigation to that pixel's contributors.
- Platform-appropriate stepping units: raster events for C64, GPU command events for PS1, RDP events for N64 and cels/CCBs plus relevant writes for 3DO. Basic pixel-contributor navigation is separate from CPU instruction stepping.
- Deterministic scratch replay from the capture checkpoint, with sparse intermediate checkpoints for responsive seeking. Display the incremental result and highlight changes; make pre-existing buffer contents explicit.
- Replay CPU/device effects as needed to reproduce rendering dependencies. A recorded command list alone is insufficient if commands read memory modified later in the interval. Isolate HLE objects, input, clocks and host side effects from the live machine.

Acceptance: forward, backward and arbitrary seeks produce the same intermediate result on repeated replay. End-of-capture output equals the captured frame. Interleave inspection and Run without changing the live continuation. Test offscreen sources and textures modified during rendering. Target cached steps below 100 ms; show progress and cancellation for expensive uncached seeks.

### U8 — Release the unified application on Pages

Deliverables:

- One reproducible build/deploy configuration for the existing site plus emulator subtree, with pinned toolchain, asset manifest, matching worker/WASM versions, hosted firmware hashes, correct MIME types and cache behavior. Existing site URLs and new emulator routes work on direct navigation and refresh. Do not prune or relocate the current extracted assets and explanation texts.
- Emulator network audit: app/code/artwork/metadata and hosted C64 firmware downloads only; no user-game-image upload or download, localhost dependencies, required fixture endpoint or server-side emulator execution. Reporting stays local. Existing asset viewers retain their current asset/CDN requests; this is not a network-policy rewrite of the old site.
- Automated core/replay tests and browser acceptance on Chromium, Firefox and Safari where the necessary APIs are supported. Exercise local media selection, multi-file discs, state-file round trips, gamepad input and failures; do not infer browser support from the native tests.
- Measure cold load, time to useful scene, pause/capture latency, replay responsiveness and memory growth. Run at least a 30-minute session and repeated reset/capture/load cycles on each platform. Disc and trace caches must remain bounded; fix growing session allocations rather than merely hiding their counters.
- Publish the tested format/firmware and game-compatibility matrix, inherited accuracy limits and unsupported features. Missing experimental capabilities are not shown as working controls.

Acceptance: starting from a clean browser session on the deployed Pages URL, a user follows the existing site's Emulators link, selects a console and their local game media, automatically obtains hosted C64 firmware, reaches a tested scene, operates controls, saves and restores across reload, pauses into an inspectable frame, clicks a pixel, steps its rendering, and resumes on all four systems. No local application or emulator server runs. This is the release gate, beyond merely seeing four launch cards.

## Later work: site integration, state, memory and lessons

Keep these out of the first release's inspector scope, while giving every captured event a stable identity and replay cursor so they can be added without changing the core architecture.

- Brainstorm the eventual relationship between the existing asset/explanation site, emulator exploration and guided lessons. Shared game pages, navigation redesign and content cleanup are deferred; the simple link is sufficient for this release.
- GPU/PPU/VIC/RDP/cel state at the selected rendering step, with register/value differences from the previous step.
- Memory views and watches synchronized to that same historical moment, with clear address spaces, banking and source texture/palette links. Writes and CPU-writer navigation are separate optional depth levels.
- Known-image lesson packages containing narrative, event predicates, annotations and infographic playback. Preserve the Fort work, but do not require lesson generation to use an unknown game. Known compatibility settings and lessons remain separate concepts.
- Brainstorm and prototype the workspace with real captures before choosing a fixed layout. Candidate: stable large display and transport, a thin timeline, and one contextual side panel switching among Pixel, State and Memory. Use progressive disclosure, not simultaneous banks of registers.
- Validate a no-page-scroll desktop workspace at an agreed reference viewport (initial proposal: 1440×900 CSS pixels), with browser zoom/accessibility and smaller-screen alternatives. Allow contained lists/tabs rather than hiding overflowing information. Decide which values remain visible while stepping and how users follow pixel → texture → memory without losing context.

Do not lock in the full GPU/memory layout in U1. U6/U7 need a usable compact inspector, but the richer state/watch UI deserves the later brainstorming the user requested.

## Implementation order and checkpoints

Start with U0, then build U1 and the four U2 submilestones. Establish U3/U4 for all systems before declaring a shared debugging experience. Prove U5–U7 first on PS1, where command capture and snapshots already exist; adapt the same contracts to C64 raster rendering, then N64 and 3DO. Each platform must pass its own acceptance tests before the milestone is complete across the product.

Useful review points are: (1) static shell plus unrestricted imports, (2) reliable transport and persistent states on all systems, (3) the first complete pause → pixel → rendering-step workflow, and (4) the four-platform Pages release. These are progress checkpoints, not requirements to pause work for approval after every milestone.

No calendar estimates are committed before U0/U2 establish the generic-boot gaps. Packaging is comparatively small; compatibility, resumable execution, snapshots and faithful provenance are the substantial work.
