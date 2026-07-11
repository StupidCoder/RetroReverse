package clv

// The stage decoder is exercised against the real image: first_us.arc's GARC
// carries the GIMG sector directory that locates st_flower01.clv inside
// DATA.BIN, and the parsed stage must reproduce the invariants pinned from the
// running game (the layout bounds and grid, the relocation-table size with
// every slot valid, the strip totals, and the terrain material by name).
// Strip-level ground truth is the game's own display lists: every in-image
// PRIM of a PSP_GE_DEBUG log matches a decoded strip's offset and vertex
// count (cmd/clvdump -verify).

import (
	"os"
	"testing"

	"retroreverse.com/games/loco-roco-psp/extract/garc"
	"retroreverse.com/games/loco-roco-psp/extract/gprs"
	"retroreverse.com/tools/platform/psp"
)

func TestParseStFlower01(t *testing.T) {
	var im *psp.Image
	for _, p := range []string{os.Getenv("PSP_IMAGE"), "../../image/LocoRoco.cso"} {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			var err error
			if im, err = psp.OpenImage(p); err != nil {
				t.Fatalf("open %s: %v", p, err)
			}
			break
		}
	}
	if im == nil {
		t.Skip("no PSP image (set PSP_IMAGE)")
	}
	defer im.Close()

	raw, err := im.ReadFile("PSP_GAME/USRDIR/data/first_us.arc")
	if err != nil {
		t.Fatalf("read first_us.arc: %v", err)
	}
	dec, err := gprs.Decompress(raw)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	a, err := garc.Parse(dec)
	if err != nil {
		t.Fatalf("garc: %v", err)
	}
	sect, ok := a.Find("sector_usa.bin")
	if !ok {
		t.Fatalf("no sector_usa.bin in first_us.arc")
	}
	dir, err := garc.ParseGimg(a.Data(sect))
	if err != nil {
		t.Fatalf("gimg: %v", err)
	}
	if len(dir) != 249 {
		t.Fatalf("gimg entries = %d, want 249", len(dir))
	}
	var stage garc.GimgEntry
	for _, e := range dir {
		if e.Name == "st_flower01.clv" {
			stage = e
		}
	}
	if stage.Name == "" {
		t.Fatalf("st_flower01.clv not in directory")
	}
	// the pack's base LBN is 23472 (DATA.BIN : LBN[23472]); read the raw extent
	clvRaw, err := im.Volume.ReadFile(rawExtentPath(23472+stage.Sector, stage.Size))
	if err != nil {
		t.Fatalf("read stage extent: %v", err)
	}

	c, err := Parse(clvRaw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(c.Data) != 3348480 {
		t.Errorf("image size = %d, want 3348480", len(c.Data))
	}
	if len(c.Reloc) != 24510 {
		t.Errorf("reloc slots = %d, want 24510", len(c.Reloc))
	}
	L := c.Layout
	if L.X != -1800 || L.Y != -1400 || L.W != 3600 || L.H != 2800 ||
		L.Cols != 9 || L.Rows != 7 || L.CellSize != 400 {
		t.Errorf("layout = %+v", L)
	}
	nb, ns, nv := 0, 0, 0
	terrain := false
	for i := range L.Cells {
		for _, b := range L.Cells[i].Batches {
			nb++
			if b.MaterialName == "stage_a_tex" {
				terrain = true
			}
			for _, s := range b.Strips {
				ns++
				nv += len(s.Verts)
			}
		}
	}
	if nb != 172 || ns != 3157 || nv != 15308 {
		t.Errorf("batches/strips/verts = %d/%d/%d, want 172/3157/15308", nb, ns, nv)
	}
	if !terrain {
		t.Errorf("no stage_a_tex batch found")
	}
}

func rawExtentPath(lbn, size uint32) string {
	return "sce_lbn0x" + hex(lbn) + "_size0x" + hex(size)
}

func hex(v uint32) string {
	const d = "0123456789ABCDEF"
	if v == 0 {
		return "0"
	}
	var b [8]byte
	i := 8
	for v != 0 {
		i--
		b[i] = d[v&0xF]
		v >>= 4
	}
	return string(b[i:])
}
