// Command uwinfo is the Part I reconnaissance tool for Ultima Underworld:
// it decodes the UW.EXE MZ header and load layout, and inventories the
// on-disk game data folders (DATA / CRIT / CUTS / SOUND) by extension.
//
// Usage:
//
//	go run ./cmd/uwinfo -game "../game"
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/dos"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder containing UW.EXE")
	flag.Parse()

	exePath := filepath.Join(*game, "UW.EXE")
	data, err := os.ReadFile(exePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uwinfo:", err)
		os.Exit(1)
	}

	fmt.Printf("== UW.EXE ==\n")
	fmt.Printf("path   %s\n", exePath)
	fmt.Printf("size   %d bytes\n", len(data))
	fmt.Printf("md5    %x\n\n", md5.Sum(data))

	m, err := dos.ParseMZ(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uwinfo:", err)
		os.Exit(1)
	}

	fmt.Printf("== MZ header ==\n")
	fmt.Printf("pages            %d  (last page %d bytes)\n", m.Pages, m.LastPageBytes)
	fmt.Printf("header           %d paragraphs = %d bytes\n", m.HeaderParas, m.LoadModuleOffset)
	fmt.Printf("load module      offset %#x .. %#x  (%d bytes DOS loads)\n",
		m.LoadModuleOffset, m.LoadImageEnd, m.LoadModuleSize)
	fmt.Printf("appended data    %d bytes past the load image (swapped overlays?)\n", m.AppendedSize)
	fmt.Printf("entry  CS:IP     %04X:%04X  (module offset %#x)\n", m.InitCS, m.InitIP, m.EntryLinear())
	fmt.Printf("stack  SS:SP     %04X:%04X\n", m.InitSS, m.InitSP)
	fmt.Printf("alloc  min/max   %#x / %#x paragraphs\n", m.MinAlloc, m.MaxAlloc)
	fmt.Printf("relocations      %d  (table at %#x)\n", m.Relocations, m.RelocOffset)
	if len(m.Relocs) > 0 {
		fmt.Printf("first relocs     ")
		for i := 0; i < 6 && i < len(m.Relocs); i++ {
			fmt.Printf("%04X:%04X ", m.Relocs[i].Segment, m.Relocs[i].Offset)
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Printf("== game data inventory (%s) ==\n", *game)
	for _, sub := range []string{"DATA", "CRIT", "CUTS", "SOUND", "."} {
		inventory(filepath.Join(*game, sub), sub)
	}
}

// inventory prints a per-extension file count and total size for one folder.
func inventory(dir, label string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type agg struct {
		count int
		bytes int64
	}
	byExt := map[string]*agg{}
	var files int
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		ext := strings.ToUpper(filepath.Ext(e.Name()))
		if ext == "" {
			ext = "(none)"
		}
		a := byExt[ext]
		if a == nil {
			a = &agg{}
			byExt[ext] = a
		}
		a.count++
		a.bytes += info.Size()
		files++
	}
	if files == 0 {
		return
	}
	exts := make([]string, 0, len(byExt))
	for k := range byExt {
		exts = append(exts, k)
	}
	sort.Strings(exts)
	fmt.Printf("%-6s %d files\n", label+":", files)
	for _, ext := range exts {
		a := byExt[ext]
		fmt.Printf("    %-8s %3d files  %8d bytes\n", ext, a.count, a.bytes)
	}
}
