// Command bootoracle boots Ridge Racer V on the PS2 oracle (tools/platform/ps2) and
// gives the instruments to look at what it does.
//
// Ridge Racer V's disc is different from Jak's in the one way that matters to the boot:
// its IOP boot image (IOPRP15.IMG) is an *update* that carries only four kernel modules,
// where Jak's carries all twelve. The other eight come from the console ROM, so this
// oracle needs a BIOS (-bios) — the whole reason the platform grew SetBIOS. It is also a
// plain ELF + archive disc: no GOAL, no runtime linker, so this oracle is much smaller
// than Jak's. It is deliberately a focused boot-and-look tool; the interactive debugger
// (framedbg, via the ps2adapter) is the richer front end and reads the BIOS out of this
// game's debug.json.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/ps2"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func main() {
	image_ := flag.String("image", "", "disc image (.iso or raw .bin)")
	bios := flag.String("bios", "", "console ROM (rom0, e.g. scph10000.bin) — needed because RRV's IOPRP image carries only some kernel modules")
	exeName := flag.String("exe", "", "boot a specific executable rather than the one SYSTEM.CNF names")
	stepsS := flag.String("steps", "300000000", "instruction budget (hex or decimal)")
	files := flag.Bool("files", false, "list the disc's files and exit")
	syms := flag.Bool("syms", false, "list the boot ELF's symbols and exit")
	dis := flag.String("dis", "", "disassemble ADDR[:N] and exit (hex)")
	dump := flag.String("dump", "", "hex-dump ADDR:LEN and exit (hex)")
	loadstate := flag.String("loadstate", "", "start from a snapshot")
	savestate := flag.String("savestate", "", "write a snapshot at the end of the run")
	poke := flag.String("poke", "", "write ADDR:VALUE (hex) after loading, before running")
	scan := flag.String("scan", "", "scan EE memory for 32-bit WORD[:MASK] (hex) and name every hit")
	gsFrame := flag.String("gsframe", "", "write the frame the GS would scan out to FILE.png at the end of the run")
	eeProf := flag.Int("eeprof", 0, "sample the EE's PC every N steps and report where the time goes, by symbol")
	pad := flag.String("pad", "", "press controller buttons: BUTTON@VBLANK[:HOLD],... (default hold 30 vblanks)")
	verbose := flag.Bool("v", false, "log every kernel call as it happens")

	var bps, logpcs, watches, rwatches, gsFBs, gsRegs multiFlag
	flag.Var(&bps, "bp", "halting breakpoint (hex); repeatable")
	flag.Var(&logpcs, "logpc", "non-halting breakpoint: log GPRs and continue (hex); repeatable")
	flag.Var(&watches, "watch", "write-watch ADDR[:LEN] (hex); repeatable")
	flag.Var(&rwatches, "rwatch", "read-watch ADDR[:LEN] (hex); repeatable")
	watchn := flag.Int("watchn", 100, "limit watch reports")
	flag.Var(&gsFBs, "gsfb", "dump a PSMCT32 buffer of GS memory as BASE:FBW:H:FILE.png; repeatable")
	flag.Var(&gsRegs, "gsreg", "log the first N writes to a GS register as REG[:N] (REG hex, e.g. 0x4C = FRAME_1)")
	iopThreads := flag.Bool("iopthreads", false, "walk THREADMAN's control blocks and report every thread's state")

	flag.Parse()

	if err := run(cfg{
		image: *image_, bios: *bios, exeName: *exeName, steps: *stepsS,
		files: *files, syms: *syms, dis: *dis, dump: *dump, scan: *scan,
		loadstate: *loadstate, savestate: *savestate, poke: *poke,
		gsFrame: *gsFrame, eeProf: *eeProf, pad: *pad, verbose: *verbose,
		bps: bps, logpcs: logpcs, watches: watches, rwatches: rwatches, watchn: *watchn,
		gsFBs: gsFBs, gsRegs: gsRegs, iopThreads: *iopThreads,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

type cfg struct {
	image, bios, exeName, steps    string
	files, syms, verbose           bool
	dis, dump, scan                string
	loadstate, savestate, poke     string
	gsFrame                        string
	eeProf                         int
	pad                            string
	bps, logpcs, watches, rwatches multiFlag
	watchn                         int
	gsFBs, gsRegs                  multiFlag
	iopThreads                     bool
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

func parseCount(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		v, err := hx(s)
		return uint64(v), err
	}
	return strconv.ParseUint(s, 10, 64)
}

// parseAddrLen reads "ADDR" or "ADDR:LEN" (hex), defaulting len to def.
func parseAddrLen(s string, def uint32) (uint32, uint32, error) {
	a, l, ok := strings.Cut(s, ":")
	addr, err := hx(a)
	if err != nil {
		return 0, 0, err
	}
	n := def
	if ok {
		if n, err = hx(l); err != nil {
			return 0, 0, err
		}
	}
	return addr, n, nil
}

func run(c cfg) error {
	if c.image == "" {
		return fmt.Errorf("no -image given")
	}
	raw, err := os.ReadFile(c.image)
	if err != nil {
		return err
	}
	vol, err := iso9660.OpenBytes(raw)
	if err != nil {
		return err
	}

	if c.files {
		return vol.Walk(func(e iso9660.Entry) error {
			fmt.Println(e)
			return nil
		})
	}

	// SYSTEM.CNF names the executable to boot; read it rather than assume the filename.
	exePath := c.exeName
	if exePath == "" {
		cnf, err := vol.ReadFile("SYSTEM.CNF")
		if err != nil {
			return fmt.Errorf("reading SYSTEM.CNF: %w", err)
		}
		for _, line := range strings.Split(string(cnf), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(line), "BOOT2"); ok {
				exePath = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "="))
				break
			}
		}
		if exePath == "" {
			return fmt.Errorf("SYSTEM.CNF names no BOOT2 executable")
		}
		fmt.Fprintf(os.Stderr, "SYSTEM.CNF: BOOT2=%s\n", exePath)
	}

	elfRaw, err := vol.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", exePath, err)
	}
	exe, err := ps2.LoadELF(elfRaw)
	if err != nil {
		return err
	}
	if c.syms {
		for _, s := range exe.Symbols {
			kind := " "
			if s.Func {
				kind = "F"
			}
			fmt.Printf("%08X %6d %s %s\n", s.Addr, s.Size, kind, s.Name)
		}
		return nil
	}

	m := ps2.NewMachine()
	m.SetVolume(vol)
	if c.bios != "" {
		biosRaw, err := os.ReadFile(c.bios)
		if err != nil {
			return fmt.Errorf("reading BIOS %s: %w", c.bios, err)
		}
		m.SetBIOS(biosRaw)
	}
	m.LoadExecutable(exe)
	fmt.Fprintf(os.Stderr, "%s", exe.Describe())
	if c.verbose {
		os.Setenv("PS2_SYSCALL_TRACE", "1")
	}

	// Static inspection that needs no run.
	if c.dis != "" {
		addr, n, err := parseAddrLen(c.dis, 1)
		if err != nil {
			return fmt.Errorf("bad -dis %q", c.dis)
		}
		for i := uint32(0); i < n; i++ {
			a := addr + i*4
			fmt.Printf("%-40s %08X  %s\n", m.Sym(a), a, m.DisasmAt(a))
		}
		return nil
	}
	if c.dump != "" {
		addr, n, err := parseAddrLen(c.dump, 0x40)
		if err != nil {
			return fmt.Errorf("bad -dump %q", c.dump)
		}
		b := m.ReadMem(addr, int(n))
		for i := 0; i < len(b); i += 16 {
			end := i + 16
			if end > len(b) {
				end = len(b)
			}
			fmt.Printf("%08X ", addr+uint32(i))
			for j := i; j < end; j++ {
				fmt.Printf(" %02x", b[j])
			}
			fmt.Println()
		}
		return nil
	}

	if c.loadstate != "" {
		if err := m.LoadStateFile(c.loadstate); err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
	} else if err := m.RebootIOP(); err != nil {
		return fmt.Errorf("bringing the IOP up before the EE runs: %w", err)
	}

	if c.iopThreads {
		fmt.Print(m.IOP.IOPThreads())
	}

	if c.scan != "" {
		word, mask := c.scan, "0xFFFFFFFF"
		if w, mk, ok := strings.Cut(c.scan, ":"); ok {
			word, mask = w, mk
		}
		wv, err1 := hx(word)
		mv, err2 := hx(mask)
		if err1 != nil || err2 != nil {
			return fmt.Errorf("bad -scan %q (want WORD[:MASK])", c.scan)
		}
		hits := 0
		for a := uint32(0); a < 0x02000000; a += 4 {
			if m.Read32(a)&mv == wv&mv {
				fmt.Printf("%08X  %08X  %s\n", a, m.Read32(a), m.Sym(a))
				hits++
			}
		}
		fmt.Printf("scan: %d hits\n", hits)
		return nil
	}

	if c.poke != "" {
		a, v, ok := strings.Cut(c.poke, ":")
		av, err1 := hx(a)
		vv, err2 := hx(v)
		if !ok || err1 != nil || err2 != nil {
			return fmt.Errorf("bad -poke %q (want ADDR:VALUE)", c.poke)
		}
		m.Write32(av, vv)
	}

	if c.pad != "" {
		script, err := parsePadScript(c.pad)
		if err != nil {
			return err
		}
		m.PadScript = script
	}

	for _, spec := range c.gsRegs {
		parts := strings.SplitN(spec, ":", 2)
		r, err := hx(parts[0])
		if err != nil {
			return fmt.Errorf("bad -gsreg %q", spec)
		}
		n := 40
		if len(parts) > 1 {
			n, _ = strconv.Atoi(parts[1])
		}
		m.GSRegLog, m.GSRegLogN = uint8(r), n
	}

	// Watches. The machine carries one write range and one read range (lo..hi), each
	// with a callback — the same shape Jak's oracle uses.
	watchHits := 0
	if len(c.watches) > 0 {
		lo, n, err := parseAddrLen(c.watches[0], 4)
		if err != nil {
			return fmt.Errorf("bad -watch %q", c.watches[0])
		}
		m.WatchLo, m.WatchHi = lo, lo+n
		m.OnWrite = func(addr, val, pc uint32) {
			if watchHits++; watchHits <= c.watchn {
				fmt.Printf("write 0x%08X = 0x%08X   from %s\n", addr, val, m.Sym(pc))
			}
		}
	}
	if len(c.rwatches) > 0 {
		lo, n, err := parseAddrLen(c.rwatches[0], 4)
		if err != nil {
			return fmt.Errorf("bad -rwatch %q", c.rwatches[0])
		}
		m.RWatchLo, m.RWatchHi = lo, lo+n
		m.OnRead = func(addr, val, pc uint32) {
			if watchHits++; watchHits <= c.watchn {
				fmt.Printf("read  0x%08X = 0x%08X   from %s\n", addr, val, m.Sym(pc))
			}
		}
	}

	for _, b := range c.bps {
		a, err := hx(b)
		if err != nil {
			return fmt.Errorf("bad -bp %q", b)
		}
		m.SetBreakpoint(a)
		fmt.Fprintf(os.Stderr, "breakpoint at %s (0x%08X)\n", m.Sym(a), a)
	}
	logAt := map[uint32]bool{}
	for _, l := range c.logpcs {
		a, err := hx(l)
		if err != nil {
			return fmt.Errorf("bad -logpc %q", l)
		}
		logAt[a] = true
		fmt.Fprintf(os.Stderr, "logging registers at 0x%08X\n", a)
	}

	// The run. eeprof and logpc both ride OnStep.
	prof := map[uint32]int{}
	profTick, profTotal := 0, 0
	if len(logAt) > 0 || c.eeProf > 0 {
		m.OnStep = func(mm *ps2.Machine, pc uint32) {
			if c.eeProf > 0 {
				if profTick++; profTick >= c.eeProf {
					profTick = 0
					prof[pc]++
					profTotal++
				}
			}
			if logAt[pc] {
				fmt.Printf("%-30s a0=%08X a1=%08X a2=%08X a3=%08X v0=%08X  ra=%s\n",
					mm.Sym(pc),
					uint32(mm.CPU.Reg(4)), uint32(mm.CPU.Reg(5)),
					uint32(mm.CPU.Reg(6)), uint32(mm.CPU.Reg(7)),
					uint32(mm.CPU.Reg(2)), mm.Sym(uint32(mm.CPU.Reg(31))))
			}
		}
	}

	steps, err := parseCount(c.steps)
	if err != nil {
		return fmt.Errorf("bad -steps %q", c.steps)
	}
	res := m.Run(steps)
	fmt.Println(res.String())
	fmt.Printf("reached: %s\n", m.Sym(res.PC))

	if c.eeProf > 0 {
		type pc struct {
			addr  uint32
			count int
		}
		var top []pc
		for a, n := range prof {
			top = append(top, pc{a, n})
		}
		for i := 0; i < len(top); i++ {
			for j := i + 1; j < len(top); j++ {
				if top[j].count > top[i].count {
					top[i], top[j] = top[j], top[i]
				}
			}
		}
		fmt.Printf("\nEE profile: %d samples, every %d steps\n", profTotal, c.eeProf)
		for i := 0; i < len(top) && i < 30; i++ {
			fmt.Printf("  %5.2f%%  0x%08X  %s\n", 100*float64(top[i].count)/float64(profTotal), top[i].addr, m.Sym(top[i].addr))
		}
	}

	for _, spec := range c.gsFBs {
		if err := writeGSBuffer(m, spec); err != nil {
			return err
		}
	}
	if c.gsFrame != "" {
		if err := writeGSFrame(m, c.gsFrame); err != nil {
			return err
		}
	}
	if c.savestate != "" {
		if err := m.SaveStateFile(c.savestate); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		fmt.Fprintf(os.Stderr, "wrote savestate %s\n", c.savestate)
	}
	return nil
}

// parsePadScript reads the -pad schedule: BUTTON@VBLANK[:HOLD], comma-separated.
func parsePadScript(s string) ([]ps2.PadPress, error) {
	var script []ps2.PadPress
	for _, ent := range strings.Split(s, ",") {
		ent = strings.TrimSpace(ent)
		if ent == "" {
			continue
		}
		name, rest, ok := strings.Cut(ent, "@")
		if !ok {
			return nil, fmt.Errorf("bad -pad entry %q (want BUTTON@VBLANK[:HOLD])", ent)
		}
		bits, ok := ps2.PadButton(strings.TrimSpace(name))
		if !ok {
			return nil, fmt.Errorf("bad -pad button %q (want one of %s)", name, strings.Join(ps2.PadButtonNames(), ", "))
		}
		atS, holdS, _ := strings.Cut(rest, ":")
		at, err := strconv.ParseUint(atS, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("bad -pad vblank %q", atS)
		}
		hold := uint64(30)
		if holdS != "" {
			if hold, err = strconv.ParseUint(holdS, 10, 32); err != nil {
				return nil, fmt.Errorf("bad -pad hold %q", holdS)
			}
		}
		script = append(script, ps2.PadPress{Buttons: bits, At: uint32(at), Hold: uint32(hold)})
	}
	return script, nil
}

func writeGSFrame(m *ps2.Machine, path string) error {
	pix, w, h := m.GSFrame()
	if pix == nil {
		fmt.Println("gsframe: the GS has no displayable frame (no DISPFB set, or an unread format)")
		return nil
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	nonBlack := 0
	for i := 0; i < len(pix); i += 4 {
		if pix[i] != 0 || pix[i+1] != 0 || pix[i+2] != 0 {
			nonBlack++
		}
	}
	fmt.Printf("gsframe: wrote %dx%d to %s (%d of %d pixels non-black)\n", w, h, path, nonBlack, w*h)
	return nil
}

func writeGSBuffer(m *ps2.Machine, spec string) error {
	parts := strings.SplitN(spec, ":", 4)
	if len(parts) != 4 {
		return fmt.Errorf("bad -gsfb %q (want BASE:FBW:H:FILE.png)", spec)
	}
	base, err1 := hx(parts[0])
	fbw, err2 := strconv.Atoi(parts[1])
	h, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("bad -gsfb %q", spec)
	}
	pix, w := m.GSBuffer(base, uint32(fbw), h)
	if pix == nil {
		fmt.Printf("gsfb %s: nothing to dump\n", spec)
		return nil
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(parts[3])
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	fmt.Printf("gsfb: wrote %dx%d to %s\n", w, h, parts[3])
	return nil
}
