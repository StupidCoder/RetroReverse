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
// The format is as plain as it sounds. A directory of fixed 16-byte records sits at
// the very start of the file:
//
//	name[10]   NUL-padded; an empty name ends the directory
//	extinfo    u16, the size of this entry's slice of the EXTINFO blob
//	size       u32, the size of the file itself
//
// and the file bodies follow one another from offset zero, each padded up to a
// 16-byte boundary. The first three entries describe the archive's own furniture —
// RESET, ROMDIR and EXTINFO — and the directory is thus the first thing in the file
// *and* an entry within it, which is a pleasing little knot but not a complication:
// walking the records and accumulating a padded offset lands on every body.

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

// ReadROMDIR parses a ROMDIR archive.
func ReadROMDIR(raw []byte) ([]RomEntry, error) {
	var out []RomEntry
	off, body := 0, 0

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
