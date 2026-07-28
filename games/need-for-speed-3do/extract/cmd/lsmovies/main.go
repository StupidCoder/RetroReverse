// lsmovies is a scratch probe: lists Movies/ entries on the disc, or with
// -chunks dumps the chunk structure of one .stream (used to reverse the SNDS
// audio track layout).
package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	chunks := flag.String("chunks", "", "dump chunk structure of this stream path on the disc")
	sdx2 := flag.String("sdx2", "", "decode the -chunks stream's SNDS audio to this s16le file")
	shot := flag.String("shot", "", "decode the -chunks stream and write frame -shotn as this PNG")
	shotn := flag.Int("shotn", 40, "frame number for -shot")
	max := flag.Int("n", 40, "max chunks to dump")
	doSums := flag.Bool("sums", false, "print SSMP capacity vs declared byteCount sums for every movie")
	flag.Parse()
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		panic(err)
	}
	if *chunks != "" {
		raw, err := vol.ReadFile(*chunks)
		if err != nil {
			panic(err)
		}
		if *sdx2 != "" {
			if err := sdx2ToFile(raw, *sdx2); err != nil {
				panic(err)
			}
			return
		}
		if *shot != "" {
			if err := shotFrame(raw, *shotn, *shot); err != nil {
				panic(err)
			}
			return
		}
		dump(raw, *max)
		return
	}
	type ent struct {
		p string
		n int
	}
	var list []ent
	vol.Walk(func(e threedo.Entry) error {
		if !e.IsDir && strings.Contains(e.Path, "Movies") {
			list = append(list, ent{e.Path, int(e.Size)})
		}
		return nil
	})
	sort.Slice(list, func(i, j int) bool { return list[i].p < list[j].p })
	if *doSums {
		var paths []string
		for _, e := range list {
			paths = append(paths, e.p)
		}
		printSums(vol, paths)
		return
	}
	for _, e := range list {
		fmt.Printf("%9d  %s\n", e.n, e.p)
	}
}

func dump(data []byte, max int) {
	counts := map[string]int{}
	shown := 0
	for off := 0; off+8 <= len(data); {
		tag := string(data[off : off+4])
		size := int(binary.BigEndian.Uint32(data[off+4 : off+8]))
		if size < 8 || off+size > len(data) {
			fmt.Printf("%08x  BAD %q size=%d\n", off, tag, size)
			break
		}
		sub := ""
		if off+20 <= len(data) {
			sub = string(data[off+16 : off+20])
		}
		counts[tag+"/"+sub]++
		if shown < max {
			n := 48
			if off+8+n > len(data) {
				n = len(data) - off - 8
			}
			fmt.Printf("%08x  %s size=%-7d sub=%s  %s\n", off, tag, size, sub,
				hex.EncodeToString(data[off+8:off+8+n]))
			shown++
		}
		off += size
	}
	fmt.Println("--- chunk census ---")
	var keys []string
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%-12s %d\n", k, counts[k])
	}
}
