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

// NARCFile is one named sub-file of a NARC.
type NARCFile struct {
	Name string // empty when the archive is nameless
	Data []byte
}

// ParseNARCFiles returns the sub-files of a NARC with their BTNF names (asset NARCs
// usually carry a flat root directory of real file names). Files the name table
// doesn't cover keep an empty name. Accepts a raw `.carc` (LZ77) directly.
func ParseNARCFiles(data []byte) ([]NARCFile, error) {
	data = Decompress(data)
	raw, err := ParseNARC(data)
	if err != nil {
		return nil, err
	}
	out := make([]NARCFile, len(raw))
	for i, d := range raw {
		out[i].Data = d
	}
	// Locate BTNF and walk its (flat) root sub-table, assigning sequential file IDs —
	// the same FNT structure the cartridge filesystem uses.
	le := binary.LittleEndian
	off := int(le.Uint16(data[0x0C:]))
	n := int(le.Uint16(data[0x0E:]))
	for i := 0; i < n && off+8 <= len(data); i++ {
		magic := string(data[off : off+4])
		size := int(le.Uint32(data[off+4:]))
		if magic == "BTNF" && size >= 16 {
			base := off + 8
			sub := base + int(le.Uint32(data[base:]))
			id := int(le.Uint16(data[base+4:]))
			for p := sub; p < len(data); {
				ctrl := int(data[p])
				p++
				if ctrl == 0 || ctrl >= 0x80 { // end, or a subdirectory (not used by asset NARCs)
					break
				}
				if p+ctrl > len(data) {
					break
				}
				if id >= 0 && id < len(out) {
					out[id].Name = string(data[p : p+ctrl])
				}
				p += ctrl
				id++
			}
			break
		}
		if size <= 0 {
			break
		}
		off += size
	}
	return out, nil
}

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
