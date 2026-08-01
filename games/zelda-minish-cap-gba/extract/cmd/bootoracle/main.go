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

// parseAddr parses an ADDRESS, always as hex. Addresses go through their own
// parser because guessing the base from the digits is a silent-wrong: "3005600"
// contains no letters, so a guessing parser reads it as decimal 3,005,600 and
// dumps 0x2DDCA0 without complaint — a plausible-looking hex dump of entirely
// the wrong memory.
func parseAddr(s string) uint64 {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "$")
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		die("bad address %q (addresses are hex)", s)
	}
	return v
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
	dumpbin := flag.String("dumpbin", "", "write raw memory to a file after the run: ADDR:LEN:PATH (hex addr/len) — for disassembling code the game relocated into IWRAM")
	dump := flag.String("dump", "", "hex-dump memory after the run: ADDR:LEN (hex)")
	bp := flag.String("bp", "", "halting breakpoint PC (hex, comma-separated)")
	irqlog := flag.Bool("irqlog", false, "log every interrupt dispatched")
	shotEvery := flag.Uint64("shotevery", 0, "with -shot: also write a numbered capture every N frames (BASE-000123.png)")
	logpc := flag.String("logpc", "", "non-halting breakpoint: log registers every time this PC is reached (hex, comma-separated)")
	declog := flag.Bool("declog", false, "log every BIOS decompression: SWI, ROM/RAM source, destination, size — the game's own asset loads")
	readwatch := flag.String("readwatch", "", "log READS of an address range ADDR:LEN (hex), with the PC — for finding the code that consumes a table")
	watch := flag.String("watch", "", "log writes to an address range ADDR:LEN (hex), with the PC that made them")
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

	if *declog {
		m.OnDecompress = func(swi, src, dst uint32, size int, pc, lr uint32) {
			kind := map[uint32]string{0x11: "LZ77", 0x12: "LZ77vram", 0x13: "Huffman", 0x14: "RLE", 0x15: "RLEvram"}[swi]
			fmt.Printf("dec: %-8s src=%08X dst=%08X size=%6d  (caller lr=%08X, frame %d)\n",
				kind, src, dst, size, lr, m.Frame())
		}
	}
	if *readwatch != "" {
		fs := strings.SplitN(*readwatch, ":", 2)
		if len(fs) != 2 {
			die("bad -readwatch (want ADDR:LEN)")
		}
		lo := uint32(parseAddr(fs[0]))
		hi := lo + uint32(parseAddr(fs[1]))
		seen := map[uint32]int{}
		m.OnRead = func(addr uint32, v byte, pc uint32) {
			if addr < lo || addr >= hi {
				return
			}
			if seen[pc]++; seen[pc] == 1 {
				fmt.Printf("readwatch: first read of %08X from pc %08X (frame %d)\n", addr, pc, m.Frame())
			}
		}
		defer func() {
			for pc, n := range seen {
				fmt.Printf("readwatch: pc %08X read %d bytes in range\n", pc, n)
			}
		}()
	}
	if *watch != "" {
		fs := strings.SplitN(*watch, ":", 2)
		if len(fs) != 2 {
			die("bad -watch (want ADDR:LEN)")
		}
		lo := uint32(parseAddr(fs[0]))
		hi := lo + uint32(parseAddr(fs[1]))
		seen := map[uint32]int{}
		m.OnWrite = func(addr uint32, v byte, pc uint32) {
			if addr < lo || addr >= hi {
				return
			}
			// One line per WRITER, not per write: a tilemap upload is thousands
			// of stores from a handful of instructions, and the useful fact is
			// which instructions.
			if seen[pc]++; seen[pc] == 1 {
				fmt.Printf("watch: first write to %08X from pc %08X (frame %d)\n", addr, pc, m.Frame())
			}
		}
		defer func() {
			type wc struct {
				pc, n int
			}
			var l []wc
			for pc, n := range seen {
				l = append(l, wc{int(pc), n})
			}
			sort.Slice(l, func(i, j int) bool { return l[i].n > l[j].n })
			for i, e := range l {
				if i >= 10 {
					break
				}
				fmt.Printf("watch: pc %08X wrote %d bytes in range\n", uint32(e.pc), e.n)
			}
		}()
	}

	if *bp != "" {
		for _, s := range strings.Split(*bp, ",") {
			m.AddBreakpoint(uint32(parseAddr(s)))
		}
	}
	if *irqlog {
		m.OnIRQ = func(src uint16, handler, ret uint32) {
			fmt.Printf("IRQ %04X -> handler 0x%08X (from 0x%08X, frame %d line %d)\n",
				src, handler, ret, m.Frame(), m.Line())
		}
	}

	if *logpc != "" {
		want := map[uint32]bool{}
		for _, x := range strings.Split(*logpc, ",") {
			want[uint32(parseAddr(x))] = true
		}
		prev := m.OnStep
		m.OnStep = func(pc uint32) {
			if want[pc] {
				r := m.Regs()
				fmt.Printf("logpc %08X:", pc)
				for i := 0; i <= 7; i++ {
					fmt.Printf(" r%d=%08X", i, r[i])
				}
				fmt.Printf(" lr=%08X (frame %d)\n", r[14], m.Frame())
			}
			if prev != nil {
				prev(pc)
			}
		}
	}

	// The tracer: from a PC (or from the start), print n instructions.
	if *trace || *traceFrom != "" {
		var armed = *traceFrom == ""
		var from uint32
		if !armed {
			from = uint32(parseAddr(*traceFrom))
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
			fmt.Printf("%08X: %-30s", pc, inst.Text)
			for i := 0; i < 13; i++ {
				fmt.Printf(" r%d=%08X", i, r[i])
			}
			fmt.Printf(" sp=%08X lr=%08X\n", r[13], r[14])
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
	// Periodic capture and the key script share OnFrame, so they compose.
	var frameHooks []func()
	if *shotEvery > 0 && *shot != "" {
		base := strings.TrimSuffix(*shot, ".png")
		frameHooks = append(frameHooks, func() {
			f := m.Frame()
			if f%*shotEvery != 0 {
				return
			}
			if err := m.Screenshot(fmt.Sprintf("%s-%06d.png", base, f)); err != nil {
				die("%v", err)
			}
		})
	}
	if len(script) > 0 {
		frameHooks = append(frameHooks, func() {
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
		})
	}
	if len(frameHooks) > 0 {
		m.OnFrame = func() {
			for _, h := range frameHooks {
				h()
			}
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
		addr, n := uint32(parseAddr(fs[0])), uint32(parseAddr(fs[1]))
		data := m.Snapshot(addr, n)
		for i := uint32(0); i < n; i += 16 {
			end := i + 16
			if end > n {
				end = n
			}
			fmt.Printf("%08X: % X\n", addr+i, data[i:end])
		}
	}
	if *dumpbin != "" {
		fs := strings.SplitN(*dumpbin, ":", 3)
		if len(fs) != 3 {
			die("bad -dumpbin (want ADDR:LEN:PATH)")
		}
		data := m.Snapshot(uint32(parseAddr(fs[0])), uint32(parseAddr(fs[1])))
		if err := os.WriteFile(fs[2], data, 0o644); err != nil {
			die("%v", err)
		}
		fmt.Printf("dumpbin: %d bytes from 0x%08X -> %s\n", len(data), parseAddr(fs[0]), fs[2])
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
