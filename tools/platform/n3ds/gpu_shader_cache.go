package n3ds

// gpu_shader_cache.go pre-decodes the vertex shader's instruction words.
//
// The interpreter used to re-derive everything about an instruction from its raw
// word every single time it executed it — and it executes each instruction once
// per vertex, for every vertex of every draw. Captain Toad's opening stage ran
// 26% of its frame inside `readSrc`/`arith`/`writeDst` doing that: extracting the
// opcode, masking out the operand-descriptor index, loading the descriptor,
// pulling the negate bit and the 8-bit swizzle out of it, and — per operand read —
// building a three-element slice literal just to pick a shift out of it.
//
// None of that depends on the vertex. A shader instruction's decode is a pure
// function of its word and the operand descriptor it names, both of which change
// only when the game uploads a new program. So decode each instruction once and
// keep it.
//
// The cache is keyed by an epoch rather than cleared: every write to the code FIFO,
// the operand-descriptor FIFO, *the geometry-shader code port — which writes into
// the same Code array* — and every savestate restore bumps the epoch, and a decoded
// entry is only valid if its epoch matches. Uploads happen a handful of times a
// frame against tens of thousands of vertex-instructions, so the invalidation is
// free and, more to the point, it cannot go stale by omission: a decode carrying
// last epoch's number is simply re-decoded.
//
// It is a cache of *decoding*, not of results: the arithmetic executed is the same
// arithmetic, in the same order, on the same values.

// The instruction classes the interpreter dispatches on.
const (
	shBad   = iota // not decoded yet, or an opcode arith will halt on
	shNop          // 0x21
	shEnd          // 0x22
	shFlow         // CALL/IF/LOOP/JMP — decoded inline; they are rare and cheap
	shArith        // the one- and two-source arithmetic ops
	shMad          // MAD/MADI (three sources)
	shCmp          // CMP
)

// shSrc is one source operand, fully resolved: which register file, which register
// in it, the relative-addressing mode (uniforms only), and the descriptor's negate
// flag and per-component swizzle selectors.
type shSrc struct {
	bank uint8    // shBankIn | shBankTmp | shBankUni
	reg  uint8    // index within that bank
	idx  uint8    // 0 = none, 1 = a0.x, 2 = a0.y, 3 = aL
	neg  bool     // negate every component
	sw   [4]uint8 // component i takes base[sw[i]]

	// plain marks the overwhelmingly common operand: the identity swizzle with no
	// negation, i.e. "the register itself". It turns the gather-and-negate into one
	// 16-byte copy.
	plain bool
}

const (
	shBankIn = iota
	shBankTmp
	shBankUni
)

// shInst is one decoded instruction.
type shInst struct {
	kind uint8
	op   uint8

	src [3]shSrc

	dstTmp  bool    // destination is a temporary (r), not an output (o)
	dst     uint8   // index within that file
	mask    [4]bool // the descriptor's component write mask (bit 3 = x)
	maskAll bool    // every component is written: store the whole vector at once

	cmpX, cmpY uint8 // CMP's two comparison operators
}

// shaderInst returns the decoded instruction at pc, decoding it if this is the
// first time it has been seen since the last upload.
//
// It MUTATES the cache, so it must not be called from a parallel vertex loop —
// decodeAll is what makes the cache read-only for the duration of a draw.
func (g *GPU) shaderInst(pc int) *shInst {
	if g.decEpoch[pc] != g.shEpoch {
		g.decode(pc)
		g.decEpoch[pc] = g.shEpoch
	}
	return &g.dec[pc]
}

// decodeAll decodes the whole code memory, so that the cache is pure data for the
// rest of this epoch and several goroutines may read it at once.
//
// Lazy decoding cannot be made safe for a parallel vertex loop by decoding "the
// instructions the program reaches" up front, because which instructions it reaches
// depends on the vertex: CMP and the conditional branches read per-vertex data. So
// the whole 4,096-word code memory is decoded, once per upload. A decode is a few
// nanoseconds of bit-twiddling and an upload happens a handful of times a frame,
// against tens of thousands of vertices.
func (g *GPU) decodeAll() {
	if g.decodedAll == g.shEpoch {
		return
	}
	for pc := range g.Code {
		if g.decEpoch[pc] != g.shEpoch {
			g.decode(pc)
			g.decEpoch[pc] = g.shEpoch
		}
	}
	g.decodedAll = g.shEpoch
}

// invalidateShaders drops every decoded instruction. Called from the code and
// operand-descriptor upload paths (gpu.go) and from a savestate restore (state.go),
// which replaces the whole Code/Opdesc memory underneath the cache.
func (g *GPU) invalidateShaders() {
	g.shEpoch++
	if g.shEpoch == 0 { // wrapped: every entry's stale epoch would now compare equal
		g.decEpoch = [len(g.Code)]uint32{}
		g.shEpoch = 1
	}
}

// decode fills the cache entry for pc. It mirrors, exactly, the field extraction
// the interpreter used to do inline — see arith()/readSrc()/writeDst().
func (g *GPU) decode(pc int) {
	in := g.Code[pc]
	op := in >> 26
	d := &g.dec[pc]
	*d = shInst{kind: shBad, op: uint8(op)}

	switch op {
	case 0x21:
		d.kind = shNop
		return
	case 0x22:
		d.kind = shEnd
		return
	case 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2C, 0x2D:
		d.kind = shFlow
		return
	}

	switch {
	case op >= 0x30: // MAD (wide src2) / MADI (wide src3); the descriptor field is 5 bits
		desc := g.Opdesc[in&0x1F]
		d.kind = shMad
		d.setDst(int(in>>24&0x1F), desc)
		d.src[0] = decodeSrc(int(in>>17&0x1F), 0, desc, 0)
		if op >= 0x38 {
			d.src[1] = decodeSrc(int(in>>10&0x7F), int(in>>22&3), desc, 1)
			d.src[2] = decodeSrc(int(in>>5&0x1F), 0, desc, 2)
		} else {
			d.src[1] = decodeSrc(int(in>>12&0x1F), 0, desc, 1)
			d.src[2] = decodeSrc(int(in>>5&0x7F), int(in>>22&3), desc, 2)
		}

	case op>>1 == 0x17: // CMP: its first comparison operator spills into the opcode's low bit
		desc := g.Opdesc[in&0x7F]
		d.kind = shCmp
		d.src[0] = decodeSrc(int(in>>12&0x7F), int(in>>19&3), desc, 0)
		d.src[1] = decodeSrc(int(in>>7&0x1F), 0, desc, 1)
		d.cmpX = uint8(in >> 24 & 7)
		d.cmpY = uint8(in >> 21 & 7)

	default:
		desc := g.Opdesc[in&0x7F]
		d.kind = shArith
		d.setDst(int(in>>21&0x1F), desc)
		idx := int(in >> 19 & 3)
		switch op {
		case 0x18, 0x19, 0x1A, 0x1B: // DPHI, DSTI, SGEI, SLTI: the wide field is src2
			d.src[0] = decodeSrc(int(in>>14&0x1F), 0, desc, 0)
			d.src[1] = decodeSrc(int(in>>7&0x7F), idx, desc, 1)
		default:
			d.src[0] = decodeSrc(int(in>>12&0x7F), idx, desc, 0)
			d.src[1] = decodeSrc(int(in>>7&0x1F), 0, desc, 1)
		}
	}
}

// setDst resolves the destination register file and the descriptor's write mask.
func (d *shInst) setDst(reg int, desc uint32) {
	if reg < 0x10 {
		d.dstTmp, d.dst = false, uint8(reg)
	} else {
		d.dstTmp, d.dst = true, uint8(reg-0x10)
	}
	d.maskAll = true
	for i := uint(0); i < 4; i++ {
		d.mask[i] = desc>>(3-i)&1 != 0
		d.maskAll = d.maskAll && d.mask[i]
	}
}

// shSrcShift is the descriptor's bit offset for operand slot n: the negate bit,
// then the 8-bit swizzle above it. This used to be built as a slice literal on
// every operand read, which is a heap-free but far from free way to index three
// constants.
var shSrcShift = [3]uint{4, 13, 22}

func decodeSrc(reg, idx int, desc uint32, n int) shSrc {
	var s shSrc
	switch {
	case reg < 0x10:
		s.bank, s.reg = shBankIn, uint8(reg)
	case reg < 0x20:
		s.bank, s.reg = shBankTmp, uint8(reg-0x10)
	default:
		s.bank, s.reg = shBankUni, uint8(reg-0x20)
		s.idx = uint8(idx) // relative addressing applies to uniform reads only
	}
	shift := shSrcShift[n]
	s.neg = desc>>shift&1 != 0
	sw := desc >> (shift + 1) & 0xFF
	for i := uint(0); i < 4; i++ {
		s.sw[i] = uint8(sw >> (6 - 2*i) & 3)
	}
	s.plain = !s.neg && s.sw == [4]uint8{0, 1, 2, 3}
	return s
}
