// missiontext decodes Elite's mission briefing messages from the recursive
// message-token table at $0E00. The text dispatcher print_dispatch ($238D)
// reads that table EOR-$57 obfuscated and $00-separated; a token number selects
// the n-th string, and inside a string the bytes are a mix of literal letters,
// nested token references, two-letter "digram" pairs (from the table at $254B)
// and control codes for dynamic inserts (the commander's name, etc.).
//
// This reproduces the four mission tokens the scripted-mission handlers print
// (Elite.md Part V §12): $0B, $C7, $DE, $DF. It runs on the reconstructed
// engine image (the code/data the loader hid under $D000-$EFFF is present as
// RAM here), so no external data is used — everything comes from the image.
package main

import (
	"fmt"
	"os"
	"strings"

	"elite/extract/shipmodel"
)

const (
	tableBase = 0x0E00 // message-token table (EOR-$57, $00-separated)
	tableEnd  = 0x1D00 // ends where the decrypted game code begins ($1D1F)
	sep       = 0x57   // stored separator byte (decodes to $00)
	digrams   = 0x2549 // digram pairs, indexed as $2549 + (byte-$D7)*2
)

// decoder expands message tokens against the engine image.
type decoder struct {
	mem    []byte
	pieces [][]byte
}

func newDecoder(mem []byte) *decoder {
	// Split the raw table on the separator byte; piece[n] is token n.
	return &decoder{mem: mem, pieces: splitRaw(mem[tableBase:tableEnd], sep)}
}

func splitRaw(b []byte, s byte) [][]byte {
	var out [][]byte
	cur := []byte{}
	for _, x := range b {
		if x == s {
			out = append(out, cur)
			cur = []byte{}
			continue
		}
		cur = append(cur, x)
	}
	return append(out, cur)
}

// digram returns the two-letter pair for an obfuscation-decoded byte >= $D7.
func (d *decoder) digram(c byte) string {
	x := int(c-0xD7) * 2
	var s strings.Builder
	if a := d.mem[digrams+x]; a >= 0x20 && a < 0x7f {
		s.WriteByte(a)
	}
	if b := d.mem[digrams+x+1]; b >= 0x20 && b < 0x7f && b != '?' {
		s.WriteByte(b)
	}
	return s.String()
}

// expand recursively expands token tok into readable text. Control codes for
// dynamic inserts become angle-bracket placeholders; pure formatting codes are
// dropped.
func (d *decoder) expand(tok int, depth int) string {
	if depth > 40 || tok < 0 || tok >= len(d.pieces) {
		return ""
	}
	var s strings.Builder
	for _, raw := range d.pieces[tok] {
		c := raw ^ sep
		switch {
		case c == 0:
			// separator inside a piece can't happen (we split on it)
		case c >= 0x20 && c <= 0x5A: // literal letter / digit / punctuation
			s.WriteByte(c)
		case c >= 0x5B && c <= 0x80: // random-synonym selector ($243E)
			s.WriteString("‹reward›")
		case c >= 0x81 && c <= 0xD6: // nested token
			s.WriteString(d.expand(int(c), depth+1))
		case c >= 0xD7: // digram pair
			s.WriteString(d.digram(c))
		case c == 0x04: // print_token $04 -> commander name
			s.WriteString("‹cmdr›")
		case c == 0x11 || c == 0x12 || c == 0x1B: // name_from_buf / digram_name / captain
			s.WriteString("‹name›")
		default: // formatting control code (clear, newline, colour, pause, beep)
		}
	}
	return s.String()
}

// clean collapses the runs of spaces left by dropped control codes.
func clean(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func main() {
	dir := "../extracted"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	mem, err := shipmodel.LoadEngine(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "missiontext:", err)
		os.Exit(1)
	}
	d := newDecoder(mem)

	tokens := []struct {
		tok  int
		note string
	}{
		{0x0B, "mission 1 invitation (handler $3D7A)"},
		{0xC7, "trade offer (handler $3DBD)"},
		{0xDE, "mission 2 briefing (handler $3D8A)"},
		{0xDF, "mission complete (handler $3D98)"},
	}
	for _, t := range tokens {
		fmt.Printf("token $%02X — %s:\n  %s\n\n", t.tok, t.note, clean(d.expand(t.tok, 0)))
	}
}
