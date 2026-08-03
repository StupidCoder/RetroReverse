package merc

// strips.go builds an effect's triangles from the DISC DATA alone.
//
// The reconstruction mirrors what the microprogram does with the same bytes
// (read off mercmicro.dis, verified slot-for-slot against emulated packets):
//
//   - Each vertex scatter-writes its {ST, RGBA, XYZF2} triple into one or
//     two output positions: the two dest bytes of lump q1 (x/y lanes) are
//     quadword offsets into the fragment's output packet. The GS consumes
//     the packet in address order, so ASCENDING DEST ORDER is the strip
//     order; gaps between dests hold mid-strip A+D (adgif) packets, not
//     vertices.
//   - The ADC bit is per WRITE, controlled by lump q0's y-lane byte through
//     the VIF row bias (16-bit truncation: c = int16(byte + row&0xFFFF),
//     row = the ctrl+44 magic whose low half is 0xFF80, so c = byte-0x80):
//     c > 0 both writes clean; c == 0 the D1 write is ADC, the D2 write
//     clean (the microcode rebuilds vf21 without the ADC add between the
//     two stores, 0x718-0x748); c < 0 both writes ADC.
//   - After the scatter, the fragment-end path (0x3C40-0x3E90) runs the
//     COPY TABLE in the byte header: hdr[0] = table offset in quadwords,
//     hdr[7] entries copying within this fragment's output, then hdr[8]
//     entries copying from the PREVIOUS fragment's output (the other
//     buffer half). Each 4-byte entry is {src dest, dst dest, ?, b3}; the
//     tail re-derives the copy's ADC via itof15.w + add.w of b3, so the
//     copy's ADC = source ADC XOR b3.
//   - The per-vertex records at hdr[1] are RGBA colors (alpha 0x80 = 1.0),
//     not topology; strips span fragments (mid-strip packets are tagless),
//     so the whole effect is one continuous vertex sequence whose ADC bits
//     gate the triangle kicks.

import (
	"sort"

	"retroreverse.com/tools/lib/glb"
)

// slotRef is one output position's content after a fragment's scatter+copies.
type slotRef struct {
	vert int // effect-wide vertex index
	adc  bool
}

// walkFragment computes a fragment's output map (dest byte -> content) from
// file data. prev is the previous fragment's map (for cross-fragment
// copies); vbase is the effect-wide index of the fragment's vertex 0.
func walkFragment(fr *Fragment, prev map[byte]slotRef, vbase int) map[byte]slotRef {
	vs := fr.Vertices()
	slots := make(map[byte]slotRef, len(vs)+8)
	for i, v := range vs {
		// c = int16(byte + rowMagic&0xFFFF); the logo's magic low half is
		// 0xFF80, i.e. c = byte - 0x80.
		c := int(v.Ctl) - 0x80
		slots[v.D1] = slotRef{vert: vbase + i, adc: c <= 0}
		if v.D2 != v.D1 {
			slots[v.D2] = slotRef{vert: vbase + i, adc: c < 0}
		} else if c == 0 {
			// same slot written twice: the second (clean) store wins
			slots[v.D1] = slotRef{vert: vbase + i, adc: false}
		}
	}
	tbl := int(fr.ByteData[0]) * 4
	nA := int(fr.ByteData[7])
	nB := int(fr.ByteData[8])
	for k := 0; k < nA+nB; k++ {
		o := tbl + k*4
		if o+4 > len(fr.ByteData) {
			break
		}
		src, dst, b3 := fr.ByteData[o], fr.ByteData[o+1], fr.ByteData[o+3]
		var from slotRef
		var ok bool
		if k < nA {
			from, ok = slots[src]
		} else {
			from, ok = prev[src]
		}
		if ok {
			slots[dst] = slotRef{vert: from.vert, adc: from.adc != (b3 != 0)}
		}
	}
	return slots
}

// EffectSequence flattens the effect into the GS-visible vertex stream:
// each fragment's occupied output positions in ascending address order.
func EffectSequence(e *Effect) []TopoVert {
	var seq []TopoVert
	var prev map[byte]slotRef
	vbase := 0
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		slots := walkFragment(fr, prev, vbase)
		dests := make([]int, 0, len(slots))
		for d := range slots {
			dests = append(dests, int(d))
		}
		sort.Ints(dests)
		for _, d := range dests {
			r := slots[byte(d)]
			seq = append(seq, TopoVert{Index: r.vert, ADC: r.adc})
		}
		prev = slots
		vbase += fr.LumpQWC / 3
	}
	return seq
}

// StripPrim builds one effect's primitive from file data only.
func StripPrim(e *Effect, base [4]float32) glb.Prim {
	var p glb.Prim
	p.BaseColor = base
	p.DoubleSided = true
	for fi := range e.Fragments {
		for _, v := range e.Fragments[fi].Vertices() {
			p.Positions = append(p.Positions, [3]float32{v.X, v.Y, v.Z})
			p.Normals = append(p.Normals, normalize(float32(v.NX), float32(v.NY), float32(v.NZ)))
		}
	}
	var w [3]int
	n := 0
	for _, tv := range EffectSequence(e) {
		w[0], w[1], w[2] = w[1], w[2], tv.Index
		if n < 2 {
			n++
			continue
		}
		if !tv.ADC && w[0] != w[1] && w[1] != w[2] && w[0] != w[2] {
			p.Tris = append(p.Tris, [3]uint32{uint32(w[0]), uint32(w[1]), uint32(w[2])})
		}
	}
	return p
}
