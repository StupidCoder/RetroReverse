package main

import (
	"fmt"
	"image/png"
	"os"
	"retroreverse.com/tools/platform/xbox"
	"strconv"
	"time"
)

func runX86() {
	d, e := xbox.Open(os.Args[2])
	check(e)
	defer d.Close()
	b, e := d.ReadFile("default.xbe")
	check(e)
	x, e := xbox.ParseXBE(b)
	check(e)
	m, e := xbox.NewMachine(x, d)
	check(e)
	m.EnableGPU()
	m.AttachPad(0)
	if p := os.Getenv("RR_GO_STATE"); p != "" {
		check(m.LoadStateFile(p))
	}
	if p := os.Getenv("RR_EXPORT"); p != "" {
		exportState("xbox", p, m)
	}
	frames := uint64(0)
	m.OnFlip = func(m *xbox.Machine) { frames++; m.StopRequested = true }
	n, _ := strconv.ParseUint(os.Args[3], 10, 64)
	maxSteps := uint64(^uint64(0))
	if s := os.Getenv("RR_MAX_STEPS"); s != "" {
		maxSteps, _ = strconv.ParseUint(s, 10, 64)
	}
	start := time.Now()
	for frames < n && !m.CPU.Halted && m.CPU.Steps < maxSteps {
		m.Run(10000)
	}
	if p := os.Getenv("RR_SAVE_END"); p != "" {
		exportState("xbox", p, m)
	}
	if os.Getenv("RR_LOG") != "" {
		for _, s := range m.Log {
			fmt.Fprintln(os.Stderr, s)
		}
	}
	f, e := m.RenderPresented()
	if e != nil {
		fmt.Printf("steps=%d pc=%x ram=%d error=%s\n", m.CPU.Steps, m.CPU.IP, hash(m.RAM), e)
		return
	}
	fmt.Printf("{\"steps\":%d,\"frames\":%d,\"pc\":%d,\"ram\":%d,\"rgba\":%d,\"seconds\":%f,\"halt\":%q}\n", m.CPU.Steps, frames, m.CPU.IP, hash(m.RAM), hash(f.Pix), time.Since(start).Seconds(), m.CPU.HaltReason)
	if len(os.Args) > 4 {
		o, e := os.Create(os.Args[4])
		check(e)
		defer o.Close()
		check(png.Encode(o, f))
	}
}
