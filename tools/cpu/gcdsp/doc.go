// Package gcdsp is an interpreter for the GameCube/Wii audio DSP — the custom 16-bit
// Macronix-fabbed fixed-point processor that sits behind the console's DSP interface and does
// the sound mixing. It is a CPU core in the same mould as the others in tools/cpu: a state
// struct, a decoder/disassembler, and a stepping interpreter, reached through a small memory
// interface the host implements.
//
// The DSP is a Harvard machine with separate instruction and data address spaces, each 16-bit
// word-addressed:
//
//   - IRAM  0x0000..0x0FFF  instruction RAM — the game's microcode is DMA'd here and run
//   - IROM  0x8000..0x8FFF  instruction ROM — the console's boot ROM (not on the game disc)
//   - DRAM  0x0000..0x0FFF  data RAM — the ucode's working memory and the command buffers
//   - DROM  0x1000..0x1FFF  coefficient ROM — resampling/filter tables (console-resident)
//
// The microcode a game runs is on its disc and is loaded over the DSP-interface mailboxes, so
// running the game's own ucode is fully first-party. The two ROMs, by contrast, live in the
// console silicon, not on the disc. The boot IROM's only job is the mailbox bootstrap that
// receives the ucode — which the host models directly by loading the ucode into IRAM — so a
// running ucode does not need it. The coefficient DROM is read by the mixing ucode for its
// resampling tables; a core with no DROM behind it halts loudly the moment the ucode reads one,
// naming the address, rather than returning a plausible zero.
//
// The register file is 32 entries, most of them 16-bit, with the accumulators built from
// several of them (see cpu.go). The instruction set is documented hardware, implemented here
// from that documentation: the Duddie/Tratax DSP manual and the gamecube-tools opcode table
// for the paper-documented core, and — for the corners the paper predates (the shift-by-
// register family, the address-register wrap formulas, the mode-bit polarities, rounding and
// saturation details) — the hardware-verified behaviour recorded in Dolphin's DSP-LLE
// interpreter, consulted with the user's explicit approval as the ground truth of record
// where the documentation runs out. Everything is reimplemented; nothing is copied.
package gcdsp
