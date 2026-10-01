# C2 compatibility and evidence ledger

Validated 2026-10-01 with Apple clang, Emscripten 4.0.16 and Node 26.3.1.
All checks below pass natively and in WASM; native UBSan also passes. This is
bounded evidence, not a claim of exhaustive NMOS or C64 hardware conformance.
Reproduction commands are in [README.md](README.md).

## CPU

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

## CIA and board

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
The board has no IEC devices or 1541 yet. VIA will be a separate implementation.

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

## Next acceptance gates

C3 must replace the raster/SID scaffolds with actual VIC-II scheduling, pixels,
DMA/arbitration, sprite/collision behavior and SID-visible oscillator/noise and
envelope evolution. Real Fort Apocalypse loading/gameplay remains unproven on
this core. Later milestones add 1541/IEC, debugger/browser APIs, portable states,
and prepared lesson migration. The shipped third-party core is still active.
