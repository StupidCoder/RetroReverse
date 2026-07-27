// webexport builds The Need for Speed's Retro-X game tree straight from the
// disc image: each of the nine courses becomes a scene3d level (road +
// trackside slice geometry from the .trk segment spline, the streamed TRKD
// slice rows and the static slice topology in LaunchMe; the horizon rings as
// a camera-attached sky dome; every RoadObjects placement at its engine
// transform), and every car .wrapFam a model3d object (ORI3 model + SPoT
// textures via the SHPM "!ori" face map). All geometry comes from the nfs
// decoders — verified against the running game by cmd/geomoracle.
//
// Usage (from games/need-for-speed-3do/):
//
//	go run ./extract/cmd/webexport -in "Need for Speed.bin"
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/need-for-speed-3do/extract/nfs"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
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
	cli.Main("need-for-speed-3do", run)
}

// group maps a course id to its run's Studio browse group — the disc's own
// announcer audio names the three runs CITY, COASTAL and ALPINE.
func group(id string) string {
	switch id[:2] {
	case "cy":
		return "City"
	case "cl":
		return "Coastal"
	default:
		return "Alpine"
	}
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in 'Need for Speed.bin' [-o DIR]")
	}
	data, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	vol, err := threedo.Open(data)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("The Need for Speed")
	b.SetPlatform("3DO")
	b.SetYear(1994)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 240},
		TickHz: 60,
		Filter: "crt",
		// the 3DO cel engine point-samples its cels
		TexFilter: "nearest",
	})
	ctx.Stage("objects")
	ctx.Stage("levels")

	// The decoders write format-1 GLBs + side files into a scratch tree;
	// everything is then registered through the builder.
	scratch, err := os.MkdirTemp("", "nfs-webexport-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	if err := os.MkdirAll(filepath.Join(scratch, "models"), 0o755); err != nil {
		return err
	}

	copyTo := func(src, dir, name string) error {
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()
		dst, err := b.Path(dir, name)
		if err != nil {
			return err
		}
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}

	for ci, c := range courses {
		a := loadCourse(vol, c.id)
		courseFile, err := exportCourse(a, scratch)
		if err != nil {
			return fmt.Errorf("course %s: %v", c.id, err)
		}
		objectsFile, err := exportObjects(a, scratch)
		if err != nil {
			return fmt.Errorf("objects %s: %v", c.id, err)
		}
		skyFile, err := exportSky(a, scratch)
		if err != nil {
			return fmt.Errorf("sky %s: %v", c.id, err)
		}

		if err := copyTo(filepath.Join(scratch, courseFile), "levels", filepath.Base(courseFile)); err != nil {
			return err
		}
		if err := copyTo(filepath.Join(scratch, skyFile), "levels", filepath.Base(skyFile)); err != nil {
			return err
		}

		// every roadside prop the course places becomes a model3d asset
		var objsDoc objectsDoc
		raw, err := os.ReadFile(filepath.Join(scratch, objectsFile))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &objsDoc); err != nil {
			return err
		}
		refs := map[string]string{} // "models/obj-cy1-10.glb" -> asset id
		for _, o := range objsDoc.Objects {
			if _, ok := refs[o.Model]; ok {
				continue
			}
			file := filepath.Base(o.Model)
			id := strings.TrimSuffix(file, ".glb")
			if err := copyTo(filepath.Join(scratch, o.Model), "objects", file); err != nil {
				return err
			}
			name := fmt.Sprintf("%s prop %02d", c.name, o.Def)
			b.AddObject(schema.Asset{ID: id, Name: name, Group: group(c.id) + " scenery"}, &schema.Object{
				Type: schema.ObjectModel3D, Name: name, Model: file,
			})
			refs[o.Model] = id
		}

		cam := startCamera(a.track)
		// The horizon dome is a unit-radius cylinder around the camera — the
		// game paints it as a pure backdrop, so it must draw first with the
		// depth test off or it occludes the whole course.
		noDepth := false
		doc := &schema.Level{
			Type: schema.LevelScene3D,
			Scene: &schema.Scene{Layers: []schema.Layer{
				{ID: "course", File: filepath.Base(courseFile)},
				{ID: "sky", Name: "Horizon", File: filepath.Base(skyFile), Mode: "toggle",
					Attach: "camera", Role: "sky", RenderOrder: -1, DepthTest: &noDepth},
			}},
			Camera: &schema.Camera{
				Mode: "fly", FOV: 55, Near: 0.2, Far: 8000, Fly: &schema.Fly{Speed: 30},
				Pos:    cam.Pos[:],
				Target: cam.Target[:],
			},
		}
		for pi, o := range objsDoc.Objects {
			pl := schema.Placement{
				ID: pi, Object: refs[o.Model],
				Pos: []float64{o.Pos[0], o.Pos[1], o.Pos[2]},
				Props: map[string]any{
					"def": o.Def, "type": o.Type, "segment": o.Segment,
					"rawYaw": o.RawYaw, "flags": o.Flags,
				},
			}
			if o.Rot != [3]float64{} {
				pl.Rot = []float64{o.Rot[0], o.Rot[1], o.Rot[2]}
			}
			doc.Placements = append(doc.Placements, pl)
		}
		b.AddLevel(schema.Asset{ID: c.id, Name: c.name, Group: group(c.id) + " tracks"}, doc)
		ctx.Progress("levels", ci+1, len(courses), fmt.Sprintf("%s: %d placements", c.id, len(doc.Placements)))
	}

	// the cars — written by the format-1 exporter into scratch/models, then
	// registered as model3d objects grouped by the fleet sections
	cars, err := exportCars(vol, scratch)
	if err != nil {
		return fmt.Errorf("car: %v", err)
	}
	for i, m := range cars {
		file := filepath.Base(m.File)
		if err := copyTo(filepath.Join(scratch, m.File), "objects", file); err != nil {
			return err
		}
		id := strings.TrimSuffix(file, ".glb")
		b.AddObject(schema.Asset{ID: id, Name: m.Name, Group: m.Section}, &schema.Object{
			Type: schema.ObjectModel3D, Name: m.Name, Model: file,
		})
		ctx.Progress("objects", i+1, len(cars), m.Name)
	}
	return nil
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
