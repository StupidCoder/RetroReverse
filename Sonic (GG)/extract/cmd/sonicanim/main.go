// sonicanim captures Sonic's animation state frame by frame while he idles in Green
// Hills Act 1: the anim id (IX+20), the tile-stream source ($D289 = slot-base +
// graphicFrame*192, set by $4E9A) and the metasprite layout ($D40C). Idling long
// enough triggers the bored animation — this logs its exact frame sequence and
// timing for the static extractor to reproduce.
//
// Usage: sonicanim <rom.gg> [frames]
package main

import (
	"fmt"
	"os"
	"strconv"

	"retroreverse.com/tools/gamegear"
)

func main() {
	rom, _ := os.ReadFile(os.Args[1])
	n := 1500
	if len(os.Args) > 2 {
		n, _ = strconv.Atoi(os.Args[2])
	}
	m := gamegear.NewMachine(rom)
	word := func(a uint16) uint16 { return uint16(m.Read(a)) | uint16(m.Read(a+1))<<8 }
	for i := 0; i < 700; i++ {
		m.RunFrame()
	}
	for round := 0; round < 40 && word(0xD26F) == 0; round++ {
		m.Pad00 = 0x7F
		for i := 0; i < 8; i++ {
			m.RunFrame()
		}
		m.Pad00 = 0xFF
		for k := 0; k < 242 && word(0xD26F) == 0; k++ {
			m.RunFrame()
		}
	}
	last := ""
	lastF := 0
	for f := 0; f < n; f++ {
		m.RunFrame()
		anim := m.Read(0xD3FD + 20)
		src := word(0xD289)
		lay := word(0xD40C)
		s := fmt.Sprintf("anim=%02X src=%04X frame=%2d layout=%04X", anim, src, (int(src)-0x4000)/192, lay)
		if s != last {
			fmt.Printf("frame %4d (+%3d): %s\n", f, f-lastF, s)
			last, lastF = s, f
		}
	}
}
