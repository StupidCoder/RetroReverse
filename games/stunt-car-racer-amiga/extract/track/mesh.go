package track

// mesh.go reimplements the colour and decal-line assignment of the game's own
// pre-race track preview ($604B4): the per-cell scene walk $602E4/$60324 computes
// every section at full rung resolution ($6521C -> $65756 strips), the emit
// $672CC writes one display-list strip per rung pair with a 4-byte descriptor
// {section, type, flag, piece}, and the flush $68F28 fills and strokes each strip:
//
//	road fill   $68D82: palette 1, or 2 when the section number is odd
//	            ($691C4 BTST #0 of the descriptor's section byte); 0 when the
//	            strip's flag byte carries the crease bit $20
//	wall fills  $68A56/$68BEC: palette $F (white), or $A (red, the byte at
//	            $691B2) when the section is odd; $9 (the byte at $691B3) when
//	            the cell is viewed from its back side ($1BB66 — view-dependent,
//	            so the static mesh keeps $F/$A for the outer faces and $9 for
//	            the inner faces)
//	curb strokes  $689CE/$68A12: the road-edge lengthwise lines, colour = the
//	            descriptor type byte — 9, or 3 on alternating rungs: $67120
//	            parity ((($1BBDC<<1) ^ d1) & 4) with the per-section phase
//	            $1BBDC seeded from bit 7 of the section's right-profile handle
//	            byte ($1C524[sec], $5FE82 ROXL) — i.e. type(k) = (seed^k)&1 ? 3 : 9
//	wall verticals $6892A/$6897C: rail-to-ground lines at rungs where the parity
//	            is even (type 9) or the rung is a drawn edge ($67284 gate);
//	            stroked in palette 9 only on type-9 strips
//	cross lines  emitted at rungs whose flag-strip byte has bit 7 clear (the
//	            baked polygon edges; the preview forces the first and last rung
//	            of every section visible, $6045C); stroked only as the finish
//	            line, in palette $F ($688FC via $1BB61 bit 0)
//
// The preview flag $1BB68=$80 makes $654C2 emit a wall-bottom vertex for every
// rail point at height $200 — the ground plane — so the walls drop to the ground
// along the whole circuit. Verified against the engine by cmd/coloracle.

// Palette returns the preview/race palette: the 16 words at $2B974 (staged via
// $1BA70/$64470 and pushed to the copper COLOR block by $1BA82, which maps each
// 4-bit channel c -> 2c | (c!=0)). Channels are the displayed 4-bit values.
func (im *Image) Palette() [16][3]uint8 {
	var pal [16][3]uint8
	for i := 0; i < 16; i++ {
		w := im.u16(palTable + 2*i)
		for c := 0; c < 3; c++ {
			v := uint8(w>>(8-4*c)) & 0xF
			v <<= 1
			if v != 0 {
				v |= 1
			}
			pal[i][c] = v
		}
	}
	return pal
}

const palTable = 0x2B974

// MeshRung is the preview renderer's decisions for one rung k of a section: the
// strip ENDING at rung k (the quad between rungs k-1 and k, k >= 1) plus the
// lateral lines AT rung k. Type is the descriptor byte $1D — the curb-stroke
// colour of the strip. All colours are 4-bit palette indexes into Palette().
type MeshRung struct {
	Type    byte // 9 or 3: curb colour, alternating per rung with per-section seed
	RoadPal byte // road fill of strip (k-1,k): 1/2 by section parity; 0 when the
	// span this strip belongs to ends at a crease rung (fills are per batch)
	WallPal byte // outer wall fill of strip (k-1,k): $F/$A by section parity
	Cross   bool // lateral road line emitted at rung k (flag bit 7 clear)
	Crease  bool // flag bit 5 at rung k
	Verts   bool // wall verticals emitted at rung k (parity even or Cross)
	VertS   bool // wall verticals stroked (palette 9): Verts and Type != 3
	Finish  bool // flag bit 0: the finish-line rung (cross line stroked $F)
}

// MeshSection is one section's rungs (index 0 duplicates the previous section's
// last rung; its strip has lateral lines only).
type MeshSection struct {
	Rungs []MeshRung
}

// Mesh reproduces the preview's per-rung colour/decal algorithm for every
// section of the track. Geometry (rail and wall-bottom positions) comes from
// Geometry(); this carries the colours the flush would assign.
func (im *Image) Mesh(t *Track) TrackMesh {
	out := TrackMesh{Sections: make([]MeshSection, len(t.Nodes))}

	// the curb-phase seeds: the loader ($5AE46) accumulates a two-state phase
	// across sections ($5B046: $1BBDC = ($1BBDC + cnt-2) & 2, from 0) and stores
	// each section's phase bit as bit 7 of $1C524[sec] ($5B010/$5AF7C) — this is
	// what keeps the yellow/dark curb dashes continuous across section joints
	seeds := make([]int, len(t.Nodes))
	phase := 0
	for s := range t.Nodes {
		seeds[s] = phase >> 1
		nib := t.Nodes[s].Type & 0x0F
		shp := handle(im.u16(dataBase + 2*nib))
		off := im.u8(shp)
		cnt := im.u8(shp + off)
		phase = (phase + cnt - 2) & 2
	}

	for sec := range t.Nodes {
		n := &t.Nodes[sec]
		nib := n.Type & 0x0F
		shp := handle(im.u16(dataBase + 2*nib))
		off := im.u8(shp)
		cnt := im.u8(shp + off)
		rungs := cnt / 2

		// the flag strip: the previous cell's emit leaves every rung consumed
		// ($670EA marks $80), then $65756 re-marks this section's edge rungs
		st := &bakeState{}
		for i := range st.flags {
			st.flags[i] = 0x80
		}
		finish := t.FinishIdx
		im.heights65756(st, n, sec, finish, cnt, rungs-1, 0)
		// the preview walk clears the first and last rung's flag byte outright
		// ($6045C MOVE.b #0) — forcing them visible and, as a side effect, wiping
		// any crease bit and the finish-line bit there: the preview never draws
		// the white finish stroke ($688FC is a race-view feature)
		st.flags[0] = 0
		last := ((cnt - 2) << 1) & 0xFF
		st.flags[last] = 0

		seed := seeds[sec] // $5FE82: ROXL of $1C524[sec]<<1
		parity := func(k int) int { return (seed ^ k) & 1 }

		ms := MeshSection{Rungs: make([]MeshRung, rungs)}
		wall := byte(0xF)
		road := byte(1)
		if sec&1 != 0 {
			wall, road = 0xA, 2
		}
		for k := 0; k < rungs; k++ {
			f := st.flags[(4*k)&0xFF]
			r := MeshRung{
				Type:    9,
				RoadPal: road,
				WallPal: wall,
				Cross:   f&0x80 == 0,
				Crease:  f&0x20 != 0,
				Finish:  f&0x01 != 0,
			}
			if parity(k) != 0 {
				r.Type = 3
			}
			r.Verts = parity(k) == 0 || r.Cross
			r.VertS = r.Verts && r.Type != 3
			ms.Rungs[k] = r
		}
		// the flush fills the road in batches: each span of strips ending at an
		// edge rung is one polygon, coloured by THAT rung's descriptor — so a
		// crease bit blackens (palette 0) the whole span leading up to it
		for k := rungs - 1; k >= 1; k-- {
			if !ms.Rungs[k].Cross {
				continue
			}
			pal := ms.Rungs[k].RoadPal
			if ms.Rungs[k].Crease {
				pal = 0
			}
			for j := k; j >= 1 && (j == k || !ms.Rungs[j].Cross); j-- {
				ms.Rungs[j].RoadPal = pal
			}
		}
		out.Sections[sec] = ms
	}
	return out
}

// TrackMesh is the whole-track preview mesh colouring (name distinct from the method).
type TrackMesh struct {
	Sections []MeshSection
}
