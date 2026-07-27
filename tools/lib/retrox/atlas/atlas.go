// Package atlas builds Retro-X sprite atlases: one PNG per object where each
// row is one animation and each column one frame, on a uniform cell grid.
//
// Input frames may have different sizes; what keeps an animation registered
// is each frame's ANCHOR (its game-space origin, e.g. the sprite's draw
// position or feet). The packer computes the smallest cell that fits every
// frame with all anchors coincident, composites frames so every anchor lands
// on the common cell anchor, and reports that anchor for the object document.
package atlas

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
)

// Frame is one input frame: an image plus its anchor point in image pixels.
type Frame struct {
	Image  image.Image
	Anchor image.Point
}

// Animation is one atlas row.
type Animation struct {
	ID     string
	Frames []Frame
}

// Packed is the result of Pack.
type Packed struct {
	Image  *image.NRGBA
	CellW  int
	CellH  int
	Anchor image.Point // common anchor within every cell
	Rows   int         // == len(animations)
	Cols   int         // widest animation
}

// Pack lays out the animations row per animation, column per frame.
func Pack(anims []Animation) (*Packed, error) {
	if len(anims) == 0 {
		return nil, fmt.Errorf("atlas: no animations")
	}
	// The common cell must fit, for every frame, anchor→left/right/top/bottom.
	var left, right, top, bottom, cols int
	for _, a := range anims {
		if len(a.Frames) == 0 {
			return nil, fmt.Errorf("atlas: animation %q has no frames", a.ID)
		}
		if len(a.Frames) > cols {
			cols = len(a.Frames)
		}
		for i, f := range a.Frames {
			if f.Image == nil {
				return nil, fmt.Errorf("atlas: animation %q frame %d has no image", a.ID, i)
			}
			b := f.Image.Bounds()
			ax, ay := f.Anchor.X-b.Min.X, f.Anchor.Y-b.Min.Y
			left = max(left, ax)
			right = max(right, b.Dx()-ax)
			top = max(top, ay)
			bottom = max(bottom, b.Dy()-ay)
		}
	}
	cellW, cellH := left+right, top+bottom
	if cellW <= 0 || cellH <= 0 {
		return nil, fmt.Errorf("atlas: degenerate cell %dx%d", cellW, cellH)
	}
	out := image.NewNRGBA(image.Rect(0, 0, cellW*cols, cellH*len(anims)))
	for row, a := range anims {
		for col, f := range a.Frames {
			b := f.Image.Bounds()
			ax, ay := f.Anchor.X-b.Min.X, f.Anchor.Y-b.Min.Y
			// Place so the frame's anchor lands on the cell anchor (left, top).
			dst := image.Rect(0, 0, b.Dx(), b.Dy()).
				Add(image.Pt(col*cellW+left-ax, row*cellH+top-ay))
			draw.Draw(out, dst, f.Image, b.Min, draw.Over)
		}
	}
	return &Packed{
		Image:  out,
		CellW:  cellW,
		CellH:  cellH,
		Anchor: image.Pt(left, top),
		Rows:   len(anims),
		Cols:   cols,
	}, nil
}

// EncodePNG writes the packed atlas.
func (p *Packed) EncodePNG(w io.Writer) error { return png.Encode(w, p.Image) }

// PackTiles builds a tile-sheet PNG for tilemap atlases: fixed-size tiles,
// `cols` per row, each tile optionally extruded by `gutter` pixels on every
// side (edge-pixel bleed guard for GPU sampling).
func PackTiles(tiles []image.Image, tileSize, cols, gutter int) (*image.NRGBA, error) {
	if len(tiles) == 0 {
		return nil, fmt.Errorf("atlas: no tiles")
	}
	if tileSize <= 0 || cols <= 0 {
		return nil, fmt.Errorf("atlas: need positive tileSize and cols")
	}
	pitch := tileSize + 2*gutter
	rows := (len(tiles) + cols - 1) / cols
	out := image.NewNRGBA(image.Rect(0, 0, cols*pitch, rows*pitch))
	for i, t := range tiles {
		if t == nil {
			continue
		}
		b := t.Bounds()
		if b.Dx() != tileSize || b.Dy() != tileSize {
			return nil, fmt.Errorf("atlas: tile %d is %dx%d, want %dx%d", i, b.Dx(), b.Dy(), tileSize, tileSize)
		}
		ox, oy := (i%cols)*pitch+gutter, (i/cols)*pitch+gutter
		draw.Draw(out, image.Rect(ox, oy, ox+tileSize, oy+tileSize), t, b.Min, draw.Src)
		for g := 1; g <= gutter; g++ {
			// Extrude edges outward: rows above/below, then full columns.
			draw.Draw(out, image.Rect(ox, oy-g, ox+tileSize, oy-g+1), t, b.Min, draw.Src)
			draw.Draw(out, image.Rect(ox, oy+tileSize+g-1, ox+tileSize, oy+tileSize+g), t, image.Pt(b.Min.X, b.Max.Y-1), draw.Src)
		}
		for g := 1; g <= gutter; g++ {
			src := out.SubImage(image.Rect(ox, oy-gutter, ox+1, oy+tileSize+gutter)).(*image.NRGBA)
			draw.Draw(out, image.Rect(ox-g, oy-gutter, ox-g+1, oy+tileSize+gutter), src, src.Bounds().Min, draw.Src)
			src = out.SubImage(image.Rect(ox+tileSize-1, oy-gutter, ox+tileSize, oy+tileSize+gutter)).(*image.NRGBA)
			draw.Draw(out, image.Rect(ox+tileSize+g-1, oy-gutter, ox+tileSize+g, oy+tileSize+gutter), src, src.Bounds().Min, draw.Src)
		}
	}
	return out, nil
}
