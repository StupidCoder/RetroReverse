// dgo reads the disc's DGO/CGO archives and reproduces the game's object
// linker offline (goalobj).
//
// Usage:
//
//	dgo -image DISC.iso -list                          list every archive and its objects
//	dgo -image DISC.iso -archive DGO/TIT.DGO -list     list one archive
//	dgo -image DISC.iso -archive DGO/TIT.DGO -entry logo-cam [-nth 0] -raw FILE
//	    write the raw (unlinked) payload
//	dgo -image DISC.iso -archive DGO/TIT.DGO -entry logo-cam -base 0x1234560 \
//	    -symtab ../work/goal.txt -out FILE [-report]
//	    link the object at the given runtime base and write the image
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/goalobj"
	"retroreverse.com/tools/lib/iso9660"
)

func parseHex(s string) (uint32, error) {
	s = strings.TrimPrefix(s, "0x")
	v, err := strconv.ParseUint(s, 16, 32)
	return uint32(v), err
}

func main() {
	image := flag.String("image", "", "disc image (.iso)")
	archive := flag.String("archive", "", "archive path on the disc (e.g. DGO/TIT.DGO)")
	list := flag.Bool("list", false, "list archives/objects")
	entry := flag.String("entry", "", "object name inside the archive")
	nth := flag.Int("nth", 0, "which occurrence of -entry (archives may repeat a name)")
	raw := flag.String("raw", "", "write the raw payload to FILE")
	baseS := flag.String("base", "", "runtime base address to link at (hex)")
	symtab := flag.String("symtab", "", "GOAL symbol table dump (goal.txt / -goalsyms output)")
	out := flag.String("out", "", "write the linked image to FILE")
	report := flag.Bool("report", false, "print the relocation report")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "dgo: -image is required")
		os.Exit(1)
	}
	f, err := os.Open(*image)
	check(err)
	st, err := f.Stat()
	check(err)
	vol, err := iso9660.Open(f, st.Size())
	check(err)

	if *archive == "" && *list {
		check(vol.Walk(func(e iso9660.Entry) error {
			p := strings.TrimSuffix(e.Path, ";1")
			if strings.HasSuffix(p, ".DGO") || strings.HasSuffix(p, ".CGO") {
				listArchive(vol, e.Path)
			}
			return nil
		}))
		return
	}
	if *archive == "" {
		fmt.Fprintln(os.Stderr, "dgo: -archive or -list is required")
		os.Exit(1)
	}
	path := *archive
	if !strings.Contains(path, ";") {
		path += ";1"
	}
	if *list {
		listArchive(vol, path)
		return
	}

	data, err := vol.ReadFile(path)
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	var obj []byte
	seen := 0
	for _, e := range d.Entries {
		if e.Name == *entry {
			if seen == *nth {
				obj = e.Data
				break
			}
			seen++
		}
	}
	if obj == nil {
		fmt.Fprintf(os.Stderr, "dgo: %s has no entry %q (occurrence %d)\n", d.Name, *entry, *nth)
		os.Exit(1)
	}
	if *raw != "" {
		check(os.WriteFile(*raw, obj, 0o644))
		fmt.Printf("%s: wrote %d raw bytes to %s\n", *entry, len(obj), *raw)
		return
	}

	base, err := parseHex(*baseS)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dgo: -base is required for linking (hex)")
		os.Exit(1)
	}
	tab, err := goalobj.LoadSymTab(*symtab)
	check(err)
	linked, rep, err := goalobj.Link(obj, base, tab)
	check(err)
	if *out != "" {
		check(os.WriteFile(*out, linked, 0o644))
	}
	fmt.Printf("%s: v%d, %d bytes linked at 0x%08X, %d pointer fixups, %d symbol patches\n",
		*entry, rep.Version, len(linked), base, len(rep.Pointers), len(rep.SymbolRef))
	if len(rep.Missing) > 0 {
		fmt.Printf("  MISSING symbols: %s\n", strings.Join(rep.Missing, ", "))
	}
	if *report {
		offs := make([]int, 0, len(rep.SymbolRef))
		for o := range rep.SymbolRef {
			offs = append(offs, int(o))
		}
		sort.Ints(offs)
		for _, o := range offs {
			fmt.Printf("  +0x%06x -> %s\n", o, rep.SymbolRef[uint32(o)])
		}
		if len(rep.Pointers) > 0 {
			fmt.Printf("  pointer fixups at:")
			for _, p := range rep.Pointers {
				fmt.Printf(" 0x%x", p)
			}
			fmt.Println()
		}
	}
}

func listArchive(vol *iso9660.Volume, path string) {
	data, err := vol.ReadFile(path)
	check(err)
	d, err := goalobj.ReadDGO(data)
	check(err)
	fmt.Printf("%s (%q): %d objects\n", path, d.Name, len(d.Entries))
	for _, e := range d.Entries {
		ver := "?"
		if len(e.Data) >= 12 {
			ver = fmt.Sprint(uint32(e.Data[8]) | uint32(e.Data[9])<<8)
		}
		fmt.Printf("  %8d  v%-2s %s\n", len(e.Data), ver, e.Name)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "dgo:", err)
		os.Exit(1)
	}
}
