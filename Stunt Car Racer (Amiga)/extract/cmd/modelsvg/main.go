// modelsvg renders each track's absolute plan (track.Geometry — the engine's own
// baked model placed at its grid cells) as a top-down SVG, and reports section-joint
// continuity: the distance between each section's last rung and the next section's
// rung 0 (they are the same physical rung; the bake skips rung 0 for exactly that
// reason). A well-formed decode shows joint distance 0 everywhere.
//
// Usage: modelsvg game.dec.bin [-out dir]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"stuntcar/extract/track"
)

var names = []string{
	"little_ramp", "stepping_stones", "hump_back", "big_ramp",
	"ski_jump", "draw_bridge", "high_jump", "roller_coaster",
}

func main() {
	out := flag.String("out", ".", "output directory")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: modelsvg game.dec.bin [-out dir]")
		os.Exit(2)
	}
	img, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "modelsvg:", err)
		os.Exit(1)
	}
	im := track.New(img)
	for id, name := range names {
		t := im.Spine(id)
		g := im.Geometry(&t)
		n := len(g)

		// joint continuity: last rung of sec k vs rung 0 of sec k+1
		worst, bad := 0, 0
		for sec := 0; sec < n; sec++ {
			a := g[sec][len(g[sec])-1]
			b := g[(sec+1)%n][0]
			d := abs(a.LX-b.LX) + abs(a.LZ-b.LZ) + abs(a.RX-b.RX) + abs(a.RZ-b.RZ)
			if d != 0 {
				bad++
				if d > worst {
					worst = d
				}
			}
		}

		minX, minZ, maxX, maxZ := 1<<30, 1<<30, -(1 << 30), -(1 << 30)
		for _, secr := range g {
			for _, r := range secr {
				for _, p := range [][2]int{{r.LX, r.LZ}, {r.RX, r.RZ}} {
					if p[0] < minX {
						minX = p[0]
					}
					if p[0] > maxX {
						maxX = p[0]
					}
					if p[1] < minZ {
						minZ = p[1]
					}
					if p[1] > maxZ {
						maxZ = p[1]
					}
				}
			}
		}
		pad := 256
		var b strings.Builder
		fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%d %d %d %d" style="background:#0d1117">`+"\n",
			minX-pad, minZ-pad, maxX-minX+2*pad, maxZ-minZ+2*pad)
		line := func(x1, z1, x2, z2 int, col string) {
			fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="14"/>`+"\n", x1, z1, x2, z2, col)
		}
		// rails: consecutive rungs within and across sections (skip each section's
		// rung 0 — it duplicates the previous section's last rung, as the bake does)
		type pt struct{ r track.RungAbs }
		var strip []track.RungAbs
		for sec := 0; sec < n; sec++ {
			for k, r := range g[sec] {
				if k == 0 && sec != 0 {
					continue
				}
				strip = append(strip, r)
			}
		}
		m := len(strip)
		for i := 0; i < m; i++ {
			a, c := strip[i], strip[(i+1)%m]
			line(a.LX, a.LZ, c.LX, c.LZ, "#54d669")
			line(a.RX, a.RZ, c.RX, c.RZ, "#54d669")
		}
		// rungs: only the baked polygon edges (the game's own wireframe density)
		for _, secr := range g {
			for _, r := range secr {
				if r.Edge && !r.Hidden {
					col := "#54d669"
					if r.Finish {
						col = "#e5c04b"
					}
					line(r.LX, r.LZ, r.RX, r.RZ, col)
				}
			}
		}
		b.WriteString("</svg>\n")
		path := filepath.Join(*out, fmt.Sprintf("%d_%s.svg", id, name))
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "modelsvg:", err)
			os.Exit(1)
		}
		fmt.Printf("track %d %-16s sections %-3d joints: %d discontinuous (worst |d|=%d)  -> %s\n",
			id, name, n, bad, worst, path)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
