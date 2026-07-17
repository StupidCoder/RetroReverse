package xbox

// nv2a_vsh.go is the LLE NV2A vertex-program ("transform program") interpreter — the
// Kelvin analogue of the PICA200 shader interpreter (n3ds/gpu_shader.go). The program
// is the game's own data, uploaded four dwords at a time through the Kelvin
// SET_TRANSFORM_PROGRAM window (methods 0x0B00-0x0B7C) into 136 instruction slots;
// running it instruction by instruction is the clean-room equivalent of LLE'ing a
// microcode — no assumption about what transform the game "probably" does.
//
// The machine: 16 input registers v0-15, temporaries R0-R11 plus R12 (which IS the
// position output — writing oPos writes R12 and the D3D-appended screen-space epilogue
// reads it back), 192 vec4 constants c0-191, 13 output registers (o0=HPOS, o3=D0
// diffuse, o4=D1 specular, o5=FOG, o6=PTS, o7/o8=back colors, o9-12=TEX0-3), one
// address register a0.x.
//
// The 128-bit instruction word (4 dwords; dword 0 is unused padding) dual-issues a
// MAC op and an ILU op. The field map below is the NV2A hardware's (envytools-class
// platform documentation), and it has been verified field-by-field against OutRun's
// own uploaded program: the loading screen's 12-instruction transform decodes as a
// classic pre-transformed-vertex path (MOV R1,v0; dual MOV o3,v3 + RCP R1.w;
// MOV o5,v4.w fog-from-specular-alpha; MUL R2,R1,c0 + MOV o4,v4 ...), each field
// landing exactly where this map says it does.
//
//	dword 1: [27:25] ILU op   [24:21] MAC op   [20:13] const index  [12:9] input reg
//	         [8] A negate     [7:6] A swz x  [5:4] y  [3:2] z  [1:0] w
//	dword 2: [31:28] A temp reg  [27:26] A mux  [25] B neg  [24:17] B swz
//	         [16:13] B temp reg  [12:11] B mux  [10] C neg  [9:2] C swz
//	         [1:0] C temp reg (high bits)
//	dword 3: [31:30] C temp reg (low bits)  [29:28] C mux
//	         [27:24] MAC dest mask (bit3=x)  [23:20] MAC temp dest reg
//	         [19:16] ILU dest mask (ILU temps always write R1)
//	         [15:12] output mask  [11] output bank (1=o register, 0=constant)
//	         [10:3] output address  [2] output mux (0=MAC result, 1=ILU)
//	         [1] relative const addressing (+a0.x)  [0] final instruction
//
// mux values: 1 = R (temp), 2 = V (input), 3 = C (constant). An idle output port
// encodes mask 0, bank 1, address 0xFF.
//
// Unknown ops halt loudly — the honest bring-up posture; ops graduate as the game's
// programs demand them.

import (
	"fmt"
	"math"
	"os"
	"strings"
)

var nvVSTrace = os.Getenv("RR_NV_VS") != ""

const (
	kelvinProgData  = 0x0B00 // ..0x0B7C: transform-program upload window (4 dwords = 1 instruction)
	kelvinConstData = 0x0B80 // ..0x0BFC: transform-constant upload window (4 dwords = 1 vec4)

	kelvinTransformExecMode = 0x1E94 // bits 1:0 — 0 fixed-function, 2 program
	kelvinCxtWriteEnable    = 0x1E98 // program may write constant memory
	kelvinProgLoad          = 0x1E9C // instruction slot the next program upload lands in
	kelvinProgStart         = 0x1EA0 // instruction slot execution starts at
	kelvinConstLoad         = 0x1EA4 // vec4 slot the next constant upload lands in

	vshProgSlots  = 136
	vshConstSlots = 192
)

// MAC opcodes.
const (
	macNOP = iota
	macMOV
	macMUL
	macADD // reads inputs A and C
	macMAD
	macDP3
	macDPH
	macDP4
	macDST
	macMIN
	macMAX
	macSLT
	macSGE
	macARL
)

// ILU opcodes (the scalar unit; its input is always C).
const (
	iluNOP = iota
	iluMOV
	iluRCP
	iluRCC
	iluRSQ
	iluEXP
	iluLOG
	iluLIT
)

var macNames = [...]string{"NOP", "MOV", "MUL", "ADD", "MAD", "DP3", "DPH", "DP4", "DST", "MIN", "MAX", "SLT", "SGE", "ARL"}
var iluNames = [...]string{"NOP", "MOV", "RCP", "RCC", "RSQ", "EXP", "LOG", "LIT"}

// vshSrc is one decoded source operand.
type vshSrc struct {
	mux int // 1=R, 2=V, 3=C
	reg int // temp register when mux==1
	neg bool
	swz [4]int
}

// vshInst is one decoded instruction.
type vshInst struct {
	mac, ilu   int
	constIdx   int
	inputReg   int
	a, b, c    vshSrc
	macMask    uint32 // bit3=x .. bit0=w
	macDst     int    // temp register the MAC result writes
	iluMask    uint32 // ILU temp writes always land in R1
	outMask    uint32
	outIsO     bool // true: o register; false: constant memory
	outAddr    int
	outFromILU bool
	relConst   bool // const index is relative to a0.x
	final      bool
}

func vshField(w *[4]uint32, dword, lo, n int) uint32 {
	return w[dword] >> lo & (1<<uint(n) - 1)
}

// vshDecode decodes the 4-dword instruction at program slot i.
func (g *pgraph) vshDecode(i int) vshInst {
	w := &g.Prog[i]
	swz := func(dword, lo int) [4]int {
		v := vshField(w, dword, lo, 8)
		return [4]int{int(v >> 6 & 3), int(v >> 4 & 3), int(v >> 2 & 3), int(v & 3)}
	}
	return vshInst{
		ilu:      int(vshField(w, 1, 25, 3)),
		mac:      int(vshField(w, 1, 21, 4)),
		constIdx: int(vshField(w, 1, 13, 8)),
		inputReg: int(vshField(w, 1, 9, 4)),
		a: vshSrc{
			neg: vshField(w, 1, 8, 1) != 0,
			swz: swz(1, 0),
			reg: int(vshField(w, 2, 28, 4)),
			mux: int(vshField(w, 2, 26, 2)),
		},
		b: vshSrc{
			neg: vshField(w, 2, 25, 1) != 0,
			swz: swz(2, 17),
			reg: int(vshField(w, 2, 13, 4)),
			mux: int(vshField(w, 2, 11, 2)),
		},
		c: vshSrc{
			neg: vshField(w, 2, 10, 1) != 0,
			swz: swz(2, 2),
			reg: int(vshField(w, 2, 0, 2)<<2 | vshField(w, 3, 30, 2)),
			mux: int(vshField(w, 3, 28, 2)),
		},
		macMask:    vshField(w, 3, 24, 4),
		macDst:     int(vshField(w, 3, 20, 4)),
		iluMask:    vshField(w, 3, 16, 4),
		outMask:    vshField(w, 3, 12, 4),
		outIsO:     vshField(w, 3, 11, 1) != 0,
		outAddr:    int(vshField(w, 3, 3, 8)),
		outFromILU: vshField(w, 3, 2, 1) != 0,
		relConst:   vshField(w, 3, 1, 1) != 0,
		final:      vshField(w, 3, 0, 1) != 0,
	}
}

// vshState is one vertex's execution state.
type vshState struct {
	g *pgraph
	v *[16][4]float32 // inputs
	r [13][4]float32  // temporaries; r[12] is oPos
	o [13][4]float32  // outputs (o[0] is aliased to r[12] on read-out)
	a int32           // a0.x
}

// src fetches a source operand: pick the file, apply relative addressing (constants),
// swizzle, negate.
func (s *vshState) src(inst *vshInst, o *vshSrc) [4]float32 {
	var base [4]float32
	switch o.mux {
	case 1:
		base = s.r[o.reg&0xF%13] // temp regs are R0-R12
	case 2:
		base = s.v[inst.inputReg]
	case 3:
		c := inst.constIdx
		if inst.relConst {
			c += int(s.a)
		}
		if c >= 0 && c < vshConstSlots {
			base = f32vec(&s.g.Const[c])
		}
	default:
		// mux 0 is "no operand" — reads as zero, which an op that uses the slot
		// never encodes (verified against the live stream: unused slots carry a
		// stale-but-well-formed mux).
	}
	out := [4]float32{base[o.swz[0]], base[o.swz[1]], base[o.swz[2]], base[o.swz[3]]}
	if o.neg {
		out[0], out[1], out[2], out[3] = -out[0], -out[1], -out[2], -out[3]
	}
	return out
}

func f32vec(c *[4]uint32) [4]float32 {
	return [4]float32{
		math.Float32frombits(c[0]), math.Float32frombits(c[1]),
		math.Float32frombits(c[2]), math.Float32frombits(c[3]),
	}
}

func maskWrite(dst *[4]float32, val [4]float32, mask uint32) {
	if mask&8 != 0 {
		dst[0] = val[0]
	}
	if mask&4 != 0 {
		dst[1] = val[1]
	}
	if mask&2 != 0 {
		dst[2] = val[2]
	}
	if mask&1 != 0 {
		dst[3] = val[3]
	}
}

// vshRun executes the uploaded program for one vertex. v is the input register file;
// the outputs land in out (o0/HPOS is the R12 alias). Returns false after halting the
// machine on anything unimplemented.
func (g *pgraph) vshRun(v *[16][4]float32, out *[13][4]float32) bool {
	s := vshState{g: g, v: v}
	pc := int(g.Regs[kelvinProgStart>>2]) % vshProgSlots
	for steps := 0; ; steps++ {
		if steps >= vshProgSlots {
			g.m.CPU.Halt("nv2a vsh: no FINAL instruction within %d slots from start %d", vshProgSlots, g.Regs[kelvinProgStart>>2])
			return false
		}
		inst := g.vshDecode(pc)
		pc++

		var macRes [4]float32
		if inst.mac != macNOP {
			a := s.src(&inst, &inst.a)
			switch inst.mac {
			case macMOV:
				macRes = a
			case macMUL:
				b := s.src(&inst, &inst.b)
				for i := range macRes {
					macRes[i] = a[i] * b[i]
				}
			case macADD:
				c := s.src(&inst, &inst.c)
				for i := range macRes {
					macRes[i] = a[i] + c[i]
				}
			case macMAD:
				b, c := s.src(&inst, &inst.b), s.src(&inst, &inst.c)
				for i := range macRes {
					macRes[i] = a[i]*b[i] + c[i]
				}
			case macDP3:
				b := s.src(&inst, &inst.b)
				dp := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
				macRes = [4]float32{dp, dp, dp, dp}
			case macDPH:
				b := s.src(&inst, &inst.b)
				dp := a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + b[3]
				macRes = [4]float32{dp, dp, dp, dp}
			case macDP4:
				b := s.src(&inst, &inst.b)
				dp := a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3]
				macRes = [4]float32{dp, dp, dp, dp}
			case macDST:
				b := s.src(&inst, &inst.b)
				macRes = [4]float32{1, a[1] * b[1], a[2], b[3]}
			case macMIN:
				b := s.src(&inst, &inst.b)
				for i := range macRes {
					macRes[i] = minf32(a[i], b[i])
				}
			case macMAX:
				b := s.src(&inst, &inst.b)
				for i := range macRes {
					macRes[i] = maxf32(a[i], b[i])
				}
			case macSLT:
				b := s.src(&inst, &inst.b)
				for i := range macRes {
					if a[i] < b[i] {
						macRes[i] = 1
					}
				}
			case macSGE:
				b := s.src(&inst, &inst.b)
				for i := range macRes {
					if a[i] >= b[i] {
						macRes[i] = 1
					}
				}
			case macARL:
				s.a = int32(floorf32(a[0]))
			default:
				g.m.CPU.Halt("nv2a vsh: MAC op %d unimplemented", inst.mac)
				return false
			}
			if inst.mac != macARL && inst.macMask != 0 {
				maskWrite(&s.r[inst.macDst%13], macRes, inst.macMask)
			}
		}

		var iluRes [4]float32
		if inst.ilu != iluNOP {
			c := s.src(&inst, &inst.c)
			switch inst.ilu {
			case iluMOV:
				iluRes = c
			case iluRCP:
				r := rcpf32(c[0])
				iluRes = [4]float32{r, r, r, r}
			case iluRCC:
				r := rccf32(c[0])
				iluRes = [4]float32{r, r, r, r}
			case iluRSQ:
				r := rsqf32(c[0])
				iluRes = [4]float32{r, r, r, r}
			case iluEXP:
				e := floorf32(c[0])
				iluRes = [4]float32{exp2f32(e), c[0] - e, exp2f32(c[0]), 1}
			case iluLOG:
				iluRes = log2partial(c[0])
			case iluLIT:
				iluRes = litf32(c)
			default:
				g.m.CPU.Halt("nv2a vsh: ILU op %d unimplemented", inst.ilu)
				return false
			}
			// The ILU's temp write port is hard-wired to R1 (the MAC owns the
			// addressable temp destination).
			if inst.iluMask != 0 {
				maskWrite(&s.r[1], iluRes, inst.iluMask)
			}
		}

		if inst.outMask != 0 {
			val := macRes
			if inst.outFromILU {
				val = iluRes
			}
			if inst.outIsO {
				oreg := inst.outAddr & 0xF
				if oreg < 13 {
					maskWrite(&s.o[oreg], val, inst.outMask)
					if oreg == 0 {
						maskWrite(&s.r[12], val, inst.outMask) // oPos IS R12
					}
				}
			} else if g.Regs[kelvinCxtWriteEnable>>2] != 0 && inst.outAddr < vshConstSlots {
				// A program with constant-write enable may write constant memory.
				var bits [4]uint32
				for i, f := range val {
					bits[i] = math.Float32bits(f)
				}
				maskWrite32(&g.Const[inst.outAddr], bits, inst.outMask)
			}
		}

		if inst.final {
			break
		}
	}
	s.o[0] = s.r[12] // the position output is the R12 alias
	*out = s.o
	return true
}

func maskWrite32(dst *[4]uint32, val [4]uint32, mask uint32) {
	if mask&8 != 0 {
		dst[0] = val[0]
	}
	if mask&4 != 0 {
		dst[1] = val[1]
	}
	if mask&2 != 0 {
		dst[2] = val[2]
	}
	if mask&1 != 0 {
		dst[3] = val[3]
	}
}

// --- program/constant upload (dispatched from kelvinMethod) ---

// progData receives one dword of a SET_TRANSFORM_PROGRAM upload. Four dwords complete
// an instruction at the load slot, which then advances — the 0x0B00-0x0B7C window is a
// FIFO; the offset within it does not address anything.
func (g *pgraph) progData(arg uint32) {
	g.progBuf[g.progBufN] = arg
	g.progBufN++
	if g.progBufN < 4 {
		return
	}
	g.progBufN = 0
	slot := g.ProgLoad % vshProgSlots
	g.Prog[slot] = g.progBuf
	g.ProgLoad = slot + 1
	if nvVSTrace {
		fmt.Printf("VSH prog[%3d] = %08X %08X %08X %08X  %s\n",
			slot, g.progBuf[0], g.progBuf[1], g.progBuf[2], g.progBuf[3], g.vshDisasm(int(slot)))
	}
}

// constData receives one dword of a SET_TRANSFORM_CONSTANT upload (same FIFO shape;
// four dwords complete a vec4 at the constant load slot).
func (g *pgraph) constData(arg uint32) {
	g.constBuf[g.constBufN] = arg
	g.constBufN++
	if g.constBufN < 4 {
		return
	}
	g.constBufN = 0
	slot := g.ConstLoad % vshConstSlots
	g.Const[slot] = g.constBuf
	g.ConstLoad = slot + 1
	if nvVSTrace {
		v := f32vec(&g.Const[slot])
		fmt.Printf("VSH const[%3d] = (%g, %g, %g, %g)\n", slot, v[0], v[1], v[2], v[3])
	}
	if nvVPTrace && slot >= 56 && slot <= 60 {
		v := f32vec(&g.Const[slot])
		fmt.Printf("VP const-load c%d = (%g, %g, %g, %g) draws=%d\n", slot, v[0], v[1], v[2], v[3], g.Draws)
	}
}

// vshDisasm renders one instruction slot for the trace — enough to read a program's
// intent off a dump, not a full assembler round-trip.
func (g *pgraph) vshDisasm(i int) string {
	in := g.vshDecode(i)
	srcStr := func(o *vshSrc) string {
		var name string
		switch o.mux {
		case 1:
			name = fmt.Sprintf("R%d", o.reg%13)
		case 2:
			name = fmt.Sprintf("v%d", in.inputReg)
		case 3:
			if in.relConst {
				name = fmt.Sprintf("c[a0+%d]", in.constIdx)
			} else {
				name = fmt.Sprintf("c%d", in.constIdx)
			}
		default:
			name = "?"
		}
		sw := ""
		if o.swz != [4]int{0, 1, 2, 3} {
			comp := "xyzw"
			sw = "." + string([]byte{comp[o.swz[0]], comp[o.swz[1]], comp[o.swz[2]], comp[o.swz[3]]})
		}
		if o.neg {
			return "-" + name + sw
		}
		return name + sw
	}
	maskStr := func(m uint32) string {
		if m == 0xF {
			return ""
		}
		s := "."
		for b, c := range []byte("xyzw") {
			if m&(8>>uint(b)) != 0 {
				s += string(c)
			}
		}
		return s
	}
	var parts []string
	if in.mac != macNOP {
		p := macNames[in.mac]
		if in.mac == macARL {
			p += " a0.x, " + srcStr(&in.a)
		} else {
			dsts := []string{}
			if in.macMask != 0 {
				dsts = append(dsts, fmt.Sprintf("R%d%s", in.macDst%13, maskStr(in.macMask)))
			}
			if in.outMask != 0 && !in.outFromILU {
				dsts = append(dsts, outName(in.outIsO, in.outAddr)+maskStr(in.outMask))
			}
			if len(dsts) == 0 {
				dsts = append(dsts, "_")
			}
			p += " " + strings.Join(dsts, "/") + ", " + srcStr(&in.a)
			switch in.mac {
			case macMUL, macDP3, macDPH, macDP4, macDST, macMIN, macMAX, macSLT, macSGE:
				p += ", " + srcStr(&in.b)
			case macADD:
				p += ", " + srcStr(&in.c)
			case macMAD:
				p += ", " + srcStr(&in.b) + ", " + srcStr(&in.c)
			}
		}
		parts = append(parts, p)
	}
	if in.ilu != iluNOP {
		dsts := []string{}
		if in.iluMask != 0 {
			dsts = append(dsts, "R1"+maskStr(in.iluMask))
		}
		if in.outMask != 0 && in.outFromILU {
			dsts = append(dsts, outName(in.outIsO, in.outAddr)+maskStr(in.outMask))
		}
		if len(dsts) == 0 {
			dsts = append(dsts, "_")
		}
		parts = append(parts, iluNames[in.ilu]+" "+strings.Join(dsts, "/")+", "+srcStr(&in.c))
	}
	out := strings.Join(parts, " ; ")
	if out == "" {
		out = "NOP"
	}
	if in.final {
		out += " [FINAL]"
	}
	return out
}

func outName(isO bool, addr int) string {
	if !isO {
		return fmt.Sprintf("c%d", addr)
	}
	names := [...]string{"oPos", "o1", "o2", "oD0", "oD1", "oFog", "oPts", "oB0", "oB1", "oT0", "oT1", "oT2", "oT3"}
	if a := addr & 0xF; a < len(names) {
		return names[a]
	}
	return fmt.Sprintf("o%d", addr&0xF)
}

// --- float helpers (float32 semantics, matching the shader's own precision) ---

func minf32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func maxf32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func floorf32(f float32) float32 { return float32(math.Floor(float64(f))) }
func exp2f32(f float32) float32  { return float32(math.Exp2(float64(f))) }

func rcpf32(x float32) float32 {
	if x == 0 {
		return float32(math.Inf(1))
	}
	return 1 / x
}

// rccf32 is the clamped reciprocal: the result's magnitude is clamped into
// [5.42101e-20, 1.884467e19], preserving sign — the hardware's guard against a
// divide-by-w blowing up downstream arithmetic.
func rccf32(x float32) float32 {
	const lo, hi = 5.42101e-20, 1.884467e19
	y := float64(1) / float64(x)
	neg := math.Signbit(y)
	a := math.Abs(y)
	if a > hi || math.IsInf(a, 0) || math.IsNaN(a) {
		a = hi
	} else if a < lo {
		a = lo
	}
	if neg {
		return float32(-a)
	}
	return float32(a)
}

func rsqf32(x float32) float32 {
	a := math.Abs(float64(x))
	if a == 0 {
		return float32(math.Inf(1))
	}
	return float32(1 / math.Sqrt(a))
}

// log2partial is the LOG op's split result: exponent, mantissa, full log2.
func log2partial(x float32) [4]float32 {
	a := math.Abs(float64(x))
	if a == 0 {
		return [4]float32{float32(math.Inf(-1)), 1, float32(math.Inf(-1)), 1}
	}
	fr, exp := math.Frexp(a) // a = fr * 2^exp, fr in [0.5, 1)
	return [4]float32{float32(exp - 1), float32(fr * 2), float32(math.Log2(a)), 1}
}

// litf32 is the lighting coefficient op: x=1, y=max(src.x,0), z = src.x>0 ?
// max(src.y,0)^w : 0, w=1, with w clamped to [-128, 128].
func litf32(c [4]float32) [4]float32 {
	out := [4]float32{1, 0, 0, 1}
	p := float64(c[3])
	if p < -128 {
		p = -128
	} else if p > 128 {
		p = 128
	}
	if c[0] > 0 {
		out[1] = c[0]
		if c[1] > 0 {
			out[2] = float32(math.Pow(float64(c[1]), p))
		} else if p == 0 {
			out[2] = 1
		}
	}
	return out
}
