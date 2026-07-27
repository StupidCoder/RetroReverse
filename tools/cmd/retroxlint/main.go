// retroxlint validates a Retro-X tree — a whole publication root (index.json)
// or a single game directory (manifest.json). It is the reference validator
// for RETROX.md §10 and works on any conforming tree, not just this repo's.
//
// Usage:
//
//	retroxlint site/public            # validate the whole publication
//	retroxlint site/public/sonic-gg   # validate one game
//	retroxlint -q PATH                # errors only
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/lib/retrox/schema"
)

func main() {
	quiet := flag.Bool("q", false, "report errors only (suppress warnings)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: retroxlint [-q] PATH")
		os.Exit(2)
	}
	root := flag.Arg(0)
	fsys := os.DirFS(root)

	var issues []schema.Issue
	switch {
	case exists(root + "/index.json"):
		issues = schema.ValidateTree(fsys)
	case exists(root + "/manifest.json"):
		issues = schema.ValidateGame(fsys, "")
	default:
		fmt.Fprintf(os.Stderr, "%s holds neither index.json nor manifest.json\n", root)
		os.Exit(2)
	}

	errs, warns := 0, 0
	for _, iss := range issues {
		if iss.Level == "warn" {
			warns++
			if *quiet {
				continue
			}
		} else {
			errs++
		}
		fmt.Println(iss.String())
	}
	if errs > 0 {
		fmt.Fprintf(os.Stderr, "%d error(s), %d warning(s)\n", errs, warns)
		os.Exit(1)
	}
	if warns > 0 && !*quiet {
		fmt.Fprintf(os.Stderr, "clean: 0 errors, %d warning(s)\n", warns)
	} else {
		fmt.Fprintln(os.Stderr, "clean")
	}
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
