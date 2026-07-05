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

// cstr reads a NUL-terminated string at p (empty if out of range/unterminated).
func (b *abin) cstr(p uint32) string {
	// (has() alone can pass on p near 0xFFFFFFFF via uint32 wrap of p+4)
	if p < b.base || uint64(p)+4 > uint64(b.base)+uint64(len(b.data)) {
		return ""
	}
	d := b.data[p-b.base:]
	for i := 0; i < len(d) && i < 64; i++ {
		if d[i] == 0 {
			return string(d[:i])
		}
	}
	return ""
}

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
	// actor -> create fn: the ARM9 profile array, plus RTTI profiles in the
	// loaded overlays for bank-resident actors.
	creates := map[int]uint32{}
	for id := 0; id < numActors; id++ {
		pp := le.Uint32(ls.arm9[profArrayOff+id*4:])
		for _, b := range bins {
			if b.has(pp) {
				creates[id] = b.u32(pp)
			}
		}
	}
	for _, b := range bins[1:] {
		for id, create := range rttiProfiles(b) {
			if _, dup := creates[id]; !dup {
				creates[id] = create.fn
			}
		}
	}
	return ls.traceModels(bins, creates), nil
}

// rttiProfile is an actor profile located by its C++ typeinfo pointer.
type rttiProfile struct {
	fn   uint32 // create function
	name string // demangled-ish class name (as embedded, e.g. "7daKrb_c")
}

// rttiProfiles scans one binary for actor profiles carrying RTTI: the record
// holds the create function at +0, the u16 actor ID at +4, and a typeinfo
// pointer at +$20 whose +4 points at the mangled class name. (The engine-actor
// profiles in the ARM9 array have no typeinfo; only overlay actors do. The
// second u16 at +6 is NOT a duplicate of the ID in the enemy banks — the
// goomba's profile at $02130924 carries $1A there — so the ID is validated by
// the typeinfo alone.)
func rttiProfiles(b *abin) map[int]rttiProfile {
	out := map[int]rttiProfile{}
	for i := 0; i+0x24 <= len(b.data); i += 4 {
		ti := le.Uint32(b.data[i+0x20:])
		if !b.has(ti) || !b.has(ti+4) {
			continue
		}
		id := int(le.Uint16(b.data[i+4:]))
		if id >= 1024 {
			continue
		}
		create := le.Uint32(b.data[i:])
		if !isCodeAddr(create) {
			continue
		}
		np := b.u32(ti + 4)
		if !b.has(np) {
			continue
		}
		name := b.cstr(np)
		if !plausibleClass(name) {
			continue
		}
		if _, dup := out[id]; !dup {
			out[id] = rttiProfile{fn: create, name: name}
		}
	}
	return out
}

// plausibleClass accepts the game's actor class-name shape: an optional length
// prefix, then "da..._c" / "dSc...", as embedded by the compiler's RTTI.
func plausibleClass(n string) bool {
	if len(n) < 4 || len(n) > 40 {
		return false
	}
	i := 0
	for i < len(n) && n[i] >= '0' && n[i] <= '9' {
		i++
	}
	rest := n[i:]
	if !strings.HasPrefix(rest, "da") && !strings.HasPrefix(rest, "dSc") && !strings.HasPrefix(rest, "dBg") && !strings.HasPrefix(rest, "dEn") {
		return false
	}
	return strings.HasSuffix(rest, "_c")
}

func isCodeAddr(p uint32) bool {
	return (p >= 0x02004000 && p < 0x020AA420) || (p >= 0x020AA420 && p < 0x02150000) || (p >= 0x01FF8000 && p < 0x02000000)
}

// TraceBankActorModels harvests actor -> model bindings from the enemy-bank
// overlays (60-102): each bank carries RTTI actor profiles and registers its
// model files in its static initialisers. The same actor ID can be provided by
// different banks with different classes (the per-level bank set decides which
// is loaded); those ambiguous IDs are skipped rather than guessed.
func (ls *LevelSet) TraceBankActorModels() (ActorModels, error) {
	arm9 := &abin{ls.arm9, 0x02004000}
	shared := []*abin{arm9}
	for _, id := range []int{0, 1, 2} {
		d, err := ls.overlayData(id)
		if err != nil {
			return nil, err
		}
		shared = append(shared, &abin{d, ls.ovls[id].RAMAddr})
	}
	classOf := map[int]string{}
	ambiguous := map[int]bool{}
	merged := ActorModels{}
	// The ARM9 profile array itself points into bank space for bank actors (the
	// entry is only meaningful when that bank is loaded — actor 200's entry
	// $021308EC lands in overlay 84, three records before the goomba's). Since
	// banks share RAM slots, a pointer can fall inside several banks; each
	// candidate is tried and conflicting resolutions are dropped.
	arrayProf := map[int]uint32{} // actor -> overlay-space profile address
	for id := 0; id < numActors; id++ {
		pp := le.Uint32(ls.arm9[profArrayOff+id*4:])
		if pp >= 0x020AA420 && pp < 0x02150000 {
			arrayProf[id] = pp
		}
	}
	for ovl := 60; ovl <= 102; ovl++ {
		if _, ok := ls.ovls[ovl]; !ok {
			continue
		}
		d, err := ls.overlayData(ovl)
		if err != nil {
			continue
		}
		bank := &abin{d, ls.ovls[ovl].RAMAddr}
		profs := rttiProfiles(bank)
		if len(profs) == 0 {
			continue
		}
		creates := map[int]uint32{}
		for id, pr := range profs {
			if seen, ok := classOf[id]; ok && seen != pr.name {
				ambiguous[id] = true
				continue
			}
			classOf[id] = pr.name
			creates[id] = pr.fn
		}
		for id, pp := range arrayProf {
			if _, have := creates[id]; have {
				continue
			}
			if !bank.has(pp) || !bank.has(pp+4) {
				continue
			}
			fn := bank.u32(pp)
			if !isCodeAddr(fn) || int(le.Uint16(bank.data[pp+4-bank.base:])) != id {
				continue
			}
			if seen, ok := classOf[id]; ok && seen != "" {
				_ = seen
			}
			creates[id] = fn
		}
		bins := append(append([]*abin{}, shared...), bank)
		for id, stems := range ls.traceModels(bins, creates) {
			if prev, dup := merged[id]; dup {
				if len(prev) > 0 && len(stems) > 0 && prev[0] != stems[0] {
					ambiguous[id] = true
				}
				continue
			}
			merged[id] = stems
		}
	}
	for id := range ambiguous {
		delete(merged, id)
	}
	return merged, nil
}

// traceModels runs the two binding passes over one bin set: the file-handle
// slot registrations, then each create function's body/callee/vtable scan.
func (ls *LevelSet) traceModels(bins []*abin, creates map[int]uint32) ActorModels {
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
	isCode := isCodeAddr
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
	return out
}

// die helper for commands
func Die(err error) {
	fmt.Fprintln(os.Stderr, "sm64ds:", err)
	os.Exit(1)
}
