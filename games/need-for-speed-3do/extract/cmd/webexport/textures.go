package main

// textures.go resolves the packet's texture tables the way the game does
// (setup at 0x1BEB8): root child 0 = the slice texture groups (face material →
// group, texture sub-index → cel within the group), root child 1 = the
// roadside-object billboard textures (4 mip cels per object texture index).
// Empty (nil) slots inherit the previous entry — the loader's rebase pass
// leaves a zero offset pointing at the last non-empty child.

import (
	"fmt"
	"image"

	"retroreverse.com/tools/platform/threedo"
)

type texCache struct {
	pkt    []byte
	root   *threedo.WrapNode
	images map[int]*image.RGBA // by leaf offset
}

func newTexCache(pkt []byte, root *threedo.WrapNode) *texCache {
	return &texCache{pkt: pkt, root: root, images: map[int]*image.RGBA{}}
}

// child returns node.Children[i] with the game's inherit-previous rule.
func child(node *threedo.WrapNode, i int) *threedo.WrapNode {
	if node == nil {
		return nil
	}
	for ; i >= 0 && i < len(node.Children); i-- {
		if node.Children[i] != nil {
			return node.Children[i]
		}
	}
	return nil
}

// celImage decodes the cel leaf at a packet offset (cached).
func (tc *texCache) celImage(leaf *threedo.WrapNode) (*image.RGBA, error) {
	if leaf == nil || leaf.Kind != "cel" {
		return nil, fmt.Errorf("not a cel leaf")
	}
	if img, ok := tc.images[leaf.Offset]; ok {
		return img, nil
	}
	cel, err := threedo.ParseCel(tc.pkt[leaf.Offset:])
	if err != nil {
		return nil, err
	}
	img, err := cel.Image()
	if err != nil {
		return nil, err
	}
	tc.images[leaf.Offset] = img
	return img, nil
}

// sliceTexture resolves a slice face's texture: material byte (already
// remapped through the group header) → texture group, sub-index → cel.
func (tc *texCache) sliceTexture(group, sub int) (*image.RGBA, error) {
	groups := tc.root.Children[0]
	g := child(groups, group)
	if g == nil || g.Kind != "wwww" {
		return nil, fmt.Errorf("slice texture group %d missing", group)
	}
	return tc.celImage(child(g, sub))
}

// objectTexture resolves a roadside-object texture index (mip 0).
func (tc *texCache) objectTexture(idx int) (*image.RGBA, error) {
	return tc.celImage(child(tc.root.Children[1], idx))
}
