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
  sonic: {},
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
stack-resetting jump the moment fire is pressed, and from then on it waits for the frame counter to
change, runs the per-frame logic chain — the object engines, zone checks and state dispatch — and
loops. Because it is gated on the frame counter, the loop runs in lock-step with the raster handlers
that drive the screen.</p>
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
  },
  turrican: {},
  marble: {},
  stuntcar: {},
  elite: {},
};

// HTML for a game/tab, or null if nothing has been written for it yet.
export function infoHtml(gameId, tabId) {
  const game = INFO_CONTENT[gameId];
  return (game && game[tabId]) || null;
}
