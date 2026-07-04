// Package nds reads the Nintendo DS cartridge (".nds") container format: the 512-byte
// ROM header, the two CPU binaries (ARM9/ARM7) and their overlays, and the built-in
// filesystem — a FAT (file allocation table, a flat list of start/end offsets) plus
// an FNT (file name table, the directory tree that names those files). It is the DS
// counterpart of tools/amiga/adf and tools/c64/tap: a platform container reader that
// makes no assumptions about the game inside it.
//
// The format is little-endian throughout. Layout of a retail cartridge image:
//
//	0x0000 header (0x200 used, 0x4000 reserved)
//	ARM9 binary, ARM9 overlay table + overlay files
//	ARM7 binary, ARM7 overlay table + overlay files
//	FNT (directory tree), FAT (start/end offset pairs), banner (icon/title)
//	the file data the FAT points into
//
// References cross-checked against the real Mario Kart DS (Europe) image and the
// public GBATEK cartridge-header documentation.
package nds

import (
	"encoding/binary"
	"errors"
	"strings"
)

// Header is the DS cartridge header (the fields at the start of the image). Offsets
// in the comments are byte offsets into the header.
type Header struct {
	Title      string // 0x00, 12 bytes, ASCII, NUL-padded
	GameCode   string // 0x0C, 4 bytes ("AMCP" = Mario Kart DS, PAL)
	MakerCode  string // 0x10, 2 bytes ("01" = Nintendo)
	UnitCode   byte   // 0x12, 0=NDS, 2=NDS+DSi, 3=DSi
	SeedSelect byte   // 0x13
	DeviceCap  byte   // 0x14, chip size = 128KB << DeviceCap
	ROMVersion byte   // 0x1E
	AutoStart  byte   // 0x1F

	ARM9ROMOff  uint32 // 0x20
	ARM9Entry   uint32 // 0x24
	ARM9RAMAddr uint32 // 0x28
	ARM9Size    uint32 // 0x2C
	ARM7ROMOff  uint32 // 0x30
	ARM7Entry   uint32 // 0x34
	ARM7RAMAddr uint32 // 0x38
	ARM7Size    uint32 // 0x3C

	FNTOff  uint32 // 0x40
	FNTSize uint32 // 0x44
	FATOff  uint32 // 0x48
	FATSize uint32 // 0x4C

	ARM9OverlayOff  uint32 // 0x50
	ARM9OverlaySize uint32 // 0x54
	ARM7OverlayOff  uint32 // 0x58
	ARM7OverlaySize uint32 // 0x5C

	IconOff uint32 // 0x68 icon/banner offset (0 = none)

	SecureCRC   uint16 // 0x6C secure-area checksum
	ARM9HookRAM uint32 // 0x70 auto-load-list hook (ARM9)
	ARM7HookRAM uint32 // 0x74 auto-load-list hook (ARM7)

	TotalUsedSize uint32 // 0x80 total used ROM size
	HeaderSize    uint32 // 0x84 (0x4000)

	LogoCRC   uint16 // 0x15C Nintendo-logo checksum
	HeaderCRC uint16 // 0x15E header checksum (over 0x00..0x15D)
}

// ChipBytes returns the cartridge chip size implied by DeviceCap.
func (h Header) ChipBytes() int { return (128 * 1024) << h.DeviceCap }

// ParseHeader decodes the cartridge header from the start of data.
func ParseHeader(data []byte) (Header, error) {
	if len(data) < 0x200 {
		return Header{}, errors.New("nds: image shorter than a 0x200 header")
	}
	le := binary.LittleEndian
	str := func(off, n int) string {
		return strings.TrimRight(string(data[off:off+n]), "\x00")
	}
	return Header{
		Title:      str(0x00, 12),
		GameCode:   str(0x0C, 4),
		MakerCode:  str(0x10, 2),
		UnitCode:   data[0x12],
		SeedSelect: data[0x13],
		DeviceCap:  data[0x14],
		ROMVersion: data[0x1E],
		AutoStart:  data[0x1F],

		ARM9ROMOff:  le.Uint32(data[0x20:]),
		ARM9Entry:   le.Uint32(data[0x24:]),
		ARM9RAMAddr: le.Uint32(data[0x28:]),
		ARM9Size:    le.Uint32(data[0x2C:]),
		ARM7ROMOff:  le.Uint32(data[0x30:]),
		ARM7Entry:   le.Uint32(data[0x34:]),
		ARM7RAMAddr: le.Uint32(data[0x38:]),
		ARM7Size:    le.Uint32(data[0x3C:]),

		FNTOff:  le.Uint32(data[0x40:]),
		FNTSize: le.Uint32(data[0x44:]),
		FATOff:  le.Uint32(data[0x48:]),
		FATSize: le.Uint32(data[0x4C:]),

		ARM9OverlayOff:  le.Uint32(data[0x50:]),
		ARM9OverlaySize: le.Uint32(data[0x54:]),
		ARM7OverlayOff:  le.Uint32(data[0x58:]),
		ARM7OverlaySize: le.Uint32(data[0x5C:]),

		IconOff: le.Uint32(data[0x68:]),

		SecureCRC:   le.Uint16(data[0x6C:]),
		ARM9HookRAM: le.Uint32(data[0x70:]),
		ARM7HookRAM: le.Uint32(data[0x74:]),

		TotalUsedSize: le.Uint32(data[0x80:]),
		HeaderSize:    le.Uint32(data[0x84:]),

		LogoCRC:   le.Uint16(data[0x15C:]),
		HeaderCRC: le.Uint16(data[0x15E:]),
	}, nil
}

// FATEntry is one file's byte range in the image (start inclusive, end exclusive).
type FATEntry struct {
	Start, End uint32
}

func (e FATEntry) Size() uint32 { return e.End - e.Start }

// ROM is an opened cartridge image with its header, FAT and directory tree parsed.
type ROM struct {
	Data   []byte
	Header Header
	FAT    []FATEntry     // by file ID
	Files  []FileInfo     // every file, with its full path (FAT order)
	byPath map[string]int // path (slash-separated, no leading slash) -> file ID
}

// FileInfo names a file in the filesystem and its ID (index into FAT).
type FileInfo struct {
	ID   int
	Path string
}

// Open parses a cartridge image: header, FAT, and the FNT directory tree, resolving
// every file's full path.
func Open(data []byte) (*ROM, error) {
	h, err := ParseHeader(data)
	if err != nil {
		return nil, err
	}
	r := &ROM{Data: data, Header: h, byPath: map[string]int{}}
	if err := r.parseFAT(); err != nil {
		return nil, err
	}
	if err := r.parseFNT(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *ROM) parseFAT() error {
	off, size := int(r.Header.FATOff), int(r.Header.FATSize)
	if off+size > len(r.Data) {
		return errors.New("nds: FAT out of range")
	}
	le := binary.LittleEndian
	n := size / 8
	r.FAT = make([]FATEntry, n)
	for i := 0; i < n; i++ {
		b := r.Data[off+i*8:]
		r.FAT[i] = FATEntry{Start: le.Uint32(b), End: le.Uint32(b[4:])}
	}
	return nil
}

// parseFNT walks the file name table's directory tree, assigning every file its full
// slash-separated path. The FNT's main table is an array of 8-byte directory records;
// record 0 is the root, whose parent field holds the directory count.
func (r *ROM) parseFNT() error {
	le := binary.LittleEndian
	base := int(r.Header.FNTOff)
	if base+8 > len(r.Data) {
		return errors.New("nds: FNT out of range")
	}
	dirCount := int(le.Uint16(r.Data[base+6:])) // root record's "parent" field = total dirs
	if dirCount == 0 || dirCount > 0x1000 {
		dirCount = 1 // some images store 1 here; walk defensively from the root only
	}

	// walk one directory's sub-table, recursing into subdirectories.
	var walk func(dirID uint16, prefix string) error
	walk = func(dirID uint16, prefix string) error {
		idx := int(dirID & 0x0FFF)
		rec := base + idx*8
		if rec+8 > len(r.Data) {
			return errors.New("nds: FNT directory record out of range")
		}
		sub := base + int(le.Uint32(r.Data[rec:]))
		fileID := int(le.Uint16(r.Data[rec+4:]))
		p := sub
		for {
			if p >= len(r.Data) {
				return errors.New("nds: FNT sub-table overran image")
			}
			ctrl := r.Data[p]
			p++
			if ctrl == 0 { // end of this directory
				break
			}
			nameLen := int(ctrl & 0x7F)
			if p+nameLen > len(r.Data) {
				return errors.New("nds: FNT name overran image")
			}
			name := string(r.Data[p : p+nameLen])
			p += nameLen
			full := name
			if prefix != "" {
				full = prefix + "/" + name
			}
			if ctrl&0x80 != 0 { // subdirectory: a 2-byte child dir id follows
				childID := le.Uint16(r.Data[p:])
				p += 2
				if err := walk(childID, full); err != nil {
					return err
				}
			} else { // file: consumes the next sequential file ID
				if fileID < len(r.FAT) {
					r.Files = append(r.Files, FileInfo{ID: fileID, Path: full})
					r.byPath[full] = fileID
				}
				fileID++
			}
		}
		return nil
	}
	return walk(0xF000, "")
}

// File returns the bytes of file id (its FAT slice), or nil if the id is invalid.
func (r *ROM) File(id int) []byte {
	if id < 0 || id >= len(r.FAT) {
		return nil
	}
	e := r.FAT[id]
	if int(e.End) > len(r.Data) || e.Start > e.End {
		return nil
	}
	return r.Data[e.Start:e.End]
}

// FileByPath returns the bytes of the file at a slash-separated path (no leading
// slash), or nil if there is no such file.
func (r *ROM) FileByPath(path string) []byte {
	id, ok := r.byPath[strings.TrimPrefix(path, "/")]
	if !ok {
		return nil
	}
	return r.File(id)
}

// Overlay is one entry of an ARM9/ARM7 overlay table: a separately-loadable code+data
// module swapped into a fixed RAM region on demand. Its bytes are the FAT file FileID.
type Overlay struct {
	ID              uint32 // overlay number
	RAMAddr         uint32 // load address
	RAMSize         uint32 // loaded size
	BSSSize         uint32 // trailing zero-fill
	StaticInitStart uint32 // constructor list (run after load)
	StaticInitEnd   uint32
	FileID          uint32 // FAT id holding the overlay data
	CompressedSize  uint32 // size of the compressed data (bits 0..23 of the flags word)
	Compressed      bool   // bit 24: the overlay is BLZ-compressed
}

// ARM9Overlays parses the ARM9 overlay table (32-byte records at ARM9OverlayOff).
func (r *ROM) ARM9Overlays() []Overlay {
	return r.overlays(r.Header.ARM9OverlayOff, r.Header.ARM9OverlaySize)
}

func (r *ROM) overlays(off, size uint32) []Overlay {
	if size == 0 || int(off)+int(size) > len(r.Data) {
		return nil
	}
	le := binary.LittleEndian
	n := int(size) / 32
	out := make([]Overlay, n)
	for i := 0; i < n; i++ {
		b := r.Data[int(off)+i*32:]
		flags := le.Uint32(b[28:])
		out[i] = Overlay{
			ID:              le.Uint32(b[0:]),
			RAMAddr:         le.Uint32(b[4:]),
			RAMSize:         le.Uint32(b[8:]),
			BSSSize:         le.Uint32(b[12:]),
			StaticInitStart: le.Uint32(b[16:]),
			StaticInitEnd:   le.Uint32(b[20:]),
			FileID:          le.Uint32(b[24:]),
			CompressedSize:  flags & 0x00FFFFFF,
			Compressed:      flags&0x01000000 != 0,
		}
	}
	return out
}

// ARM9 returns the ARM9 boot binary as stored in the image.
func (r *ROM) ARM9() []byte { return r.slice(r.Header.ARM9ROMOff, r.Header.ARM9Size) }

// ARM7 returns the ARM7 boot binary as stored in the image.
func (r *ROM) ARM7() []byte { return r.slice(r.Header.ARM7ROMOff, r.Header.ARM7Size) }

func (r *ROM) slice(off, size uint32) []byte {
	if int(off)+int(size) > len(r.Data) {
		return nil
	}
	return r.Data[off : off+size]
}
