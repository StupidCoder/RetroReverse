// ramdiff loads two bootoracle savestates of the same disc and reports every main-RAM word
// that differs, grouped into runs — the working set of whatever code ran between the two
// snapshots. Two states taken one video field apart under an idle boot expose exactly the
// globals the interrupt handlers touch, which is how a stalled boot's last living code is
// found without symbols.
//
// usage: ramdiff -image DISC.iso -a STATE1 -b STATE2 [-max N]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "disc image both states belong to")
	a := flag.String("a", "", "first savestate")
	b := flag.String("b", "", "second savestate")
	max := flag.Int("max", 200, "maximum diff runs to print")
	flag.Parse()

	ma, err := load(*image, *a)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ramdiff:", err)
		os.Exit(1)
	}
	mb, err := load(*image, *b)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ramdiff:", err)
		os.Exit(1)
	}

	runs := 0
	i := 0
	for i < len(ma.RAM) && i < len(mb.RAM) {
		if ma.RAM[i] == mb.RAM[i] {
			i++
			continue
		}
		start := i
		for i < len(ma.RAM) && i < len(mb.RAM) && ma.RAM[i] != mb.RAM[i] {
			i++
		}
		runs++
		if runs <= *max {
			fmt.Printf("0x%08X..0x%08X (%d bytes)  a=% X  b=% X\n",
				0x80000000+start, 0x80000000+i, i-start,
				ma.RAM[start:min(i, start+16)], mb.RAM[start:min(i, start+16)])
		}
	}
	fmt.Printf("%d differing runs\n", runs)
}

func load(image, state string) (*gc.Machine, error) {
	disc, err := gc.Open(image)
	if err != nil {
		return nil, err
	}
	m, err := gc.NewMachine(disc)
	if err != nil {
		return nil, err
	}
	if err := m.LoadStateFile(state); err != nil {
		return nil, err
	}
	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
