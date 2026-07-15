// xbeinfo inspects an original-Xbox disc image or a bare XBE executable.
//
// Given -image ISO it lists the XDVDFS file tree, extracts default.xbe from the disc,
// and prints that XBE's header — base, entry point, sections, and the sorted list of
// xboxkrnl.exe ordinals it imports. Given -image FILE.xbe (or -xbe FILE.xbe) it parses
// that XBE directly. This is static, analysis-only tooling: no CPU, no emulation. It
// advances the oracle's first purpose — enumerate and verify a title's assets — and it
// produces the import-ordinal census that scopes the kernel HLE the machine will need.
//
// It mirrors the per-platform readers (gcinfo/pspinfo shape): standard -image flag,
// -extract to pull a file out, -o to name the output.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func main() {
	var (
		image   = flag.String("image", "", "path to an Xbox disc image (.iso) or a bare .xbe")
		xbePath = flag.String("xbe", "", "path to a bare .xbe (overrides -image for XBE parsing)")
		list    = flag.Bool("list", false, "list the XISO file tree (default when -image is a disc)")
		extract = flag.String("extract", "", "extract this disc path (e.g. /default.xbe) to -o")
		out     = flag.String("o", "", "output path for -extract (default: the basename in the CWD)")
		md5sum  = flag.Bool("md5", false, "also print the image MD5 (hashes the whole file)")
	)
	flag.Parse()

	if *image == "" && *xbePath == "" {
		fmt.Fprintln(os.Stderr, "usage: xbeinfo -image DISC.iso [-list] [-extract /path -o out]")
		fmt.Fprintln(os.Stderr, "       xbeinfo -xbe default.xbe")
		os.Exit(2)
	}

	// A bare XBE, either via -xbe or an -image that is not a disc.
	if *xbePath != "" {
		if err := dumpBareXBE(*xbePath); err != nil {
			fatal(err)
		}
		return
	}
	if strings.EqualFold(filepath.Ext(*image), ".xbe") {
		if err := dumpBareXBE(*image); err != nil {
			fatal(err)
		}
		return
	}

	img, err := xbox.Open(*image)
	if err != nil {
		fatal(err)
	}
	defer img.Close()

	fmt.Printf("XISO: %s (%d bytes)\n", *image, img.Size)
	fmt.Printf("  game partition base: %#x  root sector-relative address space\n\n", img.Base)

	if *md5sum {
		h, err := img.MD5()
		if err != nil {
			fatal(err)
		}
		fmt.Printf("  MD5: %s\n\n", h)
	}

	if *extract != "" {
		data, err := img.ReadFile(*extract)
		if err != nil {
			fatal(err)
		}
		dst := *out
		if dst == "" {
			dst = filepath.Base(strings.TrimRight(*extract, "/"))
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("extracted %s -> %s (%d bytes)\n", *extract, dst, len(data))
		return
	}

	// List the tree (always useful context) and then the default.xbe census.
	if *list || true {
		fmt.Println("file tree:")
		var files, dirs int
		var bytes uint64
		if err := img.Walk(func(e xbox.Entry) error {
			fmt.Printf("  %s\n", e)
			if e.IsDir {
				dirs++
			} else {
				files++
				bytes += uint64(e.Size)
			}
			return nil
		}); err != nil {
			fatal(err)
		}
		fmt.Printf("\n  %d files, %d directories, %d bytes total\n\n", files, dirs, bytes)
	}

	// The title's executable: find default.xbe (case-insensitive, anywhere at root).
	xbeData, xbeName, err := findDefaultXBE(img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		return
	}
	x, err := xbox.ParseXBE(xbeData)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("default XBE: %s (%d bytes)\n", xbeName, len(xbeData))
	printXBE(x)
}

func dumpBareXBE(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	x, err := xbox.ParseXBE(data)
	if err != nil {
		return err
	}
	fmt.Printf("XBE: %s (%d bytes)\n", path, len(data))
	printXBE(x)
	return nil
}

func findDefaultXBE(img *xbox.Image) ([]byte, string, error) {
	for _, p := range []string{"/default.xbe", "/DEFAULT.XBE"} {
		if data, err := img.ReadFile(p); err == nil {
			return data, p, nil
		}
	}
	// Search the tree for any default.xbe.
	var found string
	img.Walk(func(e xbox.Entry) error {
		if !e.IsDir && strings.EqualFold(e.Name, "default.xbe") && found == "" {
			found = e.Path
		}
		return nil
	})
	if found == "" {
		return nil, "", fmt.Errorf("no default.xbe on the disc")
	}
	data, err := img.ReadFile(found)
	return data, found, err
}

func printXBE(x *xbox.XBE) {
	variant := "retail"
	if !x.Retail {
		variant = "debug"
	}
	fmt.Printf("  title:  %q  (id %#08x)\n", x.TitleName, x.TitleID)
	fmt.Printf("  base:   %#08x   image size: %#x\n", x.Base, x.ImageSize)
	fmt.Printf("  entry:  %#08x   (%s keys)\n", x.Entry, variant)
	fmt.Printf("  thunks: %#08x\n\n", x.ThunkAddr)

	fmt.Printf("  %-10s %-10s %-10s %-10s %-10s %s\n", "SECTION", "VADDR", "VSIZE", "FILEOFF", "RAWSIZE", "FLAGS")
	for _, s := range x.Sections {
		fmt.Printf("  %-10s %#08x  %#08x  %#08x  %#08x  %s\n",
			trunc(s.Name, 10), s.VAddr, s.VSize, s.RawAddr, s.RawSize, s.FlagString())
	}

	fmt.Printf("\n  xboxkrnl imports: %d ordinals\n", len(x.Ordinals))
	const perLine = 12
	for i, o := range x.Ordinals {
		if i%perLine == 0 {
			fmt.Printf("   ")
		}
		fmt.Printf(" %3d", o)
		if i%perLine == perLine-1 || i == len(x.Ordinals)-1 {
			fmt.Println()
		}
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "xbeinfo:", err)
	os.Exit(1)
}
