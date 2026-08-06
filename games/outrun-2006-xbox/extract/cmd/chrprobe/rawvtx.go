package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// rawVtxDump prints the first vertices of every pair of a part of an
// inflated pmt payload (local file), raw-decoding stride-28 fmt 0x116.
// partSummary prints every part's pair formats and position bbox.
func partSummary(path string) {
	data := mustPMT(path)
	nParts := int(u32(data, 0))
	szA := int(u32(data, 8))
	a, b := data[0x10:0x10+szA], data[0x10+szA:]
	for pi := 0; pi < nParts; pi++ {
		rec := 0x18 + pi*0x3C
		w := make([]uint32, 15)
		for i := range w {
			w[i] = u32(a, rec+4*i)
		}
		ent := int(w[5])
		nPairs := int(u32(a, ent+4*9))
		var mn, mx [3]float32
		for c := 0; c < 3; c++ {
			mn[c], mx[c] = 1e30, -1e30
		}
		total := 0
		fmts := ""
		for k := 0; k < nPairs; k++ {
			t := int(w[12]) + k*0x2C
			vbBytes := u32(a, t+0x1C)
			fmtWord := u32(a, t+0x20)
			stride := u32(a, t+0x24)
			vbDesc := u32(a, int(w[2])+4*(4*k))
			vbOff := u32(a, int(vbDesc)+4)
			nv := int(vbBytes / stride)
			total += nv
			fmts += fmt.Sprintf(" %#x/%d", fmtWord, stride)
			for i := 0; i < nv; i++ {
				o := int(vbOff) + i*int(stride)
				for c := 0; c < 3; c++ {
					v := f32(b, o+4*c)
					if v < mn[c] {
						mn[c] = v
					}
					if v > mx[c] {
						mx[c] = v
					}
				}
			}
		}
		fmt.Printf("part %2d: %4d verts fmts%s bbox (%7.3f %7.3f %7.3f)..(%7.3f %7.3f %7.3f)\n",
			pi, total, fmts, mn[0], mn[1], mn[2], mx[0], mx[1], mx[2])
	}
}

func mustPMT(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	if data[0] == 0x78 {
		zr, _ := zlib.NewReader(bytes.NewReader(data))
		if d2, err := io.ReadAll(zr); err == nil {
			data = d2
		}
	}
	return data
}

func rawVtxDump(path string, pi int, n int) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	if data[0] == 0x78 {
		zr, _ := zlib.NewReader(bytes.NewReader(data))
		if d2, err := io.ReadAll(zr); err == nil {
			data = d2
		}
	}
	nParts := int(u32(data, 0))
	szA := int(u32(data, 8))
	a, b := data[0x10:0x10+szA], data[0x10+szA:]
	if pi >= nParts {
		fatal("part %d out of %d", pi, nParts)
	}
	rec := 0x18 + pi*0x3C
	w := make([]uint32, 15)
	for i := range w {
		w[i] = u32(a, rec+4*i)
	}
	ent := int(w[5])
	nPairs := int(u32(a, ent+4*9))
	fmt.Printf("part %d: %d pairs (w4=%#x w6=%#x w7=%#x)\n", pi, nPairs, w[4], w[6], w[7])
	for k := 0; k < nPairs; k++ {
		t := int(w[12]) + k*0x2C
		vbBytes := u32(a, t+0x1C)
		fmtWord := u32(a, t+0x20)
		stride := u32(a, t+0x24)
		vbDesc := u32(a, int(w[2])+4*(4*k))
		vbOff := u32(a, int(vbDesc)+4)
		nv := int(vbBytes / stride)
		fmt.Printf(" pair %d: fmt=%#x stride=%d verts=%d vbOff=%#x\n", k, fmtWord, stride, nv, vbOff)
		for i := 0; i < nv && i < n; i++ {
			o := int(vbOff) + i*int(stride)
			fmt.Printf("  v%3d: pos=(%8.4f %8.4f %8.4f)", i, f32(b, o), f32(b, o+4), f32(b, o+8))
			fmt.Printf(" rest=")
			for j := 12; j < int(stride); j += 4 {
				fmt.Printf(" %08x", u32(b, o+j))
			}
			// candidate float reads of the trailing dwords
			for j := 16; j < int(stride)-8; j += 4 {
				fmt.Printf(" f%d=%g", j, f32(b, o+j))
			}
			if stride >= 28 {
				fmt.Printf(" uv=(%.3f %.3f)", f32(b, o+int(stride)-8), f32(b, o+int(stride)-4))
			}
			fmt.Println()
		}
	}
}
