// disrsp is a linear disassembler for Reality Signal Processor microcode — the
// vector coprocessor inside the Nintendo 64's RCP.
//
// The microcode a game runs on the RSP is data on its cartridge: the display-list
// interpreter, the audio mixer, and whatever else the developer wrote. Reversing
// the render pipeline means reading it, which is what this is for. Feed it a
// dump of instruction memory, or the region of the ROM a task DMAs into IMEM.
//
// Usage:
//
//	disrsp [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] imem.bin
//
// Addresses are IMEM offsets, 12 bits wide, and hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/rsp"
)

func main() {
	base := flag.String("base", "0", "IMEM address the image is loaded at (hex)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	start := flag.String("start", "", "first address to disassemble (hex, default -base)")
	end := flag.String("end", "", "last address to disassemble (hex, default end of image)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disrsp [-base A] [-skip N] [-start A] [-end A] imem.bin")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *start, *end); err != nil {
		fmt.Fprintln(os.Stderr, "disrsp:", err)
		os.Exit(1)
	}
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

func run(path, baseS string, skip int, startS, endS string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	base, err := hx(baseS)
	if err != nil {
		return fmt.Errorf("bad -base %q", baseS)
	}
	if skip < 0 || skip > len(raw) {
		return fmt.Errorf("bad -skip %d", skip)
	}
	mem := raw[skip:]

	lo, hi := base, base+uint32(len(mem))-1
	if startS != "" {
		if lo, err = hx(startS); err != nil {
			return fmt.Errorf("bad -start %q", startS)
		}
	}
	if endS != "" {
		if hi, err = hx(endS); err != nil {
			return fmt.Errorf("bad -end %q", endS)
		}
	}
	if lo < base || hi < lo || int(hi-base) >= len(mem) {
		return fmt.Errorf("range $%03X-$%03X lies outside the image at $%03X (%d bytes)", lo, hi, base, len(mem))
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, line := range rsp.Disassemble(mem[lo-base:hi-base+1], lo) {
		fmt.Fprintln(w, line)
	}
	return nil
}
