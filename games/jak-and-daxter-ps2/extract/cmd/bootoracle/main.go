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
	iopOnly := flag.Bool("iop", false, "boot the IOP alone, load its modules, and report — the second processor's bring-up harness")
	iopMods := flag.String("iopmods", "", "extra IRX modules to load after the kernel, comma-separated (e.g. SIO2MAN,PADMAN,OVERLORD)")
	iopDis := flag.String("iopdis", "", "disassemble IOP memory at ADDR[:N] after the IOP boots (hex or symbol) — reads the modules as loaded, with every kernel stub named")
	iopIO := flag.Bool("iopio", false, "trace every IOP peripheral-register access, with the routine that made it")
	iopION := flag.Int("iopion", 400, "limit traced IOP register accesses")
	iopWatch := flag.String("iopwatch", "", "write-watch on IOP memory: ADDR[:LEN] (hex)")
	iopTrap := flag.String("ioptrap", "", "halt the IOP when it reaches ADDR (hex or symbol) and print the instructions that led there")
	iopCalls := flag.Int("iopcalls", 0, "trace the first N calls the IOP's modules make through their import stubs — the protocol between the modules")
	iopDump := flag.String("iopdump", "", "dump IOP memory as words at ADDR[:LEN] (hex or symbol), naming any word that points into a module")
	iopCallsFrom := flag.String("iopcallsfrom", "", "only trace stub calls once this module has started (e.g. 989SND.IRX)")
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
		iopOnly: *iopOnly, iopMods: *iopMods, iopDis: *iopDis,
		iopIO: *iopIO, iopION: *iopION, iopWatch: *iopWatch, iopTrap: *iopTrap,
		iopCalls: *iopCalls, iopCallsFrom: *iopCallsFrom,
		iopDump: *iopDump,
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
	iopOnly                               bool
	iopMods, iopDis                       string
	iopIO                                 bool
	iopION                                int
	iopWatch                              string
	iopTrap                               string
	iopCalls                              int
	iopCallsFrom                          string
	iopDump                               string
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

// parseSymRange is parseRange over the executable, and it takes a symbol name where the
// plain one takes only a number. The boot ELF ships its symbol table, so there is no reason
// `-dis sceSifInitRpc` should have to be spelled as an address the reader looked up by hand —
// the IOP's disassembler has taken symbols since it was written, and this is that parity.
func parseSymRange(m *ps2.Machine, s string) (lo, ln uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = parseAddr(m, parts[0]); err != nil {
		return
	}
	ln = 4
	if len(parts) == 2 {
		ln, err = hx(parts[1])
	}
	return
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

// parseIOPRange is parseRange over IOP memory, and it accepts a symbol name where the
// EE's takes only a number. The modules the game ships are not stripped, so
// `-iopdis DMA_SendToSPUAndSync:40` says what it means — and says it against whatever
// address the module landed at this run, which changes whenever the load order does.
func parseIOPRange(m *ps2.Machine, s string) (lo, ln uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = hx(parts[0]); err != nil {
		a, ok := m.IOP.SymAddr(parts[0])
		if !ok {
			return 0, 0, fmt.Errorf("%q is neither a hex address nor a symbol in any loaded IOP module", parts[0])
		}
		lo, err = a, nil
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

	// The IOP's bring-up harness: boot the second processor on the disc's own kernel
	// modules, with the EE left standing at its entry point, and report what it did.
	//
	// It exists because the IOP is a machine you can be wrong about for a very long
	// time if the only way to reach it is through a three-billion-instruction EE boot.
	// This is the short way in.
	if c.iopOnly {
		return bootIOP(m, c)
	}

	// The full boot has an IOP in it — the game reboots it itself — so the second processor's
	// instruments belong here too, and the SIF is the reason: the conversation between the two
	// chips is the one thing the bring-up harness structurally cannot show.
	if err := armIOP(m, c); err != nil {
		return err
	}

	// The second processor is already running when the game's first instruction executes, and
	// the game proves it: before it ever reboots the IOP it initialises the SIF, binds to the
	// file-system and CD services and reads the disc — "Initializing CD drive", "Disk type 0" —
	// and only then prints "Rebooting IOP...". All of that is a conversation with an IOP that
	// the BIOS booted at power-on, and it happens before the game has named an image.
	//
	// We have no BIOS. What we have is the module set the game itself asks for a moment later,
	// on its own disc, so that is what the IOP is booted on here. It is not the ROM's kernel
	// and this does not pretend it is; it is Sony's kernel, off this disc, chosen by the game.
	// The reboot that follows is then the second boot, exactly as it is on the board.
	if err := m.RebootIOP(); err != nil {
		return fmt.Errorf("bringing the IOP up before the EE runs: %w", err)
	}

	if c.loadstate != "" {
		if err := m.LoadStateFile(c.loadstate); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "loaded state from %s\n", c.loadstate)
	}

	// The probes that answer without running: they operate on the machine as loaded
	// (or as restored from a savestate) and exit.
	if c.dis != "" {
		addr, n, err := parseSymRange(m, c.dis)
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
		addr, n, err := parseSymRange(m, c.dump)
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
	fmt.Print(m.SIFCensus())
	fmt.Println()
	fmt.Print(m.HardwareCensus())

	// The second processor, if the game brought it up. It does now — the reboot is a SIF
	// packet the game sends — so the full boot has an IOP in it, and an IOP nobody reports on
	// is an IOP that can be blocked, halted or idle for a whole run without anyone noticing.
	if m.IOP != nil {
		if tty := m.IOP.TTY(); tty != "" {
			fmt.Printf("\n--- what the IOP printed\n%s\n", tty)
		}
		fmt.Printf("\n--- %s", m.IOP.IOPInterrupts())
		if prof := m.IOP.IOPProfile(); prof != "" {
			fmt.Printf("\n--- %s", prof)
		}
	}

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

// bootIOP is the second processor's bring-up harness.
//
// It boots the IOP on IOPRP221.IMG — the kernel modules the game itself reboots it
// onto — then loads whatever else was asked for, and prints three things: what the
// modules printed, what they asked the kernel for that nothing answered, and what
// hardware they touched. The middle one is the work list.
// iopFreeRun is how long the harness lets the IOP run once its modules are up. It is
// long enough for the scheduler to make many passes and for every thread to reach the
// place where it waits for the EE, which is where a healthy IOP with no EE ends up.
const iopFreeRun = 20_000_000

// armIOP attaches the second processor's instruments.
//
// It is called by both harnesses, and it has to be, because the IOP the *game* boots is the
// same processor as the one `-iop` boots — and the interesting half of its life, the half
// where it talks to the EE, only happens in the full boot. An instrument that exists only in
// the bring-up harness cannot see the conversation it was built to watch.
//
// Everything here hangs off OnIOPStart because the IOP does not exist yet: it is created
// inside the reboot, and its modules are driving hardware inside the very entry points the
// reboot calls. By the time a caller could reach for the processor, the boot is over.
func armIOP(m *ps2.Machine, c cfg) error {
	// The register trace has to be armed before the IOP exists, because StartIOP is
	// inside RebootIOP and the modules begin driving hardware in their entry points —
	// TIMEMANI is programming a timer before the boot is four modules old.
	if c.iopIO {
		n := 0
		m.OnIOPStart = func(p *ps2.IOP) {
			p.OnIO = func(addr, val uint32, write bool, pc uint32) {
				if n >= c.iopION {
					return
				}
				n++
				op := "read "
				if write {
					op = "write"
				}
				fmt.Printf("  io %s 0x%08X = %08X  %-10s %s\n",
					op, addr, val, ps2.IOPRegionName(addr), p.Sym(pc))
			}
		}
	}

	// The stub-call trace: what the modules ask each other for. It can be held back until a
	// named module starts, because the interesting protocol is usually the last module's and
	// the ones before it produce thousands of calls to wade through.
	if c.iopCalls > 0 {
		n := 0
		armed := c.iopCallsFrom == ""
		prev := m.OnIOPStart
		m.OnIOPStart = func(p *ps2.IOP) {
			if prev != nil {
				prev(p)
			}
			p.OnCall = func(name string, args [4]uint32, from uint32) {
				if !armed || n >= c.iopCalls {
					return
				}
				n++
				fmt.Printf("  call %-16s (0x%X, 0x%X, 0x%X, 0x%X)  from %s\n",
					name, args[0], args[1], args[2], args[3], p.Sym(from))
			}
		}
		m.OnIOPModule = func(p *ps2.IOP, name string) {
			if name == c.iopCallsFrom {
				armed = true
			}
		}
	}

	// The trap. It takes a hex address or a symbol, and a symbol cannot be resolved here —
	// the IOP has no modules in it yet. So the name is handed to the loader, which arms the
	// trap as soon as the module carrying that symbol is placed.
	if c.iopTrap != "" {
		prev := m.OnIOPStart
		m.OnIOPStart = func(p *ps2.IOP) {
			if prev != nil {
				prev(p)
			}
			if a, e := hx(c.iopTrap); e == nil {
				p.Trap = a
			} else {
				p.TrapSym = c.iopTrap
			}
		}
	}

	if c.iopWatch != "" {
		lo, ln, e := parseRange(c.iopWatch)
		if e != nil {
			return e
		}
		n := 0
		prev := m.OnIOPStart
		m.OnIOPStart = func(p *ps2.IOP) {
			if prev != nil {
				prev(p)
			}
			p.WatchLo, p.WatchHi = lo, lo+ln
			p.OnWrite = func(addr, val, pc uint32) {
				if n >= 200 {
					return
				}
				n++
				fmt.Printf("  iop write 0x%08X = %02X  from %s\n", addr, val, p.Sym(pc))
			}
		}
	}
	return nil
}

func bootIOP(m *ps2.Machine, c cfg) error {
	extra, dis := c.iopMods, c.iopDis

	if err := armIOP(m, c); err != nil {
		return err
	}

	// A snapshot resumes the second processor where it was left, with its modules resident,
	// its threads scheduled and its timers running — which turns a boot that costs minutes
	// into one that costs nothing, and is the whole reason the state carries the IOP.
	var err error
	if c.loadstate != "" {
		if err = m.LoadStateFile(c.loadstate); err != nil {
			return err
		}
		if m.IOP == nil {
			return fmt.Errorf("%s holds no IOP: it was taken before the second processor was started", c.loadstate)
		}
		fmt.Printf("resumed the IOP from %s: %d modules, %d instructions in\n",
			c.loadstate, len(m.IOP.Modules()), m.IOP.Steps())
	} else {
		err = m.RebootIOP()
		if err != nil {
			fmt.Printf("\nthe IOP's boot stopped: %v\n", err)
		}
	}

	if err == nil && c.loadstate == "" && extra != "" {
		for _, name := range strings.Split(extra, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			path := "/DRIVERS/" + strings.ToUpper(name)
			if !strings.HasSuffix(path, ".IRX") {
				path += ".IRX"
			}
			if err = m.IOP.LoadModuleFromDisc(path); err != nil {
				fmt.Printf("\nloading %s: %v\n", name, err)
				break
			}
		}
	}

	for _, mod := range m.IOP.Modules() {
		fmt.Printf("  loaded %-14s at 0x%08X  %6d bytes\n", mod.Name, mod.Base, mod.Size)
	}

	// Every module is loaded and every entry point has returned. Now let the second
	// processor simply run: its threads are THREADMAN's to schedule, its interrupts come
	// from its own timers, and nothing is driving it. This is the IOP being a processor
	// rather than a library, and it is the state it spends a whole game in.
	if err == nil {
		fmt.Printf("\nrunning the IOP on its own for %d instructions\n", iopFreeRun)
		m.IOP.Run(iopFreeRun)
	}

	if c.savestate != "" {
		if e := m.SaveStateFile(c.savestate); e != nil {
			return e
		}
		fmt.Printf("wrote %s (the IOP at %d instructions, PC %s)\n",
			c.savestate, m.IOP.Steps(), m.IOP.Sym(m.IOP.CPU.PC))
	}

	fmt.Printf("\n--- the log\n")
	for _, l := range m.Log {
		fmt.Printf("  %s\n", l)
	}
	if tty := m.IOP.TTY(); tty != "" {
		fmt.Printf("\n--- what the IOP printed\n%s\n", tty)
	}
	fmt.Printf("\n--- %s", m.IOP.IOPInterrupts())
	if prof := m.IOP.IOPProfile(); prof != "" {
		fmt.Printf("\n--- %s", prof)
	}
	if census := m.IOP.IOPCensus(); census != "" {
		fmt.Printf("\n--- %s", census)
	}

	if dis != "" {
		addr, n, err := parseIOPRange(m, dis)
		if err != nil {
			return err
		}
		if n < 8 {
			n = 32 * 4
		}
		fmt.Printf("\n--- IOP memory as loaded\n")
		for a := addr; a < addr+n; a += 4 {
			fmt.Printf("  %-24s %08X  %s\n", m.IOP.Sym(a), a, m.IOP.DisasmAt(a))
		}
	}

	// -iopdump prints words rather than instructions, and it names any word that is an
	// address inside a loaded module. The IOP's interesting state is not code: it is the
	// control blocks the modules pass each other by pointer — a dispatch table whose entries
	// are {function, argument} says what a module will answer and what it will ignore, and it
	// is unreadable as instructions.
	if c.iopDump != "" {
		addr, n, err := parseIOPRange(m, c.iopDump)
		if err != nil {
			return err
		}
		fmt.Printf("\n--- IOP memory at 0x%08X\n", addr)
		for a := addr; a < addr+n; a += 4 {
			v := m.IOP.Read32(a)
			fmt.Printf("  %08X  +%-4d  %08X  %s\n", a, a-addr, v, m.IOP.Sym(v))
		}
	}
	return nil
}
