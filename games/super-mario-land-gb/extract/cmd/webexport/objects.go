// objects.go is the objects stage: each known object/enemy type's metasprite
// animation poses (from its script timeline, level.TypeTimeline; fallback: the
// one base metasprite) become a sprite2d object — a horizontal strip of
// 40x40 cells with the metasprite cursor origin at (8,24) — plus Mario's idle
// sprite. The tile pixels come from a short oracle run per world (OBJ tiles +
// sprite palette); the frame ids and metasprite layouts are decoded from ROM
// (level.TypeFrames / level.DecodeMetasprite).
//
// A type is ONE asset shared by every world whose rendered art is identical;
// worlds whose OBJ tiles draw it differently get their own world-named
// variant. Types placed in levels but without decodable art share a marker
// asset so their placements still have an identity.
package main

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sort"

	"retroreverse.com/games/super-mario-land-gb/extract/level"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/gameboy"
)

// Object-icon cell geometry: each pose is composited into one objCell-square
// cell. objOrigin is the cell pixel where the metasprite's cursor origin (0,0)
// sits, so a placement lines up with the object's engine position. A 40x40
// cell with the origin at (8,24) fits every base metasprite (tiles span
// DX -8..+32, DY -24..+16) without clipping.
const objCell = 40

var objOrigin = [2]int{8, 24}

// marioFrame is Mario's idle standing sprite: a 2x2 of OBJ tiles, each
// {tile, dx, dy} from the sprite top-left (read from the running game's OAM
// at spawn).
var marioFrame = []level.Sprite{
	{Tile: 0x00}, {Tile: 0x01, DX: 8}, {Tile: 0x10, DY: 8}, {Tile: 0x11, DX: 8, DY: 8},
}

var worldNames = []string{"", "World 1", "World 2", "World 3", "World 4"}

// exportObjects writes the object assets and returns the "w<world>/<type>"
// (and "w<world>/mario", "u/<type>") -> asset map placements resolve through.
func exportObjects(ctx *cli.Context, rom []byte) (map[string]objRef, error) {
	b := ctx.Builder
	refs := map[string]objRef{}

	// ---- pass 1: render every (world, type) strip, keeping them in memory ----
	type record struct {
		world     int
		key       string // refs key
		typ       int    // -1 = Mario
		name      string
		img       *image.NRGBA
		frames    int
		durations []int
	}
	var recs []record

	typeFrame := level.TypeFrames(rom)
	var types []int
	for t := range typeFrame {
		types = append(types, int(t))
	}
	sort.Ints(types)

	// A (world, type) pair is only rendered when the type is PLACED in that
	// world: compositing a type with another world's OBJ tiles produces
	// garbage art (the tiles belong to whatever that world loads there).
	placed := map[int]map[int]bool{}
	for world := 1; world <= 4; world++ {
		placed[world] = map[int]bool{}
		for lv := 1; lv <= 3; lv++ {
			for _, o := range level.DecodeObjectsByID(rom, byte(world<<4|lv)) {
				placed[world][int(o.Type)] = true
			}
		}
	}

	for world := 1; world <= 4; world++ {
		vram, obp0 := worldSpriteData(rom, byte(world))
		stamp := func(img *image.NRGBA, s level.Sprite, ox, oy int) {
			tl := gameboy.DecodeTile(vram[int(s.Tile)*16:])
			for py := 0; py < 8; py++ {
				for px := 0; px < 8; px++ {
					sx, sy := px, py
					if s.XFlip {
						sx = 7 - px
					}
					if s.YFlip {
						sy = 7 - py
					}
					if v := tl[sy][sx]; v != 0 { // OBJ colour 0 = transparent
						g := []uint8{0xff, 0xaa, 0x55, 0x00}[(obp0>>(2*v))&3]
						img.Set(ox+s.DX+px, oy+s.DY+py, color.NRGBA{g, g, g, 0xff})
					}
				}
			}
		}
		pose := func(img *image.NRGBA, frame byte, cellX int) {
			for _, s := range level.DecodeMetasprite(rom, int(frame)) {
				stamp(img, s, cellX*objCell+objOrigin[0], objOrigin[1])
			}
		}

		for _, t := range types {
			if !placed[world][t] {
				continue
			}
			frames, durations := 1, []int(nil)
			tl := level.TypeTimeline(rom, byte(t))
			if tl != nil {
				frames = len(tl)
				for _, st := range tl {
					durations = append(durations, st.Frames)
				}
			}
			img := image.NewNRGBA(image.Rect(0, 0, objCell*frames, objCell))
			if tl != nil {
				for f, st := range tl {
					pose(img, st.Frame, f)
				}
			} else {
				pose(img, typeFrame[byte(t)], 0)
			}
			recs = append(recs, record{
				world: world, key: fmt.Sprintf("w%d/%d", world, t), typ: t,
				name: fmt.Sprintf("Object $%02X", t), img: img, frames: frames, durations: durations,
			})
		}

		mario := image.NewNRGBA(image.Rect(0, 0, objCell, objCell))
		for _, s := range marioFrame {
			stamp(mario, s, objOrigin[0], objOrigin[1])
		}
		recs = append(recs, record{
			world: world, key: fmt.Sprintf("w%d/mario", world), typ: -1,
			name: "Mario", img: mario, frames: 1,
		})
		ctx.Progress("objects", world, 4, fmt.Sprintf("world %d: %d types + mario rendered", world, len(types)))
	}

	// ---- pass 2: collapse identical art, one asset per variant ----------------
	sig := func(r record) string {
		h := sha256.New()
		h.Write(r.img.Pix)
		fmt.Fprintf(h, "|%v", r.durations)
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	type variant struct {
		rec    record
		worlds []int
		keys   []string
	}
	variants := map[int][]*variant{} // type -> distinct-art variants, world order
	for _, r := range recs {
		vs := variants[r.typ]
		s := sig(r)
		found := false
		for _, v := range vs {
			if sig(v.rec) == s {
				v.worlds = append(v.worlds, r.world)
				v.keys = append(v.keys, r.key)
				found = true
				break
			}
		}
		if !found {
			variants[r.typ] = append(vs, &variant{rec: r, worlds: []int{r.world}, keys: []string{r.key}})
		}
	}

	emit := func(t int) error {
		vs := variants[t]
		for _, v := range vs {
			r := v.rec
			name := r.name
			if len(vs) > 1 && len(v.worlds) == 1 {
				name = fmt.Sprintf("%s (%s)", name, worldNames[v.worlds[0]])
			}
			id := slugify(name)
			f, err := b.CreateFile("objects", id+".png")
			if err != nil {
				return err
			}
			err = png.Encode(f, r.img)
			f.Close()
			if err != nil {
				return err
			}
			anim := schema.Animation{ID: "main", Frames: r.frames, Loop: "loop"}
			if r.frames == 1 {
				anim.Loop = "hold"
			}
			if r.frames > 1 {
				anim.Durations = r.durations
			}
			worlds := make([]string, len(v.worlds))
			for i, w := range v.worlds {
				worlds[i] = worldNames[w]
			}
			group := "Objects"
			if t == -1 {
				group = "Player"
			}
			doc := &schema.Object{
				Type: schema.ObjectSprite2D,
				Name: name,
				Atlas: &schema.SpriteAtlas{
					File: id + ".png", CellW: objCell, CellH: objCell,
					Anchor: []int{objOrigin[0], objOrigin[1]},
				},
				Animations: []schema.Animation{anim},
				Props:      map[string]any{"worlds": worlds},
			}
			if t >= 0 {
				doc.Props["type"] = fmt.Sprintf("0x%02X", t)
			}
			b.AddObject(schema.Asset{ID: id, Name: name, Group: group}, doc)
			for _, k := range v.keys {
				refs[k] = objRef{asset: id}
			}
		}
		return nil
	}
	if err := emit(-1); err != nil { // Mario first: the Player group leads the list
		return nil, err
	}
	for _, t := range types {
		if err := emit(t); err != nil {
			return nil, err
		}
	}

	// ---- marker for placed types without decodable art -----------------------
	unknown := map[int]bool{}
	for world := 1; world <= 4; world++ {
		for lv := 1; lv <= 3; lv++ {
			for _, o := range level.DecodeObjectsByID(rom, byte(world<<4|lv)) {
				if _, ok := typeFrame[o.Type]; !ok {
					unknown[int(o.Type)] = true
				}
			}
		}
	}
	if len(unknown) > 0 {
		f, err := b.CreateFile("objects", "marker.png")
		if err != nil {
			return nil, err
		}
		err = png.Encode(f, markerImage())
		f.Close()
		if err != nil {
			return nil, err
		}
		var us []int
		for t := range unknown {
			us = append(us, t)
		}
		sort.Ints(us)
		for _, t := range us {
			name := fmt.Sprintf("Object $%02X", t)
			id := slugify(name)
			b.AddObject(schema.Asset{ID: id, Name: name, Group: "Objects"}, &schema.Object{
				Type: schema.ObjectSprite2D,
				Name: name,
				Atlas: &schema.SpriteAtlas{
					File: "marker.png", CellW: 16, CellH: 16, Anchor: []int{0, 0},
				},
				Animations: []schema.Animation{{ID: "main", Frames: 1, Loop: "hold"}},
				Props: map[string]any{"type": fmt.Sprintf("0x%02X", t),
					"art": "metasprite not decoded; shared marker shown"},
			})
			refs[fmt.Sprintf("u/%d", t)] = objRef{asset: id}
		}
	}
	ctx.Logf("objects: %d records -> %d refs (%d marker-backed types)", len(recs), len(refs), len(unknown))
	return refs, nil
}

// markerImage draws the shared placement marker: a translucent amber box with
// a solid outline, 16x16 (two tiles).
func markerImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	fill := color.NRGBA{255, 190, 40, 90}
	edge := color.NRGBA{255, 190, 40, 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c := fill
			if x == 0 || y == 0 || x == 15 || y == 15 {
				c = edge
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func slugify(s string) string {
	var out []rune
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+32)
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// worldSpriteData warps the oracle into `world` and returns its VRAM (OBJ
// tiles) and OBP0 sprite palette register — the pixels the compositor needs.
func worldSpriteData(rom []byte, world byte) (vram []byte, obp0 byte) {
	m := gameboy.NewMachine(rom)
	m.RunFrames(80)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0
	id := byte(world<<4 | 1)
	for f := 0; f < 40; f++ {
		m.Write(0xFFB4, id)
		m.RunFrame()
	}
	return m.VRAM(), m.Read(0xFF48)
}
