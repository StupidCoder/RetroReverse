package main

// atlas.go builds textured meshes for the GLB export. Every textured quad
// references a PSX texture page + CLUT pair; the builder bakes each distinct
// pair to a 256×256 RGBA tile (sampled from the replayed VRAM exactly as the
// rasterizer would) and packs the tiles into one atlas per model. Vertices
// are deduplicated on position + texel, and PSX coordinates (Y down, Z into
// the screen) become glTF's Y up / Z out via (x, -y, -z).

import (
	"image"
	"math"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
	"retroreverse.com/tools/lib/glb"
)

// worldScale converts game units to GLB units (a 2048-unit grid cell becomes
// 2 units).
const worldScale = 1.0 / 1024

const tileSize = 256 // one PSX texture page, texel-addressed 0..255

type pageKey struct {
	page, clut uint16
	set        uint8       // scenery set for quadrant pages; 0 otherwise
	win        rr.TexWindow // texture window (48-byte records); zero otherwise
}

type vkey struct {
	x, y, z int32 // world-space, game units
	tile    int   // -1 = untextured
	u, v    byte
}

// meshBuilder accumulates one model's geometry: textured tris over a shared
// tile atlas plus flat-coloured tris. vrams holds one texture state per
// scenery set (course.go); a model with no set-dependent pages passes one.
type meshBuilder struct {
	vrams []*rr.VRAM

	tiles   []pageKey
	tileIdx map[pageKey]int

	verts   []vkey
	index   map[vkey]uint32
	texTris [][3]uint32
	colTris map[[3]uint8][][3]uint32
}

func newMeshBuilder(vrams ...*rr.VRAM) *meshBuilder {
	return &meshBuilder{
		vrams:   vrams,
		tileIdx: map[pageKey]int{},
		index:   map[vkey]uint32{},
		colTris: map[[3]uint8][][3]uint32{},
	}
}

func (b *meshBuilder) vertex(k vkey) uint32 {
	if i, ok := b.index[k]; ok {
		return i
	}
	i := uint32(len(b.verts))
	b.verts = append(b.verts, k)
	b.index[k] = i
	return i
}

// quad splits a PSX quad (vertex order TL, TR, BL, BR) into two triangles.
func quadTris(a, b, c, d uint32) [][3]uint32 {
	return [][3]uint32{{a, b, c}, {b, d, c}}
}

// quadrantPage reports whether a texture page lies in the race scenery
// quadrant, VRAM (640,256)-(1024,512) — the pages whose content depends on
// the scenery set.
func quadrantPage(page uint16) bool {
	return page&0x10 != 0 && page&0xF >= 0xA
}

// AddTextured adds one textured quad at a world offset (game units), sampling
// the given scenery set's texture state for quadrant pages and applying the
// quad's texture window (48-byte records; zero for the rest).
func (b *meshBuilder) AddTextured(v [4][3]int16, uv [4]rr.UV, page, clut uint16, off [3]int32, set int, win rr.TexWindow) {
	if !quadrantPage(page) {
		set = 0 // content identical in every set; share the tile
	}
	key := pageKey{page, clut, uint8(set), win}
	tile, ok := b.tileIdx[key]
	if !ok {
		tile = len(b.tiles)
		b.tiles = append(b.tiles, key)
		b.tileIdx[key] = tile
	}
	var idx [4]uint32
	for i := 0; i < 4; i++ {
		idx[i] = b.vertex(vkey{
			x: int32(v[i][0]) + off[0], y: int32(v[i][1]) + off[1], z: int32(v[i][2]) + off[2],
			tile: tile, u: uv[i].U, v: uv[i].V,
		})
	}
	b.texTris = append(b.texTris, quadTris(idx[0], idx[1], idx[2], idx[3])...)
}

// AddFlat adds one untextured quad with a PSX 8-bit-per-channel colour.
func (b *meshBuilder) AddFlat(v [4][3]int16, rgb [3]uint8, off [3]int32) {
	var idx [4]uint32
	for i := 0; i < 4; i++ {
		idx[i] = b.vertex(vkey{
			x: int32(v[i][0]) + off[0], y: int32(v[i][1]) + off[1], z: int32(v[i][2]) + off[2],
			tile: -1,
		})
	}
	b.colTris[rgb] = append(b.colTris[rgb], quadTris(idx[0], idx[1], idx[2], idx[3])...)
}

// Write bakes the atlas and writes the GLB.
func (b *meshBuilder) Write(path string) error {
	atlas, cols := b.bakeAtlas()

	pos := make([][3]float32, len(b.verts))
	uvs := make([][2]float32, len(b.verts))
	aw := float32(1)
	ah := float32(1)
	if atlas != nil {
		aw = float32(atlas.Bounds().Dx())
		ah = float32(atlas.Bounds().Dy())
	}
	for i, k := range b.verts {
		pos[i] = [3]float32{
			float32(k.x) * worldScale,
			-float32(k.y) * worldScale,
			-float32(k.z) * worldScale,
		}
		if k.tile >= 0 {
			tx := float32((k.tile % cols) * tileSize)
			ty := float32((k.tile / cols) * tileSize)
			uvs[i] = [2]float32{(tx + float32(k.u) + 0.5) / aw, (ty + float32(k.v) + 0.5) / ah}
		}
	}

	var texGroups []glb.TexturedGroup
	if len(b.texTris) > 0 {
		texGroups = append(texGroups, glb.TexturedGroup{Tris: b.texTris, Image: atlas})
	}
	var colGroups []glb.TriGroup
	for rgb, tris := range b.colTris {
		colGroups = append(colGroups, glb.TriGroup{Tris: tris, Color: linColor(rgb)})
	}
	return glb.WriteTextured(path, pos, uvs, texGroups, colGroups)
}

// bakeAtlas renders every referenced page+CLUT pair to its 256×256 tile.
// VRAM texel value 0 is fully transparent (the PSX convention); everything
// else is opaque sRGB.
func (b *meshBuilder) bakeAtlas() (*image.RGBA, int) {
	if len(b.tiles) == 0 {
		return nil, 1
	}
	cols := len(b.tiles)
	if cols > 4 {
		cols = 4
	}
	rows := (len(b.tiles) + cols - 1) / cols
	img := image.NewRGBA(image.Rect(0, 0, cols*tileSize, rows*tileSize))
	for t, key := range b.tiles {
		ox := (t % cols) * tileSize
		oy := (t / cols) * tileSize
		vram := b.vrams[0]
		if int(key.set) < len(b.vrams) {
			vram = b.vrams[key.set]
		}
		for v := 0; v < tileSize; v++ {
			for u := 0; u < tileSize; u++ {
				px := vram.TexelW(key.page, key.clut, byte(u), byte(v), key.win)
				o := img.PixOffset(ox+u, oy+v)
				if px == 0 {
					continue // transparent black, zero-initialised
				}
				img.Pix[o+0] = uint8((px & 0x1F) << 3)
				img.Pix[o+1] = uint8(((px >> 5) & 0x1F) << 3)
				img.Pix[o+2] = uint8(((px >> 10) & 0x1F) << 3)
				img.Pix[o+3] = 255
			}
		}
	}
	return img, cols
}

// linColor converts an 8-bit sRGB colour to the linear baseColorFactor glTF
// expects.
func linColor(rgb [3]uint8) [3]float32 {
	var out [3]float32
	for i, c := range rgb {
		s := float64(c) / 255
		if s <= 0.04045 {
			out[i] = float32(s / 12.92)
		} else {
			out[i] = float32(math.Pow((s+0.055)/1.055, 2.4))
		}
	}
	return out
}
