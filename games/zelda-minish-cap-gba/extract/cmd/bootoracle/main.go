// bootoracle boots The Minish Cap in the GBA machine model and reports what
// happened — the platform's dynamic-verification instrument.
//
//	bootoracle [-rom file.gba] [-frames N] [-steps N] [-shot out.png] ...
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/gba"
	"retroreverse.com/tools/platform/gba/gbamachine"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "bootoracle: "+format+"\n", a...)
	os.Exit(2)
}

func parseNum(s string) uint64 {
	s = strings.TrimPrefix(s, "0x")
	if v, err := strconv.ParseUint(s, 16, 64); err == nil && strings.ContainsAny(s, "abcdefABCDEF") {
		return v
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		v, err = strconv.ParseUint(s, 16, 64)
		if err != nil {
			die("bad number %q", s)
		}
	}
	return v
}

func main() {
	rom := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	frames := flag.Uint64("frames", 0, "stop after N more frames")
	steps := flag.String("steps", "40000000", "instruction budget (decimal or 0x hex)")
	shot := flag.String("shot", "", "write the composed frame as a PNG when the run ends")
	keys := flag.String("keys", "", "buttons to hold: a,b,select,start,right,left,up,down,r,l — or FRAME:buttons[,FRAME:buttons...] press scripts")
	showIO := flag.Bool("io", false, "list the I/O registers the run programmed")
	showLog := flag.Bool("log", false, "list the hardware the model did NOT implement")
	savestate := flag.String("savestate", "", "write a snapshot when the run ends")
	loadstate := flag.String("loadstate", "", "start from a snapshot instead of reset")
	trace := flag.Bool("trace", false, "trace instructions (with -tracen)")
	tracen := flag.Int("tracen", 200, "how many instructions to trace")
	traceFrom := flag.String("tracefrom", "", "start tracing at this PC (hex)")
	dump := flag.String("dump", "", "hex-dump memory after the run: ADDR:LEN (hex)")
	bp := flag.String("bp", "", "halting breakpoint PC (hex, comma-separated)")
	irqlog := flag.Bool("irqlog", false, "log every interrupt dispatched")
	snd := flag.Bool("snd", false, "dump the sound block: mixing registers, PSG channels, and the Direct Sound timer/DMA chain")
	wav := flag.String("wav", "", "capture the sound the game's driver mixes and write it as a WAV")
	flag.Parse()

	data, err := os.ReadFile(*rom)
	if err != nil {
		die("%v", err)
	}
	cart, err := gba.Parse(data)
	if err != nil {
		die("%v", err)
	}
	m := gbamachine.New(cart)

	if *loadstate != "" {
		if err := m.LoadState(*loadstate); err != nil {
			die("%v", err)
		}
		fmt.Printf("resumed at frame %d, PC 0x%08X\n", m.Frame(), m.PC())
	}

	if *wav != "" {
		m.AudioCapture(true)
	}

	if *bp != "" {
		for _, s := range strings.Split(*bp, ",") {
			m.AddBreakpoint(uint32(parseNum(s)))
		}
	}
	if *irqlog {
		m.OnIRQ = func(src uint16, handler, ret uint32) {
			fmt.Printf("IRQ %04X -> handler 0x%08X (from 0x%08X, frame %d line %d)\n",
				src, handler, ret, m.Frame(), m.Line())
		}
	}

	// The tracer: from a PC (or from the start), print n instructions.
	if *trace || *traceFrom != "" {
		var armed = *traceFrom == ""
		var from uint32
		if !armed {
			from = uint32(parseNum(*traceFrom))
		}
		left := *tracen
		m.OnStep = func(pc uint32) {
			if !armed {
				if pc != from {
					return
				}
				armed = true
			}
			if left <= 0 {
				return
			}
			left--
			code := m.Snapshot(pc, 4)
			inst := arm.DecodeVariant(code, pc, m.ThumbState(), arm.V4T)
			r := m.Regs()
			fmt.Printf("%08X: %-32s r0=%08X r1=%08X r2=%08X r3=%08X sp=%08X lr=%08X\n",
				pc, inst.Text, r[0], r[1], r[2], r[3], r[13], r[14])
		}
	}

	// Key scripts: either "start,a" (hold for the whole run) or
	// "120:start,240:a" (press for ~20 frames at the given frame numbers).
	type press struct {
		frame uint64
		mask  uint16
	}
	var script []press
	if *keys != "" {
		names := map[string]uint16{"a": 1, "b": 2, "select": 4, "start": 8,
			"right": 16, "left": 32, "up": 64, "down": 128, "r": 256, "l": 512}
		if !strings.Contains(*keys, ":") {
			var hold uint16
			for _, k := range strings.Split(*keys, ",") {
				hold |= names[strings.ToLower(strings.TrimSpace(k))]
			}
			m.SetKeys(hold)
		} else {
			for _, part := range strings.Split(*keys, ",") {
				fs := strings.SplitN(part, ":", 2)
				if len(fs) != 2 {
					die("bad -keys entry %q", part)
				}
				var mask uint16
				for _, k := range strings.Split(fs[1], "+") {
					mask |= names[strings.ToLower(strings.TrimSpace(k))]
				}
				script = append(script, press{parseNum(fs[0]), mask})
			}
			sort.Slice(script, func(i, j int) bool { return script[i].frame < script[j].frame })
		}
	}
	if len(script) > 0 {
		m.OnFrame = func() {
			f := m.Frame()
			var mask uint16
			for _, p := range script {
				// A press is held for 20 frames — a game waits for edges, and a
				// one-frame tap can fall between its polls.
				if f >= p.frame && f < p.frame+20 {
					mask |= p.mask
				}
			}
			m.SetKeys(mask)
		}
	}

	var res gbamachine.Result
	if *frames > 0 {
		res = m.RunFrames(*frames, parseNum(*steps))
	} else {
		res = m.Run(parseNum(*steps), nil)
	}

	fmt.Printf("run: %s\n", res.Reason)
	fmt.Printf("frames %d, steps %d, instrs %d, PC 0x%08X (%s), %s\n",
		res.Frames, res.Steps, m.Instrs(), m.PC(), thumbTag(m), m.Parked())
	ie, ifl, ime := m.IRQState()
	fmt.Printf("IE=%04X IF=%04X IME=%v DISPCNT=%04X\n", ie, ifl, ime, m.Reg(0x000))

	if *showLog {
		for _, l := range m.Log {
			fmt.Println("log:", l)
		}
	}
	if *snd {
		for _, l := range m.SoundState() {
			fmt.Println("snd:", l)
		}
	}
	if *showIO {
		fmt.Println("(io register file: values last written)")
		for _, r := range []uint32{0x000, 0x004, 0x008, 0x00A, 0x00C, 0x00E, 0x050, 0x200} {
			fmt.Printf("  0x%03X = 0x%04X\n", r, m.Reg(r))
		}
	}
	if *dump != "" {
		fs := strings.SplitN(*dump, ":", 2)
		addr, n := uint32(parseNum(fs[0])), uint32(parseNum(fs[1]))
		data := m.Snapshot(addr, n)
		for i := uint32(0); i < n; i += 16 {
			end := i + 16
			if end > n {
				end = n
			}
			fmt.Printf("%08X: % X\n", addr+i, data[i:end])
		}
	}
	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			die("%v", err)
		}
		fmt.Println("shot:", *shot)
	}
	if *wav != "" {
		if err := m.WriteWAV(*wav); err != nil {
			die("%v", err)
		}
		fmt.Printf("wav: %s (%d frames, %.2fs)\n", *wav, m.AudioSamples(),
			float64(m.AudioSamples())/32768)
	}
	if *savestate != "" {
		if err := m.SaveState(*savestate); err != nil {
			die("%v", err)
		}
		fmt.Println("savestate:", *savestate)
	}
	if h, why := m.Halted(); h {
		fmt.Println("HALTED:", why)
		os.Exit(1)
	}
}

func thumbTag(m *gbamachine.Machine) string {
	if m.ThumbState() {
		return "Thumb"
	}
	return "ARM"
}
