package ps2

// vif.go is the VIF — the VPU Interface, the unit that stands between a DMA channel and
// a vector unit. There are two: VIF0 feeds VU0, and VIF1 feeds VU1 and doubles as the
// second road into the GS (PATH2), which is what makes it part of the render path even
// before the vector unit exists.
//
// The stream is 32-bit VIFcodes, each optionally followed by payload:
//
//	bits  0..15  IMMEDIATE   an argument
//	bits 16..23  NUM         a count
//	bits 24..30  CMD         the command
//	bit  31      the interrupt request
//
// What this machine does with each command is honest about what exists. The decode
// state (STCYCL, STMASK, STROW/STCOL) is kept; microcode (MPG) and data (UNPACK) are
// stored into the vector unit's two memories so they are there the day it runs; MSCAL —
// "run the microprogram" — is counted, because there is nothing to run it on yet; and
// DIRECT, the PATH2 door into the GS, is fed to the GIF parser whole. The census is the
// work list: it says how much of a frame goes each road, which is the only fact that
// says whether the first triangle needs a vector unit or just a rasteriser.
//
// A command's payload routinely spans DMA chain links (a REF tag's quadwords finish a
// DIRECT another link began), so the parser is a stream: what a feed cannot finish it
// carries, and the next feed completes.

import (
	"math"

	"retroreverse.com/tools/cpu/vu"
)

// f32 reads a little-endian float out of a byte slice, for the input-buffer dump.
func f32(b []byte) float32 { return math.Float32frombits(le32gs(b)) }

// The VIFcode commands, by number.
const (
	vifNOP      = 0x00
	vifSTCYCL   = 0x01
	vifOFFSET   = 0x02 // VIF1 only
	vifBASE     = 0x03 // VIF1 only
	vifITOP     = 0x04
	vifSTMOD    = 0x05
	vifMSKPATH3 = 0x06 // VIF1 only
	vifMARK     = 0x07
	vifFLUSHE   = 0x10
	vifFLUSH    = 0x11 // VIF1 only
	vifFLUSHA   = 0x13 // VIF1 only
	vifMSCAL    = 0x14
	vifMSCALF   = 0x15
	vifMSCNT    = 0x17
	vifSTMASK   = 0x20
	vifSTROW    = 0x30
	vifSTCOL    = 0x31
	vifMPG      = 0x4A
	vifDIRECT   = 0x50 // VIF1 only: IMMEDIATE quadwords straight to the GS (PATH2)
	vifDIRECTHL = 0x51 // VIF1 only: the same, yielding to PATH3's IMAGE mode
)

// The vector units' memories. VU1 has 16 KiB of each; VU0 has 4 KiB of each. They are
// filled by MPG and UNPACK and will be read the day the unit executes.
const (
	vu0MemSize = 4 << 10
	vu1MemSize = 16 << 10
)

// vif is one VPU interface's state.
type vif struct {
	idx int // 0 or 1
	m   *Machine

	// The sticky decode registers, each set by the command of the same name.
	cl, wl     uint32 // STCYCL: the write cycle
	mode       uint32 // STMOD: the addition mode UNPACK applies
	mask       uint32 // STMASK: the write mask
	row, col   [4]uint32
	base, ofst uint32 // VIF1's double buffer registers
	itop, mark uint32

	// The double-buffer state MSCAL owns: UNPACK writes land at TOPS (with the FLG
	// bit); MSCAL hands TOPS to the program as TOP (read by XTOP) and flips TOPS to
	// the other buffer. itop above is the staging ITOPS the same way.
	tops uint32

	// The vector unit itself (VIF1's runs; VU0's macro mode belongs to the EE's COP2).
	vu         *vu.VU
	vuSteps    uint64
	kickDumped bool
	lastStart  uint32 // the last MSCAL'd program address, naming XGKICK packets' producer
	dumpN      int    // how many VU1DumpIn snapshots have been written

	// The command in flight: its code word, the bytes its payload still needs, and the
	// payload gathered so far.
	cmd     uint32
	pending int
	buf     []byte

	// The vector unit's memories, filled by MPG and UNPACK.
	micro []byte
	data  []byte

	// The census: every command seen, counted; and each distinct microprogram start
	// MSCAL asks for, because "which programs does a frame run" is the specification
	// for the vector unit this machine does not have yet.
	census map[string]int
	mscal  map[uint32]int

	// Each distinct MPG payload, keyed by content hash. A frame re-uploads its
	// microcode constantly (half a million MPGs a boot), but the set of distinct
	// programs is small — and that set, not the upload count, is the size of the
	// vector unit's workload.
	mpgSeen map[uint64]*mpgInfo
}

// mpgInfo is one distinct microprogram the stream delivered.
type mpgInfo struct {
	addr  uint32 // VU micro byte address it loads at
	size  int
	count int
}

// ensureVIF creates a VPU interface — and its vector unit — the first time anything
// needs it. VU0 doubles as the EE's COP2: the same instance the VIF fills is the one
// the EE's macro-mode instructions execute on, which is the whole point — GOAL runs its
// matrix and lighting math through COP2, and the results land in registers a VCALLMS
// program then reads.
func (m *Machine) ensureVIF(idx int) *vif {
	if m.vifs[idx] == nil {
		size := uint32(vu0MemSize)
		if idx == 1 {
			size = vu1MemSize
		}
		v := &vif{
			idx: idx, m: m, cl: 1, wl: 1,
			micro:   make([]byte, size),
			data:    make([]byte, size),
			census:  map[string]int{},
			mscal:   map[uint32]int{},
			mpgSeen: map[uint64]*mpgInfo{},
		}
		v.vu = vu.New(v.micro, v.data)
		if idx == 1 {
			v.vu.XGKick = v.xgkick
		} else {
			m.CPU.COP2 = v.vu
		}
		m.vifs[idx] = v
	}
	return m.vifs[idx]
}

// vifStart runs a transfer the DMA controller has handed a VIF channel.
func (m *Machine) vifStart(idx int, c *dmacChan) {
	v := m.ensureVIF(idx)
	switch (c.chcr & dChcrModeM) >> 2 {
	case 1: // source chain — the render path's shape
		m.dmacSourceChain(dmacChanForVIF(idx), c, v.feed, false)
	default:
		if c.qwc == 0 {
			return
		}
		v.feed(m.dmaBytes(c.madr, c.qwc))
	}
}

func dmacChanForVIF(idx int) int {
	if idx == 1 {
		return dmacChVIF1
	}
	return dmacChVIF0
}

// feed advances the parser over one stretch of the stream.
func (v *vif) feed(data []byte) {
	i := 0
	for i < len(data) {
		if v.pending > 0 {
			n := v.pending
			if n > len(data)-i {
				n = len(data) - i
			}
			v.buf = append(v.buf, data[i:i+n]...)
			v.pending -= n
			i += n
			if v.pending == 0 {
				v.finish()
			}
			continue
		}
		if len(data)-i < 4 {
			return // a stray tail; codes are word-aligned and a chain link is quadwords
		}
		w := le32gs(data[i:])
		i += 4
		v.code(w)
	}
}

// code decodes one VIFcode and either acts on it or arms the payload gather.
func (v *vif) code(w uint32) {
	cmd := (w >> 24) & 0x7F
	num := (w >> 16) & 0xFF
	imm := w & 0xFFFF

	switch cmd {
	case vifNOP:

	case vifSTCYCL:
		v.cl = imm & 0xFF
		v.wl = imm >> 8
		if v.cl == 0 {
			v.cl = 1 // a zero cycle divides by it below; the hardware treats CL=0 as unusable
		}

	case vifOFFSET:
		v.ofst = imm
	case vifBASE:
		v.base = imm
		v.tops = imm // TOPS starts at BASE; MSCAL flips it between the two buffers
	case vifITOP:
		v.itop = imm
	case vifSTMOD:
		v.mode = imm & 3
	case vifMSKPATH3:
		v.count("mskpath3")
	case vifMARK:
		v.mark = imm

	case vifFLUSHE, vifFLUSH, vifFLUSHA:
		// Wait for the vector unit (and, for FLUSHA, PATH3) to finish. Nothing runs, so
		// nothing waits.

	case vifMSCAL, vifMSCALF:
		// Run the microprogram at IMMEDIATE (in 64-bit units).
		v.count("mscal")
		if v.mscal[imm]++; v.mscal[imm] == 1 {
			v.m.note("VIF%d: MSCAL of the microprogram at 0x%X (VU%d micro address 0x%X)",
				v.idx, imm, v.idx, imm*8)
		}
		v.runVU(imm*8, false)
	case vifMSCNT:
		// Run the microprogram from where the last one ended.
		v.count("mscnt")
		v.runVU(0, true)

	case vifSTMASK:
		v.arm(w, 4)
	case vifSTROW, vifSTCOL:
		v.arm(w, 16)

	case vifMPG:
		n := num
		if n == 0 {
			n = 256
		}
		v.arm(w, int(n)*8)

	case vifDIRECT, vifDIRECTHL:
		n := uint32(imm)
		if n == 0 {
			n = 0x10000
		}
		v.arm(w, int(n)*16)

	default:
		if cmd >= 0x60 && cmd < 0x80 {
			v.arm(w, v.unpackBytes(cmd, num))
			return
		}
		v.count(sprintf("cmd 0x%02X unknown", cmd))
	}
}

// arm starts gathering a command's payload.
func (v *vif) arm(w uint32, n int) {
	v.cmd = w
	v.pending = n
	v.buf = v.buf[:0]
	if n == 0 {
		v.finish()
	}
}

// finish acts on a command whose payload has fully arrived.
func (v *vif) finish() {
	cmd := (v.cmd >> 24) & 0x7F
	imm := v.cmd & 0xFFFF

	switch cmd {
	case vifSTMASK:
		v.mask = le32gs(v.buf)
	case vifSTROW:
		for i := 0; i < 4; i++ {
			v.row[i] = le32gs(v.buf[i*4:])
		}
	case vifSTCOL:
		for i := 0; i < 4; i++ {
			v.col[i] = le32gs(v.buf[i*4:])
		}

	case vifMPG:
		// The microcode lands in the unit's program memory at IMMEDIATE*8, so it is there
		// to disassemble now and to run later.
		v.count("mpg")
		copy(v.micro[min32(imm*8, uint32(len(v.micro))):], v.buf)
		h := fnv64(v.buf)
		if info := v.mpgSeen[h]; info != nil {
			info.count++
		} else {
			v.mpgSeen[h] = &mpgInfo{addr: imm * 8, size: len(v.buf), count: 1}
		}

	case vifDIRECT, vifDIRECTHL:
		// PATH2: the payload is GIF packets, straight to the GS.
		v.count("direct")
		v.m.ensureGS().src = "path2 direct"
		v.m.gifPacket(v.buf)

	default:
		if cmd >= 0x60 && cmd < 0x80 {
			v.unpack(cmd, v.buf)
		}
	}
	v.buf = v.buf[:0]
}

// unpackBytes is the source length of an UNPACK: how many bytes of stream the command
// consumes, which depends on the format in the command and on the STCYCL cycle.
//
// The format is in the low bits of the command: vn (bits 2..3) is the vector width minus
// one, vl (bits 0..1) the element size — 32, 16 or 8 bits, with vl=3 the special V4-5551,
// one halfword unpacking to four elements. NUM is how many vectors are WRITTEN; when the
// write cycle is longer than the read cycle (WL > CL) the unit fills the gap from ROW, so
// the stream supplies data only for CL of every WL writes.
func (v *vif) unpackBytes(cmd, num uint32) int {
	if num == 0 {
		num = 256
	}
	vn := (cmd >> 2) & 3
	vl := cmd & 3

	vectors := num
	if v.wl > v.cl {
		full := num / v.wl
		rem := num % v.wl
		if rem > v.cl {
			rem = v.cl
		}
		vectors = full*v.cl + rem
	}

	var bytesPer uint32
	if vl == 3 { // V4-5551: one halfword per vector
		bytesPer = 2
	} else {
		bytesPer = (vn + 1) * (4 >> vl)
	}
	n := vectors * bytesPer
	return int((n + 3) &^ 3) // the stream advances in whole words
}

// unpack writes an UNPACK's payload into the vector unit's data memory. The formats are
// expanded faithfully — each vector becomes four 32-bit fields at 16 bytes per vector —
// because this memory is what the microprogram will read the day it runs, and data
// placed wrongly now would be a bug found then.
func (v *vif) unpack(cmd uint32, data []byte) {
	imm := v.cmd & 0xFFFF
	num := (v.cmd >> 16) & 0xFF
	if num == 0 {
		num = 256
	}
	vn := (cmd >> 2) & 3
	vl := cmd & 3
	usn := v.cmd&(1<<14) != 0
	flg := v.cmd&(1<<15) != 0

	cyc := ""
	if v.wl != v.cl {
		cyc = sprintf(" cl%d wl%d", v.cl, v.wl)
	}
	v.count(sprintf("unpack V%d-%d%s", vn+1, 32>>vl, cyc))
	if v.mask != 0 && v.cmd&(1<<28) != 0 {
		v.count(sprintf("unpack with STMASK %08X (mask not applied)", v.mask))
	}

	addr := (imm & 0x3FF) * 16
	if flg {
		addr += v.tops * 16 // the double buffer MSCAL is filling next (TOPS)
	}

	// The read cursor over the payload and the element reader for the format.
	pos := 0
	read := func() uint32 {
		var e uint32
		switch vl {
		case 0: // 32-bit
			if pos+4 <= len(data) {
				e = le32gs(data[pos:])
			}
			pos += 4
		case 1: // 16-bit
			if pos+2 <= len(data) {
				e = uint32(data[pos]) | uint32(data[pos+1])<<8
				if !usn && e&0x8000 != 0 {
					e |= 0xFFFF0000
				}
			}
			pos += 2
		case 2: // 8-bit
			if pos < len(data) {
				e = uint32(data[pos])
				if !usn && e&0x80 != 0 {
					e |= 0xFFFFFF00
				}
			}
			pos++
		}
		return e
	}

	wl := v.wl
	if wl == 0 {
		wl = v.cl
	}
	for i := uint32(0); i < num; i++ {
		var f [4]uint32
		inCycle := wl <= v.cl || i%wl < v.cl
		switch {
		case vl == 3 && vn == 3: // V4-5551: one halfword to four fields
			var h uint32
			if pos+2 <= len(data) {
				h = uint32(data[pos]) | uint32(data[pos+1])<<8
			}
			pos += 2
			f[0] = (h & 0x1F) << 3
			f[1] = (h >> 5 & 0x1F) << 3
			f[2] = (h >> 10 & 0x1F) << 3
			f[3] = (h >> 15) << 7
		case !inCycle:
			// A fill cycle: the stream supplies nothing; ROW supplies everything.
			f = v.row
		default:
			for e := uint32(0); e <= vn; e++ {
				f[e] = read()
			}
			// S-formats broadcast their one element; V2/V3 leave the rest as they are.
			if vn == 0 {
				f[1], f[2], f[3] = f[0], f[0], f[0]
			}
			// STMOD: offset mode adds ROW to what the stream supplied; difference mode
			// also makes the sum the new ROW — a delta-compressed stream reconstructing
			// itself. The addition is a plain 32-bit add, whatever the bits mean.
			last := vn
			if vn == 0 {
				last = 3
			}
			switch v.mode {
			case 1:
				for e := uint32(0); e <= last; e++ {
					f[e] += v.row[e]
				}
			case 2:
				for e := uint32(0); e <= last; e++ {
					f[e] += v.row[e]
					v.row[e] = f[e]
				}
			}
		}
		// The write address honours STCYCL's skip mode: with CL > WL the unit writes
		// WL vectors then skips ahead to the next CL boundary — the road a game uses
		// to interleave separate position/texcoord/colour streams into one vertex
		// array. Writing contiguously here scrambled every interleaved record.
		at := addr + i*16
		if v.cl > wl {
			at = addr + (i/wl*v.cl+i%wl)*16
		}
		if int(at)+16 <= len(v.data) {
			for e := 0; e < 4; e++ {
				v.data[at+uint32(e)*4+0] = byte(f[e])
				v.data[at+uint32(e)*4+1] = byte(f[e] >> 8)
				v.data[at+uint32(e)*4+2] = byte(f[e] >> 16)
				v.data[at+uint32(e)*4+3] = byte(f[e] >> 24)
			}
		}
	}
}

// runVU is the MSCAL family: latch the double-buffer pointers for the program to read
// (XTOP/XITOP), flip TOPS to the buffer the next UNPACKs fill, and run the microprogram
// to its E bit. cont is MSCNT — continue at the address the last program ended at.
//
// The run is synchronous where the hardware's is parallel: on the board the VIF goes on
// feeding while the program runs, and MSCAL *waits* if the previous program hasn't
// ended. Running to completion at the MSCAL is the same ordering contract with the wait
// collapsed to zero.
func (v *vif) runVU(start uint32, cont bool) {
	v.vu.Top = uint16(v.tops)
	v.vu.ITop = uint16(v.itop)
	if v.ofst != 0 {
		if v.tops == v.base {
			v.tops = v.base + v.ofst
		} else {
			v.tops = v.base
		}
	}
	if cont {
		start = v.vu.PC
	}
	v.lastStart = start
	if v.idx == 1 && v.m.VU1DumpIn == int64(start) && v.dumpN < 12 {
		// Whole-memory snapshots of the first dozen runs, for offline analysis: the
		// in-place transform destroys its input, and the batch lives at a rotating
		// buffer base the caller can't predict.
		name := sprintf("vu1in-%02d.bin", v.dumpN)
		v.dumpN++
		snap := append([]byte(nil), v.data...)
		_ = writeFile(name, snap)
		v.m.note("VU1 input snapshot at MSCAL of 0x%X (top %d): %s", start, v.vu.Top, name)
	}
	// The budget is a corrupt-program guard, far above any real program (the biggest
	// ones here are a few thousand steps over a full input buffer).
	steps, ended := v.vu.Run(start, 1<<20)
	v.vuSteps += uint64(steps)
	if !ended {
		v.count("vu1 program hit the step budget")
	}
}

// xgkick is PATH1: the program hands the GS a GIF packet it built in data memory.
func (v *vif) xgkick(qw uint32) {
	v.count("xgkick")
	gs := v.m.ensureGS()
	gs.src = sprintf("vu1 program 0x%X (kick at qw %d, top %d)", v.lastStart, qw, v.vu.Top)
	data := v.data[qw*16:]
	n := gifPacketLen(data)
	gs.srcData = data[:n]
	in := uint32(v.vu.Top) * 16
	if int(in) < len(v.data) {
		gs.srcIn = v.data[in:]
	}
	gs.srcMicro, gs.srcVUData = v.micro, v.data
	if !v.kickDumped {
		// The first packet, once, in the raw: what the microprogram actually builds is
		// the specification for everything downstream of it.
		v.kickDumped = true
		dump := n
		if dump > 192 {
			dump = 192
		}
		for o := 0; o < dump; o += 16 {
			v.m.note("VU1 first XGKICK qw+%d: %08X %08X %08X %08X", o/16,
				le32gs(data[o:]), le32gs(data[o+4:]), le32gs(data[o+8:]), le32gs(data[o+12:]))
		}
	}
	v.m.gifPacket(data[:n])
}

func (v *vif) count(what string) { v.census[sprintf("VIF%d %s", v.idx, what)]++ }

// fnv64 hashes an MPG payload, for the distinct-program census.
func fnv64(b []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range b {
		h = (h ^ uint64(c)) * 1099511628211
	}
	return h
}

// VUMicro returns a vector unit's program memory as the VIF has filled it — what an
// MSCAL would run, and therefore what a disassembler reads to size up building the unit.
func (m *Machine) VUMicro(idx int) []byte {
	if idx < 0 || idx > 1 || m.vifs[idx] == nil {
		return nil
	}
	return m.vifs[idx].micro
}

// VUDataMem returns a vector unit's data memory, the other half of what a microprogram
// sees.
func (m *Machine) VUDataMem(idx int) []byte {
	if idx < 0 || idx > 1 || m.vifs[idx] == nil {
		return nil
	}
	return m.vifs[idx].data
}

// VIFCensus reports what the two VPU interfaces were asked to do — the only account of
// how much of a frame goes to the vector units (PATH1, a unit not built yet) and how
// much comes straight through to the GS (PATH2 DIRECT).
func (m *Machine) VIFCensus() string {
	s := ""
	for _, v := range m.vifs {
		if v == nil {
			continue
		}
		for _, kv := range sortedCounts(v.census) {
			s += sprintf("      %-24s %d\n", kv.name, kv.n)
		}
		if v.vuSteps > 0 {
			s += sprintf("      VU%d instructions run    %d\n", v.idx, v.vuSteps)
		}
		if len(v.mpgSeen) > 0 {
			var progs []*mpgInfo
			for _, i := range v.mpgSeen {
				progs = append(progs, i)
			}
			for i := 1; i < len(progs); i++ { // insertion sort by address, then size
				for j := i; j > 0 && (progs[j].addr < progs[j-1].addr ||
					(progs[j].addr == progs[j-1].addr && progs[j].size < progs[j-1].size)); j-- {
					progs[j], progs[j-1] = progs[j-1], progs[j]
				}
			}
			s += sprintf("      VIF%d distinct microprograms: %d\n", v.idx, len(progs))
			for _, i := range progs {
				s += sprintf("        micro 0x%04X  %5d bytes  uploaded %d times\n", i.addr, i.size, i.count)
			}
		}
	}
	if s == "" {
		return ""
	}
	return "what the VPU interfaces were asked (the render transport):\n" + s
}

type countKV struct {
	name string
	n    int
}

func sortedCounts(m map[string]int) []countKV {
	var all []countKV
	for k, n := range m {
		all = append(all, countKV{k, n})
	}
	for i := 1; i < len(all); i++ { // insertion sort; the list is short
		for j := i; j > 0 && (all[j].n > all[j-1].n || (all[j].n == all[j-1].n && all[j].name < all[j-1].name)); j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	return all
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
