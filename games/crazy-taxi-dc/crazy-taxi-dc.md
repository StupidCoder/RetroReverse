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

## Part VII — The fare

Driving needed one more instrument: the `-keys` script learned the
joystick (`jx<0-255>`, 0x80 centered), because Crazy Taxi steers on the
analog stick and the d-pad does nothing on the boulevard. With gas on the
right trigger and a steerable cab, the oracle drove at the first waving
customer — and sat beside the orange dashed ring for thirty seconds of
game time while the customer waved and nobody boarded.

The suspect list ran from sound to geometry. The AICA census showed the
sound driver reading registers `2810` and `2814` — the envelope and
current-address monitors for whichever of the 64 voice slots `280C`
selects — tens of thousands of times a second, against registers nothing
had ever written, answering a constant zero. That is precisely the shape
of the Jak & Daxter lesson (a cutscene frozen because the SPU's playback
position never advanced), so the machine grew an honest voice-position
model: KYONEX/KYONB key-on across all slots, per-slot pitch stepping at
the 44.1 kHz sample tick, loop wrap with the LP flag, monitor reads that
report where the sample *would* be. The gate classified the change
exactly as the multi-hash design intends — the drive pin's AICA hash
moved, frame, RAM and CPU hashes did not.

It also wasn't the boarding gate. The real answer was embarrassingly
physical: the accept radius is much smaller than the drawn circle. Stop
*at the customer's feet* and everything happens at once — they sprint to
the door, the camera swings in, game time gains a bonus, the meter lights
at $144.40, a destination panel names "Cable car stop TOP", and a green
arrow rises over the cab. The passenger has opinions, too: drive off the
wrong way and the text over the cab reads 「反対方向だぞ！」— *you're
going the wrong way!* — and, after an accidental reverse-gear flight off
a ledge scored "Crazy Jump!", 「行きすぎたぞ！」— *you've passed it!*

The gearbox came out of the same empirical probing as everything else:
**A shifts to R, B shifts to D** — the two positions of the arcade
shifter, not a toggle — and X pops the destination preview. In-game the
mainline eats single-field button edges far more often than the menus
ever did; the discipline is to tap well after a savestate resume and read
the gear widget for confirmation, never to blind-repeat a shift.

The RAM dug up along the way now reads like a dashboard: game time in
frames at `0C1EC4B4`, the fare countdown at `0C1EC4EC`, the destination
name table at `0C08CC50` — and `0C13D89C`, which spent an hour
masquerading as a distance before revealing itself as the guide arrow's
relative bearing in hundredths of a degree, zero dead ahead. That scalar
navigated the cab to within metres of the cable-car stop; what ran out
was the fare timer, spent on U-turns judged one screenshot at a time. The
delivery — stopping inside the zone, the fare banked into TOTAL EARNED —
is the open milestone, and it wants a live cab-position instrument
first, so the steering can be computed instead of guessed.

## Part VIII — The sound

The voice-position model of Part VII knew where every sample *would* be;
this part makes the machine actually play it. The step in between was a
census, not a guess: the AICA register file only stores what a write has
touched, so dumping the stored map from the deepest state says exactly
which of each slot's 32 words the game's driver uses. The answer
(`bootoracle -aicaregs`): words `+00` through `+44` on all 64 slots, and
never anything above. Decoded against the slot layout, that is sample
address and format, loop points, the amplitude envelope, pitch, the DSP
send, the direct send with its pan, and the total level — while the LFO
word is *always zero* and the filter registers sit pinned at their neutral
values. The synthesis model (`aica_synth.go`) is exactly the census: no
LFO, no filter, and no DSP mix (the driver does load a reverb program and
point slots at it; the dry path carries the music, and the missing wet
path is census-logged rather than silently absent).

What was added, then: sample fetch in the three formats the header can
name (16-bit PCM, 8-bit PCM, and Yamaha's 4-bit ADPCM — a sequential
decoder, so the model walks it forward nibble by nibble and caches the
predictor state the first time it crosses the loop start, which is what
the hardware's loop does); a real amplitude envelope generator
(attack/decay/decay-2/release, effective-rate timing in Yamaha's
exponential-attack, linear-dB-decay shape) that replaces the old model's
instant key-off kill; and a stereo mix at the 44.1 kHz tick through each
voice's envelope, total level, direct-send level and pan, under the
master volume. The envelope is honest to the *protocol* — the `2810`
monitor now answers with the generator's real state and level, and a
one-shot still ends where the sample does — while its milliseconds are an
approximation nothing in the driver's handshake depends on.

The mix always runs, listener or not: decoder and envelope state are part
of the machine and savestate with it, and a run with `-wav` must be
byte-identical to a run without. `-wav` only copies each mixed pair into
an instrument buffer and writes a RIFF file at the end, with a
peak/RMS summary so a silent run states its silence in numbers.

The verification is the same shape as every first picture: from a cold
boot with only the two Start taps scripted, the machine produces
twenty-five seconds of honest silence through the boot and the warning
screen, the splash sting, and then the attract soundtrack — beat, bass
and vocal formants plainly structured in the spectrogram. Replaying the
Part VII boarding recipe captures the pickup: the customer's shout, the
passenger's lines, the effects over the music. The hash gate
classified the whole change the way the multi-hash design intends: the
drive pin's AICA hash moved (the envelope generator answers the monitor
differently, so the ARM driver's trajectory shifts), and frame, RAM,
VRAM and CPU hashes held still on all three pins.

## Part IX — The frame debugger

The Dreamcast joined the frame-debugger suite (`framedbg -serve`,
`tools/debug/dcadapter`): the library lists the game, a savestate lands it
at the drive, and one captured frame is 9,213 commands you can scrub
through, click, and interrogate — which command drew this pixel, which
pixels this command drew.

Every platform in the suite has had to answer "which buffer is the frame?"
in its own way, and the Dreamcast's answer is: **the draw target is one
flip ahead of the scanout**. The PVR rasterises a whole recorded TA session
into the buffer `FB_W_SOF1` names when `STARTRENDER`'s deferred completion
lands; the guest sees the render-done interrupt and only then flips
`FB_R_SOF1` onto it. So the capture's picture is the write framebuffer,
taken inside the completion — after the last pixel, before the guest can
flip — and the paused stage view (the scanout) honestly shows the frame
*before* the one just captured. A "command" is one TA parameter of the
frame's stream — a header, a vertex, an `END_OF_LIST` — plus the background
clear as command zero, which reports every pixel it writes: the clear is
usually the answer to "why is this pixel black", and provenance that says
"nothing wrote here" across the whole background would be a lie. The render
walk is a plain Go loop inside one deferred completion, so the scrubber's
mid-frame halt is a counter (`RenderStopAfter`), not a sentinel panic.

The instrument-purity fight this time was with the compiler, not the
machine. The first version reported fragments from inside the rasteriser's
pixel loop, nil-checked and dormant — and the drive gate's picture moved by
two pixels, each one 565 quantum off. The hooks weren't running; the *edit*
was: any textual change inside `tri()`'s loop shifts Go's FMA contraction
on arm64, and with it the barycentric interpolation's last bit (the same
trap `bench_test.go`'s header pins to a machine, met from a new side). The
fix is structural: the fragment hook lives in `plot()`, which is
all-integer — hooking it cannot move a float — and depth-rejected
fragments, whose only witness is the float loop itself, are deliberately
not reported. All five hashes on all three pins are unmoved by the whole
feature, which is the claim an instrument must be able to make.

The pad is the Keyer's Dreamcast variant: a level the Maple bus samples
once per field, queued whole (buttons, triggers and stick latch together)
and released one state per **three** fields — the tap length the oracle's
`-keys` scripts had already proven against the game's occasionally
overrunning mainline. Arrows drive the analog stick in the wire's own
convention (0x00 is up/left, the same convention the `jx`/`jy` scripts
document), with the gate modelled so a keyboard corner reads like a stick
corner and not a 1.41× diagonal no pad can produce. The whole path was
verified the way a user takes it — a raw WebSocket probe against a live
`-serve`: open, load, step, scrub to black, scrub to the drawn frame,
pixel → command 1201, keys accepted.

## Part X — The containers

The asset work opens the way every format effort here opens: an inventory,
and the layers that describe themselves. The tool is `ctarc`
(`extract/cmd/ctarc`), which reads everything straight off the disc.

**AFS** is a bare table: `AFS\0`, a u32 entry count, then (offset, size)
u32 pairs, data from 0x80000, and — on this disc — no name table (the word
after the last pair is zero, so an entry's identity is its index). The
census over every container says what the disc is made of. `BINC1-3.AFS`
hold **1,014 entries of one model format**: nearly every entry opens
`01 00 00 00  01 00 00 00`, a run of floats that reads as bounds, then
control words with `0x8xxxxxxx` shapes — and `POLDC1.BIN` opens the same
way, so the streamed objects and the course geometry are one family.
`SONG01.AFS` and `VOICE01.AFS` name their own contents: every entry starts
with the ADX signature (`0x8000`, copyright offset `0x20`, encoding 3,
block size 0x12, 4-bit), and the channel byte splits them exactly as the
filenames promise — stereo songs, 699 mono voice clips.

**PVRT** (`0GDTEX.PVR`) is the self-describing texture file: magic, pixel
format, data type, width, height. The decoder handles square-twiddled,
VQ and raster rectangles, and its verification is by eye: the file decodes
to the disc-label artwork the console's boot menu shows — the checkered
claw around the hub hole and the CRAZY TAXI wordmark, ARGB4444 with a
transparent border.

**LANDDC** entries are raster texture pages, no header at all: 32 KB each,
and the readable geometry is **256×64 at RGB565** — tile 37 of `LANDDC1`
is the EAST SIDE POLICE billboard, sign text and badge crisp. (128×128
shears it and a twiddled read shreds it; the page width is a fact about
the file, found by trying the three readouts a 32 KB page allows.)

`OBJDC1-3.BIN` are 32-byte placement records — a world position, a type
id, and a pointer into `0x0CE9xxxx` — RAM addresses baked into the file,
which is the tell that these tables link against model data loaded at a
fixed address. `COLDC*` opens with unit normals, plane constants and
vertex triples: the collision mesh. `MOTDC.BIN` opens with an offset
table. None of these three is guessed further than that on purpose: the
model, collision and motion layouts are the game's own parser's business,
and the next step is the proven one — watch the loader land `POLDC0` in
RAM, catch the code that walks it, and read the format off the
disassembly rather than off the bytes' shapes.

## Part XI — The model format, read off the game's own renderer

The promise at the end of Part X was kept literally. `-gd` (the disc-read
log, now carrying each read's RAM destination) named where everything
lands: `POLDC0` at `0C390000`, `COLDC1` at `0CB40000`, `MOTDC` at
`0C880000`, `OBJDC1` at `0CE90000` — the exact address its own baked
pointers promised — and the `BINC1` entries streaming through `0C6Axxxx`
during the drive. A new instrument, `-watchprof`, aggregates watch hits
into a per-PC histogram (its control: the input record at `0C1EBF6C`,
whose readers are known); pointed at a loaded model it named the code
that walks one. From there the format was read off the disassembly:
the dispatcher at `0C079600` (a model whose first word is 1 goes to
`0C080E20`, which steps past the 24-byte header), the block walker at
`0C080E44` (per-block cull sphere, then per-list-type staging through the
store queues; a culled block skips ahead by the size word at +76), and
the vertex loops at `0C081200` (strips) and `0C081C00` (triangle lists,
strip-flags bit 3).

The format, as `assets/model.go` decodes it: a model is
`[kind=1][sub][bounding sphere ×4f]` followed by a word stream. Word 0
ends the model; a word with bit 31 set opens an 80-byte block header —
real PowerVR TA templates `[PCW][ISP][TSP][TCW]`, a cull sphere, an aux
word, a lighting mode byte, an intensity, base and offset RGBA, and the
byte size of the strip stream that follows; a positive word is a strip:
`[flags][count]` then count vertex records (3×count for triangle lists).
A vertex record is either 32 inline bytes `[xyz][normal][uv]` — the
inline flag lives in the LSB of x's mantissa, which is why the files are
full of `3F800001` — or an 8-byte back-reference `[ctl][s32 offset]`
whose target is offset bytes past the pair: a vertex already in the
stream. After the end word most models carry their total vertex-record
count, a free consistency check on the whole parse.

Coverage is the argument that the reading is right: all four `POLDC`
files (1,696 models, 100,000+ vertex records) and all 1,025 non-empty
`BINC` entries parse with zero errors, and every trailer count matches
what was read. The `BINC` chunks turn out to be **world-placed** — their
vertices are in course coordinates — so `ctmodel -file BINC1.AFS -o
city.glb` concatenates 495 entries and **the arcade city assembles
itself**: downtown, the coast road, the grid, the stadium. `POLDC` also
holds local-coordinate object models (cull spheres of ~80 units — the
props and vehicles OBJDC places at runtime). The `-png` flag renders the
preview from the written GLB's own accessors, never from the structs
that wrote it.

## Part XII — The texture pipeline, closed offline

The model blocks carry TA texture words, but the file copies are
**zero-addressed**: format bits only, no VRAM address. Watching a block's
TCW in RAM shows two writes — the disc load, then a patch. The patcher at
`0C072FE0` stamps every textured block at course load from a table of
60-byte records (`rec = base + id·60`, base pointer at `[0C2AFE58]`):
width, height, byte size, flags, the allocated VRAM address, TSP low
bits, and the final TCW. The id it uses is **the block's aux word** — the
80-byte header's word at +32 is a texture id, dense per course, and over
all 2,512 `POLDC0` blocks it collides with nothing.

So the offline question became: where do the records come from? The
create-texture call chain was climbed with nothing but a literal-pool
scanner (SH-4 code addresses its neighbours through `mov.l @(disp,pc)`
pools, so "who calls X" is "which pool word holds X, and who loads that
word"): the record filler `0C15A2A2` ← the wrapper `0C15994C` ← the
create-by-id function `0C072B10` — which receives a **16-byte descriptor**
`[w u16][h u16][fmt u8][kind u8][.. ][staging ptr u32][alias u32]` and
the texture id, multiplies the id by 60, and fills the record ← the
course loader `0C079780`, which indexes **a static descriptor array** by
the aux id. The arrays live in `1ST_READ.BIN` itself. They were found
without another boot: the drive2 RAM dump holds the loader's finished
record table, so a scan for stride-16 arrays whose (w, h) pairs match a
run of records located all of them, and byte-comparing against the disc
copy of `1ST_READ.BIN` proved them static:

| course | array (RAM) | file offset | entries | pairs with |
|-------:|------------|------------:|--------:|-----------|
| 0 | `0C140570` | `0x130570` | 438 | `TEXDC0.BIN` |
| 1 | `0C144CB0` | `0x134CB0` | 336 | `TEXDC1.BIN` |
| 2 | `0C1461C0` | `0x1361C0` | 324 | `TEXDC2.BIN` |
| 3 | `0C147610` | `0x137610` | 225 | `TEXDC3.BIN` |

The **staging pointer is the file offset, baked at build time**: the
course loader streams `TEXDC<n>.BIN` to `0CB40000`, and every
descriptor's pointer is `0CB40000` plus its texture's cumulative offset.
The empirical truth table from the Part X-era cold boot (216 aux-id →
file-offset anchors, joined through c2d spans) matches `TEXDC0`'s array
on **all 216 points**, and each array's last offset plus its last size
lands on its file's byte size. The pairing with the model containers is
equally clean: the maximum aux id in `POLDC<n>`/`BINC<n>` is exactly the
entry count of `TEXDC<n>`'s array, minus one, for every n.

The kind byte was named by correlating descriptors against the final
TCWs the loader built (mip bit 31, VQ bit 30, format bits): kind 4 =
VQ + mips, 3 = VQ, 2 = twiddled + mips, 1 = twiddled, and `0xD` — only
ever on the non-square textures — raster pages. Format byte 0/1/2 =
ARGB1555/RGB565/ARGB4444, the PVR's own numbering. Sizes follow from the
kinds (VQ: 2 KB codebook + one index byte per 2×2 block; mip chains
store the sub-top levels first), each blob 32-byte aligned — the baked
consecutive offsets prove every formula.

One trap remained. Byte-comparing `TEXDC1` against the drive2 VRAM dump
at the records' own addresses verified 241 of 336 textures immediately —
and every mismatch was kind 2. `-watchprof` on a kind-2 staging address
found only ordinary copy loops, no transform, which killed the
"runtime-modified texture" theory; the real difference is two bytes of
layout. A PVR mip chain pads the sub-top levels with three dummy texels
before the top level; **the file pads with one**. At that +2 shift every
top-level texel of every probed texture matches — so the check was
rewritten kind-aware: 241 textures byte-identical, 95 kind-2 textures
texel-identical at the top level, **336 of 336, nothing unexplained**.

`assets/texdir.go` decodes the whole pipeline offline — descriptor
arrays out of `1ST_READ.BIN`, twiddle/VQ/mip/raster out of the `TEXDC`
payload — and `ctmodel` grew `-tex` (texture the export; the course is
the digit in the file name), `-dumptex` (every texture as PNG), and a
`-png` preview that decodes the **shipped GLB's own embedded images**
and samples them per pixel. `ctmodel -file BINC1.AFS -tex -o city.glb`
now assembles the arcade city with its shopfronts, brick, foliage,
crosswalks, stadium crowd and sea — every texel of it decoded from the
disc alone.

The three courses shipped to the Studio (`extract/cmd/webexport`): the
arcade city, the original city — tower, Ferris wheel, harbour and
lighthouse — and the Crazy Box arenas under their painted sky dome, each
a textured GLB with a fly camera. Course 0 is not a course: POLDC0 and
TEXDC0 load at boot and hold the shared object models, which is the next
frontier along with OBJDC's placement chains.

## Part XIII — The sky, the cabs, the drivers, the traffic

The drive renders under a photographic blue sky, but nothing in the
course files looked like one. The suspects fell in order: `SPLDC1.BIN`
(five loads at course time, never examined) censused as 99.6% floats —
world-coordinate triples marching down the road at even spacing, a
**spline file**, not geometry; `POLDC1`'s 45 models held thirty copies of
one translucent sea animation and ten untextured hills, no dome. So the
sky was traced instead of guessed: per-pixel provenance over a drive2
frame (the framedbg capture) put **every sky pixel under one polygon
header**, whose TCW named a 256×128 RGB565 texture at VRAM `5FCDC0` —
**twiddled**, per the scan-order bit. The loader's own record table
mapped that address to a 64 KB blob byte-identical to `TEXDC1` entry
218; decoding it as a row of two 128-sided twiddled squares produced the
cloud panorama exactly (and the old "raster" decode produced the striped
shreds that had hidden it in the texture dump). Kind `0xD` was never
raster: **every non-square record in the live table has scan-order
clear** — non-square twiddled, laid out as min(w,h)-sided squares in row
order. `texdir.go` now decodes it that way, with the synthetic test
pinning both squares.

The geometry followed from the texture: searching every course model for
blocks referencing entry 218 found only four horizon patches — but
searching **POLDC0** for its own 256×128 slot (aux 150) found **model
272: a 154-vertex dome centred at the origin, cull radius 4193** — and
the drive2 RAM copy of its block header carries the *exact TCW the sky
pixels traced to*. Its constant companion is **model 258, a translucent
horizon ring** (aux 108, the coastline strip photo). The cloud panorama
ships byte-identical in TEXDC0/150, TEXDC1/218 and TEXDC3/139, and the
ring's VQ codebook matches TEXDC0/108's — **all courses share the boot
directory's one sky**. The captured draw transform places both models at
the camera position, world-yawed, at uniform scale 2 — so the Studio
ships them as each course's camera-attached `sky` layer (painted first,
never depth-tested), the exact discipline the game uses.

The cabs and drivers came from the dispatch pointer, not from shapes.
The renderer enters every drawn model through `jsr 0C080E20` with r4 =
the model; logging r4 across the driver-select carousel — where the
screen itself captions AXEL, B.D.Joe, GENA, GUS — isolates each
character's ~20 body parts, and logging four separate drives (one per
selected driver, the licence plates on screen matching the select art)
isolates each cab: **Axel m9, B.D.Joe m25, Gena m51, Gus m37**, each
with its roof-sign pair, an interior model, and four wheels — Gena's and
Gus's their own, **Axel's and B.D.Joe's cabs sharing one wheel set**
(m7/8/10/11).

Placement numbers were captured, not invented: the game loads each
model's transform into **XMTRX** before the dispatch, so a breakpoint
harness (`mtxcap`, scratch) snapshots the back FP bank per dispatch. The
matrices are singular by construction (the loader negates the fourth
row against the first — `frchg` + `fneg` at `0C0795C0`), but rows 0–2
solve relative placement: `R = B₃⁻¹C₃, t = B₃⁻¹(c−b)`. That yielded
every wheel anchor in cab space (per-cab track *and* wheelbase:
±15.5/±15.25/±14.6/±14.5 — identity rotation at standstill), proved the
sign and interior sit at cab identity, located the seated driver on the
**+x side with the pedals at +z** (left-hand drive, so the wheel labels
are honest), and measured the sky pair's camera translation and scale-2
draw. For the drivers the same captures at the four carousel positions
bake the standing poses — with **one shared reference** (Axel's torso
capture), valid because all four positions render through the same
select camera, so the view factor cancels and all four figures land
upright in one frame (a per-driver torso reference laid Gus on his
side; the fix is the shared-camera observation, not a fudge).

The traffic fleet is four photo-quad impostor families: a coarse body
(190–212 vertices whose flanks are photographs of a whole car) plus one
16-vertex **wheel model placed once per axle** — captured anchors (0,
3.1–3.4, ±13.5…±15.9) — whose quads carry the wheel photo and spin
about x at speed. Each family also ships a 56-vertex variant head
sharing the body's textures (the glass panels in a raised pose; the
live frame shows traffic glass closed, so the raised copies are a state
the runtime resolves — noted open, not chased). The Studio gets the
four assembled vehicles.

`webexport` (`objects.go` + the generated `poses.go`) now ships, beside
the three courses with their new sky layer: four assembled cabs (body,
roof sign, interior, four placed wheels — every part a named glTF node,
the carex pattern), four posed drivers under the game's own names, and
the four traffic cars. Verified the only honest way: every rendered
check reopens the **shipped GLB** (`ctmodel -glbin`), and the cab, the
upright drivers, the closed sky dome over its coastline ring all read
back from the files the site serves.
