// disz80 linearly disassembles Z80 code from a Sega Master System / Game Gear
// cartridge. Because the Z80 address space is 16-bit but the ROM is larger and paged
// through the mapper, there are two ways to say which bytes to decode:
//
// Flat mode — disassemble a raw file slice mapped at a Z80 address:
//
//	disz80 [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] rom.gg
//
// -base is the Z80 address the first byte (after -skip file bytes) maps to;
// -start/-end select an absolute address sub-range (default: the whole file).
//
// Bank mode — disassemble a Z80 address range with a given slot configuration, so
// banked code (and its cross-references) decodes against the right banks:
//
//	disz80 -slots b0,b1,b2 -start ADDR [-end ADDR] rom.gg
//
// e.g. the bank-3 dispatcher (bank 3 paged into slot 1):
//
//	disz80 -slots 0,3,2 -start 0x4000 -end 0x4080 rom.gg
//
// All addresses are hex; -skip is decimal. In bank mode -end defaults to -start + $80.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/gamegear"
	"retroreverse.com/tools/cpu/z80"
)

func hx(s string) (int, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 32)
	return int(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "disz80: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	baseF := flag.String("base", "0", "flat mode: Z80 address the first byte (after -skip) maps to (hex)")
	skipF := flag.Int("skip", 0, "flat mode: leading file bytes to drop before -base maps")
	slotsF := flag.String("slots", "", "bank mode: ROM banks in slots 0,1,2 (hex, e.g. 0,3,2)")
	startF := flag.String("start", "", "start address (hex); flat: default base, bank: required")
	endF := flag.String("end", "", "end address, exclusive (hex); flat: default end of file, bank: default start+$80")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: disz80 [-base A] [-skip N] [-start A] [-end A] | -slots b0,b1,b2 -start A [-end A] rom")
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die("%v", err)
	}

	if *slotsF != "" {
		bankMode(raw, *slotsF, *startF, *endF)
		return
	}

	// Flat mode.
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
	for _, l := range z80.Disassemble(code[start-base:end-base], uint16(start)) {
		fmt.Println(l)
	}
}

func bankMode(rom []byte, slotsArg, startArg, endArg string) {
	parts := strings.Split(slotsArg, ",")
	if len(parts) != 3 {
		die("-slots wants three banks, e.g. 0,3,2")
	}
	var slots [3]int
	for i, p := range parts {
		b, err := hx(strings.TrimSpace(p))
		if err != nil || b < 0 {
			die("bad bank in -slots: %q", p)
		}
		slots[i] = b
	}
	if startArg == "" {
		die("bank mode needs -start")
	}
	start, err := hx(startArg)
	if err != nil || start < 0 || start >= 0xC000 {
		die("bad -start (must be $0000-$BFFF)")
	}
	end := start + 0x80
	if endArg != "" {
		if end, err = hx(endArg); err != nil {
			die("bad -end")
		}
	}
	if end > 0xC000 {
		end = 0xC000
	}
	if end <= start {
		die("-end must be greater than -start")
	}
	view := gamegear.BankView(rom, slots)
	for _, l := range z80.Disassemble(view[start:end], uint16(start)) {
		fmt.Println(l)
	}
}
