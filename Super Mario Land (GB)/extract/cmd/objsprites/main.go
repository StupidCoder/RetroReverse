// objsprites is a GUIDE harness that maps Super Mario Land object types to their sprite
// animation frame. The object sprite draw ($25B7) renders each slot's metasprite using
// the slot's frame field (slot+6, seen as $FFC6) as an index into the metasprite pointer
// table $2FD9. So an object's type (slot+0) and current frame (slot+6) are both right in
// the slot — we just play a level and tally, per type, the frames it actually uses.
//
// This only tells us WHICH frame id a type animates with; the frame's tiles and pixels are
// decoded from ROM ($2FD9 metasprite table + the OBJ tiles), not scraped here.
//
//	go run ./cmd/objsprites [-rom PATH] [-id NN] [-frames N]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	"retroreverse.com/tools/gameboy"
)

func main() {
	rom := flag.String("rom", "../Super Mario Land (World).gb", "ROM path")
	id := flag.String("id", "", "level id (hex); empty = default 1-1")
	frames := flag.Int("frames", 1800, "gameplay frames to run")
	flag.Parse()

	data, err := os.ReadFile(*rom)
	if err != nil {
		fmt.Fprintln(os.Stderr, "objsprites:", err)
		os.Exit(1)
	}
	m := gameboy.NewMachine(data)

	// type -> frame(slot+6) -> count
	typeFrame := map[byte]map[byte]int{}
	tally := func() {
		for s := 0; s < 10; s++ {
			base := uint16(0xD100 + s*0x10)
			typ := m.Read(base)
			if typ == 0xFF {
				continue
			}
			fr := m.Read(base + 6)
			if typeFrame[typ] == nil {
				typeFrame[typ] = map[byte]int{}
			}
			typeFrame[typ][fr]++
		}
	}

	m.RunFrames(80)
	for f := 0; f < 6; f++ {
		m.Buttons = gameboy.BtnStart
		m.RunFrame()
	}
	m.Buttons = 0

	// Visit every level (or one, if -id) and play it; the camera only scrolls while
	// Mario walks, and he dies on the first enemy, so each level mainly samples its
	// opening objects — visiting all 12 broadens the type coverage.
	var ids []byte
	if *id != "" {
		idv, _ := strconv.ParseUint(*id, 16, 8)
		ids = []byte{byte(idv)}
	} else {
		for w := 1; w <= 4; w++ {
			for l := 1; l <= 3; l++ {
				ids = append(ids, byte(w<<4|l))
			}
		}
	}
	for _, lid := range ids {
		for f := 0; f < 40; f++ {
			m.Write(0xFFB4, lid)
			m.RunFrame()
		}
		for f := 0; f < *frames; f++ {
			if f%160 < 10 {
				m.Buttons = gameboy.BtnRight | gameboy.BtnA
			} else {
				m.Buttons = gameboy.BtnRight
			}
			m.RunFrame()
			tally()
		}
		m.Buttons = 0
	}

	var types []int
	for t := range typeFrame {
		types = append(types, int(t))
	}
	sort.Ints(types)
	fmt.Println("type -> frames (slot+6), by frequency:")
	for _, t := range types {
		fm := typeFrame[byte(t)]
		var frs []int
		for fr := range fm {
			frs = append(frs, int(fr))
		}
		sort.Slice(frs, func(i, j int) bool { return fm[byte(frs[i])] > fm[byte(frs[j])] })
		fmt.Printf("  type $%02X: ", t)
		for i, fr := range frs {
			if i >= 8 {
				fmt.Print("…")
				break
			}
			fmt.Printf("$%02X(%d) ", fr, fm[byte(fr)])
		}
		fmt.Println()
	}

	// Emit a Go map literal of type -> most-common frame, to bake into the exporter.
	fmt.Println("\n// type -> representative metasprite frame (observed):")
	fmt.Print("var typeFrame = map[byte]byte{")
	for _, t := range types {
		fm := typeFrame[byte(t)]
		best, bn := byte(0), -1
		for fr, n := range fm {
			if n > bn {
				best, bn = fr, n
			}
		}
		fmt.Printf("0x%02X: 0x%02X, ", t, best)
	}
	fmt.Println("}")
}
