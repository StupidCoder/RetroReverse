package main

// The room object tables, decoded from the game's own loader.
//
// The lookup is three levels deep:
//
//	$080D50FC[area] -> a per-room array
//	            [room] -> a per-SLOT array (8 slots)
//	              [slot] -> the list itself
//
// The eight slots are not eight lists of the same thing. The room loader
// treats each index differently:
//
//	slot 0, 1  entity records, spawned on every room entry ($0804ADDC)
//	slot 2     entity records with respawn tracking ($0804B058): the walker
//	           counts records 0..31 and checks a per-index kill flag
//	           ($08049D1C) before spawning a class-3 (enemy) record; a killed
//	           enemy stays dead. The entity remembers its index in +0x6C.
//	slot 3     8-byte COMMAND records, zero-terminated ($0804B1AC): an opcode
//	           1..13 dispatched through the jump table at $0804B1CC.
//	slot 4..7  a single function pointer, called directly ($08000EF2) — the
//	           room's own code hooks, not data.
//
// An entity record is 16 bytes, $FF-terminated, read by $0804ADF8:
//
//	+0x00  low nibble: CLASS (entity+8); high nibble: initial-state flags
//	+0x01  low nibble: target entity list (0xF = the class's default);
//	       high nibble: spawn MODE — 0 place, 1 raw (no position, no extra
//	       fields), 2 place + flag 0x20, 4 place + ATTACH the +0x0C word
//	       ($0807DAD0), 5 skip if one with this class+id already exists
//	+0x02  ID within the class (entity+9)   +0x03  sub-id (entity+0xA)
//	+0x04  -> entity+0xB                    +0x05  -> entity+0xE
//	+0x06  u16, meaning depends on the class
//	+0x08  x, +0x0A y — RELATIVE TO THE ROOM, added to the room's origin
//	       (class 9 records are never placed: the spawner skips $0804AF0C)
//	+0x0C  mode 4: the attached pointer — for NPCs this is a ROM script for
//	       the 139-opcode interpreter ($0807DFBC); otherwise per-species data
//
// What a class IS comes from the loader and the dispatch tables:
//
//	class 3  enemy    (slot 2's kill flags apply only to class 3)
//	class 6  object   (pots, bushes, chests — the bulk of the placed world)
//	class 7  NPC      (88 of 111 records attach a dialogue/behaviour script)
//	class 9  manager  (no position; room-logic entities)
//
// The behaviour handler is NOT in the record. The per-class update functions
// (table at $080B2248, indexed by entity+8) dispatch entity+9 through a
// per-class table:
//
//	class 3 -> $080D3BF8[id]      class 4 -> $08129320[id]
//	class 6 -> $080B2D4C[id]      class 7 -> $080B313C + id*12
//	class 8 -> $080B2CE8[id]      class 9 -> $080B3054[id]
//
// A slot-3 command record is 8 bytes {op, b1, b2, b3, h4:u16, h6:u16}:
//
//	op  1  play song b1 ($0807CCB4)
//	op  2  register {flag=b1, pos=h4}: while flag b1 is unset, metatile
//	       0x74 is written at pos ($0804B16C -> $0807B314)
//	op  4  spawn manager 0x24 at pixel (h4, h6), params b1, b2 ($0804B300)
//	op  7  global <- b3
//	op  9  $0805BB00(b3, 1)
//	op 10  if flag b2h (u16 at +2) is set, write metatile h6 at pos h4;
//	       else spawn manager 0x2A once ($0804B340)
//	op 11  copy 32 bytes from $...[b1*32] (a settings block)
//	op 12  global <- b1, then $08054524
//	op 13  global+9 <- b3
//
// A metatile position packs as y*64+x ($0807B314 decodes with mask 0x3F),
// in metatile units of 16 pixels.

import "fmt"

const objTable = 0x080D50FC

// classTables maps an entity class to its behaviour dispatch table: the
// handler for (class, id) is what groups equal objects across rooms.
var classTables = map[int]struct {
	Base   uint32
	Stride int
}{
	3: {0x080D3BF8, 4},
	4: {0x08129320, 4},
	6: {0x080B2D4C, 4},
	7: {0x080B313C, 12},
	8: {0x080B2CE8, 4},
	9: {0x080B3054, 4},
}

// ClassNames are the structurally identified classes: 3 from the kill-flag
// walker, 7 from the attached scripts, 9 from the skipped placement, 6 from
// what the placed records land on.
var ClassNames = map[int]string{
	3: "enemy", 4: "class-4", 6: "object", 7: "npc", 8: "class-8", 9: "manager",
}

// HandlerFor resolves the behaviour function the class dispatcher would call
// for this id, or 0 if the class has no table.
func HandlerFor(rom []byte, class, id int) uint32 {
	t, ok := classTables[class]
	if !ok {
		return 0
	}
	return rd32(rom, t.Base+uint32(id*t.Stride))
}

// Object is one 16-byte entity record.
type Object struct {
	Addr    string `json:"addr"`
	Area    int    `json:"area"`
	Room    int    `json:"room"`
	Slot    int    `json:"slot"`
	Class   int    `json:"class"`
	Flags   int    `json:"flags"` // record[0] high nibble
	Mode    int    `json:"mode"`  // record[1] high nibble
	List    int    `json:"list"`  // record[1] low nibble
	ID      int    `json:"id"`
	Sub     int    `json:"sub"`
	B4      int    `json:"b4"`
	B5      int    `json:"b5"`
	H6      int    `json:"h6"`
	X       int    `json:"x"` // room-relative
	Y       int    `json:"y"`
	Extra   uint32 `json:"extra"`   // the +0x0C word
	KillIdx int    `json:"killIdx"` // slot-2 respawn index, -1 otherwise
	Handler uint32 `json:"handler"`
}

// Placed reports whether the spawner gives this record a position at all.
func (o Object) Placed() bool { return o.Class != 9 && o.Mode != 1 }

// Command is one 8-byte slot-3 record.
type Command struct {
	Addr string `json:"addr"`
	Area int    `json:"area"`
	Room int    `json:"room"`
	Op   int    `json:"op"`
	B1   int    `json:"b1"`
	B2   int    `json:"b2"`
	B3   int    `json:"b3"`
	H2   int    `json:"h2"` // u16 at +2 (op 10's flag)
	H4   int    `json:"h4"`
	H6   int    `json:"h6"`
}

// Pos returns the pixel position a command refers to, if it has one:
// op 4 carries pixels directly, ops 2 and 10 a y*64+x metatile index.
func (c Command) Pos() (x, y int, ok bool) {
	switch c.Op {
	case 4:
		return c.H4, c.H6, true
	case 2, 10:
		return (c.H4 & 0x3F) * 16, (c.H4 >> 6) * 16, true
	}
	return 0, 0, false
}

// ListFor resolves table[area][room][slot].
func ListFor(rom []byte, area, room, slot int) uint32 {
	arr := rd32(rom, uint32(objTable+area*4))
	if arr>>24 != 0x08 {
		return 0
	}
	per := rd32(rom, arr+uint32(room*4))
	if per>>24 != 0x08 {
		return 0
	}
	l := rd32(rom, per+uint32(slot*4))
	if l>>24 != 0x08 {
		return 0
	}
	return l
}

// Objects walks one entity list to its $FF terminator.
func Objects(rom []byte, list uint32, slot int) []Object {
	var out []Object
	for i := 0; i < 256; i++ {
		a := list + uint32(i*16)
		o := int(a & 0x01FFFFFF)
		if o+16 > len(rom) || rom[o] == 0xFF {
			break
		}
		ob := Object{
			Addr: fmt.Sprintf("%08X", a), Slot: slot,
			Class: int(rom[o] & 0xF), Flags: int(rom[o] >> 4),
			Mode: int(rom[o+1] >> 4), List: int(rom[o+1] & 0xF),
			ID: int(rom[o+2]), Sub: int(rom[o+3]),
			B4: int(rom[o+4]), B5: int(rom[o+5]), H6: int(rd16(rom, a+6)),
			X: int(rd16(rom, a+8)), Y: int(rd16(rom, a+10)),
			Extra: rd32(rom, a+12), KillIdx: -1,
		}
		if slot == 2 && ob.Class == 3 && i <= 0x1F {
			ob.KillIdx = i
		}
		ob.Handler = HandlerFor(rom, ob.Class, ob.ID)
		out = append(out, ob)
	}
	return out
}

// Commands walks one slot-3 list to its zero terminator.
func Commands(rom []byte, list uint32) []Command {
	var out []Command
	for i := 0; i < 512; i++ {
		a := list + uint32(i*8)
		o := int(a & 0x01FFFFFF)
		if o+8 > len(rom) || rom[o] == 0 {
			break
		}
		out = append(out, Command{
			Addr: fmt.Sprintf("%08X", a),
			Op:   int(rom[o]), B1: int(rom[o+1]), B2: int(rom[o+2]), B3: int(rom[o+3]),
			H2: int(rd16(rom, a+2)), H4: int(rd16(rom, a+4)), H6: int(rd16(rom, a+6)),
		})
	}
	return out
}
