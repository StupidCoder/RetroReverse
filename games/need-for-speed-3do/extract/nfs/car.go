package nfs

import (
	"fmt"
	"image"

	"retroreverse.com/tools/platform/threedo"
)

// Car texturing, traced from the texset builder at 0x24800 and verified
// bit-exact against the running game's per-face texture array ([car+0x4E0]):
//
// A car .wrapFam is a wwww container whose children pair up as (ORI3 model,
// SHPM shape) per LOD, plus one trailing tuning-data leaf. The shape's SPoT
// directory ({char4 name, u32 offset} entries) holds N small textures plus a
// "!ori" entry: that chunk carries, after a 0x18-byte header and
// [chunk+0x10] × 40-byte records, one 8-byte {faceIndex, spotIndex} entry per
// model face — the face → texture map. Each SPoT record is
// {u32 flags, u16 w, u16 h, u32 0, u32 0x00320032, pixels}, uncoded RGB555.
// The cel engine maps the whole texture onto the quad, so UVs are the full
// texture per face.

// CarLOD is one (model, textures) pair from a car .wrapFam.
type CarLOD struct {
	Model   *threedo.Model
	FaceTex []int          // per-face SPoT index into Textures
	Textures []*image.RGBA // decoded SPoT textures
}

// ParseCarFam decodes every LOD of a car .wrapFam. LODs come in disc order
// (the Diablo's is smallest first: DIABL300, DIABL200, DIABLO00). The (model,
// shape) pairs sit under a nested wwww node; consecutive leaves pair up.
func ParseCarFam(fam []byte) ([]*CarLOD, error) {
	leaves, err := threedo.ParseWrap(fam)
	if err != nil {
		return nil, err
	}
	var lods []*CarLOD
	for i := 0; i+1 < len(leaves); i++ {
		m, s := leaves[i], leaves[i+1]
		if m.Kind != "model" || s.Kind != "shape" {
			continue
		}
		models, err := threedo.ParseModels(fam[m.Offset:])
		if err != nil || len(models) == 0 {
			return nil, fmt.Errorf("nfs: car LOD %d: bad ORI3: %v", len(lods), err)
		}
		lod := &CarLOD{Model: models[0]}
		if err := parseCarShape(fam[s.Offset:], lod); err != nil {
			return nil, fmt.Errorf("nfs: car LOD %d: %v", len(lods), err)
		}
		if len(lod.FaceTex) < len(lod.Model.Faces) {
			return nil, fmt.Errorf("nfs: car LOD %d: %d face-map entries for %d faces",
				len(lods), len(lod.FaceTex), len(lod.Model.Faces))
		}
		lods = append(lods, lod)
		i++ // consumed the pair
	}
	if len(lods) == 0 {
		return nil, fmt.Errorf("nfs: no (model, shape) pairs in wrapFam")
	}
	return lods, nil
}

// parseCarShape decodes the SPoT textures and the !ori face map of one shape.
func parseCarShape(shape []byte, lod *CarLOD) error {
	if len(shape) < 0x10 || string(shape[0:4]) != "SHPM" {
		return fmt.Errorf("not a SHPM shape")
	}
	count := int(be32(shape[8:]))
	if count <= 0 || count > 1024 {
		return fmt.Errorf("implausible SPoT count %d", count)
	}
	oriOff := -1
	var spotOffs []int
	for k := 0; k < count; k++ {
		name := string(shape[0x10+8*k : 0x10+8*k+4])
		off := int(be32(shape[0x14+8*k:]))
		if name == "!ori" || name == "!ORI" {
			oriOff = off
			continue
		}
		if off < len(shape) && shape[off] == 0x20 {
			// a bare PLUT chunk listed in the directory (the traffic cars'
			// shared "plt0".."plt3" recolour palettes) — the textures reach
			// it by signed offset; it is not itself a texture. These always
			// trail the numbered textures, so skipping keeps !ori indices.
			continue
		}
		spotOffs = append(spotOffs, off)
	}
	if oriOff < 0 {
		return fmt.Errorf("shape has no !ori face map")
	}
	for _, off := range spotOffs {
		img, err := decodeSpot(shape, off)
		if err != nil {
			return err
		}
		lod.Textures = append(lod.Textures, img)
	}
	// face map: 0x18-byte header, u32 count of 40-byte records at +0x10,
	// then 8-byte {faceIdx, spotIdx} entries.
	n40 := int(be32(shape[oriOff+0x10:]))
	tbl := oriOff + 0x18 + 40*n40
	for p := tbl; p+8 <= len(shape); p += 8 {
		idx := int(sbe32(shape[p:]))
		spot := int(sbe32(shape[p+4:]))
		if idx != len(lod.FaceTex) { // entries are 0,1,2,... — stop at the end
			break
		}
		if spot < 0 || spot >= len(lod.Textures) {
			spot = 0
		}
		lod.FaceTex = append(lod.FaceTex, spot)
	}
	return nil
}

// spotBPP maps the SPoT type byte to bits per pixel — the game's own table
// (bytes 00 01 02 04 06 08 10 next to the CCB template the cel constructor
// 0x3B9C4 copies; type is word0's top byte). The table is exactly the CCB
// PRE0 bpp-code map: a type byte with bit 7 set skips the table and is the
// PRE0 low bits directly (the BMI path at 0x3BA44: PRE0 = type & 0x17, so
// bits 0-2 = bpp code, bit 4 = uncoded) — 0x84 is 6bpp coded, 0x85 8bpp
// coded. Types 5 and 0x85 render with the global recolour palette the
// traffic-car loader stamps (LDREQ lr, [r4, #-0x4] at 0x3BA0C/0x3BA90);
// statically their chain resolves to the file's "plt0", the default scheme.
var spotBPP = [7]int{0, 1, 2, 4, 6, 8, 16}

// decodeSpot decodes one SPoT texture record:
//
//	+0x00 u8 type (index into spotBPP), s24 PLUT-chunk offset
//	+0x04 u16 w, u16 h
//	+0x08 u32 0, +0x0C u32 0x00320032
//	+0x10 pixel rows, MSB-first bit-packed, word-aligned, minimum 8 bytes
//	At the PLUT offset: {u8 0x20, s24 next, u32 ?, u32 ?, u32 external-ptr
//	(0 = colors inline), RGB555 colors at +0x10}. Coded pixels are a 5-bit
//	PLUT index with the P-bit above it (the cel engine's 6bpp PDEC layout).
//
// The PLUT offset and the chain hops are SIGNED 24-bit, relative to the
// record (the constructor at 0x3B9C4 does ADD r9, r9, r2, ASR #8 with
// r2 = word0 << 8): the traffic cars point every texture backwards at one
// shared "plt0" chunk instead of carrying a PLUT each.
func decodeSpot(shape []byte, off int) (*image.RGBA, error) {
	if off+0x10 > len(shape) {
		return nil, fmt.Errorf("SPoT record at 0x%X out of range", off)
	}
	typ := int(shape[off])
	bpp, uncoded := 0, false
	switch {
	case typ < len(spotBPP):
		bpp = spotBPP[typ]
	case typ >= 0x80:
		bpp = spotBPP[typ&7] // PRE0-direct: bits 0-2 are the code
		uncoded = typ&0x10 != 0
	}
	if bpp == 0 {
		return nil, fmt.Errorf("SPoT type 0x%X unsupported", typ)
	}
	if uncoded && bpp != 16 {
		return nil, fmt.Errorf("SPoT type 0x%X: uncoded %dbpp unsupported", typ, bpp)
	}
	w := int(be16(shape[off+4:]))
	h := int(be16(shape[off+6:]))
	if w <= 0 || h <= 0 || w > 512 || h > 512 {
		return nil, fmt.Errorf("implausible SPoT size %dx%d", w, h)
	}
	// PLUT colors (coded depths only)
	var plut []uint16
	if bpp < 16 {
		chunk := off + s24(be32(shape[off:]))
		for chunk >= 0 && chunk+0x10 <= len(shape) && shape[chunk] != 0x20 {
			next := s24(be32(shape[chunk:]))
			if next == 0 {
				break
			}
			chunk += next
		}
		if chunk < 0 || chunk+0x10 > len(shape) || shape[chunk] != 0x20 {
			return nil, fmt.Errorf("SPoT at 0x%X: no PLUT chunk", off)
		}
		colors := chunk + 0x10
		n := 1 << min(bpp, 5)
		for i := 0; i < n && colors+2*i+2 <= len(shape); i++ {
			plut = append(plut, be16(shape[colors+2*i:]))
		}
	}
	if typ >= 0x80 {
		// The negative-type path also ORRs CCB_PACKED (0x200) into the
		// template flags (0x3BA10): the pixel stream is per-line RLE, not
		// row-strided. Reuse the platform cel decoder; PIXC pass-through and
		// BGND clear match the unpacked path's verified behaviour.
		cel := &threedo.Cel{
			Width: w, Height: h, BPP: bpp,
			Packed: true, Coded: !uncoded,
			PLUT: plut, PDAT: shape[off+0x10:],
			PIXC: 0x1F001F00,
		}
		return cel.Image()
	}
	rowBytes := ((w*bpp + 31) / 32) * 4
	if rowBytes < 8 {
		rowBytes = 8
	}
	start := off + 0x10
	if start+h*rowBytes > len(shape) {
		return nil, fmt.Errorf("SPoT pixels truncated (%dx%d@%dbpp at 0x%X)", w, h, bpp, off)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := shape[start+y*rowBytes:]
		for x := 0; x < w; x++ {
			var c uint16
			if bpp == 16 {
				c = be16(row[2*x:])
			} else {
				bit := x * bpp
				v := 0
				for k := 0; k < bpp; k++ {
					v = v<<1 | int(row[(bit+k)/8]>>(7-(bit+k)%8)&1)
				}
				c = plut[v&(len(plut)-1)]
			}
			if c&0x7FFF == 0 {
				// PDEC: resolved black is transparent (the CCB template the
				// car loader stamps has CCB_BGND clear) — the cutout around
				// the wheel discs.
				continue
			}
			img.SetRGBA(x, y, threedo.RGB555(c))
		}
	}
	return img, nil
}

// ParseShapeInto decodes a SHPM shape's textures and !ori face map into lod —
// exported for the track packets' (model, shape) scenery pairs, which use the
// same binding as the car families.
func ParseShapeInto(shape []byte, lod *CarLOD) error { return parseCarShape(shape, lod) }

// SpotOffsets returns the absolute file offsets of the SPoT records of the
// SHPM shape paired with the ORI3 at oriOff (the next SHPM in the file) —
// used by geomoracle to match the game's live texset pointers.
func SpotOffsets(fam []byte, oriOff int) []int {
	shp := -1
	for o := oriOff; o+4 <= len(fam); o++ {
		if string(fam[o:o+4]) == "SHPM" {
			shp = o
			break
		}
	}
	if shp < 0 {
		return nil
	}
	count := int(be32(fam[shp+8:]))
	if count <= 0 || count > 1024 {
		return nil
	}
	var offs []int
	for k := 0; k < count; k++ {
		name := string(fam[shp+0x10+8*k : shp+0x10+8*k+4])
		if name == "!ori" || name == "!ORI" {
			continue
		}
		off := shp + int(be32(fam[shp+0x14+8*k:]))
		if off < len(fam) && fam[off] == 0x20 {
			continue // shared plt palette chunk, not a texture
		}
		offs = append(offs, off)
	}
	return offs
}
