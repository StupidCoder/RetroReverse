package mkds

import (
	"encoding/binary"
	"fmt"
)

// NKM ("NKMD") is Mario Kart DS's course-map file — everything about a track that
// is not geometry: object placements, routes, start/respawn points, checkpoints,
// the CPU racers' driving line, item-probe points, trigger areas and cameras. A
// header (magic, u16 version=37, u16 headerSize=$4C) is followed by 17 u32 section
// offsets (relative to headerSize); each section is a 4-char magic + u16 entry
// count + fixed-size entries. Entry sizes verified by the offset deltas of the
// real files (e.g. OBJI $3C, CPOI $24, EPOI $18, CAME $48).
//
// Coordinates are fx32 world units (1.0 = $1000), matching the course model's
// vertices after its POSSCALE — verified by overlaying checkpoints on the model.

var le = binary.LittleEndian

type Vec3 struct{ X, Y, Z float64 }

// NKM is a decoded course map (the sections the analysis uses; the rest are kept
// as raw counts).
type NKM struct {
	Objects     []Object     // OBJI
	Paths       []Path       // PATH/POIT routes (moving objects, cameras)
	Starts      []Start      // KTPS
	Respawns    []Respawn    // KTPJ
	Checkpoints []Checkpoint // CPOI
	CheckPaths  []SectPath   // CPAT
	ItemPoints  []ItemPoint  // IPOI
	ItemPaths   []SectPath   // IPAT
	EnemyPoints []EnemyPoint // EPOI
	EnemyPaths  []SectPath   // EPAT
	NCame       int
	NArea       int
}

// Object is one OBJI entry: a placed course object.
type Object struct {
	Pos     Vec3
	Rot     Vec3
	Scale   Vec3
	ID      int  // object type (the engine's object table index)
	RouteID int  // PATH the object follows (-1 none)
	TT      bool // present in time-trial mode
}

// Path groups PATH+POIT: a route of points.
type Path struct {
	Loop   bool
	Points []Vec3
}

// Start is a KTPS entry: the grid start position/orientation.
type Start struct {
	Pos Vec3
	Rot Vec3
}

// Respawn is a KTPJ entry: where Lakitu drops you.
type Respawn struct {
	Pos Vec3
	Rot Vec3
}

// Checkpoint is a CPOI entry: a gate line (X1,Z1)-(X2,Z2) on the track plane.
type Checkpoint struct {
	X1, Z1, X2, Z2 float64
	KeyID          int // -1 = ordinary; 0 = lap line; >0 = key checkpoint (shortcut guard)
	Respawn        int
}

// EnemyPoint is an EPOI entry: one node of the CPU drive line.
type EnemyPoint struct {
	Pos    Vec3
	Radius float64 // lateral tolerance around the line
	Drift  int
}

// ItemPoint is an IPOI entry: one node of the item-probe line (red shells etc.).
type ItemPoint struct {
	Pos    Vec3
	Radius float64
}

// SectPath is a CPAT/IPAT/EPAT entry: a section of consecutive points plus the
// indices of the sections that follow/precede it (branching: -1 = none).
type SectPath struct {
	Start, Len int
	Next, Prev [3]int
}

func fx(v uint32) float64 { return float64(int32(v)) / 4096 }

// ParseNKM decodes an NKMD blob.
func ParseNKM(data []byte) (*NKM, error) {
	if len(data) < 8 || string(data[:4]) != "NKMD" {
		return nil, fmt.Errorf("mkds: not an NKMD file")
	}
	hdrSize := int(le.Uint16(data[6:]))
	sect := map[string][]byte{}
	counts := map[string]int{}
	for o := 8; o < hdrSize; o += 4 {
		p := hdrSize + int(le.Uint32(data[o:]))
		if p+8 > len(data) {
			continue
		}
		magic := string(data[p : p+4])
		n := int(le.Uint16(data[p+4:]))
		sect[magic] = data[p+8:]
		counts[magic] = n
	}
	nkm := &NKM{NCame: counts["CAME"], NArea: counts["AREA"]}

	if b, n := sect["OBJI"], counts["OBJI"]; b != nil {
		for i := 0; i < n && (i+1)*0x3C <= len(b); i++ {
			e := b[i*0x3C:]
			o := Object{
				Pos:   vec3(e, 0),
				Rot:   vec3(e, 12),
				Scale: vec3(e, 24),
				ID:    int(le.Uint16(e[36:])),
				TT:    le.Uint32(e[56:]) != 0,
			}
			o.RouteID = int(int16(le.Uint16(e[38:])))
			nkm.Objects = append(nkm.Objects, o)
		}
	}
	// PATH entries are {u8 index, u8 loop, u16 numPoints}; POIT are {pos, ...} of 0x14.
	if pb, pn := sect["PATH"], counts["PATH"]; pb != nil {
		tb := sect["POIT"]
		pt := 0
		for i := 0; i < pn && (i+1)*4 <= len(pb); i++ {
			e := pb[i*4:]
			p := Path{Loop: e[1] != 0}
			np := int(le.Uint16(e[2:]))
			for j := 0; j < np && (pt+1)*0x14 <= len(tb); j++ {
				p.Points = append(p.Points, vec3(tb[pt*0x14:], 0))
				pt++
			}
			nkm.Paths = append(nkm.Paths, p)
		}
	}
	if b, n := sect["KTPS"], counts["KTPS"]; b != nil {
		for i := 0; i < n && (i+1)*0x1C <= len(b); i++ {
			nkm.Starts = append(nkm.Starts, Start{Pos: vec3(b[i*0x1C:], 0), Rot: vec3(b[i*0x1C:], 12)})
		}
	}
	if b, n := sect["KTPJ"], counts["KTPJ"]; b != nil {
		for i := 0; i < n && (i+1)*0x20 <= len(b); i++ {
			nkm.Respawns = append(nkm.Respawns, Respawn{Pos: vec3(b[i*0x20:], 0), Rot: vec3(b[i*0x20:], 12)})
		}
	}
	if b, n := sect["CPOI"], counts["CPOI"]; b != nil {
		for i := 0; i < n && (i+1)*0x24 <= len(b); i++ {
			e := b[i*0x24:]
			nkm.Checkpoints = append(nkm.Checkpoints, Checkpoint{
				X1: fx(le.Uint32(e)), Z1: fx(le.Uint32(e[4:])),
				X2: fx(le.Uint32(e[8:])), Z2: fx(le.Uint32(e[12:])),
				KeyID:   int(int16(le.Uint16(e[24:]))),
				Respawn: int(e[27]),
			})
		}
	}
	nkm.CheckPaths = sectPaths(sect["CPAT"], counts["CPAT"])
	if b, n := sect["IPOI"], counts["IPOI"]; b != nil {
		for i := 0; i < n && (i+1)*0x14 <= len(b); i++ {
			e := b[i*0x14:]
			nkm.ItemPoints = append(nkm.ItemPoints, ItemPoint{Pos: vec3(e, 0), Radius: fx(le.Uint32(e[12:]))})
		}
	}
	nkm.ItemPaths = sectPaths(sect["IPAT"], counts["IPAT"])
	if b, n := sect["EPOI"], counts["EPOI"]; b != nil {
		for i := 0; i < n && (i+1)*0x18 <= len(b); i++ {
			e := b[i*0x18:]
			nkm.EnemyPoints = append(nkm.EnemyPoints, EnemyPoint{
				Pos: vec3(e, 0), Radius: fx(le.Uint32(e[12:])), Drift: int(int16(le.Uint16(e[16:]))),
			})
		}
	}
	nkm.EnemyPaths = sectPaths(sect["EPAT"], counts["EPAT"])
	return nkm, nil
}

func vec3(b []byte, off int) Vec3 {
	return Vec3{fx(le.Uint32(b[off:])), fx(le.Uint32(b[off+4:])), fx(le.Uint32(b[off+8:]))}
}

// sectPaths decodes a CPAT/IPAT/EPAT list: {u16 start, u16 len, s8 next[3], s8 prev[3], u16 order}.
func sectPaths(b []byte, n int) []SectPath {
	var out []SectPath
	for i := 0; i < n && (i+1)*12 <= len(b); i++ {
		e := b[i*12:]
		sp := SectPath{Start: int(le.Uint16(e)), Len: int(le.Uint16(e[2:]))}
		for k := 0; k < 3; k++ {
			sp.Next[k] = int(int8(e[4+k]))
			sp.Prev[k] = int(int8(e[7+k]))
		}
		out = append(out, sp)
	}
	return out
}
