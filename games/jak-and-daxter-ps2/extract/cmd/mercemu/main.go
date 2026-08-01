package main

// mercemu: run the merc microprogram on a fragment with a semantically
// chosen low-mem block (neutral camera/lights) and report the kicked
// strip structure.
import (
	"encoding/binary"
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/cpu/vu"
)

func putw(b []byte, o int, v uint32) { binary.LittleEndian.PutUint32(b[o:], v) }

func main() {
	obj, err := os.ReadFile(os.Args[1])
	check(err)
	micro, err := os.ReadFile(os.Args[2])
	check(err)
	c, err := merc.Parse(obj, 0x1244)
	check(err)
	fr := &c.Effects[0].Fragments[0]

	low := make([]byte, 140*16)
	// qw0: the GIF tag template the code copies to the output header
	// (NLOOP patched at runtime; NREG=3, regs ST,RGBAQ,XYZF2).
	putw(low, 0*16+4, 0x301E4000)
	putw(low, 0*16+8, 0x412)
	// qw1: the second tag template (materials path).
	putw(low, 1*16+0, 5)
	putw(low, 1*16+4, 0x10000000)
	// qw2: XYZ offset = 0. qw5/6/7: benign scales.
	for _, q := range []int{5, 6, 7} {
		for k := 0; k < 4; k++ {
			putw(low, q*16+k*4, 0x3F800000)
		}
	}
	// 132-138: light rows zero, q divisor 132.w = 1, ambient 138 = 1s.
	putw(low, 132*16+12, 0x3F800000)
	for k := 0; k < 4; k++ {
		putw(low, 138*16+k*4, 0x3F800000)
	}
	// 139: ctrl+28 row from the object (the merc-ctrl float block).
	for k := 0; k < 4; k++ {
		putw(low, 139*16+k*4, binary.LittleEndian.Uint32(obj[0x1244+28+k*4:]))
	}

	bones := map[byte][]byte{}
	ident := make([]byte, 7*16)
	for r := 0; r < 4; r++ {
		putw(ident, r*16+r*4, 0x3F800000)
	}
	for r := 4; r < 7; r++ {
		putw(ident, r*16+(r-4)*4, 0x3F800000)
	}
	for _, m := range fr.Mats {
		bones[m.Index] = ident
	}
	cfg := &merc.EmuConfig{Micro: micro, LowMem: low, Bones: bones, Entry: 0x128, Init: true, Top: 0}
	seen := map[uint32]bool{}
	cfg.TracePC = func(v *vu.VU, pc uint32) {
		if (pc == 0x3A00 || pc == 0x3A58 || pc == 0x3C40 || pc == 0x498 || pc == 0x9F0) && !seen[pc] {
			seen[pc] = true
			fmt.Printf("pc %04X: vi01=%d vi02=%d vi03=%d vi04=%d vi05=%d vi06=%d vi07=%d vi08=%d vi12=%d vi15=%d\n",
				pc, v.VI[1], v.VI[2], v.VI[3], v.VI[4], v.VI[5], v.VI[6], v.VI[7], v.VI[8], v.VI[12], v.VI[15])
		}
	}
	vs, err := merc.Emulate(cfg, fr, 0x4B010000)
	check(err)
	fmt.Printf("%d out-verts\n", len(vs))
	adc := ""
	for i, v := range vs {
		if i >= 80 {
			break
		}
		if v.ADC {
			adc += "A"
		} else {
			adc += "."
		}
	}
	fmt.Println("adc:", adc)
	for i := 0; i < 4 && i < len(vs); i++ {
		fmt.Printf("v%d: st=%.3f,%.3f q=%.4f rgba=%v xyz=%v adc=%v\n", i,
			vs[i].ST[0], vs[i].ST[1], vs[i].ST[2], vs[i].RGBA, vs[i].XYZ, vs[i].ADC)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercemu:", err)
		os.Exit(1)
	}
}
