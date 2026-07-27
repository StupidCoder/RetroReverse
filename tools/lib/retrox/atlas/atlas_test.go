package atlas

import (
	"image"
	"image/color"
	"testing"
)

// solid returns a w×h image filled with c, with a marker pixel at the anchor.
func solid(w, h int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func TestPackRegistersAnchors(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	blue := color.NRGBA{0, 0, 255, 255}
	// Two frames of different sizes; anchors at the sprite's feet-centre.
	// Frame A: 10x8, anchor (5,8). Frame B: 16x12, anchor (6,12).
	packed, err := Pack([]Animation{
		{ID: "walk", Frames: []Frame{
			{Image: solid(10, 8, red), Anchor: image.Pt(5, 8)},
			{Image: solid(16, 12, blue), Anchor: image.Pt(6, 12)},
		}},
		{ID: "idle", Frames: []Frame{
			{Image: solid(4, 4, red), Anchor: image.Pt(2, 4)},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// left = max(5,6,2) = 6; right = max(10-5,16-6,4-2) = 10 → cellW 16
	// top = max(8,12,4) = 12; bottom = max(0,0,0) = 0 → cellH 12
	if packed.CellW != 16 || packed.CellH != 12 {
		t.Fatalf("cell %dx%d, want 16x12", packed.CellW, packed.CellH)
	}
	if packed.Anchor != image.Pt(6, 12) {
		t.Fatalf("anchor %v, want (6,12)", packed.Anchor)
	}
	if packed.Rows != 2 || packed.Cols != 2 {
		t.Fatalf("grid %dx%d, want 2x2", packed.Cols, packed.Rows)
	}
	b := packed.Image.Bounds()
	if b.Dx() != 32 || b.Dy() != 24 {
		t.Fatalf("image %dx%d, want 32x24", b.Dx(), b.Dy())
	}
	// Frame A occupies cell (0,0): its anchor (5,8) must land on the cell
	// anchor (6,12) → the image's top-left is at (1,4). Check a red pixel
	// just inside and transparency just outside.
	if got := packed.Image.NRGBAAt(1, 4); got != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("frame A top-left: %v", got)
	}
	if got := packed.Image.NRGBAAt(0, 4); got.A != 0 {
		t.Fatalf("expected transparent left padding, got %v", got)
	}
	// Frame B in cell (1,0): top-left at (16 + 6-6, 12-12) = (16, 0).
	if got := packed.Image.NRGBAAt(16, 0); got != (color.NRGBA{0, 0, 255, 255}) {
		t.Fatalf("frame B top-left: %v", got)
	}
	// idle in row 1: top-left at (6-2, 12+12-4) = (4, 20).
	if got := packed.Image.NRGBAAt(4, 20); got != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("idle top-left: %v", got)
	}
	// The idle row's second column is empty.
	if got := packed.Image.NRGBAAt(20, 20); got.A != 0 {
		t.Fatalf("expected empty cell, got %v", got)
	}
}

func TestPackRejectsEmpty(t *testing.T) {
	if _, err := Pack(nil); err == nil {
		t.Fatal("empty pack accepted")
	}
	if _, err := Pack([]Animation{{ID: "x"}}); err == nil {
		t.Fatal("frameless animation accepted")
	}
}

func TestPackTilesGutter(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	green := color.NRGBA{0, 255, 0, 255}
	tiles := []image.Image{solid(8, 8, red), solid(8, 8, green)}
	img, err := PackTiles(tiles, 8, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 20 || b.Dy() != 10 {
		t.Fatalf("sheet %dx%d, want 20x10 (2 tiles of pitch 10)", b.Dx(), b.Dy())
	}
	// Tile 0 body at (1,1); its extruded gutter must replicate the edge.
	if got := img.NRGBAAt(1, 1); got != red {
		t.Fatalf("tile body: %v", got)
	}
	if got := img.NRGBAAt(1, 0); got != red { // top gutter
		t.Fatalf("top gutter: %v", got)
	}
	if got := img.NRGBAAt(0, 1); got != red { // left gutter
		t.Fatalf("left gutter: %v", got)
	}
	if got := img.NRGBAAt(0, 0); got != red { // corner
		t.Fatalf("corner gutter: %v", got)
	}
	// Tile 1 body at (11,1).
	if got := img.NRGBAAt(11, 1); got != green {
		t.Fatalf("tile 1 body: %v", got)
	}
}
