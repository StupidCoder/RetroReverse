// disr5900 is a linear disassembler for Emotion Engine code — the CPU of the
// PlayStation 2 — the counterpart of disr4300 / dismips / disallegrex.
//
// It decodes every 4 bytes in the selected range as an instruction, making no
// attempt to tell code from data; use codetracer5900 for that.
//
// A PS2 boot ELF is recognised and loaded at the address it asks for, so -base is
// usually unnecessary: the file already says where it lives. Where the ELF ships a
// symbol table, function boundaries are labelled in the listing. Anything that is
// not an ELF is loaded flat at -base.
//
// Usage:
//
//	disr5900 [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] SCUS_971.24
//
// All addresses are hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/r5900"
	"retroreverse.com/tools/platform/ps2"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex; ignored for an ELF)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	start := flag.String("start", "", "first address to disassemble (hex, default the load address)")
	end := flag.String("end", "", "last address to disassemble (hex, default end of image)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disr5900 [-base A] [-skip N] [-start A] [-end A] SCUS_971.24")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *start, *end); err != nil {
		fmt.Fprintln(os.Stderr, "disr5900:", err)
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

	var (
		mem  []byte
		base uint32
		exe  *ps2.Executable
	)
	if e, err := ps2.LoadELF(raw); err == nil {
		exe = e
		base, mem = e.Flat()
		fmt.Fprintf(os.Stderr, "disr5900: ELF loaded at $%08X (%d bytes), entry $%08X, %d symbols\n",
			base, len(mem), e.Entry, len(e.Symbols))
	} else {
		if base, err = hx(baseS); err != nil {
			return fmt.Errorf("bad -base %q", baseS)
		}
		if skip < 0 || skip > len(raw) {
			return fmt.Errorf("bad -skip %d", skip)
		}
		mem = raw[skip:]
	}

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

	for a := lo &^ 3; a <= hi; a += 4 {
		// Where the ELF names this address, head the instruction with the label. It
		// is the difference between reading a listing and decoding one.
		if exe != nil {
			if name, off, ok := exe.Lookup(a); ok && off == 0 {
				fmt.Fprintf(w, "\n; ==== %s  $%08X ====\n", name, a)
			}
		}
		o := int(a - base)
		in := r5900.Decode(mem[o:], a)
		fmt.Fprintf(w, "%08X  %02X %02X %02X %02X  %s\n",
			a, mem[o], mem[o+1], mem[o+2], mem[o+3], in.Text)
	}
	return nil
}
