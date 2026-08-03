package main

import (
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/tpage"
	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	f, _ := os.Open(os.Args[1])
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	data, err := vol.ReadFile(os.Args[3] + ";1")
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	tab, err := goalobj.LoadSymTab(os.Args[2])
	check(err)
	for _, name := range os.Args[4:] {
		for _, e := range d.Entries {
			if e.Name != name {
				continue
			}
			obj, _, err := goalobj.Link(e.Data, 0, tab)
			check(err)
			pg, err := tpage.Load(obj)
			check(err)
			fmt.Printf("%s: %q id %d\n", name, pg.Name, pg.ID)
			for i, t := range pg.Textures {
				fmt.Printf("  [%d] %-28s %4dx%-4d psm=%02x clutpsm=%02x mips=%d tbp=%v tbw=%v cbp=%04x\n",
					i, t.Name, t.W, t.H, t.PSM, t.ClutPSM, t.Mips, t.TBP, t.TBW, t.CBP)
			}
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
