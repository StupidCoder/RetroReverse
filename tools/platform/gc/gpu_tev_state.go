package gc

// gpu_tev_state.go decodes the TEV's configuration once per draw, where it stops being
// per-anything.
//
// THE ARGUMENT THAT MAKES THIS BIT-EXACT, and it is the same one every time: nothing the TEV
// reads can change between the fragments of a draw, because a BP register write only happens
// in the FIFO command processor, and the command processor is not running while a draw is.
// drawPrimitive is called from inside the FIFO interpreter's own loop; the Gekko is not
// running either. So the values below are constants for the whole draw, and hoisting them
// changes when they are computed, not what they are.
//
// What was being re-derived for every one of ~1.4 million fragments a field: the four working
// registers' seeds (four eleven-bit-field decodes each), the stage count, and then per stage —
// up to sixteen of them — the ordering register and its odd-stage shift, both combiner
// registers, TWO swap-table reads (four BP registers between them), and the constant colour
// and alpha, each of which re-read a KSEL register and could decode a whole colour register
// again on top. None of it depends on the pixel.
//
// This is the third time this repository has made the same trick pay (the 3DS's shader
// instructions, the 3DS's TEV, and now this one), and it is worth saying that IT DOES NOT
// GENERALISE: the same hoist applied to the 3DS's vertex attribute FETCH measured 1.7%
// slower, because the format decode there was a few shifts and the indirection cost more than
// it saved. The reason it pays here is the arithmetic per fragment is large and the state is
// small, not because hoisting is good.
//
// WHAT IS DELIBERATELY NOT HOISTED. The swap TABLES are per-draw; the swizzled VALUES are not.
// gpu_tev.go is explicit that writing a swizzle back over the loop's own ras would hand the
// next stage a colour the rasteriser never produced, so sras and tex stay per-fragment locals
// and only the table they are looked up through moves here. Likewise the alpha combiner's
// compare mode reads the COLOUR pipe's operands, which are per-fragment by construction; that
// coupling is untouched.

// tevStage is one stage's decoded configuration — everything about it that the pixel cannot
// change.
type tevStage struct {
	cc, ac           uint32 // the colour and alpha combiner registers, verbatim
	texmap, texcoord int
	texEnable        bool
	rasSel           uint32 // which colour channel this stage's RAS input reads
	swapRas, swapTex [4]int
	konstC           [3]float32
	konstA           float32
	cdest, adest     uint32
}

// tevState is the whole combiner pipeline as one draw sees it.
type tevState struct {
	numStages int
	stages    [16]tevStage
	seed      [4][4]float32 // the working registers' starting values

	// The texture maps the enabled stages name, decoded once. sampleTexmap used to rebuild
	// one of these from four BP registers and a dozen shifts FOR EVERY TEXEL — every sample
	// of every stage of every fragment. Only the maps a stage actually names are filled;
	// texValid says which, so a stage naming an unbound map is still the game's own error
	// and not a phantom this decode invented.
	tex      [8]texState
	texValid [8]bool

	// canHalt says whether sampling this draw's textures could reach a CPU.Halt — an unknown
	// texture format, or a paletted one with an unknown palette format. It is asked here,
	// serially, so that the fill can refuse to fan out a draw that might halt: a worker
	// goroutine must not halt the machine, and the alternative (teaching the whole fragment
	// path to defer a halt) would put an error check in the hottest loop for a case that
	// never happens in a working scene.
	//
	// The halt then lands on the machine's own goroutine, at exactly the fragment it always
	// did — this decides HOW a draw runs, never WHETHER it faults.
	canHalt bool
}

// texCanHalt reports whether sampling this map could reach a CPU.Halt.
//
// It mirrors decodeTexel's format switch and tlutColor's palette switch, which is a
// duplication and therefore a liability: the two must agree or a worker can halt. Its
// direction of failure is safe — a format this does not recognise is treated as "could halt"
// and merely costs the draw its parallelism — but agreement is pinned by
// TestTexCanHaltAgreesWithDecoder, which asks the decoder itself, for every format and every
// palette format, rather than trusting this list to have been kept up to date.
func texCanHalt(tx *texState) bool {
	switch tx.format {
	case texI4, texI8, texIA4, texIA8, texRGB565, texRGB5A3, texRGBA8, texCMPR:
		return false
	case texC4, texC8, texC14X2:
		// A paletted format reaches tlutColor, which knows three entry formats.
		return tx.tlutFmt > 2
	default:
		return true
	}
}

// tevstate decodes the TEV as the registers currently stand. Call it once per draw.
func (g *gpu) tevstate() tevState {
	var t tevState

	for i := 0; i < 4; i++ {
		r, gg, b, a := tevColorReg(g.TevColorReg[i])
		t.seed[i] = [4]float32{r, gg, b, a}
	}

	t.numStages = int((g.BP[0x00]>>10)&0xF) + 1
	for s := 0; s < t.numStages; s++ {
		st := &t.stages[s]

		ord := g.BP[0x28+uint32(s/2)]
		if s&1 == 1 {
			ord >>= 12
		}
		st.texmap = int(ord & 7)
		st.texcoord = int((ord >> 3) & 7)
		st.texEnable = (ord>>6)&1 != 0

		// Which colour channel this stage rasterises. A stage does not have to read the
		// channel the vertex was lit into, and in this game's scene program most of them do
		// not: its five-stage room shader reads NULL in stages 0, 1 and 4 (they are pure
		// texture stages), channel 0 in stage 2, and channel 1 in stage 3. Leaving this field
		// unread and handing every stage channel 0 is not a small error — the three NULL
		// stages get a colour where the hardware gives them zero, and the specular stage gets
		// the diffuse channel.
		st.rasSel = (ord >> 7) & 7

		st.cc = g.BP[0xC0+uint32(s)*2]
		st.ac = g.BP[0xC1+uint32(s)*2]

		// Both tables are read whether or not the stage samples a texture. The old code
		// skipped the texture swap when texEnable was clear, which saved two register reads
		// on a stage that then did not use the result — here it is two reads per draw, and
		// the branch costs more than it saves.
		st.swapRas = g.swapTable(int(st.ac & 3))
		st.swapTex = g.swapTable(int((st.ac >> 2) & 3))

		st.konstC = g.konstColor(s)
		st.konstA = g.konstAlpha(s)

		st.cdest = (st.cc >> 22) & 3
		st.adest = (st.ac >> 22) & 3

		// The map this stage samples. Two stages naming the same map decode it once.
		if st.texEnable && !t.texValid[st.texmap] {
			t.tex[st.texmap] = g.texSetup(st.texmap)
			t.texValid[st.texmap] = true
			if texCanHalt(&t.tex[st.texmap]) {
				t.canHalt = true
			}
		}
	}
	return t
}
