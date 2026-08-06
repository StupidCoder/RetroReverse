package main

import (
	"fmt"
	"math"
	"os"
)

func f32(b []byte, o int) float32 { return math.Float32frombits(u32(b, o)) }

func cstr(b []byte, o int) string {
	e := o
	for e < len(b) && b[e] != 0 {
		e++
	}
	return string(b[o:e])
}

type boneRec struct {
	name     string
	parent   int
	a, bcnt  uint16 // +4: u16,u16 (parent-ish pair, unresolved)
	f        [8]float32
	children []int
}

type skeleton struct {
	name  string
	bones []boneRec
}

// parseBoneLib reads /Common/bone.bin: 16-byte directory {1, boneCount u16,
// FFFF, 0, recordsOff u32, nameOff u32}; bone record = 0x38 bytes.
func parseBoneLib(b []byte) []skeleton {
	var sks []skeleton
	for o := 0; o+16 <= len(b); o += 16 {
		if u16(b, o) != 1 {
			break
		}
		cnt := int(u16(b, o+2))
		recOff := int(u32(b, o+8))
		nameOff := int(u32(b, o+12))
		sk := skeleton{name: cstr(b, nameOff)}
		for i := 0; i < cnt; i++ {
			r := recOff + i*0x38
			br := boneRec{
				name: cstr(b, int(u32(b, r))),
				a:    u16(b, r+4), bcnt: u16(b, r+6),
			}
			for j := 0; j < 8; j++ {
				br.f[j] = f32(b, r+0x10+4*j)
			}
			nChild := int(u32(b, r+0x30))
			childOff := int(u32(b, r+0x34))
			for c := 0; c < nChild; c++ {
				br.children = append(br.children, int(u16(b, childOff+2*c)))
			}
			sk.bones = append(sk.bones, br)
		}
		// derive parents from child lists
		for i := range sk.bones {
			sk.bones[i].parent = -1
		}
		for i, br := range sk.bones {
			for _, c := range br.children {
				if c < len(sk.bones) {
					sk.bones[c].parent = i
				}
			}
		}
		sks = append(sks, sk)
	}
	return sks
}

func boneDump(path, which string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	sks := parseBoneLib(b)
	fmt.Printf("%s: %d skeletons\n", path, len(sks))
	for _, sk := range sks {
		if which != "" && sk.name != which {
			fmt.Printf("  %s: %d bones\n", sk.name, len(sk.bones))
			continue
		}
		fmt.Printf("  %s: %d bones\n", sk.name, len(sk.bones))
		for i, br := range sk.bones {
			fmt.Printf("   %2d %-16s par=%2d a=%#x b=%#x f=[%7.3f %7.3f %7.3f | %7.3f %7.3f %7.3f | %5.2f %5.2f] ch=%v\n",
				i, br.name, br.parent, br.a, br.bcnt,
				br.f[0], br.f[1], br.f[2], br.f[3], br.f[4], br.f[5], br.f[6], br.f[7], br.children)
		}
	}
}
