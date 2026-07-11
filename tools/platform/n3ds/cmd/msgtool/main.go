// msgtool decodes a Nintendo 3DS message archive (SZS → NARC → MSBT) from a
// title's RomFS and prints every label→text pair. It is how we read what a game's
// dialogs, menus and prompts actually say — for Super Mario 3D Land it revealed
// that the boot "NULL" dialog is the StreetPass first-launch welcome.
//
// Usage:
//
//	msgtool game.cci /LocalizedData/EuEnglish/MessageData/SystemMessage.szs
//	msgtool -grep StreetPass game.cci <path>
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	grep := flag.String("grep", "", "only print messages whose label or text contains this substring")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: msgtool [-grep S] game.cci <romfs-path.szs> [more.szs...]")
		os.Exit(2)
	}
	img, err := os.ReadFile(flag.Arg(0))
	must(err)
	ncsd, err := n3ds.ParseNCSD(img)
	must(err)
	c, err := ncsd.Partition(0)
	must(err)
	rfs, err := c.RomFS()
	must(err)

	for _, path := range flag.Args()[1:] {
		raw, err := rfs.File(path)
		if err != nil {
			fmt.Printf("%s: %v\n", path, err)
			continue
		}
		data := raw
		if len(raw) >= 4 && string(raw[:4]) == "Yaz0" {
			data, err = n3ds.Yaz0(raw)
			must(err)
		}
		files, err := n3ds.NARCFiles(data)
		must(err)
		fmt.Printf("=== %s: %d files ===\n", path, len(files))
		for fi, fd := range files {
			msgs, err := n3ds.ParseMSBT(fd)
			if err != nil {
				continue
			}
			for _, m := range msgs {
				text := strings.ReplaceAll(m.Text, "\n", "\\n")
				text = strings.ReplaceAll(text, "￿", "{$}") // control-code placeholder
				if *grep != "" && !strings.Contains(m.Label, *grep) && !strings.Contains(text, *grep) {
					continue
				}
				fmt.Printf("  file%02d[%d] %-28s %s\n", fi, m.Index, m.Label, text)
			}
		}
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "msgtool:", err)
		os.Exit(1)
	}
}
