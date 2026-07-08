package level

// Object placement (enemies, moving platforms, the level-end Daisy/coin-lift, etc.)
// is a separate per-level list from the background map, decoded by the ROM's spawner
// at $2492 (init at $2453). It is NOT encoded in the column RLE — only static blocks
// ($70/$80/$5F ?-blocks & breakables) live there.
//
// The chain (reimplemented here, traced from the code, not guessed):
//
//	$401A[ffe4]  (in the world's data bank) -> L, the level's placement list
//	each entry is 3 bytes, sorted ascending by trigger column:
//	  byte0  col   : the scroll column ($C0AB) at which the object spawns. $C0AB
//	                 advances once per 16 px of scroll, so the object's map column
//	                 (8 px tiles) is col*2.
//	  byte1  pos   : bits 0-4 -> map row (packed&$1F); bits 6-7 -> fine X, 4px per
//	                 unit (the spawner adds (pos&$C0)>>4 to the spawn X).
//	  byte2  type  : bits 0-6 -> object type (indexes the $336C init table);
//	                 bit 7 -> "hard mode" flag (the object is only spawned, or spawned
//	                 differently, on a second quest; $FF9A gates it at $24E6).
//	the list is terminated by an entry whose col byte is $FF.

// Object is one placed object. Col/Row are the map-tile coordinates (8x8 tiles) of the
// object's metasprite origin (where the engine puts $FFC3/$FFC2). Type is the object type
// id, Hard is the bit-7 second-quest flag, and FineX is the 0-3 nudge from the position
// byte's top two bits, worth 4px each: world X = Col*8 + FineX*4 (exact — the spawner's
// $D0 screen offset, the $C0AB start value of 12 and its camera%16==8 tick phase cancel;
// oracle-verified per spawn by cmd/spawnverify).
type Object struct {
	Col, Row int
	Type     byte
	Hard     bool
	FineX    int
}

// DecodeObjectsByID returns the placed objects for a level id (0x11=1-1 .. 0x43=4-3),
// using the same world->bank selection as DecodeLevelByID.
func DecodeObjectsByID(rom []byte, id byte) []Object {
	world := int(id >> 4)
	level := int(id & 0x0F)
	ffe4 := byte((world-1)*3 + (level - 1))
	return DecodeObjects(rom, worldBank[world], ffe4)
}

// --- object metasprites (the sprite graphics an object draws) ---
//
// The object sprite draw ($25B7) renders a slot's metasprite using the slot's frame field
// (slot+6) as an index into the pointer table at $2FD9 (bank-0 fixed). The pointed-at
// stream is "turtle graphics": a byte with bit7 clear moves the cursor (low nibble:
// bit3 up / bit2 down / bit1 left / bit0 right, 8 px each) and carries the OAM attribute;
// a byte with bit7 set stamps an 8x8 OBJ sprite of that tile id (the byte itself, bit7 and
// all) at the cursor. ($FF ends.) SML runs sprites in 8x8 mode (LCDC bit2 = 0), so each
// emitted byte is one 8x8 tile — e.g. the Goomba (frame $01) is the single tile $90.

// The draw routine ($25B7) chooses one of two metasprite pointer tables by the object's
// facing flag ($FFC5 bit0): $2FD9 for one direction, $30AB for the other — the two hold
// mirror-image layouts (and per-tile X-flip attrs). For a static render we want the
// natural, unflipped layout (tiles ordered left-to-right like the tile sheet / Mario's
// $04,$05), which is $30AB; $2FD9 reverses the columns.
const msTable = 0x30AB // object metasprite pointer table (natural facing; bank-0 fixed)

// Sprite is one 8x8 OBJ sprite of a metasprite: the tile id, its pixel offset from
// the metasprite origin (DY negative = above the origin), and its flips. A control
// byte carries the OAM attribute alongside the cursor movement: the draw routine
// ($25F0) rotates it left into the attr ($D000; RES 4 masks the up-move bit), so
// control bit 4 -> OAM bit 5 = X flip and bit 5 -> bit 6 = Y flip — the Goomba's
// second walk pose is just "attr $10, tile $90": the same tile mirrored.
type Sprite struct {
	Tile         byte
	DX, DY       int
	XFlip, YFlip bool
}

// DecodeMetasprite decodes object metasprite frame `f` from the $2FD9 table into its 8x8
// sprites (reimplements the $25B7 stream walker; data is bank-0 fixed so no bank needed).
func DecodeMetasprite(rom []byte, f int) []Sprite {
	p := int(rom[msTable+f*2]) | int(rom[msTable+f*2+1])<<8
	var out []Sprite
	cx, cy := 0, 0
	xf, yf := false, false // current attr, replaced by each control byte
	for i := 0; i < 64 && p >= 0 && p < len(rom); i++ {
		b := rom[p]
		p++
		if b == 0xFF {
			break
		}
		if b&0x80 != 0 {
			out = append(out, Sprite{Tile: b, DX: cx, DY: cy, XFlip: xf, YFlip: yf})
			continue
		}
		xf, yf = b&0x10 != 0, b&0x20 != 0
		if b&0x08 != 0 {
			cy -= 8
		}
		if b&0x04 != 0 {
			cy += 8
		}
		if b&0x02 != 0 {
			cx -= 8
		}
		if b&0x01 != 0 {
			cx += 8
		}
	}
	return out
}

// Object behaviour is script-driven: the runner at $263F looks up each object's per-type
// script via the table at $3495 (indexed by type id) and interprets it ($26AC). Script
// command $F8 xx sets the animation frame ($FFC6); a type's *base* frame is the first $F8
// in its script. So we can recover a representative metasprite frame for every type by
// reading its script — no per-level observation needed (that earlier approach only saw the
// enemies the oracle's auto-play reached). Verified to match those observations.
const (
	scriptTable = 0x3495 // per-type object script pointer table (bank-0 fixed)
	opSetFrame  = 0xF8   // script opcode: next byte = animation frame id
)

// TypeBaseFrame returns the base metasprite frame for object type `typ` by scanning its
// script for the first $F8 (set-frame) command. ok is false if the type has no valid script.
func TypeBaseFrame(rom []byte, typ byte) (frame byte, ok bool) {
	p := int(rom[scriptTable+int(typ)*2]) | int(rom[scriptTable+int(typ)*2+1])<<8
	if p < 0x3500 || p >= 0x3D00 { // outside the script block => not a real object type
		return 0, false
	}
	for i := 0; i < 96 && p+1 < len(rom); i++ {
		if rom[p] == opSetFrame {
			return rom[p+1], true
		}
		p++
	}
	return 0, false
}

// TypeFrames builds the full type -> base-frame map from the script table.
func TypeFrames(rom []byte) map[byte]byte {
	m := map[byte]byte{}
	for t := 0; t < 0x80; t++ {
		if fr, ok := TypeBaseFrame(rom, byte(t)); ok {
			m[byte(t)] = fr
		}
	}
	return m
}

// FrameStep is one pose of an object's animation timeline: metasprite frame Frame
// shown for Frames game frames (60 Hz).
type FrameStep struct {
	Frame  byte
	Frames int
}

// nextScriptStart returns the smallest script start > p within the script block —
// the boundary at which a script without an $FF must stop (same rule as
// cmd/objscript).
func nextScriptStart(rom []byte, p int) int {
	end := 0x3D00
	for t := 0; t < 0x80; t++ {
		s := int(rom[scriptTable+t*2]) | int(rom[scriptTable+t*2+1])<<8
		if s > p && s < end {
			end = s
		}
	}
	return end
}

// TypeTimeline extracts an object type's looping animation from its behaviour
// script: a move op (<$E0) is 1 duration unit, a coast $E0-$EF is op&$0F more,
// $F8 xx switches the pose, and $FF restarts (the loop). One duration unit is NOT
// one frame: the runner's sub-step counter ($2689, the $FFC9 nibble pair set by
// $F4 xx) only ticks the duration every (step&$0F)+1 frames — the Goomba's $F4 02
// makes its 3-unit poses last 9 real frames (oracle-measured). Side-effect commands
// are skipped. A $F3 (become another type) means the script is a transient — nil
// (static icon). A player gate ($F6 wait / $FB proximity restart) discards what
// came before it: the exported loop is the action segment the gate triggers (the
// piranha plant's chomp), with the gate itself ignored. Returns nil unless the
// result is a real cycle (>= 2 distinct poses).
func TypeTimeline(rom []byte, typ byte) []FrameStep {
	p := int(rom[scriptTable+int(typ)*2]) | int(rom[scriptTable+int(typ)*2+1])<<8
	if p < 0x3500 || p >= 0x3D00 {
		return nil
	}
	end := nextScriptStart(rom, p)
	var steps []FrameStep
	cur, dur := -1, 0
	scale := 1 // frames per duration unit: ($F4 arg & $0F)+1, cleared to 1 at spawn
	flush := func() {
		if cur < 0 || dur == 0 {
			dur = 0
			return
		}
		if n := len(steps); n > 0 && steps[n-1].Frame == byte(cur) {
			steps[n-1].Frames += dur
		} else {
			steps = append(steps, FrameStep{byte(cur), dur})
		}
		dur = 0
	}
	for i := 0; i < 96 && p < end; i++ {
		op := rom[p]
		switch {
		case op == 0xFF: // restart: the loop closes here
			flush()
			p = end // done
		case op == 0xF3: // become another type: transient spawn-in, not a cycle
			return nil
		case op == 0xF6 || op == 0xFB: // player gate: keep only what follows
			steps, dur = steps[:0], 0
			p += 2
		case op == 0xF8: // set pose
			if int(rom[p+1]) != cur {
				flush()
				cur = int(rom[p+1])
			}
			p += 2
		case op == 0xF4: // set sub-step divider: units now last (arg&$0F)+1 frames
			scale = int(rom[p+1]&0x0F) + 1
			p += 2
		case op == 0xF7 || op == 0xFE: // argless commands
			p++
		case op >= 0xF0: // other commands: side effects, one arg byte
			p += 2
		case op >= 0xE0: // coast op&$0F more units
			dur += int(op&0x0F) * scale
			p++
		default: // move: one unit
			dur += scale
			p++
		}
	}
	flush()
	// Merge the wrap-around (last pose == first pose reads as one hold).
	if n := len(steps); n > 1 && steps[0].Frame == steps[n-1].Frame {
		steps[0].Frames += steps[n-1].Frames
		steps = steps[:n-1]
	}
	distinct := map[byte]bool{}
	total := 0
	for _, s := range steps {
		distinct[s.Frame] = true
		total += s.Frames
	}
	if len(distinct) < 2 || total < 2 {
		return nil
	}
	return steps
}

// SolidTile reports whether a background tile id is solid ground — something Mario and
// enemies stand on (as opposed to passable scenery). Solidity is decided purely by the
// tile id: the engine's foot/side checks read the BG tile under the actor ($0153) and
// compare it to a threshold. Mario ($17B3) treats id >= $60 as floor; the enemy checks
// ($2B7B/$2B91/$2BB2/...) treat the range [$5F, $F0) as solid (tiles >= $F0 are special
// metadata tiles, never floor). We use the shared range [$60, $F0).
func SolidTile(id byte) bool { return id >= 0x60 && id < 0xF0 }

// Pipe is one warp pipe: standing on the pipe tile ($70) at (Screen,Col) of the main path
// and pressing Down sends Mario to screen Dest (the bonus rooms are screens 1 and 2 of the
// order table); leaving the bonus room returns him to screen RetScreen at pixel (RetX,RetY).
type Pipe struct {
	Screen, Col     int
	Dest, RetScreen int
	RetX, RetY      int
}

// DecodePipes decodes a level's warp-pipe table. The $70-tile handler ($22A0) reads this
// from $651C[ffe4] — a pointer table in bank 3 (the handler always pages bank 3) to a list
// of 6-byte entries [screen, col, dest, retScreen, retX, retY], terminated by screen=$FF.
// The handler stamps the 4 data bytes into the parallel metadata map at VRAM+$3000 above
// the pipe tile; at runtime the pipe-entry code ($175C) reads them back as the destination.
func DecodePipes(rom []byte, id byte) []Pipe {
	world := int(id >> 4)
	ffe4 := (world-1)*3 + int(id&0x0F) - 1
	const bank = 3 // the warp table is always in bank 3
	list := bankWord(rom, bank, 0x651C+uint16(ffe4)*2)
	var out []Pipe
	ptr := list
	for i := 0; i < 64; i++ {
		scr := bankByte(rom, bank, ptr)
		if scr == 0xFF {
			break
		}
		out = append(out, Pipe{
			Screen:    int(scr),
			Col:       int(bankByte(rom, bank, ptr+1)),
			Dest:      int(bankByte(rom, bank, ptr+2)),
			RetScreen: int(bankByte(rom, bank, ptr+3)),
			RetX:      int(bankByte(rom, bank, ptr+4)),
			RetY:      int(bankByte(rom, bank, ptr+5)),
		})
		ptr += 6
	}
	return out
}

// DecodeObjects decodes the placement list for global level index ffe4 (0-11) from the
// pointer table at $401A in the given bank.
func DecodeObjects(rom []byte, bank int, ffe4 byte) []Object {
	list := bankWord(rom, bank, 0x401A+uint16(ffe4)*2)
	var objs []Object
	ptr := list
	for i := 0; i < 256; i++ {
		col := bankByte(rom, bank, ptr)
		if col == 0xFF { // list terminator
			break
		}
		pos := bankByte(rom, bank, ptr+1)
		typ := bankByte(rom, bank, ptr+2)
		ptr += 3
		// pos&$1F is the object's SCREEN tile-row: the engine sets Y=((pos&$1F)<<3)+$10,
		// drawn as OAM Y (so screen pixel Y = Y-16, screen row = pos&$1F). The level's
		// column data is blitted to BG rows 2-17 (rows 0-1 are the HUD), so the matching
		// row in the 16-row map is pos&$1F - 2. (Verified against the oracle: the first
		// 1-1 Goomba has slot Y=$88 = screen row 15 = map row 13.)
		objs = append(objs, Object{
			Col:   int(col) * 2,
			Row:   int(pos&0x1F) - 2,
			Type:  typ & 0x7F,
			Hard:  typ&0x80 != 0,
			FineX: int(pos&0xC0) >> 6,
		})
	}
	return objs
}
