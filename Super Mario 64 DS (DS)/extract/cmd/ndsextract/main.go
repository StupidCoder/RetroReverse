// ndsextract pulls the two CPU boot binaries (arm9.bin, arm7.bin) and the ARM9 code
// overlays out of the Super Mario 64 DS cartridge, writing them to a directory so the
// ARM disassembler/tracer can work on flat images. Built on the shared tools/nds
// container reader.
//
//	ndsextract [-o DIR] [-fs] rom.nds
//
// -o is the output directory (default ../extracted). -fs also dumps every filesystem
// file under DIR/files/<path>. Extracted output is regenerable and git-ignored.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/nds"
)

func main() {
	out := flag.String("o", "../extracted", "output directory")
	fs := flag.Bool("fs", false, "also extract every filesystem file under DIR/files/")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: ndsextract [-o DIR] [-fs] rom.nds")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(data)
	if err != nil {
		die(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die(err)
	}

	write := func(name string, b []byte) {
		p := filepath.Join(*out, name)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			die(err)
		}
		fmt.Printf("  %-24s %8d bytes\n", name, len(b))
	}

	fmt.Println("boot binaries:")
	arm9 := rom.ARM9()
	write("arm9.bin", arm9)
	if nds.IsBLZ(arm9) {
		dec := nds.DecompressBLZ(arm9)
		write("arm9_dec.bin", dec)
	}
	arm7 := rom.ARM7()
	write("arm7.bin", arm7)
	if nds.IsBLZ(arm7) {
		write("arm7_dec.bin", nds.DecompressBLZ(arm7))
	}

	ovls := rom.ARM9Overlays()
	if len(ovls) > 0 {
		fmt.Printf("ARM9 overlays (%d):\n", len(ovls))
		for _, o := range ovls {
			raw := rom.File(int(o.FileID))
			write(fmt.Sprintf("ovl9_%03d.bin", o.ID), raw)
			if o.Compressed && nds.IsBLZ(raw) {
				write(fmt.Sprintf("ovl9_%03d_dec.bin", o.ID), nds.DecompressBLZ(raw))
			}
			fmt.Printf("      → RAM 0x%08X  ramsize 0x%X  bss 0x%X  fileID %d  compressed=%v\n",
				o.RAMAddr, o.RAMSize, o.BSSSize, o.FileID, o.Compressed)
		}
	}

	if *fs {
		fmt.Printf("filesystem (%d files):\n", len(rom.Files))
		for _, f := range rom.Files {
			p := filepath.Join(*out, "files", filepath.FromSlash(f.Path))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				die(err)
			}
			if err := os.WriteFile(p, rom.File(f.ID), 0o644); err != nil {
				die(err)
			}
		}
		fmt.Printf("  → %s/files/\n", *out)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "ndsextract:", err)
	os.Exit(1)
}
