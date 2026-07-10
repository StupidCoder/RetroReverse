// romextract dumps every chunk of the cartridge archive to a directory, one
// file per chunk, decompressing GZIP-wrapped payloads on the way out. It is the
// raw material for the per-format decoders — and, because it needs no machine,
// the proof that the assets are on the medium rather than produced by running it.
//
// Files are named <index>-<FORMTYPE>[-<n>]-<CHUNKTAG>.bin. The directory FORM is
// index -1 and comes out as "dir-UVRM-TABL.bin".
//
// Usage:
//
//	romextract -image ROM -o work/archive [-type UVMD]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image := flag.String("image", "", "cartridge image")
	out := flag.String("o", "", "output directory")
	typ := flag.String("type", "", "extract only resources of this FORM type")
	skipPad := flag.Bool("skippad", true, "do not write the PAD chunks")
	flag.Parse()

	if *image == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "romextract: -image and -o are required")
		os.Exit(2)
	}
	rom, err := n64.Load(*image)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}
	if err := a.Check(); err != nil {
		die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}

	var files, bytes int
	for _, f := range a.Forms {
		if *typ != "" && f.Type != *typ {
			continue
		}
		// Several chunks of a FORM can share a tag (PDAT's 24,401 RPKTs), so the
		// index within the FORM disambiguates.
		seen := map[string]int{}
		for _, c := range f.Chunks {
			if *skipPad && strings.TrimSpace(c.Tag) == "PAD" {
				continue
			}
			data, err := a.Data(c)
			if err != nil {
				die(err)
			}
			name := strings.TrimSpace(c.Tag)
			if c.Compressed() {
				name = strings.TrimSpace(c.InnerTag)
			}
			id := fmt.Sprintf("%04d", f.Index)
			if f.Index < 0 {
				id = "dir"
			}
			base := fmt.Sprintf("%s-%s-%s", id, f.Type, name)
			if n := seen[name]; n > 0 || countTag(f, name) > 1 {
				base = fmt.Sprintf("%s-%s-%03d-%s", id, f.Type, n, name)
			}
			seen[name]++
			if err := os.WriteFile(filepath.Join(*out, base+".bin"), data, 0o644); err != nil {
				die(err)
			}
			files++
			bytes += len(data)
		}
	}
	fmt.Printf("wrote %d chunks, %d bytes to %s\n", files, bytes, *out)
}

func countTag(f pwad.Form, name string) int {
	n := 0
	for _, c := range f.Chunks {
		t := strings.TrimSpace(c.Tag)
		if c.Compressed() {
			t = strings.TrimSpace(c.InnerTag)
		}
		if t == name {
			n++
		}
	}
	return n
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "romextract:", err)
	os.Exit(1)
}
