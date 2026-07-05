// Actor-model binding, traced statically from the game's actor system (Part V §2):
//
//   - The factory at $02043098 spawns actors through a pointer array indexed by
//     actor ID. The boot code at $0201A128 assigns that global its static value:
//     the 326-entry profile-pointer array at ARM9 $02090864. A profile starts
//     with its create function; +4 is the actor ID.
//   - Models reach actors two ways, both statically visible:
//     1. file-handle slots: `$02017ACC(slot, fileID)` (and the $02017B4C
//     variant) bind a BSS slot to an internal file ID — 26 call sites in the
//     ARM9/engine overlay plus more in each level overlay's constructors.
//     Actor code then references its slot address as an LDR literal.
//     2. direct loads: a file ID materialized as an LDR literal feeds a load
//     call a few instructions later (e.g. the tree code at $020EC240 reads
//     its 5-entry model table at $0210ABB8 and calls $02016F9C).
//   - The create function allocates the object and installs its vtable (a
//     literal pointing at an array of method pointers); the model load usually
//     sits in a vtable method (the tree's table read at $020EC240), so the
//     trace follows create -> callees -> vtable methods, collecting .bmd files
//     from both mechanisms. Bank-resident actors (IDs past the ARM9 array, e.g.
//     339 daObjMc_Metalnet) are picked up from their RTTI profiles in the
//     loaded level overlay.
package sm64ds

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	profArrayOff = 0x8C864 // ARM9 file offset of the profile-pointer array ($02090864)
	numActors    = 326
	regSlotA     = 0x02017ACC // file-handle registration (slot, fileID)
	regSlotB     = 0x02017B4C
)

type abin struct {
	data []byte
	base uint32
}

func (b *abin) has(p uint32) bool   { return p >= b.base && p+4 <= b.base+uint32(len(b.data)) }
func (b *abin) u32(p uint32) uint32 { return le.Uint32(b.data[p-b.base:]) }

// blTarget decodes a BL instruction's destination, or 0.
func blTarget(w, addr uint32) uint32 {
	if w&0x0F000000 != 0x0B000000 || w>>28 != 0xE {
		return 0
	}
	off := int32(w<<8) >> 6
	return uint32(int64(addr) + 8 + int64(off))
}

// ActorModels is the result of the static trace: for each actor ID, the .bmd
// stems its create-code loads (order preserved; the tree actor's list is its
// param-indexed model table).
type ActorModels map[int][]string

// TraceActorModels harvests actor -> model bindings for the given overlay set
// (arm9 + engine overlays + one level overlay, so bank-resident actors resolve
// against the right level).
func (ls *LevelSet) TraceActorModels(levelOverlay int) (ActorModels, error) {
	arm9 := &abin{ls.arm9, 0x02004000}
	bins := []*abin{arm9}
	for _, id := range []int{0, 1, 2, levelOverlay} {
		d, err := ls.overlayData(id)
		if err != nil {
			return nil, err
		}
		bins = append(bins, &abin{d, ls.ovls[id].RAMAddr})
	}
	find := func(p uint32) *abin {
		// later bins win (the level overlay shadows the shared bank base)
		var r *abin
		for _, b := range bins {
			if b.has(p) {
				r = b
			}
		}
		return r
	}

	isBMD := map[uint32]string{}
	for i, n := range ls.intName {
		// skip ID 0: a zero literal is far more often a null/init value than a
		// real reference to the first file
		if i > 0 && strings.HasSuffix(n, ".bmd") {
			isBMD[uint32(i)] = strings.TrimSuffix(filepath.Base(n), ".bmd")
		}
	}

	// pass 1: file-handle slots (slot -> file) from registration call sites
	slots := map[uint32]uint32{}
	for _, b := range bins {
		for a := b.base; b.has(a); a += 4 {
			t := blTarget(b.u32(a), a)
			if t != regSlotA && t != regSlotB {
				continue
			}
			var slot, file uint32 = 0, 0xFFFFFFFF
			for k := uint32(4); k <= 16; k += 4 {
				if !b.has(a - k) {
					break
				}
				w := b.u32(a - k)
				if w&0x0FFF0000 != 0x059F0000 {
					continue
				}
				lit := a - k + 8 + (w & 0xFFF)
				if !b.has(lit) {
					continue
				}
				switch (w >> 12) & 0xF {
				case 0:
					slot = b.u32(lit)
				case 1:
					file = b.u32(lit)
				}
			}
			if slot != 0 && file < 2058 {
				if _, bmd := isBMD[file]; bmd {
					slots[slot] = file
				}
			}
		}
	}

	// scanFn collects model files referenced by one function body (create fn or
	// callee): slot literals and direct file-ID literals.
	isCode := func(p uint32) bool {
		return (p >= 0x02004000 && p < 0x020AA420) || (p >= 0x020AA420 && p < 0x02150000) || (p >= 0x01FF8000 && p < 0x02000000)
	}
	budget := 0
	var scanFn func(b *abin, fn uint32, depth int, out *[]uint32, seen map[uint32]bool)
	scanFn = func(b *abin, fn uint32, depth int, out *[]uint32, seen map[uint32]bool) {
		if budget <= 0 || seen[fn] {
			return
		}
		seen[fn] = true // function addrs and file IDs share the set; ranges are disjoint
		budget--
		end := fn + 0x400
		for a := fn; a < end && b.has(a); a += 4 {
			w := b.u32(a)
			// literal pools: LDR rX,[pc,#imm]
			if w&0x0FFF0000 == 0x059F0000 {
				lit := a + 8 + (w & 0xFFF)
				if b.has(lit) {
					v := b.u32(lit)
					if fid, ok := slots[v]; ok && !seen[fid] {
						seen[fid] = true
						*out = append(*out, fid)
					} else if _, ok := isBMD[v]; ok && v < 2058 && !seen[v] {
						seen[v] = true
						*out = append(*out, v)
					} else if tb := find(v); tb != nil && depth > 0 && tb.has(v) && isCode(tb.u32(v)) {
						// vtable: an array of method pointers — scan the first entries
						for j := uint32(0); j < 48; j += 4 {
							if !tb.has(v + j) {
								break
							}
							m := tb.u32(v + j)
							if !isCode(m) {
								break
							}
							if mb := find(m); mb != nil {
								scanFn(mb, m, depth-1, out, seen)
							}
						}
					} else if tb := find(v); tb != nil && v&1 == 0 {
						// u16 table of file IDs (tree pattern): 2+ consecutive bmd ids
						n := 0
						for j := uint32(0); j < 16; j += 2 {
							if !tb.has(v + j) {
								break
							}
							e := uint32(le.Uint16(tb.data[v+j-tb.base:]))
							if _, ok := isBMD[e]; !ok {
								break
							}
							n++
						}
						if n >= 2 {
							for j := uint32(0); j < uint32(n)*2; j += 2 {
								e := uint32(le.Uint16(tb.data[v+j-tb.base:]))
								if !seen[e] {
									seen[e] = true
									*out = append(*out, e)
								}
							}
						}
					}
				}
			}
			// descend into direct callees once
			if depth > 0 {
				if t := blTarget(w, a); t != 0 && t != regSlotA && t != regSlotB {
					if cb := find(t); cb != nil {
						scanFn(cb, t, depth-1, out, seen)
					}
				}
			}
			// function end: pop with pc or bx lr (past the prologue)
			if a > fn && (w&0x0FFF8000 == 0x08BD8000 || w == 0xE12FFF1E) {
				break
			}
		}
	}
	// actor -> create fn: the ARM9 profile array, plus RTTI profiles in the
	// loaded overlays for bank-resident actors ({createFn, id, id, typeinfo}).
	creates := map[int]uint32{}
	for id := 0; id < numActors; id++ {
		pp := le.Uint32(ls.arm9[profArrayOff+id*4:])
		if pb := find(pp); pb != nil && pb.has(pp) {
			creates[id] = pb.u32(pp)
		}
	}
	for _, b := range bins[1:] {
		for i := 0; i+0x24 <= len(b.data); i += 4 {
			create := le.Uint32(b.data[i:])
			if !isCode(create) {
				continue
			}
			id1, id2 := le.Uint16(b.data[i+4:]), le.Uint16(b.data[i+6:])
			if id1 != id2 || id1 < numActors || id1 >= 1024 {
				continue
			}
			ti := le.Uint32(b.data[i+0x20:])
			if b.has(ti) && b.has(b.u32(ti)+4) {
				if _, dup := creates[int(id1)]; !dup {
					creates[int(id1)] = create
				}
			}
		}
	}

	out := ActorModels{}
	for id, create := range creates {
		cb := find(create)
		if cb == nil {
			continue
		}
		var files []uint32
		budget = 48
		scanFn(cb, create, 2, &files, map[uint32]bool{})
		if len(files) == 0 {
			continue
		}
		var stems []string
		for _, f := range files {
			stems = append(stems, isBMD[f])
		}
		out[id] = stems
	}
	return out, nil
}

// die helper for commands
func Die(err error) {
	fmt.Fprintln(os.Stderr, "sm64ds:", err)
	os.Exit(1)
}
