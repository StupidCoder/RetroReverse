// bootoracle boots Ridge Racer's executable under the PSX oracle (tools/psx) and
// runs it, exposing the repo's reverse-engineering instrumentation: an
// instruction trace, breakpoints, a "who wrote this address" watch, the game's
// BIOS/TTY output, and a stop reason. It reads the disc image directly, finds
// the boot executable via SYSTEM.CNF, and loads it — no manual extraction step.
//
// Usage:
//
//	bootoracle -image "Ridge Racer (Track 01).bin" [-steps N] [-trace -tracen N]
//	           [-bp ADDR] [-watch ADDR[:LEN]] [-log] [-tty]
//
// The oracle boots the C-runtime and game init and lands in Ridge Racer's CD-
// ready wait loop; -trace shows live execution, and -watch/-bp help map the code
// that produces a given memory structure. Addresses are hex.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/psx"
)

func hx(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return uint32(v), err
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "bootoracle: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	stepsF := flag.String("steps", "30000000", "max instructions to run (hex or decimal)")
	trace := flag.Bool("trace", false, "disassemble each executed instruction")
	tracen := flag.Int("tracen", 200, "with -trace, stop printing after this many instructions")
	bpS := flag.String("bp", "", "breakpoint: stop when PC reaches this address (hex)")
	watchS := flag.String("watch", "", "watch: log who writes ADDR or ADDR:LEN (hex)")
	showLog := flag.Bool("log", false, "print the machine's diagnostic notes")
	showTTY := flag.Bool("tty", true, "print the game's BIOS/TTY output")
	flag.Parse()
	if *image == "" {
		die("need -image")
	}

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	name, exe, err := vol.BootEXE()
	if err != nil {
		die("boot exe: %v", err)
	}
	fmt.Fprintf(os.Stderr, "booting %s: entry 0x%08X, text 0x%08X (%d bytes)\n", name, exe.PC0, exe.TAddr, exe.TSize)

	steps, err := parseCount(*stepsF)
	if err != nil {
		die("bad -steps")
	}

	m := psx.NewMachine()
	m.SetDisc(vol)
	m.LoadEXE(exe)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// Instrumentation wiring.
	var bp uint32
	haveBP := false
	if *bpS != "" {
		if bp, err = hx(*bpS); err != nil {
			die("bad -bp")
		}
		haveBP = true
	}
	if *watchS != "" {
		lo, ln, err := parseWatch(*watchS)
		if err != nil {
			die("bad -watch")
		}
		m.WatchLo, m.WatchHi = lo, lo+ln
		m.OnWrite = func(addr, val, pc uint32) {
			fmt.Fprintf(w, "write [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
		}
	}
	traced := 0
	hitBP := false
	var bpPC uint32
	if *trace || haveBP {
		m.OnStep = func(mm *psx.Machine, pc uint32) {
			if haveBP && !hitBP && (pc&0x1FFFFFFF) == (bp&0x1FFFFFFF) {
				hitBP, bpPC = true, pc
			}
			if *trace && traced < *tracen {
				fmt.Fprintf(w, "%08X  %s\n", pc, strings.TrimSpace(mm.DisasmAt(pc)))
				traced++
			}
		}
	}

	// Run, honouring the breakpoint by capping steps between checks.
	res := runWithBP(m, steps, &hitBP)
	if hitBP {
		res.PC = bpPC
	}

	w.Flush()
	fmt.Fprintf(os.Stderr, "\n%s\n", res)
	fmt.Fprintf(os.Stderr, "instructions: %d\n", m.CPU.Steps)
	fmt.Fprintf(os.Stderr, "BIOS calls: %v\n", m.BiosCalls())
	if *showTTY {
		if tty := m.TTY(); tty != "" {
			fmt.Fprintf(os.Stderr, "TTY:\n%s\n", tty)
		}
	}
	if *showLog {
		for _, l := range m.Log {
			fmt.Fprintln(os.Stderr, "note:", l)
		}
		if cmds := m.CDCommands(); len(cmds) > 0 {
			fmt.Fprintf(os.Stderr, "CD commands: %v\n", cmds)
		}
		for _, l := range m.CDTrace() {
			fmt.Fprintln(os.Stderr, "cd:", l)
		}
	}
}

// runWithBP runs in short bursts so a breakpoint set by OnStep can stop the run
// promptly. Each Run(n) executes up to n more instructions and preserves state.
func runWithBP(m *psx.Machine, steps uint64, hit *bool) psx.Result {
	const burst = 200000
	var res psx.Result
	for remaining := steps; remaining > 0; {
		n := uint64(burst)
		if remaining < n {
			n = remaining
		}
		res = m.Run(n)
		remaining -= n
		if *hit {
			res.Reason, res.PC = "breakpoint", m.CPU.PC
			break
		}
		if !strings.HasPrefix(res.Reason, "budget") {
			break
		}
	}
	res.Steps = m.CPU.Steps
	return res
}

func parseCount(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		v, err := hx(s)
		return uint64(v), err
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseWatch(s string) (lo, ln uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = hx(parts[0]); err != nil {
		return
	}
	ln = 1
	if len(parts) == 2 {
		ln, err = hx(parts[1])
	}
	return
}
