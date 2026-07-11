package clv

// objects.go decodes the stage's prop placements. An instance record carries a
// property list of 16-byte tuples {u32 typeId, u32 count, ptr value, ptr key}
// whose keys are shared strings ("pos", "rotZ", "scale", "visibility",
// "direction", "object"); the values sit inline in the record — "pos" is an
// (x, y, z) float triple preceded by the instance's Maya DAG name, "object"
// points at the imported bundle the instance places. Every instance has an
// "object" tuple, so instances are enumerated through the relocation table:
// the slots holding a pointer to the shared "object" key string are exactly
// the object-tuple key slots, one per instance.

import "sort"

// Placement is one prop instance: the trailing DAG-path segment names the
// object ("toge_normal_10"), Pos/RotZ/Scale place it.
type Placement struct {
	Name      string
	DagPath   string
	Pos       [3]float32
	RotZ      float32
	Scale     [3]float32
	BundleOff uint32 // offset of the placed bundle record
}

// Placements enumerates the stage's prop instances.
func (c *Clv) Placements() []Placement {
	n := uint32(len(c.Data))
	// the shared "object\0" key string: find a tuple key slot pointing at it.
	// Key slots hold string offsets; collect every distinct string those slots
	// name, and pick the offsets whose string is exactly "object".
	var objKeys []uint32
	seen := map[uint32]bool{}
	for _, slot := range c.Reloc {
		t := c.ptr(slot)
		if t == 0 || t+7 > n || seen[t] {
			continue
		}
		seen[t] = true
		if string(c.Data[t:t+7]) == "object\x00" {
			objKeys = append(objKeys, t)
		}
	}
	var out []Placement
	for _, key := range objKeys {
		// every slot pointing at this key string is one instance's object
		// tuple: {key @slot, typeId @slot+4, count @slot+8, value @slot+12}
		for _, slot := range c.Reloc {
			if c.ptr(slot) != key || slot+16 > n {
				continue
			}
			p := Placement{BundleOff: c.ptr(slot + 12)}
			// the record's earlier tuples precede this one at 16-byte stride
			for t := slot - 16; t >= 16 && t <= slot; t -= 16 {
				name := c.ptr(t)
				val := c.ptr(t + 12)
				if name == 0 || name+4 > n || val == 0 || c.u32(t+8) != 1 {
					break
				}
				stop := false
				switch c.cstr(name) {
				case "pos":
					if val+12 <= n {
						p.Pos = [3]float32{c.f32(val), c.f32(val + 4), c.f32(val + 8)}
						p.DagPath, p.Name = c.nameBefore(val)
					}
					stop = true // pos is the record's first tuple
				case "rotZ":
					if val+4 <= n {
						p.RotZ = c.f32(val)
					}
				case "scale":
					if val+12 <= n {
						p.Scale = [3]float32{c.f32(val), c.f32(val + 4), c.f32(val + 8)}
					}
				case "visibility", "direction":
					// present on every instance; nothing to extract yet
				default:
					stop = true // unknown key: past the record head
				}
				if stop {
					break
				}
			}
			if p.Name != "" {
				out = append(out, p)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// nameBefore reads the C string that ends immediately before offset v — the
// instance's DAG path precedes its inline pos value — returning the full path
// and its trailing segment.
func (c *Clv) nameBefore(v uint32) (path, name string) {
	end := v
	for end > 0 && c.Data[end-1] == 0 {
		end--
	}
	start := end
	for start > 0 && c.Data[start-1] != 0 && end-start < 256 {
		b := c.Data[start-1]
		if b < 0x20 || b > 0x7E {
			break
		}
		start--
	}
	if start == end {
		return "", ""
	}
	path = string(c.Data[start:end])
	name = path
	if i := lastByte(path, '|'); i >= 0 {
		name = path[i+1:]
	}
	return path, name
}

func lastByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}
