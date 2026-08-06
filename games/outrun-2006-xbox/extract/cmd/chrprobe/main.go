// chrprobe — scratch analyzer for the OutRun 2006 character pipeline:
// /Common/bone.bin skeletons, /Chr/CHR_*.bin descriptors, /Anims/mot_*.sz
// motion clips. Dumps and hypothesis tests; the real export lives in carex.
package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func fatal(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "chrprobe: "+f+"\n", a...)
	os.Exit(1)
}

func main() {
	imagePath := flag.String("image", "", "Xbox disc image")
	dump := flag.String("dump", "", "disc path to dump raw (inflating .sz) to -o")
	outPath := flag.String("o", "", "output file for -dump")
	list := flag.String("list", "", "list disc dir (prefix match)")
	mot := flag.String("mot", "", "analyze a dumped motion file (local path)")
	bone := flag.String("bone", "", "analyze a dumped bone.bin (local path)")
	chr := flag.String("chr", "", "analyze a dumped CHR_*.bin descriptor (local path)")
	skel := flag.String("skel", "", "with -bone: dump only this skeleton's records")
	mtab := flag.String("mtab", "", "analyze a dumped motdata_table.bin (local path)")
	mfilter := flag.String("mfilter", "", "with -mtab: only files whose name contains this")
	census := flag.Bool("census", false, "census every *_pmt.sz on the disc (nParts, nTextures)")
	rawvtx := flag.String("rawvtx", "", "dump raw vertices of a pmt part (local .sz or inflated path)")
	rvPart := flag.Int("part", 0, "with -rawvtx: part index")
	rvN := flag.Int("n", 8, "with -rawvtx: vertices per pair")
	cfilter := flag.String("cfilter", "", "with -census: path substring filter")
	verbose := flag.Bool("v", false, "verbose channel dump")
	flag.Parse()

	if *mot != "" {
		motDump(*mot, *verbose)
		return
	}
	if *bone != "" {
		boneDump(*bone, *skel)
		return
	}
	if *chr != "" {
		chrDump(*chr)
		return
	}
	if *rawvtx != "" {
		if *rvPart < 0 {
			partSummary(*rawvtx)
		} else {
			rawVtxDump(*rawvtx, *rvPart, *rvN)
		}
		return
	}
	if *mtab != "" {
		motTableDump(*mtab, *mfilter)
		return
	}

	disc, err := xbox.Open(*imagePath)
	if err != nil {
		fatal("open image: %v", err)
	}
	defer disc.Close()

	if *census {
		censusPMT(disc, *cfilter)
		return
	}
	if *list != "" {
		low := strings.ToLower(*list)
		disc.Walk(func(e xbox.Entry) error {
			if strings.HasPrefix(strings.ToLower(e.Path), low) {
				fmt.Printf("%9d  %s\n", e.Size, e.Path)
			}
			return nil
		})
		return
	}
	if *dump != "" {
		raw, err := disc.ReadFile(*dump)
		if err != nil {
			fatal("read %s: %v", *dump, err)
		}
		data := raw
		if strings.HasSuffix(strings.ToLower(*dump), ".sz") {
			zr, err := zlib.NewReader(bytes.NewReader(raw))
			if err != nil {
				fatal("zlib: %v", err)
			}
			if data, err = io.ReadAll(zr); err != nil {
				fatal("inflate: %v", err)
			}
		}
		out := *outPath
		if out == "" {
			out = filepath.Base(*dump) + ".bin"
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			fatal("%v", err)
		}
		fmt.Printf("%s: %d bytes -> %s\n", *dump, len(data), out)
		return
	}
	fatal("nothing to do")
}
