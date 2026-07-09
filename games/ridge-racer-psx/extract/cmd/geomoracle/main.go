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
	"sort"
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
	pcVisRTPT  = 0x8004615C // visibility precheck (0x80046114), all draw variants (record in $a0)
	pcGT2RTPT  = 0x80046D64 // translucent 64/72-byte lit path: RTPT (record in $t1)
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
	if all || want["dynamics"] {
		fail += checkDynamics(vol)
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

// event is one trapped GTE load: which path, the record RAM address, the six
// data registers as loaded (packed halfwords), and — on the 40-byte-quad path
// — the GTE control registers (rotation + translation) for the placement
// check.
type event struct {
	pc   uint32
	addr uint32
	r    [6]uint32
	ctrl [8]uint32
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
		if pc == pcQuadRTPT {
			for i := range ev.ctrl {
				ev.ctrl[i] = mm.GTE.ReadCtrl(uint32(i))
			}
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

	fail := 0
	if name == "track" {
		fail += checkPlacement(vol, track, events)
		fail += checkRaceVRAM(vol, m)
		fail += checkObjectPlacement(vol)
	}
	return fail
}

// checkObjectPlacement verifies the roadside object placements (objects.go):
// it drives into the race, captures each drawn OBJ.RRO object's id (from the
// record address) and its GTE-recovered camera-relative world position
// (Rᵀ·TR), and checks that every pairwise position delta between drawn objects
// equals the decoded placement-table delta — pinning both the table's
// positions and the ×4 unit scale, camera-independently.
func checkObjectPlacement(vol *psx.Volume) int {
	objData, err := vol.ReadFile("OBJ.RRO")
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	objs, err := rr.ParseRRO(objData)
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	// OBJ.RRO record spans, by resident RAM address, to map a record to its id.
	type span struct{ lo, hi uint32 }
	spans := make([]span, len(objs))
	addr := uint32(objBase) + 4 + uint32(len(objs))*16
	for i := range objs {
		o := &objs[i]
		n := uint32(len(o.FT)*40 + len(o.FT8)*48 + len(o.F)*32 + len(o.GT)*64 + len(o.GT8)*72 + len(o.G)*56)
		spans[i] = span{addr, addr + n}
		addr += n
	}
	idOf := func(a uint32) int {
		for i, s := range spans {
			if a >= s.lo && a < s.hi {
				return i
			}
		}
		return -1
	}

	m := boot(vol, pressRace)
	m.Run(429_500_000)
	type pos struct{ x, z int64 }
	seen := map[int]pos{}
	m.OnStep = func(mm *psx.Machine, pc uint32) {
		reg := regA0
		switch pc {
		case pcQuadRTPT:
		case pcGTRTPT:
			reg = regT1
		default:
			return
		}
		id := idOf(mm.CPU.Reg(uint32(reg)) & 0x1FFFFF)
		if id < 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		var c [8]int32
		for i := range c {
			c[i] = int32(mm.GTE.ReadCtrl(uint32(i)))
		}
		R := [3][3]int64{
			{int64(int16(c[0])), int64(int16(c[0] >> 16)), int64(int16(c[1]))},
			{int64(int16(c[1] >> 16)), int64(int16(c[2])), int64(int16(c[2] >> 16))},
			{int64(int16(c[3])), int64(int16(c[3] >> 16)), int64(int16(c[4]))},
		}
		tr := [3]int64{int64(c[5]), int64(c[6]), int64(c[7])}
		seen[id] = pos{
			(R[0][0]*tr[0] + R[1][0]*tr[1] + R[2][0]*tr[2]) / 4096,
			(R[0][2]*tr[0] + R[1][2]*tr[1] + R[2][2]*tr[2]) / 4096,
		}
	}
	m.Run(1_000_000)
	m.OnStep = nil

	_, exe, err := vol.BootEXE()
	if err != nil {
		die("boot exe: %v", err)
	}
	// Only objects placed at a single position are usable for the pairwise
	// check; repeated objects (the barrier is placed dozens of times) have no
	// unique table position to compare a drawn instance against.
	count := map[int]int{}
	tbl := map[int]rr.Placement{}
	for _, p := range rr.Placements(exe.Text) {
		count[p.Obj]++
		tbl[p.Obj] = p
	}
	for id, c := range count {
		if c > 1 {
			delete(tbl, id)
		}
	}
	var ids []int
	for id := range seen {
		if _, ok := tbl[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	if len(ids) < 2 {
		fmt.Fprintf(os.Stderr, "[track] object placement: only %d drawn objects in the table (skipped)\n", len(ids))
		return 0
	}
	ok, bad := 0, 0
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a, b := ids[i], ids[j]
			dwx := seen[a].x - seen[b].x
			dwz := seen[a].z - seen[b].z
			dtx := int64(tbl[a].X - tbl[b].X)
			dtz := int64(tbl[a].Z - tbl[b].Z)
			if abs64(dwx-dtx) <= 8 && abs64(dwz-dtz) <= 8 {
				ok++
			} else {
				if bad < 5 {
					fmt.Fprintf(os.Stderr, "[track] object MISMATCH %d vs %d: draw d=(%d,%d) table d=(%d,%d)\n",
						a, b, dwx, dwz, dtx, dtz)
				}
				bad++
			}
		}
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "[track] object placement FAIL: %d/%d pairs off\n", bad, ok+bad)
		return 1
	}
	fmt.Fprintf(os.Stderr, "[track] object placement OK: %d drawn objects, all %d pairwise deltas match the table\n", len(ids), ok)
	return 0
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// checkDynamics verifies the dynamic-object catalog (rr.Dynamics) against the
// running game. The transform setup (0x80012148) computes TR = 4·R_cam·(pos −
// cam) with the camera position read as quarter-unit halfwords at 0x801DCB84,
// and loads R = R_cam·R_obj — so at each trapped draw, R_obj·(Rᵀ·TR)/4 + cam
// reconstructs the object's absolute position (mod 2¹⁶ quarter units), which
// must land on the decoded EXE vector. Two capture windows cover the targets:
// the race countdown (the number girl) and the attract demo lap (beacon, start
// banner, both big screens, the crowd gate). Spinning and path-flying objects
// have no constant pose to check and are excluded. A small fraction of events
// read the camera variables one frame stale (the game moves the camera between
// a draw's transform setup and our trap), so the check tolerates a minority of
// off-by-a-frame outliers; a wrong decode is off on every event.
func checkDynamics(vol *psx.Volume) int {
	objData, err := vol.ReadFile("OBJ.RRO")
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	objs, err := rr.ParseRRO(objData)
	if err != nil {
		die("OBJ.RRO: %v", err)
	}
	type span struct{ lo, hi uint32 }
	spans := make([]span, len(objs))
	addr := uint32(objBase) + 4 + uint32(len(objs))*16
	for i := range objs {
		o := &objs[i]
		n := uint32(len(o.FT)*40 + len(o.FT8)*48 + len(o.F)*32 + len(o.GT)*64 + len(o.GT8)*72 + len(o.G)*56)
		spans[i] = span{addr, addr + n}
		addr += n
	}
	idOf := func(a uint32) int {
		for i, s := range spans {
			if a >= s.lo && a < s.hi {
				return i
			}
		}
		return -1
	}

	_, exe, err := vol.BootEXE()
	if err != nil {
		die("boot exe: %v", err)
	}
	// Targets: dynamics with a fixed position and a constant pose.
	type target struct {
		d   rr.Dynamic
		hit bool
	}
	targets := map[int][]*target{} // base object id -> candidate instances
	wantNames := map[string]bool{
		"Number girl": true, "Beacon (palette-cycled)": true, "Start banner": true,
		"Big screen A": true, "Big screen B": true, "Start gate crowd (far LOD)": true,
	}
	for _, d := range rr.Dynamics(exe.Text) {
		if !wantNames[d.Name] {
			continue
		}
		targets[d.Objs[0]] = append(targets[d.Objs[0]], &target{d: d})
	}

	rot := func(R [3][3]int32, v [3]int64) [3]int64 {
		var o [3]int64
		for r := 0; r < 3; r++ {
			o[r] = (int64(R[r][0])*v[0] + int64(R[r][1])*v[1] + int64(R[r][2])*v[2]) / 4096
		}
		return o
	}

	okc := map[string]int{}
	badc := map[string]int{}
	capture := func(m *psx.Machine, window uint64) {
		camHW := func(a uint32) int64 {
			return int64(uint32(m.Read(a)) | uint32(m.Read(a+1))<<8)
		}
		m.OnStep = func(mm *psx.Machine, pc uint32) {
			reg := regA0
			switch pc {
			case pcQuadRTPT, pcVisRTPT:
			case pcGTRTPT, pcGT2RTPT:
				reg = regT1
			default:
				return
			}
			id := idOf(mm.CPU.Reg(uint32(reg)) & 0x1FFFFF)
			if id < 0 {
				return
			}
			cands, ok := targets[id]
			if !ok {
				return
			}
			var c [8]int32
			for i := range c {
				c[i] = int32(mm.GTE.ReadCtrl(uint32(i)))
			}
			R := [3][3]int64{
				{int64(int16(c[0])), int64(int16(c[0] >> 16)), int64(int16(c[1]))},
				{int64(int16(c[1] >> 16)), int64(int16(c[2])), int64(int16(c[2] >> 16))},
				{int64(int16(c[3])), int64(int16(c[3] >> 16)), int64(int16(c[4]))},
			}
			tr := [3]int64{int64(c[5]), int64(c[6]), int64(c[7])}
			raw := [3]int64{
				(R[0][0]*tr[0] + R[1][0]*tr[1] + R[2][0]*tr[2]) / 4096,
				(R[0][1]*tr[0] + R[1][1]*tr[1] + R[2][1]*tr[2]) / 4096,
				(R[0][2]*tr[0] + R[1][2]*tr[1] + R[2][2]*tr[2]) / 4096,
			}
			// The camera position the transform setup subtracted, quarter units.
			cam := [3]int64{camHW(0x801DCB84), camHW(0x801DCB88), camHW(0x801DCB8C)}
			matched := false
			var bx, by, bz int64
			for _, t := range cands {
				Ro := rr.YawMatrix(t.d.Yaw)
				if t.d.Pitch != 0 {
					Ro = mul33(rr.YawMatrix(t.d.Yaw), rr.PitchMatrix(t.d.Pitch))
				}
				rel := rot(Ro, raw) // 4·(pos − cam) in model units
				var err [3]int64
				for i, want := range []int64{int64(t.d.X), int64(t.d.Y), int64(t.d.Z)} {
					got := (cam[i] + rel[i]/4) & 0xFFFF     // quarter units, u16 wrap
					err[i] = int64(int16(got - (want/4)&0xFFFF)) * 4
				}
				if abs64(err[0]) <= 32 && abs64(err[1]) <= 32 && abs64(err[2]) <= 32 {
					matched, t.hit = true, true
					break
				}
				if e := abs64(err[0]) + abs64(err[1]) + abs64(err[2]); bx+by+bz == 0 || e < bx+by+bz {
					bx, by, bz = abs64(err[0]), abs64(err[1]), abs64(err[2])
				}
			}
			name := cands[0].d.Name
			if matched {
				okc[name]++
			} else {
				badc[name]++
				if badc[name] < 4 {
					fmt.Fprintf(os.Stderr, "[dynamics] MISMATCH obj %d: err (%d,%d,%d)\n", id, bx, by, bz)
				}
			}
		}
		m.Run(window)
		m.OnStep = nil
	}

	fmt.Fprintln(os.Stderr, "[dynamics] booting into the race (countdown window)...")
	m := boot(vol, pressRace)
	m.Run(415_000_000)
	capture(m, 20_000_000)

	fmt.Fprintln(os.Stderr, "[dynamics] booting the attract demo (lap window)...")
	m = boot(vol, "")
	m.Run(600_000_000)
	capture(m, 3_400_000_000)

	fail := 0
	for _, cands := range targets {
		for _, t := range cands {
			name := t.d.Name
			if t.hit || okc[name] > 0 {
				continue // hit, or the sibling instance covered the shared id
			}
			fmt.Fprintf(os.Stderr, "[dynamics] %s (obj %d): not observed/verified\n", name, t.d.Objs[0])
			fail = 1
		}
	}
	for name, n := range okc {
		fmt.Fprintf(os.Stderr, "[dynamics] %-28s %5d position checks ok, %d off\n", name, n, badc[name])
		if badc[name] > n/8 {
			fail = 1
		}
	}
	if fail == 0 {
		fmt.Fprintln(os.Stderr, "[dynamics] OK")
	}
	return fail
}

// mul33 multiplies two 4096-scaled 3×3 matrices.
func mul33(a, b [3][3]int32) [3][3]int32 {
	var o [3][3]int32
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			var s int64
			for k := 0; k < 3; k++ {
				s += int64(a[r][k]) * int64(b[k][c])
			}
			o[r][c] = int32(s / 4096)
		}
	}
	return o
}

// checkPlacement verifies the grid placement: the 40-byte-quad path's GTE
// translation is the camera-rotated cell offset, so Rᵀ·TR recovers each drawn
// section's camera-relative world position. Differences between two sections
// must equal their grid-cell deltas × CellModel (8192 model units per cell) —
// this pins the cell pitch, the x mirror and the orientation in one check.
func checkPlacement(vol *psx.Volume, track *rr.Track, events []event) int {
	idxData, err := vol.ReadFile("IDX.HED")
	if err != nil {
		die("IDX.HED: %v", err)
	}
	grid, err := rr.ParseIDX(idxData)
	if err != nil {
		die("IDX.HED: %v", err)
	}
	cellOf := map[int][2]int{}
	for z := 0; z < 32; z++ {
		for x := 0; x < 32; x++ {
			if s := grid.Section(x, z); s != rr.Empty {
				cellOf[int(s)] = [2]int{x, z}
			}
		}
	}
	// Section spans by resident address, to map a trapped record to its section.
	type span struct {
		lo, hi uint32
		sec    int
	}
	var spans []span
	addr := uint32(mapBase) + 4 + uint32(len(track.Sections))*8
	for i, s := range track.Sections {
		n := uint32(len(s.A)+len(s.B)+len(s.C)) * 40
		spans = append(spans, span{addr, addr + n, i})
		addr += n
	}
	secOf := func(a uint32) int {
		for _, s := range spans {
			if a >= s.lo && a < s.hi {
				return s.sec
			}
		}
		return -1
	}

	// One recovered position per section: world = Rᵀ·TR / 4096.
	type pos struct{ x, y, z float64 }
	got := map[int]pos{}
	for _, ev := range events {
		if ev.pc != pcQuadRTPT {
			continue
		}
		sec := secOf(ev.addr)
		if sec < 0 {
			continue // an OBJ.RRO flat-textured record, not a track section
		}
		if _, ok := got[sec]; ok {
			continue
		}
		var R [3][3]float64
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				i := r*3 + c
				h := uint16(ev.ctrl[i/2] >> (16 * uint(i%2)))
				R[r][c] = float64(int16(h))
			}
		}
		tr := [3]float64{float64(int32(ev.ctrl[5])), float64(int32(ev.ctrl[6])), float64(int32(ev.ctrl[7]))}
		var w pos
		w.x = (R[0][0]*tr[0] + R[1][0]*tr[1] + R[2][0]*tr[2]) / 4096
		w.y = (R[0][1]*tr[0] + R[1][1]*tr[1] + R[2][1]*tr[2]) / 4096
		w.z = (R[0][2]*tr[0] + R[1][2]*tr[1] + R[2][2]*tr[2]) / 4096
		got[sec] = w
	}
	if len(got) < 2 {
		fmt.Fprintf(os.Stderr, "[track] placement FAIL: only %d sections recovered\n", len(got))
		return 1
	}
	const tol = 256.0 // fixed-point noise; a wrong pitch or mirror is ≥ 8192
	secs := make([]int, 0, len(got))
	for s := range got {
		secs = append(secs, s)
	}
	bad := 0
	for i := 0; i < len(secs); i++ {
		for j := i + 1; j < len(secs); j++ {
			a, b := secs[i], secs[j]
			ca, cb := cellOf[a], cellOf[b]
			dx := got[a].x - got[b].x - float64((ca[0]-cb[0])*rr.CellModel)
			dy := got[a].y - got[b].y
			dz := got[a].z - got[b].z - float64((ca[1]-cb[1])*rr.CellModel)
			if dx < -tol || dx > tol || dy < -tol || dy > tol || dz < -tol || dz > tol {
				if bad < 5 {
					fmt.Fprintf(os.Stderr, "[track] placement MISMATCH sec %d vs %d: err=(%.0f,%.0f,%.0f)\n",
						a, b, dx, dy, dz)
				}
				bad++
			}
		}
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "[track] placement FAIL: %d pair(s) off\n", bad)
		return 1
	}
	fmt.Fprintf(os.Stderr, "[track] placement OK: %d sections, all pairwise cell deltas match ×%d\n",
		len(got), rr.CellModel)
	return 0
}

// checkRaceVRAM compares the race-time texture half of VRAM against the pure
// file-byte reconstruction (boot replay with the scenery quadrant restored
// from its pre-menu snapshot).
func checkRaceVRAM(vol *psx.Volume, m *psx.Machine) int {
	var rects [5][]rr.Rect
	for i, name := range []string{"TEX4.TMS", "TEX0.TMS", "TEX1.TMS", "TEX2.TMS", "TEX3.TMS"} {
		d, err := vol.ReadFile(name)
		if err != nil {
			die("%s: %v", name, err)
		}
		if rects[i], err = rr.ParseTMS(d); err != nil {
			die("%s: %v", name, err)
		}
	}
	replay := rr.NewRaceVRAM(rects[0], rects[1], rects[2], rects[3], rects[4])
	vram := m.VRAM()
	diff := 0
	for y := 0; y < rr.VRAMH; y++ {
		for x := 320; x < rr.VRAMW; x++ {
			if vram[y*rr.VRAMW+x] != replay.At(x, y) {
				diff++
			}
		}
	}
	fmt.Fprintf(os.Stderr, "[track] race texture VRAM: %d words compared, %d mismatched\n",
		(rr.VRAMW-320)*rr.VRAMH, diff)
	if diff > 0 {
		return 1
	}
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
