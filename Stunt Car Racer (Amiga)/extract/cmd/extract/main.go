// extract pulls the program blobs off the Stunt Car Racer disk. The disk has no
// AmigaDOS filesystem (Stunt_Car_Racer.md Part I): it is read by *logical sector*
// (512 bytes; sector N = byte N*512 in the .adf), and the boot block / loader name
// fixed sector ranges. Both blobs are stored raw — there is no compression — so
// extracting them is a straight slice of the image.
//
//   loader : sectors 22..97   (offset $2C00, $9800 bytes)  — the custom track loader
//            the boot block reads and JMPs to.
//   game   : sectors 110..914 (offset $DC00, 805 sectors)  — the whole engine, which
//            the loader reads to $E700 and enters there.
//
// The engine init ($ED56, Stunt_Car_Racer.md Part III) XOR-$80 decrypts the run-time
// range $F4B8..$1AA4A in place before the main entry. We reproduce that here and also
// emit game.dec.bin, so the encrypted region disassembles to real code/data statically.
//
// Usage: extract disk.adf [-out dir]   (defaults to ./extracted, beside the .adf)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const sector = 512

// region is one named blob on the disk, given as a sector range and its run-time
// load address (the address the loader/boot places it at — used as the disasm base).
type region struct {
	name    string
	start   int    // first sector
	count   int    // sector count
	loadAt  uint32 // run-time load address
	desc    string
}

var regions = []region{
	{"loader.bin", 22, 76, 0, "custom track loader (boot reads to AllocMem'd chip RAM, JMPs to it)"},
	{"game.bin", 110, 805, 0xE700, "the game engine + data (loader reads to $E700, entry $E700)"},
}

// The engine's anti-tamper pass ($ED56): EOR.b #$80 over the run-time range
// [encStart,encEnd) of the $E700-based game image. We emit a decrypted copy so the
// region disassembles statically.
const (
	gameBase = 0xE700
	encStart = 0xF4B8
	encEnd   = 0x1AA4A
)

func main() {
	out := flag.String("out", "", "output directory (default: <adf dir>/extracted)")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: extract disk.adf [-out dir]")
		os.Exit(2)
	}
	adf, err := os.ReadFile(flag.Arg(0))
	must(err)

	dir := *out
	if dir == "" {
		dir = filepath.Join(filepath.Dir(flag.Arg(0)), "extracted")
	}
	must(os.MkdirAll(dir, 0o755))

	for _, r := range regions {
		off := r.start * sector
		n := r.count * sector
		if off+n > len(adf) {
			fmt.Fprintf(os.Stderr, "extract: %s region [%d..%d] past end of image\n", r.name, r.start, r.start+r.count)
			os.Exit(1)
		}
		p := filepath.Join(dir, r.name)
		blob := adf[off : off+n]
		must(os.WriteFile(p, blob, 0o644))
		fmt.Printf("%-10s sectors %d..%d  offset $%X  %d bytes  load $%X  — %s\n",
			r.name, r.start, r.start+r.count-1, off, n, r.loadAt, r.desc)

		// Emit a decrypted copy of the game image (the engine's $ED56 EOR-$80 pass).
		if r.loadAt == gameBase {
			dec := make([]byte, len(blob))
			copy(dec, blob)
			for a := encStart; a < encEnd; a++ {
				dec[a-gameBase] ^= 0x80
			}
			dp := filepath.Join(dir, "game.dec.bin")
			must(os.WriteFile(dp, dec, 0o644))
			fmt.Printf("%-10s                          %d bytes  load $%X  — EOR-$80 decrypt of $%X..$%X (engine $ED56)\n",
				"game.dec.bin", len(dec), r.loadAt, encStart, encEnd)
		}
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "extract:", err)
		os.Exit(1)
	}
}
