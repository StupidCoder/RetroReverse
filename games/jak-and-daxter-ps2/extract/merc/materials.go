package merc

// materials.go: the material side of the fragment format.
//
// A fragment whose fp region is larger than the origin quadword carries one
// or more 5-quadword adgif blocks — the GS register state for a texture:
// {TEX0, TEX1, MIPTBP1, CLAMP, ALPHA} as A+D pairs. On disc the TEX0/TEX1
// data words are templates (zero); at login `adgif-shader-login` (0x611F84)
// reads the texture id from the TEX1 quadword's third word, remaps it
// through the level's texture-remap table (`level-remap-texture` 0x68EAE4:
// sorted 8-byte {original-id, loaded-id} pairs, shipped in the level's vis
// object), and fills the registers from the texture-page descriptor. The
// id encodes (page << 20) | (index << 8).
//
// The first adgif block applies from the fragment's packet start; each
// further block is inserted mid-strip into a 6-quadword hole in the dest
// sequence (adgif tag + 5 A+D), so a dest gap of 6 beyond the vertex
// stride marks the switch point.

import "encoding/binary"

// ShaderRef is one adgif block: the disc texture id plus the register
// payloads that are real on disc.
type ShaderRef struct {
	RawID uint32 // texture id word (masked 0xFFFFFF00)
	Clamp uint64 // CLAMP_1 data
	Alpha uint64 // ALPHA_1 data
}

// Shaders parses the fragment's adgif blocks ((fpQWC-1)/5 of them).
func (fr *Fragment) Shaders() []ShaderRef {
	n := (fr.FPQWC - 1) / 5
	var out []ShaderRef
	for k := 0; k < n; k++ {
		base := (1 + k*5) * 16
		if base+5*16 > len(fr.FPData) {
			break
		}
		id := binary.LittleEndian.Uint32(fr.FPData[base+16+8:]) & 0xFFFFFF00
		cl := binary.LittleEndian.Uint64(fr.FPData[base+3*16:])
		al := binary.LittleEndian.Uint64(fr.FPData[base+4*16:])
		out = append(out, ShaderRef{RawID: id, Clamp: cl, Alpha: al})
	}
	return out
}

// RemapTable is the level's texture-id remap: sorted {orig, new} pairs.
type RemapTable [][2]uint32

// LoadRemapTable reads the table from a linked (base-0) level vis object:
// header word +0x34 is the table pointer, +0x38 the entry count (which may
// include a zero terminator — zero keys are dropped).
func LoadRemapTable(vis []byte) RemapTable {
	if len(vis) < 0x40 {
		return nil
	}
	ptr := binary.LittleEndian.Uint32(vis[0x34:])
	cnt := binary.LittleEndian.Uint32(vis[0x38:])
	var out RemapTable
	for i := uint32(0); i < cnt; i++ {
		o := ptr + i*8
		if int(o)+8 > len(vis) {
			break
		}
		k := binary.LittleEndian.Uint32(vis[o:])
		v := binary.LittleEndian.Uint32(vis[o+4:])
		if k != 0 {
			out = append(out, [2]uint32{k, v})
		}
	}
	return out
}

// Lookup remaps a texture id; ids not in the table pass through (the
// engine does the same when the level has no table).
func (t RemapTable) Lookup(id uint32) uint32 {
	key := id & 0xFFFFFF00
	lo, hi := 0, len(t)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		switch {
		case t[mid][0] == key:
			return t[mid][1]
		case t[mid][0] < key:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return key
}

// StreamVert is one packet-order vertex with its active material ordinal.
type StreamVert struct {
	Index int // effect-wide vertex index
	ADC   bool
	Mat   int // index into the effect's shader list
}

// EffectStream flattens the effect like EffectSequence but also tracks the
// active adgif block per vertex, and returns the effect's shader list.
func EffectStream(e *Effect) ([]StreamVert, []ShaderRef) {
	var seq []StreamVert
	var shaders []ShaderRef
	var prev map[byte]slotRef
	vbase := 0
	mat := -1
	for fi := range e.Fragments {
		fr := &e.Fragments[fi]
		fragShaders := fr.Shaders()
		nextShader := 0
		if len(fragShaders) > 0 {
			shaders = append(shaders, fragShaders[0])
			mat = len(shaders) - 1
			nextShader = 1
		}
		slots := walkFragment(fr, prev, vbase)
		dests := make([]int, 0, len(slots))
		for d := range slots {
			dests = append(dests, int(d))
		}
		sortInts(dests)
		last := -1
		for _, d := range dests {
			// a hole beyond the 3-qw stride that fits an adgif insert
			// (tag + 5 A+D = 6 quadwords) = a mid-strip material switch
			if last >= 0 && d-last-3 >= 6 && nextShader < len(fragShaders) {
				shaders = append(shaders, fragShaders[nextShader])
				mat = len(shaders) - 1
				nextShader++
			}
			last = d
			r := slots[byte(d)]
			seq = append(seq, StreamVert{Index: r.vert, ADC: r.adc, Mat: mat})
		}
		prev = slots
		vbase += fr.LumpQWC / 3
	}
	return seq, shaders
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}
