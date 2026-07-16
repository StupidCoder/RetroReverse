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
// "\??\D:\<p>" symbolic-link form both mean the DVD, i.e. the mounted disc.
//
// The HDD's title partitions (T:/U: persistent data, Z: the utility/cache drive) are a
// WRITABLE in-memory filesystem: on a real console these partitions always exist, and
// OutRun's menu loader depends on that — it unpacks its menu resources into
// z:\MENU.PAK on first boot and re-opens that cache forever after. When they were
// unbacked (every open answered STATUS_OBJECT_NAME_NOT_FOUND), the cache could never
// be built and the loading screen retried the open for the rest of the run. A missing
// FILE is still honest (a fresh console has an empty cache); a missing PARTITION is
// not. The store is part of the savestate.

import "strings"

// cacheFile is one file on the writable HDD partitions (keyed by drive-qualified
// upper-case path, e.g. "Z:/MENU.PAK").
type cacheFile struct {
	Data []byte
}

// fileObject is an open file: a disc entry (read-only) or a cache file (writable, with
// its store key so a savestate can re-link it), plus the current byte offset.
type fileObject struct {
	entry Entry
	cache *cacheFile
	key   string
	off   uint32
}

// size is the object's current byte length (a cache file grows as it is written).
func (fo *fileObject) size() uint32 {
	if fo.cache != nil {
		return uint32(len(fo.cache.Data))
	}
	return fo.entry.Size
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

// resolveCachePath maps an Xbox object path to a key in the writable HDD store, or
// reports that it does not name one of the writable partitions.
func resolveCachePath(p string) (string, bool) {
	p = strings.ReplaceAll(p, "\\", "/")
	for _, pre := range []string{"/??/T:", "/??/U:", "/??/Z:", "T:", "U:", "Z:"} {
		if len(p) >= len(pre) && strings.EqualFold(p[:len(pre)], pre) {
			drive := pre[len(pre)-2:]
			rest := p[len(pre):]
			if !strings.HasPrefix(rest, "/") {
				rest = "/" + rest
			}
			return strings.ToUpper(drive + rest), true
		}
	}
	return "", false
}

// DebugCacheFS lists the writable HDD store: key -> size. DebugCacheFile returns one
// file's bytes. Diagnostic accessors (like DebugThreads) — the verify-the-shipped-file
// check compares an installed Z: copy byte-for-byte against its disc source.
func (m *Machine) DebugCacheFS() map[string]int {
	out := make(map[string]int, len(m.cacheFS))
	for k, v := range m.cacheFS {
		out[k] = len(v.Data)
	}
	return out
}

func (m *Machine) DebugCacheFile(key string) []byte {
	if cf := m.cacheFS[key]; cf != nil {
		return cf.Data
	}
	return nil
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

// NtCreateFile dispositions and the IOSB Information codes an open reports.
const (
	dispSupersede   = 0
	dispOpen        = 1
	dispCreate      = 2
	dispOpenIf      = 3
	dispOverwrite   = 4
	dispOverwriteIf = 5

	infoOpened      = 1
	infoCreated     = 2
	infoOverwritten = 3
)

// openFile resolves a path against the disc (read-only) or the writable HDD
// partitions, creates a file handle, and writes the standard FileHandle /
// IoStatusBlock outputs. NtOpenFile calls it with dispOpen; NtCreateFile passes the
// caller's CreateDisposition, which on the HDD store may create or truncate.
func (m *Machine) openFile(handleOut, oa, iosb uint32, disposition uint32) uint32 {
	path := m.readObjectAttributesPath(oa)

	newHandle := func(fo *fileObject, info uint32) uint32 {
		h := m.allocKObject(0x40)
		m.objects[h] = &kobject{kind: "file", addr: h, signaled: true}
		// A FILE_OBJECT is a dispatcher object: signalled while no I/O is in flight.
		// The guest header must agree — objAt resyncs from it, and the XAPI
		// GetOverlappedResult path waits on the file handle itself when the caller
		// supplied no event (TitleScreen.xmv's streaming pump, tid 1, site 0x44E13).
		m.writeSignal(h, true)
		m.files[h] = fo
		if handleOut != 0 {
			m.write32(handleOut, h)
		}
		return m.finishOpen(iosb, h, info, 0)
	}

	if key, ok := resolveCachePath(path); ok {
		cf := m.cacheFS[key]
		switch {
		case cf == nil:
			if disposition == dispOpen || disposition == dispOverwrite {
				m.logf("NtOpen/CreateFile: %q (hdd %q) -> not found (disp %d)", path, key, disposition)
				return m.finishOpen(iosb, 0, 0, 0xC0000034) // STATUS_OBJECT_NAME_NOT_FOUND
			}
			cf = &cacheFile{}
			m.cacheFS[key] = cf
			m.logf("NtCreateFile: %q -> hdd %q CREATED", path, key)
			return newHandle(&fileObject{cache: cf, key: key}, infoCreated)
		case disposition == dispCreate:
			m.logf("NtCreateFile: %q (hdd %q) -> collision", path, key)
			return m.finishOpen(iosb, 0, 0, 0xC0000035) // STATUS_OBJECT_NAME_COLLISION
		case disposition == dispSupersede, disposition == dispOverwrite, disposition == dispOverwriteIf:
			cf.Data = cf.Data[:0]
			m.logf("NtCreateFile: %q -> hdd %q overwritten", path, key)
			return newHandle(&fileObject{cache: cf, key: key}, infoOverwritten)
		default:
			m.logf("NtOpenFile: %q -> hdd %q (%d bytes)", path, key, len(cf.Data))
			return newHandle(&fileObject{cache: cf, key: key}, infoOpened)
		}
	}

	disc, ok := resolveDiscPath(path)
	if !ok || m.Disc == nil {
		m.logf("NtOpenFile: %q -> not on disc", path)
		return m.finishOpen(iosb, 0, 0, 0xC0000034)
	}
	e, err := m.Disc.resolve(disc)
	if err != nil {
		m.logf("NtOpenFile: %q (disc %q) -> not found", path, disc)
		return m.finishOpen(iosb, 0, 0, 0xC0000034)
	}
	m.logf("NtOpenFile: %q -> disc %q (%d bytes)", path, disc, e.Size)
	return newHandle(&fileObject{entry: e}, infoOpened)
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
	Handle uint32 // the file handle: its FILE_OBJECT de-signals while the I/O is in
	// flight and signals at completion — the wait target when the caller gave no event
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
func (m *Machine) readFile(handle, event, apcRoutine, apcCtx, iosb, buffer, length, byteOffsetPtr uint32) uint32 {
	if apcRoutine != 0 {
		// No caller has exercised APC completion yet; if one appears it must be
		// implemented, not dropped — a swallowed completion callback deadlocks the
		// caller's pump with nothing in the log to say why.
		m.CPU.Halt("NtReadFile: ApcRoutine %08X (ctx %08X) passed from %08X — APC completion not implemented", apcRoutine, apcCtx, m.retAddr())
		return 0
	}
	fo := m.files[handle]
	if fo == nil {
		return m.finishOpen(iosb, handle, 0, 0xC0000008) // STATUS_INVALID_HANDLE
	}
	off := fo.off
	if byteOffsetPtr != 0 {
		off = m.read32(byteOffsetPtr) // low 32 bits of the LARGE_INTEGER offset
	}
	size := fo.size()
	if off > size {
		off = size
	}
	n := length
	if off+n > size {
		n = size - off
	}
	if n > 0 {
		var data []byte
		if fo.cache != nil {
			data = fo.cache.Data[off : off+n]
		} else {
			var err error
			data, err = m.Disc.Read(int64(fo.entry.Sector)*sectorSize+int64(off), int(n))
			if err != nil {
				return m.finishOpen(iosb, handle, 0, 0xC0000008)
			}
		}
		for i, b := range data {
			m.Write(buffer+uint32(i), b)
		}
	}
	fo.off = off + n
	if iosb != 0 && m.read32(iosb) == statusPending {
		if o := m.objects[handle]; o != nil {
			o.signaled = false // I/O in flight: the file object de-signals
			m.writeSignal(handle, false)
		}
		m.pendingIO = append(m.pendingIO, pendingIO{
			Due: m.tick + ioCompletionTicks(n), IOSB: iosb, Event: event, Info: n, Handle: handle,
		})
		m.logf("NtReadFile: handle %08X off %d len %d -> %d bytes (async, event %08X, iosb %08X, buf %08X, due +%d, from %08X)",
			handle, off, length, n, event, iosb, buffer, ioCompletionTicks(n), m.retAddr())
		m.setRet(statusPending)
		return statusPending
	}
	m.logf("NtReadFile: handle %08X off %d len %d -> %d bytes (sync, tick %d)", handle, off, length, n, m.tick)
	return m.finishOpen(iosb, handle, n, 0) // Information = bytes read
}

// writeFile stores bytes into an open HDD-partition file, growing it as needed.
// NtWriteFile's signature mirrors NtReadFile (FileHandle, Event, ApcRoutine,
// ApcContext, IoStatusBlock, Buffer, Length, ByteOffset*). Disc files are read-only
// media — a write to one halts rather than fake success. HDD writes complete
// synchronously (the pre-marked-IOSB async protocol readFile honours has only ever
// been seen on the DVD streaming path; if a writer pre-marks, the same pacing can
// graduate here).
func (m *Machine) writeFile(handle, event, iosb, buffer, length, byteOffsetPtr uint32) uint32 {
	fo := m.files[handle]
	if fo == nil {
		return m.finishOpen(iosb, handle, 0, 0xC0000008) // STATUS_INVALID_HANDLE
	}
	if fo.cache == nil {
		m.CPU.Halt("NtWriteFile: write to a read-only disc file (handle %08X) from %08X", handle, m.retAddr())
		return 0
	}
	off := fo.off
	if byteOffsetPtr != 0 {
		off = m.read32(byteOffsetPtr)
	}
	if need := int(off) + int(length); need > len(fo.cache.Data) {
		fo.cache.Data = append(fo.cache.Data, make([]byte, need-len(fo.cache.Data))...)
	}
	for i := uint32(0); i < length; i++ {
		fo.cache.Data[off+i] = m.Read(buffer + i)
	}
	fo.off = off + length
	m.logf("NtWriteFile: handle %08X off %d len %d (file now %d bytes)", handle, off, length, len(fo.cache.Data))
	return m.finishOpen(iosb, handle, length, 0)
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
		if p.Handle != 0 {
			if o := m.objects[p.Handle]; o != nil {
				o.signaled = true // I/O complete: the file object signals
				m.writeSignal(p.Handle, true)
				m.wakeWaiters(p.Handle)
			}
		}
	}
	m.pendingIO = kept
}
