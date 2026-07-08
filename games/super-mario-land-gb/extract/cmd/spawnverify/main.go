// spawnverify oracle-checks the placement-list -> world-X mapping: it hooks the
// spawner's own X write (LDH ($FFC3),A at $24B9) and Y write ($24AF), reads the
// camera at that instant, and compares the resulting world position with the
// formula decoded from the ROM (x = col*16 + fine*4, screen row = pos&$1F).
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/tools/platform/gameboy"
	"retroreverse.com/games/super-mario-land-gb/extract/level"
)

func main() {
	romPath := flag.String("rom", "../Super Mario Land (World).gb", "ROM path")
	id := flag.String("id", "11", "level id (hex)")
	frames := flag.Int("frames", 1800, "gameplay frames")
	jumpEvery := flag.Int("jump", 24, "press A every N frames (0=never)")
	jumpHold := flag.Int("hold", 12, "hold A for N frames of each jump period")
	flag.Parse()

	data, err := os.ReadFile(*romPath)
	must(err)
	idv, _ := strconv.ParseUint(*id, 16, 8)
	objs := level.DecodeObjectsByID(data, byte(idv))

	m := gameboy.NewMachine(data)

	camWraps := 0
	lastCam := byte(0)
	idx := 0
	tracing := false
	sampleCam := func() int {
		cur := m.Read(0xFFA4)
		if cur < lastCam && lastCam-cur > 128 {
			camWraps++
		}
		lastCam = cur
		return camWraps*256 + int(cur)
	}
	var pendY byte
	m.OnWrite = func(pc, addr uint16, v byte) {
		if !tracing {
			return
		}
		switch {
		case addr == 0xD010 && pc == 0x2474:
			// objlist_init ($2453) ran: level (re)start — stop, only the first
			// life has trustworthy camera tracking.
			fmt.Printf("-- list re-init (death/restart): stopping trace\n")
			tracing = false
		case addr == 0xFFC2 && pc == 0x24B1:
			pendY = v
		case addr == 0xFFC3 && pc == 0x24BB:
			cam := sampleCam()
			worldX := cam + int(v) - 8
			scrRow := (int(pendY) - 16) / 8
			if idx >= len(objs) {
				fmt.Printf("?? spawn beyond decoded list (idx %d)\n", idx)
				return
			}
			o := objs[idx]
			wantX := o.Col*8 + o.FineX*4
			wantRow := o.Row + 2
			mark := "OK"
			if worldX != wantX || scrRow != wantRow {
				mark = "MISMATCH"
			}
			fmt.Printf("entry %2d type=$%02X hard=%-5v  oracle worldX=%4d row=%2d | list col*16+fine*4=%4d row=%2d  dx=%+d  %s\n",
				idx, o.Type, o.Hard, worldX, scrRow, wantX, wantRow, worldX-wantX, mark)
			idx++
		}
	}

	// Boot to title, press Start, forcing both the level id ($FFB4) and the global
	// level index ($FFE4) through the whole boot + load so every loader stage sees
	// the target level.
	ffe4 := byte((idv>>4-1)*3 + (idv & 0x0F) - 1)
	m.RunFrames(80)
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
	fmt.Printf("after warp: FFB4=%02X FFE4=%02X\n", m.Read(0xFFB4), m.Read(0xFFE4))

	tracing = true
	for f := 0; f < *frames; f++ {
		// Keep the target level forced so death-restarts reload the same level.
		m.Write(0xFFB4, byte(idv))
		m.Buttons = gameboy.BtnRight
		if *jumpEvery > 0 && f%*jumpEvery < *jumpHold {
			m.Buttons |= gameboy.BtnA
		}
		m.RunFrame()
		sampleCam()
	}
	fmt.Println("done.")
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "spawnverify:", err)
		os.Exit(1)
	}
}
