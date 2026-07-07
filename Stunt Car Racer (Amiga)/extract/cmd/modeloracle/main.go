// modeloracle verifies the Go reimplementation of the engine's track-model bake
// ($65BEC, track/model.go) against the real engine on the tools/m68k core: it runs
// the race-init chain ($5AE46 load, $64304 grid, $5A794 start section, $696FC config)
// and then the actual $65BEC, reads back the baked per-rung records ($7ABDA, indexed
// per section at $7AA1A) and compares every record — word0, plan X/Z and height for
// both rails — against track.Bake. Per the project rule the oracle only verifies; it
// is never the source of shipped data.
//
// Usage: modeloracle game.dec.bin [trackid]
package main

import (
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/tools/m68k"
	"stuntcar/extract/track"
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
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: modeloracle game.dec.bin [trackid]")
		os.Exit(2)
	}
	img, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "modeloracle:", err)
		os.Exit(1)
	}
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if len(os.Args) >= 3 {
		id, _ := strconv.Atoi(os.Args[2])
		ids = []int{id}
	}

	fail := 0
	for _, id := range ids {
		if !run(img, id) {
			fail++
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
}

func run(img []byte, id int) bool {
	bus := &flatBus{m: make([]byte, 1<<24)}
	copy(bus.m[base:], img)
	c := m68k.NewCPU(bus)
	rd16 := func(a uint32) int { return int(bus.m[a])<<8 | int(bus.m[a+1]) }
	rd32 := func(a uint32) uint32 {
		return uint32(bus.m[a])<<24 | uint32(bus.m[a+1])<<16 | uint32(bus.m[a+2])<<8 | uint32(bus.m[a+3])
	}
	call := func(pc uint32, regs map[int]uint32) {
		c.A[7] = stackTop - 4
		r := uint32(sentinel)
		bus.Write(c.A[7], byte(r>>24))
		bus.Write(c.A[7]+1, byte(r>>16))
		bus.Write(c.A[7]+2, byte(r>>8))
		bus.Write(c.A[7]+3, byte(r))
		for reg, v := range regs {
			c.D[reg] = v
		}
		c.PC = pc
		for steps := 0; c.PC != sentinel; steps++ {
			if c.Halted || steps > 20_000_000 {
				fmt.Printf("HALT/cap at $%X\n", c.PC)
				os.Exit(1)
			}
			c.Step()
		}
	}

	// race-init chain (as cmd/physoracle), then the real bake
	call(0x5AE46, map[int]uint32{1: uint32(id)})
	call(0x64304, nil)
	call(0x5A794, nil)
	call(0x696FC, nil)
	if v := rd16(0x1BB22); v != 0 {
		fmt.Printf("track %d: NOTE $1BB22=%04X (expected 0)\n", id, v)
	}
	if v := rd16(0x1BB26); v != 0 {
		fmt.Printf("track %d: NOTE $1BB26=%04X (expected 0)\n", id, v)
	}
	if v := bus.m[0x1BB7F]; v != 0 {
		fmt.Printf("track %d: NOTE $1BB7F=%02X (expected 0)\n", id, v)
	}
	call(0x65BEC, map[int]uint32{1: 0, 2: 0})

	im := track.New(img)
	t := im.Spine(id)
	model := im.Bake(&t)

	n := int(bus.m[0x1CA1A])
	if n != len(t.Nodes) {
		fmt.Printf("track %d: section count mismatch oracle %d go %d\n", id, n, len(t.Nodes))
		return false
	}
	end := rd32(0x66102)
	mism, total := 0, 0
	for sec := 0; sec < n; sec++ {
		start := rd32(0x7AA1A + uint32(sec*4))
		secEnd := end
		if sec+1 < n {
			secEnd = rd32(0x7AA1A + uint32((sec+1)*4))
		}
		var orecs [][7]int
		for p := start; p+14 <= secEnd; p += 14 {
			var r [7]int
			for i := 0; i < 7; i++ {
				r[i] = rd16(p + uint32(2*i))
			}
			orecs = append(orecs, r)
		}
		grecs := model.Sections[sec]
		if len(orecs) != len(grecs) {
			fmt.Printf("track %d sec %d: record count oracle %d go %d\n", id, sec, len(orecs), len(grecs))
			mism++
			continue
		}
		for i, or := range orecs {
			g := grecs[i]
			gr := [7]int{g.Word0,
				int(uint16(int16(g.L.X))), int(uint16(int16(g.L.Z))), int(uint16(int16(g.L.H))),
				int(uint16(int16(g.R.X))), int(uint16(int16(g.R.Z))), int(uint16(int16(g.R.H)))}
			if or != gr {
				mism++
				total++
				fmt.Printf("track %d sec %d rec %d (d1=%d):\n  oracle %04X L(%d,%d,h%d) R(%d,%d,h%d)\n  go     %04X L(%d,%d,h%d) R(%d,%d,h%d)\n",
					id, sec, i, g.D1,
					or[0], int16(uint16(or[1])), int16(uint16(or[2])), int16(uint16(or[3])), int16(uint16(or[4])), int16(uint16(or[5])), int16(uint16(or[6])),
					gr[0], g.L.X, g.L.Z, g.L.H, g.R.X, g.R.Z, g.R.H)
				if total > 12 {
					fmt.Println("  ... (truncated)")
					return false
				}
			}
		}
		total += len(orecs)
	}
	if mism == 0 {
		fmt.Printf("track %d: OK — baked model (%d sections, %d records) Go == $65BEC oracle (byte-exact)\n", id, n, total)
		return true
	}
	fmt.Printf("track %d: %d MISMATCHES\n", id, mism)
	return false
}
