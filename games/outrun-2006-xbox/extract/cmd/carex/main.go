// carex extracts OutRun 2006's car models from the disc into glTF binaries.
//
// The chain it reimplements was derived by tracing the game's own loader and
// draw loop (see outrun-2006-xbox.md, the car-model part):
//
//   - every asset is an .sz file: a bare zlib stream (78 DA), streamed by the
//     game through two 64 KB ping-pong buffers into a statically-linked zlib
//     (inflate at VA 0x1BD79B, its caller the stream-reader at 0x2A820);
//   - the inflated *_pmt payload is a 16-byte header {nParts, nTextures, szA,
//     szB} followed by section A (CPU tables: part records, draw batches,
//     materials, the 16-bit index pool, and an XPR0 texture bank header) and
//     section B (GPU data: texture pixels, then vertex buffers);
//   - the loader (0x363xx state machine) reads the header, allocates szA+8,
//     streams both sections, and 0x35DC0 fixes every nonzero offset in the
//     part records up into a pointer (zero stays zero — that fixup is what
//     told us which fields are offsets);
//   - at draw time (0x12070 batch loop) each batch names a primitive type, a
//     first index and a primitive count in the part's index-pool slice, a
//     base vertex, and a 0x58 material whose four 20-byte stages carry the
//     texture indices; SetStreamSource strides and the NV2A vertex
//     declaration (read live off SET_VERTEX_DATA_ARRAY_FORMAT) pin the
//     vertex layouts: stride 16 = {pos f32x3, normal 11:11:10}, stride 24 =
//     {pos f32x3, normal 11:11:10, uv f32x2}.
//
// carex re-derives all of that from the file alone and validates the
// invariants the trace established (section sizes sum to the file, batches
// consume their index slice exactly, base+index stays inside the vertex
// buffer). Textures decode from the XPR0 bank: the Format dword's nibbles are
// {vlog2, ulog2, mips} over an NV2A format byte (0x0C DXT1, 0x0E DXT23, 0x0F
// DXT45), low-byte bit 2 = cube map (six faces, level 0 each).
//
// Usage:
//
//	carex -image DISC.iso -all -o out/cars          # every /Cars model
//	carex -image DISC.iso -file /Cars/obj_plcar_dino_pmt.sz -o out -dumptex
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/curation"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/xbox"
)

func u32(b []byte, off int) uint32  { return binary.LittleEndian.Uint32(b[off:]) }
func u16(b []byte, off int) uint16  { return binary.LittleEndian.Uint16(b[off:]) }
func f32(b []byte, off int) float32 { return math.Float32frombits(u32(b, off)) }

// pmt is one inflated *_pmt payload split into its two sections.
type pmt struct {
	name   string
	nParts int
	nTex   int
	a, b   []byte
}

func parsePMT(name string, data []byte) (*pmt, error) {
	if len(data) < 0x18 {
		return nil, fmt.Errorf("short file (%d bytes)", len(data))
	}
	nParts, nTex := int(u32(data, 0)), int(u32(data, 4))
	szA, szB := int(u32(data, 8)), int(u32(data, 12))
	if 0x10+szA+szB != len(data) {
		return nil, fmt.Errorf("size check failed: 0x10+%#x+%#x != %#x", szA, szB, len(data))
	}
	p := &pmt{name: name, nParts: nParts, nTex: nTex, a: data[0x10 : 0x10+szA], b: data[0x10+szA:]}
	// Section A's own sub-header repeats the counts.
	if int(u32(p.a, 4)) != nParts || int(u32(p.a, 8)) != nTex {
		return nil, fmt.Errorf("sub-header disagrees with file header")
	}
	return p, nil
}

// texInfo is one decoded XPR0 texture bank entry.
type texInfo struct {
	dataOff uint32
	format  uint32
	w, h    int
	mips    int
	fmtByte int
	cube    bool
	img     image.Image // nil if the format is not decoded
}

// parseTextures locates the XPR0 bank header in section A and decodes level 0
// of every texture out of section B (whose head is the bank's pixel data).
func (p *pmt) parseTextures() ([]texInfo, int, error) {
	xi := bytes.Index(p.a, []byte("XPR0"))
	if xi < 0 {
		if p.nTex == 0 {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("no XPR0 bank but nTex=%d", p.nTex)
	}
	total, hdrSize := int(u32(p.a, xi+4)), int(u32(p.a, xi+8))
	// hdrSize is padded (0xAD fill after a 0xFFFFFFFF terminator): the entry
	// count is the file header's nTex; every live entry carries the D3D
	// texture Common word 0x00040001.
	n := p.nTex
	if 12+20*n > hdrSize {
		return nil, 0, fmt.Errorf("XPR bank header too small for %d textures", n)
	}
	for i := 0; i < n; i++ {
		if u32(p.a, xi+12+20*i) != 0x00040001 {
			return nil, 0, fmt.Errorf("XPR entry %d is not a texture header", i)
		}
	}
	pixBytes := total - hdrSize // size of section B's texture region
	texs := make([]texInfo, n)
	for i := range texs {
		e := p.a[xi+12+20*i:]
		t := texInfo{dataOff: u32(e, 4), format: u32(e, 12)}
		t.w = 1 << (t.format >> 20 & 0xF)
		t.h = 1 << (t.format >> 24 & 0xF)
		t.mips = int(t.format >> 16 & 0xF)
		t.fmtByte = int(t.format >> 8 & 0xFF)
		t.cube = t.format&4 != 0
		if int(t.dataOff) < pixBytes {
			t.img = decodeTexture(p.b[t.dataOff:], t.w, t.h, t.fmtByte)
		}
		texs[i] = t
	}
	return texs, pixBytes, nil
}

// decodeTexture decodes one top-level surface. The car set uses the DXT
// family, which the NV2A stores linearly (no swizzle).
func decodeTexture(data []byte, w, h, fmtByte int) image.Image {
	switch fmtByte {
	case 0x0C:
		return decodeDXT(data, w, h, 1)
	case 0x0E:
		return decodeDXT(data, w, h, 23)
	case 0x0F:
		return decodeDXT(data, w, h, 45)
	}
	return nil
}

// decodeDXT is a standard BC1/BC2/BC3 block decoder (variant 1, 23 or 45).
func decodeDXT(data []byte, w, h, variant int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	bw, bh := (w+3)/4, (h+3)/4
	blockLen := 16
	if variant == 1 {
		blockLen = 8
	}
	if len(data) < bw*bh*blockLen {
		return nil
	}
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			block := data[(by*bw+bx)*blockLen:]
			var alpha [16]uint8
			for i := range alpha {
				alpha[i] = 0xFF
			}
			colorBlock := block
			switch variant {
			case 23: // explicit 4-bit alpha
				for i := 0; i < 16; i++ {
					a4 := block[i/2] >> uint(i%2*4) & 0xF
					alpha[i] = a4<<4 | a4
				}
				colorBlock = block[8:]
			case 45: // interpolated alpha
				a0, a1 := block[0], block[1]
				bits := uint64(0)
				for i := 5; i >= 0; i-- {
					bits = bits<<8 | uint64(block[2+i])
				}
				for i := 0; i < 16; i++ {
					code := bits >> uint(3*i) & 7
					switch {
					case code == 0:
						alpha[i] = a0
					case code == 1:
						alpha[i] = a1
					case a0 > a1:
						alpha[i] = uint8(((8-code)*uint64(a0) + (code-1)*uint64(a1)) / 7)
					case code == 6:
						alpha[i] = 0
					case code == 7:
						alpha[i] = 255
					default:
						alpha[i] = uint8(((6-code)*uint64(a0) + (code-1)*uint64(a1)) / 5)
					}
				}
				colorBlock = block[8:]
			}
			c0, c1 := u16(colorBlock, 0), u16(colorBlock, 2)
			var pal [4][3]uint8
			pal[0] = rgb565(c0)
			pal[1] = rgb565(c1)
			opaque4 := variant != 1 || c0 > c1
			if opaque4 {
				for k := 0; k < 3; k++ {
					pal[2][k] = uint8((2*int(pal[0][k]) + int(pal[1][k])) / 3)
					pal[3][k] = uint8((int(pal[0][k]) + 2*int(pal[1][k])) / 3)
				}
			} else {
				for k := 0; k < 3; k++ {
					pal[2][k] = uint8((int(pal[0][k]) + int(pal[1][k])) / 2)
				}
			}
			codes := u32(colorBlock, 4)
			for i := 0; i < 16; i++ {
				x, y := bx*4+i%4, by*4+i/4
				if x >= w || y >= h {
					continue
				}
				ci := codes >> uint(2*i) & 3
				a := alpha[i]
				if variant == 1 && !opaque4 && ci == 3 {
					a = 0
				}
				img.SetNRGBA(x, y, color.NRGBA{pal[ci][0], pal[ci][1], pal[ci][2], a})
			}
		}
	}
	return img
}

func rgb565(v uint16) [3]uint8 {
	r := uint8(v >> 11 & 0x1F)
	g := uint8(v >> 5 & 0x3F)
	b := uint8(v & 0x1F)
	return [3]uint8{r<<3 | r>>2, g<<2 | g>>4, b<<3 | b>>2}
}

// buffers describes one of a part's stream pairs: a slice of the 16-bit index
// pool in section A and a vertex buffer in section B.
type bufPair struct {
	ibOff   uint32 // section-A offset of the index slice (from the w1 descriptor)
	ibBytes uint32
	vbOff   uint32 // section-B offset of the vertex data (from the w2 descriptor)
	vbBytes uint32
	stride  uint32
}

type batch struct {
	prim    uint32 // D3D8 primitive type; 6 = TRIANGLESTRIP
	first   uint32 // first index, relative to the pair's index slice
	prims   uint32 // primitive count
	baseVtx uint32
	matIdx  uint32 // into the part's 0x58 material array
	pair    int
}

type material struct {
	texIdx  int // first bound 2-D texture, -1 if none
	diffuse [3]float32
}

type part struct {
	pairs   []bufPair
	batches []batch
	mats    []material
}

// parsePart reads part record p (15 dwords at A+0x18+p*0x3C) and everything it
// references, validating the invariants the loader trace established.
func (p *pmt) parsePart(pi int) (*part, error) {
	rec := 0x18 + pi*0x3C
	w := make([]uint32, 15)
	for i := range w {
		w[i] = u32(p.a, rec+4*i)
	}
	ent := int(w[5])
	e := make([]uint32, 13)
	for i := range e {
		e[i] = u32(p.a, ent+4*i)
	}
	nPairs, nBatch, nMat48 := int(e[9]), int(e[10]), int(e[12])

	pt := &part{}
	// Stream pairs: sizes+stride from the w12 table (0x2C stride), index-slice
	// offsets from the w1 descriptors, vertex-buffer offsets from the w2
	// descriptors (4 stream slots per pair, slot 0 carries the buffer).
	for k := 0; k < nPairs; k++ {
		t := int(w[12]) + k*0x2C
		bp := bufPair{
			ibBytes: u32(p.a, t+0x18),
			vbBytes: u32(p.a, t+0x1C),
			stride:  u32(p.a, t+0x24),
		}
		ibDesc := u32(p.a, int(w[1])+4*k)
		bp.ibOff = u32(p.a, int(ibDesc)+4)
		vbDesc := u32(p.a, int(w[2])+4*(4*k))
		bp.vbOff = u32(p.a, int(vbDesc)+4)
		if bp.stride != 16 && bp.stride != 24 && bp.stride != 32 {
			return nil, fmt.Errorf("part %d pair %d: unexpected stride %d", pi, k, bp.stride)
		}
		if int(bp.vbOff)+int(bp.vbBytes) > len(p.b) {
			return nil, fmt.Errorf("part %d pair %d: VB out of section B", pi, k)
		}
		if int(bp.ibOff)+int(bp.ibBytes) > len(p.a) {
			return nil, fmt.Errorf("part %d pair %d: IB out of section A", pi, k)
		}
		pt.pairs = append(pt.pairs, bp)
	}

	// Materials: 0x58 entries {id -> 0x48 colour material, stateKey, 4 stages
	// of {state, ?, ?, ?, texIdx}}.
	for mi := 0; mi < nBatch; mi++ {
		m := int(w[13]) + mi*0x58
		id := int(u32(p.a, m))
		mat := material{texIdx: -1, diffuse: [3]float32{1, 1, 1}}
		if id >= 0 && id < nMat48 {
			c := int(w[14]) + id*0x48
			mat.diffuse = [3]float32{f32(p.a, c), f32(p.a, c+4), f32(p.a, c+8)}
		}
		for s := 0; s < 4; s++ {
			ti := u32(p.a, m+8+20*s+16)
			if ti != 0xFFFFFFFF && int(ti) < p.nTex {
				mat.texIdx = int(ti)
				break
			}
		}
		pt.mats = append(pt.mats, mat)
	}

	// Batches: the 32-byte descriptors give {baseVertex, matIdx, drawIdx};
	// drawIdx names the 16-byte draw entry {prim, firstIndex, primCount, nv}.
	// firstIndex is relative to the pair's index slice; the pair advances when
	// the walk has consumed the slice exactly (validated below).
	pair, consumed := 0, uint32(0)
	for bi := 0; bi < nBatch; bi++ {
		d := int(w[10]) + bi*32
		baseVtx, matIdx, drawIdx := u32(p.a, d), u32(p.a, d+4), u32(p.a, d+8)
		if int(drawIdx) >= nBatch {
			return nil, fmt.Errorf("part %d batch %d: drawIdx %d out of range", pi, bi, drawIdx)
		}
		dr := int(w[11]) + int(drawIdx)*16
		b := batch{
			prim: u32(p.a, dr), first: u32(p.a, dr+4), prims: u32(p.a, dr+8),
			baseVtx: baseVtx, matIdx: matIdx,
		}
		idxCount, ok := indexCount(b.prim, b.prims)
		if !ok {
			return nil, fmt.Errorf("part %d batch %d: unhandled primitive %d", pi, bi, b.prim)
		}
		if b.first < consumed && pair+1 < len(pt.pairs) {
			// firstIndex went backwards: the walk crossed into the next slice.
			pair++
			consumed = 0
		}
		if (b.first+idxCount)*2 > pt.pairs[pair].ibBytes+2 {
			return nil, fmt.Errorf("part %d batch %d: indices overrun slice %d (%#x+%d idx > %#x bytes)",
				pi, bi, pair, b.first*2, idxCount, pt.pairs[pair].ibBytes)
		}
		consumed = b.first + idxCount
		b.pair = pair
		pt.batches = append(pt.batches, b)
	}
	return pt, nil
}

// indexCount mirrors the game's own primitive table at VA 0x248758:
// {mul, add} per Kelvin primitive type, idxCount = prims*mul + add.
func indexCount(prim, prims uint32) (uint32, bool) {
	switch prim {
	case 5: // TRIANGLES
		return prims * 3, true
	case 6, 7: // TRISTRIP, TRIFAN
		return prims + 2, true
	case 8: // QUADS
		return prims * 4, true
	case 9: // QUADSTRIP
		return prims*2 + 2, true
	}
	return 0, false
}

// triangulate turns one batch's index run into triangles (degenerates dropped).
func triangulate(prim uint32, raw []uint32) [][3]uint32 {
	var tris [][3]uint32
	emit := func(a, b, c uint32) {
		if a != b && b != c && a != c {
			tris = append(tris, [3]uint32{a, b, c})
		}
	}
	switch prim {
	case 5:
		for i := 0; i+2 < len(raw); i += 3 {
			emit(raw[i], raw[i+1], raw[i+2])
		}
	case 6:
		for i := 0; i+2 < len(raw); i++ {
			a, b, c := raw[i], raw[i+1], raw[i+2]
			if i%2 == 1 {
				a, b = b, a
			}
			emit(a, b, c)
		}
	case 7:
		for i := 1; i+1 < len(raw); i++ {
			emit(raw[0], raw[i], raw[i+1])
		}
	case 8:
		for i := 0; i+3 < len(raw); i += 4 {
			emit(raw[i], raw[i+1], raw[i+2])
			emit(raw[i], raw[i+2], raw[i+3])
		}
	case 9:
		for i := 0; i+3 < len(raw); i += 2 {
			emit(raw[i], raw[i+1], raw[i+2])
			emit(raw[i+2], raw[i+1], raw[i+3])
		}
	}
	return tris
}

// decodeVerts expands one pair's vertex buffer into positions and UVs.
func (p *pmt) decodeVerts(bp bufPair) (pos [][3]float32, uv [][2]float32) {
	n := int(bp.vbBytes / bp.stride)
	pos = make([][3]float32, n)
	uv = make([][2]float32, n)
	for i := 0; i < n; i++ {
		v := int(bp.vbOff) + i*int(bp.stride)
		pos[i] = [3]float32{f32(p.b, v), f32(p.b, v+4), f32(p.b, v+8)}
		if bp.stride >= 24 {
			uv[i] = [2]float32{f32(p.b, v+16), f32(p.b, v+20)}
		}
	}
	return pos, uv
}

// export builds one GLB from every part of the file.
// partPose is a static per-part placement. The model files hold every part in
// its own local space (wheels and doors at the origin) and the game places
// them each frame from its chassis/physics state — there is no rest pose in
// the file. These translations were captured once from the game's own draw
// matrices on the race grid (bootoracle -carvtx: M_rel = M_body⁻¹·M_part,
// view and projection cancel), so an assembled export shows the pose the game
// itself showed. Rotation is dropped (the capture had the wheels mid-spin);
// the steering wheel keeps its 25° column tilt, applied about X.
type partPose struct {
	t     [3]float32
	tiltX float64 // radians, applied about X before translating
}

// grid-rest pose for obj_plcar_dino, captured 2026-07-27 from race-driving.state.
var dinoPose = map[int]partPose{
	5:  {t: [3]float32{-0.657, 0, -0.664}},
	6:  {t: [3]float32{0.657, 0, -0.664}},
	8:  {t: [3]float32{-0.004, 0.273, -0.312}},
	9:  {t: [3]float32{-0.331, 0.688, -0.258}, tiltX: -25 * math.Pi / 180},
	10: {t: [3]float32{-0.713, 0.299, -1.105}},
	11: {t: [3]float32{0.713, 0.299, -1.105}},
	12: {t: [3]float32{-0.715, 0.312, 1.235}},
	13: {t: [3]float32{0.715, 0.312, 1.235}},
	16: {t: [3]float32{0, 0.05, 0}},
}

// poses maps the model name (file base without _pmt.sz) to its captured pose.
var poses = map[string]map[int]partPose{
	"obj_plcar_dino": dinoPose,
}

func export(p *pmt, texs []texInfo, outPath string) (string, error) {
	pose := poses[p.name]
	var positions [][3]float32
	var uvs [][2]float32
	texTris := map[int][][3]uint32{}      // texture index -> triangles
	colorTris := map[[3]int][][3]uint32{} // quantised colour -> triangles

	totalTris := 0
	for pi := 0; pi < p.nParts; pi++ {
		pt, err := p.parsePart(pi)
		if err != nil {
			return "", err
		}
		// Per pair, decode the vertex buffer once and remember its base in the
		// merged arrays.
		vbase := make([]uint32, len(pt.pairs))
		for k, bp := range pt.pairs {
			vbase[k] = uint32(len(positions))
			pos, uv := p.decodeVerts(bp)
			if pp, ok := pose[pi]; ok {
				s, c := math.Sin(pp.tiltX), math.Cos(pp.tiltX)
				for i := range pos {
					y, z := pos[i][1], pos[i][2]
					pos[i][1] = float32(c)*y - float32(s)*z
					pos[i][2] = float32(s)*y + float32(c)*z
					pos[i][0] += pp.t[0]
					pos[i][1] += pp.t[1]
					pos[i][2] += pp.t[2]
				}
			}
			positions = append(positions, pos...)
			uvs = append(uvs, uv...)
		}
		for _, b := range pt.batches {
			bp := pt.pairs[b.pair]
			idxCount, _ := indexCount(b.prim, b.prims)
			raw := make([]uint32, idxCount)
			nVerts := bp.vbBytes / bp.stride
			for i := range raw {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+uint32(i))*2)) + b.baseVtx
				if ix >= nVerts {
					return "", fmt.Errorf("index %d out of VB (%d verts)", ix, nVerts)
				}
				raw[i] = ix + vbase[b.pair]
			}
			tris := triangulate(b.prim, raw)
			totalTris += len(tris)
			m := pt.mats[b.matIdx]
			if m.texIdx >= 0 && texs[m.texIdx].img != nil && !texs[m.texIdx].cube {
				texTris[m.texIdx] = append(texTris[m.texIdx], tris...)
			} else {
				key := [3]int{int(m.diffuse[0] * 255), int(m.diffuse[1] * 255), int(m.diffuse[2] * 255)}
				colorTris[key] = append(colorTris[key], tris...)
			}
		}
	}

	var texGroups []glb.TexturedGroup
	for ti, tris := range texTris {
		texGroups = append(texGroups, glb.TexturedGroup{
			Tris: tris, Image: texs[ti].img, WrapS: 10497, WrapT: 10497,
		})
	}
	var colorGroups []glb.TriGroup
	for key, tris := range colorTris {
		colorGroups = append(colorGroups, glb.TriGroup{
			Tris:  tris,
			Color: [3]float32{float32(key[0]) / 255, float32(key[1]) / 255, float32(key[2]) / 255},
		})
	}
	if err := glb.WriteTextured(outPath, positions, uvs, texGroups, colorGroups); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d parts, %d verts, %d tris, %d textured groups, %d colour groups",
		p.nParts, len(positions), totalTris, len(texGroups), len(colorGroups)), nil
}

func main() {
	imagePath := flag.String("image", "", "Xbox disc image")
	one := flag.String("file", "", "single disc path to extract (e.g. /Cars/obj_plcar_dino_pmt.sz)")
	all := flag.Bool("all", false, "extract every /Cars/*_pmt.sz model")
	outDir := flag.String("o", "out", "output directory")
	dumpTex := flag.Bool("dumptex", false, "also write each decoded texture as PNG")
	site := flag.String("site", "", "Studio export: write the curated roster + manifest.json under this directory")
	flag.Parse()

	if *site != "" {
		if *imagePath == "" {
			fmt.Fprintln(os.Stderr, "usage: carex -image DISC.iso -site site/public/outrun-2006-xbox")
			os.Exit(2)
		}
		exportSite(*imagePath, *site)
		return
	}
	if *imagePath == "" || (*one == "" && !*all) {
		fmt.Fprintln(os.Stderr, "usage: carex -image DISC.iso (-file /Cars/... | -all) [-o dir]")
		os.Exit(2)
	}
	disc, err := xbox.Open(*imagePath)
	if err != nil {
		fatal("open image: %v", err)
	}
	defer disc.Close()

	var files []string
	if *one != "" {
		files = []string{*one}
	} else {
		if err := disc.Walk(func(e xbox.Entry) error {
			if !e.IsDir && strings.HasPrefix(strings.ToLower(e.Path), "/cars/") &&
				strings.HasSuffix(strings.ToLower(e.Path), "_pmt.sz") {
				files = append(files, e.Path)
			}
			return nil
		}); err != nil {
			fatal("walk: %v", err)
		}
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal("%v", err)
	}

	for _, f := range files {
		raw, err := disc.ReadFile(f)
		if err != nil {
			fatal("read %s: %v", f, err)
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			fatal("%s: zlib: %v", f, err)
		}
		data, err := io.ReadAll(zr)
		if err != nil {
			fatal("%s: inflate: %v", f, err)
		}
		base := strings.TrimSuffix(filepath.Base(f), "_pmt.sz")
		p, err := parsePMT(base, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "carex: %s: %v (skipped)\n", f, err)
			continue
		}
		texs, pixBytes, err := p.parseTextures()
		if err != nil {
			fmt.Fprintf(os.Stderr, "carex: %s: %v (skipped)\n", f, err)
			continue
		}
		if *dumpTex {
			for i, t := range texs {
				if t.img == nil {
					fmt.Printf("  tex%02d fmt=%08x: undecoded\n", i, t.format)
					continue
				}
				writePNG(filepath.Join(*outDir, fmt.Sprintf("%s_tex%02d.png", base, i)), t.img)
			}
		}
		out := filepath.Join(*outDir, base+".glb")
		summary, err := export(p, texs, out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "carex: %s: %v (skipped)\n", f, err)
			continue
		}
		fmt.Printf("%-28s -> %s (%s; %d textures, %#x pixel bytes)\n", f, out, summary, len(texs), pixBytes)
	}
}

// exportSite writes the Studio's curated car set: the assembled player Dino
// plus the fifteen AI ("rc") rivals, whose wheels are modeled in place. The
// unassembled player models, the `_t` variants, the origin-stacked traffic
// bundle and rc_all stay out until their open items (chassis rest pose,
// variant semantics) are resolved — see the game markdown, Part XIX.
func exportSite(imagePath, siteDir string) {
	disc, err := xbox.Open(imagePath)
	if err != nil {
		fatal("open image: %v", err)
	}
	defer disc.Close()
	// Retro-X: the curated roster becomes model3d object assets under a
	// builder-written tree (validated on Write).
	b := build.New(siteDir, "outrun-2006-xbox")
	b.SetTitle("OutRun 2006: Coast 2 Coast")
	b.SetPlatform("Original Xbox")
	b.SetYear(2006)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 640, H: 480},
		TickHz: 60,
		// The NV2A filters textures bilinearly.
		TexFilter: "linear",
	})
	if cur, err := curation.Load("curation"); err == nil {
		if err := b.ApplyCuration(cur); err != nil {
			fatal("curation: %v", err)
		}
	}

	// Model codes are the disc's own file names; the labels just uppercase them.
	rivals := []string{
		"250gto", "328gts", "360sp", "512bb", "550b", "575sa",
		"dayts", "dino", "f355sp", "f40", "f430", "f50", "fx", "gto", "testa",
	}

	doOne := func(discPath, outName, name, section string) {
		raw, err := disc.ReadFile(discPath)
		if err != nil {
			fatal("read %s: %v", discPath, err)
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			fatal("%s: zlib: %v", discPath, err)
		}
		data, err := io.ReadAll(zr)
		if err != nil {
			fatal("%s: inflate: %v", discPath, err)
		}
		base := strings.TrimSuffix(filepath.Base(discPath), "_pmt.sz")
		p, err := parsePMT(base, data)
		if err != nil {
			fatal("%s: %v", discPath, err)
		}
		texs, _, err := p.parseTextures()
		if err != nil {
			fatal("%s: %v", discPath, err)
		}
		out, err := b.Path("objects", outName)
		if err != nil {
			fatal("%v", err)
		}
		summary, err := export(p, texs, out)
		if err != nil {
			fatal("%s: %v", discPath, err)
		}
		id := strings.TrimSuffix(outName, ".glb")
		b.AddObject(schema.Asset{ID: id, Name: name, Group: section}, &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: outName,
			Props: map[string]any{"source": discPath},
		})
		fmt.Printf("%-34s -> %s (%s)\n", discPath, out, summary)
	}

	doOne("/Cars/obj_plcar_dino_pmt.sz", "plcar-dino.glb", "Dino 246 GTS (player)", "Player car")
	for _, code := range rivals {
		doOne("/Cars/obj_rc_"+code+"_pmt.sz", "rc-"+code+".glb",
			strings.ToUpper(code), "Rivals (AI)")
	}

	if err := b.Write(); err != nil {
		fatal("%v", err)
	}
	for _, w := range b.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fatal("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fatal("%v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "carex: "+format+"\n", args...)
	os.Exit(1)
}
