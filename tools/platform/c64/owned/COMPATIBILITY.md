# Owned C64 compatibility and evidence ledger

Validated 2026-10-01 with Apple clang, Emscripten 4.0.16 and Node 26.3.1.
All checks below pass natively and in WASM; native UBSan also passes. This is
bounded evidence, not a claim of exhaustive NMOS or C64 hardware conformance.
Reproduction commands are in [README.md](README.md).

## C2: CPU

| Area | Implemented and checked | Limits |
| --- | --- | --- |
| Documented instructions | All 151 encodings; addressing, stack, page/branch cycles, decimal ADC/SBC, indirect JMP wrap, old/new RMW writes | Independent vectors are sampled |
| Undocumented instructions | SLO/RLA/SRE/RRA/DCP/ISC, SAX, memory LAX, LAS, ANC, ALR, ARR (including decimal), AXS, SBC EB and NOP forms; 237 supported encodings total | Unstable variants below deliberately stop |
| Bus behavior | 15,168 independent vectors; 66,181 cycles; digest `1d597cc1` | Not all upstream cases or analog behavior |
| Sustained execution | Klaus Dormann functional test reaches success at $3469 after 96,241,364 cycles / 30,646,176 instructions | Does not prove device or interrupt timing |
| IRQ/NMI | Prior-phi2 sampling, CLI/SEI/PLP old-I behavior, RTI restored-I, branch polling windows, page-cross poll, NMI/BRK vector hijacking at cycle edges | Authored edge tests based on independent timing descriptions, not exhaustive transistor comparison |
| Reset/RDY/state | Seven reset reads, mid-write reset, RDY read stalls without suppressing writes, NMI during stall, partial-RMW replay | VIC arbitration itself is C3 |

The seven silicon-dependent encodings are intentionally `Unknown` and enter an
explicit diagnostic `Fault`: XAA ($8B), immediate LAX ($AB), AHX ($93/$9F), TAS
($9B), SHY ($9C), SHX ($9E). XAA/immediate LAX involve chip-dependent internal
bus terms; the high-byte-masked stores have address/data coupling on crossings,
and TAS also changes SP. There is no universal guessed constant or silently
substituted NOP. Their actual behavior must be characterized against a chosen
chip profile and targeted evidence before enabling them.

The twelve JAM encodings enter a distinct `Jam` state, recoverable by reset.
The host stops bounded execution and device progression. This does **not** model
the transistor-level repeating JAM bus sequence or a running board with a jammed
CPU. All 256 encodings are accounted for by supported/JAM/unsupported metadata.

CPU timing references:
[CPU interrupt timing](https://www.nesdev.org/wiki/CPU_interrupts),
[Visual6502 interrupt timing](https://www.nesdev.org/wiki/Visual6502wiki/6502_Timing_of_Interrupt_Handling),
[interrupt hijacking](https://www.nesdev.org/wiki/Visual6502wiki/6502_Interrupt_Hijacking).
These informed independently authored control-line tests and implementation.

The optional sustained test is `bin_files/6502_functional_test.bin` from
[Klaus Dormann's tests](https://github.com/Klaus2m5/6502_65C02_functional_tests/tree/7954e2dbb49c469ea286070bf46cdd71aeb29e4b),
revision `7954e2dbb49c469ea286070bf46cdd71aeb29e4b`, SHA256
`fa12bfc761e6f9057e4cc01a665a7b800ff01ae91f598af1e39a1201d01953fd`.
The harness enforces this hash, starts at $0400 and caps execution at 200 million
cycles. It does not redistribute this binary.

## C2: CIA and board

| Area | Evidence | Limits |
| --- | --- | --- |
| Timer pipeline | Published CIA1TB123/CIA2TB123 control-write sequences and CIA1TAB cascaded timer/PB/ICR trace | Not the complete Lorenz suite; no claim covering all chip revisions |
| CIA registers | ICR acknowledgement/cancellation/mask behavior, FLAG edges, one-shot, latch writes, serial receive, TOD latch/rollover/alarm | Serial output is preliminary; exhaustive serial, CNT-mode and revision-specific ICR quirks remain unvalidated |
| Memory | All eight no-cartridge banking combinations, writes under ROM, color nibble, non-mutating peeks | Cartridge lines and expansion devices not implemented |
| Input | Forward/reverse keyboard scans, joystick port mapping; switch propagation models ghosting | No browser input adapter yet |
| Tape | TAP v0/v1 parser, rejection without replacing current media, play/motor/sense, pulse boundaries, consecutive one-cycle edges, end-of-media and replay | No real-game tape-loader acceptance yet; recording and tape analog behavior absent |
| Real firmware | READY banner, 38911 free BASIC bytes, physical-keyboard PRINT 2+2 producing 4, second READY | Screen RAM checked; no VIC-rendered pixels yet |
| State | Partial CPU state and deterministic CIA/tape replay with same media | No portable serialized full-machine checkpoints |

CIA reference: Wolfgang Lorenz,
[A Software Model of the CIA6526, EDK 2.15](https://ist.uwaterloo.ca/~schepers/MJK/cia6526.html).
Keyboard reference: [How the keyboard works](https://www.c64os.com/post/howthekeyboardworks).
CIA2 IRQ and RESTORE are wired to NMI; CIA1 and VIC scaffold IRQ to IRQ.
At C2 the board had no IEC devices or 1541. C4 adds the drive and a separate VIA implementation.

Real boot runs 3,000,000 cycles, types through matrix switches (100,000 cycles
each pressed/released), then runs another 300,000: 5,300,000 total, final PC
$E5D4. No ROM-call traps, keyboard-buffer injection or prepared game RAM. No
unsupported SID reads occur in this boot. Firmware SHA256 identities:

| ROM | Bytes | SHA256 |
| --- | ---: | --- |
| BASIC | 8192 | `89878cea0a268734696de11c4bae593eaaa506465d2029d619c0e0cbccdfa62d` |
| KERNAL | 8192 | `83c60d47047d7beab8e5b7bf6f67f80daa088b7a6a27de0d7e016f6484042721` |
| Character | 4096 | `fd0d53b8480e86163ac98998976c72cc58d5dd8eb824ed7b829774e74213b420` |

The boot harness accepts correctly sized locally supplied ROMs and reports their
hashes; the evidence above applies to these exact images.

## C3: VIC-II and SID-visible behavior

The independently authored `vic.cpp` and `sid.cpp` replace the C2 scaffolds.
CPU conformance results above remain unchanged. Native, native UBSan and WASM
run the same synthetic device assertions and authentic Fort acceptance.

| Area | Acceptance evidence | Limits |
| --- | --- | --- |
| PAL VIC schedule | 63 clocks/line, 312 lines; c/g/p/s, refresh and idle accesses; 40 c-accesses on badlines; BA three-clock warning and AEC isolation; pending CPU write drains under BA | Not a complete transistor/revision conformance suite; dynamic badline cancellation/late activation, VSP/FLI, sprite crunch and NTSC remain unvalidated |
| Display | Standard/multicolor text and bitmap, ECM text, scrolling, independent border flip-flops, banked character ROM, color RAM latches | Illegal mode output approximated; no analog PAL signal or lightpen implementation |
| Raster effects | Compare IRQ and acknowledgement, delayed line-zero IRQ, D018 changes between glyph fetches, writes after a fetch preserve earlier pixels | Subcycle register-write effects and unstable chip-specific effects remain unvalidated |
| Sprites | Pointer and data DMA slots, X/Y expansion, multicolor, foreground/lower-sprite priority, collision latches and IRQs | Extreme right-edge start timing and sprite-crunch tricks remain unvalidated |
| Source capture | Each actual fetch exposes cycle/phase/address/value/kind/slot/ROM identity; matrix, graphics and sprite latches retain fetched bytes | Current-cycle/line records, not yet browser per-pixel history; C6 adds that integration |
| SID readback | Free-running 24-bit oscillators, saw/triangle/pulse, ring/sync, TEST phase reset, noise reference prefix, ADSR attack/sustain/release and rate-change delay | Digital approximation; combined waveforms use AND, without analog coupling/noise feedback. `combinedWaveformUsed` records such writes |
| SID replay | All oscillator, noise, envelope, divider and bus phases are value state; synthetic readback digest `3546492723`; real Fort read stream below | Not matched to a physical chip's initial analog state or exhaustive hardware vectors |

The SID profile starts noise at $7FFFF8, clocks the 23-bit polynomial from
accumulator bit 19 with a two-clock shift delay, and models TEST discharge after
$8000 clocks. Clearing TEST shifts using the inverted bit-17 feedback. The
fixed post-shift noise prefix FE/FC/FC/FC/F8 agrees with Alstrup's reported
sequence; the onset timing is a chosen profile, **not** hardware-validated.
TEST discharge and power-on state vary physically. Bus retention is simplified
to $2000 clocks, disconnected pots read $FF, and combined-waveform behavior is
not a complete SID model. Envelope stepping uses published periods and an
exponential divider; fine gate-write pipelines and LFSR rate-counter phase on
mid-period rate changes remain unvalidated. Audio/filter output is deferred.

Reference material (hardware descriptions and measurements, not copied emulator
implementations):

- [Christian Bauer's VIC-II description](https://www.cebix.net/VIC-Article.txt),
  2024-09-29 edition, for bus/display scheduling and border/sprite rules.
- [MOS 6581 datasheet reproduction](https://www.waitingforfriday.com/?p=661),
  for register-visible oscillator/envelope behavior.
- [Graham's SID reference](https://www.oxyron.de/html/registers_sid.html),
  for the digital noise polynomial and output taps.
- [Asger Alstrup's noise measurements](https://codebase64.net/doku.php?id=base%3Anoise_waveform),
  for the fixed noise prefix; initial TEST timing is explicitly not a conformance
  claim. Earlier write-ups use a shifted register convention for output taps.
- [Thorsten Klose's SID experiments](https://forum.midibox.org/t/mb-sid-v2-discussion/6341?page=3),
  for the free-running envelope rate counter/delay behavior.

### Authentic Fort Apocalypse acceptance

Reference: U.S. Gold / SYNSOFT, NOVALOAD D100701, PAL. TAP SHA256:
`9e444c4576bac52ba691f0ffe2c0a7efb0f62f3fe2be7cbe78dba08672dda00b`.
Firmware identities are the same as the C2 table above.

The test starts from power-on, types LOAD/RUN, advances real tape pulses and
presses joystick fire. It uses no loading traps, supplied game RAM or game-code
patches. Comparisons use independently decoded Go extractor output:

| Fixture | SHA256 |
| --- | --- |
| expected.bin | `598fa977a379f114a4b03fb8cbe485e9ce85fbd09ba5e812439e5902f00d9039` |
| pages.bin | `7b32cdcc1b0b4bb6e0fd4c6b5cd09490fe608a12809f120c66604a11b23d8f32` |
| graphics.bin | `a2baf6f329350171a93a287ce1ef7eb323548e6582c8585cac5831013ff3abab` |
| graphics-mask.bin | `0983c70662f816657829eb2f3efe44f3f074f5dfa672fc444d4ddf051d7a9547` |

Acceptance checks 21,504 live Novaload writes, all main-program bytes before
entry, 2,251 immutable generated charset/sprite bytes, title mode 1 and gameplay
mode 2. Animated/scanner/noise-modified glyphs are intentionally excluded from
the immutable-byte comparison; actual OSC3 reads are checked separately.

Native loading/title/gameplay PPM captures were inspected. The accepted run:

| Observation | Machine cycle | PC | Tape pulse |
| --- | ---: | --- | ---: |
| KERNAL loaded | 34,220,000 | $FCDB | 48,236 |
| Loading screen | 46,400,677 | $038C | 73,083 |
| Game entry | 115,317,157 | $8600 | 225,787 |
| Title | 118,317,157 | $8AA1 | 225,787 |
| Gameplay | 123,997,157 | $ADF5 | 225,787 |

Loading replay covers 50,000 clocks. Gameplay replay covers 2,000,000 clocks,
with identical RAM and framebuffer hashes plus 1,357 completed guest OSC3 reads.
Native/WASM acceptance digests (unsigned 32-bit FNV-1a): RAM `1676350321`, raw
palette-index framebuffer `3657676438`, guest OSC3 stream `1563375142`.
These are reproducibility digests for this core/profile, not physical-hardware
or third-party-emulator golden frames. Image palette/cropping is presentation.

## Next acceptance gates

C4 provides the independent 1541 and IEC bus. C5 validates fastloaders and drive
inspection. C6 integrates the browser, full pixel provenance and portable states;
C7 governs the default switch. The production third-party core remains active.


## C4 — independent 1541 and ordinary DOS media

Two independently clocked CPUs communicate over resolved IEC lines. Drive RAM,
ROM, both 6522 VIAs, motor/stepper, recovered GCR bits, sync and CA1/SO byte-ready
signals are modeled. No guest file-loading calls are intercepted.

Acceptance includes an authored 35-track disk with six linked sectors on tracks
1/17/19/25/31/35 (all four zones); every one of its 1,522 payload bytes is checked.
Real ROM directory loading, BASIC SAVE/export/reload, protected SAVE and a
protected-to-protected disk swap pass. Snapshots during an IEC load and during
an actual disk write reproduce 20,000 clocks of CPU PCs, head bits and IEC events,
plus final RAM, raw tracks and dirty flags. Media-free checks cover timers,
handshakes, latches, IRQ acknowledgement, external shift, SO edges, exact clock
ratio, wired-AND/ATN behavior, D64 35/40-track roundtrips and G64 half-track/speed
maps. Invalid mounts leave the old media intact.

The optional firmware is the 16 KiB concatenation of the Commodore 1541
`325302-01` C000 and `901229-05` E000 halves. SHA256:
`d1d45afb46fd4e2b48d93ca367b889d75654a3b7acf73044e51f6d880c09369e`.
Local-only source archive: [Zimmers 1541 firmware](https://www.zimmers.net/anonftp/pub/cbm/firmware/drives/new/1541/).
No firmware is redistributed in this milestone.

The user-supplied `Great_Giana_Sisters_The.g64` is 333,744 bytes, 84 half-track
slots with 42 populated whole tracks. SHA256:
`5ce29ce04786eca6518fb08dfe659abb3eee079b4135a3f7606f9d17a501ec77`.
Its real-ROM directory reads `GIANA-GAME`; `LOAD"F",8,1` loads 992 bytes at
$CC00 with FNV-1a `735537004`. The independent Python GCR/checksum/chain oracle
reports PRG SHA256 `f6e28e64d68e9bd1c4dcc9e853adafa7782f93888da64d7b97f88aa597f2e6b4`
(including the two-byte load address). This does **not** certify full game startup.
The image is optional local test input, not redistributed here.

C4 limitations: read timing follows recorded GCR speed zones, without analog PLL
acquisition or weak-bit randomness; missing half-tracks read empty. Angular
position across head moves is approximate. Normal same-density writes are tested;
arbitrary formatting/density changes are not. No analog head overlap, motor
spin-up or exhaustive NMOS VIA shift-register corner-case validation is claimed.
The optical insertion/removal sensor sequence is deterministic (200 ms), not a
model of the user's physical insertion speed. C5 owns custom-loader validation.

Hardware/format references used for the independently authored implementation:

- [6522 register/timer/handshake documentation](https://www.westerndesigncenter.com/wdc/documentation/w65c22.pdf) (modern W65C22; NMOS edge quirks require separate validation).
- [Commodore IEC circuit description](https://www.commodore.ca/manuals/funet/cbm/schematics/drives/new/1541/service/Page_09.html).
- [1541 sync and byte-ready hardware investigation](https://luigidifraia.wordpress.com/2020/12/09/how-the-block-sync-and-byte-sync-signals-of-a-commodore-1541-drive-work/).
- [G64 format specification](https://vice-emu.sourceforge.io/vice_17.html) (format documentation only; no VICE device implementation copied).
- [D64 sector/BAM/directory format](https://www.theflatnet.de/pub/cbm/65xx/text/d64.html).
