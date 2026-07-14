// bootoracle boots Super Mario 64 DS on the Nintendo DS machine model
// (tools/platform/nds/dsmachine): both CPUs, the display and its timing, DMA, the
// timers, the cartridge, the ARM7's SPI bus, and both graphics engines. It runs the
// game's own code — nothing here knows anything about Mario — and lets us watch what
// it loads, what it computes and what it draws.
//
//	bootoracle -image rom.nds -frames 600 -shot title
//
// The flag vocabulary is the one every oracle in this repo shares (STANDARDS §3),
// plus the instruments the DS wanted: -io to see the hardware the boot programmed,
// -log to see the hardware we did NOT model (an honest list of gaps, not a silence),
// and -logpc to watch a routine run without halting on it.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/dsmachine"
)

// The ARM9 DTCM base this game programs through CP15 (established in Part II). The
// machine needs it up front because the BIOS interrupt vectors live at the top of
// DTCM, and an interrupt delivered to the wrong address is an interrupt the game
// never takes.
const dtcm9 = 0x023C0000

type addrList []uint32

func (a *addrList) String() string { return "" }
func (a *addrList) Set(s string) error {
	v, err := parseAddr(s)
	if err != nil {
		return err
	}
	*a = append(*a, v)
	return nil
}

func main() {
	image := flag.String("image", "", "ROM image (.nds)")
	steps := flag.String("steps", "0", "instruction budget (hex or decimal; 0 = no limit)")
	frames := flag.Uint64("frames", 0, "stop after N frames (a graphics workload is measured in frames)")
	quantum := flag.Int("quantum", 64, "ARM9 instructions per scheduler quantum")
	shot := flag.String("shot", "", "write both screens as PNG (BASE_top.png, BASE_bottom.png)")
	shotEvery := flag.Uint64("shotevery", 0, "with -shot: also write a numbered capture every N frames")
	keys := flag.String("keys", "", "an input script FILE, or buttons to hold (a,b,start,…). A game waits for a press EDGE, so a held button is usually not enough — see dsmachine/script.go")
	touch := flag.String("touch", "", "hold the stylus at X,Y for the whole run")
	showIO := flag.Bool("io", false, "list the I/O registers the run programmed")
	showLog := flag.Bool("log", false, "list the hardware the model did NOT implement")
	gfx := flag.Bool("gfx", false, "dump the graphics state: both engines, VRAM banks, the 3D engine's frame")
	gxdump := flag.Bool("gxdump", false, "histogram of the 3D commands the geometry engine executed")
	cardlog := flag.Bool("cardlog", false, "log every cartridge transfer: the map of what the game loads, when")
	rtshot := flag.String("rtshot", "", "write the 3D engine's own render target as a PNG, before the 2D engine composites it")
	savestate := flag.String("savestate", "", "write a snapshot of the machine when the run ends")
	loadstate := flag.String("loadstate", "", "start from a snapshot instead of from reset")
	trace := flag.Bool("trace", false, "trace instructions")
	tracen := flag.Int("tracen", 200, "with -trace: how many instructions to trace")
	traceFrom := flag.String("tracefrom", "", "start tracing when this ARM9 address is first reached")
	dump := flag.String("dump", "", "hex-dump memory after the run: ADDR:LEN")
	var bps, logpcs addrList
	flag.Var(&bps, "bp", "halting breakpoint (ARM9 address; repeatable)")
	flag.Var(&logpcs, "logpc", "non-halting breakpoint: log registers and continue (repeatable)")
	flag.Parse()

	romPath := *image
	if romPath == "" && flag.NArg() == 1 {
		romPath = flag.Arg(0)
	}
	if romPath == "" {
		fmt.Fprintln(os.Stderr, "usage: bootoracle -image rom.nds [-frames N] [-shot BASE] ...")
		os.Exit(2)
	}
	data, err := os.ReadFile(romPath)
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(data)
	if err != nil {
		die(err)
	}

	budget, err := parseUint(*steps)
	if err != nil {
		die(fmt.Errorf("-steps: %v", err))
	}
	if budget == 0 {
		budget = 1 << 62
	}

	m := dsmachine.New(rom, dtcm9)

	// A cold boot to the title screen is over a billion scheduler steps. Restoring one
	// is the difference between asking a question of the title screen in a second and
	// asking it in a minute, which in practice is the difference between asking it and
	// not bothering.
	if *loadstate != "" {
		if err := m.LoadState(*loadstate); err != nil {
			die(err)
		}
		fmt.Printf("restored %s: frame %d\n", *loadstate, m.Frame())
	}

	// -keys takes either an input script or a list of buttons to hold. Which one it is
	// is decided by whether the argument names a file, so the common case (hold Start)
	// stays a one-liner and the case that actually reaches new states (a timed
	// sequence with real press edges) is available under the same flag.
	if *keys != "" {
		if _, err := os.Stat(*keys); err == nil {
			sc, err := dsmachine.LoadScript(*keys)
			if err != nil {
				die(err)
			}
			m.Play(sc)
		} else {
			mask, ok := dsmachine.ParseKeys(*keys)
			if !ok {
				die(fmt.Errorf("-keys: %q is neither a script file nor a list of buttons", *keys))
			}
			m.SetKeys(mask)
		}
	}
	if *touch != "" {
		var x, y int
		if _, err := fmt.Sscanf(*touch, "%d,%d", &x, &y); err != nil {
			die(fmt.Errorf("-touch: want X,Y"))
		}
		m.SetTouch(x, y, true)
	}

	// The instruments. Each is a hook on the machine, so they compose: a breakpoint
	// and a trace and an I/O log can all be live in one run.
	ioSeen := map[uint32]uint32{}
	ioCore := map[uint32]string{}
	m.OnIO = func(arm9, write bool, addr, val, pc uint32) {
		if write {
			ioSeen[addr] = val
			if arm9 {
				ioCore[addr] = "ARM9"
			} else {
				ioCore[addr] = "ARM7"
			}
		}
	}

	bpSet := map[uint32]bool{}
	for _, a := range bps {
		bpSet[a] = true
	}
	logSet := map[uint32]bool{}
	for _, a := range logpcs {
		logSet[a] = true
	}
	var from uint32
	tracing := *trace
	if *traceFrom != "" {
		from, err = parseAddr(*traceFrom)
		if err != nil {
			die(fmt.Errorf("-tracefrom: %v", err))
		}
		tracing = false
	}
	traced := 0
	stop := ""

	if tracing || from != 0 || len(bpSet) > 0 || len(logSet) > 0 {
		m.OnStep = func(arm9 bool, pc uint32) {
			if !arm9 {
				return
			}
			if from != 0 && pc == from {
				tracing = true
			}
			if bpSet[pc] {
				stop = fmt.Sprintf("breakpoint at 0x%08X", pc)
			}
			if logSet[pc] {
				r := m.Regs(true)
				fmt.Printf("  logpc 0x%08X  r0=%08X r1=%08X r2=%08X r3=%08X lr=%08X\n",
					pc, r[0], r[1], r[2], r[3], r[14])
			}
			if tracing && traced < *tracen {
				code := m.Snapshot(true, pc, 4)
				in := arm.DecodeARM(code, pc)
				fmt.Printf("  %08X  %s\n", pc, in.Text)
				traced++
			}
		}
	}

	shots := 0
	if *shot != "" && *shotEvery > 0 {
		prev := m.OnFrame // the input script may already be on this hook
		m.OnFrame = func() {
			if prev != nil {
				prev()
			}
			if m.Frame()%*shotEvery == 0 {
				name := fmt.Sprintf("%s_f%04d", *shot, m.Frame())
				if err := m.WriteShot(name); err != nil {
					die(err)
				}
				shots++
			}
		}
	}

	cards := 0
	if *cardlog {
		m.OnCardXfer(func(cmd [8]byte, addr uint32, size int) {
			cards++
			if cards <= 40 {
				fmt.Printf("  card: cmd=%02X addr=%08X size=%d\n", cmd[0], addr, size)
			}
		})
	}

	milestones := map[uint32]string{
		0x02007000: "main()",
	}

	var res dsmachine.Result
	if *frames > 0 {
		res = m.RunFrames(*frames, budget, *quantum)
	} else {
		res = m.Run(budget, *quantum, milestones)
	}
	if stop != "" {
		res.Reason = stop
	}

	fmt.Printf("\nstopped: %s\n", res.Reason)
	fmt.Printf("frames: %d   scheduler steps: %d\n", res.Frames, res.Steps)
	fmt.Printf("ARM9 PC 0x%08X (parked=%v)   ARM7 PC 0x%08X (parked=%v)\n",
		m.ARM9PC(), m.Parked(true), m.ARM7PC(), m.Parked(false))
	for pc, name := range milestones {
		if at, ok := res.ARM9Milest[pc]; ok {
			fmt.Printf("reached %-8s 0x%08X after %d steps\n", name, pc, at)
		}
	}

	if *shot != "" {
		if err := m.WriteShot(*shot); err != nil {
			die(err)
		}
		fmt.Printf("wrote %s_top.png and %s_bottom.png", *shot, *shot)
		if shots > 0 {
			fmt.Printf(" (+%d numbered captures)", shots)
		}
		fmt.Println()
	}

	if *savestate != "" {
		if err := m.SaveState(*savestate); err != nil {
			die(err)
		}
		fmt.Printf("wrote %s (frame %d)\n", *savestate, m.Frame())
	}

	if *rtshot != "" {
		if err := m.Write3DShot(*rtshot); err != nil {
			die(err)
		}
		fmt.Printf("wrote %s (the 3D render target; magenta = the rasteriser drew nothing there)\n", *rtshot)
	}

	if *cardlog {
		fmt.Printf("cartridge transfers: %d\n", cards)
	}

	if *gfx {
		a, b := m.EngineStats()
		polys, swaps := m.GX()
		emitted, clipped := m.GXClip()
		fmt.Println("\ngraphics state:")
		fmt.Printf("  POWCNT1     = %04X   (engine A drives the %s screen)\n",
			m.Reg(0x04000304), map[bool]string{true: "TOP", false: "BOTTOM"}[m.Reg(0x04000304)&0x8000 != 0])
		fmt.Printf("  DISPCNT A   = %08X   B = %08X\n", m.Reg(0x04000000), m.Reg(0x04001000))
		fmt.Printf("  VRAMCNT A-D = %08X   E-G = %06X   H,I = %04X\n",
			m.Reg(0x04000240), m.Reg(0x04000244)&0xFFFFFF, m.Reg(0x04000248)&0xFFFF)
		fmt.Printf("  non-black pixels: engine A %5d/49152   engine B %5d/49152\n", a, b)
		fmt.Printf("  3D: %d polygons at the last buffer swap, %d swaps\n", polys, swaps)
		fmt.Printf("  3D: %d primitives assembled, %d rejected by the clipper", emitted, clipped)
		if emitted > 0 {
			fmt.Printf(" (%.1f%%)", 100*float64(clipped)/float64(emitted))
		}
		fmt.Println()
	}

	if *gxdump {
		h := m.GXHist()
		fmt.Println("\n3D commands executed:")
		for c := 0; c < 256; c++ {
			if h[c] > 0 {
				fmt.Printf("  %02X %-16s x%d\n", c, gxName(uint8(c)), h[c])
			}
		}
	}

	if *dump != "" {
		parts := strings.SplitN(*dump, ":", 2)
		addr, err := parseAddr(parts[0])
		if err != nil {
			die(err)
		}
		n := uint32(64)
		if len(parts) == 2 {
			v, err := parseUint(parts[1])
			if err != nil {
				die(err)
			}
			n = uint32(v)
		}
		hexdump(m.Snapshot(true, addr, n), addr)
	}

	if *showIO {
		addrs := make([]uint32, 0, len(ioSeen))
		for a := range ioSeen {
			addrs = append(addrs, a)
		}
		sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })
		fmt.Printf("\nI/O registers written (%d):\n", len(addrs))
		for _, a := range addrs {
			fmt.Printf("  %s 0x%08X = 0x%08X   %s\n", ioCore[a], a, ioSeen[a], ioName(a))
		}
	}

	// The gap list. This is the honest half of the run: everything the game asked of
	// the hardware that we do not implement. A short list is a claim; a silent run
	// would be an assumption.
	if *showLog || len(m.Log) > 0 {
		fmt.Printf("\nhardware the model did not implement (%d):\n", len(m.Log))
		for _, s := range m.Log {
			fmt.Println("  " + s)
		}
	}
}

// gxName labels the 3D engine's commands, so a -gxdump reads as the geometry the
// game submitted rather than as a column of opcodes.
func gxName(c uint8) string {
	names := map[uint8]string{
		0x10: "MTX_MODE", 0x11: "MTX_PUSH", 0x12: "MTX_POP", 0x13: "MTX_STORE",
		0x14: "MTX_RESTORE", 0x15: "MTX_IDENTITY", 0x16: "MTX_LOAD_4x4", 0x17: "MTX_LOAD_4x3",
		0x18: "MTX_MULT_4x4", 0x19: "MTX_MULT_4x3", 0x1A: "MTX_MULT_3x3", 0x1B: "MTX_SCALE",
		0x1C: "MTX_TRANS", 0x20: "COLOR", 0x21: "NORMAL", 0x22: "TEXCOORD",
		0x23: "VTX_16", 0x24: "VTX_10", 0x25: "VTX_XY", 0x26: "VTX_XZ", 0x27: "VTX_YZ",
		0x28: "VTX_DIFF", 0x29: "POLYGON_ATTR", 0x2A: "TEXIMAGE_PARAM", 0x2B: "PLTT_BASE",
		0x30: "DIF_AMB", 0x31: "SPE_EMI", 0x32: "LIGHT_VECTOR", 0x33: "LIGHT_COLOR",
		0x34: "SHININESS", 0x40: "BEGIN_VTXS", 0x41: "END_VTXS", 0x50: "SWAP_BUFFERS",
		0x60: "VIEWPORT", 0x70: "BOX_TEST", 0x71: "POS_TEST", 0x72: "VEC_TEST",
	}
	if n, ok := names[c]; ok {
		return n
	}
	return "?"
}

func hexdump(b []byte, base uint32) {
	for i := 0; i < len(b); i += 16 {
		fmt.Printf("  %08X ", base+uint32(i))
		for j := 0; j < 16 && i+j < len(b); j++ {
			fmt.Printf(" %02X", b[i+j])
		}
		fmt.Println()
	}
}

func parseAddr(s string) (uint32, error) {
	v, err := parseUint(s)
	return uint32(v), err
}

func parseUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

// ioName labels the DS I/O registers, so an -io listing reads as the hardware the
// boot programmed rather than as a column of addresses.
func ioName(a uint32) string {
	switch a {
	case 0x04000000:
		return "DISPCNT (engine A)"
	case 0x04000004:
		return "DISPSTAT"
	case 0x04000060:
		return "DISP3DCNT"
	case 0x04000208:
		return "IME"
	case 0x04000210:
		return "IE"
	case 0x04000214:
		return "IF"
	case 0x04000247:
		return "WRAMCNT"
	case 0x04000304:
		return "POWCNT1"
	case 0x04001000:
		return "DISPCNT (engine B)"
	}
	switch {
	case a >= 0x04000008 && a < 0x04000060:
		return "engine-A BG/affine"
	case a >= 0x040000B0 && a < 0x040000E0:
		return "DMA"
	case a >= 0x04000100 && a < 0x04000110:
		return "timers"
	case a >= 0x04000130 && a < 0x04000140:
		return "keys / RTC"
	case a >= 0x04000180 && a < 0x04000190:
		return "IPC (sync/FIFO)"
	case a >= 0x040001A0 && a < 0x040001B0:
		return "cartridge"
	case a >= 0x040001C0 && a < 0x040001C4:
		return "SPI (firmware/touch/power)"
	case a >= 0x04000240 && a < 0x0400024A:
		return "VRAMCNT"
	case a >= 0x04000280 && a < 0x040002C0:
		return "divide / sqrt"
	case a >= 0x04000320 && a < 0x040006A4:
		return "3D engine"
	case a >= 0x04000400 && a < 0x04000520:
		return "sound"
	case a >= 0x04001008 && a < 0x04001060:
		return "engine-B BG/affine"
	}
	return ""
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "bootoracle:", err)
	os.Exit(1)
}
