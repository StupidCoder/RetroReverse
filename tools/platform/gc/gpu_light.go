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

// rasChannel computes BOTH colour channels for one vertex: the values the rasteriser
// interpolates and each TEV stage selects between as its RAS input (see gpu_tev_state.go's
// rasSel). The vertex colour, eye-space position and eye-space normal come from the fetch; the
// channel controls decide what mixes into each result.
//
// Two channels, not one. The channel COUNT register says how many the transform unit produces,
// and a channel past that count is not computed — the hardware leaves it alone, and a stage
// naming it reads whatever the rasteriser last had. Here an uncomputed channel reads as zero,
// which is what the count-0 case has always done.
//
// This game's scene draws set the count to 2 and split the work: channel 0 carries the room's
// spot lights, channel 1 carries a specular highlight from a light the ambient channel never
// sees. A renderer that computed only channel 0 and handed it to every stage gave the specular
// stage the diffuse colour — every surface got a highlight-shaped multiply of the wrong term.
func (g *gpu) rasChannel(m *Machine, vr, vg, vb, va uint8, ex, ey, ez, nx, ny, nz float32) [2][4]uint8 {
	var out [2][4]uint8
	nchan := int(g.XFMem[0x1009])
	vtx := [4]float32{float32(vr) / 255, float32(vg) / 255, float32(vb) / 255, float32(va) / 255}

	for c := 0; c < 2 && c < nchan; c++ {
		amb := g.xfColor(0x100A + c)
		mat := g.xfColor(0x100C + c)
		rgb := g.evalChannel(m, g.XFMem[0x100E+uint32(c)], false, vtx, amb, mat, ex, ey, ez, nx, ny, nz)
		alpha := g.evalChannel(m, g.XFMem[0x1010+uint32(c)], true, vtx, amb, mat, ex, ey, ez, nx, ny, nz)
		out[c] = [4]uint8{chanU8(rgb[0]), chanU8(rgb[1]), chanU8(rgb[2]), chanU8(alpha[3])}
	}
	return out
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
		ldx := g.xfFloat(base + 13)
		ldy := g.xfFloat(base + 14)
		ldz := g.xfFloat(base + 15)
		a0, a1, a2 := g.xfFloat(base+4), g.xfFloat(base+5), g.xfFloat(base+6)
		k0, k1, k2 := g.xfFloat(base+7), g.xfFloat(base+8), g.xfFloat(base+9)

		// dx,dy,dz is the unit vector from the vertex TOWARDS the light, and dist the distance
		// to it — the two quantities the diffuse and distance terms are built from. SPEC
		// replaces both below: its light has no position to be a distance from.
		dx, dy, dz := lpx-ex, lpy-ey, lpz-ez
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist > 0 {
			dx, dy, dz = dx/dist, dy/dist, dz/dist
		}

		var attn float32
		if !attnNotNone { // NONE: flat, and the distance polynomial is not consulted
			attn = 1
		} else if attnNotSpec { // SPOT
			// The cone term. The stored direction is the light's aim NEGATED — the game's own
			// GXInitLightDir (0x801F5F70) stores -nx,-ny,-nz — and the vertex-to-light vector
			// on the cone's axis is likewise the reverse of the aim, so the two agree at +1 in
			// the middle of the cone and the dot is taken as it stands. This is not a
			// convention worth guessing at: this game's own cos polynomials pin it. Its lights
			// load a = (0, -3.274, 4.274) and (0, -4.530, 5.530), which evaluate to exactly 1.0
			// at cs = +1 and fall through zero around cs = 0.77 — curves built to peak on the
			// axis. Negating the dot puts that peak BEHIND the light: the cone goes dark, the
			// room behind it lights up at 7.5x, and every shadow the flashlight casts sprays
			// outwards as though the lamp were a bare point source.
			cs := dx*ldx + dy*ldy + dz*ldz
			if cs < 0 {
				cs = 0
			}
			num := a0 + a1*cs + a2*cs*cs
			if num < 0 {
				num = 0
			}
			den := k0 + k1*dist + k2*dist*dist
			if den != 0 {
				attn = num / den
			}
		} else { // SPEC
			// The specular form, and it reads the light object's fields as different things:
			// GXInitSpecularDir (0x801F5F8C) writes pos = 2^20 * -n (a direction, parked far
			// enough away to be parallel) and dir = normalize((-nx, -ny, 1-nz)) (the half-angle
			// vector between the light and the eye's own +z axis). Pinned against this scene's
			// light 7, whose pos/2^20 gives the unit n = (0.5, 0.7071, -0.5) and whose stored
			// dir is exactly normalize((-0.5, -0.7071, 1.5)) = (-0.2887, -0.4082, 0.8660).
			//
			// So the light direction is the normalised POSITION, and the value the polynomials
			// are evaluated on is the normal dotted with the HALF-ANGLE — a highlight that
			// sharpens as a2 grows, rather than a cone in space. A surface facing away from the
			// light has no highlight at all, whatever the half-angle says, which is what the
			// facing test is for.
			dx, dy, dz = normalize3f(lpx, lpy, lpz)
			cs := float32(0)
			if nx*dx+ny*dy+nz*dz >= 0 {
				cs = nx*ldx + ny*ldy + nz*ldz
				if cs < 0 {
					cs = 0
				}
			}
			num := a0 + a1*cs + a2*cs*cs
			if num < 0 {
				num = 0
			}
			// The distance polynomial is evaluated on the same cosine — there is no distance —
			// and the hardware normalises its coefficients unless the diffuse function is NONE.
			kx, ky, kz := k0, k1, k2
			if diffFn != 0 {
				kx, ky, kz = normalize3f(k0, k1, k2)
			}
			den := kx + ky*cs + kz*cs*cs
			if den != 0 {
				attn = num / den
			}
		}

		// The diffuse factor by the channel's diffuse function, against whichever direction the
		// attenuation form above settled on.
		diff := float32(1)
		if diffFn != 0 {
			diff = nx*dx + ny*dy + nz*dz
			if diffFn == 2 && diff < 0 { // CLAMP
				diff = 0
			}
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

// normalize3f scales a vector to unit length, leaving a degenerate one alone.
func normalize3f(x, y, z float32) (float32, float32, float32) {
	l := float32(math.Sqrt(float64(x*x + y*y + z*z)))
	if l == 0 {
		return x, y, z
	}
	return x / l, y / l, z / l
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
