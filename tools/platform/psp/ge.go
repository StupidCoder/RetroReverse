package psp

// ge.go is the PSP graphics engine (GE) — the GPU. A game builds a display list of
// 32-bit commands in memory (top 8 bits = command, low 24 = argument) and submits it
// with sceGeListEnQueue; the GE walks the list, maintaining render state and drawing
// primitives into the framebuffer in VRAM. This file captures a submitted list by
// following its control flow (JUMP/CALL/RET) to the END, and decodes the commands;
// ge_raster.go executes them.

import "fmt"

// GE command opcodes (top byte of each list word).
const (
	geNOP     = 0x00
	geVADDR   = 0x01
	geIADDR   = 0x02
	gePRIM    = 0x04
	geJUMP    = 0x08
	geBJUMP   = 0x09
	geCALL    = 0x0A
	geRET     = 0x0B
	geFINISH  = 0x0F
	geEND     = 0x0C
	geBASE    = 0x10
	geVTYPE   = 0x12
	geOFFADDR = 0x13
	geREGION1 = 0x15
	geREGION2 = 0x16
	geCLEAR   = 0xD3 // CLEAR mode enable/flags
	geWORLD_N = 0x3A
	geWORLD_D = 0x3B
	geVIEW_N  = 0x3C
	geVIEW_D  = 0x3D
	gePROJ_N  = 0x3E
	gePROJ_D  = 0x3F
	geFBP     = 0x9C // frame buffer pointer
	geFBW     = 0x9D // frame buffer width + high bits
)

// GeList is a captured display list: the flattened command words in execution order.
type GeList struct {
	Start uint32
	Words []uint32
}

// captureList walks a display list from addr, following JUMP/CALL/RET, and returns the
// commands in execution order up to FINISH/END (or a safety cap). base is the GE JUMP
// base address (from geBASE); addresses in the list are (base|arg) & 0x0FFFFFFF within
// the CPU space.
func (m *Machine) captureList(addr uint32) GeList {
	const cap = 1 << 20
	out := GeList{Start: addr}
	pc := addr
	base := addr & 0xFF000000
	var callStack []uint32
	seen := 0
	for seen < cap {
		w := m.read32(pc)
		out.Words = append(out.Words, w)
		cmd := w >> 24
		arg := w & 0x00FFFFFF
		pc += 4
		seen++
		switch cmd {
		case geBASE:
			base = (arg << 8) & 0xFF000000
		case geJUMP:
			pc = (base | (arg & 0x00FFFFFF)) &^ 3
		case geCALL:
			callStack = append(callStack, pc)
			pc = (base | (arg & 0x00FFFFFF)) &^ 3
		case geRET:
			if len(callStack) > 0 {
				pc = callStack[len(callStack)-1]
				callStack = callStack[:len(callStack)-1]
			}
		case geFINISH:
			// FINISH is usually followed by END; keep going one step to catch END.
		case geEND:
			return out
		}
		if pc < ramBase && pc >= vramBase+vramSize {
			return out
		}
	}
	return out
}

// execGeList executes a captured display list into the framebuffer (ge_raster.go).
func (m *Machine) execGeList(list GeList) { m.rasterList(list) }

// GeCmdName returns a short mnemonic for a GE command byte, for list dumps.
func GeCmdName(cmd uint32) string {
	switch cmd {
	case geNOP:
		return "NOP"
	case geVADDR:
		return "VADDR"
	case geIADDR:
		return "IADDR"
	case gePRIM:
		return "PRIM"
	case geJUMP:
		return "JUMP"
	case geCALL:
		return "CALL"
	case geRET:
		return "RET"
	case geFINISH:
		return "FINISH"
	case geEND:
		return "END"
	case geBASE:
		return "BASE"
	case geVTYPE:
		return "VTYPE"
	case geOFFADDR:
		return "OFFADDR"
	case geFBP:
		return "FBP"
	case geFBW:
		return "FBW"
	case geCLEAR:
		return "CLEAR"
	}
	return fmt.Sprintf("0x%02X", cmd)
}
