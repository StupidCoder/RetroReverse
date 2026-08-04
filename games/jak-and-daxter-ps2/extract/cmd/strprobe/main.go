package main

// strprobe: parse a STR spool container (sector-0 chunk table, v2 GOAL
// object chunks), link each chunk and list its art-joint-anims.
import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/lib/iso9660"
)

func u32(b []byte, o int) uint32 {
	return uint32(b[o]) | uint32(b[o+1])<<8 | uint32(b[o+2])<<16 | uint32(b[o+3])<<24
}

func main() {
	imageF := flag.String("image", "image/Jak and Daxter - The Precursor Legacy.iso", "disc image")
	symtab := flag.String("symtab", "work/goal.txt", "GOAL symbol table")
	file := flag.String("file", "STR/NDINTRO.STR", "STR path on disc")
	flag.Parse()
	f, err := os.Open(*imageF)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)
	data, err := vol.ReadFile(*file + ";1")
	check(err)

	// sector-0 chunk table: start sectors until a zero word
	var starts []int
	for o := 0; o < 2048; o += 4 {
		v := int(u32(data, o))
		if v == 0 {
			break
		}
		starts = append(starts, v*2048)
	}
	fmt.Printf("%s: %d bytes, %d chunks at %v\n", *file, len(data), len(starts), starts)
	animType := tab.Syms["art-joint-anim"].Value
	for ci, s := range starts {
		end := len(data)
		if ci+1 < len(starts) {
			end = starts[ci+1]
		}
		obj, rep, err := goalobj.Link(data[s:end], 0, tab)
		if err != nil {
			fmt.Printf("chunk %d: link error: %v\n", ci, err)
			continue
		}
		fmt.Printf("chunk %d: %d bytes linked, %d symbol refs\n", ci, len(obj), len(rep.SymbolRef))
		for _, ap := range merc.FindAnims(obj, animType) {
			a, err := merc.DecodeJointAnim(obj, ap)
			if err != nil {
				fmt.Printf("  anim @%06x: %v\n", ap, err)
				continue
			}
			fmt.Printf("  anim @%06x: %-24s %d joints, %d frames\n", ap, a.Name, a.NumJoints, len(a.Frames))
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
