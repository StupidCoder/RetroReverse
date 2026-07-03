// paceprobe measures how fast the enemy engines actually run. The gameplay
// main loop ($8BB1) does NOT wait for a frame in mode 2 — it free-runs, so
// the tank/prisoner/mine cadences ($F0 countdown, the $EA and $E8 slot
// cursors) are counted in main-loop PASSES, and a pass costs however many
// cycles one full trip through the engines plus the interleaved raster IRQs
// takes. This harness runs the real game from its entry ($8927) under the
// mos6502 core with a PAL raster-IRQ model, joysticks a game into mode 2,
// and reports (a) frames per main-loop pass and (b) the direct per-mover
// step intervals in frames, measured from their column-variable writes.
//
// Usage: paceprobe [-prg ../extracted/FORT-fast-7000.prg] [-seconds 30]
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/mos6502"
)

const (
	loadAddr    = 0x7000
	entry       = 0x8927
	cyclesLine  = 63  // PAL: 63 cycles per raster line
	linesFrame  = 312 // PAL: 312 lines per frame
	cyclesFrame = cyclesLine * linesFrame
)

// bus is the minimal C64 the game code needs at runtime: RAM, a VIC raster
// model (line counter + raster IRQ), CIA joystick ports and SID noise.
type bus struct {
	ram [65536]byte
	io  [4096]byte // $D000-$DFFF backing store (the IRQ swaps $D022/$D023)
	cpu *mos6502.CPU

	joy0 byte // $DC00 (joystick 2)
	joy1 byte // $DC01 (joystick 1)
	rng  uint32

	rasterTarget byte // last write to $D012
	irqEnabled   bool // $D01A bit 0
	irqPending   bool // raster latch, cleared by writing 1 to $D019 bit 0
	prevLine     int
}

func (b *bus) line() int { return int(b.cpu.Cycles%cyclesFrame) / cyclesLine }

func (b *bus) Read(addr uint16) byte {
	switch {
	case addr == 0xD011:
		v := b.io[0x011] & 0x7F
		if b.line() > 255 {
			v |= 0x80
		}
		return v
	case addr == 0xD012:
		return byte(b.line())
	case addr == 0xD019:
		if b.irqPending {
			return 0x81
		}
		return 0
	case addr == 0xDC00:
		return b.joy0
	case addr == 0xDC01:
		return b.joy1
	case addr == 0xD41B: // SID oscillator 3 noise: the game's RNG source
		b.rng = b.rng*1103515245 + 12345
		return byte(b.rng >> 16)
	case addr >= 0xD000 && addr < 0xE000:
		return b.io[addr-0xD000]
	}
	return b.ram[addr]
}

func (b *bus) Write(addr uint16, v byte) {
	switch {
	case addr == 0xD012:
		b.rasterTarget = v
		b.io[0x012] = v
		return
	case addr == 0xD019:
		if v&1 != 0 {
			b.irqPending = false
		}
		return
	case addr == 0xD01A:
		b.irqEnabled = v&1 != 0
		b.io[0x01A] = v
		return
	case addr >= 0xD000 && addr < 0xE000:
		b.io[addr-0xD000] = v
		return
	}
	b.ram[addr] = v
}

// pollIRQ latches the raster IRQ when the beam crosses the armed line and
// delivers it (hardware push + the KERNAL stub's A/X/Y push + jump through
// the $0314 vector) once the I flag allows.
func (b *bus) pollIRQ() {
	l := b.line()
	if b.irqEnabled && l != b.prevLine && l == int(b.rasterTarget) {
		b.irqPending = true
	}
	b.prevLine = l
	if !b.irqPending || b.cpu.I {
		return
	}
	c := b.cpu
	c.Push(byte(c.PC >> 8))
	c.Push(byte(c.PC))
	p := byte(0x20) // bit 5 always set, B clear (hardware interrupt)
	if c.C {
		p |= 0x01
	}
	if c.Z {
		p |= 0x02
	}
	if c.I {
		p |= 0x04
	}
	if c.D {
		p |= 0x08
	}
	if c.V {
		p |= 0x40
	}
	if c.N {
		p |= 0x80
	}
	c.Push(p)
	c.I = true
	// KERNAL IRQ stub: push A/X/Y, jump through the RAM vector
	c.Push(c.A)
	c.Push(c.X)
	c.Push(c.Y)
	c.PC = uint16(b.ram[0x0314]) | uint16(b.ram[0x0315])<<8
}

func main() {
	prg := flag.String("prg", "../extracted/FORT-fast-7000.prg", "game file ($7000)")
	seconds := flag.Int("seconds", 30, "emulated gameplay seconds to measure")
	flag.Parse()
	if err := run(*prg, *seconds); err != nil {
		fmt.Fprintln(os.Stderr, "paceprobe:", err)
		os.Exit(1)
	}
}

func run(prgPath string, seconds int) error {
	raw, err := os.ReadFile(prgPath)
	if err != nil {
		return err
	}
	if int(raw[0])|int(raw[1])<<8 != loadAddr {
		return fmt.Errorf("%s: not a $7000 file", prgPath)
	}
	b := &bus{joy0: 0xFF, joy1: 0xFF, rng: 0x1234567}
	copy(b.ram[loadAddr:], raw[2:])
	b.cpu = mos6502.NewCPU(b)
	b.cpu.PC = entry

	frame := func() uint64 { return b.cpu.Cycles / cyclesFrame }

	// Measurement state.
	var (
		passFrames  []float64 // frame time of each $8BC3 visit (mode-2 pass start)
		lastPass    = -1.0
		passDeltas  []float64
		stepDeltas  = map[string][]float64{} // mover -> per-step frame deltas
		lastStep    = map[string]float64{}
		gameFrames  uint64
		gameStart   = uint64(0)
		fireHeld    bool
	)
	watch := func(name string, addr uint16, cur *byte) {
		if b.ram[addr] != *cur {
			*cur = b.ram[addr]
			f := float64(b.cpu.Cycles) / cyclesFrame
			if last, ok := lastStep[name]; ok {
				stepDeltas[name] = append(stepDeltas[name], f-last)
			}
			lastStep[name] = f
		}
	}
	var tankCol, prisCol, mineCol byte
	mineSlot := -1

	// budget: the title sequence plus `seconds` of gameplay, in PAL frames
	maxCycles := uint64((seconds+25)*50) * cyclesFrame
	for b.cpu.Cycles < maxCycles && !b.cpu.Halted && gameFrames < uint64(seconds)*50 {
		b.pollIRQ()
		pc := b.cpu.PC
		if pc == 0x8BC3 { // mode-2 engine dispatch: one main-loop pass
			f := float64(b.cpu.Cycles) / cyclesFrame
			passFrames = append(passFrames, f)
			if lastPass >= 0 {
				passDeltas = append(passDeltas, f-lastPass)
			}
			lastPass = f
		}
		b.cpu.Step()

		// Hold fire from the title until gameplay starts, then release.
		mode := b.ram[0x9D]
		if mode == 2 {
			if gameStart == 0 {
				gameStart = frame()
			}
			gameFrames = frame() - gameStart
			if fireHeld {
				b.joy0, b.joy1 = 0xFF, 0xFF
				fireHeld = false
			}
			if mineSlot < 0 { // find an active mine slot once
				for i := 0; i < 39; i++ {
					if b.ram[0x3700+uint16(i)]&0x0F == 2 {
						mineSlot = i
						mineCol = b.ram[0x3500+uint16(i)]
						break
					}
				}
			}
			watch("tank", 0x00C4, &tankCol)
			watch("prisoner", 0x3608, &prisCol)
			if mineSlot >= 0 {
				watch("mine", 0x3500+uint16(mineSlot), &mineCol)
			}
		} else if b.cpu.Instrs > 400_000 && !fireHeld {
			b.joy0, b.joy1 = 0xEF, 0xEF // fire pressed on both ports
			fireHeld = true
		}
	}
	if b.cpu.Halted {
		fmt.Println("HALT:", b.cpu.HaltReason)
		fmt.Printf("trace: PC=$%04X mode=$%02X\n", b.cpu.PC, b.ram[0x9D])
	}

	fmt.Printf("emulated %.1f frames total, final mode $%02X, %d mode-2 passes\n",
		float64(b.cpu.Cycles)/cyclesFrame, b.ram[0x9D], len(passFrames))
	report := func(name string, xs []float64) {
		if len(xs) == 0 {
			fmt.Printf("%-10s no samples\n", name)
			return
		}
		sort.Float64s(xs)
		var sum float64
		for _, x := range xs {
			sum += x
		}
		fmt.Printf("%-10s n=%4d  mean %.2f  median %.2f  p10 %.2f  p90 %.2f (frames)\n",
			name, len(xs), sum/float64(len(xs)), xs[len(xs)/2], xs[len(xs)/10], xs[len(xs)*9/10])
	}
	report("pass", passDeltas)
	report("tank", stepDeltas["tank"])
	report("prisoner", stepDeltas["prisoner"])
	report("mine", stepDeltas["mine"])
	return nil
}
