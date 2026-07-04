// codetracearm is a recursive-descent ("code-tracing") disassembler for the Nintendo
// DS ARM cores (ARM9 / ARM7), the counterpart of codetrace68k / codetracesm83.
// Starting from one or more entry points it follows every branch, call, jump and
// interworking BX/BLX, marks which bytes are reachable code, and leaves everything
// else as data — so graphics, tables and the like don't get mis-decoded.
//
// The ARM twist over the other CPUs is two instruction sets. Each address is traced
// in a definite state (ARM or Thumb); the interworking branches switch state, and
// the worklist carries the state with each pending address. Entry points default to
// ARM; append "t" to an entry (or set -thumb) to seed it as Thumb. Indirect
// transfers whose target isn't statically known (BX/BLX through a register, LDR pc,
// function pointers) are reported as unresolved — supply their jump tables with
// -table, or add the discovered targets as further -entry points.
//
// Usage:
//
//	codetracearm [-base ADDR] [-skip N] [-thumb] -entry A,Bt,C [-table ADDR:N] [-annotate FILE] [-o out] image.bin
//
// The image is loaded flat at -base (default 0); -skip drops a header. Addresses are
// hex. An entry/table pointer with bit 0 set selects Thumb, matching the hardware's
// interworking convention.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/arm"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex)")
	skip := flag.String("skip", "0", "bytes to drop from the front of the file (hex)")
	thumbDefault := flag.Bool("thumb", false, "entries with no a/t suffix default to Thumb")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex; suffix t=Thumb, a=ARM)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N 32-bit pointers); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write disassembly to this file (default stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetracearm [-base A] [-skip N] [-thumb] -entry A,Bt,C [-table ADDR:N] [-annotate F] [-o out] image.bin")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *thumbDefault, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetracearm:", err)
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

// seed is a pending trace point: an address and the state to decode it in.
type seed struct {
	addr  uint32
	thumb bool
}

// parseEntry parses "ADDR", "ADDRt" or "ADDRa" into a seed. An odd address also
// selects Thumb (the interworking convention) and is aligned down.
func parseEntry(s string, thumbDefault bool) (seed, error) {
	thumb := thumbDefault
	if strings.HasSuffix(s, "t") {
		thumb, s = true, strings.TrimSuffix(s, "t")
	} else if strings.HasSuffix(s, "a") {
		thumb, s = false, strings.TrimSuffix(s, "a")
	}
	a, err := hx(s)
	if err != nil {
		return seed{}, err
	}
	return fromTarget(a, thumb), nil
}

// fromTarget normalises a code pointer: bit 0 selects Thumb, and the returned
// address is aligned to the instruction width.
func fromTarget(a uint32, thumb bool) seed {
	if a&1 == 1 {
		thumb = true
	}
	if thumb {
		return seed{a &^ 1, true}
	}
	return seed{a &^ 3, false}
}

func run(path, baseS, skipS string, thumbDefault bool, entryStr string, tables multiFlag, annPath, outPath string) error {
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
	skip, err := hx(skipS)
	if err != nil || int(skip) > len(raw) {
		return fmt.Errorf("bad -skip %q", skipS)
	}
	mem := raw[skip:]

	var seeds []seed
	for _, s := range strings.Split(entryStr, ",") {
		if s == "" {
			continue
		}
		sd, err := parseEntry(s, thumbDefault)
		if err != nil {
			return fmt.Errorf("bad -entry %q: %v", s, err)
		}
		seeds = append(seeds, sd)
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
			seeds = append(seeds, fromTarget(rd32(a+uint32(i)*4), thumbDefault))
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
	instr    map[uint32]arm.Inst
	covered  []bool
	callers  map[uint32]int
	indirect []uint32
	stops    []uint32
}

func trace(mem []byte, base uint32, seeds []seed) *traced {
	t := &traced{instr: map[uint32]arm.Inst{}, covered: make([]bool, len(mem)), callers: map[uint32]int{}}
	inRange := func(a uint32) bool { return a >= base && int(a-base) < len(mem) }

	work := append([]seed(nil), seeds...)
	queued := map[seed]bool{}
	for _, s := range seeds {
		queued[s] = true
	}
	push := func(a uint32, thumb bool) {
		s := fromTarget(a, thumb)
		if !queued[s] {
			queued[s] = true
			work = append(work, s)
		}
	}
	for len(work) > 0 {
		s := work[len(work)-1]
		work = work[:len(work)-1]
		pc, thumb := s.addr, s.thumb
		for {
			if !inRange(pc) {
				break
			}
			if _, done := t.instr[pc]; done {
				break
			}
			in := arm.Decode(mem[pc-base:], pc, thumb)
			t.instr[pc] = in
			for i := 0; i < in.Len && int(pc-base)+i < len(mem); i++ {
				t.covered[int(pc-base)+i] = true
			}
			switch in.Flow {
			case arm.FlowBranch:
				push(in.Target, in.TargetThumb)
				pc += uint32(in.Len)
			case arm.FlowCall:
				t.callers[in.Target]++
				push(in.Target, in.TargetThumb)
				pc += uint32(in.Len)
			case arm.FlowIndCall:
				t.indirect = append(t.indirect, in.Addr)
				pc += uint32(in.Len) // an indirect call returns: keep tracing the fall-through
			case arm.FlowJump:
				push(in.Target, in.TargetThumb)
				goto pathEnd
			case arm.FlowReturn:
				goto pathEnd
			case arm.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case arm.FlowStop:
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
		if in, ok := t.instr[a]; ok {
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
			st := "  "
			if in.Thumb {
				st = "T " // mark Thumb-state instructions
			}
			fmt.Fprintf(w, "%08X  %s%-12s %s\n", a, st, strings.Join(raw, " "), in.Text)
			pos += in.Len
			continue
		}
		if named {
			fmt.Fprintf(w, "\n; --- %s  $%08X — %s (data) ---\n", an.name, a, an.desc)
		}
		start := pos
		pos++
		for pos < len(mem) {
			if _, ok := t.instr[base+uint32(pos)]; ok {
				break
			}
			if _, ok := ann[base+uint32(pos)]; ok {
				break
			}
			pos++
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
