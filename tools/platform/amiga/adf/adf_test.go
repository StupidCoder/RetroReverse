package adf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Block/field offsets used to hand-build a test image (mirrors adf.go).
const (
	tbByteSize = bsize - 188                 // 0x144
	tbHashNext = bsize - 16                  // 0x1F0
	tbSecType  = bsize - 4                   // 0x1FC
	tbData0    = offTable + (tableSlots-1)*4 // first data-block pointer slot
)

type image struct{ b []byte }

func newImage(nblocks int) *image { return &image{b: make([]byte, nblocks*bsize)} }

func (im *image) put(blk, off int, v uint32) {
	binary.BigEndian.PutUint32(im.b[blk*bsize+off:], v)
}

func (im *image) name(blk int, s string) {
	o := blk*bsize + offName
	im.b[o] = byte(len(s))
	copy(im.b[o+1:], s)
}

// checksum sets a block's checksum field so its 128 longs sum to zero.
func (im *image) checksum(blk int) {
	im.put(blk, offChecksum, 0)
	var sum uint32
	for i := 0; i < bsize; i += 4 {
		sum += binary.BigEndian.Uint32(im.b[blk*bsize+i:])
	}
	im.put(blk, offChecksum, -sum)
}

// header writes a root/dir/file header skeleton and returns blk for chaining.
func (im *image) header(blk int, primary uint32, sec int32, name string) {
	im.put(blk, offType, primary)
	im.put(blk, tbSecType, uint32(sec))
	im.name(blk, name)
}

// ofsData writes an OFS data block carrying payload (must fit in 488 bytes).
func (im *image) ofsData(blk int, payload []byte) {
	im.put(blk, offType, tData)
	im.put(blk, 0x0C, uint32(len(payload))) // data_size
	copy(im.b[blk*bsize+24:], payload)
	im.checksum(blk)
}

// buildOFS lays out a tiny OFS volume:
//
//	root "TEST"
//	  HELLO            -> "Hello, Amiga!"   (same hash bucket as SUB)
//	  SUB/             (chained after HELLO)
//	    WORLD          -> "World payload."
func buildOFS(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()
	const (
		root  = 10 // nblocks/2 for a 20-block image
		hello = 11
		hdata = 12
		sub   = 13
		world = 14
		wdata = 15
	)
	hi := []byte("Hello, Amiga!")
	wo := []byte("World payload.")
	im := newImage(20)
	copy(im.b, "DOS\x00") // OFS

	im.header(root, tHeader, stRoot, "TEST")
	im.put(root, offTable, hello) // hash bucket 0 -> HELLO
	im.checksum(root)

	im.header(hello, tHeader, stFile, "HELLO")
	im.put(hello, offHighSeq, 1)
	im.put(hello, tbByteSize, uint32(len(hi)))
	im.put(hello, tbData0, hdata)
	im.put(hello, tbHashNext, sub) // chain HELLO -> SUB in the same bucket
	im.checksum(hello)
	im.ofsData(hdata, hi)

	im.header(sub, tHeader, stUserDir, "SUB")
	im.put(sub, offTable, world) // SUB's bucket 0 -> WORLD
	im.checksum(sub)

	im.header(world, tHeader, stFile, "WORLD")
	im.put(world, offHighSeq, 1)
	im.put(world, tbByteSize, uint32(len(wo)))
	im.put(world, tbData0, wdata)
	im.checksum(world)
	im.ofsData(wdata, wo)

	return im.b, hi, wo
}

func TestOpenAndRead(t *testing.T) {
	img, hello, world := buildOFS(t)
	v, err := Open(img)
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != "TEST" {
		t.Errorf("volume name = %q, want TEST", v.Name)
	}
	if v.FFS {
		t.Error("FFS flag set on an OFS image")
	}
	if !v.ChecksumOK(v.rootBlk) {
		t.Error("root checksum reported invalid")
	}

	root, err := v.ReadDir("")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]Entry{}
	for _, e := range root {
		names[e.Name] = e
	}
	// Both HELLO and SUB share hash bucket 0; the chain must surface both.
	if len(root) != 2 {
		t.Errorf("root listing has %d entries, want 2: %v", len(root), root)
	}
	if got := names["HELLO"]; got.IsDir || got.Size != len(hello) {
		t.Errorf("HELLO entry = %+v", got)
	}
	if !names["SUB"].IsDir {
		t.Error("SUB not reported as a directory")
	}

	if got, err := v.ReadFile("HELLO"); err != nil || !bytes.Equal(got, hello) {
		t.Errorf("ReadFile(HELLO) = %q, %v", got, err)
	}
	if got, err := v.ReadFile("sub/world"); err != nil || !bytes.Equal(got, world) {
		t.Errorf("ReadFile(sub/world) = %q, %v (case-insensitive path)", got, err)
	}
}

func TestWalk(t *testing.T) {
	img, _, _ := buildOFS(t)
	v, err := Open(img)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	if err := v.Walk(func(e Entry) error { paths = append(paths, e.Path); return nil }); err != nil {
		t.Fatal(err)
	}
	want := []string{"HELLO", "SUB", "SUB/WORLD"}
	if len(paths) != len(want) {
		t.Fatalf("Walk visited %v, want %v", paths, want)
	}
	for i, p := range want {
		if paths[i] != p {
			t.Errorf("Walk[%d] = %q, want %q", i, paths[i], p)
		}
	}
}

func TestRejectsNonDOS(t *testing.T) {
	if _, err := Open(make([]byte, 20*bsize)); err == nil {
		t.Error("Open accepted an image with no DOS signature")
	}
}
