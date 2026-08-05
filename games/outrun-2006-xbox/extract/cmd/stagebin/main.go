// stagebin explores the non-container (_bin / oso) members of a /Stage
// folder: decompress each and hexdump the head, or dump one whole file.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func main() {
	image := flag.String("image", "", "Xbox disc image")
	dir := flag.String("dir", "/Stage/BEAC", "stage folder to list")
	survey := flag.Bool("survey", false, "survey every non-container /Stage file's structural invariants")
	file := flag.String("file", "", "one disc path: hexdump the whole decompressed file")
	f32s := flag.Bool("f32", false, "with -file: also print each dword as float where plausible")
	n := flag.Int("n", 96, "head bytes to dump per file in -dir mode")
	flag.Parse()
	disc, err := xbox.Open(*image)
	if err != nil {
		panic(err)
	}
	read := func(p string) []byte {
		raw, err := disc.ReadFile(p)
		if err != nil {
			panic(err)
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			fmt.Printf("%s: not zlib (%v), raw %d bytes\n", p, err, len(raw))
			return raw
		}
		data, _ := io.ReadAll(zr)
		return data
	}
	dump := func(data []byte, off, cnt int) {
		for o := off; o < off+cnt && o < len(data); o += 16 {
			end := o + 16
			if end > len(data) {
				end = len(data)
			}
			var hexs, asc []string
			for _, b := range data[o:end] {
				hexs = append(hexs, fmt.Sprintf("%02x", b))
				if b >= 32 && b < 127 {
					asc = append(asc, string(b))
				} else {
					asc = append(asc, ".")
				}
			}
			fmt.Printf("  %06x  %s  %s\n", o, strings.Join(hexs, " "), strings.Join(asc, ""))
		}
	}
	if *survey {
		surveyAll(disc, read)
		return
	}
	if *file != "" {
		data := read(*file)
		fmt.Printf("%s: %d bytes decompressed\n", *file, len(data))
		dump(data, 0, len(data))
		if *f32s {
			fmt.Println("dwords as float:")
			for o := 0; o+4 <= len(data); o += 4 {
				u := binary.LittleEndian.Uint32(data[o:])
				f := math.Float32frombits(u)
				fmt.Printf("  +%04x  %08x  %g\n", o, u, f)
			}
		}
		return
	}
	var files []string
	disc.Walk(func(e xbox.Entry) error {
		if !e.IsDir && strings.HasPrefix(e.Path, *dir+"/") {
			files = append(files, e.Path)
		}
		return nil
	})
	for _, p := range files {
		data := read(p)
		fmt.Printf("\n== %s: %d bytes ==\n", p, len(data))
		dump(data, 0, *n)
	}
}
