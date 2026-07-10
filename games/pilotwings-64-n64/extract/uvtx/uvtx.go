// Package uvtx decodes Pilotwings 64's texture resources.
//
// A UVTX resource is one COMM chunk (usually GZIP-compressed; 23 of the 463
// ship raw). Its layout, read off the game's own reader at 0x80226A54 — which
// pulls the fields one at a time through a cursor helper at 0x80225394:
//
//	u16 dataSize      texel bytes that follow the header. The reader clamps this
//	                  to 4096 and logs an error above that, so 4 KiB — one TMEM —
//	                  is the hard limit on a Pilotwings texture.
//	u16 formatCode    an authoring format code (12 distinct values). The reader
//	                  parses and discards it; it is redundant with the template
//	                  below, which is what the RDP actually consumes.
//	u32 x4            read and discarded. Zero in most files, but not all.
//	u8  texels[dataSize]
//	... a Fast3D display list: the material template
//
// The template is the authority on how the texels are interpreted. It is an
// ordinary display list — `G_TEXTURE`, othermode words, `G_SETTIMG` with a
// **zero address word the loader patches** once the texels land in RDRAM, a
// `Load_Block` into tile 7, then `Set_Tile`/`Set_Tile_Size` pairs describing up
// to six mip levels — terminated by `G_ENDDL`. All 463 templates in the archive
// parse to completion, none loads a palette, and every `Load_Block` uses dxt=0,
// so the texels ship pre-swizzled exactly as TMEM wants them.
//
// Decoding therefore means: walk the template with the Fast3D interpreter in
// extract/f3d, take the tile `G_TEXTURE` selects for drawing, and read the
// texels through it. That reuses the sampler semantics the oracle's RDP was
// verified against, rather than writing a second, unverified texel reader.
package uvtx

import (
	"encoding/binary"
	"fmt"
	"image"

	"retroreverse.com/games/pilotwings-64-n64/extract/f3d"
)

// HeaderSize is the fixed header ahead of the texels.
const HeaderSize = 20

// maxDataSize is the clamp the game's reader applies (and complains about).
const maxDataSize = 4096

// Texture is a decoded UVTX resource.
type Texture struct {
	Image *image.RGBA

	DataSize   int    // texel bytes
	FormatCode uint16 // the header's authoring code, kept for the record
	Fmt, Siz   uint32 // the RDP format/size the template's draw tile declares
	Width      int
	Height     int
	Levels     int // mip levels the template configures

	// Tile is the drawing tile as the material template left it, with Img
	// pointing at the texels inside the chunk and the extent resolved. A frame
	// may re-issue Set_Tile_Size to sample a *window* of this texture (the ocean
	// wraps a 65x65 window over a 64x64 image), but never changes the format,
	// line stride or wrap bits — which makes those the fields worth checking
	// against the running game.
	Tile f3d.TileDesc
}

// Format names the RDP texel format, for listings.
func (t Texture) Format() string {
	names := map[[2]uint32]string{
		{0, 2}: "RGBA16", {0, 3}: "RGBA32",
		{2, 0}: "CI4", {2, 1}: "CI8",
		{3, 0}: "IA4", {3, 1}: "IA8", {3, 2}: "IA16",
		{4, 0}: "I4", {4, 1}: "I8",
	}
	if n, ok := names[[2]uint32{t.Fmt, t.Siz}]; ok {
		return n
	}
	return fmt.Sprintf("fmt%d/siz%d", t.Fmt, t.Siz)
}

func be16(b []byte, o int) uint16 { return binary.BigEndian.Uint16(b[o:]) }

// Decode parses one UVTX COMM chunk.
func Decode(data []byte) (*Texture, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("uvtx: chunk is %d bytes, shorter than the header", len(data))
	}
	size := int(be16(data, 0))
	code := be16(data, 2)
	if size > maxDataSize {
		return nil, fmt.Errorf("uvtx: dataSize %d exceeds the game's own %d-byte clamp", size, maxDataSize)
	}
	if HeaderSize+size > len(data) {
		return nil, fmt.Errorf("uvtx: dataSize %d overruns the %d-byte chunk", size, len(data))
	}
	dl := uint32(HeaderSize + size)
	if int(dl) == len(data) {
		return nil, fmt.Errorf("uvtx: no material template after %d texel bytes", size)
	}

	// Walk the template. The chunk is the walker's whole address space: the
	// template's own G_SETTIMG carries a zero address (patched at load time), so
	// nothing here resolves outside it.
	w := f3d.New(data, false)
	w.Walk(dl)

	tile := w.Tile(int(w.TexTile()))
	// Most templates give the drawing tile a Set_Tile_Size (they configure six
	// mip levels). Thirty-two of them draw through tile 0 and set no size at
	// all: for those the extent is the tile's wrap mask, 2^mask texels on each
	// axis — which is what the RDP itself would clamp to. Deriving it any other
	// way (from `line`, or from dataSize) would guess where the hardware reads.
	// Most templates give the drawing tile a Set_Tile_Size. Where they do not,
	// the extent is recovered in the order the hardware would find it:
	//
	//  1. a sibling tile at the same TMEM base and line stride that *does* carry
	//     a size — that tile is the same level-0 image under another index (the
	//     mip-mapped templates re-declare level 0 as tile 0 at the end);
	//  2. failing that, the tile's own wrap mask, 2^mask texels per axis, which
	//     is where the RDP clamps. Such a template configures no mip levels, so
	//     the image must then account for every stored texel byte exactly —
	//     asserted, because a wrong extent would otherwise decode happily.
	if tile.SH <= tile.SL && tile.TH <= tile.TL {
		sized := false
		for i := 0; i < 8; i++ {
			s := w.Tile(i)
			if s.Tmem == tile.Tmem && s.Line == tile.Line && (s.SH > s.SL || s.TH > s.TL) {
				tile.SL, tile.TL, tile.SH, tile.TH = s.SL, s.TL, s.SH, s.TH
				sized = true
				break
			}
		}
		if !sized {
			if tile.MaskS == 0 || tile.MaskT == 0 {
				return nil, fmt.Errorf("uvtx: drawing tile has neither a size, a sized sibling, nor a wrap mask")
			}
			tile.SL, tile.TL = 0, 0
			tile.SH = uint32(1<<tile.MaskS-1) << 2
			tile.TH = uint32(1<<tile.MaskT-1) << 2
			if rowBytes := int(tile.Line) * 8; rowBytes*(1<<tile.MaskT) != size {
				return nil, fmt.Errorf("uvtx: mask says %dx%d (%d bytes) but %d texel bytes are stored",
					1<<tile.MaskS, 1<<tile.MaskT, rowBytes*(1<<tile.MaskT), size)
			}
		}
	}
	// Decode the texels through a base the sampler's swizzle can cope with. The
	// odd-row word-half swap is `off ^= 4` on the *absolute* byte offset, so it
	// only lands right when the texels begin on an 8-byte boundary — which they
	// do in RDRAM (the game's allocator aligns to 8) but do not in the chunk,
	// where the 20-byte header pushes them to offset 20. Copying them to an
	// aligned base reproduces the layout the RDP sees. (Offset 8 rather than 0
	// because a zero Img means "untextured" to the walker.)
	const texBase = 8
	buf := make([]byte, texBase+size)
	copy(buf[texBase:], data[HeaderSize:HeaderSize+size])
	tile.Img = texBase

	tw := f3d.New(buf, false)
	g := &f3d.Group{Tile: tile, TLUT: w.TLUT(), Scale: w.TexScale()}
	img := tw.DecodeTexture(g)
	if img == nil {
		return nil, fmt.Errorf("uvtx: no decoder for fmt=%d siz=%d (format code %#x)", tile.Fmt, tile.Size, code)
	}
	rgba, ok := img.(*image.RGBA)
	if !ok {
		return nil, fmt.Errorf("uvtx: decoder returned %T", img)
	}

	// The level-0 image must fit inside the texels the header declares. It need
	// not fill them: a template with mip levels packs every level into the same
	// block. An extent that overruns is a misread, not a smaller texture.
	if rowBytes := int(tile.Line) * 8; rowBytes > 0 {
		if need := rowBytes * (int(tile.TH>>2-tile.TL>>2) + 1); need > size {
			return nil, fmt.Errorf("uvtx: level 0 needs %d bytes but only %d texel bytes are stored", need, size)
		}
	}

	levels := 0
	for i := 0; i < 8; i++ {
		if t := w.Tile(i); t.SH > t.SL || t.TH > t.TL {
			levels++
		}
	}
	b := rgba.Bounds()
	return &Texture{
		Image: rgba, DataSize: size, FormatCode: code,
		Fmt: tile.Fmt, Siz: tile.Size, Tile: tile,
		Width: b.Dx(), Height: b.Dy(), Levels: levels,
	}, nil
}
