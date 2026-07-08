package track

// model.go reimplements the engine's track-model bake $65BEC (Part IV §8): the routine
// the race-setup chain runs once per race to turn the loaded section arrays into the
// polygon model the renderer draws — per-rung records {word0, Lx,Lz,Lh, Rx,Rz,Rh} in
// the buffer at $7ABDA, indexed per section at $7AA1A. The records are CELL-LOCAL:
// each section's plan vertices are the piece-shape (x,z) pairs rotated by the section
// quadrant only ($1BBF2 = -$1BC4A, i.e. camera quadrant 0), in [0,$800] within the
// section's 16x16 grid cell; the per-frame draw ($65EC4) re-places them at the
// camera-relative cell (cell delta x $800, halved) and rotates whole cells by the
// camera quadrant. So the absolute plan position of a vertex is
//
//	world = gridCell(p1) * $800 + local
//
// with no accumulation, no fitting and no spline anywhere. Heights are stored at full
// precision (profile entry + $1C650/$1C718 base, no >>5). The bake also DECIMATES:
// a rung is emitted only when the cross-section profile byte carries the bit-7 "edge"
// marker (cleared vertex flag), it is the section's last rung, the piece is a type-2
// ramp, or the height second-difference is >= $50; rung 0 is always skipped (it is
// the previous section's last rung). Verified byte-exact against the engine running
// the real $65BEC (cmd/modeloracle).
//
// Annotated flow (game.dec.bin, base $E700):
//
//	$65BEC  bake entry: $65DEA (per-section LOD byte table $65E70), records ptr
//	        $66102 = $7ABDA, $1BBC5 = 0, a6 = $1BE70 height strip, then per section:
//	$65C0E  $7AA1A[sec] = record ptr; $66100 = 0; $5C51A advance -> peek NEXT section
//	        via $5FE56: if its piece flags ($1BB4D = a0[1]) bit7: $66100 = $4000, and
//	        if ($1BC44 ^ $1BC32) bit7 also $8000 (ramp joint markers);
//	$65C66  $5C538 retreat + $5FE56 setup CURRENT section; $1BC26 = 0;
//	$65C7A  $1BBF2 = -$1BC4A (NEG.b) — the section quadrant in the world frame;
//	$65C8A  $65756: fill height strip $1BE70 (left/right full-precision heights per
//	        rung) and the vertex flag strip $1BDD0 (bit-7 profile markers clear the
//	        flag = rung is an edge; last rung preset to 1 when sec == $1CA1C finish;
//	        |dh| >= $280 vs previous rung sets the $20 crease bit; a bit-7 marker on
//	        the byte after the last entry hides the last rung = gap piece);
//	$65C9C  emit loop $65CDC over d1 = 0,4,..,2*cnt: skip flagged ($80) rungs and
//	        d1 < 4; emit when d1 >= $1BB5A (last rung; or last-1 when last hidden),
//	        piece is ramp-2 ($1BB4D bit7), p2 == $25 && d1 == $18 (one piece quirk),
//	        or |h[d1]/2 - (h[d1-4]+h[d1+4])/4| >= $50; record = word0 (flag&$3F | sec,
//	        | $66100 unless current piece is ramp-2) then for L (d1) and R (d1+2):
//	        plan offset d2 = $1BB91 +- 2*d1 ($1BB91 = a0[0]+7, + (cnt-1)*4 and minus
//	        when reversed $1BC32/type&$10 — reversed pieces also swap rails), JSR
//	        $5C6C4 (vertex = base + quadrant-rotated LE16 pair), height word $1BE70[d1].
type BakedVert struct {
	X, Z int // cell-local plan, $5C6C4 in the world frame (base 0)
	H    int // full-precision rail height (int16; -32768 = hidden marker)
}

// BakedRung is one emitted record of the game's baked track model.
type BakedRung struct {
	D1    int // engine vertex counter (rung index = D1/4)
	Word0 int // record header: (vertex flag & $3F)<<8 | ramp-joint bits | section
	L, R  BakedVert
}

// BakedModel is the game's own polygon model of one circuit, per section.
// HiddenLast marks sections whose profile carries the end-of-piece bit-7 marker
// ($659F8): the last rung is hidden — a gap piece (e.g. a Stepping Stone edge).
type BakedModel struct {
	Sections   [][]BakedRung
	HiddenLast []bool
}

// bakeState is the persistent engine state across sections: the vertex flag strip
// $1BDD0, the height strip $1BE70 (word-indexed by the engine's d1), and $1BC26.
type bakeState struct {
	flags [256]byte
	hts   [256]int16
	bc26  int
}

func (im *Image) heightNib(a, d2 int, base int16) int16 {
	v := im.u8(a + d2)
	return int16(uint16(((v<<1)&0xE0)|((v&0xF)<<8)) + uint16(base))
}

func (im *Image) heightTwo(a, d2 int, base int16) int16 {
	return int16(uint16((im.u8(a+d2)&0x7F)<<8|im.u8(a+d2+1)) + uint16(base))
}

// heights65756 reproduces $65756 for the current section: fill the height strip and
// the vertex flag strip. bb7F mirrors the engine byte $1BB7F (0 in the bake).
// Returns whether the end-of-piece marker hid the last rung (gap piece).
func (im *Image) heights65756(st *bakeState, n *Node, sec, finish, cnt, bb6A int, bb7F byte) (hiddenLast bool) {
	a4 := handle(im.u16(shapeTab + (2*n.P2)&0xFF))
	a5 := handle(im.u16(shapeTab + (2*n.Attr)&0xFF))
	baseL, baseR := n.X, n.Z
	bb5A := ((cnt - 2) << 1) & 0xFF

	d7 := 0x2E - st.bc26
	if d7 < 0 {
		d7 = 0
	}
	st.bc26 += bb6A

	// preset the last rung's flag: 1 = finish line, else 0 ($657B0..$657D0)
	if sec == finish {
		st.flags[bb5A] = 1
	} else {
		st.flags[bb5A] = 0
	}

	// crease check shared by both read modes ($65872../$65958..): compare with the
	// previous rung's left height, re-deriving it from the shape when negative.
	crease := func(d1, d2, prevStride int, hv int16, read func(a, d2 int, base int16) int16) {
		if st.flags[d1]&0x80 != 0 {
			return
		}
		prev := st.hts[(d1-4)&0xFF]
		if prev < 0 {
			if bb7F == 0 || d1 < 4 {
				return
			}
			prev = read(a4, d2-prevStride, baseL)
		}
		diff := int16(uint16(hv) - uint16(prev))
		if diff < 0 {
			diff = -diff
		}
		if diff >= 0x280 && d1 >= 4 {
			st.flags[d1] |= 0x20
		}
	}

	// Both read modes are the same loop shape; goto labels mirror the engine's flow
	// (body -> tail -> head / last-body / exit), as elsewhere in this project's ports.
	d1, d2, bb7A := 0, 0, 0

	if n.P2&0x80 == 0 { // nibble-packed profile ($1BB79 >= 0)
		if im.u8(a4)&0x80 != 0 { // first entry marker ($65806 BMI $65846)
			st.flags[0] = 0
		}
	nibBody: // $6584C
		hv := im.heightNib(a4, d2, baseL)
		st.hts[d1] = hv
		crease(d1, d2, 1, hv, im.heightNib)
		st.hts[(d1+2)&0xFF] = im.heightNib(a5, d2, baseR)
	nibTail: // $658EC
		d2++
		d1 = (d1 + 4) & 0xFF
		if d1 < bb5A { // loop head $65826
			if im.u8(a4+d2)&0x80 != 0 {
				st.flags[d1] = 0 // $65846
				goto nibBody
			}
			if d2 >= d7 { // beyond LOD: mark hidden, skip the read ($65834)
				st.hts[d1] = -0x8000
				st.hts[(d1+2)&0xFF] = -0x8000
				goto nibTail
			}
			goto nibBody
		}
		if d1 == bb5A { // last rung, read directly with no head check ($658FA)
			goto nibBody
		}
		d2-- // $659EE
		if im.u8(a4+d2)&0x80 != 0 { // end marker: hide the last rung ($659F8)
			st.flags[(d1-4)&0xFF] = 0x80
			hiddenLast = true
			if bb6A >= d7 {
				st.hts[(d1-4)&0xFF] = -0x8000
				st.hts[(d1-2)&0xFF] = -0x8000
			}
		}
		return
	}

	// two-byte profile ($1BB79 < 0)
	if im.u8(a4)&0x80 != 0 { // $65930
		st.flags[0] = 0
	}
twoBody: // $65936
	{
		hv := im.heightTwo(a4, d2, baseL)
		st.hts[d1] = hv
		crease(d1, d2, 2, hv, im.heightTwo)
		st.hts[(d1+2)&0xFF] = im.heightTwo(a5, d2, baseR)
	}
twoTail: // $659CA
	bb7A++
	d1 = (d1 + 4) & 0xFF
	if d1 < bb5A { // loop head $65902
		d2 = 2 * bb7A
		if im.u8(a4+d2)&0x80 != 0 {
			st.flags[d1] = 0 // $65930
			goto twoBody
		}
		if bb7A >= d7 { // $6591E
			st.hts[d1] = -0x8000
			st.hts[(d1+2)&0xFF] = -0x8000
			goto twoTail
		}
		goto twoBody
	}
	if d1 == bb5A { // $659E0: last rung read directly
		d2 = 2 * bb7A
		goto twoBody
	}
	if im.u8(a4+d2)&0x80 != 0 { // $659F0 (d2 = offset of the last entry read)
		st.flags[(d1-4)&0xFF] = 0x80
		hiddenLast = true
		if bb6A >= d7 {
			st.hts[(d1-4)&0xFF] = -0x8000
			st.hts[(d1-2)&0xFF] = -0x8000
		}
	}
	return
}

// vertex5C6C4 reproduces $5C6C4 in the bake's world frame (base $1BB22/$1BB26 = 0):
// read the LE16 (u,v) pair at byte offset d2 in the piece-shape and rotate by the
// section quadrant byte ($1BBF2 = -$1BC4A).
func (im *Image) vertex5C6C4(shp, d2 int, quad byte) (x, z int16) {
	u := int16(uint16(im.s16le(shp + d2)))
	v := int16(uint16(im.s16le(shp + ((d2 + 2) & 0xFF))))
	switch {
	case quad&0x80 == 0 && quad&0x40 == 0:
		return u, v
	case quad&0x80 == 0 && quad&0x40 != 0:
		return int16(0x800) - v, u
	case quad&0x40 == 0:
		return int16(0x800) - u, int16(0x800) - v
	default:
		return v, int16(0x800) - u
	}
}

// Bake reproduces the whole $65BEC bake for track id (the track must be the one the
// Spine decode describes; the section arrays come from the verified pure-Go decode).
func (im *Image) Bake(t *Track) BakedModel {
	st := &bakeState{}
	model := BakedModel{
		Sections:   make([][]BakedRung, len(t.Nodes)),
		HiddenLast: make([]bool, len(t.Nodes)),
	}
	n := len(t.Nodes)
	for sec := 0; sec < n; sec++ {
		// $66100 ramp-joint flags from the NEXT section's piece header
		nx := &t.Nodes[(sec+1)%n]
		shpN := handle(im.u16(dataBase + 2*(nx.Type&0x0F)))
		fl66100 := 0
		if im.u8(shpN+1)&0x80 != 0 {
			fl66100 = 0x4000
			offN := im.u8(shpN)
			v := im.u8(shpN + offN + 1)
			c44 := ((v >> 1) | ((v & 1) << 7)) & 0x80
			c32 := (nx.Type & 0x10) << 3
			if (c44^c32)&0x80 != 0 {
				fl66100 = 0x8000 // MOVE.b (replaces the $40), not OR ($65C5E)
			}
		}

		// current section setup ($5FE56 essentials)
		cur := &t.Nodes[sec]
		shp := handle(im.u16(dataBase + 2*(cur.Type&0x0F)))
		off := im.u8(shp)
		cnt := im.u8(shp + off)
		bb59 := (cnt << 1) & 0xFF
		bb5A := ((cnt - 2) << 1) & 0xFF
		bb6A := ((cnt >> 1) - 1) & 0xFF
		ramp := im.u8(shp + 1) // $1BB4D
		raised := cur.Type&0x10 != 0
		quad := byte(-(byte(cur.Type & 0xC0))) // $1BBF2 = NEG.b $1BC4A
		bb91 := (off + 7) & 0xFF
		if raised {
			bb91 = (bb91 + (((cnt - 1) << 2) & 0xFF)) & 0xFF
		}

		st.bc26 = 0 // $65C72
		model.HiddenLast[sec] = im.heights65756(st, cur, sec, t.FinishIdx, cnt, bb6A, 0)

		var recs []BakedRung
		d1 := 0
		for {
			f := st.flags[d1]
			if f&0x80 == 0 {
				st.flags[d1] = 0x80
				if d1 >= 4 {
					emit := false
					switch {
					case int8(d1) >= int8(bb5A):
						emit = true
					default:
						d3 := bb5A
						if st.flags[d3]&0x80 != 0 { // last rung hidden (gap)
							d3 = (d3 - 4) & 0xFF
							if d1 == d3 {
								emit = true
							}
						}
						if !emit && ramp&0x80 != 0 {
							emit = true
						}
						if !emit && cur.P2 == 0x25 && d1 == 0x18 {
							emit = true
						}
						if !emit {
							d3v := uint16(st.hts[d1]) >> 1
							d4v := (uint16(st.hts[(d1-4)&0xFF]) + uint16(st.hts[(d1+4)&0xFF])) >> 2
							diff := int16(d3v - d4v)
							if diff < 0 {
								diff = -diff
							}
							if diff >= 0x50 {
								emit = true
							}
						}
					}
					if emit {
						w0 := int(f&0x3F) << 8
						if ramp&0x80 == 0 {
							w0 |= fl66100
						}
						w0 = (w0 & 0xFF00) | sec
						var vv [2]BakedVert
						for h := 0; h < 2; h++ {
							dv := (d1 + 2*h) & 0xFF
							var d2 int
							if raised {
								d2 = (bb91 - ((2 * dv) & 0xFF)) & 0xFF
							} else {
								d2 = (bb91 + ((2 * dv) & 0xFF)) & 0xFF
							}
							x, z := im.vertex5C6C4(shp, d2, quad)
							vv[h] = BakedVert{X: int(x), Z: int(z), H: int(st.hts[dv])}
						}
						recs = append(recs, BakedRung{D1: d1, Word0: w0, L: vv[0], R: vv[1]})
					}
				}
			}
			d1 = (d1 + 4) & 0xFF
			if d1 == bb59 {
				break
			}
		}
		model.Sections[sec] = recs
	}
	return model
}

// RungAbs is one rung of a section in ABSOLUTE plan coordinates: the engine's baked
// cell-local vertex ($5C6C4 with section-quadrant rotation, engine order — reversed
// pieces run their shape backwards which also swaps the rails) placed at the section's
// 16x16 grid cell (p1), one cell = $800 units. Heights are full precision (profile
// entry + $1C650/$1C718 base), exactly the engine's $1BE70 strip values.
type RungAbs struct {
	LX, LZ, RX, RZ int
	HL, HR         int
	Edge           bool // the bake emitted a record for this rung (a drawn polygon edge)
	Hidden         bool // last rung of a gap piece ($659F8) — the game does not draw it
	Crease         bool // record flag $20: silhouette crease (|dh| >= $280)
	Finish         bool // record flag 1: the finish line rung
}

// Geometry returns the absolute per-rung geometry of every section (all rungs, not
// just the decimated baked records; Edge marks the baked ones). The placement is the
// game's own: the per-frame draw ($65EC4) positions each section's cell-local record
// at (cell delta) * $800, so world = gridCell * $800 + local.
func (im *Image) Geometry(t *Track) [][]RungAbs {
	model := im.Bake(t)
	out := make([][]RungAbs, len(t.Nodes))
	for sec := range t.Nodes {
		n := &t.Nodes[sec]
		nib := n.Type & 0x0F
		shp := handle(im.u16(dataBase + 2*nib))
		off := im.u8(shp)
		cnt := im.u8(shp + off)
		rungs := cnt / 2
		raised := n.Type&0x10 != 0
		quad := -(byte(n.Type & 0xC0)) // $1BBF2 = NEG.b $1BC4A
		bb91 := (off + 7) & 0xFF
		if raised {
			bb91 = (bb91 + (((cnt - 1) << 2) & 0xFF)) & 0xFF
		}
		a4 := handle(im.u16(shapeTab + (2*n.P2)&0xFF))
		a5 := handle(im.u16(shapeTab + (2*n.Attr)&0xFF))

		edges := map[int]BakedRung{}
		for _, r := range model.Sections[sec] {
			edges[r.D1/4] = r
		}

		rs := make([]RungAbs, rungs)
		for k := 0; k < rungs; k++ {
			var r RungAbs
			// plan: engine emit order (d1 = 4k), vertex L = 2k, R = 2k+1
			for h := 0; h < 2; h++ {
				dv := 2*k + h
				var d2 int
				if raised {
					d2 = (bb91 - ((4 * dv) & 0xFF)) & 0xFF
				} else {
					d2 = (bb91 + ((4 * dv) & 0xFF)) & 0xFF
				}
				x, z := im.vertex5C6C4(shp, d2, quad)
				ax := n.PlanX*0x800 + int(x)
				az := n.PlanY*0x800 + int(z)
				if h == 0 {
					r.LX, r.LZ = ax, az
				} else {
					r.RX, r.RZ = ax, az
				}
			}
			// heights: the $65756 strip is filled forward in profile order
			if n.P2&0x80 == 0 {
				r.HL = int(im.heightNib(a4, k, n.X))
				r.HR = int(im.heightNib(a5, k, n.Z))
			} else {
				r.HL = int(im.heightTwo(a4, 2*k, n.X))
				r.HR = int(im.heightTwo(a5, 2*k, n.Z))
			}
			if br, ok := edges[k]; ok {
				r.Edge = true
				r.Crease = br.Word0&0x2000 != 0
				r.Finish = br.Word0&0x0100 != 0
			}
			if k == rungs-1 && model.HiddenLast[sec] {
				r.Hidden = true
			}
			rs[k] = r
		}
		out[sec] = rs
	}
	return out
}
