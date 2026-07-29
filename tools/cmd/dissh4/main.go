// dissh4 is a linear disassembler for Hitachi SH-4 code — the CPU of the Sega
// Dreamcast — the counterpart of disr4300 / disgekko.
//
// It decodes every 2 bytes in the selected range as an instruction, making no
// attempt to tell code from the literal pools SH compilers weave between
// functions; use codetracesh4 for that.
//
// Usage:
//
//	dissh4 [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] 1ST_READ.BIN
//
// The image is loaded flat at -base (default 0); -skip drops that many leading
// file bytes. All addresses are hex. A Dreamcast main binary loads at
// 8C010000.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/sh4"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	start := flag.String("start", "", "first address to disassemble (hex, default -base)")
	end := flag.String("end", "", "last address to disassemble (hex, default end of image)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: dissh4 [-base A] [-skip N] [-start A] [-end A] 1ST_READ.BIN")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *start, *end); err != nil {
		fmt.Fprintln(os.Stderr, "dissh4:", err)
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
		return fmt.Errorf("range $%08X-$%08X lies outside the image at $%08X (%d bytes)", lo, hi, base, len(mem))
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, line := range sh4.Disassemble(mem[lo-base:hi-base+1], lo) {
		fmt.Fprintln(w, line)
	}
	return nil
}
