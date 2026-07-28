package main

// -sums: for every movie on the disc print the SSMP payload capacity sum
// (chunk size - 24, what ffmpeg decodes) and the declared byteCount sum (what
// our demuxer honours) — the padding-invariant check for the SDX2 verify.

import (
	"encoding/binary"
	"fmt"

	"retroreverse.com/tools/platform/threedo"
)

func sums(data []byte) (cap, decl int) {
	for off := 0; off+8 <= len(data); {
		size := int(binary.BigEndian.Uint32(data[off+4 : off+8]))
		if size < 8 || off+size > len(data) {
			break
		}
		if string(data[off:off+4]) == "SNDS" && off+24 <= len(data) &&
			string(data[off+16:off+20]) == "SSMP" {
			cap += size - 24
			decl += int(binary.BigEndian.Uint32(data[off+20 : off+24]))
		}
		off += size
	}
	return
}

func printSums(vol interface {
	ReadFile(string) ([]byte, error)
}, paths []string) {
	for _, p := range paths {
		raw, err := vol.ReadFile(p)
		if err != nil {
			fmt.Printf("ERR %s %v\n", p, err)
			continue
		}
		c, d := sums(raw)
		fps, hdr := 0, 0
		if mv, err := threedo.DemuxStream(raw); err == nil {
			fps, hdr = mv.FPS, mv.HeaderRate
		}
		fmt.Printf("%s cap=%d decl=%d fps=%d hdr=%d\n", p, c, d, fps, hdr)
	}
}
