// dissm83 linearly disassembles Sharp LR35902 (Game Boy) code from a cartridge ROM.
// The CPU address space is 16-bit; on an MBC1 cart bank 0 is fixed at $0000-$3FFF and
// a selected bank is paged into $4000-$7FFF, so there are two ways to say which bytes
// to decode:
//
// Flat mode — disassemble a raw file slice mapped at a CPU address:
//
//	dissm83 [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] rom.gb
//
// -base is the CPU address the first byte (after -skip file bytes) maps to;
// -start/-end select an absolute address sub-range (default: the whole file).
//
// Bank mode — disassemble a CPU address range with bank 0 fixed and a chosen bank in
// the $4000-$7FFF window, so banked code decodes against the right bank:
//
//	dissm83 -bank N -start ADDR [-end ADDR] rom.gb
//
// e.g. the sound engine the timer ISR calls (in bank 3, at $7FF0):
//
//	dissm83 -bank 3 -start 0x7FF0 -end 0x8000 rom.gb
//
// All addresses are hex; -skip is decimal. In bank mode -end defaults to -start + $80.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/sm83"
)

const bankSize = 0x4000 // 16 KB

func hx(s string) (int, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 32)
	return int(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "dissm83: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	baseF := flag.String("base", "0", "flat mode: CPU address the first byte (after -skip) maps to (hex)")
	skipF := flag.Int("skip", 0, "flat mode: leading file bytes to drop before -base maps")
	bankF := flag.String("bank", "", "bank mode: ROM bank paged into $4000-$7FFF (hex)")
	startF := flag.String("start", "", "start address (hex); flat: default base, bank: required")
	endF := flag.String("end", "", "end address, exclusive (hex); flat: default end of file, bank: default start+$80")
	flag.Parse()
	if flag.NArg() != 1 {
		die("usage: dissm83 [-base A] [-skip N] [-start A] [-end A] | -bank N -start A [-end A] rom")
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die("%v", err)
	}

	if *bankF != "" {
		bankMode(raw, *bankF, *startF, *endF)
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
	for _, l := range sm83.Disassemble(code[start-base:end-base], uint16(start)) {
		fmt.Println(l)
	}
}

// bankMode assembles a 32 KB CPU view (bank 0 at $0000-$3FFF, the chosen bank at
// $4000-$7FFF) and disassembles a [start,end) range of it.
func bankMode(rom []byte, bankArg, startArg, endArg string) {
	bank, err := hx(bankArg)
	if err != nil || bank < 0 || (bank+1)*bankSize > len(rom) {
		die("bad -bank (rom has %d banks)", len(rom)/bankSize)
	}
	if startArg == "" {
		die("bank mode needs -start")
	}
	start, err := hx(startArg)
	if err != nil || start < 0 || start >= 0x8000 {
		die("bad -start (must be $0000-$7FFF)")
	}
	end := start + 0x80
	if endArg != "" {
		if end, err = hx(endArg); err != nil {
			die("bad -end")
		}
	}
	if end > 0x8000 {
		end = 0x8000
	}
	if end <= start {
		die("-end must be greater than -start")
	}
	view := make([]byte, 0x8000)
	copy(view[0:bankSize], rom[0:bankSize])                           // bank 0, fixed
	copy(view[bankSize:0x8000], rom[bank*bankSize:(bank+1)*bankSize]) // selected bank
	for _, l := range sm83.Disassemble(view[start:end], uint16(start)) {
		fmt.Println(l)
	}
}
