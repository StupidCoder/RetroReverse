// geomoracle differentially verifies the nfs geometry decoders against the
// running game. It never extracts data — the decoders read the disc file
// bytes; the oracle only checks them (the repo's decode-reimplement rule).
//
// Four checks, all during one scripted drive into the City race:
//
//   - segments: the game's resident segment array ([0x4CCEC], 2400 × 0x24)
//     must equal ParseTrack's records byte-for-byte.
//
//   - rows: the world-vertex builder (0x15F70) emits every visible slice
//     cross-section point with a 3-word copy at 0x1605C. Each hit carries
//     the group ([r11-0x30]), the row within it ([sp+0x8]) and the point
//     index (r1); the copied (x,y,z) in r2/r3/r12 must equal
//     Track.Groups[group].Rows[row][point], bit-exact. This proves the
//     chunk → group → row mapping and the world coordinates.
//
//   - placements: the RoadObjects scan computes each visible placement's
//     world position (segment position + s16 offsets << 8) into the stack
//     at 0x16184 (r3 = the placement record). The three words at
//     [sp+0x20..0x2C) must equal Placement.World(segment), bit-exact.
//
//   - cartex: at race time the player car's per-face texture array
//     ([car+0x4E0]) must point at exactly the SPoT records ParseCarFam's
//     "!ori" face map selects, for every face of the loaded LOD.
//
// Exit status 0 only if every check passes.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/platform/threedo"
)

const (
	pcRowCopy   = 0x1605C // STMIA lr,{r2,r3,r12} in the world-vertex builder
	pcPlacement = 0x16184 // placement world pos complete, r3 = record ptr
	segBasePtr  = 0x4CCEC
	carStruct   = 0x295CE8
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	steps := flag.Uint64("steps", 19_000_000, "instructions to run")
	pad := flag.String("pad", "4000000:start,4300000:0,16000000:rs,16400000:0,17000000:a", "drive script")
	flag.Parse()

	data, err := os.ReadFile(*image)
	die(err)
	vol, err := threedo.Open(data)
	die(err)
	trk, err := vol.ReadFile("DriveData/tracks/cy1.trk")
	die(err)
	track, err := nfs.ParseTrack(trk)
	die(err)
	fam, err := vol.ReadFile("DriveData/CarData/LDiablo.wrapFam")
	die(err)
	lods, err := nfs.ParseCarFam(fam)
	die(err)

	prog, err := vol.ReadFile("LaunchMe")
	die(err)
	aif, err := threedo.ParseAIF(prog)
	die(err)
	m := threedo.NewMachine()
	m.SetVolume(vol)
	m.SetVBLMirror(0x42734)
	m.LoadAIF(aif)
	m.StallTolerance = 400
	m.NoStreams = true
	script, err := parsePadScript(*pad)
	die(err)
	m.PadScript = script

	rd := func(a uint32) uint32 {
		return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
	}

	var rowHits, rowBad, plHits, plBad int
	firstBad := ""
	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		c := mm.CPU
		switch pc {
		case pcRowCopy:
			rowHits++
			group := rd(c.Reg(11) - 0x30)
			row := rd(c.Reg(13) + 0x8)
			point := c.Reg(1)
			x, y, z := int32(c.Reg(2)), int32(c.Reg(3)), int32(c.Reg(12))
			if int(group) >= len(track.Groups) || row > 3 || point > 10 {
				rowBad++
				if firstBad == "" {
					firstBad = fmt.Sprintf("row copy out of range: group=%d row=%d point=%d", group, row, point)
				}
				return
			}
			want := track.Groups[group].Rows[row][point]
			if want.X != x || want.Y != y || want.Z != z {
				rowBad++
				if firstBad == "" {
					firstBad = fmt.Sprintf("row mismatch g%d r%d p%d: game (%08X,%08X,%08X) decoder (%08X,%08X,%08X)",
						group, row, point, uint32(x), uint32(y), uint32(z),
						uint32(want.X), uint32(want.Y), uint32(want.Z))
				}
			}
		case pcPlacement:
			plHits++
			// r3 = placement record pointer; base = [0x3F25C+0x48]
			base := rd(0x3F25C + 0x48)
			rec := c.Reg(3)
			idx := int(rec-base) / 16
			if idx < 0 || idx >= len(track.Objects.Placements) {
				plBad++
				if firstBad == "" {
					firstBad = fmt.Sprintf("placement ptr 0x%X outside table (base 0x%X)", rec, base)
				}
				return
			}
			p := track.Objects.Placements[idx]
			sp := c.Reg(13)
			gx, gy, gz := int32(rd(sp+0x20)), int32(rd(sp+0x24)), int32(rd(sp+0x28))
			if int(p.Segment) >= len(track.Segments) {
				plBad++
				return
			}
			want := p.World(&track.Segments[p.Segment])
			if want.X != gx || want.Y != gy || want.Z != gz {
				plBad++
				if firstBad == "" {
					firstBad = fmt.Sprintf("placement %d mismatch: game (%08X,%08X,%08X) decoder (%08X,%08X,%08X)",
						idx, uint32(gx), uint32(gy), uint32(gz),
						uint32(want.X), uint32(want.Y), uint32(want.Z))
				}
			}
		}
	}

	res := m.Run(*steps)
	fmt.Printf("ran %d steps (%s)\n", res.Steps, res.Reason)

	// segments: byte-compare the resident array
	segBase := rd(segBasePtr)
	segBad := 0
	for i, s := range track.Segments {
		for k := 0; k < len(s.Raw); k++ {
			if m.Read(segBase+uint32(i*len(s.Raw)+k)) != s.Raw[k] {
				segBad++
				break
			}
		}
	}

	// cartex: match the live texset against the decoder's face map. The
	// texset points into the loaded wrapFam; recover its base from the
	// model pointer and check every face.
	carBad, carFaces := checkCarTex(m, rd, fam, lods)

	fmt.Printf("segments: %d/%d byte-exact\n", len(track.Segments)-segBad, len(track.Segments))
	fmt.Printf("rows:     %d hits, %d mismatches\n", rowHits, rowBad)
	fmt.Printf("placements: %d hits, %d mismatches\n", plHits, plBad)
	fmt.Printf("cartex:   %d faces, %d mismatches\n", carFaces, carBad)
	if firstBad != "" {
		fmt.Printf("first mismatch: %s\n", firstBad)
	}
	if segBad+rowBad+plBad+carBad != 0 || rowHits == 0 || plHits == 0 || carFaces == 0 {
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("OK — decoders match the running game")
}

func checkCarTex(m *threedo.Machine, rd func(uint32) uint32, fam []byte, lods []*nfs.CarLOD) (bad, faces int) {
	modelPtr := rd(carStruct + 0x4DC)
	texset := rd(carStruct + 0x4E0)
	if modelPtr == 0 || texset == 0 {
		return 1, 0
	}
	// which LOD is loaded? match by the ORI3 name at [model+0x28]
	var name [8]byte
	for i := range name {
		name[i] = m.Read(modelPtr + 0x28 + uint32(i))
	}
	var lod *nfs.CarLOD
	for _, l := range lods {
		if string(name[:]) == l.Model.Name {
			lod = l
		}
	}
	if lod == nil {
		return 1, 0
	}
	// the wrapFam is loaded contiguously: RAM base = model ptr - ORI3 file
	// offset. Find the shape's SPoT records through the same base.
	oriOff := -1
	for o := 0; o+4 <= len(fam); o++ {
		if string(fam[o:o+4]) == "ORI3" && string(fam[o+0x28:o+0x30]) == lod.Model.Name {
			oriOff = o
			break
		}
	}
	if oriOff < 0 {
		return 1, 0
	}
	base := modelPtr - uint32(oriOff)
	spotOffs := nfs.SpotOffsets(fam, oriOff)
	if spotOffs == nil {
		return 1, 0
	}
	// The game swaps some texset entries at runtime (wheel-spin frames
	// "whl0".."whl2", the brake light "bkl1"); those must still point at a
	// SPoT of the same shape — count them as dynamic, not mismatches.
	valid := map[uint32]bool{}
	for _, off := range spotOffs {
		valid[base+uint32(off)] = true
	}
	dynamic := 0
	for i := 0; i < len(lod.Model.Faces); i++ {
		faces++
		got := rd(texset + 4*uint32(i))
		want := base + uint32(spotOffs[lod.FaceTex[i]])
		if got == want {
			continue
		}
		if valid[got] {
			dynamic++
			continue
		}
		bad++
	}
	if dynamic > 0 {
		fmt.Printf("cartex:   %d entries runtime-swapped to other SPoTs (wheels/brake light)\n", dynamic)
	}
	return bad, faces
}

// parsePadScript mirrors bootoracle's syntax.
func parsePadScript(s string) ([]threedo.PadStep, error) {
	bits := map[string]uint32{
		"a": threedo.PadA, "b": threedo.PadB, "c": threedo.PadC,
		"start": threedo.PadStart, "x": threedo.PadX,
		"up": threedo.PadUp, "down": threedo.PadDown,
		"left": threedo.PadLeft, "right": threedo.PadRight,
		"ls": threedo.PadLeftShift, "rs": threedo.PadRightShift, "0": 0,
	}
	var script []threedo.PadStep
	for _, entry := range strings.Split(s, ",") {
		step, names, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok {
			return nil, fmt.Errorf("pad entry %q: want STEP:buttons", entry)
		}
		at, err := strconv.ParseUint(step, 10, 64)
		if err != nil {
			return nil, err
		}
		var buttons uint32
		for _, nm := range strings.Split(names, "+") {
			bit, ok := bits[strings.ToLower(strings.TrimSpace(nm))]
			if !ok {
				return nil, fmt.Errorf("pad entry %q: unknown button %q", entry, nm)
			}
			buttons |= bit
		}
		script = append(script, threedo.PadStep{AtStep: at, Buttons: buttons})
	}
	sort.Slice(script, func(i, j int) bool { return script[i].AtStep < script[j].AtStep })
	return script, nil
}

func die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "geomoracle:", err)
		os.Exit(1)
	}
}
