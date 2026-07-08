package gfx

import "image"

// RenderBitmapMC renders a C64 multicolor bitmap screen (160x200 colour
// pixels, shown as 320x200 with each colour pixel doubled) at pixel scale s.
//
//	bitmap  : 8000 bytes of VIC bitmap data ($2000 region), cell-row major,
//	          8 bytes per 8x8 cell.
//	matrix  : 1000 bytes of video-matrix colour (high nibble -> bit-pair 01,
//	          low nibble -> bit-pair 10).
//	colorRAM: 1000 bytes of colour RAM (low nibble -> bit-pair 11).
//	bg      : the background colour ($D021) used for bit-pair 00.
//
// Shorter slices are tolerated (missing cells render as background).
func RenderBitmapMC(bitmap, matrix, colorRAM []byte, bg byte, s int) *image.RGBA {
	const cols, rows = 40, 25
	img := image.NewRGBA(image.Rect(0, 0, 320*s, 200*s))
	bgCol := Palette[bg&0x0F]
	at := func(b []byte, i int) byte {
		if i < len(b) {
			return b[i]
		}
		return 0
	}
	for cy := 0; cy < rows; cy++ {
		for cx := 0; cx < cols; cx++ {
			cell := cy*cols + cx
			m := at(matrix, cell)
			pair := [4]int{int(bg & 0x0F), int(m >> 4), int(m & 0x0F), int(at(colorRAM, cell) & 0x0F)}
			for line := 0; line < 8; line++ {
				b := at(bitmap, cell*8+line)
				px := cx * 8
				py := cy*8 + line
				for p := 0; p < 4; p++ { // four colour pixels per byte
					bits := (b >> uint(6-2*p)) & 3
					c := bgCol
					if bits != 0 {
						c = Palette[pair[bits]&0x0F]
					}
					// each colour pixel is two hires pixels wide
					for dy := 0; dy < s; dy++ {
						for dx := 0; dx < 2*s; dx++ {
							img.SetRGBA((px+p*2)*s+dx, py*s+dy, c)
						}
					}
				}
			}
		}
	}
	return img
}
