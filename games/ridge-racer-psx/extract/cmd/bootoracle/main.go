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
	rwatchS := flag.String("rwatch", "", "watch: log who reads ADDR or ADDR:LEN (hex; data reads only)")
	watchN := flag.Int("watchn", 50, "with -watch/-rwatch, stop printing raw accesses after this many "+
		"(a per-PC summary is always printed at the end)")
	gplog := flag.Int("gplog", 0, "log up to N completed GP0 drawing commands (decoded)")
	gpfrom := flag.String("gpfrom", "0", "with -gplog, start logging at this instruction count (hex or decimal)")
	gpop := flag.String("gpop", "", "with -gplog, only log these GP0 opcodes (comma-separated hex, e.g. 2C,3C)")
	dmalog := flag.Bool("dmalog", false, "log each DMA transfer (channel, MADR, BCR, CD sector)")
	dumpS := flag.String("dump", "", "after the run, dump RAM ranges to files: ADDR:LEN:FILE[,ADDR:LEN:FILE...] (hex)")
	vramS := flag.String("vram", "", "after the run, dump raw VRAM (1024x512 little-endian 16-bit words) to this file")
	showLog := flag.Bool("log", false, "print the machine's diagnostic notes")
	showTTY := flag.Bool("tty", true, "print the game's BIOS/TTY output")
	shot := flag.String("shot", "", "after the run, write the GPU display to this PNG")
	isrS := flag.String("isr", "8004DF48", "vectored-interrupt entry (hex); Ridge Racer's own "+
		"interrupt dispatcher, traced — the retail BIOS installs it via HookEntryInt into a slot "+
		"the HLE BIOS leaves empty. Set 0 to use the (empty) registered chain handler")
	saveS := flag.String("save", "", "after the run, write a machine savestate to this file")
	loadS := flag.String("load", "", "before the run, restore a machine savestate from this file "+
		"(the -steps budget then counts from the restored instruction count)")
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
		script, err := psx.ParsePress(*press)
		if err != nil {
			die("bad -press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)
	if *loadS != "" {
		if err := m.LoadStateFile(*loadS); err != nil {
			die("load state: %v", err)
		}
		fmt.Fprintf(os.Stderr, "restored state: PC 0x%08X, %d instructions\n", m.CPU.PC, m.CPU.Steps)
	}

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
	var wsum, rsum *accessSummary
	if *watchS != "" {
		lo, ln, err := parseWatch(*watchS)
		if err != nil {
			die("bad -watch")
		}
		m.WatchLo, m.WatchHi = lo, lo+ln
		wsum = &accessSummary{kind: "write", pcs: map[uint32]*pcStat{}}
		m.OnWrite = func(addr, val, pc uint32) {
			if wsum.raw < *watchN {
				fmt.Fprintf(w, "write [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
			}
			wsum.add(addr, pc)
		}
	}
	if *rwatchS != "" {
		lo, ln, err := parseWatch(*rwatchS)
		if err != nil {
			die("bad -rwatch")
		}
		m.RWatchLo, m.RWatchHi = lo, lo+ln
		rsum = &accessSummary{kind: "read", pcs: map[uint32]*pcStat{}}
		m.OnRead = func(addr, val, pc uint32) {
			if rsum.raw < *watchN {
				fmt.Fprintf(w, "read  [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
			}
			rsum.add(addr, pc)
		}
	}
	if *gplog > 0 {
		gpFrom, err := parseCount(*gpfrom)
		if err != nil {
			die("bad -gpfrom")
		}
		var gpOps map[byte]bool
		if *gpop != "" {
			gpOps = map[byte]bool{}
			for _, s := range strings.Split(*gpop, ",") {
				op, err := hx(strings.TrimSpace(s))
				if err != nil || op > 0xFF {
					die("bad -gpop entry %q", s)
				}
				gpOps[byte(op)] = true
			}
		}
		remaining := *gplog
		m.OnGP0(func(words []uint32) {
			if remaining <= 0 || m.CPU.Steps < gpFrom {
				return
			}
			if gpOps != nil && !gpOps[byte(words[0]>>24)] {
				return
			}
			remaining--
			fmt.Fprintf(w, "gp0 @%d  %s\n", m.CPU.Steps, describePrim(words))
		})
	}
	if *dmalog {
		m.OnDMA = func(ch int, madr, bcr, chcr uint32, lba int) {
			mode := "block"
			if (chcr>>9)&3 == 2 {
				mode = "list"
			} else if chcr&1 == 0 {
				mode = "to-RAM"
			}
			if ch == 3 {
				fmt.Fprintf(w, "dma @%d  ch3 CD->RAM  madr=0x%06X bcr=0x%08X lba=%d\n",
					m.CPU.Steps, madr, bcr, lba)
			} else {
				fmt.Fprintf(w, "dma @%d  ch%d %-7s  madr=0x%06X bcr=0x%08X\n",
					m.CPU.Steps, ch, mode, madr, bcr)
			}
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

	wsum.print(w)
	rsum.print(w)

	if *saveS != "" {
		if err := m.SaveStateFile(*saveS); err != nil {
			die("save state: %v", err)
		}
		fmt.Fprintf(os.Stderr, "saved state to %s (at %d instructions)\n", *saveS, m.CPU.Steps)
	}
	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			die("screenshot: %v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote frame to %s\n", *shot)
	}
	if *dumpS != "" {
		m.OnRead = nil // our own reads, not the program's
		for _, spec := range strings.Split(*dumpS, ",") {
			addr, ln, file, err := parseDump(spec)
			if err != nil {
				die("bad -dump entry %q: %v", spec, err)
			}
			buf := make([]byte, ln)
			for i := uint32(0); i < ln; i++ {
				buf[i] = m.Read(addr + i)
			}
			if err := os.WriteFile(file, buf, 0644); err != nil {
				die("dump: %v", err)
			}
			fmt.Fprintf(os.Stderr, "dumped 0x%08X:0x%X to %s\n", addr, ln, file)
		}
	}
	if *vramS != "" {
		vram := m.VRAM()
		buf := make([]byte, len(vram)*2)
		for i, px := range vram {
			buf[i*2] = byte(px)
			buf[i*2+1] = byte(px >> 8)
		}
		if err := os.WriteFile(*vramS, buf, 0644); err != nil {
			die("vram: %v", err)
		}
		fmt.Fprintf(os.Stderr, "dumped VRAM to %s\n", *vramS)
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

// accessSummary aggregates watch hits per accessing PC so a long run ends with
// "which routines touch this range" instead of a megabyte of raw lines.
type accessSummary struct {
	kind string
	raw  int
	pcs  map[uint32]*pcStat
}

type pcStat struct {
	count  int
	lo, hi uint32 // address range touched (inclusive)
}

func (s *accessSummary) add(addr, pc uint32) {
	s.raw++
	st := s.pcs[pc]
	if st == nil {
		st = &pcStat{lo: addr, hi: addr}
		s.pcs[pc] = st
	}
	st.count++
	if addr < st.lo {
		st.lo = addr
	}
	if addr > st.hi {
		st.hi = addr
	}
}

func (s *accessSummary) print(w *bufio.Writer) {
	if s == nil || len(s.pcs) == 0 {
		return
	}
	pcs := make([]uint32, 0, len(s.pcs))
	for pc := range s.pcs {
		pcs = append(pcs, pc)
	}
	sort.Slice(pcs, func(i, j int) bool { return s.pcs[pcs[i]].count > s.pcs[pcs[j]].count })
	fmt.Fprintf(w, "\n%s summary (%d accesses, %d PCs):\n", s.kind, s.raw, len(pcs))
	for _, pc := range pcs {
		st := s.pcs[pc]
		fmt.Fprintf(w, "  PC 0x%08X  %8d %ss  [0x%08X..0x%08X]\n", pc, st.count, s.kind, st.lo, st.hi)
	}
}

// parseDump splits an ADDR:LEN:FILE spec (ADDR and LEN hex).
func parseDump(s string) (addr, ln uint32, file string, err error) {
	parts := strings.SplitN(strings.TrimSpace(s), ":", 3)
	if len(parts) != 3 {
		return 0, 0, "", fmt.Errorf("want ADDR:LEN:FILE")
	}
	if addr, err = hx(parts[0]); err != nil {
		return
	}
	if ln, err = hx(parts[1]); err != nil {
		return
	}
	return addr, ln, parts[2], nil
}

// describePrim renders one completed GP0 command for the -gplog trace: the
// opcode with its class decoded, and for polygons/rectangles the vertices,
// texcoords, CLUT and texture page as the words encode them.
func describePrim(words []uint32) string {
	op := byte(words[0] >> 24)
	var b strings.Builder
	fmt.Fprintf(&b, "%02X", op)
	vtx := func(w uint32) {
		x := int16(w&0xFFFF) << 5 >> 5 // signed 11-bit
		y := int16(w>>16) << 5 >> 5
		fmt.Fprintf(&b, " (%d,%d)", x, y)
	}
	switch {
	case op >= 0x20 && op <= 0x3F: // polygon
		verts := 3
		if op&0x08 != 0 {
			verts = 4
		}
		gouraud := op&0x10 != 0
		textured := op&0x04 != 0
		tag := "tri"
		if verts == 4 {
			tag = "quad"
		}
		if textured {
			tag += " tex"
		}
		if gouraud {
			tag += " gouraud"
		} else {
			tag += " flat"
		}
		if op&0x02 != 0 {
			tag += " semi"
		}
		fmt.Fprintf(&b, " %s rgb=%06X", tag, words[0]&0xFFFFFF)
		i := 1
		var clut, page uint32
		for v := 0; v < verts; v++ {
			if gouraud && v > 0 {
				i++ // per-vertex colour word
			}
			vtx(words[i])
			i++
			if textured {
				uv := words[i]
				fmt.Fprintf(&b, " uv=(%d,%d)", uv&0xFF, (uv>>8)&0xFF)
				if v == 0 {
					clut = uv >> 16
				}
				if v == 1 {
					page = uv >> 16
				}
				i++
			}
		}
		if textured {
			fmt.Fprintf(&b, " clut=0x%04X page=0x%04X", clut, page)
		}
	case op >= 0x60 && op <= 0x7F: // rectangle
		fmt.Fprintf(&b, " rect rgb=%06X", words[0]&0xFFFFFF)
		vtx(words[1])
		i := 2
		if op&0x04 != 0 {
			uv := words[i]
			fmt.Fprintf(&b, " uv=(%d,%d) clut=0x%04X", uv&0xFF, (uv>>8)&0xFF, uv>>16)
			i++
		}
		if op&0x18 == 0 && i < len(words) {
			fmt.Fprintf(&b, " wh=(%d,%d)", words[i]&0xFFFF, words[i]>>16)
		}
	case op >= 0xA0 && op <= 0xBF:
		fmt.Fprintf(&b, " img->vram dst=(%d,%d) wh=(%d,%d)",
			words[1]&0x3FF, (words[1]>>16)&0x1FF, words[2]&0xFFFF, words[2]>>16)
	case op >= 0xC0 && op <= 0xDF:
		fmt.Fprintf(&b, " vram->cpu src=(%d,%d) wh=(%d,%d)",
			words[1]&0x3FF, (words[1]>>16)&0x1FF, words[2]&0xFFFF, words[2]>>16)
	case op >= 0x80 && op <= 0x9F:
		fmt.Fprintf(&b, " vram->vram src=(%d,%d) dst=(%d,%d) wh=(%d,%d)",
			words[1]&0x3FF, (words[1]>>16)&0x1FF, words[2]&0x3FF, (words[2]>>16)&0x1FF,
			words[3]&0x3FF, words[3]>>16)
	case op == 0x02:
		fmt.Fprintf(&b, " fill rgb=%06X xy=(%d,%d) wh=(%d,%d)", words[0]&0xFFFFFF,
			words[1]&0x3FF, (words[1]>>16)&0x1FF, words[2]&0x3FF, (words[2]>>16)&0x1FF)
	default:
		for _, wd := range words {
			fmt.Fprintf(&b, " %08X", wd)
		}
	}
	return b.String()
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
