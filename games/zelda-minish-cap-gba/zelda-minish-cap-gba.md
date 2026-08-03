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
* **Part IV** — graphics: the starting area lifted out of VRAM — tiles,
  tilemaps and palettes, verified against the machine's own render — and what
  that says about where the ROM-side format hunt begins;
* **Part V** — the world: maps, rooms, the Minish scale mechanic, entities;
* **Part VI** — the script: where the dialogue lives and how a message is built;
* **Part VII** — music and sound;
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
- [Part IV — Graphics: the first export](#part-iv--graphics-the-first-export)
  - [1. Reaching the starting area](#1-reaching-the-starting-area)
  - [2. What the starting area is made of](#2-what-the-starting-area-is-made-of)
  - [3. The exporter, and how it is checked](#3-the-exporter-and-how-it-is-checked)
  - [4. VRAM holds a window, not a room](#4-vram-holds-a-window-not-a-room)
  - [5. The ROM side: finding the room](#5-the-rom-side-finding-the-room)
  - [6. The room format](#6-the-room-format)
  - [7. Decoding it from the ROM alone](#7-decoding-it-from-the-rom-alone)
  - [8. The asset lists — every room, not one](#8-the-asset-lists--every-room-not-one)
  - [9. In the Studio](#9-in-the-studio)
- [Part V — Areas: how the rooms fit together](#part-v--areas-how-the-rooms-fit-together)
  - [1. The room descriptor](#1-the-room-descriptor)
  - [2. The five area tables](#2-the-five-area-tables)
  - [3. Palettes](#3-palettes)
  - [4. 419 rooms, 79 places](#4-419-rooms-79-places)

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

# Part IV — Graphics: the first export

## 1. Reaching the starting area

The oracle is driven from a cold boot into gameplay with one `-keys` script:
title → file select → name entry → the stained-glass story slides → the field
outside Link's house. Gameplay begins around **frame 6400**, and `-shotevery`
turns a single run into a contact sheet, which is how the transition was found
without guessing frame numbers.

## 2. What the starting area is made of

`DISPCNT = 0x1740` — **mode 0** (four text backgrounds), BG0/BG1/BG2 enabled,
sprites on, 1D object mapping. The three layers divide the work exactly as a
2-D Zelda would:

| layer | BGCNT | size | char base | screen base | tiles used | role |
|---|---|---|---|---|---|---|
| BG0 | — | 32×32 | `$0C000` | `$0F800` | 1 | the HUD/text layer, empty in the field |
| BG1 | — | 32×32 | `$04000` | `$0E800` | 54 | the overlay: treetops, grass fringes, things Link walks *behind* |
| BG2 | — | 32×32 | `$00000` | `$0E000` | 176 | the terrain: ground, paths, fences, water |

All three are **4bpp** (16 colours per tile, palette bank chosen per map cell)
and all three are 32×32 tiles — one screenblock, 256×256 pixels. BG1 and BG2
share a scroll position (`8,14`); BG0 does not scroll, as a HUD layer should
not.

A detail worth stating because it is easy to get wrong in an exporter: **a 4bpp
tile carries no colour of its own.** The map entry picks the 16-colour bank, so
a tile sheet rendered with bank 0 throughout is not the game's art — it is one
arbitrary recolouring of it. The exporter colours each tile with the bank the
map actually pairs it with, and reports how many tiles appear under more than
one bank (6 on BG2, 1 on BG1 — the shared fringe tiles).

## 3. The exporter, and how it is checked

`mapexport` writes, per enabled text layer: the tilemap as raw bytes and as
JSON, the character data as raw bytes, a tile sheet of the tiles the map
references (each in its own palette bank), and the whole layer rendered at
native size — plus the palette, the composed screen, and a manifest naming
every register the decode depended on.

The decoding lives in **`tools/platform/gba/tiles.go`**, deliberately separate
from the PPU's scanline renderer: the two answer different questions ("what does
the wire carry on line Y, with this scroll and these effects" versus "what does
the whole map look like"). Two implementations of one format drift apart
silently, so a test renders a synthetic scene through **both** and requires the
pixels to agree.

The export itself is checked the way this repository requires — by opening the
files it shipped, not by re-rendering the structs that wrote them. `-verify`
reads the layer PNGs back off disk, recomposes them with the scroll and priority
from the manifest, and compares against the machine's own background-only render
(`RenderLayers(false)`, which re-runs the PPU over the current video state
without stepping the CPU, since the game would rewrite `DISPCNT` if it were
poked). The result:

```
verify: recomposed the exported PNGs — 0/38400 pixels differ (0.00%)
```

## 4. VRAM holds a window, not a room

The honest limit of this export, established by experiment rather than assumed.
Walking Link a few tiles east and south and re-exporting changes **65% of the
BG2 tilemap bytes** while only **1% of the character data** moves.

So the 32×32 screenblock in VRAM is a **sliding window** that the game refills
as the camera scrolls — not the starting area's map. The tileset is largely
stable across the walk (the field shares its art), but the map is streamed. What
`mapexport` lifts is therefore *the scene the game had already decompressed at
that frame*, bounded by what is on screen; a room the game has not loaded does
not exist to it.

That makes this the right first step and not the last one. It gives byte-exact
ground truth — tilemap, tiles and palette that provably reconstruct the frame —
against which a **ROM-side** decoder can be checked once the compressed map
format is found. The BIOS decompressors are already reimplemented (Part II §4),
which is where that hunt starts.

## 5. The ROM side: finding the room

The hunt started by making the game narrate itself. `gbamachine` implements the
BIOS decompressors, so `bootoracle -declog` reports every one the game
performs — source, destination, size — which is a complete asset-load log
obtained without reversing a single line of the game's loader.

The first attempt returned **nothing**, and that nearly ended the investigation
in the wrong place. The control that saved it was a unit test proving the hook
fires; the real fault was the *window* — the run stopped at frame 1000 and the
first decompression happens at frame 2102. **A zero-hit search is a claim about
the instrument and about where you pointed it.**

With the window widened, the starting area's load appears in full:

| ROM source | destination | size | role |
|---|---|---|---|
| `$0836E448` | VRAM `$06000000` | 16384 | tileset stream 0 |
| `$08370C18` | VRAM `$06004000` | 16384 | tileset stream 1 |
| `$083734F0` | VRAM `$06008000` | 16384 | tileset stream 2 |
| `$08381E9C` | RAM `$0202CEB4` | 16384 | BG2 metatile table |
| `$08384068` | RAM `$02012654` | 16368 | BG1 metatile table |
| `$08385E68` | RAM `$02025EB4` | 5418 | BG2 metatile-id grid |
| `$08386AA8` | RAM `$0200B654` | 5418 | BG1 metatile-id grid |

Everything is **BIOS LZ77** (SWI `$11`/`$12`), reached through a thunk table at
`$080B14D8`/`$080B14DC`.

## 6. The room format

A room is **not stored as a tilemap**, which is why searching a ROM for
something tilemap-shaped finds nothing. It is a grid of **metatile ids**, each
indexing an 8-byte table entry that holds four tile entries — the 2×2 block that
id expands to. The expander was found by watching writes to the destination
buffer (`bootoracle -watch`) and reading the routine at **`$0801AB40`**:

```
LDRH r0,[r6,#0]     ; id from the grid
CMP  r0,r9          ; r9 = 0x3FFF: above it, a separate (animated) path
LSL  r0,r0,#2       ; id*4 ...
LSL  r0,r0,#1       ; ... *2 = id*8 — an 8-byte table entry
ADD  r1,r8,r0       ; base + 0x7004 + id*8
LDRH r0,[r1,#0] / STRH r0,[r5,#0]   ; top-left
LDRH r0,[r1,#2] / STRH r0,[r5,#2]   ; top-right
LDRH r0,[r1,#4] / STRH r0,[r4,#0]   ; bottom-left   (r4 = r5 + one row)
LDRH r0,[r1,#6] / STRH r0,[r4,#2]   ; bottom-right
```

The geometry came out of a disagreement rather than a guess. Comparing our
decompressed id array against the game's own decompressed copy in RAM, the two
matched for exactly **63 entries** and then diverged. 5418 bytes is 2709 ids,
and **2709 = 63 × 43**: the ROM packs the grid **63 metatiles wide**, and the
game copies it into a RAM buffer with a **64-wide stride**. With that, both id
arrays match their RAM copies **100%**.

So the starting area is **63 × 43 metatiles = 126 × 86 tiles**, in two layers
that each carry their own id grid *and* their own metatile table. The pairing
matters: crossing BG2's ids with BG1's table decodes to plausible-looking
garbage.

## 7. Decoding it from the ROM alone

`tools/platform/gba/compress.go` reimplements LZ77 (and RLE) as pure functions
over bytes — a second implementation from the BIOS HLE's bus-based one, guarded
by a test that runs both over the same stream. `roomexport` then decompresses,
expands the metatiles, and renders, with **no emulator running**.

Two verifications, both against the game's own work:

* **the decompressor** — our output versus the bytes the game's own BIOS call
  produced in RAM: the metatile tables and attribute tables are **byte-identical
  (16384/16384 and 4096/4096)**;
* **the whole pipeline** — the expanded map versus the tilemap the game uploaded
  to VRAM: **every entry the game had uploaded matches, on both layers, at the
  same room alignment (56,35)**. Two independently decoded layers agreeing on
  one camera position is not a coincidence one can arrange by accident.

That second check needed care to be honest. A 32×32 screenblock is taller than
the 20 visible tile rows, so its bottom rows are still zero when the export is
taken. Scoring those as mismatches reports **71.9%** for a *perfect* decode —
which reads exactly like a decoder that is nearly right, the most expensive kind
of wrong answer. The verifier now scores only the cells the game actually
uploaded and reports the rest separately.

One more trap worth recording: a 4bpp tile index is 10 bits, so a layer
addresses **1024 tiles = 32 KiB = two** of the 16 KiB tileset streams from its
character base. Handing the renderer only the first stream draws garbage for
every index above 511 — which looks like a decode bug and is not one.

The id arrays, unlike the tables, are **live state**: the game rewrites them as
the world changes, so a comparison must be made against a freshly loaded room,
not a savestate taken after play.

## 8. The asset lists — every room, not one

The four addresses in §5 were found for one area. The index that supplies them
turned out to be hiding in plain sight: searching the ROM for a pointer to any
of those streams finds **nothing**, and the control (searching for a pointer
that certainly exists — the crt0's `AgbMain` literal) proves the search itself
works. They are not stored as pointers.

The loader at **`$080197D4`** explains why. It walks a list of 12-byte records:

| offset | meaning |
|---|---|
| `+0` | source, as an offset **relative to `$08324AE4`**, with bit 31 = "another record follows" |
| `+4` | destination address |
| `+8` | size, with bit 31 = "compressed" |

It also picks SWI `$12` over SWI `$11` when the destination is in VRAM — the
byte-store rule from Part II §3, showing up in the game's own code.

A room needs **two** lists: one carries its tilesets and its two metatile-id
grids, the other the metatile tables, which are **shared between areas** (several
rooms point at the same one). A third, fuller tileset list exists because a room
list is often an *incremental* load that replaces only the character banks that
changed since the last area — take it as the whole story and a bank stays zeroed.

`roomexport -scan` enumerates every list in the ROM that loads a metatile-id
grid: **600 of them**. `roomexport -list ADDR` then decodes any one. Rooms other
than the starting area decode into coherent maps, with two caveats worth stating
rather than papering over: the metatile grid **width is still assumed to be 63**
(confirmed only for the starting area — other rooms almost certainly carry their
own dimensions somewhere not yet found), and **palettes are per-area** and are
not yet resolved from the lists, so another room renders in the starting area's
colours or in greyscale.

## 9. In the Studio

`webexport` publishes the decoded room as a Retro-X game: the two layers are
composited as the hardware stacks them, the result is sliced into 8×8 cells,
and the distinct cells become an atlas (**932 of them** for 126×86 tiles) with
an index grid. The level opens at the camera position the decode was verified
at.

Deduplicating by pixel **content** rather than by the game's tile id is
deliberate: a GBA cell is a tile id *plus* a palette bank and two flip bits, so
one id renders as up to 64 different pictures, and keying the atlas on the id
alone would collapse them and repaint half the map wrong.

---

# Part V — Areas: how the rooms fit together

600 room maps is not a map of Hyrule; it is a pile of screens. The question
that turns the pile into places is *where does each room sit* — and the game
answers it in its own tables.

## 1. The room descriptor

Part IV had to be told a room's width. The game is told too, and finding where
took one climb. The metatile expander writes into a 64-wide RAM buffer, but the
ROM packs the grid at the room's real width, so something re-lays it out: the
routine at **`$0807C8B0`**, which takes `(buffer, width, height)` and copies a
W×H grid into the 64×64 buffer, zero-filling the rest. Logging its arguments
for the starting area gives **63 × 43** — the same numbers derived in Part IV
§6 from a disagreement in the data, now stated by the game itself.

Its caller reads them from a **room descriptor** assembled in RAM at
`$020342CC`:

| offset | field |
|---|---|
| `+0x00` | width in pixels |
| `+0x02` | height in pixels |
| `+0x04`,`+0x06` | position on the area canvas |
| `+0x08` | tileset asset list |
| `+0x0C` | room asset list |
| `+0x10` | metatile-table asset list |
| `+0x14` | the area's script entry point |

Every address Part IV hunted for by hand is in there — and so, crucially, is a
**position**.

## 2. The five area tables

The descriptor is built at **`$08053020`** from five tables, all indexed by area
id (the code holds `area*4` in `r6`):

| table | contents |
|---|---|
| `$0811E214` | → the area's **room geometry**: 10-byte records, `$FFFF`-terminated |
| `$08107988` | → an array of **room asset lists**, one per room |
| `$0810246C` | → an array of **tileset lists**, chosen per room |
| `$0810309C` | → the area's **metatile-table list**, shared by all its rooms |
| `$080B755C` | → the area's **script** |

A geometry record is `{x, y, w, h, tilesetIndex}` — **all in pixels**. That is
the whole answer: an area is a set of rooms placed on one shared canvas. The
starting area is **area 3, room 1**, at (1488, 2480), 1008×688 — and its nine
siblings tile around it into the field that surrounds Hyrule Castle.

## 3. Palettes

The tileset list carries one record whose destination is **0**, which the loader
routes to `$0801D714` instead of copying. Its word is a **palette set id**. A
set indexes `$080FF850` to a run of 4-byte records `{srcIndex, destSlot,
count|flag}`, each copying *count* 16-colour palettes from `$085A2E80 +
srcIndex*32` into palette slot *destSlot*.

An area set covers only the slots it replaces — the starting area's set fills
2&ndash;14 — and the common banks come from **base sets loaded before it**
(`$0B` and `$0C`, read off the running game rather than guessed). With those
applied first the rebuilt table matches the palette the console holds in **254
of 256** entries; the two that differ are both the index-0 slot of a bank, which
is the transparent/backdrop entry and is never drawn.

## 4. Doors

A transition calls `$08051F9C` with **(area, room, x, y)** — the room to enter
and where to arrive — which stores the pair in the room state block and rebuilds
the descriptor through `$08052FF4`.

Counting the call sites settles how doors work: `$08051F9C` is called from
exactly **two** places. One is `$08051F8C`, and the code just above it is the
general mechanism:

```
LDR  r0, [pc, …]      ; the table at $080FCA20
ADD  r3, r3, r0       ; + an offset chosen per door
LDRB r0, [r3, #0x0]   ; area
LDRB r1, [r3, #0x1]   ; room
LDRH r2, [r3, #0x4]   ; arrival x
LDRH r3, [r3, #0x6]   ; arrival y
```

So doors **are** data after all: 8-byte records `{area, room, …, x, y}` at
`$080FCA20`. The other call site, `$080539CE`, loads the destination as
immediate operands (`MOV r0,#0x22`, `MOV r1,#0x11`) and is the *scripted* entry
the game's opening uses to put Link in his house — a special case, not the rule.

That script is one step of a global transition state machine: a state byte at
`$02000086` indexes a table of function pointers at `$080FCCFC`, and each step
advances the byte. It is the sequence that runs a transition, not a per-door
handler.

What remains open is narrow and specific: the record **offset** arrives in `r3`
at `$08051F80`, put there by whatever the player touched. Catching a normal door
in the oracle and reading `r3` there gives the index, and with it the mapping
from a place on the map to a record — the last piece a door marker needs. The
scripted house entry does not take this path, so it cannot answer the question;
a second, ordinary doorway has to be walked through.

### Why that is harder than it sounds

Driving the oracle to an ordinary doorway needs the player to actually have
control, and for a long stretch after the field appears they do not. Measured by
holding a direction and diffing memory against a no-input run of the same
length, the game ignores the d-pad for roughly **3,000 frames** past the point
the field first draws: the intro is still running, and what looked like Link
walking into his house was the opening sequence moving him.

Even past that, movement can be blocked for an ordinary reason — the state that
first responds to input has a **dialogue box open** ("Good morning, Master
Smith"), and the box holds the player still until it is dismissed. A memory diff
across four directions in that state shows the key register changing and
*nothing else*, which reads exactly like "input is not wired up" and is not.

Message selection resists the same probes as the door records, and for the same
reason. A read-watch on the string finds only the renderer at `$0805EFA6`; the
chooser never touches the text. Nor is there a table to find by shape: the ROM
holds **no** absolute pointer table into the script region and no flat run of
increasing offsets. The pointer is already resolved in a message-box context at
`$020227A0` before anything observable happens to it, built from an id the
object supplied — the same shape as the door's record offset arriving in `r3`.

Both therefore need the object the player interacted with, which needs a
genuinely controllable Link and a real interaction.

## 5. Objects

Every room places objects — pots, doorways, NPCs, enemies — and the list
that describes them is **three levels deep**, which is why nothing flat ever
found it:

```
$080D50FC[area] -> a per-room array
            [room] -> a per-SLOT array (8 slots)
              [slot] -> the list itself
```

The first reading of this data called all eight slots "object lists" and
counted 21,036 records. That number was wrong, because the slots are not
eight of the same thing — the room loader treats each index differently:

| slot | contents | walker |
|---|---|---|
| 0, 1 | entity records, spawned on every entry | `$0804ADDC` |
| 2 | entity records with **respawn tracking** | `$0804B058` |
| 3 | 8-byte **command** records, zero-terminated | `$0804B1AC` |
| 4–7 | one function pointer, called directly | `$08000EF2` |

Walked honestly that is **3,605 entities and 564 commands across 467 rooms**;
the other 17k "objects" were slot-3 command bytes mis-parsed as 16-byte
records. (The room index must still be bounded by the area's own geometry
table — the object table carries no length of its own.)

### The entity record

A slot-0/1/2 list is 16-byte records ending in a `$FF` byte, read by
`$0804ADF8`:

| offset | field |
|---|---|
| `+0x00` | low nibble: **class**; high nibble: initial-state flags |
| `+0x01` | low nibble: target entity list; high nibble: spawn **mode** |
| `+0x02` | **id** within the class (entity+9), `+0x03` sub-id (entity+0xA) |
| `+0x04`, `+0x05` | copied to entity+0xB and entity+0xE |
| `+0x06` | u16, meaning depends on the class |
| `+0x08`, `+0x0A` | x and y — **relative to the room**, added to the room's origin |
| `+0x0C` | mode 4: an attached pointer; otherwise per-species data |

Spawn mode 0 places the entity; mode 1 spawns it raw (no position, no extra
fields); mode 4 places it and **attaches the `+0x0C` word** via `$0807DAD0`;
mode 5 skips the spawn if class+id already exists. Class 9 records are never
placed at all — the spawner skips the position step — so their x/y bytes mean
whatever the species wants them to mean.

### What a class is

The class nibble is the entity's kind byte (entity+8), and the evidence for
what each one *is* comes from the loader itself:

* **class 3 — enemies.** Slot 2's walker counts records 0..31 and checks a
  per-index kill flag (`$08049D1C`) before spawning, class 3 only; the index
  is remembered in entity+0x6C. That is the machinery that keeps a defeated
  enemy dead — and only enemies need it.
* **class 6 — objects.** The placed props of the world, 1,865 records.
* **class 7 — NPCs.** 88 of the 111 records use mode 4, and the attached
  pointer is a ROM script for the 139-opcode interpreter at `$0807DFBC`
  (opcode table `$0811E524`) — dialogue and behaviour. Enemies never attach
  scripts.
* **class 9 — managers.** Room-logic entities, never placed.

The behaviour handler is **not in the record**. The per-class update
functions (table at `$080B2248`, indexed by entity+8) dispatch entity+9
through per-class tables:

```
class 3 -> $080D3BF8[id]      class 4 -> $08129320[id]
class 6 -> $080B2D4C[id]      class 7 -> $080B313C + id*12
class 8 -> $080B2CE8[id]      class 9 -> $080B3054[id]
```

Every class's observed maximum id fits its table's extent — 194 entries for
class 6 against a maximum id of 188, 58 for class 9 against 56 — which is the
check that the mapping is real. `objdump -handlers` groups the cartridge's
records by resolved handler; equal ids mean equal behaviour everywhere.

Two species could be named from the map alone. Object id `00` appears in
16-pixel point grids that tile entry regions — six of them cover the door of
Link's house — with the sub byte selecting the destination: a **doorway**.
Object id `2D` sits exactly on the drawn **postbox** beside that door. The
transition a doorway triggers reads an 8-byte `{area, room, _, _, x:u16,
y:u16}` record from the table at `$080FCA20`, index arriving in a game-state
byte (`$02032EC0+3`, read at `$08051F78`) — entry `0x01` of that table lands
back in front of Link's house.

### The slot-3 commands

A slot-3 record is 8 bytes `{op, b1, b2, b3, h4:u16, h6:u16}`, opcode
dispatched through the jump table at `$0804B1CC`. The ones that matter:

* **op 1** — play song `b1` (`$0807CCB4`).
* **op 2 / op 10** — flag-gated **metatile patches**: op 10 writes metatile
  `h6` at position `h4` when save-flag `h2` is set (`$0807B314`); op 2
  registers the inverse while its flag is unset. A position packs as
  `y*64+x` in metatile units — this is how a burned bush or opened passage
  persists on the map. An op-2 patch can name a tile in a *different* room
  of the area, so its position is only trusted inside its own room.
* **op 4** — spawn manager `0x24` at pixel `(h4, h6)` with two byte
  parameters (`$0804B300`) — the only command carrying a pixel position.

### On the Studio maps

`webexport` turns all of this into clickable markers on every area: entities
from slots 0–2 (managers excluded — inventing positions for them would be a
lie) and the positioned commands, each at room origin + record position, the
same arithmetic the compositor uses for the pixels. The info popup shows the
identifiers: class, id, sub-id, slot, respawn index, resolved handler
address, attached script pointer, record address. A marker whose record
position falls outside its own room is dropped rather than plotted — a few
species (enemy `49`) reuse the position bytes for other data, and a wrong
marker is worse than none.

The visual check on the whole decode: the doorway grid tiles Link's front
door, a doorway pair sits on the hollow-tree Minish entrance, the postbox
marker lands on the drawn postbox, and the flag-gated patches cluster where
the map shows disturbable ground. A wrong decode scatters.

## 6. 419 rooms, 79 places

`areamap -list` walks the tables: **419 rooms across the areas that resolve**.
`areamap -all` assembles each area by decoding every room and compositing it at
its own coordinates — **79 areas, every room placed**, in about three seconds.
Two areas (13 and 113) decode no room at all and are reported rather than
quietly skipped.

The shape of the result is the game's own structure showing through. Areas 0–26
are large contiguous overworld chunks; area 3 assembles to 2208×2448 pixels with
a hole in the middle where Hyrule Castle sits — because the castle is a
*different area*. Areas 32 upward hold many small rooms on a sparse canvas:
interiors and dungeons, grouped by region rather than joined into one space.

The Studio now publishes **one level per area** instead of one per room.

---

# Part VI — The script

The dialogue is **plain ASCII**. After a cartridge whose maps hide behind a
metatile indirection, a relative-offset asset list and BIOS LZ77, the script is
simply there: NUL-terminated strings filling the back of the ROM from about
`$08900000`, with `$0A` for a line break.

## 1. Control codes

A message is not only text. Bytes below `$10` are inline commands — speaker and
portrait, the colour that highlights a proper noun, pauses, and the branch a
question takes. They do **not** all carry an operand, and assuming they do
quietly eats the following letter: `0E 59 6F 75` is the code `$0E`, then
`"You"`. Only `$01`&ndash;`$03` were observed to take one.

Highlighted terms are the visible half of this: `[02 02]Lon Lon Ranch[02]` is
the same colour-on, colour-off pairing the game uses for every item and place
name it wants the player to notice.

## 2. Extracting it

`textdump` walks the ROM for NUL-terminated runs and keeps the ones that read as
prose — inside the script region, containing spaces, and at least 70% letters.
All three tests are needed: compressed graphics produce long printable runs
constantly, and length alone finds **10,984** "messages" of which most are
binary. The three together give **3,746**, and spot checks across them are
dialogue, item names and character names throughout. The filter is a heuristic,
so a handful of false positives survive at the edges.

## 3. What it does not give

The messages come out in ROM order, not by who says them. Which line a sign or a
character shows is chosen by an index that has not been decoded: the text engine
at `$0805EFA6` walks a pointer that is already resolved by the time it runs.
Attributing a message to a signpost needs that lookup, which is the same
unfinished thread as the door records — both are *selection*, and both are
reached from the object that the player interacted with.

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
