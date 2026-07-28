package threedo

// snds.go demuxes and decodes the audio track of a 3DO .Stream movie: SNDS
// chunks carrying an SHDR sound header plus SSMP sample payloads, compressed
// with SDX2 — the SDK DataStreamer's "squareroot-delta-exact" DPCM (one byte
// per sample; the DSP undoes it on real hardware). Like the Cinepak side
// (cvid.go) this is a platform codec reimplemented from the format, and the
// decoder is verified byte-identical to an independent reference (FFmpeg's
// sdx2_dpcm) across the Need for Speed movies.
//
// Chunk layout (all big-endian, sizes include the 8-byte tag+size header):
//
//	SNDS/SHDR  time(4) chan(4) 'SHDR' ver(4) ?(4) ?(4) ?(4) ?(4)
//	           bitsPerSample(4) sampleRate(4) channels(4) fourcc ('SDX2')
//	SNDS/SSMP  time(4) chan(4) 'SSMP' byteCount(4) then byteCount SDX2 bytes,
//	           frames interleaved L R for stereo, one byte per sample.

import (
	"encoding/binary"
	"fmt"
)

// SndsTrack is a demuxed .Stream audio track: the SHDR parameters and the
// concatenated SDX2 payload bytes of every SSMP chunk, in stream order.
type SndsTrack struct {
	SampleRate int
	Channels   int
	Bits       int    // decoded sample width from the SHDR (16)
	Codec      string // SHDR compression fourcc; "SDX2" on every disc movie
	Data       []byte // raw SDX2 bytes (one per sample, channel-interleaved)
}

// DemuxSnds walks a .Stream file and pulls out its SNDS audio track. A stream
// with no SNDS chunks returns (nil, nil) — silent movies exist and are not an
// error.
func DemuxSnds(data []byte) (*SndsTrack, error) {
	var t *SndsTrack
	for off := 0; off+8 <= len(data); {
		tag := string(data[off : off+4])
		size := int(binary.BigEndian.Uint32(data[off+4 : off+8]))
		if size < 8 || off+size > len(data) {
			break // truncated or misaligned; stop at the last good chunk
		}
		if tag == "SNDS" && off+20 <= len(data) {
			switch string(data[off+16 : off+20]) {
			case "SHDR":
				if off+56 <= off+size {
					t = &SndsTrack{
						Bits:       int(binary.BigEndian.Uint32(data[off+40 : off+44])),
						SampleRate: int(binary.BigEndian.Uint32(data[off+44 : off+48])),
						Channels:   int(binary.BigEndian.Uint32(data[off+48 : off+52])),
						Codec:      string(data[off+52 : off+56]),
					}
				}
			case "SSMP":
				if t == nil {
					return nil, fmt.Errorf("SSMP before SHDR at 0x%x", off)
				}
				if off+24 <= off+size {
					n := int(binary.BigEndian.Uint32(data[off+20 : off+24]))
					end := off + 24 + n
					if end > off+size || end > len(data) {
						end = off + size
					}
					t.Data = append(t.Data, data[off+24:end]...)
				}
			}
		}
		off += size
	}
	if t != nil && t.Codec != "SDX2" {
		return nil, fmt.Errorf("unsupported audio codec %q", t.Codec)
	}
	return t, nil
}

// DecodeSDX2 expands the track to interleaved 16-bit PCM. Each byte b holds
// one sample: 2·b·|b| (a sign-preserving square, which the name's "squareroot"
// points back at) — taken absolutely when b is even ("exact"), added to the
// channel's previous sample when b is odd ("delta").
func (t *SndsTrack) DecodeSDX2() []int16 {
	ch := t.Channels
	if ch < 1 {
		ch = 1
	}
	prev := make([]int32, ch)
	out := make([]int16, len(t.Data))
	for i, u := range t.Data {
		b := int32(int8(u))
		s := 2 * b * b
		if b < 0 {
			s = -s
		}
		if b&1 != 0 {
			s += prev[i%ch]
		}
		if s > 32767 {
			s = 32767
		} else if s < -32768 {
			s = -32768
		}
		prev[i%ch] = s
		out[i] = int16(s)
	}
	return out
}
