// Package clv decodes the game's stage files (st_*.clv): GPRS-compressed,
// pointer-patched images the engine consumes in place. All stored pointers are
// file offsets biased by +1 (0 = NULL); a relocation table lists every pointer
// slot, and the loader adds the load base - 1 to each. The scene section holds
// the stage layout — world bounds over a grid of cells, each cell a list of
// draw batches {material, triangle strips} whose strips are stored in the GE's
// own vertex format (u16 u,v texcoords + float x,y,z).
package clv

import (
	"encoding/binary"
	"fmt"
	"math"

	"retroreverse.com/games/loco-roco-psp/extract/gprs"
)

// Clv is a parsed stage file.
type Clv struct {
	Data  []byte   // the decompressed, unpatched image
	Reloc []uint32 // pointer-slot offsets from the relocation table

	SceneOff uint32 // offset of the scene root record
	RelocOff uint32 // offset of the relocation table

	Layout Layout
}

// Layout is the stage's spatial organization: world bounds and a grid of
// cells, each holding the draw batches whose strips fall in that cell.
type Layout struct {
	X, Y, W, H float32 // world bounds
	Z0, Z1     float32
	Cols, Rows uint32
	CellSize   float32
	Cells      []Cell // Cols*Rows entries, row-major
}

// Cell is one grid cell: its draw batches.
type Cell struct {
	Batches []Batch
}

// Batch is one material's strips within a cell.
type Batch struct {
	MaterialOff  uint32 // offset of the material record
	MaterialName string // the Maya shading-group name ("stage_a_tex", ...)
	Color        uint32 // material RGBA (ABGR as stored)
	Strips       []Strip
}

// Strip is one GE triangle strip.
type Strip struct {
	Flags uint16
	Off   uint32 // offset of the vertex data in the file
	Verts []Vert
}

// Vert is one strip vertex: fractional s16 texcoords and float world position.
type Vert struct {
	U, V    int16
	X, Y, Z float32
}

func (c *Clv) u32(o uint32) uint32 {
	return binary.LittleEndian.Uint32(c.Data[o : o+4])
}
func (c *Clv) f32(o uint32) float32 { return math.Float32frombits(c.u32(o)) }

// ptr reads a +1-biased pointer slot: the stored value minus 1 is a file
// offset; 0 is NULL (returned as 0, valid offsets are never 0).
func (c *Clv) ptr(o uint32) uint32 {
	v := c.u32(o)
	if v == 0 {
		return 0
	}
	return v - 1
}

// Parse decompresses and decodes a .clv stage file.
func Parse(raw []byte) (*Clv, error) {
	data, err := gprs.Decompress(raw)
	if err != nil {
		return nil, fmt.Errorf("clv: %w", err)
	}
	c := &Clv{Data: data}
	n := uint32(len(data))
	if n < 0x20 {
		return nil, fmt.Errorf("clv: too small")
	}
	c.SceneOff = c.ptr(8)
	c.RelocOff = c.ptr(0xC)
	if c.SceneOff == 0 || c.SceneOff >= n || c.RelocOff == 0 || c.RelocOff >= n {
		return nil, fmt.Errorf("clv: bad header (scene %#x reloc %#x)", c.SceneOff, c.RelocOff)
	}
	// relocation table: +1-biased slot offsets, 0-terminated. Every listed
	// slot must itself hold a valid biased offset — this validates the whole
	// pointer scheme over the file.
	for o := c.RelocOff; o+4 <= n; o += 4 {
		v := c.u32(o)
		if v == 0 {
			break
		}
		if v > n {
			return nil, fmt.Errorf("clv: reloc entry %#x out of range", v)
		}
		slot := v - 1
		if t := c.u32(slot); t == 0 || t > n {
			return nil, fmt.Errorf("clv: slot %#x holds invalid pointer %#x", slot, t)
		}
		c.Reloc = append(c.Reloc, slot)
	}
	if err := c.parseScene(); err != nil {
		return nil, err
	}
	return c, nil
}

// parseScene walks scene root -> layout -> cells -> batch lists -> strips.
func (c *Clv) parseScene() error {
	// scene root: {u32 type, u32 count, ptr materials, ptr objects, ptr layout}
	layoutOff := c.ptr(c.SceneOff + 0x10)
	if layoutOff == 0 {
		return fmt.Errorf("clv: no layout pointer")
	}
	L := &c.Layout
	L.X = c.f32(layoutOff)
	L.Y = c.f32(layoutOff + 4)
	L.W = c.f32(layoutOff + 8)
	L.H = c.f32(layoutOff + 12)
	L.Z0 = c.f32(layoutOff + 16)
	L.Z1 = c.f32(layoutOff + 20)
	L.Cols = c.u32(layoutOff + 24)
	L.Rows = c.u32(layoutOff + 28)
	L.CellSize = c.f32(layoutOff + 32)
	cellsOff := c.ptr(layoutOff + 36)
	if L.Cols == 0 || L.Rows == 0 || L.Cols*L.Rows > 4096 || cellsOff == 0 {
		return fmt.Errorf("clv: bad layout %dx%d cells@%#x", L.Cols, L.Rows, cellsOff)
	}
	L.Cells = make([]Cell, L.Cols*L.Rows)
	for i := range L.Cells {
		listOff := c.ptr(cellsOff + uint32(i)*4)
		if listOff == 0 {
			continue
		}
		cell, err := c.parseBatchList(listOff)
		if err != nil {
			return fmt.Errorf("clv: cell %d: %w", i, err)
		}
		L.Cells[i] = cell
	}
	return nil
}

// parseBatchList reads a cell's batch list: {u16 pad, u16 count, u32 zero,
// ptr entries}, entries of 20 bytes {u32 zero, ptr material, u32 stripCount,
// ptr stripTable, u32 zero}.
func (c *Clv) parseBatchList(off uint32) (Cell, error) {
	var cell Cell
	count := uint32(binary.LittleEndian.Uint16(c.Data[off+2 : off+4]))
	entries := c.ptr(off + 8)
	if entries == 0 {
		return cell, fmt.Errorf("batch list %#x has no entries", off)
	}
	for i := uint32(0); i < count; i++ {
		e := entries + i*20
		b := Batch{
			MaterialOff: c.ptr(e + 8),
		}
		if b.MaterialOff != 0 {
			// material record: {ptr name, u32, u32 rgba, ptr texture}
			if nameOff := c.ptr(b.MaterialOff); nameOff != 0 {
				b.MaterialName = c.cstr(nameOff)
			}
			b.Color = c.u32(b.MaterialOff + 8)
		}
		stripCount := c.u32(e + 12)
		table := c.ptr(e + 16)
		if table == 0 || stripCount > 4096 {
			return cell, fmt.Errorf("batch %#x: bad strip table (%d strips @%#x)", e, stripCount, table)
		}
		for k := uint32(0); k < stripCount; k++ {
			t := table + k*8
			flags := binary.LittleEndian.Uint16(c.Data[t : t+2])
			nverts := uint32(binary.LittleEndian.Uint16(c.Data[t+2 : t+4]))
			verts := c.ptr(t + 4)
			if verts == 0 || nverts == 0 || verts+nverts*16 > uint32(len(c.Data)) {
				return cell, fmt.Errorf("strip %#x: bad verts (%d @%#x)", t, nverts, verts)
			}
			st := Strip{Flags: flags, Off: verts, Verts: make([]Vert, nverts)}
			for v := uint32(0); v < nverts; v++ {
				p := verts + v*16
				st.Verts[v] = Vert{
					U: int16(binary.LittleEndian.Uint16(c.Data[p : p+2])),
					V: int16(binary.LittleEndian.Uint16(c.Data[p+2 : p+4])),
					X: c.f32(p + 4),
					Y: c.f32(p + 8),
					Z: c.f32(p + 12),
				}
			}
			b.Strips = append(b.Strips, st)
		}
		cell.Batches = append(cell.Batches, b)
	}
	return cell, nil
}

func (c *Clv) cstr(o uint32) string {
	end := o
	for end < uint32(len(c.Data)) && c.Data[end] != 0 {
		end++
	}
	return string(c.Data[o:end])
}
