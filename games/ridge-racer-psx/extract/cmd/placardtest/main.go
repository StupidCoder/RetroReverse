// placardtest probes the warning-placard physics: it boots into the race,
// moves placard record 0 (table 0x8006E85C) onto the racing line just past the
// start, drives into it (hold cross), and reports every non-drawer PC that
// reads the table plus every write to it — the collision/knock-over code, if
// any exists.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/platform/psx"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "placardtest: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	shots := flag.String("shots", "", "screenshot dir")
	flag.Parse()
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
	script, err := psx.ParsePress("start@380000000:380000,cross@386000000:380000,cross@430000000:1200000000")
	if err != nil {
		die("press: %v", err)
	}
	m.PadScript = script
	m.LoadEXE(exe)

	m.Run(429_500_000) // in the race, pre-GO

	// Move placard 0 in front of the grid: the start straight runs toward -X
	// at Z≈39344 (the girl stands at (34048, 0, 39344)). Place it a little
	// down the road on the racing line.
	poke32 := func(addr, v uint32) {
		for i := 0; i < 4; i++ {
			m.Write(addr+uint32(i), byte(v>>(8*uint(i))))
		}
	}
	poke32(0x8006E860, uint32(int32(31500))) // X
	poke32(0x8006E864, 0)                    // Y
	poke32(0x8006E868, uint32(int32(39344))) // Z
	fmt.Fprintln(os.Stderr, "placard 0 moved to (31500,0,39344)")

	// Watch the table: reads by PC (drawers are known), writes by PC.
	readPCs := map[uint32]int{}
	writePCs := map[uint32]int{}
	m.RWatchLo, m.RWatchHi = 0x8006E85C, 0x8006E85C+0x90
	m.OnRead = func(addr, val, pc uint32) { readPCs[pc]++ }
	m.WatchLo, m.WatchHi = 0x8006E85C, 0x8006E85C+0x90
	m.OnWrite = func(addr, val, pc uint32) {
		writePCs[pc]++
		if writePCs[pc] < 5 {
			fmt.Fprintf(os.Stderr, "WRITE [%06X] = %02X by PC %08X\n", addr, val, pc)
		}
	}

	if *shots != "" {
		os.MkdirAll(*shots, 0o755)
	}
	for i := 0; i < 12; i++ {
		m.Run(100_000_000)
		if *shots != "" {
			m.Screenshot(fmt.Sprintf("%s/p%02d.png", *shots, i))
		}
	}

	fmt.Println("table readers:")
	var pcs []uint32
	for pc := range readPCs {
		pcs = append(pcs, pc)
	}
	sort.Slice(pcs, func(i, j int) bool { return pcs[i] < pcs[j] })
	for _, pc := range pcs {
		fmt.Printf("  PC %08X  %d reads\n", pc, readPCs[pc])
	}
	fmt.Printf("table writers: %d PCs\n", len(writePCs))
	for pc, n := range writePCs {
		fmt.Printf("  PC %08X  %d writes\n", pc, n)
	}
}
