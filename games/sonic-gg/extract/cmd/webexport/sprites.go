// sprites.go is the objects stage: every placed (zone, type) with renderable
// art becomes one Retro-X sprite2d object — a single row-per-animation PNG
// atlas (48x48 cells, the engine's metasprite box) plus its document. Types
// whose art cannot be extracted (the bosses decompress their own graphics and
// draw from bytecode scripts) fall back to one shared "marker" object.
//
// An object's on-screen sprite is a metasprite: a 3-row x 6-col grid of 8x16-sprite tile
// indices, walked by the sprite writer $2F07. An animated object selects its frame through
// the shared animation routine $7C75 as layoutBase + frameId*18. The referenced tiles live
// in the per-zone sprite tile set the level loader decompresses to VRAM $2000 (descriptor
// +23/+24/+25, $0406 codec), coloured by the sprite palette (descriptor +26 -> $7400).
package main

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"retroreverse.com/games/sonic-gg/extract/decomp"
	"retroreverse.com/games/sonic-gg/extract/objplace"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/gamegear"
)

// objRef points a placement at its object asset + default animation.
type objRef struct{ asset, anim string }

// objKey indexes objRef maps by placed (zone, type).
func objKey(zone, typ int) string { return fmt.Sprintf("%d/%02x", zone, typ) }

// spriteTiles builds act i's full sprite sheet (objplace.SpriteSheet: the zone's own tiles
// $00-$7F + the common block $80-$BF) and resolves its sprite palette.
func spriteTiles(rom []byte, act int) ([]byte, color.Palette) {
	d := descTable + w(rom, descTable+act*2)
	return objplace.SpriteSheet(rom, act), romPalette(rom, int(rom[d+26]))
}

// Sonic's sprite is not in the zone sheets: his handler keeps one layout ($5C1B, his 16x32
// box) and re-streams the tile GRAPHICS per pose (bank 8 + frame*192, 3bpp — $4E9A). His
// animations are per-frame byte sequences in the $5C5B table ($4E6D sequencer); the exported
// strip plays the idle -> bored program. objplace.SonicSeq/SonicFrameTiles supply both.
const sonicLayout = 0x5C1B

// sonicStrip renders Sonic's idle/bored strip: unique graphic frames plus the play sequence
// (strip-frame index, hold frames).
func sonicStrip(rom []byte, pal color.Palette) (*image.RGBA, [][]int) {
	bored, loopStep := objplace.SonicSeq(rom, 0x0D)
	gfx := []int{0}
	idxOf := map[int]int{0: 0}
	for _, st := range bored {
		if _, ok := idxOf[st.Layout]; !ok {
			idxOf[st.Layout] = len(gfx)
			gfx = append(gfx, st.Layout)
		}
	}
	strip := image.NewRGBA(image.Rect(0, 0, len(gfx)*48, 48))
	for i, f := range gfx {
		cell := renderMeta(rom[sonicLayout:sonicLayout+18], objplace.SonicFrameTiles(rom, f), pal)
		blit(strip, cell, i*48)
	}
	seq := [][]int{{0, 360}}
	for _, st := range bored {
		seq = append(seq, []int{idxOf[st.Layout], st.Frames})
	}
	for cycle := 0; cycle < 4; cycle++ {
		for _, st := range bored[loopStep:] {
			seq = append(seq, []int{idxOf[st.Layout], st.Frames})
		}
	}
	return strip, seq
}

// renderMeta renders one 18-byte metasprite layout (3 rows x 6 cols of 8x16 sprites) into a
// 48x48 RGBA image; colour index 0 of the sprite palette is transparent.
func renderMeta(layout, tiles []byte, pal color.Palette) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 6*8, 3*16))
	p := 0
	for row := 0; row < 3; row++ {
		if p >= len(layout) || layout[p] == 0xFF {
			break
		}
		for col := 0; col < 6; col++ {
			b := layout[p]
			p++
			if b >= 0xFE { // $FE skip, $FF treated as skip mid-row
				continue
			}
			for half := 0; half < 2; half++ { // 8x16 sprite = tiles b (top), b+1 (bottom)
				ti := (int(b) + half) * 32
				if ti+32 > len(tiles) {
					continue
				}
				t := gamegear.DecodeTile(tiles[ti:])
				ox, oy := col*8, row*16+half*8
				for y := 0; y < 8; y++ {
					for x := 0; x < 8; x++ {
						if v := t[y][x]; v != 0 {
							img.Set(ox+x, oy+y, pal[v])
						}
					}
				}
			}
		}
	}
	return img
}

// drawCell blits one 8x16 sprite cell (tiles t, t+1) into img at (ox, oy).
func drawCell(img *image.RGBA, tiles []byte, pal color.Palette, t, ox, oy int) {
	for half := 0; half < 2; half++ {
		ti := (t + half) * 32
		if ti+32 > len(tiles) {
			continue
		}
		cell := gamegear.DecodeTile(tiles[ti:])
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				if v := cell[y][x]; v != 0 {
					img.Set(ox+x, oy+half*8+y, pal[v])
				}
			}
		}
	}
}

func blit(dst *image.RGBA, src *image.RGBA, ox int) {
	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			dst.Set(ox+x, y, src.At(x, y))
		}
	}
}

// renderAnimStrip renders a type's full animation as a horizontal strip of 48x48 frames
// (objplace.Anim), compositing the pickup family's screen-icon overlay (drawn on top).
func renderAnimStrip(rom, tiles []byte, pal color.Palette, typ, zone int) (*image.RGBA, []int) {
	frames := objplace.Anim(rom, typ, zone)
	if len(frames) == 0 {
		return nil, nil
	}
	icons := objplace.IconOverlays(rom, typ)
	strip := image.NewRGBA(image.Rect(0, 0, len(frames)*48, 48))
	durs := make([]int, len(frames))
	for i, f := range frames {
		if f.Layout <= 0 || f.Layout+18 > len(rom) {
			return nil, nil
		}
		cell := renderMeta(rom[f.Layout:f.Layout+18], tiles, pal)
		for _, ic := range icons {
			drawCell(cell, tiles, pal, ic.Tile, ic.X, ic.Y)
		}
		blit(strip, cell, i*48)
		durs[i] = f.Frames
	}
	// Collapse animations whose frames are all pixel-identical: the pickup TV's layout blink
	// alternates a screen cell that the opaque icon fully covers — a sprite-per-scanline budget
	// trick, not a visible animation.
	if len(frames) > 1 {
		same := true
		for i := 1; same && i < len(frames); i++ {
			for y := 0; same && y < 48; y++ {
				for x := 0; x < 48; x++ {
					if strip.RGBAAt(i*48+x, y) != strip.RGBAAt(x, y) {
						same = false
						break
					}
				}
			}
		}
		if same {
			one := image.NewRGBA(image.Rect(0, 0, 48, 48))
			blit(one, strip, 0)
			return one, []int{0}
		}
	}
	return strip, durs
}

// goalStrip renders the goal sign's idle spin from its own graphics: the handler decompresses
// its sheet over VRAM $2000 and switches to sprite palette $0E (objplace.GoalAnim).
func goalStrip(rom []byte) (*image.RGBA, [][]int) {
	steps, src, palIdx := objplace.GoalAnim(rom)
	tiles := make([]byte, 256*32)
	copy(tiles, decomp.Decompress(rom, src))
	copy(tiles[0x80*32:], decomp.Decompress(rom, objplace.CommonTilesFile))
	pal := romPalette(rom, palIdx)
	var gfx []int
	idxOf := map[int]int{}
	for _, st := range steps {
		if _, ok := idxOf[st.Layout]; !ok {
			idxOf[st.Layout] = len(gfx)
			gfx = append(gfx, st.Layout)
		}
	}
	strip := image.NewRGBA(image.Rect(0, 0, len(gfx)*48, 48))
	for i, lay := range gfx {
		blit(strip, renderMeta(rom[lay:lay+18], tiles, pal), i*48)
	}
	var seq [][]int
	for _, st := range steps {
		seq = append(seq, []int{idxOf[st.Layout], st.Frames})
	}
	return strip, seq
}

// trimBBox crops an RGBA to its non-transparent bounding box; used here only to
// detect empty layouts.
func trimBBox(src *image.RGBA) (int, int, *image.RGBA) {
	b := src.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, al := src.At(x, y).RGBA(); al != 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX > maxX {
		return 0, 0, image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	out := image.NewRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			out.Set(x-minX, y-minY, src.At(x, y))
		}
	}
	return minX, minY, out
}

// artZone maps an act's engine zone to the sprite-art zone whose sheet and
// palette its objects draw with. Engine zone 7 (the Sky Base teleporter
// interior) carries Sky Base objects.
func artZone(zone int) int {
	if zone == 7 {
		return 5
	}
	return zone
}

// placedTypesByZone censuses the distinct object types placed in EVERY act
// (including the hidden teleporter sub-scenes and the special stages),
// bucketed by art zone.
func placedTypesByZone(rom []byte) [][]int {
	out := make([][]int, 7)
	for _, a := range parseActs(rom) {
		z := artZone(a.zone)
		if z < 0 || z > 6 {
			continue
		}
		seen := map[int]bool{}
		for _, t := range out[z] {
			seen[t] = true
		}
		for _, o := range objectTable(rom, a.num) {
			t := int(o.Type)
			if !seen[t] { // every placed type, even the unnamed high ones ($FF pads)
				seen[t] = true
				out[z] = append(out[z], t)
			}
		}
	}
	for z := range out {
		sortInts(out[z])
	}
	return out
}

func sortInts(v []int) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

// slugify makes a display name URL-safe ("swing platform" -> "swing-platform").
func slugify(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ' || r == '-' || r == '_':
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}

func displayName(name string, typ int) string {
	if name == "" {
		return fmt.Sprintf("Object $%02X", typ)
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// objCategories buckets the identified types for the browse list (the Part V
// handler map is the source); unnamed types land in "Misc". The order below
// is the order the groups appear in the viewer's object menu.
var categoryOrder = []string{"Player", "Enemies", "Bosses", "Platforms", "Pickups", "Level objects", "Misc"}

var objCategories = map[byte]string{
	0x00: "Player",
	0x08: "Enemies", 0x0E: "Enemies", 0x10: "Enemies", 0x26: "Enemies", 0x2D: "Enemies",
	0x12: "Bosses", 0x2C: "Bosses", 0x48: "Bosses", 0x49: "Bosses",
	0x09: "Platforms", 0x0B: "Platforms", 0x0F: "Platforms", 0x29: "Platforms",
	0x3B: "Platforms", 0x4E: "Platforms",
	0x01: "Pickups", 0x02: "Pickups", 0x03: "Pickups", 0x04: "Pickups",
	0x06: "Pickups", 0x52: "Pickups",
	0x07: "Level objects", 0x13: "Level objects", 0x21: "Level objects",
	0x25: "Level objects", 0x51: "Level objects",
}

func categoryOf(typ int) string {
	if c := objCategories[byte(typ)]; c != "" {
		return c
	}
	return "Misc"
}

// exportObjects writes the object assets and returns the (zone,type) -> asset
// map placements resolve through. An object type is ONE asset shared by every
// zone whose rendered art is identical; zones whose sheet/palette give it
// different art get their own zone-named variant. Assets are grouped by
// category (enemies, platforms, ...) rather than by zone.
func exportObjects(ctx *cli.Context, rom []byte) (map[string]objRef, error) {
	b := ctx.Builder
	index := map[string]objRef{}

	// representative act (and thus sprite sheet/palette) for each of the 7 viewer zones.
	zoneAct := []int{0, 3, 6, 9, 12, 15, 28}

	// ---- pass 1: render every placed (zone, type), keeping strips in memory ----
	type record struct {
		zone, typ int
		animID    string
		strip     *image.RGBA // nil = marker-backed (art not extractable)
		durations []int
		steps     [][]int
		path      [][]int
	}
	var recs []record
	add := func(zone, typ int, animID string, strip *image.RGBA, durations []int, steps [][]int, path [][]int) {
		recs = append(recs, record{zone, typ, animID, strip, durations, steps, path})
	}

	paths := objplace.PlatformPaths(rom)
	pathOf := func(typ int) [][]int {
		var out [][]int
		for _, p := range paths[fmt.Sprintf("%02x", typ)] {
			out = append(out, []int{p[0], p[1]})
		}
		return out
	}
	placed := placedTypesByZone(rom)
	for z, act := range zoneAct {
		tiles, pal := spriteTiles(rom, act)

		// Sonic (type $00): his own ROM tiles, this zone's sprite palette.
		sonic, seq := sonicStrip(rom, pal)
		add(z, 0, "idle", sonic, nil, seq, nil)

		for _, t := range placed[z] {
			if t == 0 || t == 0x50 || t == 0x40 { // handled elsewhere / not visible objects
				continue
			}
			if t == 0x07 { // the goal sign: own gfx but a plain metasprite — render its spin
				strip, seq := goalStrip(rom)
				add(z, t, "spin", strip, nil, seq, nil)
				continue
			}
			if t == 0x29 { // the floating log: 3 table-driven roll layouts ($803A+)
				lays := objplace.LogAnim()
				strip := image.NewRGBA(image.Rect(0, 0, len(lays)*48, 48))
				for i, lay := range lays {
					blit(strip, renderMeta(rom[lay:lay+18], tiles, pal), i*48)
				}
				seq := [][]int{{0, 240}}
				for c := 0; c < 2; c++ {
					seq = append(seq, []int{0, 6}, []int{1, 6}, []int{2, 6})
				}
				add(z, t, "roll", strip, nil, seq, pathOf(t))
				continue
			}
			// Types whose art cannot be extracted fall through to a nil strip:
			// the bosses are self-contained set-pieces that decompress their
			// OWN graphics over the zone sprite sheet and draw multi-part
			// sprites from a bytecode script — a single extracted metasprite
			// renders as garbage. They still get their own (marker-backed)
			// asset so their identity, prose and stats have a home.
			r := objplace.AnalyzeSprite(rom, t, z)
			ownGfx := false
			if _, _, own := objplace.OwnGfx(rom, t); own || t == 0x25 {
				ownGfx = true
			}
			if r.Kind == "" || r.Layout == 0 || r.Layout+18 > len(rom) || ownGfx {
				add(z, t, "main", nil, nil, nil, nil)
				continue
			}
			tt := objplace.ApplyIconUpload(rom, tiles, t)
			strip, durs := renderAnimStrip(rom, tt, pal, t, z)
			empty := strip == nil
			if !empty {
				if _, _, bb := trimBBox(strip); bb.Rect.Dx() <= 1 && bb.Rect.Dy() <= 1 {
					empty = true // empty layout
				}
			}
			if empty {
				add(z, t, "main", nil, nil, nil, nil)
				continue
			}
			add(z, t, "main", strip, durs, nil, pathOf(t))
		}
		ctx.Progress("objects", z+1, len(zoneAct), fmt.Sprintf("%-12s %d (zone,type) pairs rendered", zoneNames[z], len(recs)))
	}

	// ---- pass 2: collapse identical art, emit one asset per variant ----------
	// The signature is the rendered strip's pixels plus its animation program;
	// two zones whose sheets/palettes draw the type identically share one asset.
	sig := func(r record) string {
		h := sha256.New()
		if r.strip != nil {
			h.Write(r.strip.Pix)
			fmt.Fprint(h, r.strip.Rect)
		}
		fmt.Fprintf(h, "|%s|%v|%v|%v", r.animID, r.durations, r.steps, r.path)
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	type variant struct {
		rec   record
		zones []int
	}
	variants := map[int][]*variant{} // type -> distinct-art variants, in zone order
	var types []int
	for _, r := range recs {
		vs := variants[r.typ]
		if vs == nil {
			types = append(types, r.typ)
		}
		s := sig(r)
		found := false
		for _, v := range vs {
			if sig(v.rec) == s {
				v.zones = append(v.zones, r.zone)
				found = true
				break
			}
		}
		if !found {
			variants[r.typ] = append(vs, &variant{rec: r, zones: []int{r.zone}})
		}
	}
	sortInts(types)

	if err := writePNG(b, markerImage(), "objects", "marker.png"); err != nil {
		return nil, err
	}
	usedIDs := map[string]bool{}
	uniqueID := func(base string) string {
		id := base
		for n := 2; usedIDs[id]; n++ {
			id = fmt.Sprintf("%s-%d", base, n)
		}
		usedIDs[id] = true
		return id
	}
	// Types sharing one display name (the three bonus-TV types) carry their
	// type byte so the list can tell them apart.
	nameTypes := map[string]int{}
	for _, t := range types {
		nameTypes[displayName(objNames[byte(t)], t)]++
	}
	total, markers := 0, 0
	for _, cat := range categoryOrder {
		for _, t := range types {
			if categoryOf(t) != cat {
				continue
			}
			vs := variants[t]
			for _, v := range vs {
				name := displayName(objNames[byte(t)], t)
				if nameTypes[name] > 1 {
					name = fmt.Sprintf("%s ($%02X)", name, t)
				}
				// A type with several looks names each confined variant by its
				// zone; a variant shared across zones is the common look and
				// keeps the plain name.
				if len(vs) > 1 && len(v.zones) == 1 {
					name = fmt.Sprintf("%s (%s)", name, zoneNames[v.zones[0]])
				}
				id := uniqueID(slugify(name))
				zones := make([]string, len(v.zones))
				for i, z := range v.zones {
					zones[i] = zoneNames[z]
				}
				doc := &schema.Object{
					Type:  schema.ObjectSprite2D,
					Name:  name,
					Props: map[string]any{"type": fmt.Sprintf("0x%02X", t), "zones": zones},
				}
				r := v.rec
				if r.strip == nil {
					doc.Atlas = &schema.SpriteAtlas{File: "marker.png", CellW: 32, CellH: 32}
					doc.Animations = []schema.Animation{{ID: "main", Frames: 1, Loop: "hold"}}
					doc.Props["art"] = "not extractable (own-graphics set-piece); shared marker shown"
					markers++
				} else {
					frames := r.strip.Rect.Dx() / 48
					anim := schema.Animation{ID: r.animID, Frames: frames, Loop: "loop"}
					if frames == 1 {
						anim.Loop = "hold"
					}
					if len(r.steps) > 0 {
						anim.Steps = r.steps
					} else if frames > 1 {
						anim.Durations = r.durations
					}
					anim.Path = r.path
					doc.Atlas = &schema.SpriteAtlas{File: id + ".png", CellW: 48, CellH: 48}
					doc.Animations = []schema.Animation{anim}
					if err := writePNG(b, r.strip, "objects", id+".png"); err != nil {
						return nil, err
					}
					total++
				}
				b.AddObject(schema.Asset{ID: id, Name: name, Group: cat}, doc)
				for _, z := range v.zones {
					index[objKey(z, t)] = objRef{asset: id, anim: r.animID}
				}
			}
		}
	}
	ctx.Logf("objects: %d assets (%d with art, %d marker-backed) from %d placed (zone,type) pairs",
		total+markers, total, markers, len(recs))
	return index, nil
}

// markerImage draws the shared placement marker: a translucent amber box with
// a solid outline, 32x32 (one block half).
func markerImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fill := color.RGBA{255, 190, 40, 90}
	edge := color.RGBA{255, 190, 40, 255}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c := fill
			if x == 0 || y == 0 || x == 31 || y == 31 {
				c = edge
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func writePNG(b *build.Builder, img image.Image, rel ...string) error {
	f, err := b.CreateFile(rel...)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
