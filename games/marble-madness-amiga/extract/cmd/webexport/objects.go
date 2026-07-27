// objects.go is the scenery-objects stage: the overlay sprite pieces the level
// placements reference (the drawbridge, goal flags, Aerial's pistons, the
// WAVE, the vacuum hoods, …) become sprite2d objects, grouped by their course
// (each course renders its pieces in its own colour-band palette). The decode
// is exportOverlays (overlay.go); here its per-piece strips are repacked onto
// a uniform cell grid, which is what the sprite2d atlas format speaks.
package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/amiga/adf"
)

// exportObjects writes the scenery sprite2d assets and returns the
// "<course>/<piece>" -> asset map placements resolve through.
func exportObjects(ctx *cli.Context, vol *adf.Volume, paths map[string]string) (map[string]objRef, error) {
	b := ctx.Builder
	refs := map[string]objRef{}

	scratch, err := os.MkdirTemp("", "mm-sprites-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)

	for idx, c := range courses {
		cr, err := loadCourse(vol, paths, c.key, c.track)
		if err != nil {
			return nil, err
		}
		index := map[string]any{}
		if _, _, err := exportOverlays(vol, paths, c.key, cr.prog.Image, cr.co, cr.bake.paletteAt, scratch, index, true); err != nil {
			return nil, err
		}

		var keys []string
		for k := range index {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			e := index[key].(map[string]any)
			frames := e["frames"].([][4]int)
			src := strings.TrimPrefix(e["src"].(string), "sprites/")
			img, err := loadPNG(filepath.Join(scratch, src))
			if err != nil {
				return nil, err
			}

			// Repack the piece's (possibly ragged) frame rects onto a uniform
			// cell grid, each frame's top-left at its cell's top-left.
			cw, ch := 0, 0
			for _, f := range frames {
				if f[2] > cw {
					cw = f[2]
				}
				if f[3] > ch {
					ch = f[3]
				}
			}
			strip := image.NewNRGBA(image.Rect(0, 0, cw*len(frames), ch))
			for i, f := range frames {
				dst := image.Rect(i*cw, 0, i*cw+f[2], f[3])
				draw.Draw(strip, dst, img, image.Pt(f[0], f[1]), draw.Src)
			}

			piece := key[strings.Index(key, "/")+1:]
			id := slugify(c.key + "-" + piece)
			f, err := b.CreateFile("objects", id+".png")
			if err != nil {
				return nil, err
			}
			err = png.Encode(f, strip)
			f.Close()
			if err != nil {
				return nil, err
			}

			anim := schema.Animation{ID: "main", Frames: len(frames), Loop: "loop"}
			if steps, ok := e["steps"].([][2]int); ok && len(steps) > 0 {
				for _, st := range steps {
					anim.Steps = append(anim.Steps, []int{st[0], st[1]})
				}
			} else if len(frames) == 1 {
				anim.Loop = "hold"
			}
			name := fmt.Sprintf("Piece %s", piece)
			b.AddObject(schema.Asset{ID: id, Name: name, Group: c.name}, &schema.Object{
				Type: schema.ObjectSprite2D,
				Name: name,
				Atlas: &schema.SpriteAtlas{
					File: id + ".png", CellW: cw, CellH: ch,
				},
				Animations: []schema.Animation{anim},
				Props:      map[string]any{"course": c.name, "piece": piece},
			})
			refs[key] = objRef{asset: id}
		}
		ctx.Progress("objects", idx+1, len(courses), fmt.Sprintf("%-12s %d overlay pieces", c.name, len(keys)))
	}
	return refs, nil
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}
