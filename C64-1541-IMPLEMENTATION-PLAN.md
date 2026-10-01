# Owned C64 and 1541 implementation

## Goal and rollout

Replace the vendored C64 hardware implementation with independently authored
C++20 hardware models, compiled from the same sources for native tests and WASM.
First target: PAL C64, one 1541, standard firmware, TAP v0/v1 and D64. SID audio
output is deferred; observable oscillator/noise/envelope behavior is required.
Fort Apocalypse reads $D41B for graphics and gameplay, so substituting a host RNG
or constant is not acceptable. Native and WASM replay must reproduce those reads.

The shipped core remains the default until the replacement meets the acceptance
gates below. Keep the old core as an optional comparison tool during development;
it is not the specification. Do not relabel copied third-party hardware code as
our implementation. Public hardware documentation, independent conformance data,
our Go analysis code and existing observations provide reference material.
Game images, firmware and game-derived checkpoints remain local test inputs.
Commit and push each completed milestone.

The Go CPU is useful reference material, not a cycle-accurate starting point:
`tools/cpu/mos6502/cpu.go` currently approximates instruction timing and halts on
BRK. Its analysis callers must remain compatible while the new core develops.

## Architecture

- `tools/platform/c64/owned/`: independent core, native/WASM harness and tests.
- One NMOS 6502 engine, one externally visible bus transaction per tick; the
  board implements the 6510 port and VIC bus arbitration. The 1541 reuses the
  CPU engine with its own memory map and clock.
- Explicit machine state: registers, partial instructions, interrupt latches,
  device pipelines, bus lines and clock phases. No wall clock, hidden RNG or
  pointer-dependent snapshots. Inspection reads never execute device reads.
- Board scheduler uses integer clock phases for PAL C64 and the independently
  clocked drive. Avoid assuming their CPU clocks are identical. IEC is shared
  open-collector line state, with transitions attributable to either machine.
- Device events carry machine/CPU identity, emulated time, address, value and
  source. Separate opt-in tracing from the normal execution path.
- Preserve the existing browser transport and panel contracts where possible;
  adapt the existing debug, pixel provenance, recording and replay hooks.
- Save-state format identifies the core and schema; existing vendored-core
  binary snapshots are incompatible and must never be loaded by reinterpretation.
  Prepared lesson recipes can regenerate compatible checkpoints.

## Milestones and acceptance

### C0 — Specification and implementation boundaries

Record scope, provenance, evidence and compatibility limits. Specify native and
WASM builds outside the shipped core; no deployment switch. Tests need no game
media. Mark progress below only after the corresponding checks have run.

### C1 — CPU foundation

Implement documented NMOS instructions with real read/write cycles, indexed
page-cross behavior, zero-page wrapping, stack operations, NMOS indirect JMP,
read-modify-write dummy writes, binary/decimal arithmetic, BRK/RTI and reset.
Define RDY stalls and IRQ/NMI input/state contracts. Test partial-instruction
replay and native/WASM equivalence. Unsupported opcodes stop explicitly.
Use independent single-step vectors with bus activity, plus hand-authored
multi-instruction and control-line cases. State precisely which interrupt edge
cases and undocumented opcodes remain; C1 alone is not a playable C64.

### C2 — CPU compatibility and C64 board

Complete stable undocumented instructions and characterize unstable variants;
validate interrupt polling, CLI/SEI/PLP delays, branch quirks, NMI/BRK hijacking,
reset during execution and VIC-compatible read stalls against independent tests.
Implement 6510 DDR/port, memory banking, ROM/RAM/color RAM, both CIAs, keyboard
matrix, joystick ports and tape wiring. Boot real KERNAL/BASIC without ROM-call
traps. VIA is a distinct device, not a renamed CIA. Gate: CPU conformance ledger,
real boot/keyboard and deterministic CIA/tape tests in native and WASM.

### C3 — VIC-II and SID-visible behavior

PAL fetch schedule, badlines, sprite DMA, bus arbitration, character/bitmap and
multicolor modes, scrolling, borders, raster IRQs and collisions. Capture source
fetches at the moment they occur. Implement SID register writes and observable
voice 3 oscillator/noise/envelope evolution; test checkpoint continuity and
Fort's actual $D41B uses. Gate: display and raster tests, authentic Fort loading,
title and gameplay; retain explicit limits for untested hardware corner cases.

### C4 — 1541 and normal disk loading

Drive CPU, 2 KiB RAM/ROM map, two 6522 VIAs, IEC bus, motor/head mechanics,
track density, GCR bit stream, sync detection and byte-ready behavior. Run real
drive ROM; synthesize standard encoded tracks from D64 sectors. Implement
write-protect and local disk-write state deliberately, with dirty image state
included in snapshots. Gate: ROM-driven directory and file loading, checksummed
payloads, head moves, media changes and repeatable C64/drive replay. No
file-loading trap may stand in for this gate.

### C5 — Fastloaders and drive investigation

Validate uploaded drive routines and timing-sensitive transfers using a named
media corpus. G64 parsing was brought forward into C4 for the supplied Giana
image; validate nonstandard track behavior here. D64 does not preserve arbitrary
protection/track layouts. Expose drive CPU disassembly,
breakpoints, memory, VIA state, head/bit position and IEC transitions. Global
pause freezes both machines; stepping either CPU keeps the other synchronized.
Gate: ordinary and selected custom loaders, drive-side breakpoints and a
repeatable inspection sequence explaining a real transfer.

### C6 — Browser and research parity

Integrate the owned core behind an explicit development selection. Pass current
Code/Memory/Rendering/Storage fixtures, partial-instruction debugger behavior,
recording, provenance, experiments and cancellation. Regenerate prepared lessons
with new core identity. Fort and Elite are separate acceptance cases; record
loader progress/failures rather than inferring success from a screenshot.

### C7 — Default switch

Switch only after C2–C6 acceptance and a documented native/WASM/browser report.
Benchmark uninstrumented gameplay and instrumented capture separately. Retain a
rollback path, explain checkpoint incompatibility, and remove the vendored
runtime dependency only after the new shipped bundle is verified. NTSC,
cartridges, more drives and broader protection compatibility are follow-on work.

## Evidence sources

- Existing Fort boot/loader/gameplay and pixel provenance tests under
  `tools/platform/c64/browser` and `tools/browser/tests`.
- [SingleStepTests/65x02](https://github.com/SingleStepTests/65x02): independently
  generated instruction states and bus cycles. Pin downloads and retain license
  attribution for any redistributed vectors; a passing sample is not exhaustive.
- [Klaus Dormann tests](https://github.com/Klaus2m5/6502_65C02_functional_tests):
  sustained instruction/interrupt/decimal validation, subject to each test's scope.
- [VIC-II description](https://www.cebix.net/VIC-Article.txt).
- [1540/1541 technical manual archive](https://www.zimmers.net/anonftp/pub/cbm/schematics/drives/new/1541/tech/).
- [VICE image formats](https://vice-emu.sourceforge.io/vice_17.html).

## Progress

C0 complete: scope, architecture, acceptance gates and isolated implementation
location are recorded.

C1 complete: independently authored documented-instruction CPU, native/WASM
build harness, pinned offline instruction/bus vectors and microstate replay
checks. Both targets pass 9,664 vectors / 38,749 bus cycles with trace digest
`6a229904`; native UBSan passes. See `tools/platform/c64/owned/README.md` for
reproduction and the precise CPU limitations at that milestone.

C2 complete: 237 supported CPU encodings, explicit unstable/JAM diagnostics,
interrupt-edge tests, 6510 banking, both CIAs, keyboard/joystick and TAP wiring.
Native/WASM pass 15,168 independent vectors / 66,181 bus cycles (digest
`1d597cc1`) and the sustained Klaus Dormann test (30,646,176 instructions).
Real KERNAL/BASIC boots to READY and executes PRINT 2+2 entered through physical
keyboard switches; CIA/tape timing and replay checks pass. Native UBSan passes.
See [the compatibility ledger](tools/platform/c64/owned/COMPATIBILITY.md) for
identities, evidence and limits. VIC raster registers and SID were scaffolds at
this milestone.

C3 complete: independently authored PAL VIC fetch/render pipeline, badlines,
sprite DMA and CPU bus arbitration, text/bitmap/multicolor/ECM display, scrolling,
borders, raster IRQs, sprite priority/collisions and captured source fetches.
SID oscillator/noise/envelope state supplies actual OSC3/ENV3 reads and survives
replay. Authentic Fort TAP boot passes without ROM traps or injected game RAM:
21,504 loader writes and 2,251 immutable generated graphics bytes match the
independent Go extractors; title and gameplay render, and loading/gameplay
replays agree. Native/WASM match gameplay RAM/framebuffer and 1,357 guest OSC3
reads; native UBSan passes. Synthetic display/bus/SID cases and all prior CPU
vectors pass. The ledger records exact inputs, digests and unvalidated VIC/SID
corner cases; SID audio and complete analog behavior are not claimed. Local
comparison fixtures can be regenerated with `owned/tests/fixtures`, and no
media or game-derived captures were added to the repository.

C6 is in progress (see the checkpoint foundation below); C7 remains pending.
The production C64 core remains unchanged.


### C4 completion — 2026-10-01

Implemented the independent drive CPU/RAM/ROM map, two 6522 VIAs, CPU SO,
open-collector IEC including ATN acknowledgement, rational dual-machine clocks,
head/motor/LED state, GCR sync/byte-ready and mutable encoded media. D64 synthesis
and G64 v0 raw half-tracks/per-byte speed metadata share the same drive path.
G64 support was brought forward in response to the supplied Giana Sisters disk.

Acceptance uses real C64 and 1541 firmware with physical keyboard commands:
directory loading, a verified authored payload across six tracks/all four zones,
SAVE/reload, write protection, protected-to-protected disk swaps, and in-flight
read/write snapshot replay. The supplied Giana G64 directory and file `F` are
separate normal-DOS checks with an independent checksum-validating GCR oracle.
At C4, custom-loader validation remained C5 work; the current evidence and
remaining game/protection limits are recorded below.

Native, WASM and native UBSan checks cover these gates and the earlier CPU,
board, video and authentic Fort acceptance. The compatibility ledger records
media identities and limits. C4's bit-cell model is not an analog flux/PLL model;
weak bits, changed-density formatting and exhaustive NMOS VIA quirks remain
unvalidated. The production browser core and shipped artifacts are unchanged.


### C5 completion — 2026-10-01

Added processor-qualified breakpoints, safe live disassembly/memory inspection,
VIA/head/bit state, bounded IEC event history and actual drive CPU bus records.
Global pause and stepping operate on an individually ordered dual-clock timeline,
including consecutive drive edges. Restore preserves in-flight port samples and
clears obsolete debugger trace/bypass state. Browser presentation remains C6.

The pinned Giana G64 is the first real custom disk-loader gate: boot `LADER`,
verify 512 bytes uploaded through real DOS `M-W`, stop/step drive code, replay a
byte transfer, compare all 20,168 protocol bytes and all 20,086 resulting payload
stores, then reach the authentic unpacker. The independent GCR oracle supplies
the payload identity. The test exposed and fixed late IEC input sampling by
preserving each CIA/VIA sample from its PHI2 rising read phase.

Native, WASM and native UBSan agree on debugger and custom-loader checks;
ordinary C4 disk gates and earlier CPU/board/video/Fort regressions pass. The
named corpus and precise limits are in the compatibility ledger; the new
`owned/DRIVE-INSPECTION.md` provides a repeatable code/trace investigation.
A longer exploratory run reaches the release's animated intro, but full Giana
gameplay, other custom loaders, weak-bit protection and analog timing are not
certified. No production backend switch or shipped WASM changes were made.


### C6 progress — portable checkpoint foundation — 2026-10-01

Implemented versioned field-wise hardware serialization for a standalone C64
or the complete C64/1541/IEC system. Checkpoints retain in-flight instructions,
device pipelines, video/fetch state, SID noise, tape position, mutable raw disk
tracks/speed zones, and sampled IEC inputs. Host-supplied core/firmware/media
identities and configuration must match; bounded parsing and validation occur
before an atomic restore. See `owned/PORTABLE-STATE.md` for the exact contract.

Acceptance includes native-produced checkpoint bytes imported into WASM,
thirteen clock-phase replay cases, dirty-media and G64 speed-map restoration,
malformed/mismatched input rejection, plus portable Fort and Giana replay.
The separate Elite gate boots its real tape through KERNAL and checks the
initial 52-byte vector block against raw pulses, then repeats from a checkpoint.
It exposed an I/O-selection bug during VIC BA stalls: the PLA deselects I/O
reads during the warning phase, while writes still drain. Fixing that rule
prevents an early CIA ICR acknowledgement from corrupting Elite's byte `$032F`.

C6 is **not complete**. Remaining integration gates:

- Owned browser ABI and explicit development backend selection.
- Code/Memory/Rendering/Storage parity, including exact pixel provenance.
- Browser recording, debugger partial-instruction behavior, experiments and
  cancellation with the owned core.
- Regenerated prepared lesson identities/checkpoints and browser acceptance.
- Continue recording Elite's later loader progress separately from Fort;
  the first-block check does not certify the full self-modifying loader or game.

The shipped backend, WASM artifacts and existing knowledge-package recipes
remain unchanged. C7's default switch is still gated on full C6 acceptance.
