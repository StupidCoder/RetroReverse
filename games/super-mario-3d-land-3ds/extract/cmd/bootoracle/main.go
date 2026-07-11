// bootoracle boots Super Mario 3D Land's ARM11 application code on the shared
// 3DS machine (tools/platform/n3ds) and runs it, exposing the standard oracle
// instrumentation (STANDARDS §3). It is the executable counterpart of the static
// tools: where n3dsdump reads the cartridge's containers, this one runs the code
// they hold.
//
// The machine is a first-milestone process-level model: it loads the code
// segments, lays out the userland memory map, and high-level-emulates the Horizon
// supervisor calls needed to run the C runtime. It stops explicitly at the first
// kernel facility not yet implemented (typically the srv:/GSP IPC handshake), so a
// run reports concretely how far the bring-up reaches rather than diverging
// silently. See the game writeup for the current frontier.
//
// Usage:
//
//	bootoracle -image game.cci [-steps N] [-trace] [-tracen N] [-bp A]... [-watch A[:L]]...
//	           [-svclog] [-savestate F] [-loadstate F]
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

// parseNum accepts hex (0x…) or decimal, matching the other oracles.
func parseNum(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func main() {
	image := flag.String("image", "", "3DS cartridge image (decrypted .cci/.3ds)")
	steps := flag.String("steps", "2000000", "instruction budget (hex with 0x, else decimal)")
	trace := flag.Bool("trace", false, "trace executed instructions")
	tracen := flag.Int("tracen", 200, "limit -trace to this many instructions")
	verbose := flag.Bool("v", false, "log every supervisor call and unmapped access as it happens")
	svclog := flag.Bool("svclog", false, "print the ordered supervisor-call log after the run")
	var bps, watches multiFlag
	flag.Var(&bps, "bp", "breakpoint address (hex); repeatable")
	flag.Var(&watches, "watch", "memory watch ADDR[:LEN] (hex); repeatable")
	saveState := flag.String("savestate", "", "after the run, dump the machine snapshot to this file")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before running")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "bootoracle: -image is required")
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*image, *steps, *trace, *tracen, *verbose, *svclog, bps, watches, *saveState, *loadState); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

func run(imagePath, stepsStr string, trace bool, tracen int, verbose, svclog bool, bps, watches multiFlag, saveState, loadState string) error {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}
	m, err := n3ds.NewMachine(img)
	if err != nil {
		return err
	}
	m.Verbose = verbose
	m.SetTrace(trace, tracen)

	for _, b := range bps {
		v, err := parseNum(b)
		if err != nil {
			return fmt.Errorf("bad -bp %q: %w", b, err)
		}
		m.AddBreakpoint(uint32(v))
	}
	for _, w := range watches {
		addr, length := w, "4"
		if i := strings.IndexByte(w, ':'); i >= 0 {
			addr, length = w[:i], w[i+1:]
		}
		a, err := parseNum(addr)
		if err != nil {
			return fmt.Errorf("bad -watch %q: %w", w, err)
		}
		l, err := parseNum(length)
		if err != nil {
			return fmt.Errorf("bad -watch length %q: %w", w, err)
		}
		m.AddWatch(uint32(a), uint32(l))
	}

	if loadState != "" {
		if err := m.LoadState(loadState); err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
		fmt.Printf("restored snapshot from %s (at %d instructions)\n", loadState, m.Instrs())
	}

	budget, err := parseNum(stepsStr)
	if err != nil {
		return fmt.Errorf("bad -steps: %w", err)
	}

	fmt.Printf("boot: entry 0x%08X, running up to %d instructions\n", m.Entry(), budget)
	ran := m.Run(int(budget))

	fmt.Printf("\nran %d instructions (%d total)\n", ran, m.Instrs())
	if r := m.HaltReason(); r != "" {
		fmt.Printf("halt: %s\n", r)
	} else {
		fmt.Printf("still runnable (budget reached)\n")
	}
	if dbg := m.DebugString(); dbg != "" {
		fmt.Printf("\nOutputDebugString:\n%s\n", dbg)
	}
	if ports := m.Ports(); len(ports) > 0 {
		fmt.Printf("\nservice ports connected:\n")
		for h, name := range ports {
			fmt.Printf("  0x%08X  %s\n", h, name)
		}
	}
	if ipc := m.IPCLog(); len(ipc) > 0 {
		fmt.Printf("\nIPC requests: %d\n", len(ipc))
		counts := map[string]int{}
		for _, e := range ipc {
			counts[e.Service()]++
		}
		for svc, n := range counts {
			fmt.Printf("  %-10s %d requests\n", svc, n)
		}
	}
	if vb := m.VBlanks(); vb > 0 {
		sub, swp := m.FrameStats()
		fmt.Printf("\ngraphics: %d VBlanks delivered, %d GPU command lists submitted, %d frame swaps\n", vb, sub, swp)
	}
	if svclog {
		printSVCSummary(m)
	}

	if saveState != "" {
		if err := m.SaveState(saveState); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		fmt.Printf("\nwrote snapshot to %s\n", saveState)
	}
	return nil
}

func printSVCSummary(m *n3ds.Machine) {
	log := m.SVCLog()
	fmt.Printf("\nsupervisor calls: %d total\n", len(log))
	counts := map[string]int{}
	for _, e := range log {
		counts[e.Name]++
	}
	// Print the first several in order, then the histogram.
	fmt.Println("  first calls, in order:")
	for i, e := range log {
		if i >= 40 {
			fmt.Printf("  … and %d more\n", len(log)-i)
			break
		}
		fmt.Printf("    %3d  0x%02X %-20s r0=%08X r1=%08X r2=%08X r3=%08X\n",
			i, e.Num, e.Name, e.Args[0], e.Args[1], e.Args[2], e.Args[3])
	}
	fmt.Println("  histogram:")
	for name, n := range counts {
		fmt.Printf("    %-24s %d\n", name, n)
	}
}
