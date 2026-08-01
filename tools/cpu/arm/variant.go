package arm

// Variant selects which revision of the little-endian ARM architecture the
// decoder and the execution core implement. The three in use here:
//
//   - V5TE — ARM946E-S, the Nintendo DS's ARM9. The baseline this package was
//     written for, and the zero value.
//   - V6K  — ARM11 MPCore, the Nintendo 3DS's application processor. A strict
//     superset of V5TE's instruction set.
//   - V4T  — ARM7TDMI, the Game Boy Advance (and, in principle, the DS's ARM7).
//     A strict SUBSET of V5TE: no BLX (either form), CLZ, LDRD/STRD, PLD, the
//     saturating QADD family, the SMLAxy signed multiplies, or BKPT.
//
// V4T exists for the same reason the V5TE/V6K split does, in the other
// direction. A V5TE decoder pointed at ARMv4T code never *fails*; the v5-only
// encodings are undefined on the real chip, so a listing or a core that decodes
// them anyway invents instructions the hardware does not have — and on a machine
// where those bit patterns are reached as data (a jump table, a mis-set Thumb
// bit, an over-run) the honest answer is "undefined", not a plausible LDRD.
//
// The variant is not an ordering. Earlier code tested `Arch >= V6K` to mean "is
// this the ARMv6 core"; with a third value that test is a bug waiting for its
// third caller, so the checks are explicit equality (isV6) instead.
//
// The variant is not a cosmetic flag. Several ARMv6 instructions are encoded in
// slots that ARMv5TE assigns to *other* instructions rather than leaving
// undefined, so a V5TE decoder does not merely fail to recognise them — it
// silently decodes them as something else:
//
//	LDREX/STREX  sit in the SWP/SWPB slot (bits 27:24 == 0001, 7:4 == 1001)
//	UMAAL        sits in the MUL/MLA slot (bits 27:20 == 0x04)
//
// So the V6K path must be consulted *before* the V5TE decode, not after it as a
// fallback for undefined encodings. The rest of the ARMv6 additions — the media
// space, the packing/saturation/reversal group, CPS/SETEND, CLREX — do live in
// space ARMv5TE leaves undefined, and could have been appended; they are handled
// alongside the two exceptions for one coherent entry point.
//
// ARMv6K is deliberately *not* ARMv6T2: the bitfield instructions (BFI, BFC,
// SBFX, UBFX), MLS, RBIT and the Thumb-2 32-bit encodings arrived with T2 and do
// not exist on the ARM11 MPCore. They are left undefined here rather than
// decoded, so a listing that hits them says so instead of inventing an
// instruction the hardware does not have.
type Variant int

const (
	V5TE Variant = iota // Nintendo DS: ARM946E-S (ARM9). The zero value.
	V6K                 // Nintendo 3DS: ARM11 MPCore
	V4T                 // Game Boy Advance: ARM7TDMI
)

func (v Variant) String() string {
	switch v {
	case V5TE:
		return "ARMv5TE"
	case V6K:
		return "ARMv6K"
	case V4T:
		return "ARMv4T"
	}
	return "ARM?"
}

// isV6 reports the ARMv6K core. Written as equality rather than an ordering so
// that adding a variant cannot silently opt it into the ARMv6 decode paths.
func (v Variant) isV6() bool { return v == V6K }

// v5OrLater reports whether the variant has the ARMv5 additions (BLX, CLZ,
// LDRD/STRD, PLD, the saturating and SMLAxy groups, BKPT).
func (v Variant) v5OrLater() bool { return v != V4T }

// DecodeVariant decodes one instruction for the given architecture variant.
func DecodeVariant(code []byte, addr uint32, thumb bool, v Variant) Inst {
	if thumb {
		// ARMv6K's Thumb is Thumb-1 — the same 16-bit set ARMv5TE implements,
		// plus a handful of ARMv6 additions in the "miscellaneous" space
		// (CPS, SETEND, REV/REV16/REVSH, SXTB/SXTH/UXTB/UXTH). Those are not yet
		// decoded here: 3DS application code is overwhelmingly ARM, so they are
		// left to be added lazily when Thumb code first reaches them, matching
		// this package's convention of implementing encodings on first contact
		// rather than pre-emptively. ARMv6K does *not* include Thumb-2.
		return DecodeThumb(code, addr)
	}
	return DecodeARMVariant(code, addr, v)
}

// DecodeARMVariant decodes one 32-bit ARM instruction at addr for variant v.
func DecodeARMVariant(code []byte, addr uint32, v Variant) Inst {
	if v == V4T {
		return v4tFilter(DecodeARM(code, addr))
	}
	if !v.isV6() {
		return DecodeARM(code, addr)
	}
	w, ok := word(code)
	if !ok {
		return Inst{Addr: addr, Len: len(code), Mnem: ".word", Text: ".word ; truncated", Flow: FlowStop, Cond: condAL}
	}
	in := Inst{Addr: addr, Len: 4, Flow: FlowSeq, Cond: int(w >> 28)}
	if out, handled := decodeARMv6(w, addr, in); handled {
		return out
	}
	return DecodeARM(code, addr)
}

// v4tMnem lists the mnemonics ARMv5/v5TE added, which the ARM7TDMI does not
// have. The Thumb BLX forms are covered by name too (the Thumb decoder emits
// "BLX" for both the register and the long-branch pair).
var v4tMnem = map[string]bool{
	"BLX": true, "CLZ": true, "BKPT": true, "PLD": true,
	"LDRD": true, "STRD": true,
	"QADD": true, "QSUB": true, "QDADD": true, "QDSUB": true,
	"SMLABB": true, "SMLABT": true, "SMLATB": true, "SMLATT": true,
	"SMLAWB": true, "SMLAWT": true, "SMULWB": true, "SMULWT": true,
	"SMLALBB": true, "SMLALBT": true, "SMLALTB": true, "SMLALTT": true,
	"SMULBB": true, "SMULBT": true, "SMULTB": true, "SMULTT": true,
}

// v4tFilter rewrites an instruction the ARMv5 decoder recognised but the
// ARM7TDMI does not implement into an explicit undefined word.
//
// This is the decoder's half of the variant's purpose. A listing that prints
// "LDRD r4, [r0]" for a word the GBA would have refused is not a small
// cosmetic error: it is a fact about the program that is not true, and the
// reader has no way to tell it apart from the ones that are.
func v4tFilter(in Inst) Inst {
	mnem := in.Mnem
	// Strip a condition suffix ("CLZEQ" -> "CLZ") before the lookup.
	for _, c := range condName {
		if c != "" && len(mnem) > len(c) && mnem[len(mnem)-len(c):] == c {
			mnem = mnem[:len(mnem)-len(c)]
			break
		}
	}
	if !v4tMnem[mnem] {
		return in
	}
	return Inst{
		Addr: in.Addr, Len: in.Len, Mnem: ".word", Cond: in.Cond,
		Text: ".word ; " + in.Text + " — undefined on ARMv4T",
		Flow: FlowStop,
	}
}
