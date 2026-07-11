package audio

import (
	"encoding/binary"
	"fmt"
)

// seqBankMagic is the "S1" tag that opens Pilotwings' sequence bank. libultra's
// own ALSeqFile has no magic; this two-byte tag + count is the game's wrapper,
// but the per-song payload below it is standard Type-0 "compressed MIDI".
const seqBankMagic = 0x5331

// Song is one entry in the sequence bank: the raw Type-0 MIDI-derived bytes and
// its per-channel track offsets. The header is 16 big-endian words, one per MIDI
// channel, each a byte offset into Data where that channel's event stream
// begins (0 = channel unused). Channel 9 is the GM percussion track. Tempo and
// timing live inline as FF-meta events, not in the header.
type Song struct {
	Index    int
	Data     []byte  // the song's own bytes, starting at its header
	Track    [16]int // byte offset of each channel's stream, 0 if absent
	Division int     // ticks per quarter note (the word after the 16-track header)
}

// SeqBank is a parsed "S1" sequence bank: N songs carved out of one blob.
type SeqBank struct {
	Songs []*Song
}

// ParseSeqBank decodes the "S1" directory at the start of blob. Layout: u16
// magic, u16 count, then count records of {u32 offset, u32 len}, each offset
// relative to the start of blob. The song bytes follow the directory.
func ParseSeqBank(blob []byte) (*SeqBank, error) {
	if len(blob) < 4 {
		return nil, fmt.Errorf("audio: seq bank too small (%d bytes)", len(blob))
	}
	if m := binary.BigEndian.Uint16(blob); m != seqBankMagic {
		return nil, fmt.Errorf("audio: bad seq bank magic 0x%04x (want 0x%04x)", m, seqBankMagic)
	}
	n := int(binary.BigEndian.Uint16(blob[2:]))
	sb := &SeqBank{}
	for i := 0; i < n; i++ {
		rec := 4 + i*8
		if rec+8 > len(blob) {
			return nil, fmt.Errorf("audio: seq directory truncated at record %d", i)
		}
		off := int(binary.BigEndian.Uint32(blob[rec:]))
		ln := int(binary.BigEndian.Uint32(blob[rec+4:]))
		if off < 0 || ln < 0 || off+ln > len(blob) {
			return nil, fmt.Errorf("audio: song %d out of range (off=0x%x len=0x%x blob=0x%x)", i, off, ln, len(blob))
		}
		s := &Song{Index: i, Data: blob[off : off+ln]}
		if err := s.parseHeader(); err != nil {
			return nil, fmt.Errorf("audio: song %d: %w", i, err)
		}
		sb.Songs = append(sb.Songs, s)
	}
	return sb, nil
}

// parseHeader reads the 16-word channel-offset table. A zero offset means that
// channel is silent; every non-zero offset must fall inside the song.
func (s *Song) parseHeader() error {
	if len(s.Data) < 0x40 {
		return fmt.Errorf("song shorter than 16-word header (0x%x bytes)", len(s.Data))
	}
	for c := 0; c < 16; c++ {
		off := int(binary.BigEndian.Uint32(s.Data[c*4:]))
		if off != 0 && (off < 0 || off >= len(s.Data)) {
			return fmt.Errorf("channel %d track offset 0x%x outside song (0x%x)", c, off, len(s.Data))
		}
		s.Track[c] = off
	}
	// A division word sits between the 16-track header (0x40) and the first
	// track's stream (0x44), giving ticks per quarter note.
	if len(s.Data) >= 0x44 {
		s.Division = int(binary.BigEndian.Uint32(s.Data[0x40:]))
	}
	if s.Division == 0 {
		s.Division = 48 // libultra's default, if the field is absent
	}
	return nil
}

// ActiveTracks returns the channels this song actually uses, in order.
func (s *Song) ActiveTracks() []int {
	var t []int
	for c, off := range s.Track {
		if off != 0 {
			t = append(t, c)
		}
	}
	return t
}
