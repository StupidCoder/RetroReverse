package main

// gateprobe: dump an actor's profile, vtable and the disassembly of its
// create / init / step, so what gates the actor on the player's progress can
// be read off the game's own code.
import (
	"encoding/binary"
	"encoding/json"
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

// callDepth is how many calls deep the sweeps follow from an actor's own
// methods (-depth).
var callDepth = 1

type src struct {
	name string
	base uint32
	data []byte
}

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
	in := flag.String("in", "", "rom (only for the overlay table)")
	ext := flag.String("ext", "extracted", "extracted dir")
	actor := flag.Int("actor", 0, "actor id")
	ovl := flag.String("ovl", "", "restrict to this overlay (e.g. ovl15)")
	fn := flag.String("fn", "", "disassemble this address instead (hex)")
	n := flag.Int("n", 90, "instructions")
	cover := flag.Int("cover", 0, "list the overlays covering this address")
	depth := flag.Int("depth", 1, "how many calls deep the sweeps follow")
	sweep := flag.Bool("sweep", false, "which placed actors query the save's star bits")
	msgs := flag.Bool("msgs", false, "which placed actors reach the message system")
	dropped := flag.Bool("dropped", false, "placements the exporter drops because the oracle bound no model")
	classOf := flag.String("class", "", "comma-separated actor ids: print each one's C++ class name")
	callersOf := flag.String("callers", "", "list every BL site targeting this address")
	actorMsgs := flag.Int("actormsgs", 0, "for this actor, resolve every placement's par1 as a message id")
	msgid := flag.String("msgid", "", "resolve comma-separated EXTERNAL message ids to their text")
	pars := flag.Int("pars", 0, "list every placement of this level (1-based) with its params")
	calls := flag.String("calls", "", "list BL targets from this address")
	words := flag.String("words", "", "dump raw words at this address")
	tables := flag.Int("tables", 0, "dump every object-table entry of this level (1-based)")
	flag.Parse()
	callDepth = *depth

	img, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	rom, _ := nds.Open(img)
	a9b, _ := os.ReadFile(filepath.Join(*ext, "arm9_dec.bin"))
	a9 := src{"arm9", arm9Base, a9b}
	srcs := []src{a9}
	for _, o := range rom.ARM9Overlays() {
		d, err := os.ReadFile(filepath.Join(*ext, fmt.Sprintf("ovl9_%03d_dec.bin", o.ID)))
		if err != nil {
			continue
		}
		srcs = append(srcs, src{fmt.Sprintf("ovl%d", o.ID), o.RAMAddr, d})
	}

	if *sweep {
		sweepStarGates(srcs, a9, *in, *ext)
		return
	}
	if *msgs {
		// $020B8EC0 translates an actor's external message ID to an INF1 index
		// (the range table at $0208EEEC; sm64ds.MsgIndex); $020BB060 is what
		// actually opens the message window. An actor that reaches either has
		// words of its own.
		sweepCallers(srcs, a9, *in, *ext, map[uint32]string{
			0x020B8EC0: "msgIndex", 0x020BB060: "openWindow",
		}, "the message system")
		return
	}
	if *dropped {
		reportDropped(*in, *ext)
		return
	}
	if *classOf != "" {
		for _, tok := range strings.Split(*classOf, ",") {
			var id int
			fmt.Sscanf(strings.TrimSpace(tok), "%d", &id)
			found := false
			for _, s := range srcs {
				c := ctx{own: s, a9: a9}
				pp, ok := c.word(0x02090864 + uint32(id)*4)
				if !ok || !c.isCode(pp) {
					continue
				}
				if w, ok := c.word(pp + 4); !ok || int(w&0xFFFF) != id {
					continue
				}
				ti, _ := c.word(pp + 0x20)
				nm, _ := c.word(ti + 4)
				b, ok := c.at(nm)
				if !ok {
					continue
				}
				n := 0
				for n < len(b) && b[n] != 0 && n < 64 {
					n++
				}
				fmt.Printf("actor %3d  %-7s profile %08X  class %q\n", id, s.name, pp, string(b[:n]))
				found = true
				break
			}
			if !found {
				fmt.Printf("actor %3d  no profile/typeinfo found\n", id)
			}
		}
		return
	}
	if *callersOf != "" {
		var want uint32
		fmt.Sscanf(*callersOf, "%x", &want)
		fmt.Printf("=== BL sites targeting %08X ===\n", want)
		for _, s := range srcs {
			c := ctx{own: s, a9: a9}
			for i := 0; i+4 <= len(s.data); i += 4 {
				in := arm.DecodeARM(s.data[i:], s.base+uint32(i))
				if in.HasTarget && strings.HasPrefix(in.Mnem, "BL") && in.Target == want {
					fmt.Printf("  %-8s %08X\n", s.name, s.base+uint32(i))
				}
			}
			_ = c
		}
		return
	}
	if *actorMsgs > 0 {
		actorMsgSurvey(*in, *ext, *actorMsgs)
		return
	}
	if *msgid != "" {
		showMsg(*in, *ext, *msgid)
		return
	}
	if *pars > 0 {
		listPars(*in, *ext, *pars-1)
		return
	}
	if *calls != "" {
		var a uint32
		fmt.Sscanf(*calls, "%x", &a)
		for _, s := range srcs {
			if *ovl != "" && s.name != *ovl {
				continue
			}
			c := ctx{own: s, a9: a9}
			code, ok := c.at(a)
			if !ok {
				continue
			}
			fmt.Printf("--- BL targets from %08X (%s), %d insts\n", a, s.name, *n)
			seen := map[uint32]int{}
			var order []uint32
			for k := 0; k < *n && (k+1)*4 <= len(code); k++ {
				in := arm.DecodeARM(code[k*4:], a+uint32(k*4))
				if in.HasTarget && strings.HasPrefix(in.Mnem, "BL") {
					if seen[in.Target] == 0 {
						order = append(order, in.Target)
					}
					seen[in.Target]++
				}
			}
			for _, t := range order {
				where := "ovl"
				if t >= arm9Base && t < arm9Base+uint32(len(a9.data)) {
					where = "ARM9"
				}
				fmt.Printf("  %08X  x%d  %s\n", t, seen[t], where)
			}
			return
		}
		return
	}
	if *words != "" {
		var a uint32
		fmt.Sscanf(*words, "%x", &a)
		for _, s := range srcs {
			if *ovl != "" && s.name != *ovl {
				continue
			}
			c := ctx{own: s, a9: a9}
			if _, ok := c.at(a); !ok {
				continue
			}
			fmt.Printf("--- %08X in %s\n", a, s.name)
			for k := 0; k < *n; k++ {
				w, _ := c.word(a + uint32(k*4))
				fmt.Printf("  +%02X  %08X\n", k*4, w)
			}
			return
		}
		return
	}
	if *tables != 0 {
		dumpTables(srcs, a9, *ext, *tables)
		return
	}
	if *cover != 0 {
		fmt.Printf("overlays covering %08X:\n", *cover)
		for _, s := range srcs {
			if uint32(*cover) >= s.base && uint32(*cover) < s.base+uint32(len(s.data)) {
				fmt.Printf("  %-8s %08X..%08X\n", s.name, s.base, s.base+uint32(len(s.data)))
			}
		}
		return
	}
	if *fn != "" {
		var a uint32
		fmt.Sscanf(*fn, "%x", &a)
		for _, s := range srcs {
			if *ovl != "" && s.name != *ovl {
				continue
			}
			c := ctx{own: s, a9: a9}
			if code, ok := c.at(a); ok {
				fmt.Printf("--- %08X in %s\n", a, s.name)
				dump(c, a, *n)
				_ = code
				return
			}
		}
		return
	}

	// The engine's own actor->profile array at $02090864 (783 entries; the
	// factory $02043098 does LDR r0,[[table]+actor*4]; BLX [r0]). Resolving
	// through it beats pattern-matching profile records out of the binaries.
	const profileTable = 0x02090864
	for _, s := range srcs {
		if *ovl != "" && s.name != *ovl {
			continue
		}
		c := ctx{own: s, a9: a9}
		pp, ok := c.word(profileTable + uint32(*actor)*4)
		if !ok || !c.isCode(pp) {
			continue
		}
		create, _ := c.word(pp)
		if !c.isCode(create) {
			continue
		}
		id := uint32(0)
		if w, ok := c.word(pp + 4); ok {
			id = w & 0xFFFF
		}
		vt := vtableOf(c, create)
		fmt.Printf("=== actor %d  via %s  profile %08X  create %08X  id@+4 %d  vtable %08X\n",
			*actor, s.name, pp, create, id, vt)
		if id != uint32(*actor) {
			fmt.Println("    (profile id does not match — wrong overlay banked here)")
			continue
		}
		fmt.Println("--- create")
		dump(c, create, *n)
		if vt == 0 {
			continue
		}
		for _, sl := range []struct {
			off  uint32
			name string
		}{{0, "vt+00"}, {0x18, "vt+18 step"}} {
			f, _ := c.word(vt + sl.off)
			fmt.Printf("--- %s %08X\n", sl.name, f)
			dump(c, f, *n)
		}
		return
	}
	fmt.Println("no profile found")
}

func dump(c ctx, a uint32, n int) {
	code, ok := c.at(a)
	if !ok {
		fmt.Println("   (unmapped)")
		return
	}
	for k := 0; k < n && (k+1)*4 <= len(code); k++ {
		pc := a + uint32(k*4)
		in := arm.DecodeARM(code[k*4:], pc)
		note := ""
		if in.Text != "" && strings.Contains(in.Text, "[pc,") {
			var lit uint32
			w := le.Uint32(code[k*4:])
			t := pc + 8
			if w&0x00800000 != 0 {
				t += w & 0xFFF
			} else {
				t -= w & 0xFFF
			}
			lit, _ = c.word(t)
			note = fmt.Sprintf("   ; =%08X", lit)
		}
		fmt.Printf("  %08X  %s%s\n", pc, in.Text, note)
		if in.Text == "BX lr" && k > 2 {
			break
		}
	}
}

func vtableOf(c ctx, create uint32) uint32 {
	code, ok := c.at(create)
	if !ok {
		return 0
	}
	for k := 0; k < 64 && (k+1)*4 <= len(code); k++ {
		w := le.Uint32(code[k*4:])
		if w&0x0F7F0000 != 0x051F0000 {
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

// dumpTables walks a level's settings block exactly like the engine's
// $020FE190 and prints EVERY objects-table entry — including the types
// sm64ds/level.go does not decode.
func dumpTables(srcs []src, a9 src, ext string, level1 int) {
	level := level1 - 1
	lvOvl, err := os.ReadFile(fmt.Sprintf("%s/ovl9_%03d_dec.bin", ext, levelOverlay(a9, level)))
	if err != nil {
		log.Fatal(err)
	}
	oid := levelOverlay(a9, level)
	var base uint32
	for _, s := range srcs {
		if s.name == fmt.Sprintf("ovl%d", oid) {
			base = s.base
		}
	}
	c := ctx{own: src{"lv", base, lvOvl}, a9: a9}
	hdrRAM, _ := c.word(0x02092208 + uint32(level)*4)
	fmt.Printf("level %d overlay %d settings %08X\n", level, oid, hdrRAM)
	misc, _ := c.word(hdrRAM + 4)
	areaPtr, _ := c.word(hdrRAM + 0x10)
	nb, _ := c.at(hdrRAM + 0x14)
	nArea := int(nb[0])
	walk := func(label string, t uint32) {
		nw, ok := c.at(t)
		if !ok {
			return
		}
		n := int(le.Uint16(nw))
		p, _ := c.word(t + 4)
		fmt.Printf("  %s table %08X: %d entries, list %08X\n", label, t, n, p)
		for i := 0; i < n; i++ {
			e, ok := c.at(p + uint32(i)*8)
			if !ok {
				break
			}
			typ, cnt, layer := e[0]&0x1F, e[1], e[0]>>5
			lp := le.Uint32(e[4:])
			fmt.Printf("    type %2d layer %d count %3d list %08X", typ, layer, cnt, lp)
			if typ == 11 || typ == 6 {
				if d, ok := c.at(lp); ok {
					fmt.Printf("   first words: %08X %08X %08X", le.Uint32(d), le.Uint32(d[4:]), le.Uint32(d[8:]))
				}
			}
			fmt.Println()
		}
	}
	walk("misc", misc)
	for a := 0; a < nArea; a++ {
		t, _ := c.word(areaPtr + uint32(a)*12)
		walk(fmt.Sprintf("area%d", a), t)
	}
}

func levelOverlay(a9 src, level int) int {
	// levelOvlTable: ARM9 file offset 0x718C8, u32 per level (sm64ds/level.go)
	return int(le.Uint32(a9.data[0x718C8+level*4:]))
}

func listPars(rom, ext string, level int) {
	ls, err := sm64ds.OpenLevels(rom, ext)
	if err != nil {
		log.Fatal(err)
	}
	lv, err := ls.Level(level)
	if err != nil {
		log.Fatal(err)
	}
	for _, o := range lv.Objects {
		if o.Simple {
			continue
		}
		fmt.Printf("  actor %3d objid %3d layer %d par1 %04X (hi %02X lo %02X) par2 %04X par3 %04X  pos %.0f,%.0f,%.0f\n",
			o.Actor, o.ID, o.Layer, o.Params[0], o.Params[0]>>8, o.Params[0]&0xFF,
			o.Params[1], o.Params[2], o.X, o.Y, o.Z)
	}
}

// starGates are the save-star predicate and its two wrappers:
//
//	$020137E0(course, star) = saveStars[$0209CAB4 + course] & (1 << star)
//	$020136F8(level, star)  = $020137E0(Course(level), star)
//	$0202A6C8(star)         = $020136F8(currentLevel @ $0209F2F8, star)
var starGates = map[uint32]string{0x020137E0: "isStarCollected", 0x020136F8: "byLevel", 0x0202A6C8: "currentLevel"}

// sweepStarGates walks every actor that appears as a placement, resolves its
// profile through the engine's own $02090864 table inside the overlay the
// actor oracle recorded for it, and looks for a call to the star predicate in
// its vtable methods — directly, and one call deep.
func sweepStarGates(srcs []src, a9 src, rom, ext string) {
	sweepCallers(srcs, a9, rom, ext, starGates, "the save's star bits")
}

// sweepCallers walks every actor that appears as a placement, resolves its
// profile through the engine's own $02090864 table inside the overlay the
// actor oracle recorded for it, and looks for a call to any of `targets` in
// its vtable methods — directly, and one call deep.
func sweepCallers(srcs []src, a9 src, rom, ext string, targets map[uint32]string, what string) {
	ls, err := sm64ds.OpenLevels(rom, ext)
	if err != nil {
		log.Fatal(err)
	}
	placedIn := map[int]map[string]bool{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		for _, o := range lv.Objects {
			if placedIn[o.Actor] == nil {
				placedIn[o.Actor] = map[string]bool{}
			}
			placedIn[o.Actor][stem] = true
		}
	}
	var actors []int
	for a := range placedIn {
		actors = append(actors, a)
	}
	sort.Ints(actors)

	reached, hits := 0, 0
	gatedLevels := map[string]bool{}
	stages := map[string]bool{}
	for a := range placedIn {
		for st := range placedIn[a] {
			stages[st] = true
		}
	}
	nStage := len(stages)
	for _, a := range actors {
		var got []string
		var where string
		for _, s := range srcs {
			c := ctx{own: s, a9: a9}
			pp, ok := c.word(0x02090864 + uint32(a)*4)
			if !ok || !c.isCode(pp) {
				continue
			}
			if w, ok := c.word(pp + 4); !ok || int(w&0xFFFF) != a {
				continue
			}
			create, _ := c.word(pp)
			vt := vtableOf(c, create)
			if vt == 0 {
				continue
			}
			where = s.name
			// Breadth-first from the actor's create and every vtable slot,
			// recording the DEPTH a target is first reached at — a hit two
			// helpers down is weaker evidence than one in the actor's own step,
			// and the depth is what says which.
			frontier := []uint32{create}
			for k := 0; k < 16; k++ {
				if f, ok := c.word(vt + uint32(k*4)); ok && c.isCode(f) {
					frontier = append(frontier, f)
				}
			}
			seenFn := map[uint32]bool{}
			for depth := 0; depth <= callDepth && len(frontier) > 0; depth++ {
				var next []uint32
				for _, f := range frontier {
					if seenFn[f] {
						continue
					}
					seenFn[f] = true
					for _, t := range blTargets(c, f, 400) {
						if n, ok := targets[t]; ok {
							got = append(got, fmt.Sprintf("%s@%d", n, depth))
						}
						next = append(next, t)
					}
				}
				frontier = next
			}
			break
		}
		if where == "" {
			continue
		}
		reached++
		if len(got) == 0 {
			continue
		}
		hits++
		for st := range placedIn[a] {
			gatedLevels[st] = true
		}
		var lv []string
		for s := range placedIn[a] {
			lv = append(lv, s)
		}
		sort.Strings(lv)
		if len(lv) > 6 {
			lv = append(lv[:6], "...")
		}
		fmt.Printf("  actor %3d  %-7s calls %s (%s)   in %s\n",
			a, where, what, uniq(got), strings.Join(lv, " "))
	}
	fmt.Printf("%d placed actors, %d resolved to a vtable, %d call %s\n",
		len(actors), reached, hits, what)
	fmt.Printf("%d of the %d stages place at least one of them\n", len(gatedLevels), nStage)
}

func uniq(in []string) string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return strings.Join(out, ",")
}

func blTargets(c ctx, a uint32, n int) []uint32 {
	code, ok := c.at(a)
	if !ok {
		return nil
	}
	var out []uint32
	for k := 0; k < n && (k+1)*4 <= len(code); k++ {
		in := arm.DecodeARM(code[k*4:], a+uint32(k*4))
		if in.HasTarget && strings.HasPrefix(in.Mnem, "BL") {
			out = append(out, in.Target)
		}
		if in.Text == "BX lr" && k > 2 {
			break
		}
	}
	return out
}

// showMsg resolves external message IDs the way the message window does
// (sm64ds.MsgIndex, the range table at $0208EEEC) and prints the text.
func showMsg(rom, ext, ids string) {
	ls, err := sm64ds.OpenLevels(rom, ext)
	if err != nil {
		log.Fatal(err)
	}
	msgs, err := sm64ds.LoadBMG(filepath.Join(ext, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		log.Fatal(err)
	}
	for _, tok := range strings.Split(ids, ",") {
		var id int
		fmt.Sscanf(strings.TrimSpace(tok), "%d", &id)
		idx := ls.MsgIndex(id)
		if idx < 0 || idx >= len(msgs) {
			fmt.Printf("id %d -> index %d: OUT OF RANGE\n", id, idx)
			continue
		}
		fmt.Printf("id %d -> index %d:\n%s\n\n", id, idx, msgs[idx])
	}
}

// actorMsgSurvey tests the hypothesis "this actor's par1 is a message ID" the
// only way that means anything: resolve it for EVERY placement of the actor in
// the game and see whether they all land on real, apt messages.
func actorMsgSurvey(rom, ext string, actor int) {
	ls, err := sm64ds.OpenLevels(rom, ext)
	if err != nil {
		log.Fatal(err)
	}
	msgs, err := sm64ds.LoadBMG(filepath.Join(ext, "files/data/message/msg_data_eng.bin"))
	if err != nil {
		log.Fatal(err)
	}
	n, ok := 0, 0
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		for _, o := range lv.Objects {
			if o.Actor != actor {
				continue
			}
			n++
			idx := ls.MsgIndex(o.Params[0])
			text := "OUT OF RANGE"
			if idx >= 0 && idx < len(msgs) {
				ok++
				text = strings.ReplaceAll(msgs[idx], "\n", " ")
				if len(text) > 110 {
					text = text[:110] + "…"
				}
			}
			fmt.Printf("lvl %2d %-20s par1 %5d -> idx %4d  %s\n", id, stem, o.Params[0], idx, text)
		}
	}
	fmt.Printf("%d placements, %d resolve to a message\n", n, ok)
}

// reportDropped counts the placements webexport silently discards: modelFor
// returns "" when the actor oracle recorded no model for the actor, and
// addObjOff then returns without emitting anything at all.
func reportDropped(rom, ext string) {
	ls, err := sm64ds.OpenLevels(rom, ext)
	if err != nil {
		log.Fatal(err)
	}
	var table map[string][]struct {
		Params [3]int   `json:"params"`
		Models []string `json:"models"`
	}
	buf, err := os.ReadFile(filepath.Join(ext, "actorbind.json"))
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(buf, &table); err != nil {
		log.Fatal(err)
	}
	hasModel := func(actor int) bool {
		for _, b := range table[fmt.Sprint(actor)] {
			if len(b.Models) > 0 {
				return true
			}
		}
		return false
	}
	perActor := map[int]int{}
	total, kept, seen := 0, 0, map[string]bool{}
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil || lv.BMDPath == "" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		if seen[stem] {
			continue
		}
		seen[stem] = true
		for _, o := range lv.Objects {
			total++
			if hasModel(o.Actor) {
				kept++
			} else {
				perActor[o.Actor]++
			}
		}
	}
	type row struct {
		actor, n int
	}
	var rows []row
	for a, n := range perActor {
		rows = append(rows, row{a, n})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].n > rows[j].n })
	fmt.Printf("%d placements across the shipped stages; %d have a model, %d are dropped\n",
		total, kept, total-kept)
	for _, r := range rows {
		fmt.Printf("  actor %3d  x%-4d\n", r.actor, r.n)
	}
}
