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
