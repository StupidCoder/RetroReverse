package main

// sdx2verify: -sdx2 <out.raw> decodes the -chunks stream's SNDS track with our
// SDX2 decoder and writes little-endian s16 PCM for a byte-diff against
// ffmpeg's independent sdx2_dpcm decode of the same stream.

import (
	"encoding/binary"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func sdx2ToFile(raw []byte, out string) error {
	t, err := threedo.DemuxSnds(raw)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("no SNDS track")
	}
	pcm := t.DecodeSDX2()
	buf := make([]byte, len(pcm)*2)
	for i, s := range pcm {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}
	fmt.Fprintf(os.Stderr, "SNDS %s %d Hz %dch %d bits, %d samples\n", t.Codec, t.SampleRate, t.Channels, t.Bits, len(pcm))
	return os.WriteFile(out, buf, 0o644)
}
