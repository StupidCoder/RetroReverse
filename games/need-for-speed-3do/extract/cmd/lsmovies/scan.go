package main

// -scanreels: walk every file on the disc and report N.M-shaped strings —
// the hunt for the magazine's movie tables.

import (
	"fmt"
	"regexp"
	"sort"

	"retroreverse.com/tools/platform/threedo"
)

var reelPat = regexp.MustCompile(`[0-9]{1,3}\.[0-9]`)

func scanReels(vol *threedo.Volume) {
	type hit struct {
		path  string
		reels []string
	}
	var hits []hit
	vol.Walk(func(e threedo.Entry) error {
		if e.IsDir {
			return nil
		}
		raw, err := vol.ReadFile(e.Path)
		if err != nil {
			return nil
		}
		seen := map[string]bool{}
		var reels []string
		for _, loc := range reelPat.FindAllIndex(raw, -1) {
			// require a plausible string context: preceded and followed by
			// NUL or non-alphanumeric (skip float text like "1.5x" in docs)
			s, epos := loc[0], loc[1]
			if s > 0 {
				c := raw[s-1]
				if c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '.' {
					continue
				}
			}
			if epos < len(raw) {
				c := raw[epos]
				if c != 0 && c != '.' {
					continue
				}
			}
			m := string(raw[s:epos])
			if !seen[m] {
				seen[m] = true
				reels = append(reels, m)
			}
		}
		if len(reels) > 0 {
			sort.Strings(reels)
			hits = append(hits, hit{e.Path, reels})
		}
		return nil
	})
	for _, h := range hits {
		if len(h.reels) > 12 {
			fmt.Printf("%-40s %d reels: %v ...\n", h.path, len(h.reels), h.reels[:12])
		} else {
			fmt.Printf("%-40s %d reels: %v\n", h.path, len(h.reels), h.reels)
		}
	}
}
