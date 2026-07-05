// Level-data decoding, established by tracing the game's own code (Part V §1):
//
//   - Levels are selected through a 52-entry level->overlay table at ARM9
//     $020758C8 (file $718C8): the loader at $0202DED4 does
//     `LDR r5,[$020758C8 + level*4]`, compares against -1 and hands the overlay
//     ID to the overlay loader at $02018028. The shipped table is the identity
//     mapping [8,9,...,59].
//   - Each level's SETTINGS BLOCK is found through a second 52-entry pointer
//     table at ARM9 $02092208 (file $8E208): the level-start code at $0202D274
//     does `LDRSB r3,[currentLevel]; LDR r2,[$02092208 + r3*4]` and passes the
//     block to the level-data processor at $020FE190 (overlay 2).
//   - $020FE190 consumes the block: +$0A u16 collision-map internal file ID,
//     +$04 misc objects table, +$10 area table with +$14 u8 count (12-byte
//     entries, +$00 = the area's objects table), +$08 u16 level-model file ID.
//   - "Internal file IDs" index a 2058-entry table of filename pointers at
//     overlay-0 +$13098 ($020BD4B8): overlay 0's initializer at $020AA420 loops
//     `LDR r0,[$020BD4B8 + i*4]` over exactly $80A entries, resolving each path
//     and registering index -> file.
//   - An objects table is {u16 count, u32 entries}; each 8-byte entry is
//     {u8 type|layer<<5, u8 count, u16 pad, u32 list}. The walker at $020FE33C
//     extracts type = b & $1F, layer = (b>>5) & 7, skips entries whose layer
//     differs from the current star, and dispatches through a 15-entry handler
//     table at $0210CBB8.
//   - Type-0 handler $020FE8AC (standard objects, 16-byte stride): u16 object
//     ID at +0 (translated through the object->actor table at $0210CBF4:
//     `LDRH [$0210CBF4 + id*2]`), s16 x/y/z at +2/+4/+6, each `LSL #12` into
//     fx20.12 — the s16 IS the world coordinate. A model's own space is its
//     raw fx4.12 vertices times the 2^shift in its header, so a placement in
//     stage-model units is world >> stageShift. Params at +8..+D
//     passed by pointer (u16 par2 at +8, s16 y-rotation at +$A in standard DS
//     angle-index units where $10000 = 360°, u16 par3 at +$C), u16 par1 at +$E.
//   - Type-5 handler $020FE960 (simple objects, 8-byte stride): u16 at +0 is
//     id & $1FF (the mask is a literal in the handler) with par = word >> 9,
//     s16 x/y/z at +2/+4/+6 as above. Same actor-table translation.
//
// Actor identity: the factory at $02043098 calls
// `LDR r0,[[profileTable] + actor*4]; BLX [r0]` — each actor profile starts
// with its create function, carries its actor ID twice at +4/+6, and points at
// its C++ typeinfo at +$20, whose +4 names the class (a length-prefixed string
// like "18daObjMc_Metalnet_c"). Those class names are the game's own object
// names; ScanActorProfiles sweeps the binaries for them.
package sm64ds

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"retroreverse.com/tools/nds"
)

const (
	levelOvlTable = 0x718C8 // ARM9 file offset: level -> overlay ID (u32 x52)
	settingsTable = 0x8E208 // ARM9 file offset: level -> settings-block RAM addr (u32 x52)
	fileTableOff  = 0x13098 // overlay-0 offset: internal file ID -> name pointer
	fileTableLen  = 2058    // loop bound in overlay 0's initializer ($080A)
	objActorTable = 0x5F594 // overlay-2 offset of the object->actor u16 table ($0210CBF4)
	arm9Base      = 0x02004000
	NumLevels     = 52
)

// LevelObject is one placed object, in world units (the fx20.12 integer part;
// divide by 2^stageShift for stage-model/GLB units).
type LevelObject struct {
	ID      int     // object ID (index into the object->actor table)
	Actor   int     // actor ID (the table's translation)
	Layer   int     // 0 = all stars, 1-7 = that star only
	X, Y, Z float64 // world units
	RotY    float64 // degrees
	Params  [3]int  // par1; par2, par3 (standard objects only)
	Simple  bool    // from an 8-byte "simple" entry
}

// Level is one decoded level: its stage files and every placed object.
type Level struct {
	ID       int
	Overlay  int
	BMDPath  string
	KCLPath  string
	NumAreas int
	Objects  []LevelObject
}

// LevelSet gives access to all levels of the cartridge image.
type LevelSet struct {
	rom      *nds.ROM
	extDir   string
	arm9     []byte
	intName  []string // internal file ID -> path
	objActor []uint16 // object ID -> actor ID
	ovls     map[int]nds.Overlay
}

// OpenLevels loads the ROM image and the extracted (decompressed) binaries.
func OpenLevels(romPath, extractedDir string) (*LevelSet, error) {
	img, err := os.ReadFile(romPath)
	if err != nil {
		return nil, err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return nil, err
	}
	ls := &LevelSet{rom: rom, extDir: extractedDir, ovls: map[int]nds.Overlay{}}
	for _, o := range rom.ARM9Overlays() {
		ls.ovls[int(o.ID)] = o
	}
	if ls.arm9, err = os.ReadFile(filepath.Join(extractedDir, "arm9_dec.bin")); err != nil {
		return nil, err
	}

	// internal-file table (overlay 0)
	ovl0, err := ls.overlayData(0)
	if err != nil {
		return nil, err
	}
	base0 := ls.ovls[0].RAMAddr
	ls.intName = make([]string, fileTableLen)
	for i := 0; i < fileTableLen; i++ {
		off := int(le.Uint32(ovl0[fileTableOff+i*4:]) - base0)
		if off < 0 || off >= len(ovl0) {
			continue
		}
		e := off
		for e < len(ovl0) && ovl0[e] != 0 {
			e++
		}
		ls.intName[i] = string(ovl0[off:e])
	}

	// object -> actor table (overlay 2)
	ovl2, err := ls.overlayData(2)
	if err != nil {
		return nil, err
	}
	n := 512 // generous bound; object IDs are 9-bit (the type-5 mask is $1FF)
	ls.objActor = make([]uint16, n)
	for i := 0; i < n && objActorTable+i*2+2 <= len(ovl2); i++ {
		ls.objActor[i] = le.Uint16(ovl2[objActorTable+i*2:])
	}
	return ls, nil
}

func (ls *LevelSet) overlayData(id int) ([]byte, error) {
	return os.ReadFile(filepath.Join(ls.extDir, fmt.Sprintf("ovl9_%03d_dec.bin", id)))
}

// InternalName resolves an internal file ID to its filesystem path.
func (ls *LevelSet) InternalName(id int) string {
	if id < 0 || id >= len(ls.intName) {
		return ""
	}
	return ls.intName[id]
}

// Actor translates an object ID through the game's object->actor table.
func (ls *LevelSet) Actor(objID int) int {
	if objID < 0 || objID >= len(ls.objActor) {
		return -1
	}
	return int(ls.objActor[objID])
}

// Level decodes level `id` (0..51) from its overlay via the two ARM9 tables.
func (ls *LevelSet) Level(id int) (*Level, error) {
	if id < 0 || id >= NumLevels {
		return nil, fmt.Errorf("sm64ds: level %d out of range", id)
	}
	oid := int(le.Uint32(ls.arm9[levelOvlTable+id*4:]))
	o, ok := ls.ovls[oid]
	if !ok {
		return nil, fmt.Errorf("sm64ds: level %d: overlay %d not in table", id, oid)
	}
	b, err := ls.overlayData(oid)
	if err != nil {
		return nil, err
	}
	base := o.RAMAddr
	off := func(p uint32) int { return int(p - base) }
	inRAM := func(p uint32) bool { return p >= base && p < base+uint32(len(b)) }

	hdrRAM := le.Uint32(ls.arm9[settingsTable+id*4:])
	if !inRAM(hdrRAM) {
		return nil, fmt.Errorf("sm64ds: level %d: settings %08x outside overlay %d", id, hdrRAM, oid)
	}
	hdr := off(hdrRAM)
	misc := le.Uint32(b[hdr+4:])
	bmdI := int(le.Uint16(b[hdr+8:]))
	kclI := int(le.Uint16(b[hdr+10:]))
	areaPtr := le.Uint32(b[hdr+0x10:])
	nArea := int(b[hdr+0x14])

	lv := &Level{ID: id, Overlay: oid, BMDPath: ls.InternalName(bmdI), KCLPath: ls.InternalName(kclI), NumAreas: nArea}

	// walk one {count, entries} objects table, keeping standard + simple objects
	parse := func(t int) {
		if t < 0 || t+8 > len(b) {
			return
		}
		n := int(le.Uint16(b[t:]))
		p := le.Uint32(b[t+4:])
		if n <= 0 || !inRAM(p) {
			return
		}
		for i := 0; i < n; i++ {
			e := off(p) + i*8
			typ, cnt, layer := int(b[e]&0x1F), int(b[e+1]), int(b[e]>>5)
			lp := le.Uint32(b[e+4:])
			if !inRAM(lp) {
				continue
			}
			lo := off(lp)
			switch typ {
			case 0: // standard, 16-byte stride
				for j := 0; j < cnt; j++ {
					q := lo + j*16
					if q+16 > len(b) {
						break
					}
					oid := int(le.Uint16(b[q:]))
					lv.Objects = append(lv.Objects, LevelObject{
						ID:    oid,
						Actor: ls.Actor(oid),
						Layer: layer,
						X:     float64(int16(le.Uint16(b[q+2:]))),
						Y:     float64(int16(le.Uint16(b[q+4:]))),
						Z:     float64(int16(le.Uint16(b[q+6:]))),
						RotY:  float64(int16(le.Uint16(b[q+10:]))) * 360 / 0x10000, // idx units, 0x10000 = 360°
						Params: [3]int{
							int(le.Uint16(b[q+14:])), // par1 (+$E)
							int(le.Uint16(b[q+8:])),  // par2 (+$8)
							int(le.Uint16(b[q+12:])), // par3 (+$C)
						},
					})
				}
			case 5: // simple, 8-byte stride
				for j := 0; j < cnt; j++ {
					q := lo + j*8
					if q+8 > len(b) {
						break
					}
					w := le.Uint16(b[q:])
					oid := int(w & 0x1FF)
					lv.Objects = append(lv.Objects, LevelObject{
						ID:     oid,
						Actor:  ls.Actor(oid),
						Layer:  layer,
						X:      float64(int16(le.Uint16(b[q+2:]))),
						Y:      float64(int16(le.Uint16(b[q+4:]))),
						Z:      float64(int16(le.Uint16(b[q+6:]))),
						Params: [3]int{int(w >> 9), 0, 0},
						Simple: true,
					})
				}
			}
		}
	}
	parse(off(misc))
	for a := 0; a < nArea; a++ {
		e := off(areaPtr) + a*12
		if e+12 > len(b) {
			break
		}
		if objT := le.Uint32(b[e:]); objT != 0 && inRAM(objT) {
			parse(off(objT))
		}
	}
	return lv, nil
}

// mangled class name: decimal length prefix + identifier, e.g. "18daObjMc_Metalnet_c"
var classNameRe = regexp.MustCompile(`^[0-9]+([A-Za-z_][A-Za-z0-9_]*)$`)

// ScanActorProfiles sweeps a binary for actor profiles — records whose first
// word is a create-function pointer, whose +4/+6 carry the actor ID twice, and
// whose +$20 points at C++ typeinfo naming the class (the layout the factory
// at $02043098 consumes). Returns actorID -> class name.
func ScanActorProfiles(data []byte, base uint32, isCode func(uint32) bool) map[int]string {
	out := map[int]string{}
	end := base + uint32(len(data))
	inData := func(p uint32) bool { return p >= base && p < end }
	for i := 0; i+0x24 <= len(data); i += 4 {
		create := le.Uint32(data[i:])
		if !isCode(create) {
			continue
		}
		id1, id2 := le.Uint16(data[i+4:]), le.Uint16(data[i+6:])
		if id1 != id2 || id1 >= 1024 {
			continue
		}
		ti := le.Uint32(data[i+0x20:])
		if !inData(ti) || int(ti-base)+8 > len(data) {
			continue
		}
		namePtr := le.Uint32(data[int(ti-base)+4:])
		if !inData(namePtr) {
			continue
		}
		no := int(namePtr - base)
		e := no
		for e < len(data) && data[e] != 0 && e-no < 64 {
			e++
		}
		m := classNameRe.FindStringSubmatch(string(data[no:e]))
		if m == nil {
			continue
		}
		name := strings.TrimSuffix(m[1], "_c")
		if _, dup := out[int(id1)]; !dup {
			out[int(id1)] = name
		}
	}
	return out
}
