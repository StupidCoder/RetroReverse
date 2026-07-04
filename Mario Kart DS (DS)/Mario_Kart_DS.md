# Mario Kart DS (Nintendo DS) — cartridge format and game analysis

A reverse-engineering reference for `Mario Kart DS (Europe) (En,Fr,De,Es,It).nds`, the 2005 Nintendo kart racer. This is the repository's first **Nintendo DS** title and its first **ARM** project: where every earlier game runs on a single 8- or 16-bit CPU, the DS is a dual-processor machine — an **ARM9** (ARM946E-S, ARMv5TE, the 67 MHz main CPU) beside an **ARM7** (ARM7TDMI, ARMv4T, the 33 MHz sound/IO CPU) — and its `.nds` cartridge is a real container with a header, two CPU binaries, code overlays and an on-cartridge filesystem, unlike the flat mask-ROM of a Game Boy cartridge or the raw disk of an Amiga floppy. The writeup follows the same shape as the others, in reading order:

* **Part I** — the cartridge image: the `.nds` container, its header, the ARM9/ARM7 binaries and overlays, and the FNT/FAT filesystem that names ~600 asset files;
* **Part II** — the boot chain: the cartridge header's entry points, the secure area, and how the ARM9 and ARM7 come up and hand off;
* **Part III** — program architecture: the runtime memory map, how the game initialises through the OS layer, its interrupt and IPC-FIFO setup, and the ARM9↔ARM7 rendezvous — pinned by running the boot on the `tools/arm` core as an oracle;
* **Part IV** — graphics and data formats: the NITRO asset formats (`NSBMD` models, `NSBTX` textures, `NSBCA` animation), the `NARC` archive and its LZ compression, and the 2D banner/UI graphics;
* **Part V** — game mechanics: track data, kart physics and item behaviour;
* **Appendix A** — toolchain and reproduction.

Methods: purely static analysis of the `.nds` image. The DS needed a new toolchain — the shared 6502/68000/Z80/LR35902 decoders do not apply — so this project is built on the new `tools/nds` cartridge reader (header, FNT/FAT, overlays, BLZ decompression) and the new `tools/arm` ARM/Thumb disassembler and CPU core, with `tools/nds/cmd/ndsinfo` for the container catalog, the game's `mariokartds/extract` `ndsextract` to pull the CPU binaries, and `retroreverse.com/tools/cmd/disarm` / `codetracearm` for the code. All addresses are 32-bit ARM addresses (`$02000000`-style main-RAM addresses, or the ARM9/ARM7 BIOS and I/O regions) unless a *file offset* into the ROM image is explicitly called out; bytes are little-endian. **Parts I–III are complete (the post-rendezvous main loop and overlay streaming are Part III's frontier); Parts IV–V are stubs.**

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

*(frontier)* The compression and container layers (LZ77 `0x10`/`0x11`, Huffman, RLE; the `NARC` archive), then the NITRO asset formats — `NSBMD` models, `NSBTX` textures, `NSBCA`/`NSBTP`/`NSBTA`/`NSBMA` animation, the `NCLR`/`NCGR`/`NSCR` 2D graphics, and the `SDAT` sound bank.

# Part V — Game mechanics

*(frontier)* Track collision and route data, kart physics, item behaviour, the CPU racers, and the staff-ghost replay format in `data/Ghost`.

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
```

Toolchain (all under the `retroreverse.com/tools` module unless noted, this repository):

- **`tools/nds`** — the Nintendo DS cartridge container reader: header parse + CRC-16 verification, the FAT, the FNT directory-tree walk, the ARM9 overlay table, and the **BLZ** (backward-LZSS) decompressor used by the ARM9 and its overlays. New for this project.
- **`tools/nds/cmd/ndsinfo`** — the container inspector for Part I (`-files`, `-tree`, `-grep`).
- **`tools/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware). New for this project.
- **`tools/cmd/disarm`** — linear ARM/Thumb disassembler.
- **`tools/cmd/codetracearm`** — recursive-descent code-tracer that follows ARM↔Thumb interworking; used to trace both boot chains.
- **`mariokartds/extract/cmd/ndsextract`** — the game's extractor: writes `arm9.bin`/`arm7.bin` and the overlays, and their BLZ-decompressed forms (`arm9_dec.bin`, `ovl9_00N_dec.bin`), into `extracted/` (regenerable, git-ignored).
- **`mariokartds/extract/cmd/bootoracle`** — runs the ARM9 boot on the `tools/arm` core over a flat DS memory (with the BIOS `SWI`s the startup needs): cross-checks BLZ against the game's own decompressor, logs the I/O registers programmed, and stops at the ARM9↔ARM7 IPCSYNC rendezvous. The DS analogue of the Amiga per-game oracles.

Rendered figures go in `Mario Kart DS (DS)/rendered/`; annotated disassembly in `disasm/`.
