# Owned C64/1541 hardware core

RetroReverse's independently authored C++20 replacement now has a tested NMOS
CPU, C64 board, PAL video and SID register-visible behavior. The reference Fort
Apocalypse TAP boots through KERNAL and Novaload into rendered gameplay. This
core is still isolated from the production browser app and has no 1541 yet. See
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

`Vic` performs scheduled phi1/phi2 memory fetches, controls BA/RDY and AEC,
and emits eight palette-index pixels per clock. The framebuffer is a raw
504×312 PAL raster in VIC X coordinates; presentation should crop blanking.
`Vic::fetches[0..fetchCount)` records this clock's addresses, sampled bytes,
color nibbles, ROM/RAM identity, phase, kind, slot and emulated cycle. Matrix,
graphics and sprite latches retain captured values. Observers must consume
fetch records as execution advances; later RAM reads cannot reconstruct them.
`BoardState::lastBus` separately exposes the actual CPU transaction (including
held reads and AEC disconnection). Full browser pixel provenance is C6 work.

`Sid` clocks all three oscillators/envelopes, exposing voice 3 through OSC3 and
ENV3. Noise uses its own chip-clocked shift register, with no host randomness.
Audio synthesis remains deferred. Combined waveforms, power-on state, TEST
noise discharge and bus decay are explicit approximations; see the ledger
before treating this as a transistor-accurate SID reference.

Devices keep running during VIC CPU stalls. CIA/tape timing precedes the CPU
transaction; the SID also advances on every machine clock. AEC prevents CPU
memory accesses, while BA allows pending writes to finish during its warning.

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

## Authentic Fort acceptance

Generate local comparison data using the existing Go tape/graphics extractors:

```
GOCACHE=/private/tmp/retroreverse-go-cache go run ./tools/platform/c64/owned/tests/fixtures \
  -image games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  -out /private/tmp/rr-owned-fixtures
python3 tools/platform/c64/owned/check.py \
  --emcc /path/to/emscripten/em++ \
  --out /private/tmp/rr-owned-acceptance \
  --firmware-dir site/emulators/firmware/c64 \
  --fort-tap games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  --fort-fixtures /private/tmp/rr-owned-fixtures
```

Only the documented reference release is covered. The generator checks its
SHA256 and all KERNAL/fastloader checksums. Expected bytes are comparison data;
the harness never injects them, patches the game or traps ROM calls. It enters
LOAD and RUN through keyboard switches, starts the tape and presses joystick
fire. It checks every payload store, immutable generated graphics, title/game
modes, loading/gameplay replay, framebuffer and actual guest OSC3 reads.

Native runs write `fort-loading.ppm`, `fort-title.ppm` and `fort-gameplay.ppm`
into the output directory. The 392×272 images use a fixed presentation palette,
not an analog PAL color model. Game-derived data and images remain scratch
artifacts and must not be committed. Native UBSan and WASM repeat the complete
boot; the latter compares the same acceptance output/digests.

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
