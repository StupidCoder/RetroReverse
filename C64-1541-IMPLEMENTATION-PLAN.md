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
media corpus. Add G64 when nonstandard track contents are required; D64 does
not preserve arbitrary protection/track layouts. Expose drive CPU disassembly,
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
location are recorded. C1 in progress. The production C64 core remains unchanged.
