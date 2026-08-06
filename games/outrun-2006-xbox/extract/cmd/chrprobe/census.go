package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

// censusPMT walks the disc for *_pmt.sz and prints {nParts, nTextures} of each.
func censusPMT(disc *xbox.Image, filter string) {
	disc.Walk(func(e xbox.Entry) error {
		low := strings.ToLower(e.Path)
		if e.IsDir || !strings.HasSuffix(low, "_pmt.sz") || !strings.Contains(low, filter) {
			return nil
		}
		raw, err := disc.ReadFile(e.Path)
		if err != nil {
			return nil
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil
		}
		data, err := io.ReadAll(zr)
		if err != nil || len(data) < 16 {
			return nil
		}
		fmt.Printf("%-55s parts=%3d tex=%3d\n", e.Path, u32(data, 0), u32(data, 4))
		return nil
	})
}
