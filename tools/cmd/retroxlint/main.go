// retroxlint validates a Retro-X tree — a whole publication root (index.json)
// or a single game directory (manifest.json). It is the reference validator
// for RETROX.md §10 and works on any conforming tree, not just this repo's.
//
// Usage:
//
//	retroxlint site/public            # validate the whole publication
//	retroxlint site/public/sonic-gg   # validate one game
//	retroxlint -q PATH                # errors only
//	retroxlint -rev HEAD site/public  # validate the tree AS GIT HAS IT
//
// -rev is the one that catches a publication whose documents reference files
// nobody committed. Without it the validator checks the working tree, where the
// exporter's output is sitting on disk whether or not it was ever added; with
// it, what a host would serve is what gets checked. `git commit -- <path>`
// commits changes to tracked files and never adds new ones, so an export that
// creates new assets can pass every local check and still deploy broken.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"retroreverse.com/tools/lib/retrox/schema"
)

func main() {
	quiet := flag.Bool("q", false, "report errors only (suppress warnings)")
	rev := flag.String("rev", "", "validate PATH as it exists at this git revision, not on disk")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: retroxlint [-q] [-rev REV] PATH")
		os.Exit(2)
	}
	root := flag.Arg(0)
	if *rev != "" {
		tmp, err := checkout(*rev, root)
		if err != nil {
			fmt.Fprintln(os.Stderr, "retroxlint -rev:", err)
			os.Exit(2)
		}
		defer os.RemoveAll(tmp)
		root = filepath.Join(tmp, root)
	}
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

// checkout extracts `path` at a revision into a temp dir, so the validator sees
// the committed tree rather than the working one.
func checkout(rev, path string) (string, error) {
	tmp, err := os.MkdirTemp("", "retroxlint-")
	if err != nil {
		return "", err
	}
	ar := exec.Command("git", "archive", rev, path)
	pipe, err := ar.StdoutPipe()
	if err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	tar := exec.Command("tar", "-x", "-C", tmp)
	tar.Stdin = pipe
	tar.Stderr = os.Stderr
	ar.Stderr = os.Stderr
	if err := ar.Start(); err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	if err := tar.Run(); err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	if err := ar.Wait(); err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("git archive %s %s: %w", rev, path, err)
	}
	return tmp, nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
