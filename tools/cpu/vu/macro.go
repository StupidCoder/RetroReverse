package vu

// macro.go is VU0's second face: macro mode, where the EE reaches the vector unit one
// instruction at a time through the COP2 opcode. The instruction word's low bits carry
// the same field layout as a micro-mode UPPER instruction — dest, ft, fs, fd, opcode —
// and the wide-group encoding additionally hosts what micro mode keeps on the lower
// pipe (DIV, MOVE, MTIR, the EFU), plus VCALLMS, which starts a microprogram on the
// unit proper. So macro mode is a dispatcher over the two executors exec.go already
// has, and the same VU instance serves both faces: the VIF fills its memories, the EE
// pokes at its registers.
//
// The methods implement the r5900's Coprocessor2 interface.

// Macro executes one macro-mode COP2 instruction word.
func (v *VU) Macro(w uint32) {
	op := w & 0x3F
	if op < 0x30 {
		var res upperResult
		v.execUpper(w, &res)
		v.commitUpper(&res)
		return
	}
	if op < 0x3C {
		switch op {
		case 0x38: // VCALLMS: run the microprogram at imm15*8
			v.Run(w>>6&0x7FFF*8, 1<<20)
		case 0x39: // VCALLMSR: run at the address CMSAR0 holds
			v.Run(uint32(v.CMSAR0)*8, 1<<20)
		default:
			// The integer ops (VIADD, VISUB, VIADDI, VIAND, VIOR) share the
			// lower-special encodings, and execLowerSpecial already serves them.
			v.execLowerSpecial(w)
		}
		return
	}
	op2 := w>>4&0x7C | w&3
	if op2 <= 0x2F {
		var res upperResult
		v.execUpper(w, &res)
		v.commitUpper(&res)
		return
	}
	v.execLowerSpecial(w)
}

// CMSAR0 is the start address VCALLMSR uses, in 64-bit pairs.
// (Declared here: it belongs to the macro face alone.)

// ReadVF reads one 32-bit field of a vector register (qmfc2/sqc2).
func (v *VU) ReadVF(reg, field uint32) uint32 {
	return v.VF[reg&31][field&3]
}

// WriteVF writes one field (qmtc2/lqc2). VF00 stays pinned.
func (v *VU) WriteVF(reg, field, val uint32) {
	if reg&31 != 0 {
		v.VF[reg&31][field&3] = val
	}
}

// The control-register numbers ctc2/cfc2 speak: 0..15 are the integer registers, the
// rest the unit's state.
const (
	ctrlStatus = 16
	ctrlMac    = 17
	ctrlClip   = 18
	ctrlR      = 20
	ctrlI      = 21
	ctrlQ      = 22
	ctrlTPC    = 26
	ctrlCMSAR0 = 27
	ctrlFBRST  = 28
	ctrlVPU    = 29
	ctrlCMSAR1 = 31
)

// ReadCtrl serves cfc2.
func (v *VU) ReadCtrl(reg uint32) uint32 {
	reg &= 31
	if reg < 16 {
		return uint32(v.VI[reg])
	}
	switch reg {
	case ctrlStatus:
		return uint32(v.Status)
	case ctrlMac:
		return uint32(v.Mac)
	case ctrlClip:
		return v.Clip
	case ctrlR:
		return v.R
	case ctrlI:
		return float32bits(v.I)
	case ctrlQ:
		return float32bits(v.Q)
	case ctrlTPC:
		return v.PC / 8
	case ctrlCMSAR0:
		return uint32(v.CMSAR0)
	case ctrlVPU:
		// VPU_STAT: this machine's programs run to completion inside the call that
		// starts them, so neither unit is ever still busy when the EE looks.
		return 0
	}
	return 0
}

// WriteCtrl serves ctc2.
func (v *VU) WriteCtrl(reg, val uint32) {
	reg &= 31
	if reg == 0 {
		return // VI00 is pinned
	}
	if reg < 16 {
		v.VI[reg] = uint16(val)
		return
	}
	switch reg {
	case ctrlStatus:
		v.Status = uint16(val)
	case ctrlMac:
		v.Mac = uint16(val)
	case ctrlClip:
		v.Clip = val & 0xFFFFFF
	case ctrlR:
		v.R = val & 0x7FFFFF
	case ctrlI:
		v.I = float32frombits(val)
	case ctrlQ:
		v.Q = float32frombits(val)
	case ctrlCMSAR0:
		v.CMSAR0 = uint16(val)
	case ctrlFBRST:
		// Force-break / reset bits; nothing here runs asynchronously, so a reset has
		// nothing to interrupt.
	case ctrlCMSAR1:
		// Writing CMSAR1 starts VU1 from the EE. This VU is VU0; VU1 belongs to the
		// VIF. The hook is for the machine to wire if a game ever uses it.
		if v.StartVU1 != nil {
			v.StartVU1(val)
		}
	}
}
