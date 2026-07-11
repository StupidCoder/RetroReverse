package psp

// cso.go is a random-access reader for the CISO ("CSO") compressed-ISO container
// that UMD dumps are commonly stored in. A CISO holds an ISO 9660 image as a run
// of fixed-size logical blocks, each stored either raw or raw-DEFLATE compressed,
// with an index table giving every block's byte offset. The 2048-byte block size
// matches the ISO logical block, so the CISO block index *is* the ISO LBA — the
// iso.go reader sits straight on top of ReadBlock.
//
// Header (little-endian):
//
//	0x00  4  magic "CISO"
//	0x04  4  header size (reserved; 0 in practice)
//	0x08  8  total uncompressed size (the ISO byte length)
//	0x10  4  block size (2048)
//	0x14  1  version (1)
//	0x15  1  align — index offsets are stored right-shifted by this many bits
//	0x16  2  reserved
//	0x18  .  index: (nblocks+1) uint32 entries
//
// Each index entry's bit 31 flags a block stored *uncompressed*; bits 0..30 are
// the block's file offset >> align. A block's stored length is the gap to the next
// entry's offset. Compressed blocks are raw DEFLATE (no zlib/gzip wrapper).
//
// The container is hundreds of MiB; the reader keeps only the file handle, the
// index, and a small LRU of decoded blocks resident, inflating each block on
// demand through an io.ReaderAt so the whole image is never held in memory.

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const csoBlockSize = 2048

// CSO is an opened CISO container.
type CSO struct {
	r       io.ReaderAt
	closer  io.Closer
	index   []uint32 // nblocks+1 raw index entries (top bit = uncompressed flag)
	align   uint
	total   uint64
	nblocks int

	cache *blockLRU
}

// Open opens the CISO file at path. The returned CSO must be Closed.
func Open(path string) (*CSO, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	c, err := newCSO(f, st.Size())
	if err != nil {
		f.Close()
		return nil, err
	}
	c.closer = f
	return c, nil
}

// newCSO builds a CSO over r (whose backing store is size bytes) by reading and
// validating the header and index.
func newCSO(r io.ReaderAt, size int64) (*CSO, error) {
	var hdr [0x18]byte
	if _, err := r.ReadAt(hdr[:], 0); err != nil {
		return nil, fmt.Errorf("cso: reading header: %w", err)
	}
	if string(hdr[0:4]) != "CISO" {
		return nil, fmt.Errorf("cso: bad magic %q", hdr[0:4])
	}
	total := binary.LittleEndian.Uint64(hdr[0x08:])
	block := binary.LittleEndian.Uint32(hdr[0x10:])
	if block != csoBlockSize {
		return nil, fmt.Errorf("cso: unsupported block size %d", block)
	}
	align := uint(hdr[0x15])

	nblocks := int((total + csoBlockSize - 1) / csoBlockSize)
	idxBytes := make([]byte, (nblocks+1)*4)
	if _, err := r.ReadAt(idxBytes, 0x18); err != nil {
		return nil, fmt.Errorf("cso: reading index (%d blocks): %w", nblocks, err)
	}
	index := make([]uint32, nblocks+1)
	for i := range index {
		index[i] = binary.LittleEndian.Uint32(idxBytes[i*4:])
	}
	return &CSO{
		r:       r,
		index:   index,
		align:   align,
		total:   total,
		nblocks: nblocks,
		cache:   newBlockLRU(128),
	}, nil
}

// TotalSize is the uncompressed ISO byte length.
func (c *CSO) TotalSize() int64 { return int64(c.total) }

// Close releases the underlying file handle, if any.
func (c *CSO) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}

// ReadBlock returns the decompressed 2048-byte logical block n. The returned
// slice is owned by the LRU cache — callers must not mutate it.
func (c *CSO) ReadBlock(n int) ([]byte, error) {
	if n < 0 || n >= c.nblocks {
		return nil, fmt.Errorf("cso: block %d out of range (0..%d)", n, c.nblocks-1)
	}
	if b, ok := c.cache.get(n); ok {
		return b, nil
	}
	const flag = 1 << 31
	off := uint64(c.index[n]&^flag) << c.align
	end := uint64(c.index[n+1]&^flag) << c.align
	if end < off {
		return nil, fmt.Errorf("cso: block %d has negative length", n)
	}
	stored := make([]byte, end-off)
	if _, err := c.r.ReadAt(stored, int64(off)); err != nil {
		return nil, fmt.Errorf("cso: reading block %d payload: %w", n, err)
	}

	out := make([]byte, csoBlockSize)
	if c.index[n]&flag != 0 {
		// Stored uncompressed.
		copy(out, stored)
	} else {
		fr := flate.NewReader(bytes.NewReader(stored))
		if _, err := io.ReadFull(fr, out); err != nil && err != io.ErrUnexpectedEOF {
			fr.Close()
			return nil, fmt.Errorf("cso: inflating block %d: %w", n, err)
		}
		fr.Close()
	}
	c.cache.put(n, out)
	return out, nil
}

// ReadAt reads len(p) bytes from the uncompressed image starting at byte offset
// off, satisfying io.ReaderAt so callers can treat the CISO as a flat ISO.
func (c *CSO) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, fmt.Errorf("cso: negative offset")
	}
	got := 0
	for got < len(p) {
		pos := off + int64(got)
		if pos >= int64(c.total) {
			return got, io.EOF
		}
		blk := int(pos / csoBlockSize)
		within := int(pos % csoBlockSize)
		b, err := c.ReadBlock(blk)
		if err != nil {
			return got, err
		}
		n := copy(p[got:], b[within:])
		got += n
	}
	return got, nil
}
