// spinprobe answers one question from the cartridge: which PLACED actors turn
// on the spot every tick, and how fast?
//
// The coin's step function adds a constant to the yaw short at +$8E once per
// 30 Hz actor tick — the spinning-coin look is geometry, not a texture
// animation. Rather than trust a hand-picked actor list, this sweeps every
// actor that appears in a level's object table:
//
//	profile record {u32 create, u16 actorID, …}
//	  -> the create function's literal pool
//	    -> the vtable it stores at [this+0] (16 consecutive code pointers)
//	      -> the step slot at vtable+$18
//	        -> an `ADD rX, rX, #imm` feeding an `STRH`
//
// ARM9 overlays overlap in address space, so every address is resolved inside
// the overlay the profile was found in first and the ARM9 static second.
//
// Usage (from games/super-mario-64-ds/):
//
//	go run ./extract/cmd/spinprobe -in <rom.nds> [-ext extracted] [-all]
//
// -all drops the #$C00 filter and reports every constant-add/STRH pair, as the
// control that the sweep is not simply finding what it was told to look for.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
	"retroreverse.com/tools/cpu/arm"
	"retroreverse.com/tools/platform/nds"
)

// arm9Base: the ARM9 static loads at $02004000 (header), NOT at the start of
// main RAM — the low 16 KB stay free. Every file offset in extracted/arm9_dec.bin
// is that much below its address.
const arm9Base = 0x02004000

var le = binary.LittleEndian

type src struct {
	name string
	base uint32
	data []byte
}

// ctx resolves an address inside one overlay, then the ARM9 static.
type ctx struct{ own, a9 src }

func (c ctx) at(a uint32) ([]byte, bool) {
	for _, s := range []src{c.own, c.a9} {
		if a >= s.base && a < s.base+uint32(len(s.data)) {
			return s.data[a-s.base:], true
		}
	}
	return nil, false
}

func (c ctx) word(a uint32) (uint32, bool) {
	b, ok := c.at(a)
	if !ok || len(b) < 4 {
		return 0, false
	}
	return le.Uint32(b), true
}

func (c ctx) isCode(p uint32) bool {
	if p&3 != 0 || p == 0 {
		return false
	}
	_, ok := c.at(p)
	return ok
}

func main() {
	in := flag.String("in", "", "cartridge image")
	ext := flag.String("ext", "extracted", "extracted binaries directory")
	all := flag.Bool("all", false, "report every constant yaw-ish add, not just #$C00")
	flag.Parse()
	if *in == "" {
		log.Fatal("usage: spinprobe -in <rom.nds> [-ext extracted] [-all]")
	}

	img, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	rom, err := nds.Open(img)
	if err != nil {
		log.Fatal(err)
	}
	a9b, err := os.ReadFile(filepath.Join(*ext, "arm9_dec.bin"))
	if err != nil {
		log.Fatal(err)
	}
	a9 := src{"arm9", arm9Base, a9b}
	srcs := []src{a9}
	for _, o := range rom.ARM9Overlays() {
		d, err := os.ReadFile(filepath.Join(*ext, fmt.Sprintf("ovl9_%03d_dec.bin", o.ID)))
		if err != nil {
			continue
		}
		srcs = append(srcs, src{fmt.Sprintf("ovl%d", o.ID), o.RAMAddr, d})
	}

	ls, err := sm64ds.OpenLevels(*in, *ext)
	if err != nil {
		log.Fatal(err)
	}
	placed := map[int]int{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		for _, o := range lv.Objects {
			placed[o.Actor]++
		}
	}

	type hit struct {
		actor, n               int
		prof, create, vt, step uint32
		src, spin              string
	}
	var hits []hit
	seen := map[string]bool{}
	withProfile, withVtable := map[int]bool{}, map[int]bool{}
	for _, s := range srcs {
		c := ctx{own: s, a9: a9}
		for i := 0; i+8 <= len(s.data); i += 4 {
			create := le.Uint32(s.data[i:])
			if !c.isCode(create) {
				continue
			}
			id := int(le.Uint16(s.data[i+4:]))
			if placed[id] == 0 {
				continue
			}
			withProfile[id] = true
			vt := vtableOf(c, create)
			if vt == 0 {
				continue
			}
			withVtable[id] = true
			step, _ := c.word(vt + 0x18)
			spin := spinsYaw(c, step, *all)
			if spin == "" {
				continue
			}
			k := fmt.Sprintf("%d/%08X", id, create)
			if seen[k] {
				continue
			}
			seen[k] = true
			hits = append(hits, hit{id, placed[id], s.base + uint32(i), create, vt, step, s.name, spin})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].actor < hits[j].actor })
	// The denominator: a zero hit for an actor means nothing unless the sweep
	// actually reached that actor's step.
	fmt.Printf("placed actors %d; with a profile record %d; with a recovered vtable %d; matching steps %d\n",
		len(placed), len(withProfile), len(withVtable), len(hits))
	for _, h := range hits {
		fmt.Printf("  actor %3d  n=%4d  %-6s profile %08X create %08X vtable %08X step %08X\n      %s\n",
			h.actor, h.n, h.src, h.prof, h.create, h.vt, h.step, h.spin)
	}
}

// vtableOf recovers the vtable a create function stores at [this+0]: the first
// pc-relative literal that points at 16 consecutive code pointers.
func vtableOf(c ctx, create uint32) uint32 {
	code, ok := c.at(create)
	if !ok {
		return 0
	}
	for k := 0; k < 64 && (k+1)*4 <= len(code); k++ {
		w := le.Uint32(code[k*4:])
		if w&0x0F7F0000 != 0x051F0000 { // LDR rX, [pc, #±imm]
			continue
		}
		a := create + uint32(k*4) + 8
		if w&0x00800000 != 0 {
			a += w & 0xFFF
		} else {
			a -= w & 0xFFF
		}
		lit, ok := c.word(a)
		if !ok {
			continue
		}
		good := 0
		for j := 0; j < 16; j++ {
			p, ok := c.word(lit + uint32(j*4))
			if ok && c.isCode(p) {
				good++
			}
		}
		if good == 16 {
			return lit
		}
	}
	return 0
}

// accumulates keeps only `ADD rX, rX, #imm` — a value advanced every tick.
// `ADD r0, r4, #0xC00` with r4 the object base is an ADDRESS, and one actor's
// step computes exactly that before an unrelated STRH.
func accumulates(text string) bool {
	f := strings.Split(text, ", ")
	if len(f) != 3 {
		return false
	}
	dst := f[0]
	if i := strings.LastIndex(dst, " "); i >= 0 {
		dst = dst[i+1:]
	}
	return dst == f[1]
}

// spinsYaw reports a constant per-tick add that lands in a stored halfword.
func spinsYaw(c ctx, step uint32, all bool) string {
	code, ok := c.at(step)
	if !ok {
		return ""
	}
	for k := 0; k < 400 && (k+1)*4 <= len(code); k++ {
		a := step + uint32(k*4)
		in := arm.DecodeARM(code[k*4:], a)
		if !strings.HasPrefix(in.Mnem, "ADD") || !accumulates(in.Text) {
			continue
		}
		if all {
			if !strings.Contains(in.Text, ", #0x") || strings.Contains(in.Text, "sp") {
				continue
			}
		} else if !strings.Contains(in.Text, "#0xC00") {
			continue
		}
		var st string
		for j := k; j < k+6 && (j+1)*4 <= len(code); j++ {
			i2 := arm.DecodeARM(code[j*4:], step+uint32(j*4))
			if strings.HasPrefix(i2.Mnem, "STRH") {
				st = "  ->  " + i2.Text
			}
		}
		if st == "" {
			continue
		}
		return fmt.Sprintf("%08X  %s%s", a, in.Text, st)
	}
	return ""
}
