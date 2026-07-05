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
  first-person engine read from the oracle's live memory. *(started — the camera/view-matrix
  transform and the renderer's divide-error mechanism are mapped; `disasm/uw-render.*`)*
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

## 3. Still open

The per-frame render entry (from the main loop `2252:0410`), the perspective projection, the
wall/floor/texture span rasteriser and its texture mapper, object/billboard sprites, and the
off-screen-buffer → `A000` blit (a copy loop was seen at `231D:0070`). These are the next targets;
the peripheral-panel noise in `rendered/dungeon.png` is expected to resolve as the UI/HUD draw path
is mapped.

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
- `disasm/` — the code-knowledge store: `uw.asm` (code-traced C-runtime startup) +
  `uw.annotations.txt` (the startup chain annotated with oracle-pinned addresses); see its README.

All addresses in this document are real-mode `segment:offset` or explicit file offsets; bytes are
little-endian. The `game/` data is copyrighted and not committed; verify a local copy of
`UW.EXE` against the MD5 above before reusing any offsets here.
