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

import (
	"sort"
	"strings"
)

// cacheFile is one file on the writable HDD partitions (keyed by drive-qualified
// upper-case path, e.g. "Z:/MENU.PAK").
type cacheFile struct {
	Data []byte
}

// fileObject is an open file: a disc entry (read-only) or a cache file (writable, with
// its store key so a savestate can re-link it), plus the current byte offset.
//
// dir marks a handle onto a DIRECTORY of the writable HDD store. That store is a flat
// key->bytes map with no directory records, so such a handle has neither a cache file nor
// a disc entry to carry its identity — only its key and this flag. A disc directory needs
// no flag: its Entry already says IsDir.
// scan is the directory-enumeration cursor: the index into listDir of the next child
// NtQueryDirectoryFile will report. The Xbox call returns ONE entry per call, so the
// cursor is the whole of the enumeration's state and must ride in the savestate.
type fileObject struct {
	entry Entry
	cache *cacheFile
	key   string
	dir   bool
	off   uint32
	scan  int
}

// isDir reports whether the handle names a directory, from whichever side backs it.
func (fo *fileObject) isDir() bool { return fo.dir || fo.entry.IsDir }

// fnv64 is FNV-1a over a store key, used where the HLE must mint an opaque unique file
// id for a cache file (NtQueryInformationFile's FileInternalInformation). Hashing the
// path rather than counting keeps the id stable across a savestate round trip.
func fnv64(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
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

// ---------------------------------------------------------------------------------------
// THE RAW HDD DEVICE.
// ---------------------------------------------------------------------------------------
//
// rawDevicePath maps an Xbox object path to the one raw block device this HLE serves, or
// reports that it does not name it.
//
// WHY THERE IS ONE AT ALL. XONLINE's singleton — the object 71 sites read out of the global
// at 0x57D6BC — is created by exactly one writer (0x218724), and only if its Init succeeds.
// Init's single E_FAIL (0x2183FF) is guarded by one thing:
//
//	00214699  RtlInitAnsiString(&s, "\Device\Harddisk0\partition0")   ordinal 289
//	002146D4  NtOpenFile(&h, $C0100000, &oa, &iosb, 3, $10)           ordinal 202
//	002146DA  TEST EAX,EAX / JGE
//	002146DE  OR DWORD [EBP-$4], $FFFFFFFF     the open failed -> the handle is -1
//	002183F4  CMP EAX, $FFFFFFFF               ...and Init reads that back
//	002183F7  MOV [EBX+$94], EAX               (the handle the enumerate later uses)
//	002183FD  JNZ / 002183FF MOV EAX,$80004005 -> E_FAIL, and the factory frees the object
//
// So refusing this device is what left the singleton NULL — and the game does NOT check the
// HRESULT that reports it. Its list-enumerate (0x223A7B) returns 0x80150005 WITHOUT writing
// the out-count it was given, and the caller (0x3F8C7) stores that untouched local as a
// record count. The local is MSVC's `PUSH ECX` reserve-a-dword idiom, so it still held
// `this` = 0x574618: 5,719,064 iterations over a 16-entry array, ~80M instructions and ~91MB
// of scribbled strings PER FRAME. On the console the singleton exists, the count is written
// as 0, and the JLE skips the loop.
//
// The failure was not the network. The XNET gate in that same factory (0x21868F -> 0x205105)
// returns 0; measured, not assumed. This device is the raw disk, and it is why "offline"
// needs no networking modelled at all.
//
// PARTITION0 ONLY, AND THAT IS MEASURED. An earlier cut of this served every
// \Device\Harddisk0\partitionN and the boot DIED AFTER 1,465 INSTRUCTIONS on an
// unimplemented ordinal — serving partition1 diverts the XAPI's own mount away from the
// path this port has spent five phases making work. The narrowness is not tidiness; it is
// the only version that boots.
//
// WHAT IT SAYS IS AN INVENTION, AND IT IS DECLARED. There is no HDD image here — a disc
// cannot carry one — so nothing derives this device's contents. Two halves, and only the
// first is forced:
//
//  1. THE DEVICE EXISTS. Every console has one. Refusing it is the fiction, and this file's
//     header already makes exactly this argument for T:/U:/Z: ("a missing FILE is still
//     honest; a missing PARTITION is not"). That much is not a choice.
//  2. IT IS BLANK, and that is a choice. Init and the enumerate each read a 0x1EC-byte
//     record through the handle (0x21844B, 0x223AC0) and test a signature (0x21843B /
//     0x223AB0: CMP DWORD [EBX+$1C], $56525347 — 'GSRV' on the wire). Zeros fail that test
//     and take the branch the code already has for it. The claim being made is "this console
//     has no XONLINE account", which is the same thing the title already prints across the
//     top of the screen it hangs on: NOT SIGNED IN.
//
// The risk is named because this port has paid it before: a value our own stub invents comes
// back as an observation. So — a run that depends on a NON-blank account block will not
// silently get a plausible one; it will get zeros, take the not-matched branch, and this
// comment is the thing to come back to.
func rawDevicePath(p string) (string, bool) {
	q := strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	q = strings.TrimSuffix(q, "/")
	if q == "/device/harddisk0/partition0" {
		return rawPartition0Key, true
	}
	return "", false
}

// rawPartition0Key is the raw device's key in the store, so it rides the savestate with
// every other writable byte. It is deliberately not a path any resolveCachePath prefix can
// mint, so a guest path can never collide with it.
const rawPartition0Key = "\x00RAW/Device/Harddisk0/partition0"

// rawPartition0Size is how much of the device this HLE holds.
//
// The real thing is the whole disk — gigabytes — and allocating that to hold nothing would be
// a lie in the other direction. So it is a MEASURED bound, not a guess: served 4 MiB with
// every access logged, a full cold boot touches it SIX times and never past 0x1400 —
//
//	NtOpenFile  \Device\Harddisk0\partition0        (twice, and closed in between)
//	READ  off 0x1000 len 0x200      sector 8 of a 512-byte-sector device...
//	WRITE off 0x1000 len 0x200      ...read-modify-write, twice
//	WRITE off 0x1000 len 0x200
//	READ  off 0x1200 len 0x200      sector 9
//
// — so 0x10000 is sixteen times the footprint anything has been observed to touch, and reads
// past the end return SHORT rather than inventing more (readFile clamps to size()). If a
// later frontier reads past it, the short read is the tell and this is the constant to move.
//
// THE GUEST WRITES TO IT, which is worth stating because it changes what "blank" means. This
// is not a read-only fiction handing back zeros forever: the title reads sector 8, modifies
// it, and writes it back. So the device is blank ONLY until the guest fills it, after which
// it reads back exactly what the guest put there — the behaviour of a real fresh disk, and
// the reason this lives in cacheFS: those writes ride the savestate like every other
// writable byte on this machine.
const rawPartition0Size = 0x10000

// rawDevice returns the raw device's backing bytes, minting them blank on first open.
//
// It exists as its own function so the blankness is reachable from a test: the invention this
// model makes is "the device exists and has no account on it", and a test that builds its own
// zero-filled buffer would assert that claim against itself rather than against the code that
// mints one. Lazily, because a machine nobody drives into XONLINE never allocates it.
func (m *Machine) rawDevice(key string) *cacheFile {
	if cf := m.cacheFS[key]; cf != nil {
		return cf
	}
	cf := &cacheFile{Data: make([]byte, rawPartition0Size)}
	m.cacheFS[key] = cf
	return cf
}

// statPath answers "what is at this path" without opening a handle — what
// NtQueryFullAttributesFile (ordinal 210) needs. It reports the byte size, whether the
// path is a directory, and whether it exists at all.
//
// The disc answers from its own directory tree. The writable HDD store is a flat
// key->bytes map with no directory records, so a directory there is inferred: a path
// exists as a directory when some key sits under it. That is not a shortcut — with no
// way to create an empty directory (openFile only ever mints files), "some file is under
// it" is exactly the set of directories this store can hold. The partition roots are the
// one exception: on a real console T:/U:/Z: always exist, empty or not (see the header).
func (m *Machine) statPath(path string) (size uint32, isDir, found bool) {
	if key, ok := resolveCachePath(path); ok {
		if cf := m.cacheFS[key]; cf != nil {
			return uint32(len(cf.Data)), false, true
		}
		dir := strings.TrimSuffix(key, "/")
		if len(dir) == 2 { // "Z:" — a partition root, always present
			return 0, true, true
		}
		for k := range m.cacheFS {
			if strings.HasPrefix(k, dir+"/") {
				return 0, true, true
			}
		}
		return 0, false, false
	}
	disc, ok := resolveDiscPath(path)
	if !ok || m.Disc == nil {
		return 0, false, false
	}
	e, err := m.Disc.resolve(disc)
	if err != nil {
		return 0, false, false
	}
	return e.Size, e.IsDir, true
}

// dirEntry is one child of an open directory, as NtQueryDirectoryFile reports it.
type dirEntry struct {
	Name  string
	Size  uint32
	IsDir bool
}

// listDir enumerates an open directory handle's children, in a stable order (the scan
// cursor is an INDEX into this list and rides in the savestate, so the order must not
// depend on Go's map iteration). A disc directory reads from the XISO's own tree; an HDD
// directory is reconstructed from the flat store's keys — every key under the directory
// contributes either a file (no further separator) or the subdirectory it lies in.
func (m *Machine) listDir(fo *fileObject) []dirEntry {
	if fo.dir {
		prefix := strings.TrimSuffix(fo.key, "/") + "/"
		byName := map[string]dirEntry{}
		for k, cf := range m.cacheFS {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			rest := k[len(prefix):]
			if i := strings.Index(rest, "/"); i >= 0 {
				byName[rest[:i]] = dirEntry{Name: rest[:i], IsDir: true}
			} else if rest != "" {
				byName[rest] = dirEntry{Name: rest, Size: uint32(len(cf.Data))}
			}
		}
		out := make([]dirEntry, 0, len(byName))
		for _, e := range byName {
			out = append(out, e)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
		return out
	}
	if m.Disc == nil || !fo.entry.IsDir {
		return nil
	}
	es, err := m.Disc.ReadDir(fo.entry.Path)
	if err != nil {
		return nil
	}
	out := make([]dirEntry, 0, len(es))
	for _, e := range es {
		out = append(out, dirEntry{Name: e.Name, Size: e.Size, IsDir: e.IsDir})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// matchPattern is the NT directory-search wildcard test: '*' spans any run, '?' any one
// character, everything else is a case-insensitive literal. The only pattern this title
// is observed to pass is "*" (a 1-char OBJECT_STRING; the save-game enumeration wants
// every child) — the rest of the subset is here because it is the same three lines, not
// because anything has exercised it.
func matchPattern(pat, name string) bool {
	pat, name = strings.ToUpper(pat), strings.ToUpper(name)
	// Iterative backtracking match: p/n walk the inputs, star/mark remember the last '*'.
	p, n, star, mark := 0, 0, -1, 0
	for n < len(name) {
		switch {
		case p < len(pat) && (pat[p] == '?' || pat[p] == name[n]):
			p, n = p+1, n+1
		case p < len(pat) && pat[p] == '*':
			star, mark = p, n
			p++
		case star >= 0:
			p, mark = star+1, mark+1
			n = mark
		default:
			return false
		}
	}
	for p < len(pat) && pat[p] == '*' {
		p++
	}
	return p == len(pat)
}

// The geometry NtQueryVolumeInformationFile reports for the writable HDD partitions.
// 512-byte sectors, 32 to an allocation unit: a 16 KiB unit, which is the "block" the
// console's own UI counts saves in — and the unit this title's caller reconstructs, by
// multiplying these two fields together (site 0x4291C: IMUL [buf+0x10], [buf+0x14]).
//
// hddTotalUnits is a MODEL CHOICE and the one number here that is not derived from
// anything: our HDD store is an in-memory map with no size, and the size of a console's
// save partition is a property of the console, which the disc image cannot tell us.
// 262144 units = 4 GiB, retail-plausible and comfortably past the only constraint the
// title has ever put on it — it wants 120 blocks (~2 MB) to save a game, and said so on
// screen when a failed open left it believing there were none.
const (
	hddBytesPerSector  = 512
	hddSectorsPerUnit  = 32
	hddBytesPerUnit    = hddBytesPerSector * hddSectorsPerUnit
	hddTotalUnits      = 262144
	discBytesPerSector = 2048 // XDVDFS sectors (xiso.go's sectorSize)
)

// volumeUnits reports a volume's size and free space in allocation units, for the
// FileFsSizeInformation query. The HDD's free space tracks the store's REAL contents —
// every cache file rounded up to a whole unit — so a title that fills the partition sees
// it fill. A disc is read-only and has no free space at all; its size is the image's.
func (m *Machine) volumeUnits(fo *fileObject) (total, avail uint64) {
	if fo.cache == nil && !fo.dir { // a disc file or directory
		if m.Disc == nil {
			return 0, 0
		}
		return uint64(m.Disc.Size) / discBytesPerSector, 0
	}
	used := uint64(0)
	for k, cf := range m.cacheFS {
		// The raw device shares the store (so it rides the savestate) but is NOT a file on
		// a title partition, and must not be charged to one. It is only 4 units, which is
		// exactly why this is worth a line: a raw disk quietly eating the save partition's
		// free space is the shape of the bug Part VII already spent a session on, where the
		// title told the player their console was full over a number the HLE made up.
		if k == rawPartition0Key {
			continue
		}
		used += (uint64(len(cf.Data)) + hddBytesPerUnit - 1) / hddBytesPerUnit
	}
	if used > hddTotalUnits {
		return hddTotalUnits, 0
	}
	return hddTotalUnits, hddTotalUnits - used
}

// writeNetworkOpenInfo fills a FILE_NETWORK_OPEN_INFORMATION (0x38 bytes): four
// timestamps, AllocationSize, EndOfFile, FileAttributes. Shared by the two calls that
// report it — NtQueryInformationFile's class 0x22 (by handle) and
// NtQueryFullAttributesFile (by path). Times stay 0: the XISO stamps the volume, not its
// files, so there is no per-file time to tell the truth with.
func (m *Machine) writeNetworkOpenInfo(buf, size uint32, isDir bool) {
	for i := uint32(0); i < 0x38; i += 4 {
		m.write32(buf+i, 0)
	}
	m.write32(buf+0x20, size)    // AllocationSize (low; high already 0)
	m.write32(buf+0x28, size)    // EndOfFile
	attrs := uint32(0x01 | 0x80) // READONLY|NORMAL (a DVD file)
	if isDir {
		attrs = 0x11 // READONLY|DIRECTORY
	}
	m.write32(buf+0x30, attrs)
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
	return m.readObjectString(m.read32(oa + 4))
}

// readObjectString reads a counted ANSI OBJECT_STRING. The opens reach it through an
// OBJECT_ATTRIBUTES; NtQueryDirectoryFile is handed one directly, as its search pattern.
func (m *Machine) readObjectString(p uint32) string {
	if p == 0 {
		return ""
	}
	length := uint32(m.read16(p))
	buf := m.read32(p + 4)
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

	// The raw HDD device, before the drive-letter store: it is a device, not a path within
	// one, and XONLINE's Init opens it by its \Device name. See rawDevicePath — the whole
	// argument for serving it, and for serving only this one, is there.
	if key, ok := rawDevicePath(path); ok {
		cf := m.rawDevice(key)
		m.logf("NtOpenFile: %q -> raw device (%d bytes, blank: no XONLINE account)",
			path, len(cf.Data))
		return newHandle(&fileObject{cache: cf, key: key}, infoOpened)
	}

	if key, ok := resolveCachePath(path); ok {
		cf := m.cacheFS[key]
		// A DIRECTORY of the store opens before any file lookup: with no directory
		// records to find, the flat map would answer "no such file" for a path that
		// plainly exists — and did. The XAPI opens the save partition's root (U:\) to
		// ask the volume how much room is left; the failed open meant it never got to
		// ask, and the title told the player their console was full.
		if cf == nil && disposition != dispCreate {
			if _, isDir, found := m.statPath(path); found && isDir {
				m.logf("NtOpen/CreateFile: %q -> hdd %q (directory)", path, key)
				return newHandle(&fileObject{key: key, dir: true}, infoOpened)
			}
		}
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

// setFileLength makes an open cache file exactly n bytes long, truncating or zero-filling.
// It returns an NTSTATUS, 0 on success.
//
// It is the action behind NtSetInformationFile's two 8-byte length classes (0x13 and 0x14 —
// kernel_objects.go says why they are one action here and not two), and it is deliberately
// the whole of what those classes do: this store has no allocation to reserve separately
// from the bytes it holds.
//
// GROWTH ZERO-FILLS, and that is not Go's slice semantics being borrowed by accident. A file
// extended past what was written has to read back as SOMETHING, and the two candidates are
// not equally honest: zeroes are what a filesystem is required to hand back for a hole,
// while whatever `append` finds in a recycled backing array is our allocator's leftovers
// leaking into the guest's save. `append` to a slice with spare capacity does not clear it.
func (m *Machine) setFileLength(fo *fileObject, n uint32) uint32 {
	if fo.cache == nil {
		// A disc file. The XISO is a read-only image and there is nothing to resize; this
		// halts rather than failing quietly because nothing in the image has ever asked,
		// so a status here would be a guess about a path that does not exist. writeFile
		// takes the same line for the same reason.
		m.CPU.Halt("NtSetInformationFile: set length on a read-only disc file (%q) from %08X",
			fo.entry.Path, m.retAddr())
		return 0xC000000D // STATUS_INVALID_PARAMETER, if the halt is ever cleared
	}
	if fo.isDir() {
		m.CPU.Halt("NtSetInformationFile: set length on a directory (%q) from %08X",
			fo.key, m.retAddr())
		return 0xC000000D
	}
	switch cur := uint32(len(fo.cache.Data)); {
	case n < cur:
		fo.cache.Data = fo.cache.Data[:n]
	case n > cur:
		fo.cache.Data = append(fo.cache.Data, make([]byte, n-cur)...)
	}
	// The POSITION IS LEFT ALONE, including past the new end. It is the caller's, not this
	// call's: the one site that truncates does so precisely BECAUSE the position is already
	// where it wants the file to end (0x44238 queries class 0xE and hands that value
	// straight back), so a model that helpfully clamped the position here would be moving
	// something the guest never asked to move — and would do it invisibly, since that site
	// closes the file next.
	m.logf("NtSetInformationFile: handle file %q length -> %d bytes", fo.key, n)
	return 0
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
				o.noteSignal("io", 0, m.tick)
				m.wakeWaiters(p.Event)
			}
		}
		if p.Handle != 0 {
			if o := m.objects[p.Handle]; o != nil {
				o.signaled = true // I/O complete: the file object signals
				m.writeSignal(p.Handle, true)
				o.noteSignal("io", 0, m.tick)
				m.wakeWaiters(p.Handle)
			}
		}
	}
	m.pendingIO = kept
}
