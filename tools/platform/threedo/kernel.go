package threedo

import (
	"fmt"
	"sort"

	"retroreverse.com/tools/cpu/arm60"
)

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

	// The kernel folio's I/O and message SWIs, by the same KERNELSWI-relative
	// index the boot code issues them at (index found from LaunchMe's disassembly
	// and the Portfolio kernel folio function table):
	//
	//	14 printf   15 GetThisMsg  16 PutMsg    18 ReplyMsg   19 GetMsg
	//	24 SendIO   25 AbortIO     34 CompleteIO 37 DoIO       41 WaitIO
	swiPrintf     = 0x1000E
	swiGetThisMsg = 0x1000F
	swiPutMsg     = 0x10010
	swiReplyMsg   = 0x10012
	swiGetMsg     = 0x10013
	swiSendIO     = 0x10018
	swiAbortIO    = 0x10019
	swiCompleteIO = 0x10022
	swiDoIO       = 0x10025
	swiWaitIO     = 0x10029

	// Kernel item subtypes we care about.
	typeSemaphore = 0x07
	typeMsg       = 0x09 // MESSAGENODE (created by CreateMsg right before use)
	typeMsgPort   = 0x0A
	typeIOReq     = 0x0E
	typeDevice    = 0x0F

	// Item-creation tag numbers (TAG_ITEM_LAST = 9, so item-specific tags start at
	// 0xA). MsgPort: CREATEPORT_TAG_SIGNAL. IOReq: CREATEIOREQ_TAG_REPLYPORT then
	// _TAG_DEVICE. TAG_ITEM_NAME (a base item tag) carries a found item's name.
	tagItemName       = 1
	tagPortSignal     = 0x0A
	tagIOReqReplyPort = 0x0A
	tagIOReqDevice    = 0x0B
	tagMsgReplyPort   = 0x0A // CREATEMSG_TAG_REPLY_PORT (TAG_ITEM_LAST+1)

	sigfIODONE = 0x08 // SIGF_IODONE: the default I/O-completion signal
)

// item is a kernel Item tracked by the HLE.
type item struct {
	num  int32
	typ  uint32 // full item type (folio<<8 | subtype)
	addr uint32 // an allocated in-RAM struct for the item, if any
	tags uint32 // the tag-args pointer the item was created/found with
	name string // item name (devices found by name, named ports, ...)

	owner  int32  // task that created/owns this item
	signal uint32 // message port: the signal raised on its owner when a msg arrives
	msgs   []int32 // message port: queued Message items, FIFO

	// IOReq fields (typeIOReq): the device it drives and its reply port (0 = the
	// completion is delivered as SIGF_IODONE to the owner task instead).
	device    int32
	replyPort int32
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
		m.initItemFromTags(it)
		c.SetReg(0, uint32(it.num))
	case swiFindItem:
		c.SetReg(0, uint32(m.findItem(c.Reg(0), c.Reg(1)).num))
	case swiPrintf:
		c.SetReg(0, m.kprintf())
	case swiSendIO, swiDoIO:
		m.serviceIO(c)
	case swiWaitIO:
		// The I/O already completed synchronously in serviceIO, so there is
		// nothing to wait for; return its stored error (0 = success).
		c.SetReg(0, m.ioError(int32(c.Reg(0))))
	case swiAbortIO, swiCompleteIO:
		c.SetReg(0, 0)
	case swiPutMsg, swiReplyMsg, swiGetMsg, swiGetThisMsg:
		m.serviceMsg(c, swi)
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
		c.SetReg(0, m.allocSignalFor(m.curTask().num))
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
	it := &item{num: m.nextItem, typ: typ, tags: tags, owner: m.curTask().num}
	// Give the item a backing struct so field writes/reads have somewhere to go.
	structSize := uint32(0x100)
	if size > structSize {
		structSize = size
	}
	it.addr = m.dheap.alloc(structSize)
	if it.addr != 0 {
		// Fill the ItemNode header fields programs read back off the struct:
		// n_SubsysType/n_Type at +8/+9 and n_Item at +0x18 (nodes.h layout).
		m.Write(it.addr+8, byte(typ>>8))
		m.Write(it.addr+9, byte(typ))
		m.writeWord(it.addr+0x18, uint32(it.num))
	}
	m.items[it.num] = it
	return it
}

// initItemFromTags fills in the type-specific fields (message-port signal, IOReq
// device/reply-port) of a freshly created item from its TagArg list.
func (m *Machine) initItemFromTags(it *item) {
	switch it.typ & 0xFF {
	case typeMsgPort:
		it.signal = m.tagArg(it.tags, tagPortSignal)
		if it.signal == 0 { // no explicit signal: allocate one for the owner
			it.signal = m.allocSignalFor(it.owner)
		}
		if it.addr != 0 {
			// mp_Signal sits right after the ItemNode (+0x24); programs read it
			// off the struct (GetMsgPortSignal) to build their WaitSignal masks.
			m.writeWord(it.addr+0x24, it.signal)
		}
	case typeIOReq:
		it.replyPort = int32(m.tagArg(it.tags, tagIOReqReplyPort))
		it.device = int32(m.tagArg(it.tags, tagIOReqDevice))
	case typeMsg:
		// CreateMsg(..., CREATEMSG_TAG_REPLY_PORT): reply routing lives on the
		// message struct where ReplyMsg reads it back.
		if it.addr != 0 {
			m.writeWord(it.addr+msgReplyPort, m.tagArg(it.tags, tagMsgReplyPort))
		}
	}
}

// allocSignalFor hands out a free signal bit to the given task. User signals
// start above the reserved SIGF_* bits (0x100 and up), matching AllocSignal.
func (m *Machine) allocSignalFor(taskNum int32) uint32 {
	t := m.taskByNum(taskNum)
	if t == nil {
		t = m.curTask()
	}
	bit := uint32(0x100)
	for bit != 0 && t.allocSigs&bit != 0 {
		bit <<= 1
	}
	t.allocSigs |= bit
	return bit
}

// ItemsSummary lists every live kernel item for the oracle's diagnostics.
func (m *Machine) ItemsSummary() []string {
	var nums []int32
	for n := range m.items {
		nums = append(nums, n)
	}
	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
	var out []string
	for _, n := range nums {
		it := m.items[n]
		s := fmt.Sprintf("item %d type=0x%X owner=#%d", it.num, it.typ, it.owner)
		if it.name != "" {
			s += fmt.Sprintf(" name=%q", it.name)
		}
		if it.signal != 0 {
			s += fmt.Sprintf(" signal=0x%X", it.signal)
		}
		if len(it.msgs) > 0 {
			s += fmt.Sprintf(" queued=%v", it.msgs)
		}
		if it.device != 0 {
			s += fmt.Sprintf(" device=%d", it.device)
		}
		if it.replyPort != 0 {
			s += fmt.Sprintf(" replyPort=%d", it.replyPort)
		}
		if it.typ&0xFF == typeMsg && it.addr != 0 {
			s += fmt.Sprintf(" msgReplyPort=%d result=0x%X", m.read32(it.addr+msgReplyPort), m.read32(it.addr+msgResult))
		}
		out = append(out, s)
	}
	return out
}

// findItem returns a per-(type,name) item, creating it on first request so
// repeated finds of the same subsystem object resolve to the same Item number.
func (m *Machine) findItem(typ, tags uint32) *item {
	name := m.tagString(tags, tagItemName)
	key := typ
	if it, ok := m.itemByType[key]; ok && it.name == name {
		return it
	}
	it := m.createItem(typ, tags, 0)
	it.name = name
	m.itemByType[key] = it
	return it
}
