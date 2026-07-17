package ps2

// romdir.go reads the archive format the IOP's own boot images use.
//
// The PS2 keeps the IOP's kernel in a ROM this repository does not have and could
// not take. What it does have is IOPRP221.IMG, which sits on Jak's disc in
// /DRIVERS/ — and which turns out to be a ROMDIR archive holding the IOP kernel
// modules themselves: SIFCMD, SIFMAN, THREADMAN, IOMAN, MODLOAD, FILEIO, CDVDMAN,
// CDVDFSV, LOADFILE, TIMEMANI, ROMDRV, EESYNC.
//
// That is the whole reason the IOP can be run rather than faked. A game that
// reboots the IOP hands it an image like this one and the IOP comes up on the
// modules inside it, so the disc carries the very code the second processor is
// supposed to be executing. Everything a boot needs, except the handful of base
// libraries the ROM keeps to itself (see iopkernel.go), is right here.
//
// The format is as plain as it sounds. A directory of fixed 16-byte records:
//
//	name[10]   NUL-padded; an empty name ends the directory
//	extinfo    u16, the size of this entry's slice of the EXTINFO blob
//	size       u32, the size of the file itself
//
// and the file bodies follow one another from offset zero, each padded up to a
// 16-byte boundary. The first three entries describe the archive's own furniture —
// RESET, ROMDIR and EXTINFO — and the directory is thus an entry within itself,
// which is a pleasing little knot but not a complication: walking the records and
// accumulating a padded offset lands on every body.
//
// THE DIRECTORY IS NOT NECESSARILY AT OFFSET ZERO, and a real BIOS proves it. In an
// IOPRP image the first body *is* the directory, so the two coincide and a reader
// that assumes zero works by luck. In a console ROM the first body is RESET — the
// processor's reset vector, which the hardware demands be the first thing in the
// chip — and the directory sits after it (0x2700 in SCPH-10000, 0x2740 in
// SCPH-70004). So the directory is *found*, by looking for the RESET/ROMDIR/EXTINFO
// furniture that every archive opens with, and only the bodies start at zero. The
// three names together are the anchor: RESET alone appears in the ROM's code as
// ordinary bytes, and matching on it would land the directory in the middle of a
// module.

import (
	"encoding/binary"
	"errors"
	"strings"
)

// RomEntry is one file in a ROMDIR archive.
type RomEntry struct {
	Name string
	Data []byte
}

// romEntrySize is the size of one directory record.
const romEntrySize = 16

// FindROMDIR reports the offset of the archive's directory, which is the start of
// the file for an IOPRP image and a good way in for a console ROM. It looks for the
// three furniture records every archive opens with; see the note at the top of this
// file for why one name is not enough.
func FindROMDIR(raw []byte) (int, bool) {
	name := func(off int) string {
		if off+romEntrySize > len(raw) {
			return ""
		}
		return strings.TrimRight(string(raw[off:off+10]), "\x00")
	}
	for off := 0; off+3*romEntrySize <= len(raw); off += 16 {
		if name(off) == "RESET" && name(off+romEntrySize) == "ROMDIR" && name(off+2*romEntrySize) == "EXTINFO" {
			return off, true
		}
	}
	return 0, false
}

// ReadROMDIR parses a ROMDIR archive.
func ReadROMDIR(raw []byte) ([]RomEntry, error) {
	var out []RomEntry
	off, ok := FindROMDIR(raw)
	if !ok {
		return nil, errors.New("ps2: no ROMDIR directory (no RESET/ROMDIR/EXTINFO records) in this image")
	}
	body := 0

	for {
		if off+romEntrySize > len(raw) {
			return nil, errors.New("ps2: ROMDIR ran off the end of the image without a terminator")
		}
		name := strings.TrimRight(string(raw[off:off+10]), "\x00")
		size := int(binary.LittleEndian.Uint32(raw[off+12:]))
		if name == "" {
			break // the empty record that ends the directory
		}
		if body+size > len(raw) {
			return nil, errors.New("ps2: a ROMDIR entry claims more bytes than the image holds")
		}
		out = append(out, RomEntry{Name: name, Data: raw[body : body+size]})

		body += (size + 15) &^ 15
		off += romEntrySize
	}
	if len(out) == 0 {
		return nil, errors.New("ps2: the ROMDIR is empty")
	}
	return out, nil
}

// ROMDIREntry returns one named entry's bytes, or false if the archive has no such
// entry.
func ROMDIREntry(raw []byte, name string) ([]byte, bool) {
	all, err := ReadROMDIR(raw)
	if err != nil {
		return nil, false
	}
	for _, e := range all {
		if e.Name == name {
			return e.Data, true
		}
	}
	return nil, false
}

// IOPBootConf reads the IOP boot order out of a console ROM's IOPBTCONF entry — the
// list of kernel modules the machine itself starts, in the order it starts them.
// The file is one module name per line; a leading "@800" line sets the load
// address and is not a module. This is the machine's own authority on the order our
// iopBootOrder was hand-derived to match, and it is the same across ROM revisions
// four years apart, so it is a real independent reference and not a coincidence.
func IOPBootConf(bios []byte) ([]string, error) {
	blob, ok := ROMDIREntry(bios, "IOPBTCONF")
	if !ok {
		return nil, errors.New("ps2: this ROM has no IOPBTCONF entry")
	}
	var out []string
	for _, line := range strings.Fields(string(blob)) {
		if line == "" || strings.HasPrefix(line, "@") {
			continue // the "@800" load-address directive, not a module
		}
		out = append(out, line)
	}
	return out, nil
}

// ROMDIRModules returns the entries of a ROMDIR archive that are IRX modules,
// skipping the archive's own furniture (RESET, ROMDIR, EXTINFO) and anything else
// that does not begin with an ELF header.
func ROMDIRModules(raw []byte) ([]RomEntry, error) {
	all, err := ReadROMDIR(raw)
	if err != nil {
		return nil, err
	}
	var out []RomEntry
	for _, e := range all {
		if len(e.Data) >= 4 && string(e.Data[:4]) == "\x7fELF" {
			out = append(out, e)
		}
	}
	return out, nil
}
