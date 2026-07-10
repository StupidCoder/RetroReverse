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
	saveState := flag.String("savestate", "", "after the run, dump the full machine snapshot to this file")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before running")
	out := flag.String("o", "", "output directory")
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
	if err := run(*image, *steps, *trace, *tracen, bps, watches, callLogs,
		*shot, *shotEvery, *shotBase, *dmaLog, *stopField, *saveState, *loadState, *out); err != nil {
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

func run(image, stepsS string, trace bool, tracen int, bps, watches, callLogs multiFlag,
	shot string, shotEvery, shotBase int, dmaLog string, stopField int,
	saveState, loadState, out string) error {
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
	m.OnDisplay = func(mm *n64.Machine) {
		fields++
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
				fmt.Printf("call %08X field=%d a0=%08X a1=%08X a2=%08X a3=%08X ra=%08X\n",
					pc, fields, uint32(r[4]), uint32(r[5]), uint32(r[6]), uint32(r[7]), uint32(r[31]))
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
	fmt.Fprintf(os.Stderr, "bootoracle: osMemSize = 0x%08X\n", m.OSMemSize())
	if len(m.Log) > 0 {
		fmt.Fprintf(os.Stderr, "bootoracle: machine notes:\n")
		for _, l := range m.Log {
			fmt.Fprintf(os.Stderr, "  - %s\n", l)
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
