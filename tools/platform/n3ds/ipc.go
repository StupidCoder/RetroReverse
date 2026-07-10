package n3ds

import "fmt"

// ipc.go high-level-emulates the Horizon IPC layer: the service-manager port
// ("srv:") and the individual OS services a title talks to over SendSyncRequest.
// A request's command buffer lives in the calling thread's TLS at +0x80: a
// header word (commandID<<16 | normalParams<<6 | translateParams), then the
// parameters, then translate descriptors (handles, mapped buffers). The service
// writes its reply back into the same buffer: the response header, a result
// code, and return values.
//
// The surface is grown lazily, exactly as the supervisor calls were: a request
// this layer does not recognise halts with the port name and command ID, so a
// run always reports the precise next facility the game needs. The aim of this
// stage is to carry a title through the srv:/APT/GSP init far enough to reach
// the point where it *submits its first GPU frame* — the honest first-frame
// milestone for a machine that does not (yet) emulate the PICA200 GPU.

// ipcHeader unpacks a command-buffer header word.
type ipcHeader struct {
	Command    uint16
	Normal     int // normal parameter words
	Translate  int // translate parameter words
}

func parseIPCHeader(w uint32) ipcHeader {
	return ipcHeader{
		Command:   uint16(w >> 16),
		Normal:    int(w >> 6 & 0x3F),
		Translate: int(w & 0x3F),
	}
}

const tlsCmdBuf = tlsBase + 0x80

// arg reads normal parameter i (1-based: arg 1 is the word after the header).
func (m *Machine) ipcArg(i int) uint32 { return m.ReadWord(tlsCmdBuf + uint32(i)*4) }

// reply writes a standard response: the response header (same command ID, the
// given result-word count, no translate), result code 0, then the values.
func (m *Machine) ipcReply(cmd uint16, values ...uint32) {
	m.WriteWord(tlsCmdBuf, uint32(cmd)<<16|uint32(len(values)+1)<<6)
	m.WriteWord(tlsCmdBuf+4, resultSuccess)
	for i, v := range values {
		m.WriteWord(tlsCmdBuf+8+uint32(i)*4, v)
	}
}

// handleIPC dispatches a SendSyncRequest. It replaces the halting stub: the
// handle names a port ("srv:") or a service the game acquired from srv:.
func (m *Machine) handleIPC(handle uint32) bool {
	hdr := parseIPCHeader(m.ReadWord(tlsCmdBuf))
	name := m.ports[handle]
	if svc, ok := m.services[handle]; ok {
		name = svc
	}
	m.ipcLog = append(m.ipcLog, ipcCall{service: name, command: hdr.Command})
	if m.Verbose {
		fmt.Printf("  IPC handle=0x%08X %-14s cmd 0x%04X (%d normal, %d translate)\n", handle, name, hdr.Command, hdr.Normal, hdr.Translate)
	}

	switch name {
	case "srv:", "srv:pm":
		return m.ipcSrv(hdr)
	default:
		return m.ipcService(name, hdr)
	}
}

// ipcSrv services the service-manager port.
func (m *Machine) ipcSrv(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0001: // RegisterClient
		m.ipcReply(hdr.Command)
		return true
	case 0x0002: // EnableNotification → returns a semaphore handle (translate)
		h := m.newHandle("notification-semaphore", true)
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|1<<6|2)
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0) // translate header: move-handle
		m.WriteWord(tlsCmdBuf+12, h)
		return true
	case 0x0005: // GetServiceHandle(name[8], nameLen, flags)
		name := m.readServiceName()
		h := m.newHandle("service:"+name, false)
		m.services[h] = name
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|1<<6|2)
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0)
		m.WriteWord(tlsCmdBuf+12, h)
		if m.Verbose {
			fmt.Printf("    GetServiceHandle %q -> 0x%08X\n", name, h)
		}
		return true
	case 0x0009: // Subscribe
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("srv: command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// knownService reports whether name's family is one the HLE recognises — used to
// disambiguate the service-name byte order (see readServiceName).
func knownService(name string) bool {
	switch serviceBase(name) {
	case "APT", "gsp", "hid", "cfg", "fs", "ndm", "ptm", "ac", "frd", "cecd",
		"boss", "nim", "mic", "csnd", "dsp", "y2r", "am", "ns", "pxi", "srv", "cam", "mcu":
		return true
	}
	return false
}

// readServiceName reads the service name from the srv: GetServiceHandle command
// buffer (params 1..2, 8 bytes; param 3 the length). One unresolved quirk: some
// requests store the name with each 32-bit word's two 16-bit halves swapped and
// some do not — "APT:U" arrives as words 0x50413A54/0x00556663 (half-swapped)
// while "ndm:u" arrives as 0x3A6D646E/0xFFFF0075 (straight), both from the same
// thread. Until the cause is understood, both interpretations are read and the
// one whose family the HLE knows is chosen (falling back to the straight order).
func (m *Machine) readServiceName() string {
	straight := m.decodeName(false)
	if knownService(straight) {
		return straight
	}
	swapped := m.decodeName(true)
	if knownService(swapped) {
		return swapped
	}
	return straight
}

func (m *Machine) decodeName(swap bool) string {
	n := int(m.ReadWord(tlsCmdBuf + 12)) // param 3: name length
	if n <= 0 || n > 8 {
		n = 8
	}
	var b []byte
	for w := 0; w < 2; w++ {
		word := m.ReadWord(tlsCmdBuf + 4 + uint32(w)*4)
		if swap {
			word = word>>16 | word<<16
		}
		for i := 0; i < 4; i++ {
			ch := byte(word >> (uint(i) * 8))
			if ch == 0 || len(b) >= n {
				return string(b)
			}
			b = append(b, ch)
		}
	}
	return string(b)
}

func (m *Machine) newHandle(kind string, signalled bool) uint32 {
	h := m.nextHandle
	m.nextHandle++
	m.handles[h] = &kobject{kind: kind, signal: signalled}
	return h
}

// ipcCall records one IPC request for reporting.
type ipcCall struct {
	service string
	command uint16
}

// Service returns the service/port name the request targeted ("?" if unknown).
func (c ipcCall) Service() string {
	if c.service == "" {
		return "?"
	}
	return c.service
}

// IPCLog returns the ordered IPC requests seen so far.
func (m *Machine) IPCLog() []ipcCall { return m.ipcLog }
