// Package psp is the Sony PSP machine oracle and its UMD format libraries.
//
// It boots a PSP game image on the Allegrex core (tools/cpu/allegrex) the way the
// other platform packages in this repo do: a Machine decodes the PSP memory map and
// hardware I/O, high-level-emulates the PSP kernel (the sceXxx syscall surface), and
// exposes the tracing/watch/savestate instrumentation used to reverse the game.
//
// The format libraries stand alone and are useful without the CPU:
//
//   - cso.go   — the CISO ("CSO") compressed-ISO container UMD dumps ship in
//   - iso.go   — the ISO 9660 UMD filesystem
//   - sfo.go   — PARAM.SFO title metadata
//   - kirk.go  — the KIRK crypto engine that decrypts the ~PSP executable
//   - prx.go / elf.go — the decrypted PRX/ELF module loader
//
// cmd/pspinfo is the Part I inspector: list the disc, dump PARAM.SFO, decrypt and
// describe the boot executable, extract files.
package psp
