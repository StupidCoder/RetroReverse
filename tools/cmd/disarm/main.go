// disarm linearly disassembles ARM code from a raw binary — an extracted .bin, a
// memory dump, or a slice of a cartridge. It covers both little-endian ARM cores
// in this repository: the Nintendo DS's ARM9/ARM7 (ARMv5TE, the default) and the
// Nintendo 3DS's ARM11 (ARMv6K, selected with -v6). The address space is 32-bit;
// you tell disarm the CPU address the first byte maps to and the instruction set
// to decode in.
//
//	disarm [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] [-thumb] [-v6] file.bin
//
// -base is the CPU address the first byte (after -skip file bytes are dropped) maps
// to; -start/-end select an absolute address sub-range (default: the whole file).
//
// A linear disassembler decodes at one fixed width the whole way through; it cannot
// follow a BX that switches between ARM and Thumb at run time, so decode each state's
// region separately (-thumb selects Thumb) — or use codetracearm, which tracks the
// state changes for you. All addresses are hex; -skip is decimal.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/arm"
)

func hx(s string) (int, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return int(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "disarm: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	baseF := flag.String("base", "0", "CPU address the first byte (after -skip) maps to (hex)")
	skipF := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	startF := flag.String("start", "", "start address (hex), default: base")
	endF := flag.String("end", "", "end address (hex, exclusive), default: end of file")
	thumb := flag.Bool("thumb", false, "decode as Thumb (16-bit) instead of ARM (32-bit)")
	v6 := flag.Bool("v6", false, "decode for ARMv6K (Nintendo 3DS ARM11); default is ARMv5TE (DS)")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: disarm [-base A] [-skip N] [-start A] [-end A] [-thumb] file.bin")
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die("%v", err)
	}
	if *skipF < 0 || *skipF > len(raw) {
		die("bad -skip (file is %d bytes)", len(raw))
	}
	code := raw[*skipF:]
	base, err := hx(*baseF)
	if err != nil {
		die("bad -base")
	}
	start, end := base, base+len(code)
	if *startF != "" {
		if start, err = hx(*startF); err != nil {
			die("bad -start")
		}
	}
	if *endF != "" {
		if end, err = hx(*endF); err != nil {
			die("bad -end")
		}
	}
	if start < base || start >= base+len(code) || end > base+len(code) || end <= start {
		die("range $%X-$%X outside loaded blob ($%X-$%X)", start, end, base, base+len(code))
	}
	variant := arm.V5TE
	if *v6 {
		variant = arm.V6K
	}
	for _, l := range arm.DisassembleVariant(code[start-base:end-base], uint32(start), *thumb, variant) {
		fmt.Println(l)
	}
}
