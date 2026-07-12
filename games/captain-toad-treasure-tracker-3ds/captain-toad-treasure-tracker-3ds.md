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

**The frontier has moved past audio**, and its shape is pinned: the main thread loops a ~once-per-
second timed poll (helper `0x001211A8` → sleep wrapper `0x003006D0(1)` → the poll object at heap
`0x146EFE14`, vtable `0x004D7250`, poll function `0x0033AF78` — a 64-bit tick-delta dispatcher). The
poll's argument points *into worker thread 4's live stack frame* (a condvar/critical-section pair
around `arb@0x17BDFD08`, precisely where that prio-38 worker parks waiting for a job); worker 5
(prio 35) parks the same way on `arb@0x0054F59C`. So the main thread awaits a completion its worker
should produce, and the worker waits for a job nobody posts — only one file has been opened (~14
reads), and GSP is never brought up. A 9-billion-instruction probe reaches the identical steady
state, so this is not a timeout that expires. Tracing who was supposed to post that worker's job is
the next session's work — note each poll second costs ~100M *executed* instructions now, because the
game's own mixer really runs every audio frame.
