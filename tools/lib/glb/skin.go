package glb

// skin.go: glTF skinning and named per-joint animation clips for Scene.
// Single-influence skinning (JOINTS_0 with weight 1) is what fixed-palette
// console renderers (PS2 merc, etc.) map to.

import (
	"encoding/binary"
	"math"
)

// AddSkin registers a skin over existing nodes and returns its index. ibm
// is one inverse-bind matrix per joint node (column-major, glTF order).
func (s *Scene) AddSkin(jointNodes []int, ibm [][16]float32) int {
	buf := make([]byte, 64*len(ibm))
	for i, m := range ibm {
		for k := 0; k < 16; k++ {
			binary.LittleEndian.PutUint32(buf[i*64+k*4:], math.Float32bits(m[k]))
		}
	}
	vi := s.b.addView(buf)
	s.b.accessors = append(s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(ibm), Type: "MAT4",
	})
	s.skins = append(s.skins, map[string]any{
		"joints": jointNodes, "inverseBindMatrices": len(s.b.accessors) - 1,
	})
	return len(s.skins) - 1
}

// SetNodeSkin attaches a skin to a mesh node.
func (s *Scene) SetNodeSkin(node, skin int) {
	s.nodes[node]["skin"] = skin
}

// Clip is a named animation being assembled.
type Clip struct {
	s        *Scene
	name     string
	channels []map[string]any
	samplers []map[string]any
}

// NewClip starts a named animation clip.
func (s *Scene) NewClip(name string) *Clip {
	return &Clip{s: s, name: name}
}

func (c *Clip) times(times []float32) int {
	buf := make([]byte, 4*len(times))
	mn, mx := float32(math.Inf(1)), float32(math.Inf(-1))
	for i, t := range times {
		putF32(buf[i*4:], t)
		if t < mn {
			mn = t
		}
		if t > mx {
			mx = t
		}
	}
	vi := c.s.b.addView(buf)
	c.s.b.accessors = append(c.s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(times), Type: "SCALAR",
		Min: []float64{float64(mn)}, Max: []float64{float64(mx)},
	})
	return len(c.s.b.accessors) - 1
}

func (c *Clip) channel(node int, path string, tAcc, vAcc int) {
	c.samplers = append(c.samplers, map[string]any{
		"input": tAcc, "output": vAcc, "interpolation": "LINEAR",
	})
	c.channels = append(c.channels, map[string]any{
		"sampler": len(c.samplers) - 1,
		"target":  map[string]any{"node": node, "path": path},
	})
}

// Rotations adds a LINEAR quaternion channel on a node.
func (c *Clip) Rotations(node int, times []float32, quats [][4]float32) {
	buf := make([]byte, 16*len(quats))
	for i, q := range quats {
		for k := 0; k < 4; k++ {
			putF32(buf[i*16+k*4:], q[k])
		}
	}
	vi := c.s.b.addView(buf)
	c.s.b.accessors = append(c.s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(quats), Type: "VEC4",
	})
	vAcc := len(c.s.b.accessors) - 1
	c.channel(node, "rotation", c.times(times), vAcc)
}

// Vec3s adds a LINEAR vec3 channel ("translation" or "scale") on a node.
func (c *Clip) Vec3s(node int, path string, times []float32, vals [][3]float32) {
	buf := make([]byte, 12*len(vals))
	for i, v := range vals {
		for k := 0; k < 3; k++ {
			putF32(buf[i*12+k*4:], v[k])
		}
	}
	vi := c.s.b.addView(buf)
	c.s.b.accessors = append(c.s.b.accessors, accessor{
		BufferView: vi, ComponentType: 5126, Count: len(vals), Type: "VEC3",
	})
	vAcc := len(c.s.b.accessors) - 1
	c.channel(node, path, c.times(times), vAcc)
}

// Finish registers the clip as a named glTF animation.
func (c *Clip) Finish() {
	if len(c.channels) == 0 {
		return
	}
	c.s.clips = append(c.s.clips, map[string]any{
		"name": c.name, "channels": c.channels, "samplers": c.samplers,
	})
}
