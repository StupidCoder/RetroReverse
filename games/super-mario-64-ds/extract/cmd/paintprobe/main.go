// paintprobe asks the game how big a painting is, by running the painting
// actor's own create+init and reading the mesh it builds.
//
// Actor 307 (overlay 80) is every framed painting in the castle. Its init
// (`$02126CA0`) reads the spawn parameter word at obj+$8 and:
//
//	param & $F        -> width  = (n+1) * 100.0 world units
//	(param >> 4) & $F -> height = (n+1) * 100.0
//	(param >> 8) & $1F-> which picture ($02125630 resolves the texture)
//	(param >> 13) & 3 -> the behaviour mode; >= 2 collapses the grid to 2x2
//
// It allocates `rows*cols` records of $18 bytes at obj+$1A0 (the grid size
// comes from a per-width table at $02127714) and hands off to the mode's own
// method. Reading that buffer back after a real init is the measurement: no
// guess about extents, orientation or subdivision.
//
//	paintprobe [-rom img] [-extracted dir] [-par 0xNNNN]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

const (
	paintActor = 307
	paintCfg   = 80
	offParam   = 0x08  // the spawn parameter word the base ctor copies in
	offVerts   = 0x1A0 // the allocated vertex/grid buffer
	offRows    = 0x1BA
	offCols    = 0x1BB
	offCount   = 0x1B8 // obj+$100 + $B8
	recSize    = 0x18
)

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	pars := flag.String("par", "", "comma-separated par1 values (hex ok with 0x); default = every placed one")
	pics := flag.Bool("pics", false, "dump the per-picture records overlay 80's static init builds")
	dump := flag.Int("dump", 0, "print this many vertex records")
	place := flag.Bool("place", false, "list every placement of the actor with its decoded parameter word")
	modes := flag.Bool("modes", false, "dump the four behaviour-mode records the init dispatches through")
	tex := flag.Bool("tex", false, "for every .bmd under -texdir: the material's TEXIMAGE_PARAM, the texture format, and how many texels the decode makes transparent")
	texDir := flag.String("texdir", "data/picture", "subtree of the extracted filesystem -tex walks")
	grid := flag.String("grid", "", "run this par1 (hex ok) and dump the mesh grid the actor builds")
	ticks := flag.Int("ticks", 0, "step the -grid actor this many ticks, dumping the grid each tick")
	flag.Parse()

	if *tex {
		dumpTextures(*ext, *texDir)
		return
	}

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := o.InitEngine(); err != nil {
		sm64ds.Die(err)
	}
	if err := o.LoadConfig(paintCfg); err != nil {
		sm64ds.Die(err)
	}

	if *grid != "" {
		var v int
		if strings.HasPrefix(*grid, "0x") {
			fmt.Sscanf((*grid)[2:], "%x", &v)
		} else {
			fmt.Sscanf(*grid, "%d", &v)
		}
		dumpGrid(o, v, *ticks)
		return
	}
	if *pics {
		dumpPics(o, ls)
		return
	}
	if *place {
		listPlacements(ls)
		return
	}
	if *modes {
		dumpModes(o)
		return
	}
	var list [][3]int
	if *pars != "" {
		for _, t := range strings.Split(*pars, ",") {
			var v int
			t = strings.TrimSpace(t)
			if strings.HasPrefix(t, "0x") {
				fmt.Sscanf(t[2:], "%x", &v)
			} else {
				fmt.Sscanf(t, "%d", &v)
			}
			list = append(list, [3]int{v, 0, 0})
		}
	} else {
		seen := map[[3]int]bool{}
		for id := 0; id < sm64ds.NumLevels; id++ {
			lv, err := ls.Level(id)
			if err != nil {
				continue
			}
			for _, ob := range lv.Objects {
				if ob.Actor == paintActor && !seen[ob.Params] {
					seen[ob.Params] = true
					list = append(list, ob.Params)
				}
			}
		}
	}

	fmt.Printf("%-8s %-5s %-5s %-4s %-5s %-6s  %s\n",
		"par1", "w", "h", "mode", "pic", "grid", "mesh bounds from the game's own buffer (world units)")
	for _, par := range list {
		if err := o.LoadConfig(paintCfg); err != nil {
			sm64ds.Die(err)
		}
		run := o.RunActor(paintActor, paintCfg, par)
		if run.Obj == 0 {
			fmt.Printf("%04X     refused\n", par[0])
			continue
		}
		p := uint32(par[0]) | uint32(par[1])<<16
		w := ((p & 0xF) + 1) * 100
		h := ((p >> 4 & 0xF) + 1) * 100
		mode := p >> 13 & 3
		pic := p >> 8 & 0x1F
		obj := run.Obj
		if os.Getenv("PAINTDBG") != "" {
			// Where did the parameter actually land on the object?
			b := o.ReadBytes(obj, 0x1C0)
			var at []string
			for i := 0; i+2 <= len(b); i += 2 {
				if int(b[i])|int(b[i+1])<<8 == par[0] {
					at = append(at, fmt.Sprintf("+%X", i))
				}
			}
			fmt.Printf("      obj %08X  par1 %04X found at %v  word@8=%08X\n", obj, par[0], at, o.R32(obj+8))
		}
		rows := o.ReadBytes(obj+offRows, 1)[0]
		cols := o.ReadBytes(obj+offCols, 1)[0]
		n := o.R16(obj + offCount)
		buf := o.R32(obj + offVerts)
		fmt.Printf("%04X     %-5d %-5d %-4d %-5d %dx%d=%-3d %s\n",
			par[0], w, h, mode, pic, rows, cols, n, bounds(o, buf, int(n), *dump))
	}
}

// dumpTextures reports what each picture's own .bmd says about its texture.
//
// The painting's draw does not use the material record — it rebuilds
// TEXIMAGE_PARAM at $02126290 out of the texture resource's authored param,
// carrying the format, the two size fields and bit 29 (colour index 0 is
// transparent) and nothing else. So the question "does this picture have a
// cut-out?" is a question about bit 29 and the format, and this prints both
// beside the count of texels the decode actually zeroes.
func dumpTextures(ext, sub string) {
	var paths []string
	filepath.Walk(filepath.Join(ext, "files", filepath.FromSlash(sub)), func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.HasSuffix(p, ".bmd") {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	fmt.Printf("%-16s %-10s %-4s %-6s %-5s %-6s %s\n",
		"model", "texparam", "fmt", "col0tr", "alpha", "size", "transparent texels")
	for _, p := range paths {
		m, err := sm64ds.LoadBMD(p)
		if err != nil {
			fmt.Printf("%-16s %v\n", filepath.Base(p), err)
			continue
		}
		for _, mat := range m.Mats {
			t, ok := m.Texs[mat.Texture]
			if !ok {
				continue
			}
			zero := 0
			if t.Img != nil {
				for i := 3; i < len(t.Img.Pix); i += 4 {
					if t.Img.Pix[i] == 0 {
						zero++
					}
				}
			}
			fmt.Printf("%-16s %08X   %-4d %-6t %-5d %dx%-3d %d\n",
				m.Name, mat.TexParam, t.Format, mat.TexParam&(1<<29) != 0,
				mat.Alpha, t.Width, t.Height, zero)
		}
	}
}

// dumpGrid runs one painting for real and reads back the mesh its actor builds,
// then steps it. The grid records are $18 bytes: +0/+4/+8 an fx20.12 position,
// +$C the vertex's own phase and +$10 the packed NORMAL the draw emits.
func dumpGrid(o *sm64ds.Oracle, par1, ticks int) {
	run := o.RunActorBanked(paintActor, paintCfg, [3]int{par1, 0, 0}, func(extra int) error {
		if extra < 0 {
			return o.LoadConfig(paintCfg)
		}
		return o.LoadConfigMulti([]int{paintCfg, extra})
	})
	if run.Obj == 0 {
		fmt.Println("refused:", run.Notes)
		return
	}
	obj := run.Obj
	fmt.Printf("obj=%08X  rows=%d cols=%d count=%d buf=%08X  phase=%d ring=%d  notes=%v\n",
		obj, o.ReadBytes(obj+offRows, 1)[0], o.ReadBytes(obj+offCols, 1)[0],
		o.R16(obj+offCount), o.R32(obj+offVerts),
		int16(o.R16(obj+0x1B4)), o.R16(obj+0x1B6), run.Notes)
	for t := 0; t <= ticks; t++ {
		n, buf := int(o.R16(obj+offCount)), o.R32(obj+offVerts)
		if buf == 0 || n == 0 {
			fmt.Println("  (no buffer)")
			return
		}
		lo, hi := 1<<30, -(1 << 30)
		var zs []string
		for i := 0; i < n; i++ {
			z := int(int32(o.R32(buf + uint32(i*recSize+8))))
			if z < lo {
				lo = z
			}
			if z > hi {
				hi = z
			}
			if i < 8 {
				zs = append(zs, fmt.Sprintf("%7.2f", float64(z)/4096))
			}
		}
		fmt.Printf("  tick %-3d phase=%-7d ring=%-4d  z %7.2f..%7.2f   first: %s\n",
			t, int16(o.R16(obj+0x1B4)), o.R16(obj+0x1B6), f(int32(lo)), f(int32(hi)), strings.Join(zs, " "))
		o.StepActor(run)
	}
}

// dumpModes reads the mode table the init indexes with (par1 >> 13) & 3.
//
//	$02126E3C  r0 = param >> 13 & 3
//	$02126E54  MLA r0, r0, #$18, $02128628   -> obj+$1A4 = &modeRec[mode]
//	$02126E5C  the member-function pair at +0/+4 is then called on the object
//
// The table is in overlay 80's BSS, so it only exists after the overlay's
// static initialisers have run — which LoadConfig does for real.
const modeTable = 0x02128628

func dumpModes(o *sm64ds.Oracle) {
	fmt.Printf("%-5s %-10s %s\n", "mode", "record", "$18 bytes")
	for m := 0; m < 4; m++ {
		rec := uint32(modeTable + m*recSize)
		fmt.Printf("%-5d %08X  ", m, rec)
		for k := 0; k < recSize; k += 4 {
			fmt.Printf("%08X ", o.R32(rec+uint32(k)))
		}
		fmt.Println()
	}
}

// listPlacements walks every level's object table and prints each actor-307
// record with its parameter word split into the fields the init reads. The
// point is to see WHICH placement carries which mode: the mode nibble is the
// only thing distinguishing the framed wall pictures from whatever else the
// same actor is asked to be.
func listPlacements(ls *sm64ds.LevelSet) {
	fmt.Printf("%-20s %-6s %-5s %-5s %-4s %-4s %7s %7s %7s  %s\n",
		"level", "par1", "w", "h", "mode", "pic", "rotX", "rotY", "rotZ", "position (world units)")
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		for _, ob := range lv.Objects {
			if ob.Actor != paintActor {
				continue
			}
			p := uint32(ob.Params[0])
			fmt.Printf("%-20s %04X   %-5d %-5d %-4d %-4d %7.2f %7.2f %7.2f  %8.1f %8.1f %8.1f\n",
				stem, p, (p&0xF+1)*100, (p>>4&0xF+1)*100, p>>13&3, p>>8&0x1F,
				ob.RotX, ob.RotY, ob.RotZ, ob.X, ob.Y, ob.Z)
		}
	}
}

// bounds reads the grid records back and reports the extent they span. A record
// is $18 bytes; the first three words are read as an fx20.12 position.
func bounds(o *sm64ds.Oracle, buf uint32, n, dump int) string {
	if buf == 0 || n == 0 {
		return "(no buffer)"
	}
	lo := [3]int32{1 << 30, 1 << 30, 1 << 30}
	hi := [3]int32{-1 << 30, -1 << 30, -1 << 30}
	nz := 0
	for i := 0; i < n; i++ {
		var v [3]int32
		for k := 0; k < 3; k++ {
			v[k] = int32(o.R32(buf + uint32(i*recSize+k*4)))
		}
		if v[0]|v[1]|v[2] != 0 {
			nz++
		}
		for k := 0; k < 3; k++ {
			if v[k] < lo[k] {
				lo[k] = v[k]
			}
			if v[k] > hi[k] {
				hi[k] = v[k]
			}
		}
		if i < dump {
			fmt.Printf("      rec %2d  %8.2f %8.2f %8.2f\n", i, f(v[0]), f(v[1]), f(v[2]))
		}
	}
	if nz == 0 {
		return "all-zero (init does not fill it)"
	}
	return fmt.Sprintf("x %.0f..%.0f  y %.0f..%.0f  z %.0f..%.0f  (%d/%d non-zero)",
		f(lo[0]), f(hi[0]), f(lo[1]), f(hi[1]), f(lo[2]), f(hi[2]), nz, n)
}

func f(v int32) float64 { return float64(v) / 4096 }

var _ = os.Stdout

// dumpPics reads the per-picture table the painting code indexes with
// (par1 >> 8) & $1F. $0212775C is 19 pointers to 8-byte records, and both live
// in overlay 80's BSS — zero in the file image, built by the overlay's static
// initialiser, which the oracle runs for real. Reading them back after
// LoadConfig(80) is the only way to see them.
func dumpPics(o *sm64ds.Oracle, ls *sm64ds.LevelSet) {
	const picTable = 0x0212775C
	fmt.Printf("%-4s %-10s %s\n", "pic", "record", "8 bytes  ->  resolved")
	for i := 0; i < 19; i++ {
		rec := o.R32(picTable + uint32(i)*4)
		if rec == 0 {
			fmt.Printf("%-4d %-10s\n", i, "(null)")
			continue
		}
		a := o.R32(rec)
		b := o.R32(rec + 4)
		fmt.Printf("%-4d %08X   %08X %08X   file[%d]=%q  file[%d]=%q\n",
			i, rec, a, b, a&0xFFFF, ls.InternalName(int(a&0xFFFF)), b&0xFFFF, ls.InternalName(int(b&0xFFFF)))
	}
}
