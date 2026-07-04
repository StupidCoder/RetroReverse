package nds

import (
	"encoding/binary"
	"fmt"
)

// NARC (Nintendo ARChive) is the DS's generic bundle format — the second layer of a
// `.carc` (an LZ77 stream, blz.go/lz77.go, wrapping a NARC), and the container many
// assets ship in. It reuses the same on-cartridge filesystem idea as the cartridge
// itself: a file-allocation table of start/end offsets, an (often flat) name table,
// and a blob of file data. Three chunks:
//
//	NARC header  : "NARC", BOM 0xFFFE, version, file size, header size, block count
//	BTAF (FAT)   : file count, then count × (u32 start, u32 end) into the image data
//	BTNF (FNT)   : the name table (usually just a flat root for asset NARCs)
//	GMIF (image) : the file data the BTAF offsets point into
//
// Offsets in BTAF are relative to the start of the GMIF data.

// ParseNARC returns the sub-files of a NARC archive, in FAT order. It transparently
// LZ77-decompresses the input first, so a raw `.carc` may be passed directly.
func ParseNARC(data []byte) ([][]byte, error) {
	data = Decompress(data)
	if len(data) < 0x10 || string(data[0:4]) != "NARC" {
		return nil, fmt.Errorf("nds: not a NARC (bad magic)")
	}
	le := binary.LittleEndian

	// Walk the block list to find BTAF and GMIF (their order is fixed in practice,
	// but scan to be safe).
	var fatOff, imgOff int
	off := int(le.Uint16(data[0x0C:])) // header size → first block
	nBlocks := int(le.Uint16(data[0x0E:]))
	for i := 0; i < nBlocks && off+8 <= len(data); i++ {
		magic := string(data[off : off+4])
		size := int(le.Uint32(data[off+4:]))
		switch magic {
		case "BTAF":
			fatOff = off
		case "GMIF":
			imgOff = off
		}
		if size <= 0 {
			break
		}
		off += size
	}
	if fatOff == 0 || imgOff == 0 {
		return nil, fmt.Errorf("nds: NARC missing BTAF/GMIF")
	}

	count := int(le.Uint16(data[fatOff+8:]))
	imgData := imgOff + 8
	out := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		e := fatOff + 0x0C + i*8
		if e+8 > len(data) {
			return nil, fmt.Errorf("nds: NARC FAT out of range")
		}
		start := imgData + int(le.Uint32(data[e:]))
		end := imgData + int(le.Uint32(data[e+4:]))
		if start > end || end > len(data) {
			return nil, fmt.Errorf("nds: NARC file %d out of range", i)
		}
		out = append(out, data[start:end])
	}
	return out, nil
}
