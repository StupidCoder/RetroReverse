package main

// atlas.go — every texture a level uses, packed into one image, so the whole
// dungeon draws in a handful of calls instead of one per texture.
//
// A primitive is a draw call, and a level was one primitive per texture: 74 on
// level 1 after the per-material merge. On a Quest 3 the draw calls were the
// scene's whole cost, so the last thing separating them had to go. Once the
// pictures live in one atlas, what is left to separate primitives is state —
// whether the material cuts out, and whether it is drawn from both sides.
//
// This only works because UW stretches a wall texture over the wall rather
// than tiling it: every UV lands inside the unit square, so a sub-rectangle of
// an atlas can stand in for a whole texture. atlasGroups checks that for
// itself and declines the job if it is ever not true, leaving the caller's
// per-texture primitives alone.

import (
	"image"
	"image/draw"
	"sort"

	"retroreverse.com/tools/lib/glb"
)

// maxAtlas is the conservative texture-size ceiling. Every level's textures
// fit well inside it; a level that did not would keep its own primitives.
const maxAtlas = 4096

// pad is the ring of wrapped edge texels around each packed texture. The
// textures were sampled with REPEAT, so a fetch that lands a hair past an edge
// used to come back around the other side; in an atlas it would land in the
// neighbour. Copying the opposite edge into the ring reproduces the wrap.
const pad = 1

// atlasGroups returns the merged groups and the rewritten UVs, or nil to
// decline (which leaves the caller's groups untouched).
func atlasGroups(groups []glb.TexturedGroup, uvs [][2]float32) ([]glb.TexturedGroup, [][2]float32) {
	if len(groups) < 2 {
		return nil, nil
	}
	for _, g := range groups {
		if g.Image == nil {
			return nil, nil
		}
		for _, t := range g.Tris {
			for _, vi := range t {
				u, v := uvs[vi][0], uvs[vi][1]
				if u < -1e-4 || u > 1+1e-4 || v < -1e-4 || v > 1+1e-4 {
					return nil, nil // a texture that repeats cannot be atlased
				}
			}
		}
	}

	// Shelf-pack, tallest first. A level's textures are a few dozen squares of
	// two sizes, so a shelf wastes almost nothing here.
	order := make([]int, len(groups))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		ra, rb := groups[order[a]].Image.Bounds(), groups[order[b]].Image.Bounds()
		if ra.Dy() != rb.Dy() {
			return ra.Dy() > rb.Dy()
		}
		return ra.Dx() > rb.Dx()
	})
	area, widest := 0, 1
	for _, i := range order {
		b := groups[i].Image.Bounds()
		area += (b.Dx() + 2*pad) * (b.Dy() + 2*pad)
		if w := b.Dx() + 2*pad; w > widest {
			widest = w
		}
	}
	W := pow2(widest)
	if s := pow2(isqrt(area)); s > W {
		W = s
	}
	var rects []image.Rectangle
	var H int
	for ; W <= maxAtlas; W *= 2 {
		rects, H = shelf(groups, order, W)
		if rects != nil && pow2(H) <= maxAtlas {
			H = pow2(H)
			break
		}
		rects = nil
	}
	if rects == nil {
		return nil, nil
	}

	atlas := image.NewRGBA(image.Rect(0, 0, W, H))
	for i, r := range rects {
		if r.Empty() {
			continue
		}
		drawWrapped(atlas, groups[i].Image, r)
	}

	// Rewrite each vertex's UV into its texture's rectangle. Positions are
	// emitted three-per-triangle and never shared, so a vertex belongs to
	// exactly one group — done[] holds that claim to account rather than
	// trusting it.
	out := make([][2]float32, len(uvs))
	copy(out, uvs)
	done := make([]bool, len(uvs))
	fw, fh := float32(W), float32(H)
	for i, g := range groups {
		r := rects[i]
		x0, y0 := float32(r.Min.X), float32(r.Min.Y)
		w, h := float32(r.Dx()), float32(r.Dy())
		for _, t := range g.Tris {
			for _, vi := range t {
				if done[vi] {
					continue
				}
				done[vi] = true
				out[vi][0] = (x0 + out[vi][0]*w) / fw
				out[vi][1] = (y0 + out[vi][1]*h) / fh
			}
		}
	}

	// One primitive per material class, all sharing the one atlas image (glb
	// dedupes textures by image pointer, so that is also one GPU texture).
	type class struct{ opaque, single, blend, additive bool }
	merged := map[class]*glb.TexturedGroup{}
	var order2 []class
	for _, g := range groups {
		k := class{g.Opaque, g.SingleSided, g.Blend, g.Additive}
		m, ok := merged[k]
		if !ok {
			m = &glb.TexturedGroup{
				Image: atlas, SingleSided: k.single, Opaque: k.opaque,
				Blend: k.blend, Additive: k.additive,
				// CLAMP, not REPEAT: the ring already answers what wrapping
				// used to, and clamping keeps a stray UV inside its own tile
				// instead of fetching a neighbour's picture.
				WrapS: 33071, WrapT: 33071,
			}
			merged[k] = m
			order2 = append(order2, k)
		}
		m.Tris = append(m.Tris, g.Tris...)
	}
	res := make([]glb.TexturedGroup, 0, len(order2))
	for _, k := range order2 {
		res = append(res, *merged[k])
	}
	return res, out
}

// shelf lays the textures out in rows of the given width, returning where each
// group's texture goes (indexed like groups) and the height used.
func shelf(groups []glb.TexturedGroup, order []int, W int) ([]image.Rectangle, int) {
	rects := make([]image.Rectangle, len(groups))
	x, y, rowH := 0, 0, 0
	for _, i := range order {
		b := groups[i].Image.Bounds()
		w, h := b.Dx()+2*pad, b.Dy()+2*pad
		if w > W {
			return nil, 0
		}
		if x+w > W {
			x, y, rowH = 0, y+rowH, 0
		}
		rects[i] = image.Rect(x+pad, y+pad, x+pad+b.Dx(), y+pad+b.Dy())
		x += w
		if h > rowH {
			rowH = h
		}
	}
	return rects, y + rowH
}

// drawWrapped blits src into r and fills the surrounding ring with the texels
// REPEAT would have fetched there — the same image offset by its own size.
func drawWrapped(dst *image.RGBA, src image.Image, r image.Rectangle) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	ring := image.Rect(r.Min.X-pad, r.Min.Y-pad, r.Max.X+pad, r.Max.Y+pad)
	for _, dy := range []int{-h, 0, h} {
		for _, dx := range []int{-w, 0, w} {
			at := image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Min.X+dx+w, r.Min.Y+dy+h).Intersect(ring)
			if at.Empty() {
				continue
			}
			off := b.Min.Add(image.Pt(at.Min.X-(r.Min.X+dx), at.Min.Y-(r.Min.Y+dy)))
			draw.Draw(dst, at, src, off, draw.Src)
		}
	}
}

func pow2(n int) int {
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

func isqrt(n int) int {
	r := 0
	for r*r < n {
		r++
	}
	return r
}
