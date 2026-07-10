// physprobe instruments the car-physics code paths under the PSX oracle.
// It restores a race savestate, drives the car with a scripted pad, and logs
// the player car state block (0x80080194) once per frame — the tool that maps
// the physics fields and the collision/drift behaviour.
//
// Usage:
//
//	physprobe -image DISC -load STATE [-press SCRIPT] [-window N] [-o out.csv]
//	          [-car ADDR] [-len N]
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/psx"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "physprobe: "+format+"\n", a...)
	os.Exit(1)
}

const pcFrame = 0x8001B9D4 // steering-input routine: runs once per physics frame

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	load := flag.String("load", "", "machine savestate to restore")
	press := flag.String("press", "", "pad script BUTTON@STEP:HOLD,...")
	window := flag.Uint64("window", 30_000_000, "steps to run")
	out := flag.String("o", "", "write per-frame CSV here (default stdout)")
	carS := flag.String("car", "80080194", "car state base address (hex)")
	lenF := flag.Int("len", 0xC0, "bytes of car state to log per frame")
	flag.Parse()
	if *image == "" || *load == "" {
		die("need -image and -load")
	}
	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}
	_, exe, err := vol.BootEXE()
	if err != nil {
		die("%v", err)
	}
	m := psx.NewMachine()
	m.SetDisc(vol)
	m.ISRHandler = 0x8004DF48
	if *press != "" {
		script, err := psx.ParsePress(*press)
		if err != nil {
			die("press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)
	if err := m.LoadStateFile(*load); err != nil {
		die("load state: %v", err)
	}

	var car uint32
	fmt.Sscanf(*carS, "%x", &car)
	w := os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			die("%v", err)
		}
		defer f.Close()
		w = f
	}

	// Header: one column per 32-bit word of the car block.
	fmt.Fprintf(w, "step")
	for o := 0; o < *lenF; o += 4 {
		fmt.Fprintf(w, ",+%02X", o)
	}
	fmt.Fprintln(w)

	m.OnStep = func(mm *psx.Machine, pc uint32) {
		if pc != pcFrame {
			return
		}
		fmt.Fprintf(w, "%d", mm.CPU.Steps)
		for o := uint32(0); o < uint32(*lenF); o += 4 {
			v := int32(mm.Read(car+o)) | int32(mm.Read(car+o+1))<<8 |
				int32(mm.Read(car+o+2))<<16 | int32(mm.Read(car+o+3))<<24
			fmt.Fprintf(w, ",%d", v)
		}
		fmt.Fprintln(w)
	}
	res := m.Run(*window)
	fmt.Fprintln(os.Stderr, res)
}
