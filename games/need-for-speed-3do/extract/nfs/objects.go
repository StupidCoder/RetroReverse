package nfs

import "fmt"

// RoadObjects is the roadside-object block the loader reads right after the
// track head ("RoadObjects" allocation at 0x15A38): 64 object definitions
// followed by ~1000 placements, both 16 bytes. Decoded from the consumers:
// the placement scan in the world-vertex builder (0x16094: world position =
// segment position + the three s16 offsets << 8) and the draw dispatch at
// 0x1A2E8 (def type 1 = 3D model, 4 = upright billboard, 6 = two-anchor cel;
// billboard yaw = -(placement.Yaw<<16) - (segment.Heading<<10), width from
// the def's first extent, height from its third).
type RoadObjects struct {
	Defs       []ObjectDef
	Placements []Placement
}

// ObjectDef is one 16-byte object definition.
type ObjectDef struct {
	Flat bool   // byte 0 bit 0: horizontal (ground) billboard, not upright
	Type byte   // byte 1: 1 = 3D model, 4 = billboard, 6 = two-anchor cel
	Tex  byte   // byte 2: texture index (billboards: root child 1, 4 mips each)
	Tex2 byte   // byte 3: second texture index (scan-path variant)
	W1   Fx16   // +0x4: width (16.16 world units)
	W2   Fx16   // +0x8: width again on every def inspected (far-LOD width?)
	H    Fx16   // +0xC: height (upright) / depth (flat)
	Raw  [16]byte
}

// Placement puts one def somewhere along the track.
type Placement struct {
	Segment  uint32 // record word 0; the array is sorted by it
	Def      byte   // +0x4: index into Defs
	Yaw      byte   // +0x5: extra yaw, 256 = full circle (consumer shifts <<16)
	AnimFlag byte   // +0x8: nonzero selects the animation-frame table
	DX       int16  // +0xA: world-space offsets from the segment position,
	DY       int16  // +0xC: in 1/256 world units (consumer shifts <<8)
	DZ       int16  // +0xE
	Raw      [16]byte
}

// World returns the placement's world position given its segment record.
func (p *Placement) World(seg *Segment) Vec3 {
	return Vec3{
		seg.Pos.X + Fx16(p.DX)<<8,
		seg.Pos.Y + Fx16(p.DY)<<8,
		seg.Pos.Z + Fx16(p.DZ)<<8,
	}
}

func parseRoadObjects(trk []byte) (*RoadObjects, error) {
	defCount := int(be32(trk[defCountOff:]))
	placeCount := int(be32(trk[placeCountOff:]))
	if defCount <= 0 || defCount > 256 || placeCount <= 0 || placeCount > 4096 {
		return nil, fmt.Errorf("nfs: implausible RoadObjects counts %d/%d", defCount, placeCount)
	}
	// block header {u32 ?, u32 size, u32 ?} at headSize; payload follows.
	size := int(be32(trk[headSize+4:]))
	payload := headSize + 12
	if payload+size-12 > len(trk) {
		return nil, fmt.Errorf("nfs: RoadObjects block truncated")
	}
	ro := &RoadObjects{}
	for i := 0; i < defCount; i++ {
		o := payload + 16*i
		var d ObjectDef
		copy(d.Raw[:], trk[o:o+16])
		d.Flat = trk[o]&1 != 0
		d.Type = trk[o+1]
		d.Tex = trk[o+2]
		d.Tex2 = trk[o+3]
		d.W1 = sbe32(trk[o+4:])
		d.W2 = sbe32(trk[o+8:])
		d.H = sbe32(trk[o+12:])
		ro.Defs = append(ro.Defs, d)
	}
	base := payload + 16*defCount
	for i := 0; i < placeCount; i++ {
		o := base + 16*i
		seg := be32(trk[o:])
		if seg == 0xFFFFFFFF { // padding tail
			break
		}
		var p Placement
		copy(p.Raw[:], trk[o:o+16])
		p.Segment = seg
		p.Def = trk[o+4]
		p.Yaw = trk[o+5]
		p.AnimFlag = trk[o+8]
		p.DX = sbe16(trk[o+0xA:])
		p.DY = sbe16(trk[o+0xC:])
		p.DZ = sbe16(trk[o+0xE:])
		ro.Placements = append(ro.Placements, p)
	}
	return ro, nil
}
