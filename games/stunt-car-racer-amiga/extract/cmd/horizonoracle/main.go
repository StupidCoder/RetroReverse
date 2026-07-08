// horizonoracle verifies the Go decode of the horizon mountain range
// (track.Horizon) against the engine's yaw-placed object renderer $6953E on
// the tools/m68k core. It runs the race-init chain (whose $696FC fills the
// placement tables from the config list at $69A80), builds a level camera,
// then calls the real $6953E at several compass headings with PC hooks on
// every edge emission ($668A8) and face-colour set ($6770A), and compares
// each visible object's edges — placed with the engine's own formula
// (x_left = (yaw*256 - cameraYaw16) >> 3, y = -($1BC42>>3) - y_model) —
// coordinate-exact, plus the face colour sequence. Per the project rule the
// oracle only verifies; track.Horizon decodes purely from the image.
//
// Usage: horizonoracle -in game.dec.bin [-id N]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/track"
	"retroreverse.com/tools/cpu/m68k"
)

const (
	base     = 0xE700
	sentinel = 0xFFFFFE
	stackTop = 0x300000
)

type flatBus struct{ m []byte }

func (b *flatBus) Read(a uint32) byte     { return b.m[a&0xFFFFFF] }
func (b *flatBus) Write(a uint32, v byte) { b.m[a&0xFFFFFF] = v }

func main() {
	in := flag.String("in", "", "input decoded game binary (game.dec.bin)")
	idFlag := flag.Int("id", 0, "track id")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: horizonoracle -in game.dec.bin [-id N]")
		os.Exit(2)
	}
	img, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "horizonoracle:", err)
		os.Exit(1)
	}

	bus := &flatBus{m: make([]byte, 1<<24)}
	copy(bus.m[base:], img)
	c := m68k.NewCPU(bus)
	rw := func(a uint32) int16 { return int16(uint16(bus.m[a])<<8 | uint16(bus.m[a+1])) }

	type edge struct{ ax, ay, bx, by int }
	var edges []edge
	var colors []int
	hooks := map[uint32]func(){
		0x668A8: func() {
			d1, d2 := uint32(c.D[1]&0xFFFF), uint32(c.D[2]&0xFFFF)
			edges = append(edges, edge{
				int(rw(0x1BFB0 + d1)), int(rw(0x1C0F0 + d1)),
				int(rw(0x1BFB0 + d2)), int(rw(0x1C0F0 + d2)),
			})
		},
		0x6964C: func() { colors = append(colors, int(c.D[0]&0xFF)) },
	}

	call := func(pc uint32, regs map[int]uint32) {
		c.A[7] = stackTop - 4
		r := uint32(sentinel)
		for i := uint32(0); i < 4; i++ {
			bus.Write(c.A[7]+i, byte(r>>(24-8*i)))
		}
		for reg, v := range regs {
			c.D[reg] = v
		}
		c.PC = pc
		for steps := 0; c.PC != sentinel; steps++ {
			if c.Halted || steps > 50_000_000 {
				fmt.Printf("HALT/cap at $%X\n", c.PC)
				os.Exit(1)
			}
			if h, ok := hooks[c.PC]; ok {
				h()
			}
			c.Step()
		}
	}

	id := *idFlag
	bus.m[0x1CA33] = byte(id)
	call(0x5AE46, map[int]uint32{1: uint32(id)})
	call(0x64304, nil)
	call(0x5A794, nil)
	call(0x696FC, nil) // fills $69736/$69766/$69799 from the $69A80 config
	// level camera (identity roll matrix), as cmd/caroracle
	bus.m[0x1BB0A], bus.m[0x1BB0B] = 0, 0
	call(0x605B6, map[int]uint32{1: uint32(bus.m[0x1CA1B])})
	bus.m[0x1BC42], bus.m[0x1BC43] = 0x07, 0x00
	call(0x64E18, nil)

	im := track.New(img)
	place, models := im.Horizon()
	if int(bus.m[0x69799]) != len(place) {
		fmt.Printf("track %d: placement count oracle %d go %d\n", id, bus.m[0x69799], len(place))
		os.Exit(1)
	}

	d5 := -(int(rw(0x1BC42)) >> 3)
	// the $62518 roll rotation the renderer applies per vertex: fixed-point
	// rotate by the camera matrix, >>2 scale and screen re-centre
	m20, m22 := int(rw(0x1C230+0x20)), int(rw(0x1C230+0x22))
	hw := func(m, v int) int { return int(int16(uint16((int32(m) * int32(v) * 2) >> 16))) }
	rot := func(x, y int) (int, int) {
		sx := int(int16(uint16(hw(m22, x)+hw(m20, y))))>>2 + 0x80
		sy := int(int16(uint16(hw(m22, y)-hw(m20, x))))>>2 + 0x40
		return sx, sy
	}
	mism, checked := 0, 0
	for _, camHi := range []int{0x00, 0x20, 0x40, 0x60, 0x80, 0xA0, 0xC0, 0xE0} {
		bus.m[0x1BCE6], bus.m[0x1BCE7] = byte(camHi), 0
		edges = edges[:0]
		colors = colors[:0]
		call(0x6953E, nil)

		// expected: entries processed from the last placement down, visible
		// when (yaw - camHi + $1C) & $FF < $2C
		var wantE []edge
		var wantC []int
		for i := len(place) - 1; i >= 0; i-- {
			p := place[i]
			if (int(p.Yaw)-camHi+0x1C)&0xFF >= 0x2C {
				continue
			}
			dy := int(int8(byte(int(p.Yaw) - camHi)))
			xl := (dy * 256) >> 3
			m := models[p.Model]
			for _, e := range m.Edges {
				a, b := m.Verts[e[0]], m.Verts[e[1]]
				ax, ay := rot(xl+a[0], d5-a[1])
				bx, by := rot(xl+b[0], d5-b[1])
				wantE = append(wantE, edge{ax, ay, bx, by})
			}
			for _, f := range m.Faces {
				wantC = append(wantC, int(f.Pal))
			}
		}
		if len(edges) != len(wantE) || len(colors) != len(wantC) {
			fmt.Printf("track %d cam %02X: oracle %d edges/%d colours, go %d/%d\n",
				id, camHi, len(edges), len(colors), len(wantE), len(wantC))
			mism++
			continue
		}
		for i := range edges {
			if edges[i] != wantE[i] {
				fmt.Printf("track %d cam %02X edge %d: oracle %v go %v\n", id, camHi, i, edges[i], wantE[i])
				mism++
			}
		}
		for i := range colors {
			if colors[i] != wantC[i] {
				fmt.Printf("track %d cam %02X colour %d: oracle %d go %d\n", id, camHi, i, colors[i], wantC[i])
				mism++
			}
		}
		checked += len(edges)
	}
	if mism == 0 {
		fmt.Printf("track %d: OK — horizon decode matches $6953E over 8 headings (%d edges, %d placements, %d models)\n",
			id, checked, len(place), len(models))
		return
	}
	fmt.Printf("track %d: %d MISMATCHES\n", id, mism)
	os.Exit(1)
}
