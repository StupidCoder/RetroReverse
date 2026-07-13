// irxinfo is the IOP's static inspector: it reads the second processor's modules
// off a disc without running them.
//
// The PS2's IOP boots its base kernel from a ROM this repository does not have.
// What the disc has instead is IOPRP221.IMG — a ROMDIR archive carrying the kernel
// modules a game reboots the IOP onto — plus the game's own drivers in /DRIVERS/.
// Between them they are almost the whole of the IOP, and this tool is what says
// *how much* of it, and what is missing.
//
// The last section of the default report is the one that matters. Every module
// declares the libraries it imports and the libraries it exports, so taking the
// difference over the whole disc yields the libraries nothing on the disc provides:
// the exact surface that must be written in Go, function by function, before the
// real IOP can run. It is a work list derived from the image rather than assumed.
//
// Usage:
//
//	irxinfo -image DISC.iso                 # every module, and the Go work list
//	irxinfo -image DISC.iso -mod OVERLORD   # one module in full
//	irxinfo -file OVERLORD.IRX -syms        # a module from a file, with its symbols
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/ps2"
)

func main() {
	image := flag.String("image", "", "disc image (.iso)")
	file := flag.String("file", "", "a single .IRX file, instead of a disc")
	mod := flag.String("mod", "", "describe only the module whose name contains this")
	syms := flag.Bool("syms", false, "list each module's symbols")
	flag.Parse()

	if err := run(*image, *file, *mod, *syms); err != nil {
		fmt.Fprintln(os.Stderr, "irxinfo:", err)
		os.Exit(1)
	}
}

// module is one IRX and where it came from.
type module struct {
	origin string // the path on the disc, or the archive it came out of
	irx    *ps2.IRX
}

func run(image, file, mod string, syms bool) error {
	var mods []module

	switch {
	case file != "":
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		m, err := ps2.ReadIRX(raw)
		if err != nil {
			return err
		}
		mods = append(mods, module{file, m})

	case image != "":
		var err error
		if mods, err = discModules(image); err != nil {
			return err
		}

	default:
		return fmt.Errorf("give -image DISC.iso or -file MODULE.IRX")
	}

	for _, m := range mods {
		if mod != "" && !strings.Contains(strings.ToUpper(m.origin), strings.ToUpper(mod)) {
			continue
		}
		describe(m, syms || mod != "")
	}

	if mod == "" && file == "" {
		workList(mods)
	}
	return nil
}

// discModules gathers every IOP module on a disc: the loose ones in /DRIVERS/, and
// the kernel modules inside the IOP boot image alongside them.
func discModules(image string) ([]module, error) {
	f, err := os.Open(image)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	vol, err := iso9660.Open(f, st.Size())
	if err != nil {
		return nil, err
	}

	var paths []string
	vol.Walk(func(e iso9660.Entry) error {
		if !e.IsDir {
			paths = append(paths, e.Path)
		}
		return nil
	})
	sort.Strings(paths)

	var mods []module
	for _, p := range paths {
		up := strings.ToUpper(p)
		isIRX := strings.Contains(up, ".IRX")
		isIMG := strings.Contains(up, ".IMG")
		if !isIRX && !isIMG {
			continue
		}
		raw, err := vol.ReadFile(p)
		if err != nil {
			return nil, err
		}

		if isIRX {
			m, err := ps2.ReadIRX(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p, err)
			}
			mods = append(mods, module{p, m})
			continue
		}

		// A boot image: an archive of modules.
		entries, err := ps2.ROMDIRModules(raw)
		if err != nil {
			fmt.Printf("%s: not a ROMDIR archive (%v)\n", p, err)
			continue
		}
		for _, e := range entries {
			m, err := ps2.ReadIRX(e.Data)
			if err != nil {
				return nil, fmt.Errorf("%s!%s: %w", p, e.Name, err)
			}
			mods = append(mods, module{p + "!" + e.Name, m})
		}
	}
	return mods, nil
}

func describe(m module, syms bool) {
	x := m.irx
	fmt.Printf("\n%s\n", m.origin)
	fmt.Printf("  %-24s version %d.%d   entry 0x%05X   gp 0x%05X   %d bytes loaded, %d in memory\n",
		x.Name, x.Version>>8, x.Version&0xFF, x.Entry, x.GP, len(x.Image), x.MemSz)

	for _, e := range x.Exports {
		fmt.Printf("    exports %-9s v%d.%d  %d functions\n",
			e.Library, e.Version>>8, e.Version&0xFF, len(e.Entries))
	}
	for _, i := range x.Imports {
		fmt.Printf("    imports %-9s v%d.%d  %v\n", i.Library, i.Version>>8, i.Version&0xFF, i.IDs)
	}
	fmt.Printf("    %d relocations, %d symbols\n", len(x.Relocs), len(x.Symbols))

	if syms && len(x.Symbols) > 0 {
		for _, s := range x.Symbols {
			if s.Func {
				fmt.Printf("      0x%05X  %s\n", s.Addr, s.Name)
			}
		}
	}
}

// workList prints the libraries the disc imports but does not provide.
//
// This is the whole point of the tool. Anything exported by a module on the disc can
// be *run*; anything else has to be written. The difference is the IOP kernel's
// Go surface, and every function in it is listed, because a library is not a unit of
// work — a function is.
func workList(mods []module) {
	provided := map[string]string{} // library -> the module that exports it
	wanted := map[string]map[uint16][]string{}

	for _, m := range mods {
		for _, e := range m.irx.Exports {
			provided[e.Library] = m.irx.Name
		}
	}
	for _, m := range mods {
		for _, i := range m.irx.Imports {
			if wanted[i.Library] == nil {
				wanted[i.Library] = map[uint16][]string{}
			}
			for _, id := range i.IDs {
				wanted[i.Library][id] = append(wanted[i.Library][id], m.irx.Name)
			}
		}
	}

	var missing, onDisc []string
	for lib := range wanted {
		if _, ok := provided[lib]; ok {
			onDisc = append(onDisc, lib)
		} else {
			missing = append(missing, lib)
		}
	}
	sort.Strings(missing)
	sort.Strings(onDisc)

	fmt.Printf("\n\n=== the IOP, as this disc describes it ===\n\n")

	fmt.Printf("%d libraries are provided by modules on the disc, and can be run rather than written:\n  ",
		len(onDisc))
	fmt.Printf("%s\n", strings.Join(onDisc, ", "))

	total := 0
	fmt.Printf("\n%d libraries are imported but provided by nothing on the disc. They are the ROM's,\n"+
		"and they are the Go work list:\n\n", len(missing))
	for _, lib := range missing {
		ids := make([]uint16, 0, len(wanted[lib]))
		for id := range wanted[lib] {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		total += len(ids)

		fmt.Printf("  %-9s %2d functions:", lib, len(ids))
		for _, id := range ids {
			fmt.Printf(" %d", id)
		}
		fmt.Println()
	}
	fmt.Printf("\n%d functions in all.\n", total)
}
