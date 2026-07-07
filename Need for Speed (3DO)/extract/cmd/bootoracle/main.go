// bootoracle loads Need for Speed's LaunchMe (or any 3DO AIF program) into the
// tools/threedo ARM60 machine and runs it, tracing the startup: the executed
// instructions and the Portfolio folio/kernel calls it makes. This is the 3DO
// equivalent of the Ridge Racer PSX bootoracle.
//
//	bootoracle -image "Need for Speed.bin" -prog LaunchMe -steps 200000 -trace
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	prog := flag.String("prog", "LaunchMe", "AIF program path within the disc")
	file := flag.String("f", "", "load a standalone AIF file instead of -image/-prog")
	steps := flag.Uint64("steps", 200000, "max instructions to run")
	trace := flag.Bool("trace", false, "print each executed instruction")
	tracen := flag.Uint64("tracen", 0, "print only the first N executed instructions")
	hot := flag.Bool("hot", false, "profile the most-executed instruction addresses")
	breakAt := flag.Uint64("break", 0, "log lr + r0-r3/r12 each time PC == this address")
	spinbreak := flag.Bool("spinbreak", false, "poke past flag spin-waits (exploration; advances PC, not OS state)")
	flag.Parse()

	var data []byte
	var err error
	if *file != "" {
		data, err = os.ReadFile(*file)
	} else if *image != "" {
		var vol *threedo.Volume
		if data, err = os.ReadFile(*image); err == nil {
			var v *threedo.Volume
			if v, err = threedo.Open(data); err == nil {
				vol = v
				data, err = vol.ReadFile(*prog)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "usage: bootoracle -image DISC -prog LaunchMe | -f FILE [-steps N] [-trace]")
		os.Exit(2)
	}
	if err != nil {
		die(err)
	}

	aif, err := threedo.ParseAIF(data)
	if err != nil {
		die(err)
	}
	fmt.Print(aif.Describe())

	m := threedo.NewMachine()
	m.SpinBreak = *spinbreak
	m.LoadAIF(aif)

	if *trace || *tracen > 0 {
		var n uint64
		limit := *tracen
		m.OnStep = func(mm *threedo.Machine, pc uint32) {
			if limit == 0 || n < limit {
				fmt.Println(" ", mm.DisasmAt(pc))
				n++
			}
		}
	}
	hits := map[uint32]uint64{}
	if *hot {
		m.OnStep = func(mm *threedo.Machine, pc uint32) { hits[pc]++ }
	}
	var brk []string
	if *breakAt != 0 {
		ba := uint32(*breakAt)
		var prev uint32
		m.OnStep = func(mm *threedo.Machine, pc uint32) {
			if pc == ba {
				c := mm.CPU
				brk = append(brk, fmt.Sprintf("hit 0x%08X (from 0x%08X)  r4=%08X r5=%08X r6=%08X r8=%08X r9=%08X",
					pc, prev, c.Reg(4), c.Reg(5), c.Reg(6), c.Reg(8), c.Reg(9)))
			}
			prev = pc
		}
	}

	fmt.Printf("\n--- running (max %d steps) ---\n", *steps)
	res := m.Run(*steps)
	fmt.Printf("stopped: %s  after %d steps, pc=0x%08X\n", res.Reason, res.Steps, res.PC)

	if len(brk) > 0 {
		fmt.Printf("\n--- breakpoint hits at 0x%X (last 12 of %d) ---\n", *breakAt, len(brk))
		for _, s := range brk[max(0, len(brk)-12):] {
			fmt.Println(" ", s)
		}
	}

	if tty := m.TTY(); tty != "" {
		fmt.Printf("\n[TTY]\n%s\n", tty)
	}
	fmt.Printf("\n--- Portfolio folio/kernel calls (%d) ---\n", len(m.KernelCalls))
	seen := map[uint32]int{}
	for _, k := range m.KernelCalls {
		seen[k.Offset]++
	}
	shown := 0
	for _, k := range m.KernelCalls {
		if shown < 24 {
			fmt.Printf("  folio[-0x%X] from 0x%08X  args=%08X %08X %08X %08X\n",
				k.Offset, k.From, k.Args[0], k.Args[1], k.Args[2], k.Args[3])
			shown++
		}
	}
	fmt.Printf("  (%d distinct folio offsets)\n", len(seen))

	fmt.Printf("\n--- tasks ---\n")
	for _, s := range m.TaskSummary() {
		fmt.Println(" ", s)
	}

	fmt.Printf("\n--- kernel SWIs (%d) ---\n", len(m.SWICalls))
	for i, k := range m.SWICalls {
		if i >= 30 {
			break
		}
		fmt.Printf("  SWI 0x%-5X from 0x%08X  args=%08X %08X %08X %08X\n",
			k.Offset, k.From, k.Args[0], k.Args[1], k.Args[2], k.Args[3])
	}
	if *hot {
		type hp struct {
			pc uint32
			n  uint64
		}
		var hs []hp
		for pc, n := range hits {
			hs = append(hs, hp{pc, n})
		}
		sort.Slice(hs, func(i, j int) bool { return hs[i].n > hs[j].n })
		fmt.Printf("\n--- hottest instruction addresses ---\n")
		for i := 0; i < 12 && i < len(hs); i++ {
			fmt.Printf("  0x%08X  x%d   %s\n", hs[i].pc, hs[i].n, m.DisasmAt(hs[i].pc))
		}
	}
	if len(m.Log) > 0 {
		fmt.Printf("\n--- notes ---\n")
		for _, s := range m.Log {
			fmt.Println(" ", s)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "bootoracle:", err)
	os.Exit(1)
}
