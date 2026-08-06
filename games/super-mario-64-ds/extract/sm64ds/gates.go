package sm64ds

// Mission gates (Part V §8).
//
// The object table's star layer is not the only thing that decides whether an
// object is in a mission. Fifty of the 301 placed actors read the save's star
// bits through $020137E0(course, star) = saveStars[$0209CAB4+course] & (1<<star),
// usually via $0202A6C8(star), which asks about the level you are standing in.
// Whomp's Fortress is the clearest case: the Whomp King destroys itself and the
// summit tower refuses to be created on exactly complementary conditions over
// the level, the star being played ($0209F220 — the same global the layer
// walker compares layers against) and whether star 1 is already collected.
//
// Rather than decode fifty bespoke conditions, this RUNS them. For each mission
// the oracle is told the level, the star, and "stars 1..star-1 collected" — the
// state a player is in the first time they play that mission — and every
// placement of a gated actor is created and initialised AT ITS OWN POSITION
// (the tower's condition includes its own height). An actor that exists under
// some missions and not others is gated; one that refuses under all of them is
// the oracle's bare environment talking, not the game, and is left alone. The
// difference is the measurement; the absolute outcome is not trusted.
//
// Only the numbered painting courses are swept. "The first time you play
// mission N" is a real state there. It is not one for the castle hub, whose
// doors count stars across the whole game, so those levels are deliberately
// left alone rather than depicted under a made-up save.

import (
	"path/filepath"
	"sort"
	"strings"
)

// MissionGate is one placement's mission membership as the game's own code
// decides it, for a placement the star layer alone does not separate.
type MissionGate struct {
	Actor    int    `json:"actor"`
	Params   [3]int `json:"params"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Z        int    `json:"z"`
	Missions []int  `json:"missions"` // the stars in which the actor exists
}

// GatedActors are the placed actors whose code reads the save's star bits,
// measured by cmd/gateprobe -sweep: 50 of the 301 placed actors, with 287 of
// the 301 resolving to a vtable, so what is left out is a measured set and not
// an assumption. Only these can answer differently from one mission to the
// next, so only these are run seven times.
var GatedActors = map[int]bool{
	13: true, 15: true, 16: true, 17: true, 19: true, 20: true, 21: true, 22: true, 23: true,
	24: true, 25: true, 37: true, 43: true, 44: true, 46: true, 49: true, 52: true, 56: true,
	57: true, 61: true, 164: true, 165: true, 176: true, 185: true, 187: true, 188: true,
	189: true, 200: true, 201: true, 202: true, 209: true, 210: true, 216: true, 221: true,
	234: true, 242: true, 245: true, 248: true, 259: true, 262: true, 263: true, 272: true,
	298: true, 321: true, 322: true, 323: true, 324: true, 340: true, 341: true, 350: true,
}

// ConfigFn resolves the overlay an actor's profile is meaningful under — an
// actor profile only means anything while its overlay is banked, and several
// banks share each RAM slot, so this is not derivable from the address.
type ConfigFn func(actor int, par [3]int) (int, bool)

type gateKey struct {
	actor   int
	par     [3]int
	x, y, z int
}

// MissionGates sweeps every numbered painting course and returns, per stage
// stem, the placements whose existence varies from mission to mission.
func MissionGates(ls *LevelSet, o *Oracle, cfgFor ConfigFn, log func(stem string, n int)) map[string][]MissionGate {
	out := map[string][]MissionGate{}
	seenStem := map[string]bool{}
	for id := 0; id < NumLevels; id++ {
		course := ls.Course(id)
		if course < 0 || course >= StarNameCourses {
			continue // not a numbered course: no honest "first play of mission N"
		}
		lv, err := ls.Level(id)
		if err != nil || lv.BMDPath == "" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if seenStem[stem] {
			continue // the exporter keeps the first level that uses a stage
		}
		seenStem[stem] = true
		if g := levelGates(o, lv, id, course, cfgFor); len(g) > 0 {
			out[stem] = g
			if log != nil {
				log(stem, len(g))
			}
		}
	}
	o.ClearMission()
	o.SetSpawnPos(0, 0, 0)
	return out
}

func levelGates(o *Oracle, lv *Level, id, course int, cfgFor ConfigFn) []MissionGate {
	var order []gateKey
	seen := map[gateKey]bool{}
	for _, ob := range lv.Objects {
		if !GatedActors[ob.Actor] {
			continue
		}
		k := gateKey{ob.Actor, ob.Params, int(ob.X), int(ob.Y), int(ob.Z)}
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
		}
	}
	if len(order) == 0 {
		return nil
	}
	alive := map[gateKey][]int{}
	ran := map[gateKey]bool{}
	for star := 1; star <= StarsPerCourse; star++ {
		o.SetMission(id, course, star, FirstPlaySave(star))
		for _, k := range order {
			cfg, ok := cfgFor(k.actor, k.par)
			if !ok {
				continue
			}
			o.SetSpawnPos(int32(k.x)<<12, int32(k.y)<<12, int32(k.z)<<12)
			if err := o.LoadConfig(cfg); err != nil {
				continue
			}
			run := o.RunActor(k.actor, cfg, k.par)
			if run.Create == 0 {
				continue // no profile under this config: not a measurement
			}
			ran[k] = true
			if !run.Refused && !run.Destroyed {
				alive[k] = append(alive[k], star)
			}
		}
	}
	var out []MissionGate
	for _, k := range order {
		m := alive[k]
		if !ran[k] || len(m) == 0 || len(m) == StarsPerCourse {
			continue // never measured, always dead (the bare oracle), or always alive
		}
		out = append(out, MissionGate{Actor: k.actor, Params: k.par, X: k.x, Y: k.y, Z: k.z, Missions: m})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Actor < out[j].Actor })
	return out
}

// FirstPlaySave is the per-course star bitmask for "the first time you play
// mission `star`": every earlier mission of the course is done and this one is
// not. $020137E0 tests bit `star` directly, so star N is bit N.
func FirstPlaySave(star int) uint8 {
	var bits uint8
	for s := 1; s < star; s++ {
		bits |= 1 << uint(s)
	}
	return bits
}
