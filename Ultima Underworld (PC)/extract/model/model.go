// Package model decodes UW's 3D object models ("chairs, tables, doors") from
// UW.EXE. The models are BYTECODE PROGRAMS for the game's display-list VM — the
// same interpreter (07F7, jump table 499D:2738) that renders the level. Every
// opcode's semantics here were derived from that interpreter's code (see
// disasm/uw-render.annotations.txt §10-§10c); this decoder simulates the
// vertex-building opcodes in model space to reconstruct the mesh.
//
// Location (verified against live memory with the oracle): the 64-entry offset
// table sits in DGROUP right after the VM jump table, at 499D:28A0
// (file 0x4E370), and model N's record is at 499D:292E + table[N]:
//
//	+0  word   (consumed by the object-emission pass)
//	+2  word 0
//	+4  extents X,Y,Z (3 words) — read by the draw pass as [SI-6/-4/-2]
//	+A  the program; SI starts here (first opcode is 0x7A, plant a vertex)
//
// Opcodes used by models (word value = byte offset into the 2738 table; the
// models use the high alias range 0x78+ of the same handlers):
//
//	0x00                RET — end of (sub)program
//	0x7A  x,y,z,dst     transform one word-coord vertex into pool slot dst (50BE)
//	0x78  pool,hx,hy,hz,pad box frustum test: 8 corners = pool ± half-extents run
//	                    through the matrix; all-inside → no-clip fast mode
//	                    (5040(FFFF)); both exits skip the pad word. No vertices.
//	0x82  n,off,n*(xyz) batch: n word-coord vertices into pool from byte off
//	0x86  src,off,dst   pool[dst] = pool[src] + off along model X (matrix row 1)
//	0x88  src,off,dst   ... along model Y (row 2)
//	0x8A  src,off,dst   ... along model Z (row 3)
//	0x8C  a,b,dst?      pool add; 0xC6 pool subtract (rare)
//	0x6A  link,anchor   face-begin (packed): plane test d+ΔX<0 vs the anchor →
//	                    CALL the draw sub-program at the link target (metadata
//	                    d,shift,bounds at target-0xA); continue after the anchor
//	0x38  ...           face-begin (inline bounds variant)
//	0x34/0x36 desc,4*(v,uv) textured QUAD from 4 pool slots (0x36) / flat (0x34)
//	0x40/0x44/0xD6/0xD8 select projection/draw back-end (no geometry)
//	0x02/0x04/0x0A      pokes (state)
//	0x32  addr          jump indirect (computed goto)  0x48 rel jump
//	0x14/0x16 n,a,imm   conditional skip
package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

const (
	magicFileOff = 0 // found by scanning for the magic
	numModels    = 64
)

var magic = []byte{0xB6, 0x4A, 0x06, 0x40}

// Vertex is a model-space vertex (word coordinates as the stream stores them).
type Vertex struct{ X, Y, Z int16 }

// Quad is one textured quad: four pool slots and the descriptor/UV codes.
type Quad struct {
	Desc     uint16    // polygon-descriptor index (texture binding by the emitter)
	V        [4]uint16 // pool slot numbers (byte offset / 8)
	UV       [4]uint16 // per-vertex UV code words
	Textured bool      // opcode 0x36 (vs 0x34 flat)
}

// Poly is an N-gon: gathered pool slots and, for textured n-gons (ops
// 0xA8/0xB4), the per-vertex texture-coordinate codes. Color's top nibble
// (0xF000) marks a textured poly; otherwise it is a flat-shade colour set by op
// 0xBC through the light-shade table. UV, when present, is parallel to V: the
// raw U,V code words the draw handler (07F7:5C5E) scales into the texture
// (U = code*texW>>8, V = code*texH>>16), so a normalised UV is code/65536.
type Poly struct {
	Color uint16
	V     []uint16
	UV    [][2]uint16 // per-vertex (U,V) code words; nil for flat polys
}

// UVScale is the divisor that normalises a Poly.UV code word to 0..1.
const UVScale = 65536.0

// Face is a plane-checked group: the anchor (plane point) and the quads its
// draw sub-program emits.
type Face struct {
	Anchor Vertex
	D      int16 // plane scalar (visible when d + ΔX >= 0)
	Quads  []int // indices into Model.Quads
}

// Model is one decoded model.
type Model struct {
	Index    int
	Offset   uint16 // table offset
	Extents  [3]int16
	Shift    [3]int16          // anchor shift (op 0x4A): add to every vertex when placing
	Pool     map[uint16]Vertex // pool slot -> model-space vertex
	Quads    []Quad
	Polys    []Poly
	Faces    []Face
	Ops      []string // human-readable decoded listing
	Warnings []string
	// EnvSlots are pool slots the program READS but never writes: the
	// object-emission pass pre-loads them with environment vectors (e.g. the
	// tile's floor-to-ceiling vector) before running the program. This is how a
	// doorframe or pillar adapts to the room: its top corners are derived as
	// bottom + envSlot (op 0x8C). Extent 1024 on an axis marks such models.
	EnvSlots []uint16
}

type decoder struct {
	data    []byte // UW.EXE
	m       *Model
	seen    map[int]bool
	color   uint16   // current colour (op 0xBC)
	gather  []uint16 // pending N-gon pool slots (op 0x22)
	written map[uint16]bool
	swing   float64 // door-swing angle (radians) applied to 0xBA rotation scopes
}

// wr marks a pool slot written; rd flags an environment slot (read before any
// write — pre-loaded by the emitter, e.g. the floor-to-ceiling vector).
func (d *decoder) wr(slot uint16) { d.written[slot] = true }
func (d *decoder) rd(slot uint16) {
	if d.written[slot] {
		return
	}
	for _, s := range d.m.EnvSlots {
		if s == slot {
			return
		}
	}
	d.m.EnvSlots = append(d.m.EnvSlots, slot)
}

// Decode parses UW.EXE and returns all 64 models.
func Decode(exe []byte) ([]*Model, error) { return DecodeWithEnv(exe, nil) }

// DecodeWithEnv decodes the models with environment slots pre-loaded, the way
// the object-emission pass runs them: env maps a pool slot (e.g. 256, 128) to
// the vector the emitter would plant there — for ceiling-adaptive models the
// tile's floor-to-ceiling vector. With env nil those slots read as zero.
func DecodeWithEnv(exe []byte, env map[uint16]Vertex) ([]*Model, error) {
	return DecodeWithEnvSwing(exe, env, 0)
}

// DecodeWithEnvSwing additionally swings the door leaf: swingRad (radians) is
// applied to every 0xBA scoped-rotation sub-program (the game rotates its
// vertices about the vertical Y axis by the angle at [292A], 0 = closed). Pass
// π/2 to render doors fully open.
func DecodeWithEnvSwing(exe []byte, env map[uint16]Vertex, swingRad float64) ([]*Model, error) {
	i := bytes.Index(exe, magic)
	if i < 0 {
		return nil, fmt.Errorf("model: magic not found")
	}
	tbl := i
	base := i + 4 + 138 // == 499D:292E when loaded; verified in live memory
	var models []*Model
	for n := 0; n < numModels; n++ {
		off := binary.LittleEndian.Uint16(exe[tbl+2*n:])
		m := &Model{Index: n, Offset: off, Pool: map[uint16]Vertex{}}
		s := base + int(off)
		if off == 0 || s+10 >= len(exe) {
			m.Warnings = append(m.Warnings, "empty/stub")
			models = append(models, m)
			continue
		}
		for k := 0; k < 3; k++ {
			m.Extents[k] = int16(binary.LittleEndian.Uint16(exe[s+4+2*k:]))
		}
		for slot, v := range env {
			m.Pool[slot] = v
		}
		d := &decoder{data: exe, m: m, seen: map[int]bool{}, written: map[uint16]bool{}, swing: swingRad}
		d.run(s+10, 0)
		models = append(models, m)
	}
	return models, nil
}

func (d *decoder) w(p int) uint16         { return binary.LittleEndian.Uint16(d.data[p:]) }
func (d *decoder) sw(p int) int16         { return int16(d.w(p)) }
func (d *decoder) log(f string, a ...any) { d.m.Ops = append(d.m.Ops, fmt.Sprintf(f, a...)) }

// run interprets the vertex-building/drawing stream at p (file offset) until a
// RET. depth guards face-subprogram recursion.
func (d *decoder) run(p int, depth int) {
	if depth > 64 {
		d.m.Warnings = append(d.m.Warnings, "recursion cap")
		return
	}
	for steps := 0; steps < 4096; steps++ {
		if d.seen[p] {
			d.log("%04X: (already decoded, stop)", p)
			return
		}
		d.seen[p] = true
		op := d.w(p)
		switch op {
		case 0x00:
			d.log("%04X: 00 RET", p)
			return
		case 0x7A: // vertex x,y,z -> pool[dst]
			x, y, z, dst := d.sw(p+2), d.sw(p+4), d.sw(p+6), d.w(p+8)
			d.m.Pool[dst/8] = Vertex{x, y, z}
			d.wr(dst / 8)
			d.log("%04X: 7A vertex (%d,%d,%d) -> v%d", p, x, y, z, dst/8)
			p += 10
		case 0x78: // box classify: pool, hx, hy, hz, pad (both handler exits ADD SI,2)
			pv, hx, hy, hz := d.w(p+2), d.sw(p+4), d.sw(p+6), d.sw(p+8)
			d.log("%04X: 78 boxtest v%d half(%d,%d,%d)", p, pv/8, hx, hy, hz)
			p += 12
		case 0x82: // batch: n, off, n*(x,y,z)
			n, off := int(d.w(p+2)), d.w(p+4)
			p += 6
			for k := 0; k < n; k++ {
				x, y, z := d.sw(p), d.sw(p+2), d.sw(p+4)
				d.m.Pool[off/8+uint16(k)] = Vertex{x, y, z}
				d.wr(off/8 + uint16(k))
				d.log("%04X: 82 vertex (%d,%d,%d) -> v%d", p, x, y, z, off/8+uint16(k))
				p += 6
			}
		case 0x86, 0x88, 0x8A: // src, off, dst — offset along model axis X/Y/Z
			src, off, dst := d.w(p+2), d.sw(p+4), d.w(p+6)
			d.rd(src / 8)
			d.wr(dst / 8)
			ax := map[uint16]int{0x86: 0, 0x88: 1, 0x8A: 2}[op]
			v := d.m.Pool[src/8]
			switch ax {
			case 0:
				v.X += off
			case 1:
				v.Y += off
			case 2:
				v.Z += off
			}
			d.m.Pool[dst/8] = v
			d.log("%04X: %02X v%d = v%d %+d on axis %d", p, op, dst/8, src/8, off, ax)
			p += 8
		case 0x8C, 0xC6: // pool add/sub: a, b, dst (3BB1/3BEE). With b = an
			// emitter-pre-loaded environment slot (v128/v256) this is
			// "raise vertex to the ceiling" — the doorframe/pillar adaptation.
			a, b, dst := d.w(p+2), d.w(p+4), d.w(p+6)
			d.rd(a / 8)
			d.rd(b / 8)
			d.wr(dst / 8)
			va, vb := d.m.Pool[a/8], d.m.Pool[b/8]
			if op == 0x8C {
				d.m.Pool[dst/8] = Vertex{va.X + vb.X, va.Y + vb.Y, va.Z + vb.Z}
			} else {
				d.m.Pool[dst/8] = Vertex{va.X - vb.X, va.Y - vb.Y, va.Z - vb.Z}
			}
			d.log("%04X: %02X v%d = v%d %s v%d", p, op, dst/8, a/8,
				map[uint16]string{0x8C: "+", 0xC6: "-"}[op], b/8)
			p += 8
		case 0x90, 0x92, 0x94: // two-axis displacement: off1, off2, src, dst
			o1, o2, src, dst := d.sw(p+2), d.sw(p+4), d.w(p+6), d.w(p+8)
			d.rd(src / 8)
			d.wr(dst / 8)
			v := d.m.Pool[src/8]
			switch op {
			case 0x90: // X, Y axes (matrix rows 1602, 1608)
				v.X += o1
				v.Y += o2
			case 0x92: // X, Z axes (1602, 160E)
				v.X += o1
				v.Z += o2
			case 0x94: // Z, Y axes (160E, 1608)
				v.Z += o1
				v.Y += o2
			}
			d.m.Pool[dst/8] = v
			d.log("%04X: %02X v%d = v%d %+d,%+d (2-axis)", p, op, dst/8, src/8, o1, o2)
			p += 10
		case 0x18: // inline int32 vertex -> current regs (no pool)
			d.log("%04X: 18 current-vertex (int32)", p)
			p += 14
		case 0x20: // int32 vertex + pool index
			x := int16(d.w(p + 2)) // low words carry the model-space value
			y := int16(d.w(p + 6))
			z := int16(d.w(p + 10))
			dst := d.w(p + 14)
			d.m.Pool[dst/8] = Vertex{x, y, z}
			d.log("%04X: 20 vertex32 (%d,%d,%d) -> v%d", p, x, y, z, dst/8)
			p += 16
		case 0x4A: // adjust current deltas by 3 words: shifts the model's anchor
			// relative to the object position. The door leaf uses (-48,0,0),
			// which puts its 0..128 span exactly onto the doorframe's opening
			// (frame model 1 builds the opening at X -48..+80).
			d.m.Shift[0] += d.sw(p + 2)
			d.m.Shift[1] += d.sw(p + 4)
			d.m.Shift[2] += d.sw(p + 6)
			d.log("%04X: 4A delta -= (%d,%d,%d)", p, d.sw(p+2), d.sw(p+4), d.sw(p+6))
			p += 8
		case 0x5E, 0x60, 0x62: // two-term plane-side conditional skip (3F00/3F29/3F52)
			skip := int(d.w(p + 2))
			d.log("%04X: %02X if plane2(%d,%d,%d,%d) skip %d", p, op,
				d.sw(p+4), d.sw(p+6), d.sw(p+8), d.sw(p+10), skip)
			d.run(p+12+skip, depth+1)
			p += 12
		case 0x64, 0x66, 0x68: // one-axis plane-side conditional skip (3F7B/3F92/3FA9)
			skip, factor, off := int(d.w(p+2)), d.sw(p+4), d.sw(p+6)
			d.log("%04X: %02X if plane(f=%d,off=%d) skip %d", p, op, factor, off, skip)
			d.run(p+8+skip, depth+1) // decode the skip target too
			p += 8
		case 0x22, 0x7E: // gather N pool verts AND draw the flat N-gon (2E73 falls
			// through into projection; op 0x80 is only the standalone re-draw)
			n := int(d.w(p + 2))
			d.gather = d.gather[:0]
			pp := p + 4
			for k := 0; k < n; k++ {
				d.gather = append(d.gather, d.w(pp)/8)
				pp += 2
			}
			d.m.Polys = append(d.m.Polys, Poly{Color: d.color, V: append([]uint16{}, d.gather...)})
			d.log("%04X: %02X ngon color=%03X %v", p, op, d.color, d.gather)
			p = pp
		case 0x80: // draw the gathered flat N-gon with the current colour
			pl := Poly{Color: d.color, V: append([]uint16{}, d.gather...)}
			d.m.Polys = append(d.m.Polys, pl)
			d.log("%04X: 80 draw ngon color=%03X %v", p, d.color, pl.V)
			p += 2
		case 0xBC: // set colour via the light-shade table: addr, colour
			d.color = d.w(p + 4)
			d.log("%04X: BC color [%04X] base=%03X", p, d.w(p+2), d.color)
			p += 6
		case 0xA4, 0xA6: // textured QUAD, compact: desc word + 4 pool-index BYTES,
			// implicit full-texture corner UVs (5D6F/5D77 -> 5D7D)
			q := Quad{Desc: d.w(p + 2), Textured: op == 0xA4}
			for k := 0; k < 4; k++ {
				q.V[k] = uint16(d.data[p+4+k])
			}
			d.m.Quads = append(d.m.Quads, q)
			d.log("%04X: %02X quadB desc=%d v(%d,%d,%d,%d)", p, op, q.Desc, q.V[0], q.V[1], q.V[2], q.V[3])
			p += 8
		case 0xA8, 0xAA: // textured N-gon: desc, count, count*(pool,U,V) (5C7E/5C86)
			desc, n := d.w(p+2), int(d.w(p+4))
			pl := Poly{Color: 0xF000 | desc} // mark textured
			pp := p + 6
			for k := 0; k < n; k++ {
				pl.V = append(pl.V, d.w(pp)/8)
				pl.UV = append(pl.UV, [2]uint16{d.w(pp + 2), d.w(pp + 4)})
				pp += 6
			}
			d.m.Polys = append(d.m.Polys, pl)
			d.log("%04X: %02X tex-ngon desc=%d n=%d %v", p, op, desc, n, pl.V)
			p = pp
		case 0xB4, 0xCE: // textured N-gon, current desc: count, count*(pool,U,V)
			n := int(d.w(p + 2))
			pl := Poly{Color: 0xF000}
			pp := p + 4
			for k := 0; k < n; k++ {
				pl.V = append(pl.V, d.w(pp)/8)
				pl.UV = append(pl.UV, [2]uint16{d.w(pp + 2), d.w(pp + 4)})
				pp += 6
			}
			d.m.Polys = append(d.m.Polys, pl)
			d.log("%04X: %02X tex-ngon (cur desc) n=%d %v", p, op, n, pl.V)
			p = pp
		case 0xA0, 0xA2: // textured QUAD, byte indices, implicit corner UVs,
			// reversed winding vs 0xA4/0xA6 (5CED/5CF5 -> 5CFB, CX=4 fixed)
			q := Quad{Desc: d.w(p + 2), Textured: true}
			for k := 0; k < 4; k++ {
				q.V[k] = uint16(d.data[p+4+k])
			}
			d.m.Quads = append(d.m.Quads, q)
			d.log("%04X: %02X quadB' desc=%d v(%d,%d,%d,%d)", p, op, q.Desc, q.V[0], q.V[1], q.V[2], q.V[3])
			p += 8
		case 0xB2: // set polygon descriptor: desc word (5C30)
			d.log("%04X: B2 set-desc %d", p, d.w(p+2))
			p += 4
		case 0xD0: // set span-mode flag: word (5C45)
			d.log("%04X: D0 span-mode %04X", p, d.w(p+2))
			p += 4
		case 0xAC: // no-op (5C78 dispatches immediately)
			d.log("%04X: AC nop", p)
			p += 2
		case 0xD4: // per-vertex shade: count, addr, count*(pool word + shade byte),
			// then INC SI; AND SI,~1 (round up to even)
			n := int(d.w(p + 2))
			d.log("%04X: D4 vertex-shades n=%d src=[%04X]", p, n, d.w(p+4))
			p = (p + 6 + 3*n + 1) &^ 1
		case 0x50, 0x70, 0x72: // scoped rotation, direct angle: angle, rel -> sub
			ang, rel := d.sw(p+2), int(int16(d.w(p+4)))
			d.log("%04X: %02X rotate(angle=%d) sub@%04X", p, op, ang, p+6+rel)
			d.run(p+6+rel, depth+1) // the rotated sub-program (leaf geometry)
			p += 6
		case 0xBA, 0xC2, 0xC4: // scoped rotation, angle from memory: addr, rel -> sub
			addr, rel := d.w(p+2), int(int16(d.w(p+4)))
			d.log("%04X: %02X rotate(angle=[%04X]) sub@%04X", p, op, addr, p+6+rel)
			// The door leaf ([292A] via 0xBA) swings about the vertical Y axis
			// (routine 5426 mixes matrix X/Z). Rotate the sub-program's own
			// vertices by the swing angle so a door can be rendered open. Only
			// 0xBA (the door) is swung; 0xC2/0xC4 are other axes left as-is.
			before := map[uint16]bool{}
			if op == 0xBA && d.swing != 0 {
				for k := range d.m.Pool {
					before[k] = true
				}
			}
			d.run(p+6+rel, depth+1)
			if op == 0xBA && d.swing != 0 {
				s, c := math.Sin(d.swing), math.Cos(d.swing)
				for k, v := range d.m.Pool {
					if before[k] {
						continue
					}
					x, z := float64(v.X), float64(v.Z)
					d.m.Pool[k] = Vertex{X: int16(x*c + z*s), Y: v.Y, Z: int16(-x*s + z*c)}
				}
			}
			p += 6
		case 0xBE: // copy word: src, ptr
			d.log("%04X: BE *[%04X]=*[%04X]", p, d.w(p+4), d.w(p+2))
			p += 6
		case 0x06: // two-sided face: 3-term plane (nx,ox,ny,oy,nz,oz) then two
			// sub-calls; back side runs them in reverse painter's order (30AE/3090)
			relA, relB := int(int16(d.w(p+14))), int(int16(d.w(p+16)))
			d.log("%04X: 06 plane3 n(%d,%d,%d) o(%d,%d,%d) subs @%04X @%04X", p,
				d.sw(p+2), d.sw(p+6), d.sw(p+10), d.sw(p+4), d.sw(p+8), d.sw(p+12),
				p+16+relA, p+18+relB)
			d.run(p+16+relA, depth+1)
			d.run(p+18+relB, depth+1)
			p += 18
		case 0x0C, 0x0E, 0x10: // two-sided face, 2-term plane variants
			relA, relB := int(int16(d.w(p+10))), int(int16(d.w(p+12)))
			d.log("%04X: %02X plane2 subs @%04X @%04X", p, op, p+12+relA, p+14+relB)
			d.run(p+12+relA, depth+1)
			d.run(p+14+relB, depth+1)
			p += 14
		case 0x58: // three-term plane-side conditional skip (3FC0)
			skip := int(d.w(p + 2))
			d.log("%04X: 58 if plane3 skip %d", p, skip)
			d.run(p+16+skip, depth+1)
			p += 16
		case 0x6A: // face-begin packed
			link := int(d.w(p + 2))
			ax, ay, az := d.sw(p+4), d.sw(p+8), d.sw(p+12)
			tail := p + 4 + link // DI = SI+link where SI = after link word
			f := Face{Anchor: Vertex{ax, ay, az}}
			if tail-10 > 0 && tail < len(d.data) {
				f.D = d.sw(tail - 10)
			}
			d.log("%04X: 6A face anchor(%d,%d,%d) d=%d sub@%04X", p, ax, ay, az, f.D, tail)
			nq := len(d.m.Quads)
			d.run(tail, depth+1) // the draw sub-program (CALLed until RET)
			for i := nq; i < len(d.m.Quads); i++ {
				f.Quads = append(f.Quads, i)
			}
			d.m.Faces = append(d.m.Faces, f)
			p += 16 // continue after the 12-byte anchor
		case 0x38: // face-begin inline: link,shift,bx,X:2,by,Y:2,bz,Z:2,d
			link := int(d.w(p + 2))
			tail := p + 4 + link
			ax := int16(d.w(p + 8))
			ay := int16(d.w(p + 14))
			az := int16(d.w(p + 20))
			f := Face{Anchor: Vertex{ax, ay, az}, D: d.sw(p + 24)}
			d.log("%04X: 38 face anchor(%d,%d,%d) d=%d sub@%04X", p, ax, ay, az, f.D, tail)
			nq := len(d.m.Quads)
			d.run(tail, depth+1)
			for i := nq; i < len(d.m.Quads); i++ {
				f.Quads = append(f.Quads, i)
			}
			d.m.Faces = append(d.m.Faces, f)
			p += 26
		case 0x34, 0x36: // quad: desc, 4*(pool, uv)
			q := Quad{Desc: d.w(p + 2), Textured: op == 0x36}
			pp := p + 4
			for k := 0; k < 4; k++ {
				q.V[k] = d.w(pp) / 8
				q.UV[k] = d.w(pp + 2)
				pp += 4
			}
			d.m.Quads = append(d.m.Quads, q)
			d.log("%04X: %02X quad desc=%d v(%d,%d,%d,%d)", p, op, q.Desc, q.V[0], q.V[1], q.V[2], q.V[3])
			p = pp
		case 0x40, 0x44, 0xD6, 0xD8: // back-end selects, no operands
			d.log("%04X: %02X select-backend", p, op)
			p += 2
		case 0x02, 0x04: // poke: addr, value
			d.log("%04X: %02X poke [%04X]=%04X", p, op, d.w(p+2), d.w(p+4))
			p += 6
		case 0x0A:
			d.log("%04X: 0A poke [BP+%04X]=%04X", p, d.w(p+2), d.w(p+4))
			p += 6
		case 0x32: // jump indirect
			d.log("%04X: 32 jmp [%04X] (stop: runtime target)", p, d.w(p+2))
			return
		case 0x48: // jump relative
			rel := int(int16(d.w(p + 2)))
			d.log("%04X: 48 jmp %+d", p, rel)
			p = p + 4 + rel
		case 0x14, 0x16: // conditional skip: N, addr, imm
			n := int(d.w(p + 2))
			d.log("%04X: %02X if [%04X] %s %d skip %d", p, op, d.w(p+4),
				map[uint16]string{0x14: ">", 0x16: "<"}[op], int16(d.w(p+6)), n)
			p += 8
			_ = n // both paths continue linearly here; the skip target is p+n (decoded on the fall-through path)
		case 0x12: // call relative
			rel := int(int16(d.w(p + 2)))
			d.log("%04X: 12 call %+d", p, rel)
			d.run(p+4+rel, depth+1)
			p += 4
		default:
			d.m.Warnings = append(d.m.Warnings,
				fmt.Sprintf("unhandled op %04X at %04X — stop", op, p))
			d.log("%04X: ?? %04X (stop)", p, op)
			return
		}
	}
	d.m.Warnings = append(d.m.Warnings, "step cap")
}
