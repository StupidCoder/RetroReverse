package rr

// flight.go decodes the two flying objects' movement data from the EXE.
//
// The HELICOPTER (objects 188 + 189) is flown by a waypoint bytecode
// interpreter (dispatcher 0x800380A0, jump table 0x800107FC, flight struct
// 0x801DB8FC). Two route scripts exist — 0x800734D0 and 0x80073790 — and the
// end-of-script opcode 9 restarts with one of them picked at random (pointer
// pair at 0x8007A10C). The opcodes, each a run of little-endian halfwords:
//
//	0  X, Y, Z, pitch, yaw, roll      teleport + set attitude (X/Z unsigned)
//	1  X, Y, Z, gateS, gateL, dur     fly to waypoint; the gate halfword
//	                                  (short-/long-course variant, <<8) holds
//	                                  the segment until race progress passes
//	                                  it, dur is the flight time in frames
//	2  dur                            hover in place for dur frames
//	3|4|5  angle                      set pitch|yaw|roll (current and target)
//	6|7|8  target, rate               seek pitch|yaw|roll target at rate/frame
//	9  —                              restart with a random route
//
// The drawer (0x80038A38) applies pitch = angle0 − 1024 — the model is
// authored nose-up and flown pitched −90° — then roll, then yaw, and spins the
// rotor (189) 331/frame on top.
//
// The AIRPLANE (object 190 day / 251 night, drawer 0x80038FE4) has no script:
// its updater (0x80038C74) spawns it at the position vector 0x800739E4 when
// the race passes a progress gate, then integrates pos += (dir·200)>>16 each
// frame along the 16.16 direction vector at 0x800739F4 while its age counter
// runs 1..1800. Positions are quarter model units wrapping at 2¹⁶ (the world
// is a u16 torus; SetTransform subtracts the camera as halfwords).

import "fmt"

// HeliKey is one flight-path key: a waypoint position in model units, the
// flight time to reach it, an extra hover after arriving, and the yaw the
// script last commanded (raw units, 4096 = one turn).
type HeliKey struct {
	X, Y, Z int32
	Dur     int // frames flying to this key (30 fps race ticks)
	Hold    int // frames hovering after arrival
	Yaw     int16
}

// HeliScript is one decoded route: the teleport start pose and the waypoints.
type HeliScript struct {
	Addr    uint32
	X, Y, Z int32 // start position, model units
	Yaw     int16
	Keys    []HeliKey
}

// heliScriptAddrs are the two routes the op-9 restart table (0x8007A10C)
// points at; the second is also the initial route.
var heliScriptAddrs = []uint32{0x800734D0, 0x80073790}

// HeliScripts decodes both helicopter routes from the executable image.
func HeliScripts(text []byte) ([]HeliScript, error) {
	var out []HeliScript
	for _, addr := range heliScriptAddrs {
		s, err := decodeHeliScript(text, addr)
		if err != nil {
			return nil, fmt.Errorf("heli script %08X: %v", addr, err)
		}
		out = append(out, s)
	}
	return out, nil
}

func decodeHeliScript(text []byte, addr uint32) (HeliScript, error) {
	s := HeliScript{Addr: addr}
	off := int(addr - exeTextBase)
	lhu := func() int32 { v := int32(uint16(s16(text, off))); off += 2; return v }
	lh := func() int32 { v := int32(s16(text, off)); off += 2; return v }
	var yaw int16
	for n := 0; ; n++ {
		if n > 200 {
			return s, fmt.Errorf("no terminator after %d ops", n)
		}
		op := lh()
		switch op {
		case 0:
			s.X, s.Y, s.Z = lhu()*4, lh()*4, lhu()*4
			lh() // pitch
			yaw = int16(lh())
			lh() // roll
			s.Yaw = yaw
		case 1:
			k := HeliKey{X: lhu() * 4, Y: lh() * 4, Z: lhu() * 4}
			lh() // progress gate, short courses
			lh() // progress gate, long courses
			k.Dur = int(lh())
			k.Yaw = yaw
			s.Keys = append(s.Keys, k)
		case 2:
			hold := int(lh())
			if len(s.Keys) > 0 {
				s.Keys[len(s.Keys)-1].Hold += hold
			}
		case 3, 5:
			lh()
		case 4:
			yaw = int16(lh())
		case 6, 8:
			lh()
			lh()
		case 7:
			yaw = int16(lh())
			lh() // rate
		case 9:
			return s, nil
		default:
			return s, fmt.Errorf("unknown opcode %d at %08X", op, exeTextBase+uint32(off)-2)
		}
	}
}

// PlanePath is the airplane's linear flight: spawn position, per-frame delta
// (both model units; the game wraps quarter units at 2¹⁶) and lifetime.
type PlanePath struct {
	X, Y, Z    int32
	DX, DY, DZ int32
	Frames     int
}

// Airplane decodes the airplane's path: spawn vector at 0x800739E4, 16.16
// direction at 0x800739F4, speed 200, alive for 1800 frames.
func Airplane(text []byte) PlanePath {
	const speed = 200
	pos := int(uint32(0x800739E4) - exeTextBase)
	dir := int(uint32(0x800739F4) - exeTextBase)
	step := func(d int32) int32 {
		return int32(int64(d)*speed/65536) * 4 // trunc toward zero, ×4 model units
	}
	return PlanePath{
		X: int32(u32(text, pos)) * 4, Y: int32(u32(text, pos+4)) * 4, Z: int32(u32(text, pos+8)) * 4,
		DX: step(int32(u32(text, dir))), DY: step(int32(u32(text, dir+4))), DZ: step(int32(u32(text, dir+8))),
		Frames: 1800,
	}
}
