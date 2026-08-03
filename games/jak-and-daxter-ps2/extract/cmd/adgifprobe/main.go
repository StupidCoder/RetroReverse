package main

// adgifprobe: dump the A+D register blocks in merc fragment fp regions.
// With -ram FILE -base ADDR, also show the same quadwords from a live RAM
// image (where the engine's login has filled the TEX0/TEX1 templates) —
// the instrument for learning the texture-id -> TEX0 mapping.
import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

var regName = map[byte]string{
	0x06: "TEX0_1", 0x08: "CLAMP_1", 0x14: "TEX1_1", 0x34: "MIPTBP1_1", 0x42: "ALPHA_1",
}

func tex0(lo, hi uint32) string {
	v := uint64(lo) | uint64(hi)<<32
	return fmt.Sprintf("TBP0=%04x TBW=%d PSM=%02x %dx%d TCC=%d TFX=%d CBP=%04x CPSM=%x CSA=%d",
		v&0x3FFF, (v>>14)&0x3F, (v>>20)&0x3F, 1<<((v>>26)&0xF), 1<<((v>>30)&0xF),
		(v>>34)&1, (v>>35)&3, (v>>37)&0x3FFF, (v>>51)&0xF, (v>>56)&0x1F)
}

func main() {
	ram := flag.String("ram", "", "RAM image")
	base := flag.String("base", "0", "art group base in RAM (hex)")
	flag.Parse()
	obj, _ := os.ReadFile(flag.Arg(0))
	var rdata []byte
	var rbase uint64
	if *ram != "" {
		rdata, _ = os.ReadFile(*ram)
		rbase, _ = strconv.ParseUint(*base, 0, 64)
	}
	for _, a := range flag.Args()[1:] {
		p, _ := strconv.ParseUint(a, 0, 32)
		c, err := merc.Parse(obj, uint32(p))
		if err != nil {
			panic(err)
		}
		fmt.Printf("=== ctrl %s\n", a)
		for ei := range c.Effects {
			for fi := range c.Effects[ei].Fragments {
				fr := &c.Effects[ei].Fragments[fi]
				if fr.FPQWC <= 2 {
					continue
				}
				fpOff := int(fr.Off) + (fr.ByteQWC+3)/4*16 + (fr.LumpQWC+3)/4*16
				fmt.Printf("e%d f%d fp=%d (fpOff 0x%x):\n", ei, fi, fr.FPQWC, fpOff)
				for q := 1; q*16+15 < len(fr.FPData); q++ {
					w := make([]uint32, 4)
					for k := range w {
						w[k] = binary.LittleEndian.Uint32(fr.FPData[q*16+k*4:])
					}
					reg := byte(w[2])
					n := regName[reg]
					if n == "" {
						n = fmt.Sprintf("reg%02x", reg)
					}
					fmt.Printf("  qw%-2d disc: %08x %08x %08x %08x %-9s\n", q, w[0], w[1], w[2], w[3], n)
					if rdata != nil {
						ro := int(rbase) + fpOff + q*16
						var r [4]uint32
						for k := range r {
							r[k] = binary.LittleEndian.Uint32(rdata[ro+k*4:])
						}
						extra := ""
						if reg == 0x06 {
							extra = "  " + tex0(r[0], r[1])
						}
						fmt.Printf("       ram:  %08x %08x %08x %08x%s\n", r[0], r[1], r[2], r[3], extra)
					}
				}
			}
		}
	}
}
