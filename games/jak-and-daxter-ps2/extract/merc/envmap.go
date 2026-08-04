package merc

// envmap.go: the flag-0x100 effects' extra-info block (effect record +28) —
// the second, additive draw the game layers over hair, armor, eyes and the
// title logo's electric glow.
//
// Read from draw-bones (0x6EBA74; the per-effect loop at 0x6EBFC8): the
// record's flag byte (+27, the 0x100 bit) gates the pass; the params row
// sits at extra + extra[0]*16 as {f32 slope, f32 base, u8 tint[4]}. The
// runtime intensity is clamp(slope*camDist + base, 0, 1.0) (constants at
// draw-bones fp+10600 = 1.0 ceiling, fp+10604 = 1/128), and the tint bytes
// scale by intensity x the level's env color ([s7-8136]+92, default
// {128,128,128} = x1.0) into *merc-bucket-info* + effect*8 + 124 — the
// per-effect vertex color of the additive pass (+128 = the router flag that
// sends the effect through the alpha bucket).
//
// When the header's second byte is 2, a 5-quadword adgif shader follows the
// params ({TEX0, TEX1, MIPTBP1, CLAMP, ALPHA} rows; the texture id in the
// TEX1 quadword's third word, the ALPHA data 0x58 = Cs*Ad + Cd, additive by
// destination alpha — the mask the base pass left in the framebuffer's
// alpha channel): a real environment map, sphere-mapped by the microcode
// (bam-eyelight for the eyes; the characters' shine texture on their own
// pages). With byte 2 = 0 there is no shader: the pass re-draws the
// fragment's own textures with plain additive blending (ALPHA 0x68, by the
// FIX constant) — the logo's electric glow.

import "encoding/binary"

// EnvSpec is one effect's parsed extra-info.
type EnvSpec struct {
	Slope, Base float32  // intensity = clamp(Slope*dist + Base, 0, 1)
	Tint        [4]uint8 // GS scale, 0x80 = 1.0
	RawID       uint32   // envmap shader texture id; 0 = plain additive pass
}

// ParseExtraInfo reads the extra-info block at object offset off (0 = none).
func ParseExtraInfo(obj []byte, off uint32) *EnvSpec {
	if off == 0 || int(off)+32 > len(obj) {
		return nil
	}
	p := off + uint32(obj[off])*16
	if int(p)+16 > len(obj) {
		return nil
	}
	s := &EnvSpec{
		Slope: float32frombits(binary.LittleEndian.Uint32(obj[p:])),
		Base:  float32frombits(binary.LittleEndian.Uint32(obj[p+4:])),
		Tint:  [4]uint8{obj[p+8], obj[p+9], obj[p+10], obj[p+11]},
	}
	// find the adgif rows (reg bytes {06,14,34,08,42} at +8 of each
	// quadword) anywhere in the item list
	for r := off + 16; int(r)+80 <= len(obj) && r < off+16*16; r += 16 {
		if obj[r+8] == 0x06 && obj[r+24] == 0x14 && obj[r+40] == 0x34 &&
			obj[r+56] == 0x08 && obj[r+72] == 0x42 {
			s.RawID = binary.LittleEndian.Uint32(obj[r+24:]) & 0xFFFFFF00
			break
		}
	}
	return s
}
