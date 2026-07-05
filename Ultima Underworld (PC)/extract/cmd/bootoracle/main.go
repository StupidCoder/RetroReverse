// Command bootoracle loads UW.EXE into the tools/x86 execution core and runs
// it as a DOS program — the Ultima Underworld oracle. It reports how far the
// boot gets: where it halts (program exit, an unimplemented opcode, a spin, or
// the step cap), the INT 21h services it used, the files it opened, and the
// final register/segment state. Use it to follow the C-runtime startup through
// the indirect handoff into the game and toward the overlay manager (Part III).
//
// Usage:
//
//	go run ./cmd/bootoracle [-game ../game] [-steps N] [-trace] [-log]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/x86"
	"ultimaunderworld/extract/uw"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder (contains UW.EXE)")
	steps := flag.Uint64("steps", 20_000_000, "maximum instructions to execute")
	trace := flag.Bool("trace", false, "print a PC trace line for the first -tracen instructions")
	tracen := flag.Uint64("tracen", 200, "number of instructions to trace with -trace")
	showLog := flag.Bool("log", false, "print the full DOS event log")
	dump := flag.String("dump", "", "after the run, hex-dump SEG:OFF:LEN (hex), e.g. 5C4B:0040:0020")
	dis := flag.String("dis", "", "after the run, disassemble SEG:OFF:LEN (hex) from live (relocated) memory")
	irq := flag.Bool("irq", false, "inject periodic timer IRQ0 (recommended: drives the frame waits, cutscenes and menus)")
	bp := flag.String("bp", "", "halt when execution reaches SEG:OFF (hex), e.g. 0FD5:010D")
	bpal := flag.Int("bpal", -1, "with -bp, only halt when AL equals this value (decimal; -1 = any)")
	watch := flag.String("watch", "", "log writes to SEG:OFF[:LEN] (hex)")
	shot := flag.String("shot", "", "after the run, write the mode-13h screen (A000 + DAC palette) as PNG")
	keys := flag.String("keys", "", "script keyboard input via IRQ1, e.g. \"down,down,enter,wait:40,enter\" (implies -irq)")
	flag.Parse()

	var bpSeg, bpOff uint32
	bpSet := false
	if *bp != "" {
		fmt.Sscanf(*bp, "%x:%x", &bpSeg, &bpOff)
		bpSet = true
	}

	exe := filepath.Join(*game, "UW.EXE")
	m, err := uw.LoadEXE(exe, *game)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
	m.EnableIRQ = *irq
	if *keys != "" {
		ev, err := uw.ParseKeys(*keys)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bootoracle:", err)
			os.Exit(1)
		}
		m.SetKeys(ev)
		m.EnableIRQ = true // key pacing rides the timer tick
	}
	if *watch != "" {
		var s, o, l uint32
		if n, _ := fmt.Sscanf(*watch, "%x:%x:%x", &s, &o, &l); n < 3 {
			l = 1
		}
		m.WatchAddr = (s<<4 + o) & 0xFFFFF
		m.WatchLen = l
	}
	c := m.CPU
	fmt.Printf("loaded %s: entry CS:IP=%04X:%04X SS:SP=%04X:%04X DS=%04X (irq=%v)\n\n",
		exe, c.Seg[x86.CS], c.IP, c.Seg[x86.SS], c.Reg16(x86.SP), c.Seg[x86.DS], *irq)

	// Optional short instruction trace at the entry.
	if *trace {
		var n uint64
		c.OnStep = func(cpu *x86.CPU) {
			if n < *tracen {
				lin := (uint32(cpu.Seg[x86.CS]) << 4) + cpu.IP
				in := x86.Decode(m.Mem[lin&0xFFFFF:], cpu.IP)
				fmt.Printf("%04X:%04X  %s\n", cpu.Seg[x86.CS], cpu.IP, in.Text)
			}
			n++
		}
	}

	// Spin detection + a ring buffer of the last executed PCs, so a step-cap or
	// spin stop can show exactly what the CPU was doing.
	spin := newSpinDetector(c)
	const ringSize = 400
	ring := make([]uint32, ringSize) // packed CS<<16 | IP
	ri := 0
	zeroRun := 0
	prevHook := c.OnStep
	c.OnStep = func(cpu *x86.CPU) {
		if prevHook != nil {
			prevHook(cpu)
		}
		ring[ri%ringSize] = uint32(cpu.Seg[x86.CS])<<16 | (cpu.IP & 0xFFFF)
		ri++
		if bpSet && uint32(cpu.Seg[x86.CS]) == bpSeg && (cpu.IP&0xFFFF) == bpOff &&
			(*bpal < 0 || cpu.Reg8(x86.AL) == byte(*bpal)) {
			cpu.Halt("breakpoint at %04X:%04X (AX=%04X BX=%04X)", cpu.Seg[x86.CS], cpu.IP, cpu.Reg16(x86.AX), cpu.Reg16(x86.BX))
		}
		// Runaway detector: executing a run of 00 00 bytes means the CPU jumped
		// into zero-filled memory — halt so the ring shows the culprit transfer.
		l := (uint32(cpu.Seg[x86.CS]) << 4) + (cpu.IP & 0xFFFF)
		if m.Mem[l&0xFFFFF] == 0 && m.Mem[(l+1)&0xFFFFF] == 0 {
			zeroRun++
			if zeroRun > 6 {
				cpu.Halt("runaway: executing zero-filled memory at %04X:%04X", cpu.Seg[x86.CS], cpu.IP)
			}
		} else {
			zeroRun = 0
		}
		spin.check(cpu)
	}

	c.Run(*steps)

	// Print the tail of the execution ring (deduplicated consecutive repeats).
	fmt.Printf("\n== last instructions ==\n")
	start := ri - ringSize
	if start < 0 {
		start = 0
	}
	var lastLine string
	for i := start; i < ri; i++ {
		v := ring[i%ringSize]
		cs, ip := uint16(v>>16), v&0xFFFF
		l := (uint32(cs) << 4) + ip
		in := x86.Decode(m.Mem[l&0xFFFFF:], ip)
		line := fmt.Sprintf("%04X:%04X  %s", cs, ip, in.Text)
		if line != lastLine {
			fmt.Println("  " + line)
			lastLine = line
		}
	}

	fmt.Printf("\n== stopped after %d instructions (%d were 386 0F/66/67 ops) ==\n", c.Steps, c.Ext386)
	switch {
	case m.Terminated:
		fmt.Printf("reason: program terminated (exit code %d)\n", m.ExitCode)
	case c.Halted:
		fmt.Printf("reason: %s\n", c.HaltReason)
	default:
		fmt.Printf("reason: step cap reached (still running at %04X:%04X)\n", c.Seg[x86.CS], c.IP)
	}

	fmt.Printf("\nfinal state:\n")
	fmt.Printf("  CS:IP=%04X:%04X  SS:SP=%04X:%04X  DS=%04X ES=%04X\n",
		c.Seg[x86.CS], c.IP, c.Seg[x86.SS], c.Reg16(x86.SP), c.Seg[x86.DS], c.Seg[x86.ES])
	fmt.Printf("  AX=%04X BX=%04X CX=%04X DX=%04X SI=%04X DI=%04X BP=%04X\n",
		c.Reg16(x86.AX), c.Reg16(x86.BX), c.Reg16(x86.CX), c.Reg16(x86.DX),
		c.Reg16(x86.SI), c.Reg16(x86.DI), c.Reg16(x86.BP))

	fmt.Printf("\nINT 21h services used: ")
	if s := m.IntSummary(); len(s) > 0 {
		for _, e := range s {
			fmt.Printf("%s  ", e)
		}
	}
	fmt.Println()

	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			fmt.Fprintln(os.Stderr, "screenshot:", err)
		} else {
			fmt.Printf("\nscreenshot written to %s\n", *shot)
		}
	}

	if *dump != "" {
		var seg, off, ln uint32
		if _, err := fmt.Sscanf(*dump, "%x:%x:%x", &seg, &off, &ln); err == nil {
			base := (seg<<4 + off) & 0xFFFFF
			fmt.Printf("\n== dump %04X:%04X len %X ==\n", seg, off, ln)
			for i := uint32(0); i < ln; i += 16 {
				fmt.Printf("%04X:%04X ", seg, off+i)
				for j := uint32(0); j < 16 && i+j < ln; j++ {
					fmt.Printf("%02X ", m.Mem[(base+i+j)&0xFFFFF])
				}
				fmt.Println()
			}
		}
	}

	if *dis != "" {
		var seg, off, ln uint32
		if _, err := fmt.Sscanf(*dis, "%x:%x:%x", &seg, &off, &ln); err == nil {
			base := (seg<<4 + off) & 0xFFFFF
			fmt.Printf("\n== disasm %04X:%04X len %X (live memory) ==\n", seg, off, ln)
			end := base + ln
			if end > uint32(len(m.Mem)) {
				end = uint32(len(m.Mem))
			}
			for _, l := range x86.Disassemble(m.Mem[base:end], off) {
				fmt.Println("  " + l)
			}
		}
	}

	if *showLog {
		fmt.Printf("\n== DOS event log (%d) ==\n", len(m.Log))
		for _, l := range m.Log {
			fmt.Println(l)
		}
	} else if len(m.Log) > 0 {
		fmt.Printf("\nlast events:\n")
		start := 0
		if len(m.Log) > 15 {
			start = len(m.Log) - 15
		}
		for _, l := range m.Log[start:] {
			fmt.Println("  " + l)
		}
	}
}

// spinDetector stops the CPU when the PC revisits the same address too many
// times without the surrounding window advancing — a stand-in for hardware
// polling loops the oracle can't satisfy.
type spinDetector struct {
	cpu       *x86.CPU
	lastPC    uint32
	sameCount int
}

func newSpinDetector(c *x86.CPU) *spinDetector { return &spinDetector{cpu: c} }

func (s *spinDetector) check(c *x86.CPU) {
	pc := (uint32(c.Seg[x86.CS]) << 4) + c.IP
	if pc == s.lastPC {
		s.sameCount++
		if s.sameCount > 5_000_000 {
			c.Halt("spin detected at %04X:%04X (%d repeats)", c.Seg[x86.CS], c.IP, s.sameCount)
		}
		return
	}
	s.lastPC = pc
	s.sameCount = 0
}
