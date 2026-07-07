package main

// 3D object placement — doors, pillars, bridges, levers — baked into the level
// mesh. The pipeline mirrors the game's object-emission pass (overlay 2DFE,
// traced in the oracle):
//
//   - An object is 3D when its COMOBJ.DAT render class (byte0 & 3) is 2.
//   - Items whose low 6 bits are >= 0x10 map to a model number through a
//     32-entry word table at DGROUP:056C (UW.EXE file offset 0x5F85C).
//   - Items with low6 < 0x10 are the door family, emitted by dedicated code
//     (2DFE:0CBA): a doorframe (model 1) plus a leaf — the wooden door
//     (model 14) for the closed variants, the portcullis (model 12), or the
//     open-door leaf (model 15).
//   - The model program runs with pool slots 128/256 pre-loaded with the
//     floor-to-ceiling vector (0, 1024-z, 0) — model Y units are the level's
//     fine height units where the ceiling sits at 16*64 = 1024 (the emitter
//     writes exactly 0x400 - Z). That is how frames and pillars stretch.
//   - World scale: one tile = 256 model units on every axis; the object's
//     heading (w1 bits 7-9) turns the model in 45° steps.
//   - Door textures: the level texture list's trailing 6 bytes pick DOORS.GR
//     images for door variants 0-5.

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"

	"ultimaunderworld/extract/crit"
	"ultimaunderworld/extract/lev"
	"ultimaunderworld/extract/model"
	"ultimaunderworld/extract/tex"
)

// modelTableFileOff is the 32-entry item->model word table (DGROUP:056C) in
// UW.EXE; -1 entries have no model. Verified equal to live memory.
const modelTableFileOff = 0x5F85C

// modelFlagsFileOff is the per-model render-flags table (DGROUP:04F4, stride 4,
// byte0 = flags) in UW.EXE. Bit 0x20 = "texture the model's faces with the
// tile's WALL texture" (the emit function 2DFE:05E0 reads it into DI and, when
// set, binds arg wallTex+0x3A). Bit 0x08 = ceiling-adaptive (env vector).
// So the stone doorframe (model 1, flags 0x21) is wall-textured while the
// wooden leaf (model 14, flags 0x11) keeps its own DOORS.GR image.
const modelFlagsFileOff = 0x5F7E4

// modelWallTextured reports whether model n's faces take the tile wall texture.
func modelWallTextured(exe []byte, n int) bool {
	off := modelFlagsFileOff + n*4
	return off < len(exe) && exe[off]&0x20 != 0
}

const modelUnitsPerTile = 256.0

// Door-family variants (item id & 0xF): 0-5 are the six leveled door types,
// 6 the portcullis, 7 the secret door; +8 = the same doors standing open.
const (
	doorLeafModel   = 14
	doorFrameModel  = 1
	portcullisModel = 12
	openLeafModel   = 15
)

// doorOpeningHeight is the door opening / leaf height in model units (leaf model
// 14 is 208 tall). The frame's lintel env vector is measured from the opening
// top, so the frame is decoded with env reduced by this amount.
const doorOpeningHeight = 208

// flat-poly colours. A flat face's opcode (0xBC) carries a shade LEVEL (0,1,2…);
// the VM resolves the final palette index through a runtime-built light table
// (CS:696E, generated from the palette — not in the EXE) as
// shadeTable[level+light][baseColor], where baseColor is the object's poked
// [2920] value. For the wooden door (baseColor 236) that walks the palette's
// brown ramp pal[235..238] = the real torch-lit door shades below (higher level
// = darker face). Objects with a different material poke a different baseColor;
// static export uses this wood ramp as the common case (metals get an explicit
// colour, e.g. the portcullis).
var polyShades = map[uint16]color.RGBA{
	0: {88, 52, 12, 255}, // pal[235] — brightest face
	1: {72, 40, 12, 255}, // pal[236] — [2920] base for the door
	2: {60, 36, 12, 255}, // pal[237] — darker face
}

type objMaterials struct {
	o       *outMesh
	matOf   map[string]int // material key -> index in o.Textures
	doorGR  *tex.GR
	pal     tex.Palette
	wallTR  *tex.TR
	texMap  *lev.TexMap
	nextMat int
}

func (m *objMaterials) doorMat(img int) int {
	key := fmt.Sprintf("door:%d", img)
	if i, ok := m.matOf[key]; ok {
		return i
	}
	im, err := m.doorGR.Image(img, m.pal)
	must(err)
	i := len(m.o.Textures)
	m.o.Textures = append(m.o.Textures, outTexture{Num: 3000 + img, PNG: toDataURI(im)})
	m.matOf[key] = i
	return i
}

func (m *objMaterials) colorMat(c color.RGBA) int {
	key := fmt.Sprintf("rgb:%d,%d,%d", c.R, c.G, c.B)
	if i, ok := m.matOf[key]; ok {
		return i
	}
	im := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for p := 0; p < 4; p++ {
		im.Pix[p*4+0], im.Pix[p*4+1], im.Pix[p*4+2], im.Pix[p*4+3] = c.R, c.G, c.B, 255
	}
	i := len(m.o.Textures)
	m.o.Textures = append(m.o.Textures, outTexture{Num: 4000 + i, PNG: toDataURI(im)})
	m.matOf[key] = i
	return i
}

// wallMat reuses the tile's wall texture for models the game binds it to.
func (m *objMaterials) wallMat(idx uint8) int {
	num := int(m.texMap.WallTexture(idx))
	key := fmt.Sprintf("wall:%d", num)
	if i, ok := m.matOf[key]; ok {
		return i
	}
	im, err := m.wallTR.Image(num%m.wallTR.Count(), m.pal)
	must(err)
	i := len(m.o.Textures)
	m.o.Textures = append(m.o.Textures, outTexture{Wall: true, Num: num, PNG: toDataURI(im)})
	m.matOf[key] = i
	return i
}

// placedModel positions one decoded model in the level.
type placedModel struct {
	m       *model.Model
	tileX   float32 // world position, tile units
	tileY   float32
	height  float32 // fine height units (z*8)
	heading uint8   // 0-7, 45° steps
	texMat  int     // material for textured polys (desc-bound), -1 = none
}

// spriteWorldPerTexel scales an object sprite's texels to world (tile) units.
// UW object sprites are ~16 texels; 1/32 tile per texel makes a 16-texel sprite
// half a tile — a reasonable size for floor items and creatures.
const spriteWorldPerTexel = 1.0 / 32

// creatureIdleFps is the playback rate of a creature's idle cycle. The segment
// lists imply one frame per animation tick but the tick rate isn't traced; the
// idle is mostly a slow look-around, so 1 fps (one pose per second) reads right
// — faster looks like the creature is shaking its head.
const creatureIdleFps = 1

// appendBillboards emits the level's class-0/1 objects as camera-facing sprites.
// Items (class 0/1 non-creature) use OBJECTS.GR frame [item_id]; CREATURES
// (class-1 mobile objects, item 64-127) use their CRIT/CR<octal>PAGE.N01 sprite.
// The base of each sprite sits on the object's floor position.
func appendBillboards(o *outMesh, grid *lev.Grid, block []byte, comObj *lev.ComObj,
	objGR *tex.GR, allPals []byte, pal tex.Palette, gamePath string) {

	critCache := map[int]*crit.Page{} // creature index -> parsed page (nil = missing)
	loadCrit := func(idx int) *crit.Page {
		if p, ok := critCache[idx]; ok {
			return p
		}
		// Creature N → CR<octal N>PAGE.N01 (the files are octal-numbered).
		fn := fmt.Sprintf("CR%02oPAGE.N01", idx)
		var p *crit.Page
		if b, err := os.ReadFile(filepath.Join(gamePath, "CRIT", fn)); err == nil {
			p, _ = crit.ParsePage(b)
		}
		critCache[idx] = p
		return p
	}

	texOf := map[string]int{} // sprite key -> SpriteTex index
	spriteTex := func(img *image.RGBA, key string) int {
		if ti, ok := texOf[key]; ok {
			return ti
		}
		ti := len(o.SpriteTex)
		o.SpriteTex = append(o.SpriteTex, toDataURI(img))
		texOf[key] = ti
		return ti
	}
	for _, obj := range lev.Objects(grid, block) {
		cls := comObj.RenderClass(obj.ItemID)
		if cls != lev.RenderNone && cls != lev.RenderBillboard {
			continue // 3D models and specials handled elsewhere
		}
		tx := float32(obj.TileX) + (float32(obj.FineX)+0.5)/8
		ty := float32(obj.TileY) + (float32(obj.FineY)+0.5)/8
		base := float32(obj.Z) / 32 // floor height in tile units (Z*8/256)

		if cls == lev.RenderBillboard && obj.Mobile && obj.ItemID >= 64 && obj.ItemID <= 127 {
			// Creature: emit all eight compass views, each as its idle frame cycle
			// (segment 0-7 of the primary animation). The viewer selects a view per
			// render from the camera bearing and the object heading and loops the
			// cycle; non-directional creatures collapse to view 0 for all eight.
			pg := loadCrit(int(obj.ItemID) - 64)
			if pg == nil || pg.NumFrames() == 0 {
				continue
			}
			views := make([][]outDir, 8)
			okAll := true
			for v := 0; v < 8 && okAll; v++ {
				for _, frame := range pg.ViewCycle(v) {
					im, err := pg.Frame(frame, pal)
					if err != nil {
						okAll = false
						break
					}
					views[v] = append(views[v], outDir{
						Tex: spriteTex(im, fmt.Sprintf("crit:%d:%d", obj.ItemID, frame)),
						W:   float32(im.Bounds().Dx()) * spriteWorldPerTexel,
						H:   float32(im.Bounds().Dy()) * spriteWorldPerTexel,
					})
				}
			}
			if !okAll {
				continue
			}
			o.Creatures = append(o.Creatures, outCreature{
				Pos: [3]float32{tx, base, -ty}, Heading: int(obj.Heading),
				Views: views, Fps: creatureIdleFps,
			})
			continue
		}

		// Item: OBJECTS.GR frame [item_id] (the emitter's [BP-4]).
		frame := int(obj.ItemID)
		if frame >= objGR.Count() {
			continue
		}
		im, err := objGR.Sprite(frame, pal, allPals)
		if err != nil {
			continue
		}
		ti := spriteTex(im, fmt.Sprintf("obj:%d", frame))
		o.Sprites = append(o.Sprites, outSprite{
			Pos: [3]float32{tx, base, -ty},
			W:   float32(im.Bounds().Dx()) * spriteWorldPerTexel,
			H:   float32(im.Bounds().Dy()) * spriteWorldPerTexel,
			Tex: ti,
		})
	}
}

// appendObjects decodes the level's 3D-class objects and bakes their models
// into the mesh as extra triangle groups.
func appendObjects(o *outMesh, grid *lev.Grid, block []byte, exeBytes []byte,
	comObj *lev.ComObj, tm *lev.TexMap, doorGR *tex.GR, wallTR *tex.TR, pal tex.Palette) {

	mats := &objMaterials{o: o, matOf: map[string]int{}, doorGR: doorGR, pal: pal, wallTR: wallTR, texMap: tm}

	// item->model table (DGROUP:056C image).
	var itemModel [32]int16
	for i := range itemModel {
		itemModel[i] = int16(uint16(exeBytes[modelTableFileOff+2*i]) | uint16(exeBytes[modelTableFileOff+2*i+1])<<8)
	}

	// Models are re-decoded per distinct env height (the floor-to-ceiling
	// vector differs by object Z) and swing angle (doors are rendered open);
	// cache the runs.
	type decKey struct {
		envH  int16
		swing float64
	}
	decoded := map[decKey][]*model.Model{}
	modelsForSwing := func(envH int16, swing float64) []*model.Model {
		k := decKey{envH, swing}
		if ms, ok := decoded[k]; ok {
			return ms
		}
		env := map[uint16]model.Vertex{
			128: {Y: envH},
			256: {Y: envH},
		}
		ms, err := model.DecodeWithEnvSwing(exeBytes, env, swing)
		must(err)
		decoded[k] = ms
		return ms
	}
	modelsFor := func(envH int16) []*model.Model { return modelsForSwing(envH, 0) }

	var placed []placedModel
	for _, obj := range lev.Objects(grid, block) {
		if comObj.RenderClass(obj.ItemID) != lev.RenderModel {
			continue
		}
		fineH := int16(obj.Z) * 8 // object Z in fine height units (z<<3)
		envH := int16(1024) - fineH
		ms := modelsFor(envH)
		low6 := obj.ItemID & 0x3F

		place := func(mm []*model.Model, n int, texMat int, snap bool) {
			if n < 0 || n >= len(mm) || mm[n] == nil || mm[n].Offset == 0 {
				return
			}
			pm := placedModel{
				m:      mm[n],
				tileX:  float32(obj.TileX) + (float32(obj.FineX)+0.5)/8,
				tileY:  float32(obj.TileY) + (float32(obj.FineY)+0.5)/8,
				height: float32(fineH), heading: obj.Heading, texMat: texMat,
			}
			if snap {
				// Doors CENTRE in their doorway tile, not at the object's stored
				// fine offset: the level-0 door is stored at fine (5,3) but the
				// game renders it filling the tile (placing it at the raw fine
				// position leaves it off-centre with a jamb poking out — verified
				// by rendering both). The frame spans a full tile (-112..+144
				// model units: opening -48..+80 with 64-unit jambs), so we anchor
				// it on the tile edge along the leaf axis.
				sin, cos := headingSinCos(pm.heading)
				lo := float32(-112) / modelUnitsPerTile
				if cos > 0.5 || cos < -0.5 { // leaf axis = world X
					pm.tileX = float32(obj.TileX) - lo*cos
					if cos < 0 {
						pm.tileX = float32(obj.TileX) + 1 + lo*(-cos)
					}
				} else { // leaf axis = world Y
					pm.tileY = float32(obj.TileY) - lo*sin
					if sin < 0 {
						pm.tileY = float32(obj.TileY) + 1 + lo*(-sin)
					}
				}
			}
			placed = append(placed, pm)
		}

		wallIdx := grid.At(obj.TileX, obj.TileY).WallTex

		// modelTex picks a model's textured-face material the way the emitter
		// does: wall-flagged models take the tile wall texture; a wooden door
		// leaf takes its DOORS.GR image; the portcullis is iron; anything else
		// leaves its faces to the flat-shade colours (texMat -1).
		modelTex := func(n, doorImg int) int {
			switch {
			case modelWallTextured(exeBytes, n):
				return mats.wallMat(wallIdx)
			case n == portcullisModel:
				return mats.colorMat(color.RGBA{70, 70, 78, 255}) // iron grate
			case doorImg >= 0:
				return mats.doorMat(doorImg)
			}
			return -1
		}

		if low6 < 0x10 { // door family: stone frame + leaf
			variant := low6 & 7
			leaf, doorImg := doorLeafModel, -1
			switch {
			case variant == 6:
				leaf = portcullisModel
			case variant == 7: // secret door — leaf reads as the surrounding wall
			default:
				doorImg = int(tm.DoorTexture(uint8(variant)))
			}
			if low6 >= 8 { // standing open
				leaf = openLeafModel
			}
			// The doorframe's ceiling-adaptive lintel (op 0x8C adds the env vector
			// to the opening-top vertices at y=208) needs env measured from the
			// OPENING TOP, not the floor: the game raises [3322] by the opening
			// height before computing 0x400-[3322] (captured pool slot 256 = 176,
			// not 384). Decode the frame with env reduced by that opening height
			// so its lintel lands exactly on the ceiling instead of overshooting.
			frameModels := modelsFor(envH - doorOpeningHeight)
			place(frameModels, doorFrameModel, modelTex(doorFrameModel, -1), true) // model 1 -> wall
			leafMat := modelTex(leaf, doorImg)
			if variant == 7 { // secret door leaf: force the wall texture
				leafMat = mats.wallMat(wallIdx)
			}
			// Swing the wooden leaf fully open (portcullis raises via a bar count,
			// not a rotation, so it is left un-swung).
			leafModels := ms
			if leaf == doorLeafModel {
				leafModels = modelsForSwing(envH, math.Pi/2)
			}
			place(leafModels, leaf, leafMat, true)
			continue
		}
		mn := itemModel[low6-0x10]
		if mn < 0 {
			continue
		}
		place(ms, int(mn), modelTex(int(mn), -1), false)
	}

	for _, pm := range placed {
		bakeModel(o, mats, pm)
	}
}

// bakeModel appends one placed model's triangles.
func bakeModel(o *outMesh, mats *objMaterials, pm placedModel) {
	sin, cos := headingSinCos(pm.heading)
	// model (X width, Y up, Z thickness) -> world tile units, rotated about the
	// vertical, then to the viewer's Y-up frame (x, up, -y) like the level mesh.
	xf := func(v model.Vertex) [3]float32 {
		mx := float32(v.X+pm.m.Shift[0]) / modelUnitsPerTile
		my := float32(v.Y+pm.m.Shift[1]) / modelUnitsPerTile
		mz := float32(v.Z+pm.m.Shift[2]) / modelUnitsPerTile
		wx := pm.tileX + mx*cos - mz*sin
		wy := pm.tileY + mx*sin + mz*cos
		wup := pm.height/modelUnitsPerTile + my
		return [3]float32{wx, wup, -wy}
	}

	appendTri := func(mat int, a, b, c [3]float32, uv [3][2]float32) {
		// winding doesn't matter: the viewer renders every material DoubleSide
		g := findGroup(o, mat)
		for _, p := range [][3]float32{a, b, c} {
			o.Positions = append(o.Positions, p[0], p[1], p[2])
		}
		for _, t := range uv {
			o.UVs = append(o.UVs, t[0], t[1])
		}
		g.Count += 3
	}

	fan := func(mat int, slots []uint16, textured bool) {
		if len(slots) < 3 {
			return
		}
		vs := make([]model.Vertex, len(slots))
		for i, s := range slots {
			vs[i] = pm.m.Pool[s]
		}
		// Planar UVs over the polygon's own extent: one texture copy per face,
		// matching the compact quad ops' implicit corner UVs. U/V span the two
		// widest model axes; V runs downward from the top (paint-order match).
		lo := [3]int16{vs[0].X, vs[0].Y, vs[0].Z}
		hi := lo
		for _, v := range vs {
			for k, c := range [3]int16{v.X, v.Y, v.Z} {
				if c < lo[k] {
					lo[k] = c
				}
				if c > hi[k] {
					hi[k] = c
				}
			}
		}
		span := [3]int{int(hi[0] - lo[0]), int(hi[1] - lo[1]), int(hi[2] - lo[2])}
		// smallest span = the face's flat axis; the other two carry U (width)
		// and V (height: model Y if it is in play).
		flat := 0
		for k := 1; k < 3; k++ {
			if span[k] < span[flat] {
				flat = k
			}
		}
		ua, va := 0, 1
		switch flat {
		case 0:
			ua, va = 2, 1
		case 1:
			ua, va = 0, 2
		case 2:
			ua, va = 0, 1
		}
		uvOf := func(v model.Vertex) [2]float32 {
			c := [3]int16{v.X, v.Y, v.Z}
			var u, w float32
			if span[ua] > 0 {
				u = float32(c[ua]-lo[ua]) / float32(span[ua])
			}
			if span[va] > 0 {
				w = float32(c[va]-lo[va]) / float32(span[va])
			}
			return [2]float32{u, w}
		}
		for i := 1; i+1 < len(vs); i++ {
			appendTri(mat,
				xf(vs[0]), xf(vs[i]), xf(vs[i+1]),
				[3][2]float32{uvOf(vs[0]), uvOf(vs[i]), uvOf(vs[i+1])})
		}
	}

	// fanUV triangulates a polygon using the model's own per-vertex texture
	// coordinates (op 0xA8/0xB4). The stream stores U,V as code words the draw
	// handler scales by texW>>8 / texH>>16, so a normalised coordinate is
	// code/65536; V is flipped for three.js's bottom-left texture origin.
	fanUV := func(mat int, slots []uint16, uv [][2]uint16) {
		if len(slots) < 3 {
			return
		}
		vs := make([]model.Vertex, len(slots))
		for i, s := range slots {
			vs[i] = pm.m.Pool[s]
		}
		tc := func(i int) [2]float32 {
			return [2]float32{
				float32(uv[i][0]) / model.UVScale,
				1 - float32(uv[i][1])/model.UVScale,
			}
		}
		for i := 1; i+1 < len(vs); i++ {
			appendTri(mat, xf(vs[0]), xf(vs[i]), xf(vs[i+1]),
				[3][2]float32{tc(0), tc(i), tc(i + 1)})
		}
	}

	for _, q := range pm.m.Quads {
		mat := pm.texMat
		if mat < 0 || !q.Textured {
			mat = mats.colorMat(shadeColor(q.Desc))
		}
		fan(mat, q.V[:], true)
	}
	for _, pl := range pm.m.Polys {
		if pl.Color&0xF000 != 0 && pm.texMat >= 0 { // textured n-gon
			if len(pl.UV) == len(pl.V) {
				fanUV(pm.texMat, pl.V, pl.UV) // real per-vertex UVs
			} else {
				fan(pm.texMat, pl.V, true)
			}
			continue
		}
		fan(mats.colorMat(shadeColor(pl.Color&0xFFF)), pl.V, false)
	}
}

func shadeColor(base uint16) color.RGBA {
	if c, ok := polyShades[base]; ok {
		return c
	}
	return color.RGBA{100, 100, 100, 255}
}

func headingSinCos(h uint8) (sin, cos float32) {
	// 45° steps; exact values for the multiples keep the math clean
	tab := [8][2]float32{
		{0, 1}, {0.7071068, 0.7071068}, {1, 0}, {0.7071068, -0.7071068},
		{0, -1}, {-0.7071068, -0.7071068}, {-1, 0}, {-0.7071068, 0.7071068},
	}
	return tab[h&7][0], tab[h&7][1]
}

// findGroup returns (creating if needed) the trailing group for material mat;
// object groups are appended after the level's groups.
func findGroup(o *outMesh, mat int) *outGroup {
	if n := len(o.Groups); n > 0 && o.Groups[n-1].Material == mat &&
		o.Groups[n-1].Start+o.Groups[n-1].Count == len(o.Positions)/3 {
		return &o.Groups[n-1]
	}
	o.Groups = append(o.Groups, outGroup{Start: len(o.Positions) / 3, Material: mat})
	return &o.Groups[len(o.Groups)-1]
}
