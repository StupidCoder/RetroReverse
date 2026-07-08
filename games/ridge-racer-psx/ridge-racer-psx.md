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
  texture pages and CLUTs, and the executable's car table.
* **Part VII** — **extraction and verification**: the pure-Go decoders in `extract/rr`, the
  `geomoracle` differential that checks them bit-exact against the running game, and the
  `webexport` pipeline that bakes the textured car and course GLBs.

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
| `games/ridge-racer-psx/extract/cmd/bootoracle` | boot the game under the oracle (`-trace`/`-bp`/`-watch`/`-shot`/`-press`/`-isr`/`-log`/`-tty`) |

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
texture pages `0x1A`–`0x1E` the track's buildings and roadside art sample — is uploaded by
`TEX0`/`TEX1`, then **saved to RAM** the moment `TEX1`'s stream completes (eight 384×32 VRAM→CPU
read-backs, GP0 `0xC0`), and the later archives paint the menu screens over the same pages. When
the race starts, the saved image is **restored row by row** (256 paired `0xC0`/`0xA0` 384×1
transfers — the read half banks the menu art for the way back). The race therefore renders from
the quadrant as it stood at the end of `TEX1`, a state that exists in no single archive and at no
single moment of the boot replay.

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
| 40 | textured flat quad | identical to `TrackQuad` |
| 48 | textured flat quad + tail | `TrackQuad` + 8 bytes |
| 32 | untextured flat quad | verts + RGB colour at +24, depth bias at +28 (drawn as GP0 `0x2B`; the car shadows) |
| 64 | textured **lit** quad | verts + 4 × int16 normals at +24 + texture words at +48 (as `TrackQuad`'s, uv3's spare halfword unused) |
| 72 | textured lit quad + tail | the 64-byte layout + 8 bytes carrying the base colour |
| 56 | untextured lit quad | verts + normals + a ready-made GP0 colour word `{r,g,b,0x38}` at +48 + bias at +52 |

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

---

# Part VII — Extraction and verification

## 1. The decoders (extract/rr)

`extract/rr` reimplements every format of Parts V–VI as pure decoders over the CD file bytes:
`idx.go` (the grid), `rrm.go` (sections and `TrackQuad`s), `obj.go` (objects and the six record
types), `tms.go` (the upload blocks) and `vram.go` — a 1024×512 virtual VRAM built by replaying
all five TMS streams, whose `Texel(page, clut, u, v)` mirrors the rasterizer's addressing. The
oracle never supplies data; it only verifies.

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
  cell pitch, the x mirror and the orientation at once) — and the **race texture VRAM**, compared
  word-for-word against the pure file-byte reconstruction (the boot replay with the scenery
  quadrant restored from its end-of-`TEX1` snapshot): 360,448 words, zero mismatched.

## 3. webexport — the shipped models

`extract/cmd/webexport` builds `site/public/ridge-racer-psx/` from the image alone. The **models**
stage emits one GLB per car, composed exactly as the carousel draws it (body + canopy + underbody +
both axles); the **levels** stage places all 258 sections at their grid cells (8,192 model units
apart) and emits the whole course as one GLB. Textures resolve against the **race-time** virtual
VRAM (the TMS replay with the scenery quadrant restored from its end-of-`TEX1` snapshot, §VI.1)
and are baked per referenced page+CLUT pair into a tiled atlas embedded in the GLB
(nearest-neighbour sampling, texel 0 cut out via alpha masking); PSX coordinates become glTF's
through `(x, -y, -z)` at 1/1024 scale. The manifest lists the 13 cars and the course; the Studio
renders them with its stock GLB viewer.
