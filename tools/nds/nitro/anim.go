package nitro

import (
	"fmt"
)

// BTA0 (NSBTA) — material texture-SRT animation: per material, five tracks
// {scaleS, scaleT, rotation, transS, transT} over a frame count, driving the
// texture matrix each frame (the DS's scrolling water, waterfalls and boost-panel
// arrows). Layout, pinned on the cartridge's 47 course BTA0s:
//
//	BTA0 header (NITRO container) → one SRT0 block
//	SRT0: dict of animations (named after the target model); entry u32 =
//	      animation offset relative to the SRT0 block start
//	animation: magic "M\0AT", u16 numFrames, u16 flags, then a dict of materials
//	      whose 0x28-byte entries are the five 8-byte component records inline:
//	      {u16 lastFrame, u8 0, u8 flags, u32 v}
//	component flags: bit7 = per-frame track (v = u16-sample array offset relative
//	      to the "M\0AT" record, sampled every 4th frame — verified: the beach
//	      dash's last sample 0xFBD == 4096*60/61 exactly); else v is a constant.
//	      Rotation components hold a (sin,cos) fx12 pair in v's two halves.
//	      Values are fx12 (1.0 = 0x1000) in normalized texture space.
type TexAnim struct {
	Model    string // target model (SRT0 dict name)
	Material string // target material within that model
	Frames   int
	ScaleS   Track
	ScaleT   Track
	RotSin   float64 // constant rotation only (no animated rotation on cartridge)
	RotCos   float64
	TransS   Track
	TransT   Track
}

// Track is one animated component: constant, or per-frame samples (one per
// sampleStep frames, linearly interpolated between).
type Track struct {
	Const   float64
	Samples []float64 // empty = constant
	Step    int       // frames per sample (4 on cartridge)
}

// At returns the track value at a (fractional) frame.
func (t Track) At(frame float64) float64 {
	if len(t.Samples) == 0 {
		return t.Const
	}
	pos := frame / float64(t.Step)
	i := int(pos)
	if i >= len(t.Samples)-1 {
		return t.Samples[len(t.Samples)-1]
	}
	f := pos - float64(i)
	return t.Samples[i]*(1-f) + t.Samples[i+1]*f
}

func fx12(v uint32) float64  { return float64(int32(v)) / 4096 }
func fx12u(v uint16) float64 { return float64(int16(v)) / 4096 }

// DecodeNSBTA decodes every material animation in a BTA0 file.
func DecodeNSBTA(data []byte) ([]TexAnim, error) {
	if len(data) < 0x14 || string(data[0:4]) != "BTA0" {
		return nil, fmt.Errorf("nitro: not a BTA0 file")
	}
	srt := int(le.Uint32(data[0x10:]))
	if srt+8 > len(data) || string(data[srt:srt+4]) != "SRT0" {
		return nil, fmt.Errorf("nitro: no SRT0 block")
	}
	anims, err := parseDict(data, srt+8)
	if err != nil {
		return nil, err
	}
	var out []TexAnim
	for _, a := range anims {
		mat := srt + int(le.Uint32(padded(a.data)))
		if mat+8 > len(data) || string(data[mat:mat+2]) != "M\x00" || string(data[mat+2:mat+4]) != "AT" {
			return nil, fmt.Errorf("nitro: animation %q: no M-AT record", a.name)
		}
		frames := int(le.Uint16(data[mat+4:]))
		mats, err := parseDict(data, mat+8)
		if err != nil {
			return nil, err
		}
		for _, m := range mats {
			if len(m.data) < 0x28 {
				return nil, fmt.Errorf("nitro: animation %q material %q: short record", a.name, m.name)
			}
			ta := TexAnim{Model: a.name, Material: m.name, Frames: frames}
			comp := func(i int) (Track, error) {
				o := i * 8
				last := int(le.Uint16(m.data[o:]))
				flags := m.data[o+3]
				v := le.Uint32(m.data[o+4:])
				if flags&0x80 == 0 {
					return Track{Const: fx12(v)}, nil
				}
				const step = 4
				n := last/step + 1
				ao := mat + int(v)
				if ao+2*n > len(data) {
					return Track{}, fmt.Errorf("nitro: animation %q material %q: track overruns file", a.name, m.name)
				}
				t := Track{Step: step, Samples: make([]float64, n)}
				for k := 0; k < n; k++ {
					t.Samples[k] = fx12u(le.Uint16(data[ao+2*k:]))
				}
				return t, nil
			}
			var errs [5]error
			ta.ScaleS, errs[0] = comp(0)
			ta.ScaleT, errs[1] = comp(1)
			// rotation: constant (sin,cos) fx12 pair
			rv := le.Uint32(m.data[2*8+4:])
			ta.RotSin, ta.RotCos = fx12u(uint16(rv)), fx12u(uint16(rv>>16))
			ta.TransS, errs[3] = comp(3)
			ta.TransT, errs[4] = comp(4)
			for _, e := range errs {
				if e != nil {
					return nil, e
				}
			}
			out = append(out, ta)
		}
	}
	return out, nil
}
