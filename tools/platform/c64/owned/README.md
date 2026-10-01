# Owned C64/1541 hardware core

RetroReverse's independently authored C++20 replacement now has a tested NMOS
CPU and a C64 board that boots real KERNAL/BASIC and accepts keyboard input.
It is not yet a playable C64 or selected by the production browser app. See
[the implementation plan](../../../../C64-1541-IMPLEMENTATION-PLAN.md) and
[the compatibility ledger](COMPATIBILITY.md). The existing `browser/core` and
vendored dependencies remain intact until machine-level acceptance is complete.

## CPU and board contracts

`cpu.h`/`cpu.cpp` expose one pending bus transaction per CPU cycle. A host
services `cpu.bus()` and supplies its read value to `cpu.tick()`. Writes use the
value already on the bus. Opcode fetches have `sync=true`; RDY stalls reads,
not writes. The future drive uses another instance with its own clock and map.

`start(pc)` is a debug/test entry point. Machines use `reset()` and tick its
seven-cycle read sequence. `boundary()` denotes a pending opcode fetch before
consumption. `faulted()` distinguishes unsupported instructions and JAM from an
ordinary boundary. Execution requests must remain bounded.

`board.h`/`board.cpp` connect the processor port, banking, RAM/ROM/color RAM,
both CIAs, keyboard switches, joysticks and TAP media. `power()` clears machine
state; `resetCpu()` resets only the CPU. `read()` performs device side effects;
`peek()` is safe for inspection. Keyboard coordinates are physical matrix
columns/rows, not character codes. Firmware arrays must be populated by the
caller. `loadTape()` validates TAP v0/v1 before replacing current media.

`CpuState` and `BoardState` contain value-owned registers, device state, pending
transactions, timing phases and interrupt latches. Copying board state resumes
partial instructions and tape pulses with the **same firmware and media**.
Firmware and decoded pulses live outside that state. This is not a versioned
portable checkpoint format, nor compatible with the previous core's snapshots.

The PAL raster clock/register scaffold lets firmware read raster progress; it
has no fetches, pixels, badlines, sprite DMA or CPU arbitration yet. SID writes
are retained; OSC3/ENV3 reads return zero and set `unimplementedSidRead` so use
is visible to tests. Neither is a substitute for C3's devices. The board ticks
CIAs before the CPU bus access; tape pulse boundaries inject CIA1 FLAG events.

The opcode table uses existing RetroReverse documented metadata and authored
undocumented-instruction metadata in `instruction_set.py`:

```
python3 tools/platform/c64/owned/generate-opcodes.py
```

No vendored CPU or device implementation is compiled into this core.

## Reproduce acceptance

Offline, without game media or firmware:

```
python3 tools/platform/c64/owned/check.py --native-only
python3 tools/platform/c64/owned/check.py --emcc /path/to/emscripten/em++
```

Optional real firmware boot and independent sustained functional test:

```
python3 tools/platform/c64/owned/check.py \
  --emcc /path/to/emscripten/em++ \
  --firmware-dir site/emulators/firmware/c64 \
  --functional-rom /absolute/path/6502_functional_test.bin
```

The ledger records exact firmware and functional-test identities. The harness
never downloads media. Optional files are embedded only in scratch WASM test
outputs; they are not redistributed here. `--out /absolute/scratch/path` retains
builds; otherwise outputs are temporary. `CXX`, `EMXX` and `NODE` select tools.
`tools/browser/check.py` also calls native-only, media-free acceptance.

Native and WASM run the same assertions and compare output, including the CPU
bus/register digest. A native undefined-behavior sanitizer run must also agree.
AddressSanitizer previously stalled during macOS runtime initialization before
main; it is not claimed as passing.

## Independent instruction fixtures

The harness verifies every read/write address, value and cycle count, final
registers, and final memory for 64 cases per supported encoding. This is a
bounded sample, **not** the complete upstream suite. `tests/vectors.json` records
the source revision, selection and hash. `tests/fetch-vectors.py` explicitly
reproduces downloads; ordinary checks never use the network.

Source: [SingleStepTests/65x02](https://github.com/SingleStepTests/65x02), revision
`2f6980a2d95757486c7bee24355c360e40e2a224`, NMOS `6502/v1`.
The redistributed data's license is
[SingleStepTests-LICENSE](tests/SingleStepTests-LICENSE).
`vectors.bin` is a compact lossless encoding: LE `CPV1` magic and u32 count;
each case has opcode u8, source index u16, initial and final
`(PC:u16,S,A,X,Y,P:u8)`, initial/final RAM lists
`(count:u16,(address:u16,value:u8)[])`, and
`(cycleCount:u8,(address:u16,value:u8,write:u8)[])`.

Nothing here changes the existing Go analysis CPU, shipped WASM or lesson cache
identities. Browser integration, portable snapshots and the 1541 remain later
milestones.
