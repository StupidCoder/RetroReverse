// codetracearm60 is a recursive-descent code tracer for ARM60 (big-endian,
// ARMv3) images. Starting from one or more entry points it follows every
// statically reachable path — branches, calls (and their fall-through), returns
// — and prints the discovered instructions in address order, marking gaps as
// data. It mirrors tools/cmd/codetracearm for the DS, minus Thumb.
//
//	codetracearm60 -base 0x0 -entry 0x80 LaunchMe.code
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/arm60"
)

func main() {
	base := flag.Uint64("base", 0, "load address of the first byte")
	entries := flag.String("entry", "", "comma-separated entry addresses (hex ok, e.g. 0x80,0x120)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: codetracearm60 -entry A[,B...] [-base A] file")
		os.Exit(2)
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "codetracearm60:", err)
		os.Exit(1)
	}
	baseAddr := uint32(*base)
	end := baseAddr + uint32(len(raw))

	at := func(addr uint32) ([]byte, bool) {
		if addr < baseAddr || addr >= end {
			return nil, false
		}
		return raw[addr-baseAddr:], true
	}

	seen := map[uint32]arm60.Inst{}
	var work []uint32
	for _, s := range strings.Split(*entries, ",") {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad entry %q\n", s)
			os.Exit(2)
		}
		work = append(work, uint32(v))
	}
	if len(work) == 0 {
		work = append(work, baseAddr) // default: the start
	}

	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		for {
			if _, done := seen[pc]; done {
				break
			}
			code, ok := at(pc)
			if !ok {
				break
			}
			in := arm60.Decode(code, pc)
			seen[pc] = in

			if in.HasTarget && in.Target >= baseAddr && in.Target < end {
				work = append(work, in.Target)
			}
			switch in.Flow {
			case arm60.FlowJump, arm60.FlowReturn, arm60.FlowIndJump, arm60.FlowStop:
				pc = 0
			default: // Seq, Branch, Call — fall through
				pc = in.Addr + uint32(in.Len)
				continue
			}
			break
		}
	}

	addrs := make([]uint32, 0, len(seen))
	for a := range seen {
		addrs = append(addrs, a)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

	var prev uint32
	first := true
	for _, a := range addrs {
		if !first && a != prev {
			fmt.Printf("        ; --- gap (data) 0x%08X..0x%08X ---\n", prev, a)
		}
		in := seen[a]
		fmt.Printf("%08X  %s\n", a, in.Text)
		prev = a + uint32(in.Len)
		first = false
	}
	fmt.Fprintf(os.Stderr, "traced %d instructions\n", len(seen))
}
