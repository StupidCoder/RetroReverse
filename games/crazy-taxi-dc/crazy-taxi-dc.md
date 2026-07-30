# Crazy Taxi (Sega Dreamcast)

**Image:** `Crazy Taxi (US).bin` + `Crazy Taxi (US).cue` — 1,188,056,352 bytes, MD5 `e446e7a09dc23985359c7cb9d1034639`.
Not committed (copyright); supply your own copy.

Sega's 1999 open-city taxi game, on the repo's Dreamcast platform: an SH-4
(SH7091) toolchain in `tools/cpu/sh4` (`dissh4`, `codetracesh4`, execution
core), the machine in `tools/platform/dc`, the static disc inspector
`tools/platform/dc/cmd/gdinfo`, and the oracle in
`extract/cmd/bootoracle`.

## Part I — The disc

### The container

The rip is a `.cue` in **cdrdao TOC syntax** (not CDRWIN cue) over a single
`.bin` of raw 2352-byte sectors:

```
TRACK MODE1_RAW   DATAFILE "Crazy Taxi (US).bin" 00:08:00
TRACK AUDIO       DATAFILE "Crazy Taxi (US).bin" #1058400 09:52:00
TRACK MODE1_RAW   DATAFILE "Crazy Taxi (US).bin" #2295552 112:02:00
```

`#N` is a starting byte offset in the `.bin`; the trailing MSF is the track's
*length*, and it lies: track 2's declared 104,428,800 bytes overrun track 3's
offset by two orders of magnitude — the audio track is truncated in the dump.
The reader (`tools/platform/dc/cue.go`, `disc.go`) therefore trusts only
modes and `#` offsets; lengths derive from the next track's offset and the
file size.

### The GD-ROM layout

A GD-ROM has two sessions: a low-density CD area (tracks 1-2 here) and the
high-density data area, which on every retail disc begins at absolute LBA
45000. The anchor is not taken from the TOC: each raw data sector carries its
own address as BCD minutes:seconds:frames in the 16-byte header, offset by
the 150-frame pregap, and track 3's first sector says 10:02:00 = LBA 45000.

| track | mode | file offset | bytes | absolute LBA |
|---|---|---|---|---|
| 1 | MODE1_RAW | 0 | 1,058,400 | 0–449 |
| 2 | AUDIO | 1,058,400 | 1,237,152 | — |
| 3 | MODE1_RAW | 2,295,552 | 1,185,760,800 | 45000–549149 |

### IP.BIN

The high-density area's first 16 sectors are the initial program. Sector 0
opens with the boot metadata, fixed-width space-padded text:

| field | value |
|---|---|
| hardware | `SEGA SEGAKATANA` |
| maker | `SEGA ENTERPRISES` |
| device | `EDB5 GD-ROM1/1` |
| area | ` U` |
| peripherals | `0799A10` |
| product | `MK-51035` `V1.004` |
| release | `19991219` |
| boot file | `1ST_READ.BIN` |
| title | `CRAZY TAXI` |

### The filesystem

The data area is ISO 9660 with the multi-session convention: the PVD sits at
session+16 (LBA 45016, volume `CRAZY_TAXI`), its volume-size field is
session-relative (504,150 blocks), but every directory extent is *absolute* —
the root lands at 45020. The disc therefore serves absolute LBAs as an
`iso9660.BlockSource` and the volume opens through `iso9660.OpenVolumeAt`.

51 files, flat in the root. The census in brief: `1ST_READ.BIN` (1,468,208
bytes, LBA 548283 — outer-edge placement, where the drive reads fastest),
`AICADRV.BIN` (the 57 KB sound-CPU driver), three of each per-city set —
`BINC*.AFS` archives, `COLDC*.BIN` collision, `LANDDC*.AFS`, `POLDC*.BIN`
geometry, `TEXDC*.BIN`/`TEXPVR*.BIN` textures — plus `MOTDC.BIN` motion
data, `SEQON*.BIN`, `VOICE.BIN`, `GDTEX.PVR` (the disc-icon texture), and
`CT2/` sequel-preview assets. `gdinfo -files` lists them all; `-at LBA`
names the file any raw read lands in.

## Part II — Boot chain and toolchain

### The bootstrap

`1ST_READ.BIN` loads at 8C010000 and its first instructions are a copy loop
(`dissh4 -base 8C010000`):

```
8C010000  D006  mov.l @(0x18,pc), r0      ; = AC010100
8C010002  D107  mov.l @(0x1C,pc), r1      ; = AC014000
8C010004  D207  mov.l @(0x1C,pc), r2      ; = AC004000
8C010006  6302  mov.l @r0, r3             ; copy loop
8C010008  2232  mov.l r3, @r2
8C01000A  7004  add #4, r0
8C01000C  7204  add #4, r2
8C01000E  3100  cmp/eq r0, r1
8C010010  8BF9  bf 0x8C010006
8C010012  D004  mov.l @(0x10,pc), r0      ; = AC004000
8C010014  402B  jmp @r0
8C010016  0009  nop
```

The loader relocates its second stage from 8C010100 down to 8C004000 through
the uncached P2 mirror and jumps there; the second stage then loads the main
program high (0C14xxxx and up), zeroes the vacated pages, and enters the
game. The literal pools between these functions are why the tracer treats
`mov.l @(disp,PC)` targets as first-class data (`codetracesh4` renders them
as `.word` and refuses to decode into them).

### What the machine model provides at boot

No BIOS image is used. `dc.Boot` reproduces the hand-over state the BIOS
leaves: the boot binary at 8C010000, SR `400000F0`, FPSCR `00040001`, stack
at 8C00F400, and in low RAM exactly six planted facts —

- the five syscall vectors (8C0000B0 sysinfo, B4 romfont, B8 flashrom,
  BC gdrom/misc, E0 menu), each a pointer to a trap PC the HLE services;
- the console ID at 8C000068;
- `rte; nop` at 8C000010: the BIOS's uncached return-from-exception stub.
  The game's own interrupt epilogue is the evidence — it restores the whole
  context including SSR/SPC, hand-builds the constant AC000010 (P2, so the
  `rte` executes uncached) and jumps there with nothing left to do but
  return;
- zeros at 8C00FFF0/FFF4: the game keeps a lazily-initialised work-structure
  pointer pair there and its "already initialised?" test is `tst` — a fresh
  boot must read zero.

Every other byte of the BIOS work area is poisoned `F00D` + offset, so a
dependency on an unplanted value surfaces as a self-naming register value,
never as an invented fact.

### The boot, observed

The oracle runs the retail boot for a billion instructions
(`bootoracle -steps 1000000000 -gd -v`):

1. the bootstrap relocates and the runtime initialises;
2. a serial driver probes the expansion socket at 03010000 (an empty G2
   socket reads all-ones and the probe gives up) and prints over the SCIF —
   the on-chip UART's transmit FIFO always drains, and every byte lands in
   the oracle's `serial` capture;
3. the runtime installs VBR at 8C00F400, hooks the VBlank through Holly's
   level-6 mask (INTEVT 0x320), and services fields at 60 Hz;
4. GD-ROM syscalls read the session lead-in and then stream `AICADRV.BIN`
   (28 sectors) into sound RAM; the Maple bus enumerates the port-A
   controller and block-reads for memory cards (answered "function
   unsupported"; the VMU slots answer as empty sockets);
5. 4,568 words of geometry go to the tile accelerator by ch2 DMA (counted,
   not rendered — the PVR is a later milestone);
6. the boot parks polling a handshake word in sound RAM at A080005C: it is
   waiting for the AICA's ARM7, which does not run yet. That poll is the
   frontier of this milestone, named by the machine itself.

Savestates round-trip the whole machine (a run continued past a snapshot and
a run restored from it stay byte-identical), and the unmodelled-hardware
census after a run is the worklist: the AICA block, the GD DMA registers,
the RTC words at 00710000, the PVR beyond its register file.

### The SH-4 toolchain

`tools/cpu/sh4` follows the repo's CPU-package shape: `decode.go` dispatches
the 16-bit halfwords the way the manual's tables group them (`decode_fpu.go`
holds group 1111), `exec.go`/`exec_fpu.go` interpret under the r4300-style
delay-slot pipeline with the SH-4 rules — interrupts are never accepted
between a delayed transfer and its slot, delayed conditionals run their slot
on both paths, R0-R7 swap banks on SR.RB writes, the FPU is two banks of raw
bits reinterpreted by FPSCR. The TMU, INTC, store queues and SCIF live in
the CPU package because P4 accesses never reach the external bus. Validation
is the assembled-loop suite in `cpu_test.go` plus the golden decode table
seeded with this game's own bootstrap; an env-gated per-instruction
conformance harness (`SH4_SST_DIR`) is deliberately deferred until the suite
files are in hand to pin the format against.

## Part III — The sound processor and the frame loop

The boot's declared frontier was a poll of sound RAM at `A080005C`: the SH-4
waiting for a handshake from the AICA's own ARM7, which did not exist yet.
The milestone that followed gave the machine its second processor — the
shared `tools/cpu/arm` core, paced at one step per nine SH-4 instructions,
running `AICADRV.BIN` from reset vector 0 in sound RAM — and the sound
block's timers, counted in 44.1 kHz samples, feeding the SCIPD/MCIPD
pending registers. When the driver answered the handshake, the game's
mainline came alive: a 60 Hz field loop against a real line-sweeping SPG
(lines 0..524, VBlank where the game's own `SPG_VBLANK_INT` asked for it).

Two disciplines from earlier platforms were applied preemptively rather
than re-learned:

- **completions take time** — render and DMA interrupts are counted down,
  never raised inside the store that starts the operation (the GameCube's
  instant-I/O lesson);
- **a device must not vanish between commands** — Maple commands the pad
  does not speak answer "function unsupported" (0xFE), never the
  empty-socket word.

## Part IV — The display gate, the software PVR, and the first picture

The game ran for billions of instructions with a configured-but-disabled
display: framebuffer registers set, `FB_R_CTRL` enable clear, `VO_CONTROL`
blanked, and a render-command queue that filled and timed out every frame.
The gate was the tile accelerator's *list discipline*, recovered from the
game's own bookkeeping (an expected-mask word at `[0C2E7E88]` against an
accumulator at `[0C2E7E8C]`):

1. only polygon-path (`10xxxxxx`) submissions carry TA parameters — a
   texture-path job is pixels, and scanning pixels as parameters
   manufactures completions the game never made;
2. parameters have real sizes — 64-byte vertices exist, and a header can
   owe a second 32-byte chunk;
3. lists are strictly ordered and close *implicitly*: `TA_LIST_INIT`
   leaves the opaque list open, and a header naming a different list fires
   the open list's completion.

With the completion count exact, the game wrote `STARTRENDER`, flipped its
double buffer, set the enable bit, and unblanked. The renderer behind
`STARTRENDER` is `pvr.go`: a software rasteriser over the *recorded* TA
stream — recorded per session and double-buffered like the hardware's own
(`TA_LIST_INIT` stashes the recording; the render draws the closed session,
not the recording one). The first picture it produced was the no-VMU
warning screen, in Japanese — the flash is erased, so the console has no
language setting, and the game falls back to its first language.

## Part V — The Start button, the second submission path, and the city

### An input is consumed as edges

The warning screen says "press Start" and ignored a Start held for three
thousand fields. The chain, traced end to end: the game's VBlank ISR
triggers a Maple DMA every field; the response parser (`0C14C548`) fills a
52-byte record per port (port A at `0C2B6248`) with current buttons
(active-high, +8) and *press edges* — current AND NOT previous — at +0x10,
where "previous" is one field ago; the mainline copies driver records into
game records (`0C029AA0` → port A at `0C1EBF6C`) once per loop; and the
screen's handler advances on the *edge* word. An edge lives exactly one
field. The mainline usually runs one loop per field, but occasionally
overruns — and an overrun that swallows the edge's field loses the press
forever, no matter how long the button stays down. Real hardware never
overruns this screen; the model sometimes does, an accepted timing-fidelity
gap. The oracle's `-keys` scripts therefore *tap* — three-field holds,
repeated — and the screen advances: the ADX/CRI middleware splash, then the
attract sequence.

Two dead ends the trace also named: the record's identity pointer reads
"(no device)" (the game classifies controllers by product-name bytes,
'R'→wheel, 'F' at +10→fishing rod; the fallback is the standard-pad path,
so the label is cosmetic), and a second consumer of the current-buttons
word turned out to be the A+B+X+Y+Start soft-reset combo checker, not the
screen.

### The FIFO is half the hardware

The attract scene submitted 23 million TA words and drew a fragmentary
skeleton. The census said nothing was unimplemented — so the missing
geometry was never *reaching* the rasteriser. It wasn't: direct stores into
the TA FIFO window (`0x10000000`, fed by the SH-4's store queues — the
mainline's registers hold `E000FAxx` addresses mid-frame) were counted and
discarded, a leftover from the boot milestone. The game submits its UI by
ch2 DMA and its *world* by store queue. One `taFeed` now consumes both
paths chunk by chunk, with parameter sizes carried across job boundaries,
and FIFO END_OF_LIST completions paced on their own countdown.

### Two windows, one memory

The last visible defect was a rectangle of coloured noise on the road: a
256×256 VQ texture whose control-word address held zeros when dumped, yet
sampled noise when drawn. The contradiction is the VRAM architecture: 8MB
as two 4MB banks, addressable through a 32-bit window (banks in sequence)
and a 64-bit window (banks interleaved a word at a time). The texture was
uploaded through one window and sampled through the other — under the
machine's old "both windows are linear" simplification, those are the same
bytes; on hardware they are not. The VRAM array now lives in the 64-bit
path's layout (texture control words address it directly), and the 32-bit
window — including the framebuffer pointers, which are 32-bit-path
addresses — goes through the interleave (`vram32to64`).

The rasteriser grew what the attract census named, feature by feature:
intensity colour modes (face colour from the header, scaled by a per-vertex
float), floating colour, 16-bit strip UVs, mipmap chain offsets for every
format, a 1/w depth buffer honouring the ISP compare mode, and
perspective-correct UVs — the PVR hands the rasteriser 1/w as its depth
value, so `u·z, v·z, z` are exactly the screen-linear quantities. The
warning-screen render stayed byte-identical through every change, pinned by
hash; and the attract mode now renders the city — storefronts, billboards,
lane markings, and the black convertible the camera follows down the hill.

### The pad's identity papers, and a state chain reborn

Two aftershocks followed the city. First: the attract's taxi rendered solid
black, and the hunt for its paint ran through five hypotheses before a fresh
boot under the current build showed the texture uploading perfectly — every
savestate in use descended from a boot made by an older build whose
texture path never wrote that region of VRAM. A savestate freezes not just
the guest but the machine model that produced its memory; content written
once at load fossilises, and only content regenerated per frame heals when
the model is fixed. The whole state chain was re-cut from a clean boot, and
the warning screen came back wearing the flame-swoosh CRAZY TAXI logo it
had silently been missing all along.

Second: on the re-cut chain the pad went dead — and the reason is the best
kind of lesson. The old boots' Maple response headers were malformed, so
the game's device enumeration had always *failed*, into a tolerant fallback
that accepted condition data from an unidentified device. The fresh boot's
correct headers made enumeration *succeed* — and success meant scrutiny:
the driver byte-reverses the DEVINFO function-type word and tests the
CONTROLLER bit every single field before it will read a button
(`0C14C80E`, `tst #33`), reads the device-status block in the hardware's
own layout (area code and connector direction ahead of the 30-byte product
name), and classifies the device by its name bytes — a leading 'R' selects
the racing-wheel input path, which our pad, then named "RetroReverse
Dreamcast Pad", would have taken. The machine now answers with the retail
pad's own 112-byte block, function word raw on the wire, named "Dreamcast
Controller". The reward for full protocol honesty was the attract sequence
in full colour: the yellow cab riding a green car-transporter past the
park, a purple pickup alongside, PRESS START BUTTON across the road.

## Part VI — The drive

With the pad speaking the hardware's own protocol, the rest of the road to
gameplay is the game's own menus, driven by `-keys` taps: Start at the
warning screen, Start at the attract's PRESS START BUTTON, then A through
MODE SELECTION (arcade highlighted first), the ARCADE rules menu, and the
driver select — where Axel stands as a fully shaded, textured 3D model
beside the line-art taxi, the other three drivers' portraits waiting above.
One more A and the machine is *in the game*: the yellow cab dropped onto
the boulevard under a blue sky, palm trees and traffic streaming past, the
game-time counter running, TOTAL EARNED's odometer wheels at zero, the D
gear lit in the corner.

The regression gate pins three states along that path — the warning screen,
the attract, and the drive itself — five hashes each (RAM, the scanout
picture, VRAM, the SH-4's trajectory, the AICA side), thirty fields from
each savestate, with a determinism test underneath. What remains for the
rasteriser is honesty at the margins: the ISP background plane (the
in-game sky is drawn geometry, so nothing visible depends on it yet), the
translucent auto-sort, and whatever the census names next. What remains
for the writeup is the game itself: the asset containers (`BINC*.AFS`),
the world and car geometry (`POLDC*`/`TEXDC*`), collision (`COLDC*`) and
motion (`MOTDC`) — the formats the next parts will crack.
