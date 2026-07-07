// psxinfo inspects a PlayStation CD image: it prints the ISO 9660 volume header,
// lists the filesystem, and extracts individual files. This is the first step in
// reverse engineering a PSX disc — it locates and pulls out the boot executable
// named in SYSTEM.CNF.
//
// Usage:
//
//	psxinfo -ls image.bin                     list the volume and its files
//	psxinfo -pvd image.bin                    dump the volume descriptor fields
//	psxinfo -extract "SCUS_943.00;1" -o exe image.bin   extract one file
//	psxinfo -exe "SCUS_943.00;1" image.bin    dump PS-X EXE header fields
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/psx"
)

func main() {
	ls := flag.Bool("ls", false, "list all files in the volume")
	pvd := flag.Bool("pvd", false, "print the Primary Volume Descriptor fields")
	extract := flag.String("extract", "", "extract this file path from the image")
	exe := flag.String("exe", "", "parse and dump the PS-X EXE header of this file")
	out := flag.String("o", "", "output file for -extract (default stdout)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: psxinfo [-ls] [-pvd] [-extract PATH -o FILE] [-exe PATH] image.bin")
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
	vol, err := psx.Open(image)
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
	case *exe != "":
		data, err := vol.ReadFile(*exe)
		if err != nil {
			die(err)
		}
		e, err := psx.ParseEXE(data)
		if err != nil {
			die(err)
		}
		fmt.Print(e.Describe())
	case *pvd:
		fmt.Printf("system:  %q\n", vol.System)
		fmt.Printf("volume:  %q\n", vol.Name)
	default: // -ls or no flag
		_ = ls
		fmt.Printf("PLAYSTATION disc  system=%q volume=%q\n", vol.System, vol.Name)
		var files, dirs int
		err = vol.Walk(func(e psx.Entry) error {
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
	fmt.Fprintln(os.Stderr, "psxinfo:", err)
	os.Exit(1)
}
