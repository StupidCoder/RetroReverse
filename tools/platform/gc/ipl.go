package gc

// ipl.go stands in for the console's boot ROM: it lays down the low-memory globals the
// game expects, then runs the disc's own apploader to load the executable and the
// filesystem, exactly as the real IPL does.
//
// The apploader protocol is small and worth stating in full, because running it — rather
// than emulating what it would have done — is the whole reason this machine needs no BIOS.
// The IPL copies the loader off the disc, jumps to its entry point, and the entry point
// writes three function pointers back through pointers handed to it in r3, r4 and r5:
//
//	init(report_fn)              one-time setup; report_fn is an OSReport-style logger
//	main(&dst, &size, &offset)   called repeatedly; returns 1 while it wants more, and
//	                             each time asks the IPL to read `size` bytes from disc
//	                             `offset` into `dst`
//	close()                      returns the game's entry point
//
// So the apploader does not touch the disc hardware at all: it asks, and the IPL reads.
// That is what makes this the first real milestone — running it exercises the CPU, the
// block-address translation, the cache instructions and the IPL handoff, and proves them
// all at once, on the game's own code, before any device beyond memory is needed. The
// gate is that close() returns 0x80003100, the entry point the DOL header names.
//
// The three low-memory globals that cannot be derived from the disc — the memory size, the
// console type, and the two clock speeds — are the smallest of the three console-resident
// substitutions named in the package doc. The rest are the disc's own header, laid where
// the game's __start reads it.

import (
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/gekko"
)

// Where the IPL puts the apploader, and the two host trampolines the apploader calls back
// through. The trampolines are addresses the run loop recognises rather than real code: a
// call to one is serviced in Go and returns via the link register.
const (
	apploaderAddr  = 0x81200000 // the loader's body, as the real IPL stages it
	reportTramp    = 0x81330000 // init()'s OSReport callback lands here
	returnSentinel = 0x81340000 // a called apploader function returns to here
)

// bootMagic marks a booted console: __start checks for it. Its ASCII is "disk ea5e".
const bootMagic = 0x0D15EA5E

// setupLowMem lays down the globals the console's boot ROM leaves in the bottom of memory.
//
// Most are the disc's own header, copied where the game looks for it. Four are not on the
// disc and are the substitution: the physical memory size, the console type, and the bus
// and core clock speeds — the two clocks being exactly what the tracer caught __start
// reading at 0x800000F8 and 0x800000FC.
func (m *Machine) setupLowMem() {
	// The disc ID: the first 0x20 bytes of boot.bin, which name the game to its own code.
	boot, _ := m.disc.Read(0, 0x20)
	m.dmaToRAM(0x80000000-0x80000000, boot) // physical 0..0x20

	set := func(addr, v uint32) { m.setRAM32(addr, v) }

	set(0x00000020, bootMagic)  // "this console has booted a disc"
	set(0x00000024, 1)          // the boot-ROM version
	set(0x00000028, RAMSize)    // physical memory size — a substitution
	set(0x0000002C, 3)          // production console (retail hardware) — a substitution
	set(0x00000030, 0x00000000) // ARENA lo: the bottom of the free heap
	set(0x00000034, 0x817FE8C0) // ARENA hi: the top, below the stack
	set(0x00000038, m.disc.Header.FSTAddr)
	set(0x0000003C, m.disc.Header.FSTMaxSize)
	set(0x000000CC, 0)         // video mode: 0 = NTSC — a substitution, from the SRAM
	set(0x000000F0, RAMSize)   // "simulated" memory size, equal to the real one here
	set(0x000000F8, BusClock)  // the bus clock — what __start reads at 0x800000F8
	set(0x000000FC, CoreClock) // the core clock — and at 0x800000FC

	// 0x800000F4 holds a pointer to the boot-info-2 block, or null. __start dereferences
	// it only when non-null; left null, it takes the ordinary path. If a title turns out
	// to need it, the poison-and-watch derivation (see the oracle's -lowmem) is how its
	// exact contents are found rather than guessed.
	set(0x000000F4, 0)
}

// BusClock and CoreClock are declared in tools/cpu/gekko (the timer paces itself by them),
// but the machine writes them into low memory for the game to read, so it names them here
// with the same values.
const (
	BusClock  = 162_000_000
	CoreClock = 486_000_000
)

// setupState puts the processor into the post-IPL state: translation on through the
// standard block map, the machine state register as a booted console leaves it, and the
// stack somewhere valid. Everything after this runs as the game's own code.
func (m *Machine) setupState() {
	c := m.CPU

	// The block-address translations every GameCube program runs on. Two map main memory
	// twice — cached at 0x80000000, uncached at 0xC0000000 — and one maps the hardware
	// registers. This IS the memory map; without it the game's first load faults.
	//
	// A BAT's upper word is (effective base) | (block-length field << 2) | valid bits; the
	// lower is (physical base) | protection. A 256 MiB block has length field 0x7FF.
	const bl256 = 0x7FF << 2
	c.DBAT[0] = [2]uint32{0x80000000 | bl256 | 2, 0x00000000 | 2}            // cached RAM
	c.DBAT[1] = [2]uint32{0xC0000000 | bl256 | 2, 0x00000000 | (2 << 3) | 2} // uncached RAM (guarded)
	c.IBAT[0] = [2]uint32{0x80000000 | bl256 | 2, 0x00000000 | 2}
	// The hardware registers: a 256 MiB uncached block at 0xC0000000 already covers
	// 0xCC000000, but a game maps them explicitly too; DBAT1 above reaches them.

	c.MSR = gekko.MSRFP | gekko.MSRDR | gekko.MSRIR | gekko.MSRME
	c.GPR[1] = 0x816FFF00 // a stack pointer somewhere sensible in the top of memory
	c.GPR[2] = 0x81600000 // a plausible small-data area pointer
	c.GPR[13] = 0x81500000

	// The syscall handoff. The SDK's cache-range routines (DCFlushRange and its kin) end
	// in `sc` — not to call anything, but as a heavyweight synchronisation barrier whose
	// handler does nothing but return. At apploader time the game has not installed its
	// exception vectors yet, so there is no handler at 0xC00, and taking the exception
	// would run into the zeroed vector page. Until a handler is installed, `sc` is
	// therefore serviced as the no-op sync it is; once the game installs its own, the
	// architectural exception is taken and its handler runs. That switch is the whole of
	// the substitution — it is the IPL's default syscall behaviour, and it lasts exactly
	// until the game replaces it.
	c.SC = m.handleSyscall
}

// handleSyscall services an `sc`. It returns true (handled, continue past it) while the
// syscall vector is empty — the IPL's do-nothing default — and false once the game has
// installed a handler, so the real exception is taken.
func (m *Machine) handleSyscall(c *gekko.CPU) bool {
	vec := uint32(0x00000C00)
	if c.MSR&gekko.MSRIP != 0 {
		vec = 0xFFF00C00
	}
	if m.ram32(vec) == 0 {
		return true // no handler yet: `sc` is the SDK's sync barrier, and it returns
	}
	return false // the game's handler is installed; take the exception
}

// RunApploader is the IPL handoff: set up memory, run the disc's loader, and leave the
// machine at the game's entry point, ready to run. It returns the entry point the loader
// reported, which the caller checks against the DOL header.
func (m *Machine) RunApploader() (entry uint32, err error) {
	m.setupLowMem()
	m.setupState()

	// Stage the apploader off the disc and jump to its entry. Its entry point takes three
	// pointers — to where it should write its init, main and close addresses.
	code, err := m.disc.ApploaderCode()
	if err != nil {
		return 0, err
	}
	m.dmaToRAM(apploaderAddr-0x80000000, code)

	// Three scratch words for the loader to write its function pointers into.
	const (
		pInit  = 0x81300000
		pMain  = 0x81300004
		pClose = 0x81300008
	)
	c := m.CPU
	c.GPR[3], c.GPR[4], c.GPR[5] = pInit, pMain, pClose
	if err := m.call(m.disc.Apploader.Entry); err != nil {
		return 0, fmt.Errorf("apploader entry: %w", err)
	}

	initFn := m.ram32(pInit)
	mainFn := m.ram32(pMain)
	closeFn := m.ram32(pClose)
	if initFn == 0 || mainFn == 0 || closeFn == 0 {
		return 0, fmt.Errorf("the apploader did not report its functions (init=0x%08X main=0x%08X close=0x%08X)", initFn, mainFn, closeFn)
	}

	// init(report_fn): hand it the report trampoline, so its OSReport calls are printed.
	c.GPR[3] = reportTramp
	if err := m.call(initFn); err != nil {
		return 0, fmt.Errorf("apploader init: %w", err)
	}

	// main(&dst, &size, &offset), in a loop, servicing each read the loader asks for.
	const (
		pDst  = 0x8130000C
		pSize = 0x81300010
		pOff  = 0x81300014
	)
	for i := 0; i < 1000; i++ {
		c.GPR[3], c.GPR[4], c.GPR[5] = pDst, pSize, pOff
		if err := m.call(mainFn); err != nil {
			return 0, fmt.Errorf("apploader main (iteration %d): %w", i, err)
		}
		if c.GPR[3] == 0 {
			break // the loader is done: it returned 0
		}
		dst := m.ram32(pDst)
		size := m.ram32(pSize)
		off := m.ram32(pOff)
		// The offset the loader gives is a byte offset on the disc; it asks for the DOL's
		// segments and then the FST, and the -dvd log names each file the read lands in.
		if m.OnDVDRead != nil {
			m.OnDVDRead(int64(off), size, dst)
		}
		data, rerr := m.disc.Read(int64(off), int(size))
		if rerr != nil {
			return 0, fmt.Errorf("apploader read (offset 0x%X size %d): %w", off, size, rerr)
		}
		m.dmaToRAM(dst&0x03FFFFFF, data)
	}

	// close(): it returns the game's entry point.
	if err := m.call(closeFn); err != nil {
		return 0, fmt.Errorf("apploader close: %w", err)
	}
	entry = c.GPR[3]
	c.PC = entry
	return entry, nil
}

// call runs a Gekko function to its return: it sets the link register to a sentinel the
// run loop recognises, jumps to the function, and steps until control comes back — handling
// the report-trampoline callback along the way. It is how the host invokes the game's own
// code and gets an answer.
func (m *Machine) call(fn uint32) error {
	c := m.CPU
	c.LR = returnSentinel
	c.PC = fn
	const budget = 200_000_000
	trace := iplTrace
	var ring [24]uint32
	ri := 0
	for i := 0; i < budget; i++ {
		switch c.PC {
		case returnSentinel:
			return nil // the function returned
		case reportTramp:
			m.serviceReport()
			continue
		}
		if trace {
			ring[ri%len(ring)] = c.PC
			ri++
			// The apploader runs entirely in the top of memory; a PC below it means an
			// exception was taken to a vector the loader has not installed yet. Print the
			// path that led there and stop.
			if c.PC < 0x80000000 {
				fmt.Fprintf(os.Stderr, "ipl: control left the apploader for 0x%08X. Last PCs:\n", c.PC)
				for k := 0; k < len(ring); k++ {
					fmt.Fprintf(os.Stderr, "    0x%08X\n", ring[(ri+k)%len(ring)])
				}
				return fmt.Errorf("apploader jumped to 0x%08X (SRR0=0x%08X SRR1=0x%08X)", c.PC, c.SRR0, c.SRR1)
			}
		}
		m.tickVI()
		c.Step()
		if c.Halted {
			return fmt.Errorf("halted at 0x%08X: %s", c.PC, c.HaltReason)
		}
	}
	return fmt.Errorf("did not return within %d instructions (PC 0x%08X)", budget, c.PC)
}

// serviceReport handles a call to the apploader's report callback: it prints the format
// string (the varargs are not expanded — the string alone is what makes the loader's
// progress legible) and returns via the link register.
func (m *Machine) serviceReport() {
	c := m.CPU
	msg := m.readCString(c.GPR[3])
	if msg != "" {
		m.logf("apploader: %s", trimReport(msg))
	}
	c.PC = c.LR // return
}

// readCString reads a NUL-terminated string out of the game's memory, through the CPU's
// own view of it (translation and locked cache included), so it reads what the game wrote.
func (m *Machine) readCString(addr uint32) string {
	var b []byte
	for i := uint32(0); i < 256; i++ {
		ch := m.CPU.ReadMem(addr + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

func trimReport(s string) string {
	// Report strings often end in a newline; drop it so the log stays one line per report.
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// iplTrace, set from the RR_GC_IPLTRACE environment variable, turns on the apploader
// path trace — the diagnostic that shows how control left the loader.
var iplTrace = os.Getenv("RR_GC_IPLTRACE") != ""
