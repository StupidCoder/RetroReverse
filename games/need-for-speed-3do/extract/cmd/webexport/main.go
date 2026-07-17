// webexport extracts The Need for Speed's shipped web assets straight from the
// disc image: a textured GLB of a course (road + trackside slice geometry,
// built from the .trk segment spline, the streamed TRKD slice rows and the
// static slice topology in LaunchMe), per-object GLBs plus a placement
// manifest for the instanced roadside objects (RoadObjects block), and a
// textured GLB of a car (ORI3 model + SPoT textures via the SHPM "!ori" face
// map). All geometry comes from the nfs decoders — verified against the
// running game by cmd/geomoracle — following the repo's webexport standard
// (see games/ridge-racer-psx/extract/cmd/webexport).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/platform/threedo"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "webexport: "+format+"\n", a...)
	os.Exit(1)
}

// Manifest is the format-2 asset index the Studio loads.
type Manifest struct {
	Format   int          `json:"format"`
	Game     string       `json:"game"`
	Platform string       `json:"platform"`
	Native   Size         `json:"native"`
	TickHz   int          `json:"tickHz"`
	Models   []ModelIndex `json:"models,omitempty"`
}

type Size struct {
	W int `json:"w"`
	H int `json:"h"`
}

type ModelIndex struct {
	Name        string      `json:"name"`
	File        string      `json:"file"`
	Kind        string      `json:"kind"`                  // routes to a Studio renderer plugin
	Section     string      `json:"section,omitempty"`     // Studio browse-list group
	W           int         `json:"w,omitempty"`           // native pixel size (movies)
	H           int         `json:"h,omitempty"`
	ObjectsFile string      `json:"objectsFile,omitempty"` // placement manifest for the object layer
	Sky         string      `json:"sky,omitempty"`         // camera-centred horizon dome GLB
	Fly         bool        `json:"fly,omitempty"`         // present with the free-flight camera
	Camera      *CameraPose `json:"camera,omitempty"`      // opening view (course only)
}

// CameraPose is a viewer opening view in GLB axes (x, y, -z).
type CameraPose struct {
	Pos    [3]float64 `json:"pos"`
	Target [3]float64 `json:"target"`
}

// courses is the full course set — the disc's three runs (its own announcer
// audio names them CITY/COASTAL/ALPINE) of three stages each.
var courses = []struct{ id, name string }{
	{"cy1", "City 1"}, {"cy2", "City 2"}, {"cy3", "City 3"},
	{"cl1", "Coastal 1"}, {"cl2", "Coastal 2"}, {"cl3", "Coastal 3"},
	{"al1", "Alpine 1"}, {"al2", "Alpine 2"}, {"al3", "Alpine 3"},
}

// startCamera derives a course's opening view from its spline. Calibration is
// the City 1 driver's-eye camera captured from the running game at the grid
// (camObj @[0x40014], position at +0xDC/+0xE0/+0xE4 in 16.16 world units,
// orientation the identity matrix at [0x4001C]=0x485A0): it sits at segment 16
// (96 m in), 2.10 m right of the track centreline, 0.94 m up, looking straight
// down the heading. Every course starts near heading 0, so the same offsets
// place the grid view on all nine; for City 1 this reproduces the captured
// values. Heading units: 0x4000 per full circle (the billboard corner code
// shifts heading<<10 into a 0x1000000 circle). GLB axes are (x, y, -z).
func startCamera(t *nfs.Track) *CameraPose {
	seg := t.Segments[16]
	theta := float64(seg.Heading) / 0x4000 * 2 * math.Pi
	fwd := [3]float64{math.Sin(theta), 0, math.Cos(theta)}
	right := [3]float64{math.Cos(theta), 0, -math.Sin(theta)}
	p := [3]float64{
		nfs.Float(seg.Pos.X) + 2.098*right[0],
		nfs.Float(seg.Pos.Y) + 0.94,
		nfs.Float(seg.Pos.Z) + 2.098*right[2],
	}
	return &CameraPose{
		Pos:    [3]float64{p[0], p[1], -p[2]},
		Target: [3]float64{p[0] + 40*fwd[0], p[1], -(p[2] + 40*fwd[2])},
	}
}

// assets bundles a course's decoded inputs.
type assets struct {
	course string // "cy1", …
	track  *nfs.Track
	slices *nfs.SliceTables
	pkt    []byte             // the course's DriveArt packet
	root   *threedo.WrapNode  // its wwww tree
	tex    *texCache
}

func main() {
	image := flag.String("image", "", "3DO disc image")
	out := flag.String("o", "", "output directory")
	movies := flag.Bool("movies", false, "also export Movies/*.stream FMV as MP4 (needs ffmpeg)")
	flag.Parse()
	if *image == "" || *out == "" {
		die("usage: webexport -image DISC -o OUTDIR")
	}

	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		die("%v", err)
	}

	if err := os.MkdirAll(filepath.Join(*out, "models"), 0o755); err != nil {
		die("%v", err)
	}

	var models []ModelIndex
	for _, c := range courses {
		a := loadCourse(vol, c.id)
		courseFile, err := exportCourse(a, *out)
		if err != nil {
			die("course %s: %v", c.id, err)
		}
		objectsFile, err := exportObjects(a, *out)
		if err != nil {
			die("objects %s: %v", c.id, err)
		}
		skyFile, err := exportSky(a, *out)
		if err != nil {
			die("sky %s: %v", c.id, err)
		}
		models = append(models, ModelIndex{
			Name: c.name, File: courseFile,
			Kind: "nfs-course", Section: "Tracks", ObjectsFile: objectsFile,
			Sky: skyFile, Fly: true, Camera: startCamera(a.track),
		})
	}

	cars, err := exportCars(vol, *out)
	if err != nil {
		die("car: %v", err)
	}
	models = append(models, cars...)

	if *movies {
		movieModels, err := exportMovies(vol, *out)
		if err != nil {
			die("movies: %v", err)
		}
		models = append(models, movieModels...)
		fmt.Fprintf(os.Stderr, "[movies] %d exported\n", len(movieModels))
	}

	m := Manifest{
		Format: 2, Game: "The Need for Speed", Platform: "3DO",
		Native: Size{W: 320, H: 240}, TickHz: 60, Models: models,
	}
	j, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(j, '\n'), 0o644); err != nil {
		die("%v", err)
	}
	fmt.Fprintf(os.Stderr, "[manifest] %d models -> %s\n", len(models), *out)
}

// loadCourse reads and decodes everything one course needs.
func loadCourse(vol *threedo.Volume, course string) *assets {
	trk, err := vol.ReadFile("DriveData/tracks/" + course + ".trk")
	if err != nil {
		die("%v", err)
	}
	lm, err := vol.ReadFile("LaunchMe")
	if err != nil {
		die("%v", err)
	}
	// "cy1" -> "Cy1_PKT_000"
	pktName := fmt.Sprintf("DriveData/DriveArt/%c%s_PKT_000", course[0]-'a'+'A', course[1:])
	pkt, err := vol.ReadFile(pktName)
	if err != nil {
		die("%v", err)
	}
	track, err := nfs.ParseTrack(trk)
	if err != nil {
		die("%v", err)
	}
	slices, err := nfs.LoadSliceTables(lm)
	if err != nil {
		die("%v", err)
	}
	root, err := threedo.ParseWrapTree(pkt)
	if err != nil {
		die("%v", err)
	}
	return &assets{
		course: course, track: track, slices: slices,
		pkt: pkt, root: root, tex: newTexCache(pkt, root),
	}
}

// gl maps a game world-space 16.16 point into GLB axes: the game is Y-up with
// Z forward, glTF is Y-up with -Z forward.
func gl(v nfs.Vec3) [3]float32 {
	return [3]float32{float32(nfs.Float(v.X)), float32(nfs.Float(v.Y)), -float32(nfs.Float(v.Z))}
}
