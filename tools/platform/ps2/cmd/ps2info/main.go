// ps2info is the PS2's static inspector: it reads a disc without running it.
//
// It answers the questions a boot oracle should not have to — what is on the disc,
// what is the sector geometry, which executable does SYSTEM.CNF boot, where does that
// executable load, and what does its symbol table name — so that the oracle can be
// about execution and nothing else.
//
// Usage:
//
//	ps2info -image DISC.iso [-files] [-elf] [-syms [PATTERN]] [-cat PATH]
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/ps2"
)

func main() {
	image := flag.String("image", "", "disc image (.iso)")
	files := flag.Bool("files", false, "list every file on the disc")
	showELF := flag.Bool("elf", false, "describe the boot executable")
	syms := flag.String("syms", "", "list symbols whose name contains this (use \"*\" for all)")
	cat := flag.String("cat", "", "write a file from the disc to stdout")
	at := flag.Int("at", -1, "name the file whose extent contains this logical block")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: ps2info -image DISC.iso [-files] [-elf] [-syms PAT] [-cat PATH] [-at LBA]")
		os.Exit(2)
	}
	if err := run(*image, *files, *showELF, *syms, *cat, *at); err != nil {
		fmt.Fprintln(os.Stderr, "ps2info:", err)
		os.Exit(1)
	}
}

func run(image string, files, showELF bool, syms, cat string, at int) error {
	f, err := os.Open(image)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}

	geom, err := iso9660.Detect(f, st.Size())
	if err != nil {
		return err
	}
	vol, err := iso9660.Open(f, st.Size())
	if err != nil {
		return err
	}

	if cat != "" {
		b, err := vol.ReadFile(cat)
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, strings.NewReader(string(b)))
		return err
	}

	if at >= 0 {
		e, ok := vol.FileAt(at)
		if !ok {
			fmt.Printf("LBA %d is not inside any file\n", at)
			return nil
		}
		fmt.Printf("LBA %d is inside %s (extent starts at %d, %d bytes; offset 0x%X into the file)\n",
			at, e.Path, e.Block, e.Size, (at-e.Block)*iso9660.BlockSize)
		return nil
	}

	if !files && !showELF && syms == "" {
		// The default: describe the disc.
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}
		raw, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		fmt.Printf("image:    %s\n", image)
		fmt.Printf("size:     %d bytes\n", len(raw))
		fmt.Printf("md5:      %x\n", md5.Sum(raw))
		fmt.Printf("geometry: %v  (%d sectors)\n", geom, st.Size()/int64(geom.SectorSize))
		fmt.Printf("system:   %q\n", vol.System)
		fmt.Printf("volume:   %d logical blocks\n", vol.Blocks)

		nFiles, nDirs, bytes := 0, 0, 0
		vol.Walk(func(e iso9660.Entry) error {
			if e.IsDir {
				nDirs++
			} else {
				nFiles++
				bytes += e.Size
			}
			return nil
		})
		fmt.Printf("contents: %d files in %d directories, %d bytes\n", nFiles, nDirs, bytes)

		cnf, err := vol.ReadFile("SYSTEM.CNF")
		if err == nil {
			fmt.Printf("\nSYSTEM.CNF:\n")
			for _, line := range strings.Split(strings.TrimSpace(string(cnf)), "\n") {
				fmt.Printf("  %s\n", strings.TrimSpace(line))
			}
		}
		return nil
	}

	if files {
		return vol.Walk(func(e iso9660.Entry) error {
			fmt.Println(e)
			return nil
		})
	}

	// Everything below needs the executable.
	exe, err := bootELF(vol)
	if err != nil {
		return err
	}

	if showELF {
		fmt.Print(exe.Describe())
		return nil
	}

	if syms != "" {
		n := 0
		for _, s := range exe.Symbols {
			if syms != "*" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(syms)) {
				continue
			}
			kind := "data"
			if s.Func {
				kind = "func"
			}
			fmt.Printf("0x%08X  %-5s %6d  %s\n", s.Addr, kind, s.Size, s.Name)
			n++
		}
		fmt.Fprintf(os.Stderr, "%d symbols\n", n)
	}
	return nil
}

// bootELF finds and loads the executable SYSTEM.CNF names.
func bootELF(vol *iso9660.Volume) (*ps2.Executable, error) {
	cnf, err := vol.ReadFile("SYSTEM.CNF")
	if err != nil {
		return nil, fmt.Errorf("reading SYSTEM.CNF: %w", err)
	}
	for _, line := range strings.Split(string(cnf), "\n") {
		v, ok := strings.CutPrefix(strings.TrimSpace(line), "BOOT2")
		if !ok {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "="))
		raw, err := vol.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return ps2.LoadELF(raw)
	}
	return nil, fmt.Errorf("SYSTEM.CNF names no BOOT2 executable")
}
