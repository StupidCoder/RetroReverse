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
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/ps2"
)

// padButtonBits names the controller's buttons by the bit each occupies in the pad
// protocol's two button bytes (little end first, active-high here — the wire inverts).
var padButtonBits = map[string]uint16{
	"SELECT": 0x0001, "L3": 0x0002, "R3": 0x0004, "START": 0x0008,
	"UP": 0x0010, "RIGHT": 0x0020, "DOWN": 0x0040, "LEFT": 0x0080,
	"L2": 0x0100, "R2": 0x0200, "L1": 0x0400, "R1": 0x0800,
	"TRIANGLE": 0x1000, "CIRCLE": 0x2000, "CROSS": 0x4000, "X": 0x4000, "SQUARE": 0x8000,
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
		bits, ok := padButtonBits[strings.ToUpper(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("bad -pad button %q", name)
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
	scan := flag.String("scan", "", "scan EE memory for 32-bit WORD[:MASK] (hex) and name every hit — finds every instruction that references a GOAL symbol cell")
	goalPS := flag.String("goalps", "", "walk a GOAL process tree from ROOT (hex or symbol, e.g. *active-pool*) and print every process — the walk search-process-tree performs, done from outside")
	files := flag.Bool("files", false, "list the disc's files and exit")
	syms := flag.Bool("syms", false, "list the boot ELF's symbols (addr size F name) and exit — the map of the engine's named C surface")
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
	iopThreads := flag.Bool("iopthreads", false, "walk THREADMAN's control blocks and report every thread's state and the PC it is parked at — the blocked-thread inspector for a deadlock")
	iopCallsFrom := flag.String("iopcallsfrom", "", "only trace stub calls once this module has started (e.g. 989SND.IRX)")
	var iopPokes multiFlag
	flag.Var(&iopPokes, "ioppoke", "write ADDR:VALUE (hex) into IOP memory every time the IOP finishes booting; repeatable. Sony's modules carry their own tracing behind a verbosity word — CDVDMAN's is at 0x29F90 — and turning one on makes a stripped module narrate itself")
	iopIELog := flag.String("iopielog", "", "log every IOP interrupt-enable event (suspend/resume/deliver/frame save+load) to FILE — the instrument for an enable bit lost across a thread switch")
	goalSyms := flag.String("goalsyms", "", "write the GOAL symbol table (name, address, value) to FILE at the end of the run — the runtime-linked engine's own names, read the way find_symbol_from_c reads them")
	eeProf := flag.Int("eeprof", 0, "sample the EE's PC every N steps and report where the time goes, by symbol (use with -goalnames to see engine code) — the only thing that tells an engine idling from an engine working")
	goalNames := flag.String("goalnames", "", "read a -goalsyms dump back in, so -dis/-bp/-logpc and every trace can name GOAL engine code (symbol values that point into RAM become function names)")
	gsFrame := flag.String("gsframe", "", "write the frame the GS would be scanning out (the DISPFB rectangle, deswizzled) to FILE.png at the end of the run")
	vu1In := flag.String("vu1in", "", "dump a VU1 program's input buffer (96 qw at TOP) at its next MSCAL — hex byte address; the in-place transforms destroy the input by kick time")
	gsPixel := flag.String("gspixel", "", "log the next N writes landing on window pixel X:Y[:N] of any render target, with the colour, blend inputs, target and producer — the 'who painted this pixel' instrument a uniform fill needs")
	pad := flag.String("pad", "", "press controller buttons: BUTTON@VBLANK[:HOLD],... (e.g. X@1100:30,START@1400:30; default hold 30 vblanks) — a digital pad sits in port 0 either way, this is what it reports pressed")
	gsBig := flag.Int("gsbig", 0, "print the first N completed GS primitives whose bounding box exceeds 1024px, naming the VU1 program or PATH that produced each — the huge-triangle hunter")
	gsVerts := flag.Int("gsverts", 0, "print the first N completed GS primitives with their exact vertex data (position, Z, RGBA, ST/Q) — one column per hypothesis: huge positions = transform bug, black RGBA = lighting bug, zero alpha or Q = unpack bug")
	gsReg := flag.String("gsreg", "", "log the first N writes to a GS register as REG[:N] (REG hex, e.g. 0x40 = SCISSOR_1), with the value and the producer — the instrument for a register value nobody admits to writing")
	vu1Data := flag.String("vu1data", "", "write VU1's data memory (as the VIF unpacked it) to FILE at the end of the run — the input side of a microprogram, where the matrix rows and the vertex block sit")
	vu0Data := flag.String("vu0data", "", "write VU0's data memory to FILE at the end of the run — where the vcallms palette lives")
	vu0Micro := flag.String("vu0micro", "", "write VU0's program memory (as VIF0 filled it) to FILE at the end of the run — where the EE's vcallms microprograms live")
	vu0Regs := flag.Bool("vu0regs", false, "print VU0's register file at the end of the run — the state a vcallms at a breakpoint would be issued with, for hand-executing a microprogram")
	vu1Micro := flag.String("vu1micro", "", "write VU1's program memory (as the VIF filled it) to FILE at the end of the run — the input for sizing up the vector unit")
	var gsFBs multiFlag
	flag.Var(&gsFBs, "gsfb", "dump a PSMCT32 buffer of GS memory as BASE:FBW:H:FILE.png (base = word address as the census prints, FBW in 64px units, H in pixels); repeatable")
	var gsTexs multiFlag
	flag.Var(&gsTexs, "gstex", "dump a texture as the sampler sees it (swizzle+CLUT+TEXA) as TEX0:FILE.png (TEX0 = the raw register word the draw census prints); repeatable")
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
		dis: *dis, dump: *dump, scan: *scan, goalPS: *goalPS, files: *files, syms: *syms, verbose: *verbose,
		iopOnly: *iopOnly, iopMods: *iopMods, iopDis: *iopDis,
		iopIO: *iopIO, iopION: *iopION, iopWatch: *iopWatch, iopTrap: *iopTrap,
		iopCalls: *iopCalls, iopCallsFrom: *iopCallsFrom, iopPokes: iopPokes,
		iopDump: *iopDump, iopThreads: *iopThreads, iopIELog: *iopIELog, goalSyms: *goalSyms, goalNames: *goalNames, eeProf: *eeProf, gsFrame: *gsFrame, gsVerts: *gsVerts, gsReg: *gsReg, gsBig: *gsBig, gsPixel: *gsPixel, pad: *pad, vu1In: *vu1In, vu1Micro: *vu1Micro, vu0Micro: *vu0Micro, vu0Data: *vu0Data, vu0Regs: *vu0Regs, vu1Data: *vu1Data,
		gsFBs: gsFBs, gsTexs: gsTexs,
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
	scan                                  string
	goalPS                                string
	files, syms, verbose                  bool
	iopOnly                               bool
	iopMods, iopDis                       string
	iopIO                                 bool
	iopION                                int
	iopWatch                              string
	iopTrap                               string
	iopCalls                              int
	iopCallsFrom                          string
	iopDump                               string
	iopThreads                            bool
	iopPokes                              multiFlag
	iopIELog                              string
	goalSyms                              string
	goalNames                             string
	eeProf                                int
	gsFrame                               string
	gsVerts                               int
	gsReg                                 string
	gsBig                                 int
	gsPixel                               string
	pad                                   string
	vu1In                                 string
	vu1Data                               string
	vu1Micro                              string
	vu0Micro                              string
	vu0Data                               string
	vu0Regs                               bool
	gsFBs                                 multiFlag
	gsTexs                                multiFlag
}

// ieLogFlush, if the interrupt-enable log is on, flushes it. It is set by armIOP and
// called once the run is over, from whichever harness ran.
var ieLogFlush func()

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

// goalName maps a GOAL symbol name to the address its value held in the -goalnames
// dump, so -bp/-logpc/-dis can be set on "display-loop" rather than a number.
var goalName = map[string]uint32{}

// loadGoalNames reads a -goalsyms dump back in: symbol values that point into RAM
// become names for the machine's Sym, and entries for parseAddr. The dump line is
// "CELLADDR OFFSET VALUE name"; the value is what names code, the cell is not code.
func loadGoalNames(m *ps2.Machine, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var syms []ps2.Symbol
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 4 {
			continue
		}
		v, err := strconv.ParseUint(fields[2], 16, 32)
		if err != nil {
			continue
		}
		val := uint32(v)
		name := fields[3]
		if off, err := strconv.ParseInt(fields[1], 10, 32); err == nil {
			goalCell[int32(off)] = name
		}
		// Only values that can be code name addresses: the GOAL heaps live above
		// the C kernel. Small ints, #f/#t (symbol addresses) and zeros do not.
		if val < 0x100000 || val >= 0x02000000 {
			continue
		}
		syms = append(syms, ps2.Symbol{Name: name, Addr: val, Func: true})
		if _, dup := goalName[name]; !dup {
			goalName[name] = val
		}
	}
	m.AddSymbols(syms)
	fmt.Fprintf(os.Stderr, "goalnames: %d named addresses from %s\n", len(syms), path)
	return sc.Err()
}

// goalCell maps a GOAL symbol-table offset (from $s7) to its name, for annotating
// the `lw $v1, -17128($s7)` idiom GOAL compiles every symbol reference into.
var goalCell = map[int32]string{}

// goalRefNote annotates a disassembled instruction that references OFFSET($s7)
// with the GOAL symbol that offset is — the difference between reading engine
// code and reading numbers.
func goalRefNote(m *ps2.Machine, a uint32) string {
	if len(goalCell) == 0 {
		return ""
	}
	w := m.Fetch32(a)
	op := w >> 26
	// The loads and stores GOAL uses on symbol cells (lw/lwu/sw/ld/sd/lq/sq and
	// daddiu for taking a symbol's address), base register $s7 (23).
	base := (w >> 21) & 31
	if base != 23 {
		return ""
	}
	switch op {
	case 0x23, 0x27, 0x2B, 0x37, 0x3F, 0x1E, 0x1F, 0x19, 0x09: // lw lwu sw ld sd lq sq daddiu addiu
	default:
		return ""
	}
	off := int32(int16(w))
	if name, ok := goalCell[off]; ok {
		return "    ; " + name
	}
	return ""
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
	if a, ok := goalName[s]; ok {
		return a, nil
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
	m.SetImageHash(sum)
	m.SetVolume(vol)
	m.LoadExecutable(exe)
	fmt.Fprintf(os.Stderr, "%s", exe.Describe())

	if c.goalNames != "" {
		if err := loadGoalNames(m, c.goalNames); err != nil {
			return err
		}
	}

	// The pokes are wired in here, before either path forks, because both need them: the
	// full boot applies them on the game's own IOP reboot, and the -iop bring-up harness
	// applies them on its boot and on a -loadstate resume. Populating them after the fork
	// (as this once did) left the harness path with an empty map, so -ioppoke silently did
	// nothing there — the one path where turning a module's tracing on matters most.
	if len(c.iopPokes) > 0 {
		m.IOPPokes = map[uint32]uint32{}
		for _, s := range c.iopPokes {
			parts := strings.SplitN(s, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("bad -ioppoke %q (want ADDR:VALUE)", s)
			}
			a, err1 := hx(parts[0])
			v, err2 := hx(parts[1])
			if err1 != nil || err2 != nil {
				return fmt.Errorf("bad -ioppoke %q", s)
			}
			m.IOPPokes[a] = v
		}
	}

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
			fmt.Printf("%-40s %08X  %s%s\n", m.Sym(a), a, m.DisasmAt(a), goalRefNote(m, a))
		}
		return nil
	}
	if c.scan != "" {
		parts := strings.SplitN(c.scan, ":", 2)
		word, err := hx(parts[0])
		if err != nil {
			return fmt.Errorf("bad -scan %q", c.scan)
		}
		mask := ^uint32(0)
		if len(parts) == 2 {
			if mask, err = hx(parts[1]); err != nil {
				return fmt.Errorf("bad -scan mask %q", parts[1])
			}
		}
		n := 0
		for a := uint32(0x80000); a < 0x02000000 && n < 500; a += 4 {
			if w := m.Read32(a); w&mask == word&mask {
				fmt.Printf("%08X  %08X  %s\n", a, w, m.Sym(a))
				n++
			}
		}
		fmt.Printf("scan: %d hits\n", n)
		return nil
	}
	if c.goalPS != "" {
		root, err := parseAddr(m, c.goalPS)
		if err != nil {
			return err
		}
		s7 := uint32(m.CPU.Reg(23))
		dumpGoalTree(m, s7, root, 0)
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

	if c.gsBig > 0 {
		m.GSBigDump = c.gsBig
	}
	if c.gsPixel != "" {
		parts := strings.Split(c.gsPixel, ":")
		if len(parts) < 2 {
			return fmt.Errorf("bad -gspixel %q (want X:Y[:N])", c.gsPixel)
		}
		x, err1 := strconv.Atoi(parts[0])
		y, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return fmt.Errorf("bad -gspixel %q", c.gsPixel)
		}
		n := 40
		if len(parts) > 2 {
			n, _ = strconv.Atoi(parts[2])
		}
		m.GSPixelX, m.GSPixelY, m.GSPixelN = int32(x), int32(y), n
	}
	if c.vu1In != "" {
		a, err := hx(c.vu1In)
		if err != nil {
			return fmt.Errorf("bad -vu1in %q", c.vu1In)
		}
		m.VU1DumpIn = int64(a)
	}
	if c.pad != "" {
		script, err := parsePadScript(c.pad)
		if err != nil {
			return err
		}
		m.PadScript = script
	}
	if c.gsVerts > 0 {
		m.GSVertDump = c.gsVerts
	}
	if c.gsReg != "" {
		parts := strings.Split(c.gsReg, ":")
		r, err := hx(parts[0])
		if err != nil {
			return fmt.Errorf("bad -gsreg %q (want REG[:N][:VALUE], REG hex)", c.gsReg)
		}
		n := 40
		if len(parts) > 1 {
			n, _ = strconv.Atoi(parts[1])
		}
		m.GSRegLog, m.GSRegLogN = uint8(r), n
		if len(parts) > 2 {
			v, err := strconv.ParseUint(strings.TrimPrefix(parts[2], "0x"), 16, 64)
			if err != nil {
				return fmt.Errorf("bad -gsreg value %q", parts[2])
			}
			m.GSRegDumpPacket, m.GSRegDumpVal = true, v
		}
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
	profSamples := map[uint32]int{}
	profTick := 0
	if tracing || traceFrom != 0 || len(logAt) > 0 || c.eeProf > 0 {
		m.OnStep = func(mm *ps2.Machine, pc uint32) {
			if c.eeProf > 0 {
				if profTick++; profTick >= c.eeProf {
					profTick = 0
					profSamples[pc]++
				}
			}
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
	if ieLogFlush != nil {
		ieLogFlush()
	}

	fmt.Println()
	fmt.Println(res)
	fmt.Printf("reached: %s\n", m.Sym(res.PC))
	fmt.Printf("vblanks: %d\n", m.VBlanks())
	// A halting breakpoint is a question about state; answer it without a second run.
	if len(c.bps) > 0 {
		fmt.Println()
		fmt.Print(m.Registers())
	}

	if c.eeProf > 0 {
		// Aggregate the PC samples by the function that owns them. The name is
		// resolved here, not at sample time, so sampling stays cheap.
		byFunc := map[string]int{}
		total := 0
		for pc, n := range profSamples {
			name := m.Sym(pc)
			if i := strings.IndexByte(name, '+'); i > 0 {
				name = name[:i]
			}
			byFunc[name] += n
			total += n
		}
		type fn struct {
			name string
			n    int
		}
		var fns []fn
		for name, n := range byFunc {
			fns = append(fns, fn{name, n})
		}
		sort.Slice(fns, func(i, j int) bool { return fns[i].n > fns[j].n })
		fmt.Printf("\nEE profile: %d samples, every %d steps\n", total, c.eeProf)
		for i, f := range fns {
			if i >= 40 || f.n*1000 < total {
				break
			}
			fmt.Printf("  %6.2f%%  %s\n", 100*float64(f.n)/float64(total), f.name)
		}
	}
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
	fmt.Println()
	fmt.Print(m.GSStatus())
	if v := m.VIFCensus(); v != "" {
		fmt.Print(v)
	}

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
		// The second processor's work list. It was printed only by the -iop harness, which is
		// exactly the run in which the interesting half of it cannot appear: the calls and the
		// registers that only happen once the EE has asked the IOP for something.
		if census := m.IOP.IOPCensus(); census != "" {
			fmt.Printf("\n--- %s", census)
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

	if c.goalSyms != "" {
		if err := writeGoalSyms(m, c.goalSyms); err != nil {
			return err
		}
	}
	if c.gsFrame != "" {
		if err := writeGSFrame(m, c.gsFrame); err != nil {
			return err
		}
	}
	if c.vu0Data != "" {
		data := m.VUDataMem(0)
		if data == nil {
			fmt.Println("vu0data: no VU0 yet")
		} else if err := os.WriteFile(c.vu0Data, data, 0o644); err != nil {
			return err
		} else {
			fmt.Printf("vu0data: wrote %d bytes to %s\n", len(data), c.vu0Data)
		}
	}
	if c.vu0Micro != "" {
		micro := m.VUMicro(0)
		if micro == nil {
			fmt.Println("vu0micro: no VU0 yet")
		} else if err := os.WriteFile(c.vu0Micro, micro, 0o644); err != nil {
			return err
		} else {
			fmt.Printf("vu0micro: wrote %d bytes to %s\n", len(micro), c.vu0Micro)
		}
	}
	if c.vu0Regs {
		if regs := m.VURegs(0); regs != "" {
			fmt.Print(regs)
		} else {
			fmt.Println("vu0regs: no VU0 yet")
		}
	}
	if c.vu1Micro != "" {
		micro := m.VUMicro(1)
		if micro == nil {
			fmt.Println("vu1micro: VIF1 never started; no program memory to dump")
		} else if err := os.WriteFile(c.vu1Micro, micro, 0o644); err != nil {
			return err
		} else {
			fmt.Printf("vu1micro: wrote %d bytes to %s\n", len(micro), c.vu1Micro)
		}
	}
	if c.vu1Data != "" {
		data := m.VUDataMem(1)
		if data == nil {
			fmt.Println("vu1data: VIF1 never started; no data memory to dump")
		} else if err := os.WriteFile(c.vu1Data, data, 0o644); err != nil {
			return err
		} else {
			fmt.Printf("vu1data: wrote %d bytes to %s\n", len(data), c.vu1Data)
		}
	}
	for _, spec := range c.gsFBs {
		if err := writeGSBuffer(m, spec); err != nil {
			return err
		}
	}
	for _, spec := range c.gsTexs {
		if err := writeGSTexture(m, spec); err != nil {
			return err
		}
	}
	return nil
}

// writeGSTexture dumps a texture as the sampler resolves it, named as TEX0:FILE.png.
func writeGSTexture(m *ps2.Machine, spec string) error {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("bad -gstex %q (want TEX0:FILE.png)", spec)
	}
	tex0, err := strconv.ParseUint(strings.TrimPrefix(parts[0], "0x"), 16, 64)
	if err != nil {
		return fmt.Errorf("bad -gstex %q", spec)
	}
	pix, w, h := m.GSTexture(tex0)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(parts[1])
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	fmt.Printf("gstex: wrote %dx%d to %s\n", w, h, parts[1])
	return nil
}

// goalObjName renders a GOAL name field: a gstring's chars live at +4; a symbol's
// name is found through the parallel table find_symbol_in_area reads (+0xFF38).
func goalObjName(m *ps2.Machine, ref uint32) string {
	if ref == 0 || ref >= 0x02000000 {
		return sprintfHex(ref)
	}
	if s := m.CString(ref + 4); isPrintable(s) && s != "" {
		return s
	}
	if strp := m.Read32(ref + 0xFF38); strp > 0x80000 && strp < 0x02000000 {
		if s := m.CString(strp + 4); isPrintable(s) && s != "" {
			return s
		}
	}
	return sprintfHex(ref)
}

func sprintfHex(v uint32) string { return fmt.Sprintf("0x%08X", v) }

// dumpGoalTree walks a GOAL process tree the way search-process-tree does:
// child handle at [node+16], each entry's process at [handle+0], the brother
// handle at [process+12]. Mask bit 0x100 marks a container rather than a process.
func dumpGoalTree(m *ps2.Machine, s7, node uint32, depth int) {
	if depth > 12 || node == 0 || node == s7 || node >= 0x02000000 {
		return
	}
	typ := m.Read32(node - 4)
	typeName := "?"
	if typ > 0x80000 && typ < 0x02000000 {
		typeName = goalObjName(m, m.Read32(typ))
	}
	mask := m.Read32(node + 4)
	fmt.Printf("%*s%08X  %-24s mask 0x%08X  %s\n",
		depth*2, "", node, goalObjName(m, m.Read32(node)), mask, typeName)
	h := m.Read32(node + 16)
	for n := 0; h != 0 && h != s7 && h < 0x02000000 && n < 512; n++ {
		p := m.Read32(h)
		if p == 0 || p == s7 || p >= 0x02000000 {
			break
		}
		dumpGoalTree(m, s7, p, depth+1)
		h = m.Read32(p + 12)
	}
}

// writeGoalSyms dumps the GOAL symbol table: every runtime-linked name the engine
// knows itself by, with its current value. The layout is the game's own, read off
// find_symbol_from_c / find_symbol_in_area (boot ELF, symbols shipped):
//
//	offset = (hash << 19) >> 16          — 8-byte cells spanning $s7 ± 0x8000
//	[cell + 0xFF34] = the name's hash    — the parallel info table
//	[cell + 0xFF38] = gstring*, chars at +4
//	[cell]          = the symbol's value
//
// $s7 is the GOAL symbol-base register (the C kernel itself indexes off it), so the
// base is read from the CPU, not guessed.
func writeGoalSyms(m *ps2.Machine, path string) error {
	s7 := uint32(m.CPU.Reg(23))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintf(w, "s7 (symbol base) = 0x%08X\n", s7)
	n := 0
	for off := int32(-0x8000); off < 0x8000; off += 8 {
		cell := s7 + uint32(off)
		strp := m.Read32(cell + 0xFF38)
		if strp < 0x80000 || strp >= 0x02000000 {
			continue
		}
		name := m.CString(strp + 4)
		if name == "" || len(name) > 96 {
			continue
		}
		ok := true
		for _, ch := range name {
			if ch < 0x21 || ch > 0x7E {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%08X %+6d %08X %s\n", cell, off, m.Read32(cell), name)
		n++
	}
	fmt.Printf("goalsyms: %d symbols to %s\n", n, path)
	return nil
}

// writeGSBuffer dumps an arbitrary PSMCT32 buffer named as BASE:FBW:H:FILE.png.
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
	nonBlack := 0
	for i := 0; i < len(pix); i += 4 {
		if pix[i] != 0 || pix[i+1] != 0 || pix[i+2] != 0 {
			nonBlack++
		}
	}
	fmt.Printf("gsfb: wrote %dx%d to %s (%d non-black)\n", w, h, parts[3], nonBlack)
	return nil
}

// writeGSFrame dumps the frame the GS would be scanning out as a PNG — the eyes of the
// render bring-up, and it reads what is actually in GS memory through the real deswizzle,
// not a re-render of anything.
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
	// Whether the frame is black because black was drawn or because nothing was is
	// the difference between a colour bug and a missing draw; say which.
	nonBlack := 0
	for i := 0; i < len(pix); i += 4 {
		if pix[i] != 0 || pix[i+1] != 0 || pix[i+2] != 0 {
			nonBlack++
		}
	}
	fmt.Printf("gsframe: wrote %dx%d to %s (%d of %d pixels non-black)\n", w, h, path, nonBlack, w*h)
	return nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
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

	// The interrupt-enable log: every suspend/resume/deliver/save/load, one line each,
	// with the PC and the caller. It is the instrument for the one bug family nothing
	// else can see — an enable bit that goes false and stays false across a thread
	// switch, faithfully saved and faithfully restored by everything downstream.
	if c.iopIELog != "" {
		f, e := os.Create(c.iopIELog)
		if e != nil {
			return e
		}
		w := bufio.NewWriterSize(f, 1<<20)
		ieLogFlush = func() { w.Flush(); f.Close() }
		prev := m.OnIOPStart
		m.OnIOPStart = func(p *ps2.IOP) {
			if prev != nil {
				prev(p)
			}
			p.OnIntrState = func(ev ps2.IOPIntrEvent) {
				fmt.Fprintf(w, "%d %s pc=%s ra=%s addr=%08X val=%08X ie=%d d=%d i=%d\n",
					ev.Step, ev.Kind, p.Sym(ev.PC), p.Sym(ev.RA), ev.Addr, ev.Val,
					b2i(ev.Enabled), ev.Depth, ev.InIntr)
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
	if ieLogFlush != nil {
		ieLogFlush()
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

	if c.iopThreads {
		fmt.Printf("\n--- %s", m.IOP.IOPThreads())
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
