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
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/psx"
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
	shot := flag.String("shot", "", "after the run, write the GPU display to this PNG")
	vramdump := flag.String("vramdump", "", "DEBUG: after the run, write full 1024x512 VRAM to this PNG")
	isrS := flag.String("isr", "8004DF48", "vectored-interrupt entry (hex); Ridge Racer's own "+
		"interrupt dispatcher, traced — the retail BIOS installs it via HookEntryInt into a slot "+
		"the HLE BIOS leaves empty. Set 0 to use the (empty) registered chain handler")
	gpuwatch := flag.String("gpuwatch", "", "DEBUG: log GPU writes into VRAM rect x0,y0,x1,y1")
	press := flag.String("press", "", "scripted controller input: comma-separated BUTTON@STEP:HOLD "+
		"entries (e.g. start@380000000:400000,right@390000000:400000). BUTTON is start/select/up/"+
		"down/left/right/cross/circle/triangle/square/l1/r1/l2/r2. STEP is the instruction count to "+
		"press at, HOLD how many instructions to hold. Fills the HLE pad buffer at VBlank cadence")
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
	if isr, err := hx(*isrS); err == nil {
		m.ISRHandler = isr
	}
	if *press != "" {
		script, err := parsePress(*press)
		if err != nil {
			die("bad -press: %v", err)
		}
		m.PadScript = script
	}
	if *gpuwatch != "" {
		var a, b, c, d int
		if n, _ := fmt.Sscanf(*gpuwatch, "%d,%d,%d,%d", &a, &b, &c, &d); n == 4 {
			m.DebugWatchVRAM(a, b, c, d, func(s string) {
				fmt.Fprintf(os.Stderr, "gpu@%d: %s\n", m.CPU.Steps, s)
			})
		}
	}
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

	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			die("screenshot: %v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote frame to %s\n", *shot)
	}
	if *vramdump != "" {
		if err := m.DumpVRAM(*vramdump); err != nil {
			die("vramdump: %v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote VRAM to %s\n", *vramdump)
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

// padNames maps -press button names to their active-low mask bit.
var padNames = map[string]uint16{
	"select": psx.PadSelect, "start": psx.PadStart,
	"up": psx.PadUp, "right": psx.PadRight, "down": psx.PadDown, "left": psx.PadLeft,
	"l2": psx.PadL2, "r2": psx.PadR2, "l1": psx.PadL1, "r1": psx.PadR1,
	"triangle": psx.PadTriangle, "circle": psx.PadCircle, "cross": psx.PadCross, "square": psx.PadSquare,
}

// parsePress turns "start@380000000:400000,right@390000000:400000" into a
// time-ordered pad schedule: each entry presses a button at its step for a hold
// span, then releases. Overlapping holds are OR-combined so chords work.
func parsePress(spec string) ([]psx.PadEvent, error) {
	type edge struct {
		step uint64
		bit  uint16
		down bool
	}
	var edges []edge
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		at := strings.IndexByte(part, '@')
		col := strings.LastIndexByte(part, ':')
		if at < 0 || col < at {
			return nil, fmt.Errorf("want BUTTON@STEP:HOLD, got %q", part)
		}
		bit, ok := padNames[strings.ToLower(part[:at])]
		if !ok {
			return nil, fmt.Errorf("unknown button %q", part[:at])
		}
		step, err := parseCount(part[at+1 : col])
		if err != nil {
			return nil, fmt.Errorf("bad step in %q", part)
		}
		hold, err := parseCount(part[col+1:])
		if err != nil {
			return nil, fmt.Errorf("bad hold in %q", part)
		}
		edges = append(edges, edge{step, bit, true}, edge{step + hold, bit, false})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].step < edges[j].step })
	var script []psx.PadEvent
	held := uint16(0)
	for _, e := range edges {
		if e.down {
			held |= e.bit
		} else {
			held &^= e.bit
		}
		buttons := psx.PadReleased &^ held
		if n := len(script); n > 0 && script[n-1].AtStep == e.step {
			script[n-1].Buttons = buttons // coalesce simultaneous edges
		} else {
			script = append(script, psx.PadEvent{AtStep: e.step, Buttons: buttons})
		}
	}
	return script, nil
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
