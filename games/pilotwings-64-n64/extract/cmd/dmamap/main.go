// dmamap traces RDRAM asset regions back to cartridge offsets.
//
// It reads a PI-DMA log from a scratch boot (bootoracle -dmalog), an RDRAM
// snapshot from the field being studied (rdpdbg -snapram), and the ROM, and
// answers: which cart offset delivered the bytes at this RDRAM region, and are
// they still byte-identical at snapshot time? A region no DMA covers, or one
// whose bytes differ from every DMA that touched it, was built or unpacked at
// runtime — a finding about the loader, not a mapping.
//
// Usage:
//
//	dmamap -image ROM -dmalog FILE -ram SNAP.bin -region ADDR:LEN [-region ...]
//	dmamap -image ROM -dmalog FILE -ram SNAP.bin -scan
//
// -scan sweeps the whole snapshot: for every PI DMA it reports whether the
// snapshot still holds the ROM's bytes (intact / partial / overwritten).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n64"
)

type dma struct {
	field       int
	dram, cart  uint32
	length      uint32
}

func main() {
	image := flag.String("image", "", "cartridge image")
	dmaLog := flag.String("dmalog", "", "DMA log from bootoracle -dmalog")
	ramFile := flag.String("ram", "", "RDRAM snapshot to verify against")
	scan := flag.Bool("scan", false, "report snapshot intactness for every PI DMA")
	var regions multiFlag
	flag.Var(&regions, "region", "RDRAM region ADDR:LEN (hex); repeatable")
	flag.Parse()

	rom, err := n64.Load(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ram, err := os.ReadFile(*ramFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dmas, err := readLog(*dmaLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%d PI DMAs, ROM %d bytes, RAM %d bytes\n", len(dmas), len(rom.Data), len(ram))

	// cartBytes reads the ROM as the PI sees it: cart addresses are
	// 0x10000000-based.
	cartByte := func(addr uint32) (byte, bool) {
		off := int(addr) - 0x10000000
		if off < 0 || off >= len(rom.Data) {
			return 0, false
		}
		return rom.Data[off], true
	}

	// match counts how many of the DMA's bytes the snapshot still holds.
	match := func(d dma) (same, total int) {
		for i := uint32(0); i < d.length; i++ {
			b, ok := cartByte(d.cart + i)
			if !ok || int(d.dram+i) >= len(ram) {
				continue
			}
			total++
			if ram[d.dram+i] == b {
				same++
			}
		}
		return
	}

	if *scan {
		for _, d := range dmas {
			same, total := match(d)
			state := "OVERWRITTEN"
			switch {
			case total == 0:
			case same == total:
				state = "intact"
			case same > total*9/10:
				state = "mostly"
			case same > total/10:
				state = "partial"
			}
			fmt.Printf("field %4d  ram %06X..%06X  cart %07X  len %6X  %s (%d/%d)\n",
				d.field, d.dram, d.dram+d.length, d.cart&0x0FFFFFFF, d.length, state, same, total)
		}
		return
	}

	for _, r := range regions {
		parts := strings.SplitN(r, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "bad -region %q\n", r)
			continue
		}
		addr, _ := strconv.ParseUint(strings.TrimPrefix(parts[0], "0x"), 16, 32)
		length, _ := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 32)
		lo, hi := uint32(addr), uint32(addr+length)
		fmt.Printf("\n== region %06X..%06X\n", lo, hi)
		covered := make([]bool, length)
		n := 0
		for _, d := range dmas {
			if d.dram >= hi || d.dram+d.length <= lo {
				continue
			}
			n++
			same, total := match(d)
			for i := uint32(0); i < d.length; i++ {
				if a := d.dram + i; a >= lo && a < hi {
					covered[a-lo] = true
				}
			}
			fmt.Printf("  field %4d  ram %06X..%06X  cart %07X  len %6X  snapshot %d/%d identical\n",
				d.field, d.dram, d.dram+d.length, d.cart&0x0FFFFFFF, d.length, same, total)
		}
		nc := 0
		for _, c := range covered {
			if c {
				nc++
			}
		}
		fmt.Printf("  %d DMAs touch the region; %d/%d bytes ever DMA'd\n", n, nc, len(covered))
	}
}

func readLog(path string) ([]dma, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []dma
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var kind string
		var d dma
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if _, err := fmt.Sscanf(line, "%s %d %x %x %x", &kind, &d.field, &d.dram, &d.cart, &d.length); err != nil {
			continue
		}
		if kind == "pi" {
			out = append(out, d)
		}
	}
	return out, sc.Err()
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}
