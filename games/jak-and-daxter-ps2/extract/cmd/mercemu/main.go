package main

// mercemu: smoke harness for the merc VU emulation — run one fragment and
// report kick count, vertex count, ADC pattern.
import (
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, err := os.ReadFile(os.Args[1])
	check(err)
	micro, err := os.ReadFile(os.Args[2])
	check(err)
	c, err := merc.Parse(obj, 0x1244)
	check(err)
	fr := &c.Effects[0].Fragments[0]

	low := make([]byte, 140*16)
	// vf07 row (ambient) = 1s so lighting passes vf23 through; vf01.w != 0.
	for i, f := range []uint32{0x3F800000, 0x3F800000, 0x3F800000, 0x3F800000} {
		putw(low, 138*16+i*4, f)
		putw(low, 7*16+i*4, f)
	}
	bones := map[byte][]byte{}
	ident := make([]byte, 7*16)
	for r := 0; r < 4; r++ {
		putw(ident, r*16+r*4, 0x3F800000)
	}
	for _, m := range fr.Mats {
		bones[m.Index] = ident
	}
	cfg := &merc.EmuConfig{Micro: micro, LowMem: low, Bones: bones, Entry: 0x128, Top: 0}
	vs, err := merc.Emulate(cfg, fr, 0x4B010000)
	check(err)
	fmt.Printf("%d out-verts\n", len(vs))
	adc := ""
	for i, v := range vs {
		if i >= 70 {
			break
		}
		if v.ADC {
			adc += "A"
		} else {
			adc += "."
		}
	}
	fmt.Println("adc pattern:", adc)
	if len(vs) > 2 {
		fmt.Printf("v2: st=%v rgba=%v xyz=%v\n", vs[2].ST, vs[2].RGBA, vs[2].XYZ)
	}
}

func putw(b []byte, o int, v uint32) {
	b[o] = byte(v)
	b[o+1] = byte(v >> 8)
	b[o+2] = byte(v >> 16)
	b[o+3] = byte(v >> 24)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercemu:", err)
		os.Exit(1)
	}
}
