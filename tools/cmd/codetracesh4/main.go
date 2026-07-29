// codetracesh4 is a recursive-descent ("code-tracing") disassembler for
// Hitachi SH-4 (Sega Dreamcast) code, the counterpart of codetracer4300 /
// codetracearm. Starting from one or more entry points it follows every
// branch, call and jump, marks which bytes are reachable code, and leaves
// everything else as data — so tables, textures and the like don't get
// mis-decoded.
//
// Two SH-4 features shape the descent. The delay slot: the instruction after
// a delayed transfer (bra, bsr, jmp, jsr, rts, rte, bt/s, bf/s) executes
// before control moves, so it is decoded and covered before the transfer is
// honoured. And the literal pools: the only 32-bit constant load the ISA has
// is mov.w/mov.l @(disp,PC), a read from data words the compiler drops
// between functions, so every traced literal load contributes its pool word
// (Inst.LitAddr) to a data set that is both rendered as .word/.hword and
// enforced as a barrier — a traced path that would fall through into a known
// pool stops instead of misdecoding it.
//
// Indirect transfers whose target isn't statically known (jmp/jsr through a
// register, braf/bsrf) are reported as unresolved — supply their tables with
// -table, or add discovered targets as further -entry points. That feedback
// loop, often seeded from the oracle's live trace, is the reverse-engineering
// workflow.
//
// Usage:
//
//	codetracesh4 [-base ADDR] [-skip N] -entry A,B,C [-table ADDR:N] [-annotate FILE] [-o out] 1ST_READ.BIN
//
// The image is loaded flat at -base (default 0); -skip drops that many
// leading file bytes. All addresses are hex. A Dreamcast main binary loads at
// 8C010000.
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

	"retroreverse.com/tools/cpu/sh4"
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
		fmt.Fprintln(os.Stderr, "usage: codetracesh4 [-base A] [-skip N] -entry A,B,C [-table ADDR:N] [-annotate F] [-o out] 1ST_READ.BIN")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetracesh4:", err)
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
		seeds = append(seeds, a&^1)
	}
	// Pointer tables hold little-endian words, like everything else on the SH-4.
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
			seeds = append(seeds, rd32(a+uint32(i)*4)&^1)
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
	fmt.Fprintf(os.Stderr, "traced $%08X-$%08X: %d code, %d data; %d routines, %d literal-pool words, %d unresolved indirect, %d stop-hits\n",
		base, base+uint32(len(mem))-1, code, len(mem)-code, len(tr.callers), len(tr.lit), len(tr.indirect), len(tr.stops))
	return nil
}

type traced struct {
	instr    map[uint32]sh4.Inst
	covered  []bool
	callers  map[uint32]int
	lit      map[uint32]int // literal-pool word address -> size (2 or 4)
	indirect []uint32
	stops    []uint32
}

func trace(mem []byte, base uint32, seeds []uint32) *traced {
	t := &traced{instr: map[uint32]sh4.Inst{}, covered: make([]bool, len(mem)), callers: map[uint32]int{}, lit: map[uint32]int{}}
	inRange := func(a uint32) bool { return a >= base && int(a-base) < len(mem) }
	cover := func(a uint32, n int) {
		for i := 0; i < n && int(a-base)+i < len(mem); i++ {
			t.covered[int(a-base)+i] = true
		}
	}
	inLit := func(a uint32) bool {
		// A pool word is a barrier at any byte: check both the halfword itself
		// and the longword that would straddle it.
		if _, ok := t.lit[a]; ok {
			return true
		}
		if sz, ok := t.lit[a-2]; ok && sz == 4 {
			return true
		}
		return false
	}

	work := append([]uint32(nil), seeds...)
	queued := map[uint32]bool{}
	for _, s := range seeds {
		queued[s] = true
	}
	push := func(a uint32) {
		a &^= 1
		if !queued[a] {
			queued[a] = true
			work = append(work, a)
		}
	}
	// decodeAt records an instruction once, covers its bytes, and registers
	// any literal-pool word it reaches.
	decodeAt := func(pc uint32) sh4.Inst {
		if in, ok := t.instr[pc]; ok {
			return in
		}
		in := sh4.Decode(mem[pc-base:], pc)
		t.instr[pc] = in
		cover(pc, in.Len)
		if in.LitSize != 0 && inRange(in.LitAddr) {
			if old, ok := t.lit[in.LitAddr]; !ok || in.LitSize > old {
				t.lit[in.LitAddr] = in.LitSize
			}
		}
		return in
	}

	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		for {
			if !inRange(pc) || inLit(pc) {
				break
			}
			if _, done := t.instr[pc]; done {
				break
			}
			in := decodeAt(pc)
			// A delayed transfer always executes its slot first; decode and
			// cover it before acting on the control transfer.
			if in.HasDelay && inRange(pc+2) && !inLit(pc+2) {
				decodeAt(pc + 2)
			}
			step := uint32(in.Len)
			if in.HasDelay {
				step += 2 // past the instruction and its delay slot
			}
			switch in.Flow {
			case sh4.FlowBranch:
				push(in.Target)
				pc += step
			case sh4.FlowCall:
				t.callers[in.Target]++
				push(in.Target)
				pc += step
			case sh4.FlowIndCall:
				t.indirect = append(t.indirect, in.Addr)
				pc += step // jsr/bsrf returns: keep tracing after the delay slot
			case sh4.FlowJump:
				push(in.Target)
				goto pathEnd
			case sh4.FlowReturn:
				goto pathEnd
			case sh4.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case sh4.FlowStop:
				t.stops = append(t.stops, in.Addr)
				goto pathEnd
			default: // FlowSeq
				pc += step
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
			fmt.Fprintf(w, "%08X  %02X %02X        %s\n", a, mem[pos], mem[pos+1], strings.TrimSpace(in.Text))
			pos += in.Len
			continue
		}
		// A literal-pool word renders with its value: the constant is the point.
		if sz, ok := t.lit[a]; ok && pos+sz <= len(mem) {
			if named {
				fmt.Fprintf(w, "\n; --- %s  $%08X — %s (literal) ---\n", an.name, a, an.desc)
			}
			if sz == 4 {
				fmt.Fprintf(w, "%08X  .word 0x%08X\n", a, binary.LittleEndian.Uint32(mem[pos:]))
			} else {
				fmt.Fprintf(w, "%08X  .hword 0x%04X\n", a, binary.LittleEndian.Uint16(mem[pos:]))
			}
			pos += sz
			continue
		}
		if named {
			fmt.Fprintf(w, "\n; --- %s  $%08X — %s (data) ---\n", an.name, a, an.desc)
		}
		start := pos
		pos += 2
		for pos < len(mem) {
			b := base + uint32(pos)
			if in, ok := t.instr[b]; ok && in.Len > 0 {
				break
			}
			if _, ok := t.lit[b]; ok {
				break
			}
			if _, ok := ann[b]; ok {
				break
			}
			pos += 2
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
