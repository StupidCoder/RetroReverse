// kcltrace pins the .kcl collision format by driving the game's OWN collision
// code in the oracle (Part VI). It builds the collision world exactly as the
// level-data processor at $020FE190 does — load the level's .kcl through the
// game loader, run the in-place offset->pointer fixup at $02039760, construct
// a dBgW_Kc collider ($020398C8, vtable $020993DC), set it up ($020396F0) and
// register it in the 24-slot collider table at $020A0C80 ($02039184) — then
// casts the signpost's ground ray (ctor $02037570, origin $0203748C, execute
// $02038F44, hit Y at query+$44) with a READ WATCH over the served .kcl
// buffer: every byte of the file the walker touches is logged with the PC
// that touched it. The read log is the file's structure; the PCs are the
// routines to disassemble for its semantics.
//
//	kcltrace [-rom img] [-extracted dir] [-level N] [-x f -z f] [-grid N]
//	         [-itcm out.bin] [-v]
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

const (
	settingsTab = 0x02092208 // level -> settings block (RAM, u32 x52)
	loaderThunk = 0x0201816C // load by internal file ID -> buffer
	kclFixup    = 0x02039760 // .kcl header: 4 relative offsets -> absolute pointers
	kcCtor      = 0x020398C8 // dBgW_Kc constructor
	kcSetup     = 0x020396F0 // dBgW_Kc setup(ctx, kcl, scaleVec)
	kcRegister  = 0x02039184 // register collider (first free of 24 slots)
	slotArray   = 0x020A0C80 // the collider slot table
	rayCtor     = 0x02037570 // ground-ray query ctor (0x50-byte object)
	raySetOrig  = 0x0203748C // set query origin (vec3 fx20.12 at +0x38)
	rayExec     = 0x02038F44 // execute against all registered colliders
	rayDtor     = 0x02037534
	qHitY       = 0x44 // query field: ground Y on hit
)

func main() {
	rom := flag.String("rom", "../Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "../extracted", "extracted binaries dir")
	level := flag.Int("level", 1, "level to load")
	fx := flag.Float64("x", 0, "ray X (world units; default: first entrance)")
	fy := flag.Float64("y", 500, "ray Y (world units)")
	fz := flag.Float64("z", 0, "ray Z (world units; default: first entrance)")
	useEnt := flag.Bool("ent", true, "start at the level's first entrance")
	grid := flag.Int("grid", 0, "also cast an NxN ray grid around the origin (span +-2048)")
	verify := flag.Int("verify", 0, "verify the Go reimplementation against the game over N random rays")
	seed := flag.Int64("seed", 1, "random seed for -verify")
	itcm := flag.String("itcm", "", "dump ITCM (0x01FF8000+0x8000) to this file")
	verbose := flag.Bool("v", false, "trace progress")
	flag.Parse()

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		sm64ds.Die(err)
	}
	if *verbose {
		o.Trace = func(s string) { fmt.Fprintln(os.Stderr, "  |", s) }
	}
	if err := o.InitEngine(); err != nil {
		sm64ds.Die(err)
	}
	lv, err := ls.Level(*level)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := o.LoadConfig(lv.Overlay); err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("level %d (overlay %d): kcl=%s\n", *level, lv.Overlay, lv.KCLPath)

	if *itcm != "" {
		if err := os.WriteFile(*itcm, o.ReadBytes(0x01FF8000, 0x8000), 0o644); err != nil {
			sm64ds.Die(err)
		}
		fmt.Printf("ITCM dumped to %s\n", *itcm)
	}

	// --- build the collision world the way $020FE190 does ---
	settings := o.R32(settingsTab + uint32(*level)*4)
	fid := o.R16(settings + 0xA)
	clpsP := o.R32(settings) // settings+0 -> the level's CLPS surface-attribute block
	clpsN := o.R16(clpsP + 6)
	clps := sm64ds.NewCLPS(o.ReadBytes(clpsP, 8+clpsN*8))
	fmt.Printf("settings=%08X collision file id=%d CLPS=%08X (%q, %d entries)\n",
		settings, fid, clpsP, o.ReadBytes(clpsP, 4), clpsN)

	kcl, note := o.Call(loaderThunk, fid, 0, 0, 0, 5_000_000)
	if note != "" || kcl == 0 {
		sm64ds.Die(fmt.Errorf("kcl load: ptr=%08X %s", kcl, note))
	}
	lo, hi, name := o.LastServed()
	fmt.Printf("kcl buffer %08X..%08X (%s, %d bytes)\n", lo, hi, name, hi-lo)
	kclBytes := o.ReadBytes(lo, hi-lo) // pre-fixup copy for the Go decoder
	fmt.Printf("header before fixup: %08X %08X %08X %08X\n", o.R32(kcl), o.R32(kcl+4), o.R32(kcl+8), o.R32(kcl+12))
	if _, note := o.Call(kclFixup, kcl, 0, 0, 0, 10_000); note != "" {
		sm64ds.Die(fmt.Errorf("fixup: %s", note))
	}
	fmt.Printf("header after  fixup: %08X %08X %08X %08X\n", o.R32(kcl), o.R32(kcl+4), o.R32(kcl+8), o.R32(kcl+12))
	fmt.Printf("header words +10..+38:")
	for a := kcl + 0x10; a < kcl+0x38; a += 4 {
		fmt.Printf(" %08X", o.R32(a))
	}
	fmt.Println()

	ctx := o.Alloc(0x400)
	if _, note := o.Call(kcCtor, ctx, 0, 0, 0, 100_000); note != "" {
		sm64ds.Die(fmt.Errorf("ctor: %s", note))
	}
	if _, note := o.Call(kcSetup, ctx, kcl, clpsP, 0, 100_000); note != "" {
		sm64ds.Die(fmt.Errorf("setup: %s", note))
	}
	if r, note := o.Call(kcRegister, ctx, 0, 0, 0, 100_000); r != 1 || note != "" {
		sm64ds.Die(fmt.Errorf("register: r=%d %s", r, note))
	}
	if got := o.R32(slotArray); got != ctx {
		sm64ds.Die(fmt.Errorf("slot 0 = %08X, want ctx %08X", got, ctx))
	}
	fmt.Printf("collider registered: ctx=%08X (slot 0)\n", ctx)

	// --- ray origin ---
	x, y, z := *fx, *fy, *fz
	if *useEnt && len(lv.Entrances) > 0 {
		e := lv.Entrances[0]
		x, y, z = e.X, e.Y+100, e.Z
	}

	// --- read-watched single ray ---
	reads := map[uint32]int{}    // file offset -> count
	firstPC := map[uint32]bool{} // "pc:off" first-touch log
	type touch struct{ pc, off uint32 }
	var touches []touch
	o.SetReadWatch(lo, hi, func(a uint32) {
		off := a - lo
		reads[off]++
		pc := o.PC()
		k := pc ^ off<<16
		if !firstPC[k] {
			firstPC[k] = true
			touches = append(touches, touch{pc, off})
		}
	})
	r := castRay(o, int32(x*4096), int32(y*4096), int32(z*4096))
	o.SetReadWatch(0, 0, nil)
	fmt.Printf("\nray (%.0f, %.0f, %.0f): hit=%d groundY=%.2f prism=%d %s\n",
		x, y, z, r.hit, float64(r.groundY)/4096, r.prism, r.note)

	if *verbose {
		k, err := sm64ds.ParseKCL(kclBytes)
		if err == nil {
			sm64ds.KCLTrace = func(f string, a ...any) { fmt.Printf("  go| "+f+"\n", a...) }
			ours, ok := k.RaycastDown(int32(x*4096), int32(y*4096), int32(z*4096), r.floorBelow, clps, 1)
			sm64ds.KCLTrace = nil
			fmt.Printf("go reimpl: hit=%v y=%d prism=%d\n", ok, ours.Y, ours.Prism)
		}
	}

	// classify touched offsets by header section
	sec := make([]uint32, 4)
	for i := range sec {
		sec[i] = o.R32(kcl+uint32(i)*4) - lo
	}
	fmt.Printf("sections at file offsets: %X %X %X %X (file size %X)\n", sec[0], sec[1], sec[2], sec[3], hi-lo)
	var offs []uint32
	for off := range reads {
		offs = append(offs, off)
	}
	sort.Slice(offs, func(i, j int) bool { return offs[i] < offs[j] })
	fmt.Printf("%d distinct bytes read:\n", len(offs))
	for _, r := range coalesce(offs) {
		fmt.Printf("  %06X..%06X (%d bytes) section %s\n", r[0], r[1], r[1]-r[0]+1, secName(r[0], sec))
	}
	fmt.Println("code that touched the file (pc -> offset, first touches):")
	for i, t := range touches {
		if i >= 60 {
			fmt.Printf("  ... %d more\n", len(touches)-60)
			break
		}
		fmt.Printf("  %08X reads %06X (%s)\n", t.pc, t.off, secName(t.off, sec))
	}

	// --- optional verification grid (no watch: fast) ---
	if *grid > 1 {
		n := *grid
		span := 2048.0
		fmt.Printf("\ngrid %dx%d around (%.0f,%.0f), y=%.0f:\n", n, n, x, z, y)
		for iz := 0; iz < n; iz++ {
			for ix := 0; ix < n; ix++ {
				gx := x - span + 2*span*float64(ix)/float64(n-1)
				gz := z - span + 2*span*float64(iz)/float64(n-1)
				r := castRay(o, int32(gx*4096), int32(y*4096), int32(gz*4096))
				if r.hit != 0 {
					fmt.Printf("%d %.2f %.2f %.2f\n", r.hit, gx, float64(r.groundY)/4096, gz)
				} else {
					fmt.Printf("0 %.2f - %.2f\n", gx, gz)
				}
			}
		}
	}

	// --- oracle-verify the Go reimplementation ---
	if *verify > 0 {
		k, err := sm64ds.ParseKCL(kclBytes)
		if err != nil {
			sm64ds.Die(err)
		}
		rng := rand.New(rand.NewSource(*seed))
		// world-unit area bounds from the header the game itself walks
		minX, minY, minZ := k.Hdr.Min[0]>>6, k.Hdr.Min[1]>>6, k.Hdr.Min[2]>>6
		spanX, spanY, spanZ := int32(^k.Hdr.Mask[0]), int32(^k.Hdr.Mask[1]), int32(^k.Hdr.Mask[2])
		mismatch := 0
		hits := 0
		for i := 0; i < *verify; i++ {
			// random fx20.12 points across the area (10% deliberately outside)
			rx := (minX - spanX/10 + rng.Int31n(spanX+spanX/5)) << 12
			rz := (minZ - spanZ/10 + rng.Int31n(spanZ+spanZ/5)) << 12
			ry := (minY + rng.Int31n(spanY+spanY/4)) << 12
			rx += rng.Int31n(0x1000)
			rz += rng.Int31n(0x1000)
			ry += rng.Int31n(0x1000)
			g := castRay(o, rx, ry, rz)
			if g.note != "" {
				fmt.Printf("  game error at (%d,%d,%d): %s\n", rx, ry, rz, g.note)
				mismatch++
				continue
			}
			ours, ok := k.RaycastDown(rx, ry, rz, g.floorBelow, clps, 1)
			okGame := g.hit != 0
			if ok != okGame || (ok && (ours.Y != g.groundY || ours.Prism != g.prism)) {
				mismatch++
				fmt.Printf("  MISMATCH (%.1f,%.1f,%.1f): game hit=%v y=%d prism=%d | ours hit=%v y=%d prism=%d\n",
					float64(rx)/4096, float64(ry)/4096, float64(rz)/4096,
					okGame, g.groundY, g.prism, ok, ours.Y, ours.Prism)
			}
			if okGame {
				hits++
			}
		}
		fmt.Printf("\nverify: %d rays, %d hits, %d mismatches\n", *verify, hits, mismatch)
	}
}

// rayResult is one game-side down-ray answer.
type rayResult struct {
	hit        uint32
	groundY    int32 // q+0x44 after execute (fx20.12)
	floorBelow int32 // q+0x44 before execute (the walker's initial best)
	prism      int   // q+0x28 u16: winning prism record index
	note       string
}

// castRay performs the signpost's ground snap: down-ray from (x,y,z).
func castRay(o *sm64ds.Oracle, x, y, z int32) rayResult {
	q := o.Alloc(0x50)
	vec := o.Alloc(12)
	o.W32(vec, uint32(x))
	o.W32(vec+4, uint32(y))
	o.W32(vec+8, uint32(z))
	if _, n := o.Call(rayCtor, q, 0, 0, 0, 100_000); n != "" {
		return rayResult{note: "ctor: " + n}
	}
	if _, n := o.Call(raySetOrig, q, vec, 0, 0, 100_000); n != "" {
		return rayResult{note: "setOrigin: " + n}
	}
	// the manager entry resets the query first ($02037464): q+0x44 (the
	// walker's initial "best floor") becomes the -inf sentinel 0x80000000
	r := rayResult{floorBelow: -0x80000000}
	r.hit, r.note = o.Call(rayExec, q, 0, 0, 0, 10_000_000)
	r.groundY = int32(o.R32(q + qHitY))
	r.prism = int(o.R16(q + 0x28))
	o.Call(rayDtor, q, 0, 0, 0, 100_000)
	return r
}

func coalesce(offs []uint32) [][2]uint32 {
	var out [][2]uint32
	for _, o := range offs {
		if len(out) > 0 && o <= out[len(out)-1][1]+4 {
			out[len(out)-1][1] = o
			continue
		}
		out = append(out, [2]uint32{o, o})
	}
	return out
}

func secName(off uint32, sec []uint32) string {
	names := []string{"0 (+00)", "1 (+04)", "2 (+08)", "3 (+0C)"}
	best, bi := "header", -1
	for i, s := range sec {
		if off >= s && (bi < 0 || s >= sec[bi]) {
			best, bi = names[i], i
		}
	}
	if off < minU32(sec) {
		return "header"
	}
	return best
}

func minU32(v []uint32) uint32 {
	m := v[0]
	for _, x := range v {
		if x < m {
			m = x
		}
	}
	return m
}
