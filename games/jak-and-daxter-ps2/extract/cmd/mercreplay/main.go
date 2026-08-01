package main

// mercreplay plays the game's own merc DMA chains from a RAM image with the
// logo art group's color records index-encoded, and reports the harvested
// packet topology.
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	ram, err := os.ReadFile(os.Args[1])
	check(err)
	obj, err := os.ReadFile(os.Args[2]) // logo linked at base 0
	check(err)
	mercMicro, err := os.ReadFile(os.Args[3])
	check(err)

	// locate the art group in RAM by its file-info source path
	needle := []byte("/src/next/data/art-group6/logo-ag.go")
	pObj := bytes.Index(obj, needle)
	pRAM := bytes.Index(ram, needle)
	if pObj < 0 || pRAM < 0 {
		fmt.Println("logo-ag path not found", pObj, pRAM)
		os.Exit(1)
	}
	base := uint32(pRAM - pObj)
	fmt.Printf("logo art group at RAM 0x%X\n", base)

	// index-encode all fragments' color records, globally numbered
	gbase := 0
	type frag struct {
		g  int
		fr *merc.Fragment
	}
	var frags []frag
	for _, off := range []uint32{0x1244, 0x29CE4} {
		c, err := merc.Parse(obj, off)
		check(err)
		for e := range c.Effects {
			for f := range c.Effects[e].Fragments {
				fr := &c.Effects[e].Fragments[f]
				nv := fr.LumpQWC / 3
				cb := base + fr.Off + (uint32(fr.ByteData[12])+1)*4
				for v := 0; v < nv+2; v++ {
					o := cb + uint32(v)*4
					if int(o)+4 > len(ram) {
						break
					}
					idx := gbase + v + 1
					ram[o] = byte(idx)
					ram[o+1] = byte(idx >> 8)
					ram[o+2] = 0
					ram[o+3] = 0x80
				}
				frags = append(frags, frag{gbase, fr})
				gbase += nv
			}
		}
	}
	fmt.Printf("%d fragments, %d vertices encoded\n", len(frags), gbase)

	// find merc chain heads: {0x1000000A}{0}{STCYCL 4,4}{UNPACK V4-32}
	var heads []uint32
	for a := uint32(0); a+16 < uint32(len(ram)); a += 16 {
		if binary.LittleEndian.Uint32(ram[a:]) == 0x1000000A &&
			binary.LittleEndian.Uint32(ram[a+8:]) == 0x01000404 {
			heads = append(heads, a)
		}
	}
	fmt.Println("chain heads:", len(heads))

	// dry pass over all heads to collect the resident microcode
	pre := merc.NewReplayer(ram)
	pre.SkipRun = true
	for _, h := range heads {
		if err := pre.Play(h); err != nil {
			fmt.Printf("dry head 0x%X: %v\n", h, err)
		}
	}
	fmt.Printf("micro head after dry: % X\n", pre.V.Micro[:24])
	total, resolved := 0, 0
	for _, h := range heads {
		r := merc.NewReplayer(ram)
		r.RefLo, r.RefHi = base, base+0x2B000
		copy(r.V.Micro, mercMicro)
		if err := r.Play(h); err != nil {
			fmt.Printf("head 0x%X: %v (packets so far %d)\n", h, err, len(r.Packets))
		}
		hres := 0
		for _, pk := range r.Packets {
			for _, v := range pk {
				total++
				if v.RGBA[3] == 0x80 && v.RGBA[2] == 0 && v.RGBA[1] < 0x40 {
					enc := int(v.RGBA[0]) | int(v.RGBA[1])<<8
					if enc >= 1 && enc <= gbase {
						resolved++
						hres++
					}
				}
			}
		}
		if hres > 0 || r.RefHits > 0 {
			fmt.Printf("head 0x%X: resolved %d, logo-refs %d, packets %d\n", h, hres, r.RefHits, len(r.Packets))
		}

	}
	fmt.Printf("total packet verts %d, resolved %d\n", total, resolved)
	_ = resolved
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mercreplay:", err)
		os.Exit(1)
	}
}
