package gekko

// decode_ps.go decodes the instructions that are the Gekko's and not PowerPC's: the
// paired-single arithmetic in primary opcode 4, and the quantised loads and stores in
// opcodes 4, 56, 57, 60 and 61.
//
// Opcode 4 packs three differently-shaped extended-opcode fields into one primary
// opcode, and the decoder has to try them in the right order:
//
//	XO(26..30), five bits   the A-forms: three float operands (ps_madd, ps_sum0, ...)
//	XO(25..30), six bits    the indexed quantised loads (psq_lx, psq_stx, ...)
//	XO(21..30), ten bits    everything else (ps_merge, ps_neg, ps_cmpu0, dcbz_l)
//
// The five-bit values the A-forms use are {10-15, 18, 20, 21, 23-26, 28-31}. Every
// ten-bit form and every six-bit form in this opcode reduces, modulo 32, to a value
// outside that set, so testing narrowest-first is unambiguous — and decode_test.go
// asserts it over all 2^32 encodings' worth of field combinations rather than trusting
// the claim.
//
// The quantised load/store is the interesting instruction on this machine. psq_l reads
// one or two values from memory, in one of five formats — IEEE single, unsigned byte,
// unsigned halfword, signed byte, signed halfword — multiplies by 2^-scale, and lands
// two floats in a register. psq_st does the reverse, saturating on the way out. Neither
// the format nor the scale is in the instruction: both come from one of eight graphics
// quantisation registers, named by a three-bit field, and read at execution time. So the
// same psq_l instruction means different things at different moments, and a disassembly
// can only name the GQR, not the format. That is stated in the text it prints, rather
// than papered over.

// The A-form (five-bit XO) members of opcode 4.
var psAForm = map[uint32]string{
	10: "ps_sum0", 11: "ps_sum1", 12: "ps_muls0", 13: "ps_muls1",
	14: "ps_madds0", 15: "ps_madds1", 18: "ps_div", 20: "ps_sub", 21: "ps_add",
	23: "ps_sel", 24: "ps_res", 25: "ps_mul", 26: "ps_rsqrte",
	28: "ps_msub", 29: "ps_madd", 30: "ps_nmsub", 31: "ps_nmadd",
}

// psShape says which operands each A-form actually reads, since they do not agree:
// some are A×C, some A+B, some the full A×C±B.
const (
	shapeAB  = iota // frD, frA, frB
	shapeAC         // frD, frA, frC        (the multiplies)
	shapeACB        // frD, frA, frC, frB   (the multiply-accumulates and ps_sel)
	shapeB          // frD, frB             (the estimates)
)

var psShape = map[uint32]int{
	10: shapeACB, 11: shapeACB, // ps_sum0/1 take all three
	12: shapeAC, 13: shapeAC, // ps_muls0/1
	14: shapeACB, 15: shapeACB, // ps_madds0/1
	18: shapeAB, 20: shapeAB, 21: shapeAB, // ps_div, ps_sub, ps_add
	23: shapeACB,           // ps_sel
	24: shapeB, 26: shapeB, // ps_res, ps_rsqrte
	25: shapeAC,                                            // ps_mul
	28: shapeACB, 29: shapeACB, 30: shapeACB, 31: shapeACB, // the madd family
}

// The ten-bit members of opcode 4.
var psX = map[uint32]string{
	0: "ps_cmpu0", 32: "ps_cmpo0", 64: "ps_cmpu1", 96: "ps_cmpo1",
	40: "ps_neg", 72: "ps_mr", 136: "ps_nabs", 264: "ps_abs",
	528: "ps_merge00", 560: "ps_merge01", 592: "ps_merge10", 624: "ps_merge11",
	1014: "dcbz_l",
}

// decodePS decodes primary opcode 4.
func decodePS(in *Inst, w uint32) {
	d, a, b, c := rs(w), ra(w), rb(w), rc(w)

	// Narrowest field first: the A-forms.
	if name, ok := psAForm[xo5(w)]; ok {
		switch psShape[xo5(w)] {
		case shapeAB:
			in.set(dot(name, w), "%s,%s,%s", fr(d), fr(a), fr(b))
		case shapeAC:
			in.set(dot(name, w), "%s,%s,%s", fr(d), fr(a), fr(c))
		case shapeACB:
			in.set(dot(name, w), "%s,%s,%s,%s", fr(d), fr(a), fr(c), fr(b))
		case shapeB:
			in.set(dot(name, w), "%s,%s", fr(d), fr(b))
		}
		return
	}

	// Then the six-bit indexed quantised forms.
	switch xo6(w) {
	case 6:
		psqX(in, "psq_lx", w)
		return
	case 7:
		psqX(in, "psq_stx", w)
		return
	case 38:
		psqX(in, "psq_lux", w)
		return
	case 39:
		psqX(in, "psq_stux", w)
		return
	}

	// Then the ten-bit rest.
	name, ok := psX[xo10(w)]
	if !ok {
		return // unknown: DecodeWord renders it as .word and stops the path
	}
	switch name {
	case "ps_cmpu0", "ps_cmpo0", "ps_cmpu1", "ps_cmpo1":
		in.set(name, "cr%d,%s,%s", crfD(w), fr(a), fr(b))
	case "dcbz_l":
		// Not a floating-point instruction at all: it allocates a line in the locked
		// cache, zeroed, without reading main memory. It is how a program uses the
		// scratchpad, and it is here only because the opcode map put it here.
		base := r(a)
		if a == 0 {
			base = "0"
		}
		in.set(name, "%s,%s", base, r(b))
	case "ps_merge00", "ps_merge01", "ps_merge10", "ps_merge11":
		in.set(dot(name, w), "%s,%s,%s", fr(d), fr(a), fr(b))
	default: // ps_neg, ps_mr, ps_nabs, ps_abs
		in.set(dot(name, w), "%s,%s", fr(d), fr(b))
	}
}

// psqX renders an indexed quantised load or store: the GQR is named, the format and the
// scale it holds are not — because they are not in the instruction.
func psqX(in *Inst, mnem string, w uint32) {
	base := r(ra(w))
	if ra(w) == 0 {
		base = "0"
	}
	wBit := (w >> 15) & 1 // 1 = transfer one value rather than two
	gqr := (w >> 12) & 7
	in.set(mnem, "%s,%s,%s,%d,gqr%d", fr(rs(w)), base, r(rb(w)), wBit, gqr)
}

// decodePSQ decodes the non-indexed quantised loads and stores, primary opcodes 56, 57,
// 60 and 61. Their displacement is twelve bits, not sixteen: the top four bits of what
// would be the displacement field hold the W bit and the GQR index instead.
func decodePSQ(in *Inst, w uint32) {
	var mnem string
	switch opcd(w) {
	case 56:
		mnem = "psq_l"
	case 57:
		mnem = "psq_lu"
	case 60:
		mnem = "psq_st"
	case 61:
		mnem = "psq_stu"
	}
	d := psqDisp(w)
	wBit := (w >> 15) & 1
	gqr := (w >> 12) & 7

	base := r(ra(w))
	if ra(w) == 0 {
		base = "0"
	}
	in.set(mnem, "%s,%d(%s),%d,gqr%d", fr(rs(w)), d, base, wBit, gqr)
}

// psqDisp sign-extends the twelve-bit displacement.
func psqDisp(w uint32) int32 {
	return int32(w&0xFFF) << 20 >> 20
}
