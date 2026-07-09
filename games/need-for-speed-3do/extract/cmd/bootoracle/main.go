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
	"image/png"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	prog := flag.String("prog", "LaunchMe", "AIF program path within the disc")
	file := flag.String("f", "", "load a standalone AIF file instead of -image/-prog")
	steps := flag.Uint64("steps", 200000, "max instructions to run")
	trace := flag.Bool("trace", false, "print each executed instruction")
	tracen := flag.Uint64("tracen", 0, "print only the first N executed instructions")
	hot := flag.Bool("hot", false, "profile the most-executed instruction addresses")
	breakAt := flag.Uint64("bp", 0, "breakpoint: log lr + r0-r3/r12 each time PC == this address")
	spinbreak := flag.Bool("spinbreak", false, "poke past flag spin-waits (exploration; advances PC, not OS state)")
	fbOut := flag.String("shot", "", "after the run, capture the VRAM framebuffer (320x240 RGB555) to this PNG")
	fbBase := flag.Uint64("fbbase", 0x200000, "framebuffer base address in VRAM")
	watch := flag.Uint64("watch", 0, "log writes to [watch, watch+watchlen) with the writing PC")
	watchLen := flag.Uint64("watchlen", 4, "byte span for -watch")
	dump := flag.Uint64("dump", 0, "after the run, dump memory words at [dump, dump+dumplen)")
	dumpLen := flag.Uint64("dumplen", 0x40, "byte span for -dump")
	vblMirror := flag.Uint64("vblmirror", 0x42734, "game global the VBL manager keeps at the elapsed-field count (0 = off)")
	stall := flag.Int("stall", 1, "deadlock-guard tolerance multiplier (raise for programs with settled main loops)")
	movies := flag.Bool("movies", false, "let the game open .stream movies (FMV subsystem not modelled yet: crashes in the movie player)")
	pad := flag.String("pad", "", "control-pad script: STEP:btn+btn,STEP:0,... (btns: a b c start x up down left right ls rs; 0=release)")
	celDebugAt := flag.Uint64("celdebug", 0, "record a per-cel render summary for the frames after this step")
	sportDebug := flag.Bool("sportdebug", false, "log the full IOInfo of every SPORT request")
	perspTint := flag.Bool("persptint", false, "paint perspective cels solid magenta")
	flag.Parse()

	var data []byte
	var err error
	var vol *threedo.Volume
	if *file != "" {
		data, err = os.ReadFile(*file)
	} else if *image != "" {
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
	m.StallTolerance = *stall
	m.NoStreams = !*movies
	m.SportDebug = *sportDebug
	m.PerspTint = *perspTint
	if *pad != "" {
		script, err := parsePadScript(*pad)
		if err != nil {
			die(err)
		}
		m.PadScript = script
	}
	if vol != nil {
		m.SetVolume(vol)
	}
	m.SetVBLMirror(uint32(*vblMirror))
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
	if *celDebugAt != 0 {
		var n uint64
		m.OnStep = func(mm *threedo.Machine, pc uint32) {
			if n++; n >= *celDebugAt {
				mm.CelDebug = true
			}
		}
	}
	var brk []string
	if *breakAt != 0 {
		ba := uint32(*breakAt)
		m.OnStep = func(mm *threedo.Machine, pc uint32) {
			if pc == ba {
				c := mm.CPU
				brk = append(brk, fmt.Sprintf("hit 0x%08X r0=%08X r1=%08X r2=%08X r3=%08X r4=%08X r5=%08X r6=%08X r12=%08X lr=%08X",
					pc, c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3), c.Reg(4), c.Reg(5), c.Reg(6), c.Reg(12), c.Reg(14)))
			}
		}
	}

	var watchHits []string
	if *watch != 0 {
		m.WatchLo = uint32(*watch)
		m.WatchHi = uint32(*watch + *watchLen)
		m.OnWrite = func(addr, val, pc uint32) {
			if len(watchHits) < 2000 {
				watchHits = append(watchHits, fmt.Sprintf("write [0x%08X]=0x%02X from pc=0x%08X", addr, val&0xFF, pc))
			}
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

	if *watch != 0 {
		fmt.Printf("\n--- watch [0x%X,+0x%X) writes (%d) ---\n", *watch, *watchLen, len(watchHits))
		for _, s := range watchHits {
			fmt.Println(" ", s)
		}
	}

	if *dump != 0 {
		fmt.Printf("\n--- memory dump 0x%X..0x%X ---\n", *dump, *dump+*dumpLen)
		rd := func(a uint32) uint32 {
			return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
		}
		for a := uint32(*dump); a < uint32(*dump+*dumpLen); a += 16 {
			fmt.Printf("  %08X: %08X %08X %08X %08X\n", a, rd(a), rd(a+4), rd(a+8), rd(a+12))
		}
	}

	fmt.Printf("VRAM non-zero bytes (first 640KB): %d\n", m.VRAMNonZero(640*1024))
	if *fbOut != "" {
		base := uint32(*fbBase)
		if db := m.DisplayBuffer(); db != 0 && *fbBase == 0x200000 {
			base = db // default to the screen the game is actually showing
			fmt.Printf("capturing DisplayScreen buffer 0x%08X\n", base)
		}
		img := m.CaptureVRAM(base, 320, 240)
		f, err := os.Create(*fbOut)
		if err != nil {
			die(err)
		}
		if err := png.Encode(f, img); err != nil {
			die(err)
		}
		f.Close()
		fmt.Fprintf(os.Stderr, "wrote framebuffer to %s\n", *fbOut)
	}

	if *celDebugAt != 0 {
		fmt.Printf("\n--- last displayed frame's cels (%d) ---\n", len(m.CelFrameLog))
		for _, s := range m.CelFrameLog {
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

// parsePadScript turns "20000000:start,21000000:0,30000000:a+down" into the
// machine's step-scheduled control-pad states.
func parsePadScript(s string) ([]threedo.PadStep, error) {
	bits := map[string]uint32{
		"a": threedo.PadA, "b": threedo.PadB, "c": threedo.PadC,
		"start": threedo.PadStart, "p": threedo.PadStart, "x": threedo.PadX,
		"up": threedo.PadUp, "down": threedo.PadDown,
		"left": threedo.PadLeft, "right": threedo.PadRight,
		"ls": threedo.PadLeftShift, "rs": threedo.PadRightShift, "0": 0,
	}
	var script []threedo.PadStep
	for _, entry := range strings.Split(s, ",") {
		step, names, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok {
			return nil, fmt.Errorf("pad entry %q: want STEP:buttons", entry)
		}
		at, err := strconv.ParseUint(step, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("pad entry %q: %v", entry, err)
		}
		var buttons uint32
		for _, n := range strings.Split(names, "+") {
			bit, ok := bits[strings.ToLower(strings.TrimSpace(n))]
			if !ok {
				return nil, fmt.Errorf("pad entry %q: unknown button %q", entry, n)
			}
			buttons |= bit
		}
		script = append(script, threedo.PadStep{AtStep: at, Buttons: buttons})
	}
	sort.Slice(script, func(i, j int) bool { return script[i].AtStep < script[j].AtStep })
	return script, nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "bootoracle:", err)
	os.Exit(1)
}
