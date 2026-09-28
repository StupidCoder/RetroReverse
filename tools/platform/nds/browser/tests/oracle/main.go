package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/dsmachine"
	"strconv"
)

func hash(b []byte) uint32 { h := fnv.New32a(); h.Write(b); return h.Sum32() }
func main() {
	b, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	rom, e := nds.Open(b)
	if e != nil {
		panic(e)
	}
	m := dsmachine.New(rom, 0)
	n := uint64(120)
	if len(os.Args) > 2 {
		n, _ = strconv.ParseUint(os.Args[2], 10, 64)
	}
	for i := uint64(0); i <= n; i++ {
		if i > 0 {
			if len(os.Args) > 3 {
				if i == 320 || i == 420 {
					m.SetTouch(128, 120, true)
				}
				if i == 340 || i == 440 {
					m.SetTouch(0, 0, false)
				}
			}
			r := m.RunFrames(1, 10000000, 64)
			if r.Frames < i {
				panic(r.Reason)
			}
		}
		if true {
			a, b := m.Screens()
			v := map[string]any{"frame": m.Frame(), "steps": m.Steps, "arm9": m.Regs(true), "arm7": m.Regs(false), "ram": hash(m.Snapshot(true, 0x2000000, 0x400000)), "top": hash(a.Pix), "bottom": hash(b.Pix)}
			o, _ := json.Marshal(v)
			fmt.Println(string(o))
		}
	}
}
