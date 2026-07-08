// dualoracle co-executes Super Mario 64 DS's ARM9 and ARM7 boot binaries on two
// tools/arm cores sharing one main RAM, wired by the DS's IPC hardware (the
// tools/nds/dsmachine model). Where the single-CPU bootoracle (Part III) stops at
// the ARM9↔ARM7 rendezvous, this lets the ARM7 answer the handshake so the ARM9
// clears it and runs on into main() and the frame loop.
//
//	dualoracle [-budget N] [-quantum N] [-log] rom.nds
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/dsmachine"
)

const dtcm9 = 0x023C0000 // the ARM9 DTCM base this game programs (Part II)

func main() {
	budget := flag.Uint64("budget", 200_000_000, "instruction budget (both cores)")
	quantum := flag.Int("quantum", 64, "instructions per core per scheduler round")
	showLog := flag.Bool("log", false, "list the stubbed hardware accesses the model satisfied")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: dualoracle [-budget N] [-quantum N] [-log] rom.nds")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(data)
	if err != nil {
		die(err)
	}

	m := dsmachine.New(rom, dtcm9)
	milestones := map[uint32]string{
		0x0205BB54: "IPCSYNC rendezvous poll (a single core stops here)",
		0x02059E48: "past the rendezvous → post-sync PXI init",
		0x0205B864: "receiving the ARM7 boot message (PXI FIFO)",
		0x02007000: "main()",
	}
	res := m.Run(*budget, *quantum, milestones)

	s9, s7 := m.SyncNibbles()
	cleared := res.ARM9Milest[0x02059E48] != 0
	fmt.Println()
	if cleared {
		fmt.Println("=> the ARM9↔ARM7 IPCSYNC rendezvous is CLEARED (the single-core oracle cannot pass it):")
		fmt.Println("   both sync nibbles ratcheted to 0 and the ARM9 ran on into the post-sync PXI exchange.")
	}
	fmt.Printf("stopped: %s\n", res.Reason)
	fmt.Printf("rounds: %d\n", res.Steps)
	fmt.Printf("ARM9 final PC 0x%08X (parked=%v)   ARM7 final PC 0x%08X (parked=%v)\n",
		m.ARM9PC(), m.Parked(true), m.ARM7PC(), m.Parked(false))
	fmt.Printf("IPCSYNC nibbles: ARM9 out=%d  ARM7 out=%d\n", s9, s7)

	if *showLog {
		fmt.Printf("stubs satisfied (%d):\n", len(m.Log))
		for _, s := range m.Log {
			fmt.Println("  " + s)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "dualoracle:", err)
	os.Exit(1)
}
