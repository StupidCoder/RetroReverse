package clv

// texture.go decodes the stage's texture bank. A batch's material record
// points to a texture-reference record (the material's UV scale and the
// texture object), and the texture object describes the pixel data: CLUT4
// texels in the GE's swizzled block layout, preceded by their 16-entry
// RGBA8888 palette, with the mip chain stored consecutively.

import (
	"encoding/binary"
	"fmt"
	"image"
)

// Material is a batch's resolved material: the shading-group name, the
// material colour, the UV scale, and the texture object.
type Material struct {
	Name           string
	Color          uint32 // RGBA as stored (byte0 = R)
	UScale, VScale float32
	TexName        string
	TexW, TexH     int
	TexFmt         int    // texture object format byte (1 = CLUT4)
	Mips           int    // mip level count
	ClutOff        uint32 // 16 x RGBA8888 palette
	PixOff         uint32 // mip 0 texels (CLUT4, swizzled)
}

// Material resolves a batch's material chain.
func (c *Clv) Material(b Batch) (Material, error) {
	m := Material{Name: b.MaterialName, Color: b.Color, UScale: 1, VScale: 1}
	if b.MaterialOff == 0 {
		return m, fmt.Errorf("clv: batch has no material")
	}
	ref := c.ptr(b.MaterialOff + 12)
	if ref == 0 {
		// untextured material (flat colour: water, hidden shapes, plain fills)
		return m, nil
	}
	// texture reference: {ptr name, u32 0, u32 0, u32 flags, u32 0, u32 0,
	// float uscale, float vscale, ptr texture object}
	if n := c.ptr(ref); n != 0 {
		m.TexName = c.cstr(n)
	}
	m.UScale = c.f32(ref + 0x18)
	m.VScale = c.f32(ref + 0x1C)
	obj := c.ptr(ref + 0x20)
	if obj == 0 {
		return m, fmt.Errorf("clv: texture ref %q has no texture object", m.TexName)
	}
	// texture object: {ptr source path, u32 0, u16 w, u16 h, u8 fmt, u8 mips,
	// u16 0, ptr clut+pixels, ptr next}
	m.TexW = int(binary.LittleEndian.Uint16(c.Data[obj+8 : obj+10]))
	m.TexH = int(binary.LittleEndian.Uint16(c.Data[obj+10 : obj+12]))
	m.TexFmt = int(c.Data[obj+12])
	m.Mips = int(c.Data[obj+13])
	data := c.ptr(obj + 16)
	if data == 0 {
		return m, fmt.Errorf("clv: texture %q has no data", m.TexName)
	}
	m.ClutOff = data
	m.PixOff = data + 0x40
	return m, nil
}

// DecodeTexture renders a material's mip-0 texels to an RGBA image. Formats 1
// and 3 share the CLUT4 + RGBA8888-palette storage (the engine binds every
// level texture as CLUT4 with an 8888 palette); format 3 marks the translucent
// class — its palette entries carry partial alpha.
func (c *Clv) DecodeTexture(m Material) (*image.RGBA, error) {
	if m.TexFmt != 1 && m.TexFmt != 3 {
		return nil, fmt.Errorf("clv: texture %q has unhandled format %d", m.TexName, m.TexFmt)
	}
	if m.TexW <= 0 || m.TexH <= 0 || m.TexW > 1024 || m.TexH > 1024 {
		return nil, fmt.Errorf("clv: texture %q has bad size %dx%d", m.TexName, m.TexW, m.TexH)
	}
	var clut [16][4]uint8
	for i := range clut {
		o := m.ClutOff + uint32(i)*4
		clut[i] = [4]uint8{c.Data[o], c.Data[o+1], c.Data[o+2], c.Data[o+3]}
	}
	img := image.NewRGBA(image.Rect(0, 0, m.TexW, m.TexH))
	rowBytes := uint32(m.TexW / 2) // CLUT4: two texels per byte
	for y := 0; y < m.TexH; y++ {
		for x := 0; x < m.TexW; x++ {
			off := swizzleOff(uint32(x/2), uint32(y), rowBytes)
			raw := c.Data[m.PixOff+off] >> (4 * (uint32(x) & 1)) & 0xF
			p := clut[raw]
			i := img.PixOffset(x, y)
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = p[0], p[1], p[2], p[3]
		}
	}
	return img, nil
}

// swizzleOff maps a (byte-x, y) texel position through the GE's swizzled
// block layout: 16-byte x 8-row blocks stored contiguously.
func swizzleOff(xb, y, rowBytes uint32) uint32 {
	if rowBytes < 16 {
		return y*rowBytes + xb
	}
	rowBlocks := rowBytes / 16
	block := (y/8)*rowBlocks + xb/16
	return block*128 + (y%8)*16 + xb%16
}
