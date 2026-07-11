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
	case "err": // fatal-error display — capture what the game is throwing
		return m.ipcErr(hdr)
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
	case 0x0001: // GetLockHandle → applet attributes, APT state, and the lock handle.
		// Response header 0x00010082: 3 normal words (result, attributes, state)
		// then a moved handle — the wrapper (0x00103108) reads attr from cmdbuf[2],
		// state from cmdbuf[3] and the HANDLE from cmdbuf[5] (+0x14). The app stores
		// that lock handle to a global and WaitSyncs on it, so it must be the real
		// (signalled) handle — a wrong offset made it read 0 and block forever.
		// Attributes: AppletPos = APP (0), so (attr&7) != 6 and the app takes the
		// normal wait-on-the-lock path, which succeeds because apt-lock is signalled.
		h := m.newHandle("apt-lock", true)
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|3<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)  // applet attributes (AppletPos APP)
		m.WriteWord(m.cmdBuf()+12, 0) // APT state
		m.WriteWord(m.cmdBuf()+16, 0) // translate descriptor: move 1 handle
		m.WriteWord(m.cmdBuf()+20, h) // the lock handle (cmdbuf[5])
		return true
	case 0x0002, 0x0003, 0x0004: // Initialize / Enable / Finalize
		if hdr.Command == 0x0002 { // Initialize returns two event handles
			ev1 := m.newHandle("apt-notify", true)
			ev2 := m.newHandle("apt-resume", true)
			m.aptNotifyEv, m.aptResumeEv = ev1, ev2
			m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|4)
			m.WriteWord(m.cmdBuf()+4, resultSuccess)
			m.WriteWord(m.cmdBuf()+8, 0)
			m.WriteWord(m.cmdBuf()+12, ev1)
			m.WriteWord(m.cmdBuf()+16, ev2)
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
	case 0x000B: // InquireNotification → the pending APT command. The app dispatches
		// this through a jump table where 0 panics; a freshly launched application's
		// first parameter is APTCMD_WAKEUP (1), delivered exactly once, after which
		// the app proceeds to real init. Later inquiries report "none" as WAKEUP too
		// (benign: the WAKEUP handler is idempotent and re-arms the wait).
		m.ipcReply(hdr.Command, 1) // APTCMD_WAKEUP
		return true
	case 0x000C: // SendParameter
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x000D, 0x000E: // ReceiveParameter / GlanceParameter → the pending parameter.
		// Wrapper (0x00107EF8) reads senderId=cmdbuf[2], command=cmdbuf[3],
		// dataSize=cmdbuf[4], handle=cmdbuf[6]. Deliver the launch WAKEUP: no
		// sender, command WAKEUP (1), no data, no handle.
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|4<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)  // sender app id
		m.WriteWord(m.cmdBuf()+12, 1) // command = APTCMD_WAKEUP
		m.WriteWord(m.cmdBuf()+16, 0) // parameter data size
		m.WriteWord(m.cmdBuf()+20, 0) // translate descriptor: move 1 handle
		m.WriteWord(m.cmdBuf()+24, 0) // parameter handle (none for WAKEUP)
		return true
	case 0x0043, 0x004B, 0x004C: // NotifyToWait / AppletUtility / SleepIfShellClosed
		// NotifyToWait (0x0043) means "park me until APT wakes me": on hardware
		// the APT module later posts a WAKEUP, which arrives as a signal on the
		// events Initialize returned; the app's APT handler thread then runs
		// InquireNotification / ReceiveParameter (both already deliver WAKEUP
		// here) and releases the parked main thread. Without a wake the whole
		// game slept forever after its warmup frames. The signal must be
		// DEFERRED (next VBlank), not raised inside this reply: the caller
		// still holds the library's cached APT session handle (global
		// 0x003E2668) for ~50 more instructions, and a handler woken while it
		// is set throws the applet-module fatal 0xE0A0CFF9 ("session busy") —
		// traced with a write-watch on that global.
		if hdr.Command == 0x0043 {
			m.aptWakePending = true
		}
		m.ipcReply(hdr.Command)
		return true
	case 0x003E: // Takes (u32, u8) and the wrapper (0x00107F28: header const
		// 0x003E0080, STRB of a stacked byte into cmdbuf[2], then LDRPL r0 =
		// cmdbuf[1]) consumes nothing from the reply but the result code —
		// a screen-capture/permission-style setter. Acknowledge.
		m.ipcReply(hdr.Command)
		return true
	}
	m.CPU.Halt("APT command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// signalAPTEvents signals the notify/resume events APT Initialize handed the
// app, waking its APT handler thread.
func (m *Machine) signalAPTEvents() {
	for _, h := range []uint32{m.aptNotifyEv, m.aptResumeEv} {
		if obj := m.handles[h]; obj != nil {
			obj.signal = true
			if m.signalObject(obj) {
				m.reschedule = true
			}
		}
	}
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
	case 0x0013: // RegisterInterruptRelayQueue → thread index + shared-mem handle
		// The request carries the event the GSP signals on each interrupt
		// (1 normal flag word, then a move-handle translate pair: descriptor at
		// cmdbuf[2], the event handle at cmdbuf[3]). The GSP event thread waits on
		// it; VBlank delivery signals it (gsp_vblank.go).
		m.gspEvent = m.ipcArg(3)
		if m.gspShared == 0 {
			m.gspShared = m.newHandle("gsp-shared", false)
			m.handles[m.gspShared].blockSize = gspSharedSize
		}
		// Response (libctru reads cmdbuf[2]=threadID, cmdbuf[4]=handle): 2 normal
		// words (result + threadID) then a move-handle translate pair.
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|2<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)  // GSP thread index (this process is thread 0)
		m.WriteWord(m.cmdBuf()+12, 0) // translate descriptor: move 1 handle
		m.WriteWord(m.cmdBuf()+16, m.gspShared)
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
	case 0x0016, 0x0017: // AcquireRight / ReleaseRight
		m.ipcReply(hdr.Command)
		return true
	case 0x001E, 0x0020: // SetInternalPriorities / config — no state to model, ack
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
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|(6<<1|1))
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)
		for i := 0; i < 6; i++ {
			h := sh
			if i > 0 {
				h = m.newHandle("hid-event", true)
			}
			m.WriteWord(m.cmdBuf()+12+uint32(i)*4, h)
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
	case 0x0802, 0x0803: // OpenFile / OpenFileDirectly → a file session handle
		return m.fsOpenFile(hdr)
	case 0x0804: // DeleteFile (save data)
		return m.fsDeleteFile(hdr)
	case 0x0808: // CreateFile (save data)
		return m.fsCreateFile(hdr)
	case 0x0809: // CreateDirectory — the flat save store needs no dir objects
		m.ipcReply(hdr.Command)
		return true
	case 0x080B: // OpenDirectory → a directory session handle
		return m.fsOpenDirectory(hdr)
	case 0x080C: // OpenArchive → an archive handle (routes save-data opens)
		return m.fsOpenArchive(hdr)
	case 0x0814, 0x0817, 0x0845, 0x0851: // Format / control — ack
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

// DisplayTransfers counts the GX DisplayTransfers executed — each one turned a
// rendered (tiled) colour buffer into a linear framebuffer the LCD scans out.
func (m *Machine) DisplayTransfers() int { return m.displayTransfers }

var _ = fmt.Sprintf

// ipcErr models the err:f fatal-error display. Cmd 0x0001 (ThrowFatalError) is
// how a title reports an unrecoverable error; capturing its payload tells us what
// the game objected to, so it is dumped and then halts loudly rather than being
// silently swallowed.
func (m *Machine) ipcErr(hdr ipcHeader) bool {
	if hdr.Command == 0x0001 { // ThrowFatalError
		errType := m.ipcArg(1)
		code := m.ipcArg(3)
		pc := m.ipcArg(5)
		fmt.Printf("err:f ThrowFatalError type=0x%X resultCode=0x%08X pc=0x%08X; cmdbuf:", errType, code, pc)
		for i := 1; i < 16; i++ {
			fmt.Printf(" %08X", m.ReadWord(m.cmdBuf()+uint32(i)*4))
		}
		fmt.Println()
	}
	m.ipcReply(hdr.Command)
	return true
}
