// gdinfo is the static inspector for a GD-ROM rip — the counterpart of
// gcinfo / xbeinfo. It opens a cdrdao-style .cue + .bin, anchors the tracks,
// and prints what the disc says about itself: the track table, the IP.BIN
// boot metadata, and the ISO 9660 census of the high-density area.
//
// Usage:
//
//	gdinfo -image game.cue [-tracks] [-ip] [-files] [-at LBA] [-x FILE[:OUT]] [-md5]
//
// With no selection flags it prints tracks, IP.BIN and the volume summary.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/dc"
)

func main() {
	image := flag.String("image", "", "disc image (.cue, cdrdao TOC syntax)")
	tracks := flag.Bool("tracks", false, "print the track table")
	ip := flag.Bool("ip", false, "print the IP.BIN metadata")
	files := flag.Bool("files", false, "list every file in the high-density area")
	at := flag.String("at", "", "name the file containing this absolute LBA (hex ok)")
	x := flag.String("x", "", "extract FILE[:OUT] from the volume")
	md5f := flag.Bool("md5", false, "print the .bin's MD5 (reads the whole image)")
	flag.Parse()
	if *image == "" {
		fmt.Fprintln(os.Stderr, "usage: gdinfo -image game.cue [-tracks] [-ip] [-files] [-at LBA] [-x FILE[:OUT]] [-md5]")
		os.Exit(2)
	}
	if err := run(*image, *tracks, *ip, *files, *at, *x, *md5f); err != nil {
		fmt.Fprintln(os.Stderr, "gdinfo:", err)
		os.Exit(1)
	}
}

func run(image string, tracks, ip, files bool, at, x string, md5f bool) error {
	d, err := dc.OpenDisc(image)
	if err != nil {
		return err
	}
	all := !tracks && !ip && !files && at == "" && x == "" && !md5f

	if tracks || all {
		fmt.Printf("tracks:\n")
		for _, t := range d.Tracks {
			lba := "-"
			if t.StartLBA >= 0 {
				lba = fmt.Sprintf("%d-%d", t.StartLBA, t.StartLBA+int(t.Length/2352)-1)
			}
			fmt.Printf("  %d  %-10s  offset %-11d  %11d bytes  lba %s\n", t.Number, t.Mode, t.FileOffset, t.Length, lba)
		}
	}
	if ip || all {
		i := d.IP
		fmt.Printf("ip.bin: %s %s (%s) — %q\n", i.ProductNo, i.Version, i.ReleaseDate, i.Title)
		fmt.Printf("  hardware %q  maker %q  device %q\n", i.HardwareID, i.MakerID, i.DeviceInfo)
		fmt.Printf("  area %q  peripherals %q  boot %q  company %q\n", i.AreaSyms, i.Peripherals, i.BootFile, i.Company)
	}
	if all {
		fmt.Printf("volume: %q, %d blocks (session-relative), data area at lba %d\n", d.Vol.Name, d.Vol.Blocks, d.DataLBA())
	}
	if files {
		return d.Vol.Walk(func(e iso9660.Entry) error {
			fmt.Println(e)
			return nil
		})
	}
	if at != "" {
		lba, err := strconv.ParseInt(strings.TrimPrefix(at, "0x"), 0, 64)
		if err != nil {
			return fmt.Errorf("bad -at %q", at)
		}
		if e, ok := d.Vol.FileAt(int(lba)); ok {
			fmt.Printf("lba %d: %s\n", lba, e)
		} else {
			fmt.Printf("lba %d: no file extent\n", lba)
		}
	}
	if x != "" {
		name, out := x, ""
		if i := strings.IndexByte(x, ':'); i >= 0 {
			name, out = x[:i], x[i+1:]
		}
		if out == "" {
			out = filepath.Base(strings.ReplaceAll(name, ";1", ""))
		}
		data, err := d.Vol.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("extracted %s -> %s (%d bytes)\n", name, out, len(data))
	}
	if md5f {
		sum, err := d.MD5()
		if err != nil {
			return err
		}
		fmt.Printf("md5: %s\n", sum)
	}
	return nil
}
