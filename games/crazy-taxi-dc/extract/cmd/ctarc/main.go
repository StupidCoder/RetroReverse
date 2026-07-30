// ctarc reads Crazy Taxi's asset containers straight off the disc image:
// AFS tables, PVRT textures, and raw 16-bpp pages, plus a census of what
// every container holds — the inventory the format work starts from.
//
//	ctarc -image game.cue -census                      # every AFS: entries, header histogram
//	ctarc -image game.cue -file BINC1.AFS              # list one container
//	ctarc -image game.cue -file BINC1.AFS -x 0:out.bin # extract one entry
//	ctarc -image game.cue -pvr "0GDTEX.PVR;1" -out x.png
//	ctarc -image game.cue -file LANDDC1.AFS -raw16 0:128x128:tile.png [-fmt 1]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/crazy-taxi-dc/extract/assets"
	"retroreverse.com/tools/platform/dc"
)

func main() {
	image_ := flag.String("image", "image/Crazy Taxi (US).cue", "disc image (.cue)")
	file := flag.String("file", "", "disc file to open (AFS or raw)")
	census := flag.Bool("census", false, "walk every .AFS on the disc and histogram entry headers")
	x := flag.String("x", "", "extract AFS entry N[:out]")
	pvr := flag.String("pvr", "", "decode a PVRT texture file from the disc")
	raw16 := flag.String("raw16", "", "decode AFS entry as raw 16bpp: N:WxH:out.png")
	pixFmt := flag.Int("fmt", 1, "raw16/pvr pixel format override: 0=1555 1=565 2=4444")
	out := flag.String("out", "", "output path for -pvr")
	flag.Parse()

	disc, err := dc.OpenDisc(*image_)
	if err != nil {
		die("%v", err)
	}

	switch {
	case *census:
		runCensus(disc)
	case *pvr != "":
		data := readDisc(disc, *pvr)
		t, err := assets.OpenPVR(data)
		if err != nil {
			die("%v", err)
		}
		fmt.Printf("%s: pixfmt %d, type %#x, %dx%d\n", *pvr, t.PixFmt, t.DataType, t.W, t.H)
		img, err := t.Decode()
		if err != nil {
			die("%v", err)
		}
		writePNG(*out, img)
	case *file != "":
		data := readDisc(disc, *file)
		a, err := assets.OpenAFS(data)
		if err != nil {
			die("%s: %v", *file, err)
		}
		switch {
		case *x != "":
			n, path := splitNumPath(*x)
			e, err := a.Data(n)
			if err != nil {
				die("%v", err)
			}
			if path == "" {
				path = fmt.Sprintf("%s.%03d.bin", strings.TrimSuffix(*file, ".AFS"), n)
			}
			if err := os.WriteFile(path, e, 0o644); err != nil {
				die("%v", err)
			}
			fmt.Printf("extracted entry %d (%d bytes) -> %s\n", n, len(e), path)
		case *raw16 != "":
			parts := strings.SplitN(*raw16, ":", 3)
			if len(parts) != 3 {
				die("-raw16 wants N:WxH:out.png")
			}
			var n, w, h int
			fmt.Sscanf(parts[0], "%d", &n)
			fmt.Sscanf(parts[1], "%dx%d", &w, &h)
			e, err := a.Data(n)
			if err != nil {
				die("%v", err)
			}
			img, err := assets.Decode16(e, uint8(*pixFmt), w, h)
			if err != nil {
				die("%v", err)
			}
			writePNG(parts[2], img)
		default:
			fmt.Printf("%s: %d entries\n", *file, len(a.Entries))
			for _, e := range a.Entries {
				d, _ := a.Data(e.Index)
				fmt.Printf("  %4d  %#8x  %8d  %s\n", e.Index, e.Offset, e.Size, headWords(d))
			}
		}
	default:
		flag.Usage()
	}
}

// runCensus opens every .AFS on the disc and histograms the entries' first
// words — the cheap classifier that says which containers share a format.
func runCensus(disc *dc.Disc) {
	var names []string
	entries, err := disc.Vol.ReadDir("/")
	if err != nil {
		die("%v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name, ".AFS") {
			names = append(names, e.Name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		data := readDisc(disc, name)
		a, err := assets.OpenAFS(data)
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
			continue
		}
		hist := map[string]int{}
		var total uint64
		for _, e := range a.Entries {
			d, _ := a.Data(e.Index)
			hist[classify(d)]++
			total += uint64(e.Size)
		}
		fmt.Printf("%s: %d entries, %d data bytes\n", name, len(a.Entries), total)
		type kv struct {
			k string
			n int
		}
		var ks []kv
		for k, n := range hist {
			ks = append(ks, kv{k, n})
		}
		sort.Slice(ks, func(i, j int) bool { return ks[i].n > ks[j].n })
		for _, e := range ks {
			fmt.Printf("   %5d  %s\n", e.n, e.k)
		}
	}
}

// classify names an entry by the shape of its head — enough to group a
// container's contents, honest about being only that.
func classify(d []byte) string {
	if len(d) < 16 {
		return fmt.Sprintf("tiny (%d bytes)", len(d))
	}
	w0 := binary.LittleEndian.Uint32(d)
	w1 := binary.LittleEndian.Uint32(d[4:])
	switch {
	case string(d[:4]) == "PVRT" || string(d[:4]) == "GBIX":
		return "PVRT texture"
	case w0 == 0x80000000 && string(d[4:8]) == "\x00\x00\x00\x00":
		return "ADX?"
	case w0 == 1 && w1 == 1:
		return "model (01 01 + bounds)"
	case w0 == 1:
		return fmt.Sprintf("01 + %08x", w1)
	case allSame(d[:32]):
		return fmt.Sprintf("fill %02x page", d[0])
	}
	return fmt.Sprintf("%08x %08x", w0, w1)
}

func allSame(d []byte) bool {
	for _, b := range d {
		if b != d[0] {
			return false
		}
	}
	return true
}

func headWords(d []byte) string {
	var b strings.Builder
	for i := 0; i+4 <= len(d) && i < 16; i += 4 {
		fmt.Fprintf(&b, "%08x ", binary.LittleEndian.Uint32(d[i:]))
	}
	return b.String()
}

// readDisc reads a file by name, tolerant of the ";1" ISO version suffix.
func readDisc(disc *dc.Disc, name string) []byte {
	for _, cand := range []string{name, name + ";1"} {
		if data, err := disc.Vol.ReadFile(cand); err == nil {
			return data
		}
	}
	die("no file %q on the disc", name)
	return nil
}

func splitNumPath(s string) (int, string) {
	var n int
	path := ""
	if i := strings.IndexByte(s, ':'); i >= 0 {
		fmt.Sscanf(s[:i], "%d", &n)
		path = s[i+1:]
	} else {
		fmt.Sscanf(s, "%d", &n)
	}
	return n, path
}

func writePNG(path string, img *image.RGBA) {
	if path == "" {
		die("no output path (use -out / the :out.png form)")
	}
	f, err := os.Create(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		die("%v", err)
	}
	fmt.Printf("wrote %s (%dx%d)\n", path, img.Rect.Dx(), img.Rect.Dy())
}

func die(f string, args ...any) {
	fmt.Fprintf(os.Stderr, "ctarc: "+f+"\n", args...)
	os.Exit(1)
}
