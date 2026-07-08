// Package sdat decodes the NitroSDK sound archive (SDAT) and renders its
// sequenced music: SSEQ (the MIDI-like sequence bytecode), SBNK (instrument
// banks binding note ranges to waves or PSG channels), and SWAR (wave archives
// of PCM8/PCM16/IMA-ADPCM samples). The sequencer and mixer follow the NitroSDK
// sound driver's documented semantics (48 ticks per quarter note, the DS's
// 16-channel mixer model) — the DS analogue of tools/c64/sid.
package sdat

import (
	"encoding/binary"
	"fmt"
)

var le = binary.LittleEndian

// SDAT is a parsed sound archive. Retail archives often ship without the SYMB
// name block (Mario Kart DS does); sequences are then known only by index.
type SDAT struct {
	data  []byte
	files [][]byte // FAT entries

	Seqs     []SeqInfo
	Banks    []BankInfo
	Wavearcs []WavearcInfo
}

// SeqInfo is one INFO SEQ record: a playable sequence. Name is the SYMB-block
// symbol (empty when the archive ships without SYMB).
type SeqInfo struct {
	FileID int
	Bank   int
	Vol    int
	Name   string
}

// BankInfo is one INFO BANK record: an instrument bank plus up to four wave
// archives its instruments draw samples from.
type BankInfo struct {
	FileID int
	Swars  [4]int // -1 = unused
}

// WavearcInfo is one INFO WAVEARC record.
type WavearcInfo struct{ FileID int }

// Parse decodes an SDAT container.
func Parse(data []byte) (*SDAT, error) {
	if len(data) < 0x40 || string(data[:4]) != "SDAT" {
		return nil, fmt.Errorf("sdat: not an SDAT file")
	}
	s := &SDAT{data: data}
	infoOff := int(le.Uint32(data[0x18:]))
	fatOff := int(le.Uint32(data[0x20:]))
	if string(data[infoOff:infoOff+4]) != "INFO" || string(data[fatOff:fatOff+4]) != "FAT " {
		return nil, fmt.Errorf("sdat: INFO/FAT block missing")
	}

	// FAT: {u32 off, u32 size, u64 pad} records
	nfat := int(le.Uint32(data[fatOff+8:]))
	for i := 0; i < nfat; i++ {
		o := fatOff + 12 + i*16
		fo, fs := int(le.Uint32(data[o:])), int(le.Uint32(data[o+4:]))
		if fo+fs > len(data) {
			return nil, fmt.Errorf("sdat: FAT entry %d out of bounds", i)
		}
		s.files = append(s.files, data[fo:fo+fs])
	}

	// INFO: eight sub-lists of u32 record offsets (rel to INFO block)
	list := func(kind int) [][]byte {
		base := infoOff + int(le.Uint32(data[infoOff+8+kind*4:]))
		n := int(le.Uint32(data[base:]))
		var out [][]byte
		for i := 0; i < n; i++ {
			ro := int(le.Uint32(data[base+4+i*4:]))
			if ro == 0 {
				out = append(out, nil)
				continue
			}
			out = append(out, data[infoOff+ro:])
		}
		return out
	}
	for _, r := range list(0) { // SEQ
		if r == nil {
			s.Seqs = append(s.Seqs, SeqInfo{FileID: -1})
			continue
		}
		s.Seqs = append(s.Seqs, SeqInfo{
			FileID: int(le.Uint32(r)),
			Bank:   int(le.Uint16(r[4:])),
			Vol:    int(r[6]),
		})
	}
	for _, r := range list(2) { // BANK
		if r == nil {
			s.Banks = append(s.Banks, BankInfo{FileID: -1})
			continue
		}
		b := BankInfo{FileID: int(le.Uint32(r))}
		for k := 0; k < 4; k++ {
			v := le.Uint16(r[4+k*2:])
			if v == 0xFFFF {
				b.Swars[k] = -1
			} else {
				b.Swars[k] = int(v)
			}
		}
		s.Banks = append(s.Banks, b)
	}
	for _, r := range list(3) { // WAVEARC
		if r == nil {
			s.Wavearcs = append(s.Wavearcs, WavearcInfo{FileID: -1})
			continue
		}
		s.Wavearcs = append(s.Wavearcs, WavearcInfo{FileID: int(le.Uint32(r)) & 0xFFFFFF})
	}
	s.parseSYMB()
	return s, nil
}

// parseSYMB reads the optional SYMB name block: like INFO it starts with eight
// sub-list offsets (SEQ, SEQARC, BANK, WAVEARC, PLAYER, GROUP, PLAYER2, STRM),
// each sub-list a u32 count followed by count u32 offsets — but here the
// offsets (relative to the SYMB block) point at NUL-terminated symbol names.
// Retail archives often strip the whole block (header offset 0).
func (s *SDAT) parseSYMB() {
	data := s.data
	symbOff := int(le.Uint32(data[0x10:]))
	if symbOff == 0 || symbOff+12 > len(data) || string(data[symbOff:symbOff+4]) != "SYMB" {
		return
	}
	base := symbOff + int(le.Uint32(data[symbOff+8:])) // SEQ sub-list
	n := int(le.Uint32(data[base:]))
	for i := 0; i < n && i < len(s.Seqs); i++ {
		no := int(le.Uint32(data[base+4+i*4:]))
		if no == 0 {
			continue
		}
		p := symbOff + no
		e := p
		for e < len(data) && data[e] != 0 {
			e++
		}
		s.Seqs[i].Name = string(data[p:e])
	}
}

// File returns FAT file i.
func (s *SDAT) File(i int) []byte {
	if i < 0 || i >= len(s.files) {
		return nil
	}
	return s.files[i]
}

// NumFiles returns the FAT entry count.
func (s *SDAT) NumFiles() int { return len(s.files) }
