// codetracex86 is a recursive-descent ("code-tracing") 16-bit real-mode x86
// disassembler — the x86 counterpart to codetrace6502/codetrace68k. Starting
// from one or more entry points it follows every branch, jump and call (using
// x86's Flow classification), marks which bytes are reachable code, and leaves
// the rest as data, so jump tables and embedded data aren't mis-decoded as
// instructions. A CALL/JMP through a register or memory operand can't be
// followed statically; those are reported as unresolved.
//
// Usage:
//
//	codetracex86 -base HEX [-skip n] -entry A,B,C [-table ADDR:N ...] [-annotate FILE] [-o out] file
//
// file is a raw x86 code blob loaded at -base; -skip steps past the MZ header
// (e.g. -skip 0x3200 for UW.EXE). -table seeds N 16-bit little-endian near
// offsets (relative to -base) as code — the usual real-mode jump-table form.
// Addresses are hex. The flat model treats -base as load-segment 0, so near
// targets and module-relative far pointers resolve into the same blob.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/x86"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 32)
	return uint32(v), err
}

func main() {
	base := flag.String("base", "0", "load address of the first byte (hex)")
	skip := flag.Int("skip", 0, "skip this many leading bytes (e.g. past the MZ header)")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N 16-bit LE near offsets); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write disassembly to this file (default stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetracex86 -base HEX [-skip n] -entry A,B,C [-table ADDR:N ...] [-annotate FILE] [-o out] file")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetracex86:", err)
		os.Exit(1)
	}
}

type annot struct{ name, desc string }

// loadAnnotations reads "ADDR name rest-of-line-is-description" entries.
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

func run(path, baseStr string, skip int, entryStr string, tables multiFlag, annPath, outPath string) error {
	ann, err := loadAnnotations(annPath)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if skip < 0 || skip > len(raw) {
		return fmt.Errorf("-skip out of range")
	}
	code := raw[skip:]
	base, err := hx(baseStr)
	if err != nil {
		return fmt.Errorf("bad -base: %v", err)
	}
	lo, hi := base, base+uint32(len(code))

	var seeds []uint32
	for _, s := range strings.Split(entryStr, ",") {
		if s == "" {
			continue
		}
		a, err := hx(s)
		if err != nil {
			return fmt.Errorf("bad -entry %q: %v", s, err)
		}
		seeds = append(seeds, a)
	}
	for _, tbl := range tables {
		parts := strings.SplitN(tbl, ":", 2)
		a, err := hx(parts[0])
		if err != nil || len(parts) != 2 {
			return fmt.Errorf("bad -table %q (want ADDR:N)", tbl)
		}
		nw, _ := strconv.Atoi(parts[1])
		for i := 0; i < nw; i++ {
			p := int(a-base) + i*2
			if p < 0 || p+2 > len(code) {
				continue
			}
			seeds = append(seeds, base+uint32(code[p])|uint32(code[p+1])<<8)
		}
	}

	tr := trace(code, base, seeds)

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
	emit(w, code, base, tr, ann)

	codeBytes := 0
	for _, in := range tr.instr {
		codeBytes += in.Len
	}
	fmt.Fprintf(os.Stderr, "image $%06X-$%06X (%d bytes): %d code, %d data; %d routines, %d unresolved indirect jumps, %d indirect calls, %d stop-hits\n",
		lo, hi-1, hi-lo, codeBytes, int(hi-lo)-codeBytes, len(tr.callers), len(tr.indirect), len(tr.indCalls), len(tr.stops))
	return nil
}

type traced struct {
	instr    map[uint32]x86.Inst
	covered  []bool
	callers  map[uint32]int // CALL target -> caller count
	indirect []uint32       // JMP through a register/memory
	indCalls []uint32       // CALL through a register/memory
	stops    []uint32       // undefined/HLT/data hits
}

func trace(code []byte, base uint32, seeds []uint32) *traced {
	hi := base + uint32(len(code))
	inRange := func(a uint32) bool { return a >= base && a < hi }
	t := &traced{instr: map[uint32]x86.Inst{}, covered: make([]bool, len(code)), callers: map[uint32]int{}}

	work := append([]uint32(nil), seeds...)
	queued := map[uint32]bool{}
	for _, s := range seeds {
		queued[s] = true
	}
	push := func(a uint32) {
		if inRange(a) && !queued[a] {
			queued[a] = true
			work = append(work, a)
		}
	}
	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		for inRange(pc) {
			if _, done := t.instr[pc]; done {
				break
			}
			off := int(pc - base)
			in := x86.Decode(code[off:], pc)
			if in.Len <= 0 {
				break
			}
			t.instr[pc] = in
			for i := 0; i < in.Len && off+i < len(code); i++ {
				t.covered[off+i] = true
			}
			switch in.Flow {
			case x86.FlowBranch:
				if in.HasTarget {
					push(in.Target)
				}
				pc += uint32(in.Len)
			case x86.FlowCall:
				if in.HasTarget {
					t.callers[in.Target]++
					push(in.Target)
				} else {
					t.indCalls = append(t.indCalls, in.Addr) // indirect call: continue past it
				}
				pc += uint32(in.Len)
			case x86.FlowJump:
				if in.HasTarget {
					push(in.Target)
				}
				goto pathEnd
			case x86.FlowReturn:
				goto pathEnd
			case x86.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case x86.FlowStop:
				t.stops = append(t.stops, in.Addr)
				goto pathEnd
			default: // FlowSeq
				pc += uint32(in.Len)
			}
		}
	pathEnd:
	}
	sort.Slice(t.indirect, func(i, j int) bool { return t.indirect[i] < t.indirect[j] })
	sort.Slice(t.indCalls, func(i, j int) bool { return t.indCalls[i] < t.indCalls[j] })
	return t
}

// emit writes the listing: a header before each subroutine (CALL target) or
// annotated address, the decoded instructions, and condensed data runs.
func emit(w *bufio.Writer, code []byte, base uint32, t *traced, ann map[uint32]annot) {
	hi := base + uint32(len(code))
	for pos := base; pos < hi; {
		off := int(pos - base)
		an, named := ann[pos]
		if in, ok := t.instr[pos]; ok {
			switch {
			case t.callers[pos] > 0 && named:
				fmt.Fprintf(w, "\n; ==== %s  $%06X  (%d caller%s) — %s ====\n", an.name, pos, t.callers[pos], plural(t.callers[pos]), an.desc)
			case t.callers[pos] > 0:
				fmt.Fprintf(w, "\n; ==== sub_%06X (%d caller%s) ====\n", pos, t.callers[pos], plural(t.callers[pos]))
			case named:
				fmt.Fprintf(w, "\n; --- %s  $%06X — %s ---\n", an.name, pos, an.desc)
			}
			raw := make([]string, 0, in.Len)
			for i := 0; i < in.Len && off+i < len(code); i++ {
				raw = append(raw, fmt.Sprintf("%02X", code[off+i]))
			}
			fmt.Fprintf(w, "%06X  %-29s %s\n", pos, strings.Join(raw, " "), in.Text)
			pos += uint32(in.Len)
			continue
		}
		if named {
			fmt.Fprintf(w, "\n; --- %s  $%06X — %s (data) ---\n", an.name, pos, an.desc)
		}
		start := pos
		pos++
		for pos < hi {
			if _, ok := t.instr[pos]; ok {
				break
			}
			if _, ok := ann[pos]; ok {
				break
			}
			pos++
		}
		emitData(w, code, base, start, pos)
	}
}

func emitData(w *bufio.Writer, code []byte, base, start, end uint32) {
	for p := start; p < end; p += 16 {
		n := end - p
		if n > 16 {
			n = 16
		}
		bs := make([]string, n)
		asc := make([]byte, n)
		for i := uint32(0); i < n; i++ {
			c := code[p-base+i]
			bs[i] = fmt.Sprintf("%02X", c)
			if c >= 0x20 && c < 0x7f {
				asc[i] = c
			} else {
				asc[i] = '.'
			}
		}
		fmt.Fprintf(w, "%06X  .db   %-47s ; %s\n", p, strings.Join(bs, " "), string(asc))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
