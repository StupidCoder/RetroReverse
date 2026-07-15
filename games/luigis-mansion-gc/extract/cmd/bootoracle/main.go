// bootoracle boots Luigi's Mansion on the GameCube machine and runs it, with the tracing
// and inspection the other oracles in this repository provide.
//
// The GameCube has no operating system to service, so this oracle is about execution and
// hardware and nothing else: it lays down the IPL globals, runs the disc's own apploader
// to load the game, and then steps the Gekko while the machine's devices answer. The most
// useful thing it can report is what the game reads off its own disc, by file name — that
// is -dvd, and it is the game telling us, in its own words, what it is loading.
//
// Usage:
//
//	bootoracle -image DISC.iso [-steps N] [-trace] [-dvd] [-loaddol] [-shot BASE] ...
//
// The DOL has no symbol table, so unlike the PS2 oracle a breakpoint or a watch takes a
// hex address or a segment keyword (entry, text0..text6, bss), not a name.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/cpu/gekko"
	"retroreverse.com/tools/platform/gc"
)

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

func main() {
	image := flag.String("image", "", "disc image (.iso / .gcm)")
	stepsS := flag.String("steps", "500000000", "instruction budget (hex or decimal)")
	trace := flag.Bool("trace", false, "trace execution")
	tracen := flag.Int("tracen", 200, "limit traced instructions")
	tracefrom := flag.String("tracefrom", "", "start tracing when this address is first reached (hex or segment keyword)")
	var bps multiFlag
	flag.Var(&bps, "bp", "halting breakpoint (hex or segment keyword); repeatable")
	var logpcs multiFlag
	flag.Var(&logpcs, "logpc", "non-halting breakpoint: log registers and continue; repeatable")
	var watches multiFlag
	flag.Var(&watches, "watch", "write-watch ADDR[:LEN] (hex); repeatable")
	var rwatches multiFlag
	flag.Var(&rwatches, "rwatch", "read-watch ADDR[:LEN] (hex); repeatable")
	watchn := flag.Int("watchn", 100, "limit watch reports")
	loadDOL := flag.Bool("loaddol", false, "load the DOL directly, skipping the apploader (the bisection tool)")
	dvd := flag.Bool("dvd", false, "log every disc read with the file it lands in")
	lowmem := flag.Bool("lowmem", false, "poison the low-memory globals and report which the game reads")
	shot := flag.String("shot", "", "write the framebuffer (XFB) as a PNG to this path when the run ends")
	savestate := flag.String("savestate", "", "write a snapshot at the end of the run")
	loadstate := flag.String("loadstate", "", "start from a snapshot")
	poke := flag.String("poke", "", "write ADDR:VALUE (hex) after loading, before running")
	dis := flag.String("dis", "", "disassemble ADDR[:N] and exit (hex)")
	dump := flag.String("dump", "", "hex-dump ADDR:LEN and exit (hex)")
	files := flag.Bool("files", false, "list the disc's files and exit")
	verbose := flag.Bool("v", false, "report the unmodelled-hardware census at the end")
	nospin := flag.Bool("nospin", false, "do not stop on a tight loop (the OS idle/scheduler loop looks like one)")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: bootoracle -image DISC.iso [-steps N] [-dvd] [-loaddol] [-shot BASE] ...")
		os.Exit(2)
	}
	if err := run(cfg{
		image: *image, steps: *stepsS, trace: *trace, tracen: *tracen, tracefrom: *tracefrom,
		bps: bps, logpcs: logpcs, watches: watches, rwatches: rwatches, watchn: *watchn,
		loadDOL: *loadDOL, dvd: *dvd, lowmem: *lowmem, shot: *shot,
		savestate: *savestate, loadstate: *loadstate, poke: *poke,
		dis: *dis, dump: *dump, files: *files, verbose: *verbose, nospin: *nospin,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

type cfg struct {
	image                                string
	steps                                string
	trace                                bool
	tracen                               int
	tracefrom                            string
	bps, logpcs, watches, rwatches       multiFlag
	watchn                               int
	loadDOL, dvd, lowmem, files, verbose bool
	nospin                               bool
	shot, savestate, loadstate, poke     string
	dis, dump                            string
}

func run(c cfg) error {
	disc, err := gc.Open(c.image)
	if err != nil {
		return err
	}
	defer disc.Close()

	if c.files {
		for _, f := range disc.FST.Files() {
			fmt.Printf("%10d  0x%010X  %s\n", f.Size, f.Offset, f.Path)
		}
		return nil
	}

	m, err := gc.NewMachine(disc)
	if err != nil {
		return err
	}

	// -dis and -dump answer without running: they load the executable and inspect it.
	if c.dis != "" || c.dump != "" {
		if _, err := m.LoadDOL(); err != nil {
			return err
		}
		if c.dis != "" {
			return doDis(m, c.dis)
		}
		return doDump(m, c.dump)
	}

	// The disc-read log: resolve each read to the file it lands in.
	if c.dvd {
		m.OnDVDRead = func(off int64, length, memAddr uint32) {
			if f, within, ok := disc.FST.ByOffset(off); ok {
				fmt.Printf("  DVD read: %s + 0x%X  (%d bytes -> 0x%08X)\n", f.Path, within, length, memAddr)
			} else {
				fmt.Printf("  DVD read: 0x%X  (%d bytes -> 0x%08X) — not in any file\n", off, length, memAddr)
			}
		}
	}

	if c.lowmem {
		m.PoisonLowMem()
		reads := 0
		m.RWatchLo, m.RWatchHi = 0, 0x3000
		m.OnRead = func(addr, val, pc uint32) {
			if val&0xFFFF0000 == 0xF00D0000 && reads < 100 { // still poison: the game read a global we did not set
				fmt.Printf("  low-mem read of UNSET 0x%08X (pc 0x%08X)\n", addr, pc)
				reads++
			}
		}
	}

	// Load the game: through the apploader (the real path) or directly (the bisection one).
	var entry uint32
	if c.loadstate != "" {
		if err := m.LoadStateFile(c.loadstate); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "bootoracle: resumed from %s at PC 0x%08X\n", c.loadstate, m.CPU.PC)
	} else if c.loadDOL {
		entry, err = m.LoadDOL()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "bootoracle: loaded the DOL directly; entry 0x%08X\n", entry)
	} else {
		entry, err = m.RunApploader()
		if err != nil {
			return fmt.Errorf("the apploader did not complete: %w", err)
		}
		fmt.Fprintf(os.Stderr, "bootoracle: the apploader ran and returned the entry point 0x%08X\n", entry)
		// Gate 2.1: the entry the loader reports must be the one the DOL header names.
		dol, _ := disc.DOL()
		if entry == dol.Entry {
			fmt.Fprintf(os.Stderr, "bootoracle: it matches the DOL header (0x%08X) — the IPL handoff is correct\n", dol.Entry)
		} else {
			fmt.Fprintf(os.Stderr, "bootoracle: WARNING it does not match the DOL header (0x%08X)\n", dol.Entry)
		}
	}

	if c.poke != "" {
		addr, val, err := parsePoke(c.poke)
		if err != nil {
			return err
		}
		m.Poke(addr, val)
	}

	// Instruments.
	for _, s := range c.bps {
		a, err := parseHex(s)
		if err != nil {
			return fmt.Errorf("bad -bp %q", s)
		}
		m.SetBreakpoint(a)
	}
	if len(c.watches) > 0 {
		lo, hi, err := parseWatch(c.watches[0])
		if err != nil {
			return err
		}
		n := 0
		m.WatchLo, m.WatchHi = lo, hi
		m.OnWrite = func(addr, val, pc uint32) {
			if n < c.watchn {
				fmt.Printf("  write 0x%08X = 0x%08X (pc 0x%08X)\n", addr, val, pc)
				n++
			}
		}
	}
	if len(c.rwatches) > 0 && !c.lowmem {
		lo, hi, err := parseWatch(c.rwatches[0])
		if err != nil {
			return err
		}
		n := 0
		m.RWatchLo, m.RWatchHi = lo, hi
		m.OnRead = func(addr, val, pc uint32) {
			if n < c.watchn {
				fmt.Printf("  read 0x%08X = 0x%08X (pc 0x%08X)\n", addr, val, pc)
				n++
			}
		}
	}
	if c.trace {
		traced := 0
		on := c.tracefrom == ""
		var fromAddr uint32
		if c.tracefrom != "" {
			fromAddr, _ = parseHex(c.tracefrom)
		}
		m.OnStep = func(mm *gc.Machine, pc uint32) {
			if !on && pc == fromAddr {
				on = true
			}
			if on && traced < c.tracen {
				w := mm.ReadVirt32(pc)
				fmt.Printf("%08X  %s\n", pc, gekko.DecodeWord(w, pc).Text)
				traced++
			}
		}
	}
	for _, s := range c.logpcs {
		a, err := parseHex(s)
		if err != nil {
			return fmt.Errorf("bad -logpc %q", s)
		}
		prev := m.OnStep
		target := a
		m.OnStep = func(mm *gc.Machine, pc uint32) {
			if prev != nil {
				prev(mm, pc)
			}
			if pc == target {
				fmt.Printf("  logpc 0x%08X: r3=0x%08X r4=0x%08X r5=0x%08X lr=0x%08X\n",
					pc, mm.CPU.GPR[3], mm.CPU.GPR[4], mm.CPU.GPR[5], mm.CPU.LR)
			}
		}
	}

	// The run.
	steps, err := parseUint(c.steps)
	if err != nil {
		return fmt.Errorf("bad -steps %q", c.steps)
	}
	m.SetSpinDetect(!c.nospin)
	res := m.Run(steps)
	fmt.Fprintf(os.Stderr, "bootoracle: %s\n", res)
	fmt.Fprintf(os.Stderr, "bootoracle: VI fields elapsed: %d\n", m.VIField())
	if c.verbose {
		fmt.Fprintf(os.Stderr, "bootoracle: intr: %s\n", m.IntrState())
		fmt.Fprintf(os.Stderr, "bootoracle: backtrace:\n%s", m.BacktraceString())
	}

	if c.shot != "" {
		if err := writeShot(m, c.shot); err != nil {
			fmt.Fprintln(os.Stderr, "bootoracle: -shot:", err)
		}
	}
	if c.savestate != "" {
		if err := m.SaveStateFile(c.savestate); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "bootoracle: wrote %s\n", c.savestate)
	}

	if c.verbose || len(m.Census()) > 0 {
		cen := m.Census()
		fmt.Fprintf(os.Stderr, "\nbootoracle: %d unmodelled-hardware events:\n", len(cen))
		for _, s := range cen {
			fmt.Fprintf(os.Stderr, "  %s\n", s)
		}
	}
	return nil
}

// --- Instruments --------------------------------------------------------------------

func writeShot(m *gc.Machine, path string) error {
	img, err := m.RenderXFB()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func doDis(m *gc.Machine, spec string) error {
	addr, n, err := parseAddrN(spec, 32)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		a := addr + uint32(i*4)
		w := m.ReadVirt32(a)
		in := gekko.DecodeWord(w, a)
		fmt.Printf("%08X  %08X  %s\n", a, w, in.Text)
	}
	return nil
}

func doDump(m *gc.Machine, spec string) error {
	addr, n, err := parseAddrN(spec, 64)
	if err != nil {
		return err
	}
	for i := 0; i < n; i += 16 {
		fmt.Printf("%08X ", addr+uint32(i))
		var ascii []byte
		for j := 0; j < 16 && i+j < n; j++ {
			b := m.ReadVirt8(addr + uint32(i+j))
			fmt.Printf(" %02X", b)
			if b >= 0x20 && b < 0x7F {
				ascii = append(ascii, b)
			} else {
				ascii = append(ascii, '.')
			}
		}
		fmt.Printf("  |%s|\n", ascii)
	}
	return nil
}

// --- Small parsers ------------------------------------------------------------------

func parseUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseHex(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimSpace(s), "0x"), 16, 64)
	return uint32(v), err
}

func parseWatch(s string) (lo, hi uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = parseHex(parts[0]); err != nil {
		return
	}
	length := uint32(4)
	if len(parts) == 2 {
		n, e := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 32)
		if e != nil {
			return 0, 0, e
		}
		length = uint32(n)
	}
	return lo, lo + length, nil
}

func parsePoke(s string) (addr, val uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad -poke %q, want ADDR:VALUE", s)
	}
	if addr, err = parseHex(parts[0]); err != nil {
		return
	}
	val, err = parseHex(parts[1])
	return
}

func parseAddrN(spec string, def int) (addr uint32, n int, err error) {
	parts := strings.SplitN(spec, ":", 2)
	if addr, err = parseHex(parts[0]); err != nil {
		return
	}
	n = def
	if len(parts) == 2 {
		n64, e := strconv.Atoi(parts[1])
		if e != nil {
			return 0, 0, e
		}
		n = n64
	}
	return
}
