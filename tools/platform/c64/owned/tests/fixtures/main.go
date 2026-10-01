// Command fixtures creates local comparison data using the existing independent
// Go tape decoder/graphics extractor. None of its bytes are injected into the CPU.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/fort-apocalypse-c64/extract/fastload"
	"retroreverse.com/games/fort-apocalypse-c64/extract/fortgfx"
	"retroreverse.com/tools/platform/c64/cbmtape"
	"retroreverse.com/tools/platform/c64/tap"
)

func main() {
	image := flag.String("image", "games/fort-apocalypse-c64/Fort_Apocalypse.tap", "local reference TAP")
	out := flag.String("out", "", "scratch output directory (required)")
	flag.Parse()
	if err := run(*image, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(image, out string) error {
	if out == "" {
		return fmt.Errorf("supply -out with a local scratch directory")
	}
	raw, err := os.ReadFile(image)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "9e444c4576bac52ba691f0ffe2c0a7efb0f62f3fe2be7cbe78dba08672dda00b" {
		return fmt.Errorf("expected the reference U.S. Gold / SYNSOFT NOVALOAD D100701 TAP")
	}
	tape, err := tap.Parse(raw)
	if err != nil {
		return err
	}
	blocks := cbmtape.ScanBlocks(tape.Pulses)
	if len(blocks) != 4 {
		return fmt.Errorf("expected four KERNAL records")
	}
	for _, block := range blocks {
		if !block.ChecksumOK {
			return fmt.Errorf("KERNAL checksum failure")
		}
	}
	decoded := fastload.Decode(tape.Pulses, blocks[3].EndPulse)
	if decoded.Err != nil {
		return decoded.Err
	}
	if !decoded.Terminated || len(decoded.Records) != 84 {
		return fmt.Errorf("incomplete fastloader stream")
	}
	pages := make([]byte, len(decoded.Records))
	for i, record := range decoded.Records {
		if !record.ChecksumOK {
			return fmt.Errorf("page checksum failure")
		}
		pages[i] = record.Page
	}
	memory := make([]byte, 65536)
	for address, value := range decoded.Memory {
		memory[address] = value
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	program, err := os.CreateTemp(out, "game-*.prg")
	if err != nil {
		return err
	}
	defer os.Remove(program.Name())
	if _, err = program.Write(append([]byte{0, 0x70}, memory[0x7000:0xb900]...)); err != nil {
		program.Close()
		return err
	}
	if err := program.Close(); err != nil {
		return err
	}
	game, err := fortgfx.LoadGame(program.Name())
	if err != nil {
		return err
	}
	graphics, mask := make([]byte, 65536), make([]byte, 65536)
	put := func(address int, data []byte) {
		copy(graphics[address:], data)
		for i := range data {
			mask[address+i] = 1
		}
	}
	// Compare only immutable output: exclude live scanner, animation and SID-noise cells.
	put(0x500f, game.HUDCharset()[0xf:0x2e0])
	chars := game.PlayfieldCharset()
	for ch := 0x21; ch < 128; ch++ {
		if ch == 0x3f || ch == 0x47 || ch == 0x59 || ch == 0x5a || ch == 0x71 || ch == 0x72 || ch >= 0x7c || ch >= 0x4c && ch <= 0x4f {
			continue
		}
		put(0x5800+8*ch, chars[8*ch:8*ch+8])
	}
	for i, shape := range game.SpriteShapes() {
		put(0x4040+64*i, shape)
	}
	for _, entry := range []struct {
		name string
		data []byte
	}{{"expected.bin", memory}, {"pages.bin", pages}, {"graphics.bin", graphics}, {"graphics-mask.bin", mask}} {
		if err := os.WriteFile(filepath.Join(out, entry.name), entry.data, 0644); err != nil {
			return err
		}
		fmt.Printf("%s SHA256 %x\n", entry.name, sha256.Sum256(entry.data))
	}
	return nil
}
