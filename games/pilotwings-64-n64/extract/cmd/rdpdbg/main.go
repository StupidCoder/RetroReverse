// rdpdbg attributes framebuffer pixels to the RDP commands that produced them.
//
// It restores a savestate, runs to a target video field, and for a set of probe
// pixels records every OnPixel event together with the command that caused it.
// It then prints, per probe, the draws that touched it in the target frame, and
// decodes those commands' words. This is the census method: localise an artefact
// to one command and one pipeline stage before changing the renderer.
//
// Usage:
//
//	rdpdbg -image ROM -loadstate FILE -shotbase 1184 -field 1200 \
//	       -px X,Y -px X,Y ... [-steps N] [-dumptmem CMDIDX:FILE]
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n64"
)

type probe struct{ x, y uint32 }

type pixelRec struct {
	cmdIdx    int
	field     int
	colorAddr uint32
	ev        n64.PixelEvent
}

type cmd struct {
	op    uint32
	words []uint64
}

func main() {
	image := flag.String("image", "", "cartridge image")
	loadState := flag.String("loadstate", "", "machine snapshot to restore")
	steps := flag.Uint64("steps", 20000000, "instruction budget")
	shotBase := flag.Int("shotbase", 0, "field number the restored state starts at")
	field := flag.Int("field", 1200, "target video field")
	var pxFlags multiFlag
	flag.Var(&pxFlags, "px", "probe pixel X,Y; repeatable")
	dumpTMem := flag.String("dumptmem", "", "CMDIDX:FILE — dump TMEM before this command")
	triCheck := flag.Bool("tricheck", false, "verify edge-coefficient conventions across all recorded triangles")
	tasks := flag.Bool("tasks", false, "decode the OSTask block for every RSP task started")
	var dumpRAM multiFlag
	flag.Var(&dumpRAM, "dumpram", "ADDR:LEN:FILE (hex addr/len) — dump RDRAM at the target field; repeatable")
	snapRAM := flag.String("snapram", "", "FILE — snapshot all of RDRAM at the first graphics task of the target field")
	flag.Parse()

	var probes []probe
	for _, s := range pxFlags {
		parts := strings.Split(s, ",")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "bad -px %q\n", s)
			os.Exit(2)
		}
		x, _ := strconv.Atoi(parts[0])
		y, _ := strconv.Atoi(parts[1])
		probes = append(probes, probe{uint32(x), uint32(y)})
	}

	rom, err := n64.Load(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	m := n64.NewMachine(rom)
	if err := m.Boot(rom, n64.DefaultBoot()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := m.LoadState(*loadState); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var (
		cmds      []cmd
		colorAddr uint32
		curField  = *shotBase
		recs      = map[probe][]pixelRec{}
		origin    uint32
	)
	for _, p := range probes {
		recs[p] = nil
	}

	dumpIdx, dumpFile := -1, ""
	if *dumpTMem != "" {
		parts := strings.SplitN(*dumpTMem, ":", 2)
		dumpIdx, _ = strconv.Atoi(parts[0])
		dumpFile = parts[1]
	}

	m.OnRDPCmd = func(mm *n64.Machine, op uint32, words []uint64) {
		if len(cmds) == dumpIdx {
			os.WriteFile(dumpFile, mm.RDPTMem(), 0o644)
			fmt.Fprintf(os.Stderr, "dumped TMEM before cmd %d to %s\n", dumpIdx, dumpFile)
		}
		w := make([]uint64, len(words))
		copy(w, words)
		cmds = append(cmds, cmd{op, w})
		if op == 0x3F { // Set_Color_Image
			colorAddr = uint32(words[0]) & 0x03FFFFFF
		}
	}
	m.OnPixel = func(x, y uint32, ev n64.PixelEvent) {
		p := probe{x, y}
		if r, ok := recs[p]; ok {
			recs[p] = append(r, pixelRec{len(cmds) - 1, curField, colorAddr, ev})
		}
	}
	m.OnDisplay = func(mm *n64.Machine) {
		curField++
		if curField == *field {
			origin = mm.Origin()
			mm.CPU.Halt("target field reached")
		}
	}
	if *tasks {
		// The OSTask block is libultra's task descriptor, DMA'd to the top of
		// DMEM before the RSP starts. Offset 0xFC0: type, flags, then pointer/
		// size pairs for boot ucode, ucode text, ucode data, dram stack, output
		// buffer, task data (the display list for a graphics task), yield data.
		m.OnRSPTask = func(mm *n64.Machine, pc uint32) {
			be := func(off int) uint32 {
				d := mm.DMEM[0xFC0+off:]
				return uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
			}
			fmt.Printf("task field=%d pc=%03X type=%d flags=%X ucode=%08X/%d data=%08X/%d dl=%08X/%d out=%08X\n",
				curField, pc, be(0), be(4), be(16), be(20), be(24), be(28), be(48), be(52), be(40))
		}
	}
	if *snapRAM != "" {
		prev := m.OnRSPTask
		m.OnRSPTask = func(mm *n64.Machine, pc uint32) {
			if prev != nil {
				prev(mm, pc)
			}
			be := func(off int) uint32 {
				d := mm.DMEM[0xFC0+off:]
				return uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
			}
			// Snapshot at a graphics task close to the target field, so the RAM
			// matches the frame that field displays. The DL pointer is printed;
			// the walker takes it from there.
			if be(0) == 1 && curField >= *field-2 && *snapRAM != "" {
				os.WriteFile(*snapRAM, mm.RDRAM, 0o644)
				fmt.Fprintf(os.Stderr, "snapped RDRAM at field %d gfx task: dl=%08X/%d ucode=%08X data=%08X -> %s\n",
					curField, be(48), be(52), be(16), be(24), *snapRAM)
				*snapRAM = ""
			}
		}
	}

	res := m.Run(*steps)
	fmt.Fprintf(os.Stderr, "run: %s\n", res)
	for _, s := range dumpRAM {
		parts := strings.SplitN(s, ":", 3)
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "bad -dumpram %q\n", s)
			continue
		}
		addr, _ := strconv.ParseUint(strings.TrimPrefix(parts[0], "0x"), 16, 32)
		length, _ := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 32)
		if int(addr+length) <= len(m.RDRAM) {
			os.WriteFile(parts[2], m.RDRAM[addr:addr+length], 0o644)
			fmt.Fprintf(os.Stderr, "dumped %d bytes at %06X to %s\n", length, addr, parts[2])
		}
	}
	fmt.Printf("field %d origin=0x%08X, %d commands recorded\n\n", *field, origin, len(cmds))

	if *triCheck {
		checkTriangles(cmds)
		return
	}

	// Report each probe's history, filtered to draws into the displayed buffer.
	culprits := map[int]bool{}
	for _, p := range probes {
		fmt.Printf("== probe (%d,%d)\n", p.x, p.y)
		rs := recs[p]
		// Only the last frame drawn into the displayed buffer matters: walk
		// backwards, stop after the first gap of >1 field.
		for _, r := range rs {
			// The VI origin is offset a row or two into the buffer; match loosely.
			if diff := int64(origin) - int64(r.colorAddr); diff < 0 || diff > 0x1000 {
				continue
			}
			e := r.ev
			verdict := "DRAWN"
			if e.ZReject {
				verdict = "zrej "
			}
			if e.AlphaReject {
				verdict = "arej "
			}
			fmt.Printf("  field %d cmd %6d fb=%06X %-22s %s rgba(%3d,%3d,%3d,%3d) tex(%3d,%3d,%3d,%3d) st(%7.2f,%7.2f) z=%d\n",
				r.field, r.cmdIdx, r.colorAddr, opName(cmds[r.cmdIdx].op), verdict,
				e.R, e.G, e.B, e.A, e.TexR, e.TexG, e.TexB, e.TexA,
				float64(e.TexS)/32, float64(e.TexT)/32, e.Z>>16)
			culprits[r.cmdIdx] = true
		}
		fmt.Println()
	}

	// Decode every culprit command.
	fmt.Println("== culprit commands")
	idxs := sortedKeys(culprits)
	for _, i := range idxs {
		fmt.Printf("cmd %6d: %s\n", i, decode(cmds, i))
	}
}

func opName(op uint32) string {
	names := map[uint32]string{
		0x08: "Triangle", 0x09: "Triangle_Z", 0x0A: "Triangle_Tex", 0x0B: "Triangle_Tex_Z",
		0x0C: "Triangle_Shade", 0x0D: "Triangle_Shade_Z",
		0x0E: "Triangle_Shade_Tex", 0x0F: "Triangle_Shade_Tex_Z",
		0x24: "Texture_Rectangle", 0x25: "Texture_Rectangle_Flip",
		0x36: "Fill_Rectangle",
	}
	if n, ok := names[op]; ok {
		return n
	}
	return fmt.Sprintf("op%02X", op)
}

func sortedKeys(m map[int]bool) []int {
	var out []int
	for k := range m {
		out = append(out, k)
	}
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// decode renders one command, with the state commands that configured it: the
// nearest preceding Set_Tile / Set_Tile_Size / Set_Other_Modes / Set_Combine.
func decode(cmds []cmd, i int) string {
	c := cmds[i]
	var b strings.Builder
	switch {
	case c.op >= 0x08 && c.op <= 0x0F:
		w := c.words
		lft := w[0] >> 55 & 1
		tileIdx := w[0] >> 48 & 7
		yl := s14(uint32(w[0] >> 32 & 0x3FFF))
		ym := s14(uint32(w[0] >> 16 & 0x3FFF))
		yh := s14(uint32(w[0] & 0x3FFF))
		xl, dxldy := int32(w[1]>>32), int32(w[1])
		xh, dxhdy := int32(w[2]>>32), int32(w[2])
		xm, dxmdy := int32(w[3]>>32), int32(w[3])
		fmt.Fprintf(&b, "%s lft=%d tile=%d\n", opName(c.op), lft, tileIdx)
		fmt.Fprintf(&b, "  yh=%d.%d ym=%d.%d yl=%d.%d (quarter-lines: %d %d %d)\n",
			yh>>2, (yh&3)*25, ym>>2, (ym&3)*25, yl>>2, (yl&3)*25, yh, ym, yl)
		fmt.Fprintf(&b, "  xh=%d+%04X/64k dxhdy=%s\n", xh>>16, uint32(xh)&0xFFFF, fx(dxhdy))
		fmt.Fprintf(&b, "  xm=%d+%04X/64k dxmdy=%s\n", xm>>16, uint32(xm)&0xFFFF, fx(dxmdy))
		fmt.Fprintf(&b, "  xl=%d+%04X/64k dxldy=%s\n", xl>>16, uint32(xl)&0xFFFF, fx(dxldy))
		next := 4
		if c.op&0x04 != 0 {
			bl := w[next : next+8]
			fmt.Fprintf(&b, "  shade base RGBA %d,%d,%d,%d dx %s,%s,%s,%s de %s,%s,%s,%s\n",
				int16(bl[0]>>48), int16(bl[0]>>32), int16(bl[0]>>16), int16(bl[0]),
				fx(pair32(bl[1], bl[3], 48)), fx(pair32(bl[1], bl[3], 32)), fx(pair32(bl[1], bl[3], 16)), fx(pair32(bl[1], bl[3], 0)),
				fx(pair32(bl[4], bl[6], 48)), fx(pair32(bl[4], bl[6], 32)), fx(pair32(bl[4], bl[6], 16)), fx(pair32(bl[4], bl[6], 0)))
			next += 8
		}
		if c.op&0x02 != 0 {
			bl := w[next : next+8]
			fmt.Fprintf(&b, "  tex base s=%s t=%s w=%s\n",
				fx(pair32(bl[0], bl[2], 48)), fx(pair32(bl[0], bl[2], 32)), fx(pair32(bl[0], bl[2], 16)))
			fmt.Fprintf(&b, "  tex dx  s=%s t=%s w=%s\n",
				fx(pair32(bl[1], bl[3], 48)), fx(pair32(bl[1], bl[3], 32)), fx(pair32(bl[1], bl[3], 16)))
			fmt.Fprintf(&b, "  tex de  s=%s t=%s w=%s\n",
				fx(pair32(bl[4], bl[6], 48)), fx(pair32(bl[4], bl[6], 32)), fx(pair32(bl[4], bl[6], 16)))
			next += 8
		}
		if c.op&0x01 != 0 {
			bl := w[next : next+2]
			fmt.Fprintf(&b, "  z base=%s dx=%s de=%s\n",
				fx(int32(bl[0]>>32)), fx(int32(bl[0])), fx(int32(bl[1]>>32)))
		}
	case c.op == 0x24 || c.op == 0x25:
		w := c.words
		fmt.Fprintf(&b, "%s tile=%d xh=%d yh=%d xl=%d yl=%d s0=%d t0=%d dsdx=%d dtdy=%d\n",
			opName(c.op), w[0]>>24&7,
			w[0]>>12&0xFFF>>2, w[0]&0xFFF>>2, w[0]>>44&0xFFF>>2, w[0]>>32&0xFFF>>2,
			int16(w[1]>>48), int16(w[1]>>32), int16(w[1]>>16), int16(w[1]))
	default:
		fmt.Fprintf(&b, "%s %016X\n", opName(c.op), c.words[0])
	}

	// The state that was live: scan backwards for the configuring commands.
	tileIdx := -1
	if c.op >= 0x08 && c.op <= 0x0F && c.op&0x02 != 0 {
		tileIdx = int(c.words[0] >> 48 & 7)
	}
	if c.op == 0x24 || c.op == 0x25 {
		tileIdx = int(c.words[0] >> 24 & 7)
	}
	seen := map[string]bool{}
	for j := i - 1; j >= 0 && len(seen) < 8; j-- {
		p := cmds[j]
		switch p.op {
		case 0x2F: // Set_Other_Modes
			if !seen["om"] {
				seen["om"] = true
				fmt.Fprintf(&b, "  [cmd %d] Set_Other_Modes %014X cycle=%d\n", j, p.words[0]&0xFFFFFFFFFFFFFF, p.words[0]>>52&3)
			}
		case 0x3C:
			if !seen["cc"] {
				seen["cc"] = true
				fmt.Fprintf(&b, "  [cmd %d] Set_Combine %014X\n", j, p.words[0]&0xFFFFFFFFFFFFFF)
			}
		case 0x35: // Set_Tile
			if tileIdx >= 0 && int(p.words[0]>>24&7) == tileIdx && !seen["tile"] {
				seen["tile"] = true
				w := p.words[0]
				fmt.Fprintf(&b, "  [cmd %d] Set_Tile %d fmt=%d size=%d line=%d tmem=%d pal=%d cmT=%d maskT=%d shiftT=%d cmS=%d maskS=%d shiftS=%d\n",
					j, tileIdx, w>>53&7, w>>51&3, w>>41&0x1FF, w>>32&0x1FF, w>>20&15,
					w>>18&3, w>>14&15, w>>10&15, w>>8&3, w>>4&15, w&15)
			}
		case 0x32: // Set_Tile_Size
			if tileIdx >= 0 && int(p.words[0]>>24&7) == tileIdx && !seen["tsz"] {
				seen["tsz"] = true
				w := p.words[0]
				fmt.Fprintf(&b, "  [cmd %d] Set_Tile_Size %d sl=%d tl=%d sh=%d th=%d (10.2)\n",
					j, tileIdx, w>>44&0xFFF, w>>32&0xFFF, w>>12&0xFFF, w&0xFFF)
			}
		case 0x33: // Load_Block
			if !seen["lb"] {
				seen["lb"] = true
				w := p.words[0]
				fmt.Fprintf(&b, "  [cmd %d] Load_Block tile=%d sl=%d tl=%d sh=%d dxt=%d\n",
					j, w>>24&7, w>>44&0xFFF, w>>32&0xFFF, w>>12&0xFFF, w&0xFFF)
			}
		case 0x34: // Load_Tile
			if !seen["lt"] {
				seen["lt"] = true
				w := p.words[0]
				fmt.Fprintf(&b, "  [cmd %d] Load_Tile tile=%d sl=%d tl=%d sh=%d th=%d (10.2)\n",
					j, w>>24&7, w>>44&0xFFF, w>>32&0xFFF, w>>12&0xFFF, w&0xFFF)
			}
		case 0x3D: // Set_Texture_Image
			if !seen["ti"] {
				seen["ti"] = true
				w := p.words[0]
				fmt.Fprintf(&b, "  [cmd %d] Set_Texture_Image fmt=%d size=%d width=%d addr=0x%06X\n",
					j, w>>53&7, w>>51&3, (w>>32&0x3FF)+1, uint32(w)&0x03FFFFFF)
			}
		}
	}
	return b.String()
}

func s14(v uint32) int32 { return int32(v<<18) >> 18 }

// checkTriangles tests two hypotheses over every triangle in the stream:
//
//  1. XH and XM are the edge x values at YH rounded down to a whole scanline
//     (YH & ~3), stepped per quarter-line by slope/4. If so, both edges meet at
//     the top vertex when evaluated at YH.
//  2. XL is the low edge's x at exactly YM (so xm evaluated at YM equals XL).
//
// The error is reported in 16.16 x units; a convention that holds should sit
// within a couple of ULPs of the microcode's own divides.
func checkTriangles(cmds []cmd) {
	tiles := map[string]int{}
	var n, apexBad, xlBad int
	var apexMax, xlMax float64
	for _, c := range cmds {
		if c.op == 0x35 {
			w := c.words[0]
			k := fmt.Sprintf("fmt=%d size=%d cmT=%d maskT=%2d cmS=%d maskS=%2d", w>>53&7, w>>51&3, w>>18&3, w>>14&15, w>>8&3, w>>4&15)
			tiles[k]++
		}
		if c.op < 0x08 || c.op > 0x0F {
			continue
		}
		w := c.words
		yl := int64(s14(uint32(w[0] >> 32 & 0x3FFF)))
		ym := int64(s14(uint32(w[0] >> 16 & 0x3FFF)))
		yh := int64(s14(uint32(w[0] & 0x3FFF)))
		xl := int64(int32(w[1] >> 32))
		xh, dxhdy := int64(int32(w[2]>>32)), int64(int32(w[2]))
		xm, dxmdy := int64(int32(w[3]>>32)), int64(int32(w[3]))
		_ = yl
		n++
		base := yh &^ 3
		// apex: xh(yh) vs xm(yh), both from the scanline boundary
		sub := yh - base
		ah := xh + dxhdy*sub/4
		am := xm + dxmdy*sub/4
		d := float64(ah-am) / 65536
		if d < 0 {
			d = -d
		}
		if d > apexMax {
			apexMax = d
		}
		if d > 0.05 {
			apexBad++
			if apexBad <= 5 {
				fmt.Printf("apex mismatch: yh=%d ym=%d yl=%d xh(yh)=%.3f xm(yh)=%.3f\n",
					yh, ym, yl, float64(ah)/65536, float64(am)/65536)
			}
		}
		// xl origin: xm evaluated at ym vs XL, under two hypotheses for where XL
		// is specified: (A) at exactly YM, (B) at YM & ~3. Only well-conditioned
		// M edges discriminate — a one-quarter-tall top edge has a huge slope and
		// extrapolating it to ym is meaningless.
		dxl := int64(int32(w[1]))
		if ym >= yl || ym-yh < 8 || dxmdy > 3<<16 || dxmdy < -3<<16 {
			continue
		}
		if ym&3 == 0 {
			continue // hypotheses coincide
		}
		subm := ym - base
		amy := xm + dxmdy*subm/4
		errA := float64(amy-xl) / 65536
		errB := float64(amy-(xl+dxl*(ym&3)/4)) / 65536
		if errA < 0 {
			errA = -errA
		}
		if errB < 0 {
			errB = -errB
		}
		if errA > xlMax {
			xlMax = errA
		}
		xlBad++ // reuse as a counter of discriminating triangles
		if xlBad <= 12 {
			fmt.Printf("xl: ym&3=%d errA(at ym)=%.4f errB(at ym&~3)=%.4f\n", ym&3, errA, errB)
		}
	}
	fmt.Printf("%d triangles: apex bad %d (max err %.4f px), xl bad %d (max err %.4f px)\n",
		n, apexBad, apexMax, xlBad, xlMax)
	fmt.Println("Set_Tile census:")
	for k, c := range tiles {
		fmt.Printf("  %6d  %s\n", c, k)
	}
}

func pair32(intWord, fracWord uint64, shift uint) int32 {
	i := int32(int16(uint16(intWord >> shift)))
	f := int32(uint16(fracWord >> shift))
	return i<<16 | f
}

func fx(v int32) string { return fmt.Sprintf("%d.%05d", v>>16, uint64(uint32(v)&0xFFFF)*100000/65536) }

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}
