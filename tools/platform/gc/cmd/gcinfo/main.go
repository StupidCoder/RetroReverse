// gcinfo is the GameCube's static inspector: it reads a disc without running it.
//
// It answers the questions a boot oracle should not have to — what the disc calls
// itself, where its loader and executable and filesystem are, what the executable's
// segments are and where they load, and what files the disc carries — so that the
// oracle can be about execution and nothing else.
//
// Usage:
//
//	gcinfo -image DISC.iso [-tree] [-files] [-dol] [-grep PAT] [-at OFF] [-md5]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "disc image (.iso / .gcm)")
	tree := flag.Bool("tree", false, "print the filesystem as a tree")
	files := flag.Bool("files", false, "list every file with its extent")
	showDOL := flag.Bool("dol", false, "describe the executable's segments")
	grep := flag.String("grep", "", "list files whose path contains this")
	at := flag.String("at", "", "name the file whose extent contains this byte offset (hex)")
	sum := flag.Bool("md5", false, "hash the image (reads the whole disc)")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: gcinfo -image DISC.iso [-tree] [-files] [-dol] [-grep PAT] [-at OFF] [-md5]")
		os.Exit(2)
	}
	if err := run(*image, *tree, *files, *showDOL, *grep, *at, *sum); err != nil {
		fmt.Fprintln(os.Stderr, "gcinfo:", err)
		os.Exit(1)
	}
}

func run(image string, tree, list, showDOL bool, grep, at string, sum bool) error {
	d, err := gc.Open(image)
	if err != nil {
		return err
	}
	defer d.Close()
	h := d.Header

	fmt.Printf("image        %s (%d bytes)\n", image, d.Size)
	if sum {
		m, err := d.MD5()
		if err != nil {
			return err
		}
		fmt.Printf("md5          %s\n", m)
	}
	fmt.Printf("game id      %s\n", h.GameID)
	fmt.Printf("title        %s\n", h.Title)
	fmt.Printf("disc/version %d / %d\n", h.DiscID, h.Version)
	fmt.Printf("apploader    %s  entry 0x%08X  body %d bytes (0x%x code + 0x%x trailer)\n",
		d.Apploader.Date, d.Apploader.Entry, d.Apploader.Body(), d.Apploader.Size, d.Apploader.TrailerSize)
	fmt.Printf("executable   0x%08X on disc\n", h.DOLOffset)
	fmt.Printf("filesystem   0x%08X on disc, %d bytes (%d entries, %d files); loads at 0x%08X\n",
		h.FSTOffset, h.FSTSize, len(d.FST.Entries), len(d.FST.Files()), h.FSTAddr)
	fmt.Printf("user data    0x%08X on disc, 0x%X bytes\n", h.UserOffset, h.UserSize)

	dol, err := d.DOL()
	if err != nil {
		return err
	}
	fmt.Printf("entry point  0x%08X\n", dol.Entry)

	// The extent check is the real proof the filesystem was read correctly: a misparse
	// produces wild offsets, and wild offsets run off the disc or collide with a neighbour.
	if err := d.FST.Validate(d.Size); err != nil {
		fmt.Printf("filesystem   INVALID: %v\n", err)
	} else {
		fmt.Printf("filesystem   valid: every extent is inside the image and none overlap\n")
	}

	if showDOL {
		fmt.Printf("\nexecutable (%d segments, entry 0x%08X)\n", len(dol.Segments), dol.Entry)
		for _, s := range dol.Segments {
			fmt.Printf("  %-6s 0x%08X..0x%08X  %8d bytes  (at 0x%08X on disc)\n",
				s.Name(), s.Addr, s.Addr+s.Size, s.Size, h.DOLOffset+s.Offset)
		}
		fmt.Printf("  %-6s 0x%08X..0x%08X  %8d bytes\n", "bss", dol.BSSAddr, dol.BSSAddr+dol.BSSSize, dol.BSSSize)
	}

	if at != "" {
		off, err := strconv.ParseInt(strings.TrimPrefix(at, "0x"), 16, 64)
		if err != nil {
			return fmt.Errorf("-at: %w", err)
		}
		f, within, ok := d.FST.ByOffset(off)
		if !ok {
			fmt.Printf("\n0x%X is in no file\n", off)
		} else {
			fmt.Printf("\n0x%X is %s + 0x%X (of %d bytes)\n", off, f.Path, within, f.Size)
		}
	}

	if grep != "" {
		fmt.Println()
		for _, f := range d.FST.Files() {
			if strings.Contains(strings.ToLower(f.Path), strings.ToLower(grep)) {
				fmt.Printf("%10d  0x%08X  %s\n", f.Size, f.Offset, f.Path)
			}
		}
	}

	if list {
		fmt.Println()
		fs := d.FST.Files()
		sort.Slice(fs, func(i, j int) bool { return fs[i].Offset < fs[j].Offset })
		var total int64
		for _, f := range fs {
			fmt.Printf("%10d  0x%08X  %s\n", f.Size, f.Offset, f.Path)
			total += f.Size
		}
		fmt.Printf("\n%d files, %d bytes\n", len(fs), total)
	}

	if tree {
		fmt.Println()
		printTree(d.FST)
	}
	return nil
}

// printTree walks the flat entry array back into the hierarchy it encodes. Depth is
// the path's own slash count, which needs no stack of its own.
func printTree(f *gc.FST) {
	for _, e := range f.Entries[1:] {
		depth := strings.Count(e.Path, "/") - 1
		indent := strings.Repeat("  ", depth)
		if e.Dir {
			fmt.Printf("%s%s/  (%d entries)\n", indent, e.Name, e.Size)
		} else {
			fmt.Printf("%s%-*s  %9d  0x%08X\n", indent, 40-len(indent), e.Name, e.Size, e.Offset)
		}
	}
}
