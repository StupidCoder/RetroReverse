package threedo

// shpm.go decodes the ".3sh" shape files (magic "SHPM") that Need for Speed uses
// for its full-screen 16-bit artwork — car photos, the title screen, the EA
// logo, menu backdrops. Unlike a standard 3DO cel, SHPM is Electronic Arts' own
// container, so its layout is derived structurally from the disc rather than
// from platform docs; the loader trace in a later milestone confirms it.
//
// Layout, confirmed against several .3sh files on the disc:
//
//	0x00  "SHPM"                     magic
//	0x04  u32  total size            (= file length)
//	0x08  u32  shape count           (1 for the single-image files)
//	0x0C  "SPoT" sub-record:
//	         +0x00 "SPoT"
//	         +0x04 char[4] name       ("ansx", "eal2", …)
//	         +0x08 u32 header size
//	         +0x10 u16 width, u16 height
//	         +0x1C raw pixels         width*height 16-bit RGB555, big-endian
//
// The pixel body is an uncoded, unpacked RGB555 bitmap (bit 15 is the P/W
// control bit, ignored here).

import (
	"fmt"
	"image"
)

// ParseSHPM decodes the first shape of a .3sh file into an RGBA image.
func ParseSHPM(data []byte) (*image.RGBA, error) {
	if len(data) < 0x2C || string(data[0:4]) != "SHPM" {
		return nil, fmt.Errorf("threedo: not a SHPM shape")
	}
	spot := indexTag(data, "SPoT")
	if spot < 0 || spot+0x1C > len(data) {
		return nil, fmt.Errorf("threedo: SHPM has no SPoT record")
	}
	w := int(uint16(data[spot+0x10])<<8 | uint16(data[spot+0x11]))
	h := int(uint16(data[spot+0x12])<<8 | uint16(data[spot+0x13]))
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("threedo: implausible SHPM size %dx%d", w, h)
	}
	start := spot + 0x1C
	if start+w*h*2 > len(data) {
		return nil, fmt.Errorf("threedo: SHPM pixel data truncated (%dx%d, %d bytes left)", w, h, len(data)-start)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := start + (y*w+x)*2
			c := uint16(data[o])<<8 | uint16(data[o+1])
			img.SetRGBA(x, y, rgb555(c))
		}
	}
	return img, nil
}

// indexTag finds the byte offset of a 4-char tag, or -1.
func indexTag(data []byte, tag string) int {
	for i := 0; i+4 <= len(data); i++ {
		if string(data[i:i+4]) == tag {
			return i
		}
	}
	return -1
}
