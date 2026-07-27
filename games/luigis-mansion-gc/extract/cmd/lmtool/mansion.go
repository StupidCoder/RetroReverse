package main

// mansion.go exports the whole mansion: every room's room.bin as a GLB,
// every furniture .bin once (content-deduped across the 75 room arcs, so
// the eleven o_isu chairs ship one GLB), and placements.json — the
// furniture placement database from Map/map2.szp jmp/furnitureinfo, each
// record resolved to its GLB. Rooms are modelled in mansion-global
// coordinates and furniture positions live in the same frame, so a viewer
// recreates a room by loading its GLB and instancing furniture at
// {pos, rotDeg, scale}.

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

type placement struct {
	Model string     `json:"model"`         // furniture GLB, e.g. "furniture/chest.glb"
	Tag   string     `json:"tag,omitempty"` // the record's second string, when set
	Pos   [3]float32 `json:"pos"`
	Rot   [3]float32 `json:"rotDeg"`
	Scale [3]float32 `json:"scale"`
}

type roomEntry struct {
	Room      string      `json:"room"` // e.g. "room_02"
	Model     string      `json:"model"`
	Furniture []placement `json:"furniture"`
}

type doorEntry struct {
	Model  string     `json:"model,omitempty"` // doors/door_09.glb; empty = open archway
	Pos    [3]float32 `json:"pos"`
	Axis   int        `json:"axis"` // 1: leaf faces z, 2: faces x, 4: floor opening
	Size   [3]float32 `json:"size"`
	Double bool       `json:"double,omitempty"`
	Rooms  [2]int     `json:"rooms"`
}

// The mansion's door table lives in the DOL, not in any file: the map-info
// header at 0x8030377C (GLME01/USA) carries the room table at +0x14 and the
// door list at +0x18 — 28-byte records {u8 axis, u8 flag, …, u8 typ@6,
// u8 id@7, s32 pos[3]@8, …, u16 size[3]@20, u8 roomA@26, u8 roomB@27},
// terminated by axis 0. typ 0xFF is a doorless opening; otherwise it indexes
// the 20-byte type table at 0x802FF95C: {u8 kind (2 = double), …,
// u8 model@19} with model naming /iwamoto/Door/{saku,door_NN}.bin (the path
// table at 0x802FF868).
const (
	dolMapHeaderVA = 0x8030377C
	dolDoorTypesVA = 0x802FF95C
)

func doorTable(d *gc.Disc) ([]doorEntry, map[int]bool, error) {
	dol, err := d.DOL()
	if err != nil {
		return nil, nil, err
	}
	// Flatten the DOL segments into a VA-addressable reader.
	segs := map[uint32][]byte{}
	dol.Load(func(addr uint32, b []byte) { segs[addr] = b })
	read := func(va uint32, n int) []byte {
		for base, b := range segs {
			if va >= base && va+uint32(n) <= base+uint32(len(b)) {
				return b[va-base : va-base+uint32(n)]
			}
		}
		return nil
	}
	u32 := func(va uint32) uint32 {
		b := read(va, 4)
		if b == nil {
			return 0
		}
		return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}
	listVA := u32(dolMapHeaderVA + 0x18)
	if listVA < 0x80000000 {
		return nil, nil, fmt.Errorf("no door list behind the map header")
	}
	var doors []doorEntry
	models := map[int]bool{} // model indices actually used
	for i := 0; ; i++ {
		rec := read(listVA+uint32(i)*28, 28)
		if rec == nil || rec[0] == 0 {
			break
		}
		s32 := func(o int) float32 {
			return float32(int32(uint32(rec[o])<<24 | uint32(rec[o+1])<<16 | uint32(rec[o+2])<<8 | uint32(rec[o+3])))
		}
		u16 := func(o int) float32 { return float32(uint32(rec[o])<<8 | uint32(rec[o+1])) }
		e := doorEntry{
			Axis:  int(rec[0]),
			Pos:   [3]float32{s32(8), s32(12), s32(16)},
			Size:  [3]float32{u16(20), u16(22), u16(24)},
			Rooms: [2]int{int(rec[26]), int(rec[27])},
		}
		// The whole word at +4 reading 255 marks a doorless opening (this is
		// the check the panel counter at 0x8001A0C0 makes); otherwise byte 6
		// picks the type.
		leafless := uint32(rec[4])<<24|uint32(rec[5])<<16|uint32(rec[6])<<8|uint32(rec[7]) == 255
		if typ := rec[6]; !leafless {
			t := read(dolDoorTypesVA+uint32(typ)*20, 20)
			if t != nil {
				model := int(t[19])
				name := "saku"
				if model > 0 {
					name = fmt.Sprintf("door_%02d", model)
				}
				e.Model = "doors/" + name + ".glb"
				e.Double = t[0] == 2
				models[model] = true
			}
		}
		doors = append(doors, e)
	}
	return doors, models, nil
}

func mansionExport(image, outDir string) error {
	d, err := gc.Open(image)
	if err != nil {
		return err
	}
	defer d.Close()
	read := func(path string) ([]byte, error) {
		for _, f := range d.FST.Entries {
			if !f.Dir && strings.EqualFold(f.Path, path) {
				b, err := d.Read(f.Offset, int(f.Size))
				if err != nil {
					return nil, err
				}
				if len(b) >= 4 && string(b[:4]) == "Yay0" {
					return lm.Yay0(b)
				}
				return b, nil
			}
		}
		return nil, fmt.Errorf("no file %q on the disc", path)
	}

	for _, dir := range []string{"rooms", "furniture", "doors"} {
		if err := os.MkdirAll(filepath.Join(outDir, dir), 0o755); err != nil {
			return err
		}
	}

	// Doors: the DOL-side list, plus the leaf models from game_usa.szp with
	// their swing clips (pull opens toward the walker, push away).
	doors, doorModels, err := doorTable(d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  doors: %v\n", err)
	}
	doorCount := 0
	if len(doors) > 0 {
		gb, err := read("/Game/game_usa.szp")
		if err != nil {
			return err
		}
		members, err := lm.RARC(gb)
		if err != nil {
			return err
		}
		var swing []anmClip
		for _, mem := range members {
			if mem.Dir == "iwamoto/door" && (mem.Name == "pull.anm" || mem.Name == "push.anm") {
				if a, err := lm.ParseAnm(mem.Data); err == nil {
					swing = append(swing, anmClip{Name: strings.TrimSuffix(mem.Name, ".anm"), Anm: a})
				}
			}
		}
		sortClips(swing)
		for idx := range doorModels {
			name := "saku"
			if idx > 0 {
				name = fmt.Sprintf("door_%02d", idx)
			}
			for _, mem := range members {
				if mem.Dir == "iwamoto/door" && mem.Name == name+".bin" {
					m, err := lm.ParseBin(mem.Data)
					if err != nil {
						fmt.Fprintf(os.Stderr, "  skip door %s: %v\n", name, err)
						continue
					}
					if err := binGLBAnimated(m, swing, filepath.Join(outDir, "doors", name+".glb"), name); err != nil {
						fmt.Fprintf(os.Stderr, "  skip door %s: %v\n", name, err)
						continue
					}
					doorCount++
				}
			}
		}
	}

	// Sweep the room arcs: rooms export directly, furniture dedupes by
	// content hash. byName tracks hash→file per member name so a name
	// reused for different geometry gets a numbered variant.
	var roomArcs []string
	for _, f := range d.FST.Entries {
		if !f.Dir && strings.HasPrefix(f.Path, "/Iwamoto/map2/room_") && strings.HasSuffix(f.Path, ".arc") {
			roomArcs = append(roomArcs, f.Path)
		}
	}
	sort.Strings(roomArcs)

	type variant struct{ hash, file string }
	byName := map[string][]variant{}          // member base name → content variants
	resolve := map[[2]string]string{}         // {room, member base} → furniture file
	furnCount, roomCount, failCount := 0, 0, 0
	for _, arc := range roomArcs {
		roomName := strings.TrimSuffix(filepath.Base(arc), ".arc")
		b, err := read(arc)
		if err != nil {
			return err
		}
		members, err := lm.RARC(b)
		if err != nil {
			return err
		}
		// The arc's anm/ directory: clips keyed by the furniture base name
		// (anm/chest_0.anm and anm/chest_1.anm belong to chest.bin).
		clipsFor := map[string][]anmClip{}
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
				fmt.Fprintf(os.Stderr, "  skip %s:anm/%s: %v\n", roomName, mem.Name, err)
				failCount++
				continue
			}
			clipsFor[base] = append(clipsFor[base], anmClip{Name: clip, Anm: a})
		}
		for _, mem := range members {
			if !strings.HasSuffix(mem.Name, ".bin") || len(mem.Data) < 0x60 {
				continue
			}
			base := strings.TrimSuffix(mem.Name, ".bin")
			m, err := lm.ParseBin(mem.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skip %s:%s: %v\n", roomName, mem.Name, err)
				failCount++
				continue
			}
			if mem.Name == "room.bin" {
				if err := binGLB(m, filepath.Join(outDir, "rooms", roomName+".glb"), roomName); err != nil {
					return fmt.Errorf("%s: %w", roomName, err)
				}
				roomCount++
				continue
			}
			h := fmt.Sprintf("%x", sha256.Sum256(mem.Data))
			var file string
			for _, v := range byName[base] {
				if v.hash == h {
					file = v.file
					break
				}
			}
			if file == "" {
				file = base + ".glb"
				if n := len(byName[base]); n > 0 {
					file = fmt.Sprintf("%s-%d.glb", base, n+1)
				}
				var err error
				if clips := clipsFor[base]; len(clips) > 0 {
					sortClips(clips)
					err = binGLBAnimated(m, clips, filepath.Join(outDir, "furniture", file), base)
				} else {
					err = binGLB(m, filepath.Join(outDir, "furniture", file), base)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "  skip %s:%s: %v\n", roomName, mem.Name, err)
					failCount++
					continue
				}
				byName[base] = append(byName[base], variant{h, file})
				furnCount++
			}
			resolve[[2]string{roomName, base}] = file
		}
	}

	// The placement database.
	mapb, err := read("/Map/map2.szp")
	if err != nil {
		return err
	}
	members, err := lm.RARC(mapb)
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

	rooms := map[string]*roomEntry{}
	unresolved := 0
	for r := range table.Records {
		model := table.Str(r, lm.JMPDMDName)
		if model == "" || model == "(null)" {
			continue
		}
		roomName := fmt.Sprintf("room_%02d", table.U32(r, lm.JMPRoomNo))
		file, ok := resolve[[2]string{roomName, model}]
		if !ok {
			// The member lives in another arc (shared geometry): any
			// same-named export will do if the name is unambiguous.
			if vs := byName[model]; len(vs) == 1 {
				file = vs[0].file
			} else {
				fmt.Fprintf(os.Stderr, "  unresolved: %s in %s (%d variants)\n", model, roomName, len(byName[model]))
				unresolved++
				continue
			}
		}
		e := rooms[roomName]
		if e == nil {
			e = &roomEntry{Room: roomName, Model: "rooms/" + roomName + ".glb"}
			rooms[roomName] = e
		}
		e.Furniture = append(e.Furniture, placement{
			Model: "furniture/" + file,
			Tag:   nullless(table.Str(r, 0x00B64611)),
			Pos:   [3]float32{table.F32(r, lm.JMPPosX), table.F32(r, lm.JMPPosY), table.F32(r, lm.JMPPosZ)},
			Rot:   [3]float32{table.F32(r, lm.JMPDirX), table.F32(r, lm.JMPDirY), table.F32(r, lm.JMPDirZ)},
			Scale: [3]float32{table.F32(r, lm.JMPSclX), table.F32(r, lm.JMPSclY), table.F32(r, lm.JMPSclZ)},
		})
	}
	var list []roomEntry
	for _, arc := range roomArcs {
		name := strings.TrimSuffix(filepath.Base(arc), ".arc")
		if e := rooms[name]; e != nil {
			list = append(list, *e)
		} else {
			list = append(list, roomEntry{Room: name, Model: "rooms/" + name + ".glb"})
		}
	}
	j, err := json.MarshalIndent(map[string]any{"rooms": list, "doors": doors}, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "placements.json"), j, 0o644); err != nil {
		return err
	}
	fmt.Printf("mansion: %d rooms, %d unique furniture GLBs, %d placements, %d doors (%d leaf models), %d unresolved, %d skipped\n",
		roomCount, furnCount, countPlacements(list), len(doors), doorCount, unresolved, failCount)
	return nil
}

func countPlacements(list []roomEntry) int {
	n := 0
	for _, e := range list {
		n += len(e.Furniture)
	}
	return n
}

func nullless(s string) string {
	if s == "(null)" {
		return ""
	}
	return s
}
