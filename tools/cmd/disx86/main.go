// disx86 linearly disassembles a raw 16-bit real-mode x86 code blob loaded at a
// base address — the x86 counterpart to dis6502/dis68k/disarm. It is a plain
// linear sweep (data between routines will mis-decode); use codetracex86 to
// follow control flow and separate code from data.
//
// Usage: disx86 [-base addr] [-skip n] [-start addr] [-end addr] file
//
// For UW.EXE, disassemble the MZ load module by skipping its 12,800-byte header
// (-skip 0x3200) and giving the module a base of 0.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/x86"
)

func parseAddr(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 32)
	return uint32(v), err
}

func main() {
	baseF := flag.String("base", "0", "load address of the first byte (hex)")
	skipF := flag.Int("skip", 0, "skip this many leading bytes (e.g. past the MZ header)")
	startF := flag.String("start", "", "start address (hex), default: base")
	endF := flag.String("end", "", "end address (hex, exclusive), default: end of file")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disx86 [-base addr] [-skip n] [-start addr] [-end addr] file")
		os.Exit(2)
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "disx86:", err)
		os.Exit(1)
	}
	base, err := parseAddr(*baseF)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad -base:", err)
		os.Exit(2)
	}
	if *skipF < 0 || *skipF > len(raw) {
		fmt.Fprintln(os.Stderr, "disx86: -skip out of range")
		os.Exit(1)
	}
	code := raw[*skipF:]

	start, end := base, base+uint32(len(code))
	if *startF != "" {
		if start, err = parseAddr(*startF); err != nil {
			fmt.Fprintln(os.Stderr, "bad -start:", err)
			os.Exit(2)
		}
	}
	if *endF != "" {
		if end, err = parseAddr(*endF); err != nil {
			fmt.Fprintln(os.Stderr, "bad -end:", err)
			os.Exit(2)
		}
	}
	if start < base || start >= base+uint32(len(code)) || end > base+uint32(len(code)) || end <= start {
		fmt.Fprintf(os.Stderr, "disx86: range $%X-$%X outside loaded blob ($%X-$%X)\n",
			start, end, base, base+uint32(len(code)))
		os.Exit(1)
	}
	for _, l := range x86.Disassemble(code[start-base:end-base], start) {
		fmt.Println(l)
	}
}
