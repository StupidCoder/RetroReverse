// Command stringsdump decodes DATA/STRINGS.PAK. With no -block it lists every
// block id and its string count; with -block N it prints that block's strings.
//
// Usage: stringsdump [-game ../game] [-block N] [-idx I]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ultimaunderworld/extract/strpak"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder")
	block := flag.Int("block", -1, "print this block's strings (default: list all blocks)")
	idx := flag.Int("idx", -1, "with -block, print only this string index")
	flag.Parse()

	data, err := os.ReadFile(filepath.Join(*game, "DATA/STRINGS.PAK"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "stringsdump:", err)
		os.Exit(1)
	}
	a, err := strpak.Parse(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stringsdump:", err)
		os.Exit(1)
	}

	if *block < 0 {
		total := 0
		for _, id := range a.BlockIDs() {
			ss, _ := a.Block(id)
			total += len(ss)
			fmt.Printf("block %4d: %4d strings   e.g. %q\n", id, len(ss), first(ss))
		}
		fmt.Printf("%d blocks, %d strings\n", len(a.BlockIDs()), total)
		return
	}

	if *idx >= 0 {
		s, ok := a.String(*block, *idx)
		if !ok {
			fmt.Fprintf(os.Stderr, "no block %d string %d\n", *block, *idx)
			os.Exit(1)
		}
		fmt.Print(s)
		return
	}
	ss, ok := a.Block(*block)
	if !ok {
		fmt.Fprintf(os.Stderr, "no block %d\n", *block)
		os.Exit(1)
	}
	for i, s := range ss {
		fmt.Printf("[%d] %s\n", i, strings.ReplaceAll(strings.TrimRight(s, "\n"), "\n", "\\n"))
	}
}

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return strings.TrimRight(ss[0], "\n")
}
