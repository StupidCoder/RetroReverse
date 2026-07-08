// Command bootoracle loads UW.EXE into the tools/x86 execution core and runs
// it as a DOS program — the Ultima Underworld oracle. It reports how far the
// boot gets: where it halts (program exit, an unimplemented opcode, a spin, or
// the step cap), the INT 21h services it used, the files it opened, and the
// final register/segment state. Use it to follow the C-runtime startup through
// the indirect handoff into the game and toward the overlay manager (Part III).
//
// Usage:
//
//	go run ./cmd/bootoracle [-game ../game] [-steps N] [-trace] [-log]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/games/ultima-underworld-pc/extract/tex"

	"retroreverse.com/tools/platform/dos"
	"retroreverse.com/tools/cpu/x86"
)

func main() {
	game := flag.String("game", "../game", "path to the game/ folder (contains UW.EXE)")
	image := flag.String("image", "", "alias for -game (the installed game/ directory)")
	steps := flag.Uint64("steps", 20_000_000, "maximum instructions to execute")
	trace := flag.Bool("trace", false, "print a PC trace line for the first -tracen instructions")
	tracen := flag.Uint64("tracen", 200, "number of instructions to trace with -trace")
	showLog := flag.Bool("log", false, "print the full DOS event log")
	dump := flag.String("dump", "", "after the run, hex-dump SEG:OFF:LEN (hex), e.g. 5C4B:0040:0020")
	dis := flag.String("dis", "", "after the run, disassemble SEG:OFF:LEN (hex) from live (relocated) memory")
	irq := flag.Bool("irq", false, "inject periodic timer IRQ0 (recommended: drives the frame waits, cutscenes and menus)")
	bp := flag.String("bp", "", "halt when execution reaches SEG:OFF (hex), e.g. 0FD5:010D")
	bpal := flag.Int("bpal", -1, "with -bp, only halt when AL equals this value (decimal; -1 = any)")
	watch := flag.String("watch", "", "log writes to SEG:OFF[:LEN] (hex)")
	shot := flag.String("shot", "", "after the run, write the mode-13h screen (A000 + DAC palette) as PNG")
	texid := flag.Int("texid", -1, "extract global texture id N from the EMS cache (resolves [499D:B10F] -> handle -> EMS page) to -texout")
	texout := flag.String("texout", "/tmp/texid.png", "output PNG for -texid")
	keys := flag.String("keys", "", "script keyboard input via IRQ1, e.g. \"down,down,enter,wait:40,enter\" (implies -irq)")
	vgaprof := flag.Uint64("vgaprof", 0, "after this many instructions, tally code addrs writing to the profiled range; print the hottest on stop")
	profrange := flag.String("profrange", "", "with -vgaprof, profile writes to SEG:OFF:LEN (hex) instead of the A000 framebuffer")
	rdprof := flag.Uint64("rdprof", 0, "after this many instructions, tally code addrs READING -rdrange; print the hottest on stop")
	rdrange := flag.String("rdrange", "", "with -rdprof, the SEG:OFF:LEN (hex) range to profile reads of")
	loadSave := flag.Bool("loadsave", false, "pre-seed SAVE0 with the shipped save templates (test booting into a loaded game)")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before running (see -savestate)")
	saveState := flag.String("savestate", "", "after the run, dump the full machine snapshot to this file")
	poke := flag.String("poke", "", "after loadstate, write word(s): SEG:OFF:VAL[,SEG:OFF:VAL...] (hex)")
	find := flag.String("find", "", "after the run, scan all RAM for this hex byte pattern and print linear addresses")
	flag.Parse()
	if *image != "" {
		*game = *image // -image is the standard alias for -game
	}

	var bpSeg, bpOff uint32
	bpSet := false
	if *bp != "" {
		fmt.Sscanf(*bp, "%x:%x", &bpSeg, &bpOff)
		bpSet = true
	}

	exe := filepath.Join(*game, "UW.EXE")
	m, err := dos.LoadEXE(exe, *game)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
	m.SeedDir("SAVE0") // UW aborts without a SAVE0 working dir (empty on first run)
	if *loadSave {
		// Pre-place the shipped save templates into SAVE0 so the game may boot into
		// a loaded game (Journey Onward) rather than character creation.
		for src, dst := range map[string]string{
			"PLAYER.DAT": `SAVE0\PLAYER.DAT`, "BABGLOBS.DAT": `SAVE0\BGLOBALS.DAT`,
			"LEV.ARK": `SAVE0\LEV.ARK`,
		} {
			b, err := os.ReadFile(filepath.Join(*game, "DATA", src))
			if err != nil {
				fmt.Fprintln(os.Stderr, "bootoracle: -loadsave:", err)
				os.Exit(1)
			}
			if err := m.SeedSaveFile(dst, b); err != nil {
				fmt.Fprintln(os.Stderr, "bootoracle: -loadsave:", err)
				os.Exit(1)
			}
		}
	}
	m.EnableIRQ = *irq
	m.VGAProfileAt = *vgaprof
	if *profrange != "" {
		var s, o, l uint32
		fmt.Sscanf(*profrange, "%x:%x:%x", &s, &o, &l)
		m.ProfLo = (s<<4 + o) & 0xFFFFF
		m.ProfHi = m.ProfLo + l
	}
	m.RdProfileAt = *rdprof
	if *rdrange != "" {
		var s, o, l uint32
		fmt.Sscanf(*rdrange, "%x:%x:%x", &s, &o, &l)
		m.RdLo = (s<<4 + o) & 0xFFFFF
		m.RdHi = m.RdLo + l
	}
	if *loadState != "" {
		if err := m.LoadState(*loadState); err != nil {
			fmt.Fprintln(os.Stderr, "bootoracle: -loadstate:", err)
			os.Exit(1)
		}
		m.EnableIRQ = *irq // hooks were reset by the fresh load; re-apply run options
		fmt.Printf("restored snapshot %s (at %d instructions)\n", *loadState, m.CPU.Steps)
	}
	if *poke != "" {
		for _, spec := range strings.Split(*poke, ",") {
			var s, o, v uint32
			if n, _ := fmt.Sscanf(spec, "%x:%x:%x", &s, &o, &v); n < 3 {
				fmt.Fprintln(os.Stderr, "bootoracle: -poke bad spec:", spec)
				os.Exit(1)
			}
			lin := (s<<4 + o) & 0xFFFFF
			m.Mem[lin] = byte(v)
			m.Mem[(lin+1)&0xFFFFF] = byte(v >> 8)
			fmt.Printf("poke %04X:%04X <- %04X\n", s, o, v)
		}
	}
	if *keys != "" {
		ev, err := dos.ParseKeys(*keys)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bootoracle:", err)
			os.Exit(1)
		}
		m.SetKeys(ev)
		m.EnableIRQ = true // key pacing rides the timer tick
	}
	if *watch != "" {
		var s, o, l uint32
		if n, _ := fmt.Sscanf(*watch, "%x:%x:%x", &s, &o, &l); n < 3 {
			l = 1
		}
		m.WatchAddr = (s<<4 + o) & 0xFFFFF
		m.WatchLen = l
	}
	c := m.CPU
	fmt.Printf("loaded %s: entry CS:IP=%04X:%04X SS:SP=%04X:%04X DS=%04X (irq=%v)\n\n",
		exe, c.Seg[x86.CS], c.IP, c.Seg[x86.SS], c.Reg16(x86.SP), c.Seg[x86.DS], *irq)

	// Optional short instruction trace at the entry.
	if *trace {
		var n uint64
		c.OnStep = func(cpu *x86.CPU) {
			if n < *tracen {
				lin := (uint32(cpu.Seg[x86.CS]) << 4) + cpu.IP
				in := x86.Decode(m.Mem[lin&0xFFFFF:], cpu.IP)
				fmt.Printf("%04X:%04X  %s\n", cpu.Seg[x86.CS], cpu.IP, in.Text)
			}
			n++
		}
	}

	// Spin detection + a ring buffer of the last executed PCs, so a step-cap or
	// spin stop can show exactly what the CPU was doing.
	spin := newSpinDetector(c)
	const ringSize = 400
	ring := make([]uint32, ringSize) // packed CS<<16 | IP
	ri := 0
	zeroRun := 0
	prevHook := c.OnStep
	c.OnStep = func(cpu *x86.CPU) {
		if prevHook != nil {
			prevHook(cpu)
		}
		ring[ri%ringSize] = uint32(cpu.Seg[x86.CS])<<16 | (cpu.IP & 0xFFFF)
		ri++
		if bpSet && uint32(cpu.Seg[x86.CS]) == bpSeg && (cpu.IP&0xFFFF) == bpOff &&
			(*bpal < 0 || cpu.Reg8(x86.AL) == byte(*bpal)) {
			cpu.Halt("breakpoint at %04X:%04X (AX=%04X BX=%04X CX=%04X DX=%04X SI=%04X DI=%04X BP=%04X SP=%04X DS=%04X ES=%04X SS=%04X)",
				cpu.Seg[x86.CS], cpu.IP, cpu.Reg16(x86.AX), cpu.Reg16(x86.BX), cpu.Reg16(x86.CX), cpu.Reg16(x86.DX),
				cpu.Reg16(x86.SI), cpu.Reg16(x86.DI), cpu.Reg16(x86.BP), cpu.Reg16(x86.SP),
				cpu.Seg[x86.DS], cpu.Seg[x86.ES], cpu.Seg[x86.SS])
		}
		// Runaway detector: executing a run of 00 00 bytes means the CPU jumped
		// into zero-filled memory — halt so the ring shows the culprit transfer.
		l := (uint32(cpu.Seg[x86.CS]) << 4) + (cpu.IP & 0xFFFF)
		if m.Mem[l&0xFFFFF] == 0 && m.Mem[(l+1)&0xFFFFF] == 0 {
			zeroRun++
			if zeroRun > 6 {
				cpu.Halt("runaway: executing zero-filled memory at %04X:%04X", cpu.Seg[x86.CS], cpu.IP)
			}
		} else {
			zeroRun = 0
		}
		spin.check(cpu)
	}

	c.Run(*steps)

	if *find != "" {
		var pat []byte
		for _, h := range strings.Fields(strings.ReplaceAll(*find, ",", " ")) {
			var v uint32
			fmt.Sscanf(h, "%x", &v)
			pat = append(pat, byte(v))
		}
		hits := 0
		scan := func(name string, mem []byte) {
			for i := 0; i+len(pat) <= len(mem); i++ {
				if mem[i] == pat[0] && string(mem[i:i+len(pat)]) == string(pat) {
					fmt.Printf("found in %s at %05X\n", name, i)
					if hits++; hits >= 20 {
						return
					}
				}
			}
		}
		scan("RAM", m.Mem[:])
		scan("EMS", m.EmsBacking())
		if hits == 0 {
			fmt.Println("pattern not found")
		}
	}

	if *texid >= 0 {
		extractTexID(m, *texid, *texout, *game)
	}

	if *saveState != "" {
		if err := m.SaveState(*saveState); err != nil {
			fmt.Fprintln(os.Stderr, "bootoracle: -savestate:", err)
		} else {
			fmt.Printf("\nsnapshot written to %s (%d instructions)\n", *saveState, m.CPU.Steps)
		}
	}

	if len(m.Int33Hist) > 0 {
		total := 0
		fmt.Printf("\n== INT33 mouse polling ==\n")
		for _, ah := range []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 0xB, 0xC, 0xF, 0x14} {
			if c := m.Int33Hist[ah]; c > 0 {
				fmt.Printf("  AH=%02X: %d\n", ah, c)
				total += c
			}
		}
		if s := m.CPU.Steps; s > 0 {
			fmt.Printf("  ~1 poll per %d instructions\n", s/uint64(total+1))
		}
	}

	// Print the tail of the execution ring (deduplicated consecutive repeats).
	fmt.Printf("\n== last instructions ==\n")
	start := ri - ringSize
	if start < 0 {
		start = 0
	}
	var lastLine string
	for i := start; i < ri; i++ {
		v := ring[i%ringSize]
		cs, ip := uint16(v>>16), v&0xFFFF
		l := (uint32(cs) << 4) + ip
		in := x86.Decode(m.Mem[l&0xFFFFF:], ip)
		line := fmt.Sprintf("%04X:%04X  %s", cs, ip, in.Text)
		if line != lastLine {
			fmt.Println("  " + line)
			lastLine = line
		}
	}

	fmt.Printf("\n== stopped after %d instructions (%d were 386 0F/66/67 ops) ==\n", c.Steps, c.Ext386)
	switch {
	case m.Terminated:
		fmt.Printf("reason: program terminated (exit code %d)\n", m.ExitCode)
	case c.Halted:
		fmt.Printf("reason: %s\n", c.HaltReason)
	default:
		fmt.Printf("reason: step cap reached (still running at %04X:%04X)\n", c.Seg[x86.CS], c.IP)
	}

	fmt.Printf("\nfinal state:\n")
	fmt.Printf("  CS:IP=%04X:%04X  SS:SP=%04X:%04X  DS=%04X ES=%04X\n",
		c.Seg[x86.CS], c.IP, c.Seg[x86.SS], c.Reg16(x86.SP), c.Seg[x86.DS], c.Seg[x86.ES])
	fmt.Printf("  AX=%04X BX=%04X CX=%04X DX=%04X SI=%04X DI=%04X BP=%04X\n",
		c.Reg16(x86.AX), c.Reg16(x86.BX), c.Reg16(x86.CX), c.Reg16(x86.DX),
		c.Reg16(x86.SI), c.Reg16(x86.DI), c.Reg16(x86.BP))

	fmt.Printf("\nINT 21h services used: ")
	if s := m.IntSummary(); len(s) > 0 {
		for _, e := range s {
			fmt.Printf("%s  ", e)
		}
	}
	fmt.Println()

	if *vgaprof > 0 {
		prof := m.VGAProfile()
		fmt.Printf("\n== A000 writers since step %d (hottest 20 of %d) ==\n", *vgaprof, len(prof))
		for i, w := range prof {
			if i >= 20 {
				break
			}
			fmt.Printf("  %04X:%04X  %d writes\n", w.Seg, w.Off, w.Count)
		}
	}

	if *rdprof > 0 {
		prof := m.ReadProfile()
		fmt.Printf("\n== readers of the range since step %d (hottest 20 of %d) ==\n", *rdprof, len(prof))
		for i, w := range prof {
			if i >= 20 {
				break
			}
			fmt.Printf("  %04X:%04X  %d reads\n", w.Seg, w.Off, w.Count)
		}
	}

	if *shot != "" {
		if err := m.Screenshot(*shot); err != nil {
			fmt.Fprintln(os.Stderr, "screenshot:", err)
		} else {
			fmt.Printf("\nscreenshot written to %s\n", *shot)
		}
	}

	if *dump != "" {
		var seg, off, ln uint32
		if _, err := fmt.Sscanf(*dump, "%x:%x:%x", &seg, &off, &ln); err == nil {
			base := (seg<<4 + off) & 0xFFFFF
			fmt.Printf("\n== dump %04X:%04X len %X ==\n", seg, off, ln)
			for i := uint32(0); i < ln; i += 16 {
				fmt.Printf("%04X:%04X ", seg, off+i)
				for j := uint32(0); j < 16 && i+j < ln; j++ {
					fmt.Printf("%02X ", m.Mem[(base+i+j)&0xFFFFF])
				}
				fmt.Println()
			}
		}
	}

	if *dis != "" {
		var seg, off, ln uint32
		if _, err := fmt.Sscanf(*dis, "%x:%x:%x", &seg, &off, &ln); err == nil {
			base := (seg<<4 + off) & 0xFFFFF
			fmt.Printf("\n== disasm %04X:%04X len %X (live memory) ==\n", seg, off, ln)
			end := base + ln
			if end > uint32(len(m.Mem)) {
				end = uint32(len(m.Mem))
			}
			for _, l := range x86.Disassemble(m.Mem[base:end], off) {
				fmt.Println("  " + l)
			}
		}
	}

	if *showLog {
		fmt.Printf("\n== DOS event log (%d) ==\n", len(m.Log))
		for _, l := range m.Log {
			fmt.Println(l)
		}
	} else if len(m.Log) > 0 {
		fmt.Printf("\nlast events:\n")
		start := 0
		if len(m.Log) > 15 {
			start = len(m.Log) - 15
		}
		for _, l := range m.Log[start:] {
			fmt.Println("  " + l)
		}
	}
}

// spinDetector stops the CPU when the PC revisits the same address too many
// times without the surrounding window advancing — a stand-in for hardware
// polling loops the oracle can't satisfy.
type spinDetector struct {
	cpu       *x86.CPU
	lastPC    uint32
	sameCount int
}

func newSpinDetector(c *x86.CPU) *spinDetector { return &spinDetector{cpu: c} }

func (s *spinDetector) check(c *x86.CPU) {
	pc := (uint32(c.Seg[x86.CS]) << 4) + c.IP
	if pc == s.lastPC {
		s.sameCount++
		if s.sameCount > 5_000_000 {
			c.Halt("spin detected at %04X:%04X (%d repeats)", c.Seg[x86.CS], c.IP, s.sameCount)
		}
		return
	}
	s.lastPC = pc
	s.sameCount = 0
}

// extractTexID pulls a texture out of the game's EMS texture cache by its global
// id, exactly as the 3D renderer's 0xC0 opcode does (07F7:79D2): read the handle
// from [499D:B10F + id*2], then the texture lives at logical page (handle>>12) of
// the texture-cache EMS handle ([499D:C4D7]), byte offset (handle & 0x3FF)*16
// within that page. The cached bitmap keeps its .GR frame layout
// [format][W][H][u16 size][pixels]. This lets us confirm which archive frame a
// model's texture id resolves to without reverse-engineering the load-time id
// assignment or the page mapping.
func extractTexID(m *dos.Machine, id int, out, game string) {
	ds := uint32(0x499D) << 4
	rd16 := func(off uint32) uint16 {
		return uint16(m.Mem[ds+off]) | uint16(m.Mem[ds+off+1])<<8
	}
	handle := rd16(0xB10F + uint32(id)*2)
	emsH := uint16(m.Mem[ds+0xC4D7])
	bank := int(handle >> 12)
	base := m.EmsHandleBase(emsH)
	off := int(handle&0x3FF) * 16
	fmt.Printf("texid %d: handle=%04X bank(logical page)=%d emsHandle=%d base=%d off=0x%X\n", id, handle, bank, emsH, base, off)
	if base < 0 {
		fmt.Println("  no such EMS handle")
		return
	}
	backing := m.EmsBacking()
	lin := (base+bank)*dos.EmsPageSize + off
	if lin < 0 || lin+5 > len(backing) {
		fmt.Println("  resolved location out of range")
		return
	}
	t := backing[lin:]
	format, w, h := t[0], int(t[1]), int(t[2])
	size := int(t[3]) | int(t[4])<<8
	fmt.Printf("  format=0x%02X W=%d H=%d size=%d\n", format, w, h, size)
	if w == 0 || h == 0 || w > 256 || h > 256 {
		fmt.Println("  header not a texture (id likely not cached) — dumping first bytes:")
		fmt.Printf("  % X\n", t[:16])
		return
	}
	palsB, _ := os.ReadFile(filepath.Join(game, "DATA/PALS.DAT"))
	pal, _ := tex.LoadPalette(palsB, 0)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	px := t[5:]
	for i := 0; i < w*h && i < len(px); i++ {
		c := pal[px[i]]
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = c.R, c.G, c.B, 255
	}
	// scale x6 for viewing
	big := image.NewRGBA(image.Rect(0, 0, w*6, h*6))
	for y := 0; y < h*6; y++ {
		for x := 0; x < w*6; x++ {
			big.Set(x, y, img.At(x/6, y/6))
		}
	}
	f, err := os.Create(out)
	if err != nil {
		fmt.Println("  create:", err)
		return
	}
	png.Encode(f, big)
	f.Close()
	fmt.Printf("  wrote %s (%dx%d, format %02X)\n", out, w, h, format)

	// Byte-match the cached pixels against TMOBJ.GR frames to name the exact frame
	// (definitive: identical palette-index bytes), so the model->frame formula can
	// be confirmed rather than guessed.
	want := px[:w*h]
	grB, err := os.ReadFile(filepath.Join(game, "DATA/TMOBJ.GR"))
	if err != nil {
		return
	}
	// .GR: byte0 tag, u16 count, count u32 offsets, then frames [fmt][W][H][u16 size][pixels]
	if len(grB) < 3 {
		return
	}
	cnt := int(grB[1]) | int(grB[2])<<8
	for i := 0; i < cnt; i++ {
		o := int(grB[3+i*4]) | int(grB[4+i*4])<<8 | int(grB[5+i*4])<<16 | int(grB[6+i*4])<<24
		if o == 0 || o+5 > len(grB) {
			continue
		}
		fw, fh := int(grB[o+1]), int(grB[o+2])
		if fw != w || fh != h || grB[o] != format {
			continue
		}
		fp := grB[o+5:]
		if len(fp) < w*h {
			continue
		}
		match := true
		for k := 0; k < w*h; k++ {
			if fp[k] != want[k] {
				match = false
				break
			}
		}
		if match {
			fmt.Printf("  == TMOBJ.GR frame %d (byte-identical)\n", i)
		}
	}
}
