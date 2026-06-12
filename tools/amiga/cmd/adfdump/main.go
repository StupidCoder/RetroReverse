// adfdump lists the contents of an AmigaDOS floppy image (ADF) and can extract
// its files. This is the usual first step when reverse engineering an Amiga
// game disk: it shows the filesystem type, the volume name and the directory
// tree, and -x writes every file out preserving the directory structure.
//
// Usage:
//
//	adfdump disk.adf            list the volume and its files
//	adfdump -x outdir disk.adf  also extract all files into outdir/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"stupidcoder.com/tools/amiga/adf"
)

func main() {
	out := flag.String("x", "", "extract all files into this directory")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: adfdump [-x outdir] disk.adf")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	image, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "adfdump:", err)
		os.Exit(1)
	}
	vol, err := adf.Open(image)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adfdump:", err)
		os.Exit(1)
	}

	fs := "OFS"
	if vol.FFS {
		fs = "FFS"
	}
	fmt.Printf("Volume %q  (%s%s%s, %d blocks, root checksum ok=%v)\n",
		vol.Name, fs, flagStr(" intl", vol.Intl), flagStr(" dircache", vol.Dircache),
		blocks(image), vol.ChecksumOK(blocks(image)/2))

	var files, dirs int
	err = vol.Walk(func(e adf.Entry) error {
		fmt.Println(" ", e)
		if e.IsDir {
			dirs++
			return nil
		}
		files++
		if *out == "" {
			return nil
		}
		data, err := vol.ReadFile(e.Path)
		if err != nil {
			return fmt.Errorf("%s: %w", e.Path, err)
		}
		dst := filepath.Join(*out, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "adfdump:", err)
		os.Exit(1)
	}
	fmt.Printf("%d files, %d directories\n", files, dirs)
	if *out != "" {
		fmt.Printf("extracted to %s/\n", *out)
	}
}

func blocks(image []byte) int { return len(image) / 512 }

func flagStr(s string, on bool) string {
	if on {
		return s
	}
	return ""
}
