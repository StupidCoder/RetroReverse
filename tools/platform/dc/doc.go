// Package dc models the Sega Dreamcast around the SH-4 core in
// tools/cpu/sh4: 16 MB main RAM, 8 MB video RAM, 2 MB sound RAM, the Holly
// system bus with its three-level interrupt fan-in, the GD-ROM drive served
// from a cdrdao-style rip, and the CRTC scanout of the framebuffer registers.
//
// What is high-level emulation here, and why — the package's honesty ledger:
//
//   - The BIOS. No boot ROM image is required or read. The oracle loads the
//     disc's own 1ST_READ.BIN directly (boot.go) and services the syscall
//     vectors the BIOS would have installed at 8C0000B0-8C0000E0 in Go
//     (bios.go), gdrom reads included. Everything else the BIOS would have
//     left in low RAM is poisoned, so a dependency on an unplanted value
//     becomes a named gap, not an invented fact.
//
//   - The PowerVR tile accelerator. Not modeled this milestone: TA FIFO
//     writes are counted (Machine.TAWrites) and the PVR register file is
//     stored raw, so the game's configuration survives for the renderer to
//     come. The framebuffer scanout (video.go) is real, which is enough to
//     see anything drawn through the CPU.
//
//   - The AICA and its ARM7. Sound RAM exists and is writable — the game
//     uploads its driver there — but the ARM never runs and every AICA
//     register access is gap-logged.
//
//   - The GD-ROM ATA register protocol. The syscall HLE is the supported
//     path; direct drive-register traffic is gap-logged until a game is seen
//     using it.
//
// Everything unmodelled follows one discipline: log once, count, surface
// through Census(), return zero — never a plausible "ready" bit. The census
// after a run is the worklist, in the order the machine itself proposes it.
package dc
