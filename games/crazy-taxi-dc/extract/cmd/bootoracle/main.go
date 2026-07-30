// bootoracle boots Crazy Taxi on the Dreamcast machine model
// (tools/platform/dc) and reports what happens — the ground-truth instrument
// every reverse-engineering claim in the writeup is checked against.
//
// The flag vocabulary is the repo standard (STANDARDS.md §3):
//
//	bootoracle -image game.cue -steps N [-frames N] [-trace] [-bp A] [-keys ...]
//	           [-shot out.png] [-savestate f] [-loadstate f] [-v] ...
//
// A run ends with the machine's own verdict: where it stopped and why, the
// registers, and under -v the census of everything unmodelled it touched —
// the worklist for the next platform session.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/dc"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

func main() {
	image := flag.String("image", "games/crazy-taxi-dc/image/Crazy Taxi (US).cue", "disc image (.cue, cdrdao TOC)")
	steps := flag.String("steps", "500000000", "instruction budget (decimal, or hex with 0x)")
	frames := flag.Uint64("frames", 0, "run this many VBlank fields instead of the full step budget")
	trace := flag.Bool("trace", false, "print every executed instruction")
	tracen := flag.Uint64("tracen", 200, "cap on -trace lines")
	tracefrom := flag.String("tracefrom", "", "start -trace when PC first reaches this address (hex)")
	var bps, logpcs, watches, rwatches multiFlag
	flag.Var(&bps, "bp", "halting breakpoint (hex address); repeatable")
	flag.Var(&logpcs, "logpc", "non-halting PC log with registers (hex); repeatable")
	flag.Var(&watches, "watch", "write watch ADDR[:LEN] (hex, physical); repeatable")
	flag.Var(&rwatches, "rwatch", "read watch ADDR[:LEN] (hex, physical); repeatable")
	poke := flag.String("poke", "", "write ADDR:VALUE32 after boot/loadstate (hex)")
	keys := flag.String("keys", "", "pad script BUTTON@FIELD[:HOLD],... e.g. start@240,a@600:10")
	shot := flag.String("shot", "", "render the scanout framebuffer to this PNG at the end")
	vramshot := flag.String("vramshot", "", "raw VRAM viewer OFFSET:WxH:out.png (hex offset)")
	dis := flag.String("dis", "", "disassemble ADDR[:N] and exit (hex)")
	dump := flag.String("dump", "", "hex-dump ADDR:LEN and exit (hex)")
	ramraw := flag.String("ramraw", "", "write the 16MB of main RAM to this file after the run")
	vramraw := flag.String("vramraw", "", "write the 8MB of VRAM (64-bit-path layout) to this file after the run")
	aicaregs := flag.Bool("aicaregs", false, "print the AICA register census (which slot words the driver has written) after the run")
	wav := flag.String("wav", "", "capture the AICA's stereo mix to this WAV over the whole run")
	files := flag.Bool("files", false, "list the disc's files and exit")
	gd := flag.Bool("gd", false, "log every GD-ROM read with the file it lands in")
	watchprof := flag.Bool("watchprof", false, "aggregate watch hits into a per-PC histogram instead of printing each")
	c2dlog := flag.Bool("c2dlog", false, "log every ch2 texture-path DMA (RAM src -> VRAM dst, bytes)")
	savestate := flag.String("savestate", "", "write a savestate at the end of the run")
	loadstate := flag.String("loadstate", "", "resume from a savestate instead of booting")
	nospin := flag.Bool("nospin", false, "disable tight-spin detection")
	cpuprofile := flag.String("cpuprofile", "", "write a pprof CPU profile")
	verbose := flag.Bool("v", false, "print the unmodelled-hardware census after the run")
	flag.Parse()

	if err := run(*image, *steps, *frames, *trace, *tracen, *tracefrom, bps, logpcs, watches, rwatches,
		*poke, *keys, *shot, *vramshot, *dis, *dump, *ramraw, *vramraw, *wav, *aicaregs, *files, *gd, *watchprof, *c2dlog, *savestate, *loadstate,
		*nospin, *cpuprofile, *verbose); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

func run(image, stepsS string, frames uint64, trace bool, tracen uint64, tracefrom string,
	bps, logpcs, watches, rwatches multiFlag, poke, keys, shot, vramshot, dis, dump, ramraw, vramraw, wav string,
	aicaregs, files, gd, watchprof, c2dlog bool, savestate, loadstate string, nospin bool, cpuprofile string, verbose bool) error {

	maxSteps, err := strconv.ParseUint(strings.TrimPrefix(stepsS, "0x"), pick(strings.HasPrefix(stepsS, "0x"), 16, 10), 64)
	if err != nil {
		return fmt.Errorf("bad -steps %q", stepsS)
	}

	disc, err := dc.OpenDisc(image)
	if err != nil {
		return err
	}
	if files {
		return disc.Vol.Walk(func(e iso9660.Entry) error {
			fmt.Println(e)
			return nil
		})
	}

	m := dc.NewMachine(disc)
	if loadstate != "" {
		if err := m.LoadStateFile(loadstate); err != nil {
			return err
		}
	} else {
		if err := m.Boot(); err != nil {
			return err
		}
		warnIfScrambled(m)
	}

	if poke != "" {
		a, v, err := pair(poke)
		if err != nil {
			return fmt.Errorf("bad -poke %q", poke)
		}
		m.Write32(a&0x1FFFFFFF, v)
	}
	cfg := dc.RunConfig{Breakpoints: map[uint32]bool{}, NoSpin: nospin}
	for _, s := range bps {
		a, err := hx(s)
		if err != nil {
			return fmt.Errorf("bad -bp %q", s)
		}
		cfg.Breakpoints[a] = true
	}

	logSet := map[uint32]bool{}
	for _, s := range logpcs {
		a, err := hx(s)
		if err != nil {
			return fmt.Errorf("bad -logpc %q", s)
		}
		logSet[a] = true
	}

	tracing := trace && tracefrom == ""
	var traceStart uint32
	if tracefrom != "" {
		if traceStart, err = hx(tracefrom); err != nil {
			return err
		}
	}
	var traced uint64
	if trace || len(logSet) > 0 {
		m.OnStep = func(pc uint32) {
			if tracefrom != "" && !tracing && pc == traceStart {
				tracing = true
			}
			if tracing && traced < tracen {
				fmt.Println(m.Disasm(pc, 1)[0])
				traced++
			}
			if logSet[pc] {
				fmt.Printf("logpc %08X:\n%s", pc, m.RegString())
			}
		}
	}

	if len(watches) > 0 || len(rwatches) > 0 {
		for _, s := range watches {
			r, err := watchRange(s)
			if err != nil {
				return err
			}
			m.WatchW = append(m.WatchW, r)
		}
		for _, s := range rwatches {
			r, err := watchRange(s)
			if err != nil {
				return err
			}
			m.WatchR = append(m.WatchR, r)
		}
		if watchprof {
			type wkey struct {
				pc    uint32
				write bool
			}
			prof := map[wkey]uint64{}
			m.OnWatch = func(write bool, addr, v uint32, size int, pc uint32) {
				prof[wkey{pc, write}]++
			}
			defer func() {
				type row struct {
					k wkey
					n uint64
				}
				rows := make([]row, 0, len(prof))
				for k, n := range prof {
					rows = append(rows, row{k, n})
				}
				sort.Slice(rows, func(i, j int) bool { return rows[i].n > rows[j].n })
				fmt.Printf("watchprof: %d distinct PCs\n", len(rows))
				for _, r := range rows {
					dir := "r"
					if r.k.write {
						dir = "w"
					}
					fmt.Printf("watchprof %s PC %08X x%d\n", dir, r.k.pc, r.n)
				}
			}()
		} else {
			m.OnWatch = func(write bool, addr, v uint32, size int, pc uint32) {
				dir := "r"
				if write {
					dir = "w"
				}
				fmt.Printf("watch %s%d %08X = %08X (PC %08X)\n", dir, size*8, addr, v, pc)
			}
		}
	}

	if c2dlog {
		m.OnC2DTexture = func(src, dst, length uint32) {
			fmt.Printf("c2d tex %08X -> vram %06X len %#x\n", src, dst, length)
		}
	}

	if gd {
		m.OnGDRead = func(fad, count, buf uint32) {
			lba := int(fad) - 150
			name := "?"
			off := 0
			if e, ok := disc.Vol.FileAt(lba); ok {
				name = e.Path
				off = (lba - e.Block) * 2048
			}
			fmt.Printf("gd read FAD %d x%d -> %08X <- %s +0x%X\n", fad, count, buf, name, off)
		}
	}

	if keys != "" {
		if err := installKeys(m, keys); err != nil {
			return err
		}
	}
	if wav != "" {
		m.AudioCapture = true
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			return err
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// -dis and -dump inspect the machine AFTER the run; -steps 0 makes them
	// static probes of the freshly booted (or just-restored) state.
	var r dc.Result
	if frames > 0 {
		r = m.RunFields(frames, maxSteps)
	} else if maxSteps > 0 {
		r = m.Run(maxSteps, cfg)
	}
	if dis != "" {
		addr, n := uint32(0), 16
		if i := strings.IndexByte(dis, ':'); i >= 0 {
			if addr, err = hx(dis[:i]); err != nil {
				return err
			}
			n, _ = strconv.Atoi(dis[i+1:])
		} else if addr, err = hx(dis); err != nil {
			return err
		}
		for _, l := range m.Disasm(addr, n) {
			fmt.Println(l)
		}
		return nil
	}
	if vramraw != "" {
		if err := os.WriteFile(vramraw, m.VRAM, 0o644); err != nil {
			return err
		}
	}
	if ramraw != "" {
		if err := os.WriteFile(ramraw, m.RAM, 0o644); err != nil {
			return err
		}
		fmt.Println("ramraw:", ramraw)
	}
	if aicaregs {
		fmt.Print(m.AICARegCensus())
	}
	if dump != "" {
		a, n, err := pair(dump)
		if err != nil {
			return fmt.Errorf("bad -dump %q", dump)
		}
		b := m.ReadBytes(a, int(n))
		for i := 0; i < len(b); i += 16 {
			end := i + 16
			if end > len(b) {
				end = len(b)
			}
			fmt.Printf("%08X  % X\n", a+uint32(i), b[i:end])
		}
		return nil
	}
	fmt.Println(r)
	fmt.Print(m.RegString())
	if tx := m.CPU.SerialTX; len(tx) > 0 {
		fmt.Printf("serial (%d bytes):\n%s\n", len(tx), strings.ToValidUTF8(string(tx), "."))
	}

	if shot != "" {
		img, err := m.RenderFB()
		if err != nil {
			fmt.Println("shot:", err) // honest: no frame is a fact, not a black PNG
		} else if err := writePNG(shot, img); err != nil {
			return err
		} else {
			fmt.Println("shot:", shot)
		}
	}
	if vramshot != "" {
		if err := doVRAMShot(m, vramshot); err != nil {
			return err
		}
	}
	if savestate != "" {
		if err := m.SaveStateFile(savestate); err != nil {
			return err
		}
		fmt.Println("savestate:", savestate)
	}
	if wav != "" {
		if err := m.WriteWAV(wav); err != nil {
			fmt.Println("wav:", err) // honest: silence is a fact, not an empty file
		} else {
			fmt.Printf("wav: %s (%s)\n", wav, m.AudioSummary())
		}
	}
	if verbose {
		census := m.Census()
		sort.Strings(census)
		fmt.Printf("census (%d):\n", len(census))
		for _, c := range census {
			fmt.Println(" ", c)
		}
	}
	return nil
}

// warnIfScrambled decodes the entry's first halfwords: a dump whose boot
// binary is scrambled (some .cdi rips) produces mostly .hword there.
func warnIfScrambled(m *dc.Machine) {
	bad := 0
	for i, l := range m.Disasm(0x8C010000, 8) {
		_ = i
		if strings.Contains(l, ".hword") {
			bad++
		}
	}
	if bad > 2 {
		fmt.Fprintln(os.Stderr, "bootoracle: entry code does not decode as SH-4 — possibly a scrambled 1ST_READ.BIN dump")
	}
}

// installKeys arms the pad script: BUTTON@FIELD[:HOLD], hold defaulting to 30
// fields. "l"/"r" pull the analog triggers; "jx<0-255>"/"jy<0-255>" set the
// joystick axes for the window (0x80 centered otherwise, 0 = left/up).
func installKeys(m *dc.Machine, script string) error {
	type ev struct {
		btn        uint16
		trig       byte // 'l', 'r' or 0
		joy        byte // 'x', 'y' or 0
		joyVal     uint8
		from, till uint64
	}
	var evs []ev
	for _, part := range strings.Split(script, ",") {
		at := strings.IndexByte(part, '@')
		if at < 0 {
			return fmt.Errorf("bad -keys entry %q (want BUTTON@FIELD[:HOLD])", part)
		}
		name := strings.ToLower(part[:at])
		rest := part[at+1:]
		hold := uint64(30)
		if i := strings.IndexByte(rest, ':'); i >= 0 {
			h, err := strconv.ParseUint(rest[i+1:], 10, 64)
			if err != nil {
				return fmt.Errorf("bad hold in %q", part)
			}
			hold, rest = h, rest[:i]
		}
		field, err := strconv.ParseUint(rest, 10, 64)
		if err != nil {
			return fmt.Errorf("bad field in %q", part)
		}
		e := ev{from: field, till: field + hold}
		switch {
		case name == "l" || name == "r":
			e.trig = name[0]
		case strings.HasPrefix(name, "jx") || strings.HasPrefix(name, "jy"):
			v, err := strconv.ParseUint(name[2:], 10, 8)
			if err != nil {
				return fmt.Errorf("bad joystick value in %q (want jx<0-255>)", part)
			}
			e.joy, e.joyVal = name[1], uint8(v)
		default:
			// dc.PadButton is the one table both this script parser and the
			// frame debugger's keyer resolve names through.
			b, ok := dc.PadButton(name)
			if !ok {
				return fmt.Errorf("unknown button %q", name)
			}
			e.btn = b
		}
		evs = append(evs, e)
	}
	prev := m.OnDisplay
	m.OnDisplay = func(field uint64) {
		if prev != nil {
			prev(field)
		}
		buttons := uint16(0xFFFF)
		var lt, rt uint8
		jx, jy := uint8(0x80), uint8(0x80)
		for _, e := range evs {
			if field >= e.from && field < e.till {
				switch {
				case e.trig == 'l':
					lt = 255
				case e.trig == 'r':
					rt = 255
				case e.joy == 'x':
					jx = e.joyVal
				case e.joy == 'y':
					jy = e.joyVal
				default:
					buttons &^= e.btn
				}
			}
		}
		m.Pad.Buttons, m.Pad.LT, m.Pad.RT = buttons, lt, rt
		m.Pad.JoyX, m.Pad.JoyY = jx, jy
	}
	return nil
}

func pair(s string) (uint32, uint32, error) {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return 0, 0, fmt.Errorf("want A:B")
	}
	a, err1 := hx(s[:i])
	b, err2 := hx(s[i+1:])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("bad hex in %q", s)
	}
	return a, b, nil
}

func watchRange(s string) (dc.WatchRange, error) {
	a, l := s, "4"
	if i := strings.IndexByte(s, ':'); i >= 0 {
		a, l = s[:i], s[i+1:]
	}
	addr, err1 := hx(a)
	ln, err2 := hx(l)
	if err1 != nil || err2 != nil {
		return dc.WatchRange{}, fmt.Errorf("bad -watch %q", s)
	}
	return dc.WatchRange{Start: addr & 0x1FFFFFFF, Len: ln}, nil
}

func doVRAMShot(m *dc.Machine, spec string) error {
	parts := strings.SplitN(spec, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf("bad -vramshot %q (want OFF:WxH:out.png)", spec)
	}
	off, err := hx(parts[0])
	if err != nil {
		return err
	}
	var w, h int
	if _, err := fmt.Sscanf(parts[1], "%dx%d", &w, &h); err != nil {
		return fmt.Errorf("bad -vramshot size %q", parts[1])
	}
	img, err := m.RenderVRAM(off, w, h)
	if err != nil {
		return err
	}
	return writePNG(parts[2], img)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func pick(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
