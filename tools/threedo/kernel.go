package threedo

import "retroreverse.com/tools/arm60"

// kernel.go high-level-emulates the Portfolio kernel folio's SWI calls — the
// item system, semaphores and signals the game uses during boot. The SWI numbers
// and item types are from the 3DO SDK (KERNELSWI = 0x10000): CreateItem finds and
// tracks kernel objects by an "Item" number the game passes to later calls.
//
// The kernel folio SWI functions, by index off KERNELSWI:
//
//	+0 CreateSizedItem  +1 WaitSignal   +2 SendSignal   +3 DeleteItem
//	+4 FindItem         +5 OpenItem     +6 UnlockItem    +7 LockItem
//	+8 CloseItem        +10 SetItemPri
//
// Item types (folio 1 << 8 | subtype): 0x104 Folio, 0x107 Semaphore,
// 0x10A MsgPort, 0x10E IOReq, 0x10F Device.
const (
	swiCreateSizedItem = 0x10000
	swiWaitSignal      = 0x10001
	swiSendSignal      = 0x10002
	swiDeleteItem      = 0x10003
	swiFindItem        = 0x10004
	swiOpenItem        = 0x10005
	swiUnlockItem      = 0x10006
	swiLockItem        = 0x10007
	swiCloseItem       = 0x10008
	swiSetItemPri      = 0x1000A
	swiAllocSignal     = 0x10015
	swiFreeSignal      = 0x10016

	// Kernel item subtypes we care about.
	typeSemaphore = 0x07
	typeMsgPort   = 0x0A
	typeIOReq     = 0x0E
	typeDevice    = 0x0F
)

// item is a kernel Item tracked by the HLE.
type item struct {
	num  int32
	typ  uint32 // full item type (folio<<8 | subtype)
	addr uint32 // an allocated in-RAM struct for the item, if any
	tags uint32 // the tag-args pointer the item was created/found with
}

// kernelSWI services a Portfolio kernel folio SWI. It sets r0 to the result and
// returns true if handled; unknown SWIs return false to be logged as stubs.
func (m *Machine) kernelSWI(c *arm60.CPU, swi uint32) bool {
	switch swi {
	case swiCreateSizedItem:
		if c.Reg(0) == 0x105 { // TASKNODE — spawn a real cooperative task
			c.SetReg(0, uint32(m.spawnTask(c.Reg(1))))
			break
		}
		it := m.createItem(c.Reg(0), c.Reg(1), c.Reg(2))
		c.SetReg(0, uint32(it.num))
	case swiFindItem:
		c.SetReg(0, uint32(m.findItem(c.Reg(0), c.Reg(1)).num))
	case swiOpenItem:
		// OpenItem(item, args) — return the item, opened.
		c.SetReg(0, c.Reg(0))
	case swiLockItem:
		// LockItem(sema, waitflag) — in a single context the lock is always free.
		c.SetReg(0, 1)
	case swiUnlockItem, swiCloseItem, swiDeleteItem, swiSetItemPri:
		c.SetReg(0, 0) // success
	case swiWaitSignal:
		// WaitSignal(mask): if any awaited signal is already pending, take it and
		// return; otherwise block this task and ask the run loop to switch. When
		// resumed (SendSignal woke us) we continue past here with r0 = the mask.
		t := m.curTask()
		mask := c.Reg(0)
		if got := t.sig & mask; got != 0 || mask == 0 {
			t.sig &^= got
			c.SetReg(0, got)
		} else {
			c.SetReg(0, mask) // resumed value
			t.wait = mask
			t.state = stWaiting
			m.needSchedule = true
		}
	case swiSendSignal:
		// SendSignal(task, sigs)
		m.sendSignal(int32(c.Reg(0)), c.Reg(1))
		c.SetReg(0, 0)
	case swiAllocSignal:
		// AllocSignal(mask) allocates a free signal bit for the current task. User
		// signals start above the reserved SIGF_* bits (0x100 and up).
		t := m.curTask()
		bit := uint32(0x100)
		for bit != 0 && t.allocSigs&bit != 0 {
			bit <<= 1
		}
		t.allocSigs |= bit
		c.SetReg(0, bit)
	case swiFreeSignal:
		m.curTask().allocSigs &^= c.Reg(0)
		c.SetReg(0, 0)
	default:
		return false
	}
	return true
}

// createItem makes a new tracked item of the given type and returns it. IOReq and
// similar items get a small zeroed struct in DRAM so the game can fill their
// fields; the item number is what the game holds onto.
func (m *Machine) createItem(typ, tags, size uint32) *item {
	m.nextItem++
	it := &item{num: m.nextItem, typ: typ, tags: tags}
	// Give the item a backing struct so field writes/reads have somewhere to go.
	structSize := uint32(0x100)
	if size > structSize {
		structSize = size
	}
	it.addr = m.dheap.alloc(structSize)
	m.items[it.num] = it
	return it
}

// findItem returns a per-(type) singleton item, creating it on first request so
// repeated finds of the same subsystem object resolve to the same Item number.
func (m *Machine) findItem(typ, tags uint32) *item {
	if it, ok := m.itemByType[typ]; ok {
		return it
	}
	it := m.createItem(typ, tags, 0)
	m.itemByType[typ] = it
	return it
}
