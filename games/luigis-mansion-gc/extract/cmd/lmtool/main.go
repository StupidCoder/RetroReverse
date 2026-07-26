// lmtool works on Luigi's Mansion's own files: it pulls a file out of the disc
// by its path, undoes the Yay0 compression the .szp files wear, and inspects
// what is inside.
//
// Usage:
//
//	lmtool -image DISC.iso -cat /Ajioka/ADemo/opeg.szp > opeg.bin   (decompressed)
//	lmtool -image DISC.iso -raw /Ajioka/ADemo/opeg.szp > opeg.szp   (as stored)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "disc image (.iso / .gcm)")
	cat := flag.String("cat", "", "write this disc file to stdout, decompressed if it is Yay0")
	raw := flag.String("raw", "", "write this disc file to stdout exactly as stored")
	ls := flag.String("ls", "", "list the RARC archive inside this disc file")
	x := flag.String("x", "", "extract the RARC archive inside this disc file into -into DIR")
	into := flag.String("into", ".", "directory -x extracts into")
	mdl := flag.String("mdl", "", "export a model as GLB: /disc/file.szp:member.mdl")
	anim := flag.String("anim", "", "with -mdl: member.key animation in the same archive; makes the export skinned+animated")
	out := flag.String("o", "out.glb", "output path for -mdl")
	pose := flag.Float64("pose", -1, "with -mdl/-anim: print evaluated node matrices at this frame instead of exporting (flat and tree-composed), for comparing against the game's RAM")
	inplace := flag.Bool("inplace", false, "with -mdl/-anim: pin the root's x/z translation at frame 0 (walk clips stay under the camera)")
	flag.Parse()

	if *mdl != "" && *anim != "" && *pose >= 0 {
		if err := poseDump(*image, *mdl, *anim, float32(*pose)); err != nil {
			fmt.Fprintln(os.Stderr, "lmtool:", err)
			os.Exit(1)
		}
		return
	}
	if *mdl != "" && *anim != "" {
		if err := skinnedExport(*image, *mdl, *anim, *out, *inplace); err != nil {
			fmt.Fprintln(os.Stderr, "lmtool:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", *out)
		return
	}

	if *mdl != "" {
		m, name, err := mdlFromArchive(*image, *mdl)
		if err == nil {
			err = staticGLB(m, *out, name)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "lmtool:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", *out)
		return
	}

	if *ls != "" || *x != "" {
		if err := runArchive(*image, *ls, *x, *into); err != nil {
			fmt.Fprintln(os.Stderr, "lmtool:", err)
			os.Exit(1)
		}
		return
	}
	if *image == "" || (*cat == "" && *raw == "") {
		fmt.Fprintln(os.Stderr, "usage: lmtool -image DISC.iso (-cat|-raw) /path/on/disc")
		os.Exit(2)
	}
	if err := run(*image, *cat, *raw); err != nil {
		fmt.Fprintln(os.Stderr, "lmtool:", err)
		os.Exit(1)
	}
}

func discFile(image, path string) ([]byte, error) {
	d, err := gc.Open(image)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	for _, f := range d.FST.Entries {
		if !f.Dir && strings.EqualFold(f.Path, path) {
			return d.Read(f.Offset, int(f.Size))
		}
	}
	return nil, fmt.Errorf("no file %q on the disc", path)
}

func runArchive(image, ls, x, into string) error {
	path := ls
	if path == "" {
		path = x
	}
	b, err := discFile(image, path)
	if err != nil {
		return err
	}
	if len(b) >= 4 && string(b[:4]) == "Yay0" {
		if b, err = lm.Yay0(b); err != nil {
			return err
		}
	}
	files, err := lm.RARC(b)
	if err != nil {
		return err
	}
	for _, f := range files {
		p := f.Name
		if f.Dir != "" {
			p = f.Dir + "/" + f.Name
		}
		if x != "" {
			dst := filepath.Join(into, filepath.FromSlash(p))
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(dst, f.Data, 0o644); err != nil {
				return err
			}
		}
		fmt.Printf("%8d  %s\n", len(f.Data), p)
	}
	return nil
}

func run(image, cat, raw string) error {
	d, err := gc.Open(image)
	if err != nil {
		return err
	}
	defer d.Close()

	path := cat
	if path == "" {
		path = raw
	}
	var found *gc.File
	for i, f := range d.FST.Entries {
		if !f.Dir && strings.EqualFold(f.Path, path) {
			found = &d.FST.Entries[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("no file %q on the disc", path)
	}
	b, err := d.Read(found.Offset, int(found.Size))
	if err != nil {
		return err
	}
	if cat != "" && len(b) >= 4 && string(b[:4]) == "Yay0" {
		if b, err = lm.Yay0(b); err != nil {
			return err
		}
	}
	_, err = os.Stdout.Write(b)
	return err
}
