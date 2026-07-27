// webexport serializes the decoded Sonic levels, objects and music as a
// Retro-X game tree (see RETROX.md). Everything is reconstructed from the
// cartridge by the same decode path as cmd/levelmap and handed to the shared
// builder (tools/lib/retrox):
//
//	manifest.json                game meta + asset index (written by the builder)
//	levels/act<NN>.json          per act: tilemap body, placements, paletteFx
//	levels/shared/atlas_<k>.png  256 tiles (8x8) at a zone palette, deduped across acts
//	levels/shared/shapes.json    48 collision height profiles ($3E7A) + angles ($3978)
//	objects/<id>.json|.png       one sprite2d object per placed (zone, type) with art
//	music/<track>.mp3            the 7 zone themes, pure-ROM synth (music.go)
//
// The levels stage uses the machine oracle only to capture the paletteFx
// cycle; objects and music are oracle-free. Cross-stage references (a level's
// music binding, placements' object refs) are emitted only when the target
// stage is enabled, so partial -only runs still validate.
//
// Usage (from games/sonic-gg/): go run ./extract/cmd/webexport -in <rom.gg>
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"retroreverse.com/games/sonic-gg/extract/decomp"
	"retroreverse.com/games/sonic-gg/extract/objplace"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/gamegear"
)

const (
	descTable = 0x15600 // bank 5 $5600: 18 word-pointers -> per-act descriptors
	blockBase = 0x10000 // block tile tables: file $10000 + descriptor word
	tileBase  = 0x30000 // compressed tile sets: file $30000 + descriptor word
	palTable  = 0x23400 // bank 8 $7400: per-index palette offset table
	objTable  = 0x15600 // object tables: $15600 + descriptor word +30
	attrPtrs  = 0x343D  // per-zone block-attribute table pointers (bit7 priority, bits0-5 shape)
	shapeTbl  = 0x3E7A  // floor height-profile pointer table (shape*2 -> 32-byte profile)
	angleTbl  = 0x3978  // per-shape surface angle (signed)
	numShapes = 48      // collision shapes $00-$2F (the table ends where profile data begins)
)

var zoneNames = []string{"Green Hills", "Bridge", "Jungle", "Labyrinth", "Scrap Brain", "Sky Base", "Special Stage"}

// Bonus/special stages: descriptor indices 28-35 ($1C-$23), reached when an act is
// cleared with >=50 rings (goal handler $61F8 -> $D282=4 -> sets bonus flag IY+7 bit0;
// $D239 counts from $1C and selects the next descriptor each time). All 8 share one zone-6
// tilemap/tileset/palette; they differ in spawn, camera bounds and object (ring) layout.
const bonusFirst, bonusCount = 28, 8

// objNames maps the object type bytes we have identified (Part V §1) to display names.
var objNames = map[byte]string{
	0x00: "Sonic", 0x01: "bonus", 0x02: "bonus", 0x03: "bonus", 0x04: "shield",
	0x06: "emerald", 0x07: "goal", 0x08: "crab", 0x09: "swing platform",
	0x0B: "sinking platform", 0x0E: "bird", 0x0F: "moving platform", 0x10: "beetle", 0x12: "world 1 boss",
	0x25: "capsule", 0x26: "fish", 0x2C: "world 3 boss", 0x2D: "porcupine",
	0x48: "world 2 boss", 0x49: "world 4 boss", 0x4E: "seesaw",
	0x50: "bg animator", 0x51: "checkpoint", 0x13: "teleporter",
	0x21: "bumper", 0x52: "continue", 0x3B: "bobbing platform", 0x29: "floating log", // special-stage objects ($52 = Continue powerup)
	// (rings are not objects: they are baked into the block map as $79-$7B, like normal zones)
}

type Act struct {
	num      int
	zone     int
	name     string
	mapFile  int
	mapLen   int
	widthBlk int
	stride   int
	engZone  int // descriptor +0 = the engine zone byte ($D2D5): selects the $343D
	// collision-attribute table. act/3 except Sky Base 3 + the teleporter interiors (7).
	blkTable int
	tileFile int
	bgPal    int
	spawnX   int // descriptor +13 (blockX; the loader stores +13/+14 at ($D217)->$D362)
	spawnY   int // descriptor +14 (blockY, verbatim — no adjustment)
}

func w(rom []byte, o int) int { return int(rom[o]) | int(rom[o+1])<<8 }

// mkAct reads one scene descriptor (index in the $5600 table) into an Act.
func mkAct(rom []byte, idx, zone int, name string) Act {
	d := descTable + w(rom, descTable+idx*2)
	return Act{
		num: idx, zone: zone, name: name,
		engZone:  int(rom[d]),
		mapFile:  0x14000 + w(rom, d+15),
		mapLen:   w(rom, d+17),
		widthBlk: w(rom, d+7) / 32,
		stride:   w(rom, d+1),
		blkTable: blockBase + w(rom, d+19),
		tileFile: tileBase + w(rom, d+21),
		bgPal:    int(rom[d+29]),
		spawnX:   int(rom[d+13]),
		spawnY:   int(rom[d+14]),
	}
}

func parseActs(rom []byte) []Act {
	var acts []Act
	for i := 0; i < 18; i++ {
		acts = append(acts, mkAct(rom, i, i/3, fmt.Sprintf("%s Act %d", zoneNames[i/3], i%3+1)))
		// Each teleporter sub-scene (Part V §1) is listed right under its parent act.
		for _, e := range hiddenScenes {
			if e.parent == i {
				acts = append(acts, mkAct(rom, e.idx, e.zone, e.name))
			}
		}
	}
	// Bonus/special stages (zone 6): same descriptor layout, one shared map.
	for n := 0; n < bonusCount; n++ {
		acts = append(acts, mkAct(rom, bonusFirst+n, 6, fmt.Sprintf("Special Stage %d", n+1)))
	}
	return acts
}

// hiddenScenes are the teleporter-only sub-scenes, each listed under its parent act (num).
var hiddenScenes = []struct {
	idx, zone, parent int
	name              string
}{
	{20, 4, 13, "Scrap Brain Act 2a"}, {21, 4, 13, "Scrap Brain Act 2b"}, {22, 4, 13, "Scrap Brain Act 2c"},
	{23, 4, 13, "Scrap Brain Act 2d"}, {24, 4, 13, "Scrap Brain Act 2e"}, {25, 4, 13, "Scrap Brain Act 2f"},
	{26, 7, 16, "Sky Base Act 2a"},
}

func romPalette(rom []byte, idx int) color.Palette {
	off := w(rom, palTable+idx*2)
	return gamegear.Palette(rom[palTable+off : palTable+off+32])
}

// Obj is one object-table entry with its settled rest position.
type Obj struct {
	Type   byte
	X, Y   int // rest position (world px): spawn + the engine's placement
	Name   string
	bx, by int // spawn blocks (internal: settle input)
}

func objectTable(rom []byte, act int) []Obj {
	d := descTable + w(rom, descTable+act*2)
	t := objTable + w(rom, d+30)
	count := int(rom[t])
	objs := make([]Obj, 0, count)
	for i := 0; i < count; i++ {
		p := t + 1 + i*3
		typ := rom[p]
		bx, by := int(rom[p+1]), int(rom[p+2])
		objs = append(objs, Obj{Type: typ, X: bx * 32, Y: by * 32, Name: objNames[typ], bx: bx, by: by})
	}
	return objs
}

// settleObjects replaces each object's raw spawn (blockX*32, blockY*32) with its rest
// position on its first live frame, exactly as the engine computes it (objplace,
// verified against the running game by cmd/objsettle).
func settleObjects(rom []byte, objs []Obj, lvl *objplace.Level) {
	for i := range objs {
		t := int(objs[i].Type)
		x, y, grounded := lvl.Settle(t, objs[i].bx*32, objs[i].by*32)
		// A handler with per-frame gravity (and terrain collision) pulls its object
		// down to the floor below when the spawn block has none — the porcupine in
		// Bridge 1 drops 48 px onto the lower ground on activation (oracle-verified).
		if !grounded && objplace.HasGravity(rom, t) && !objplace.NoCollide(rom, t) {
			x, y, _ = lvl.DropToFloor(t, x, y)
		}
		objs[i].X, objs[i].Y = x, y
	}
}

// cellAnims emits one cell animator per type-$50 object (the growing-flower /
// sea-twinkle strips), all sharing the engine's global phase table.
func cellAnims(rom []byte, objs []Obj) []schema.CellAnim {
	var phases []schema.CellPhase
	for _, p := range objplace.BgAnim(rom) {
		phases = append(phases, schema.CellPhase{Tiles: p.Tiles[:], Frames: p.Frames})
	}
	var out []schema.CellAnim
	for _, o := range objs {
		if o.Type == 0x50 {
			out = append(out, schema.CellAnim{TX: o.bx * 4, TY: o.by * 4, TW: 2, TH: 4, Phases: phases})
		}
	}
	return out
}

// blockShapes returns block index -> collision shape (0-47) for a zone, from the $343D
// per-zone attribute table (low 6 bits of each block's attribute byte). Home-config
// addresses map 1:1 to file offsets, so the pointer is read directly.
func blockShapes(rom []byte, zone int) []int {
	p := w(rom, attrPtrs+zone*2)
	out := make([]int, 256)
	for b := 0; b < 256; b++ {
		out[b] = int(rom[p+b]) & 0x3F
	}
	return out
}

// animFrame is a runtime-animated tile group (see cmd/levelmap).
type animFrame struct{ vramTile, fileOff, nTiles int }

var ringAnim = animFrame{252, 0x2F73D, 4}
var zoneTileAnims = map[int][]animFrame{0: {{12, 0x2FA3D, 4}}}

func applyAnimFrame(rom, tiles []byte, zone int) {
	apply := func(a animFrame) {
		n := a.nTiles * 32
		if a.fileOff+n <= len(rom) && a.vramTile*32+n <= len(tiles) {
			copy(tiles[a.vramTile*32:], rom[a.fileOff:a.fileOff+n])
		}
	}
	// Tiles 252-255 are empty in every zone's base set (including the special stage); the
	// engine fills them with the spinning ring frames at runtime. The special stage's
	// rectangular ring fields are blocks $79-$7B (= those tiles), so it needs this too.
	apply(ringAnim)
	for _, a := range zoneTileAnims[zone] {
		apply(a)
	}
}

// Animated tile sources (bank 11, contiguous): the rings spin through 6 frames, the
// Green Hills flowers through 2, each frame being 4 tiles (16x16). Frame 0 is the one
// already baked into the base tile set; frames 1+ are appended to the atlas.
const (
	ringSrc      = 0x2F73D
	ringFrames   = 6
	flowerSrc    = 0x2FA3D
	flowerFrames = 2
)

func drawTile(img *image.RGBA, pal color.Palette, data []byte, atlasIdx int) {
	t := gamegear.DecodeTile(data)
	ox, oy := (atlasIdx%16)*8, (atlasIdx/16)*8
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(ox+x, oy+y, pal[t[y][x]])
		}
	}
}

// renderAtlas paints the 256 base tiles into a 16-wide RGBA atlas, then appends the
// animation frames (frames 1+) below them, returning the per-group tile animations.
// Atlas height is 18 rows (the appended frames fit in rows 16-17).
func renderAtlas(rom, tiles []byte, pal color.Palette, zone int) (*image.RGBA, []schema.TileAnim) {
	const cols, rows = 16, 18
	img := image.NewRGBA(image.Rect(0, 0, cols*8, rows*8))
	for ti := 0; ti < 256; ti++ {
		drawTile(img, pal, tiles[ti*32:], ti)
	}
	next := 256
	appendGroup := func(base []int, src, nframes int) schema.TileAnim {
		g := schema.TileAnim{Tiles: base, Frames: [][]int{base}, PeriodFrames: framesPerTick}
		for f := 1; f < nframes; f++ {
			idx := []int{next, next + 1, next + 2, next + 3}
			for i := 0; i < 4; i++ {
				drawTile(img, pal, rom[src+f*128+i*32:], idx[i])
			}
			g.Frames = append(g.Frames, idx)
			next += 4
		}
		return g
	}
	anims := []schema.TileAnim{appendGroup([]int{252, 253, 254, 255}, ringSrc, ringFrames)}
	if zone == 0 {
		anims = append(anims, appendGroup([]int{12, 13, 14, 15}, flowerSrc, flowerFrames))
	}
	return img, anims
}

// cyclingTiles lists the tile indices (0-255) whose pixels use one of the cycling palette
// slots. The cycle must be limited to these — other tiles may share a colour with the
// cycling slots at rest (e.g. sky in Bridge) but use a different, static palette index.
func cyclingTiles(tiles []byte, slots []int) []int {
	slot := map[int]bool{}
	for _, s := range slots {
		slot[s] = true
	}
	var out []int
	for t := 0; t < 256; t++ {
		px := gamegear.DecodeTile(tiles[t*32:])
		found := false
		for y := 0; y < 8 && !found; y++ {
			for x := 0; x < 8; x++ {
				if slot[int(px[y][x])] {
					found = true
					break
				}
			}
		}
		if found {
			out = append(out, t)
		}
	}
	return out
}

// capturePaletteCycle boots the act and watches CRAM for a BG-palette cycle, returning the
// cycling slots + per-step colours (step 0 = the static palette) + period, or nil. It is
// the one oracle-assisted part of the export (the cycle is driven at runtime).
func capturePaletteCycle(rom []byte, act int, staticPal color.Palette) *schema.PaletteCycle {
	m := gamegear.NewMachine(rom)
	m.CapturePC = 0x0A73
	for i := 0; i < 700; i++ {
		m.RunFrame()
	}
	for r := 0; r < 40 && !m.Captured; r++ {
		m.Pad00 = 0x7F
		m.Write(0xD238, byte(act))
		for i := 0; i < 8; i++ {
			m.RunFrame()
			m.Write(0xD238, byte(act))
		}
		m.Pad00 = 0xFF
		for k := 0; k < 242 && !m.Captured; k++ {
			m.Write(0xD238, byte(act))
			m.RunFrame()
		}
	}
	for i := 0; i < 200; i++ { // let the palette fade finish
		m.RunFrame()
	}
	const N = 120
	rec := make([][]string, N)
	for f := 0; f < N; f++ {
		m.RunFrame()
		rec[f] = paletteHex(gamegear.Palette(m.VDP.CRAM[:32]))
	}
	eq := func(a, b []string) bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	// Frames where the palette changes; the period is the gap BETWEEN changes (the first
	// change is just the remainder of the step the recording started in).
	var changes []int
	for f := 1; f < N; f++ {
		if !eq(rec[f], rec[f-1]) {
			changes = append(changes, f)
		}
	}
	if len(changes) < 2 {
		return nil
	}
	P := changes[1] - changes[0]
	// The distinct states in cycle order: the state at the start, then after each change.
	runVals := [][]string{rec[0]}
	for _, cf := range changes {
		runVals = append(runVals, rec[cf])
	}
	L := 0
	for cand := 1; cand <= len(runVals)/2; cand++ {
		ok := true
		for i := 0; i+cand < len(runVals); i++ {
			if !eq(runVals[i], runVals[i+cand]) {
				ok = false
				break
			}
		}
		if ok {
			L = cand
			break
		}
	}
	if L < 2 {
		return nil
	}
	steps := runVals[:L]
	var slots []int
	for i := 0; i < 16; i++ {
		for k := 1; k < len(steps); k++ {
			if steps[k][i] != steps[0][i] {
				slots = append(slots, i)
				break
			}
		}
	}
	if len(slots) == 0 {
		return nil
	}
	sp := paletteHex(staticPal)
	rot := 0
	for k := 0; k < len(steps); k++ {
		match := true
		for _, s := range slots {
			if steps[k][s] != sp[s] {
				match = false
				break
			}
		}
		if match {
			rot = k
			break
		}
	}
	pc := &schema.PaletteCycle{Slots: slots, PeriodFrames: P}
	for k := 0; k < len(steps); k++ {
		st := steps[(rot+k)%len(steps)]
		row := make([]string, len(slots))
		for j, s := range slots {
			row[j] = st[s]
		}
		pc.Steps = append(pc.Steps, row)
	}
	return pc
}

// framesPerTick is the engine's tile-animation cadence (the $15FF update's ~10-frame
// cycle), emitted as each tileAnims group's periodFrames.
const framesPerTick = 10

// sectionOf derives the menu group from the act's name: the text before " Act "
// ("Green Hills Act 1" -> "Green Hills"); the special stages group under their own name.
func sectionOf(a Act) string {
	if i := strings.Index(a.name, " Act "); i > 0 {
		return a.name[:i]
	}
	if strings.HasPrefix(a.name, "Special Stage") {
		return "Special Stage"
	}
	return a.name
}

// musicTrack returns the act's background-music track name. The music id is descriptor byte
// +36 (the loader $1A66 stores it to $D2F7 and plays it via RST $18); the id indexes the
// $4716 song table. Acts map to ids 0-5 by zone, with two Sky Base acts reusing Scrap Brain's
// theme and the special stage on id 16 (Part VI). music.go bakes these by the same names.
func musicTrack(rom []byte, act int) string {
	d := descTable + w(rom, descTable+act*2)
	names := map[byte]string{0: "greenhills", 1: "bridge", 2: "jungle", 3: "labyrinth",
		4: "scrapbrain", 5: "skybase", 16: "special"}
	return names[rom[d+36]]
}

// underwaterPalette returns the 16 static underwater BG colours the IRQ line-split writes
// to CRAM from the bank-0 table at file $0216 (Part V §3).
func underwaterPalette(rom []byte) []string {
	pal := gamegear.Palette(rom[0x0216 : 0x0216+32])
	return paletteHex(pal)
}

// waterLine returns the water surface world-Y (px) for a Labyrinth act, or -1 if the act
// has no water. The water surface is object type $40 (the first object placed); its block-Y
// times 32 is the surface line (loader $185D arms the split only for acts 9-11).
func waterLine(rom []byte, act int) int {
	if act < 9 || act > 11 {
		return -1
	}
	for _, o := range objectTable(rom, act) {
		if o.Type == 0x40 {
			return o.by * 32
		}
	}
	return -1
}

func main() {
	cli.Main("sonic-gg", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in <rom.gg> [-o <outdir>] [-only levels,objects,music,all]")
	}
	rom, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Sonic the Hedgehog")
	b.SetPlatform("Game Gear")
	b.SetYear(1991)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 160, H: 144},
		TickHz: 60,
		Filter: "gg",
	})

	var objIndex map[string]objRef
	if ctx.Stage("objects") {
		objIndex, err = exportObjects(ctx, rom)
		if err != nil {
			return err
		}
	}
	if ctx.Stage("music") {
		if err := exportMusic(ctx, rom); err != nil {
			return err
		}
	}
	if ctx.Stage("levels") {
		if err := exportLevels(ctx, rom, objIndex); err != nil {
			return err
		}
	}
	return nil
}

// exportLevels writes the per-act level documents, the shared atlases and the
// collision profiles. objIndex may be nil (objects stage disabled): then the
// levels ship without placements so the tree still validates.
func exportLevels(ctx *cli.Context, rom []byte, objIndex map[string]objRef) error {
	b := ctx.Builder

	// levels/shared/shapes.json (global collision profiles)
	sh := &schema.Shapes{Count: numShapes}
	for s := 0; s < numShapes; s++ {
		p := w(rom, shapeTbl+s*2)
		prof := make([]int, 32)
		for c := 0; c < 32; c++ {
			prof[c] = int(int8(rom[p+c])) // signed; -128 ($80) = no surface
		}
		sh.Profiles = append(sh.Profiles, prof)
		sh.Angles = append(sh.Angles, float64(int8(rom[angleTbl+s])))
	}
	b.AddSideDoc("levels/shared/shapes.json", sh)

	acts := parseActs(rom)
	// Dedup atlases by (tileFile, bgPal): same tiles + palette -> one PNG.
	atlasName := map[[2]int]string{}
	atlasAnims := map[string][]schema.TileAnim{}
	cycleCache := map[int]*schema.PaletteCycle{} // by bgPal (oracle-captured)

	const screenBlk = 5
	for idx, a := range acts {
		tiles := decomp.Decompress(rom, a.tileFile)
		applyAnimFrame(rom, tiles, a.zone)
		pal := romPalette(rom, a.bgPal)

		// atlas (deduped) -> levels/shared/
		key := [2]int{a.tileFile, a.bgPal}
		atlas, ok := atlasName[key]
		if !ok {
			atlas = fmt.Sprintf("atlas_%d.png", len(atlasName))
			atlasName[key] = atlas
			img, tileAnims := renderAtlas(rom, tiles, pal, a.zone)
			atlasAnims[atlas] = tileAnims
			if err := writePNG(b, img, "levels", "shared", atlas); err != nil {
				return err
			}
		}

		// map + geometry. The map window is usually 4096 bytes, but the hidden teleporter
		// rooms decode to a smaller buffer, so size the height from the actual decoded length.
		mp := decomp.LoadMapRLE(rom, a.mapFile, a.mapLen)
		cols := clampi(a.widthBlk+screenBlk, 1, a.stride)
		rows := len(mp) / a.stride
		cells := make([]int, cols*rows)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				cells[r*cols+c] = int(mp[r*a.stride+c])
			}
		}

		// block tile table (256 blocks x 16 tiles)
		bt := make([][]int, 256)
		for blk := 0; blk < 256; blk++ {
			row := make([]int, 16)
			for i := 0; i < 16; i++ {
				row[i] = int(rom[a.blkTable+blk*16+i])
			}
			bt[blk] = row
		}

		// Sonic's spawn is not in the object table: the loader stores descriptor +13/+14
		// at $D362 (pointed to by $D217) and $1AB3 spawns slot 0 at (blockX*32, blockY*32).
		// His handler then pulls him down onto the first floor line (gravity + the $2CD4
		// snap), so his visible start is the dropped rest position. Objects likewise rest
		// where the engine's first live frame puts them (objplace, verified by objsettle).
		objs := objectTable(rom, a.num)
		lvl := objplace.NewLevel(rom, mp, a.stride, a.engZone)
		settleObjects(rom, objs, lvl)
		anims50 := cellAnims(rom, objs) // the $50 background animators become cellAnims...
		visible := objs[:0]
		for _, o := range objs {
			if o.Type != 0x50 && o.Type != 0x40 { // ...$40 is the Labyrinth water surface line
				visible = append(visible, o)
			}
		}
		objs = visible
		sx, sy, _ := lvl.DropToFloor(0, a.spawnX*32, a.spawnY*32)

		doc := &schema.Level{
			Type:  schema.LevelTilemap,
			Camera: &schema.Camera{Mode: "map2d", Map2D: &schema.Map2D{MaxNativeFactor: 1}},
			Tilemap: &schema.Tilemap{
				TileSize: 8, Width: cols, Height: rows,
				Atlas:  schema.TileAtlas{File: "shared/" + atlas, Cols: 16},
				Cells:  cells,
				Blocks: &schema.Blocks{Size: 4, Tiles: bt, Shapes: blockShapes(rom, a.engZone)},
				// initial framing: one GG screen centred on the spawn block's anchor
				View:      &schema.Rect{X: a.spawnX*32 + 8 - 80, Y: a.spawnY*32 + 16 - 72, W: 160, H: 144},
				Spawn:     &schema.Spawn{X: sx, Y: sy},
				TileAnims: atlasAnims[atlas],
				CellAnims: anims50,
			},
			Collision: &schema.Collision{Kind: "profiles", File: "shared/shapes.json"},
		}
		if ctx.Enabled("music") {
			doc.Music = musicTrack(rom, a.num)
		}
		if objIndex != nil {
			if ref, ok := objIndex[objKey(artZone(a.zone), 0)]; ok {
				doc.Tilemap.Spawn.Object = ref.asset
				doc.Tilemap.Spawn.Anim = ref.anim
			}
			for i, o := range objs {
				ref, ok := objIndex[objKey(artZone(a.zone), int(o.Type))]
				if !ok { // a type outside the placed-type census (never seen in practice)
					fmt.Fprintf(os.Stderr, "warning: %s: type 0x%02X has no object asset; placement dropped\n", a.name, o.Type)
					continue
				}
				doc.Placements = append(doc.Placements, schema.Placement{
					ID:     i,
					Object: ref.asset,
					Anim:   ref.anim,
					Pos:    []float64{float64(o.X), float64(o.Y)},
					Name:   o.Name,
					Props:  map[string]any{"type": fmt.Sprintf("0x%02X", o.Type)},
				})
			}
		}

		// The palette cycle is oracle-captured by booting the act; only do it for the real
		// zones (the bonus stages can't be reached by forcing $D238 and have no water/lava
		// cycle anyway).
		fx := &schema.PaletteFx{Palette: paletteHex(pal)}
		var pc *schema.PaletteCycle
		if a.zone < 6 {
			var seen bool
			pc, seen = cycleCache[a.bgPal]
			if !seen {
				pc = capturePaletteCycle(rom, a.num, pal)
				cycleCache[a.bgPal] = pc
			}
		}
		if pc != nil {
			actPC := *pc // per-act copy: the cycling tiles depend on this act's tile set
			actPC.Tiles = cyclingTiles(tiles, pc.Slots)
			fx.Cycle = &actPC
		}
		if ly := waterLine(rom, a.num); ly >= 0 {
			fx.WaterLine = &schema.WaterLine{Y: ly, Palette: underwaterPalette(rom)}
		}
		if fx.Cycle != nil || fx.WaterLine != nil {
			doc.Tilemap.PaletteFx = fx
		}

		id := fmt.Sprintf("act%02d", a.num+1)
		b.AddLevel(schema.Asset{ID: id, Name: a.name, Group: sectionOf(a)}, doc)
		ctx.Progress("levels", idx+1, len(acts), fmt.Sprintf("%-16s %3dx%-3d blocks  %d placements", a.name, cols, rows, len(doc.Placements)))
	}
	return nil
}

func paletteHex(p color.Palette) []string {
	out := make([]string, len(p))
	for i, c := range p {
		r, g, bl, _ := c.RGBA()
		out[i] = fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, bl>>8)
	}
	return out
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
