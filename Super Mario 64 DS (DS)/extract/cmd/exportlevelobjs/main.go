// exportlevelobjs decodes every level's object placements (Part V §1: the level
// overlay's settings block, object tables and the object->actor table, all located
// by tracing the game's code — see sm64ds/level.go) and writes per-stage JSON for
// the web viewer.
//
// Placements bind to models through the ACTOR-BINDING ORACLE's table
// (extracted/actorbind.json, written by cmd/actororacle — see sm64ds/oracle.go):
// each actor's create/init code was RUN on the tools/arm CPU with the game's
// loader trapped, so a binding is what the actor's own code loaded or built its
// render object on, per parameter set. No pattern matching, no special cases:
// an actor without a binding provably loads nothing in create/init.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"supermario64ds/extract/sm64ds"
)

// Binding mirrors cmd/actororacle's table entry.
type Binding struct {
	Params    [3]int            `json:"params"`
	Config    int               `json:"config"`
	Models    []string          `json:"models,omitempty"`
	Clips     []string          `json:"clips,omitempty"`
	KCL       []string          `json:"kcl,omitempty"`
	Colliders []sm64ds.Collider `json:"colliders,omitempty"`
	Notes     []string          `json:"notes,omitempty"`
}

type jsonObj struct {
	Actor int    `json:"a"`
	Model string `json:"m,omitempty"`
	// collision: the .kcl the actor's own collider walks (collision/<stem>.glb)
	// plus its registered transform — 9 unitless rotation/scale entries (the
	// MtxFx43's 3x3, row-vector order) and a stage-unit translation, captured
	// from the collider object at +$134 by the oracle (origin spawn, so the
	// placement pose composes on top). cy = the SclY class's local-Y scale.
	Col   string    `json:"c,omitempty"`
	ColM  []float64 `json:"cm,omitempty"`
	ColSY float64   `json:"cy,omitempty"`
	Scale float64   `json:"s,omitempty"` // display scale: 1/125 (engine NULL-scale-vector size, Part V §5)
	Bill  bool      `json:"b,omitempty"` // billboard (bone flag +$3C bit 0 set on the model's bones)
	Pos   []float64 `json:"p"`
	RotY  float64   `json:"ry,omitempty"`
	Layer int       `json:"l,omitempty"`
	// Txt: the signpost's in-game text (actor 184; par1 is an external message
	// ID resolved through the $0208EEEC range table — see sm64ds.MsgIndex —
	// into data/message/msg_data_eng.bin).
	Txt string `json:"txt,omitempty"`
}

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	glbDir := flag.String("glb", "../extracted/glb", "exported models dir (to check bindings)")
	bindPath := flag.String("bind", "../extracted/actorbind.json", "oracle binding table (cmd/actororacle)")
	outDir := flag.String("o", "../extracted/objects", "output dir for per-stage object JSON")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	var bindings map[int][]Binding
	buf, err := os.ReadFile(*bindPath)
	if err != nil {
		die(fmt.Errorf("%w (run cmd/actororacle first)", err))
	}
	if err := json.Unmarshal(buf, &bindings); err != nil {
		die(err)
	}

	// modelFor picks a placement's model: the oracle ran every distinct
	// parameter combination the levels place, so an exact match exists for any
	// bound actor; the first model of the matching run is the actor's own
	// display binding (see Oracle.Models).
	modelFor := func(actor int, par [3]int) string {
		var loose string
		for _, b := range bindings[actor] {
			if len(b.Models) == 0 {
				continue
			}
			if b.Params == par {
				return b.Models[0]
			}
			if loose == "" && b.Params[0] == par[0] {
				loose = b.Models[0]
			}
		}
		return loose
	}

	// colFor picks a placement's collider the same way: the first collider the
	// actor's own create/init/first-step registered under these parameters.
	colFor := func(actor int, par [3]int) *sm64ds.Collider {
		var loose *sm64ds.Collider
		for i := range bindings[actor] {
			b := &bindings[actor][i]
			if len(b.Colliders) == 0 || b.Colliders[0].KCL == "" {
				continue
			}
			if b.Params == par {
				return &b.Colliders[0]
			}
			if loose == nil && b.Params[0] == par[0] {
				loose = &b.Colliders[0]
			}
		}
		return loose
	}

	// extract archive-member models the bindings reference (arcN_M stems have
	// no filesystem file; decode straight from the NARC)
	extractArchiveGLBs(ls, bindings, *glbDir)

	hasGLB := func(name string) bool {
		if name == "" {
			return false
		}
		_, err := os.Stat(filepath.Join(*glbDir, name+".glb"))
		return err == nil
	}
	colDir := filepath.Join(filepath.Dir(*glbDir), "collision")
	hasCol := func(name string) bool {
		if name == "" {
			return false
		}
		_, err := os.Stat(filepath.Join(colDir, name+".glb"))
		return err == nil
	}
	// billboard = the model's own bone flag (+$3C bit 0): the engine billboards
	// per BONE, so the whole-model flag applies only to models that ARE a single
	// flagged bone (tree quads, clouds, number sprites, the chain link).
	// Multi-bone models with a flagged part (the bob-omb's body_bill) carry the
	// flag as glTF node extras in their skinned export and the viewer orients
	// just that bone.
	bill := billboardChecker(ls, *ext)

	// The signposts' text: actor 184 carries its message as par1, an external
	// message ID the game resolves through $020B8EC0's range table (see
	// sm64ds.MsgIndex); the text itself is the English BMG message file.
	msgs, err := sm64ds.LoadBMG(filepath.Join(*ext, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		die(err)
	}
	const signpostActor = 184
	signText := func(o sm64ds.LevelObject) string {
		if o.Actor != signpostActor {
			return ""
		}
		if idx := ls.MsgIndex(o.Params[0]); idx >= 0 && idx < len(msgs) {
			return msgs[idx]
		}
		return ""
	}

	// objScale is the display scale for an object instance, from the traced engine
	// transform (Part V §5): object models draw under MTX_SCALE(2^shift) — already
	// baked into the exported GLB — while the stage draws under an extra uniform
	// 125.0 ($020755D4), so in stage-GLB units an object drawn with a NULL scale
	// vector (the common case) shows at its baked GLB size / 125.
	const objScale = 1.0 / 125

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
		// Placement shorts are fx world coordinates (the spawner shifts them <<12 —
		// traced at $020FE8AC/$020FE960). The engine renders in world/8 units
		// (ASR #3 at every world->render seam) and draws the stage under
		// MTX_SCALE(2^stageShift) * MTX_SCALE(125.0), so one stage-vertex unit is
		// 2^stageShift * 1000 world-fx units. In stage-GLB units (the export bakes
		// 2^stageShift onto the vertices) a placement is short / 1000, independent
		// of the stage shift (Part V §5).
		const toStage = 1.0 / 1000
		var objs []jsonObj
		seen := map[string]bool{}
		for _, o := range lv.Objects {
			key := fmt.Sprintf("%d/%.3f/%.3f/%.3f", o.Actor, o.X, o.Y, o.Z)
			if seen[key] {
				continue // same placement listed for several star layers
			}
			seen[key] = true
			fill := func(j *jsonObj, actor int, par [3]int) {
				if m := modelFor(actor, par); m != "" && hasGLB(m) {
					j.Model = m
					j.Bill = bill(m)
					j.Scale = objScale
				}
				if c := colFor(actor, par); c != nil && hasCol(c.KCL) {
					j.Col = c.KCL
					if c.Class != "Kc" { // Mbg classes carry their own transform
						cm := make([]float64, 12)
						identity := true
						for k := 0; k < 9; k++ {
							cm[k] = r5(float64(c.Mtx[k]) / 4096)
							want := 0.0
							if k%4 == 0 {
								want = 1
							}
							if cm[k] != want {
								identity = false
							}
						}
						for k := 0; k < 3; k++ {
							cm[9+k] = r5(float64(c.Mtx[9+k]) / 4096 * toStage)
							if cm[9+k] != 0 {
								identity = false
							}
						}
						if !identity {
							j.ColM = cm
						}
						if c.ScaleY != 0 && c.ScaleY != 0x1000 {
							j.ColSY = r5(float64(c.ScaleY) / 4096)
						}
					}
				}
			}
			j := jsonObj{Actor: o.Actor, Pos: []float64{r3(o.X * toStage), r3(o.Y * toStage), r3(o.Z * toStage)}, RotY: r3(o.RotY), Layer: o.Layer, Txt: signText(o)}
			fill(&j, o.Actor, o.Params)
			if j.Model != "" {
				bound++
			}
			total++
			objs = append(objs, j)
			// The chain chomp's stake: daWanwan_c's init spawns a pile
			// (actor 27, param $11) at its own position — $02112C6C:
			// `MOV r0,#27; MOV r1,#$11; ADD r2,actor,#$5C (own pos);
			// BL $02010E2C` — keeps it at +$608 and marks the pile's
			// chomp-stake mode byte (+$320). The actor oracle can't see
			// spawned children (the spawn dies in its bare environment),
			// so the traced child is emitted here.
			if o.Actor == 219 {
				s := jsonObj{Actor: 27, Pos: j.Pos, Layer: o.Layer}
				fill(&s, 27, [3]int{65535, 0, 0})
				objs = append(objs, s)
			}
		}
		if len(objs) == 0 {
			continue
		}
		// Mario's spawn: the level's first type-1 entrance.
		var mario map[string]interface{}
		if len(lv.Entrances) > 0 {
			e := lv.Entrances[0]
			mario = map[string]interface{}{
				"p":  []float64{r3(e.X * toStage), r3(e.Y * toStage), r3(e.Z * toStage)},
				"ry": r3(e.RotY),
			}
		}
		out := map[string]any{"objects": objs}
		// the level's own collision map (Part VI), for the viewer's red overlay
		if kstem := strings.TrimSuffix(filepath.Base(lv.KCLPath), ".kcl"); hasCol(kstem) {
			out["col"] = kstem
		}
		if mario != nil {
			out["mario"] = mario
		}
		// the level's skybox (drawn camera-centred at the engine's NULL-scale
		// object size, GLB/125 — see sm64ds/level.go)
		if sky := strings.TrimSuffix(filepath.Base(lv.SkyPath), ".bmd"); lv.SkyPath != "" && hasGLB(sky) {
			out["sky"] = sky
		}
		buf, _ := json.Marshal(out)
		if err := os.WriteFile(filepath.Join(*outDir, stem+".json"), buf, 0o644); err != nil {
			die(err)
		}
		stages++
	}
	fmt.Printf("exported %d stages, %d placements (%d bound to models)\n", stages, total, bound)
}

// extractArchiveGLBs decodes every archive member the bindings name (arcN_M)
// into a GLB next to the filesystem models.
func extractArchiveGLBs(ls *sm64ds.LevelSet, bindings map[int][]Binding, glbDir string) {
	done := map[string]bool{}
	for _, bs := range bindings {
		for _, b := range bs {
			for _, stem := range b.Models {
				i := strings.LastIndexByte(stem, '_')
				if i < 0 || done[stem] {
					continue
				}
				done[stem] = true
				ref, ok := archiveRefByStem(ls, stem)
				if !ok {
					continue
				}
				data, err := ls.ArchiveMember(ref)
				if err != nil || !sm64ds.PlausibleBMD(data) {
					continue
				}
				m, err := sm64ds.Decode(data, stem)
				if err != nil {
					continue
				}
				glb, err := m.GLB()
				if err != nil {
					continue
				}
				os.WriteFile(filepath.Join(glbDir, stem+".glb"), glb, 0o644)
			}
		}
	}
}

// archiveRefByStem parses "arc0_5" back into an archive reference.
func archiveRefByStem(ls *sm64ds.LevelSet, stem string) (sm64ds.ArchiveRef, bool) {
	i := strings.LastIndexByte(stem, '_')
	if i < 0 {
		return sm64ds.ArchiveRef{}, false
	}
	var member int
	if _, err := fmt.Sscanf(stem[i+1:], "%d", &member); err != nil {
		return sm64ds.ArchiveRef{}, false
	}
	name := stem[:i]
	// only names that are actually archives resolve (filesystem stems with
	// underscores fall through)
	for _, arc := range []string{"arc0", "ar1", "c2d", "cee", "cef", "ceg", "cei", "ces", "en1", "vs1", "vs2", "vs3", "vs4"} {
		if name == arc {
			return sm64ds.ArchiveRef{Archive: name, Member: member}, true
		}
	}
	return sm64ds.ArchiveRef{}, false
}

// billboardChecker reports whether a model's bones carry the camera-facing
// flag, decoding each stem once (filesystem or archive member).
func billboardChecker(ls *sm64ds.LevelSet, extDir string) func(stem string) bool {
	// stem -> path for every .bmd in the internal file table
	path := map[string]string{}
	for i := 0; i < 2058; i++ {
		n := ls.InternalName(i)
		if strings.HasSuffix(n, ".bmd") {
			s := strings.TrimSuffix(filepath.Base(n), ".bmd")
			if _, dup := path[s]; !dup {
				path[s] = n
			}
		}
	}
	cache := map[string]bool{}
	return func(stem string) bool {
		if v, ok := cache[stem]; ok {
			return v
		}
		var m *sm64ds.Model
		if p, ok := path[stem]; ok {
			m, _ = sm64ds.LoadBMD(filepath.Join(extDir, "files", filepath.FromSlash(strings.TrimPrefix(p, "/"))))
		} else if ref, ok := archiveRefByStem(ls, stem); ok {
			if data, err := ls.ArchiveMember(ref); err == nil && sm64ds.PlausibleBMD(data) {
				m, _ = sm64ds.Decode(data, stem)
			}
		}
		v := m != nil && len(m.Skel) == 1 && m.Skel[0].Billboard
		cache[stem] = v
		return v
	}
}

func r3(v float64) float64 { return float64(int(v*1000+0.5*sign(v))) / 1000 }
func r5(v float64) float64 { return float64(int(v*100000+0.5*sign(v))) / 100000 }
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
