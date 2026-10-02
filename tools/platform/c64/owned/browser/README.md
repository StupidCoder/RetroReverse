# Owned C64 browser adapter

The production C64 uses this independently authored hardware core. The shared
worker, debugger, memory, rendering, tours and experiments use its C ABI.
See [the rollout report](../C7-RELEASE.md) for benchmarks and checkpoint rollback.

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

Build and package production, then serve the repository root:

```sh
python3 tools/platform/c64/browser/build.py --emcc /path/to/emscripten/em++
python3 tools/browser/package.py --core c64
python3 -m http.server 8765 --bind 127.0.0.1
```

Open `/site/emulators/c64/`. The normal page and source development page both
select the owned core. `c64Core=owned` and the older `production` alias resolve
to the same packaged core. Scratch builds remain excluded from packaging.
`/site/emulators/c64/legacy.html` loads the frozen previous bundle for old
checkpoints. It never resolves the new core. Neither core accepts the other's
states; restarting a prepared lesson builds a checkpoint for the selected core.

The normal worker binds actual WASM/firmware/TAP identities before permitting
checkpoint operations. Save files include the owned backend in configuration;
system suspension retains it. A failed candidate load cannot change the active
backend or capabilities. Prepared lessons are listed only when their core and
firmware identities match, so old-core starts are not offered here.

Run `/tools/browser/tests/owned-c64-browser.html` on the server for actual-browser
acceptance with a generated TAP and local bundled firmware. It exercises the
worker, DOM panels, memory recording/cancellation, debugger, state transport,
identity rejection, rendering capture/provenance, independent drive inspectors,
D64 head telemetry, synchronized drive stepping/breakpoints, global pause and
failed-load recovery. It uses authored drive firmware and disk bytes.

The source UI accepts TAP, D64 and G64. For disks, select a local 16 KiB 1541
ROM under optional firmware, then load the image. Disk media is write protected
in this browser profile; the original selected file is never modified. Select
**1541 drive** in any viewport for drive disassembly, registers, VIA state, IEC
levels/transition count, head/bit position, instruction stepping and breakpoints.
Multiple drive panels have independent snapshot identities. Advanced drive
step-over/out and write watchpoints are not exposed. Both CPUs share one timeline.
Storage shows logical D64 sectors or raw G64 bytes/half-tracks, with a live head
ring. G64 filesystem decoding and physical angular alignment are not inferred.

Optional real-media UI checks:
`/tools/browser/tests/prepared-browser.html?core=owned&game=fort` and
`?core=owned&game=elite`. These read the private local images named by the tests;
no media or checkpoint is uploaded. They exercise fresh preparation, cancellation,
local-cache restore and the existing guided lesson/experiment panels.

## Execution and inspection contract

- `rr_init` copies BASIC/KERNAL/character ROMs from `rr_input`, powers the C64 and
  removes old media. `rr_tape` loads validated TAP v0/v1. For a drive session,
  call `rr_drive_rom` with 16 KiB before executing any C64 clocks, then `rr_disk`
  with D64/G64 bytes and an explicit write-protection flag. Inputs are copied.
- Without drive firmware, only the board executes. With it, each C64 step also
  executes the intervening 1 MHz drive edges through the owned scheduler.
  `rr_drive_status` reports live head, bit, motor and LED state. The drive debugger advances ordered scheduler edges and shares global pause.
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
  worker provides decoded tape region 5 (7 with a drive attached). Drive RAM and
  firmware are snapshot-only regions; no drive access heatmap is claimed. Bounded access records retain CPU
  fetch/read/write events and tape-head movement; overflow is counted. Writes
  under visible ROM belong to physical RAM. Inspection does not tick hardware.
- The 392×272 RGBA display copies the owned VIC framebuffer. SID oscillators,
  noise and envelope state execute, but no audio samples are generated.
  Audio remains explicitly unavailable. Rendering capture and pixel provenance
  are supported through the live owned VIC observer (see below).

`rr_prepare` is an explicitly synthetic test/debug start, not a game loader.
Real-media acceptance uses physical keyboard input, KERNAL and original loader
code. The synthetic ROM-execution guard prevents silently substituting it for
an authentic boot.

## Checkpoint transport

Before portable save/load, the host must write 228 bytes to `rr_input` and call
`rr_state_bind`: seven raw SHA256 digests (core, BASIC, KERNAL, characters,
1541 firmware, original TAP, original disk), then a little-endian configuration
word. An absent optional input uses a zero digest. The current test host uses
bit 0 for drive presence and bit 1 for writable disk media. The browser worker
hashes the actual current inputs, including local drive firmware and disk. It
binds standalone TAP sessions with configuration zero and protected drive
sessions with configuration one.
Media replacement or pulse editing invalidates the binding and internal slots.

The adapter wraps the [hardware codec](../PORTABLE-STATE.md) with format magic
`0x41343643`, version 2, drive mode, current instruction PC/start cycle,
synthetic-start status, logical pressed keys, actual fetched instruction bytes
and interrupt nesting. This retains writer attribution after partial-instruction
restore without re-decoding overwritten RAM. It appends an FNV-1a checksum;
the nested hardware payload has its own CRC32 and identity checks. Parsing is
bounded to 12 MiB + 4096 bytes and occurs before atomic hardware restore.
This development format is distinct from the old core's checkpoints.

Sixteen in-memory slots are tied to the current media generation. Portable
restore clears slots and observer history. Trace IDs/settings, memory history,
breakpoint jobs and profiler counters are not machine state; observing a run
must not change its checkpoint digest. The existing JavaScript services retain
host input queues and their own tour/experiment ownership separately.

## Rendering contract

The live owned VIC reports real matrix/color/graphics/sprite fetches and pixel
composition. Each captured pixel retains the source bytes, fetch cycles,
bank/pointer controls and exact CPU writer versions. Same-value writes remain
distinct; pixels drawn in a CPU-write clock use the preceding register version.
Instruction bytes come from actual CPU reads, including self-modifying code.
Unknown pre-checkpoint writers remain unknown after restoring a checkpoint.

Each visible line freezes RAM, color RAM and VIC/bank registers. Background and
sprite previews run a separate owned VIC against that frozen memory. The actual
screen accumulates historical pixels; it is not regenerated from end-of-frame
RAM. Charset previews also use the selected line's memory. Seeking, clicking
pixels and replaying the capture do not tick or mutate the live machine.

Capture is bounded to 524,288 references, 65,536 writes and 65,536 control states.
Overflow and incomplete frames are explicit. Tests cover all five valid video
modes, character ROM, sprite priority, charset switching, changed glyphs,
register versions, partial-instruction restore and complete checkpoint equality
with observation on/off. Real Fort gameplay additionally checks all 272 lines,
its charset split, source/writer values and final output equality.

## Prepared lessons and acceptance

The game knowledge packages retain the original production recipes and add
`terrain-owned`, `enemy-ai-owned` and `loader-prefix-owned`. The UI filters by
actual core and firmware hashes. Recipes use physical keyboard/joystick inputs
and bounded execution; they contain no game RAM or firmware. Their complete
checkpoint digests are checked before fresh or cached use.

To reproduce or deliberately regenerate them after an owned build change:

```sh
RR_C64_CORE="$PWD/site/emulators/cores/c64/core.js" \
  node tools/platform/c64/owned/tests/prepare-lessons.mjs \
  site/emulators/firmware/c64 games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  games/elite-c64/Elite.tap
# Add --write to regenerate package identities after reviewing acceptance.
```

The shared service suites pass partial instructions, interrupts/RDY, atomic
edits, deterministic replay, same-value writes, cancellation, overflow, stale
requests and complete session/input rollback. The authored actual-browser
suite covers worker integration and independent panel ownership.

Fort boots through KERNAL/Novaload and passes the three-stop terrain tour
(215/40 interval writes). The owned AI recipe stops at cycle 126,553,519,
110 natural AI entries after the earlier candidate. This SID-noise phase lets
the unchanged teleport/spawn guards pass. The original branch repeats exactly
and records terrain contact/death; the $55-to-$A5 candidate survives the same
200-frame interval. No game-state guard or hardware behavior was changed to
force this result. It remains a scenario-specific repair hypothesis.

Elite separately passes authentic KERNAL boot, its four-stop first-byte lesson,
all 52 initial vector stores against independently decoded tape bits, and
fresh/cached prepared-start replay. Later fastloader stages, protocol changes,
gameplay and object slots remain unvalidated. Giana retains the C5 gate: 512
uploaded drive bytes, 20,168 protocol bytes, 20,086 payload stores and unpacker
entry, with dual-machine replay. Full Giana gameplay is not certified.

C7 switches the default after native/WASM/browser acceptance and separate
execution/capture measurements. The frozen legacy release and its license are
retained; the vendored runtime sources are no longer part of the build.
