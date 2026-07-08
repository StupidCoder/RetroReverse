// models.go — the "models" stage: one GLB per circuit, built from the verified
// track geometry (track.Geometry) coloured exactly as the game's pre-race
// preview renders it (track.Mesh, byte-verified by cmd/coloracle): the road
// ribbon in the two alternating greys (background-colour spans up to a crease),
// side walls white/red by section parity dropping to the ground plane (inner
// faces in the dark back-side red), and the stroked decal lines as true glTF
// LINES primitives — curb strokes alternating palette 9/3 per rung and the
// wall-vertical struts on type-9 rungs.
//
// Coordinates: glTF Y-up. X = wx*s, Y = (h-$200)/4*s, Z = -wz*s with s = 1/1024
// (one $800 grid cell = 2 units). The 1:4 height:plan ratio is the engine's own
// ($624C2 heights ASR#2 into the same projection the plan enters unscaled); the
// preview's extra $4C1B vertical squeeze is a screen transform and is NOT baked
// in. $200 is the wall-bottom ground height $654C2 plants ($1BB68=$80), so the
// ground sits at Y=0. Z is negated to keep the circuit plan un-mirrored in the
// right-handed glTF frame (same convention as the stunt-track viewer).
//
// Colours: 4-bit displayed channels (Palette, already through the copper-push
// 2c|1 map) -> sRGB c/15 -> linearised into baseColorFactor, so an sRGB-output
// renderer reproduces the exact Amiga RGB.
package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/track"
	"retroreverse.com/tools/lib/glb"
)

// ModelIndex is one manifest models[] entry (STANDARDS.md §4.2).
type ModelIndex struct {
	Name string `json:"name"`
	File string `json:"file"`
}

const (
	glbScale  = 1.0 / 1024 // one $800 grid cell = 2 GLB units
	groundH   = 0x200      // wall-bottom height the preview plants ($654C2)
	heightDiv = 4          // the engine's height:plan ratio ($624C2 ASR #2)
)

// linColor converts a displayed 4-bit-per-channel Amiga colour to a linear-light
// baseColorFactor (sRGB EOTF).
func linColor(c [3]uint8) [3]float32 {
	var out [3]float32
	for i, v := range c {
		s := float64(v) / 15
		if s <= 0.04045 {
			out[i] = float32(s / 12.92)
		} else {
			out[i] = float32(math.Pow((s+0.055)/1.055, 2.4))
		}
	}
	return out
}

// meshBuilder accumulates deduplicated positions plus triangle and line groups
// keyed by palette index, in deterministic first-use order.
type meshBuilder struct {
	pos   [][3]float32
	index map[[3]int]uint32
}

func newMeshBuilder() *meshBuilder {
	return &meshBuilder{index: map[[3]int]uint32{}}
}

// vert returns the GLB-space vertex index for game-world (x, z, h).
func (b *meshBuilder) vert(x, z, h int) uint32 {
	key := [3]int{x, z, h}
	if i, ok := b.index[key]; ok {
		return i
	}
	i := uint32(len(b.pos))
	b.index[key] = i
	b.pos = append(b.pos, [3]float32{
		float32(x) * glbScale,
		float32(h-groundH) / heightDiv * glbScale,
		-float32(z) * glbScale,
	})
	b.index[key] = i
	return i
}

func (b *meshBuilder) xyz(i uint32) [3]float32 { return b.pos[i] }

// quad appends a,b,c,d as two triangles. If outward is non-nil the winding is
// chosen so the face normal points along it (for single-sided walls).
func quadTris(mb *meshBuilder, tris *[][3]uint32, a, b, c, d uint32, outward *[3]float32) {
	if outward != nil {
		pa, pb, pc := mb.xyz(a), mb.xyz(b), mb.xyz(c)
		e1 := [3]float32{pb[0] - pa[0], pb[1] - pa[1], pb[2] - pa[2]}
		e2 := [3]float32{pc[0] - pa[0], pc[1] - pa[1], pc[2] - pa[2]}
		n := [3]float32{
			e1[1]*e2[2] - e1[2]*e2[1],
			e1[2]*e2[0] - e1[0]*e2[2],
			e1[0]*e2[1] - e1[1]*e2[0],
		}
		if n[0]*outward[0]+n[1]*outward[1]+n[2]*outward[2] < 0 {
			b, d = d, b
		}
	}
	*tris = append(*tris, [3]uint32{a, b, c}, [3]uint32{a, c, d})
}

// exportModels writes models/<slug>.glb for the eight circuits and returns the
// manifest entries.
func exportModels(inPath, outDir string) []ModelIndex {
	img, err := os.ReadFile(inPath)
	chk(err)
	im := track.New(img)
	chk(os.MkdirAll(filepath.Join(outDir, "models"), 0o755))
	pal := im.Palette()

	var out []ModelIndex
	for id := 0; id < 8; id++ {
		t := im.Spine(id)
		geo := im.Geometry(&t)
		mesh := im.Mesh(&t)

		mb := newMeshBuilder()
		// triangle groups: road greys 1/2, crease background 0, outer walls
		// $F/$A, inner walls $9 — fixed order for byte-stable output
		triPal := []byte{1, 2, 0, 0xF, 0xA, 9}
		triIdx := map[byte]int{}
		tris := make([][][3]uint32, len(triPal))
		for i, p := range triPal {
			triIdx[p] = i
		}
		const innerWall = 5 // tris[] slot for the $9 inner faces
		// line groups: palette 9 (curbs + struts), palette 3 (curbs)
		linePal := []byte{9, 3}
		lines := make([][][2]uint32, len(linePal))
		lineIdx := map[byte]int{9: 0, 3: 1}

		for sec := range geo {
			rs := geo[sec]
			ms := mesh.Sections[sec].Rungs
			for k := 1; k < len(rs); k++ {
				p, q := rs[k-1], rs[k]
				mr := ms[k]

				lp := mb.vert(p.LX, p.LZ, p.HL)
				rp := mb.vert(p.RX, p.RZ, p.HR)
				lq := mb.vert(q.LX, q.LZ, q.HL)
				rq := mb.vert(q.RX, q.RZ, q.HR)
				lpg := mb.vert(p.LX, p.LZ, groundH)
				rpg := mb.vert(p.RX, p.RZ, groundH)
				lqg := mb.vert(q.LX, q.LZ, groundH)
				rqg := mb.vert(q.RX, q.RZ, groundH)

				// road quad (double-sided, the fill has no facing)
				g := triIdx[mr.RoadPal]
				quadTris(mb, &tris[g], lp, rp, rq, lq, nil)

				// walls: outer face in the section colour, inner face $9.
				// outward = away from the other rail, in GLB space
				pl, pr := mb.xyz(lp), mb.xyz(rp)
				outL := [3]float32{pl[0] - pr[0], 0, pl[2] - pr[2]}
				outR := [3]float32{pr[0] - pl[0], 0, pr[2] - pl[2]}
				inL := [3]float32{-outL[0], 0, -outL[2]}
				inR := [3]float32{-outR[0], 0, -outR[2]}
				w := triIdx[mr.WallPal]
				quadTris(mb, &tris[w], lp, lq, lqg, lpg, &outL)
				quadTris(mb, &tris[w], rp, rq, rqg, rpg, &outR)
				quadTris(mb, &tris[innerWall], lp, lq, lqg, lpg, &inL)
				quadTris(mb, &tris[innerWall], rp, rq, rqg, rpg, &inR)

				// curb strokes: the road-edge lengthwise lines, colour = type
				cg := lineIdx[mr.Type]
				lines[cg] = append(lines[cg], [2]uint32{lp, lq}, [2]uint32{rp, rq})

				// wall-vertical struts at rung k, stroked palette 9 on type-9 strips
				if mr.VertS {
					lines[0] = append(lines[0], [2]uint32{lq, lqg}, [2]uint32{rq, rqg})
				}
			}
		}

		var triGroups []glb.TriGroup
		for i, p := range triPal {
			c := linColor(pal[p])
			triGroups = append(triGroups, glb.TriGroup{
				Tris:        tris[i],
				Color:       c,
				SingleSided: i >= 3, // walls: outer and inner faces are one-sided
			})
		}
		var lineGroups []glb.LineGroup
		for i, p := range linePal {
			lineGroups = append(lineGroups, glb.LineGroup{Lines: lines[i], Color: linColor(pal[p])})
		}

		name := trackNames[id]
		file := "models/" + slug(name) + ".glb"
		chk(glb.WriteMixed(filepath.Join(outDir, file), mb.pos, triGroups, lineGroups))
		out = append(out, ModelIndex{Name: name, File: file})
		fmt.Fprintf(os.Stderr, "[models] %d/8 %s (%d verts)\n", id+1, filepath.Base(file), len(mb.pos))
	}
	return out
}
