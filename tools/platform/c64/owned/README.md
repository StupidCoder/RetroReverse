# Owned C64/1541 hardware core

RetroReverse's independently authored C++20 replacement now has a tested NMOS
CPU, C64 board, PAL video, SID register-visible behavior and an independent
1541 with ROM-driven IEC disk loading and saving. The reference Fort
Apocalypse TAP boots through KERNAL and Novaload into rendered gameplay. This
core is still isolated from the production browser app. See
[the implementation plan](../../../../C64-1541-IMPLEMENTATION-PLAN.md) and
[the compatibility ledger](COMPATIBILITY.md). The existing `browser/core` and
vendored dependencies remain intact until machine-level acceptance is complete.

## CPU and board contracts

`cpu.h`/`cpu.cpp` expose one pending bus transaction per CPU cycle. A host
services `cpu.bus()` and supplies its read value to `cpu.tick()`. Writes use the
value already on the bus. Opcode fetches have `sync=true`; RDY stalls reads,
not writes. The drive uses another instance with its own clock and map. The optional
`so` argument is the asserted active-low SO pin; its edge sets overflow even
while RDY holds a read.

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
identities. Browser integration and portable snapshots remain later milestones.

## 1541 and media contracts

`System` connects a PAL `Board` at 985,248 Hz to a `Drive` at 1 MHz. Its integer
scheduler orders every clock edge; a drive edge at the same time precedes the
C64 edge. Both CPUs execute their own real firmware. IEC ATN, CLK and DATA
resolve as open-collector lines, including the drive's hardware ATN acknowledge.
`IecState` reports participants, transition count and the latest timestamped
transition. A future inspector must consume these events while running.

The drive has 2 KiB RAM, a 16 KiB firmware array, two independent `Via` devices,
a stepper position, motor/LED, GCR sync and byte-ready/SO signals. `read` has VIA
side effects; `peek` does not. Firmware is supplied by the caller. There are no
KERNAL or drive-ROM file-loading traps.

`Drive::mount(bytes, writeProtected=true)` validates before replacing media;
`eject()` removes it. Media changes generate a bounded optical-sensor sequence,
including protected-to-protected swaps. Ordinary 35/40-track D64 images are
encoded into headers, GCR data/checksums, sync and gaps. Error-map D64 variants
are rejected. G64 v0 stores up to 84 raw half-track slots and per-byte speed zones;
missing half-tracks stay empty. Source padding is not meaningful track content.

Media is write-protected by default. Guest writes alter only value-owned encoded
tracks and dirty flags; they never modify the host's input file. `exportD64`
requires a D64-origin image with complete valid normal sectors. `exportG64`
retains raw tracks and speed maps; both return bytes for an explicit caller-led
export. No automatic persistence or conversion of protected layouts occurs.

`System::save/restore` captures both CPUs, VIAs, C64 devices, clock phases, IEC,
head/bit position, media-change phase and all mutable media. This resumes
partial transfers and writes with the same external firmware and TAP data.
The C++ value snapshot is not the portable browser checkpoint format.

The C4 model works at recovered GCR bit-cell level. It uses recorded density
for read timing and selected VIA density for writing. It does not model analog
PLL acquisition, weak bits, head width, motor spin-up, or arbitrary changed-density
formatting. Half-track moves approximate angular continuity by scaling bit
position. VIA shift behavior has bounded authored tests, not an exhaustive NMOS
silicon suite. These limits matter for protection and fastloaders; C5 is the
separate acceptance gate for named custom loaders and drive inspection.

## Reproduce disk acceptance

Supply the concatenation of local 1541 ROM halves `325302-01` (C000) and
`901229-05` (E000), in that order. The harness pins the combined identity; it
never fetches firmware or game images:

```
python3 tools/platform/c64/owned/check.py \
  --emcc /path/to/emscripten/em++ \
  --out /private/tmp/rr-owned-disk-acceptance \
  --firmware-dir site/emulators/firmware/c64 \
  --drive-rom /absolute/path/1541.rom \
  --giana-g64 games/great-giana-sisters-c64/Great_Giana_Sisters_The.g64
```

`--giana-g64` is optional. Authored offline tests always cover VIA behavior,
SO, media parsing/roundtrips/protection, IEC, clock ratios and snapshots. With
firmware, physical keyboard commands load the authored directory and a
1,522-byte file spanning tracks 1, 17, 19, 25, 31 and 35, save and reload a BASIC
program, replay live IEC reads/writes and attempt a write-protected SAVE. With
the pinned G64, a protected-to-protected media swap must yield `GIANA-GAME`,
and `LOAD"F",8,1` must reproduce the independently decoded 992-byte payload.
This is normal DOS loading acceptance, not a claim that Giana gameplay boots.

Reproduce the independent G64 payload identity without emulating either CPU:

```
python3 tools/platform/c64/owned/tests/g64fixture.py \
  games/great-giana-sisters-c64/Great_Giana_Sisters_The.g64
```

That oracle validates GCR header/data checksums, follows DOS chains and prints
lengths/hashes only. It is never used to inject bytes into a running machine.
