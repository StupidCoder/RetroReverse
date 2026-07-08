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

	ioDone = 0x01 // IO_DONE bit in io_Flags

	// Standard device commands (CMD_*) and file device commands (FILECMD_*).
	cmdWrite       = 0
	cmdRead        = 1
	cmdStatus      = 2
	fileCmdReadDir = 3
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
	recvBuf := m.read32(info + ioiRecvBuf)
	recvLen := m.read32(info + ioiRecvLen)

	// Copy the caller's IOInfo into the IOReq's io_Info, as the kernel does, so
	// code that inspects the request through its item sees the submitted command.
	if it != nil && it.addr != 0 {
		for i := uint32(0); i < ioInfoBytes; i++ {
			m.Write(it.addr+ioInfoOff+i, m.Read(info+i))
		}
	}

	actual, ioErr := m.performIO(it, cmd, offset, recvBuf, recvLen)

	if it != nil && it.addr != 0 {
		m.write32(it.addr+ioActualOff, uint32(actual))
		m.write32(it.addr+ioFlagsOff, m.read32(it.addr+ioFlagsOff)|ioDone)
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
// error). It reads file/directory data straight off the mounted disc image. When
// the request cannot yet be mapped to disc data it is logged and reported as a
// zero-length success so the boot proceeds and the gap is visible.
func (m *Machine) performIO(it *item, cmd, offset, buf, length uint32) (int32, int32) {
	dev := ""
	if it != nil && it.device != 0 {
		if d := m.items[it.device]; d != nil {
			dev = d.name
		}
	}
	m.note(fmt.Sprintf("IO dev=%q cmd=%d offset=0x%X buf=0x%08X len=0x%X", dev, cmd, offset, buf, length))
	switch cmd {
	case cmdRead:
		return m.readDevice(it, offset, buf, length)
	default:
		// STATUS/READDIR/WRITE and unmapped commands: acknowledge with no data.
		return 0, 0
	}
}

// readDevice copies length bytes at the given offset from the disc file backing
// the IOReq's device into buf. It relies on the device name being a disc path;
// mapping opened-file items to disc files is refined as the boot reveals them.
func (m *Machine) readDevice(it *item, offset, buf, length uint32) (int32, int32) {
	if m.vol == nil || it == nil || buf == 0 {
		return 0, 0
	}
	dev := m.items[it.device]
	if dev == nil || dev.name == "" {
		return 0, 0
	}
	data, err := m.vol.ReadFile(dev.name)
	if err != nil {
		return 0, 0
	}
	if int(offset) >= len(data) {
		return 0, 0
	}
	n := len(data) - int(offset)
	if uint32(n) > length {
		n = int(length)
	}
	for i := 0; i < n; i++ {
		m.Write(buf+uint32(i), data[int(offset)+i])
	}
	return int32(n), 0
}

// --- messages ----------------------------------------------------------------

// serviceMsg handles the message-port SWIs the boot uses alongside I/O. The
// oracle delivers I/O completions as signals, so a full message queue is not
// needed yet; these keep the call sites progressing and return benign results.
func (m *Machine) serviceMsg(c *arm60.CPU, swi uint32) {
	switch swi {
	case swiPutMsg:
		// PutMsg(port, msg): signal the port's owner so a WaitPort wakes.
		if p := m.items[int32(c.Reg(0))]; p != nil {
			m.sendSignal(p.owner, p.signal)
		}
		c.SetReg(0, 0)
	case swiGetMsg, swiGetThisMsg:
		// No queued message: report "none".
		c.SetReg(0, 0)
	case swiReplyMsg:
		c.SetReg(0, 0)
	}
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
