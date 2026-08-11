# Captain Toad: Treasure Tracker (Nintendo 3DS)

The repo's **second Nintendo 3DS title**, and the first real test of whether the 3DS work built for
*Super Mario 3D Land* is a **platform** or just one game's scaffolding. It is the same hardware — an
ARM11 (ARMv6K) application processor under the Horizon microkernel — so the whole toolchain
(`tools/platform/n3ds` containers, machine and service tree; `tools/cpu/arm` in its V6K variant) and the
oracle apply unchanged. Nothing here is game-specific until the frontier.

Image: `Captain Toad - Treasure Tracker (Europe) (En,Fr,De,Es,It,Nl).cci`, 512 MiB, MD5
`c52bafa56dadc8777b0b14a151d0ba51`. A **decrypted** dump (the NCCH NoCrypto bit is set), which is the
only kind this project can read — the AES-CTR keys of a retail dump are console state, and none are
embedded, so an encrypted image is refused rather than faked.

## Part I — The cartridge, for free

The container layer read it with no changes at all: NCSD → four partitions, NCCH → ExHeader/ExeFS/RomFS,
the BLZ-compressed `.code` decompressed, the RomFS tree listed. The title is noticeably bigger than
SM3DL — `.code` is 0x36A310 of text against SM3DL's 0x2F4000 — and its RomFS carries the expected asset
trees (`/StageData`, `/TextureData`, `/UIX2`, …).

That the container work transferred with zero effort is the first evidence that the 3DS layer is a real
platform port rather than a single-title hack.

## Part II — What a second title finds: two holes in the CPU

The boot then halted twice inside its first 400,000 instructions, both times on ARMv6K instructions
that Super Mario 3D Land simply never executes. These were the honest "lazy gaps" left from the
original core bring-up — the decoder recognised them, the executor refused to guess — and a second
title is exactly the thing that finds them.

- **`LDREXD` / `STREXD`**, the 64-bit members of the load/store-exclusive family. The value is a
  register *pair*, `Rt` and `Rt+1`, low word first, and the subtlety is that **`Rt2` is implicit**: it
  is not encoded anywhere in the instruction. The exclusive monitor tags the base address exactly as
  the word-sized forms do, so an *untagged* `STREXD` must report failure and leave memory untouched —
  the property the lock code's retry loop is built on.
- **The parallel (SIMD) add/subtract group**, of which only the two truncating classes existed. Three
  things vary independently, and the encoding keeps them apart: the lane width and pairing (`ADD16`,
  `ASX`, `SAX`, `SUB16`, `ADD8`, `SUB8` — the two "exchange" forms cross the halfwords of `Rm`),
  signedness, and **what happens to a lane result that will not fit**: truncate (`S`/`U`, which publish
  the per-lane `GE` flags a following `SEL` consumes), saturate (`Q`/`UQ`), or halve (`SH`/`UH`, which
  shift the full-precision result right by one so no lane *can* overflow). Only the truncating classes
  write `GE`; the others must leave it alone. Captain Toad halted on `UQSUB8` — a byte lane that must
  clamp at zero where the plain `USUB8` wraps to `0xFF`, which is the kind of difference that produces
  quietly wrong audio or pixels rather than a crash.

Both are now implemented and unit-tested (the pair semantics, the untagged-`STREXD` failure path,
`UQSUB8` clamping where `USUB8` wraps, negative saturation, the halving forms, both exchange forms, and
that the saturating classes do not clobber `GE`). The DS regression suite still passes on the shared
core.

## Part III — The OS handshake, and one new service each

Past the CPU, the boot walked the OS handshake the SM3DL work already built — `srv:`, the APT applet
lifecycle, `ndm`, `cfg`, `hid`, `fs` — and stopped only where this title asks for something SM3DL never
did. Each was settled by disassembling the game's own IPC wrapper rather than guessing:

| service | command | what the wrapper shows |
|---|---|---|
| `APT:A` | `0x0055` | header `0x00550040`, one byte in, reply is the result word only — a plain setter |
| `fs:USER` | `0x0861` | header `0x08610042`: an SDK version, plus the constant `0x20` — the **ProcessId** translate descriptor the kernel fills in. Session init |
| `fs:USER` | `0x0862` | header `0x08620040`, one word in, result only |
| `APT:A` | `0x0101` | header `0x01010000`, no arguments, returns a single **byte** |

`APT 0x0101` is a boolean capability query of the New-3DS class ("is this a New 3DS / an extended memory
layout"). This machine models an **original 3DS** — one application core, the 64 MiB `APPMEMALLOC`
budget its heap sizing assumes — so the honest answer is *false*. Answering true would promise a second
core and a larger budget that nothing here provides.

Note also that Captain Toad talks to **`APT:A`** where SM3DL uses `APT:U`. The service tree already keys
on the family (`serviceBase` strips the suffix), so this needed no change — another small confirmation
that the layer generalises.

With those four, the title runs its C runtime, completes the applet handshake, spawns six threads, and
keeps going without halting.

## Part IV — The frontier: this game needs a real DSP

And then it stops making progress — not with a halt, but with something more interesting. The game's
**sound thread runs its mixer forever in pure userspace**: ten billion instructions without a single
supervisor call. Because Horizon is priority-preemptive and that thread sits at priority 36, above the
main thread's 48, it **starves the game outright**. The main thread sits `ready` at a fixed PC and never
advances.

This is not an emulator bug — it is the correct consequence of an absent DSP. On hardware the DSP is a
separate audio processor that raises an interrupt once per audio frame; the sound thread blocks on the
event it registered (`dsp::DSP` `RegisterInterruptEvents` / `GetSemaphoreEventHandle`), so its mixer runs
at the DSP's cadence. With no DSP, there is nothing for it to wait on, so it never yields. SM3DL never
exposed this because its audio thread is structured differently; Captain Toad's is load-bearing.

**The obvious shortcut does not work, and it is worth recording why.** Two half-measures were tried and
both made things *worse*:

1. *Hand back a real event handle from `GetSemaphoreEventHandle` and pulse it every frame.* The sound
   thread duly blocks and gets a clock — and then wakes up and starts actually **using** the DSP,
   hammering one pipe command 7,504 times against fabricated replies until the game crash-restarts (the
   main thread reappears in its bss-clear loop at the entry point). Giving a thread a clock is useless
   if the thing it then talks to is a fiction.
2. *Just mint the handle* (without pulsing it), on the theory that replying with the result word alone
   is plainly a bug, since it leaves the game to read whatever happens to be in the buffer as its
   handle. It is a bug — but "fixing" it is a **regression**: a real handle that nothing ever signals is
   worse than a bogus one, because the game blocks on it forever, where the bogus handle fails its wait
   immediately and the boot falls through. Verified against SM3DL, where minting the handle costs the
   entire render loop (4,053 GPU command lists → 1).

So the DSP is left honestly unmodelled, and the one thing that *is* answered is the data-register
handshake (`RecvData` / `RecvDataIsReady`), which SM3DL's applet-resume path polls until the component
reports "running".

The conclusion was a clean statement of the next piece of work: **Captain Toad needs a real `dsp::DSP`
model** — the component load, the pipe protocol, the shared-memory audio frame structures, and the
frame interrupt — a subsystem of a scale comparable to the PICA200 GPU, and the natural next platform
phase now that a second title demands it.

## Part V — A real DSP, and the sound system comes alive

The DSP model now exists (`tools/platform/n3ds/dsp.go`), and this title is what drove — and what
validates — it.

**The first decision was LLE versus HLE, settled by dumping what `LoadComponent` actually carries.**
The buffer is 49,756 bytes: a 0x100-byte signature followed by a `DSP1` magic — Nintendo's *signed
Teak firmware container*. The DSP is a CEVA TeakLite-class core; running that blob would mean building
a second CPU emulator. So the model takes the same posture as the Horizon kernel HLE: don't run the
firmware, implement what it *does*. (Clean-room note: the DSP is platform hardware, not game data, so
platform-level documentation and other emulators' implementations were consulted under the same
user-approved exception as the PSP KIRK keys — the protocol and structure layout follow 3dbrew and the
Citra project's DSP HLE, reimplemented in Go; everything about what *this title* does still comes from
tracing its code with our own tools.)

**What the firmware does, and what the model therefore implements:**

- A 512 KiB **DSP RAM window** at `0x1FF00000`. `ConvertProcessAddressFromDspDram` maps a DSP word
  address to `(addr << 1) + 0x1FF40000` — this is the command Captain Toad hammered 7,504 times
  against the old fabricated replies; against real ones it is called exactly 30 times (15 structure
  addresses × 2 regions) and never again.
- Two 0x8000-byte **shared-memory regions** (`0x1FF50000` / `0x1FF70000`) holding the 15 audio-frame
  structures — per-source configurations and statuses, the DSP configuration and status, the mix
  buffers, and a trailing **frame counter** per region. The app and the DSP double-buffer through
  them: the DSP reads whichever region has the higher counter and answers into the other, and the app
  increments the counter of the region it finishes.
- **Pipes.** Pipe 2 (audio) carries the control protocol: the app writes a state change (Initialize /
  Shutdown / Wakeup / Sleep); on Initialize the DSP resets the pipes, answers with a count and the 15
  structure addresses as DSP words, raises the pipe-2 interrupt event, and starts running.
- The **audio-frame clock**: 160 samples ≈ 1,310,720 ARM11 cycles per frame (Citra's
  hardware-verified ratio), paced on the machine's monotonic instruction counter, *independent of the
  VBlank* — the sound thread starts waiting before the game brings up graphics at all, so the run
  loop's idle fast-forward now jumps to whichever machine event fires next (GX completion, DSP frame,
  or VBlank). Each frame the DSP consumes the source configurations (clearing their dirty flags — the
  app's protocol depends on that), advances each source's buffer queue by sample count, publishes the
  per-source statuses (`sync_count` echo, play position, `current_buffer_id` / `last_buffer_id` with
  the report-once dirty byte), raises the pipe-2 interrupt, and signals the **frame semaphore** —
  the event `GetSemaphoreEventHandle` returns, and the thing Captain Toad's sound thread actually
  blocks on.
- What it deliberately does *not* do: decode or mix any PCM. The sources model the **control
  protocol** — positions advance, buffer ids complete in order — not the audio. Fidelity is a later
  phase.

One delta from the Citra reference is deliberate and traced: the frame clock arms at **LoadComponent**
(when the Teak core would boot), not at pipe Initialize — Super Mario 3D Land waits on the semaphore
event without ever writing a pipe command, so a clock gated on Initialize would re-create failed
shortcut #2 for it exactly.

**Result, on this title.** The whole init conversation now runs clean — LoadComponent →
RegisterInterruptEvents(pipe 2) → GetSemaphoreEventHandle → SetSemaphoreMask → Initialize →
ReadPipeIfPossible → 30 address conversions — and then the sound thread enters its real frame loop:
it **blocks on the semaphore, wakes once per audio frame, completes the shared-memory exchange, and
parks again**. The proof is in DSP memory itself: after a boot the two frame counters read
consecutive values in the thousands (7714/7715 after ~800M executed instructions), incremented by the
*game's* sound thread once per frame, forever. No starvation — the main thread runs, finishes its
init, and moves on; no crash-restart; no pipe-command hammering. Both prior failure modes are gone.

**Status.** Container layer: complete, unchanged. CPU: two ARMv6K gaps found and closed. OS: four new
service commands, all traced. DSP: modelled; the sound system runs at full cadence and makes no
further IPC after init (39 requests total, then pure shared-memory exchange — exactly the firmware
contract).

## Part VI — The "mystery poll" was the oracle's own bug: a sleeping thread nobody would wake

The post-DSP frontier looked like a game-side puzzle: the main thread apparently polling a condition
once a second forever, worker threads parked with no jobs, GSP never brought up. Unpicking it found
something better — a **scheduler bug in the oracle**, introduced by the DSP itself.

**What the "poll" really was.** The sleep helper (`0x001211A8`) is one iteration of the game's
*synchronous file read*: submit an async fs request, then `svcSleepThread(10 ms)` and re-check until
it completes (the 10 ms comes from a global at `0x004F81E8`, converted ×1,000,000 to nanoseconds —
read straight out of the sleep wrapper `0x003006D0`). The "poll object" at `0x146EFE14` is the
pool-recycled async request; its dispatcher (`0x0033AF78`) indexes a method table and its read
method builds an `IFile Read` in the caller's TLS. Watching the request's offset field across the
boot enumerated the whole conversation: 6 requests through the worker queue at `arb@0x17BDFD08`
(each: main posts, t4 executes, one 10 ms wait) — and the new `-at` instrument on `n3dsdump` names
them: `/SoundData/ADSRINFO.DAT`, two slices of `SFXSCRIPTFILES.DAT`, `STREAMFILEINFO.DAT`, two of
`BgmRhythmInfo.szs`. Then request #7 is prepared: offset `0x0AF6E3C0`, end `0x0C34D828` — byte-exact
the span of **`/SoundData/sound_data.bcsar`, the 20.8 MB sound archive** — the main thread issues
its 10 ms sleep… and every later thread dump shows it still `sleeping`. One sleep, never completed,
for billions of instructions.

**The actual bug.** In the oracle, sleeping threads were only readied by the idle path's
`advanceIdle`, which ran when no timed machine event fired first. The DSP's audio frame recurs every
1.31 M instructions and *always wakes the sound thread* — so from the moment the DSP existed, the
scheduler always had either a runnable thread or a nearer DSP deadline, `advanceIdle` was starved
forever, and with it every `svcSleepThread` in the game. The healthier the sound system, the deader
every sleeper. (This is the counterpart lesson to the DSP shortcuts: the DSP was necessary, and its
correct heartbeat then exposed the run loop's idle logic as the next lie.) The fix unified the
machine's clocks — the tick now advances at exactly 2 per instruction through idle jumps too — and
made the idle path a single earliest-event selection over {GX completion, DSP frame, sleeper
deadline, VBlank}, with due sleepers also woken inside the scheduling loop, since a machine with a
recurring event may never be idle at all.

**With sleepers alive, the boot sprints.** The bcsar read completes, the whole 31-machine-second
stall collapses to a boot that reaches each next frontier in ~100 M instructions, and the frontiers
fall one by one, each traced from its wrapper before answering: `act:u 0x0001` (NNID account init —
version + ProcessId + a handle, result-only ack), APT `0x003B` (byte setter) and `0x002B` (bare
query, result-only), APT `0x002C` (the app hands over a capture-info buffer), `nfc:u` `0x0001`
(amiibo reader init) plus its bare status queries `0x000B/0x000C/0x000F` (zeros: no adapter
activity, no tag). The game brings up **GSP**, spawns its full engine thread set (13 threads),
submits its first GPU command list and starts swapping frames.

**One frontier needed a policy decision: APT `0x0044` is `GetSharedFont`.** The wrapper reads a font
address and a shared-memory-block *handle* from the reply; zeros sent the game straight into
`svcMapMemoryBlock(0)`. The system shared font is **console firmware data** — it lives in NAND, not
on the cartridge — so per the same policy as encrypted images and AES keys it is not fabricated:
the reply is an explicit failure. **The title tolerates it**: the boot continues past the font.

**Where the boot stood at the end of that session.** An 8-billion-instruction run completed without
a single halt: the game ran its full 15-thread complement, drove GSP (176 requests), kept its
sound system at cadence (88 dsp requests then pure shared memory), and appeared to present ~98
frames (6 GPU command lists, 3 display transfers) — **still black: zero draws submitted**. Not a
deadlock: a memory watch on the global `0x004F35E4` the main thread appeared parked on showed a
hot, constantly cycled lock — the engine's steady-state loop alive and spinning around something
that never became ready to render.

## Part VII — Zero draws: the frame that could never finish presenting

The suspects going in were game-side: the refused shared font's consumer, an intro movie, a
resource-build job for the idle worker pool. All wrong — and the two instruments that proved it
wrong were a one-line trace on the "swap" IPC command and a 500-instruction trace of the main
thread's spin.

**First, the font question closed quickly.** The `GetSharedFont` consumer is a font-manager init at
`0x00225B14`: it builds the game's font table (its own `.bcfnt` files under `/LayoutData/FontData`),
then calls the shared-font fetch (`0x0033DD44`) and polls a three-state getter (`0x0033DD94` — the
mapped font's first word, or **3** if the font pointer is null) in a sleep loop at `0x00225DA4`:
state 2 = ready, use it; state 3 = give up, return; anything else = sleep 10 ms and re-poll. With
our explicit `0xC8A0CFFC` failure the pointer stays null, the getter returns 3 on the first poll,
and the function returns cleanly — logpc probes confirm the loop ran exactly once and the state-2
path never executed. The game genuinely tolerates the missing console font.

**Then the "98 frame swaps" dissolved under one printf.** The gsp HLE had labeled command `0x001F`
"SetBufferSwap" and counted it as frames presented. Logging each call's arguments showed address +
length pairs (`0x14BC2380/0xA0`, … `0x14000010/0x21298`) from ten different call sites — that is
**StoreDataCache**, a cache-maintenance hint. The mislabel had been harmless under SM3DL, but here
it had manufactured the whole "engine presents black frames" narrative. The truth: **this game had
never presented a single frame, and frame presentation never goes through IPC at all.**

**What presentation actually is — and what the oracle wasn't doing.** A `-tracefrom` on the hot
lock's acquire caught main's spin in the act, and it is a *present fence* at `0x00271840`: each
iteration re-reads two per-screen counters (`+0x154/+0x158` of the screen manager at `0x004F3580`)
and a byte via `0x003428C8(screen)` — `[ptr+1]` where ptr comes from the manager's per-screen table
at `+0x5C`: `0x10002200` / `0x10002240`. Those addresses are **GSP shared memory +0x200/+0x240: the
per-screen framebuffer-info structures** of the 3DS buffer-swap protocol. The writer
(`0x00126F28`, caught by a memory watch) presents a frame by filling the entry at index
`1 − current` — `{active_framebuf, fb0_vaddr, fb1_vaddr, stride, format, dispselect, attr}`,
0x1C bytes, two entries after a 4-byte header — then LDREX/STREX-ing the header to
`{byte0 = new index, byte1 = 1}`. On hardware the **GSP module consumes the flagged entry at the
next VBlank: it points the LCD at the new framebuffer and clears byte 1**, and that clear is
exactly what the fence polls for. The oracle's VBlank pushed interrupts and signalled events — but
never consumed framebuffer info. So the game's *very first* present hung forever: frame 1 never
finished presenting, the engine never started frame 2, and every screen stayed black. The six GPU
command lists it did submit were pure state setup (list 0 alone is a 6,590-register pipeline
init) — the game initialises the GPU, presents once, and waits.

The fix is the GSP module's missing half: `consumeFBInfo` in `gsp_vblank.go` — each VBlank, for
each screen with the new-data byte set, read entry `[header byte0]`, record it as the presented
framebuffer (new machine state `screenFB`, on the savestate), clear the flag. "Frames swapped" now
counts consumed framebuffer-info entries — the real thing — and gsp `0x001F` is relabeled the
cache hint it always was.

**With presents completing, the engine rendered its first draws immediately** — 913 of them, 3,400
GPU command lists — and the boot ran into three further walls, each of a different character:

*Two ordinary service gaps.* The streaming layer halted loudly on `IFile` commands SM3DL never
sends: `0x080C` **OpenLinkFile** (a second session onto the same open file — its wrapper
`0x0033C170` sends a bare header and reads a handle from the reply's translate slot; the streamer
clones its RomFS session per in-flight read) and `0x080A/0x080B` **Set/GetPriority**. Both
implemented from the wrappers, both routine.

*An oracle bug: completions gated on submissions.* At ~590 presented frames the engine froze —
every thread waiting, presentation stopped. The game's graphics driver runs a dedicated
command-list thread (t10, stack-tagged `CmdL`) that it **pauses** (a flag polled through the frame
pacer at `0x00303704`) whenever outstanding GPU submissions haven't retired; retirement is counted
by the DMA-completion interrupt callback (registered, notably, on interrupt id 6 — *DMA*, not
VBlank). The freeze: six GX commands sat accepted-but-never-completed in the oracle, because
`processGXQueue` returned early when the game had posted nothing new — and the only other pump
site, the idle path, never ran because the sound thread's cadence kept some thread always
runnable. Main held the driver paused waiting for those six completions; paused, the driver could
never post again. Completions are paced by the machine clock alone now — `pumpGX` runs
unconditionally at the queue check.

*A protocol fault in the DSP model: two heartbeats per frame.* The next freeze was nastier: the
**sound thread spinning forever inside a linked-list append** — the list had acquired a cycle
(a node whose `next` pointed at itself), and at priority 36 the spin starved every thread below
it, main included. The corruption was reconstructed watch by watch: the sound system queues
per-voice command nodes (fixed slots, `0x18` apart) onto a global list at `0x0054A020`, drained
per audio frame by a routine that **requires the DSP's per-source status to echo an exact
sync-count** before it will send. Our `dspTick` raised *both* the pipe-2 interrupt *and* the frame
semaphore every audio frame; the game's sound thread treats each signal as a frame, ran its frame
processing at double rate (the shared-region frame counters read ~2,200 after ~1,000 real DSP
frames), bumped its sync-counts twice per DSP consumption — and the exact-match echo skipped past
the expected value and never matched again. The drain wedged; five music-track starts later the
voice allocator recycled a still-queued node, the append walked into `X.next = X`, and the game
was gone. On hardware the per-frame heartbeat is the **semaphore alone** — pipe interrupts
accompany actual pipe messages (the `Initialize` reply). One deleted signal call fixed it.

**With those four fixed, the engine runs at full cadence.** A 2-billion-step run presents 63,020
frames — one per VBlank, 60 fps for the entire run — with 157,676 GPU command lists, 78,832
display transfers and 188,447 draws; the sound thread parks and wakes healthily on its event; no
halts, no stalls. And yet every screen was still black, because **every single triangle the game
submitted was being rejected**.

## Part VIII — Black for one reason: a shader constant cannot hold a NaN

The rejection was total and it was in the transform: the clip-space position of every vertex came
out `NaN`. Tracing one draw's uniforms found exactly one poisoned constant — **`c4`** — and
tracing the uploads found the game writing the quiet-NaN word `0x7FC00000` into it deliberately,
every frame.

`c4` is the **stereoscopic-3D parameter block**, and the vertex shader (entry `0x06D`) guards its
use. Position is `dp4(c0..c3, r3)`, and `r3` is built by a subroutine that would apply a
stereo eye-shift:

```
065: mov  r11, c4.wzyx            ; the stereo parameters
066: add  r11.z, c34-.xxxx, r11.zzzz
067: cmp  c5.xxxx ne|lt r11.xzzz  ; c5.x is the constant 0
068: jmpc dst=0x6C                ; …skip the whole block
069: rcp  r11.z, r11.zzzz         ; (the shift: r3.x += iod / (z − focus))
06A: add  r3.x, r3.xxxx, r11.xxxx
06B: mad  r3.x, r11-.yyyy, r11.zzzz, r3.xxxx
```

The guard reads *"skip the eye-shift unless this parameter differs from zero"* — the game's way of
saying "3D slider off, don't shift". Writing a NaN into it is the game telling the hardware
nothing at all: **a PICA200 shader constant is a float24** — one sign bit, seven exponent bits,
sixteen mantissa bits — and that exponent field is too narrow to hold the f32's all-ones exponent.
There is no NaN in float24. A uniform uploaded in the "f32 mode" of register `0x2C0` is converted
on the way into the register file, the poison pattern does not survive the trip, the parameter
reads as an ordinary finite number, the comparison behaves, and the block is skipped.

Our model stored the constant as a true IEEE `float32`. IEEE says a NaN differs from *everything*,
so the `ne` comparison came out true where hardware's is false, the guard did not skip, the stereo
maths ran on a NaN, `r3.x` became NaN — and from there every vertex, every triangle, every frame.
The fix is `toF24` in `gpu_float.go`: quantise an f32-mode uniform upload to what the register file
can actually hold (rebias the exponent 127 → 63, keep sixteen mantissa bits; NaN, infinity and
float24 underflow all fall to zero, because none of them have an encoding).

The effect is not subtle. On one draw, before and after:

| | before | after |
|---|---|---|
| triangles rejected at w ≤ 0 | 4,610 | **0** |
| pixels drawn | 0 | **89,768,446** |

This is the fifth core-assumption bug of the port, and it is the same shape as the others: not a
missing feature, but a place where the model was *more* precise than the hardware and the game
depended on the imprecision. It is also a clean example of why a second title matters — Super
Mario 3D Land poisons `c0`–`c3` too, but only on warm-up frames, and always uploads real matrices
before a real draw; it never *draws* through a poisoned uniform, so it never noticed that our
uniforms could hold a value the hardware cannot.

## Part IX — The DSP grows a voice: sample formats, resampling, a mixer, and actual sound

The DSP model built in Part V spoke the *protocol* — pipes, the shared-memory frame exchange, the
dirty-flag handshake — but it decoded nothing. A voice's play position advanced by a sample count
against a buffer length, and no PCM was ever read. That was honest for bringing the sound thread to
life, and useless for anything else: a model that never touches the audio cannot tell you whether
the game's sequencer, its streaming or its effects are doing the right thing. So it now runs the
firmware's real per-frame pipeline.

**The format is in the configuration, and everything follows from it.** `flags1` at `+0xB4` carries
the channel count (bits 0–1) and the codec (bits 2–3) — and the buffer's *length is in samples*, so
until the codec is known, the byte arithmetic cannot be done at all. Three codecs, and Captain Toad
uses the third for everything: PCM8 (one byte per sample per channel), PCM16 (two), and **GC-ADPCM**
— eight-byte blocks, each a header byte and fourteen four-bit samples. The header's low nibble is a
scale (a power of two), its high nibble an index into the sixteen coefficients that arrive in the
ADPCM-coefficient structure — two taps of a second-order predictor, `y[n] = x[n]·scale + c1·y[n-1] +
c2·y[n-2]`, computed in 11-bit fixed point. Its state (`y[n-1]`, `y[n-2]`) carries across blocks
*and* across buffers, and a buffer may re-prime it from the ADPCM fields the app queued with it.

**Resampling, and one detail that is load-bearing.** Each frame a voice produces exactly 160 output
samples, stepping over its decoded buffer at its rate multiplier with a fixed-point position (24
fractional bits) and interpolating between input samples. The position, *and the two input samples
around it*, carry across the frame boundary — and also across the **buffer** boundary: consecutive
buffers of a stream are one continuous signal, and an interpolator that restarted at each boundary
would click at every buffer. Then the voice's two optional filters run, a single-pole and a biquad,
both normalised with the feedback coefficients pre-negated by the application.

**Then the mixers.** Each voice adds its frame into three quadraphonic intermediate mixes at twelve
gains (four channels × three mixes). Two of those mixes can be routed through an **aux bus**, which
is a round trip through the *application*: the DSP writes the mix into the shared region and takes
back whatever the app's own effects code left there — the ARM11 is the effects processor, one frame
late. Finally the three mixes are downmixed to stereo at their volumes (master, aux-return 0, aux-
return 1) and summed into the 160-sample frame the region publishes. The firmware's limiter and
compressor are *not* modelled; the mix is what a disabled limiter produces, and the code says so
rather than pretending.

**It makes sound.** `bootoracle -wav` captures the final mix to a playable file — Super Mario 3D
Land's dialog scene yields 4.9 seconds (peak 5060), Captain Toad's opening stage 6.8 seconds (peak
15,137). That mixer is the point: it is the verification oracle any future reimplementation of the
game's sequencer or effects gets checked against. And the regression gate held — SM3DL's welcome
dialog is still pixel-identical.

One deviation from the protocol survives, and it is worth naming because it is *measured*, not
assumed: `sync_count` is read out of the configuration **every frame**, not on its dirty bit. Gating
it correctly was tried, and the game got *worse* — it stopped queueing buffers on its streaming
voices altogether. The app writes each region with only the fields it has changed, so a region's
`sync_count` word can belong to an older generation than its frame counter suggests; something about
how the game reconciles that is still not understood, and the honest state of it is a flag in the
code, not a silent choice.

## Part X — One spinning thread, and the stream channel that was never started

With a real voice model the boot still stops in the same place: 588 presented frames, then the sound
thread (t6, priority 36) spins forever and starves every thread below it. The model didn't fix it —
but it made the whole conversation visible (`-dsptrace` logs every configuration the DSP consumes
and every status it publishes; none of it goes through IPC, so nothing else could see it), and the
chase finally reached the bottom.

**The spin, exactly.** `0x00132D2C` appends a command node to a voice: it null-terminates the node's
`next` (`+0x14`), takes the voice's lock, and walks the list from the head (`voice+0x30`) to the tail
to link it. Which means appending a node *that is already in the list* makes the tail's `next` point
at the node itself. The live node at `0x140D96B0` has `+0x14` = `0x140D96B0`. Whoever walks the list
next — the sound thread, every frame — never reaches the end.

**Who appended it twice.** Logging every append across the boot: the node is appended to voice
`0x0054A020` at instruction 10,989,842 and again at 11,997,688, with no retire in between. Nodes are
pooled and re-used all the time — that is normal — but this one was never released, because the code
that would release it never ran.

**Why it never ran, in one chain.** Voice `0x0054A020` is DSP source 5. The per-frame flush that
turns a voice's queued command nodes into a DSP buffer queue (`0x00132444`) is only called for a
voice whose state byte `+0xD` is zero (`0x00132118`); the state is 2, meaning *allocated but not
started*. The start (`setVoiceState(voice, 0)` at `0x001312B0`) is issued by the voice's **player
object**, and only when the player's byte `+0x15` is non-zero. The two players are twins — the same
stream's two channels, allocated 475 instructions apart, identical in every field but one:

| | player `0x080C6568` (left) | player `0x080C6748` (right) |
|---|---|---|
| voice | `0x00549FA4` → DSP source 4 | `0x0054A020` → DSP source 5 |
| `+0x15` (block ready) | **1** | **0** |
| started | yes (instr 11,039,281) | **never** |
| buffers flushed to the DSP | 4 | **0** |

So: the right channel's player never receives its first decoded block, so it never starts its voice,
so the voice is never flushed, so its five command nodes sit in the list forever, so the streamer —
topping up a queue that never drains — appends node #1 a second time, and the list eats itself. One
un-started stream channel is the entire black screen.

That reframes the frontier, and it is a better place to stand than "the sound thread spins". The DSP
side of this voice is *not* the problem: source 4, its twin, streams correctly through the same model
— configuration consumed, buffers queued, statuses reported, buffers retired. What is missing is one
step in the game's own stream pipeline: the block that should be handed to the second channel's
player (a structure copy that sets `+0x15`, seen twice for the left channel and never for the right).
The next question is who produces that block, and what it is waiting for.

## Part XI — The sixth core-assumption bug: an unaligned word store

The producer of the missing block was worth chasing one hop further, and the hop ended somewhere that
had nothing to do with audio.

The stream's channel table (`0x140DB120`) holds two channel entries — and both pointed at the *same*
object. The two entries are written from a **track descriptor**, whose two channel-index bytes sit at
`+0xB` and `+0xC`; both read 0, where the second should read 1. The descriptor is filled by a copy
loop at `0x0012D5BC` that walks the game's track table with a **13-byte stride** — three words and a
byte per track — so every track after the first is written through an *unaligned* pointer. Watching
the byte itself settled it in two lines:

```
watch 0x140D9180: 0x00000000 -> 0x00000001 at pc=0x0012D5F4   # track 0's second channel index: 1 ✓
watch 0x140D9180: 0x00000001 -> 0x00000000 at pc=0x0012D5DC   # track 1's first word store: ate it
```

Track 1's word store lands at `0x140D9181`, one byte past track 0's last byte. It should write the
four bytes *at* `0x140D9181`. Our core wrote the word at `0x140D9180` — the address rounded down.

`read32` was split by architecture when the message-archive bug was found: ARMv6 (the ARM11, with
unaligned access enabled by Horizon) does a **true unaligned load**; ARMv5 (the DS) forces the address
aligned and rotates. **`write32` never got the same treatment** — it masked the address on every
architecture. On ARMv6 that is not a rounding error, it is silent corruption of the three bytes *in
front of* the target.

So the whole chain, end to end, from one masked address:

> unaligned `STR` writes low → track 0's second channel index is overwritten with 0 → the stream
> resolves both channels to the same object → the right channel's player never receives a decoded
> block → it never starts its voice → that voice is never flushed to the DSP → its command nodes
> accumulate → the streamer re-appends a node that is still linked → `X.next = X` → the sound thread
> spins forever on the list walk → every thread below priority 36 starves → black screen.

With the store fixed, the sound thread **parks on its DSP interrupt event** instead of spinning, and
**DSP source 5 streams in lockstep with source 4** — the stereo stream has both of its channels for
the first time. `LDM`/`STM` and `SWP` keep the aligned path (they ignore the low address bits on every
architecture). The DS regressions pass and Super Mario 3D Land's welcome dialog is pixel-identical.

This is the sixth core-assumption bug of the port, and — like the unaligned load, the float24
constant, and the per-thread instruction counter — it was found by a *second title* exercising a path
the first one never took. Super Mario 3D Land never stores through an unaligned pointer; Captain Toad
does it once per track in a table it parses at every stream start.

**Where that leaves the boot.** A 1.6-billion-instruction run from the pre-stage snapshot now finishes
with no spinning thread and no deadlock: all seventeen threads are parked on real objects (the sound
thread on its DSP event, the render thread on the VBlank condvar, the workers on their queues), 24,731
VBlanks are delivered, 107,305 GPU command lists are submitted, and the DSP mixes **65 seconds of
continuous music** (peak 18,514). The engine is *alive* in a way it has never been.

It is also still black — and now for a different reason than before. The command lists keep coming but
carry **no draw calls**: the scene has no actors in it. That is the same loose thread the RomFS trace
found earlier (the game reaches Season 1's opening stage, parses the map, and then loads four of its
2,364 `/ObjectData` files), and with the sound system finally healthy — it is also the engine's
resource-loader producer — it is the next thing to chase.

## Part XII — Listening to the mix finds the last lie

The captured WAV is not a souvenir, it is an instrument, and playing it back caught something no
counter had: the music plays for about two seconds, jumps back a second, and repeats — 1,2,3 then
2,3,4 then 3,4,5. Tempo perfect, sound effects in place, and the whole thing sliding backwards.

Hashing the mix frame by frame ruled out the easy explanation immediately: **no audio frame is ever
emitted twice** (only silence repeats). Nothing in the DSP was re-publishing frames — a *voice* was
replaying its buffers. The new `-dsptrace` DEQUEUE log said so in one glance:

```
1 2 3 4 5 6   3 4 5 6 7   4 5 6 7 8   5 6 7 ...
```

The stream plays six buffers, then goes back to buffer 3. Buffers play lowest-id-first, so this means
buffers were being **enqueued twice** — and the ENQUEUE log showed the application doing it: at frame
1582 it marks its queue slots dirty again with ids **4, 5, 6**, all of which it had already handed over
five hundred frames earlier, and only then appends the new id 7.

**Why would the app re-declare buffers it has already played?** Because it did not know it had played
them. Its per-voice update (`0x00131FAC`) compares the DSP's `sync_count` echo against its own count on
the *first instruction* and bails if they differ — and our echo was stale, because `sync_count` was the
one field the model deliberately read **every frame** rather than on its dirty bit. The application
writes each shared-memory region with only the fields it has changed, so that word can belong to an
older generation than the region's frame counter suggests. Measured, on the streaming voice:

| | stale echo (read every frame) | dirty-gated echo (the protocol) |
|---|---|---|
| update calls that pass the sync gate | 363 / 3,280 | **2,596 / 3,280** |
| buffer completions the app actually sees | 6 | **25** (all of them) |
| stream dequeue order | 1,2,3,4,5,6, **3**,4,5,6,7… | **1,2,3,4,5,6,7,8,9,10…** |

Missing the completions, the app's command nodes never retired; and when the voice then ran dry, the
DSP reported `current_buffer_id = 0` — which is the app's signal to **zero its in-flight count**
(`0x001320A0`) and re-push every node still in its list, including the ones already played. Hence the
slide backwards.

**And the deviation was never real.** Part IX recorded, honestly, that gating `sync_count` on its dirty
bit had been *measured worse* — the game stopped queueing buffers on its streaming voices. That
measurement was taken against a game whose memory was being corrupted by the unaligned-store bug of
Part XI. With the store fixed, the protocol-correct latch is not merely safe, it is the fix. The
workaround was the bug wearing a disguise, and the only thing that unmasked it was listening to the
output.

## Part XIII — 107,000 command lists, no draws: the game submits only the head of its display list

The engine was alive — no spinning thread, no deadlock, every thread parked on a real object, 24,729
VBlanks, sixty-five seconds of correct music — and both screens were black. 107,271 GPU command lists
had been submitted and not one pixel had been drawn. That is a strange shape of failure, and it has a
cheap first fork: *look at what is actually in the lists.* `-gxdump` captures every submitted PICA
command buffer at submission time; `picadump` decodes one.

**The answer was in the first pass.** Of 795 captured lists, **zero** contained a draw register
(`0x22E`/`0x22F`). So we were not dropping draws — the game was not sending any. But 383 of those 795
ended the same way:

```
reg 0x238 = 000000EA      CMDBUF_SIZE0   (bytes >> 3)
reg 0x239 = 00000004      CMDBUF_SIZE1
reg 0x23A = 029784FC      CMDBUF_ADDR0   (address >> 3)
reg 0x23B = 029255C6      CMDBUF_ADDR1
reg 0x23C = 00000001      CMDBUF_JUMP0   ← trigger
```

Those are the PICA command processor's **own chain registers**: two preloaded buffer slots and a
trigger apiece. Writing a JUMP makes the processor continue in that buffer — and it is a *jump*, not
a call: the rest of the current buffer is never executed. We had never implemented them, so the GPU
ran the buffer the game submitted and stopped at its end.

The game's frames are not lists, they are **chains**. Following the jumps showed the shape: a *spine*
of 32-byte buffers in FCRAM, each one preloading a content buffer and jumping into it; each content
buffer ends by jumping back to the next spine entry. One frame is **1,633 hops** deep. Here is a
content buffer, 96 bytes, entire — a complete draw and the jump out:

```
7fff0001 000f02b0    vertex-shader bools
00000001 000f025f    restart primitive
00583b84 000f0227    index buffer
00000198 000f0228    408 vertices
00000001 000f022f    ← DrawElements
00000001 000f0231    clear post-vertex cache
00000001 000f023d    ← CMDBUF_JUMP1: onward
```

`Execute` now follows the chain (bounded, so a corrupt jump register halts loudly instead of hanging).
The draws appeared immediately. **Super Mario 3D Land had been missing most of its buffers too** —
45,000 draws became 3.4 million, and its welcome dialog came out pixel-identical, which is exactly
what you want from a fix like this: much more work, byte-for-byte the same picture.

## Part XIV — Two more, found by the draws that now existed

**A megabyte of depth over the game's own command buffers.** With the chain followed, the boot began
halting inside VRAM: a chained buffer that had executed cleanly a hundred times decoded as garbage on
a later pass. Its contents had changed under us. The culprit was our own rasteriser.

Captain Toad's shadow pass renders a 512×512 map, and it *keeps the depth-test bits set in `0x107`*
while never re-pointing the depth buffer — so it inherited the main pass's depth address and we wrote
512×512×4 = one megabyte of depth from `0x1F300000`, straight over the main pass's depth buffer, over
the command buffers the game keeps in VRAM, and over the shadow pass's own colour target. The game's
own memory map proves it can't have meant that: main depth *ends* at `0x1F380000`, and the command
buffer sits at `0x1F380720`, in the gap.

What the shadow pass does instead is switch the depth buffer off one level lower down:

```
reg 0x112 = 0000000F   colour buffer read   enabled
reg 0x113 = 0000000F   colour buffer write  enabled
reg 0x114 = 00000000   depth buffer read    DISABLED
reg 0x115 = 00000000   depth buffer write   DISABLED
```

These are the **buffer access enables**, a second gate in front of `0x107` that permits the memory
traffic itself. Zero means the pass does not touch that buffer at all. Honouring them: 3,052,450
bogus depth kills → **0**, and pixels drawn 482,438 → **103 million**. (Super Mario 3D Land writes
zero to both, always — it never enables the depth test either. It is a game that never used a depth
buffer, which is why the gap survived a whole title.)

**And the scene is coloured by a unit we had stubbed.** With the draws landing and the depth correct,
the colour buffer filled with `FD 00 00 00` per pixel: alpha, and **no colour at all**. Captain Toad's
world geometry carries *black vertex colours* — every pixel of it is coloured by the PICA's
**fragment-lighting unit**, which we had been halting on. It is now implemented (`gpu_light.go`), and
the interesting part is the normal: the PICA does not carry one. The vertex shader emits a
**tangent-space quaternion** (output semantics `0x04`-`0x07`, next to the view vector at `0x12`-`0x14`)
and the fragment normal is that quaternion applied to `(0,0,1)`. The game's own output map says so:

```
o0 = 03020100   position x,y,z,w
o1 = 07060504   quaternion          ← the normal
o2 = 1F141312   view vector
o3 = 0B0A0908   colour
```

Four lights, warm diffuse, a global ambient. Ambient and diffuse are modelled; the specular lookup
tables are named and left out, and the code says which — a missing highlight, not a missing image.

## Part XV — The frontier: the game poisons its view matrix, and then draws with it

The scene now rasterises: 11,200 draws a frame into a 256×512 render target, correctly depth-sorted.
It is still black, and the reason is now located to a single uniform.

Every one of the scene's draws runs its vertex shader from entry `0x000`, which is a dispatcher — five
`call`s and an `end`. Every path through it finishes the same way:

```
0A7: dp4  r15.x, c90, r10      ← r15 = the view matrix × the model-space position
0A8: dp4  r15.y, c91, r10
0A9: dp4  r15.z, c92, r10
0AE: dp4  o0.x, c86, r15       ← then the projection
0B1: dp4  o0.w, c89, r15
```

The projection (`c86`–`c89`) arrives correct and recognisable: `c86 = [0, 3.732, 0, 0]` and
`c89 = [0, 0, -1, 0]` — a 30° perspective rotated for the 3DS's portrait framebuffer, with the y-scale
3.732 in the x-row because the screen is turned on its side. But `c90`–`c92`, the **view matrix**, is
all zeros in our register file, so `r15` comes out `[0,0,0,1]`, `o0` collapses to the projection's
translation column — exactly the `[0, -0, 100.087, 0]` every vertex reports — `w` is zero, and every
triangle in the scene is rejected.

It is zero because **the game deliberately writes NaN into it**, in the same command buffer that
uploads the projection:

```
8000005a 000f02c0    uniform upload: index 0x5A = c90, f32 mode
ffc00000 00bf02c1    12 words follow
7fc00000 7fc00000 7fc00000 ffc00000
7fc00000 7fc00000 7fc00000 ffc00000     ← c90, c91, c92: all quiet NaN
```

And logging the uploads against the draws shows the order is not an accident:

```
identity × 5
REAL view matrix (43AC59BE …)
POISON (FFC00000 …)          ← immediately after the real one
88 lit draws                 ← every one of them drawn with the poison live
```

The game uploads its real view matrix, poisons it, and *then* draws eighty-eight objects — and on
hardware those objects appear. So the poison is not poison to the hardware, and our `toF24` (Part
VIII: a float24 has no NaN, so NaN → 0) cannot be the whole rule for the uniform path. Part VIII was
*measured* — it took one draw's rejected-triangle count from 4,610 to zero — so the answer is not to
undo it but to find what distinguishes the two cases. That is the next question, and it is a small,
sharp one: **what does a PICA200 uniform upload actually do with an all-ones exponent, such that
`c4`'s stereo sentinel reads as "disabled" and `c90`'s view matrix still transforms?**

Everything downstream of it is now standing and waiting: the chain executes, the depth is right, the
lighting unit is built, the quaternion normals are wired, and the moment `r15` is a real position the
frame has somewhere to go.

## Part XVI — The NaN was ours: an unfilled config block, a 0/0, and a camera full of nothing

The question Part XV left was the wrong question. It assumed the poison was the game's, and asked what
hardware does with it. The game never wrote it.

The tell was in the arithmetic, and it was decisive before a single line of the GPU was touched. Look
again at what reaches the projection register `c87`, and at what the game's *own* projection matrix
holds in memory at the same moment:

```
the game's struct @0x1504ECE4        what reaches the GPU
row0  [0, 3.732, 0, 0]               [0, 3.732, 0, 0]        identical
row1  [-2.2392, -0, -0, -0]          [NaN, -0, NaN, -0]      ← two NaNs
row2  [-0, -0, 1.00087, 100.087]     [0, 0, 1.00087, 100.087]
row3  [0, 0, -1, 0]                  [0, 0, -1, 0]           identical
```

The game has the right number. `-2.2392` is exactly `-3.732 × (240/400)` — the y-scale divided by the
screen's aspect, which is what row 1 of a portrait-rotated perspective must be. And now the trap in
Part XV's framing springs shut: row 1 needs `-2.2392` in its `x` and `0` in its `z`, and the buffer
holds **the same NaN word in both**. No conversion function — flush-to-zero, exponent wrap, saturate,
preserve — can turn one bit pattern into two different numbers. There was never a float24 rule that
could explain this. The uniform path was never the bug.

So where does a NaN come from, if not from a constant? It comes from arithmetic. A NaN is *born*, in
an invalid operation, and there are only so many of those. So the ARM core got an instrument: report
any VFP operation whose result is NaN when neither of its inputs was. Across two million instructions
of a running frame, in a machine drawing eleven thousand times a frame, it fired on exactly **one**
instruction:

```
$004015D4: LDR   r1, =0x00546348      ← a global
$004015D8: VLDR  s0, [r0, #+0x48]     ← the camera's depth
$004015E4: VLDR  s1, [r1, #+0x8]      ← a float from that global — reads ZERO
$004015E8: VDIV.F32 s2, s0, s1        ← 0 / 0  =  NaN
$004015EC: LDRB  r1, [0x1FF81084]     ← the 3D LED state
$004015F8: VMUL.F32 s1, s2, s3        ← × 0.5
$004015FC: VLDREQ s0, [0x1FF81080]    ← the 3D depth slider = 0.0
$00401600: VMUL.F32 s0, s1, s0        ← NaN × 0 = NaN, not 0
```

This is the stereoscopic eye separation. The console's 3D slider is all the way down — we report it as
`0.0`, correctly, because this machine has no 3D — and the routine dutifully multiplies the separation
by zero. On hardware that yields `0.0`: no stereo, mono camera, done. Here it yields NaN, because IEEE
754 says `NaN × 0` is NaN, and the NaN was already there. **The multiply that was supposed to make
stereo a no-op cannot switch off a NaN.** From there it flows into the camera, and the engine bakes the
camera once, at scene setup, into static command buffers (a `VLDM`/`VSTR` serializer at `0x003091CC`
that loads twelve floats and stores them out reversed, w-first). Baked once, never rewritten — so the
whole opening stage was frozen around a view matrix of nothing, and every triangle in it was
w-rejected.

That leaves the divisor. `0x00546348` is a 32-byte singleton in `.bss`, lazily constructed, and the
constructor is an IPC wrapper: header `0x00010082`, a 32-byte write buffer, block id `0x00050005` —
**`cfg:u GetConfigInfoBlk2`, the console's factory stereo-camera calibration.** Our `cfg` HLE
implemented four config blocks and zero-filled the rest, and this is what an earlier pass wrote beside
the one that mattered:

```go
// 0x00050005 (stereo-camera calibration, 32B) and any others: zero is
// benign for boot.
```

It was not benign. It was the black screen. The block is eight floats of display geometry; the game
reads two of them — `[2]` as that divisor, `[4]` as the camera's default depth. Fill them with the
numbers every 3DS ships with and the division is an ordinary small number, the slider multiplies it to
a clean zero, and the camera is a camera. What the render depends on is not the exact calibration — the
slider annihilates that term — but simply that the numbers are *real*, and non-zero.

Two lessons, and they are the same lesson twice. **A zero is not a safe default; it is a value, and the
game will do arithmetic with it.** Every stub that "returns zeros for now" is a claim that zero is in
the domain of whatever consumes it, and here that claim was false in the one way floating point
punishes hardest. And: when a measurement (Part VIII's NaN → 0, which really did take one draw's
rejected triangles from 4,610 to zero) contradicts new evidence, the resolution may be that the
measurement was real and its *explanation* was invented. `toF24` was fitting a rule to a symptom whose
cause was three subsystems away.

### What the fix reached

The block is eight floats; fill them and the camera is a camera. A cold boot — and it has to be a cold
boot, because the `cfg` read happens early and the poison is *baked* into static command buffers, so
every savestate we had was already contaminated — now runs the full budget with no halt, and the
counters invert:

| | before | after |
|---|---|---|
| pixels drawn | 0 | **13,346,300,283** |
| draws | (all rejected) | 789,590 |
| back-faces culled | 0 | 27,556,096 |
| depth-killed fragments | 0 | 13,619,400,772 |
| w-rejected triangles | *every one* | 1,935,795 (a near-plane trickle) |

And the render target has *something* in it.

**Which is where I claimed victory, and was wrong.** The frame is not a stage. It is a thicket of long
thin spiky triangles fanning in every direction — structure, certainly, but not the structure of a
Captain Toad diorama. I read "no longer black" as "correct", and quoted the counters as if they
corroborated it. They do not: thirteen billion depth-killed fragments and twenty-seven million culled
back-faces are exactly what garbage geometry also produces. What the numbers above actually establish
is narrower, and worth stating precisely: **the camera is no longer NaN, and the triangles are no
longer all w-rejected.** Nothing in them says the triangles are the right triangles.

So the fix stands — it is arithmetic, and the `0/0` is gone — but it bought a different frontier than
the one I announced.

Closing the NaN immediately exposed three things the scene had never been able to reach, each fixed
here: texture format **0xD (ETC1A4)** — ETC1 colour blocks each preceded by a 4-bit alpha plane, the
nibbles in ETC1's own column-major pixel order; texture format **0x6 (HILO8)**, two 8-bit channels for
a bump map (with these two, all fourteen PICA formats `0x0`–`0xD` now decode); and a real bounds bug in
`sampleTexture`, where border-wrap's "outside" sentinel (`-1`) was **y-flipped before the negative
check** — `h-1-(-1) = h` slipped past the guard and indexed a row past the image.

### The next question: the vertices, or the transform?

Spiky triangles radiating from points are a signature, and it is not the signature of a bad matrix. A
wrong matrix mangles a scene *coherently* — everything sheared the same way, scaled the same way,
recognisably a world seen through a broken lens. What this looks like instead is a **malformed vertex
stream**: a wrong stride, a wrong attribute format, a wrong index width, where some vertices decode to
sane model-space positions and their neighbours decode to junk, and the triangles stretched between
them go wherever the junk points. That is a hypothesis, and the next session's first job is to falsify
it, not to build on it — which is cheap, because the two halves separate cleanly. Decode one draw's
raw attributes by hand and ask whether the model-space positions are a plausible cluster. Poke the view
matrix to identity and the projection to an orthographic and render again: if it is still a thicket,
the vertices are wrong and the camera is innocent.

There is one candidate that would explain the picture and links both halves. The scene's vertex shader
does not simply multiply by a model matrix — it *indexes a matrix palette with the address register*
(`code[007]` is a `MOVA`, and `code[008]` a `DP4` against `c25[a0.x]`). If `a0` is loaded from a vertex
attribute the fetch decodes wrongly, every object picks a **random matrix out of the palette**, and
the result would look precisely like this.

The display path is broken too — the top screen shows only part of the target, because the panel's
image is the 240×400 sub-window at byte offset `0x1C000` of the 256×512 buffer and our
`DisplayTransfer` detiles it with the wrong stride — but that is queued behind this, and deliberately
so. A correctly-strided window onto garbage is still garbage.

And the lesson, which is the same one this document keeps having to relearn in a new costume: *render
the artefact and look at it* only works if the looking ends in a judgement. "Not black" is not a
judgement. The question a rendered frame has to survive is **"does this look like the game?"** — asked
out loud, of the picture, before a single counter is quoted in its defence.

## Part XVII — It was the vertices: a permutation read backwards, and a framebuffer anchored to the wrong edge

Two bugs, found in that order, and the opening stage renders.

### The permutation ran backwards

The hypothesis inherited from Part XVI — a malformed vertex stream — was half right, and the wrong
half of it was the half we would have spent the session on. The **fetch** was innocent, and it said so
immediately. A scene draw's attribute-format word is `0x000DB7BB`, which decodes to five attributes —
three floats, three floats, two floats, three floats, four unsigned bytes — totalling `12+12+8+12+4 =
48` bytes, and the loader buffer's configured stride is **exactly 48**. A wrong stride cannot produce
that agreement. The UVs came out inside `[0,1]` and the colours at 255. Whatever was broken, it was
not the reading of the bytes.

It was the step immediately after. Registers `0x2BB`/`0x2BC` map the fetched attributes onto the
vertex shader's input registers `v0..v15`, and the map runs **attribute → register**: nibble *a* names
the register that attribute *a* is delivered to. We had it as its inverse — register *j* takes
attribute `perm[j]`.

The game's own configuration settles the direction without needing a hardware reference, because the
inverse of a permutation is not a permutation *of the same set* unless the map happens to be an
involution. One draw fetches attributes 0–2 with `perm = 0x340`. Read backwards, that hands `v1` and
`v2` attributes **4 and 3 — which this draw does not fetch at all** — while the UV and the colour it
*does* fetch reach no register whatsoever, and `v5..v15` all quietly receive the position. That is not
a permutation; it is lossy, and no shipping game configures a three-attribute fetch and then throws
two of them away. Read forwards it is a clean bijection: `v0 ← position`, `v4 ← UV`, `v3 ← colour`.

The evidence had been sitting in the trace the whole time, in plain sight. The shader was emitting
texture coordinates of `[-19204.4, -34630.41]` — *the position values themselves*, because under the
inverted map the register the shader read its UV from had been handed attribute 0.

Fixed, the vertices decode into exactly what a vertex should be: a position, a **unit normal**
(‖-0.62, 0.49, 0.61‖ ≈ 1.0), a UV, a **unit tangent**, and a grey vertex colour. And the scene
rasterises into a Captain Toad diorama — blocky terrain, grass and dirt, a sky. Depth-kills fell from
13.6 billion to 8.5 million and culled triangles from 27.5 million to 1.09 million, but those numbers
are the *consequence* of looking at the picture, not the reason for believing it.

Two claims from earlier parts die here, and both died of the same cause. **The stage's vertex colours
are not black** — they are `(198,198,198)`; Part XIV's "Toad's world geometry has black vertex colours,
so every pixel of it is coloured by the fragment-lighting unit" was reading the colour attribute out of
a register that had been handed something else. The fragment-lighting unit is real and still does the
work; the premise offered for *why* was an artefact of this bug.

Super Mario 3D Land renders **pixel-identically** either way. Its permutation is an involution, so the
inverse is invisible in it. That is how this survived an entire title.

### The framebuffer is anchored to its bottom edge

With the geometry coherent, the diorama was still shoved against the right of the top screen and cut
off at the edge. Not the camera — *where in the render target we put the pixels*.

The PICA anchors the viewport to the render target's **bottom** — an OpenGL-style bottom-left origin —
so a viewport shorter than its buffer renders into the buffer's bottom, not its top. Both games'
`DisplayTransfer` calls say so, and they say it three times without ambiguity; every one reads the
bottom-aligned window:

| screen | viewport | target | transfer source offset | target − viewport |
|---|---|---|---|---|
| Toad, top | 240×400 | 256×512 | `+0x1C000` = 112 rows | 512 − 400 = **112** |
| Toad, bottom | 240×320 | 256×512 | `+0x30000` = 192 rows | 512 − 320 = **192** |
| SM3DL, bottom | 240×320 | 240×400 | `+0x12C00` = 80 rows | 400 − 320 = **80** |

We rasterised top-aligned, so any pass whose viewport was shorter than its buffer landed too high by
exactly that difference and ran off the panel. The fix is one line — the target row is `height-1-y` —
plus dropping the compensating flip the screenshot rotation had been carrying. The two cancel exactly
whenever buffer height equals viewport height, which is the case for Super Mario 3D Land's *top*
screen, and that is the whole reason this went unnoticed.

### The regression gate was guarding a broken frame

Its **bottom** screen had no such excuse, and this is the part worth sitting with.

The pinned regression md5 `d8efe1f5b7baba85385dd2bb84b2b4d7` — quoted across several sessions as proof
that the welcome dialog still rendered correctly, and used to clear every change to the GPU since —
was guarding a **cropped image**. The dialog it locked in reads `SUPER MARIO 3D LAN`. The panel is
shoved 80 pixels to the left with a black band down the edge. It now renders complete and centred:
"Welcome to SUPER MARIO 3D LAND." with the Ⓐ OK button where it belongs. The new reference is
`1a8e9853c14cbd67a4aeec1efc273dbb`.

Nobody was careless. The hash was cut from a frame that *did* look like a 3DS dialog, and from then on
it was quoted rather than looked at. **A pinned hash guards against change, not against being wrong.**
It proved only that we were reproducing the same cropped picture with great consistency. A hash is
worth exactly as much as the one look someone took at the image before pinning it — and that look, at
a dialog whose text ran off the edge, is the one this project keeps having to learn to take.

The display-path bug Part XVI queued behind the geometry — "our `DisplayTransfer` detiles the window
with the wrong stride" — **does not exist**. The stride is taken from the transfer's source width,
which is 256, which is right. What looked like a stride error was the geometry sitting 112 rows above
where the panel reads. Diagnosing a symptom while a known-broken subsystem sits upstream of it produces
a confident description of a bug that isn't there.

---

## Part XVIII — The shading: six bugs in the registers we had never read

Part XVII got the diorama on screen. It was still wrong, and wrong in a way that a screenshot makes
obvious and a counter never will: the stone read **black**, the cast shadows were **pitch**, the star
sat **behind** the ruin it floats in front of, and the decorative ring around the archway — a half
circle of voussoirs with triangular crenellations above it — **was not there at all**.

Six bugs. Every one of them was a field in a register the model had never looked at. None of them was
subtle once read. The pattern of this session is worth stating plainly: **we had implemented the
registers we knew about and left the rest latching quietly into the register file, and every one of
those was load-bearing.**

### 1. A light colour is not a 10-bit fraction

The light colour registers pack three channels into 10-bit fields. We divided by the field width. But
**255, not 1023, is 1.0** — the extra two bits are headroom that lets a light be configured brighter
than full white. Scaling by the field width made *every light in the game, and the global ambient with
them, exactly four times too dark.*

That is most of why the stone was black. Nothing else needed to be wrong.

### 2. TEV sources 0, 1 and 2 are three different values

`gpu_raster.go` overwrote the vertex colour with the lit colour before calling the fragment stage, so
the TEV's PrimaryColor, PrimaryFragmentColor and SecondaryFragmentColor all returned the same thing.
On the PICA they are the **interpolated vertex colour**, the **lighting unit's diffuse output**, and
its **specular output**.

Captain Toad's stone material is, verbatim from its registers:

```
stage0 rgb: fragpri × tex1          ← lit colour × albedo
stage1 rgb: fragsec × tex1 + prev   ← specular × albedo, added
stage2 rgb: prev × vtxcol           ← × the vertex colour
```

It needs all three to be distinct. Part XIV built the lighting unit on the premise that *Toad's world
geometry has black vertex colours* — which is why collapsing them looked harmless. **That premise was
false.** It was an artefact of the vertex-permutation bug Part XVII fixed: the colour attribute was
being read out of a register holding something else. The vertex colours are `(198,198,198)`, and they
are not decoration — they vary per vertex (238 here, 112 there) because the game **bakes ambient
occlusion into them**.

### 3. The lookup tables are consulted, not merely uploaded

Decoding `0x1C3`/`0x1C4` — printed raw by `-gputrace` for sessions and decoded by nobody — says exactly
what the game turns on:

| | |
|---|---|
| `cfg0 = 90810401` | shadow enabled, applied to the primary colour, sampled from **texture unit 0**; **normal map** from **texture unit 2**; light environment 0 |
| `cfg1 = F30EFBFC` | shadow for lights **0,1**; spotlight for light **2**; distance attenuation for lights **2,3**; the **D0** table live, the rest off |

Which is a completely coherent lighting rig: **two shadowed directional key lights** (warm, sun-coloured)
and **two unshadowed positional fill lights** that carry their falloff entirely in distance-attenuation
tables. And the tables the game uploads through the `0x1C5`/`0x1C8` FIFO are *precisely* the ones those
bits name — D0, RR, SP2, DA2, DA3, and nothing else.

The configuration was self-consistent and sitting there unread. (The data FIFO is at `0x1C8`, not
`0x1C6`; we had the address wrong too, so nothing was ever captured.)

### 4. The shadow is a shadow *texture*, and it needs `texcoord0.w`

Nobody had established how the shadow was consumed; stencil was the leading guess. It is neither
stencil nor guesswork — **texture unit 0's type field says `2 = Shadow2D`**, and it says so in a
register we were reading the low four bits of and ignoring the top three.

The whole mechanism:

- the shadow pass sets `0x100`'s **fragment-operation mode** to Shadow, which *replaces the output
  merger*: no alpha test, no depth test, no blender. Each fragment writes a packed texel — 24 bits of
  depth in the low three bytes, an 8-bit **density** in the fourth;
- the scene draws bind that buffer as a Shadow2D texture. The sampler compares the fragment's own
  distance from the light against the stored depth and returns the density, or zero if occluded;
- the **lighting unit** multiplies the shadowed lights' diffuse by it.

The fragment's distance from the light arrives in **`texcoord0.w`** — output semantic `0x10`, which
the pipeline had a comment explaining it did not consume. The shadow projection is orthographic
(`0x08B` bit 0), so the game puts the depth there deliberately.

Captain Toad's shadow pass writes density **0** on every single fragment, so its shadow map is a pure
depth map and its shadows are binary. That is the hardware's design, not a shortcut.

### 5. The depth buffer was W-buffered, and we were Z-buffering

This is the one that hid the geometry, and it is the best example in this port of a **default that is
not neutral**.

`0x06D` bit 0 selects the depth mode. **Clear means W-buffering** — the value the register holds if
nobody ever writes it — and the game leaves it clear:

```
depth = z·scale + w·offset  ==  (z/w·scale + offset)·w
```

We were computing `z/w·scale + offset` and stopping. On its own that is a different depth curve; paired
with the scale the game actually uses — **-1/110000** — it is a catastrophe. Under Z-buffering that
scale crushes the entire scene into the **bottom 151 values of a 16.7-million-value depth buffer**.
Every object in the diorama lands on top of every other one and the winner is decided by rounding.

That is what put the star behind the ruin, the round pillar behind the wall it stands in front of, and
deleted the archway's decorations into the stone. It looked like missing geometry. Rasterising with the
depth test forced to always pass brought all of it back — which is how the model was convicted, before
a line was changed.

### 6. A negative attenuation does not dim a light. It inverts it.

With everything above fixed the picture was right and the shadows were still far too dark — 9% of the
lit value where the hardware gives 63%.

The lighting probe said the two *unshadowed* fill lights were contributing **exactly nothing**, and
then said something impossible: their distance attenuation was **-0.5001**.

A LUT entry is a 12-bit unsigned value plus a 12-bit signed slope to the *next* entry. Both fill lights
sit far outside their own falloff range, so both clamp to the very end of the table — index 255, delta
1.0 — where the model dutifully applied the last entry's slope and **ran off the end of the table**
into whatever the game had left in that field. Captain Toad leaves `-2048` there.

An attenuation of -0.5 does not fail to light a surface. It **subtracts** light from it — and the fill
lights are warm, so what it subtracted was red and green, which is why the shadows were not merely dark
but dark *blue*. Clamping the lookup to what an unsigned 12-bit entry can actually hold took the
shadows from **9% of the lit value to 49%**, against a hardware reference of **63%**.

### What the hardware says, measured

Sampling the real console's frame and ours at the same surface:

| | lit stone | shadowed stone | ratio |
|---|---|---|---|
| hardware | (250, 234, 229) | (157, 147, 152) | **0.63** |
| before | (237, 206, 211) | (21, 25, 58) | 0.09 |
| after | (255, 224, 217) | (125, 118, 133) | **0.49** |

Not yet exact. But it is a *measurement against the machine*, which is the thing this port kept
skipping, and the shape is right: the shadow is cool because what it removes is warm.

### The lesson, again, in a new costume

Part XVI: *a zero is not a safe default.* Part XVII: *a pinned hash is not correctness.* This one:

**An unread register is a decision, and the default is rarely the one the game wanted.** Five of these
six bugs are fields we were latching into `g.Regs[]` and never looking at — the depth mode, the texture
type, the two lighting config words, the LUT FIFO. The sixth is a table we ran off the end of. Nothing
here required cleverness; it required *reading the registers the game writes and asking what each field
means* — and then looking at the picture instead of the counters.

The SM3DL regression md5 is unchanged through all of it. It would be: SM3DL never enables the depth
test and has no fragment lighting. It is a floor, not a gate.

## Part XIX — The title screen was in Japanese, and the language block was innocent

Booting the European cartridge, the splash that flashes before the title is the *Japanese* logo:
進め！キノピオ隊長, `CAPTAIN KINOPIO`. The cart is `(Europe) (En,Fr,De,Es,It,Nl)` and our config
service has been answering `cfg` block `0x000A0002` — system language — with **English** since Part XVI.
So the language was right and the logo was Japanese anyway, which means the logo is not chosen by
the language.

A print in `ipcCFG` says what the game actually asks for. Two commands, not one:

```
CFGTRACE cmd=0001 args=00000001 000A0002 ...   ← GetConfigInfoBlk2(size=1, blk=language)
CFGTRACE cmd=0002 args=00000000 00000000 ...   ← ???  (no arguments at all)
```

Command `0x0002` takes no arguments. We were treating it as a second config-block read
(`GetConfigInfoBlk2 / Blk8`), passing it a block ID scraped from a word the caller never wrote, and
acking with a single zero value word. Its wrapper says what it really is:

```
$00116140: MOV  r0, #0x20000        ; IPC header 0x00020000 — command 2, no params
$00116144: STR  r0, [r4, #0x80]!    ; …into the command buffer
$00116150: SWI  #0x32               ; SendSyncRequest
$0011615C: LDRB r0, [r4, #0x8]      ; cmdbuf[2], low BYTE
$00116160: STRB r0, [r5]            ; → the caller's u8
```

A byte out, nothing in: `SecureInfoGetRegion`. And the zero we were answering with is `JPN`. **The
region, not the language, picks the logo** — so a European console that spoke English booted a
Japanese title screen. Answer it with the same console the language and country blocks already
describe (`EUR` = 2) and the splash is English.

### The blanket ack was hiding a second one

The old handler also had a catch-all: commands `0x0003`–`0x0008`, all acked with one zero word.
Deleting it — the frontier is supposed to halt loudly — made 3D Land halt on `0x0003`, which it had
been calling all along, silently, for ~200M instructions. Its wrapper (`0x001F2A30`) sends one salt
word and reads the reply with **`LDRD r0, [r4, #0x8]`**: a *64-bit* value from `cmdbuf[2..3]`. That is
`GenHashConsoleUnique`, the console's factory `LocalFriendCodeSeed` hashed with the caller's salt —
and our ack filled `cmdbuf[2]` only, so the top half of the hash 3D Land took home was whatever bytes
of its own *request* were still sitting in the command buffer. It is modelled now: one fixed synthetic
console seed, mixed with the salt, deterministic across runs (savestates and the frame debugger's
replays require that much).

### The lesson

**A reply the caller believes is data, whatever we meant it as.** The zero we sent as "success, no
value" was read as a region; the word we never wrote was read as half a hash. Both games booted
through both bugs — the only symptom either ever produced was a logo in the wrong language, and only
because someone looked at the screen. *Answer the command the game sent, not the one whose number you
recognise.*

---

## Part XX — The banner, and the material that was a stencil

The HOME Menu banner — the little 3-D scene shown when the title is highlighted — comes out of the
cartridge by the same chain as 3D Land's, with no new code: ExeFS `banner` → **CBMD** (one shared
CGFX, no language variants, no CWAV) → **LZ11** → a 140,928-byte **CGFX**. What is inside it is the
opposite of the other title's banner, and that is what made it worth extracting.

```
CGFX  revision 0x05000000  fileSize 140928
    Models    1  COMMON (CMDL)
    Textures  3  COMMON1 COMMON2 COMMON3
```

No camera, no light, no scene environment, **no animation**. Three meshes, three materials, four
bones, twelve triangles. The scene is a **folded triptych**: `Logo` is two quads creased down the
middle (the crease sits 2.7 units nearer the viewer than the outer edges), `Bird` is three quads
behind it carrying the tower and its cast, and `TM` is a single quad. The fold is the whole 3-D
effect — on the HOME Menu the parallax as the panels turn is what sells the diorama.

Every mapper is a plain UV0 sampler, scale (1,1), REPEAT. Where 3D Land's banner is a rigged, animated
scene whose atlases are addressed through per-mapper texture matrices, this one is three flat
billboards with untransformed coordinates, and the same exporter covers both — which is the useful
result: `tools/platform/n3ds/cgfxglb` is now a *format* exporter, not one game's scaffolding.

### ETC1A4 — the alpha plane the CGFX decoder had never been asked for

`COMMON1` (the logo) and `COMMON2` (the tower scene) are 256×256 in GL format **`0x675B`**: ETC1
colour blocks with a **4-bit alpha plane** apiece, 8 bits per texel against ETC1's 4. The decoder for
it already existed — the emulated GPU has decoded PICA texture format `0xD` since the fragment
pipeline work in Part XVIII — but the CGFX texture path had only ever seen `0x675A`, plain ETC1,
because that is all 3D Land's banner uses. One case in the format switch, and the size check
(`width × height × bpp / 8 == dataSize`) confirms it: 8 bpp, not 4.

### The ™ was an opaque black square

The first export put a black box beside the logo. The mesh, the UVs and the texture were all correct:
`COMMON3` is a 16×16 **ETC1** image of a white "TM" on black, and ETC1 has no alpha channel at all, so
sampled as a picture it *is* an opaque black square.

The material says otherwise, in its **texture-environment stage 0** — five registers the MTOB carries
as a command entry (`0x0C0` sources, `0x0C1` operands, `0x0C2` combine ops, `0x0C3` the stage
constant, `0x0C4` scale), sitting after the mapper blocks so its offset moves by `0x8C` per extra
mapper. Decoded with the same field layout the emulated GPU uses:

| material | colour | alpha |
|----------|--------|-------|
| `COMMON4` (Logo) | `Replace(Texture0)` | `Replace(Texture0.a)` |
| `COMMON5` (Bird) | `Add(Texture0, black)` | `Replace(Texture0.a)` |
| `COMMON1` (™) | **`Replace(Constant)`** = black | **`Replace(Texture0.`*red*`)`** |

The ™ is a **stencil**: a flat black decal whose shape comes entirely from a colour channel. The alpha
operand `2` — "source red" — is the standard way a format without alpha carries a cutout mask, and it
is not a Captain Toad quirk: 3D Land's title planes use the identical operand to take their alpha from
the separate 4-bit mask, which is why the combiner bake written for them was right. The same fact,
read out of the file this time instead of assumed.

`Material.Stage0` now carries the stage, with two derived questions the exporter asks of it —
`AlphaFromRed()` and `ConstantColor()` — and a stencil material bakes to a one-colour image cut by
that channel. It is deliberately not a combiner model: six stages of TEV cannot be folded into glTF's
one texture per material, and the three shapes a banner material actually takes can.

The GLB is in the Studio under a new Captain Toad entry, beside 3D Land's.

---

## Part XXI — The opening diorama, taken out of the cartridge

The oracle has been drawing this stage since Part XVII. Part XXI takes it out of the cartridge as
data: the ruin, its textures and its placement, rebuilt from the files rather than photographed from
the emulator. Four formats had to come apart, and none of them is the CGFX the HOME Menu banner uses —
the *game* assets are a different stack entirely.

```
/StageData/Season1OpeningStageMap1.szs   Yaz0 → SARC → BYML   where things stand
/ObjectData/Season1OpeningStepA.szs      Yaz0 → SARC → BCH    the ruin, and its collision
/ObjectData/Season1OpeningTextures.szs   Yaz0 → SARC → BCH    30 textures, nothing else
```

### SARC, and an archive that checks itself

Every `.szs` is Yaz0 (already decompressed for the message archives in Part IV) wrapping a **SARC**:
a header, an SFAT node table, an SFNT string table. What makes it pleasant is that a node stores both
its name's *offset* and a **hash of the name**, under a key the header gives. Rehashing every name and
demanding it match proves the node stride, the string-table base, the name lengths and the hash key
simultaneously — a layout read wrongly cannot produce names that hash correctly. `ParseSARC` insists on
it rather than treating it as optional.

### BCH — where a mesh is a command list

The models are **BCH** ("Binary CTR H3D"). Its header gives six sections an offset and a length that
**tile the file exactly**, which is what pins every offset in it: the relocation table's end is the
file's end, to the byte, on every `.bch` in the cartridge. The main header is fifteen
(pointer table, count, dictionary) triples — models, materials, shaders, textures, LUTs, lights,
cameras, fogs, the animation families, scenes — and the dictionaries are patricia trees a linear walk
enumerates, exactly like CGFX's.

Then the part worth the trip: **a mesh does not describe its geometry, it programs it.** The vertex
format, the buffer offset, the stride, the index array and the draw call are stored as **PICA register
writes**, in the same command encoding the game submits at run time. So the decoder does not parse a
vertex-buffer struct — it runs the mesh's command lists through a bare register file with the same
`DecodePICA` the emulated GPU uses, and then reads registers `0x200`-`0x205`, `0x227`, `0x228` and
`0x2BB` exactly as `gpu_raster.go` reads them when the game draws. The attribute→shader-input
permutation of Part XVII does double duty here: because the H3D vertex shader takes position in `v0`,
normal in `v1`, tangent in `v2`, colour in `v3` and texture coordinates in `v4`-`v6`, **the permutation
is the file's own statement of what each attribute means**, and no attribute order has to be assumed.
The stage's vertices come out `3f pos + 3f normal + 2f uv + 3f tangent + 4ub colour` = 48 bytes,
matching the configured stride exactly.

**The index width is not where the hardware keeps it.** Register `0x227` has a bit for 8- versus
16-bit indices, and in the file that bit is *zero* for every mesh — the address it shares the word with
has not been relocated yet, so the whole word is a placeholder. The width is instead carried by the
**relocation type** of the entry that patches that word: `0x58` for 16-bit, `0x5A` for 8-bit, ten of
one and forty-four of the other, fifty-four together, which is exactly the model's face count. Read as
8-bit, a 16-bit mesh decodes to an index list of `0 0 1 0 2 0 …` — the low bytes and their zero high
bytes — which is a real triangle list of half the size, and looks plausible enough to ship.

**The decode proves itself.** A model's meshes take their vertices from one shared array, each naming
its own byte offset into it, and the arrays **tile end to end**: mesh *k*'s offset plus its vertex
count times its stride is exactly where the next mesh begins. The vertex count is stored nowhere — it
comes from the largest index in the mesh's index array, so it is only right if the index *width* is
right. All 53 arrays tiling with no gap and no overlap is a coincidence nothing but a correct decode
produces, and it is what `TestBCHVertexArraysTile` asserts. Mutating either the stride or the index
width breaks it.

Textures are the same idea: a texture object is a triple of command lists, one per unit, and the
unit-0 list carries the size (`0x082`), the format (`0x08E`) and the pixel offset (`0x085`). The
stage's thirty are all ETC1 or ETC1A4 — decoders Part XVI already built for the emulated GPU, now
factored out to run on a byte slice.

### The material names its texture, the object names its archive

A model carries a per-material binding table of 0x2C-byte entries whose slot 0 is the texture the
material samples on unit 0. The stage's textures are not in the model's own file at all, and nothing is
inferred from the filename: the object archive's **`InitModel.byml` names its texture archive**
(`TextureArc: Season1OpeningTextures`), and the material names the texture inside it.

One trap. An unset slot is a zero word — and zero is a *valid* string-table offset, landing on the
table's first entry, which in this file is `$shadowmap`. Resolving it gives the two vertex-coloured
grass materials a confident, wrong texture name. An unset slot has to stay empty.

### BYML, and where the ruin stands

The configuration files are **BYML** ("binary YAML"): a tree of dictionaries, arrays and scalars over
two string tables, one for keys and one for values. The opening stage's map decodes to nine lists —
`ObjectList`, `SkyList`, `PlayerList`, `GoalList`, `DemoObjList`, `AreaList`, `Rails`, `Objs`,
`GoalList` — and the placements read straight off:

| list | object | at |
|------|--------|-----|
| ObjectList | `Season1OpeningStepA` | (200, −750, 50) — the ruin |
| SkyList | `SkySeason1Opening` | (0, 850, 0) |
| PlayerList | `Player` | (−400, −800, 550), yaw 90 |
| GoalList | `KinopioBrigadeChecker` | (550, 50, −150) — the star atop the ruin |

It even keeps the authoring path it was built from, `…/StageData/Season1OpeningStage/Map/Season1Opening
StageMap.muunt`.

### What is in the Studio, and what is not

`webexport` reads the map, loads each placed object's archive, resolves its textures through the
archive its own `InitModel.byml` names, and writes the stage as one GLB with every object at the map's
transform — the opening diorama, 7,552 triangles, textured. It is a new **Stages** entry beside the
banner.

Two things are deliberately left out, and are the next pass rather than approximations shipped as
fact:

- **The sky.** `SkyList` places a dome whose material binds only a cloud *mask* to unit 0; the blue
  behind the clouds (`SkyBlueSunny_color`, a 512×64 texture sitting unbound in the same archive) is
  applied by something this decoder does not yet read. Exporting the dome as it stands would ship a
  black sky, so it is not exported.
- **The characters.** Toad, the birds and the star are placed by `PlayerList`, `ObjectList` and
  `GoalList` and are skinned, animated objects — `KinopioAnimationSeason1OpeningStage.szs` alone is
  2.6 MB. They are the object pass.

The albedo comes out much lighter than the stage looks in play, and that is correct: the stone's
texture really is near-white (mean RGB 200, 200, 191) and the dark, mossy look the game has is its
**fragment lighting**, the unit Part XVIII spent six bugs on. An unlit extraction shows the albedo and
the baked ambient occlusion in the vertex colours, which is what every other model in the Studio shows.

### Every texture has an alpha plane. Almost none of them means it.

The first export was full of holes: the floor showed through between the clover leaves, and the stone
walls and paving were pocked with transparent patches. The cause is a property of the *format*, not of
any one asset. The stage's albedos are **ETC1A4**, so every single one carries a 4-bit alpha plane
whether or not anything samples it as coverage — and an exporter that turns a texture with alpha into
an alpha-masked material will cut holes wherever that plane happens to dip.

The textures say so themselves, if you look at the histograms rather than the pixels. They fall into
two populations: the genuine cut-outs run the full 0-255 (`clowers_alb`, `tile_alb`, `moss_alb`,
`Ground_Grass`), while the rest sit in narrow mid-range bands — and **all five stone albedos carry the
identical band, 85 to 221, mean 149.2**. Five different textures with the same alpha histogram is not
coverage.

The authority is the material, and it keeps its answer where it keeps everything else: in the PICA
state it programs. Each material object carries its own command list at `+0xC8`, and two of its
registers settle the question — `0x101`, the blend function, and `0x104`, the alpha test:

| | value | meaning |
|---|---|---|
| blend function `0x101` | `0x01010000` on **all 53** | src×ONE + dst×ZERO — no blending anywhere |
| alpha test `0x104` | `0x8061` on 10, disabled on 43 | GREATER than 128 — a cut-out, on ten materials |

So: nothing in this stage blends; ten materials cut out at a threshold of 128; the other forty-three
ignore alpha completely. The ten are exactly the ones that should be — the clovers, the ivy billboards,
the moss, the tile decal, the grass overlays. `glb.Prim` grew an `Opaque` flag to say so, since the
writer had only ever emitted `MASK` or `BLEND` for a textured material.

Fixing it brought back more than the holes: the **soil layers** banding the underside of the diorama
had been invisible, because `SoilLayer00_AI_alb`'s alpha plane sits at 17-85 and a 0.5 cutoff deletes
every texel of it. The island had been floating on nothing.

*A format that obliges every texture to carry alpha has made alpha uninformative; the material is the
only thing that knows whether it is coverage.*

---

## Part XXII — The light rig, and why it is baked

The diorama's albedo is near-white stone and bright green lawn; everything that makes it *look* like
Captain Toad is the lighting. Part XVIII made the emulator's fragment-lighting unit correct. This gets
the same rig out of the cartridge and into the Studio.

### Where a stage keeps its lights

Three files, and none of it is guessed:

```
/ObjectData/Season1OpeningScene.szs   → BCH, Lights group: mainLight, secondLight,
                                        AmbientLight, mainLight_noshadow
/StageData/…Design1.szs → LightAreas.byml   which lights each light area applies
/StageData/…Map1.szs    → AreaList          where those areas are
```

A light object is small and its layout is checkable: flags at `+0x28`, ambient at `+0x30`, and — when
the flags' bit 3 is set — diffuse, two speculars and a direction at `+0x40`. Bit 3 is not asserted, it
is **tested**: across every scene archive in the cartridge, each light that bit marks carries a vector
of exactly unit length, which a misread offset does not produce. Objects without one are `0x34` bytes
long and stop before the field.

The stage's two light areas are commented `default` and `no_shadow` and name `mainLight + AmbientLight`
and `mainLight_noshadow + AmbientLight` — the same colour and the same direction, differing only in
whether the key casts a depth shadow. So the rig is:

| | colour | |
|---|---|---|
| `mainLight` | `#ffbd4d` (255, 189, 77) | a warm key, travelling (−0.431, −0.666, −0.609) |
| `AmbientLight` | `#6d7580` (109, 117, 128) | a cool sky ambient |

**The stored vector is the direction the light travels, not the direction towards it.** Its Y is
negative — the sun is above, so `L = −direction`; reading it the other way lights the underside of
everything.

### Checked against the hardware, not against taste

`bootoracle -gputrace` dumps the lighting unit's state per draw, so the file's values can be compared
with the registers the machine is actually handed while it draws this stage. The terrain's draws report
**two** lights — the count the light area declares — with a global ambient of raw `0x06C1D07F` and a
key diffuse of raw `0x0FA2F04C`. Those are 10-bit fields carrying 8-bit colours: they decode to
(108, 116, 127) and (254, 188, 76), against the file's (109, 117, 128) and (255, 189, 77). One unit of
quantisation apart. The scene file's lights *are* the ones the hardware gets.

(The same trace shows other draws using four lights, two of them positional with spotlight and distance
attenuation, and the key's colour sliding down through a series of dimmer values — the characters' own
light set, and the intro's fade. Neither belongs to the terrain.)

### Why the rig is baked rather than handed to the renderer

The material's combiner is `PrimaryFragmentColor × Texture1` — the lighting unit's output times the
albedo — and a later stage multiplies by the vertex colour, which carries the artists' baked ambient
occlusion. **All three multiplies happen on 8-bit values, in gamma space.**

A linear-space renderer does not reproduce that. Computing `albedo × colour × N·L` in linear and
encoding the result to sRGB returns the colour product unchanged but the geometric factor as
`N·L^(1/2.2)` — at a typical 0.67 that is 0.83, a visibly flatter and brighter picture than the
console's. The lighting model is not the disagreement; the space the multiply happens in is.

So the lit term is evaluated per vertex in gamma space, multiplied by the occlusion exactly as the
combiner does, and written to `COLOR_0` — converted to linear on the way out, because glTF defines
that attribute as linear. The material stays unlit and performs the one multiply that is left. Per
vertex is not an approximation here: the normal is a vertex attribute, so a hard edge already has split
vertices and a face's shading is constant across it either way.

The scene's lights are still published in the level document (`scene.lights`, with the direction
already negated into the towards-the-light sense). They are decoded fact, and they are what a renderer
that wanted to light this dynamically would need.

### How close it gets

Measured on the same surfaces, ours against a frame the oracle drew:

| | top / side brightness ratio |
|---|---|
| the oracle's frame | 1.47, 1.48, 1.41 |
| the Studio's | 1.26, 1.34, 1.22 |

and the lit stone's hue is warmer than the console's (blue at 0.74 of red against the reference's
0.93). Both residuals point the same way: **what is modelled is the diffuse rig — ambient plus `N·L` —
and what is not is everything that adds a neutral highlight or takes light away.** The shadow map is
the next section; the specular and Fresnel lookup tables and the unit-2 normal maps are still on the
floor.

### The shadow: the one thing that composes with a bake

A shadow **multiplies**, and a multiply is the single operation that survives the gamma/linear round
trip unchanged — `(a·s)` and `(a^2.2·s^2.2)^(1/2.2)` are the same number. So unlike the diffuse rig, the
shadow does not have to be baked to be exact: it can be evaluated live, in whichever space the renderer
happens to be in, and still land on the value the console computes.

The caster was already in the cartridge and already being skipped. Every object archive holds a coarse
second model beside its visible one — `Season1OpeningStepA_shd`, 136 triangles — and its
`InitExecutor.byml` lists it under the draw categories `デプスシャドウ[地形]` and `Ｚプリパス[地形]`,
depth-shadow and Z-prepass. That is what the game renders its shadow map from, so it is what the Studio
renders its shadow map from: the export publishes it as a level layer with role `shadow`, geometry that
exists to be rendered from the light and never to be seen.

**And the stage ships the numbers for it.** The Design archive's `DepthShadow.byml` holds the pass's
own configuration, and reading it settled two bugs that guessing had put in:

| | |
|---|---|
| `Near` 500, `Far` 4800 | the depth range the pass renders over |
| `BiasFactor` 0.0185 | a fraction **of that range** — about eighty world units |
| `ShadowMapWidth/Height` 512 | not the 2048 a modern default reaches for |
| `ColorA` 150 | how much of the light the shadow takes: 59%, not all of it |

The bias is the interesting one, because eighty units is not an acne epsilon — **it is the stand-off of
the caster**. The proxy is a coarse *shell* stretched over the light-facing side of the object (128 of
its 136 triangles face the light, and 132 of its edges are boundaries, so it is not a closed solid and
cannot be cast from its back faces either), and it stands well clear of the surface it stands for.
Compared against with a small bias it does exactly what a first attempt here did: it moirés every face
lying near the shell, and it shadows every recess the shell bridges over — a slab set back under an
overhang came out uniformly dark when the game has it lit. With the guest's own bias both go away.

`site/src/shadowmap.js` renders the proxy once — the scene and the light are both static — into a
packed-depth target from an orthographic camera over the guest's near/far, and patches every other
layer's material to multiply by the comparison. Two details make it right rather than decorative:

- The injection sits **after `<colorspace_fragment>`**, where `gl_FragColor` is already in the output's
  gamma space. That is the space the guest's combiner multiplies in, so the multiply needs no
  conversion at all.
- **A shadow removes one light's contribution, not all of them.** The hardware's shadow attenuation
  scales the *casting light's* term; the ambient — and any second light that does not cast — keeps
  shining inside the shadow. The baked vertex colour holds `ambient + Σ light·N·L`, so the factor is
  `(everything except the caster) / (everything)`, computed per fragment from the same colours the bake
  used, and scaled by the guest's own shadow strength. A constant "shadow floor" would have been wrong
  in both hue and depth. Two things fall out of the exact form for free: the factor reaches 1 precisely
  where the caster's `N·L` reaches 0, so a face already turned away from the light is not darkened
  twice, and the shadowed stone comes out the cool blue-grey of the sky ambient rather than black —
  which is what the console shows.

*Every number this pass needed was in the stage's own files. The two artefacts came from the two that
were not read.*

---

## Part XXIII — The characters: a skeleton that proves itself, and a space that does not

Toad, Toadette and the star are skinned, animated actors, and the archives are generous:
`Kinopio.szs` is a 22-bone character with 40 meshes, `KinopicoNpc.szs` a 28-bone one with 38, and
`KinopioAnimationSeason1OpeningStage.szs` carries **148 skeletal animations** for the opening alone
(plus 600 visibility animations and 5 material ones). Toadette brings her own 50.

### The skeleton, and the invariant that settles it

A model header carries its skeleton at `+0x70` with the bone count at `+0x74` — zero for the terrain,
22 for Toad, which is how the field was found. A bone is `0x64` bytes: flags, an int16 parent, scale,
an XYZ Euler triple in radians, a translation, a 3×4 matrix, and its name last at `+0x5C`. Toad's
reads as a clean humanoid — `AllRoot → JointRoot → Hip → FootL/ToeL, FootR/ToeR`, and
`Spine1 → Spine2 → Head → HeadBelt → HeadLight`, that last one being the lamp on his cap, plus a
rucksack chain and two arms.

The 3×4 is the inverse bind matrix, and **a skeleton carries the proof of its own decode**: compose
each bone's world matrix down the parent chain and multiply by the inverse bind beside it, and the
result must be the identity, because that is what "inverse bind" means. The check settles four things
at once — the bone stride, the parent links, the Euler order and the matrix layout — and it is sharp.
The rotation is **Rz·Ry·Rx**, the X term applied first:

| order | worst residual |
|---|---|
| `Rz·Ry·Rx` | **0.0096** |
| next best | 30.7 |
| the one after | 66.7 |

measured over Toadette's skeleton. It has to be hers: eight of her bones turn about all three axes,
and Toad's turn about at most two, which cannot tell those orderings apart at all. *A degenerate
sample agrees with every hypothesis.*

### The skin

The vertex format carries the rest. Shader inputs 7 and 8 are the bone index and the bone weight, and
a mesh takes one of three shapes: indices and weights (smooth, two influences), indices alone (rigid,
one bone per vertex), or neither (rigid, the whole draw on one bone). The weights are unsigned bytes
and they are **percentages** — a two-influence vertex stores 70 and 30, not 179 and 76 — which the
decoder both relies on and checks: every skinned vertex's influences sum to one, over 4,470 of them.

A vertex does not name a bone directly. It names a slot in the **matrix palette** the face header
carries as twenty u16s before its command list — the same `0x2C` bytes that are all zero on the
terrain — and the palette names the bone. An upper-body mesh's palette reads
`JointRoot, HeadBelt, Spine2, ShoulderL, ArmL1, Head, HandL, ArmL2, ArmR1, ArmR2, HandR, ShoulderR`,
which is exactly what an upper body needs.

One more rule fell out on the way: a vertex's attributes may fall short of its stride by up to three
bytes, because a vertex is padded to a multiple of four. The tolerance is exactly the alignment and
no more — the terrain's 48-byte vertices sum to 48 exactly, and Toad's rigid meshes sum to 37 in a
stride of 40.

### What is not yet right: the space the vertices live in

Assembling the character is still open, and the writeup says so rather than shipping a guess. Neither
of the two obvious rules puts Toad together:

| skinning matrix | resulting extent |
|---|---|
| `world × invBind` (glTF's rule) | y from −103 to 86 — a body hanging below the floor |
| `world` alone (vertices bone-local) | y from −18 to 187 — closer to standing, but wrong in detail |

and both render as a jumble. The skeleton itself is not in doubt: its bones land where a Toad's should
(Hip at y = 40, HeadBelt at 124, HeadLight at 146 and 39 forward — the lamp). What is in doubt is how
the stored positions relate to it. The next move is to read the vertex shader the model names
(`DefaultShader-rs-nl`) rather than to keep proposing rules for it to disagree with: the shader is in
the cartridge, it is what the hardware runs, and it will say which matrix it multiplies by.


## Part XXIV — Two skinning spaces, and the game that said which

Part XXIII left the characters decoded and not assembled. The skeleton was right and the skin was
right, and neither of the two rules that could join them produced a Toad. The reason turns out to be
that the question had a false premise: **a character does not have one skinning rule. It has two, and
the file says per mesh which applies.**

### Reading the register instead of proposing a third rule

The PICA200 has no matrix-palette registers. A skinned draw's bone matrices arrive as ordinary
vertex-shader **float uniforms**, uploaded through `0x2C0`/`0x2C1` — which the emulator already
decodes, and never printed. Two instruments closed that gap:

- **`-drawcensus FILE`** writes one line per draw: index count, vertex stride, and the
  attribute→shader-input map. It is how a mesh decoded from a file is *located* among the 44,000
  draws of a frame, and the match is not approximate. Toad's rucksack is 562 vertices, 1,914 indices,
  60-byte stride, with attributes reaching inputs 7 and 8 — the bone index and weight. Exactly one
  draw in the frame looks like that.
- **`-gputracefrom N` + `-gpuuniforms`** then print that draw's whole uniform file.

The answer was in it at a glance. Uniforms **c25 onward hold one three-row matrix per palette slot,
in the palette's own order** — proved by three draws with different palettes: the rucksack (six
bones) fills c25–c42; the body (four bones, a prefix of the same list) is handed the first four of
them unchanged; and the scarf (three bones, 8/14/18) is handed the rucksack's slots 1–3 verbatim.

And the matrices say two different things:

| mesh | mode | what the game hands the shader |
|---|---|---|
| rucksack, body, scarf, skin, buckle | 1 | six matrices all within **three units** of the placement itself |
| lamp strap (HeadBelt) | 2 | a matrix translating 123 units **above** the placement |
| head (Head) | 2 | 72 above |
| headlamp (HeadLight) | 2 | 146 above |

With Toad placed at (−400, −800, 550): the first row is `world × invBind` at a pose near the bind
pose, which is the placement and nothing else — six bones' *world* matrices would be a hundred units
apart in Y. The rest are bone world matrices with no inverse bind anywhere: HeadBelt stands 124 units
up the bind pose and is handed −677; Head stands at 72 and is handed −728; HeadLight at 146, −653.

So:

- **mode 1 (smooth)** — vertices in the *model's* space, matrix `world(bone) × invBind(bone)`. glTF's
  own rule.
- **mode 2 (rigid)** — vertices in the *bone's* space, matrix `world(bone)`.
- **mode 0 (none)** — no per-vertex skinning; the same rule as rigid. The goal star settles that one:
  its glow quads are modelled centred on nothing, and the bone they name stands 73 units up, exactly
  where the star they surround is.

Applying either rule to a whole character breaks the half that wanted the other. That is precisely
what "the body hangs below the floor" and "closer to standing, still a jumble" were, and why no
amount of staring at either picture was going to resolve it.

### The palette's first two words were never bones

The face header does not begin with the palette. It begins with the **skinning mode** and the **bone
count**, and the bones follow — so reading the reserved block straight through put every vertex's
joint two slots wrong. The game's own uniform uploads say so: each draw is handed exactly *count*
matrices. With the count read properly, every one of Toad's 4,470 vertex influences names a palette
slot that exists, which a two-slot shift does not survive.

### The check that has a control

`TestBCHSkinSpaces` asserts that every mesh, assembled at the bind pose by the header's own rule,
sits closer to the bones its palette names than the same mesh does under the opposite rule — 78
meshes across the two characters and the star, none worse, and all but the degenerate ones decisively
better.

The first version of that test used a **whole-character bounding box**, and it had to be thrown away:
the all-rigid hypothesis produces a box of the right size and shape while scattering the parts inside
it. It agreed with a wrong model because the criterion could not disagree. The test now logs the fact
so the next person does not reach for the same measure.

### A character's colour is not in its texture

With Toad standing, he was standing in the wrong colours: a **black cap with white spots**. The cap's
32×32 texture really is black with a white blob, because it is not an albedo at all — it is a *mask*,
and the material builds the surface out of it with two stage constants:

    stage2   (1 − mask) × (242,246,255)     the white cap
    stage3   mask × (196,30,25) + prev      the red spots

A decoder that stops at the texture name paints the inverse of the character. The colours are in the
material, and **the material is a program** — six PICA combiner stages, which the BCH stores as the
same register writes the running game submits.

`BCHMaterial.BakeAlbedo` runs that program with the lighting inputs *neutral* — diffuse white,
specular black, vertex colour white — which is what "unlit albedo" means for a combiner and which on
these materials reduces exactly: the stages that consume the lighting become identities. The stage
loop is factored out of the GPU's own fragment path so the same code does both, and the oracle's
frame is byte-identical across the move.

⚠ **Only the stages the material writes are the material's.** PICA registers latch, so the register
file during a draw also holds whatever the *previous* material left in the stages this one does not
touch. The file is the only thing that says which are which.

### Three more things the material had to be asked, not told

1. **The vertex colour is a scalar.** Toad's meshes carry one whose red channel runs the full 0–255
   range while green and blue sit at 255, and every colour stage that reads it reads it through the
   *source-red* operand. Multiplying a baked light by all three channels drags green and blue down
   independently of red — and painted him teal. The stage terrain, by contrast, reads `vtxcol(rgb)`
   and means it. `BCHMaterial.VertexColorScalar()` asks the material rather than the game.
2. **It is not even a plain multiply.** His cap's chain ends `(vtxcol.r + 0.46) × everything` — an
   *add* before the multiply, so an occlusion above about half does not darken him at all. Treating
   it as a multiplier makes him a third too dark. The shading factor is therefore computed as a
   **ratio of the material's own chain to itself**, lit over neutral, which cancels whatever the chain
   does to the texture and leaves exactly the part that varies per vertex. `BakeCheck` measures
   whether that ratio depends on which texel it is taken at — for 38 of Toad's 40 materials it is
   under two per cent, and the two that fail are the ones sampling an environment map.
3. **A mesh with no colour attribute means white, not black.** His hands carry position, normal and a
   bone index and nothing else, and their chain ends `× vtxcol`. A zero there renders them black; the
   shader input is simply never written, and the identity is what the hardware's picture implies.

### The animations

BCH group 8. The object is a name, a loop flag, a **curve count**, a frame count and a table of
per-bone elements; an element is nine transform slots — scale XYZ, rotate XYZ, translate XYZ — each
either absent, a stored constant, or a pointer to a hermite curve. The flags carry both: bit `16+s`
marks slot *s* absent, and bit `6+s` (for the six scale and rotation slots) or `7+s` (for the three
translation ones) marks it constant. **Bit 12 is skipped**, and reading straight through it makes
every translation look like a curve.

Two counts prove the whole reading, and neither can be satisfied by accident:

- the header's own curve count equals the number of slots this decode calls curves — over all 198
  animations of the two characters;
- every curve stores its ordinal within the animation, and it comes out as exactly its position in
  that enumeration.

The key encodings are named by the high byte of the encoding word, and their widths are *measured*
rather than assumed: the curve headers and key blocks tile the section, so the distance from a key
block to whatever starts next, over the key count, is the key size. It comes out constant per
encoding — 96 bits for encoding 3 across 9,797 curves, 128 for encoding 0, 64 for encoding 1.
Encoding 3 (`frame, value, slope`), 0 (`frame, value, in-slope, out-slope`) and 1 (a packed word
holding the frame in the low twelve bits and a signed twenty-bit value above it, then a float slope)
are implemented; they are 97.8% of the curves. **The rest halt loudly.** Their frame fields are
already pinned by the same invariant that pinned encoding 1's — the frames must ascend, start no
earlier than the curve's start and land exactly on its end — and are recorded here as the lead:
encoding 5 and 2 carry the frame in the low **8** bits, encoding 7 in the low **12**. What is not yet
pinned is the width of their value fields, and a value field guessed is a pose invented.

**The frame rate is 30 Hz, and that is a measurement.** Six consecutive rendered frames of the intro
give three distinct matrix palettes: the engine presents at the 3DS's 60 and advances its animations
at half that.

### What the map says, and what it does not

Nothing about the actors is named by convention. The map's `GoalList` entry is a logic object,
`KinopioBrigadeChecker`, which has no archive at all; what it *links* is a **`GoalItem`** — the star,
`/ObjectData/GoalItem.szs` — and a **`PlayerNpc` whose ModelName is `KinopicoNpc`**, which is
Toadette, each with its own translation. Toad comes from `PlayerList`, and the clip the stage starts
him in is written down: the `DemoObjList` entry beside him says `StartAnimation: CourseInRove`.

Their light rig is not the terrain's, and the oracle says how it differs: the fragment-lighting unit
is programmed for character draws with a brighter ambient and a second light. That second light is
**Toad's own headlamp** — his `InitLightActor.byml` declares a `spotLight` on the `HeadLight` joint —
and it is deliberately not baked, because it is aimed every frame and a baked cone would be a lie
about a light that moves with his head. The key light *is* the stage's `mainLight`: the unit is
handed a view-space direction of (0.735, 0.610, 0.294), and rotating that back through the game's own
view matrix gives (0.431, 0.666, 0.609) — the stage file's `mainLight` travel vector negated, to four
decimals. Two independent routes to the same vector.

### What is still open

- **Which clip the intro is playing.** No single animation reproduces the pose the savestate holds;
  the closest is out by about three degrees. The reason is in the file: `InitModel.byml` says
  `BlendAnimMax: 4`, and `AnimInterp.byml` gives a per-clip blend length. The game is blending, and
  matching one clip to a blended pose is the wrong question.
- **The visibility animations.** A character carries every face, hand and eye it can wear — nine
  faces, seven hands a side — and the game shows one of each. The mechanism is decoded: for each
  skeletal clip there are four visibility animations named `<clip>Eye`, `<clip>Face`, `<clip>HandL`,
  `<clip>HandR` (148 clips × 4 ≈ the 600 in the archive), and each names its variants with a constant
  0 or 1 — `WalkFace` says `Face01`. What is *not* decoded is how those names bind to a model's
  meshes, which carry no names of their own. So the exported set is the one the **running game
  draws**, read off the draw census by matching each draw's vertex-buffer offset to a decoded mesh:
  `HandR02`, `HandL02`, `Face00`, `EyeBall`.
- **The headlamp's aim.** The game holds `HeadLight` at a rotation of about 27° that no clip in the
  archive puts there. Something outside the skeletal animation points the lamp.

### Shipped

`site/public/captain-toad-treasure-tracker-3ds/objects/{kinopio,kinopico,goalitem}.glb` — Captain
Toad (2,554 tris, 5 clips), Toadette (2,764, 3) and the goal star (180, 3), each skinned, each with
its own light bake, and each placed in the opening diorama at the position, facing and starting
animation the stage's own map gives it.


## Part XXV — Three bugs the Studio found that my own renderer could not

Part XXIV verified the characters by rendering the shipped GLB with `tools/cmd/rendersheet` and looking
at the picture. They were right in that picture and wrong in the Studio, which is the renderer that
matters. Opening the artefact is not enough if you open it somewhere else.

### 1. A plain clone does not rebind a skeleton

In the viewer, Toad arrived as a heap with his cap under his chin. The cause is one field:

    engine3d._model:  const node = doc.skinnedClone ? SkeletonUtils.clone(src) : src.clone(true);

`Object3D.clone(true)` copies a `SkinnedMesh` but leaves its `skeleton` pointing at the **original**
bones — bones that live under a prototype nobody adds to the scene, so their world matrices are never
updated and stay the identity. Every *bone-space* mesh then draws at the origin while the model-space
ones stay put, which is precisely a character with its rigid parts collapsed into a pile. It is the
Part XXIV finding again, seen from the other end: the two spaces are real, and the failure mode of
losing the skeleton is exactly the failure mode of applying the wrong rule.

`skinnedClone: true` asks for `SkeletonUtils.clone` instead, and also keeps the object out of the
instanced-batching path, which drops skinning altogether. Every other game in the repo that ships a
skinned model already sets it; this export did not. It is now derived, not declared: an actor sets it
when any of its meshes got a skin.

### 2. Texture unit 0 is not always the shadow map

The goal star came out as a solid yellow card with the star behind it. `BakeAlbedo` was substituting
an opaque white for unit 0, on the grounds that the terrain and the characters bind `$shadowmap`
there. The star binds **real textures** to unit 0 and builds its alpha out of them — its glow quad's
chain is `alpha = tex0.r + tex1.r + tex2.r` — so the substitution made every glow and sparkle fully
opaque.

The fix is to stop reasoning about the name and ask the archive: a bound texture the archive holds is
sampled, and a name it does not hold is something the engine renders at run time (the shadow map, a
cube map) and gets the neutral value that means "not there". The glow and the sparkles now blend and
add as the material says.

### 3. A wrap mode read out of a register nobody wrote

Chasing a row of dark slots across the star's face, I decoded the materials' texture-unit wrap modes
and got a clean, plausible answer: **CLAMP_TO_EDGE, all 97 materials across three models.** It was
the decoder's invention. Not one of those materials writes its unit's parameter register, and an
unwritten register reads zero, and zero is CLAMP_TO_EDGE. `WrapModes` now reports whether the
material said anything at all — [[stored-registers-are-not-read-registers]], caught this time by
asking before believing.

What the game actually programs, for the only draws the oracle can be asked about, is REPEAT (the
character textures come back with wrap 2/2), so that is the fallback and it is right for the terrain
and the characters.

The star's slots are still there, and now precisely understood: its eye mesh's UVs span
**u −0.47 … 4.49** — five tiles of a texture holding one eye. A material carries a texture *matrix*
as well as a wrap, and this decoder does not read it yet. See [[uv-set-is-not-a-texture-coord]],
which is the same lesson from the banner: a vertex UV set is not a texture coordinate until the
material's mapper has had its say.


## Part XXVI — Where a texture coordinate actually comes from

The characters were assembled, lit and animated, and still wearing their textures wrongly: Toad's cap
spots ran off its edge, his and Toadette's faces were a smear, and the goal star had a row of dark
slots where its eyes belong. Part XXV had narrowed it to "the material carries a texture matrix and
this decoder does not read it". Both halves turned out to be missing, and neither is where a PICA
programmer would look first.

### Neither the wrap nor the matrix is in a command list

The BCH stores most render state as PICA register writes, and this decoder reads them with the same
`DecodePICA` the emulated GPU uses. So the natural place to look for the sampler's wrap mode is the
texture unit's parameter register — `0x083`/`0x093`/`0x09B`, bits 12-14 and 8-10. **Nothing writes
it.** Not the material's command list (which writes only the TEV stages, the output merger and some
shader uniforms), not the mesh's, and not the texture object's — which carries three command lists of
its own, one per unit, and sets only the dimensions, the address and the format.

Reading it out of the register file anyway produced a clean, plausible, entirely invented answer:
CLAMP_TO_EDGE, for all 97 materials across three models, because an unwritten register reads zero and
zero is CLAMP_TO_EDGE. That is [[stored-registers-are-not-read-registers]] for the third time in this
game, and this time it was caught by asking whether the register had been written before believing
what it said.

They live in two places the decoder had walked past:

- **the wrap** in a 0x10-byte *texture mapper* block per unit at material **+0x110**, wrap S at +1
  and wrap T at +2;
- **the matrix** in the vertex-shader **float uniforms the material uploads** — three rows per unit,
  c11-c13 for unit 0, c14-c16 for unit 1, c17-c19 for unit 2 — which the shader multiplies the vertex
  UV by before sampling.

### Found by asking the game, then finding the field that agrees

The material's own draws are observable, so the answer came first and the field second. The oracle's
per-draw dump prints each texture unit's parameter word, and over Toad's fourteen draws in the intro
savestate that is **sixteen ground-truth wrap pairs** across three units. With that table in hand,
scanning the material object for a byte that *is* the wrap finds +0x121 and +0x122 immediately — and
then the same scan over the other units gives the 0x10 stride.

The first version of that scan found nothing useful and looked like it had: with every material
holding a distinct byte at a given offset, "this byte predicts the wrap" is true of almost any offset
and means nothing. The criterion has to be that the same value implies the same wrap, over few enough
distinct values to be a claim.

### What they were hiding

Neither is a formality:

| material | wrap (unit 1) | matrix |
|---|---|---|
| `KinopioRuckMat00` | REPEAT | identity |
| `KinopioHeadMat00` (the cap) | **MIRRORED_REPEAT** | `u' = 2u − 2, v' = 2v − 3` |
| `Kinopio_Face00` | **CLAMP_TO_EDGE** | `u' = 2u, v' = 2v − 1` |
| `Kinopio_EyeBall` | MIRRORED / REPEAT | `u' = 8u` |
| `GoalItem/Eyes` | MIRRORED / CLAMP_TO_BORDER | `u' = 8u` |

A single guessed mode is wrong for most of a character. And the star's eyes are the clearest case of
what the pair does together: **one texture holding half an eye**, an 8× scale, and a mirrored wrap —
which is how eight tiles become four symmetric eyes, of which the mesh uses two. Guessing REPEAT
tiled the half-eye into a row of slots; guessing CLAMP smeared one across the whole face.

### And one assumption underneath both

Fixing the matrix promptly broke the star's sparkles, which went from cross-shaped flares to solid
white cards. `bchUnitAlbedo = 1` had been true of everything the decoder had met — the terrain and the
characters bind `$shadowmap` to unit 0, so their first real texture is on unit 1 — but the sparkles
and Toad's headlamp put theirs on **unit 0**, and asking a material for unit 1's texture matrix when
it has none returns all zeros, which collapses every UV onto one texel. The albedo unit is now the
lowest-numbered unit whose named texture the archive actually holds, which is a question the data can
answer.

Guarded by `TestBCHMaterialTexCoords`, which checks all sixteen measured wraps and four matrices, and
fails if the fixture stops exercising all three wrap modes.

### The filtering was already right

The 3DS filters bilinearly, and the Studio was already doing it: both titles' manifests declare
`texFilter: "linear"`, every exported GLB's samplers declare `magFilter` LINEAR, and the viewer's
`applyTexFilter` upgrades minification to mipmapped linear on load. The blockiness was the wrap and
the matrix sampling the wrong texels with hard tile seams, not point sampling.

"linear" rather than "bilinear" is the right one of the two, and the cartridge says so: these textures
**ship mip chains**. Their pixel blocks do not tile the data section at `w*h*bpp/8` — a 32x32 ETC1
texture occupies 640 bytes where its base level needs 512 (base plus one level), and a 64x64 one
occupies 10,880 where it needs 8,192 (base plus three). Only mip 0 is decoded and the viewer generates
the rest, which is an approximation of the chain the cartridge carries, not a substitute for one.


## Part XXVII — A filter that empties itself stops filtering

Both characters were still wearing two pairs of eyes, one over the other. It looked like a texture
fault — a wrong tile of the 32x128 eye strip, or the mirrored wrap catching a neighbour — and it was
not a texture fault at all. **The GLB contained six eye meshes.**

A character carries every eye it can wear (`EyeAngry`, `EyeBall`, `EyeBallBig`, `EyeSad`,
`Eyelid00`-`02`) and the export names the one the running game draws. The filter read:

    if len(want) > 0 {
        if !want[mat.Name] { continue }
        delete(want, mat.Name)          // <- also the "have we used this name" tracker
    }

One collection doing two jobs: the membership test, and the record of which names had been matched so
an unused one could be reported. Striking each name off as it matched **emptied the set**, and an
empty set is read one line above as "no filter" — so every mesh after the last match went straight
through. Toad's list ends at his eyeball, which is fifth from last in the model, so out came
`EyeBallBig`, `EyeSad` and three eyelids on top of it.

Nothing about the artefact looked like that. Fourteen names went in, twenty meshes came out, and the
picture was a plausible-looking character with slightly wrong eyes — which is why it survived a
careful look at Toad, Toadette, the star and the diorama. The invariant that catches it is one line,
and it is now enforced: **one mesh out per name in.**

Two sets, and a count that has to match.


## Part XXVIII — The variants stop being a list, and the shadows the game actually casts

Two things the export was doing by hand: naming which of a character's variant
meshes to draw, and deciding who casts a shadow. Both are written down in the
cartridge.

### A mesh knows which node it is

Part XXIV shipped the variant set as a list of material names, read off the
oracle's draw census, because the visibility animations name things like
`Face00` and a BCH mesh appeared to carry no name at all. It carries an index.

The model header holds a **mesh-node name dictionary at +0x7C** with its count at
+0x80 — for Toad, thirty-one entries reading `Kinopio00`, `EyeAngry`, `EyeBall`,
`EyeBallBig`, `EyeSad`, `Eyelid00`-`02`, `Face00`-`08`, `HandL00`-`06`,
`HandR00`-`06`. Exactly the names the visibility animations use. And **mesh
header +0x04 is a u16 index into it**.

That field was found by its shape rather than by reading bytes hopefully: ten of
Toad's forty meshes are the fixed body, so the field had to take one value ten
times over and thirty others once each. One offset in the header does that. The
cross-check is independent of the search — every variant mesh's *material* is
named after the node its header points at, twenty-seven of thirty exactly and the
other three differing only in whether the L in `EyeLid` is capital.

⚠ The dictionary's names sit at `dict+0x18+i*12`, four bytes further in than the
main header's groups put theirs. The two pointers do not address the same part of
the structure, and reading this one with the other's offset returns garbage.

### The visibility value is a bitfield, not a boolean

The elements looked simple — a name, a pointer, a count of one, and a `0` or `1`
behind the pointer — and 336 of the archive's 600 decoded that way. The other 264
had counts like 196 and 1,539.

They are **one bit per frame**. `+0x0C` is the element's own end frame and `+0x18`
the number of frames; the bit blocks tile the data section end to end at
`ceil(frames/32)` words apiece, which is what pins the width and the count
together. Captain Toad blinks, and that is where it is stored.

The reading proves itself frame by frame: **the number of nodes an animation
shows never changes over its frames**, across 650 animations and 63,477 frames.
That is a claim about the decode rather than about one file's habits, because the
two archives disagree about the number — the opening stage's animations are split
per part and show one node each, while Toadette's own put every part in one
animation and show four. A wrong bit order or stride would make either wander.

One trap on the way out: gathering a clip's animations by name prefix. `Walk`
also prefixes `WalkSlowlyFace`, `WalkSpeedyEye` and `WalkWaterHandL` — sixty-four
animations instead of four, with whichever was enumerated first deciding the
face. A visibility animation belongs to the **longest** skeletal clip name that
prefixes its own, and the archive's own clip list is what resolves it.

### Who casts a shadow is also written down

The obvious thing to do with "can the placed objects cast shadows too" is to add
their depth-shadow proxies to the caster layer. They all have one. It would have
been wrong for two of the three.

Every object's `InitExecutor.byml` lists the draw categories it renders into, and
the depth-shadow pass is one of them:

| object | categories |
|---|---|
| `Season1OpeningStepA` (the terrain) | 地形, 撮影用[地形], **デプスシャドウ[地形]**, Ｚプリパス[地形] |
| `GoalItem` (the star) | アイテム[フォワード], **デプスシャドウ[キャラクター]** |
| `Kinopio`, `KinopicoNpc` | プレイヤー — and nothing else |

**The characters do not cast a depth shadow.** What they have is an
`InitShadowMask.byml`: a blob projected beneath them, with its own joint, offset,
radius, drop length and colour —

    Toad       cylinder, JointRoot, offset y 28, radius 85, drop 600, colour (0.2,0.2,0.2)
    Toadette   cylinder, JointRoot, offset y 28, radius 75, drop 150
    the star   ellipsoid, GoalItem, offset y 12, scale 70, drop 100

and that is also why forcing the depth shadow would have failed anyway: the
stage's shadow bias is 0.0185 of a 4,300-unit range, some eighty world units of
stand-off tuned for the terrain's coarse shell — most of a 190-unit character.

So the star joins the caster layer, posed at the frame its placement starts (its
proxy is skinned, unlike the terrain's, so it has to be posed rather than
copied), and all three get their mask: a disc dropped by ray-casting straight
down against the stage's own triangles, at the radius and colour the game states.
Placements now receive the depth shadow too, which they never did.

### And a blend mode that was never written

The first masks came out invisible. They were in the file, at the right place,
with the right colours — and `alphaMode` was absent, because `glb`'s material
writer only decided it *inside the branch that adds a texture*. A primitive that
asks to blend and carries its coverage in COLOR_0 got glTF's default, which is
OPAQUE: the alpha was written to the file and ignored by every renderer that read
it. Blending is a property of the material, not of having a texture.

Then they were invisible again, for a duller reason: a disc fanned from a single
centre vertex to a transparent rim has its alpha fall linearly across the whole
radius, which averages to a smudge. The mask the game projects is a cylinder,
which is mostly solid — so the disc has a full-strength core and only the outer
band fades.

### A shadow is a factor, not a layer of paint

And then it was visible and still wrong. The mask's `Color` is a **multiplier** —
the game darkens what is under it to a fifth of itself — and an alpha-blended
grey disc is not that. Alpha blending moves the surface *towards* grey, so it
lightens dark ground as much as it darkens light ground: over the clover it was
a pale film, over pale stone a grey coin. Both are wrong for the same reason.

So the disc multiplies: a new Retro-X material blend, `multiply`, which the
viewer maps to `dst x src` with a decal's polygon offset and no depth write. It
fades to **white** at the rim rather than to transparent, because multiplying by
one is what "no shadow here" means.

The colours go out through `srgbToLinear` on the way, for the reason
[[gamma-space-multiply]] gives: COLOR_0 is linear by glTF's definition and the
renderer encodes the result back to gamma before blending, which is the space the
guest multiplies in. Skipping the conversion turns a 0.2 shadow into a 0.48 one.

One consequence worth stating: the goal star's placement now plays no animation.
Its idle is `Wait`, which uses a key encoding this decoder still refuses, and
every clip it *does* have — the opening cutscene, the goal poses — flies it off
its pedestal. It stands where the map puts it, and its clips are still there to
play in the object viewer.

