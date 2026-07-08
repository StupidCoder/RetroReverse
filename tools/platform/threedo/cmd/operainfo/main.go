// operainfo inspects a 3DO Opera CD image: it prints the volume label, lists the
// file system, and extracts individual files. This is the first step in reverse
// engineering a 3DO disc.
//
// Usage:
//
//	operainfo -label image.bin                    dump the volume label fields
//	operainfo -ls image.bin                       list the whole file tree
//	operainfo -extract path/to/file -o out image.bin   extract one file
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	label := flag.Bool("label", false, "print the Opera volume label fields")
	ls := flag.Bool("ls", false, "list every file and directory")
	extract := flag.String("extract", "", "extract this file path from the image")
	out := flag.String("o", "", "output file for -extract (default stdout)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: operainfo [-label] [-ls] [-extract PATH -o FILE] image.bin")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	image, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die(err)
	}
	vol, err := threedo.Open(image)
	if err != nil {
		die(err)
	}

	switch {
	case *extract != "":
		data, err := vol.ReadFile(*extract)
		if err != nil {
			die(err)
		}
		if *out == "" || *out == "-" {
			os.Stdout.Write(data)
		} else if err := os.WriteFile(*out, data, 0o644); err != nil {
			die(err)
		} else {
			fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(data), *out)
		}
	case *label:
		fmt.Printf("Opera volume\n")
		fmt.Printf("  label   = %q\n", vol.Label)
		fmt.Printf("  comment = %q\n", vol.Comment)
		fmt.Printf("  id      = 0x%08X\n", vol.ID)
	default: // -ls or no flag
		_ = ls
		fmt.Printf("Opera disc  label=%q id=0x%08X\n", vol.Label, vol.ID)
		var files, dirs int
		err = vol.Walk(func(e threedo.Entry) error {
			fmt.Println(" ", e)
			if e.IsDir {
				dirs++
			} else {
				files++
			}
			return nil
		})
		if err != nil {
			die(err)
		}
		fmt.Printf("%d files, %d directories\n", files, dirs)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "operainfo:", err)
	os.Exit(1)
}
