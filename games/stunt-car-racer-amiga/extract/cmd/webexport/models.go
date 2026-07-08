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

	"retroreverse.com/games/stunt-car-racer-amiga/extract/carmodel"
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
		mb, triGroups, lineGroups := buildCircuit(im, id, pal)

		name := trackNames[id]
		file := "models/" + slug(name) + ".glb"
		if id == track.DrawBridgeTrack {
			// the bridge ramps are animated at runtime ($5A794): the base mesh
			// is the lowered pose (phase 15, tri 0), one morph target raises it
			// to the top pose (phase 0, tri 15), and the weight animation is
			// the traced triangle-wave cadence
			lo, loT, loL := buildCircuit(im.Drawbridge(15), id, pal)
			hi, _, _ := buildCircuit(im.Drawbridge(0), id, pal)
			if len(lo.pos) != len(hi.pos) {
				chk(fmt.Errorf("drawbridge morph: vertex count mismatch %d vs %d", len(lo.pos), len(hi.pos)))
			}
			deltas := make([][3]float32, len(lo.pos))
			for i := range deltas {
				for c := 0; c < 3; c++ {
					deltas[i][c] = hi.pos[i][c] - lo.pos[i][c]
				}
			}
			// cadence (traced): one phase step per race frame, stretched by the
			// $EE time-base accumulator ($5DB34: 18 of 256 frames skip the
			// step), two steps held at each extreme; the race loop itself is
			// render-bound (no VBlank throttle), so the absolute rate is
			// machine-dependent — a nominal 12.5 fps A500-class race rate is
			// used for the GLB clock
			const nominalFPS = 12.5
			step := float32((1.0 / nominalFPS) * 256.0 / 238.0)
			m := &glb.MorphAnim{
				Name:    "drawbridge",
				Deltas:  deltas,
				Times:   []float32{0, 15 * step, 16 * step, 31 * step, 32 * step},
				Weights: []float32{1, 0, 0, 1, 1},
				// resting pose = the game's own first preview (phase 1, tri 14)
				Default: 14.0 / 15.0,
			}
			chk(glb.WriteMixedMorph(filepath.Join(outDir, file), lo.pos, loT, loL, m))
			out = append(out, ModelIndex{Name: name, File: file})
			fmt.Fprintf(os.Stderr, "[models] %d/10 %s (%d verts, morph)\n", id+1, filepath.Base(file), len(lo.pos))
			continue
		}
		chk(glb.WriteMixed(filepath.Join(outDir, file), mb.pos, triGroups, lineGroups))
		out = append(out, ModelIndex{Name: name, File: file})
		fmt.Fprintf(os.Stderr, "[models] %d/10 %s (%d verts)\n", id+1, filepath.Base(file), len(mb.pos))
	}
	out = append(out, exportCar(pal, outDir))
	out = append(out, exportHorizon(im, pal, outDir))
	return out
}

// buildCircuit assembles one circuit's coloured geometry (quads and decal
// lines, deterministic vertex order) from the given image — for the Draw
// Bridge the caller passes Drawbridge-patched images to build the animation
// poses.
func buildCircuit(im *track.Image, id int, pal [16][3]uint8) (*meshBuilder, []glb.TriGroup, []glb.LineGroup) {
	{
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
		// line groups: palette 9 (curbs + struts), palette 3 (curbs), and
		// palette $F — the white start/finish line the race view strokes
		// ($688FC; the preview wipes its flag, but it belongs to the circuit)
		linePal := []byte{9, 3, 0xF}
		lines := make([][][2]uint32, len(linePal))
		lineIdx := map[byte]int{9: 0, 3: 1, 0xF: 2}

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

				// the race view's white start/finish cross-line ($688FC,
				// baked flag bit 0 on the finish section's last rung)
				if q.Finish {
					lines[2] = append(lines[2], [2]uint32{lq, rq})
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
		return mb, triGroups, lineGroups
	}
}

// exportHorizon writes the race view's mountain range (track.Horizon, verified
// by cmd/horizonoracle) as models/horizon.glb: the 32 placed silhouettes
// arranged as the 360-degree ring the renderer implements — one compass yaw
// unit is 32 horizon pixels, so the full 256-unit circle is 8192 px; each
// vertex's x maps to arc length on a cylinder of radius 8192/2pi px and its y
// to height above the horizon plane (Y=0). Scale: 1 GLB unit = 32 px (1 yaw
// unit), radius ~40.7 units, faces double-sided facing the viewer inside.
func exportHorizon(im *track.Image, pal [16][3]uint8, outDir string) ModelIndex {
	place, models := im.Horizon()
	const pxPerYaw = 32.0
	const circle = 256 * pxPerYaw
	radius := circle / (2 * math.Pi)

	var pos [][3]float32
	index := map[[2]int]uint32{}
	vert := func(xpx, ypx int) uint32 {
		key := [2]int{xpx, ypx}
		if i, ok := index[key]; ok {
			return i
		}
		phi := float64(xpx) / circle * 2 * math.Pi
		i := uint32(len(pos))
		index[key] = i
		pos = append(pos, [3]float32{
			float32(radius * math.Sin(phi) / pxPerYaw),
			float32(float64(ypx) / pxPerYaw),
			float32(-radius * math.Cos(phi) / pxPerYaw),
		})
		return i
	}

	// one TriGroup per palette index used, in fixed order (5 = the grey
	// mountains; 4 appears only in the unplaced two-tone shape)
	triPal := []byte{5, 4}
	triIdx := map[byte]int{5: 0, 4: 1}
	tris := make([][][3]uint32, len(triPal))

	for _, p := range place {
		m := models[p.Model]
		xl := int(p.Yaw) * int(pxPerYaw)
		for _, f := range m.Faces {
			// chain the face's boundary edges into a vertex ring
			es := make([][2]int, len(f.Edges))
			for i, ei := range f.Edges {
				es[i] = m.Edges[ei]
			}
			ring := []int{es[0][0], es[0][1]}
			used := map[int]bool{0: true}
			for len(ring) < len(es) {
				cur := ring[len(ring)-1]
				for i, e := range es {
					if used[i] {
						continue
					}
					if e[0] == cur {
						ring = append(ring, e[1])
						used[i] = true
						break
					}
					if e[1] == cur {
						ring = append(ring, e[0])
						used[i] = true
						break
					}
				}
			}
			var ids []uint32
			for _, vi := range ring {
				v := m.Verts[vi]
				ids = append(ids, vert(xl+v[0], v[1]))
			}
			g := triIdx[f.Pal]
			for i := 2; i < len(ids); i++ {
				tris[g] = append(tris[g], [3]uint32{ids[0], ids[i-1], ids[i]})
			}
		}
	}

	var groups []glb.TriGroup
	for i, p := range triPal {
		groups = append(groups, glb.TriGroup{Tris: tris[i], Color: linColor(pal[p])})
	}
	file := "models/horizon.glb"
	chk(glb.WriteMixed(filepath.Join(outDir, file), pos, groups, nil))
	fmt.Fprintf(os.Stderr, "[models] 10/10 horizon.glb (%d placements, %d verts)\n", len(place), len(pos))
	return ModelIndex{Name: "Horizon", File: file}
}

// exportCar writes the opponent car's rest-pose model (carmodel.Rest — the
// verbatim-ported $599E2 construction run through orthographic views) as
// models/opponent-car.glb, in the same scale and colour conventions as the
// circuit GLBs, resting on Y=0. Faces stay double-sided: the engine's 2-D
// fills repaint every face each frame with no facing test.
func exportCar(pal [16][3]uint8, outDir string) ModelIndex {
	m := carmodel.Rest()
	minY := m.Verts[0].Y
	for _, v := range m.Verts {
		if v.Y < minY {
			minY = v.Y
		}
	}
	pos := make([][3]float32, len(m.Verts))
	for i, v := range m.Verts {
		pos[i] = [3]float32{
			float32(v.X * glbScale),
			float32((v.Y - minY) * glbScale),
			float32(-v.Z * glbScale),
		}
	}
	// one TriGroup per palette index used, in fixed order
	carPal := []byte{0xA, 0xF, 9, 0xC, 5, 0}
	idx := map[byte]int{}
	tris := make([][][3]uint32, len(carPal))
	for i, p := range carPal {
		idx[p] = i
	}
	for _, q := range m.Quads {
		g := idx[q.Pal]
		a, b, c, d := uint32(q.V[0]), uint32(q.V[1]), uint32(q.V[2]), uint32(q.V[3])
		tris[g] = append(tris[g], [3]uint32{a, b, c}, [3]uint32{a, c, d})
	}
	var groups []glb.TriGroup
	for i, p := range carPal {
		groups = append(groups, glb.TriGroup{Tris: tris[i], Color: linColor(pal[p])})
	}
	file := "models/opponent-car.glb"
	chk(glb.WriteMixed(filepath.Join(outDir, file), pos, groups, nil))
	fmt.Fprintf(os.Stderr, "[models] 9/10 opponent-car.glb (%d verts)\n", len(pos))
	return ModelIndex{Name: "Opponent Car", File: file}
}
