package main

import (
	"fmt"
	"image/png"
	"os"
	"reflect"
	"retroreverse.com/tools/platform/dc"
	"retroreverse.com/tools/platform/gba"
	"retroreverse.com/tools/platform/gba/gbamachine"
	"strconv"
	"time"
)

func fieldBytes(m any, name string) []byte {
	v := reflect.ValueOf(m).Elem().FieldByName(name)
	b := make([]byte, v.Len())
	for i := range b {
		b[i] = byte(v.Index(i).Uint())
	}
	return b
}
func runAdvance() {
	n, _ := strconv.ParseUint(os.Args[3], 10, 64)
	if os.Args[1] == "gba" {
		b, e := os.ReadFile(os.Args[2])
		check(e)
		rom, e := gba.Parse(b)
		check(e)
		m := gbamachine.New(rom)
		if p := os.Getenv("RR_GO_STATE"); p != "" {
			check(m.LoadState(p))
		}
		if p := os.Getenv("RR_EXPORT"); p != "" {
			exportState("gba", p, m)
		}
		end := m.Frame() + n
		start := time.Now()
		for m.Frame() < end {
			if os.Getenv("RR_AUTO") != "" {
				b := uint16(0)
				if m.Frame()%120 < 8 {
					b = 9
				}
				m.SetKeys(b)
			}
			m.RunFrames(1, 10000)
			if h, r := m.Halted(); h {
				panic(r)
			}
		}
		f := m.Screen()
		fmt.Printf("{\"steps\":%d,\"instrs\":%d,\"frames\":%d,\"pc\":%d,\"ram\":%d,\"iwram\":%d,\"vram\":%d,\"rgba\":%d,\"seconds\":%f}\n", m.Steps, m.Instrs(), m.Frame(), m.PC(), hash(fieldBytes(m, "ewram")), hash(fieldBytes(m, "iwram")), hash(fieldBytes(m, "vram")), hash(f.Pix), time.Since(start).Seconds())
		if p := os.Getenv("RR_SAVE_END"); p != "" {
			exportState("gba", p, m)
		}
		if len(os.Args) > 4 {
			o, e := os.Create(os.Args[4])
			check(e)
			defer o.Close()
			png.Encode(o, f)
		}
	} else {
		d, e := dc.OpenDisc(os.Args[2])
		check(e)
		m := dc.NewMachine(d)
		check(m.Boot())
		if p := os.Getenv("RR_GO_STATE"); p != "" {
			check(m.LoadStateFile(p))
		}
		if p := os.Getenv("RR_EXPORT"); p != "" {
			exportState("dc", p, m)
		}
		m.OnDisplay = func(uint64) { m.StopRequested = true }
		end := m.Fields + n
		start := time.Now()
		for m.Fields < end {
			if os.Getenv("RR_AUTO") != "" {
				m.Pad.Buttons = 0xffff
				if m.Fields%120 < 8 {
					m.Pad.Buttons = ^uint16(12)
				}
			}
			r := m.Run(10000, dc.RunConfig{NoSpin: true})
			if r.Reason != "steps" && r.Reason != "stop requested" {
				panic(r.Reason)
			}
		}
		f, e := m.RenderFB()
		rgba := uint32(0)
		if e == nil {
			rgba = hash(f.Pix)
		}
		fmt.Printf("{\"steps\":%d,\"frames\":%d,\"pc\":%d,\"ram\":%d,\"vram\":%d,\"aica\":%d,\"rgba\":%d,\"seconds\":%f}\n", m.Instrs, m.Fields, m.CPU.PC, hash(m.RAM), hash(m.VRAM), hash(m.AICARAM), rgba, time.Since(start).Seconds())
		if p := os.Getenv("RR_SAVE_END"); p != "" {
			exportState("dc", p, m)
		}
		if len(os.Args) > 4 && e == nil {
			o, e := os.Create(os.Args[4])
			check(e)
			defer o.Close()
			png.Encode(o, f)
		}
	}
}
