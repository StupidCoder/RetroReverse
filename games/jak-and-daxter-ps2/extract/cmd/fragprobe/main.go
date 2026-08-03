package main

// fragprobe: probe 1 of the topology hunt. For single fragments, print the
// file-only slot walk (strips.go's model) next to the emulated packet
// sequence (ground truth), and diff the triangle sets. Instrument only —
// the deliverable stays merc/strips.go.
import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
	"retroreverse.com/tools/cpu/vu"
)

type tri [3]int

func norm(t tri) tri {
	sort.Ints(t[:])
	return t
}

// fileWalk reproduces strips.go's reconstruction for one fragment, returning
// the slot table and the triangles (local vertex indices).
func fileWalk(fr *merc.Fragment, verbose bool) []tri {
	vs := fr.Vertices()
	recBase := int(fr.ByteData[12]) * 4
	flagOf := func(v int) byte {
		o := recBase + v*4 + 3
		if o < len(fr.ByteData) {
			return fr.ByteData[o]
		}
		return 0x80
	}
	slots := map[int]int{}
	term := map[int]bool{}
	for i, v := range vs {
		low := v.D1 < 7 || v.D2 < 7
		if v.Slot1 >= 0 {
			slots[v.Slot1] = i
			if low {
				term[v.Slot1] = true
			}
		}
		if v.Slot2 >= 0 && v.Slot2 != v.Slot1 {
			slots[v.Slot2] = i
			if low {
				term[v.Slot2] = true
			}
		}
	}
	mx := -1
	for s := range slots {
		if s > mx {
			mx = s
		}
	}
	if verbose {
		fmt.Printf("  file slots (slot:vert flag[T=term]):\n    ")
		for s := 0; s <= mx; s++ {
			v, ok := slots[s]
			if !ok {
				fmt.Printf("%d:-- ", s)
				continue
			}
			t := ""
			if term[s] {
				t = "T"
			}
			fmt.Printf("%d:%d/%02x%s ", s, v, flagOf(v), t)
			if (s+1)%10 == 0 {
				fmt.Printf("\n    ")
			}
		}
		fmt.Println()
	}
	var tris []tri
	var w [3]int
	n := 0
	for s := 0; s <= mx; s++ {
		v, ok := slots[s]
		if !ok {
			n = 0
			continue
		}
		w[0], w[1], w[2] = w[1], w[2], v
		if n < 2 {
			n++
		} else if flagOf(v) == 0x80 {
			if w[0] != w[1] && w[1] != w[2] && w[0] != w[2] {
				tris = append(tris, tri{w[0], w[1], w[2]})
			}
		}
		if term[s] {
			n = 0
		}
	}
	return tris
}

// emuTris converts a global TopoVert stream into triangles by strip
// semantics: every non-ADC vertex after the window fills kicks a triangle
// with its two predecessors (ADC verts advance the window without kicking).
func emuTris(seq []merc.TopoVert) []tri {
	var tris []tri
	var w [3]int
	n := 0
	for _, tv := range seq {
		if tv.Index < 0 {
			n = 0 // unresolved vertex: break rather than invent
			continue
		}
		w[0], w[1], w[2] = w[1], w[2], tv.Index
		if n < 2 {
			n++
			continue
		}
		if !tv.ADC && w[0] != w[1] && w[1] != w[2] && w[0] != w[2] {
			tris = append(tris, tri{w[0], w[1], w[2]})
		}
	}
	return tris
}

func main() {
	obj, err := os.ReadFile(os.Args[1])
	check(err)
	micro, err := os.ReadFile(os.Args[2])
	check(err)
	effIdx := 0
	maxFrag := 6
	if len(os.Args) > 3 {
		effIdx, _ = strconv.Atoi(os.Args[3])
	}
	if len(os.Args) > 4 {
		maxFrag, _ = strconv.Atoi(os.Args[4])
	}
	ctrlOff := uint32(0x1244)
	if v := os.Getenv("CTRL"); v != "" {
		n, _ := strconv.ParseUint(v, 0, 32)
		ctrlOff = uint32(n)
	}
	c, err := merc.Parse(obj, ctrlOff)
	check(err)
	if len(os.Args) > 3 && os.Args[3] == "-trace43" {
		trace43(obj, micro, c)
		return
	}
	e := &c.Effects[effIdx]
	fmt.Printf("effect %d: %d frags, records say %d tris, %d dverts\n",
		effIdx, e.FragCount, e.TriCount, e.DVertCount)

	low := merc.DefaultLowMem()
	stMagic := u32(obj, ctrlOff+44)
	sess, err := merc.NewSession(micro, low, obj[ctrlOff+28:ctrlOff+44], stMagic)
	check(err)

	// global base per fragment
	gbase := make([]int, len(e.Fragments)+1)
	for i := range e.Fragments {
		gbase[i+1] = gbase[i] + e.Fragments[i].LumpQWC/3
	}
	// run the whole effect, tagging each harvested vert with arrival order
	var seq []merc.TopoVert
	perFragArrive := make([][]merc.TopoVert, len(e.Fragments))
	for fi := range e.Fragments {
		topo, terr := sess.RunFragment(&e.Fragments[fi], gbase[fi])
		if terr != nil {
			fmt.Printf("f%d: EMU ERR %v\n", fi, terr)
			continue
		}
		perFragArrive[fi] = topo
		seq = append(seq, topo...)
	}
	if fl, ferr := sess.Flush(); ferr == nil {
		seq = append(seq, fl...)
		perFragArrive = append(perFragArrive, fl)
	}

	// ---- per fragment report ----
	locOf := func(g int) (int, int) { // global -> frag, local
		for i := 0; i < len(e.Fragments); i++ {
			if g >= gbase[i] && g < gbase[i+1] {
				return i, g - gbase[i]
			}
		}
		return -1, g
	}
	for fi := 0; fi < maxFrag && fi < len(e.Fragments); fi++ {
		fr := &e.Fragments[fi]
		nv := fr.LumpQWC / 3
		fmt.Printf("\n== frag %d: nv=%d ctrl={%d,%d,%d,%d} hdr[0..15]=% x\n",
			fi, nv, fr.ByteQWC, fr.LumpQWC, fr.FPQWC, len(fr.Mats), fr.ByteData[:16])
		fileWalk(fr, true)
		fmt.Printf("  emu packets arriving during f%d run (frag:local A=adc):\n    ", fi)
		for i, tv := range perFragArrive[fi] {
			a := ""
			if tv.ADC {
				a = "A"
			}
			if tv.Index < 0 {
				fmt.Printf("?%s ", a)
			} else {
				f, l := locOf(tv.Index)
				fmt.Printf("%d:%d%s ", f, l, a)
			}
			if (i+1)%12 == 0 {
				fmt.Printf("\n    ")
			}
		}
		fmt.Println()
	}

	// ---- per-slot ADC vs lump control-byte correlation (frag0 only) ----
	// The packet arrival order IS slot order, so emu entry s = slot s.
	if len(perFragArrive) > 0 {
		fr := &e.Fragments[0]
		vs0 := fr.Vertices()
		slots := map[int][]int{}
		for i, v := range vs0 {
			if v.Slot1 >= 0 {
				slots[v.Slot1] = append(slots[v.Slot1], i)
			}
			if v.Slot2 >= 0 && v.Slot2 != v.Slot1 {
				slots[v.Slot2] = append(slots[v.Slot2], i)
			}
		}
		fmt.Println("\nfrag0 per-slot: slot vert q0b0 q0b1 q2b0 q2b1 D1 D2 -> ADC")
		for s, tv := range perFragArrive[0] {
			vi := -1
			if list, ok := slots[s]; ok {
				vi = list[0]
			}
			mark := "."
			if tv.ADC {
				mark = "A"
			}
			if vi >= 0 {
				b := fr.LumpData[vi*12 : vi*12+12]
				second := ""
				if vs0[vi].Slot2 >= 0 && vs0[vi].Slot2 != vs0[vi].Slot1 && vs0[vi].Slot2 == s {
					second = "2nd"
				} else if vs0[vi].Slot1 == s {
					second = "1st"
				}
				fmt.Printf("  s%02d v%02d  %02x %02x  %02x %02x  d=%3d,%3d %s %s\n",
					s, vi, b[0], b[1], b[8], b[9], b[4], b[5], second, mark)
			} else {
				fmt.Printf("  s%02d v?? (no file write) emu=%d %s\n", s, tv.Index, mark)
			}
		}
	}

	// ---- whole-effect triangle diff ----
	emuT := emuTris(seq)
	emuSet := map[tri]bool{}
	for _, t := range emuT {
		emuSet[norm(t)] = true
	}
	fileSet := map[tri]bool{}
	for _, t := range emuTris(merc.EffectSequence(e)) {
		fileSet[norm(t)] = true
	}
	both, fileOnly, emuOnly := 0, 0, 0
	var fileOnlyEx, emuOnlyEx []tri
	for t := range fileSet {
		if emuSet[t] {
			both++
		} else {
			fileOnly++
			if len(fileOnlyEx) < 15 {
				fileOnlyEx = append(fileOnlyEx, t)
			}
		}
	}
	for t := range emuSet {
		if !fileSet[t] {
			emuOnly++
			if len(emuOnlyEx) < 15 {
				emuOnlyEx = append(emuOnlyEx, t)
			}
		}
	}
	fmt.Printf("\nEFFECT %d TRI DIFF: emu=%d file=%d agree=%d file-only=%d emu-only=%d\n",
		effIdx, len(emuSet), len(fileSet), both, fileOnly, emuOnly)
	pr := func(name string, ts []tri) {
		fmt.Printf("  %s:", name)
		for _, t := range ts {
			f0, l0 := locOf(t[0])
			f1, l1 := locOf(t[1])
			f2, l2 := locOf(t[2])
			fmt.Printf(" (%d:%d,%d:%d,%d:%d)", f0, l0, f1, l1, f2, l2)
		}
		fmt.Println()
	}
	pr("file-only", fileOnlyEx)
	pr("emu-only", emuOnlyEx)

	// ---- sequence-level diff: first divergences ----
	fseq := merc.EffectSequence(e)
	fmt.Printf("\nseq lens: emu=%d file=%d\n", len(seq), len(fseq))
	shown := 0
	off := 0
	for i := 0; i < len(seq) && i+off < len(fseq) && shown < 40; i++ {
		if seq[i].Index < 0 {
			continue
		}
		a, b := seq[i], fseq[i+off]
		if a.Index != b.Index {
			// try resync: emu has extra entries (deferred kicks may repeat)
			af, al := locOf(a.Index)
			bf, bl := locOf(b.Index)
			fmt.Printf("  pos %d(+%d): emu=%d:%d file=%d:%d -- resyncing\n", i, off, af, al, bf, bl)
			shown++
			// slide file cursor back one to retry emu's next against same file pos
			off--
			continue
		}
		if a.ADC != b.ADC {
			af, al := locOf(a.Index)
			am := ""
			if a.ADC {
				am = "A"
			}
			fmt.Printf("  pos %d(+%d): ADC diff %d:%d emu=%s\n", i, off, af, al, am)
			shown++
		}
	}
	if v := os.Getenv("ADDRWIN"); v != "" {
		var lo, hi int
		fmt.Sscanf(v, "%d:%d", &lo, &hi)
		for i := lo; i <= hi && i < len(seq); i++ {
			af, al := locOf(seq[i].Index)
			m := ""
			if seq[i].ADC {
				m = "A"
			}
			fmt.Printf("  seq[%d] = %d:%d%s addr=%d\n", i, af, al, m, seq[i].Addr)
		}
	}
	fmt.Printf("  emu tail:")
	for i := len(seq) - 16; i < len(seq); i++ {
		if i < 0 {
			continue
		}
		af, al := locOf(seq[i].Index)
		m := ""
		if seq[i].ADC {
			m = "A"
		}
		fmt.Printf(" %d:%d%s", af, al, m)
	}
	fmt.Printf("\n  file tail:")
	for i := len(fseq) - 16; i < len(fseq); i++ {
		if i < 0 {
			continue
		}
		af, al := locOf(fseq[i].Index)
		m := ""
		if fseq[i].ADC {
			m = "A"
		}
		fmt.Printf(" %d:%d%s", af, al, m)
	}
	fmt.Println()
	_ = shown
	shown = 0
	for i := 0; i < len(seq) && i < len(fseq) && shown < 25; i++ {
		a, b := seq[i], fseq[i]
		if a.Index != b.Index || a.ADC != b.ADC {
			af, al := locOf(a.Index)
			bf, bl := locOf(b.Index)
			am, bm := "", ""
			if a.ADC {
				am = "A"
			}
			if b.ADC {
				bm = "A"
			}
			fmt.Printf("  pos %d: emu=%d:%d%s file=%d:%d%s\n", i, af, al, am, bf, bl, bm)
			shown++
		}
	}
}

// trace43 runs frag0 standalone (top=0, output vertices from qw 378) and logs
// every sq store into the slot-42..44 quadword range, naming the storing PC.
func trace43(obj, micro []byte, c *merc.Ctrl) {
	fr := &c.Effects[0].Fragments[0]
	cfg := &merc.EmuConfig{
		Micro: micro, LowMem: merc.DefaultLowMem(),
		Bones: merc.IdentityBones(fr),
		Entry: 0x88, Init: true, Top: 0,
		ParseAt: uint32(371 + fr.FPQWC),
		STMagic: u32(obj, 0x1244+44),
	}
	// Log everything the stitch tail does: loads and stores with addresses,
	// plus the table pointers at loop entry.
	var lastPC uint32 = 0xFFFF
	cfg.TracePC = func(v *vu.VU, pc uint32) {
		if pc >= 0x3900 && pc <= 0x3F18 && pc != lastPC+8 {
			fmt.Printf("JUMP -> %04X (from %04X) vi04=%d vi05=%d vi06=%d vi07=%d vi08=%d vi09=%d vi10=%d\n",
				pc, lastPC, int16(v.VI[4]), int16(v.VI[5]), int16(v.VI[6]), int16(v.VI[7]),
				int16(v.VI[8]), int16(v.VI[9]), int16(v.VI[10]))
		}
		if pc >= 0x3900 && pc <= 0x3F18 {
			lastPC = pc
		}
		if pc < 0x3A00 || pc > 0x3E98 || int(pc)+8 > len(micro) {
			return
		}
		if pc == 0x3A00 || pc == 0x3C08 || pc == 0x3D20 {
			fmt.Printf("pc %04X: vi02=%d vi04=%d vi05=%d vi06=%d vi07=%d vi08=%d vi09=%d vi10=%d vi15=%d\n",
				pc, int16(v.VI[2]), int16(v.VI[4]), int16(v.VI[5]), int16(v.VI[6]),
				int16(v.VI[7]), int16(v.VI[8]), int16(v.VI[9]), int16(v.VI[10]), int16(v.VI[15]))
		}
		raw := binary.LittleEndian.Uint64(micro[pc:])
		txt := vu.Decode(raw, pc).Lower
		var vf, off, viN int
		if n, _ := fmt.Sscanf(txt, "sq.xyzw vf%d, %d(vi%d)", &vf, &off, &viN); n == 3 {
			addr := int(int16(v.VI[viN])) + off
			fmt.Printf("  pc %04X: %-28s -> qw %d (slot %d.%d)\n", pc, txt, addr, (addr-378)/3, (addr-378)%3)
		}
		if n, _ := fmt.Sscanf(txt, "lq.xyzw vf%d, %d(vi%d)", &vf, &off, &viN); n == 3 && pc >= 0x3D20 {
			addr := int(int16(v.VI[viN])) + off
			fmt.Printf("  pc %04X: %-28s <- qw %d (slot %d.%d)\n", pc, txt, addr, (addr-378)/3, (addr-378)%3)
		}
	}
	vs, err := merc.EmulateTopology(cfg, fr)
	check(err)
	for i, v := range vs {
		if i >= 41 && i <= 46 {
			fmt.Printf("out s%d: idx=%d adc=%v\n", i, v.Index, v.ADC)
		}
	}
}

func u32(b []byte, o uint32) uint32 {
	return uint32(b[o]) | uint32(b[o+1])<<8 | uint32(b[o+2])<<16 | uint32(b[o+3])<<24
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "fragprobe:", err)
		os.Exit(1)
	}
}
