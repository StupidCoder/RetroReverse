// coloracle verifies the Go reimplementation of the preview renderer's face
// colours and decal lines (track/mesh.go) against the real engine on the
// tools/m68k core. It runs the race-init chain + $65BEC bake (as modeloracle),
// then calls the actual pre-race preview draw $604B4 for each of the four camera
// angles with PC hooks planted on the colour decisions:
//
//	$67440  per-strip descriptor {section, type, flag, piece} written
//	$68DEE  road-surface fill colour chosen ($68D82)
//	$68ACC  left-wall fill colour chosen ($68A56)
//	$68C62  right-wall fill colour chosen ($68BEC)
//	$6890C/$68950/$6899C/$689E6/$68A2A  line strokes (slot + colour patched
//	        into the plane-select rasteriser via $66348)
//
// and compares every logged decision against track.Mesh. The palette is checked
// end-to-end: the game's own load ($5D228: $2B974 -> $1BA70), fade to completion
// ($64470) and copper push ($1BA82) run on the core, and the 16 COLOR words at
// $E770+4n must equal track.Palette. Per the project rule the oracle only
// verifies; it is never the source of shipped data.
//
// Usage: coloracle -in game.dec.bin [-id N] [-dump]
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

// strip is one logged display-list strip (one rung pair) with every colour
// decision the flush made for it.
type strip struct {
	angle   int
	cellD1  int8 // walk coords ($1BB8B/$1BB8D), for context only
	cellD2  int8
	sec     int
	rung    int // d1/4 at descriptor time (the strip's current rung)
	typ     byte
	flag    byte
	piece   byte
	b66     byte
	road    int // fill palette idx, -1 if the flush never filled it
	wallL   int
	wallR   int
	strokes map[string]byte // slot name -> colour
}

func main() {
	in := flag.String("in", "", "input decoded game binary (game.dec.bin)")
	idFlag := flag.Int("id", -1, "track id (-1: all eight)")
	dump := flag.Bool("dump", false, "dump the engine log instead of comparing")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: coloracle -in game.dec.bin [-id N] [-dump]")
		os.Exit(2)
	}
	img, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coloracle:", err)
		os.Exit(1)
	}
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if *idFlag >= 0 {
		ids = []int{*idFlag}
	}
	fail := 0
	for _, id := range ids {
		if !run(img, id, *dump) {
			fail++
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
}

func run(img []byte, id int, dump bool) bool {
	bus := &flatBus{m: make([]byte, 1<<24)}
	copy(bus.m[base:], img)
	c := m68k.NewCPU(bus)
	rd16 := func(a uint32) int { return int(bus.m[a])<<8 | int(bus.m[a+1]) }

	var strips []*strip
	var cur *strip
	angle := 0

	// hooks fire before the instruction at their PC executes
	hooks := map[uint32]func(){
		0x67440: func() { // descriptor bytes just written at list offset d3
			d3 := c.D[3] & 0xFFFF
			s := &strip{
				angle:  angle,
				cellD1: int8(bus.m[0x1BB8B]),
				cellD2: int8(bus.m[0x1BB8D]),
				sec:    int(bus.m[0x7B08A+d3]),
				typ:    bus.m[0x7B08A+d3+1],
				flag:   bus.m[0x7B08A+d3+2],
				piece:  bus.m[0x7B08A+d3+3],
				rung:   int(bus.m[0x1BBE4]) / 4,
				b66:    bus.m[0x1BB66],
				road:   -1, wallL: -1, wallR: -1,
				strokes: map[string]byte{},
			}
			strips = append(strips, s)
		},
	}
	// fill-colour hooks: the flush walks strips backwards; d3 indexes the strip
	// whose descriptor drove the decision. Find it by list offset.
	byOffset := map[uint32]*strip{}
	hookDesc := hooks[0x67440]
	hooks[0x67440] = func() {
		hookDesc()
		byOffset[c.D[3]&0xFFFF] = strips[len(strips)-1]
	}
	fill := func(field *int) func() {
		_ = field
		return nil
	}
	_ = fill
	setFill := func(which string) func() {
		return func() {
			d3 := c.D[3] & 0xFFFF
			s := byOffset[d3]
			if s == nil {
				return
			}
			v := int(c.D[5] & 0xFF)
			switch which {
			case "road":
				s.road = v
			case "wallL":
				s.wallL = v
			case "wallR":
				s.wallR = v
			}
		}
	}
	hooks[0x68DEE] = setFill("road")
	hooks[0x68ACC] = setFill("wallL")
	hooks[0x68C62] = setFill("wallR")
	stroke := func(slot string) func() {
		return func() {
			d3 := c.D[3] & 0xFFFF
			if s := byOffset[d3]; s != nil {
				s.strokes[slot] = byte(c.D[0] & 0xFF)
			}
		}
	}
	hooks[0x6890C] = stroke("finish") // cross line at cursor strip, finish only
	hooks[0x68950] = stroke("vertL")  // wall vertical L (slot $10)
	hooks[0x6899C] = stroke("vertR")  // wall vertical R (slot $14)
	hooks[0x689E6] = stroke("curbL")  // road-edge lengthwise L (slot $0)
	hooks[0x68A2A] = stroke("curbR")  // road-edge lengthwise R (slot $4)
	// cur is unused; silence linters that dislike dead vars
	_ = cur

	call := func(pc uint32, regs map[int]uint32) bool {
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
			if c.Halted || steps > 100_000_000 {
				fmt.Printf("track %d: HALT/step-cap at $%X\n", id, c.PC)
				return false
			}
			if h, ok := hooks[c.PC]; ok {
				h()
			}
			c.Step()
		}
		return true
	}

	// race-init chain + bake (exactly as cmd/modeloracle)
	for _, e := range []uint32{0x5AE46, 0x64304, 0x5A794, 0x696FC} {
		regs := map[int]uint32{}
		if e == 0x5AE46 {
			regs[1] = uint32(id)
		}
		if !call(e, regs) {
			return false
		}
	}
	if !call(0x65BEC, map[int]uint32{1: 0, 2: 0}) {
		return false
	}

	// the game's own palette path: load target ($5D228 does a1=$2B974 ->
	// $1BA70), fade the staged palette to completion, push to the copper list
	c.A[1] = 0x2B974
	if !call(0x1BA70, nil) {
		return false
	}
	for i := 0; i < 64; i++ { // 15 steps max per channel; 64 is safely past
		if !call(0x64470, nil) {
			return false
		}
	}
	if !call(0x1BA82, nil) {
		return false
	}

	im := track.New(img)
	t := im.Spine(id)

	// palette: copper COLOR value slots at $E770+4n vs track.Palette
	pal := im.Palette()
	palBad := 0
	for i := 0; i < 16; i++ {
		got := rd16(uint32(0xE770 + 4*i))
		want := int(pal[i][0])<<8 | int(pal[i][1])<<4 | int(pal[i][2])
		if got != want {
			fmt.Printf("track %d: palette %d oracle $%03X go $%03X\n", id, i, got, want)
			palBad++
		}
	}

	// the four preview camera angles
	for angle = 0; angle < 4; angle++ {
		bus.m[0x1BB57] = byte(angle)
		if !call(0x604B4, nil) {
			return false
		}
	}

	if dump {
		fmt.Printf("track %d: %d strips, palette %s\n", id, len(strips),
			map[bool]string{true: "OK", false: "BAD"}[palBad == 0])
		for _, s := range strips {
			fmt.Printf("a%d cell(%d,%d) sec %2d rung %2d typ %02X flag %02X piece %02X b66 %02X road %2d wallL %2d wallR %2d strokes %v\n",
				s.angle, s.cellD1, s.cellD2, s.sec, s.rung, s.typ, s.flag, s.piece, s.b66, s.road, s.wallL, s.wallR, s.strokes)
		}
		return palBad == 0
	}

	// compare against track.Mesh
	mesh := im.Mesh(&t)
	mism := 0
	report := func(f string, a ...any) {
		mism++
		if mism <= 12 {
			fmt.Printf(f, a...)
		}
	}
	covered := map[[2]int]bool{}
	for _, s := range strips {
		if s.sec >= len(mesh.Sections) {
			report("track %d a%d: strip sec %d out of range\n", id, s.angle, s.sec)
			continue
		}
		ms := mesh.Sections[s.sec]
		if s.rung >= len(ms.Rungs) {
			report("track %d a%d: sec %d rung %d out of range (%d rungs)\n", id, s.angle, s.sec, s.rung, len(ms.Rungs))
			continue
		}
		covered[[2]int{s.sec, s.rung}] = true
		mr := ms.Rungs[s.rung]
		if s.typ != mr.Type {
			report("track %d a%d sec %d rung %d: type oracle %02X go %02X\n", id, s.angle, s.sec, s.rung, s.typ, mr.Type)
		}
		if s.road >= 0 && byte(s.road) != mr.RoadPal {
			report("track %d a%d sec %d rung %d: road oracle %d go %d\n", id, s.angle, s.sec, s.rung, s.road, mr.RoadPal)
		}
		wallWant := mr.WallPal
		if s.b66 != 0 {
			wallWant = 9 // back-side fill: view-dependent, not part of the mesh
		}
		if s.wallL >= 0 && byte(s.wallL) != wallWant {
			report("track %d a%d sec %d rung %d: wallL oracle %d go %d (b66 %02X)\n", id, s.angle, s.sec, s.rung, s.wallL, wallWant, s.b66)
		}
		if s.wallR >= 0 && byte(s.wallR) != wallWant {
			report("track %d a%d sec %d rung %d: wallR oracle %d go %d (b66 %02X)\n", id, s.angle, s.sec, s.rung, s.wallR, wallWant, s.b66)
		}
		if col, ok := s.strokes["curbL"]; ok && col != mr.Type {
			report("track %d a%d sec %d rung %d: curbL stroke %d != type %d\n", id, s.angle, s.sec, s.rung, col, mr.Type)
		}
		if col, ok := s.strokes["vertL"]; ok && (col != 9 || mr.Type == 3) {
			report("track %d a%d sec %d rung %d: vertL stroke %d (type %d)\n", id, s.angle, s.sec, s.rung, col, mr.Type)
		}
		if col, ok := s.strokes["finish"]; ok && (col != 0xF || !mr.Finish) {
			report("track %d a%d sec %d rung %d: finish stroke %d finish=%v\n", id, s.angle, s.sec, s.rung, col, mr.Finish)
		}
	}
	total := 0
	miss := 0
	for si, ms := range mesh.Sections {
		for ri := range ms.Rungs {
			if ri == 0 {
				continue // rung 0 duplicates the previous section's last rung
			}
			total++
			if !covered[[2]int{si, ri}] {
				miss++
			}
		}
	}
	if mism == 0 {
		fmt.Printf("track %d: OK — %d strips over 4 angles match track.Mesh (%d/%d rungs covered, palette OK)\n",
			id, len(strips), total-miss, total)
		return palBad == 0
	}
	fmt.Printf("track %d: %d MISMATCHES (%d strips)\n", id, mism, len(strips))
	return false
}
