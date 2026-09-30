// knowledgeasset resolves a pinned static asset and optionally exports its bytes/PNG.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"retroreverse.com/tools/knowledge"
	"strings"
)

type roles map[string]string

func (m roles) String() string { return "role=path" }
func (m roles) Set(s string) error {
	k, v, ok := strings.Cut(s, "=")
	if !ok || k == "" || v == "" || m[k] != "" {
		return fmt.Errorf("expected unique role=path")
	}
	m[k] = v
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	media := roles{}
	flag.Var(media, "media", "local raw media (repeat role=path)")
	release := flag.String("release", "", "exact release ID")
	asset := flag.String("asset", "", "asset ID")
	out := flag.String("out", "", "optional exported asset path (must not exist)")
	report := flag.String("report", "", "optional JSON provenance path (must not exist)")
	list := flag.Bool("decoders", false, "list registered decoder versions")
	flag.Parse()
	if *list {
		return json.NewEncoder(os.Stdout).Encode(knowledge.Decoders())
	}
	if flag.NArg() != 1 || *release == "" || *asset == "" {
		return fmt.Errorf("usage: knowledgeasset -release id -asset id -media role=path [-out path] [-report path] package.json")
	}
	p, e := knowledge.Read(flag.Arg(0))
	if e != nil {
		return e
	}
	r, e := knowledge.OpenResolver(p, *release, media)
	if e != nil {
		return e
	}
	defer r.Close()
	result, e := r.ResolveAsset(*asset)
	if e != nil {
		return e
	}
	b, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	if *out != "" {
		if e = writeNew(*out, result.Data); e != nil {
			return e
		}
	}
	if *report != "" {
		if e = writeNew(*report, b); e != nil {
			return e
		}
	}
	_, e = os.Stdout.Write(b)
	return e
}
func writeNew(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	c := f.Close()
	if e != nil {
		os.Remove(path)
		return e
	}
	return c
}
