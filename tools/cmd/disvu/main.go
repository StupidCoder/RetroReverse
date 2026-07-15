// disvu linearly disassembles PS2 Vector Unit microcode from a raw binary — a dump of
// VU micro memory (bootoracle -vu1micro) or an MPG payload. Each 64-bit word is an
// instruction pair: the upper half runs on the floating-point FMACs, the lower half on
// the integer/branch pipe; both columns are printed side by side with the pair's I and
// E flags. All numbers are hex.
//
//	disvu [-base ADDR] [-skip N] [-zeros] file.bin
//
// -base is the micro-memory byte address the first decoded byte maps to (branch targets
// print absolute against it); -zeros keeps the all-zero pairs a dump is padded with.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/vu"
)

func hx(s string) (int, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return int(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "disvu: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	baseF := flag.String("base", "0", "micro-memory byte address the first byte (after -skip) maps to (hex)")
	skipF := flag.Int("skip", 0, "leading file bytes to drop")
	zerosF := flag.Bool("zeros", false, "print all-zero pairs too (a dump's padding)")
	flag.Parse()

	if flag.NArg() != 1 {
		die("usage: disvu [-base ADDR] [-skip N] [-zeros] file.bin")
	}
	base, err := hx(*baseF)
	if err != nil {
		die("bad -base %q", *baseF)
	}
	b, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die("%v", err)
	}
	if *skipF > 0 && *skipF <= len(b) {
		b = b[*skipF:]
	}

	for a := 0; a+8 <= len(b); a += 8 {
		raw := binary.LittleEndian.Uint64(b[a:])
		if raw == 0 && !*zerosF {
			continue
		}
		in := vu.Decode(raw, uint32(base+a))
		flags := ""
		if in.E {
			flags += "[E]"
		}
		if in.I {
			flags += "[I]"
		}
		fmt.Printf("%04X  %-44s %-36s %s\n", base+a, in.Upper, in.Lower, flags)
	}
}
