package main

// pose.go prints the animation pose as matrices, for byte-comparing our .key
// evaluation against the world matrices the game keeps in RAM (the model
// object's +0x14 array).

import (
	"fmt"
	"math"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

// srt34 builds the game's local matrix: Rz·Ry·Rx with the translation set.
func srt34(p lm.Pose) lm.Mtx34 {
	sx, cx := sincos(p.Rot[0])
	sy, cy := sincos(p.Rot[1])
	sz, cz := sincos(p.Rot[2])
	return lm.Mtx34{
		cy * cz, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx, p.Translate[0],
		sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx, p.Translate[1],
		-sy, cy * sx, cy * cx, p.Translate[2],
	}
}

func sincos(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func poseDump(image, spec, animName string, frame float32) error {
	path, member, _ := cutSpec(spec)
	m, key, err := loadModelAnim(image, path, member, animName)
	if err != nil {
		return err
	}
	local := make([]lm.Mtx34, len(m.Nodes))
	for i := range m.Nodes {
		local[i] = srt34(key.Eval(i, frame))
	}
	world := make([]lm.Mtx34, len(m.Nodes))
	for i := range world {
		world[i] = local[i]
	}
	var walk func(idx int, parent lm.Mtx34)
	walk = func(idx int, parent lm.Mtx34) {
		for idx >= 0 && idx < len(m.Nodes) {
			world[idx] = parent.Mul(local[idx])
			if m.Nodes[idx].Child >= 0 {
				walk(m.Nodes[idx].Child, world[idx])
			}
			idx = m.Nodes[idx].Sibling
		}
	}
	walk(0, lm.Identity34)
	for i := range m.Nodes {
		fmt.Printf("flat %3d %v\n", i, local[i])
		fmt.Printf("tree %3d %v\n", i, world[i])
	}
	return nil
}

func cutSpec(spec string) (path, member string, ok bool) {
	for i := len(spec) - 1; i >= 0; i-- {
		if spec[i] == ':' {
			return spec[:i], spec[i+1:], true
		}
	}
	return spec, "", false
}

func loadModelAnim(image, path, member, animName string) (*lm.MDL, *lm.Key, error) {
	b, err := discFile(image, path)
	if err != nil {
		return nil, nil, err
	}
	if len(b) >= 4 && string(b[:4]) == "Yay0" {
		if b, err = lm.Yay0(b); err != nil {
			return nil, nil, err
		}
	}
	files, err := lm.RARC(b)
	if err != nil {
		return nil, nil, err
	}
	var m *lm.MDL
	var key *lm.Key
	for _, f := range files {
		if equalFold(f.Name, member) {
			if m, err = lm.ParseMDL(f.Data); err != nil {
				return nil, nil, err
			}
		}
		if equalFold(f.Name, animName) {
			if key, err = lm.ParseKey(f.Data); err != nil {
				return nil, nil, err
			}
		}
	}
	if m == nil || key == nil {
		return nil, nil, fmt.Errorf("member %q or %q not found in %s", member, animName, path)
	}
	return m, key, nil
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 32
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
