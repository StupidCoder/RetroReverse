// knowledgecheck validates a package and optionally verifies local media identity.
package main

import (
	"flag"
	"fmt"
	"os"
	"retroreverse.com/tools/knowledge"
	"strings"
)

type mediaFlags map[string]string

func (m mediaFlags) String() string { return "role=path" }
func (m mediaFlags) Set(s string) error {
	role, path, ok := strings.Cut(s, "=")
	if !ok || role == "" || path == "" {
		return fmt.Errorf("expected role=path")
	}
	if _, ok := m[role]; ok {
		return fmt.Errorf("duplicate role %s", role)
	}
	m[role] = path
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	media := mediaFlags{}
	files := mediaFlags{}
	flag.Var(files, "file", "verify logical file-set member (repeat name=path)")
	flag.Var(media, "media", "verify raw local media (repeat role=path)")
	artifacts := flag.String("artifacts", "", "verify curated preview files beneath this site root")
	release := flag.String("release", "", "require this release when checking media")
	flag.Parse()
	if flag.NArg() != 1 {
		return fmt.Errorf("usage: knowledgecheck [-media role=path | -file name=path] [-release id] package.json")
	}
	p, e := knowledge.Read(flag.Arg(0))
	if e != nil {
		return e
	}
	if *artifacts != "" {
		if e = p.VerifyArtifacts(*artifacts); e != nil {
			return e
		}
		fmt.Println("ARTIFACTS VERIFIED")
	}
	fmt.Printf("VALID %s revision %d sha256=%s\n", p.ID, p.Revision, p.Hash)
	if len(media) > 0 && len(files) > 0 {
		return fmt.Errorf("choose raw -media or logical -file identity, not both")
	}
	if len(files) > 0 {
		members := []knowledge.Member{}
		for name, path := range files {
			n, h, e := knowledge.HashFile(path)
			if e != nil {
				return e
			}
			members = append(members, knowledge.Member{Path: name, Size: n, SHA256: h})
		}
		r, e := p.MatchFileSet(members)
		if e != nil {
			return e
		}
		if *release != "" && *release != r {
			return fmt.Errorf("matched %s, expected %s", r, *release)
		}
		fmt.Println("FILE SET VERIFIED", r)
	} else if len(media) > 0 {
		r, e := p.MatchMedia(media)
		if e != nil {
			return e
		}
		if *release != "" && *release != r {
			return fmt.Errorf("matched %s, expected %s", r, *release)
		}
		fmt.Println("MEDIA VERIFIED", r)
	} else if *release != "" {
		return fmt.Errorf("-release requires -media or -file")
	}
	return nil
}
