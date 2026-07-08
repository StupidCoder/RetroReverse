# Analysis playbook — how these projects were reverse-engineered

A field guide to the methods used across the games in this repository, written so
that a future session (human or AI) can start a new project with a plan instead
of a blank page. It distils what worked across the collection — C64 tape (Elite,
Fort Apocalypse), Amiga disk (Marble Madness, Turrican, Stunt Car Racer),
cartridge ROM (Sonic on Game Gear, Super Mario Land on Game Boy, Super Mario 64 DS
and Mario Kart DS on the DS), optical disc (Ridge Racer on PlayStation, Need for
Speed on 3DO), and a DOS x86 executable (Ultima Underworld) — and generalises the
moves that transfer to an unknown medium and unknown formats. §2 and §3 walk the
two media worked in most depth (tape, disk); the cross-cutting sections §4–§7
carry the transferable technique, and the per-game writeups hold the specifics.

This document is the *methodology*. The repository's *structure* — the tool
layout, the CPU/oracle/extract command flags, the extracted-asset formats, and the
writeup style — is specified separately in [`STANDARDS.md`](STANDARDS.md); the
data interchange format the extractors produce and the viewer consumes is in
[`FORMAT2.md`](FORMAT2.md). Where this guide names a path or a tool it follows
those; when they disagree, they win.

The house rule for all of it: **static analysis first.** No third-party
emulator, debugger, or disassembler; no looking up released source or other RE
projects. Build the tools, read the bytes, disassemble the code, and only reach
for a *self-built* emulator or debugger when the answer is genuinely not on the
medium (state that the running machine/OS produces). When you do, the deliverable
is still *documented understanding plus a clean reimplementation*, never dumped
data.

---

## 0. The shape every project converges on

Every game, regardless of medium, converges on the same arc. When you start a new
game, this is your default table of contents — fill it top to bottom, because each
part unlocks the next:

| Part | Question | C64 (tape) | Amiga (disk) |
|------|----------|------------|--------------|
| **I — The image** | How is the medium framed, and what files does it hold? | TAP pulse stream → KERNAL + fastloader encodings → program segments | ADF sectors → OFS/FFS filesystem → catalogued files |
| **II — Boot chain** | What runs first, and how does it bring in the rest? | autostart → bootstrap → turbo loader → multi-stage load | bootblock → dos.library → startup-sequence → launcher → loaders |
| **III — Program architecture** | How does the loaded game decrypt/relocate/initialise itself? | decrypt+relocate, hardware init, IRQ architecture, memory map | multi-stage decrypt, the decryptor, copy protection, static limits |
| **IV — Graphics & data** | What are the asset formats? | charsets, sprites, bitmaps, maps, compression | IFF ILBM, icons, custom sprite banks, level/sound modules |
| **V — Game mechanics** | How do the objects behave? | object tables, movement/AI, collisions, scoring | object/actor structs, movement/physics, state machines, collisions, scoring |
| **VI — Sound & music** | How is audio produced? | SID register writes, SFX player, music sequencer/bytecode | Paula sample channels, mod/sequencer playback |

The cells name the *medium-typical* form each part tends to take — not a fixed
checklist, and not the specifics of any one game (those live in the per-game
writeups). This is also the canonical Part numbering the writeups use (see
STANDARDS §5); games subdivide a Part (separate model / level / collision /
message formats) but keep the order and number sequentially.

**Only Parts I–II change shape with the medium; III–VI are largely medium-neutral.**
The two media worked in most depth are tape (§2) and disk (§3); the others slot the
same way, differing mainly in how Part I frames the image and Part II reaches the
entry point:

- **Cartridge ROM** (Game Gear, Game Boy, DS) — Part I is a memory-mapped ROM with a
  header and bank scheme (GG/GB mapper registers; DS `FNT`/`FAT` + overlays), read
  directly rather than decoded from a signal. Part II is the on-cart boot/reset
  vector into the engine; there is no filesystem-loader stage to reverse.
- **Optical disc** (PlayStation, 3DO) — Part I parses the disc image (ISO/OperaFS
  directory) and the platform executable (PSX-EXE, 3DO AIF); Part II is the BIOS/OS
  hand-off (self-relocation, BSS, entry) that a boot oracle drives (§6).
- **DOS x86 executable** (Ultima Underworld) — Part I is the MZ image and the
  real-mode memory/segment map; Part II is DOS/BIOS service emulation (INT 10h/21h/
  33h) up to the game's own entry.

Write it up *as you go*, with byte examples and assembly snippets. The writeup is
the deliverable; the tools are how you earn it.

---

## 1. The reusable toolkit

Everything is one Go workspace (`go.work`) over a shared `tools` module plus a
per-game `extract/` module. The shared tools are grouped by role: CPU cores in
`tools/cpu/<arch>/`, platform machines and their format/codec libraries in
`tools/platform/<platform>/`, the disassembler/tracer command binaries in
`tools/cmd/`, and genuinely cross-platform helpers in `tools/lib/` (see STANDARDS
§1 for the full layout). The investment that paid off most was a **CPU core you
can both disassemble with and execute**, per instruction-set:

- **Disassemblers, two flavours each — both still earn their keep.** A *linear*
  sweep (`dis6502`, `dis68k`) and a *recursive-descent* tracer (`codetrace6502`,
  `codetrace68k`). Reach for the **tracer** to make a whole 60 KB binary legible: it
  follows control flow from entry points, segments functions, and merges an
  **annotations** sidecar so labels and notes ride along into the listing — feed it
  every entry you find (vectors, jump tables, library stubs). Reach for the **linear**
  sweep for a quick bounded read of a *known* address range, and — the case that
  recurs — for code the tracer marks as **data** because it is only reached by
  computed/indirect jumps (a `JMP (pc,Dn)` dispatch, self-modified vectors). Rather
  than hunting for an entry point to feed the tracer, just linearly disassemble that
  byte range; the two are complements, not a hierarchy.
- **An executable CPU core** (`tools/cpu/mos6502`, `tools/cpu/m68k`, and one per
  instruction-set: `z80`, `sm83`, `arm`, `arm60`, `mips`, `x86`). This is the secret
  weapon. When a loader's bit-framing or a decryptor's key schedule is painful to
  reverse purely on paper, *run it* on your own core with the real input and
  watch what it produces. It is also how you validate a reimplementation
  bit-for-bit. Expect to implement instructions lazily — add them when the target
  code first hits them (Marble Madness needed EXG, the bit ops, NOT/NEG/NEGX
  added to the 68k core mid-project; each came with a unit test). Where a published
  per-instruction conformance suite exists, gate the core on it (SingleStepTests for
  6502, 8088/x86, and MIPS, run only when the test data is present); cores without a
  suite keep hand-written assembled-loop tests. STANDARDS §2 lists the current state.
- **Container/format readers** per platform: `tools/platform/c64/{tap,cbmtape}`,
  `tools/platform/amiga/{adf,hunk,iff,icon}`. Each is a library plus a `cmd/`
  front-end (`tapdump`, `adfdump`, `hunkload`, `amigapng`).
- **Graphics** (`tools/platform/c64/gfx`): palettes, char/sprite/bitmap rendering,
  PNG and APNG output. Rendering an asset is how you *prove* you decoded it.

Principle: **make each finding reproducible from the raw image with one
command.** Each game separates three output roles (STANDARDS §1): the raw
ROM/disk image and any regenerable scratch are **not** committed — the image lives
in `games/<slug>/image/` and is `.gitignored` for copyright, dev/debug renders and
dumps go to `games/<slug>/work/` (also `.gitignored`); the curated web deliverables
the `webexport` tool produces are committed under `site/public/<slug>/`; and the
handful of images a writeup actually embeds are committed under
`games/<slug>/figures/` — *only* those, nothing a writeup does not reference. There
is no `rendered/` or `extracted/` dir. The rule of thumb: if a tool can re-emit it
and no writeup shows it, it belongs in `work/`, not in git.

---

## 2. Flow A — C64 tape games (Elite, Fort Apocalypse)

The medium is a `.tap`: a stream of **pulse-width** bytes, not bytes of program.
The whole of Part I is turning pulses back into program.

1. **Pulse histogram first (`tapdump`).** Plot the distribution of pulse widths.
   Commodore tapes cluster around a few widths (short/medium/long); the peaks
   tell you the threshold values *this* tape uses. A standard KERNAL tape and a
   custom turbo each have their own characteristic histogram — you can often see
   the boundary in the image where one gives way to the other.
2. **Decode the standard KERNAL part (`cbmtape`).** The first file is almost
   always a normal ROM-loadable bootstrap: pilot tone → sync bytes → header
   (load address, length) → data, with the KERNAL's checksum and the
   pilot/sync/dipole framing. Decoding this gives you the bootstrap program *and*
   teaches you the tape's standard timings as a baseline.
3. **Reverse the fastloader from the bootstrap.** The bootstrap installs a turbo
   loader (often into the tape buffer / low RAM) and hands the tape to it.
   Disassemble it. You are looking for: how it times a pulse (the threshold
   compare), how it frames a **bit** (pulse → 0/1), how it frames a **byte**
   (bit order, sync), how it frames a **block** (count, load address, checksum),
   and where it stores the result. Watch for **self-modifying code** — Elite's
   loader rewrites itself from the tape, so the bytes you disassemble statically
   are not the bytes that run; trace the modifications.
4. **Emulate when framing is ambiguous.** If the exact bit/byte rule is hard to
   pin on paper, run the disassembled loader on the 6502 core with the real pulse
   stream as input and capture what it writes to memory. This is faster and more
   certain than arguing about edge cases, and the C64 machine model
   (`tools/platform/c64/c64`) exists precisely to run hostile loaders.
5. **Extract the segments → `.prg` files.** Now you have the program in memory
   images. Map the multi-stage chain (each stage often loads the next, sometimes
   with a different encoding or under an IRQ).
6. **Program architecture (Part III).** Trace the loaded code: the
   decrypt/relocate step (Elite `SYS 30215 → $7607`; in-place decrypt and
   hand-off), hardware init, and the **interrupt architecture** — C64 games live
   in raster IRQs, so find the IRQ vector chain early; it structures everything.
   Build a memory map.
7. **Assets (Part IV) and mechanics (Part V).** Charsets, sprites
   (packed-column formats appear — Fort Apocalypse), bitmaps, and **compression**
   (Fort Apocalypse uses a table-selective RLE; identify the control-byte scheme
   and the literal runs). Then the object tables and behaviours: the player,
   enemies, bullets, collision matrix, scoring, level progression.

C64-specific instincts: everything is anchored to fixed addresses (`$0314` IRQ
vector, VIC-II `$D000`, SID `$D400`, CIA `$DC00`); banking via `$01` swaps ROM
for RAM under the I/O space; sprites and charsets are page-aligned, which helps
you spot them by address.

---

## 3. Flow B — Amiga disk games (Marble Madness)

The medium is an `.adf`: a flat dump of 1760 × 512-byte sectors. Unlike a tape,
there is usually a real **filesystem**, so Part I is filesystem archaeology, not
signal processing.

1. **Parse the filesystem (`adfdump`).** OFS/FFS: boot block (blocks 0–1, "open
   dos.library"), root block (880), the bitmap (free-space) blocks, and directory
   **hash chains** of file-header blocks. Each OFS file-header block lists its
   data blocks; each **OFS data block** is a 512-byte sector = 24-byte header
   (type 8, header-key, sequence, size, next-pointer, checksum) + 488 payload
   bytes. Extract every file, preserving the tree.
2. **Know storage layout vs content.** The OFS 24-byte-per-sector framing is a
   *storage layout*, not compression — the reader undoes it on extraction. Don't
   confuse a high-entropy *extracted* file (encryption/compression) with the
   on-disk block framing.
3. **AmigaDOS hunk format (`hunk`/`hunkload`).** Executables are hunk files:
   `HUNK_HEADER` (table of segment sizes) then `CODE`/`DATA`/`BSS`, `RELOC32`
   relocation tables, optional `SYMBOL` and `DEBUG`. Load to a flat, relocated
   image for the disassembler. **Keep the `SYMBOL` table if present** — it names
   the routines (Marble Madness's `c/bootscr` shipped 106 symbols and read almost
   like source); a stripped file you name from behaviour.
4. **Boot chain (Part II).** Bootblock → `dos.library` resident → `startup-
   sequence` → Workbench or CLI → the game launcher. Disassemble the launcher
   (`codetrace68k` + the m68k core). Library calls are negative offsets off a
   base in `a6` (exec base at absolute `$4`; `AllocMem = -198`, `FindTask =
   -294`, `DoIO = -456`, …) — resolving those LVOs is most of reading Amiga code.
5. **Encrypted payloads.** Spot them by **entropy** (≈7.95/8 bits/byte) and a
   shared packer signature. Reverse the decryptor as ordinary compiled C: find
   the input layer, the structure walk, and the key schedule. Marble Madness's
   `c/zzz` is a `LoadSeg` replacement whose decode is a seeded table + an
   additive lagged-Fibonacci keystream XORed over *selected* fields (type markers
   and symbol names stay plaintext). **Reimplement the cipher in Go** and validate
   it against known plaintext (a hunk file's header longwords are mostly zero/
   known — derive the keystream from them).
6. **Direct-hardware loaders.** A game may bypass the filesystem for speed and
   protection: Marble Madness's decrypted `c/xxx` drives Paula disk DMA (`$DFF000`
   — `DSKPT`/`DSKLEN` written twice to arm DMA) and the CIAs (`$BFD100` drive
   control, `$BFE001` status) directly, reading raw MFM tracks, finding `$4489`
   sync words, MFM-decoding standard `[$FF,track,sector,gap]` sectors. The tell
   that it reads **by physical position, not by name**: the main data file never
   appears as a filename string in the launcher, yet the disk is 100% filesystem
   (classify every block to prove there's no hidden region). The file is a real
   DOS file so the disk stays bootable and its blocks stay reserved and
   contiguous; it's read by track for speed and as an anti-copy hook.
7. **Graphics (Part IV).** Standard formats decode with stock readers: IFF
   `FORM…ILBM` (BMHD/CMAP/BODY ByteRun1/CRNG cycling) via `iff`; Workbench
   `.info` icons (`$E310` `DiskObject` + planar `Image`s) via `icon`. Custom
   sprite banks you reverse by hand: a fixed-size descriptor table + a packed
   bitmap stream — derive the cell geometry from the descriptor stride, then
   crack the pack codec (or note where it resists static solving because its only
   spec lives in the still-encrypted consumer).

---

## 4. Cross-cutting techniques (the transferable bits)

These are the moves that recur regardless of platform — the ones worth reaching
for on any unknown format:

- **Entropy triage.** Bytes/byte near 8.0 ⇒ compressed *or* encrypted. A packer
  signature header + full-entropy body that decodes to *full-size* output (sizes
  match the header's segment table) ⇒ **encryption, not compression**. Compressed
  data decodes to something *larger* than the stored bytes; encrypted data is the
  same size. (Marble Madness's `.dat`/`c/xxx` are encrypted, full-size — calling
  the loader a "decruncher" was the early mislabel; it is a decryptor.)
- **Structure as a crib.** Most formats keep *some* fields in clear: hunk type
  markers, symbol names, IFF chunk IDs, a file's own size table. Those are
  known-plaintext. For an additive/XOR keystream cipher, known plaintext at a
  known position gives you keystream values directly; solve the generator or the
  table from them.
- **Emulation as an oracle.** When reversing a transform on paper is slow or
  uncertain (loader bit-framing, a key schedule, a self-modifying routine), run
  the real code on your own CPU core with real input and *observe*. This both
  answers the question and gives you a golden reference to test the
  reimplementation against. It is the single highest-leverage tool in the kit.
  But the oracle is a *guide and a verifier, not the data*: read the structure
  out of the **code**, then use the oracle to confirm it — don't scrape decoded
  bytes out of its RAM and call that the format.
- **Follow the game's own pointers — do NOT identify data heuristically.** The
  game does not guess where its data is; it reads a pointer from a table, indexed
  by a level/stage/object number. **That table, that index, and the bank/segment
  it lives in are all somewhere in the code — find them.** So when you need "where
  is level *N*'s map / the tileset / the palette," do not eyeball the ROM for
  plausible-looking data and infer offsets by shape: set a watch on the load,
  capture the *PC* that reads or computes the pointer, and disassemble *backwards*
  from there to the literal table. A render that "looks coherent" is not proof you
  found the right table — Super Mario Land's level maps decoded into a believable
  level from the *wrong* start index and a *mis-located* order table, twice, before
  tracing `level-id → $0470 index → $0D64 bank → $4000[idx]` gave the real one. The
  only time data is found heuristically is when *the game itself* does so (rare);
  otherwise heuristic offset-hunting is a smell that you skipped the lookup.
- **Climb the write-profiler up the pipeline.** When you want the *producer* of a
  value — who filled this buffer, who built this table, who wrote this VGA latch —
  don't reason about it, measure it: profile "which PC wrote address X" (Ultima
  Underworld's DOS oracle has `-vgaprof`/`-profrange` for exactly this), then treat
  that writer's *inputs* as the next X and profile again, climbing the call stack
  one producer at a time until you reach the source data or a literal table. This
  measures the pipeline instead of guessing it, and it catches the case that defeats
  static reading: a **self-modifying patcher** whose output byte was written by code
  that itself was written moments earlier — a write-profile sees the real writer, a
  disassembly sees stale bytes.
- **Hook PCs to log decisions; verify algorithms as sequences.** The write-profiler
  answers "who wrote X"; the complementary move verifies a *reimplemented algorithm*
  (a colour-assignment rule, a procedural geometry build) rather than a data format.
  Plant hooks on the exact instruction addresses where the engine commits each
  decision — a descriptor write, the colour load before a fill call, every edge-emit
  call — log the registers/memory *at that moment*, and diff the whole logged
  sequence against the Go reimplementation's output. On a step-loop core this is one
  map lookup per instruction (`if h, ok := hooks[c.PC]; ok { h() }`); Stunt Car
  Racer's preview colours, car build and horizon objects were each proven this way
  (every strip/edge/fill across all tracks, exact). Capture values **at the hook**,
  not after the run — work buffers get reused (the car recomputes one four-vertex
  slot set per wheel), so post-hoc reads see only the last occupant.
- **Expect "the asset is code."** Not every drawn thing has stored data, so a failed
  pointer-hunt for vertex data is itself a clue. Stunt Car Racer's opponent car is
  built procedurally each frame from four projected wheel points (fixed fractions of
  the axle vector — no model anywhere on disk); its drawbridge ramps are *rewritten
  into the track data* every frame; its horizon range is one 4-vertex silhouette
  whose negative coordinate words are placeholders filled from per-instance streams.
  When data won't turn up, look for a **builder**: find the code that writes where
  the renderer reads, and reimplement the construction instead of a decode.
- **Locally weird decoded data = suspect a runtime patcher.** If a verified pipeline
  produces nonsense in one small region (the Draw Bridge ramps decoded as narrow
  height spikes while 76 of 78 sections were perfect), don't rationalise it —
  byte-search the binary for the region's absolute address and find who *writes* it.
  The spikes were on-disk placeholder bytes the game never displays: a per-frame
  animator regenerates that table before anything draws, and even the "static"
  preview runs one pass of it first. Related tell: several handles decoding to
  *identical* records ⇒ diff the handle values — they were overlapping 9-entry
  windows into one 36-entry table, not four copies.
- **An oracle comparison only verifies the state you actually established.** A
  routine that begins with a compare-and-return (`CMPI.b #5,$1CA33; BEQ …; RTS`) is
  gated on state your harness may never set — and then it silently no-ops on *both*
  sides, so the comparison passes while verifying the wrong world. Stunt Car Racer's
  bake oracle reported track 5 byte-exact for weeks while both engine and Go were in
  the never-displayed placeholder state, because the animator checks a menu byte the
  harness didn't write. When you add a call to an oracle chain, read its first few
  instructions and confirm every gate variable is established; and phrase results
  honestly — "byte-exact" is a statement about the state you created, not about the
  game as played.
- **Absence of a throttle is a finding — don't invent a frame rate.** Before claiming
  real-time cadence for anything (an animation cycle, a physics rate), trace the
  actual pacing: VBlank counters, raster-position waits, blitter waits. Stunt Car
  Racer's race loop has none — it free-runs render-bound, so the honest statement is
  "one step per rendered frame, machine-dependent," with any absolute seconds
  explicitly nominal. While you're at it, check *reachability*: the one 3-VBlank wait
  routine in the binary has zero callers (dead code), and a phase-update block sat
  behind a branch that can never fall through (`MOVE.b #0` sets N=0, so the following
  `BPL` always takes) — verify a branch can go both ways before modelling its body.
- **Round-trip verification.** You haven't decoded something until you've
  re-emitted it and checked: relocations apply cleanly, checksums hold, the image
  renders to a recognisable picture, the extractor's output matches a golden
  test. "Looks plausible" is not done.
- **Separate the layers.** Storage framing (filesystem, sector headers, tape
  block structure) ≠ compression ≠ encryption ≠ the asset format underneath.
  Peel them one at a time and name each explicitly; conflating them is the most
  common source of a wrong writeup.
- **Logical vs physical access.** A file in the directory may be read by the game
  through the OS *or* by raw position on the medium. If the loader touches
  hardware directly, ask *where on the medium* it reads, and whether that maps to
  a catalogued file — the answer is often "the same bytes, reached a faster/
  sneakier way," not "secret data."
- **Register-write greps miss coprocessor-written state.** "The CPU never writes
  `$DFF0A0–$DFF0DC`" does *not* mean "no hardware sprites" — on the Amiga the
  copper writes chip registers from a list the CPU builds as *data* (Marble
  Madness sets all eight sprite pointers, the sprite colours, and `BPLCON2` that
  way; the same trap exists for any DMA-list-driven hardware). When a register
  scan comes up empty, read the **live copper/DMA list** out of emulator RAM —
  it is the ground truth for the display architecture, and one dump settles
  bitplane count, playfield mode, priorities and sprite usage at once. Corollary:
  a mechanism that "explains" the visuals (a depth-sorted list ⇒ painter's
  algorithm) is not proof; trace the draw path to the hardware before writing it
  up — the sort turned out to order *sprite-channel assignment*, and the
  occlusion was a per-frame software mask punched out of sprite DMA data.
- **Read-watch the whole file's RAM extent, summarised per PC.** The read-side
  complement of the write-profiler climb, with aggregation doing the analysis:
  watch the *entire* resident copy of an asset file for one run and report, per
  reading PC, the hit count and the low/high address range. One run maps every
  consumer at once, and the numbers *are* the structure: PCs whose ranges cover
  only the file's head are the directory parser; PCs covering the tail are the
  record walkers; a cluster of PCs whose range starts differ by 4 is one routine
  reading consecutive words (Ridge Racer's six-PC cluster at `+0,+4,…,+20` was
  the packed-vertex loader — 4 × int16 xyz — before any disassembly). Raw
  per-access logging drowns at this scale; the per-PC `{count, lo..hi}` summary
  is what makes watching a 400 KB file tractable.
- **Prove a directory with arithmetic before reading a single record.** When you
  suspect `header + count × entry + Σ(counts × recordSize)`, compute it against
  the file length first: MAP.RRM's `4 + 258×8 + 6737×40 = 271,548` and OBJ.RRO's
  `4 + 319×16 + 440,240 = 445,348` both land on the exact byte count, zero
  leftover — at that point the container structure is essentially proven and
  every record boundary is known. Leftover bytes are equally informative: they
  mean a region you haven't explained yet. This is minutes of work and it
  anchors everything that follows.
- **Expect records shaped as the hardware's own packets.** On console-era 3-D
  titles, asset records are often the GPU/DSP command layout with the file
  fields in command order: Ridge Racer's 40-byte track quads store their
  texture words exactly as the `POLY_FT4` packet wants them, with the packet's
  two *spare* high halfwords repurposed for depth bias and shade; the `.TMS`
  archives are literally the VRAM upload command stream (each record = transfer
  header + dest rect + pixels) played out one block per frame; and the object
  directory reserves a word the loader patches with a runtime pointer. The tell
  is field order matching the hardware spec — once seen, the "decoder" is
  mostly the packet layout, and placement data you feared was hardcoded turns
  out to live in the file itself.
- **Verify at the narrowest seam the engine itself uses.** To prove a decoder
  you do not need to replicate everything downstream of it (projection,
  culling, near-quad subdivision, palette shifts) — trap the exact instructions
  where the engine hands decoded data to a narrow interface, and compare what
  crosses it. Ridge Racer's `geomoracle` breaks on the renderers' own
  RTPT/RTPS/NCT/NCS ops and checks, per event, the record address in the
  pointer register and the GTE data registers just loaded against the Go
  decoder's record at that file offset: bit-exact or fail, 8,000+ events, and
  not one line of the projection pipeline reimplemented. Choosing the seam well
  also keeps the check non-circular: the *register contents* verify field
  interpretation, the *record address* verifies structure — comparing decoded
  bytes against the RAM copy of the file would verify nothing.
- **A state differential needs a timeline.** Comparing rebuilt state (a VRAM
  image, a sound-RAM bank) against the oracle *at the end of a run* fails for a
  reason that isn't a decode bug: later producers legitimately overwrite it.
  Ridge Racer's title screen redraws parts of the texture pages while the last
  archive is still streaming, so the honest differential compares each
  archive's cells at the moment *its own* stream completes, and excludes cells
  a later transfer touched (track upload order per cell). If a differential
  fails, first ask **who else writes the target after your producer finished** —
  the answer converted 4,531 "mismatches" into an exact pass without changing
  the decoder. And the timeline can *fork*: a game may **snapshot hardware
  state and restore it in another scene** — Ridge Racer saves its scenery
  texture quadrant to RAM the moment one archive's stream ends, lets the menu
  paint over it, and restores it at race start, so the race renders from a
  VRAM state that exists at *no single moment* of a linear replay. The
  diagnostic that found the fork without tracing any copy loop: classify each
  mismatching cell by **which replay-prefix state its live value matches**
  (after file 1? file 2? …) — every cell matched the post-`TEX1` prefix, which
  *is* the snapshot point.
- **Know what your instrumentation hides.** Filters you add for signal quality
  become blind spots: muting the read-watch during DMA (so the OT walk doesn't
  drown the log) made an entire file look "never read" — it was consumed purely
  by DMA. An empty watch under a filtered hook is evidence about the filter as
  much as the program; before concluding "nothing reads this," re-check with
  the filter off, or list what the hook deliberately ignores.
- **Count event multiplicities to decompose composite draws.** Bucket trapped
  draw calls per time window and count per-object frequencies before reading
  any code: Ridge Racer's car-select frames showed `{body×1, canopy×2, axle×4,
  underbody×2}` — the ×2s meant *two cars on screen* (the carousel draws the
  neighbour), the ×4 meant a shared axle drawn per corner pair, and the whole
  parts-list structure of the car table fell out of ratios alone. Frequencies
  are a structure probe that costs one run and no disassembly.
- **Placement can be right while the scale is wrong — check for unit forks.**
  Verifying every *record* bit-exact says nothing about the transform that
  places records in the world, and engines routinely keep two unit systems:
  Ridge Racer's positions (camera, cars, the 2048-per-cell grid) are quarter
  model units, converted by a single `sll ...,2` where the grid walk builds
  the GTE translation — missed on first reading, so the exported course packed
  its sections 4× too close while every section rendered perfectly. When
  composed output looks scrambled but the pieces verify, audit the
  translation-building path for shifts/scales, and settle it with ground
  truth: trap the transform's inputs at the coprocessor (Rᵀ·TR recovers each
  section's true world offset; pairwise deltas expose the real cell pitch to
  within rounding).
- **Delay slots execute — simulate them.** On MIPS (and SPARC/PA-RISC), the
  instruction *after* every branch, jump and `jal` runs before control
  transfers, and hand-tracing that wrong sends a format hypothesis into the
  weeds: Ridge Racer's TMS walker "made no sense" for an hour because a
  `j`/`addiu` pair advances the cursor *before* the target label, and a CLUT
  branch's delay-slot `addiu` put the record pointer four bytes past where a
  linear reading said it was. When a hand-derived walk contradicts observed
  behaviour on a delay-slot ISA, re-read every branch with its slot before
  doubting the data.

---

## 5. Reversing behaviour, not just formats

Decoding a *format* (entropy → crib → emulate → round-trip, §4) and understanding
what the code *does* (physics, state machines, game logic) are different modes — and
the second one bit hardest here. The Marble Madness slope/physics trace took many
wrong turns before it converged. What those wrong turns taught, in order of how much
time each would have saved:

- **Trace from the effect backwards; never infer behaviour from the shape of data.**
  The move that finally worked for "what makes the marble roll" was to find the exact
  instruction that *writes the velocity*, then walk backward to its inputs and forward
  to where that data is built. Staring at a byte array and guessing ("this looks like
  a heightmap / these must be the slopes") produced a string of confident, wrong
  stories. If you cannot name the instruction and address that causes the effect, you
  are guessing — say so, or go find it.
- **Decode a concrete instance's real data early; it settles arguments code-reading
  can't.** Hours of reasoning about "the structure" went in circles; decoding one
  actual course's bytes (66 static slope records vs 13 dynamic scripts) settled the
  whole static-vs-scripted question in minutes. Look at the data the code actually
  consumes, for a real instance, not just the code that consumes it.
- **Don't generalise one proven path to the whole system.** A mechanism proven for one
  case (the scripted/dynamic regions, pulling toward a reference point) got written up
  as *the* answer for all slopes — but the static slopes ran an entirely separate code
  path (a height-field gradient). Always ask: does the case I care about actually flow
  through the path I just proved? A plain user question — "is a static slope just a
  one-frame animation?" — exposed the over-reach. Re-derive for each case.
- **Same address/offset ≠ same thing.** One work buffer (`$CCA`) was reused across
  phases at three different strides; one struct offset (`+$1A`) meant different things
  on the marble vs a region vs an enemy. A raw "who writes offset N" scan is therefore
  contaminated — confirm the *base register's identity* (is this pointer really the
  marble?) before trusting a hit.
- **Count from the bound, not by eye.** A dispatch's range check *is* the table size:
  `CMPI.l #$C` ⇒ 12 entries (states 0–11). Reading one slot past the table invented a
  phantom "13th state."
- **A visualisation is a hypothesis test — check it against ground truth.** The
  height-field render became trustworthy only once it reproduced a *known* image (the
  course tilemap) feature-for-feature; doing that comparison caught a real
  horizontal-mirror bug and a profile-indexing error (visible as speckle) that
  code-reading alone had missed.
- **Never fudge data to make the picture look right — a marker in the "wrong" place is
  information, not a defect.** When spawn markers landed off the course mesh, the
  tempting fix was to snap each to the nearest tile; that would have *hidden* the exact
  gap worth seeing. Plot it where the data says. Two outcomes, both useful: it was real
  (the creatures are placed off the path — confirmed because the *placement* objects,
  same grid, calibrated dead-on the course), or it would have flagged a misread (e.g.
  the `+$20` spawner takes its position from a definition table, not the record cell I
  was plotting). Snapping erases both signals. A second-channel **calibration overlay**
  — plot a layer you *know* the ground truth for (course features must sit on the course)
  — turns "is my coordinate mapping right?" into a yes/no you can see.
- **The user's hands-on knowledge outranks a plausible code reading — so when the data
  looks wrong, show it and ask them to look.** "Dizzy" was asserted as a death/respawn
  animation straight from the code; the user, from *playing the game*, knew it is a
  recoverable stun. They also catch placement errors a disassembly can't: someone who
  knows the course can connect "that marker is in the wrong spot" to "oh, that's the
  bird's entry point," pinpointing the misread. State game behaviour as fact only once
  it is traced; otherwise flag it as inference and put the honest (even visibly-wrong)
  artefact in front of the domain expert in the loop.
- **Re-emitting a renderer's output as a portable asset: separate world geometry from
  screen framing.** A game renderer mixes what the scene *is* (world positions, face
  colours) with how it *frames* it (projection scale, a preview-only vertical squeeze,
  screen re-centring, view-dependent recolouring like back-face darkening). Exporting
  "exactly what the game draws" means keeping the first and stripping-but-documenting
  the second — bake the true height:plan ratio, not the preview's squeeze. And when
  the construction is *linear* in its projected inputs (Stunt Car Racer's car is
  fixed fractions of two projected wheel points), a verbatim port becomes a measuring
  instrument: feed it orthographic projections from two axes and the screen-space
  arithmetic hands you the exact 3-D shape it encodes — verify the port against the
  engine's real emitted edges first, then trust the extraction.
- **Keep generated tools deterministic, and verify the artefact, not the run.**
  Map-iteration order made PNG output change run-to-run; a `head` on a regenerate
  SIGPIPE-killed it mid-way and left stale files. `git diff`/`status` on the *committed
  output* caught both — sort before emitting, and check what actually changed, not just
  that the command exited 0.

---

## 6. The executable machine — and when static analysis legitimately stops

A self-built machine is no longer an emergency measure; for every ROM/disc/DOS
platform it is a **standard, first-class deliverable** — a machine model plus a CPU
core, driven by a per-game `bootoracle` (STANDARDS §3). These boot the real image
all the way through the interesting part: the PlayStation core boots Ridge Racer to
its CD-wait loop and renders the attract demo, the 3DO ARM60 machine runs Need for
Speed's `LaunchMe` through self-relocation/BSS into its entry, the DS `dsmachine`
runs both ARM cores, the Game Boy core boots Super Mario Land, and Ultima
Underworld's DOS machine boots into the texture-mapped dungeon. The clean-room
discipline is what keeps even a full boot honest: the oracle is a **guide and a
verifier**, never the data — you read the structure out of the code and reimplement
the decode in Go, using the running machine to confirm it and to supply only what is
provably not on the medium.

That last clause is the narrow case this section is really about. Some state is
simply not on the medium: it is produced by the booted machine.
Marble Madness binds its decryption key to the host **Kickstart ROM** exception-
vector pages *and* the running process's `tc_ExceptCode`/`tc_TrapCode` handler
fields — values that exist only once AmigaDOS has constructed the process. No
amount of reading the disk recovers them.

The escalation ladder, in order of preference:

1. **Boot the ROM on your own core** to read fixed ROM-derived state (a minimal
   reset/early-init run was enough to read the cold-start vector table).
2. **Known-plaintext / linear-algebra** against the cipher to recover as much key
   as the structure pins down (got the header and first segments; stalled where
   too few reliable equations met too many unknowns).
3. **A self-built debugger to capture *only* the irreducible runtime bytes.** For
   Marble Madness that meant building a headless FS-UAE with a GDB stub
   (`tools/platform/amiga/fsuae-debug/`), booting a real Kickstart, and reading
   ~25 bytes of OS state at decode time — then baking those constants into the Go
   decoder. The algorithm stays reimplemented in Go; the emulator supplies and
   verifies a handful of bytes, nothing more.

Capture tricks that helped: patch the `startup-sequence` in place to auto-run the
target from CLI (no GUI/mouse needed), and read the launcher's *own* task fields
live (CLI commands run in-process, so `FindTask(0)` is the right task). Keep the
debugger itself honest — a working continue/breakpoint/step loop is worth the
effort to fix, because it turns a one-shot guess into repeatable measurement.

**Drive the oracle, don't just watch it.** Some state only exists after the player
acts — a menu selection, a name typed at character creation, the first frame of
gameplay. Rather than reconstruct that state by hand, make the oracle **play the
game**: feed it a `-keys` input script and inject the events through the platform's
*own* input path. Ultima Underworld's oracle walks the entire character-creation
flow, types the character's name, and "Journeys Onward" into the first-person
dungeon by pushing key events into the game's own INT 9 keyboard ring buffer
(phase-offset from the timer tick so the game's interrupt-flag logic accepts them)
and mouse events through INT 33h (scale and corner-homing pinned from the game's own
cursor arithmetic). Injecting through the game's real input mechanism — not a
synthetic poke of the resulting variable — is what keeps it clean-room: the game
computes its own consequences, and the oracle only supplies the keystrokes a player
would.

---

## 7. A starting checklist for a new, unknown image

1. `file`/`hexdump` the image; measure size; compute an entropy profile across
   it (flat high-entropy regions = packed/encrypted; structured low-entropy =
   headers/tables/code).
2. Identify the **container**: a signal stream (tape pulses), a sector dump with a
   filesystem (disk), a cartridge ROM (bank/header scheme), an optical-disc image
   (ISO/OperaFS + a platform executable), a DOS MZ program, or a raw dump? Build or
   reuse the container reader.
3. Find the **first code that runs** (autostart vector, boot block, reset
   vector) and disassemble outward from it with recursive descent.
4. Map the **load chain** — what each stage reads, decodes, and hands to the
   next — before diving into any one stage.
5. For any opaque blob: entropy → signature → is it compression or encryption?
   → find and reverse the transform → reimplement → verify by round-trip.
6. Bring up an **executable CPU core** early; use it to run loaders/decryptors
   you can't fully reason about, and to validate every reimplementation.
7. Only when a value is provably not on the medium, escalate to ROM-boot, then
   known-plaintext, then a self-built debugger — capturing the minimum.
8. When the question turns from *format* to *behaviour* (what does this code do?),
   switch modes (§5): trace from the effect — the write to the field that matters —
   backward to its inputs and forward to its data source; anchor every claim to an
   instruction + address; decode one concrete instance's data to check the story; and
   don't assume one proven path covers every case.
9. Write each part up with byte/asm examples as you confirm it, in the neutral
   technical-manual voice the writeups use (STANDARDS §5). Keep raw images and
   copyrighted ROMs out of git (`games/<slug>/image/`); commit only the web
   deliverables (`site/public/<slug>/`) and the figures the writeup embeds
   (`games/<slug>/figures/`); keep every result one command from reproducible.
