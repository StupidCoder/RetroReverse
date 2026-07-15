package gc

// debug.go is the machine's window for a debugger and an oracle: reading memory through
// the CPU's own eyes, decoding the framebuffer, and walking the call stack.
//
// The backtrace is worth the few lines it costs. PowerPC's calling convention keeps a
// linked list of stack frames in memory — each frame's first word points at the caller's
// frame, and the caller's return address sits at a fixed offset in it — so the whole call
// stack can be read out of memory at any moment, with no debug information at all. On a
// machine whose executable has no symbol table, "which functions are we inside" is often
// the only structure available, and this is how it is recovered.

import (
	"fmt"
	"image"
	"strings"
)

// ReadVirt8 and ReadVirt32 read the game's memory through the CPU's translation and locked
// cache, so a debugger sees exactly what the program sees.
func (m *Machine) ReadVirt8(addr uint32) uint8 { return m.CPU.ReadMem(addr) }

func (m *Machine) ReadVirt32(addr uint32) uint32 {
	return uint32(m.CPU.ReadMem(addr))<<24 | uint32(m.CPU.ReadMem(addr+1))<<16 |
		uint32(m.CPU.ReadMem(addr+2))<<8 | uint32(m.CPU.ReadMem(addr+3))
}

// VIField is how many video fields have elapsed — the frame count, and the proof that the
// heartbeat is beating.
func (m *Machine) VIField() uint64 { return m.vi.Field }

// Backtrace walks the stack frames from the current one, returning the return addresses,
// caller by caller. It stops at a null back-chain, a wild pointer, or a depth limit —
// whichever comes first — so a corrupt stack yields a short trace rather than a loop.
func (m *Machine) Backtrace() []uint32 {
	var out []uint32
	sp := m.CPU.GPR[1]
	for depth := 0; depth < 64; depth++ {
		if sp == 0 || sp&3 != 0 || sp >= 0x81800000 || sp < 0x80000000 {
			break
		}
		next := m.ReadVirt32(sp)
		lr := m.ReadVirt32(sp + 4) // the saved link register, at frame+4
		if lr != 0 {
			out = append(out, lr)
		}
		if next <= sp {
			break // the chain must climb toward the top of the stack
		}
		sp = next
	}
	return out
}

// BacktraceString renders it.
func (m *Machine) BacktraceString() string {
	bt := m.Backtrace()
	if len(bt) == 0 {
		return "  (no frames)"
	}
	s := ""
	for i, a := range bt {
		s += fmt.Sprintf("  #%d  0x%08X\n", i, a)
	}
	return s
}

// RenderXFB decodes the framebuffer VI is scanning out — the picture that is on screen —
// into an RGBA image. The XFB is YUV 4:2:2: two pixels share a chroma pair, packed as
// Y0 Cb Y1 Cr, so the decode reads four bytes and produces two pixels.
//
// The dimensions come from the display configuration the game programmed; a standard NTSC
// frame is 640x480. When VI has not been set up — no framebuffer address yet — there is
// nothing to render, and that is reported rather than a black rectangle invented.
func (m *Machine) RenderXFB() (*image.RGBA, error) {
	addr := m.vi.XFBAddr()
	if addr == 0 {
		return nil, fmt.Errorf("VI has no framebuffer address yet (the game has not set TFBL)")
	}
	w, h := m.xfbSize()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	stride := w * 2 // two bytes per pixel in YUY2

	for y := 0; y < h; y++ {
		row := addr + uint32(y*stride)
		for x := 0; x < w; x += 2 {
			o := row + uint32(x*2)
			if int(o)+3 >= len(m.RAM) {
				break
			}
			y0 := m.RAM[o]
			cb := m.RAM[o+1]
			y1 := m.RAM[o+2]
			cr := m.RAM[o+3]
			r0, g0, b0 := yuv2rgb(y0, cb, cr)
			r1, g1, b1 := yuv2rgb(y1, cb, cr)
			setPix(img, x, y, r0, g0, b0)
			setPix(img, x+1, y, r1, g1, b1)
		}
	}
	return img, nil
}

// xfbSize is the framebuffer's dimensions. Until the video mode is decoded from the DCR
// registers, the standard NTSC frame is the right default and the one Luigi's Mansion uses.
func (m *Machine) xfbSize() (int, int) {
	return 640, 480
}

func setPix(img *image.RGBA, x, y int, r, g, b uint8) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	i := img.PixOffset(x, y)
	img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, 255
}

// yuv2rgb converts one YUV 4:2:2 sample to RGB, with the standard BT.601 coefficients the
// GameCube's video encoder uses.
func yuv2rgb(y, cb, cr uint8) (uint8, uint8, uint8) {
	yf := float64(y)
	u := float64(cb) - 128
	v := float64(cr) - 128
	r := yf + 1.371*v
	g := yf - 0.336*u - 0.698*v
	b := yf + 1.732*u
	return clamp8(r), clamp8(g), clamp8(b)
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// IntrState reports the interrupt plumbing, for diagnosing a boot that waits on an
// interrupt that is not arriving: the pending causes, the mask, whether the CPU's external
// line and MSR[EE] agree that one could be taken, and whether any video display interrupt
// is armed to fire.
func (m *Machine) IntrState() string {
	viArmed := false
	for _, d := range m.vi.DI {
		if d&(1<<28) != 0 {
			viArmed = true
		}
	}
	dspState := "none"
	if m.dsp.Core != nil {
		dspState = fmt.Sprintf("PC=0x%04X blocked=%v halt=%v", m.dsp.Core.PC, m.dsp.CoreBlocked, m.dsp.Core.Halted)
	}
	return fmt.Sprintf("PI cause=0x%08X mask=0x%08X | CPU ExtInt=%v MSR[EE]=%v | VI armed=%v field=%d | PE reg0=0x%04X | TFBL=0x%08X XFB=0x%08X | DSP core %s toDSP=0x%08X fromDSP=0x%08X csr=0x%04X",
		m.pi.Cause, m.pi.Mask, m.CPU.ExtInt, m.CPU.MSR&(1<<15) != 0, viArmed, m.vi.Field, m.pe.Reg[0], m.vi.TFBL, m.vi.XFBAddr(),
		dspState, m.dsp.ToDSP, m.dsp.FromDSP, m.dsp.CSR)
}

// RegString is the integer register file, for reading a fault: when the machine halts on a
// bad access, the offending pointer is in one of these, and its neighbours usually say where
// it came from.
func (m *Machine) RegString() string {
	var b strings.Builder
	for i := 0; i < 32; i += 4 {
		fmt.Fprintf(&b, "  r%-2d %08X  r%-2d %08X  r%-2d %08X  r%-2d %08X\n",
			i, m.CPU.GPR[i], i+1, m.CPU.GPR[i+1], i+2, m.CPU.GPR[i+2], i+3, m.CPU.GPR[i+3])
	}
	fmt.Fprintf(&b, "  PC %08X  LR %08X  CTR %08X  SRR0 %08X  SRR1 %08X\n",
		m.CPU.PC, m.CPU.LR, m.CPU.CTR, m.CPU.SRR0, m.CPU.SRR1)
	return b.String()
}
