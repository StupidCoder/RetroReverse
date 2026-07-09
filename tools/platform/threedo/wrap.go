package threedo

// wrap.go decodes the "wwww" resource container that packages Need for Speed's
// track packets (`DriveData/DriveArt/*_PKT_*`) and the car `.WrapFam` files. It
// is a recursive tree: a `wwww` node holds a count and that many **absolute**
// file offsets, each pointing to a child that is either another `wwww` node or a
// leaf resource — a cel (`CCB `), a 3D object (`ORI3`), or a shape (`SHPM`).
// Offsets of 0 are empty slots.
//
// A track packet decodes to hundreds of cels (road and scenery textures), a few
// ORI3 track-slice objects and SHPM backdrops — every one in a format this
// package already decodes (cel.go, model.go, shpm.go). So the whole track packet
// is readable without booting the game; the remaining piece is the road segment
// layout that references these resources.

import "fmt"

// Resource is one leaf in a wwww tree: its file offset and detected kind.
type Resource struct {
	Offset int
	Kind   string // "cel", "model", "shape", or "unknown"
	Depth  int    // nesting depth in the container tree
}

// ParseWrap walks a wwww container and returns its leaf resources in tree order.
func ParseWrap(data []byte) ([]Resource, error) {
	if len(data) < 8 || string(data[0:4]) != "wwww" {
		return nil, fmt.Errorf("threedo: not a wwww container")
	}
	var out []Resource
	seen := map[int]bool{}
	var walk func(off, depth int)
	walk = func(off, depth int) {
		if off < 0 || off+8 > len(data) || seen[off] || depth > 8 {
			return
		}
		seen[off] = true
		if string(data[off:off+4]) != "wwww" {
			out = append(out, Resource{Offset: off, Kind: kindOf(data, off), Depth: depth})
			return
		}
		n := int(be32(data[off+4:]))
		if n < 0 || n > 100000 {
			return
		}
		// Child offsets are relative to this container's own base (the outer
		// container sits at 0, so its offsets look absolute).
		for i := 0; i < n; i++ {
			p := off + 8 + i*4
			if p+4 > len(data) {
				break
			}
			if rel := int(be32(data[p:])); rel != 0 {
				walk(off+rel, depth+1)
			}
		}
	}
	walk(0, 0)
	return out, nil
}

// kindOf classifies a leaf by its 4-char magic.
func kindOf(data []byte, off int) string {
	if off+4 > len(data) {
		return "unknown"
	}
	switch string(data[off : off+4]) {
	case "CCB ":
		return "cel"
	case "ORI3":
		return "model"
	case "SHPM":
		return "shape"
	default:
		return "unknown"
	}
}

// Inventory counts leaf resources by kind.
func Inventory(res []Resource) map[string]int {
	m := map[string]int{}
	for _, r := range res {
		m[r.Kind]++
	}
	return m
}

// WrapNode is one node of a wwww tree with its children resolved, mirroring
// how the game itself navigates a packet: the loader rebases each child-offset
// word into a pointer, so game code addresses resources as "root child k",
// "child k of child 0", and so on. Children[i] is nil for an empty (0) slot.
type WrapNode struct {
	Offset   int    // node or leaf offset within the container
	Kind     string // "wwww", "cel", "model", "shape", or "unknown"
	Children []*WrapNode
}

// ParseWrapTree decodes a wwww container preserving the tree structure, so a
// caller can navigate by child index exactly like the game's loader does.
func ParseWrapTree(data []byte) (*WrapNode, error) {
	if len(data) < 8 || string(data[0:4]) != "wwww" {
		return nil, fmt.Errorf("threedo: not a wwww container")
	}
	seen := map[int]bool{}
	var walk func(off, depth int) *WrapNode
	walk = func(off, depth int) *WrapNode {
		if off < 0 || off+8 > len(data) || seen[off] || depth > 8 {
			return nil
		}
		if string(data[off:off+4]) != "wwww" {
			return &WrapNode{Offset: off, Kind: kindOf(data, off)}
		}
		seen[off] = true
		n := int(be32(data[off+4:]))
		if n < 0 || n > 100000 {
			return nil
		}
		node := &WrapNode{Offset: off, Kind: "wwww"}
		for i := 0; i < n; i++ {
			p := off + 8 + i*4
			if p+4 > len(data) {
				break
			}
			var child *WrapNode
			if rel := int(be32(data[p:])); rel != 0 {
				child = walk(off+rel, depth+1)
			}
			node.Children = append(node.Children, child)
		}
		return node
	}
	root := walk(0, 0)
	if root == nil {
		return nil, fmt.Errorf("threedo: malformed wwww container")
	}
	return root, nil
}
