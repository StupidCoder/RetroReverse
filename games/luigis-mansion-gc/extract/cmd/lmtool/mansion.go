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

	for _, dir := range []string{"rooms", "furniture"} {
		if err := os.MkdirAll(filepath.Join(outDir, dir), 0o755); err != nil {
			return err
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
				if err := binGLB(m, filepath.Join(outDir, "furniture", file), base); err != nil {
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
	j, err := json.MarshalIndent(map[string]any{"rooms": list}, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "placements.json"), j, 0o644); err != nil {
		return err
	}
	fmt.Printf("mansion: %d rooms, %d unique furniture GLBs, %d placements, %d unresolved, %d skipped\n",
		roomCount, furnCount, countPlacements(list), unresolved, failCount)
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
