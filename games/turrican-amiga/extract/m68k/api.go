package m68k

// api.go — the exported surface used to RUN Turrican's own code (rather than
// reimplementing it). The CPU core in cpu.go/exec.go was lifted verbatim from the
// music driver interpreter; this file adds a general "call a routine and read back
// memory" harness with address traps (to stub out helper routines we don't want to
// execute) and panic recovery (an unimplemented opcode fails just that call).

import "fmt"

// Machine is the exported handle to the 68000 interpreter.
type Machine = m68k

// New allocates a machine with a zeroed 16 MB address space.
func New() *Machine { return newM68k() }

// Load copies data into the address space at addr.
func (m *Machine) Load(addr uint32, data []byte) { copy(m.mem[addr:], data) }

// W8/W16/W32 write directly to memory (bypassing I/O interception).
func (m *Machine) W8(a uint32, v uint8)   { m.mem[a] = v }
func (m *Machine) W16(a uint32, v uint16) { m.mem[a] = byte(v >> 8); m.mem[a+1] = byte(v) }
func (m *Machine) W32(a uint32, v uint32) { m.W16(a, uint16(v>>16)); m.W16(a+2, uint16(v)) }

// R8/R16/R32 read directly from memory.
func (m *Machine) R8(a uint32) uint8   { return m.mem[a] }
func (m *Machine) R16(a uint32) uint16 { return uint16(m.mem[a])<<8 | uint16(m.mem[a+1]) }
func (m *Machine) R32(a uint32) uint32 { return uint32(m.R16(a))<<16 | uint32(m.R16(a+2)) }

// Reg register indices for Call's regs map: 0-7 = D0-D7, 8-15 = A0-A7.
const (
	A5 = 13
	A6 = 14
	SP = 15
)

// Call runs the routine at addr with the given register presets until it returns to
// the sentinel (an RTS off the initial frame) or halts. traps holds addresses that,
// when reached, are treated as an immediate RTS — used to stub helper routines whose
// side effects we don't need (sound, spawn helpers). It recovers from unimplemented
// opcodes and runaway loops, reporting them as an error so one bad handler can't abort
// a batch. maxInsn bounds execution.
func (m *Machine) Call(addr uint32, regs map[int]uint32, traps map[uint32]bool, maxInsn int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("m68k panic: %v", r)
		}
	}()
	const sentinel = 0x00FFFFF0
	m.a[7] = 0x00FF0000 // fresh supervisor-ish stack, clear of loaded code/data
	m.a[7] -= 4
	m.wr32(m.a[7], sentinel)
	for k, v := range regs {
		if k < 8 {
			m.d[k] = v
		} else {
			m.a[k-8] = v
		}
	}
	m.pc = addr
	m.halt = false
	for n := 0; n < maxInsn; n++ {
		if m.pc == sentinel {
			return nil
		}
		if traps[m.pc] { // stub: pop return address and continue
			m.pc = m.rd32(m.a[7])
			m.a[7] += 4
			continue
		}
		if m.halt {
			return nil
		}
		m.step()
	}
	return fmt.Errorf("runaway: exceeded %d instructions at PC=%06X", maxInsn, m.pc)
}
