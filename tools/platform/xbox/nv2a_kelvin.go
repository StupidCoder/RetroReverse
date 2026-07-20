package xbox

// nv2a_kelvin.go implements the NV20_KELVIN_PRIMITIVE object (class 0x0097) — the Xbox's
// 3D graphics engine. It is the destination for the method stream the PFIFO pusher
// decodes on the subchannel the Direct3D runtime binds Kelvin to. Most methods latch
// state into the object register file (pgraph.Regs, indexed by method>>2); a few are
// triggers that run the pipeline (nv2a_pgraph.go dispatches those). The method numbers
// are the NV2A hardware's.
//
// This starts as the survey's graduation target: methods move here from the survey's
// "seen" set one at a time, each pinned from the live stream, so an unmodelled method
// halts and names itself instead of the engine drawing something plausible.

import (
	"fmt"
	"math"
	"os"
)

var nvSemTrace = os.Getenv("RR_NV_SEM") != ""
var nvVPTrace = os.Getenv("RR_NV_VP") != ""
var nvSurfTrace = os.Getenv("RR_NV_SURF") != ""
var shadowTrace = os.Getenv("RR_SHADOW") != ""
var lowWriteTrace = os.Getenv("RR_LOWWRITE") != ""

// RR_SHADOWCMP overrides the shadow-map depth-compare function (lt/le/gt/ge) — the
// derivation probe Part XV used to render the candidate frames against the caster
// geometry. Unset means the derived compare (shadowComparePass).
var shadowCmpEnv = os.Getenv("RR_SHADOWCMP")

// RR_SHADOWFRAG traces receiver fragments at the shadow sample: the interpolated
// oT3 s,t,r after the q divide against the stored 24-bit texel — the comparand
// evidence the compare derivation rests on.
var shadowFragTrace = os.Getenv("RR_SHADOWFRAG") != ""

// RR_TEXSHADER dumps the full 4-unit texture-shader config (stage modes, texture
// types, per-fragment texcoords) at the first fragment that would halt on an
// unmodelled cube/depth stage mode — so the exact reflection setup can be read
// before it is modelled. Investigation-only; not committed.
var texShaderTrace = os.Getenv("RR_TEXSHADER") != ""

// RR_NV_SERIAL forces the raster onto the serial path — the A/B control for
// verifying the parallel raster is byte-identical (it must always be).
var rasterSerial = os.Getenv("RR_NV_SERIAL") != ""

// flipVSyncDefault is the default for Machine.FlipVSync: model FLIP_STALL's vsync wait by
// advancing the guest clock a whole field per present, so OutRun's RDTSC catch-up steps its
// simulation every frame instead of every ~6th. ON by default — it is the correct behaviour;
// without it OutRun's sim runs at ~10 FPS and the debugger is unusably slow. RR_FLIP_VSYNC=0
// turns it off (the A/B control for measuring the old under-paced clock, never a mode to ship).
// See Machine.FlipVSync / creditFlipVBlank.
var flipVSyncDefault = os.Getenv("RR_FLIP_VSYNC") != "0"

// Kelvin methods with modelled side effects (NV2A method numbers).
const (
	kelvinCtxDmaSemaphore  = 0x01A4 // SET_CONTEXT_DMA_SEMAPHORE: DMA handle for the semaphore surface
	kelvinSemaphoreOffset  = 0x1D6C // SET_SEMAPHORE_OFFSET: byte offset within that surface
	kelvinSemaphoreRelease = 0x1D70 // BACK_END_WRITE_SEMAPHORE_RELEASE: write the value there

	// kelvinFlipStall is NV097_FLIP_STALL: the method Direct3D's Present compiles to. On
	// hardware it stalls the pusher until the flip the CRTC owes has retired; here the
	// pipeline is synchronous and there is nothing to wait for, so it is a pure marker —
	// but it is THE marker, because it is the title saying "this frame is finished and
	// meant for the screen". It is the machine's frame boundary (the debugger's OnFlip).
	kelvinFlipStall = 0x0130

	// SET_VIEWPORT_OFFSET / SET_VIEWPORT_SCALE (4 floats each). These do not just latch:
	// on the NV2A the viewport lives IN the transform-constant file, at the slots the
	// D3D-appended screen-space epilogue of every 3D vertex program reads — c59 (offset,
	// the added term) and c58 (scale, the multiplied term). Derived from the image alone,
	// no piece optional:
	//
	//   - the epilogue computes oPos.xyz = clip.xyz/w * c58.xyz + c59.xyz, so the
	//     multiplied constant is the scale and the added one the offset;
	//   - the game writes (±half-extent, zrange) vectors here — (320,-240,2^24-1,0) for
	//     the 640x480 pass, (64,-64,2^24-1,0) for the 128x128 reflection targets — and
	//     (center+bias) vectors to 0x0A20 — (320.53125, 240.53125, 0, 0) etc.;
	//   - across the whole reached gameplay window not one SET_TRANSFORM_CONSTANT load
	//     targets slots 56..60 (measured under RR_NV_VP, 1200 viewport-method writes,
	//     zero const loads), so this aliasing is the only mechanism by which the
	//     program's c58/c59 can ever hold the viewport the game configured.
	//
	// The pre-transformed 2D passes are untouched: their programs carry their own
	// viewport in c0/c1 (explicit const loads) and never reference c58/c59.
	kelvinViewportOffset = 0x0A20 // ..0x0A2C → c59
	kelvinViewportScale  = 0x0AF0 // ..0x0AFC → c58

	vshSlotViewportScale  = 58
	vshSlotViewportOffset = 59
)

// kelvinMethod handles one method write to the bound Kelvin (3D) object. Every method
// latches into the register file; the cases below are the ones with modelled side
// effects (triggers and FIFO-shaped data ports). A method with no case is a plain
// state latch the pipeline reads back at draw time — or an unmodelled one, which the
// unhandled map records so a run's reach through the 3D command surface stays a
// concrete statement.
func (g *pgraph) kelvinMethod(method, arg uint32) {
	// THE FRAME BOUNDARY. FLIP_STALL is what Direct3D's Present compiles to, and it is the
	// only thing in the stream that means "this frame is finished and meant for the
	// screen". The hook fires before the latch below, while the colour surface still names
	// the buffer the frame was built in — so a hook that renders sees the completed frame.
	//
	// Every plausible alternative is wrong, and each was measured rather than reasoned:
	//
	//   - AvSetDisplayMode, the kernel call that registers the scanout, is called ONCE per
	//     boot (measured: one call in 340M instructions, against thousands of frames). It
	//     reads like a swap — the D3D swap path is where it is called from — but it is a
	//     mode set.
	//   - BACK_END_WRITE_SEMAPHORE_RELEASE, D3D's fence, fires TWICE per frame (odd and
	//     ascending by 2), so it would report two frames for every real one.
	//   - SET_SURFACE_COLOR_OFFSET moving to the next buffer of the swap chain is the
	//     seductive one, and it is right often enough to fool a test: at the logo phase it
	//     fires exactly once per frame, in lockstep with FLIP_STALL (209 and 209). At the
	//     TITLE screen it fires three times per frame — the title renders its movie into an
	//     off-screen target first (269 re-points to 02B7B200, one per frame) and then twice
	//     more into the back buffer. A capture bounded by it there ends mid-frame, on a
	//     buffer nothing has drawn into yet, and reports a blank white frame that looks
	//     entirely plausible.
	//   - The vertical blank is a 60 Hz scanout clock that ticks whether or not the title
	//     drew: a field, not a frame.
	//
	// Measured, once per frame at BOTH fixtures: 209 flips at the logo, 269 at the title.
	if method == kelvinFlipStall {
		// Model the vsync wait: on hardware Present stalls until the field the CRTC owes
		// retires, and the guest's RDTSC accrues that whole field. Advance the tick to the
		// next vblank so the game's fixed-timestep loop sees one field per present. Just move
		// the clock — the normal schedTick delivers the vblank interrupt at an instruction
		// boundary; raising it here (mid-store) would run the ISR reentrantly. Off unless
		// FlipVSync (a fidelity change; the default path is byte-identical).
		if g.m.FlipVSync {
			g.m.creditFlipVBlank()
		}
		g.recordPresented()
		// Close the frame's profile BEFORE the debugger's flip hook runs, so the hook (which
		// reads FrameProfile) sees the completed frame and a debugger pausing in it does not
		// bank the pause as x86. profFrame splits the run and pusher timers in flight here —
		// the flip is mid-drain (profile.go).
		g.m.profFrame()
		if g.m.OnFlip != nil {
			g.m.OnFlip(g.m)
		}
	}
	if method < uint32(len(g.Regs))*4 {
		g.Regs[method>>2] = arg
	}
	if nvVPTrace && (method >= 0x0A20 && method < 0x0A30 || method >= 0x0AF0 && method < 0x0B00) {
		fmt.Printf("VP method %04X = %08X (%g) draws=%d\n", method, arg, math.Float32frombits(arg), g.Draws)
	}
	if nvSurfTrace && method >= 0x0200 && method <= 0x0214 {
		fmt.Printf("SURF method %04X = %08X draws=%d\n", method, arg, g.Draws)
	}
	if shadowTrace {
		switch method {
		case kelvinFlipStall:
			fmt.Printf("SHADOW FLIP draws=%d tid=%d tick=%d\n", g.Draws, g.m.threadID(), g.m.tick)
		case kelvinSurfaceColorOffset:
			fmt.Printf("SHADOW COLOR=%08X draws=%d\n", arg, g.Draws)
		case kelvinSurfaceZetaOffset:
			fmt.Printf("SHADOW ZETA=%08X draws=%d\n", arg, g.Draws)
		case kelvinTexOffset, kelvinTexOffset + 0x40, kelvinTexOffset + 0x80, kelvinTexOffset + 0xC0:
			u := (method - kelvinTexOffset) / 0x40
			fmt.Printf("SHADOW TEXOFF unit=%d off=%08X draws=%d\n", u, arg, g.Draws)
		}
	}
	if nvSemTrace {
		switch method {
		case 0x0100, 0x0110, 0x0130, 0x1D6C, 0x1D70, 0x17D0:
			fmt.Printf("KELVIN sync: method %04X arg %08X (steps=%d)\n", method, arg, g.m.CPU.Steps)
		}
	}
	switch {
	case method == kelvinClearSurface:
		// CLEAR_SURFACE: fill the clear rect of the color/zeta surfaces with the
		// latched clear values (nv2a_frame.go) — the first Kelvin method that
		// produced pixels.
		tc := g.m.profStart()
		g.clearSurface(arg)
		g.m.profEnd(bucketClear, tc)
		return
	case method == kelvinSemaphoreRelease:
		// BACK_END_WRITE_SEMAPHORE_RELEASE: the back end writes the release value into
		// the bound semaphore surface at the latched offset — this is how the Direct3D
		// runtime observes GPU progress (its per-frame sync values arrive here, odd and
		// ascending by 2). The D3D busy-wait at 0x1AE550 additionally polls PGRAPH
		// register 0x400B10, comparing its bits 2..6 against the semaphore value <<2 —
		// the register mirrors the back end's semaphore progress. Our pipeline is
		// synchronous, so both update the instant the method executes: the release has
		// retired by the time the CPU can look.
		base, limit := g.m.dmaObjectTarget(g.Regs[kelvinCtxDmaSemaphore>>2])
		off := g.Regs[kelvinSemaphoreOffset>>2]
		if base != 0 && off <= limit {
			g.m.write32(base+off, arg)
		}
		g.m.nv.reg[nvPGRAPH_SEMAPHORE>>2] = arg << 2
		return

	// --- the vertex front end (nv2a_vertex.go) ---
	case method == kelvinBeginEnd:
		g.rastValid = false // the batch's raster state is decoded fresh per BEGIN/END
		g.beginEnd(arg)
		return
	case method == kelvinInlineArray:
		g.inline = append(g.inline, arg)
		return
	case method == kelvinElement16:
		g.elems = append(g.elems, arg&0xFFFF, arg>>16)
		return
	case method == kelvinElement32:
		g.elems = append(g.elems, arg)
		return
	case method == kelvinDrawArrays:
		g.ranges = append(g.ranges, [2]uint32{arg & 0xFFFFFF, (arg >> 24) + 1})
		return
	case method >= kelvinVertexData4C && method < kelvinVertexData4C+0x40:
		// SET_VERTEX_DATA4UB: a persistent attribute value as 4 unsigned bytes
		// (RGBA byte order), read by every vertex whose arrays do not supply it.
		i := (method - kelvinVertexData4C) >> 2
		g.vtxAttr[i] = [4]float32{
			float32(arg&0xFF) / 255, float32(arg>>8&0xFF) / 255,
			float32(arg>>16&0xFF) / 255, float32(arg>>24&0xFF) / 255,
		}
		return

	// --- the transform program (nv2a_vsh.go) ---
	case method >= kelvinProgData && method < kelvinProgData+0x80:
		g.progData(arg)
		return
	case method >= kelvinConstData && method < kelvinConstData+0x80:
		g.constData(arg)
		return
	case method == kelvinProgLoad:
		g.ProgLoad = arg
		g.progBufN = 0
		return
	case method == kelvinConstLoad:
		g.ConstLoad = arg
		g.constBufN = 0
		return

	// --- the viewport, aliased into the transform-constant file (see the method
	// constants above for the derivation) ---
	case method >= kelvinViewportOffset && method < kelvinViewportOffset+0x10:
		g.Const[vshSlotViewportOffset][(method-kelvinViewportOffset)>>2] = arg
		return
	case method >= kelvinViewportScale && method < kelvinViewportScale+0x10:
		g.Const[vshSlotViewportScale][(method-kelvinViewportScale)>>2] = arg
		return
	}
	// Everything else is a plain state latch (the raster/texture/combiner registers
	// the pipeline reads at draw time) or an unmodelled method; the survey/unhandled
	// map is what states the frontier.
	g.unhandled[classKelvin<<16|(method&0xFFFF)]++
}
