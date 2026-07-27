package main

// mansion.go — the whole-mansion level. Room shells stream in storey areas
// (the dollhouse peel), furniture instances come from Map/map2.szp
// jmp/furnitureinfo with the vacuum interaction on their last .anm clip, and
// the DOL door list becomes door-leaf placements with the swing clip held at
// its apex on click.

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
)

// Storey areas: a shell lands in the area of its lowest storey (the same
// ladder the game's own room list implies), and the porch/roof shells are
// the exterior.
var areas = []schema.Area{
	{ID: "basement", Name: "Basement"},
	{ID: "ground", Name: "Ground floor"},
	{ID: "first", Name: "First floor"},
	{ID: "attic", Name: "Attic"},
	{ID: "exterior", Name: "Exterior"},
}

func areaOf(minY float32) string {
	switch {
	case minY < -300:
		return "basement"
	case minY < 300:
		return "ground"
	case minY < 900:
		return "first"
	default:
		return "attic"
	}
}

var exteriorRooms = map[string]bool{
	"room_11": true, "room_16": true, "room_23": true, "room_37": true,
	"room_59": true, "room_60": true, "room_72": true,
}

// Furniture display names for pieces the write-up identified; everything else
// keeps its archive base name.
var furnitureNames = map[string]string{
	"otukue": "Tea table", "oisu": "Chair", "otoire": "Toilet",
	"chest": "Chest of drawers", "mirror": "Foyer mirror", "syan": "Chandelier",
	"ota1": "Tool cabinet",
}

func slugify(s string) string {
	var out []rune
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	return strings.Trim(string(out), "-")
}

func exportMansion(ctx *cli.Context, src *export.Source, doObjects, doLevels bool) error {
	b := ctx.Builder

	// --- doors: leaf models from game_usa.szp with their swing clips -------
	doors, doorModels, err := export.DoorTable(src.Disc)
	if err != nil {
		return fmt.Errorf("door table: %w", err)
	}
	doorAsset := map[string]string{}       // "doors/door_09.glb" (table ref) → asset id
	doorInteraction := map[string]string{} // asset id → swing clip id
	if doObjects {
		members, err := src.Archive("/Game/game_usa.szp")
		if err != nil {
			return err
		}
		var swing []export.AnmClip
		for _, mem := range members {
			if mem.Dir == "iwamoto/door" && (mem.Name == "pull.anm" || mem.Name == "push.anm") {
				if a, err := lm.ParseAnm(mem.Data); err == nil {
					swing = append(swing, export.AnmClip{Name: strings.TrimSuffix(mem.Name, ".anm"), Anm: a})
				}
			}
		}
		export.SortClips(swing)
		count := 0
		for idx := range doorModels {
			name := "saku"
			if idx > 0 {
				name = fmt.Sprintf("door_%02d", idx)
			}
			mem := export.Member(members, name+".bin")
			if mem == nil || mem.Dir != "iwamoto/door" {
				// Member() matches by name only; double-check the directory.
				mem = nil
				for i := range members {
					if members[i].Dir == "iwamoto/door" && members[i].Name == name+".bin" {
						mem = &members[i]
						break
					}
				}
				if mem == nil {
					continue
				}
			}
			m, err := lm.ParseBin(mem.Data)
			if err != nil {
				ctx.Logf("skip door %s: %v", name, err)
				continue
			}
			id := slugify(name)
			glbPath, err := b.Path("objects", id+".glb")
			if err != nil {
				return err
			}
			if err := export.BinGLBAnimated(m, swing, glbPath, name); err != nil {
				ctx.Logf("skip door %s: %v", name, err)
				continue
			}
			metas, interaction := clipMeta(swing)
			dispName := "Door " + strings.TrimPrefix(name, "door_")
			if name == "saku" {
				dispName = "Gate leaf"
			}
			b.AddObject(schema.Asset{ID: id, Name: dispName, Group: "Doors"}, &schema.Object{
				Type: schema.ObjectModel3D, Name: dispName, Model: id + ".glb",
				Instanced:  true,
				Animations: metas,
				Props: map[string]any{
					"source": "/Game/game_usa.szp iwamoto/door/" + name + ".bin",
					"clips":  "pull opens toward the walker, push away; a full open-and-shut cycle each",
				},
			})
			doorAsset["doors/"+name+".glb"] = id
			doorInteraction[id] = interaction
			count++
		}
		ctx.Progress("objects", count, len(doorModels), "door leaf models")
	}

	// --- sweep the 75 room arcs: shells + content-deduped furniture --------
	roomArcs := src.List("/Iwamoto/map2/room_", ".arc")
	sort.Strings(roomArcs)

	type variant struct{ hash, assetID, interaction string }
	byName := map[string][]variant{}   // member base name → content variants
	resolve := map[[2]string]variant{} // {room, member base} → variant
	usedIDs := map[string]bool{}
	type roomInfo struct {
		id   int
		name string
		aabb schema.AABB
		area string
	}
	var roomList []roomInfo
	furnCount, roomCount := 0, 0

	shellID := map[string]int{} // arc name → unique shell id ("room_28A" is its own shell)
	for ai, arc := range roomArcs {
		roomName := strings.TrimSuffix(arc[strings.LastIndex(arc, "/")+1:], ".arc")
		roomNo := len(shellID)
		shellID[roomName] = roomNo
		members, err := src.Archive(arc)
		if err != nil {
			return err
		}
		// The arc's anm/ directory: clips keyed by the furniture base name
		// (anm/chest_0.anm and anm/chest_1.anm belong to chest.bin).
		clipsFor := map[string][]export.AnmClip{}
		for _, mem := range members {
			if mem.Dir != "anm" || !strings.HasSuffix(mem.Name, ".anm") {
				continue
			}
			clip := strings.TrimSuffix(mem.Name, ".anm")
			base := clip
			if i := strings.LastIndex(clip, "_"); i > 0 {
				base = clip[:i]
			}
			a, err := lm.ParseAnm(mem.Data)
			if err != nil {
				ctx.Logf("skip %s:anm/%s: %v", roomName, mem.Name, err)
				continue
			}
			clipsFor[base] = append(clipsFor[base], export.AnmClip{Name: clip, Anm: a})
		}
		for _, mem := range members {
			if !strings.HasSuffix(mem.Name, ".bin") || len(mem.Data) < 0x60 {
				continue
			}
			base := strings.TrimSuffix(mem.Name, ".bin")
			m, err := lm.ParseBin(mem.Data)
			if err != nil {
				ctx.Logf("skip %s:%s: %v", roomName, mem.Name, err)
				continue
			}
			if mem.Name == "room.bin" {
				mn, mx, ok := export.BinBounds(m)
				info := roomInfo{id: roomNo, name: roomName}
				if ok {
					info.aabb = schema.AABB{Min: float64toF(mn), Max: float64toF(mx)}
					info.area = areaOf(mn[1])
				}
				if exteriorRooms[roomName] {
					info.area = "exterior"
				}
				if doLevels {
					glbPath, err := b.Path("levels", "mansion", "rooms", roomName+".glb")
					if err != nil {
						return err
					}
					if err := export.BinGLB(m, glbPath, roomName); err != nil {
						return fmt.Errorf("%s: %w", roomName, err)
					}
				}
				roomList = append(roomList, info)
				roomCount++
				continue
			}
			if !doObjects {
				continue
			}
			h := fmt.Sprintf("%x", sha256.Sum256(mem.Data))
			var got *variant
			for i := range byName[base] {
				if byName[base][i].hash == h {
					got = &byName[base][i]
					break
				}
			}
			if got == nil {
				id := slugify(base)
				if usedIDs[id] || id == "" {
					id = fmt.Sprintf("%s-%d", id, len(byName[base])+1)
				}
				usedIDs[id] = true
				glbPath, err := b.Path("objects", id+".glb")
				if err != nil {
					return err
				}
				clips := clipsFor[base]
				export.SortClips(clips)
				if len(clips) > 0 {
					err = export.BinGLBAnimated(m, clips, glbPath, base)
				} else {
					err = export.BinGLB(m, glbPath, base)
				}
				if err != nil {
					ctx.Logf("skip %s:%s: %v", roomName, mem.Name, err)
					continue
				}
				metas, interaction := clipMeta(clips)
				name := furnitureNames[base]
				if name == "" {
					name = base
				}
				b.AddObject(schema.Asset{ID: id, Name: name, Group: "Furniture"}, &schema.Object{
					Type: schema.ObjectModel3D, Name: name, Model: id + ".glb",
					Instanced:  true,
					Animations: metas,
				})
				byName[base] = append(byName[base], variant{h, id, interaction})
				got = &byName[base][len(byName[base])-1]
				furnCount++
			}
			resolve[[2]string{roomName, base}] = *got
		}
		if (ai+1)%15 == 0 {
			ctx.Progress("levels", ai+1, len(roomArcs), roomName)
		}
	}
	ctx.Logf("mansion: %d rooms, %d unique furniture objects", roomCount, furnCount)

	if !doLevels {
		return nil
	}

	// --- the level document -------------------------------------------------
	// The tour starts inside the lobby: room_02 is the foyer (the room whose
	// jmp/furnitureinfo record carries the foyer mirror), entered through the
	// mansion's front (negative-z) face. Eye height just inside the doors,
	// looking across the room. If the shell is ever missing, fall back to an
	// establishing shot derived from the assembled shells' bounds.
	var lo, hi [3]float64
	first := true
	for _, r := range roomList {
		if r.aabb.Min == r.aabb.Max {
			continue
		}
		if first {
			lo, hi, first = r.aabb.Min, r.aabb.Max, false
			continue
		}
		for k := 0; k < 3; k++ {
			lo[k] = math.Min(lo[k], r.aabb.Min[k])
			hi[k] = math.Max(hi[k], r.aabb.Max[k])
		}
	}
	cx, cy, cz := (lo[0]+hi[0])/2, (lo[1]+hi[1])/2, (lo[2]+hi[2])/2
	span := math.Max(hi[0]-lo[0], hi[2]-lo[2])
	pos := []float64{cx, hi[1] + span*0.3, lo[2] - span*0.55}
	target := []float64{cx, cy, cz}
	const lobbyRoom = "room_02"
	for _, r := range roomList {
		if r.name != lobbyRoom || r.aabb.Min == r.aabb.Max {
			continue
		}
		lx := (r.aabb.Min[0] + r.aabb.Max[0]) / 2
		pos = []float64{lx, r.aabb.Min[1] + 220, r.aabb.Min[2] + 180}
		target = []float64{lx, r.aabb.Min[1] + 400, r.aabb.Max[2]}
		break
	}
	doc := &schema.Level{
		Type: schema.LevelScene3D,
		Camera: &schema.Camera{
			Mode:   "fly",
			Pos:    pos,
			Target: target,
			FOV:    50, Near: 5, Far: 200000,
			Fly: &schema.Fly{Speed: 1250},
		},
		Scene: &schema.Scene{
			Rooms: &schema.Rooms{Areas: areas, Stream: true},
		},
	}
	for _, r := range roomList {
		aabb := r.aabb
		doc.Scene.Rooms.List = append(doc.Scene.Rooms.List, schema.Room{
			ID: r.id, Name: r.name, File: "mansion/rooms/" + r.name + ".glb",
			Area: r.area, AABB: &aabb,
		})
	}
	if doObjects {
		// Furniture placements from jmp/furnitureinfo.
		members, err := src.Archive("/Map/map2.szp")
		if err != nil {
			return err
		}
		var table *lm.JMPTable
		for _, mem := range members {
			if mem.Dir == "jmp" && mem.Name == "furnitureinfo" {
				if table, err = lm.ParseJMP(mem.Data); err != nil {
					return err
				}
			}
		}
		if table == nil {
			return fmt.Errorf("no jmp/furnitureinfo in /Map/map2.szp")
		}
		pid := 0
		unresolved := 0
		for r := range table.Records {
			model := table.Str(r, lm.JMPDMDName)
			if model == "" || model == "(null)" {
				continue
			}
			roomName := fmt.Sprintf("room_%02d", int(table.U32(r, lm.JMPRoomNo)))
			v, ok := resolve[[2]string{roomName, model}]
			if !ok {
				// The member lives in another arc (shared geometry): any
				// same-named export will do if the name is unambiguous.
				if vs := byName[model]; len(vs) == 1 {
					v = vs[0]
				} else {
					unresolved++
					continue
				}
			}
			pl := schema.Placement{
				ID:     pid,
				Object: v.assetID,
				Pos: []float64{
					float64(table.F32(r, lm.JMPPosX)),
					float64(table.F32(r, lm.JMPPosY)),
					float64(table.F32(r, lm.JMPPosZ)),
				},
			}
			if sid, ok := shellID[roomName]; ok {
				pl.Room = intp(sid)
			}
			rot := [3]float64{
				float64(table.F32(r, lm.JMPDirX)),
				float64(table.F32(r, lm.JMPDirY)),
				float64(table.F32(r, lm.JMPDirZ)),
			}
			scale := [3]float64{
				float64(table.F32(r, lm.JMPSclX)),
				float64(table.F32(r, lm.JMPSclY)),
				float64(table.F32(r, lm.JMPSclZ)),
			}
			setTRS(&pl, pl.Pos, rot, scale)
			if tag := table.Str(r, 0x00B64611); tag != "" && tag != "(null)" {
				if pl.Props == nil {
					pl.Props = map[string]any{}
				}
				pl.Props["tag"] = tag
			}
			if v.interaction != "" {
				pl.OnClick = &schema.OnClick{
					Action: schema.ActionAnimate, Clip: v.interaction, Toggle: true,
				}
			}
			doc.Placements = append(doc.Placements, pl)
			pid++
		}
		if unresolved > 0 {
			ctx.Logf("mansion: %d placements unresolved (ambiguous shared members)", unresolved)
		}

		// Door-leaf placements: the record's position is the opening's
		// centre; the leaf model is 200x300 with the hinge at its local x=0
		// edge. A single leaf sits at the record position (hinge on the
		// right jamb); a double's two leaves shift half an opening each way,
		// mirrored. Axis 2 turns the plane onto x; axis 4 lays it flat.
		for di, dr := range doors {
			if dr.Model == "" {
				continue // a doorless archway — just an opening
			}
			asset := doorAsset[dr.Model]
			if asset == "" {
				continue
			}
			w := float64(dr.Size[0])
			if dr.Axis == 2 {
				w = float64(dr.Size[2])
			}
			if w == 0 {
				w = 200
			}
			h := float64(dr.Size[1])
			if h == 0 {
				h = 300
			}
			off := math.Max(0, (w-200)/2)
			leaves := [][2]float64{{0, 1}}
			if dr.Double {
				leaves = [][2]float64{{off, 1}, {-off, -1}}
			}
			for li, leaf := range leaves {
				pl := schema.Placement{
					ID:     100000 + di*10 + li,
					Object: asset,
					Matrix: doorLeafMatrix(dr, leaf[0], leaf[1], h),
					Name:   "Door",
					OnClick: &schema.OnClick{
						Action: schema.ActionAnimate,
						Clip:   doorInteraction[asset],
						HoldAt: 0.5, // the swing clip is a full open-and-shut; hold the apex
						Toggle: true,
					},
					Props: map[string]any{
						"rooms": []int{dr.Rooms[0], dr.Rooms[1]},
						"axis":  dr.Axis,
					},
				}
				doc.Placements = append(doc.Placements, pl)
			}
		}
	}

	b.AddLevel(schema.Asset{ID: "mansion", Name: "The mansion"}, doc)
	ctx.Progress("levels", len(roomArcs), len(roomArcs),
		fmt.Sprintf("mansion: %d rooms, %d placements", roomCount, len(doc.Placements)))
	return nil
}

// doorLeafMatrix composes T(pos) · R(axis) · T(offX, -h/2, 0) · S(sx,1,1) as
// a column-major glTF matrix.
func doorLeafMatrix(dr export.DoorEntry, offX, sx, h float64) []float64 {
	// Rotation about Y (axis 2) or X (axis 4) by +90°.
	var r [3][3]float64
	switch dr.Axis {
	case 2: // leaf plane onto x
		r = [3][3]float64{{0, 0, 1}, {0, 1, 0}, {-1, 0, 0}}
	case 4: // trapdoor: flat
		r = [3][3]float64{{1, 0, 0}, {0, 0, -1}, {0, 1, 0}}
	default:
		r = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	}
	// local leaf offset in the holder's frame, then rotated into world.
	lx, ly, lz := offX, -h/2, 0.0
	wx := r[0][0]*lx + r[0][1]*ly + r[0][2]*lz + float64(dr.Pos[0])
	wy := r[1][0]*lx + r[1][1]*ly + r[1][2]*lz + float64(dr.Pos[1])
	wz := r[2][0]*lx + r[2][1]*ly + r[2][2]*lz + float64(dr.Pos[2])
	// Columns: basis * mirror scale on local x.
	return []float64{
		r[0][0] * sx, r[1][0] * sx, r[2][0] * sx, 0,
		r[0][1], r[1][1], r[2][1], 0,
		r[0][2], r[1][2], r[2][2], 0,
		wx, wy, wz, 1,
	}
}

// setTRS emits pos/rot/scale when the rotation is effectively single-axis
// (the ZYX-vs-XYZ Euler order distinction vanishes), and a full matrix when
// it is not — the JMP records store ZYX degrees, Retro-X speaks XYZ radians.
func setTRS(pl *schema.Placement, pos []float64, rotDeg, scale [3]float64) {
	nonzero := 0
	for _, d := range rotDeg {
		if d != 0 {
			nonzero++
		}
	}
	uniform := scale[0] == scale[1] && scale[1] == scale[2]
	if nonzero <= 1 {
		var rot []float64
		if nonzero == 1 {
			rot = []float64{rad(rotDeg[0]), rad(rotDeg[1]), rad(rotDeg[2])}
		}
		pl.Rot = rot
		if uniform {
			if scale[0] != 1 {
				pl.Scale = schema.Scale{scale[0]}
			}
		} else {
			pl.Scale = schema.Scale{scale[0], scale[1], scale[2]}
		}
		return
	}
	// Full ZYX-composed matrix.
	rx, ry, rz := rad(rotDeg[0]), rad(rotDeg[1]), rad(rotDeg[2])
	sx, cx := math.Sincos(rx)
	sy, cy := math.Sincos(ry)
	sz, cz := math.Sincos(rz)
	// R = Rz * Ry * Rx (the .bin node convention, binNodeMatrix).
	r := [3][3]float64{
		{cy * cz, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx},
		{sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx},
		{-sy, cy * sx, cy * cx},
	}
	pl.Pos = nil
	pl.Matrix = []float64{
		r[0][0] * scale[0], r[1][0] * scale[0], r[2][0] * scale[0], 0,
		r[0][1] * scale[1], r[1][1] * scale[1], r[2][1] * scale[1], 0,
		r[0][2] * scale[2], r[1][2] * scale[2], r[2][2] * scale[2], 0,
		pos[0], pos[1], pos[2], 1,
	}
}

func rad(d float64) float64 { return d * math.Pi / 180 }

func intp(v int) *int { return &v }

// float64toF is a tiny helper because schema.AABB is [3]float64.
func float64toF(v [3]float32) [3]float64 {
	return [3]float64{float64(v[0]), float64(v[1]), float64(v[2])}
}
