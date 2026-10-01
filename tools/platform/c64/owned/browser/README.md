# Owned C64 development browser adapter

This is C6 work in progress. It builds the owned hardware as an ES module with
an Emscripten ABI compatible with the existing debugger, memory, tour,
experiment and prepared-start services. The production worker can select it explicitly for development; the default
backend and shipped WASM remain unchanged.

## Build and test

Use an explicit scratch output directory:

```sh
python3 tools/platform/c64/owned/browser/build.py \
  --emcc /path/to/emscripten/em++ --out /private/tmp/rr-owned-adapter
RR_C64_CORE=/private/tmp/rr-owned-adapter/core.mjs node tools/browser/tests/tours.mjs
RR_C64_CORE=/private/tmp/rr-owned-adapter/core.mjs node tools/browser/tests/experiments.mjs
RR_C64_CORE=/private/tmp/rr-owned-adapter/core.mjs node tools/platform/c64/owned/tests/browser-services.mjs
```

`owned/check.py` builds and checks this adapter automatically, in addition to
native, UBSan and WASM ABI tests. Its optional `--elite-tap` and `--fort-tap`
inputs also run the corresponding real-media browser-service acceptance.
Firmware/game/checkpoint bytes are never fetched or committed by these tests.
Without `RR_C64_CORE`, the shared tour/experiment tests still use the production
core. The test host hashes actual WASM, firmware and media bytes; it does not
replace any execution or checkpoint implementation.

## Local browser integration

Build the opt-in artifacts, then serve the repository root:

```sh
python3 tools/platform/c64/owned/browser/build.py \
  --emcc /path/to/emscripten/em++ --out site/emulators/cores/c64-owned
python3 -m http.server 8765 --bind 127.0.0.1
```

Open `/tools/browser/owned-c64.html?c64Core=owned` on that local server.
Without the query parameter, the same source UI uses the production backend.
Only the fixed `owned` and `production` choices are accepted. Generated owned
artifacts are ignored by Git and excluded from release packaging. The build
writes a SHA256 manifest; the worker verifies the WASM before initialization.
Missing development artifacts produce a load error and retain any live session.

The normal worker binds actual WASM/firmware/TAP identities before permitting
checkpoint operations. Save files include the owned backend in configuration;
system suspension retains it. A failed candidate load cannot change the active
backend or capabilities. Prepared lessons are listed only when their core and
firmware identities match, so old-core starts are not offered here.

Run `/tools/browser/tests/owned-c64-browser.html` on the server for actual-browser
acceptance with a generated TAP and local bundled firmware. It exercises the
worker, DOM panels, memory recording/cancellation, debugger, state transport,
identity rejection and failed-load recovery. The worker also rejects unsupported
render capture directly; every independent Render panel disables that action.
This integration currently accepts TAP only: drive firmware/disk selection and
Storage head telemetry still need UI/worker integration.

## Execution and inspection contract

- `rr_init` copies BASIC/KERNAL/character ROMs from `rr_input`, powers the C64 and
  removes old media. `rr_tape` loads validated TAP v0/v1. For a drive session,
  call `rr_drive_rom` with 16 KiB before executing any C64 clocks, then `rr_disk`
  with D64/G64 bytes and an explicit write-protection flag. Inputs are copied.
- Without drive firmware, only the board executes. With it, each C64 step also
  executes the intervening 1 MHz drive edges through the owned scheduler.
  `rr_drive_status` reports live head, bit, motor and LED state. Drive-specific
  debugger controls and Storage UI wiring are still pending.
- `rr_run` accepts at most one million C64 clocks per slice; `rr_debug_run`
  accepts at most 10,000. JavaScript services own yielding, budgets and
  cancellation. No host timer is used as a source of emulated time.
- The debugger stops at pending opcode-fetch boundaries. It handles partial
  instructions, RDY-held fetches, IRQ/NMI entry, next-match breakpoints,
  step-over and conventional JSR/RTS step-out. Abnormal returns and mapping
  changes produce explicit stop results. A write watch stops on an actual bus
  write, including the dummy write of a read/modify/write instruction.
- Snapshot bytes are live, side-effect-free peeks. CPU ports and I/O are marked
  unavailable rather than acknowledging registers. Checked RAM edits validate
  the entire batch before changing anything, and edits at PC affect the next
  real opcode read.
- Memory regions expose underlying RAM, three ROMs and color RAM. The shared
  worker provides decoded tape region 5. Bounded access records retain CPU
  fetch/read/write events and tape-head movement; overflow is counted. Writes
  under visible ROM belong to physical RAM. Inspection does not tick hardware.
- The 392×272 RGBA display copies the owned VIC framebuffer. SID oscillators,
  noise and envelope state execute, but no audio samples are generated.
  `rr_capabilities` explicitly reports audio, render capture and pixel
  provenance as unavailable. Those rendering APIs are not exported.

`rr_prepare` is an explicitly synthetic test/debug start, not a game loader.
Real-media acceptance uses physical keyboard input, KERNAL and original loader
code. The synthetic ROM-execution guard prevents silently substituting it for
an authentic boot.

## Checkpoint transport

Before portable save/load, the host must write 228 bytes to `rr_input` and call
`rr_state_bind`: seven raw SHA256 digests (core, BASIC, KERNAL, characters,
1541 firmware, original TAP, original disk), then a little-endian configuration
word. An absent optional input uses a zero digest. The current test host uses
bit 0 for drive presence and bit 1 for writable disk media. The browser worker hashes the actual current inputs and binds standalone TAP
sessions with configuration zero; drive sessions remain a test-host API.
Media replacement or pulse editing invalidates the binding and internal slots.

The adapter wraps the [hardware codec](../PORTABLE-STATE.md) with format magic
`0x41343643`, version 1, drive mode, current instruction PC/start cycle,
synthetic-start status and logical pressed keys. It appends an FNV-1a checksum;
the nested hardware payload has its own CRC32 and identity checks. Parsing is
bounded to 12 MiB + 4096 bytes and occurs before atomic hardware restore.
This development format is distinct from the old core's checkpoints.

Sixteen in-memory slots are tied to the current media generation. Portable
restore clears slots and observer history. Trace IDs/settings, memory history,
breakpoint jobs and profiler counters are not machine state; observing a run
must not change its checkpoint digest. The existing JavaScript services retain
host input queues and their own tour/experiment ownership separately.

## Acceptance and remaining gaps

The shared worker suites pass tours and experiments with authored programs,
including atomic edits, deterministic replay, same-value writes, cancellation,
overflow, stale requests and complete session/input rollback. Native ABI tests
also cover IRQ/NMI entry, RDY, nested calls through interrupts, bad returns,
mapping changes, self-modification, keyboard combinations and dual-machine
partial-state replay.

The pinned Elite TAP passes its real KERNAL boot, four-stop first-byte lesson,
all 52 initial vector stores against independently decoded tape bits, and a
fresh/cached prepared-start replay through the existing service. The test
creates an owned-core recipe identity and expected digest in memory only.
It does not update the shipped recipe or certify later loader stages/gameplay.

The pinned Fort TAP boots through Novaload and passes all three terrain-tour
stops, including 215 and 40 writes. Its **existing enemy-AI experiment is not
compatible yet**: after satisfying the start guard, the native teleport produces
player/camera X=209/188, while the package requires 53/34. The game chooses among
four random drop points; a core/timing-specific recipe needs investigation.
The acceptance test retains the guards and verifies complete rollback on this
known rejection. It does not claim a successful original/patched AI comparison.

C6 still needs render capture/pixel provenance, Storage/drive-debug integration,
and regenerated accepted lesson/experiment recipes. C7's default switch remains
blocked on those acceptance gates.
