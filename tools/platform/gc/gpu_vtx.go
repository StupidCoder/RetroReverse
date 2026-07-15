package gc

// gpu_vtx.go decodes the vertex format the command processor must know before it can step
// over — or, in the next stage, fetch — the vertices a draw primitive carries. A draw opcode
// names a vertex-attribute-table index but not a size; the size is implied by two register
// sets the game programmed into the CP earlier in the stream: the vertex descriptor (which
// attributes are present, and whether each is inline or an index into an array) and the eight
// vertex-attribute tables (how each present attribute is formatted). Multiplying that
// per-vertex size by the primitive's vertex count is what tells the parser where the next
// command begins, so a size computed even one byte wrong desynchronises the whole FIFO —
// which is why the parser halts loudly the moment such a size leads it to an opcode it does
// not recognise.

// The two-bit descriptor value an attribute has in the vertex descriptor: absent, inline, or
// an 8- or 16-bit index into an array that lives elsewhere in memory.
const (
	descNone    = 0
	descDirect  = 1
	descIndex8  = 2
	descIndex16 = 3
)

// componentBytes is the size of one component of a position/normal/texture-coordinate
// attribute, by its three-bit format code.
func componentBytes(format uint32) int {
	switch format {
	case 0, 1: // u8, s8
		return 1
	case 2, 3: // u16, s16
		return 2
	case 4: // f32
		return 4
	}
	return 0
}

// colorBytes is the inline size of a direct colour attribute, by its three-bit format code.
func colorBytes(comp uint32) int {
	switch comp {
	case 0: // RGB565
		return 2
	case 1: // RGB888
		return 3
	case 2: // RGB888x
		return 4
	case 3: // RGBA4444
		return 2
	case 4: // RGBA6666
		return 3
	case 5: // RGBA8888
		return 4
	}
	return 0
}

// vertexSize returns how many bytes one vertex of the given attribute-table index occupies in
// the FIFO stream, computed from the current vertex descriptor (CP registers 0x50/0x60) and
// that table (0x70/0x80/0x90 + index).
func (g *gpu) vertexSize(vat int) int {
	lo := g.CPReg[0x50]
	hi := g.CPReg[0x60]
	g0 := g.CPReg[0x70+uint32(vat)]
	g1 := g.CPReg[0x80+uint32(vat)]
	g2 := g.CPReg[0x90+uint32(vat)]

	size := 0

	// The nine matrix-index attributes (the position matrix, then eight texture matrices) are
	// each a single inline byte when their descriptor bit is set.
	for b := uint32(0); b < 9; b++ {
		if lo&(1<<b) != 0 {
			size++
		}
	}

	// add accounts for an attribute that may be absent, inline, or an index. An index costs
	// one or two bytes whatever the array element's size; only a direct attribute carries its
	// data inline, and only then does the format-derived size apply.
	add := func(desc uint32, direct int) {
		switch desc {
		case descIndex8:
			size++
		case descIndex16:
			size += 2
		case descDirect:
			size += direct
		}
	}

	// Position: 2 or 3 components.
	posComps := 2
	if g0&1 != 0 {
		posComps = 3
	}
	add((lo>>9)&3, posComps*componentBytes((g0>>1)&7))

	// Normal: 3 components, or 9 when the normal/binormal/tangent set is enabled.
	nrmComps := 3
	if (g0>>9)&1 != 0 {
		nrmComps = 9
	}
	add((lo>>11)&3, nrmComps*componentBytes((g0>>10)&7))

	// Two colours, each with its own component format.
	add((lo>>13)&3, colorBytes((g0>>14)&7))
	add((lo>>15)&3, colorBytes((g0>>18)&7))

	// Eight texture coordinates: the first is formatted in group 0, the next three in group 1,
	// and the last four in group 2. Each has a 1/2-component element bit and a format code.
	texBytes := func(elem, format uint32) int {
		comps := 1
		if elem != 0 {
			comps = 2
		}
		return comps * componentBytes(format)
	}
	texDesc := func(k int) uint32 { return (hi >> (2 * uint32(k))) & 3 }

	add(texDesc(0), texBytes((g0>>21)&1, (g0>>22)&7))
	add(texDesc(1), texBytes((g1>>0)&1, (g1>>1)&7))
	add(texDesc(2), texBytes((g1>>9)&1, (g1>>10)&7))
	add(texDesc(3), texBytes((g1>>18)&1, (g1>>19)&7))
	add(texDesc(4), texBytes((g1>>27)&1, (g1>>28)&7))
	add(texDesc(5), texBytes((g2>>5)&1, (g2>>6)&7))
	add(texDesc(6), texBytes((g2>>14)&1, (g2>>15)&7))
	add(texDesc(7), texBytes((g2>>23)&1, (g2>>24)&7))

	return size
}
