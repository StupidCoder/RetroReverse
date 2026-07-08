// Package strpak decodes DATA/STRINGS.PAK, Ultima Underworld's Huffman-packed
// text archive (every in-game string: UI messages, object names, descriptions,
// conversation lines, cutscene text…).
//
// Layout (all little-endian), derived from the bytes:
//
//	uint16 nnodes
//	nnodes x { uint8 char, uint8 parent, uint8 left, uint8 right }  // Huffman tree
//	uint16 nblocks
//	nblocks x { uint16 blockID, uint32 fileOffset }                 // block directory
//	each block @ fileOffset:
//	  uint16 nstrings
//	  nstrings x uint16 stringOffset      // byte offset into this block's bitstream
//	  <packed Huffman bitstream>
//
// A tree node is a leaf when left == right == 0xFF, and then char is its symbol;
// otherwise left/right are child node indices. The root is the LAST node. A
// string is decoded by walking the tree from the root: read the bitstream MSB
// first, take the left child on a 0 bit and the right child on a 1 bit, emit a
// leaf's char on arrival and restart at the root. The pipe '|' (0x7C) terminates
// a string. Verified: block 1 string 2 == "The key does not fit.\n".
package strpak

import (
	"encoding/binary"
	"fmt"
)

type node struct{ char, parent, left, right byte }

// Archive is a parsed STRINGS.PAK. Decoding is done lazily from the retained
// file bytes, so parsing is cheap.
type Archive struct {
	data   []byte
	nodes  []node
	root   int
	blocks map[int]int // block id -> file offset
	ids    []int       // block ids in directory order
}

// Terminator is the character that ends every packed string.
const Terminator = '|'

// Parse reads the archive header (Huffman tree + block directory).
func Parse(data []byte) (*Archive, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("strpak: too short")
	}
	nn := int(binary.LittleEndian.Uint16(data))
	end := 2 + nn*4
	if nn == 0 || end+2 > len(data) {
		return nil, fmt.Errorf("strpak: bad node count %d", nn)
	}
	nodes := make([]node, nn)
	for i := range nodes {
		o := 2 + i*4
		nodes[i] = node{data[o], data[o+1], data[o+2], data[o+3]}
	}
	off := end
	nb := int(binary.LittleEndian.Uint16(data[off:]))
	off += 2
	blocks := make(map[int]int, nb)
	ids := make([]int, 0, nb)
	for i := 0; i < nb; i++ {
		if off+6 > len(data) {
			return nil, fmt.Errorf("strpak: block directory truncated at %d", i)
		}
		id := int(binary.LittleEndian.Uint16(data[off:]))
		fo := int(binary.LittleEndian.Uint32(data[off+2:]))
		if fo < 0 || fo+2 > len(data) {
			return nil, fmt.Errorf("strpak: block %d offset %d out of range", id, fo)
		}
		blocks[id] = fo
		ids = append(ids, id)
		off += 6
	}
	return &Archive{data: data, nodes: nodes, root: nn - 1, blocks: blocks, ids: ids}, nil
}

// BlockIDs returns the block ids in directory order.
func (a *Archive) BlockIDs() []int {
	out := make([]int, len(a.ids))
	copy(out, a.ids)
	return out
}

// Block decodes every string in the block; ok is false if there is no such block.
func (a *Archive) Block(id int) (strs []string, ok bool) {
	bo, ok := a.blocks[id]
	if !ok {
		return nil, false
	}
	ns := int(binary.LittleEndian.Uint16(a.data[bo:]))
	dataStart := bo + 2 + ns*2
	out := make([]string, ns)
	for i := 0; i < ns; i++ {
		so := int(binary.LittleEndian.Uint16(a.data[bo+2+i*2:]))
		out[i] = a.decode(dataStart + so)
	}
	return out, true
}

// String decodes a single string by block id and index.
func (a *Archive) String(block, index int) (string, bool) {
	bo, ok := a.blocks[block]
	if !ok {
		return "", false
	}
	ns := int(binary.LittleEndian.Uint16(a.data[bo:]))
	if index < 0 || index >= ns {
		return "", false
	}
	dataStart := bo + 2 + ns*2
	so := int(binary.LittleEndian.Uint16(a.data[bo+2+index*2:]))
	return a.decode(dataStart + so), true
}

// decode walks the Huffman tree from the root, reading the bitstream MSB-first
// starting at byte offset start, until the '|' terminator (or the data runs out).
func (a *Archive) decode(start int) string {
	var out []byte
	n := a.root
	bit := start * 8
	for bit>>3 < len(a.data) {
		b := (a.data[bit>>3] >> (7 - uint(bit&7))) & 1
		bit++
		if b == 0 {
			n = int(a.nodes[n].left)
		} else {
			n = int(a.nodes[n].right)
		}
		if n < 0 || n >= len(a.nodes) {
			break
		}
		if a.nodes[n].left == 0xFF { // leaf: emit its symbol
			if a.nodes[n].char == Terminator {
				break
			}
			out = append(out, a.nodes[n].char)
			n = a.root
		}
	}
	return string(out)
}
