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
