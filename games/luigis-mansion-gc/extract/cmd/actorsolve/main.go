// actorsolve recovers a demo shot's per-actor placement matrix A from a
// bootoracle savestate taken inside the shot. The demo player keeps every
// actor's world matrices as a flat array of 3x4 floats (the model object's
// +0x14 pointer), each ram[i] = A · composed[i](f) where composed is our own
// .key evaluation through the node tree and f is the shot frame the state
// happens to hold. Neither the array address, A, nor f is known up front:
// the tool scans main RAM for plausible node-count-sized matrix arrays, then
// for each candidate searches all frames for the one that makes
// A := ram[0]·composed[0]⁻¹ constant across every node.
//
// The world-space sets (op_mansion, the forest) solve to A = identity — the
// built-in control that the scan, the evaluator and the composition all
// agree with the game before any character's A is trusted.
//
//	usage: actorsolve -image DISC.iso -state STATE \
//	       -arc /Ajioka/ADemo/opwf.szp -mdl opwf_luigi.mdl -key opwf_luigi.key
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "disc image")
	state := flag.String("state", "", "bootoracle savestate inside the shot")
	state2 := flag.String("state2", "", "second savestate a few fields later (census liveness filter)")
	arc := flag.String("arc", "", "shot archive, e.g. /Ajioka/ADemo/opwf.szp")
	mdl := flag.String("mdl", "", "actor model member")
	keyName := flag.String("key", "", "actor .key member")
	tol := flag.Float64("tol", 1.5, "max per-element error against A·composed")
	array := flag.Uint64("array", 0, "known world-matrix array address: skip the scan, solve A directly")
	frame := flag.Float64("frame", -1, "known shot frame (with -array); -1 searches all frames")
	actors := flag.String("actors", "", "the shot's actors as mdl:key[,mdl:key...] (census mode)")
	doCensus := flag.Bool("census", false, "find every actor object in the state, attribute to -actors, solve each placement A")
	emit := flag.String("emit", "", "with -census: write the accepted placements as this blocking-table JSON")
	flag.Parse()

	var err error
	switch {
	case *doCensus:
		err = census(*image, *state, *state2, *arc, *actors, float32(*tol)*100, *emit)
	default:
		err = run(*image, *state, *arc, *mdl, *keyName, float32(*tol), uint32(*array), float32(*frame))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "actorsolve:", err)
		os.Exit(1)
	}
}

// loadRAM loads a savestate and returns main RAM.
func loadRAM(image, state string) ([]byte, error) {
	disc, err := gc.Open(image)
	if err != nil {
		return nil, err
	}
	mach, err := gc.NewMachine(disc)
	if err != nil {
		return nil, err
	}
	if err := mach.LoadStateFile(state); err != nil {
		return nil, err
	}
	return mach.RAM, nil
}

func run(image, state, arc, mdlName, keyName string, tol float32, array uint32, frame float32) error {
	src, err := export.Open(image)
	if err != nil {
		return err
	}
	defer src.Close()
	files, err := src.Archive(arc)
	if err != nil {
		return err
	}
	m, key, err := export.LoadSkinned(files, mdlName, keyName)
	if err != nil {
		return err
	}
	n := len(m.Nodes)

	disc, err := gc.Open(image)
	if err != nil {
		return err
	}
	mach, err := gc.NewMachine(disc)
	if err != nil {
		return err
	}
	if err := mach.LoadStateFile(state); err != nil {
		return err
	}
	ram := mach.RAM

	// Pre-compose our pose over the clip, at half-frame steps: the 30 fps
	// demo clock lands between clip frames on odd video fields.
	steps := (int(key.Duration()) + 1) * 2
	composed := make([][]lm.Mtx34, steps)
	for s := 0; s < steps; s++ {
		composed[s] = pose(m, key, float32(s)/2)
	}

	// Direct mode: a known array (the model object's +0x14 pointer), solve A
	// there and report the residual per frame (or at one known frame).
	if array != 0 {
		base := int(array - 0x80000000)
		got := make([]lm.Mtx34, n)
		for i := 0; i < n; i++ {
			got[i] = readMtx(ram, base+i*48)
		}
		lo, hi := 0, steps
		if frame >= 0 {
			lo, hi = int(frame*2), int(frame*2)+1
		}
		bestErr, bestS := float32(math.MaxFloat32), -1
		for s := lo; s < hi; s++ {
			a := got[0].Mul(invert34(composed[s][0]))
			worst := float32(0)
			for i := 0; i < n; i++ {
				if e := maxDiff(a.Mul(composed[s][i]), got[i]); e > worst {
					worst = e
				}
			}
			if worst < bestErr {
				bestErr, bestS = worst, s
			}
		}
		a := got[0].Mul(invert34(composed[bestS][0]))
		fmt.Printf("array 0x%08X frame %.1f maxerr %.4f\n", array, float32(bestS)/2, bestErr)
		fmt.Printf("A row0 % .6f % .6f % .6f % .3f\n", a[0], a[1], a[2], a[3])
		fmt.Printf("A row1 % .6f % .6f % .6f % .3f\n", a[4], a[5], a[6], a[7])
		fmt.Printf("A row2 % .6f % .6f % .6f % .3f\n", a[8], a[9], a[10], a[11])
		return nil
	}

	// Scan for candidate arrays: n consecutive 3x4 float rows whose rotation
	// block is orthogonal-ish and whose translation is world-sized.
	found := 0
	for base := 0; base+n*48 <= len(ram); base += 4 {
		if !plausible(ram, base) || !plausible(ram, base+48) {
			continue
		}
		ok := true
		for i := 0; i < n; i++ {
			if !plausible(ram, base+i*48) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		got := make([]lm.Mtx34, n)
		for i := 0; i < n; i++ {
			got[i] = readMtx(ram, base+i*48)
		}
		// Which frame (if any) makes A constant across all nodes?
		bestErr, bestS := float32(math.MaxFloat32), -1
		for s := 0; s < steps; s++ {
			a := got[0].Mul(invert34(composed[s][0]))
			worst := float32(0)
			for i := 0; i < n; i++ {
				e := maxDiff(a.Mul(composed[s][i]), got[i])
				if e > worst {
					worst = e
				}
				if worst > tol && worst > bestErr {
					break
				}
			}
			if worst < bestErr {
				bestErr, bestS = worst, s
			}
		}
		if bestS < 0 || bestErr > tol {
			// report near misses for diagnosis when nothing passes
			if bestS >= 0 && bestErr < tol*200 {
				fmt.Printf("near miss: array 0x%08X frame %.1f maxerr %.3f\n", 0x80000000+base, float32(bestS)/2, bestErr)
			}
			continue
		}
		a := got[0].Mul(invert34(composed[bestS][0]))
		found++
		fmt.Printf("array 0x%08X frame %.1f maxerr %.4f\n", 0x80000000+base, float32(bestS)/2, bestErr)
		fmt.Printf("A row0 % .6f % .6f % .6f % .3f\n", a[0], a[1], a[2], a[3])
		fmt.Printf("A row1 % .6f % .6f % .6f % .3f\n", a[4], a[5], a[6], a[7])
		fmt.Printf("A row2 % .6f % .6f % .6f % .3f\n", a[8], a[9], a[10], a[11])
		det := float64(a[0])*(float64(a[5])*float64(a[10])-float64(a[6])*float64(a[9])) -
			float64(a[1])*(float64(a[4])*float64(a[10])-float64(a[6])*float64(a[8])) +
			float64(a[2])*(float64(a[4])*float64(a[9])-float64(a[5])*float64(a[8]))
		fmt.Printf("A det % .4f\n\n", det)
		base += n*48 - 4 // skip past this array; overlapping windows alias it
	}
	if found == 0 {
		return fmt.Errorf("no RAM array matched any frame of %s (%d nodes)", keyName, n)
	}
	return nil
}

// pose composes the clip's world matrices at frame f through the node tree —
// the same Rz·Ry·Rx + T composition the game's evaluator uses.
func pose(m *lm.MDL, key *lm.Key, f float32) []lm.Mtx34 {
	local := make([]lm.Mtx34, len(m.Nodes))
	for i := range m.Nodes {
		local[i] = srt34(key.Eval(i, f))
	}
	world := make([]lm.Mtx34, len(m.Nodes))
	var walk func(idx int, parent lm.Mtx34)
	walk = func(idx int, parent lm.Mtx34) {
		for idx >= 0 && idx < len(m.Nodes) {
			world[idx] = parent.Mul(local[idx])
			if m.Nodes[idx].Child >= 0 {
				walk(m.Nodes[idx].Child, world[idx])
			}
			idx = m.Nodes[idx].Sibling
		}
	}
	walk(0, lm.Identity34)
	return world
}

func srt34(p lm.Pose) lm.Mtx34 {
	sx, cx := sincos(p.Rot[0])
	sy, cy := sincos(p.Rot[1])
	sz, cz := sincos(p.Rot[2])
	return lm.Mtx34{
		cy * cz, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx, p.Translate[0],
		sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx, p.Translate[1],
		-sy, cy * sx, cy * cx, p.Translate[2],
	}
}

func sincos(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

// plausible screens one 3x4 candidate: finite floats, near-orthogonal unit
// rows in the 3x3 block, world-sized translation.
func plausible(ram []byte, off int) bool {
	var v [12]float32
	for i := range v {
		bits := binary.BigEndian.Uint32(ram[off+i*4:])
		f := math.Float32frombits(bits)
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return false
		}
		v[i] = f
	}
	for r := 0; r < 3; r++ {
		l := v[r*4]*v[r*4] + v[r*4+1]*v[r*4+1] + v[r*4+2]*v[r*4+2]
		if l < 0.25 || l > 4 {
			return false
		}
	}
	for _, t := range [3]float32{v[3], v[7], v[11]} {
		if t < -100000 || t > 100000 {
			return false
		}
	}
	return true
}

func readMtx(ram []byte, off int) lm.Mtx34 {
	var out lm.Mtx34
	for i := range out {
		out[i] = math.Float32frombits(binary.BigEndian.Uint32(ram[off+i*4:]))
	}
	return out
}

func maxDiff(a, b lm.Mtx34) float32 {
	var worst float32
	for i := range a {
		d := a[i] - b[i]
		if d < 0 {
			d = -d
		}
		// translations are thousands of units; weigh rotation cells up so a
		// 0.01 rotation error counts like a 10-unit translation error.
		if i%4 != 3 {
			d *= 1000
		}
		if d > worst {
			worst = d
		}
	}
	return worst
}

// invert34 inverts an affine 3x4 matrix (row-major, translation in col 3).
func invert34(m lm.Mtx34) lm.Mtx34 {
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[4], m[5], m[6]
	g, h, i := m[8], m[9], m[10]
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	id := 1 / det
	var r lm.Mtx34
	r[0] = (e*i - f*h) * id
	r[1] = (c*h - b*i) * id
	r[2] = (b*f - c*e) * id
	r[4] = (f*g - d*i) * id
	r[5] = (a*i - c*g) * id
	r[6] = (c*d - a*f) * id
	r[8] = (d*h - e*g) * id
	r[9] = (b*g - a*h) * id
	r[10] = (a*e - b*d) * id
	tx, ty, tz := m[3], m[7], m[11]
	r[3] = -(r[0]*tx + r[1]*ty + r[2]*tz)
	r[7] = -(r[4]*tx + r[5]*ty + r[6]*tz)
	r[11] = -(r[8]*tx + r[9]*ty + r[10]*tz)
	return r
}
