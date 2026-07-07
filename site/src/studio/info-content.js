// Technical-manual text for the info panel, one entry per game per tab. The prose is derived
// from each game's Markdown write-up but rewritten in a neutral reference style: the
// reverse-engineering narrative and history in the source docs are dropped, leaving a
// description of how the shipped game works.
//
// The five tabs fold the Markdown's parts into reader-facing sections:
//   loader   -> Parts I & II  (the disk/tape image and the boot/loader chain)
//   engine   -> Part III      (the game program's architecture / main loop)
//   graphics -> Part IV       (graphics and data formats)
//   music    -> Part VI       (the music engine and tracks)
//   gameplay -> Part V        (game mechanics)
//
// Content is filled in over subsequent passes. INFO_CONTENT[gameId][tabId] is an HTML string
// (rendered inside .info-doc); a missing entry shows a "not written yet" placeholder.

export const INFO_TABS = [
  { id: 'loader', label: 'Image & Loader' },
  { id: 'engine', label: 'Game Engine' },
  { id: 'graphics', label: 'Graphics' },
  { id: 'music', label: 'Music' },
  { id: 'gameplay', label: 'Gameplay' },
];

export const INFO_CONTENT = {
  sonic: {
    loader: `
<div class="info-eyebrow">Sonic the Hedgehog · Image &amp; Loader</div>
<p>Sonic runs from a Game Gear cartridge, and a cartridge image is far simpler than a disk or tape image: it is
a verbatim copy of the mask-ROM chip, with <strong>no container, filesystem or loader</strong>. The only structure the
console imposes is a memory map (the ROM is larger than the Z80 can address) and a small Sega header,
so there is nothing to unpack — boot is a short reset routine that brings the hardware up and hands
off to the game.</p>

<h2>The cartridge and its memory map</h2>
<p>The ROM is 256&nbsp;KB. The Z80's address bus is 16-bit, so it can only see 64&nbsp;KB at once —
a quarter of the cartridge. The ROM is divided into sixteen 16&nbsp;KB <strong>banks</strong>, and the
standard Sega mapper pages a chosen bank into one of three 16&nbsp;KB <strong>slots</strong> in the low
48&nbsp;KB; the top 16&nbsp;KB is work RAM. Four mapper registers at the very top of the address space
select which bank shows in each slot — and they physically <em>are</em> the top four bytes of RAM
(mirrored there), so a write both stores the byte and reprograms the mapper. At reset the slots default
to banks 0/1/2, which is why the engine and boot code sit in the first 48&nbsp;KB. One subtlety: the
first 1&nbsp;KB is <strong>hard-wired to bank 0</strong> and never paged, so the CPU vectors and the
mapper-setup code stay reachable however slot 0 is paged. The graphics and sound hardware are not in the
memory map at all — the Z80 reaches the VDP and PSG through I/O ports.</p>

<h2>The header and vectors</h2>
<p>Sega stamps a 16-byte <code>"TMR SEGA"</code> header near the front: a magic string, a product code
and version, and a region / ROM-size nibble (here Game Gear export, 256&nbsp;KB). On the Master System
that field gates a BIOS check; the Game Gear has no such gate, so it is informational. The Z80's fixed
entry points all sit at the bottom of the always-present bank 0 — the reset address, eight one-byte
<code>RST</code> call targets, and the interrupt vectors. Sonic routes its hottest common subroutines
through the <code>RST</code> slots (each is a one-byte call), and the maskable-interrupt vector jumps to
the per-frame VDP handler.</p>

<h2>Reset and cold start</h2>
<p>The reset code is the textbook Master System / Game Gear opening: disable interrupts, select interrupt
mode 1, busy-wait on the VDP until the raster reaches a known line, then jump to the real cold start. The
cold start re-asserts the default bank layout, clears the 8&nbsp;KB of work RAM with the classic
overlapping <code>LDIR</code> fill (stopping 16 bytes short so it does not clobber the mapper-register
mirror) and parks the stack at the top, brings the VDP up in <strong>Mode 4</strong> from a register table
(keeping a RAM shadow the interrupt handler reads back), and <strong>hides all 64 sprites</strong> by
setting their Y off-screen before the display comes on. It then calls a setup routine in another bank
through the banked-call gateway and hands off to the main entry.</p>

<h2>Cross-bank calls</h2>
<p>Because most of the game lives outside the 64&nbsp;KB the CPU can see at once, the engine reaches it
through <strong>banked-call thunks</strong> that are the <code>RST $18/$20/$28</code> vectors: each pages
a fixed bank into a slot, calls a fixed entry in it, then restores the previous bank — read back from a
RAM shadow that tracks the current slot banks so the calls can nest. <code>RST $20</code> lands in bank 3,
which holds a dispatcher the engine reaches through these one-byte calls. The main entry enables
interrupts, sets the bank shadows to the running configuration, runs the subsystem initialisers, and sets
the top-level game-mode variable — and from there control is in the main loop.</p>
`,
    engine: `
<div class="info-eyebrow">Sonic the Hedgehog · Game Engine</div>
<p>After boot, Sonic runs an attract sequence over a small <strong>state machine</strong>: a scene system
that loads the logo, the title, the world map and the levels in turn, driven by a one-byte scene counter
and a function-table dispatcher that reaches across the ROM's banks.</p>

<h2>The attract loop and scene state machine</h2>
<p>The main entry loads the SEGA logo and enters an <strong>attract loop</strong> driven by a one-byte
scene counter: each pass loads a screen, fades it in, runs it, and advances; when the counter passes the
last scene it restarts from the logo. Each scene is described by a roughly 40-byte <strong>descriptor</strong>
in a table (18 entries), which the dispatcher copies into RAM to run and maps to a screen type — the
early scenes are the title background, the later ones the world map. It only reloads the background when
the type changes, so the title stays up across its scenes while the foreground animates, and the map
likewise persists across its scenes.</p>

<h2>Title and Start</h2>
<p>While the title is up, the scene scripts animate Sonic (the finger-tapping pose) and blink
<code>PRESS START BUTTON</code>. The wait for Start is not a poll loop — it is folded into the attract
loop through two flags: the controller is read every frame, and when Start is pressed a handler raises a
<strong>launch flag</strong> and writes a target scene. On the next pass the loop sees the flag, skips the
idle wait, and uses the target scene instead of the counter — so Start jumps the sequence out of the demo
and into the post-title flow rather than letting it free-run. The same Start bit is checked during the
logo, which is why the logo and title are skippable.</p>

<h2>The world map</h2>
<p>There are <strong>two</strong> world-map screens — a different tile set and a different stored map for
each — and the engine picks between them by the level you are heading into. A countdown byte, decremented
after each level, selects the wide island for early levels and zooms in on the mountain-top goal for late
ones. It is the same island both times: the castle on the peak of the wide shot is the city that fills the
zoomed shot. On either map the route and the zone name are not in the stored map — they are a per-scene
overlay drawn by the string blitter, and the blink is that overlay repainted on a timer.</p>

<h2>Reaching a level</h2>
<p>Pressing Start does not "return into the game" — the attract loop never exits. Instead Start selects a
target scene whose script <em>itself</em> loads the level and jumps into the gameplay engine, which runs
from banked code (the gameplay engine and the level variables live in bank 1, reached through the bank-3
dispatcher). Once in play, a per-frame <strong>scroll / camera update</strong> recomputes the tile-stream
source from the camera position each frame, feeding the streamer that draws the level a column at a time.</p>
`,
    graphics: `
<div class="info-eyebrow">Sonic the Hedgehog · Graphics</div>
<p>Game Gear graphics are two layers: the fixed <strong>VDP hardware formats</strong> the data ends up in,
and the game-specific <strong>compression</strong> it is stored under in the ROM. Almost all of Sonic's
art is compressed, and levels are built from a small alphabet of reusable blocks.</p>

<h2>VDP formats</h2>
<p>These are standard Mode 4 formats. A <strong>tile</strong> is 8&times;8 pixels in four bitplanes (16
colours), 32 bytes, stored row-interleaved. The <strong>palette</strong> is 32 entries of 12-bit colour in
BGR order — 16 background, 16 sprite. The <strong>name table</strong> is the 32&times;28 background map in
VRAM; each cell is a two-byte word giving a 9-bit tile number plus per-cell flip, palette-select and
priority bits.</p>

<h2>Compression</h2>
<p>A single codec compresses almost all of Sonic's art — a <strong>four-byte-unit LZ</strong> whose unit is
one tile row (four bitplane bytes), so a repeated row (a blank row, a flat fill) costs a single bit. A
compressed block is addressed as a <code>(bank, address)</code> pair, and its prologue normalises the
address and maps two consecutive banks into the slots, so a block can span banks. The block holds a control
bitmap (one bit per output unit), a literal stream of the distinct four-byte rows in first-appearance order,
and a match-info stream of back-reference offsets; decoding walks the units, emitting either the next
literal row or a back-reference to an earlier one.</p>

<h2>The opening screens</h2>
<p>The two screens before the menu reach the same VRAM by opposite routes. The <strong>SEGA logo</strong>'s
tile map is computed in code — there is no stored map: a loop lays down a plain identity grid one vertical
column at a time, left to right, so the logo wipes in column by column behind <strong>Sonic</strong> as he
jumps across the screen. He leaps from the left, the logo drawing in behind him as he goes, then jumps back
to the left and comes to rest beside the finished logo — his parabolic arc comes from a symmetric offset
table the frame interrupt flushes to the sprite list each vblank. The <strong>title</strong> takes the other
approach, loading its name table wholesale from a stored, compressed map (a tiny RLE name-table loader runs
twice, composing a priority base layer drawn in front of sprites and a plain overlay).
<code>PRESS START BUTTON</code> is painted on afterwards by a string blitter and blinks — repainted every
time it toggles.</p>

<h2>Level maps</h2>
<p>A level is far wider than the screen, scrolls, and is built from reusable <strong>blocks</strong> rather
than per-cell tile numbers. Sonic stores it as three nested layers plus a palette:</p>
<ul>
  <li>a compressed <strong>block-index map</strong> — an RLE list of block indices that decompresses to a
  fixed 4096-byte grid in a RAM window the engine reads as it scrolls. The scenery — hills, palms, flowers,
  even the ring graphics — is part of this block map; what it does <em>not</em> contain is the object layer.</li>
  <li>a <strong>block → tiles table</strong> — each block index expands to a 4&times;4 grid of 8&times;8
  tiles (a 32&times;32-pixel chunk), 16 bytes per block. A column expander reads one map column, looks up
  each block's tiles, and uploads a column to VRAM as the camera crosses each 8-pixel boundary.</li>
  <li>the <strong>tile graphics</strong> — 256 background tiles, compressed with the same codec, plus a
  palette loaded by index through a table.</li>
</ul>
<p>Each act is driven by a 37-byte <strong>descriptor</strong> — the zone, scroll bounds, spawn block, the
offsets of the map, block table and tile set, the object table, the music id and flags. Because that table
is static, every act of a zone follows from it: the three acts of a zone share their tile set, block table
and palette, differing only in the block-index map and the width.</p>

<h2>A map that isn't always wide</h2>
<p>The decompressed map is always 4096 bytes but not always 16&times;256: the descriptor carries a
<strong>stride</strong> that sets the column count, so the grid is <code>(4096 / stride)</code> rows tall.
Most acts are wide (stride 256 → 16&times;256), but a small stride is a tall, narrow level — Jungle Act 2
is a 256-row &times; 16-column vertical waterfall climb. The grid is always 4096 cells; how they are laid
out is data-driven per act.</p>

<h2>Animated tiles</h2>
<p>Some graphics animate by swapping tile data in place. The <strong>rings</strong> (a 16&times;16 graphic
is four 8&times;8 tiles) spin through about six frames, copied from a fixed source for every zone; the Green
Hills <strong>flowers</strong> are a two-frame toggle gated on the zone. Both tile sets are empty in the base
set — a per-frame update copies a fresh frame into them each cycle. There is no generic animation table; each
animation is hardcoded.</p>

<h2>Sprites</h2>
<p>Objects — Sonic, the enemies, items — are drawn as <strong>hardware sprites</strong> (8&times;16 mode),
not from the tilemap. A per-zone sprite tile set is decompressed into VRAM, and the sprite-attribute table
lists each active sprite's position and tile; the per-object draw builds a display list in RAM that is
flushed to the attribute table each vblank. Sonic's animation is <strong>data-driven and streamed</strong>:
his pose indexes an animation table to a frame sequence (a byte stream with loop/jump controls), and when the
frame changes its 16 tiles are streamed into the sprite VRAM — so only the live frame (and its neighbours) is
resident, not a full sheet.</p>
`,
    gameplay: `
<div class="info-eyebrow">Sonic the Hedgehog · Gameplay</div>
<p>Sonic himself is just object type <code>$00</code>, and everything interactive in a level — the player,
the enemies, the platforms, the springs — runs through one <strong>object system</strong> and one master
dispatch.</p>

<h2>The object system</h2>
<p>Each act's descriptor points to an <strong>object table</strong> — a count followed by three-byte entries
(type, block X, block Y) — expanded at load into an array of fixed-size records; record 0 is Sonic, placed
from the spawn pointer. Every live object is processed once per frame by a <strong>master dispatch</strong>:
a word-pointer table indexed by the type byte selects the object's behaviour handler, which runs and then
falls through to common "apply velocity to position" code. The type also indexes a bounding-box table for
the object's size and a descriptor for its sprite class. So a single type byte ties an object to its
behaviour, its size and its sprite.</p>

<h2>Sonic's movement</h2>
<p>Sonic's handler is the largest in the game, turning the controller into motion. The pad is sampled once
per frame and is active-low (a pressed button reads <code>0</code>). Each frame the handler picks a
<strong>physics parameter set</strong> — acceleration, friction and a top-speed cap — by Sonic's state and
whether he is on the ground, then accelerates him while a direction is held and decelerates by friction when
nothing is; the resulting speed integrates into his world position. Pressing <strong>Down</strong> sets the
rolling/ball flag (and, only if he is moving, plays the roll sound — standing still and pressing Down is a
crouch); rolling selects low-acceleration, <strong>low-friction</strong> constants, so a roll holds its
momentum — you steer only a little and mostly coast, carrying farther than Sonic would slide to a stop on
foot. <strong>Jumping</strong> gives variable height: a hold timer is seeded, and while the button
stays held and the timer counts down the upward impulse keeps being added; once released or expired, gravity
is applied each frame instead.</p>

<h2>Terrain and rings</h2>
<p>Collision reads the same block map the level is built from, through a shared sampler that returns the
block at a world point (it accounts for the level's stride, so a sampled point lands on the right block
whatever the level's shape). The plain <strong>solid floor</strong> is a generic per-object routine: it
samples the block at the feet, reads a collision shape from a per-zone attribute table (shape 0 means
non-solid — Sonic passes through), and uses a per-column <strong>height profile</strong> so he lands on
slopes at the right height and angle. Special blocks are a separate interaction layer — springs, spikes and
conveyors mapped through a per-zone table to handlers (a vertical spring sets an upward velocity, a
horizontal spring an X velocity, spikes hurt, conveyors add a steady drift), each gated on which 16-pixel
sub-cell of the block Sonic is in. <strong>Rings</strong> are baked into the block map, not objects: certain
block indices mark a left/right pair of rings, and the low bits of the index say which halves still hold one.
Collecting one is a single byte-write back into the map (the graphic downgrades two → one → none), a sparkle,
and a bump to the BCD ring count — which grants an extra life every time it passes 100.</p>

<h2>Enemies and platforms</h2>
<p>Each object type is a self-contained behaviour on its record. The <strong>platforms</strong> carry Sonic
explicitly: when he stands on one, the handler glues him on and adds the platform's per-frame motion directly
to his position. There are five kinds. A <strong>horizontal platform</strong> glides one pixel per frame
160&nbsp;px out and back; a <strong>swinging platform</strong> reads an arc table to trace a semicircle below
its pivot. A <strong>sinking platform</strong> holds still until Sonic steps on, then gives under his weight —
settling about half a pixel per frame until it has dropped 16&nbsp;px, and floating back up once he leaves.
Two more ride the water: a <strong>bobbing platform</strong> (Bridge's floating logs) drifts down from where
it spawns and then bobs gently on a fixed cycle with no terrain contact of its own, while the Jungle's
<strong>floating log</strong> is rideable in a second way — pushing left or right rolls it, and it carries
Sonic along at half his speed, its end-grain turning as it goes, until it fetches up against solid ground. The
<strong>enemies</strong> share
a small script engine — a counter into a state table, one entry per state for a fixed run of frames: the crab
walks, stops and fires a projectile to each side; the beetle marches faster with no attack; the fish hides
underwater and leaps straight up; the porcupine is a slow spiky walker whose spikes are the threat. The
<strong>seesaw</strong> is simulated rather than scripted: a tilt angle with a weight that falls under
gravity, and Sonic's landing impact is transferred into the tilt — so a harder landing flings whatever is on
the other end higher, with no fixed launch height. Each <strong>boss</strong> sets itself up as a
self-contained set-piece, decompressing its own graphics and running a small bytecode script; the world-1
boss takes eight hits, scored only when Sonic touches it while rolling or jumping.</p>

<h2>Checkpoints and scenery that lives</h2>
<p>Some "objects" are not enemies at all. A <strong>background animator</strong> repaints its own map cell
every frame through the same request registers the scrolling engine uses, cycling a four-phase block
sequence — these are Green Hills' growing flowers and twinkling sea. (It was long mislabelled a "camera
lock": the register it writes is the scroll-draw <em>request</em>, not the camera.) A
<strong>checkpoint</strong>, on contact, writes its own block position into a per-act respawn table — so a
death returns Sonic to the checkpoint, one block above it.</p>

<h2>Underwater</h2>
<p>Labyrinth is the only zone with water, and its acts are half-flooded, split by a horizontal water line.
The <strong>water surface is itself an object</strong> — the first one placed in each Labyrinth act, its
block-Y the water level — and each frame it writes the current surface Y (with a sine bob, so the line
ripples) and derives the scanline where the split falls. Crossing the line changes three things at once, all
driven by one underwater flag: the <strong>palette</strong> swaps via a mid-frame raster split (a line
interrupt rewrites the 16 background colours to a static underwater set below the line, restored at vblank
above it); the <strong>physics</strong> load a slower constant set (roughly quarter acceleration and gravity
and a weaker jump, so Sonic drifts down slowly); and an <strong>air timer</strong> starts that, past about 13
seconds, triggers the drowning countdown. Drowning is often assumed to be a feature only of the 16-bit
Mega Drive games, but this 8-bit version has it too.</p>

<h2>Bonus stages</h2>
<p>Clearing Act 1 or 2 of a zone with 50 or more rings sends you to a <strong>bonus stage</strong> instead of
the next act. The goal sign checks the ring count and sets a bonus flag; the next scene load runs the bonus
path, swapping in a separate bonus-stage cursor that advances each visit. The bonus stages are eight more
descriptors in the same scene table, in a seventh "zone", and they all share one tilemap, tile set, block
table and palette — what differs per round is the spawn, the camera bounds and the object layout: the rings,
a collectible, and bumpers that reverse Sonic's velocity on contact.</p>

<h2>The hidden Scrap Brain maze</h2>
<p>An invisible <strong>teleporter</strong> object converts its own position to a key, looks it up in a small
table, and launches the matched destination scene — the very mechanism Start uses on the title. Because each
teleporter's destination is hardwired to where it sits, the placements expose something the act list hides:
<strong>Scrap Brain Act 2 is not one level but seven</strong> — the listed act plus six sub-scenes reachable
only through teleporters, wired into a maze with deliberate loops so a wrong choice sends you in circles. A
second teleporter, in Sky Base, warps to a hidden fortress room holding a goal sign and a Chaos Emerald,
outside the normal act flow entirely.</p>
`,
    music: `
<div class="info-eyebrow">Sonic the Hedgehog · Music</div>
<p>The Game Gear's audio is the <strong>SN76489 PSG</strong> — three square-wave tone channels and one noise
channel, programmed through a single write port. There is no FM on the Game Gear; the music is square waves,
reprogrammed once per video frame.</p>

<h2>The sound driver</h2>
<p>A sound is started by a one-byte <code>RST</code> call with a sound id; the gateway pages in the sound
bank and indexes a song-pointer table. A song begins with a header of <strong>five relative channel
offsets</strong> — three square tones, the noise channel, and a sound-effect channel that stays idle during
music — which the loader resolves to absolute pointers in RAM. A per-frame <strong>sequencer</strong>
advances each channel and renders it to the PSG.</p>

<h2>The channel format</h2>
<p>Each channel is a byte stream with a duration counter decremented each frame; when it runs out the decoder
reads the next bytes:</p>
<ul>
  <li>a <strong>note</strong> is two bytes — an octave/note value (the pitch is a base period from a
  one-octave table, shifted down by the octave) plus a duration;</li>
  <li>a <strong>rest</strong> silences the channel for a duration;</li>
  <li>a <strong>voice</strong> command picks an instrument (a noise mode plus an ADSR envelope) — and on the
  noise channel the "notes" are voice commands, each a drum hit, which is the zone's percussion;</li>
  <li>a <strong>command</strong> byte (tempo, volume, instrument envelope, vibrato, detune, block-repeat with
  a nested loop stack, a loop-point mark and an end/loop) carries no time of its own.</li>
</ul>
<p>The loop is <strong>in the data</strong>: a loop-point mark and a loop command bracket the repeating
section, which is why each track loops seamlessly. Per frame the render forms the pitch as base + detune +
vibrato, shifted by the octave, and scales the volume with the ADSR envelope before writing the PSG.</p>

<h2>Which track plays</h2>
<p>The track for an act is its descriptor's <strong>music id</strong>, indexing the song table. The ids
resolve to about fifteen distinct tunes — the six zone themes (Sky Base acts 1 and 3 reuse Scrap Brain's),
the special-stage theme, the world map, act-clear, a shared boss theme, and assorted jingles — so, for
example, Sky Base Act 1 correctly plays the Scrap Brain music.</p>
`,
  },
  fort: {
    loader: `
<div class="info-eyebrow">Fort Apocalypse · Image &amp; Loader</div>
<p>Fort Apocalypse ships on a single cassette, preserved as a <strong>TAP image</strong>: a
recording of the raw signal the C64 reads off tape. Loading runs in two phases — a short
bootstrap in the ROM's standard tape format, then a custom high-speed loader that streams the
rest of the game in while a U.S. Gold loading screen plays.</p>

<h2>The tape image</h2>
<p>A TAP file stores the cassette signal as a list of pulse lengths, one byte each. After a
20-byte header (the magic string <code>C64-TAPE-RAW</code>, a version byte and a little-endian
data length), every non-zero byte is a single pulse of <code>n &times; 8</code> clock cycles —
985,248&nbsp;Hz on PAL. A zero byte marks a pause and is followed by a 24-bit cycle count. Only a
handful of distinct pulse widths carry data: roughly 300 and 670 cycles for the fastloader's 0 and
1 bits, and three medium widths used by the KERNAL bootstrap.</p>
<p>Three regions sit back to back, separated by pauses: a KERNAL header block, a KERNAL data block,
and the fastloader stream. The two KERNAL blocks are each recorded twice for reliability.</p>

<h3>The KERNAL bootstrap</h3>
<p>The first two blocks use the C64's built-in ROM tape format. Each bit is a <em>pair</em> of
pulses (short+medium = 0, medium+short = 1); a byte is a marker pair, eight data bits LSB-first,
and an odd parity bit. Each record carries a pilot tone, a nine-byte countdown sync sequence, the
payload, and an XOR checksum. Two records load:</p>
<ul>
  <li>a <strong>header block</strong> announcing a relocatable BASIC program named <code>FORT</code>.
  Its payload is nominally the 16-character filename — but the bytes after the name are not padding.
  The KERNAL copies the whole header into the cassette buffer at <code>$033C</code>, which quietly
  plants machine code at <strong><code>$0351–$03F5</code></strong>: the fastloader's interrupt
  handler.</li>
  <li>a <strong>data block</strong> loaded to <code>$0801</code> — a one-line BASIC program,
  <code>SYS 2061</code>, followed by the loader-setup code at <code>$080D</code>.</li>
</ul>
<p>So <code>LOAD"",1</code> then <code>RUN</code> runs the BASIC stub, which <code>SYS</code>es into
the setup routine — and the real loader is already resident in the tape buffer, smuggled in
disguised as a filename.</p>

<h3>The NOVALOAD fastloader</h3>
<p>The bulk of the game arrives through a custom turbo loader (it names itself <strong>NOVALOAD</strong>,
serial D100701, on screen) that reads <strong>one bit per pulse</strong> rather than two. CIA timer A
is latched and force-reloaded on every cassette edge; the interrupt handler reads the timer's high
byte and treats a short pulse as 0, a long pulse as 1. Bits are rotated in with <code>ROR</code>, so
the first pulse of a byte is its least significant bit. The shift register starts at <code>$7F</code>,
and any run of eight-or-more zero bits ending in a one bit reads back as the pilot byte
<code>$80</code> — which is how the decoder self-synchronises without an explicit reset.</p>
<p>The stream is a pilot tone, a sync byte (<code>$AA</code>), a key byte (<code>$55</code>), then
<strong>84 records</strong>, each a page number, 256 data bytes, and a checksum
(<code>page + sum of bytes, mod 256</code>). Every record loads to <code>page &lt;&lt; 8</code>, so
pages may arrive in any order and gaps are harmless. One record carries page <code>$F0</code>, which
arms "end mode"; after it, a page number of <code>$00</code> ends the load. The pages come in two
groups: first the stage-2 loading screen (<code>$E000–$E6FF</code> and <code>$EE00–$F1FF</code>),
then the main game (<code>$7000–$B8FF</code>) streamed in behind it.</p>

<h2>The boot chain</h2>
<p>End to end, control flows:</p>
<ol>
  <li><code>SYS 2061</code> runs the loader setup at <code>$080D</code>.</li>
  <li>Setup banks out the BASIC and KERNAL ROMs, points the CPU's IRQ vector at the planted handler
  at <code>$0351</code>, arms a CIA FLAG interrupt that fires once per tape pulse, and busy-waits
  until page <code>$F0</code> has arrived.</li>
  <li>With the loading screen now in memory, it calls stage 2 at <code>$E000</code> while the
  interrupt keeps streaming the game in the background.</li>
  <li>On success, stage 2 fades the music, banks the ROMs back in, and jumps to the game's
  initialisation at <code>$8600</code>.</li>
</ol>

<h3>Loader setup ($080D)</h3>
<p>Besides redirecting interrupts, setup clears the screen and prints the filename and the
<code>NOVALOAD D100701</code> banner, primes the SID for the loading-noise effect (each loaded byte
is also written to a SID register), and lays out its zero-page state: a store pointer, a page
offset, the checksum seed, and a status byte (loading / done / error). It also aims the BASIC text
pointer at a planted <code>:RUN</code> token sequence — a decoy that makes a memory snapshot look
like a harmless return to BASIC.</p>

<h3>The interrupt handler ($0351)</h3>
<p>The handler runs once per tape pulse. After demodulating the bit and assembling a byte, a
self-modifying branch offset dispatches a small state machine: search for the pilot, match the sync
byte, verify the checksum, read a page number (<code>$F0</code> arming end mode, <code>$00</code>
afterwards completing the load), or store 256 data bytes and accumulate the checksum. A bad checksum
sets the error status and halts the load.</p>

<h2>The loading screen</h2>
<p>While the game streams in, stage 2 paints the U.S. Gold loading screen and runs a scroller and
three-voice music. The screen is drawn from a compact <strong>display script</strong> — border and
background colours, then runs of screen codes with a single escape byte for newlines, colour changes
and end-of-script. It includes the three-digit "BLOCKS TO LOAD" counter that the tape interrupt
decrements as each page arrives. The scrolling message is stored <strong>reversed</strong> and read
backwards through a self-modified pointer.</p>

<h3>A tune that is also a program</h3>
<p>The music is not merely audio: it is a small bytecode the player interprets. Commands play a note
for a duration, set the read pointer (to loop the tune), or — the notable one — copy the next
<em>n</em> stream bytes to <strong>any address in memory</strong>, implemented by patching the
operand of a store instruction. The tune uses that copy command, on its very first tick, to rewrite
the machine itself:</p>
<ul>
  <li>it redirects the KERNAL NMI vector so that RUN/STOP–RESTORE becomes a clean no-op during play;</li>
  <li>it re-initialises the SID and some player variables;</li>
  <li>and, crucially, it overwrites the loader's epilogue at <strong><code>$03F5</code></strong> with
  <code>JMP $8600</code> — the jump that actually starts the game.</li>
</ul>
<p>As loaded from tape, that epilogue ends in an innocuous <code>RTS</code> followed by the decoy
<code>:RUN</code> bytes. The real entry address appears nowhere in the code; it exists only as data
inside the music stream, and only the act of playing the tune assembles it.</p>

<h3>Error handling</h3>
<p>On a clean load, stage 2 fades the volume over a few seconds, clears the SID, restores the ROMs,
and takes the patched jump into the game. If a tape error stalls the loader — detected as a frozen
byte counter — stage 2 instead <strong>wipes all RAM except its own page and jumps through the reset
vector</strong>, a response that is as much anti-tamper as error recovery.</p>
`,
    engine: `
<div class="info-eyebrow">Fort Apocalypse · Game Engine</div>
<p>Once loaded, Fort Apocalypse is an almost entirely <strong>interrupt-driven</strong> program. A
brief setup routine builds the world in memory and arms a raster interrupt, then deliberately parks
the processor in a tight infinite loop — every frame of the game is produced by the raster handlers
and a main loop that they release.</p>

<h2>Initialization ($8600)</h2>
<p>Entry at <code>$8600</code> jumps straight to the init routine at <code>$8927</code>. It runs once:</p>
<ul>
  <li><code>SEI</code>, <code>CLD</code>, clear zero page, and set <code>$01 = $2E</code> — <strong>BASIC
  ROM banked out, KERNAL left in</strong>, so the game's own code in the <code>$A000–$B8FF</code> region
  is called directly underneath where BASIC used to be.</li>
  <li>Point the VIC at bank 1 (<code>$4000–$7FFF</code>) through CIA2, with the screen at
  <code>$4400</code>.</li>
  <li>Reset the SID — and set voice 3 to noise, whose output at <code>$D41B</code> becomes the game's
  <strong>random-number source</strong>.</li>
  <li>Zero <code>$0380–$6FFF</code>, then build both character sets and expand all sprite shapes
  (see Graphics).</li>
  <li>Draw the HUD frame and title text with a double-width font renderer: each glyph is drawn as
  character <code>n</code> alongside character <code>n+$20</code>.</li>
  <li>Install the title raster interrupt at line <code>$F9</code> and finish with
  <code>$8A9F: JMP $8A9F</code> — a one-instruction halt. Everything after this point happens inside
  interrupts.</li>
</ul>

<h2>The raster architecture</h2>
<p>The display is split into two horizontal bands, each served by its own interrupt handler that
reprograms the VIC mid-frame and chains to the next split:</p>
<ul>
  <li><strong>Line <code>$F9</code> — the HUD handler (<code>$9BD4</code>).</strong> Selects the HUD
  character set (<code>$D018 = $14</code>, charset <code>$5000</code>), sets the scroll registers,
  latches the sprite-collision registers, increments the frame counter, reads keyboard and joystick,
  updates the player sprite, bullets and the enemy sprite, drives sound, and schedules the next
  interrupt for line <code>$76</code>.</li>
  <li><strong>Line <code>$76</code> — the playfield handler (<code>$AE19</code>).</strong> Selects the
  playfield character set (<code>$D018 = $16</code>, charset <code>$5800</code>), applies fine
  scrolling, sets the per-level colours, runs the in-place charset animations, copies the scrolling
  playfield window, applies SID effects, and schedules the next interrupt back at line
  <code>$F9</code>.</li>
</ul>
<p>The consequence of the split is that screen rows 0–6 (the HUD and scanner) and rows 7–24 (the
playfield) are drawn from <strong>two different character sets</strong>, swapped partway down every
frame.</p>

<h2>The main loop and game state</h2>
<p>The main game loop lives at <code>$8BB1</code>. It is entered from the title interrupt by a
stack-resetting jump the moment fire is pressed, and from then on it runs the game-logic chain — the
object engines, zone checks and state dispatch — and loops. Outside gameplay it waits for the frame
counter to change each time around, but <strong>during play the loop free-runs</strong>: the game-state
check branches straight past that wait, so its pace is set purely by how long one trip through the
engines plus the interleaved raster interrupts takes — roughly <strong>2½ frames per pass</strong>. The
raster handlers themselves stay locked to the frame; only this main loop runs untied to it. That
distinction matters because the character-based actors (tanks, mines, prisoners) are stepped from the
main loop, so their patrol speeds are measured in loop passes, not frames — a tank advances a cell
about every ten frames, not every four.</p>
<p>A single byte at <code>$9D</code> holds the overall game state and selects what that chain does:
<code>1</code> title / attract, <code>9</code> demo game, <code>3</code> new game, <code>4</code> "get
ready", <code>5</code> life lost, <code>2</code> playing, <code>6</code> game over and debrief,
<code>7</code> a transition lock, and <code>$0A</code> the cavern teleport.</p>

<h2>Memory layout in play</h2>
<p>With the ROMs banked the way they are, the 64&nbsp;KB address space is densely packed. Zero page
holds the live state — game state, frame counter, the camera position, the player block and a set of
pointers. The VIC's bank 1 contains the screen at <code>$4400</code>, the sprite shape blocks at
<code>$4000</code> (blocks 1–14 are the enemy helicopter's animation frames), and the two character
sets at <code>$5000</code> (HUD) and <code>$5800</code> (playfield).</p>
<p>The current level is held as a <strong>decompressed map</strong> from <code>$0503</code> — 40 rows
of one page each — beside a soft <strong>scanner bitmap</strong> that backs the radar display, and
small per-object coordinate and state tables for the char-based actors (tanks, prisoners, mines). The
loaded game file itself occupies <code>$7000–$B8FF</code>: the two level maps and their RLE-packed
scanner bitmaps, the HUD screen image, the packed sprite shapes, then the bulk of the code and its
data tables, and finally the raw character-set data. The stage-2 loader and loading screen are left
as dead remnants higher in memory, never referenced again.</p>
`,
    graphics: `
<div class="info-eyebrow">Fort Apocalypse · Graphics</div>
<p>Fort Apocalypse is a character-mapped game: the playfield is built from an 8&times;8 tile set, the
moving actors are a mix of hardware sprites and animated characters, and the levels are stored as
compressed grids of screen codes. None of the data is encrypted — the only transformations applied
to it are a simple run-length scheme, the sprite packing, and a <code>$7F</code> mask on map bytes.</p>

<h2>Compression</h2>
<p>A single decompressor at <code>$8CDB</code> serves all level data. It reads a byte; if that value
appears in the active <strong>run-table</strong>, the following byte is a repeat count (with
<code>0</code> meaning 256) and the value is emitted that many times; otherwise the byte is a single
literal. Two run-tables pick which values are eligible to repeat — one for terrain, a smaller one
(<code>$00 $55 $AA $FF</code>) for the scanner bitmap — so there are no escape codes at all. Every
decompressed byte is masked with <code>AND #$7F</code>, which keeps all map codes below
<code>$80</code>.</p>

<h2>Character sets</h2>
<p>Both character sets are built at init from uncompressed data, copied in overlapping 256-byte
strips. They are swapped mid-frame by the raster handlers, so the HUD and the playfield draw from
different sets.</p>
<h3>HUD set ($5000)</h3>
<p>Selected by <code>$D018 = $14</code> for screen rows 0–6. It holds the score font and the HUD
furniture. Its high characters are left as <strong>soft characters</strong> into which the radar
window is rendered at runtime.</p>
<h3>Playfield set ($5800)</h3>
<p>Selected by <code>$D018 = $16</code> for rows 7–24. It holds the terrain glyphs — 8&times;8
multicolor dither patterns, including the mountain-slope, flat-dither and solid-block tiles. The low
characters <code>$00–$20</code> are reserved as <strong>soft characters animated in place</strong> by
the playfield interrupt: the energy barriers cycle between a stored pattern and blank on a timer; the
laser-grid segments each flip on or off independently and are re-rolled periodically; a four-character
group lights one member per phase to rotate; the explosion character and the fort core are masked
against the SID noise register (<code>$D41B</code>) every frame for a live flicker; the reactor-gate
walls pick one of two solid forms per life; and the missile-exhaust rows are noise-flickered each
frame. The same alphabet glyphs that form the double-width HUD font also serve as object graphics —
distinct glyph ranges are the prisoners, the self-propelled mines, and the tanks and their missiles.</p>

<h2>Sprites</h2>
<p>Fourteen sprite shapes are stored in a <strong>packed column format</strong>: 36 bytes per shape,
arranged as two 18-byte pixel columns (the left column's rows, then the right column's), located by a
pointer table. Init expands each shape into a 64-byte VIC sprite block, laying out <code>[left][right][pad]</code>
per row. The sprites are hi-res — no sprite multicolor — and the player and enemy sprites are
horizontally expanded.</p>
<p>Both helicopters, player and enemy, draw from <strong>one shared animation table</strong> of 18
entries indexed by bank/tilt: seven banking poses &times; two rotor frames, with the level-flight pose
covering three tilt steps. The player toggles its rotor frame every frame; the enemy every fourth
frame. The two bullet sprites are built at runtime from a nine-byte dot pattern — one block carries
the pattern twice for angled shots, the other once for straight-down shots.</p>

<h2>The level maps</h2>
<p>Each level's terrain is decompressed from a per-level source into a buffer at <code>$0503</code> —
one 256-byte page per map row, 40 rows. The map bytes <strong>are screen character codes directly</strong>,
with no tile-index indirection. Two placeholder codes are resolved after decompression: one is replaced
by a random pick from three cave-rock glyphs and another by a different trio, driven by the SID noise,
which gives the cave rock its mottled texture. The two levels are <em>Vaults of Draconis</em> (the
surface, with fuel depots and the landing pad) and <em>Crystalline Caves</em> (the Kralthan fortress,
with its central shaft and a large field of destructible rock).</p>
<h3>A cylindrical world</h3>
<p>The 256-byte rows are wider than the visible playfield. Columns 0–214 hold the 215 columns of level
content, columns 215–254 are padding, and <strong>column 255 is a copy of column 0</strong>. The world
is a horizontal cylinder: the camera column wraps around, and at the wrap point the right edge of the
screen displays that stored copy of the leftmost column, so the world's left edge meets its right edge
without a seam.</p>
<h3>Scrolling</h3>
<p>When the camera advances a full character — or every 8 frames regardless, so that map-embedded
objects keep animating — the engine rewrites the source operands of an unrolled copy loop and
block-copies <strong>16 rows &times; 40 columns</strong> from the map buffer straight to the screen.
Sub-character movement between copies is done with hardware fine-scroll. Because moving objects write
themselves <em>into the map buffer</em>, this periodic re-copy doubles as their on-screen update.</p>

<h2>The scanner</h2>
<p>The radar is backed by a second compressed stream that decompresses per level into a 1600-byte soft
bitmap — the whole map as a 320&times;40-pixel image (40 chars &times; 5 rows). The HUD rows are a
prebuilt screen image whose scanner window is made of soft characters; each frame a 12&times;3-character
window of the bitmap, following the camera, is copied into those characters' definitions. Blips are
XOR-plotted through a pixel-pair mask table — the player every frame, the enemy helicopter and the tank
bases blinking.</p>

<h2>The HUD</h2>
<p>The status display shows the score (six BCD digits), a bonus that counts down during play and is set
to 9999 when the fort is destroyed, the fuel gauge (four BCD digits), the "MEN TO RESCUE" count, and a
message row for flashing texts such as "LOW ON FUEL". The digits are drawn with leading-zero blanking.</p>
`,
    gameplay: `
<div class="info-eyebrow">Fort Apocalypse · Gameplay</div>
<p>Fort Apocalypse is a rescue-and-destroy game: you pilot the Rocket Copter through a surface and a
fortress of caverns, lift out trapped men and blow the enemy's reactor core, against tanks, mines,
homing missiles and a hunting enemy helicopter. Almost every interaction in the game follows from one
unusual rule about what counts as solid.</p>

<h2>The collision model</h2>
<ul>
  <li><strong>Solidity is defined by pixels, not tables.</strong> The core test takes the character
  drawn under an actor and scans its eight charset bytes; any non-zero byte is a hit. So blanking a
  character's definition makes every cell drawn with it non-solid <em>at once</em> — the basis for all
  the dynamic walls and barriers below.</li>
  <li><strong>Character-based actors carry their own collision.</strong> Tanks, mines, missiles and
  prisoners draw themselves into the map buffer (saving the background underneath) and react to the
  character codes they find around them.</li>
  <li><strong>Hardware sprites use the VIC collision latches</strong>, read once per frame —
  sprite-to-sprite and sprite-to-background.</li>
  <li><strong>Bullets bridge the two worlds.</strong> They fly as sprites but stamp an explosion
  character (<code>$20</code>) into the map on impact, and the character actors die from touching it.</li>
</ul>

<h2>The player — Rocket Copter</h2>
<p>Left and right build a <strong>bank</strong> that steers the copter, aims its gun and indexes the
sprite shape so it visibly tilts; up and down move it directly; and gravity pulls it down at a rate set
by the gravity option. The camera keeps the copter within a horizontal band and scrolls the cylindrical
world beneath it. (The title attract mode flies the copter by replaying a recorded joystick sequence.)</p>
<p>Contact with terrain is fatal <em>unless</em> the cell is a legal landing surface — the landing pad,
a fuel depot, the walkway floor, or a prisoner — in which case the copter bounces gently and the spot
becomes the <strong>respawn checkpoint</strong>. Setting down on a fuel depot refuels, the depot draining
visibly as it does. Fuel falls slowly in flight; at zero the engine sputters and "LOW ON FUEL" flashes.
A crash — from enemy or enemy-bullet contact, or a hard landing on an empty tank — sends the copter into
a flashing fall and costs a life; running out of lives ends the game. Brief grace timers protect the
moments just after spawning or teleporting.</p>

<h2>Bullets</h2>
<p>The gun fires from the nose along the current bank angle — from full-left, through level (which fires
<em>straight down</em>), to full-right — using the same bank-to-trajectory mapping as the enemy's gun.
Two impacts are special: the reactor core on level 1 triggers the <strong>fort-destruction sequence</strong>
(an expanding explosion, sixteen colour flashes, a 9999 bonus), and destructible rock is permanently
cleared. Every other hit stamps the explosion character into the cell, and what follows depends on the
victim. Against plain terrain the explosion lingers a few frames, then the original character is restored.
Against an object — a mine, tank, missile or prisoner — the bullet is freed at once and the object's own
engine finds the explosion in its cell, dies, and restores its background. So a direct hit kills any of
them; the sole exception is the enemy helicopter, a sprite that dies through the collision latch instead.
A consequence worth noting: prisoners can be shot, by you or by the enemy.</p>

<h2>The enemy helicopter</h2>
<p>Only one is active at a time. After a delay it spawns at a random patrol point — but never within
roughly 34 columns or 8 rows of the player, so it cannot materialise on top of you. It then hunts by
<strong>pure per-axis pursuit</strong>: each tick it steps one cell toward the player horizontally and
then vertically, with no pathfinding, testing the cells ahead so it only advances into clear corridor.
It banks into its motion — which in turn aims its shots — and fires periodically while on screen. It
cannot chase you across the cylinder's wrap. Off-camera it keeps hunting in map coordinates, with only
its sprite and gun going live once it is back in view, and a watchdog quietly resets it to a fresh patrol
point if it spends too long stuck off-screen while you are underground. Its only exits are death — flying
into terrain, or being hit by a player bullet — after which it explodes and waits to respawn. Its
climbing is notably erratic, incidentally: an easter-egg signature left in the binary overwrote one
opcode in its upward-probe routine, so its ceiling checks read a garbage column and it often stalls or
clips going up.</p>

<h2>Tanks, missiles and mines</h2>
<p>These are the character-based enemies. <strong>Tanks</strong> — six per level — are three body cells
plus a turret that always aims at the player; they patrol horizontally <strong>in lockstep</strong> —
one shared timer steps all six together — turning back only at open air or water while driving straight
through every other kind of terrain, and respawn at fixed home positions once cleared. Each tank can launch one <strong>homing missile</strong>
when the player passes within range above it: the missile flies in its facing direction, steering toward
the player's row, and falls once its fuel runs out, the player slips behind it, or it leaves its column
range — detonating on anything solid. <strong>Self-Propelled Mines</strong> (the manual's name for the
small drones) patrol the corridors in numbers set by difficulty; they spawn at random empty cells, fly
horizontally and reverse at obstacles, and do not respawn once destroyed until the next level. All three
die the same way — an explosion character or a missile in their cell — and because a missile's own
character kills the mines, tanks and prisoners it passes through, missiles can be lured into the other
enemies.</p>

<h2>Prisoners — "men to rescue"</h2>
<p>Eight per level, placed wherever the level builder finds a floor cell with rock directly above. Each
runs back and forth along its walkway. Flying into one within a few cells rescues him: he boards, the
rescued count rises, and the on-screen tally is reprinted. He can also be killed — by shooting away the
floor beneath him, or by an explosion or missile — so a stray shot, yours or the enemy's, can kill the
very men you need. Either way he leaves the "men to rescue" count, and <strong>both level exits stay
locked until that count reaches zero</strong>.</p>

<h2>The dynamic fortress</h2>
<p>None of the fortress's walls, gates and hazards use object slots. The map cells never change; their
character glyphs are <strong>redefined at runtime</strong>, and because solidity is pixel-based,
redefining one glyph opens or closes every cell drawn with it simultaneously. This drives:</p>
<ul>
  <li><strong>Reactor gate walls</strong> — two gates on level 1; at each life one is filled solid and
  the other left passable, chosen at random, so the safe route changes every life. Destroying the core
  opens both for the escape.</li>
  <li><strong>Sweeping walls</strong> — a band of four glyphs of which exactly one is solid at a time,
  advancing in phase so a wall section appears to march along the corridor. Its direction is reversed by
  every shot you fire, anywhere on the map.</li>
  <li><strong>Laser grids</strong> — four glyphs re-rolled every couple of seconds, each independently
  lit or dark at even odds; a lit segment is lethal, a dark one open air. There is no pattern to learn —
  passage is a gamble.</li>
  <li><strong>Energy barriers</strong> — two interleaved groups that are blank except for a brief lethal
  flash each cycle, the two groups flashing in alternation. On level 0 they form diagonal "scissor gates"
  across the cavern passages; on level 1 they are rails and shaft columns. Destroying the fort forces
  them permanently blank.</li>
</ul>
<p>The barriers double as the level-0 transport system. Flying into a lit barrier on level 0 from beneath
a scissor gate does not kill — it <strong>teleports</strong> the copter to one of four random cavern drop
points, each beside another gate, with a grace flag so the arrival cannot crash. On level 1 a barrier
always kills; there the hazard of the gates is the rock around the funnel, not the barrier itself. (Some
walls also carry a purely cosmetic shimmer that never affects collision.)</p>

<h2>Difficulty</h2>
<p>Three options on the title screen tune a run: <strong>Gravity Skill</strong> (how fast the copter
sinks), <strong>Pilot Skill</strong> (the speed of the enemy helicopter, tanks, missiles, barriers and
sweeping walls, plus the number of active mines — 13, 26 or 39), and <strong>Robo Pilots</strong> (three,
five or seven lives).</p>

<h2>Progression and rank</h2>
<p>Two playfields loop with rising difficulty. Clear and rescue the surface — <em>Vaults of Draconis</em>
— then land on the bottom-centre pad and sink through the floor into the fortress, <em>Crystalline
Caves</em>; there, rescue the men and shoot the reactor core, then fly out the top opening. The third pass
is the surface again, harder, and landing back on the base deck ends the mission. The debrief tallies
rescued men and bonuses into a <strong>rank from 0 to 15</strong>, shown as one of four bird names —
Sparrow, Condor, Hawk, Eagle — and a class number from 4 up to 1, with Eagle Class 1 at the very top.</p>
`,
    music: `
<div class="info-eyebrow">Fort Apocalypse · Music</div>
<p>Fort Apocalypse has no separate in-game score. Its one piece of music is the
<strong>loading-screen tune</strong> — the three-voice SID piece that plays under the U.S. Gold
loading screen while the game streams in from tape. That tune is more than music: it is a small
bytecode program whose commands also patch the machine as it plays, including writing the jump that
actually starts the game. It is described in full under <strong>Image &amp; Loader</strong>.</p>
<p>Once play begins, the SID is given over entirely to <strong>sound effects</strong>. The two raster
handlers drive the audio each frame — the copter's engine, gunfire, explosions, and the warning tones —
rather than a continuous melody, so the gameplay itself runs without a backing track.</p>
`,
  },
  turrican: {
    loader: `
<div class="info-eyebrow">Turrican · Image &amp; Loader</div>
<p>Turrican ships on a single double-density Amiga floppy that carries <strong>no filesystem</strong> —
only a boot block the Kickstart ROM will run, and behind it the whole game laid out in a private format
the loader pulls off by absolute sector. The build preserved here is a one-disk cracked release by
Tristar &amp; Red Sector, whose loader decompresses the game from a crunched image on boot.</p>

<h2>The disk image</h2>
<p>An ADF is a flat dump of the floppy's 1760 logical blocks of 512 bytes — block <em>N</em> is simply
the bytes at offset <em>N</em>&times;512. The first four bytes are the <code>"DOS\\0"</code> boot
signature, enough for the ROM to accept the disk and run its boot code, and that is the only
AmigaDOS-conformant thing on it. There is no directory: the program, graphics and levels sit in a
private layout addressed by absolute byte offset through <code>trackdisk.device</code>, never through
files. The disk falls into three regions — the boot block (blocks 0–1), a first-stage loader in plain
68000 code (blocks 2–21), and from block 22 to the end the <strong>crunched main part</strong>: the
entire game — program, graphics and levels — stored compressed.</p>

<h2>The boot block</h2>
<p>The boot block is a complete sector loader. The ROM enters it with the boot device's I/O request
ready; it reads blocks 2–9 to <code>$30000</code> and runs them (the cracker's intro), then clears the
bitplane, copper and sprite DMA. It allocates a work buffer (the largest FAST-memory chunk, or the chip
region on a 512&nbsp;KB machine), issues the main read — about 143&nbsp;KB of the crunched main part to
<code>$50000</code> — and stops the drive. It adapts to the CPU (on a 68010 or better it installs a
<code>TRAP</code> handler running a <code>MOVEC</code>, so the rest of the loader can treat the machine
as a bare 68000), seizes the machine (supervisor mode, interrupts off, stack at <code>$80000</code>),
copies a 512-byte tail routine to <code>$7F800</code> and jumps to it. The boot block never touches
<code>dos.library</code>; it drives the hardware directly.</p>

<h2>The intro and trainer</h2>
<p>Being a cracked release, blocks 2–9 are the cracker's intro: it opens <code>graphics.library</code>
and the Topaz font, allocates a chip buffer for its bitplanes, and scrolls the group's greetings and a
prompt over a copper display. Pressing the joystick button after decrunching enables the
<strong>trainer</strong> — 99 lives — and the high-score save is redirected to track 0. The game itself
appears only once the main part is decrunched.</p>

<h2>Decrunching the main part</h2>
<p>The tail at <code>$7F800</code> is the bridge from loader to game: it calls the decruncher at the head
of the crunched blob, then enters the game. The crunched main part is not one packed stream but the
output of <strong>three compressors applied in series</strong>, so unpacking runs three decoders in turn
— a Huffman bit-reader, then an LZ77 copier, then an RLE expander — each relocating its intermediate
result to the top of memory and decoding it back down. Two of the three are byte-dispatched: they build a
256-entry jump table whose default handler copies a literal and whose few escape values trigger
match/run handlers, and the loop <strong>writes each control byte to the background-colour register</strong>
(<code>$DFF180</code>) as it runs — the flickering colour bars across the screen that a cracked game
shows while it "decrunches". The result is a
214,400-byte image at <code>$43880</code>, with the game entered partway into it at <code>$5F500</code>.
The tail then applies the trainer (overwriting two longwords with branches into the cheat code) and jumps
into the decrunched game.</p>
`,
    engine: `
<div class="info-eyebrow">Turrican · Game Engine</div>
<p>The decruncher hands control to a flat image at <code>$43880</code>. The first thing the game does is
split itself in two and bring up the hardware; from there it is a <strong>vertical-blank-driven loop</strong>
with a function-pointer state machine, pulling each world's code and data off the disk as it goes.</p>

<h2>Two segments</h2>
<p>The image does not run where it is loaded. On entry the game copies the <strong>resident engine</strong>
— roughly 112&nbsp;KB — down to low memory from <code>$10</code> onward, where <code>$10–$FF</code> is the
68000 exception vector table and <code>$100</code> is the engine's internal jump table. The rest of the
image — the setup code, the interrupt handler and data — runs in place. So the program is two segments:
the relocated resident engine (the bulk of the game) and the in-place setup/ISR.</p>

<h2>Bring-up</h2>
<p>Entry reads the fire buttons to pick the trainer and option settings, waits for a press and release,
and branches into <code>game_init</code> — the hardware bring-up. It enters supervisor mode, turns all
interrupts and DMA off, then unpacks and runs several sub-modules, installs the level-3 (vertical-blank)
interrupt vector, and enables the display. The vblank interrupt bumps a frame counter, cycles the
palette, and calls the per-frame game tick.</p>

<h2>The resident engine</h2>
<p><code>game_init</code>'s last act jumps into the relocated segment through its internal jump table;
slot 0 is <code>game_start</code>, which seizes the machine, runs the OS-interface module, initialises
the object table, the map grid and the display interrupt, and falls into the main loop. The engine
re-uses the <strong>same three-pass decoder</strong> at runtime to unpack graphics and level blocks,
alongside a PowerPacker decompressor and a floppy trackloader that streams level data off the disk during
play.</p>

<h2>The streamed modules</h2>
<p>The engine does not ship complete in the resident image; three more modules stream in at startup:</p>
<ul>
  <li>the <strong>music / sound driver</strong> (see Music);</li>
  <li>a <strong>loader-sound player</strong> that installs its own vblank handler and plays the
  disk-access sound during loading;</li>
  <li>a PowerPacker-compressed <strong>OS-interface module</strong> — the engine's bridge to the system:
  it opens <code>graphics.library</code> and <code>dos.library</code>, installs a <code>TRAP</code>
  handler and saves and replaces CPU vectors for the display and disk.</li>
</ul>

<h2>The game loop</h2>
<p><code>game_start</code> falls into level setup, which clears the playfield with the blitter, installs
three triple-buffered display buffers, primes the level state and runs a chain of subsystem inits, then
drops into the game loop. Two things define its shape:</p>
<ul>
  <li><strong>Mode dispatch.</strong> A single pointer holds the current game-mode handler, called once
  per frame. Swapping it switches state — title, play, and so on — without touching the surrounding
  pipeline: the classic function-pointer state machine. It is driven by a <strong>scene system</strong>:
  a scene id indexes a descriptor table, and the descriptor's handler fields become the primary and
  secondary per-frame handlers. The descriptors are not in the resident image — they are
  <strong>streamed off the disk per world</strong>, so the states and their code change with the level.</li>
  <li><strong>Frame sync.</strong> The loop raises a flag and spins until the vblank interrupt clears it,
  locking the pipeline to the vertical blank.</li>
</ul>
<p>Around the mode call sits the fixed pipeline: blitter copies that draw the playfield and object layers
from a draw list, plus a dozen further per-frame subsystems. The engine also carries its own copy of the
sound driver on resident state, so it runs the music and the sound effects as two independent player
instances.</p>
`,
    graphics: `
<div class="info-eyebrow">Turrican · Graphics</div>
<p>The engine and its modules are only the loader and runtime; the <strong>worlds themselves are streamed
off the floppy</strong> as the game runs. Each is a self-describing block of tiles, a palette, a map, a
collision layer and sprite graphics, decoded straight into memory.</p>

<h2>Worlds streamed off disk</h2>
<p>Loading a level reads a per-world entry from a table, pulls the packed block off the disk into a buffer
just past the resident image, and decompresses it with the same three-pass decoder used at boot. Each of
the five worlds decodes to a fixed block of about 260&nbsp;KB at a known address. A small
<strong>section directory</strong> at the head of the block points at the tile data, a collision layer,
the 16-colour palette, the sprite/object graphics, and a TFMX music slot — so a single decoded block
carries everything that world needs.</p>

<h2>Tiles and palette</h2>
<p>The palette is 16 big-endian 12-bit RGB words. Tiles are reached through an offset table — a list of
longword byte-offsets, with entry 0 equal to the table's own size, so the tile count and the start of
the pixels both fall out of it. Each tile is <strong>32&times;32 pixels in four bitplanes</strong>,
interleaved per row (512 bytes), drawn through the palette. World 0 has 209 tiles, world 1 has 215, and
so on — the cave and planet surface, the machine world, and the rest.</p>

<h2>The tile map</h2>
<p>Each world holds several <strong>scenes</strong>, one per sub-map, each described by a descriptor: a
pointer to the map data, its width and height in tiles, and the scene's per-frame handlers. The map is a
<strong>column-major array of one byte per cell</strong>; a value below the tile count is a tile index,
and a value at or above it is the same tile drawn <strong>horizontally flipped</strong>. World 0's three
scenes are 137&times;51, 153&times;51 and 115&times;51 tiles, laid out back to back, while other worlds
are shaped very differently — one world opens on a tall 12&times;269 vertical shaft.</p>

<h2>Collision</h2>
<p>Solidity is not a parallel grid; it is a <strong>per-tile-type shape</strong>. A collision section gives
16 bytes per tile — a 4&times;4 grid of 8&times;8-pixel-block solidity — so each 32&times;32 tile carries
sub-tile collision. The values are not merely solid or empty: passable, solid wall or ground, solid but
reacting to shots (a hit sparks and stops), breakable or trigger (contact spawns an effect and clears the
cell), and hazard (contact drains the player's energy). At scene load the playfield builder copies each
map tile's shape into a screen-sized collision buffer, and the player check reads one byte at the player's
position at 8-pixel granularity; flipped map cells mirror their columns. This is the layer the viewer's
<strong>Collision</strong> toggle overlays.</p>

<h2>Sprites — the BOB format</h2>
<p>Enemies and effects are <strong>blitter objects</strong> cookie-cut into the back buffer: a
four-bitplane bitmap and a one-plane mask blitted through the playfield's 16-colour palette, with plane 3
doubling as the mask so opaque pixels carry colours 8–15 and colour 0 is transparent. Each is described by
a 14-byte <strong>descriptor</strong> — bitmap pointer, mask pointer, modulo, a packed size word, and a
y-adjust and flag — and a flat array of these descriptors is the animation table: an object's draw routine
picks a frame group, then the current frame within it, then draws that descriptor. The <strong>player</strong>
is the exception — he is drawn by a dedicated routine as a multi-part composite (three body parts plus the
orbiting spinning weapon), indexed by his animation state.</p>
`,
    music: `
<div class="info-eyebrow">Turrican · Music</div>
<p>Turrican's score is Chris Hülsbeck's, in his own <strong>TFMX</strong> format — Turrican is the
canonical TFMX game. The music is driven by a dedicated sound overlay, and the engine carries
<strong>two copies of the same player</strong> so the music and the sound effects run independently.</p>

<h2>The sound driver</h2>
<p>The music and sound driver is a separate module streamed off the disk at startup, decoded with the same
three-pass decoder as everything else and loaded at a fixed address. Its body opens with a
<strong>branch dispatch table</strong> — its public API, which the engine calls at fixed entry points to
start playback, initialise the player from the song and sample pointers, and set the master volume and
channel mask. Its vertical-blank entry runs the player once per frame: it processes the voices, each with
a period LFO (vibrato), a pitch slide (portamento) and a volume envelope, writing the Amiga's audio period
and volume registers, while a silence call zeros all four channel volumes. The engine keeps a second,
byte-identical copy of this player on its own state, so the music and the effects play as two independent
instances.</p>

<h2>The TFMX module</h2>
<p>The score itself is a TFMX module. It is <em>not</em> played from the per-world scene block — that
block's "TFMX-SONG" slot is an empty stub — but from the sound overlay, which carries the in-game player
and two data pointers: the song data and about 50&nbsp;KB of raw signed 8-bit PCM, the instrument samples.
The song data is a set of tables — a song table of start, end and tempo per sub-song (three real ones), a
pattern pointer table, a macro pointer table, and a trackstep table that lays out the eight channels. A
<strong>pattern</strong> is a stream of note-plus-macro entries and commands; a <strong>macro</strong> is a
stream of instrument commands (set sample, volume, period, vibrato, portamento, envelope, DMA, wait, loop)
— effectively a small instrument VM. The player runs a song sequencer and trackstep processor feeding a
pattern reader and a per-voice update with the macro VM, driving Paula's four channels.</p>
`,
    gameplay: `
<div class="info-eyebrow">Turrican · Gameplay</div>
<p>Turrican is a run-and-gun platform shooter across five large, multi-scene worlds. Crucially, the worlds
differ in their <strong>enemies and backdrops, not their mechanics</strong> — every world runs on the same
engine, object system and sound interface, bringing only its own enemy roster and scene code. The parts
that are not self-describing data — the objects, where they spawn, and how they behave — are driven by
code.</p>

<h2>The object system</h2>
<p>Active enemies and effects are a <strong>doubly-linked list</strong> of 58-byte nodes drawn from a
39-node pool; spare nodes sit on a free list. A spawn pops a free node, fills its fields and links it into
the active list; a kill unlinks it back to the pool. Each node holds its position, its current frame and
frame table, an active flag and an AI-handler pointer. Every frame the engine walks the list once calling
each node's handler, then walks it again drawing each node — cookie-cutting its sprite through its frame
table at the current frame and position. So a spawn is simply a node whose frame table points at one of
the world's sprites.</p>

<h2>Enemy behaviour and per-world code</h2>
<p>Each world's scene block carries its own code in two parts. The <strong>scene handlers</strong> run the
animated parallax backdrop and trigger ambient sounds — they only call the resident sound API, never the
spawn routines; world 1's drives the waterfall, and worlds differ here only in which backdrop they animate.
The <strong>enemy-AI handlers</strong> — six to eighteen per world — are the enemy roster: each is a
complete behaviour on an object node, setting its sprite and health, animating a loop, applying damage and,
on death, freeing the node. The per-world differences are the enemies and the backdrop, not new mechanics.</p>

<h2>Enemy placement</h2>
<p>Which enemy is seeded where is read by a <strong>scroll-triggered spawner</strong>, called twice per
frame. It builds a spawn window from the camera — the visible screen plus a margin — and spawns any entry
inside it, which is why enemies appear <em>just</em> as the screen reaches them. The layout is a 2D bucket
grid, not a flat list: a per-camera-row offset table and the grid yield, for a given camera position, a
pointer into the entry data; each distinct pointer heads a run of 6-byte entries (type, x, y in 8-pixel
units) ending at a terminator. An entry's <strong>type</strong> selects its handler in two tiers — low
types index a resident handler table (engine-wide objects like the little rotating mine), higher types
index the scene's own handler table — and the handler installs the object's sprite. So the chain from a
placement entry to a drawn enemy is <strong>type → handler → sprite</strong>.</p>

<h2>Starting position</h2>
<p>Each scene also records its initial camera tile and the player's on-screen offset, so the player spawns
at camera-plus-offset — the scene's intended starting position, which is the point the viewer frames each
scene on.</p>

<h2>Weapons</h2>
<p>Turrican's signature weapon is the <strong>spinning energy beam</strong>: holding the fire button deploys
it, and while it is active the player can sweep it through its 32 rotation angles but cannot move, before it
releases in a short burst. Its sprite is one of the shared resident sprites — the same 32-frame rotation
plus burst — rather than a per-world enemy graphic, so it is available in every world.</p>
`,
  },
  marble: {
    loader: `
<div class="info-eyebrow">Marble Madness · Image &amp; Loader</div>
<p>Marble Madness ships on a single Amiga floppy that boots through entirely <strong>stock AmigaDOS</strong>:
the disk is an ordinary bootable filesystem, with no custom fastloader on the boot path. The protection
is elsewhere: the main program is encrypted, and a from-scratch track loader reads it off the platter by
physical position.</p>

<h2>The disk and filesystem</h2>
<p>An ADF is a flat dump of the floppy's 1760 blocks of 512 bytes. The disk is a normal AmigaDOS
<strong>OFS volume</strong> named <code>MarbleMadness!</code>, with a standard boot block, a root block and
three directories holding 50 files. Almost every file is an Amiga loadable <strong>hunk object</strong> — a
relocatable code/data segment AmigaDOS brings in with <code>LoadSeg</code> — so the game is not one binary
but a launcher, a main program, and a large set of per-course overlays loaded on demand. Two files are
exceptions: the main program and a helper named <code>xxx</code> are stored <strong>encrypted</strong>,
near-random at the byte level, and decrypted at load.</p>

<h2>Booting to Workbench</h2>
<p>The boot block is the unmodified AmigaDOS boot code: it finds the resident <code>dos.library</code> and
hands off to it. DOS runs a two-line startup script — <code>LoadWb</code> then <code>endcli</code> — that
brings up Workbench (the desktop is a service of the Kickstart ROM; the disk only bundles
<code>icon.library</code> and its icon files so it can show its own window on a bare machine). The player
launches the game by double-clicking the marble icon, which <code>LoadSeg</code>s and runs the launcher.</p>

<h2>The launcher and the encrypted program</h2>
<p>The launcher is a small compiled program that, finding it was started from Workbench rather than a shell,
displays the splash screen and brings the game in. Because the main program is encrypted, that load does not
go through plain <code>LoadSeg</code> — it goes through a small decryptor named <code>zzz</code>, a custom
<strong>decrypting <code>LoadSeg</code> replacement</strong>. It reads the encrypted file, undoes a keystream
XOR, relocates the hunks, and hands the segments back. The cipher is keyed in part to machine state, so the
decryption is bound to a particular Kickstart and a booted process, not just the disk (see Game Engine).</p>

<h2>A from-scratch track loader</h2>
<p>The other encrypted file, <code>xxx</code>, is decrypted first and run as code: it is a custom floppy
<strong>track loader</strong>. It reads neither through AmigaDOS nor through <code>trackdisk.device</code>'s
normal commands — it drives the floppy hardware directly: the CIA drive-control and status ports to seek and
check readiness, and Paula's disk DMA to pull a whole raw MFM track in one burst, then MFM-decode and
validate each sector in the CPU. It reads the main program by <strong>physical track and sector position,
not by name</strong> — the filename never appears in the launcher at all. The program still exists as a real
DOS file (so the disk stays a valid bootable volume and its blocks are laid down contiguously), but it is read
by location for speed and as a copy-protection hook: a from-scratch reader can demand non-standard formatting
and bypass file-level tampering.</p>
`,
    engine: `
<div class="info-eyebrow">Marble Madness · Game Engine</div>
<p>Reaching the game's own code means getting through the encryption and the copy protection wrapped around it.
Once decrypted, the program is a stripped Amiga hunk object — almost entirely code — that drives the hardware
directly and runs as two cooperating tasks.</p>

<h2>The multi-stage load</h2>
<p>The launcher loads the game in stages over a shared control block: it decrypts the track loader
<code>xxx</code> with the decryptor <code>zzz</code> (with an empty key), runs <code>xxx</code> to read the
175&nbsp;KB main program off the disk by physical position, then mutates the key and runs <code>zzz</code>
again to decrypt the program. The track loader is the fast raw reader; the decryptor is the cipher; they
cooperate through the shared block, and the key changes between the two decrypt passes.</p>

<h2>The cipher</h2>
<p>The on-disk format is a standard AmigaDOS hunk with <strong>selective encryption</strong>: the hunk-header
magic, the block-type markers and the symbol names are left in plaintext, so the file's structure stays
legible, while the hunk sizes, relocation tables and code/data bodies are XORed with a keystream — one
keystream longword per stored longword. There is no compression; the bodies are full size and the high entropy
is the cipher. The keystream is an additive <strong>lagged-Fibonacci generator</strong> over a 55-entry table
built from a fixed seed by a multiply-hash.</p>

<h2>The copy protection</h2>
<p>The teeth are in the key setup. Before the keystream runs, the table is perturbed by folding in two pieces of
<strong>live machine state</strong>: bytes from the host's CPU exception and TRAP vector table in low memory,
and the running task's exception- and trap-handler pointers. Because those entries feed the generator, the
keystream past its first stretch depends on the vector table and the task — which is why the file's structure
decodes regardless (its keystream is drawn before the perturbed entries propagate) while the bodies scramble.
On Kickstart 1.x every exception vector points at the same ROM handler, so the relevant byte is just the ROM
page, tying the protection to the 1.x ROM layout; the handler pointers only exist once AmigaDOS has constructed
the launcher's process, so the full key is not present until the game is actually booting on the right vintage
of machine. This is why such titles are Kickstart-version-locked: the decryption key is, in effect, the ROM
page.</p>

<h2>Inside the program</h2>
<p>Decrypted, the main program is a stripped hunk load file — not merged into a few segments but <strong>347
hunks</strong> (about 115 object modules), each keeping its own code/data/BSS triple, with no symbol or debug
blocks. It is mostly code: the bitmaps and samples live in separate per-course files, so the program carries no
pixel or sample data — only the engine that drives them, and it drives them at the metal, writing the full
blitter register block and <code>DMACON</code> directly and reading the mouse ports. The small data payload
is the UI text (the course banners, "GAME OVER", the player labels), the per-course level filenames it loads at
run time, and a few lookup tables.</p>

<h2>Two cooperating tasks</h2>
<p>The running game is two contexts that talk over exec messages: a <strong>main thread</strong> (the Intuition
front-end and the game-state machine) and a separate vblank-synced <strong>"Framer" task</strong> that owns the
display refresh. There is no single linear loop — the gameplay update and the display refresh run in different
contexts. The Framer task wakes once per vertical blank, animates the cycling colours, and rebuilds the
copper/display list when a fresh world frame is ready; the main thread runs the state machine (set up the
course, then play), which each frame integrates every object and draws it.</p>
`,
    graphics: `
<div class="info-eyebrow">Marble Madness · Graphics</div>
<p>Marble Madness's graphics are <strong>blitted, not sprited</strong>: the program draws everything itself from
per-course banks of tiles and obstacle cells, scrolling a single tall course vertically. The boot screen uses
standard Amiga formats; the per-course art uses one shared RLE codec.</p>

<h2>The splash screen</h2>
<p>The title screen is a standard IFF ILBM bitmap — 320&times;200, four bitplanes (16 colours), its pixels
<strong>ByteRun1 (PackBits)</strong> compressed, with a palette and four colour-cycling ranges so parts of the
logo animate. A small boot-screen overlay loads the image and puts it up while the game streams in.</p>

<h2>Tiles</h2>
<p>Each course's floor, walls and railings — everything the marble rolls on — are a <strong>tile map</strong> in
its own file, a single PackBits stream. Unpacked, it holds a 16-colour palette, four bitplanes of tile graphics,
and a tilemap. Tiles are 8&times;8 pixels in four bitplanes; the tilemap is a row-major stream of tile-index
words, a constant <strong>36 tiles (288 px) wide</strong> — Marble Madness scrolls only vertically, so the width
is fixed and the height varies per course. Placing each tile by index reproduces the whole course. The map's
leading word is the <strong>playable</strong> height; rows stored beyond it can never scroll on screen and serve
as hidden storage — Ultimate keeps <strong>three extra variants of its final screen</strong> there, which the
engine's tile-repaint machinery cycles through in play so the narrow paths to the goal appear and disappear
(the map view replays this as a tile animation, collision heights and all in the real game). The stored palette
is only the top of the course: each Track carries a <strong>display block</strong> of copper colour bands that
recolour six palette slots per vertical stretch of the map — the courses fade as the marble descends (Aerial
turns from tan to pure red to yellow; Beginner's ice pit is emerald). The map view bakes these bands in as
recoloured tiles. On Ultimate the same block also drives a three-phase <strong>gold glitter</strong> — the copper
rotates three gold shades through the checker pattern every four frames — which the map view plays too. (The
tilemap is only the visual surface; the physics rolls the marble on a separate height field.)</p>

<h2>Obstacle cells</h2>
<p>Each course also carries a bank of <strong>obstacle sprites</strong> — the goal flag, moving barriers,
drawbridges and the like — also one PackBits stream, holding a count and a table of cell descriptors over
contiguous planar pixel data. Each cell is one complete animation frame, in one of two layouts chosen by a type
byte: "stored" free sprites (the flag, the marble, the creatures) keep their bitplanes row-interleaved — the
Amiga hardware-sprite data layout — and carry a per-object colour ramp, while "composited" scenery is sequential
plane blocks whose bitplanes the loader ORs together into a <strong>silhouette mask</strong>, a black-and-white
image of that piece of level geometry used to occlude sprites behind it. The moving creatures and the marble
itself live in separate banks that share this container.</p>

<h2>The course layout</h2>
<p>A third per-course file holds everything else a course needs — not just object positions but all of its
gameplay data. It is a plain hunk module loaded at course init, opening with a header of relocated pointers the
engine fans out to the actor-system globals: the static slope field, a placement/feature table, the coarse-zone
partition, the animation scripts, the creature spawn lists and the actor list.</p>
`,
    gameplay: `
<div class="info-eyebrow">Marble Madness · Gameplay</div>
<p>The marble is a real <strong>3-D simulation</strong> projected to the isometric view — rolled by the
<strong>mouse</strong> over a height-mapped course, with a state machine governing rolling, falling, landing and
the dizzy spin.</p>

<h2>The mouse and the marble</h2>
<p>Input is the mouse's quadrature counters, per player — a <strong>relative</strong> device, so moving
faster pushes harder — accumulated into a roll-force. The marble is not a 2-D sprite but a <strong>point
mass</strong> with velocity and position in three dimensions; each frame the engine integrates position by
velocity and then iso-projects to the screen, so the isometric look is a projection of a real 3-D model. Exactly
three things write the marble's velocity: the scaled mouse force; friction and an octagonal speed cap
(clamped per surface — that selector is the ice/grating friction); and the surface force from the terrain.</p>

<h2>The course as a height map</h2>
<p>The terrain is two independent systems. The course itself is a <strong>static slope field</strong>: a list of
region records, each a flat isometric-tile rectangle carrying a base height, a direction and a one-dimensional
profile, rasterised at load into a <strong>corner-height mesh</strong> — a grid of cells each holding the four
corner heights of one tile. All the regions compose into one continuous 2.5-D height map, and the triangular
slope faces you see are emergent: a quad with non-coplanar corners is two triangles. Each frame the engine
samples the four mesh cells around the marble, picks which of the tile's two triangles it is over, computes the
surface gradient, and accelerates the marble down it — except on the <strong>Silly</strong> course, which adds
instead of subtracts, so the marble rolls <strong>up</strong> the slopes instead of down. The walls fall out of the same mesh: a height step between
neighbouring cells becomes a side the velocity is clamped against. One height map drives both the roll and the
walls, with no per-cell terrain codes.</p>

<h2>Scripted regions</h2>
<p>A few moving or interactive surfaces — seesaws, the rail-guarded holes, the start and finish triggers, the
ball-catcher — are a separate system: a per-course list of regions whose reference point is emitted by a small
<strong>bytecode animation script</strong> (keyframes, a move opcode for a sliding slope or seesaw, sound
triggers). Each frame the engine matches a region to the marble's tile and dispatches on its terrain code: slope
codes push the marble toward the reference point, trigger codes raise wall flags that snap and bounce it.</p>

<h2>The marble's state machine</h2>
<p>All of this runs under a <strong>twelve-state machine</strong> on the marble. Three states are
player-controllable — they run input and physics — while the rest are animation or transition states that only
redraw the marble: rolling, landing after a drop, an edge reaction, falling and settling onto a surface, an
object-bump on contact with an enemy, the course-intro run, the spawn, and the hole/region capture. A notable one
is <strong>dizzy</strong>: a survivable hard hit (by another marble) or fall is <em>not</em> death — it sets a
stun flag, and the rolling state hands off to a swirl-spin that plays out and returns to rolling. Death is
running off the edge onto no terrain at all, <strong>falling from too great a height</strong>, or the hazards
and the marble-munchers.</p>

<h2>Actors</h2>
<p>The moving things — the goal flag, the enemies, the munchers — are <strong>actors</strong> fed by the
course-layout data. Each frame the engine walks an array of actor records, each holding a sprite-cell pointer,
an animation-script pointer (a cell list advanced when a frame timer expires, with randomised durations) and a
position. The whole moving cast rides the Amiga's <strong>hardware sprites</strong>: the "stored" cell format —
16 pixels wide, two bitplanes, row-interleaved — <em>is</em> the sprite-DMA data layout, copied into a sprite
channel's buffer each frame, and every piece carries its own three sprite colours that the copper loads
mid-screen. Eight channels cover everything by <strong>copper multiplexing</strong> down the screen; wider
pieces use several 16-pixel columns. The display is a fixed 512-pixel-tall bitmap used as a <strong>circular
scroll buffer</strong>: as the course scrolls vertically the visible window wraps around that buffer. (The
course itself does not wrap — only the scroll buffer does.)</p>

<h2>Hiding behind the level — the mask punch</h2>
<p>The marble can roll <em>behind</em> parts of the course — under the raised drawbridge, into holes — yet the
tiles are plain 8&times;8 squares that can hold background and foreground in one tile, and the hardware priority
is fixed with every sprite <em>in front of</em> the playfield. The engine's answer is beautifully direct:
<strong>it erases the occluded pixels from the sprite itself</strong>. Every scenery piece is a cell from the
obstacle bank, and at load the engine builds each cell's <strong>silhouette mask</strong> — a black-and-white
image of that chunk of level geometry. Each frame, after the marble's sprite data is queued, the engine checks
the marble's position against the course's occluding features (the drawbridge, the holes, the funnels); if the
marble is behind one, it takes the piece's mask for its <em>current</em> animation state, shifts it to the exact
pixel offset, inverts it, and ANDs it into the marble's sprite words — punching the scenery's shape out of the
marble so the playfield shows through, pixel-perfectly. Sprite-versus-sprite layering falls out of the hardware
(lower channels appear in front), ordered by an isometric depth sort. The "level" pieces themselves — the bridge
plank, the flag poles, Practice's pop-up start ramp — are cells from the obstacle bank anchored by the
course-layout data; the map view's <strong>scenery overlays</strong> toggle draws exactly these pieces, placed by
replaying the same data: the goal flags land on the GOAL banner of every course, in their record's own sprite
colours.</p>
`,
    music: `
<div class="info-eyebrow">Marble Madness · Music</div>
<p>Each course has its own theme — in fact two per course, fourteen tunes in all — and the notable thing is how
they play. Unlike the bare-metal C64 and Turrican drivers that bang Paula's registers directly, Marble Madness
plays its music through the operating system's <strong><code>audio.device</code></strong>: it sequences the song
and, per voice, hands the OS a sample pointer, a period and a volume, letting <code>audio.device</code> perform
the actual Paula DMA. So the player is much simpler than a TFMX-class engine — no macro VM, no software mixing,
just "play this sample at this pitch and volume".</p>

<h2>Where the music lives</h2>
<p>The six per-course sound banks are ordinary AmigaDOS hunk modules of <strong>pure data</strong> — there is no
player code in the file; the player lives in the main program. Their data blocks carry the chip-memory flag (the
only memory Paula can DMA from) and hold a song header, instrument and envelope tables, and 8-bit signed PCM
sample waveforms. The song names its instruments by relocated pointer rather than by index.</p>

<h2>The sound engine</h2>
<p>The player runs as its own exec <strong>task</strong>, and the rest of the engine talks to it only by message
— a clean producer/consumer split. It is <strong>dual-clocked</strong>: an audio-reply clock advances each voice's
note list (when <code>audio.device</code> finishes a voice's sample it replies, and the player writes that
voice's next note), and a 60&nbsp;Hz timer clock ticks the per-voice pitch and volume envelopes on sustained
notes. The engine is a general sampled-sound driver — music is just a set of voices it triggers — addressed by
sound number through a per-bank directory; a course's music is one of those entries, whose event lists are the
long, looping score.</p>

<h2>The song format</h2>
<p>The music itself is a <strong>Soundtracker-style arrangement</strong>: an order table per channel of
<code>(repeat, pattern)</code> entries — play a pattern so many times, then advance, with a zero repeat looping
back to the start — over patterns that are byte-streams of note commands (a note as octave-and-semitone, a rest,
a note-length class, an end marker). Each note is voiced from the bank's <strong>single shared waveform</strong>:
the semitone indexes the standard Amiga/ProTracker period table for the fine pitch, and the octave selects the
length of the looped waveform slice — the classic one-sample-many-octaves trick. A per-note volume envelope (a
list of rate/target segments ramped one step per frame) gives each note its pluck shape rather than a flat tone.
Advancing the score is a third, separate clock: a per-frame music tick driven by the video frame at about
50&nbsp;Hz steps through the patterns, emitting notes into the same voice path — distinct from the driver's
60&nbsp;Hz envelope timer and its sample-reply note clock, each doing its own job.</p>
`,
  },
  sml: {
    loader: `
<div class="info-eyebrow">Super Mario Land · Image &amp; Loader</div>
<p>Super Mario Land ships on a 64&nbsp;KB Game Boy cartridge with a simple bank-switching chip. There is no
loader to speak of: the console's boot ROM only runs the cartridge after it has <strong>verified the
cartridge itself</strong>, and then jumps into the game's own cold-start code.</p>

<h2>The cartridge image</h2>
<p>The image is four 16&nbsp;KB ROM banks behind an <strong>MBC1</strong> mapper. Bank&nbsp;0 is permanently
visible at <code>$0000–$3FFF</code>; banks&nbsp;1–3 share the switchable window at <code>$4000–$7FFF</code>,
selected by <em>writing</em> the bank number to ROM space (<code>$2000–$3FFF</code>) — an MBC1 register, not a
memory store. The cartridge declares no save RAM, so there is no battery-backed high-score table. The header at
<code>$0100–$014F</code> carries the entry point, the Nintendo logo, the title <code>SUPER MARIOLAND</code> in
the old 16-byte form, and two checksums. The boot ROM refuses to start a cartridge whose logo isn't byte-exact
and whose header checksum doesn't match — both pass here, which is what lets the game run at all.</p>

<h2>Cold start</h2>
<p>The entry point at <code>$0100</code> jumps to the cold-start routine at <code>$0185</code>. It disables
interrupts, then enables only the VBlank and STAT sources (the timer interrupt that drives sound stays off for
now). To clear video RAM — which can only be touched while the screen is off — it runs an <strong>LCD
"safe-off" dance</strong>: turn the LCD on, wait for a known scanline, then switch it off. It sets the
palettes, switches the sound hardware on, sets up the stack, and clears work RAM, video RAM, sprite memory and
high RAM. It copies a 12-byte sprite-DMA routine into high RAM — the transfer has to be kicked from there
because the CPU can't see ROM while it runs — and initialises the sound engine by paging in bank&nbsp;3 and
calling it.</p>

<h2>Interrupts and the main loop</h2>
<p>Three interrupt vectors do the real-time work. <strong>VBlank</strong> (<code>$0060</code>) is the render
half of each frame. <strong>STAT</strong> (<code>$0095</code>) is a mid-frame raster split that holds the status
bar still. The <strong>timer</strong> (<code>$0050</code>) runs the sound engine on its own clock, independent
of the video frame. Once init finishes turning the LCD back on and enabling interrupts, the game settles into a
main loop that runs the frame's logic and then <code>HALT</code>s until the VBlank handler sets a frame-done
flag at <code>$FF85</code> — so one trip round the loop is exactly one displayed frame.</p>
`,
    engine: `
<div class="info-eyebrow">Super Mario Land · Game Engine</div>
<p>Super Mario Land is a <strong>frame-synced state machine</strong>. A single state byte selects the whole
behaviour of the frame through a jump table, and every frame splits into a logic half and a render half.</p>

<h2>The state machine</h2>
<p>The current state is a byte at <code>$FFB3</code>. Each frame the main body loads it and executes
<code>RST $28</code> — a one-byte gateway that jumps through a <strong>62-entry word table</strong> to the
handler for that state. The flow runs boot settle &rarr; title screen (which polls for a newly-pressed Start)
&rarr; level load &rarr; in-level gameplay, with further states for death and transitions. Adding behaviour to
the game is adding a state and its table entry.</p>

<h2>Two halves of a frame</h2>
<p>Game logic runs during the visible scan, while the screen is being drawn; rendering happens only in VBlank,
the brief window when video RAM and sprite memory are writable. The <code>$FF85</code> flag interlocks the two
so the loop never races the display.</p>

<h2>Input</h2>
<p>The joypad register <code>$FF00</code> is read in two passes — the d-pad, then the buttons — and the result
is stored as two bitmaps in high RAM: the buttons currently <em>held</em> (<code>$FF80</code>) and the buttons
<em>newly pressed</em> this frame (<code>$FF81</code>). Handlers test the edge-triggered byte, so a jump or a
menu choice fires once per press rather than every frame the button is down.</p>

<h2>Bank shadowing and the status bar</h2>
<p>Because the sound engine and the level data both live in the switchable bank window, the engine follows a
strict <strong>save / switch / call / restore</strong> pattern around cross-bank calls, with the active bank
shadowed in high RAM so the timer interrupt — which always pages in bank&nbsp;3 for sound — can put it back
afterwards. The <strong>status bar</strong> is simply the top rows of the background map held still while the
playfield scrolls underneath: VBlank resets the scroll to zero for the bar, and the STAT handler waits for
H-blank at the split line and reloads the playfield's scroll value. (The Game Boy's window layer, often used for
a status bar, is here the <em>pause</em> overlay instead.)</p>
`,
    graphics: `
<div class="info-eyebrow">Super Mario Land · Graphics</div>
<p>The graphics are ordinary Game Boy tiles, but the level is <strong>streamed a column at a time</strong> from
a compressed map, and every enemy is a metasprite assembled on the fly from 8&times;8 tiles.</p>

<h2>Tiles and the screen</h2>
<p>Every graphic is a 2-bits-per-pixel tile, 16 bytes holding two bitplanes. The background is a 32&times;32
tilemap scrolled past the 160&times;144 screen; sprites are entries in the hardware sprite table, copied each
frame by DMA from a shadow buffer in work RAM. Super Mario Land runs in <strong>8&times;8 sprite mode</strong>
and uses the signed tile-addressing mode for its background art.</p>

<h2>The level map</h2>
<p>A level is a list of screen pointers — an <strong>order table</strong> — and the same screen pointer can
appear at many positions, which is where repeated stretches of terrain come from. Each screen is 20 columns of
<strong>run-length-encoded</strong> tiles: a run byte packs a starting row and a count, the next bytes are tiles
placed downward, and fill and end-of-column markers finish it. The low order-table indices are reserved — a
lead-in screen, then the two pipe-accessed bonus rooms — so the main path begins at the third entry. The map is
decoded straight from ROM; nothing is rasterised ahead of time.</p>

<h2>Column streaming</h2>
<p>As the camera scrolls, a builder decodes the next map column into a work-RAM buffer and blits its 16 tiles
into the background map just off the right edge of the screen. A handful of tile ids — pipes
(<code>$70</code>) and the breakable and question blocks (<code>$80</code>/<code>$5F</code>/<code>$81</code>)
— are normal tiles that are <em>also</em> recorded as interactive blocks in side tables as they are laid down.</p>

<h2>Metasprites</h2>
<p>Each object carries a frame id that indexes a pointer table to a <strong>turtle-graphics stream</strong>:
control bytes move an 8&times;8 cursor and set the sprite attribute, and high-bit bytes stamp a single tile at
the cursor. A facing flag picks between two mirror-image layout tables, which is how an enemy faces left or
right from the same stream. The object tile art is bulk-loaded into video RAM per world, so each world brings
its own cast of enemy graphics into the same tile region.</p>
`,
    music: `
<div class="info-eyebrow">Super Mario Land · Music</div>
<p>The music plays on the Game Boy's four-channel sound chip — two square-wave voices, one wavetable voice and
one noise voice — driven by a sound engine in bank&nbsp;3 that runs on a <strong>hardware timer, not the video
frame</strong>, so the tempo is independent of how busy the screen is.</p>

<h2>The sound engine</h2>
<p>The timer interrupt pages in bank&nbsp;3 and calls the engine on every tick, 64 times a second. The engine
services a set of request slots in work RAM: per-channel slots for sound effects, and one
<strong>music selector</strong> at <code>$DFE8</code>. Writing a song id there starts a piece of music; the
engine then advances that song's channels on each tick.</p>

<h2>The song format</h2>
<p>A song is a header — a master byte, a duration table, and four channel pointers, one per voice. Each channel
pointer leads to an <strong>order list</strong> of pattern pointers ending in a loop target: the patterns before
the target play once as an intro, then playback loops from the target onward. A pattern is a short bytecode —
set-voice (a volume envelope and a duty), set-duration (an index into the duration table), repeat-the-previous
note, end-of-pattern, or a note number that reads its pitch from a frequency table. A duration unit is one tick,
1/64&nbsp;s.</p>

<h2>One theme per level group</h2>
<p>Music is chosen per level by a small table indexed by the level number, so each piece covers several levels:
the bright overworld-style theme plays in 1-1, 1-2 and 3-1; the underground theme in 1-3, 3-2 and 3-3; the Muda
(water) theme in 2-1 and 2-2; the Chai theme in 4-1 and 4-2; a tense boss-and-vehicle theme in the
auto-scrolling 2-3 and 4-3; and a short jingle in the pipe bonus rooms. The tracks in this viewer are named for
the levels that table assigns them to.</p>
`,
    gameplay: `
<div class="info-eyebrow">Super Mario Land · Gameplay</div>
<p>Super Mario Land's mechanics — where enemies appear, how they behave, what counts as solid ground, the pipe
warps, and Mario's own movement — are driven by code and small per-level tables rather than self-describing
data.</p>

<h2>The object system</h2>
<p>Up to ten enemies and effects are live at once in a bank of object slots. A <strong>scroll-triggered
spawner</strong> walks a per-level placement list — 3-byte entries giving a trigger column, a packed
row-and-fine-position, and a type — sorted by column, spawning each entry as its column scrolls in off the right
edge of the screen. That is why an enemy appears <em>just</em> as the screen reaches its position.</p>

<h2>Behaviour scripts</h2>
<p>Each object type's behaviour is a <strong>script in a small bytecode interpreter</strong>. Its opcodes set a
velocity, coast for a number of frames, flip the facing, set the animation frame, spawn a child object,
transform into another type (or despawn entirely), queue a sound effect, gate on how close the player is, and
loop. A Goomba is a two-frame walk loop; world&nbsp;4's high-numbered types are the Tatanga boss fight, where one
type spits projectiles and another fans out a spread of them. Items that pop from a block all share the same toss
arc.</p>

<h2>Collision</h2>
<p>Solidity is a <strong>pure tile-id threshold</strong> — there is no separate collision map. An actor reads
the background tile underneath it and treats any id at or above a fixed value as solid floor; lower ids are
passable scenery (sky, clouds, palms, the decorative pyramids and statues); the very top range is special
metadata that is never floor. The tileset is laid out in that order on purpose, so the test is a single
comparison.</p>

<h2>Pipes and bonus rooms</h2>
<p>A pipe is a <code>$70</code> tile whose destination is recorded in a <strong>parallel metadata map</strong>
that shadows the background. Pressing Down while standing on one reads the destination, animates Mario sliding
in, and re-points the screen index at one of the reserved bonus-room screens; leaving the room repositions him
back where he entered. A "bonus room", then, is just another screen entry the engine warps to and returns
from.</p>

<h2>Mario's movement</h2>
<p>Mario is a <strong>special object on the same velocity integrator as the enemies</strong>, flagged so that
moving him also scrolls the camera. Jump height is variable through a hold timer — a tap gives a short hop, a
held button a full jump — and the B button makes him run. Gravity accelerates the fall a pixel per frame at a
time, and dropping below the bottom of the playfield is a death. He starts each level at the same fixed
on-screen position, which is the point this viewer frames every level on.</p>
`,
  },
  stuntcar: {
    loader: `
<div class="info-eyebrow">Stunt Car Racer · Image &amp; Loader</div>
<p>Stunt Car Racer ships on a single double-density Amiga floppy that carries <strong>no filesystem</strong> —
a boot block just conformant enough for the Kickstart ROM to run it, and behind it the whole game laid out
in a private, sector-addressed format the loader pulls off by hand. There is no directory and there are no
files: the program, the tracks and the physics tables are all located by following the loader.</p>

<h2>The disk image</h2>
<p>An ADF is a flat dump of the floppy's 1760 blocks of 512 bytes — block <em>N</em> is simply the bytes at
offset <em>N</em>&times;512, in the usual <strong>80 cylinders &times; 2 heads &times; 11 sectors</strong>
layout. The first four bytes are the <code>"DOS\\0"</code> boot signature, enough for the ROM to accept the
disk and run its boot code, and that is the only AmigaDOS-conformant thing on it. The boot block names a root
block, but there is no valid root header there and no directory — the disk is <strong>custom-formatted</strong>,
a flat region of code and data only the game's own loader understands. It falls into three raw regions: the
boot block (sector 0), a self-contained loader (sectors 22&ndash;97), and from sector 110 to the end the whole
game — engine, tracks and physics — read verbatim into <code>$E700</code>.</p>

<h2>The boot block</h2>
<p>The boot code is a compact track loader. The ROM enters it with the boot device's I/O request ready; it
stops multitasking (<code>Forbid</code>), allocates 38&nbsp;KB of cleared chip RAM, and issues one read —
the 38&nbsp;KB blob at disk offset <code>$2C00</code> (sector 22) into that buffer — retrying on error, then
stops the drive, resumes (<code>Permit</code>) and jumps into the blob. The boot block does nothing
game-specific beyond fetching that blob; the loaded code is the real bootstrap.</p>

<h2>The custom loader</h2>
<p>The 38&nbsp;KB blob is position-independent 68000 code — every reference is PC-relative, so it runs wherever
it is placed. It does four things. It <strong>seizes the machine</strong>: a <code>TRAP&nbsp;#0</code> drops into
supervisor mode, then it clears interrupts and DMA, installs its own vertical-blank autovector and re-enables the
master and VERTB interrupts. It loads an initial stage to <code>$4000</code>. It <strong>shows the title screen</strong>,
unpacking a four-bitplane image into chip RAM at <code>$78000</code> with its palette and a Copper list. Finally it
<strong>streams the game in</strong>: 805 sectors starting at sector 110 into <code>$E700</code>, sets the user and
supervisor stacks, and jumps to <code>$E700</code>.</p>

<h2>The disk reader</h2>
<p>Behind the loader is a logical-sector reader, not a filesystem: given a start sector, a count and a destination,
it validates the range against the 1760-block disk, converts a logical sector to a physical track (11 sectors per
track-side) and MFM-reads whole tracks. Crucially it <strong>only reads</strong> — there is no decompression — so
on disk everything is stored raw and sector-aligned. That is why the game code at <code>$E700</code> is directly
executable the moment it lands, with no unpacking step.</p>
`,
    engine: `
<div class="info-eyebrow">Stunt Car Racer · Game Engine</div>
<p>The loader streams a 412&nbsp;KB engine to <code>$E700</code> and jumps in. The first bytes are not code
reached by falling through — they are the entry, a small supervisor trampoline, and an embedded Copper list —
so the engine is best read by following the transfers out of the entry. From there it is a
<strong>vertical-blank-paced loop</strong> that never touches the operating system.</p>

<h2>Self-check and supervisor</h2>
<p>The very first thing the engine does is <strong>checksum itself</strong>: it sums roughly 408&nbsp;KB of the
image as 16-bit words and compares the result against a stored value — a tamper and load-integrity guard. It then
installs its own <code>TRAP&nbsp;#0</code> handler to drop into supervisor mode, sets the supervisor stack, and
branches to two routines in turn: the hardware bring-up and the game bootstrap.</p>

<h2>Hardware bring-up</h2>
<p>Bring-up (<code>$ED56</code>) is a textbook bare-metal Amiga take-over. It masks interrupts, clears and then
re-enables <code>INTENA</code> (master, vertical-blank and the audio levels), turns on <code>DMACON</code> for the
bitplanes, Copper, blitter, sprites, disk and audio, and wires the seven 68000 interrupt levels to the engine's own
handlers — the level-3 vertical-blank handler drives the frame timing. It sets up the CIA timers, points the Copper
at the embedded display list, and primes the audio buffers. From here on the engine owns the machine and drives the
hardware registers directly; it never calls <code>graphics.library</code> or <code>dos.library</code>.</p>

<h2>The in-place decryptor</h2>
<p>Before the game runs, bring-up walks a <strong>~46&nbsp;KB region</strong> (<code>$F4B8</code>&ndash;<code>$1AA4A</code>)
and <code>EOR</code>s every byte with <code>$80</code> in place. That block is stored <strong>obfuscated</strong> on the
disk and only becomes real 68000 code once this pass rewrites it — a second layer of protection on top of the
self-checksum, and the reason a static read of that range is meaningless until it is unscrambled. Everything outside
it, including the bootstrap and the main loop, is plain on disk.</p>

<h2>The main loop</h2>
<p>The bootstrap clears its state tables and working screen, initialises the lookup tables, and enters the top level.
From there the engine is a conventional loop locked to the vertical blank: a <strong>state word</strong> selects the
current screen — title, menu, track-select or race — and swapping it switches mode without disturbing the surrounding
pipeline. Two double-buffered colour/Copper tables are exchanged each frame, and the level-3 handler paces everything.
The filled-vector race renderer, the track interpreter (see Graphics) and the car simulation (see Gameplay) all hang
off the race state of this loop.</p>
`,
    graphics: `
<div class="info-eyebrow">Stunt Car Racer · Graphics</div>
<p>The tracks are the game's centrepiece and they are <strong>3-D vector geometry, not bitmaps</strong>: each of the
eight circuits is a short byte stream — a few hundred bytes — that expands into a complete elevated ribbon of road,
which the engine draws as filled polygons every frame. The compactness is the whole trick: an entire rollercoaster
of a track is well under a kilobyte.</p>

<h2>Eight circuits from a byte stream</h2>
<p>The eight tracks — Little Ramp, Stepping Stones, Hump Back, Big Ramp, Ski Jump, Draw Bridge, High Jump and Roller
Coaster — are reached through a pointer table, each 124&ndash;213 bytes long. A short header gives the
<strong>section count</strong> and the start/finish section; then comes a run of per-section records. The stream is
<strong>run-length-encoded</strong>: a marker nibble repeats the previous section with one parameter stepped by a
fixed amount, which is how a long constant curve or a steady gradient is stored in a couple of bytes.</p>

<h2>The 16&times;16 grid footprint</h2>
<p>Each section carries its cell on a <strong>16&times;16 track grid</strong>, packed into one byte (low nibble the
grid X, high nibble the grid Y). Consecutive sections are always <em>adjacent</em> cells, so the plan view is a
closed loop built with no accumulation and no drift — the true footprint of the circuit seen from above, whether a
diamond, a rectangle, or a lap doubling back along an inner peninsula.</p>

<h2>Reusable piece shapes</h2>
<p>A section's <strong>type</strong> indexes a shared table of <strong>piece shapes</strong> — straights, mirror-image
arcs, ramps — so a handful of shapes describe every track. Each piece shape holds two things: the local
<code>(x,z)</code> outline of the rung strip (a straight marches its rails forward; a curve carries its real arc), and
a per-rung <strong>height profile</strong> for the left and right rails. The section's grid cell places the piece in
the world; the shape supplies its geometry.</p>

<h2>Surface, ramps and camber</h2>
<p>Along a section the road is a strip of <strong>rungs</strong>, and each rung has an <em>independent</em> left and
right rail height. That single fact carries the whole surface: a smooth run of heights is a drivable slope or a hill,
while a sudden jump between rungs is a <strong>hard edge</strong> — a ramp lip, a stepping-stone gap, a ski-jump
launch. The mean of the two rails is the road's elevation and their difference is its <strong>banking</strong>
(camber), so a bend can lean while it climbs. Nothing decides "step versus slope" separately; it is entirely in what
the profile values do.</p>

<h2>The baked model and the render</h2>
<p>At race start the engine <strong>bakes</strong> the section stream into the polygon model it will draw: per-rung
records of the left and right rail positions and heights, placed at absolute coordinates from each section's grid
cell. The bake <strong>decimates</strong> — it keeps a cross-road polygon edge only where the shape marks one or the
height changes sharply — reverses the pieces that run backwards, hides the open ends of gap pieces, and tags the
start/finish. Each frame the renderer re-places these baked records for the current camera and fills them as
flat-shaded polygons through the blitter: the elevated ribbon on its support walls, drawn as solid vectors rather
than sprites.</p>
`,
    gameplay: `
<div class="info-eyebrow">Stunt Car Racer · Gameplay</div>
<p>What gave Stunt Car Racer its reputation is that the car is a <strong>sprung rigid body, not a point</strong>: the
chassis pitches and rolls on its suspension, the wheels gain and lose contact over crests and on landings, a landing
that is too hard damages the car, and how it handles depends on speed and the road's gradient. All of that is one
per-frame simulation.</p>

<h2>The simulation frame</h2>
<p>A physics tick is a <strong>semi-implicit Euler rigid-body step in fixed point</strong>, run <em>twice</em> per
displayed frame — once for the player's car and once for the opponent — at a fixed 50&nbsp;Hz. A <code>0.93</code>
damping factor applied at both integration stages is the drag that keeps the sim stable. Forces are summed in the
car's <strong>body frame</strong>, rotated into the world through a 3&times;3 orientation matrix rebuilt each tick
from the roll, yaw and pitch, and integrated to velocity and then position. <strong>Gravity</strong> is a fixed
world-down vector re-expressed in the tilted body frame every tick, so it always pulls straight down whatever the
car's attitude.</p>

<h2>Suspension on the real surface</h2>
<p>The car sits on a <strong>three-point suspension</strong>. Each point's spring compression is the track surface
height under it minus the chassis contact height minus a rest length; the force is a spring-plus-damper of that
compression. The chassis contact heights follow the car's height tilted by its roll and pitch — and the
<strong>track surface heights come straight from the track geometry</strong>: the simulation locates the section the
car is over, finds where each wheel sits across the rung strip, and interpolates the same per-rung rail heights the
renderer draws. So the springs literally ride the decoded ramps and bumps, and a wheel running off the lateral edge
is detected. The three springs combine into a net lift, a roll torque and a pitch torque, and set the on-ground
flag that the rest of the model keys off.</p>

<h2>Drive, grip and steering</h2>
<p>Drive force comes from the throttle and gear, decays under wheelspin, and is clamped to the available
<strong>grip</strong> — tyre load times two, and only when grounded, so there is <em>no</em> grip in the air. The
lateral tyre force is grip-limited the same way, flagging a <strong>slide</strong> when the demand exceeds grip: a
friction-circle model. A drag opposes motion, stiff when grounded and speed-proportional when rolling free. An
<strong>auto-steer</strong> measures the car's heading against the section's centreline and nudges it to follow the
track.</p>

<h2>Damage</h2>
<p>When a spring force exceeds its tolerance and stays there for a run of frames, <strong>damage</strong> accumulates
on a 0&ndash;255 scale, with a separate counter for hard bottoming slams — the "land too hard and you wreck the car"
mechanic that the whole track design plays against.</p>

<h2>A copy-protection trap in the physics</h2>
<p>Buried in the steering code is a piece of the disk <strong>copy protection</strong>. It compares a value against a
constant and, on mismatch, <strong>zeroes the pitch-stabilisation torque</strong> — quietly degrading the handling
rather than refusing to run. Both operands are obfuscated so they cannot be searched for: the address is built from a
subtraction, the magic value from a sum. The compared value is not a setting — it is written by a routine that reads
the <strong>physical disk hardware</strong> (the CIA drive-control and disk-DMA registers), so it is derived from the
disk's protection tracks. On a genuine disk the check passes; on a plain copy it fails and the car handles a little
worse — an anti-piracy nudge, not a lock. The controls that feed all of this are a single decoded joystick byte:
throttle and brake, steer left and right, and fire.</p>
`,
  },
  mariokart: {
    loader: `
<div class="info-eyebrow">Mario Kart DS &middot; Image &amp; Loader</div>
<p>Mario Kart DS runs from a 32&nbsp;MiB Nintendo&nbsp;DS Game Card. A DS card is not a memory-mapped
cartridge like the Game&nbsp;Boy or Game&nbsp;Gear: it is a serial device behind a controller, and its
contents are an on-card <strong>filesystem</strong> read on demand. And the DS is a <strong>two-processor
machine</strong> &mdash; an <code>ARM9</code> and an <code>ARM7</code> that share one main memory &mdash; so
the card carries two independent code images that boot side by side.</p>

<h2>Two processors, one memory</h2>
<p>The card hands the two CPUs <em>different</em> load and entry addresses. The <strong>ARM9</strong>
(an ARMv5TE core) loads to <code>$02000000</code> at the base of main RAM and starts at <code>$02000800</code>;
the <strong>ARM7</strong> (an ARMv4T core) loads to <code>$02380000</code>, also inside main RAM but near its
top. Both share the single 4&nbsp;MiB main memory <code>$02000000</code>&ndash;<code>$023FFFFF</code>: the
ARM7's code simply occupies a region the ARM9 agrees to leave alone. Each CPU also has fast private
tightly-coupled memory, its own BIOS, and its own view of the DS I/O block at <code>$04000000</code>. The
ARM9 owns the card and drives the game and both screens; the ARM7 services sound, the touchscreen, buttons
and wireless. Both cores run 32-bit <strong>ARM</strong> and 16-bit <strong>Thumb</strong> code, switched with
<code>BX</code>/<code>BLX</code> &mdash; boot is ARM, but much of the game is Thumb to fit the 32&nbsp;MiB.</p>

<h2>The card header</h2>
<p>The first 512 bytes are the header the DS BIOS reads. It stamps the title <code>"MARIOKARTDS"</code>, the
game code <code>"AMCP"</code> (DS card, Mario&nbsp;Kart, PAL region) and a device-capacity code of
<code>$08</code> that decodes to 32&nbsp;MiB. It then fully describes where each processor starts: ARM9 image
at card offset <code>$4000</code> (loading to <code>$02000000</code>, entry <code>$02000800</code>), ARM7 at
<code>$141600</code> (loading to and starting at <code>$02380000</code>), plus the filesystem tables, an
overlay table, and three independent checksums (a header CRC over <code>$000</code>&ndash;<code>$15D</code>, a
secure-area CRC, and the Nintendo-logo CRC the BIOS validates before it runs any code).</p>

<h2>The filesystem</h2>
<p>Everything that is not boot code &mdash; models, textures, tracks, UI, the sound bank &mdash; lives in a
card filesystem addressed by two tables. The <strong>FAT</strong> is a flat array of <code>{start, end}</code>
offset pairs, one per file; the <strong>FNT</strong> is the directory tree that gives those numbered files
names like <code>data/Course/airship_course.carc</code>. Of the 610 file slots the first four are the ARM9
overlays (referenced directly, with no name); the remaining 606 are named files. A file is <em>not</em> the
asset, though: a single path is three nested layers &mdash; <strong>FAT slice &rarr; LZ77 stream &rarr; NARC
archive &rarr; NITRO resource</strong> &mdash; each unpacked in turn (the graphics chapter picks this up).</p>

<h2>Booting the ARM9</h2>
<p>The ARM9 comes up first. Its startup stub kills interrupts (<code>IME&nbsp;=&nbsp;0</code>), then programs
the <strong>CP15</strong> system coprocessor: invalidate caches, set the MPU regions, and enable the two
tightly-coupled memories the core relies on &mdash; <strong>ITCM</strong> at <code>$01FF8000</code> for fast
code and <strong>DTCM</strong> at <code>$027E0000</code> for fast data and the stacks. The image on the card
is <em>compressed</em>: the bulk of the ARM9 is packed with Nintendo's <strong>BLZ</strong>, an LZSS variant
that decodes <strong>backward</strong> and expands the image <em>in place</em> &mdash; the stub decompresses
itself, growing from <code>$020E751C</code> to <code>$021773D8</code> before it can even reach the code it is
about to call. It then runs the autoload list (copying initialised data and the ITCM/DTCM-resident code into
place), zeroes <code>.bss</code>, cleans the caches, and jumps to <code>main</code>, whose real entry is the
game init at <code>$020365F0</code>.</p>

<h2>Booting the ARM7</h2>
<p>The ARM7 image is <em>not</em> compressed, so its startup is shorter: interrupts off, clear its private
64&nbsp;KiB WRAM at <code>$03800000</code>, park the exception stacks at the top of that WRAM, run its own
autoload and <code>.bss</code> clear. Notably, the address it finally branches to &mdash; <code>$037F8534</code>
&mdash; is <em>in WRAM</em>, not in the <code>$02380000</code> load region: the ARM7 <strong>relocates its hot
code into WRAM</strong> and runs it from there, close to the sound and input hardware it drives.</p>
`,
    engine: `
<div class="info-eyebrow">Mario Kart DS &middot; Game Engine</div>
<p>Past boot, the game is a large C++ program that dispatches almost everything through function pointers and
virtual tables. It lays out the 4&nbsp;MiB main memory, brings up the NitroSDK operating-system layer, and
then &mdash; before it does anything else &mdash; reaches for the second processor.</p>

<h2>The runtime memory map</h2>
<p>The ARM9 partitions main RAM by convention: its decompressed static code and data occupy
<code>$02000000</code>&ndash;<code>$0216F340</code> (with <code>main</code> at <code>$02003000</code> and the
game init at <code>$020365F0</code>), its zero-initialised <code>.bss</code> runs up to
<code>$021804E0</code>, and the four code <strong>overlays</strong> bank in at
<code>$021804E0</code> onward, with the heap growing above them. The ARM7 image sits near the top at
<code>$02380000</code>. Each core keeps its working set in fast private memory &mdash; the ARM9's stacks and
hot data in <strong>DTCM</strong> (<code>$027E0000</code>), fast code in ITCM &mdash; and a small
system-reserved block at <code>$027FF000</code> holds the configuration both processors share.</p>

<h2>Initialisation</h2>
<p>The game init at <code>$020365F0</code> is a compact sequence of framework brings-up: an OS/system
initialiser that walks a list of subsystem constructors, then two routines that each program a contiguous
block of 16-bit hardware registers with <code>STRH</code> bursts &mdash; the 2D-graphics-engine and DMA/timer
register banks. Tick and system-config readers in the <code>$0200Exxx</code>/<code>$0200Fxxx</code> range are
the game's copy of the NitroSDK OS layer; one of them computes an address into the shared
<code>$027FFxxx</code> config block and reads the value both CPUs agreed on.</p>

<h2>Interrupts and the IPC FIFO</h2>
<p>The most telling thing about early boot is how <em>few</em> hardware registers it touches before it needs
the second CPU &mdash; only five, and every one is about inter-processor communication. It clears its
<code>IPCSYNC</code> mailbox (<code>$04000180</code>), enables the 64-bit <code>IPCFIFO</code> hardware
queue (<code>$04000184&nbsp;=&nbsp;$C408</code>), and enables <strong>exactly one interrupt source</strong>:
bit&nbsp;18, <em>"IPC receive FIFO not empty."</em> The ARM9's entire initial interrupt architecture exists to
hear the ARM7. The model underneath is standard DS: <code>IME</code> the master switch, <code>IE</code> the
per-source mask, <code>IF</code> the write-1-to-clear request latch, with the BIOS vectoring an IRQ through a
handler pointer the runtime installs in DTCM.</p>

<h2>The two processors meet</h2>
<p>Having enabled that one interrupt, the game init calls an OS routine that <strong>blocks on
<code>IPCSYNC</code> waiting for the ARM7</strong>. The spin loop reads the four-bit value the <em>other</em>
CPU posted, compares it to an expected step number, and polls until they match &mdash; the
<code>OS_SyncWithOtherProc</code> handshake, in which each CPU ratchets a short sequence of step numbers so
neither races ahead during boot. The ARM9 has done its half and waits; the ARM7, meanwhile, is booting from
<code>$02380000</code> and relocating into WRAM. Only once they meet do the two <code>main</code> routines
begin exchanging FIFO commands &mdash; input and touch state up from the ARM7, sound and DMA requests down
from the ARM9.</p>

<h2>Overlays and the frame loop</h2>
<p>Past the rendezvous the ARM9 enters its frame loop and streams the code <strong>overlays</strong>: each is
a BLZ-compressed blob named by a file ID, loaded to a fixed RAM address with its <code>.bss</code> zeroed and
its constructors run. Because they all load to the same address they are mutually exclusive banks &mdash; only
one resident at a time &mdash; so switching game mode is card-DMA the overlay in, decompress, clear, construct,
jump. The battle- and mission-mode objects, for instance, live in an overlay that pages in with those
modes.</p>
`,
    graphics: `
<div class="info-eyebrow">Mario Kart DS &middot; Graphics</div>
<p>The DS renders the tracks on dedicated 2D and 3D graphics engines, and the card stores their art in the
NitroSDK's family of <strong>NITRO</strong> formats. Reaching the pixels means peeling three layers &mdash;
compression, archive, then the resource itself.</p>

<h2>Asset layers</h2>
<p>A <code>.carc</code> file is a Nintendo <strong>LZ77</strong> stream (a length/distance back-reference
format, distinct from the backward-BLZ used for the boot code) wrapping a <strong>NARC</strong> archive, and
the NARC in turn holds the individual NITRO resources: models, texture sets, animations, collision. A single
track archive therefore expands into its course model, its far-backdrop model, its map-object models, its
texture animation and its collision, all as sibling sub-files.</p>

<h2>Textures and palettes</h2>
<p>A texture set is a <code>TEX0</code> block: two dictionaries &mdash; one naming textures, one naming
palettes &mdash; over packed image and colour data. The DS supports <strong>seven texture formats</strong>:
paletted 4/16/256-colour, two alpha-plus-index formats (A3I5, A5I3), 16-bit direct colour, and a
<strong>4&times;4 block-compressed</strong> mode whose texel indices and per-block palette selectors live in
two separate regions that tile the block back-to-back. Palettes are 15-bit <strong>BGR555</strong>. The
DS's texture units are tiny &mdash; often 32&times;32 texels &mdash; so this viewer keeps them
nearest-filtered to preserve the original pixel art.</p>

<h2>Tiled 2D graphics</h2>
<p>The menus and HUD use the classic tile pipeline: an <code>NCLR</code> palette, an <code>NCGR</code> tile
sheet (4- or 8-bit), and an <code>NSCR</code> screen map whose entries pick a tile, a palette row and
horizontal/vertical flips. Screens and their tiles are split across a base archive and a per-language archive,
so a composed picture is the union of both.</p>

<h2>3D models</h2>
<p>A model is an <code>NSBMD</code> file: a tree of joint <strong>nodes</strong> (each with a
translate/rotate/scale, rotations sometimes stored as a compact pivot form), a stream of scene bytecode that
walks the nodes and binds materials, and per-shape <strong>GX display lists</strong> &mdash; the same command
stream the DS's geometry engine consumes. The display-list interpreter handles every vertex form the hardware
offers, including <em>delta</em> vertices (small signed offsets from the previous vertex) and
<strong>quad strips</strong>, whose winding pairs each new vertex against the one two back. Materials carry
the texture's repeat/mirror/flip flags and, when present, a texture <strong>scale/rotate/translate</strong>
that maps a shared texture onto part of a surface &mdash; a kart tyre's single strip texture, say, stretched
across the tread.</p>

<h2>Track scenes</h2>
<p>Course geometry is stored small and scaled up: vertices are divided by a power-of-two at authoring time and
the scene bytecode re-applies it, and the whole model lives at <strong>one-sixteenth of world scale</strong>
&mdash; the size the physics and course-map data use. Each track ships a companion far model that is its
<strong>skybox</strong>: a backdrop dome drawn <em>relative to the camera</em> so it turns as you look around
but never gets closer. The archive also carries the track's <strong>map objects</strong> &mdash; item boxes,
trees, pipes, Chain&nbsp;Chomps, Piantas &mdash; which the engine spawns onto the course; some of them are
two-triangle <strong>billboards</strong>, flat sprites kept turned to face the camera, exactly as sprites in a
3D world are drawn.</p>

<h2>Scrolling and rippling surfaces</h2>
<p>The moving water, waterfalls and boost-panel arrows are <strong>texture animation</strong>, stored as a
<code>BTA0</code> resource: per material, five tracks &mdash; two scale, one rotation, two translation &mdash;
each either a constant or a series of samples taken every fourth frame. Driving those tracks into the texture
matrix each frame is what scrolls a river downstream and sweeps the arrows across a dash panel; nothing about
the geometry moves.</p>
`,
    music: `
<div class="info-eyebrow">Mario Kart DS &middot; Music</div>
<p>All of the game's sound lives in one 3.3&nbsp;MiB file, <code>sound_data.sdat</code> &mdash; the NitroSDK
sound archive. Every note of music is <strong>sequenced</strong>, MIDI-style: there is no streamed audio at
all. The retail archive is also shipped with its <em>symbol block stripped</em>, so nothing on the card names
its own tracks; they are known only by number.</p>

<h2>The sound archive</h2>
<p>The archive holds 284 files, bound together by an information block: <strong>76 sequences</strong>
(<code>SSEQ</code> &mdash; the music and jingles), <strong>99 instrument banks</strong> (<code>SBNK</code>),
<strong>104 wave archives</strong> (<code>SWAR</code>), and a handful of sound-effect archives. An info record
gives each sequence its bank and master volume; each bank names up to four wave archives its instruments draw
samples from.</p>

<h2>Sequences</h2>
<p>A sequence is a compact bytecode stream, MIDI-like: a byte below <code>$80</code> is a note-on carrying a
key, a velocity and a duration; other opcodes rest, change program (instrument), set pan, volume, expression,
transpose, pitch bend and vibrato, or set the tempo. A sequence opens by declaring its tracks and can
<strong>call, return and jump</strong>; a backward jump is the loop point that makes a 40-second race theme
play forever. Multiple tracks advance in parallel off the one stream.</p>

<h2>Instruments and waves</h2>
<p>An instrument is either a <strong>sampled wave</strong> (a reference into a wave archive, with a base note,
an attack/decay/sustain/release envelope and a pan), one of the DS's two <strong>PSG</strong> channels &mdash;
a Game&nbsp;Boy-style pulse wave with selectable duty, or an LFSR noise generator &mdash; or a composite: a
<em>drum kit</em> mapping each key to its own instrument, or an eight-way <em>key split</em>. The waves
themselves are stored as 8- or 16-bit PCM or as <strong>IMA-ADPCM</strong> (four-bit deltas expanded from a
seed), each with its own sample rate and loop point.</p>

<h2>Playback</h2>
<p>Playing a sequence means running its bytecode against a clock of <strong>48 ticks per quarter note</strong>
(so tempo is real beats-per-minute), stepping every voice's envelope at the sound driver's 192&nbsp;Hz frame
rate, resampling each instrument's wave for the note's pitch, and mixing the voices with constant-power pan.
The envelopes work in decibels &mdash; attack curves the level up toward full, decay falls to the sustain
level, release fades to a silence floor that is exactly the noise floor of 16-bit audio. This viewer plays the
rendered sequences directly.</p>
`,
    gameplay: `
<div class="info-eyebrow">Mario Kart DS &middot; Gameplay</div>
<p>Everything about a track that is not its geometry &mdash; where a lap begins and ends, the line the
computer racers drive, where items steer, where Lakitu drops you, and every object placed on the course &mdash;
is one data file, the <strong>course map</strong> (<code>NKM</code>). The engine reads it at load and builds
the live race from it.</p>

<h2>The course map</h2>
<p>The map is a header followed by seventeen sections, identified purely by position, each a fixed-size record
list: object placements, routes, start and respawn points, checkpoints, the lap graph, the item line, the
computer drive line, trigger areas and cameras. Coordinates are world units &mdash; sixteen times the scale of
the display model, which is why the drive line, laid over the track, sits exactly on the asphalt.</p>

<h2>Laps and checkpoints</h2>
<p>A lap is read from a chain of <strong>checkpoint gates</strong>. Crossing gates advances your position; the
gate keyed as the lap line closes a lap, and gates keyed above it are <strong>key checkpoints</strong> you must
cross <em>in order</em> for the lap to count &mdash; which is precisely the game's shortcut protection, and why
cutting the course does not register a lap. Each checkpoint also names the respawn point Lakitu uses if you
fall out along its stretch.</p>

<h2>Computer racers and items</h2>
<p>The computer karts follow the <strong>enemy drive line</strong> &mdash; a polyline of points, each with a
lateral radius saying how far a racer may wander from it and a drift hint for corners. Item routing (red
shells) steers along a parallel <em>item line</em>. Both, and the lap graph, can <strong>branch</strong>: a
section can name several successors, which is how alternate routes are encoded &mdash; a plain circuit is one
section end to end, while Waluigi&nbsp;Pinball's line splits into ten around the bumpers.</p>

<h2>Placing the objects</h2>
<p>Each object placement names an object by ID plus a position, rotation and scale. The engine matches the ID
against a table that binds it to a model and the callbacks that build it, then spawns the instance &mdash; and
the spawn transform is <em>per type</em>: an object's setup code may adjust it. The <strong>item boxes</strong>
are the clear example: they are placed at road height, and it is the item box's own setup that lifts each one a
fixed <strong>twelve world units</strong> into the air, which is why they float. Ground objects &mdash; trees,
crates, pipes &mdash; sit as placed. A handful of placements are logic-only markers with no model (sound
emitters, effect points) and are silent in the scene.</p>

<h2>Moving objects</h2>
<p>An object whose placement names a route walks that route &mdash; a stored polyline of points &mdash; at a
steady speed, facing the way it travels, wrapping on a closed loop or reversing on an open path. This is what
sends the Cheep&nbsp;Cheep&nbsp;Beach crabs scuttling across the sand, circles Desert&nbsp;Hills' sun over the
course, drives Airship&nbsp;Fortress's Bullet&nbsp;Bills along their track, and swings Waluigi&nbsp;Pinball's
iron balls through the table.</p>
`,
  },
  elite: {
    loader: `
<div class="info-eyebrow">Elite · Image &amp; Loader</div>
<p>Elite for the C64 loads from cassette, and the tape is as much a <strong>copy-protection device</strong>
as a loader. The game is split across a dozen tape segments in a custom turbo format with no checksums, and
the loader rewrites its own wire format between blocks — so the bytes can only be read by running the loader
the tape itself carries.</p>

<h2>The tape image</h2>
<p>A TAP file records the length of every pulse the datasette delivers. The tape opens with two ordinary
CBM ROM-format segments — a header naming the file <code>ELITE</code> and a 289-byte boot program — then
switches to a custom turbo encoding for the dozen segments that carry the game, its graphics and its data.
The ROM-format part is the bootstrap; the turbo part is everything else.</p>

<h2>The bootstrap and autostart</h2>
<p>The first two segments load through the C64's own ROM tape loader. The trick is the boot program's
<strong>load address</strong>: the KERNAL saves the IRQ vector while it uses the tape and restores it from a
fixed pair of bytes when the load finishes — and the boot file is positioned so that those two bytes become
the IRQ vector, pointing into the just-loaded code. So the first timer interrupt after "FOUND ELITE" jumps
straight into the boot code, with no <code>RUN</code> needed. The boot code fills memory, turns off the
KERNAL messages, and jumps into the turbo loader.</p>

<h2>The turbo loader</h2>
<p>From there the tape uses a custom one-bit-per-pulse encoding — a short pulse is 0, a long pulse is 1,
told apart by a CIA timer restarted on every cassette edge. A byte is nine pulses (a start bit and eight data
bits), and the stream is a pilot tone, a sync byte, then blocks <strong>back to back with no checksums and no
gaps</strong>. Each block is a four-byte header (its end and start address) followed by its data, stored
straight into memory. The store loop deliberately <em>never terminates on its own</em> — it always branches
back for another block header.</p>

<h2>A self-rewriting loader</h2>
<p>That non-terminating loop is the hook the protection hangs on. To end a load, the tape simply sends a block
whose address range <strong>covers the loop itself</strong>: as it loads, the branch instruction at the loop's
tail is overwritten, so when the block completes the loop falls through into freshly loaded code. Between
payload blocks the tape sends tiny one-to-three-byte blocks aimed at single instructions of the loader,
<strong>changing the wire format on the fly</strong> — flipping the bit order between MSB-first and LSB-first,
changing how many bits make a byte frame, growing the header with a decoy byte, and rewriting the whole loader
tail (sometimes byte-identical, a decoy that has to race the executing loop). The main payload is sent as many
small blocks with patch blocks interleaved between almost every page, so the data cannot be read off the tape
without executing the patches. It is a copy-protection scheme, not just a fastloader.</p>

<h2>The multi-stage load</h2>
<p>The chain runs through a short BASIC stub (a tape <code>LOAD</code> from inside a running BASIC program
restarts it afterwards, so a flag variable makes each line run once) into the game's own multi-load routine,
which pulls in the remaining segments — the engine, the in-game code, the colour data and the bitmap loading
picture — behind the loading screen, then jumps into the game.</p>
`,
    engine: `
<div class="info-eyebrow">Elite · Game Engine</div>
<p>Everything in the game image arrives <strong>encrypted</strong> and is unpacked in stages, then relocated —
much of the engine hidden as RAM under the I/O area — before the game brings up its own interrupt-driven
display and reaches its first frame.</p>

<h2>Layered decryption</h2>
<p>None of the game code is ever in plaintext on the tape, and even after loading it is decrypted in pieces,
at different times, with different keys. Three near-identical decryptors do the work, each a
<strong>rolling-subtraction cipher</strong>: every byte is the previous plaintext byte subtracted from the
ciphertext, so the key rolls forward as it goes, and the loop's address range is self-modified. The first
stage decrypts the loaded blob in two downward passes with different keys; a later stage decrypts the rest in
place once it has loaded.</p>

<h2>Relocation under I/O</h2>
<p>The decrypted engine relocates itself. Part of it is copied into low memory; the bulk — about 8&nbsp;KB —
is copied <strong>underneath the I/O area</strong> at <code>$D000–$EFFF</code>, hidden as RAM the routines
reach by toggling the processor port's bank bits (banking I/O out, reading or writing, then banking it back
in). The ship-model data and much of the engine live there, out of the way of the normal memory map.</p>

<h2>Hardware init and interrupts</h2>
<p>Hardware init neutralises RUN/STOP–RESTORE, banks the KERNAL and BASIC ROMs out, and points the CPU's IRQ
and NMI hardware vectors at <strong>RAM</strong> — so with no ROM in the path, interrupts dispatch straight
into the game's own handlers. The interrupt is a <strong>table-driven raster-split engine</strong>: it reads
the current split index and loads that split's VIC register values from a set of two-entry tables, giving two
splits per frame (the bitmap space view and the dashboard). When the second split completes the frame, the
handler falls through to its per-frame work — a three-voice SID player — before returning.</p>

<h2>Reaching the first frame</h2>
<p>From the init onward the game is interrupt-driven: the game start runs its one-time setup, builds the
title / commander screen, and settles into the main loop. The picture shown during the long loads is a
multicolor bitmap — the 3-D "ELITE" logo and a Cobra over a starfield — stored uncompressed and assembled
from three tape segments, displayed before the two largest loads so it masks the slowest part. Once flight
begins, two control flows cooperate: the <strong>raster interrupt</strong> only swaps VIC registers between
the bitmap and dashboard and ticks the music, with no game logic, while a <strong>foreground loop</strong>
does everything else — moving and drawing every object, spawning, combat, scoring and reading the controls.</p>
`,
    graphics: `
<div class="info-eyebrow">Elite · Graphics</div>
<p>Elite's graphics are <strong>vector, not bitmap</strong>: ships are wireframe models drawn line by line,
and the whole universe of star systems is generated from a tiny seed rather than stored. Even the text is held
as compressed tokens, not letters.</p>

<h2>Ship models</h2>
<p>Each ship is a filled-edge wireframe — a list of 3-D vertices joined by edges, with face normals used to
hide the back. The models live in the engine block hidden under the I/O area, reached through a blueprint
pointer table of 33 ship types. A blueprint is a 20-byte header followed by three packed arrays — vertices,
edges and faces — with <strong>no explicit counts</strong>: each array's length falls out of the offsets,
because every record is a fixed size. A vertex is six bytes (its three coordinate magnitudes, a byte packing
the three signs and a visibility distance, and the numbers of up to four faces it belongs to); an edge is four
bytes (a visibility distance, the two faces on either side, and its two vertex numbers); a face is four bytes
(the signed normal plus a visibility distance). The model is a <strong>wireframe</strong> with hidden-face
removal — there is no shading, so nothing in a face record is an illumination value. A face number of 15 is a
<strong>sentinel</strong> meaning "no
face on this side" — an edge carrying it is always drawn, never culled, which is how flat models like the alloy
plate show their outline from any angle.</p>

<h2>Drawing a ship</h2>
<p>Drawing runs a fixed pipeline: copy the ship's state into a zero-page workspace and point at its blueprint;
rotate the model and drop or simplify it by distance (level of detail); project each vertex into view space
with a perspective divide; back-face-test each face with its normal and append every visible edge to a per-ship
<strong>line heap</strong> as endpoint records; then walk the heap drawing each as a Bresenham line into the
multicolor space-view bitmap. Because lines are plotted with EOR, the previous frame's ship is erased by drawing
its old line heap again before the new one is built — no separate clear is needed. There is no double buffer,
though, so a busy line heap is often still being erased and redrawn as the raster sweeps past it: the result is
the <strong>characteristic flicker</strong> of the C64 version, part of the game's signature look. Distant ships
and stars are single dots of one, two or four pixels by distance.</p>

<h2>A universe from a seed</h2>
<p>Elite's universe — hundreds of star systems, each with a name, an economy, a government and a position — is
<strong>not stored</strong> anywhere; a search of the image for names like Lave or Diso finds nothing. Every
name and attribute is generated on demand from a tiny seed. A whole galaxy is defined by a <strong>six-byte
seed</strong> (the only one stored sits in the default commander block), and a single Fibonacci-style
pseudo-random generator — a two-word lagged sum over a four-byte seed — advances it deterministically, so the
same starting seed always yields the same stream and a system's name and data are reproducible. To work on a
system, the code loads that system's seed and runs the generator. A name is one to four two-letter
<strong>digram</strong> pairs drawn from a small letter-pair table — the only stored fragment of any name —
and coordinates and the star/dust field come from the same seed and generator. The other galaxies are
transforms of the same seed, so the entire universe lives in those six bytes plus the algorithm.</p>

<h2>Text as tokens</h2>
<p>Almost every word the game prints — commodity names, government and economy descriptions, combat ranks,
mission briefings, even "GAME OVER" — is stored not as letters but as compressed, EOR-obfuscated
<strong>tokens</strong>: recursive tokens that expand to other tokens, two-letter digrams from the same table
the names use, and control codes that insert the commander's name, a generated captain's name or a number. A
raw search for any in-game phrase finds nothing, because the letters are never there in sequence.</p>
`,
    gameplay: `
<div class="info-eyebrow">Elite · Gameplay</div>
<p>Once flight begins, the game is about the <strong>other ships</strong> — how traffic appears, moves and
fights, and when it vanishes again. It runs in a foreground loop over a small, fixed set of object slots.</p>

<h2>The universe of slots</h2>
<p>The universe is at most ten object <strong>slots</strong> — the planet, the station or sun, and up to a
handful of ships — held in a slot array of ship-type bytes and a table of pointers to each slot's 37-byte
record (position, orientation, speed, AI and state flags). The update pass copies one slot's record into a
zero-page workspace, works on it there (so the same fast code serves every slot), and copies it back. The
<strong>player's ship never moves</strong>: the universe is rotated and translated around a stationary camera,
so flying forward makes everything stream toward you.</p>

<h2>Spawning</h2>
<p>New ships go through one routine that finds a free slot (ten full means nothing spawns), looks up the
blueprint, and places ordinary traffic far away so it approaches from a distance. A <strong>spawn
director</strong> runs periodically — never while docked — and makes weighted random rolls: a chance of lone
trader traffic; law-enforcement ships scaled by how many police are already present and by the player's
accumulated bounty; and pirates gated by the system's government and the commander's danger rating, as either a
single ship or a pack of up to four. So traffic is a mix of traders, police scaled to your record, and pirates
scaled to the system's lawlessness — plus the station and any escorts the AI launches itself.</p>

<h2>Movement and AI</h2>
<p>Each ship is moved forward along its nose vector, rotated by its angular velocity, then re-orthonormalised so
rounding errors do not accumulate, and handed to the wireframe renderer. Ships with hostile AI run a tactics
routine, but only <strong>every eighth frame</strong> — the expensive AI is time-sliced across ships and
frames. The station does not fly; its "AI" is a launch controller that sends an enforcement ship after a wanted
player, and mothership-class ships launch escorts the same way. Ordinary combat ships track their target
(normally the player), accelerate toward their speed limit, steer to point at or away, and — by their
aggression and the player's bounty — fire lasers or break into evasive manoeuvres; a ship that takes enough
damage can flip from attacking to fleeing.</p>

<h2>Despawning</h2>
<p>After a ship is moved, its fate is decided three ways, all funnelling into one removal routine: marked for
removal (it finished what it was doing), <strong>killed</strong> (the reward is banked first — and killing a
police ship adds to your bounty rather than your bank, which is where a clean trader becomes a fugitive), or
<strong>out of range</strong> (drifted too far on any axis — which is why traffic quietly thins out behind you).
Removal decrements the population counters and compacts the slot and pointer arrays so the list stays
contiguous.</p>

<h2>Legal status and bounty</h2>
<p>One byte holds your standing with the law, shown as <strong>Clean</strong>, <strong>Offender</strong> or
<strong>Fugitive</strong>. The byte is not a tidy counter: every ship carries an offence value, and destroying
a ship <strong>bitwise-ORs</strong> that value into the byte rather than adding it. Merging bits means the byte
can only gain set bits on a kill, so your standing ratchets upward and never partly cools from a single act;
lawful ships carry the largest value, so shooting a police ship sets enough high bits to turn a clean pilot into
a Fugitive at once. The same byte is then read as a plain number for the Clean / Offender / Fugitive thresholds.
It comes back down on travel: each hyperspace jump shifts the byte one bit to the right — a literal halving — so
minor offences fade over a few quiet jumps while a serious one lingers. It both follows your behaviour and
drives the game: a dirty record spawns more police, and past a threshold ships engage you on sight.</p>

<h2>Combat rating, and winning</h2>
<p>Separately, every kill pays a bounty into a cumulative score that drives a nine-step <strong>combat
rating</strong> from Harmless up to Elite — it only ever climbs, and reaching the top takes thousands of kills.
Rating and legal status are independent axes: an Elite pilot can still be a Fugitive. There is <strong>no
victory state</strong> in the game at all — the only way the main loop is left is the death sequence (energy
gone under fire, a hard collision, or flying into a planet or sun), after which you resume from your last saved
commander. Elite is deliberately open-ended; "Elite" is the nominal goal, but reaching it only changes a
label.</p>

<h2>Docking</h2>
<p>Flight is entirely manual — but docking is the maneuver that turns that into a test of nerve, and the only
one you can buy a machine to fly for you. A legal dock requires four things at once:
approaching the station's <strong>correct face</strong>, pointing <strong>nearly straight in</strong>,
<strong>laterally lined up</strong> with the slot, and <strong>rolled</strong> to fit the rectangular opening —
all while slow enough that the contact stays under the fatal threshold. Because the station continuously
rotates, you must roll with it to keep aligned, which is the whole skill. Miss the slot at speed and you hit the
hull — the same death path as crashing into a planet. (A purchasable docking computer flies the same approach
automatically.)</p>

<h2>Thargoids</h2>
<p>The alien <strong>Thargoids</strong> sit on top of the ordinary traffic: a tough, aggressive mothership and
the small drones it launches. They arrive either through a <strong>misjump into witchspace</strong> — a random
check on a hyperspace jump that dumps you into a nest of them — or, rarely, as a random in-system attack. The
mothership keeps manufacturing fresh drones while it lives, so a drawn-out fight only grows the swarm. The neat
trick is that the drones read the same population counter that tracks live motherships: while it is non-zero
they attack, but the moment the last mothership dies the drones go inert and drift — there is no explicit
"deactivate my children" code. Kill the motherships and the swarm switches off.</p>

<h2>Missions, economy and government</h2>
<p>On top of the open-ended trading and combat is a small <strong>scripted mission chain</strong>, gated by how
many times you have docked and by specific destination systems: a briefing on docking sends you to a named
system to hunt a unique, tougher enemy or to carry stolen Thargoid plans, with a large fixed reward on return.
Each system's character comes from its <strong>seed</strong>: an economy and a government (each three bits),
with tech level and productivity derived from them and coupled — lawless systems are poor and backward,
well-governed wealthy ones high-tech. Economy drives prices through a signed per-commodity slope, so the same
good is cheap in one economy and dear in another — the whole basis of trade routes; government drives danger,
gating how freely pirates spawn; and tech level decides what equipment is for sale. So the risk and reward of
where to fly fall out of a few bytes per system.</p>
`,
    music: `
<div class="info-eyebrow">Elite · Music</div>
<p>Elite has one piece of music — and it is <strong>Johann Strauss II's <em>The Blue Danube</em></strong>,
the classic nod to <em>2001: A Space Odyssey</em>, played while the docking computer flies you into the
station. The track here is that waltz, rendered from the game's own engine. In flight there is no music,
only sound effects; the title and station screens are silent.</p>

<h2>Two sound engines</h2>
<p>Both run once per frame from the raster interrupt. A <strong>sound-effects</strong> engine drives three
SID voices from a small set of tables — each effect is a short pitch sweep (the laser, the hyperspace whine,
explosions). Separately, a <strong>music sequencer</strong> plays the waltz; the two are gated by their own
flags, so effects can sound with the music off and vice versa. The music starts as the docking sequence
begins and is silenced the instant you dock.</p>

<h2>The music sequencer</h2>
<p>The waltz is a compact <strong>nibble-packed bytecode</strong>: a play pointer walks a command stream
whose opcodes (two per byte) set SID registers or step time — note-on per voice, chords, ADSR, pulse width,
waveform, filter and tempo — looping at the end. Voices 1 and 2 are sawtooth, plucked for the
"oom-pah-pah" bass and accompaniment; voice 3 is a sustained pulse through the resonant low-pass filter — the
melody. Voices 2 and 3 each play against a copy of themselves a fraction sharper, the gentle detuned shimmer
that gives the piece its character.</p>

<h2>From bytecode to audio</h2>
<p>The audio you hear is reconstructed end to end: the command stream is interpreted exactly as the engine
does, and its SID-register writes are fed to a <strong>reimplemented SID chip</strong> — three oscillators,
the envelope generator and the multimode filter — clocked at the C64's rate and bandlimited the way the real
chip's output is. No recording of the original is used; it is played from the bytes.</p>
`,
  },
  sm64ds: {
    loader: `
<div class="info-eyebrow">Super Mario 64 DS &middot; Image &amp; Loader</div>
<p>Super Mario 64 DS ships on a 16&nbsp;MiB Nintendo DS Game Card. A DS card is not a memory-mapped cartridge
like the Game Boy's: it is a serial device whose contents are an on-card <strong>filesystem</strong> the console
reads on demand. The DS is also a <strong>two-processor machine</strong> &mdash; an ARM9 and an ARM7 sharing one
4&nbsp;MiB main memory &mdash; so the card carries two code images that boot side by side, described by a
512-byte header at the front of the image.</p>

<h2>Two CPUs, one memory</h2>
<p>The header hands each CPU a different load and entry address. The <strong>ARM9</strong> (an ARMv5TE core that
owns the card, the game and both screens) loads to <code>$02004000</code> and starts at <code>$02004800</code>,
over its secure-area landing. The <strong>ARM7</strong> (an ARMv4T core servicing sound, the touchscreen,
buttons and wireless) loads near the top of the same RAM at <code>$02380000</code>. Both see the single main
memory <code>$02000000</code>&ndash;<code>$023FFFFF</code> and coordinate through the DS's <strong>IPC</strong>
hardware in the <code>$04000000</code> I/O block. Unlike Mario Kart DS, which bases its ARM9 at the very bottom
<code>$02000000</code>, this game starts <code>$4000</code> bytes higher, leaving the low 16&nbsp;KiB of main
RAM free for the system. Both cores run 32-bit <strong>ARM</strong> and 16-bit <strong>Thumb</strong> code,
switched by <code>BX</code>/<code>BLX</code>.</p>

<h2>Overlays: code paged in on demand</h2>
<p>The ARM9 static binary is only the spine. The game is really a tree of <strong>code overlays</strong> &mdash;
separately loadable code+data blobs swapped into fixed RAM regions as the player moves between the castle, a
painting-world, a minigame or a menu. Its overlay table at <code>$061510</code> lists <strong>103 overlays</strong>
(32-byte records) that occupy just <strong>22 distinct RAM addresses</strong>: overlays sharing an address are
mutually-exclusive banks, only one resident at a time. Each record names the file holding its bytes, a
<code>.bss</code> range to zero and a static-initialiser range to run after loading, and every overlay is
individually <strong>BLZ-compressed</strong> (a backward-decompressing LZ variant the boot code also uses on the
main binaries). The ARM7 has no overlays.</p>

<h2>The filesystem</h2>
<p>Everything that is not boot code &mdash; models, animation, collision, textures, the sound bank, the 2D UI
&mdash; lives in an on-card filesystem addressed by two tables. The <strong>FAT</strong> is a flat array of
start/end offset pairs, one per file (2175 files); the <strong>FNT</strong> is the directory tree that gives
those numbered files names and a hierarchy. File IDs <code>0</code>&ndash;<code>102</code> are the ARM9 overlays
(referenced directly, unnamed); IDs <code>103</code> onward are the named files, yielding paths like
<code>data/stage/castle_1f/castle_1f_all.bmd</code> and <code>ARCHIVE/arc0.narc</code>. A filesystem file is
rarely the raw asset: it is usually a compressed stream, and the top-level <code>ARCHIVE/</code> holds
<strong>NARC</strong> archives that bundle further sub-files.</p>
`,
    engine: `
<div class="info-eyebrow">Super Mario 64 DS &middot; Game Engine</div>
<p>Once the two boot chains meet, the game runs as a dual-core program: the ARM9 drives simulation and rendering
while the ARM7 handles sound and input, the two exchanging messages over dedicated hardware. On top of that base
sits an <strong>overlay tree</strong> that reshapes the ARM9's code map to match wherever the player is.</p>

<h2>The two cores meet</h2>
<p>Each CPU boots from its own entry: an ARM9 startup stub programs the <strong>CP15</strong> system control
coprocessor, sets up the tightly-coupled memories and stacks, then self-decompresses the BLZ-packed body, zeroes
<code>.bss</code>, runs the static initialisers and calls <code>main</code>. The ARM7 comes up in parallel. Before
either reaches its real work they <strong>rendezvous</strong> through the <code>IPCSYNC</code> register &mdash; a
handshake where each core waits to see the other's agreed value &mdash; and then settle into a steady exchange
over the <strong>PXI FIFO</strong>, the message pipe that carries sound commands and input state between them for
the rest of the session.</p>

<h2>Overlays as a state machine</h2>
<p>Because the 103 overlays collapse onto only 22 RAM addresses, the resident set of code <em>is</em> the game's
current state. Entering a world's painting loads that world's overlay bank into a shared region, displacing
whatever was there; returning to the castle swaps it back. The engine is thus organised as a large tree of
on-demand modules keyed to place &mdash; the castle hub, each course, the minigame suite, the menus &mdash;
rather than one monolithic binary, which is how a Nintendo&nbsp;64-scale game fits inside 4&nbsp;MiB of DS RAM.</p>
`,
    graphics: `
<div class="info-eyebrow">Super Mario 64 DS &middot; Graphics</div>
<p>The DS renders in 3D on a dedicated geometry engine, and this game stores its 3D art in a
<strong>bespoke model format</strong> with the extension <code>.bmd</code>. Despite the name it is <em>not</em>
the NitroSDK's <code>BMD0</code> container that Mario Kart DS uses &mdash; the low-level primitives are the same DS
silicon, but the file that wraps them is the game's own.</p>

<h2>The container</h2>
<p>A <code>.bmd</code> file begins with an <code>LZ77</code> magic tag over a standard <strong>LZ10</strong>
compressed stream; decompressed, it opens with a fixed header of <strong>(count, offset)</strong> pairs pointing
at arrays of <strong>bones</strong>, <strong>display lists</strong>, <strong>materials</strong>,
<strong>textures</strong> and <strong>palettes</strong>, plus a single model-wide scale stored as a
power-of-two shift. There is no <code>BMD0</code> stamp and no NITRO resource dictionary; the indices that bind
the pieces together are explicit fields, not named lookups.</p>

<h2>Bones draw the geometry</h2>
<p>Each <strong>bone</strong> carries a transform and a small <strong>render list</strong>: pairs of
(material index, display-list index) telling the engine to draw a given shape with a given material. A
<strong>display list</strong> is a stream of <strong>GX commands</strong> &mdash; the exact command language the
DS's geometry engine consumes &mdash; so that decoder is shared with the NITRO models. It handles every vertex
form the hardware offers, including <em>delta</em> vertices (small signed offsets from the previous vertex) and
the full <strong>matrix pipeline</strong> the display lists drive directly: push/pop of a matrix stack, load,
multiply, scale and translate, which is how the more elaborate stages position their parts. A model's two
sub-display-lists are one continuous command stream and must be decoded together, or a chunk that opens on a
delta vertex snaps back to the origin and flings triangles off into spikes.</p>

<h2>Textures and palettes</h2>
<p>The textures are embedded in the model, not in loose files. Each material names an explicit texture and
palette index; a texture is decoded through the DS's <strong>seven texture formats</strong> &mdash; paletted
4/16/256-colour, two alpha-plus-index formats, 16-bit direct colour, and a <strong>4&times;4 block-compressed</strong>
mode whose texel indices and per-block palette selectors sit in two separate regions. Palettes are 15-bit
<strong>BGR555</strong>. DS texture units are tiny &mdash; often 32&times;32 texels &mdash; so this viewer keeps
them nearest-filtered to preserve the original pixel art. Level textures tile, so their materials repeat on both
axes unless the GX flags say otherwise.</p>

<h2>From cartridge to viewer</h2>
<p>The models on display are the game's own stage and object files, decoded straight from these structures and
converted to glTF with their textures baked in &mdash; the castle floors, each course's terrain, and the
creatures and props that populate them. Course geometry is authored small and scaled up by the model's
power-of-two shift, the size the physics and collision meshes use.</p>

<h2>How the engine places everything</h2>
<p>The render transform is traced end to end in the ARM9 draw code. The engine renders in <strong>world &divide; 8</strong>
units &mdash; an arithmetic shift right by 3 at every seam between simulation and rendering &mdash; and every model
draws under a hardware matrix scale of its header's 2<sup>shift</sup>. The stage alone is drawn with one extra
hard-coded uniform scale of <strong>125.0</strong>, which makes a stage vertex unit exactly
2<sup>shift</sup>&nbsp;&times;&nbsp;1000 world units: the courses are authored in kilo-units. That single constant fixes
where every object stands &mdash; a placement's 16-bit coordinate divided by 1000 lands it on the stage model with no
per-course tuning. Object behavior is C++ virtual methods, not scripts; the traces so far give the viewer its
moving parts: the coin's step adds <code>$C00</code> to its yaw every frame &mdash; a flat quad revolving about 1.4
times a second &mdash; and the signposts are proximity dialogs that snap themselves to the ground with a collision
ray when they spawn. Click one in a course view.</p>
`,
    music: `
<div class="info-eyebrow">Super Mario 64 DS &middot; Music</div>
<p>All of the game's sound lives in one 4.4&nbsp;MiB file, <code>sound_data.sdat</code> &mdash; the NitroSDK
sound archive, the same format Mario Kart DS uses. Every note is <strong>sequenced</strong>, MIDI-style;
there is no streamed audio anywhere on the card. But where Mario Kart DS shipped its archive with the name
block stripped, this one keeps it: <strong>every tune carries the sound project's own symbol</strong>, and
the track list you see here is those names, straight off the cartridge.</p>

<h2>The archive, with its names</h2>
<p>The archive binds <strong>79 sequences</strong> (<code>SSEQ</code>), <strong>109 instrument banks</strong>
(<code>SBNK</code>) and <strong>89 wave archives</strong> (<code>SWAR</code>) under the project prefix
<code>NCS_</code>: <code>NCS_BGM_TITLE</code>, <code>NCS_BANK_SE_ACTION</code>,
<code>NCS_WAVE_RESIDENT</code>. The music names are Japanese romanized development names, and they sort the
soundtrack at a glance: <code>SHIRO</code> (castle), <code>CHIJOU</code> (overground), <code>OBAKE</code>
(the haunted house), <code>KUPPA</code> (Bowser &mdash; his Japanese name, Koopa), <code>DOLPIC</code> for
the beach and its sung twin <code>VOCAL</code>. A few records share one file &mdash; the minigame versions of
the Bob-omb and Piranha themes are literally the same sequence played through a different setup.</p>

<h2>A soundtrack in families</h2>
<p>The names split the 83 playable records into clear families: the long looping <strong>world themes</strong>
(a minute and more), a set of short <strong>power-up loops</strong> (<code>MUTEKI</code> invincibility,
<code>METAL</code>, <code>FIRE</code>, <code>BALLOON</code>), the <code>MINI*</code> minigame set played on
the touch screen, a <code>VS*</code> set for wireless multiplayer, and a fanfare for everything that can
happen to you &mdash; from <code>FIRSTCAP</code> (finding your first cap) and <code>KAGI</code> (a key) down
to <code>GAMEOVER</code>.</p>

<h2>Playback</h2>
<p>What you hear is rendered from the bytes: the sequence bytecode runs against the sound driver's clock of
<strong>48 ticks per quarter note</strong>, envelopes step in decibels at the driver's 192&nbsp;Hz frame rate,
instruments resample their PCM or ADPCM waves for each note's pitch, and the voices mix at the DS's
32768&nbsp;Hz &mdash; the same reimplemented sequencer and synthesizer that plays Mario Kart DS's music,
reused unchanged. Each track plays two loops and fades.</p>
`,
    gameplay: `
<div class="info-eyebrow">Super Mario 64 DS &middot; Gameplay</div>
<p>Everything that moves in a course &mdash; coins, enemies, signposts, the platforms you ride &mdash; is an
<strong>actor</strong> placed by the level data and run by the engine, and every actor feels the world through a
<strong>collision mesh separate from the model you see</strong>.</p>

<h2>From a level number to its objects</h2>
<p>Loading a course indexes two 52-entry tables in the ARM9 data: one maps the level to its code overlay, the other
to a 28-byte <strong>settings block</strong> that names the level's display model, its collision mesh and its object
tables as internal file IDs. Each placed object is a compact record &mdash; an object id and a signed 16-bit
<code>x/y/z</code> whose shorts, shifted up, <em>are</em> the world coordinate, plus a rotation and a couple of
parameters. Standard objects are 16 bytes; <strong>entrances</strong> share the form and the level's first one is
where Mario spawns; and 8-byte <strong>simple objects</strong> carry set-dressing like trees and coins. Across the
52 levels this comes to about <strong>4,350 placements</strong>. Every object id is first translated through an
<strong>object&rarr;actor table</strong> before spawning &mdash; the id in the level data is not the actor the
engine runs.</p>

<h2>Actors are C++ objects</h2>
<p>An actor's behavior is <strong>C++ virtual methods, not a script</strong>: each profile installs a vtable whose
slots the engine calls every frame &mdash; an init, a per-frame step, a draw. What an actor <em>does</em> lives in
its step, and the ones traced here show the range:</p>
<ul>
  <li>a <strong>coin</strong> adds a fixed amount to its yaw every frame &mdash; a flat quad turning about 1.4 times
  a second, so the spin is <em>geometry</em>, not a texture animation;</li>
  <li>a <strong>signpost</strong> casts a downward collision ray when it spawns to <strong>snap itself to the
  ground</strong>, then watches an interaction flag and starts a dialog when the player is in range (its words come
  from the message system);</li>
  <li>a <strong>goomba</strong> wanders: it eases its forward speed toward a per-state value, eases its heading
  toward a target, repicks a random heading on a 100-frame timer (three turns in four, a pause the fourth), stays
  <strong>leashed</strong> to its spawn point, and drops into a chase inside its sight radius;</li>
  <li>a <strong>chain chomp</strong> lunges on a chain strung to a stake it <strong>spawns as a child actor</strong>,
  and a <strong>bob-omb</strong> wanders home-biased and sprints once its fuse is lit.</li>
</ul>

<h2>Collision &mdash; the .kcl mesh</h2>
<p>The mesh you see is <strong>never</strong> queried for physics. Every level, and every solid platform, pairs its
display model with a <code>.kcl</code> <strong>collision mesh</strong>, and actors probe it through a small family
of walkers that live in the DS's zero-wait instruction memory: a <strong>down-ray</strong> ("the highest floor
under this point"), a segment sweep and a sphere. The mesh is a plane-based triangle soup indexed by an
<strong>octree</strong>, so a query descends straight to the few triangles near the point; a triangle is stored not
as three corners but as one anchor vertex plus normal indices and an edge length &mdash; the compact plane-based
encoding.</p>
<p>What a surface <em>does</em> &mdash; floor, wall, water, death plane, slippery ice &mdash; is deliberately
<strong>not</strong> in the mesh. Each triangle carries only an index into a separate <code>CLPS</code> attribute
table, and every query carries a flag byte that <strong>filters which surfaces it can see</strong>. So the same
mesh answers a plain ground ray (only ordinary floors) and a water probe (the water plane) differently &mdash; the
reason a signpost's ground-snap ignores the pool it stands beside. The viewer's <strong>Collision</strong> toggle
draws this mesh in red; what stays uncoloured &mdash; bridge rails, flags, window glass &mdash; has no collision at
all.</p>
`,
  },
  uw: {
    loader: `
<div class="info-eyebrow">Ultima Underworld &middot; Image &amp; Loader</div>
<p>Ultima Underworld is a 16-bit <strong>MS-DOS real-mode</strong> program from 1992, before DOS extenders were
common. <code>UW.EXE</code> begins with the <code>MZ</code> signature and carries no <code>PE</code>, extender stub
or protected-mode header &mdash; it runs on the bare 8086-compatible instruction set, addressing memory through the
classic <code>segment:offset</code> scheme, and it pages its own code out of the executable because the whole game
cannot fit in the 640&nbsp;KB a real-mode PC offers.</p>

<h2>The executable and its overlays</h2>
<p>The header describes a load module of roughly <strong>408&nbsp;KB</strong> that DOS copies into memory, but the
file is larger still: about <strong>141&nbsp;KB of extra data sits appended past that image</strong>, which DOS never
loads. That trailing region is a <strong>self-managed code-overlay store</strong> &mdash; segments of code the program
pages in on demand as the player moves between the world, the menus and the intro. The resident image is only a spine;
the 3-D renderer, the world code and much else live in overlays fetched from the store as needed. The header also
carries a 3,176-entry relocation table, since every far pointer baked into the image must be fixed up with the
runtime load segment.</p>

<h2>The data catalog</h2>
<p>Beside the executable sit four folders of data. <strong><code>DATA/</code></strong> is the core database: the two
big archives <code>LEV.ARK</code> (the dungeon levels) and <code>CNV.ARK</code> (the conversation scripts), the
<code>.GR</code> image banks (objects, weapons, heads, cursors), the <code>.TR</code> wall and floor
<strong>textures</strong> at 16/32/64 pixels, full-screen <code>.BYT</code> images, small <code>.DAT</code> tables
(objects, palettes, skills), the packed text in <code>STRINGS.PAK</code> and the bitmap fonts.
<strong><code>CRIT/</code></strong> holds the creatures as paged animation-frame banks, <strong><code>CUTS/</code></strong>
the cutscenes as sequenced page files, and <strong><code>SOUND/</code></strong> the audio &mdash; digitized
<code>.VOC</code> voice, sequenced <code>.XMI</code> music and the Miles sound-driver files. None of these is a
memory-mapped resource: the game opens and parses each file as it needs it.</p>
`,
    engine: `
<div class="info-eyebrow">Ultima Underworld &middot; Game Engine</div>
<p>Ultima Underworld manages its own memory and its own code, and it drives the PC hardware directly &mdash; leaning
on a handful of DOS, BIOS and hardware services more heavily than most games of its day. Understanding how it runs
means understanding its <strong>overlay system</strong> and the machine it expects underneath it.</p>

<h2>Code paged in on demand</h2>
<p>Because the program is far larger than it loads, its C runtime installs an <strong><code>INT 3Fh</code> overlay
dispatcher</strong>. A call into overlaid code traps through that vector; the handler reads a small thunk off the
return address, looks the overlay up in a resident directory, and <strong>reads its code out of <code>UW.EXE</code>
into a fresh segment</strong> &mdash; standard Microsoft-C overlay machinery. So which code is resident depends on
what the game is doing: the renderer's overlay (segment <code>07F7</code>), the tile tessellator (<code>1FF9</code>)
and the software rasteriser page in and out as the view is drawn.</p>

<h2>The machine it expects</h2>
<p>UW is demanding about its environment. It <strong>requires an EMS (expanded memory) driver</strong> &mdash; without
one it prints "Out of EMS Memory" and quits &mdash; and uses it to stream whole screens through a 64&nbsp;KB page
frame. It manages RAM by resizing its own program block through DOS and sub-allocating inside it. It talks to the
hardware directly: it detects a VGA through the video BIOS and sets <strong>mode 13h</strong>, probes for a Sound
Blaster Pro, handshakes the keyboard controller and reads the timer, and hooks the timer interrupt while chaining the
previous handler. Input arrives at two layers: UW installs its <strong>own keyboard interrupt handler</strong> that
reads each raw scancode into a 64-byte ring, and it polls a Microsoft-compatible <code>INT 33h</code> mouse driver for
the pointer &mdash; the two feeding its real-time first-person UI.</p>
`,
    graphics: `
<div class="info-eyebrow">Ultima Underworld &middot; Graphics</div>
<p>The dungeon view is a real-time <strong>texture-mapped first-person 3-D renderer</strong>, written entirely in
software on the CPU &mdash; there was no hardware 3-D in 1992. It draws the scene into an off-screen buffer and blits
that to the screen, and its whole reason for existing is speed: there is no per-pixel divide anywhere in it.</p>

<h2>World to pixels</h2>
<p>The pipeline is the textbook one, done in fixed point. A world point is rotated by the <strong>camera basis</strong>
&mdash; a 1.15 fixed-point orientation matrix, re-orthonormalised each frame &mdash; into view space, then run through
a <strong>perspective projection</strong> (<code>screenX = X &middot; scale / Z + centre</code>, one divide per axis)
to a screen vertex. Polygon setup computes the constant edge and texture slopes, and a <strong>two-level affine
stepper</strong> draws it: a vertical loop steps the left and right edges down the polygon, and for each scanline a
horizontal loop steps a texture coordinate across the span, copying one texel per pixel into a linear
one-byte-per-pixel <strong>chunky buffer</strong>.</p>

<h2>Self-modifying spans</h2>
<p>The stepper has no divide in its inner loops: the only divides are two per span, for the texture gradients. Every
per-step delta is a <strong>self-modifying immediate</strong> that the level above patches &mdash; the edge slopes
written into the scanline loop, the texture gradients into the span loop. Drawing a textured wall is therefore:
project the corners, compute the constant slopes and patch them in, then step down the polygon patching and running
each span. Textures are the loaded <code>.TR</code> bitmaps, 32 texels wide, sampled straight through.</p>

<h2>Divide by design</h2>
<p>The renderer <em>never</em> guards its divides. When a basis vector is axis-aligned the normalised value is exactly
1.0 &mdash; a quotient one past the signed-16 maximum, which <strong>overflows the divide</strong>. Rather than test
for that on every operation, the renderer <strong>arms its own divide-error interrupt handler</strong> and lets the
fault fire; the handler saturates the result and returns. It does the same for the projection when a point sits at the
eye (<code>Z</code> near zero). It is the classic fixed-point-renderer trick, routed through the CPU's exception
vector.</p>

<h2>The Mode&nbsp;X framebuffer</h2>
<p>Nothing reaches the screen until a final <strong>chunky&rarr;planar blit</strong>. Right after setting mode 13h the
game turns chain-4 off &mdash; the unchained <strong>Mode&nbsp;X</strong> &mdash; so VGA memory is four parallel planes.
The blit copies the chunky buffer out in four passes, one plane at a time, deinterleaving it into the planes; that
planar layout is why a naive linear read of the framebuffer would show the picture repeated four times across.</p>
`,
    gameplay: `
<div class="info-eyebrow">Ultima Underworld &middot; Gameplay</div>
<p>The Stygian Abyss is a stack of <strong>nine dungeon levels</strong>, and each is a 64&times;64 grid of
<strong>tiles</strong> carrying its own geometry, textures and a list of the objects standing on it &mdash; the data
the renderer turns into the world you walk through in real time.</p>

<h2>The level archive</h2>
<p><code>LEV.ARK</code> is a block archive: a block count, a table of offsets, then the blocks. The first nine blocks
are the nine dungeon levels, each a fixed size; the rest are the per-level texture lists, automap and animation data.
A level's tile map is the first thing in its block.</p>

<h2>The tile map</h2>
<p>Each level begins with a <strong>64&times;64 array, two words per tile</strong>. A tile carries its
<strong>type</strong> &mdash; solid rock, open floor, one of four diagonal (half-solid) corners, or a floor sloping
up toward a compass direction &mdash; a <strong>floor height</strong>, and 6-bit <strong>floor and wall texture</strong>
indices into the level's own texture list. The diagonals are one shape in four orientations and the slopes likewise,
rotated by a small remap table indexed by facing, so the geometry code only ever handles a canonical orientation.</p>

<h2>Tiles become polygons</h2>
<p>The visible geometry is <strong>rebuilt every frame</strong>: a tessellator walks the tiles in view and emits, per
tile, floor, ceiling and wall quads whose corners are tile-space coordinates and whose textures are that tile's
indices, into a display list the geometry overlay then transforms and projects. Nothing is pre-baked &mdash; a tile
becomes polygons on the fly, which is how the engine renders a large, vertically varied dungeon from a compact grid.</p>

<h2>Objects</h2>
<p>Right after the tile map sit the <strong>objects</strong>: 256 mobile objects of 27 bytes and 768 static ones of 8
bytes, each an item id plus a packed position and a <strong>link to the next object</strong> in its tile's list. Each
tile holds the head of a linked list of the things resting on it, so only the fine offset within the tile is stored.
Doors, wall decorations like moss and vines, and 3-D props are all objects. The whole 3-D view is in fact a
<strong>display-list bytecode</strong> &mdash; a small virtual machine with control flow, vertex transforms and draws,
which the geometry overlay runs each frame &mdash; so an animated door is a sub-program the engine selects from the
door's own state.</p>
`,
  },
};

// HTML for a game/tab, or null if nothing has been written for it yet.
export function infoHtml(gameId, tabId) {
  const game = INFO_CONTENT[gameId];
  return (game && game[tabId]) || null;
}
