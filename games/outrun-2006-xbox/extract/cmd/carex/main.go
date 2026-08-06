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
	"sort"
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
	cube   []bool // per texture-bank entry: cube-map flag (set by parseTextures)
	// vis, when set, is the course's per-segment visibility database (the
	// cs_*_bin sibling); export then takes the placement-forest path
	// (stage.go) instead of the flat all-batches one.
	vis *visBin
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
	p.cube = make([]bool, len(texs))
	for i, t := range texs {
		p.cube[i] = t.cube
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
	// fmtWord is the pair table's +0x20 vertex-format code, a bitfield the
	// stage files exercise fully (cars only ever carry 0x012/0x112/0x212):
	// low bits 0x12 = {pos f32x3 @0, normal packed-11:11:10 @12}; bit 0x40 =
	// a D3DCOLOR at 16; bits 0x300 = UV-set count, f32x2 each, after the
	// colour. fmtWord 0 with stride 44 is the reflective-water special:
	// {pos, normal, uv0 @16, uv1 @24, tangent f32x3 @32 on the tex2 slot} —
	// the tangent feeds the texm3x3 basis of the sea's per-pixel cube-map
	// reflection (bootoracle -vshdump 11 + cmd/tanprobe). Every layout read
	// live off SET_VERTEX_DATA_ARRAY_FORMAT/_OFFSET (-vtxdecl on the beach
	// race).
	fmtWord uint32
}

// hasColor / uvSets decode fmtWord (see above).
func (bp bufPair) hasColor() bool { return bp.fmtWord&0x40 != 0 }
func (bp bufPair) uvSets() int {
	if bp.fmtWord == 0 && bp.stride == 44 {
		return 2
	}
	return int(bp.fmtWord >> 8 & 3)
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
	alpha   float32 // D3DMATERIAL8 diffuse alpha — the glass batches carry < 1
	// additive: the stateKey's blend-factor byte (bits 8-15, indexing the
	// game's own factor table at VA 0x248B54, applied by the state routine
	// at 0x15B90 to NV2A methods 0x0344/0x0348) reads SRC_ALPHA/ONE — the
	// glow overlays. sheen: a stage samples an environment cube (the game
	// adds the reflection; the viewer applies the model's static cube).
	additive bool
	sheen    bool
	// wrapS/wrapT are glTF sampler wrap enums, decoded from the binding
	// stage's first word: bits 10-12 = U mode, 13-15 = V mode, in the NV2A
	// address enum (1 wrap, 2 mirror, 3 clamp). The LOD bodies map one
	// mirrored atlas half across the car and need MIRRORED_REPEAT.
	wrapS, wrapT int
}

// glWrap maps an NV2A texture-address mode to the glTF sampler enum.
func glWrap(mode uint32) int {
	switch mode {
	case 2:
		return 33648 // MIRRORED_REPEAT
	case 3, 4, 5:
		return 33071 // CLAMP_TO_EDGE
	default:
		return 10497 // REPEAT
	}
}

type part struct {
	pairs   []bufPair
	batches []batch
	mats    []material
	// descStart/descCount map a 32-byte batch descriptor index to its run of
	// expanded entries in batches — the unit the placement forest's e2 ranges
	// address.
	descStart, descCount []int
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
	// Entry counts: e[9]=pairs, e[10]=batches, e[11]=0x58 materials,
	// e[12]=0x48 colour materials. The material count used to be read as
	// nBatch — the cars never had more batches than materials, so the
	// over-read stayed inside section A; the stage files (229 batches, 17
	// materials) put the real count field beyond doubt.
	nPairs, nBatch, nMat58, nMat48 := int(e[9]), int(e[10]), int(e[11]), int(e[12])

	pt := &part{}
	// Stream pairs: sizes+stride from the w12 table (0x2C stride), index-slice
	// offsets from the w1 descriptors, vertex-buffer offsets from the w2
	// descriptors (4 stream slots per pair, slot 0 carries the buffer).
	for k := 0; k < nPairs; k++ {
		t := int(w[12]) + k*0x2C
		bp := bufPair{
			ibBytes: u32(p.a, t+0x18),
			vbBytes: u32(p.a, t+0x1C),
			fmtWord: u32(p.a, t+0x20),
			stride:  u32(p.a, t+0x24),
		}
		ibDesc := u32(p.a, int(w[1])+4*k)
		bp.ibOff = u32(p.a, int(ibDesc)+4)
		vbDesc := u32(p.a, int(w[2])+4*(4*k))
		bp.vbOff = u32(p.a, int(vbDesc)+4)
		// The stride the format promises must be the stride the table carries.
		want := uint32(16 + 8*bp.uvSets())
		if bp.hasColor() {
			want += 4
		}
		if bp.fmtWord == 0 && bp.stride == 44 {
			want = 44 // the course-geometry special: uv0, uv1, f32x3 @32
		}
		if bp.stride != want {
			return nil, fmt.Errorf("part %d pair %d: fmt %#x promises stride %d, table says %d",
				pi, k, bp.fmtWord, want, bp.stride)
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
	for mi := 0; mi < nMat58; mi++ {
		m := int(w[13]) + mi*0x58
		id := int(u32(p.a, m))
		mat := material{texIdx: -1, diffuse: [3]float32{1, 1, 1}, alpha: 1}
		if id >= 0 && id < nMat48 {
			c := int(w[14]) + id*0x48
			mat.diffuse = [3]float32{f32(p.a, c), f32(p.a, c+4), f32(p.a, c+8)}
			mat.alpha = f32(p.a, c+12)
		}
		stateKey := u32(p.a, m+4)
		mat.additive = stateKey>>8&0xFF == 0x14 // SRC_ALPHA / ONE
		for s := 0; s < 4; s++ {
			ti := u32(p.a, m+8+20*s+16)
			if ti == 0xFFFFFFFF || int(ti) >= p.nTex {
				continue
			}
			if len(p.cube) > int(ti) && p.cube[ti] {
				mat.sheen = true
				continue
			}
			if mat.texIdx < 0 {
				mat.texIdx = int(ti)
				w0 := u32(p.a, m+8+20*s)
				mat.wrapS, mat.wrapT = glWrap(w0>>10&7), glWrap(w0>>13&7)
			}
		}
		pt.mats = append(pt.mats, mat)
	}

	// Batches: the 32-byte descriptors give {baseVertex, matIdx,
	// firstDrawIdx, drawCount, sphere×4}; each of the drawCount
	// consecutive 16-byte draw entries {prim, firstIndex, primCount, nv}
	// is one strip, all sharing the batch's material and base vertex (the
	// cars kept drawCount at 1, which let the old reader treat word3 as
	// padding). firstIndex is relative to the pair's index slice, and each
	// pair's draws tile its slice in file order — but the stream
	// interleaves the pairs freely (the beach course's part 1 weaves four
	// pairs through 125 batches; the cars kept it monotonic, which is why
	// a simple pair-advance walk used to pass). The stream structure is a
	// set of cursors: a draw extends the cursor sitting exactly at its
	// firstIndex, and firstIndex 0 opens a new cursor. Which cursor is
	// which pair is decided at the end, where each cursor must have
	// consumed some pair's slice exactly — that bijection is the
	// invariant, and index-bounds checks break size ties.
	for bi := 0; bi < nBatch; bi++ {
		d := int(w[10]) + bi*32
		baseVtx, matIdx := u32(p.a, d), u32(p.a, d+4)
		drawIdx, drawCount := u32(p.a, d+8), u32(p.a, d+12)
		pt.descStart = append(pt.descStart, len(pt.batches))
		pt.descCount = append(pt.descCount, int(drawCount))
		if drawCount == 0 || drawCount > 64 {
			return nil, fmt.Errorf("part %d batch %d: draw count %d out of range", pi, bi, drawCount)
		}
		if int(matIdx) >= nMat58 {
			return nil, fmt.Errorf("part %d batch %d: matIdx %d out of range (%d materials)", pi, bi, matIdx, nMat58)
		}
		for di := uint32(0); di < drawCount; di++ {
			dr := int(w[11]) + int(drawIdx+di)*16
			b := batch{
				prim: u32(p.a, dr), first: u32(p.a, dr+4), prims: u32(p.a, dr+8),
				baseVtx: baseVtx, matIdx: matIdx, pair: -1,
			}
			if _, ok := indexCount(b.prim, b.prims); !ok {
				return nil, fmt.Errorf("part %d batch %d: unhandled primitive %d", pi, bi, b.prim)
			}
			pt.batches = append(pt.batches, b)
		}
	}
	if err := p.assignPairs(pt, pi); err != nil {
		return nil, err
	}
	return pt, nil
}

// assignPairs resolves which pair each draw reads. A draw extends the
// cursor sitting exactly at its firstIndex, and firstIndex 0 opens a new
// cursor; at the end each cursor must have consumed some pair's slice
// exactly (±1 index of tail padding), one cursor per pair, every index in
// bounds. Two cursors can collide on the same value (two pairs both paused
// 11 indices in — obj_rc_512bb part 23 does it), so collisions are search
// branches: the bijection at the end is what makes the assignment right,
// and it is unique in every file on the disc.
func (p *pmt) assignPairs(pt *part, pi int) error {
	n := len(pt.batches)
	assign := make([]int, n)
	var cursors []uint32
	streamBatches := func(si int) []int {
		var out []int
		for i := 0; i < n; i++ {
			if assign[i] == si {
				out = append(out, i)
			}
		}
		return out
	}
	fits := func(si, k int) bool {
		bp := pt.pairs[k]
		nVerts := bp.vbBytes / bp.stride
		for _, bi := range streamBatches(si) {
			b := pt.batches[bi]
			idxCount, _ := indexCount(b.prim, b.prims)
			for i := uint32(0); i < idxCount; i++ {
				if uint32(u16(p.a, int(bp.ibOff)+int(b.first+i)*2))+b.baseVtx >= nVerts {
					return false
				}
			}
		}
		return true
	}
	sizeOK := func(c uint32, ibBytes uint32) bool {
		return c*2 == ibBytes || c*2 == ibBytes-2
	}
	var finalize func() bool
	finalize = func() bool {
		taken := make([]bool, len(pt.pairs))
		var match func(si int) bool
		match = func(si int) bool {
			if si == len(cursors) {
				for k, bp := range pt.pairs {
					if !taken[k] && bp.ibBytes > 2 {
						return false
					}
				}
				return true
			}
			for k := range pt.pairs {
				if !taken[k] && sizeOK(cursors[si], pt.pairs[k].ibBytes) && fits(si, k) {
					taken[k] = true
					for _, bi := range streamBatches(si) {
						pt.batches[bi].pair = k
					}
					if match(si + 1) {
						return true
					}
					taken[k] = false
				}
			}
			return false
		}
		return match(0)
	}
	var dfs func(i int) bool
	dfs = func(i int) bool {
		if i == n {
			return finalize()
		}
		b := pt.batches[i]
		idxCount, _ := indexCount(b.prim, b.prims)
		for si, c := range cursors {
			if c == b.first {
				cursors[si] = b.first + idxCount
				assign[i] = si
				if dfs(i + 1) {
					return true
				}
				cursors[si] = c
			}
		}
		if b.first == 0 && len(cursors) < len(pt.pairs) {
			cursors = append(cursors, idxCount)
			assign[i] = len(cursors) - 1
			if dfs(i + 1) {
				return true
			}
			cursors = cursors[:len(cursors)-1]
		}
		return false
	}
	if !dfs(0) {
		return fmt.Errorf("part %d: no draw-to-pair assignment satisfies the slice invariants", pi)
	}
	return nil
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

// decodeVerts expands one pair's vertex buffer into positions, normals, UVs,
// the baked D3DCOLOR and the second UV set. Every layout carries the NV2A's
// packed 11:11:10 normal at +12; the D3DCOLOR (when fmtWord has it) sits at
// +16 — stored BGRA in memory, presented RGBA by the UB_D3D attribute type —
// and pushes the UV sets to +20 (-vtxdecl census on the beach race). col and
// uv2 are nil when the layout lacks them. The stride-44 water layout's
// trailing f32x3 is a tangent (see bufPair.fmtWord); glTF TANGENT wants a
// normalised vec4 with handedness and these are neither, so it stays
// unexported — the viewer has no per-pixel water reflection to feed anyway.
func (p *pmt) decodeVerts(bp bufPair) (pos, nrm [][3]float32, uv, uv2 [][2]float32, col [][4]uint8) {
	n := int(bp.vbBytes / bp.stride)
	pos = make([][3]float32, n)
	nrm = make([][3]float32, n)
	uv = make([][2]float32, n)
	uvOff := 16
	if bp.hasColor() {
		uvOff = 20
		col = make([][4]uint8, n)
	}
	if bp.uvSets() > 1 {
		uv2 = make([][2]float32, n)
	}
	for i := 0; i < n; i++ {
		v := int(bp.vbOff) + i*int(bp.stride)
		pos[i] = [3]float32{f32(p.b, v), f32(p.b, v+4), f32(p.b, v+8)}
		nrm[i] = unpackNormal(u32(p.b, v+12))
		if col != nil {
			col[i] = [4]uint8{p.b[v+18], p.b[v+17], p.b[v+16], p.b[v+19]}
		}
		if bp.uvSets() > 0 {
			uv[i] = [2]float32{f32(p.b, v+uvOff), f32(p.b, v+uvOff+4)}
		}
		if uv2 != nil {
			uv2[i] = [2]float32{f32(p.b, v+uvOff+8), f32(p.b, v+uvOff+12)}
		}
	}
	return pos, nrm, uv, uv2, col
}

// unpackNormal decodes the NV2A CMP (11:11:10 signed) vertex normal: x in bits
// 0-10, y in 11-21 (11-bit two's complement, /1023), z in 22-31 (10-bit, /511).
func unpackNormal(v uint32) [3]float32 {
	sext := func(u uint32, bits int) float32 {
		s := int32(u<<(32-bits)) >> (32 - bits)
		return float32(s) / float32(int32(1)<<(bits-1)-1)
	}
	x := sext(v&0x7FF, 11)
	y := sext(v>>11&0x7FF, 11)
	z := sext(v>>22&0x3FF, 10)
	// Normalise away the quantisation (and clamp the -1024/-512 endpoints).
	l := float32(math.Sqrt(float64(x*x + y*y + z*z)))
	if l > 0 {
		x, y, z = x/l, y/l, z/l
	}
	return [3]float32{x, y, z}
}

// inspect prints every part's header, batches and materials — the fields the
// export consumes plus the ones it currently drops (stateKey, diffuse alpha,
// the part-header kind) — with a computed bounding box per batch.
func inspect(p *pmt, texs []texInfo) error {
	for pi := 0; pi < p.nParts; pi++ {
		rec := 0x18 + pi*0x3C
		w := make([]uint32, 15)
		for i := range w {
			w[i] = u32(p.a, rec+4*i)
		}
		pt, err := p.parsePart(pi)
		if err != nil {
			return err
		}
		hdr := int(w[7])
		fmt.Printf("part %2d: kind=%#x centre=(%.3f,%.3f,%.3f) r=%.3f pairs=%d batches=%d\n",
			pi, u32(p.a, hdr), f32(p.a, hdr+4), f32(p.a, hdr+8), f32(p.a, hdr+12), f32(p.a, hdr+16),
			len(pt.pairs), len(pt.batches))
		for k, bp := range pt.pairs {
			fmt.Printf("  pair %d: vbOff=%#x vbBytes=%#x stride=%d\n", k, bp.vbOff, bp.vbBytes, bp.stride)
		}
		for mi := range pt.mats {
			m := int(w[13]) + mi*0x58
			id := int(u32(p.a, m))
			stateKey := u32(p.a, m+4)
			var stages []string
			for s := 0; s < 4; s++ {
				ti := u32(p.a, m+8+20*s+16)
				switch {
				case ti == 0xFFFFFFFF:
				case int(ti) < len(texs) && texs[ti].cube:
					stages = append(stages, fmt.Sprintf("s%d=tex%d(cube)", s, ti))
				default:
					stages = append(stages, fmt.Sprintf("s%d=tex%d", s, ti))
				}
			}
			line := fmt.Sprintf("  mat %2d: id=%d stateKey=%08x %s", mi, id, stateKey, strings.Join(stages, " "))
			if id >= 0 && id < 1<<16 {
				c := int(w[14]) + id*0x48
				line += fmt.Sprintf(" diffuse=(%.3f,%.3f,%.3f,a=%.3f)",
					f32(p.a, c), f32(p.a, c+4), f32(p.a, c+8), f32(p.a, c+12))
			}
			fmt.Println(line)
		}
		for bi, b := range pt.batches {
			bp := pt.pairs[b.pair]
			idxCount, _ := indexCount(b.prim, b.prims)
			mn := [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
			mx := [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
			for i := uint32(0); i < idxCount; i++ {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+i)*2)) + b.baseVtx
				v := int(bp.vbOff) + int(ix)*int(bp.stride)
				for c := 0; c < 3; c++ {
					f := f32(p.b, v+4*c)
					if f < mn[c] {
						mn[c] = f
					}
					if f > mx[c] {
						mx[c] = f
					}
				}
			}
			d := int(w[10]) + bi*32
			flag := u32(p.a, d+12)
			fmt.Printf("  batch %2d: pair=%d stride=%d prim=%d idx=%d mat=%d flag=%#x base=%d attrOff=%#x bbox=(%.2f,%.2f,%.2f)..(%.2f,%.2f,%.2f)\n",
				bi, b.pair, bp.stride, b.prim, idxCount, b.matIdx, flag,
				b.baseVtx, bp.vbOff+b.baseVtx*bp.stride,
				mn[0], mn[1], mn[2], mx[0], mx[1], mx[2])
		}
	}
	return nil
}

// onlyParts, when non-nil, restricts export to the listed part indices — the
// -parts diagnostic for comparing candidate part sets against game captures.
var onlyParts map[int]bool

// ---- the chassis table --------------------------------------------------------
//
// The rc cars' wheel placements are static data in default.xbe: per car and per
// LOD group, four {partHandle, x, y, z} wheel entries (plus the group's part
// handles, mirror points and the caster slot). The handle encodes
// (modelId<<16)|partIndex; the table was found by value-searching the wheel
// positions out of the game's runtime group descriptors and confirmed against
// the live race (dayts = id 0x82 matches its captured draw matrices to a
// millimetre, dino = 0x83 matches the captured grid pose exactly) — and
// against reality: every row's wheelbase is the real car's to a centimetre
// (Dino 2.34 m, Daytona 2.40, F40 2.45, Enzo 2.65, Testarossa 2.55, ...).

// wheelSpec is one placed wheel: the part index and its translation. The part
// geometry is modeled as the LEFT wheel; entries with x > 0 are mirrored
// (capture-pinned convention).
type wheelSpec struct {
	part int
	t    [3]float32
}

// chassisWheels maps model name -> [LOD0 wheels, LOD1 wheels], parsed from the
// XBE. Records are found by signature, not offset: a LOD0 group opens with
// handles {id:1, id:0, id:6, ...} and -1 at +0x18, a LOD1 group with
// {id:0xA, id:0, id:0xE, ...}; the four wheel entries sit at +0x38.
// loadChassis reads default.xbe once and fills both chassis tables.
func loadChassis(disc *xbox.Image) error {
	xbe, err := disc.ReadFile("/default.xbe")
	if err != nil {
		return err
	}
	if chassis, err = chassisWheels(xbe); err != nil {
		return err
	}
	if plcars, err = plcarChassis(xbe); err != nil {
		return err
	}
	if traffic, err = trafficChassis(xbe); err != nil {
		return err
	}
	return nil
}

func chassisWheels(xbe []byte) (map[string][2][]wheelSpec, error) {
	// Model id -> rc file base name. Ids 0x80..0x89 are the original OutRun2
	// ten, 0x1D5..0x1D9 the Coast-2-Coast five, each in reverse order of the
	// XBE's name-string blocks; 0x1DA..0x1E8 repeat all 15 for the _t twins.
	// Pinned live: 0x82 = dayts, 0x83 = dino; corroborated by per-car
	// wheelbases (see the writeup, Part XXI).
	ids := map[uint32]string{
		0x80: "obj_rc_250gto", 0x81: "obj_rc_360sp", 0x82: "obj_rc_dayts",
		0x83: "obj_rc_dino", 0x84: "obj_rc_fx", 0x85: "obj_rc_512bb",
		0x86: "obj_rc_f40", 0x87: "obj_rc_f50", 0x88: "obj_rc_gto",
		0x89: "obj_rc_testa",
		0x1D5: "obj_rc_f355sp", 0x1D6: "obj_rc_328gts", 0x1D7: "obj_rc_f430",
		0x1D8: "obj_rc_550b", 0x1D9: "obj_rc_575sa",
	}
	// The _t twins repeat all fifteen at 0x1DA..0x1E8 in the same order
	// (their rows carry byte-identical wheel data to the base cars').
	for i, base := range []uint32{0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x1D5, 0x1D6, 0x1D7, 0x1D8, 0x1D9} {
		ids[0x1DA+uint32(i)] = ids[base] + "_t"
	}
	out := map[string][2][]wheelSpec{}
	for off := 0; off+0x78 <= len(xbe); off += 4 {
		v0 := u32(xbe, off)
		id := v0 >> 16
		name, known := ids[id]
		if !known || id == 0 {
			continue
		}
		slot := -1
		switch v0 & 0xFFFF {
		case 0x1:
			slot = 0 // LOD0 group {1, 0, 6, ...}
		case 0xA:
			slot = 1 // LOD1 group {0xA, 0, 0xE, ...}
		default:
			continue
		}
		if u32(xbe, off+4) != id<<16 || u32(xbe, off+0x18) != 0xFFFFFFFF {
			continue
		}
		if g := u32(xbe, off+8); g != id<<16|0x6 && g != id<<16|0xE {
			continue
		}
		var wheels []wheelSpec
		for k := 0; k < 4; k++ {
			w := off + 0x38 + k*16
			h := u32(xbe, w)
			if h>>16 != id {
				return nil, fmt.Errorf("chassis %s: wheel %d handle %#x not model %#x", name, k, h, id)
			}
			wheels = append(wheels, wheelSpec{
				part: int(h & 0xFFFF),
				t:    [3]float32{f32(xbe, w+4), f32(xbe, w+8), f32(xbe, w+12)},
			})
		}
		cur := out[name]
		if cur[slot] != nil {
			return nil, fmt.Errorf("chassis %s: duplicate group %d", name, slot)
		}
		cur[slot] = wheels
		out[name] = cur
	}
	for id, name := range ids {
		if out[name][0] == nil || out[name][1] == nil {
			return nil, fmt.Errorf("chassis: incomplete groups for %s (id %#x)", name, id)
		}
	}
	return out, nil
}

// chassis is the parsed table, loaded once per disc open.
var chassis map[string][2][]wheelSpec

// trafficVehicle is one traffic type from the traffic table: two LOD groups
// (body + four placed wheels each), the additive glow parts, the caster hull
// and the vehicle's part range within obj_othcar. The ground shadow slot
// points into a separate shared shadow model and is not exported.
type trafficVehicle struct {
	bodies [2]int
	wheels [2][]placement
	glows  []int
	caster int
	lo, hi int
}

// traffic is the parsed traffic table (12 vehicles).
var traffic []trafficVehicle

// trafficChassis finds the traffic records in the XBE by signature: a record
// opens with an othcar handle (model 0xE — its registry slot, pinned via the
// model-object array against dayts=0x82 and the player dino=2), carries eight
// role slots {body, shadow(0xBA:n), beam glow, -, glass glow x2, flare L/R},
// four wheel entries {handle, x, y, z} at +0x38 and the caster at +0x98;
// records pair into vehicles by their shared caster part.
func trafficChassis(xbe []byte) ([]trafficVehicle, error) {
	const othID = 0xE
	type rec struct {
		body, glow, glassA, glassB, flareL, flareR, caster int
		wheels                                             []placement
	}
	part := func(off int) int {
		v := u32(xbe, off)
		if v == 0xFFFFFFFF || v>>16 != othID || v&0xFFFF >= 0x1BD {
			return -1
		}
		return int(v & 0xFFFF)
	}
	var recs []rec
	for off := 0; off+0xA8 <= len(xbe); off += 4 {
		if part(off) < 0 {
			continue
		}
		ok := true
		var wheels []placement
		for k := 0; k < 4 && ok; k++ {
			w := off + 0x38 + k*16
			pi := part(w)
			t := [3]float32{f32(xbe, w+4), f32(xbe, w+8), f32(xbe, w+12)}
			ax, ay, az := t[0], t[1], t[2]
			if ax < 0 {
				ax = -ax
			}
			if az < 0 {
				az = -az
			}
			if pi < 0 || ax < 0.3 || ax > 2 || ay < 0.1 || ay > 1.2 || az < 0.5 || az > 6 {
				ok = false
				break
			}
			wheels = append(wheels, placement{part: pi, mirrorX: t[0] > 0, t: t, label: cornerLabel(t)})
		}
		if !ok || part(off+0x98) < 0 {
			continue
		}
		r := rec{body: part(off), glow: part(off + 8), glassA: part(off + 0x10), glassB: part(off + 0x14),
			flareL: part(off + 0x18), flareR: part(off + 0x1C), caster: part(off + 0x98), wheels: wheels}
		recs = append(recs, r)
		off += 0xA8 - 4
	}
	byCaster := map[int][]rec{}
	var order []int
	for _, r := range recs {
		if len(byCaster[r.caster]) == 0 {
			order = append(order, r.caster)
		}
		byCaster[r.caster] = append(byCaster[r.caster], r)
	}
	var out []trafficVehicle
	for _, c := range order {
		pair := byCaster[c]
		if len(pair) != 2 {
			return nil, fmt.Errorf("traffic: caster part %d has %d records, want 2", c, len(pair))
		}
		v := trafficVehicle{caster: c}
		for i, r := range pair {
			v.bodies[i] = r.body
			v.wheels[i] = r.wheels
			for _, g := range []int{r.glow, r.glassA, r.glassB, r.flareL, r.flareR} {
				if g >= 0 {
					v.glows = append(v.glows, g)
				}
			}
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].bodies[0] < out[j].bodies[0] })
	for i := range out {
		out[i].lo = out[i].bodies[0]
		if i+1 < len(out) {
			out[i].hi = out[i+1].bodies[0]
		} else {
			out[i].hi = 0x1BD
		}
	}
	if len(out) != 12 {
		return nil, fmt.Errorf("traffic: found %d vehicles, want 12", len(out))
	}
	return out, nil
}

// plcarSpec is one player car's 0x128-byte chassis record: the full rest pose
// the Part XIX capture could only sample for the Dino. Layout (offsets from
// the record head): +0 body, +4 five static slots (shadow + shells), +0x18
// gear {handle, xyz}, +0x28/+0x38 doors L/R {handle, xyz}, +0x48 steering
// {handle, xyz, tiltX, tiltB}, +0x60 two glow slots, +0x7C three front-state
// slots, +0x88 four wheel corners {up to four part handles, xyz} (the big
// cars stack tire/rim/disc parts per corner), +0x108 two attach points.
type plcarSpec struct {
	statics []int       // body first, then the four shadow/shell slots (+0x04..+0x10)
	panel   int         // the +0x14 slot: an extra body panel the renderer draws
	// unconditionally with the body (the F40's engine cover under the rear
	// spoiler, the F50's tail panel); -1 when absent (most cars).
	attach  []placement // gear, doors, steering (tilt applied about X)
	glows   []int
	fronts  []int
	wheels  []placement
}

// plcars is the parsed player-car table, keyed by model name.
var plcars map[string]plcarSpec

// plcarChassis finds every player-car record in the XBE by signature (four
// same-id wheel rows at +0x88.., doors/steering same-id or empty) and decodes
// the ones whose model id is mapped to a file.
//
// The id→file mapping: ids 1..0xB are the OutRun2 ten (plus f355sp), ids
// 0x14A..0x14D the Coast-2-Coast four, 0x14E..0x15C the _t twins. It was
// established by matching each record's four wheel positions against the
// already-pinned rc rows (11 of 15 match their rc counterpart exactly, the
// rest to a few cm), and every record's highest referenced part index equals
// its file's part count minus one — 15 of 15 (writeup, Part XXII).
func plcarChassis(xbe []byte) (map[string]plcarSpec, error) {
	ids := map[uint32]string{
		0x1: "obj_plcar_f50", 0x2: "obj_plcar_dino", 0x3: "obj_plcar_gto",
		0x4: "obj_plcar_fx", 0x5: "obj_plcar_dayts", 0x6: "obj_plcar_f355sp",
		0x7: "obj_plcar_testa", 0x8: "obj_plcar_360sp", 0x9: "obj_plcar_f40",
		0xA: "obj_plcar_512bb", 0xB: "obj_plcar_250gto",
		0x14A: "obj_plcar_328gts", 0x14B: "obj_plcar_f430",
		0x14C: "obj_plcar_550b", 0x14D: "obj_plcar_575sa",
		0x14E: "obj_plcar_f50_t", 0x14F: "obj_plcar_dino_t", 0x150: "obj_plcar_gto_t",
		0x151: "obj_plcar_fx_t", 0x152: "obj_plcar_dayts_t", 0x153: "obj_plcar_f355sp_t",
		0x154: "obj_plcar_testa_t", 0x155: "obj_plcar_360m_t", 0x156: "obj_plcar_f40_t",
		0x157: "obj_plcar_512bb_t", 0x158: "obj_plcar_250gto_t",
		0x159: "obj_plcar_328gts_t", 0x15A: "obj_plcar_f430sp_t",
		0x15B: "obj_plcar_550b_t", 0x15C: "obj_plcar_575sa_t",
	}
	part := func(off int, id uint32) (int, bool) {
		v := u32(xbe, off)
		if v == 0xFFFFFFFF || v>>16 != id || v&0xFFFF > 0x3F {
			return -1, false
		}
		return int(v & 0xFFFF), true
	}
	out := map[string]plcarSpec{}
	for off := 0; off+0x128 <= len(xbe); off += 4 {
		v := u32(xbe, off)
		id := v >> 16
		name, known := ids[id]
		if !known || v&0xFFFF != 0 {
			continue
		}
		ok := true
		for _, w := range []int{0x88, 0xA8, 0xC8, 0xE8} {
			if _, valid := part(off+w, id); !valid {
				ok = false
				break
			}
		}
		for _, a := range []int{0x28, 0x38, 0x48} {
			if v := u32(xbe, off+a); v != 0xFFFFFFFF && (v>>16 != id || v&0xFFFF > 0x3F) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		s := plcarSpec{panel: -1}
		s.statics = append(s.statics, 0)
		for i := 0; i < 4; i++ {
			if pi, valid := part(off+4+4*i, id); valid {
				s.statics = append(s.statics, pi)
			}
		}
		if pi, valid := part(off+0x18-4, id); valid {
			s.panel = pi
		}
		for i, a := range []int{0x18, 0x28, 0x38} {
			if pi, valid := part(off+a, id); valid {
				s.attach = append(s.attach, placement{part: pi,
					t:     [3]float32{f32(xbe, off+a+4), f32(xbe, off+a+8), f32(xbe, off+a+12)},
					label: []string{"gear stick", "door left", "door right"}[i]})
			}
		}
		if pi, valid := part(off+0x48, id); valid {
			s.attach = append(s.attach, placement{part: pi,
				t:     [3]float32{f32(xbe, off+0x4C), f32(xbe, off+0x50), f32(xbe, off+0x54)},
				tiltX: float64(f32(xbe, off+0x58)),
				label: "steering wheel"})
		}
		for _, g := range []int{0x60, 0x64} {
			if pi, valid := part(off+g, id); valid {
				s.glows = append(s.glows, pi)
			}
		}
		for i := 0; i < 3; i++ {
			if pi, valid := part(off+0x7C+4*i, id); valid {
				s.fronts = append(s.fronts, pi)
			}
		}
		for _, w := range []int{0x88, 0xA8, 0xC8, 0xE8} {
			t := [3]float32{f32(xbe, off+w+16), f32(xbe, off+w+20), f32(xbe, off+w+24)}
			for i := 0; i < 4; i++ {
				if pi, valid := part(off+w+4*i, id); valid {
					s.wheels = append(s.wheels, placement{part: pi, t: t, label: cornerLabel(t)})
				}
			}
		}
		if _, dup := out[name]; dup {
			return nil, fmt.Errorf("plcar chassis: duplicate record for %s", name)
		}
		out[name] = s
	}
	for _, name := range ids {
		if _, found := out[name]; !found {
			return nil, fmt.Errorf("plcar chassis: no record for %s", name)
		}
	}
	return out, nil
}

// placement is one exported instance of a part: an optional X mirror (the
// game draws each rc wheel part twice with a mirrored transform), a rotation
// about X and a translation. label names the logical part; placements sharing
// a label merge into one named, pickable node of the exported variant (empty
// labels merge into a single anonymous node).
type placement struct {
	part    int
	mirrorX bool
	t       [3]float32
	tiltX   float64
	label   string
}

// cornerLabel names a wheel placement by its corner (+z is the car's front,
// -x its left).
func cornerLabel(t [3]float32) string {
	fr, lr := "R", "R"
	if t[2] > 0 {
		fr = "F"
	}
	if t[0] < 0 {
		lr = "L"
	}
	return "wheel " + fr + lr
}

// labeled builds one-part placements with a shared label.
func labeled(label string, idx ...int) []placement {
	out := make([]placement, len(idx))
	for i, pi := range idx {
		out[i] = placement{part: pi, label: label}
	}
	return out
}

// labeledEach gives every part its own "<prefix> (part N)" node.
func labeledEach(prefix string, idx []int) []placement {
	out := make([]placement, len(idx))
	for i, pi := range idx {
		out[i] = placement{part: pi, label: fmt.Sprintf("%s (part %d)", prefix, pi)}
	}
	return out
}

// rcPlan builds the export plan for an obj_rc_* rival from the game's own
// draw evidence (bootoracle -carvtx over race states; the analysis lives in
// the game markdown, Part XIX):
//
//   - the visible near-distance set is parts {0 ground shadow, 1 LOD0 body,
//     2 front detail, 7, 8 wheels}. Part 9 draws ONLY under the shadow pass's
//     light projection (x-scale 0.052 vs the camera's 1.19) — it is the
//     shadow-caster proxy hull, never the visible car — and the remaining
//     parts are front-state variants, effect/glow overlays and lower-LOD
//     copies, none drawn at the near distance.
//   - the wheel parts sit in local space (one per axle) and the game draws
//     each twice with mirrored transforms. The placements are not in the
//     file's tables, but the far-LOD bodies carry the same wheels BAKED at
//     their true positions: clustering the vertices of body batches whose
//     materials use only the wheel textures gives the four wheel centres.
//     For dayts (the captured car) the derived centres match the game's own
//     draw matrices to a millimetre: file (±0.721, 0.341, −1.246) rear /
//     (±0.709, 0.341, 1.154) front vs captured (±0.72, 0.34, −1.246) /
//     (±0.72, 0.34, 1.154). Part 7 is the rear axle, 8 the front (captured
//     assignment, declared for the fleet).
func rcPlan(p *pmt) ([]placement, error) {
	wheels, ok := chassis[p.name]
	if !ok {
		return nil, fmt.Errorf("%s: no chassis-table entry", p.name)
	}
	return rcNearPlan(wheels[0]), nil
}

// rcNearPlan is the near-distance visible set: the game's LOD-0 group as its
// group table lists it — {0 ground shadow, 1 body, 2 front detail} plus the
// four wheels exactly as the chassis table places them.
func rcNearPlan(wheels []wheelSpec) []placement {
	plan := []placement{
		{part: 0, label: "ground shadow"},
		{part: 1, label: "body"},
		{part: 2, label: "front detail"},
	}
	return append(plan, wheelPlacements(wheels)...)
}

// wheelPlacements turns the chassis table's wheel entries into placements. The
// wheel geometry is modeled as the LEFT wheel; the game draws the x > 0
// entries with a mirrored transform (capture-pinned convention).
func wheelPlacements(wheels []wheelSpec) []placement {
	out := make([]placement, len(wheels))
	for i, w := range wheels {
		out[i] = placement{part: w.part, mirrorX: w.t[0] > 0, t: w.t, label: cornerLabel(w.t)}
	}
	return out
}

// checkChassis cross-validates the chassis table's id->file mapping against
// the model's own baked far-LOD wheel clusters where they derive cleanly: the
// z centres (axle positions) agreed to ~2 cm on every derivable car when the
// mapping was established, so a gross disagreement means a row landed on the
// wrong file.
func checkChassis(p *pmt, wheels []wheelSpec) {
	clusters, err := rcWheelCentres(p)
	if err != nil {
		return // fx/328gts: no derivable clusters — nothing to check against
	}
	for _, w := range wheels {
		sx, sz := sign(w.t[0]), sign(w.t[2])
		c := clusters[[2]int{sx, sz}]
		if dz := c[2] - w.t[2]; dz > 0.1 || dz < -0.1 {
			fmt.Fprintf(os.Stderr, "carex: %s: chassis wheel z %.3f vs baked cluster %.3f — id mapping suspect\n",
				p.name, w.t[2], c[2])
		}
	}
}

// rcWheelCentres derives the four wheel centres; it errors when the
// derivation has no clean source in the file.
func rcWheelCentres(p *pmt) (map[[2]int][3]float32, error) {
	// The tire/rim textures, taken from the LOD1 wheel parts (15/16), whose
	// materials are the cleanest source (the LOD0 wheels also sample detail
	// textures shared with the body on some cars).
	wtex := map[int]bool{}
	for _, pi := range []int{15, 16} {
		if pi >= p.nParts {
			continue
		}
		pt, err := p.parsePart(pi)
		if err != nil {
			return nil, err
		}
		for _, m := range pt.mats {
			if m.texIdx >= 0 {
				wtex[m.texIdx] = true
			}
		}
	}
	// Baked wheel batches: in car-sized parts (bounding sphere r > 1), every
	// material texture drawn from the wheel set.
	type acc struct {
		mn, mx [3]float32
		n      int
	}
	quads := map[[2]int]*acc{}
	for pi := 0; pi < p.nParts; pi++ {
		if pi == 7 || pi == 8 || pi == 15 || pi == 16 {
			continue
		}
		rec := 0x18 + pi*0x3C
		hdr := int(u32(p.a, rec+4*7))
		if f32(p.a, hdr+16) < 1 {
			continue
		}
		pt, err := p.parsePart(pi)
		if err != nil {
			return nil, err
		}
		for _, b := range pt.batches {
			m := pt.mats[b.matIdx]
			if m.texIdx < 0 || !wtex[m.texIdx] {
				continue
			}
			bp := pt.pairs[b.pair]
			idxCount, _ := indexCount(b.prim, b.prims)
			// Wheels touch the ground: a baked wheel batch's lowest vertex is
			// at y≈0. Reject look-alike batches higher up (badges, mirrors).
			ymin := float32(math.MaxFloat32)
			for i := uint32(0); i < idxCount; i++ {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+i)*2)) + b.baseVtx
				if y := f32(p.b, int(bp.vbOff)+int(ix)*int(bp.stride)+4); y < ymin {
					ymin = y
				}
			}
			if ymin > 0.08 {
				continue
			}
			for i := uint32(0); i < idxCount; i++ {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+i)*2)) + b.baseVtx
				v := int(bp.vbOff) + int(ix)*int(bp.stride)
				x, y, z := f32(p.b, v), f32(p.b, v+4), f32(p.b, v+8)
				key := [2]int{sign(x), sign(z)}
				a := quads[key]
				if a == nil {
					a = &acc{mn: [3]float32{x, y, z}, mx: [3]float32{x, y, z}}
					quads[key] = a
				}
				for c, f := range [3]float32{x, y, z} {
					if f < a.mn[c] {
						a.mn[c] = f
					}
					if f > a.mx[c] {
						a.mx[c] = f
					}
				}
				a.n++
			}
		}
	}
	out := map[[2]int][3]float32{}
	for _, sx := range []int{1, -1} {
		for _, sz := range []int{1, -1} {
			a := quads[[2]int{sx, sz}]
			if a == nil {
				return nil, fmt.Errorf("%s: no baked wheel cluster at quadrant %d,%d", p.name, sx, sz)
			}
			c := [3]float32{(a.mn[0] + a.mx[0]) / 2, (a.mn[1] + a.mx[1]) / 2, (a.mn[2] + a.mx[2]) / 2}
			if c[1] < 0.1 || c[1] > 0.7 || c[0]*float32(sx) < 0.4 || c[0]*float32(sx) > 1.3 ||
				c[2]*float32(sz) < 0.4 || c[2]*float32(sz) > 2.0 {
				return nil, fmt.Errorf("%s: implausible wheel centre %v at quadrant %d,%d", p.name, c, sx, sz)
			}
			out[[2]int{sx, sz}] = c
		}
	}
	return out, nil
}

func sign(f float32) int {
	if f < 0 {
		return -1
	}
	return 1
}

// plan builds the placement list for a model: the rc rivals get the
// group-table visible set, the player cars their chassis-record rest pose,
// and everything else all parts in place.
func (p *pmt) plan() ([]placement, error) {
	if strings.HasPrefix(p.name, "obj_rc_") && p.name != "obj_rc_all" && onlyParts == nil {
		return rcPlan(p)
	}
	if spec, ok := plcars[p.name]; ok && onlyParts == nil {
		return plcarCarPlan(p, spec), nil
	}
	// Stage geometry (cs_*): one named node per part, so the viewer can pick
	// them apart — and so the export takes the node path, which carries the
	// baked COLOR_0 and TEXCOORD_1 the stage layouts declare.
	if strings.HasPrefix(p.name, "cs_") && onlyParts == nil {
		var out []placement
		for pi := 0; pi < p.nParts; pi++ {
			out = append(out, placement{part: pi, label: fmt.Sprintf("part %d", pi)})
		}
		return out, nil
	}
	var out []placement
	for pi := 0; pi < p.nParts; pi++ {
		out = append(out, placement{part: pi})
	}
	return out, nil
}

// partMinY is the lowest vertex Y across a part's vertex buffers — near zero
// when the part carries baked ground-touching wheels.
func (p *pmt) partMinY(pi int) float32 {
	pt, err := p.parsePart(pi)
	if err != nil {
		return 0
	}
	min := float32(math.MaxFloat32)
	for _, bp := range pt.pairs {
		for i := uint32(0); i < bp.vbBytes/bp.stride; i++ {
			if y := f32(p.b, int(bp.vbOff+i*bp.stride)+4); y < min {
				min = y
			}
		}
	}
	return min
}

// partKind reads a part's piece-node flags word (the "kind": 2 marks the flat
// baked ground-shadow part).
func partKind(p *pmt, pi int) uint32 {
	hdr := int(u32(p.a, 0x18+pi*0x3C+28))
	return u32(p.a, hdr)
}

// plcarRoles splits a player-car record into the visible rest pose and the
// leftovers: the shadow (the kind-2 static slot), the other static shells,
// and any file parts the record never references.
func plcarRoles(p *pmt, spec plcarSpec) (shadow int, shells, extras []int) {
	shadow = -1
	for _, pi := range spec.statics[1:] {
		if shadow < 0 && partKind(p, pi) == 2 {
			shadow = pi
			continue
		}
		shells = append(shells, pi)
	}
	ref := map[int]bool{}
	if spec.panel >= 0 {
		ref[spec.panel] = true
	}
	for _, pi := range spec.statics {
		ref[pi] = true
	}
	for _, pl := range spec.attach {
		ref[pl.part] = true
	}
	for _, pl := range spec.wheels {
		ref[pl.part] = true
	}
	for _, pi := range spec.glows {
		ref[pi] = true
	}
	for _, pi := range spec.fronts {
		ref[pi] = true
	}
	for pi := 0; pi < p.nParts; pi++ {
		if !ref[pi] {
			extras = append(extras, pi)
		}
	}
	return shadow, shells, extras
}

// plcarCarPlan is the assembled rest pose: body + ground shadow + the default
// front state + the placed doors/gear/steering and wheels.
func plcarCarPlan(p *pmt, spec plcarSpec) []placement {
	shadow, _, _ := plcarRoles(p, spec)
	plan := []placement{{part: spec.statics[0], label: "body"}}
	if spec.panel >= 0 {
		plan = append(plan, placement{part: spec.panel, label: "body panel"})
	}
	if shadow >= 0 {
		plan = append(plan, placement{part: shadow, label: "ground shadow"})
	}
	if len(spec.fronts) > 0 {
		plan = append(plan, placement{part: spec.fronts[0], label: "front panel"})
	}
	plan = append(plan, spec.attach...)
	return append(plan, spec.wheels...)
}

// modelVariant is one exportable alternate of a model: a named placement plan
// that becomes one glTF scene of the GLB.
type modelVariant struct {
	id, name, desc string
	plan           []placement
}

func parts(idx ...int) []placement {
	out := make([]placement, len(idx))
	for i, pi := range idx {
		out[i] = placement{part: pi}
	}
	return out
}

// variants lists a model's alternates. The part roles come from the game's
// own runtime group tables and draw captures (writeup Part XX): the rc
// template's LOD groups {1,0,6,2,3|4} / {10,0,14,11,12|13} / {17,0,21,19,20}
// / {22} / {23}, part 9 = the shadow-caster proxy (drawn only under the
// light's projection), 5 = the effect quad stack, 6/14/21 = light and road
// glow overlays. The first variant is the default scene.
func (p *pmt) variants() ([]modelVariant, error) {
	if strings.HasPrefix(p.name, "obj_rc_") && p.name != "obj_rc_all" {
		wheels, ok := chassis[p.name]
		if !ok {
			return nil, fmt.Errorf("%s: no chassis-table entry", p.name)
		}
		checkChassis(p, wheels[0])
		// The mid LODs usually bake their wheels in place; where a body floats
		// (328gts' LOD 2/3, 550b/575sa's LOD 3 — lowest vertex well above the
		// ground) the variant borrows the placed LOD-1 wheels. LOD 4 is
		// wheel-less on every car and ships as the game has it.
		withWheels := func(plan []placement, bodyPart int) []placement {
			if p.partMinY(bodyPart) > 0.08 {
				return append(plan, wheelPlacements(wheels[1])...)
			}
			return plan
		}
		lodPlan := func(body, front int) []placement {
			plan := append(labeled("ground shadow", 0), labeled("body", body)...)
			if front >= 0 {
				plan = append(plan, labeled("front detail", front)...)
			}
			return plan
		}
		return []modelVariant{
			{"car", "Car", "", rcNearPlan(wheels[0])},
			{"lod1", "LOD 1", "", append(lodPlan(10, 11), wheelPlacements(wheels[1])...)},
			{"lod2", "LOD 2", "", withWheels(lodPlan(17, 19), 17)},
			{"lod3", "LOD 3", "", withWheels(lodPlan(22, -1), 22)},
			{"lod4", "LOD 4", "", lodPlan(23, -1)},
			{"caster", "Shadow caster", "the proxy hull the game draws into the shadow map — never directly visible", labeled("caster hull", 9)},
			{"overlays", "Light & effect overlays", "the livery texture-carrier quads (never rendered — the game binds their atlases in place of the body texture) and the additive glow quads", append(labeled("livery texture carrier", 5), labeled("light & road glow", 6)...)},
		}, nil
	}
	if spec, ok := plcars[p.name]; ok {
		shadow, shells, extras := plcarRoles(p, spec)
		if shadow < 0 {
			fmt.Fprintf(os.Stderr, "carex: %s: no kind-2 ground shadow among static slots %v\n", p.name, spec.statics)
		}
		vars := []modelVariant{{"car", "Car", "", plcarCarPlan(p, spec)}}
		if len(spec.fronts) > 1 {
			vars = append(vars, modelVariant{"alt", "Alternate panels",
				"the front-panel state variants the game swaps in (headlights, damage)", labeledEach("front state", spec.fronts[1:])})
		}
		if len(spec.glows) > 0 {
			vars = append(vars, modelVariant{"overlays", "Light overlays",
				"the additively-blended brake/headlight glow quads", labeledEach("light overlay", spec.glows)})
		}
		if rest := append(append([]int{}, shells...), extras...); len(rest) > 0 {
			vars = append(vars, modelVariant{"shells", "Shells & extras",
				"the record's static shell slots and parts outside the chassis record — proxy hulls and state shells", labeledEach("shell", rest)})
		}
		return vars, nil
	}
	def, err := p.plan()
	if err != nil {
		return nil, err
	}
	return []modelVariant{{"car", "Car", "", def}}, nil
}

// export writes the model as a GLB: a single scene when it has one variant,
// or one glTF scene per variant (scene 0 = the default "car") when several.
func export(p *pmt, texs []texInfo, outPath string) (string, error) {
	writeOne := func(v glb.ModelVariant, summary string) (string, error) {
		if len(v.Nodes) > 0 {
			v.Name = "car"
			return summary, glb.WriteVariantScenes(outPath, []glb.ModelVariant{v})
		}
		return summary, glb.WriteTexturedLit(outPath, v.Positions, v.UVs, v.Normals, v.TexGroups, v.ColorGroups)
	}
	// A course with its visibility database exports as the walker draws it:
	// world geometry + placed instances (stage.go).
	if p.vis != nil && onlyParts == nil {
		nodes, _, _, summary, err := p.buildStageNodes(texs, false)
		if err != nil {
			return "", err
		}
		return summary, glb.WriteVariantScenes(outPath, []glb.ModelVariant{{Name: "car", Nodes: nodes}})
	}
	if onlyParts != nil {
		plan, err := p.plan()
		if err != nil {
			return "", err
		}
		v, summary, err := buildVariant(p, texs, plan)
		if err != nil {
			return "", err
		}
		return writeOne(v, summary)
	}
	vars, err := p.variants()
	if err != nil {
		return "", err
	}
	return exportVariantList(p, texs, outPath, vars, writeOne)
}

// exportVariantList builds and writes an explicit variant list.
func exportVariantList(p *pmt, texs []texInfo, outPath string, vars []modelVariant,
	writeOne func(glb.ModelVariant, string) (string, error)) (string, error) {
	out := make([]glb.ModelVariant, len(vars))
	var summary string
	for i, mv := range vars {
		v, s, err := buildVariant(p, texs, mv.plan)
		if err != nil {
			return "", fmt.Errorf("variant %s: %w", mv.id, err)
		}
		v.Name = mv.id
		out[i] = v
		if i == 0 {
			summary = s
		}
	}
	if len(out) == 1 && writeOne != nil {
		return writeOne(out[0], summary)
	}
	if err := glb.WriteVariantScenes(outPath, out); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s + %d variants", summary, len(out)-1), nil
}

// buildVariant assembles one placement plan into a variant. Placements
// sharing a label become one named node (pickable in the viewer); an empty
// label merges everything into a single anonymous node (the legacy shape for
// uncurated models).
func buildVariant(p *pmt, texs []texInfo, plan []placement) (glb.ModelVariant, string, error) {
	type texKey struct {
		tex             int
		additive, sheen bool
		wrapS, wrapT    int
	}
	type colKey struct {
		rgba            [4]int
		additive, sheen bool
	}
	type nodeAcc struct {
		name      string
		positions [][3]float32
		normals   [][3]float32
		uvs       [][2]float32
		uv2s      [][2]float32
		colors    [][4]uint8
		hasUV2    bool // any pair actually carried a second UV set
		hasColor  bool // any pair actually carried a baked colour
		texTris   map[texKey][][3]uint32
		colorTris map[colKey][][3]uint32
	}
	var order []*nodeAcc
	byLabel := map[string]*nodeAcc{}
	node := func(label string) *nodeAcc {
		if n, ok := byLabel[label]; ok {
			return n
		}
		n := &nodeAcc{name: label, texTris: map[texKey][][3]uint32{}, colorTris: map[colKey][][3]uint32{}}
		byLabel[label] = n
		order = append(order, n)
		return n
	}

	totalTris, totalVerts := 0, 0
	for _, pl := range plan {
		if onlyParts != nil && !onlyParts[pl.part] {
			continue
		}
		if pl.part >= p.nParts {
			continue
		}
		na := node(pl.label)
		pt, err := p.parsePart(pl.part)
		if err != nil {
			return glb.ModelVariant{}, "", err
		}
		// Per pair, decode the vertex buffer once and remember its base in the
		// node's merged arrays.
		vbase := make([]uint32, len(pt.pairs))
		for k, bp := range pt.pairs {
			vbase[k] = uint32(len(na.positions))
			pos, nrm, uv, uv2, col := p.decodeVerts(bp)
			sn, c := float32(math.Sin(pl.tiltX)), float32(math.Cos(pl.tiltX))
			for i := range pos {
				if pl.mirrorX {
					pos[i][0] = -pos[i][0]
					nrm[i][0] = -nrm[i][0]
				}
				y, z := pos[i][1], pos[i][2]
				pos[i][1] = c*y - sn*z
				pos[i][2] = sn*y + c*z
				pos[i][0] += pl.t[0]
				pos[i][1] += pl.t[1]
				pos[i][2] += pl.t[2]
				ny, nz := nrm[i][1], nrm[i][2]
				nrm[i][1] = c*ny - sn*nz
				nrm[i][2] = sn*ny + c*nz
			}
			na.positions = append(na.positions, pos...)
			na.normals = append(na.normals, nrm...)
			na.uvs = append(na.uvs, uv...)
			// The node's colour/uv2 arrays run parallel to positions across all
			// pairs, so pairs lacking the attribute contribute the neutral value
			// (white multiplies to identity; uv2 zero). If no pair ever carries
			// the attribute the array is dropped at finish, keeping the /Cars
			// exports (which never carry either) byte-identical.
			if uv2 == nil {
				uv2 = make([][2]float32, len(pos))
			} else {
				na.hasUV2 = true
			}
			if col == nil {
				col = make([][4]uint8, len(pos))
				for i := range col {
					col[i] = [4]uint8{255, 255, 255, 255}
				}
			} else {
				na.hasColor = true
			}
			na.uv2s = append(na.uv2s, uv2...)
			na.colors = append(na.colors, col...)
		}
		for _, b := range pt.batches {
			bp := pt.pairs[b.pair]
			idxCount, _ := indexCount(b.prim, b.prims)
			raw := make([]uint32, idxCount)
			nVerts := bp.vbBytes / bp.stride
			for i := range raw {
				ix := uint32(u16(p.a, int(bp.ibOff)+int(b.first+uint32(i))*2)) + b.baseVtx
				if ix >= nVerts {
					return glb.ModelVariant{}, "", fmt.Errorf("index %d out of VB (%d verts)", ix, nVerts)
				}
				raw[i] = ix + vbase[b.pair]
			}
			tris := triangulate(b.prim, raw)
			if pl.mirrorX {
				for i := range tris {
					tris[i][1], tris[i][2] = tris[i][2], tris[i][1]
				}
			}
			totalTris += len(tris)
			m := pt.mats[b.matIdx]
			if m.texIdx >= 0 && texs[m.texIdx].img != nil && !texs[m.texIdx].cube {
				k := texKey{m.texIdx, m.additive, m.sheen, m.wrapS, m.wrapT}
				na.texTris[k] = append(na.texTris[k], tris...)
			} else {
				k := colKey{[4]int{int(m.diffuse[0] * 255), int(m.diffuse[1] * 255), int(m.diffuse[2] * 255), int(m.alpha * 255)}, m.additive, m.sheen}
				na.colorTris[k] = append(na.colorTris[k], tris...)
			}
		}
	}

	finish := func(na *nodeAcc) glb.VariantNode {
		// Group order is output order: sort the map keys so an export is
		// byte-stable run to run (map iteration used to decide it, and every
		// re-export rewrote every GLB in the site).
		tkeys := make([]texKey, 0, len(na.texTris))
		for k := range na.texTris {
			tkeys = append(tkeys, k)
		}
		sort.Slice(tkeys, func(i, j int) bool {
			a, b := tkeys[i], tkeys[j]
			if a.tex != b.tex {
				return a.tex < b.tex
			}
			if a.wrapS != b.wrapS {
				return a.wrapS < b.wrapS
			}
			if a.wrapT != b.wrapT {
				return a.wrapT < b.wrapT
			}
			if a.additive != b.additive {
				return !a.additive
			}
			return !a.sheen && b.sheen
		})
		var texGroups []glb.TexturedGroup
		for _, k := range tkeys {
			texGroups = append(texGroups, glb.TexturedGroup{
				Tris: na.texTris[k], Image: texs[k.tex].img, WrapS: k.wrapS, WrapT: k.wrapT,
				Additive: k.additive, Sheen: k.sheen,
				// The game's own alpha test: enabled with ref 0x01 at every
				// opaque stage/car draw (bootoracle -carvtx at=1:…:01) — only
				// fully transparent texels are cut. The road's asphalt alpha
				// (~0.25, a lane-marking combiner input) must survive.
				AlphaCutoff: 1.0 / 255,
			})
		}
		ckeys := make([]colKey, 0, len(na.colorTris))
		for k := range na.colorTris {
			ckeys = append(ckeys, k)
		}
		sort.Slice(ckeys, func(i, j int) bool {
			a, b := ckeys[i], ckeys[j]
			if a.rgba != b.rgba {
				for c := 0; c < 4; c++ {
					if a.rgba[c] != b.rgba[c] {
						return a.rgba[c] < b.rgba[c]
					}
				}
			}
			if a.additive != b.additive {
				return !a.additive
			}
			return !a.sheen && b.sheen
		})
		var colorGroups []glb.TriGroup
		for _, k := range ckeys {
			colorGroups = append(colorGroups, glb.TriGroup{
				Tris:     na.colorTris[k],
				Color:    [3]float32{float32(k.rgba[0]) / 255, float32(k.rgba[1]) / 255, float32(k.rgba[2]) / 255},
				Alpha:    float32(k.rgba[3]) / 255,
				Additive: k.additive,
				Sheen:    k.sheen,
			})
		}
		totalVerts += len(na.positions)
		n := glb.VariantNode{Name: na.name, Positions: na.positions, Normals: na.normals,
			UVs: na.uvs, TexGroups: texGroups, ColorGroups: colorGroups}
		if na.hasUV2 {
			n.UV2 = na.uv2s
		}
		if na.hasColor {
			n.Colors = na.colors
		}
		return n
	}

	var v glb.ModelVariant
	if len(order) == 1 && order[0].name == "" {
		n := finish(order[0])
		v = glb.ModelVariant{Positions: n.Positions, Normals: n.Normals, UVs: n.UVs,
			UV2: n.UV2, Colors: n.Colors, TexGroups: n.TexGroups, ColorGroups: n.ColorGroups}
	} else {
		for _, na := range order {
			v.Nodes = append(v.Nodes, finish(na))
		}
	}
	summary := fmt.Sprintf("%d parts, %d verts, %d tris, %d nodes",
		p.nParts, totalVerts, totalTris, len(order))
	return v, summary, nil
}

func main() {
	imagePath := flag.String("image", "", "Xbox disc image")
	one := flag.String("file", "", "single disc path to extract (e.g. /Cars/obj_plcar_dino_pmt.sz)")
	all := flag.Bool("all", false, "extract every /Cars/*_pmt.sz model")
	outDir := flag.String("o", "out", "output directory")
	dumpTex := flag.Bool("dumptex", false, "also write each decoded texture as PNG")
	insp := flag.Bool("inspect", false, "print part/batch/material tables instead of exporting")
	forest := flag.Bool("forest", false, "print a course's placement forest (e0 nodes) instead of exporting")
	nodesFlag := flag.String("nodes", "", "part:id,... — export each named forest node as its own GLB (diagnostic)")
	partsFlag := flag.String("parts", "", "comma-separated part indices: export only these parts (diagnostic)")
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
	if *partsFlag != "" {
		onlyParts = map[int]bool{}
		for _, tok := range strings.Split(*partsFlag, ",") {
			var pi int
			if _, err := fmt.Sscanf(strings.TrimSpace(tok), "%d", &pi); err != nil {
				fatal("bad -parts %q", *partsFlag)
			}
			onlyParts[pi] = true
		}
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
	if err := loadChassis(disc); err != nil {
		fatal("chassis table: %v", err)
	}

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
		if err := loadVisBin(disc, f, p); err != nil {
			fmt.Fprintf(os.Stderr, "carex: %s: %v (skipped)\n", f, err)
			continue
		}
		texs, pixBytes, err := p.parseTextures()
		if err != nil {
			fmt.Fprintf(os.Stderr, "carex: %s: %v (skipped)\n", f, err)
			continue
		}
		if *forest {
			fmt.Printf("== %s: %d parts ==\n", f, p.nParts)
			if err := p.inspectForest(); err != nil {
				fmt.Fprintf(os.Stderr, "carex: %s: forest: %v\n", f, err)
			}
			continue
		}
		if *nodesFlag != "" {
			if err := p.exportNodes(texs, *nodesFlag, *outDir); err != nil {
				fmt.Fprintf(os.Stderr, "carex: %s: nodes: %v\n", f, err)
			}
			continue
		}
		if *insp {
			fmt.Printf("== %s: %d parts, %d textures ==\n", f, p.nParts, p.nTex)
			if err := inspect(p, texs); err != nil {
				fmt.Fprintf(os.Stderr, "carex: %s: inspect: %v\n", f, err)
			}
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

// envFaces decodes the six faces of the model's first environment cube (the
// XPR bank stores them consecutively, face stride = the level-0 size, D3D
// order +x,-x,+y,-y,+z,-z). Nil when the bank has no cube.
func envFaces(p *pmt, texs []texInfo) []image.Image {
	for _, t := range texs {
		if !t.cube || t.fmtByte == 0 {
			continue
		}
		blockLen := 16
		if t.fmtByte == 0x0C {
			blockLen = 8
		}
		faceSize := ((t.w + 3) / 4) * ((t.h + 3) / 4) * blockLen
		// The face stride is padded (the 8x8 cubes store 0x80-byte faces for
		// 0x40 bytes of blocks): derive it from the bank's own layout — the
		// distance to the next texture's pixels spans exactly six faces.
		stride := faceSize
		next := -1
		for _, o := range texs {
			if d := int(o.dataOff) - int(t.dataOff); d > 0 && (next < 0 || d < next) {
				next = d
			}
		}
		if next > 0 && next%6 == 0 && next/6 >= faceSize {
			stride = next / 6
		}
		out := make([]image.Image, 6)
		for f := 0; f < 6; f++ {
			img := decodeTexture(p.b[int(t.dataOff)+f*stride:], t.w, t.h, t.fmtByte)
			if img == nil {
				return nil
			}
			out[f] = img
		}
		return out
	}
	return nil
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
	if err := loadChassis(disc); err != nil {
		fatal("chassis table: %v", err)
	}
	// Retro-X: the curated roster becomes model3d object assets under a
	// builder-written tree (validated on Write).
	b := build.New(siteDir, "outrun-2006-xbox")
	b.SetTitle("OutRun 2006: Coast 2 Coast")
	b.SetPlatform("Original Xbox")
	b.SetYear(2006)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 640, H: 480},
		TickHz: 60,
		// 480i on a CRT; under the filter the viewer renders at 480 lines.
		Filter: "crt",
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
		if err := loadVisBin(disc, discPath, p); err != nil {
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
		obj := &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: outName,
			Props: map[string]any{"source": discPath},
		}
		// The model's static environment cube (in the game the body materials
		// sample it — substituted for the live reflection RTT at runtime; at
		// rest it is the sheen). Exported as six face PNGs for sheen-marked
		// materials in the viewer.
		if faces := envFaces(p, texs); faces != nil {
			for fi, img := range faces {
				fn := fmt.Sprintf("%s-env-%s.png", id, [6]string{"px", "nx", "py", "ny", "pz", "nz"}[fi])
				path, err := b.Path("objects", fn)
				if err != nil {
					fatal("%v", err)
				}
				writePNG(path, img)
				obj.EnvMap = append(obj.EnvMap, fn)
			}
		}
		if vars, err := p.variants(); err == nil && len(vars) > 1 {
			for _, mv := range vars {
				obj.Variants = append(obj.Variants, schema.ModelVariant{
					ID: mv.id, Name: mv.name, Scene: mv.id, Description: mv.desc,
				})
			}
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: section}, obj)
		fmt.Printf("%-34s -> %s (%s)\n", discPath, out, summary)
	}

	// The player fleet, assembled from the chassis records; the Dino keeps its
	// established id and label.
	doOne("/Cars/obj_plcar_dino_pmt.sz", "plcar-dino.glb", "Dino 246 GTS (player)", "Player cars")
	for _, code := range []string{
		"250gto", "328gts", "360sp", "512bb", "550b", "575sa",
		"dayts", "f355sp", "f40", "f430", "f50", "fx", "gto", "testa",
	} {
		doOne("/Cars/obj_plcar_"+code+"_pmt.sz", "plcar-"+code+".glb",
			strings.ToUpper(code)+" (player)", "Player cars")
	}
	for _, code := range rivals {
		doOne("/Cars/obj_rc_"+code+"_pmt.sz", "rc-"+code+".glb",
			strings.ToUpper(code), "Rivals (AI)")
	}

	// The traffic fleet: twelve vehicles assembled out of obj_othcar per the
	// traffic table (bodies, placed wheels, caster, additive glows). The
	// shared ground-shadow model the table references is not part of this
	// file and is left out.
	exportTraffic(disc, b)

	// The first /Stage resident (Parts XXVI-XXIX): the beach course assembled
	// as the game assembles it — world geometry plus the placed-object forest
	// (palms, gantries, grandstands) under their w4 matrices, the distant
	// scenery ring, and the sky dome. The rest of the stage family
	// (collision, spline, fog/sun params) is still closed.
	exportStages(disc, b)

	if err := b.Write(); err != nil {
		fatal("%v", err)
	}
	for _, w := range b.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
}

// trafficNames are display names for the twelve table vehicles, in body-part
// order (curated from their renders; the data itself carries no names).
var trafficNames = [12]string{
	"Hatchback", "Coach", "Motorhome", "Pickup", "Coupé", "Sedan",
	"Compact", "Minivan", "Taxi", "Flatbed truck", "Truck", "Tanker truck",
}

// exportTraffic writes the twelve traffic vehicles from obj_othcar.
func exportTraffic(disc *xbox.Image, b *build.Builder) {
	const discPath = "/Cars/obj_othcar_pmt.sz"
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
	p, err := parsePMT("obj_othcar", data)
	if err != nil {
		fatal("%s: %v", discPath, err)
	}
	texs, _, err := p.parseTextures()
	if err != nil {
		fatal("%s: %v", discPath, err)
	}
	// One shared set of environment-cube faces for the whole fleet.
	var envRefs []string
	if faces := envFaces(p, texs); faces != nil {
		for fi, img := range faces {
			fn := fmt.Sprintf("othcar-env-%s.png", [6]string{"px", "nx", "py", "ny", "pz", "nz"}[fi])
			path, err := b.Path("objects", fn)
			if err != nil {
				fatal("%v", err)
			}
			writePNG(path, img)
			envRefs = append(envRefs, fn)
		}
	}
	for i, v := range traffic {
		id := fmt.Sprintf("othcar-%02d", i+1)
		name := trafficNames[i]
		ref := map[int]bool{v.caster: true}
		car := append(labeled("body", v.bodies[0]), v.wheels[0]...)
		lod1 := append(labeled("body", v.bodies[1]), v.wheels[1]...)
		for _, pl := range append(append([]placement{}, v.wheels[0]...), v.wheels[1]...) {
			ref[pl.part] = true
		}
		ref[v.bodies[0]], ref[v.bodies[1]] = true, true
		glows := map[int]bool{}
		for _, g := range v.glows {
			glows[g] = true
			ref[g] = true
		}
		var glowList, extras []int
		for pi := v.lo; pi < v.hi; pi++ {
			if glows[pi] {
				glowList = append(glowList, pi)
			} else if !ref[pi] {
				extras = append(extras, pi)
			}
		}
		vars := []modelVariant{
			{"car", "Car", "", car},
			{"lod1", "LOD 1", "", lod1},
			{"caster", "Shadow caster", "the proxy hull the game draws into the shadow map — never directly visible", labeled("caster hull", v.caster)},
			{"overlays", "Light overlays", "headlight/brake flares and road glow, drawn additively at night", labeledEach("light overlay", glowList)},
		}
		if len(extras) > 0 {
			vars = append(vars, modelVariant{"extras", "Extra parts",
				"parts of this vehicle's range the traffic table never places", labeledEach("extra", extras)})
		}
		out, err := b.Path("objects", id+".glb")
		if err != nil {
			fatal("%v", err)
		}
		summary, err := exportVariantList(p, texs, out, vars, nil)
		if err != nil {
			fatal("traffic %s: %v", id, err)
		}
		obj := &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: id + ".glb",
			EnvMap: envRefs,
			Props:  map[string]any{"source": discPath, "bodyPart": v.bodies[0]},
		}
		for _, mv := range vars {
			obj.Variants = append(obj.Variants, schema.ModelVariant{ID: mv.id, Name: mv.name, Scene: mv.id, Description: mv.desc})
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Traffic"}, obj)
		fmt.Printf("%-34s -> %s (%s)\n", discPath+"#"+id, out, summary)
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
