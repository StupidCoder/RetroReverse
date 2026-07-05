# Super Mario 64 DS (Nintendo DS) — cartridge format and game analysis

Super Mario 64 DS is a Nintendo DS launch title (2004) — a remake of the Nintendo 64 original that rebuilds the castle and its worlds on the DS's twin 2D/3D engines, adds four playable characters, a touch-screen map, and a suite of minigames. This document reverse-engineers the *shipped cartridge* from the `.nds` image alone, the same way the other titles in this repository are taken apart: no third-party emulator, debugger or disassembler, and nothing read from released source.

It is the second Nintendo DS title analysed here (after [[Mario Kart DS]]), so it stands on the same toolchain — the shared `tools/nds` cartridge reader (header, FNT/FAT, overlays, BLZ decompression) and the `tools/arm` ARM/Thumb disassembler and code-tracer — and the two make an instructive pair: same platform, same NitroSDK runtime, but different choices at almost every turn (where the ARM9 loads, how many overlays it carries, which asset formats it stores).

Image: `Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds`, 16 MiB, game code **ASMP**, MD5 `867b3d17ad268e10357c9754a77147e5` (pinned in `../README.md`).

## Contents

* **Part I** — the cartridge image: the ROM header and its integrity checks, the two CPU binaries and the unusually large **overlay table**, the FNT/FAT filesystem, and the asset catalog;
* **Part II** — the boot chain: both NitroSDK `crt0` startup stubs, the ARM9's in-place BLZ self-decompression, and the handoff to each processor's `main`;
* **Part III** — program architecture: the boot re-run on our own CPU core as an *oracle* (BLZ cross-check, the runtime memory map, the interrupt/IPC setup, the ARM9↔ARM7 rendezvous), and the overlay system that carries the game's 103 code modules.
* **Part IV** — the 3D model format: the game's *own* `.bmd` container (not the NITRO `BMD0` its extension suggests) and how its display lists and textures decode;
* **Part V** — level data and object placements: the level→overlay and settings-block tables, the object-table format and its spawner, the actor system, and the trace from each actor's create function to the model it draws.

Part VI (the SDAT music) is future work; the toolchain and container decoding carry straight over from the Mario Kart DS analysis.

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
| `.bmd` | 455 | 3D model — the game's *own* container (NOT NITRO `BMD0`; see Part IV), textures embedded |
| `.kcl` | 241 | **collision mesh** — the `KCL` triangle-soup/octree format the physics runs on |
| `.btp` | 105 | NITRO texture-pattern animation (frame-swapped textures) |
| `.narc` | 13 | `NARC` archive (bundles of the above) |
| `.sdat` | 1 | the entire SDAT sound bank (SSEQ/SBNK/SWAR), 4.4 MiB |

Two things distinguish this catalog from Mario Kart DS's. First, the **naming**: Super Mario 64 DS uses the terse extensions `.bmd`/`.bca` where the later Mario Kart DS writes `.nsbmd`/`.nsbca`. The suffixes *look* like an older spelling of the same NITRO formats — and `.bca` animation is — but the `.bmd` model is **not** NITRO `BMD0` at all: it is the game's own bespoke container (an `LZ77`-tagged LZ10 stream over a fixed header of bone/display-list/material/texture/palette arrays), decoded in Part IV. Only the low-level pieces inside — the GX display lists and the DS texture formats — are the shared NITRO silicon. Second, there is **no separate texture-set extension** (`.nsbtx`): with 455 models and no loose texture files, the models carry their textures **embedded**, which the later graphics Part decodes. The abundance of `.kcl` (241 collision meshes — one per world, object and platform) foreshadows a game that is, at heart, a physics playground.

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

Both binaries load into the same 4 MiB main RAM (the ARM9 at `$02004000`, the ARM7 near the top at `$02380000`) and coordinate through the DS's **inter-processor communication** hardware in the `$04000000` I/O block — the `IPCSYNC` mailbox and the `IPCFIFO` queue — with main RAM carrying the bulk payloads: input and touch state up from the ARM7, sound commands and DMA lists down from the ARM9. Part III picks up exactly here, pinning where and how they first synchronise.

# Part III — Program architecture

Part II followed the boot to each processor's `main`; Part III is about the machine that `main` builds — the runtime memory layout, the OS/interrupt scaffolding, the point where the ARM9 first has to talk to the ARM7, and the overlay system that pages the game's code in and out. A retail DS game is a large C++ program that dispatches almost everything through function pointers and virtual tables, so a purely static trace fans out into hundreds of unreachable indirect calls. This Part therefore leans on the repository's standard technique — an **oracle**, the game's own code run on our own CPU core — to establish structure from behaviour.

## 1. The oracle: running the boot on our own core

`supermario64ds/extract`'s `bootoracle` loads the *compressed* ARM9 binary to `$02004000` exactly as the BIOS would, points the `tools/arm` core at the entry `$02004800`, and lets it run: a flat memory for RAM/TCM/WRAM, the handful of BIOS `SWI`s the startup needs (`CpuSet`/`CpuFastSet` moves, `WaitByLoop`), CP15 accepted and ignored, and every write to the `$04000000` I/O block logged. It runs the real startup and reports what the code *did*.

Two things fall out immediately, each a verification rather than a guess:

- **The BLZ decompression is confirmed by the game itself.** After boot, the bytes the game's own `crt0` decompressor wrote into `$02004000`… are **identical** to what `tools/nds`' independent `DecompressBLZ` produces (`bootoracle` diffs them) over all code and data — the two agree bit-for-bit through `.bss` start, above which the `crt0` zero-fills. The Part II reimplementation and the real decompressor match exactly.
- **The startup runs cleanly on our core** — the decoder, CPU core, mode/bank handling and memory model execute millions of instructions of real ARM9 code (self-decompression, autoload, OS init) without a wrong turn.

## 2. The runtime memory map

Combining the module-params/overlay tables (Parts I–II) with what the oracle observes, the ARM9's runtime layout of the 4 MiB main RAM (`$02000000`–`$023FFFFF`) plus its tightly-coupled memories is:

| Region | Range | Contents |
|---|---|---|
| *(system-reserved)* | `$02000000`–`$02003FFF` | the low 16 KiB the ARM9 image is loaded *above* |
| ITCM | ~`$01FF8000` window | fast code, filled from the autoload list |
| ARM9 static code+data | `$02004000`–`$0209B000` | the decompressed ARM9 (`crt0`, `main` `$02007000`) |
| ARM9 `.bss` | `$0209B000`–`$020AA420` | zero-initialised statics (module-params `+0x0C`/`+0x10`) |
| ARM9 overlays | `$020AA420`–`$02148A80` | the 103 §I.4 overlays, banked into 22 address slots |
| ARM9 heap | above the overlays | dynamic allocation, up to the ARM7 region |
| ARM7 binary | `$02380000`–`$023A4B24` | the ARM7 image (its hot code runs from WRAM, Part II §5) |
| DTCM | `$023C0000`–`$023C3FFF` | fast data + the SVC/IRQ/System stacks (top-down from `+$3FC0`) |
| system-config | `$027FF000`… | the ARM9/ARM7 shared config block |

Against Mario Kart DS the shape is the same but the coordinates differ throughout: that game bases its ARM9 at `$02000000` and its DTCM at `$027E0000`; Super Mario 64 DS bases the ARM9 at `$02004000` and pulls DTCM down to `$023C0000`, just above the ARM7 image — a per-title arrangement, not a platform constant.

## 3. Initialisation and the interrupt/IPC setup

The single most useful thing the oracle extracts is the *exact* set of hardware registers the boot programs before it needs the second CPU — only **four**, and every one is about interrupts and inter-processor communication:

```
0x04000180 = 0x00000000   IPCSYNC     — clear our sync nibble / enable IPC-sync IRQ
0x04000184 = 0x0000C408   IPCFIFOCNT  — enable the IPC FIFO (send-clear, error-ack, enable)
0x04000208 = 0x04000000   IME         — master interrupt enable touched
0x04000210 = 0x00040000   IE          — enable exactly bit 18: IPC recv-FIFO-not-empty
```

The ARM9 enables **one** interrupt source before anything else: **bit 18, "IPC receive FIFO not empty"** — its entire early interrupt architecture exists to hear the ARM7, exactly as in Mario Kart DS (which additionally writes `IF` to acknowledge). The model underneath is standard DS: `IME` the master switch, `IE` the per-source mask, `IF` the write-1-to-clear latch, with the BIOS vectoring an IRQ through a handler pointer the runtime installs in DTCM.

## 4. The ARM9↔ARM7 rendezvous — before `main`

Here Super Mario 64 DS diverges sharply from Mario Kart DS. That game reaches `main` and its game-init and *then* blocks on the ARM7; Super Mario 64 DS **blocks during `crt0`'s SDK initialisers, before `main` (`$02007000`) is ever reached.** After ≈6.47 million instructions the oracle settles into a tight loop at `$0205BB54`:

```
0205BB48  LDR  r3, =0x04000180   ; IPCSYNC
0205BB54  LDRH r0, [r3]          ; read the sync register
0205BB58  ANDS lr, r0, #0x0F     ; the 4-bit value the *other* CPU posted
   …
0205BB9C  CMP  r0, lr            ; == the expected step value?
0205BBA0  BEQ  0x0205BB84        ; …keep polling…
```

This is `OS_SyncWithOtherProc` handshaking: the routine enables the IPC-receive interrupt (the `IE = 0x40000` above), writes the ARM9's outgoing `IPCSYNC` nibble, then spins until it reads the matching step number back from the ARM7's incoming nibble — ratcheting a short sequence so neither CPU races ahead during boot. The ARM9 has done its half and waits; the ARM7, per Part II §5, is meanwhile booting from `$02380000` and relocating into WRAM. With no ARM7, the single-CPU oracle stops here — the game's two boot halves meet at this point, and only past it do the `main` routines begin exchanging FIFO traffic.

## 5. The overlay system

The defining feature of this game's architecture is its **103 ARM9 overlays** (Part I §4) — where Mario Kart DS has four. They are the game's code, split into modules paged into main RAM on demand and banked into just 22 address slots, so at most a couple of dozen are resident at once. The load path is a NitroSDK filesystem routine the boot leaves in place, `FS_StartOverlay` (`$0205DD9C`), and tracing it confirms the overlay-record layout decoded in Part I and reveals how a module comes alive:

```
0205DE3C  LDRB r0, [r5, #0x1F]   ; the record's flag byte
0205DE40  ANDS r0, r0, #1        ; compressed?
0205DE48  LDR  r0, [r5, #0x04]   ; the overlay's RAM address
0205DE50  BL   0x020048D8        ; → BLZ-decompress it in place (same decompressor as the crt0)
0205DE54  LDR  r6, [r5, #0x10]   ; static-init (constructor) list start
0205DE58  LDR  r4, [r5, #0x14]   ; …list end
0205DE68  LDR  r0, [r6]          ; walk the list…
0205DE74  BLX  r0                ; …calling each constructor
```

So loading an overlay is: card-DMA its file into its slot, **BLZ-decompress in place if its flag bit is set** — using the *same* backward-LZSS decompressor (`$020048D8`) the `crt0` used on the whole ARM9, its only two callers being the self-decompression and this — then run the record's constructor list to register the module. The overlay *records* are read 32 bytes at a time from the ROM overlay table on demand rather than held in a resident array, so there is no static in-RAM table to read off; the choice of *which* overlay to load is made by the caller.

Which of the 103 overlays backs which part of the game — the castle hub, each painting-world, each minigame, the menus — is the natural next question, and reaching it means getting the ARM9 *past* the rendezvous the single-CPU oracle stops at.

## 6. The dual-core oracle: past the rendezvous

The rendezvous only blocks a *lone* ARM9, so the next tool is a **dual-core oracle** — both processors on `tools/arm` cores over one shared main RAM, wired by the DS's IPC hardware. It lives in `tools/nds/dsmachine` (game-neutral, for any DS title) and models the "full machine" the bare CPU core leaves to its caller: the shared 4 MiB main RAM and 32 KiB WRAM, each core's private TCM/WRAM and BIOS vectors, the cross-wired **IPCSYNC** mailbox and the two directional **IPC FIFOs**, a per-core interrupt controller, and the BIOS `SWI`s. `bootoracle` was single-core; `dualoracle` co-executes the two.

Watching the two cores run the handshake together resolves it. The routine is a mutual echo-and-count: the ARM7 posts a nibble that counts down `8,7,…,1,0`, reloading to 8 whenever the ARM9's echo lags; the ARM9 (at `$0205BB54`) echoes whatever the ARM7 posts and exits only when it reads a **0** after five or more rounds. The catch is timing — between each post the ARM7 spins in a BIOS `WaitByLoop`, and that delay is exactly what lets the other core's echo catch up, so the model has to *honour* `WaitByLoop` (yield to the other core) rather than skip it as a single-core trace would. With that, **both sync nibbles ratchet cleanly to 0 and the ARM9 clears `$0205BB54`** — the frontier the single-core oracle could not pass — and runs on into the post-sync **PXI** exchange (`$02059E48`), where it waits to *receive* the ARM7's boot message over the FIFO.

That is where the current model settles: the ARM7 reaches its idle loop, but the boot message it should send depends on its firmware/user-settings init read over **SPI**, which the machine stubs to zero. Modelling that ARM7 hardware (SPI, the RTC, the sound/power management the ARM7 owns) is what remains between here and the frame loop — and, with it, the overlay-to-state map. The dual-core scaffold that gets across the rendezvous is the reusable part; the rest is per-subsystem stubbing, and carries straight over to future DS titles.


# Part IV — The 3D model format

The catalog's 455 `.bmd` files *look* like the NitroSDK's `BMD0` models under an older file suffix — that is what Part I's first read assumed — but they are nothing of the sort: Super Mario 64 DS ships its **own model container**. A `.bmd` opens with an `LZ77` magic tag over a standard LZ10 stream; decompressed, there is no `BMD0` stamp and no NITRO resource dictionary, just a fixed header of `(count, offset)` pairs pointing at flat arrays:

| Section | Stride | Contents |
|---|---|---|
| header `+$00` | — | u32 scale: a power-of-two shift; raw fx4.12 vertices × 2^shift = world size |
| bones `+$04/+$08` | $40 | transform + a render list: count at `+$30`, material-index bytes at `+$34`, display-list-index bytes at `+$38` |
| display lists `+$0C/+$10` | **$08** | `{u32 subCount, u32 subHeaderPtr}` → subCount 16-byte sub-headers, each locating a GX command chunk (size `+$08`, data `+$0C`) |
| textures `+$14/+$18` | $14 | name, data offset, size, `texImageParam`; format-5 palette-select data follows the texels |
| palettes `+$1C/+$20` | $10 | name, data offset, size (BGR555) |
| materials `+$24/+$28` | $30 | name, **explicit texture and palette indices** (`+$04`/`+$08`), texture-matrix scale, GX polygon attribute (`+$24`) |

Only the low-level pieces are the shared DS silicon — the GX geometry display lists (including the full in-list matrix stack: push/pop/load/mult/scale/translate, which the larger stages drive) and the seven hardware texture formats — so those decoders carry over from the Mario Kart DS work unchanged (`tools/nds/nitro`); the container parser is this game's own (`extract/sm64ds/bmd.go`). Two traps mattered in practice: the display-list stride is **8 bytes, not 16** (a 16-byte read merges adjacent lists and scrambles which material draws what), and a display list's two GX chunks are one *continuous* command stream (a chunk may open with a delta vertex relative to the previous chunk's last).

# Part V — Level data and object placements

## 1. From level number to level data

Everything in this part was located by tracing the ARM9's own pointer tables — no signature scanning.

Two 52-entry tables in the ARM9 static data drive level loading:

* **level → overlay** at `$020758C8` (file `$718C8`): the loader at `$0202DED4` does `LDR r5, [$020758C8 + level*4]`, compares the entry against `-1`, and hands it to the overlay loader at `$02018028`. The shipped table is the identity mapping `[8, 9, …, 59]` — one overlay per level, which is *why* levels live in overlays 8–59.
* **level → settings block** at `$02092208` (file `$8E208`): the level-start code at `$0202D274` does `LDRSB r3, [currentLevel]; LDR r2, [$02092208 + r3*4]`, stores the pointer in a global and calls the level-data processor at `$020FE190` (overlay 2 — the always-resident engine overlay). The blocks sit at a different offset in every overlay; this table is how the engine finds them.

The **settings block** is 28 bytes. The processor at `$020FE190` consumes it field by field: `+$04` the *misc* objects table, `+$08`/`+$0A` the level model and collision map as u16 **internal file IDs**, `+$10` the area table (12-byte entries, `+$00` = that area's objects table) with the area count at `+$14`.

Internal file IDs are the game's own file namespace: a 2058-entry array of filename pointers at overlay-0 `+$13098` (`$020BD4B8`). Overlay 0's initializer at `$020AA420` loops over exactly `$80A` entries — the bound is a literal in its pool — resolving each path against the filesystem and registering *index → file*. Every file reference in the level data goes through this table.

## 2. Object tables and the spawner

An objects table is `{u16 count, u32 entries}`; each 8-byte entry is `{u8 type|layer<<5, u8 count, u16 pad, u32 list}`. The walker at `$020FE33C` decodes `type = b & $1F` and `layer = (b >> 5) & 7`, **skips entries whose layer differs from the current star** (layer 0 = every star), and dispatches through a 15-entry handler table at `$0210CBB8` — one handler per object type.

Two types carry the placements this analysis extracts:

* **Type 0 — standard objects** (handler `$020FE8AC`, 16 bytes each): u16 object ID at `+$00`, signed 16-bit x/y/z at `+$02/$04/$06` — each shifted `LSL #12` into fx20.12, so the short *is* the world coordinate — a parameter at `+$08`, the y-rotation at `+$0A` (standard DS angle-index units, `$10000` = 360°), another parameter at `+$0C` and the primary parameter at `+$0E`.
* **Type 5 — simple objects** (handler `$020FE960`, 8 bytes each): u16 at `+$00` packing `id & $1FF` (the mask is a literal in the handler) with a 7-bit parameter above it, then the same three position shorts. Trees, coins and other set-dressing use this compact form.

Both handlers translate the object ID through the **object → actor table** at `$0210CBF4` (u16 entries; 326 objects) before spawning — the object namespace in the level data is not the actor namespace the engine runs.

Across the 52 levels this decodes **4,350 distinct placements** (types 0 and 5, all star layers).

## 3. The actor system and its model bindings

The spawn call lands in the factory at `$02043098`: `LDR r0, [[profileTable] + actor*4]; BLX [r0]`. The boot code at `$0201A128` assigns that global its value — a static 326-entry **profile-pointer array at `$02090864`**. A profile begins with its *create function*; `+$04` carries the actor ID. Profiles for the always-loaded engine actors live in overlay 2; level-specific actors (IDs past the array) carry C++ RTTI in their overlay, whose typeinfo names the class — `daObjMc_Metalnet`, `daSBird`, `daMcFlag` — the only object *names* the cartridge contains.

Models reach actors through two statically visible mechanisms, and tracing them from each placed actor's create function (its body, its callees, and the methods of the vtable it installs) yields the model bindings:

* **File-handle slots**: `$02017ACC(slot, fileID)` — 26 call sites in the ARM9/engine overlay plus more in each level overlay's constructors — bind a BSS slot to an internal file ID at registration time; actor code then references its slot as an `LDR` literal. The metalnet constructor, for instance, registers its slot with file 1681 (`mc_metalnet.bmd`) and a second slot with the companion file 1682.
* **Direct loads**: a file ID (or a table of them) materializes as a literal and feeds a load call. The tree actor is the exemplar: its code at `$020EC240` computes `type = (par >> 4) & 7`, clamps it to 4, and indexes a **five-entry model table at `$0210ABB8`** — `bomb_tree`, `toge_tree`, `yuki_tree`, `yashi_tree`, and an archive-resident fifth entry (a bit-15-flagged ID) for the castle-grounds tree, which lives inside a NARC and is not yet extracted.

The static trace (`extract/sm64ds/actors.go`, report via `cmd/actortrace`) currently binds ~24 placed actors to their models — trees, signboards, mushrooms, switches, sea-weed, push-blocks and the level-local set-dressing — covering ~580 placements; the rest render as markers in the viewer until their create functions are traced further. Two calibrations tie placements to the exported GLBs: positions divide by 4096 (the exported GLBs store fx/4096 floats, so one GLB unit is one engine fx 1.0 — confirmed by 99.9% of all placements landing inside their stage's bounding box, against 72.9% for the initial /1000 guess), and object models place at **1/16 scale** — the same object-to-world ratio Mario Kart DS uses, and the size the tree's shift-4 bake cancels out to exactly.

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

# Part III — run the ARM9 boot on the tools/arm core as an oracle: BLZ cross-check,
# the I/O registers it programs, and the ARM9↔ARM7 IPCSYNC rendezvous it stops at
( cd extract && go run ./cmd/bootoracle -io "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )

# Part IV — decode every .bmd model and export GLBs (+ the viewer manifest)
go run ./cmd/exportbmd -all

# Part V — decode all 52 levels' object placements into per-stage JSON
go run ./cmd/exportlevelobjs
# and report the traced actor -> model bindings
go run ./cmd/actortrace

# Part III §6 — the dual-core oracle: both CPUs on shared RAM + IPC, clearing the
# rendezvous the single core stops at (into the post-sync PXI FIFO exchange)
( cd extract && go run ./cmd/dualoracle "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )
```

Toolchain (shared `retroreverse.com/tools`, this repository):

- **`tools/nds`** — the Nintendo DS Game Card container reader: header parse + CRC-16 verification, the FAT, the FNT directory-tree walk, the ARM9 overlay table, and the **BLZ** (backward-LZSS) decompressor the SDK applies to the ARM9 static module and every overlay. Shared with the [[Mario Kart DS]] analysis; makes no assumptions about the game inside.
- **`tools/nds/dsmachine`** — the reusable **dual-core DS machine**: two `tools/arm` cores over one shared main RAM + WRAM, per-core private TCM/WRAM/BIOS, the cross-wired IPCSYNC mailbox and the two IPC FIFOs, a per-core interrupt controller, and the BIOS `SWI`s. Game-neutral; the model any DS title's dual-core oracle builds on. (`arm.CPU.Exception`, added for the BIOS-style IRQ dispatch, is its one addition to the core.)
- **`tools/nds/cmd/ndsinfo`** — the container inspector for Part I (`-files`, `-tree`, `-grep`).
- **`tools/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware), with `tools/cmd/disarm` (linear) and `tools/cmd/codetracearm` (recursive-descent, follows ARM↔Thumb interworking) as CLIs.
- **`supermario64ds/extract/cmd/ndsextract`** — this game's extractor: writes `arm9.bin`/`arm7.bin` and the 103 overlays, and their BLZ-decompressed forms (`arm9_dec.bin`, `ovl9_NNN_dec.bin`), into `extracted/` (regenerable, git-ignored).
- **`supermario64ds/extract/cmd/bootoracle`** — runs the ARM9 boot on the `tools/arm` core over a flat DS memory (with the BIOS `SWI`s the startup needs): cross-checks BLZ against the game's own decompressor, logs the I/O registers programmed, and stops at the ARM9↔ARM7 `IPCSYNC` rendezvous. The DS analogue of the Amiga per-game oracles; the counterpart of Mario Kart DS's `bootoracle`.
- **`supermario64ds/extract/cmd/dualoracle`** — co-executes both boot binaries on the `tools/nds/dsmachine` dual-core model, clearing the rendezvous the single-core oracle stops at and running the ARM9 into the post-sync PXI exchange (Part III §6).

Rendered figures will go in `Super Mario 64 DS (DS)/rendered/`; annotated disassembly in `disasm/`.
