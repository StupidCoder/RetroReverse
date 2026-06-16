// leveltrace drives the Game Gear machine model from boot into actual gameplay —
// pressing Start to get past the title — and snapshots the live screen plus the
// scroll/map state. It is the oracle-assisted step for locating the level map:
// rather than chase reused pointers statically, run a level and read what $D2AF
// actually points to and which bank is paged. That LOCATED the Zone 0 map in bank 4
// ($D2AF = $7B99 = file $13B99), which is then analysed statically.
//
// Limitation: the simplified machine loads and renders the level (the screenshot is
// real Green Hills graphics) and the controller read registers ($D203 reflects the
// injected D-pad), but the player physics do not advance — driving live scrolling
// exceeds the non-cycle-accurate model's fidelity. So this captures the map's
// location and a static frame, not a streaming trace.
//
// Usage: leveltrace <rom.gg> <outdir>
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"stupidcoder.com/tools/gamegear"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: leveltrace <rom.gg> <outdir>")
		os.Exit(2)
	}
	rom, err := os.ReadFile(os.Args[1])
	chk(err)
	outdir := os.Args[2]
	chk(os.MkdirAll(outdir, 0o755))

	m := gamegear.NewMachine(rom)

	word := func(a uint16) uint16 { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
	snap := func(tag string) {
		pal := gamegear.Palette(m.VDP.CRAM[:])
		full := gamegear.RenderNameTable(m.VDP.VRAM[0x3800:], m.VDP.VRAM[:], 32, 28, pal)
		gg := full.SubImage(image.Rect(48, 24, 48+160, 24+144)).(*image.Paletted)
		writePNG(filepath.Join(outdir, "lvl_"+tag+".png"), scale(gg, 3))
		fmt.Printf("%-7s scene=$%02X mode=$%02X  D2AF=$%04X  cam(D2AB/D2AD)=$%04X/$%04X  "+
			"bounds(D26D/D26F)=$%04X/$%04X\n", tag, m.Read(0xD238), m.Read(0xD240),
			word(0xD2AF), word(0xD2AB), word(0xD2AD), word(0xD26D), word(0xD26F))
	}

	frame := 0
	run := func(n int) {
		for i := 0; i < n; i++ {
			m.RunFrame()
			frame++
		}
	}

	dump := func(tag string) {
		bank := m.Read(0xD22F)
		ptr := word(0xD2AF)
		fmt.Printf("%-12s in=$%02X(D203) cam=$%04X/$%04X slot1=b%-2d D2AF=$%04X bytes:",
			tag, m.Read(0xD203), word(0xD2AB), word(0xD2AD), bank, ptr)
		for i := uint16(0); i < 16; i++ {
			fmt.Printf(" %02X", m.Read(ptr + i))
		}
		fmt.Println()
	}

	run(700) // reach the title
	// Tap Start until we enter gameplay (a non-zero right scroll bound), then STOP
	// (further Start presses would pause / leave the level).
	for round := 0; round < 8 && word(0xD26F) == 0; round++ {
		m.Pad00 = 0x7F
		run(8)
		m.Pad00 = 0xFF
		run(242)
	}
	snap(fmt.Sprintf("%04d_inlevel", frame))
	dump("inlevel")
	chk(os.WriteFile(filepath.Join(outdir, "vram.bin"), m.VDP.VRAM[:], 0o644))
	chk(os.WriteFile(filepath.Join(outdir, "cram.bin"), m.VDP.CRAM[:], 0o644))
	ram := make([]byte, 0x2000)
	for i := range ram {
		ram[i] = m.Read(uint16(0xC000 + i))
	}
	chk(os.WriteFile(filepath.Join(outdir, "ram.bin"), ram, 0o644))

	// Hold Right + Jump and watch whether the scroll position EVER moves over a long run.
	m.PadDC = 0xE7 // Right (bit3) + Button 1 (bit4) low
	fmt.Printf("\nholding Right+Jump; watching scroll $D3FF / drawn $D254:\n")
	for round := 0; round < 20; round++ {
		run(100)
		fmt.Printf("  frame %4d: D3FF=$%04X D254=$%04X cam=$%04X in(D203)=$%02X\n",
			frame, word(0xD3FF), word(0xD254), word(0xD2AB), m.Read(0xD203))
		if round == 3 { // a varied surface view, before Sonic sinks into the fill
			snap(fmt.Sprintf("%04d_surface", frame))
			chk(os.WriteFile(filepath.Join(outdir, "vram_scroll.bin"), m.VDP.VRAM[:], 0o644))
			chk(os.WriteFile(filepath.Join(outdir, "cram_scroll.bin"), m.VDP.CRAM[:], 0o644))
			// Which code writes the name table while actively scrolling?
			m.Watch(0x3800, 0x3F00)
			run(3)
			type pc2 struct {
				a uint16
				n int
			}
			var h []pc2
			for a, n := range m.WatchPCs {
				h = append(h, pc2{a, n})
			}
			sort.Slice(h, func(i, j int) bool { return h[i].n > h[j].n })
			fmt.Printf("  name-table writers while scrolling: ")
			for i := 0; i < 14 && i < len(h); i++ {
				fmt.Printf("$%04X(%d) ", h[i].a, h[i].n)
			}
			fmt.Println()
			m.WatchPCs = nil
		}
	}
	snap(fmt.Sprintf("%04d_right", frame))

	// Profile one frame: what code is the machine actually running in the level?
	m.PCHist = map[uint16]int{}
	m.Sample = true
	m.RunFrame()
	m.Sample = false
	type pc struct {
		a uint16
		n int
	}
	var hot []pc
	for a, n := range m.PCHist {
		hot = append(hot, pc{a, n})
	}
	sort.Slice(hot, func(i, j int) bool { return hot[i].n > hot[j].n })
	fmt.Printf("\nhottest PCs in one in-level frame (where the CPU spends time):\n")
	for i := 0; i < 16 && i < len(hot); i++ {
		fmt.Printf("  $%04X x%d\n", hot[i].a, hot[i].n)
	}
}

func scale(src *image.Paletted, n int) *image.Paletted {
	b := src.Bounds()
	out := image.NewPaletted(image.Rect(0, 0, b.Dx()*n, b.Dy()*n), src.Palette)
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			ci := src.ColorIndexAt(b.Min.X+x, b.Min.Y+y)
			for dy := 0; dy < n; dy++ {
				for dx := 0; dx < n; dx++ {
					out.SetColorIndex(x*n+dx, y*n+dy, ci)
				}
			}
		}
	}
	return out
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	chk(err)
	defer f.Close()
	chk(png.Encode(f, img))
}

func chk(e error) {
	if e != nil {
		panic(e)
	}
}
