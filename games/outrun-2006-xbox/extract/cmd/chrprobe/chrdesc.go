package main

import (
	"fmt"
	"os"
)

// chrDump reads /Chr/CHR_*.bin: header {skeletonIdx u16, nAttach u16,
// nLOD u16, nMat u16} + u32 offsets {attachOff, lodOff, 0x38, 0, matOff,
// extraOff}. Attach records 8 B {boneIdx u16, flags u8, pad, partId u16,
// w u16}; LOD records 20 B; matrices 0x40 B row-vector.
func chrDump(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	skel, nAttach, nLOD, nMat := u16(b, 0), int(u16(b, 2)), int(u16(b, 4)), int(u16(b, 6))
	attOff, lodOff := int(u32(b, 8)), int(u32(b, 12))
	matOff, extraOff := int(u32(b, 24)), int(u32(b, 28))
	fmt.Printf("%s: %d bytes skel=%d nAttach=%d nLOD=%d nMat=%d att=%#x lod=%#x mat=%#x extra=%#x hdr16=%#x hdr20=%#x\n",
		path, len(b), skel, nAttach, nLOD, nMat, attOff, lodOff, matOff, extraOff, u32(b, 16), u32(b, 20))
	fmt.Printf("attach records @%#x (%d):\n", attOff, nAttach)
	for i := 0; i < nAttach; i++ {
		o := attOff + i*8
		fmt.Printf("  %2d: bone=%5d fl=%#02x b3=%#02x part=%5d w=%d\n",
			i, u16(b, o), b[o+2], b[o+3], int16(u16(b, o+4)), u16(b, o+6))
	}
	fmt.Printf("LOD records @%#x (%d):\n", lodOff, nLOD)
	for i := 0; i < nLOD; i++ {
		o := lodOff + i*20
		fmt.Printf("  %2d:", i)
		for k := 0; k < 10; k++ {
			fmt.Printf(" %6d", int16(u16(b, o+2*k)))
		}
		fmt.Println()
	}
	fmt.Printf("matrices @%#x (%d):\n", matOff, nMat)
	for i := 0; i < nMat; i++ {
		o := matOff + i*0x40
		var r [3][3]float32
		var t [3]float32
		for row := 0; row < 3; row++ {
			for col := 0; col < 3; col++ {
				r[row][col] = f32(b, o+16*row+4*col)
			}
		}
		for col := 0; col < 3; col++ {
			t[col] = f32(b, o+48+4*col)
		}
		// row-vector convention: p' = p·M + t. Bind origin solves p·M + t = 0
		// → p = -t·Mᵀ (orthonormal M).
		var org [3]float32
		for col := 0; col < 3; col++ {
			org[col] = -(t[0]*r[col][0] + t[1]*r[col][1] + t[2]*r[col][2])
		}
		fmt.Printf("  %2d: [%7.3f %7.3f %7.3f | %7.3f %7.3f %7.3f | %7.3f %7.3f %7.3f | t %7.3f %7.3f %7.3f]  origin (%7.3f %7.3f %7.3f)\n",
			i,
			r[0][0], r[0][1], r[0][2],
			r[1][0], r[1][1], r[1][2],
			r[2][0], r[2][1], r[2][2],
			t[0], t[1], t[2], org[0], org[1], org[2])
	}
	if extraOff > 0 && extraOff < len(b) {
		fmt.Printf("extra @%#x:", extraOff)
		for o := extraOff; o < len(b) && o < extraOff+0x20; o += 4 {
			fmt.Printf(" %#x", u32(b, o))
		}
		fmt.Println()
	}
}
