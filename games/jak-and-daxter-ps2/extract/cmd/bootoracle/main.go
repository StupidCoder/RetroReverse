// bootoracle boots Jak and Daxter's disc on the PS2 machine model and lets you watch
// what it does.
//
// It follows the boot-oracle contract in STANDARDS.md §3, and adds the instruments
// ORACLES.md names as worth porting. Three of them are sharper here than anywhere
// else in the repository, because this game's boot ELF ships its symbol table: every
// address the oracle prints — a breakpoint, a trace line, a backtrace frame, an
// unmodelled kernel call — is named.
//
// Usage:
//
//	bootoracle -image DISC.iso [-steps N] [-trace] [-bp ADDR] [-logpc ADDR] ...
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
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
	image := flag.String("image", "", "disc image (.iso)")
	exeName := flag.String("exe", "", "boot a specific executable rather than the one SYSTEM.CNF names")
	stepsS := flag.String("steps", "100000000", "instruction budget (hex or decimal)")
	trace := flag.Bool("trace", false, "trace execution")
	tracen := flag.Int("tracen", 200, "limit traced instructions")
	tracefrom := flag.String("tracefrom", "", "start tracing when this address is first reached (hex)")
	var bps multiFlag
	flag.Var(&bps, "bp", "halting breakpoint (hex or symbol name); repeatable")
	var logpcs multiFlag
	flag.Var(&logpcs, "logpc", "non-halting breakpoint: log registers and continue (hex or symbol); repeatable")
	var watches multiFlag
	flag.Var(&watches, "watch", "write-watch ADDR[:LEN] (hex); repeatable")
	var rwatches multiFlag
	flag.Var(&rwatches, "rwatch", "read-watch ADDR[:LEN] (hex); repeatable")
	watchn := flag.Int("watchn", 100, "limit watch reports")
	savestate := flag.String("savestate", "", "write a snapshot at the end of the run")
	loadstate := flag.String("loadstate", "", "start from a snapshot")
	poke := flag.String("poke", "", "write ADDR:VALUE (hex) after loading, before running")
	dis := flag.String("dis", "", "disassemble ADDR[:N] and exit (hex)")
	dump := flag.String("dump", "", "hex-dump ADDR:LEN and exit (hex)")
	files := flag.Bool("files", false, "list the disc's files and exit")
	verbose := flag.Bool("v", false, "log every kernel call as it happens")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: bootoracle -image DISC.iso [-steps N] [-trace] [-bp ADDR] ...")
		os.Exit(2)
	}
	if err := run(cfg{
		image: *image, exeName: *exeName, steps: *stepsS,
		trace: *trace, tracen: *tracen, tracefrom: *tracefrom,
		bps: bps, logpcs: logpcs, watches: watches, rwatches: rwatches, watchn: *watchn,
		savestate: *savestate, loadstate: *loadstate, poke: *poke,
		dis: *dis, dump: *dump, files: *files, verbose: *verbose,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

type cfg struct {
	image, exeName, steps                 string
	trace                                 bool
	tracen                                int
	tracefrom                             string
	bps, logpcs, watches, rwatches        multiFlag
	watchn                                int
	savestate, loadstate, poke, dis, dump string
	files, verbose                        bool
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

// parseAddr accepts a hex address or a symbol name from the executable, so a
// breakpoint can be set on "KernelCheckAndDispatch__Fv" rather than on a number
// looked up by hand.
func parseAddr(m *ps2.Machine, s string) (uint32, error) {
	if a, err := hx(s); err == nil {
		return a, nil
	}
	if exe := m.Exe(); exe != nil {
		for _, sym := range exe.Symbols {
			if sym.Name == s || strings.HasPrefix(sym.Name, s+"__") {
				return sym.Addr, nil
			}
		}
	}
	return 0, fmt.Errorf("%q is neither a hex address nor a symbol in this executable", s)
}

func parseRange(s string) (lo, ln uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = hx(parts[0]); err != nil {
		return
	}
	ln = 4
	if len(parts) == 2 {
		ln, err = hx(parts[1])
	}
	return
}

func run(c cfg) error {
	raw, err := os.ReadFile(c.image)
	if err != nil {
		return err
	}
	sum := fmt.Sprintf("%x", md5.Sum(raw))

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

	// SYSTEM.CNF names the executable to boot. Reading it rather than assuming the
	// filename is the difference between a tool that works on this disc and one that
	// works on a PS2 disc.
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

	m := ps2.NewMachine()
	m.SetImageHash(sum)
	m.SetVolume(vol)
	m.LoadExecutable(exe)
	fmt.Fprintf(os.Stderr, "%s", exe.Describe())

	if c.loadstate != "" {
		if err := m.LoadStateFile(c.loadstate); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "loaded state from %s\n", c.loadstate)
	}

	// The probes that answer without running: they operate on the machine as loaded
	// (or as restored from a savestate) and exit.
	if c.dis != "" {
		addr, n, err := parseRange(c.dis)
		if err != nil {
			return err
		}
		if n < 8 {
			n = 32 * 4
		}
		for a := addr; a < addr+n; a += 4 {
			fmt.Printf("%-40s %08X  %s\n", m.Sym(a), a, m.DisasmAt(a))
		}
		return nil
	}
	if c.dump != "" {
		addr, n, err := parseRange(c.dump)
		if err != nil {
			return err
		}
		b := m.ReadMem(addr, int(n))
		for i := 0; i < len(b); i += 16 {
			end := i + 16
			if end > len(b) {
				end = len(b)
			}
			fmt.Printf("%08X  % x\n", addr+uint32(i), b[i:end])
		}
		return nil
	}

	if c.poke != "" {
		parts := strings.SplitN(c.poke, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("bad -poke %q (want ADDR:VALUE)", c.poke)
		}
		a, err1 := hx(parts[0])
		v, err2 := hx(parts[1])
		if err1 != nil || err2 != nil {
			return fmt.Errorf("bad -poke %q", c.poke)
		}
		m.Write32(a, v)
		fmt.Fprintf(os.Stderr, "poked 0x%08X = 0x%08X\n", a, v)
	}

	for _, s := range c.bps {
		a, err := parseAddr(m, s)
		if err != nil {
			return err
		}
		m.SetBreakpoint(a)
		fmt.Fprintf(os.Stderr, "breakpoint at %s (0x%08X)\n", m.Sym(a), a)
	}

	// -logpc: a non-halting breakpoint. The workhorse for "how often, and with what,
	// does this routine run?" across a boot too long to trace.
	logAt := map[uint32]bool{}
	for _, s := range c.logpcs {
		a, err := parseAddr(m, s)
		if err != nil {
			return err
		}
		logAt[a] = true
		fmt.Fprintf(os.Stderr, "logging calls to %s (0x%08X)\n", m.Sym(a), a)
	}

	watchHits := 0
	if len(c.watches) > 0 {
		lo, ln, err := parseRange(c.watches[0])
		if err != nil {
			return err
		}
		m.WatchLo, m.WatchHi = lo, lo+ln
		m.OnWrite = func(addr, val, pc uint32) {
			if watchHits++; watchHits <= c.watchn {
				fmt.Printf("write 0x%08X = 0x%08X   from %s\n", addr, val, m.Sym(pc))
			}
		}
	}
	if len(c.rwatches) > 0 {
		lo, ln, err := parseRange(c.rwatches[0])
		if err != nil {
			return err
		}
		m.RWatchLo, m.RWatchHi = lo, lo+ln
		m.OnRead = func(addr, val, pc uint32) {
			if watchHits++; watchHits <= c.watchn {
				fmt.Printf("read  0x%08X = 0x%08X   from %s\n", addr, val, m.Sym(pc))
			}
		}
	}

	var traceFrom uint32
	tracing := c.trace
	if c.tracefrom != "" {
		if traceFrom, err = parseAddr(m, c.tracefrom); err != nil {
			return err
		}
		tracing = false
	}

	traced := 0
	if tracing || traceFrom != 0 || len(logAt) > 0 {
		m.OnStep = func(mm *ps2.Machine, pc uint32) {
			if traceFrom != 0 && pc == traceFrom {
				tracing = true
			}
			if logAt[pc] {
				fmt.Printf("%-44s a0=%08X a1=%08X a2=%08X a3=%08X v0=%08X  ra=%s\n",
					mm.Sym(pc),
					uint32(mm.CPU.Reg(4)), uint32(mm.CPU.Reg(5)),
					uint32(mm.CPU.Reg(6)), uint32(mm.CPU.Reg(7)),
					uint32(mm.CPU.Reg(2)), mm.Sym(uint32(mm.CPU.Reg(31))))
				// Any argument register that points at a string names what the game asked
				// for, which is usually the whole answer.
				for i, r := range []uint32{4, 5, 6, 7} {
					if s := mm.CString(uint32(mm.CPU.Reg(r))); isPrintable(s) {
						fmt.Printf("    a%d -> %q\n", i, s)
					}
				}
			}
			if tracing && traced < c.tracen {
				traced++
				fmt.Printf("%08X  %-40s %s\n", pc, mm.DisasmAt(pc), mm.Sym(pc))
			}
		}
	}

	if c.verbose {
		os.Setenv("PS2_SYSCALL_TRACE", "1")
	}

	steps, err := parseCount(c.steps)
	if err != nil {
		return fmt.Errorf("bad -steps %q", c.steps)
	}

	res := m.Run(steps)

	fmt.Println()
	fmt.Println(res)
	fmt.Printf("reached: %s\n", m.Sym(res.PC))
	fmt.Printf("vblanks: %d\n", m.VBlanks())
	if tty := m.TTY(); tty != "" {
		fmt.Printf("\n--- the game's own output ---\n%s\n", tty)
	}
	fmt.Println()
	fmt.Print(m.SyscallCensus())
	fmt.Println()
	fmt.Print(m.Threads())
	fmt.Println()
	fmt.Print(m.HardwareCensus())
	if len(m.Log) > 0 {
		fmt.Println("\nlog:")
		for _, l := range m.Log {
			fmt.Println(" ", l)
		}
	}

	if c.savestate != "" {
		if err := m.SaveStateFile(c.savestate); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\nwrote state to %s\n", c.savestate)
	}
	return nil
}

// isPrintable reports whether a string read out of guest memory looks like text
// rather than the first few bytes of a pointer.
func isPrintable(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			return false
		}
	}
	return true
}
