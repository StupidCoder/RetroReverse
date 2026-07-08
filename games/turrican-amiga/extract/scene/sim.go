package scene

// sim.go — resolve each placed object's *displayed* sprite/frame/position by RUNNING
// its AI handler's spawn-init in a 68000 interpreter, exactly as the engine does, rather
// than guessing. This is the same "run the code, don't reimplement it" approach that
// nailed the music driver.
//
// The spawner (enemy_spawner $1710) allocates a zeroed object node, writes the placement
// low byte into node+$1E (the orientation/direction selector) and the AI handler into
// node+$22, then object_update ($1B94) calls the handler with a5 = node. On that first
// call the handler runs its init path (node+$1E has bit 7 set = "just spawned"): it picks
// a frame table (node+$12), a frame index (node+$C) and adjusts the position (node+$18/$1A)
// according to the orientation. We reproduce exactly that, then read those fields back.
//
// Isolation: one machine per world, but before each object we restore the engine + block
// image (undo any scribbles), re-run the pool init so obj_alloc ($108 -> $1ABA, which some
// init paths call) hands out valid scratch nodes, zero the node and reset the few globals
// the init paths read (camera/scroll/player = origin). A handler that faults or produces an
// implausible result is left with its static (frame-0, raw-position) values.

import "retroreverse.com/games/turrican-amiga/extract/m68k"

const (
	nodeAddr    = 0x80000   // scratch object node (above every scene block, below the stack)
	objPoolInit = 0x1A68    // the engine's obj_pool_init (builds the free list obj_alloc pops)
	simMaxInsn  = 4_000_000 // per-init instruction cap (inits are short; this is a runaway guard)
)

// Sim runs object-handler inits for one world.
type Sim struct {
	m     *m68k.Machine
	res   []byte // full resident image; res[:residHi] is the relocated engine
	block []byte // this world's scene block
	w     int
}

// NewSim builds a simulator for world w: the relocated engine at address 0 and the scene
// block at $1B980, with the object pool initialised.
func (g *Game) NewSim(w int) *Sim {
	s := &Sim{m: m68k.New(), res: g.Resident.Data, block: g.blocks[w], w: w}
	s.reset()
	return s
}

// posGuard: the sim's position is trusted only when it stays within this many pixels of
// the raw placement. Legit init adjustments (grid-snaps, small offsets) are tens of
// pixels; a handful of scrolling-scene flyers hard-code a *screen* position (e.g.
// node.x=$150), which with an origin camera would teleport the object across the level.
// Those keep the raw placement position — while still taking the sim's correct sprite/frame.
const posGuard = 80

// reset restores a clean memory image and object pool, with the camera/scroll/player at
// the origin so grid-snaps and small offsets are deterministic.
func (s *Sim) reset() {
	s.m.Load(0, s.res[:residHi]) // relocated engine ($10..$1B780)
	s.m.Load(BlockBase, s.block) // scene block at $1B980
	for _, a := range []uint32{0x104, 0x106, 0x120, 0x122, 0x158, 0x15A, 0x172, 0x174} {
		s.m.W16(a, 0)
	}
	s.m.W32(0x236, 0) // active list head
	s.m.W16(0x14A, 0) // active count
	s.m.Call(objPoolInit, nil, nil, 200_000)
}

// Resolve runs o's handler init and, on success, overwrites o.X/Y/FT/Frame/Handler with
// the values the engine would display, and marks o.Simmed. A fault or implausible result
// leaves o unchanged (its static frame-0 / raw-position values stand).
func (g *Game) Resolve(s *Sim, o *Object) {
	if o.Handler == 0 {
		return
	}
	rawX, rawY := o.X, o.Y
	s.reset()
	m := s.m
	for a := uint32(nodeAddr); a < nodeAddr+0x60; a += 2 {
		m.W16(a, 0)
	}
	m.W16(nodeAddr+0x18, uint16(int16(o.X)))
	m.W16(nodeAddr+0x1A, uint16(int16(o.Y)))
	m.W8(nodeAddr+0x1E, uint8(o.Orient))
	m.W32(nodeAddr+0x22, uint32(o.Handler))

	if err := m.Call(uint32(o.Handler), map[int]uint32{m68k.A5: nodeAddr, m68k.A6: 0xDFF000}, nil, simMaxInsn); err != nil {
		return
	}

	ft := int(m.R32(nodeAddr + 0x12))
	frame := int(int16(m.R16(nodeAddr + 0x0C)))
	x := int(int16(m.R16(nodeAddr + 0x18)))
	y := int(int16(m.R16(nodeAddr + 0x1A)))
	ai := int(m.R32(nodeAddr + 0x22))

	resident := ft < BlockBase
	if _, ok := g.FrameAt(s.w, ft, frame, resident); !ok {
		return // implausible sprite — keep the static resolution
	}
	// Trust the sprite/frame always; trust the position only if it stayed near the
	// placement (else a screen-hard-coding flyer would jump across the level).
	if abs(x-rawX)+abs(y-rawY) <= posGuard {
		o.X, o.Y = x, y
	}
	o.FT, o.Frame, o.Resident = ft, frame, resident
	if ai != 0 {
		o.Handler = ai // the steady-state AI the init installed
	}
	o.Simmed = true
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
