# RetroReverse

Are the latest LLMs smart enough to reverse engineer old
software purely by static analysis, without access to a 3rd party debugger? This
repository contains the results of a few tests, attempting to answer that
question. All code and documents contained within (except this introduction) were
written by Claude Fable 5 or Opus 4.8. In some cases my Fable session got flagged, probably due to the nature of reverse engineering, and downgraded to Opus. Opus still did a fine job though.

Even though I explicitly restricted internet searches (released source code or other reverse engineering projects) Claude had detailed knowledge if the games I've had it analyze so far. When I asked how it knew exact function names for Elite it admitted that the original source code of the game, which has been publically released years ago, was part of its training data. The same is true for character and level names of Fort Apocalypse, for example.

Some of the prompts I used (for reference):

> Write a GoLang application that extracts program files from a C64 tape image
> in TAP format. The image in question might use a non-standard fastloader,
> that you need to reverse engineer by disassembling the loader. Use only
> static analysis of the image file without looking up external sources or
> external tools like Vice. Ignore memories from previous reverse engineering
> sessions. Document your findings about the tape image format in general and
> the fastloader in particular in a markdown file. Use example byte sequences
> in that documentation where appropriate. The tape image file in question is
> <GAME.TAP> in the current folder.

> Now that the program has been extracted from the tape, I want to enhance the scope of the project. Using the disassembler you already wrote, analyze the extracted program's startup and document your findings in a markdown file, again using assembly snippets or byte sequences as example where appropriate.

> Next, analyze the actual game code in the same way, again producing a markdown file that describes the initialization of the game until it reaches the main game loop. Put extra focus on the graphical elements like character sets, sprite data and level map. Describe compression or encryption schemes you find and build a memory map.

> Analyze the game code and describe the different types of objects (player, enemies, obstacles) in the game and how they behave in a markdown file. Be as detailed as possible about each object's behavior, movement patterns and collision behaviors.

## Methodology

[`ANALYSIS-PLAYBOOK.md`](ANALYSIS-PLAYBOOK.md) is a field guide to *how* these
projects were reverse-engineered — the common five-part arc, the reusable
toolkit, the end-to-end flows for C64 tape games and Amiga disk games, the
cross-cutting techniques (entropy triage, known-plaintext attacks, using a
self-built CPU core as an oracle), and a checklist for tackling a new unknown
image. Read it first if you're starting a new game or picking up an existing one.

## Repository structure

Each game lives in its own `games/<slug>/` folder following a common contract,
and the reusable building blocks live in a shared `tools` module at the root.
Tools are grouped into CPU cores (`tools/cpu/`), platform machine models and
their format/codec libraries (`tools/platform/`), cross-platform helpers
(`tools/lib/`), and the disassembler/tracer command binaries (`tools/cmd/`). A
`go.work` ties the modules together. The full contract is in
[`STANDARDS.md`](STANDARDS.md).

```
RetroReverse/
├── go.work                     # Go workspace over tools + each game's extract/
├── README.md                   # this file
├── STANDARDS.md                # repo-wide layout / CLI / asset-format contract
├── ANALYSIS-PLAYBOOK.md        # methodology: how the games are reverse-engineered
├── FORMAT2.md                  # the extracted asset / level "format 2" spec
├── tools/                      # shared tooling (module retroreverse.com/tools)
│   ├── cpu/                    #   CPU disassemblers + executable cores
│   │   ├── mos6502/            #     MOS 6502 (any 6502 platform)
│   │   ├── m68k/               #     Motorola 68000 (Amiga, ST, Genesis, …)
│   │   ├── z80/                #     Zilog Z80 (Game Gear, Master System, …)
│   │   ├── sm83/               #     Sharp LR35902 (Game Boy) — Z80 relative, not a Z80
│   │   ├── arm/                #     ARM (ARMv5TE/ARMv4T), little-endian — Nintendo DS
│   │   ├── arm60/              #     ARM60 (ARMv3), big-endian — 3DO
│   │   ├── mips/               #     MIPS R3000 + GTE — PlayStation
│   │   └── x86/                #     Intel x86, 16-bit real mode — MS-DOS
│   ├── platform/               #   platform machine models + format/codec libs
│   │   ├── c64/                #     tap, cbmtape, c64 machine, gfx, sid; cmd/tapdump
│   │   ├── amiga/              #     adf, hunk, iff, icon, powerpacker; cmd/adfdump, amigapng, …
│   │   ├── gamegear/           #     VDP decoders + z80-oracle machine
│   │   ├── gameboy/            #     MBC1 machine + DMG graphics decoders (sm83 oracle)
│   │   ├── nds/                #     cartridge reader (FNT/FAT, BLZ/LZ77, NARC), nitro, sdat, dsmachine; cmd/ndsinfo
│   │   ├── dos/                #     MZ loader + real-mode DOS/PC machine (x86 oracle)
│   │   ├── psx/                #     CD/EXE loader + PlayStation machine (mips oracle)
│   │   └── threedo/            #     OperaFS + 3DO machine, CEL/AIF decoders (arm60 oracle)
│   ├── lib/                    #   cross-platform helpers
│   │   └── glb/                #     binary glTF 2.0 model writer
│   └── cmd/                    #   dis<cpu> / codetrace<cpu> command binaries
│       ├── dis6502/ dis68k/ disz80/ dissm83/ disarm/ disarm60/ dismips/ disx86/
│       └── codetrace6502/ … codetracex86/
│
├── games/                      # one folder per game (module retroreverse.com/games/<slug>/extract)
│   ├── elite-c64/
│   │   ├── elite-c64.md        # writeup; pins the image MD5 near the top
│   │   ├── image/              # ROM/disk image(s) — .gitignored (copyright)
│   │   ├── extract/            # Go module: extraction + cmd/webexport + other tools
│   │   ├── disasm/             # optional annotated disassembly
│   │   ├── figures/            # only the images embedded in the writeup — tracked
│   │   └── work/               # regenerable dev/debug scratch — .gitignored
│   ├── ridge-racer-psx/
│   │   ├── ridge-racer-psx.md
│   │   ├── image/
│   │   └── extract/            # psx CD+EXE oracle + cmd/webexport
│   └── …                       # 12 games total (see the slug list below)
│
└── site/                       # the Studio web app (no-build PixiJS 2D + three.js 3D)
    ├── index.html
    ├── src/                    # the Studio and per-game viewers
    └── public/<slug>/          # canonical extracted assets: manifest.json + levels/ models/ music/ sprites/ bitmaps/
```

The twelve slugs are `elite-c64`, `fort-apocalypse-c64`, `marble-madness-amiga`,
`mario-kart-ds`, `need-for-speed-3do`, `ridge-racer-psx`, `sonic-gg`,
`stunt-car-racer-amiga`, `super-mario-64-ds`, `super-mario-land-gb`,
`turrican-amiga`, `ultima-underworld-pc`.

Per-game folder contract: a `<slug>.md` writeup (pinning the image MD5), an
`image/` directory for the raw ROM/disk/disc, and an `extract/` Go module whose
`cmd/webexport` command produces the web asset tree under `site/public/<slug>/`.
Where used, `disasm/` holds text disassembly, `figures/` holds only the images
embedded in the writeup (tracked), and `work/` holds regenerable scratch
(git-ignored). The old per-game `rendered` and `extracted` output directories
are gone.

## Image files

Every game is analyzed from one exact image (ROM, disk, or disc); the
documentation and the golden extraction tests assume it byte for byte. Game
images are **copyrighted and never committed** — they are `.gitignore`d, and
each game's writeup pins the precise copy by filename, size, and MD5 in its
**Image** line near the top. Supply your own copy under
`games/<slug>/image/` before running that game's tools, and verify it, e.g.
`md5 games/elite-c64/image/Elite.tap` (`md5sum` on Linux).

## Shared tools (`tools`)

CPU cores live under `tools/cpu/`, their command-line front-ends under
`tools/cmd/`, platform machine models and format/codec libraries under
`tools/platform/`, and cross-platform helpers under `tools/lib/`. Import paths
are `retroreverse.com/tools/cpu/...`, `.../tools/platform/...`,
`.../tools/lib/...` and `.../tools/cmd/...`.

### CPU cores (`tools/cpu`)

Each core pairs a `Decode`/`Disassemble` disassembler with an instruction-level
`CPU` execution core over a shared `Bus`/`Step()` interface, so the same package
both disassembles and *runs* code.

| Package | What it does |
|---------|--------------|
| `cpu/mos6502` | One opcode table driving both a `Disassemble` function and an executable `CPU` core (all documented opcodes, binary + BCD) — usable by any 6502 platform. |
| `cpu/m68k` | Motorola 68000 toolkit: a disassembler (full documented instruction set, all addressing modes) **and** an execution core. The core runs a minimal-but-correct opcode subset (MOVE/ALU/shift/branch/jump/DBcc/MOVEM/LINK with proper X/N/Z/V/C flags) and halts on anything not yet implemented, so gaps are explicit. Usable by any 68k platform (Amiga, ST, Genesis, …). |
| `cpu/z80` | Zilog Z80 toolkit: a disassembler over the CPU's x/y/z/p/q opcode bit-fields (all `CB`/`ED`/`DD`/`FD` prefix pages, including `DDCB` and the undocumented `IXH`/`IXL`/`SLL`) **and** a practically complete execution core (block moves, IM-1 interrupts, documented flags). Usable by any Z80 platform (Game Gear, Master System, ZX, MSX, …). |
| `cpu/sm83` | Sharp **LR35902** (Game Boy CPU; "GBZ80"/SM83) toolkit: a disassembler **and** an execution core, with `Step` returning T-cycles for the machine to drive timing. A Z80 relative but **not** a Z80, so it is its own package: only four flags (Z/N/H/C), no `IX`/`IY` or `DD`/`FD`/`ED` prefixes (only `CB`), no `IN`/`OUT` (the Game-Boy-only `LDH ($FF00+n)`/`(C)` high-page ops instead), the `LD (HL±)` auto-inc/dec loads, `LD ($nnnn),SP`, `ADD SP,e`, `LD HL,SP+e`, `RETI`, `STOP`, CB-page `SWAP`, and eleven illegal opcodes. Unit-tested and verified end-to-end booting Super Mario Land. |
| `cpu/arm` | **ARM** toolkit (little-endian) for the two Nintendo DS CPUs — the ARM9 (ARM946E-S, ARMv5TE) and ARM7 (ARM7TDMI, ARMv4T): a disassembler **and** an execution core handling **both** DS instruction sets — 32-bit ARM (conditional execution, barrel-shifter operand, multiply/long-multiply, LDR/STR/LDM/STM, the ARMv5 `BLX`/`CLZ`/`BKPT`, saturating `QADD…` and signed `SMLA…` DSP ops) and 16-bit `Thumb` — and the `BX`/`BLX` interworking between them, tracked per address (`Inst.Thumb`/`TargetThumb`). Models the CPSR flags, banked register file and mode switching (SWI/IRQ via caller hooks; CP15 via a `Coproc` hook); caches, MPU, timing and the 2D/3D video hardware are the caller's "full machine". Unit-tested (incl. ARM↔Thumb interworking). |
| `cpu/arm60` | **ARM60** toolkit (ARMv3, **big-endian**) for the 3DO's ARM60 CPU — the big-endian counterpart to `cpu/arm`, ARMv3 (no Thumb, no ARMv5 ops). Disassembler + execution core. |
| `cpu/mips` | **MIPS R3000** toolkit for the PlayStation: a disassembler + execution core (the R3000 integer set with load-delay slots and branch-delay slots, COP0, and the **GTE** geometry coprocessor for 3D). Differentially validated against a SingleStepTests/psx suite (`PSX_SST_DIR`, env-gated). |
| `cpu/x86` | **Intel x86** toolkit for 16-bit real-mode code (MS-DOS, Ultima Underworld's `UW.EXE`): a disassembler over the variable-length x86 encoding — prefix runs, ModR/M + SIB in 16- and 32-bit addressing, the 8086/186/286 integer set plus the common 386 additions (`MOVZX`/`MOVSX`, `SHLD`/`SHRD`, `BT`/`BSF`/…, `SETcc`, near `Jcc`), and the x87 `D8`–`DF` escapes — **and** an execution core (real-mode segmentation, correct flags, REP string ops, IN/OUT via port hooks, software INT/IRET via an `IntHook`, maskable hardware-IRQ injection with a `MOV SS` interrupt-shadow). **Differentially validated against the SingleStepTests/8088 suite** (`harte_test.go`, `HARTE_DIR`, ~155 opcodes — caught a real 8086 segment-wrap bug) and driven by the Ultima Underworld DOS oracle through the game's entire initialisation. |

### Disassembler / tracer commands (`tools/cmd`)

Every CPU has a linear disassembler (`dis<cpu>`) and a recursive-descent code
tracer (`codetrace<cpu>`) at parity. The tracers share one flag vocabulary
(`-base`/`-skip`, `-entry`, `-table ADDR:N`, `-annotate FILE`, `-o FILE`) with
per-architecture extras (`-thumb`, `-bank`, …).

| Command | What it does |
|---------|--------------|
| `cmd/dis6502` | Linear disassembler for a `.prg` file (2-byte load address + data), optionally over an address range. |
| `cmd/dis68k` | Linear disassembler for a raw 68000 code blob loaded at a given base address (`-skip` steps past an AmigaDOS hunk header). |
| `cmd/disz80` | Linear disassembler for a raw Z80 code slice mapped at a given address (`-off`/`-len`/`-base`). |
| `cmd/dissm83` | Linear disassembler for a Game Boy ROM: a flat file slice (`-off`/`-len`/`-base`) or an MBC1 bank view (`-bank N -start A -end A`). |
| `cmd/disarm` | Linear disassembler for a raw ARM/Thumb blob (`-off`/`-len`/`-base`, `-thumb` selects Thumb). |
| `cmd/disarm60` | Linear disassembler for a raw big-endian ARM60 (ARMv3) blob. |
| `cmd/dismips` | Linear disassembler for a raw MIPS R3000 code slice. |
| `cmd/disx86` | Linear disassembler for a raw 16-bit real-mode x86 blob (`-base`/`-skip`/`-start`/`-end`); `-skip` steps past UW.EXE's MZ header. |
| `cmd/codetrace6502` | Recursive-descent 6502 disassembler: from given entry points (and seeded jump tables) it follows every branch/jump/call, marks reachable code vs data, and lists routines and unresolved indirect jumps — so tables and graphics aren't mis-decoded as instructions. |
| `cmd/codetrace68k` | The 68000 counterpart: same recursive trace over a raw blob loaded at `-base` (`-skip` past a hunk header). |
| `cmd/codetracez80` | The Z80 counterpart over a banked ROM (`-load` selects the resident bank). |
| `cmd/codetracesm83` | The Game Boy counterpart over a 32 KB bank view (`-bank N`; bank 0 fixed, bank *N* in `$4000`–`$7FFF`). |
| `cmd/codetracearm` | The ARM counterpart over a flat image, tracking ARM/Thumb state through `BX`/`BLX` interworking (entries suffixed `t`/`a`, or bit 0 of a pointer, select the state). |
| `cmd/codetracearm60` | The big-endian ARM60 counterpart. |
| `cmd/codetracemips` | The MIPS R3000 counterpart. |
| `cmd/codetracex86` | The x86 counterpart over the MZ load module (`-base`/`-skip`), with `-table` (16-bit LE near-offset jump tables). |

### Platform tools (`tools/platform`)

Each platform folder holds its machine model (driving the matching CPU core as
an *emulation oracle*) plus its format/codec libraries.

| Package / command | What it does |
|-------------------|--------------|
| `platform/c64/tap` | Parse a TAP v0/v1 image (C64/C16) into a pulse stream; `Segmentize` splits it at pauses. |
| `platform/c64/cbmtape` | Decode the standard Commodore KERNAL (ROM loader) tape encoding: blocks, headers, and paired header+data files with checksum verification. |
| `platform/c64/c64` | A minimal C64 machine model — RAM, the `cpu/mos6502` CPU, a CIA pulse-feed tape model, a PC-hook registry, a RAM write log and an optional read probe — for *running* a self-modifying loader instead of decoding it, or tracing which routine touches which memory. Optional standard KERNAL tape hooks included. |
| `platform/c64/gfx` | C64/VIC rendering (palette, multicolor characters, hires sprites, multicolor bitmaps) plus general 2-D helpers (line drawing, markers, still/animated PNG output). |
| `platform/c64/sid` | A reusable MOS 6581 SID emulator for rendering C64 music. |
| `platform/c64/cmd/tapdump` | Print a pulse-width histogram and the pause-delimited segment map of a `.tap` — the usual first look at an unknown tape. |
| `platform/amiga/adf` | Read a standard AmigaDOS floppy image (ADF): detect OFS/FFS, walk the directory tree, and extract file contents (hash chains, OFS data-block headers, multi-block extension chains). |
| `platform/amiga/hunk` | Load an AmigaDOS hunk object/executable: place its CODE/DATA/BSS segments from a base, apply the 32-bit relocations, and return a flat image (and each segment's base) ready to disassemble. |
| `platform/amiga/iff` | Decode an IFF `FORM…ILBM` bitmap (planar BODY, ByteRun1/uncompressed, CMAP palette) into a Go image. |
| `platform/amiga/icon` | Decode a Workbench `.info` icon (DiskObject + planar Image structs), using the standard Workbench palette. |
| `platform/amiga/powerpacker` | Decompress PowerPacker (`PP20`) data — one of the most common Amiga crunchers. Faithful reimplementation of the standard backward bit-reader decode loop. |
| `platform/amiga/cmd/adfdump` | List an `.adf`'s volume, directory tree and file sizes; `-x outdir` extracts every file preserving the directory structure. |
| `platform/amiga/cmd/hunkload` | Print a hunk file's segment map; `-o` writes its flat relocated image and `-syms` writes its symbol table as a `codetrace68k` annotations file. |
| `platform/amiga/cmd/amigapng` | Render an IFF ILBM or a `.info` icon to PNG (auto-detects the format). |
| `platform/amiga/cmd/ppdecrunch` | Decompress a `PP20` file, or a `PP20` block embedded at a `-off`/`-len` slice of a larger file. |
| `platform/gamegear` | Sega Game Gear VDP graphics — the 4-bitplane tile, 12-bit CRAM palette and name-table decoders — plus a minimal `Machine` (8 KB RAM + Sega cartridge mapper + VDP ports) that drives the `cpu/z80` core as an oracle. Usable by any Game Gear (and, for the tiles, Master System) game. |
| `platform/gameboy` | Game Boy (DMG) machine model driving the `cpu/sm83` core as an oracle: the MBC1 mapper, the full memory map, and the timer and LCD scanline counter with their VBlank/STAT/timer interrupts (`RunFrame`/`RunFrames`, plus a PC histogram and a VRAM write-watch). Also the fixed DMG **graphics decoders** (`gb.go`): the 2bpp tile, `BGP`/`OBP` palettes, tile-sheet and 32×32 background-map composition, and an OAM/sprite screen compositor. |
| `platform/nds` | Nintendo DS cartridge (`.nds`) container reader: the ROM header (CRC-16 verification), the ARM9/ARM7 binaries and overlay tables, the on-cartridge filesystem (**FAT** joined to the **FNT** tree), and the **BLZ** backward-LZSS decompressor (`DecompressBLZ`/`IsBLZ`). Also `DecompressLZ77` (forward LZ10/LZ11), `ParseNARC` (splits a Nintendo ARChive, decompressing a `.carc` wrapper first), the `nitro` and `sdat` decoders, and the `dsmachine` dual-core model. |
| `platform/nds/nitro` | NITRO-System resource decoders. 3D textures: `DecodeNSBTX` turns a `BTX0`/`TEX0` set into Go images (**all seven** DS texture formats incl. the 4x4-block-compressed one, BGR555 palettes, name-similarity texture↔palette pairing). 2D tile art: `ParseNCLR`/`ParseNCGR`/`ParseNSCR` + `ComposeScreen`/`TileSheet`. **3D models**: `ParseNSBMD` (nodes/TRS, SBC scene bytecode, materials, shape display lists), `RunSBC` + `DecodeDL` (the GX geometry-command interpreter), and `ExportGLB`. |
| `platform/nds/dsmachine` | Reusable **dual-core DS machine**: the ARM9 and ARM7 on two `cpu/arm` cores over one shared main RAM + WRAM, wired by the cross-wired **IPCSYNC** mailbox and the two **IPC FIFOs**, a per-core interrupt controller and the BIOS `SWI`s. Co-runs both boot chains and clears the ARM9↔ARM7 rendezvous a single core cannot. IPC cross-wiring unit-tested. |
| `platform/nds/cmd/ndsinfo` | DS container inspector: prints the header, integrity checks, the ARM9/ARM7/overlay layout and the filesystem catalog (`-files`, `-tree`, `-grep`). |
| `platform/dos` | Real-mode MS-DOS/PC machine driving the `cpu/x86` core as an oracle: the MZ loader (parse + apply relocations), an MCB-chain memory manager, LIM EMS 4.0, DOS `INT 21h` file I/O, a VGA video BIOS (`INT 10h`) with planar/Mode-X emulation, an 8042 keyboard / Sound Blaster port model, and injected timer IRQs. Game-agnostic; the MZ parser is unit-tested. |
| `platform/psx` | PlayStation machine driving the `cpu/mips` core (with GTE): the CD-ROM controller + disc image reader, the PS-EXE loader, a GPU rasteriser, and the BIOS-call layer. |
| `platform/threedo` | 3DO machine driving the `cpu/arm60` core: the **OperaFS** disc filesystem reader, the CEL image and AIF audio decoders, and the kernel/folio/task model. |

### Cross-platform helpers (`tools/lib`)

| Package | What it does |
|---------|--------------|
| `lib/glb` | A standalone binary glTF 2.0 (GLB) writer — triangle meshes with materials, embedded PNG textures, per-material single-/double-sided flags — shared by every game that exports 3D models to the Studio. |

## Building and running

The `go.work` workspace lets each game's `extract` module resolve the local
`tools` module. Build and test each module from its own directory:

```sh
( cd tools && go test ./... )
( cd games/elite-c64/extract && go test ./... )
( cd games/ridge-racer-psx/extract && go test ./... )
```

(The integration tests skip automatically when the game image is absent.)
The shared module's import path is `retroreverse.com/tools`, so its commands
are run by that full path from anywhere in the workspace (the `go.work` is
found by walking up from the current directory), e.g.
`go run retroreverse.com/tools/platform/c64/cmd/tapdump ...`.

`tapdump` is generic — point it at any C64 `.tap` to see its pulse encoding
and segment layout (the first step when approaching an unfamiliar tape):

```sh
go run retroreverse.com/tools/platform/c64/cmd/tapdump path/to/any.tap
```

`adfdump` is likewise generic for Amiga disks — it lists and extracts the files
of any standard AmigaDOS floppy image:

```sh
go run retroreverse.com/tools/platform/amiga/cmd/adfdump path/to/disk.adf        # list
go run retroreverse.com/tools/platform/amiga/cmd/adfdump -x out path/to/disk.adf # extract
```

Each game's `extract` module exposes a single `webexport` command that runs the
whole extraction from the raw image and writes the **format-2** asset tree
(`manifest.json` + `levels/ models/ music/ sprites/ bitmaps/`) to
`site/public/<slug>/`, which the Studio at `site/` consumes. `-in` is the input
image, `-o` the output root (defaults to `../../site/public/<slug>`), and
`-only` selects a subset of stages (`levels,models,music,…`). Run it from the
game's `extract/` folder:

```sh
cd games/elite-c64/extract && go run ./cmd/webexport -in ../image/Elite.tap

cd games/sonic-gg/extract && go run ./cmd/webexport \
    -in "../image/Sonic The Hedgehog (Japan, USA).gg" -only levels
```

Each `webexport` is game-specific — written for that game's loader and formats —
so extracting a *new* game means writing its own `extract`/`webexport` on top of
`tools` (see "Two extraction strategies" below), not reusing one of these.

Disassemble a code slice with the shared tools — `dis6502` for 6502, `dis68k`
for 68000 (`-skip` steps past an Amiga hunk header):

```sh
go run retroreverse.com/tools/cmd/dis6502 -start 8927 -end 8A40 fort-fast-7000.prg

go run retroreverse.com/tools/cmd/dis68k -skip 36 amiga-code-hunk.bin
```

## Two extraction strategies

The two tapes need fundamentally different approaches, and the shared tools
support both:

- **Declarative decoding** (Fort Apocalypse). The fastloader is static: a
  fixed Novaload-family format with page records and checksums. The extractor
  reimplements that format on top of `platform/c64/tap` + `platform/c64/cbmtape`
  and reads the payload straight off the pulse stream.

- **Run-the-loader emulation** (Elite). The fastloader rewrites its own wire
  format on the fly — bit order, bits-per-byte, header size and sync handling
  all change mid-load, driven by patch blocks that arrive on the tape itself.
  Reimplementing the protocol is hopeless, so the extractor *runs the actual
  loader* on the `platform/c64/c64` + `cpu/mos6502` machine model, feeding it
  the tape pulses and logging every memory write the loader performs.

For a new C64 tape the workflow is: `tapdump` to see the encoding, `cbmtape`
to read the boot file, `dis6502` to disassemble the loader, then choose a
strategy — a static loader is a new decoder package; a hostile self-modifying
one is a set of hooks on the `platform/c64/c64` machine model.
