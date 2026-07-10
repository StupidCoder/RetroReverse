package main

// The object layer, and the sixteen sets each world is cut into.
//
// A terrain chunk carries placed objects (extract/uvct), and each one names its
// model by UVMD ordinal and carries a 16-bit `mask`. The engine draws an object
// only where `mask & selector` is non-zero, for a selector the game sets per
// scene: the same island is dressed differently for each mission. What the
// selector's bits mean is not traced, so the export does not choose for the
// viewer — it writes **one object list per mask bit**, sixteen at most, and lets
// the Studio's browse list offer them side by side. A world with every object at
// once is a world full of rings and balloons that never coexist in the game.
//
// Objects are not baked into the terrain GLB. Each set is a small JSON list of
// (model, matrix) placements over the world's single mesh, in the format-2 shape
// the shared placeObjects/installPicker helpers already read.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/pilotwings-64-n64/extract/uvct"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvmd"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
)

// MaskBits is the width of an object's scene mask.
const MaskBits = 16

// Object is one placement in a world's object list.
type Object struct {
	Model string `json:"model"`

	// Mat is the object's world transform for the Studio: a 16-element glTF
	// matrix (column-major, Y-up). See worldMatrix.
	Mat []float32 `json:"mat"`

	Type  int        `json:"type"`  // the UVMD ordinal the object draws
	UVMD  int        `json:"uvmd"`  // that model's archive resource index
	Mask  string     `json:"mask"`  // the object's raw 16-bit scene mask
	Chunk int        `json:"chunk"` // the UVCT ordinal it was placed in
	Cell  [2]int     `json:"cell"`  // that chunk's grid cell, (col, row)
	Pos   [3]float32 `json:"pos"`   // game-space position, for the info card
}

type objectDoc struct {
	Objects []Object `json:"objects"`
}

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
// already rotated to Y-up, so its vertices are v*R and the node's matrix must be
// R^-1 * M * R, not M — otherwise the rotation is applied in the wrong basis and
// only objects with no rotation look right.
//
// glTF stores a column-vector matrix column-major, which is the row-vector
// matrix listed row by row: the flatten below is not a transpose in disguise.
func worldMatrix(pose, cell uvmd.Matrix) []float32 {
	m := mul(rotYInv, mul(mul(pose, cell), rotY))
	out := make([]float32, 0, 16)
	for i := 0; i < 4; i++ {
		out = append(out, m[i][0], m[i][1], m[i][2], m[i][3])
	}
	return out
}

// worldObjects returns one object list per mask bit for a world, indexed by bit.
// A bit no object in the world carries yields a nil list.
//
// An object's pose block holds one matrix per part of its model, and part 0's is
// the object's placement: its translation is the position stored four fields
// later, and its 3x3 carries the object's scale and yaw. Parts 1.. are the
// model's own rest poses, requantized — the exported GLB already bakes them, so
// the viewer needs pose[0] alone. (Not instances: poseCount equals the model's
// part count for all 1,364 objects, and the tail agrees with the model's rest
// poses to 2^-16, the fixed-point step.)
func worldObjects(w *uvtr.World, chunks []*uvct.Chunk, modelFiles []string, uvmdRes []int) [MaskBits][]Object {
	var sets [MaskBits][]Object
	for i := range w.Cells {
		c := &w.Cells[i]
		if !c.Present {
			continue
		}
		for _, o := range chunks[c.Chunk].Objects {
			file := modelFiles[o.Type]
			if file == "" {
				continue // a model with no triangles at LOD 0; counted by the caller
			}
			if len(o.Poses) == 0 {
				continue
			}
			obj := Object{
				Model: file,
				Mat:   worldMatrix(o.Poses[0], c.Matrix),
				Type:  int(o.Type),
				UVMD:  uvmdRes[o.Type],
				Mask:  fmt.Sprintf("0x%04X", o.Mask),
				Chunk: c.Chunk,
				Cell:  [2]int{c.Col, c.Row},
				Pos:   [3]float32{o.X, o.Y, o.Z},
			}
			for b := 0; b < MaskBits; b++ {
				if o.Mask&(1<<b) != 0 {
					sets[b] = append(sets[b], obj)
				}
			}
		}
	}
	return sets
}

// writeObjectSets writes each non-empty set and returns the files it wrote, by bit.
func writeObjectSets(dir, prefix string, sets [MaskBits][]Object) (map[int]string, error) {
	files := map[int]string{}
	for b, objs := range sets {
		if len(objs) == 0 {
			continue
		}
		name := fmt.Sprintf("%s-set-%02d.objects.json", prefix, b+1)
		j, err := json.Marshal(objectDoc{Objects: objs})
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(j, '\n'), 0o644); err != nil {
			return nil, err
		}
		files[b] = name
	}
	return files, nil
}
