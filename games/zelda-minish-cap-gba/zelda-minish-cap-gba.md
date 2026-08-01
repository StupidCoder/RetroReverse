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
* **Part II** — the machine: the GBA oracle (`gbamachine`), what it models and
  what it deliberately does not, and the boot it reproduces — title screen,
  file select, and the opening story sequence;
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
- [Part II — The machine](#part-ii--the-machine)
  - [1. What the oracle is](#1-what-the-oracle-is)
  - [2. The CPU: ARMv4T is a subset, and subsets need saying](#2-the-cpu-armv4t-is-a-subset-and-subsets-need-saying)
  - [3. The bus and its byte-store quirks](#3-the-bus-and-its-byte-store-quirks)
  - [4. The BIOS without a BIOS](#4-the-bios-without-a-bios)
  - [5. The PPU](#5-the-ppu)
  - [6. The EEPROM, and the bug that framed it](#6-the-eeprom-and-the-bug-that-framed-it)
  - [7. Sound: one chip synthesised, one transported](#7-sound-one-chip-synthesised-one-transported)
  - [8. What the boot reaches](#8-what-the-boot-reaches)
  - [9. What is not modelled](#9-what-is-not-modelled)

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

# Part II — The machine

## 1. What the oracle is

**`tools/platform/gba/gbamachine`** is a Game Boy Advance: one ARM7TDMI on
`tools/cpu/arm` over the flat AGB bus, the PPU with its palette/VRAM/OAM, four
DMA channels, four timers, the keypad, the interrupt controller, and the
cartridge's serial EEPROM. It is a **low-level emulation** like `dsmachine`,
`n64` and `psx`: there is no operating system on a GBA, only hardware. The one
thing lifted above the metal is the BIOS's software interrupts (§4).

It is deliberately the single-core sibling of `dsmachine` and shares its shape:
a **scanline scheduler**, interrupt delivery through a Go model of the BIOS's
dispatch shim, an **honest unimplemented-hardware log** (`Machine.Log`) rather
than registers that read back the last value written, and **savestates in the
first phase** — the repository's oracle-capability-parity rule, since every
platform that added them late wished it hadn't. A snapshot taken at frame 900
and resumed for 600 more frames produces a frame **byte-identical** to the
straight 1500-frame run.

The GBA is the smallest machine here since the Game Boy, and the resemblance to
the DS is not a coincidence in the other direction either: the DS's 2D engines
are this PPU's descendants, and the DS BIOS's interrupt protocol is this one's.
Much of `gbamachine` is `dsmachine` with a part removed.

## 2. The CPU: ARMv4T is a subset, and subsets need saying

The ARM7TDMI is **ARMv4T** — and the existing core defaulted to ARMv5TE,
because that is what the DS's ARM9 is. The difference is not academic in the
direction that matters: ARMv4T is a strict *subset*, so a V5TE decoder pointed
at GBA code never *fails*. It silently decodes `BLX`, `CLZ`, `LDRD`/`STRD`,
`PLD`, the saturating `QADD` family and the `SMLAxy` multiplies — none of which
exist on this chip — as those instructions, and executes them.

`tools/cpu/arm` therefore gained a third variant, **`arm.V4T`**, which marks
those encodings undefined in both the decoder and the execution core. Two
details were worth getting right:

* the variant is **not an ordering**. The old code tested `Arch >= V6K` to mean
  "is this the ARMv6 core"; with a third value that is a bug waiting for its
  third caller, so those became explicit `isV6()` equality checks;
* `V5TE` stays the **zero value**, so every existing machine model keeps the CPU
  it had. A test pins that specifically — it is the kind of change that would
  otherwise be discovered by a DS game quietly losing `BLX`.

The V4T tests each carry a **control**: before asserting that V4T *rejects* an
encoding, the same word is pushed through V5TE to confirm it is recognised
there. Those controls immediately earned their keep by failing — and what they
found was two pre-existing bugs in the shared ARMv5 decoder, not in the new
code. `BKPT` was matched against op field `0b1010` when the encoding's is
`0b1001`, and `PLD`'s mask compared against `0xF5` where bits 27:20 masked with
`0xF7` yield `0x55`. **Neither instruction had ever decoded**; both fell through
to `.word` in every DS and 3DS listing this repository has produced. Both are
fixed.

Switching Minish Cap's core from V5TE to V4T left its rendered frame
byte-identical — the expected result, and the one worth recording: the game does
not execute any v5-only encoding, so the guard costs nothing and the listings
are now honest.

## 3. The bus and its byte-store quirks

The address decoder is Part I §2's map. Three behaviours are load-bearing:

* every region is **mirrored** across its 16 MiB window. Top-of-IWRAM
  (`$03FFFFxx`) is not a curiosity — it is how the BIOS's handler slot is
  reached;
* a **byte store to palette or BG VRAM writes the byte into *both halves* of the
  addressed halfword**, and a byte store to OBJ VRAM or OAM is **ignored
  entirely**. A model that just stores the byte boots fine and mis-renders
  sprites, and worse, hides that whole class of bug in the guest;
* the ROM appears three times (`$08`/`$0A`/`$0C`) with different waitstates,
  which this nominal-timing model treats identically, and the `$0D` window is
  where the EEPROM answers.

## 4. The BIOS without a BIOS

There is **no BIOS image**, and none is needed. As `dsmachine` (and the PSX
model) already argue: the BIOS is a *library*, not a kernel — memory fills, a
divider, `ArcTan2`, the affine-matrix helpers, the LZ77/Huffman/RLE
decompressors, and the interrupt waits. Reimplementing that in Go is simpler and
more auditable than interpreting a ROM the repository does not carry, and the
decompressors are the same formats Part IV will need offline.

Two pieces have real behaviour rather than arithmetic:

* **`Halt`/`IntrWait`/`VBlankIntrWait` park the CPU.** It executes nothing until
  the scheduler delivers an interrupt it asked for, so an idle game costs
  nothing and a *missing* interrupt source shows up as a core parked in
  `IntrWait 0x1` rather than as a busy loop that merely looks slow;
* the **interrupt dispatch shim**. The game's handler is not the vector — it is
  a routine the BIOS *calls*. The BIOS pushes `{r0-r3, r12, lr}`, calls through
  the pointer the crt0 planted at `$03007FFC` (Part I §4 watched it do exactly
  that), and returns with `subs pc, lr, #4`, which restores the CPSR — the mode,
  the interrupt mask, and the **Thumb bit**. On a machine whose game code is
  almost entirely Thumb, jumping straight to the handler instead returns into
  ARM state at a Thumb address and the boot dies somewhere with no visible
  relationship to interrupts. `IntrWait`'s other half matters too: the BIOS only
  *clears* the check flags at `$03007FF8`; the **game's own handler** must set
  them, and a game whose handler never does hangs on hardware too.

## 5. The PPU

A per-scanline software renderer for all six modes — 0/1/2 (text and affine
backgrounds), 3/4/5 (the bitmap modes) — with regular and affine sprites in
4/8bpp and both 1D and 2D character mapping, the two rectangular windows plus
the OBJ window, and the full blend/brighten/darken effects including the
semi-transparent-sprite override. Each visible line composes its enabled layers
into buffers and then applies the per-pixel priority, window and blending rules.

Forced blank (DISPCNT bit 7) blanks the screen to **white**, not black — pinned
by a test, because it is exactly the kind of rule that is easy to get backwards
and that looks plausible either way on a screenshot.

## 6. The EEPROM, and the bug that framed it

The cart's serial EEPROM (Part I §5) is addressed one **bit** per bus access
through the `$0D` window: the game lays a request out in RAM a bit per halfword,
DMAs it out, then DMAs 68 halfwords back for the answer.

The first implementation parsed the bit stream as it arrived, and Minish Cap
told me precisely how wrong that was — twice, in the game's own words. The file
select reported *"The data in File 2 is corrupted"*, and name entry ended in
*"Unable to save file."*

The cause is that **a request is framed by the DMA, not by its content**. A read
request is `"11"` + address + stop: 9 bits on a 512 B part (6-bit address), 17
on an 8 KiB one (14-bit). The first nine bits of a 17-bit request are a
perfectly well-formed 9-bit request. A bit-by-bit parser therefore commits to
the wrong address width on the very first request it ever sees, sizes itself to
512 B, and reads and writes the wrong blocks for ever after. Real hardware has
the same ambiguity and resolves it the same way the model now does: the transfer
ends when the DMA ends, so the device counts the bits it was handed and only
*then* decides what it was asked. With that fixed the corruption dialog is gone
— a blank EEPROM reads as a new file, which is what it is.

This is the "a stub that reads ready is indistinguishable from working
hardware" trap in its save-device form, and the lesson generalises: **when a
serial device's request lengths are prefixes of each other, the framing is the
protocol.**

## 7. Sound: one chip synthesised, one transported

The AGB carries the Game Boy's four PSG channels forward and adds its own two
**Direct Sound** PCM channels. The model treats the two halves differently
because the hardware does, and the difference is the whole point:

* the **PSG channels are synthesised**. They are a *description* of a waveform —
  frequency, duty, envelope, sweep — so the model keeps a phase accumulator per
  channel and evaluates it per output sample, exactly as `tools/platform/gameboy`
  does for the same four channels;
* the **Direct Sound channels are transported**. They describe nothing. The
  game's own sound driver mixes its PCM into a buffer in RAM, a DMA channel in
  *special* timing refills a 32-byte FIFO, and a **timer overflow pops one
  signed 8-bit sample** out of it. There is nothing to synthesise; the model's
  entire job is to move the game's bytes at the rate the game's own timer asks
  for.

That second path is why a GBA game's music can be a streamed mixdown rather than
a chiptune, and it makes sound inseparable from the DMA and timer models: get
the timer rate wrong and the music plays at the wrong pitch while every register
in the sound block still reads back correctly.

`bootoracle -wav` captures the final stereo mix at 32768 Hz — the same
verification artefact `tools/platform/dc` and `n3ds` produce, so a future
reimplementation of the game's own sequencer can be checked against the sound
its driver actually drove out of the hardware. `-snd` dumps the whole chain.

**Timing honesty:** the scheduler ticks timers once per scanline, so a FIFO pop
lands with scanline granularity (~16 µs) rather than at its exact cycle. The
*rate* is exact — pops per second equal the timer's overflow rate — so pitch and
duration are right and the error is sub-scanline jitter.

### What the audio found that the video could not

Comparing a captured WAV across a save/resume boundary exposed a **scheduler
bug** that four byte-identical screenshot comparisons had already sailed past.
The run loop checked its stop-at-frame condition *after* `startLine()`, so a run
ending at frame *N* broke out of the line that had just incremented the frame
counter — skipping that line's timers and audio. The display never noticed,
because that line's pixels had already been composed. The sound did: a resumed
run dropped one scanline of timer ticks, produced exactly **two fewer output
samples**, and its Direct Sound FIFO ran one sample behind the straight run for
27 output frames before a refill resynchronised it.

The lesson is the repository's own, in a new costume: *a comparison is blind to
whatever its instrument does not carry.* A frame buffer cannot see a dropped
timer tick. With the check moved before `startLine()`, a resumed run's audio is
now **sample-identical** to the straight run's, and the video stays
byte-identical.

### What the sound says about Minish Cap

Both paths are verified by tests that drive them directly — all four PSG
channels made to sound, and the full Direct Sound chain (timer → DMA → FIFO →
mixer) carrying a known ramp. In the game itself:

* **Direct Sound A carries everything audible.** It plays the logo and intro
  music, and the file-select music after START;
* **the PSG channels are allocated but silent.** `-snd` shows channels 1, 2 and
  4 enabled with frequencies set and **volume 0** throughout the boot — which is
  why "enabled" is a useless diagnostic on its own, and why the dump prints
  volume and frequency next to it;
* **the title screen is silent in this model**, and the reason is not a broken
  audio path: the transport is provably identical to the moments that *do* play,
  and the value being carried is zero. The game's own driver is mixing silence
  there.

Whether real hardware plays a title theme at that point is **open** — settling
it needs a reference capture rather than a guess, and this repository's rule is
to derive facts from the image and our own tools, not from outside sources.

## 8. What the boot reaches

From a cold start, with no BIOS image and no key input, the oracle reaches the
**title screen** (~frame 900) and idles there correctly. Driven with `-keys`
press scripts it goes through **file select**, **name entry**, creates a file,
and plays the **opening stained-glass story sequence** with its text
(`figures/title.png`, `figures/story.png`). The CPU never halts on an
unimplemented instruction along the way, and the only entries in the honest-gaps
log are the sound registers and two mosaic uses (§8).

Instruments on the `bootoracle`, all from the first phase: `-shot` (PNG),
`-savestate`/`-loadstate`, `-keys` (hold or frame-scripted presses, held 20
frames because a game waits for a press *edge*), `-trace`/`-tracefrom`, `-bp`,
`-irqlog`, `-dump`, `-io` and `-log`.

## 9. What is not modelled

Stated plainly, because a gap that is not written down becomes a wrong
conclusion later:

* **Mosaic.** Logged when used; Minish Cap's story sequence uses it on two
  backgrounds, so it is the smallest remaining gap.
* **The BIOS `SoundDriver*` SWIs** (0x1A-0x1E), beyond the harmless
  `SoundDriverVSyncOn/Off` stubs. Minish Cap drives its own mixer and does not
  need them; a game that uses the BIOS driver would.
* **The serial/link port**, which reads as an idle link.
* **Cycle-accurate timing.** The scheduler is nominal: a scanline's worth of
  instructions per line, waitstates and prefetch ignored. Timers tick per
  scanline, which will need to be finer once the sound FIFOs are real.
* **The BIOS as an image** — by design (§4). A direct read of the BIOS region
  (an anti-emulator probe) is logged and returns zero rather than the
  prefetch-latch value real hardware exposes.

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
