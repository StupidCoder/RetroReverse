// Package gc models the Nintendo GameCube: an IBM "Gekko" PowerPC 750
// (tools/cpu/gekko) wired to 24 MiB of main memory, 16 MiB of auxiliary RAM, the
// "Flipper" graphics chip, and a disc drive — and it boots the retail disc and runs
// it, with the tracing and profiling instrumentation the other machine models in
// this repository provide.
//
// What is distinctive about this machine is what it does *not* need. The GameCube
// has no operating system in the sense the PlayStation 2 or the 3DS has one. There
// is no kernel to answer syscalls, no service manager, no second processor running
// modules loaded off the disc. The routines a program calls — OSInit, DVDRead,
// GXBegin, the thread scheduler, the exception vectors — are library code
// *statically linked into the game's own executable*, and they run on the Gekko as
// ordinary PowerPC instructions. So there is nothing here to high-level emulate:
// only hardware to model. The N64 is the same shape (see tools/platform/n64, whose
// libultra likewise ships inside the cartridge); the PS2 and the 3DS are not.
//
// The price of that is that the hardware has to be right. Every routine that would
// otherwise have been serviced in Go instead pokes a real register, and there is no
// seam at which to paper over a wrong bit: a mistake in the disc interface or the
// video clock does not produce a wrong answer, it produces a machine that stops.
// The discipline this package therefore keeps is the N64's — model the register, or
// halt loudly naming it, but never return a plausible zero.
//
// # What is not on the disc
//
// Three things the machine needs are resident in the console rather than on the
// medium, and each is therefore *substituted* — named, and justified, rather than
// quietly faked. They are the whole of the high-level emulation in this package:
//
//   - The IPL. The console's boot ROM initializes a page of low-memory globals and
//     then reads the disc's own apploader and runs it. We do the same: fill the
//     globals, then run the apploader — which is real PowerPC code, on the disc, and
//     it is the apploader that goes on to load the game's executable and filesystem.
//     So the machine needs no BIOS image, and every instruction after the globals
//     are laid down is the disc's own. See ipl.go.
//   - The SRAM and real-time clock behind the EXI bus, which hold console settings —
//     video mode, language, the counter bias. See exi.go.
//   - The DSP's boot ROM, which the sound system handshakes with over a pair of
//     mailboxes before it uploads the microcode that the disc *does* carry. See dsp.go.
//
// This is the same class of substitution as the N64's PIF handoff (tools/platform/n64
// boot.go) and for the same reason: it is not on the medium, so it cannot be derived
// from the medium.
//
// Addresses in this package are physical. The Gekko translates virtual addresses —
// through the block-address translation registers, which is how a GameCube program
// reaches memory, rather than through a page table — before touching the bus.
package gc
