package sdat

import "fmt"

// Instrument types (SBNK record type byte).
const (
	InstNone     = 0
	InstPCM      = 1  // sampled wave from a SWAR
	InstPSGPulse = 2  // GB-style pulse channel; "swav" field = duty (0-7)
	InstPSGNoise = 3  // LFSR noise channel
	InstDrums    = 16 // per-key instrument map over a key range
	InstSplit    = 17 // eight key regions, each its own instrument
)

// Region is a playable instrument definition: the wave (or PSG setting) plus
// its envelope, for some range of keys.
type Region struct {
	Type     int // InstPCM, InstPSGPulse, InstPSGNoise
	Swav     int // wave index (PCM) or duty (pulse)
	Swar     int // index into the bank's four SWAR slots
	BaseNote int
	Attack   int
	Decay    int
	Sustain  int
	Release  int
	Pan      int
}

// Instrument is one SBNK program: simple types have a single region; drumsets
// and keysplits select a region by key.
type Instrument struct {
	Type    int
	Region  Region  // simple types
	Lo, Hi  int     // drumset key range
	Keys    []byte  // split upper bounds (8, for InstSplit)
	Regions []Region
}

// Bank is a decoded SBNK.
type Bank struct{ Instruments []Instrument }

func region(rec []byte, typ int) Region {
	return Region{
		Type:     typ,
		Swav:     int(le.Uint16(rec)),
		Swar:     int(le.Uint16(rec[2:])),
		BaseNote: int(rec[4]),
		Attack:   int(rec[5]),
		Decay:    int(rec[6]),
		Sustain:  int(rec[7]),
		Release:  int(rec[8]),
		Pan:      int(rec[9]),
	}
}

// ParseSBNK decodes an instrument bank.
func ParseSBNK(data []byte) (*Bank, error) {
	if len(data) < 0x3C || string(data[:4]) != "SBNK" {
		return nil, fmt.Errorf("sdat: not an SBNK file")
	}
	n := int(le.Uint32(data[0x38:]))
	b := &Bank{}
	ok := func(off, need int) bool { return off >= 0 && off+need <= len(data) }
	for i := 0; i < n; i++ {
		o := 0x3C + i*4
		if !ok(o, 4) {
			break
		}
		typ := int(data[o])
		off := int(le.Uint16(data[o+1:]))
		inst := Instrument{Type: typ}
		switch typ {
		case InstPCM, InstPSGPulse, InstPSGNoise:
			if !ok(off, 10) {
				inst.Type = InstNone
				break
			}
			inst.Region = region(data[off:], typ)
		case InstDrums:
			if !ok(off, 2) {
				inst.Type = InstNone
				break
			}
			lo, hi := int(data[off]), int(data[off+1])
			inst.Lo, inst.Hi = lo, hi
			for k := 0; k <= hi-lo; k++ {
				ro := off + 2 + k*12
				if !ok(ro, 12) {
					inst.Hi = lo + k - 1
					break
				}
				inst.Regions = append(inst.Regions, region(data[ro+2:], int(le.Uint16(data[ro:]))))
			}
		case InstSplit:
			if !ok(off, 8+8*12) {
				inst.Type = InstNone
				break
			}
			inst.Keys = data[off : off+8]
			for k := 0; k < 8; k++ {
				ro := off + 8 + k*12
				inst.Regions = append(inst.Regions, region(data[ro+2:], int(le.Uint16(data[ro:]))))
			}
		}
		b.Instruments = append(b.Instruments, inst)
	}
	return b, nil
}

// RegionFor returns the region that plays key, or nil.
func (in *Instrument) RegionFor(key int) *Region {
	switch in.Type {
	case InstPCM, InstPSGPulse, InstPSGNoise:
		return &in.Region
	case InstDrums:
		if key < in.Lo || key > in.Hi {
			return nil
		}
		return &in.Regions[key-in.Lo]
	case InstSplit:
		for k := 0; k < 8; k++ {
			if key <= int(in.Keys[k]) {
				return &in.Regions[k]
			}
		}
	}
	return nil
}
