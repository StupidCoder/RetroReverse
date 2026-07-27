package main

// The object layer, and the sixteen scene sets each world is cut into.
//
// A terrain chunk carries placed objects (extract/uvct), and each one names
// its model by UVMD ordinal and carries a 16-bit `mask`. The engine draws an
// object only where `mask & selector` is non-zero, for a selector the game
// sets per scene: the same island is dressed differently for each mission.
// What the selector's bits mean is not traced, so the export does not choose
// for the viewer — every placement ships ONCE, carrying its mask bits as
// level-variant membership ("set-01".."set-16"), and the level declares one
// variant per bit. A world with every object at once is a world full of rings
// and balloons that never coexist in the game.

import (
	"fmt"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/tools/lib/retrox/schema"
)

// MaskBits is the width of an object's scene mask.
const MaskBits = 16

func setID(bit int) string { return fmt.Sprintf("set-%02d", bit+1) }

// rotY is the row-vector matrix taking game space (Z up) to glTF space (Y up):
// (x, y, z) -> (x, z, -y). Its inverse is its transpose.
var rotY = uvmd.Matrix{
	{1, 0, 0, 0},
	{0, 0, -1, 0},
	{0, 1, 0, 0},
	{0, 0, 0, 1},
}

var rotYInv = uvmd.Matrix{
	{1, 0, 0, 0},
	{0, 0, 1, 0},
	{0, -1, 0, 0},
	{0, 0, 0, 1},
}

// mul multiplies row-vector matrices: a point p transforms as p*a*b.
func mul(a, b uvmd.Matrix) uvmd.Matrix {
	var m uvmd.Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[i][k] * b[k][j]
			}
			m[i][j] = s
		}
	}
	return m
}

// worldMatrix composes an object's transform for the viewer.
//
// The game's matrices are row-vector (a point is a row, translation in row 3),
// and the object's own pose block gives the placement in its chunk's space; the
// chunk's cell transform lifts that into the world. The exported model GLB is
// already rotated to Y-up, so its vertices are v*R and the placement matrix
// must be R^-1 * M * R, not M — otherwise the rotation is applied in the wrong
// basis and only objects with no rotation look right.
//
// glTF stores a column-vector matrix column-major, which is the row-vector
// matrix listed row by row: the flatten below is not a transpose in disguise.
func worldMatrix(pose, cell uvmd.Matrix) []float64 {
	m := mul(rotYInv, mul(mul(pose, cell), rotY))
	out := make([]float64, 0, 16)
	for i := 0; i < 4; i++ {
		out = append(out, float64(m[i][0]), float64(m[i][1]), float64(m[i][2]), float64(m[i][3]))
	}
	return out
}

// worldPlacements returns a world's placements (each once, with its mask bits
// as variant membership), which mask bits are present at all, and how many
// placements were dropped for naming a triangle-less model.
//
// An object's pose block holds one matrix per part of its model, and part 0's
// is the object's placement: its translation is the position stored four
// fields later, and its 3x3 carries the object's scale and yaw — which is why
// a placement ships a full matrix rather than a decomposed TRS. Parts 1.. are
// the model's own rest poses, requantized; the exported GLB already bakes
// them, so the viewer needs pose[0] alone.
func worldPlacements(w *uvtr.World, chunks []*uvct.Chunk, assetByOrd []string, uvmdRes []int) ([]schema.Placement, [MaskBits]bool, int) {
	var placements []schema.Placement
	var present [MaskBits]bool
	dropped := 0
	id := 0
	for i := range w.Cells {
		c := &w.Cells[i]
		if !c.Present {
			continue
		}
		for _, o := range chunks[c.Chunk].Objects {
			asset := assetByOrd[o.Type]
			if asset == "" || len(o.Poses) == 0 {
				dropped++ // a model with no triangles at LOD 0
				continue
			}
			if o.Mask == 0 {
				dropped++ // never drawn in any scene; absent variants would mean "always"
				continue
			}
			var variants []string
			for bit := 0; bit < MaskBits; bit++ {
				if o.Mask&(1<<bit) != 0 {
					variants = append(variants, setID(bit))
					present[bit] = true
				}
			}
			placements = append(placements, schema.Placement{
				ID:       id,
				Object:   asset,
				Matrix:   worldMatrix(o.Poses[0], c.Matrix),
				Variants: variants,
				Props: map[string]any{
					"uvmd":  uvmdRes[o.Type],
					"type":  int(o.Type),
					"mask":  fmt.Sprintf("0x%04X", o.Mask),
					"chunk": c.Chunk,
					"cell":  []int{c.Col, c.Row},
					"pos":   []float64{float64(o.X), float64(o.Y), float64(o.Z)}, // game-space, for the info card
				},
			})
			id++
		}
	}
	return placements, present, dropped
}
