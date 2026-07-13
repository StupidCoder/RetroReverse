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
defines a `DebugTarget` interface, `tools/debug/n64adapter` implements it, and any platform whose
machine can capture a display-list and attribute pixels can follow (PICA200 next).

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
| **Command scrub** | drag through the frame's RDP command stream and watch the picture assemble, command by command |
| **Click a pixel → the command that drew it** | plus its **full overdraw history**, including the writes the rasteriser produced and then *threw away* on a depth or alpha test — usually the answer to "why is this pixel not the colour I expect?" |
| **Select a command → every pixel it drew** | highlighted as an overlay |
| **Inspect** | CPU registers, RDRAM hex |

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

Dual ARM9/ARM7 on `tools/platform/nds/dsmachine`. Deliberately minimal (`-image -steps -io`), because
the DS work is driven by two *special-purpose* oracles instead:

- **`actororacle`** (SM64DS) — the strongest "reimplement, don't scrape" instrument in the repo. It
  runs the game's own actor create/init code natively for each of 4,048 of 4,350 actors, so behaviour
  comes from the game rather than from heuristics. `-boot`, `-actor ID`, `-ovl`, `-par`.
- **`dualoracle`** — dual-CPU scheduling harness (`-budget`, `-quantum`, `-log`).

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
| **The frame debugger (`framedbg -serve`)** | N64 | watch a frame being built, command by command; click a pixel to get the command that drew it and its overdraw history. Needs three things of a platform — a per-command hook, a per-pixel hook, and an in-memory snapshot — and every LLE machine here could offer them. `tools/debug` is platform-agnostic and wants adapters (PICA200 next) |
| **`-poke`** | PSX/3DS | falsify a hypothesis in one run by forcing the value the game waits for |
| **Pad/key scripts (`-keys`, `-keypulse`)** | DOS/PSP/3DS | an oracle that can *play* reaches states no boot ever will. `-keypulse` (fresh press edges) is the non-obvious part |
