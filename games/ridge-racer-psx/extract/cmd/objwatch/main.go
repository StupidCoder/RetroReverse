// objwatch enumerates every code path that draws an OBJ.RRO object. It boots
// Ridge Racer under the oracle, runs to a window, and read-watches the OBJ.RRO
// directory (0x800C291C, 319 × 16-byte entries): any draw of object N must
// fetch its geometry pointer from entry N's +12 word, so the set of reading
// PCs is the set of draw sites — independent of which inline record walker
// each site uses. For every +12 fetch it records the PC, the caller ($ra) and
// the object id; other entry offsets (the record counts) are tallied per PC.
//
// Usage:
//
//	objwatch -image disc.bin [-press SCRIPT] [-warmup N] [-window N] [-shots DIR]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/platform/psx"
)

const (
	dirBase  = 0x000C291C // physical address of OBJ.RRO directory entry 0
	dirCount = 319
)

type site struct {
	pc    uint32
	ids   map[int]int    // object id -> fetch count
	ras   map[uint32]int // caller $ra -> count
	first uint64         // step of first fetch
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "objwatch: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	press := flag.String("press", "", "pad script (psx.ParsePress syntax)")
	warmup := flag.Uint64("warmup", 429_500_000, "steps before the watch window")
	window := flag.Uint64("window", 2_000_000, "steps to watch")
	shotEvery := flag.Uint64("shotevery", 0, "during the window, save a screenshot every N steps")
	shots := flag.String("shots", "", "with -shotevery, PNG directory")
	flag.Parse()
	if *image == "" {
		die("need -image")
	}
	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	_, exe, err := vol.BootEXE()
	if err != nil {
		die("%v", err)
	}
	m := psx.NewMachine()
	m.SetDisc(vol)
	m.ISRHandler = 0x8004DF48
	if *press != "" {
		script, err := psx.ParsePress(*press)
		if err != nil {
			die("press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)

	fmt.Fprintf(os.Stderr, "[objwatch] warmup %d steps...\n", *warmup)
	m.Run(*warmup)

	type siteKey struct{ pc, ra uint32 }
	sites := map[siteKey]*site{} // geometry-pointer (+12) fetches, by (pc, caller)
	countPC := map[uint32]int{}  // record-count (+0..+10) reads
	var steps uint64             // steps into the window (approximate: counted per fetch)
	m.RWatchLo, m.RWatchHi = 0x800C291C, 0x800C291C+dirCount*16
	m.OnRead = func(addr, val, pc uint32) {
		off := addr - dirBase
		if off%16 != 12 {
			countPC[pc]++
			return
		}
		id := int(off / 16)
		k := siteKey{pc, m.CPU.Reg(31)}
		s := sites[k]
		if s == nil {
			s = &site{pc: pc, ids: map[int]int{}, ras: map[uint32]int{}, first: steps}
			sites[k] = s
		}
		s.ids[id]++
		s.ras[k.ra]++
	}
	if *shotEvery > 0 && *shots != "" {
		os.MkdirAll(*shots, 0o755)
		var done uint64
		for done < *window {
			m.Run(*shotEvery)
			done += *shotEvery
			steps = done
			p := fmt.Sprintf("%s/w%011d.png", *shots, *warmup+done)
			if err := m.Screenshot(p); err != nil {
				die("screenshot: %v", err)
			}
		}
	} else {
		m.Run(*window)
	}
	m.OnRead = nil

	var keys []siteKey
	for k := range sites {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ra != keys[j].ra {
			return keys[i].ra < keys[j].ra
		}
		return keys[i].pc < keys[j].pc
	})
	fmt.Printf("geometry-pointer fetch sites (%d caller/site pairs):\n", len(keys))
	for _, k := range keys {
		s := sites[k]
		total := 0
		var ids []int
		for id, n := range s.ids {
			ids = append(ids, id)
			total += n
		}
		sort.Ints(ids)
		fmt.Printf("  ra %08X -> site %08X  %6d fetches  ids:", k.ra, k.pc, total)
		for _, id := range ids {
			fmt.Printf(" %d(%d)", id, s.ids[id])
		}
		fmt.Println()
	}
	var cpcs []uint32
	for pc := range countPC {
		cpcs = append(cpcs, pc)
	}
	sort.Slice(cpcs, func(i, j int) bool { return cpcs[i] < cpcs[j] })
	fmt.Printf("record-count read sites (%d):\n", len(cpcs))
	for _, pc := range cpcs {
		fmt.Printf("  PC %08X  %6d reads\n", pc, countPC[pc])
	}
}
