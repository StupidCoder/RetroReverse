package main

// gsaudit: replay a captured frame's DMA chains and report the GS register
// state every drawn strip actually uses — PRIM (ABE), TEST (ATE/ATST/AREF),
// ALPHA, TEX0 — from the game's own packets. PATH2 DIRECT payloads (bucket
// init) and PATH1 XGKick packets both feed one shadow GS.
import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

type gsState struct {
	prim          uint64 // last PRIM (A+D or tag PRE)
	test          [2]uint64
	alpha         [2]uint64
	tex0          [2]uint64
	texa          uint64
	pabe          uint64
}

type drawKey struct {
	src   string // "kick@ENTRY" or "direct"
	prim  uint64
	test  uint64
	alpha uint64
	tex0  uint64
}

var draws = map[drawKey]int{}   // -> vertex count
var drawPkts = map[drawKey]int{}

// feed walks a GIF stream (byte slice, 16-byte qwords) updating the shadow
// state and recording draw events.
func (g *gsState) feed(mem []byte, src string) {
	qw := 0
	nqw := len(mem) / 16
	for qw < nqw {
		tag := binary.LittleEndian.Uint64(mem[qw*16:])
		nloop := int(tag & 0x7FFF)
		eop := tag&0x8000 != 0
		pre := tag>>46&1 != 0
		primField := tag >> 47 & 0x7FF
		flg := int(tag >> 58 & 3)
		nreg := int(tag >> 60 & 0xF)
		if nreg == 0 {
			nreg = 16
		}
		regs := binary.LittleEndian.Uint64(mem[qw*16+8:])
		qw++
		if pre {
			g.prim = primField
		}
		if flg != 0 { // REGLIST/IMAGE — not used by these chains for draws
			// REGLIST: nloop*nreg regs in 2-per-qw; IMAGE: nloop qws of data
			if flg == 2 || flg == 3 {
				qw += nloop
			} else {
				qw += (nloop*nreg + 1) / 2
			}
			if eop {
				return
			}
			continue
		}
		verts := 0
		for l := 0; l < nloop && qw < nqw; l++ {
			for ri := 0; ri < nreg && qw < nqw; ri++ {
				reg := regs >> (4 * ri) & 0xF
				data := binary.LittleEndian.Uint64(mem[qw*16:])
				switch reg {
				case 0x0: // PRIM
					g.prim = data & 0x7FF
				case 0x4, 0x5, 0xC, 0xD: // XYZF2/XYZ2/XYZF3/XYZ3
					if reg == 0x4 || reg == 0x5 {
						verts++
					}
				case 0xE: // A+D
					ad := mem[qw*16+8]
					switch ad {
					case 0x00:
						g.prim = data & 0x7FF
					case 0x06, 0x07:
						g.tex0[ad-0x06] = data
					case 0x3B:
						g.texa = data
					case 0x42, 0x43:
						g.alpha[ad-0x42] = data
					case 0x47, 0x48:
						g.test[ad-0x47] = data
					case 0x49:
						g.pabe = data
					}
				}
				qw++
			}
		}
		if verts > 0 {
			ctx := int(g.prim >> 9 & 1)
			k := drawKey{src, g.prim, g.test[ctx], g.alpha[ctx], g.tex0[ctx]}
			draws[k] += verts
			drawPkts[k]++
		}
		if eop {
			return
		}
	}
}

func decodePrim(p uint64) string {
	return fmt.Sprintf("type=%d ABE=%d TME=%d FGE=%d IIP=%d CTXT=%d",
		p&7, p>>6&1, p>>4&1, p>>5&1, p>>3&1, p>>9&1)
}

func decodeTest(t uint64) string {
	return fmt.Sprintf("ATE=%d ATST=%d AREF=0x%02X AFAIL=%d DATE=%d ZTE=%d ZTST=%d",
		t&1, t>>1&7, t>>4&0xFF, t>>12&3, t>>14&1, t>>16&1, t>>17&3)
}

func decodeAlpha(a uint64) string {
	return fmt.Sprintf("A=%d B=%d C=%d D=%d FIX=0x%02X", a&3, a>>2&3, a>>4&3, a>>6&3, a>>32&0xFF)
}

func decodeTex0(t uint64) string {
	return fmt.Sprintf("TBP=0x%03X TBW=%d PSM=0x%02X TW=%d TH=%d TCC=%d CBP=0x%03X",
		t&0x3FFF, t>>14&0x3F, t>>20&0x3F, t>>26&0xF, t>>30&0xF, t>>34&1, t>>37&0x3FFF)
}

func main() {
	ramF := flag.String("ram", "", "RAM image")
	flag.Parse()
	ram, err := os.ReadFile(*ramF)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var heads []uint32
	for a := uint32(0); a+16 < uint32(len(ram)); a += 16 {
		if binary.LittleEndian.Uint32(ram[a:]) == 0x1000000A &&
			binary.LittleEndian.Uint32(ram[a+8:]) == 0x01000404 {
			heads = append(heads, a)
		}
	}
	fmt.Printf("%d chain heads\n", len(heads))

	pre := merc.NewReplayer(ram)
	pre.SkipRun = true
	for _, h := range heads {
		pre.Play(h)
	}

	for _, h := range heads {
		r := merc.NewReplayer(ram)
		copy(r.V.Micro, pre.V.Micro)
		gs := &gsState{}
		r.Direct = func(qws []byte) { gs.feed(qws, "direct") }
		r.V.XGKick = func(qw uint32) {
			end := uint32(len(r.V.Data))
			gs.feed(r.V.Data[qw*16:end], fmt.Sprintf("kick@%03X", r.LastEntry))
		}
		r.Play(h)
	}

	type row struct {
		k drawKey
		v int
	}
	var rows []row
	for k, v := range draws {
		rows = append(rows, row{k, v})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].v > rows[j].v })
	for _, r := range rows {
		fmt.Printf("%-10s %6d verts %4d pkts | PRIM{%s}\n           TEST{%s}\n           ALPHA{%s}\n           TEX0{%s}\n",
			r.k.src, r.v, drawPkts[r.k], decodePrim(r.k.prim), decodeTest(r.k.test), decodeAlpha(r.k.alpha), decodeTex0(r.k.tex0))
	}
}
