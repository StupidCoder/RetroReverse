// Package uvlv decodes Pilotwings 64's scene manifests.
//
// The archive holds exactly one UVLV resource, carrying 136 uncompressed COMM
// chunks. Each chunk is one scene, and a scene is nothing but a list of what the
// game must have resident to run it. Its parser is at 0x80226B20, and the shape
// it reads is ten counted arrays back to back:
//
//	Slot[10]:  u16 count, u16[count]
//
// The values are **ordinals within a FORM type**, not resource indices: slot 4's
// `2` means the third UVCT resource, not resource 2. Which slot names which type
// is not written down anywhere, and was not guessed:
//
//   - Slot 8 is 0..8 in all 136 chunks. Nine values, always resident, and the
//     archive holds exactly nine UVFT resources.
//   - Slot 0 names UVTR worlds, and slot 4 UVCT terrain chunks. The 13 chunks
//     that use slot 0 pin both at once: take the union of the UVCT ordinals that
//     those worlds' grid cells name, and it equals slot 4 exactly, thirteen times
//     out of thirteen. Two independently-decoded structures agreeing on which
//     terrain a scene needs is not a coincidence a wrong reading survives.
//   - Slot 3 names UVMD models. An object placed in a UVCT chunk selects its
//     model by ordinal (see package uvct), so every object type in a scene's
//     terrain must appear in that scene's model list. All 13 do, covering all
//     183 types the game uses.
//   - Slot 5 names UVTX textures, by the same argument the DMA log makes: the
//     resources the running game fetches for a scene are exactly slots 0/3/4/5/8
//     plus the mission's own vehicle and HUD assets.
//
// Slot 1 is empty in every chunk; slots 2, 6, 7 and 9 are used but unidentified,
// and are carried through rather than named. Six of the 13 terrain scenes share
// world 9's single cell and three share world 3's terrain with different models
// and textures: the game's map variations are scene manifests over shared
// terrain, not duplicated maps.
package uvlv

import (
	"encoding/binary"
	"fmt"
)

// Slots is the number of counted arrays in every scene. The parser is unrolled,
// not a loop, and it reads ten of them into a resident struct of ten
// (pointer, count) pairs: the last `addiu $a0, $s0, 76` before its `jr $ra` is
// the tenth. Counting them mattered — nine parses without error, because slot 9
// is empty in 107 of the 136 scenes and its zero count is indistinguishable from
// the alignment padding that follows.
const Slots = 10

// Named slots, for callers that want to say what they mean.
const (
	SlotWorld   = 0 // UVTR ordinals
	SlotModel   = 3 // UVMD ordinals
	SlotTerrain = 4 // UVCT ordinals
	SlotTexture = 5 // UVTX ordinals
	SlotFont    = 8 // UVFT ordinals, always 0..8
)

// A Scene is one UVLV chunk: nine lists of resource ordinals.
type Scene struct {
	Slot [Slots][]uint16

	// Padding is the number of bytes left after the tenth array: zero, two, four
	// or six, and always zero-valued. A scene that ends anywhere else is a misparse.
	Padding int
}

// Worlds, Models, Terrain, Textures and Fonts name the slots this package has
// identified. A scene with no terrain returns nil.
func (s *Scene) Worlds() []uint16   { return s.Slot[SlotWorld] }
func (s *Scene) Models() []uint16   { return s.Slot[SlotModel] }
func (s *Scene) Terrain() []uint16  { return s.Slot[SlotTerrain] }
func (s *Scene) Textures() []uint16 { return s.Slot[SlotTexture] }
func (s *Scene) Fonts() []uint16    { return s.Slot[SlotFont] }

// Decode reads one scene out of a UVLV COMM chunk.
func Decode(data []byte) (*Scene, error) {
	var s Scene
	p := 0
	for i := 0; i < Slots; i++ {
		if p+2 > len(data) {
			return nil, fmt.Errorf("uvlv: slot %d: chunk ends after %d of %d bytes", i, p, len(data))
		}
		n := int(binary.BigEndian.Uint16(data[p:]))
		p += 2
		if p+2*n > len(data) {
			return nil, fmt.Errorf("uvlv: slot %d: %d entries overrun the chunk", i, n)
		}
		if n > 0 {
			s.Slot[i] = make([]uint16, n)
			for j := range s.Slot[i] {
				s.Slot[i][j] = binary.BigEndian.Uint16(data[p+2*j:])
			}
		}
		p += 2 * n
	}
	for _, b := range data[p:] {
		if b != 0 {
			return nil, fmt.Errorf("uvlv: %d trailing bytes and not all zero", len(data)-p)
		}
	}
	s.Padding = len(data) - p
	return &s, nil
}
