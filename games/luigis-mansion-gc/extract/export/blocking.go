package export

// blocking.go — the demo player's per-frame actor placement ("A" in the
// actorsolve write-up: world = A(f) · composed(f)). The tables are captured
// from the running game by extract/cmd/actorsolve (-census + -bake) because
// no file on the disc stores them; SkinnedGLB bakes a table into the export
// as animation channels on the GLB's wrapper node, inside the same clip as
// the .key animation, so the shot scrubs as one unit.

import (
	"encoding/json"
	"fmt"
	"math"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

// Blocking is one actor's placement track in one shot.
type Blocking struct {
	Identity bool            `json:"identity,omitempty"`
	Frames   [][]float64     `json:"frames,omitempty"` // [f, m00..m23] row-major 3x4
	Attach   *BlockingAttach `json:"attach,omitempty"`
}

// BlockingTable is one shot's actor→track map, keyed "model.mdl+clip.key".
type BlockingTable struct {
	Shot   string               `json:"shot"`
	Note   string               `json:"note"`
	Actors map[string]*Blocking `json:"actors"`
}

// ParseBlocking reads a baked blocking table.
func ParseBlocking(b []byte) (*BlockingTable, error) {
	var t BlockingTable
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// trs is one decomposed placement sample.
type trs struct {
	t     [3]float32
	q     [4]float32
	sz    float32 // z scale: +1, or -1 when the placement mirrors (det < 0)
	frame float32
}

// decompose splits the 3x4 rows into translation + rotation + mirror. The
// captured matrices are rigid-or-mirrored (actorsolve screens |det| ≈ 1); a
// det < 0 sample is R·diag(1,1,-1), the demo's actor mirror.
func decompose(row []float64) (trs, error) {
	var m [12]float64
	copy(m[:], row[1:13])
	det := m[0]*(m[5]*m[10]-m[6]*m[9]) - m[1]*(m[4]*m[10]-m[6]*m[8]) + m[2]*(m[4]*m[9]-m[5]*m[8])
	out := trs{frame: float32(row[0]), sz: 1}
	if det < 0 {
		out.sz = -1
		// R = L·diag(1,1,-1): negate the third column.
		m[2], m[6], m[10] = -m[2], -m[6], -m[10]
	}
	// orthonormality check on R
	for r := 0; r < 3; r++ {
		l := m[r*4]*m[r*4] + m[r*4+1]*m[r*4+1] + m[r*4+2]*m[r*4+2]
		if math.Abs(l-1) > 0.05 {
			return out, fmt.Errorf("blocking frame %.0f: row %d length² %.3f, not rigid", row[0], r, l)
		}
	}
	out.t = [3]float32{float32(m[3]), float32(m[7]), float32(m[11])}
	out.q = matToQuat(m)
	return out, nil
}

// matToQuat converts a row-major 3x4's rotation block to an xyzw quaternion.
func matToQuat(m [12]float64) [4]float32 {
	tr := m[0] + m[5] + m[10]
	var x, y, z, w float64
	switch {
	case tr > 0:
		s := math.Sqrt(tr+1) * 2
		w = s / 4
		x = (m[9] - m[6]) / s
		y = (m[2] - m[8]) / s
		z = (m[4] - m[1]) / s
	case m[0] > m[5] && m[0] > m[10]:
		s := math.Sqrt(1+m[0]-m[5]-m[10]) * 2
		x = s / 4
		w = (m[9] - m[6]) / s
		y = (m[1] + m[4]) / s
		z = (m[2] + m[8]) / s
	case m[5] > m[10]:
		s := math.Sqrt(1+m[5]-m[0]-m[10]) * 2
		y = s / 4
		w = (m[2] - m[8]) / s
		x = (m[1] + m[4]) / s
		z = (m[6] + m[9]) / s
	default:
		s := math.Sqrt(1+m[10]-m[0]-m[5]) * 2
		z = s / 4
		w = (m[4] - m[1]) / s
		x = (m[2] + m[8]) / s
		y = (m[6] + m[9]) / s
	}
	return [4]float32{float32(x), float32(y), float32(z), float32(w)}
}

// bakeBlocking decomposes every frame, keeps quaternions on one hemisphere,
// and asserts the mirror sign never flips mid-shot.
func bakeBlocking(b *Blocking) ([]trs, error) {
	if b == nil || b.Identity || len(b.Frames) == 0 {
		return nil, nil
	}
	out := make([]trs, 0, len(b.Frames))
	for _, row := range b.Frames {
		if len(row) != 13 {
			return nil, fmt.Errorf("blocking row has %d values, want 13", len(row))
		}
		s, err := decompose(row)
		if err != nil {
			return nil, err
		}
		if n := len(out); n > 0 {
			if out[n-1].sz != s.sz {
				return nil, fmt.Errorf("blocking mirror flips at frame %.0f", s.frame)
			}
			if dot := out[n-1].q[0]*s.q[0] + out[n-1].q[1]*s.q[1] + out[n-1].q[2]*s.q[2] + out[n-1].q[3]*s.q[3]; dot < 0 {
				for i := range s.q {
					s.q[i] = -s.q[i]
				}
			}
		}
		out = append(out, s)
	}
	return out, nil
}

// BlockingAttach declares a prop riding a carrier joint: the prop's world
// node 0 equals worldCarrier(joint)·Offset (nil offset = identity). Solved by
// cmd/attachprobe from the game's own matrix arrays; the track itself is
// derived from the keys at export time — pure disc data.
type BlockingAttach struct {
	Carrier string    `json:"carrier"`
	Joint   int       `json:"joint"`
	Offset  []float64 `json:"offset,omitempty"`
}

// srt34World composes a key pose's world matrices through the node tree at
// frame f — the game's own Rz·Ry·Rx + T convention.
func srt34World(m *lm.MDL, key *lm.Key, f float32) []lm.Mtx34 {
	local := make([]lm.Mtx34, len(m.Nodes))
	for i := range m.Nodes {
		p := key.Eval(i, f)
		sx, cx := sincosf(p.Rot[0])
		sy, cy := sincosf(p.Rot[1])
		sz, cz := sincosf(p.Rot[2])
		local[i] = lm.Mtx34{
			cy * cz, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx, p.Translate[0],
			sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx, p.Translate[1],
			-sy, cy * sx, cy * cx, p.Translate[2],
		}
	}
	world := make([]lm.Mtx34, len(m.Nodes))
	var walk func(idx int, parent lm.Mtx34)
	walk = func(idx int, parent lm.Mtx34) {
		for idx >= 0 && idx < len(m.Nodes) {
			world[idx] = parent.Mul(local[idx])
			if m.Nodes[idx].Child >= 0 {
				walk(m.Nodes[idx].Child, world[idx])
			}
			idx = m.Nodes[idx].Sibling
		}
	}
	walk(0, lm.Identity34)
	return world
}

func sincosf(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func rowToMtx(row []float64) lm.Mtx34 {
	var m lm.Mtx34
	for i := 0; i < 12; i++ {
		m[i] = float32(row[i+1])
	}
	return m
}

// ExpandAttachments turns every attach-mode actor in the table into explicit
// per-frame wrapper rows: A_prop(f) = A_carrier(f)·composedCarrier(joint,f)·
// Offset·composedProp(0,f)⁻¹, so worldProp(0,f) lands exactly on the carrier
// joint. load resolves an actor key ("model.mdl+clip.key") to its parsed
// model and clip.
func (t *BlockingTable) ExpandAttachments(load func(spec string) (*lm.MDL, *lm.Key, error)) error {
	// carrierA returns the carrier's constant placement (attach chains hang
	// off constant-placed actors; the recursion below re-expands rides on
	// rides, e.g. cone → handlight → Luigi).
	var expand func(spec string, depth int) error
	expand = func(spec string, depth int) error {
		if depth > 4 {
			return fmt.Errorf("blocking: attachment chain too deep at %s", spec)
		}
		bl := t.Actors[spec]
		if bl == nil || bl.Attach == nil || len(bl.Frames) > 0 {
			return nil // absent, not attached, or already expanded
		}
		at := bl.Attach
		if err := expand(at.Carrier, depth+1); err != nil {
			return err
		}
		car := t.Actors[at.Carrier]
		if car == nil {
			return fmt.Errorf("blocking: %s rides unknown carrier %s", spec, at.Carrier)
		}
		carM, carKey, err := load(at.Carrier)
		if err != nil {
			return err
		}
		propM, propKey, err := load(spec)
		if err != nil {
			return err
		}
		_ = propM
		off := lm.Identity34
		if len(at.Offset) == 12 {
			for i := range off {
				off[i] = float32(at.Offset[i])
			}
		}
		carA := func(f int) lm.Mtx34 {
			switch {
			case car.Identity || len(car.Frames) == 0:
				return lm.Identity34
			case len(car.Frames) == 1:
				return rowToMtx(car.Frames[0])
			default:
				i := f
				if i >= len(car.Frames) {
					i = len(car.Frames) - 1
				}
				return rowToMtx(car.Frames[i])
			}
		}
		frames := int(propKey.Duration()) + 1
		if cf := int(carKey.Duration()) + 1; cf < frames {
			frames = cf
		}
		for f := 0; f < frames; f++ {
			carWorld := srt34World(carM, carKey, float32(f))
			if at.Joint < 0 || at.Joint >= len(carWorld) {
				return fmt.Errorf("blocking: %s joint %d out of range", spec, at.Joint)
			}
			propRoot := srt34World(propM, propKey, float32(f))[0]
			A := carA(f).Mul(carWorld[at.Joint]).Mul(off).Mul(invert34(propRoot))
			row := make([]float64, 13)
			row[0] = float64(f)
			for i := 0; i < 12; i++ {
				row[i+1] = float64(A[i])
			}
			bl.Frames = append(bl.Frames, row)
		}
		bl.Attach = nil
		return nil
	}
	for spec := range t.Actors {
		if err := expand(spec, 0); err != nil {
			return err
		}
	}
	return nil
}
