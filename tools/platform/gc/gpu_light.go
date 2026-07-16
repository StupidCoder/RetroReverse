package gc

// gpu_light.go is the XF's colour channel — the fixed-function lighting stage that decides
// the "rasterised colour" the TEV combines with textures. It sits between the vertex fetch
// and the rasteriser: for each vertex the channel picks a material colour (from the vertex or
// a register), and either passes it through or multiplies it by an illumination sum built
// from an ambient term and up to eight hardware lights.
//
// This stage is where this game's UI hides panes. Every J2D pane draws with opaque white
// vertex colours and a channel control that enables lighting with NO lights and the ambient
// from a register — so the channel output is vertexColour x ambientRegister, and the pane's
// visibility rides the ambient register's alpha: 0 for a hidden pane (the alpha test then
// rejects every pixel), 255 for the shown one. A renderer that skips the channel and takes
// the vertex colour raw paints every hidden dialog opaque — the overprinted select screen.
//
// Every register layout here is pinned from the game's own GX library, not a wiki:
//   - GXSetChanCtrl (0x801F65xx, writes XF 0x100E+chan): material source at bit 0 (1 =
//     vertex), lighting enable at bit 1, lights 0-3 at bits 2-5, ambient source at bit 6,
//     diffuse function at bits 7-8, the attenuation encoding at bits 9-10 (bit 9 is
//     "attnfn != NONE(2)", bit 10 is "attnfn != SPEC(0)"), lights 4-7 at bits 11-14.
//   - GXLoadLightObjImm (0x801F60A0): light data at XF 0x600 + index*16 — three reserved
//     words, the colour (RGBA, red high), cos-attenuation a0,a1,a2, distance-attenuation
//     k0,k1,k2, position x,y,z, direction x,y,z.
//   - GXLoadNrmMtxImm (0x801FA1CC): the normal matrix at XF 0x400 + index*3, nine floats,
//     row-major 3x3, sharing the position-matrix index.
// The ambient/material registers are XF 0x100A/0x100B and 0x100C/0x100D (channel 0/1), the
// colour controls 0x100E/0x100F, the alpha controls 0x1010/0x1011, the channel count 0x1009.

import "math"

// xfColor unpacks an XF colour register (ambient, material, or a light's colour): RGBA with
// red in the high byte, each channel scaled to [0,1].
func (g *gpu) xfColor(addr int) [4]float32 {
	w := uint32(0)
	if addr >= 0 && addr < len(g.XFMem) {
		w = g.XFMem[addr]
	}
	return [4]float32{
		float32(w>>24) / 255,
		float32(w>>16&0xFF) / 255,
		float32(w>>8&0xFF) / 255,
		float32(w&0xFF) / 255,
	}
}

// lightMask extracts the eight-light enable mask from a channel control word: lights 0-3 at
// bits 2-5, lights 4-7 at bits 11-14.
func lightMask(ctl uint32) uint32 {
	return (ctl>>2)&0xF | (ctl>>7)&0xF0
}

// rasChannel computes colour channel 0 for one vertex: the value the rasteriser interpolates
// and the TEV reads as its RAS input. The vertex colour, eye-space position and eye-space
// normal come from the fetch; the channel controls decide what mixes into the result.
func (g *gpu) rasChannel(m *Machine, vr, vg, vb, va uint8, ex, ey, ez, nx, ny, nz float32) (r, gg, b, a uint8) {
	if g.XFMem[0x1009] == 0 { // no colour channels: the TEV's RAS input reads as zero
		return 0, 0, 0, 0
	}

	vtx := [4]float32{float32(vr) / 255, float32(vg) / 255, float32(vb) / 255, float32(va) / 255}
	rgb := g.evalChannel(m, g.XFMem[0x100E], false, vtx, g.xfColor(0x100A), g.xfColor(0x100C), ex, ey, ez, nx, ny, nz)
	alpha := g.evalChannel(m, g.XFMem[0x1010], true, vtx, g.xfColor(0x100A), g.xfColor(0x100C), ex, ey, ez, nx, ny, nz)
	return chanU8(rgb[0]), chanU8(rgb[1]), chanU8(rgb[2]), chanU8(alpha[3])
}

func chanU8(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

// evalChannel runs one channel control (colour or alpha) for one vertex. For the alpha
// control only the [3] lane of the result is meaningful; for the colour control the [0..2]
// lanes are. Both use the same equation, the alpha one reading each colour's alpha.
func (g *gpu) evalChannel(m *Machine, ctl uint32, isAlpha bool, vtx, amb, mat [4]float32,
	ex, ey, ez, nx, ny, nz float32) [4]float32 {

	if ctl&1 != 0 { // material from the vertex colour
		mat = vtx
	}
	if ctl&2 == 0 { // lighting disabled: the material passes through
		return mat
	}

	if ctl&(1<<6) != 0 { // ambient from the vertex colour
		amb = vtx
	}
	illum := amb

	mask := lightMask(ctl)
	diffFn := (ctl >> 7) & 3
	attnNotNone := ctl&(1<<9) != 0
	attnNotSpec := ctl&(1<<10) != 0

	for i := 0; i < 8; i++ {
		if mask&(1<<i) == 0 {
			continue
		}
		base := 0x600 + i*16
		lcol := g.xfColor(base + 3)
		lpx := g.xfFloat(base + 10)
		lpy := g.xfFloat(base + 11)
		lpz := g.xfFloat(base + 12)

		// The vector from the vertex to the light, and its length.
		dx, dy, dz := lpx-ex, lpy-ey, lpz-ez
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist > 0 {
			dx, dy, dz = dx/dist, dy/dist, dz/dist
		}

		// The diffuse factor by the channel's diffuse function.
		diff := float32(1)
		if diffFn != 0 {
			diff = nx*dx + ny*dy + nz*dz
			if diffFn == 2 && diff < 0 { // CLAMP
				diff = 0
			}
		}

		// The attenuation. SPOT evaluates the cos polynomial on the angle between the
		// light's aim direction and the light-to-vertex ray, over the distance polynomial;
		// NONE is flat; SPEC (the half-angle form) has no user yet and halts loudly rather
		// than shade wrong.
		attn := float32(1)
		if attnNotSpec {
			if attnNotNone { // SPOT
				ldx := g.xfFloat(base + 13)
				ldy := g.xfFloat(base + 14)
				ldz := g.xfFloat(base + 15)
				cs := -(dx*ldx + dy*ldy + dz*ldz)
				if cs < 0 {
					cs = 0
				}
				a0, a1, a2 := g.xfFloat(base+4), g.xfFloat(base+5), g.xfFloat(base+6)
				k0, k1, k2 := g.xfFloat(base+7), g.xfFloat(base+8), g.xfFloat(base+9)
				num := a0 + a1*cs + a2*cs*cs
				if num < 0 {
					num = 0
				}
				den := k0 + k1*dist + k2*dist*dist
				if den != 0 {
					attn = num / den
				} else {
					attn = 0
				}
			}
		} else if attnNotNone { // SPEC
			m.CPU.Halt("XF channel: specular attenuation not yet implemented (ctl 0x%04X)", ctl)
			return mat
		}

		if isAlpha {
			illum[3] += attn * diff * lcol[3]
		} else {
			illum[0] += attn * diff * lcol[0]
			illum[1] += attn * diff * lcol[1]
			illum[2] += attn * diff * lcol[2]
		}
	}

	for k := range illum {
		if illum[k] < 0 {
			illum[k] = 0
		}
		if illum[k] > 1 {
			illum[k] = 1
		}
	}
	return [4]float32{mat[0] * illum[0], mat[1] * illum[1], mat[2] * illum[2], mat[3] * illum[3]}
}

// normalToEye takes a model-space normal through the normal matrix that shares the position
// matrix's index (XF 0x400 + index*3, nine floats row-major), then normalises it.
func (g *gpu) normalToEye(mtxIdx int, nx, ny, nz float32) (ox, oy, oz float32) {
	base := 0x400 + mtxIdx*3
	ox = g.xfFloat(base+0)*nx + g.xfFloat(base+1)*ny + g.xfFloat(base+2)*nz
	oy = g.xfFloat(base+3)*nx + g.xfFloat(base+4)*ny + g.xfFloat(base+5)*nz
	oz = g.xfFloat(base+6)*nx + g.xfFloat(base+7)*ny + g.xfFloat(base+8)*nz
	l := float32(math.Sqrt(float64(ox*ox + oy*oy + oz*oz)))
	if l > 0 {
		ox, oy, oz = ox/l, oy/l, oz/l
	}
	return
}
