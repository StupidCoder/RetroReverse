package lev

import (
	"os"
	"path/filepath"
	"testing"
)

// The wooden door (item 320) the player meets just past the start room sits at
// tile (33,8) in level 1 — a fixed anchor that verifies the object chain decode.
func TestObjectsDoorAtStart(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "game", "DATA", "LEV.ARK"))
	if err != nil {
		t.Skip("game data not present")
	}
	ark, err := ParseArk(b)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := ark.Block(0)
	grid, _ := DecodeGrid(block)
	objs := Objects(grid, block)
	if len(objs) < 100 {
		t.Fatalf("only %d objects decoded", len(objs))
	}
	found := false
	for _, o := range objs {
		if o.TileX == 33 && o.TileY == 8 && o.ItemID == 320 {
			found = true
		}
		if o.ItemID > 511 || o.FineX > 7 || o.FineY > 7 || o.Z > 127 || o.Heading > 7 {
			t.Errorf("object out of range: %+v", o)
		}
	}
	if !found {
		t.Errorf("wooden door (item 320) not found at tile (33,8)")
	}
}
