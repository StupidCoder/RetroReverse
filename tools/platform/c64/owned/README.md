# Owned C64/1541 hardware core

This directory starts RetroReverse's independently authored C++20 C64/1541
replacement. It is not yet a playable C64 and is not selected by the production
browser app. See [the implementation plan](../../../../C64-1541-IMPLEMENTATION-PLAN.md).
The existing `browser/core` and its vendored dependencies remain intact until
machine-level acceptance is complete.

## C1: documented NMOS CPU foundation

`cpu.h`/`cpu.cpp` implement one pending bus transaction per CPU cycle. A host
services `cpu.bus()` and supplies its read value to `cpu.tick()`. Writes use the
value already on the bus. Opcode fetches have `sync=true`; RDY stalls reads,
not writes. The board will handle the 6510 processor port and VIC bus arbitration.
The drive will use another instance with its own clock and memory map.

Implemented: all 151 documented opcodes, indexed and indirect addressing,
page-cross and branch cycles, zero-page wrapping, NMOS indirect JMP wrapping,
stack traffic, old/new RMW writes, decimal arithmetic, BRK/RTI, seven-cycle reset,
basic level IRQ/edge NMI and explicit unsupported-opcode stop. `CpuState` holds
registers, pending bus transaction, execution phase, temporary operands and
interrupt latches without pointers. Copying it preserves partial instructions;
this is not yet a versioned portable on-disk machine-state format.

`start(pc)` is a test/debug entry point, not a simulated reset. Boards use
`reset()` and tick the reset bus sequence. `boundary()` denotes a pending opcode
fetch, before that fetch is consumed. Hosts must distinguish `faulted()` from an
ordinary instruction boundary and keep all execution requests bounded.

The opcode table is generated from RetroReverse's existing instruction metadata:

```
python3 tools/platform/c64/owned/generate-opcodes.py
```

No vendored CPU or device code is compiled into this implementation. The
independent test corpus is third-party MIT-licensed data with attribution below.

## Reproduce acceptance offline

```
python3 tools/platform/c64/owned/check.py --native-only
python3 tools/platform/c64/owned/check.py --emcc /path/to/emscripten/em++
```

The default uses temporary build outputs. `--out /absolute/scratch/path` retains
them. `CXX`, `EMXX` and `NODE` select toolchains. Native-only acceptance is also
called by `tools/browser/check.py`. WASM acceptance runs the same C++ harness
under Node, including the independent vectors embedded in its test filesystem.
No game media, ROMs or browser deployment are required for C1.

The harness verifies every read/write address, value and cycle count, final
registers, and final memory from 64 cases per documented opcode. Fixtures are a
bounded sample, **not** the complete upstream suite. `vectors.json` records the
source revision, selection rule and content hash. Run `tests/fetch-vectors.py`
explicitly to reproduce downloads; ordinary checks never use the network.
The source is [SingleStepTests/65x02](https://github.com/SingleStepTests/65x02),
revision `2f6980a2d95757486c7bee24355c360e40e2a224`, NMOS `6502/v1`.
The redistributed data's license is [SingleStepTests-LICENSE](tests/SingleStepTests-LICENSE).
`vectors.bin` is a compact lossless encoding of the selected states and cycles:
LE `CPV1` magic and u32 count; each case has opcode u8, source index u16, initial
and final `(PC:u16,S,A,X,Y,P:u8)`, initial/final RAM lists `(count:u16,(address:u16,
value:u8)[])`, and `(cycleCount:u8,(address:u16,value:u8,write:u8)[])`.

Additional authored tests cover multiple instructions, repeated RDY reads,
un-stalled writes, reset stack reads, IRQ stack contents and CLI delay, an NMI
edge during a read stall, unsupported-opcode stopping, and replay across the two
writes of an in-progress RMW instruction.

Validated on 2026-10-01 with Apple clang, Emscripten 4.0.16 and Node 26.3.1:

- 9,664 independent vectors, 38,749 observed bus cycles.
- Native and WASM bus/register trace digest: `6a229904`.
- All authored control-line and replay cases pass in both builds.
- Native undefined-behavior sanitizer passes with the same digest.
- AddressSanitizer was attempted but stalled during macOS sanitizer runtime
  initialization, before main; it is not claimed as passing.

## Remaining CPU acceptance before board integration

C2 must complete undocumented opcodes, characterize unstable variants, and test
interrupt polling at precise cycle edges (including branch timing, NMI/BRK
hijacking and reset/RDY interactions). Current IRQ/NMI logic provides the basic
state contract and tested cases; it is not an exhaustive NMOS timing claim.
The complete upstream vectors, sustained functional ROM tests and independent
hardware timing tests remain broader gates. C1 passes alone do not demonstrate
KERNAL boot, a working VIC/CIA/SID, tape compatibility or 1541 behavior.

Nothing here changes the existing Go analysis CPU, shipped WASM, prepared lesson
cache identities, or third-party-core save states. Those are addressed by the
later board, device and integration milestones.
