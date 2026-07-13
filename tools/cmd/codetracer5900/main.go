// codetracer5900 is a recursive-descent ("code-tracing") disassembler for Emotion
// Engine (PlayStation 2) code, the counterpart of codetracer4300 / codetracemips.
// Starting from one or more entry points it follows every branch, call and jump,
// marks which bytes are reachable code, and leaves everything else as data — so
// tables and packed vector data don't get mis-decoded.
//
// Two MIPS features shape the descent. The delay slot: the instruction after a
// branch or jump always executes before control transfers, so it is decoded and
// covered before the transfer is honoured. And the branch-likely family (beql,
// bnel, and the REGIMM and COP1/COP2 variants), whose delay slot executes only when
// the branch is taken — it is still reachable code, so it is still covered, but the
// fall-through path resumes after it exactly as for an ordinary branch.
//
// A PS2 boot ELF is recognised: it supplies the load address, the entry point, and —
// where it ships a symbol table — the names of every function, which are used as
// both extra descent seeds and the listing's labels. So on a symbol-bearing
// executable the useful invocation carries no -entry at all.
//
// Indirect transfers whose target isn't statically known (jr/jalr through a
// register, vtable dispatch) are reported as unresolved — supply their tables with
// -table, or add discovered targets as further -entry points.
//
// Usage:
//
//	codetracer5900 [-base A] [-skip N] [-entry A,B,C] [-table ADDR:N] [-annotate F] [-o out] SCUS_971.24
//
// All addresses are hex.
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/r5900"
	"retroreverse.com/tools/platform/ps2"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex; ignored for an ELF)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex; an ELF supplies its own)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N 32-bit pointers); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write disassembly to this file (default stdout)")
	symbols := flag.Bool("symbols", true, "seed the descent from the ELF's function symbols")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetracer5900 [-base A] [-entry A,B] [-table ADDR:N] [-o out] SCUS_971.24")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out, *symbols); err != nil {
		fmt.Fprintln(os.Stderr, "codetracer5900:", err)
		os.Exit(1)
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

type annot struct{ name, desc string }

func loadAnnotations(path string) (map[uint32]annot, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := map[uint32]annot{}
	for _, line := range strings.Split(string(data), "\n") {
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
			return nil, fmt.Errorf("annotations: bad address %q", f[0])
		}
		name := f[1]
		rest := strings.TrimSpace(line[len(f[0]):])
		m[a] = annot{name: name, desc: strings.TrimSpace(rest[len(name):])}
	}
	return m, nil
}

func run(path, baseS string, skip int, entryStr string, tables multiFlag, annPath, outPath string, useSyms bool) error {
	ann, err := loadAnnotations(annPath)
	if err != nil {
		return err
	}
	if ann == nil {
		ann = map[uint32]annot{}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var (
		mem   []byte
		base  uint32
		seeds []uint32
	)
	if e, err := ps2.LoadELF(raw); err == nil {
		base, mem = e.Flat()
		seeds = append(seeds, e.Entry)
		if useSyms {
			// Every function the symbol table names is a descent seed. This matters far
			// more than it sounds: GOAL's kernel reaches most of itself through function
			// pointers, so a descent from the entry point alone would leave the majority
			// of the .text unvisited and print it as data.
			for _, s := range e.Symbols {
				if s.Func {
					seeds = append(seeds, s.Addr)
					if _, taken := ann[s.Addr]; !taken {
						ann[s.Addr] = annot{name: s.Name}
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "codetracer5900: ELF at $%08X (%d bytes), entry $%08X, %d symbol seeds\n",
			base, len(mem), e.Entry, len(seeds)-1)
	} else {
		if base, err = hx(baseS); err != nil {
			return fmt.Errorf("bad -base %q", baseS)
		}
		if skip < 0 || skip > len(raw) {
			return fmt.Errorf("bad -skip %d", skip)
		}
		mem = raw[skip:]
	}

	for _, s := range strings.Split(entryStr, ",") {
		if s == "" {
			continue
		}
		a, err := hx(s)
		if err != nil {
			return fmt.Errorf("bad -entry %q: %v", s, err)
		}
		seeds = append(seeds, a&^3)
	}
	// Pointer tables hold little-endian words on this machine.
	rd32 := func(a uint32) uint32 {
		o := int(a - base)
		if o < 0 || o+4 > len(mem) {
			return 0
		}
		return binary.LittleEndian.Uint32(mem[o:])
	}
	for _, t := range tables {
		parts := strings.SplitN(t, ":", 2)
		a, err := hx(parts[0])
		if err != nil || len(parts) != 2 {
			return fmt.Errorf("bad -table %q (want ADDR:N)", t)
		}
		n, _ := strconv.Atoi(parts[1])
		for i := 0; i < n; i++ {
			seeds = append(seeds, rd32(a+uint32(i)*4)&^3)
		}
	}
	if len(seeds) == 0 {
		return fmt.Errorf("no -entry points given, and the image is not an ELF with symbols")
	}

	tr := trace(mem, base, seeds)

	w := bufio.NewWriter(os.Stdout)
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = bufio.NewWriter(f)
	}
	defer w.Flush()
	emit(w, mem, base, tr, ann)

	code := 0
	for _, ok := range tr.covered {
		if ok {
			code++
		}
	}
	fmt.Fprintf(os.Stderr, "traced $%08X-$%08X: %d code, %d data (%.1f%% code); %d routines, %d unresolved indirect, %d stop-hits\n",
		base, base+uint32(len(mem))-1, code, len(mem)-code,
		100*float64(code)/float64(len(mem)), len(tr.callers), len(tr.indirect), len(tr.stops))
	return nil
}

type traced struct {
	instr    map[uint32]r5900.Inst
	covered  []bool
	callers  map[uint32]int
	indirect []uint32
	stops    []uint32
}

func trace(mem []byte, base uint32, seeds []uint32) *traced {
	t := &traced{instr: map[uint32]r5900.Inst{}, covered: make([]bool, len(mem)), callers: map[uint32]int{}}
	inRange := func(a uint32) bool { return a >= base && int(a-base)+4 <= len(mem) }
	cover := func(a uint32, n int) {
		for i := 0; i < n && int(a-base)+i < len(mem); i++ {
			t.covered[int(a-base)+i] = true
		}
	}

	var work []uint32
	queued := map[uint32]bool{}
	push := func(a uint32) {
		a &^= 3
		if !queued[a] {
			queued[a] = true
			work = append(work, a)
		}
	}
	for _, s := range seeds {
		push(s)
	}

	decodeAt := func(pc uint32) r5900.Inst {
		if in, ok := t.instr[pc]; ok {
			return in
		}
		in := r5900.Decode(mem[pc-base:], pc)
		t.instr[pc] = in
		cover(pc, in.Len)
		return in
	}

	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		for {
			if !inRange(pc) {
				break
			}
			if _, done := t.instr[pc]; done {
				break
			}
			in := decodeAt(pc)
			// A branch/jump/call always executes its delay slot first; decode and cover
			// it before acting on the control transfer. A branch-likely's slot runs only
			// on the taken path, but it is code either way.
			if in.HasDelay && inRange(pc+4) {
				decodeAt(pc + 4)
			}
			switch in.Flow {
			case r5900.FlowBranch:
				push(in.Target)
				pc += 8 // past the instruction and its delay slot
			case r5900.FlowCall:
				t.callers[in.Target]++
				push(in.Target)
				pc += 8
			case r5900.FlowIndCall:
				t.indirect = append(t.indirect, in.Addr)
				pc += 8 // jalr returns: keep tracing after the delay slot
			case r5900.FlowJump:
				push(in.Target)
				goto pathEnd
			case r5900.FlowReturn:
				goto pathEnd
			case r5900.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case r5900.FlowStop:
				t.stops = append(t.stops, in.Addr)
				goto pathEnd
			default: // FlowSeq
				pc += uint32(in.Len)
			}
		}
	pathEnd:
	}
	sort.Slice(t.indirect, func(i, j int) bool { return t.indirect[i] < t.indirect[j] })
	return t
}

func emit(w *bufio.Writer, mem []byte, base uint32, t *traced, ann map[uint32]annot) {
	pos := 0
	for pos < len(mem) {
		a := base + uint32(pos)
		an, named := ann[a]
		if in, ok := t.instr[a]; ok && in.Len > 0 {
			switch {
			case t.callers[a] > 0 && named:
				fmt.Fprintf(w, "\n; ==== %s  $%08X  (%d caller%s) %s====\n",
					an.name, a, t.callers[a], plural(t.callers[a]), descOf(an))
			case t.callers[a] > 0:
				fmt.Fprintf(w, "\n; ==== sub_%08X (%d caller%s) ====\n", a, t.callers[a], plural(t.callers[a]))
			case named:
				fmt.Fprintf(w, "\n; --- %s  $%08X %s---\n", an.name, a, descOf(an))
			}
			raw := make([]string, in.Len)
			for i := 0; i < in.Len; i++ {
				raw[i] = fmt.Sprintf("%02X", mem[pos+i])
			}
			fmt.Fprintf(w, "%08X  %-11s %s\n", a, strings.Join(raw, " "), strings.TrimSpace(in.Text))
			pos += in.Len
			continue
		}
		if named {
			fmt.Fprintf(w, "\n; --- %s  $%08X %s(data) ---\n", an.name, a, descOf(an))
		}
		start := pos
		pos += 4
		for pos < len(mem) {
			if in, ok := t.instr[base+uint32(pos)]; ok && in.Len > 0 {
				break
			}
			if _, ok := ann[base+uint32(pos)]; ok {
				break
			}
			pos += 4
		}
		if pos > len(mem) {
			pos = len(mem)
		}
		emitData(w, mem, base, start, pos)
	}
}

func descOf(a annot) string {
	if a.desc == "" {
		return ""
	}
	return "— " + a.desc + " "
}

func emitData(w *bufio.Writer, mem []byte, base uint32, start, end int) {
	for p := start; p < end; p += 16 {
		n := end - p
		if n > 16 {
			n = 16
		}
		bs := make([]string, n)
		asc := make([]byte, n)
		for i := 0; i < n; i++ {
			bs[i] = fmt.Sprintf("%02X", mem[p+i])
			c := mem[p+i]
			if c >= 0x20 && c < 0x7f {
				asc[i] = c
			} else {
				asc[i] = '.'
			}
		}
		fmt.Fprintf(w, "%08X  .byte %-47s ; %s\n", base+uint32(p), strings.Join(bs, " "), string(asc))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
