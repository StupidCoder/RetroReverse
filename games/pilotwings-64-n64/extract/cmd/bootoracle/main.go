// bootoracle boots the Pilotwings 64 cartridge on the N64 machine model
// (tools/platform/n64) and runs it, reporting where it stopped and why.
//
// Nothing about the game is emulated on its behalf: libultra ships inside the
// cartridge and runs on the VR4300 as ordinary code. Only the console's hardware
// is modelled, plus the PIF boot handoff, which is not on the medium.
//
// Usage:
//
//	bootoracle -image ROM [-steps N] [-bp ADDR] [-watch ADDR[:LEN]] [-trace] [-tracen N]
//	           [-savestate FILE] [-loadstate FILE] [-o DIR]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n64"
)

func main() {
	image := flag.String("image", "", "cartridge image (.z64/.v64/.n64)")
	steps := flag.String("steps", "10000000", "instruction budget (hex with 0x, else decimal)")
	trace := flag.Bool("trace", false, "trace executed instructions to stderr")
	tracen := flag.Int("tracen", 200, "limit -trace to this many instructions")
	var bps, watches multiFlag
	flag.Var(&bps, "bp", "breakpoint address (hex); repeatable")
	flag.Var(&watches, "watch", "memory watch ADDR[:LEN] (hex); repeatable")
	shot := flag.String("shot", "", "write a PNG of the framebuffer to this directory")
	shotEvery := flag.Int("shotevery", 60, "write a -shot PNG every this many video fields (1 = every field)")
	shotBase := flag.Int("shotbase", 0, "number the first -shot field as this, for runs resumed from a savestate")
	dmaLog := flag.String("dmalog", "", "log every DMA transfer (kind, field, addresses, length) to this file")
	stopField := flag.Int("stopfield", 0, "stop the run after this many video fields (0 = never)")
	var callLogs multiFlag
	flag.Var(&callLogs, "calllog", "log a0-a3 and ra each time this PC executes (hex); repeatable")
	rwatch := flag.String("rwatch", "", "read-watch ADDR:LEN (hex): summarise, per reading PC, the hit count and address range")
	keys := flag.String("keys", "", "field-timed controller script, e.g. \"2160:+start,2166:-start,2200:+a\"")
	saveState := flag.String("savestate", "", "after the run, dump the full machine snapshot to this file")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before running")
	out := flag.String("o", "", "output directory")
	pcmDump := flag.String("pcmdump", "", "write the raw PCM the audio thread plays (AI buffers) to this file, plus a .rate sidecar")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "bootoracle: -image is required")
		flag.Usage()
		os.Exit(2)
	}
	if *shotEvery < 1 {
		fmt.Fprintln(os.Stderr, "bootoracle: -shotevery must be at least 1")
		os.Exit(2)
	}
	if err := run(*image, *steps, *trace, *tracen, bps, watches, callLogs, *rwatch, *keys,
		*shot, *shotEvery, *shotBase, *dmaLog, *stopField, *saveState, *loadState, *out, *pcmDump); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

// num parses a hex (0x-prefixed or bare) or decimal count.
func num(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		return strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func hx(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
	return uint32(v), err
}

// readSite aggregates one PC's reads inside the watch window. Raw per-access
// logging drowns at this scale; the {count, lo..hi} summary per PC *is* the
// structure — a directory parser only touches the head, a record walker the
// tail, and a cluster of PCs whose ranges start four bytes apart is one routine
// reading consecutive words.
type readSite struct {
	count  int
	lo, hi uint32
}

// buttonNames maps a script token to the joybus button bit. Setting these in
// m.Controllers[0] *is* the game's real input path: Pilotwings polls the joybus
// itself, so nothing is injected past the hardware the way Ultima Underworld's
// oracle has to inject past DOS.
var buttonNames = map[string]uint16{
	"a": n64.BtnA, "b": n64.BtnB, "z": n64.BtnZ, "start": n64.BtnStart,
	"dup": n64.BtnDUp, "ddown": n64.BtnDDown, "dleft": n64.BtnDLeft, "dright": n64.BtnDRight,
	"l": n64.BtnL, "r": n64.BtnR,
	"cup": n64.BtnCUp, "cdown": n64.BtnCDown, "cleft": n64.BtnCLeft, "cright": n64.BtnCRight,
}

// keyEvent is one scripted controller change, applied when the field is reached.
type keyEvent struct {
	field  int
	press  bool
	button uint16
	stickX *int8
	stickY *int8
}

// parseKeys reads "FIELD:+start,FIELD:-start,FIELD:x=40,FIELD:y=-20".
func parseKeys(s string) ([]keyEvent, error) {
	var out []keyEvent
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		i := strings.IndexByte(tok, ':')
		if i < 0 {
			return nil, fmt.Errorf("-keys: %q has no FIELD:", tok)
		}
		field, err := strconv.Atoi(tok[:i])
		if err != nil {
			return nil, fmt.Errorf("-keys: bad field in %q", tok)
		}
		act := tok[i+1:]
		switch {
		case strings.HasPrefix(act, "x="), strings.HasPrefix(act, "y="):
			v, err := strconv.Atoi(act[2:])
			if err != nil || v < -128 || v > 127 {
				return nil, fmt.Errorf("-keys: bad stick value in %q", tok)
			}
			b := int8(v)
			e := keyEvent{field: field}
			if act[0] == 'x' {
				e.stickX = &b
			} else {
				e.stickY = &b
			}
			out = append(out, e)
		case strings.HasPrefix(act, "+"), strings.HasPrefix(act, "-"):
			bit, ok := buttonNames[strings.ToLower(act[1:])]
			if !ok {
				return nil, fmt.Errorf("-keys: unknown button %q", act[1:])
			}
			out = append(out, keyEvent{field: field, press: act[0] == '+', button: bit})
		default:
			return nil, fmt.Errorf("-keys: %q is not +button, -button, x= or y=", act)
		}
	}
	return out, nil
}

func run(image, stepsS string, trace bool, tracen int, bps, watches, callLogs multiFlag, rwatch, keys string,
	shot string, shotEvery, shotBase int, dmaLog string, stopField int,
	saveState, loadState, out, pcmDump string) error {
	rom, err := n64.Load(image)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "bootoracle: %s\n", rom)
	fmt.Fprintf(os.Stderr, "bootoracle: md5 %s\n", rom.MD5)

	budget, err := num(stepsS)
	if err != nil {
		return fmt.Errorf("bad -steps %q", stepsS)
	}

	m := n64.NewMachine(rom)
	if err := m.Boot(rom, n64.DefaultBoot()); err != nil {
		return err
	}
	if loadState != "" {
		if err := m.LoadState(loadState); err != nil {
			return fmt.Errorf("-loadstate: %w", err)
		}
		fmt.Fprintf(os.Stderr, "bootoracle: restored %s\n", loadState)
	}

	for _, s := range bps {
		a, err := hx(s)
		if err != nil {
			return fmt.Errorf("bad -bp %q", s)
		}
		m.SetBreakpoint(a)
	}
	for _, s := range watches {
		addr, length := s, "4"
		if i := strings.IndexByte(s, ':'); i >= 0 {
			addr, length = s[:i], s[i+1:]
		}
		a, err := hx(addr)
		if err != nil {
			return fmt.Errorf("bad -watch %q", s)
		}
		n, err := strconv.Atoi(length)
		if err != nil {
			return fmt.Errorf("bad -watch length in %q", s)
		}
		m.WatchLo, m.WatchHi = a, a+uint32(n)
		m.OnWrite = func(addr, val, pc uint32) {
			fmt.Fprintf(os.Stderr, "  write [%08X] = %08X  from %08X  (%s)\n", addr, val, pc, m.DisasmAt(uint64(pc)))
		}
	}
	// A run resumed from a savestate starts counting fields at zero, so
	// -shotbase restores the numbering the original boot would have given
	// them. That is what lets a snapshot stand in for the run that made it.
	fields := shotBase
	if shot != "" {
		if err := os.MkdirAll(shot, 0o755); err != nil {
			return err
		}
	}
	var script []keyEvent
	if keys != "" {
		var err error
		if script, err = parseKeys(keys); err != nil {
			return err
		}
		// The controller must be attached before anything is pressed on it.
		m.Controllers[0].Present = true
	}
	m.OnDisplay = func(mm *n64.Machine) {
		fields++
		// Apply the script at a field boundary: the game polls the joybus once
		// per frame, so a change here is seen exactly once, like a real press.
		for _, e := range script {
			if e.field != fields {
				continue
			}
			c := &mm.Controllers[0]
			switch {
			case e.stickX != nil:
				c.StickX = *e.stickX
			case e.stickY != nil:
				c.StickY = *e.stickY
			case e.press:
				c.Buttons |= e.button
			default:
				c.Buttons &^= e.button
			}
			fmt.Fprintf(os.Stderr, "  field %d: controller now buttons=%04X stick=(%d,%d)\n",
				fields, c.Buttons, c.StickX, c.StickY)
		}
		if stopField > 0 && fields >= stopField {
			mm.StopRequested = true
		}
		if shot == "" || fields%shotEvery != 0 {
			return
		}
		p := filepath.Join(shot, fmt.Sprintf("frame-%04d.png", fields))
		if err := mm.Screenshot(p); err != nil {
			fmt.Fprintf(os.Stderr, "  field %d: %v\n", fields, err)
			return
		}
		fmt.Fprintf(os.Stderr, "  wrote %s\n", p)
	}
	if dmaLog != "" {
		f, err := os.Create(dmaLog)
		if err != nil {
			return fmt.Errorf("-dmalog: %w", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()
		fmt.Fprintf(w, "# kind field dram(cart-side for sp) src length\n")
		m.OnDMA = func(kind string, dramAddr, cartAddr, length uint32) {
			fmt.Fprintf(w, "%s %d %08X %08X %X\n", kind, fields, dramAddr, cartAddr, length)
		}
	}
	if pcmDump != "" {
		f, err := os.Create(pcmDump)
		if err != nil {
			return fmt.Errorf("-pcmdump: %w", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()
		var lastRate uint32
		var buffers, samples int
		m.OnAIBuffer = func(dramAddr, length, dacRate uint32) {
			end := dramAddr + length
			if int(end) > len(m.RDRAM) {
				return
			}
			w.Write(m.RDRAM[dramAddr:end])
			lastRate, buffers, samples = dacRate, buffers+1, samples+int(length)/4
		}
		defer func() {
			// The DAC clock is the VI crystal; rate = clock/(dacRate+1). Write it
			// beside the PCM so a renderer can tag the WAV correctly.
			const dacClock = 48681812
			rate := 0
			if lastRate != 0 {
				rate = dacClock / int(lastRate+1)
			}
			os.WriteFile(pcmDump+".rate", []byte(fmt.Sprintf("%d\n", rate)), 0o644)
			fmt.Fprintf(os.Stderr, "bootoracle: captured %d AI buffers, %d stereo samples, rate %d Hz\n", buffers, samples, rate)
		}()
	}
	var sites map[uint32]*readSite
	if rwatch != "" {
		addr, length := rwatch, "4"
		if i := strings.IndexByte(rwatch, ':'); i >= 0 {
			addr, length = rwatch[:i], rwatch[i+1:]
		}
		a, err := hx(addr)
		if err != nil {
			return fmt.Errorf("bad -rwatch %q", rwatch)
		}
		n, err := hx(length)
		if err != nil {
			return fmt.Errorf("bad -rwatch length in %q", rwatch)
		}
		m.RWatchLo, m.RWatchHi = a, a+n
		sites = map[uint32]*readSite{}
		m.OnRead = func(addr, val, pc uint32) {
			s := sites[pc]
			if s == nil {
				s = &readSite{lo: addr, hi: addr}
				sites[pc] = s
			}
			s.count++
			if addr < s.lo {
				s.lo = addr
			}
			if addr > s.hi {
				s.hi = addr
			}
		}
	}
	if len(callLogs) > 0 {
		pcs := map[uint32]bool{}
		for _, s := range callLogs {
			a, err := hx(s)
			if err != nil {
				return fmt.Errorf("bad -calllog %q", s)
			}
			pcs[a] = true
		}
		m.OnStep = func(mm *n64.Machine, pc uint32) {
			if pcs[pc] {
				r := mm.CPU.R
				// v0 as well as the arguments: an allocator's answer is the address
				// everything downstream watches, and it is only ever in v0.
				fmt.Printf("call %08X field=%d a0=%08X a1=%08X a2=%08X a3=%08X v0=%08X ra=%08X\n",
					pc, fields, uint32(r[4]), uint32(r[5]), uint32(r[6]), uint32(r[7]), uint32(r[2]), uint32(r[31]))
			}
		}
	}
	if trace {
		left := tracen
		m.OnStep = func(mm *n64.Machine, pc uint32) {
			if left <= 0 {
				return
			}
			left--
			fmt.Fprintf(os.Stderr, "  %08X  %s\n", pc, mm.DisasmAt(uint64(pc)))
		}
	}

	res := m.Run(budget)
	fmt.Fprintf(os.Stderr, "\nbootoracle: %s\n", res)
	fmt.Fprintf(os.Stderr, "bootoracle: the game polled the controller %d times\n", m.ContPolls)
	if len(m.JoybusCmds) > 0 {
		var ops []int
		for op := range m.JoybusCmds {
			ops = append(ops, int(op))
		}
		sort.Ints(ops)
		fmt.Fprintf(os.Stderr, "bootoracle: joybus commands:")
		for _, op := range ops {
			fmt.Fprintf(os.Stderr, " %02X=%d", op, m.JoybusCmds[byte(op)])
		}
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintf(os.Stderr, "bootoracle: SI DMAs: %d command blocks written, %d read back\n", m.SIWrites, m.SIReads)
	fmt.Fprintf(os.Stderr, "bootoracle: osMemSize = 0x%08X\n", m.OSMemSize())
	if len(m.Log) > 0 {
		fmt.Fprintf(os.Stderr, "bootoracle: machine notes:\n")
		for _, l := range m.Log {
			fmt.Fprintf(os.Stderr, "  - %s\n", l)
		}
	}

	if sites != nil {
		pcs := make([]uint32, 0, len(sites))
		for pc := range sites {
			pcs = append(pcs, pc)
		}
		sort.Slice(pcs, func(i, j int) bool { return sites[pcs[i]].count > sites[pcs[j]].count })
		fmt.Printf("read-watch %08X..%08X: %d reading PCs\n", m.RWatchLo, m.RWatchHi, len(pcs))
		fmt.Printf("%-10s %9s  %-19s %s\n", "pc", "reads", "range", "instruction")
		for _, pc := range pcs {
			s := sites[pc]
			fmt.Printf("%08X   %9d  %06X..%06X  %s\n", pc, s.count, s.lo, s.hi, m.DisasmAt(uint64(pc)))
		}
	}

	if saveState != "" {
		if err := m.SaveState(saveState); err != nil {
			return fmt.Errorf("-savestate: %w", err)
		}
		fmt.Fprintf(os.Stderr, "bootoracle: wrote %s\n", saveState)
	}
	_ = out
	return nil
}
