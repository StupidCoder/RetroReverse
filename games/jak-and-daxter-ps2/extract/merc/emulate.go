package merc

// emulate.go runs the merc VU1 microprogram (tools/cpu/vu) on a fragment's
// unpacked input and parses the GIF packet it kicks — the exact vertices,
// ST/RGBA, ADC bits and strip structure, with no hand-derived pipeline
// logic. The VIF side reproduces draw-bones-merc's chain: STROW
// {0x47800000, 0x4B010000, stMagic, stMagic}, STMOD 1 for the two V4-8
// unpacks (row bits + byte, as the VIF adds them), STMOD 0 for the V4-32
// fp region at TOP+140+byteNUM+lumpNUM, bone matrices at their control
// pairs' VU addresses, and MSCAL at the merc entry.

import (
	"encoding/binary"
	"fmt"
	"os"

	"retroreverse.com/tools/cpu/vu"
)

// EmuConfig carries what the runtime would supply: the microcode, the
// low-memory constant block, and the bone matrix palette.
type EmuConfig struct {
	Micro  []byte
	LowMem []byte // 140 quadwords: rows 0-7, camera/lights 132-139 as chosen
	// Bones maps a matrix-palette index to its 7-quadword upload image.
	Bones map[byte][]byte
	Entry uint32 // microcode entry, byte address (0x128 for the merc main entry)
	Init  bool   // run the constant-derivation prologue (entry 0) first
	TracePC func(v *vu.VU, pc uint32)
	DumpMem func(mem []byte)
	ParseAt uint32 // if nonzero, parse the GIF packet here instead of the kick address
	STMagic uint32 // the merc-ctrl's +44 word (the STROW x/y value)
	Top   uint16 // input double-buffer base in quadwords
}

// OutVert is one GIF vertex from the kicked packet.
type OutVert struct {
	ST    [3]float32 // s, t, q
	RGBA  [4]uint8
	XYZ   [3]float32 // fixed 12.4, converted
	ADC   bool
	Addr  uint32
	XYZQW [16]byte // raw XYZF2 quadword, for carried-vertex matching
}

// Emulate runs one fragment and returns the kicked GIF packets' vertices in
// kick order.
func Emulate(cfg *EmuConfig, fr *Fragment, stMagic uint32) ([]OutVert, error) {
	data := make([]byte, 16384)
	copy(data, cfg.LowMem)

	// STROW as draw-bones-merc's chain assembles it: the ctrl+44 word,
	// then 0, 65536.0, 8454144.0 (the tag's vif1 slot is overwritten with
	// ctrl+44 and becomes the first STROW data word).
	row := [4]uint32{stMagic, stMagic, 0x47800000, 0x4B010000}
	top := uint32(cfg.Top)

	put := func(qw uint32, vals [4]uint32) {
		for k, v := range vals {
			binary.LittleEndian.PutUint32(data[qw*16+uint32(k)*4:], v)
		}
	}
	// V4-8 with STMOD 1: each source byte lands in one word as row[k]+b.
	// Byte region: STMOD 0 — raw bytes (t2's vif words are {NOP, UNPACK}).
	addr := top + 140
	for r := 0; r < fr.ByteQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = uint32(fr.ByteData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	for r := 0; r < fr.LumpQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = row[k] + uint32(fr.LumpData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	// V4-32, STMOD 0: raw quadwords.
	for r := 0; r < fr.FPQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = binary.LittleEndian.Uint32(fr.FPData[r*16+k*4:])
		}
		put(addr, w)
		addr++
	}
	// Bone matrices at their control destinations (7 qw each).
	for _, m := range fr.Mats {
		img, ok := cfg.Bones[m.Index]
		if !ok || len(img) < 7*16 {
			return nil, fmt.Errorf("merc emulate: no bone image for palette %d", m.Index)
		}
		copy(data[uint32(m.Dest)*16:], img[:7*16])
	}

	micro := cfg.Micro
	if len(micro) < 16384 {
		m := make([]byte, 16384)
		copy(m, micro)
		micro = m
	}
	v := vu.New(micro, data)
	v.Top = cfg.Top
	v.ResetBranchLog(16)
	if cfg.TracePC != nil {
		v.Trace = func(vm *vu.VU, pc uint32, raw uint64) { cfg.TracePC(vm, pc) }
	}
	var kicks []uint32
	v.XGKick = func(qw uint32) { kicks = append(kicks, qw) }
	if cfg.Init {
		if _, ok := v.Run(0, 4000); !ok {
			return nil, fmt.Errorf("merc emulate: init prologue did not halt; %x", v.BranchTrail())
		}
	}
	halted := false
	if _, ok := v.Run(cfg.Entry, 400000); ok {
		halted = true
	}
	if cfg.DumpMem != nil {
		cfg.DumpMem(v.Data)
	}
	var out []OutVert
	if cfg.ParseAt != 0 {
		kicks = []uint32{cfg.ParseAt}
	}
	for _, k := range kicks {
		vs, err := parseGIF(v.Data, k)
		if err != nil {
			return nil, err
		}
		out = append(out, vs...)
	}
	if !halted {
		return out, fmt.Errorf("merc emulate: did not halt (%d kicks first); last branches %x", len(kicks), v.BranchTrail())
	}
	return out, nil
}

// parseGIF walks a GIF packet in VU memory: PACKED mode with the merc
// register list (ST, RGBAQ, XYZF2); EOP ends the packet.
func parseGIF(mem []byte, qw uint32) ([]OutVert, error) {
	var out []OutVert
	for {
		if qw*16+16 > uint32(len(mem)) {
			return out, nil
		}
		tag := binary.LittleEndian.Uint64(mem[qw*16:])
		nloop := int(tag & 0x7FFF)
		eop := tag&0x8000 != 0
		nreg := int(tag >> 60 & 0xF)
		if nreg == 0 {
			nreg = 16
		}
		regs := binary.LittleEndian.Uint64(mem[qw*16+8:])
		qw++
		for l := 0; l < nloop; l++ {
			var v OutVert
			sawXYZ := false
			for r := 0; r < nreg; r++ {
				reg := regs >> (4 * r) & 0xF
				base := qw * 16
				if base+16 > uint32(len(mem)) {
					return out, nil
				}
				switch reg {
				case 0x2: // ST(Q)
					v.ST[0] = f32bits(binary.LittleEndian.Uint32(mem[base:]))
					v.ST[1] = f32bits(binary.LittleEndian.Uint32(mem[base+4:]))
					v.ST[2] = f32bits(binary.LittleEndian.Uint32(mem[base+8:]))
				case 0x1: // RGBAQ
					v.RGBA = [4]uint8{mem[base], mem[base+4], mem[base+8], mem[base+12]}
				case 0x4: // XYZF2, ADC in bit 15 of the fourth word
					x := binary.LittleEndian.Uint16(mem[base:])
					y := binary.LittleEndian.Uint16(mem[base+4:])
					z := binary.LittleEndian.Uint32(mem[base+8:])
					w := binary.LittleEndian.Uint32(mem[base+12:])
					v.XYZ = [3]float32{float32(int16(x)) / 16, float32(int16(y)) / 16, float32(z >> 4)}
					v.ADC = w&0x8000 != 0
					v.Addr = qw
					copy(v.XYZQW[:], mem[base:base+16])
					sawXYZ = true
				}
				qw++
			}
			if sawXYZ {
				out = append(out, v)
			}
		}
		if eop {
			return out, nil
		}
	}
}

func f32bits(v uint32) float32 { return float32frombits(v) }

// TopoVert is one packet vertex: which input vertex it is, and its ADC bit.
type TopoVert struct {
	Index int
	ADC   bool
	Addr  uint32 // slot address of the vertex's RGBA quadword
}

// EmulateTopology runs the fragment with vertex indices encoded into the
// color-record stream (byte-region records, one per vertex, read one behind
// the UV cursor), then reads them back from the packet's RGBA slots: the
// exact strip order and ADC bits, keyed to input vertices.
func EmulateTopology(cfg *EmuConfig, fr *Fragment) ([]TopoVert, error) {
	patched := *fr
	patched.ByteData = append([]byte(nil), fr.ByteData...)
	// The color cursor starts one quadword after the UV cursor
	// (hdr.byte12+1); record k in that stream colors vertex k.
	base := (int(fr.ByteData[12]) + 1) * 4
	nv := fr.LumpQWC / 3
	for v := 0; base+v*4+3 < len(patched.ByteData) && v < nv+2; v++ {
		patched.ByteData[base+v*4] = byte(v + 1) // +1: 0 is unused (visible vs zero-fill)
		patched.ByteData[base+v*4+1] = 0
		patched.ByteData[base+v*4+2] = 0
		patched.ByteData[base+v*4+3] = 0x80
	}
	cfg2 := *cfg
	cfg2.ParseAt = uint32(371 + fr.FPQWC)
	vs, err := Emulate(&cfg2, &patched, cfg.STMagic)
	if err != nil {
		return nil, err
	}
	out := make([]TopoVert, 0, len(vs))
	for _, v := range vs {
		out = append(out, TopoVert{Index: int(v.RGBA[0]) - 1, ADC: v.ADC})
	}
	return out, nil
}

// FindMicro locates the merc microprogram in ENGINE.CGO's raw entries: the
// vu-function header {0x7E5, 0, 0x3F3, 0} followed by the code (VU code
// carries no relocations, so the raw payload bytes are the program).
func FindMicro(entries [][]byte) []byte {
	needle := []byte{0xE5, 0x07, 0, 0, 0, 0, 0, 0, 0xF3, 0x03, 0, 0, 0, 0, 0, 0}
	for _, e := range entries {
		for i := 0; i+16+0x3F30 <= len(e); i++ {
			match := true
			for k, b := range needle {
				if e[i+k] != b {
					match = false
					break
				}
			}
			if match {
				return e[i+16 : i+16+0x3F30]
			}
		}
	}
	return nil
}

// DefaultLowMem builds the constant block the chain would upload, using the
// engine's own tag templates and a reference *math-camera* (captured live;
// any well-formed camera works — the packet structure, not the projected
// positions, is what topology extraction consumes).
func DefaultLowMem() []byte {
	low := make([]byte, 140*16)
	put := func(qw, k int, v uint32) {
		binary.LittleEndian.PutUint32(low[qw*16+k*4:], v)
	}
	put(0, 1, 0x301E4000) // GIF tag template: NREG 3, regs ST/RGBAQ/XYZF2
	put(0, 2, 0x412)
	put(1, 0, 5) // adgif tag: NLOOP 5, A+D
	put(1, 1, 0x10000000)
	put(1, 2, 14)
	cam := [4][4]uint32{ // *math-camera* rows at title-logo
		{0xBED66E1D, 0, 0, 0},
		{0, 0xBE7A2B23, 0, 0},
		{0, 0, 0xC5C808F1, 0xB8151DE3},
		{0, 0, 0x4B4807A9, 0},
	}
	for r, row := range cam {
		for k, v := range row {
			put(3+r, k, v)
		}
	}
	for k, v := range []uint32{0x45000000, 0x4AFFBF9B, 0x438428EE, 0x40800000} {
		put(2, k, v) // hvdf offset
	}
	for k, v := range []uint32{0xBD3EA7E6, 0x3F800000, 0x3F800000, 0} {
		put(7, k, v) // fog row
	}
	put(132, 3, 0x3F800000) // q divisor 1.0
	for k := 0; k < 4; k++ {
		put(138, k, 0x3F800000) // ambient 1: colors pass through
	}
	return low
}

// IdentityBones returns a bone image map covering the fragment's transfers
// with identity matrices (rest/object pose).
func IdentityBones(fr *Fragment) map[byte][]byte {
	ident := make([]byte, 7*16)
	for r := 0; r < 4; r++ {
		binary.LittleEndian.PutUint32(ident[r*16+r*4:], 0x3F800000)
	}
	for r := 4; r < 7; r++ {
		binary.LittleEndian.PutUint32(ident[r*16+(r-4)*4:], 0x3F800000)
	}
	m := map[byte][]byte{}
	for _, x := range fr.Mats {
		m[x.Index] = ident
	}
	return m
}

// Session runs an effect's fragments sequentially in one VU with real
// double-buffering, so cross-fragment vertex reuse (header bytes 8/9) finds
// the previous fragment's actual output.
type Session struct {
	v     *vu.VU
	magic uint32
	top   uint16
	first bool
	kicks []uint32
	// carried maps a previous packet's raw XYZ quadword to its resolved
	// vertex index, for stitch verts whose color record was not encoded.
	carried map[[16]byte]int
}

// NewSession builds the VU with the microprogram, low-mem block and ctrl row
// installed, and runs the init prologue.
func NewSession(micro, lowMem, ctrlRow []byte, stMagic uint32) (*Session, error) {
	m := micro
	if len(m) < 16384 {
		mm := make([]byte, 16384)
		copy(mm, m)
		m = mm
	}
	data := make([]byte, 16384)
	copy(data, lowMem)
	if len(ctrlRow) >= 16 {
		copy(data[139*16:], ctrlRow[:16])
	}
	v := vu.New(m, data)
	if _, ok := v.Run(0, 4000); !ok {
		return nil, fmt.Errorf("merc session: init prologue did not halt")
	}
	// VIF double buffering: BASE 0x1BA, OFFSET 582 — TOPS alternates
	// 442, (442+582)%1024 = 0, starting at BASE (DBF=0).
	s2 := &Session{v: v, magic: stMagic, top: 442, first: true, carried: map[[16]byte]int{}}
	v.XGKick = func(qw uint32) { s2.kicks = append(s2.kicks, qw) }
	return s2, nil
}

// RunFragment feeds one fragment (vertex indices globally encoded from
// gbase into the color records) and returns the vertices of the packets the
// microprogram KICKS during the run — the GIF tag is patched by kick time,
// so NLOOP bounds the parse and no signature scanning is needed. Encoded
// RGBA words resolve identity directly; stitch-carried verts (copied
// verbatim from the previous packet) resolve by raw-XYZ-quadword match.
func (s *Session) RunFragment(fr *Fragment, gbase int) ([]TopoVert, error) {
	return s.runOnce(fr, gbase, 0)
}

func (s *Session) runOnce(fr *Fragment, gbase, bias int) ([]TopoVert, error) {
	patched := *fr
	patched.ByteData = append([]byte(nil), fr.ByteData...)
	base := (int(fr.ByteData[12]) + 1) * 4
	nv := fr.LumpQWC / 3
	for v := 0; base+v*4+3 < len(patched.ByteData) && v < nv+2; v++ {
		idx := gbase + v + 1 + bias
		patched.ByteData[base+v*4] = byte(idx)
		patched.ByteData[base+v*4+1] = byte(idx >> 8)
		patched.ByteData[base+v*4+2] = 0
		patched.ByteData[base+v*4+3] = 0x80
	}

	data := s.v.Data
	row := [4]uint32{s.magic, s.magic, 0x47800000, 0x4B010000}
	top := uint32(s.top)
	put := func(qw uint32, vals [4]uint32) {
		for k, val := range vals {
			binary.LittleEndian.PutUint32(data[(qw&1023)*16+uint32(k)*4:], val)
		}
	}
	addr := top + 140
	for r := 0; r < patched.ByteQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = uint32(patched.ByteData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	for r := 0; r < patched.LumpQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = row[k] + uint32(patched.LumpData[r*4+k])
		}
		put(addr, w)
		addr++
	}
	for r := 0; r < patched.FPQWC; r++ {
		var w [4]uint32
		for k := 0; k < 4; k++ {
			w[k] = binary.LittleEndian.Uint32(patched.FPData[r*16+k*4:])
		}
		put(addr, w)
		addr++
	}
	for _, m := range fr.Mats {
		copy(data[uint32(m.Dest)*16:], identBone())
	}

	s.v.Top = s.top
	entry := uint32(0x88)
	if s.first {
		entry = 0xA0
		s.first = false
	}
	s.kicks = nil
	if os.Getenv("MERCDBG") != "" {
		hits := map[uint32]int{}
		s.v.Trace = func(vm *vu.VU, pc uint32, raw uint64) {
			if pc >= 0x3E80 && pc <= 0x3F18 {
				hits[pc]++
			}
		}
		defer func() {
			s.v.Trace = nil
			fmt.Fprintf(os.Stderr, "[tailhits=%v]\n", hits)
		}()
	}
	if _, ok := s.v.Run(entry, 400000); !ok {
		return nil, fmt.Errorf("merc session: fragment did not halt")
	}
	if s.top == 0 {
		s.top = 442
	} else {
		s.top = 0
	}

	out := s.harvest(bias)
	return out, nil
}

// Flush kicks the pending packet after the last fragment (the microcode
// kicks a packet during the NEXT run's buffer swap; the chain ends with a
// terminator call that reaches the kick tail).
func (s *Session) Flush() ([]TopoVert, error) {
	s.kicks = nil
	if _, ok := s.v.Run(0x3E98, 40000); !ok {
		return nil, fmt.Errorf("merc session: flush did not halt")
	}
	return s.harvest(0), nil
}

// harvest resolves the kicked packets collected since the last reset.
func (s *Session) harvest(bias int) []TopoVert {
	if os.Getenv("MERCDBG") != "" {
		fmt.Fprintf(os.Stderr, "[kicks=%d %v", len(s.kicks), s.kicks)
		for _, k := range s.kicks {
			for dq := -3; dq <= 3; dq++ {
				q := (int(k) + dq) & 1023
				fmt.Fprintf(os.Stderr, " [%+d]={%08X %08X %08X %08X}", dq,
					binary.LittleEndian.Uint32(s.v.Data[q*16:]),
					binary.LittleEndian.Uint32(s.v.Data[q*16+4:]),
					binary.LittleEndian.Uint32(s.v.Data[q*16+8:]),
					binary.LittleEndian.Uint32(s.v.Data[q*16+12:]))
			}
		}
		fmt.Fprintln(os.Stderr, "]")
	}
	var out []TopoVert
	nextCarried := map[[16]byte]int{}
	for k, v := range s.carried {
		nextCarried[k] = v
	}
	for _, k := range s.kicks {
		vs, err := parseGIF(s.v.Data, k)
		if err != nil {
			continue
		}
		for _, v := range vs {
			idx := -1
			if v.RGBA[3] == 0x80 && v.RGBA[2] == 0 && v.RGBA[1] < 0x40 {
				enc := int(v.RGBA[0]) | int(v.RGBA[1])<<8
				if enc >= 1 {
					idx = enc - 1 - bias
					if idx >= 0x2000 {
						idx -= 0x2000
					}
				}
			}
			if idx < 0 {
				if ci, ok := s.carried[v.XYZQW]; ok {
					idx = ci
				}
			}
			if idx >= 0 {
				nextCarried[v.XYZQW] = idx
			}
			out = append(out, TopoVert{Index: idx, ADC: v.ADC, Addr: v.Addr})
		}
	}
	s.carried = nextCarried
	return out
}

func identBone() []byte {
	b := make([]byte, 7*16)
	for r := 0; r < 4; r++ {
		binary.LittleEndian.PutUint32(b[r*16+r*4:], 0x3F800000)
	}
	for r := 4; r < 7; r++ {
		binary.LittleEndian.PutUint32(b[r*16+(r-4)*4:], 0x3F800000)
	}
	return b
}
