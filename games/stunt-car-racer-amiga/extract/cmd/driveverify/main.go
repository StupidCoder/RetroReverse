// driveverify locksteps the Go faithful-drive orchestration (the real per-frame render
// coupling + the verified $6185C physics) against the engine on the m68k core. It proves the
// car drives *grounded* — the launch artefact of the old bare-placement drive is gone once
// the real coupling runs each frame.
//
// Setup (mirrors the race intro $5D402..): load the track, place the car with the real
// $605B6 (posY = 16.0), run the pre-arm coupling + one unarmed $6185C + $64E4C, then arm
// ($1BB72). That armed placed state is snapshotted and both sides run forward from it.
//
// Per frame both sides run the decomposed coupling that is byte-identical to the render
// $64E4C for the physics-relevant state: $60190 camera, zero the view offsets $1BBD5/$1BBD6,
// $5FE04 grid->section into $1BB85, $5BE44 placement; then $6185C. The Go side uses the
// package reimplementations (Camera60190/Section5FE04/Couple5BE44/Frame6185C).
//
// Usage: driveverify -in game.dec.bin [-frames N] [-drive N]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/physics"
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

var bus *flatBus
var c *m68k.CPU

func call(pc uint32) byte {
	c.A[7] = stackTop - 4
	r := uint32(sentinel)
	for i := uint32(0); i < 4; i++ {
		bus.Write(c.A[7]+i, byte(r>>(24-8*i)))
	}
	c.PC = pc
	for steps := 0; c.PC != sentinel; steps++ {
		if c.Halted {
			fmt.Printf("HALT at $%X: %s\n", c.PC, c.HaltReason)
			os.Exit(1)
		}
		if steps > 20_000_000 {
			fmt.Printf("STEP CAP at $%X\n", c.PC)
			os.Exit(1)
		}
		c.Step()
	}
	return byte(c.D[0])
}

func callD1(pc uint32, d1 uint32) { c.D[1] = d1; call(pc) }

// engineCouple runs the decomposed per-frame coupling on the oracle.
func engineCouple() {
	call(0x60190)
	bus.Write(0x1BBD5, 0)
	bus.Write(0x1BBD6, 0)
	d0 := call(0x5FE04)
	if d0 != 0xFF {
		bus.Write(0x1BB85, d0)
		call(0x5BE44)
	}
}

// goCouple runs the same on the Go package memory.
func goCouple(m *physics.Mem) {
	m.Camera60190()
	m.B[0x1BBD5] = 0
	m.B[0x1BBD6] = 0
	sec, off := m.Section5FE04()
	if !off && byte(sec) != 0xFF {
		m.B[0x1BB85] = byte(sec)
		m.Couple5BE44()
	}
}

// checked car-state addresses (word granularity; 32-bit values are two words).
var checks = []struct {
	name string
	addr uint32
}{
	{"posXhi", 0x1BCD8}, {"posXlo", 0x1BCDA},
	{"posYhi", 0x1BCDC}, {"posYlo", 0x1BCDE},
	{"posZhi", 0x1BCE0}, {"posZlo", 0x1BCE2},
	{"roll", 0x1BCE4}, {"yaw", 0x1BCE6}, {"pit", 0x1BCE8},
	{"velX", 0x1BCEA}, {"velY", 0x1BCEC}, {"velZ", 0x1BCEE},
	{"amR", 0x1BCF0}, {"amP", 0x1BCF2}, {"amY", 0x1BCF4},
}

func rw(m []byte, a uint32) int16 { return int16(uint16(m[a])<<8 | uint16(m[a+1])) }

func main() {
	in := flag.String("in", "../extracted/game.dec.bin", "input decoded game binary")
	frames := flag.Int("frames", 120, "physics frames")
	drive := flag.Int("drive", 0, "drive force $1BD2A held each frame")
	noCrash := flag.Bool("nocrash", false, "zero $1BBDF after arming (skip spawn crash-recovery)")
	tracks := flag.String("tracks", "1,3,7", "comma track ids (unused placeholder)")
	flag.Parse()
	_ = tracks
	img, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fails := 0
	for _, id := range []int{1, 3, 7} {
		bus = &flatBus{m: make([]byte, 1<<24)}
		copy(bus.m[base:], img)
		c = m68k.NewCPU(bus)

		bus.Write(0x1CA33, byte(id))
		callD1(0x5AE46, uint32(id))
		call(0x64304)
		call(0x5A794)
		call(0x696FC)
		call(0x65BEC)
		bus.Write(0x1BB57, 1)
		copy(bus.m[0x64AEC:], []byte{0x9C, 0xED, 0xCD, 0x02})
		call(0x605B6) // real placement (posY = 16.0)
		call(0x5E778)
		call(0x60CDE)
		call(0x64E4C)
		call(0x6185C) // one unarmed tick
		call(0x64E4C)
		bus.Write(0x1BB72, 0x80) // arm
		if *noCrash {
			bus.Write(0x1BBDF, 0) // skip the spawn crash-recovery countdown ($5B32E machine)
		}

		// snapshot the armed placed state
		m0 := make([]byte, 1<<24)
		copy(m0, bus.m)

		gm := &physics.Mem{B: make([]byte, 1<<24)}
		copy(gm.B, m0)

		bad := 0
		for f := 0; f < *frames; f++ {
			// Real loop order ($5D48A then $5D496): physics tick, THEN the render
			// coupling that sets up the next tick. The intro's final $64E4C already
			// primed the coupling state read by this first physics tick.
			// oracle
			bus.Write(0x1BD2A, byte(*drive>>8))
			bus.Write(0x1BD2B, byte(*drive))
			call(0x6185C)
			engineCouple()
			// go
			gm.SetW(0x1BD2A, int16(*drive))
			gm.Frame6185C()
			goCouple(gm)

			for _, ck := range checks {
				if rw(gm.B, ck.addr) != rw(bus.m, ck.addr) {
					bad++
					if bad <= 6 {
						fmt.Printf("  t%d f%d MISMATCH %s @%X: go=%d eng=%d\n",
							id, f, ck.name, ck.addr, rw(gm.B, ck.addr), rw(bus.m, ck.addr))
					}
					break
				}
			}
		}
		if bad == 0 {
			// report the final grounded state as a sanity signal
			py := int32(uint32(rw(bus.m, 0x1BCDC))<<16 | uint32(uint16(rw(bus.m, 0x1BCDE))))
			fmt.Printf("Drive lockstep track %d (%d frames)  OK  (final posY=%d onG=%02x sec=%d)\n",
				id, *frames, py>>16, bus.m[0x1BB7E], bus.m[0x1BB1C])
		} else {
			fmt.Printf("Drive lockstep track %d  %d FAIL\n", id, bad)
			fails += bad
		}
	}
	if fails == 0 {
		fmt.Println("ALL OK")
	} else {
		os.Exit(1)
	}
}
