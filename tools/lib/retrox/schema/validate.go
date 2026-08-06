package schema

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Issue is one validation finding. Level is "error" or "warn".
type Issue struct {
	Level string
	Path  string // file (and optionally element) the issue is about
	Msg   string
}

func (i Issue) String() string { return fmt.Sprintf("%s: %s: %s", i.Level, i.Path, i.Msg) }

// HasErrors reports whether any issue is an error (not just a warning).
func HasErrors(issues []Issue) bool {
	for _, i := range issues {
		if i.Level == "error" {
			return true
		}
	}
	return false
}

type checker struct {
	fsys   fs.FS
	issues []Issue
}

func (c *checker) errf(p, format string, args ...any) {
	c.issues = append(c.issues, Issue{"error", p, fmt.Sprintf(format, args...)})
}

func (c *checker) warnf(p, format string, args ...any) {
	c.issues = append(c.issues, Issue{"warn", p, fmt.Sprintf(format, args...)})
}

// ref resolves a file reference relative to the directory of the referencing
// document and checks it stays inside the tree and exists. Returns the
// resolved path ("" when invalid).
func (c *checker) ref(docPath, r string) string {
	where := docPath + ": " + r
	if r == "" {
		c.errf(where, "empty file reference")
		return ""
	}
	if strings.HasPrefix(r, "/") || strings.Contains(r, "..") {
		c.errf(where, "file reference must be relative and must not contain '..'")
		return ""
	}
	full := path.Join(path.Dir(docPath), r)
	if _, err := fs.Stat(c.fsys, full); err != nil {
		c.errf(where, "referenced file does not exist (%s)", full)
		return ""
	}
	return full
}

func loadJSON[T any](fsys fs.FS, p string) (*T, error) {
	b, err := fs.ReadFile(fsys, p)
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *checker) pngSize(p string) (w, h int, ok bool) {
	f, err := c.fsys.Open(p)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		c.errf(p, "not a decodable PNG: %v", err)
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

// ValidateTree validates a whole publication rooted at fsys (must contain
// index.json).
func ValidateTree(fsys fs.FS) []Issue {
	c := &checker{fsys: fsys}
	idx, err := loadJSON[Index](fsys, "index.json")
	if err != nil {
		c.errf("index.json", "%v", err)
		return c.issues
	}
	if err := idx.Check(); err != nil {
		c.errf("index.json", "%v", err)
	}
	seen := map[string]bool{}
	for _, g := range idx.Games {
		if seen[g] {
			c.errf("index.json", "duplicate game id %q", g)
			continue
		}
		seen[g] = true
		sub, err := fs.Sub(fsys, g)
		if err != nil {
			c.errf("index.json", "game %q: %v", g, err)
			continue
		}
		if _, err := fs.Stat(sub, "manifest.json"); err != nil {
			c.errf("index.json", "game %q has no manifest.json", g)
			continue
		}
		for _, iss := range ValidateGame(sub, g) {
			iss.Path = g + "/" + iss.Path
			c.issues = append(c.issues, iss)
		}
	}
	return c.issues
}

// ValidateGame validates one game tree (fsys rooted at the game directory).
// wantID, when non-empty, must match the manifest id.
func ValidateGame(fsys fs.FS, wantID string) []Issue {
	c := &checker{fsys: fsys}
	const mp = "manifest.json"
	man, err := loadJSON[Manifest](fsys, mp)
	if err != nil {
		c.errf(mp, "%v", err)
		return c.issues
	}
	if err := man.Check(); err != nil {
		c.errf(mp, "%v", err)
	}
	if wantID != "" && man.ID != wantID {
		c.errf(mp, "manifest id %q does not match directory %q", man.ID, wantID)
	}
	if man.Title == "" {
		c.errf(mp, "title is required")
	}
	if man.Display.Native.W <= 0 || man.Display.Native.H <= 0 {
		c.errf(mp, "display.native must be a positive resolution")
	}
	if man.Display.TickHz <= 0 {
		c.errf(mp, "display.tickHz must be positive")
	}
	if man.Logo != "" {
		c.ref(mp, man.Logo)
	}
	docIDs := map[string]bool{}
	for _, d := range man.Docs {
		if docIDs[d.ID] {
			c.errf(mp, "duplicate doc id %q", d.ID)
		}
		docIDs[d.ID] = true
		c.ref(mp, d.File)
	}

	// Asset index passes: uniqueness, categories, file existence.
	assets := map[string]*Asset{}
	catOK := map[string]bool{}
	for _, cat := range Categories {
		catOK[cat] = true
	}
	for i := range man.Assets {
		a := &man.Assets[i]
		where := mp + ": asset " + a.ID
		if a.ID == "" {
			c.errf(mp, "asset #%d has no id", i)
			continue
		}
		if assets[a.ID] != nil {
			c.errf(where, "duplicate asset id")
			continue
		}
		assets[a.ID] = a
		if !catOK[a.Category] {
			c.errf(where, "unknown category %q", a.Category)
		}
		if a.Name == "" {
			c.warnf(where, "asset has no name")
		}
		c.ref(mp, a.File)
	}
	for _, a := range assets {
		for _, r := range a.Related {
			if assets[r] == nil {
				c.errf(mp+": asset "+a.ID, "related asset %q does not exist", r)
			}
		}
	}

	// Load all object documents (levels reference their animations).
	objects := map[string]*Object{}
	for _, a := range assets {
		if a.Category != CategoryObject {
			continue
		}
		doc, err := loadJSON[Object](fsys, a.File)
		if err != nil {
			c.errf(a.File, "%v", err)
			continue
		}
		if err := doc.Check(); err != nil {
			c.errf(a.File, "%v", err)
		}
		objects[a.ID] = doc
		c.checkObject(a.File, doc, assets)
	}

	// Levels.
	for _, a := range assets {
		if a.Category != CategoryLevel {
			continue
		}
		doc, err := loadJSON[Level](fsys, a.File)
		if err != nil {
			c.errf(a.File, "%v", err)
			continue
		}
		if err := doc.Check(); err != nil {
			c.errf(a.File, "%v", err)
		}
		c.checkLevel(a.File, doc, assets, objects)
	}

	// Media (rule 12: declared picture dimensions match the PNG).
	for _, a := range assets {
		if a.Category != CategoryPicture {
			continue
		}
		if w, h, ok := c.pngSize(a.File); ok && (a.W != 0 || a.H != 0) {
			if a.W != w || a.H != h {
				c.errf(mp+": asset "+a.ID, "declared %dx%d but PNG is %dx%d", a.W, a.H, w, h)
			}
		}
	}
	return c.issues
}

// ---------------------------------------------------------------------------

func (c *checker) checkLevel(p string, l *Level, assets map[string]*Asset, objects map[string]*Object) {
	switch l.Type {
	case LevelTilemap:
		if l.Tilemap == nil {
			c.errf(p, "type tilemap requires a tilemap body")
		}
		if l.Scene != nil {
			c.errf(p, "tilemap level must not carry a scene body")
		}
	case LevelScene3D:
		if l.Scene == nil {
			c.errf(p, "type scene3d requires a scene body")
		}
		if l.Tilemap != nil {
			c.errf(p, "scene3d level must not carry a tilemap body")
		}
		if l.Camera == nil {
			c.errf(p, "scene3d level requires a camera block")
		}
	default:
		c.errf(p, "unknown level type %q", l.Type)
	}

	if l.Music != "" {
		if a := assets[l.Music]; a == nil || a.Category != CategoryMusic {
			c.errf(p, "music %q is not a music asset", l.Music)
		}
	}
	if l.PixelGrid != nil && l.PixelGrid.Lines <= 0 {
		c.errf(p, "pixelGrid.lines must be positive")
	}

	variants := map[string]bool{}
	defaults := 0
	for _, v := range l.Variants {
		if variants[v.ID] {
			c.errf(p, "duplicate variant id %q", v.ID)
		}
		variants[v.ID] = true
		if v.Default {
			defaults++
		}
	}
	if defaults > 1 {
		c.errf(p, "more than one default variant")
	}
	checkVariants := func(where string, ids []string) {
		for _, id := range ids {
			if !variants[id] {
				c.errf(where, "variant %q is not declared by the level", id)
			}
		}
	}

	routes := map[string]bool{}
	for _, r := range l.Routes {
		where := p + ": route " + r.ID
		if routes[r.ID] {
			c.errf(where, "duplicate route id")
		}
		routes[r.ID] = true
		if len(r.Points) < 2 {
			c.errf(where, "a route needs at least 2 points")
		}
		for _, pt := range r.Points {
			if len(pt) != 2 && len(pt) != 3 {
				c.errf(where, "route points must be [x,y] or [x,y,z]")
				break
			}
		}
	}

	layers := map[string]bool{}
	groupLayers := map[string]string{} // file-less scene layers -> where; cleared as placements name them
	rooms := map[int]bool{}
	if l.Tilemap != nil {
		for i := range l.Tilemap.Layers {
			ly := &l.Tilemap.Layers[i]
			if ly.ID == "" {
				c.errf(p, "tilemap layer #%d has no id", i)
				continue
			}
			if layers[ly.ID] {
				c.errf(p+": layer "+ly.ID, "duplicate layer id")
			}
			layers[ly.ID] = true
		}
	}
	if l.Scene != nil {
		if len(l.Scene.Layers) == 0 && l.Scene.Rooms == nil && len(l.Placements) == 0 {
			c.errf(p, "a scene needs at least one layer, a room graph, or placements")
		}
		exclusiveVisible := map[string]int{}
		for i := range l.Scene.Layers {
			ly := &l.Scene.Layers[i]
			where := p + ": layer " + ly.ID
			if ly.ID == "" {
				c.errf(p, "layer #%d has no id", i)
				continue
			}
			if layers[ly.ID] {
				c.errf(where, "duplicate layer id")
			}
			layers[ly.ID] = true
			// A layer with no file is a placement group (schema.Layer.File):
			// legal, but only if something actually lands in it.
			if ly.File != "" {
				c.ref(p, ly.File)
			} else {
				groupLayers[ly.ID] = where
			}
			switch {
			case ly.Mode == "" || ly.Mode == LayerBase || ly.Mode == LayerToggle:
			case strings.HasPrefix(ly.Mode, "exclusive:"):
				if ly.IsVisible() {
					exclusiveVisible[ly.Mode]++
				}
			default:
				c.errf(where, "unknown layer mode %q", ly.Mode)
			}
			switch ly.Attach {
			case "", "world", "camera", "cameraYaw":
			default:
				c.errf(where, "unknown attach %q", ly.Attach)
			}
		}
		for group, n := range exclusiveVisible {
			if n != 1 {
				c.errf(p, "%s: exactly one layer of an exclusive group must be initially visible (got %d)", group, n)
			}
		}
		if l.Scene.Rooms != nil {
			areaIDs := map[string]bool{}
			for _, a := range l.Scene.Rooms.Areas {
				if areaIDs[a.ID] {
					c.errf(p, "duplicate area id %q", a.ID)
				}
				areaIDs[a.ID] = true
			}
			for _, rm := range l.Scene.Rooms.List {
				where := fmt.Sprintf("%s: room %d", p, rm.ID)
				if rooms[rm.ID] {
					c.errf(where, "duplicate room id")
				}
				rooms[rm.ID] = true
				c.ref(p, rm.File)
				if rm.Area != "" && !areaIDs[rm.Area] {
					c.errf(where, "area %q is not declared", rm.Area)
				}
			}
		}
	}

	if l.Tilemap != nil {
		c.checkTilemap(p, l)
	}
	if l.Collision != nil {
		c.checkCollision(p, l)
	}

	// Placements.
	placements := map[int]*Placement{}
	for i := range l.Placements {
		pl := &l.Placements[i]
		where := fmt.Sprintf("%s: placement %d", p, pl.ID)
		if placements[pl.ID] != nil {
			c.errf(where, "duplicate placement id")
			continue
		}
		placements[pl.ID] = pl

		obj := c.checkObjectRef(where, pl.Object, assets, objects)
		if pl.Anim != "" && obj != nil && findAnim(obj, pl.Anim) == nil {
			c.errf(where, "anim %q does not exist on object %q", pl.Anim, pl.Object)
		}
		if len(pl.Matrix) == 0 {
			want2 := l.Type == LevelTilemap
			switch {
			case len(pl.Pos) == 0:
				c.errf(where, "placement needs pos or matrix")
			case want2 && len(pl.Pos) != 2:
				c.errf(where, "tilemap placement pos must be [x,y]")
			case !want2 && len(pl.Pos) != 3:
				c.errf(where, "scene3d placement pos must be [x,y,z]")
			}
		} else if len(pl.Matrix) != 16 {
			c.errf(where, "matrix must have 16 numbers")
		}
		if n := len(pl.Scale); n != 0 && n != 1 && n != 3 {
			c.errf(where, "scale must be a scalar or [sx,sy,sz]")
		}
		if len(pl.Rot) != 0 && len(pl.Rot) != 3 {
			c.errf(where, "rot must be [rx,ry,rz]")
		}
		if pl.Layer != "" && !layers[pl.Layer] {
			c.errf(where, "layer %q does not exist", pl.Layer)
		}
		delete(groupLayers, pl.Layer)
		if pl.Room != nil && !rooms[*pl.Room] {
			c.errf(where, "room %d does not exist", *pl.Room)
		}
		checkVariants(where, pl.Variants)
		if pl.Collision != nil {
			c.ref(p, pl.Collision.File)
			if n := len(pl.Collision.Matrix); n != 0 && n != 12 {
				c.errf(where, "collision.matrix must have 12 numbers")
			}
		}
		if pl.Route != nil {
			if !routes[pl.Route.ID] {
				c.errf(where, "route %q does not exist", pl.Route.ID)
			}
			if m := pl.Route.Mode; m != "" && m != "loop" && m != "pingpong" {
				c.errf(where, "unknown route mode %q", m)
			}
		}
		if pl.Behavior != nil {
			switch pl.Behavior.Kind {
			case BehaviorSpin, BehaviorFlyer:
			default:
				c.warnf(where, "unknown behavior kind %q (viewers render the object static)", pl.Behavior.Kind)
			}
		}
	}

	// onClick needs the full placement table (targets).
	for i := range l.Placements {
		pl := &l.Placements[i]
		if pl.OnClick == nil {
			continue
		}
		where := fmt.Sprintf("%s: placement %d onClick", p, pl.ID)
		oc := pl.OnClick
		switch oc.Action {
		case ActionAnimate:
			target := pl
			if oc.Target != "" && oc.Target != "self" {
				id, err := strconv.Atoi(oc.Target)
				if err != nil || placements[id] == nil {
					c.errf(where, "target %q is not a placement id", oc.Target)
					target = nil
				} else {
					target = placements[id]
				}
			}
			if oc.Clip == "" {
				c.errf(where, "animate needs a clip")
			} else if target != nil {
				if obj := objects[target.Object]; obj != nil && findAnim(obj, oc.Clip) == nil {
					c.errf(where, "clip %q does not exist on object %q", oc.Clip, target.Object)
				}
			}
			for _, s := range oc.SFX {
				if a := assets[s.ID]; a == nil || a.Category != CategorySFX {
					c.errf(where, "sfx %q is not an sfx asset", s.ID)
				}
			}
		case ActionText:
			if oc.Body == "" {
				c.warnf(where, "text action has no body")
			}
		default:
			c.warnf(where, "unknown action %q (viewers ignore it)", oc.Action)
		}
	}

	// Pools.
	poolIDs := map[string]bool{}
	for i := range l.Pools {
		po := &l.Pools[i]
		where := p + ": pool " + po.ID
		if poolIDs[po.ID] {
			c.errf(where, "duplicate pool id")
		}
		poolIDs[po.ID] = true
		c.checkObjectRef(where, po.Object, assets, objects)
		if po.Count <= 0 {
			c.errf(where, "count must be positive")
		}
		if po.Count > len(po.Candidates) {
			c.warnf(where, "count %d exceeds %d candidates", po.Count, len(po.Candidates))
		}
		for _, cand := range po.Candidates {
			if len(cand) != 2 {
				c.errf(where, "candidates must be [x,y]")
				break
			}
		}
		if po.Anim != "" {
			if obj := objects[po.Object]; obj != nil && findAnim(obj, po.Anim) == nil {
				c.errf(where, "anim %q does not exist on object %q", po.Anim, po.Object)
			}
		}
		checkVariants(where, po.Variants)
	}
	for id, where := range groupLayers {
		c.errf(where, "layer %q has no file and no placement names it", id)
	}

	// Camera block.
	if l.Camera != nil {
		c.checkCamera(p, l.Camera, routes)
	}

	// Cutscene scripts.
	for _, sr := range l.Scripts {
		full := c.ref(p, sr.File)
		if full == "" {
			continue
		}
		doc, err := loadJSON[Script](c.fsys, full)
		if err != nil {
			c.errf(full, "%v", err)
			continue
		}
		if err := doc.Check(); err != nil {
			c.errf(full, "%v", err)
		}
		c.checkScript(full, doc, layers, placements, objects, assetsOnly(assets, CategorySFX))
	}
}

func assetsOnly(assets map[string]*Asset, cat string) map[string]bool {
	out := map[string]bool{}
	for id, a := range assets {
		if a.Category == cat {
			out[id] = true
		}
	}
	return out
}

func (c *checker) checkObjectRef(where, id string, assets map[string]*Asset, objects map[string]*Object) *Object {
	if id == "" {
		c.errf(where, "missing object reference")
		return nil
	}
	if a := assets[id]; a == nil || a.Category != CategoryObject {
		c.errf(where, "object %q is not an object asset", id)
		return nil
	}
	return objects[id]
}

func findAnim(o *Object, id string) *Animation {
	for i := range o.Animations {
		if o.Animations[i].ID == id {
			return &o.Animations[i]
		}
	}
	return nil
}

func (c *checker) checkTilemap(p string, l *Level) {
	t := l.Tilemap
	if t.TileSize <= 0 || t.Width <= 0 || t.Height <= 0 {
		c.errf(p, "tilemap needs positive tileSize/width/height")
		return
	}
	if len(t.Cells) != t.Width*t.Height {
		c.errf(p, "cells has %d entries, want width*height = %d", len(t.Cells), t.Width*t.Height)
	}
	if t.Wrap != "" && t.Wrap != "none" && t.Wrap != "x" {
		c.errf(p, "unknown wrap %q", t.Wrap)
	}

	// Tile capacity from the atlas PNG.
	capacity := -1
	if full := c.ref(p, t.Atlas.File); full != "" && t.Atlas.Cols > 0 {
		if w, h, ok := c.pngSize(full); ok {
			pitch := t.TileSize + 2*t.Atlas.Gutter
			if w%pitch != 0 || h%pitch != 0 {
				c.warnf(p, "atlas %dx%d is not a multiple of the tile pitch %d", w, h, pitch)
			}
			capacity = (w / pitch) * (h / pitch)
			if w/pitch != t.Atlas.Cols {
				c.errf(p, "atlas.cols %d but PNG holds %d tiles per row", t.Atlas.Cols, w/pitch)
			}
		}
	}

	maxCell := 0
	for _, cell := range t.Cells {
		id := cell
		if t.HFlipMask != 0 {
			id = cell &^ t.HFlipMask
		}
		if id > maxCell {
			maxCell = id
		}
	}
	if t.Blocks != nil {
		b := t.Blocks
		if b.Size <= 0 {
			c.errf(p, "blocks.size must be positive")
		}
		if maxCell >= len(b.Tiles) {
			c.errf(p, "cell id %d out of range (%d blocks)", maxCell, len(b.Tiles))
		}
		if len(b.Shapes) != 0 && len(b.Shapes) != len(b.Tiles) {
			c.errf(p, "blocks.shapes has %d entries, want one per block (%d)", len(b.Shapes), len(b.Tiles))
		}
		maxTile := 0
		for i, blk := range b.Tiles {
			if len(blk) != b.Size*b.Size {
				c.errf(p, "block %d has %d tiles, want size*size = %d", i, len(blk), b.Size*b.Size)
				break
			}
			for _, tid := range blk {
				if tid > maxTile {
					maxTile = tid
				}
			}
		}
		if capacity >= 0 && maxTile >= capacity {
			c.errf(p, "block tile id %d out of range (atlas holds %d tiles)", maxTile, capacity)
		}
	} else if capacity >= 0 && maxCell >= capacity {
		c.errf(p, "cell id %d out of range (atlas holds %d tiles)", maxCell, capacity)
	}

	for i, ta := range t.TileAnims {
		for _, frame := range ta.Frames {
			if len(frame) != len(ta.Tiles) {
				c.errf(p, "tileAnim %d: every frame must list %d tiles", i, len(ta.Tiles))
				break
			}
		}
		if ta.PeriodFrames <= 0 {
			c.errf(p, "tileAnim %d: periodFrames must be positive", i)
		}
	}
	for i, ca := range t.CellAnims {
		for _, ph := range ca.Phases {
			if len(ph.Tiles) != ca.TW*ca.TH {
				c.errf(p, "cellAnim %d: a phase must list tw*th = %d tiles", i, ca.TW*ca.TH)
				break
			}
		}
	}
}

func (c *checker) checkCollision(p string, l *Level) {
	col := l.Collision
	switch col.Kind {
	case "grid":
		sub := col.Sub
		if sub == 0 {
			sub = 1
		}
		if len(col.Solid)%(sub*sub) != 0 {
			c.errf(p, "collision.solid length %d is not a multiple of sub*sub = %d", len(col.Solid), sub*sub)
		}
	case "profiles":
		full := c.ref(p, col.File)
		if full == "" {
			return
		}
		doc, err := loadJSON[Shapes](c.fsys, full)
		if err != nil {
			c.errf(full, "%v", err)
			return
		}
		if err := doc.Check(); err != nil {
			c.errf(full, "%v", err)
		}
		if doc.Count != len(doc.Profiles) || doc.Count != len(doc.Angles) {
			c.errf(full, "count %d but %d profiles / %d angles", doc.Count, len(doc.Profiles), len(doc.Angles))
		}
		for i, prof := range doc.Profiles {
			if len(prof) != 32 {
				c.errf(full, "profile %d has %d columns, want 32", i, len(prof))
				break
			}
		}
	default:
		c.errf(p, "unknown collision kind %q", col.Kind)
	}
}

func (c *checker) checkCamera(p string, cam *Camera, routes map[string]bool) {
	switch cam.Mode {
	case "map2d":
	case "orbit", "fly", "ortho", "pan2d":
		if len(cam.Pos) != 3 || len(cam.Target) != 3 {
			c.errf(p, "camera mode %q needs pos and target [x,y,z]", cam.Mode)
		}
		if cam.Mode == "fly" && (cam.Fly == nil || cam.Fly.Speed <= 0) {
			c.errf(p, "fly camera needs fly.speed > 0")
		}
		if cam.Mode == "ortho" && (cam.Ortho == nil || len(cam.Ortho.Dir) != 3) {
			c.errf(p, "ortho camera needs ortho.dir [x,y,z]")
		}
	default:
		c.errf(p, "unknown camera mode %q", cam.Mode)
	}
	if cam.Drive != nil {
		if !routes[cam.Drive.Route] {
			c.errf(p, "drive route %q does not exist", cam.Drive.Route)
		}
		if cam.Drive.Speed <= 0 {
			c.errf(p, "drive.speed must be positive")
		}
	}
}

func (c *checker) checkScript(p string, s *Script, layers map[string]bool, placements map[int]*Placement, objects map[string]*Object, sfx map[string]bool) {
	if s.FPS <= 0 {
		c.errf(p, "fps must be positive")
	}
	if len(s.Shots) == 0 {
		c.errf(p, "a script needs at least one shot")
	}
	shotIDs := map[string]bool{}
	for i := range s.Shots {
		sh := &s.Shots[i]
		where := p + ": shot " + sh.ID
		if shotIDs[sh.ID] {
			c.errf(where, "duplicate shot id")
		}
		shotIDs[sh.ID] = true
		if sh.Frames <= 0 {
			c.errf(where, "frames must be positive")
		}
		for _, ly := range sh.Layers {
			if !layers[ly] {
				c.errf(where, "layer %q does not exist in the owning level", ly)
			}
		}
		for _, a := range sh.Actors {
			pl := placements[a.Placement]
			if pl == nil {
				c.errf(where, "actor placement %d does not exist", a.Placement)
				continue
			}
			if a.Clip != "" {
				if obj := objects[pl.Object]; obj != nil && findAnim(obj, a.Clip) == nil {
					c.errf(where, "actor clip %q does not exist on object %q", a.Clip, pl.Object)
				}
			}
			if n := len(a.Matrix); n != 0 && n != 16 {
				c.errf(where, "actor matrix must have 16 numbers")
			}
		}
		// Camera: inline track or track file; sample count must match frames.
		track := sh.Camera.Track
		frames := sh.Frames
		if sh.Camera.TrackFile != "" {
			if len(track) != 0 {
				c.errf(where, "camera has both track and trackFile")
			}
			if full := c.ref(p, sh.Camera.TrackFile); full != "" {
				doc, err := loadJSON[CameraTrack](c.fsys, full)
				if err != nil {
					c.errf(full, "%v", err)
					continue
				}
				if err := doc.Check(); err != nil {
					c.errf(full, "%v", err)
				}
				if len(doc.Track) != doc.Frames {
					c.errf(full, "track has %d samples, want frames = %d", len(doc.Track), doc.Frames)
				}
				if doc.Frames != frames {
					c.warnf(where, "shot is %d frames but camera track is %d", frames, doc.Frames)
				}
			}
		} else if len(track) != 0 {
			if len(track) != frames {
				c.errf(where, "inline camera track has %d samples, want frames = %d", len(track), frames)
			}
		} else {
			c.errf(where, "shot needs a camera track or trackFile")
		}
		for _, cue := range sh.Sounds {
			if !sfx[cue.SFX] {
				c.errf(where, "sound cue sfx %q is not an sfx asset", cue.SFX)
			}
			if cue.End != 0 && cue.End < cue.Start {
				c.errf(where, "sound cue ends before it starts")
			}
		}
	}
}

// ---------------------------------------------------------------------------

func (c *checker) checkObject(p string, o *Object, assets map[string]*Asset) {
	if o.Name == "" {
		c.warnf(p, "object has no name")
	}
	animIDs := map[string]bool{}
	for i := range o.Animations {
		a := &o.Animations[i]
		if a.ID == "" {
			c.errf(p, "animation #%d has no id", i)
			continue
		}
		if animIDs[a.ID] {
			c.errf(p, "duplicate animation id %q", a.ID)
		}
		animIDs[a.ID] = true
		switch a.Loop {
		case "once", "loop", "pingpong", "hold":
		default:
			c.errf(p+": animation "+a.ID, "unknown loop mode %q", a.Loop)
		}
	}

	switch o.Type {
	case ObjectSprite2D:
		c.checkSpriteAtlas(p, o, assets, false)
	case ObjectBillboard3D:
		if o.Views <= 0 {
			c.errf(p, "billboard3d needs views >= 1")
		}
		if len(o.Size) != 2 {
			c.errf(p, "billboard3d needs size [w,h]")
		}
		switch o.Mode {
		case "", "camera", "yaw":
		default:
			c.errf(p, "unknown billboard mode %q", o.Mode)
		}
		switch o.Blend {
		case "", "opaque", "alpha", "additive":
		default:
			c.errf(p, "unknown blend %q", o.Blend)
		}
		switch o.AnchorMode {
		case "", "center", "bottom":
		default:
			c.errf(p, "unknown anchorMode %q", o.AnchorMode)
		}
		c.checkSpriteAtlas(p, o, assets, true)
	case ObjectModel3D:
		full := c.ref(p, o.Model)
		if full != "" {
			c.checkGLBBindings(p, full, o)
		}
		if o.Billboard != "" && o.Billboard != "yaw" && o.Billboard != "camera" {
			c.errf(p, "unknown model billboard %q", o.Billboard)
		}
		if o.AtlasPicture != "" {
			c.ref(p, o.AtlasPicture)
		}
		if len(o.EnvMap) > 0 {
			if len(o.EnvMap) != 6 {
				c.errf(p, "envMap needs exactly 6 faces (+x,-x,+y,-y,+z,-z), got %d", len(o.EnvMap))
			}
			for _, f := range o.EnvMap {
				c.ref(p, f)
			}
		}
		for _, fb := range o.Flipbooks {
			where := p + ": flipbook " + fb.Material
			for _, tex := range fb.Textures {
				c.ref(p, tex)
			}
			for _, s := range fb.Sequence {
				if s < 0 || s >= len(fb.Textures) {
					c.errf(where, "sequence index %d out of range (%d textures)", s, len(fb.Textures))
					break
				}
			}
			if fb.Step <= 0 {
				c.errf(where, "step must be positive")
			}
		}
	case ObjectWireframe3D:
		if o.Wireframe == nil {
			c.errf(p, "wireframe3d needs a wireframe body")
			return
		}
		w := o.Wireframe
		if len(w.Positions)%3 != 0 {
			c.errf(p, "positions length must be a multiple of 3")
		}
		nv := len(w.Positions) / 3
		if len(w.Faces) != len(w.FaceCenters) {
			c.errf(p, "%d faces but %d faceCenters", len(w.Faces), len(w.FaceCenters))
		}
		for i, e := range w.Edges {
			if len(e) != 4 {
				c.errf(p, "edge %d must be [v0,v1,faceA,faceB]", i)
				break
			}
			if e[0] >= nv || e[1] >= nv || e[0] < 0 || e[1] < 0 {
				c.errf(p, "edge %d vertex index out of range", i)
				break
			}
			// -1 = no face on that side (an open edge, always drawn) — RETROX.md §6.4
			if e[2] >= len(w.Faces) || e[3] >= len(w.Faces) || e[2] < -1 || e[3] < -1 {
				c.errf(p, "edge %d face index out of range", i)
				break
			}
		}
	default:
		c.errf(p, "unknown object type %q", o.Type)
	}

	// Animation events reference sfx assets.
	for i := range o.Animations {
		for _, ev := range o.Animations[i].Events {
			if a := assets[ev.SFX]; a == nil || a.Category != CategorySFX {
				c.errf(p+": animation "+o.Animations[i].ID, "event sfx %q is not an sfx asset", ev.SFX)
			}
		}
	}
}

func (c *checker) checkSpriteAtlas(p string, o *Object, assets map[string]*Asset, billboard bool) {
	if o.Atlas == nil {
		c.errf(p, "%s needs an atlas", o.Type)
		return
	}
	if len(o.Animations) == 0 {
		c.errf(p, "%s needs at least one animation", o.Type)
	}
	full := c.ref(p, o.Atlas.File)
	if full == "" || o.Atlas.CellW <= 0 || o.Atlas.CellH <= 0 {
		if o.Atlas.CellW <= 0 || o.Atlas.CellH <= 0 {
			c.errf(p, "atlas needs positive cellW/cellH")
		}
		return
	}
	w, h, ok := c.pngSize(full)
	if !ok {
		return
	}
	if w%o.Atlas.CellW != 0 || h%o.Atlas.CellH != 0 {
		c.errf(p, "atlas %dx%d is not a multiple of the cell %dx%d", w, h, o.Atlas.CellW, o.Atlas.CellH)
		return
	}
	rows, cols := h/o.Atlas.CellH, w/o.Atlas.CellW
	if billboard && o.Views > rows {
		c.errf(p, "views %d but the atlas has %d rows", o.Views, rows)
	}
	for i := range o.Animations {
		a := &o.Animations[i]
		where := p + ": animation " + a.ID
		if billboard {
			if a.FramesPerView <= 0 {
				c.errf(where, "framesPerView must be positive")
				continue
			}
			if a.Col+a.FramesPerView > cols {
				c.errf(where, "columns [%d,%d) exceed the atlas (%d cols)", a.Col, a.Col+a.FramesPerView, cols)
			}
		} else {
			if a.Frames <= 0 {
				c.errf(where, "frames must be positive")
				continue
			}
			if a.Row < 0 || a.Row >= rows {
				c.errf(where, "row %d out of range (%d rows)", a.Row, rows)
			}
			if a.Frames > cols {
				c.errf(where, "%d frames exceed the atlas (%d cols)", a.Frames, cols)
			}
			if len(a.Durations) != 0 && len(a.Steps) != 0 {
				c.errf(where, "durations and steps are mutually exclusive")
			}
			if len(a.Durations) != 0 && len(a.Durations) != a.Frames {
				c.errf(where, "%d durations for %d frames", len(a.Durations), a.Frames)
			}
			for _, st := range a.Steps {
				if len(st) != 2 || st[0] < 0 || st[0] >= a.Frames {
					c.errf(where, "steps must be [frameIndex < frames, hold]")
					break
				}
			}
			for _, ev := range a.Events {
				if ev.Frame < 0 || ev.Frame >= a.Frames {
					c.errf(where, "event frame %d out of range", ev.Frame)
					break
				}
			}
			if a.Mirror != "" {
				found := false
				for j := range o.Animations {
					if o.Animations[j].ID == a.Mirror {
						found = true
						break
					}
				}
				if !found {
					c.errf(where, "mirror %q is not an animation of this object", a.Mirror)
				}
			}
		}
	}
	_ = assets
}

func (c *checker) checkGLBBindings(p, glbPath string, o *Object) {
	f, err := c.fsys.Open(glbPath)
	if err != nil {
		return // existence already reported by ref()
	}
	defer f.Close()
	info, err := ReadGLBInfo(f)
	if err != nil {
		c.errf(glbPath, "%v", err)
		return
	}
	clips := map[string]bool{}
	for _, a := range info.Animations {
		clips[a] = true
	}
	mats := map[string]bool{}
	for _, m := range info.Materials {
		mats[m] = true
	}
	for i := range o.Animations {
		a := &o.Animations[i]
		if a.Clip != "" && !clips[a.Clip] {
			c.errf(p+": animation "+a.ID, "clip %q is not in the GLB (has: %s)", a.Clip, strings.Join(info.Animations, ", "))
		}
	}
	for _, ua := range o.UVAnims {
		if !mats[ua.Material] {
			c.errf(p, "uvAnim material %q is not in the GLB", ua.Material)
		}
	}
	for _, fb := range o.Flipbooks {
		if !mats[fb.Material] {
			c.errf(p, "flipbook material %q is not in the GLB", fb.Material)
		}
	}
	if len(o.Variants) > 0 {
		if len(o.Variants) == 1 {
			c.errf(p, "variants with a single entry is pointless — omit the list")
		}
		scenes := map[string]int{}
		for i, s := range info.Scenes {
			scenes[s] = i
		}
		ids := map[string]bool{}
		for i, v := range o.Variants {
			where := p + ": variant " + v.ID
			if v.ID == "" || v.Name == "" || v.Scene == "" {
				c.errf(where, "id, name and scene are all required")
				continue
			}
			if ids[v.ID] {
				c.errf(where, "duplicate variant id")
			}
			ids[v.ID] = true
			si, ok := scenes[v.Scene]
			if !ok {
				c.errf(where, "scene %q is not in the GLB (has: %s)", v.Scene, strings.Join(info.Scenes, ", "))
			} else if i == 0 && si != 0 {
				c.errf(where, "the first variant must be the GLB's default scene (scene 0 is %q)", info.Scenes[0])
			}
		}
	}
}
