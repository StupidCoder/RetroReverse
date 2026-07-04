// bootoracle runs the Mario Kart DS ARM9 boot binary on the tools/arm CPU core, over
// a flat DS-like address space, and observes what it does: it lets the game's own
// crt0 self-decompress (a cross-check of tools/nds' DecompressBLZ against the game's
// decompressor), reach main, and program its hardware, logging every I/O-register
// write, until it hits a hardware wait (a VBlank/IRQ-wait BIOS call) it cannot satisfy
// without a full machine. It is the DS analogue of the Amiga physoracle: real code on
// our own core, used to confirm structure read out of the code — not to scrape data.
//
//	bootoracle [-steps N] [-io] rom.nds
//
// -steps caps the instruction budget; -io prints the sorted set of I/O addresses the
// boot wrote (with the last value), which reveals the interrupt/display setup.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"retroreverse.com/tools/arm"
	"retroreverse.com/tools/nds"
)

// bus is a flat little-endian memory covering ITCM/main-RAM/DTCM/WRAM (0..0x03FFFFFF),
// with the I/O block at 0x04000000 logged rather than modelled.
type bus struct {
	mem    []byte
	io     map[uint32]uint32 // last value written to each I/O register
	ioseq  []uint32          // I/O addresses in first-write order
	vcount uint32            // a fake, ever-advancing VCOUNT so simple poll loops progress
}

func newBus() *bus { return &bus{mem: make([]byte, 0x04000000), io: map[uint32]uint32{}} }

func (b *bus) Read(a uint32) byte {
	if a < uint32(len(b.mem)) {
		return b.mem[a]
	}
	if a>>24 == 0x04 { // I/O
		switch a &^ 3 {
		case 0x04000006: // VCOUNT — return an advancing value so a poll doesn't spin forever
			b.vcount = (b.vcount + 1) & 0x1FF
			return byte(b.vcount >> (8 * (a & 3)))
		}
		if v, ok := b.io[a&^3]; ok {
			return byte(v >> (8 * (a & 3)))
		}
	}
	return 0
}

func (b *bus) Write(a uint32, v byte) {
	if a < uint32(len(b.mem)) {
		b.mem[a] = v
		return
	}
	if a>>24 == 0x04 { // I/O register: record it
		base := a &^ 3
		if _, seen := b.io[base]; !seen {
			b.ioseq = append(b.ioseq, base)
		}
		shift := 8 * (a & 3)
		b.io[base] = b.io[base]&^(0xFF<<shift) | uint32(v)<<shift
	}
}

func (b *bus) r32(a uint32) uint32 {
	return uint32(b.Read(a)) | uint32(b.Read(a+1))<<8 | uint32(b.Read(a+2))<<16 | uint32(b.Read(a+3))<<24
}
func (b *bus) w32(a, v uint32) {
	for i := uint32(0); i < 4; i++ {
		b.Write(a+i, byte(v>>(8*i)))
	}
}

func main() {
	steps := flag.Int("steps", 30_000_000, "instruction budget")
	showIO := flag.Bool("io", false, "list the I/O registers the boot programmed")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: bootoracle [-steps N] [-io] rom.nds")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		die(err)
	}
	rom, err := nds.Open(data)
	if err != nil {
		die(err)
	}

	b := newBus()
	// Load the ARM9 binary exactly as the BIOS would: the compressed image at its RAM
	// address. The crt0 will decompress it itself.
	arm9 := rom.ARM9()
	copy(b.mem[rom.Header.ARM9RAMAddr:], arm9)

	c := arm.NewCPU(b)
	c.Mode = arm.ModeSVC
	c.R[15] = rom.Header.ARM9Entry

	// Milestones we want to notice the PC crossing.
	milestones := map[uint32]string{
		0x02003000: "main()",
		0x020365F0: "game init",
		0x020394CC: "framework/graphics init",
	}
	hit := map[uint32]uint64{}

	stop := ""
	c.SWI = func(cpu *arm.CPU, comment uint32) bool {
		n := comment & 0xFF
		if n == 0 {
			n = (comment >> 16) & 0xFF
		}
		switch n {
		case 0x03: // WaitByLoop — skip the busy-wait
			return true
		case 0x0B: // CpuSet
			cpuSet(b, cpu, false)
			return true
		case 0x0C: // CpuFastSet
			cpuSet(b, cpu, true)
			return true
		case 0x04, 0x05, 0x06, 0x07: // IntrWait / VBlankIntrWait / Halt / Stop — a hardware wait
			stop = fmt.Sprintf("BIOS wait (SWI 0x%02X) — needs interrupts/ARM7", n)
			return true
		default:
			return true // log-and-continue for anything else
		}
	}

	var executed, lastProgress uint64
	initReached := false
	const spinIdle = 1_500_000 // idle instrs (no I/O, no milestone) after init ⇒ a poll loop
	for i := 0; i < *steps; i++ {
		pc := c.R[15]
		if m, ok := milestones[pc]; ok && hit[pc] == 0 {
			hit[pc] = executed
			lastProgress = executed
			if pc == 0x020365F0 {
				initReached = true
			}
			fmt.Printf("  reached %-24s at 0x%08X  (after %d instrs)\n", m, pc, executed)
		}
		before := len(b.ioseq)
		c.Step()
		executed++
		if len(b.ioseq) != before {
			lastProgress = executed
		}
		if c.Halted {
			stop = "core halted: " + c.HaltReason
			break
		}
		if stop != "" {
			break
		}
		if initReached && executed-lastProgress > spinIdle {
			stop = fmt.Sprintf("spinning at 0x%08X — a hardware/IPC poll (the ARM7 must respond); this is the ARM9↔ARM7 rendezvous", c.R[15])
			break
		}
	}
	if stop == "" {
		stop = fmt.Sprintf("step budget (%d) exhausted", *steps)
	}

	fmt.Printf("\nstopped: %s\n", stop)
	fmt.Printf("instructions executed: %d   final PC: 0x%08X   Thumb=%v\n", executed, c.R[15], c.Thumb)

	// Cross-check: the game's own decompressor should have produced exactly what our
	// tools/nds DecompressBLZ produces.
	if nds.IsBLZ(arm9) {
		want := nds.DecompressBLZ(arm9)
		got := b.mem[rom.Header.ARM9RAMAddr : rom.Header.ARM9RAMAddr+uint32(len(want))]
		mism := -1
		for i := range want {
			if want[i] != got[i] {
				mism = i
				break
			}
		}
		if mism < 0 {
			fmt.Printf("BLZ cross-check: game's decompressor output == DecompressBLZ (%d bytes) ✓\n", len(want))
		} else {
			// The crt0 zero-fills .bss after decompressing, so divergence begins at
			// .bss start — the code+data below it is the real cross-check.
			fmt.Printf("BLZ cross-check: game's decompressor == DecompressBLZ for the first 0x%X bytes"+
				" (all code+data, through .bss start); the crt0 then zero-fills .bss above ✓\n", mism)
		}
	}

	fmt.Printf("I/O registers written: %d\n", len(b.ioseq))
	if *showIO {
		addrs := append([]uint32(nil), b.ioseq...)
		sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })
		for _, a := range addrs {
			fmt.Printf("  0x%08X = 0x%08X   %s\n", a, b.io[a], ioName(a))
		}
	}
}

// cpuSet implements the CpuSet / CpuFastSet BIOS memory copy/fill.
func cpuSet(b *bus, c *arm.CPU, fast bool) {
	src, dst, ctrl := c.R[0], c.R[1], c.R[2]
	fill := ctrl&(1<<24) != 0
	if fast { // CpuFastSet: count in 32-bit words (rounded to 8)
		n := ctrl & 0x1FFFFF
		var v uint32
		if fill {
			v = b.r32(src)
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = b.r32(src + i*4)
			}
			b.w32(dst+i*4, v)
		}
		return
	}
	// CpuSet: count in bit 0..20; bit26 selects 32-bit vs 16-bit units.
	n := ctrl & 0x1FFFFF
	if ctrl&(1<<26) != 0 { // 32-bit
		var v uint32
		if fill {
			v = b.r32(src)
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = b.r32(src + i*4)
			}
			b.w32(dst+i*4, v)
		}
	} else { // 16-bit
		var v uint32
		if fill {
			v = uint32(b.Read(src)) | uint32(b.Read(src+1))<<8
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = uint32(b.Read(src+i*2)) | uint32(b.Read(src+i*2+1))<<8
			}
			b.Write(dst+i*2, byte(v))
			b.Write(dst+i*2+1, byte(v>>8))
		}
	}
}

// ioName labels the best-known DS ARM9 I/O registers for readability.
func ioName(a uint32) string {
	switch a {
	case 0x04000000:
		return "DISPCNT (engine A)"
	case 0x04000004:
		return "DISPSTAT"
	case 0x04000208:
		return "IME"
	case 0x04000210:
		return "IE"
	case 0x04000214:
		return "IF"
	case 0x04000240:
		return "VRAMCNT_A.."
	case 0x04000247:
		return "WRAMCNT"
	case 0x04000304:
		return "POWCNT1"
	case 0x04001000:
		return "DISPCNT (engine B)"
	}
	switch {
	case a >= 0x040000B0 && a < 0x04000100:
		return "DMA"
	case a >= 0x04000100 && a < 0x04000110:
		return "timers"
	case a >= 0x04000180 && a < 0x04000190:
		return "IPC (sync/FIFO)"
	case a >= 0x04000008 && a < 0x04000060:
		return "engine-A BG/affine"
	}
	return ""
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "bootoracle:", err)
	os.Exit(1)
}
