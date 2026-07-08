// ndsinfo prints the header, integrity checks and filesystem catalog of a Nintendo
// DS cartridge image, built on tools/nds. It is the DS counterpart of amiga/cmd/
// adfdump: a container inspector.
//
//	ndsinfo [-files] [-tree] [-grep SUBSTR] rom.nds
//
// -files lists every file with its ID, byte range and size; -tree groups by
// directory; -grep filters the listing to paths containing SUBSTR.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/nds"
)

func main() {
	files := flag.Bool("files", false, "list every file (ID, range, size, path)")
	tree := flag.Bool("tree", false, "list directories with per-directory file counts")
	grep := flag.String("grep", "", "only show files whose path contains this substring")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: ndsinfo [-files] [-tree] [-grep S] rom.nds")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ndsinfo:", err)
		os.Exit(1)
	}
	rom, err := nds.Open(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ndsinfo:", err)
		os.Exit(1)
	}
	h := rom.Header

	fmt.Printf("Title        %q\n", h.Title)
	fmt.Printf("Game code    %s   Maker %s   Unit %d   Version %d\n", h.GameCode, h.MakerCode, h.UnitCode, h.ROMVersion)
	fmt.Printf("Device cap   0x%02X  → %d MB chip   (image %d bytes, used 0x%X)\n",
		h.DeviceCap, h.ChipBytes()/(1024*1024), len(data), h.TotalUsedSize)
	comp, ok := rom.VerifyHeaderCRC()
	fmt.Printf("Header CRC   stored 0x%04X computed 0x%04X %s\n", h.HeaderCRC, comp, check(ok))
	fmt.Printf("Logo CRC     stored 0x%04X\n", h.LogoCRC)
	fmt.Println()
	fmt.Printf("ARM9  ROM 0x%08X  size 0x%X  → RAM 0x%08X  entry 0x%08X\n", h.ARM9ROMOff, h.ARM9Size, h.ARM9RAMAddr, h.ARM9Entry)
	fmt.Printf("ARM7  ROM 0x%08X  size 0x%X  → RAM 0x%08X  entry 0x%08X\n", h.ARM7ROMOff, h.ARM7Size, h.ARM7RAMAddr, h.ARM7Entry)
	fmt.Printf("ovl9  ROM 0x%08X  size 0x%X  (%d overlays)\n", h.ARM9OverlayOff, h.ARM9OverlaySize, h.ARM9OverlaySize/32)
	fmt.Printf("ovl7  ROM 0x%08X  size 0x%X  (%d overlays)\n", h.ARM7OverlayOff, h.ARM7OverlaySize, h.ARM7OverlaySize/32)
	fmt.Printf("FNT   0x%08X  size 0x%X\n", h.FNTOff, h.FNTSize)
	fmt.Printf("FAT   0x%08X  size 0x%X  → %d files\n", h.FATOff, h.FATSize, h.FATSize/8)
	fmt.Printf("Icon  0x%08X\n", h.IconOff)
	fmt.Println()
	fmt.Printf("Filesystem   %d named files\n", len(rom.Files))

	if *tree {
		dirs := map[string]int{}
		for _, f := range rom.Files {
			d := "/"
			if i := strings.LastIndex(f.Path, "/"); i >= 0 {
				d = f.Path[:i]
			}
			dirs[d]++
		}
		keys := make([]string, 0, len(dirs))
		for k := range dirs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %-40s %4d files\n", k, dirs[k])
		}
	}

	if *files {
		list := append([]nds.FileInfo(nil), rom.Files...)
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
		for _, f := range list {
			if *grep != "" && !strings.Contains(f.Path, *grep) {
				continue
			}
			e := rom.FAT[f.ID]
			fmt.Printf("  %4d  0x%08X-0x%08X  %8d  %s\n", f.ID, e.Start, e.End, e.Size(), f.Path)
		}
	}
}

func check(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}
