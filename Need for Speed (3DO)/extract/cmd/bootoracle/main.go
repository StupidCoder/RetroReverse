// bootoracle loads Need for Speed's LaunchMe (or any 3DO AIF program) into the
// tools/threedo ARM60 machine and runs it, tracing the startup: the executed
// instructions and the Portfolio folio/kernel calls it makes. This is the 3DO
// equivalent of the Ridge Racer PSX bootoracle.
//
//	bootoracle -image "Need for Speed.bin" -prog LaunchMe -steps 200000 -trace
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	prog := flag.String("prog", "LaunchMe", "AIF program path within the disc")
	file := flag.String("f", "", "load a standalone AIF file instead of -image/-prog")
	steps := flag.Uint64("steps", 200000, "max instructions to run")
	trace := flag.Bool("trace", false, "print each executed instruction")
	tracen := flag.Uint64("tracen", 0, "print only the first N executed instructions")
	flag.Parse()

	var data []byte
	var err error
	if *file != "" {
		data, err = os.ReadFile(*file)
	} else if *image != "" {
		var vol *threedo.Volume
		if data, err = os.ReadFile(*image); err == nil {
			var v *threedo.Volume
			if v, err = threedo.Open(data); err == nil {
				vol = v
				data, err = vol.ReadFile(*prog)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "usage: bootoracle -image DISC -prog LaunchMe | -f FILE [-steps N] [-trace]")
		os.Exit(2)
	}
	if err != nil {
		die(err)
	}

	aif, err := threedo.ParseAIF(data)
	if err != nil {
		die(err)
	}
	fmt.Print(aif.Describe())

	m := threedo.NewMachine()
	m.LoadAIF(aif)

	if *trace || *tracen > 0 {
		var n uint64
		limit := *tracen
		m.OnStep = func(mm *threedo.Machine, pc uint32) {
			if limit == 0 || n < limit {
				fmt.Println(" ", mm.DisasmAt(pc))
				n++
			}
		}
	}

	fmt.Printf("\n--- running (max %d steps) ---\n", *steps)
	res := m.Run(*steps)
	fmt.Printf("stopped: %s  after %d steps, pc=0x%08X\n", res.Reason, res.Steps, res.PC)

	if tty := m.TTY(); tty != "" {
		fmt.Printf("\n[TTY]\n%s\n", tty)
	}
	fmt.Printf("\n--- Portfolio folio/kernel calls (%d) ---\n", len(m.KernelCalls))
	seen := map[uint32]int{}
	for _, k := range m.KernelCalls {
		seen[k.Offset]++
	}
	shown := 0
	for _, k := range m.KernelCalls {
		if shown < 24 {
			fmt.Printf("  folio[-0x%X] from 0x%08X  args=%08X %08X %08X %08X\n",
				k.Offset, k.From, k.Args[0], k.Args[1], k.Args[2], k.Args[3])
			shown++
		}
	}
	fmt.Printf("  (%d distinct folio offsets)\n", len(seen))
	if len(m.Log) > 0 {
		fmt.Printf("\n--- notes ---\n")
		for _, s := range m.Log {
			fmt.Println(" ", s)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "bootoracle:", err)
	os.Exit(1)
}
