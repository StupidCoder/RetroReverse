// go32dump inspects a DJGPP go32/COFF DOS executable (e.g. Quake's quake.exe):
// it prints the MZ-stub/COFF layout and sections, then linearly disassembles the
// program from its entry point in 32-bit mode. A feature census (x87 escapes,
// 0x0F two-byte opcodes, INT 31h/DPMI call sites, and any bytes the decoder does
// not recognise) scopes the work needed to actually execute the image: it shows
// how soon the FPU appears and which DPMI services the runtime reaches for.
//
// This is analysis-only tooling — it does not execute the program.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/cpu/x86"
	"retroreverse.com/tools/platform/dos"
)

func main() {
	var (
		path   = flag.String("image", "", "path to the go32/COFF executable (e.g. quake.exe)")
		count  = flag.Int("n", 80, "number of instructions to disassemble from the entry point")
		census = flag.Int("census", 0, "if >0, disassemble this many bytes of .text for the feature census (default: whole .text)")
	)
	flag.Parse()
	if *path == "" {
		fmt.Fprintln(os.Stderr, "usage: go32dump -image quake.exe [-n 80] [-census N]")
		os.Exit(2)
	}

	data, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	c, err := dos.ParseGo32COFF(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}

	fmt.Printf("go32/COFF image: %s (%d bytes)\n", *path, len(data))
	fmt.Printf("  MZ stub ends / COFF begins at file offset %#x\n", c.StubEnd)
	fmt.Printf("  sections=%d  filehdr flags=%#04x\n", c.NSections, c.Flags)
	fmt.Printf("  entry=%#08x  text_start=%#08x  data_start=%#08x\n", c.Entry, c.TextStart, c.DataStart)
	fmt.Printf("  tsize=%#x dsize=%#x bsize=%#x\n\n", c.TextSize, c.DataSize, c.BSSSize)

	fmt.Printf("  %-10s %-10s %-10s %-10s %s\n", "SECTION", "VADDR", "SIZE", "FILEOFF", "KIND")
	for _, s := range c.Sections {
		kind := "data"
		switch {
		case s.IsText():
			kind = "text"
		case s.IsBSS():
			kind = "bss"
		}
		fmt.Printf("  %-10s %#08x  %#08x  %#08x  %s\n", s.Name, s.VAddr, s.Size, s.FileOff, kind)
	}
	fmt.Println()

	text := c.TextSection()
	if text == nil {
		fmt.Fprintln(os.Stderr, "no executable section found")
		os.Exit(1)
	}

	// Disassemble from the entry point.
	entryOff := int(c.Entry - text.VAddr)
	if entryOff < 0 || entryOff >= len(text.Data) {
		fmt.Fprintf(os.Stderr, "entry %#x not inside %s [%#x,%#x)\n", c.Entry, text.Name, text.VAddr, text.VAddr+text.Size)
		os.Exit(1)
	}
	fmt.Printf("=== disassembly from entry %#08x (%s+%#x), 32-bit ===\n", c.Entry, text.Name, entryOff)
	pc := entryOff
	for i := 0; i < *count && pc < len(text.Data); i++ {
		in := x86.Decode32(text.Data[pc:], text.VAddr+uint32(pc))
		fmt.Printf("  %08x: %s\n", text.VAddr+uint32(pc), in.Text)
		if in.Len <= 0 {
			break
		}
		pc += in.Len
	}
	fmt.Println()

	// Feature census over the whole (or -census bytes of) .text section.
	limit := len(text.Data)
	if *census > 0 && *census < limit {
		limit = *census
	}
	var total, fpu, twobyte, unknown, ints int
	dpmi := map[byte]int{} // INT n histogram
	mnem := map[string]int{}
	prevWasIntImm := byte(0)
	pc = 0
	for pc < limit {
		in := x86.Decode32(text.Data[pc:], text.VAddr+uint32(pc))
		total++
		mnem[in.Mnem]++
		b := text.Data[pc]
		// account for a leading prefix byte so we classify the real opcode
		op, isFPU := classify(text.Data[pc:])
		if isFPU {
			fpu++
		}
		if op == 0x0F {
			twobyte++
		}
		if in.Mnem == ".byte" {
			unknown++
		}
		if in.Mnem == "INT" {
			ints++
			// the immediate follows the 0xCD opcode
			if pc+1 < len(text.Data) && b == 0xCD {
				dpmi[text.Data[pc+1]]++
			}
		}
		_ = prevWasIntImm
		if in.Len <= 0 {
			pc++
			continue
		}
		pc += in.Len
	}

	fmt.Printf("=== feature census over %d bytes of %s (%d instructions) ===\n", limit, text.Name, total)
	fmt.Printf("  x87 FPU escapes (D8-DF): %d\n", fpu)
	fmt.Printf("  0x0F two-byte opcodes:   %d\n", twobyte)
	fmt.Printf("  software INT sites:      %d\n", ints)
	fmt.Printf("  unrecognised (.byte):    %d\n", unknown)
	if len(dpmi) > 0 {
		fmt.Printf("  INT vectors used:")
		var ns []int
		for n := range dpmi {
			ns = append(ns, int(n))
		}
		sort.Ints(ns)
		for _, n := range ns {
			fmt.Printf(" %02Xh(%d)", n, dpmi[byte(n)])
		}
		fmt.Println()
	}

	// Top mnemonics by frequency.
	type mc struct {
		m string
		n int
	}
	var top []mc
	for m, n := range mnem {
		top = append(top, mc{m, n})
	}
	sort.Slice(top, func(i, j int) bool { return top[i].n > top[j].n })
	fmt.Printf("  top mnemonics:")
	for i := 0; i < len(top) && i < 20; i++ {
		fmt.Printf(" %s(%d)", top[i].m, top[i].n)
	}
	fmt.Println()
}

// classify returns the primary opcode byte of an instruction (skipping prefix
// bytes) and whether it is an x87 FPU escape (0xD8-0xDF).
func classify(code []byte) (op byte, isFPU bool) {
	i := 0
	for i < len(code) {
		b := code[i]
		switch b {
		case 0x26, 0x2E, 0x36, 0x3E, 0x64, 0x65, 0x66, 0x67, 0xF0, 0xF2, 0xF3:
			i++
			continue
		}
		break
	}
	if i >= len(code) {
		return 0, false
	}
	op = code[i]
	return op, op >= 0xD8 && op <= 0xDF
}
