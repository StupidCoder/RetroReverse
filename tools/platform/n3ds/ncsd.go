package n3ds

import (
	"encoding/binary"
	"fmt"
)

// The NCSD header sits at 0x100, after a 0x100-byte RSA-2048 signature over it.
const (
	ncsdSigSize    = 0x100
	ncsdHeaderOff  = 0x100
	ncsdHeaderSize = 0x100
	ncsdMagic      = "NCSD"

	// NumPartitions is the fixed size of the NCSD partition table. Slots the
	// cartridge does not use have a zero offset and length.
	NumPartitions = 8
)

// Partition is one entry of the NCSD partition table, resolved from media units
// to byte offsets into the image.
type Partition struct {
	Index  int
	Offset int64  // byte offset of the partition within the image
	Size   int64  // byte length of the partition
	FSType byte   // partitionsFsType[i]
	Crypt  byte   // partitionsCryptType[i]
	ID     uint64 // partitionIdTable[i]
}

// Empty reports whether the partition slot is unused.
func (p Partition) Empty() bool { return p.Size == 0 }

// NCSD is the parsed cartridge-image header.
type NCSD struct {
	raw []byte // the whole image

	ImageSize     int64 // total image size in bytes, from the header
	MediaID       uint64
	MediaUnitSize int64 // 512 << flags[6]
	Flags         [8]byte
	Partitions    [NumPartitions]Partition
}

// Raw returns the backing image bytes.
func (n *NCSD) Raw() []byte { return n.raw }

// ParseNCSD parses the NCSD header of a CCI image.
//
// The media-unit size is taken from flags[6] rather than assumed: every offset
// and length in both the NCSD and the NCCH headers is expressed in these units,
// and a wrong unit silently mislocates every partition.
func ParseNCSD(img []byte) (*NCSD, error) {
	if len(img) < ncsdHeaderOff+ncsdHeaderSize {
		return nil, fmt.Errorf("n3ds: image too short for an NCSD header (%d bytes)", len(img))
	}
	h := img[ncsdHeaderOff:]
	if string(h[0:4]) != ncsdMagic {
		return nil, fmt.Errorf("n3ds: not an NCSD image: magic %q at 0x%x, want %q", h[0:4], ncsdHeaderOff, ncsdMagic)
	}

	n := &NCSD{raw: img}
	copy(n.Flags[:], h[0x88:0x90])
	n.MediaUnitSize = 512 << n.Flags[6]
	n.ImageSize = int64(binary.LittleEndian.Uint32(h[0x04:])) * n.MediaUnitSize
	n.MediaID = binary.LittleEndian.Uint64(h[0x08:])

	for i := 0; i < NumPartitions; i++ {
		off := int64(binary.LittleEndian.Uint32(h[0x20+i*8:])) * n.MediaUnitSize
		size := int64(binary.LittleEndian.Uint32(h[0x24+i*8:])) * n.MediaUnitSize
		p := Partition{
			Index:  i,
			Offset: off,
			Size:   size,
			FSType: h[0x10+i],
			Crypt:  h[0x18+i],
			ID:     binary.LittleEndian.Uint64(h[0x90+i*8:]),
		}
		if !p.Empty() {
			if off < 0 || size < 0 || off+size > int64(len(img)) {
				return nil, fmt.Errorf("n3ds: partition %d [0x%x+0x%x] runs past the image end (0x%x)", i, off, size, len(img))
			}
		}
		n.Partitions[i] = p
	}
	return n, nil
}

// Bytes returns the raw bytes of partition i, or nil if the slot is unused.
func (n *NCSD) Bytes(i int) []byte {
	p := n.Partitions[i]
	if p.Empty() {
		return nil
	}
	return n.raw[p.Offset : p.Offset+p.Size]
}

// Partition parses partition i as an NCCH container.
func (n *NCSD) Partition(i int) (*NCCH, error) {
	if i < 0 || i >= NumPartitions {
		return nil, fmt.Errorf("n3ds: partition index %d out of range", i)
	}
	b := n.Bytes(i)
	if b == nil {
		return nil, fmt.Errorf("n3ds: partition %d is empty", i)
	}
	return ParseNCCH(b)
}

// Executable returns partition 0, the application NCCH (a "CXI").
func (n *NCSD) Executable() (*NCCH, error) { return n.Partition(0) }
