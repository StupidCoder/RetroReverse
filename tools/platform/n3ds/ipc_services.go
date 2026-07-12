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
	case "dsp": // the audio coprocessor — modelled in dsp.go: component load,
		// the pipe protocol, the shared-memory audio-frame exchange and the
		// frame clock the sound threads block on
		return m.ipcDSP(hdr)
	case "ndm", "ptm", "ac", "frd", "cecd", "boss", "nim", "mic", "csnd", "y2r":
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
	case 0x0006: // GetAppletInfo(appId) → {u64, u8, u8, u8, u32}. The game's wrapper
		// (0x00296E40) reads a u64 from cmdbuf[2/3], bytes from cmdbuf[4]/[5]/[6]
		// and a word from cmdbuf[7]; after PreloadLibraryApplet its poll loop
		// (0x002324FC) re-queries every 10ms and only proceeds once the cmdbuf[5]
		// AND cmdbuf[6] bytes are both nonzero — the applet's "registered" and
		// "loaded" flags. We ack preloads without running library applets, so
		// report the queried applet as present and loaded.
		m.ipcReply(hdr.Command, 0, 0, 1, 1, 1, 0)
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
	case 0x000C: // SendParameter(sender, dest applet, signal, size + handle/buffer):
		// the app messages a library applet (file-select sends applet 0x402 a
		// 32-byte parameter, signal 2) and then parks its main thread on the APT
		// condvar 0x003E242C until the applet answers. We do not run library
		// applets, so model the applet answering immediately: arm the deferred
		// APT wake (next VBlank, same rule as NotifyToWait — waking inside the
		// reply races the caller's cached session handle). The woken APT thread
		// (loop 0x00102CB4) runs InquireNotification, whose WAKEUP(1) dispatches
		// through the jump table 0x00102E9C to the handler that sets state
		// [apt+0x2C]=6 and signals that condvar.
		if m.Verbose {
			fmt.Printf("    SendParameter sender=0x%X dest=0x%X signal=%d size=0x%X\n",
				m.ReadWord(m.cmdBuf()+4), m.ReadWord(m.cmdBuf()+8),
				m.ReadWord(m.cmdBuf()+12), m.ReadWord(m.cmdBuf()+16))
		}
		// The answer carries a shared-memory block the app maps whole (size 0)
		// with rw permissions and reads the applet's response from — the map
		// thunk at 0x00296D70 feeds svcMapMemoryBlock via 0x001EFBD4. Mint a
		// zero-filled one-page block for it.
		h := m.newHandle("apt-reply-shared", false)
		m.handles[h].blockSize = 0x1000
		m.aptParams = append(m.aptParams, aptParam{
			Sender:  m.ReadWord(m.cmdBuf() + 8), // dest applet answers as sender
			Command: 3,
			Handle:  h,
		})
		m.aptWakePending = true
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x000D, 0x000E: // ReceiveParameter / GlanceParameter → the pending parameter.
		// Wrapper (0x00107EF8) reads senderId=cmdbuf[2], command=cmdbuf[3],
		// dataSize=cmdbuf[4], handle=cmdbuf[6]. Default: deliver the launch
		// WAKEUP — no sender, command WAKEUP (1), no data, no handle.
		//
		// After a SendParameter to a library applet, deliver the applet's
		// answer instead: the app's wait loop (0x00232BA0, via the deadline-
		// receive 0x0028D750) re-parks until it sees sender == the applet it
		// messaged AND command == 3 (checked at 0x00232C7C), with a NULL data
		// buffer — only sender+command matter. Receive consumes the pending
		// answer; Glance (the APT thread peeks first) leaves it pending.
		p := aptParam{Sender: 0, Command: 1} // default: launch WAKEUP
		if len(m.aptParams) > 0 {
			p = m.aptParams[0]
			if hdr.Command == 0x000D { // Receive consumes; Glance peeks
				m.aptParams = m.aptParams[1:]
				if len(m.aptParams) > 0 {
					// More queued answers: arm the next deferred wake so the
					// app's APT thread comes back for them.
					m.aptWakePending = true
				}
			}
		}
		if len(p.Data) > 0 {
			// The payload travels in the receiver's static buffer 0: the
			// wrapper (0x00107F6C) declares it in the TLS descriptor pair at
			// +0x180 (word0 = size<<14|2, word1 = pointer).
			ptr := m.ReadWord(m.curThread.tlsBase + 0x184)
			max := m.ReadWord(m.curThread.tlsBase+0x180) >> 14
			n := uint32(len(p.Data))
			if n > max {
				n = max
			}
			for i := uint32(0); i < n; i++ {
				m.Write(ptr+i, p.Data[i])
			}
		}
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|4<<6|2)
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, p.Sender)
		m.WriteWord(m.cmdBuf()+12, p.Command)
		m.WriteWord(m.cmdBuf()+16, uint32(len(p.Data))) // parameter data size
		m.WriteWord(m.cmdBuf()+20, 0)                   // translate descriptor: move 1 handle
		m.WriteWord(m.cmdBuf()+24, p.Handle)
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
	case 0x0018: // PrepareToStartLibraryApplet(appId) — wrapper 0x0029708C,
		// header 0x00180040, reply is the result word only. The applet is
		// "prepared" (we run none); StartLibraryApplet (0x001E) then queues the
		// fabricated answers.
		m.ipcReply(hdr.Command)
		return true
	case 0x0016, 0x0017: // PreloadLibraryApplet / FinishPreloadingLibraryApplet
		// The title preloads a library applet (an on-demand helper such as a
		// keyboard or selector) ahead of the file-select menu. This HLE does not run
		// library applets; acking success lets the game proceed.
		m.ipcReply(hdr.Command)
		return true
	case 0x001E: // StartLibraryApplet(appId, paramSize, handle, paramBuffer) —
		// wrapper 0x00296F50, header 0x001E0084, reply is the result word only.
		// We do not run the applet: treat it as starting and exiting at once.
		// Queue the exit answer the app then waits for — its buffered receive
		// loop (0x002915E4, buffer 0x84 bytes) loops until the parameter
		// command is one it accepts ({1,0xA,0xB,0xC} or {0xD,0xE,0xF,0x11}),
		// and the post-accept dispatch (0x002917D8) maps command 0xA to its
		// own "applet finished" class. Response payload: zeros.
		//
		// The exit choreography then needs a SECOND parameter: command 8, the
		// APT module's "resume the application" order. The game's registered
		// APT callback (0x00104200, invoked from its APT thread on each wake)
		// glances the pending parameter and dispatches commands {2,5,8,9}
		// (0x001044A4); command 8 is the ONLY path that restarts the frame
		// pacer (0x0028DBD0(0) → the frame-request walk 0x0028B9B0 → the
		// per-VBlank latch that paces the game's render thread). Without it
		// the in-game renderer never pumps its command ring, which fills and
		// deadlocks the first gameplay scene switch.
		m.aptParams = append(m.aptParams, aptParam{
			Sender:  m.ReadWord(m.cmdBuf() + 4),
			Command: 0xA,
			Data:    make([]byte, 0x84),
		},
			aptParam{Command: 8})
		m.aptWakePending = true
		m.ipcReply(hdr.Command)
		return true
	case 0x0040: // The app hands APT a buffer (wrapper 0x00296FD4: header
		// 0x00400042 = size, static-buffer descriptor, pointer) and consumes
		// only the result word — issued right after the library-applet answer
		// during the file-select flow. Nothing to model: acknowledge.
		m.ipcReply(hdr.Command)
		return true
	case 0x0101, 0x0102: // Boolean capability queries: no arguments, and the
		// wrapper (0x0010FC30, header 0x01010000) reads a single BYTE out of
		// cmdbuf[2]. These are the New-3DS-class questions ("is this a New 3DS /
		// an extended memory layout"). This machine models an ORIGINAL 3DS — one
		// application core, the 64 MiB APPMEMALLOC budget the heap sizing assumes
		// — so the honest answer is false. Answering true would promise a second
		// core and a bigger budget that nothing here provides.
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x0055: // A one-byte setter (Captain Toad's wrapper 0x00104948: header
		// const 0x00550040, STRB of a stacked byte into cmdbuf[1], and it reads
		// nothing back but the result). Nothing to model: acknowledge.
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
	case 0x0019, 0x001A: // a paired save/restore issued around the library-applet
		// hand-off after the file-select confirms a slot; both wrappers
		// (0x002961D0, 0x00107C48) send a bare header — 0x00190000 / 0x001A0000,
		// no arguments — and read only the result word: acknowledge.
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
		m.handles[sh].blockSize = hidSharedSize
		m.hidShared = sh
		m.hidEvents = m.hidEvents[:0]
		m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6|(6<<1|1))
		m.WriteWord(m.cmdBuf()+4, resultSuccess)
		m.WriteWord(m.cmdBuf()+8, 0)
		for i := 0; i < 6; i++ {
			h := sh
			if i > 0 {
				h = m.newHandle("hid-event", true)
				m.hidEvents = append(m.hidEvents, h)
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
	case 0x0001, 0x0002: // GetConfigInfoBlk2 / GetConfigInfoBlk8 — fill the block buffer
		size, blkID, out := m.ipcArg(1), m.ipcArg(2), m.ipcArg(4)
		m.writeConfigBlock(blkID, out, size)
		m.ipcReply(hdr.Command, 0)
		return true
	case 0x0003, 0x0004, 0x0005, 0x0006, 0x0007, 0x0008:
		m.ipcReply(hdr.Command, 0)
		return true
	}
	m.CPU.Halt("cfg command 0x%04X unimplemented at 0x%08X after %d instructions", hdr.Command, m.CPU.PC(), m.CPU.Instrs)
	return true
}

// writeConfigBlock fills a cfg configuration-block buffer with the value a
// European, English-language console would hold. This matters for boot: the game
// reads the system-language block (0x000A0002) to choose which LocalizedData/*
// message archive to load; a zeroed buffer reads as language 0 (Japanese), whose
// folder this European cartridge does not contain, so the message table loads
// empty and every lookup renders "NULL". Values are the documented cfg block IDs;
// unmodelled blocks are zero-filled and reported so the frontier stays explicit.
func (m *Machine) writeConfigBlock(blkID, out, size uint32) {
	if out == 0 {
		return
	}
	buf := make([]byte, size)
	switch blkID {
	case 0x000A0002: // system language (u8): 1 = English (JP=0, EN=1, FR=2, DE=3, IT=4, ES=5, ZH=6, KO=7, NL=8, PT=9, RU=10)
		if size >= 1 {
			buf[0] = cfgLangEnglish
		}
	case 0x000B0000: // region/country code (u8): EUR = 2
		if size >= 1 {
			buf[0] = cfgRegionEUR
		}
	case 0x00070001: // sound output mode (u8): 1 = stereo
		if size >= 1 {
			buf[0] = 1
		}
	case 0x00130000: // agreed EULA version (u16 minor, u16 major or u32): nonzero = accepted
		if size >= 4 {
			buf[0], buf[1] = 0x01, 0x00 // minor
			buf[2], buf[3] = 0x01, 0x00 // major
		}
	default:
		// 0x00050005 (stereo-camera calibration, 32B) and any others: zero is
		// benign for boot. Report once so an unexpected reliance shows up.
		if m.Verbose {
			fmt.Printf("cfg block 0x%08X (%d bytes) zero-filled\n", blkID, size)
		}
	}
	for i := uint32(0); i < size; i++ {
		m.Write(out+i, buf[i])
	}
}

const (
	cfgLangEnglish = 1
	cfgRegionEUR   = 2
)

func (m *Machine) ipcFS(hdr ipcHeader) bool {
	switch hdr.Command {
	case 0x0861: // InitializeWithSdkVersion(version, ProcessId) — Captain Toad's
		// wrapper (0x00114230) writes the version to cmdbuf[1] and the constant
		// 0x20 to cmdbuf[2]: the ProcessId translate descriptor, which the kernel
		// fills in for the caller. Reply is the result word only.
		m.ipcReply(hdr.Command)
		return true
	case 0x0862, 0x0863: // SetPriority / GetPriority-shaped session setters
		// (wrapper 0x0010EEE0, header 0x08620040: one word in, result only).
		m.ipcReply(hdr.Command, 0)
		return true
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
	case 0x080E: // CloseArchive(archive u64) → result — wrapper 0x0023E3D4, header 0x080E0080
		delete(m.fsArchives, m.ipcArg(1))
		m.ipcReply(hdr.Command)
		return true
	case 0x0814, 0x0817, 0x0851: // control — ack
		m.ipcReply(hdr.Command, 0, 0)
		return true
	case 0x080D: // ControlArchive(archive u64, action, in/out sizes + mapped
		// buffers) — wrapper 0x001EDFE8, reply is the result word only. The
		// game issues it to commit the save archive after writing; the
		// in-memory store needs no commit: acknowledge.
		m.ipcReply(hdr.Command)
		return true
	case 0x0845: // GetFormatInfo(archive, path) → {u32 size, u32 dirs, u32 files,
		// u8 duplicateData}. The wrapper (0x001EDF50 site) reads cmdbuf[2..4]
		// and a byte at cmdbuf[5]. An unformatted save must report "not found"
		// — a bare success ack left stale request words in the reply, which the
		// game took for a valid existing save: it skipped creation, opened
		// /GameData.bin, got NotFound, and threw fatal 0xC8804478 (the "Saving…"
		// hang after the file-select). The error must be the right CLASS: the
		// game's handler (0x001A5A68) forgives an fs-module error only when its
		// description is in [0x154,0x168) — the "save not formatted/absent"
		// group — and throws anything else. Description 0x154 with NotFound's
		// level/summary/module = 0xC8804554 routes it to FormatSaveData.
		if !m.saveFormatted {
			m.WriteWord(m.cmdBuf(), uint32(hdr.Command)<<16|1<<6)
			m.WriteWord(m.cmdBuf()+4, resultFSSaveNotFound)
			return true
		}
		m.ipcReply(hdr.Command, m.saveFormatInfo[0], m.saveFormatInfo[1], m.saveFormatInfo[2], m.saveFormatInfo[3])
		return true
	case 0x084C: // FormatSaveData(archive, path type/size, blocks, dirs, files,
		// dir/file hash-buckets, duplicateData byte + path) — wrapper 0x001EE04C,
		// header 0x084C0242, reply is the result word only. Formatting erases
		// the archive and records the layout GetFormatInfo echoes back.
		m.saveFormatted = true
		m.saveFormatInfo = [4]uint32{
			m.ipcArg(4),        // size in media blocks
			m.ipcArg(5),        // directory count
			m.ipcArg(6),        // file count
			m.ipcArg(9) & 0xFF, // duplicate-data flag
		}
		m.saveFiles = map[string][]byte{}
		m.ipcReply(hdr.Command)
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
