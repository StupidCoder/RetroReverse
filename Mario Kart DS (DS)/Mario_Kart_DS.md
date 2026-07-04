# Mario Kart DS (Nintendo DS) — cartridge format and game analysis

A reverse-engineering reference for `Mario Kart DS (Europe) (En,Fr,De,Es,It).nds`, the 2005 Nintendo kart racer. This is the repository's first **Nintendo DS** title and its first **ARM** project: where every earlier game runs on a single 8- or 16-bit CPU, the DS is a dual-processor machine — an **ARM9** (ARM946E-S, ARMv5TE, the 67 MHz main CPU) beside an **ARM7** (ARM7TDMI, ARMv4T, the 33 MHz sound/IO CPU) — and its `.nds` cartridge is a real container with a header, two CPU binaries, code overlays and an on-cartridge filesystem, unlike the flat mask-ROM of a Game Boy cartridge or the raw disk of an Amiga floppy. The writeup follows the same shape as the others, in reading order:

* **Part I** — the cartridge image: the `.nds` container, its header, the ARM9/ARM7 binaries and overlays, and the FNT/FAT filesystem that names ~600 asset files;
* **Part II** — the boot chain: the cartridge header's entry points, the secure area, and how the ARM9 and ARM7 come up and hand off;
* **Part III** — program architecture: the runtime memory map, how the game initialises through the OS layer, its interrupt and IPC-FIFO setup, and the ARM9↔ARM7 rendezvous — pinned by running the boot on the `tools/arm` core as an oracle;
* **Part IV** — graphics and data formats: peeling the asset layers (`.carc` → LZ77 → `NARC` archive → NITRO resources), decoding all seven `NSBTX` texture formats, the `NCLR`/`NCGR`/`NSCR` 2D tile pipeline, and the `NSBMD` 3D models (nodes, scene bytecode, GX display lists) — every track texture and UI screen rendered, every menu kart and character exported to GLB;
* **Part V** — game mechanics: the NKM course map (checkpoints and the lap graph, the CPU drive line, item routes, spawns, objects — the full track layout), the engine's object-spawn chain (the NKM loader, its 17 positional section parsers, and the 124-entry map-object descriptor table) and the route-follower that moves objects along their paths, with collision and kart physics ahead;
* **Appendix A** — toolchain and reproduction.

Methods: purely static analysis of the `.nds` image. The DS needed a new toolchain — the shared 6502/68000/Z80/LR35902 decoders do not apply — so this project is built on the new `tools/nds` cartridge reader (header, FNT/FAT, overlays, BLZ decompression) and the new `tools/arm` ARM/Thumb disassembler and CPU core, with `tools/nds/cmd/ndsinfo` for the container catalog, the game's `mariokartds/extract` `ndsextract` to pull the CPU binaries, and `retroreverse.com/tools/cmd/disarm` / `codetracearm` for the code. All addresses are 32-bit ARM addresses (`$02000000`-style main-RAM addresses, or the ARM9/ARM7 BIOS and I/O regions) unless a *file offset* into the ROM image is explicitly called out; bytes are little-endian. **Parts I–IV are essentially complete (IV: every 2D format, every texture and screen, and the `NSBMD` models — karts, characters and all course scenes with their skyboxes — rendered and exported to GLB in the viewer site; `NCER` cells and `SDAT` remain). Part V has begun: the NKM course map — the full track layout — is decoded and figure-verified for all 59 courses; collision and physics are its frontier.**

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
  - [1. Two CPUs, two entry points](#1-two-cpus-two-entry-points)
  - [2. The ARM9 startup stub: CP15, the TCMs and the stacks](#2-the-arm9-startup-stub-cp15-the-tcms-and-the-stacks)
  - [3. The ARM9 self-decompression (BLZ)](#3-the-arm9-self-decompression-blz)
  - [4. Autoload, `.bss`, and the handoff to `main`](#4-autoload-bss-and-the-handoff-to-main)
  - [5. The ARM7 startup](#5-the-arm7-startup)
  - [6. The two processors meet: shared RAM and the IPC FIFO](#6-the-two-processors-meet-shared-ram-and-the-ipc-fifo)
- [Part III — Program architecture](#part-iii--program-architecture)
  - [1. The oracle: running the boot on our own core](#1-the-oracle-running-the-boot-on-our-own-core)
  - [2. The runtime memory map](#2-the-runtime-memory-map)
  - [3. Initialisation: from `crt0` to the framework](#3-initialisation-from-crt0-to-the-framework)
  - [4. Interrupts and the IPC FIFO](#4-interrupts-and-the-ipc-fifo)
  - [5. The ARM9↔ARM7 rendezvous](#5-the-arm9arm7-rendezvous)
  - [6. Overlays and the main loop](#6-overlays-and-the-main-loop)
- [Part IV — Graphics and data formats](#part-iv--graphics-and-data-formats)
  - [1. Peeling the asset layers: LZ77 and NARC](#1-peeling-the-asset-layers-lz77-and-narc)
  - [2. NITRO textures: the `TEX0` block](#2-nitro-textures-the-tex0-block)
  - [3. Texture formats and palettes](#3-texture-formats-and-palettes)
  - [4. The 4x4-compressed format](#4-the-4x4-compressed-format)
  - [5. The 2D tile pipeline: NCLR, NCGR, NSCR](#5-the-2d-tile-pipeline-nclr-ncgr-nscr)
  - [6. NSBMD models: nodes, scene bytecode, display lists](#6-nsbmd-models-nodes-scene-bytecode-display-lists)
  - [7. Course scenes: the track, its far model, and the world scale](#7-course-scenes-the-track-its-far-model-and-the-world-scale)
  - [8. Frontier: sprite cells and sound](#8-frontier-sprite-cells-and-sound)
- [Part V — Game mechanics](#part-v--game-mechanics)
  - [1. The NKM course map: the full track layout](#1-the-nkm-course-map-the-full-track-layout)
  - [2. Frontier: collision, physics, ghosts](#2-frontier-collision-physics-ghosts)
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

Part I read the header the DS BIOS reads; Part II follows what runs after it. Both CPU images are extracted with the game's `ndsextract` (which also decompresses them — see §3), and traced from the header's entry points with `codetracearm`. What comes out is a pair of near-textbook **NitroSDK** startup stubs (the compiler runtime `crt0`): the ARM9's is the interesting one — it decompresses the whole game in place before jumping to `main` — while the ARM7's is a smaller, uncompressed sibling.

## 1. Two CPUs, two entry points

On power-up each CPU runs its on-chip **BIOS**, which validates the header (the logo and header checksums from Part I), copies each CPU's binary to the RAM address the header names, and jumps to its entry point:

| CPU | ROM offset | → RAM load | Entry | Size (stored) |
|---|---|---|---|---:|
| ARM9 | `$004000` | `$02000000` | `$02000800` | `$0E751C` (compressed) |
| ARM7 | `$141600` | `$02380000` | `$02380000` | `$0289A4` |

The ARM9 binary's first 16 KiB (`$02000000`–`$02003FFF`) is the **secure area** — on a real cartridge it is KEY1/Blowfish-encrypted and the BIOS decrypts it during a challenge/response handshake with the cartridge (the `port 40001A4h` settings and secure-CRC `$DE47` from the header drive that). In this dump the secure area is already decrypted: its first eight bytes read `FF DE FF E7  FF DE FF E7` — the two `$E7FFDEFF` marker words the SDK writes over the encrypted "secure-area ID" once decryption succeeds — so the startup code below is plain ARM. The ARM9 entry sits at `$02000800`, `$800` past the load base, stepping over that ID/marker region.

The ARM9 comes up first; it is the CPU that owns the cartridge and main RAM, and (§6) releases the ARM7.

## 2. The ARM9 startup stub: CP15, the TCMs and the stacks

The entry does two things immediately — kill interrupts, then configure the CP15 system-control coprocessor — before it can even use a stack:

```
02000800  MOV  r12, #0x04000000        ; I/O base
02000804  STR  r12, [r12, #0x208]      ; IME = 0x04000000 (bit0 clear) → interrupts off
02000808  BL   0x02000A5C              ; CP15 / MPU / TCM setup
```

`sub_02000A5C` is the CP15 sequence (all `MRC`/`MCR p15` — the ARM9's caches, MPU regions and tightly-coupled memories):

```
02000A5C  MRC  p15, 0, r0, c1, c0, 0   ; read control register
02000A64  BIC  r0, r0, r1              ; clear cache/MPU enables
02000A68  MCR  p15, 0, r0, c1, c0, 0
02000A70  MCR  p15, 0, r0, c7, c5, 0   ; invalidate I-cache
02000A74  MCR  p15, 0, r0, c7, c6, 0   ; invalidate D-cache
02000A78  MCR  p15, 0, r0, c7, c10, 4  ; drain write buffer
02000A80  MCR  p15, 0, r0, c6, c0, 0   ; MPU protection region 0
02000A88  MCR  p15, 0, r0, c6, c1, 0   ; MPU protection region 1  …
```

It programs the MPU protection regions and enables the two **tightly-coupled memories** the ARM9 relies on: **ITCM** at `$01FF8000` (fast code, just below main RAM) and **DTCM** at `$027E0000` (fast data — and, crucially, the stacks). With the TCMs live, the stub sets up a stack per CPU mode by switching mode with `MSR CPSR_c` and pointing `sp` into the top of DTCM:

```
0200080C  MOV r0,#0x13 ; MSR CPSR_c,r0   ; Supervisor mode
02000814  LDR r0,=0x027E0000 ; ADD r0,#0x3FC0 ; MOV sp,r0   ; SVC stack at DTCM+0x3FC0
02000820  MOV r0,#0x12 ; MSR CPSR_c,r0   ; IRQ mode  → its own stack
02000840  MOV r0,#0x1F ; MSR CPSR_c,r0   ; System mode (the mode main runs in)
```

DTCM (`$027E0000`, 16 KiB) holds all three exception stacks, packed down from its top (`+$3FC0`). The word-fill helper `sub_02000920` (`ADD r12,r1,r2; STMLTIA r1!,{r0}; …`) is then used to clear/seed a few regions before the heavy lifting.

## 3. The ARM9 self-decompression (BLZ)

The header's "ARM9 size" (`$0E751C`) is a *compressed* size: the bulk of the ARM9 image is packed with Nintendo's **BLZ** — an LZSS variant that decodes **backward**, expanding the image in place. The stub decompresses itself:

```
0200087C  LDR  r1, =0x02000B4C         ; the NitroSDK "module params" struct
02000880  LDR  r0, [r1, #0x14]         ; r0 = compressedStaticEnd = 0x020E751C
02000884  BL   0x02000934              ; MIi_UncompressBackward (BLZ)
```

`sub_02000934` is the BLZ decompressor itself — it reads the footer at the end of the compressed image and decodes downward:

```
02000940  LDMDB r0, {r1, r2}           ; r1 = footer[-8] (enc_len | hdr<<24), r2 = footer[-4] (inc_len)
02000944  ADD   r2, r0, r2             ; r2 = decompressed end
02000948  SUB   r3, r0, r1, LSR #24    ; r3 = start of control stream (skip hdr_len footer)
0200094C  BIC   r1, r1, #0xFF000000    ; r1 = enc_len
02000950  SUB   r1, r0, r1             ; r1 = start of compressed region
02000960  LDRB  r5, [r3, #-1]!         ; …read flag/data bytes backward, copy back-refs…
```

The **module-params** struct at `$02000B4C` is the map the runtime uses; decoded:

```
$02000B4C +0x00  0x021773C0   autoload list start
          +0x04  0x021773D8   autoload list end        (= end of the decompressed ARM9)
          +0x08  0x0216F340   autoload data start
          +0x0C  0x0216F340   static .bss start
          +0x10  0x021804E0   static .bss end          (= the overlay load base, Part I §4)
          +0x14  0x020E751C   compressed-static end     → passed to the BLZ decompressor
          +0x1C  0xDEC00621   NitroSDK signature
```

The footer values (`enc_len $E351C`, `hdr_len $08`, `inc_len $8FEBC`) put the compressed region at ARM9 offset `$4000` upward (the secure-area/startup stub below it stays verbatim), and give a decompressed size of `$E751C + $8FEBC = $1773D8`. The ARM9 therefore expands from `$02000000`–`$020E751C` to `$02000000`–`$021773D8`, which is exactly why the far calls that follow (e.g. `BL 0x02133DA8`) land in valid code only *after* this step.

BLZ is reimplemented from scratch in Go (`tools/nds`, `DecompressBLZ`) — reversing the compressed stream, running a forward LZSS, and reversing the result. Verified per the decode-reimplement rule: the output is exactly `$1773D8` bytes, its far-call targets disassemble as coherent ARM code, and the tiny 20-byte fourth overlay round-trips to its 32 known bytes (a unit-test golden vector). The game never ships decompressed; `ndsextract` regenerates `arm9_dec.bin` on demand.

## 4. Autoload, `.bss`, and the handoff to `main`

With the image expanded, the stub finishes the C runtime:

- **Autoload** (`sub_020009E0`) walks the `$021773C0`–`$021773D8` list, copying each autoload block from its packed location (`$0216F340`) to its run address — this is how the ITCM/DTCM-resident code and initialised data reach their fast memories.
- **`.bss` clear** zeroes `$0216F340`–`$021804E0` (`STRCC r0,[r1],#4` loop) — the static zero-initialised data, ending exactly at the overlay base.
- a **cache clean/invalidate** loop (`MCR p15, c7, …` over 32-byte lines) makes the freshly-written code visible to the instruction side.
- two far initialisers run (`BL 0x02133DA8`, `BL 0x02142540` — SDK/OS init in the now-decompressed image).

Then the handoff, `main` and the exit vector loaded from the literal pool:

```
020008F0  LDR r1, =0x02003000          ; main
020008F4  LDR lr, =0xFFFF0000          ; return address = ARM9 BIOS (halt on return)
020008F8  BX  r1
```

`main` at `$02003000` is a thin wrapper — `STMDB sp!,{lr}; BL 0x020365F0; …; BL 0x020365BC; BX lr` — so the real game entry is `$020365F0`, the subject of Part III. If it ever returns, `lr = $FFFF0000` drops the ARM9 into its BIOS.

## 5. The ARM7 startup

The ARM7 binary is **not** compressed (its body disassembles coherently end to end; there is no module-params decompression call), so its `crt0` is shorter. From entry `$02380000`:

```
02380000  MOV r12,#0x04000000 ; STR r12,[r12,#0x208]   ; IME = 0 (interrupts off)
02380008  LDR r1,=…  ; MOV r0,#0x03800000 ; CMP;MOVPL   ; clear range in ARM7-private WRAM
02380020  …STMLTIA r1!,{r0}…                            ; zero ARM7 WRAM
0238002C  MOV r0,#0x13 ; MSR CPSR_c,r0 ; LDR sp,=0x0380FF00   ; SVC stack (ARM7 WRAM top)
02380038  MOV r0,#0x12 ; MSR CPSR_c,r0 ; …                    ; IRQ stack (0x0380FFC0)
02380050  MOV r0,#0x1F ; MSR CPSR_c,r0 ; …                    ; System stack (0x0380FF80)
02380068  …LDR r3,[r0],#4 ; STR r3,[r1],#4…                   ; copy IRQ vectors / fixed sections
02380090  BL  0x02380100                                      ; autoload
02380094  …STRCC r0,[r1],#4…                                  ; .bss clear
023800C0  LDR r1,=0x037F8534 ; LDR lr,=0xFFFF0000 ; BX r1     ; jump to ARM7 main
```

The ARM7's stacks live at the top of its private 64 KiB WRAM (`$03800000`–`$0380FFFF`). Its autoload (`sub_02380100`) copies its resident code up into shared WRAM, and — notably — the entry point it finally branches to, `$037F8534`, is *in WRAM*, not in the `$02380000` load region: the ARM7 relocates its hot code to WRAM and runs it there. As with the ARM9, `lr = $FFFF0000` returns to the ARM7 BIOS.

## 6. The two processors meet: shared RAM and the IPC FIFO

Both binaries load into the same 4 MiB main RAM (`$02000000`; ARM9 at the base, ARM7 near the top at `$02380000`), and both reach `main` independently after the sequences above. From there they coordinate through the DS's **inter-processor communication** hardware in the `$04000000` I/O block — the `IPCSYNC` register (`$04000180`, a 4-bit mailbox each way plus an IRQ line) and the `IPCFIFO` (control at `$04000184`, the 64-byte send/receive FIFO at `$04100000`) — with main RAM used for the bulk payloads (input state, sound commands, DMA lists). The ARM9, owning the cartridge, streams the filesystem (Part I) and drives the game; the ARM7 services sound (the single `SDAT` bank), the touchscreen, buttons and wireless, reporting back over the FIFO.

Part III §4–§5 picks this up: running the ARM9 boot on the `tools/arm` core as an oracle pins the exact IPCSYNC handshake by which the ARM9 blocks waiting for the ARM7. The full bidirectional command protocol still needs a dual-core oracle and remains a frontier.

# Part III — Program architecture

Part II followed the boot to the game's `main`; Part III is about the machine that `main` builds — the runtime memory layout, the OS/interrupt scaffolding, and the point where the ARM9 first has to talk to the ARM7. A retail DS game is a large C++ program that dispatches almost everything through function pointers and virtual tables, so a purely static trace fans out into hundreds of unreachable indirect calls. So this Part leans on the repository's standard technique: an **oracle** — the game's own code run on our own CPU core — to establish, from behaviour, the structure a static read can only guess at.

## 1. The oracle: running the boot on our own core

`mariokartds/extract`'s `bootoracle` loads the *compressed* ARM9 binary to `$02000000` exactly as the BIOS would, points the `tools/arm` core at the entry `$02000800`, and lets it run: a flat memory for RAM/TCM/WRAM, the handful of BIOS `SWI`s the startup needs (`CpuSet`/`CpuFastSet` memory moves, `WaitByLoop`), CP15 accepted and ignored, and every write to the `$04000000` I/O block logged. It runs the real startup — no shortcuts — and reports what the code *did*.

It executes ≈12.3 million instructions to reach `main` (`$02003000`), then the game init (`$020365F0`), and halts cleanly at the first thing a lone ARM9 cannot get past (§5). Two things fall straight out of that run, each a verification rather than a guess:

- **The BLZ decompression is confirmed by the game itself.** After boot, the bytes the game's own `crt0` decompressor wrote into `$02000000…$0216F340` are **identical** to what `tools/nds`' independent `DecompressBLZ` produces (`bootoracle` diffs them) — divergence begins only at `$0216F340`, exactly `.bss` start, which the `crt0` then zero-fills. The Part II reimplementation and the real decompressor agree bit-for-bit over all code and data.
- **The startup reaches the framework on our core** — the decoder, the CPU core, the mode/bank handling and the memory model are correct enough to run millions of instructions of real ARM9 code, including the self-decompression, autoload and OS init, without a wrong turn.

This is the oracle in the repository's sense: real code, our core, used to *confirm* structure — never to scrape decoded data out of RAM.

## 2. The runtime memory map

Combining the module-params/overlay tables (Parts I–II) with what the oracle observes the boot populate, the ARM9's runtime layout of the 4 MiB main RAM (`$02000000`–`$023FFFFF`) plus its tightly-coupled memories is:

| Region | Range | Contents |
|---|---|---|
| ITCM | `$01FF8000`–`$01FFFFFF` | fast code, filled from the autoload list |
| ARM9 static code+data | `$02000000`–`$0216F340` | the decompressed ARM9 (entry/`crt0`, `main` `$02003000`, game init `$020365F0`) |
| ARM9 `.bss` | `$0216F340`–`$021804E0` | zero-initialised statics (module-params `+0x0C`/`+0x10`) |
| ARM9 overlays | `$021804E0`… / `$021B7BA0`… | the four §I.4 overlays, banked in per game mode |
| ARM9 heap | above the overlays | dynamic allocation, up to the ARM7 region |
| ARM7 binary | `$02380000`–`$023A89A4` | the ARM7 image (its hot code runs from WRAM, Part II §5) |
| DTCM | `$027E0000`–`$027E3FFF` | fast data + the SVC/IRQ/System stacks (top-down from `+$3FC0`) |
| system-reserved | `$027FF000`–`$027FFFFF` | the ARM9/ARM7 shared config block (read via `$0200F168`) |

The two CPUs partition the one main memory by convention — the ARM9 owns the low end and grows a heap upward; the ARM7 sits at `$02380000` near the top — and each keeps its hot working set in its private TCM/WRAM.

## 3. Initialisation: from `crt0` to the framework

`main` (`$02003000`) is a two-call stub (Part II §4); the real work is the game init at **`$020365F0`**, which the oracle confirms is where control settles after startup. Statically it is a compact sequence of framework brings-up:

```
020365F0  STMDB sp!,{r4,lr}
020365F4  BL 0x020366BC        ; OS / system init (BL 0x0200EF10, 0x0200EEC0, …)
020365F8  BL 0x0203907C        ; a hardware register block (STRH bursts)
0203660C  BL 0x0200F168        ; read OS tick / system-config word ($027FFxxx)
02036638  BL 0x02036C4C        ; …
0203665C  BL 0x020394CC        ; a second register block init
```

The `$0200Exxx`/`$0200Fxxx` range is the game's copy of the **NitroSDK OS layer** (tick/timer readers like `$0200F168`, which computes `$027FF000 + (id<<2) + $DA0` and loads the shared config). `$020366BC` walks a list of subsystem initialisers; `$0203907C` and `$020394CC` each program a contiguous block of 16-bit hardware registers (`STRH` to `[base+0]`, `[base+2]`, …) — the 2D-graphics-engine and DMA/timer register banks. The oracle watches this init drive the machine into the state §4 describes and then reach for the ARM7.

## 4. Interrupts and the IPC FIFO

The single most useful thing the oracle extracts is the *exact* set of hardware registers the boot programs before it needs the second CPU — only five, and every one is about interrupts and inter-processor communication:

```
0x04000180 = 0x00000000   IPCSYNC     — clear our sync nibble / enable IPC-sync IRQ
0x04000184 = 0x0000C408   IPCFIFOCNT  — enable the IPC FIFO (send-clear, error-ack, enable)
0x04000208 = 0x04000000   IME         — master interrupt enable register touched
0x04000210 = 0x00040000   IE          — enable exactly bit 18: IPC recv-FIFO-not-empty
0x04000214 = 0x00040000   IF          — acknowledge that bit
```

The ARM9 enables **one** interrupt source before anything else: **bit 18, "IPC receive FIFO not empty."** Its entire early interrupt architecture exists to hear the ARM7. The DS interrupt model underneath is standard: `IME` (`$04000208`) is the master switch, `IE` (`$04000210`) the per-source enable mask, `IF` (`$04000214`) the request/acknowledge latch (write-1-to-clear); the ARM9 BIOS vectors an IRQ through a handler pointer the runtime installs in DTCM (`$027E3FFC`) after saving state. The `IPCFIFOCNT = $C408` write turns on the 64-bit-wide hardware FIFO (data port at `$04100000`) with its send-empty/receive-not-empty IRQ wiring; `IPCSYNC` (`$04000180`) is the 4-bits-each-way mailbox the two CPUs use to hand-shake before the FIFO carries real traffic.

## 5. The ARM9↔ARM7 rendezvous

Having enabled the IPC receive interrupt, the game init calls into an OS routine that **blocks on `IPCSYNC` waiting for the ARM7** — and the oracle stops there, because with no ARM7 the condition never comes true. The spin loop it lands in is unmistakable:

```
0214D8D0  LDRH r0, [r3]        ; r3 = IPCSYNC ($04000180)
0214D8D8  AND  r0, r0, #0x0F   ; the 4-bit value the *other* CPU posted
0214D8DC  CMP  r0, lr          ; == the expected step value?
0214D8E0  BNE  …               ; …keep polling…
0214D8F0  LDRH r0, [r3]        ; re-read
0214D8F8  AND  r0, r0, #0x0F
0214D8FC  CMP  r0, lr
```

This is `OS_SyncWithOtherProc`-style handshaking: each CPU writes a step number into its outgoing `IPCSYNC` nibble and spins until it reads the matching number from the other's incoming nibble, ratcheting through a short sequence so that neither races ahead of the other during early boot. The ARM9 has done its half (the `IPCSYNC = 0` write above) and now waits for the ARM7 — which, per Part II §5, is itself booting from `$02380000` and relocating into WRAM.

So the boot's two halves meet here: the ARM9, having set up memory, the OS layer and the IPC interrupt, parks on `IPCSYNC` until the ARM7 answers; only then do the two `main` routines begin exchanging FIFO commands (input and touch state up from the ARM7; sound and DMA requests down from the ARM9). Pinning this point resolves the Part II frontier — the *location and mechanism* of the rendezvous — using the oracle.

## 6. Overlays and the main loop

Past the rendezvous the ARM9 enters its frame loop and starts streaming the §I.4 overlays. The overlay mechanism itself is already established: each overlay is a BLZ-compressed blob (Part II's `DecompressBLZ` handles them — `ndsextract` decompresses all four) named by a FAT id, loaded to a fixed RAM address (`$021804E0`, or `$021B7BA0` for overlay 3) with its `.bss` zeroed and its static-initialiser list run, so switching game mode is: card-DMA the overlay's file in, decompress, clear `.bss`, run constructors, jump. What is *not* yet pinned is which overlay backs which mode, and the frame loop's VBlank cadence.

*(frontier)* Both the frame/VBlank main loop and the overlay-to-mode mapping live **past** the IPC rendezvous, so reaching them on the oracle needs the ARM7 running too — a **dual-core oracle** (ARM9 + ARM7 sharing main RAM, stepping in lockstep through the IPCSYNC/FIFO exchange). Standing that up, and then tracing the frame loop and the per-mode overlay loads, is the next step, and the bridge into Part IV's asset decoding.

# Part IV — Graphics and data formats

Part I catalogued the filesystem and noted (the "separate the layers" rule) that an asset path is several nested formats. Part IV peels them: the storage layers first, then the NITRO texture format, decoded to actual pixels. Unlike Part III this needs no oracle — the assets are inert data, decoded statically straight from the cartridge.

## 1. Peeling the asset layers: LZ77 and NARC

A `.carc` file is two layers over the real content, both reimplemented from scratch in `tools/nds` and verified against the image:

- **LZ77** — Nintendo's standard *forward* LZSS (distinct from the *backward* BLZ the boot code uses, Part II §3). A 4-byte header (`0x10`/`0x11` type + 24-bit decompressed size) precedes flag-driven blocks of literals and back-references. `DecompressLZ77` decompresses `data/CharacterKartSelect.carc` (`0x10`, declared size `$850C`) to exactly `$850C` bytes beginning `NARC`. ✓
- **NARC** (Nintendo ARChive) — the bundle underneath, the same idea as the cartridge filesystem in miniature: a `BTAF` file-allocation table (start/end offsets), a usually-flat `BTNF` name table, and a `GMIF` blob of file data. `ParseNARC` (which LZ77-decompresses first, so a raw `.carc` is accepted directly) splits `beach_courseTex.carc` into its 11 sub-files, and `mario_course.carc` into 21.

That single command reveals the engine's whole asset vocabulary. A course's `…Tex.carc` holds `BTX0` texture sets plus `NCGR`/`NCLR`/`NSCR` 2D graphics; a `…course.carc` holds `BMD0` 3D geometry, `BTA0`/`BTP0` texture animation, and **`NKMD`** — "Nitro Kart Map Data", the collision/route data that is Part V's target.

## 2. NITRO textures: the `TEX0` block

Textures ship as an NSBTX file: a NITRO `BTX0` container wrapping a single `TEX0` block (standalone for the character/kart menu emblems, or packed in a course's texture `NARC`). A `TEX0` holds four things located by offsets in its header — the texel data, the palette-colour data, and two **NITRO resource dictionaries** naming the textures and the palettes:

```
TEX0 + $0E  u16  texture-dictionary offset
TEX0 + $14  u32  texel-data offset
TEX0 + $34  u32  palette-dictionary offset
TEX0 + $38  u32  palette-colour-data offset
```

The **resource dictionary** is the reusable NITRO structure (shared with models, animations, everything): a 4-byte header (revision, entry count, size), a Patricia-tree block of `4+(count+1)*4` bytes for name lookup (skippable for a linear read), a 4-byte data-block header giving the per-entry unit size, `count` fixed-size entries, and finally `count` 16-byte names at `dict + size − count*16`. Each texture entry's 32-bit `texImageParam` packs everything needed to decode it:

```
DK_emblem "emblem": param = $2D200000
  bits 0-15   texel offset (>>3)      = 0
  bits 20-22  width  = 8 << 2         = 32
  bits 23-25  height = 8 << 2         = 32
  bits 26-28  format = 3              = 16-colour (4bpp)
  bit  29     colour 0 transparent    = yes
```

## 3. Texture formats and palettes

Palette colours are 15-bit **BGR555** words (`tools/nds/nitro` expands each 5-bit channel to 8-bit); colour index 0 is transparent when the texture's flag says so. All seven DS texture formats are decoded: 2 (4-colour, 2bpp), 3 (16-colour, 4bpp), 4 (256-colour, 8bpp), 1 (**A3I5**: 3-bit alpha + 5-bit index) and 6 (**A5I3**: 5-bit alpha + 3-bit index) for translucency, 7 (direct 16bpp), and 5 (4x4-compressed, §4). Two facts had to be pinned empirically against real sets:

- the palette dictionary's offset word is in **8-byte units** (`<<3`, like the `TEX0` size fields) — proved by the last palettes of a full course set landing exactly inside the palette region, where a `<<4` read overshoots it;
- textures and palettes are named *and ordered independently*, and the naming is erratic — `emblem`↔`emblem_pl`, but also `nr_road5`↔`road5_pl` (prefix dropped), `nr_dash_02`↔`nr_dash2`, `nr_start_line2`↔`nr_line_pl` — so `DecodeNSBTX` pairs them by **name similarity** (exact → containment → common suffix/substring). The authoritative binding lives in the material block of the `NSBMD` model that consumes the texture (§6); a texture set alone doesn't carry it.

## 4. The 4x4-compressed format

Format 5, the DS's only compressed texture format, carries most course art (a course's `…Tex.carc` is typically 80% format 5). The texture is a grid of **4x4-pixel blocks**: per block, one 32-bit word of sixteen 2-bit values, plus one 16-bit palette word — bits 0–13 a sub-palette offset (4-byte steps, relative to the texture's palette), bits 14–15 a mode: the four 2-bit values select `{c0, c1, c2, transparent}`, `{c0, c1, (c0+c1)/2, transparent}`, `{c0, c1, c2, c3}` or `{c0, c1, (5·c0+3·c1)/8, (3·c0+5·c1)/8}`.

The subtlety that initially broke the decode is *where* the two block streams live: format-5 texel words go in a **dedicated region** at `TEX0+$24` (not the ordinary texel region), and the palette words in a third region at `TEX0+$28`, indexed at **half** the texture's address. The proof is in the region arithmetic of a real file — in `beach_courseTex` the four regions tile the block exactly, back to back:

```
$0404  ordinary texels   ($0F00 bytes = size field $01E0 << 3)
$1304  4x4 texel words   ($5BC0 bytes = size field $0B78 << 3)
$6EC4  4x4 palette words ($2DE0 bytes — exactly half the texel words)
$9CA4  palette colours   ($4300 bytes = size field $0860 << 3)  → end of block ✓
```

With that layout every format-5 texture in the game decodes cleanly — Rainbow Road's rainbow-gradient roadway, Cheep Cheep Beach's palm-jungle skyline, Mario Circuit's trackside billboards (`rendered/course/`).

## 5. The 2D tile pipeline: NCLR, NCGR, NSCR

The menus, HUD and every flat screen use the DS's 2D tile engines, and ship in three NITRO files that mirror the hardware's three memories — decoded in `tools/nds/nitro/g2d.go`:

- **`NCLR`** (`RLCN`, block `TTLP`; the SDK also emits an `RPCN` variant, seen on the debug font) — palette RAM: BGR555 colours, 4bpp (16×16) or 8bpp (256);
- **`NCGR`** (`RGCN`, block `RAHC`) — tile VRAM: 8x8-pixel characters at 4bpp or 8bpp (sprite-destined ones mark their dimensions `$FFFF` — "scattered", laid out by a cell file, §6);
- **`NSCR`** (`RCSN`, block `NRCS`) — map VRAM: 16-bit entries packing tile number (bits 0–9), H/V flip (10–11) and a palette row (12–15); index 0 is the transparent backdrop.

Composing screen→tiles→palette reproduces what the 2D engine displays. Two game conventions matter: names suffixed **`_b`** are the background bank and **`_o`**/`.nce` the sprite bank (a screen must pair with `_b` — the composer prefers it); and a scene's screens may sit in the *base* archive while their tiles ship in the **language variants** (`Title.carc` holds `title_m1_EU.NSCR`; the 8bpp title-logo tiles are in `Title_us.carc`…`_it.carc`), so composition merges each language archive with its base — the same join the game performs at load time.

`extract/cmd/renderall` sweeps the whole filesystem through every decoder above and renders **every texture and every screen in the game** into `rendered/` — 1,360 textures (all 45+ course sets and every kart/character model texture), 596 composed screens and 376 tile sheets, zero archives skipped. Among the verified figures: the boot **Nintendo logo**, the **MARIO KART DS title screen**, all 32 **cup-select course previews**, the race HUD and menu backdrops, the full **debug font** (`data/Boot/dbgfont`), and the 32×32 cartridge **banner icon** — Shy Guy in a kart — from the raw `.nbfc`/`.nbfp` pair. (`data/Boot/builddate.bin`, incidentally, dates the build: *"Build: 2005 10/8 (Sat) 23:05:54"*.)

## 6. NSBMD models: nodes, scene bytecode, display lists

An `NSBMD` (`BMD0`, block `MDL0`) is a small named *scene*, not just a mesh, decoded in `tools/nds/nitro/model.go` + `displaylist.go` and pinned structure-by-structure against the game's own files (the 476-byte shadow-quad model was the Rosetta stone — small enough to hand-verify every field):

- the **model header** carries section offsets (`+$04` SBC, `+$08` materials, `+$0C` shapes) and counts; then a resource dictionary of **nodes** (joints), each a packed TRS record — a flags word (bit 0 no-T, 1 no-R, 2 no-S, 3 *pivot*), fx32 translation, and either a full fx16 3x3 or a **pivot-compressed rotation**: one matrix element is ±1 (its index in flags bits 4–7, sign bit 8) and the 2x2 remainder is `{A,B;±B,±A}` (bits 9–10) from just two fx16 values;
- the **SBC scene bytecode** walks the scene: `NODEDESC` (node × parent → the joint's world matrix, optionally stored to a GX matrix-stack slot — opcode bits `$20/$40` add store/restore operands), `MTX` (restore a slot), `MAT` (bind material), `SHP` (draw shape). Mario's menu model is seven `NODEDESC`s (root→body→arms→head) then four draw commands;
- the **material section**'s tex/pal-to-material lists give the *authoritative* texture↔palette binding (`kart_body`←`kart_MR_a`+`kart_MR_a_pl`, …) that §3's name heuristics approximate — each list entry points (relative to the material section) at the index bytes of the materials using that texture. The material record carries the GX `TEXIMAGE_PARAM` (`+$14`: repeat/flip addressing) and the texture size (`+$20`);
- a **shape** is `{u16 itemTag, u16 recSize, u32 flags, u32 dlOffset (record-relative), u32 dlSize}` → a raw **display list**: the DS geometry engine's own command stream, four packed command IDs per word. The decoder executes it: `MTX_RESTORE` (joint binding), `COLOR`, `TEXCOORD` (12.4 texel units), and the vertex forms `VTX_16`, `VTX_10`, `VTX_XY/XZ/YZ` (two coords + one kept) and `VTX_DIFF` (10-bit deltas in *raw fx12* units — ±0.125; the scale that, wrong, explodes every character model into spikes). Primitives 0–3 are triangles/quads/tri-strips/quad-strips; a **quad strip's** vertices arrive `v0 v1 v2 v3…` but each quad's spatial cycle is `v0→v1→v3→v2` (the newest pair swaps), so splitting on the wrong diagonal turns every lathed surface — the tires — into bowties;
- a material whose record is longer than the base `$2C` bytes carries a **texture-SRT matrix**, flagged in the `+$1C` flag word and applied when `TEXIMAGE_PARAM` bits 30–31 select the texcoord-source transform: `+$2C`/`+$30` are fx32 **scaleS/scaleT**. The kart wheels are the textbook case — their 32×32 texture is half tread pattern, half hub face, and the tread strips' texcoords only land on the tread half after the material's `scale(2, 1)`; without it the hub face smears around the tyre.

Verified by rendering (`extract/cmd/rendermodel`, a z-buffered software rasteriser): Mario's **B-Dasher** with the "M" emblem on the hood and gold-hubbed tires, the **Mario menu character** (cap logo, moustache, gloves), Yoshi, Bowser's kart with his emblem on the tail — all 12 characters and all 29 menu karts, `rendered/models/`. Sanity anchor: the shadow model decodes to exactly the 6 quads its header declares.

**GLB export.** `nitro.ExportGLB` serialises any decoded model as a standard **binary glTF 2.0** — per-material primitives with positions, normalised UVs and vertex colours, the NSBTX textures embedded as PNGs, GX repeat/flip mapped to glTF sampler wrap modes. `extract/cmd/exportglb -all` emits all 41 menu models plus a `models.json` manifest, and the set is published on the repository's viewer site: **Mario Kart DS is in the Studio** (`site/`, "Nintendo DS" system) with a three.js GLB viewer (`site/src/mariokart/viewer.js`) — orbit controls, nearest-filtered textures, characters and karts selectable per section.

## 7. Course scenes: the track, its far model, and the world scale

A course archive (`data/Course/<name>.carc`, nameless NARC) is the whole racing scene. Decoded for Mario Circuit:

| Sub-file | Type | Contents |
|---|---|---|
| 0 | `BMD0` | **`mario_course`** — the drivable course (1,625 tris, 22 materials/shapes, `POSSCALE` 64) |
| 1 | `BMD0` | **`mario_course_V`** — the *skybox* (24 tris): the camera-relative backdrop panorama |
| 2 | `BTA0` | texture animation (the water) |
| 3 | *(raw)* | **collision** — vertex/normal/prism/octree sections at `$3C/$FC0/$596C/$9DCC` (Part V) |
| 4, 19, 20 | `NKMD` | the **course maps** — single-player, plus variants (Part V §1) |
| 5–18 | `BMD0`+anims | the **map objects**: `MarioTree3`, `kuribo` (Goomba), the `Pakkun*` Piranha-Plant parts, `FireBall`, `water_efct`, with `BTP0`/`BCA0`/`BMA0` animations |

So the segmentation is real, but not LOD-per-segment: the course body is **one model** whose 20-odd shapes are its material batches (road, grass, water, castle walls…), and the `_V` companion is not a distance LOD at all but the course's **skybox** — the panorama drawn around the world, positioned relative to the camera so it never parallaxes: `mario_course_V` is a 24-triangle backdrop to the main model's 1,625, `rainbow_course_V` a 146-triangle starfield dome (at its own `POSSCALE` 128, matching its camera-space scale). Multiplayer gets a simplification pass one level up: `beach_courseD`/`cross_courseD`/`mansion_courseD` are whole reduced *course archives* for the DS's download-play/multiplayer mode.

Three findings pin the geometry to the world:

- **`POSSCALE`**: course vertices are stored divided by a power-of-two scale (`$40` for Mario Circuit — the fx16 vertex range is only ±8) and the SBC re-applies it (`$0B` between `NODEDESC` and the draws) — implemented in `RunSBC`.
- **The ×16 world scale**: even after `POSSCALE`, the model spans ±250 units while the course-map data (§V.1) spans ±2,900 — the render model lives at **1/16 of kart-world scale**. Overlaying the two at ×16 puts the CPU drive line exactly on the asphalt (the `rendered/tracks/` figures), which is the proof.
- **Textures cross archives**: the course model's textures live in the sibling `<name>Tex.carc` (two `BTX0` sets — the main model's and the `_V` model's), resolved through the model's own material bindings.

And a cartridge archaeology bonus: `data/Course` still contains **unused development courses** — `donkey_course`, `luigi_course`, `nokonoko_course`, `dokan_course`, `test1_course`, `test_circle`, plus `wario_course` (an earlier Wario Stadium with only 3 placed objects) and the `StaffRoll` credits fly-through — all complete with geometry and course maps, all decoded by the same pipeline.

All course scenes are exported to GLB alongside the karts (`exportglb -all` emits **377 models**: characters, karts, and every course — the nitro tracks, the **16 retro tracks** brought forward from the SNES/N64/GBA/GameCube games plus their multiplayer variants, and the battle/mission stages — each with its `_V` skybox **and its map-object models**, the archive-mates the engine spawns from the NKM's `OBJI` section: Chain Chomps, Goombas, Piantas, bridges, pipes, per-theme trees…). In the Studio each track is its own section holding the track plus its objects. The retro archives name their scene after the archive (`old_baby_gc`, not `*_course`), so the exporter keys off the archive stem to pick the main model and its `_V` companion.

A format detail the objects forced out: a `BMD0` may **embed its own `TEX0` block** after the `MDL0`. The course's road textures live in the sibling `<name>Tex.carc`, but the self-contained map objects (the Chain Chomp's `ob_wan_body`, the Goomba, the crates) carry their textures inside their own model file — `nitro.DecodeContainerTextures` decodes the `TEX0` of any NITRO container, and with it every material binding across all 59 course archives resolves (verified: 0 missing). Several objects turn out to be 2-triangle **camera-facing billboards** (the Goomba, the Pianta) — sprites in the 3-D world, exactly as the DS renders them.

**And the objects are placed.** `exportglb` also writes `<course>.objects.json`: every `OBJI` placement whose object ID resolves to an exported model — the binding comes from the ARM9's own descriptor table (Part V §2, via `mkds.ObjectModelBindings`), position converted to GLB space, rotation kept as the entry's Y-dominant Euler degrees (the two Delfino drawbridges sit at ±90°), per-entry scale (the Delfino trees ship non-uniform, e.g. 0.9/1.0/0.9). The itembox (object `0x65`) is the one placed model living outside the course archives — exported once from `data/Main/MapObj.carc` and shared by every track. Across the cartridge ~1,190 placements resolve; the ~370 skipped are logic objects with no model (their descriptors name no resources — sound emitters, effect points) and multi-part composites the engine assembles itself (the snowman's `sman_top`+`sman_bottom`, the Chain Chomp's body+chain+stake). In the Studio the placements load with the track behind an *Objects* toggle, and **billboard placements yaw around world-up toward the camera each frame** — their up axis stays world-aligned, matching how the DS draws them.

A placement subtlety, traced in code: the item boxes are authored at **road height** — it is the itembox *init* (`$020DE6B0`, first thing it does) that adds the fx32 constant **12.0 world units** (`0xC000`) to Y, which is why they float. The spawn transform is per-type: each object's init callback may adjust it, and ground objects place as-authored. The exporter bakes the traced +12.0 into the itembox placements.

**Route-following objects move.** Placements whose `OBJI` entry names a route carry their `PATH`/`POIT` polyline in the JSON (36 across the cartridge: the Airship Fortress Bullet Bills and burners, the beach crabs, the desert Pokeys — and Desert Hills' *sun*, which circles the track on a route — the mansion's walking trees, Waluigi Pinball's iron balls). The viewer walks them along the polyline at constant world speed with the engine's follower semantics (Part V §3): wrap on looped paths, out-and-back on open ones, facing the travel direction. The per-type *speeds* live in engine code not yet traced, so a shared plausible speed stands in.

**BTA0 texture animation, decoded** (`tools/nds/nitro/anim.go`, `DecodeNSBTA`). The `BTA0` sub-files of the course archives are material **texture-SRT animations** — the scrolling water, waterfalls and boost-panel arrows. One `SRT0` block holds a dict of animations (named for their target model); each is an `"M\0AT"` record `{u16 numFrames, u16 flags}` plus a dict of materials whose 0x28-byte entries are five 8-byte component records inline — `{u16 lastFrame, u8 0, u8 flags, u32 v}` for *scaleS, scaleT, rotation, transS, transT*. Component flag bit 7 selects a sampled track: `v` becomes the offset (relative to the `M\0AT` record) of u16 fx12 samples taken **every 4th frame**; otherwise `v` is an fx12 constant (rotation packs a `(sin,cos)` pair). The pin that proves the layout: the beach dash panel's last sample is `0xFBD` = 4096×60/61 *exactly* — one full texture wrap per 61-frame loop. All **47 course BTA0s decode** (81 material animations, 39 with sampled tracks; Rainbow Road alone has 11). `exportglb` emits them as `<course>.anims.json` and the viewer replays them at 60 fps on the matching GLB materials — the drawbridge's `Bdash` **boost-panel arrows scroll**, Yoshi Falls pours, and the course waters ripple as shipped.

Two extras ride alongside each course GLB for the viewer:

- **Skybox in the round.** glTF has no notion of a skybox, so the `_V` backdrop stays an ordinary GLB; the viewer (`site/src/mariokart/viewer.js`) loads it, pins *its own centre* to the camera every frame and draws it depth-test-off behind the scene — exactly the DS's camera-relative backdrop that turns with you but never gets closer. (Pinning the model *origin* is wrong: these domes are modelled around their centre, e.g. `old_peach_agb_V` sits at y≈92.)
- **Drive the CPU line.** `exportglb` also writes `<course>.path.json` — the enemy drive line (`EPOI`/`EPAT`, §V.1) flattened into one lap (start at section 0, follow `next[0]`) and converted to the GLB's own frame (world ÷ 16). A "Drive the CPU line" toggle flies the camera along a Catmull-Rom through it, at kart eye-height, looking ahead — a first-person lap of the track inside its skybox. 47 of the courses ship an enemy line.

(GLB export is deterministic — materials, buffer views and primitives are emitted in sorted material order — so re-exporting a model reproduces it byte-for-byte and the committed assets don't churn.)

## 8. Frontier: sprite cells and sound

*(frontier)* Remaining in Part IV: **`NCER` sprite cells** (with `NANR` animations) — the layouts assembling the "scattered" sprite `NCGR`s; and the **`SDAT` sound bank** (`sound_data.sdat`: sequences, banks, wave archives).

# Part V — Game mechanics

## 1. The NKM course map: the full track layout

Everything about a track that is not geometry lives in its **NKM** file (`NKMD`, up to three per course — single-player, plus mode variants). The format, decoded in `mariokartds/extract/mkds` and verified by overlay: a header (u16 version `37`, u16 header size `$4C`) followed by **17 section offsets**; each section is a 4-char magic, a u16 entry count, and fixed-size entries (sizes pinned by the offset deltas). Mario Circuit's map:

| Section | Count | Entry | Contents |
|---|---:|---:|---|
| `OBJI` | 41 | `$3C` | placed objects: position/rotation/scale (fx32), object ID, route, time-trial flag |
| `PATH`/`POIT` | 20/102 | `$04`/`$14` | routes — point lists that moving objects and cameras follow |
| `STAG` | — | `$2C` | stage settings (course ID, laps, fog) |
| `KTPS` | 1 | `$1C` | the grid start position + facing |
| `KTPJ` | 7 | `$20` | respawn points — where Lakitu drops you, one per track sector |
| `KTP2/KTPC/KTPM` | 1/0/0 | `$1C` | cannon/mission variants of the same |
| `CPOI` | 52 | `$24` | **checkpoints**: a gate line (x1,z1)–(x2,z2) on the track plane, key ID, respawn link |
| `CPAT` | 1 | `$0C` | checkpoint *sections*: `{start, len, next[3], prev[3]}` — the lap graph |
| `IPOI`/`IPAT` | 64/4 | `$14`/`$0C` | the **item-probe line** (what red shells steer along), same section scheme |
| `EPOI`/`EPAT` | 60/1 | `$18`/`$0C` | the **CPU drive line**: points with a lateral radius and a drift hint |
| `AREA` | 14 | `$48` | trigger volumes (camera/effect zones) |
| `CAME` | 19 | `$48` | cameras (intro fly-by, race cams) |

How a lap actually works, read straight from the data: the 52 `CPOI` gates are chained by `CPAT` into sections; crossing gates advances your position, the gate with **key ID 0 is the lap line**, and higher key IDs are *key checkpoints* — you must cross them in order for a lap to count, which is exactly the game's shortcut protection. Each checkpoint names the `KTPJ` respawn to use if you fall out inside its stretch. The CPU racers follow `EPOI` — a polyline with a per-point **radius** (how far they may wander from it) and a **drift** hint; item routing uses the parallel `IPOI` line. All three share the section scheme with up to three `next[]` links, and that is how **alternate routes** are encoded: Mario Circuit is one section end-to-end, but Waluigi Pinball's drive line splits into **10 sections** branching around the bumper field, Delfino Square has 6, DK Pass 4.

The `rendered/tracks/` figures (59 courses, `extract/cmd/trackmap`) draw it all in one frame: the course geometry top-down, the CPU line (cyan), item line (yellow), every checkpoint gate (red; key checkpoints gold, the lap line white), respawns (magenta), objects (white) and the start (green). That the drive line sits pixel-on-the-asphalt for every course is the joint verification of the model decoder, the NKM decoder and the ×16 world scale.

## 2. How the engine spawns the course objects

§1 described the `OBJI` records as data; this section traces, in the ARM9 code, how they become live objects. The chain was found by the pointer-table method: start from the path string the loader uses, follow the literal pools.

**Loading the course map.** The engine addresses the NKM *by name inside the mounted course archive* — the format strings sit together in `.data` (`$021647D8`ff) and their single referrer is the loader at **`$02042064`** (Thumb). It picks the file by game mode: `/course_map.nkm` for a race, `/MissionRun/%s_tool.nkm` for missions, `/Net/net_course_map.nkm` (falling back to `/course_map.nkm`) for download play. (The course NARC's own file table names its members — `course_model.nsbmd`, `course_map.nkm`, **`course_collision.kcl`** — pinning the collision format's name.) The loader allocates a **0x110-byte runtime CourseMap struct**, pointer in the global **`$02175620`**, NKM version word at `+0x10C`.

**17 positional section parsers.** The loader walks the header's 17 offsets in order, calling a **function table at `$0215407C`** — one parser per section, in exactly the OBJI…CAME order §1 assumed, which *proves* the sections are identified by position (no parser checks its magic; each just skips 4 bytes, reads the u16 count, and stores `{ptr, count}` into the CourseMap struct at `8×index`: `OBJI` at `+0x00/+0x04`, `PATH` at `+0x08/+0x0C`, `POIT` at `+0x10/+0x14`, … `CAME` at `+0x80/+0x84`).

**The map-object descriptor table.** Placed objects dispatch through a static table in `.data` at **`$0216B288`**: zero-terminated records of `{u32 objectID, u32 descriptor, u32 auxFn}` — **124 entries**, IDs `0x001`–`0x200`. The map-object manager (`$020D3xxx`, whose literal pools are how the table was found) matches each placed object's ID by **linear scan** (`$020D3452`: compare, step `+0xC`). Every descriptor starts with the **instance size** at `+0x04` and callback slots filled at boot by tiny registration trampolines (e.g. the cow's at `$020918A8`) through registrar thunks (`$0209BEF0`ff) that allocate a 0x2C-byte runtime record: `{+0x08 loadResources, +0x0C unload, +0x10 createInstance, +0x18 category}`.

One chain, verified end to end — object ID **0x134 = the Moo Moo Farm cow**: table entry `{0x134, $021665C4}` → trampoline `$020918A8` → *load* `$0209196C` (literal pool names `cow.nsbmd` + `cow.nsbtp`, stores the model in a global) → *create* `$020918CC`, which reads the instance's back-pointer to its OBJI record (instance`+0x9C`) and consumes the OBJI *settings* field (`LDRH +0x28` — the u16s after position/rotation/scale/ID/route are per-object parameters). `extract/cmd/objtable` dumps the whole table with each descriptor's resources recovered from its callbacks' literal pools: trees per course theme (`BeachTree1`, `Snow_Tree1`…), `teresa` (Boo), `bakubaku` (Cheep Chomp), `flipper` (0x1A9, the Waluigi Pinball flippers, instance size 0xEF4 — the biggest), `PakkunBody/ZHead` (Piranha Plants), `crab`, `sun`, `NsKiller1/2` (Bullet Bills), `IronBall`, `sanbo` (Pokey)…

Three details round out the picture. **(a)** The itembox (`0x65`, 852 placements — the most-placed object in the game) keeps its assets not in course archives but in the shared **`data/Main/MapObj.carc`** (`itembox.nsbmd`, its `.nsbca` bob animation, the shattered `itembox_hahen`); the same archive holds **`grpconf.tbl`**, 91 × 16-byte records `{u16 id, u16 flag, u16 params[5]}` of per-object config (the values pattern as visibility/clip ranges), whose path string sits in `.data` immediately before the descriptor table. **(b)** Descriptors at `$021A7xxx` (IDs `0x159`, `0x1F5`–`0x200`, …) point into the **overlay bank** — battle/mission objects whose code pages in with their mode. **(c)** A handful of simple IDs (`0x1`, `0x3`, `0x6`, `0x9`, `0xC`) share one generic descriptor (`$021580A4`) and differ only in the per-ID `auxFn` — parameterised variants of one water-effect object.

## 3. How route objects move

Objects with a route ID (`OBJI+0x26`) follow the `PATH`/`POIT` sections through a shared route-follower engine — and its point accessors are the code that proves how the raw data binds together.

**The derived path table.** Raw `PATH` entries are 4 bytes (`{u8 index, u8 loop, u16 numPoints}`) with *no start offset* — so which `POIT` points belong to which path is implicit. Immediately after parsing, **`$02042FD8`** builds a derived table (CourseMap`+0x94`) of 8-byte records `{rawPathEntry*, firstPoint*}` by walking the paths **in order, advancing a cursor through `POIT` by `numPoints × 0x14`** — the game's own code executing exactly the accumulation rule `mkds.ParseNKM` implements, upgrading that from inference to fact. The point accessor everything uses is `$02042044`: `point(path, j) = derived[path].points + j*0x14`.

**The follower.** A route-follower state block keeps `{+0x08 path index, +0x0C point index, +0x14 progress t, +0x1C previous point, +0x20 current point}`, with `t` a **fx24 fraction** (1.0 = `0x1000000`) of the current segment. The advance routine (`$020D8D30`ff, the accessor's main caller cluster) steps `t` each frame; on `t ≥ 1.0` it subtracts 1.0, advances the point index, and fetches the next point — wrapping to point 0 **iff the raw `PATH` entry's loop byte is set** (read through the derived record), otherwise reversing/stopping per object. Each object's expanded per-point records (0x54 bytes, factors at `+0x30/+0x34` applied by `SMULL…LSR #12` fixed-point multiplies) rescale the per-frame step per segment — i.e. **constant world speed across segments of different lengths**, with the position interpolated between previous and current points by `t`.

So the answer to "how do objects move" is layered: the *placement* (`OBJI`) names a route; the *route* is a `PATH`-ordered slice of `POIT`; the *follower* advances a normalised fx24 parameter with per-segment speed compensation; and what the object *does* along the way (a Goomba walking, the Airship's Bullet Bills firing) is its descriptor's per-type update callback.

## 4. Frontier: collision, physics, ghosts

*(frontier)* The collision file (`course_collision.kcl` — vertices/normals/prisms/octree, the format the kart actually drives on), kart physics, item behaviour, the CPU racers' use of the drive line, and the staff-ghost replay format in `data/Ghost`. On the object side: the generic factory that allocates instances (descriptor size `+0x04`) and fills the common header before the per-type `create` runs, and the per-type update callbacks themselves.

---

# Appendix A — Toolchain and reproduction

Everything here is derived from the `.nds` image with this repository's own tools: static disassembly for Parts I–II, and — for Part III — the game's own code run on our `tools/arm` core as an oracle. No third-party emulator, debugger or disassembler was used, and nothing was read from released source. Verify the image, then reproduce the catalog, the boot trace and the oracle run:

```sh
# identity (size + MD5 pinned in ../README.md#image-files)
md5 "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"

# Part I — header, integrity checks, overlay/filesystem summary (and -files / -tree)
go run retroreverse.com/tools/nds/cmd/ndsinfo "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"

# Part II — extract + BLZ-decompress the ARM9/ARM7 binaries and overlays into extracted/
( cd extract && go run ./cmd/ndsextract "../Mario Kart DS (Europe) (En,Fr,De,Es,It).nds" )

# trace the ARM9 boot chain from its entry (ARM state) over the decompressed image
go run retroreverse.com/tools/cmd/codetracearm -base 0x02000000 -entry 0x02000800 extracted/arm9_dec.bin

# trace the ARM7 boot chain from its entry
go run retroreverse.com/tools/cmd/codetracearm -base 0x02380000 -entry 0x02380000 extracted/arm7.bin

# Part III — run the ARM9 boot on the tools/arm core as an oracle: BLZ cross-check,
# the I/O registers it programs, and the ARM9↔ARM7 IPCSYNC rendezvous it stops at
( cd extract && go run ./cmd/bootoracle -io "../Mario Kart DS (Europe) (En,Fr,De,Es,It).nds" )

# Part IV — extract the filesystem, then render EVERY texture and 2D screen in the
# game (rendered/course, rendered/tex, rendered/ui) with one sweep
( cd extract && go run ./cmd/ndsextract -fs "../Mario Kart DS (Europe) (En,Fr,De,Es,It).nds" )
( cd extract && go run ./cmd/renderall )

# or individually: one texture set / one UI archive
( cd extract && go run ./cmd/rendertex -o ../rendered/course ../extracted/files/data/Course/beach_courseTex.carc )
( cd extract && go run ./cmd/render2d  -o ../rendered/ui     ../extracted/files/data/CupPicture )

# 3D models: inspect structure, software-render to PNG, export all 377 to GLB (+manifest + drive-line JSON)
( cd extract && go run ./cmd/modeldump   ../extracted/files/data/KartModelMenu/kart/mario/kart_MR_a.nsbmd )
( cd extract && go run ./cmd/rendermodel ../extracted/files/data/Course/mario_course.carc )
( cd extract && go run ./cmd/exportglb -all )   # → extracted/glb/, copied to site/public/mariokart/

# Part V — the full track layout: course geometry + NKM overlay (checkpoints, CPU
# line, item line, spawns, objects) for one course or all of rendered/tracks/
( cd extract && go run ./cmd/trackmap ../extracted/files/data/Course/mario_course.carc )

# Part V §2 — the map-object descriptor table ($0216B288): 124 object IDs with their
# instance sizes and the NSBMD/NSBCA resources recovered from the callbacks' literal pools
( cd extract && go run ./cmd/objtable )
```

Toolchain (all under the `retroreverse.com/tools` module unless noted, this repository):

- **`tools/nds`** — the Nintendo DS cartridge container reader: header parse + CRC-16 verification, the FAT, the FNT directory-tree walk, the ARM9 overlay table, and the **BLZ** (backward-LZSS) decompressor used by the ARM9 and its overlays. New for this project.
- **`tools/nds/cmd/ndsinfo`** — the container inspector for Part I (`-files`, `-tree`, `-grep`).
- **`tools/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware). New for this project.
- **`tools/cmd/disarm`** — linear ARM/Thumb disassembler.
- **`tools/cmd/codetracearm`** — recursive-descent code-tracer that follows ARM↔Thumb interworking; used to trace both boot chains.
- **`mariokartds/extract/cmd/ndsextract`** — the game's extractor: writes `arm9.bin`/`arm7.bin` and the overlays, and their BLZ-decompressed forms (`arm9_dec.bin`, `ovl9_00N_dec.bin`), into `extracted/` (regenerable, git-ignored).
- **`mariokartds/extract/cmd/bootoracle`** — runs the ARM9 boot on the `tools/arm` core over a flat DS memory (with the BIOS `SWI`s the startup needs): cross-checks BLZ against the game's own decompressor, logs the I/O registers programmed, and stops at the ARM9↔ARM7 IPCSYNC rendezvous. The DS analogue of the Amiga per-game oracles.
- **`tools/nds` LZ77 + NARC** — `DecompressLZ77`/`Decompress` (the forward LZ10/LZ11 the filesystem uses) and `ParseNARC` (splits a NARC, transparently decompressing a `.carc` first). Both unit-tested.
- **`tools/nds/nitro`** — the NITRO resource decoders. `DecodeNSBTX` turns a `BTX0`/`TEX0` texture set into Go images (resource-dictionary parse, `texImageParam`, all seven texture formats including 4x4-compressed, BGR555, name-matched palettes); `ParseNCLR`/`ParseNCGR`/`ParseNSCR` + `ComposeScreen`/`TileSheet` decode and compose the 2D tile pipeline.
- **`mariokartds/extract/cmd/rendertex`** — renders every texture in an NSBTX, or in a `.carc`/NARC's `BTX0` blocks, to PNG.
- **`mariokartds/extract/cmd/render2d`** — composes the NSCR screens of a directory or `.carc` through their NCGR/NCLR (name-paired, `_b` background banks preferred).
- **`mariokartds/extract/cmd/renderall`** — the whole-filesystem sweep: every texture, every screen (merging language-variant archives with their base), every leftover tile sheet, and the raw `.nbfc`/`.nbfp` banner — regenerates all 2,300+ figures in `rendered/`.
- **`tools/nds/nitro` models** — `ParseNSBMD` (nodes/TRS + pivot rotations, SBC, materials with the authoritative tex↔pal binding, shapes), `RunSBC` (joint matrices → matrix stack → draw list), `DecodeDL` (the GX display-list interpreter: all vertex forms, strips), `ExportGLB` (standard binary glTF 2.0 with embedded PNG textures), `DecodeContainerTextures` (the TEX0 of any NITRO container, incl. BMD0-embedded), and `DecodeNSBTA` (BTA0 texture-SRT animation: per-material scale/rotation/translation tracks, constant or sampled every 4th frame).
- **`mariokartds/extract/cmd/modeldump` / `rendermodel` / `exportglb`** — model structure dump; z-buffered software render to PNG (`rendered/models/`); GLB export of all 377 models (characters, karts, every course scene + skybox + map objects, including the retro tracks) plus each course's `<name>.path.json` drive line and the `models.json` manifest that `site/public/mariokart/` serves. The Studio site (`site/`) carries them under the "Nintendo DS" system (`site/src/mariokart/viewer.js`), one section per track (the track plus its map objects), with camera-locked skyboxes and a drive-the-CPU-line fly-through.
- **`mariokartds/extract/mkds`** — the game-specific plumbing: `LoadModels`/`LoadTextures` (loose files, NARC-embedded models, the cross-archive `<name>Tex.carc` convention) and `ParseNKM`, the course-map decoder.
- **`mariokartds/extract/cmd/trackmap`** — the track-layout figure generator (`rendered/tracks/`, 59 courses): top-down course geometry at the ×16 world scale, overlaid with every NKM element.
- **`mariokartds/extract/cmd/objtable`** — dumps the Part V §2 map-object descriptor table from the ARM9 image: 124 `{id, descriptor, auxFn}` records at `$0216B288`, each with its instance size and the resource names recovered from the callbacks' literal pools.

Rendered figures go in `Mario Kart DS (DS)/rendered/`; annotated disassembly in `disasm/`.
