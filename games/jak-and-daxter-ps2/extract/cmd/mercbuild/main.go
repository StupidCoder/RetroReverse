package main

// mercbuild replays the live frame, harvests merc packets whose vertices
// carry encoded identities, and writes the model as a GLB: positions from
// the verified direct lattice decode, topology (strip order + ADC) from the
// microprogram's own kicked packets.
import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/lib/glb"
)

func main() {
	ramF := flag.String("ram", "", "RAM image")
	objF := flag.String("obj", "", "logo art group linked at base 0")
	microF := flag.String("micro", "", "merc microcode")
	out := flag.String("out", "", "output glb")
	flag.Parse()

	ram, err := os.ReadFile(*ramF)
	check(err)
	obj, err := os.ReadFile(*objF)
	check(err)
	micro, err := os.ReadFile(*microF)
	check(err)

	needle := []byte("/src/next/data/art-group6/logo-ag.go")
	base := uint32(bytes.Index(ram, needle) - bytes.Index(obj, needle))
	fmt.Printf("art group at 0x%X\n", base)

	var pos, nrm [][3]float32
	gbase := 0
	for _, off := range []uint32{0x1244, 0x29CE4} {
		c, err := merc.Parse(obj, off)
		check(err)
		for e := range c.Effects {
			for f := range c.Effects[e].Fragments {
				fr := &c.Effects[e].Fragments[f]
				for _, v := range fr.Vertices() {
					pos = append(pos, [3]float32{v.X, v.Y, v.Z})
					nrm = append(nrm, norm(v.NX, v.NY, v.NZ))
				}
				nv := fr.LumpQWC / 3
				// The art group is resident twice (the second instance is
				// what the live chain references); patch every copy.
				for _, b := range []uint32{base, base + 0x78AB30} {
					cb := b + fr.Off + (uint32(fr.ByteData[12])+1)*4
					for v := 0; v < nv+2; v++ {
						o := cb + uint32(v)*4
						if int(o)+4 > len(ram) {
							break
						}
						idx := gbase + v + 1
						ram[o], ram[o+1], ram[o+2], ram[o+3] = byte(idx), byte(idx>>8), 0, 0x80
					}
				}
				gbase += nv
			}
		}
	}
	fmt.Printf("%d vertices\n", len(pos))

	var heads []uint32
	for a := uint32(0); a+16 < uint32(len(ram)); a += 16 {
		if binary.LittleEndian.Uint32(ram[a:]) == 0x1000000A &&
			binary.LittleEndian.Uint32(ram[a+8:]) == 0x01000404 {
			heads = append(heads, a)
		}
	}

	// Accumulate every program the frame uploads: one head's chain only
	// carries part of the code, so a dry pass over all of them builds the
	// program memory the real VU would hold.
	pre := merc.NewReplayer(ram)
	pre.SkipRun = true
	for _, h := range heads {
		pre.Play(h)
	}

	hits := 0
	var allData []uint32
	seen := map[string]bool{}
	var tris [][3]uint32
	kept := 0
	for _, h := range heads {
		r := merc.NewReplayer(ram)
		r.RefLo, r.RefHi = 0x13A1210, 0x13A1210+0x2B000
		copy(r.V.Micro, pre.V.Micro)
		_ = micro
		r.Play(h)
		hits += r.RefHits
		allData = append(allData, r.Refs...)
		for _, pk := range r.Packets {
			idx := make([]int, 0, len(pk))
			adc := make([]bool, 0, len(pk))
			res := 0
			for _, v := range pk {
				i := -1
				if v.RGBA[3] == 0x80 && v.RGBA[2] == 0 && v.RGBA[1] < 0x40 {
					if enc := int(v.RGBA[0]) | int(v.RGBA[1])<<8; enc >= 1 && enc <= len(pos) {
						i = enc - 1
						res++
					}
				}
				idx = append(idx, i)
				adc = append(adc, v.ADC)
			}
			if res < 8 {
				continue
			}
			key := fmt.Sprint(idx)
			if seen[key] {
				continue
			}
			seen[key] = true
			kept++
			var w [3]int
			n := 0
			for i := range idx {
				if idx[i] < 0 {
					n = 0
					continue
				}
				w[0], w[1], w[2] = w[1], w[2], idx[i]
				if n < 2 {
					n++
					continue
				}
				if adc[i] {
					continue
				}
				a, b, c := uint32(w[0]), uint32(w[1]), uint32(w[2])
				if a == b || b == c || a == c {
					continue
				}
				if i&1 == 0 {
					tris = append(tris, [3]uint32{a, b, c})
				} else {
					tris = append(tris, [3]uint32{a, c, b})
				}
			}
		}
	}
	fmt.Printf("%d packets kept, %d triangles, artgroup-sourced tags %d\n", kept, len(tris), hits)
	buckets := map[uint32]int{}
	for _, a := range allData {
		buckets[a>>16]++
	}
	fmt.Println("big-transfer sources by 64K page:")
	for k, v := range buckets {
		if v > 20 {
			fmt.Printf("  0x%X0000: %d\n", k, v)
		}
	}
	if len(tris) == 0 {
		return
	}
	scene := glb.NewScene()
	node := scene.AddNode("logo", -1, [3]float32{}, [4]float32{0, 0, 0, 1}, [3]float32{1, 1, 1})
	check(scene.AddMesh(node, "logo", []glb.Prim{{
		Positions: pos, Normals: nrm, Tris: tris,
		BaseColor: [4]float32{0.83, 0.68, 0.28, 1}, DoubleSided: true,
	}}))
	check(scene.Write(*out, "title-logo"))
	fmt.Println("wrote", *out)
}

func norm(x, y, z uint8) [3]float32 {
	fx, fy, fz := float32(x)-128, float32(y)-128, float32(z)-128
	l := fx*fx + fy*fy + fz*fz
	if l == 0 {
		return [3]float32{0, 1, 0}
	}
	v := l
	for i := 0; i < 16; i++ {
		v = 0.5 * (v + l/v)
	}
	return [3]float32{fx / v, fy / v, fz / v}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercbuild:", err)
		os.Exit(1)
	}
}
