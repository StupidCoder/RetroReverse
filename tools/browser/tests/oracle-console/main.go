// Private-media differential harness. No media or checkpoints are distributed.
package main

import (
	"fmt"
	"image/png"
	"os"
	"retroreverse.com/tools/lib/iso9660"
	"retroreverse.com/tools/platform/gc"
	"retroreverse.com/tools/platform/ps2"
	"strconv"
	"strings"
	"time"
)

func hash(b []byte) uint32 {
	h := uint32(2166136261)
	for _, v := range b {
		h = (h ^ uint32(v)) * 16777619
	}
	return h
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	if os.Args[1] == "xbox" {
		runX86()
		return
	}
	if os.Args[1] == "gba" || os.Args[1] == "dc" {
		runAdvance()
		return
	}
	if os.Args[1] == "ps2" {
		runPS2()
		return
	}
	runGC()
}
func runGC() {
	d, e := gc.Open(os.Args[2])
	check(e)
	defer d.Close()
	m, e := gc.NewMachine(d)
	check(e)
	defer m.Close()
	m.SingleThreaded = true
	if p := os.Getenv("RR_GO_STATE"); p != "" {
		check(m.LoadStateFile(p))
	} else {
		_, e = m.RunApploader()
		check(e)
	}
	if p := os.Getenv("RR_EXPORT"); p != "" {
		exportState("gc", p, m)
	}
	m.OnDisplay = func(m *gc.Machine) { m.StopRequested = true }
	n, _ := strconv.ParseUint(os.Args[3], 10, 64)
	end := m.VIField() + n
	start := time.Now()
	for m.VIField() < end {
		if os.Getenv("GC_AUTO") != "" {
			b := uint16(0)
			if m.VIField()%120 < 10 {
				b = 0x1100
			}
			m.SetPadButtons(0, b)
		}
		r := m.Run(10000)
		if r.Reason != "step budget exhausted" && r.Reason != "stop requested" {
			panic(r.Reason)
		}
	}
	if p := os.Getenv("RR_SAVE_END"); p != "" {
		exportState("gc", p, m)
	}
	f, e := m.RenderXFB()
	check(e)
	fmt.Printf("{\"steps\":%d,\"frames\":%d,\"pc\":%d,\"ram\":%d,\"rgba\":%d,\"seconds\":%f}\n", m.Instrs, m.VIField(), m.CPU.PC, hash(m.RAM), hash(f.Pix), time.Since(start).Seconds())
	if len(os.Args) > 4 {
		o, _ := os.Create(os.Args[4])
		defer o.Close()
		png.Encode(o, f)
	}
}
func runPS2() {
	f, e := os.Open(os.Args[2])
	check(e)
	defer f.Close()
	st, e := f.Stat()
	check(e)
	v, e := iso9660.Open(f, st.Size())
	check(e)
	cnf, e := v.ReadFile("SYSTEM.CNF")
	check(e)
	exe := ""
	for _, line := range strings.Split(string(cnf), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "BOOT2" {
			exe = strings.TrimSpace(value)
			if p := strings.Index(exe, ":"); p >= 0 {
				exe = exe[p+1:]
			}
		}
	}
	b, e := v.ReadFile(exe)
	check(e)
	elf, e := ps2.LoadELF(b)
	check(e)
	m := ps2.NewMachine()
	defer m.Close()
	m.SingleThreaded = true
	m.SetVolume(v)
	if p := os.Getenv("PS2_BIOS"); p != "" {
		bios, e := os.ReadFile(p)
		check(e)
		m.SetBIOS(bios)
	}
	m.LoadExecutable(elf)
	check(m.RebootIOP())
	if p := os.Getenv("RR_GO_STATE"); p != "" {
		check(m.LoadStateFile(p))
	}
	if p := os.Getenv("RR_EXPORT"); p != "" {
		exportState("ps2", p, m)
	}
	m.OnVBlank = func(m *ps2.Machine) { m.StopRequested = true }
	n, _ := strconv.ParseUint(os.Args[3], 10, 64)
	end := uint64(m.VBlanks()) + n
	start := time.Now()
	for uint64(m.VBlanks()) < end {
		r := m.Run(2000000)
		if r.Reason != "step budget exhausted" && r.Reason != "stop requested" {
			panic(r.Reason)
		}
	}
	pix, w, h := m.GSFrame()
	fmt.Printf("{\"frames\":%d,\"pc\":%d,\"ram\":%d,\"rgba\":%d,\"width\":%d,\"height\":%d,\"seconds\":%f}\n", m.VBlanks(), m.CPU.PC, hash(m.ReadMem(0, 32*1024*1024)), hash(pix), w, h, time.Since(start).Seconds())
}
