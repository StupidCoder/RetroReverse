// bootoracle boots an original-Xbox title's XBE on the Xbox machine
// (tools/platform/xbox) and runs its XDK/CRT boot code, exposing the standard oracle
// instrumentation (STANDARDS §3). Where xbeinfo (Phase 0) reads the disc's containers
// and the executable's header statically, this one RUNS the code they hold: it loads
// default.xbe at its fixed base, high-level-emulates the xboxkrnl exports the boot path
// reaches, and stops at the first NV2A push-buffer kick — or, before that, at the first
// kernel ordinal not yet modelled, naming it, so each run reports concretely how far
// the bring-up reaches.
//
// Usage:
//
//	bootoracle -image DISC.iso [-steps N] [-trace] [-tracen N] [-bp ADDR]...
//	           [-watch ADDR[:LEN]] [-rwatch ADDR[:LEN]] [-watchn N] [-v] [-ordinals]
//	           [-dump ADDR[:LEN]]... [-dis ADDR[:N]]... [-poke ADDR:VALUE] [-stack]
//	           [-keys SPEC] [-savestate FILE] [-loadstate FILE]
//
// Addresses are flat hex (this machine runs in flat protected mode, so there is no
// SEG:OFF to write — an earlier version of this comment promised one, alongside -bp and
// -watch flags that were never registered at all).
package main

import (
	"flag"
	"fmt"
	"os"
	"crypto/md5"
	"runtime/pprof"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func parseNum(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

// parseHex reads an address argument. Addresses are hex whether or not they carry the
// 0x — STANDARDS §3 ("address arguments are the platform's natural form"), and nobody
// writing 5E5158 means five million.
func parseHex(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimSpace(s), "0x"), 16, 64)
	return uint32(v), err
}

// parseAddrN reads "ADDR[:N]", where N is a decimal count with a per-flag default.
func parseAddrN(spec string, def int) (addr uint32, n int, err error) {
	parts := strings.SplitN(spec, ":", 2)
	if addr, err = parseHex(parts[0]); err != nil {
		return 0, 0, fmt.Errorf("bad address %q: %w", parts[0], err)
	}
	n = def
	if len(parts) == 2 {
		if n, err = strconv.Atoi(strings.TrimSpace(parts[1])); err != nil {
			return 0, 0, fmt.Errorf("bad count %q: %w", parts[1], err)
		}
	}
	return addr, n, nil
}

// parseWatch reads "ADDR[:LEN]" into the half-open window [lo,hi) the machine's watches
// take. LEN is hex like the address, and defaults to a dword.
func parseWatch(s string) (lo, hi uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if lo, err = parseHex(parts[0]); err != nil {
		return 0, 0, fmt.Errorf("bad watch address %q: %w", parts[0], err)
	}
	length := uint32(4)
	if len(parts) == 2 {
		n, e := parseHex(parts[1])
		if e != nil {
			return 0, 0, fmt.Errorf("bad watch length %q: %w", parts[1], e)
		}
		length = n
	}
	return lo, lo + length, nil
}

// keyPress is one scheduled pad input: hold the control named `control` from frame
// `atFrame` onward, releasing after `holdFrames` frames (0 = held for the rest of the run).
//
// It carries a NAME rather than a bit mask, which it did not have to before the pad could
// say more than eight bits: a stick direction is not a bit, and two of them held at once
// resolve to one axis pair rather than to an OR. So the schedule collects the names that
// are held on a given frame and xbox.PadStateOf turns the set into a level — the same call
// the debugger's keyboard makes, out of the same table.
//
// The unit is the FRAME — the title's flip — because that is the boundary the title's own
// poll is synchronous with, and because a run that starts from a savestate wants to count
// from where it starts rather than from a boot it did not do. It edge-detects presses from
// consecutive polls (its own ~prev&cur), so a release between two presses of one button is
// what makes them two presses rather than one long hold.
type keyPress struct {
	atFrame    uint64
	holdFrames uint64
	control    string
}

// parseKeys turns a "-keys a@120,stickleft@900:10" spec into a schedule. NAME@FRAME holds
// from that frame on; NAME@FRAME:N holds for N frames and then releases.
func parseKeys(spec string) ([]keyPress, error) {
	var sched []keyPress
	for _, tok := range strings.Split(spec, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		name, frameS, ok := strings.Cut(tok, "@")
		if !ok {
			return nil, fmt.Errorf("bad -keys token %q, want NAME@FRAME[:HOLD]", tok)
		}
		name = strings.ToLower(strings.TrimSpace(name))
		if _, ok := xbox.PadControlByName(name); !ok {
			return nil, fmt.Errorf("unknown pad control %q in -keys (have %s)",
				name, strings.Join(xbox.PadControlNames(), " "))
		}
		frameS = strings.TrimSpace(frameS)
		var hold uint64
		if f2, holdS, has := strings.Cut(frameS, ":"); has {
			frameS = f2
			h, err := parseNum(strings.TrimSpace(holdS))
			if err != nil || h == 0 {
				return nil, fmt.Errorf("bad hold in -keys token %q: want a positive frame count", tok)
			}
			hold = h
		}
		f, err := parseNum(frameS)
		if err != nil {
			return nil, fmt.Errorf("bad frame in -keys token %q: %w", tok, err)
		}
		sched = append(sched, keyPress{atFrame: f, holdFrames: hold, control: name})
	}
	return sched, nil
}

// parsePoke reads "ADDR:VALUE", both hex.
func parsePoke(s string) (addr, val uint32, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad -poke %q, want ADDR:VALUE", s)
	}
	if addr, err = parseHex(parts[0]); err != nil {
		return 0, 0, err
	}
	val, err = parseHex(parts[1])
	return addr, val, err
}

func main() {
	image := flag.String("image", "", "Xbox disc image (.iso / XISO)")
	xbePath := flag.String("xbe", "/default.xbe", "path of the XBE within the disc to boot")
	steps := flag.String("steps", "50000000", "instruction budget (hex with 0x, else decimal)")
	verbose := flag.Bool("v", false, "log kernel calls and events live")
	ordinals := flag.Bool("ordinals", false, "after the run, print the histogram of xboxkrnl ordinals called")
	irqs := flag.Bool("irqs", false, "after the run, print the interrupt vectors the title has connected")
	threads := flag.Bool("threads", false, "after the run, print every thread's scheduler state, wait objects and signal provenance")
	stackDump := flag.Bool("stack", false, "on halt, dump the top of the stack (the caller's argument frame)")
	trace := flag.Bool("trace", false, "trace executed instructions")
	tracen := flag.Int("tracen", 200, "limit -trace to this many instructions")
	savestate := flag.String("savestate", "", "after the run, write a machine snapshot to this file")
	pngOut := flag.String("png", "", "after the run, write the display scanout to this PNG")
	surfOut := flag.String("surfpng", "", "after the run, write the Kelvin render surface (AA-resolved) to this PNG")
	loadstate := flag.String("loadstate", "", "restore a machine snapshot before running")
	gpu := flag.Bool("gpu", false, "Phase C: run the NV2A DMA pusher on each kick (do not stop at first push)")
	survey := flag.Bool("survey", false, "with -gpu: record the PGRAPH method surface and print it")
	watch := flag.String("watch", "", "break-free write watch on ADDR[:LEN] (hex): log each write with its PC")
	rwatch := flag.String("rwatch", "", "read watch on ADDR[:LEN] (hex): log each read with its PC")
	watchn := flag.Int("watchn", 40, "limit -watch/-rwatch to this many reported accesses")
	poke := flag.String("poke", "", "write ADDR:VALUE (hex) after loading, before running — a probe, not a model")
	keys := flag.String("keys", "", "pad-1 input script: NAME@FRAME[:HOLD][,...] — holds pad control NAME from that frame (the title's flip) for HOLD frames (default: forever). Names: see xbox.PadControlNames. e.g. -keys start@120,a@300:10,stickleft@400:8")
	stopflip := flag.Int("stopflip", 0, "stop the run at the Nth FLIP_STALL — the hook fires while the completed frame is still the bound colour surface, so -surfpng captures a whole presented frame instead of a mid-frame slice")
	ramhash := flag.Bool("ramhash", false, "after the run, print an md5 of guest RAM + the CPU position (divergence comparator)")
	cpuprofile := flag.String("cpuprofile", "", "write a host pprof CPU profile of the run to this file")
	var bps multiFlag
	flag.Var(&bps, "bp", "execution breakpoint at ADDR (hex); repeatable")
	var dumps multiFlag
	flag.Var(&dumps, "dump", "hex-dump ADDR:LEN of memory after the run (hex); repeatable")
	var diss multiFlag
	flag.Var(&diss, "dis", "disassemble N instructions at ADDR[:N] after the run (default 32); repeatable")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "bootoracle: -image is required")
		os.Exit(2)
	}

	budget, err := parseNum(*steps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: bad -steps %q: %v\n", *steps, err)
		os.Exit(2)
	}

	disc, err := xbox.Open(*image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: open image: %v\n", err)
		os.Exit(1)
	}
	defer disc.Close()

	xbeBytes, err := disc.ReadFile(*xbePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: read %s: %v\n", *xbePath, err)
		os.Exit(1)
	}
	xbe, err := xbox.ParseXBE(xbeBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: parse XBE: %v\n", err)
		os.Exit(1)
	}

	m, err := xbox.NewMachine(xbe, disc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: build machine: %v\n", err)
		os.Exit(1)
	}
	m.SetVerbose(*verbose)
	if os.Getenv("RR_HOTPC") != "" {
		m.EnableHotPC()
	}
	if *gpu {
		m.EnableGPU()
		if *survey {
			m.PGraph().SetSurvey(true)
		}
	}

	if *loadstate != "" {
		if err := m.LoadStateFile(*loadstate); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: loadstate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("restored machine state from %s\n", *loadstate)
		if m.CPU.Halted {
			// A frontier state: it stopped on an unimplemented ordinal, whose trap is
			// retried on resume (the sentinel EIP was saved untouched).
			fmt.Printf("state was halted (%s) — clearing to retry\n", m.CPU.HaltReason)
			m.ClearHalt()
		}
	}

	// -poke lands after the restore and before the run: it is a probe against the state
	// the run starts from, and a savestate loaded over it would undo it.
	if *poke != "" {
		addr, val, err := parsePoke(*poke)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: %v\n", err)
			os.Exit(2)
		}
		m.Poke(addr, val)
		fmt.Printf("poked %08X = %08X\n", addr, val)
	}

	// Instruments.
	for _, s := range bps {
		a, err := parseHex(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: bad -bp %q: %v\n", s, err)
			os.Exit(2)
		}
		m.SetBreakpoint(a)
	}
	if *watch != "" {
		lo, hi, err := parseWatch(*watch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: %v\n", err)
			os.Exit(2)
		}
		seen := 0
		m.SetWriteWatch(lo, hi, func(addr, val, pc uint32) {
			if seen < *watchn {
				fmt.Printf("  write %08X = %08X (pc %08X)\n", addr, val, pc)
				seen++
			}
		})
	}
	if *rwatch != "" {
		lo, hi, err := parseWatch(*rwatch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: %v\n", err)
			os.Exit(2)
		}
		seen := 0
		m.SetReadWatch(lo, hi, func(addr, val, pc uint32) {
			if seen < *watchn {
				fmt.Printf("  read %08X = %08X (pc %08X)\n", addr, val, pc)
				seen++
			}
		})
	}

	// The pad. A -keys schedule is what plugs a controller in: a run that was not asked
	// to drive one leaves the root hub empty, which is both the truth and what every
	// boot before this flag did.
	//
	// The schedule is recomputed from the frame counter on each flip rather than stepped
	// through a cursor, which is what makes it savestate-stable for free: there is no
	// oracle-side position to save, and a restored machine resumes the same schedule at
	// whatever frame it restored into.
	if *keys != "" {
		sched, err := parseKeys(*keys)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: %v\n", err)
			os.Exit(2)
		}
		m.AttachPad(0)
		frame := uint64(0)
		prev := m.OnFlip
		m.OnFlip = func(mm *xbox.Machine) {
			if prev != nil {
				prev(mm) // OnFlip is the machine's one frame hook; the pad does not own it
			}
			frame++
			held := map[string]bool{}
			for _, k := range sched {
				if frame >= k.atFrame && (k.holdFrames == 0 || frame < k.atFrame+k.holdFrames) {
					held[k.control] = true
				}
			}
			mm.SetPad(0, xbox.PadStateOf(held))
		}
	}

	if *trace {
		m.SetTrace(*tracen) // print the first -tracen executed instructions (PC trail)
	}

	if *stopflip > 0 {
		n := *stopflip
		prev := m.OnFlip
		m.OnFlip = func(mm *xbox.Machine) {
			if prev != nil {
				prev(mm)
			}
			n--
			if n == 0 {
				mm.StopRequested = true
			}
		}
	}

	if *cpuprofile != "" {
		pf, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: -cpuprofile: %v\n", err)
			os.Exit(2)
		}
		pprof.StartCPUProfile(pf)
		defer func() {
			pprof.StopCPUProfile()
			pf.Close()
		}()
	}

	reason, n := m.Run(budget)
	fmt.Printf("\n=== run ended: %s after %d instructions ===\n", reason, n)
	fmt.Print(m.Report())

	if *stackDump && m.CPU.Halted {
		sp := m.CPU.Regs[4]
		fmt.Printf("stack @ ESP=%08X:\n", sp)
		for i := uint32(0); i < 12; i++ {
			a := sp + i*4
			fmt.Printf("  [ESP+%02X] %08X = %08X\n", i*4, a, m.MemRead32(a))
		}
		// The call site: disassemble a window ending at the return address so the
		// argument pushes (which pin the ordinal's signature) are visible.
		ret := m.CallerReturnAddr()
		if ret > 48 {
			fmt.Printf("call site (around return %08X):\n", ret)
			for _, line := range m.DisasmForward(ret-40, 20) {
				fmt.Println(line)
			}
		}
	}

	for _, d := range dumps {
		parts := strings.SplitN(d, ":", 2)
		addr, _ := parseNum(parts[0])
		length := uint64(64)
		if len(parts) == 2 {
			length, _ = parseNum(parts[1])
		}
		hexDump(m, uint32(addr), uint32(length))
	}

	for _, d := range diss {
		addr, n, err := parseAddrN(d, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: bad -dis %q: %v\n", d, err)
			continue
		}
		fmt.Printf("disasm %08X (%d instructions):\n", addr, n)
		for _, line := range m.DisasmForward(addr, n) {
			fmt.Println(line)
		}
	}

	if *threads {
		fmt.Println("\nthreads:")
		for _, line := range m.DebugThreads() {
			fmt.Println(line)
		}
	}

	if os.Getenv("RR_HOTPC") != "" {
		fmt.Println("\nhot PCs (sampled every 256 instructions):")
		for _, line := range m.HotPCReport(30) {
			fmt.Println(line)
		}
	}

	if *ordinals {
		fmt.Println("\nxboxkrnl ordinals reached:")
		for _, line := range m.OrdinalHistogram() {
			fmt.Println(line)
		}
	}

	// The connected interrupt vectors. HalGetInterruptVector is modelled as identity, so
	// a "vector" here is the bus level the XDK passed — which is why these are read off a
	// live machine rather than written down: a device's level is what the title says it
	// is. The registrations ride in the savestate, so a restored state answers too.
	if *irqs {
		fmt.Println("\nconnected interrupt vectors:")
		found := false
		for v := uint32(0); v < 64; v++ {
			ki := m.DebugInterruptKI(v)
			if ki == 0 {
				continue
			}
			found = true
			fmt.Printf("  vector %2d -> KINTERRUPT %08X (routine %08X ctx %08X)\n",
				v, ki, m.MemRead32(ki), m.MemRead32(ki+4))
		}
		if !found {
			fmt.Println("  (none)")
		}
	}

	if *gpu && *survey {
		fmt.Println()
		for _, line := range m.PGraph().SurveyReport() {
			fmt.Println(line)
		}
	}

	if *ramhash {
		fmt.Printf("ramhash: %x cpu: PC=%08X steps=%d\n", md5.Sum(m.RAM), m.CPU.LinearPC(), m.CPU.Steps)
	}
	if *pngOut != "" {
		if data, err := m.FramePNG(); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: png: %v\n", err)
		} else if err := os.WriteFile(*pngOut, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: png: %v\n", err)
		} else {
			fmt.Printf("wrote frame to %s\n", *pngOut)
		}
	}
	if *surfOut != "" {
		if data, err := m.SurfacePNG(); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: surfpng: %v\n", err)
		} else if err := os.WriteFile(*surfOut, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: surfpng: %v\n", err)
		} else {
			fmt.Printf("wrote render surface to %s\n", *surfOut)
		}
	}
	if *savestate != "" {
		if err := m.SaveStateFile(*savestate); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: savestate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("wrote machine snapshot to %s\n", *savestate)
	}
}

func hexDump(m *xbox.Machine, addr, length uint32) {
	fmt.Printf("dump %08X..%08X:\n", addr, addr+length)
	for off := uint32(0); off < length; off += 16 {
		fmt.Printf("  %08X ", addr+off)
		var ascii strings.Builder
		for i := uint32(0); i < 16 && off+i < length; i++ {
			b := m.MemReadByte(addr + off + i)
			fmt.Printf("%02X ", b)
			if b >= 0x20 && b < 0x7F {
				ascii.WriteByte(b)
			} else {
				ascii.WriteByte('.')
			}
		}
		fmt.Printf(" %s\n", ascii.String())
	}
}
