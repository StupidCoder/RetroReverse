// dlwalk walks a Fast3D display list out of an RDRAM snapshot and maps the
// scene: the matrix stack, vertex loads, triangles, and texture bindings.
//
// Pilotwings 64 runs SGI's Fast3D microcode ("RSP SW Version: 2.0D, 04-01-96"
// in its ucode data, GBI version 1 opcodes on the wire). The walker implements
// the command set the game actually uses and halts loudly on anything else, so
// an unhandled command is a finding rather than a silent gap.
//
// Usage:
//
//	dlwalk -ram RAM.bin -dl 2A15C0 [-v] [-obj DIR]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"

	"retroreverse.com/tools/lib/glb"
)

type walker struct {
	ram     []byte
	seg     [16]uint32
	verbose bool

	depth int

	// current state
	mtxStack []mtx44
	proj     mtx44
	vtx      [64]vertex // Fast3D's vertex buffer is 16 entries; keep headroom
	texImg   uint32
	texFmt   uint32
	texSiz   uint32
	texScale [2]uint32
	texOn    bool
	tlut     uint32 // RDRAM address of the palette Load_TLUT last read
	tile     [8]tileDesc
	texTile  uint32            // the tile G_TEXTURE selects for drawing
	tmemSrc  map[uint32]uint32 // TMEM word address -> RDRAM source of the last load into it
	geoMode  uint32            // G_SETGEOMETRYMODE bits: lighting decides colour-vs-normal

	// per-draw grouping: one group per (modelview matrix, texture image)
	groups map[string]*group
	order  []string

	tris int
}

// tileDesc mirrors the Set_Tile / Set_Tile_Size state a draw samples through.
type tileDesc struct {
	fmt, size, line, tmem, pal     uint32
	cmS, maskS, shiftS             uint32
	cmT, maskT, shiftT             uint32
	sl, tl, sh, th                 uint32 // 10.2
	img                            uint32 // RDRAM source of the texels
	width                          uint32 // texture image width at load time
}

type mtx44 [4][4]float64

type vertex struct {
	x, y, z    int16
	flag       uint16
	s, t       int16
	r, g, b, a uint8
	// world-space position after the modelview at load time
	wx, wy, wz float64
}

type group struct {
	name   string
	texImg uint32
	tile   tileDesc
	tlut   uint32
	scale  [2]uint32
	mtx    mtx44
	lit    bool // G_LIGHTING: vertex bytes 12..15 are a normal, not a colour
	texGen bool // G_TEXTURE_GEN: s,t come from the normal (environment map)
	verts  []vertex
	faces  [][3]int
}

// F3D geometry mode bits.
const (
	geoLighting   = 0x00020000
	geoTextureGen = 0x00040000
)

func main() {
	ramFile := flag.String("ram", "", "RDRAM snapshot")
	dlAddr := flag.String("dl", "", "display list physical address (hex)")
	verbose := flag.Bool("v", false, "print every command")
	objDir := flag.String("obj", "", "write one .obj per draw group into this directory")
	glbDir := flag.String("glb", "", "write one textured .glb per draw group into this directory")
	var objects multiFlag
	flag.Var(&objects, "object", `NAME=idx,idx,... — merge the listed draw groups into NAME.glb
(written to the -glb directory). Groups are placed relative to the first
group's matrix, so an articulated model reassembles from its parts.`)
	flag.Parse()

	ram, err := os.ReadFile(*ramFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	addr, err := strconv.ParseUint(strings.TrimPrefix(*dlAddr, "0x"), 16, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad -dl")
		os.Exit(1)
	}

	w := &walker{ram: ram, verbose: *verbose, groups: map[string]*group{}, tmemSrc: map[uint32]uint32{}}
	w.mtxStack = []mtx44{identity()}
	w.walk(uint32(addr))

	fmt.Printf("\n%d triangles in %d groups\n", w.tris, len(w.groups))
	for _, k := range w.order {
		g := w.groups[k]
		flags := ""
		if g.lit {
			flags += " lit"
		}
		if g.texGen {
			flags += " texgen"
		}
		fmt.Printf("  %-40s tex=%06X %4d verts %4d tris  T=(%8.1f %8.1f %8.1f)%s\n",
			g.name, g.texImg, len(g.verts), len(g.faces),
			g.mtx[3][0], g.mtx[3][1], g.mtx[3][2], flags)
	}

	if *objDir != "" {
		os.MkdirAll(*objDir, 0o755)
		for _, k := range w.order {
			g := w.groups[k]
			writeOBJ(*objDir, g)
		}
	}
	if *glbDir != "" && len(objects) == 0 {
		w.writeGLBs(*glbDir)
	}
	for _, spec := range objects {
		eq := strings.IndexByte(spec, '=')
		if eq < 0 {
			fmt.Fprintf(os.Stderr, "bad -object %q\n", spec)
			continue
		}
		name := spec[:eq]
		var gs []*group
		for _, s := range strings.Split(spec[eq+1:], ",") {
			i, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || i < 0 || i >= len(w.order) {
				fmt.Fprintf(os.Stderr, "bad group index %q in %q\n", s, spec)
				continue
			}
			gs = append(gs, w.groups[w.order[i]])
		}
		if len(gs) > 0 {
			w.writeObject(*glbDir, name, gs)
		}
	}
}

// resolve maps a segmented or KSEG0 address to a physical RDRAM offset.
func (w *walker) resolve(a uint32) uint32 {
	seg := a >> 24 & 0xF
	return (w.seg[seg] + a&0xFFFFFF) & 0x3FFFFF
}

func (w *walker) be64(a uint32) uint64 { return binary.BigEndian.Uint64(w.ram[a:]) }

func (w *walker) cur() *mtx44 { return &w.mtxStack[len(w.mtxStack)-1] }

func (w *walker) logf(format string, args ...interface{}) {
	if w.verbose {
		fmt.Printf("%s%s\n", strings.Repeat("  ", w.depth), fmt.Sprintf(format, args...))
	}
}

func (w *walker) walk(pc uint32) {
	for {
		cmd := w.be64(pc)
		op := uint32(cmd >> 56)
		w0 := uint32(cmd >> 32)
		w1 := uint32(cmd)

		switch op {
		case 0x01: // G_MTX
			p := w0 >> 16 & 0xFF
			a := w.resolve(w1)
			m := readMtx(w.ram[a:])
			proj := p&1 != 0
			load := p&2 != 0
			push := p&4 != 0
			w.logf("G_MTX %s%s%s addr=%06X", iff(proj, "PROJ ", "MV "), iff(load, "LOAD", "MUL"), iff(push, " PUSH", ""), a)
			if proj {
				if load {
					w.proj = m
				} else {
					w.proj = mul(m, w.proj)
				}
			} else {
				if push {
					w.mtxStack = append(w.mtxStack, *w.cur())
				}
				if load {
					*w.cur() = m
				} else {
					*w.cur() = mul(m, *w.cur())
				}
			}
		case 0xBD: // G_POPMTX
			if len(w.mtxStack) > 1 {
				w.mtxStack = w.mtxStack[:len(w.mtxStack)-1]
			}
			w.logf("G_POPMTX")
		case 0x03: // G_MOVEMEM
			w.logf("G_MOVEMEM idx=%02X addr=%06X", w0>>16&0xFF, w.resolve(w1))
		case 0x04: // G_VTX: (n-1)<<20 | v0<<16 | bytes-1
			n := w0>>20&0xF + 1
			v0 := w0 >> 16 & 0xF
			a := w.resolve(w1)
			w.logf("G_VTX v0=%d n=%d addr=%06X", v0, n, a)
			for i := uint32(0); i < n; i++ {
				v := readVertex(w.ram[a+i*16:])
				m := w.cur()
				x, y, z := float64(v.x), float64(v.y), float64(v.z)
				v.wx = m[0][0]*x + m[1][0]*y + m[2][0]*z + m[3][0]
				v.wy = m[0][1]*x + m[1][1]*y + m[2][1]*z + m[3][1]
				v.wz = m[0][2]*x + m[1][2]*y + m[2][2]*z + m[3][2]
				w.vtx[v0+i] = v
			}
		case 0xBF: // G_TRI1: indices are byte offsets, 10 per vertex
			i0, i1, i2 := w1>>16&0xFF/10, w1>>8&0xFF/10, w1&0xFF/10
			w.addTri(i0, i1, i2)
		case 0xB5: // G_QUAD (two tris)
			// w1 bytes: v0,v1,v2 then v0,v2,v3 per GBI
			i0, i1, i2, i3 := w1>>24&0xFF/10, w1>>16&0xFF/10, w1>>8&0xFF/10, w1&0xFF/10
			w.addTri(i0, i1, i2)
			w.addTri(i0, i2, i3)
		case 0x06: // G_DL
			a := w.resolve(w1)
			if w0>>16&0xFF == 0 {
				w.logf("G_DL call %06X", a)
				w.depth++
				w.walk(a)
				w.depth--
			} else {
				w.logf("G_DL branch %06X", a)
				pc = a
				continue
			}
		case 0xB8: // G_ENDDL
			w.logf("G_ENDDL")
			return
		case 0xBC: // G_MOVEWORD
			idx := w1 & 0 // placeholder; real layout below
			_ = idx
			index := uint32(cmd>>32) & 0xFF
			offset := uint32(cmd>>40) & 0xFFFF
			if index == 6 { // G_MW_SEGMENT
				w.seg[offset/4] = w1 & 0xFFFFFF
				w.logf("G_MOVEWORD segment[%d] = %06X", offset/4, w1&0xFFFFFF)
			} else {
				w.logf("G_MOVEWORD idx=%d off=%X val=%08X", index, offset, w1)
			}
		case 0xBB: // G_TEXTURE
			w.texScale[0] = w1 >> 16
			w.texScale[1] = w1 & 0xFFFF
			w.texTile = w0 >> 8 & 7
			w.texOn = w0&0xFF != 0
			w.logf("G_TEXTURE on=%v tile=%d scale=%04X,%04X", w.texOn, w.texTile, w.texScale[0], w.texScale[1])
		case 0xFD: // G_SETTIMG
			w.texImg = w.resolve(w1)
			w.texFmt = w0 >> 21 & 7
			w.texSiz = w0 >> 19 & 3
			w.logf("G_SETTIMG fmt=%d size=%d addr=%06X", w.texFmt, w.texSiz, w.texImg)
		case 0xF5: // Set_Tile
			t := &w.tile[w0>>24&7]
			t.fmt, t.size = w0>>21&7, w0>>19&3
			t.line, t.tmem = w0>>9&0x1FF, w0&0x1FF
			t.pal = w1 >> 20 & 15
			t.cmT, t.maskT, t.shiftT = w1>>18&3, w1>>14&15, w1>>10&15
			t.cmS, t.maskS, t.shiftS = w1>>8&3, w1>>4&15, w1&15
			w.logf("G_SETTILE %d fmt=%d size=%d line=%d tmem=%d", w0>>24&7, t.fmt, t.size, t.line, t.tmem)
		case 0xF2: // Set_Tile_Size
			t := &w.tile[w1>>24&7]
			t.sl, t.tl = w0>>12&0xFFF, w0&0xFFF
			t.sh, t.th = w1>>12&0xFFF, w1&0xFFF
			w.logf("G_SETTILESIZE %d %d..%d x %d..%d", w1>>24&7, t.sl>>2, t.sh>>2, t.tl>>2, t.th>>2)
		case 0xF3: // Load_Block: loads land in TMEM; draws sample through another
			// tile pointed at the same TMEM address, so the source is remembered
			// by TMEM word, not by tile index.
			ti := w1 >> 24 & 7
			w.tmemSrc[w.tile[ti].tmem] = w.texImg
			w.logf("G_LOADBLOCK tile=%d sh=%d dxt=%d src=%06X tmem=%d", ti, w1>>12&0xFFF, w1&0xFFF, w.texImg, w.tile[ti].tmem)
		case 0xF4: // Load_Tile
			ti := w1 >> 24 & 7
			w.tmemSrc[w.tile[ti].tmem] = w.texImg
			w.logf("G_LOADTILE tile=%d src=%06X", ti, w.texImg)
		case 0xF0: // Load_TLUT
			w.tlut = w.texImg
			w.logf("G_LOADTLUT src=%06X", w.texImg)
		case 0xB7: // G_SETGEOMETRYMODE
			w.geoMode |= w1
			w.logf("G_SETGEOMETRYMODE %08X -> %08X", w1, w.geoMode)
		case 0xB6: // G_CLEARGEOMETRYMODE
			w.geoMode &^= w1
			w.logf("G_CLEARGEOMETRYMODE %08X -> %08X", w1, w.geoMode)
		case 0xB2, 0xB3, // G_RDPHALF_CONT / G_RDPHALF_2
			0xB4, // G_PERSPNORMALIZE
			0xB9, 0xBA, 0xBE, 0xE4, 0xE6, 0xE7, 0xE8, 0xE9, 0xED, 0xEE, 0xEF,
			0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB, 0xFC, 0xFE, 0xFF:
			// geometry mode, othermode, cull, and RDP passthroughs: not needed
			// for geometry extraction, logged for the record.
			w.logf("op %02X %08X %08X", op, w0, w1)
		case 0x00: // G_SPNOOP
		default:
			fmt.Printf("UNHANDLED op %02X at %06X: %08X %08X\n", op, pc, w0, w1)
			return
		}
		pc += 8
	}
}

func (w *walker) addTri(i0, i1, i2 uint32) {
	w.tris++
	// Group key: the modelview matrix content + the texture the draw samples,
	// through the tile G_TEXTURE selected and the TMEM address it reads.
	m := w.cur()
	t := w.tile[w.texTile]
	t.img = w.tmemSrc[t.tmem]
	key := fmt.Sprintf("%x-%06X-%d", mtxHash(m), t.img, t.fmt)
	g, ok := w.groups[key]
	if !ok {
		g = &group{
			name: fmt.Sprintf("group-%03d-tex%06X", len(w.groups), t.img),
			texImg: t.img, tile: t, tlut: w.tlut, scale: w.texScale, mtx: *m,
			lit: w.geoMode&geoLighting != 0, texGen: w.geoMode&geoTextureGen != 0,
		}
		w.groups[key] = g
		w.order = append(w.order, key)
	}
	base := len(g.verts)
	for _, i := range []uint32{i0, i1, i2} {
		g.verts = append(g.verts, w.vtx[i])
	}
	g.faces = append(g.faces, [3]int{base, base + 1, base + 2})
}

func mtxHash(m *mtx44) uint64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			h ^= uint64(int64(m[i][j] * 65536))
			h *= 1099511628211
		}
	}
	return h
}

func readVertex(d []byte) vertex {
	return vertex{
		x: int16(binary.BigEndian.Uint16(d[0:])), y: int16(binary.BigEndian.Uint16(d[2:])),
		z:    int16(binary.BigEndian.Uint16(d[4:])),
		flag: binary.BigEndian.Uint16(d[6:]),
		s:    int16(binary.BigEndian.Uint16(d[8:])), t: int16(binary.BigEndian.Uint16(d[10:])),
		r: d[12], g: d[13], b: d[14], a: d[15],
	}
}

// readMtx reads libultra's 4x4 s15.16 matrix: 32 bytes of integer parts then
// 32 bytes of fraction parts, row-major.
func readMtx(d []byte) mtx44 {
	var m mtx44
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			hi := int16(binary.BigEndian.Uint16(d[(i*4+j)*2:]))
			lo := binary.BigEndian.Uint16(d[32+(i*4+j)*2:])
			m[i][j] = float64(hi) + float64(lo)/65536
		}
	}
	return m
}

func mul(a, b mtx44) mtx44 {
	var r mtx44
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return r
}

func identity() mtx44 {
	var m mtx44
	for i := 0; i < 4; i++ {
		m[i][i] = 1
	}
	return m
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func iff(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

func writeOBJ(dir string, g *group) {
	var b strings.Builder
	fmt.Fprintf(&b, "o %s\n", g.name)
	for _, v := range g.verts {
		fmt.Fprintf(&b, "v %f %f %f %f %f %f\n", v.wx, v.wy, v.wz,
			float64(v.r)/255, float64(v.g)/255, float64(v.b)/255)
	}
	for _, f := range g.faces {
		fmt.Fprintf(&b, "f %d %d %d\n", f[0]+1, f[1]+1, f[2]+1)
	}
	os.WriteFile(fmt.Sprintf("%s/%s.obj", dir, g.name), []byte(b.String()), 0o644)
}

// --- texture decode -----------------------------------------------------

// decodeTexture reads a group's texture straight out of the RDRAM snapshot,
// using the same texel formats as the oracle's sampler (rdp_texture.go). The
// data in RDRAM is laid out for a dxt=0 Load_Block — pre-swizzled, odd rows
// with their 32-bit word halves exchanged — so the reader undoes the same swap
// the sampler does.
func (w *walker) decodeTexture(g *group) image.Image {
	t := g.tile
	if t.sh <= t.sl && t.th <= t.tl {
		return nil
	}
	tw := int(t.sh>>2-t.sl>>2) + 1
	th := int(t.th>>2-t.tl>>2) + 1
	if tw <= 0 || th <= 0 || tw > 1024 || th > 1024 || t.img == 0 {
		return nil
	}
	img := image.NewRGBA(image.Rect(0, 0, tw, th))
	rowBytes := int(t.line) * 8
	texel := func(off int) byte {
		if off < 0 || off >= len(w.ram) {
			return 0
		}
		return w.ram[off]
	}
	for y := 0; y < th; y++ {
		row := int(t.img) + y*rowBytes
		for x := 0; x < tw; x++ {
			var r, gg, b, a uint32
			swz := func(off int) int {
				if y&1 != 0 {
					return off ^ 4
				}
				return off
			}
			switch {
			case t.fmt == 0 && t.size == 2: // RGBA16
				v := uint16(texel(swz(row+x*2)))<<8 | uint16(texel(swz(row+x*2)+1))
				r, gg, b = uint32(v>>11&31)<<3, uint32(v>>6&31)<<3, uint32(v>>1&31)<<3
				a = uint32(v&1) * 255
			case t.fmt == 4 && t.size == 0: // I4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				i := n * 17
				r, gg, b, a = i, i, i, i
			case t.fmt == 4 && t.size == 1: // I8
				i := uint32(texel(swz(row + x)))
				r, gg, b, a = i, i, i, i
			case t.fmt == 3 && t.size == 0: // IA4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				i := (n >> 1) * 36
				r, gg, b = i, i, i
				a = uint32(n&1) * 255
			case t.fmt == 3 && t.size == 1: // IA8
				v := texel(swz(row + x))
				i := uint32(v>>4) * 17
				r, gg, b, a = i, i, i, uint32(v&15)*17
			case t.fmt == 3 && t.size == 2: // IA16
				i := uint32(texel(swz(row + x*2)))
				a = uint32(texel(swz(row+x*2) + 1))
				r, gg, b = i, i, i
			case t.fmt == 2 && t.size == 1: // CI8
				idx := uint32(texel(swz(row + x)))
				v := uint16(texel(int(w.tlut)+int(idx)*2))<<8 | uint16(texel(int(w.tlut)+int(idx)*2+1))
				r, gg, b = uint32(v>>11&31)<<3, uint32(v>>6&31)<<3, uint32(v>>1&31)<<3
				a = uint32(v&1) * 255
			case t.fmt == 2 && t.size == 0: // CI4
				v := texel(swz(row + x/2))
				n := uint32(v >> 4)
				if x&1 == 1 {
					n = uint32(v & 15)
				}
				idx := t.pal<<4 | n
				pv := uint16(texel(int(w.tlut)+int(idx)*2))<<8 | uint16(texel(int(w.tlut)+int(idx)*2+1))
				r, gg, b = uint32(pv>>11&31)<<3, uint32(pv>>6&31)<<3, uint32(pv>>1&31)<<3
				a = uint32(pv&1) * 255
			default:
				return nil
			}
			img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(gg), uint8(b), uint8(a)})
		}
	}
	return img
}

// --- GLB export ---------------------------------------------------------

// writeObject merges draw groups into one GLB. Vertices are model-space; a
// group whose matrix differs from the first group's is baked through the
// relative transform, so an articulated model (a gyrocopter's rotor over its
// body) reassembles the way the frame posed it.
func (w *walker) writeObject(dir, name string, gs []*group) {
	os.MkdirAll(dir, 0o755)
	root := invertAffine(gs[0].mtx)

	var pos [][3]float32
	var uvs [][2]float32
	var cols [][4]uint8
	var texGroups []glb.TexturedGroup
	for _, g := range gs {
		rel := mul(g.mtx, root)
		img := w.decodeTexture(g)
		tw := float64(g.tile.sh>>2-g.tile.sl>>2) + 1
		tth := float64(g.tile.th>>2-g.tile.tl>>2) + 1
		sScale := float64(g.scale[0]) / 65536
		tScale := float64(g.scale[1]) / 65536

		var tris [][3]uint32
		for _, f := range g.faces {
			base := uint32(len(pos))
			for _, vi := range f {
				v := g.verts[vi]
				x, y, z := float64(v.x), float64(v.y), float64(v.z)
				pos = append(pos, [3]float32{
					float32(rel[0][0]*x + rel[1][0]*y + rel[2][0]*z + rel[3][0]),
					float32(rel[0][1]*x + rel[1][1]*y + rel[2][1]*z + rel[3][1]),
					float32(rel[0][2]*x + rel[1][2]*y + rel[2][2]*z + rel[3][2]),
				})
				uvs = append(uvs, [2]float32{
					float32(float64(v.s) / 32 * sScale / tw),
					float32(float64(v.t) / 32 * tScale / tth),
				})
				cols = append(cols, [4]uint8{v.r, v.g, v.b, v.a})
			}
			tris = append(tris, [3]uint32{base, base + 1, base + 2})
		}
		if img == nil {
			img = image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.(*image.RGBA).SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})
		}
		texGroups = append(texGroups, glb.TexturedGroup{
			Tris: tris, Image: img,
			WrapS: wrapMode(g.tile.cmS, g.tile.maskS),
			WrapT: wrapMode(g.tile.cmT, g.tile.maskT),
		})
	}

	path := fmt.Sprintf("%s/%s.glb", dir, name)
	if err := glb.WriteTexturedColored(path, pos, uvs, cols, texGroups, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return
	}
	fmt.Printf("wrote %s: %d groups, %d verts\n", path, len(gs), len(pos))
}

// wrapMode maps a tile's cm/mask to the glTF sampler wrap enum. The cm bits
// are libultra's G_TX constants: 1 mirrors, 2 clamps; no mask forces a clamp.
func wrapMode(cm, mask uint32) int {
	switch {
	case cm&2 != 0 || mask == 0:
		return 33071 // CLAMP_TO_EDGE
	case cm&1 != 0:
		return 33648 // MIRRORED_REPEAT
	default:
		return 10497 // REPEAT
	}
}

// invertAffine inverts a rigid-ish modelview (rotation+scale+translation).
func invertAffine(m mtx44) mtx44 {
	// Invert the 3x3 by adjugate, then the translation.
	a := [3][3]float64{
		{m[0][0], m[0][1], m[0][2]},
		{m[1][0], m[1][1], m[1][2]},
		{m[2][0], m[2][1], m[2][2]},
	}
	det := a[0][0]*(a[1][1]*a[2][2]-a[1][2]*a[2][1]) -
		a[0][1]*(a[1][0]*a[2][2]-a[1][2]*a[2][0]) +
		a[0][2]*(a[1][0]*a[2][1]-a[1][1]*a[2][0])
	if det == 0 {
		return identity()
	}
	inv := func(i, j int) float64 {
		i1, i2 := (i+1)%3, (i+2)%3
		j1, j2 := (j+1)%3, (j+2)%3
		return (a[j1][i1]*a[j2][i2] - a[j1][i2]*a[j2][i1]) / det
	}
	var r mtx44
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			r[i][j] = inv(i, j)
		}
	}
	// translation' = -T * R'
	for j := 0; j < 3; j++ {
		r[3][j] = -(m[3][0]*r[0][j] + m[3][1]*r[1][j] + m[3][2]*r[2][j])
	}
	r[3][3] = 1
	return r
}

// writeGLBs writes one textured GLB per draw group, in model space (the raw
// vertex coordinates the game keeps in RDRAM), with UVs derived from the
// vertex s/t, the G_TEXTURE scale, and the tile extent.
func (w *walker) writeGLBs(dir string) {
	os.MkdirAll(dir, 0o755)
	for _, k := range w.order {
		g := w.groups[k]
		if len(g.faces) == 0 {
			continue
		}
		img := w.decodeTexture(g)

		tw := float64(g.tile.sh>>2-g.tile.sl>>2) + 1
		tth := float64(g.tile.th>>2-g.tile.tl>>2) + 1
		sScale := float64(g.scale[0]) / 65536
		tScale := float64(g.scale[1]) / 65536

		var pos [][3]float32
		var uvs [][2]float32
		var tris [][3]uint32
		for _, f := range g.faces {
			base := uint32(len(pos))
			for _, vi := range f {
				v := g.verts[vi]
				pos = append(pos, [3]float32{float32(v.x), float32(v.y), float32(v.z)})
				// s,t are 10.5 texel coordinates before the texture scale.
				u := float64(v.s) / 32 * sScale / tw
				vv := float64(v.t) / 32 * tScale / tth
				uvs = append(uvs, [2]float32{float32(u), float32(vv)})
			}
			tris = append(tris, [3]uint32{base, base + 1, base + 2})
		}

		path := fmt.Sprintf("%s/%s.glb", dir, g.name)
		var err error
		if img != nil {
			err = glb.WriteTextured(path, pos, uvs, []glb.TexturedGroup{{Tris: tris, Image: img}}, nil)
		} else {
			err = glb.WriteTrianglesMat(path, pos, []glb.TriGroup{{Tris: tris}})
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		}
	}
}
