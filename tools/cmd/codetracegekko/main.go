// codetracegekko is a recursive-descent ("code-tracing") disassembler for Gekko code —
// the PowerPC 750 of the Nintendo GameCube — the counterpart of codetracer4300 /
// codetracer5900 / codetracemips / codetracearm.
//
// Starting from one or more entry points it follows every branch, call and jump, marks
// which words are reachable code, and leaves everything else as data. That distinction is
// not academic here: a GameCube executable has its compiler's copyright strings, its
// jump tables and its constant pools sitting *inside* the text segment, and a linear
// disassembler walks straight into them and decodes "Metrowerks Target Resident Kernel
// for PowerPC" as a run of xoris and andi. instructions. The descent stops at the blr
// before the string and never sees it.
//
// Three things about PowerPC shape the descent, and all three differ from MIPS:
//
//   - There is no delay slot. Nothing executes after a taken branch, so a transfer is
//     honoured immediately rather than one instruction later.
//   - An indirect branch can be CONDITIONAL. A `bclr` with a condition may return and may
//     also fall through, so the descent must walk the fall-through path as well — it is
//     classified FlowBranch, not FlowReturn, and the eight-value Flow enum absorbs it.
//   - `bl .+4` is not a call. It is how position-independent code reads its own program
//     counter, and treating it as a call would invent a function at every one of them.
//
// Indirect transfers whose target is not statically known — a `bctr` through a register,
// which in C++ code means a virtual call or a switch — are reported as unresolved. Supply
// their tables with -table, or add discovered targets as further -entry points. That
// feedback loop, seeded from the boot oracle's live trace, is the reverse-engineering
// workflow.
//
// Usage:
//
//	codetracegekko [-base ADDR] [-skip N] [-entry A,B,C] [-table ADDR:N] [-annotate FILE] [-o out] IMAGE
//
// A GameCube disc or a raw .dol loads its own segments and defaults -entry to the
// executable's entry point. All addresses are hex.
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/gekko"
	"retroreverse.com/tools/platform/gc"
)

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex; ignored for a disc or DOL)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex; defaults to the executable's)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N 32-bit pointers); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write the disassembly to this file (default stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetracegekko [-base A] [-skip N] [-entry A,B,C] [-table ADDR:N] [-annotate F] [-o out] IMAGE")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetracegekko:", err)
		os.Exit(1)
	}
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "$"), "0x"), 16, 64)
	return uint32(v), err
}

// traced is what the descent learns.
type traced struct {
	instr    map[uint32]gekko.Inst
	covered  []bool // one per word of the image
	callers  map[uint32]int
	indirect []uint32 // the sites whose target we could not resolve
	stops    []uint32
}

func run(path, baseS string, skip int, entryS string, tables multiFlag, annotate, out string) error {
	base, mem, dol, err := load(path, baseS, skip)
	if err != nil {
		return err
	}

	var seeds []uint32
	if entryS != "" {
		for _, s := range strings.Split(entryS, ",") {
			a, err := hx(s)
			if err != nil {
				return fmt.Errorf("bad -entry %q", s)
			}
			seeds = append(seeds, a)
		}
	} else if dol != nil {
		seeds = append(seeds, dol.Entry)
		fmt.Fprintf(os.Stderr, "codetracegekko: no -entry given; starting at the executable's entry, 0x%08X\n", dol.Entry)
	}
	if len(seeds) == 0 {
		return fmt.Errorf("no entry point: give -entry, or an image that names its own")
	}

	// A jump table is a run of pointers. Seeding one turns a switch statement's
	// unresolved bctr into a set of reachable targets.
	for _, t := range tables {
		parts := strings.SplitN(t, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("bad -table %q, want ADDR:N", t)
		}
		addr, err := hx(parts[0])
		if err != nil {
			return fmt.Errorf("bad -table address %q", parts[0])
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("bad -table count %q", parts[1])
		}
		for i := 0; i < n; i++ {
			o := int(addr-base) + i*4
			if o < 0 || o+4 > len(mem) {
				return fmt.Errorf("-table %s runs outside the image", t)
			}
			seeds = append(seeds, binary.BigEndian.Uint32(mem[o:]))
		}
	}

	t := trace(mem, base, seeds, dol)

	w := os.Stdout
	if out != "" {
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	emit(bw, mem, base, t, loadAnnotations(annotate))

	// The coverage summary is the point of the tool: how much of the image is code we can
	// account for, and what is still unreachable.
	code := 0
	for _, c := range t.covered {
		if c {
			code++
		}
	}
	fmt.Fprintf(os.Stderr, "codetracegekko: %d instructions reached (%d bytes, %.1f%% of the image), %d functions, %d unresolved indirect branches\n",
		len(t.instr), code*4, 100*float64(code*4)/float64(len(mem)), len(t.callers), len(t.indirect))
	return nil
}

// load resolves the input the same way disgekko does.
func load(path, baseS string, skip int) (base uint32, mem []byte, dol *gc.DOL, err error) {
	if d, derr := gc.Open(path); derr == nil {
		defer d.Close()
		dol, err = d.DOL()
		if err != nil {
			return 0, nil, nil, err
		}
		base, mem = dol.Flat()
		return base, mem, dol, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, nil, err
	}
	if d, derr := gc.ParseDOL(raw); derr == nil && d.Entry != 0 {
		base, mem = d.Flat()
		return base, mem, d, nil
	}
	if base, err = hx(baseS); err != nil {
		return 0, nil, nil, fmt.Errorf("bad -base %q", baseS)
	}
	if skip < 0 || skip > len(raw) {
		return 0, nil, nil, fmt.Errorf("bad -skip %d", skip)
	}
	return base, raw[skip:], nil, nil
}

// trace walks every path reachable from the seeds.
func trace(mem []byte, base uint32, seeds []uint32, dol *gc.DOL) *traced {
	t := &traced{
		instr:   map[uint32]gekko.Inst{},
		covered: make([]bool, len(mem)/4),
		callers: map[uint32]int{},
	}
	queued := map[uint32]bool{}
	var work []uint32

	push := func(a uint32) {
		if queued[a] {
			return
		}
		queued[a] = true
		work = append(work, a)
	}
	for _, s := range seeds {
		push(s)
		t.callers[s] = 0 // a seed is a function, whether or not anything calls it
	}

	inRange := func(a uint32) bool {
		return a >= base && int(a-base)+4 <= len(mem) && a%4 == 0
	}

	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]

		for {
			if !inRange(pc) {
				break
			}
			// A descent that wanders out of a code segment is a descent that has gone
			// wrong; stop rather than decode the constant pool.
			if dol != nil && !dol.Text(pc) {
				t.stops = append(t.stops, pc)
				break
			}
			idx := (pc - base) / 4
			if t.covered[idx] {
				break // already walked from here
			}
			t.covered[idx] = true

			in := gekko.Decode(mem[pc-base:], pc)
			t.instr[pc] = in

			switch in.Flow {
			case gekko.FlowSeq, gekko.FlowIndCall:
				// An indirect call returns, so the path carries on. Note the site.
				if in.Flow == gekko.FlowIndCall {
					t.indirect = append(t.indirect, pc)
				}
				pc += 4
				continue

			case gekko.FlowBranch:
				// Conditional: BOTH the target and the fall-through are live. On PowerPC
				// a conditional bclr/bcctr is also a FlowBranch and has no static target,
				// which is why HasTarget is checked rather than assumed.
				if in.HasTarget {
					push(in.Target)
				} else {
					t.indirect = append(t.indirect, pc)
				}
				pc += 4
				continue

			case gekko.FlowCall:
				if in.HasTarget {
					t.callers[in.Target]++
					push(in.Target)
				}
				pc += 4
				continue

			case gekko.FlowJump:
				if in.HasTarget {
					pc = in.Target
					continue
				}
				t.stops = append(t.stops, pc)

			case gekko.FlowIndJump:
				// A bctr with no known target: a switch, or a tail call through a vtable.
				t.indirect = append(t.indirect, pc)
				t.stops = append(t.stops, pc)

			case gekko.FlowReturn, gekko.FlowStop:
				t.stops = append(t.stops, pc)
			}
			break // the path ended
		}
	}
	return t
}

type annot struct {
	name string
	desc string
}

func loadAnnotations(path string) map[uint32]annot {
	ann := map[uint32]annot{}
	if path == "" {
		return ann
	}
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codetracegekko: -annotate:", err)
		return ann
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		a, err := hx(f[0])
		if err != nil {
			continue
		}
		ann[a] = annot{name: f[1], desc: strings.Join(f[2:], " ")}
	}
	return ann
}

// emit prints the listing: the reachable code as instructions, headed by a function
// label wherever something calls in, and everything else as data.
func emit(w *bufio.Writer, mem []byte, base uint32, t *traced, ann map[uint32]annot) {
	n := len(mem) / 4
	i := 0
	for i < n {
		if !t.covered[i] {
			// A run of data. Print it compactly and move on — the point of the tracer is
			// that this is NOT code, and pretending otherwise is what it exists to avoid.
			start := i
			for i < n && !t.covered[i] {
				i++
			}
			emitData(w, mem, base, start, i)
			continue
		}
		a := base + uint32(i*4)
		if callers, ok := t.callers[a]; ok {
			name := fmt.Sprintf("sub_%08X", a)
			desc := ""
			if x, ok := ann[a]; ok {
				name, desc = x.name, x.desc
			}
			fmt.Fprintf(w, "\n; ==== %s  0x%08X", name, a)
			if callers > 0 {
				fmt.Fprintf(w, "  (%d callers)", callers)
			}
			if desc != "" {
				fmt.Fprintf(w, "  — %s", desc)
			}
			fmt.Fprintln(w, " ====")
		} else if x, ok := ann[a]; ok {
			fmt.Fprintf(w, "\n; ---- %s — %s ----\n", x.name, x.desc)
		}

		in := t.instr[a]
		o := i * 4
		fmt.Fprintf(w, "%08X  %02X %02X %02X %02X  %s\n",
			a, mem[o], mem[o+1], mem[o+2], mem[o+3], in.Text)
		i++
	}
}

func emitData(w *bufio.Writer, mem []byte, base uint32, start, end int) {
	if end-start == 0 {
		return
	}
	fmt.Fprintf(w, "\n; ---- data: 0x%08X..0x%08X (%d bytes) ----\n",
		base+uint32(start*4), base+uint32(end*4), (end-start)*4)
	// Print at most a few lines of it: the listing is about the code, and a megabyte of
	// hex helps nobody.
	const maxLines = 4
	for i, line := start, 0; i < end && line < maxLines; i, line = i+4, line+1 {
		a := base + uint32(i*4)
		fmt.Fprintf(w, "%08X ", a)
		var ascii []byte
		for j := i; j < i+4 && j < end; j++ {
			o := j * 4
			fmt.Fprintf(w, " %02X%02X%02X%02X", mem[o], mem[o+1], mem[o+2], mem[o+3])
			for k := 0; k < 4; k++ {
				c := mem[o+k]
				if c >= 0x20 && c < 0x7F {
					ascii = append(ascii, c)
				} else {
					ascii = append(ascii, '.')
				}
			}
		}
		fmt.Fprintf(w, "  |%s|\n", ascii)
	}
	if end-start > maxLines*4 {
		fmt.Fprintf(w, "; ... %d more bytes\n", (end-start-maxLines*4)*4)
	}
}
