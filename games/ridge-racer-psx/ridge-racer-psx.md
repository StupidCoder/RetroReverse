# Ridge Racer (PlayStation) — technical reference

**Image:** `Ridge Racer (Track 01).bin` — 3,683,232 bytes, MD5 `ec755afa6ca49432445384b870412dfb` (data track only). Not committed (copyright); supply your own copy.

Ridge Racer (Namco, 1993 arcade / 1994 PlayStation) is a texture-mapped, GTE-transformed 3D
racing game running at 30 fps. This document reconstructs the shipped PlayStation disc from its
bytes alone: no third-party emulator, debugger or disassembler, no released source, and nothing
about the file or instruction formats taken from external documentation. Everything is derived
from the image with purpose-built tools.

The PlayStation runs a **MIPS R3000A** (the LSI CW33300 core) with a fixed-point 3D coprocessor,
the **GTE**. Analysing the disc requires a MIPS toolchain: a disassembler, a recursive code-tracer,
an execution core, a bit-exact GTE, and a **machine oracle** that boots the game. The whole game
loads into the console's 2 MiB of RAM — which is why the game CD can be swapped for any music CD
mid-race — so the data track holds the executable and every asset.

The image is the disc's **data track only**; the in-game music lives on separate Redbook
CD-audio tracks (the CD-swap feature) that are not part of this file.

## Contents

* **Part I** — the disc, the filesystem, and the executable: the raw **2352-byte CD sectors**, the
  **ISO 9660** volume and its file catalog, `SYSTEM.CNF`, and the **PS-X EXE** boot executable with
  its load layout and embedded string tables.
* **Part II** — the **MIPS R3000 / PSX toolchain**: a MIPS disassembler, a recursive code-tracer,
  an execution core (with the R3000's branch- and load-delay slots), a bit-exact **GTE**, and a
  **machine oracle** with an HLE BIOS that boots the game.
* **Part III** — the **program architecture**: the runtime memory map, the interrupt system
  (`I_STAT`/`I_MASK`, the synthetic VBlank, the game's own vectored dispatcher), the **CD-ROM
  controller**, the **DMA controller**, and the interrupt-driven main loop with its controller input.
* **Part IV** — the **graphics pipeline**: the GPU's GP0/GP1 command stream, the 1 MiB VRAM, the
  ordering-table DMA, the triangle/quad/rectangle rasterizer, the GTE lighting that shades the
  models, and the texture staging (CPU→VRAM upload and VRAM→CPU read-back) that populates VRAM for
  each scene.
* **Part V** — the **asset index and load map**: the boot load sequence, the staging buffer, where
  each file resides in RAM, and `IDX.HED` — the 32×32 course grid that places every track section
  in the world.
* **Part VI** — the **asset formats**: the `TEX*.TMS` texture-upload streams, the `MAP.RRM` track
  geometry, the `OBJ.RRO` object models with their six polygon record types, how polygons reference
  texture pages and CLUTs, the executable's car table, the roadside placement tables, and the
  dynamic objects (the number girl, helicopter, airplane, start balloons, big screens and friends).
* **Part VII** — **extraction and verification**: the pure-Go decoders in `extract/rr`, the
  `geomoracle` differential that checks them bit-exact against the running game, and the
  `webexport` pipeline that bakes the textured car and course GLBs.
* **Part VIII** — **the driving simulation**: the car state block and its 30 Hz update pipeline,
  the engine/gear model, the polar-coordinate velocity core that produces the drift, the
  road-edge and car-to-car collision with its ×0.7 bump and damped-sine wobble, and the opponent
  AI — a centerline follower with a player-aware speed state machine (the rubber band).

Addresses are MIPS virtual addresses (`0x8000xxxx` in the cached KSEG0 window) or byte offsets into
a file (called out explicitly); bytes are little-endian.

---

# Part I — The disc, the filesystem, and the executable

## 1. What kind of image this is

The file is exactly **1,566 × 2,352 bytes**, and 2,352 is the size of a *raw* CD sector — sync,
header and error-correction included — so this is a raw dump, not a "cooked" 2048-byte ISO. Sector
0 opens with the CD sync mark and header:

```
$0000  00 FF FF FF FF FF FF FF FF FF FF 00   12-byte sync
$000C  00 02 00                              address (min=00, sec=02, frame=00, BCD)
$000F  02                                    mode = 2
$0010  00 00 08 00 ...                       Mode-2 subheader (submode $08 = data)
$0018  ... 2048 bytes of user data ...
```

Mode byte `$02` and a subheader submode with the data bit set make every sector **Mode 2, Form 1**:
2,048 bytes of user data starting at offset `$18`, wrapped in the CD's sync/header/subheader/EDC/ECC
framing. Stripping that framing to expose the 2,048-byte payload of each logical block is the one
job unique to a CD image versus a flat disk; the reader folds it into the block accessor:

```
tools/platform/psx/cd.go  →  func (v *Volume) block(n int) []byte   // img[n*2352+0x18 : +2048]
```

## 2. The ISO 9660 filesystem

On top of that 2,048-byte-per-block payload sits a textbook **ISO 9660** volume. The Primary Volume
Descriptor is at logical block 16 (`\x01CD001\x01`), and it identifies the disc:

```
system identifier : "PLAYSTATION"
volume identifier : "RIDGERACERUSA"      → the NTSC-U release
```

Walking the root directory (both-endian fields, little-endian half taken) gives the complete file
catalog. Counts and sizes are observed facts; the *role* of each file is an inference from its name
and size, and the internal layout of the asset archives is outside the scope of this reference:

| File | Bytes | LBA | Inferred role (name-implied) |
|---|---:|---:|---|
| `SYSTEM.CNF` | 68 | 23 | boot configuration (below) |
| `SCUS-943.00` | 438,272 | 24 | the boot executable — a **PS-X EXE** (§3) |
| `RR.VB` | 491,056 | 238 | VAB **body** — the ADPCM sample bank |
| `RR.VH` | 32,288 | 478 | VAB **header** — region/pitch/ADSR tables |
| `TEX4.TMS` | 14,996 | 494 | texture archive |
| `TEX0.TMS` | 699,908 | 502 | texture archive (the largest asset) |
| `TEX1.TMS` | 197,128 | 844 | texture archive |
| `TEX2.TMS` | 140,704 | 941 | texture archive |
| `TEX3.TMS` | 110,144 | 1010 | texture archive |
| `MAP.RRM` | 271,548 | 1064 | track / map geometry |
| `OBJ.RRO` | 445,348 | 1197 | objects / 3D models |
| `IDX.HED` | 2,048 | 1415 | an index / header table |

`RR.VB`/`RR.VH` are the standard PlayStation **VAB** pair (a `.VH` header describing voices that
index ADPCM samples in the `.VB` body); the `.TMS` files are the game's texture archives; `.RRM`
and `.RRO` are the track and object data. The runtime loads these into RAM and feeds them to the
subsystems in Parts III–IV. Re-list at any time:

```
go run retroreverse.com/tools/platform/psx/cmd/psxinfo -ls "games/ridge-racer-psx/Ridge Racer (Track 01).bin"
```

`SYSTEM.CNF` is the 68-byte boot script the console reads first — plain text:

```
BOOT = cdrom:SCUS-943.00;1
TCB = 4
EVENT = 10
STACK = 801FFF00
```

`BOOT` names the executable to load and run; `TCB`/`EVENT` size the BIOS thread/event tables; `STACK`
sets the initial stack pointer. `SCUS-943.00` is the US product code (the `SCUS-94300` on the disc),
which is why the volume is `RIDGERACERUSA`.

## 3. The PS-X EXE

`SCUS-943.00` begins with the ASCII bytes **`PS-X EXE`**, the PlayStation executable signature. The
format is a fixed 2,048-byte (`$800`) header followed by the raw text image; the loader
(`tools/platform/psx/exe.go`) decodes it:

| Field | Value | Meaning |
|---|---:|---|
| entry PC | `0x80040004` | first instruction executed |
| gp | `0x00000000` | initial global pointer (unused here) |
| text address | `0x80010000` | where the text image is loaded in RAM |
| text size | `436,224` (`$6A800`) | copied verbatim into RAM |
| initial SP | `0x801FFFF0` | near the top of the 2 MiB RAM |
| region marker | `"Sony Computer Entertainment Inc. for North America area"` | NTSC-U |

`$800` (header) + `$6A800` (text) = `$6B000` = 438,272 bytes, exactly the file size — a single
contiguous text section, no separate data or bss. The load address `0x80010000` is in the **KSEG0**
cached window; physically it is RAM offset `0x10000`, leaving the low 64 KiB for the kernel/BIOS
scratch. Verify:

```
go run retroreverse.com/tools/platform/psx/cmd/psxinfo -exe "SCUS-943.00;1" "games/ridge-racer-psx/Ridge Racer (Track 01).bin"
```

## 4. What's inside the executable

The first few kilobytes of the text image are not code but a string table — surfaced immediately by
the code-tracer (Part II §2), which correctly leaves it as data rather than mis-decoding it. It is a
tour of the game's contents: the selectable cars are named after other Namco titles, and the
soundtrack titles are the real Ridge Racer BGM tracks:

```
"F/A RACING"  "RT RYUKYU"  "RT YELLOW SOLVALOU"  "RT BLUE SOLVALOU"
"RT PINK MAPPY"  "RT BLUE MAPPY"  "GALAGA RT PLID'S"  "GALAGA RT CARROT"
"RT BOSCONIAN"  "RT NEBULASRAY"  "RT XEVIOUS RED"  "RT XEVIOUS GREEN"  "13\" RACING"
"TM&© 1993 1994 NAMCO LTD."  "ALL RIGHTS RESERVED"  "PUSH START BUTTON"
"RANDOM PLAY"  "RIDGE RACER"  "RARE HERO"  "FEELING OVER"  "ROTTERDAM NATION..."
```

("Rotterdam Nation", "Feeling Over" and "Rare Hero" are the names of the game's music tracks.) Past
this table the image is MIPS code; Part II builds the tools to read it.

---

# Part II — The MIPS R3000 / PSX toolchain

Everything past Part I needs to read and run MIPS. The toolchain lives at **`tools/cpu/mips`** (the
CPU) and **`tools/platform/psx`** (the disc reader, executable loader and machine), platform-neutral
under the shared `retroreverse.com/tools` module.

## 1. The disassembler (`tools/cpu/mips`)

A table-driven decoder for the **MIPS I** instruction set plus the PSX's coprocessors (COP0 system
control, COP2 = the GTE), returning the repository's common
`Inst{Addr, Len, Mnem, Text, Flow, Target, HasTarget}` shape with the shared `Flow` enum
(Seq/Branch/Jump/Call/Return/IndJump/IndCall/Stop). MIPS makes this clean — a
fixed 4-byte instruction with 6-bit opcode/funct fields — but adds one thing no other core here has:
the **delay slot**. Every branch and jump is marked so the tracer and the interpreter can honour it
(the instruction *after* a branch always executes before control transfers). `jr $ra` classifies as
`Return`, `jalr` as an indirect call, `jr` through any other register as an indirect jump, and
`j`/`jal` carry region-computed targets.

**Validated on the game's own code.** Disassembling from the entry point produces clean, coherent
MIPS — the C-runtime prologue, a classic BSS-clear loop with its branch delay slot:

```
80040004  lui   $v0, 0x8008
80040008  addiu $v0, $v0, -24080     ; $v0 = 0x8007A1F0  (clear from here…)
8004000C  lui   $v1, 0x801F
80040010  addiu $v1, $v1, -9420      ; …to 0x801EDB34)
80040014  sw    $zero, 0($v0)        ; loop: zero a word
80040018  addiu $v0, $v0, 4
8004001C  sltu  $at, $v0, $v1
80040020  bne   $at, $zero, $80040014
80040024  nop                        ; ← branch delay slot
```

## 2. The CLIs (`dismips`, `codetracemips`)

- **`cmd/dismips`** — a linear disassembler (`-off`/`-len`/`-base`).
- **`cmd/codetracemips`** — the recursive-descent tracer (`-entry`, `-table` for jump tables,
  `-annotate`, `-o`), following the repo's `codetrace*` convention. Its one MIPS-specific twist is
  that it decodes and covers the **delay-slot** instruction before honouring each branch, and
  resumes a conditional branch/call *after* the slot. From the entry point it cleanly separates code
  from data, names subroutines by their call targets, and reports the indirect jumps it cannot
  resolve statically:

  ```
  traced $80010000-$8007A7FF: 56992 code, 379232 data; 213 routines,
                              51 unresolved indirect, 36 stop-hits
  ```

  The 51 unresolved indirect transfers are `jr`-through-register jump tables — the seed list for the
  static↔dynamic loop, resolved by watching the oracle run them live (§5).

## 3. The execution core (`tools/cpu/mips` CPU)

`tools/cpu/mips` carries a full R3000 interpreter (`cpu.go`/`exec.go`) — a `CPU` over a small `Bus`,
covering the whole MIPS I set: the ALU/shift ops, `mult`/`div` (with the R3000's specific
divide-by-zero and `INT_MIN/-1` results), the aligned and unaligned loads/stores
(`lwl`/`lwr`/`swl`/`swr`), all branches and jumps, the HI/LO pair, and COP0 exceptions
(`syscall`/`break`/overflow/address-error/interrupt) with `rfe` and the RAM/ROM vector select.
Unimplemented opcodes halt with the offending word so gaps are explicit.

Two R3000 pipeline hazards are modelled because real code depends on them:

- **Branch delay slot** — a PC/nextPC pair: `Step` advances PC to the delay slot before the branch
  runs, so the slot executes whether or not the branch is taken, and the branch only chooses the
  next PC.
- **Load delay slot** — a two-register-file scheme (reads see `R`, writes go to `out`, a pending
  load lands in `out` at the start of the *next* instruction). This reproduces the rule that the
  instruction immediately after a load sees the *old* register value, and that an ALU write in the
  load's delay slot beats the load — including the `lwl`/`lwr` forwarding idiom.

It is unit-tested with small hand-assembled programs proving both hazards, `jal`/`jr` call-return,
`mult`/`div`, `$zero` hardwiring and the add-overflow trap, and there is an env-gated differential
harness (`singlestep_test.go`) that diffs the core against the **SingleStepTests/psx** vectors
instruction-by-instruction against a reference test suite.

## 4. The GTE (`tools/cpu/mips/gte.go`)

The PlayStation's 3D math runs on the **GTE** (Geometry Transformation Engine, coprocessor 2): a
fixed-point matrix/vector engine with no floating point, reached through `cop2` commands and
`mfc2`/`mtc2`/`cfc2`/`ctc2` register moves. It is the one genuinely new, hard component. The GTE
implements the register file (with the screen-XY / screen-Z ordering-table FIFOs, the IRGB colour
packing and the leading-zero-count registers) and the transform + ordering-table + matrix-vector
command subset a racer leans on: **RTPS/RTPT** (rotate-translate-perspective), **NCLIP** (back-face
cross product), **AVSZ3/AVSZ4** (ordering-table Z averaging) and **MVMVA** — with the exact 44-bit
MAC overflow, IR saturation and FLAG semantics, and the hardware **unsigned Newton-Raphson**
perspective divide (the 257-entry reciprocal seed table). Correctness matters bit-for-bit because it
feeds both rendering and any later verification: golden tests confirm an identity RTPS reproduces its
input vector exactly, NCLIP computes signed area, AVSZ averages Z, and the UNR divide matches the
closed-form quotient to within one unit in the last place. It hangs off the CPU as an optional
coprocessor hook.

## 5. The machine oracle (`tools/platform/psx`)

The tracer's static reach ends at those 51 indirect jumps; resolving them needs an execution core
that *runs* the game so its behaviour can be observed. `tools/platform/psx` wires the CPU +
GTE to the PSX memory map: 2 MiB RAM (with its KSEG0/KSEG1 mirrors), the 1 KiB scratchpad, the
hardware I/O register window (`I_STAT`/`I_MASK`, `GPUSTAT`, the CD-ROM and DMA controllers,
timers), and a **high-level-emulated BIOS**. Rather than a firmware image, the BIOS `A0`/`B0`/`C0`
call vectors are intercepted at their fixed entry addresses and serviced in Go —
`InitHeap`/`malloc`, `memcpy`/`memset`, `printf`, `putchar`, the event/pad calls, `FlushCache` —
with unhandled calls logged once so the next one to implement is obvious. A minimal exception handler
at `0x80000080` services the critical-section `syscall`s.

Loading `SCUS-943.00` and running it, the C-runtime clears BSS, sets up the stack, and the game's
initialization proceeds through the BIOS into its CD setup. Early in the boot the game reaches its
**CD-ready wait loop** — a bounded counter poll around `0x8004B648` — which advances once the CD-ROM
controller and interrupt delivery (Part III) supply the status it watches. The game's **own debug
output** comes back through the HLE `printf` as TTY text:

```
$ bootoracle -image "Ridge Racer (Track 01).bin" -steps 50000000
booting SCUS-943.00;1: entry 0x80040004, text 0x80010000 (436224 bytes)

stopped at 0x8004B648 after 50000000 steps: budget reached
BIOS calls: InitHeap:1 printf:1 CdRemove:1 HookEntryInt:1 ChangeClearPad:1
            ChangeClearRCnt:1 syscall(1):1 syscall(2):1  A0(0x13):1 A0(0x9F):1
TTY:
CD_init:addr=800794ac
```

That line — `CD_init:addr=800794ac` — is Ridge Racer's own `CD_init` routine reporting the RAM
address of its CD subsystem, printed by the game and pulled through the emulated BIOS.

`extract/cmd/bootoracle` drives the oracle and exposes the instrumentation as flags:
`-trace`/`-tracen` (live disassembly of each executed instruction), `-bp ADDR` (breakpoint),
`-watch ADDR[:LEN]` (who wrote this address, attributed to the storing instruction's PC), `-shot`
(write the GPU display to a PNG), `-press` (scripted controller input), `-isr` (the vectored
interrupt entry, Part III), `-log` and `-tty`. With `codetracemips`, the static trace gives the
labeled disassembly and the oracle resolves the indirect jumps the static pass leaves open.

### Oracle tooling

| Command | Role |
|---|---|
| `tools/platform/psx/cmd/psxinfo` | inspect the disc: `-ls`, `-pvd`, `-extract PATH -o FILE`, `-exe PATH` |
| `tools/cmd/dismips` | linear MIPS disassembler (`-off`/`-len`/`-base`) |
| `tools/cmd/codetracemips` | recursive-descent tracer (`-entry`/`-table`/`-annotate`), delay-slot aware |
| `games/ridge-racer-psx/extract/cmd/bootoracle` | boot the game under the oracle (`-trace`/`-bp`/`-watch`/`-shot`/`-press`/`-isr`/`-log`/`-tty`, `-save`/`-load` machine savestates) |
| `games/ridge-racer-psx/extract/cmd/physprobe` | restore a race savestate, drive, and log the car state block once per physics frame as CSV |
| `games/ridge-racer-psx/extract/cmd/calltrace` | record the dynamic call tree under a root function (one frame of the physics pipeline in one run) |
| `games/ridge-racer-psx/extract/cmd/pccount` | trap a PC list and census hits per `(pc, $a0, $ra)` — "who runs this routine with which struct" |

The oracle also supports **machine savestates** (`tools/platform/psx/state.go` with the CPU/GTE
half in `tools/cpu/mips/state.go`): everything the game can observe — RAM, scratchpad, CPU
pipeline state (including a pending load-delay slot), GTE registers, GPU with its VRAM, CD
controller, IRQ/DMA/pad/BIOS-HLE state — serializes to a gob+gzip file (~1 MiB). Host
configuration (hooks, pad script, the disc) deliberately stays outside the snapshot, so one
500-M-instruction boot into a race becomes a reusable branch point for every Part VIII experiment.

---

# Part III — Program architecture

Past initialization the game is **interrupt-driven**: a per-field vertical-blank interrupt paces the
frame, the CD-ROM controller signals command completions and sector arrivals, and DMA completions
signal transfers. The runtime memory map is the 2 MiB main RAM (KSEG0 `0x80000000`, mirrored
uncached at KSEG1 `0xA0000000`), the 1 KiB scratchpad at `0x1F800000`, and the hardware I/O window
at `0x1F801000`; the executable occupies RAM from `0x00010000`, the BIOS/kernel scratch and the
interrupt-vector area sit in the low 64 KiB, and the stack grows down from `0x801FFFF0`.

## 1. Interrupt delivery

Two hardware registers gate every interrupt: **`I_STAT`** (`0x1F801070`, the pending-request bits)
and **`I_MASK`** (`0x1F801074`, the enable bits). A device sets its `I_STAT` bit; the CPU takes a
COP0 interrupt only when the bit is also set in `I_MASK` and interrupts are enabled in the status
register. The game unmasks three lines — **`I_MASK = 0xD`**: bit 0 (VBlank), bit 2 (CD-ROM) and
bit 3 (DMA). `I_STAT` bits are acknowledged by writing a word with zeros in the bits to clear.

The oracle raises the **vertical blank** synthetically on a fixed cadence of **250,000 instructions**,
approximating one NTSC field (on the order of half a million CPU cycles at roughly two cycles per
instruction), setting `I_STAT` bit 0 and refreshing the controller buffer each field. The game's
frame-delay loops count instructions against a VBlank-advanced counter, so the cadence is what makes
those waits terminate. When an enabled, unmasked line is pending, the CPU vectors to the exception
entry at **`0x80000080`**.

Ridge Racer installs its **own vectored dispatcher** rather than relying on the kernel's event
system. It registers the handler through the BIOS `HookEntryInt` call; on the retail machine the ROM
chains it in at `0x8004DF48`, an address that is never a stored pointer in the image (it is reached
only through the kernel's hook chain), so the oracle vectors delivery there via `-isr`. The handler
runs on a **dedicated interrupt stack** (top `0x8000D000`) so it cannot corrupt the interrupted
code's frame, and returns through a sentinel return address (`0x000000E0`) that the oracle catches to
restore the saved context. Nested delivery is suppressed while a handler is active — the handler
runs to completion with interrupts disabled, as the hardware entry leaves them. Critical sections
(`EnterCriticalSection`/`ExitCriticalSection`) toggle the previous-interrupt-enable bit in the COP0
status register, which the exception entry and `rfe` shift into and out of the current-enable bit.

## 2. The CD-ROM controller

The CD-ROM controller is four **byte-wide, index-banked** registers at `0x1F801800`–`0x1F801803`:
`0x1F801800` selects the register bank (index 0–3) in its low two bits, and the other three change
meaning with that index (command / parameter FIFO, response FIFO, interrupt-enable/flags, data
request). The game reaches them not through `lui`-immediate addresses but through a **pointer table
shipped in the executable at `0x8007945C`** and dereferenced at run time.

A command is issued by writing parameter bytes then a command byte; the controller answers
asynchronously with a queued sequence of interrupts — **INT3** (acknowledge), **INT2** (completion)
and **INT1** (data ready) — each delivering a response FIFO. The boot performs audio-volume setup
(index-2/3 writes), then `GetStat`, then the CD-DA table queries `GetTN` (`0x13`) and `GetTD`
(`0x14`); on a data-only disc these report a single track. A `ReadN`/`ReadS` sector read signals
INT1, and the ready sector is moved into RAM by DMA channel 3. The whole game is resident after the
initial load, so play issues no further reads — which is what lets the disc be swapped for a music
CD mid-race.

## 3. The DMA controller

Seven DMA channels share a control register (`DPCR`) and an interrupt register (**`DICR`**,
`0x1F8010F4`, whose per-channel completion flags in bits 24–30 are **write-1-to-clear**). Each
channel has three registers at `0x1F801080 + ch·0x10`: `MADR` (`+0`, RAM address), `BCR` (`+4`,
block size and count) and `CHCR` (`+8`, direction, sync mode and the `0x01000000` start bit). A
`CHCR` write with the start bit runs the transfer, after which the oracle clears the start bit, sets
the `DICR` channel-done flag and raises the DMA interrupt line. The game uses **channel 3** for
CD-ROM sector reads and **channel 2** for the GPU (Part IV). A channel's completion dispatcher polls
`DICR` and clears the flag by writing it back.

## 4. Controller input

The BIOS `InitPad` call registers the game's port-1 pad buffer and `StartPad` arms polling. Each
vertical blank the oracle deposits a standard digital-pad packet into that buffer, as the hardware
BIOS does: **byte 0 = `0x00`** (a controller is present), **byte 1 = `0x41`** (the digital-pad id),
and **bytes 2–3 = the 16-bit button state, active-low** (a zero bit is a pressed button; the idle
state is `0xFFFF`). Host input is scripted by instruction count through the oracle's `-press` flag,
which fills the packet at the VBlank cadence.

Driving that input walks the menus: **START** at the title advances to the GAME START menu (course /
transmission / car / sound panels and START / OPTION), where the cursor moves with the d-pad,
**CROSS** confirms and the shape buttons back out; confirming START begins a race. The car-select
screen shows each car's name and its acceleration / handling / traction / top-speed radar, updated
as the selection cycles.

---

# Part IV — The graphics pipeline

The GPU owns a **1 MiB frame buffer** — 1024 × 512 pixels of 16-bit colour, **5:5:5** with bit 15 as
a mask/semi-transparency flag — and is driven through two ports: **GP0** (`0x1F801810`) for drawing
and image transfers, **GP1** (`0x1F801814`) for display control. `GPUSTAT` reports the ready/idle
bits a busy-wait poll needs; `GPUREAD` returns image data during a VRAM→CPU copy. The GTE (Part II
§4) has already transformed geometry to screen coordinates; the GPU rasterizes it into VRAM, and the
display is scanned from a programmable window within the same VRAM.

## 1. The GP0 command stream

Each GP0 command is a byte opcode in the top of its first word followed by a fixed or image-sized run
of parameter words; the length is decoded per opcode so the command assembler knows when a command is
complete. The families the game uses:

| Opcode | Command | Words |
|---|---|---|
| `0x02` | fill rectangle (flat) | 3 |
| `0x20`–`0x3F` | polygon: flat/gouraud × textured, triangle/quad | 4–12 |
| `0x60`–`0x7F` | rectangle / sprite (fixed or variable size) | 2–4 |
| `0x80`–`0x9F` | VRAM→VRAM copy | 4 |
| `0xA0`–`0xBF` | CPU→VRAM copy (image upload) | 3 + pixels |
| `0xC0`–`0xDF` | VRAM→CPU copy (image store) | 3, then read-out |
| `0xE1`–`0xE6` | draw-mode settings | 1 |

The draw-mode settings carry the **texture page** base and colour depth (`0xE1`), the **texture
window** mask/offset (`0xE2`), the **drawing-area** clip rectangle (`0xE3`/`0xE4`) and the signed
**drawing offset** added to every vertex (`0xE5`). GP1 handles display state: reset (`0x00`/`0x01`),
display enable (`0x03`), the **display-area start** in VRAM (`0x05`) and the **display mode** (`0x08`,
selecting horizontal resolution 256/320/368/512/640 and 240- or 480-line/interlaced output).

## 2. Ordering-table DMA

Most drawing arrives not through direct GP0 writes but through **DMA channel 2 in linked-list sync
mode**. The game builds an **ordering table** — a chain of nodes each headed by a word of
`[word-count : 8][next-node address : 24]` — and points the channel at its head; the DMA walks the
chain, feeding each node's body words to GP0, until the end-of-list marker (`next` with bit 23 set).
Each node holds one primitive whose leading word is its GP0 opcode-and-colour.

Culled primitives — roadside sprites off this frame, whole track objects not currently visible — are
left linked in the table with their **opcode-and-colour word zeroed** so the leading opcode reads as a
no-op. A node whose primitive opcode is a no-op is skipped whole; feeding its interior texcoord/CLUT
words to GP0 would misread their high bytes as further opcodes and desynchronize the command stream.

## 3. Rasterization and shading

Triangles are filled by a **barycentric edge walk**; a quad is two triangles. Flat and gouraud
primitives interpolate the vertex colours. Textured primitives sample the current texture page — **4,
8 or 15 bits per texel**, the indexed depths resolving through a **CLUT** (colour lookup table) whose
VRAM position rides in the primitive — with the texture window applied so a repeated texture wraps
within its tile. A texel of index/value zero is treated as transparent and leaves the destination
untouched.

Textured surfaces are **modulated** by the primitive colour: `out = texel × colour / 128`, so a
colour of `0x80` per channel is neutral and brighter or darker colours scale the texel. A fully lit
model therefore needs real vertex colours, which come from the GTE. The **normal-colour lighting**
ops (`NCS`/`NCT`, `NCCS`/`NCDS`, `CC`, `CDP`) multiply the light matrix by each vertex normal, add
the light-colour matrix and background, and push the shaded colour to the GTE's RGB FIFO — **keeping
the command byte in the word's top lane**, which the game copies into the primitive's opcode slot.
`INTPL`/`DPCS` interpolate depth-cue fog. The rasterizer reads that lit colour back as the primitive
colour and modulates the texture by it; an all-zero colour, left by a depth-cue path the model does
not fully drive, is treated as a raw unmodulated texel rather than black.

## 4. Texture staging

Textures reach VRAM by **CPU→VRAM image upload** (GP0 `0xA0`): a destination rectangle and a run of
pixel words, streamed by a block-mode channel-2 DMA. The texture atlas occupies the **right half of
VRAM** (x ≥ 320), packed with the car, track, scenery, sprite and font pages plus their CLUTs.

For a scene change the game does more than upload: it **reads staged textures back out of VRAM** with
a **VRAM→CPU image store** (GP0 `0xC0`). The store names a source rectangle; `GPUREAD` then streams
those pixels out two per word, and — the path the scenery and font use — a **to-RAM (direction-0)
channel-2 DMA drains them into a RAM work buffer** (for the race scenery, `0x80176D10`). The game
transforms the buffer row by row and re-uploads it to a new VRAM location for the upcoming scene.
Without the read-back the work buffer holds nothing and the re-upload writes a blank region, so the
store direction is as essential to a correct frame as the upload.

## 5. Display and output

The display is **double-buffered** in the **left half of VRAM** (x 0–320): two frame buffers at rows
0–240 and 240–480. The game renders into the off-screen buffer while the other is scanned out, then
swaps by moving the display-area start (GP1 `0x05`) at vertical blank.

Rendered end to end, the oracle produces the game's frames to PNG with `-shot`: the attract **title
screen** (the waving checkered flag, the RIDGE RACER logo and `namco`), the attract demo's **textured
3-D race**, the GTE-lit **car-select model**, and the **in-race scene** — the road with lane markings,
opponent cars shaded in 3-D, the grandstand crowd, the city skyline, the `namco` barrier walls, the
speedometer and course map, and the TIME / POSITION / LAP HUD text:

```
bootoracle -image "games/ridge-racer-psx/Ridge Racer (Track 01).bin" -steps 430000000 \
           -press "start@380000000:380000,cross@386000000:380000" -shot race.png
```

---

# Part V — The asset index and load map

## 1. The boot load sequence

The loader streams every asset file off the disc during the boot and attract sequence, one file at
a time, through a **staging buffer at `0x8008045C`**. Each CD read is a channel-3 DMA of one 2048-byte
sector; the per-file destination and order (instruction counts are the oracle's synthetic timeline):

| File | Sectors (LBA) | RAM destination | Fate |
|---|---|---|---|
| `RR.VH` | 478–493 | `0x801DCBB4` | resident (VAB header) |
| `RR.VB` | 238–477 | `0x8008045C` | staged, uploaded to the SPU |
| `TEX4.TMS` | 494–501 | `0x8008045C` | staged, streamed to VRAM (Part VI §1) |
| `TEX0.TMS` | 502–843 | `0x8008045C` | staged, streamed to VRAM |
| `TEX1.TMS` | 844–940 | `0x8008045C` | staged, streamed to VRAM |
| `TEX2.TMS` | 941–1009 | `0x8008045C` | staged, streamed to VRAM |
| `TEX3.TMS` | 1010–1063 | `0x8008045C` | staged, streamed to VRAM |
| `MAP.RRM` | 1064–1196 | `0x8008045C` | **resident** — rendered in place |
| `OBJ.RRO` | 1197–1414 | `0x800C2918` | **resident** — rendered in place |
| `IDX.HED` | 1415 | `0x8007F134` | **resident** — the course grid |

A texture archive is fully consumed (streamed to VRAM at a paced rate, §VI.1) before the next file
overwrites the staging buffer. `MAP.RRM` is the last user of the buffer and simply stays there;
`OBJ.RRO` lands directly behind it (`0x8008045C` + 271,548 = `0x800C2918`). The renderer walks both
files in place every frame — nothing is unpacked or converted.

## 2. IDX.HED — the course grid

The game world is a **32×32 grid of cells**. `IDX.HED` is exactly this grid: **1024 little-endian
halfwords**, each the `MAP.RRM` section index occupying that cell, or `0xFFFF` for empty ground. A
pointer to the resident copy lives at `0x801DBB24`.

Two unit systems meet at the grid. Positions (the camera, the cars) are kept in **quarter model
units**: a cell is 2048 *position* units, but the section records' vertex coordinates are in model
units, where a cell is **8192** across — the grid walk shifts the rotated cell translation left by
two (`0x80012568`) before handing it to the GTE. (This is also why a section's geometry, spanning
thousands of model units, still fits its cell.)

The per-frame grid walk at `0x80012478` selects the visible cells: the camera's cell is its
position arithmetically shifted right by 11 (`x >> 11`, `z >> 11`); a direction-dependent table of
64 signed byte pairs (blocks of 128 pairs at `0x80056D64`, selected by the view octant) supplies
the neighbourhood offsets; and a cell (x, z) — both coordinates must be < 32 — is looked up as

```
section = grid[z*32 + 30 - x]        ; halfword; 0xFFFF = nothing to draw
```

Each visible cell gets a 16-byte entry in the cell list at `0x801DB060`: the section index and the
cell's camera-relative translation, `(x*2048 - camX, -camY, z*2048 - camZ)` in position units,
rotated by the camera matrix and scaled ×4 into the GTE translation — a section's world origin is
its cell corner, `(x*8192, 0, z*8192)` in model units. 258 cells are occupied, each by a distinct
section (0–257, matching `MAP.RRM`'s section count exactly); drawn as a map, the grid is the Ridge
Racer island — the circuit, the seaside, and the scenery hills around it.

---

# Part VI — Asset formats

## 1. TEX*.TMS — the texture-upload streams

A `.TMS` archive is a **VRAM upload schedule**: the sequence of CPU→VRAM transfers (Part IV §4)
that populate the texture atlas, stored ready to send. The streamer at `0x80037DB8` walks it a
block per frame — texture loading is spread across the attract sequence, which is why the title
screen appears long before the race pages are complete.

```
u32 format                     ; 0x100
repeat:
  u32 blockLen                 ; next block header at +4 + (blockLen & ~3); ≤ 0 ends the file
  u32 flags                    ; bit 3: a CLUT record precedes the image record
  if flags & 8:
    u32 len                    ; record length: 12 + w*h*2
    u16 x, y, w, h             ; destination rect in VRAM (w in 16-bit units)
    u16 pixels[w*h]
  u32 len                      ; image record, same shape; skipped when w or h is 0
  u16 x, y, w, h
  u16 pixels[w*h]
```

Both record kinds are raw VRAM rects; a CLUT is just a 16×1 strip (the 4-bit palettes here) that
the streamer re-sends with its block. An image whose area exceeds **2048 pixels** is uploaded in
**eight horizontal slices** of `h/8` rows, one per frame — the file stores the rect whole and only
the pacing splits it. Per archive: `TEX4` 163 blocks (23 distinct images — the early sky and font),
`TEX0` 143 (the biggest pages), `TEX1` 120, `TEX2` 102, `TEX3` 102.

The archives paint VRAM twice over: the **race scenery quadrant** — VRAM (640,256)–(1024,512), the
texture pages `0x1A`–`0x1E` the track's buildings and roadside art sample — holds one of **three
scenery sets**, and the archives deliver all three: the quadrant as it stands at the end of
`TEX1`'s stream is the **city** set (0), at the end of `TEX2`'s the **seaside** set (1), and at the
end of `TEX3`'s the third set (2). Each boundary is banked to RAM with eight 384×32 VRAM→CPU
read-backs (GP0 `0xC0`); the rotator at `0x800375FC` then pages sets between VRAM and the two RAM
banks (`0x80176D10`, `0x801A6D10`) row by row (paired `0xC0`/`0xA0` 384×1 transfers), tracking each
row's occupancy in the per-row arrays at `0x801DB460`/`0x801DB560` — the displaced set always lands
in the bank the incoming set vacated.

Which set is in VRAM follows the car's **course progress**: the selector at `0x800374F0` measures
the circular distance (`0x80037434`, positions modulo 65,536) from the car's progress to two
triggers — immediates in the course init at `0x80012D60`: `0xD00` and `0x6800` (`0x6400` for the
alternate course variant) — and keeps the city set while the car is between them, the seaside set
on the rest of the lap, sliding the row boundary gradually across a ±512-unit window around each
trigger. The third set is requested explicitly by attract-demo scenes and course variants, never by
lap position on the standard course.

Course progress itself is defined by the executable's **checkpoint table at `0x80059164`**: 256
gates of 20 bytes, searched per frame (`0x8001B68C`) for the gate containing the car; progress =
gate index × 256 + fraction. A gate decodes as `X = 0xF000 − (word0 >> 14)`,
`Z = word4 >> 14` (position units, i.e. model units ÷ 4; the X origin `0xF000` is the immediate at
`0x8001547C`), with the gate heading in the halfword at +10 and half-widths at +14/+16.

## 2. MAP.RRM — the track

`MAP.RRM` holds every track section's geometry as **GPU-shaped 40-byte textured-quad records**,
grouped per section into three consecutively stored draw classes (the renderer batches each class
separately — near geometry, far geometry, and a trailing class, each with its own shading and
subdivision treatment):

```
u16 sectionCount               ; 258
u16 pad
sectionCount × {
  u16 nA, nB, nC               ; record counts of the three classes
  u16 pad
}
Σ(nA+nB+nC) × TrackQuad        ; 40 bytes each, section by section, class A then B then C
```

The directory walker at `0x80011D64` converts the counts to running record pointers (each section's
geometry starts where the previous one ended); the layout accounts for the file exactly:
4 + 258×8 + **6,737**×40 = 271,548 bytes.

A `TrackQuad` is the in-file image of a textured quad, vertices first, then the texture words in
exactly the order the GPU packet wants them — the two spare high halfwords carry the depth bias and
the shade pair:

```
+0   int16 x,y,z  × 4          ; the corners (TL, TR, BL, BR), local to the section cell
+24  u8 u0,v0   u16 clut
+28  u8 u1,v1   u16 tpage
+32  u8 u2,v2   int16 bias     ; added to the ordering-table depth
+36  u8 u3,v3   u8 shade0, shade1
```

The transform helper at `0x800461E8` loads the packed vertices into the GTE (`RTPT` for the first
three corners, `RTPS` for the fourth), stores the projected `SXY` pairs, runs `AVSZ4` for the
ordering-table depth, and returns `IR0` — the depth-cue factor. The class walker (`0x800341A0` for
class B) screen-clips the quad, fades the two shade bytes with distance (`shade - OTZ>>10`,
clamped at zero), depth-cues the colour through `DPCS` (`0x80047844`), subdivides near quads
(`0x80047E38`, the workspace splits a quad into perspective-corrected pieces), and links the packet
at `OT[OTZ>>shift + bias]`.

## 3. OBJ.RRO — the object models

`OBJ.RRO` holds every 3-D object — the cars, the roadside scenery, the grandstand, the blimp — as
**319 objects**, each a run of polygon records of six fixed-size types:

```
u32 objectCount                ; 319
objectCount × {
  u16 n40, n48, n32, n64, n72, n56   ; record counts, one per type
  u32 runtime                        ; zero on disc; the loader stores the object's
}                                    ;   geometry address here (0x80011E48)
per object, in directory order:
  n40 × 40 bytes   n48 × 48   n32 × 32   n64 × 64   n72 × 72   n56 × 56
```

4 + 319×16 + 440,240 = 445,348 bytes, again exact. The types:

| Size | Shape | Layout |
|---|---|---|
| 40 | textured flat quad | identical to `TrackQuad`; no texture window |
| 48 | textured flat quad + tail | `TrackQuad` + an 8-byte texture-window rectangle (below) |
| 32 | untextured flat quad | verts + RGB colour at +24, depth bias at +28 (drawn as GP0 `0x2B`; the car shadows) |
| 64 | textured **lit** quad | verts + 4 × int16 normals at +24 + texture words at +48 (as `TrackQuad`'s, uv3's spare halfword unused); no window |
| 72 | textured lit quad + tail | the 64-byte layout + an 8-byte texture-window rectangle |
| 56 | untextured lit quad | verts + normals + a ready-made GP0 colour word `{r,g,b,0x38}` at +48 + bias at +52 |

The 48- and 72-byte tails are a **texture window** — a rectangle in the page `(u0, v0, w, h)` as four
halfwords — from which the draw code builds a GP0 `0xE2` window before the primitive so that
rectangle repeats. The register's raw 5-bit fields are `offsetX = u0/8`, `offsetY = v0/8`,
`maskX = (256 − w)/8`, `maskY = (256 − h)/8` (a full-size axis → mask 0, no repeat), and the GPU
samples `texel = (uv &^ (mask·8)) | ((offset & mask)·8)`. The buildings rely on this to tile a small
facade strip up their walls; sampling the raw UVs instead lands on the neighbouring texture.

The lit-quad renderer at `0x80046B18` is the busiest: `RTPT`/`RTPS` project the corners, **`NCLIP`
culls by winding** — a context flag selects which sign survives, so a model can be one- or
two-sided — the normals feed **`NCT`** (three) and **`NCS`** (fourth) with `RGBC = 0x808080` to
produce the four gouraud colours of a `POLY_GT4` packet, and the record's CLUT halfword gets a
**per-instance offset added** (`addu` at `0x80046BF0`) as it is copied into the packet — palette
recolouring, applied per drawn object.

## 4. Texture references

Every textured record carries the two GPU texture halfwords verbatim (Part IV §3): `tpage` selects
a 64×256-halfword page in VRAM (bits 0–3 page X ×64, bit 4 page Y ×256, bits 7–8 the colour depth:
4-bit, 8-bit or direct 15-bit) and `clut` names the palette strip (bits 0–5 X ×16, bits 6–14 Y).
UV bytes address texels within the page. Texel value 0 is fully transparent. Resolving a record's
texels therefore needs only the VRAM image the TMS streams build — no other lookup exists.

## 5. The car table

The executable's **car table at `0x80056B40`** defines the 13 cars (12 selectable plus the
unlockable `13" RACING`), 16 bytes each:

```
+0   u16   —                    ; aux object
+2   u16 body                   ; the display body object (car select, podium)
+4   u16 race                   ; the reduced in-race body (objects 24–35)
+6   u16 family                 ; canopy object; +1 axle, +2 shadow, +3 underbody
+8   i16   —                    ; −79…−83, family-dependent
+10  i16 wheelbase              ; the second axle's Z offset (−335, −317 or −320 by family)
+12  u16 spec                   ; per-car stat/livery selector
+14  i16   —                    ; -32767
```

The car-select carousel (`0x8001D2F4`) composes a car as: **body**, **canopy** (`family`),
**shadow** (`family+2`, twice, through the translucent path at `0x8001232C`), **underbody**
(`family+3`), and the **axle** (`family+1`) drawn twice — once at the car origin (the rear) and
once translated by the yaw-rotated `(0, 0, wheelbase)` (the front), with a spin matrix so the
wheels turn. Cars share part families: `F/A RACING` and `RT RYUKYU` use family 1, the SOLVALOUs
and XEVIOUSes family 17, the MAPPYs and GALAGAs family 7. The car names are the NUL-terminated
strings that open the text segment (file offset `0x800`), in table order: F/A RACING, RT RYUKYU,
RT YELLOW SOLVALOU, RT BLUE SOLVALOU, RT PINK MAPPY, RT BLUE MAPPY, GALAGA RT PLID'S, GALAGA RT
CARROT, RT BOSCONIAN, RT NEBULASRAY, RT XEVIOUS RED, RT XEVIOUS GREEN, 13" RACING.

## 6. Roadside object placement

Every `OBJ.RRO` object the race draws funnels through one small family of draw-object routines
(`0x800348E8` plain, `0x80034F78` translucent, `0x800372B0` shadow-class-only) that fetch the
geometry pointer from the object directory entry (`id×16+12`); read-watching the directory over a
full attract lap therefore enumerates **every** draw path. They split into three groups: the static
placement tables below, the **dynamic objects** placed by dedicated code (§7), and the car
compositor (§5).

The scenery that lines the course — buildings, signs, the grandstand, barriers and the first
tunnel's warning placards — comes from **static placement tables** in the executable, drawn each
frame by four near-identical iterators (`0x800157D8` plain, `0x800158E8` translucent, `0x800159F8`,
`0x80036778`) and culled against the frame's visible-cell mask. Every table is a run of **24-byte
records** ending on a negative id:

```
+0   s16 id        ; OBJ.RRO object index; a negative id ends the table
+2   s16 —         ; 0
+4   s32 X         ; world position in quarter model units (×4 → model units)
+8   s32 Y
+12  s32 Z
+16  s32 —         ; 0 in the model lists; a per-list draw-mode flag otherwise
+20  s16 yaw       ; Y-axis rotation, 4096 = one turn
+22  s16 —
```

The iterator reads the record's X/Z (as `X>>11`, `Z>>11`) to index the visible-cell bitmask and skip
off-screen objects, builds a Y-axis rotation from the yaw (`RotMatrix` at `0x80017EAC`, matrix
`[cos,0,-sin / 0,1,0 / sin,0,cos]`), and draws the object's geometry under that rotation translated to
`(X, Y, Z)`. The positions are in the same quarter-model-unit space as the grid (§V.2), so a placement
sits at `(X, Y, Z)×4` in the track's model units.

Two dispatchers select the tables. The **buildings dispatcher** (`0x800129E0`, mirrored at
`0x80014F80`) always draws `0x8006E85C` — the six warning placards (ids 191–194) in the first
tunnel — and `0x8006E904` (translucent glow/flag panels 247–249), then the day/night halfword at
`0x8017693C` picks the day pair (`0x8006EAFC` via `0x80036778`, `0x8006E9AC` via `0x800158E8`) or
the night pair (`0x8006F09C` via `0x800159F8`, `0x8006EA54` via `0x800157D8`) — the same buildings
as different object variants (e.g. 177 by day, 178 by night). The **structures dispatcher** (the
head of the scenery routine at `0x80015B98`) draws the start barriers (`0x80070360`, at night
`0x80070468`) and the grandstand list (`0x800703C0` with objects 60/61; at night a pointer variable
at `0x80079F98/9C` selects a single-record 61-variant list at `0x80070408/0x80070438`), then the
course-length word at `0x80176CBC` picks the course-split barrier walls (`0x80070510` short courses,
`0x800705E8` long), and at night `0x800706C0` adds translucent extras (101, 142, 165). The day pair
and night pair carry the same 59 building placements, so a placement is uniquely
`(id, X, Y, Z, yaw)`.

**The placards are draw-only.** Watched over a full attract lap and a scripted drive through a
placard relocated onto the racing line, table `0x8006E85C` is read by the draw iterator and the
transform setup only, and written by nothing: no collision system knows the placards exist in this
build. (Object 193 is a flattened sign lying on the road; the *swinging* sign hanging over the same
tunnel is dynamic — §7.)

## 7. Dynamic objects

Everything that moves — and a handful of fixed but animated objects — is placed by dedicated code,
one traced block per object, mostly inside the scenery dispatcher at `0x80015B98` (the full block
list with addresses lives in `extract/rr/dynamics.go`). Each block culls by the same visible-cell
mask, builds its own rotation, calls `SetTransform` (`0x80012148`) with a **16-byte position vector
in the executable's data segment** (X, Y, Z, W words in quarter model units), and draws through the
shared draw-object routines. The id immediates encode a fallback: every block clamps to object 1
when the loaded directory has fewer objects than it needs, which is why the iterators carry the
same `id ≥ count → 1` guard.

| Object | Ids | Position vector | Behaviour (traced) |
|---|---|---|---|
| Number girl | 250 | `0x80070870` | on the grid; upright billboard (pitch `0x400`, yaw `W+0x400`), translucent path |
| Swinging tunnel sign | 192 | `0x80070780` | hangs over the first tunnel's road; yaw `0xC00 ± sin(frame)/8` |
| Rotating sign | 175 | `0x80070720` | spins 16/frame (reversed at night) |
| Beacon | 257 | `0x80070820` | unrotated; CLUT offset cycles four palettes every four frames (the per-instance CLUT add of §3) |
| Start banner | 182/181 | `0x80070730` day, `0x80070740` night | gantry banner; 181 on a scheduled flicker frame |
| Big screens ×2 | 185 off, 183/184 on | `0x80070750`, `0x80070760` | video walls; the drawer re-points the screen quad's UVs through the directory each frame |
| Start-gate crowd | 67 + 68–83 | `0x800707A0` (night twin `0x80070810`) | the crowd panel waves: frame `(counter>>2)&15` |
| — near LOD | 146,145,152,153,164 + 147–151 + 154–158 | `0x800707A0`–`0x800707E0` | drawn instead of 67 when the camera word `0x80130CC4` < 900: base + parts + crowd-wave sequence (byte table `0x800100F4`, ids 98+b) + the start-semaphore column (lit frame 154+n from the rig table `0x80056C44`) |
| Camera crane | 258–260 (row of `0x80056C44`) | pivot `0x800707F0` | articulated rig (`0x800154B8`); one joint spins frame×16 |
| Start balloons | 166–173 + 279–318 | `0x80070830` day, `0x80070840` night (W = yaw) | released at the start; drift −X and climb with race progress `(0x80176AE4−90)/3` ∈ [0,360); the topper is `279+8·bank+frame` with the bank byte (`0x80079FA0`) picked by lap — the lap banner |
| Hot-air balloons ×2 | 139 | `0x80070850`, `0x80070860` | night classes only (flag bits 0x20/0x40 at `0x80080444`); rise 256 and 544 per frame |
| Helicopter | 188 body + 189 rotor | scripts `0x800734D0`, `0x80073790` | the rotor spins 331/frame about the pitched axis; **waypoint bytecode scripts** fly it around the whole course (below) |
| Airplane | 190 day / 251 night | struct `0x801D6DBC` | spawn vector `0x800739E4` = (22976, 0, 30912), then position += `(dir·200)>>16` per tick along the 16.16 direction at `0x800739F4` = (−8448, −4096, −17664) → Δ(−25, −12, −53) quarter units/tick (pitch −192, yaw `0xF00`), alive 1800 ticks; spawned at race start with a random age and re-spawned at course-progress gates. Positions wrap at 2¹⁶ quarter units — the world is a u16 torus |

The helicopter's flight is a small **bytecode program** (interpreter `0x800380A0`, jump table
`0x800107FC`, flight struct `0x801DB8FC`; the drawer applies pitch = angle0 − 1024, the model being
authored nose-up). Each opcode is a run of halfwords:

```
0  X, Y, Z, pitch, yaw, roll     teleport + attitude (X/Z unsigned)
1  X, Y, Z, gateS, gateL, dur    fly to waypoint; the gate (short-/long-course
                                 variant, <<8) holds until race progress passes
                                 it, dur is the flight time in ticks
2  dur                           hover for dur ticks
3|4|5  angle                     set pitch|yaw|roll
6|7|8  target, rate              seek pitch|yaw|roll at rate/tick
9  —                             restart with one of the two routes at random
                                 (pointer pair 0x8007A10C)
```

Both routes lift off from the helipad east of the start (36 and 30 waypoints), circle the course
and land back on it, hovering 9000 ticks before the next route. `rr/flight.go` decodes the scripts
and the airplane's glide; `geomoracle -only flight` verifies them (the airplane's per-tick delta
bit-exact over 1591 ticks; the interpreter's waypoint-target sequence matches the decoded key list
in order, 22/22 in the race window).

The day/night split (`0x80176BF8`) is the race class — the higher classes run at dusk/night and
swap models; `0x80176CBC` is the course length. `extract/rr/dynamics.go` decodes the catalog
(positions from the EXE vectors, ids and behaviour from the traced blocks), and
`geomoracle -only dynamics` verifies it: over the attract lap, each drawn dynamic object's
GTE-recovered position (`R_obj·Rᵀ·TR`, camera cancelled against a same-frame static anchor) must
land on the decoded vector.

---

# Part VII — Extraction and verification

## 1. The decoders (extract/rr)

`extract/rr` reimplements every format of Parts V–VI as pure decoders over the CD file bytes:
`idx.go` (the grid), `rrm.go` (sections and `TrackQuad`s), `obj.go` (objects and the six record
types), `objects.go` (the roadside placement tables and the yaw rotation), `dynamics.go` (the
dynamic-object catalog of §VI.7 — ids and behaviour from the traced code blocks, positions from
the EXE data vectors), `course.go` (the
checkpoint table and scenery-set triggers), `tms.go` (the upload blocks) and `vram.go` — a
1024×512 virtual VRAM built by replaying all five TMS streams, whose `Texel(page, clut, u, v)`
mirrors the rasterizer's addressing. The oracle never supplies data; it only verifies.

## 2. geomoracle — the differential

`extract/cmd/geomoracle` proves the decoders against the running game, three ways:

* **vram** — boots the game and, at the instant each TMS stream completes (the staging buffer's
  reuse order guarantees the boundary), compares the machine's VRAM **word-for-word** against the
  replay over every rect that archive uploads. Cells the game re-uploads afterwards (the title
  screen redraws parts of the pages while the last archive still streams) are excluded by
  upload-order tracking. All five archives match exactly.
* **cars** — drives the menu into car select and traps the lit-quad renderer's GTE loads (the
  `RTPT`/`RTPS`/`NCT`/`NCS` at `0x80046B6C`/`0x80046BDC`/`0x80046C74`/`0x80046CA0`). Every trap
  yields the record's address (register `$t1`) and the vertex/normal registers just loaded; each
  must land on a decoder-predicted record boundary and match its decoded vectors **bit-exact**
  (3,763 events, zero mismatches).
* **track** — the same in-race, adding the 40-byte-quad path (`0x80046250`/`0x8004626C`, record in
  `$a0`), which covers `MAP.RRM` sections and the flat-textured objects (4,475 events, zero
  mismatches). Two further in-race checks ride along: **placement** — the quad path's GTE
  translation is the camera-rotated cell offset, so `Rᵀ·TR` recovers each drawn section's world
  position, and every pairwise difference must equal the grid-cell delta × 8192 (this pins the
  cell pitch, the x mirror and the orientation at once); the **race texture VRAM**, compared
  word-for-word against the pure file-byte reconstruction (the boot replay with the scenery
  quadrant restored from its end-of-`TEX1` snapshot): 360,448 words, zero mismatched; and **object
  placement** — each drawn `OBJ.RRO` object's `Rᵀ·TR` world position, whose every pairwise delta
  must equal the decoded placement table's, pinning both the table and the ×4 unit scale.
* **dynamics** — runs the attract demo lap and, for every drawn dynamic object with a fixed
  position and constant rotation (the girl, beacon, start banner, both big screens, the crowd
  gate), recovers `R_obj·Rᵀ·TR` and checks it against the decoded catalog position through a
  same-frame static anchor (the camera cancels in the pairwise delta).
* **flight** — samples the helicopter's and airplane's flight structs once per game tick in the
  race: the airplane's position delta must equal the decoded `(dir·speed)>>16` step exactly, and
  the helicopter's waypoint-target field must walk the decoded script's key list in order.

## 3. webexport — the shipped models

`extract/cmd/webexport` builds `site/public/ridge-racer-psx/` from the image alone. The **models**
stage emits one GLB per car, composed exactly as the carousel draws it (body + canopy + underbody +
both axles); the **levels** stage emits the road as one `track.glb` (all 258 sections at their grid
cells, 8,192 model units apart), each roadside object as its own `obj-NN.glb`, and
`course.objects.json` — the placement list (`{model, pos, rot}` plus each record's address, flag and
yaw), including every fixed-position dynamic object of §VI.7 (flagged `dynamic`; the airplane has
no fixed position and appears only as a model). The dynamic objects also ship as named
`special-NNN.glb` viewer models (the helicopter with its rotor, the camera crane assembled), and
`course.paths.json` carries the flight paths — the two helicopter routes (waypoints in GLB units
with per-key flight/hover ticks and commanded yaws, plus body/rotor part GLBs so the rotor spins)
and the airplane's glide (spawn, per-tick delta, 1800-tick life). The course renderer animates
both on the 30 Hz tick timeline: the helicopter flies the routes alternately (terminal helipad
hovers capped for watchability), the airplane jumps back to its spawn at the end of its life. Textures resolve against the race texture states of §VI.1: all three scenery-set VRAM images
are rebuilt from the TMS replay, and a section's (or object's) quadrant-page quads sample the set
active at its position — the nearest checkpoint gate gives the progress, the traced trigger rules
give the set (no object id spans two sets). Every distinct page+CLUT(+set) is baked into a tiled
atlas embedded in the GLB (nearest-neighbour sampling, texel 0 cut out via alpha masking); PSX
coordinates become glTF's through `(x, -y, -z)` at 1/1024 scale, and an object's yaw becomes a
three.js `rotation.y` of the same angle. The Studio's `rr-course` renderer flies through the course
(shared `FlyCam`), toggles the roadside-object layer, and shows each object's record address, flag
and rotation in a click-to-inspect card.

---

# Part VIII — The driving simulation

Everything below was traced in a *driven* race: the oracle's pad script walks the menus, a
savestate is taken just as the start lights go out (~520 M instructions), and every experiment
branches from that state — accelerate, steer, scrape a wall — while `-watch` attributes each field
of the car state to the instruction that wrote it. One boot-time fact matters up front: the
**attract demo is a replay, not a simulation**. A leaf at `0x8002A894` copies 40-byte recorded
frames (position, angles, progress for two cars) out of a table at `[0x80080434]` straight into
the car blocks, and none of the physics below runs. Physics facts can only be observed by playing.

## 1. The car state block and the frame pipeline

The player car is a **~280-byte state block at `0x80080194`**; the eleven opponents use the same
layout in an array at `0x801ECE34 + n × 0x114`. The fields that matter (offsets in hex):

| Offset | Field |
|---|---|
| +00 | active flag; +02 car model index (0–12, the Part VI §5 table order) |
| +08 | course progress: `gate × 256 + fraction`, **decreasing** as the car advances, wrapping at the course length (`[0x801FCC04]`) |
| +10/+14/+18 | world X, Y, Z (position units; the grid cell is 2,048) |
| +20/+24/+28 | drawn pitch / **travel direction** / drawn roll (angles, 4096 = 360°) |
| +38 | wheel-rotation accumulator (`+= speed`, bit 0x1000 flags "slow") |
| +58… | a sub-block the code addresses as `car+88`: +74 steering (±4096), +82 gear, +84 engine/wheel speed, +B8 airborne flag, +BA air time, +BC/+C0 vertical velocity |
| +60/+68 | velocity X / Z (used by the track sampler and the drawers) |
| +A0 | **speed** — the velocity magnitude the whole model revolves around |
| +A4 | per-frame **thrust magnitude** (engine force after drag) |
| +A8 | **velocity direction** (where the car is actually going) |
| +AC | **steer target** (where the front wheels point) |
| +B0 | drive-force/speedometer value (16.16; the mph readout derives from it) |
| +B4 | traction state (0 = grip … 3 = wheels loose: launch burnout, drift) |
| +C8/+CA | throttle / brake flags (0x100 while held) |

The player's per-frame driver is `0x8001BD80` (reached through a mode dispatcher; `jal` sites at
`0x80013448`…`0x80014C70` select attract/race/etc.). Traced with `calltrace`, one 30 Hz frame is:

```
0x8001B9D4  read pad        → steering ±4096 (slew 1280/frame, snap through zero)
0x800195B0  engine          → gears, throttle force, drag → thrust +A4, speed +A0
0x800176A4  track sampler   → road frame + four wheel heights under the car
0x8001B68C  gate search     → progress +08; fails if the new position leaves the road
0x8001AC54  car collision   → 6-point footprint vs. every active opponent
0x80032E94…0x80033600        five post passes: bump impulse, wobble oscillator,
                             jump/landing, brake dive, suspension roll/pitch
```

then the driver either commits the integrated position (`0x8001C450/454/458`) or, if the gate
search or the car check flagged contact, runs the collision response of §5.

Input, incidentally, does **not** come from the BIOS pad buffer the menus use. The game reads the
raw packet at `0x80176948` (the `InitPad` buffer), rebuilds it active-high with the bytes swapped
(`0x8002E12C`: `buttons = ~(b[2]<<8 | b[3])`, the SDK's `PADL*/PADR*` layout), and per-function
**mask globals** at `0x801DAFAC+` map buttons to actions — steer left/right `0x8000/0x2000`,
accelerate `0x0040` (cross), brake `0x0080` (square). A pad-id check for `0x23` selects a NeGcon
path with analog steering.

## 2. Engine, gears, forces

`0x800195B0` owns the longitudinal model. The transmission is a **7-row × 20-byte gear table
built at race init at `0x801DB71C`** — per gear `{ratio, torque-low, torque-high, back-torque,
speed ceiling}`; the ceilings run 6180, 6500, 6820, 7140, 7460, 7780, 8100 speed units across the
six forward gears. Throttle is smoothed into `[0x8007F12C]` (0–256, slewing ±25/frame — even a
digital button gives the engine an analog ramp), torque comes from the gear row, and a drag
accumulator collects: rolling/aero drag, extra drag when the car's travel direction disagrees
with the road direction (`0x80019528` vs. the gate frame), a wheelspin penalty while the traction
state (+B4) is 3, and a **landing penalty** (`200 + 20 × airtime/3`) after a jump. The net force
becomes the frame's thrust `+A4`; the speed `+A0` also decays by a drag factor — ×0.996 normally,
×0.94 in the heavy path (off-throttle/brake). Traction state 3 additionally **halves the steering
rate** and feeds a random jitter (`rand & 3 × (brake+8) >> 8`) into the model — the burnout
wiggle.

## 3. The velocity core — where the drift comes from

The lateral model lives in `0x80027BE4` + `0x800265D0`, and it is neither a bicycle model nor
rails: the car's velocity is stored in **polar form** — magnitude `+A0`, direction `+A8` — and
each frame the engine's thrust is added to it as a second vector:

1. The steer target `+AC` rotates by `steering × sensitivity ([0x80079FA8]) >> 8` per frame
   (scaled by speed when slow, halved when the wheels are loose).
2. The travel direction `+24` **eases toward `+AC` at 20 % per frame** (`0x80027C10`:
   `dir += Δ × 20 / 100`) — the front wheels drag the drivetrain around, not the car.
3. `0x800265D0` forms the vector sum `v = speed·dir(+A8) + thrust·dir(+24)`: the new magnitude
   comes from the **law of cosines** (`0x800465F4` is an integer square root), the new direction
   from an **atan2** (`0x8001805C`) — velocity integration done entirely in angles and
   magnitudes, GTE-style fixed point, no Cartesian velocity state at all.

Drift is then *emergent*, not a scripted mode: at 90 mph the thrust vector (~45 units) is tiny
against the velocity vector (~700), so the velocity direction can only creep toward where the
wheels point — the car runs wide, nose in, tail out. Log of a full-lock left at 102 mph
(steer target / travel dir / velocity dir, in angle units):

```
frame   +AC steer   +24 travel   +A8 velocity
  4       1186        1198          1199
 10       1065        1130          1166
 16        939        1019          1080
 22        879         923           978      ← target clamped at lock
 28        879         892           918      ← still 39 units of tail-out slide
```

Braking during the turn shrinks `+A0` (the ×0.94 drag path) while the 20 %-easing keeps rotating
the thrust — which is exactly the arcade power-slide: brake-tap to break the speed, the nose
whips in, throttle catches the exit.

## 4. The road: suspension, jumps

`0x800176A4` samples the checkpoint tables (`0x80059164` and a parallel edge table at
`0x800596C8`) around the car and produces the road frame plus **four wheel-contact heights**
(globals `0x80176C28/38/48/58`). The fifth post pass (`0x80033600`) turns them into the body
attitude: left/right height difference → drawn roll (+28), front/rear difference → drawn pitch
(+20), the average → suspension-settled Y. The fourth pass (`0x800334B4`) adds pitch dive — an
accumulator that ramps while braking (or during the race-start launch) and relaxes ×3/4 per
frame. There is no spring/damper integration per wheel — attitude is *derived* from the sampled
surface each frame, plus these two accumulators.

Jumps are **data-driven**: `0x80033294` compares the car's progress remainder against per-class
course positions (`0x3400/0x7900/0xE900/0x5D00/0x7700` — the hill crests) and, above a speed
threshold (1216), launches the car with a vertical velocity from a per-crest table at
`0x80073430`. While airborne (+B8), gravity integrates the vertical velocity and steering keeps
working; landing fires event 1 (below), the landing drag penalty of §2, and — after long air
time — a dust effect (`0x80015B14`).

## 5. Collision: the bump and the wobble

Two detectors feed one response path in the driver's tail:

* **Road edge** — the gate search `0x8001B68C` re-locates the car's *proposed* position (plus
  four footprint corners via `0x80017404`) against the road quad between checkpoint gates; if the
  position falls outside, the frame's move is rejected (`s2 ≠ 0`) — the "wall" is the road
  boundary itself, there is no separate wall geometry.
* **Car-to-car** — `0x8001AC54` transforms every active opponent into the player's local frame
  and tests a **6-point footprint** (from a table at `0x80010128`: rear corners ±(−19, 22),
  mid ±(35, 24), nose ±(79, 25)) against the opponent's extent.

On contact the driver **skips the position commit** and instead (`0x8001C500…0x8001C668`):

* computes a push vector — for car contact, the blocked position delta; for a road-edge hit, the
  road direction rotated ±90°/16 (`0x80026B50` + sin/cos), i.e. *away from the edge, along the
  road* — and hands it to `0x80032F6C`;
* `0x80032F6C` latches the impulse into globals (`0x801ECCA0/A8`), arms a **30-frame bump state**
  (`[0x80131040] = 30`, flag `0x801DB6CC`), and fires a wobble event;
* multiplies the speed `+A0` by **0.7** and the drive force by 0.8 (`0x8001C618–0x8001C668`) —
  the signature Ridge Racer wall tax, applied once per contact frame;
* docks 5,000 progress units from a standings accumulator (`0x80176AC8 −= 5000`).

The **bump** is the first post pass (`0x80032E94`): while the bump state is armed, position gets
`impulse/8` added each frame and the impulse decays ×7/8 — an exponential shove away from the
contact that dies out over about a second. Because the car keeps thrusting toward the wall, a
shallow scrape settles into a limit cycle: watch of a 100-mph wall graze shows contact frames
about every 28 frames, each one a clean ×0.70 on the velocity vector (`(15152, −4220) →
(10603, −2953)`) with the heading untouched.

The **wobble** is the second post pass (`0x80033140`): wobble events (`0x80032FB4`, codes 1–5 —
landing, road-edge, car contact, …) arm a counter, and each frame adds a **damped sinusoid** —
`sin(counter × 3 × 4096 / 30) × counter × amplitude >> 7`, three full oscillations over 30 frames
with linearly dying amplitude — to the drawn **roll** for side contact or the drawn **pitch** for
landings. The same frame log shows it directly: after an impact the roll rings
`0 → 17 → 0 → −13 → 0 → 9 → …` — the car visibly rocks three diminishing times over one second.
The bump moves the car; the wobble only moves the *body* (and the chase camera reads the same
angles) — the handling model underneath is disturbed only by the one-off ×0.7.

Opponents get the same treatment in miniature: their updater carries the airborne/landing state
and a landing **bounce** — a decaying sine subtracted from Y (`0x80023E5C`) — so AI cars crest
the hills and slam down believably without owning a suspension.

## 6. The opponents — rails with a race engineer

The eleven AI cars are **not** driven by the player's physics. Their update splits by distance:

* **Far opponents** (out of sight) run a one-dimensional integrator inside the scheduler
  (`0x80024B9C`, the writer at `0x800253F0`): progress advances along the racing line at the
  commanded speed, and the position is simply the track sampler's centerline point — pure rails.
* **Near opponents** (the ~5 around the player; the same set the drawer at `0x80020BB0` renders)
  run a kinematic updater (`0x80023B48`): `speed +88` integrates `accel(+100)/3` per frame,
  velocity is `speed × (sin, cos)(heading +C4)`, a slide angle (+78) rotates the velocity during
  avoidance, and `0x80023570` steers the heading to stay inside a **lateral envelope** on the
  road (bounds `[0x80130CBC]`/`[0x8008044C]`, the latter maintained per-frame — lane keeping, and
  the mechanism that makes them swerve around the player rather than through). They also get the
  jump/landing treatment of §5. There is no engine, no grip, no drag — the "physics" is position
  integration; everything else is control law.

The interesting part is the **speed controller** (`0x800228A0` per opponent per frame). Each car
steps its speed toward a target at a state-given rate (`0x80022EE8`), and the target comes from a
per-car mode field (+A2) that a **player-proximity classifier** (`0x800213AC–0x800214FC`)
switches: it computes the opponent's circular progress distance to the player
(`opp+08 − [0x8008019C]`, wrap-corrected) and, inside ~200–500 progress units, promotes the car
out of its cruise mode:

| Mode | Target speed |
|---|---|
| default | scripted cruise speed (+FC / +108, set per car at init) |
| 1 | +10C boosted by **+10 %** when its duel flags allow |
| 2 | +110 boosted by **+7 %** when the player-side flag `[0x8008018C]` is set |
| 3 | the speed of the opponent indexed by +96 (**the car ahead**) **+6.7 %** — a chase mode |

That is the rubber band, and it is deliberately asymmetric: the boosts only arm near the player,
so back markers brake to a beatable pace when you catch them while the leaders cruise their
script far ahead. The AI never reads the player's *inputs* and never gets the player's physics —
`pccount` over a live race shows the full pipeline (`0x8001BD80` → `0x800195B0` → `0x80027BE4` →
`0x800265D0` → `0x8001AC54`) executing for exactly one struct: `0x80080194`.

So: the opponents are rails plus a control layer — lane-keeping steering, scripted cruise speeds,
a proximity-triggered chase state machine — while the player's car is a genuine (if stylised)
simulation: polar-form velocity with vector-added thrust, emergent drift from the 20 % easing,
data-driven jumps, and a collision response tuned to feel like a bump (decaying impulse), a
wobble (damped sine on the body) and a penalty (×0.7) rather than a crash.
