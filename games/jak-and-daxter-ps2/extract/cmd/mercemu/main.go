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
	// low-mem 0-7 as initialize-chain uploads them (from *math-camera*):
	// VU0 = GIF tag template, VU1 = adgif tag, VU2 = hvdf offset,
	// VU3-6 = camera matrix (identity here: object-space out), VU7 = fog.
	putw(low, 0*16+4, 0x301E4000)
	putw(low, 0*16+8, 0x412)
	putw(low, 1*16+0, 5)
	putw(low, 1*16+4, 0x10000000)
	putw(low, 1*16+8, 14)
	// The real *math-camera* values captured at title-logo: perspective
	// rows (VU3-6) and the hvdf offset (VU2). row3.z's 0x4B4807A9 is the
	// mantissa-trick bias that turns lump dest bytes into VU addresses.
	campos := [][4]uint32{
		{0xBED66E1D, 0, 0, 0},
		{0, 0xBE7A2B23, 0, 0},
		{0, 0, 0xC5C808F1, 0xB8151DE3},
		{0, 0, 0x4B4807A9, 0},
	}
	for r, row := range campos {
		for k, v := range row {
			putw(low, (3+r)*16+k*4, v)
		}
	}
	for k, v := range []uint32{0x45000000, 0x4AFFBF9B, 0x438428EE, 0x40800000} {
		putw(low, 2*16+k*4, v)
	}
	for k, v := range []uint32{0xBD3EA7E6, 0x3F800000, 0x3F800000, 0} {
		putw(low, 7*16+k*4, v)
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
	cfg := &merc.EmuConfig{Micro: micro, LowMem: low, Bones: bones, Entry: 0x88, Init: true, Top: 0}
	seen := map[uint32]bool{}
	stores := 0
	cfg.TracePC = func(v *vu.VU, pc uint32) {
		if (pc == 0x3A00 || pc == 0x3A58 || pc == 0x3C40 || pc == 0x498 || pc == 0x9F0) && !seen[pc] {
			seen[pc] = true
			fmt.Printf("pc %04X: vi01=%d vi02=%d vi03=%d vi04=%d vi05=%d vi06=%d vi07=%d vi08=%d vi12=%d vi15=%d\n",
				pc, v.VI[1], v.VI[2], v.VI[3], v.VI[4], v.VI[5], v.VI[6], v.VI[7], v.VI[8], v.VI[12], v.VI[15])
		}
		// log sq stores in the vertex loop region
		if stores < 60 && pc >= 0x600 && pc < 0xA00 {
			raw := binary.LittleEndian.Uint64(micro[pc:])
			txt := vu.Decode(raw, pc).Lower
			if len(txt) > 2 && txt[:2] == "sq" {
				// parse "sq.xyzw vfNN, OFF(viMM)"
				var vf, off, vi int
				n, _ := fmt.Sscanf(txt, "sq.xyzw vf%d, %d(vi%d)", &vf, &off, &vi)
				if n == 3 {
					fmt.Printf("  store pc %04X vf%02d -> %d(vi%02d)=%d\n", pc, vf, off, vi, int(int16(v.VI[vi]))+off)
					stores++
				}
			}
		}
	}
	cfg.DumpMem = func(mem []byte) {
		// report nonzero quadword runs outside input/low-mem
		start := -1
		for q := 371; q < 1024; q++ {
			nz := false
			for k := 0; k < 16; k++ {
				if mem[q*16+k] != 0 {
					nz = true
					break
				}
			}
			if nz && start < 0 {
				start = q
			}
			if (!nz || q == 1023) && start >= 0 {
				fmt.Printf("  mem qw %d..%d nonzero\n", start, q-1)
				start = -1
			}
		}
	}
	cfg.ParseAt = 377
	cfg.STMagic = binary.LittleEndian.Uint32(obj[0x1244+44:])
	vs, err := merc.Emulate(cfg, fr, cfg.STMagic)
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
	for ei := range c.Effects {
		e := &c.Effects[ei]
		for fi := range e.Fragments {
			f2 := &e.Fragments[fi]
			cfg2 := *cfg
			cfg2.Bones = merc.IdentityBones(f2)
			topo, err := merc.EmulateTopology(&cfg2, f2)
			nv := f2.LumpQWC / 3
			valid := 0
			if err == nil {
				for _, t := range topo {
					if t.Index >= 0 && t.Index < nv {
						valid++
					}
				}
			}
			status := ""
			if err != nil {
				status = "ERR " + err.Error()[:30]
			}
			if fi < 6 || valid*10 < len(topo)*9 {
				fmt.Printf("  e%d f%d: nv=%d topo=%d valid=%d hdr=%v %s\n", ei, fi, nv, len(topo), valid, f2.ByteData[:16], status)
			}
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercemu:", err)
		os.Exit(1)
	}
}
