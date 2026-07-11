package psp

// sfo.go parses a PARAM.SFO — the PSP's system parameter file, a small key/value
// table carrying a title's display name, disc id, category and firmware
// requirement. The bootoracle and the writeup read the title metadata from it.
//
// Layout (little-endian):
//
//	0x00  4  magic "\0PSF"
//	0x04  4  version (0x00000101)
//	0x08  4  key-table offset (from file start)
//	0x0C  4  data-table offset (from file start)
//	0x10  4  entry count
//	0x14  .  index: one 16-byte record per entry
//	          0x00 2  key offset (into the key table)
//	          0x02 2  data format (0x0004 UTF-8 special, 0x0204 UTF-8, 0x0404 u32)
//	          0x04 4  data length (used bytes)
//	          0x08 4  data max length (reserved bytes)
//	          0x0C 4  data offset (into the data table)
//	   .      key table: NUL-terminated key strings
//	   .      data table: values

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// SFO is a parsed PARAM.SFO. Values are either string or uint32.
type SFO struct {
	Keys   []string // in file order
	Values map[string]any
}

// ParseSFO reads a PARAM.SFO image.
func ParseSFO(b []byte) (*SFO, error) {
	if len(b) < 0x14 || string(b[0:4]) != "\x00PSF" {
		return nil, fmt.Errorf("sfo: bad magic")
	}
	keyTab := int(binary.LittleEndian.Uint32(b[0x08:]))
	dataTab := int(binary.LittleEndian.Uint32(b[0x0C:]))
	count := int(binary.LittleEndian.Uint32(b[0x10:]))

	s := &SFO{Values: make(map[string]any, count)}
	for i := 0; i < count; i++ {
		rec := 0x14 + i*16
		if rec+16 > len(b) {
			return nil, fmt.Errorf("sfo: index record %d out of range", i)
		}
		keyOff := int(binary.LittleEndian.Uint16(b[rec:]))
		fmtCode := binary.LittleEndian.Uint16(b[rec+2:])
		dataLen := int(binary.LittleEndian.Uint32(b[rec+4:]))
		dataOff := int(binary.LittleEndian.Uint32(b[rec+12:]))

		key := cstr(b, keyTab+keyOff)
		start := dataTab + dataOff
		if start+dataLen > len(b) {
			return nil, fmt.Errorf("sfo: value for %q out of range", key)
		}
		val := b[start : start+dataLen]

		switch fmtCode {
		case 0x0404: // uint32
			if len(val) >= 4 {
				s.Values[key] = binary.LittleEndian.Uint32(val)
			}
		default: // 0x0004 / 0x0204 UTF-8
			s.Values[key] = strings.TrimRight(string(val), "\x00")
		}
		s.Keys = append(s.Keys, key)
	}
	return s, nil
}

// cstr reads a NUL-terminated string starting at off.
func cstr(b []byte, off int) string {
	if off < 0 || off >= len(b) {
		return ""
	}
	end := off
	for end < len(b) && b[end] != 0 {
		end++
	}
	return string(b[off:end])
}

// String returns the value of key as a string ("" if absent or not a string).
func (s *SFO) String(key string) string {
	if v, ok := s.Values[key].(string); ok {
		return v
	}
	return ""
}

// Describe renders the table sorted by key, one "KEY = value" line per entry.
func (s *SFO) Describe() string {
	keys := append([]string(nil), s.Keys...)
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		switch v := s.Values[k].(type) {
		case uint32:
			fmt.Fprintf(&b, "%-18s = %d (0x%X)\n", k, v, v)
		default:
			fmt.Fprintf(&b, "%-18s = %v\n", k, v)
		}
	}
	return b.String()
}
