// disr4300 is a linear disassembler for NEC VR4300 code — the CPU of the
// Nintendo 64 — the counterpart of dismips / disarm60.
//
// It decodes every 4 bytes in the selected range as an instruction, making no
// attempt to tell code from data; use codetracer4300 for that.
//
// A cartridge image is accepted in any of the three byte orders (z64, v64, n64)
// and normalised to the cartridge's native big-endian before decoding, so a
// file whose extension lies about its contents still disassembles.
//
// Usage:
//
//	disr4300 [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] image.z64
//
// The image is loaded flat at -base (default 0); -skip drops that many leading
// file bytes. All addresses are hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/r4300"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	start := flag.String("start", "", "first address to disassemble (hex, default -base)")
	end := flag.String("end", "", "last address to disassemble (hex, default end of image)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disr4300 [-base A] [-skip N] [-start A] [-end A] image.z64")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *start, *end); err != nil {
		fmt.Fprintln(os.Stderr, "disr4300:", err)
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
	if order, ok := n64.Normalize(raw); ok && order != n64.OrderZ64 {
		fmt.Fprintf(os.Stderr, "disr4300: image stored as %s; normalised to big-endian\n", order)
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
	for _, line := range r4300.Disassemble(mem[lo-base:hi-base+1], lo) {
		fmt.Fprintln(w, line)
	}
	return nil
}
