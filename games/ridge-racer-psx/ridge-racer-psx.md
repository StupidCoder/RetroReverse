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
  its load layout and embedded string tables. *(done)*
* **Part II** — the **MIPS R3000 / PSX toolchain**: a MIPS disassembler, a recursive code-tracer,
  an execution core (with the R3000's branch- and load-delay slots), a bit-exact **GTE**, and a
  **machine oracle** with an HLE BIOS that **boots the real game**. *(done — the oracle boots Ridge
  Racer into its CD-ready wait loop, and the game's own
  debug output comes back through the emulated BIOS)*
* **Part III** — CD-ROM + interrupt delivery to advance the boot past the CD wait loop into the
  interrupt-driven main loop. *(planned)*
* **Part IV** — the asset formats: the `RR.VB`/`RR.VH` sound bank, the `TEX*.TMS` textures, and the
  `MAP.RRM`/`OBJ.RRO` track and model data. *(planned)*
* **Part V** — the **GPU**: the GP0/GP1 command stream, VRAM and the ordering-table DMA, and a
  screenshot of the first rendered frame. *(planned — the GTE it depends on is done)*

Methods: purely static analysis of the shipped image, plus dynamic analysis via the PSX oracle.
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
job unique to a CD image versus a flat disk; the reader folds it into the block accessor, exactly as
the Amiga `adf` reader hands back 512-byte blocks:

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
catalog. Counts and sizes are observed facts; the *purpose* of each format is an inference to be
confirmed by decoding in Parts IV–V, not asserted here:

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
and `.RRO` are the track and object data. All of these are Part IV/V targets. Re-list at any time:

```
go run retroreverse.com/tools/platform/psx/cmd/psxinfo -ls "Ridge Racer (PSX)/Ridge Racer (Track 01).bin"
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
go run retroreverse.com/tools/platform/psx/cmd/psxinfo -exe "SCUS-943.00;1" "Ridge Racer (PSX)/Ridge Racer (Track 01).bin"
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

Everything past Part I needs to read and run MIPS. The toolchain mirrors the existing per-CPU
packages (`tools/cpu/arm`, `tools/cpu/x86`, `tools/cpu/sm83`) and lives at **`tools/cpu/mips`** (the CPU) and
**`tools/platform/psx`** (the disc reader, executable loader and machine), platform-neutral under the shared
`retroreverse.com/tools` module.

## 1. The disassembler (`tools/cpu/mips`) — done

A table-driven decoder for the **MIPS I** instruction set plus the PSX's coprocessors (COP0 system
control, COP2 = the GTE), returning the repository's common
`Inst{Addr, Len, Mnem, Text, Flow, Target, HasTarget}` shape with the same `Flow` enum
(Seq/Branch/Jump/Call/Return/IndJump/IndCall/Stop) as the other cores. MIPS makes this clean — a
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

## 2. The CLIs (`dismips`, `codetracemips`) — done

- **`cmd/dismips`** — a linear disassembler (`-off`/`-len`/`-base`), paired with `disarm`/`disx86`.
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

## 3. The execution core (`tools/cpu/mips` CPU) — done

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

## 4. The GTE (`tools/cpu/mips/gte.go`) — done

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
coprocessor hook, exactly as `arm.CPU` routes CP15.

## 5. The machine oracle (`tools/platform/psx`) — done, and it boots the game

The tracer's static reach ends at those 51 indirect jumps; the answer is the same one the DS and
DOS titles needed — an execution core that *runs* the game so its behaviour can be observed. `tools/platform/psx` wires the CPU +
GTE to the PSX memory map: 2 MiB RAM (with its KSEG0/KSEG1 mirrors), the 1 KiB scratchpad, the
hardware I/O register window (`I_STAT`/`I_MASK`, a "ready" `GPUSTAT`, DMA channel auto-complete,
timers), and a **high-level-emulated BIOS**. Rather than a firmware image, the BIOS `A0`/`B0`/`C0`
call vectors are intercepted at their fixed entry addresses and serviced in Go —
`InitHeap`/`malloc`, `memcpy`/`memset`, `printf`, `putchar`, the event/pad calls, `FlushCache` — the
same approach `tools/platform/dos` takes with `INT 21h`, with unhandled calls logged once so the next one to
implement is obvious. A minimal exception handler at `0x80000080` services the critical-section
`syscall`s.

**Result: the oracle boots the real game.** Loading `SCUS-943.00` and running it, the C-runtime
clears BSS, sets up the stack, and the game's initialization proceeds through the BIOS into its CD
setup — tens of millions of instructions of real code with no wrong turn and no unimplemented-opcode
halt. It parks in **Ridge Racer's own CD-ready wait loop** (a bounded counter poll around
`0x8004B648`), where it idles because the CD drive and its interrupt are not yet wired to change
the status byte it is watching — precisely the Part III entry point.

The clinching evidence is the game's **own debug output**, captured through the HLE `printf` and
returned as TTY text:

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
address of its CD subsystem, printed by the game and pulled straight out of the emulated BIOS. It
both validates the CPU/BIOS on a large body of real code and confirms exactly where the boot now
stands.

`extract/cmd/bootoracle` drives all of this and exposes the repo's RE instrumentation as flags:
`-trace`/`-tracen` (live disassembly of each executed instruction), `-bp ADDR` (breakpoint),
`-watch ADDR[:LEN]` ("who wrote this address", attributed to the storing instruction's PC), `-log`
and `-tty`. Together with `codetracemips`, that closes the loop: the static trace gives the labeled
disassembly, and the oracle resolves the indirect jumps the static pass left open.

### Oracle tooling

| Command | Role |
|---|---|
| `tools/platform/psx/cmd/psxinfo` | inspect the disc: `-ls`, `-pvd`, `-extract PATH -o FILE`, `-exe PATH` |
| `tools/cmd/dismips` | linear MIPS disassembler (`-off`/`-len`/`-base`) |
| `tools/cmd/codetracemips` | recursive-descent tracer (`-entry`/`-table`/`-annotate`), delay-slot aware |
| `Ridge Racer (PSX)/extract/cmd/bootoracle` | boot the game under the oracle (`-trace`/`-bp`/`-watch`/`-log`/`-tty`) |
