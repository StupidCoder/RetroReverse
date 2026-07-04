# Mario Kart DS (Nintendo DS) — cartridge format and game analysis

A reverse-engineering reference for `Mario Kart DS (Europe) (En,Fr,De,Es,It).nds`, the 2005 Nintendo kart racer. This is the repository's first **Nintendo DS** title and its first **ARM** project: where every earlier game runs on a single 8- or 16-bit CPU, the DS is a dual-processor machine — an **ARM9** (ARM946E-S, ARMv5TE, the 67 MHz main CPU) beside an **ARM7** (ARM7TDMI, ARMv4T, the 33 MHz sound/IO CPU) — and its `.nds` cartridge is a real container with a header, two CPU binaries, code overlays and an on-cartridge filesystem, unlike the flat mask-ROM of a Game Boy cartridge or the raw disk of an Amiga floppy. The writeup follows the same shape as the others, in reading order:

* **Part I** — the cartridge image: the `.nds` container, its header, the ARM9/ARM7 binaries and overlays, and the FNT/FAT filesystem that names ~600 asset files;
* **Part II** — the boot chain: the cartridge header's entry points, the secure area, and how the ARM9 and ARM7 come up and hand off;
* **Part III** — program architecture: how the ARM9 image decompresses/relocates itself, sets up the TCMs and the interrupt table, and the overlay loading that streams code in per game mode;
* **Part IV** — graphics and data formats: the NITRO asset formats (`NSBMD` models, `NSBTX` textures, `NSBCA` animation), the `NARC` archive and its LZ compression, and the 2D banner/UI graphics;
* **Part V** — game mechanics: track data, kart physics and item behaviour;
* **Appendix A** — toolchain and reproduction.

Methods: purely static analysis of the `.nds` image. The DS needed a new toolchain — the shared 6502/68000/Z80/LR35902 decoders do not apply — so this project is built on the new `tools/nds` cartridge reader (header, FNT/FAT, overlays) and the new `tools/arm` ARM/Thumb disassembler and CPU core, with `tools/nds/cmd/ndsinfo` for the container catalog and `retroreverse.com/tools/cmd/disarm` / `codetracearm` for the code. All addresses are 32-bit ARM addresses (`$02000000`-style main-RAM addresses, or the ARM9/ARM7 BIOS and I/O regions) unless a *file offset* into the ROM image is explicitly called out; bytes are little-endian. **Part I is complete; Parts II–V are stubs.**

---

## Contents

- [Part I — The cartridge image](#part-i--the-cartridge-image)
  - [1. The ROM dump](#1-the-rom-dump)
  - [2. The DS dual-CPU architecture and the address map](#2-the-ds-dual-cpu-architecture-and-the-address-map)
  - [3. The cartridge header (`$000`–`$15F`)](#3-the-cartridge-header-00015f)
  - [4. The ARM9 and ARM7 binaries and overlays](#4-the-arm9-and-arm7-binaries-and-overlays)
  - [5. The filesystem: FNT and FAT](#5-the-filesystem-fnt-and-fat)
  - [6. The file catalog](#6-the-file-catalog)
- [Part II — The boot chain](#part-ii--the-boot-chain)
- [Part III — Program architecture](#part-iii--program-architecture)
- [Part IV — Graphics and data formats](#part-iv--graphics-and-data-formats)
- [Part V — Game mechanics](#part-v--game-mechanics)
- [Appendix A — Toolchain and reproduction](#appendix-a--toolchain-and-reproduction)

---

# Part I — The cartridge image

Where a Game Boy `.gb` is a verbatim copy of a mask-ROM chip with no container at all, and an Amiga `.adf` is a raw track dump you have to prove is *not* an AmigaDOS volume, a Nintendo DS `.nds` is a genuine **structured container**: a fixed header at offset 0 points at two separate CPU boot binaries, their code overlays, and a small filesystem — a directory tree (**FNT**, file name table) and an allocation table (**FAT**, a flat array of start/end offsets) — that names and locates the hundreds of asset files the game streams off the cartridge. This Part walks that structure top-down and pins every offset against the real image.

## 1. The ROM dump

The image is **33,554,432 bytes = 32 MiB** exactly, a power of two, matching the cartridge chip size the header advertises (§3: device capacity `$08` ⇒ `128 KiB << 8` = 32 MiB). There is no copier/loader header — byte *N* of the file is byte *N* the cartridge hardware reads. Of the 32 MiB, `$01860B9E` (≈ 25.4 MiB) is the *used* size the header records (§3, `$080`); the tail is pad.

Two integrity fields in the header are checkable without running anything, and both verify:

- the **header checksum** at `$15E` is a CRC-16 (MODBUS: reflected poly `$A001`, init `$FFFF`) over header bytes `$000`–`$15D`. Computed `$73A0`, stored `$73A0` ✓.
- the **Nintendo-logo checksum** at `$15C` covers the compressed logo bitmap at `$0C0`–`$15B` that the BIOS validates on boot; stored `$CF56`, the fixed value every genuine retail cartridge carries ✓.

The image is pinned by size and MD5 in the repository [README](../README.md#image-files); the Appendix repeats the verify command. All of Part I is produced by `tools/nds/cmd/ndsinfo` reading those header/FNT/FAT structures directly out of the image.

## 2. The DS dual-CPU architecture and the address map

The header (§3) hands the two CPUs *different* load and entry addresses, which is the first hard evidence of the DS's split-brain design:

- the **ARM9** binary loads to `$02000000` in **main RAM** and starts at `$02000800`;
- the **ARM7** binary loads to `$02380000` — also inside main RAM, near its top — and starts there.

Both CPUs share the single 4 MiB main memory at `$02000000`–`$023FFFFF`; the ARM7's code simply lives in a region the ARM9 agrees to leave alone. The rest of each CPU's view is fixed DS hardware. The ARM9 memory map, from its own perspective:

| Range | Size | Contents |
|---|---:|---|
| `$00000000`–`$00007FFF` | 32 KiB | ITCM — instruction tightly-coupled memory (fast, CPU-local) |
| `$02000000`–`$023FFFFF` | 4 MiB | Main RAM (shared with ARM7; ARM9 image loads at the base) |
| *(configurable)* | 16 KiB | DTCM — data tightly-coupled memory, base set by code at boot (Part III) |
| `$03000000`–`$03007FFF` | 32 KiB | Shared WRAM (mapping split with ARM7 via `WRAMCNT`) |
| `$04000000`–`$04FFFFFF` | — | Memory-mapped I/O (video, DMA, timers, IPC, cartridge) |
| `$05000000`–`$050007FF` | 2 KiB | Palette RAM (BG + OBJ, both screens) |
| `$06000000`–`$06FFFFFF` | ≤ 656 KiB | VRAM banks A–I (mapped to BG/OBJ/texture/LCDC uses) |
| `$07000000`–`$070007FF` | 2 KiB | OAM (sprite attributes, both screens) |
| `$08000000`–`$09FFFFFF` | 32 MiB | GBA slot ROM (Slot-2) |
| `$FFFF0000`–`$FFFF7FFF` | 32 KiB | ARM9 BIOS |

The ARM7 sees a smaller, different map: its own 16 KiB BIOS at `$00000000`, the shared main RAM at `$02000000`, a private 64 KiB WRAM at `$03800000` (plus its slice of the shared `$03000000` WRAM), its own I/O block at `$04000000` (sound, SPI/touchscreen, RTC) and the Wi-Fi region at `$04800000`. The two processors coordinate through the **IPC** hardware FIFO and shared-memory mailboxes in the `$04000000` I/O space — the mechanism traced in Part II.

The ARM9 is an ARMv5TE core and the ARM7 an ARMv4T core; both run the 32-bit **ARM** and 16-bit **Thumb** instruction sets, switched via `BX`/`BLX`. Disassembling either binary therefore means tracking ARM-vs-Thumb state per address, which is exactly what `tools/arm` and `codetracearm` do (Appendix); the boot code is ARM, but large amounts of the game are Thumb to save the 32 MiB.

## 3. The cartridge header (`$000`–`$15F`)

The first 512 bytes are the cartridge header (the reserved header region is `$4000`, `$084`). Decoding it against the real image — raw bytes shown little-endian as stored, with the assembled value on the right:

```
$000  4D 41 52 49 4F 4B 41 52 54 44 53 00   title            "MARIOKARTDS"
$00C  41 4D 43 50                            game code        "AMCP"  (A=DS card, MC=Mario Kart, P=Europe/PAL)
$010  30 31                                  maker code       "01"    — Nintendo
$012  00                                     unit code        — NDS (0)
$013  00                                     encryption seed  — 0
$014  08                                     device capacity  — 128 KiB << 8 = 32 MiB  ✓ matches file size
$01E  00                                     ROM version      — 0
$01F  00                                     autostart flags  — 0
$020  00 40 00 00                            ARM9 ROM offset  = $00004000
$024  00 08 00 02                            ARM9 entry addr  = $02000800
$028  00 00 00 02                            ARM9 RAM addr    = $02000000
$02C  1C 75 0E 00                            ARM9 size        = $000E751C   (≈ 948 KiB)
$030  00 16 14 00                            ARM7 ROM offset  = $00141600
$034  00 00 38 02                            ARM7 entry addr  = $02380000
$038  00 00 38 02                            ARM7 RAM addr    = $02380000
$03C  A4 89 02 00                            ARM7 size        = $000289A4   (≈ 166 KiB)
$040  00 A0 16 00                            FNT offset       = $0016A000
$044  9E 2E 00 00                            FNT size         = $00002E9E
$048  00 D0 16 00                            FAT offset       = $0016D000
$04C  10 13 00 00                            FAT size         = $00001310   (÷8 = 610 entries)
$050  00 B6 0E 00                            ARM9 overlay tbl = $000EB600
$054  80 00 00 00                            ARM9 overlay sz  = $00000080   (÷32 = 4 overlays)
$058  00 00 00 00                            ARM7 overlay tbl = none
$05C  00 00 00 00                            ARM7 overlay sz  = 0
$068  00 E4 16 00                            icon/banner off  = $0016E400
$06C  47 DE                                  secure-area CRC  = $DE47
$070  58 0A 00 02                            ARM9 autoload hook = $02000A58
$074  58 01 38 02                            ARM7 autoload hook = $02380158
$080  9E 0B 86 01                            total used size  = $01860B9E   (≈ 25.4 MiB of 32)
$084  00 40 00 00                            header size      = $00004000
$0C0  24 FF AE 51 …                          Nintendo logo bitmap ($0C0–$15B, BIOS-validated)
$15C  56 CF                                  logo checksum    = $CF56  ✓
$15E  A0 73                                  header checksum  = $73A0  ✓ (CRC-16 over $000–$15D)
```

Observations:

- **Two independent code images.** The ARM9 (`$4000`, 948 KiB) and ARM7 (`$141600`, 166 KiB) binaries are stored back-to-back with their overlays between them; each carries its own RAM-load address and entry point, so the header alone fully describes where each CPU starts (Part II).
- The **ARM9 entry `$02000800`** sits `$800` bytes into its load region `$02000000` — the first `$800` bytes are the header-copy/secure-area landing that the BIOS fills in during the encrypted boot handshake (Part II).
- The **autoload hooks** (`$070`/`$074`) are the addresses of each binary's auto-load list, the table the boot stub walks to copy its `.data` and clear its `.bss` before jumping to `main` — the seam for Part III's relocation trace.
- The header checksum covers only `$000`–`$15D`; the secure-area CRC (`$06C`) and logo CRC (`$15C`) cover other regions, so the three are independent and all check out.

## 4. The ARM9 and ARM7 binaries and overlays

Beyond the two main binaries, the ARM9 uses **code overlays** — separately loadable code+data blobs swapped into a fixed RAM region on demand, the DS's answer to fitting far more code than RAM holds. The overlay *table* at `$0EB600` (`$054` = `$80` bytes ⇒ **4 overlays**) is an array of 32-byte records; decoding it:

```
ovl  RAM addr    size      bss       static-init range      FAT id
 0   $021804E0   $034780   $002F40   $021B07E4–$021B07E8      0
 1   $021804E0   $029500   $020A60   $021A7684–$021A7688      2
 2   $021804E0   $000520   $000005   …                        3
 3   $021804E0   …                   …                        1
```

Every overlay loads to the **same** RAM address `$021804E0` — they are mutually-exclusive banks, only one resident at a time — and each names the **FAT file ID** that holds its bytes. The four overlays claim FAT IDs `{0, 1, 2, 3}` (the mapping is `ovl0→0, ovl1→2, ovl2→3, ovl3→1`), which is why the filesystem's named files begin at ID 4 (§5). Each record also carries a `.bss` size to zero and a static-initialiser (constructor) address range the loader runs after copying — the same auto-load convention as the main binaries. The ARM7 has no overlays (`$058`/`$05C` = 0). *Which* overlay is loaded for which game mode, and the loader that reads these records, is a Part III trace.

## 5. The filesystem: FNT and FAT

Everything that is not boot code — models, textures, tracks, UI graphics, the sound bank — lives in an on-cartridge filesystem addressed by two tables the header locates:

- the **FAT** (file allocation table) at `$16D000`, `$1310` bytes, is a flat array of 8-byte records: a **start** and **end** file offset per file. `$1310 / 8 = 610` files. File *ID* is just the index; the bytes are `image[start:end]`. This is *storage framing only* — it says where a file is, nothing about what it contains.
- the **FNT** (file name table) at `$16A000`, `$2E9E` bytes, is the **directory tree** that gives those numbered files names and a hierarchy. Its main table is an array of 8-byte directory records (record 0 = the root, whose "parent" field instead holds the total directory count); each record points at a **sub-table** listing that directory's entries. A sub-table entry is a control byte — `$00` ends the directory, `$01`–`$7F` is a file whose name length is the low 7 bits, `$81`–`$FF` is a subdirectory — followed by the name, and for a subdirectory a 2-byte child directory ID (`$F000 | n`). Files take **sequential IDs** starting from the directory record's "first file ID", which is how names bind to FAT indices.

`tools/nds` walks the tree and joins the two tables: of the 610 FAT entries, IDs **0–3 are the four ARM9 overlays** (they have no filesystem name — the overlay table references them directly, §4), and IDs **4–609 are the 606 named files**. Walking from the root yields paths like `data/Course/airship_course.carc` and `data/KartModelMenu/character/mario/…`.

One layering note, per the "separate the layers" rule: a filesystem file is *not* the asset. The first named file, `data/CharacterKartSelect.carc` (ID 4, `$27CA00`–`$28069F`), begins:

```
10 0C 85 00  00 4E 41 52 43 …            ".....NARC"
```

`$10` is the Nintendo **LZ77 (type 0x10)** compression tag and `$00850C` the decompressed size (34,060 bytes); after decompression the magic is **`NARC`** — a Nintendo ARChive, itself a mini-container of sub-files. So a single asset path is three nested layers — *FAT slice → LZ77 stream → NARC archive → NITRO asset* — each decoded in its own Part (the compression and NARC/NITRO formats are Part IV). Part I stops at the storage framing.

## 6. The file catalog

The 606 named files group cleanly by extension, and the histogram already sketches the engine's asset pipeline before any of it is decoded:

| Ext | Count | What it is (to be confirmed in Part IV) |
|---|---:|---|
| `.carc` | 286 | LZ77-compressed `NARC` archive (bundles of the below) |
| `.nsbmd` | 82 | NITRO 3D model (geometry + material) |
| `.nsbtx` | 56 | NITRO texture set |
| `.bin` | 40 | raw/engine-specific data |
| `.nsbca` | 21 | NITRO character (joint) animation |
| `.prm` | 8 | (parameter tables — TBD) |
| `.nsbtp` | 6 | NITRO texture-pattern animation |
| `.nsbma` | 2 | NITRO material animation |
| `.spa` | 1 | particle/effect archive |
| `.sdat` | 1 | the entire SDAT sound bank (SSEQ/SBNK/SWAR) |
| `.nsbta` | 1 | NITRO texture SRT animation |
| `.nbfc` / `.nbfp` | 1 / 1 | 2D tile graphics / palette |

The top-level directories tell the same story: `data/Course` (118 files — the tracks and their textures, e.g. `airship_course.carc` + `airship_courseTex.carc`), `data/CupPicture` (96), `data/Ghost` (33 — the staff-ghost replays), `data/KartModelMenu` with a `character/<name>/` and `kart/<name>/` subtree per driver (mario, luigi, peach, yoshi, donkey, wario, waluigi, daisy, koopa, kinopio, robo, karon), plus `data/Boot`, `data/CupPicture`, `data/Race`, `data/MainMenu` and `data/MapObj`. The full catalog is reproducible with `ndsinfo -files` (Appendix). Decoding what is *inside* these files is Parts IV–V.

---

# Part II — The boot chain

*(frontier)* How the two CPUs come up: the BIOS's encrypted KEY1 handshake and the **secure area** at `$4000` (the first 16 KiB of the ARM9 binary, secure-CRC `$DE47` at header `$06C`); the ARM9 boot stub at entry `$02000800` and its auto-load list at `$02000A58`; the ARM7 boot at `$02380000`; and the IPC/FIFO synchronisation by which the ARM9 hands the ARM7 its work. Traced with `codetracearm` over the ARM binaries extracted by `ndsinfo`, confirmed on the `tools/arm` core.

# Part III — Program architecture

*(frontier)* ARM9 self-relocation and `.bss` clear via the auto-load list; ITCM/DTCM setup through CP15 (the `MCR p15` sequence); the interrupt table and the main game loop; and the **overlay loader** that reads the §4 overlay table to swap code per game mode.

# Part IV — Graphics and data formats

*(frontier)* The compression and container layers (LZ77 `0x10`/`0x11`, Huffman, RLE; the `NARC` archive), then the NITRO asset formats — `NSBMD` models, `NSBTX` textures, `NSBCA`/`NSBTP`/`NSBTA`/`NSBMA` animation, the `NCLR`/`NCGR`/`NSCR` 2D graphics, and the `SDAT` sound bank.

# Part V — Game mechanics

*(frontier)* Track collision and route data, kart physics, item behaviour, the CPU racers, and the staff-ghost replay format in `data/Ghost`.

---

# Appendix A — Toolchain and reproduction

Everything in Part I is derived by pure static inspection of the `.nds` image; no emulator, debugger, or third-party tool was used, and nothing was read from released source. Verify the image, then reproduce the catalog:

```sh
# identity (size + MD5 pinned in ../README.md#image-files)
md5 "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"

# header, integrity checks, overlay/filesystem summary
go run retroreverse.com/tools/nds/cmd/ndsinfo "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"

# full file catalog (ID, byte range, size, path)
go run retroreverse.com/tools/nds/cmd/ndsinfo -files "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"

# directory tree with per-directory file counts
go run retroreverse.com/tools/nds/cmd/ndsinfo -tree "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"
```

Toolchain (all under the `retroreverse.com/tools` module, this repository):

- **`tools/nds`** — the Nintendo DS cartridge container reader: header parse + CRC-16 verification, the FAT, and the FNT directory-tree walk that names every file. New for this project.
- **`tools/nds/cmd/ndsinfo`** — the container inspector used throughout Part I (`-files`, `-tree`, `-grep`).
- **`tools/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware), for Parts II–V. New for this project.
- **`tools/cmd/disarm`** — linear ARM/Thumb disassembler.
- **`tools/cmd/codetracearm`** — recursive-descent code-tracer that follows ARM↔Thumb interworking, for tracing the extracted ARM9/ARM7 binaries and overlays.

Game-specific extraction tools live in the `mariokartds/extract` module (`Mario Kart DS (DS)/extract`); rendered figures go in `Mario Kart DS (DS)/rendered/`.
