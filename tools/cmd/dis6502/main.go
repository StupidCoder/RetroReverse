// dis6502 disassembles a C64 .prg file (2-byte load address + data).
//
// Usage: dis6502 [-base addr] [-skip n] [-start addr] [-end addr] file.prg
//
// With no -base the load address is taken from the .prg's 2-byte header (read at
// the -skip offset, default 0). Give -base to override it and treat the file as a
// raw blob whose first byte (after -skip) maps to that address. -start/-end select
// an absolute address sub-range.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/mos6502"
)

func parseAddr(s string) (uint16, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 16)
	return uint16(v), err
}

func main() {
	baseF := flag.String("base", "", "load address (hex); default: from the .prg 2-byte header")
	skipF := flag.Int("skip", 0, "leading file bytes to drop before -base maps (default 0)")
	startF := flag.String("start", "", "start address (hex), default: load address")
	endF := flag.String("end", "", "end address (hex, exclusive), default: end of file")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: dis6502 [-base addr] [-skip n] [-start addr] [-end addr] file.prg")
		os.Exit(2)
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *skipF < 0 || *skipF > len(raw) {
		fmt.Fprintln(os.Stderr, "dis6502: -skip out of range")
		os.Exit(1)
	}
	var load uint16
	var code []byte
	if *baseF == "" {
		// Auto: the .prg's 2-byte little-endian load address at the -skip offset.
		if len(raw) < *skipF+3 {
			fmt.Fprintln(os.Stderr, "dis6502: file too short")
			os.Exit(1)
		}
		load = uint16(raw[*skipF]) | uint16(raw[*skipF+1])<<8
		code = raw[*skipF+2:]
	} else {
		if load, err = parseAddr(*baseF); err != nil {
			fmt.Fprintln(os.Stderr, "bad -base:", err)
			os.Exit(2)
		}
		code = raw[*skipF:]
	}
	start, end := load, load+uint16(len(code))
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
	if start < load || int(start) >= int(load)+len(code) || int(end) > int(load)+len(code) || end <= start {
		fmt.Fprintf(os.Stderr, "dis6502: range $%04X-$%04X outside file ($%04X-$%04X)\n",
			start, end, load, int(load)+len(code)-1)
		os.Exit(1)
	}
	for _, l := range mos6502.Disassemble(code[start-load:end-load], start) {
		fmt.Println(l)
	}
}
