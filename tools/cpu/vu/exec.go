package vu

// exec.go is the VU interpreter. A VU instruction is a pair issued together: the upper
// half on the floating-point FMACs, the lower half on the integer/branch/load-store
// pipe. This interpreter keeps the architectural contract and drops the timing: results
// are visible immediately, Q and P are ready the moment they are asked for (WAITQ and
// WAITP are no-ops), and the pipelines never stall. What it keeps of the parallel issue
// is the one visible rule: the two halves of a pair read the machine as it was BEFORE
// the pair — the upper result is computed first but committed after the lower half has
// read its operands.
//
// Vector registers are held as raw bits, not floats: FTOI writes integer bit patterns
// through the same registers, and a float32 round-trip would corrupt them. Arithmetic
// converts at the edges and sanitises its results — the VU's floats have no NaN or
// infinity (the exponent saturates), so a Go NaN/Inf is clamped to the largest value of
// the right sign before it lands in a register.

import "math"

// VU is one vector unit.
type VU struct {
	VF  [32][4]uint32 // raw bits; VF00 is pinned to (0, 0, 0, 1.0)
	VI  [16]uint16    // VI00 is pinned to 0
	ACC [4]float32
	Q   float32
	P   float32
	I   float32
	R   uint32

	Micro []byte
	Data  []byte

	PC uint32 // byte address into Micro

	Mac    uint16 // per-lane zero/sign of the last FMAC result
	Status uint16
	Clip   uint32 // 24 bits, shifted 6 per CLIP

	// The FMAC pipeline is four stages deep, and the flag registers a LOWER instruction
	// reads (FMAND, FSAND, FCAND...) are the flags of the upper instruction four pairs
	// BACK — the microcode is scheduled around that latency, and a kernel's cull branch
	// tests a specific instruction's flags by position. flagPipe carries the in-flight
	// values; vis* are what this pair's lower half sees.
	flagPipe                    [4]flagVals
	visMac, visStatus           uint16
	visClip                     uint32

	// Top and ITop are the VIF's double-buffer pointers, latched at MSCAL and read by
	// XTOP/XITOP.
	Top, ITop uint16

	// XGKick, if set, is handed the data-memory quadword address of a GIF packet the
	// program has finished building.
	XGKick func(qwAddr uint32)

	// CMSAR0 is VCALLMSR's start address (64-bit pairs); StartVU1, if set, serves a
	// macro-mode write of CMSAR1 — the EE starting the OTHER unit. Both belong to the
	// macro face (macro.go).
	CMSAR0   uint16
	StartVU1 func(startPair uint32)

	Steps uint64
}

// flagVals is one pipeline stage's worth of flag state.
type flagVals struct {
	mac    uint16
	status uint16
	clip   uint32
}

// New makes a VU over the two memories the VIF fills.
func New(micro, data []byte) *VU {
	v := &VU{Micro: micro, Data: data}
	v.VF[0][3] = math.Float32bits(1.0) // VF00 is the constant (0, 0, 0, 1)
	return v
}

// Run executes from a byte address in program memory until the E bit ends the program
// or the step budget runs out. It reports the steps taken and whether the program ended
// by its own E bit.
func (v *VU) Run(start uint32, maxSteps int) (int, bool) {
	if len(v.Micro) == 0 {
		return 0, false
	}
	mask := uint32(len(v.Micro)-1) &^ 7
	v.PC = start & mask

	delayed := int64(-1) // a taken branch's target, applied after its delay slot
	endIn := -1          // instructions left after the E bit
	var res upperResult
	for steps := 0; steps < maxSteps; steps++ {
		raw := le64m(v.Micro[v.PC:])
		up := uint32(raw >> 32)
		lo := uint32(raw)

		// What this pair's lower half sees of the flags is the pipeline's oldest stage.
		v.visMac = v.flagPipe[0].mac
		v.visStatus = v.flagPipe[0].status
		v.visClip = v.flagPipe[0].clip

		v.execUpper(up, &res)
		var taken int64 = -1
		if up&(1<<31) != 0 { // I: the lower word is the I register's new value
			v.I = sane(math.Float32frombits(lo))
		} else {
			taken = v.execLower(lo)
		}
		v.commitUpper(&res)
		v.flagPipe[0], v.flagPipe[1], v.flagPipe[2] = v.flagPipe[1], v.flagPipe[2], v.flagPipe[3]
		v.flagPipe[3] = flagVals{mac: v.Mac, status: v.Status, clip: v.Clip}
		v.Steps++

		next := (v.PC + 8) & mask
		if delayed >= 0 {
			next = uint32(delayed) & mask
			delayed = -1
		}
		if taken >= 0 {
			delayed = taken
		}
		v.PC = next

		if endIn >= 0 {
			endIn--
			if endIn < 0 {
				return steps + 1, true
			}
		} else if up&(1<<30) != 0 { // E: one more pair, then stop
			endIn = 0
		}
	}
	return maxSteps, false
}

// --- register access ------------------------------------------------------------

func (v *VU) getF(r, lane uint32) float32 {
	return math.Float32frombits(v.VF[r][lane])
}

// setLanes writes a float result through the dest mask, sanitised.
func (v *VU) setLanes(r uint32, destMask uint32, val [4]float32) {
	if r == 0 {
		return
	}
	for lane := uint32(0); lane < 4; lane++ {
		if destMask&(8>>lane) != 0 {
			v.VF[r][lane] = math.Float32bits(sane(val[lane]))
		}
	}
}

func (v *VU) setBits(r uint32, destMask uint32, val [4]uint32) {
	if r == 0 {
		return
	}
	for lane := uint32(0); lane < 4; lane++ {
		if destMask&(8>>lane) != 0 {
			v.VF[r][lane] = val[lane]
		}
	}
}

func (v *VU) setVI(r uint32, val uint16) {
	if r != 0 && r < 16 {
		v.VI[r&15] = val
	}
}

// sane clamps the values Go arithmetic produces that a VU cannot: no NaN, no infinity —
// the hardware's exponent saturates instead.
func sane(f float32) float32 {
	if math.IsNaN(float64(f)) {
		return 0
	}
	if f > math.MaxFloat32 {
		return math.MaxFloat32
	}
	if f < -math.MaxFloat32 {
		return -math.MaxFloat32
	}
	return f
}

func le64m(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

// --- the upper pipe ---------------------------------------------------------------

// upperResult is what the FMAC half of a pair wants written, computed before the lower
// half runs and committed after it — the one visible consequence of the parallel issue.
type upperResult struct {
	kind uint8 // which commit applies
	reg  uint32
	mask uint32
	val  [4]float32
	bits [4]uint32
	clip uint32
}

const (
	upNone  = iota
	upVF    // val -> VF[reg] through mask, with MAC flags
	upACC   // val -> ACC through mask, with MAC flags
	upBits  // bits -> VF[reg] through mask, no flags (the conversions and moves)
	upClip  // clip judged; shift the ring
)

// execUpper computes the FMAC half into res.
func (v *VU) execUpper(i uint32, res *upperResult) {
	res.kind = upNone
	op := i & 0x3F
	if op < 0x30 || op >= 0x3C {
		// fall through to the decode below
	} else {
		return
	}

	d := i >> 21 & 0xF
	t, s, f := i>>16&0x1F, i>>11&0x1F, i>>6&0x1F

	var vs, vt [4]float32
	for l := uint32(0); l < 4; l++ {
		vs[l] = v.getF(s, l)
		vt[l] = v.getF(t, l)
	}

	// The second operand: a vector, or a broadcast of one lane / Q / I.
	setB := func(x float32) [4]float32 { return [4]float32{x, x, x, x} }

	vf := func(r uint32, val [4]float32) {
		res.kind, res.reg, res.mask, res.val = upVF, r, d, val
	}
	acc := func(val [4]float32) {
		res.kind, res.mask, res.val = upACC, d, val
	}

	addv := func(b [4]float32) [4]float32 {
		return [4]float32{vs[0] + b[0], vs[1] + b[1], vs[2] + b[2], vs[3] + b[3]}
	}
	subv := func(b [4]float32) [4]float32 {
		return [4]float32{vs[0] - b[0], vs[1] - b[1], vs[2] - b[2], vs[3] - b[3]}
	}
	mulv := func(b [4]float32) [4]float32 {
		return [4]float32{vs[0] * b[0], vs[1] * b[1], vs[2] * b[2], vs[3] * b[3]}
	}
	maddv := func(b [4]float32) [4]float32 {
		return [4]float32{v.ACC[0] + vs[0]*b[0], v.ACC[1] + vs[1]*b[1], v.ACC[2] + vs[2]*b[2], v.ACC[3] + vs[3]*b[3]}
	}
	msubv := func(b [4]float32) [4]float32 {
		return [4]float32{v.ACC[0] - vs[0]*b[0], v.ACC[1] - vs[1]*b[1], v.ACC[2] - vs[2]*b[2], v.ACC[3] - vs[3]*b[3]}
	}
	maxv := func(b [4]float32) [4]float32 {
		var r [4]float32
		for l := 0; l < 4; l++ {
			if vs[l] > b[l] {
				r[l] = vs[l]
			} else {
				r[l] = b[l]
			}
		}
		return r
	}
	minv := func(b [4]float32) [4]float32 {
		var r [4]float32
		for l := 0; l < 4; l++ {
			if vs[l] < b[l] {
				r[l] = vs[l]
			} else {
				r[l] = b[l]
			}
		}
		return r
	}

	bcVal := vt[i&3]

	switch {
	case op < 0x04:
		vf(f, addv(setB(bcVal)))
	case op < 0x08:
		vf(f, subv(setB(bcVal)))
	case op < 0x0C:
		vf(f, maddv(setB(bcVal)))
	case op < 0x10:
		vf(f, msubv(setB(bcVal)))
	case op < 0x14:
		vf(f, maxv(setB(bcVal)))
	case op < 0x18:
		vf(f, minv(setB(bcVal)))
	case op < 0x1C:
		vf(f, mulv(setB(bcVal)))
	case op == 0x1C:
		vf(f, mulv(setB(v.Q)))
	case op == 0x1D:
		vf(f, maxv(setB(v.I)))
	case op == 0x1E:
		vf(f, mulv(setB(v.I)))
	case op == 0x1F:
		vf(f, minv(setB(v.I)))
	case op == 0x20:
		vf(f, addv(setB(v.Q)))
	case op == 0x21:
		vf(f, maddv(setB(v.Q)))
	case op == 0x22:
		vf(f, addv(setB(v.I)))
	case op == 0x23:
		vf(f, maddv(setB(v.I)))
	case op == 0x24:
		vf(f, subv(setB(v.Q)))
	case op == 0x25:
		vf(f, msubv(setB(v.Q)))
	case op == 0x26:
		vf(f, subv(setB(v.I)))
	case op == 0x27:
		vf(f, msubv(setB(v.I)))
	case op == 0x28:
		vf(f, addv(vt))
	case op == 0x29:
		vf(f, maddv(vt))
	case op == 0x2A:
		vf(f, mulv(vt))
	case op == 0x2B:
		vf(f, maxv(vt))
	case op == 0x2C:
		vf(f, subv(vt))
	case op == 0x2D:
		vf(f, msubv(vt))
	case op == 0x2E: // opmsub: the outer product's second half, xyz only
		res.kind, res.reg, res.mask = upVF, f, d&0xE
		res.val = [4]float32{
			v.ACC[0] - vs[1]*vt[2],
			v.ACC[1] - vs[2]*vt[0],
			v.ACC[2] - vs[0]*vt[1],
		}
	case op == 0x2F:
		vf(f, minv(vt))
	default: // 0x3C..0x3F: the wide group
		op2 := i>>4&0x7C | i&3
		switch {
		case op2 < 0x04:
			acc(addv(setB(bcVal)))
		case op2 < 0x08:
			acc(subv(setB(bcVal)))
		case op2 < 0x0C:
			acc(maddv(setB(bcVal)))
		case op2 < 0x10:
			acc(msubv(setB(bcVal)))
		case op2 <= 0x13: // ITOF0/4/12/15
			shift := [4]float32{1, 16, 4096, 32768}[op2-0x10]
			res.kind, res.reg, res.mask = upBits, t, d
			for l := uint32(0); l < 4; l++ {
				res.bits[l] = math.Float32bits(sane(float32(int32(v.VF[s][l])) / shift))
			}
		case op2 <= 0x17: // FTOI0/4/12/15
			shift := [4]float32{1, 16, 4096, 32768}[op2-0x14]
			res.kind, res.reg, res.mask = upBits, t, d
			for l := uint32(0); l < 4; l++ {
				fv := float64(v.getF(s, l)) * float64(shift)
				switch {
				case fv > math.MaxInt32:
					fv = math.MaxInt32
				case fv < math.MinInt32:
					fv = math.MinInt32
				case math.IsNaN(fv):
					fv = 0
				}
				res.bits[l] = uint32(int32(fv))
			}
		case op2 < 0x1C:
			acc(mulv(setB(bcVal)))
		case op2 == 0x1C:
			acc(mulv(setB(v.Q)))
		case op2 == 0x1D: // ABS
			res.kind, res.reg, res.mask = upBits, t, d
			for l := 0; l < 4; l++ {
				res.bits[l] = v.VF[s][l] &^ (1 << 31)
			}
		case op2 == 0x1E:
			acc(mulv(setB(v.I)))
		case op2 == 0x1F: // CLIP: six judgements against |ft.w|, shifted into the ring
			w := vt[3]
			if w < 0 {
				w = -w
			}
			var bits uint32
			if vs[0] > w {
				bits |= 1 << 0
			}
			if vs[0] < -w {
				bits |= 1 << 1
			}
			if vs[1] > w {
				bits |= 1 << 2
			}
			if vs[1] < -w {
				bits |= 1 << 3
			}
			if vs[2] > w {
				bits |= 1 << 4
			}
			if vs[2] < -w {
				bits |= 1 << 5
			}
			res.kind, res.clip = upClip, bits
		case op2 == 0x20:
			acc(addv(setB(v.Q)))
		case op2 == 0x21:
			acc(maddv(setB(v.Q)))
		case op2 == 0x22:
			acc(addv(setB(v.I)))
		case op2 == 0x23:
			acc(maddv(setB(v.I)))
		case op2 == 0x24:
			acc(subv(setB(v.Q)))
		case op2 == 0x25:
			acc(msubv(setB(v.Q)))
		case op2 == 0x26:
			acc(subv(setB(v.I)))
		case op2 == 0x27:
			acc(msubv(setB(v.I)))
		case op2 == 0x28:
			acc(addv(vt))
		case op2 == 0x29:
			acc(maddv(vt))
		case op2 == 0x2A:
			acc(mulv(vt))
		case op2 == 0x2C:
			acc(subv(vt))
		case op2 == 0x2D:
			acc(msubv(vt))
		case op2 == 0x2E: // opmula: the outer product's first half, xyz only
			res.kind, res.mask = upACC, d&0xE
			res.val = [4]float32{vs[1] * vt[2], vs[2] * vt[0], vs[0] * vt[1]}
		}
	}
}

// commitUpper writes the FMAC result after the lower half has read its operands.
func (v *VU) commitUpper(res *upperResult) {
	switch res.kind {
	case upVF:
		v.setLanes(res.reg, res.mask, res.val)
		v.macFlags(res.mask, res.val)
	case upACC:
		for l := uint32(0); l < 4; l++ {
			if res.mask&(8>>l) != 0 {
				v.ACC[l] = sane(res.val[l])
			}
		}
		v.macFlags(res.mask, res.val)
	case upBits:
		v.setBits(res.reg, res.mask, res.bits)
	case upClip:
		v.Clip = (v.Clip<<6 | res.clip) & 0xFFFFFF
	}
}


// macFlags files the per-lane zero/sign of a result, x in the high bit of each nibble
// (the xyzw-from-MSB order the manual draws), and folds them into the status flags.
func (v *VU) macFlags(destMask uint32, val [4]float32) {
	var z, sgn uint16
	for l := uint32(0); l < 4; l++ {
		if destMask&(8>>l) == 0 {
			continue
		}
		bit := uint16(8 >> l)
		if val[l] == 0 {
			z |= bit
		}
		if math.Signbit(float64(val[l])) {
			sgn |= bit
		}
	}
	v.Mac = z | sgn<<4
	v.Status = v.Status & 0xFC0
	if z != 0 {
		v.Status |= 1 | 1<<6
	}
	if sgn != 0 {
		v.Status |= 2 | 2<<6
	}
}

// --- the lower pipe ----------------------------------------------------------------

// execLower runs the integer half. It returns a branch target (byte address) when the
// instruction branches, or -1.
func (v *VU) execLower(i uint32) int64 {
	if i == 0 {
		return -1
	}
	op7 := i >> 25 & 0x7F
	d := i >> 21 & 0xF
	t, s := i>>16&0x1F, i>>11&0x1F
	imm11 := int32(i<<21) >> 21
	imm15 := i>>10&0x7800 | i&0x7FF
	dmask := uint32(len(v.Data)-1) &^ 15

	target := func() int64 { return int64(v.PC+8+uint32(imm11)*8) & int64(uint32(len(v.Micro)-1)) }

	switch op7 {
	case 0x00: // LQ
		v.loadQ(t, d, (uint32(int32(v.VI[s&15])+imm11)*16)&dmask)
	case 0x01: // SQ
		v.storeQ(s, d, (uint32(int32(v.VI[t&15])+imm11)*16)&dmask)
	case 0x04: // ILW
		a := (uint32(int32(v.VI[s&15])+imm11)*16 + laneOf(d)*4) & (uint32(len(v.Data)-1) &^ 3)
		v.setVI(t, uint16(uint32(v.Data[a])|uint32(v.Data[a+1])<<8))
	case 0x05: // ISW
		a := (uint32(int32(v.VI[s&15])+imm11)*16 + laneOf(d)*4) & (uint32(len(v.Data)-1) &^ 3)
		val := v.VI[t&15]
		v.Data[a] = byte(val)
		v.Data[a+1] = byte(val >> 8)
		v.Data[a+2] = 0
		v.Data[a+3] = 0
	case 0x08:
		v.setVI(t, v.VI[s&15]+uint16(imm15))
	case 0x09:
		v.setVI(t, v.VI[s&15]-uint16(imm15))
	case 0x10: // FCEQ
		v.setVI(1, b2u(v.visClip&0xFFFFFF == i&0xFFFFFF))
	case 0x11: // FCSET: an architectural write — the whole pipeline sees it
		v.Clip = i & 0xFFFFFF
		for st := range v.flagPipe {
			v.flagPipe[st].clip = v.Clip
		}
	case 0x12: // FCAND
		v.setVI(1, b2u(v.visClip&i&0xFFFFFF != 0))
	case 0x13: // FCOR
		v.setVI(1, b2u((v.visClip|i&0xFFFFFF)&0xFFFFFF == 0xFFFFFF))
	case 0x14: // FSEQ
		v.setVI(t, b2u(uint32(v.visStatus) == i>>10&0x800|i&0x7FF))
	case 0x15: // FSSET: only the sticky bits are writable, everywhere in the pipe
		v.Status = v.Status&0x3F | uint16(i>>10&0x800|i&0x7FF)&0xFC0
		for st := range v.flagPipe {
			v.flagPipe[st].status = v.flagPipe[st].status&0x3F | v.Status&0xFC0
		}
	case 0x16: // FSAND
		v.setVI(t, v.visStatus&uint16(i>>10&0x800|i&0x7FF))
	case 0x17: // FSOR
		v.setVI(t, v.visStatus|uint16(i>>10&0x800|i&0x7FF))
	case 0x18: // FMEQ
		v.setVI(t, b2u(v.visMac == v.VI[s&15]))
	case 0x1A: // FMAND
		v.setVI(t, v.visMac&v.VI[s&15])
	case 0x1B: // FMOR
		v.setVI(t, v.visMac|v.VI[s&15])
	case 0x1C: // FCGET
		v.setVI(t, uint16(v.visClip&0xFFF))
	case 0x20: // B
		return target()
	case 0x21: // BAL
		v.setVI(t, uint16((v.PC+16)/8))
		return target()
	case 0x24: // JR
		return int64(uint32(v.VI[s&15])*8) & int64(uint32(len(v.Micro)-1))
	case 0x25: // JALR
		v.setVI(t, uint16((v.PC+16)/8))
		return int64(uint32(v.VI[s&15])*8) & int64(uint32(len(v.Micro)-1))
	case 0x28:
		if v.VI[t&15] == v.VI[s&15] {
			return target()
		}
	case 0x29:
		if v.VI[t&15] != v.VI[s&15] {
			return target()
		}
	case 0x2C:
		if int16(v.VI[s&15]) < 0 {
			return target()
		}
	case 0x2D:
		if int16(v.VI[s&15]) > 0 {
			return target()
		}
	case 0x2E:
		if int16(v.VI[s&15]) <= 0 {
			return target()
		}
	case 0x2F:
		if int16(v.VI[s&15]) >= 0 {
			return target()
		}
	case 0x40:
		return v.execLowerSpecial(i)
	}
	return -1
}

// execLowerSpecial is the op7 == 0x40 group.
func (v *VU) execLowerSpecial(i uint32) int64 {
	d := i >> 21 & 0xF
	t, s, f := i>>16&0x1F, i>>11&0x1F, i>>6&0x1F
	dmask := uint32(len(v.Data)-1) &^ 15

	switch i & 0x3F {
	case 0x30:
		v.setVI(f, v.VI[s&15]+v.VI[t&15])
		return -1
	case 0x31:
		v.setVI(f, v.VI[s&15]-v.VI[t&15])
		return -1
	case 0x32:
		v.setVI(t, uint16(int16(v.VI[s&15])+int16(int32(i<<21)>>27)))
		return -1
	case 0x34:
		v.setVI(f, v.VI[s&15]&v.VI[t&15])
		return -1
	case 0x35:
		v.setVI(f, v.VI[s&15]|v.VI[t&15])
		return -1
	}
	if i&0x3C != 0x3C {
		return -1
	}

	op2 := i>>4&0x7C | i&3
	fsf := i >> 21 & 3
	ftf := i >> 23 & 3
	switch op2 {
	case 0x30: // MOVE
		v.setBits(t, d, v.VF[s])
	case 0x31: // MR32: x<-y, y<-z, z<-w, w<-x
		src := v.VF[s]
		v.setBits(t, d, [4]uint32{src[1], src[2], src[3], src[0]})
	case 0x34: // LQI
		v.loadQ(t, d, uint32(v.VI[s&15])*16&dmask)
		v.setVI(s, v.VI[s&15]+1)
	case 0x35: // SQI
		v.storeQ(s, d, uint32(v.VI[t&15])*16&dmask)
		v.setVI(t, v.VI[t&15]+1)
	case 0x36: // LQD
		v.setVI(s, v.VI[s&15]-1)
		v.loadQ(t, d, uint32(v.VI[s&15])*16&dmask)
	case 0x37: // SQD
		v.setVI(t, v.VI[t&15]-1)
		v.storeQ(s, d, uint32(v.VI[t&15])*16&dmask)
	case 0x38: // DIV
		v.Q = vuDiv(v.getF(s, fsf), v.getF(t, ftf))
	case 0x39: // SQRT
		v.Q = sane(float32(math.Sqrt(math.Abs(float64(v.getF(t, ftf))))))
	case 0x3A: // RSQRT
		v.Q = vuDiv(v.getF(s, fsf), float32(math.Sqrt(math.Abs(float64(v.getF(t, ftf))))))
	case 0x3B: // WAITQ: Q is always ready here
	case 0x3C: // MTIR
		v.setVI(t, uint16(v.VF[s][fsf]))
	case 0x3D: // MFIR
		bits := uint32(int32(int16(v.VI[s&15])))
		v.setBits(t, d, [4]uint32{bits, bits, bits, bits})
	case 0x3E: // ILWR
		a := (uint32(v.VI[s&15])*16 + laneOf(d)*4) & (uint32(len(v.Data)-1) &^ 3)
		v.setVI(t, uint16(uint32(v.Data[a])|uint32(v.Data[a+1])<<8))
	case 0x3F: // ISWR
		a := (uint32(v.VI[s&15])*16 + laneOf(d)*4) & (uint32(len(v.Data)-1) &^ 3)
		val := v.VI[t&15]
		v.Data[a] = byte(val)
		v.Data[a+1] = byte(val >> 8)
		v.Data[a+2] = 0
		v.Data[a+3] = 0
	case 0x40: // RNEXT: advance the 23-bit sequence, then read it
		v.rNext()
		fallthrough
	case 0x41: // RGET
		bits := 0x3F800000 | v.R&0x7FFFFF
		v.setBits(t, d, [4]uint32{bits, bits, bits, bits})
	case 0x42: // RINIT
		v.R = v.VF[s][fsf] & 0x7FFFFF
	case 0x43: // RXOR
		v.R = (v.R ^ v.VF[s][fsf]) & 0x7FFFFF
	case 0x64: // MFP
		bits := math.Float32bits(v.P)
		v.setBits(t, d, [4]uint32{bits, bits, bits, bits})
	case 0x68: // XTOP
		v.setVI(t, v.Top)
	case 0x69: // XITOP
		v.setVI(t, v.ITop)
	case 0x6C: // XGKICK
		if v.XGKick != nil {
			v.XGKick(uint32(v.VI[s&15]) & uint32(len(v.Data)/16-1))
		}
	case 0x70: // ESADD
		v.P = sane(dot3(v.VF[s]))
	case 0x71: // ERSADD
		v.P = vuDiv(1, dot3(v.VF[s]))
	case 0x72: // ELENG
		v.P = sane(float32(math.Sqrt(float64(dot3(v.VF[s])))))
	case 0x73: // ERLENG
		v.P = vuDiv(1, float32(math.Sqrt(float64(dot3(v.VF[s])))))
	case 0x74: // EATANxy
		v.P = sane(float32(math.Atan2(float64(v.getF(s, 1)), float64(v.getF(s, 0)))))
	case 0x75: // EATANxz
		v.P = sane(float32(math.Atan2(float64(v.getF(s, 2)), float64(v.getF(s, 0)))))
	case 0x76: // ESUM
		v.P = sane(v.getF(s, 0) + v.getF(s, 1) + v.getF(s, 2) + v.getF(s, 3))
	case 0x78: // ESQRT
		v.P = sane(float32(math.Sqrt(math.Abs(float64(v.getF(s, fsf))))))
	case 0x79: // ERSQRT
		v.P = vuDiv(1, float32(math.Sqrt(math.Abs(float64(v.getF(s, fsf))))))
	case 0x7A: // ERCPR
		v.P = vuDiv(1, v.getF(s, fsf))
	case 0x7B: // WAITP
	case 0x7C: // ESIN
		v.P = sane(float32(math.Sin(float64(v.getF(s, fsf)))))
	case 0x7D: // EATAN
		v.P = sane(float32(math.Atan(float64(v.getF(s, fsf)))))
	case 0x7E: // EEXP: e^-x, the EFU's decaying exponential
		v.P = sane(float32(math.Exp(-float64(v.getF(s, fsf)))))
	}
	return -1
}

// loadQ reads data memory into a register's masked lanes.
func (v *VU) loadQ(r, destMask, addr uint32) {
	var val [4]uint32
	for l := uint32(0); l < 4; l++ {
		o := addr + l*4
		val[l] = uint32(v.Data[o]) | uint32(v.Data[o+1])<<8 | uint32(v.Data[o+2])<<16 | uint32(v.Data[o+3])<<24
	}
	v.setBits(r, destMask, val)
}

// storeQ writes a register's masked lanes to data memory.
func (v *VU) storeQ(r, destMask, addr uint32) {
	for l := uint32(0); l < 4; l++ {
		if destMask&(8>>l) == 0 {
			continue
		}
		o := addr + l*4
		bits := v.VF[r][l]
		v.Data[o] = byte(bits)
		v.Data[o+1] = byte(bits >> 8)
		v.Data[o+2] = byte(bits >> 16)
		v.Data[o+3] = byte(bits >> 24)
	}
}

// laneOf maps a single-bit dest mask to its lane index (ILW/ISW name one lane this way).
func laneOf(destMask uint32) uint32 {
	switch destMask {
	case 8:
		return 0
	case 4:
		return 1
	case 2:
		return 2
	default:
		return 3
	}
}

// vuDiv is the DIV/RSQRT contract: a zero divisor yields the largest value of the
// quotient's sign instead of an infinity.
func vuDiv(a, b float32) float32 {
	if b == 0 {
		if math.Signbit(float64(a)) != math.Signbit(float64(b)) {
			return -math.MaxFloat32
		}
		return math.MaxFloat32
	}
	return sane(a / b)
}

func dot3(r [4]uint32) float32 {
	x := math.Float32frombits(r[0])
	y := math.Float32frombits(r[1])
	z := math.Float32frombits(r[2])
	return x*x + y*y + z*z
}

// rNext advances the R register's 23-bit sequence. The hardware's exact feedback taps
// are not documented in what this repository derives from; a maximal 23-bit LFSR stands
// in, which keeps RGET/RNEXT random-looking without claiming the silicon's sequence.
func (v *VU) rNext() {
	bit := (v.R >> 22) ^ (v.R >> 17)
	v.R = (v.R<<1 | bit&1) & 0x7FFFFF
}

// b2u is a boolean as the 0/1 the flag instructions store.
func b2u(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}
