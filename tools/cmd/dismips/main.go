// dismips linearly disassembles MIPS R3000 (PlayStation) code from a raw binary
// — an extracted PS-X EXE text image, a memory dump, or a slice of one. You tell
// it the CPU address the first decoded byte maps to; all numbers are hex.
//
//	dismips [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] file.bin
//
// -base is the CPU address the first byte (after -skip file bytes are dropped) maps
// to; -start/-end select an absolute address sub-range (default: the whole file).
//
// A linear disassembler decodes straight through and cannot follow indirect
// jumps or separate code from data — use codetracemips for that. To disassemble
// a PS-X EXE, skip its 0x800 header with -skip 2048 and set -base to t_addr.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/mips"
)

func hx(s string) (int, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return int(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "dismips: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	baseF := flag.String("base", "0", "CPU address the first byte (after -skip) maps to (hex)")
	skipF := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	startF := flag.String("start", "", "start address (hex), default: base")
	endF := flag.String("end", "", "end address (hex, exclusive), default: end of file")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: dismips [-base A] [-skip N] [-start A] [-end A] file.bin")
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
	for _, l := range mips.Disassemble(code[start-base:end-base], uint32(start)) {
		fmt.Println(l)
	}
}
