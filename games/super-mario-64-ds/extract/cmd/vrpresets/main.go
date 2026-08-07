// vrpresets writes the per-level world-mode (VR) preset for every shipped
// stage, from the cartridge's own numbers.
//
// A preset is what a walk-in session needs that no amount of measuring a GLB
// can tell you (site/src/xrpreset.js): how big a level unit is in metres, where
// you stand, which way you face, and whether there is a sky. Retro-X carries no
// units and is right not to, so this is written down once per level.
//
// It is AUTHORED data — you tune one from inside a headset and reload — so this
// never overwrites a file that already exists. site/vr/super-mario-64-ds/
// bombhei-map.json is the hand-tuned reference every value here follows, and it
// is left alone.
//
//	vrpresets [-rom img] [-extracted dir] [-levels DIR] [-o DIR] [-index FILE]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/platform/nds"
)

// metresPerUnit is Mario. He is the one character whose height Nintendo states
// — 155 cm — and every actor placement in every SM64DS level document carries
// the same 0.008 scale, so the ruler is the same in all of them: idle Mario
// (su_wait, measured POSED) is 0.1415 stage units, and 1.55 / 0.1415 = 10.95
// metres per unit. Derived once in bombhei-map.json, which explains the working.
const metresPerUnit = 10.95

// marioAhead is how far in front of you the ruler stands, in metres. He is
// there so the scale can be checked by eye: if he does not come up to your
// eyes, metresPerUnit is wrong.
const marioAhead = 2.0

const toStage = 1.0 / 1000 // placement short -> stage-GLB units

type preset struct {
	Doc           []string       `json:"_"`
	MetresPerUnit float64        `json:"metresPerUnit"`
	Spawn         spawn          `json:"spawn"`
	SkyDoc        []string       `json:"_sky,omitempty"`
	Sky           any            `json:"sky"`
	FogDoc        []string       `json:"_fog,omitempty"`
	Fog           *fog           `json:"fog"`
	Torch         map[string]any `json:"torch"`
	FloorDoc      []string       `json:"_floor,omitempty"`
	Floor         map[string]any `json:"floor"`
	Teleport      map[string]any `json:"teleport"`
	PropsDoc      []string       `json:"_props,omitempty"`
	Props         []prop         `json:"props"`
	Sounds        []any          `json:"sounds"`
}

type spawn struct {
	Pos []float64 `json:"pos"`
	Dir []float64 `json:"dir"`
}
type fog struct {
	Color string  `json:"color"`
	Near  float64 `json:"near"`
	Far   float64 `json:"far"`
}
type prop struct {
	Object string    `json:"object"`
	Pos    []float64 `json:"pos"`
	RotY   float64   `json:"rotY"`
	Scale  float64   `json:"scale"`
	Anim   string    `json:"anim"`
}

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	levels := flag.String("levels", "../../site/public/super-mario-64-ds/levels", "shipped level documents")
	out := flag.String("o", "../../site/vr/super-mario-64-ds", "preset directory")
	index := flag.String("index", "../../site/vr/index.json", "the preset index the viewer fetches")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		sm64ds.Die(err)
	}
	wrote, kept := 0, 0
	var ids []string
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil || lv.BMDPath == "" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		slug := slugify(strings.TrimSuffix(stem, "_all"))
		// Only stages the Studio actually ships a document for: a preset for a
		// level with no level document is a preset nothing can open.
		if _, err := os.Stat(filepath.Join(*levels, slug+".json")); err != nil {
			continue
		}
		ids = append(ids, "super-mario-64-ds/"+slug)
		path := filepath.Join(*out, slug+".json")
		if _, err := os.Stat(path); err == nil {
			kept++ // authored already; never clobber a tuned preset
			continue
		}
		p, err := build(ls, *ext, lv, stem)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%-24s %v\n", stem, err)
			continue
		}
		buf, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			sm64ds.Die(err)
		}
		if err := os.WriteFile(path, append(buf, '\n'), 0o644); err != nil {
			sm64ds.Die(err)
		}
		fmt.Printf("%-24s spawn %7.3f %7.3f %7.3f  yaw %7.2f  %s\n",
			slug, p.Spawn.Pos[0], p.Spawn.Pos[1], p.Spawn.Pos[2], p.Props[0].RotY,
			map[bool]string{true: "sky", false: "indoor"}[p.Fog != nil])
		wrote++
	}
	sort.Strings(ids)
	if err := writeIndex(*index, ids); err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("%d presets written, %d already authored, %d in the index\n", wrote, kept, len(ids))
}

// build derives one level's preset.
func build(ls *sm64ds.LevelSet, ext string, lv *sm64ds.Level, stem string) (*preset, error) {
	if len(lv.Entrances) == 0 {
		return nil, fmt.Errorf("no entrance record")
	}
	// SPAWN. Not chosen: the game's own. Object-table type 1 is an entrance
	// record and the level's FIRST one is where the game stands the player
	// (handler $020FE6C8, writeup Part V section 2). Its y is NOT used — that
	// is the height you are DROPPED from, and world mode wants the floor, so a
	// ray goes down onto the level's own collision, the same query the game
	// walks ($01FFD3F8, reimplemented in sm64ds.KCL.RaycastDown).
	e := lv.Entrances[0]
	kcl, kerr := loadKCL(ls, ext, strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl"))
	y := e.Y
	floorSource := "visible"
	if kerr == nil {
		floorSource = "collision"
		if h, ok := groundUnder(kcl, e.X, e.Y, e.Z); ok {
			y = h
		}
	}
	yaw := e.RotY * math.Pi / 180
	dx, dz := math.Sin(yaw), math.Cos(yaw) // the heading convention every placement uses
	px, pz := e.X*toStage, e.Z*toStage
	ahead := marioAhead / metresPerUnit

	p := &preset{
		MetresPerUnit: metresPerUnit,
		Spawn: spawn{
			Pos: []float64{r3(px), r3(y * toStage), r3(pz)},
			Dir: []float64{r4(dx), 0, r4(dz)},
		},
		Torch:    map[string]any{"flicker": 0},
		Floor:    map[string]any{"source": floorSource, "maxSlope": 50},
		Teleport: map[string]any{"speed": 9, "maxRange": 40, "maxRise": 2.5, "markerRadius": 0.6, "snapTurn": 30, "blockers": true},
		Props: []prop{{
			Object: "mario-model-mg",
			Pos:    []float64{r3(px + dx*ahead), r3(y * toStage), r3(pz + dz*ahead)},
			RotY:   r3(e.RotY), Scale: 0.008, Anim: "su_wait",
		}},
		Sounds: []any{},
	}
	p.Doc = []string{
		fmt.Sprintf("%s, at Mario's own scale.", stem),
		"",
		"Derived by extract/cmd/vrpresets from the cartridge, following the",
		"reference preset bombhei-map.json — which explains the working for every",
		"number here and is the one to read first. Authored data: edit freely, the",
		"generator will not overwrite a file that exists.",
		"",
		"SCALE. Mario is the ruler. Idle mario_model_mg at the 0.008 scale every",
		"actor placement carries is 0.1415 stage units posed, and 1.55 / 0.1415 =",
		"10.95 metres per unit. The same in every level, because the scale is.",
		"",
		fmt.Sprintf("SPAWN. The level's own first entrance record: (%.0f, %.0f, %.0f) yaw %.0f.",
			e.X, e.Y, e.Z, e.RotY),
		fmt.Sprintf("The %0.f is a drop height, not a floor, so a ray onto the level's own", e.Y),
		fmt.Sprintf("collision puts the feet at %.0f world units instead.", y),
	}
	p.PropsDoc = []string{
		"Mario, two metres ahead and facing the way you are: he is the ruler this",
		"preset is measured with, so if he does not come up to your eyes the scale",
		"is wrong.",
	}
	p.FloorDoc = []string{
		"The game's own walkable surface (the level's .kcl, shipped as the hidden",
		"col_* layer) rather than the art. maxSlope 50 because Mario's hills are",
		"steep.",
	}
	if kerr != nil {
		p.FloorDoc = []string{
			"No .kcl for this stage, so the visible mesh is the floor.",
		}
	}

	// SKY and FOG. A stage with a skybox is outdoors: draw the dome, and give
	// it aerial perspective in the colour of its own horizon. A stage without
	// one is an interior, where distance haze would be an invention — so no fog
	// at all, which is also what leaves the far plane alone.
	if lv.SkyPath == "" {
		p.Sky = false
		p.SkyDoc = []string{"No skybox in the level settings: this stage is an interior."}
		// An interior has no sky to take a haze colour from and its rooms are
		// lit, not a dungeon — so the honest answer is no fog. World mode does
		// not accept that answer: a preset with fog:null enters the session and
		// draws nothing (checked A/B on this very level, same clean load, only
		// the fog field differing). "Fog IS the torch" is the mode's own design
		// note and the far-plane policy hangs off it.
		//
		// So the fog here is present in order to exist, and set beyond the room
		// so it tints nothing: black, starting past the stage's own diagonal.
		// If the no-fog path is ever fixed, delete the block and the interiors
		// look identical.
		far := 250.0
		if lo, hi, ok := stageExtent(ls, ext, lv.BMDPath); ok {
			far = math.Round(math.Hypot(hi[0]-lo[0], hi[2]-lo[2])*metresPerUnit*1.5/10) * 10
		}
		if far < 80 {
			far = 80
		}
		p.Fog = &fog{Color: "#000000", Near: math.Round(far * 0.9), Far: far}
		p.FogDoc = []string{
			"Not atmosphere: a workaround. An interior has no sky to take a haze",
			"colour from, so the honest value is none — but world mode draws",
			"nothing at all when fog is null (A/B on this level, same load, only",
			"that field differing), because the mode's far-plane policy hangs off",
			"the fog. This one is black and starts past the far wall, so it tints",
			"nothing; it is here only to exist. Delete it if that is ever fixed.",
		}
		return p, nil
	}
	p.Sky = map[string]any{"scale": "auto", "fog": false}
	p.SkyDoc = []string{
		"scale:auto: the dome is authored tens of thousands of units across, which",
		"is outside any far plane worth having. The viewer centres it on the EYE",
		"being drawn, so the binocular disparity is zero however near it is pulled.",
	}
	col, err := skyHorizon(ls, ext, lv.SkyPath)
	if err != nil {
		return p, nil
	}
	// Aerial perspective scaled to the stage: bombhei-map is a 175 m course
	// with fog 40..380, and 380 is 1.5x its own horizontal diagonal. Take the
	// same ratio from each stage's own extent so a small courtyard does not get
	// a mountain's haze.
	diag := 250.0
	if lo, hi, ok := stageExtent(ls, ext, lv.BMDPath); ok {
		diag = math.Hypot(hi[0]-lo[0], hi[2]-lo[2]) * metresPerUnit
	}
	far := math.Round(diag*1.5/10) * 10
	if far < 80 {
		far = 80
	}
	p.Fog = &fog{Color: col, Near: math.Round(far / 9.5), Far: far}
	p.FogDoc = []string{
		"Daylight, so this is aerial perspective and not a torch: distance haze",
		"only, no flicker. The colour is the lightest band of the stage's own",
		fmt.Sprintf("skybox (%s), over its upper half. The distance is 1.5x the stage's own", filepath.Base(lv.SkyPath)),
		fmt.Sprintf("horizontal diagonal, %.0f m, the ratio bombhei-map was tuned to.", diag),
	}
	return p, nil
}

// skyHorizon returns the colour of the stage's own sky as "#rrggbb": the mean
// of the LARGEST texture in the skybox model, over its upper half.
//
// Three things that rule gets right, each of which a simpler one gets wrong. A
// vrbox is more than one texture — vr01 carries a 32x32 `sun` as well as its
// 256x256 dome — and picking the brightest texel lands on the sun every time.
// Within the dome, the brightest BAND is the cloud layer (vr01's row 112 is
// #b3dffa, nearly white), which is not what distance haze should be. And the
// dome's lower half is a second, darker band below the horizon, which drags a
// whole-texture mean away from the sky you are standing under.
//
// The check is the hand-tuned reference: this returns #57b3ec for vr01, against
// the #60baf2 that was picked by eye in bombhei-map.json. Nine, seven and six
// out of 255 — the same sky.
func skyHorizon(ls *sm64ds.LevelSet, ext, path string) (string, error) {
	m, err := sm64ds.LoadBMD(filepath.Join(ext, "files", filepath.FromSlash(strings.TrimPrefix(path, "/"))))
	if err != nil {
		return "", err
	}
	var dome *image.NRGBA
	for _, t := range m.Texs {
		if t.Img == nil {
			continue
		}
		if dome == nil || t.Img.Bounds().Dx()*t.Img.Bounds().Dy() > dome.Bounds().Dx()*dome.Bounds().Dy() {
			dome = t.Img
		}
	}
	if dome == nil {
		return "", fmt.Errorf("no texture in %s", path)
	}
	b := dome.Bounds()
	var r, g, bl, n float64
	for y := b.Min.Y; y < b.Min.Y+b.Dy()/2; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := dome.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			r, g, bl, n = r+float64(c.R), g+float64(c.G), bl+float64(c.B), n+1
		}
	}
	if n == 0 {
		return "", fmt.Errorf("%s: empty texture", path)
	}
	return fmt.Sprintf("#%02x%02x%02x", int(r/n+0.5), int(g/n+0.5), int(bl/n+0.5)), nil
}

// stageExtent returns the level mesh's bounding box in stage-GLB units.
func stageExtent(ls *sm64ds.LevelSet, ext, path string) (lo, hi [3]float64, ok bool) {
	m, err := sm64ds.LoadBMD(filepath.Join(ext, "files", filepath.FromSlash(strings.TrimPrefix(path, "/"))))
	if err != nil {
		return lo, hi, false
	}
	lo = [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi = [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, tris := range m.ByMat {
		for _, t := range tris {
			for _, v := range t.V {
				for k, c := range [3]float64{v.X, v.Y, v.Z} {
					if c < lo[k] {
						lo[k] = c
					}
					if c > hi[k] {
						hi[k] = c
					}
				}
			}
		}
	}
	return lo, hi, !math.IsInf(lo[0], 1)
}

// groundUnder drops a ray onto the level's collision, in world units — the
// game's own ground query (Part VI section 5).
func groundUnder(k *sm64ds.KCL, x, y, z float64) (float64, bool) {
	if k == nil {
		return 0, false
	}
	const floorBelow = -0x80000000
	h, ok := k.RaycastDown(int32(x*4096), int32(y*4096), int32(z*4096), floorBelow, nil, 1)
	if !ok {
		return 0, false
	}
	return float64(h.Y) / 4096, true
}

func loadKCL(ls *sm64ds.LevelSet, ext, stem string) (*sm64ds.KCL, error) {
	var found string
	filepath.Walk(filepath.Join(ext, "files"), func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.TrimSuffix(filepath.Base(p), ".kcl") == stem &&
			strings.HasSuffix(p, ".kcl") {
			found = p
		}
		return nil
	})
	if found == "" {
		return nil, fmt.Errorf("no .kcl named %q", stem)
	}
	data, err := os.ReadFile(found)
	if err != nil {
		return nil, err
	}
	if len(data) > 4 && string(data[:4]) == "LZ77" {
		data = nds.Decompress(data[4:])
	}
	return sm64ds.ParseKCL(data)
}

// writeIndex rewrites the viewer's one-fetch list of which (game, asset) pairs
// have a preset, keeping every other game's entries.
func writeIndex(path string, mine []string) error {
	var doc struct {
		Format  string   `json:"format"`
		Version int      `json:"version"`
		Presets []string `json:"presets"`
	}
	if buf, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(buf, &doc); err != nil {
			return err
		}
	}
	if doc.Format == "" {
		doc.Format, doc.Version = "retro-x-vr", 1
	}
	keep := map[string]bool{}
	for _, p := range doc.Presets {
		if !strings.HasPrefix(p, "super-mario-64-ds/") {
			keep[p] = true
		}
	}
	for _, p := range mine {
		keep[p] = true
	}
	doc.Presets = doc.Presets[:0]
	for p := range keep {
		doc.Presets = append(doc.Presets, p)
	}
	sort.Strings(doc.Presets)
	buf, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(buf, '\n'), 0o644)
}

func slugify(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		default:
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}

func r3(v float64) float64 { return math.Round(v*1000) / 1000 }
func r4(v float64) float64 { return math.Round(v*10000) / 10000 }

