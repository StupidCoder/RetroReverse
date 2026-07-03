// codetrace68k is a recursive-descent ("code-tracing") Motorola 68000
// disassembler — the 68k counterpart to codetrace6502. Starting from one or
// more entry points it follows every branch, jump and call (using m68k's Flow
// classification), marks which bytes are reachable code, and leaves the rest as
// data, so jump tables and graphics aren't mis-decoded as instructions. JMP/JSR
// through a register or indexed effective address can't be followed statically;
// those are reported as unresolved.
//
// Usage:
//
//	codetrace68k -base HEX [-skip n] -entry A,B,C [-table ADDR:N ...] [-annotate FILE] [-o out] file
//
// file is a raw 68000 code blob loaded at -base; -skip steps past an AmigaDOS
// hunk header. -table seeds N big-endian 32-bit pointers as code. Addresses are
// hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/m68k"
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
	skip := flag.Int("skip", 0, "skip this many leading bytes (e.g. past a hunk header)")
	entry := flag.String("entry", "", "comma-separated entry addresses (hex)")
	var tables multiFlag
	flag.Var(&tables, "table", "jump table to seed as code, ADDR:N (N big-endian longs); repeatable")
	annotate := flag.String("annotate", "", "annotations file: lines \"ADDR name description\" (# comments)")
	out := flag.String("o", "", "write disassembly to this file (default stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetrace68k -base HEX [-skip n] -entry A,B,C [-table ADDR:N ...] [-annotate FILE] [-o out] file")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *entry, tables, *annotate, *out); err != nil {
		fmt.Fprintln(os.Stderr, "codetrace68k:", err)
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
	for _, t := range tables {
		parts := strings.SplitN(t, ":", 2)
		a, err := hx(parts[0])
		if err != nil || len(parts) != 2 {
			return fmt.Errorf("bad -table %q (want ADDR:N)", t)
		}
		nw, _ := strconv.Atoi(parts[1])
		for i := 0; i < nw; i++ {
			p := int(a-base) + i*4
			if p < 0 || p+4 > len(code) {
				continue
			}
			seeds = append(seeds, uint32(code[p])<<24|uint32(code[p+1])<<16|uint32(code[p+2])<<8|uint32(code[p+3]))
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
	fmt.Fprintf(os.Stderr, "image $%06X-$%06X (%d bytes): %d code, %d data; %d routines, %d unresolved indirect jumps, %d stop-hits\n",
		lo, hi-1, hi-lo, codeBytes, int(hi-lo)-codeBytes, len(tr.callers), len(tr.indirect), len(tr.stops))
	return nil
}

type traced struct {
	instr    map[uint32]m68k.Inst
	covered  []bool
	callers  map[uint32]int // BSR/JSR target -> caller count
	indirect []uint32       // JMP through a register/indexed EA
	stops    []uint32       // ILLEGAL/STOP/data hits
}

func trace(code []byte, base uint32, seeds []uint32) *traced {
	hi := base + uint32(len(code))
	inRange := func(a uint32) bool { return a >= base && a < hi }
	t := &traced{instr: map[uint32]m68k.Inst{}, covered: make([]bool, len(code)), callers: map[uint32]int{}}

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
			in := m68k.Decode(code[off:], pc)
			if in.Len <= 0 {
				break
			}
			t.instr[pc] = in
			for i := 0; i < in.Len && off+i < len(code); i++ {
				t.covered[off+i] = true
			}
			switch in.Flow {
			case m68k.FlowBranch:
				push(in.Target)
				pc += uint32(in.Len)
			case m68k.FlowCall:
				t.callers[in.Target]++
				push(in.Target)
				pc += uint32(in.Len)
			case m68k.FlowJump:
				push(in.Target)
				goto pathEnd
			case m68k.FlowReturn:
				goto pathEnd
			case m68k.FlowIndJump:
				t.indirect = append(t.indirect, in.Addr)
				goto pathEnd
			case m68k.FlowStop:
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

// emit writes the listing: a header before each subroutine (BSR/JSR target) or
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
		fmt.Fprintf(w, "%06X  .dc.b %-47s ; %s\n", p, strings.Join(bs, " "), string(asc))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
