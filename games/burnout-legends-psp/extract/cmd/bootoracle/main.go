// bootoracle boots Burnout Legends' executable under the PSP oracle (tools/platform/psp)
// and runs it, exposing the repo's reverse-engineering instrumentation: an
// instruction trace, breakpoints, memory watches, the kernel syscall census, the
// game's Kprintf/stdout, and a savestate. It reads the UMD image (a .cso CISO or a
// .iso), KIRK-decrypts EBOOT.BIN, relocates and loads the module — no manual
// extraction step.
//
// Usage:
//
//	bootoracle -image "image/Burnout Legends.cso" [-steps N] [-trace -tracen N]
//	           [-bp ADDR] [-watch ADDR[:LEN]] [-savestate FILE] [-loadstate FILE]
//
// The oracle boots the C runtime and module start, streams the game's assets off
// the UMD, renders its frames (-shot), and plays them: -keys schedules pad input
// by VBlank (through the language screen, the title menu, the intro dialogue and
// into the stage). -trace shows live execution, -watch maps the code that produces
// a memory structure, -find locates byte patterns in RAM, -gelog/-gedump summarize
// or dump the submitted GE display lists, and the syscall census names the kernel
// functions the boot path invokes.
package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/psp"
)

func hx(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return uint32(v), err
}

func parseCount(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		v, err := hx(s)
		return uint64(v), err
	}
	return strconv.ParseUint(s, 10, 64)
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bootoracle: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	exe := flag.String("exe", "PSP_GAME/SYSDIR/EBOOT.BIN", "boot executable path on the disc")
	stepsF := flag.String("steps", "20000000", "max instructions to run (hex or decimal)")
	trace := flag.Bool("trace", false, "disassemble each executed instruction")
	tracen := flag.Int("tracen", 200, "with -trace, stop printing after this many instructions")
	traceThread := flag.String("tracethread", "", "with -trace, only print instructions of this thread")
	bpS := flag.String("bp", "", "breakpoint: stop when PC reaches this address (hex)")
	watchS := flag.String("watch", "", "watch: log who writes ADDR or ADDR:LEN (hex)")
	rwatchS := flag.String("rwatch", "", "watch: log who reads ADDR or ADDR:LEN (hex)")
	rprofS := flag.String("rprofile", "", "profile reads of ADDR:LEN: per reading PC, report count and lo..hi range at exit")
	watchN := flag.Int("watchn", 50, "with -watch/-rwatch, stop printing raw accesses after this many")
	keysS := flag.String("keys", "", "pad input script: file of '<vblank> [buttons...]' lines (cross, start, up, ...)")
	saveS := flag.String("savestate", "", "after the run, write a machine savestate to this file")
	loadS := flag.String("loadstate", "", "before the run, restore a machine savestate from this file")
	shot := flag.String("shot", "", "after the run, write the display framebuffer to this PNG")
	geLog := flag.Int("gelog", 0, "print a command summary of the first N GE display lists")
	geDump := flag.Int("gedump", 0, "print every command word of the first N GE display lists")
	disS := flag.String("dis", "", "disassemble ADDR:LEN of loaded memory and exit (after -loadstate)")
	dumpS := flag.String("dump", "", "hex-dump ADDR:LEN of loaded memory and exit (after -loadstate)")
	findS := flag.String("find", "", "search loaded RAM for hex bytes (e.g. F0208908) and exit (after -loadstate)")
	dumpBinS := flag.String("dumpbin", "", "write ADDR:LEN of loaded memory to a file (ADDR:LEN:FILE) and exit (after -loadstate)")
	showNotes := flag.Bool("notes", false, "print the machine's diagnostic notes")
	flag.Parse()
	if *image == "" {
		die("need -image")
	}

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	im, err := psp.OpenImage(*image)
	if err != nil {
		die("%v", err)
	}
	defer im.Close()
	fmt.Fprintf(os.Stderr, "volume: system=%q name=%q\n", im.System, im.Name)

	mod, err := im.LoadExecutable(*exe)
	if err != nil {
		die("load %s: %v", *exe, err)
	}
	fmt.Fprintf(os.Stderr, "module %q: entry 0x%08X gp 0x%08X, %d imports\n",
		mod.Name, mod.EntryPC, mod.GP, len(mod.Imports))

	steps, err := parseCount(*stepsF)
	if err != nil {
		die("bad -steps")
	}

	m := psp.NewMachine()
	m.SetImageHash(fmt.Sprintf("%x", md5.Sum(data)))
	m.SetVolume(im.Volume)
	if err := m.LoadModule(mod); err != nil {
		die("load module: %v", err)
	}
	if *loadS != "" {
		if err := m.LoadStateFile(*loadS); err != nil {
			die("load state: %v", err)
		}
		fmt.Fprintf(os.Stderr, "restored state: PC 0x%08X, %d instructions, vblank %d\n", m.CPU.PC, m.CPU.Steps, m.Vblanks())
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	if *disS != "" {
		lo, ln := parseWatch(*disS)
		for a := lo; a < lo+ln; a += 4 {
			fmt.Fprintf(w, "%08X  %s\n", a, strings.TrimSpace(m.DisasmAt(a)))
		}
		return
	}
	if *findS != "" {
		var pat []byte
		for i := 0; i+1 < len(*findS); i += 2 {
			v, err := strconv.ParseUint((*findS)[i:i+2], 16, 8)
			if err != nil {
				die("bad -find hex")
			}
			pat = append(pat, byte(v))
		}
		if len(pat) == 0 {
			die("empty -find pattern")
		}
		const ramLo, ramHi = 0x08800000, 0x0A000000
		for a := uint32(ramLo); a < ramHi-uint32(len(pat)); a++ {
			hit := true
			for i, b := range pat {
				if m.Read(a+uint32(i)) != b {
					hit = false
					break
				}
			}
			if hit {
				fmt.Fprintf(w, "found at 0x%08X\n", a)
			}
		}
		return
	}
	if *dumpBinS != "" {
		parts := strings.SplitN(*dumpBinS, ":", 3)
		if len(parts) != 3 {
			die("-dumpbin wants ADDR:LEN:FILE")
		}
		lo, _ := hx(parts[0])
		ln, _ := hx(parts[1])
		buf := make([]byte, ln)
		for i := uint32(0); i < ln; i++ {
			buf[i] = m.Read(lo + i)
		}
		if err := os.WriteFile(parts[2], buf, 0644); err != nil {
			die("%v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote 0x%X bytes at 0x%08X to %s\n", ln, lo, parts[2])
		return
	}
	if *dumpS != "" {
		lo, ln := parseWatch(*dumpS)
		for a := lo; a < lo+ln; a += 16 {
			fmt.Fprintf(w, "%08X ", a)
			for i := uint32(0); i < 16; i++ {
				fmt.Fprintf(w, " %02X", m.Read(a+i))
			}
			fmt.Fprint(w, "  ")
			for i := uint32(0); i < 16; i++ {
				b := m.Read(a + i)
				if b < 0x20 || b > 0x7E {
					b = '.'
				}
				fmt.Fprintf(w, "%c", b)
			}
			fmt.Fprintln(w)
		}
		return
	}

	if *keysS != "" {
		evs, err := parseKeys(*keysS)
		if err != nil {
			die("%v", err)
		}
		m.SetPadScript(evs)
	}

	// Instrumentation.
	var bp uint32
	haveBP := false
	if *bpS != "" {
		if bp, err = hx(*bpS); err != nil {
			die("bad -bp")
		}
		haveBP = true
	}
	if *watchS != "" {
		lo, ln := parseWatch(*watchS)
		m.WatchLo, m.WatchHi = lo, lo+ln
		raw := 0
		m.OnWrite = func(addr, val, pc uint32) {
			if raw < *watchN {
				fmt.Fprintf(w, "write [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
				raw++
			}
		}
	}
	if *rwatchS != "" {
		lo, ln := parseWatch(*rwatchS)
		m.RWatchLo, m.RWatchHi = lo, lo+ln
		raw := 0
		m.OnRead = func(addr, val, pc uint32) {
			if raw < *watchN {
				fmt.Fprintf(w, "read  [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
				raw++
			}
		}
	}
	type rprof struct {
		count  uint64
		lo, hi uint32
	}
	var rprofile map[uint32]*rprof
	if *rprofS != "" {
		lo, ln := parseWatch(*rprofS)
		m.RWatchLo, m.RWatchHi = lo, lo+ln
		rprofile = make(map[uint32]*rprof)
		m.OnRead = func(addr, val, pc uint32) {
			p := rprofile[pc]
			if p == nil {
				p = &rprof{lo: addr, hi: addr}
				rprofile[pc] = p
			}
			p.count++
			if addr < p.lo {
				p.lo = addr
			}
			if addr > p.hi {
				p.hi = addr
			}
		}
	}
	if *geDump > 0 {
		dumped := 0
		m.OnGeList = func(l psp.GeList) {
			if dumped >= *geDump {
				return
			}
			dumped++
			fmt.Fprintf(w, "GE list %d @0x%08X: %d words\n", dumped, l.Start, len(l.Words))
			for i, word := range l.Words {
				fmt.Fprintf(w, "  [%4d] %08X  %-14s arg 0x%06X\n", i, word, psp.GeCmdName(word>>24), word&0xFFFFFF)
			}
		}
	} else if *geLog > 0 {
		logged := 0
		m.OnGeList = func(l psp.GeList) {
			if logged >= *geLog {
				return
			}
			logged++
			counts := map[string]int{}
			prims := ""
			for _, word := range l.Words {
				cmd := word >> 24
				counts[psp.GeCmdName(cmd)]++
				if cmd == 0x04 && len(prims) < 200 { // PRIM: type + vertex count
					prims += fmt.Sprintf(" %d:%d", (word>>16)&7, word&0xFFFF)
				}
			}
			fmt.Fprintf(w, "GE list %d @0x%08X: %d words:", logged, l.Start, len(l.Words))
			for _, k := range []string{"PRIM", "VTYPE", "VADDR", "FBP", "CLEAR", "JUMP", "CALL", "FINISH", "END"} {
				if counts[k] > 0 {
					fmt.Fprintf(w, " %s=%d", k, counts[k])
				}
			}
			if prims != "" {
				fmt.Fprintf(w, " prim(type:n):%s", prims)
			}
			fmt.Fprintln(w)
		}
	}
	traced := 0
	hitBP := false
	var bpPC uint32
	if *trace || haveBP {
		m.OnStep = func(mm *psp.Machine, pc uint32) {
			if haveBP && !hitBP && (pc&0x1FFFFFFF) == (bp&0x1FFFFFFF) {
				hitBP, bpPC = true, pc
				mm.Halted, mm.HaltReason = true, "breakpoint"
				names := [32]string{"zr", "at", "v0", "v1", "a0", "a1", "a2", "a3",
					"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
					"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
					"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra"}
				for r := 1; r < 32; r++ {
					fmt.Fprintf(w, "$%s=%08X ", names[r], mm.CPU.Reg(uint32(r)))
					if r%8 == 7 {
						fmt.Fprintln(w)
					}
				}
				fmt.Fprintf(w, "  thread=%s\n", mm.CurrentThread())
			}
			if *trace && traced < *tracen &&
				(*traceThread == "" || mm.CurrentThread() == *traceThread) {
				fmt.Fprintf(w, "%08X  %s\n", pc, strings.TrimSpace(mm.DisasmAt(pc)))
				traced++
			}
		}
	}

	res := m.Run(steps)
	if hitBP {
		res.PC, res.Reason = bpPC, "breakpoint"
	}

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
		fmt.Fprintf(os.Stderr, "wrote framebuffer (%s) to %s\n", m.FramebufferInfo(), *shot)
	}

	if rprofile != nil {
		var pcs []uint32
		for pc := range rprofile {
			pcs = append(pcs, pc)
		}
		sort.Slice(pcs, func(i, j int) bool { return pcs[i] < pcs[j] })
		fmt.Fprintf(w, "read profile of 0x%08X..0x%08X: %d reading PCs\n", m.RWatchLo, m.RWatchHi, len(pcs))
		for _, pc := range pcs {
			p := rprofile[pc]
			fmt.Fprintf(w, "  PC 0x%08X  %9d reads  0x%08X..0x%08X (+0x%06X..+0x%06X)\n",
				pc, p.count, p.lo, p.hi, p.lo-m.RWatchLo, p.hi-m.RWatchLo)
		}
	}
	w.Flush()
	fmt.Fprintf(os.Stderr, "\n%s (vblank %d)\n", res, m.Vblanks())
	if tty := m.TTY(); tty != "" {
		fmt.Fprintf(os.Stderr, "TTY:\n%s\n", tty)
	}
	printCensus(m)
	for _, l := range m.Threads() {
		fmt.Fprintln(os.Stderr, l)
	}
	for _, l := range m.KObjects() {
		fmt.Fprintln(os.Stderr, l)
	}
	if *showNotes {
		for _, l := range m.Log {
			fmt.Fprintln(os.Stderr, "note:", l)
		}
	}
}

func printCensus(m *psp.Machine) {
	type kv struct {
		name string
		n    int
	}
	var all []kv
	for k, v := range m.SyscallCalls {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].n > all[j].n })
	fmt.Fprintf(os.Stderr, "kernel syscalls reached (%d):\n", len(all))
	for _, e := range all {
		fmt.Fprintf(os.Stderr, "  %5d  %s\n", e.n, e.name)
	}
}

// padButtons maps script names to pspctrl button bits (platform spec).
var padButtons = map[string]uint32{
	"select": 0x000001, "start": 0x000008,
	"up": 0x000010, "right": 0x000020, "down": 0x000040, "left": 0x000080,
	"ltrigger": 0x000100, "rtrigger": 0x000200,
	"triangle": 0x001000, "circle": 0x002000, "cross": 0x004000, "square": 0x008000,
}

// parseKeys reads a pad script: one event per line, "<vblank> [button ...]".
// The named buttons are held from that VBlank until the next event (a line with
// no buttons releases them all). '#' starts a comment.
func parseKeys(path string) ([]psp.PadEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var evs []psp.PadEvent
	for ln, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		at, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: bad vblank %q", path, ln+1, fields[0])
		}
		var buttons uint32
		for _, name := range fields[1:] {
			bit, ok := padButtons[strings.ToLower(name)]
			if !ok {
				return nil, fmt.Errorf("%s:%d: unknown button %q", path, ln+1, name)
			}
			buttons |= bit
		}
		evs = append(evs, psp.PadEvent{AtVblank: uint32(at), Buttons: buttons})
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].AtVblank < evs[j].AtVblank })
	return evs, nil
}

func parseWatch(s string) (lo, ln uint32) {
	parts := strings.SplitN(s, ":", 2)
	lo, _ = hx(parts[0])
	ln = 1
	if len(parts) == 2 {
		ln, _ = hx(parts[1])
	}
	return
}
