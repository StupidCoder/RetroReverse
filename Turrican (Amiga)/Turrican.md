# Turrican (Amiga) — disk format and game analysis

A reverse-engineering reference for `Turrican.adf`, the Amiga release of
Turrican (Rainbow Arts / Factor 5, 1990). This is the second Amiga title in this
repository and the writeup follows the same shape as the others, in reading
order:

* **Part I** — the disk image: the ADF container and the disk's *custom* layout.
  Unlike Marble Madness, this is **not** an AmigaDOS volume — it is a bootable
  non-DOS disk whose boot block is a hand-written sector loader, so Part I is
  about mapping the raw disk rather than walking a filesystem;
* **Part II** — the boot chain: the boot block's multi-stage load and the
  unpacking / decryption of the main program from the packed track data;
* **Part III** — the game program: the 68000 startup, the interrupt/copper
  setup and the memory map;
* **Part IV** — graphics and data formats: the tile, sprite, level and audio
  encodings;
* **Part V** — game mechanics: the player, weapons, enemies, the levels and
  progression.
* **Appendices** — toolchain and reproduction.

Methods: purely static analysis of the disk image, plus the 68000 toolchain in
the shared `tools/` module — the AmigaDOS reader (`tools/amiga/adf`), the
disassemblers (`tools/cmd/dis68k`, `tools/cmd/codetrace68k`) and the 68000
execution core (`tools/m68k`) for dynamic verification. All addresses are 68000
addresses; sizes are `.b`/`.w`/`.l` (8/16/32-bit). **Part I is complete; Parts
II–V are stubs.**

---

## Contents

- [Part I — The disk image](#part-i--the-disk-image)
  - [1. The ADF container](#1-the-adf-container)
  - [2. A custom boot disk, not an AmigaDOS volume](#2-a-custom-boot-disk-not-an-amigados-volume)
  - [3. The boot block: a raw-sector loader](#3-the-boot-block-a-raw-sector-loader)
  - [4. The disk map](#4-the-disk-map)
- [Part II — Boot chain](#part-ii--boot-chain)
- [Part III — Game program architecture](#part-iii--game-program-architecture)
- [Part IV — Graphics and data formats](#part-iv--graphics-and-data-formats)
- [Part V — Game mechanics](#part-v--game-mechanics)
- [Appendix A — Toolchain and reproduction](#appendix-a--toolchain-and-reproduction)

---

# Part I — The disk image

## 1. The ADF container

An ADF is the simplest possible disk image: a flat dump of the floppy's logical
blocks with no header or metadata. Turrican ships on one standard
double-density disk — **1760 blocks of 512 bytes = 901,120 bytes** — so block
*N* is simply the 512 bytes at file offset *N* × 512. The exact copy this
analysis is based on is pinned by size and MD5 in the repository
[README](../README.md#image-files).

## 2. A custom boot disk, not an AmigaDOS volume

The first four bytes are `44 4F 53 00` — the `"DOS\0"` boot-block signature — so
the Kickstart ROM will accept the disk and run its boot code. But that is as far
as AmigaDOS goes: there is **no filesystem on the disk**. The boot block's
block-8 field still carries the conventional root-block pointer (`$00000370` =
880, the standard value for a DD disk), yet block 880 is not a valid root block,
and the AmigaDOS reader rejects it:

```
$ adfdump Turrican.adf
adfdump: adf: root block is not a valid root header
```

This is the usual shape of a commercial Amiga game disk: the `"DOS\0"` signature
and a valid boot-block checksum are the *only* AmigaDOS-conformant things on it.
Everything else — the program, the graphics, the levels — is laid out in a
private format and pulled off the disk by the game's own loader, addressing the
medium by absolute byte offset through `trackdisk.device`, never through files.
(Contrast Marble Madness, whose disk is a real OFS volume — see that writeup's
Part I.) So this Part maps the raw disk; decoding what the loader fetches is
Part II.

## 3. The boot block: a raw-sector loader

The boot block is blocks 0–1 (1024 bytes): the `"DOS\0"` tag, a checksum
(`$090B08A1` at `+4`, which the ROM verifies before it will boot), the vestigial
root pointer at `+8`, and from `+12` the boot code the ROM jumps to with the
boot device's I/O request in `A1`. That code is a complete sector loader.

It begins with a first read (`BSR $2C0`), filling low memory and running a
first-stage routine off the end of it:

```
$2C2  MOVE.w #$2,$1C(a1)        ; io_Command = CMD_READ
$2C8  MOVE.l #$30000,$28(a1)    ; io_Length  = 192 KB
$2D0  MOVE.l #$1000,$24(a1)     ; io_Data    = $1000
$2D8  MOVE.l #$400,$2C(a1)      ; io_Offset  = $400  (block 2)
$2E0  JSR  -$1C8(a6)            ; DoIO  (a6 = ExecBase, -456 = DoIO)
$2E4  JSR  $30000               ; run the loaded first stage
$2EA  MOVE.w #$1A0,$DFF096      ; DMACON: clear bitplane/copper/sprite DMA
```

So blocks 2–385 (`$400`…`$30400`) are read to `$1000`, and the routine that ends
up at `$30000` is called — the disk's first-stage loader/unpacker (it lives at
the *end* of the packed blob it travels with, the classic "decompressor stub
after the data" arrangement). Control then returns to the main boot code, which:

1. blanks the border (`CLR.w $DFF180`) and takes `A6 = ExecBase` (`$4.w`);
2. sizes and grabs a work buffer — `AvailMem`/`AllocMem` (`-$D8`/`-$C6` on
   `ExecBase`) for the **largest FAST chunk** (`MEMF_FAST|MEMF_LARGEST`,
   `$20004`), or, on a 512 KB chip-only machine, the chip region `$80000`…
   `$100000`;
3. issues the **main read** — `io_Offset $2C00` (block 22), `io_Length $50000`
   (320 KB), `io_Data $22E00` — pulling the main program into chip RAM, then
   stops the drive motor (`io_Command = 9`, `TD_MOTOR`);
4. adapts to the CPU: on a 68010 or better (`BTST #1,AttnFlags`) it installs a
   `TRAP #0` handler that executes a `MOVEC` — the standard fixup so the rest of
   the loader can keep treating the machine like a bare 68000;
5. seizes the machine: `MOVE #$2000,SR` (supervisor, no interrupts), stack at
   `$80000`, copies a 512-byte tail routine to **`$7F800`** and `JMP $7F800` to
   carry on the load/decrypt.

The boot block never touches `dos.library`; it reads the disk by absolute byte
offset and drives the hardware directly. Following the `$7F800` stage and the
unpacking is Part II.

## 4. The disk map

Reading the boot block's offsets back onto the image, and confirming it with a
byte-entropy sweep, the disk falls into three regions:

| blocks | offset | entropy | contents |
|--------|--------|---------|----------|
| 0–1 | `$0`–`$400` | ~3.7 | **boot block** — the sector loader (§3) |
| 2–21 | `$400`–`$2C00` | ~3.7 | **first-stage loader code** — plain 68000 (it opens `MOVEM.l d0-d7/a0-a6,-(a7)` then drives `DMACON`/`$DFF096`), loaded to `$1000` |
| 22–1759 | `$2C00`–end | ~7.99 | **packed/encrypted payload** — the main program, graphics and level data, read in bulk (the `$50000` main read starts here) |

The low two regions are recognizable 68000 code (entropy well under 4 bits/byte);
from block 22 on the image is essentially incompressible (entropy ~7.99 of a
possible 8), which is the signature of packed or encrypted data, not a
filesystem or raw bitplanes. There is no directory to enumerate — the boot
block's two reads (§3) are the disk's entire "table of contents." Turning that
packed payload back into program is the work of Part II.

---

# Part II — Boot chain

> **Stub.** Trace the `$7F800` stage and the first-stage routine at `$30000`:
> the unpacking/decryption of the `$22E00` payload, any further `trackdisk`
> reads (the rest of the disk past the main read), and the hand-off to the
> game's own startup.

# Part III — Game program architecture

> **Stub.** The 68000 startup of the unpacked program: interrupt and copper
> setup, the main loop, and the memory map.

# Part IV — Graphics and data formats

> **Stub.** Tile maps, sprites/BOBs, the level encodings, fonts and audio
> (Turrican's music is a TFMX/“Chris Hülsbeck” player).

# Part V — Game mechanics

> **Stub.** Player movement and the weapon system, enemies, the worlds and
> level structure, scoring and progression.

# Appendix A — Toolchain and reproduction

The disk facts above reproduce with the shared tools:

```sh
# size + hash (must match the README image table)
md5 "Turrican (Amiga)/Turrican.adf"

# confirm it is not an AmigaDOS volume
go run stupidcoder.com/tools/amiga/cmd/adfdump "Turrican (Amiga)/Turrican.adf"

# disassemble the boot block (code starts at +12)
go run stupidcoder.com/tools/cmd/dis68k -skip 12 -base 0xc "Turrican (Amiga)/Turrican.adf"
```
