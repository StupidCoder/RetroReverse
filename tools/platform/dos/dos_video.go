package dos

// A minimal VGA video BIOS (INT 10h). Ultima Underworld detects the display
// adapter through the BIOS before it will set its video mode — and it gates the
// whole video-memory-arena initialisation on that mode-set, so with INT 10h
// stubbed out the arena stays uninitialised and the resource loader dies with
// "Could not read data". This model answers the detection calls as a VGA with
// 256 KiB, tracks the current mode in the BIOS Data Area, and accepts the
// palette/DAC programming. Pixel writes land in plain RAM at A000:0 (planar
// Mode-X semantics are not modelled — game logic, not pixels, is the goal).

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"retroreverse.com/tools/cpu/x86"
)

func (m *Machine) int10(c *x86.CPU) bool {
	ah := c.Reg8(x86.AH)
	bda := uint32(0x40) << 4

	// Log the first occurrence of each function for visibility.
	key := 0x1000 | uint16(ah)
	if m.video == nil {
		m.video = map[uint16]int{}
	}
	m.video[key]++
	if m.video[key] == 1 {
		m.logf("INT 10h AH=%02X AL=%02X BX=%04X at %04X:%04X",
			ah, c.Reg8(x86.AL), c.Reg16(x86.BX), c.Seg[x86.CS], c.IP)
	}

	switch ah {
	case 0x00: // set video mode (bit 7 of AL = don't clear screen)
		mode := c.Reg8(x86.AL) & 0x7F
		m.Mem[bda+0x49] = mode
		m.vgaInit13h(c.Reg8(x86.AL)&0x80 == 0) // chain-4 defaults; games unchain themselves
		cols := byte(80)
		if mode <= 1 || mode == 4 || mode == 5 || mode == 0x0D || mode == 0x13 {
			cols = 40
		}
		if mode == 0x13 {
			cols = 40
		}
		m.w16(bda+0x4A, uint16(cols))
		return true
	case 0x01, 0x02, 0x05, 0x06, 0x07, 0x09, 0x0A, 0x0B, 0x0C, 0x0E, 0x13:
		// cursor shape/pos, page select, scroll, write char/pixel/teletype/string
		return true
	case 0x03: // read cursor position
		c.SetReg16(x86.DX, 0)
		c.SetReg16(x86.CX, 0x0607)
		return true
	case 0x08: // read char/attribute at cursor
		c.SetReg16(x86.AX, 0x0720) // space, light grey on black
		return true
	case 0x0D: // read pixel
		c.SetReg8(x86.AL, 0)
		return true
	case 0x0F: // get current video mode
		c.SetReg8(x86.AL, m.Mem[bda+0x49])
		c.SetReg8(x86.AH, byte(m.r16(bda+0x4A)))
		c.SetReg8(x86.BH, 0) // active page
		return true
	case 0x10: // palette / DAC programming
		switch c.Reg8(x86.AL) {
		case 0x10: // set one DAC register: BX=index, DH/CH/CL = R/G/B
			i := int(c.Reg16(x86.BX)) % 256 * 3
			m.io.Pal[i] = c.Reg8(x86.DH)
			m.io.Pal[i+1] = c.Reg8(x86.CH)
			m.io.Pal[i+2] = c.Reg8(x86.CL)
		case 0x12: // set block of DAC registers from ES:DX (BX=start, CX=count)
			src := lin(c.Seg[x86.ES], c.Reg16(x86.DX))
			start := int(c.Reg16(x86.BX)) * 3
			n := int(c.Reg16(x86.CX)) * 3
			for i := 0; i < n && start+i < 768; i++ {
				m.io.Pal[start+i] = m.Mem[(src+uint32(i))&0xFFFFF]
			}
		case 0x15: // read one DAC register -> DH,CH,CL
			c.SetReg16(x86.CX, 0)
			c.SetReg8(x86.DH, 0)
		case 0x17: // read block of DAC registers into ES:DX
			n := int(c.Reg16(x86.CX)) * 3
			dst := lin(c.Seg[x86.ES], c.Reg16(x86.DX))
			for i := 0; i < n; i++ {
				m.Mem[(dst+uint32(i))&0xFFFFF] = 0
			}
		}
		return true
	case 0x11: // character generator
		if c.Reg8(x86.AL) == 0x30 { // get font information
			c.Seg[x86.ES] = 0xF000 // conventional ROM font location
			c.SetReg16(x86.BP, 0xFA6E)
			c.SetReg16(x86.CX, 16) // bytes per character
			c.SetReg8(x86.DL, 24)  // rows - 1
		}
		return true
	case 0x12: // alternate select
		if c.Reg8(x86.BL) == 0x10 { // get EGA info
			c.SetReg8(x86.BH, 0) // colour mode
			c.SetReg8(x86.BL, 3) // 256 KiB
			c.SetReg16(x86.CX, 0x0009)
		} else {
			c.SetReg8(x86.AL, 0x12) // function supported
		}
		return true
	case 0x1A: // get/set display combination code — the VGA detection call
		if c.Reg8(x86.AL) == 0x00 {
			c.SetReg8(x86.AL, 0x1A) // function supported => VGA BIOS present
			c.SetReg8(x86.BL, 0x08) // active display: VGA with analog colour
			c.SetReg8(x86.BH, 0x00) // no alternate display
		}
		return true
	case 0x1B: // functionality/state information — not provided
		c.SetReg8(x86.AL, 0)
		return true
	}
	return true
}

// Screenshot reconstructs the 320×200 display from the planar VGA memory and
// renders it through the captured DAC palette to a PNG file. In chain-4 (mode
// 13h) pixel i lives in plane i&3 at offset i>>2; unchained (Mode X) the
// display fetches four planes per byte address, starting at the CRTC start
// address with the CRTC offset register as the pitch — honouring both means
// the shot always shows what a monitor would.
func (m *Machine) Screenshot(path string) error {
	v := m.vga
	start := int(v.crtc[0x0C])<<8 | int(v.crtc[0x0D])
	pitch := int(v.crtc[0x13]) * 2
	if pitch == 0 {
		pitch = 80
	}
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			var pix byte
			if v.chained() {
				pi := y*320 + x
				pix = v.planes[pi&3][(pi>>2)&0xFFFF]
			} else {
				pix = v.planes[x&3][(start+y*pitch+x/4)&0xFFFF]
			}
			p := int(pix) * 3
			img.Set(x, y, color.RGBA{ // 6-bit DAC components → 8-bit
				m.io.Pal[p] << 2, m.io.Pal[p+1] << 2, m.io.Pal[p+2] << 2, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
