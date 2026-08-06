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
	flag.Parse()

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

	if *pics {
		dumpPics(o, ls)
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
