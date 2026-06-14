// regions renders the per-course SLOPE FIELD — the static terrain the marble
// rolls on — from each course's Track file, as isometric PNGs.
//
// The slope field is the course descriptor at Track header +0 (global $9A6),
// reverse-engineered in Marble_Madness.md Part V §4. It is an array of 8-byte
// region records (descriptor +$1A = count, +$1C = table), each:
//
//	[0] x0 (s8)  [1] y0   [2] xSize  [3] ySize   (axis-aligned rect in ISO TILES)
//	[4..5] baseHeight (word)
//	[6] low5 = edge-shape selector -> a height-delta profile (table at +$20)
//	[7] low3 = slope direction 0..7 (the 4 iso diagonals x 2 fill orders)
//	[7] bit3 = flip (negate the profile)
//
// The engine routine build_region ($E158) rasterises these into a corner-height
// mesh; this tool replays that height generation — value = baseHeight ± profile,
// the profile consumed one byte per cell in $E158's exact diagonal fill order
// (reset on a $80 marker) — and plots the result in iso tile space.
//
// NOTE this is the STATIC slope field (the checkerboard). The handful of dynamic
// regions (seesaws / holes / triggers) live in a separate scripted list at Track
// header +$14 and are not drawn here.
//
// Usage: regions <disk.adf> <outdir>
//
//	writes <outdir>/<course>.regions.png (tiles coloured by slope direction)
//	   and <outdir>/<course>.height.png  (relief-shaded height field)
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"stupidcoder.com/tools/amiga/adf"
	"stupidcoder.com/tools/amiga/hunk"
)

var courses = []struct{ key, track string }{
	{"practy", "PrcTrack"}, {"beginr", "BegTrack"}, {"interm", "IntTrack"},
	{"aerial", "AerTrack"}, {"silly", "SilTrack"}, {"ultima", "UltTrack"},
}

// $2504 direction table: the 4 iso diagonals, indexed by the record's 3-bit dir.
var dirVec = [8][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, {1, 1}, {-1, 1}, {-1, -1}, {1, -1}}

func u16(b []byte, o uint32) int {
	if int(o)+2 > len(b) {
		return 0
	}
	return int(b[o])<<8 | int(b[o+1])
}
func u32(b []byte, o uint32) uint32 {
	if int(o)+4 > len(b) {
		return 0
	}
	return uint32(b[o])<<24 | uint32(b[o+1])<<16 | uint32(b[o+2])<<8 | uint32(b[o+3])
}
func s8(v byte) int {
	if v > 127 {
		return int(v) - 256
	}
	return int(v)
}

type record struct{ x0, y0, xs, ys, bh, edge, dir, flip int }

// profileSeq expands an edge-shape profile to n signed deltas, resetting to the
// start on a $80 marker (build_region $E158 logic).
func profileSeq(im []byte, etbl uint32, edge, n int) []int {
	ep := u32(im, etbl+uint32(edge)*4)
	seq := make([]int, 0, n)
	for i := uint32(0); len(seq) < n; {
		if int(ep+i) >= len(im) {
			break
		}
		v := im[ep+i]
		if v == 0x80 {
			i = 0
			continue
		}
		seq = append(seq, s8(v))
		i++
		if i > 255 {
			break
		}
	}
	return seq
}

type cell struct {
	h, dir int
	sloped bool
}

// buildField replays build_region for every record into a per-tile height map.
func buildField(im []byte) (map[[2]int]cell, int, int) {
	d := u32(im, 0) // header +0 -> $9A6 descriptor
	cnt := u16(im, d+0x1A)
	tbl := u32(im, d+0x1C)
	etbl := u32(im, d+0x20)
	field := map[[2]int]cell{}
	loSlope, hiSlope := 1<<30, -(1 << 30)
	for k := 0; k < cnt; k++ {
		o := tbl + uint32(k)*8
		if int(o)+8 > len(im) {
			break
		}
		r := record{s8(im[o]), int(im[o+1]), int(im[o+2]), int(im[o+3]),
			u16(im, o+4), int(im[o+6]) & 0x1F, int(im[o+7]) & 7, (int(im[o+7]) >> 3) & 1}
		dx, dy := dirVec[r.dir][0], dirVec[r.dir][1]
		xEnd, yEnd := r.x0+r.xs-1, r.y0+r.ys-1
		seq := profileSeq(im, etbl, r.edge, r.xs*r.ys+4)
		sloped := false
		for i := 0; i < r.xs*r.ys && i < len(seq); i++ {
			if seq[i] != 0 {
				sloped = true
				break
			}
		}
		fi := 0
		put := func(tx, ty int) {
			delta := seq[len(seq)-1]
			if fi < len(seq) {
				delta = seq[fi]
			}
			fi++
			h := r.bh + delta
			if r.flip == 1 {
				h = r.bh - delta
			}
			field[[2]int{tx, ty}] = cell{h, r.dir, sloped}
			if h > 8000 {
				if h < loSlope {
					loSlope = h
				}
				if h > hiSlope {
					hiSlope = h
				}
			}
		}
		// dir<4: outer y, inner x ; dir>=4: outer x, inner y (the two $E158 loops)
		xs := seqRange(r.x0, xEnd, dx)
		ys := seqRange(r.y0, yEnd, dy)
		if r.dir < 4 {
			for _, ty := range ys {
				for _, tx := range xs {
					put(tx, ty)
				}
			}
		} else {
			for _, tx := range xs {
				for _, ty := range ys {
					put(tx, ty)
				}
			}
		}
	}
	return field, loSlope, hiSlope
}

func seqRange(a, b, step int) []int {
	var out []int
	if step > 0 {
		for i := a; i <= b; i++ {
			out = append(out, i)
		}
	} else {
		for i := b; i >= a; i-- {
			out = append(out, i)
		}
	}
	return out
}

const tileS = 12

func iso(tx, ty int) (int, int) { return (tx - ty) * tileS, (tx + ty) * tileS / 2 }

// fillDiamond draws a tile diamond centred at (cx,cy), half-width tileS.
func fillDiamond(img *image.RGBA, cx, cy int, c color.RGBA) {
	for dy := -tileS / 2; dy <= tileS/2; dy++ {
		w := tileS - 2*abs(dy)
		for dx := -w; dx <= w; dx++ {
			img.SetRGBA(cx+dx, cy+dy, c)
		}
	}
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

var dirCol = map[int]color.RGBA{
	0: {70, 120, 220, 255}, 4: {220, 90, 70, 255}, 2: {90, 200, 90, 255},
	1: {200, 200, 80, 255}, 3: {180, 90, 200, 255}, 5: {80, 200, 200, 255},
	6: {200, 140, 70, 255}, 7: {150, 150, 150, 255},
}

// heightRamp maps t in [0,1] -> blue(low)..white(high).
func heightRamp(t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	stops := []struct {
		p          float64
		r, g, b    int
	}{{0, 30, 40, 120}, {0.3, 40, 140, 150}, {0.55, 60, 170, 80}, {0.78, 220, 205, 70}, {1, 250, 250, 250}}
	for i := 0; i < len(stops)-1; i++ {
		if t <= stops[i+1].p {
			f := (t - stops[i].p) / (stops[i+1].p - stops[i].p + 1e-9)
			return color.RGBA{
				clamp8(stops[i].r + int(float64(stops[i+1].r-stops[i].r)*f)),
				clamp8(stops[i].g + int(float64(stops[i+1].g-stops[i].g)*f)),
				clamp8(stops[i].b + int(float64(stops[i+1].b-stops[i].b)*f)), 255}
		}
	}
	return color.RGBA{250, 250, 250, 255}
}

func render(field map[[2]int]cell, lo, hi int, height bool) *image.RGBA {
	minX, minY, maxX, maxY := 1<<30, 1<<30, -(1 << 30), -(1 << 30)
	for t := range field {
		x, y := iso(t[0], t[1])
		minX, maxX = min(minX, x), max(maxX, x)
		minY, maxY = min(minY, y), max(maxY, y)
	}
	const M = 44
	W := (maxX - minX) + 2*M + 2*tileS
	H := (maxY - minY) + 2*M + 2*tileS
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	for i := range img.Pix {
		if i%4 == 3 {
			img.Pix[i] = 255
		} else {
			img.Pix[i] = 12
		}
	}
	hget := func(tx, ty int) int {
		if c, ok := field[[2]int{tx, ty}]; ok {
			return c.h
		}
		return -1
	}
	for t, c := range field {
		x, y := iso(t[0], t[1])
		cx, cy := x-minX+M+tileS, y-minY+M+tileS
		var col color.RGBA
		if height {
			if c.h < 8000 {
				col = color.RGBA{8, 8, 40, 255} // pit
			} else {
				col = heightRamp(float64(c.h-lo) / (float64(hi-lo) + 1e-9))
				// relief shading from local gradient
				h := c.h
				gx := neigh(hget, t[0]+1, t[1], h) - neigh(hget, t[0]-1, t[1], h)
				gy := neigh(hget, t[0], t[1]+1, h) - neigh(hget, t[0], t[1]-1, h)
				f := float64(-gx-gy) / 8.0
				if f > 1 {
					f = 1
				}
				if f < -1 {
					f = -1
				}
				k := 1.0 + 0.5*f
				col = color.RGBA{clamp8(int(float64(col.R) * k)), clamp8(int(float64(col.G) * k)), clamp8(int(float64(col.B) * k)), 255}
			}
		} else {
			col = dirCol[c.dir]
			if !c.sloped { // flat region -> dimmed
				col = color.RGBA{col.R/3 + 30, col.G/3 + 30, col.B/3 + 30, 255}
			}
		}
		fillDiamond(img, cx, cy, col)
	}
	return img
}

func neigh(hget func(int, int) int, tx, ty, def int) int {
	if h := hget(tx, ty); h >= 0 {
		return h
	}
	return def
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: regions <disk.adf> <outdir>")
		os.Exit(2)
	}
	img, err := os.ReadFile(os.Args[1])
	chk(err)
	vol, err := adf.Open(img)
	chk(err)
	outdir := os.Args[2]
	chk(os.MkdirAll(outdir, 0o755))

	paths := map[string]string{}
	chk(vol.Walk(func(e adf.Entry) error {
		if !e.IsDir {
			paths[strings.ToLower(e.Name)] = e.Path
		}
		return nil
	}))

	for _, c := range courses {
		p, ok := paths[strings.ToLower(c.track)]
		if !ok {
			fmt.Printf("%s: %s not found\n", c.key, c.track)
			continue
		}
		data, err := vol.ReadFile(p)
		chk(err)
		prog, err := hunk.Load(data, 0)
		chk(err)
		field, lo, hi := buildField(prog.Image)
		writePNG(filepath.Join(outdir, c.key+".regions.png"), render(field, lo, hi, false))
		writePNG(filepath.Join(outdir, c.key+".height.png"), render(field, lo, hi, true))
		fmt.Printf("%s: %d tiles, height %d..%d -> %s.{regions,height}.png\n", c.key, len(field), lo, hi, c.key)
	}
}

func writePNG(path string, img *image.RGBA) {
	f, err := os.Create(path)
	chk(err)
	defer f.Close()
	chk(png.Encode(f, img))
}

func chk(e error) {
	if e != nil {
		panic(e)
	}
}
