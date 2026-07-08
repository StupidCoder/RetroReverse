// geomoracle differentially verifies the rr decoders against the running
// game. It never extracts data — the decoders read the CD file bytes; the
// oracle only checks them (the repo's decode-reimplement rule).
//
// Three checks, selected with -only (comma separated, default all):
//
//   - vram: boots the game until every TEX*.TMS archive has streamed to the
//     GPU, then compares the machine's VRAM word-for-word against the pure-Go
//     replay (rr.ParseTMS + rr.NewVRAM) over every uploaded rect.
//
//   - cars: drives the menu to car select and, for a window of frames, traps
//     the model renderer's GTE loads (the RTPT/RTPS at 0x80046B6C/0x80046BDC,
//     the NCT/NCS at 0x80046C74/0x80046CA0). Each trap yields the record's
//     RAM address (register $t1) and the GTE vertex/normal registers the game
//     just loaded; both must match the rr.ParseRRO record at that file
//     offset, bit-exact. Every trapped address must land on a decoder-
//     predicted record boundary.
//
//   - track: drives the menu into the race and traps the shared 40-byte-quad
//     transform (RTPT/RTPS at 0x80046250/0x8004626C, record in $a0) plus the
//     lit-object renderer, verifying MAP.RRM track quads and OBJ.RRO records
//     against rr.ParseRRM/rr.ParseRRO the same way.
//
// Exit status 0 only if every trapped event matches.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
	"retroreverse.com/tools/platform/psx"
)

// Traced constants (see ridge-racer-psx.md): resident file bases and the
// renderer PCs the checks trap.
const (
	mapBase = 0x0008045C // MAP.RRM in RAM (physical)
	objBase = 0x000C2918 // OBJ.RRO in RAM (physical)

	pcQuadRTPT = 0x80046250 // 40-byte quad path: RTPT (record in $a0)
	pcQuadRTPS = 0x8004626C // 40-byte quad path: 4th-vertex RTPS
	pcGTRTPT   = 0x80046B6C // 64/72-byte lit path: RTPT (record in $t1)
	pcGTRTPS   = 0x80046BDC // 64/72-byte lit path: 4th-vertex RTPS
	pcGTNCT    = 0x80046C74 // 64/72-byte lit path: NCT (normals 0-2)
	pcGTNCS    = 0x80046CA0 // 64/72-byte lit path: NCS (normal 3)

	regA0 = 4 // $a0
	regT1 = 9 // $t1

	pressCarSelect = "start@380000000:380000,up@386000000:380000,right@388000000:380000," +
		"right@390000000:380000,cross@392000000:380000"
	pressRace = "start@380000000:380000,cross@386000000:380000"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "geomoracle: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "PlayStation CD image (.bin)")
	only := flag.String("only", "all", "checks to run: vram,cars,track or all")
	flag.Parse()
	if *image == "" {
		die("need -image")
	}
	data, err := os.ReadFile(*image)
	if err != nil {
		die("%v", err)
	}
	vol, err := psx.Open(data)
	if err != nil {
		die("%v", err)
	}

	want := map[string]bool{}
	for _, s := range strings.Split(*only, ",") {
		want[strings.TrimSpace(s)] = true
	}
	all := want["all"]

	fail := 0
	if all || want["vram"] {
		fail += checkVRAM(vol)
	}
	if all || want["cars"] {
		fail += checkGeometry(vol, "cars", pressCarSelect, 396_000_000, 1_000_000)
	}
	if all || want["track"] {
		fail += checkGeometry(vol, "track", pressRace, 429_500_000, 1_000_000)
	}
	if fail > 0 {
		die("%d check(s) failed", fail)
	}
	fmt.Println("geomoracle: all checks passed")
}

// boot loads the disc's EXE into a fresh machine with Ridge Racer's traced
// interrupt dispatcher and an optional pad script.
func boot(vol *psx.Volume, press string) *psx.Machine {
	_, exe, err := vol.BootEXE()
	if err != nil {
		die("boot exe: %v", err)
	}
	m := psx.NewMachine()
	m.SetDisc(vol)
	m.ISRHandler = 0x8004DF48
	if press != "" {
		script, err := psx.ParsePress(press)
		if err != nil {
			die("press: %v", err)
		}
		m.PadScript = script
	}
	m.LoadEXE(exe)
	return m
}

// --- vram --------------------------------------------------------------

// vramFiles pairs each archive with the boot step at which its per-frame
// upload stream has finished (the next file reuses the staging buffer, so a
// file's stream always completes before the next one loads). Comparing each
// file at its own checkpoint keeps the check exact even though the title
// screen later redraws parts of the texture pages.
var vramFiles = []struct {
	name string
	step uint64
}{
	{"TEX4.TMS", 50_000_000},
	{"TEX0.TMS", 160_000_000},
	{"TEX1.TMS", 216_000_000},
	{"TEX2.TMS", 256_000_000},
	{"TEX3.TMS", 292_000_000},
}

func checkVRAM(vol *psx.Volume) int {
	m := boot(vol, "")

	// Track every CPU→VRAM upload the game issues, in order, so a cell the
	// game re-uploads after a record's own transfer (the title screen redraws
	// parts of the texture pages while the last archive is still streaming)
	// can be excluded instead of misread as a decode error.
	type up struct{ x, y, w, h int }
	var ups []up
	m.OnGP0(func(words []uint32) {
		if op := byte(words[0] >> 24); op >= 0xA0 && op <= 0xBF {
			ups = append(ups, up{
				x: int(words[1] & 0x3FF), y: int((words[1] >> 16) & 0x1FF),
				w: int(words[2] & 0xFFFF), h: int(words[2] >> 16),
			})
		}
	})
	lastTouch := make([]int, rr.VRAMW*rr.VRAMH)

	replay := rr.NewVRAM()
	var ran uint64
	fail := 0
	for _, f := range vramFiles {
		fmt.Fprintf(os.Stderr, "[vram] booting to the end of the %s stream...\n", f.name)
		m.Run(f.step - ran)
		ran = f.step

		d, err := vol.ReadFile(f.name)
		if err != nil {
			die("%s: %v", f.name, err)
		}
		rects, err := rr.ParseTMS(d)
		if err != nil {
			die("%s: %v", f.name, err)
		}
		for _, r := range rects {
			replay.Blit(r)
		}
		for i, u := range ups {
			for y := 0; y < u.h; y++ {
				for x := 0; x < u.w; x++ {
					lastTouch[((u.y+y)&(rr.VRAMH-1))*rr.VRAMW+((u.x+x)&(rr.VRAMW-1))] = i + 1
				}
			}
		}
		vram := m.VRAM()
		var diff, total, skipped int
		for _, r := range rects {
			// The record's own upload is the last observed transfer with
			// exactly its rect (chunked records have none; then no cell can
			// be excluded, which is safe — chunking splits only rects the
			// game uploads early, before any redraw).
			own := 0
			for i := len(ups) - 1; i >= 0; i-- {
				if ups[i] == (up{r.X, r.Y, r.W, r.H}) {
					own = i + 1
					break
				}
			}
			for y := 0; y < r.H; y++ {
				for x := 0; x < r.W; x++ {
					vx := (r.X + x) & (rr.VRAMW - 1)
					vy := (r.Y + y) & (rr.VRAMH - 1)
					if own > 0 && lastTouch[vy*rr.VRAMW+vx] > own {
						skipped++ // re-uploaded by the game after this record
						continue
					}
					total++
					if vram[vy*rr.VRAMW+vx] != replay.At(vx, vy) {
						diff++
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[vram] %-9s %7d words compared (%d later redrawn), %d mismatched\n",
			f.name, total, skipped, diff)
		if diff > 0 {
			fail = 1
		}
	}
	return fail
}

// --- geometry ----------------------------------------------------------

// event is one trapped GTE load: which path, the record RAM address and the
// six data registers as loaded (packed halfwords).
type event struct {
	pc   uint32
	addr uint32
	r    [6]uint32
}

func checkGeometry(vol *psx.Volume, name, press string, warmup uint64, window uint64) int {
	fmt.Fprintf(os.Stderr, "[%s] booting to the capture window...\n", name)
	m := boot(vol, press)
	m.Run(warmup)

	var events []event
	trap := map[uint32]bool{
		pcQuadRTPT: true, pcQuadRTPS: true,
		pcGTRTPT: true, pcGTRTPS: true, pcGTNCT: true, pcGTNCS: true,
	}
	m.OnStep = func(mm *psx.Machine, pc uint32) {
		if !trap[pc] {
			return
		}
		reg := regA0
		if pc == pcGTRTPT || pc == pcGTRTPS || pc == pcGTNCT || pc == pcGTNCS {
			reg = regT1
		}
		ev := event{pc: pc, addr: mm.CPU.Reg(uint32(reg)) & 0x1FFFFF}
		for i := 0; i < 6; i++ {
			ev.r[i] = mm.GTE.Read(uint32(i))
		}
		events = append(events, ev)
	}
	m.Run(window)
	m.OnStep = nil
	fmt.Fprintf(os.Stderr, "[%s] %d GTE loads trapped\n", name, len(events))
	if len(events) == 0 {
		fmt.Fprintf(os.Stderr, "[%s] FAIL: nothing rendered in the window\n", name)
		return 1
	}

	// Decode the files and index every record by its resident RAM address.
	mapData, err := vol.ReadFile("MAP.RRM")
	if err != nil {
		die("MAP.RRM: %v", err)
	}
	objData, err := vol.ReadFile("OBJ.RRO")
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	track, err := rr.ParseRRM(mapData)
	if err != nil {
		die("MAP.RRM: %v", err)
	}
	objs, err := rr.ParseRRO(objData)
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	index := buildIndex(track, objs)

	bad := 0
	for _, ev := range events {
		if err := verify(index, ev); err != nil {
			if bad < 10 {
				fmt.Fprintf(os.Stderr, "[%s] MISMATCH pc=%08X addr=%06X: %v\n", name, ev.pc, ev.addr, err)
			}
			bad++
		}
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "[%s] FAIL: %d/%d events mismatched\n", name, bad, len(events))
		return 1
	}
	fmt.Fprintf(os.Stderr, "[%s] OK: %d events, all bit-exact\n", name, len(events))
	return 0
}

// rec is a decoder-predicted record: vertex and normal blocks by RAM address.
type rec struct {
	v [4][3]int16
	n *[4][3]int16
}

// buildIndex places every decoded record at its resident address, exactly as
// the loader lays the files out.
func buildIndex(track *rr.Track, objs []rr.Object) map[uint32]rec {
	idx := map[uint32]rec{}
	// MAP.RRM: directory then quads, in section/class order.
	addr := uint32(mapBase) + 4 + uint32(len(track.Sections))*8
	for _, s := range track.Sections {
		for _, class := range [][]rr.TrackQuad{s.A, s.B, s.C} {
			for _, q := range class {
				idx[addr] = rec{v: q.V}
				addr += 40
			}
		}
	}
	// OBJ.RRO: directory then per-object records in type order.
	addr = uint32(objBase) + 4 + uint32(len(objs))*16
	for i := range objs {
		o := &objs[i]
		for _, q := range o.FT {
			idx[addr] = rec{v: q.V}
			addr += 40
		}
		for _, q := range o.FT8 {
			idx[addr] = rec{v: q.V}
			addr += 48
		}
		for _, q := range o.F {
			idx[addr] = rec{v: q.V}
			addr += 32
		}
		for k := range o.GT {
			idx[addr] = rec{v: o.GT[k].V, n: &o.GT[k].N}
			addr += 64
		}
		for k := range o.GT8 {
			idx[addr] = rec{v: o.GT8[k].V, n: &o.GT8[k].N}
			addr += 72
		}
		for _, q := range o.G {
			idx[addr] = rec{v: q.V, n: &q.N}
			addr += 56
		}
	}
	return idx
}

// verify checks one trapped event against the decoder's record at that
// address. The walkers load vertices packed: reg0 = x0|y0<<16, reg1 low = z0,
// reg2 = x1|y1<<16, reg3 low = z1, reg4 = x2|y2<<16, reg5 low = z2; the
// 4th-vertex RTPS carries x3|y3<<16 in reg0 and z3 in reg1's low half.
func verify(index map[uint32]rec, ev event) error {
	switch ev.pc {
	case pcQuadRTPT, pcGTRTPT:
		r, ok := index[ev.addr]
		if !ok {
			return fmt.Errorf("no decoded record at this address")
		}
		return checkTriple(ev.r, [3][3]int16{r.v[0], r.v[1], r.v[2]})
	case pcQuadRTPS, pcGTRTPS:
		r, ok := index[ev.addr]
		if !ok {
			return fmt.Errorf("no decoded record at this address")
		}
		return checkSingle(ev.r, r.v[3])
	case pcGTNCT:
		r, ok := index[ev.addr]
		if !ok || r.n == nil {
			return fmt.Errorf("no decoded lit record at this address")
		}
		return checkTriple(ev.r, [3][3]int16{r.n[0], r.n[1], r.n[2]})
	case pcGTNCS:
		r, ok := index[ev.addr]
		if !ok || r.n == nil {
			return fmt.Errorf("no decoded lit record at this address")
		}
		return checkSingle(ev.r, r.n[3])
	}
	return fmt.Errorf("unexpected trap pc")
}

func checkTriple(r [6]uint32, want [3][3]int16) error {
	got := [3][3]int16{
		{int16(r[0]), int16(r[0] >> 16), int16(r[1])},
		{int16(r[2]), int16(r[2] >> 16), int16(r[3])},
		{int16(r[4]), int16(r[4] >> 16), int16(r[5])},
	}
	if got != want {
		return fmt.Errorf("vectors %v != decoded %v", got, want)
	}
	return nil
}

func checkSingle(r [6]uint32, want [3]int16) error {
	got := [3]int16{int16(r[0]), int16(r[0] >> 16), int16(r[1])}
	if got != want {
		return fmt.Errorf("vector %v != decoded %v", got, want)
	}
	return nil
}
