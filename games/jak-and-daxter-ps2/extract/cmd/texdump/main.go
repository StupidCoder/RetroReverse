package main

// texdump: resolve a list of merc shader raw-ids through a level's remap
// table and write the referenced textures as PNGs (alpha as-is, x2 view
// alongside), with a stat line per texture.
import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	imageF := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	symtab := flag.String("symtab", "work/goal.txt", "symbol table")
	visDGO := flag.String("visdgo", "DGO/TIT.DGO", "level DGO for the remap table")
	vis := flag.String("vis", "title-vis", "vis object name")
	pages := flag.String("pages", "CGO/ART.CGO", "comma-separated archives whose tpages to load")
	out := flag.String("out", ".", "output dir")
	flag.Parse()

	f, err := os.Open(*imageF)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)

	dgoEntry := func(d *goalobj.DGO, name string) []byte {
		for _, e := range d.Entries {
			if e.Name == name && len(e.Data) >= 12 && e.Data[8] >= 4 {
				return e.Data
			}
		}
		return nil
	}
	data, err := vol.ReadFile(*visDGO + ";1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	visObj, _, err := goalobj.Link(dgoEntry(d, *vis), 0, tab)
	check(err)
	remap := merc.LoadRemapTable(visObj)

	pgs := map[int]*tpage.Page{}
	for _, arc := range strings.Split(*pages, ",") {
		data, err := vol.ReadFile(arc + ";1")
		check(err)
		d, err := goalobj.ReadDGO(data)
		check(err)
		for _, e := range d.Entries {
			if !strings.HasPrefix(e.Name, "tpage-") {
				continue
			}
			pobj, _, err := goalobj.Link(e.Data, 0, tab)
			check(err)
			pg, err := tpage.Load(pobj)
			check(err)
			if _, ok := pgs[pg.ID]; !ok {
				pgs[pg.ID] = pg
			}
		}
	}

	for _, a := range flag.Args() {
		raw64, err := strconv.ParseUint(a, 0, 32)
		check(err)
		raw := uint32(raw64)
		id := remap.Lookup(raw)
		pg := pgs[int(id>>20)]
		idx := int(id >> 8 & 0xFFF)
		if pg == nil {
			fmt.Printf("%08x -> %08x: page %d NOT LOADED\n", raw, id, id>>20)
			continue
		}
		if idx >= len(pg.Textures) || pg.Textures[idx].W == 0 {
			fmt.Printf("%08x -> %08x: page %d idx %d EMPTY\n", raw, id, id>>20, idx)
			continue
		}
		tex := &pg.Textures[idx]
		img, err := pg.Decode(tex, 0)
		check(err)
		// alpha census
		b := img.Bounds()
		total, zeros, mids, opaq := 0, 0, 0, 0
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				_, _, _, al := img.At(x, y).RGBA()
				a8 := int(al>>8) * 2
				total++
				switch {
				case a8 < 16:
					zeros++
				case a8 < 240:
					mids++
				default:
					opaq++
				}
			}
		}
		name := fmt.Sprintf("%08x-%s", raw, tex.Name)
		fmt.Printf("%08x -> %08x: page %d idx %d %-24s %dx%d zeros %4.1f%% mids %4.1f%% opaque %4.1f%%\n",
			raw, id, id>>20, idx, tex.Name, tex.W, tex.H,
			100*float64(zeros)/float64(total), 100*float64(mids)/float64(total), 100*float64(opaq)/float64(total))
		of, err := os.Create(filepath.Join(*out, name+".png"))
		check(err)
		check(png.Encode(of, img))
		of.Close()
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
