package n3ds

// message.go decodes the message archives a 3DS title ships in its RomFS — the
// text of every dialog, menu and prompt. For Super Mario 3D Land the chain is
// SZS (Yaz0-compressed) → NARC (a Nitro archive of files) → per-file MSBT
// ("MsgStdBn", Nintendo's Message Studio Binary Text): a label table (LBL1) and a
// UTF-16 text block (TXT2). Everything here is derived from the byte layout of the
// game's own files, decoded with our own reader — no external message tooling.

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// Yaz0 decompresses a Yaz0 stream (magic "Yaz0", big-endian uncompressed size at
// +4, then run-length groups: a control byte whose high-to-low bits select, per
// step, a literal byte or a back-reference (nibble length + 12-bit distance, with
// a trailing length byte when the nibble is zero)).
func Yaz0(b []byte) ([]byte, error) {
	if len(b) < 16 || string(b[:4]) != "Yaz0" {
		return nil, fmt.Errorf("not a Yaz0 stream")
	}
	size := binary.BigEndian.Uint32(b[4:])
	out := make([]byte, 0, size)
	src := 16
	for uint32(len(out)) < size {
		if src >= len(b) {
			return nil, fmt.Errorf("Yaz0: truncated at %d/%d", len(out), size)
		}
		ctrl := b[src]
		src++
		for i := 0; i < 8 && uint32(len(out)) < size; i++ {
			if ctrl&0x80 != 0 {
				out = append(out, b[src])
				src++
			} else {
				b1, b2 := b[src], b[src+1]
				src += 2
				dist := (int(b1&0x0F)<<8 | int(b2)) + 1
				n := int(b1 >> 4)
				if n == 0 {
					n = int(b[src]) + 0x12
					src++
				} else {
					n += 2
				}
				start := len(out) - dist
				for j := 0; j < n; j++ {
					out = append(out, out[start+j])
				}
			}
			ctrl <<= 1
		}
	}
	return out, nil
}

// NARCFiles splits a NARC ("NARC" + BOM + version + fileSize + headerSize +
// sectionCount) into its member files. BTAF holds the file allocation table
// (start/end pairs into the GMIF image); BTNF (names) is skipped; GMIF holds the
// data. Sections follow the 0x10-byte NARC header in order.
func NARCFiles(d []byte) ([][]byte, error) {
	if len(d) < 0x10 || string(d[:4]) != "NARC" {
		return nil, fmt.Errorf("not a NARC")
	}
	le := binary.LittleEndian
	if string(d[0x10:0x14]) != "BTAF" {
		return nil, fmt.Errorf("NARC: BTAF not at 0x10")
	}
	btafSize := le.Uint32(d[0x14:])
	n := le.Uint32(d[0x18:])
	fat := make([][2]uint32, n)
	for i := uint32(0); i < n; i++ {
		off := 0x1C + i*8
		fat[i] = [2]uint32{le.Uint32(d[off:]), le.Uint32(d[off+4:])}
	}
	btnfOff := uint32(0x10) + btafSize
	if string(d[btnfOff:btnfOff+4]) != "BTNF" {
		return nil, fmt.Errorf("NARC: BTNF not after BTAF")
	}
	gmifOff := btnfOff + le.Uint32(d[btnfOff+4:])
	if string(d[gmifOff:gmifOff+4]) != "GMIF" {
		return nil, fmt.Errorf("NARC: GMIF not after BTNF")
	}
	gmifData := gmifOff + 8
	files := make([][]byte, n)
	for i, ext := range fat {
		files[i] = d[gmifData+ext[0] : gmifData+ext[1]]
	}
	return files, nil
}

// MSBTMessage is one label→text entry from an MSBT file.
type MSBTMessage struct {
	Index int
	Label string
	Text  string // UTF-16 decoded; embedded control codes appear as U+FFFF runs
}

// ParseMSBT decodes an MSBT ("MsgStdBn") file into its label→text list. It reads
// the section count from the header, walks the 0x10-aligned sections, pairs LBL1
// (hash-bucketed label → message index) with TXT2 (count + offset table + UTF-16
// strings). Labels not present leave the entry's Label empty.
func ParseMSBT(fd []byte) ([]MSBTMessage, error) {
	if len(fd) < 0x20 || string(fd[:8]) != "MsgStdBn" {
		return nil, fmt.Errorf("not an MSBT")
	}
	le := binary.LittleEndian
	nsec := int(le.Uint16(fd[0x0E:]))
	labels := map[int]string{}
	var texts []string
	p := 0x20
	for s := 0; s < nsec && p+0x10 <= len(fd); s++ {
		mag := string(fd[p : p+4])
		size := int(le.Uint32(fd[p+4:]))
		body := fd[p+0x10 : p+0x10+size]
		switch mag {
		case "TXT2":
			cnt := int(le.Uint32(body[0:]))
			for i := 0; i < cnt; i++ {
				start := int(le.Uint32(body[4+i*4:]))
				end := len(body)
				if i+1 < cnt {
					end = int(le.Uint32(body[4+(i+1)*4:]))
				}
				texts = append(texts, decodeUTF16(body[start:end]))
			}
		case "LBL1":
			ngrp := int(le.Uint32(body[0:]))
			for g := 0; g < ngrp; g++ {
				num := int(le.Uint32(body[4+g*8:]))
				off := int(le.Uint32(body[4+g*8+4:]))
				q := off
				for k := 0; k < num; k++ {
					ln := int(body[q])
					q++
					name := string(body[q : q+ln])
					q += ln
					idx := int(le.Uint32(body[q:]))
					q += 4
					labels[idx] = name
				}
			}
		}
		p += 0x10 + size
		p = (p + 0xF) &^ 0xF
	}
	out := make([]MSBTMessage, len(texts))
	for i, t := range texts {
		out[i] = MSBTMessage{Index: i, Label: labels[i], Text: t}
	}
	return out, nil
}

func decodeUTF16(b []byte) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		v := uint16(b[i]) | uint16(b[i+1])<<8
		if v == 0 {
			break
		}
		u = append(u, v)
	}
	return string(utf16.Decode(u))
}
