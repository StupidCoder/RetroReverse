# C7: owned C64 production release

Release date: 2026-10-02. Bundle `d8bde8bba9eb279a0008`. C0–C7 are complete. The default C64 browser build
now compiles the owned CPU, board, PAL VIC, digital SID, dual-clock system and
1541. All fifteen other emulator binaries retain their previous hashes.

## Rollout and rollback

`tools/platform/c64/browser/build.py` delegates to the owned build. Its output
is packaged at `cores/c64/core.js` and `core.wasm`. No vendor headers are used.
The core reports `owned-c64`, production capabilities and mandatory portable
state identity binding. The worker hashes the actual WASM, firmware and medium.
D64/G64 need a local 16 KiB 1541 ROM; browser disks remain write protected.

The Session panel links to `site/emulators/c64/legacy.html`. That page loads
release `9dc547ff2dea4b78b50a`, including its original worker, knowledge and
vendored binary. It does not resolve current assets. `c64/rollback.json` pins
all 103 assets; packaging retains this release independently of its ordinary
previous-release retention. Release checks verify every rollback hash. The
legacy runtime's license and modified-source revision are retained alongside
its page. Source at `d66634b1` remains available for rebuilding/comparison.

There is no cross-core checkpoint conversion. Old saves need the legacy page;
new saves need the matching owned build. Both reject wrong core, firmware,
media or configuration before replacing the active session. Prepared lessons
are filtered by actual binary and firmware identity and regenerate locally
from physical inputs. New complete-state hashes are recorded in the original
game knowledge packages; no private media or derived save states are shipped.
The legacy page retains the recipes it was originally released with.

## Acceptance

- The C6 Linux CI fix passed GitHub Actions run 36968305822. The release CI
  allowance is now 25 minutes, allowing headroom for all sixteen systems plus
  owned native and UBSan fixtures rather than racing the previous 15-minute cap.
- Full owned suite: native, UBSan and WASM CPU, board, PAL video/digital SID,
  VIA/disk/dual-clock hardware, debugger, portable state, browser ABI and
  rendering; native/WASM traces match. Native-to-WASM state transfer passes.
- Real firmware/media: ordinary 1541 LOAD, SAVE, protected SAVE and replay;
  Giana G64 custom-loader upload, 20,168 IEC bytes, 20,086 verified payload
  stores, unpacker entry and dual-machine replay; Fort's 21,504 loader stores,
  2,251 extracted graphics bytes, gameplay and deterministic state replay;
  Elite's 52 initial vector stores against independently decoded tape bits.
- Shared release suite: knowledge/schema/export, all JS services and all
  sixteen platforms' native release checks pass. Debugger, tours and experiments
  now exercise the owned production WASM. Vendor-specific fixtures are replaced
  by owned browser/render/debugger/state suites, not retained as a dependency.
- Real browser, packaged worker: default selection, full rendering capture and
  provenance, step/record/cancel, save/restore, independent render and drive
  panels, drive run-to/step/global pause, D64/G64 head telemetry, bound disk
  firmware, suspension and failed-load recovery. Legacy save/restore passes;
  cross-core restore is rejected. Both `production` and `owned` caller aliases
  resolve to the owned core in the new application.
- Fort's production prepared lesson: cancellation preserves the session,
  cold terrain preparation reaches its tour, repeat preparation uses the
  verified cache, cold AI preparation passes unchanged setup guards. The real
  service test repeats original death deterministically and measures survival
  with the candidate patch for the same 200 frames, then restores the session.
- Elite's four-stop first-byte tour uses the production Code, Memory atlas and
  Storage panels. Memory snapshot busy periods now disable conflicting debugger
  and tour controls, then repaint their availability when the snapshot finishes.

## Performance method

`tools/browser/tests/c64-performance.{js,mjs,html}` measures the actual shipped
WASM in Node and the in-app browser. Each core physically boots the reference
Fort tape using its own hash-verified terrain recipe, advances to the next PAL
frame boundary and checkpoints that state. These are comparable gameplay entry
points, not identical hardware states across the two implementations.

Each measurement restores its anchor outside the timer. Normal execution runs
985,248 cycles (one PAL second) with event tracing, memory recording and render
capture disabled. Normal built-in writer history/profiling overhead remains
included: this is the application core, not a stripped CPU microbenchmark.
Capture enables the renderer's complete fetch/pixel/writer history for one
19,656-cycle display, including capture begin/end. One warm-up is discarded;
five samples and their median are reported. Boot, state copying, DOM drawing,
worker transport, raster replay and network fetches are excluded. No 1541 is
attached in this Fort benchmark, so it does not establish disk-game throughput.

The native harness applies the same input recipes and timing procedure to
Clang `-O2` builds. Native timing is additional evidence, not a substitute for
browser measurements. Emscripten is pinned to 4.0.16; JS builds use `-O2`.

Measurements on this macOS arm64 host, Apple Clang 17, Node 26.3.1 and the
Codex in-app browser (worker execution). Medians of five runs:

| Runtime | Owned gameplay / PAL second | Legacy gameplay / PAL second | Owned frame capture | Legacy frame capture |
| --- | ---: | ---: | ---: | ---: |
| Native | 148.90 ms (6.72×) | 77.63 ms (12.88×) | 18.68 ms | 2.52 ms |
| Node WASM | 161.64 ms (6.19×) | 149.05 ms (6.71×) | 18.56 ms | 4.45 ms |
| Browser WASM | 166.80 ms (6.00×) | 155.40 ms (6.44×) | 19.00 ms | 4.60 ms |

The browser's ordinary execution takes about 7.3% more host time while retaining
roughly sixfold real-time headroom on this machine. Capture takes about 4.1 times
as long, around 19 ms for a full PAL display. Native execution is substantially
slower relative to the legacy native build; it must not be used to predict the
browser ratio. These are one-machine, one-game measurements, not a mobile-device
or all-game guarantee. Detailed capture is a paused inspection operation. No
performance optimization or accuracy relaxation was made to improve these results.
[Recorded samples](C7-PERFORMANCE.json) accompany the reproducible harnesses.

Owned WASM: 355,150 bytes; legacy WASM: 418,600 bytes (uncompressed).
Owned SHA256: `d2fd1a1811e5094c01db842e6f5d93494908e8d6f1af4e86ac36811aeb809306`.
Legacy SHA256: `5cd52834dca1457e8a6715719a28b4fcb1de3de4a0df7b0510395cee4643f999`.

## Reproduce

Build/package as described in [the adapter guide](browser/README.md), then run:

```sh
python3 tools/browser/check.py
python3 tools/platform/c64/owned/check.py --emcc /path/to/em++ \
  --firmware-dir site/emulators/firmware/c64 --drive-rom /path/to/1541.rom \
  --giana-g64 /path/to/Great_Giana_Sisters_The.g64 \
  --elite-tap games/elite-c64/Elite.tap \
  --fort-tap games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  --fort-fixtures /path/to/independent-fort-fixtures
node tools/browser/tests/c64-performance.mjs \
  site/emulators/cores/c64/core.js \
  site/emulators/releases/9dc547ff2dea4b78b50a/cores/c64/core.js \
  games/fort-apocalypse-c64/Fort_Apocalypse.tap
python3 tools/browser/tests/c64-performance-native.py
```

For the optional native legacy comparison, extract commit `d66634b1` paths
`tools/platform/c64/browser`, `tools/browser/core` and `tools/browser/state`
into a scratch root, then pass `--legacy-root /path/to/scratch` to the native
harness. This does not restore a vendored dependency to the repository.

Serve the repository and open `/tools/browser/tests/owned-c64-browser.html`,
`/tools/browser/tests/prepared-browser.html?game=fort`, the same page with
`?game=elite`, and `/tools/browser/tests/c64-performance.html`. The last three
require the local reference tapes. Nothing is uploaded by these tests.

## Explicit follow-on work

This is PAL/TAP/D64/G64 acceptance for the tested cases, not general C64
compatibility certification. Full Elite loading, encoding changes, gameplay and
object-slot tours, and full Giana gameplay remain unvalidated. NTSC, cartridges,
additional drives, broader protection schemes, audible SID synthesis and G64
filesystem extraction remain future work. Fort's opcode patch is a measured
scenario-specific hypothesis, not a universal AI repair.
