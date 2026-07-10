// romls lists the cartridge archive: one line per resource, with its FORM type,
// cart offset, byte length, and chunk shape. Nothing is emulated — the archive
// is read straight out of the ROM (see extract/pwad).
//
// -calllog cross-checks the static reading against the running game: every
// cartridge address the game's loader ever fetched an asset from must be the
// body of a chunk this reader found. That ties the archive we parse to the
// archive the game reads, which no amount of self-consistency can do.
//
// Usage:
//
//	romls -image ROM [-type UVMD] [-census] [-chunks]
//	romls -image ROM -calllog work/calllog-2100.log
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
)

func main() {
	image := flag.String("image", "", "cartridge image")
	typ := flag.String("type", "", "list only resources of this FORM type")
	census := flag.Bool("census", false, "print the type census and the chunk vocabulary instead of the listing")
	chunks := flag.Bool("chunks", false, "print each resource's chunk list")
	callLog := flag.String("calllog", "", "verify every loader fetch in this log lands on a chunk body")
	flag.Parse()

	rom, err := n64.Load(*image)
	if err != nil {
		die(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}
	if err := a.Check(); err != nil {
		die(err)
	}

	if *callLog != "" {
		if err := verifyLoads(a, *callLog); err != nil {
			die(err)
		}
		return
	}
	if *census {
		printCensus(a)
		return
	}

	fmt.Printf("archive 0x%07X..0x%07X — %d resources\n\n", pwad.ArchiveStart, a.End, len(a.Dir))
	fmt.Printf("%5s  %-4s  %-9s %9s  %s\n", "idx", "type", "offset", "bytes", "chunks")
	for _, f := range a.Forms[1:] {
		if *typ != "" && f.Type != *typ {
			continue
		}
		fmt.Printf("%5d  %-4s  0x%07X %9d  %s\n", f.Index, f.Type, f.Off, f.Total(), shape(f))
		if *chunks {
			for _, c := range f.Chunks {
				if c.Compressed() {
					fmt.Printf("           GZIP %-4s  %6d stored -> %7d\n", c.InnerTag, c.Size, c.UncompressedSize)
				} else {
					fmt.Printf("           %-4s       %6d\n", c.Tag, c.Size)
				}
			}
		}
	}
}

// shape summarises a FORM's chunk list, collapsing runs: "PAD ×2, GZIP(COMM)".
func shape(f pwad.Form) string {
	var parts []string
	var last string
	n := 0
	flush := func() {
		if n == 0 {
			return
		}
		if n > 1 {
			parts = append(parts, fmt.Sprintf("%s×%d", last, n))
		} else {
			parts = append(parts, last)
		}
	}
	for _, c := range f.Chunks {
		name := strings.TrimSpace(c.Tag)
		if c.Compressed() {
			name = "GZIP(" + c.InnerTag + ")"
		}
		if name == last {
			n++
			continue
		}
		flush()
		last, n = name, 1
	}
	flush()
	return strings.Join(parts, ", ")
}

func printCensus(a *pwad.Archive) {
	types := map[string]int{}
	bytes := map[string]uint32{}
	// chunk vocabulary per FORM type; GZIP recorded by its inner tag
	vocab := map[string]map[string]int{}
	for _, f := range a.Forms[1:] {
		types[f.Type]++
		bytes[f.Type] += f.Total()
		if vocab[f.Type] == nil {
			vocab[f.Type] = map[string]int{}
		}
		for _, c := range f.Chunks {
			name := strings.TrimSpace(c.Tag)
			if c.Compressed() {
				name = "GZIP(" + c.InnerTag + ")"
			}
			vocab[f.Type][name]++
		}
	}
	var keys []string
	for k := range types {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return types[keys[i]] > types[keys[j]] })

	fmt.Printf("%-4s %6s %12s   %s\n", "type", "count", "bytes", "chunks")
	for _, k := range keys {
		var cv []string
		for name, n := range vocab[k] {
			cv = append(cv, fmt.Sprintf("%s×%d", name, n))
		}
		sort.Strings(cv)
		fmt.Printf("%-4s %6d %12d   %s\n", k, types[k], bytes[k], strings.Join(cv, " "))
	}
}

// bounceBuffer is where the game stages a compressed blob before inflating it;
// a loader call targeting it names a GZIP chunk, any other cartridge fetch names
// a plain one. loaderChunkPC is the loader's internal 4 KiB-chunking recursion,
// whose calls are an implementation detail, not logical loads.
const (
	bounceBuffer  = 0x000DA800
	loaderChunkPC = 0x8022A7D4
)

// verifyLoads reads a bootoracle -calllog and checks every cartridge fetch
// against the archive. A staged (compressed) fetch must hit a GZIP chunk body;
// a direct fetch inside the archive must hit a plain chunk body. Fetches outside
// the archive are reported, not failed: the program overlay and the audio banks
// legitimately live there.
func verifyLoads(a *pwad.Archive, path string) error {
	gzipBody := map[uint32]bool{}
	rawBody := map[uint32]bool{}
	for _, f := range a.Forms {
		for _, c := range f.Chunks {
			if c.Compressed() {
				gzipBody[c.Off] = true
			} else {
				rawBody[c.Off] = true
			}
		}
	}

	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	staged, direct := map[uint32]bool{}, map[uint32]bool{}
	sc := bufio.NewScanner(fh)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var pc, field, a0, a1, a2, a3, ra uint32
		if _, err := fmt.Sscanf(sc.Text(), "call %x field=%d a0=%x a1=%x a2=%x a3=%x ra=%x",
			&pc, &field, &a0, &a1, &a2, &a3, &ra); err != nil {
			continue
		}
		if pc != 0x8022A760 || ra == loaderChunkPC || a1 >= 0x80000000 {
			continue
		}
		if a0&^0x80000000 == bounceBuffer {
			staged[a1] = true
		} else if a2 > 4 { // a 4-byte fetch is a directory probe, not an asset
			direct[a1] = true
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	var badStaged, badDirect, outside []uint32
	for off := range staged {
		if !gzipBody[off] {
			badStaged = append(badStaged, off)
		}
	}
	for off := range direct {
		switch {
		case off < pwad.ArchiveStart || off >= a.End:
			outside = append(outside, off)
		case !rawBody[off]:
			badDirect = append(badDirect, off)
		}
	}
	sort.Slice(outside, func(i, j int) bool { return outside[i] < outside[j] })

	fmt.Printf("%d staged (compressed) fetches: %d not a GZIP chunk body\n", len(staged), len(badStaged))
	fmt.Printf("%d direct fetches: %d inside the archive but not a chunk body, %d outside the archive\n",
		len(direct), len(badDirect), len(outside))
	for _, off := range outside {
		fmt.Printf("  outside: 0x%07X\n", off)
	}
	if len(badStaged)+len(badDirect) > 0 {
		for _, off := range append(badStaged, badDirect...) {
			fmt.Printf("  MISMATCH: 0x%07X\n", off)
		}
		return fmt.Errorf("%d loader fetches do not land on a chunk body", len(badStaged)+len(badDirect))
	}
	fmt.Println("every loader fetch lands on a chunk body this reader found")
	return nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "romls:", err)
	os.Exit(1)
}
