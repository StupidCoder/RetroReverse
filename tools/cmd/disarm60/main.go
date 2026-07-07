// disarm60 is a linear ARM60 disassembler (big-endian, ARMv3) for 3DO code.
//
//	disarm60 -base 0x0 -off 0x80 -len 0x200 LaunchMe
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/arm60"
)

func main() {
	base := flag.Uint64("base", 0, "load address of the first byte")
	off := flag.Uint64("off", 0, "byte offset into the file to start at")
	length := flag.Uint64("len", 0, "number of bytes to disassemble (0 = to EOF)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disarm60 [-base A] [-off O] [-len N] file")
		os.Exit(2)
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "disarm60:", err)
		os.Exit(1)
	}
	o := int(*off)
	if o > len(raw) {
		o = len(raw)
	}
	end := len(raw)
	if *length > 0 && o+int(*length) < end {
		end = o + int(*length)
	}
	for _, line := range arm60.Disassemble(raw[o:end], uint32(*base)+uint32(o)) {
		fmt.Println(line)
	}
}
