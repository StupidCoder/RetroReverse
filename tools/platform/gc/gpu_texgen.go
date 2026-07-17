package gc

// gpu_texgen.go is the transform unit's texture-coordinate generator — the stage between the
// vertex and the rasteriser that decides, per coordinate, WHERE in a texture this vertex
// samples. It is the part of the XF that gpu_xf.go's position path has no counterpart for: a
// position has one destination and needs one matrix, while a vertex may carry up to eight
// texture coordinates, each generated from a different source through a different matrix.
//
// A coordinate does not have to come off the vertex. That is the whole point of the stage and
// the reason it exists here: the generator can take the vertex's POSITION as its input and run
// it through a matrix that describes some other camera, and what comes out is where this
// vertex lands in that camera's view. Do it with the matrix of the light and the texture being
// sampled is the light's own depth buffer — which is projective shadow mapping, and is what
// this game asks for when it draws the map into Luigi's hands.
//
// The register layout, and the order the two matrices apply in:
//
//	XF 0x103F  the number of coordinates to generate — a count, not a count-1 (BP 0x00's own
//	           ntex field carries the same number, which is what pins it)
//	XF 0x1040+i  coordinate i's source row, input form, generation type and projection
//	XF 0x1050+i  coordinate i's POST-transform matrix: the second matrix, and on this game's
//	             shadow draws the one that matters — see below
//	XF 0x1012  whether that second matrix applies at all
//
// The two-matrix arrangement is the trap in this stage. A coordinate is transformed by its
// texture matrix (selected per coordinate by the matrix-index register) and THEN, if the dual
// transform is enabled, by a post-transform matrix from a separate bank. This game leaves the
// texture matrix as the identity and puts the entire light-space projection in the post
// matrix, so a generator that implemented the first matrix and skipped the second would
// multiply by identity, emit the model-space position as a texture coordinate, and look for
// all the world like a working texgen that had simply been handed a bad matrix.

import "math"

// maxTexCoord is the eight coordinates the hardware generates and interpolates.
const maxTexCoord = 8

// texCoord is one generated coordinate. q is the projective divisor: the generator writes 1
// into it for a non-projective coordinate, which makes the rasteriser's divide a no-op rather
// than a special case (see perspTexCoords).
type texCoord struct{ s, t, q float32 }

// The generation-type field.
//
// texGenColor0/1 are the stage that draws this game's flashlight, and they are worth stating
// plainly because they invert what a texture coordinate normally is. Instead of asking WHERE
// this vertex sits, they take the colour the lighting channel just computed for it (see
// gpu_light.go) and use that BRIGHTNESS as the coordinate — s = channel.r, t = channel.g. The
// texture then is not a picture of anything; it is a lookup table from "how lit is this vertex"
// to "what colour should that be", which is how the hardware buys a non-linear falloff out of a
// lighting unit that can only do polynomials. Luigi's cone is a 64x64 ramp sampled that way.
//
// The consequence for anyone reading the pipeline in order: this generator MUST run after the
// colour channels, not alongside them. It is the one texture coordinate that depends on the
// lighting result rather than the vertex.
const (
	texGenRegular = 0
	texGenEmboss  = 1
	texGenColor0  = 2
	texGenColor1  = 3
)

// The source-row field. The rows are the vertex's attributes as the XF sees them, which is why
// a position and a texture coordinate are both just "a row" to the generator.
const (
	texSrcGeom   = 0 // the vertex position — the projective source
	texSrcNormal = 1
	texSrcTex0   = 5 // ...through texSrcTex0+7 for the vertex's own eight coordinates
)

// texGenCount is how many coordinates the transform unit generates for this draw.
func (g *gpu) texGenCount() int {
	n := int(g.XFMem[0x103F] & 0xF)
	if n > maxTexCoord {
		n = maxTexCoord
	}
	return n
}

// texMtxRow is the matrix-memory row coordinate i's texture matrix starts at. The eight
// indices are packed six bits apiece across two registers, above the position matrix's own
// index in the low six bits of the first — the same register the position path reads.
func (g *gpu) texMtxRow(i int) int {
	if i < 4 {
		return int((g.CPReg[0x30] >> (6 + 6*uint32(i))) & 0x3F)
	}
	return int((g.CPReg[0x31] >> (6 * uint32(i-4))) & 0x3F)
}

// genTexCoord generates coordinate i for one vertex.
//
// mx,my,mz is the model-space position and nx,ny,nz the model-space normal — the two rows the
// generator can take as input besides the vertex's own coordinates, which arrive in vtx.
// mtxRow overrides the register-selected matrix row when the vertex carried its own index.
func (g *gpu) genTexCoord(m *Machine, i int, mtxRow int, mx, my, mz, nx, ny, nz float32, vtc *[maxTexCoord]texCoord, col *[2][4]uint8) texCoord {
	info := g.XFMem[0x1040+uint32(i)]
	proj := (info >> 1) & 1
	inputForm := (info >> 2) & 3
	genType := (info >> 4) & 7
	srcRow := (info >> 7) & 0x1F

	switch genType {
	case texGenRegular:
	case texGenColor0, texGenColor1:
		// The lit channel's red and green ARE the coordinate. No matrix, no projection: the
		// generator writes them straight through, scaled from the channel's byte range to the
		// 0..1 the sampler wants, with the divisor forced to 1 so the rasteriser's unconditional
		// divide leaves them alone.
		c := col[genType-texGenColor0]
		return texCoord{s: float32(c[0]) / 255, t: float32(c[1]) / 255, q: 1}
	default:
		m.logf("XF: texgen %d uses generation type %d, only the regular and colour types are implemented", i, genType)
		return texCoord{q: 1}
	}

	// The input row. A row is four components to the matrix that follows; the input form says
	// whether the third one is the source's own or a constant 1.
	var in [4]float32
	switch {
	case srcRow == texSrcGeom:
		in = [4]float32{mx, my, mz, 1}
	case srcRow == texSrcNormal:
		in = [4]float32{nx, ny, nz, 1}
	case srcRow >= texSrcTex0 && srcRow < texSrcTex0+maxTexCoord:
		// Any of the vertex's own eight coordinates. Which one is not tied to which coordinate
		// is being generated: a draw may generate coordinate 2 from the vertex's coordinate 1.
		v := vtc[srcRow-texSrcTex0]
		in = [4]float32{v.s, v.t, 1, 1}
	default:
		// The binormal and tangent rows, which the emboss stage reads and the fetch does not
		// capture; a named gap, not a wrong coordinate.
		m.logf("XF: texgen %d sources row %d, which the vertex fetch does not capture", i, srcRow)
		return texCoord{q: 1}
	}
	if inputForm == 0 { // AB11: the third component is a constant, not the source's
		in[2] = 1
	}

	// The texture matrix: two rows for a plain coordinate, three when the coordinate is
	// projective and needs a divisor.
	base := mtxRow * 4
	dot := func(r int) float32 {
		return g.xfFloat(base+r*4)*in[0] + g.xfFloat(base+r*4+1)*in[1] +
			g.xfFloat(base+r*4+2)*in[2] + g.xfFloat(base+r*4+3)*in[3]
	}
	out := texCoord{s: dot(0), t: dot(1), q: 1}
	if proj == 1 {
		out.q = dot(2)
	}

	// The post-transform matrix, from its own bank at 0x0500. It consumes the first matrix's
	// three outputs as a point, so a coordinate that skipped the third row above brings a 1
	// into this one.
	if g.XFMem[0x1012]&1 != 0 {
		pi := g.XFMem[0x1050+uint32(i)]
		pbase := 0x500 + int(pi&0x3F)*4
		v := [3]float32{out.s, out.t, out.q}
		if (pi>>8)&1 != 0 {
			v = normalize3(v)
		}
		pdot := func(r int) float32 {
			return g.xfFloat(pbase+r*4)*v[0] + g.xfFloat(pbase+r*4+1)*v[1] +
				g.xfFloat(pbase+r*4+2)*v[2] + g.xfFloat(pbase+r*4+3)
		}
		out = texCoord{s: pdot(0), t: pdot(1), q: pdot(2)}
	}

	// Only a projective coordinate is divided by its third component. A plain one is used as
	// it stands, so its divisor is forced back to 1 — which is what lets the rasteriser divide
	// unconditionally instead of carrying the projection bit into the pixel loop. Note this
	// must happen AFTER the post transform, whose third row writes a value into q that a
	// non-projective coordinate is not entitled to use.
	if proj != 1 {
		out.q = 1
	}
	return out
}

// normalize3 scales a vector to unit length, leaving a degenerate one alone.
func normalize3(v [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	if l == 0 {
		return v
	}
	return [3]float32{v[0] / l, v[1] / l, v[2] / l}
}
