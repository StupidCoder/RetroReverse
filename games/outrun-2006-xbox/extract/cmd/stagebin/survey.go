package main

// surveyAll checks the cheap structural invariants of the non-container stage
// files across every /Stage folder: the leading size word, the COLI magic, the
// oso record grid, the fog/sun section offsets. It prints one line per shape
// class, aggregated, so a claim like "every oso file is 40-byte records under
// a size word" rests on the whole disc rather than one folder.

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func surveyAll(disc *xbox.Image, read func(string) []byte) {
	classes := map[string][]string{}
	add := func(class, note string) { classes[class] = append(classes[class], note) }
	disc.Walk(func(e xbox.Entry) error {
		if e.IsDir || !strings.HasPrefix(e.Path, "/Stage/") || strings.HasSuffix(e.Path, "_pmt.sz") {
			return nil
		}
		name := e.Path[strings.LastIndex(e.Path, "/")+1:]
		data := read(e.Path)
		u32 := func(o int) uint32 {
			if o+4 > len(data) {
				return 0xDEAD
			}
			return binary.LittleEndian.Uint32(data[o:])
		}
		sizeOK := int(u32(0)) == len(data)-4
		switch {
		case strings.HasPrefix(name, "coli_"):
			magic := ""
			if len(data) >= 12 {
				magic = string(data[4:12])
			}
			add("coli", fmt.Sprintf("size+4=%v magic=%q", sizeOK, magic))
		case strings.HasPrefix(name, "maya_spl_"):
			add("maya_spl", fmt.Sprintf("len=%d w0=%#x", len(data), u32(0)))
		case strings.HasPrefix(name, "scn_env_"):
			add("scn_env", fmt.Sprintf("len=%d hdr={%#x,%#x,%#x,%#x}", len(data), u32(0), u32(4), u32(8), u32(12)))
		case strings.HasPrefix(name, "oso_"):
			// size word, then 40-byte records; a -99999.9 sextet terminator.
			nRec := 0
			ok := sizeOK
			for o := 4; o+40 <= len(data); o += 40 {
				if u32(o) == 0xC7C34FF3 {
					break
				}
				if u32(o) == 0 && u32(o+4) == 0 && u32(o+8) == 0 && u32(o+12) == 0 {
					break
				}
				nRec++
			}
			add("oso", fmt.Sprintf("size+4=%v recs=%d", ok, nRec))
		case strings.HasSuffix(name, "_bin.sz"):
			// cs_CS/cs_ENV placement (?) tables: {size, 0, 0, hdrSize, ...}
			add("cs_bin", fmt.Sprintf("size+4=%v w1=%#x w2=%#x hdrSize=%#x", sizeOK, u32(4), u32(8), u32(12)))
		case strings.HasSuffix(name, "_xmt.sz"):
			add("xmt", fmt.Sprintf("len=%d w0=%#x w1=%#x", len(data), u32(0), u32(4)))
		default:
			add("other:"+name, fmt.Sprintf("len=%d w0=%#x", len(data), u32(0)))
		}
		return nil
	})
	keys := make([]string, 0, len(classes))
	for k := range classes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		notes := classes[k]
		uniq := map[string]int{}
		for _, n := range notes {
			uniq[n]++
		}
		fmt.Printf("%s: %d files, %d distinct shapes\n", k, len(notes), len(uniq))
		us := make([]string, 0, len(uniq))
		for u := range uniq {
			us = append(us, u)
		}
		sort.Strings(us)
		max := 12
		for _, u := range us {
			if max == 0 {
				fmt.Printf("    ... (%d more)\n", len(us)-12)
				break
			}
			fmt.Printf("    %3dx %s\n", uniq[u], u)
			max--
		}
	}
}
