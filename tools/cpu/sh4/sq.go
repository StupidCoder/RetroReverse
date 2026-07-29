package sh4

// sq.go is the store queues: two 32-byte buffers the CPU fills with ordinary
// stores to 0xE0000000-0xE3FFFFFF and flushes with pref, 32 bytes at a time,
// to the external address QACR0/1 select the high bits of. Katana submits
// Holly register blocks and tile-accelerator data this way, so the mechanics
// are modeled rather than gap-logged: a machine that dropped SQ writes would
// hide a large fraction of what the boot actually stores.

// sqWrite lands a store in a queue. Address bit 5 picks the queue, bits 2-4
// the word; sub-word stores merge into the word the way they would into any
// memory.
func (c *CPU) sqWrite(addr uint32, size, v uint32) {
	w := &c.SQ[(addr>>5)&1][(addr>>2)&7]
	switch size {
	case 1:
		sh := 8 * (addr & 3)
		*w = *w&^(0xFF<<sh) | (v&0xFF)<<sh
	case 2:
		sh := 8 * (addr & 2)
		*w = *w&^(0xFFFF<<sh) | (v&0xFFFF)<<sh
	default:
		*w = v
	}
}

// sqRead32 reads a queue back — legal, and occasionally used.
func (c *CPU) sqRead32(addr uint32) uint32 {
	return c.SQ[(addr>>5)&1][(addr>>2)&7]
}

// sqFlush is pref inside the SQ window: the queue bursts to the external
// address built from QACRn bits 4-2 (physical address bits 28-26) and the
// virtual address bits 25-5.
func (c *CPU) sqFlush(addr uint32) {
	q := (addr >> 5) & 1
	qacr := c.QACR0
	if q == 1 {
		qacr = c.QACR1
	}
	ext := (qacr&0x1C)<<24 | addr&0x03FFFFE0
	for i := uint32(0); i < 8; i++ {
		c.bus.Write32(ext+4*i, c.SQ[q][i])
	}
}
