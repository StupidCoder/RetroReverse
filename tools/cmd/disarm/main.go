// disarm linearly disassembles Nintendo DS ARM code (ARM9 / ARM7) from a raw binary
// — an extracted .bin, a memory dump, or a slice of a cartridge. The DS address
// space is 32-bit; you tell disarm the CPU address the first byte maps to and the
// instruction set to decode in.
//
//	disarm [-off FILEOFF] [-len N] [-base ADDR] [-thumb] file.bin
//
// A linear disassembler decodes at one fixed width the whole way through; it cannot
// follow a BX that switches between ARM and Thumb at run time, so decode each state's
// region separately (-thumb selects Thumb) — or use codetracearm, which tracks the
// state changes for you. All numbers are hex.
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
	offF := flag.String("off", "0", "file offset to start at (hex)")
	lenF := flag.String("len", "", "number of bytes (hex, default: to end of file)")
	baseF := flag.String("base", "0", "CPU address the first decoded byte maps to (hex)")
	thumb := flag.Bool("thumb", false, "decode as Thumb (16-bit) instead of ARM (32-bit)")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: disarm [-off F -len N -base A] [-thumb] file.bin")
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
	for _, l := range arm.Disassemble(raw[off:off+n], uint32(base), *thumb) {
		fmt.Println(l)
	}
}
