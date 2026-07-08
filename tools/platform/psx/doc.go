// Package psx reads Sony PlayStation software and emulates the machine.
//
// It has three parts, built up across the toolchain:
//
//   - cd.go: a reader for a raw CD image (2352-byte sectors) with an ISO 9660
//     filesystem on top — enough to list and extract the files off a game disc,
//     in particular the boot executable named in SYSTEM.CNF.
//   - exe.go: a loader for the PS-X EXE executable format (the "PS-X EXE" header
//     carries the initial PC/GP/SP and the load address of the text image).
//   - machine.go and friends: the oracle — a MIPS R3000 CPU (tools/mips) wired
//     to 2 MiB of RAM, the scratchpad, the hardware I/O registers and a
//     high-level-emulated BIOS, with the tracing/profiling instrumentation the
//     other machine models in this repo expose (see tools/dos, tools/nds).
//
// The disc used for development, Ridge Racer (NTSC-U, "RIDGERACERUSA", boot
// SCUS_943.00), is a single Mode-2/Form-1 data track; the game's music lives on
// separate Redbook audio tracks that are not part of this image.
package psx
