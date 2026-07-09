package main

// paths.go exports the flying objects' movement data (rr/flight.go) as
// levels/course.paths.json for the course viewer's animation layer:
//
//   - the helicopter's two waypoint routes (positions in GLB units, per-key
//     flight time and hover in 30 fps game ticks, and the script's commanded
//     yaw converted to a three.js rotation.y), flown alternately — plus the
//     body and rotor as separate GLBs so the viewer can spin the rotor the
//     way the game does (331/4096 of a turn per tick about the pitched axis);
//   - the airplane's linear glide: spawn, per-tick delta and lifetime; the
//     viewer jumps it back to the spawn when the path ends.
//
// PSX model units become GLB units through (x, −y, −z)/1024, and a yaw
// becomes rotation.y of yaw/4096·2π (both as in objects.go).

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

type pathsDoc struct {
	FPS        int        `json:"fps"` // game ticks per second
	Helicopter heliPath   `json:"helicopter"`
	Airplane   planePath  `json:"airplane"`
}

type heliPath struct {
	Body      string     `json:"body"`
	Rotor     string     `json:"rotor"`
	RotorRate float64    `json:"rotorRate"` // radians per second
	Routes    []route    `json:"routes"`    // flown alternately, each a loop
}

type route struct {
	Addr  string     `json:"addr"`
	Start [3]float64 `json:"start"`
	Yaw   float64    `json:"yaw"` // rotation.y, radians
	Keys  []pathKey  `json:"keys"`
}

type pathKey struct {
	Pos  [3]float64 `json:"pos"`
	Dur  int        `json:"dur"`  // ticks flying to this key
	Hold int        `json:"hold"` // ticks hovering after arrival
	Yaw  float64    `json:"yaw"`
}

type planePath struct {
	Model  string     `json:"model"`
	Start  [3]float64 `json:"start"`
	Delta  [3]float64 `json:"delta"` // GLB units per tick
	Frames int        `json:"frames"`
	Yaw    float64    `json:"yaw"`
}

func glbPos(x, y, z int32) [3]float64 {
	return [3]float64{float64(x) / 1024, -float64(y) / 1024, -float64(z) / 1024}
}

func rotY(yaw int16) float64 { return float64(yaw) / 4096 * 2 * math.Pi }

// exportPaths writes levels/course.paths.json and the helicopter body/rotor
// part GLBs, returning the manifest-relative path of the json.
func exportPaths(a *assets, out string, cps []rr.Checkpoint) (string, error) {
	scripts, err := rr.HeliScripts(a.exe)
	if err != nil {
		return "", err
	}
	plane := rr.Airplane(a.exe)

	// The heli parts, pitched like the drawer (see rr.Dynamics): body 188 and
	// rotor 189 separately, so the viewer can spin the rotor.
	var heli rr.Dynamic
	for _, d := range rr.Dynamics(a.exe) {
		if d.Name == "Helicopter" {
			heli = d
		}
	}
	set := 0
	if heli.X != 0 || heli.Z != 0 {
		seg := rr.NearestSegment(cps, heli.X, heli.Z)
		set = rr.SetForProgress(int32(seg) * 256)
	}
	pitch := rr.PitchMatrix(heli.Pitch)
	for _, part := range []struct {
		id   int
		file string
	}{{188, "models/heli-body.glb"}, {189, "models/heli-rotor.glb"}} {
		b := newMeshBuilder(a.vrams[set])
		addObjectXform(b, &a.objs[part.id], pitch, [3]int32{0, 0, 0}, set)
		if err := b.Write(filepath.Join(out, part.file)); err != nil {
			return "", err
		}
	}

	doc := pathsDoc{
		FPS: 30,
		Helicopter: heliPath{
			Body:      "models/heli-body.glb",
			Rotor:     "models/heli-rotor.glb",
			RotorRate: 331.0 / 4096 * 2 * math.Pi * 30,
		},
		Airplane: planePath{
			Model:  "models/special-190.glb",
			Start:  glbPos(plane.X, plane.Y, plane.Z),
			Delta:  glbPos(plane.DX, plane.DY, plane.DZ),
			Frames: plane.Frames,
			Yaw:    rotY(0xF00),
		},
	}
	for _, s := range scripts {
		r := route{
			Addr:  fmt.Sprintf("0x%08X", s.Addr),
			Start: glbPos(s.X, s.Y, s.Z),
			Yaw:   rotY(s.Yaw),
		}
		for _, k := range s.Keys {
			r.Keys = append(r.Keys, pathKey{
				Pos: glbPos(k.X, k.Y, k.Z), Dur: k.Dur, Hold: k.Hold, Yaw: rotY(k.Yaw),
			})
		}
		doc.Helicopter.Routes = append(doc.Helicopter.Routes, r)
	}

	rel := "levels/course.paths.json"
	if err := os.MkdirAll(filepath.Join(out, "levels"), 0o755); err != nil {
		return "", err
	}
	j, _ := json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(filepath.Join(out, rel), append(j, '\n'), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[levels] flight paths: %d heli routes (%d+%d keys), airplane %d ticks\n",
		len(scripts), len(doc.Helicopter.Routes[0].Keys), len(doc.Helicopter.Routes[1].Keys), plane.Frames)
	return rel, nil
}
