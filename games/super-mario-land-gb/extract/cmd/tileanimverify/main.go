// tileanimverify oracle-checks the decoded BG tile animation (level.DecodeTileAnim):
// it boots a level, samples the animated tile's VRAM data ($95D0-$95DF, tile $5D)
// every frame, and asserts the observed patterns and phase period match the decode.
//
//	go run ./cmd/tileanimverify [-rom PATH] [-id NN]
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/tools/platform/gameboy"
	"retroreverse.com/games/super-mario-land-gb/extract/level"
)

func main() {
	romPath := flag.String("rom", "../Super Mario Land (World).gb", "ROM path")
	id := flag.String("id", "13", "level id (hex)")
	flag.Parse()

	data, err := os.ReadFile(*romPath)
	must(err)
	idv, _ := strconv.ParseUint(*id, 16, 8)
	want := level.DecodeTileAnim(data, byte(idv))

	m := gameboy.NewMachine(data)
	m.RunFrames(80)
	ffe4 := byte((idv>>4-1)*3 + (idv & 0x0F) - 1)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.Write(0xFFB4, byte(idv))
		m.Write(0xFFE4, ffe4)
		m.RunFrame()
	}
	m.Buttons = 0
	for f := 0; f < 120; f++ {
		m.Write(0xFFB4, byte(idv))
		m.Write(0xFFE4, ffe4)
		m.RunFrame()
	}

	// Sample tile $5D's 16 VRAM bytes for 64 gameplay frames.
	const tileAddr = 0x9000 + level.AnimTile*16 - 0x8000 // offset into VRAM()
	seen := map[string]int{}
	var order []string
	var phases []int
	last := ""
	for f := 0; f < 64; f++ {
		m.RunFrame()
		cur := string(m.VRAM()[tileAddr : tileAddr+16])
		if cur != last {
			if last != "" {
				phases = append(phases, f)
			}
			last = cur
			if seen[cur] == 0 {
				order = append(order, cur)
			}
		}
		seen[cur]++
	}

	if want == nil {
		if len(seen) <= 1 {
			fmt.Printf("level %x-%x: decode says static, oracle saw %d pattern(s) - OK\n", idv>>4, idv&0xf, len(seen))
			return
		}
		fmt.Printf("level %x-%x: decode says static but oracle saw %d patterns - MISMATCH\n", idv>>4, idv&0xf, len(seen))
		os.Exit(1)
	}
	ok := len(order) == 2
	for _, o := range order {
		if !bytes.Equal([]byte(o), want.Frames[0][:]) && !bytes.Equal([]byte(o), want.Frames[1][:]) {
			ok = false
		}
	}
	// Phase length: distance between changes should be AnimPeriod.
	period := 0
	if len(phases) >= 2 {
		period = phases[1] - phases[0]
	}
	if period != level.AnimPeriod {
		ok = false
	}
	status := "OK"
	if !ok {
		status = "MISMATCH"
	}
	fmt.Printf("level %x-%x: %d patterns (want 2, both matching decode), phase period %d frames (want %d) - %s\n",
		idv>>4, idv&0xf, len(order), period, level.AnimPeriod, status)
	for i, o := range order {
		which := "??"
		if bytes.Equal([]byte(o), want.Frames[0][:]) {
			which = "accent (decoded frame 0)"
		} else if bytes.Equal([]byte(o), want.Frames[1][:]) {
			which = "resting (decoded frame 1)"
		}
		fmt.Printf("  pattern %d: % x = %s\n", i, o, which)
	}
	if !ok {
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "tileanimverify:", err)
		os.Exit(1)
	}
}
