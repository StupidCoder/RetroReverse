package threedo

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm60"
)

// io.go high-level-emulates the Portfolio kernel's asynchronous I/O and message
// passing — the layer the boot loader uses to read the disc. It is reimplemented
// from the SDK io.h/msgport.h ABI and the kernel folio's sendio.c semantics (a
// clean-room reimplementation, never a copy):
//
//   - A program opens a filesystem Device by name, creates a reply MsgPort and an
//     IOReq bound to {device, reply port}, fills an IOInfo (command, offset,
//     buffers) and submits it with SendIO (async) or DoIO (synchronous).
//   - On completion the kernel writes io_Actual/io_Error, sets IO_DONE in
//     io_Flags, and notifies the caller: if the IOReq has a reply port it signals
//     that port's owner with the port's signal, otherwise it sends SIGF_IODONE to
//     the owning task. The task wakes from WaitSignal / WaitPort and reads the
//     data the driver copied into the IOInfo's receive buffer.
//
// Because the oracle reads the disc instantly, every I/O completes inside the
// submit call — SendIO and DoIO behave identically here — and the completion
// signal is delivered before the call returns, so the task's following
// WaitSignal takes it immediately.
//
// IOReq struct field offsets (from io.h; ItemNode is 0x24 bytes, MinNode 8):
//
//	+0x34 io_Info (IOInfo, 0x20 bytes)   +0x54 io_Actual   +0x58 io_Flags
//	+0x5C io_Error                       +0x68 io_MsgItem
//
// IOInfo field offsets: +0x00 ioi_Command(u8) +0x0C ioi_Offset +0x10 ioi_Send
// {buffer,len} +0x18 ioi_Recv {buffer,len}.
const (
	ioInfoOff   = 0x34 // IOReq.io_Info
	ioActualOff = 0x54 // IOReq.io_Actual
	ioFlagsOff  = 0x58 // IOReq.io_Flags
	ioErrorOff  = 0x5C // IOReq.io_Error

	ioiCommand  = 0x00 // IOInfo.ioi_Command (u8)
	ioiOffset   = 0x0C // IOInfo.ioi_Offset
	ioiSendBuf  = 0x10 // IOInfo.ioi_Send.iob_Buffer
	ioiSendLen  = 0x14 // IOInfo.ioi_Send.iob_Len
	ioiRecvBuf  = 0x18 // IOInfo.ioi_Recv.iob_Buffer
	ioiRecvLen  = 0x1C // IOInfo.ioi_Recv.iob_Len
	ioInfoBytes = 0x20

	ioDone  = 0x01 // IO_DONE bit in io_Flags
	ioQuick = 0x02 // IO_QUICK: the request completed synchronously in the submit call

	// Standard device commands (CMD_*) and file device commands (FILECMD_*,
	// filesystem.h).
	cmdWrite           = 0
	cmdRead            = 1
	cmdStatus          = 2
	fileCmdReadDir     = 3
	fileCmdAllocBlocks = 6
	fileCmdSetEOF      = 7

	// Timer device: command 3 is "wait N fields" (the game's WaitVBL sends it with
	// the field count in ioi_Offset). The real driver blocks the caller until N
	// vertical-blank fields elapse; we complete instantly and instead advance the
	// virtual field clock by N, so the game's frame-pacing accumulator (which reads
	// the elapsed-field count each WaitVBL) settles at one update per field instead
	// of spinning while it waits for the clock to catch up.
	timerCmdWaitField = 3

	// SPORT (VRAM serial port) command: FLASHWRITE fills VRAM with a constant.
	sportFlashWrite = 6
)

// --- byte helpers (big-endian, over the full address space via the bus) -------

func (m *Machine) read32(a uint32) uint32 {
	return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
}

func (m *Machine) write32(a, v uint32) {
	m.Write(a, byte(v>>24))
	m.Write(a+1, byte(v>>16))
	m.Write(a+2, byte(v>>8))
	m.Write(a+3, byte(v))
}

// readCStr reads a NUL-terminated string (bounded) from memory.
func (m *Machine) readCStr(a uint32) string {
	if a == 0 {
		return ""
	}
	var b []byte
	for i := 0; i < 256; i++ {
		c := m.Read(a + uint32(i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

// tagArg walks a {tag, value} TagArg list and returns the value for want, or 0.
// The list ends at tag 0 (TAG_END); the tag word's high control bits are ignored.
func (m *Machine) tagArg(p, want uint32) uint32 {
	for i := 0; i < 64 && p != 0; i++ {
		tag := m.read32(p)
		val := m.read32(p + 4)
		if tag == 0 {
			break
		}
		if tag&0xFFFF == want {
			return val
		}
		p += 8
	}
	return 0
}

// tagString reads the C string a TagArg points at (e.g. TAG_ITEM_NAME).
func (m *Machine) tagString(p, want uint32) string {
	return m.readCStr(m.tagArg(p, want))
}

// --- I/O ---------------------------------------------------------------------

// serviceIO handles SendIO/DoIO(r0 = IOReq item, r1 = IOInfo*). It performs the
// requested transfer, records the completion fields on the IOReq, delivers the
// completion signal and returns the error in r0. With instant disc access the
// request always completes here, so the two calls behave identically.
func (m *Machine) serviceIO(c *arm60.CPU) {
	ioNum := int32(c.Reg(0))
	info := c.Reg(1)
	it := m.items[ioNum]

	cmd := uint32(m.Read(info + ioiCommand))
	offset := m.read32(info + ioiOffset)
	sendBuf := m.read32(info + ioiSendBuf)
	sendLen := m.read32(info + ioiSendLen)
	recvBuf := m.read32(info + ioiRecvBuf)
	recvLen := m.read32(info + ioiRecvLen)

	if m.SportDebug && cmd == sportFlashWrite {
		// FLASHWRITE IOInfo forensics: ioi_Recv carries the clear's dest+byte
		// count, ioi_Offset the fill colour (SetVRAMPages, game 0x39598).
		m.note(fmt.Sprintf("FLASHWRITE dest=0x%08X bytes=0x%X val=0x%08X (send=%08X,%08X)",
			recvBuf, recvLen, offset, sendBuf, sendLen))
	}

	// Copy the caller's IOInfo into the IOReq's io_Info, as the kernel does, so
	// code that inspects the request through its item sees the submitted command.
	if it != nil && it.addr != 0 {
		for i := uint32(0); i < ioInfoBytes; i++ {
			m.Write(it.addr+ioInfoOff+i, m.Read(info+i))
		}
	}

	actual, ioErr := m.performIO(it, cmd, offset, sendBuf, sendLen, recvBuf, recvLen)

	if it != nil && it.addr != 0 {
		m.write32(it.addr+ioActualOff, uint32(actual))
		// Our disc is instant, so every request finishes inside the submit call:
		// mark it DONE and QUICK. IO_QUICK tells the caller's WaitIO the request
		// completed synchronously with no reply message queued, so it returns the
		// stored error immediately instead of blocking on a reply port.
		m.write32(it.addr+ioFlagsOff, m.read32(it.addr+ioFlagsOff)|ioDone|ioQuick)
		m.write32(it.addr+ioErrorOff, uint32(ioErr))
	}
	m.completeIO(it)
	c.SetReg(0, uint32(ioErr))
}

// completeIO delivers an IOReq's completion notification: a signal to the reply
// port's owner (the message-port path) if it has one, else SIGF_IODONE to the
// task that owns the IOReq.
func (m *Machine) completeIO(it *item) {
	if it == nil {
		return
	}
	if it.replyPort != 0 {
		if p := m.items[it.replyPort]; p != nil {
			m.sendSignal(p.owner, p.signal)
			return
		}
	}
	m.sendSignal(it.owner, sigfIODONE)
}

// ioError returns the stored io_Error of an IOReq item (0 if unknown).
func (m *Machine) ioError(ioNum int32) uint32 {
	if it := m.items[ioNum]; it != nil && it.addr != 0 {
		return m.read32(it.addr + ioErrorOff)
	}
	return 0
}

// performIO carries out one device request and returns (bytes transferred,
// error). File-device requests (the device item was opened by OpenDiskFile)
// serve disc bytes or the NVRAM store; the timer's field-wait advances the
// virtual VBlank clock. Unmapped requests are logged and acknowledged as a
// zero-length success so the boot proceeds and the gap is visible.
func (m *Machine) performIO(it *item, cmd, offset, sendBuf, sendLen, buf, length uint32) (int32, int32) {
	dev := ""
	if it != nil && it.device != 0 {
		if d := m.items[it.device]; d != nil {
			dev = d.name
		}
	}
	m.note(fmt.Sprintf("IO dev=%q cmd=%d offset=0x%X buf=0x%08X len=0x%X", dev, cmd, offset, buf, length))
	if dev == "timer" && cmd == timerCmdWaitField {
		// A field wait: advance the virtual VBlank clock by the requested count so
		// the caller's timing loop sees the fields elapse (see timerCmdWaitField).
		m.advanceVBlank(offset)
		// On hardware this request parks the caller for a whole field — an eternity
		// in which every other task runs. Completing it instantly without yielding
		// would let a frame loop starve the rest of the program, so hand the CPU to
		// the next ready task; the waiter stays ready and resumes on its next turn.
		m.curTask().state = stReady
		m.needSchedule = true
		return 0, 0
	}
	// FLASHWRITE_CMD clears a VRAM span to a constant via the VRAM serial port —
	// how the game paints the frame's background each field (SetVRAMPages, game
	// 0x39598). The fill value is a 16-bit RGB555 replicated into both halfwords
	// of ioi_Offset; the destination is ioi_Recv.iob_Buffer and the byte count
	// ioi_Recv.iob_Len (= VRAM page size × page count). Each frame issues two: the
	// sky colour over the top span, the ground colour over the rest, so the
	// horizon lands exactly where the page counts split. These clears arrive on a
	// device that is not the named "SPORT" item, so they are keyed on the command
	// landing in VRAM, not the device name (file ALLOCBLOCKS is also command 6 but
	// never targets VRAM).
	if cmd == sportFlashWrite && buf >= vramBase && buf < vramBase+vramSize && length > 0 {
		m.flashClearRange(buf, length, uint16(offset))
		return 0, 0
	}
	if dev == "SPORT" {
		return 0, 0
	}
	if dev != "" && dev != "timer" {
		return m.fileDeviceIO(dev, cmd, offset, sendBuf, sendLen, buf, length)
	}
	// Non-file devices: acknowledge with no data until they are modelled.
	return 0, 0
}

// fileDeviceIO serves a request against an opened file: STATUS fills the
// FileStatus block (DeviceStatus + fs_ByteCount), READ copies block-addressed
// data, and the write-side commands mutate the NVRAM store (the disc is
// read-only). Block addressing follows the driver: ioi_Offset counts blocks of
// the device's block size — 2048 for disc files, 1 for NVRAM bytes.
func (m *Machine) fileDeviceIO(name string, cmd, offset, sendBuf, sendLen, buf, length uint32) (int32, int32) {
	data, blockSize, nvKey, ok := m.fileData(name)
	if !ok {
		return 0, -1
	}
	switch cmd {
	case cmdStatus:
		blocks := uint32(len(data))
		if blockSize > 1 {
			blocks = (uint32(len(data)) + blockSize - 1) / blockSize
		}
		for i, w := range []uint32{
			0,                // +0x00 identity/version/family/pad
			0x28,             // +0x04 ds_MaximumStatusSize
			blockSize,        // +0x08 ds_DeviceBlockSize
			blocks,           // +0x0C ds_DeviceBlockCount
			0, 0, 0, 0, 0,    // flags, usage, last error, media counter, reserved
			uint32(len(data)), // +0x24 fs_ByteCount (FileStatus)
		} {
			if uint32(i)*4 < length {
				m.write32(buf+uint32(i)*4, w)
			}
		}
		n := int32(0x28)
		if uint32(n) > length {
			n = int32(length)
		}
		return n, 0
	case cmdRead:
		byteOff := offset * blockSize
		if buf == 0 || int(byteOff) >= len(data) {
			return 0, 0
		}
		n := len(data) - int(byteOff)
		if uint32(n) > length {
			n = int(length)
		}
		for i := 0; i < n; i++ {
			m.Write(buf+uint32(i), data[int(byteOff)+i])
		}
		return int32(n), 0
	case cmdWrite:
		if nvKey == "" {
			return 0, -1 // the disc is read-only
		}
		byteOff := offset * blockSize
		stored := m.nvram[nvKey]
		if need := int(byteOff) + int(sendLen); need > len(stored) {
			grown := make([]byte, need)
			copy(grown, stored)
			stored = grown
		}
		for i := uint32(0); i < sendLen; i++ {
			stored[byteOff+i] = m.Read(sendBuf + i)
		}
		m.nvram[nvKey] = stored
		return int32(sendLen), 0
	case fileCmdAllocBlocks:
		if nvKey == "" {
			return 0, -1
		}
		if grow := int(offset) * int(blockSize); grow > 0 {
			m.nvram[nvKey] = append(m.nvram[nvKey], make([]byte, grow)...)
		}
		return 0, 0
	case fileCmdSetEOF:
		if nvKey == "" {
			return 0, -1
		}
		stored := m.nvram[nvKey]
		if int(offset) <= len(stored) {
			m.nvram[nvKey] = stored[:offset]
		} else {
			m.nvram[nvKey] = append(stored, make([]byte, int(offset)-len(stored))...)
		}
		return 0, 0
	default:
		return 0, 0
	}
}

// --- messages ----------------------------------------------------------------

// Message struct field offsets (msgport.h; the 0x24-byte ItemNode comes first).
const (
	msgReplyPort = 0x24 // msg_ReplyPort (Item)
	msgResult    = 0x28 // msg_Result
	msgDataPtr   = 0x2C // msg_DataPtr
	msgDataSize  = 0x30 // msg_DataSize
	msgMsgPort   = 0x34 // msg_MsgPort: the port the message is queued on
)

// serviceMsg handles the message-port SWIs: real FIFO queues per port, with the
// port's signal raised on its owner at queue time — the SendMsg/GetMsg/ReplyMsg
// handshake tasks use to talk to their worker threads.
func (m *Machine) serviceMsg(c *arm60.CPU, swi uint32) {
	switch swi {
	case swiPutMsg:
		// SendMsg(port, msg, dataptr, datasize): record the payload on the message,
		// queue it and wake the port's owner.
		port, msg := m.items[int32(c.Reg(0))], m.items[int32(c.Reg(1))]
		if port == nil || msg == nil {
			c.SetReg(0, ^uint32(0)) // BADITEM
			return
		}
		if msg.addr != 0 {
			m.write32(msg.addr+msgDataPtr, c.Reg(2))
			m.write32(msg.addr+msgDataSize, c.Reg(3))
			m.write32(msg.addr+msgMsgPort, uint32(port.num))
		}
		if port.name == "eventbroker" {
			m.eventBrokerRequest(port, msg, c.Reg(2))
			c.SetReg(0, 0)
			return
		}
		port.msgs = append(port.msgs, msg.num)
		m.sendSignal(port.owner, port.signal)
		c.SetReg(0, 0)
	case swiGetMsg:
		// GetMsg(port): pop the oldest queued message (0 = none).
		port := m.items[int32(c.Reg(0))]
		if port == nil || len(port.msgs) == 0 {
			c.SetReg(0, 0)
			return
		}
		num := port.msgs[0]
		port.msgs = port.msgs[1:]
		c.SetReg(0, uint32(num))
	case swiGetThisMsg:
		// GetThisMsg(msg): remove the message from whatever port holds it.
		msg := m.items[int32(c.Reg(0))]
		if msg == nil {
			c.SetReg(0, 0)
			return
		}
		for _, p := range m.items {
			for i, qn := range p.msgs {
				if qn == msg.num {
					p.msgs = append(p.msgs[:i], p.msgs[i+1:]...)
					c.SetReg(0, uint32(msg.num))
					return
				}
			}
		}
		c.SetReg(0, uint32(msg.num))
	case swiReplyMsg:
		// ReplyMsg(msg, result, dataptr, datasize): store the result and queue the
		// message back on its reply port, waking that port's owner.
		msg := m.items[int32(c.Reg(0))]
		if msg == nil {
			c.SetReg(0, ^uint32(0))
			return
		}
		if msg.addr != 0 {
			m.write32(msg.addr+msgDataPtr, c.Reg(2))
			m.write32(msg.addr+msgDataSize, c.Reg(3))
		}
		m.replyMsg(msg, c.Reg(1))
		c.SetReg(0, 0)
	}
}

// replyMsg stores a result on a message and queues it back on its reply port,
// waking that port's owner — the completion half of the SendMsg handshake.
func (m *Machine) replyMsg(msg *item, result uint32) {
	if msg == nil || msg.addr == 0 {
		return
	}
	m.write32(msg.addr+msgResult, result)
	rp := m.items[int32(m.read32(msg.addr+msgReplyPort))]
	if rp == nil {
		m.note(fmt.Sprintf("replyMsg: msg %d has no live reply port (field=0x%X)", msg.num, m.read32(msg.addr+msgReplyPort)))
		return
	}
	if rp.name == "eventbroker" {
		// An event message coming back from a listener (EB_EventReply): the
		// broker would recycle it; nothing to do in the HLE.
		return
	}
	rp.msgs = append(rp.msgs, msg.num)
	m.sendSignal(rp.owner, rp.signal)
}

// --- event broker -------------------------------------------------------------
//
// The system input broker (event.h): programs connect by sending a
// ConfigurationRequest message to the public "eventbroker" port; the broker
// replies to it and from then on delivers EB_EventRecord messages — an
// EventBrokerHeader followed by EventFrames — to the same reply port whenever
// a watched device (the control pad) changes. No broker task exists in the
// HLE, so the PutMsg intercept plays the broker: it acknowledges requests and
// records each connector's reply port as an event listener; SendPadEvent then
// pushes control-pad frames to every listener.

// Event-broker message flavors and the control-pad event payload (event.h).
const (
	ebConfigure   = 1
	ebEventRecord = 3

	eventNumButtonUpdate = 3 // EVENTNUM_ControlButtonUpdate: current button state

	// ControlPadEventData button bits (left-justified).
	PadDown       = 0x80000000
	PadUp         = 0x40000000
	PadRight      = 0x20000000
	PadLeft       = 0x10000000
	PadA          = 0x08000000
	PadB          = 0x04000000
	PadC          = 0x02000000
	PadStart      = 0x01000000
	PadX          = 0x00800000
	PadRightShift = 0x00400000
	PadLeftShift  = 0x00200000
)

// eventBrokerRequest services a message arriving at the broker port: it logs
// the flavor, remembers EB_Configure senders' reply ports as event listeners,
// and acknowledges the request so the sender's handshake completes.
func (m *Machine) eventBrokerRequest(port, msg *item, dataPtr uint32) {
	flavor := m.read32(dataPtr)
	if flavor == ebConfigure {
		listener := int32(m.read32(msg.addr + msgReplyPort))
		category := m.read32(dataPtr + 4)
		if m.items[listener] != nil {
			m.ebListeners = append(m.ebListeners, listener)
		}
		m.note(fmt.Sprintf("eventbroker: task #%d configured (category %d) -> listener port %d", msg.owner, category, listener))
	} else {
		m.note(fmt.Sprintf("eventbroker: msg %d flavor %d from task #%d acknowledged", msg.num, flavor, msg.owner))
	}
	m.replyMsg(msg, 0)
}

// SendPadEvent delivers a control-pad button state to every event-broker
// listener as an EB_EventRecord message: header, one ControlButtonUpdate
// EventFrame carrying the bits, and a terminating zero count.
func (m *Machine) SendPadEvent(buttons uint32) {
	for _, lp := range m.ebListeners {
		port := m.items[lp]
		if port == nil {
			continue
		}
		rec := m.dheap.alloc(0x40)
		if rec == 0 {
			return
		}
		m.writeWord(rec, ebEventRecord) // EventBrokerHeader.ebh_Flavor
		f := rec + 4
		m.writeWord(f+0x00, 0x20)     // ef_ByteCount: 28-byte frame + 4 data
		m.writeWord(f+0x04, 0)        // ef_SystemID
		m.writeWord(f+0x08, m.vblank) // ef_SystemTimeStamp
		m.writeWord(f+0x0C, 0)        // ef_Submitter
		m.Write(f+0x10, eventNumButtonUpdate)
		m.Write(f+0x11, 1) // ef_PodNumber: pad 1
		m.Write(f+0x12, 1) // ef_PodPosition
		m.Write(f+0x13, 1) // ef_GenericPosition: generic controller #1
		m.Write(f+0x14, 0) // ef_Trigger
		m.writeWord(f+0x18, 0)
		m.writeWord(f+0x1C, buttons) // ControlPadEventData.cped_ButtonBits
		m.writeWord(f+0x20, 0)       // terminating frame count
		msg := m.createItem(0x100|typeMsg, 0, 0)
		if msg.addr != 0 {
			m.writeWord(msg.addr+msgReplyPort, m.brokerPortNum())
			m.write32(msg.addr+msgDataPtr, rec)
			m.write32(msg.addr+msgDataSize, 0x28)
			m.write32(msg.addr+msgMsgPort, uint32(port.num))
		}
		port.msgs = append(port.msgs, msg.num)
		m.sendSignal(port.owner, port.signal)
	}
	m.note(fmt.Sprintf("pad event 0x%08X -> %d listener(s)", buttons, len(m.ebListeners)))
}

// brokerPortNum returns the public event-broker port's item number (0 if the
// program never looked it up).
func (m *Machine) brokerPortNum() uint32 {
	for _, it := range m.items {
		if it.name == "eventbroker" {
			return uint32(it.num)
		}
	}
	return 0
}

// --- kernel printf -----------------------------------------------------------

// kprintf implements the kernel's debug printf (SWI index 14). r0 is the format
// string; the first varargs are in r1..r3 and the rest on the stack. It supports
// the conversions the boot code uses (%d %u %x %X %c %s %%) so the game's own
// diagnostics land in the oracle's TTY. Returns 0.
func (m *Machine) kprintf() uint32 {
	c := m.CPU
	format := m.readCStr(c.Reg(0))
	argi := 0
	nextArg := func() uint32 {
		var v uint32
		switch argi {
		case 0:
			v = c.Reg(1)
		case 1:
			v = c.Reg(2)
		case 2:
			v = c.Reg(3)
		default:
			v = m.read32(c.Reg(13) + uint32(argi-3)*4)
		}
		argi++
		return v
	}

	var out []byte
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch != '%' || i+1 >= len(format) {
			out = append(out, ch)
			continue
		}
		i++
		// Skip flags/width/precision/length modifiers we don't format precisely.
		for i < len(format) && (format[i] == '-' || format[i] == '+' || format[i] == ' ' ||
			format[i] == '#' || format[i] == '0' || format[i] == '.' || format[i] == 'l' ||
			(format[i] >= '1' && format[i] <= '9')) {
			i++
		}
		if i >= len(format) {
			break
		}
		switch format[i] {
		case 'd', 'i':
			out = append(out, []byte(fmt.Sprintf("%d", int32(nextArg())))...)
		case 'u':
			out = append(out, []byte(fmt.Sprintf("%d", nextArg()))...)
		case 'x':
			out = append(out, []byte(fmt.Sprintf("%x", nextArg()))...)
		case 'X':
			out = append(out, []byte(fmt.Sprintf("%X", nextArg()))...)
		case 'p':
			out = append(out, []byte(fmt.Sprintf("0x%08X", nextArg()))...)
		case 'c':
			out = append(out, byte(nextArg()))
		case 's':
			out = append(out, []byte(m.readCStr(nextArg()))...)
		case '%':
			out = append(out, '%')
		default:
			out = append(out, '%', format[i])
		}
	}
	m.tty = append(m.tty, out...)
	return 0
}
