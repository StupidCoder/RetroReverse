// picadump decodes a PICA200 command buffer (as captured by the boot oracle's
// -gxdump flag) into its register-write stream. A command list is a sequence of
// GPU register writes; this tool is the instrument-first companion to the GPU
// interpreter — it shows what register traffic a game actually generates, so the
// GPU implements what is used rather than what is imagined.
//
// Usage:
//
//	picadump cmdlist_00.bin              annotated register-write listing
//	picadump -hist cmdlist_00.bin        histogram of writes per register
//	picadump -reg 0x22E cmdlist_00.bin   only writes to one register
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	hist := flag.Bool("hist", false, "print a histogram of writes per register instead of the listing")
	shader := flag.Bool("shader", false, "extract the vertex-shader upload from the list and disassemble it")
	regFilter := flag.String("reg", "", "only show writes to this register id (hex with 0x, else decimal)")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "picadump: need a command-buffer file")
		os.Exit(2)
	}

	var only int64 = -1
	if *regFilter != "" {
		v, err := parseNum(*regFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "picadump: bad -reg %q: %v\n", *regFilter, err)
			os.Exit(2)
		}
		only = int64(v)
	}

	for _, path := range flag.Args() {
		buf, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "picadump:", err)
			os.Exit(1)
		}
		ws, derr := n3ds.DecodePICA(buf)
		// Entries and register writes are not the same number, and confusing them is
		// how a command list comes to look absurd. One entry can carry a burst of
		// hundreds of words into a FIFO — a 396-word vertex-shader upload is ONE entry
		// and 396 register writes. Report both, plus the padding, which is written to
		// register 0 with an empty byte-enable mask and does nothing at all.
		entries, padding := 0, 0
		for _, w := range ws {
			if !w.Burst {
				entries++
			}
			if w.Mask == 0 {
				padding++
			}
		}
		fmt.Printf("%s: %d bytes, %d entries, %d register writes (%d are FIFO burst words, %d are padding)\n",
			path, len(buf), entries, len(ws), len(ws)-entries, padding)
		switch {
		case *shader:
			printShader(ws)
		case *hist:
			printHist(ws)
		default:
			printListing(ws, only)
		}
		if derr != nil {
			fmt.Fprintln(os.Stderr, "picadump:", derr)
			os.Exit(1)
		}
	}
}

func printListing(ws []n3ds.PICAWrite, only int64) {
	for _, w := range ws {
		if only >= 0 && int64(w.Reg) != only {
			continue
		}
		burst := " "
		if w.Burst {
			burst = "+"
		}
		mask := ""
		if w.Mask != 0xF {
			mask = fmt.Sprintf(" mask=%04b", w.Mask)
		}
		fmt.Printf("  %06X: %s reg 0x%03X = %08X%s  %s\n", w.Off, burst, w.Reg, w.Value, mask, groupOf(w.Reg))
	}
}

func printHist(ws []n3ds.PICAWrite) {
	counts := map[uint16]int{}
	for _, w := range ws {
		counts[w.Reg]++
	}
	regs := make([]uint16, 0, len(counts))
	for r := range counts {
		regs = append(regs, r)
	}
	sort.Slice(regs, func(i, j int) bool { return regs[i] < regs[j] })
	for _, r := range regs {
		fmt.Printf("  reg 0x%03X  ×%-6d %s\n", r, counts[r], groupOf(r))
	}
	fmt.Printf("  %d distinct registers\n", len(regs))
}

// printShader replays the list's shader-engine upload registers (code at
// 0x2CB/0x2CC-0x2D3, operand descriptors at 0x2D5/0x2D6-0x2DD, the entry point
// at 0x2BA) and disassembles the program — the check that the instruction
// decode is right: a wrong field layout turns the game's own shader into
// gibberish.
func printShader(ws []n3ds.PICAWrite) {
	var code [4096]uint32
	var opdesc [128]uint32
	codeIdx, opdIdx, entry, top := 0, 0, 0, 0
	for _, w := range ws {
		switch {
		case w.Reg == 0x2CB:
			codeIdx = int(w.Value & 0xFFF)
		case w.Reg >= 0x2CC && w.Reg < 0x2D4:
			if codeIdx < len(code) {
				code[codeIdx] = w.Value
				codeIdx++
				if codeIdx > top {
					top = codeIdx
				}
			}
		case w.Reg == 0x2D5:
			opdIdx = int(w.Value & 0x7F)
		case w.Reg >= 0x2D6 && w.Reg < 0x2DE:
			if opdIdx < len(opdesc) {
				opdesc[opdIdx] = w.Value
				opdIdx++
			}
		case w.Reg == 0x2BA:
			entry = int(w.Value & 0xFFF)
		}
	}
	if top == 0 {
		fmt.Println("  no shader code uploaded in this list")
		return
	}
	fmt.Printf("  %d code words uploaded, entry 0x%03X\n", top, entry)
	for i := 0; i < top; i++ {
		marker := "  "
		if i == entry {
			marker = "->"
		}
		fmt.Printf("  %s %03X: %08X  %s\n", marker, i, code[i], n3ds.ShaderDisasm(code[i], &opdesc))
	}
}

// groupOf mirrors the platform package's register grouping for annotation.
func groupOf(reg uint16) string { return n3ds.PICARegGroup(reg) }

func parseNum(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}
