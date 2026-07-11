// bootoracle boots Loco Roco's executable under the PSP oracle (tools/platform/psp)
// and runs it, exposing the repo's reverse-engineering instrumentation: an
// instruction trace, breakpoints, memory watches, the kernel syscall census, the
// game's Kprintf/stdout, and a savestate. It reads the UMD image (a .cso CISO or a
// .iso), KIRK-decrypts EBOOT.BIN, relocates and loads the module — no manual
// extraction step.
//
// Usage:
//
//	bootoracle -image image/LocoRoco.cso [-steps N] [-trace -tracen N]
//	           [-bp ADDR] [-watch ADDR[:LEN]] [-savestate FILE] [-loadstate FILE]
//
// The oracle boots the C runtime and module start, creates and starts the main
// thread, and runs until it reaches the kernel-HLE wall; -trace shows live execution,
// -watch maps the code that produces a memory structure, and the syscall census names
// the kernel functions the boot path invokes.
package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/psp"
)

func hx(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x")
	v, err := strconv.ParseUint(s, 16, 64)
	return uint32(v), err
}

func parseCount(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		v, err := hx(s)
		return uint64(v), err
	}
	return strconv.ParseUint(s, 10, 64)
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bootoracle: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PSP UMD image (.cso or .iso)")
	exe := flag.String("exe", "PSP_GAME/SYSDIR/EBOOT.BIN", "boot executable path on the disc")
	stepsF := flag.String("steps", "20000000", "max instructions to run (hex or decimal)")
	trace := flag.Bool("trace", false, "disassemble each executed instruction")
	tracen := flag.Int("tracen", 200, "with -trace, stop printing after this many instructions")
	bpS := flag.String("bp", "", "breakpoint: stop when PC reaches this address (hex)")
	watchS := flag.String("watch", "", "watch: log who writes ADDR or ADDR:LEN (hex)")
	rwatchS := flag.String("rwatch", "", "watch: log who reads ADDR or ADDR:LEN (hex)")
	watchN := flag.Int("watchn", 50, "with -watch/-rwatch, stop printing raw accesses after this many")
	saveS := flag.String("savestate", "", "after the run, write a machine savestate to this file")
	loadS := flag.String("loadstate", "", "before the run, restore a machine savestate from this file")
	shot := flag.String("shot", "", "after the run, write the display framebuffer to this PNG")
	showNotes := flag.Bool("notes", false, "print the machine's diagnostic notes")
	flag.Parse()
	if *image == "" {
		die("need -image")
	}

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	im, err := psp.OpenImage(*image)
	if err != nil {
		die("%v", err)
	}
	defer im.Close()
	fmt.Fprintf(os.Stderr, "volume: system=%q name=%q\n", im.System, im.Name)

	mod, err := im.LoadExecutable(*exe)
	if err != nil {
		die("load %s: %v", *exe, err)
	}
	fmt.Fprintf(os.Stderr, "module %q: entry 0x%08X gp 0x%08X, %d imports\n",
		mod.Name, mod.EntryPC, mod.GP, len(mod.Imports))

	steps, err := parseCount(*stepsF)
	if err != nil {
		die("bad -steps")
	}

	m := psp.NewMachine()
	m.SetImageHash(fmt.Sprintf("%x", md5.Sum(data)))
	if err := m.LoadModule(mod); err != nil {
		die("load module: %v", err)
	}
	if *loadS != "" {
		if err := m.LoadStateFile(*loadS); err != nil {
			die("load state: %v", err)
		}
		fmt.Fprintf(os.Stderr, "restored state: PC 0x%08X, %d instructions\n", m.CPU.PC, m.CPU.Steps)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// Instrumentation.
	var bp uint32
	haveBP := false
	if *bpS != "" {
		if bp, err = hx(*bpS); err != nil {
			die("bad -bp")
		}
		haveBP = true
	}
	if *watchS != "" {
		lo, ln := parseWatch(*watchS)
		m.WatchLo, m.WatchHi = lo, lo+ln
		raw := 0
		m.OnWrite = func(addr, val, pc uint32) {
			if raw < *watchN {
				fmt.Fprintf(w, "write [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
				raw++
			}
		}
	}
	if *rwatchS != "" {
		lo, ln := parseWatch(*rwatchS)
		m.RWatchLo, m.RWatchHi = lo, lo+ln
		raw := 0
		m.OnRead = func(addr, val, pc uint32) {
			if raw < *watchN {
				fmt.Fprintf(w, "read  [0x%08X] = 0x%02X  by PC 0x%08X\n", addr, val, pc)
				raw++
			}
		}
	}
	traced := 0
	hitBP := false
	var bpPC uint32
	if *trace || haveBP {
		m.OnStep = func(mm *psp.Machine, pc uint32) {
			if haveBP && !hitBP && (pc&0x1FFFFFFF) == (bp&0x1FFFFFFF) {
				hitBP, bpPC = true, pc
			}
			if *trace && traced < *tracen {
				fmt.Fprintf(w, "%08X  %s\n", pc, strings.TrimSpace(mm.DisasmAt(pc)))
				traced++
			}
		}
	}

	res := m.Run(steps)
	if hitBP {
		res.PC, res.Reason = bpPC, "breakpoint"
	}

	if *saveS != "" {
		if err := m.SaveStateFile(*saveS); err != nil {
			die("save state: %v", err)
		}
		fmt.Fprintf(os.Stderr, "saved state to %s (at %d instructions)\n", *saveS, m.CPU.Steps)
	}

	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			die("screenshot: %v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote framebuffer (%s) to %s\n", m.FramebufferInfo(), *shot)
	}

	w.Flush()
	fmt.Fprintf(os.Stderr, "\n%s\n", res)
	if tty := m.TTY(); tty != "" {
		fmt.Fprintf(os.Stderr, "TTY:\n%s\n", tty)
	}
	printCensus(m)
	if *showNotes {
		for _, l := range m.Log {
			fmt.Fprintln(os.Stderr, "note:", l)
		}
	}
}

func printCensus(m *psp.Machine) {
	type kv struct {
		name string
		n    int
	}
	var all []kv
	for k, v := range m.SyscallCalls {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].n > all[j].n })
	fmt.Fprintf(os.Stderr, "kernel syscalls reached (%d):\n", len(all))
	for _, e := range all {
		fmt.Fprintf(os.Stderr, "  %5d  %s\n", e.n, e.name)
	}
}

func parseWatch(s string) (lo, ln uint32) {
	parts := strings.SplitN(s, ":", 2)
	lo, _ = hx(parts[0])
	ln = 1
	if len(parts) == 2 {
		ln, _ = hx(parts[1])
	}
	return
}
