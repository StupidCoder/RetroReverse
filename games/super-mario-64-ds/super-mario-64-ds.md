# Super Mario 64 DS (Nintendo DS) — technical reference

**Image:** `Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds` — 16,777,216 bytes, MD5 `867b3d17ad268e10357c9754a77147e5`. Not committed (copyright); supply your own copy.

Super Mario 64 DS is a Nintendo DS launch title (2004) — a remake of the Nintendo 64 original that rebuilds the castle and its worlds on the DS's twin 2D/3D engines, adds four playable characters, a touch-screen map, and a suite of minigames. This document describes the *shipped cartridge* derived from the `.nds` image alone: no third-party emulator, debugger or disassembler, and nothing read from released source.

The toolchain is the `tools/platform/nds` cartridge reader (header, FNT/FAT, overlays, BLZ decompression) and the `tools/cpu/arm` ARM/Thumb disassembler and code-tracer.

Image: `Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds`, 16 MiB, game code **ASMP**, MD5 `867b3d17ad268e10357c9754a77147e5` (pinned in `../README.md`).

## Contents

* **Part I** — the cartridge image: the ROM header and its integrity checks, the two CPU binaries and the unusually large **overlay table**, the FNT/FAT filesystem, and the asset catalog;
* **Part II** — the boot chain: both NitroSDK `crt0` startup stubs, the ARM9's in-place BLZ self-decompression, and the handoff to each processor's `main`;
* **Part III** — program architecture: the boot re-run on the `tools/cpu/arm` core as an *oracle* (BLZ cross-check, the runtime memory map, the interrupt/IPC setup, the ARM9↔ARM7 rendezvous), and the overlay system that carries the game's 103 code modules.
* **Part IV** — the 3D model format: the game's *own* `.bmd` container (not the NITRO `BMD0` its extension suggests) and how its display lists and textures decode;
* **Part V** — level data and object placements: the level→overlay and settings-block tables, the object-table format and its spawner, the actor system, and the trace from each actor's create function to the model it draws.
* **Part VI** — collision: the `.kcl` mesh format (vertices, packed normals, prism records and the octree the queries walk), the external `CLPS` surface-attribute table, the ITCM query walkers — all pinned by running the game's own collision code in the oracle and reimplemented bit-exactly in Go.
* **Part VII** — the music: the `SDAT` sound archive (the same NitroSDK format as Mario Kart DS's, but shipped *with* its symbol block, so every tune carries the game's own name), rendered to MP3 through the shared `tools/platform/nds/sdat` sequencer and synth.
* **Part VIII** — the message system and the game's own course names: the `BMG` text containers, the font-derived character encoding, the level→course table the save system uses, and the per-level music binding — the cartridge's own course names.
* **Part IX** — **the machine**: the DS model grows from two CPUs and a stub into a console — timing, DMA, timers, the maths units, the nine VRAM banks, the cartridge port, the ARM7's SPI bus, and both graphics engines — until the game boots from cold, loads its 2,731 files, and draws its own title screen. Eight bugs, every one of which looked like something else.

Methods: purely static analysis of the `.nds` image. All addresses are 32-bit ARM addresses (`$02000000`-style main-RAM addresses, or the BIOS and I/O regions) unless a *file offset* into the ROM image is explicitly called out; bytes are little-endian.

# Part I — The cartridge image

## 1. The ROM dump

The image is a 16 MiB Nintendo DS Game Card dump. A DS card is not a memory-mapped cartridge like the Game Boy or Game Gear: it is a serial device behind a controller, and its contents are an on-card **filesystem** the console reads on demand. The DS is also a **two-processor machine** — an ARM9 and an ARM7 sharing one 4 MiB main memory — so the card carries two independent code images that boot side by side. The 512-byte header at the front of the image describes all of this to the BIOS.

## 2. The DS dual-CPU architecture and the address map

The header hands the two CPUs *different* load and entry addresses:

- the **ARM9** binary loads to `$02004000` in **main RAM** and starts at `$02004800`;
- the **ARM7** binary loads to `$02380000` — also inside main RAM, near its top — and starts there.

Both CPUs share the single 4 MiB main memory `$02000000`–`$023FFFFF`. Two things stand out. Super Mario 64 DS bases its ARM9 image `$4000` bytes above the base of main RAM, **leaving the low 16 KiB of main RAM (`$02000000`–`$02003FFF`) free** for the system to use, and (as with any DS title) the entry sits `$800` past the load base, over the secure-area landing.

The ARM9 is an ARMv5TE core and the ARM7 an ARMv4T core; both run the 32-bit **ARM** and 16-bit **Thumb** instruction sets, switched via `BX`/`BLX`, so disassembling either binary means tracking ARM-vs-Thumb state per address — exactly what `tools/cpu/arm` and `codetracearm` do. The ARM9 owns the card and drives the game and both screens; the ARM7 sits near the top of main RAM and services sound, the touchscreen, buttons and wireless, the two coordinating through the DS's **IPC** hardware in the `$04000000` I/O block.

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

Beyond the two main binaries, the ARM9 uses **code overlays** — separately loadable code+data blobs swapped into fixed RAM regions on demand, the DS's answer to fitting far more code than RAM holds. Super Mario 64 DS's overlay table at `$061510` (`$054` = `$CE0` bytes ⇒ **103 overlays**) is an array of 32-byte records. Decoding a sample:

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

`tools/platform/nds` joins the two tables: of the 2175 FAT entries, IDs **0–102 are the ARM9 overlays** (referenced directly by the overlay table, with no filesystem name) and IDs **103–2174 are the 2072 named files**. Walking from the root yields paths like `data/enemy/wanwan/wanwan.bmd` and `ARCHIVE/arc0.narc`.

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

Two features mark this catalog. The **naming** is terse: Super Mario 64 DS uses the extensions `.bmd`/`.bca` rather than the standard `.nsbmd`/`.nsbca`. The suffixes *look* like an older spelling of the same NITRO formats — and `.bca` animation is — but the `.bmd` model is **not** NITRO `BMD0` at all: it is the game's own bespoke container (an `LZ77`-tagged LZ10 stream over a fixed header of bone/display-list/material/texture/palette arrays), decoded in Part IV. Only the low-level pieces inside — the GX display lists and the DS texture formats — are the shared NITRO silicon. There is also **no separate texture-set extension** (`.nsbtx`): with 455 models and no loose texture files, the models carry their textures **embedded**, which the later graphics Part decodes. The abundance of `.kcl` (241 collision meshes — one per world, object and platform) foreshadows a game that is, at heart, a physics playground.

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

`sub_020049F0` is the CP15 sequence (`MRC`/`MCR p15`): disable caches, invalidate the instruction and data caches, drain the write buffer, program eight **MPU protection regions** (`c6, c0`…`c6, c7`), and enable the two **tightly-coupled memories** — **ITCM** (mapped so the low addresses just below main RAM, around `$01FF8000`, hit its fast 32 KiB) and **DTCM**, which Super Mario 64 DS bases at **`$023C0000`**, high in main RAM (the base is a per-title decision). It then re-enables caches/MPU/TCM by writing the control register.

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

The compressed-static end `$02061504` is exactly the load base `$02004000` plus the header's compressed ARM9 size `$5D504`. The image expands from `$02004000`–`$02061504` to a static code+data region ending at `.bss` start `$0209B000`, with `.bss` running on to `$020AA420` — precisely the base at which the first overlay loads. The far calls that follow the decompression (e.g. `BL 0x02019780`) land in valid code only *after* this step. BLZ is reimplemented from scratch in Go (`tools/platform/nds`, `DecompressBLZ`) — reversing the stream, running a forward LZSS, reversing the result — and the same routine decompresses all 103 overlays.

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

`main` at **`$02007000`** is the real game entry. It opens by calling a run of SDK/OS initialisers (`BL 0x02059788`, `BL 0x02059BC0`, …) — the subject of a future Part III. If it ever returns, `lr = $FFFF0000` drops the ARM9 into its BIOS.

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

Part II followed the boot to each processor's `main`; Part III is about the machine that `main` builds — the runtime memory layout, the OS/interrupt scaffolding, the point where the ARM9 first has to talk to the ARM7, and the overlay system that pages the game's code in and out. A retail DS game is a large C++ program that dispatches almost everything through function pointers and virtual tables, so a purely static trace fans out into hundreds of unreachable indirect calls. This Part therefore leans on the repository's standard technique — an **oracle**, the game's own code run on the `tools/cpu/arm` core — to establish structure from behaviour.

## 1. The oracle: running the boot on the CPU core

`supermario64ds/extract`'s `bootoracle` loads the *compressed* ARM9 binary to `$02004000` exactly as the BIOS would, points the `tools/cpu/arm` core at the entry `$02004800`, and lets it run: a flat memory for RAM/TCM/WRAM, the handful of BIOS `SWI`s the startup needs (`CpuSet`/`CpuFastSet` moves, `WaitByLoop`), CP15 accepted and ignored, and every write to the `$04000000` I/O block logged. It runs the real startup and reports what the code *did*.

Two things fall out immediately, each a verification rather than a guess:

- **The BLZ decompression is confirmed by the game itself.** After boot, the bytes the game's own `crt0` decompressor wrote into `$02004000`… are **identical** to what `tools/platform/nds`' independent `DecompressBLZ` produces (`bootoracle` diffs them) over all code and data — the two agree bit-for-bit through `.bss` start, above which the `crt0` zero-fills. The Part II reimplementation and the real decompressor match exactly.
- **The startup runs cleanly on the CPU core** — the decoder, CPU core, mode/bank handling and memory model execute millions of instructions of real ARM9 code (self-decompression, autoload, OS init) without a wrong turn.

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

Super Mario 64 DS bases the ARM9 at `$02004000` and pulls DTCM down to `$023C0000`, just above the ARM7 image — a per-title arrangement, not a platform constant.

## 3. Initialisation and the interrupt/IPC setup

The single most useful thing the oracle extracts is the *exact* set of hardware registers the boot programs before it needs the second CPU — only **four**, and every one is about interrupts and inter-processor communication:

```
0x04000180 = 0x00000000   IPCSYNC     — clear our sync nibble / enable IPC-sync IRQ
0x04000184 = 0x0000C408   IPCFIFOCNT  — enable the IPC FIFO (send-clear, error-ack, enable)
0x04000208 = 0x04000000   IME         — master interrupt enable touched
0x04000210 = 0x00040000   IE          — enable exactly bit 18: IPC recv-FIFO-not-empty
```

The ARM9 enables **one** interrupt source before anything else: **bit 18, "IPC receive FIFO not empty"** — its entire early interrupt architecture exists to hear the ARM7. The model underneath is standard DS: `IME` the master switch, `IE` the per-source mask, `IF` the write-1-to-clear latch, with the BIOS vectoring an IRQ through a handler pointer the runtime installs in DTCM.

## 4. The ARM9↔ARM7 rendezvous — before `main`

Super Mario 64 DS **blocks during `crt0`'s SDK initialisers, before `main` (`$02007000`) is ever reached.** After ≈6.47 million instructions the oracle settles into a tight loop at `$0205BB54`:

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

The defining feature of this game's architecture is its **103 ARM9 overlays** (Part I §4). They are the game's code, split into modules paged into main RAM on demand and banked into just 22 address slots, so at most a couple of dozen are resident at once. The load path is a NitroSDK filesystem routine the boot leaves in place, `FS_StartOverlay` (`$0205DD9C`), and tracing it confirms the overlay-record layout decoded in Part I and reveals how a module comes alive:

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

The rendezvous only blocks a *lone* ARM9, so the next tool is a **dual-core oracle** — both processors on `tools/cpu/arm` cores over one shared main RAM, wired by the DS's IPC hardware. It lives in `tools/platform/nds/dsmachine` (game-neutral, for any DS title) and models the "full machine" the bare CPU core leaves to its caller: the shared 4 MiB main RAM and 32 KiB WRAM, each core's private TCM/WRAM and BIOS vectors, the cross-wired **IPCSYNC** mailbox and the two directional **IPC FIFOs**, a per-core interrupt controller, and the BIOS `SWI`s. `bootoracle` was single-core; `dualoracle` co-executes the two.

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
| materials `+$24/+$28` | $30 | name, **explicit texture and palette indices** (`+$04`/`+$08`), texture-matrix scale, GX texture addressing (`+$20`), GX polygon attribute (`+$24`) |

Only the low-level pieces are the shared DS silicon — the GX geometry display lists (including the full in-list matrix stack: push/pop/load/mult/scale/translate, which the larger stages drive) and the seven hardware texture formats — so those decoders carry over from the Mario Kart DS work unchanged (`tools/platform/nds/nitro`); the container parser is this game's own (`extract/sm64ds/bmd.go`). Two traps mattered in practice: the display-list stride is **8 bytes, not 16** (a 16-byte read merges adjacent lists and scrambles which material draws what), and a display list's two GX chunks are one *continuous* command stream (a chunk may open with a delta vertex relative to the previous chunk's last).

## The .bca skeletal animation

The 493 `.bca` files are the game's bone animations (`kuribo_walk.bca` and six siblings animate the goomba), and the format decodes entirely from the engine's own player. The runtime applier at `$02045394` walks the model's bone tree — the same relative parent/sibling/child links the `.bmd` bones carry — and per bone decodes a **0x24-byte track set** at `[anim+$14] + boneIndex × $24` (`$0204547C`). A track is four bytes, `{u8 rate, u8 animated, u16 index}`, and a set is nine of them: scale x/y/z (values in the fx32 array at `[anim+$08]`), rotation x/y/z (u16 array at `[anim+$0C]` — a value is `angle >> 4` and is shifted left four bits into DS angle units, exactly like the `.bmd` bone rotations), translation x/y/z (fx32 array at `[anim+$10]`). The sampler (`$020456A0` for u16, `$020457F0` for fx32) has three modes: `animated == 0` → the one value at `index`; `rate == 0` → one key per frame; otherwise keys every `2^rate` frames, linearly interpolated, with the run past the last full key stored per-frame. The header is `{u16 numBones, u16 numFrames, u32 loop, u32 scaleOff, u32 rotOff, u32 transOff, u32 trackSetsOff}`, wrapped in the same `LZ77`-magic LZ10 stream as the models. The decoded per-frame TRS feeds the same bone compose as the bind pose (`$02045074`: `R = Rx·Ry·Rz` in row-vector order, then the parent chain).

A bone's flag word at `+$3C` carries two decoded bits: bit 0 marks a **camera-facing (billboard) bone** — set on the bob-omb's `body_bill` and on the tree quads, clear on roots and feet — the same flag the engine's bone compose treats specially; the sub-header words `+$00`/`+$04` are the **matrix-slot list** (`{u32 count, u32 byteListOffset}`, through the u16 map at header `+$2C`): a display list's `MTX_RESTORE` indices are slots, and slot *k* belongs to bone `map[list[k]]` — the goomba's list is `[1,0,2]` (slot 1 the body, 0/2 the feet). The exporter seeds its matrix slots from that list, skins vertices by the mapped bone, and exports billboard bones as glTF node extras the viewer orients to the camera each frame.

Checked against the shipped data: `kuribo_walk` is 3 bones × 30 looping frames whose animated rotation tracks hold 16 keys each (`rate = 1`: every 2nd frame plus one) — a stride cycle on the legs — and `kuribo_wait`'s root keeps the bind pose's 90° Z rotation while bobbing on translate-Y, a consistency check between the two formats. Decoder: `extract/sm64ds/bca.go` (`cmd/bcainfo`). Playing these in the web viewer needs a skinned export — joints and weights per vertex instead of baked bone transforms — which is the next exporter step.

Texture *addressing* is per material, not per texture: the render-object initialiser at `$02046374` ORs the material's `+$20` word — GX repeat bits 16/17 and **flip (mirror) bits 18/19** — onto the texture record's format/size param to form the `TEXIMAGE_PARAM` the hardware sees. The castle hall's sun emblem is the visible proof: its `mat_sun` sets flip on both axes, so the floor quad's four quarters mirror into one complete emblem. The same word also settles the tree-billboard texcoord puzzle: the tree materials clear all four bits — hardware **clamp** — so their famously overflowing quad texcoords (t = −14.75..64 on a 64-texel texture) simply pin to the texture's edge rows (transparent, on a cut-out), no engine-side remap involved. The exporter's earlier UV-normalise convention for billboards is deleted; the addressing bits are exported as glTF sampler wrap modes instead.

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
* **Type 1 — entrances** (handler `$020FE6C8`, 16 bytes each): the same signed-short position triple at `+$02/$04/$06` (`LSL #12`) and yaw at `+$0A`; the level's first entry is where the player spawns. The viewer stands the playable Mario (the 16-bone `MG/mario_model_mg.bmd`, sharing the `data/player` skeleton and its `su_wait` idle clip) at that point.
* **Type 5 — simple objects** (handler `$020FE960`, 8 bytes each): u16 at `+$00` packing `id & $1FF` (the mask is a literal in the handler) with a 7-bit parameter above it, then the same three position shorts. Trees, coins and other set-dressing use this compact form.

Both handlers translate the object ID through the **object → actor table** at `$0210CBF4` (u16 entries; 326 objects) before spawning — the object namespace in the level data is not the actor namespace the engine runs.

Across the 52 levels this decodes **4,350 distinct placements** (types 0 and 5, all star layers).

## 3. The actor system and its model bindings — the actor oracle

The spawn call lands in the factory at `$02043098`: `LDR r0, [[profileTable] + actor*4]; BLX [r0]`. The engine init at `$0201A128` assigns that global its value — the **profile-pointer array at `$02090864`**, 783 entries long (the factory's own `*ERR*` string ends it). A profile begins with its *create function*; `+$04` carries the actor ID. Profiles for the always-loaded engine actors live in overlays 1–2; enemy and level actors' entries point into the banked overlay that carries their code — meaningful only while that overlay is loaded (several banks share each RAM slot). Some profiles carry C++ typeinfo whose mangled name — `daKrb_c`, `daObjMc_Metalnet`, `daWanwan2_c` — is the only object *name* the cartridge contains.

Which model belongs to which actor is answered by **running the actors' own code** — the actor oracle (`extract/sm64ds/oracle.go`, driven by `cmd/actororacle`). An earlier revision pattern-matched the create functions for file-ID literals; it misbound look-alikes (a stone Eyerok hand under Bob-omb Battlefield's bridge) and needed hand-kept exception lists, so it was replaced wholesale. The oracle:

1. **Boots the real game.** The compressed ARM9 runs from its entry on the `tools/cpu/arm` core over a flat DS memory image — the crt0 self-decompresses, builds the TCM sections and the pre-`main` heap — and then `main()` itself runs up to (not into) its frame-loop call at `$02007040`, so OS arenas, the root heap (`$0203CB04`), and the engine init (`$0201A054`) happen in the game's own order. The handful of places that need the ARM7 or the card are trapped at function level: the IPC rendezvous and PXI waits, the sound system's channel-`$B` commands, the save-file read, and the GX status waits that only an interrupt handler would clear.
2. **Runs the overlay initialisers natively.** Overlay 0's initialiser at `$020AA420` resolves all `$80A` internal file paths (the FNT answers the path→ID walk at `$020189F0`) and builds the internal→FS remap table; then overlays 1+2 load and their 28 constructors run, registering every file-handle slot (`$02017ACC`) and queueing the common-model preloads. The queued loads are serviced the way the game's loader thread would — the async callbacks (`$02017AB4`) only publish the slot to the thread's mailbox, so the oracle performs the acquire (`$02017BC4`) itself — which preloads the shared pool: coins, stars, `?`-blocks, shells, signs, Peach, the power flowers…
3. **Traps the loaders and serves real files.** The load-by-internal-ID function (`$0201818C`) and the slot loader (`$02017C54`) are replaced: the requested ID is recorded and the actual extracted file (archive members decompressed, exactly like the real branches) is served into scratch RAM, so the relocation and parse code after each load runs on real data.
4. **Spawns every placed actor.** For each level (its overlay loaded next to 1+2) and each enemy bank, every distinct `(actor, parameters)` combination the levels place is run: the factory's context store (`$02043180`) is called with the placement's parameters, then the profile's create function, then the new object's init (vtable slot 0), then the async queue is drained. Between runs the machine state rolls back by dirty-page restore.

Three signals, all from the game's own code paths, make a binding: a **load** (the actor's code pulled the file in), an **acquire** (it took a reference on an already-loaded slot), and — strongest — a **display** (it built a render object on the model: the wrapper get at `$02016E70` or the render-object sizing at `$02046564`, whose model pointer the oracle maps back to the served file). The first display/load outside the engine's own preload pool is the actor's model; `.bca` requests are its animation clips (the goomba: `kuribo_model.bmd` plus all seven `kuribo_*.bca`). Parameter-dependent actors fall out for free: the tree actor run with each placed `par1` loads the right `bomb_tree`/`toge_tree`/`yuki_tree`/`yashi_tree`/archive-tree member; the `?`-block actor 20 loads its per-parameter contents.

The result (`extracted/actorbind.json`, consumed by `cmd/exportlevelobjs`) binds **300 placed actors — 4,048 of 4,350 placements**, with no heuristics and no exception lists: every per-level mechanism the static scan missed (lifts, shutters, the pirate ship, Bowser's bridges, the fifteen `fl_14_xx` puzzle pieces of Lethal Lava Land), the collectibles by their real archive members (coin `arc0_5`, star `arc0_21`), the chain chomp's `ar1_2` body + `ar1_1` chain with `ar1_3`/`ar1_4` as its clips, and the two actors the old scan had misbound to the Eyerok hand — actor 11 is the **box switch** (`obj_box_switch`), actor 12 the **star switch** (`obj_star_switch`). Actors with no recorded request provably load nothing in create/init (cameras, triggers, spawners of other actors).

## 4. Archives: the flagged file IDs

Internal file IDs with **bit 15 set** name files inside archives. The resolver at `$020186C0` is explicit about it: `CMP id, #$8000`, then a walk over a 13-entry descriptor array at `$0208ECF4` (stride `$14`, loop bound `#$D` at `$0201874C`). Each descriptor carries the archive's flagged-ID range as two u16s at `+$08`/`+$0A` and two strings — a short name and the full `/ARCHIVE/<name>.narc` path; the member index is simply `id - first`. The shipped ranges step by `$400`: `$8000` = `arc0.narc` (198 members), `$8400` = `en1`, `$8800`–`$9400` = `vs1`–`vs4`, `$9800` = `c2d`, `$9C00` = `ar1` (17 members), `$A000`+ = the per-language archives.

That closes the tree table's fifth entry: `$9C10` = **ar1.narc member 16**, whose only material is named `mat_main_tree` — the castle-grounds tree, a two-triangle billboard like its filesystem siblings. The same archives hold the collectibles, self-identified by their material names: arc0 member 5 is the **coin** (materials `coin`/`coin_a`; member 7 the red variant), member 21 the **power star** (`mat_body`/`mat_eye` with embedded `star` strings), member 25 the **silver star** (`mat_star_silver`). The coin's load path is now code-traced too: a bulk registration block at `$021005E0` binds file `$8005` to the file-handle slot at `$0210D9B8` (and `$8007`, the red coin, to `$0210D9F8`). Two actor families read it: the placed coins (actors 288–290, creates `$020B2B48`/`$020B2AF0`/`$020B2A98`) and the shared **item actor 276** (create `$020B0580`) whose init at `$020B01C0` selects a model by subtype — `param & $F`, where subtypes `$B`/`$C` are the coin variants and the rest the mushrooms; the oracle's binding of actor 276's mushroom subtypes to `scale_up_kinoko` is that same table's other rows.

## 5. The render transform — traced end to end

Placement data is fully traced: the position shorts are fx world coordinates (`LSL #12` in the spawners at `$020FE8AC`/`$020FE960`), rotations are DS angle-index units, and the object→actor and model bindings above all come from the game's own tables and code. The render transform — how authored fx vertices and an actor's world position reach the geometry engine — is now traced end to end as well.

One correction first. An earlier revision of this section placed the trace at a "model-object creator" `$02020994`. That function is nothing of the kind: it decodes OAM attribute bit-fields (a signed-byte Y, a 9-bit X that wraps at `#$100`, shape/size crumbs) into one of two 128-entry, 8-byte shadow-OAM arrays at `$0209E674`/`$0209EA74` (main and sub screen), building affine parameters from the sine table at `$02082214` when a rotation or non-1.0 scale pair comes in. It is the **2D sprite submitter** — and actor 335, whose init at `$020F9E98` feeds it while culling against the 256×192 screen with a 16-pixel margin and tracking the four playable characters, is the **bottom-screen minimap**, not the visible level.

The 3D path lives in the ARM9 engine library — no geometry-port literal appears anywhere in overlay 2:

* **Draw entry `$020443C8`** `(renderObj, baseMatrix, scaleVec)`. The render object is `{bmd header, materials, bones, bone matrices}`. Its first act is `[$020A4BD4] = 1 << (header.byte0 + 12)` — **the .bmd header shift is real: every model draws under a `MTX_SCALE` of 2^shift**, written to the scale port for every bone by the compose function below. `baseMatrix` is published to `$020A4BD0`. It then walks the model's bones (stride `$40`) and, for each render-list entry, applies material `+$34[i]` (`$02044B30`) and runs display list `+$38[i]`, skipping materials whose polygon-attribute bit 31 is set.
* **Bone compose `$0204488C`**: per bone, in matrix mode 2 it loads identity and `MTX_MULT_3x3`s the base then the bone matrix — a product that only survives in the **vector (normal) matrix**; a bone with the billboard flag (`+$18`) first has its 3×3 rows replaced by their norms (`$020458A8`, hardware SQRT per row), killing its rotation for lighting. Then in mode 1 it rebuilds the **position matrix**: `MTX_LOAD_4x3(base)`, optional `MTX_SCALE(scaleVec)`, `MTX_MULT_4x3(bone)`, `MTX_SCALE(2^shift)` — vertices see 2^shift, then the bone transform, then the caller's scale vector, then the base matrix — and `MTX_STORE`s the result in the bone's matrix-stack slot for the display lists to restore. Display-list bytes stream out through `$0205A358`; the helpers `$0205536C`/`$02055388`/`$020553A4` are one-word wrappers pushing GX commands `$1A`/`$19`/`$17` (`MTX_MULT_3x3`/`MTX_MULT_4x3`/`MTX_LOAD_4x3`).
* **The base matrix is CPU-composed.** The model-wrapper class (ctor `$02016D58`, vtable `$0208E90C`) embeds a 4x3 world matrix at `+$1C`; its draw method (vtable `+$14` = `$02016B78`) multiplies the camera matrix at `$0209B3EC` onto it (`$02052914`) and calls the draw entry with the caller's scale vector. Its `+$10` method copies a new world matrix in and re-walks the bone hierarchy (`$02045074`).
* **Render space is world ÷ 8.** The idiom is `ASR #3` at every world→render seam: the coin builds its wrapper translation as `position >> 3` (`$020AF4EC`), and the stage's frame draw copies the camera position `>> 3` into the skybox matrix translation (`$0202B940`). A placement short (world fx `short × 4096`) therefore lands in a render matrix as `short × 512`.
* **The stage** is owned by `dScStage_c` (its RTTI string sits at `$02092198`, vtable `$020921BC`). The loader `$0202B5EC` reads the settings block's `+$08` u16 and loads the level model into an embedded wrapper at `+$86C`; on load, every material whose polygon-attribute alpha isn't 31 has its polygon ID forced to `$13` (translucency grouping), and the render object is published to `$0209F320`. A sibling loader `$0202B0FC` reads settings `+$18` bits 4–8 as a **skybox index** into the u16 file-ID table at `$02075620` (files 2040–2050, `data/vrbox/vr01`–`vr11`) and hangs that wrapper at `+$9BC` — the index is 1-based (`SUB r1, idx, #1`) and **0 means no skybox**: the loader returns before creating the wrapper, which is every indoor level. Each frame the draw copies the camera position (`>> 3`, world→render) into the wrapper matrix's translation and calls the wrapper's virtual draw with a NULL scale vector, so the sky is a camera-centred dome at the standard object size — never rotated. The vrbox models are authored at **shift 13** (vertices to ±65536), which the exporter's artifact guard must admit. The frame draw `$0202B8A4` sets fog, moves the skybox to the camera, then `$0202B164` walks the stage bones **as areas** — toggling polygon-attribute bit 31 to hide every area but the current one — and draws the stage wrapper with base = camera (its world matrix stays identity) and one hard-coded scale vector, **`$020755D4` = `{$7D000,$7D000,$7D000}` — uniform 125.0**.

Putting the units together: a stage vertex renders at `v × 2^stageShift × 125` render units while a placement short renders at `short / 8`, so **one stage-vertex unit is `2^stageShift × 1000` world-fx units** — the stages are authored in kilo-units. In the exporter's baked-GLB space (vertices × 2^shift) that collapses to `GLB position = short / 1000`, *independent of the stage shift*, and an object model drawn with a NULL scale vector (the common case — the coin passes NULL) stands `bakedGLB / 125` stage-GLB units tall; the engine object's scale vector at `+$80` enters as the `scaleVec` argument when an actor passes it.

That rule holds against the shipped data once the decoder seeds the DS's 32 addressable matrix slots from the bone matrices: the engine keeps each bone's *composed* matrix in those slots — the `MTX_STORE` in the bone compose runs after `MTX_SCALE(2^shift)`, so a stored slot always carries the model scale. Nearly every display list begins with `MTX_RESTORE`, which loads that scaled slot; seeding the slots to identity instead drops the 2^shift bake — half size on a shift-1 stage, 1/16 on a shift-4 model. With slots seeded from the bone matrices, all 48 stages contain their placements under the ÷1000 rule except five where objects legitimately leave the visible geometry: the open-sky wing-cap level (`habatake`, rings of items in the air), the secret slide's off-tube course markers, the two Bowser arenas (mines placed past the platform rim) and the castle exterior's over-water markers.

Two object systems bypass the wrapper path. A **shadow/effect list** at `$0209CEF4`, drawn by `$02015D38`: each node carries its own render object, a matrix pointer and an alpha byte; the walker builds a translation matrix from the node position (+2.0 render units on Y, clearing the ground), multiplies the node's matrix, sets the material alpha (`$02046120`) and draws two passes through `$020443C8`. And the **billboard set-dressing** (the trees): the tree placer `$020EC22C` creates no model wrapper at all — it allocates a `$4C` node holding `position >> 3` (Y `+$1E000`), chains it per model type, and registers it into the sorted billboard list at `$0209CEE8` with three constants — `$35555` (53.33, an added cull radius), `$1F4000` (500.0, a vertical cull range) and a layer mask — that the sorter `$02014AA8` consumes for distance and height culling and depth ordering; the quad's camera-facing orientation comes from the node's matrix pointer, not from the bone billboard flag.

## 6. Actor behavior — the first traces

Actor behavior is C++ virtual methods, not bytecode scripts: every profile's create function installs a vtable whose slots the engine calls each frame — `+$00` init, `+$18` step, `+$24` draw is the pattern across every actor examined (coin, item, signpost, minimap, goomba). What each actor *does* lives in its step, and three of them are traced:

* **The coin spins by `+$C00` yaw per frame.** The placed-coin step at `$020B2324` (vtable `$021087EC`, actors 288–290) adds `#$C00` to the yaw short at `+$8E` every tick — `$10000` is a full turn, so at the 30 fps actor tick that is ~1.4 revolutions per second — then runs pickup logic (`$020B12EC`/`$020B14D8`) and links its blob-shadow billboard node (`+$178`) into the draw list only when the camera is within 100.0 render units (`$020536E4` distance test against `#$64000`). The item actor's coin subtypes rotate the same field and build their wrapper matrix from it in `$020AF4EC` (`$0203BD6C`: sine-table rotation about Y, then translation `position >> 3`). The spinning-coin look is geometry, not a texture animation — a flat quad model yawing in 3D.
* **The signpost is a proximity dialog.** Its init (`$020BC240`, actor 184) loads the model wrapper from a file-handle slot, then **snaps to the ground**: a collision ray is cast from `position.y + $64000` downward (`$02037570`/`$0203748C`/`$02038F44`) and the hit height replaces the placement's Y. It registers an interaction cylinder (`$0201490C`) and its step at `$020BBEA4` watches the engine-set interaction flags in `+$B0` (bit `$4000` = player in range) — when the dialog pointer at `+$59C` is live it starts the message through `$020BB060` and parks an "in dialog" bit (`$4000000`). The message text lives in `data/message/msg_data_<lang>.bin` — decoded in Part VIII (all 711 messages, including every signpost).
* **The goomba is a bank actor.** Overlay 84 is an enemy bank carrying RTTI for `daKrb_c` (Kuribo), `daRedBombhei_c` and the piranha plants; the goomba's profile at `$02130924` (create `$0212BFF8`, actor 202) also carries per-enemy tuning — a 100.0 sight radius and 200.0 active radius. Its model bundle registers `kuribo_model.bmd` (internal file 902), a low-LOD variant (910) **and skeletal animations** — `kuribo_walk.bca` (909) among seven `.bca` files — through an animated-wrapper class (`$02016958`) the static coins never use. The step at `$0212B6EC` gates on the ARM9 enemy base class (`$02005FA0`), tests player interaction (`$0200F70C`) and routes stomp/hit responses through `$020AD838` with a `#$6C000` bounce impulse. Animating a walking goomba in the viewer therefore needs the `.bca` bone-animation format and a skinned exporter — the next format on the list.

The goomba's **wander AI** is traced from its state handlers (installed into the table at `$02130D74` by the bank's static init): state 0 (`$0212B2DC` → `$0212ABD4`) eases the forward speed at `+$98` toward the per-state value in the table at `$02130248` (2.0 world-units per frame when wandering, 8.0 chasing) by `$500` per frame, and eases the yaw toward a target heading at **`$200` angle-units per frame**. A repick timer (`+$450`) runs 100 frames: on expiry a call to the shared RNG (`$0203B990`, upper 16 bits) turns the goomba by a **uniformly random signed 16-bit angle** three times out of four and pauses it the fourth. The collision probe (`$020AE244`) reflects the heading into `+$45A` on wall contact, a **1000.0-unit leash** (`$3E8000`, checked against the spawn point kept at `+$41C`) steers it home, and falling out of the world (`+$113 ≥ 6`) teleports it back to that home point. State 3 is the chase (`$0212AF74`), entered from the profile's 100-unit sight radius.

The **chain chomp** (`daWanwan2_c`, overlay 100) resolves its previously-unbound model through the castle-grounds archive: it loads `ar1` members 1–4 — member 2 the body (`a_mat_body`/`a_mat_eye`/`a_mat_mouth`), member 1 the chain link. Its step (`$02143D64`) runs under a heavy `-$3C000` (−15.0) gravity, eases the actor scale vector at `+$80` toward 1.0 (the pre-bark inflate), and its lunge state drives the forward speed to **`$17000` — 23.0 units per frame**; the chain drawer (`$021437D4`) strings the links from a **`(0, 0, −250.0)`** anchor vector rotated by the body's orientation — a 250-unit chain to the stake. The viewer renders the chomp with its chain links strung to the spawn stake and lunges at the traced speed inside the traced radius (the lunge cadence is approximated, not traced).

**The stake is a spawned child** — and pinning that down took a wrong conclusion first. The placed Bob-omb Battlefield chomp is `daWanwan_c` (actor 219) in **overlay 14, the level's own overlay** (each level overlay carries its level-specific actors: here also the gate `daObjWanwanShutter_c` and the cannon-hole cover). An initial sweep found no stake anywhere — no pile placement at the anchor, no fifth file in the chomp's slot loads, no actor-factory calls in the overlay, plain `mat_grass` in the stage mesh and flat floor in the level `.kcl` — and this document briefly claimed the remake had dropped the N64 stake. A gameplay screenshot disproved that, and the miss was in the sweep, not the game: child actors are spawned through the helper **`$02010E2C`** (spawn actor ID at position), not through the `$020430xx` factory entry points the sweep had covered. The chomp's init does exactly that at **`$02112C6C`**: `MOV r0,#27; MOV r1,#$11; ADD r2,r4,#$5C; BL $02010E2C` — **spawn a pile (actor 27), parameter `$11`, at the chomp's own anchor** — then keeps the child pointer at `+$608` and sets a mode byte on the pile (`+$320` = 1), the chomp-stake behavior the four ordinary 2×2-grid piles elsewhere in the level don't get. Pound it three times and the chomp breaks the gate. The same `$02010E2C` scan legitimizes the level's other spawn sources: the type-6 object-table records spawn the four cannons (actor `$15B`) and `daWanwan2_c` (the star-3 and castle-grounds chomp, overlay 100) spawns items, stars and its iron-ball sub-actor 220 — but no stake, and none of this is visible to the actor oracle, whose bare environment kills child spawns (the "terminate called — resumed" notes on the chomp's runs are exactly those). `cmd/exportlevelobjs` emits the traced child, so the viewer now plants the stake at every chomp anchor. Two lessons, recorded: a "nothing draws here" conclusion is only as strong as the list of spawn mechanisms it covered, and a user with the real game beats a confident negative.

The **bob-omb** (`daBmb_c`, overlay 102) wanders differently: its heading picker (`$0214BEB4`) aims **at its home point** (`$0203B7AC` toward the anchor at `+$3C4`) **plus a random signed 16-bit offset** — erratic but home-biased — and beyond **1280 units** (`$500000`) it drops the randomness and heads straight back. Each pick sets the forward speed to **`$5000` (5.0 units/frame)** and the walk repicks when the yaw reaches the target heading (`$0214BE8C`) or when a 512-frame fallback timer (`+$3E8`, set at the state entry `$0214C108`) expires; the yaw eases at **`$400` angle-units/frame** (doubled to `$800` when chasing the target at `+$38C`, whose speed goal becomes `$10000` = 16.0 — the lit-fuse sprint). A nice detail: the walk-cycle playback rate at `+$35C` is the forward speed divided by 8, so the feet match the ground speed. The viewer gives bob-ombs (and the red buddies) this profile alongside the goomba's.

The viewer replicates what is traced: coins spin at the engine's `$C00`-per-frame rate, clicking a signpost pops the traced interaction — and, since Part VIII, quotes the sign's own in-game words — skinned enemies play their `.bca` clips, and the goombas **patrol** with the traced wander — 2.0 units/frame, `$200`/frame turning, 100-frame repicks at 75% turn / 25% pause, leashed to their spawn points.

# Part VI — Collision: the `.kcl` mesh, the octree walk and the `CLPS` surface table

The render mesh never sees a physics query. Every level (and every solid platform) pairs its `.bmd` display model with a `.kcl` **collision mesh** — 241 of them in the catalog — and the actors probe that through a small family of engine walkers. This part pins the file format and the queries, with a new kind of instrument: instead of only *reading* the code, `cmd/kcltrace` **builds the collision world inside the oracle with the game's own routines and casts the game's own rays through it**, while a read watch on the served file logs every byte the walker touches. The read log gives the structure; the flagged PCs name the code whose disassembly gives the semantics; and a Go reimplementation is then verified bit-for-bit against the running original.

## 1. From level load to a registered collider

Part V's level-data processor (`$020FE190`) consumes the settings block's collision-map file ID at `+$0A`:

* the internal-ID loader (`$0201816C`) fetches the file — `.kcl` files are stored `LZ77`-tagged on card and come out of the loader decompressed (`$02017D84`);
* `$02039760` **fixes the header up in place**: the file's first four u32s are section *offsets*, and each gains the buffer base to become a pointer — the whole format is position-independent and mutated into its runtime form in four adds;
* the fixed-up pointer is stored at `+$20` of a collider object, and `$020396F0` seeds the rest of the context (the settings block's `+$00` pointer — the `CLPS` table, §4 — goes to `+$24` via `$0203821C`);
* `$02039184` registers the collider in a flat **24-slot pointer table at `$020A0C80`** (first free slot; the level map, loading first, is always slot 0).

The collider is a C++ object, and the binary's own RTTI names the family: **`dBgW`** (the base — "background wall"), **`dBgW_Kc`** (the KCL-backed collider, vtable `$020993DC`, constructor `$020398C8`), and the moving/scaling variants **`dBgW_KcMbg`** and **`dBgW_KcMbgSclY`** that platform actors embed. The interesting vtable slots point into **ITCM** — the DS's zero-wait-state instruction TCM, where the engine parks its hottest code: `+$18` the down-ray walker (`$01FFD3F8`, §3), `+$1C` a segment/sweep walker (`$01FFB0FC` — it reads two vectors, `query+$38` and `query+$54`; the wall probes build these via the query class at `$020377B0`), `+$20` a sphere walker (`$01FFB830`, the one consumer of the header's *thickness* field), `+$14`/`+$10` the vertex and normal accessors (`$01FFD890`/`$01FFD8D8`) whose shifts betray the packed encodings below.

Queries go through the dispatcher at `$02038F44` (the routine the signpost's ground-snap calls, Part V §6): slot 0 first, then slots 1–23 — each *dynamic* collider gated by its owning actor's `+$B0` flags — with the query object collecting the best hit across all of them.

## 2. The file layout

After LZ77 decompression, a `.kcl` is a 0x38-byte header and four sections:

| Offset | Content |
|---|---|
| `+$00` | u32 → **vertex section**: 12-byte records, fx32 x/y/z in **world>>6** units (the accessor returns them `<<6`) |
| `+$04` | u32 → **normal section**: 6-byte records, s16 x/y/z in **fx.10** (1.0 = `$400`; returned `<<2` as fx.12) |
| `+$08` | u32 → **prism section**: 16-byte records (below); record 0 is a dummy — index 0 terminates leaf lists |
| `+$0C` | u32 → **octree section** (below) |
| `+$10` | fx32 **thickness** — how far behind a surface plane the sphere walker still counts contact (castle grounds: 80.0); the down-ray never reads it |
| `+$14` | fx32 ×3 **area minimum** x/y/z, pre-shifted into world>>6 units |
| `+$20` | u32 ×3 **coordinate masks** — `~mask` is the largest local coordinate (the area's extent in integer world units) |
| `+$2C` | u32 **root shift** — the descent's starting cell size (log2, in world units) |
| `+$30`/`+$34` | u32 **y/z root-index shifts** — how the root grid packs, so the walker derives nothing: the file carries its own indexing |

A **prism** — the format's collision triangle — is not three vertices. It is one vertex plus four normal indices and a length, the classic plane-based encoding:

| Offset | Field |
|---|---|
| `+$00` | fx32 **length** (the extent along the third edge's direction) |
| `+$04` | u16 **vertex index** — the triangle's anchor corner |
| `+$06` | u16 **face-normal** index |
| `+$08`/`+$0A` | u16 **edge-normal A/B** indices (planes through the anchor) |
| `+$0C` | u16 **edge-normal C** index (the far edge, bounded by *length*) |
| `+$0E` | u16 **attribute** — an index into the level's `CLPS` table (§4); the surface's behavior lives *outside* the mesh |

The **octree** section is a forest of s32 tables. The root grid is indexed by `(lz>>s)<<shiftZ | (ly>>s)<<shiftY | lx>>s` (s = root shift, local coordinates in integer world units relative to the area minimum); each word ≥ 0 is a **child-table offset relative to the table holding it**, descending one halved cell per level with the child picked as `xbit | ybit<<1 | zbit<<2`; a word with **bit 31 set** points (same relative convention, low bits) at a **leaf**: a 0-terminated u16 prism-index list starting 2 bytes in. Note the asymmetry — the root packs z highest, the children pack z highest *of three fixed bits* — both spelled out by the code, neither derivable from the other.

## 3. The down-ray walk (`$01FFD3F8`)

The ground query — "the highest floor under (x,y,z)" — works in three nested precisions: fx20.12 world coordinates arrive, are cut to **fx.6** (`>>6`) for all plane math, and to **integer world units** (`>>6` again) for cell addressing. The walker:

1. bounds-checks the local coordinates against the masks — x or z outside means *miss*, but **y above the area clamps** to the top instead (a ray from the sky still finds the ground);
2. descends the octree to the leaf under the point and tests every prism in its list:
   * face normal `y ≤ 0` → not a floor, skip; `|y| ≤ 8` (fx.10) → too vertical to solve against, skip (`$020397DC`);
   * solve the face plane for the height offset at (x,z): `h = −((dx·nx + dz·nz) / ny)` — the products truncate `>>10`, the divide runs on the **hardware divider** (`$02053258`: numerator `<<32`, quotient `+$80000 >> 20` — an fx.12 divide with round-to-nearest), the result drops 2 more bits;
   * the three **edge tests**: `dx·ex + h·ey + dz·ez` against a `±$20000` tolerance (edges A and B from above, edge C between `−$20000` and `length+$20000`) — the tolerance is why Mario doesn't fall through seams;
   * the ray origin must be on the plane's **front side** (an exact 64-bit dot with the full dy);
   * the surface must pass the **`CLPS` filter** (§4);
   * finally the height must beat the best so far, sit above the area floor, and lie **strictly below the ray origin**;
3. if the leaf yields nothing better, **steps the cell just below** (`ly = (ly & ~(cell−1)) − 1`) and re-descends from the root, until y walks off the bottom of the area.

On a hit, the best height (`<<6`, back to fx20.12) lands in the query's `+$44`, the hit byte at `+$48`, and the surface record — prism index, `CLPS` entry, face normal — is copied into the query via the global staging buffer at `$020A0CEC`. The manager resets `+$44` to the `−∞` sentinel `$80000000` before every cast (`$02037464`).

## 4. The `CLPS` surface-attribute table

The `.kcl` deliberately knows nothing about *behavior*. A prism carries only a u16 attribute; the level's settings block points (at `+$00`) at a **`CLPS`** chunk — magic, u16 entry size (checked against 8, `$020381CC`), u16 count, then 8-byte entries — and the attribute indexes it. A missing or malformed block falls back to a default entry `{$FC0, $FF}`.

Queries carry an opt-in **flag byte** at `+$04` (the base query constructor sets it to 1) and the filter at `$02039488` matches it against the entry's first word: property bit 5 demands query flag 2, bit 25 demands flag 8, bit 26 demands flag 4, bit 24 *excludes* the surface when flag 4 is set, behavior type `$11` (bits 19–23) is excluded by flag `$20` — and a surface with *no* special properties needs flag 1. So the same mesh answers differently shaped questions: the signpost's plain ground ray (flags = 1) sees only ordinary floors, while water surfaces, death planes and the like need their specific opt-ins — which is exactly what the verification first stumbled over (§5): the aquarium's water plane out-ranked the tank floor until the filter was replicated.

## 5. Verification — the game plays referee

`sm64ds/kcl.go` reimplements the parser and the down-ray walker in Go, to the bit: the fx.6/fx.10 truncations, the divider's rounding, the 32-bit wrapping edge dots, the 64-bit side test, the cell-stepping walk order, the filter. `cmd/kcltrace -verify N` then plays referee: it boots the oracle, has the *game* build the collision world (loader → fixup → `dBgW_Kc` constructor → registration), casts N random rays through the game's own `$02038F44`, and compares every answer — hit flag, ground height, *and the winning prism index* — against the reimplementation:

```
all 52 levels × 300 rays: 15,600 rays, 0 mismatches
```

Three real bugs fell out of the loop before it went clean, each a lesson the read watch caught: the child-index packing (y and z swapped — coherent-looking descent, wrong leaves), the reset sentinel (the manager, not the constructor, arms `+$44`), and the `CLPS` filter (water is not ground).

## 6. Same format in Mario Kart DS?

The name `.kcl` also appears in [[Mario Kart DS]]'s course archives (`course_collision.kcl`), and the two decode against each other cleanly: the same four ascending section offsets, the same thickness/minimum/mask/shift header fields in the same order, the same 12-byte vertices, 16-byte prisms and index structure. Two revisions show: Mario Kart DS's header is `$3C` bytes (one extra trailing word) and stores its minimum and thickness in plain fx20.12 where Super Mario 64 DS pre-shifts into world>>6 — and its ARM9 does *not* contain Super Mario 64 DS's four-add fixup sequence, so the loading convention differs in detail.

Is it a NitroSDK format, then? The evidence in the images says **no — it is a Nintendo EAD engine format**, carried between that studio's games rather than shipped in the SDK: it wears no NitroSDK container magic (the SDK's formats — `NARC`, `SDAT`, `BMD0` — all tag themselves), and the walkers are not in a library section but woven into the game's own `dBgW` class framework, RTTI names and all, in the same `d`-prefix hierarchy as the actors (`daObjMc_…`). Consistent with that: Super Mario 64 DS pairs its KCL with a *bespoke* model container (Part IV), while Mario Kart DS pairs its KCL with genuine SDK `NSBMD` models — the collision format is the constant, the SDK usage the variable, which is the signature of studio middleware, not of the platform kit.

## 7. The collision layer in the viewer

The Studio viewer gets a **Collision** toggle (like the 2D games' overlays): the level's collision world in red, shaded by face normal so slopes read, over — usually *instead of*, since collision covers every walkable surface — the render mesh. What doesn't turn red is telling: bridge rails, flags and window glass have no prisms at all.

**Triangles from prisms.** A prism stores no corner points, but the walker's own tests define the triangle exactly — it is the region of the face plane cut by the three edge half-planes (`eA·d ≤ 0`, `eB·d ≤ 0`, `eC·d ≤ length`). The corners are therefore the pairwise plane intersections: the anchor vertex itself, plus two solutions of `{fn·d = 0, e·d = 0, eC·d = length}`, by Cramer's rule `d = length·(fn × e) / (eC·(fn × e))` (`sm64ds/kcl.go Corners`). `cmd/exportkcl` converts every level map and every actor-bound collider this way and validates itself against the walker: **every reconstructed centroid must pass the walker's own inequalities — 0 failures across all 225 meshes** (a handful of genuinely degenerate slivers with parallel edge planes are skipped per file).

**Object colliders come from the oracle, transforms included.** Which `.kcl` belongs to which placement is not name-matching: the actor oracle records every `.kcl` an actor's own code loads (147 actors, 177 files), and reads back the collider it registers in the 24-slot table — platform actors register on their **first step**, not in init, so the oracle runs one step frame when a `.kcl` was loaded but no collider appeared. The registered object is one of the `dBgW_Kc` subclasses, and the moving ones carry their own transform: a **local→world `MtxFx43` at `+$134`** (inverse at `+$168` — the Mbg wrappers at `$02039CB8`/`$02039E48` transform the query into collider-local space before calling the same ITCM walkers) and, for `dBgW_KcMbgSclY`, an extra **fx12 local-Y scale at `+$1C8`**. Reading that matrix explained the sizes: object `.kcl` are authored ~10× large and the actors scale them down by `$199/$1000 ≈ 0.1` (fixed-point precision headroom). The oracle spawns at the origin with no yaw, so the captured matrix is the actor's *own* collider transform; the viewer composes the placement pose on top — the gate shutters land exactly inside their arches.

# Part VII — The music: `SDAT`, with its names intact

## 1. The same archive as Mario Kart DS — plus the symbol block

All sound lives in one file, `data/sound_data.sdat` (4.4 MB) — the NitroSDK sound archive, byte-for-byte the same container format decoded for [[Mario Kart DS]] (Part IV §8 there): an `INFO` block binding sequences to instrument banks and banks to wave archives, a `FAT` of member files, and the `FILE` payload. The parser, the `SSEQ` sequencer, the `SBNK`/`SWAR` instrument and wave decoding and the synth in `tools/platform/nds/sdat` apply **unchanged** — the whole Part is a reuse dividend.

One thing is different, and it is a gift: where Mario Kart DS stripped the optional **`SYMB`** name block from its retail archive (its 76 sequences are known only by number), Super Mario 64 DS **ships with `SYMB` intact**. The block mirrors `INFO`'s structure — eight sub-lists (SEQ, SEQARC, BANK, WAVEARC, PLAYER, GROUP, PLAYER2, STRM), each a `u32` count and `count` `u32` offsets — but its offsets (relative to the `SYMB` block) point at NUL-terminated symbol names. `tools/platform/nds/sdat` now reads it when present, and every sequence introduces itself: `NCS_BGM_TITLE`, `NCS_BGM_SHIRO` ("castle"), `NCS_BGM_CHIJOU` ("overground"), `NCS_BGM_KUPPA` (Bowser), `NCS_BGM_STAFFROLL`. The `NCS_` prefix runs through the whole archive (`NCS_BANK_SE_ACTION`, `NCS_WAVE_RESIDENT`…) — the sound project's own namespace, on the cartridge.

The census: **282 member files — 79 `SSEQ` sequences, 109 `SBNK` banks, 89 `SWAR` wave archives, 5 `SSAR` effect archives** — bound by 83 SEQ records, 134 BANK records and 89 WAVEARC records (records can share files: `PANEL`/`MINIMOTOS`/`MOTOS` are one `SSEQ` played through different player setups, as are `MINIVOCAL`/`VOCAL` and `MINIPUKKUN`/`PUKKUN` — the minigame versions of world themes are literally the same sequence). As on Mario Kart DS there is **no `STRM`** streamed audio: every note in the game is sequenced.

## 2. Rendered — all 83 sequences

`extract/cmd/musicrender` plays every sequence through the driver-faithful sequencer — 48 ticks per quarter, envelopes at 192 Hz in dB×1024 units, the DS's 32768 Hz mixer — for two loops (capped at 3 minutes) with a fade, and encodes MP3s named after the `SYMB` symbols. **All 83 render**, from the 3-second star-drop stinger to the 3-minute staff roll. The names sort the soundtrack at a glance: the world themes (`SHIRO` castle 62 s, `CHIJOU` overground 73 s, `WATER` 91 s, `SNOW2` 92 s, `OBAKE` haunted house 105 s, `DUNGEON` 124 s, `ATHRETIC` [sic] 78 s, `KUPPA`/`KUPPA3` Bowser 73/84 s, `DOLPIC` beach 90 s and its sung twin `VOCAL`), the power-up loops (`MUTEKI` invincible, `METAL`, `FIRE`, `BIG`, `BALLOON`, `OWLFLY`), the `MINI*` minigame set, the `VS*` multiplayer set, and a fanfare for everything from `FIRSTCAP` to `GAMEOVER`. The tracks land in the Studio's music panel under their cartridge names.

Honest caveat, the same as Mario Kart DS: the envelope tables and PSG behavior are the sound driver's *documented* semantics reimplemented, not a bit-exact trace of this cartridge's ARM7 driver. The level→sequence binding — which Mario Kart DS never yielded — fell out of the course-name hunt (Part VIII): the per-level info table at `$02075768` (three interleaved `s8` arrays, 3 bytes per level) carries the level's BGM sequence in its **third byte**, read at level start by `$0202D35C` (`LDRSB [$0207576A + level*3]`) and handed to the music starter `$0201320C`, where a negative value (the table's `$FF`) stops the music instead — the castle exterior and the cap-course interiors are deliberately silent. The values confirm themselves: the castle interior plays 57 `SHIRO` ("castle"), Bob-omb Battlefield and Whomp's Fortress share 58 `CHIJOU` ("overground"), Jolly Roger Bay, Dire Dire Docks *and the Secret Aquarium* share 59 `WATER`, the three Bowser roads share 64, their three boss arenas play 66/66/67 `KUPPA`/`KUPPA3`, and Sunshine Isles plays 68 `DOLPIC` — the beach tune named after Super Mario Sunshine's Delfino. `sm64ds.LevelSet.MusicSeq` exposes the table.

# Part VIII — The message system and the game's own course names

This Part exists because of a correction. The Studio's level list had shipped with **hand-written course names, guessed from the Japanese internal stems** — and several were wrong (`kaizoku_irie`, "pirate inlet", had been labelled Dire, Dire Docks when it is Jolly Roger Bay; the mountain and its slide were swapped; `suisou`, "aquarium", was labelled Wet-Dry World). The guesses are gone; everything below is read out of the cartridge and pinned by the game's own code.

## 1. The `BMG` text containers

All text lives in **BMG** containers: `data/message/msg_data_<lang>.bin` (eng/frn/gmn/itl/spn — the 711 in-game messages, `LZ77`-tagged like the `.bmd` models) plus five small BMGs embedded in the ARM9 itself (the pre-boot option menus, one per language, at file offset `$8E89C`ff). The magic is `MESG`/`bmg1` stored byte-swapped (`GSEM1gmb`), then tagged sections. The game's parser — overlay 7, `$020C951C` — walks the sections comparing each tag word against `INF1`/`FLW1`/`DAT1`/`STR1`/`FLI1` constants and stores the `INF1` base, its entry array (`INF1+$10`) and the string pool (`DAT1+8`) in globals `$02104C1C`/`$02104C18`/`$02104C24`. The string getter `$020C94A0` then computes:

```
string(id) = DAT1+8 + u32( entries[ id * (entrySizeField >> 3) ] )
```

with the `INF1` header holding the count (`u16` at `+8`) and that entry-size field (`u16` at `+$0A`, value `$40`, i.e. 8-byte entries: `u32` string offset + `u32` attributes). Reimplemented in `sm64ds/msg.go` (`LoadBMG`/`ParseBMG`).

## 2. The encoding: bytes are font-glyph indices

The strings are not ASCII: a message byte is an **index into the dialog font**. The font lives in `ARCHIVE/c2d.narc` member 13 — 16×16-pixel 4bpp glyphs, two tiles wide, in a 32-tile-per-row sheet — and *reading the sheet in index order is the decoder*: digits `0`–`9` at index 0, `A`–`Z` at `$0A`, `a`–`z` at `$2D`, punctuation between and after, and a blank cell at `$4D` — the space. `$FD` is a newline, `$FF` terminates a message, and `$FE` opens a control sequence whose next byte is the control's total length (the button and d-pad icons in the slide instructions). The mapping cross-checks against the ARM9-embedded menus — `CONTINUE`, `EXIT COURSE`, `DUAL-HAND MODE`, and in the French bank `CONTINUER`/`QUITTER NIVEAU` — and the first message of the English file is Peach's letter, exactly as the game opens: *"Dear Mario: Please come to the castle. I've baked a cake for you."* This also unblocks the signpost/dialog frontier from Part V — all 711 messages decode.

## 3. The course names, and the table that binds them

Messages **406–435** are the course names, in course-index order: `" 1 BOB-OMB BATTLEFIELD"` through `"15 RAINBOW RIDE"` (the numbered painting courses), then the unnumbered `BOWSER IN THE DARK WORLD`/`FIRE SEA`/`SKY`, the DS boss courses (`GOOMBOSS BATTLE`, `BIG BOO BATTLE`, `CHIEF CHILLY CHALLENGE`), the secret courses (`THE PRINCESS'S SECRET SLIDE`, `THE SECRET AQUARIUM`, `? SWITCH`, `THE SECRET UNDER THE MOAT`, `BEHIND THE WATERFALL`, `OVER THE RAINBOWS`, `SUNSHINE ISLES`, `THE SECRET OF BATTLE FORT`), and `CASTLE SECRET STARS`.

The binding from level to course is one table and one accessor. The **`s8` table at `$02075298`** maps each of the 52 levels to its course index; the accessor is the one-liner at **`$02013558`** (`LDRSB r0,[$02075298, r0]`), called from 35 sites across the save/star bookkeeping (star counts loop courses 0–14 for the paintings and up to 29 overall, matching the indexing). The castle hub levels, the playroom and Rec-Room share row 29 — `CASTLE SECRET STARS`, exactly the 16th entry of the pause screen's course-title table at `$020757D0`, whose contents are precisely `{406..420, 435}`. The course names for courses that own several maps confirm the stems' plain meanings: `snow_mt`+`snow_slider` are both Cool, Cool Mountain (the mountain and the slide *inside* it), `tibi_deka_*` ("tiny/huge") is Tiny-Huge Island, `water_city` is Wet-Dry World, `water_land` is Dire, Dire Docks, `horisoko` ("moat bottom") is The Secret Under the Moat, and the three `ex_*` cap courses are the DS boss battles.

`cmd/exportbmd` now derives the Studio's level names from exactly this join (`buildLevelNames`): course name from message `406+course`, title-cased (a display choice — the game shows the names in capitals; "Bob-omb"'s casing follows the game's own mixed-case dialog), with the internal stem appended to a course's secondary maps. The castle hub keeps literal stem descriptions — it is not a course, and its outdoor label `CASTLE GROUNDS` appears on the cartridge only as a pre-rendered banner image in the per-language menu archives (`ARCHIVE/cee.narc` for English), the same sheets that carry `SUNSHINE ISLES` and `PRINCESS'S SECRET SLIDE`.

One more prize fell out of the same 3-bytes-per-level info table at `$02075768`: its third byte is the level's **background-music sequence** (Part VII §2), read at level start by `$0202D35C` and started by `$0201320C` — negative stops the music, so the `$FF` rows are the deliberately silent levels.

## 4. Which signpost says what

Every readable signpost in the game is the same actor, 184 (`obj_tatefuda`, Part V §6) — so where does each one get its words? From its **placement**: the first object parameter (`par1`) is an **external message ID**. These IDs are not `INF1` indices — they live in their own namespace (the course-story signs count up from `1000 + 50·course`: Bob-omb Battlefield uses 1000–1008, Whomp's Fortress 1050–1054…; shared tutorial signs sit in an 1800 block and recur across levels). The translation is the function the message window runs every ID through: **`$020B8EC0`** walks a `{u16 firstID, u16 firstIndex}` **range table at ARM9 `$0208EEEC`** (half-open ranges, sentinel ID ≥ 8000) and returns `firstIndex + (id − firstID)`.

The join proves itself instantly: ID 1000 → index 42 = *"BEWARE OF CHAIN-CHOMP / Extreme Danger!"* — the famous sign next to the Chain Chomp's stake in Bob-omb Battlefield — and 1003 → 45, the Big Bob-omb's *"No visitors allowed, by decree of the Big Bob-omb."* Reimplemented as `sm64ds.LevelSet.MsgIndex`; `cmd/exportlevelobjs` now attaches each signpost's English text to its placement JSON (`txt`), and clicking a signpost in the Studio viewer quotes the sign itself — 102 of the 103 placed signposts resolve (the one holdout is on the unused test map, placed with parameter `$FFFF`, no valid message). Icon escapes (`$FE` controls — the button and d-pad glyphs in control hints) are skipped in the decoded text.

# Appendix A — Toolchain and reproduction

Everything here is derived from the `.nds` image with this repository's own tools: the shared `tools/platform/nds` container reader and the `tools/cpu/arm` disassembler/tracer, plus this game's `extract` module. No third-party emulator, debugger or disassembler was used, and nothing was read from released source.

```sh
# identity (size + MD5 pinned in ../README.md#image-files)
md5 "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"

# Part I — header, integrity checks, overlay/filesystem summary (and -files / -tree)
go run retroreverse.com/tools/platform/nds/cmd/ndsinfo "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
go run retroreverse.com/tools/platform/nds/cmd/ndsinfo -tree  "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"
go run retroreverse.com/tools/platform/nds/cmd/ndsinfo -files "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds"

# Part II — extract + BLZ-decompress the ARM9/ARM7 binaries and all 103 overlays into extracted/
( cd extract && go run ./cmd/ndsextract "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )
# add -fs to also dump every filesystem file under extracted/files/

# trace the ARM9 boot chain from its entry (ARM state) over the decompressed image
go run retroreverse.com/tools/cmd/codetracearm -base 0x02004000 -entry 0x02004800 extracted/arm9_dec.bin

# trace the ARM7 boot chain from its entry
go run retroreverse.com/tools/cmd/codetracearm -base 0x02380000 -entry 0x02380000 extracted/arm7.bin

# Part III — run the ARM9 boot on the tools/cpu/arm core as an oracle: BLZ cross-check,
# the I/O registers it programs, and the ARM9↔ARM7 IPCSYNC rendezvous it stops at
( cd extract && go run ./cmd/bootoracle -io "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )

# Part IV — decode every .bmd model and export GLBs (+ the viewer manifest)
go run ./cmd/exportbmd -all

# Part V — run every placed actor's create/init on the tools/cpu/arm core and record
# the files its code loads (the actor oracle; writes extracted/actorbind.json)
go run ./cmd/actororacle
# then decode all 52 levels' object placements into per-stage JSON, bound
# through the oracle table
go run ./cmd/exportlevelobjs

# Part III §6 — the dual-core oracle: both CPUs on shared RAM + IPC, clearing the
# rendezvous the single core stops at (into the post-sync PXI FIFO exchange)
( cd extract && go run ./cmd/dualoracle "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds" )

# Part VI — build the collision world with the game's own code, cast the signpost's
# ground ray under a read watch (structure log + the PCs that touched the file)
go run ./cmd/kcltrace -level 1
# oracle-verify the Go reimplementation of the .kcl walker: N random rays per level,
# game vs. sm64ds/kcl.go, comparing hit/height/prism (0 mismatches, all 52 levels)
go run ./cmd/kcltrace -level 30 -verify 300
# dump the ITCM image (the collision walkers live there) for disassembly
go run ./cmd/kcltrace -itcm ../extracted/itcm.bin

# Part VI §7 — reconstruct every collision mesh as a red viewer GLB (level maps +
# the 177 actor-bound object colliders), self-validated against the walker's tests
go run ./cmd/exportkcl
# the actor sweep also records .kcl loads and registered collider transforms
# (actorbind.json "kcl"/"colliders"), consumed by exportlevelobjs

# Part VII — render all 83 SSEQ sequences through the SDAT sequencer+synth to MP3,
# named from the archive's own SYMB symbols (needs ffmpeg)
go run ./cmd/musicrender                        # → work/music/, copied to site/public/sm64ds/music/
```

Toolchain (shared `retroreverse.com/tools`, this repository):

- **`tools/platform/nds`** — the Nintendo DS Game Card container reader: header parse + CRC-16 verification, the FAT, the FNT directory-tree walk, the ARM9 overlay table, and the **BLZ** (backward-LZSS) decompressor the SDK applies to the ARM9 static module and every overlay. Shared with the [[Mario Kart DS]] analysis; makes no assumptions about the game inside.
- **`tools/platform/nds/dsmachine`** — the reusable **dual-core DS machine**: two `tools/cpu/arm` cores over one shared main RAM + WRAM, per-core private TCM/WRAM/BIOS, the cross-wired IPCSYNC mailbox and the two IPC FIFOs, a per-core interrupt controller, and the BIOS `SWI`s. Game-neutral; the model any DS title's dual-core oracle builds on. (`arm.CPU.Exception`, added for the BIOS-style IRQ dispatch, is its one addition to the core.)
- **`tools/platform/nds/cmd/ndsinfo`** — the container inspector for Part I (`-files`, `-tree`, `-grep`).
- **`tools/cpu/arm`** — the ARM9/ARM7 disassembler and CPU core (ARMv5TE + ARMv4T; ARM + Thumb, interworking-aware), with `tools/cmd/disarm` (linear) and `tools/cmd/codetracearm` (recursive-descent, follows ARM↔Thumb interworking) as CLIs.
- **`supermario64ds/extract/cmd/ndsextract`** — this game's extractor: writes `arm9.bin`/`arm7.bin` and the 103 overlays, and their BLZ-decompressed forms (`arm9_dec.bin`, `ovl9_NNN_dec.bin`), into `extracted/` (regenerable, git-ignored).
- **`supermario64ds/extract/cmd/bootoracle`** — runs the ARM9 boot on the `tools/cpu/arm` core over a flat DS memory (with the BIOS `SWI`s the startup needs): cross-checks BLZ against the game's own decompressor, logs the I/O registers programmed, and stops at the ARM9↔ARM7 `IPCSYNC` rendezvous.
- **`supermario64ds/extract/cmd/dualoracle`** — co-executes both boot binaries on the `tools/platform/nds/dsmachine` dual-core model, clearing the rendezvous the single-core oracle stops at and running the ARM9 into the post-sync PXI exchange (Part III §6).
- **`supermario64ds/extract/cmd/kcltrace`** — the Part VI instrument: has the game itself load, fix up and register a level's `.kcl` collider in the oracle, casts ground rays through the game's own dispatcher with a **read watch** over the served file (every byte the walker touches, with the touching PC), and cross-checks `sm64ds/kcl.go`'s bit-exact reimplementation against the running original (`-verify`).
- **`supermario64ds/extract/cmd/exportkcl`** — Part VI §7: reconstructs every prism's triangle from the walker's own half-plane definition and writes the red collision GLBs for the Studio's Collision toggle (level maps + the object colliders the binding oracle saw actors load), self-validated centroid-by-centroid against the walker's inequalities.
- **`tools/platform/nds/sdat`** — the SDAT sound archive: container/INFO/FAT parse (now also the optional `SYMB` name block, which this game ships), SBNK instruments, SWAR/SWAV waves (PCM8/16, IMA-ADPCM), and the SSEQ sequencer + synth (driver-faithful timing and envelopes) rendering to stereo PCM. Built for [[Mario Kart DS]]; reused here unchanged apart from the `SYMB` reader.
- **`supermario64ds/extract/cmd/musicrender`** — Part VII: renders every sequence to MP3 (via ffmpeg), named from the `SYMB` symbols, and writes the Studio music panel's `tracks.json`.

Rendered figures will go in `Super Mario 64 DS (DS)/work/`; annotated disassembly in `disasm/`.

---

# Part IX — The machine: booting the game and drawing its title

Parts I–VIII read the cartridge. This part *runs* it.

The DS model up to now was two ARM cores, the inter-processor block, and stubs for
everything else. It could clear the ARM9↔ARM7 rendezvous (Part III §4) and no further:
`main()` was never reached. The distance from there to a title screen is not a bug list,
it is a console — the display and its timing, eight DMA channels, eight timers, the
ARM9's hardware divider and square-root units, the nine VRAM banks and the register that
decides what each of them currently *is*, the cartridge port, the ARM7's SPI bus, and both
graphics engines. That machine is `tools/platform/nds/dsmachine`, and it is a low-level
emulation like the N64's and the PSX's: a DS has no operating system to high-level-emulate,
only hardware. The one thing lifted above the metal is the BIOS's software interrupts —
they are a library, not a kernel.

It works. Super Mario 64 DS boots from cold, brings up its OS, reads **2,731 files** off
the cartridge, and renders its title sequence: the ©2004-2005 Nintendo legal screen, the
TOUCH TO START star, and — after a scripted stylus tap — the SUPER MARIO 64 DS logo with
Mario's 3D face and the VS / ADVENTURE / REC ROOM menu.

    bootoracle -image rom.nds -frames 1100 -keys start.keys -shot title

## 1. What each bug looked like

The interesting thing about the eight bugs that stood between the rendezvous and the title
screen is that not one of them *looked like what it was*. They are worth recording in that
form, because the shape of the symptom is the part that transfers.

**The boot hangs, and the ARM7 is waiting for a value nobody writes.** The ARM7 polls a
halfword at `$027FFFF4` for `$7F` and reads zero for ever. That address is past the end of
main RAM — and main RAM is **mirrored** every 4 MiB. The DS's boot parameter block lives in
the top mirror (`$027FFxxx`), and NitroSDK hard-codes it there. Map only the first 4 MiB and
the ARM9 writes the handshake to an address the ARM7 cannot see.

**The ARM7 swallows four IPC messages to receive one.** The CPU core composes a 32-bit load
out of four byte reads, and both the IPC receive FIFO and the cartridge data port are
*read to pop*. So every word read popped four, and every word the game read off the card was
assembled from the bytes of four different words. This is why `arm.BusWide` exists: a machine
whose registers have side effects on read must service wide accesses itself. It is not a
performance interface.

**The ARM7 executes its own BIOS thunk table as garbage.** The game's interrupt handler ends
with `ldr pc, [sp], #4` — a plain load, restoring nothing. It is therefore not the interrupt
*vector*; it is a routine the BIOS *calls*, and the BIOS's own epilogue (`subs pc, lr, #4`)
is what restores the CPSR, and with it the **Thumb bit**. Jump straight to the game's handler
and the ARM7 — interrupted out of Thumb code, as it usually is, because the BIOS call thunks
are Thumb — comes back in ARM state on a Thumb address and decodes the thunk table as ARM.

**The ARM7's thread scheduler spins between two threads without executing an instruction of
either.** `LDMIA r0, {r0-lr}^`. The `^`, with no PC in the list, means *transfer the user-mode
bank* — which is exactly how an OS restores a thread's context from Supervisor mode without
touching the Supervisor LR it is about to return through. Treat the `^` as a no-op and the load
clobbers that LR, and the context switch returns into its own middle.

**Every ARM9 interrupt is latched and none is dispatched.** The guard was "is a handler
installed?", implemented as "does the pointer look like a code address?" — rejecting anything
below main RAM. SM64DS's handler is at `$01FFD97C`: in **ITCM**, which is below main RAM and is
precisely where a DS game wants its interrupt handler, because ITCM is the only memory the ARM9
reaches in a single cycle. The sanity check was the bug.

**The filesystem asserts, and every byte it read was correct.** The card's block-size field is
`0x100 << n`, not `0x100 << (n-1)`. SM64DS's FS takes the name table's ROM offset (`$1D9324`),
rounds it down to a 512-byte boundary (`$1D9200`), reads that block, and indexes `$124` into what
comes back. With the off-by-one the read returns 256 bytes — all of them right — and simply stops
short of the offset the game is about to read from. It resolves its first path against padding and
hangs in `myFS_ConvertPathToFileID`.

**The ARM7 puts the console to sleep, correctly.** `EXTKEYIN` bit 7 is **0 for lid OPEN**. Every
other bit in that register reads 1 for "not pressed", so setting the lid bit the same way is the
natural mistake — and the ARM7 then does exactly the right thing with a console it believes has
been folded shut.

**The 3D engine draws, swaps, and is missing most of its own command stream.** The GXFIFO is not
one address: it is the whole 64-byte window `$04000400`–`$0400043F`, because a game feeds the
geometry engine with burst stores and an `STMIA` (or a DMA) walks the destination address up as it
goes. Match only the first address and you accept the first word of every burst and drop the rest.
What survives still decodes cleanly, still builds polygons, still swaps buffers — and never once
multiplies a matrix. The camera never arrives, and 95% of the geometry lands behind the eye and is
clipped away. The tell was in the `-gxdump` histogram: **a 3D game that never issues `MTX_MULT` is
not a 3D game.**

**The stylus works, reports a position, and the title screen ignores it.** The ARM7 asks for a
VCount-match interrupt at line 197 to sample the touchscreen. `DISPSTAT` splits that 9-bit line
number across the register — bits 15..8 hold the low eight, and bit **7**, nowhere near them, holds
the ninth. Reassemble it as `(d >> 7) & 0x100` and it lands on bit 15, so any line ≥ 128 quietly
acquires an extra 256: the ARM7 asks for line 197 and gets 453, a line the display never reaches.

## 2. The instruments the chase produced

Each of these exists because a run lied and something had to make it stop.

* **`-log`** — the honest half of every run: the hardware the model does *not* implement. A DS boot
  polls status bits constantly, so a stub that happens to read "ready" is indistinguishable from
  working silicon right up until the frame it isn't. The model logs; it does not fake.
* **`-rtshot`** — the 3D engine's render target on its own, before the 2D engine composites it.
  Untouched pixels come out magenta. A black screen is two entirely different bugs wearing one face
  (the GPU drew nothing / the GPU drew and the compositor threw it away) and no counter separates them.
* **`-gfx`** — polygons at the last swap, and how many primitives the clipper rejected. *95% clipped*
  and *0 polygons* are not the same bug and are not investigated the same way.
* **`-gxdump`** — the geometry commands actually executed. See above.
* **`-cardlog`** — every cartridge transfer, drawn by the game's own loader.
* **`-keys FILE`** — a *timed* input script. The point is the press **edge**: a DS game asks "did this
  go down since last frame", so a stylus held from reset gives the title screen nothing, and it waits
  for ever with TOUCH TO START on the screen.
* **`-savestate` / `-loadstate`** — a cold boot to the title is 1.2 billion scheduler steps and 42
  seconds; restoring it takes one. Every question asked of the title screen used to cost 42 seconds
  before it could be asked.

## 3. What is not yet right

Stated plainly, because a screenshot that looks correct is not an argument:

* **Toon-shaded polygons render pink.** A third of the title screen's polygons use polygon mode 2.
  The toon table in the registers is a clean grey ramp and Mario's skin palette in VRAM is a correct
  peach ramp — both verified — so the data is right and the *shading* is wrong. Open.
* **No sound.** The register file is real and the ARM7's sound driver runs and sequences its music;
  no samples are fetched. (The music itself is already rendered, by a different route: Part VII.)
* **Display capture, anti-aliasing, edge marking, fog** — declared, logged, not implemented.
* The 2D engines compose a whole frame from one register snapshot at line 0, so a mid-frame raster
  split would not appear. Nothing in this title needs one.
