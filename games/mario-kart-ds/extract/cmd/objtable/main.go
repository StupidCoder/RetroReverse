// objtable dumps the game's map-object descriptor table — the table the engine
// walks to turn a course map's OBJI placements into live objects (Part V §2).
//
// The table lives in ARM9 .data at $0216B288: records of {u32 objectID,
// u32 descriptor, u32 auxFn}, terminated by a zero record. It is found by the
// literal pools of the map-object manager ($020D3xxx), which walks it linearly
// comparing each placed object's ID. Each descriptor carries the instance size
// (+0x04) and callback slots; the callbacks' literal pools name the NSBMD/NSBCA
// resources the object loads, which is how the IDs get their names here.
// The decoding lives in mkds.ObjectTable, shared with the placements exporter.
//
//	objtable [-arm9 FILE]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/mario-kart-ds/extract/mkds"
)

func main() {
	arm9 := flag.String("arm9", "../extracted/arm9_dec.bin", "decompressed ARM9 binary")
	flag.Parse()
	data, err := os.ReadFile(*arm9)
	if err != nil {
		fmt.Fprintln(os.Stderr, "objtable:", err)
		os.Exit(1)
	}
	recs := mkds.ObjectTable(data)
	sort.Slice(recs, func(i, j int) bool { return recs[i].ID < recs[j].ID })

	fmt.Printf("map-object table @ $0216B288: %d records\n", len(recs))
	fmt.Printf("%-6s %-10s %-6s %-9s %s\n", "id", "descr", "size", "auxFn", "resources (from callback literal pools)")
	for _, r := range recs {
		desc := "-"
		if r.Desc != 0 {
			desc = fmt.Sprintf("$%08X", r.Desc)
		}
		aux := "-"
		if r.Aux != 0 {
			aux = fmt.Sprintf("$%07X", r.Aux&^1)
		}
		size := "-"
		if r.Size != 0 && r.Size < 0x10000 {
			size = fmt.Sprintf("0x%X", r.Size)
		}
		fmt.Printf("0x%03X  %-10s %-6s %-9s %s\n", r.ID, desc, size, aux, strings.Join(r.Names, " "))
	}
}
