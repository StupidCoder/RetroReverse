# Ultima Underworld (MS-DOS) — executable format and game analysis

Ultima Underworld: The Stygian Abyss (Blue Sky Productions / Origin Systems, 1992) is a
first-person, texture-mapped 3D dungeon crawler — one of the earliest games to render a
freely-looking, slope-and-stair 3D world in real time on a PC. This document reverse-engineers
the *shipped MS-DOS game* from its files alone, the same way the other titles in this repository
are taken apart: no third-party emulator, debugger or disassembler, no released source, and
nothing about the file formats taken from external documentation. Everything here is derived
from the bytes on disk via our own tools.

It is the repository's **first MS-DOS / x86 title**. Every prior game ran on a CPU we already
had a disassembler and (usually) an execution core for — 6502, 68000, Z80, SM83, ARM. Ultima
Underworld runs on the **Intel x86 in 16-bit real mode**, which we have *no* tooling for yet, so
a large part of this project is building that toolchain: a real-mode x86 disassembler, a
recursive code-tracer, and eventually an execution core to use as an **oracle** (Part II), in
the same spirit as `tools/arm`, `tools/m68k`, and `tools/mos6502`.

Image: `game/UW.EXE`, 561,744 bytes, MD5 `0f58c92a45b8d8d5bba498c59eb111c2` (pinned in
`../README.md`). The complete `game/` folder (executable plus ~11 MB of data) is a copyrighted
commercial release and is **not committed** — it is `.gitignore`d; only the extraction tools,
their tests and the rendered outputs live in the repository.

## Contents

* **Part I** — the DOS executable and data catalog: the **MZ** header and real-mode load
  layout, the relocation table, the appended data past the load image, and an inventory of the
  on-disk asset folders (`DATA` / `CRIT` / `CUTS` / `SOUND`). *(started — header decoded)*
* **Part II** — the x86 toolchain: a 16-bit real-mode **x86 disassembler**, a recursive
  **code-tracer**, and an **execution-core oracle** with a minimal DOS around it — all built and
  validated by booting UW.EXE's real startup code, mirroring the ARM/68k/6502 toolchains.
  *(done — the oracle boots the game to its overlay manager)*
* **Part III** — the boot chain and overlay system: the DOS entry at `CS:IP`, program startup, and
  UW's **overlaid code segments** (the Microsoft-C `INT 3Fh` scheme, overlay store at file
  `$66B30`). *(done — the oracle runs the game's own overlay handler and, with a faithful DOS/PC
  around it (MCB memory, EMS 4.0, files, video BIOS, keyboard/VGA/Sound-Blaster ports, timer),
  **runs the real game**: intro cutscene, new-game initialisation, title screen — with PNG
  screenshots straight from the emulated VGA)*
* **Part IV** — asset formats: the palettes (`PALS.DAT` / `ALLPALS.DAT`), the `.GR` image
  banks, the `.TR` wall/floor textures, the `.BYT` full-screen images, the fonts, and the packed
  string table (`STRINGS.PAK`). *(planned — static decode)*
* **Part V** — the world and the **3D renderer**: the level archive (`LEV.ARK`) plus the
  first-person engine read from the oracle's live memory. *(the renderer is mapped world-to-pixels;
  the `LEV.ARK` tile map + textures are decoded and the static level geometry is rebuilt as a
  textured 3D mesh (`extract/lev`+`levgeo`+`tex`) shown in the Studio viewer)*
* **Part VI** — audio and cutscenes: the digitized voices (`.VOC`), the sequenced music
  (`.XMI`) and its Miles driver files, and the `CUTS` cutscene animation format. *(planned)*

Methods: purely static analysis of the shipped files, plus dynamic analysis via our own x86
oracle once it exists. All addresses are given as real-mode `segment:offset` pairs or as byte
offsets into a file (called out explicitly); bytes are little-endian.

---

# Part I — The DOS executable and data catalog

## 1. What kind of program this is

`UW.EXE` begins with the ASCII bytes `MZ` — the signature of a DOS **relocatable executable**.
It is a **plain 16-bit real-mode** program: there is no `PE`, `LE`, `LX` or `NE` secondary
header, and no DOS-extender stub (DOS/4GW, Pharlap, DOS/16M) anywhere in the image. UW therefore
runs directly on the 8086-compatible real-mode instruction set (with the 286/386 extensions the
1992-era minimum hardware provided), addressing memory through the classic `segment:offset`
scheme where a linear byte address is `segment × 16 + offset`.

This is the fact that shapes the whole project: **we have no x86 tooling**, and real-mode x86 is
unlike anything already in `tools/` — variable-length instructions (1–6+ bytes, with prefixes),
segmented addressing, and a program that (as shown below) is far larger on disk than the single
image DOS loads, implying a **code-overlay** scheme it manages itself.

## 2. The MZ header

Decoding the 28-byte MZ header at file offset 0 (verified by `extract/cmd/uwinfo`, which
reimplements the parse in `tools/dos/mz.go`):

```
$00  4D 5A            signature        "MZ"
$02  30 01            bytes/last page  = 304
$04  36 03            pages (512B)     = 822        → load image ≈ 421 KiB
$06  68 0C            relocations      = 3176
$08  20 03            header size      = 800 paragraphs = 12,800 bytes
$0A  00 00            min alloc        = 0 paragraphs
$0C  FF FF            max alloc        = 0xFFFF paragraphs (asks for all it can get)
$0E  8B 63            initial SS       = $638B  (relative to load segment)
$10  80 00            initial SP       = $0080
$12  00 00            checksum         = 0 (unused)
$14  C5 0E            initial IP       = $0EC5
$16  3E 00            initial CS       = $003E  (relative to load segment)
$18  3E 00            reloc table off  = $003E
$1A  00 00            overlay number   = 0 (this is the main program)
```

Derived load layout (all file offsets):

| Quantity | Value | Meaning |
|---|---:|---|
| Header | `$0000`–`$3200` | 12,800 bytes; contains the header + the 3,176-entry relocation table |
| Load module | `$3200`–`$66B30` | **407,856 bytes** DOS copies into memory as the program image |
| Appended data | `$66B30`–`$89250` | **141,088 bytes** past what the header calls the load image |
| Entry point | `CS:IP = 0EC5:0000` | module offset `$EC50` |
| Initial stack | `SS:SP = 638B:0080` | a ~408 KiB program image needs a stack segment high in memory |

Two things stand out:

- **3,176 relocations.** Every far pointer baked into the image must be fixed up at load time by
  adding the runtime load segment. The first few point into segment `$0EC5` (the same segment as
  the entry point), i.e. the fix-ups begin right where execution begins. Our `ParseMZ` reads
  the whole table (entries are `{offset u16, segment u16}` at file offset `$3E`); a future x86
  oracle will apply them when it places the module at a chosen load segment.
- **141 KiB of data past the load image.** DOS only loads the 408 KiB module described by the
  page count; the trailing 141 KiB is *not* part of that image. For a 1992 game that had to run
  in ≤ 640 KiB of real-mode memory, this is the classic signature of a **self-managed code
  overlay** region: segments of code/data appended to the EXE that the program pages in on
  demand. Confirming exactly how UW indexes and loads these overlays is a Part III goal (it needs
  the oracle) — for now Part I simply records that the region exists and where it starts.

## 3. The on-disk data catalog

Alongside `UW.EXE`, the `game/` folder carries four data directories. `uwinfo` inventories them
by extension (counts and total bytes are observed facts from the shipped files; the *purpose* of
each format is an inference to be confirmed by decoding in Parts IV–VI, not asserted here):

**`DATA/` (74 files)** — the core game database.

| Ext | Files | Bytes | Notable members / first read |
|---|---:|---:|---|
| `.ARK` | 2 | 848,871 | `LEV.ARK` (levels) and `CNV.ARK` (conversations) — the two big archives |
| `.GR` | 33 | 596,001 | image banks: `OBJECTS.GR`, `WEAPONS.GR`, `HEADS.GR`, `CURSORS.GR`, … |
| `.TR` | 4 | 982,608 | `F16`/`F32`/`W16`/`W64` — floor/wall textures at 16/32/64 px (name-implied) |
| `.BYT` | 9 | 576,000 | 9 × 64,000 = full-screen 320×200 byte-per-pixel images (`MAIN.BYT`, `WIN1.BYT`, …) |
| `.DAT` | 17 | 28,835 | small tables: `OBJECTS.DAT`, `PALS.DAT`, `ALLPALS.DAT`, `SKILLS.DAT`, `LIGHT(S).DAT`, … |
| `.PAK` | 1 | 227,230 | `STRINGS.PAK` — the packed game text |
| `.SYS` | 6 | 10,288 | `FONT*.SYS` — bitmap fonts |
| `.CM`/`.CFG` | 2 | 77 | `WEAPONS.CM`, `UW.CFG` |

**`CRIT/` (65 files)** — creatures. Paired `CRnnPAGE.N00` / `.N01` files (32 of each) plus
`ASSOC.ANM`; the `.N00`/`.N01` split looks like paged animation-frame banks per creature id.

**`CUTS/` (72 files)** — cutscenes, as `CSnnn.N0x`/`.N1x`/`.N2x` sequenced page files (the two
largest, `.N10` at 834 KiB and `.N01` at 1.2 MB total, are the biggest data in the game).

**`SOUND/` (79 files)** — audio. 42 `.VOC` (Creative digitized voice), 24 `.XMI` (sequenced
music), and Miles-style driver/timbre files (`.ADV`, `.AD`, `.MT`) — decoded in Part VI.

None of these formats' internals are documented here yet; Part I's job is only to establish the
catalog and file sizes so the later parts have a target list. Re-run at any time with:

```
cd "Ultima Underworld (PC)/extract" && go run ./cmd/uwinfo -game ../game
```

---

# Part II — The x86 toolchain

Everything past Part I needs to read x86 code. The toolchain mirrors the existing per-CPU
packages (`tools/arm`, `tools/m68k`, `tools/mos6502`, `tools/sm83`) and lives at **`tools/x86`**,
platform-neutral under the shared module.

## 1. The disassembler (`tools/x86`) — done

A table/pattern-driven **16-bit real-mode decoder** returning the repository's common
`Inst{Addr, Len, Mnem, Text, Flow, Target, HasTarget}` shape with the same `Flow` enum
(Seq/Branch/Jump/Call/Return/IndJump/Stop) as the other cores, so the existing recursive-trace
machinery drives it unchanged. It handles the full variable-length x86 encoding:

- **Prefixes** — segment overrides (`ES:`…`GS:`), the `0x66`/`0x67` operand- and address-size
  toggles (default 16-bit; a game reaches for the 386's 32-bit forms via these), `LOCK`, and
  `REP`/`REPNE`.
- **ModR/M + SIB + displacement** — both 16-bit addressing (`[BX+SI]`, `[BP+disp]`, `[disp16]`)
  and, under a `0x67` prefix, 32-bit addressing with the SIB `base+index*scale` forms; a memory
  operand gets an explicit `BYTE`/`WORD`/`DWORD` size keyword when no register fixes its width.
- **The integer instruction set** (8086/186/286 plus the common 386 additions — `MOVZX`/`MOVSX`,
  `SHLD`/`SHRD`, the bit instructions `BT`/`BTS`/`BTR`/`BTC`/`BSF`/`BSR`, `SETcc`, the two- and
  three-operand `IMUL`, near `Jcc`), and the **x87 FPU** escapes `D8`–`DF` (decoded so the
  ModR/M and displacement are always consumed — lengths stay correct even for rare encodings).

Control flow is classified for the tracer: near `Jcc`/`LOOP`/`JCXZ` are branches with a computed
target; `JMP`/`CALL rel` carry a target; a direct far pointer (`CALLF`/`JMPF ptr16:16`) folds its
`seg:off` to a linear `Target`; an **indirect `JMP`** (through a register or memory) ends the
path as `FlowIndJump`, while an **indirect `CALL`** keeps the fall-through with `HasTarget=false`.
Anything unrecognised decodes as a `.byte` with `FlowStop` so a walk stays aligned rather than
mis-decoding the next byte. Syntax is Intel order, upper-case, `$`-hex — matching the other
disassemblers here. Unit-tested against hand-assembled encodings (addressing modes, prefixes,
every control-flow form, x87 length-correctness) in `tools/x86/x86_test.go`.

**Validated on UW.EXE's own code.** Disassembling from the real entry point (`$EC50`, the DOS C
runtime startup) produces clean, coherent x86 — the `INT 21h`/`AH=30h` get-version call, the PSP
field reads at `[$0002]`/`[$002C]`, the environment-scan `REPNE SCASB`/`JCXZ`, and the memory-
sizing arithmetic — with every branch landing on an instruction boundary and no desync:

```
0000EC58  B4 30               MOV AH, $30
0000EC5A  CD 21               INT $21          ; DOS get-version
0000EC5C  8B 2E 02 00         MOV BP, [$0002]  ; PSP: top-of-memory segment
...
0000EC80  B9 FF 7F            MOV CX, $7FFF
0000EC83  FC                  CLD
0000EC84  F2 AE               REPNE SCASB      ; scan the environment
0000EC86  E3 61               JCXZ $0000ECE9
```

## 2. The CLIs — done

- **`cmd/disx86`** — a linear disassembler (`-base`/`-skip`/`-start`/`-end`), paired with
  `dis6502`/`dis68k`/`disarm`. Disassemble the MZ load module by skipping the 12,800-byte header:
  `disx86 -skip 0x3200 -base 0 -start EC50 game/UW.EXE`.
- **`cmd/codetracex86`** — the recursive-descent tracer (`-entry`, `-table` for jump tables,
  `-annotate`, `-o`), following the repo's `codetrace*` convention and feeding the `disasm/`
  annotation store. It separates code from data, names subroutines by their CALL targets, and
  reports unresolved indirect jumps/calls. From the entry it cleanly traces the CRT startup
  (10 routines, 0 stop-hits) and stops at the **6 indirect calls** through which the runtime
  hands off to the game — which is exactly why the next piece is needed.

## 3. The execution-core oracle — done

The tracer's static reach ends where UW's startup dispatches into the game through indirect and
far calls. The answer is the same one the DS titles needed: an **execution core** that *runs*
UW.EXE and we watch. `tools/x86` now carries one (`cpu.go`/`exec.go`/`exec2.go`) — a real-mode
`CPU` over a `Bus`, covering MOV/LEA, the PUSH/POP family, the eight ALU ops in every form plus
INC/DEC/NEG/NOT/TEST, MUL/IMUL/DIV/IDIV, the shift/rotate family, the REP string ops, the full
near/short/far/indirect control transfers, the flag and BCD/ASCII-adjust ops, and INT/IRET, all
with real-mode flags. Unimplemented opcodes (notably the x87 FPU) halt with the offending byte so
gaps are explicit. It is unit-tested with small programs (a `LOOP` sum, `CALL`/`RET`, conditional
branches, `SUB` flags, memory round-trips, `REP STOSB`, `MUL`, and an `IntHook`).

Around it, `tools/dos` (`dos.go`/`dos_int.go`) is a **reusable minimal real-mode DOS**: it loads UW.EXE the
way DOS would — place a PSP and environment, copy the load module, **apply the 3,176 relocations**
(this is what turns the entry's link-time `MOV DX,$5C0F` into the correct runtime data segment),
seed `CS:IP`/`SS:SP`/`DS`/`ES` — then services the `INT 21h`/BIOS calls the program makes
(get-version, the memory-block calls, interrupt-vector install, DTA, and **real file
open/read/seek/close against the game directory**, plus termination). `extract/cmd/bootoracle`
drives it and reports where the boot goes.

**Result: the oracle boots the real game.** It runs ~10,900 instructions of the Microsoft C
run-time startup with no wrong turn — servicing get-version, resizing its memory block three times,
installing the divide/overflow/etc. interrupt vectors — and then **reaches the overlay manager**,
exactly the wall Part II §2 predicted. This both validates the CPU core on a large body of real
code and lands us precisely at the Part III entry point.

## 4. What the oracle found: the overlay system (Part III lead-in)

The boot stops at a `CALLF` into a freshly-loaded segment whose target is `INT 3Fh` — the
signature of the **Microsoft C run-time overlay manager**. The oracle's file-I/O log and a memory
dump pin the whole mechanism down:

- The manager **reopens UW.EXE** and seeks to file offset **`$66B30`** — precisely the
  `LoadImageEnd` (start of the 141 KiB appended region) computed in Part I — reads a 16-byte
  overlay-table record there, then loads a **7,616-byte overlay** from `$66B40` into a fresh
  segment. Part I's "appended data past the load image" is therefore confirmed to be the
  **overlay store**, and its directory begins at `$66B30`.
- Inter-overlay calls go through a table of **5-byte thunks**, each `CD 3F | DW entry_offset |
  DB overlay_number` (an `INT 3Fh` followed by the target offset and overlay index). The dump of
  one such table:

  ```
  5C4B:0034  CD 3F 3C 06 00   INT 3F  entry $063C  ovl 0
  5C4B:0039  CD 3F B1 02 00   INT 3F  entry $02B1  ovl 0
  5C4B:0043  CD 3F 00 00 00   INT 3F  entry $0000  ovl 0   ← the call the boot reaches
  ```

Implementing the `INT 3Fh` handler (parse the overlay directory at `$66B30`, load an overlay on
demand, transfer to `overlaySeg:entry`) is **Part III** proper — it needs the overlay-table format
reverse-engineered, which the oracle now makes straightforward. The static asset formats (Part IV)
don't need any of this and can proceed in parallel.

---

# Part III — Boot chain and overlay system

The oracle now boots UW **all the way through initialisation** — the overlay system works and the
game loads its entire asset set. The key decision was *not* to reimplement the overlay manager but
to **run the game's own handler**.

## 1. The overlay loader

The CRT installs the `INT 3Fh` handler in the IVT at boot (`IVT[0x3F] → 3AB0:04E4`); disassembling
it shows a textbook Microsoft-C overlay dispatcher — it reads the 5-byte thunk off the return
address, indexes a resident overlay directory in its own data segment (`DS = 5A4C`), and pages the
overlay in from UW.EXE with DOS file calls. Since the oracle already services those file calls, the
handler just *works*: the machine stops special-casing `INT 3Fh` and lets the CPU dispatch through
the real vector. The game then loads overlays on demand from the store at file `$66B30` (12+
overlay calls during boot, overlays 0 and 32 among them), reopening UW.EXE and reading each
overlay's code into a fresh segment — exactly as it would on a real DOS machine.

## 2. What the boot needs — a minimal DOS/PC around the CPU

Running the real engine surfaced each thing UW depends on, and the oracle grew a faithful-enough
model of each (all now in the reusable `tools/dos`):

- **LIM EMS (expanded memory), `INT 67h`** — UW *requires* an EMS driver (without one it prints
  "Out of EMS Memory" and quits). `dos_ems.go` provides a 64 KiB page frame + an 8 MiB expanded
  pool with the standard functions (allocate/map/save-restore), plus the `EMMXXXX0` driver
  signature the game probes for. Mapping is copy-on-remap over the flat address space.
- **DOS file management** — beyond open/read/seek/close: **FindFirst/FindNext** (search state
  keyed by the DTA address, as real DOS keeps it, so interleaved searches don't clobber one
  another; host dotfiles like `.DS_Store` are hidden), get/set attributes, and a **writable
  scratch overlay** — reads fall through to the read-only game folder, writes and created files
  land in a temp directory, so nothing is ever written into the (copyrighted) game data. The
  overlay also supplies the **`SAVE0` working directory** a real install's SETUP would have
  created (the game aborts without it), seeded from `DATA`.
- **PC I/O ports** (`dos_io.go`) — the 8042 keyboard controller handshake, the VGA retrace bit
  (`0x3DA`), the PIT, and enough **Sound Blaster Pro** detection (OPL FM timer, mixer read-back,
  DSP reset/version) that the `sbpfm.adv` driver's initialisation succeeds instead of aborting.
- **BIOS initialisation** — `setupBIOS` fills the interrupt-vector table with harmless `IRET`
  stubs and populates the BIOS Data Area, exactly as a real BIOS would. This matters: UW's timer
  ISR *saves and chains* the previous `INT 8` handler (`PUSHF; CALLF [saved vector]`), and with a
  zero-initialised IVT that chain far-jumped through a null pointer into unmapped memory. Stubbing
  the vectors makes the chain land on a valid instruction.
- **Timer interrupts** — the CPU core grows `Interrupt(n)` (maskable IRQ delivery honouring `IF`
  and a `MOV SS` interrupt-shadow), and the machine can inject a periodic IRQ0.

With all of this, the oracle runs **~1 million instructions** of real game code and loads UW's
entire front end: `STRINGS.PAK`, `UW.CFG`, the fonts, `ALLPALS.DAT`, the presentation screen, the
sound driver + `SOUNDS.DAT`, and the whole `.GR` graphics set (`OBJECTS`, `VIEWS`, `CURSORS`,
`INV`, `SPELLS`, `DRAGONS`, `COMPASS`, … — it even seeks through `OPTB.GR` reading individual
sub-images). This is the payoff for Part IV: the oracle can be watched opening and parsing each
asset file, which is exactly how their formats will be decoded.

## 3. Hunting the boot abort (the investigation)

The earlier crash is fixed — the BIOS IVT stops the timer ISR chaining through a null vector — and
the boot now runs cleanly to a **deterministic** stop (the same on `-irq` and default: the timer
injection is *not* involved). The game loads every asset, then aborts with `exit(232)`.

That abort was traced, with the oracle's breakpoint/ring/disassembly/memory-watch tooling, all the
way to its **root cause** — a DOS memory-management fidelity gap. The chain, from symptom inward:

- The internal error code is `0x3004` (`"D004"`, "Could not read data"), raised inside a
  **resource-verification framework** in overlay `66CF`: it walks several resource databases and
  ANDs a success flag; the first failing resource is item `$14AA`.
- Its callback allocates a per-record buffer from a **heap allocator** (`015E:004A`) over a block
  table at segment `5751`. The allocator returns 0 — no free block is big enough.
- The heap is empty because it was **sized to zero**: its init reads a descriptor at `3BDD:4106`/
  `3BDD:4108` (`heapSize = [4106] - [4108] - 1`), and those hold `0xFFFF`/`0xFFFE`.
- A memory-write watch pins where `0xFFFF` comes from: `01A0:3A61` stores `DI`, and `DI` is the
  result of a scan (`01A0:3A3F`…) over the game's **internal memory-allocation registry** — a
  min-of-tracked-blocks that *defaults to `0xFFFF` when the registry is empty*. In our emulation
  the registry **is** empty.

So the game manages memory by repeatedly **resizing its own program block** (`INT 21h/4Ah`, growing
to ~35,200 paragraphs) plus a 928 KiB EMS allocation, then sub-allocating internally. `tools/dos`
now carries a **faithful DOS memory manager** (`dos_mem.go`) — a real **MCB chain** with correct
`INT 21h/48h/49h/4Ah/52h` block, free, resize (grow-by-absorbing-free-neighbours) and coalesce
semantics, replacing the earlier bump stub. That is the right foundation and matters for the deeper
runs to come.

It did **not**, however, move the abort: with the faithful manager the run is *byte-identical*, so
the exit is confirmed **not** a DOS-memory-call issue — UW barely touches DOS memory beyond the
program-block resize, which the stub already satisfied. The trouble is one level further in, in the
game's *own* sub-allocator: `015E:000A` reserves a heap from a pool (bounds `[3BDD:4106]`/`[4108]`),
writes the heap header, but the heap's block table ends up without a usable free entry (the
allocator `015E:004A` reads an entry whose "free" field is 0 and skips it), so every record
allocation fails. The game's memory-manager data structures are simply ending up **inconsistent**
under our emulation.

That inconsistency looked like the signature of a **subtle CPU-core bug**, so the core was
cross-validated against an independent reference — the **SingleStepTests/8088** suite, which gives a
full initial CPU+RAM state, one instruction, and the exact resulting state, for thousands of cases
per opcode. A differential harness (`tools/x86/harte_test.go`, gated on `HARTE_DIR`) loads each
case, `Step()`s once, and diffs registers/memory/flags (masking flags the 8088 leaves undefined,
and skipping opcodes that genuinely differ between the 8088 and this core's 286+ choices).

That paid off: it found **one real bug** — a word/dword memory access straddling offset `0xFFFF`
wrapped the *linear* address instead of the *segment offset* (8086 wrap), so the high byte landed a
paragraph past the segment instead of at its base — now fixed and regression-tested. But across
~155 opcodes (every instruction UW's failing path uses) the core otherwise matches the real 8088
exactly. And an instruction counter shows UW executes **zero** 386 instructions before the abort —
its whole init is 8088-level code. So the exit was **not** a CPU instruction bug; that hypothesis
was disproven, redirecting the hunt to the machine model — where the real causes were found (§4).

## 4. The real causes — and the oracle running the game

With the CPU exonerated, the remaining divergences were tracked down one by one with the oracle's
debugger (`-watch` on the game's own variables was the key tool). Four fixes, in causal order:

1. **The video BIOS (`INT 10h`) — the root cause of the boot abort.** The exit-232 chain traced
   back through the game's sub-allocator to its **video-memory arena** (`[3BDD:4104/4106/4108]`)
   never being initialised, because the arena init lives in the **video-mode-set routine** — and
   the game only sets its mode after *detecting a VGA through the BIOS* (`INT 10h AX=1A00`),
   which our stub never answered. `dos_video.go` now implements a minimal VGA BIOS (mode set/get
   tracked in the BDA, the `1A00`/`12,10`/font detection answers, palette/DAC functions), and the
   game immediately detects VGA at `01A0:3400` and sets **mode 13h**. The arena initialises, the
   D004 abort vanishes, and the game runs its intro.
2. **Signed `lseek`.** DOS `INT 21h/42h` takes a *signed* CX:DX offset; treating `-4` as
   `+0xFFFFFFFC` sent the C runtime's buffered reads to nowhere.
3. **EMS 4.0 map-multiple (`INT 67h/50h`)** — the game maps batches of pages once running; added
   to the EMS model (it streams whole screens into the page frame, e.g. `OPSCR.BYT` → `E000:0`).
4. **DOS handle reuse — the subtlest one.** The Microsoft C runtime indexes its internal per-file
   flags table *by DOS handle number*, sized for DOS's small per-process handle table. Our DOS
   handed out handles monotonically (never reusing), so after ~180 opens the C runtime indexed
   past its table and `fread` failed **without issuing a single DOS read** ("D015"). Real DOS
   always returns the lowest free handle; `allocFH` now does the same, and handles stay below 10
   for the whole run.

**The result: the oracle runs the game.** It boots, detects VGA, plays the animated intro
cutscene, initialises a **new game** (creating `SAVE0\LEV.ARK` and `SAVE0\BGLOBALS.DAT` in the
scratch overlay — matching the empty `SAVE0` a fresh install ships with), loads the menu assets,
and settles at its **interactive main menu** — 700 M+ instructions, deterministic, no
terminations.

## 5. Planar VGA — pixel-perfect screenshots

The first framebuffer dumps showed each image repeated four times across the screen, with correct
colours. The cause was settled by evidence, not inference: the machine now models **true VGA
memory** (`dos_vga.go`) — four 64 KiB planes behind the `A000` window, plus the sequencer,
graphics-controller and CRTC registers that steer it (word `OUT`s to the index ports decoded as
index+data pairs). The register log then proves what the game does: right after setting mode 13h
it writes `seq[4]=F7` at `01A0:342E` — **turning chain-4 off**, the classic unchained **Mode X**
setup. From then on every byte write reaches only the planes selected by the map mask; a flat
framebuffer keeps one pixel in four (colours intact, since each plane byte is a full 8-bit pixel —
which is why the repeats looked correctly coloured), and reading it linearly interleaves rows four
to a line: the "four images".

The model implements chain-4 addressing, unchained write modes 0 (set/reset + bit mask), 1 (latch
copies, the VRAM-to-VRAM blit mode) and 2, and read modes 0/1 with latch loading. `Screenshot`
reconstructs the display the way the CRT does — plane `x&3` at offset `start + y·pitch + x/4`,
start address and pitch from the CRTC registers — so `bootoracle -shot` now yields **single,
pixel-perfect frames**: `rendered/oracle-planar-*.png` show the intro cutscene (a chainmail
knight, subtitle *"A score of us gave chase, but it fled into the Stygian Abyss with poor
Arial."*) and the **main menu** (*Introduction / Create Character / Acknowledgements*), exactly as
a monitor would show them. Beyond proof, this matters for what comes next: the plane-accurate
framebuffer is the reference for debugging the game's own rendering algorithms in Part V.

The dependable-oracle goal is met: `bootoracle -irq` runs the real game deterministically into
its interactive main menu, with faithful video output.

## 6. Input injection — driving the game into the dungeon

An oracle that only watches is half a tool; to reach the dungeon and study object behaviour it has
to *play*. UW takes its input two ways, and injection meets each at its own layer (`dos_keyboard.go`,
`dos_mouse.go`):

- **Keyboard.** UW installs its own `INT 9` (IRQ1) handler at `214A:062E`: it does `IN AL,60h`,
  stores the raw scancode into a 64-byte ring, acknowledges the keyboard (port 61h) and PIC
  (`OUT 20h,20h`), then IRETs. Injection puts a scancode in the 8042 output latch and raises IRQ1 —
  the game's *own* ISR consumes it as a real key. A key is a make code then a break code
  (make | 80h). The delivery is gated on the interrupt-enable flag, so it lands only when a real
  IRQ1 would; a subtlety fell out of this: the timer IRQ's dispatch clears `IF` for the duration of
  its handler, so keyboard injection rides a **half-tick phase offset** from the timer to find the
  game in its normal, interruptible flow.
- **Mouse.** UW's pointer UI polls the `INT 33h` driver — function 0Bh for the relative motion
  counters and 03h for position and buttons (it installs no callback). A Microsoft-compatible driver
  (`dos_mouse.go`) answers both. Injection drives the cursor by feeding motion deltas; the driver's
  scale was pinned by reading UW's own code, which computes `delta × 100 / 200` — a halving — so two
  mickeys per pixel gives a 1:1 mapping, and each `at:X,Y` first *homes* the cursor into the corner
  with a large negative delta (giving a known origin) before moving. A click is a short
  press→release pulse polled within one UI frame; a long hold reads as a drag and does nothing.

`bootoracle -keys` compiles a comma-separated script — key names, `wait:N`, `at:X,Y`,
`lclick`/`rclick` — into a schedule paced against the timer tick. That was enough to walk the whole
of character creation with mouse clicks (sex, handedness, class, skills, a weapon, a general skill,
a portrait, difficulty), type the hero's name at the keyboard, confirm, and press through — every
step verified against UW's own crosshair drawn into the framebuffer (`rendered/cc-*.png`).

**The divide-error gate.** Confirming the character launches the intro and world setup, where the
core stopped on a *divide overflow* at `07F7:4DCA` — `IDIV DI` computing `([1612] << 15) / DI` with
`[1612] == DI`, i.e. a quotient of exactly 32768, one past the signed-16 maximum. This is not a
bug: UW's fixed-point renderer installs its **own #DE handler** (IVT[0] → `589E:04D0` →
`07F7:5BD1`, inside the same overlay) and deliberately lets such divides overflow so the handler can
*saturate* the result — the classic texture-mapper trick. The core was halting instead of vectoring
the exception, so `divOp` now raises `#DE` through IVT[0] (`divErr`), exactly as an 8086 does. With
that, the game runs its own handler and drops into the **dungeon**: `rendered/dungeon.png` is the
first-person, texture-mapped view of the Stygian Abyss — a stone corridor drawn by UW's engine on
our CPU, with the character HUD beside it. (The peripheral status panels still show some noise; the
central 3D viewport is correct. Pixel-exact UI is a Part V refinement.)

This is the whole arc closed: from an MZ header we could not run to the game *being played* under
the oracle, deterministically, all the way to dungeon gameplay — the foundation for the
object-behaviour and world work of Parts IV/V.

### Oracle tooling

`bootoracle` now carries a small debugger useful for exactly this kind of trace: `-bp SEG:OFF`
(with `-bpal N` to gate on `AL`) breakpoints, `-watch SEG:OFF[:LEN]` to log writes to an address or
range (how the bad heap descriptor was pinned), `-dis SEG:OFF:LEN` to disassemble live *relocated*
memory (essential when overlay far-call targets are fixed up at load), `-dump` hex-dumps, `-shot` to
render the planar VGA framebuffer to PNG, `-keys` to script keyboard/mouse input, a decoded
instruction **ring buffer** printed on stop, and spin/runaway detectors.

# Part IV — Asset formats (planned)

Static decode, no oracle required: the palettes (`PALS.DAT`/`ALLPALS.DAT`), the `.GR` image
banks, the `.TR` textures, the `.BYT` full-screen images, the `FONT*.SYS` fonts, and
`STRINGS.PAK`. Each will be reimplemented in the game's `extract` module and rendered to `rendered/`.

# Part V — The world and the 3D renderer

`LEV.ARK` — the dungeon: the per-level tile map, the object placement lists, and the geometry the
first-person engine renders. `CNV.ARK` — the conversation scripts. *(the level archive is still
planned static decode)*

## 1. Reverse-engineering the 3D renderer — the camera transform *(started)*

Now that the oracle plays into the dungeon, the renderer can be read from **live memory** — it is
overlay-paged code, so static tracing of the raw `UW.EXE` can't reach it, but `bootoracle -dis`
disassembles the relocated, paged-in image directly, and a breakpoint on the perspective divide
catches it mid-render. The renderer's code overlay is segment **`07F7`** and its data segment is
**`499D`**; the disassembly and full commentary are in `disasm/uw-render.asm` +
`disasm/uw-render.annotations.txt`. What's mapped so far:

- **The view matrix** at `[499D:1600..1618]` — the camera orientation basis in 1.15 fixed point
  (`0x8000` = 1.0). A live capture at dungeon entry: `[1602]=0x64F2`, `[160A]=0x7FFD`≈1.0,
  `[1612]=0x61A4`, `[1616]=0x4D03`, the rest zero (sparse because the player spawns facing a
  cardinal direction).
- **Basis orthonormalisation** (`07F7`, from `4B90`): it scales the matrix rows by reciprocal
  lengths, then normalises each basis vector — length via an integer `isqrt` (`CALLF 214A:0A30`,
  Newton's method), then `component = (c<<15) / length`, saturated to ≈1.0.
- **Shared math** in segment `214A`: `0A30` = `isqrt32(CX:BX)`; `0A34` = a linear interpolation of
  two tables (use not yet confirmed — left unclaimed rather than guessed).

## 2. A divide-by-design — and a CPU-generation bug it exposed

The renderer never guards its normalisation divides. When a basis vector is axis-aligned the
normalised component is exactly 1.0, i.e. quotient `32768` — one past the signed-16 maximum — so
the `IDIV` overflows. Instead of testing for that every time, the renderer **arms its own
divide-error handler** (`MOV [SS:04D5],$5BD1` at `07F7:4D85`, so `IVT[0] = 589E:04D0 → JMPF → 07F7:5BD1`)
and lets the `#DE` fire; the handler skips the faulting `IDIV` and returns the saturated `0x7FFF`.
It is the classic fixed-point-renderer trick, done through the interrupt vector.

Reading that handler pinned a real oracle bug. It advances the pushed return IP by 2 to step over
the 2-byte `IDIV` — which only lands correctly if the pushed IP is the **faulting** instruction's,
the **286-and-later** `#DE` behaviour. `tools/x86` was pushing the *following* instruction's
address (the 8086/8088 quirk), so the handler returned one byte into the next instruction and
transiently corrupted the view matrix at setup. Since this core models a 386-class machine (UW's
target), it now pushes the faulting address (`divErr` + `CPU.instrIP` in `tools/x86/exec.go`); the
8088 differential harness skips the divide-*exception* cases as the documented generational
difference (the same class as the already-skipped `PUSHF` reserved bits). The `SingleStepTests/8088`
suite still passes.

## 3. The rasterisation pipeline

Where the view transform (§1) was found by breakpointing the perspective divide, the *pixel* code
was found by **profiling writes** — a new oracle feature (`bootoracle -vgaprof N` tallies the code
addresses that write the `A000` framebuffer once past instruction *N*; `-profrange SEG:OFF:LEN`
retargets the tally at any buffer). In the dungeon the framebuffer writes all come from one place —
segment `01A0`, offsets `0BED..0C95`, ~40 unrolled addresses each writing an identical count: an
unrolled **blit**, not the rasteriser. It reads from an off-screen buffer at segment **`41C5`**, so
that is where the scene is actually drawn. Re-profiling *that* buffer gives the pipeline:

    3D overlay 07F7  ─draws→  chunky off-screen buffer @ 41C5 (1 byte/pixel)
    01A0 primitives  ─draws→  (same buffer)
    01A0:0B96 blit   ─copies→ A000  (deinterleaved into the 4 Mode X planes)

- **The frame is drawn chunky, then blitted.** Everything renders into a linear 1-byte-per-pixel
  buffer at `41C5`; nothing touches planar VGA until `01A0:0B96` — the **chunky→planar Mode X blit**
  — copies it out in four passes (one per plane, setting the sequencer map-mask, an unrolled
  `MOVSB; ADD SI,3` deinterleave, 80 bytes per scanline). This is exactly why the naïve pre-planar
  screenshots showed four copies.
- **The texture mapper** is `01A0:02CE`: it walks a span copying one texel per pixel while stepping
  a fixed-point texture coordinate whose gradient is **self-modifying code** (the `ADD BP,imm /
  ADC AX,imm` immediates are patched per span). Textures are 32 texels wide (`SI = (V&0x3E0)+U`).
  Alongside it are a flat span fill (`01A0:0A58`, `REP STOSW`) and a `LODSW`-driven display-list
  interpreter (`01A0:0A5C`) that dispatches the primitives through a jump table.

So `01A0` is the resident **software rasteriser** (texture spans, fills, blit); `07F7` is the 3D
geometry overlay that drives it (and writes the buffer directly too); a further overlay `1FF9`
contributes buffer writes not yet mapped.

## 4. The two-level affine DDA

The texture span is the innermost of a **three-level affine rasteriser** — found by profiling the
writers of the span's inputs, then the writers of *their* deltas. There is no per-pixel divide; the
only divides are two per-span gradient `IDIV`s:

    01A0:0312  vertical DDA   — per scanline: step the left/right edge X and the span-endpoint
                               texture coords by constant per-polygon deltas; loop to the last row
      ↓ per scanline
    01A0:0296  horizontal DDA — CX = span length; IDIV out the U and V gradients (endU−startU)/CX,
                               (endV−startV)/CX and patch them into the span's step immediates
      ↓ per pixel
    01A0:02CE  texture span   — MOVSB a texel, step the coordinate, LOOP

Every level's per-step delta is a **self-modifying immediate** the level above patches — the edge
slopes into `01A0:0312`, the span gradients into `01A0:02CE`. So drawing a textured wall is:
project the vertices → compute the constant edge/texture slopes and patch them → step down the
polygon → patch and run each span. The state (edge X, texture endpoints, scanline counter) lives in
the graphics-driver data segment `3BDD` at `[07B7..07CF]`. Full disassembly in `disasm/uw-render.asm`
(PART 2 = span + blit, PART 3 = the two DDAs) with commentary in `uw-render.annotations.txt` (§5–6).

## 5. The perspective projection — the geometry heart

Climbing once more (profile who writes the *screen vertices* the polygon setup reads) lands in the
`07F7` overlay at `6148`: the **perspective projection**. It turns each view-space point (already
rotated out of world space by the camera basis, §1) into a screen vertex, with one divide-by-Z per
axis — the textbook pinhole projection:

    screenX = X · scaleX / Z + centreX          screenY = Y · scaleY / Z + centreY

The constants, read live at `499D:26B0`, are `scaleX = centreX = 86` and `scaleY = centreY = 56`, so
the view spans `[0,172] × [0,112]` centred at `(86,56)` — exactly the dungeon 3D viewport. Texture
coordinates are copied straight through; Z is stashed in the vertex for the rasteriser. And like the
basis normalisation (§1), the projection **arms its own `#DE` handler** (`[SS:04D5] = 07F7:692C`) so
a point at or behind the eye (`Z ≈ 0`) saturates instead of trapping — the same interrupt-vector
trick, a second handler for the projection's divides.

That closes the loop. The whole first-person renderer, world to pixels:

    world point
      → camera basis  (§1, orthonormalised view matrix @ 499D:1600)  → view-space point
      → perspective projection  (07F7:6148: screen = axis·scale/Z + centre)  → screen vertex
      → polygon edge setup  (01A0:0060: edge & texture slopes via IDIV, patched)
      → vertical edge DDA  (§4, 01A0:0312)  → horizontal span DDA  (§4, 01A0:0296)
      → texture span  (§3, 01A0:02CE: texels into the 41C5 chunky buffer)
      → chunky→planar Mode X blit  (§3, 01A0:0B96)  → A000  → pixels

Full disassembly in `disasm/uw-render.asm` (PART 4 = the projection) with commentary in
`uw-render.annotations.txt` (§7).

## 6. The level archive and the tile map

The renderer's geometry ultimately comes from `LEV.ARK`. Its outer format falls straight out of the
file: a `uint16` block count (135) followed by that many `uint32` block offsets, each block running
to the next offset. The first **nine** blocks are exactly **31,752 bytes** — the nine dungeon
levels; the rest are per-level texture-mapping, automap and animation blocks. The load trace
confirms it: the game copies `DATA\LEV.ARK` to `SAVE0\LEV.ARK`, reads the 2-byte count and the
540-byte offset table, then `seek $21E; read 31752 bytes` pulls block 0 (level 1) into memory at
`798D:0004`.

Inside a level block, the first `64×64×4` bytes are the **tile map** — two little-endian words per
tile. Every field was derived, not looked up — the geometry/object fields from the game's *own*
decode code (found with the new read profiler, `-rdprof`, pointed at the loaded tile map), and the
texture split from the data distribution then confirmed by rendering:

| Field | Bits | How it was derived |
|---|---|---|
| tile **type** | word0 `0-3` | `2CD3:0740` — `MOV AX,[tile]; AND AX,000F`, indexes a per-type geometry table |
| floor **height** | word0 `4-7` | `2CD3:0835` — `SHR AX,4; AND AX,000F` |
| **floor texture** | word0 `10-15` | data: clusters to a small per-level set; ~0 on solid tiles |
| **wall texture** | word1 `0-5` | data: clusters to a per-level set; present on solid *and* open tiles |
| first **object** | word1 `6-15` | `28B3:08F9` — `MOV AX,[tile+2]; SHR 6; AND 03FF` (0 = empty) |

**Geometry.** Type `0` is solid rock (full walls, no floor), `1` open flat floor, `2-5` the four
diagonal (half-solid) corners, `6-9` floors sloping up toward N/S/E/W. That the diagonals are one
shape in four orientations — and likewise the slopes — is proven by a **rotation-remap table** in
DGROUP at `034E`: its first row is the identity `00 01 02 03 04 05 06 07 08 09`, and rows 1-3 rotate
`2↔4↔5↔3` and `6↔9↔7↔8`. The game indexes it by the camera facing (`[2B1A]`, 0-3) so its
floor-height query (`2CD3:0720`) only handles a canonical orientation; the slope surface itself is a
per-subregion height table at `038F`.

**Textures** are per-level 6-bit indices into the level's texture list (a separate block, not yet
decoded). Reimplemented in Go (`extract/lev`, `cmd/levinfo`) and **verified by rendering twice**:
the tile-type map (`rendered/level1-map.png`) reproduces a coherent Level 1 with the unmistakable
**ankh room**, and colouring each tile by its texture index (`rendered/level1-tex.png`) yields
coherent regions — rooms share a wall colour, and the ankh is drawn in its own floor texture,
exactly as the game renders it. Both are the proof the fields are right. Each of the 329 non-empty
tiles also carries an object-list head, the bridge into the object/item system (still to decode).

## 7. How tiles become polygons — and how textures are referenced

Tracing the geometry-prep pass joins the level data (§6) to the renderer (§1–5). The visible
geometry is a **display list** at `499D:744A`, **rebuilt every frame** by segment `1FF9` — the
tile→polygon tessellator (which also reads the tile map). It is a command stream the `07F7` side
interprets: each word dispatches (`word/2`) through a ~24-entry jump table at `499D:2738`, the live
handlers being at `07F7:2B16 / 5E04 / 79CE`. For each vertex, `07F7:5096` reads it as **tile-space
bytes** — `(tileX, tileY, height)` — scales each by 32 (`SHL 5`, the world units per tile),
subtracts the camera position, and multiplies by the camera matrix (§1) into a view-space vertex
pool at `499D:1620` (8 bytes/vertex). A per-polygon gather (`07F7:5E2A`) then copies a polygon's
vertices out of the pool by index, attaches texture coordinates, and feeds the projection (§5). So a
level vertex is stored in *tile coordinates* and transformed on the fly.

The builder itself (`1FF9`) reads each tile's fields directly — `byte0>>4 & 0F` = floor height,
`byte0 & 0F` = type, `byte1>>2 & 0F` = a texture field (the very fields `extract/lev` decodes,
confirmed here from the builder's own code) — and branches on them to emit the tile's polygons into
the list through a write pointer at `[7250]`, one command block per polygon.
(An earlier draft of this section, before the builder was pinned, wrongly called the list *cached*
and named an interpreter at `07F7:2BF0`; that breakpoint never fires — the `065D` it seemed to use
was a stale main-loop `SI`. The list is `499D:744A`, rebuilt each frame, and needs no player
movement to observe.)

**The texture reference** falls out at the rasteriser (`01A0:0210`): each polygon descriptor carries
its texture *inline* — the draw setup reads the texture **segment** straight into `[07AF]`, the very
pointer the texture span (§3) samples, plus a texture parameter it patches into the span. That
segment is the loaded `.TR` bitmap the display-list builder chose from the tile's texture index
(§6). So the texture travels *tile index → per-level list → `.TR` segment → polygon descriptor →
`[07AF]` → span*, while the coordinates travel *projection → gather → affine DDA*.

That is "how a tile becomes polygons," from actual code: `1FF9` walks the visible tiles each frame
and emits, per tile, floor/wall/diagonal quads whose corners are tile-space vertices and whose
texture is its floor/wall index; `07F7` transforms and projects them and `01A0` rasterises them.
Full disassembly in `disasm/uw-render.asm` (PART 5 = the 07F7 interpret/transform + texture ref,
PART 6 = the 1FF9 builder) with commentary in `uw-render.annotations.txt` (§8–9).

## 8. Exporting the static level geometry, with textures, into the viewer

With the tile format (§6), the geometry rules (§7) and the world scale all derived, the level can be
rebuilt as a **textured 3D mesh** — reimplemented in Go and hooked into the Studio's three.js viewer.

- **Textures.** The `.TR` banks decode straight from their bytes (`extract/tex`): byte 1 is the
  square dimension (`W64.TR` = 64×64 walls, `F32.TR` = 32×32 floors), a `uint16` count, then per-
  texture offsets to `dim×dim` palette indices; `PALS.DAT` holds eight VGA palettes. The **per-level
  texture list** — the missing §6 piece — turned out to be the 122-byte LEV.ARK block `18 + level`:
  **48 wall texture numbers (into `W64.TR`) then 10 floor numbers (into `F32.TR`)**, which a tile's
  `WallTex`/`FloorTex` index selects.
- **Geometry.** `extract/levgeo` walks the tile map and emits, per non-solid tile, a floor quad (its
  corners carrying the sloped heights), a **ceiling** quad at the level ceiling height (on by
  default — an enclosed dungeon), and a wall quad on each edge where the neighbour is solid
  (full height to the ceiling) or higher (a step up) — floors textured `F32`, walls `W64`, through
  the texture list. The wall's top is sampled at **both shared corners** of the edge (not one), so a
  ramp meets a flush neighbour with no wall and produces a *triangular* side wall instead of a
  spurious vertical segment. Wall textures use a **uniform texel scale** (`WallTexUnitsPerCopy` — one
  copy per tile width horizontally and per tile-width vertically), so a tall wall *tiles* the texture
  rather than stretching one copy floor-to-ceiling, and the UVs are oriented upright and un-mirrored
  (V=0 at the foot, U reading left-to-right for a viewer on the open side) — verified with
  `levrender -uvtest` (colour-by-UV) and a first-person textured render of the game's arched door.
  **Diagonal tiles (types 2-5)** are emitted exactly: the solid corner
  (NW/NE/SW/SE, derived from neighbour solidity in the real levels) is cut off, leaving a *triangular*
  floor, a diagonal wall across the hypotenuse, and normal walls on only the two open edges. A
  diagonal is *solid rock along the two edges bordering its solid corner*, so a neighbouring tile
  facing one of those closed edges also gets a wall — this filled the gaps in the ankh room, where
  the loop's diagonal tiles meet the crossbar.
  **Heights** are scaled by the ratio the game's own vertex path proves: the display-list builder
  `1FF9:0006` stores tile X/Y with `SHL 3` (×8) and the height field with `SHL 1` (×2), then the
  vertex transform `07F7:5096` scales all three equally (`SHL 5` then `SHL 1` = ×64). So a tile spans
  `8×64` and a height-field unit `2×64`, i.e. a height unit is `2/8 = ` **a quarter** of a tile width
  (`HeightScale = 0.25`). (An earlier `0.5` was wrong — it read the shared transform but missed the
  unequal `×8`/`×2` in `1FF9:0006`; the symptom was walls twice too tall, doubling the door texture.)
  `cmd/levexport` groups the mesh by material and writes a self-contained JSON
  (positions, UVs, groups, and each texture as a data-URI PNG); `cmd/levrender` is a Go software
  renderer that verified the result is a coherent dungeon before any browser was involved
  (`rendered/level1-3d.png` — Level 1 with its rooms, water channels and the ankh room, in 3D).
- **Viewer.** `site/src/uw/viewer.js` loads that JSON into a three.js `BufferGeometry` with one
  textured material per group (nearest-filtered) and a fly-camera; all eight levels are
  registered in the Studio under a new **MS-DOS** system. The game world is `(X=east, Y=north, Z=up)`;
  the export maps it to three.js Y-up as `(tileX, height, -tileY)` — the `-tileY` (not a plain
  `tileY` swap) keeps the handedness, since a bare axis swap reflects the level and renders it
  mirrored (a step that is on the left in the game appeared on the right). Since the dungeon is now ceiling-enclosed,
  the export also carries a **spawn point** (the open tile nearest the map centre, at eye height) and
  the viewer starts the camera there, inside the level, rather than on an outside overview.

This is a faithful *reimplementation* grounded in the reverse-engineered format, not a byte-exact
replay of `1FF9`: the wall-adjacency rule and slope corners follow the derived tile semantics, the
diagonal cut and per-edge walls follow the derived diagonal orientations, and the height scale is the
transform-proven 0.5 (a height unit = half a tile). The layout, heights, diagonals and textures are
all the game's own; the ceiling texture (reusing the floor texture, for now) is the one piece not yet
read from the game.

## 9. Still open

Inside `1FF9`, the object/sprite geometry it also feeds (the static per-type tile polygons —
diagonals and true heights — are now emitted); the exact per-height display units `1FF9` scales the
0-15 field into (the 0.5 X:Z ratio is proven, the absolute per-height factor is not yet byte-exact).
The object lists the tile heads point into. The render *entry* is a callback of the `2252:0410` IRQ0 timer table; the
peripheral-panel noise in `rendered/dungeon.png` should resolve as the HUD draw path (a separate
`01A0` consumer) is mapped.

One capability gap worth recording: **avatar movement** can't yet be injected. The scripted
mouse-hold does reach the game — during play it reads `buttons=1` at the injected cursor via
`INT 33h` `AH=03h` — but the avatar doesn't move, because UW's movement handler (behind a
stack-switching input thunk at `214A:09F0`) keys off its *own* delta-integrated cursor rather than
the absolute position we set. Solving it would enable full playthrough automation; it turned out not
to be needed for the geometry work, since the display list rebuilds every frame regardless.

# Part VI — Audio and cutscenes (planned)

`.VOC` digitized voice, `.XMI` (XMIDI) music and its Miles driver files, and the `CUTS`
page-file cutscene format.

---

## Appendix A — Tools

- `tools/x86` — the shared 16-bit real-mode x86 disassembler (`Decode`/`Disassemble`) **and**
  execution core (`CPU`/`Bus`/`Step`).
- `tools/cmd/disx86` — linear disassembler; `tools/cmd/codetracex86` — recursive code-tracer.
- `tools/dos` — the **reusable real-mode DOS/PC machine** (the oracle) built on `tools/x86`:
  `mz.go` (MZ loader + relocations, `ParseMZ`/`LoadEXE`), `dos.go`/`dos_int.go` (`INT 21h`
  services: file I/O, FindFirst/Next, MCB memory, vectors, scratch filesystem), `dos_mem.go`
  (MCB chain), `dos_ems.go` (LIM EMS `INT 67h`), `dos_video.go`/`dos_vga.go` (VGA BIOS + planar
  VGA), `dos_mouse.go` (`INT 33h`), `dos_io.go` (PC ports + timer IRQ), `dos_keyboard.go`
  (keyboard/mouse input injection). Game-agnostic — first proven on UW but nothing here is
  UW-specific.
- `extract/cmd/bootoracle` — the UW driver: loads `game/UW.EXE` on `tools/dos`, seeds `SAVE0`,
  and exposes `-trace`/`-log`/`-dump`/`-dis`/`-shot`/`-keys`/`-irq` with a decoded instruction
  ring and spin/runaway detectors.
- `extract/cmd/uwinfo` — Part I recon: decode the UW.EXE header/layout (via `tools/dos.ParseMZ`)
  and inventory `game/`.
- `extract/lev` + `extract/cmd/levinfo` — Part V: decode `LEV.ARK` (archive blocks + the 64×64
  tile map) and render a level; the tile fields were derived from the game's own decode code via
  the oracle's `-rdprof` read profiler.
- `extract/tex` — decode the `.TR` texture banks (W64/F32) and `PALS.DAT` palettes.
- `extract/levgeo` + `extract/cmd/levexport` — rebuild a level's static geometry (floors/walls,
  textured via the per-level list) and export it as a self-contained JSON for the Studio viewer;
  `cmd/levrender` software-renders it for verification.
- `disasm/` — the code-knowledge store: `uw.asm` (code-traced C-runtime startup) +
  `uw.annotations.txt` (the startup chain annotated with oracle-pinned addresses); see its README.

All addresses in this document are real-mode `segment:offset` or explicit file offsets; bytes are
little-endian. The `game/` data is copyrighted and not committed; verify a local copy of
`UW.EXE` against the MD5 above before reusing any offsets here.
