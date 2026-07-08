package dos

// A faithful DOS conventional-memory manager built on the Memory Control Block
// (MCB) chain. DOS memory is a linked list of blocks; each block is preceded by
// a 16-byte MCB (at segment block-1):
//
//	+0    mark   'M' (0x4D) = a block follows, 'Z' (0x5A) = last block
//	+1..2 owner  PSP segment of the owner (0 = free)
//	+3..4 size   block size in paragraphs (not counting the MCB itself)
//	+8..15 name  program name (unused here)
//
// UW manages memory by repeatedly resizing its own program block (INT 21h/4Ah)
// to grab the free space above it, then sub-allocating from there; getting the
// block layout and the resize/free/allocate semantics exactly right is what its
// internal memory manager depends on. The chain runs from firstMCB up to memTop
// (the top of conventional memory); the EMS page frame and the fake EMM driver
// live above it in the upper-memory area, so they are outside the chain.

const (
	mcbNormal = 0x4D // 'M'
	mcbLast   = 0x5A // 'Z'
	mcbFree   = 0    // owner 0 => free block
)

// setupMCB builds the initial chain: the environment block (owned by the
// program) followed by the program block, which — as for a program that asked
// for the maximum allocation — owns all remaining conventional memory and is the
// last block in the chain.
func (m *Machine) setupMCB() {
	envMCB := m.envSeg - 1
	pspMCB := m.pspSeg - 1
	m.writeMCB(envMCB, mcbNormal, m.pspSeg, pspMCB-m.envSeg)
	m.writeMCB(pspMCB, mcbLast, m.pspSeg, m.memTop-m.pspSeg)
	m.firstMCB = envMCB
}

func (m *Machine) writeMCB(mcbSeg uint16, mark byte, owner, size uint16) {
	a := uint32(mcbSeg) << 4
	m.Mem[a&0xFFFFF] = mark
	m.w16(a+1, owner)
	m.w16(a+3, size)
}

func (m *Machine) readMCB(mcbSeg uint16) (mark byte, owner, size uint16) {
	a := uint32(mcbSeg) << 4
	return m.Mem[a&0xFFFFF], m.r16(a + 1), m.r16(a + 3)
}

// allocBlock services INT 21h/48h: first-fit a free block big enough for want
// paragraphs, splitting off the remainder. Returns the block segment, or (on
// failure) the largest free block found.
func (m *Machine) allocBlock(want uint16) (seg, largest uint16, ok bool) {
	mcb := m.firstMCB
	for i := 0; i < 4096; i++ {
		mark, owner, size := m.readMCB(mcb)
		if owner == mcbFree {
			if size >= want {
				m.carve(mcb, mark, size, want, m.pspSeg)
				return mcb + 1, 0, true
			}
			if size > largest {
				largest = size
			}
		}
		if mark == mcbLast {
			break
		}
		mcb += 1 + size
	}
	return 0, largest, false
}

// carve turns the free block at mcb (mark/size) into an allocated block of want
// paragraphs owned by owner, appending a free remainder block if room remains.
func (m *Machine) carve(mcb uint16, mark byte, size, want, owner uint16) {
	if size >= want+1 { // room for a remainder MCB (1 para) + at least 0 data
		rem := mcb + 1 + want
		m.writeMCB(rem, mark, mcbFree, size-want-1) // remainder inherits last-ness
		m.writeMCB(mcb, mcbNormal, owner, want)
	} else {
		m.writeMCB(mcb, mark, owner, size) // whole block (can't split)
	}
}

// freeBlock services INT 21h/49h: mark the block at blockSeg free, then coalesce
// adjacent free blocks. Returns false for an invalid block.
func (m *Machine) freeBlock(blockSeg uint16) bool {
	if blockSeg == 0 || blockSeg <= m.firstMCB {
		return false
	}
	mcb := blockSeg - 1
	mark, _, size := m.readMCB(mcb)
	if mark != mcbNormal && mark != mcbLast {
		return false
	}
	m.writeMCB(mcb, mark, mcbFree, size)
	m.coalesce()
	return true
}

// coalesce merges each run of adjacent free blocks into one.
func (m *Machine) coalesce() {
	mcb := m.firstMCB
	for i := 0; i < 4096; i++ {
		mark, owner, size := m.readMCB(mcb)
		if mark == mcbLast {
			break
		}
		next := mcb + 1 + size
		nmark, nowner, nsize := m.readMCB(next)
		if owner == mcbFree && nowner == mcbFree {
			m.writeMCB(mcb, nmark, mcbFree, size+nsize+1) // absorb next; inherit its mark
			continue                                      // re-examine with the new neighbour
		}
		mcb = next
	}
}

// resizeBlock services INT 21h/4Ah: shrink (freeing a remainder) or grow the
// block at blockSeg to want paragraphs by absorbing following free blocks.
// Returns the achieved/maximum size and whether it succeeded.
func (m *Machine) resizeBlock(blockSeg, want uint16) (uint16, bool) {
	if blockSeg == 0 || blockSeg <= m.firstMCB {
		return 0, false
	}
	mcb := blockSeg - 1
	mark, owner, size := m.readMCB(mcb)

	if want <= size { // shrink (or no change)
		if want < size {
			rem := mcb + 1 + want
			m.writeMCB(rem, mark, mcbFree, size-want-1)
			m.writeMCB(mcb, mcbNormal, owner, want)
			m.coalesce()
		}
		return want, true
	}

	// Grow: walk following free blocks, accumulating available space.
	avail := size
	lastMark := mark
	next := mcb + 1 + size
	for i := 0; i < 4096 && lastMark != mcbLast; i++ {
		nmark, nowner, nsize := m.readMCB(next)
		if nowner != mcbFree {
			break
		}
		avail += nsize + 1
		lastMark = nmark
		next += 1 + nsize
	}
	if want > avail {
		return avail, false
	}
	if avail > want { // grew but with room to spare: leave a free remainder
		rem := mcb + 1 + want
		m.writeMCB(rem, lastMark, mcbFree, avail-want-1)
		m.writeMCB(mcb, mcbNormal, owner, want)
	} else { // consumed everything up to (and including) the last absorbed block
		m.writeMCB(mcb, lastMark, owner, want)
	}
	return want, true
}
