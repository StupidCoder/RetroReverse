package mkds

import (
	"encoding/binary"
	"strings"
)

// The map-object descriptor table (Part V §2): ARM9 .data at $0216B288 holds
// zero-terminated records {u32 objectID, u32 descriptor, u32 auxFn} — the table
// the engine walks (linear scan at $020D3452) to turn a course map's OBJI
// placements into live objects. Each descriptor starts with the instance size
// (+0x04) and callback slots whose literal pools name the NSBMD/NSBCA resources
// the object loads; that is where the ID→model bindings here come from.

const (
	ramBase      = 0x02000000
	objTableAddr = 0x0216B288
)

// ObjRecord is one entry of the map-object descriptor table.
type ObjRecord struct {
	ID    int
	Desc  uint32 // descriptor address (0 = registered but empty)
	Aux   uint32 // per-ID callback for IDs sharing a generic descriptor
	Size  uint32 // instance size (descriptor+0x04)
	Names []string
}

type arm9 struct{ data []byte }

func (b arm9) u32(addr uint32) uint32 {
	off := addr - ramBase
	if int(off)+4 > len(b.data) {
		return 0
	}
	return binary.LittleEndian.Uint32(b.data[off:])
}

// str returns the NUL-terminated ASCII string at addr, or "".
func (b arm9) str(addr uint32) string {
	off := int(addr - ramBase)
	if off < 0 || off >= len(b.data) {
		return ""
	}
	end := off
	for end < len(b.data) && b.data[end] != 0 {
		if b.data[end] < 0x20 || b.data[end] > 0x7E {
			return ""
		}
		end++
	}
	if end == off || end-off > 64 {
		return ""
	}
	return string(b.data[off:end])
}

// poolStrings scans a function body for literal-pool words that point at
// resource-name strings (.nsbmd/.nsbca/.nsbtp/.nsbta) in .data.
func (b arm9) poolStrings(fn uint32, budget int, seen map[uint32]bool) []string {
	fn &^= 1 // thumb bit
	if fn < ramBase || seen[fn] {
		return nil
	}
	seen[fn] = true
	var out []string
	for off := 0; off < budget; off += 4 {
		w := b.u32(fn + uint32(off))
		if w < 0x02150000 || w >= 0x02180000 {
			continue
		}
		s := b.str(w)
		if s == "" {
			continue
		}
		l := strings.ToLower(s)
		if strings.HasSuffix(l, ".nsbmd") || strings.HasSuffix(l, ".nsbca") ||
			strings.HasSuffix(l, ".nsbtp") || strings.HasSuffix(l, ".nsbta") {
			out = append(out, s)
		}
	}
	return out
}

// ObjectTable decodes the descriptor table from a decompressed ARM9 image.
func ObjectTable(data []byte) []ObjRecord {
	b := arm9{data}
	var recs []ObjRecord
	for addr := uint32(objTableAddr); ; addr += 12 {
		id, desc, aux := b.u32(addr), b.u32(addr+4), b.u32(addr+8)
		if id == 0 && desc == 0 && aux == 0 {
			break
		}
		r := ObjRecord{ID: int(id), Desc: desc, Aux: aux}
		if desc != 0 {
			r.Size = b.u32(desc + 4)
			// Scan the descriptor's callback slots' bodies for resource names;
			// registration trampolines are tiny, so also follow one level of
			// code pointers found in their pools.
			seen := map[uint32]bool{}
			for slot := uint32(0x08); slot <= 0x2C; slot += 4 {
				fn := b.u32(desc + slot)
				if fn < ramBase || fn >= 0x021C0000 {
					continue
				}
				r.Names = append(r.Names, b.poolStrings(fn, 0x100, seen)...)
				for off := 0; off < 0x40; off += 4 {
					w := b.u32((fn &^ 1) + uint32(off))
					if w >= ramBase && w < 0x02160000 { // code region
						r.Names = append(r.Names, b.poolStrings(w, 0x200, seen)...)
					}
				}
			}
			r.Names = dedup(r.Names)
		}
		recs = append(recs, r)
	}
	return recs
}

// ObjectModelBindings reduces the table to objectID → model base name (the first
// .nsbmd resource, lowercased, directory and extension stripped) — the name the
// placements exporter matches against a course archive's model names. A few
// bindings the literal-pool scan cannot see (loaders that build their model
// names at runtime) are pinned from course context and marked as such.
func ObjectModelBindings(data []byte) map[int]string {
	out := map[int]string{}
	for _, r := range ObjectTable(data) {
		for _, n := range r.Names {
			l := strings.ToLower(n)
			if strings.HasSuffix(l, ".nsbmd") {
				l = strings.TrimSuffix(l, ".nsbmd")
				if i := strings.LastIndexByte(l, '/'); i >= 0 {
					l = l[i+1:]
				}
				out[r.ID] = l
				break
			}
		}
	}
	// Context-pinned (not from literal pools): resolved by matching a course's
	// unbound OBJI IDs against its remaining archive models — each pin is
	// unambiguous on the course(s) that use the ID.
	for id, name := range map[int]string{
		0xCC: "bridge", // Delfino Square's two drawbridges (rot.Y = ±90)
	} {
		if _, ok := out[id]; !ok {
			out[id] = name
		}
	}
	return out
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
