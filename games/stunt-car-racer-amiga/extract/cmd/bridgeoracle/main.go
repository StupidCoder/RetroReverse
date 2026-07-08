// bridgeoracle verifies the Go reimplementation of the Draw Bridge animator
// (track.Drawbridge) against the engine's real $5A794 on the tools/m68k core:
// it loads track 5, then steps the animator through two full cycles (64
// frames), comparing the whole 36-entry bridge profile table byte-exact after
// every call, and asserts the freeze gate (a car on sections $33..$37 holds
// the phase). Per the project rule the oracle only verifies; track.Drawbridge
// patches purely from the image.
//
// Usage: bridgeoracle -in game.dec.bin
package main

import (
	"bytes"
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
	tblAddr  = 0x1F785 // handle $5F: the 36-entry bridge profile table
	tblLen   = 72
)

type flatBus struct{ m []byte }

func (b *flatBus) Read(a uint32) byte     { return b.m[a&0xFFFFFF] }
func (b *flatBus) Write(a uint32, v byte) { b.m[a&0xFFFFFF] = v }

func main() {
	in := flag.String("in", "", "input decoded game binary (game.dec.bin)")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: bridgeoracle -in game.dec.bin")
		os.Exit(2)
	}
	img, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bridgeoracle:", err)
		os.Exit(1)
	}

	bus := &flatBus{m: make([]byte, 1<<24)}
	copy(bus.m[base:], img)
	c := m68k.NewCPU(bus)
	call := func(pc uint32, regs map[int]uint32) {
		c.A[7] = stackTop - 4
		for i := uint32(0); i < 4; i++ {
			bus.Write(c.A[7]+i, byte(uint32(sentinel)>>(24-8*i)))
		}
		for r, v := range regs {
			c.D[r] = v
		}
		c.PC = pc
		for steps := 0; c.PC != sentinel; steps++ {
			if c.Halted || steps > 50_000_000 {
				fmt.Printf("HALT/cap at $%X\n", c.PC)
				os.Exit(1)
			}
			c.Step()
		}
	}

	bus.m[0x1CA33] = 5 // the animator's own gate ($5A79A) reads the menu byte
	call(0x5AE46, map[int]uint32{1: 5})
	call(0x64304, nil)

	im := track.New(img)
	mism := 0
	for frame := 0; frame < 64; frame++ {
		call(0x5A794, nil)
		phase := int(bus.m[0x1BBB0])
		want := im.Drawbridge(phase)
		got := bus.m[tblAddr : tblAddr+tblLen]
		exp := make([]byte, tblLen)
		for i := 0; i < tblLen; i++ {
			exp[i] = byte(want.U8(tblAddr + i))
		}
		if !bytes.Equal(got, exp) {
			fmt.Printf("frame %d (phase %d, tri %d): TABLE MISMATCH\n  oracle % X\n  go     % X\n",
				frame, phase, track.DrawbridgeTri(phase), got, exp)
			mism++
			if mism > 3 {
				os.Exit(1)
			}
		}
	}

	// freeze gate: a car on the bridge sections holds the phase
	before := bus.m[0x1BBB0]
	bus.m[0x1BB1C] = 0x35
	call(0x5A794, nil)
	frozen := bus.m[0x1BBB0] == before
	bus.m[0x1BB1C] = 0
	if !frozen {
		fmt.Println("freeze gate FAILED: phase advanced with a car on the bridge")
		mism++
	}

	if mism == 0 {
		fmt.Println("OK — track.Drawbridge matches $5A794 over 64 frames (byte-exact), freeze gate holds")
		return
	}
	os.Exit(1)
}
