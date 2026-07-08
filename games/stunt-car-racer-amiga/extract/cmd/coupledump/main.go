// coupledump builds memory exactly the way the browser's Physics.loadTrack does (per-track
// .bin + static.bin + the baked code constants), sets a fixed test car state, runs the Go
// render-coupling chain (Camera60190 -> Section5FE04 -> Couple5BE44), and prints the coupling
// outputs as JSON. A Node harness runs the JS port over the same .bin/state and compares, so
// the JS coupling is checked against the (oracle-verified) Go.
//
// Usage: coupledump phys-dir   (containing 1.bin, static.bin)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/stunt-car-racer-amiga/extract/physics"
)

// loadTrack mirrors site/src/stuntcar/physics.js loadTrack.
func loadTrack(perTrack, static []byte) *physics.Mem {
	b := make([]byte, 1<<24)
	copy(b[0x1C900:], static)
	split := 0x1C900 - 0x1B000
	copy(b[0x1B000:], perTrack[:split])
	copy(b[0x1CA1A:], perTrack[split:])
	copy(b[0x6125A:], []byte{0, 0, 0, 217, 255, 39})
	copy(b[0x61AD4:], []byte{44, 0, 10, 0, 211, 0, 245, 0, 48, 57, 0, 1})
	copy(b[0x64AEC:], []byte{0x9C, 0xED, 0xCD, 0x02})
	copy(b[0x5C6B8:], []byte{0x00, 0xD4, 0x80, 0xD4, 0x00, 0x00, 0xAB, 0xAB, 0x40, 0x40, 0x00, 0x00})
	return &physics.Mem{B: b}
}

// xs steps the xorshift32 the Node harness mirrors, so both sides generate the same
// camera-state bytes for the per-section coupling sweep.
func xs(s uint32) uint32 {
	s ^= s << 13
	s ^= s >> 17
	s ^= s << 5
	return s
}

func main() {
	dir := os.Args[1]
	perTrack, err := os.ReadFile(filepath.Join(dir, "1.bin"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	static, err := os.ReadFile(filepath.Join(dir, "static.bin"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	type out struct {
		BC30, BB04, BB06, BB22, BB26       int
		BBF2, BBF6, BBF8, BC5E, BB10, BD5A int
		Sec                                int
	}
	var results []out
	// a few test car states (posX/Y/Z 32-bit, roll/yaw/pit, netlift) exercising the quadrants.
	type st struct{ px, py, pz, roll, yaw, pit, lift int }
	states := []st{
		{0x01C00000, 0x00100000, 0x07AC0000, 0, 0x8000, 0, 0},
		{0x02400000, 0x00100000, 0x03000000, 0x0040, 0x0000, 0x0010, 0x40},
		{0x05000000, 0x00080000, 0x06000000, -0x0080, 0x4000, -0x0020, 0x600},
		{0x00400000, 0x00200000, 0x00400000, 0x0100, 0xC000, 0x0040, 0x1200},
	}
	for _, s := range states {
		m := loadTrack(perTrack, static)
		m.SetL(0x1BCD8, int32(s.px))
		m.SetL(0x1BCDC, int32(s.py))
		m.SetL(0x1BCE0, int32(s.pz))
		m.SetW(0x1BCE4, int16(s.roll))
		m.SetW(0x1BCE6, int16(s.yaw))
		m.SetW(0x1BCE8, int16(s.pit))
		m.SetW(0x1BD38, int16(s.lift))
		m.B[0x1BBD5], m.B[0x1BBD6] = 0, 0
		m.Camera60190()
		sec, _ := m.Section5FE04()
		m.B[0x1BB1C], m.B[0x1BB85] = byte(sec), byte(sec)
		m.Couple5BE44()
		results = append(results, out{
			int(m.W(0x1BC30)), int(m.U8(0x1BB04)), int(m.U8(0x1BB06)), int(m.U8(0x1BB22)), int(m.U8(0x1BB26)),
			int(m.W(0x1BBF2)), int(m.W(0x1BBF6)), int(m.W(0x1BBF8)), int(m.W(0x1BC5E)), int(m.W(0x1BB10)), int(m.W(0x1BD5A)),
			sec,
		})
	}
	// Per-section coupling sweep over the tracks with ramp-type-2 pieces: deterministic
	// xorshift camera states (mirrored by the Node harness), Couple5BE44 run directly on
	// the placed section, all branch outputs dumped -- so the JS ramp-2 port is checked
	// against the oracle-verified Go on every section, every branch.
	type sweepOut struct {
		Track, Sec, Iter                   int
		BBF2, BBF6, BBF8, BC5E, BB10, BD5A int
		BC1C, BC36                         int
		Sec85, BB4D                        int
	}
	var sweep []sweepOut
	for _, id := range []int{1, 3, 7} {
		pt, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%d.bin", id)))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		probe := loadTrack(pt, static)
		n := probe.U8(0x1CA1A)
		for sec := 0; sec < n; sec++ {
			for iter := 0; iter < 8; iter++ {
				m := loadTrack(pt, static)
				m.B[0x1BB85] = byte(sec)
				s := uint32(id*1_000_000 + sec*1_000 + iter + 1)
				for _, a := range []uint32{0x1BB04, 0x1BB06, 0x1BC30, 0x1BB2E, 0x1BB32, 0x1BC4A, 0x1BC4B} {
					s = xs(s)
					m.B[a] = byte(s)
				}
				m.Couple5BE44()
				sweep = append(sweep, sweepOut{
					id, sec, iter,
					int(m.W(0x1BBF2)), int(m.W(0x1BBF6)), int(m.W(0x1BBF8)),
					int(m.W(0x1BC5E)), int(m.W(0x1BB10)), int(m.W(0x1BD5A)),
					int(m.W(0x1BC1C)), int(m.W(0x1BC36)),
					m.U8(0x1BB85), m.U8(0x1BB4D),
				})
			}
		}
	}

	b, _ := json.MarshalIndent(map[string]any{"states": results, "sweep": sweep}, "", " ")
	fmt.Println(string(b))
}
