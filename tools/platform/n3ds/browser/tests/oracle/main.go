package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"retroreverse.com/tools/platform/n3ds"
	"strconv"
)

func hash(b []byte) uint32 { h := fnv.New32a(); h.Write(b); return h.Sum32() }
func main() {
	b, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	m, e := n3ds.NewMachine(b)
	if e != nil {
		panic(e)
	}
	m.SingleThreaded = true
	n := 60
	if len(os.Args) > 2 {
		n, _ = strconv.Atoi(os.Args[2])
	}
	for i := 0; i <= n; i++ {
		if i > 0 {
			m.RunFrames(1, 100000000)
		}
		a, b := m.Framebuffer("top"), m.Framebuffer("bottom")
		if i == 58 && a != nil {
			os.WriteFile("/private/tmp/3ds-go58.rgba", a.Pix, 0600)
		}
		var ha, hb uint32
		if a != nil {
			ha = hash(a.Pix)
		}
		if b != nil {
			hb = hash(b.Pix)
		}
		v := map[string]any{"frame": i, "vblanks": m.VBlanks(), "instructions": m.Instrs(), "pc": m.CPU.R, "cpsr": m.CPU.CPSR(), "commands": m.PICACommands(), "top": ha, "bottom": hb, "halt": m.HaltReason()}
		o, _ := json.Marshal(v)
		fmt.Println(string(o))
		if m.CPU.Halted {
			break
		}
	}
}
