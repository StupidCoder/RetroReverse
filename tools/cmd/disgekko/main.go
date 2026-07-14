// disgekko is a linear disassembler for Gekko code — the PowerPC 750 at the heart of
// the Nintendo GameCube — the counterpart of disr4300 / disr5900 / dismips / disarm.
//
// It decodes every 4 bytes in the selected range as an instruction, making no attempt to
// tell code from data; use codetracegekko for that.
//
// The input may be a GameCube disc, a raw .dol executable, or a flat binary. A disc is
// opened, its executable found, and its segments loaded at the addresses they ask for,
// so -base is usually unnecessary: the file already says where it lives. There is no
// symbol table in a DOL — the format has no room for one — so unlike disr5900 this
// listing labels nothing, and the segment boundaries are all the structure there is.
//
// Usage:
//
//	disgekko [-base ADDR] [-skip N] [-start ADDR] [-end ADDR] [-census] IMAGE
//
// All addresses are hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/gekko"
	"retroreverse.com/tools/platform/gc"
)

func main() {
	base := flag.String("base", "0", "CPU address the image is loaded at (hex; ignored for a disc or a DOL)")
	skip := flag.Int("skip", 0, "leading file bytes to drop before -base maps")
	start := flag.String("start", "", "first address to disassemble (hex, default the load address)")
	end := flag.String("end", "", "last address to disassemble (hex, default end of image)")
	textOnly := flag.Bool("text", false, "disassemble only the executable's code segments")
	census := flag.Bool("census", false, "count the mnemonics in the range instead of listing it")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: disgekko [-base A] [-skip N] [-start A] [-end A] [-text] [-census] IMAGE")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *base, *skip, *start, *end, *textOnly, *census); err != nil {
		fmt.Fprintln(os.Stderr, "disgekko:", err)
		os.Exit(1)
	}
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

// load resolves the input to a flat image and the address it starts at, plus the
// executable it came from when there was one.
func load(path, baseS string, skip int) (base uint32, mem []byte, dol *gc.DOL, err error) {
	// A disc first: it names its own executable.
	if d, derr := gc.Open(path); derr == nil {
		defer d.Close()
		dol, err = d.DOL()
		if err != nil {
			return 0, nil, nil, err
		}
		base, mem = dol.Flat()
		fmt.Fprintf(os.Stderr, "disgekko: %s — executable loaded at 0x%08X (%d bytes), entry 0x%08X, %d segments\n",
			d.Header.GameID, base, len(mem), dol.Entry, len(dol.Segments))
		return base, mem, dol, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, nil, err
	}
	// Then a bare DOL.
	if d, derr := gc.ParseDOL(raw); derr == nil && d.Entry != 0 {
		base, mem = d.Flat()
		fmt.Fprintf(os.Stderr, "disgekko: DOL loaded at 0x%08X (%d bytes), entry 0x%08X, %d segments\n",
			base, len(mem), d.Entry, len(d.Segments))
		return base, mem, d, nil
	}
	// Otherwise flat.
	if base, err = hx(baseS); err != nil {
		return 0, nil, nil, fmt.Errorf("bad -base %q", baseS)
	}
	if skip < 0 || skip > len(raw) {
		return 0, nil, nil, fmt.Errorf("bad -skip %d", skip)
	}
	return base, raw[skip:], nil, nil
}

func run(path, baseS string, skip int, startS, endS string, textOnly, census bool) error {
	base, mem, dol, err := load(path, baseS, skip)
	if err != nil {
		return err
	}

	// The ranges to walk. Normally one; with -text, one per code segment, which is what
	// makes "does the decoder cover the game's code?" a question with an answer.
	type span struct {
		name   string
		lo, hi uint32
	}
	var spans []span

	switch {
	case textOnly && dol != nil:
		for _, s := range dol.Segments {
			if s.Text {
				spans = append(spans, span{s.Name(), s.Addr, s.Addr + s.Size - 4})
			}
		}
	default:
		lo, hi := base, base+uint32(len(mem))-4
		if startS != "" {
			if lo, err = hx(startS); err != nil {
				return fmt.Errorf("bad -start %q", startS)
			}
		}
		if endS != "" {
			if hi, err = hx(endS); err != nil {
				return fmt.Errorf("bad -end %q", endS)
			}
		}
		if lo < base || hi < lo || int(hi-base) >= len(mem) {
			return fmt.Errorf("range 0x%08X-0x%08X lies outside the image at 0x%08X (%d bytes)", lo, hi, base, len(mem))
		}
		spans = append(spans, span{"", lo, hi})
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	counts := map[string]int{}
	var total, unknown int

	for _, sp := range spans {
		if !census && sp.name != "" {
			fmt.Fprintf(w, "\n; ==== %s  0x%08X..0x%08X ====\n", sp.name, sp.lo, sp.hi+4)
		}
		for a := sp.lo &^ 3; a <= sp.hi; a += 4 {
			o := int(a - base)
			in := gekko.Decode(mem[o:], a)
			total++
			counts[in.Mnem]++
			if in.Mnem == ".word" {
				unknown++
			}
			if !census {
				fmt.Fprintf(w, "%08X  %02X %02X %02X %02X  %s\n",
					a, mem[o], mem[o+1], mem[o+2], mem[o+3], in.Text)
			}
		}
	}

	if census {
		type kv struct {
			m string
			n int
		}
		var all []kv
		for m, n := range counts {
			all = append(all, kv{m, n})
		}
		sort.Slice(all, func(i, j int) bool {
			if all[i].n != all[j].n {
				return all[i].n > all[j].n
			}
			return all[i].m < all[j].m
		})
		for _, e := range all {
			fmt.Fprintf(w, "%8d  %.2f%%  %s\n", e.n, 100*float64(e.n)/float64(total), e.m)
		}
	}

	fmt.Fprintf(os.Stderr, "disgekko: %d words, %d distinct mnemonics, %d undecoded (%.4f%%)\n",
		total, len(counts), unknown, 100*float64(unknown)/float64(total))
	return nil
}
