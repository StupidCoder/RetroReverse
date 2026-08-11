package bchglb

import (
	"fmt"

	"retroreverse.com/tools/lib/glb"
	"retroreverse.com/tools/platform/n3ds"
)

// AnimFPS is the rate a clip's frames are played at.
//
// It is measured, not assumed. The oracle's per-draw uniform dump shows the
// matrix palette Captain Toad's opening stage hands its vertex shader changing
// on every *second* presented frame — six consecutive rendered frames give
// three distinct poses — so the engine presents at the 3DS's 60 Hz and advances
// its animations at half that. It agrees with the rate the banner's CGFX
// animations are authored at, which is the same NW4C toolchain.
const AnimFPS = 30

// AddClip writes one skeletal animation over a rig as a named glTF animation.
//
// The channels are sampled at every whole frame rather than passed through as
// curves. That is not an approximation of what the game does — it is what the
// game does: it advances the clip one frame at a time and poses the skeleton
// from the value there, so straight lines between whole frames pass through
// every pose the console ever shows. Carrying the hermite tangents through
// instead would be exact for the translations and wrong for the rotations,
// which are stored as three independent Euler curves and have to become
// quaternions before a renderer can slerp them.
func AddClip(s *glb.Scene, rig *Rig, a *n3ds.SkelAnim, name string) error {
	if len(rig.Nodes) == 0 {
		return fmt.Errorf("bchglb: clip %q has no rig to play on", name)
	}
	byName := map[string]int{}
	for i, b := range rig.model.Bones {
		byName[b.Name] = i
	}

	last := int(a.Frames)
	if last < 0 {
		return fmt.Errorf("bchglb: clip %q runs to frame %g", name, a.Frames)
	}
	clip := s.NewClip(name)
	for _, m := range a.Members {
		bi, ok := byName[m.Bone]
		if !ok {
			// An animation archive is shared between several models — Toad's
			// holds Toadette's clips too — so a member naming a bone this
			// skeleton does not have is normal, and is skipped rather than
			// treated as an error.
			continue
		}
		bone := rig.model.Bones[bi]
		node := rig.Nodes[bi]

		for _, grp := range []struct {
			first int
			path  string
		}{
			{n3ds.SlotScaleX, "scale"},
			{n3ds.SlotRotX, "rotation"},
			{n3ds.SlotTransX, "translation"},
		} {
			// A group is emitted only if the clip says something about it; the
			// bone otherwise keeps the transform its skeleton node already has.
			moving, present := false, false
			for k := 0; k < 3; k++ {
				c := m.Curves[grp.first+k]
				if c == nil {
					continue
				}
				present = true
				if len(c.Keys) > 1 {
					moving = true
				}
			}
			if !present {
				continue
			}
			frames := []int{0}
			if moving {
				frames = make([]int, last+1)
				for i := range frames {
					frames[i] = i
				}
			}
			times := make([]float32, len(frames))
			for i, fr := range frames {
				times[i] = float32(fr) / AnimFPS
			}

			// The bind value is the fallback for a slot the clip leaves alone,
			// so a group with one animated component still moves only that one.
			base := [3]float32{}
			switch grp.first {
			case n3ds.SlotScaleX:
				base = bone.Scale
			case n3ds.SlotRotX:
				base = bone.Rotate
			default:
				base = bone.Trans
			}
			vals := make([][3]float32, len(frames))
			for i, fr := range frames {
				v := base
				for k := 0; k < 3; k++ {
					if c := m.Curves[grp.first+k]; c != nil {
						v[k] = c.Eval(float32(fr))
					}
				}
				vals[i] = v
			}
			if grp.path == "rotation" {
				quats := make([][4]float32, len(vals))
				for i, v := range vals {
					quats[i] = quatZYX(v)
				}
				// Keep the quaternion path continuous: a renderer slerps along
				// the shorter arc, and a sign flip between two samples of a
				// smoothly turning bone sends it the long way round.
				for i := 1; i < len(quats); i++ {
					var dot float32
					for k := 0; k < 4; k++ {
						dot += quats[i-1][k] * quats[i][k]
					}
					if dot < 0 {
						for k := 0; k < 4; k++ {
							quats[i][k] = -quats[i][k]
						}
					}
				}
				clip.Rotations(node, times, quats)
				continue
			}
			clip.Vec3s(node, grp.path, times, vals)
		}
	}
	clip.Finish()
	return nil
}
