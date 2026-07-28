package main

// census.go — the mode that turns actorsolve from a one-off prober into the
// intro's blocking extractor.
//
// -census finds every demo actor OBJECT in a savestate by its heap signature
// (float frame counter, zero word, four ascending array pointers, an
// FFFFFFFF terminator, one more pointer), and attributes each of the shot's
// (mdl,key) actors to an object by solving A = world[0]·composed[0]⁻¹ over
// the whole node array. Each object carries stage arrays: stage 0 is the
// WORLD (A is the actor's placement, constant through the shot — verified by
// running the census at states ~200 frames apart), stage 1 is the same
// matrices premultiplied by the running camera's view matrix (chasing that
// stage produced a phantom "time-varying blocking" until its A was matched
// against the baked .scd camera: A_s1(f) = View(f)·A_s0). The world arrays
// lag the object's counter by ~2 frames; looping clips (the sets' sway, the
// torch) run their counter past the clip length, so a full-range frame
// search backs up the counter window.
//
// -emit writes the accepted per-actor placements as the shot's blocking
// table JSON (cmd/webexport/blocking/<shot>.json). The placement is computed
// by the demo player's code — no file on the disc stores it (the .scd holds
// only camera and light channels, the .sco only cut bases), so solving it
// out of the running game's own matrix arrays, against our own .key
// evaluation, is the derivation; the identity solves on the world-space sets
// are the control.

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

type actorSpec struct{ mdl, key string }

func parseActors(s string) ([]actorSpec, error) {
	var out []actorSpec
	for _, part := range strings.Split(s, ",") {
		mdl, key, ok := strings.Cut(part, ":")
		if !ok {
			return nil, fmt.Errorf("actor %q: want mdl:key", part)
		}
		out = append(out, actorSpec{mdl, key})
	}
	return out, nil
}

type actorData struct {
	spec     actorSpec
	m        *lm.MDL
	key      *lm.Key
	steps    int          // half-frame steps precomposed
	composed [][]lm.Mtx34 // [step][node]
}

func loadActors(files []lm.RARCFile, specs []actorSpec) ([]*actorData, error) {
	var out []*actorData
	for _, sp := range specs {
		m, key, err := export.LoadSkinned(files, sp.mdl, sp.key)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", sp.mdl, err)
		}
		a := &actorData{spec: sp, m: m, key: key}
		a.steps = (int(key.Duration()) + 1) * 2
		a.composed = make([][]lm.Mtx34, a.steps)
		for s := 0; s < a.steps; s++ {
			a.composed[s] = pose(m, key, float32(s)/2)
		}
		out = append(out, a)
	}
	return out, nil
}

// objHit is one heap object matching the demo-actor signature.
type objHit struct {
	addr   uint32 // of the frame-counter float
	frame  float32
	arrays []uint32
}

func findObjects(ram []byte) []objHit {
	var out []objHit
	ptr := func(off int) uint32 { return binary.BigEndian.Uint32(ram[off:]) }
	isPtr := func(v uint32) bool { return v >= 0x80003000 && v < 0x80000000+uint32(len(ram)) }
	for o := 0; o+40 <= len(ram); o += 4 {
		f := math.Float32frombits(ptr(o))
		if !(f >= 0 && f <= 4000) || float32(int(f*2)) != f*2 {
			continue
		}
		if ptr(o+4) != 0 {
			continue
		}
		p := [4]uint32{ptr(o + 8), ptr(o + 12), ptr(o + 16), ptr(o + 20)}
		if !isPtr(p[0]) || !isPtr(p[1]) || !isPtr(p[2]) || !isPtr(p[3]) {
			continue
		}
		if !(p[0] < p[1] && p[1] < p[2] && p[2] < p[3]) {
			continue
		}
		if ptr(o+24) != 0xFFFFFFFF || !isPtr(ptr(o+28)) {
			continue
		}
		out = append(out, objHit{
			addr:   0x80000000 + uint32(o),
			frame:  f,
			arrays: []uint32{p[0], p[1], p[2], p[3], ptr(o + 28)},
		})
	}
	return out
}

// solveAt computes A = arr[0]·composed[0](step)⁻¹ and the max error over all
// nodes of the actor at that array.
func solveAt(ram []byte, arr uint32, a *actorData, step int) (lm.Mtx34, float32) {
	base := int(arr - 0x80000000)
	n := len(a.m.Nodes)
	if base < 0 || base+n*48 > len(ram) {
		return lm.Mtx34{}, float32(math.MaxFloat32)
	}
	got0 := readMtx(ram, base)
	A := got0.Mul(invert34(a.composed[step][0]))
	worst := float32(0)
	for i := 0; i < n; i++ {
		e := maxDiff(A.Mul(a.composed[step][i]), readMtx(ram, base+i*48))
		if e > worst {
			worst = e
		}
	}
	return A, worst
}

// solveStage finds the best A for one actor against one array over a list of
// candidate steps, rejecting degenerate (zeroed / scaled) matches.
func solveStage(ram []byte, arr uint32, a *actorData, steps []int) (lm.Mtx34, float32) {
	bestA, bestE := lm.Mtx34{}, float32(math.MaxFloat32)
	for _, step := range steps {
		if step < 0 || step >= a.steps {
			continue
		}
		A, e := solveAt(ram, arr, a, step)
		if !orthonormalish(A) {
			continue
		}
		if e < bestE {
			bestA, bestE = A, e
		}
	}
	return bestA, bestE
}

func census(image, state, state2, arc, actorsFlag string, tol float32, emit string) error {
	src, err := export.Open(image)
	if err != nil {
		return err
	}
	defer src.Close()
	files, err := src.Archive(arc)
	if err != nil {
		return err
	}
	specs, err := parseActors(actorsFlag)
	if err != nil {
		return err
	}
	actors, err := loadActors(files, specs)
	if err != nil {
		return err
	}
	ram, err := loadRAM(image, state)
	if err != nil {
		return err
	}
	objs := findObjects(ram)
	fmt.Fprintf(os.Stderr, "census: %d candidate objects\n", len(objs))

	// Liveness: a live actor's frame counter ticks — or, when its clip has
	// ended (the sets play their sway once and park), its stage-1 array
	// still changes every frame because that stage is premultiplied by the
	// moving camera's view matrix. A stale object from an earlier shot
	// freezes entirely. A second state a few fields later separates them.
	var ram2 []byte
	stale := []objHit{}
	if state2 != "" {
		ram2, err = loadRAM(image, state2)
		if err != nil {
			return err
		}
		changed := func(addr uint32, n int) bool {
			off := int(addr - 0x80000000)
			if off < 0 || off+n > len(ram) || off+n > len(ram2) {
				return false
			}
			return !bytesEqual(ram[off:off+n], ram2[off:off+n])
		}
		liveObjs := objs[:0:0]
		for _, o := range objs {
			if changed(o.addr, 4) || changed(o.arrays[0], 48) || changed(o.arrays[1], 48) {
				liveObjs = append(liveObjs, o)
			} else {
				stale = append(stale, o)
			}
		}
		objs = liveObjs
		fmt.Fprintf(os.Stderr, "census: %d live between the two states\n", len(objs))
	}

	table := &export.BlockingTable{
		Note: "Constant per-shot actor placement A, solved from the demo player's own stage-0 " +
			"world-matrix arrays as A = world[i]·composed[i]⁻¹ over every node (stage 1 is the " +
			"same premultiplied by the camera view). The demo player computes these in code; " +
			"no disc file stores them. Verified constant across states ~200 frames apart; the " +
			"world-space sets solve to identity as the control.",
		Actors: map[string]*export.Blocking{},
	}
	allOK := true
	var ranges []string

	// Each actor gets its own object (greedy best-first assignment) — the
	// torch must not ride the mansion's object just because a 3-node fit
	// is easy.
	live := objs

	window := func(center int) []int {
		var out []int
		for df := -6; df <= 0; df++ {
			out = append(out, center+df)
		}
		return out
	}
	type solved struct {
		ai, oi int
		err    float32
		errs   [2]float32
		as     [2]lm.Mtx34
	}
	var all []solved
	for ai, a := range actors {
		for oi, o := range live {
			var sv solved
			sv.ai, sv.oi = ai, oi
			for pass := 0; pass < 2; pass++ {
				var steps []int
				if pass == 0 {
					c := int(o.frame * 2)
					if c >= a.steps {
						c = c % a.steps
					}
					steps = window(c)
				} else {
					for st := 0; st < a.steps; st++ {
						steps = append(steps, st)
					}
				}
				sv.as[0], sv.errs[0] = solveStage(ram, o.arrays[0], a, steps)
				sv.as[1], sv.errs[1] = solveStage(ram, o.arrays[1], a, steps)
				sv.err = sv.errs[0]
				if sv.errs[1] < sv.err {
					sv.err = sv.errs[1]
				}
				if sv.err <= tol {
					break
				}
			}
			// Temporal coherence: a few-node actor can "solve" against a
			// foreign array with error 0. The real pairing solves on BOTH
			// states with nearly the same A (hand-held props drift a few
			// units per field; a false match jumps wildly or dies). The
			// second solve stays inside a window under state 2's counter —
			// a global search would wander to a random best fit.
			if ram2 != nil && sv.errs[0] <= tol {
				off := int(o.addr - 0x80000000)
				c2 := math.Float32frombits(binary.BigEndian.Uint32(ram2[off:]))
				center := int(c2 * 2)
				if center >= a.steps {
					center = center % a.steps
				}
				A2, e2 := solveStage(ram2, o.arrays[0], a, window(center))
				if e2 > tol || maxDiff(A2, sv.as[0]) > 1000 {
					sv.err = float32(math.MaxFloat32)
					sv.errs[0] = float32(math.MaxFloat32)
				}
			}
			all = append(all, sv)
		}
	}
	// Unresolved actors may sit on frozen objects (a set whose sway clip
	// ended under a nearly static camera): offer the stale objects too, at
	// lower priority (they sort behind every live solve of equal error).
	for ai, a := range actors {
		for si2, o := range stale {
			var sv solved
			sv.ai, sv.oi = ai, len(live)+si2
			c := int(o.frame * 2)
			if c >= a.steps {
				c = c % a.steps
			}
			sv.as[0], sv.errs[0] = solveStage(ram, o.arrays[0], a, window(c))
			sv.as[1], sv.errs[1] = solveStage(ram, o.arrays[1], a, window(c))
			sv.err = sv.errs[0]
			if sv.errs[1] < sv.err {
				sv.err = sv.errs[1]
			}
			sv.err += 0.001 // sort behind live ties
			sv.errs[0] += 0.001
			all = append(all, sv)
		}
	}
	live = append(live, stale...)

	sort.Slice(all, func(i, j int) bool { return all[i].err < all[j].err })
	pickA := make([]int, len(actors))
	for i := range pickA {
		pickA[i] = -1
	}
	usedO := make([]bool, len(live))
	// Strong actors first: a 122-node fit at err 60 is evidence, a 3-node
	// fit at err 0.000 is a coin toss that must not steal an object from a
	// stronger claim (the torch once stole the set's object this way).
	for _, sv := range all {
		if len(actors[sv.ai].m.Nodes) < 8 || pickA[sv.ai] >= 0 || usedO[sv.oi] || sv.errs[0] > tol {
			continue
		}
		pickA[sv.ai] = sv.oi
		usedO[sv.oi] = true
	}
	// The shot clock: what the strong actors' live objects tick at. Weak
	// (few-node) actors may only claim objects on that clock — a parked
	// object from another shot fits them far too easily.
	clock := float32(-1)
	for ai := range actors {
		if pickA[ai] >= 0 && len(actors[ai].m.Nodes) >= 8 {
			c := live[pickA[ai]].frame
			if clock < 0 || c > clock {
				clock = c
			}
		}
	}
	for _, sv := range all {
		if len(actors[sv.ai].m.Nodes) >= 8 || pickA[sv.ai] >= 0 || usedO[sv.oi] || sv.errs[0] > tol {
			continue
		}
		if clock >= 0 {
			d := live[sv.oi].frame - clock
			if d < -1 || d > 1 {
				continue
			}
		}
		pickA[sv.ai] = sv.oi
		usedO[sv.oi] = true
	}

	for ai, a := range actors {
		if pickA[ai] < 0 {
			fmt.Printf("actor %-40s NO MATCH (%d nodes)\n", a.spec.mdl+"+"+a.spec.key, len(a.m.Nodes))
			allOK = false
			continue
		}
		var b solved
		for _, sv := range all {
			if sv.ai == ai && sv.oi == pickA[ai] {
				b = sv
				break
			}
		}
		o := live[b.oi]
		fmt.Printf("actor %-40s counter 0x%08X frame %6.1f nodes %3d s0 err %8.3f s1 err %8.3f\n",
			a.spec.mdl+"+"+a.spec.key, o.addr, o.frame, len(a.m.Nodes), b.errs[0], b.errs[1])
		A := b.as[0] // stage 0 = world
		fmt.Printf("  A [% .4f % .4f % .4f % .1f | % .4f % .4f % .4f % .1f | % .4f % .4f % .4f % .1f]\n",
			A[0], A[1], A[2], A[3], A[4], A[5], A[6], A[7], A[8], A[9], A[10], A[11])

		bl := &export.Blocking{}
		if maxDiff(A, lm.Identity34) <= tol {
			bl.Identity = true
		} else {
			row := make([]float64, 13)
			for i := 0; i < 12; i++ {
				row[i+1] = float64(A[i])
			}
			bl.Frames = [][]float64{row}
		}
		table.Actors[a.spec.mdl+"+"+a.spec.key] = bl
		nb := 192
		if n := len(a.m.Nodes) * 48; n < nb {
			nb = n
		}
		ranges = append(ranges,
			fmt.Sprintf("%X:4", o.addr-0x80000000),
			fmt.Sprintf("%X:%X", o.arrays[0]-0x80000000, nb))
	}
	fmt.Printf("memrec %s\n", strings.Join(ranges, ","))
	if emit != "" {
		if !allOK {
			return fmt.Errorf("not all actors solved; refusing to emit %s", emit)
		}
		table.Shot = strings.TrimSuffix(emit[strings.LastIndex(emit, "/")+1:], ".json")
		j, err := json.MarshalIndent(table, "", " ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(emit, append(j, '\n'), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "census: wrote %s\n", emit)
	}
	return nil
}

func bytesEqual(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// orthonormalish screens a 3x4: unit-ish rows, |det| ≈ 1.
func orthonormalish(m lm.Mtx34) bool {
	for r := 0; r < 3; r++ {
		l := m[r*4]*m[r*4] + m[r*4+1]*m[r*4+1] + m[r*4+2]*m[r*4+2]
		if l < 0.9 || l > 1.1 {
			return false
		}
	}
	d := float64(m[0])*(float64(m[5])*float64(m[10])-float64(m[6])*float64(m[9])) -
		float64(m[1])*(float64(m[4])*float64(m[10])-float64(m[6])*float64(m[8])) +
		float64(m[2])*(float64(m[4])*float64(m[9])-float64(m[5])*float64(m[8]))
	return math.Abs(math.Abs(d)-1) < 0.1
}
