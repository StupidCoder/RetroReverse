package threedo

import "fmt"

// graphicsfolio.go high-level-emulates the Portfolio "Graphics" folio — the calls
// the game makes to set up and display its screens. Like the File folio, it is
// reached through a negative-offset vector table: the game LookupItem's the
// "Graphics" folio to gfxFolioBase (folio.go) and calls `LDR pc, [base, #-N]`,
// which the planted vectors route into the HLE window's Graphics slice, landing
// in serviceGraphicsFolio below.
//
// This game (EA) drives the folio with a CUSTOM ABI rather than the stock SDK
// tag-arg calls: it builds a Video Display List (VDL) + screen descriptor struct
// itself and submits it directly. The vector offsets are labelled from the stock
// intgraf.c GrafUserFuncs table (byte offset = 4*(idx+1)) as a guide, but the
// argument shapes are the game's own. Offsets seen from the boot trace:
//
//	-0x08 GetVBLAttrs   -0x0C SetVBLAttrs    -0x14 QueryGraphicsList
//	-0x38 DisplayScreen -0x40 SetCEControl   -0x50 FillRect
//	-0x68 SetClipWidth  -0x70 AddScreenGroup -0x74 SetBGPen  -0x78 SetFGPen
//	-0xA0 ResetReadAddress -0xA4 SetReadAddress -0xA8 CreateScreenGroup
//	-0xC0 (EA VBL-counter register: r0=enable, r1=&elapsedFieldCount)
//
// The goal is to let the boot get a live screen and start drawing into its own
// VRAM buffers, which the oracle then captures with -shot.

// gfxFuncName labels a Graphics-folio vector offset for the trace.
func gfxFuncName(foff uint32) string {
	switch foff {
	case 0x08:
		return "GetVBLAttrs"
	case 0x0C:
		return "SetVBLAttrs"
	case 0x14:
		return "QueryGraphicsList"
	case 0x38:
		return "DisplayScreen"
	case 0x40:
		return "SetCEControl"
	case 0x50:
		return "FillRect"
	case 0x68:
		return "SetClipWidth"
	case 0x70:
		return "AddScreenGroup"
	case 0x74:
		return "SetBGPen"
	case 0x78:
		return "SetFGPen"
	case 0xA0:
		return "ResetReadAddress"
	case 0xA4:
		return "SetReadAddress"
	case 0xA8:
		return "CreateScreenGroup"
	case 0xC0:
		return "RegisterVBLCounter"
	default:
		return "?"
	}
}

// serviceGraphicsFolio dispatches an intercepted Graphics-folio vector call by its
// (positive) byte offset and returns to the caller with a result in r0. Calls not
// yet modelled return 0 (a non-negative success the game's error checks accept)
// and are logged with their call site so the next one to model is visible.
func (m *Machine) serviceGraphicsFolio(foff uint32) {
	c := m.CPU
	r0, r1, r2 := c.Reg(0), c.Reg(1), c.Reg(2)
	m.note(fmt.Sprintf("Graphics[-0x%X] %s from 0x%08X (r0=0x%08X r1=0x%08X r2=0x%08X)",
		foff, gfxFuncName(foff), c.Reg(14)-8, r0, r1, r2))

	switch foff {
	case 0xC0: // EA VBL-counter register: r0=enable, r1=&elapsedFieldCount.
		// The game hands the folio the address of its own elapsed-field counter for
		// the VBL interrupt to advance. We have no interrupt, so record the address
		// and let the virtual VBL manager keep it in step (retires the hardcoded
		// -vblmirror default — the folio now discovers it the way hardware does).
		if r1 != 0 {
			m.SetVBLMirror(r1)
		}
		m.SetResultAndReturn(0)
	case 0xA8: // CreateScreenGroup: hand back a valid (non-negative) screen item.
		m.SetResultAndReturn(m.createScreenItem())
	default:
		// Non-negative default so the game's MI (error) checks pass.
		m.SetResultAndReturn(0)
	}
}

// createScreenItem allocates a kernel item to stand in for a Screen/ScreenGroup so
// the game holds a distinct non-negative handle it can pass on (e.g. to
// SetReadAddress) and later DeleteItem.
func (m *Machine) createScreenItem() uint32 {
	it := m.createItem(0x0A03, 0, 0) // graphics folio (3) << 8 | screen subtype (approx)
	return uint32(it.num)
}
