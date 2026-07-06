// Command levexport writes a level's static geometry — floors, walls and
// ceilings with real textures — as a self-contained JSON the Studio's three.js
// viewer loads. The mesh is grouped by material; each material carries its
// W64.TR/F32.TR texture as a base64 PNG data URI, so the JSON needs no side
// files.
//
// Usage: levexport [-game ../game] [-level 0] [-pal 0] -o out.json
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"ultimaunderworld/extract/lev"
	"ultimaunderworld/extract/levgeo"
	"ultimaunderworld/extract/tex"
)

type outTexture struct {
	Wall bool   `json:"wall"`
	Num  int    `json:"num"`
	PNG  string `json:"png"` // data:image/png;base64,...
}

type outGroup struct {
	Start    int `json:"start"`
	Count    int `json:"count"`
	Material int `json:"material"`
}

type outMesh struct {
	Level     int          `json:"level"`
	Positions []float32    `json:"positions"`
	UVs       []float32    `json:"uvs"`
	Groups    []outGroup   `json:"groups"`
	Textures  []outTexture `json:"textures"`
}

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	level := flag.Int("level", 0, "level index (0-7)")
	palN := flag.Int("pal", 0, "PALS.DAT palette index")
	ceil := flag.Bool("ceilings", false, "include ceiling faces (enclosed dungeon)")
	out := flag.String("o", "", "output JSON path")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "levexport: -o is required")
		os.Exit(1)
	}
	data := func(name ...string) []byte {
		b, err := os.ReadFile(filepath.Join(append([]string{*game, "DATA"}, name...)...))
		if err != nil {
			fmt.Fprintln(os.Stderr, "levexport:", err)
			os.Exit(1)
		}
		return b
	}

	ark, err := lev.ParseArk(data("LEV.ARK"))
	must(err)
	block, err := ark.Block(*level)
	must(err)
	grid, err := lev.DecodeGrid(block)
	must(err)
	tm, err := ark.TexMapForLevel(*level)
	must(err)

	pal, err := tex.LoadPalette(data("PALS.DAT"), *palN)
	must(err)
	wallTR, err := tex.ParseTR(data("W64.TR"))
	must(err)
	floorTR, err := tex.ParseTR(data("F32.TR"))
	must(err)

	mesh := levgeo.Build(grid, tm, *ceil)

	// Assign a material index per (wall, texture-number) used, and sort the quads
	// by material so each material's triangles form one contiguous group.
	type matKey struct {
		wall bool
		num  uint16
	}
	matIndex := map[matKey]int{}
	var mats []matKey
	for _, q := range mesh.Quads {
		k := matKey{q.Wall, q.Tex}
		if _, ok := matIndex[k]; !ok {
			matIndex[k] = len(mats)
			mats = append(mats, k)
		}
	}
	sort.SliceStable(mesh.Quads, func(i, j int) bool {
		return matIndex[matKey{mesh.Quads[i].Wall, mesh.Quads[i].Tex}] <
			matIndex[matKey{mesh.Quads[j].Wall, mesh.Quads[j].Tex}]
	})

	// Emit triangles (two per quad), Y-up: position = (tileX, height, tileY).
	o := &outMesh{Level: *level}
	tri := [6]int{0, 1, 2, 0, 2, 3}
	groupStart := map[int]int{}
	groupCount := map[int]int{}
	for _, q := range mesh.Quads {
		mat := matIndex[matKey{q.Wall, q.Tex}]
		if _, ok := groupStart[mat]; !ok {
			groupStart[mat] = len(o.Positions) / 3
		}
		for _, ci := range tri {
			p := q.P[ci]
			o.Positions = append(o.Positions, p[0], p[2], p[1])
			o.UVs = append(o.UVs, q.UV[ci][0], q.UV[ci][1])
			groupCount[mat]++
		}
	}
	for mat := range mats {
		o.Groups = append(o.Groups, outGroup{Start: groupStart[mat], Count: groupCount[mat], Material: mat})
	}
	sort.Slice(o.Groups, func(i, j int) bool { return o.Groups[i].Start < o.Groups[j].Start })

	// Decode each material's texture to a PNG data URI.
	for _, k := range mats {
		var im *image.RGBA
		if k.wall {
			im, err = wallTR.Image(int(k.num)%wallTR.Count(), pal)
		} else {
			im, err = floorTR.Image(int(k.num)%floorTR.Count(), pal)
		}
		must(err)
		o.Textures = append(o.Textures, outTexture{Wall: k.wall, Num: int(k.num), PNG: toDataURI(im)})
	}

	buf, err := json.Marshal(o)
	must(err)
	must(os.WriteFile(*out, buf, 0o644))
	fmt.Printf("wrote %s: %d triangles, %d materials, %d KB\n",
		*out, len(o.Positions)/9, len(mats), len(buf)/1024)
}

func toDataURI(im *image.RGBA) string {
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "levexport:", err)
		os.Exit(1)
	}
}
