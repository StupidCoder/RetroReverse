// textdump extracts the game's dialogue straight from the cartridge.
//
// The script is stored as plain ASCII — no encoding, no compression — in
// NUL-terminated runs across the back of the ROM. A message uses 0x0A for a
// line break and carries inline control codes introduced by a byte below 0x10
// (speaker and portrait selection, colour, pauses, and the branch points a
// question uses), which are shown here as [xx yy] rather than dropped.
//
// This dumps the SCRIPT, not the mapping from a sign or a character to their
// line: what selects a message is a separate index that has not been decoded
// yet, so the messages come out in ROM order rather than by who says them.
//
//	textdump [-min N] [-out FILE] [-json FILE]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// scriptFloor is where the script region begins; everything below it is code
// and graphics, and admitting it produces thousands of false messages.
const scriptFloor = 0x08900000

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "textdump: "+format+"\n", a...)
	os.Exit(2)
}

// Message is one NUL-terminated string, with its cartridge address.
type Message struct {
	Addr string `json:"addr"`
	Text string `json:"text"`
}

// printable reports whether a byte can appear inside a message: ASCII, the
// newline, the NUL terminator, or a control-code introducer.
func printable(c byte) bool {
	return c == 0 || c == 0x0A || (c >= 0x20 && c < 0x7F) || c < 0x10
}

// decode renders one message, keeping control codes visible.
//
// Control codes do NOT all take an operand, and assuming they do eats the
// following letter: "[0E 59]ou got some..." is 0x0E with no operand followed by
// the 'Y' of "You". Only the codes observed to carry one (0x01-0x03, which
// select speaker, highlight colour and portrait) consume the next byte; the
// rest are shown bare rather than guessed at.
func decode(b []byte) string {
	var sb strings.Builder
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch {
		case c == 0x0A:
			sb.WriteByte('\n')
		case c >= 0x20 && c < 0x7F:
			sb.WriteByte(c)
		case c >= 0x01 && c <= 0x03 && i+1 < len(b):
			fmt.Fprintf(&sb, "[%02X %02X]", c, b[i+1])
			i++
		case c < 0x10:
			fmt.Fprintf(&sb, "[%02X]", c)
		}
	}
	return sb.String()
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	min := flag.Int("min", 8, "shortest message to keep, in printable characters")
	out := flag.String("out", "", "write the script as text to this file (default: stdout summary only)")
	jsonOut := flag.String("json", "", "write the messages as JSON to this file")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}

	var msgs []Message
	i := 0
	for i < len(rom) {
		if !printable(rom[i]) || rom[i] == 0 {
			i++
			continue
		}
		start := i
		for i < len(rom) && printable(rom[i]) && rom[i] != 0 {
			i++
		}
		if i >= len(rom) || rom[i] != 0 {
			continue // a run must END in a NUL to be a message
		}
		raw := rom[start:i]
		i++ // step over the terminator

		// Keep it only if it reads as PROSE. Compressed data and tables are full
		// of runs that decode as printable ASCII, so length alone finds ten
		// thousand "messages" of which most are binary: the script lives in the
		// back of the cartridge, real lines contain spaces, and their letters
		// dominate. All three together are what separates a sentence from a
		// stretch of graphics data that happens to fall in the ASCII range.
		if 0x08000000+start < scriptFloor {
			continue
		}
		letters, printablen, spaces := 0, 0, 0
		for _, c := range raw {
			if c >= 0x20 && c < 0x7F {
				printablen++
				switch {
				case c == ' ':
					spaces++
				case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
					letters++
				}
			}
		}
		if printablen < *min || spaces == 0 || letters*10 < printablen*7 {
			continue
		}
		msgs = append(msgs, Message{
			Addr: fmt.Sprintf("%08X", 0x08000000+start),
			Text: decode(raw),
		})
	}

	fmt.Printf("%d messages\n", len(msgs))
	if len(msgs) > 0 {
		fmt.Println("first:", strings.ReplaceAll(msgs[0].Text, "\n", " / "))
		fmt.Println("last: ", strings.ReplaceAll(msgs[len(msgs)-1].Text, "\n", " / "))
	}
	if *out != "" {
		var sb strings.Builder
		for _, m := range msgs {
			fmt.Fprintf(&sb, "%s\n%s\n\n", m.Addr, m.Text)
		}
		if err := os.WriteFile(*out, []byte(sb.String()), 0o644); err != nil {
			die("%v", err)
		}
		fmt.Println("wrote", *out)
	}
	if *jsonOut != "" {
		b, _ := json.MarshalIndent(msgs, "", " ")
		if err := os.WriteFile(*jsonOut, b, 0o644); err != nil {
			die("%v", err)
		}
		fmt.Println("wrote", *jsonOut)
	}
}
