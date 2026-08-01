// gbarom prints the cartridge header, entry point and detected save type of a
// Game Boy Advance image.
//
//	gbarom file.gba
package main

import (
	"fmt"
	"os"

	"retroreverse.com/tools/platform/gba"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gbarom file.gba")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "gbarom:", err)
		os.Exit(2)
	}
	rom, err := gba.Parse(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gbarom:", err)
		os.Exit(2)
	}
	h := rom.Header
	fmt.Printf("image:      %d bytes (%d MiB)\n", len(data), len(data)>>20)
	fmt.Printf("title:      %q\n", h.Title)
	fmt.Printf("game code:  %s   maker: %s   version: %d\n", h.GameCode, h.MakerCode, h.Version)
	fmt.Printf("unit code:  %#02x  device: %#02x  fixed byte: %#02x\n", h.UnitCode, h.DeviceType, h.Fixed96)
	ok := "OK"
	if !h.ChecksumOK {
		ok = fmt.Sprintf("BAD (computed %#02x)", gba.ComplementCheck(data))
	}
	fmt.Printf("checksum:   %#02x %s\n", h.Checksum, ok)
	fmt.Printf("logo md5:   %s\n", h.LogoMD5)
	if h.EntryAddr != 0 {
		fmt.Printf("entry:      %08X  (word %08X: B)\n", h.EntryAddr, h.EntryInstr)
	} else {
		fmt.Printf("entry:      word %08X is not an ARM B\n", h.EntryInstr)
	}
	if id, desc := rom.SaveType(); id != "" {
		fmt.Printf("save:       %s — %s\n", id, desc)
	} else {
		fmt.Printf("save:       no driver ID string found\n")
	}
}
