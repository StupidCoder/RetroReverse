package main

import (
	"fmt"
	"math"

	"stupidcoder.com/tools/gameboy"
)

var noteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func freqName(f int) string {
	if f <= 0 || f >= 2048 {
		return "--"
	}
	hz := 131072.0 / float64(2048-f)
	n := int(math.Round(12*math.Log2(hz/440))) + 69
	if n < 0 {
		return "?"
	}
	return fmt.Sprintf("%s%d", noteNames[n%12], n/12-1)
}

// verifyMelody decodes channel 2 (the usual melody) of `id` from the data and prints it next
// to the note stream the real engine plays, so the port can be checked against ground truth.
func verifyMelody(id int) {
	ch := song(id)[1] // square 2
	fmt.Printf("=== music id $%02X — decoded ch2 (square 2), loop %d ticks ===\n", id, ch.loopTicks)
	prev := -1
	line := ""
	for i, e := range ch.events {
		if i >= 32 {
			line += "…"
			break
		}
		if e.tie {
			line += "~ "
			continue
		}
		line += freqName(e.freq) + " "
		prev = e.freq
	}
	_ = prev
	fmt.Println(" decoded:", line)

	// engine ground truth
	m := gameboy.NewMachine(rom)
	m.RunFrames(120)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0
	m.RunFrames(120)
	for f := 0; f < 4; f++ {
		m.Write(0xDFE8, byte(id))
		m.RunFrame()
	}
	var f2, cnt int
	eng := ""
	m.OnWrite = func(pc, addr uint16, v byte) {
		if addr == 0xFF18 {
			f2 = (f2 & 0x700) | int(v)
		}
		if addr == 0xFF19 && v&0x80 != 0 && cnt < 32 {
			f2 = (f2 & 0xFF) | int(v&7)<<8
			eng += freqName(f2) + " "
			cnt++
		}
	}
	m.RunFrames(360)
	fmt.Println(" engine: ", eng)
}
