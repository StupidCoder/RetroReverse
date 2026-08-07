package main

// The liquid painting-entrances ripple all the time, and the Studio can show it.
//
// Mode 1's grid builder ends by calling the ripple trigger $02125DE0 with the
// surface's own centre (w/2, h/2) — so unlike a framed painting, which only
// waves where the player touched it, these two start a wave at build time and
// never stop: mode 1 skips the decay early-out at $02125BF0 that ends the
// painting's ring.
//
// The wave is entirely in the data. $02125DE0 selects a $14-byte parameter
// record at $021277A8 by picture (mode 1 picks 12 for picture 4 and 13 for
// picture 7, $02125E20), and the step integrates it:
//
//	$021265A4  per vertex: rec+$8 (z) = $02125BB0(obj, rec+$C)
//	$021265D4  obj+$1B4 (the phase) += params+$04, once per 30 Hz tick
//	$02125D08  angle   = rec+$C * FX_Div($FFFF, params+$08) >> 12  -  phase
//	$02125C84  amp     = max(0, A - FX_Div(A, params+$0C) * rec+$C >> 12)
//	$02125CA4  z       = sin(angle) * amp        (the table at $02082214)
//
// and rec+$C is each vertex's distance from the source, filled once by
// $02125D64. So: a circular travelling wave from the centre of the surface,
// amplitude falling linearly to zero at params+$0C, wavelength params+$08.
//
// WHAT IS RECONSTRUCTED. The numbers, the grid and the wave are decoded. Two
// things are not: the game recomputes each vertex NORMAL from its displaced
// neighbours ($02125940/$02125AF0) and this takes the analytic gradient of the
// same field instead; and the oracle cannot check any of it, because actor
// 307's init dies in the picture resolver under it (paintprobe -grid), so
// there is no run to compare against. This is a reimplementation from the
// disassembly, not a capture.

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"sort"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/nds/nitro"
)

// paintGridN is the mesh resolution the init picks, indexed by width nibble + 1
// — the byte table at $02127714 behind the jump table at $02126CFC. Both liquid
// entrances land on 20, so both are 20x20 vertices.
var paintGridN = map[int]int{4: 13, 5: 14, 6: 16, 7: 20, 8: 20, 9: 25, 16: 30}

// paintWave is one $14-byte record at $021277A8, in the units the game keeps.
type paintWave struct {
	amp        float64 // +$00, world units
	phasePerTick float64 // +$04, DS angle units per 30 Hz tick
	wavelength float64 // +$08, world units
	falloff    float64 // +$0C, world units at which the amplitude reaches zero
}

// paintWaves maps a picture to its record. Only the two mode-1 pictures have
// one that matters: mode 0's ripple is triggered by touch and is not exported.
var paintWaves = map[int]paintWave{
	4: {amp: 16, phasePerTick: 3640, wavelength: 300, falloff: 900}, // record 12
	7: {amp: 32, phasePerTick: 3449, wavelength: 300, falloff: 900}, // record 13
}

// rippleGrid builds the surface: the game's own grid at its own resolution, with
// the wave carried by TWO morph targets. A travelling wave is exactly a rotating
// pair —
//
//	amp(d)*sin(kd - phi) = cos(phi) * [amp(d)*sin(kd)] + (-sin phi) * [amp(d)*cos(kd)]
//
// — so the two bracketed fields are static geometry and the animation is the
// weights. No sampling error, and it loops as exactly as the phase does.
//
// Geometry is in model units at the standard object scale (1 unit = 8 world
// units), centred on the placement like the quad it replaces.
func rippleGrid(pic, wNib, hNib int, img image.Image, alpha float64) (
	pos [][3]float32, uvs [][2]float32, nrm [][3]float32,
	grp []glb.TexturedGroup, targets []glb.MorphTarget, clip glb.MorphClip, ok bool) {

	wv, has := paintWaves[pic]
	n := paintGridN[wNib+1]
	if !has || n < 2 {
		return nil, nil, nil, nil, nil, clip, false
	}
	const unit = 8.0 // world units per model unit at objScale
	w, h := float64(wNib+1)*100, float64(hNib+1)*100
	k := 2 * math.Pi / wv.wavelength

	// S and C are the two static fields, in model units.
	field := func(x, y float64) (s, c float64) {
		d := math.Hypot(x, y)
		a := wv.amp * (1 - d/wv.falloff)
		if a < 0 {
			a = 0
		}
		return a * math.Sin(k*d) / unit, a * math.Cos(k*d) / unit
	}
	at := func(i, j int) (x, y float64) {
		return w*float64(i)/float64(n-1) - w/2, h*float64(j)/float64(n-1) - h/2
	}
	S := make([]float64, n*n)
	C := make([]float64, n*n)
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			x, y := at(i, j)
			S[j*n+i], C[j*n+i] = field(x, y)
		}
	}
	dx := w / float64(n-1) / unit // model-unit grid spacing
	dy := h / float64(n-1) / unit
	// Central differences on the grid, one-sided at the border: the normal of a
	// height field is (-dz/dx, -dz/dy, 1), and it stays linear in S and C, so
	// each field's gradient is its target's NORMAL delta.
	grad := func(f []float64, i, j int) (gx, gy float64) {
		i0, i1 := max(0, i-1), min(n-1, i+1)
		j0, j1 := max(0, j-1), min(n-1, j+1)
		gx = (f[j*n+i1] - f[j*n+i0]) / (float64(i1-i0) * dx)
		gy = (f[j1*n+i] - f[j0*n+i]) / (float64(j1-j0) * dy)
		return gx, gy
	}
	var tS, tC glb.MorphTarget
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			x, y := at(i, j)
			pos = append(pos, [3]float32{float32(x / unit), float32(y / unit), 0})
			uvs = append(uvs, [2]float32{float32(i) / float32(n-1), 1 - float32(j)/float32(n-1)})
			nrm = append(nrm, [3]float32{0, 0, 1})
			tS.Pos = append(tS.Pos, [3]float32{0, 0, float32(S[j*n+i])})
			tC.Pos = append(tC.Pos, [3]float32{0, 0, float32(C[j*n+i])})
			sx, sy := grad(S, i, j)
			cx, cy := grad(C, i, j)
			tS.Normal = append(tS.Normal, [3]float32{float32(-sx), float32(-sy), 0})
			tC.Normal = append(tC.Normal, [3]float32{float32(-cx), float32(-cy), 0})
		}
	}
	var tris [][3]uint32
	for j := 0; j < n-1; j++ {
		for i := 0; i < n-1; i++ {
			a := uint32(j*n + i)
			b, c, d := a+1, a+uint32(n)+1, a+uint32(n)
			tris = append(tris, [3]uint32{a, b, c}, [3]uint32{a, c, d})
		}
	}
	grp = []glb.TexturedGroup{{
		Tris: tris, Image: img, Opaque: true, Matcap: true, Alpha: alpha,
		WrapS: 33071, WrapT: 33071,
	}}
	targets = []glb.MorphTarget{tS, tC}

	// One keyframe per 30 Hz tick over the phase's own period. The phase is a
	// 16-bit angle stepped by phasePerTick, so a full turn takes 65536/step
	// ticks — 18.00 for picture 4 and 19.00 for picture 7. The last keyframe is
	// forced equal to the first: the residue is 16 and 5 angle units, 0.09 and
	// 0.03 degrees, which is not worth a visible seam.
	steps := int(math.Round(0x10000 / wv.phasePerTick))
	clip = glb.MorphClip{Name: "ripple"}
	for i := 0; i <= steps; i++ {
		phi := 2 * math.Pi * float64(i) * wv.phasePerTick / 0x10000
		if i == steps {
			phi = 0
		}
		clip.Times = append(clip.Times, float32(i)/30)
		clip.Weights = append(clip.Weights, []float32{float32(math.Cos(phi)), float32(-math.Sin(phi))})
	}
	return pos, uvs, nrm, grp, targets, clip, true
}

// rippleAsset maps a painting parameter word to the asset id of the rippling
// surface built for it, when one was.
var rippleAsset = map[int]string{}

// exportRippleGLBs builds one surface per liquid entrance the levels actually
// place. They cannot share a model: the wave is in WORLD units and the grid is
// sized in them, so the same picture at two sizes is two meshes (the Hazy Maze
// Cave portal is 800 in the basement and 700 in the cave).
func exportRippleGLBs(ctx *cli.Context, ls *sm64ds.LevelSet, tmp string, bindings map[int][]Binding) error {
	b := ctx.Builder
	seen := map[int]bool{}
	var pars []int
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		for _, ob := range lv.Objects {
			if ob.Actor == paintingActor && paintingModeEnvMapped(ob.Params[0]) && !seen[ob.Params[0]] {
				seen[ob.Params[0]] = true
				pars = append(pars, ob.Params[0])
			}
		}
	}
	sort.Ints(pars)
	n := 0
	for _, par1 := range pars {
		src := ""
		for _, bd := range bindings[paintingActor] {
			if len(bd.Models) > 0 && bd.Params[0] == par1 {
				src = bd.Models[0]
				break
			}
		}
		if src == "" {
			continue
		}
		m, err := sm64ds.LoadBMDBlank(
			filepath.Join(tmp, "files", "data", "picture", src+".bmd"), nitro.BlankBaseColor)
		if err != nil || len(m.Mats) == 0 {
			continue
		}
		tex, ok := m.Texs[m.Mats[0].Texture]
		if !ok || tex.Img == nil {
			continue
		}
		alpha := 0.0
		if a := paintingDrawAlpha(par1); a < 0x1F {
			alpha = float64(a) / 31
		}
		wNib, hNib := par1&0xF, par1>>4&0xF
		pos, uvs, nrm, grp, targets, clip, ok := rippleGrid(par1>>8&0x1F, wNib, hNib, tex.Img, alpha)
		if !ok {
			continue
		}
		stem := fmt.Sprintf("%s_r%d", src, (wNib+1)*100)
		gp, err := b.Path("objects", stem+".glb")
		if err != nil {
			return err
		}
		if err := glb.WriteTexturedMorph(gp, pos, uvs, nrm, grp, targets, clip); err != nil {
			return err
		}
		name := fmt.Sprintf("%s (rippling)", title(src))
		id := objectID(stem)
		b.AddObject(schema.Asset{ID: id, Name: name, Group: "Other models"},
			&schema.Object{
				Type: schema.ObjectModel3D, Name: name, Model: stem + ".glb",
				SkinnedClone: true,
				Animations: []schema.Animation{{
					ID: clip.Name, Clip: clip.Name, Loop: "loop", FPS: 30,
				}},
			})
		refs[stem] = id
		rippleAsset[par1] = stem
		n++
	}
	ctx.Logf("%d rippling liquid surfaces built", n)
	return nil
}
