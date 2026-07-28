// attachprobe tests the demo attachment hypothesis: is a prop's world node 0
// a constant offset off some Luigi joint? For each carrier node i it computes
// C_i = inv(world_luigi[i]) · world_prop[0] in two states and reports the
// nodes whose C agrees — the constant C and its node name the attachment.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"

	"retroreverse.com/tools/platform/gc"
)

type mtx [12]float32

func read(ram []byte, addr uint32, i int) mtx {
	var m mtx
	off := int(addr-0x80000000) + i*48
	for k := range m {
		m[k] = math.Float32frombits(binary.BigEndian.Uint32(ram[off+k*4:]))
	}
	return m
}

func inv(m mtx) mtx {
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[4], m[5], m[6]
	g, h, i := m[8], m[9], m[10]
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	id := 1 / det
	var r mtx
	r[0], r[1], r[2] = (e*i-f*h)*id, (c*h-b*i)*id, (b*f-c*e)*id
	r[4], r[5], r[6] = (f*g-d*i)*id, (a*i-c*g)*id, (c*d-a*f)*id
	r[8], r[9], r[10] = (d*h-e*g)*id, (b*g-a*h)*id, (a*e-b*d)*id
	tx, ty, tz := m[3], m[7], m[11]
	r[3] = -(r[0]*tx + r[1]*ty + r[2]*tz)
	r[7] = -(r[4]*tx + r[5]*ty + r[6]*tz)
	r[11] = -(r[8]*tx + r[9]*ty + r[10]*tz)
	return r
}

func mul(a, b mtx) mtx {
	var o mtx
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			o[r*4+c] = a[r*4]*b[c] + a[r*4+1]*b[4+c] + a[r*4+2]*b[8+c]
		}
		o[r*4+3] = a[r*4]*b[3] + a[r*4+1]*b[7] + a[r*4+2]*b[11] + a[r*4+3]
	}
	return o
}

func diff(a, b mtx) float32 {
	var w float32
	for i := range a {
		d := a[i] - b[i]
		if d < 0 {
			d = -d
		}
		if i%4 != 3 {
			d *= 1000
		}
		if d > w {
			w = d
		}
	}
	return w
}

func main() {
	image := flag.String("image", "", "")
	s1 := flag.String("s1", "", "state 1")
	s2 := flag.String("s2", "", "state 2")
	carrier := flag.Uint64("carrier", 0, "carrier world array addr")
	n := flag.Int("n", 122, "carrier node count")
	prop := flag.Uint64("prop", 0, "prop world array addr")
	flag.Parse()
	load := func(p string) []byte {
		disc, err := gc.Open(*image)
		if err != nil {
			panic(err)
		}
		m, err := gc.NewMachine(disc)
		if err != nil {
			panic(err)
		}
		if err := m.LoadStateFile(p); err != nil {
			panic(err)
		}
		return m.RAM
	}
	r1, r2 := load(*s1), load(*s2)
	p1, p2 := read(r1, uint32(*prop), 0), read(r2, uint32(*prop), 0)
	for i := 0; i < *n; i++ {
		c1 := mul(inv(read(r1, uint32(*carrier), i)), p1)
		c2 := mul(inv(read(r2, uint32(*carrier), i)), p2)
		if d := diff(c1, c2); d < 100 {
			fmt.Printf("node %3d agrees, diff %8.3f  C=[% .3f % .3f % .3f % .1f | % .3f % .3f % .3f % .1f | % .3f % .3f % .3f % .1f]\n",
				i, d, c1[0], c1[1], c1[2], c1[3], c1[4], c1[5], c1[6], c1[7], c1[8], c1[9], c1[10], c1[11])
		}
	}
}
