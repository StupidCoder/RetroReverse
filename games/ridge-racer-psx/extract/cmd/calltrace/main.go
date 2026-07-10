// calltrace records the dynamic call tree under a root function for one (or a
// few) invocations — the fastest way to see a per-frame pipeline's structure.
// It watches jal/jalr instructions from OnStep, tracking depth by expected
// return addresses.
//
// Usage: calltrace -image DISC -load STATE [-press ...] -root 8001BD80 [-n 1]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/psx"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "calltrace: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	load := flag.String("load", "", "machine savestate to restore")
	press := flag.String("press", "", "pad script")
	rootS := flag.String("root", "8001BD80", "root function address (hex)")
	n := flag.Int("n", 1, "how many invocations to trace")
	window := flag.Uint64("window", 30_000_000, "max steps")
	maxDepth := flag.Int("depth", 4, "max call depth to print")
	flag.Parse()
	var root uint32
	fmt.Sscanf(*rootS, "%x", &root)

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	_, exe, err := vol.BootEXE()
	if err != nil {
		die("%v", err)
	}
	m := psx.NewMachine()
	m.SetDisc(vol)
	m.ISRHandler = 0x8004DF48
	if *press != "" {
		script, err := psx.ParsePress(*press)
		if err != nil {
			die("press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)
	if err := m.LoadStateFile(*load); err != nil {
		die("load state: %v", err)
	}

	read32 := func(a uint32) uint32 {
		return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 |
			uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
	}

	var (
		active  bool
		stack   []uint32 // expected return addresses
		invoked int
	)
	m.OnStep = func(mm *psx.Machine, pc uint32) {
		if !active {
			if pc == root && invoked < *n {
				active = true
				invoked++
				stack = stack[:0]
				fmt.Printf("=== invocation %d (step %d) a0=%08X ===\n", invoked, mm.CPU.Steps, mm.CPU.Reg(4))
			}
			return
		}
		// Return?
		for len(stack) > 0 && pc == stack[len(stack)-1] {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 && pc != root {
			// look at instruction only when shallow enough
		}
		w := read32(pc)
		op := w >> 26
		var target uint32
		isCall := false
		if op == 3 { // jal
			target = (pc & 0xF0000000) | (w&0x03FFFFFF)<<2
			isCall = true
		} else if op == 0 && (w&0x3F) == 9 { // jalr
			target = mm.CPU.Reg((w >> 21) & 31)
			isCall = true
		}
		if isCall {
			depth := len(stack) + 1
			if depth <= *maxDepth {
				for i := 0; i < depth; i++ {
					fmt.Print("  ")
				}
				kind := "jal"
				if op == 0 {
					kind = "jalr"
				}
				fmt.Printf("%s %08X (from %08X, a0=%08X a1=%08X a2=%08X)\n",
					kind, target, pc, mm.CPU.Reg(4), mm.CPU.Reg(5), mm.CPU.Reg(6))
			}
			stack = append(stack, pc+8)
			return
		}
		// Root return: jr $ra at depth 0
		if len(stack) == 0 && op == 0 && (w&0x3F) == 8 && (w>>21)&31 == 31 {
			active = false
			fmt.Println("=== return ===")
			if invoked >= *n {
				mm.Halted = true
				mm.HaltReason = "calltrace done"
			}
		}
	}
	res := m.Run(*window)
	fmt.Fprintln(os.Stderr, res)
}
