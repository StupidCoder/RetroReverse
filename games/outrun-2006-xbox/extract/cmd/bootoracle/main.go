// bootoracle boots an original-Xbox title's XBE on the Xbox machine
// (tools/platform/xbox) and runs its XDK/CRT boot code, exposing the standard oracle
// instrumentation (STANDARDS §3). Where xbeinfo (Phase 0) reads the disc's containers
// and the executable's header statically, this one RUNS the code they hold: it loads
// default.xbe at its fixed base, high-level-emulates the xboxkrnl exports the boot path
// reaches, and stops at the first NV2A push-buffer kick — or, before that, at the first
// kernel ordinal not yet modelled, naming it, so each run reports concretely how far
// the bring-up reaches.
//
// Usage:
//
//	bootoracle -image DISC.iso [-steps N] [-trace] [-tracen N] [-bp SEG:OFF]...
//	           [-watch ADDR[:LEN]]... [-v] [-ordinals] [-dump ADDR:LEN]... [-stack]
//	           [-savestate FILE] [-loadstate FILE]
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/xbox"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func parseNum(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func main() {
	image := flag.String("image", "", "Xbox disc image (.iso / XISO)")
	xbePath := flag.String("xbe", "/default.xbe", "path of the XBE within the disc to boot")
	steps := flag.String("steps", "50000000", "instruction budget (hex with 0x, else decimal)")
	verbose := flag.Bool("v", false, "log kernel calls and events live")
	ordinals := flag.Bool("ordinals", false, "after the run, print the histogram of xboxkrnl ordinals called")
	stackDump := flag.Bool("stack", false, "on halt, dump the top of the stack (the caller's argument frame)")
	trace := flag.Bool("trace", false, "trace executed instructions")
	tracen := flag.Int("tracen", 200, "limit -trace to this many instructions")
	savestate := flag.String("savestate", "", "after the run, write a machine snapshot to this file")
	pngOut := flag.String("png", "", "after the run, write the display scanout to this PNG")
	surfOut := flag.String("surfpng", "", "after the run, write the Kelvin render surface (AA-resolved) to this PNG")
	loadstate := flag.String("loadstate", "", "restore a machine snapshot before running")
	gpu := flag.Bool("gpu", false, "Phase C: run the NV2A DMA pusher on each kick (do not stop at first push)")
	survey := flag.Bool("survey", false, "with -gpu: record the PGRAPH method surface and print it")
	var dumps multiFlag
	flag.Var(&dumps, "dump", "hex-dump ADDR:LEN of memory after the run (hex); repeatable")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "bootoracle: -image is required")
		os.Exit(2)
	}

	budget, err := parseNum(*steps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: bad -steps %q: %v\n", *steps, err)
		os.Exit(2)
	}

	disc, err := xbox.Open(*image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: open image: %v\n", err)
		os.Exit(1)
	}
	defer disc.Close()

	xbeBytes, err := disc.ReadFile(*xbePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: read %s: %v\n", *xbePath, err)
		os.Exit(1)
	}
	xbe, err := xbox.ParseXBE(xbeBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: parse XBE: %v\n", err)
		os.Exit(1)
	}

	m, err := xbox.NewMachine(xbe, disc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootoracle: build machine: %v\n", err)
		os.Exit(1)
	}
	m.SetVerbose(*verbose)
	if *gpu {
		m.EnableGPU()
		if *survey {
			m.PGraph().SetSurvey(true)
		}
	}

	if *loadstate != "" {
		if err := m.LoadStateFile(*loadstate); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: loadstate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("restored machine state from %s\n", *loadstate)
	}

	if *trace {
		m.SetTrace(*tracen) // print the first -tracen executed instructions (PC trail)
	}

	reason, n := m.Run(budget)
	fmt.Printf("\n=== run ended: %s after %d instructions ===\n", reason, n)
	fmt.Print(m.Report())

	if *stackDump && m.CPU.Halted {
		sp := m.CPU.Regs[4]
		fmt.Printf("stack @ ESP=%08X:\n", sp)
		for i := uint32(0); i < 12; i++ {
			a := sp + i*4
			fmt.Printf("  [ESP+%02X] %08X = %08X\n", i*4, a, m.MemRead32(a))
		}
		// The call site: disassemble a window ending at the return address so the
		// argument pushes (which pin the ordinal's signature) are visible.
		ret := m.CallerReturnAddr()
		if ret > 48 {
			fmt.Printf("call site (around return %08X):\n", ret)
			for _, line := range m.DisasmForward(ret-40, 20) {
				fmt.Println(line)
			}
		}
	}

	for _, d := range dumps {
		parts := strings.SplitN(d, ":", 2)
		addr, _ := parseNum(parts[0])
		length := uint64(64)
		if len(parts) == 2 {
			length, _ = parseNum(parts[1])
		}
		hexDump(m, uint32(addr), uint32(length))
	}

	if *ordinals {
		fmt.Println("\nxboxkrnl ordinals reached:")
		for _, line := range m.OrdinalHistogram() {
			fmt.Println(line)
		}
	}

	if *gpu && *survey {
		fmt.Println()
		for _, line := range m.PGraph().SurveyReport() {
			fmt.Println(line)
		}
	}

	if *pngOut != "" {
		if data, err := m.FramePNG(); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: png: %v\n", err)
		} else if err := os.WriteFile(*pngOut, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: png: %v\n", err)
		} else {
			fmt.Printf("wrote frame to %s\n", *pngOut)
		}
	}
	if *surfOut != "" {
		if data, err := m.SurfacePNG(); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: surfpng: %v\n", err)
		} else if err := os.WriteFile(*surfOut, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: surfpng: %v\n", err)
		} else {
			fmt.Printf("wrote render surface to %s\n", *surfOut)
		}
	}
	if *savestate != "" {
		if err := m.SaveStateFile(*savestate); err != nil {
			fmt.Fprintf(os.Stderr, "bootoracle: savestate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("wrote machine snapshot to %s\n", *savestate)
	}
}

func hexDump(m *xbox.Machine, addr, length uint32) {
	fmt.Printf("dump %08X..%08X:\n", addr, addr+length)
	for off := uint32(0); off < length; off += 16 {
		fmt.Printf("  %08X ", addr+off)
		var ascii strings.Builder
		for i := uint32(0); i < 16 && off+i < length; i++ {
			b := m.MemReadByte(addr + off + i)
			fmt.Printf("%02X ", b)
			if b >= 0x20 && b < 0x7F {
				ascii.WriteByte(b)
			} else {
				ascii.WriteByte('.')
			}
		}
		fmt.Printf(" %s\n", ascii.String())
	}
}
