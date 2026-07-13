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
			base = geBaseAddr(arg)
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

// geBaseAddr decodes the BASE register: it supplies the high address bits for
// the 24-bit addresses in VADDR/IADDR/JUMP/CALL, but the GE only implements
// bits 16-19 of the argument (address bits 24-27) — the memory map never goes
// higher. Burnout sends BASE 0x480000 for geometry streamed into the volatile
// block; taking the 0x40 literally addressed 0x484038FC (nowhere) instead of
// 0x084038FC, so every primitive drawn out of the streaming buffers read
// garbage vertices.
func geBaseAddr(arg uint32) uint32 { return (arg & 0x000F0000) << 8 }

// execGeList executes a captured display list into the framebuffer (ge_raster.go).
func (m *Machine) execGeList(list GeList) { m.rasterList(list) }

// geCmdNames names every GE command the rasteriser acts on, and the well-known ones it
// ignores. A frame debugger lists the whole command stream, not just the draws — the
// register writes that conditioned a draw are as much a part of the frame as the draw,
// and a list of anonymous hex bytes is a list nobody can read. What is genuinely unknown
// stays unknown (an empty entry prints as its hex opcode), because inventing a name for
// a register we have not reversed would be the one thing worse than not naming it.
var geCmdNames = func() [256]string {
	n := [256]string{
		0x00: "NOP", 0x01: "VADDR", 0x02: "IADDR", 0x04: "PRIM", 0x05: "BEZIER",
		0x06: "SPLINE", 0x07: "BBOX", 0x08: "JUMP", 0x09: "BJUMP", 0x0A: "CALL",
		0x0B: "RET", 0x0C: "END", 0x0F: "FINISH", 0x10: "BASE", 0x12: "VTYPE",
		0x13: "OFFADDR", 0x14: "ORIGIN", 0x15: "REGION1", 0x16: "REGION2",
		0x17: "LIGHTING", 0x1C: "CLUTON", 0x1D: "CULLON", 0x1E: "TEXENABLE",
		0x1F: "FOGON", 0x20: "DITHERON", 0x21: "BLENDON", 0x22: "ALPHATESTON",
		0x23: "ZTESTON", 0x24: "STENCILON", 0x25: "AAON", 0x26: "PATCHCULLON",
		0x27: "COLORTESTON", 0x28: "LOGICOPON", 0x2E: "BONEMTXN", 0x2F: "BONEMTXD",
		0x36: "PATCHDIV", 0x37: "PATCHPRIM", 0x38: "PATCHFACE",
		0x3A: "WORLDN", 0x3B: "WORLDD", 0x3C: "VIEWN", 0x3D: "VIEWD",
		0x3E: "PROJN", 0x3F: "PROJD", 0x40: "TEXMTXN", 0x41: "TEXMTXD",
		0x42: "VPXSCALE", 0x43: "VPYSCALE", 0x44: "VPZSCALE",
		0x45: "VPXCENTER", 0x46: "VPYCENTER", 0x47: "VPZCENTER",
		0x48: "TEXSCALEU", 0x49: "TEXSCALEV", 0x4A: "TEXOFFSETU", 0x4B: "TEXOFFSETV",
		0x4C: "OFFSETX", 0x4D: "OFFSETY", 0x50: "SHADEMODE", 0x51: "NORMALREV",
		0x53: "MATCOLORMODE", 0x54: "MATEMISSIVE", 0x55: "MATAMBIENT",
		0x56: "MATDIFFUSE", 0x57: "MATSPECULAR", 0x58: "MATALPHA",
		0x5B: "MATSPECCOEF", 0x5C: "AMBIENTCOL", 0x5D: "AMBIENTALPHA",
		0x5E: "LIGHTMODE", 0x8F: "SPOTCOEF",
		0x9B: "CULLFACE", 0x9C: "FBP", 0x9D: "FBW", 0x9E: "ZBP", 0x9F: "ZBW",
		0xB0: "CLUTADDR", 0xB1: "CLUTADDRH", 0xB2: "TRSRC", 0xB3: "TRSRCW",
		0xB4: "TRDST", 0xB5: "TRDSTW", 0xC0: "TEXMAPMODE", 0xC1: "TEXSHADELS",
		0xC2: "TEXMODE", 0xC3: "TEXFORMAT", 0xC4: "LOADCLUT", 0xC5: "CLUTFORMAT",
		0xC6: "TEXFILTER", 0xC7: "TEXWRAP", 0xC8: "TEXLEVEL", 0xC9: "TEXFUNC",
		0xCA: "TEXENVCOL", 0xCB: "TEXFLUSH", 0xCC: "TEXSYNC",
		0xCD: "FOG1", 0xCE: "FOG2", 0xCF: "FOGCOLOR",
		0xD2: "FBPIXFMT", 0xD3: "CLEARMODE", 0xD4: "SCISSOR1", 0xD5: "SCISSOR2",
		0xD6: "NEARZ", 0xD7: "FARZ", 0xDB: "ATEST", 0xDC: "STEST", 0xDD: "SOP",
		0xDE: "ZTEST", 0xDF: "BLENDFUNC", 0xE0: "BLENDFIXA", 0xE1: "BLENDFIXB",
		0xE2: "DITH0", 0xE3: "DITH1", 0xE4: "DITH2", 0xE5: "DITH3",
		0xE6: "LOGICOP", 0xE7: "ZMASK", 0xE8: "PMSKC", 0xE9: "PMSKA",
		0xEA: "TRSTART", 0xEB: "TRSRCPOS", 0xEC: "TRDSTPOS", 0xEE: "TRSIZE",
	}
	// The indexed families: eight texture levels, four lights.
	for i := 0; i < 8; i++ {
		n[0xA0+i] = fmt.Sprintf("TEXADDR%d", i)
		n[0xA8+i] = fmt.Sprintf("TEXBW%d", i)
		n[0xB8+i] = fmt.Sprintf("TEXSIZE%d", i)
	}
	for i := 0; i < 4; i++ {
		n[0x18+i] = fmt.Sprintf("LIGHT%dON", i)
		n[0x5F+i] = fmt.Sprintf("LIGHTTYPE%d", i)
		n[0x90+i] = fmt.Sprintf("LIGHTDIF%d", i)
		n[0x94+i] = fmt.Sprintf("LIGHTSPC%d", i)
		n[0x98+i] = fmt.Sprintf("LIGHTAMB%d", i)
		for c := 0; c < 3; c++ {
			n[0x63+i*3+c] = fmt.Sprintf("LIGHTPOS%d%c", i, "XYZ"[c])
			n[0x6F+i*3+c] = fmt.Sprintf("LIGHTDIR%d%c", i, "XYZ"[c])
			n[0x7B+i*3+c] = fmt.Sprintf("LIGHTATT%d%c", i, "XYZ"[c])
		}
	}
	return n
}()

// GeCmdName returns a short mnemonic for a GE command byte, for list dumps.
func GeCmdName(cmd uint32) string {
	if cmd < 256 {
		if s := geCmdNames[cmd]; s != "" {
			return s
		}
	}
	return fmt.Sprintf("0x%02X", cmd)
}
