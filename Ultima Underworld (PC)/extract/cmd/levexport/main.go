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
	Level     int           `json:"level"`
	Positions []float32     `json:"positions"`
	UVs       []float32     `json:"uvs"`
	Groups    []outGroup    `json:"groups"`
	Textures  []outTexture  `json:"textures"`
	Spawn     []float32     `json:"spawn"`     // [x,y,z] interior start (Y-up), eye height above a floor
	SpawnDir  []float32     `json:"spawnDir"`  // [x,y,z] initial look direction (Y-up)
	Sprites   []outSprite   `json:"sprites"`   // billboard objects (camera-facing)
	Creatures []outCreature `json:"creatures"` // view-dependent 8-direction sprites
	SpriteTex []string      `json:"spriteTex"` // PNG data URIs, indexed by outSprite.Tex / outDir.Tex
	AddTex    []int         `json:"addTex"`    // spriteTex indices to draw ADDITIVELY (translucent bodies)
	Picks     []outPick     `json:"picks"`     // click targets (item id + world AABB)
}

// outPick is a click target: an axis-aligned box (world Y-up, Pos = centre, Size
// = full extents) tagged with the object's item id, so the viewer can raycast it
// and report which object was clicked. Doors emit two (frame + leaf, same id).
type outPick struct {
	ID   int        `json:"id"`
	Pos  [3]float32 `json:"pos"`
	Size [3]float32 `json:"size"`
}

// outSprite is one camera-facing billboard: its base sits at Pos (Y-up), it is
// W×H world units, and Tex indexes SpriteTex.
type outSprite struct {
	Pos [3]float32 `json:"pos"`
	W   float32    `json:"w"`
	H   float32    `json:"h"`
	Tex int        `json:"tex"`
}

// outCreature is a directional (Doom-style) billboard that plays its idle
// animation. Views holds eight compass views (view 0 the creature's back, view 4
// its front), each a frame cycle to loop at Fps. The viewer picks the view each
// render from the camera-to-creature bearing and Heading (matching the game's
// emit path 2DFE:0221, remap table DGROUP:05AC) and advances the cycle by time.
// Base sits at Pos (Y-up).
type outCreature struct {
	Pos     [3]float32 `json:"pos"`
	Heading int        `json:"heading"` // 0-7, 45° steps
	Views   [][]outDir `json:"views"`   // eight view cycles (fewer → single-view creature)
	Fps     float32    `json:"fps"`     // idle-cycle playback rate
	// Translucent marks the ethereal creatures (ghost/wisp/fire/shadow) whose
	// sprites use the reserved palette index 252 — the game BRIGHTENS the
	// background there (additively) instead of drawing a colour. For them each
	// frame is split by palette index: index 252 → the additive layer (outDir.Add),
	// every other opaque index → the normal layer (outDir.Tex, e.g. the ghost's
	// black eyes, fire's flames, the shadow's body). The viewer draws Add with
	// additive blending over the normal Tex.
	Translucent bool `json:"translucent,omitempty"`
}

// outDir is one animation frame of a creature: a SpriteTex index and its world
// size (frames differ in size, so the viewer rescales as the cycle plays). Add
// indexes SpriteTex for a translucent creature's additive (index-252) layer, or
// -1 when the frame has no additive pixels.
type outDir struct {
	Tex int     `json:"tex"`
	W   float32 `json:"w"`
	H   float32 `json:"h"`
	Add int     `json:"add"` // additive layer SpriteTex index, -1 = none
}

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	level := flag.Int("level", 0, "level index (0-7)")
	palN := flag.Int("pal", 0, "PALS.DAT palette index")
	ceil := flag.Bool("ceilings", true, "include ceiling faces (enclosed dungeon)")
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

	// An interior start point so the (now ceiling-enclosed) level opens with the
	// fly-camera already inside it. Y-up, matching the position layout below.
	spawn, spawnDir := spawnPoint(grid, *level)

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

	// Emit triangles (two per quad), Y-up. The game world is (X=east, Y=north,
	// Z=up); three.js is Y-up, so map to (tileX, height, -tileY). Negating tileY
	// (rather than a plain axis swap) keeps the handedness right — a bare
	// (tileX, height, tileY) swap reflects the level and renders it mirrored.
	o := &outMesh{Level: *level, Spawn: spawn, SpawnDir: spawnDir}
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
			o.Positions = append(o.Positions, p[0], p[2], -p[1])
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

	// Bake the 3D-class objects (doors, pillars, bridges...) into the mesh as
	// extra groups; their materials append after the level's.
	exeBytes, err := os.ReadFile(filepath.Join(*game, "UW.EXE"))
	must(err)
	comObj, err := lev.ParseComObj(data("COMOBJ.DAT"))
	must(err)
	doorGR, err := tex.ParseGR(data("DOORS.GR"))
	must(err)
	tmobjGR, err := tex.ParseGR(data("TMOBJ.GR"))
	must(err)
	appendObjects(o, grid, block, exeBytes, comObj, tm, doorGR, tmobjGR, wallTR, pal)

	// Billboard sprites (class-0/1 objects: items, creatures) from OBJECTS.GR.
	objGR, err := tex.ParseGR(data("OBJECTS.GR"))
	must(err)
	appendBillboards(o, grid, block, comObj, objGR, data("ALLPALS.DAT"), pal, *game)

	buf, err := json.Marshal(o)
	must(err)
	must(os.WriteFile(*out, buf, 0o644))
	fmt.Printf("wrote %s: %d triangles, %d materials, %d KB\n",
		*out, len(o.Positions)/9, len(mats), len(buf)/1024)
}

// spawnPoint returns an interior start position and initial look direction
// (Y-up: x, height, -y — the mirror-safe mapping used above). Level 1 (index 0)
// starts where the game drops a new character: the ceremonial-door room, facing
// NORTH into the level (the big door is behind you). That room and orientation
// are derived from the level geometry and the described start; the exact stored
// sub-tile coordinate isn't read from the game's player state. Other levels have
// no fixed entry, so they start at the open tile nearest the map centre.
func spawnPoint(g *lev.Grid, level int) (pos, dir []float32) {
	if level == 0 {
		tx, ty := 31.5, 2.5 // in the start room, a couple tiles in front of the door
		floorZ := float32(g.At(31, 2).Height) * levgeo.HeightScale
		pos = []float32{float32(tx), floorZ + 0.55, -float32(ty)}
		dir = []float32{0, 0, -1} // +tileY (north) maps to -Z
		return
	}
	cx, cy := g.W/2, g.H/2
	bx, by, best := cx, cy, 1<<30
	found := false
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if g.At(x, y).Type != lev.TileOpen {
				continue
			}
			d := (x-cx)*(x-cx) + (y-cy)*(y-cy)
			if d < best {
				best, bx, by, found = d, x, y, true
			}
		}
	}
	dir = []float32{1, 0, 0}
	if !found {
		pos = []float32{float32(cx) + 0.5, 1, -(float32(cy) + 0.5)}
		return
	}
	floorZ := float32(g.At(bx, by).Height) * levgeo.HeightScale
	pos = []float32{float32(bx) + 0.5, floorZ + 0.6, -(float32(by) + 0.5)}
	return
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
