// Package cli is the standard front-end for every game's webexport command
// (STANDARDS §4.1): one flag vocabulary, stage gating via -only, uniform
// progress output on stderr, curation merge, and the final Builder.Write.
//
// A minimal exporter:
//
//	func main() {
//		cli.Main("sonic-gg", func(ctx *cli.Context) error {
//			if ctx.Stage("levels") { ... ctx.Builder.AddLevel(...) ... }
//			if ctx.Stage("music")  { ... }
//			return nil
//		})
//	}
package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/lib/retrox/build"
	"retroreverse.com/tools/lib/retrox/curation"
)

// Stages is the -only vocabulary (comma-separated; "all" = everything).
var Stages = []string{"levels", "objects", "music", "sfx", "pictures", "videos"}

type Context struct {
	In      string // -in
	Out     string // -o
	Seed    int64  // -seed
	Verbose bool   // -v

	Builder  *build.Builder
	Curation *curation.Curation

	only map[string]bool
}

// Enabled reports whether a stage is selected by -only, without announcing
// it. Use it to decide whether to emit cross-stage references (a level's
// music binding, placements' object refs) on partial runs.
func (c *Context) Enabled(name string) bool { return c.only[name] }

// Stage reports whether a stage is enabled by -only, and announces it on
// stderr when it is. Use as a gate: if ctx.Stage("music") { ... }.
func (c *Context) Stage(name string) bool {
	if !c.only[name] {
		return false
	}
	fmt.Fprintf(os.Stderr, "[%s]\n", name)
	return true
}

// Progress prints one item line within a stage.
func (c *Context) Progress(stage string, i, n int, item string) {
	fmt.Fprintf(os.Stderr, "[%s]  %d/%d  %s\n", stage, i, n, item)
}

// Logf prints when -v is set.
func (c *Context) Logf(format string, args ...any) {
	if c.Verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Main parses the standard flags, prepares the Builder and curation, runs
// extract, then writes + validates the tree. Exits non-zero on failure.
func Main(slug string, extract func(*Context) error) {
	in := flag.String("in", "", "input image / rom / game dir")
	out := flag.String("o", filepath.Join("..", "..", "site", "public", slug), "output root")
	only := flag.String("only", "all", "comma-separated stages: "+strings.Join(Stages, "|")+"|all")
	seed := flag.Int64("seed", 0, "seed for any randomized choices (reproducible output)")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	ctx := &Context{
		In:      *in,
		Out:     *out,
		Seed:    *seed,
		Verbose: *verbose,
		Builder: build.New(*out, slug),
		only:    map[string]bool{},
	}
	if *only == "all" || *only == "" {
		for _, s := range Stages {
			ctx.only[s] = true
		}
	} else {
		for _, s := range strings.Split(*only, ",") {
			s = strings.TrimSpace(s)
			ok := false
			for _, known := range Stages {
				if s == known {
					ok = true
					break
				}
			}
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown stage %q (want %s)\n", s, strings.Join(Stages, "|"))
				os.Exit(2)
			}
			ctx.only[s] = true
		}
	}

	// Curation lives in the game's repo directory, next to extract/.
	cur, err := curation.Load(filepath.Join("..", "..", "games", slug, "curation"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "curation: %v\n", err)
		os.Exit(1)
	}
	ctx.Curation = cur

	if err := extract(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := ctx.Builder.ApplyCuration(cur); err != nil {
		fmt.Fprintf(os.Stderr, "curation: %v\n", err)
		os.Exit(1)
	}
	if err := ctx.Builder.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	for _, w := range ctx.Builder.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
}
