# The oracles

An **oracle** is the game's own code, running. Where the static tools (`dis*`, `codetrace*`, the
per-platform container readers) read an image, an oracle *executes* it on our model of the machine
and lets us watch: what it loads, what it computes, what it draws. Everything this repo claims about
a game is either derived from its image by our own code or verified against the game running.

`STANDARDS.md` §3 fixes the **contract** — the flag vocabulary every `bootoracle` shares. This file
is the **inventory**: what each platform's oracle can actually do today, and why each instrument
exists. It is written to be read across sections: an instrument invented for one platform is very
often the thing another platform needs next, and the last section lists the ones worth porting.

Common vocabulary (see STANDARDS §3): `-image`, `-steps`, `-trace`/`-tracen`, `-bp`, `-watch`,
`-keys`, `-shot`, `-savestate`/`-loadstate`. Everything below is *in addition* to that, or a
platform-specific reading of it.

---

## Nintendo 3DS — `games/super-mario-3d-land-3ds/extract/cmd/bootoracle`

Runs both 3DS titles (Super Mario 3D Land, Captain Toad) on `tools/platform/n3ds` + `tools/cpu/arm`
(V6K). The richest oracle in the repo, because the 3DS is the only target where we HLE a whole
operating system (Horizon), a GPU (PICA200, LLE shader + rasteriser), and an audio DSP.

**Execution & state**
| flag | what it does |
|---|---|
| `-steps N` | instruction budget (hex or decimal) |
| `-frames N` | **stop after N VBlanks**, with `-steps` as a ceiling. A graphics workload is measured in frames; an instruction budget only contains one by guesswork, and the guess moves whenever the idle skipping does |
| `-savestate F` / `-loadstate F` | full deterministic snapshot: memory, threads, GPU, GSP, APT, DSP, fs sessions, save store. **The fast-iteration workhorse** — a cold boot to a menu is billions of instructions; a savestate replays it in seconds |
| `-poke ADDR:VALUE` | write a word after `-loadstate`, before running. A probe instrument: falsify a hypothesis by forcing the value the game is waiting for |
| `-threads` | dump every thread (state, pc, sp, lr, what it waits on) + the handle table + pending GX commands. The first thing to run when a boot stops making progress |

**Tracing & breakpoints**
| flag | what it does |
|---|---|
| `-bp ADDR` | halting breakpoint |
| `-logpc ADDR` | **non-halting breakpoint**: log registers (r0–r7, lr, top of stack) and continue. The workhorse for "how often, and with what, does this routine run?" across a billion-instruction boot. Also renders any of r0–r3 that points at a C string, so a `-logpc` on a path builder *names the resource the game asked for* |
| `-tracefrom ADDR` | start instruction tracing when this address is first reached — trace a routine deep in a long boot without drowning in the millions of instructions before it |
| `-watch ADDR[:LEN]` | report every change to a memory word, with the thread and PC that wrote it. Tagged by thread since the port went multi-threaded |
| `-v` / `-svclog` | log every supervisor call and IPC request as it happens / dump the ordered log at the end |

**Finding things in a live machine** — these turn a running game into a searchable object:
| flag | what it does |
|---|---|
| `-findascii STR` / `-findutf16 STR` | locate a string in loaded memory (found the message-archive bug: the dialog's literal `"NULL"`) |
| `-findword HEX` | locate a 32-bit word — including a code pointer, for vtable-driven code no static `BL` reaches |
| `-dump ADDR:LEN` | hex-dump memory after load/run |

**Graphics**
| flag | what it does |
|---|---|
| `-shot BASE` | write both presented framebuffers as PNG (`_top`, `_bottom`), de-rotating the 3DS's column-major panels |
| `-rtshot ADDR:WxH[:FILE]` | decode a tiled render target **straight out of memory**, at the address and dimensions the GPU's own registers name — what the rasteriser drew, *before* any DisplayTransfer moves it. This is what separates "did we rasterise it" from "did it reach the panel", which no counter can: it is how the black-screen chase found that the pixels were landing correctly and were simply being shaded black |
| `-gxdump DIR` | capture GX commands and every PICA command list. Submitted lists are captured **at submission time** (the game reuses list memory, so capturing later reads garbage); buffers the command processor reaches by a **CMDBUF_JUMP** are captured at execution time — the only moment they exist as a unit — and marked `..chained`. Without those you see roughly 1/200th of what the GPU runs |
| `-gputrace N` | per-draw summary: vertex fetch, uniforms, first clip positions, the colour/depth targets and dims, the fragment-lighting block, plus **which uniforms are NaN at draw time** — the instrument that found the float24 bug |

**Performance** — an oracle nobody can afford to run is an oracle nobody runs; see `PERFORMANCE.md`
| flag | what it does |
|---|---|
| `-profile` | time each subsystem and print the run's cost by bucket — command decode, vertex+shader, rasterise, texture decode, GX transfers, DSP, svc/IPC, and the ARM11 as a *derived remainder*. Times only boundaries that are already coarse (a list, a draw, a cache miss, an svc), because a clock read per fragment costs more than the fragment. Totals every frame rather than the last one: the game renders in bursts, and the frame a run stops on is as likely as not to have drawn nothing. Prints the work alongside the time — **fragments per millisecond is the number that moves when an optimisation lands**; milliseconds alone cannot tell a faster rasteriser from a frame that drew less |
| `-cpuprofile F` | write a pprof CPU profile of the run (`go tool pprof -top`). Brackets the run only, not the image load or the PNG writing |

**Audio**
| flag | what it does |
|---|---|
| `-wav FILE` | capture the DSP's final stereo mix (32,728 Hz) for the whole run and write it as a WAV. The verification oracle for anything that makes sound |
| `-dsptrace` | log every source configuration the DSP consumes and every status it publishes. The app↔DSP voice conversation happens entirely in shared memory with **no IPC to log**, so without this it is invisible |

**Input**
| flag | what it does |
|---|---|
| `-keys a,b,x,y,l,r,up,down,…` | inject pad state into the HID shared-memory ring, published each VBlank |
| `-keypulse N` | release the injected keys briefly every N frames, so a *fresh press edge* keeps arriving. Required to advance multi-screen dialogs — a held button gives one edge, which the open animation swallows |
| `-hidtrace` | tally the game's own reads of the HID block by offset — how the ring layout was reverse-engineered rather than guessed |

**Companion static tools:** `n3dsdump` (containers, RomFS; **`-at OFFSET` names the RomFS file behind
a traced raw read** — the instrument that showed Captain Toad loads its opening stage but none of its
object models), `picadump` (`-hist`/`-shader`/`-reg`: decode a captured command list, disassemble the
vertex shader), `msgtool` (message archives), `bannerdump`.

---

## Sony PSP — `games/{loco-roco,burnout-legends}-psp/extract/cmd/bootoracle`

Allegrex (MIPS + VFPU) on `tools/platform/psp`. Both titles share the flag set; Burnout adds
`-shotat`. **The oracle plays the game**: Loco Roco is driven from cold boot through language select,
title, dialogue and into tilt gameplay entirely by scripted pad input.

| flag | what it does |
|---|---|
| `-exe` | boot a specific EBOOT/PRX rather than the disc default |
| `-keys FILE` | **pad script** — a timed sequence of presses. This is what "the oracle plays the game" means in practice |
| `-tracethread N` | restrict tracing to one thread — essential once a title runs a dozen kernel threads |
| `-rwatch ADDR` | **read**-watch (who *reads* this?), the complement of `-watch` |
| `-rprofile` / `-watchn` | profile reads over a range; limit watch hits |
| `-gelog` / `-gedump` | log and dump GE (GPU) display lists |
| `-shot` / `-shotat N` | screenshot at end / at a given instruction |
| `-dis ADDR` / `-dump` / `-dumpbin` / `-find` | disassemble, hex-dump, dump raw, search live memory |
| `-notes` | annotate a run |
| `-savestate` / `-loadstate` | snapshot (carried over from day one per the oracle-parity rule) |

---

## Nintendo 64 — `games/pilotwings-64-n64/extract/cmd/bootoracle`

LLE RSP + RDP on `tools/platform/n64` + `tools/cpu/r4300`. Pixel-perfect on the attract sequence.
Now kept mainly as a *verification harness* — the webexport reads the cartridge directly.

| flag | what it does |
|---|---|
| `-shot` / `-shotevery N` / `-shotbase` | screenshot; periodic capture through a sequence |
| `-stopfield N` | stop at a given video field — deterministic frame targeting |
| `-dmalog` | log cartridge DMA: which region loads when, the map of the ROM's own loader |
| `-pcmdump` | dump decoded audio |
| `-calllog ADDR` | log calls to a routine (repeatable) |
| `-rwatch` | read-watch |
| `-keys` | controller script |

**Companions:** `rdpdbg` (`-px X,Y` — *click a pixel, get the RDP command that drew it*, plus
`-dumpram`), `dlwalk` (walk display lists), `dmamap`, `texdump`. Plus the **frame debugger**
(`tools/cmd/framedbg`), below — an interactive front-end onto the same machine.

---

## The frame debugger — `tools/cmd/framedbg`

Every other instrument in this file answers a question you already knew to ask. The frame debugger is
for the questions you *don't* — you watch the frame being built and see what it does. It is the one
oracle front-end with a user interface, and it is deliberately platform-agnostic: `tools/debug`
defines a small `Target` interface plus a set of *optional capabilities*, and the page builds itself
from whatever the current target says it can do — no empty panels, nothing faked. Seven adapters back
it today: `n64adapter`, `psxadapter`, `n3dsadapter`, `ndsadapter`, `pspadapter`, `threedoadapter` and
`gcadapter` (the RDP, GP0, the PICA200, the DS's geometry engine, the GE, the 3DO's cel engine and
Flipper), plus `dosadapter`, which is CPU-only and has no frames at all.

```
framedbg -image ROM [-state FILE] -serve :8088     # the interactive debugger, in a browser
framedbg -image ROM -list | -scrub N | -pixel X,Y  # the same pipeline, headless and scriptable
```

**The UI is a local web page the oracle binary serves itself** — no cgo, no GUI toolkit, no build
step, no network. `tools/debug/wsock` is a server-side RFC 6455 implementation in stdlib Go, so the
`tools` module keeps its zero-dependency property. Framebuffers go over the socket as raw binary
behind a 16-byte header, aligned so the page wraps the bytes in an `ImageData` with no copy.

| what it does | how |
|---|---|
| **Frame-step** | one video field at a time, or "step to a drawn frame" — skip the idle fields a boot is full of |
| **Play** | free-run the machine and stream the scanout (~20 fps on Pilotwings), capturing nothing. How you fast-forward to the part of the game you want to look at. Pausing lands on a full capture |
| **Command scrub** | drag through the frame's RDP command stream and watch the picture assemble, command by command. On a target that backs `debug.BatchReplayer` (3DS), the positions a drag is about to land on are replayed **in parallel on independent scratch machines** and cached — 2.3× on an eight-position drag. Free of any determinism question: each replay restores the same snapshot into a machine of its own and throws it away |
| **Click a pixel → the command that drew it** | plus its **full overdraw history**, including the writes the rasteriser produced and then *threw away* on a depth or alpha test — usually the answer to "why is this pixel not the colour I expect?" |
| **Select a command → every pixel it drew** | highlighted as an overlay |
| **Inspect** | CPU registers, RDRAM hex |
| **Profile** (3DS, PSP, 3DO, GameCube) | where the stepped frame's time went, by subsystem — one stacked bar plus the counters. Capability-gated on `debug.Profiler`, so a target that cannot honestly time its own subsystems simply has no panel rather than an empty one reading "0 ms". The times are the *machine's*, not a sampling profiler's: only the emulator knows which nanoseconds were its rasteriser and which its DSP. Counters ride alongside, because a bucket that got faster while the frame drew less has not got faster. **Which boundaries a platform may time is decided by counting them, not by taste** — a GameCube field is 8.8k FIFO bursts and 9.8k draws (timeable, 0.9% overhead) but 3.4M fragments (not), which is why it has no texture bucket where the 3DS does |

**Why the pixel questions are instant.** A frame capture carries a *provenance buffer* — one command
index per pixel — built in a single pass by the machine's own `OnPixel` hook while the frame draws.
It is handed to the page once per capture, so hovering a pixel or highlighting a command's coverage
is a local array lookup, not a request to the emulator.

**Why the scrubber is exact rather than approximate.** "The framebuffer after command k" is produced
by *replaying the frame* from an in-memory snapshot taken at its start, halting right after command k
(`Machine.RunStopAfterRDPCommand`, which unwinds the nested RSP/CPU interpreter loops with a sentinel
panic). Execution is deterministic, so this is the real thing — no per-command framebuffer copies, and
FILL/COPY writes straight to RDRAM are handled correctly, which a history-replay could not do. The
replay restarts from the snapshot every time (the scratch machine is left mid-queue and is
discarded), so the session caches the replays it has already paid for and collapses a fast scrubber
drag to the position the mouse actually landed on.

**Why it is trustworthy.** The image the socket hands the browser for command k is verified
byte-identical to the headless `RenderAfter(k)` on the same frame — the UI cannot quietly diverge
from the core.

**What the 3DS adapter had to do differently**, because it is the first target whose display processor
is not a packet machine:

- **A "command" is one register write.** A PICA200 frame is a *register-write stream*, not a command
  list: the draw happens when the stream writes `0x22E`, and everything about it — which buffers the
  vertices come from, which shader runs, whether colour is written at all — is state that the
  preceding writes latched. So the scrubber's unit is the write (Captain Toad's opening stage: ~99k of
  them in one frame), and scrubbing backwards from a draw is how you read the state it was made under.
  Provenance still names only the draws, because only a draw can produce a pixel.
- **The frame's picture is the render target, not the screen.** The GPU renders into a tiled, padded
  VRAM buffer (Captain Toad's is 256×512 at `0x1F000000`); a `DisplayTransfer` later copies a rotated
  240×400 crop of it to the LCD. They are different planes — different size, different orientation,
  different contents — so the debugger builds the frame around the one the commands actually wrote,
  and offers the scanout alongside it as a surface. **Comparing the two is how you catch a frame the
  GPU drew and the transfer never delivered.**
- **A frame ends at the VBlank**, where the GSP consumes the buffer swap — the 3DS has no scanout the
  CPU can watch.
- **The main pass is chosen by pixel census**, not by whichever buffer the register file happens to
  point at when the frame ends: a frame can render a shadow pass into its own target, and at the
  VBlank the registers still name whatever the last command list left behind.

The last two mattered immediately. Provenance is a plane of *one* target, and the page had been
indexing it by the displayed image's width — which is only correct where the scanout and the draw
target coincide, as they do on N64 and PSX. On the 3DS it would have silently reported the wrong
command for every pixel. The page now carries the provenance plane's own dimensions and refuses to
answer when they disagree with the picture on screen.

**What the DS adapter had to do differently**, because it is the first target whose rasteriser does
not run while the commands do:

- **The frame is TWO SCREENS.** A DS game puts its world on one panel and its menu, its map, its
  stylus target on the other; a debugger showing one of them shows half the picture. So the frame is
  the console's own composition — 256×392, the top panel above the bottom one with the bezel's gap.
  Which engine drives which panel is not a constant: POWCNT1 bit 15 says, and SM64DS flips it, so the
  composition asks the machine rather than assuming.
- **The scrubber replays the GEOMETRY, not the picture.** The DS's geometry engine accumulates
  polygons all frame and the rasteriser runs once, at the buffer swap. There is no partial framebuffer
  to halt at, because there is no framebuffer yet — so the scrubber stops after command k and then
  *rasterises the polygon list as it stands*. Dragging it watches a frame's geometry accumulate, which
  on this machine is the right question: the geometry **is** the frame, and the rasteriser is a
  formality that happens at the end. What it shows at command k is not "the frame, half-drawn"; it is
  "the frame, if the game stopped submitting geometry here".
- **The halt is a sentinel panic, not a flag.** One mode-7 DMA burst pushes 112 words into the GXFIFO
  *inside the store that triggered it*, so a scrubber that could only stop between instructions would
  overshoot its position by most of a display list — and the picture it drew would not be the picture
  it claimed. The panic unwinds out of the middle of the DMA, which is safe precisely because the
  machine it unwinds is a scratch replay about to be thrown away.
- **Provenance is gated on visibility.** Every polygon remembers the command that closed it, so a
  pixel names a command — but only where the 3D layer actually *won* the 2D compositor's priority
  fight. A 3D pixel with a background on top of it is one the player cannot see, and naming a draw for
  it would be a confident wrong answer, which is the worst kind a debugger can give.

Clicking Mario's face on SM64DS's title screen (12,311 geometry commands in the frame) names
`VTX_XZ` #11679 and shows seven writes to that pixel — one of them depth-rejected.

**What the GameCube adapter had to do differently**, because it is the first target whose draw target
is not in memory at all:

- **The frame is the COPY, not the field.** Flipper draws into an embedded framebuffer on the die —
  the EFB, 640×528 RGBA — and the pixel engine then copies that out to the external framebuffer the
  video interface scans, as 640×480 YUY2 in main RAM. The copy is the flip and it is where a frame
  ends. The video field is not: it is a scanout clock that ticks whether or not the game drew, and
  the copy that ends a frame **clears the EFB behind itself** — so a capture taken at the field
  boundary finds the frame's own draw target already wiped. This is the third form the "render the
  draw target, not the scanout" lesson has taken (N64/PSX: double-buffered; 3DS: different planes;
  here: different *memories*, and one of them is erased a microsecond after the frame is finished).
  The picture is therefore taken **inside** the copy, through an `OnFlip` hook that runs before the
  clear — in `StepFrame` *and* in `RenderAfter`, because the scrubber's last position is the closing
  copy and would otherwise be black.
- **No sentinel panic is needed.** The FIFO is drained by an ordinary Go loop that the write-gather
  pipe feeds, so the interpreter simply declines the next command and leaves it queued
  (`Machine.RunStopAfterGXCommand`) — the N64 and DS need the panic only because their command lists
  execute *inside* one CPU instruction.
- **The disc is a real filesystem**, so this is the second target after the 3DS to back `FileLister`
  and `FileAttributer`: a read-watch on the drive can name the game's own file being streamed.
- **The pad is the Keyer's second platform, and it inverted the interface's hazard.** A PC keyboard
  *delivers events* — make and break are each a scancode the guest's ISR consumes, so the rule there
  was "never coalesce". The GameCube's pad *is a level*: the serial interface samples it once per
  field and the game edge-detects presses off those samples. Nothing on the guest side queues, so
  nothing self-paces, and a browser key that goes down and up between two field samples is a press
  the game never sees. The adapter therefore queues button states and releases **one per field**,
  which is the touch latch's logic arriving from the opposite side: there the risk was more states
  than frames, here a state with no field to be sampled in.

A frame of the intro cutscene is 12,770 FIFO commands over a 640×528 draw target, 2,959 of them
writing pixels; clicking Luigi at (340,390) names `DRAW_TRIANGLEFAN` #1103 and shows 45 writes to
that pixel, 44 of them depth-rejected.

**What its profiler says** — measured before back-face culling landed, so it is also the worked
example of the profiler being used — and what deciding the buckets by measurement rather than taste
looked like. One field of the cutscene contains 8,785 FIFO bursts, 12,770 GX commands (9,774 draws),
3,644 DSP batches — and ~3.4 **million** fragments. The first three are timed; the last is three
orders of magnitude too many, so unlike the 3DS **there is no texture bucket** (the 3DS times texture
*cache misses*; this machine has no texture cache, so every sample would be a clock pair). Measured
overhead: 0.9%. A field comes out:

```
  rasterise                  234.61 ms  71.5%   (9,774 draws)
  gekko + rest (derived)      74.80 ms  22.8%
  dsp                          8.15 ms   2.5%   (5,466 batches)
  vertex + xf                  4.89 ms   1.5%   (9,774 draws)
  command decode (derived)     3.72 ms   1.1%   (8,822 bursts)
  pe copy                      1.85 ms   0.6%   (3 copies)
  · fragments drawn 1,155,366   · depth-rejected 6,650,838   · alpha-rejected 132,234
```

The counters are the point of the counters: the rasteriser decides **7.9 million fragments to keep
1.2 million** — 85% of the frame's dominant cost thrown away by the depth test, on a rasteriser that
did not yet cull back faces. **That reading was right, and acting on it made the rasteriser 28%
faster** (181.7 → 130.6 ms; the whole field 268.4 → 219.6 ms, −18%), by culling 8,264 of the field's
triangles and 3.6M of its fragments before they were ever shaded. See "back-face culling" below.

Two things about the nesting, because they are where this kind of profiler goes wrong. "Command
decode" is *derived* — the whole FIFO drain is timed at its one entry point and the draws and copies
inside it are subtracted out, because timing 12,770 individual register loads would double the cost
of the cheapest thing in the frame. And the frame boundary happens **inside** the timed FIFO burst
(the copy that ends a field is a command in it), so `profFrame` closes and restarts the burst's timer
and the copy books its conversion *before* calling it. That is the PSP/3DO bug — work nested in a
frame's closing call, booked after the accumulators reset, landing in the next frame — arriving in a
third shape. Note the sum test cannot catch it here, precisely *because* decode is derived and
absorbs it; the guard is a direct unit test on the timer mechanism instead.

**Back-face culling, and why a speedup is not evidence.** The profile above named the target, and
implementing it was worth **−28% on the rasteriser and −18% on the whole field**. But it is a
*correctness* change wearing a performance change's clothes, and it is the cleanest example in this
repo of a measurement that cannot check its own work:

- **Both** sign conventions cull ~8,000 triangles. **Both** make the rasteriser ~30% faster. In
  closed geometry half the triangles face each way, so culling the *wrong* half is exactly as cheap
  as culling the right one — and every counter reports a triumph either way. Rendering the frame is
  the only test that can tell them apart: the right sign changes 9.4% of the picture, the inverted
  one changes **92.7%** and you see through the tree trunks and into the back of Luigi's head.
- The first guess was the wrong one.
- What the disassembly *can* pin: the field (GEN_MODE bits 14..15) and the fact that the SDK's enum
  and the hardware disagree — the game's own `GXSetCullMode` (0x801F51C0) is almost entirely a
  `1↔2` swap. What it *cannot* pin is which winding is a front face, because that depends on our own
  projection (the viewport flips Y, which flips every triangle's area sign). So the constants are
  named `cullPosArea`/`cullNegArea` — for the sign they discard, which is verified — rather than
  front/back, which would be asserting a mapping nothing has checked.
- The residual 9.4% is culling *fixing* things: the flashlight's cone had been blended in twice
  (both faces), and the tree silhouettes had slivers of back face peeking round them.

`RR_GC_CULLMODE=off|flip` is kept as the scaffolding that settled it, and as the first thing to try
when geometry goes missing from a scene.

---

## PlayStation — `games/ridge-racer-psx/extract/cmd/bootoracle`

MIPS R3000 + GPU on `tools/platform/psx`. Boots into the race; savestates were built specifically so
the driving-physics work could iterate.

| flag | what it does |
|---|---|
| `-gplog` / `-gpfrom N` / `-gpop OP` | log GPU primitives, from a given command, filtered by opcode |
| `-dmalog` | DMA log |
| `-vram FILE` | dump VRAM (textures, CLUTs, the framebuffer) |
| `-isr` | trace interrupt service routines |
| `-tty` | the game's own debug output |
| `-press BTN` | inject a button |
| `-poke ADDR:VAL` | force a value (used to reach the reversed/EXTRA course variants) |
| `-save` / `-load` | savestates |
| `-rwatch` / `-watchn` | read-watch; bounded watch |

**Companions:** `geomoracle` (bit-exact geometry verification against the game's own transforms),
`calltrace`.

---

## 3DO — `games/need-for-speed-3do/extract/cmd/bootoracle`

ARM60 (big-endian) + a full Portfolio-OS HLE and a software cel engine. The oracle reaches in-race
rendering. Its flag set is the most *experimental* in the repo — a good place to look for ideas.

| flag | what it does |
|---|---|
| `-hot` | hot-PC profile: where is the CPU actually spending time |
| `-spinbreak` | break out of a detected spin loop — turns "hung forever" into "here is what it spins on" |
| `-stall` | report stalls |
| `-vblmirror` | mirror the VBL counter (the contract that unblocked the boot) |
| `-movies` | enable the movie/DataStreamer path |
| `-pad` | scripted pad input |
| `-celdebug` / `-sportdebug` | cel-engine and SPORT (VRAM blit) debugging |
| `-persptint` | tint by perspective term — *visualise* the cel engine's HDDX/HDDY maths |
| `-probex` / `-probey` | probe a specific screen coordinate |
| `-shots` / `-shotevery` / `-shotfrom` | capture a sequence of frames |
| `-fbbase` | override the framebuffer base |

**Companions:** `geomoracle`, `memtrace`.

---

## MS-DOS / x86 — `games/ultima-underworld-pc/extract/cmd/bootoracle`

8086/286 + DOS + BIOS + VGA on `tools/platform/dos`. **The oracle plays the game into the dungeon**:
it walks character creation, types a name, and Journeys Onward into the first-person Stygian Abyss.

| flag | what it does |
|---|---|
| `-keys FILE` | **keyboard + mouse script** — injected through the game's *own* INT 9 ring buffer (phase-offset from the timer so the interrupt flag is right) and INT 33h, with corner-homing for absolute clicks |
| `-irq N` | drive/trace a hardware interrupt |
| `-bpal SEG:OFF` | breakpoint in segment:offset form (addresses are the platform's natural shape) |
| `-vgaprof` / `-profrange` | **write-profiler**: who writes each pixel/region of VGA memory. The basis of the "write-profiler climb" technique — RE a pipeline by repeatedly profiling *who produced this value*, up the call stack |
| `-rdprof` / `-rdrange` | the same for reads |
| `-texid` / `-texout` | dump a texture by id |
| `-loadsave` | boot into a saved game |
| `-dis` / `-dump` / `-find` / `-poke` | disassemble, dump, search, force |

---

## Nintendo DS — `games/{mario-kart-ds,super-mario-64-ds}/extract/cmd/bootoracle`

Dual ARM9/ARM7 on `tools/platform/nds/dsmachine` + `tools/cpu/arm` (V5TE). **The oracle boots the
game and draws it**: Super Mario 64 DS runs from cold through its OS bring-up, reads 2,731 files off
the cartridge, and renders its title sequence — the Nintendo legal screen, the TOUCH TO START star,
and (after a scripted stylus tap) the SUPER MARIO 64 DS logo with Mario's 3D face and the menu.

Unlike the 3DS there is no operating system to HLE, only hardware, so this is an LLE machine like
`n64` and `psx`: scanline timing, eight DMA channels, eight timers, the ARM9's hardware divider and
square-root units, the nine VRAM banks and the mapping register that decides what each of them
currently *is*, the cartridge port, the ARM7's SPI bus (a synthesised firmware, the power chip, the
touchscreen), and both graphics engines — the two 2D controllers and the 3D geometry/rasterising
pipeline. Only the BIOS's software interrupts are lifted above the metal (they are a library, not a
kernel), as the PSX BIOS is.

**Execution & state**
| flag | what it does |
|---|---|
| `-frames N` | stop after N frames — a graphics workload is measured in frames, not instructions |
| `-savestate F` / `-loadstate F` | full snapshot. **The workhorse**: a cold boot to the title screen is 1.2 billion scheduler steps and 42 seconds; restoring it takes one. Captures the ARM's *banked* registers, without which the first interrupt after a restore runs on a stack pointer of zero |
| `-steps N` / `-quantum N` | instruction budget; how long a core runs before the other gets a turn |

**Input** — the oracle plays the game
| flag | what it does |
|---|---|
| `-keys FILE` | **a timed input script**: `320 touch 128,120` / `340 release` / `120 press a,start`. The point is the press EDGE — a DS game asks "did this go down since last frame", so a stylus held from reset gives the title screen nothing and it waits for ever |
| `-keys a,start` | or just buttons to hold, when that is enough |
| `-touch X,Y` | hold the stylus for the whole run |

**Graphics**
| flag | what it does |
|---|---|
| `-shot BASE` / `-shotevery N` | both screens as PNG (`_top`, `_bottom`), with POWCNT1's engine→panel swap applied — "the top screen" is a question about a register, not about an engine |
| `-rtshot F` | the **3D engine's own render target**, before the 2D engine composites it as engine A's BG0. Pixels the rasteriser never touched come out magenta, so "drew nothing" cannot be mistaken for "drew black" — the DS's answer to the 3DS's `-rtshot`, and for the same reason |
| `-gfx` | both engines, the VRAM bank mapping, and the 3D engine's frame: polygons at the last swap, and how many primitives the clipper rejected. *95% clipped* is a transform bug; *0 polygons* is a geometry bug; they are not investigated the same way |
| `-gxdump` | histogram of the 3D commands actually executed. A 3D game that never issues `MTX_MULT` is not a 3D game — it is a FIFO that is dropping words |

**Loading & tracing**
| flag | what it does |
|---|---|
| `-cardlog` | every cartridge transfer: command, ROM address, size. The map of what the game loads, when, drawn by its own loader |
| `-io` | the I/O registers the run programmed |
| `-log` | **the hardware the model did NOT implement.** The honest half of every run: a register this machine does not model is logged, not quietly read back as the last value written, because on a machine whose boot polls status bits a stub that happens to read "ready" is indistinguishable from working silicon right up until the frame it isn't |
| `-bp` / `-logpc` / `-trace` / `-tracefrom` / `-dump` | halting and non-halting breakpoints, tracing, memory |

**The frame debugger** — `framedbg -image rom.nds -dtcm 023C0000 -state title.state -serve :8088`.
The DS backs the full 3DS capability set; what it had to do differently (a frame that is two screens,
a scrubber that replays *geometry* because the rasteriser only runs at the buffer swap) is in the
frame-debugger section above.

**Declared gaps** (all logged by `-log`, none faked): the sound *mixer* — the register file is real and
the ARM7's sound driver runs and sequences, but no samples are fetched; display capture; anti-aliasing;
and toon-shaded polygons currently render pink (the toon table and the skin palette are both correct in
VRAM, so the fault is in the shading, not the data).

**Companion oracles**, still the right shape for questions about *data*:
- **`actororacle`** (SM64DS) — the strongest "reimplement, don't scrape" instrument in the repo. It
  runs the game's own actor create/init code natively for each of 4,048 of 4,350 actors, so behaviour
  comes from the game rather than from heuristics. `-boot`, `-actor ID`, `-ovl`, `-par`.
- **`dualoracle`** — the original dual-CPU scheduling harness (`-budget`, `-quantum`, `-log`).

---

## Amiga (68000) — `games/{stunt-car-racer,marble-madness,turrican}-amiga`

The 68k work uses **many small, single-purpose oracles** rather than one big one: each boots the game
and asks it one question, then we reimplement the answer in Go and use the oracle to verify.
Stunt Car Racer alone has nine:

| oracle | the question it answers |
|---|---|
| `geomoracle` | the baked track geometry (8/8 circuits byte-exact) |
| `coloracle` | preview colours and decals (8/8) |
| `caroracle` | the opponent car's procedural screen-space build (edge-exact) |
| `modeloracle` / `planoracle` / `spineoracle` | model, plan and spine data |
| `horizonoracle` | the horizon |
| `bridgeoracle` | the Draw Bridge animator |
| `physoracle` | the rigid-body car simulation (`-frames`, `-input`) |
| Marble Madness `sndoracle` | the sound engine (`-course`, `-id`, `-secs`, `-music`) |

This is the pattern to reach for when a game's *data* is procedural: don't reverse the algorithm from
the disassembly alone — run it, and check your reimplementation against it.

---

## Game Gear / Game Boy / C64 — probe-shaped oracles

Same philosophy, smaller machines: rather than one boot oracle with many flags, each question gets a
probe that boots the ROM and settles it.

- **Game Gear (Sonic)** — `oracleshot` (`-act`, `-settle`: boot, settle N frames, screenshot),
  `leveltrace`, `screentrace`, `objsettle`, `spawncheck`, `enemyprobe`, `waterprobe`, `animprobe`,
  `soundprobe`, `logprobe`.
- **Game Boy (Super Mario Land)** — `tools/platform/gameboy` machine (Sharp LR35902, *not* a Z80);
  `spawntrace` (`-id`, `-frames`), `spawnverify`, `objscript`, `tileanimverify`.
- **C64 (Elite, Fort Apocalypse)** — `tools/platform/c64` + a 6502 core; `galaxytrace` (the procedural
  galaxy, run rather than guessed), `enginedump`, `paceprobe`. The SID emulator (`tools/c64/sid`)
  renders the music by running the player.

---

## Cross-platform: what exists once and should exist everywhere

Every CPU has a matching pair of static tools — `dis<cpu>` and `codetrace<cpu>` (6502, 68k, z80, sm83,
x86, mips, allegrex, r4300, rsp, arm, arm60) — and every platform machine implements `Read`, `Run`,
savestates and a framebuffer capture. Beyond that, these instruments were invented for one platform
and are **candidates to port**:

| instrument | born on | worth porting because |
|---|---|---|
| **Savestates** | all | mandatory (`oracle-capability-parity`): they turn a 40-minute cold-boot gate into a seconds-long one. Any platform whose regression gate is a cold boot is paying this tax now |
| **`-logpc` with string rendering** | 3DS | names *what* the game asked for, not just where it asked. Any oracle that traces a resource loader wants this |
| **`-findascii` / `-findword`** | 3DS | search a *live* machine — the only way into vtable-driven code no static `BL` reaches |
| **`-at OFFSET` → filename** | 3DS | turns a raw archive read into a filename. Every platform with a big packed archive (PSP, N64, PSX, DS) has this problem |
| **Write-profiler (`-vgaprof`)** | DOS | "who wrote this value" up the call stack; catches self-modifying patchers. Would suit any framebuffer or command-ring investigation |
| **`-spinbreak` / `-hot`** | 3DO | turns "hangs forever" into "here is the loop and why". The 3DS spent three sessions doing this by hand |
| **Read-watch (`-rwatch`)** | PSX/PSP/N64 | the complement of `-watch`; the 3DS oracle still lacks it |
| **The frame debugger (`framedbg -serve`)** | N64, PSX, 3DS, DS, PSP, 3DO, **GameCube**, and DOS (CPU-only) | watch a frame being built, command by command; click a pixel to get the command that drew it and its overdraw history. Needs four things of a platform — a per-command hook, a per-pixel hook, a frame hook, and a **deep** in-memory snapshot — and every LLE machine here can offer them. Every platform after the N64 needed all four added; every one of them also had to answer "which buffer is the frame?" differently, and getting that wrong is the one mistake that still looks right |
| **The "gap log" (`-log`)** | DS | an I/O register the machine does not implement is *logged*, not quietly read back as the last value written. Every stub that reads "ready" is a lie the boot believes until the frame it doesn't, and the log is what turns a run's reach into a claim you can check. Cheap to add anywhere |
| **`-rtshot` / the render target on its own** | 3DS, DS | a black screen is two different bugs wearing one face — the GPU drew nothing, or it drew and the compositor threw it away. No counter separates them; looking at the plane the rasteriser wrote does, instantly |
| **`-poke`** | PSX/3DS | falsify a hypothesis in one run by forcing the value the game waits for |
| **Pad/key scripts (`-keys`, `-keypulse`)** | DOS/PSP/3DS | an oracle that can *play* reaches states no boot ever will. `-keypulse` (fresh press edges) is the non-obvious part |
