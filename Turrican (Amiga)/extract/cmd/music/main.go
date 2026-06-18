// Command music extracts Turrican's TFMX music data from the disk — the first
// stage of the audio pipeline (Part V). Turrican's score is Chris Hülsbeck's, in
// his own TFMX format (Turrican is *the* canonical TFMX game), played by the
// in-game sound driver in the $1BB00 sound overlay (streamed from ADF $26000 at
// game_init). The driver's api_init ($1CB62) is handed two pointers, which
// game_init fills with:
//
//	mdat  $1CFF4  the "TFMX-SONG" song/pattern/macro data
//	smpl  $20E90  raw 8-bit signed PCM sample data (to the overlay's end)
//
// The mdat header carries the song table (start/end/tempo of each sub-song at
// +$100/+$140/+$180, indexed by song number); the pattern/macro pointer tables and
// their data follow (the standard TFMX header pointers are zeroed — this build
// packs the data after a gap, the first pointer table starting at +$402). This
// command writes mdat.bin + smpl.bin and prints the song table; the player (next
// stage) reimplements the driver over these to render PCM.
//
// Usage: music [-o dir] [Turrican.adf]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"turrican/extract/decrunch"
)

const (
	soundOff  = 0x26000 // ADF offset of the packed sound overlay
	soundLen  = 0xC268
	soundBase = 0x1BB00 // its runtime load address
	mdatAddr  = 0x1CFF4 // api_init d0 (song/pattern/macro data)
	smplAddr  = 0x20E90 // api_init d1 (sample data)
)

func main() {
	out := flag.String("o", "rendered/music", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}
	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fail(err)
	}
	overlay, err := decrunch.DecrunchBlock(adf[soundOff : soundOff+soundLen])
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	mdat := overlay[mdatAddr-soundBase : smplAddr-soundBase]
	smpl := overlay[smplAddr-soundBase:]
	if err := os.WriteFile(filepath.Join(*out, "mdat.bin"), mdat, 0o644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "smpl.bin"), smpl, 0o644); err != nil {
		fail(err)
	}

	be16 := func(o int) int { return int(binary.BigEndian.Uint16(mdat[o:])) }
	fmt.Printf("mdat: %d bytes (TFMX-SONG)\nsmpl: %d bytes (8-bit PCM)\n", len(mdat), len(smpl))
	fmt.Println("sub-songs (trackstep start/end, tempo):")
	for i := 0; i < 32; i++ {
		s, e, t := be16(0x100+i*2), be16(0x140+i*2), be16(0x180+i*2)
		if s == 0 && e == 0 && t == 0 {
			continue
		}
		fmt.Printf("  %2d: start=$%04X end=$%04X tempo=$%04X\n", i, s, e, t)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "music:", err)
	os.Exit(1)
}
