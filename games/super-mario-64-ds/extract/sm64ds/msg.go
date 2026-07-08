package sm64ds

// The message system (dialog, menus, and the course names).
//
// All text lives in BMG containers: data/message/msg_data_<lang>.bin (the 711
// game messages, LZ77-tagged) plus five small BMGs embedded in the ARM9 (the
// pre-boot option menus, one per language). The format and encoding are pinned
// from the game itself:
//
//   - Container: "MESG"/"bmg1" magic (stored byte-swapped, "GSEM1gmb"), a 0x20
//     header, then tagged sections. The game's parser (overlay 7, $020C951C)
//     walks the sections comparing the tag word against INF1/FLW1/DAT1/STR1/
//     FLI1 constants and stores the INF1 section base, INF1+0x10 (the entry
//     array) and DAT1+8 (the string pool) in globals $02104C1C/$02104C18/
//     $02104C24. The string getter $020C94A0 computes
//     stringPool + u32(entries[id * (entrySizeField>>3)]) — the INF1 header
//     holds count (u16 at +8) and an entry-size field (u16 at +0xA, value
//     0x40, used as 0x40>>3 = 8 bytes per entry: u32 offset + u32 attributes).
//
//   - Encoding: a message byte is a GLYPH INDEX into the dialog font
//     (ARCHIVE/c2d.narc member 13: 16x16 4bpp glyphs, two tiles wide in a
//     32-tile-per-row sheet). Reading the sheet in index order gives the
//     charset: '0'-'9' at 0, 'A'-'Z' at $0A, 'a'-'z' at $2D, punctuation
//     between and after, and a blank cell at $4D — the space. $FD is a
//     newline, $FF terminates a message, and $FE opens a control whose next
//     byte is the control's total length (button/icon escapes like the d-pad
//     glyph in the slide instructions). The mapping is cross-checked against
//     the ARM9-embedded menus ("CONTINUE", "EXIT COURSE", "DUAL-HAND MODE",
//     French "CONTINUER"/"QUITTER NIVEAU") and Peach's opening letter.
//
// The course names are messages 406..435 (in course-index order — see
// level.go Course): the 15 numbered painting courses, the three Bowser roads,
// the three DS boss courses, the secret courses, and "CASTLE SECRET STARS".

import (
	"fmt"
	"os"

	"retroreverse.com/tools/platform/nds"
)

// glyphChars is the dialog-font glyph order, read off the font sheet
// (ARCHIVE/c2d.narc member 13). Zero value = unmapped (accents and symbols in
// the upper half of the sheet, decoded as <XX> placeholders).
var glyphChars = map[byte]rune{
	0x26: '?', 0x27: '!', 0x28: '~', 0x29: ',', 0x2A: '“', 0x2B: '”', 0x2C: '·',
	0x47: '-', 0x48: '.', 0x49: '\'', 0x4A: ':', 0x4B: ';', 0x4C: '&', 0x4D: ' ', 0x4E: '/',
}

func init() {
	for i := 0; i < 10; i++ {
		glyphChars[byte(i)] = rune('0' + i)
	}
	for i := 0; i < 26; i++ {
		glyphChars[byte(0x0A+i)] = rune('A' + i)
		glyphChars[byte(0x2D+i)] = rune('a' + i)
	}
}

// LoadBMG reads a BMG message file (LZ77-tagged on the card, like the .bmd
// models) and returns the decoded messages in INF1 order.
func LoadBMG(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) > 4 && string(raw[:4]) == "LZ77" {
		raw = nds.Decompress(raw[4:])
	}
	return ParseBMG(raw)
}

// ParseBMG decodes an uncompressed BMG container.
func ParseBMG(data []byte) ([]string, error) {
	if len(data) < 0x20 || string(data[:8]) != "GSEM1gmb" {
		return nil, fmt.Errorf("sm64ds: not a BMG message file")
	}
	var inf, dat int
	for p := 0x20; p+8 <= len(data); {
		switch string(data[p : p+4]) {
		case "1FNI":
			inf = p
		case "1TAD":
			dat = p
		}
		size := int(le.Uint32(data[p+4:]))
		if size <= 0 { // the English file declares DAT1 one byte past EOF
			break
		}
		p += size
	}
	if inf == 0 || dat == 0 {
		return nil, fmt.Errorf("sm64ds: BMG INF1/DAT1 section missing")
	}
	count := int(le.Uint16(data[inf+8:]))
	entrySize := int(le.Uint16(data[inf+0xA:])) >> 3 // the getter's LSR #3
	pool := dat + 8
	msgs := make([]string, count)
	for i := 0; i < count; i++ {
		off := int(le.Uint32(data[inf+0x10+i*entrySize:]))
		msgs[i] = decodeMsg(data, pool+off)
	}
	return msgs, nil
}

// MsgIndex translates an EXTERNAL message ID — the number actors carry, e.g.
// a signpost placement's par1 — to the message's INF1 index. Game code never
// stores raw indices: the message-window code (overlay 7) passes every ID
// through $020B8EC0, which walks a {u16 firstID, u16 firstIndex} range table
// at ARM9 $0208EEEC and returns firstIndex + (id - firstID) for the range
// containing the ID (ranges are half-open, table ends at a sentinel ID
// >= 8000). Returns -1 for an ID before the first range.
func (ls *LevelSet) MsgIndex(id int) int {
	const msgRangeTable = 0x8AEEC // ARM9 file offset of $0208EEEC
	for k := 0; ; k++ {
		first := int(le.Uint16(ls.arm9[msgRangeTable+k*4:]))
		next := int(le.Uint16(ls.arm9[msgRangeTable+(k+1)*4:]))
		if id < first {
			return -1
		}
		if id < next || next >= 8000 {
			return int(le.Uint16(ls.arm9[msgRangeTable+k*4+2:])) + id - first
		}
	}
}

// decodeMsg converts one glyph-index string (0xFF-terminated) to text.
func decodeMsg(data []byte, p int) string {
	var out []rune
	for p < len(data) && data[p] != 0xFF {
		b := data[p]
		switch {
		case b == 0xFD:
			out = append(out, '\n')
			p++
		case b == 0xFE: // control: next byte = total length
			n := 2
			if p+1 < len(data) && int(data[p+1]) > 2 {
				n = int(data[p+1])
			}
			p += n
		case glyphChars[b] != 0:
			out = append(out, glyphChars[b])
			p++
		default:
			out = append(out, []rune(fmt.Sprintf("<%02X>", b))...)
			p++
		}
	}
	return string(out)
}
