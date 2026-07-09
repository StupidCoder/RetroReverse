# Stunt Car Racer (Amiga) — technical reference

**Image:** `Stunt Car Racer.adf` — 901,120 bytes, MD5 `b6d3751e6aa636f203f3c6a8de81ebfc`. Not committed (copyright); supply your own copy.

A technical reference for `Stunt Car Racer.adf`, the Amiga release of
Geoff Crammond's *Stunt Car Racer* (1989) — a filled/wireframe-vector stunt
racing game built around an unusually advanced (for its day) rigid-body car
simulation running on elevated 3-D circuits.

The writeup proceeds in reading order, but
the centre of gravity is the last two parts — the tracks and the physics:

* **Part I** — the disk image: the ADF container and the custom (non-AmigaDOS)
  on-disk format — enough to pull every byte off the disk;
* **Part II** — the boot chain: the boot block, the custom track loader it
  bootstraps, and how the game and its data load;
* **Part III** — the game program: the 68000 startup, the interrupt/Copper/
  blitter setup and the memory map;
* **Part IV** — **the tracks**: the vector format of the elevated circuits
  (the track sections, their 3-D geometry and connectivity) and a Go
  reimplementation that extracts and re-draws them;
* **Part V** — **the physics**: the car's rigid-body simulation — the chassis,
  suspension, wheel/ground contact, drive and damage model — reverse-engineered
  from the 68000 integrator and reimplemented in Go so a track can be *driven*.
* **Part VI** — **the 3-D objects**: the opponent car (a procedural
  screen-space build, not a stored model) and the horizon mountain range —
  reverse-engineered, ported and shipped as GLBs.
* **Appendices** — toolchain and reproduction.

Methods: purely static analysis of the disk image, plus the 68000 toolchain in
the shared `tools/` module — the AmigaDOS reader (`tools/platform/amiga/adf`), the
disassemblers (`tools/cmd/dis68k`, `tools/cmd/codetrace68k`) and an
instruction-level 68000 execution core (`tools/cpu/m68k`) for dynamic verification.
All addresses are 68000 addresses; sizes are `.b`/`.w`/`.l` (8/16/32-bit).
**Status: complete. Parts I–III (disk format, loader, engine architecture), Part IV (the
track geometry — footprint, surface, camber and the preview renderer's colours and decals,
decoded and verified exact), Part V (the full rigid-body car physics, verified
frame-by-frame against the original) and Part VI (the opponent car object) are all done;
combined, a car can be driven over the decoded tracks, and the circuits and the opponent
car ship as GLB models in the game's own colours.**

---

## Contents

- [Part I — The disk image](#part-i--the-disk-image)
  - [1. The ADF container](#1-the-adf-container)
  - [2. Not an AmigaDOS filesystem](#2-not-an-amigados-filesystem)
- [Part II — Boot chain and loader](#part-ii--boot-chain-and-loader)
  - [1. The boot block](#1-the-boot-block)
  - [2. The custom track loader](#2-the-custom-track-loader)
- [Part III — Game program architecture](#part-iii--game-program-architecture)
  - [1. Entry, self-check and supervisor setup](#1-entry-self-check-and-supervisor-setup-e700)
  - [2. Hardware bring-up](#2-hardware-bring-up-ed56)
  - [3. Game bootstrap and top-level loop](#3-game-bootstrap-and-top-level-loop-1ba08--5c890)
- [Part IV — The tracks (vector circuits)](#part-iv--the-tracks-vector-circuits)
  - [1. Finding the track table](#1-finding-the-track-table-the-race-setup-path)
  - [2. The record header](#2-the-record-header)
  - [3. The section stream and spine builder](#3-the-section-stream-rle-and-the-spine-builder)
  - [4. A verified Go decoder](#4-a-verified-go-decoder)
  - [5. The spine geometry and re-drawing the circuits](#5-the-spine-geometry-and-re-drawing-the-circuits)
  - [6. The plan footprint (the 16×16 track grid)](#6-the-plan-footprint-the-1616-track-grid)
  - [7. Banking and elevation](#7-banking-and-elevation)
  - [8. The baked track model — the definitive absolute geometry](#8-the-baked-track-model-65bec--the-definitive-absolute-geometry)
  - [9. The drawn faces: colours and decals](#9-the-drawn-faces-colours-and-decals-672cc68f28)
  - [10. The Draw Bridge animator](#10-the-draw-bridge-animator-5a794)
- [Part V — The physics simulation](#part-v--the-physics-simulation)
- [Part VI — The 3-D objects: the opponent car and the horizon](#part-vi--the-3-d-objects-the-opponent-car-and-the-horizon)
  - [1. Placement: wheels ride the decoded track](#1-placement-wheels-ride-the-decoded-track)
  - [2. The screen-space build](#2-the-screen-space-build-599e2)
  - [3. Reimplementation and proof](#3-reimplementation-and-proof)
  - [4. The horizon mountain range](#4-the-horizon-mountain-range-6953e)
- [Appendix A — Toolchain and reproduction](#appendix-a--toolchain-and-reproduction)

---

# Part I — The disk image

## 1. The ADF container

`Stunt Car Racer.adf` is a raw, 901,120-byte Amiga floppy image — the usual
double-density layout of **80 cylinders × 2 heads × 11 sectors × 512 bytes**
(`80 × 2 × 11 × 512 = 901,120`). A byte offset on the disk maps linearly to a
sector: `sector = offset / 512`, with no interleave at the image level. Its
identity is pinned in the repository [README](../README.md#image-files) by size
and MD5 so the analysis stays reproducible.

The image opens with a boot block whose first four bytes are the ASCII tag
`DOS\0` (`44 4F 53 00`) — the standard "bootable AmigaDOS disk" magic — followed
by the boot-block checksum and the boot code:

```
000000  44 4F 53 00 30 AD 90 C0  00 00 03 70  ...   DOS. 0...  rootblk=$370(880)
                    └ checksum ┘  └ rootblk ┘
00000C  24 49 4F FA 03 F0 2C 78  00 04 4E AE  ...   ← 68000 boot code starts here
```

So at face value it looks like an ordinary AmigaDOS disk that boots Workbench.
It is not (§2): the `DOS\0` block is just enough to be *bootable*; the boot code
ignores any filesystem and pulls the game off the disk itself (Part II).

## 2. Not an AmigaDOS filesystem

The boot block names root block 880, but block 880 is **not** a valid AmigaDOS
root header (`tools/platform/amiga/cmd/adfdump` rejects it: *"root block is not a valid
root header"*). There is no DOS filesystem, no directory, no files — the disk is
**custom-formatted**: a flat region of game code and data that only the game's
own loader understands.

This is a common pattern for commercial
Amiga games of the era: a minimal DOS-looking boot block that bootstraps a
bespoke track-loading scheme, both to fit the data densely and to resist
copying. Everything past the boot block — the loader, the engine, the track
geometry and the physics tables — has to be located by following that loader
rather than by reading a filesystem.

---

# Part II — Boot chain and loader

## 1. The boot block

The boot code (from image offset `$0C`) is a compact, self-contained track
loader. It runs while the Kickstart still has the disk inserted, with `a1` =
the boot-time IO request (an `IOStdReq` on `trackdisk.device`) and `a6` =
`ExecBase`:

```
Forbid()                                  ; JSR -$84(a6) — stop multitasking
d0 = $9800 ; d1 = MEMF_CHIP|MEMF_CLEAR     ; ($10002)
a3 = AllocMem(d0, d1)                       ; JSR -$C6(a6) — 38 KB of cleared chip RAM
io_Data($28)   = a3                         ; read destination = the buffer
io_Length($24) = $9800                      ; 38912 bytes
io_Offset($2C) = $2C00                      ; source = disk offset $2C00 (sector 22)
io_Command($1C)= 2  (CMD_READ)
DoIO()                                      ; JSR -$1C8(a6) — read it in
... (retry on io_Error) ...
io_Command = 9 ; io_Length = 0 ; DoIO()    ; motor off
Permit()                                    ; JSR -$96(a6)
JMP (a3)                                    ; enter the loaded code
```

So the boot block **reads a 38 KB blob from disk offset `$2C00` into chip RAM
and jumps to it**. That blob is the game's real loader/engine bootstrap; the
boot block itself does nothing game-specific beyond fetching it. (The boot
block also carries a short ASCII string near offset `$76` — `Prot…` — likely a
title/copyright/"protection" tag; to be transcribed.)

## 2. The custom track loader

The `$9800`-byte blob (disk sectors 22–97) is position-independent 68000 code
(every reference is `LEA …(pc)`), so it disassembles cleanly with its file
offsets as addresses. It does four things:

1. **Take over the machine.** A `MOVE.l #handler,$80 ; TRAP #0` pair jumps into
   supervisor mode at `$10`, then it kills the OS environment: `INTENA`/`INTREQ`
   cleared (`#$7FFF → $DFF09A/$9C`), `DMACON = $7C5F` (bitplane/copper/blitter/
   sprite/disk DMA + master on), a level-3 (VBlank) autovector installed at `$6C`
   pointing to its own handler (`$13A`), then `INTENA = $C020` (master + VERTB).
2. **Load stage data to `$4000`.** With `a0 = $4000` and a return address in
   `a1`, it branches to the disk reader (`$D76`) to pull an initial chunk to
   `$4000`.
3. **Show the title screen.** It unpacks a 4-bitplane image from an embedded
   table (`$D9C`) into chip RAM at `$78000` — four interleaved planes `$1F3E`
   bytes apart, `$FA0` words each (a full lo-res screen) — loads the palette
   (`$D7C → $206`) and points a Copper list at the planes (`$4A2`).
4. **Stream the game and enter it.** The load loop calls the disk reader to read
   **805 sectors starting at sector 110 into `$E700`**, retrying on error (and
   flashing the screen red and hanging if it can't); then it sets the user stack
   (`usp = $7FFFC`), the supervisor stack (`$3D80`) and `JMP $E700` — the game.

**The disk reader (`$570`).** This is a logical-sector reader, not a filesystem.
Its arguments are `d1 = start sector`, `d2 = sector count`, `a0 = destination`
(plus mode bits in `d0`/`d3`). It validates `start + count ≤ $6E0` (1760 =
80 cyl × 2 heads × 11 sectors) and converts a logical sector to a physical track
with `sector / $B` (11 sectors per track-side), then MFM-reads whole tracks (the
helpers at `$C22`/`$664`/`$9E4`/`$746`). Crucially it **only reads** — there is
no decompression — so on disk the data is stored raw, sector-aligned. (A
`"DOW"`-plus-incrementing-byte table at `$96CE` looks like a per-sector
track/format key.)

### The disk map

Putting the boot block and loader together, the image is a handful of raw
regions read by sector — no filesystem, no packing:

| sectors | offset | bytes | loaded at | what |
|--------:|-------:|------:|----------:|------|
| 0 | `$0000` | 1024 | — | boot block (reads the loader, jumps to it) |
| 22–97 | `$2C00` | 38,912 | (chip) | the loader above |
| 110–914 | `$DC00` | 412,160 | `$E700` | **the whole game** (engine + tracks + physics) |

`extract/cmd/extract` slices these straight out of the `.adf`
(`extracted/loader.bin`, `extracted/game.bin`). Because nothing is compressed,
`extracted/game.bin` disassembles directly at base `$E700` — the input for
Parts III–V.

---

# Part III — Game program architecture

The game image (`game.bin`, `$E700`-based) is the 412 KB engine the loader streams
in. Its first bytes are not code that's reached by falling through — they are the
entry, a small trampoline, and an embedded Copper list — so the engine is best read
by following the three transfers out of the entry.

## 1. Entry, self-check and supervisor setup (`$E700`)

```
$E700  MOVEA.l #$E730,a0          ; a0 = start of the checksummed image
       MOVE.l  #$63BA8,d3         ; byte count ($63BA8 = 408 488)
$E70E  ADD.w   (a0)+,d0           ; 16-bit running sum over the whole image …
       SUBQ.l  #2,d3
       BNE     $E70E              ; … d0 = checksum
       MOVE.w  #$5834,d0          ; (expected value; a mismatch path exists)
$E730  MOVE.l  #$E73C,$80.l       ; install a TRAP #0 handler at vector $80…
       TRAP    #0                 ; … to drop into supervisor mode
$E73C  MOVEA.l #$EAD2,a7          ; set the supervisor stack
       JSR     $ED56              ; hardware init (below)
       JMP     $1BA08             ; main bootstrap (below)
```

So the image **checksums itself** ($63BA8 bytes summed as words), enters supervisor
mode through a self-installed `TRAP #0`, sets `a7`, and then does two things: the
hardware bring-up at `$ED56` and the game bootstrap at `$1BA08`. The bytes at
`$E74E` immediately after the trampoline are an **embedded Copper list** (bitplane
pointers `$078000…`, then the colour registers from `COLOR00` at `$180`), which
`$ED56` installs.

## 2. Hardware bring-up (`$ED56`)

`$ED56` is a textbook bare-metal Amiga take-over:

```
MOVE    #$2700,sr                 ; ints off, supervisor
MOVE.w  #$7FFF,$DFF09A / $DFF09C  ; INTENA / INTREQ: clear all
MOVE.w  #$E839,$DFF09A            ; INTENA = master|VERTB|… enable
MOVE.w  #$7CDF,$DFF096            ; DMACON  = master|bitplane|copper|blitter|sprite|disk|audio
MOVE.l  #$EEC8,$64 … #$F000,$7C   ; level-1…7 autovector table (handlers $EEC8/$EF0A/$EF1A/$EF5A/$EFC4/$EFF0/$F000)
…
MOVE.l  #$E74E,$DFF080            ; COP1LC = the embedded Copper list
MOVE.w  #$8380,$DFF096            ; DMACON: enable copper DMA
JSR     $EE8A                     ; CIA setup (timers/keyboard at $BFExxx/$BFDxxx)
;  ── anti-tamper ──
MOVEA.l #$F4B8,a0 ; MOVEA.l #$1AA4A,a2
EORI.b  #$80,(a0)+ ; CMPA.l a2,a0 ; BLT      ; XOR-$80 the range $F4B8..$1AA4A in place
JSR     $F402
RTS
```

Two details matter for the rest of the analysis:

* **The autovector table** at `$64..$7C` wires the seven 68000 interrupt levels to
  the engine's own handlers. The level-3 (VBlank, `$6C`) handler is `$EF1A`; the
  audio handler set lives alongside the sound-buffer pointers initialised just above
  (`$6A584/$6A588/$6A58C = $6A594` and `+$7D00`).
* **The XOR-`$80` decryptor** rewrites `$F4B8..$1AA4A` (≈ 46 KB) in place before the
  game runs. On disk that whole region is obfuscated, so a static disassembly of it
  is garbage until decrypted. `extract/cmd/extract` reproduces this pass and writes
  `extracted/game.dec.bin`; **all addresses in `$F4B8..$1AA4A` must be read from the
  decrypted image.** (Everything `≥ $1AA4A`, including the main bootstrap `$1BA08`
  and the top-level loop, is plaintext on disk.)

## 3. Game bootstrap and top-level loop (`$1BA08` → `$5C890`)

```
$1BA08 clear $EAD6[$80]                     ; small state table
       clear $7A01A..$7B6FA                 ; a screen / work region
       (conditional) fill palette at $E770
       JSR  $69CFC                          ; table init (per-entry loop over $6490…)
       JMP  $5C890                          ; top level
```

`$5C890` sets up the working screen and palette buffers and enters the game's outer
structure:

```
$5C890 MOVE.w #0,$23C32                     ; a mode/state word
       copy $1ECAA[256]  -> $7A91A          ; a 256-byte table (LUT)
       copy $5CF30[…]    -> $7A61A / $7A71A ; two parallel byte tables (championship names/scores)
       TST.b $64AF0 ; BNE …                 ; a "skip intro" flag
       JSR  $62D0A                          ; …
       BRA  $5C960                          ; the main loop body
```

From here the engine is a conventional VBlank-paced loop: a state word
(`$23C32`) selects the current screen (title / menu / track-select / race), the
championship name/score tables (`$7A61A`/`$7A71A`) hold the standings (the
`$5F29A` handlers copy 12-byte `------------` name entries and the `$1C916…`
score counters in and out of them — an early read of these as colour buffers
was wrong; the palette path is the staged table at `$E83A`/`$E85A` pushed to
the Copper list by `$1BA82`, §IV.9), and the level-3 handler (`$EF1A`) drives
timing. The filled-vector race renderer,
the track interpreter (Part IV) and the car simulation (Part V) hang off the race
state of this loop.

*Memory map (run-time, so far):* engine code/data from `$E700`; the encrypted
`$F4B8..$1AA4A` block; sound buffers at `$6A594`; the working screen/Copper at
`$78000`+ and the championship tables at `$7A61A`/`$7A71A`; assorted state
words around `$23C32`/`$64AF0`. The track and physics tables are located in
Parts IV–V.

---

# Part IV — The tracks (vector circuits)

*The first goal.* Stunt Car Racer's circuits are short, elevated 3-D tracks
built from a sequence of **sections** — straights, banked curves, humps, jumps,
ramps and the collapsing "broken" bridge — each a piece of extruded ribbon
geometry with a height profile, joined end to end into a loop. The aim of this
part is to:

1. locate the track table on the disk and decode one section's format (its
   geometry: length, curvature, gradient/height, width, type flags, and how
   sections connect);
2. enumerate the game's built-in tracks (the league circuits);
3. reimplement the decoder in Go and **re-draw each circuit** (a 3-D wireframe/
   plan view).

## 1. Finding the track table (the race-setup path)

The selected track is a single byte, **`$1CA33`** (0–7), set from the track-select
menu and printed by the name-printer `$64C3E` (which does `ASL.b #4,d1` — names are
16-byte records at **`$1EDAA`**, in track order: *LITTLE RAMP, STEPPING STONES, HUMP
BACK, BIG RAMP, SKI JUMP, DRAW BRIDGE, HIGH JUMP, ROLLER COASTER*; the AI opponents
are a parallel 14-byte table at `$1ECAA`).

Starting a race runs `$5D2CA`: `MOVE.b $1CA33,d1` then a chain
`JSR $5AE46 / $64304 / $5A794 / $696FC / $65BEC / $604B4`. The first of these,
**`$5AE46`, is the track loader**, and it is the key to the data:

```
$5AE46 MOVE.b d1,d0 ; ASL.b #1,d0 ; MOVE.b d0,d2     ; d2 = track_id * 2 (word index)
       MOVEA.l #$1F0A2,a2                            ; the track-pointer table
       MOVE.w  $0(a2,d2.w),$1BCC0                    ; word = table[id]
       MOVE.w  $1BCC0,d0
       ROL.w   #8,d0                                 ; byte-swap the word
       SUBI.w  #$B100,d0 ; ANDI.l #$FFFF,d0          ; -> a 16-bit offset
       ADDI.l  #$1EF82,d0                            ; + data base
       MOVEA.l d0,a5                                 ; a5 = this track's data stream
       MOVE.w  #0,d5
       … reads bytes through $5AE00 …
```

So a **track-pointer table of eight words at `$1F0A2`** indexes into the track data:
each word is byte-swapped, less `$B100`, added to the base `$1EF82` to give the
track's absolute address. The reader **`$5AE00`** is a plain sequential byte fetch
(`MOVE.b $0(a5,d5.w),d0 ; ADDQ #1,d5`) — the track is a **byte stream**, decoded in
order, not a fixed-size struct.

Decoding the table gives the eight tracks, stored **contiguously in track order**:

| id | track | address | bytes |
|---:|-------|--------:|------:|
| 0 | LITTLE RAMP | `$1F8E4` | 124 |
| 1 | STEPPING STONES | `$1F960` | 145 |
| 2 | HUMP BACK | `$1F9F1` | 145 |
| 3 | BIG RAMP | `$1FA82` | 142 |
| 4 | SKI JUMP | `$1FB10` | 145 |
| 5 | DRAW BRIDGE | `$1FBA1` | 213 |
| 6 | HIGH JUMP | `$1FC76` | 142 |
| 7 | ROLLER COASTER | `$1FD04` | 312 |

`extract/cmd/tracks` reimplements the `$5AE46` pointer math and writes the eight
raw streams to `extracted/tracks/<id>_<name>.bin` — the input for the format work.

## 2. The record header

Each track stream begins with a short parameter header that `$5AE46` reads first
(five bytes into `$1CA19[1..5]`, then two more into `$1BBF6..9`). The first eight
bytes of all eight tracks share a clear shape — **byte 1 always equals byte 2**:

```
LITTLE RAMP     2c 0f 0f | 25 00 05 a0 cf …
STEPPING STONES 38 2a 2a | 0e 00 0f a0 cf …
HUMP BACK       35 2e 2e | 13 40 05 60 04 …
BIG RAMP        2c 01 01 | 18 80 07 a0 c0 …
SKI JUMP        28 0f 0f | 23 40 6a aa bd …
DRAW BRIDGE     4e 2a 2a | 04 a0 11 a0 cc …
HIGH JUMP       34 1d 1d | 04 40 06 20 3f …
ROLLER COASTER  4e 00 00 | 25 00 05 a0 cf …
```

`$5AE46` reads these into `$1CA1A..$1CA1E`, which fixes their meaning:

* **byte 0 = section count** (`$1CA1A`) — the loop below runs until the section
  index reaches it. Per track: 44, 56, 53, 44, 40, 78, 52, 78 — i.e. each circuit
  is 40–78 sections long;
* **byte 1 = byte 2 = the finish/start section index** (`$1CA1B`/`$1CA1C`) — a
  duplicated value, used to wrap the lap (`$5B10C`: start from `$1CA1C+1`);
* **bytes 3–4 = the initial position seed** (`$1CA1D`/`$1CA1E`; `$1CA1E` is later
  overwritten with the total track length).

`extract/cmd/tracks` prints all of this (`sections=… finishIdx=…(dup=true) seed=…`).

## 3. The section stream (RLE) and the spine builder

After the header, the body of `$5AE46` (`$5AED0..$5B106`) walks the byte stream once,
emitting one **section** per iteration into a set of parallel arrays indexed by the
section number `d1`:

| array | per-section meaning |
|-------|---------------------|
| `$1C5EC[]` | type / flags byte |
| `$1C588[]` | parameter 1 |
| `$1C4C0[]` | parameter 2 |
| `$1C524[]` | combined attribute (`param & $7F | $1BB1B`) |
| `$1C650[]` / `$1C718[]` (word) | the section node's **X / Z** coordinate |
| `$1C7E0[]` (word) | accumulated distance along the track (`<<5`) |

**Encoding.** Each section normally costs a type byte plus 1–2 parameter bytes. Two
mechanisms compress the common case of many similar sections in a row:

* **Run-length.** If a type byte's **low nibble is `$F`**, its high nibble is a
  *run count* (`$1BB48`): the following `n` sections repeat the saved type
  (`$1BC55`) with their parameter stepped by the delta helper `$5AE0C` — which adds
  or subtracts either `$10` or `$1` depending on flag bits in `$1BB1A`. This is what
  encodes a constant curve or a constant gradient as one marker plus a step.
* **Type-nibble ≥ `$C`.** These section types take their second parameter from the
  small tables `$5B2B8`/`$5B2BA` instead of from the stream (the standard ramp/jump
  piece shapes).

**Spine layout.** As it emits each section, the loop also places it in 3-D. The
running position is `(X,Z) = ($1BBF6,$1BBF8)`, advanced per section by a helper
`$5B2D2`/`$5B2C8` that — notably — **reuses the very same handle-decode formula as
the track-pointer table** (`ROL.w #8`, `−$B100`, `+$1EF82` on a 16-bit handle in
`$1BC8C`) to follow a *section-shape* sub-table inside the track data, indexed by the
section type. So section "types" are references to reusable extruded-piece shapes,
and the per-section dx/dz come from those shape profiles rather than from a sine
table. The node coordinates are stored in `$1C650[]`/`$1C718[]`; the cumulative
distance in `$1C7E0[]`; and the grand total length in `$1CA1E` (`$1BC2A << 5`).

**Trailer.** After the sections, the stream carries a fixed 6-byte block (`$1CA2A`),
then two optional object lists guarded by counts: `$1CA2E` pairs into
`$1C8A8[]`/`$1C8C8[]` and `$1CA2F` entries into a further array — the trackside
scenery / markers.

So a track decodes to: *a header (count, finish index, seed) → an RLE section stream
→ a scenery trailer*, with each section a `(type, p1, p2)` triple where the type
indexes a shared piece-shape table that supplies the geometry.

## 4. A verified Go decoder

`extract/cmd/sections` reimplements the `$5AED0..$5B106` loop exactly — the
run-length marker, the `$5AE0C` delta step, the `≥$C` table branch and the
`bit5`/`bit6`/`bit7` flag-driven reads — and then checks itself against the data.
The whole track stream is:

```
6 (header)  +  section stream  +  6 (trailer)  +  2*trailer[4] + trailer[5] (scenery)
```

(Trailer bytes 0/1 — loaded to `$1CA2A`/`$1CA2B` — are the track's **damage windows**:
how many ticks of a sustained over-tolerance suspension impact accumulate chassis damage;
see Part V §3.)

so a correct parse must consume a track's bytes **exactly**. It does, for all seven
length-bounded tracks:

```
LITTLE RAMP     44 sections  used 124/124   STEPPING STONES 56  145/145
HUMP BACK       53 sections  used 145/145   BIG RAMP        44  142/142
SKI JUMP        40 sections  used 145/145   DRAW BRIDGE     78  213/213
HIGH JUMP       52 sections  used 142/142   ROLLER COASTER  78  used 192 (last)
```

(ROLLER COASTER, the last record, is 192 bytes — the raw slice in `cmd/tracks` runs
on into padding; the parser gives its true length.) The byte-exact match across every
track confirms the read pattern is right. As a spot check, LITTLE RAMP's first ten
sections are a single `$A0` run: a fresh `type $A0, p1 $CF, p2 $6A` section, then one
marker byte `$9F` (run = 9) followed by nine `p2` bytes, with `p1` stepping `−$10`
each section — exactly the bytes on disk.

What remains for the circuits is geometric, not structural — covered in §5.

## 5. The spine geometry, and re-drawing the circuits

The section loop also lays the track out in plan view as it parses. Per section,
given its decoded `(type, p1, p2, attr)`, the per-section setup `$5FE56`/`$5FEFA`
resolves the geometry from three data tables — all reached by the same
byte-swap/`−$B100`/`+$1EF82` handle decode as the track-pointer table:

* **`$1EF82 + ((type & $F) << 8)`** → a per-type *piece-shape* table. Its layout
  yields the section's **segment length** (`$1BB6A = $1BB97/2 − 1`, where
  `$1BB97 = shape[shape[0]]`) and related step counts.
* **`$1EFA2[p2*2]`** → the **X-offset shape** handle (`$1BC8C`); **`$1EFA2[attr*2]`**
  → the **Z-offset shape** handle (`$1BC90`). `p2` also sets the sign mode `$1BB79`.

The shape readers `$5B2D2` (X) and `$5B2C8` (Z) return the piece's offset at a given
distance along it, and the loop walks the spine by

```
node.X = cur.X − shapeX(0) ;  cur.X = node.X + shapeX(len)
node.Z = cur.Z − shapeZ(0) ;  cur.Z = node.Z + shapeZ(len)
```

storing each node in `$1C650[]` (X) / `$1C718[]` (Z) and accumulating the heading in
`$1BC2A`. So a straight piece (where `p2 == attr`, common in the runs) advances X and
Z equally — the world axes are the 45°-rotated `(x±z)` pair — and curves bend them
apart.

**The Go reimplementation (`extract/track`).** `package track` walks this spine
purely in Go, reading only the game's data tables — it does *not* run any 68000 code.
The handle math has two byte-width subtleties worth recording (the format is shared
with the 6502/Z80 ports and so cannot depend on any 68000 quirk):

* the per-type shape table is a **word table at `$1EF82 + nib*2`** — no shift (the
  `ASL.l #7` that reads out adjacent is the low word of the preceding
  `MOVEA.l #$1EF82`, not an instruction);
* the X/Z offset-shape handle index is `ASL.b #1` on the param, i.e.
  **`(param*2) & $FF`** — a byte multiply that wraps, not a 16-bit one.

`extract/cmd/spineoracle` executes the real `$5AE46` on the `tools/cpu/m68k` core (it
takes only `d1` = track id and self-initialises) and reads the `$1C650`/`$1C718`
arrays back out; `extract/cmd/spineverify` checks `package track` against it.
**All eight tracks match coordinate-exact** — the Go decoder is the source of truth,
the oracle only confirms it.

**These arrays are not (yet) the world centre-line.** Plotted directly, `$1C650`/
`$1C718` form thin slivers (plan aspect 6:1 to 23:1 — Ski Jump is almost a straight
line), which are *not* the shapes of the real circuits. Tracing the renderer shows
why: the per-section vertex builder `$5C0AA` returns `shape_offset + $1BC0E/$1BC10`,
where `$1BC0E/$1BC10` are exactly `$1C650[i]`/`$1C718[i]` — so these arrays are the
section *base positions*, but the renderer (`$5A1A6`) also reads **two** offset shapes
per vertex (`a4` from `p2`, `a5` from `attr`, alternating), and rotates everything by a
per-section **heading** (`$1BBF2 = $1BC30 − $1BC4A`, with `$1BC4A = type & $C0` the
section's quadrant and `p1` the intra-quadrant fine direction via `$5FF94`). Strong
evidence that one of `$1C650`/`$1C718` is the **vertical** (the `a4`/`a5` pair reads as
horizontal/vertical cross-sections), so the raw plot is closer to a side profile than
a plan.

So `$1C650/$1C718` are per-piece *extents*, not the plan. The plan footprint is held
elsewhere — and reading the renderer, not guessing, finds it (§6).

## 6. The plan footprint (the 16×16 track grid)

The engine renders the track by walking from the player's section outward, and to find
the next section it consults a lookup table. `$5FE04` is that lookup: it forms a byte
index `(y << 4) | x` from a grid coordinate and reads the **256-entry table `$1C280`**,
returning the section sitting in that cell (or `$FF` for empty). That table is built by
`$64304`:

```
for each section i:  $1C280[ p1[i] ] = i        ; p1 indexes the 16×16 grid
```

So **`p1` is the section's cell on a 16×16 grid** — its low nibble the grid X, its high
nibble the grid Y. That is the track's plan footprint, read straight from the data:

```
node[i].planX = p1 & $0F
node[i].planY = p1 >> 4
```

No heading accumulation, no quadrant transforms, no fitting — and it is unambiguous:
**every consecutive section is an adjacent grid cell** (0 non-adjacent steps across all
eight tracks), so the plans are smooth closed loops with no sharp corners and no
spurious self-crossings. The footprints are the real circuits — *Little Ramp* a
diamond, *Stepping Stones* a rectangle, and *Draw Bridge* a loop with the two inner
prongs that are literally its drawbridge:

```
.##..........##.        .##....########.
#..#........#..#        #..#..#........#
#..#........#..#        #..#..#........#   Roller Coaster (right):
#..#........S..#        #..#..#........S   an outer lap with an
#..#........#..#  ...   #..#..#........#   inner peninsula the
#...########...#        #...##.........#   track doubles back along
#..............#        #..............#
.##############.        .##############.
  Draw Bridge             Roller Coaster
```

`package track` sets `PlanX/PlanY` from `p1`; `cmd/trackjson` exports them; the viewer
draws them.

## 7. Banking and elevation

The game's pre-race **preview** (the "steer to rotate view" screen) draws each circuit
as an *elevated ribbon on support columns*, which makes the two vertical quantities
plain: the ribbon **banks** on the curves, and it **rises and falls** (the ramps, the
Roller-Coaster hills) on columns of varying height. Both come, cleanly, from the
loader's two per-section extent arrays.

Tracing the two vertex builders settles which is which:

* `$5C6C4` reads the section's per-type **piece-shape** and rotates *both* its
  components by the section's orientation (its four quadrant cases are a 2-D rotation).
  A quantity that rotates with heading is **in-plane** — so this builds the plan
  footprint, not the height.
* `$5C0AA` reads the two extent arrays `$1C650` (from `p2`) and `$1C718` (from `attr`),
  adds them **un-rotated** (and `>>5`), and *those* are the heights of the ribbon's
  **left and right rails** — a vertical doesn't turn with the heading.

So the two rail heights give both quantities directly:

```
elevation = ($1C650 + $1C718) / 2      (the surface height)
banking   =  $1C650 − $1C718           (the camber: rail-height difference)
```

On a straight the section's `p2 == attr`, the two rails are level, banking is zero;
on a curve they diverge and the road tilts. The elevation profile **matches the
preview screenshots**: Big Ramp is flat with a single ramp, Stepping Stones gently
undulates, Roller Coaster is a run of hills — and every circuit's height **closes**
over the lap (Hump Back exactly).

**Smooth or stepped — solved, exactly.** A long search for an explicit "step here" flag
(`$65D3C`, the `type` bits, the `p2` sign, `attr`, `a0[1]`, the height magnitude) found
nothing — *because there is no such flag*. The surface is not one height per section at
all. `$5C0AA` is a **per-rung** builder: it is called for vertex index `d1 = 0,1,2,…`
across the section, and each call adds the section's base accumulator (`$1C650` for the
left rail, `$1C718` for the right) to a **profile value** read from the `p2`/`attr`
cross-section shape, then `>>5`. So every section carries a short array of rail heights
*along its length* — the in-game surface itself:

```
rung count   = shape[shape[0]] of the per-TYPE piece-shape   ($1BB97; counts both rails)
left rail[k]  = ($1C650 + a4-profile[k]) >> 5                  (even vertex 2k,  a4 from p2)
right rail[k] = ($1C718 + a5-profile[k]) >> 5                  (odd  vertex 2k+1, a5 from attr)
read mode     = nibble-packed if p2 ≥ 0, two bytes/entry if p2 < 0   ($1BB79)
```

The step-vs-slope distinction is simply **what those profile values do**. A smooth run is
a drivable slope; a sudden jump in consecutive rungs is a hard edge. Big Ramp's takeoff
section reads `[80 85 90 95 100 105 110 115 35]` — a clean climb to the lip (115) then a
hard drop into the gap (35); the next piece resumes at 35 and the landing ramp rises again
`[35 105 100 …]`. That is **three jumps, six steep lips**, present in the data, with no
heuristic. The Stepping Stones read `[118 36 36 …]` (a flat stone with a square gap), the
Ski Jump's launch is a `704 → 389` drop mid-section, and the Roller Coaster is a long
continuous `281 → 925` climb. No magnitude rule, no extremum test — the profile *is* the
surface.

This is verified the strict way: `cmd/geomoracle` runs the engine's own `$5FE56`+`$5C0AA`
on the m68k core and the Go reimplementation in `package track` (`railHeight`/
`railProfile`) reproduces every rung of every section of all eight tracks **coordinate-
exact**.

### The plan outline (`$5C6C4`)

The other half of each piece is its **plan footprint** — the `(x,z)` shape of the rung
strip seen from above. The draw loop (`$65CDC`) reads it with `$5C6C4`, which fetches two
consecutive 16-bit little-endian values at byte offset `d2` in the per-type piece-shape
and adds the section base, rotating by quadrant. The vertex pairs are a flat array:

```
plan vertex k = LE16 pair at  a0[0] + 7 + 4k     (even k = left rail, odd = right rail)
                d2 = 2*d1 + $1BB91 ,  $1BB91 = a0[0]+7        (the draw loop's index math)
```

Read out, these are exactly the piece's local outline (`+z` forward, `x` lateral): the
straight piece (`nib 0`) is two parallel rails of width 384 marching `z = 0,256,…,2048`;
the curve pieces (`nib 6`/`7`) are mirror-image arcs; ramps carry their own profiles. The
left/right rails are independent, so the rail **width** (and any asymmetry through a bend)
is exact, not a nominal constant. `cmd/planoracle` drives `$5C6C4` in the canonical frame
(zero base, zero quadrant) and confirms `package track`'s `planProfile` matches it
**exactly** on every vertex of all eight tracks.

## 8. The baked track model (`$65BEC`) — the definitive absolute geometry

The race-setup chain (`$5D2CA`: `$5AE46` load → `$64304` grid → `$5A794` → `$696FC` →
**`$65BEC`**) runs a **bake** that turns the loaded section arrays into the polygon model
everything else draws: per-rung records `{word0, Lx, Lz, Lh, Rx, Rz, Rh}` written to the
buffer at `$7ABDA`, with a per-section index at `$7AA1A`. The per-frame renderer never
recomputes geometry; it re-places these baked records each frame. Annotated flow:

```
$65BEC  bake entry: $65DEA (per-section LOD bytes -> $65E70), record ptr $66102=$7ABDA
$65C0E  per section: $7AA1A[sec] = record ptr; $66100 = 0
$65C2E  $5C51A advance + $5FE56 = peek the NEXT section: if its piece header flags
        ($1BB4D = a0[1]) bit7: $66100 = $4000; and if ($1BC44^$1BC32) bit7: = $8000
$65C66  $5C538 retreat + $5FE56 = set up the CURRENT section; $1BC26 = 0
$65C7A  $1BBF2 = -$1BC4A (NEG.b)      ; section quadrant in a FIXED world frame
$65C8A  $65756: fill the height strip $1BE70 (left/right full-precision rail heights,
        profile entry + $1C650/$1C718 base — no >>5) and the vertex-flag strip $1BDD0:
        a bit-7 marker on a profile byte CLEARS the flag = "this rung is a polygon
        edge"; unmarked rungs keep the $80 the previous emit left = skipped; the last
        rung is preset (1 = finish line when sec == $1CA1C); |dh| >= $280 vs the
        previous rung sets the $20 crease bit; a bit-7 marker after the LAST entry
        hides the last rung entirely (gap pieces — the Stepping Stones edges)
$65CBC  $1BB91 = a0[0]+7 (vertex 0's byte offset), + (cnt-1)*4 when the piece is
        REVERSED (type&$10, $1BC32): reversed pieces run their shape backwards, which
        also swaps the left/right rails
$65CDC  emit loop over d1 = 0,4,..,2*cnt: skip flagged ($80) rungs and d1 < 4 (rung 0
        is the previous section's last rung); emit when d1 >= $1BB5A (the last rung, or
        last-1 when the last is gap-hidden), the piece is a type-2 ramp ($1BB4D bit7),
        p2 == $25 && d1 == $18 (one piece's quirk), or the height second-difference
        |h[d1]/2 - (h[d1-4]+h[d1+4])/4| >= $50 (this is what keeps a rung on a lip);
        record word0 = (flag&$3F)<<8 | $66100 | sec, then for L (d1) and R (d1+2):
        d2 = $1BB91 +- 2*d1, JSR $5C6C4, write $1BBF6, $1BBF8, height $1BE70[d1]
```

`$5C6C4` itself (traced in full) reads one vertex per call — an LE16 `(u,v)` pair at
offset `d2` in the piece-shape — and writes the **absolute plan point** `$1BBF6/$1BBF8 =
base($1BB22/$1BB26) + rot(u,v)`, where the rotation by the quadrant byte `$1BBF2` is one
of exactly four cases (`$800` = one grid cell):

```
q0 (00): X = bx + u          Z = by + v
q1 (01): X = bx + $800 - v   Z = by + u
q2 (10): X = bx + $800 - u   Z = by + $800 - v
q3 (11): X = bx + v          Z = by + $800 - u
```

In the bake the base is **zero** (the init chain leaves `$1BB22/$1BB26 = 0`, checked in
the oracle), so the records are **cell-local** coordinates in `[0,$800]`. The per-frame
draw (`$65EC4`) proves the placement convention: per visible section it calls `$5FF94`
(section cell − camera cell, rotated by the *camera* quadrant), scales the delta by one
cell (`<<10` at half resolution), applies the same four reflections (`$800 − stored`) to
the stored local values, and adds. So the absolute plan position of every vertex is

```
world = gridCell(p1) * $800  +  baked local vertex
```

— no accumulation, no fitting, no spline. The curves join the straights exactly because
each piece's arc *ends where the data says the next cell begins*.

**Reimplementation and proof.** `track.Bake` (`extract/track/model.go`) reproduces the
whole bake in pure Go from the disk image — flag strip, height strip, decimation,
reversed pieces, gap hiding, record words. `cmd/modeloracle` runs the engine's real
init chain + `$65BEC` on the m68k core and compares **every record of every section of
all eight tracks byte-exact** (51/69/54/51/44/82/59/78 records) — OK on all eight.
`track.Geometry` then emits every rung (the decimated ones flagged as the game's drawn
edges) in absolute coordinates; section joints are continuous to within 4 units out of
2048 per cell (the residue is in the game's own shape tables, not the decode), and every
circuit closes. `cmd/modelsvg` renders the top-down plans; `cmd/trackjson` exports
`rungs[i][k] = [Lx,Lz,Rx,Rz,Lh,Rh,flags]` and the viewer now draws the baked model
directly. **Part IV is complete: footprint, surface, camber
and now the absolute plan are all decoded purely from the disk in Go and verified
byte-exact against the original renderer's own build.**

## 9. The drawn faces: colours and decals (`$672CC`/`$68F28`)

The pre-race track preview (`$604B4`) is the renderer at full strength: it walks
all 16×16 grid cells (`$602E4`/`$60324`), recomputes every section at full rung
resolution (`$6521C` → the `$65756` strips of Part IV §8) and, because the
preview flag `$1BB68=$80` makes `$654C2` emit a wall-bottom vertex for every
rail point at height `$200` (the ground plane), draws the whole circuit as a
walled ribbon standing on the ground. The emit `$672CC` writes one display-list
strip per rung pair — four lengthwise edges (two road rails, two wall bottoms),
two wall verticals, one road cross-line — closed by a 4-byte descriptor
`{section, type, flag, piece}`; the flush `$68F28` walks the list backwards in
batches that end at each *drawn-edge* rung (flag-strip bit 7 clear) and fills
and strokes each batch. Everything the player sees is decided as follows:

- **Palette.** The 16 words at `$2B974` are the race/preview palette. They are
  staged at `$E83A` (current) / `$E85A` (fade target, loader `$1BA70`), the
  fade `$64470` steps each 4-bit channel per frame, and `$1BA82` pushes the
  staged words into the embedded Copper list's `COLOR` block at `$E770+4n`
  after doubling every channel and setting its low bit when non-zero
  (`c → 2c|1`): stored `$443` displays as `$997`. The displayed values:
  `$055` background (0), `$997`/`$BB9` road greys (1/2), `$FF0` yellow (3),
  `$555` (5), `$500` dark red (9), `$733` red (`$A`), `$777` (`$D`/`$E`-ish
  greys), `$FFF` white (`$F`).
- **Road surface** (`$68D82`): each batch — the span of strips up to a drawn
  edge — is filled as one polygon in palette **1**, or **2 when the section
  number is odd** (`$691C4` tests bit 0 of the descriptor's section byte). A
  batch whose descriptor carries the crease bit `$20` is filled with the
  **background colour 0** instead — the dark bands leading up to ramp lips.
- **Side walls** (`$68A56`/`$68BEC`): filled palette **`$F` (white)**, or
  **`$A` (red) when the section is odd** — the classic alternation is *per
  section*, in step with the road greys. A cell viewed from its back side
  (`$1BB66`, from the camera/section quadrants) is filled `$9` instead — the
  dark inner face of the wall.
- **Curb strokes** (`$689CE`/`$68A12`): the road-edge lengthwise lines are
  stroked in the descriptor's type byte — **9, or 3 on alternating rungs**
  (`$67120`: parity of the rung index XOR a per-section phase). The phase is
  computed *by the track loader*: `$5AE46` accumulates
  `$1BBDC = ($1BBDC + cnt−2) & 2` across sections from zero and stores each
  section's phase bit as bit 7 of its `$1C524` byte (`$5B010`/`$5AF7C`) —
  which is what keeps the yellow/dark-red dashes continuous over every joint.
- **Wall verticals** (`$6892A`/`$6897C`): emitted on parity-even or drawn-edge
  rungs and stroked palette **9** on type-9 strips only — the strut lines
  climbing the walls on every other rung.
- **Cross-lines**: emitted at drawn-edge rungs. The preview clears the first
  and last rung's flag byte outright (`$6045C`), forcing a lateral edge at
  every section joint — and wiping the crease and finish bits there, so the
  preview never draws the white finish stroke (`$688FC` is race-only).
- The screen is 4 bitplanes (`BPLCON0 = $4200`); a face colour is a 4-bit
  palette index that `$66348` patches into the plane-select fill code — there
  is no per-scanline Copper trickery in the scene.

The preview projection halves the plan (`$62454`) and squeezes heights by
`$4C1B/$8000` (`$624C2`) — screen framing only. The true proportions are the
race view's: plan enters the projection unscaled and heights are `ASR #2` —
**4 height units to 1 plan unit**.

`track.Mesh` (`extract/track/mesh.go`) reimplements the whole assignment in
pure Go — palette, fills, strokes, phases — and `cmd/coloracle` verifies it by
running the real `$604B4` on the m68k core for all four preview camera angles
with hooks on every colour decision: **all eight tracks match, every rung
covered, palette word-exact**. `cmd/webexport -only models` bakes the result
into one GLB per circuit (`models/<slug>.glb`): road and wall quads per rung
pair with the exact linearised Amiga colours, and the stroked decals as true
glTF `LINES` primitives — plus the race view's white start/finish cross-line
(`$688FC`, the baked flag bit 0 the preview wipes) as a palette-`$F` segment.

## 10. The Draw Bridge animator (`$5A794`)

Track 5's two bridge ramps (sections 51/52 and 54/55, around the gap section
53) have no fixed geometry. Their four height-profile handles (`$5F..$62`) are
**overlapping 9-entry windows into one 36-entry table**, and the disk only
holds a placeholder pattern there (a narrow spike per window — the game never
displays it). `$5A794`, gated to track 5 by the menu byte `$1CA33`, rewrites
the table:

```
tri  = |(phase & $1F) − 16|      ; triangle wave 15..0..15 over 32 steps
step = (tri + 4) << 5            ; per-rung rise, 128..608 height units
```

Entries 1–16 become the accumulating ramp `k·step` (entry 9 repeating entry
8's value — the shared joint rung between sections 51 and 52 — and entry 16,
the lip, carrying the profile's bit-7 edge marker), and every write is
mirrored at entry `35−k`, producing the second ramp descending; entries
0/17/18/35 stay zero. Through the four overlapping windows this reads back as:
section 51 ascending, 52 ascending to the marked lip then dropping to the gap
floor, 53 flat, 54 rising to its lip at rung 1 and descending through 55 —
with the lips swinging between 1 920 and 9 120 height units above the deck as
`tri` cycles.

Cadence (all traced): the race loop calls the animator once per frame
(`$5D49C`); the phase advances unless the `$EE` time-base accumulator missed
its carry (`$5DB34`: `$1BBCF += $EE` twice per frame, `$1BBCD` = the last
carry — 18 frames in 256 skip), and freezes entirely while the player's or
the opponent's section is `$33..$37` (a car on the bridge). One full up-down
cycle is 32 phase steps, holding two steps at each extreme. The race loop
itself is **unthrottled** (there is no VBlank or raster wait anywhere in the
race path), so the absolute rate is machine-dependent. Even the pre-race
preview runs one animator pass first (`$5D2DA`), advancing the phase from 0
to 1 (`tri` 14) — the near-fully-raised "static ramps" the preview shows.

`track.Drawbridge` reimplements the patch in pure Go and `cmd/bridgeoracle`
verifies it against the engine's own `$5A794` over 64 consecutive frames,
byte-exact, including the freeze gate. (Note: `cmd/modeloracle` never sets
`$1CA33`, so its track-5 verification covers the placeholder state on both
sides — consistent, but not what the game shows; the in-game state is what
`bridgeoracle` covers.) The exports use the patched profiles: the track JSON
carries the first-preview pose, and `draw-bridge.glb` ships the animation as a
**morph target** — base mesh at the lowered pose, one target raising it, and
a LINEAR weight track reproducing the triangle wave with its two-step holds
and the `$EE` stretch (nominal 12.5 fps A500-class race rate, ≈2.75 s per
cycle).

*Part V — the physics.*

---

# Part V — The physics simulation

*The headline goal — done.* The car is simulated as a sprung rigid body, not a point,
which is what gave the game its reputation: the chassis pitches and rolls on its
suspension, the wheels gain and lose contact over crests and on landings, too hard a
landing damages the car, and airtime/handling depend on speed and the track gradient.
The **entire** per-frame simulation has been reverse-engineered from the 68000 code and
reimplemented in Go (`extract/physics`), **verified coordinate-exact against the original
frame by frame** on the `tools/cpu/m68k` core (`cmd/physverify`): every sub-routine matches on
3000 random states each, and the whole frame matches bit-for-bit over 60 consecutive
frames on three tracks. The reimplementation operates directly on the game's 24-bit memory
image, so it can be checked address-for-address.

### 1. The car state and the frame

One physics frame is `$6185C`, called from the race loop (`$5D42C`) just before the
renderer. The model is **semi-implicit Euler in fixed point** with a `0.93` (`= $EE/256`)
damping factor applied at *both* integration stages — that constant is the drag that keeps
the sim stable. The car-state block lives at `$1BCxx`:

```
position  (16.16)  X $1BCD8   Y/height $1BCDC (clamped <= $3E8)   Z $1BCE0
angles    (16-bit) roll $1BCE4   yaw $1BCE6   pitch $1BCE8
velocity           $1BCEA / $1BCEC / $1BCEE
angular momentum   roll $1BCF0   pitch $1BCF2   yaw $1BCF4   (body frame)
```

Forces are summed in the **body frame**, rotated into the world, and integrated. The frame
pipeline is, in order: build the orientation matrix (`$61368`); compute the wheel contact
offsets (`$618CE`); **sample the track surface** under each wheel (`$5C1D0`); compute the
chassis contact-point heights (`$61B70`); rotate velocity into the body frame (`$6158C`)
and gravity into it (`$615E6`); run the **suspension** (`$61BCC`); and — only when grounded
(`$1BB72`) — the **drive/tyre/steer** block (`$620B8`, `$61012`, `$61618`, `$621F4`,
`$62138`, `$61B26`, `$61672`); then integrate force→velocity (`$61ADC`) and
velocity→position (`$61950`).

### 2. Frame transforms

`$61368` builds the 3×3 chassis orientation matrix at `$1C230` from the three Euler angles
(`$64D08` = sin, `$64D10` = cos, via a quarter-wave table at `$1CA42`): it seeds the slots
with sin/cos of each angle, cascades the `MULS/SWAP` products into the composite rotation,
and forms the cross terms. Everything is then shuttled through it: world velocity → body
(`$6158C`), body force → world (`$61618`), body torque → world angular rate (`$61672`), all
driven by a small index table at `$1EC46` that selects which matrix slot each term uses.
**Gravity** enters via `$615E6`: a fixed world-down vector (magnitude `$13D` = 317) is
re-expressed in the tilted body frame each tick, so it always pulls world-down whatever the
car's attitude.

### 3. The suspension — and the surface

The car has a **three-point suspension** (`$61BCC`). Each point computes a spring
compression

```
compression = trackSurfaceY - chassisContactY - restLength      (clamped to [-$300, $1400])
force       = compression + 1.078 * (compression - prevCompression)   ($6180E: spring + damper)
```

The chassis contact heights (`$1BC94/98/9C`) come from the car height tilted by sin(roll)
and sin(pitch) (`$61B70`). The track surface heights (`$1BCA4/A8/AC`) are the **direct
coupling to Part IV**: `$5C1D0` locates the car's section, computes where each wheel sits
across the rung strip, samples the four surrounding rung-corner heights with the *same*
`$5C0AA` builder that produces the rail heights in Part IV, and bilinearly interpolates them
(`$5C554`) — so the springs ride the exact decoded track surface, and a wheel running off
the lateral edge (`$1BB65`) is detected and tapered (`$5C808`).

The three spring forces combine (`$61F42`) into a **net lift** (`$1BD38` — the weighted
average `(frontAvg + rear) / 2`: the single rear point carries the same weight as the front
pair together), a **roll torque** (`3 * (left - right)`, clamped to `±$1000`) and a
**pitch torque** (front average − rear); the **on-ground flag** `$1BB7E` is set if any
spring is engaged. The net lift is projected onto the body axes by the road slope
(`$622DC`) into the tyre **loads** `$1BD40/42/44`.

**Damage.** When a spring force exceeds a tolerance (`$1BB01 << 8 + $700`) an impact counter
(`$1BB56`, shared by the three points, reset the moment the force drops back below) starts
counting, and damage accumulates into `$1BB4F/50/51` (0–255) on each of the **first
`mem[$63CE2]` ticks** of the impact — after that window a continuous impact adds nothing
more, so the damage from one slam is bounded. The window is **per-track data**: `$63CE2` is
written at race setup (`$63AD2`) from the track trailer byte `$1CA2A` (or `$1CA2B` when the
mode flag `$1C9D0` is set) — 3 ticks on Roller Coaster, 4 on Draw Bridge, up to 10 on Big
Ramp. A separate bottoming counter (`$1BB7D`) increments when a spring force jumps to
`≥ $400` from below `$200` — the "land too hard and you wreck the car" mechanic,
reproduced exactly.

### 4. Drive, grip and steering

The longitudinal drive force (`$620B8`) comes from the throttle/gear, decays on wheelspin,
and is clamped to the available **grip** (`$621DA`): grip = tyre-load × 2 *only when
grounded* — zero in the air, and load-sensitive on the ground. The lateral tyre force
(`$6217A`) is likewise grip-limited, setting a slide flag (`$1BBC1`) when the demand
exceeds grip — a friction-circle model. A velocity **drag** (`$621F4`) opposes motion:
stiff when grounded and gripping, speed-proportional over a deadzone when rolling free.

`$61012` is the **track-following auto-steer**: it measures the car's heading error against
the section's centreline and nudges the heading (`$1BCE6`) to follow the track, then
computes a pitch-stabilisation torque (`$1BCFE`).

### 5. A copy-protection trap inside the physics

`$61012` carries a piece of the disk **copy protection**. At `$611E8` it compares
`mem[$64AEC]` against a constant and, on mismatch, **zeroes the pitch-stabilisation
torque** — subtly degrading the handling. Both operands are obfuscated specifically so they
can't be grepped: the address is built as `$79360 - $14874`, the magic value as
`$667B379F + $36729563 = $9CEDCD02`. The value at `$64AEC` is not a settings byte — it is
written (at `$5CEFA`) by a routine (`$5CEB4`) that reads the **physical disk hardware**
(CIA-B `$BFD100` drive control, `DMACONR`/`DMACON` disk DMA), i.e. it is derived from the
disk's protection tracks. So on a genuine disk the protection writes `$9CEDCD02` and the
check passes; the extracted blob holds the unpatched `$2F76EA80`, so it fails. The
reimplementation reproduces the check exactly (matching the oracle on the same image); for
a playable sim, "inserting the genuine disk" is just storing the expected value at `$64AEC`.

### 6. Verification

`cmd/physverify` runs each engine routine and its Go twin from identical random snapshots
and compares the state — `Sin/Cos`, the integrator, the matrix and transforms, the
suspension and damage, the drive/tyre/drag, the load projection, the per-section setup, the
**surface sample over loaded tracks**, the auto-steer (both protection branches), and
finally the whole `$6185C` frame in lockstep for 60 frames. **All match exactly.** The
checks pin several subtle behaviours of the original code: a swapped sin/cos
quarter-mirror, a signed `CMPI.b #5` whose `NEG.b $80` stays `$80`, an unsigned shift on
a negative handle that affects only curve pieces, and a branch where a large heading
error uses the *excess* as the correction.

**Part V is complete: the full rigid-body car physics is decoded from the disk in Go and
verified coordinate-exact against the original, frame by frame.** Combined with Part IV, a
car can now be driven over the decoded tracks.

### 7. The render coupling — placing the car on the track

The `$6185C` frame is exact, but it is only half of what makes the car drive. The other
half lives in the engine's **render pass**: each frame, before the physics runs, the game
recomputes *where the car sits on the track* — the section it is over, the surface height
under each wheel, and the heading the orientation matrix is measured against. The physics
reads that coupling state (`$1BC5E`, `$1BB10`, `$1BD5A`, the section `$1BB1C`) and the
suspension samples the surface there. Run the physics *without* the coupling and the car
floats in a mismatched frame and tumbles — which is exactly what a standalone port does.

Mapping the coupling turned up several concrete facts:

- **The per-frame loop is `$5D402` → `$5D608` (a `JMP` back to the top).** Its body runs the
  coupling, the player controls, and **two** `$6185C` ticks (player car and opponent), then
  the lap timer and the 3-D draw. Two physics ticks per displayed frame — consistent with the
  fixed 50 Hz (PAL VBlank) timestep noted earlier.
- **Placement uses a local coordinate frame, not world coordinates.** `$605B6` seats the car
  at `grid×128 + 64` in X/Z, runs the coupling (`$60190` camera/projection, `$5BE44`, `$60246`)
  and one physics tick, then *forces* `posY` to **`$10` = 16.0**. So the car rests at a height
  of ~16 in a frame whose surface sits near zero — it does **not** live at the world height
  (~1000) the standalone sim drifts to. That mismatch was the "flies up at the start" symptom.
- **The controls are a decoded joystick byte at `$1BB47`** (`$5D8A2`): bits 0–1 throttle/brake
  (up/down), bits 2–3 steer (left/right), bit 4 fire — not raw hardware, so a port can inject
  it directly.

Getting this to *execute* on the m68k oracle (so it can be ground truth, per the
[reimplement rule](#)) needed two additions to the shared core, both independently useful and
regression-clean against `physverify`:

- **`ORI`/`ANDI`/`EORI` to `CCR`/`SR`** — the render hit `ORI #$1,CCR` and the core had been
  treating the special `$3C`/`$7C` effective-address field as an immediate destination.
- **`ABCD`/`SBCD`** — the lap timer keeps its score in packed BCD.

With those in place the **coupling routines run to completion on the oracle**, and running
them around a placed car reproduces a grounded, upright, stable equilibrium — the physics no
longer tumbles. What does **not** run in software is the *drawing*: the per-frame 3-D render
(`$66xxx`–`$69xxx`) walks the whole visible scene and, without the Amiga's blitter/copper and
a vertical-blank interrupt, never terminates. So the faithful path is **not** to run the whole
loop, but to reimplement the handful of coupling routines (`$5BE44`, `$60190`, the section
update, the `$1BB47` control decode) in Go — each verifiable in isolation against the now-working
oracle — and compose them around the verified `$6185C` frame.

That is what shipped: the camera-follow (`$60190`/`$600A6`), the grid→section lookup
(`$5FE04`), the plan base and point (`$5FF94`/`$5C3DA`) and the placement `$5BE44` are all
reimplemented in Go (verified per-routine against the engine in `physverify`) and hand-ported
to the browser physics (cross-checked exact against the Go over the same memory blobs). The
placement covers **all three piece branches**, including the ramp-type-2 (curved-ramp) branch
`$5BF50`: the plan point (`$5C6C4` at header offset 2) is the car relative to the **curve
centre**; a quarter-table **atan2** (`$64D66`, table `$1CC46`) and **hypot** (`$64DE8`, table
`$1DC46` — indexed by the ratio quotient the atan2 leaves in `d7`) give its polar angle and
radius; the heading `$1BC1C` and orientation reference `$1BD5A` come from the angle plus the
camera quadrant; the along-progress `$1BB10` is the angle swept from the arc start (header
word 6) times an arc coefficient (header byte 8), scaled by a shift exponent in the ramp-flag
low bits; past the piece's end it steps to the neighbouring section and — if that piece's
type nibble is 4 — **re-enters the whole coupling for it** (`$5C072`); finally the radius is
refined (`$5C65A`: `r += ((x²+z²−r²)>>8) × coef[$5C6B8]`, a Newton-style pull toward the true
distance) and `$1BC5E` = piece radius (header word 9) − car radius, sign-flipped by `$1BC44`.
`physverify` checks every section of tracks 1/3/7 under random camera state (2048 ramp-2
cases, including the recursion, all exact).

### 8. The drive, grounded — the launch artefact explained and fixed

The earlier "vertical launch" was never a physics bug. It came from running the exact
`$6185C` from a *bare* placement (`posY` ≈ 1008) without the real per-frame coupling: the
surface sample then found no valid section, the springs read a mismatched frame, and the car
either fell through or shot up. Running the **real** sequence on the oracle settles the
question. The race loop (`$5D402`…) places with `$605B6` — which forces **`posY = 16.0`** in a
local frame whose surface sits near zero — runs one unarmed tick and the render coupling, then
arms `$1BB72` and, each iteration, does **physics then coupling** (`$6185C` at `$5D48A`, the
render `$64E4C` at `$5D496`). Driven that way the car settles and stays grounded: `posY`
oscillates 16 → 14.6 → 14.8, vertical velocity damps to zero, `onGround` stays set. No launch.

The render coupling that matters for the physics is a small, decomposable part of `$64E4C`:
`$60190` camera, zero the view offsets `$1BBD5`/`$1BBD6`, `$5FE04` grid→section into `$1BB85`,
`$5BE44` placement (the rest of `$64E4C` is the scene draw and the opponent car). That subset
is **byte-identical** to the full `$64E4C` for the car-state it produces, so the drive can be
composed in Go/JS from routines already verified individually.

One spawn detail matters: `$605B6` sets a crash-recovery countdown `$1BBDF = $F0`, and while it
counts down the `$5B32E` crash state-machine runs (it is a no-op only once `$1BBDF` reaches 0).
That machine is not reimplemented, so the faithful drive **zeroes `$1BBDF`** to drop straight
into normal driving. With that, `cmd/driveverify` locksteps the Go package drive (coupling +
`$6185C`) against the engine and matches **exactly**: 200 idle frames and 150 throttle frames
on tracks 1/3/7, car grounded throughout. The browser physics (`physics.js`
`driveTickCoupled`) is the same sequence and matches the engine over a 150-frame golden trace.
Reverse into a spin-off still diverges when `$5B32E` re-engages — the one open path.

**The world↔GLB mapping is solved.** The sim's persistent world position `$1BCD8`/`$1BCE0` is
in the same 16×16 track grid as the baked GLB, only at a different scale: `$605B6` seats the car
at grid×128 (128 units/cell, 16.16 fixed), while the GLB bakes grid×`$800`/1024 = 2 units/cell.
So **GLB = (physics 16.16 position) / (65536·64)**. Overlaying a driven path on the track
top-down confirms it: the car's mapped start sits exactly on the road, and `cmd/physdump` now
emits the real placed state per track so a browser `Physics` starts grounded and on-road.

### 9. The vehicle-control layer — the car drives

The exact physics alone is not a drivable car; the race loop runs a control layer on top of it,
now reimplemented and verified. Three routines matter (found by running the real control stack
on the oracle and pruning to the minimal set that laps):

- **`$5D8A2`** — the per-frame input decode (entered at `$5D8A8`, past the `$60BAE` hardware/
  keyboard-matrix read a port bypasses by supplying `$1BB47` directly). It turns the joystick
  byte `$1BB47` (bits 0-1 throttle/brake, 2-3 steer, 4 fire) into the steering demand `$1BBC6`,
  the fire flag `$1BB70`, and the longitudinal drive-force word `$1BD2A`/`$1BD2B` — the force
  comes from the **per-car acceleration constants** `$1BAFA`/`$1BAFB` (copied at setup by
  `$5D73A` from the car table), gated on airborne/stall/crash, with a wheelspin force-doubling
  (`$608A4`) when fire is held. (Injecting `$1BD2A` directly instead is why the earlier "drive"
  crept: it read a zero force.)
- **`$5B32E`** — the spawn / crash-recovery state-machine, run in the `$61BCC` tail each frame,
  a no-op once `$1BBDF` reaches 0. `$1BBDF` is a phase clock from `$F0`: prime the pitch
  accumulator (`$F0..$E6`), a pitch + heading servo (`$E5`), wait for the pitch to settle then
  reload the clock and arm the launch (`$E4`; the reload is a deterministic `$8C` because the
  race has `$1CA22 ≥ 0`, so the `$62574` PRNG branch is never taken), and settle-and-hand-off
  (`$E3..$01`). It servos the pitch angle `$1BCE8` and a heading accumulator, injecting no
  velocity — the car rolls forward once the physics is handed control. This is what makes the
  car launch cleanly instead of flailing.
- **`$5DB34`** — the lap timer, of which only the `$EE` time-base tick `$1BBCD` matters to the
  physics (it gates the crash countdown and the wheelspin timer). The rest — the start-light
  sequence — is visual; a minimal tick reproduces the physics bit-for-bit.

The start section is `$605B6`'s `d1 = $1CA1B` (the header finish/start index), with the intro
setup `$1BB1D`/`$1BB0C`/`$1BBED`; the car seats at the start line, `posY = 16.0`.

**Verification.** Each routine is checked against the engine in `cmd/physverify` (3000 random
states, exact). `cmd/driveverify -full` locksteps the whole real stack — input decode, physics
with the `$5B32E` spawn active, coupling, timer — against the engine frame by frame with an
injected throttle: **bit-exact through the entire ~240-frame spawn and into free driving**, 300+
frames per track (all 300 on Roller Coaster). Finding this fixed a real omission the earlier
checks missed — `Integrate61950` skipped the always-run `$619E4` tail (`$1BC42 = -roll`, read by
`$622DC`). The browser physics (`physics.js`) ports the same routines and matches the Go golden
trace exactly for 300 frames on all eight tracks (`cmd/physdump` emits the placed state + the
full-drive trace). The far tail of the drive on some tracks still diverges from the engine
(a start-light-state feedback the minimal timer omits), but the spawn and the bulk of the drive
are exact.

**The car drives in the Studio.** `site/src/stunt-car-racer-amiga/model-renderer.js` loads the
per-track placed state, runs `driveTickCoupled` from a joystick byte each frame, maps the sim
world position onto the GLB by `/(65536·64)`, and keeps the car on the road with a downward
raycast onto the track mesh. The car spawns at the start line and drives the circuit; the game's
own auto-steer follows the straights, and the player steers through the corners. Controls are
**`IJKL`** (I throttle, K brake, J/L steer, O fire) so they don't collide with the free-flight
**fly-cam** (`WASD`/arrows) — the camera stays under your control while the car drives on its
own; it does not chase. The **Low res** display toggle (next to Wireframe / Screen filter)
turns the native-Amiga 200-line render off for judging exactness at full display resolution.
(Forcing the steering bias `$1BBC6` every frame — even to
0 — fights the auto-steer and throws the car airborne; it is left alone unless a steer key is
held.)

---

# Part VI — The 3-D objects: the opponent car and the horizon

The race view draws two things besides the track: the computer player's car,
and the mountain range on the horizon. The car is **not a stored vertex
model**. There is no vertex/face list anywhere
on the disk: `$599E2` assembles the car procedurally, in *screen space*, from
the four projected wheel points of the opponent's rig, and `$67AA6` fills its
faces with colours hard-coded in an unrolled paint list.

## 1. Placement: wheels ride the decoded track

The opponent runs no physics. Its state is the AI navigator's — section
(`$1BB1D`), distance along it (`$1BB0C`), across position (`$1BB13`) — advanced
by `$63996` each frame, and `$5A186`/`$5A40C` place its four wheels directly on
the Part IV track surface: the **front axle** is the across-interpolation of
the rail sample grid (the car is 81 plan units wide on the standard 384-unit
road — the `$1BBBE` across offsets), and the **rear axle** is the front pair
displaced by **1.5× the axle vector rotated 90°** (`$5A468`) — wheelbase 122.
Heights stack simply (`$63A58`, `$641B6`): body plane = the bilinear surface
sample + `$68` + a per-frame bounce (`rand & $7F`, the fake suspension); wheel
tops = body + `$50`; deck plane = surface sample + `$50`.

## 2. The screen-space build (`$599E2`)

`$5A0CE` takes an axle's two projected wheel points and derives a family of
**signed halves, quarters and eighths** of the screen axle vector and its
perpendicular (`$59B7A`), plus the axle midpoint. From those fractions alone:

- `$59C7C` builds the **front face rectangle** and, via `$59F78`/`$59F98`, the
  two front **tyre parallelograms**;
- `$59DF6` builds the **rear face rectangle**, the four **connector edges**
  running the length of the car, and the rear tyres;
- `$59B1A` closes the **deck quad** over the four wheel-bottom points, which
  `$599E2` first slides by the axle fractions (`$59AB8`/`$59AEE` — the front
  shift uses `bc6A−1`, the rear plain `bc6A`) so the deck lands under the car
  as its painted shadow/base plate.

Everything goes into a reserved display-list region at `$5E0` (32 edges), and
`$67AA6` — dispatched from the flush when it reaches the car's strip
(`$66FF0`/`$1BC12`, painter's order) — fills ten faces with fixed colours:
**front rectangle `$A` (red), top `$F` (white), bottom `9` (dark red), flanks
`$C` (light grey), deck `5` (mid grey), tyres `0`** — the background colour:
the wheels are literally holes filled with sky.

## 3. Reimplementation and proof

`extract/carmodel` ports the construction verbatim (`carmodel.Build`, the
`$1BC64…$1BC8A` fraction family included), and `cmd/caroracle` proves it: it
runs the real race init, places the opponent exactly as `$5D3C6` does, runs the
game's own per-frame chain (`$6076C` distance/visibility, `$63A58`, `$641B6`,
`$5A186`) and the real `$67AA6` draw on the m68k core, and compares every one
of the 32 emitted display-list edges against `carmodel.Build` —
**coordinate-exact on all eight tracks**.

Because the construction is linear in the projected wheel points, running the
verified port through two *orthographic* views (front and side) recovers the
exact 3-D shape the arithmetic encodes: `carmodel.Rest` does that, and
`cmd/webexport` ships it as `models/opponent-car.glb` — the red-fronted,
white-topped wedge on four background-coloured tyre panels over its grey deck
plate, in the same scale and palette conventions as the circuit GLBs. (The
original being a screen-space build, the GLB is by nature its rest-pose
interpretation; the wedge proportions, planes and colours are the engine's own
numbers.)

## 4. The horizon mountain range (`$6953E`)

The mountains are true 2-D shapes, drawn by a small yaw-placed object renderer.
The race-init config step `$696FC` loads a placement list through the pointer
table at `$69A80` (it forces entry 0, so **every track shows the same range**):
a count byte then `(yaw, model)` pairs into `$69736`/`$69766` — **32 objects
spread around the full 256-unit compass** at alternating `$A`/`$6`-unit
spacings. Each frame `$6953E` walks the list and draws every object whose
heading falls in the view window, at

```
x_left   = (yaw·256 − cameraYaw16) >> 3     ; 1 yaw unit = 32 horizon pixels
y_screen = −($1BC42 >> 3) − y               ; base pinned to the horizon line
```

then pushes every vertex through the same `$62518` roll-rotate/`>>2`/re-centre
transform as the car. A model (table `$699B8`, 8 bytes per entry: shape +
variable-stream pointers) is a word vertex count, `(x, y)` word pairs where a
**negative word is a placeholder filled from the entry's variable stream**,
an edge list and a face list `{colour, vertex count, edge refs}`. That
placeholder trick is the whole range: **all 14 placed models share one
4-vertex silhouette** — base line plus two peaks — with per-entry streams of
5 words giving each mountain its width (250–384 px) and peak heights
(12–39 px), filled as a single quad in **palette 5** (`$555` mid grey). A
second two-triangle shape (table entries 14/15, palettes 4/5) exists but the
config never places it.

`track.Horizon` decodes the placement list and models purely from the image,
and `cmd/horizonoracle` verifies it against the real `$6953E` on the m68k core
across eight camera headings — **every emitted edge coordinate-exact, face
colours included**. `cmd/webexport` ships the range as `models/horizon.glb`:
the 32 silhouettes arranged as the 360° ring the renderer implements (arc
length = the engine's 32 px per yaw unit, radius `8192/2π` px, bases on the
Y=0 horizon plane).

---

# Appendix A — Toolchain and reproduction

All work is reproducible from the image with the shared `tools/` module:

```sh
# Inspect the boot block / a raw region (disk offset maps 1:1 to bytes)
go run retroreverse.com/tools/cmd/dis68k -base 0 -skip 12 "Stunt Car Racer (Amiga)/Stunt Car Racer.adf"

# Slice the disk into loader.bin, game.bin and the decrypted game.dec.bin
cd "Stunt Car Racer (Amiga)/extract" && go run ./cmd/extract "../Stunt Car Racer.adf"

# Dump the eight track byte streams (reimplements the $5AE46 pointer math)
go run ./cmd/tracks ../extracted/game.dec.bin    # -> extracted/tracks/<id>_<name>.bin

# Decode + verify the section streams (byte-exact against each track's length)
go run ./cmd/sections ../extracted/game.dec.bin [-v]

# Run the real $5AE46 on the m68k core and read out the spine (X,Z) per section
go run ./cmd/spineoracle ../extracted/game.dec.bin [trackid]   # -> extracted/spine_<id>.csv

# Verify the pure-Go spine (package track) against the oracle, coordinate-exact
go run ./cmd/spineverify ../extracted/game.dec.bin

# Verify the Go reimplementation of the track-model bake (track.Bake) against the
# engine's real $65BEC on the m68k core — every baked record, all 8 tracks, byte-exact
go run ./cmd/modeloracle ../extracted/game.dec.bin [trackid]

# Render the absolute top-down plans (the baked model placed at its grid cells) as SVG
# and report section-joint continuity
go run ./cmd/modelsvg -out <dir> ../extracted/game.dec.bin

# Export the eight decoded circuits to JSON for the web track viewer
go run ./cmd/trackjson ../extracted/game.dec.bin   # -> site/public/stuntcar/tracks.json

# Verify the Go car physics (package physics) against the engine on the m68k core:
# every sub-routine + the full $6185C frame in lockstep for 60 frames (all exact)
go run ./cmd/physverify ../extracted/game.dec.bin

# Verify the preview colour/decal reimplementation (track.Mesh) against the real
# $604B4 preview draw on the m68k core — 4 camera angles, every colour decision
# hooked; all 8 tracks exact, palette word-exact (Part IV §9)
go run ./cmd/coloracle -in ../extracted/game.dec.bin [-id N] [-dump]

# Verify the opponent-car construction port (carmodel.Build) against the real
# $599E2/$67AA6 draw — all 32 display-list edges coordinate-exact (Part VI)
go run ./cmd/caroracle -in ../extracted/game.dec.bin [-id N]

# Verify the horizon mountain-range decode (track.Horizon) against the real
# $6953E renderer over 8 camera headings — edge- and colour-exact (Part VI §4)
go run ./cmd/horizonoracle -in ../extracted/game.dec.bin [-id N]

# Verify the Draw Bridge animator reimplementation (track.Drawbridge) against
# the real $5A794 over 64 frames — table byte-exact + freeze gate (Part IV §10)
go run ./cmd/bridgeoracle -in ../extracted/game.dec.bin

# Export the web assets: per-circuit track JSON + the GLB models (8 circuits in
# the preview's exact colours/decals + models/opponent-car.glb)
go run ./cmd/webexport [-in ../extracted/game.dec.bin] [-only tracks,models,all]

# Disassemble / trace the engine. Use game.dec.bin for anything in $F4B8..$1AA4A.
go run retroreverse.com/tools/cmd/dis68k     -base 0xE700 -start <addr> -end <addr> extracted/game.dec.bin
go run retroreverse.com/tools/cmd/codetrace68k -base 0xE700 -entry <addr>            extracted/game.dec.bin
```

Dynamic verification uses the instruction-level 68000 core in `tools/cpu/m68k`
(`m68k.CPU` over a `Bus`).

The disk image is not committed (it is a copyrighted game); its size and MD5
are recorded in the repository [README](../README.md#image-files) so the exact
copy can be verified.
