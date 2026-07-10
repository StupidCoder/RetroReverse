package upw

import (
	"os"
	"strings"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/games/pilotwings-64-n64/extract/uvtr"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

func archive(t *testing.T) *pwad.Archive {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func chunkList(t *testing.T, a *pwad.Archive, f pwad.Form) []Chunk {
	t.Helper()
	var out []Chunk
	for _, c := range f.Chunks {
		data, err := a.Data(c)
		if err != nil {
			t.Fatal(err)
		}
		tag := c.Tag
		if c.Compressed() {
			tag = c.InnerTag
		}
		out = append(out, Chunk{Tag: tag, Data: data})
	}
	return out
}

func missions(t *testing.T, a *pwad.Archive) []*Mission {
	t.Helper()
	var out []*Mission
	for _, i := range a.ByType("UPWT") {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		m, err := DecodeMission(chunkList(t, a, f))
		if err != nil {
			t.Fatalf("UPWT %d: %v", i, err)
		}
		out = append(out, m)
	}
	return out
}

func TestSixtyOneMissionsDecode(t *testing.T) {
	a := archive(t)
	ms := missions(t, a)
	if got, want := len(ms), 61; got != want {
		t.Fatalf("%d missions, want %d", got, want)
	}
	for _, m := range ms {
		if m.Name == "" {
			t.Error("mission with an empty NAME")
		}
		if m.Level > 3 {
			t.Errorf("%q names level %d", m.Name, m.Level)
		}
	}
}

// The COMM descriptor's second byte is the vehicle, and the developers named
// every mission after the craft it is flown with. That the two agree on all 42
// prefixed missions is what licenses reading that byte as the vehicle at all.
func TestVehicleByteAgreesWithTheMissionName(t *testing.T) {
	a := archive(t)
	prefix := map[string]Vehicle{
		"HG": HangGlider, "RP": RocketBelt, "GC": Gyrocopter,
		"CB": Cannonball, "SD": SkyDiving, "HM": HumanTorch,
	}
	checked := 0
	for _, m := range missions(t, a) {
		for p, v := range prefix {
			if strings.HasPrefix(m.Name, p) {
				checked++
				if m.Vehicle != v {
					t.Errorf("%q: name says %s, COMM says %s", m.Name, v, m.Vehicle)
				}
			}
		}
	}
	if checked != 42 {
		t.Errorf("checked %d prefixed missions, want 42", checked)
	}
}

// A takeoff pad is three floats read at an offset nothing forced on us. If the
// offset were wrong they would be denormals or 1e38, not points inside a world.
func TestEveryTakeoffPadLiesInsideAWorld(t *testing.T) {
	a := archive(t)
	worlds := decodeWorlds(t, a)
	for _, m := range missions(t, a) {
		ok := false
		for _, w := range worlds {
			if m.Takeoff.X >= w.Min[0] && m.Takeoff.X <= w.Max[0] &&
				m.Takeoff.Y >= w.Min[1] && m.Takeoff.Y <= w.Max[1] {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("%q: takeoff pad (%g,%g) lies inside no world", m.Name, m.Takeoff.X, m.Takeoff.Y)
		}
	}
}

// All of a level's missions must be able to share a world: the set of worlds
// containing every one of their takeoff pads may not be empty.
func TestEachLevelsMissionsShareAWorld(t *testing.T) {
	a := archive(t)
	worlds := decodeWorlds(t, a)
	byLevel := map[uint8][]*Mission{}
	for _, m := range missions(t, a) {
		byLevel[m.Level] = append(byLevel[m.Level], m)
	}
	if len(byLevel) != 4 {
		t.Fatalf("missions name %d levels, want 4", len(byLevel))
	}
	for l, ms := range byLevel {
		common := 0
		for _, w := range worlds {
			all := true
			for _, m := range ms {
				if m.Takeoff.X < w.Min[0] || m.Takeoff.X > w.Max[0] ||
					m.Takeoff.Y < w.Min[1] || m.Takeoff.Y > w.Max[1] {
					all = false
					break
				}
			}
			if all {
				common++
			}
		}
		if common == 0 {
			t.Errorf("level %d: no world holds all %d takeoff pads", l, len(ms))
		}
	}
}

func TestFourLevelsDecode(t *testing.T) {
	a := archive(t)
	idx := a.ByType("UPWL")
	if len(idx) != 4 {
		t.Fatalf("%d UPWL resources, want 4", len(idx))
	}
	for _, i := range idx {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeLevel(chunkList(t, a, f)); err != nil {
			t.Fatalf("UPWL %d: %v", i, err)
		}
	}
}

func decodeWorlds(t *testing.T, a *pwad.Archive) []*uvtr.World {
	t.Helper()
	f, err := a.Resource(a.ByType("UVTR")[0])
	if err != nil {
		t.Fatal(err)
	}
	var out []*uvtr.World
	for _, c := range f.Chunks {
		if c.Tag != "COMM" {
			continue
		}
		data, err := a.Data(c)
		if err != nil {
			t.Fatal(err)
		}
		w, err := uvtr.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, w)
	}
	return out
}
