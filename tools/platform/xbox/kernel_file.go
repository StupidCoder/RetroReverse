package xbox

// kernel_file.go is the disc-backed file I/O the boot reaches once the XDK mounts the
// title's media and starts opening files. It models the Nt* file calls against the
// mounted XISO (xiso.go): NtOpenFile resolves an Xbox device path to a disc entry and
// hands back a file handle; NtReadFile streams bytes from the disc; NtClose and the
// query calls follow. The ordinals here are pinned empirically from their call sites
// (NtOpenFile = 202, a 6-argument open with an ACCESS_MASK and an OpenOptions word),
// not from the reconstructed table, whose numbering drifts.
//
// Xbox object paths name a device and a path within it: "\Device\CdRom0\<p>" and the
// "\??\D:\<p>" symbolic-link form both mean the DVD, i.e. the mounted disc. HDD
// partitions (T:/U:/Z:, "\Device\Harddisk0\PartitionN") have no backing here — a title
// that opens a save or cache file there gets STATUS_OBJECT_NAME_NOT_FOUND, exactly as a
// freshly-formatted console would present it.

import "strings"

// fileObject is an open file: the disc entry it reads from and the current offset.
type fileObject struct {
	entry Entry
	off   uint32
}

// resolveDiscPath maps an Xbox object path to a slash path within the mounted disc, or
// reports that it does not name the disc.
func resolveDiscPath(p string) (string, bool) {
	p = strings.ReplaceAll(p, "\\", "/")
	// Case-insensitive strip of the known DVD prefixes.
	for _, pre := range []string{"/Device/CdRom0", "/??/D:", "/??/d:", "D:", "d:"} {
		if len(p) >= len(pre) && strings.EqualFold(p[:len(pre)], pre) {
			rest := p[len(pre):]
			if rest == "" {
				rest = "/"
			}
			if !strings.HasPrefix(rest, "/") {
				rest = "/" + rest
			}
			return rest, true
		}
	}
	return "", false
}

// readObjectAttributesPath reads the ANSI path out of an OBJECT_ATTRIBUTES pointer.
// OBJECT_ATTRIBUTES = { HANDLE RootDirectory; POBJECT_STRING ObjectName; ULONG Attr }.
// OBJECT_STRING (ANSI) = { USHORT Length; USHORT MaximumLength; PCHAR Buffer }.
func (m *Machine) readObjectAttributesPath(oa uint32) string {
	if oa == 0 {
		return ""
	}
	nameStr := m.read32(oa + 4)
	if nameStr == 0 {
		return ""
	}
	length := uint32(m.read16(nameStr))
	buf := m.read32(nameStr + 4)
	if buf == 0 || length == 0 || length > 1024 {
		return ""
	}
	b := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b[i] = m.Read(buf + i)
	}
	return string(b)
}

// openFile resolves a disc path and creates a file handle, writing the standard
// FileHandle / IoStatusBlock outputs. Returns the NTSTATUS.
func (m *Machine) openFile(handleOut, oa, iosb uint32) uint32 {
	path := m.readObjectAttributesPath(oa)
	disc, ok := resolveDiscPath(path)
	if !ok || m.Disc == nil {
		m.logf("NtOpenFile: %q -> not on disc", path)
		return m.finishOpen(iosb, 0, 0, 0xC0000034) // STATUS_OBJECT_NAME_NOT_FOUND
	}
	e, err := m.Disc.resolve(disc)
	if err != nil {
		m.logf("NtOpenFile: %q (disc %q) -> not found", path, disc)
		return m.finishOpen(iosb, 0, 0, 0xC0000034)
	}
	h := m.allocKObject(0x40)
	o := &kobject{kind: "file", addr: h, signaled: true}
	m.objects[h] = o
	m.files[h] = &fileObject{entry: e}
	if handleOut != 0 {
		m.write32(handleOut, h)
	}
	m.logf("NtOpenFile: %q -> disc %q (%d bytes), handle %08X", path, disc, e.Size, h)
	return m.finishOpen(iosb, h, 1, 0) // Information = FILE_OPENED(1), STATUS_SUCCESS
}

// finishOpen writes the IO_STATUS_BLOCK { NTSTATUS Status; ULONG_PTR Information } and
// returns the status (leaving the handle output to the caller).
func (m *Machine) finishOpen(iosb, _handle, info, status uint32) uint32 {
	if iosb != 0 {
		m.write32(iosb+0, status)
		m.write32(iosb+4, info)
	}
	m.setRet(status)
	return status
}

// statusPending is the NTSTATUS an in-flight asynchronous I/O reports.
const statusPending = 0x103

// ioCompletionTicks is the modelled DVD latency for an n-byte read: ~1 ms of
// seek/command overhead plus transfer at ~8 MB/s (8192 bytes per ms). An instant
// completion is a fiction real hardware never shows the guest: the title's own
// streaming engine issues an overlapped read and then runs bookkeeping that on
// hardware always finishes long before the DVD does — completing inside the call
// made the completion path release a buffer slot the issuer still held, and the
// slot refcounts went negative (the [[instant-io-races]] class, same mechanism as
// the GameCube DI).
func ioCompletionTicks(n uint32) uint64 {
	return instrsPerMs * (10 + uint64(n)/8192)
}

// pendingIO is one queued asynchronous read completion: at the due tick the
// IO_STATUS_BLOCK gets its final status/information and the optional event object
// is signalled. The read's data is already in the guest buffer — only the
// completion notification is paced.
type pendingIO struct {
	Due    uint64
	IOSB   uint32
	Event  uint32
	Info   uint32
	Status uint32
}

// readFile streams bytes from an open disc file into guest memory. NtReadFile signature
// (FileHandle, Event, ApcRoutine, ApcContext, IoStatusBlock, Buffer, Length,
// ByteOffset*) — the ByteOffset optional pointer overrides the current offset.
//
// Completion honours the caller's own protocol: the XAPI overlapped wrapper
// (0x440B3) pre-stores STATUS_PENDING in the IOSB before the call — a caller that
// did so is async-ready, gets STATUS_PENDING back, and sees the IOSB complete a
// DVD-realistic latency later (kernel_objects.go ioTick). A caller that did not
// pre-mark the IOSB is synchronous and completes inline as before.
func (m *Machine) readFile(handle, event, iosb, buffer, length, byteOffsetPtr uint32) uint32 {
	fo := m.files[handle]
	if fo == nil {
		return m.finishOpen(iosb, handle, 0, 0xC0000008) // STATUS_INVALID_HANDLE
	}
	off := fo.off
	if byteOffsetPtr != 0 {
		off = m.read32(byteOffsetPtr) // low 32 bits of the LARGE_INTEGER offset
	}
	if off > fo.entry.Size {
		off = fo.entry.Size
	}
	n := length
	if off+n > fo.entry.Size {
		n = fo.entry.Size - off
	}
	if n > 0 {
		data, err := m.Disc.Read(int64(fo.entry.Sector)*sectorSize+int64(off), int(n))
		if err != nil {
			return m.finishOpen(iosb, handle, 0, 0xC0000008)
		}
		for i, b := range data {
			m.Write(buffer+uint32(i), b)
		}
	}
	fo.off = off + n
	if iosb != 0 && m.read32(iosb) == statusPending {
		m.pendingIO = append(m.pendingIO, pendingIO{
			Due: m.tick + ioCompletionTicks(n), IOSB: iosb, Event: event, Info: n,
		})
		m.logf("NtReadFile: handle %08X off %d len %d -> %d bytes (async, event %08X, due +%d)",
			handle, off, length, n, event, ioCompletionTicks(n))
		m.setRet(statusPending)
		return statusPending
	}
	m.logf("NtReadFile: handle %08X off %d len %d -> %d bytes (sync, tick %d)", handle, off, length, n, m.tick)
	return m.finishOpen(iosb, handle, n, 0) // Information = bytes read
}

// ioTick delivers due asynchronous completions: the IOSB gets its final status and
// byte count, and the event object (if any) is signalled, waking its waiters.
func (m *Machine) ioTick() {
	if len(m.pendingIO) == 0 {
		return
	}
	kept := m.pendingIO[:0]
	for _, p := range m.pendingIO {
		if p.Due > m.tick {
			kept = append(kept, p)
			continue
		}
		if p.IOSB != 0 {
			m.write32(p.IOSB+0, p.Status)
			m.write32(p.IOSB+4, p.Info)
		}
		if p.Event != 0 {
			if o := m.objAt(p.Event); o != nil {
				o.signaled = true
				m.writeSignal(o.addr, true)
				m.wakeWaiters(p.Event)
			}
		}
	}
	m.pendingIO = kept
}
