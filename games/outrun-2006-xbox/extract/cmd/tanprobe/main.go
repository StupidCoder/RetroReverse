// tanprobe is the data-side half of naming the stride-44 layout's trailing
// f32x3 @32 (the tex2 attribute slot). The program side: bootoracle -vshdump 11
// on the beach race dumps the transform program bound at the stride-44 draws,
// which crosses v11 with the normal twice into a 3-vector basis on oT1-3 with
// the per-vertex eye ray in the w components — the NV2A texm3x3 setup for the
// per-pixel cube-map reflection (the draws' material binds a cube on stage 3).
// That reading predicts v11 lies in the surface plane; this probe checks it:
// over the beach course's one stride-44 pair (part 5 pair 2 — the sea, a flat
// y=-51.2 plane spanning the bay), dot(normal, f32x3) ~ 0 at every vertex.
// Result: mean |dot| 0.0074, max 0.055 (the 11:11:10 normal's quantisation),
// lengths 1..18 clustered small — an unnormalised UV-derivative-style tangent.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"retroreverse.com/tools/platform/xbox"
)

func main() {
	disc, err := xbox.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	raw, err := disc.ReadFile("/Stage/BEAC/cs_CS_BEAC_pmt.sz")
	if err != nil {
		panic(err)
	}
	zr, _ := zlib.NewReader(bytes.NewReader(raw))
	data, _ := io.ReadAll(zr)
	szA := int(binary.LittleEndian.Uint32(data[8:]))
	b := data[0x10+szA:]
	// Part 5 pair 2, the file's one stride-44 pair (carex -inspect offsets).
	vbOff, vbBytes, stride := 0x5ec230, 0x955c, 44
	n := vbBytes / stride
	f32 := func(o int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(b[o:])) }
	unpack := func(v uint32) [3]float64 {
		sext := func(u uint32, bits int) float64 {
			s := int32(u<<(32-bits)) >> (32 - bits)
			return float64(s) / float64(int32(1)<<(bits-1)-1)
		}
		return [3]float64{sext(v&0x7FF, 11), sext(v>>11&0x7FF, 11), sext(v>>22&0x3FF, 10)}
	}
	var dots, lens []float64
	for i := 0; i < n; i++ {
		v := vbOff + i*stride
		nr := unpack(binary.LittleEndian.Uint32(b[v+12:]))
		tx, ty, tz := float64(f32(v+32)), float64(f32(v+36)), float64(f32(v+40))
		tl := math.Sqrt(tx*tx + ty*ty + tz*tz)
		nl := math.Sqrt(nr[0]*nr[0] + nr[1]*nr[1] + nr[2]*nr[2])
		lens = append(lens, tl)
		if tl > 0 && nl > 0 {
			dots = append(dots, (tx*nr[0]+ty*nr[1]+tz*nr[2])/(tl*nl))
		}
	}
	maxAbs, sum := 0.0, 0.0
	for _, d := range dots {
		if math.Abs(d) > maxAbs {
			maxAbs = math.Abs(d)
		}
		sum += math.Abs(d)
	}
	hist := map[int]int{}
	lmin, lmax := lens[0], lens[0]
	for _, l := range lens {
		hist[int(l+0.5)]++
		if l < lmin {
			lmin = l
		}
		if l > lmax {
			lmax = l
		}
	}
	fmt.Printf("len hist: %v\n", hist)
	fmt.Printf("%d verts: |dot(n,t)| mean %.4f max %.4f; |t| range %.4f..%.4f\n",
		n, sum/float64(len(dots)), maxAbs, lmin, lmax)
}
