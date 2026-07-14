package dsmachine

// VRAM. The DS does not have "video memory" at a fixed address — it has nine
// banks of it (A..I, 656 KiB in total) and a control register per bank that says
// which of the machine's several *address spaces* that bank currently is. The same
// 128 KiB of silicon is background memory now and texture memory after the next
// write to VRAMCNT_A.
//
// So there are two layers here, and conflating them is the classic DS emulator bug:
//
//	the banks       — where the bytes actually live, and what a DMA writes into
//	the spaces      — background, sprite, texture, texture-palette, extended-palette,
//	                  and the ARM7's window; what the display engines actually read
//
// A game writes into a bank through the LCDC window (0x06800000+, where each bank
// is visible at a fixed address regardless of mapping), then re-points the bank at
// a display space and the pixels appear. Model only the spaces and the upload has
// nowhere to land; model only the banks and nothing can read them back.
//
// Mapping is rebuilt eagerly on every VRAMCNT write, into a page table of 8 KiB
// pages per space — 8 KiB because that is the extended-palette slot size, the
// finest granularity any bank mapping uses.

const vramPage = 0x2000 // 8 KiB

// The banks, in VRAMCNT_A..I order. Sizes are the hardware's.
var bankSizes = [9]int{
	0x20000, 0x20000, 0x20000, 0x20000, // A B C D — 128 KiB
	0x10000,        // E — 64 KiB
	0x4000, 0x4000, // F G — 16 KiB
	0x8000, // H — 32 KiB
	0x4000, // I — 16 KiB
}

// Each bank's fixed address in the LCDC window, where it is directly visible to
// the CPU whatever else it is mapped as.
var lcdcBase = [9]uint32{
	0x06800000, 0x06820000, 0x06840000, 0x06860000,
	0x06880000, 0x06890000, 0x06894000, 0x06898000, 0x068A0000,
}

// The address spaces a bank can be mapped into.
const (
	spBGA     = iota // engine A background   — 0x06000000, 512 KiB
	spBGB            // engine B background   — 0x06200000, 128 KiB
	spOBJA           // engine A sprites      — 0x06400000, 256 KiB
	spOBJB           // engine B sprites      — 0x06600000, 128 KiB
	spTex            // 3D texture image      — 512 KiB (four 128 KiB slots)
	spTexPal         // 3D texture palette    — 128 KiB (six 16 KiB slots)
	spBGExtA         // engine A BG extended palette  — 32 KiB (four 8 KiB slots)
	spBGExtB         // engine B BG extended palette  — 32 KiB
	spOBJExtA        // engine A OBJ extended palette — 8 KiB
	spOBJExtB        // engine B OBJ extended palette — 8 KiB
	spARM7           // the ARM7's VRAM window — 256 KiB at 0x06000000 on its bus
	numSpaces
)

var spaceSize = [numSpaces]int{
	0x80000, 0x20000, 0x40000, 0x20000,
	0x80000, 0x20000,
	0x8000, 0x8000, 0x2000, 0x2000,
	0x40000,
}

// vram holds the nine banks and the page tables that say what each space currently
// sees. A page maps to zero, one, or (when a game maps two banks over each other)
// several bank slices: reads take the first, writes go to all of them, which is
// the only behaviour that keeps an overlapping upload from silently vanishing.
type vram struct {
	bank [9][]byte
	cnt  [9]uint8

	pages [numSpaces][][][]byte // space -> page -> the bank slices mapped there
}

func newVRAM() *vram {
	v := &vram{}
	for i := range v.bank {
		v.bank[i] = make([]byte, bankSizes[i])
	}
	for s := 0; s < numSpaces; s++ {
		v.pages[s] = make([][][]byte, spaceSize[s]/vramPage)
	}
	return v
}

// setCNT records a VRAMCNT write and rebuilds the mapping.
func (v *vram) setCNT(bank int, val uint8) {
	if v.cnt[bank] == val {
		return
	}
	v.cnt[bank] = val
	v.remap()
}

// remap rebuilds every page table from the nine control registers.
func (v *vram) remap() {
	for s := 0; s < numSpaces; s++ {
		for p := range v.pages[s] {
			v.pages[s][p] = v.pages[s][p][:0]
		}
	}
	for b := 0; b < 9; b++ {
		cnt := v.cnt[b]
		if cnt&0x80 == 0 {
			continue // bank disabled: it is only reachable through the LCDC window
		}
		mst := int(cnt & 7)
		ofs := int(cnt>>3) & 3
		space, off, ok := bankTarget(b, mst, ofs)
		if !ok {
			continue
		}
		v.attach(space, off, v.bank[b])
	}
}

// attach maps a bank's bytes into a space at a byte offset, page by page.
func (v *vram) attach(space, off int, data []byte) {
	for i := 0; i < len(data); i += vramPage {
		p := (off + i) / vramPage
		if p < 0 || p >= len(v.pages[space]) {
			continue // a mapping that runs off the end of its space simply does not appear
		}
		end := i + vramPage
		if end > len(data) {
			end = len(data)
		}
		v.pages[space][p] = append(v.pages[space][p], data[i:end])
	}
}

// bankTarget is the hardware's mapping table: for a bank, its MST (what kind of
// memory it becomes) and its OFS (where in that space it lands), the space and
// byte offset it maps to. The MST encoding is *per bank* — the same MST value
// means different things for bank A and bank H — which is why this is a table and
// not a formula.
func bankTarget(b, mst, ofs int) (space, off int, ok bool) {
	switch b {
	case 0, 1: // A, B — 128 KiB
		switch mst {
		case 1:
			return spBGA, 0x20000 * ofs, true
		case 2:
			return spOBJA, 0x20000 * (ofs & 1), true
		case 3:
			return spTex, 0x20000 * ofs, true
		}
	case 2: // C — 128 KiB
		switch mst {
		case 1:
			return spBGA, 0x20000 * ofs, true
		case 2:
			return spARM7, 0x20000 * (ofs & 1), true
		case 3:
			return spTex, 0x20000 * ofs, true
		case 4:
			return spBGB, 0, true
		}
	case 3: // D — 128 KiB
		switch mst {
		case 1:
			return spBGA, 0x20000 * ofs, true
		case 2:
			return spARM7, 0x20000 * (ofs & 1), true
		case 3:
			return spTex, 0x20000 * ofs, true
		case 4:
			return spOBJB, 0, true
		}
	case 4: // E — 64 KiB
		switch mst {
		case 1:
			return spBGA, 0, true
		case 2:
			return spOBJA, 0, true
		case 3:
			return spTexPal, 0, true
		case 4:
			return spBGExtA, 0, true // all four 8 KiB slots at once
		}
	case 5, 6: // F, G — 16 KiB
		// F and G address their space in a peculiar way: OFS bit 0 selects a 16 KiB
		// step and OFS bit 1 a 64 KiB step, so the two bits are not a plain index.
		step := 0x4000*(ofs&1) + 0x10000*(ofs>>1)
		switch mst {
		case 1:
			return spBGA, step, true
		case 2:
			return spOBJA, step, true
		case 3:
			return spTexPal, step, true
		case 4:
			return spBGExtA, 0x4000 * (ofs & 1), true // two 8 KiB slots
		case 5:
			return spOBJExtA, 0, true // only the first 8 KiB of the bank is used
		}
	case 7: // H — 32 KiB
		switch mst {
		case 1:
			return spBGB, 0, true
		case 2:
			return spBGExtB, 0, true
		}
	case 8: // I — 16 KiB
		switch mst {
		case 1:
			return spBGB, 0x8000, true
		case 2:
			return spOBJB, 0, true
		case 3:
			return spOBJExtB, 0, true
		}
	}
	return 0, 0, false // MST 0 is LCDC: no space mapping, only the direct window
}

// read8 reads a byte from a space. An unmapped page reads as zero — the hardware
// returns undefined data, but zero is the reading that makes an unmapped read
// *look* unmapped (a black tile, an empty palette) rather than plausible.
func (v *vram) read8(space int, off uint32) byte {
	p := off / vramPage
	if int(p) >= len(v.pages[space]) {
		return 0
	}
	refs := v.pages[space][p]
	if len(refs) == 0 {
		return 0
	}
	i := off % vramPage
	if int(i) >= len(refs[0]) {
		return 0
	}
	return refs[0][i]
}

func (v *vram) read16(space int, off uint32) uint16 {
	return uint16(v.read8(space, off)) | uint16(v.read8(space, off+1))<<8
}

// write8 writes to every bank mapped at that page (see the type comment).
func (v *vram) write8(space int, off uint32, b byte) {
	p := off / vramPage
	if int(p) >= len(v.pages[space]) {
		return
	}
	i := off % vramPage
	for _, r := range v.pages[space][p] {
		if int(i) < len(r) {
			r[i] = b
		}
	}
}

// lcdcSlot resolves an address in the LCDC window (0x06800000+) to the bank byte
// it names. This window is how the CPU and DMA actually *fill* VRAM, and it works
// whatever the bank is mapped as — a game uploads a texture here and then points
// the bank at the texture space.
func (v *vram) lcdcSlot(a uint32) ([]byte, uint32, bool) {
	for b := 0; b < 9; b++ {
		base := lcdcBase[b]
		if a >= base && a < base+uint32(bankSizes[b]) {
			return v.bank[b], a - base, true
		}
	}
	return nil, 0, false
}

// cpuSlot resolves a CPU-side VRAM address (0x06000000..0x068FFFFF) as the given
// core sees it. The engine spaces are readable and writable by the ARM9 at their
// own addresses too — that is how a game pokes a single background tile without
// going through the LCDC window.
func (v *vram) cpuAccess(arm9 bool, a uint32) (space int, off uint32, ok bool) {
	if !arm9 {
		// The ARM7 sees only the banks mapped to it (C and/or D with MST=2), in a
		// 256 KiB window at 0x06000000. It has no view of the display spaces at all.
		if a >= 0x06000000 && a < 0x07000000 {
			return spARM7, (a - 0x06000000) & 0x3FFFF, true
		}
		return 0, 0, false
	}
	switch {
	case a >= 0x06800000 && a < 0x068A4000:
		return 0, 0, false // the LCDC window is handled directly against the banks
	case a >= 0x06000000 && a < 0x06200000:
		return spBGA, (a - 0x06000000) & 0x7FFFF, true
	case a >= 0x06200000 && a < 0x06400000:
		return spBGB, (a - 0x06200000) & 0x1FFFF, true
	case a >= 0x06400000 && a < 0x06600000:
		return spOBJA, (a - 0x06400000) & 0x3FFFF, true
	case a >= 0x06600000 && a < 0x06800000:
		return spOBJB, (a - 0x06600000) & 0x1FFFF, true
	}
	return 0, 0, false
}
