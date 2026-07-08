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
	stall := flag.Int("stall", 1, "deadlock-guard tolerance multiplier")
	vramOut := flag.String("vramout", "", "write the raw 1MB VRAM to this file after the run")
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

	type snap struct{ pc, sp, r11, lr uint32 }
	var ring [64]snap
	var ri int
	dumped := false
	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		n++
		c := mm.CPU
		ring[ri%len(ring)] = snap{pc, c.Reg(13), c.Reg(11), c.Reg(14)}
		ri++
		if !dumped && pc >= 0xB0000 && pc < 0xB1000 {
			dumped = true
			fmt.Printf("[%9d] EXEC IN STACK at %X — last %d pcs:\n", n, pc, len(ring))
			for i := 0; i < len(ring); i++ {
				s := ring[(ri+i)%len(ring)]
				fmt.Printf("  sp=%08X r11=%08X lr=%08X  %s\n", s.sp, s.r11, s.lr, mm.DisasmAt(s.pc))
			}
		}
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
		case 0x9E0EC:
			ev("readysig", 10, "SendSignal(task=%X, sig=%X)", c.Reg(0), c.Reg(1))
		case 0x9E270:
			ev("wsigmask", 10, "WaitSignal(mask=%X)", c.Reg(0))
		case 0x9E274:
			ev("wsigret", 10, "-> r0=%X", c.Reg(0))
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

	m.StallTolerance = *stall
	m.NoStreams = true
	m.PadScript = []threedo.PadStep{{AtStep: 10000000, Buttons: threedo.PadStart}, {AtStep: 10400000, Buttons: 0}}

	res := m.Run(*steps)
	fmt.Printf("stopped: %s after %d steps pc=%X\n", res.Reason, res.Steps, res.PC)
	fmt.Println("\nevent totals:")
	for k, v := range counts {
		fmt.Printf("  %-14s %d\n", k, v)
	}

	// Kernel/folio call census: which vectors dominate, and the tail of the log.
	type ofr struct{ off, from uint32 }
	byOff := map[ofr]int{}
	for _, k := range m.KernelCalls {
		byOff[ofr{k.Offset, k.From}]++
	}
	fmt.Printf("\nkernel/folio calls by offset+site (%d total):\n", len(m.KernelCalls))
	for of, cnt := range byOff {
		if cnt > 20 {
			fmt.Printf("  -0x%-6X from 0x%08X x%d\n", of.off, of.from, cnt)
		}
	}
	fmt.Println("\nlast 30 kernel/folio calls:")
	kc := m.KernelCalls
	for _, k := range kc[max(0, len(kc)-30):] {
		fmt.Printf("  folio[-0x%X] from 0x%08X args=%08X %08X %08X %08X\n", k.Offset, k.From, k.Args[0], k.Args[1], k.Args[2], k.Args[3])
	}
	bySWI := map[uint32]int{}
	for _, k := range m.SWICalls {
		bySWI[k.Offset]++
	}
	fmt.Printf("\nSWIs by number (%d total):\n", len(m.SWICalls))
	for n, cnt := range bySWI {
		fmt.Printf("  0x%-6X x%d\n", n, cnt)
	}
	msgSites := map[uint32]map[uint32]int{}
	for _, k := range m.SWICalls {
		if k.Offset == 0x10013 || k.Offset == 0x10012 || k.Offset == 0x10001 {
			if msgSites[k.Offset] == nil {
				msgSites[k.Offset] = map[uint32]int{}
			}
			msgSites[k.Offset][k.From]++
		}
	}
	for _, swi := range []uint32{0x10013, 0x10012, 0x10001} {
		fmt.Printf("SWI 0x%X sites:\n", swi)
		for from, cnt := range msgSites[swi] {
			if cnt > 2 {
				fmt.Printf("  from 0x%08X x%d\n", from, cnt)
			}
		}
	}
	fmt.Println("all SWIs except WaitVBL SendIO (chronological, first 150):")
	shown := 0
	for _, k := range m.SWICalls {
		if k.Offset == 0x10018 && k.Args[2] == 0x20 {
			continue // the frame loop's field-wait spam
		}
		fmt.Printf("  SWI 0x%-6X from 0x%08X args=%08X %08X %08X %08X\n", k.Offset, k.From, k.Args[0], k.Args[1], k.Args[2], k.Args[3])
		if shown++; shown > 150 {
			break
		}
	}
	fmt.Printf("\nVRAM non-zero (640KB): %d\n", m.VRAMNonZero(640*1024))
	for _, s := range m.TaskSummary() {
		fmt.Println(" ", s)
	}
	fmt.Println("\nitems:")
	for _, s := range m.ItemsSummary() {
		fmt.Println(" ", s)
	}
	fmt.Println("\nVRAM occupancy (16KB bins, nonzero bytes):")
	for base := uint32(0x200000); base < 0x300000; base += 0x4000 {
		nz := 0
		for a := base; a < base+0x4000; a++ {
			if m.Read(a) != 0 {
				nz++
			}
		}
		if nz > 0 {
			fmt.Printf("  %08X: %d\n", base, nz)
		}
	}
	if *vramOut != "" {
		buf := make([]byte, 0x100000)
		for i := range buf {
			buf[i] = m.Read(0x200000 + uint32(i))
		}
		if err := os.WriteFile(*vramOut, buf, 0644); err != nil {
			die(err)
		}
		fmt.Println("wrote VRAM to", *vramOut)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "memtrace:", err)
	os.Exit(1)
}
