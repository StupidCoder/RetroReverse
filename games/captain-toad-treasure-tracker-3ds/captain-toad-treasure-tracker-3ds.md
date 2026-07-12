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

The conclusion is a clean statement of the next piece of work: **Captain Toad needs a real `dsp::DSP`
model** — the component load, the pipe protocol, the shared-memory audio frame structures, and the
frame interrupt — a subsystem of a scale comparable to the PICA200 GPU, and the natural next platform
phase now that a second title demands it.

---

**Status.** Container layer: complete, unchanged. CPU: two ARMv6K gaps found and closed. OS: four new
service commands, all traced. Reach: the full OS handshake and six live threads, stopped by an absent
DSP rather than by anything unknown.
