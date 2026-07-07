// Package threedo provides the platform support for reverse-engineering 3DO
// Interactive Multiplayer discs. Like tools/psx it is built in layers:
//
//   - operafs.go  reads the big-endian "Opera" file system off a CD image
//     (volume label, directories, avatar copies, file extraction).
//   - cel.go      decodes the Madam cel-engine image format (CCB + PLUT +
//     coded/uncoded pixels, packed and unpacked) into Go images.
//   - aif.go      loads the ARM Image Format executable the disc boots.
//   - machine.go  is the boot oracle: an ARM60 (tools/arm60) wired to the 3DO
//     memory map with Madam/Clio stubbed and the Portfolio OS high-level
//     emulated, so a game can be run and traced.
//
// The 3DO CPU is an ARM60 (ARMv3, big-endian); its execution core lives in the
// sibling package tools/arm60. Everything here is derived from the disc and from
// public 3DO platform documentation, never from game-specific sources.
package threedo
