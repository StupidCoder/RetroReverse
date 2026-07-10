package n3ds

import "fmt"

// ipc_services.go implements the individual OS services, grown lazily. Each is
// modelled just far enough to keep a title's init moving toward its first frame;
// commands not yet needed halt with the service and command ID so the frontier
// is always explicit. Where a service hands back a handle (a shared-memory
// block, an event, a mutex) the HLE mints a stub and, for waitable objects,
// reports it signalled so the game does not block on a scheduler this machine
// does not run.

func (m *Machine) ipcService(name string, hdr ipcHeader) bool {
	switch serviceBase(name) {
	case "APT": // applet manager — the app-lifecycle handshake
		return m.ipcAPT(name, hdr)
	case "gsp": // graphics — the path to a frame
		return m.ipcGSP(hdr)
	case "hid": // input
		return m.ipcHID(hdr)
	case "cfg": // system configuration (region, language)
		return m.ipcCFG(hdr)
	case "fs": // filesystem (RomFS, save data)
		return m.ipcFS(hdr)
	case "ndm", "ptm", "ac", "frd", "cecd", "boss", "nim", "mic", "csnd", "dsp", "y2r":
		// Background/optional services: acknowledge init-shaped commands so the
		// game's optional subsystems do not stall the boot.
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("service %q command 0x%04X unimplemented at 0x%08X after %d instructions",
		name, hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// serviceBase strips a service's ":U"/":LR"/"::Gsp" suffix to its family.
func serviceBase(name string) string {
	for i, ch := range name {
		if ch == ':' {
			return name[:i]
		}
	}
	return name
}

// ipcAPT models the applet manager. The real APT handshake is an elaborate
// state machine (lock handle, notification events, a "wakeup" parameter the
// launcher delivers); the HLE gives the app what it needs to believe it is the
// foreground application and proceed: handles for its notification and resume
// events, and an APT_Wrap-free "you are running" status.
func (m *Machine) ipcAPT(name string, hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0001: // GetLockHandle → APT mutex + applet attributes
		h := m.newHandle("apt-lock", true)
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|3<<6|2)
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0)  // applet attributes
		m.WriteWord(tlsCmdBuf+12, 0) // APT state
		m.WriteWord(tlsCmdBuf+16, 0)
		m.WriteWord(tlsCmdBuf+20, 0) // translate: move handle
		m.WriteWord(tlsCmdBuf+24, h)
		return true
	case 0x0002, 0x0003, 0x0004: // Initialize / Enable / Finalize
		if hdr.Command == 0x0002 { // Initialize returns two event handles
			ev1 := m.newHandle("apt-notify", true)
			ev2 := m.newHandle("apt-resume", true)
			m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|1<<6|4)
			m.WriteWord(tlsCmdBuf+4, resultSuccess)
			m.WriteWord(tlsCmdBuf+8, 0)
			m.WriteWord(tlsCmdBuf+12, ev1)
			m.WriteWord(tlsCmdBuf+16, ev2)
			return true
		}
		m.ipcReply(hdr.Command)
		return true
	case 0x0005: // GetAppletManInfo
		m.ipcReply(hdr.Command, 0, 0, 0x300, 0x300) // active app id fields
		return true
	case 0x0006: // GetAppletInfo
		m.ipcReply(hdr.Command, 0, 0, 1, 0, 0, 0)
		return true
	case 0x0009: // IsRegistered
		m.ipcReply(hdr.Command, 1)
		return true
	case 0x000B, 0x000C, 0x000D: // InquireNotification / Send/ReceiveParameter
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x0043, 0x004B, 0x004C: // NotifyToWait / AppletUtility / SleepIfShellClosed
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("APT command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// ipcGSP models the graphics service — the interface a title drives to present
// a frame. The key handshake is RegisterInterruptRelayQueue, which returns the
// GSP shared-memory block (the GX command queue and the interrupt/VBlank flags)
// and the thread index; from there the game writes GPU register commands and
// triggers frame swaps. This HLE records the frame-submission calls so a run can
// report reaching the first frame, but does not execute PICA200 commands into a
// framebuffer — that GPU is a later milestone.
func (m *Machine) ipcGSP(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0013: // RegisterInterruptRelayQueue → shared-mem handle + thread index
		if m.gspShared == 0 {
			m.gspShared = m.newHandle("gsp-shared", false)
		}
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|2<<6|2)
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0) // FirstInitialization flag
		m.WriteWord(tlsCmdBuf+12, 0) // thread index
		m.WriteWord(tlsCmdBuf+16, 0)
		m.WriteWord(tlsCmdBuf+20, m.gspShared)
		return true
	case 0x0001, 0x0002, 0x0003, 0x0004, 0x0005: // Write/Read HW regs, flush cache
		m.ipcReply(hdr.Command)
		return true
	case 0x0008: // FlushDataCache
		m.ipcReply(hdr.Command)
		return true
	case 0x000A: // SetLcdForceBlack
		m.ipcReply(hdr.Command)
		return true
	case 0x000B: // TriggerCmdReqQueue — the game submits a GPU command list
		m.framesSubmitted++
		m.ipcReply(hdr.Command)
		return true
	case 0x000C, 0x000D, 0x000E: // Set/ClearInterrupt, register events
		m.ipcReply(hdr.Command)
		return true
	case 0x0016: // AcquireRight
		m.ipcReply(hdr.Command)
		return true
	case 0x0018: // ImportDisplayCaptureInfo
		m.ipcReply(hdr.Command, 0, 0, 0, 0, 0, 0, 0, 0)
		return true
	case 0x001F: // SetBufferSwap — presents a finished frame
		m.framesSwapped++
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("gsp command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

func (m *Machine) ipcHID(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x000A: // GetIPCHandles → shared-mem + 5 event handles
		sh := m.newHandle("hid-shared", false)
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|1<<6|(6<<1|1))
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0)
		for i := 0; i < 6; i++ {
			h := sh
			if i > 0 {
				h = m.newHandle("hid-event", true)
			}
			m.WriteWord(tlsCmdBuf+12+uint32(i)*4, h)
		}
		return true
	case 0x0011, 0x0012, 0x0013: // EnableAccelerometer / Gyroscope / etc.
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("hid command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

func (m *Machine) ipcCFG(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0001, 0x0002: // GetConfigInfoBlk2 / GetRegion — return a plausible EUR/EN config
		m.ipcReply(hdr.Command, 0) // block contents are written to a mapped buffer we accept as zero
		return true
	case 0x0003, 0x0004, 0x0005, 0x0006, 0x0007, 0x0008:
		m.ipcReply(hdr.Command, 0)
		return true
	}
	m.CPU.Halt("cfg command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

func (m *Machine) ipcFS(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0801: // Initialize
		m.ipcReply(hdr.Command)
		return true
	case 0x0802, 0x0803: // OpenFile / OpenFileDirectly → a file handle
		h := m.newHandle("fs-file", false)
		m.WriteWord(tlsCmdBuf, uint32(hdr.Command)<<16|1<<6|2)
		m.WriteWord(tlsCmdBuf+4, resultSuccess)
		m.WriteWord(tlsCmdBuf+8, 0)
		m.WriteWord(tlsCmdBuf+12, h)
		return true
	case 0x080C, 0x0814, 0x0817, 0x0845, 0x0851: // OpenArchive / Format / etc.
		m.ipcReply(hdr.Command, 0, 0)
		return true
	}
	m.CPU.Halt("fs command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// FrameStats reports how far toward a frame the run reached.
func (m *Machine) FrameStats() (submitted, swapped int) {
	return m.framesSubmitted, m.framesSwapped
}

var _ = fmt.Sprintf
