package rr

// dynamics.go decodes the course's dynamic objects: everything the scenery
// dispatcher at 0x80015B98 (and the two flight controllers at 0x800380A0 /
// 0x80038C48) draws from dedicated code rather than a placement table. Each
// block was traced in the disassembly: the object ids are immediates in the
// code, and every fixed position is an initialized 16-byte vector in the
// executable's data segment (X, Y, Z, W words in quarter model units; W is a
// yaw for the flyer spawns, else 0). The decoder reads those vectors from the
// EXE bytes — nothing here is captured from the running game; the oracle only
// verifies (geomoracle -only dynamics).
//
// The traced blocks, in dispatcher order (gates in parentheses; "night" is the
// non-zero state of the halfword at 0x80176BF8 — the higher race classes run
// at dusk/night and swap models):
//
//	0x80015CB0  start-gate far LOD: obj 67 + wave frames 68..83 ((frame>>2)&15)
//	            at 0x80070810, yaw 0x874 (night only — the day twin is below)
//	0x80015DB4  hot-air balloon obj 139 at 0x80070850, rises 256/frame
//	            (night, flag bit 0x20 at 0x80080444)
//	0x80015EEC  hot-air balloon obj 139 at 0x80070860, rises 544/frame
//	            (night, flag bit 0x40)
//	0x80016018  number girl obj 250 at 0x80070870: upright (pitch 0x400) then
//	            yaw = word@+12 + 0x400; drawn through the translucent path
//	0x80016128  rotating sign obj 175 at 0x80070720, yaw spins 16/frame
//	            (reversed at night)
//	0x8001620C  palette-cycled beacon obj 257 at 0x80070820, unrotated, CLUT
//	            offset cycles four palettes every four frames
//	0x800162F4  start-gate day twin at 0x800707A0: far LOD 67 + 68..83, or the
//	            near-LOD composite (camera word 0x80130CC4 < 900): base 146 +
//	            145/152/153 + shadow 164 + semaphore 155,154+n (n = column 5 of
//	            the rig table 0x80056C44, row = state 0x80130F54) + crowd wave
//	            147..151 (98 + byte sequence at 0x800100F4, (frame>>2)&15) +
//	            the camera-crane rig (0x800154B8): part ids from the same rig
//	            table row, pivot 0x800707F0, one joint spinning frame<<4
//	0x8001694C  start balloons: frames 166..173 + lap-banner 279+8·bank+frame
//	            (bank byte table 0x80079FA0 by lap), spawned from 0x80070830
//	            (day) / 0x80070840 (night, W = yaw), drifting -X and up with
//	            race progress
//	0x80016C38  start banner obj 182 (181 on a scheduled flicker frame) at
//	            0x80070740 yaw 0x2CC (night) / 0x80016D34 at 0x80070730 yaw
//	            0x400 (day)
//	0x80016E30  big screen A objs 183/184 (on, alternating every 4 frames) or
//	            185 (off) at 0x80070750, yaw 0xAC0; the drawer re-points the
//	            screen quad's UVs through the OBJ.RRO directory each frame
//	0x80016FBC  big screen B, same objs, at 0x80070760, yaw 0xB20
//	0x80038A38  helicopter: body 188 + rotor 189 (spun 331/frame), flown by a
//	            waypoint bytecode script at 0x80073790 (jump table 0x800107FC)
//	0x80038C48  airplane: obj 190 (day) / 251 (night), pitch -192, yaw 0xF00,
//	            spawned at race start (struct 0x801D6DBC), position integrated
//	            along the EXE direction vector at 0x800739F4 for 1800 frames
//	0x80015B98  swinging tunnel sign obj 192 at 0x80070780, yaw 0xC00 ±
//	            sin(frame)/8 — the hanging counterpart of placard 193 (the
//	            placards themselves are ordinary table placements, 0x8006E85C)
//
// Placard physics note: the placard table 0x8006E85C is read by the draw
// iterator and SetTransform only — watched over a full attract lap and a
// scripted drive through a placard moved onto the racing line, it is never
// written and has no other reader. In this build the placards are draw-only
// scenery; nothing detects the car hitting them.

// Dynamic is one dynamic (code-placed) object: the ids it is drawn with (base
// first, then animation frames or attached parts), its fixed position in model
// units decoded from the EXE vector at Addr, the placement yaw (raw units,
// 4096 = one turn) and upright pitch where the code applies one, and a traced
// behaviour note. Night marks objects only drawn in the dusk/night classes.
type Dynamic struct {
	Name    string
	Objs    []int
	X, Y, Z int32  // model units (vector words ×4, like table placements)
	Yaw     int16  // initial/fixed yaw the drawer applies
	Pitch   int16  // rotation about X the drawer applies before the yaw
	Addr    uint32 // EXE address of the position vector (or struct/script)
	Night   bool
	Note    string
}

// dynSpec is the traced, position-address-driven form of the catalog.
type dynSpec struct {
	name       string
	objs       []int
	vec        uint32 // 16-byte position vector in the EXE data segment
	yawFromW   bool   // yaw = vector word @+12 (+ yawBias)
	yaw        int16  // fixed yaw (or bias when yawFromW)
	pitch      int16
	night      bool
	note       string
}

var dynSpecs = []dynSpec{
	{name: "Number girl", objs: []int{250}, vec: 0x80070870, yawFromW: true, yaw: 0x400, pitch: 0x400,
		note: "stands on the grid; upright billboard, removed after the start"},
	{name: "Swinging tunnel sign", objs: []int{192}, vec: 0x80070780, yaw: 0xC00,
		note: "hangs over the first tunnel's road, yaw swings ±sin(frame)/8 around 0xC00"},
	{name: "Rotating sign", objs: []int{175}, vec: 0x80070720,
		note: "spins 16/frame about Y (reversed at night); translucent draw path"},
	{name: "Beacon (palette-cycled)", objs: []int{257}, vec: 0x80070820,
		note: "unrotated; CLUT offset cycles four palettes, one step every four frames"},
	{name: "Start banner", objs: []int{182, 181}, vec: 0x80070730, yaw: 0x400,
		note: "the start/finish gantry banner; swaps to 181 on a scheduled flicker frame"},
	{name: "Big screen A", objs: []int{185, 183, 184}, vec: 0x80070750, yaw: 0xAC0,
		note: "trackside video wall: 183/184 alternate every 4 frames when on, 185 off; UVs re-pointed per frame"},
	{name: "Big screen B", objs: []int{185, 183, 184}, vec: 0x80070760, yaw: 0xB20,
		note: "second video wall, same states"},
	{name: "Start gate crowd (far LOD)", objs: []int{67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83},
		vec: 0x800707A0,
		note: "the start-gate panel with the waving crowd: base 67 + frames 68-83, (frame>>2)&15; drawn unrotated"},
	{name: "Start gate crowd (near LOD base)", objs: []int{146, 145, 152, 153, 164},
		vec: 0x800707A0,
		note: "close-range composite drawn instead of 67 when the camera is near (0x80130CC4 < 900)"},
	{name: "Start semaphore", objs: []int{154, 155, 156, 157, 158}, vec: 0x800707D0,
		note: "the start-light column; the lit frame is 154+n from the rig table 0x80056C44"},
	{name: "Crowd wave (near LOD)", objs: []int{147, 148, 149, 150, 151}, vec: 0x800707A0,
		note: "wave panels 98+b, byte sequence at 0x800100F4 stepping (frame>>2)&15"},
	{name: "Camera crane", objs: []int{258, 259, 260}, vec: 0x800707F0,
		note: "articulated rig (0x800154B8): part ids from rig table 0x80056C44, one joint spins frame<<4"},
	{name: "Start balloons", objs: []int{166, 167, 168, 169, 170, 171, 172, 173}, vec: 0x80070830, yawFromW: true,
		note: "released at the start, drift -X and climb with race progress; 8 sway frames"},
	{name: "Lap banner (on balloons)", objs: []int{279, 287, 295, 303, 311}, vec: 0x80070830, yawFromW: true,
		note: "topper over the balloons: 279+8·bank+frame, bank by lap (byte table 0x80079FA0)"},
	{name: "Hot-air balloon A", objs: []int{139}, vec: 0x80070850, night: true,
		note: "rises 256/frame; gated by flag bit 0x20 at 0x80080444"},
	{name: "Hot-air balloon B", objs: []int{139}, vec: 0x80070860, night: true,
		note: "rises 544/frame; gated by flag bit 0x40"},
	{name: "Start gate crowd (night twin)", objs: []int{67}, vec: 0x80070810, yaw: 0x874, night: true,
		note: "second instance of the crowd gate drawn in the night classes"},
}

// Dynamics decodes the dynamic-object catalog from the executable's text+data
// image (as loaded at 0x80010000). Positions come from each entry's EXE
// vector; everything else is the traced code behaviour.
func Dynamics(text []byte) []Dynamic {
	var out []Dynamic
	for _, s := range dynSpecs {
		off := int(s.vec - exeTextBase)
		if off < 0 || off+16 > len(text) {
			continue
		}
		d := Dynamic{
			Name:  s.name,
			Objs:  s.objs,
			X:     int32(u32(text, off)) * 4,
			Y:     int32(u32(text, off+4)) * 4,
			Z:     int32(u32(text, off+8)) * 4,
			Yaw:   s.yaw,
			Pitch: s.pitch,
			Addr:  s.vec,
			Night: s.night,
			Note:  s.note,
		}
		if s.yawFromW {
			d.Yaw = int16(u32(text, off+12)) + s.yaw
		}
		out = append(out, d)
	}
	// The helicopter: spawn teleport is the script's opcode-0 record at
	// 0x80073790 — halfwords X (unsigned), Y, Z then three angles (pitch,
	// yaw, roll).
	if off := int(uint32(0x80073790) - exeTextBase); off+14 <= len(text) {
		out = append(out, Dynamic{
			Name: "Helicopter", Objs: []int{188, 189},
			X:    int32(uint16(s16(text, off+2))) * 4,
			Y:    int32(s16(text, off+4)) * 4,
			Z:    int32(uint16(s16(text, off+6))) * 4,
			Yaw:  s16(text, off+10),
			Addr: 0x80073790,
			Note: "body 188 + rotor 189 (spun 331/frame); flies the waypoint script at 0x80073790",
		})
	}
	out = append(out, Dynamic{
		Name: "Airplane", Objs: []int{190, 251},
		Yaw: 0xF00, Pitch: -192, Addr: 0x801D6DBC,
		Note: "spawned at race start (struct 0x801D6DBC), glides along the EXE direction vector at 0x800739F4 for 1800 frames; 251 is the night model — no fixed position",
	})
	return out
}
