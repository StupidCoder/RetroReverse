# M2 — C64 Code workspace and instruction debugging

Implemented 2026-09-30. This is a C64-only vertical slice. It does not add tours,
structured state panels, patches, or a verified reproduction of the Fort bug.

## User-visible behavior

The C64 player has a Code tab with the same live preview and keyboard input path
as Memory. Play, Pause, Next frame, and tape controls retain their meanings.
Instruction stepping is a separate control. The view shows A/X/Y/S/P, CPU banking,
cycle count, the fetch PC when available, interrupt/stall indicators, and mapped
instruction bytes. Live sampling is scheduled at five updates per second.

Opening Code, following PC, inspecting an address, or selecting a documented
function never executes instructions. A partial instruction is explicitly shown
as such; **Finish partial instruction** is an execution operation. Render now
also requires **Capture next display** rather than capturing on tab selection.
This prevents inspection navigation from unexpectedly advancing a paused machine.

The exact reference Fort tape exposes five entries from its single knowledge
package. Selecting an entry shows its applicability, evidence and annotations.
These are documented addresses, not automatically verified runtime functions.
The upward-probe annotation preserves the unresolved bug explanation. Generic
images still have raw disassembly and address breakpoints. No game-state meanings
or functions are guessed from filenames.

**Run to address** stops before the fetched opcode, with immediate-stop and
next-visit policies for the current PC. The binding records whether the target
maps RAM, ROM, or I/O when armed; reaching it under a different mapping reports
`mapping-invalid`. This deliberately exposes an address breakpoint, not a claim
that a documented function has been loaded or its byte signature verified.
Default limits are 9,852,480 PAL cycles and 10 seconds of wall time. Cancellation
leaves the reached machine state; a partial instruction may need normalization.

## Implementation

- `tools/platform/c64/browser/core/debug.inc`: separate opt-in debugger ABI.
  The normal `rr_run` loop has no debugger polling or per-instruction hook.
  Debug execution uses bounded calls through that existing device scheduler.
- `rr_debug_begin(mode,target,next_match)`: mode 0 normalizes a partial
  instruction, 1 steps from a fetch boundary, 2 runs to a bound address.
- `rr_debug_run(cycles)`: at most 10,000 cycles per call; the worker uses 1,000.
  Results are 0 slice exhausted, 1 normalized, 2 instruction retired,
  3 interrupt/reset entry, 4 target, 5 mapping invalid, 6 halted, -1 error.
  IRQ/NMI/reset entry stops before the first handler instruction. JAM is halted;
  RDY stalls advance devices without falsely retiring instructions.
- `rr_debug_snapshot(address)`: synchronous, read-only register/clock snapshot
  plus 256 mapped bytes and mapping tags. `-1` follows the fetch PC, or the last
  fetched instruction when currently between boundaries. CPU ports, device
  registers and open-bus color RAM are unavailable (`null`), never live-read.
  `nextPC` is the fetch candidate; `interruptPending` explicitly indicates that
  an entry sequence can run before that opcode. Unknown `lastExecutedPC` is null.
- `tools/cmd/dis6502export` exports the native decoder's official 151-opcode
  table. JS formatting includes `$55` EOR zero-page,X and wrapped relative
  targets. Unknown undocumented opcodes and unavailable/truncated bytes stop
  linear decoding; the UI never guesses the next alignment.
- `debug-worker.js` owns bounded jobs, explicit capabilities, session generation,
  request/job IDs, snapshot IDs, and observed mapping revision. Stale actions
  are rejected against the displayed snapshot and the current machine. Each
  accepted debug job produces one terminal result. `retired` is only reported
  for instruction-step jobs; it is not a fabricated run-until instruction count.
- `execution-gate.js` serializes play/frame-step, capture, seek, state save and
  memory inspection/recording across asynchronous awaits. Debug jobs exclude
  these operations at the worker boundary. Read-only live samples during play
  remain permitted. Existing explicit pause-and-capture/save controls retain
  their behavior and old worker messages; tours must use the debug protocol.
- The UI waits for a requested snapshot before enabling execution controls.
  Switching tabs does not cancel an active debug job. Pause/Cancel does.
  Candidate worker replacement still handles media load and state restore.
- `package.py --core c64` refreshes only C64 and verifies/reuses the other
  packaged cores. Both opcode and knowledge data are generated on packaging.

The mapping revision describes observed snapshot mapping changes, not a complete
bank-switch trace. Native debugger job state is ephemeral and invalidated by
machine initialization, synthetic preparation, checkpoint restore, and state
load; saved-state format and existing legacy PC status remain unchanged.

## Acceptance and scope

The same `debug-c64.cpp` corpus runs as native C++ and compiled WASM. It asserts
stop-before-execute, immediate/current-PC next-visit behavior, taken branches,
JSR/RTS, IRQ/NMI entry, safe peeks, ROM/RAM banking and invalidation, an executed
self-modifying sequence, RDY stalls, and JAM termination. `debug.mjs` tests all
151 official decoder entries, wraparound and unsupported bytes, gate tokens,
real WASM job execution, stale generations/snapshots, cancellation and budgets.

Existing C64 memory, pixel/device, raster, and serialized-state tests pass.
Knowledge, shell, input and state-container JS regressions and the Go knowledge
and mos6502 tests pass. Release checks verify pinned hashes and static network
behavior. No unrelated platform core was rebuilt or replaced.

Actual Chromium UI acceptance is in `tests/code-m2-browser.html`. It covers
preview keyboard input, live samples, pause, instruction stepping, safe peeks,
Code/Memory/Render navigation, and run cancellation. `?fort` additionally uses
the locally present reference tape to check exact matching and read-only
function selection. It does not load the tape into gameplay or prove the bug.
The browser report and measured disabled-debug benchmark are in
`results/code-m2/`. On the final bundle, browser pause/cancel observations were
21.4/20.7 ms (including a 20 ms test polling interval), with a 128 MiB WASM heap.
The interleaved disabled-debug core benchmark measured median 7.55M cycles/s
before versus 7.45M after, a 1.34% time overhead in this synthetic workload.
There are no new hooks in the normal execution loop; code layout and timing
noise can affect this comparison. Timings are samples, not worst-case guarantees.

Reproduce from the repository root:

```sh
go run ./tools/cmd/dis6502export -check
node tools/browser/tests/debug.mjs
clang++ -O2 -std=c++20 -Wno-c99-designator -Wno-address-of-temporary \
  tools/browser/tests/debug-c64.cpp -o /tmp/rr-debug-c64
/tmp/rr-debug-c64
# Compile the same test with em++ and -sENVIRONMENT=node for WASM parity.
node tools/browser/tests/code-m2-performance.mjs
python3 tools/browser/check-release.py
python3 -m http.server 8779 --bind 127.0.0.1
# Open /tools/browser/tests/code-m2-browser.html (optionally append ?fort).
```

Remaining milestones own structured state (M3), tours and verified runtime
applicability (M4), other processor adapters, and reproducible experiments.
Undocumented-opcode decoding, Safari/Firefox coverage, long-duration soak,
and actual Fort gameplay acceptance are not claimed here.
