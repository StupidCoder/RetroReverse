// caroracle traces and verifies the in-race car object renderer ($67AA6/$599E2)
// on the tools/m68k core: the car the game draws on the track is not a stored
// vertex model but a procedural build from the physics rig — four wheel contact
// points ($1BD8E plan positions, $1BF64 heights), a body rectangle and canopy
// box derived as fixed fractions of the projected wheelbase vectors ($5A0CE,
// $59B7A, $59C7C, $59DF6), and an unrolled face-fill list with hard-coded
// palette colours ($67AA6). The oracle runs the real race init, places the
// opponent by its AI nav state exactly as the race init does ($5D3C6), runs the
// game's own per-frame opponent chain ($6076C distance, $63A58 body plane,
// $641B6 wheel heights, $5A186 wheel placement on the decoded track surface),
// calls the real draw with PC hooks on every edge emission ($668A8), and
// verifies carmodel.Build — the verbatim Go port of the construction — against
// every emitted edge, coordinate-exact.
//
// Usage: caroracle -in game.dec.bin [-id N] [-yaw N]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/carmodel"
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
	yawAdj := flag.Int("yaw", 49152, "camera yaw adjustment (view direction toward the car)")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: caroracle -in game.dec.bin [-id N] [-yaw N]")
		os.Exit(2)
	}
	img, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "caroracle:", err)
		os.Exit(1)
	}

	bus := &flatBus{m: make([]byte, 1<<24)}
	copy(bus.m[base:], img)
	c := m68k.NewCPU(bus)
	rw := func(a uint32) int16 { return int16(uint16(bus.m[a])<<8 | uint16(bus.m[a+1])) }

	type edge struct {
		cursor         int
		d1, d2         int
		ax, ay, bx, by int // endpoint coords captured at emit time
	}
	var edges []edge
	var preShift [4][2]int // $118.. deck slots before the $59AB8/$59AEE shifts
	type fill struct {
		pc    uint32
		color int
	}
	var fills []fill
	hooks := map[uint32]func(){
		0x668A8: func() {
			d1, d2 := int(c.D[1]&0xFFFF), int(c.D[2]&0xFFFF)
			edges = append(edges, edge{
				int(rw(0x68148)), d1, d2,
				int(rw(uint32(0x1BFB0 + uint32(d1)))), int(rw(uint32(0x1C0F0 + uint32(d1)))),
				int(rw(uint32(0x1BFB0 + uint32(d2)))), int(rw(uint32(0x1C0F0 + uint32(d2)))),
			})
		},
		0x59AB8: func() { // deck slots still unshifted here
			for i, d1 := range []uint32{0x118, 0x11A, 0x11C, 0x11E} {
				preShift[i] = [2]int{int(rw(0x1BFB0 + d1)), int(rw(0x1C0F0 + d1))}
			}
		},
	}
	// the unrolled face fills: hook each JSR $66618/$6950C site inside $67AA6
	// and record the colour loaded just before (d0)
	for _, pc := range []uint32{0x67BAA, 0x67C4E, 0x67CEA, 0x67D86, 0x67E22, 0x67EBE, 0x67F5A, 0x67FF6, 0x68092, 0x6812E} {
		pc := pc
		hooks[pc] = func() { fills = append(fills, fill{pc, int(c.D[0] & 0xFF)}) }
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
	call(0x696FC, nil)
	// camera: the game's own player-nav -> camera-state routine ($605B6, as the
	// race loop calls at $5D402) with the player at the start section
	bus.m[0x1BB57] = 1
	bus.m[0x1BC42] = 0xFE // horizon on screen for a level view (y = (persp-$1BC42)/8)
	bus.m[0x1BC43] = 0x00
	bus.m[0x1BB0A], bus.m[0x1BB0B] = 0, 0
	call(0x605B6, map[int]uint32{1: uint32(bus.m[0x1CA1B])})
	// the start-grid camera faces the race direction; -yawadj turns it toward
	// the opponent placed further along the lap
	yaw := uint16(bus.m[0x1BCE6])<<8 | uint16(bus.m[0x1BCE7])
	yaw += uint16(*yawAdj)
	bus.m[0x1BCE6], bus.m[0x1BCE7] = byte(yaw>>8), byte(yaw)
	bus.m[0x1BCDC], bus.m[0x1BCDD] = 0x06, 0x00 // eye height ~1536, above the car

	// compute the car's wheel/body arrays: the opponent is placed by its AI nav
	// state (section $1BB1D, distance $1BB0C, across $1BB13) and $5A186 rides
	// its four wheels on the decoded track surface. Save the arrays, move the
	// viewpoint away and rebuild the camera, then restore the car geometry so
	// the draw projects it from the outside.
	// seed the opponent exactly as the race init does ($5D3C6: nav section =
	// $1CA1B, dist 4, along $4C) but 1024 track units further so the distance
	// gate passes, then run the game's own per-frame opponent chain: $6076C
	// (distance/visibility), $63A58 (body plane = surface + $68 + bounce),
	// the $641B6 tail (wheel heights = body + $50, rear extrapolation), and
	// $5A186 (wheel plan positions on the decoded track surface)
	bus.m[0x1BB1D] = bus.m[0x1CA1B]
	bus.m[0x1BB0C], bus.m[0x1BB0D] = 0x04, 0x00
	bus.m[0x1BBEC], bus.m[0x1BBED] = 0x00, 0x4C
	bus.m[0x1BB13] = 0x80
	bus.m[0x1BB1C] = bus.m[0x1CA1B]
	bus.m[0x1BB0A], bus.m[0x1BB0B] = 0, 0
	call(0x64E18, nil)
	call(0x6076C, nil)
	call(0x63A58, nil)
	call(0x641B6, nil)
	call(0x5A186, nil)
	fmt.Printf("AI nav: sec %d dist %d across %d $1BB12 %d\n",
		bus.m[0x1BB1D], rw(0x1BB0C), bus.m[0x1BB13], bus.m[0x1BB12])
	fmt.Printf("vis $1BBB8=%02X dist $1BC38=%d pitch $1BBEC=%d strip $1BC12=%d\n",
		bus.m[0x1BBB8], rw(0x1BC38), rw(0x1BBEC), rw(0x1BC12))
	fmt.Printf("car pos %d,%d,%d yaw %d\n", rw(0x1BCD8), rw(0x1BCDC), rw(0x1BCE0), rw(0x1BCE6))
	fmt.Printf("wheels X: ")
	for i := uint32(0); i < 0x20; i += 2 {
		fmt.Printf("%d ", rw(0x1BD8E+i))
	}
	fmt.Printf("\nwheels Z: ")
	for i := uint32(0); i < 0x20; i += 2 {
		fmt.Printf("%d ", rw(0x1BD8E+0x20+i))
	}
	fmt.Println()
	fmt.Printf("wheel heights $1BF64: %d %d %d %d\n", rw(0x1BF64), rw(0x1BF66), rw(0x1BF68), rw(0x1BF6A))
	fmt.Printf("body heights $1BD86: %d %d %d  $1BD66: %d %d %d %d\n", rw(0x1BD86), rw(0x1BD88), rw(0x1BD8A),
		rw(0x1BD66), rw(0x1BD68), rw(0x1BD6A), rw(0x1BD6C))
	bus.m[0x1BB8F] = 0 // race projection (no preview squeeze)
	edges = edges[:0]
	fills = fills[:0]
	call(0x67AA6, nil)

	fmt.Printf("\n%d edges emitted:\n", len(edges))
	for _, e := range edges {
		fmt.Printf("  list+%03X: v%03X -> v%03X   (%d,%d) -> (%d,%d)\n",
			e.cursor-0x5E0, e.d1, e.d2, e.ax, e.ay, e.bx, e.by)
	}
	// verify the carmodel port: rebuild the construction from the engine's own
	// projected wheel points (pre-shift) and compare every emitted edge
	var bin carmodel.BuildIn
	for i, d1 := range []uint32{0xF4, 0xF6, 0xF8, 0xFA} {
		bin.Hi[i] = carmodel.Pt{X: int(rw(0x1BFB0 + d1)), Y: int(rw(0x1C0F0 + d1))}
		bin.Lo[i] = carmodel.Pt{X: preShift[i][0], Y: preShift[i][1]}
	}
	got := carmodel.Build(bin)
	bySlot := map[int][4]int{}
	for _, e := range edges {
		off := e.cursor - 0x5E0
		if off >= 0 && off%4 == 0 {
			bySlot[off/4] = [4]int{e.ax, e.ay, e.bx, e.by}
		}
	}
	mism := 0
	for _, ge := range got {
		oe, ok := bySlot[ge.Slot]
		if !ok {
			fmt.Printf("port: slot %02X missing in engine log\n", ge.Slot)
			mism++
			continue
		}
		if oe != [4]int{ge.A.X, ge.A.Y, ge.B.X, ge.B.Y} {
			fmt.Printf("port: slot %02X engine (%d,%d)-(%d,%d) port (%d,%d)-(%d,%d)\n",
				ge.Slot, oe[0], oe[1], oe[2], oe[3], ge.A.X, ge.A.Y, ge.B.X, ge.B.Y)
			mism++
		}
	}
	if mism == 0 {
		fmt.Printf("PORT OK: all %d construction edges match the engine\n", len(got))
	} else {
		fmt.Printf("PORT: %d mismatches\n", mism)
		os.Exit(1)
	}
	fmt.Printf("\n%d fills:\n", len(fills))
	for _, f := range fills {
		fmt.Printf("  pc $%X colour %d\n", f.pc, f.color)
	}
	// vertex slot world data for correlation
	fmt.Println("\nvertex slots (screenX, screenY, heightStrip):")
	for _, d1 := range []int{0xF4, 0xF6, 0xF8, 0xFA, 0x108, 0x10A, 0x10C, 0x10E, 0x118, 0x11A, 0x11C, 0x11E} {
		fmt.Printf("  v%03X: (%d, %d) h=%d\n", d1,
			rw(uint32(0x1BFB0+d1)), rw(uint32(0x1C0F0+d1)), rw(uint32(0x1BE70+d1)))
	}
}
