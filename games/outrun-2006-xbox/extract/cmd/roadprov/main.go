// roadprov asks the frame debugger's pixel census who draws given pixels of a
// race frame — built to attribute the road surface, which the course export
// lacks. For each -px X,Y it prints the owning draw command and walks the
// command stream back to that draw's vertex-array offsets/formats, the state
// the draw ran against.
package main

import (
	"flag"
	"fmt"
	"strings"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/xboxadapter"
)

func main() {
	image := flag.String("image", "", "Xbox disc image")
	state := flag.String("state", "", "savestate to load")
	pxList := flag.String("px", "200,440 320,300 450,200", "space-separated X,Y pixels to attribute")
	flag.Parse()

	a, err := xboxadapter.New(*image, "/default.xbe")
	if err != nil {
		panic(err)
	}
	_ = debug.OpenSpec{}
	if err := a.LoadStateFile(*state); err != nil {
		panic(err)
	}
	fc, err := a.StepFrame(true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("frame %dx%d, %d commands, %d writers\n", fc.Width, fc.Height, len(fc.Commands), len(fc.Writers))
	describe := func(ci int) {
		c := fc.Commands[ci]
		fmt.Printf("  [%d] %s %s\n", ci, c.Name, c.Decoded)
		// Walk back for this draw's vertex state: the 16 array offsets (0x1720..)
		// and formats (0x1760..), plus texture offsets (0x1B00 bank) — report the
		// most recent value of each before the draw.
		var offs, fmts [16]string
		var texs [4]string
		progStart := ""
		for j := ci - 1; j >= 0; j-- {
			cc := fc.Commands[j]
			m := cc.Op
			switch {
			case m >= 0x1720 && m < 0x1760:
				k := (m - 0x1720) / 4
				if offs[k] == "" {
					offs[k] = fmt.Sprintf("%08X", cc.Words[0])
				}
			case m >= 0x1760 && m < 0x17A0:
				k := (m - 0x1760) / 4
				if fmts[k] == "" {
					fmts[k] = fmt.Sprintf("%08X", cc.Words[0])
				}
			case m >= 0x1B00 && m < 0x1C00 && (m-0x1B00)%0x40 == 0:
				k := (m - 0x1B00) / 0x40
				if texs[k] == "" {
					texs[k] = fmt.Sprintf("%08X", cc.Words[0])
				}
			case m == 0x1EA0:
				if progStart == "" {
					progStart = fmt.Sprintf("%d", cc.Words[0])
				}
			}
		}
		for k := 0; k < 16; k++ {
			if fmts[k] != "" && fmts[k] != "00000002" {
				fmt.Printf("  a%-2d off=%s fmt=%s\n", k, offs[k], fmts[k])
			}
		}
		fmt.Printf("  tex=%v progStart=%s\n", texs, progStart)
	}
	for _, tok := range strings.Fields(*pxList) {
		var x, y int
		if _, err := fmt.Sscanf(tok, "%d,%d", &x, &y); err != nil {
			continue
		}
		ci := fc.ProvAt(x, y)
		fmt.Printf("\npixel (%d,%d): last writer %d; history:\n", x, y, ci)
		if fc.Overdraw != nil {
			for _, wr := range fc.Overdraw[y*fc.Width+x] {
				tag := "drawn"
				if wr.Rejected {
					tag = "rejected"
				}
				fmt.Printf("    write by cmd %d: rgba(%d,%d,%d,%d) %s\n", wr.CmdIndex, wr.R, wr.G, wr.B, wr.A, tag)
			}
			for _, wr := range fc.Overdraw[y*fc.Width+x] {
				if !wr.Rejected {
					describe(wr.CmdIndex)
				}
			}
		} else if ci >= 0 {
			describe(ci)
		}
	}
}
