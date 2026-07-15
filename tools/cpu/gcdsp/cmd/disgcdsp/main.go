// Command disgcdsp disassembles a GameCube-DSP microcode image — a raw stream of big-endian
// 16-bit words, as the game DMAs it into the DSP's instruction RAM. It is the same decoder the
// core runs, used here to read a ucode off disk and confirm it decodes to a coherent program.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/gcdsp"
)

func main() {
	in := flag.String("in", "", "ucode image (raw big-endian 16-bit words)")
	start := flag.Int("start", 0, "first word address to disassemble")
	count := flag.Int("count", 0, "words to disassemble (0 = whole image)")
	validate := flag.Bool("validate", false, "check every branch target is instruction-aligned")
	flag.Parse()

	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: disgcdsp -in ucode.bin [-start W] [-count N]")
		os.Exit(2)
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	words := make([]uint16, len(raw)/2)
	for i := range words {
		words[i] = binary.BigEndian.Uint16(raw[i*2:])
	}
	read := func(a uint16) uint16 {
		if int(a) < len(words) {
			return words[a]
		}
		return 0
	}
	n := uint16(len(words))
	if *count > 0 {
		n = uint16(*start + *count)
	}
	if *validate {
		mis, boundaries, overrun := gcdsp.DisasmValidate(read, uint16(len(words)))
		fmt.Printf("; %s: %d words, %d instructions decoded, overrun=%v\n", *in, len(words), boundaries, overrun)
		if len(mis) == 0 {
			fmt.Println("; all branch targets land on instruction boundaries — decode is self-consistent")
		} else {
			fmt.Printf("; %d branches with misaligned targets:\n", len(mis))
			for _, w := range mis {
				t, _ := gcdsp.Disasm(read, w)
				fmt.Printf(";   0x%04X: %s\n", w, t)
			}
		}
		return
	}

	fmt.Printf("; %s: %d words\n", *in, len(words))
	fmt.Print(gcdsp.DisasmRange(read, uint16(*start), n-uint16(*start)))
}
