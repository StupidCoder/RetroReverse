// go32run loads a DJGPP go32/COFF DOS executable (e.g. Quake shareware's
// quake.exe) into the mode-aware tools/cpu/x86 core in flat 32-bit protected
// mode and runs it — the executing counterpart to go32dump. It is the Phase-A
// bring-up cockpit for the original-Xbox oracle's CPU: booting a DOS-extender
// game exercises the new protected-mode addressing and, as soon as the C runtime
// touches the FPU, surfaces the first x87 op the core must implement. The program
// halts on the first unmodelled opcode, DPMI function, or INT, so each run turns
// a remaining unknown into a concrete, addressed fact.
//
// Usage:
//
//	go run ./cmd/go32run -image ../games/quake-pc/image/quake.exe [-steps N] [-trace]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"

	"retroreverse.com/tools/cpu/x86"
	"retroreverse.com/tools/platform/dos"
)

func main() {
	image := flag.String("image", "", "path to the go32/COFF executable (e.g. quake.exe)")
	game := flag.String("game", "", "path to the game data directory (defaults to the image's folder)")
	steps := flag.Uint64("steps", 5_000_000, "maximum instructions to execute")
	trace := flag.Bool("trace", false, "print a disassembled trace of the first -tracen instructions")
	tracen := flag.Uint64("tracen", 200, "number of instructions to trace with -trace")
	bp := flag.String("bp", "", "halt when EIP reaches this flat linear address (hex)")
	dis := flag.String("dis", "", "after the run, disassemble ADDR:LEN (hex) from live memory")
	dump := flag.String("dump", "", "after the run, hex-dump ADDR:LEN (hex)")
	pngOut := flag.String("png", "", "after the run, write the VGA mode-13h framebuffer (0xA0000, 320x200) to this PNG")
	keys := flag.String("keys", "", "scripted keyboard input (e.g. \"enter,wait:30,space\") injected via IRQ1")
	loadState := flag.String("loadstate", "", "restore a PM savestate before running (skips the boot)")
	saveState := flag.String("savestate", "", "after the run, write a PM savestate to this file")
	showLog := flag.Bool("log", false, "print the full event log")
	flag.Parse()
	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: go32run -image quake.exe [-steps N] [-trace] [-bp ADDR]")
		os.Exit(2)
	}
	dir := *game
	if dir == "" {
		dir = dirOf(*image)
	}

	m, err := dos.LoadGo32(*image, dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "go32run:", err)
		os.Exit(1)
	}
	c := m.CPU

	if *loadState != "" {
		if err := m.LoadStateFile(*loadState); err != nil {
			fmt.Fprintln(os.Stderr, "go32run: -loadstate:", err)
			os.Exit(1)
		}
		fmt.Printf("restored state from %s (EIP=%08X, %d instructions in)\n", *loadState, c.IP, c.Steps)
	}

	if *keys != "" {
		events, err := dos.ParseKeys(*keys)
		if err != nil {
			fmt.Fprintln(os.Stderr, "go32run: -keys:", err)
			os.Exit(2)
		}
		m.SetKeys(events)
	}

	var bpAddr uint32
	bpSet := false
	if *bp != "" {
		fmt.Sscanf(*bp, "%x", &bpAddr)
		bpSet = true
	}

	fmt.Printf("loaded %s: mode=PM  EIP=%08X  ESP=%08X  CS=%04X DS=%04X\n\n",
		*image, c.IP, c.Regs[x86.SP], c.Seg[x86.CS], c.Seg[x86.DS])

	// A ring of the last executed EIPs so a halt can show what led up to it.
	const ringSize = 400
	ring := make([]uint32, ringSize)
	ri := 0
	var traced uint64
	spinPC, spinN := uint32(0xFFFFFFFF), 0

	c.OnStep = func(cpu *x86.CPU) {
		// Deliver scripted keyboard input (IRQ1 through the game's PM INT 9). This
		// runs at an instruction boundary before the step, the way the real-mode CPU
		// drives Machine.onStep; go32run owns OnStep, so the PM machine can't set its
		// own — it exposes PumpInput for the run loop to call here.
		m.PumpInput(cpu)
		// The linear PC is CS's descriptor base plus EIP — not EIP alone: go32
		// executes its real-mode trampolines through code-selector aliases whose
		// base is a DOS-memory block, so a bare-EIP trace decodes the wrong bytes.
		pc := cpu.SegBase[x86.CS] + cpu.IP
		if *trace && traced < *tracen {
			in := x86.Decode32(memAt(m, pc), pc)
			fmt.Printf("%08X  %s\n", pc, in.Text)
			traced++
		}
		ring[ri%ringSize] = pc
		ri++
		if bpSet && pc == bpAddr {
			cpu.Halt("breakpoint at %08X (EAX=%08X EBX=%08X ECX=%08X EDX=%08X ESI=%08X EDI=%08X EBP=%08X ESP=%08X)",
				pc, cpu.Regs[x86.AX], cpu.Regs[x86.BX], cpu.Regs[x86.CX], cpu.Regs[x86.DX],
				cpu.Regs[x86.SI], cpu.Regs[x86.DI], cpu.Regs[x86.BP], cpu.Regs[x86.SP])
			return
		}
		// Cheap spin detector: the same PC re-executing millions of times without
		// progress is a poll we can't satisfy — stop so the ring shows it.
		if pc == spinPC {
			if spinN++; spinN > 5_000_000 {
				cpu.Halt("spin detected at %08X (%d repeats)", pc, spinN)
			}
		} else {
			spinPC, spinN = pc, 0
		}
	}

	c.Run(*steps)

	// --- report ---
	fmt.Printf("== last instructions ==\n")
	start := ri - ringSize
	if start < 0 {
		start = 0
	}
	var last string
	for i := start; i < ri; i++ {
		pc := ring[i%ringSize]
		in := x86.Decode32(memAt(m, pc), pc)
		line := fmt.Sprintf("%08X  %s", pc, in.Text)
		if line != last {
			fmt.Println("  " + line)
			last = line
		}
	}

	fmt.Printf("\n== stopped after %d instructions (%d were 0F/66/67 ops) ==\n", c.Steps, c.Ext386)
	switch {
	case m.Terminated:
		fmt.Printf("reason: program terminated (exit code %d)\n", m.ExitCode)
	case c.Halted:
		fmt.Printf("reason: %s\n", c.HaltReason)
	default:
		fmt.Printf("reason: step cap reached (still running at %08X)\n", c.IP)
	}

	fmt.Printf("\nfinal state:\n")
	fmt.Printf("  EIP=%08X ESP=%08X EBP=%08X  CS=%04X DS=%04X ES=%04X SS=%04X FS=%04X GS=%04X\n",
		c.IP, c.Regs[x86.SP], c.Regs[x86.BP], c.Seg[x86.CS], c.Seg[x86.DS], c.Seg[x86.ES], c.Seg[x86.SS], c.Seg[x86.FS], c.Seg[x86.GS])
	fmt.Printf("  EAX=%08X EBX=%08X ECX=%08X EDX=%08X ESI=%08X EDI=%08X\n",
		c.Regs[x86.AX], c.Regs[x86.BX], c.Regs[x86.CX], c.Regs[x86.DX], c.Regs[x86.SI], c.Regs[x86.DI])

	if len(m.DPMICounts) > 0 {
		fmt.Printf("\nDPMI (INT 31h) functions used:")
		fns := make([]int, 0, len(m.DPMICounts))
		for fn := range m.DPMICounts {
			fns = append(fns, int(fn))
		}
		sort.Ints(fns)
		for _, fn := range fns {
			fmt.Printf(" %04X(%d)", fn, m.DPMICounts[uint16(fn)])
		}
		fmt.Println()
	}
	if len(m.Console) > 0 {
		fmt.Printf("\n== program console output (%d bytes) ==\n%s\n", len(m.Console), m.Console)
	}
	if len(m.DOSCounts) > 0 {
		fmt.Printf("\nDOS (INT 21h) AH functions used:")
		ahs := make([]int, 0, len(m.DOSCounts))
		for ah := range m.DOSCounts {
			ahs = append(ahs, int(ah))
		}
		sort.Ints(ahs)
		for _, ah := range ahs {
			fmt.Printf(" %02X(%d)", ah, m.DOSCounts[byte(ah)])
		}
		fmt.Println()
	}
	if len(m.IntCounts) > 0 {
		fmt.Printf("other INTs (stubbed):")
		for n, cnt := range m.IntCounts {
			fmt.Printf(" %02Xh(%d)", n, cnt)
		}
		fmt.Println()
	}

	if *dis != "" {
		var addr, ln uint32
		if _, err := fmt.Sscanf(*dis, "%x:%x", &addr, &ln); err == nil {
			fmt.Printf("\n== disasm %08X len %X ==\n", addr, ln)
			for _, l := range x86.Disassemble32(memRange(m, addr, ln), addr) {
				fmt.Println("  " + l)
			}
		}
	}
	if *dump != "" {
		var addr, ln uint32
		if _, err := fmt.Sscanf(*dump, "%x:%x", &addr, &ln); err == nil {
			fmt.Printf("\n== dump %08X len %X ==\n", addr, ln)
			for i := uint32(0); i < ln; i += 16 {
				fmt.Printf("%08X ", addr+i)
				for j := uint32(0); j < 16 && i+j < ln; j++ {
					fmt.Printf("%02X ", m.Mem[addr+i+j])
				}
				fmt.Println()
			}
		}
	}

	if *saveState != "" {
		if err := m.SaveStateFile(*saveState); err != nil {
			fmt.Fprintln(os.Stderr, "go32run: -savestate:", err)
		} else {
			fmt.Printf("\nwrote PM savestate to %s\n", *saveState)
		}
	}

	if *pngOut != "" {
		if err := writeMode13PNG(m, *pngOut); err != nil {
			fmt.Fprintln(os.Stderr, "go32run: -png:", err)
		} else {
			fmt.Printf("\nwrote mode-13h framebuffer to %s\n", *pngOut)
		}
	}

	if *showLog {
		fmt.Printf("\n== event log (%d) ==\n", len(m.Log))
		for _, l := range m.Log {
			fmt.Println("  " + l)
		}
	} else if len(m.Log) > 0 {
		fmt.Printf("\nlast events:\n")
		s := 0
		if len(m.Log) > 15 {
			s = len(m.Log) - 15
		}
		for _, l := range m.Log[s:] {
			fmt.Println("  " + l)
		}
	}
}

// memAt returns a slice of live memory starting at flat linear address a, capped
// so a decoder can read a full instruction without running off the end.
func memAt(m *dos.PM, a uint32) []byte {
	end := a + 16
	if end > uint32(len(m.Mem)) {
		end = uint32(len(m.Mem))
	}
	return m.Mem[a:end]
}

func memRange(m *dos.PM, a, ln uint32) []byte {
	end := a + ln
	if end > uint32(len(m.Mem)) {
		end = uint32(len(m.Mem))
	}
	return m.Mem[a:end]
}

// writeMode13PNG renders the VGA mode-13h framebuffer (256-colour, 320x200 at
// linear 0xA0000) to a PNG, colouring each byte through the DAC palette the game
// programmed via ports 0x3C8/0x3C9 (6-bit components, scaled to 8-bit).
func writeMode13PNG(m *dos.PM, path string) error {
	const w, h, fb = 320, 200, 0xA0000
	var pal [256]color.RGBA
	for i := 0; i < 256; i++ {
		pal[i] = color.RGBA{m.Pal[i*3] << 2, m.Pal[i*3+1] << 2, m.Pal[i*3+2] << 2, 255}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, pal[m.Mem[fb+y*w+x]])
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
