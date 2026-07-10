// pccount traps a list of PCs under the PSX oracle and reports, per (pc, a0),
// how often each was hit and from where ($ra) — the quick "who runs this with
// what" census used to enumerate e.g. all car structs fed to a physics routine.
//
// Usage: pccount -image DISC [-load STATE] [-press ...] -pcs 8001BD80,8001B9D4 [-window N]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/psx"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "pccount: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	load := flag.String("load", "", "machine savestate to restore")
	press := flag.String("press", "", "pad script")
	pcsS := flag.String("pcs", "", "comma-separated PCs to trap (hex)")
	window := flag.Uint64("window", 10_000_000, "steps to run")
	flag.Parse()
	if *image == "" || *pcsS == "" {
		die("need -image and -pcs")
	}
	traps := map[uint32]bool{}
	for _, s := range strings.Split(*pcsS, ",") {
		var pc uint32
		if _, err := fmt.Sscanf(strings.TrimSpace(s), "%x", &pc); err != nil {
			die("bad pc %q", s)
		}
		traps[pc] = true
	}

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	_, exe, err := vol.BootEXE()
	if err != nil {
		die("%v", err)
	}
	m := psx.NewMachine()
	m.SetDisc(vol)
	m.ISRHandler = 0x8004DF48
	if *press != "" {
		script, err := psx.ParsePress(*press)
		if err != nil {
			die("press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)
	if *load != "" {
		if err := m.LoadStateFile(*load); err != nil {
			die("load state: %v", err)
		}
	}

	type key struct{ pc, a0, ra uint32 }
	counts := map[key]int{}
	m.OnStep = func(mm *psx.Machine, pc uint32) {
		if traps[pc] {
			counts[key{pc, mm.CPU.Reg(4), mm.CPU.Reg(31)}]++
		}
	}
	res := m.Run(*window)
	m.OnStep = nil

	keys := make([]key, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].pc != keys[j].pc {
			return keys[i].pc < keys[j].pc
		}
		if keys[i].a0 != keys[j].a0 {
			return keys[i].a0 < keys[j].a0
		}
		return keys[i].ra < keys[j].ra
	})
	for _, k := range keys {
		fmt.Printf("pc=%08X a0=%08X ra=%08X hits=%d\n", k.pc, k.a0, k.ra, counts[k])
	}
	fmt.Fprintln(os.Stderr, res)
}
