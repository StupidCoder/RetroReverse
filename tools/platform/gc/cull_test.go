package gc

import (
	"image"
	"os"
	"testing"
)

// Back-face culling's failure mode is not a crash and not a slowdown — it is a scene that
// quietly loses its surfaces, or renders inside-out, while every counter says the change
// worked. Both sign conventions cull about 8,000 of this field's triangles and both make the
// rasteriser about 30% faster, because in closed geometry half the triangles face each way.
// So the tests here ask the PICTURE, which is the only thing that can tell them apart.

// renderCutsceneField draws one field of the intro cutscene under the given RR_GC_CULLMODE
// behaviour and returns the finished EFB. mode is passed directly rather than through the
// environment, because cullExperiment is read once at init.
func renderCutsceneField(t *testing.T, mode string) *image.RGBA {
	t.Helper()
	if _, err := os.Stat(discPath); err != nil {
		t.Skip("the disc image is not present")
	}
	if _, err := os.Stat(cutscenePath); err != nil {
		t.Skip("the cutscene savestate is not present (work/ is regenerable scratch)")
	}
	old := cullExperiment
	cullExperiment = mode
	t.Cleanup(func() { cullExperiment = old })

	d, err := Open(discPath)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	m, err := NewMachine(d)
	if err != nil {
		t.Fatal(err)
	}
	m.SetSpinDetect(false)
	if err := m.LoadStateFile(cutscenePath); err != nil {
		t.Fatal(err)
	}
	var shot *image.RGBA
	flips := 0
	m.OnFlip = func(mm *Machine) {
		flips++
		if flips == 2 {
			shot, _ = mm.RenderEFB()
			mm.StopRequested = true
		}
	}
	m.Run(400_000_000)
	if shot == nil {
		t.Fatal("the field never flipped")
	}
	return shot
}

// pixelsDiffering counts pixels whose colour differs by more than tol in any channel.
func pixelsDiffering(a, b *image.RGBA, tol int) (n, total int) {
	for i := 0; i < len(a.Pix) && i < len(b.Pix); i += 4 {
		total++
		for c := 0; c < 3; c++ {
			d := int(a.Pix[i+c]) - int(b.Pix[i+c])
			if d < 0 {
				d = -d
			}
			if d > tol {
				n++
				break
			}
		}
	}
	return n, total
}

// TestCullingDoesNotChangeTheOpaqueScene pins the sign convention.
//
// Culling removes triangles the depth buffer was going to hide anyway, so the opaque scene
// must come out of it essentially unchanged. If the sign is inverted the rasteriser culls the
// FRONT faces instead and the picture is gutted — you see through the tree trunks and into the
// back of Luigi's head — while the profiler happily reports that the rasteriser got faster.
//
// Measured when the convention was chosen: the correct sign changes 9.4% of the frame, the
// inverted one 92.7%. The threshold sits between those, far from both.
func TestCullingDoesNotChangeTheOpaqueScene(t *testing.T) {
	uncalled := renderCutsceneField(t, "off")
	culled := renderCutsceneField(t, "")

	n, total := pixelsDiffering(uncalled, culled, 16)
	pct := 100 * float64(n) / float64(total)
	if pct > 25 {
		t.Errorf("culling changed %.1f%% of the frame (%d/%d pixels) — at that scale it is removing "+
			"geometry the player can see, which means the winding sign in cullTest is inverted",
			pct, n, total)
	}
	// And it must do SOMETHING: a cullTest that never fires would pass the check above
	// trivially while leaving the rasteriser doing all the work it always did.
	if pct == 0 {
		t.Error("culling changed nothing at all; the cull test is not firing")
	}
}

// TestCullingTheOtherWayRoundIsVisiblyWrong is the other half of the same claim, and it is
// what makes the test above mean something. Without it, a cullTest that culled nothing would
// pass "the scene is unchanged" perfectly.
func TestCullingTheOtherWayRoundIsVisiblyWrong(t *testing.T) {
	uncalled := renderCutsceneField(t, "off")
	inverted := renderCutsceneField(t, "flip")

	n, total := pixelsDiffering(uncalled, inverted, 16)
	pct := 100 * float64(n) / float64(total)
	if pct < 50 {
		t.Errorf("inverting the cull sign changed only %.1f%% of the frame; it should gut the scene "+
			"(~92%%). Either the two conventions are no longer opposites, or culling has stopped working", pct)
	}
}

// TestCullingActuallyCulls: the register is read, the modes are honoured, and the field's
// triangles are genuinely being discarded rather than the test above passing because nothing
// happens.
func TestCullingActuallyCulls(t *testing.T) {
	if _, err := os.Stat(cutscenePath); err != nil {
		t.Skip("the cutscene savestate is not present")
	}
	d, err := Open(discPath)
	if err != nil {
		t.Skip("the disc image is not present")
	}
	defer d.Close()
	m, _ := NewMachine(d)
	m.SetSpinDetect(false)
	if err := m.LoadStateFile(cutscenePath); err != nil {
		t.Fatal(err)
	}
	m.SetProfile(true)
	// Every distinct cull mode the field programs, so a change to the field decode shows up
	// as a mode nobody uses rather than as a silently wrong picture.
	modes := map[uint32]int{}
	m.OnGXCmd = func(mm *Machine, op uint8, w []uint32) {
		if op == 0x61 && len(w) > 0 && uint8(w[0]>>24) == 0x00 {
			modes[(w[0]>>14)&3]++
		}
	}
	flips := 0
	m.OnFlip = func(mm *Machine) {
		flips++
		if flips >= 2 {
			mm.StopRequested = true
		}
	}
	m.Run(400_000_000)

	if modes[cullNone] == 0 || modes[cullNegArea] == 0 || modes[cullPosArea] == 0 {
		t.Errorf("the field no longer programs all three of the modes it used to: %v", modes)
	}
	if modes[cullAll] != 0 {
		t.Errorf("the field programs cull-everything %d times, which it never did before — "+
			"the field decode is probably reading the wrong bits", modes[cullAll])
	}
	// Mode 2 dominating is what makes it (very probably) the back-face cull. Not relied on
	// for correctness — the sign is pinned by the picture — but a reversal here would mean
	// the field decode moved, so it is worth noticing.
	if modes[cullPosArea] < modes[cullNegArea] {
		t.Errorf("mode 1 now outnumbers mode 2 (%d vs %d); the field decode has probably shifted",
			modes[cullNegArea], modes[cullPosArea])
	}
	var culled int
	for _, c := range m.FrameProfile().Counters {
		if c.Name == "tris culled" {
			culled = c.Value
		}
	}
	if culled < 1000 {
		t.Errorf("only %d triangles were culled in a field of ~9,800 draws; culling is not working", culled)
	}
}
