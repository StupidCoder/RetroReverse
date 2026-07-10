package arm

// Variant selects which revision of the little-endian ARM architecture the
// decoder and the execution core implement. The two in use here:
//
//   - V5TE — ARM946E-S / ARM7TDMI, the Nintendo DS's two cores. The baseline
//     this package was written for.
//   - V6K  — ARM11 MPCore, the Nintendo 3DS's application processor. A strict
//     superset of V5TE's instruction set.
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
	V5TE Variant = iota // Nintendo DS: ARM946E-S (ARM9) and ARM7TDMI (ARM7)
	V6K                 // Nintendo 3DS: ARM11 MPCore
)

func (v Variant) String() string {
	switch v {
	case V5TE:
		return "ARMv5TE"
	case V6K:
		return "ARMv6K"
	}
	return "ARM?"
}

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
	if v < V6K {
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
