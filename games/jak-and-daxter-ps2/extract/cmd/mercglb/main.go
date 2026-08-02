package main

// mercglb exports an art group's merc-ctrl models as a GLB scene: one node
// per merc-ctrl, one primitive per effect (extract/merc.BuildPrim).
import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/iso9660"
)

type hexList []uint32

func (h *hexList) String() string { return fmt.Sprint(*h) }
func (h *hexList) Set(s string) error {
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 32)
	*h = append(*h, uint32(v))
	return err
}

func main() {
	image := flag.String("image", "", "disc image")
	archive := flag.String("archive", "", "archive path")
	entry := flag.String("entry", "", "art group object name")
	symtab := flag.String("symtab", "", "GOAL symbol table dump")
	out := flag.String("out", "", "output .glb")
	var ctrls hexList
	flag.Var(&ctrls, "ctrl", "merc-ctrl basic offset (hex); repeatable")
	flag.Parse()

	f, err := os.Open(*image)
	check(err)
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	check(err)
	data, err := vol.ReadFile(*archive + ";1")
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
		fmt.Fprintf(os.Stderr, "mercglb: no entry %q\n", *entry)
		os.Exit(1)
	}
	obj, _, err := goalobj.Link(raw, 0, tab)
	check(err)

	engData, err := vol.ReadFile("CGO/ENGINE.CGO;1")
	check(err)
	eng, err := goalobj.ReadDGO(engData)
	check(err)
	var raws [][]byte
	for _, e := range eng.Entries {
		raws = append(raws, e.Data)
	}
	micro := merc.FindMicro(raws)
	if micro == nil {
		fmt.Fprintln(os.Stderr, "merc microcode not found in ENGINE.CGO")
		os.Exit(1)
	}
	scene := glb.NewScene()
	gold := [4]float32{0.83, 0.68, 0.28, 1}
	for ci, off := range ctrls {
		c, err := merc.Parse(obj, off)
		check(err)
		node := scene.AddNode(fmt.Sprintf("%s-%d", *entry, ci), -1,
			[3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
		var prims []glb.Prim
		for i := range c.Effects {
			p := merc.StripPrim(&c.Effects[i], gold)
			if len(p.Tris) > 0 {
				prims = append(prims, p)
			}
		}
		check(scene.AddMesh(node, fmt.Sprintf("%s-%d", *entry, ci), prims))
	}
	check(scene.Write(*out, *entry))
	fmt.Printf("wrote %s\n", *out)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercglb:", err)
		os.Exit(1)
	}
}
