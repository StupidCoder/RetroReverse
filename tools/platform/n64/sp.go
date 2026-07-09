package n64

// sp.go is the RSP's register block as the CPU sees it: the status register, the
// DMA engine that moves microcode and data between RDRAM and the RSP's DMEM and
// IMEM, and the program counter that starts a task.
//
// The RSP core itself is not here — it is a CPU, and lives in tools/cpu/rsp. At
// this stage the RSP stays halted, and an attempt to start a task halts the
// oracle rather than silently doing nothing, so the boot's first display-list
// task announces itself.

const (
	spMemAddr   = 0x00 // DMEM/IMEM address for a DMA
	spDramAddr  = 0x04
	spRdLen     = 0x08 // RDRAM -> SP memory
	spWrLen     = 0x0C // SP memory -> RDRAM
	spStatus    = 0x10
	spDMAFull   = 0x14
	spDMABusy   = 0x18
	spSemaphore = 0x1C
)

// The RSP's registers occupy two windows. The eight command/status registers sit
// at 0x04040000; the program counter is alone in a second window at 0x04080000,
// far enough away that a mask over the low bits cannot tell them apart.
const (
	spRegsBase = 0x04040000
	spPCBase   = 0x04080000
	spPCEnd    = 0x040C0000
)

// SP_STATUS read bits.
const (
	spStatusHalt       = 1 << 0
	spStatusBroke      = 1 << 1
	spStatusDMABusy    = 1 << 2
	spStatusDMAFull    = 1 << 3
	spStatusIOFull     = 1 << 4
	spStatusSingleStep = 1 << 5
	spStatusIntrBreak  = 1 << 6
	spStatusSig0       = 1 << 7
)

// SP_STATUS write bits come in clear/set pairs, so a writer can change one bit
// without a read-modify-write. Modelling this as a plain value breaks the task
// handshake, which sets and clears signals from both the CPU and the RSP.
const (
	spWClearHalt      = 1 << 0
	spWSetHalt        = 1 << 1
	spWClearBroke     = 1 << 2
	spWClearIntr      = 1 << 3
	spWSetIntr        = 1 << 4
	spWClearSStep     = 1 << 5
	spWSetSStep       = 1 << 6
	spWClearIntrBreak = 1 << 7
	spWSetIntrBreak   = 1 << 8
	// bits 9..24: the eight signal bits, clear/set paired
	spWSignalBase = 9
)

func (m *Machine) spRead(addr uint32) uint32 {
	if addr >= spPCBase && addr < spPCEnd {
		return m.spPC
	}
	switch addr & 0x1F {
	case spStatus:
		return m.sp[spStatus]
	case spDMABusy:
		return m.sp[spStatus] & spStatusDMABusy
	case spDMAFull:
		return m.sp[spStatus] & spStatusDMAFull
	case spSemaphore:
		// Reading takes the semaphore: it returns the previous value and sets it.
		v := m.sp[spSemaphore]
		m.sp[spSemaphore] = 1
		return v
	}
	return m.sp[addr&0x1F]
}

func (m *Machine) spWrite(addr uint32, v uint32) {
	if addr >= spPCBase && addr < spPCEnd {
		m.spPC = v & 0xFFC
		return
	}
	switch addr & 0x1F {
	case spMemAddr, spDramAddr:
		m.sp[addr&0x1F] = v
	case spRdLen:
		m.spDMA(v, false)
	case spWrLen:
		m.spDMA(v, true)
	case spStatus:
		m.spStatusWrite(v)
	case spSemaphore:
		m.sp[spSemaphore] = 0 // any write releases it
	default:
		m.sp[addr&0x1F] = v
	}
}

func (m *Machine) spStatusWrite(v uint32) {
	s := m.sp[spStatus]
	if v&spWSetHalt != 0 {
		s |= spStatusHalt
	}
	if v&spWClearBroke != 0 {
		s &^= spStatusBroke
	}
	if v&spWClearIntr != 0 {
		m.clearIRQ(intrSP)
	}
	if v&spWSetIntr != 0 {
		m.raiseIRQ(intrSP)
	}
	if v&spWSetSStep != 0 {
		s |= spStatusSingleStep
	}
	if v&spWClearSStep != 0 {
		s &^= spStatusSingleStep
	}
	if v&spWSetIntrBreak != 0 {
		s |= spStatusIntrBreak
	}
	if v&spWClearIntrBreak != 0 {
		s &^= spStatusIntrBreak
	}
	// The eight signal bits, clear/set paired from bit 9.
	for i := uint32(0); i < 8; i++ {
		if v&(1<<(spWSignalBase+2*i)) != 0 {
			s &^= spStatusSig0 << i
		}
		if v&(1<<(spWSignalBase+2*i+1)) != 0 {
			s |= spStatusSig0 << i
		}
	}
	// Clearing halt starts the RSP on whatever microcode is in IMEM. That is a
	// task, and running it needs the RSP core.
	if v&spWClearHalt != 0 && s&spStatusHalt != 0 {
		s &^= spStatusHalt
		m.CPU.Halt("unmodelled RSP task: SP_PC=0x%03X, started from 0x%08X (tools/cpu/rsp is not wired up yet)",
			m.spPC, m.pc())
	}
	m.sp[spStatus] = s
}

// spDMA moves a block between RDRAM and the RSP's DMEM or IMEM. The length field
// counts in 8-byte units minus one, with skip/count fields above it that the
// boot path does not use.
func (m *Machine) spDMA(lenReg uint32, toRDRAM bool) {
	length := (lenReg & 0xFFF) + 1
	memAddr := m.sp[spMemAddr] & 0x1FFF
	dramAddr := m.sp[spDramAddr] & 0x00FFFFFF

	mem := m.DMEM
	if memAddr >= spMemSize {
		mem = m.IMEM
		memAddr -= spMemSize
	}
	kind := "sp-read"
	if toRDRAM {
		kind = "sp-write"
	}
	if m.OnDMA != nil {
		m.OnDMA(kind, dramAddr, memAddr, length)
	}
	for i := uint32(0); i < length; i++ {
		d := (dramAddr + i) % uint32(len(m.RDRAM))
		s := (memAddr + i) % spMemSize
		if toRDRAM {
			m.RDRAM[d] = mem[s]
		} else {
			mem[s] = m.RDRAM[d]
		}
	}
	m.sp[spDramAddr] = dramAddr + length
	m.sp[spMemAddr] = m.sp[spMemAddr] + length
}
