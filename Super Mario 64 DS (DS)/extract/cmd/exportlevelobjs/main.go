// exportlevelobjs decodes every level's object placements (Part V §1: the level
// overlay's settings block, object tables and the object->actor table, all located
// by tracing the game's code — see sm64ds/level.go) and writes per-stage JSON for
// the web viewer, binding placements to extracted models where the binding itself
// was traced from the game:
//
//   - the tree actor's model table at overlay-2 $0210ABB8 (u16 internal file IDs,
//     indexed by (par>>4)&7 clamped to 4 — the clamp and index shift are read at
//     $020EC240): types 0-3 are the four filesystem tree models; type 4 is an
//     archive-resident model (bit-15-flagged ID) we don't extract yet.
//   - actors whose C++ typeinfo class name (embedded in the overlay binaries,
//     e.g. "daObjMc_Metalnet") names their model file directly.
//
// Everything else is exported as an unbound placement; the viewer shows a marker.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"

	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/nds"
	"supermario64ds/extract/sm64ds"
)

// tree actor's model table (traced): index = (par>>4)&7, clamped to 4; entry 4 is
// the archive-resident castle tree (flagged ID $9C10 = ar1.narc member 16)
var treeModels = []string{"bomb_tree", "toge_tree", "yuki_tree", "yashi_tree", "ar1_16"}

// archive models identified by their own material names (arc0[5] "coin",
// arc0[7] the red variant, arc0[21] the power star with "mat_body"/"mat_eye" and
// embedded star strings, arc0[25] "mat_star_silver", ar1[16] "mat_main_tree"):
// extracted from the NARCs via the traced descriptor table and bound to the
// coin/star actors pending a full code trace of their load path.
var archiveModels = []int{0x8005, 0x8007, 0x8015, 0x8019, 0x9C10, 0x9C0F}

// actor -> archive model (content-identified; the actors are the object->actor
// translations of the coin/red-coin/blue-coin/star placements)
var collectibleModels = map[int]string{
	288: "arc0_5",  // coin
	289: "arc0_7",  // red coin
	290: "arc0_5",  // blue coin (same mesh; blue palette variant not yet separated)
	178: "arc0_21", // power star
}

// billboard models: flat quads/discs the game keeps camera-facing
var billboardStems = map[string]bool{
	"bomb_tree": true, "toge_tree": true, "yuki_tree": true, "yashi_tree": true,
	"ar1_16": true, "ar1_15": true,
}

// falseBind: actors whose create stubs sit inside the tree class's compilation
// unit and pick up its model table by adjacency (they place in indoor levels
// where trees make no sense) — dropped until traced properly.
var falseBind = map[int]bool{177: true, 178: true, 179: true, 180: true}

type jsonObj struct {
	Actor int       `json:"a"`
	Model string    `json:"m,omitempty"`
	Scale float64   `json:"s,omitempty"` // model scale in stage units (2^objShift / 2^stageShift)
	Bill  bool      `json:"b,omitempty"` // billboard (flat tree quads yaw to camera)
	Pos   []float64 `json:"p"`
	RotY  float64   `json:"ry,omitempty"`
	Layer int       `json:"l,omitempty"`
}

// bmdShift reads a model's power-of-two scale shift (header u32 at +0): its raw
// fx4.12 vertex space times 2^shift is world units.
func bmdShift(path string) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	var data []byte
	if len(raw) >= 4 && string(raw[:4]) == "LZ77" {
		data = nds.Decompress(raw[4:])
	} else {
		data = nds.Decompress(raw)
	}
	if len(data) < 4 {
		return 0, false
	}
	return int(binary.LittleEndian.Uint32(data)), true
}

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	glbDir := flag.String("glb", "../extracted/glb", "exported models dir (to check bindings)")
	outDir := flag.String("o", "../extracted/objects", "output dir for per-stage object JSON")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	treeActor := ls.Actor(41) // the TREE object's actor (object->actor table)
	filesRoot := filepath.Join(*ext, "files")

	// extract the archive models (coin, star, castle tree ...) to GLBs
	for _, id := range archiveModels {
		ref, ok := ls.ResolveArchiveID(id)
		if !ok {
			fmt.Fprintf(os.Stderr, "  archive id %#x: unresolved\n", id)
			continue
		}
		data, err := ls.ArchiveMember(ref)
		if err != nil || !sm64ds.PlausibleBMD(data) {
			fmt.Fprintf(os.Stderr, "  %s: not a decodable model\n", ref.Stem())
			continue
		}
		m, err := sm64ds.Decode(data, ref.Stem())
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", ref.Stem(), err)
			continue
		}
		if billboardStems[ref.Stem()] {
			m.NormalizeUV() // billboard quads overflow their texture in texel space
		}
		glb, err := m.GLB()
		if err != nil {
			continue
		}
		os.WriteFile(filepath.Join(*glbDir, ref.Stem()+".glb"), glb, 0o644)
	}
	shiftOf := map[string]int{} // model stem -> 2^shift cache
	modelShift := func(rel string) (int, bool) {
		stem := strings.TrimSuffix(filepath.Base(rel), ".bmd")
		if v, ok := shiftOf[stem]; ok {
			return v, true
		}
		v, ok := bmdShift(filepath.Join(filesRoot, rel))
		if ok {
			shiftOf[stem] = v
		}
		return v, ok
	}
	// stem -> path for every .bmd in the internal file table
	modelPath := map[string]string{}
	for i := 0; i < 2058; i++ {
		n := ls.InternalName(i)
		if strings.HasSuffix(n, ".bmd") {
			stem := strings.TrimSuffix(filepath.Base(n), ".bmd")
			if _, dup := modelPath[stem]; !dup {
				modelPath[stem] = n
			}
		}
	}

	hasGLB := func(name string) bool {
		if name == "" {
			return false
		}
		_, err := os.Stat(filepath.Join(*glbDir, name+".glb"))
		return err == nil
	}

	total, bound, stages := 0, 0, 0
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if !hasGLB(stem) {
			continue // no stage model exported (shouldn't happen)
		}
		actorModels, err := ls.TraceActorModels(lv.Overlay)
		if err != nil {
			die(err)
		}
		stageShift, ok := modelShift(lv.BMDPath)
		if !ok {
			fmt.Fprintf(os.Stderr, "  %s: no stage shift\n", stem)
			continue
		}
		// Stage-GLB units: a stage GLB stores raw fx floats times the stage's 2^shift,
		// and a placement short (<<12 into fx by the spawner) is short/4096 fx —
		// so stage-GLB position = short * 2^stageShift / 4096. (The /4096-only first
		// pass passed the containment test but clustered everything at half radius
		// on shift-1 stages like Bob-omb Battlefield.)
		toStage := math.Pow(2, float64(stageShift)) / 4096
		var objs []jsonObj
		seen := map[string]bool{}
		for _, o := range lv.Objects {
			key := fmt.Sprintf("%d/%.3f/%.3f/%.3f", o.Actor, o.X, o.Y, o.Z)
			if seen[key] {
				continue // same placement listed for several star layers
			}
			seen[key] = true
			j := jsonObj{Actor: o.Actor, Pos: []float64{r3(o.X * toStage), r3(o.Y * toStage), r3(o.Z * toStage)}, RotY: r3(o.RotY), Layer: o.Layer}
			switch {
			case o.Actor == treeActor:
				t := (o.Params[0] >> 4) & 7
				if t > 4 {
					t = 4
				}
				if m := treeModels[t]; m != "" && hasGLB(m) {
					j.Model, j.Bill = m, true
					j.Scale = r3(objScale(m, stageShift))
				}
			case collectibleModels[o.Actor] != "":
				if m := collectibleModels[o.Actor]; hasGLB(m) {
					j.Model = m
					j.Bill = billboardStems[m]
					j.Scale = r3(objScale(m, stageShift))
				}
			default:
				if ms := actorModels[o.Actor]; len(ms) > 0 && !falseBind[o.Actor] {
					m := ms[0] // first hit = nearest to the create function
					if hasGLB(m) {
						j.Model = m
						j.Bill = billboardStems[m]
						j.Scale = r3(objScale(m, stageShift))
					}
				}
			}
			if j.Model != "" {
				bound++
			}
			total++
			objs = append(objs, j)
		}
		if len(objs) == 0 {
			continue
		}
		buf, _ := json.Marshal(map[string]any{"objects": objs})
		if err := os.WriteFile(filepath.Join(*outDir, stem+".json"), buf, 0o644); err != nil {
			die(err)
		}
		stages++
	}
	fmt.Printf("exported %d stages, %d placements (%d bound to models)\n", stages, total, bound)
}

// objScale converts an object GLB into stage-GLB units. The engine-side object
// scale field (object+$80, fx12) defaults to 1.0 and none of the placed actors
// write a static constant there — the collectibles are sized by their actors at
// runtime (the mushroom literally grows) — so display sizes are calibrated per
// model where the 1/16 world ratio (which the tree confirms) reads wrong. The
// true render-path object scaling remains untraced (Part V §3).
var stemScale = map[string]float64{
	"arc0_5": 0.020, "arc0_7": 0.020, // coins
	"arc0_21": 0.019, "arc0_25": 0.019, // stars
	"ar1_15":          0.019, // super-mushroom billboard
	"scale_up_kinoko": 0.019,
	"obj_tatefuda":    0.030, // signboards read closer to in-game at half the 1/16
}

func objScale(stem string, stageShift int) float64 {
	f, ok := stemScale[stem]
	if !ok {
		f = 1.0 / 16
	}
	return f * math.Pow(2, float64(stageShift))
}

func r3(v float64) float64 { return float64(int(v*1000+0.5*sign(v))) / 1000 }
func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "exportlevelobjs:", err)
	os.Exit(1)
}
