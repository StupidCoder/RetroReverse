package export

// doors.go — the mansion's door table. It lives in the DOL, not in any file:
// the map-info header at 0x8030377C (GLME01/USA) carries the room table at
// +0x14 and the door list at +0x18 — 28-byte records {u8 axis, u8 flag, …,
// u8 typ@6, u8 id@7, s32 pos[3]@8, …, u16 size[3]@20, u8 roomA@26, u8
// roomB@27}, terminated by axis 0. typ 0xFF is a doorless opening; otherwise
// it indexes the 20-byte type table at 0x802FF95C: {u8 kind (2 = double), …,
// u8 model@19} with model naming /iwamoto/Door/{saku,door_NN}.bin (the path
// table at 0x802FF868).

import (
	"fmt"

	"retroreverse.com/tools/platform/gc"
)

type DoorEntry struct {
	Model  string     `json:"model,omitempty"` // doors/door_09.glb; empty = open archway
	Pos    [3]float32 `json:"pos"`
	Axis   int        `json:"axis"` // 1: leaf faces z, 2: faces x, 4: floor opening
	Size   [3]float32 `json:"size"`
	Double bool       `json:"double,omitempty"`
	Rooms  [2]int     `json:"rooms"`
}

const (
	dolMapHeaderVA = 0x8030377C
	dolDoorTypesVA = 0x802FF95C
)

// DoorTable reads the DOL-side door list. The returned map holds the leaf
// model indices actually referenced (0 = saku).
func DoorTable(d *gc.Disc) ([]DoorEntry, map[int]bool, error) {
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
	var doors []DoorEntry
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
		e := DoorEntry{
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
