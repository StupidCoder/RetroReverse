# Super Mario 64 DS (Nintendo DS) — cartridge format and game analysis

Super Mario 64 DS is a Nintendo DS launch title (2004) — a remake of the Nintendo 64 original that rebuilds the castle and its worlds on the DS's twin 2D/3D engines, adds four playable characters, a touch-screen map, and a suite of minigames. This document reverse-engineers the *shipped cartridge* from the `.nds` image alone, the same way the other titles in this repository are taken apart: no third-party emulator, debugger or disassembler, and nothing read from released source.

It is the second Nintendo DS title analysed here (after [[Mario Kart DS]]), so it stands on the same toolchain — the shared `tools/nds` cartridge reader (header, FNT/FAT, overlays, BLZ decompression) and the `tools/arm` ARM/Thumb disassembler and code-tracer — and the two make an instructive pair: same platform, same NitroSDK runtime, but different choices at almost every turn (where the ARM9 loads, how many overlays it carries, which asset formats it stores).

Image: `Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds`, 16 MiB, game code **ASMP**, MD5 `867b3d17ad268e10357c9754a77147e5` (pinned in `../README.md`).

## Contents

* **Part I** — the cartridge image: the ROM header and its integrity checks, the two CPU binaries and the unusually large **overlay table**, the FNT/FAT filesystem, and the asset catalog;
* **Part II** — the boot chain: both NitroSDK `crt0` startup stubs, the ARM9's in-place BLZ self-decompression, and the handoff to each processor's `main`.

Parts III onward (program architecture, the NITRO asset formats, the game's mechanics and its SDAT music) are future work; the toolchain and container decoding carry straight over from the Mario Kart DS analysis.

Methods: purely static analysis of the `.nds` image. All addresses are 32-bit ARM addresses (`$02000000`-style main-RAM addresses, or the BIOS and I/O regions) unless a *file offset* into the ROM image is explicitly called out; bytes are little-endian.

# Part I — The cartridge image

## 1. The ROM dump

The image is a 16 MiB Nintendo DS Game Card dump. A DS card is not a memory-mapped cartridge like the Game Boy or Game Gear: it is a serial device behind a controller, and its contents are an on-card **filesystem** the console reads on demand. The DS is also a **two-processor machine** — an ARM9 and an ARM7 sharing one 4 MiB main memory — so the card carries two independent code images that boot side by side. The 512-byte header at the front of the image describes all of this to the BIOS.

## 2. The DS dual-CPU architecture and the address map

The header hands the two CPUs *different* load and entry addresses:

- the **ARM9** binary loads to `$02004000` in **main RAM** and starts at `$02004800`;
- the **ARM7** binary loads to `$02380000` — also inside main RAM, near its top — and starts there.

Both CPUs share the single 4 MiB main memory `$02000000`–`$023FFFFF`. Two things stand out immediately against Mario Kart DS, which loads its ARM9 at the very base `$02000000`: Super Mario 64 DS bases its ARM9 image `$4000` bytes higher, **leaving the low 16 KiB of main RAM (`$02000000`–`$02003FFF`) free** for the system to use, and (as with any DS title) the entry sits `$800` past the load base, over the secure-area landing.

The ARM9 is an ARMv5TE core and the ARM7 an ARMv4T core; both run the 32-bit **ARM** and 16-bit **Thumb** instruction sets, switched via `BX`/`BLX`, so disassembling either binary means tracking ARM-vs-Thumb state per address — exactly what `tools/arm` and `codetracearm` do. The ARM9 owns the card and drives the game and both screens; the ARM7 sits near the top of main RAM and services sound, the touchscreen, buttons and wireless, the two coordinating through the DS's **IPC** hardware in the `$04000000` I/O block.

## 3. The cartridge header (`$000`–`$15F`)

Decoding the header against the real image, raw bytes shown little-endian with the assembled value on the right:

```
$000  53 2E 4D 41 52 49 4F 36 34 44 53 00   title            "S.MARIO64DS"
$00C  41 53 4D 50                            game code        "ASMP"  (A=DS card, SM=Super Mario, P=Europe/PAL)
$010  30 31                                  maker code       "01"    — Nintendo
$012  00                                     unit code        — NDS (0)
$014  07                                     device capacity  — 128 KiB << 7 = 16 MiB  ✓ matches file size
$01F  00                                     autostart flags  — 0
$020  00 40 00 00                            ARM9 ROM offset  = $00004000
$024  00 48 00 02                            ARM9 entry addr  = $02004800
$028  00 40 00 02                            ARM9 RAM addr    = $02004000
$02C  04 D5 05 00                            ARM9 size        = $0005D504   (≈ 382 KiB, compressed)
$030  00 48 1B 00                            ARM7 ROM offset  = $001B4800
$034  00 00 38 02                            ARM7 entry addr  = $02380000
$038  00 00 38 02                            ARM7 RAM addr    = $02380000
$03C  24 4B 02 00                            ARM7 size        = $00024B24   (≈ 147 KiB)
$040  24 93 1D 00                            FNT offset       = $001D9324
$044  FD BE 00 00                            FNT size         = $0000BEFD
$048  24 52 1E 00                            FAT offset       = $001E5224
$04C  F8 43 00 00                            FAT size         = $000043F8   (÷8 = 2175 entries)
$050  10 15 06 00                            ARM9 overlay tbl = $00061510
$054  E0 0C 00 00                            ARM9 overlay sz  = $00000CE0   (÷32 = 103 overlays)
$058  00 00 00 00                            ARM7 overlay tbl = none
$068  00 98 1E 00                            icon/banner off  = $001E9800
$06C  70 A8                                  secure-area CRC  = $A870
$070  EC 49 00 02                            ARM9 autoload hook = $020049EC
$074  10 01 38 02                            ARM7 autoload hook = $02380110
$080  80 67 F7 00                            total used size  = $00F76780   (≈ 15.5 MiB of 16)
$084  00 40 00 00                            header size      = $00004000
$15C  56 CF                                  logo checksum    = $CF56  ✓
$15E  7C 47                                  header checksum  = $477C  ✓ (CRC-16 over $000–$15D)
```

Observations:

- **A 16 MiB card, over 96 % full.** The device-capacity code `$07` gives a 16 MiB chip, and the total-used size (`$F76780`, ≈ 15.5 MiB) leaves barely half a megabyte spare — a dense little image.
- **Two independent code images.** The ARM9 (`$4000`, 382 KiB stored) and ARM7 (`$1B4800`, 147 KiB) binaries each carry their own RAM-load address and entry point, so the header alone fully describes where each CPU starts (Part II). The ARM9 stored size is *compressed*; it expands at boot.
- **The autoload hooks** (`$070`/`$074`) point at each binary's auto-load list — the table the startup stub walks to place its `.data` and clear its `.bss` before jumping to `main` (Part II §4).
- **Three independent checksums.** The header CRC-16 covers `$000`–`$15D` and recomputes to `$477C`. The Nintendo-logo CRC (`$15C` = `$CF56`) is the standard value the BIOS validates before it runs any code — identical across every retail DS card, because the logo bitmap is the same. The secure-area CRC (`$06C` = `$A870`) covers the encrypted boot region.

## 4. The ARM9 and ARM7 binaries and overlays

Beyond the two main binaries, the ARM9 uses **code overlays** — separately loadable code+data blobs swapped into fixed RAM regions on demand, the DS's answer to fitting far more code than RAM holds. And here Super Mario 64 DS diverges sharply from Mario Kart DS's four: its overlay table at `$061510` (`$054` = `$CE0` bytes ⇒ **103 overlays**) is an array of 32-byte records. Decoding a sample:

```
ovl  RAM addr    ramsize    bss      fileID
  0  $020AA420   $0150C0    $0000       0
  1  $020AA420   $003200    $0040       6
  2  $020AD660   $060340    $3800       8
  3  $020AD660   $004240    $0000      10
  4  $020AD660   $011500    $1360      12
  5  $020BFEC0   $003160    $0000      14
 …
102  $02148A80   $005C60    $03C0      98
```

The 103 overlays occupy **just 22 distinct RAM addresses**, spread across `$020AA420`–`$02148A80` (immediately above the ARM9's static `.bss`, which ends at `$020AA420`). Overlays that share an address are mutually-exclusive banks — only one resident at a time — so the game is organised as a large tree of on-demand code modules keyed to state: the castle, each painting-world, the minigames, the menus. Each record also names the **FAT file ID** holding its bytes (IDs `0`–`102`, so the filesystem's *named* files begin at ID 103), plus a `.bss` size to zero and a static-initialiser range to run after loading — the same auto-load convention as the main binaries. Every overlay is individually **BLZ-compressed** (Part II §3 covers BLZ). The ARM7 has no overlays.

## 5. The filesystem: FNT and FAT

Everything that is not boot code — models, animations, collision, textures, the sound bank, the 2D UI — lives in an on-card filesystem addressed by two tables the header locates:

- the **FAT** (file allocation table) at `$1E5224`, `$43F8` bytes, is a flat array of 8-byte records, one **start** and **end** file offset per file. `$43F8 / 8 = 2175` files. File *ID* is the index; the bytes are `image[start:end]` — storage framing only.
- the **FNT** (file name table) at `$1D9324`, `$BEFD` bytes, is the **directory tree** that gives those numbered files names and a hierarchy. Its main table is an array of 8-byte directory records (record 0 = the root, whose "parent" field instead holds the total directory count); each points at a sub-table of that directory's entries. A sub-table entry is a control byte — `$00` ends the directory, `$01`–`$7F` a file whose name length is the low 7 bits, `$81`–`$FF` a subdirectory — followed by the name, and for a subdirectory a child directory ID. Files take **sequential IDs** from the directory's "first file ID", binding names to FAT indices.

`tools/nds` joins the two tables: of the 2175 FAT entries, IDs **0–102 are the ARM9 overlays** (referenced directly by the overlay table, with no filesystem name) and IDs **103–2174 are the 2072 named files**. Walking from the root yields paths like `data/enemy/wanwan/wanwan.bmd` and `ARCHIVE/arc0.narc`.

As always, a filesystem file is *not* the asset: the top-level `ARCHIVE/` directory holds a handful of **`NARC`** archives (`arc0.narc`, `ar1.narc`, the per-language `cee`/`cef`/`ceg`/`cei`/`ces`, the versus-mode `vs1`–`vs4`), each a mini-container of further sub-files, and every archive is itself LZ77-compressed inside its `.carc`-style wrapper. The nested layering (FAT slice → compression → NARC → NITRO resource) is decoded in the later graphics Part.

## 6. The file catalog

The 2072 named files group cleanly by extension, and the histogram sketches the engine's asset pipeline before any of it is decoded:

| Ext | Count | What it is |
|---|---:|---|
| `.bin` | 763 | raw/engine-specific data (2D UI tile maps, tables) |
| `.bca` | 493 | NITRO character (joint) animation — the shorter-named sibling of Mario Kart DS's `.nsbca` |
| `.bmd` | 455 | NITRO 3D model (`BMD0`: geometry + material, textures embedded) |
| `.kcl` | 241 | **collision mesh** — the `KCL` triangle-soup/octree format the physics runs on |
| `.btp` | 105 | NITRO texture-pattern animation (frame-swapped textures) |
| `.narc` | 13 | `NARC` archive (bundles of the above) |
| `.sdat` | 1 | the entire SDAT sound bank (SSEQ/SBNK/SWAR), 4.4 MiB |

Two things distinguish this catalog from Mario Kart DS's. First, the **naming**: Super Mario 64 DS uses the terse early-NitroSDK extensions `.bmd`/`.bca` where the later Mario Kart DS writes `.nsbmd`/`.nsbca` — the same underlying `BMD0`/`BCA0` NITRO formats under an older suffix. Second, there is **no separate texture-set extension** (`.nsbtx`): with 455 models and no loose texture files, the models carry their textures **embedded** in their own `BMD0` (a `TEX0` block after the geometry), which the later graphics Part decodes. The abundance of `.kcl` (241 collision meshes — one per world, object and platform) foreshadows a game that is, at heart, a physics playground.

The top-level directories tell the same story: `data/` (1509 files — the bulk, with a large `data/enemy/<name>/` subtree per creature: `wanwan`, `basabasa`, `bakubaku`, `battan_king`, `big_snowman`…, plus `data/2D_cad` UI graphics and `data/DSMT` map data), `MG/` (549 files — the minigame assets) and `ARCHIVE/` (13 `NARC` bundles). The full catalog is reproducible with `ndsinfo -files` (Appendix).

# Part II — The boot chain

Part I read the header the DS BIOS reads; Part II follows what runs after it. Both CPU images are extracted (and, for the ARM9, decompressed) with the game's `ndsextract`, and traced from the header's entry points with `codetracearm`. What comes out is a pair of near-textbook **NitroSDK** startup stubs — the ARM9's decompresses the whole game in place before jumping to `main`, the ARM7's is a smaller uncompressed sibling.

## 1. Two CPUs, two entry points

On power-up each CPU runs its on-chip **BIOS**, which validates the header, copies each CPU's binary to the RAM address the header names, and jumps to its entry point:

| CPU | ROM offset | → RAM load | Entry | Size (stored) |
|---|---|---|---|---:|
| ARM9 | `$004000` | `$02004000` | `$02004800` | `$05D504` (compressed) |
| ARM7 | `$1B4800` | `$02380000` | `$02380000` | `$024B24` |

The ARM9 binary's first 16 KiB (`$02004000`–`$02007FFF`) is the **secure area** — on a real cartridge it is KEY1/Blowfish-encrypted and the BIOS decrypts it during a challenge/response handshake with the card. In this dump the secure area is already decrypted: its first eight bytes read `FF DE FF E7  FF DE FF E7` — the two marker words the SDK writes over the encrypted "secure-area ID" once decryption succeeds — and the code at the entry disassembles as clean ARM. The ARM9 entry `$02004800` sits `$800` past the load base, stepping over that ID region; the ARM9 comes up first, as the CPU that owns the card and main RAM.

## 2. The ARM9 startup stub: CP15, the TCMs and the stacks

The entry kills interrupts, then configures the CP15 system-control coprocessor before it can even use a stack:

```
02004800  MOV  r12, #0x04000000        ; I/O base
02004804  STR  r12, [r12, #0x208]      ; IME = 0x04000000 (bit0 clear) → interrupts off
02004808  BL   0x020049F0              ; CP15 / MPU / TCM setup
```

`sub_020049F0` is the CP15 sequence (`MRC`/`MCR p15`): disable caches, invalidate the instruction and data caches, drain the write buffer, program eight **MPU protection regions** (`c6, c0`…`c6, c7`), and enable the two **tightly-coupled memories** — **ITCM** (mapped so the low addresses just below main RAM, around `$01FF8000`, hit its fast 32 KiB) and **DTCM**, which Super Mario 64 DS bases at **`$023C0000`**, high in main RAM (Mario Kart DS chose `$027E0000`; the base is a per-title decision). It then re-enables caches/MPU/TCM by writing the control register.

With the TCMs live, the stub sets up a stack per CPU mode by switching mode with `MSR CPSR_c` and pointing `sp` into the top of DTCM, packing the three exception stacks down from `DTCM+$3FC0`:

```
0200480C  MOV r0,#0x13 ; MSR CPSR_c,r0            ; Supervisor mode
02004814  LDR r0,=0x023C0000 ; ADD r0,#0x3FC0 ; MOV sp,r0   ; SVC stack at DTCM+0x3FC0
02004820  MOV r0,#0x12 ; MSR CPSR_c,r0            ; IRQ mode  → stack at DTCM+0x3FC0-0x40
02004840  MOV r0,#0x1F ; MSR CPSR_cxsf,r0         ; System mode (the mode main runs in)
```

## 3. The ARM9 self-decompression (BLZ)

The header's "ARM9 size" (`$05D504`) is a *compressed* size: the bulk of the ARM9 image is packed with Nintendo's **BLZ** — an LZSS variant that decodes **backward**, expanding the image in place. The stub decompresses itself, driven by the NitroSDK **module-params** struct (found via the literal pool, at `$02004AD8`):

```
02004848  SUB sp, r1, #4               ; (finish the System-mode stack)
0200484C  LDR r1, =0x02004AD8          ; module-params struct
02004850  LDR r0, [r1, #0x14]          ; r0 = compressedStaticEnd = 0x02061504
02004854  BL  0x020048D8               ; MIi_UncompressBackward (BLZ)
```

`sub_020048D8` is the BLZ decompressor itself, inline in the stub — it reads a footer at the end of the compressed image (`LDMDB r0,{r1,r2}` fetches the encoded/increment lengths) and decodes downward, copying literals and back-references from high addresses to low. Decoding the module-params struct:

```
$02004AD8 +0x00  0x020A0F60   autoload list start
          +0x04  0x020A0F78   autoload list end        (= end of the decompressed ARM9)
          +0x08  0x0209B000   autoload data start
          +0x0C  0x0209B000   static .bss start
          +0x10  0x020AA420   static .bss end          (= the overlay load base, Part I §4)
          +0x14  0x02061504   compressed-static end     → passed to the BLZ decompressor
          +0x1C  0xDEC00621   NitroSDK signature
```

The compressed-static end `$02061504` is exactly the load base `$02004000` plus the header's compressed ARM9 size `$5D504`. The image expands from `$02004000`–`$02061504` to a static code+data region ending at `.bss` start `$0209B000`, with `.bss` running on to `$020AA420` — precisely the base at which the first overlay loads. The far calls that follow the decompression (e.g. `BL 0x02019780`) land in valid code only *after* this step. BLZ is reimplemented from scratch in Go (`tools/nds`, `DecompressBLZ`) — reversing the stream, running a forward LZSS, reversing the result — and the same routine decompresses all 103 overlays.

## 4. Autoload, `.bss`, and the handoff to `main`

With the image expanded, the stub finishes the C runtime:

- **Autoload** (`sub_0200497C`) walks the `$020A0F60`–`$020A0F78` list, copying each block from its packed location (`$0209B000` onward, the end of the decompressed image) to its run address — this is how the ITCM/DTCM-resident code and initialised data reach their fast memories.
- three **memory-clear** calls (`BL 0x0205A47C` with lengths `$4000`, `$400`, `$400`) zero the static and I/O-shadow regions, and a fourth write installs the IRQ-handler pointer into the DTCM slot at `DTCM+$3FC0+$3C`.
- a **cache flush** (`BL 0x01FFAFD4`, itself ITCM-resident) makes the freshly-written code visible to the instruction side.
- two far initialisers run (`BL 0x02019780`, `BL 0x02072F94` — SDK/OS init in the now-decompressed image).

Then the handoff, `main` and the exit vector loaded from the literal pool:

```
020048AC  LDR r1, =0x02007000          ; main
020048B0  LDR lr, =0xFFFF0000          ; return address = ARM9 BIOS (halt on return)
020048B4  BX  r1
```

`main` at **`$02007000`** is the real game entry (unlike Mario Kart DS, whose `main` is a thin two-call wrapper). It opens by calling a run of SDK/OS initialisers (`BL 0x02059788`, `BL 0x02059BC0`, …) — the subject of a future Part III. If it ever returns, `lr = $FFFF0000` drops the ARM9 into its BIOS.

## 5. The ARM7 startup

The ARM7 binary is **not** compressed — its body disassembles coherently end to end, with no module-params decompression call — so its `crt0` is shorter. From entry `$02380000`:

```
02380000  MOV r12,#0x04000000 ; STR r12,[r12,#0x208]   ; IME = 0 (interrupts off)
02380008  LDR r1,=0x02380168 ; MOV r0,#0x03800000 ; CMP;MOVPL   ; range for the WRAM clear
02380020  …STMLTIA r1!,{r0}…                            ; zero ARM7-private WRAM up to 0x0380FF00
0238002C  MOV r0,#0x13 ; MSR CPSR_c,r0 ; LDR sp,=0x0380FF00   ; SVC stack (ARM7 WRAM top)
02380038  MOV r0,#0x12 ; MSR CPSR_c,r0 ; LDR sp,=0x0380FFC0   ; IRQ stack
02380050  MOV r0,#0x1F ; MSR CPSR_cxsf,r0                     ; System mode
0238005C  BL  0x023800A0                                      ; autoload
02380060  BL  0x02380114                                      ; .bss clear
02380070  LDR r1,=0x037F8300 ; LDR lr,=0xFFFF0000 ; BX r1     ; jump to ARM7 main
```

The ARM7's stacks live at the top of its private 64 KiB WRAM (`$03800000`–`$0380FFFF`): SVC at `$0380FF00`, IRQ at `$0380FFC0`, System below. Its autoload (`sub_023800A0`, the copy loop visible inline) relocates its resident code, and — notably — the entry point it finally branches to, **`$037F8300`**, is *in WRAM*, not in the `$02380000` load region: the ARM7 **relocates its hot code into WRAM** and runs it from there, close to the sound and input hardware it drives. As with the ARM9, `lr = $FFFF0000` returns to the ARM7 BIOS.

## 6. The two processors meet

Both binaries load into the same 4 MiB main RAM (the ARM9 at `$02004000`, the ARM7 near the top at `$02380000`) and both reach `main` independently. From there they coordinate through the DS's **inter-processor communication** hardware in the `$04000000` I/O block — the `IPCSYNC` mailbox and the `IPCFIFO` queue — with main RAM carrying the bulk payloads: input and touch state up from the ARM7, sound commands and DMA lists down from the ARM9. Pinning the exact IPC handshake, the runtime memory map the game builds, and the overlay-to-state mapping (which of the 103 overlays backs which world or minigame) is the work of a future Part III, following the oracle-driven approach used for Mario Kart DS.

# Appendix A — Toolchain and reproduction

Everything here is derived from the `.nds` image with this repository's own tools: the shared `tools/nds` container reader and the `tools/arm` disassembler/tracer, plus this game's `extract` module. No third-party emulator, debugger or disassembler was used, and nothing was read from released source.

```sh
# identity (size + MD5 pinned in ../README.md#image-files)
md5 "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"

# Part I — header, integrity checks, overlay/filesystem summary (and -files / -tree)
go run retroreverse.com/tools/nds/cmd/ndsinfo "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
go run retroreverse.com/tools/nds/cmd/ndsinfo -tree  "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
go run retroreverse.com/tools/nds/cmd/ndsinfo -files "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"

# Part II — extract + BLZ-decompress the ARM9/ARM7 binaries and all 103 overlays into extracted/
( cd extract && go run ./cmd/ndsextract "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )
# add -fs to also dump every filesystem file under extracted/files/

# trace the ARM9 boot chain from its entry (ARM state) over the decompressed image
go run retroreverse.com/tools/cmd/codetracearm -base 0x02004000 -entry 0x02004800 extracted/arm9_dec.bin

# trace the ARM7 boot chain from its entry
go run retroreverse.com/tools/cmd/codetracearm -base 0x02380000 -entry 0x02380000 extracted/arm7.bin
```

Toolchain (shared `retroreverse.com/tools`, this repository):

- **`tools/nds`** — the Nintendo DS Game Card container reader: header parse + CRC-16 verification, the FAT, the FNT directory-tree walk, the ARM9 overlay table, and the **BLZ** (backward-LZSS) decompressor the SDK applies to the ARM9 static module and every overlay. Shared with the [[Mario Kart DS]] analysis; makes no assumptions about the game inside.
- **`tools/nds/cmd/ndsinfo`** — the container inspector for Part I (`-files`, `-tree`, `-grep`).
- **`tools/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware), with `tools/cmd/disarm` (linear) and `tools/cmd/codetracearm` (recursive-descent, follows ARM↔Thumb interworking) as CLIs.
- **`supermario64ds/extract/cmd/ndsextract`** — this game's extractor: writes `arm9.bin`/`arm7.bin` and the 103 overlays, and their BLZ-decompressed forms (`arm9_dec.bin`, `ovl9_NNN_dec.bin`), into `extracted/` (regenerable, git-ignored).

Rendered figures will go in `Super Mario 64 DS (DS)/rendered/`; annotated disassembly in `disasm/`.
