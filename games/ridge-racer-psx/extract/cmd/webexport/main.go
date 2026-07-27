// webexport extracts Ridge Racer's shipped web assets straight from the CD
// image: one textured GLB per selectable car (the car-select models, with
// wheels) and a textured GLB of the whole course, plus manifest.json. All
// geometry and texture data comes from the rr decoders (MAP.RRM, OBJ.RRO,
// IDX.HED, TEX*.TMS) — verified bit-exact against the running game by
// cmd/geomoracle — and the car catalog comes from the executable's own car
// table (0x80056B40) and name list.
package main

import (
	"io"

	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/psx"
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
	Section     string      `json:"section,omitempty"`     // Studio browse-list group ("Course", "Cars", "Objects")
	ObjectsFile string      `json:"objectsFile,omitempty"` // placement manifest for the object layer
	PathsFile   string      `json:"pathsFile,omitempty"`   // flight-path manifest for the animation layer
	Fly         bool        `json:"fly,omitempty"`         // present the item with the fly camera
	Camera      *CameraPose `json:"camera,omitempty"`      // opening view (course only)
}

// CameraPose is a viewer opening view in GLB axes.
type CameraPose struct {
	Pos    [3]float64 `json:"pos"`
	Target [3]float64 `json:"target"`
}

// startCam is the course viewer's opening view: the race-start camera pose
// captured from the running game at the grid. The camera's world position and
// facing are recovered from the GTE view transform (R, TR) at the first race
// frame — cam = P_known − Rᵀ·TR for a known-placement object in frame, forward
// = R's third row — then converted to GLB axes (x, −y, −z)/1024. The target is
// 20 GLB units down the captured forward (0.9634, 0, 0.2683).
var startCam = CameraPose{
	Pos:    [3]float64{103.646, 0.098, -162.713},
	Target: [3]float64{122.914, 0.098, -157.347},
}

// assets bundles everything the stages need: the decoded files and the
// replayed texture VRAM.
type assets struct {
	objs  []rr.Object
	track *rr.Track
	grid  *rr.Grid
	vrams [3]*rr.VRAM // one texture state per scenery set
	exe   []byte      // text segment (0x80010000..)
}

func main() {
	cli.Main("ridge-racer-psx", run2)
}

func run2(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in 'Ridge Racer (Track 01).bin' [-o DIR]")
	}
	data, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	vol, err := psx.Open(data)
	if err != nil {
		return err
	}
	a, err := load(vol)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Ridge Racer")
	b.SetPlatform("Sony PlayStation")
	b.SetYear(1994)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 320, H: 240},
		TickHz: 60,
		Filter: "crt",
		// the PSX GPU point-samples textures
		TexFilter: "nearest",
	})
	ctx.Stage("objects")
	ctx.Stage("levels")

	// The stages write the format-1 GLBs + side files into a scratch tree;
	// everything is then registered through the builder.
	scratch, err := os.MkdirTemp("", "rr-webexport-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)

	cars, err := exportCars(a, scratch)
	if err != nil {
		return err
	}
	specials, err := exportSpecials(a, scratch)
	if err != nil {
		return err
	}
	trackIdx, err := exportTrack(a, scratch)
	if err != nil {
		return err
	}
	_ = trackIdx

	// copy every produced GLB into objects/ (track.glb goes to levels/)
	glbs, _ := filepath.Glob(filepath.Join(scratch, "models", "*.glb"))
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
	refs := map[string]string{} // "models/<f>.glb" -> asset id
	registered := map[string]bool{}
	addObj := func(file, name, group string) error {
		id := strings.TrimSuffix(file, ".glb")
		if registered[file] {
			return nil
		}
		registered[file] = true
		if err := copyTo(filepath.Join(scratch, "models", file), "objects", file); err != nil {
			return err
		}
		b.AddObject(schema.Asset{ID: id, Name: name, Group: group}, &schema.Object{
			Type: schema.ObjectModel3D, Name: name, Model: file,
		})
		refs["models/"+file] = id
		return nil
	}
	for _, m := range cars {
		if err := addObj(filepath.Base(m.File), m.Name, "Cars"); err != nil {
			return err
		}
	}
	for _, m := range specials {
		if err := addObj(filepath.Base(m.File), m.Name, "Trackside"); err != nil {
			return err
		}
	}
	for _, g := range glbs {
		f := filepath.Base(g)
		if f == "track.glb" || registered[f] {
			continue
		}
		name := strings.TrimSuffix(f, ".glb")
		if err := addObj(f, "Object "+strings.TrimPrefix(name, "obj-"), "Scenery"); err != nil {
			return err
		}
	}
	if err := copyTo(filepath.Join(scratch, "models", "track.glb"), "levels", "track.glb"); err != nil {
		return err
	}

	// the course level: track layer, race-start camera, scenery placements,
	// the helicopter + airplane as flyer behaviours
	doc := &schema.Level{
		Type:  schema.LevelScene3D,
		Scene: &schema.Scene{Layers: []schema.Layer{{ID: "track", File: "track.glb"}}},
		Camera: &schema.Camera{
			Mode: "fly", FOV: 55, Near: 0.05, Far: 3000, Fly: &schema.Fly{Speed: 20},
			Pos:    startCam.Pos[:],
			Target: startCam.Target[:],
		},
	}
	var objsDoc struct {
		Objects []struct {
			Model string     `json:"model"`
			Pos   [3]float64 `json:"pos"`
			Rot   [3]float64 `json:"rot"`
			ID    int        `json:"id"`
			Addr  string     `json:"addr"`
		} `json:"objects"`
	}
	if raw, err := os.ReadFile(filepath.Join(scratch, "levels", "course.objects.json")); err == nil {
		if err := json.Unmarshal(raw, &objsDoc); err != nil {
			return err
		}
	}
	pid := 0
	for _, o := range objsDoc.Objects {
		asset, ok := refs[o.Model]
		if !ok {
			continue
		}
		pl := schema.Placement{
			ID: pid, Object: asset,
			Pos:   []float64{o.Pos[0], o.Pos[1], o.Pos[2]},
			Props: map[string]any{"addr": o.Addr},
		}
		if o.Rot != [3]float64{} {
			pl.Rot = []float64{o.Rot[0], o.Rot[1], o.Rot[2]}
		}
		doc.Placements = append(doc.Placements, pl)
		pid++
	}
	var paths pathsDoc
	if raw, err := os.ReadFile(filepath.Join(scratch, "levels", "course.paths.json")); err == nil {
		if err := json.Unmarshal(raw, &paths); err != nil {
			return err
		}
	}
	toKeys := func(r route) []schema.FlyKey {
		var keys []schema.FlyKey
		for _, k := range r.Keys {
			keys = append(keys, schema.FlyKey{
				Pos: []float64{k.Pos[0], k.Pos[1], k.Pos[2]}, Dur: k.Dur, Hold: k.Hold, Yaw: k.Yaw,
			})
		}
		return keys
	}
	if len(paths.Helicopter.Routes) > 0 {
		r := paths.Helicopter.Routes[0]
		for _, part := range []string{paths.Helicopter.Body, paths.Helicopter.Rotor} {
			if asset, ok := refs[part]; ok {
				pl := schema.Placement{
					ID: pid, Object: asset,
					Pos:      []float64{r.Start[0], r.Start[1], r.Start[2]},
					Behavior: &schema.Behavior{Kind: "flyer", Keys: toKeys(r)},
					Props:    map[string]any{"route": r.Addr, "note": "route 0 of 2 (flown alternately in game)"},
				}
				doc.Placements = append(doc.Placements, pl)
				pid++
			}
		}
	}
	if asset, ok := refs[paths.Airplane.Model]; ok {
		p0 := paths.Airplane.Start
		end := [3]float64{
			p0[0] + paths.Airplane.Delta[0]*float64(paths.Airplane.Frames),
			p0[1] + paths.Airplane.Delta[1]*float64(paths.Airplane.Frames),
			p0[2] + paths.Airplane.Delta[2]*float64(paths.Airplane.Frames),
		}
		yaw := paths.Airplane.Yaw
		doc.Placements = append(doc.Placements, schema.Placement{
			ID: pid, Object: asset,
			Pos: []float64{p0[0], p0[1], p0[2]},
			Behavior: &schema.Behavior{Kind: "flyer", Keys: []schema.FlyKey{
				{Pos: []float64{p0[0], p0[1], p0[2]}, Dur: paths.Airplane.Frames, Yaw: yaw},
				{Pos: []float64{end[0], end[1], end[2]}, Dur: 1, Yaw: yaw},
			}},
		})
		pid++
	}

	b.AddLevel(schema.Asset{ID: "course", Name: "Ridge Racer Course", Group: "Course"}, doc)
	ctx.Progress("levels", 1, 1, fmt.Sprintf("course: %d placements", len(doc.Placements)))
	return nil
}

// load reads and decodes every needed file off the disc.
func load(vol *psx.Volume) (*assets, error) {
	a := &assets{}
	objData, err := vol.ReadFile("OBJ.RRO")
	if err != nil {
		return nil, err
	}
	if a.objs, err = rr.ParseRRO(objData); err != nil {
		return nil, err
	}
	mapData, err := vol.ReadFile("MAP.RRM")
	if err != nil {
		return nil, err
	}
	if a.track, err = rr.ParseRRM(mapData); err != nil {
		return nil, err
	}
	idxData, err := vol.ReadFile("IDX.HED")
	if err != nil {
		return nil, err
	}
	if a.grid, err = rr.ParseIDX(idxData); err != nil {
		return nil, err
	}
	var rects [][]rr.Rect
	for _, name := range []string{"TEX4.TMS", "TEX0.TMS", "TEX1.TMS", "TEX2.TMS", "TEX3.TMS"} {
		d, err := vol.ReadFile(name)
		if err != nil {
			return nil, err
		}
		rs, err := rr.ParseTMS(d)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", name, err)
		}
		rects = append(rects, rs)
	}
	// The race texture states: one per scenery set (the quadrant pages are
	// paged by course progress; everything else is identical across sets).
	a.vrams = rr.NewSceneryVRAMs(rects[0], rects[1], rects[2], rects[3], rects[4])
	_, exe, err := vol.BootEXE()
	if err != nil {
		return nil, err
	}
	a.exe = exe.Text
	return a, nil
}
