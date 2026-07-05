// actortrace reports the traced actor->model bindings (see sm64ds/actors.go).
package main

import (
	"fmt"
	"sort"

	"supermario64ds/extract/sm64ds"
)

func main() {
	ls, err := sm64ds.OpenLevels("../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "../extracted")
	if err != nil { sm64ds.Die(err) }
	used := map[int]int{}
	for i := 0; i < sm64ds.NumLevels; i++ {
		lv, err := ls.Level(i)
		if err != nil { continue }
		for _, o := range lv.Objects { used[o.Actor]++ }
	}
	am, err := ls.TraceActorModels(9)
	if err != nil { sm64ds.Die(err) }
	var ids []int
	for id := range am { if used[id] > 0 { ids = append(ids, id) } }
	sort.Slice(ids, func(i, j int) bool { return used[ids[i]] > used[ids[j]] })
	nb := 0
	for _, id := range ids {
		nb += used[id]
		fmt.Printf("actor %3d ×%-4d -> %v\n", id, used[id], am[id])
	}
	total := 0
	for _, n := range used { total += n }
	fmt.Printf("-- %d actors bound, covering %d/%d placements\n", len(ids), nb, total)
}
