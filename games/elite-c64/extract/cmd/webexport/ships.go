// ships.go is the objects stage: every decodable XX21 ship blueprint becomes a
// wireframe3d object. The document carries the full hidden-surface blueprint —
// each face's outward normal and a point on it, and each edge with the two
// faces beside it. The viewer back-face tests every face by the sign of
// dot(normal, eye - faceCenter) and draws an edge only when one of its faces
// is visible — Elite's own render rule (Elite.md Part IV §1). Face nibble $F
// (15) on an edge means "no face here"; such an edge is always drawn.
package main

import (
	"fmt"
	"strings"

	"retroreverse.com/games/elite-c64/extract/shipmodel"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

func exportShips(ctx *cli.Context) error {
	mem, err := shipmodel.LoadEngine(ctx.In)
	if err != nil {
		return err
	}
	ships := shipmodel.ParseAll(mem)

	for i, s := range ships {
		name := shipNames[s.Type]
		if name == "" {
			name = fmt.Sprintf("Ship %d", s.Type)
		}
		wf, radius := buildWireframe(s)
		id := slug(name)
		ctx.Builder.AddObject(schema.Asset{ID: id, Name: name, Group: "Ships"}, &schema.Object{
			Type:      schema.ObjectWireframe3D,
			Name:      name,
			Wireframe: wf,
			Props: map[string]any{
				"type":   fmt.Sprintf("0x%02X", s.Type),
				"radius": radius,
				"source": "XX21 blueprint table; drawn with the game's own face-visibility rule",
			},
		})
		ctx.Progress("objects", i+1, len(ships), fmt.Sprintf("%-24s %2d verts %2d edges %2d faces",
			id, len(wf.Positions)/3, len(wf.Edges), len(wf.Faces)))
	}
	return nil
}

// slug turns a ship name into a stable kebab-case id ("Cobra Mk III (pirate)"
// -> "cobra-mk-iii-pirate"). Names are unique across the XX21 table, so the
// slug is a stable per-ship identity.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}
