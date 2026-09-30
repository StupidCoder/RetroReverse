// knowledgeexport compiles validated source packages into deterministic browser data.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"retroreverse.com/tools/knowledge"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	out := flag.String("out", "", "output JS path")
	index := flag.String("index", "", "optional static HTML asset catalog")
	check := flag.Bool("check", false, "fail if output is missing or stale")
	flag.Parse()
	if *out == "" || flag.NArg() == 0 {
		return fmt.Errorf("usage: knowledgeexport -out file.js [-check] package.json ...")
	}
	ps := []*knowledge.Package{}
	for _, path := range flag.Args() {
		p, e := knowledge.Read(path)
		if e != nil {
			return fmt.Errorf("%s: %w", path, e)
		}
		ps = append(ps, p)
	}
	b, e := knowledge.ExportJS(ps)
	if e != nil {
		return e
	}
	outputs := map[string][]byte{*out: b}
	if *index != "" {
		catalog, e := knowledge.ExportCatalog(ps)
		if e != nil {
			return e
		}
		outputs[*index] = catalog
	}
	for path, data := range outputs {
		if *check {
			old, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			if !bytes.Equal(old, data) {
				return fmt.Errorf("stale export: regenerate %s", path)
			}
			fmt.Println("CURRENT", path)
		} else {
			if e = os.WriteFile(path, data, 0644); e != nil {
				return e
			}
			fmt.Println("EXPORTED", path)
		}
	}
	return nil
}
