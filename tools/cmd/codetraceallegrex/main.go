// codetraceallegrex is a recursive-descent ("code-tracing") disassembler for MIPS
// R3000 (PlayStation) code, the counterpart of codetracearm / codetracesm83.
// Starting from one or more entry points it follows every branch, call and jump,
// marks which bytes are reachable code, and leaves everything else as data — so
// tables, textures and the like don't get mis-decoded.
//
// The MIPS twist over the other CPUs is the delay slot: the instruction after a
// branch or jump always executes before control transfers. The tracer decodes
// and covers that delay-slot instruction before honouring the branch, and a
// conditional branch or call resumes tracing at the instruction *after* the slot.
//
// Indirect transfers whose target isn't statically known (jr/jalr through a
// register, jump tables) are reported as unresolved — supply their tables with
// -table, or add discovered targets as further -entry points. That feedback loop,
// often seeded from the oracle's live trace, is the reverse-engineering workflow.
//
// Usage:
//
//	codetraceallegrex [-base ADDR] [-skip N] -entry A,B,C [-table ADDR:N] [-annotate FILE] [-o out] image.bin
//
// The image is loaded flat at -base (default 0); -skip drops that many leading file
// bytes (decimal; 2048 for a PS-X EXE header). All addresses are hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/allegrex"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N 32-bit pointers); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write disassembly to this file (default stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetraceallegrex [-base A] [-skip N] -entry A,B,C [-table ADDR:N] [-annotate F] [-o out] image.bin")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetraceallegrex:", err)
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
		desc := strings.TrimSpace(rest[len(name):])
		m[a] = annot{name: name, desc: desc}
	}
	return m, nil
}

func run(path, baseS string, skip int, entryStr string, tables multiFlag, annPath, outPath string) error {
	ann, err := loadAnnotations(annPath)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	base, err := hx(baseS)
	if err != nil {
		return fmt.Errorf("bad -base %q", baseS)
	}
	if skip < 0 || skip > len(raw) {
		return fmt.Errorf("bad -skip %d", skip)
	}
	mem := raw[skip:]

	var seeds []uint32
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
	rd32 := func(a uint32) uint32 {
		o := int(a - base)
		if o < 0 || o+4 > len(mem) {
			return 0
		}
		return uint32(mem[o]) | uint32(mem[o+1])<<8 | uint32(mem[o+2])<<16 | uint32(mem[o+3])<<24
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
		return fmt.Errorf("no -entry points given")
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
	fmt.Fprintf(os.Stderr, "traced $%08X-$%08X: %d code, %d data; %d routines, %d unresolved indirect, %d stop-hits\n",
		base, base+uint32(len(mem))-1, code, len(mem)-code, len(tr.callers), len(tr.indirect), len(tr.stops))
	return nil
}

type traced struct {
	instr    map[uint32]allegrex.Inst
	covered  []bool
	callers  map[uint32]int
	indirect []uint32
	stops    []uint32
}

func trace(mem []byte, base uint32, seeds []uint32) *traced {
	t := &traced{instr: map[uint32]allegrex.Inst{}, covered: make([]bool, len(mem)), callers: map[uint32]int{}}
	inRange := func(a uint32) bool { return a >= base && int(a-base) < len(mem) }
	cover := func(a uint32, n int) {
		for i := 0; i < n && int(a-base)+i < len(mem); i++ {
			t.covered[int(a-base)+i] = true
		}
	}

	work := append([]uint32(nil), seeds...)
	queued := map[uint32]bool{}
	for _, s := range seeds {
		queued[s] = true
	}
	push := func(a uint32) {
		a &^= 3
		if !queued[a] {
			queued[a] = true
			work = append(work, a)
		}
	}
	// decodeAt records an instruction once and covers its bytes.
	decodeAt := func(pc uint32) allegrex.Inst {
		if in, ok := t.instr[pc]; ok {
			return in
		}
		in := allegrex.Decode(mem[pc-base:], pc)
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
			// A branch/jump/call always executes its delay slot first; decode and
			// cover it before acting on the control transfer.
			if in.HasDelay && inRange(pc+4) {
				decodeAt(pc + 4)
			}
			switch in.Flow {
			case allegrex.FlowBranch:
				push(in.Target)
				pc += 8 // past the instruction and its delay slot
			case allegrex.FlowCall:
				t.callers[in.Target]++
				push(in.Target)
				pc += 8
			case allegrex.FlowIndCall:
				t.indirect = append(t.indirect, in.Addr)
				pc += 8 // jalr returns: keep tracing after the delay slot
			case allegrex.FlowJump:
				push(in.Target)
				goto pathEnd
			case allegrex.FlowReturn:
				goto pathEnd
			case allegrex.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case allegrex.FlowStop:
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
				fmt.Fprintf(w, "\n; ==== %s  $%08X  (%d caller%s) — %s ====\n", an.name, a, t.callers[a], plural(t.callers[a]), an.desc)
			case t.callers[a] > 0:
				fmt.Fprintf(w, "\n; ==== sub_%08X (%d caller%s) ====\n", a, t.callers[a], plural(t.callers[a]))
			case named:
				fmt.Fprintf(w, "\n; --- %s  $%08X — %s ---\n", an.name, a, an.desc)
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
			fmt.Fprintf(w, "\n; --- %s  $%08X — %s (data) ---\n", an.name, a, an.desc)
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
