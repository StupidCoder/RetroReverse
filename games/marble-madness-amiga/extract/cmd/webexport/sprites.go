// sprites.go is the sprites stage: it renders the scenery-overlay sprite pieces the
// level objects reference (the drawbridge, goal flags, Aerial's pistons, the WAVE,
// the vacuum hoods, …) into sprites/*.png and writes the sprites/index.json the
// viewer resolves each object's "sprite" key against (Marble_Madness.md Part IV §5 /
// Part V §2). One entry per distinct cell (static) or per animated region strip.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/platform/amiga/adf"
)

// exportSprites writes the sprites/ tree (overlay PNGs + index.json). Each course's
// overlay pieces are rendered with that course's colour-band palette at each piece's row.
func exportSprites(vol *adf.Volume, paths map[string]string, outDir string) {
	spritesDir := filepath.Join(outDir, "sprites")
	chk(os.MkdirAll(spritesDir, 0o755))

	spriteIndex := map[string]any{}
	for idx, c := range courses {
		cr, err := loadCourse(vol, paths, c.key, c.track)
		chk(err)
		before := len(spriteIndex)
		objects, _, err := exportOverlays(vol, paths, c.key, cr.prog.Image, cr.co, cr.bake.paletteAt, spritesDir, spriteIndex, true)
		chk(err)
		fmt.Fprintf(os.Stderr, "[sprites] %d/%d  %-12s %d overlay objects, %d sprite entries\n",
			idx+1, len(courses), c.name, len(objects), len(spriteIndex)-before)
	}
	chk(writeJSON(filepath.Join(spritesDir, "index.json"), spriteIndex))
	fmt.Fprintf(os.Stderr, "[sprites] done: %d sprite entries\n", len(spriteIndex))
}
