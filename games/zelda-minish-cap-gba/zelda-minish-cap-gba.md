# The Legend of Zelda: The Minish Cap (Game Boy Advance) — technical reference

**Image:** `Legend of Zelda, The - The Minish Cap (USA).gba` — 16,777,216 bytes, MD5
`a104896da0047abe8bee2a6e3f4c7290`. Not committed (copyright); supply your own copy.

A technical reference for the 2004/2005 Game Boy Advance Zelda, developed by
Capcom/Flagship for Nintendo — internally, as its own build string admits,
**"ZELDA 5"**. The console's CPU is the **ARM7TDMI** (ARMv4T, ARM + Thumb),
the same core this repository already executes as the Nintendo DS's ARM7, so
the whole `tools/cpu/arm` toolchain (`disarm`, `codetracearm`, the emulator)
applies from day one. The writeup proceeds in reading order:

* **Part I** — the cartridge image: the flat ROM dump, the AGB memory map, the
  cartridge header, the boot handshake and the crt0, and the save memory;
* **Part II** — boot and initialization: from `AgbMain` through the interrupt
  scaffolding to the first rendered frame;
* **Part III** — engine architecture: the main loop, the IRQ handler at
  `$03005D90`, task/state dispatch, and the RAM layout;
* **Part IV** — graphics and data formats: tiles, palettes, OAM, the BIOS
  compression formats (LZ77/Huffman/RLE) and the game's own asset containers;
* **Part V** — the world: maps, rooms, the Minish scale mechanic, entities;
* **Part VI** — music and sound;
* **Appendix** — toolchain and reproduction.

Methods: static analysis of the ROM image with the repository's own tools
(`tools/platform/gba`, `tools/cpu/arm`), then a GBA oracle for dynamic
verification. All bus addresses are 32-bit (`$00000000`–`$0FFFFFFF` region of
interest); a *file offset* is called out explicitly where it differs. The GBA
is little-endian throughout.

---

## Contents

- [Part I — The cartridge image](#part-i--the-cartridge-image)
  - [1. The ROM dump](#1-the-rom-dump)
  - [2. The AGB memory map](#2-the-agb-memory-map)
  - [3. The cartridge header (`$08000000`–`$080000BF`)](#3-the-cartridge-header-0800000008000obf)
  - [4. Boot: the BIOS handshake and the crt0](#4-boot-the-bios-handshake-and-the-crt0)
  - [5. The save memory](#5-the-save-memory)

---

# Part I — The cartridge image

Like the Game Boy before it, a GBA cartridge is the simplest image format in
the catalogue: **no container, no filesystem, no loader, no relocation**. The
`.gba` file is a verbatim dump of the cartridge mask ROM: byte *N* of the file
is the byte the CPU reads at bus address `$08000000 + N`. Unlike the Game Boy
there is not even a bank-switching mapper — the ARM7TDMI's 32-bit bus simply
maps the whole 16 MiB chip flat. There is nothing to extract; the structure
that matters is the memory map the console imposes, a 192-byte header Nintendo
stamps on the front, and whatever conventions the game's own build system
(Nintendo's AGB SDK) left behind.

## 1. The ROM dump

| property | value |
|---|---|
| size | 16,777,216 bytes (16 MiB — the largest standard AGB mask ROM) |
| MD5 | `a104896da0047abe8bee2a6e3f4c7290` |
| used extent | last non-padding byte at file offset `$DE7D87` — 13.91 MiB used, the rest `$FF` fill |
| internal name | `AGBZELDA:THE MINISH CAP:ZELDA 5` at file offset `$11E484` |

The `$FF` padding is the erased state of the mask ROM's flash master; the
game's content ends at `$DE7D87` and — fittingly — the very last thing in ROM
is the save driver's ID string (`EEPROM_V124` at `$DE7D34`, §5).

## 2. The AGB memory map

The fixed bus map every GBA game lives in:

| bus address | size | region | notes |
|---|---|---|---|
| `$00000000` | 16 KiB | BIOS ROM | boot + SWI library; readable only while executing in it |
| `$02000000` | 256 KiB | EWRAM | external work RAM, 16-bit bus, 2 waitstates |
| `$03000000` | 32 KiB | IWRAM | internal work RAM, 32-bit bus, zero waitstates |
| `$04000000` | ~1 KiB | I/O registers | PPU, sound, DMA, timers, keypad, IRQ control |
| `$05000000` | 1 KiB | palette RAM | 256 BG + 256 OBJ entries, 15-bit BGR |
| `$06000000` | 96 KiB | VRAM | tiles, tilemaps, bitmap modes |
| `$07000000` | 1 KiB | OAM | 128 sprite entries + affine parameters |
| `$08000000` | ≤32 MiB | cartridge ROM | waitstate configuration 0 |
| `$0A000000` | mirror | cartridge ROM | waitstate configuration 1 |
| `$0C000000` | mirror | cartridge ROM | waitstate configuration 2 |
| `$0E000000` | 64 KiB | cartridge SRAM/Flash | 8-bit bus (not used by this cart — EEPROM instead, §5) |

Two facts shape everything a GBA game does. First, **IWRAM is the only fast
memory** — 32 KiB on a zero-wait 32-bit bus, while ROM and EWRAM sit on 16-bit
buses; performance-critical routines are copied into IWRAM and compiled as
ARM, everything else stays in ROM as Thumb (16-bit opcodes over a 16-bit bus).
Minish Cap follows the convention exactly: the crt0 is ARM, `AgbMain` and the
bulk of the game are Thumb, and the IRQ handler lives at `$03005D90` — in
IWRAM. Second, **the BIOS owns the exception vectors** at `$00000000`; a game
receives interrupts only through the BIOS's dispatcher, which reads a handler
pointer the game installs at `$03007FFC` (top of IWRAM). §4 shows Minish Cap
doing precisely that.

## 3. The cartridge header (`$08000000`–`$080000BF`)

The 192-byte header, decoded by `tools/platform/gba` (`gbarom` CLI):

| offset | size | field | value in this cart |
|---|---|---|---|
| `$00` | 4 | entry branch | `EA00002E` = ARM `B $080000C0` |
| `$04` | 156 | Nintendo logo bitmap | present, MD5 `e0434707845307679464ae1c22f0ec2d` |
| `$A0` | 12 | title | `GBAZELDA MC` |
| `$AC` | 4 | game code | `BZME` |
| `$B0` | 2 | maker code | `01` (Nintendo) |
| `$B2` | 1 | fixed byte | `$96` (required) |
| `$B3` | 1 | main unit code | `$00` (AGB) |
| `$B4` | 1 | device type | `$00` |
| `$BC` | 1 | software version | 0 |
| `$BD` | 1 | complement check | `$D8` — verified OK |

The game code reads as *B* (second-generation GBA title), *ZM* (Zelda
Minish), *E* (USA/English). The complement check covers bytes `$A0`–`$BC`:
`chk = -(sum + $19)` mod 256 — reimplemented and verified in
`tools/platform/gba/rom.go` (`ComplementCheck`).

The logo bitmap is a compressed image of the Nintendo wordmark that the BIOS
compares byte-for-byte against its own copy at boot — a lockout mechanism (you
cannot ship a booting cart without reproducing Nintendo's trademark), the
GBA's descendant of the Game Boy's scrolling-logo check. We record its MD5
rather than its bytes.

## 4. Boot: the BIOS handshake and the crt0

At reset the BIOS runs from `$00000000`: it shows the boot animation, verifies
the header logo and complement check, and branches to `$08000000` in ARM
state. There the entry word `B $080000C0` hops over the header into the crt0:

```
$080000C0: MOV  r0, #0x12          ; CPSR mode = IRQ
$080000C4: MSR  CPSR_cf, r0
$080000C8: LDR  sp, [pc, #0x24]    ; IRQ-mode stack  = $03007FA0
$080000CC: MOV  r0, #0x1F          ; CPSR mode = System
$080000D0: MSR  CPSR_cf, r0
$080000D4: LDR  sp, [pc, #0x1C]    ; System stack    = $03007F00
$080000D8: LDR  r1, [pc, #0x1C]    ; r1 = $03007FFC  (BIOS IRQ vector slot)
$080000DC: LDR  r0, [pc, #0x1C]    ; r0 = $03005D90  (the game's IRQ handler, IWRAM)
$080000E0: STR  r0, [r1]           ; install it
$080000E4: LDR  r1, [pc, #0x18]    ; r1 = $08055E6D  (odd ⇒ Thumb)
$080000E8: MOV  lr, pc
$080000EC: BX   r1                 ; → AgbMain, in Thumb state
$080000F0: B    $080000C0          ; if main ever returns: reboot
$080000F4: .word $03007FA0, $03007F00, $03007FFC, $03005D90, $08055E6D
```

Nine instructions and a literal pool: stacks for the two modes the game will
ever run in (both in top-of-IWRAM, below the BIOS's reserved words at
`$03007F00`–`$03007FFF`), the IRQ handler pointer installed at `$03007FFC`,
and a `BX` to an odd address — the whole game above this point is **Thumb**
code. The handler target `$03005D90` is an IWRAM address, so the handler (like
all hot code) must first be *copied* there — one of `AgbMain`'s first jobs,
traced in Part II.

`AgbMain` at `$08055E6C` opens exactly like an SDK game should:

```
$08055E6C: PUSH {r4-r6, lr}
$08055E6E: BL   $08055F70          ; six initialization calls —
$08055E72: BL   $080A3204          ;   hardware, IWRAM copy-down,
$08055E76: BL   $0805616C          ;   engine singletons ... (Part II
$08055E7A: BL   $0807CE90          ;   names each one)
$08055E7E: BL   $080560B8
$08055E82: BL   $08056208
```

## 5. The save memory

The header does not declare a save type; Nintendo's library drivers each embed
a versioned ASCII ID, and detection — on flash carts and emulators alike — is
a ROM scan. This cart carries **`EEPROM_V124`** (file offset `$DE7D34`): a
serial EEPROM addressed through the top of the ROM bus region via DMA-timed
bit transfers, either 512 B or 8 KiB (the driver probes at runtime; a
Zelda-sized save will be the 8 KiB part — confirmed when we watch it, Part
II+). No `SRAM_V`/`FLASH*_V` strings appear — the `$0E000000` region is unused
by this cart.

---

# Appendix A — Toolchain and reproduction

New for this game (first GBA title in the repository):

* **`tools/platform/gba`** — cartridge image reader: header decode,
  complement-check verification, save-driver ID scan. `gbarom` CLI:
  `gbarom <file.gba>` prints the header block, entry point and save type.

Reused unchanged:

* **`tools/cpu/arm`** — the ARM7TDMI is ARMv4T; the existing ARM+Thumb decoder
  (built for the DS's ARM7/ARM9) covers it. Part I hand-decodes were produced
  with `disarm -base 8000000 -start 80000c0 -end 8000140 <rom>` (crt0) and
  `disarm -thumb -base 8000000 -start 8055e6c <rom>` (`AgbMain`).

Reproduction of the Part I facts:

```
gbarom "Legend of Zelda, The - The Minish Cap (USA).gba"
```
