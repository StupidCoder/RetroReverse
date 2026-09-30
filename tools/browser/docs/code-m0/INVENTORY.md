# M0 capability and research inventory

Audited 2026-09-30 at commit `2284b8c8` plus existing unrelated local changes.
The audit did not rebuild cores. SHA-256 of every inspected WASM and glue file is
in [core-audit.json](../../results/code-m0/core-audit.json). Named exports may be
minified internally; the audit checks Emscripten public bindings as well. Export
presence proves linkage only. It does not prove semantics or game compatibility.

## Browser versus native versus prototype

All 16 shipped core bindings expose status, physical inspection, state save/load
and activity entry points. Only C64, GG and Amiga currently enable general Memory
activity recording (`supported=true`). The other activity entry points are not
support claims. Render provenance is separate, selective instrumentation.

The unified worker exposes Play/Pause/frame step, capture/replay and Memory, but
no instruction-debug protocol. `tools/debug/debug.go` describes native optional
CodeStepper capabilities; native support is not browser support. The standalone
C64 lesson directory and local educational plan contain useful but partly
untracked prototype UI/history. M2 must port verified behavior deliberately.

| Browser ID / CPU family | Existing execution seam | Instruction/debug qualification | Memory activity |
|---|---|---|---|
| c64 / 6510 | `core.cpp:rr_run(ticks,stop_kind,target)` | Opcode/PC/write stop; PC latch mismatch reproduced; boundary/RDY/IRQ audit required | CPU accesses + tape |
| gb / SM83 | shared handheld API, `rrAdvance` calls machine Step | Internal step, no unified instruction control | none |
| gg / Z80 | shared handheld API + platform advance | Internal step, no unified instruction control | CPU/video-port access |
| gba / ARM7TDMI | `api.cpp:rr_run`, RunFrames budget | Frame/budget API; ARM/Thumb context required | none |
| amiga / 68000 | `api.cpp:rr_run`, machine run | Musashi/device scheduling, no unified instruction control | CPU/DMA/disk |
| ps1 / R3000A | `rr_run(n,stopField)` | Instruction budget; delay slots/load delays need explicit semantics | none |
| n64 / VR4300 + RSP | `rr_run(n)`, machine Run | Budget/scheduler; RSP separate execution context | none |
| 3do / ARM60 | `rr_run_slice(n)`, resumable runSlice | Budget/scheduler, HLE tasks; not exact guest instruction API | none |
| ds / ARM9 + ARM7 | `rr_run(n)`, machine scheduler | Must define selected CPU and other CPU progression | none |
| 3ds / ARM11 | `rr_run(n)`, RunFrames | HLE process/scheduler and mapped allocations | none |
| psp / Allegrex | `rr_run(n)`, machine Run | HLE scheduling, delay slots | none |
| ps2 / EE + IOP/VU | `rr_run(n)`, machine Run | Multiple processors and VU work, not one global PC | none |
| gc / PowerPC + DSP | `rr_run(n)`, machine Run | CPU/device/DSP timing, no unified step | none |
| dc / SH4 + AICA | `rr_run(n)`, machine Run | Delay slots/device work, no unified step | none |
| dos / x86 | `rr_run(n)` calls CPU_Step | Real/protected modes; current status only linear PC | none |
| xbox / x86 | `rr_run(n)`, machine Run | Guest/HLE/device boundary, no unified step | none |

Source anchors: `site/emulators/worker.js:tick/pump/onmessage`;
`tools/platform/<platform>/browser/core/api.cpp` (C64 uses `core.cpp`);
GB/GG use `host.h` and `tools/browser/handheld/api.inc`; physical/recording adapters
are each platform's `core/memory.h` or C64 `memory.inc`. Platform directories use
`gameboy/gamegear/psx/threedo/nds/n3ds` for their shorter browser IDs.

### Decoder inventory

Repository native tooling includes mos6502, z80, sm83, m68k, mips, allegrex, ARM,
ARM60, x86, SH4 and other CPU-specific decoders. Reuse requires parity fixtures
and mode coverage, not simply finding a similarly named CPU. C64 prototype
`web/disassemble.js` is a small JS lookup table; `$55` is missing. The native
mos6502 decoder contains EOR/zpx. x86 `x86.go` has Decode/Decode32 and `modrm.go`
has linear disassembly helpers. Live bytes and known boundaries are mandatory
for UW; static EXE decoding alone misses relocated overlays.

### State and recording inventory

The worker state container pins core/media/firmware/configuration; candidate
restore uses a separate worker. Existing U4 reports native/WASM continuation for
C64, PS1, N64, 3DO; later ports expose save/load but require their own scenario
acceptance. No blanket deterministic-tour claim follows from exported save/load.
C64 low-level checkpoints have 16 slots and generation checks. Its observed
serialized state is 230,067 bytes on the synthetic fixture. Memory's bounded
byte-history cursor does not restore CPU/devices. Render seek has its own capture
identity/replay context. These histories must not be conflated in the new API.

## Research inventory and unresolved claims

### Fort Apocalypse

Source: [write-up](../../../../games/fort-apocalypse-c64/fort-apocalypse-c64.md),
sections Enemy helicopter, Player, Appendix B. Exact tape SHA-256 is already in
`memory-labels.js`: `9e444c4576bac52ba691f0ffe2c0a7efb0f62f3fe2be7cbe78dba08672dda00b`.
Candidate definitions: AI `$9C52`, hunt `$9CDA`, upward probe `$A000`, player
movement `$A4CE`; enemy mode `$6E`, positions `$76/$77`, player `$69/$6A`, bank `$67`.

Unresolved: prose gives hidden text `N0:JOE VIERMON.Z` yet attributes corruption
to a trailing `U` at the entry; establish exact bytes/boundaries rather than
transcribing this contradiction. The stored state and actual upward lookup need
live validation before a lesson predicts a crash. No invented stored velocity.
M0's synthetic PC test is independent of that game hypothesis.

### Elite

Source: [write-up](../../../../games/elite-c64/elite-c64.md), Parts II and V.
Up to ten slots; type list `$0452` terminates at zero; pointer table `$28A1`;
37-byte records; `$0009..$002D` is the per-object scratch workspace. It is not a
simple contiguous array of independently typed records. Negative type codes
identify special objects. `slot_ptr` `$3E84`, update `$1EBE`/loop `$202A`, spawn
`$855B` are candidates. Resolve occupancy, termination and reuse from actual code.

Fields include position with a separate sign byte, speed/acceleration, AI and
removal flags. Human ship names are not strings stored in the game; link labels
to documented blueprints with provenance. Do not infer a name solely from a number.

Loader research specifically describes `$0347` ROL/ROR changing bit order,
`$034B` byte-frame bit count, `$03A4` header size and `$03BB` exit opcode. This is
stronger/more precise than a generic “encoding changes” story. Verify execution,
tape position, overwritten code and write-then-restore intervals. Earlier U2
reports bad Elite rendering; current compatibility was not revalidated in M0.

### Ultima Underworld

Source: [write-up](../../../../games/ultima-underworld-pc/ultima-underworld-pc.md)
and `disasm/uw-render.annotations.txt`. README/introduction still says only
camera setup is mapped, but later sections document spans, projection and copy.
Migrate by reading the full evidence, not trusting the introductory progress text.

Observed runtime addresses: projection `07F7:6148`, edge DDA `01A0:0312`, span
setup `01A0:0296`, texture span `01A0:02CE`, blit `01A0:0B96`; off-screen segment
`41C5`, graphics state `3BDD:07B7..07CF`, camera basis `499D:1600..1618`.
These require relocation/module binding and applicability, not fixed global
segments. Gradient immediates self-modify. Reconcile “1.15 / 0x8000 = 1.0” wording
with signed arithmetic/saturation before choosing a fixed-point display type.

### Need for Speed (3DO)

Source: [write-up](../../../../games/need-for-speed-3do/need-for-speed-3do.md),
Part XI. Seven linked car structs of `0x4e4` bytes in the observed City race;
addresses `0x293f90..0x295ce8` are observations, not a global allocation guarantee.
Driver pointer `+0x4c4`: player `0xf8d8`, road AI `0x284e0`, dormant `0x2a4d8`;
placer `0x28eb0`, behavior mask `+0x188`, activity `+0x4cc`.

The documented module dispatcher `0x27800` skips dormant cars; it cannot by
itself be declared the complete spawn/AI loop boundary. Trace the outer driver
invocation/cycle before defining a tour. Identify active, dormant, opponent and
police via actual fields, not object count. Frame count is not simulation time.
HLE scheduling/input profile and scene checkpoint compatibility remain gates.

## Prioritized risks and owners by milestone

| Risk | Evidence/impact | Resolution gate |
|---|---|---|
| Wrong current instruction | C64 status reports preceding PC in measured fixture | M2 normalized boundary probe + IRQ/RDY tests |
| Reentrant jobs | Worker has epoch plus saving/memoryBusy/capturing flags | M2 single arbitration, terminal response tests |
| Incomplete decode | Prototype omits relevant opcode | M2 generated table/parity and unknown-opcode policy |
| False recording capability | All cores export activity; only three enable it | M1/M2 explicit capabilities, never export-name inference |
| Stale annotations | Fort contradiction, UW summaries lag findings | M1 evidence migration, M4/M5 actual game validation |
| Misleading object table | Elite indirection/scratch; NFS linked allocations | M3 typed pointer/occupancy fixtures |
| Nondeterministic comparison | Restore/input policies differ by context | M8 identical replay before patch experiments |
| Memory blowup | Full snapshots, trace, media, checkpoints overlap | M2/M4 budget accounting and cap tests |
| Compatibility hidden by labels | Elite historical rendering failure | M3/M4 separate current scene acceptance |
| Overpromised measurements | Synthetic baseline is not a game or UI | M2 real Code timings; per-game milestones run media fixtures |
