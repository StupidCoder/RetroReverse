// objtable dumps the game's map-object descriptor table — the table the engine
// walks to turn a course map's OBJI placements into live objects (Part V §2).
//
// The table lives in ARM9 .data at $0216B288: records of {u32 objectID,
// u32 descriptor, u32 auxFn}, terminated by a zero record. It is found by the
// literal pools of the map-object manager ($020D3xxx), which walks it linearly
// comparing each placed object's ID. Each descriptor carries the instance size
// (+0x04) and callback slots; the callbacks' literal pools name the NSBMD/NSBCA
// resources the object loads, which is how the IDs get their names here.
//
//	objtable [-arm9 FILE]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	ramBase   = 0x02000000
	tableAddr = 0x0216B288 // {id, descriptor, auxFn} records, zero-terminated
)

var le = binary.LittleEndian

type bin struct{ data []byte }

func (b bin) u32(addr uint32) uint32 {
	off := addr - ramBase
	if int(off)+4 > len(b.data) {
		return 0
	}
	return le.Uint32(b.data[off:])
}

// str returns the NUL-terminated ASCII string at addr, or "".
func (b bin) str(addr uint32) string {
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
func (b bin) poolStrings(fn uint32, budget int, seen map[uint32]bool) []string {
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

func main() {
	arm9 := flag.String("arm9", "../extracted/arm9_dec.bin", "decompressed ARM9 binary")
	flag.Parse()
	data, err := os.ReadFile(*arm9)
	if err != nil {
		fmt.Fprintln(os.Stderr, "objtable:", err)
		os.Exit(1)
	}
	b := bin{data}

	type rec struct {
		id, desc, aux uint32
		size          uint32
		names         []string
	}
	var recs []rec
	for addr := uint32(tableAddr); ; addr += 12 {
		id, desc, aux := b.u32(addr), b.u32(addr+4), b.u32(addr+8)
		if id == 0 && desc == 0 && aux == 0 {
			break
		}
		r := rec{id: id, desc: desc, aux: aux}
		if desc != 0 {
			r.size = b.u32(desc + 4)
			// The descriptor's callback slots (+0x08..+0x2C) hold function
			// pointers (often via a registration trampoline); scan each body's
			// literal pool for the resource names it loads. Trampolines are tiny,
			// so also follow one level of pointers found in their pools.
			seen := map[uint32]bool{}
			for slot := uint32(0x08); slot <= 0x2C; slot += 4 {
				fn := b.u32(desc + slot)
				if fn < ramBase || fn >= 0x021C0000 {
					continue
				}
				r.names = append(r.names, b.poolStrings(fn, 0x100, seen)...)
				// trampoline: its pool holds the real callbacks
				for off := 0; off < 0x40; off += 4 {
					w := b.u32((fn &^ 1) + uint32(off))
					if w >= ramBase && w < 0x02160000 { // code region
						r.names = append(r.names, b.poolStrings(w, 0x200, seen)...)
					}
				}
			}
			r.names = dedup(r.names)
		}
		recs = append(recs, r)
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].id < recs[j].id })

	fmt.Printf("map-object table @ $%08X: %d records\n", uint32(tableAddr), len(recs))
	fmt.Printf("%-6s %-10s %-6s %-9s %s\n", "id", "descr", "size", "auxFn", "resources (from callback literal pools)")
	for _, r := range recs {
		desc := "-"
		if r.desc != 0 {
			desc = fmt.Sprintf("$%08X", r.desc)
		}
		aux := "-"
		if r.aux != 0 {
			aux = fmt.Sprintf("$%07X", r.aux&^1)
		}
		size := "-"
		if r.size != 0 && r.size < 0x10000 {
			size = fmt.Sprintf("0x%X", r.size)
		}
		fmt.Printf("0x%03X  %-10s %-6s %-9s %s\n", r.id, desc, size, aux, strings.Join(r.names, " "))
	}
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
