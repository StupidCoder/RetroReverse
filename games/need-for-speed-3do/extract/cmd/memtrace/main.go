// memtrace is a throwaway step-stamped event tracer. Current focus: the frozen
// race sim. Main task#1 busy-waits in its GetMsg pump (SWI 0x10013 via helper
// 0x3A12C) during the race, while the attract/credits states advance fine. This
// trace logs every message queue/dequeue and signal send with step stamps so the
// working state (attract, ~28-30M steps) can be compared with the frozen race
// (>40M steps): which port does the pump poll, and who feeds it when it works?
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	image := flag.String("image", "", "3DO disc image")
	steps := flag.Uint64("steps", 60000000, "max instructions")
	stall := flag.Int("stall", 400, "deadlock-guard tolerance multiplier")
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
	// count-limited event logger: limits apply per 10M-step phase so both the
	// working attract window and the frozen race window stay visible.
	counts := map[string]int{}
	ev := func(kind string, limit int, format string, args ...any) {
		key := fmt.Sprintf("%s@%d", kind, n/10_000_000)
		counts[key]++
		if counts[key] <= limit {
			fmt.Printf("[%9d t#%-4d] %-10s %s\n", n, m.CurrentTaskNum(), kind, fmt.Sprintf(format, args...))
		}
	}

	// GetMsg poll census: port -> polls, successes, step of first/last poll.
	type pollStat struct{ polls, hits int; first, last uint64 }
	getStats := map[uint32]*pollStat{}
	// pending GetMsg: watch the instruction after the SWI for the result.
	var pendPC, pendPort uint32

	m.OnSWI = func(mm *threedo.Machine, from, swi uint32) {
		c := mm.CPU
		switch swi {
		case 0x10010: // PutMsg(port, msg)
			ev("PutMsg", 40, "port=%d msg=%d from=0x%X", c.Reg(0), c.Reg(1), from)
		case 0x10012: // ReplyMsg(msg, result)
			ev("ReplyMsg", 40, "msg=%d result=0x%X from=0x%X", c.Reg(0), c.Reg(1), from)
		case 0x10013: // GetMsg(port)
			port := c.Reg(0)
			s := getStats[port]
			if s == nil {
				s = &pollStat{first: n}
				getStats[port] = s
			}
			s.polls++
			s.last = n
			if s.polls <= 3 {
				ev("GetMsg", 60, "port=%d from=0x%X (poll #%d)", port, from, s.polls)
			}
			pendPC, pendPort = from+4, port
		case 0x10002: // SendSignal(task, sigs)
			ev("SendSig", 40, "task=%d sigs=0x%X from=0x%X", c.Reg(0), c.Reg(1), from)
		case 0x10001: // WaitSignal(mask)
			ev("WaitSig", 30, "mask=0x%X from=0x%X", c.Reg(0), from)
		case 0x40016: // MonitorAttachment(att, cue, cueAt)
			ev("MONATT", 60, "att=%d cue=%d cueAt=0x%X from=0x%X", c.Reg(0), c.Reg(1), c.Reg(2), from)
		case 0x40001: // StartInstrument(ins, tags)
			ev("STARTINS", 40, "ins=%d tags=0x%X from=0x%X", c.Reg(0), c.Reg(1), from)
		case 0x40012: // StartAttachment(att, tags)
			ev("STARTATT", 40, "att=%d tags=0x%X from=0x%X", c.Reg(0), c.Reg(1), from)
		}
	}
	m.OnMsgQueue = func(mm *threedo.Machine, port, msg int32, why string) {
		ev("QUEUE", 60, "%s -> port=%d msg=%d", why, port, msg)
	}
	m.WatchLo, m.WatchHi = 0x41D2C, 0x41D30
	m.OnWrite = func(addr, val, pc uint32) {
		if addr == 0x41D2F { // low byte of the flags word
			ev("FLAGW", 100, "[0x41D2C] byte3=%d from pc=0x%X lr=0x%X", val, pc, m.CPU.Reg(14))
		}
	}

	rd := func(a uint32) uint32 {
		return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
	}
	// Hot-PC histogram of the main task inside a step window (the static phase),
	// to see the race loop's actual per-iteration path.
	hot := map[[2]uint32]uint64{} // {task, pc} -> count
	var hotLR = map[uint32]uint64{}
	var traceRet uint32
	var traced bool
	var traceLog []uint32
	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		n++
		if pc == pendPC {
			pendPC = 0
			if r0 := mm.CPU.Reg(0); r0 != 0 {
				s := getStats[pendPort]
				s.hits++
				if s.hits <= 6 || s.hits%500 == 0 {
					ev("GOTMSG", 80, "port=%d -> msg=%d (hit #%d)", pendPort, r0, s.hits)
				}
			}
		}
		c := mm.CPU
		// One-shot full PC trace of the 30Hz full-sim call (0x1779C, even frames).
		if traceRet != 0 {
			if pc == traceRet && len(traceLog) > 2 {
				fmt.Printf("--- 0x1779C invocation trace (%d instrs) ---\n", len(traceLog))
				for _, r := range compressRanges(traceLog) {
					fmt.Println("   ", r)
				}
				traceRet = 0
				traceLog = nil
			} else if len(traceLog) < 300000 && mm.CurrentTaskNum() == 4190 {
				traceLog = append(traceLog, pc)
			}
		} else if pc == 0x17904 && n > 26_000_000 && !traced && rd(0x41D2C)&1 != 0 {
			traced = true
			traceRet = c.Reg(14)
			traceLog = traceLog[:0]
		}
		if n > 20_000_000 {
			hot[[2]uint32{uint32(mm.CurrentTaskNum()), pc}]++
			if pc == 0x27C8 { // loop head: sample where it goes
				hotLR[c.Reg(14)]++
			}
		}
		switch pc {
		case 0x34458: // RegisterVBLCallback(func): store in first free slot
			ev("VBLREG", 40, "func=0x%X lr=0x%X", c.Reg(0), c.Reg(14))
		case 0x2804: // race loop: WaitSignal(cueSig) returned r0
			ev("LOOPWAKE", 40, "r0=0x%X frame[+8]=%d", c.Reg(0), rd(0x3E5D4+8))
		case 0x27F8: // race loop: SignalAtTime(cue=[r4+0x10], t=now+4)
			ev("SIGAT", 20, "cue=%d t=%d", c.Reg(0), c.Reg(1))
		case 0x2830: // race loop: pad bits from 0x130E8
			ev("PADBITS", 20, "bits=0x%08X gate[0x40044]=0x%X", c.Reg(0), rd(0x40044))
		case 0x1C94C:
			ev("FRAMESPIN", 10, "gate=0x%X", rd(0x40044))
		case 0x17980: // world update reads simFlags
			ev("SIMFLAGS", 6, "[0x41D2C]=0x%X frame=%d", rd(0x41D2C), rd(0x41D24))
		case 0x1824: // track-stream state machine(ctx): [ctx]=state, [ctx+4]=next
			ev("STREAM", 60, "ctx=0x%X state=%d next=%d lr=0x%X", c.Reg(0), rd(c.Reg(0)), rd(c.Reg(0)+4), c.Reg(14))
		case 0xD48: // post a load request
			ev("LOADREQ", 60, "r0=0x%X r1=0x%X r2=%d lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(2), c.Reg(14))
		case 0x134E0: // steer consumer(obj, steerByte)
			ev("STEER", 12, "obj=0x%X v=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(14))
		case 0x13B58: // gas consumer(obj, nibble)
			ev("GAS", 12, "obj=0x%X v=0x%X lr=0x%X", c.Reg(0), c.Reg(1), c.Reg(14))
		}
	}

	m.StallTolerance = *stall
	m.NoStreams = true
	m.PadScript = []threedo.PadStep{{AtStep: 4000000, Buttons: threedo.PadStart}, {AtStep: 4300000, Buttons: 0}, {AtStep: 16000000, Buttons: threedo.PadA}, {AtStep: 17000000, Buttons: 0}, {AtStep: 18000000, Buttons: threedo.PadA}, {AtStep: 19000000, Buttons: 0}, {AtStep: 20000000, Buttons: threedo.PadA}}

	// DRAM snapshots for state diffing: idle (14M), A-held (28M, 30M).
	snapAt := []uint64{22_000_000, 28_000_000, 30_000_000}
	snaps := map[uint64][]byte{}
	prevOnStep := m.OnStep
	m.OnStep = func(mm *threedo.Machine, pc uint32) {
		prevOnStep(mm, pc)
		if len(snapAt) > 0 && n >= snapAt[0] {
			buf := make([]byte, 3*1024*1024)
			for i := range buf {
				buf[i] = mm.Read(uint32(i))
			}
			snaps[snapAt[0]] = buf
			snapAt = snapAt[1:]
		}
	}

	res := m.Run(*steps)
	for at, buf := range snaps {
		os.WriteFile(fmt.Sprintf("dram-%d.bin", at), buf, 0644)
	}
	fmt.Printf("stopped: %s after %d steps pc=%X\n", res.Reason, res.Steps, res.PC)

	fmt.Println("\nGetMsg poll census (port: polls/hits, first..last step):")
	var ports []uint32
	for p := range getStats {
		ports = append(ports, p)
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
	for _, p := range ports {
		s := getStats[p]
		fmt.Printf("  port %-6d polls=%-9d hits=%-6d steps [%d .. %d]\n", p, s.polls, s.hits, s.first, s.last)
	}

	fmt.Println("\nhot {task,pc} after 20M steps (top 80):")
	type hc struct {
		key [2]uint32
		c   uint64
	}
	var hcs []hc
	perTask := map[uint32]uint64{}
	for p, c := range hot {
		hcs = append(hcs, hc{p, c})
		perTask[p[0]] += c
	}
	fmt.Println("steps per task in window:")
	for t, c := range perTask {
		fmt.Printf("  task#%-5d %d\n", t, c)
	}
	sort.Slice(hcs, func(i, j int) bool { return hcs[i].c > hcs[j].c })
	for i, h := range hcs {
		if i >= 80 {
			break
		}
		fmt.Printf("  t#%-5d %8d  %s\n", h.key[0], h.c, m.DisasmAt(h.key[1]))
	}
	if f, err := os.Create("hot.csv"); err == nil {
		for _, h := range hcs {
			fmt.Fprintf(f, "%d,%08X,%d\n", h.key[0], h.key[1], h.c)
		}
		f.Close()
	}
	fmt.Println("\nloop-head LR census:")
	for lr, c := range hotLR {
		fmt.Printf("  lr=0x%X x%d\n", lr, c)
	}

	fmt.Println("\ntasks:")
	for _, s := range m.TaskSummary() {
		fmt.Println(" ", s)
	}
	fmt.Println("\nitems:")
	for _, s := range m.ItemsSummary() {
		fmt.Println(" ", s)
	}
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
		if len(out) > 800 {
			out = append(out, "...truncated")
			break
		}
	}
	out = append(out, fmt.Sprintf("%05X-%05X (%d)", start, prev, count))
	return out
}
