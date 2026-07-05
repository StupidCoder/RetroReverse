package nitro

import "image/color"

// A shape's geometry is a display list: the raw command stream of the DS geometry
// engine (the "G3"/GX commands), stored exactly as the game DMA-feeds it to the
// hardware FIFO. Each 32-bit word packs four command IDs, followed by each
// command's parameter words in order. The commands that matter for geometry:
//
//	$14 MTX_RESTORE  — switch the current matrix to a stack slot (joint binding)
//	$20 COLOR        — vertex colour (BGR555)
//	$21 NORMAL       — packed 10-bit normal
//	$22 TEXCOORD     — packed 12.4 texture coordinate
//	$23 VTX_16       — full 16-bit x,y,z vertex (two words)
//	$24 VTX_10       — packed 10-bit vertex
//	$25/$26/$27 VTX_XY/XZ/YZ — two coordinates, third kept from the last vertex
//	$28 VTX_DIFF     — 10-bit delta from the last vertex
//	$40 BEGIN_VTXS   — primitive type: 0 tris, 1 quads, 2 tri-strip, 3 quad-strip
//	$41 END_VTXS
//
// Vertices are fx12 (1.0 = $1000) model-space values; matrices from the stack the
// SBC set up (RunSBC) place each batch.

// Vertex is one emitted vertex in model space (after matrix application).
type Vertex struct {
	X, Y, Z float64
	U, V    float64 // texels (12.4 → texel units)
	C       color.NRGBA
	J    int // source matrix-stack slot (joint index for skinned export)
}

// Tri is one triangle.
type Tri struct {
	V   [3]Vertex
	Mat int // material index active when emitted
}

// dlParamCount gives each G3 command's parameter-word count.
func dlParamCount(cmd byte) int {
	switch cmd {
	case 0x16:
		return 16
	case 0x17:
		return 12
	case 0x18:
		return 16
	case 0x19:
		return 12
	case 0x1A:
		return 9
	case 0x1B, 0x1C:
		return 3
	case 0x23:
		return 2
	case 0x34:
		return 32
	case 0x00, 0x11, 0x15, 0x41:
		return 0
	default:
		return 1
	}
}

// jointOf finds the matrix-stack slot the starting matrix came from, so each
// vertex can be tagged with its source joint for skinned export (MTX_RESTORE
// switches it; other matrix ops keep the last slot).
func jointOf(cur Mat43, stack []Mat43) int {
	for i, m := range stack {
		if m == cur {
			return i
		}
	}
	return 0
}

// DecodeDL runs a display list, emitting triangles. stack is the joint-matrix
// stack (from RunSBC); cur is the starting current matrix; mat the active material
// index (both as the SBC left them before the shape's draw call).
func DecodeDL(dl []byte, stack []Mat43, cur Mat43, mat int) []Tri {
	var tris []Tri
	var prim int
	var verts []Vertex
	var last Vertex
	var curColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	var u, v float64

	// GX matrix pipeline: a current position matrix (cur), a push/pop stack, and the
	// 32 addressable slots (stack, passed in). Models that build transforms inside
	// the display list (SM64DS stages such as Big Boo's Haunt, the cave) rely on
	// these; simpler models (the castle floors) only ever MTX_RESTORE slot 0.
	var mstack []Mat43 // push/pop stack
	joint := jointOf(cur, stack)
	mtxMode := 2 // 0 proj, 1 position, 2 position+vector, 3 texture
	fx := func(v uint32) float64 { return float64(int32(v)) / 4096 }
	load43 := func(p []uint32) Mat43 {
		return Mat43{fx(p[0]), fx(p[1]), fx(p[2]), fx(p[3]), fx(p[4]), fx(p[5]),
			fx(p[6]), fx(p[7]), fx(p[8]), fx(p[9]), fx(p[10]), fx(p[11])}
	}
	load44 := func(p []uint32) Mat43 { // 4×4: drop each column's w component
		return Mat43{fx(p[0]), fx(p[1]), fx(p[2]), fx(p[4]), fx(p[5]), fx(p[6]),
			fx(p[8]), fx(p[9]), fx(p[10]), fx(p[12]), fx(p[13]), fx(p[14])}
	}
	isPos := func() bool { return mtxMode != 3 } // apply to cur unless in texture mode

	emit := func() {
		n := len(verts)
		switch prim {
		case 0: // separate triangles
			if n == 3 {
				tris = append(tris, Tri{V: [3]Vertex{verts[0], verts[1], verts[2]}, Mat: mat})
				verts = verts[:0]
			}
		case 1: // separate quads → two tris
			if n == 4 {
				tris = append(tris,
					Tri{V: [3]Vertex{verts[0], verts[1], verts[2]}, Mat: mat},
					Tri{V: [3]Vertex{verts[0], verts[2], verts[3]}, Mat: mat})
				verts = verts[:0]
			}
		case 2: // triangle strip (alternating winding)
			if n >= 3 {
				a, b, c := verts[n-3], verts[n-2], verts[n-1]
				if n%2 == 1 {
					tris = append(tris, Tri{V: [3]Vertex{a, b, c}, Mat: mat})
				} else {
					tris = append(tris, Tri{V: [3]Vertex{b, a, c}, Mat: mat})
				}
			}
		case 3: // quad strip: vertices v0 v1 v2 v3… form quads (v0,v1,v3,v2) — the last
			// two swap, so the spatial cycle is a→b→c→d with c the NEWEST vertex.
			if n >= 4 && n%2 == 0 {
				a, b, c, d := verts[n-4], verts[n-3], verts[n-1], verts[n-2]
				tris = append(tris,
					Tri{V: [3]Vertex{a, b, c}, Mat: mat},
					Tri{V: [3]Vertex{a, c, d}, Mat: mat})
			}
		}
	}

	addVertex := func(x, y, z float64) {
		wx, wy, wz := cur.Apply(x, y, z)
		last = Vertex{X: x, Y: y, Z: z} // model-space memory for VTX_XY/DIFF forms
		vv := Vertex{X: wx, Y: wy, Z: wz, U: u, V: v, C: curColor, J: joint}
		verts = append(verts, vv)
		emit()
	}

	p := 0
	next := func() uint32 {
		if p+4 > len(dl) {
			p = len(dl)
			return 0
		}
		w := le.Uint32(dl[p:])
		p += 4
		return w
	}

	for p < len(dl) {
		cmdWord := next()
		cmds := [4]byte{byte(cmdWord), byte(cmdWord >> 8), byte(cmdWord >> 16), byte(cmdWord >> 24)}
		for _, cmd := range cmds {
			np := dlParamCount(cmd)
			params := make([]uint32, np)
			for i := range params {
				params[i] = next()
			}
			switch cmd {
			case 0x10: // MTX_MODE
				mtxMode = int(params[0] & 3)
			case 0x11: // MTX_PUSH
				if isPos() {
					mstack = append(mstack, cur)
				}
			case 0x12: // MTX_POP (signed 6-bit level count)
				if isPos() {
					n := int(int8(params[0]<<2) >> 2)
					if n < 0 {
						n = -n
					}
					if n > 0 && n <= len(mstack) {
						cur = mstack[len(mstack)-n]
						mstack = mstack[:len(mstack)-n]
					}
				}
			case 0x13: // MTX_STORE
				if isPos() {
					if idx := int(params[0] & 0x1F); idx < len(stack) {
						stack[idx] = cur
					}
				}
			case 0x14: // MTX_RESTORE
				idx := int(params[0] & 0x1F)
				if isPos() && idx < len(stack) {
					cur = stack[idx]
					joint = idx
				}
			case 0x15: // MTX_IDENTITY
				if isPos() {
					cur = Identity43()
				}
			case 0x16: // MTX_LOAD_4x4
				if isPos() {
					cur = load44(params)
				}
			case 0x17: // MTX_LOAD_4x3
				if isPos() {
					cur = load43(params)
				}
			case 0x18: // MTX_MULT_4x4
				if isPos() {
					cur = cur.Mul(load44(params))
				}
			case 0x19: // MTX_MULT_4x3
				if isPos() {
					cur = cur.Mul(load43(params))
				}
			case 0x1A: // MTX_MULT_3x3 (rotation only)
				if isPos() {
					p := params
					cur = cur.Mul(Mat43{fx(p[0]), fx(p[1]), fx(p[2]), fx(p[3]), fx(p[4]), fx(p[5]), fx(p[6]), fx(p[7]), fx(p[8]), 0, 0, 0})
				}
			case 0x1B: // MTX_SCALE
				if isPos() {
					cur = cur.Mul(Mat43{fx(params[0]), 0, 0, 0, fx(params[1]), 0, 0, 0, fx(params[2]), 0, 0, 0})
				}
			case 0x1C: // MTX_TRANS
				if isPos() {
					cur = cur.Mul(Mat43{1, 0, 0, 0, 1, 0, 0, 0, 1, fx(params[0]), fx(params[1]), fx(params[2])})
				}
			case 0x20: // COLOR (BGR555)
				c := uint16(params[0])
				curColor = bgr555(c)
			case 0x21: // NORMAL — ignored (no lighting model)
			case 0x22: // TEXCOORD, 12.4 fixed, texel units
				u = float64(int16(params[0]&0xFFFF)) / 16
				v = float64(int16(params[0]>>16)) / 16
			case 0x23: // VTX_16
				x := float64(int16(params[0]&0xFFFF)) / 4096
				y := float64(int16(params[0]>>16)) / 4096
				z := float64(int16(params[1]&0xFFFF)) / 4096
				addVertex(x, y, z)
			case 0x24: // VTX_10 (10 bits each, 4.6 fixed → <<6 /4096)
				w := params[0]
				x := float64(int16(w<<6)>>6) * 64 / 4096
				y := float64(int16((w>>10)<<6)>>6) * 64 / 4096
				z := float64(int16((w>>20)<<6)>>6) * 64 / 4096
				addVertex(x, y, z)
			case 0x25: // VTX_XY
				x := float64(int16(params[0]&0xFFFF)) / 4096
				y := float64(int16(params[0]>>16)) / 4096
				addVertex(x, y, last.Z)
			case 0x26: // VTX_XZ
				x := float64(int16(params[0]&0xFFFF)) / 4096
				z := float64(int16(params[0]>>16)) / 4096
				addVertex(x, last.Y, z)
			case 0x27: // VTX_YZ
				y := float64(int16(params[0]&0xFFFF)) / 4096
				z := float64(int16(params[0]>>16)) / 4096
				addVertex(last.X, y, z)
			case 0x28: // VTX_DIFF: 10-bit signed deltas in raw fx12 units (±0.125)
				w := params[0]
				dx := float64(int16(w<<6)>>6) / 4096
				dy := float64(int16((w>>10)<<6)>>6) / 4096
				dz := float64(int16((w>>20)<<6)>>6) / 4096
				addVertex(last.X+dx, last.Y+dy, last.Z+dz)
			case 0x40: // BEGIN_VTXS
				prim = int(params[0] & 3)
				verts = verts[:0]
			case 0x41: // END_VTXS
				verts = verts[:0]
			}
		}
	}
	return tris
}

// RunSBC interprets a model's scene bytecode far enough for static geometry: it
// computes each node's world matrix (NODEDESC chains node×parent and can store to
// a matrix-stack slot) and records, for every shape draw (SHP), the material and
// current matrix in effect. Animation-only opcodes are skipped by size.
type ShapeDraw struct {
	Shape, Mat int
	M          Mat43
	Stack      []Mat43
}

func RunSBC(m Model) []ShapeDraw {
	world := make([]Mat43, len(m.Nodes)) // per-node world matrix
	for i := range world {
		world[i] = Identity43()
	}
	stack := make([]Mat43, 32)
	for i := range stack {
		stack[i] = Identity43()
	}
	cur := Identity43()
	curMat := 0
	var draws []ShapeDraw

	sbc := m.SBC
	p := 0
	for p < len(sbc) {
		op := sbc[p]
		base := op & 0x1F
		opt := op >> 5
		p++
		switch base {
		case 0x01: // RET
			p = len(sbc)
		case 0x02: // NODE: id, visibility
			p += 2
		case 0x03: // MTX: restore stack slot
			if p < len(sbc) {
				idx := int(sbc[p])
				if idx < len(stack) {
					cur = stack[idx]
				}
			}
			p++
		case 0x04: // MAT
			if p < len(sbc) {
				curMat = int(sbc[p])
			}
			p++
		case 0x05: // SHP
			if p < len(sbc) {
				draws = append(draws, ShapeDraw{Shape: int(sbc[p]), Mat: curMat, M: cur, Stack: append([]Mat43(nil), stack...)})
			}
			p++
		case 0x06: // NODEDESC: node, parent, flags [, stackID [, restoreID]]
			if p+3 > len(sbc) {
				p = len(sbc)
				break
			}
			node := int(sbc[p])
			parent := int(sbc[p+1])
			p += 3
			var stackID = -1
			if opt&1 != 0 { // has stack store slot
				stackID = int(sbc[p])
				p++
			}
			if opt&2 != 0 { // has restore slot
				if idx := int(sbc[p]); idx < len(stack) {
					cur = stack[idx]
				}
				p++
			}
			if node < len(m.Nodes) {
				pm := Identity43()
				if parent < len(world) && parent != node {
					pm = world[parent]
				}
				world[node] = pm.Mul(m.Nodes[node].M)
				cur = world[node]
				if stackID >= 0 && stackID < len(stack) {
					stack[stackID] = cur
				}
			}
		case 0x07, 0x08: // BB / BBY (billboards): same operands as NODEDESC
			p += 3
			if opt&1 != 0 {
				p++
			}
			if opt&2 != 0 {
				p++
			}
		case 0x09: // NODEMIX: skinning — skip (stackID, numMtx, then 3 bytes per mtx)
			if p+2 <= len(sbc) {
				n := int(sbc[p+1])
				p += 2 + 3*n
			} else {
				p = len(sbc)
			}
		case 0x0B: // POSSCALE: scale the current matrix by the model's up/down scale
			// (courses store vertices divided by a power-of-two "posScale" to fit the
			// fx16 vertex range; the SBC re-applies it before drawing). opt bit 0
			// selects the inverse (down) scale.
			s := m.UpScale
			if opt&1 != 0 && s != 0 {
				s = 1 / s
			}
			for i := 0; i < 9; i++ {
				cur[i] *= s
			}
		case 0x0C: // ENVMAP / 0x0D PRJMAP
			p += 2
		case 0x0D:
			p += 2
		default: // NOP and anything unmodelled: no operands
		}
	}
	return draws
}
