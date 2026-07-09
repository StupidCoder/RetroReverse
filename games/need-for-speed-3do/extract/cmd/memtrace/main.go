// memtrace is a throwaway step-stamped event tracer. Current focus: the
// no-input pre-start hold. With zero input after race entry nothing ticks (not
// even the race clock); the first pad edge starts everything. On hardware the
// opponent launches by itself after ~3-4s. This trace compares a no-input run
// against a driving run: who kicks the sim chain, what state flips on the
// first input edge, and which stubbed service should have flipped it alone.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	steps := flag.Uint64("steps", 42000000, "max instructions")
	stall := flag.Int("stall", 400, "deadlock-guard tolerance multiplier")
	pad := flag.String("pad", "4000000:start,4300000:0", "control-pad script (bootoracle syntax)")
	snap := flag.String("snap", "", "comma-separated steps at which to dump 3MB DRAM+VRAM snapshots")
	snapPrefix := flag.String("snapprefix", "mem", "snapshot file prefix")
	pokeAddr := flag.Uint64("poke", 0, "write 1 to this address at step 6M")
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
	m.StallTolerance = *stall
	m.NoStreams = true
	script, err := parsePadScript(*pad)
	if err != nil {
		die(err)
	}
	m.PadScript = script

	var snapAt []uint64
	if *snap != "" {
		for _, s := range strings.Split(*snap, ",") {
			v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
			if err != nil {
				die(err)
			}
			snapAt = append(snapAt, v)
		}
		sort.Slice(snapAt, func(i, j int) bool { return snapAt[i] < snapAt[j] })
	}

	var n uint64
	counts := map[string]int{}
	ev := func(kind string, limit int, format string, args ...any) {
		key := fmt.Sprintf("%s@%d", kind, n/10_000_000)
		counts[key]++
		if counts[key] <= limit {
			fmt.Printf("[%9d t#%-4d] %-10s %s\n", n, m.CurrentTaskNum(), kind, fmt.Sprintf(format, args...))
		}
	}

	rd := func(a uint32) uint32 {
		return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
	}

	// audio SWI census: swi -> count (all 0x40000..0x400FF)
	audioSWI := map[uint32]int{}
	sendSig := map[[2]uint32]int{} // {task,sigs} -> count

	// APCS frame-pointer backtrace (EA functions use full STMDB {..,fp,ip,lr,pc}
	// frames: [r11]=saved pc, [r11-4]=lr, [r11-0xC]=caller's r11).
	backtrace := func() string {
		fp := m.CPU.Reg(11)
		s := ""
		for i := 0; i < 8 && fp > 0x1000 && fp < 0x400000; i++ {
			s += fmt.Sprintf(" <-0x%X", rd(fp-4)-4)
			fp = rd(fp - 0xC)
		}
		return s
	}

	// Operamath vector-transform census: source arrays of MulManyVec3Mat33
	// (SWI 0x50002) enumerate every world-space vertex array the game
	// transforms per frame (road strip, scenery slice rows, cars).
	mmv := map[[2]uint32]int{}
	m.OnSWI = func(mm *threedo.Machine, from, swi uint32) {
		c := mm.CPU
		if swi == 0x50002 && n > 16_000_000 {
			k := [2]uint32{from, c.Reg(3)}
			mmv[k]++
			if mmv[k] <= 2 {
				src := c.Reg(1)
				v0 := []uint32{rd(src), rd(src + 4), rd(src + 8)}
				fmt.Printf("[%9d t#%-4d] MMV        from=0x%X dst=0x%X src=0x%X mat=0x%X n=%d v0=%08X,%08X,%08X bt:%s\n",
					n, m.CurrentTaskNum(), from, c.Reg(0), src, c.Reg(2), c.Reg(3), v0[0], v0[1], v0[2], backtrace())
			}
		}
		if swi >= 0x40000 && swi < 0x40100 {
			audioSWI[swi]++
			if audioSWI[swi] <= 4 {
				ev("AUDSWI", 80, "0x%X r0=0x%X r1=0x%X r2=0x%X from=0x%X", swi, c.Reg(0), c.Reg(1), c.Reg(2), from)
			}
		}
		switch swi {
		case 0x10002: // SendSignal(task, sigs)
			k := [2]uint32{c.Reg(0), c.Reg(1)}
			sendSig[k]++
			if sendSig[k] <= 3 {
				ev("SendSig", 40, "task=%d sigs=0x%X from=0x%X", c.Reg(0), c.Reg(1), from)
			}
		}
	}

	// one-shot PC traces of the opponent's car-sim invocation (0x1779C, r0=1)
	// at two moments (pre-GO ~15M, post-GO ~17.6M): diff the executed paths to
	// find the branch the GO flag controls.
	traceAts := []uint64{}
	var traceRet uint32
	var traceLog []uint32
	var traceName string

	var integLogs int

	// write-watch: where do the road-strip vertex rows (strip struct 0x14B290,
	// verts at +0x24) come from? Log writer PC + backtrace.
	wrPCs := map[uint32]int{}
	m.WatchLo, m.WatchHi = 0x14B2B4, 0x14B2C0
	m.OnWrite = func(addr, val, pc uint32) {
		wrPCs[pc]++
		if wrPCs[pc] <= 2 {
			fmt.Printf("[%9d t#%-4d] WR         [0x%X]=0x%02X pc=0x%X%s\n", n, m.CurrentTaskNum(), addr, val, pc, backtrace())
		}
	}

	// counters for the frame chain
	var loopHead, worldUpd uint64
	carSim := map[uint32]uint64{} // r0 (car ptr) -> count
	lastStatus := uint64(0)

	// track-stream geometry trace: the road-strip vertex rows are read straight
	// off the trk stream (ReadDiskStream into strip+0x24). Log seek/read on the
	// stream plus the chunk-index helpers 0x15704/0x15724.
	var trkStream uint32
	var resolver [][3]uint32 // {fn, arg, lr} stack for arena-resolver returns
	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		n++
		c := mm.CPU
		switch pc {
		case 0x15650: // trkOpen(name, bufsz) — caches stream ptr at [0x3F2B0]
			ev("TRKOPEN", 8, "name=%q bufsz=0x%X", readCStr(m, c.Reg(0)), c.Reg(1))
		case 0x15660: // return from OpenDiskStream inside trkOpen
			trkStream = c.Reg(0)
			ev("TRKSTRM", 8, "stream=0x%X", trkStream)
		case 0x39BD0: // ReadDiskStream(stream, buf, len)
			if c.Reg(0) == trkStream && trkStream != 0 {
				ev("TRKREAD", 60, "buf=0x%X len=0x%X bt:%s", c.Reg(1), c.Reg(2), backtrace())
			}
			if c.Reg(1) <= 0x293E20 && c.Reg(1)+c.Reg(2) > 0x293E20 {
				ev("TEXREAD", 20, "buf=0x%X len=0x%X bt:%s", c.Reg(1), c.Reg(2), backtrace())
			}
		case 0x39C0C: // SeekDiskStream(stream, pos, whence)
			if c.Reg(0) == trkStream && trkStream != 0 {
				ev("TRKSEEK", 60, "pos=0x%X whence=%d bt:%s", c.Reg(1), c.Reg(2), backtrace())
			}
		case 0x15704: // seek to chunk idx: offset = [blockBase+0x98C+idx*4]
			ev("CHUNKSEEK", 60, "idx=%d lr=0x%X", c.Reg(0), c.Reg(14))
		case 0x15724: // chunk walker: r0 = idx
			ev("CHUNKWALK", 60, "idx=%d lr=0x%X", c.Reg(0), c.Reg(14))
		case 0x15CA0, 0x15D1C: // arena resolvers: capture (pc, arg) -> return
			if n > 4_500_000 {
				resolver = append(resolver, [3]uint32{pc, c.Reg(0), c.Reg(14)})
			}
		case 0x21CC4: // drawFaces(count, projVerts, ?, ?, [sp]: faces, idxlist, texset)
			if n > 4_500_000 {
				sp := c.Reg(13)
				ev("DRAWFACES", 40, "n=%d r1=0x%X r2=0x%X r3=0x%X sp0=0x%X sp1=0x%X sp2=0x%X sp3=0x%X lr=0x%X",
					c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3), rd(sp), rd(sp+4), rd(sp+8), rd(sp+12), c.Reg(14))
			}
		case 0x352DC: // transform/prep before drawFaces
			if n > 4_500_000 {
				ev("PREPFACES", 40, "r0=0x%X r1=0x%X r2=0x%X r3=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(3), c.Reg(14))
			}
		}
		if len(resolver) > 0 && pc == resolver[len(resolver)-1][2] {
			r := resolver[len(resolver)-1]
			resolver = resolver[:len(resolver)-1]
			ev("ARENA", 120, "fn=0x%X arg=%d -> 0x%X", r[0], r[1], c.Reg(0))
		}
		if traceRet != 0 {
			if pc == traceRet {
				fmt.Printf("--- %s: 0x1779C(r0=1) path (%d instrs) ---\n", traceName, len(traceLog))
				for _, r := range compressRanges(traceLog) {
					fmt.Println("   ", r)
				}
				traceRet = 0
				traceLog = nil
			} else if len(traceLog) < 400000 {
				traceLog = append(traceLog, pc)
			}
		} else if pc == 0x1779C && c.Reg(0) == 1 && len(traceAts) > 0 && n >= traceAts[0] {
			traceName = fmt.Sprintf("at %d", n)
			traceAts = traceAts[1:]
			traceRet = c.Reg(14)
		}
		if pc == 0x28BE0 && c.Reg(4) == 0x295810 && integLogs < 8 {
			integLogs++
			fmt.Printf("[%9d t#%-4d] INTEG      opp integrator vel=0x%X,0x%X,0x%X frame=%d bt:%s\n",
				n, m.CurrentTaskNum(), rd(0x295858), rd(0x29585C), rd(0x295860), rd(0x41D24), backtrace())
		}
		if pc == 0x12918 {
			fmt.Printf("[%9d t#%-4d] STAMP      [0x3F97C]=%d (frame+1) car? bt:%s\n",
				n, m.CurrentTaskNum(), rd(0x41D24)+1, backtrace())
		}
		if *pokeAddr != 0 && n == 6_000_000 {
			m.Write(uint32(*pokeAddr), 1)
			fmt.Printf("[%9d] POKE       [0x%X]=1\n", n, *pokeAddr)
		}
		if pc == 0x128FC { // stamp condition: r0 = |speed| of the car being checked
			ev("SPDCHK", 6, "car=0x%X speed=0x%X frame=%d bt:%s", c.Reg(4), c.Reg(0), rd(0x41D24), backtrace())
		}
		// experiment: fake a brief player-car motion (speed scalar [obj+0x5C])
		// during 8.0M..8.3M — does the stamp/opponent-launch/race-clock cascade?
		if false {
			m.Write(0x295D50, 0)
			m.Write(0x295D51, 0)
			m.Write(0x295D52, 0x20)
			m.Write(0x295D53, 0)
		}
		switch pc {
		case 0x206C4: // sim-task hold path taken this frame
			ev("HOLD", 12, "countdown=%d frame=%d", rd(0x41D0C), rd(0x41D24))
		case 0x20740: // started transition
			fmt.Printf("[%9d t#%-4d] STARTED    countdown=%d frame=%d\n", n, m.CurrentTaskNum(), rd(0x41D0C), rd(0x41D24))
		case 0x27C8:
			loopHead++
		case 0x17904:
			worldUpd++
		case 0x1779C:
			carSim[c.Reg(0)]++
			if carSim[c.Reg(0)] <= 2 {
				ev("CARSIM", 20, "car=0x%X seg=%d", c.Reg(0), rd(c.Reg(0)+0x50))
			}
		case 0x1824: // track-stream state machine(ctx)
			ev("STREAM", 40, "ctx=0x%X state=%d next=%d lr=0x%X", c.Reg(0), rd(c.Reg(0)), rd(c.Reg(0)+4), c.Reg(14))
		case 0xD48: // post a load request
			ev("LOADREQ", 40, "r0=0x%X r1=0x%X r2=%d lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(14))
		case 0x134E0: // steer consumer(obj, steerByte)
			if c.Reg(0) != 0x295CF4 {
				ev("STEER-O", 40, "obj=0x%X v=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(14))
			}
		case 0x13B58: // gas consumer(obj, nibble)
			if c.Reg(0) != 0x295CF4 {
				ev("GAS-O", 40, "obj=0x%X v=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(14))
			} else if c.Reg(1) != 0 {
				ev("GAS-P", 12, "obj=0x%X v=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(14))
			}
		}
		// periodic world-state status
		if n-lastStatus >= 1_000_000 && n >= 4_000_000 {
			lastStatus = n
			fmt.Printf("[%9d] STATUS frame=%d simflags=0x%X raceseg=%d racestate=%d simstate=0x%X gate=0x%X loop=%d world=%d p.seg=%d o.seg=%d\n",
				n, rd(0x41D24), rd(0x41D2C), rd(0x3F970), rd(0x3F974), rd(0x41A84), rd(0x40044),
				loopHead, worldUpd, rd(0x295CE8+0x50), rd(0x295804+0x50))
		}
		if len(snapAt) > 0 && n >= snapAt[0] {
			buf := make([]byte, 3*1024*1024)
			for i := range buf {
				buf[i] = mm.Read(uint32(i))
			}
			os.WriteFile(fmt.Sprintf("%s-%d.bin", *snapPrefix, snapAt[0]), buf, 0644)
			fmt.Printf("[%9d] SNAP -> %s-%d.bin\n", n, *snapPrefix, snapAt[0])
			snapAt = snapAt[1:]
		}
	}

	res := m.Run(*steps)
	fmt.Printf("stopped: %s after %d steps pc=%X\n", res.Reason, res.Steps, res.PC)

	fmt.Println("\naudio SWI census:")
	var sws []uint32
	for s := range audioSWI {
		sws = append(sws, s)
	}
	sort.Slice(sws, func(i, j int) bool { return sws[i] < sws[j] })
	for _, s := range sws {
		fmt.Printf("  SWI 0x%-6X x%d\n", s, audioSWI[s])
	}

	fmt.Println("\nSendSignal census:")
	type sk struct {
		k [2]uint32
		c int
	}
	var sks []sk
	for k, c := range sendSig {
		sks = append(sks, sk{k, c})
	}
	sort.Slice(sks, func(i, j int) bool { return sks[i].c > sks[j].c })
	for i, s := range sks {
		if i >= 30 {
			break
		}
		fmt.Printf("  task=%-6d sigs=0x%-8X x%d\n", s.k[0], s.k[1], s.c)
	}

	fmt.Println("\ncar-sim census:")
	for car, c := range carSim {
		fmt.Printf("  car=0x%X x%d\n", car, c)
	}

	fmt.Println("\ntasks:")
	for _, s := range m.TaskSummary() {
		fmt.Println(" ", s)
	}
	if tty := m.TTY(); tty != "" {
		fmt.Printf("\n[TTY]\n%s\n", tty)
	}
}

// parsePadScript mirrors bootoracle's syntax.
func parsePadScript(s string) ([]threedo.PadStep, error) {
	bits := map[string]uint32{
		"a": threedo.PadA, "b": threedo.PadB, "c": threedo.PadC,
		"start": threedo.PadStart, "p": threedo.PadStart, "x": threedo.PadX,
		"up": threedo.PadUp, "down": threedo.PadDown,
		"left": threedo.PadLeft, "right": threedo.PadRight,
		"ls": threedo.PadLeftShift, "rs": threedo.PadRightShift, "0": 0,
	}
	var script []threedo.PadStep
	for _, entry := range strings.Split(s, ",") {
		step, names, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok {
			return nil, fmt.Errorf("pad entry %q: want STEP:buttons", entry)
		}
		at, err := strconv.ParseUint(step, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("pad entry %q: %v", entry, err)
		}
		var buttons uint32
		for _, nm := range strings.Split(names, "+") {
			bit, ok := bits[strings.ToLower(strings.TrimSpace(nm))]
			if !ok {
				return nil, fmt.Errorf("pad entry %q: unknown button %q", entry, nm)
			}
			buttons |= bit
		}
		script = append(script, threedo.PadStep{AtStep: at, Buttons: buttons})
	}
	sort.Slice(script, func(i, j int) bool { return script[i].AtStep < script[j].AtStep })
	return script, nil
}

func readCStr(m *threedo.Machine, addr uint32) string {
	var b []byte
	for i := uint32(0); i < 64; i++ {
		c := m.Read(addr + i)
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "memtrace:", err)
	os.Exit(1)
}

// compressRanges renders a PC trace as "start-end (xN)" call segments: each
// contiguous run of ascending PCs is one segment; branches start a new one.
func compressRanges(pcs []uint32) []string {
	var out []string
	if len(pcs) == 0 {
		return out
	}
	start, prev := pcs[0], pcs[0]
	count := 1
	for _, pc := range pcs[1:] {
		if pc == prev+4 {
			prev = pc
			count++
			continue
		}
		out = append(out, fmt.Sprintf("%05X-%05X (%d)", start, prev, count))
		start, prev = pc, pc
		count = 1
		if len(out) > 3000 {
			out = append(out, "...truncated")
			break
		}
	}
	out = append(out, fmt.Sprintf("%05X-%05X (%d)", start, prev, count))
	return out
}
