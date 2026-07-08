// Package rr decodes Ridge Racer's proprietary asset files straight from the
// CD image: the 32×32 course grid (IDX.HED), the track geometry (MAP.RRM), and
// the 3-D object models (OBJ.RRO). Every structure here was pinned by tracing
// the game's own readers under the PSX oracle (the consuming PC, then the
// disassembly around it); the decoders reimplement those readers over the file
// bytes, and cmd/geomoracle verifies the result against the running game.
//
// All three files are raw (no compression) and little-endian. The boot loader
// streams each file whole into main RAM; MAP.RRM (0x8008045C) and OBJ.RRO
// (0x800C2918) stay resident and the renderer walks them in place.
package rr

import (
	"encoding/binary"
	"fmt"
)

func u16(b []byte, off int) uint16 { return binary.LittleEndian.Uint16(b[off:]) }
func s16(b []byte, off int) int16  { return int16(binary.LittleEndian.Uint16(b[off:])) }
func u32(b []byte, off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }

// UV is one texel coordinate as the GPU packet encodes it.
type UV struct{ U, V byte }

func uvOf(h uint16) UV { return UV{byte(h), byte(h >> 8)} }

// errTruncated reports a file shorter than its own directory claims.
func errTruncated(name string, want, have int) error {
	return fmt.Errorf("%s: directory needs %d bytes, file has %d", name, want, have)
}
