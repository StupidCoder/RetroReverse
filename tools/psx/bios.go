package psx

// bios.go high-level-emulates the PlayStation BIOS. Instead of running a
// firmware image, the machine intercepts the A0/B0/C0 call vectors and the
// exception vector at their fixed addresses and services them in Go — the same
// approach tools/dos takes with INT 21h. Each unhandled call is logged once, so
// the next function worth implementing is always obvious.
//
// A BIOS call is made as `jal 0x000000{A,B,C}0` with the function number in $t1
// ($9) and arguments in $a0..$a3; the machine catches PC at that vector (see
// run.go), services the call, puts the result in $v0 and returns to $ra.

import (
	"fmt"
	"strings"
)

// biosCall services a BIOS call at vector table ('A', 'B' or 'C') and returns.
func (m *Machine) biosCall(table byte) {
	fn := m.CPU.Reg(9) & 0xFF // $t1
	name, ret := m.serviceBios(table, fn)
	m.biosCalls[name]++
	m.CPU.SetReg(2, ret) // $v0
	m.CPU.SetPC(m.CPU.Reg(31))
}

func (m *Machine) serviceBios(table byte, fn uint32) (string, uint32) {
	a0, a1, a2 := m.CPU.Reg(4), m.CPU.Reg(5), m.CPU.Reg(6)
	switch table {
	case 'A':
		switch fn {
		case 0x2A:
			m.memcpy(a0, a1, a2)
			return "memcpy", a0
		case 0x2B:
			m.memset(a0, a1, a2)
			return "memset", a0
		case 0x28:
			m.memset(a0, 0, a1)
			return "bzero", a0
		case 0x33:
			return "malloc", m.malloc(a0)
		case 0x34:
			return "free", 0
		case 0x39:
			m.initHeap(a0, a1)
			return "InitHeap", 0
		case 0x3F:
			m.printf(a0)
			return "printf", 0
		case 0x44:
			return "FlushCache", 0
		case 0x49:
			return "gpu_cw", 0
		case 0x72:
			return "CdRemove", 0
		case 0x96:
			return "AddCDROMDevice", 0
		case 0x97:
			return "AddMemCardDevice", 0
		case 0x99:
			return "add_nullcon_driver", 0
		case 0xA3:
			return "DequeueCdIntr", 0
		}
	case 'B':
		switch fn {
		case 0x3D:
			m.putchar(byte(a0))
			return "std_out_putchar", 0
		case 0x3F:
			m.puts(a0)
			return "std_out_puts", 0
		case 0x00:
			return "alloc_kernel_memory", m.malloc(a0)
		case 0x07:
			return "DeliverEvent", 0
		case 0x08:
			return "OpenEvent", m.openEvent()
		case 0x09:
			return "CloseEvent", 1
		case 0x0B:
			return "TestEvent", 1
		case 0x0C:
			return "EnableEvent", 1
		case 0x0D:
			return "DisableEvent", 1
		case 0x12:
			return "InitPad", 1
		case 0x13:
			return "StartPad", 1
		case 0x14:
			return "StopPad", 0
		case 0x17:
			return "ReturnFromException", 0
		case 0x18:
			return "ResetEntryInt", 0
		case 0x19:
			return "HookEntryInt", 0
		case 0x47:
			return "GetC0Table", 0
		case 0x4A:
			return "InitCard", 0
		case 0x4B:
			return "StartCard", 0
		case 0x56:
			return "GetC0Table", 0
		case 0x5B:
			return "ChangeClearPad", 0
		}
	case 'C':
		switch fn {
		case 0x00:
			return "EnqueueTimerAndVblankIrqs", 0
		case 0x01:
			return "EnqueueSyscallHandler", 0
		case 0x02:
			return "SysEnqIntRP", 0
		case 0x03:
			return "SysDeqIntRP", 0
		case 0x08:
			return "SysInitMemory", 0
		case 0x0A:
			return "ChangeClearRCnt", 0
		case 0x0C:
			return "InitDefInt", 0
		case 0x1C:
			return "AdjustA0Table", 0
		}
	}
	key := fmt.Sprintf("%c0(0x%02X)", table, fn)
	m.note("unhandled BIOS " + key)
	return key, 0
}

// handleException emulates the BIOS general-exception handler at 0x80000080. It
// services `syscall` (critical-section enable/disable) and returns; other
// exceptions are logged and skipped so a boot trace continues.
func (m *Machine) handleException() {
	cause := m.CPU.COP0[13]
	code := (cause >> 2) & 0x1F
	epc := m.CPU.COP0[14]
	switch code {
	case 8: // syscall
		switch m.CPU.Reg(4) { // $a0 selects the sub-function
		case 1: // EnterCriticalSection: return the old IEc, then disable interrupts
			m.CPU.SetReg(2, m.CPU.COP0[12]&1)
			m.CPU.COP0[12] &^= 1
		case 2: // ExitCriticalSection: enable interrupts
			m.CPU.COP0[12] |= 1
		}
		m.biosCalls[fmt.Sprintf("syscall(%d)", m.CPU.Reg(4))]++
	default:
		m.note(fmt.Sprintf("exception code %d at EPC 0x%08X", code, epc))
	}
	// rfe: pop the SR interrupt/kernel stack, and resume after the faulting op.
	sr := m.CPU.COP0[12]
	m.CPU.COP0[12] = (sr &^ 0x0F) | ((sr >> 2) & 0x0F)
	ret := epc + 4
	if cause&0x80000000 != 0 { // fault was in a branch delay slot
		ret = epc + 8
	}
	m.CPU.SetPC(ret)
}

// --- BIOS service helpers --------------------------------------------------

func (m *Machine) memcpy(dst, src, n uint32) {
	for i := uint32(0); i < n; i++ {
		m.Write(dst+i, m.Read(src+i))
	}
}

func (m *Machine) memset(dst, val, n uint32) {
	for i := uint32(0); i < n; i++ {
		m.Write(dst+i, byte(val))
	}
}

func (m *Machine) initHeap(base, size uint32) {
	m.heapPtr = base
	m.heapEnd = base + size
}

func (m *Machine) malloc(size uint32) uint32 {
	if m.heapPtr == 0 {
		m.heapPtr, m.heapEnd = 0x80180000, 0x80200000 // default heap in high RAM
	}
	size = (size + 3) &^ 3
	p := m.heapPtr
	if p+size > m.heapEnd {
		return 0 // out of heap
	}
	m.heapPtr += size
	return p
}

func (m *Machine) putchar(c byte) { m.tty = append(m.tty, c) }

func (m *Machine) puts(addr uint32) { m.tty = append(m.tty, m.readCStr(addr)...) }

func (m *Machine) readCStr(addr uint32) string {
	var b []byte
	for i := 0; i < 4096; i++ {
		c := m.Read(addr + uint32(i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

func (m *Machine) openEvent() uint32 {
	m.nextEvent++
	return 0xF1000000 | m.nextEvent
}

// printf renders a BIOS printf into the captured TTY, supporting the common
// conversions with arguments taken from $a1..$a3 then the stack.
func (m *Machine) printf(fmtAddr uint32) {
	f := m.readCStr(fmtAddr)
	argN := 0
	arg := func() uint32 {
		v := m.printfArg(argN)
		argN++
		return v
	}
	out := make([]byte, 0, len(f))
	for i := 0; i < len(f); i++ {
		if f[i] != '%' || i+1 >= len(f) {
			out = append(out, f[i])
			continue
		}
		// Capture the whole conversion spec ("%08x", "%-3d", "%ld", ...) and hand
		// it to Go's formatter, dropping C length modifiers Go does not accept.
		start := i
		i++
		for i < len(f) && strings.IndexByte("-+ #0123456789.lh", f[i]) >= 0 {
			i++
		}
		if i >= len(f) {
			out = append(out, f[start:]...)
			break
		}
		spec := strings.Map(func(r rune) rune {
			if r == 'l' || r == 'h' {
				return -1
			}
			return r
		}, f[start:i+1])
		switch f[i] {
		case '%':
			out = append(out, '%')
		case 'c':
			out = append(out, byte(arg()))
		case 's':
			out = append(out, []byte(fmt.Sprintf(spec, m.readCStr(arg())))...)
		case 'd', 'i':
			out = append(out, []byte(fmt.Sprintf(strings.Replace(spec, "i", "d", 1), int32(arg())))...)
		case 'u':
			out = append(out, []byte(fmt.Sprintf(strings.Replace(spec, "u", "d", 1), arg()))...)
		case 'x', 'X', 'o', 'b':
			out = append(out, []byte(fmt.Sprintf(spec, arg()))...)
		default:
			out = append(out, f[start:i+1]...)
		}
	}
	m.tty = append(m.tty, out...)
}

// printfArg returns the n-th printf vararg (0->$a1, 1->$a2, 2->$a3, then stack).
func (m *Machine) printfArg(n int) uint32 {
	switch n {
	case 0:
		return m.CPU.Reg(5)
	case 1:
		return m.CPU.Reg(6)
	case 2:
		return m.CPU.Reg(7)
	default:
		sp := m.CPU.Reg(29)
		a := sp + uint32(4*(n+1)) // args past $a3 spill above the arg-save area
		return uint32(m.Read(a)) | uint32(m.Read(a+1))<<8 | uint32(m.Read(a+2))<<16 | uint32(m.Read(a+3))<<24
	}
}
