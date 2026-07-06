package lev

import "fmt"

// Common object properties (DATA/COMOBJ.DAT): 11 bytes per item id. The game
// loads this file into DGROUP:5975 and the object-emission dispatcher
// (2DFE:0005) reads byte 0's low two bits as the RENDER CLASS, switching how a
// placed object is drawn (jump table at 2DFE:05D0):
//
//	0  no special handling
//	1  billboard sprite (the common case: items, creatures)
//	2  3D model — the display-list VM programs decoded in extract/model.
//	   For these the item's low 6 bits select the model: values 0x10-0x1F index
//	   a 16-entry model-number table (DGROUP:056C, file 0x5F85C in UW.EXE);
//	   values 0-0xF take the dedicated door path (2DFE:0CBA) which emits the
//	   doorframe (model 1) plus a leaf — door (14), portcullis (12) or the
//	   open-door variant (15) — by variant and state.
//	3  special (decals etc.)
//
// The file: an 11-byte header, then the per-item records.
const (
	comObjHeader  = 11
	comObjRecSize = 11
)

// RenderClass values from COMOBJ byte 0 & 3.
const (
	RenderNone      = 0
	RenderBillboard = 1
	RenderModel     = 2
	RenderSpecial   = 3
)

// ComObj is the parsed common-object property table.
type ComObj struct {
	rec []byte // records, 11 bytes each, from item id 0
}

// ParseComObj wraps DATA/COMOBJ.DAT.
func ParseComObj(data []byte) (*ComObj, error) {
	if len(data) < comObjHeader+comObjRecSize {
		return nil, fmt.Errorf("comobj: short file (%d bytes)", len(data))
	}
	return &ComObj{rec: data[comObjHeader:]}, nil
}

// RenderClass returns how the object-emission pass draws item id (0-511).
func (c *ComObj) RenderClass(itemID uint16) int {
	off := int(itemID) * comObjRecSize
	if off >= len(c.rec) {
		return RenderNone
	}
	return int(c.rec[off] & 3)
}
