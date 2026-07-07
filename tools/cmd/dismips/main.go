// dismips linearly disassembles MIPS R3000 (PlayStation) code from a raw binary
// — an extracted PS-X EXE text image, a memory dump, or a slice of one. You tell
// it the CPU address the first decoded byte maps to; all numbers are hex.
//
//	dismips [-off FILEOFF] [-len N] [-base ADDR] file.bin
//
// A linear disassembler decodes straight through and cannot follow indirect
// jumps or separate code from data — use codetracemips for that. To disassemble
// a PS-X EXE, skip its 0x800 header with -off 800 and set -base to t_addr.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/mips"
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
	offF := flag.String("off", "0", "file offset to start at (hex)")
	lenF := flag.String("len", "", "number of bytes (hex, default: to end of file)")
	baseF := flag.String("base", "0", "CPU address the first decoded byte maps to (hex)")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: dismips [-off F -len N -base A] file.bin")
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die("%v", err)
	}
	off, err := hx(*offF)
	if err != nil || off < 0 || off > len(raw) {
		die("bad -off (file is %d bytes)", len(raw))
	}
	n := len(raw) - off
	if *lenF != "" {
		if n, err = hx(*lenF); err != nil || n < 0 || off+n > len(raw) {
			die("bad -len (file is %d bytes)", len(raw))
		}
	}
	base, err := hx(*baseF)
	if err != nil {
		die("bad -base")
	}
	for _, l := range mips.Disassemble(raw[off:off+n], uint32(base)) {
		fmt.Println(l)
	}
}
