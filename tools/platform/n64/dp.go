package n64

// dp.go is the RDP's command-queue interface: the registers through which the
// CPU (or the RSP, through its COP0 window) hands the rasteriser a span of
// display-list commands to execute. The commands themselves are in rdp.go, and
// the pixels they produce in rdp_raster.go.

const (
	dpStart    = 0x00
	dpEnd      = 0x04
	dpCurrent  = 0x08
	dpStatus   = 0x0C
	dpClock    = 0x10
	dpBufBusy  = 0x14
	dpPipeBusy = 0x18
	dpTMem     = 0x1C
)

// DPC_STATUS bits. The source bit selects whether commands are read from RDRAM
// or from the RSP's data memory.
const (
	dpStatusXBusDMEM   = 1 << 0
	dpStatusFreeze     = 1 << 1
	dpStatusFlush      = 1 << 2
	dpStatusCmdBusy    = 1 << 6
	dpStatusCbufReady  = 1 << 7
	dpStatusDMABusy    = 1 << 8
	dpStatusEndValid   = 1 << 9
	dpStatusStartValid = 1 << 10
)

func (m *Machine) dpRead(addr uint32) uint32 {
	switch addr & 0x1F {
	case dpCurrent, dpEnd:
		// The queue drains instantly, so CURRENT always equals END: a game that
		// polls for the RDP to catch up falls straight through.
		return m.dp[dpEnd]
	case dpStatus:
		return m.dp[dpStatus] | dpStatusCbufReady
	}
	return m.dp[addr&0x1F]
}

func (m *Machine) dpWrite(addr uint32, v uint32) {
	switch addr & 0x1F {
	case dpStart:
		// Writing the start address arms the queue: the address latches and
		// start-valid is set. While armed, further start writes are ignored —
		// the hardware holds the first address until the matching end write,
		// and n64-systemtest checks that a second write changes nothing.
		// CURRENT does not move here; it moves when the end write commits.
		if m.dp[dpStatus]&dpStatusStartValid == 0 {
			m.dp[dpStart] = v & 0x00FFFFF8
			m.dp[dpStatus] |= dpStatusStartValid
		}
	case dpEnd:
		m.dp[dpEnd] = v & 0x00FFFFF8
		if m.dp[dpStatus]&dpStatusStartValid != 0 {
			m.dp[dpCurrent] = m.dp[dpStart]
			m.dp[dpStatus] &^= dpStatusStartValid
		}
		if m.dp[dpStatus]&dpStatusFreeze == 0 {
			m.runRDP()
		}
	case dpStatus:
		// Paired clear/set bits, as elsewhere in the RCP.
		if v&(1<<0) != 0 {
			m.dp[dpStatus] &^= dpStatusXBusDMEM
		}
		if v&(1<<1) != 0 {
			m.dp[dpStatus] |= dpStatusXBusDMEM
		}
		if v&(1<<2) != 0 {
			// Unfreezing releases whatever the queue received while frozen.
			m.dp[dpStatus] &^= dpStatusFreeze
			m.runRDP()
		}
		if v&(1<<3) != 0 {
			m.dp[dpStatus] |= dpStatusFreeze
		}
		if v&(1<<4) != 0 {
			m.dp[dpStatus] &^= dpStatusFlush
		}
		if v&(1<<5) != 0 {
			m.dp[dpStatus] |= dpStatusFlush
		}
	default:
		m.dp[addr&0x1F] = v
	}
}
