// tpage decodes a texture-page object from a DGO/CGO archive to PNGs.
//
// Usage:
//
//	tpage -image DISC.iso -archive CGO/ART.CGO -entry tpage-463 \
//	      -symtab ../work/goal.txt -outdir OUT [-mips]
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	imageF := flag.String("image", "", "disc image (.iso)")
	archive := flag.String("archive", "", "archive path (e.g. CGO/ART.CGO)")
	entry := flag.String("entry", "", "tpage object name")
	symtab := flag.String("symtab", "", "GOAL symbol table dump")
	outdir := flag.String("outdir", "", "write PNGs here")
	mips := flag.Bool("mips", false, "also write mip levels beyond 0")
	flag.Parse()

	f, err := os.Open(*imageF)
	check(err)
	st, err := f.Stat()
	check(err)
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	path := *archive
	if len(path) > 0 && path[len(path)-1] != '1' {
		path += ";1"
	}
	data, err := vol.ReadFile(path)
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)

	var raw []byte
	for _, e := range d.Entries {
		if e.Name == *entry {
			raw = e.Data
			break
		}
	}
	if raw == nil {
		fmt.Fprintf(os.Stderr, "tpage: no entry %q in %s\n", *entry, d.Name)
		os.Exit(1)
	}
	obj, _, err := goalobj.Link(raw, 0, tab)
	check(err)
	pg, err := tpage.Load(obj)
	check(err)
	fmt.Printf("%s: page %q id %d, %d textures, 0x%X words of VRAM\n",
		*entry, pg.Name, pg.ID, len(pg.Textures), pg.Words)

	if *outdir == "" {
		for _, t := range pg.Textures {
			fmt.Printf("  %-28s %4dx%-4d psm 0x%02X mips %d tbp 0x%03X tbw %d cbp 0x%03X cpsm 0x%02X\n",
				t.Name, t.W, t.H, t.PSM, t.Mips, t.TBP[0], t.TBW[0], t.CBP, t.ClutPSM)
		}
		return
	}
	check(os.MkdirAll(*outdir, 0o755))
	wrote := 0
	for i := range pg.Textures {
		t := &pg.Textures[i]
		top := 1
		if *mips {
			top = t.Mips
		}
		for m := 0; m < top; m++ {
			img, err := pg.Decode(t, m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skip %s mip %d: %v\n", t.Name, m, err)
				continue
			}
			name := fmt.Sprintf("%s.png", t.Name)
			if m > 0 {
				name = fmt.Sprintf("%s.mip%d.png", t.Name, m)
			}
			out, err := os.Create(filepath.Join(*outdir, name))
			check(err)
			check(png.Encode(out, img))
			check(out.Close())
			wrote++
		}
	}
	fmt.Printf("wrote %d PNGs to %s\n", wrote, *outdir)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "tpage:", err)
		os.Exit(1)
	}
}
