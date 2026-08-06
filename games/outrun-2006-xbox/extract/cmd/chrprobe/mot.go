package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func u32(b []byte, o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }
func u16(b []byte, o int) uint16 { return binary.LittleEndian.Uint16(b[o:]) }

// half decodes an IEEE 754 half-precision float.
func half(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h >> 10 & 0x1F)
	man := uint32(h & 0x3FF)
	switch exp {
	case 0:
		if man == 0 {
			return math.Float32frombits(sign)
		}
		for man&0x400 == 0 {
			man <<= 1
			exp--
		}
		exp++
		man &= 0x3FF
		return math.Float32frombits(sign | (exp+112)<<23 | man<<13)
	case 31:
		return math.Float32frombits(sign | 0xFF<<23 | man<<13)
	}
	return math.Float32frombits(sign | (exp+112)<<23 | man<<13)
}

type motKey struct {
	frame int
	flags byte
	val   float32
	extra []float32
}

type motChannel struct {
	bone int
	name string
	keys []motKey
}

type motClipData struct {
	id, seq, frames int
	channels        []motChannel
}

var chanNames = [6]string{"rx", "ry", "rz", "tx", "ty", "tz"}

// parseMot decodes a mot_*.bin motion container.
//
// Directory: 0x18-byte entries {id u32, seq u32, frames u32, descBytes u32,
// dataOff u32, sizeWords u32}, terminated by an entry with id 0xFFFF in its
// seq slot. Clip data: lead u32 (0), then descBytes of {canonicalBoneId u8,
// chanMask u8} descriptors (mask bits 0-5 = rx ry rz tx ty tz), then the
// curve stream, then a {01 20} sentinel.
//
// Curve stream: per enabled channel, in descriptor order, a key list of
// tokens {frame u8, flags u8, value f16}. Flag 0x20 = single-key constant
// (list ends). Flag 0x80 carries one extra f16 after the value, 0xC0 two
// extras (Hermite-style tangents). The list ends after the key whose frame
// is frames-1.
func parseMot(b []byte) ([]motClipData, error) {
	var clips []motClipData
	for o := 0; o+0x18 <= len(b); o += 0x18 {
		if u16(b, o+4) == 0xFFFF {
			break
		}
		id, seq, frames := int(u32(b, o)), int(u32(b, o+4)), int(u32(b, o+8))
		descBytes, dataOff := int(u32(b, o+12)), int(u32(b, o+16))
		if dataOff <= 0 || dataOff >= len(b) {
			break
		}
		c := motClipData{id: id, seq: seq, frames: frames}
		// the u32 at dataOff is pad (it may hold the previous clip's 01 20
		// sentinel); descriptors always start at dataOff+4
		p := dataOff + 4
		type desc struct{ bone, mask byte }
		var descs []desc
		for i := 0; i < descBytes/2; i++ {
			descs = append(descs, desc{b[p], b[p+1]})
			p += 2
		}
		for _, d := range descs {
			for bit := 0; bit < 6; bit++ {
				if d.mask&(1<<bit) == 0 {
					continue
				}
				ch := motChannel{bone: int(d.bone), name: chanNames[bit]}
				for {
					if p+2 > len(b) {
						return nil, fmt.Errorf("clip %#x: stream overrun at %#x", id, p)
					}
					frame, flags := int(b[p]), b[p+1]
					p += 2
					// low flag bits extend the 8-bit frame (clips reach 260 frames)
					frame |= int(flags&0x1F) << 8
					k := motKey{frame: frame, flags: flags}
					// payload halfwords by top bits: 00→none (implied value),
					// 20/40→value, 80→value+tangent, C0→value+2 tangents
					nval := 0
					switch flags & 0xE0 {
					case 0x00:
					case 0x20, 0x40:
						nval = 1
					case 0x80:
						nval = 2
					case 0xC0:
						nval = 3
					default:
						return nil, fmt.Errorf("clip %#x bone %d %s: flags %#02x at %#x", id, d.bone, ch.name, flags, p-2)
					}
					for e := 0; e < nval; e++ {
						if p+2 > len(b) {
							return nil, fmt.Errorf("clip %#x: stream overrun at %#x", id, p)
						}
						v := half(u16(b, p))
						p += 2
						if e == 0 {
							k.val = v
						} else {
							k.extra = append(k.extra, v)
						}
					}
					ch.keys = append(ch.keys, k)
					if motTrace {
						fmt.Printf("    @%#x bone%d %s f%d fl%#02x v%.4f %v\n", p, d.bone, ch.name, k.frame, k.flags, k.val, k.extra)
					}
					if k.flags&0x20 != 0 || k.frame >= frames-1 {
						break
					}
					if len(ch.keys) > 4096 {
						return nil, fmt.Errorf("clip %#x bone %d %s: runaway key list", id, d.bone, ch.name)
					}
				}
				c.channels = append(c.channels, ch)
			}
		}
		// loose consistency: sizeWords ≈ curve-stream u16 count + sentinel
		curveStart := dataOff + 4 + descBytes
		size := int(u32(b, o+20))
		if d := curveStart + 2*size - p; d < 0 || d > 4 {
			return nil, fmt.Errorf("clip %#x: stream ended %#x, size field implies %#x", id, p, curveStart+2*size)
		}
		clips = append(clips, c)
	}
	return clips, nil
}

func motDump(path string, verbose bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	clips, err := parseMot(b)
	if err != nil {
		fatal("%s: %v", path, err)
	}
	fmt.Printf("%s: %d clips\n", path, len(clips))
	for ci, c := range clips {
		nk := 0
		animated := 0
		for _, ch := range c.channels {
			nk += len(ch.keys)
			if len(ch.keys) > 1 {
				animated++
			}
		}
		fmt.Printf("clip %2d: id=%#06x seq=%d frames=%3d channels=%d keys=%d animated=%d\n",
			ci, c.id, c.seq, c.frames, len(c.channels), nk, animated)
		if verbose {
			for _, ch := range c.channels {
				if len(ch.keys) == 1 {
					fmt.Printf("   bone%3d %s const %8.4f\n", ch.bone, ch.name, ch.keys[0].val)
					continue
				}
				fmt.Printf("   bone%3d %s keys:", ch.bone, ch.name)
				for _, k := range ch.keys {
					fmt.Printf(" [f%d %#02x %.4f", k.frame, k.flags, k.val)
					for _, e := range k.extra {
						fmt.Printf(" e%.4f", e)
					}
					fmt.Printf("]")
				}
				fmt.Println()
			}
		}
	}
}

var motTrace = os.Getenv("MOTTRACE") != ""
