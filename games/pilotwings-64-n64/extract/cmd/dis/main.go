// dis disassembles a range of the game's code out of a booted machine.
//
// The game's code lives in the cartridge, but its overlays are relocated into
// RDRAM, so an address means nothing without a machine to read it from. A boot
// (or a savestate cut at the moment of interest) is the only way to see an
// overlay's instructions at the address the call log reports.
//
// Usage:
//
//	dis -image ROM [-loadstate FILE] -at ADDR [-n COUNT] [-steps N]
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n64"
)

func main() {
	image := flag.String("image", "", "cartridge image (.z64/.v64/.n64)")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before disassembling")
	at := flag.String("at", "", "start address (hex)")
	n := flag.Int("n", 64, "number of instructions")
	dump := flag.Bool("dump", false, "print words instead of instructions")
	flag.Parse()

	if *image == "" || *at == "" {
		fmt.Fprintln(os.Stderr, "dis: -image and -at are required")
		os.Exit(2)
	}
	addr, err := strconv.ParseUint(strings.TrimPrefix(*at, "0x"), 16, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dis: bad -at:", err)
		os.Exit(2)
	}

	rom, err := n64.Load(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dis:", err)
		os.Exit(1)
	}
	m := n64.NewMachine(rom)
	if err := m.Boot(rom, n64.DefaultBoot()); err != nil {
		fmt.Fprintln(os.Stderr, "dis:", err)
		os.Exit(1)
	}
	if *loadState != "" {
		if err := m.LoadState(*loadState); err != nil {
			fmt.Fprintln(os.Stderr, "dis: -loadstate:", err)
			os.Exit(1)
		}
	}
	for i := 0; i < *n; i++ {
		a := addr + uint64(i)*4
		if *dump {
			fmt.Printf("%08X  %08X\n", a, m.Read32(uint32(a)))
			continue
		}
		fmt.Printf("%08X  %s\n", a, m.DisasmAt(a))
	}
}
