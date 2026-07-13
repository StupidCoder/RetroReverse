// bootoracle boots Super Mario 3D Land's ARM11 application code on the shared
// 3DS machine (tools/platform/n3ds) and runs it, exposing the standard oracle
// instrumentation (STANDARDS §3). It is the executable counterpart of the static
// tools: where n3dsdump reads the cartridge's containers, this one runs the code
// they hold.
//
// The machine is a first-milestone process-level model: it loads the code
// segments, lays out the userland memory map, and high-level-emulates the Horizon
// supervisor calls needed to run the C runtime. It stops explicitly at the first
// kernel facility not yet implemented (typically the srv:/GSP IPC handshake), so a
// run reports concretely how far the bring-up reaches rather than diverging
// silently. See the game writeup for the current frontier.
//
// Usage:
//
//	bootoracle -image game.cci [-steps N] [-trace] [-tracen N] [-bp A]... [-watch A[:L]]...
//	           [-svclog] [-savestate F] [-loadstate F]
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

// parseNum accepts hex (0x…) or decimal, matching the other oracles.
func parseNum(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func main() {
	image := flag.String("image", "", "3DS cartridge image (decrypted .cci/.3ds)")
	steps := flag.String("steps", "2000000", "instruction budget (hex with 0x, else decimal)")
	trace := flag.Bool("trace", false, "trace executed instructions")
	tracen := flag.Int("tracen", 200, "limit -trace to this many instructions")
	verbose := flag.Bool("v", false, "log every supervisor call and unmapped access as it happens")
	svclog := flag.Bool("svclog", false, "print the ordered supervisor-call log after the run")
	var bps, watches, logpcs multiFlag
	flag.Var(&bps, "bp", "breakpoint address (hex); repeatable")
	flag.Var(&logpcs, "logpc", "log register context at this address and continue (hex); repeatable")
	var tracefroms multiFlag
	flag.Var(&tracefroms, "tracefrom", "start instruction tracing (for -tracen instrs) when this address is first reached (hex); repeatable")
	flag.Var(&watches, "watch", "memory watch ADDR[:LEN] (hex); repeatable")
	saveState := flag.String("savestate", "", "after the run, dump the machine snapshot to this file")
	loadState := flag.String("loadstate", "", "restore a machine snapshot before running")
	gxdump := flag.String("gxdump", "", "capture GX commands; write ProcessCommandList buffers to this directory")
	shot := flag.String("shot", "", "after the run, write the presented framebuffers to <base>_top.png / <base>_bottom.png")
	gputrace := flag.Int("gputrace", 0, "print a summary of the first N GPU draws")
	rtshot := flag.String("rtshot", "", "after the run, decode a tiled RGBA8 render target straight from memory: ADDR:WxH[:FILE] (hex addr) — what the GPU drew, before any DisplayTransfer")
	threads := flag.Bool("threads", false, "after the run, dump thread states and the handle table")
	hidtrace := flag.Bool("hidtrace", false, "tally reads of the HID shared-memory block by offset, dump after the run")
	findAscii := flag.String("findascii", "", "after load/run, print addresses where this ASCII string occurs in memory")
	findUtf16 := flag.String("findutf16", "", "after load/run, print addresses where this UTF-16LE string occurs in memory")
	findWord := flag.String("findword", "", "after load/run, print addresses where this 32-bit little-endian word occurs in memory (hex)")
	var dumps multiFlag
	flag.Var(&dumps, "dump", "hex-dump ADDR:LEN of memory after load/run (hex); repeatable")
	var pokes multiFlag
	flag.Var(&pokes, "poke", "write a 32-bit word ADDR:VALUE (hex) after -loadstate, before running; repeatable (probe instrument)")
	keys := flag.String("keys", "", "inject HID pad input: comma-separated button names (a,b,x,y,l,r,up,down,left,right,start,select)")
	keypulse := flag.Int("keypulse", 0, "if >0, release the injected keys briefly every N frames so a fresh press edge keeps arriving (for advancing menus/dialogs)")
	wav := flag.String("wav", "", "capture the DSP's final mix during the run and write it to this .wav file")
	dsptrace := flag.Bool("dsptrace", false, "log every source configuration the DSP consumes and every status it publishes")
	flag.Parse()

	if *image == "" {
		fmt.Fprintln(os.Stderr, "bootoracle: -image is required")
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*image, *steps, *trace, *tracen, *verbose, *svclog, bps, watches, logpcs, tracefroms, dumps, *saveState, *loadState, *gxdump, *shot, *rtshot, *gputrace, *threads, *hidtrace, *keys, *keypulse, *findAscii, *findUtf16, *findWord, pokes, *wav, *dsptrace); err != nil {
		fmt.Fprintln(os.Stderr, "bootoracle:", err)
		os.Exit(1)
	}
}

func asciiPattern(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}

func wordPattern(s string) []byte {
	if s == "" {
		return nil
	}
	v, err := parseNum(s)
	if err != nil {
		return nil
	}
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

func utf16Pattern(s string) []byte {
	if s == "" {
		return nil
	}
	b := make([]byte, 0, len(s)*2)
	for _, r := range s {
		b = append(b, byte(r), byte(r>>8))
	}
	return b
}

func run(imagePath, stepsStr string, trace bool, tracen int, verbose, svclog bool, bps, watches, logpcs, tracefroms, dumps multiFlag, saveState, loadState, gxdump, shot, rtshot string, gputrace int, threads, hidtrace bool, keys string, keypulse int, findAscii, findUtf16, findWord string, pokes multiFlag, wav string, dsptrace bool) error {
	img, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}
	m, err := n3ds.NewMachine(img)
	if err != nil {
		return err
	}
	m.Verbose = verbose
	m.SetTrace(trace, tracen)
	m.GXCapture = gxdump != ""
	m.GPU().TraceDraws = gputrace
	m.HidTrace = hidtrace
	m.AudioCapture = wav != ""
	m.DSPTrace = dsptrace
	m.HidPulse = keypulse
	if keys != "" {
		if err := m.SetKeys(keys); err != nil {
			return err
		}
	}

	for _, b := range bps {
		v, err := parseNum(b)
		if err != nil {
			return fmt.Errorf("bad -bp %q: %w", b, err)
		}
		m.AddBreakpoint(uint32(v))
	}
	for _, b := range logpcs {
		v, err := parseNum(b)
		if err != nil {
			return fmt.Errorf("bad -logpc %q: %w", b, err)
		}
		m.AddLogPC(uint32(v))
	}
	for _, b := range tracefroms {
		v, err := parseNum(b)
		if err != nil {
			return fmt.Errorf("bad -tracefrom %q: %w", b, err)
		}
		m.AddTraceFrom(uint32(v))
	}
	for _, w := range watches {
		addr, length := w, "4"
		if i := strings.IndexByte(w, ':'); i >= 0 {
			addr, length = w[:i], w[i+1:]
		}
		a, err := parseNum(addr)
		if err != nil {
			return fmt.Errorf("bad -watch %q: %w", w, err)
		}
		l, err := parseNum(length)
		if err != nil {
			return fmt.Errorf("bad -watch length %q: %w", w, err)
		}
		m.AddWatch(uint32(a), uint32(l))
	}

	if loadState != "" {
		if err := m.LoadState(loadState); err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
		fmt.Printf("restored snapshot from %s (at %d instructions)\n", loadState, m.Instrs())
	}

	for _, pk := range pokes {
		i := strings.IndexByte(pk, ':')
		if i < 0 {
			return fmt.Errorf("bad -poke %q: want ADDR:VALUE", pk)
		}
		a, err := parseNum(pk[:i])
		if err != nil {
			return fmt.Errorf("bad -poke addr %q: %w", pk, err)
		}
		v, err := parseNum(pk[i+1:])
		if err != nil {
			return fmt.Errorf("bad -poke value %q: %w", pk, err)
		}
		m.WriteWord(uint32(a), uint32(v))
		fmt.Printf("poke [0x%08X] = 0x%08X\n", a, v)
	}

	budget, err := parseNum(stepsStr)
	if err != nil {
		return fmt.Errorf("bad -steps: %w", err)
	}

	fmt.Printf("boot: entry 0x%08X, running up to %d instructions\n", m.Entry(), budget)
	ran := m.Run(int(budget))

	fmt.Printf("\nran %d instructions (%d total)\n", ran, m.Instrs())
	if r := m.HaltReason(); r != "" {
		fmt.Printf("halt: %s\n", r)
	} else {
		fmt.Printf("still runnable (budget reached)\n")
	}
	if dbg := m.DebugString(); dbg != "" {
		fmt.Printf("\nOutputDebugString:\n%s\n", dbg)
	}
	if ports := m.Ports(); len(ports) > 0 {
		fmt.Printf("\nservice ports connected:\n")
		for h, name := range ports {
			fmt.Printf("  0x%08X  %s\n", h, name)
		}
	}
	if ipc := m.IPCLog(); len(ipc) > 0 {
		fmt.Printf("\nIPC requests: %d\n", len(ipc))
		counts := map[string]int{}
		for _, e := range ipc {
			counts[e.Service()]++
		}
		for svc, n := range counts {
			fmt.Printf("  %-10s %d requests\n", svc, n)
		}
	}
	if vb := m.VBlanks(); vb > 0 {
		sub, swp := m.FrameStats()
		fmt.Printf("\ngraphics: %d VBlanks delivered, %d GPU command lists submitted, %d display transfers, %d frame swaps\n",
			vb, sub, m.DisplayTransfers(), swp)
		g := m.GPU()
		fmt.Printf("gpu: %d draws, %d pixels drawn; tris: %d zero-area, %d culled, %d w-rejected; %d depth-killed frags\n",
			g.Draws, g.PixelsDrawn, g.ZeroAreaTris, g.CulledTris, g.RejectedTris, g.DepthKilled)
		fmt.Printf("shadow: %d map writes, %d samples (%d occluded)\n",
			g.ShadowWrites, g.ShadowSamples, g.ShadowOccluded)
	}
	if svclog {
		printSVCSummary(m)
	}
	if threads {
		m.DumpThreads()
	}
	if hidtrace {
		m.DumpHIDReads()
	}
	for _, d := range dumps {
		addr, length := d, "64"
		if i := strings.IndexByte(d, ':'); i >= 0 {
			addr, length = d[:i], d[i+1:]
		}
		a, err := parseNum(addr)
		if err != nil {
			return fmt.Errorf("bad -dump %q: %w", d, err)
		}
		l, err := parseNum(length)
		if err != nil {
			return fmt.Errorf("bad -dump length %q: %w", d, err)
		}
		fmt.Printf("\ndump 0x%08X (%d bytes):\n", a, l)
		bytes := m.ReadBytes(uint32(a), uint32(l))
		for i := 0; i < len(bytes); i += 16 {
			end := i + 16
			if end > len(bytes) {
				end = len(bytes)
			}
			fmt.Printf("  0x%08X ", uint32(a)+uint32(i))
			for j := i; j < end; j++ {
				fmt.Printf("%02X ", bytes[j])
			}
			fmt.Print(" |")
			for j := i; j < end; j++ {
				c := bytes[j]
				if c < 0x20 || c > 0x7E {
					c = '.'
				}
				fmt.Printf("%c", c)
			}
			fmt.Println("|")
		}
	}
	for _, spec := range []struct {
		label string
		pat   []byte
	}{{"ASCII " + findAscii, asciiPattern(findAscii)}, {"UTF-16 " + findUtf16, utf16Pattern(findUtf16)}, {"word " + findWord, wordPattern(findWord)}} {
		if len(spec.pat) == 0 {
			continue
		}
		hits := m.FindBytes(spec.pat)
		fmt.Printf("\nfind %q: %d hit(s)\n", spec.label, len(hits))
		for _, h := range hits {
			fmt.Printf("  0x%08X\n", h)
		}
	}
	if gxdump != "" {
		if err := dumpGX(m, gxdump); err != nil {
			return err
		}
	}
	if wav != "" {
		fmt.Printf("dsp audio: %s\n", m.AudioSummary())
		if err := m.WriteWAV(wav); err != nil {
			fmt.Printf("wav: %v\n", err)
		} else {
			fmt.Printf("wrote %s\n", wav)
		}
	}

	if rtshot != "" {
		if err := writeRenderTarget(m, rtshot); err != nil {
			fmt.Printf("rtshot: %v\n", err)
		}
	}
	if shot != "" {
		if err := m.Screenshot(shot); err != nil {
			fmt.Printf("screenshot: %v\n", err)
		} else {
			fmt.Printf("wrote %s_top.png / %s_bottom.png\n", shot, shot)
		}
	}

	if saveState != "" {
		if err := m.SaveState(saveState); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		fmt.Printf("\nwrote snapshot to %s\n", saveState)
	}
	return nil
}

// gxNames labels the GX command ids for the capture listing.
var gxNames = map[uint32]string{
	0: "RequestDMA", 1: "ProcessCommandList", 2: "MemoryFill",
	3: "DisplayTransfer", 4: "TextureCopy", 5: "FlushCacheRegions",
}

// dumpGX prints every captured GX command's raw slot words and writes each
// ProcessCommandList's PICA200 command buffer to dir/cmdlist_NN.bin — the
// instrument-first artifact Phase 4 (the GPU) is built against.
func dumpGX(m *n3ds.Machine, dir string) error {
	log := m.GXLog()
	if len(log) == 0 {
		fmt.Println("\nGX capture: no commands seen")
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fmt.Printf("\nGX capture: %d commands\n", len(log))
	lists := 0
	for i, r := range log {
		id := r.Words[0] & 0x1F
		name := gxNames[id]
		if name == "" {
			name = fmt.Sprintf("cmd%d", id)
		}
		if r.Chained {
			// Not a submitted command: a buffer the command processor jumped into.
			path := filepath.Join(dir, fmt.Sprintf("cmdlist_%02d.bin", lists))
			if err := os.WriteFile(path, r.Buf, 0o644); err != nil {
				return err
			}
			fmt.Printf("  %3d @%-11d   ..chained  -> %s (%d bytes from 0x%08X)\n",
				i, r.Instr, path, len(r.Buf), r.Words[1])
			lists++
			continue
		}
		fmt.Printf("  %3d @%-11d %-18s %08X %08X %08X %08X %08X %08X %08X %08X\n",
			i, r.Instr, name,
			r.Words[0], r.Words[1], r.Words[2], r.Words[3],
			r.Words[4], r.Words[5], r.Words[6], r.Words[7])
		if r.Buf != nil {
			path := filepath.Join(dir, fmt.Sprintf("cmdlist_%02d.bin", lists))
			if err := os.WriteFile(path, r.Buf, 0o644); err != nil {
				return err
			}
			fmt.Printf("        -> %s (%d bytes from 0x%08X)\n", path, len(r.Buf), r.Words[1])
			lists++
		}
	}
	return nil
}

func printSVCSummary(m *n3ds.Machine) {
	log := m.SVCLog()
	fmt.Printf("\nsupervisor calls: %d total\n", len(log))
	counts := map[string]int{}
	for _, e := range log {
		counts[e.Name]++
	}
	// Print the first several in order, then the histogram.
	fmt.Println("  first calls, in order:")
	for i, e := range log {
		if i >= 40 {
			fmt.Printf("  … and %d more\n", len(log)-i)
			break
		}
		fmt.Printf("    %3d  0x%02X %-20s r0=%08X r1=%08X r2=%08X r3=%08X\n",
			i, e.Num, e.Name, e.Args[0], e.Args[1], e.Args[2], e.Args[3])
	}
	fmt.Println("  histogram:")
	for name, n := range counts {
		fmt.Printf("    %-24s %d\n", name, n)
	}
}

// writeRenderTarget decodes ADDR:WxH[:FILE] and dumps that colour buffer as a
// PNG, reading it as the PICA stores it (8x8 Morton tiles, RGBA8).
func writeRenderTarget(m *n3ds.Machine, spec string) error {
	name := "rendertarget.png"
	parts := strings.Split(spec, ":")
	if len(parts) == 3 {
		name = parts[2]
	} else if len(parts) != 2 {
		return fmt.Errorf("want ADDR:WxH[:FILE], got %q", spec)
	}
	addr, err := strconv.ParseUint(strings.TrimPrefix(parts[0], "0x"), 16, 32)
	if err != nil {
		return fmt.Errorf("bad address in %q: %v", spec, err)
	}
	var w, h uint32
	if _, err := fmt.Sscanf(parts[1], "%dx%d", &w, &h); err != nil {
		return fmt.Errorf("bad dimensions in %q: %v", spec, err)
	}
	img := m.RenderTarget(uint32(addr), w, h)
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%dx%d render target at 0x%08X)\n", name, w, h, addr)
	return nil
}
