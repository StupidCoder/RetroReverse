// memtrace is a throwaway step-stamped event tracer for the NFS 3DO memory-
// manager investigation: it logs, with instruction step numbers, every visit
// to the game's memory-manager entry points (InitMemMgr, InitMemList, the
// region chainer, the CCB-Manager buffer replace, FreeByPtr, findmemblock,
// panic) plus writes to the CCB-Manager current-buffer global [0x472D4].
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	steps := flag.Uint64("steps", 3000000, "max instructions")
	flag.Parse()

	data, err := os.ReadFile(*image)
	if err != nil {
		die(err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		die(err)
	}
	prog, err := vol.ReadFile("LaunchMe")
	if err != nil {
		die(err)
	}
	aif, err := threedo.ParseAIF(prog)
	if err != nil {
		die(err)
	}

	m := threedo.NewMachine()
	m.SetVolume(vol)
	m.SetVBLMirror(0x42734)
	m.LoadAIF(aif)

	var n uint64
	rd := func(a uint32) uint32 {
		return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
	}
	rstr := func(a uint32) string {
		var b []byte
		for i := uint32(0); i < 24; i++ {
			c := m.Read(a + i)
			if c == 0 {
				break
			}
			b = append(b, c)
		}
		return string(b)
	}
	// count-limited event logger
	counts := map[string]int{}
	ev := func(kind string, limit int, format string, args ...any) {
		counts[kind]++
		if counts[kind] <= limit {
			fmt.Printf("[%9d] %-12s %s\n", n, kind, fmt.Sprintf(format, args...))
		}
	}

	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		n++
		c := mm.CPU
		switch pc {
		case 0xEB6C:
			ev("sysinit", 10, "0xEB6C(r0=%X) lr=%X", c.Reg(0), c.Reg(14))
		case 0xECE4:
			ev("0xECE4", 10, "r0=%X r1=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(14))
		case 0x12DC4:
			ev("dispatch", 10, "MOV pc,r2  r0=%X r2=%X lr=%X", c.Reg(0), c.Reg(2), c.Reg(14))
		case 0x30F8C:
			ev("InitMemMgr", 10, "count=%X dram=%X ? vram=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(3), c.Reg(14))
		case 0x2EBF4:
			ev("InitMemList", 20, "memtype=%X start=%X end=%X align=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3), c.Reg(14))
		case 0x2ECF8:
			ev("ChainRegion", 20, "slot0=%X memtype=%X start=%X end=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3), c.Reg(14))
		case 0x3B7D8:
			ev("CCBreplace", 20, "r0=%X sz<<6=%X r2=%X lr=%X  cur[0x472D4]=%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(14), rd(0x472D4))
		case 0x2EF8C:
			ev("alloc", 40, "name=%q size=%X flags=%X lr=%X", rstr(c.Reg(0)), c.Reg(1), c.Reg(2), c.Reg(14))
		case 0x2F76C:
			ev("FreeByPtr", 12, "ptr=%X lr=%X", c.Reg(0), c.Reg(14))
		case 0x2F6E0:
			ev("findmemblock", 6, "ptr=%X lr=%X", c.Reg(0), c.Reg(14))
		case 0x357D0:
			ev("panic", 8, "fmt=%q lr=%X", rstr(c.Reg(0)), c.Reg(14))
		case 0x301B8:
			ev("nodepool", 10, "base=%X count=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(14))
		case 0x32BA4:
			ev("loadfile", 10, "name=%q r1=%X r2=%X lr=%X", rstr(c.Reg(0)), c.Reg(1), c.Reg(2), c.Reg(14))
		case 0x336F0:
			ev("filestat", 10, "name=%q lr=%X", rstr(c.Reg(0)), c.Reg(14))
		case 0x32BF8:
			ev("statsize", 10, "size=%X", c.Reg(0))
		case 0x32C1C:
			ev("loadalloc", 10, "handle=%X", c.Reg(0))
		case 0x39C0C:
			ev("seekglue", 10, "stream=%X off=%X whence=%X lr=%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(14))
		case 0x33820:
			ev("seekendret", 10, "r0=%X", c.Reg(0))
		case 0x12D2C:
			ev("ovlbuf", 10, "0x32B5C returned r0=%X", c.Reg(0))
		case 0x12D34:
			ev("ovlentry", 10, "0x3C0B4 returned r2=%X", c.Reg(0))
		}
	}
	m.WatchLo, m.WatchHi = 0x472D4, 0x472D8
	m.OnWrite = func(addr, val, pc uint32) {
		ev("G4write", 20, "[0x472D4] byte=%02X from pc=%X", val&0xFF, pc)
	}

	res := m.Run(*steps)
	fmt.Printf("stopped: %s after %d steps pc=%X\n", res.Reason, res.Steps, res.PC)
	fmt.Println("\nevent totals:")
	for k, v := range counts {
		fmt.Printf("  %-14s %d\n", k, v)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "memtrace:", err)
	os.Exit(1)
}
