package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

func main() {
	disc, err := xbox.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	type ent struct{ path string; size uint32 }
	var courses []ent
	disc.Walk(func(e xbox.Entry) error {
		if !e.IsDir && strings.HasPrefix(e.Path, "/Stage/") && strings.Contains(e.Path, "/cs_CS_") && strings.HasSuffix(e.Path, "_pmt.sz") && !strings.Contains(e.Path, "old") {
			courses = append(courses, ent{e.Path, e.Size})
		}
		return nil
	})
	sort.Slice(courses, func(i, j int) bool { return courses[i].path < courses[j].path })
	var total uint64
	for _, c := range courses {
		fmt.Printf("%-50s %8.1f KB\n", c.path, float64(c.size)/1024)
		total += uint64(c.size)
	}
	fmt.Printf("TOTAL %d courses, %.1f MB compressed\n", len(courses), float64(total)/1024/1024)
}
