package track

// drawbridge.go reimplements $5A794 — the Draw Bridge animator. On track 5
// the two bridge ramps (sections 51/52 and 54/55, around the gap section 53)
// have no fixed height profile: the four sections' profile handles ($5F..$62)
// are overlapping 9-entry windows into ONE 36-entry table (at handle $5F,
// entries 2 bytes each), and $5A794 rewrites that table every race frame:
//
//	tri  = |((phase & $1F)) - 16|   (NOT for the negative side: 15..0,0..15)
//	step = (tri + 4) << 5           (per-rung rise, 128..608 height units)
//
// then entries 1..16 are filled with the accumulating ramp (1..15 x step,
// with entry 9 REPEATING entry 8's value — the shared joint rung between
// sections 51 and 52 — and entry 16 carrying the bit-7 edge marker on the
// lip), and each write is mirrored at entry 35-k, producing the second ramp
// descending; entries 0/17/18/35 stay zero (the ramp feet and the gap edge).
// The disk image only holds a placeholder pattern (a narrow spike per
// window) that the game never displays: even the pre-race preview runs one
// $5A794 pass first ($5D2DA), advancing the phase counter from 0 to 1
// (tri 14) — the "static ramps" the preview shows.
//
// Cadence (all traced): the race loop calls $5A794 once per frame ($5D49C);
// the phase advances unless the $EE time-base accumulator missed its carry
// ($5DB34: $1BBCF += $EE twice per frame, $1BBCD = carry of the last add —
// 18 frames in 256 skip), and freezes entirely while the player's or
// opponent's section is $33..$37 (a car on the bridge). One full up-down
// cycle is 32 phase steps, holding two steps at each extreme (phases 15/16
// and 31/0). The race loop itself is unthrottled (render-bound) — there is
// no VBlank or raster wait in the race path — so the absolute rate is
// machine-dependent. Verified byte-exact against the engine's own $5A794 by
// cmd/bridgeoracle.

// DrawBridgeTrack is the track id the animator is gated to ($5A79A).
const DrawBridgeTrack = 5

// U8 exposes a raw image byte read (run-time address) for the oracles.
func (im *Image) U8(a int) int { return im.u8(a) }

// DrawbridgeTri converts the phase counter to the triangle-wave step index
// (0..15), exactly as $5A824-$5A830 (SUBI then NOT on the negative side).
func DrawbridgeTri(phase int) int {
	v := (phase & 0x1F) - 16
	if v < 0 {
		v = ^v // NOT.w: -v-1 -> 15..0 for phase 0..15
	}
	return v
}

// Drawbridge returns a copy of the image with the bridge profile table
// rewritten as $5A794 leaves it for the given phase-counter value (the value
// of $1BBB0 AFTER its increment — the game patches with the incremented
// phase). For other tracks the patch is meaningless; callers gate on the id.
func (im *Image) Drawbridge(phase int) *Image {
	b := make([]byte, len(im.b))
	copy(b, im.b)
	out := &Image{b}

	tbl := handle(im.u16(shapeTab + 2*0x5F)) // the $BE index of $5A858
	step := (DrawbridgeTri(phase) + 4) << 5

	acc := 0
	write := func(entry, val int, mark bool) {
		hi := byte(val>>8) & 0x7F
		if mark {
			hi |= 0x80
		}
		out.b[tbl+2*entry-Base] = hi
		out.b[tbl+2*entry+1-Base] = byte(val)
	}
	entry := 1
	for i := 0; i < 15; i++ {
		acc += step
		write(entry, acc, entry == 16)
		write(35-entry, acc, entry == 16)
		entry++
		if entry == 9 { // the $12 quirk: repeat the joint rung's value
			write(entry, acc, false)
			write(35-entry, acc, false)
			entry++
		}
	}
	return out
}
