// xrtiles.js — the GPU half of the AR tilemap renderer.
//
// Takes the plain arrays tileatlas.js produced and draws the whole map in ONE
// draw call: a handful of brick quads carrying nothing but a cell coordinate,
// and a fragment shader that resolves cell -> (block ->) tile -> slot -> array
// layer. The map's cost stops depending on how big the map is.
//
// Three things in here are load-bearing and easy to get subtly wrong:
//
//   * derivatives come from the CELL coordinate, never the atlas coordinate.
//     dFdx of an atlas UV jumps at every cell boundary and would select the
//     1x1 mip along a one-pixel-wide diagonal stripe across the whole map.
//   * each tile is its own ARRAY LAYER, so its mip chain never crosses into a
//     neighbouring tile. That is the thing a 2-D atlas cannot do with the
//     gutters these levels ship (0 or 1) and the reason for sampler2DArray.
//   * below one texel per tile no atlas scheme can be right at all, because
//     the correct average there is over neighbouring MAP CELLS. A cell-
//     resolution 'far' texture takes over, cross-faded across one lod unit.

const VERT = /* glsl */`
in vec2 aCell;
out vec2 vCell;
void main() {
  vCell = aCell;
  gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
}`;

const FRAG = /* glsl */`
precision highp float;
precision highp sampler2DArray;

in vec2 vCell;
uniform sampler2DArray uTiles;
uniform sampler2D uIndex;
uniform sampler2D uSlot;
uniform sampler2D uFar;
#ifdef BLOCKS
uniform sampler2D uBlocks;
#endif
uniform vec2 uGrid;     // cells across, down
uniform float uSub;     // tiles per cell axis (blocks.size, else 1)
uniform float uTs;      // tile size in texels
uniform float uP;       // tiles packed per array-layer axis
uniform float uLodMax;  // log2(tile size): one texel per tile
out vec4 fragColor;

int u16(vec4 t) { return int(t.r * 255.0 + 0.5) + 256 * int(t.g * 255.0 + 0.5); }

void main() {
  vec2 c = clamp(floor(vCell), vec2(0.0), uGrid - 1.0);
  vec2 f = vCell - c;

  vec4 ix = texelFetch(uIndex, ivec2(c), 0);
  float flags = ix.a * 255.0;
  if (flags >= 128.0) discard;              // cell is entirely backdrop
  int id = u16(ix);
  int variant = int(ix.b * 255.0 + 0.5);

  if (mod(flags, 2.0) >= 1.0) f.x = 1.0 - f.x;   // hflipMask, mirrors the cell

  vec2 sub = floor(f * uSub);
  f = f * uSub - sub;
#ifdef BLOCKS
  id = u16(texelFetch(uBlocks, ivec2(int(sub.x + uSub * sub.y), id), 0));
#endif

  int slot = u16(texelFetch(uSlot, ivec2(id, variant), 0));
  float lay = floor(float(slot) / (uP * uP));
  float k = float(slot) - lay * uP * uP;
  vec2 tuv = (vec2(mod(k, uP), floor(k / uP)) + clamp(f, 0.0, 1.0)) / uP;

  vec2 dx = dFdx(vCell) * uSub / uP;
  vec2 dy = dFdy(vCell) * uSub / uP;
  vec4 col = textureGrad(uTiles, vec3(tuv, lay), dx, dy);

  // Mip-averaged texels are premultiplied (keyed ones are (0,0,0,0)), so undo
  // that before mixing with the straight-alpha far texture.
  vec3 near = col.rgb / max(col.a, 1e-4);

  float texels = max(length(dx), length(dy)) * uTs * uP;
  float lod = log2(max(texels, 1e-6));
  float m = clamp(lod - (uLodMax - 1.0), 0.0, 1.0);
  vec4 farc = texture(uFar, vCell / uGrid);
  vec4 outc = vec4(mix(near, farc.rgb, m), mix(col.a, farc.a, m));

#ifdef BLEND_PASS
  fragColor = vec4(outc.rgb * outc.a, outc.a);
#else
  if (outc.a < 0.5) discard;
  fragColor = vec4(outc.rgb, 1.0);
#endif
  fragColor = linearToOutputTexel(fragColor);
}`;

export class TileMapMesh {
  // (THREE, model, { renderer, blend, aniso })
  constructor(THREE, model, opts = {}) {
    this.THREE = THREE;
    this.model = model;
    const { renderer, blend = false, aniso = 4 } = opts;
    const m = model;

    const tiles = new THREE.DataArrayTexture(m.tiles.data, m.tiles.size, m.tiles.size, m.tiles.layers);
    tiles.format = THREE.RGBAFormat;
    tiles.type = THREE.UnsignedByteType;
    tiles.colorSpace = THREE.SRGBColorSpace;
    tiles.magFilter = THREE.NearestFilter;            // game pixels stay pixels up close
    tiles.minFilter = THREE.LinearMipmapLinearFilter; // clean from across the room
    tiles.generateMipmaps = true;
    tiles.wrapS = tiles.wrapT = THREE.ClampToEdgeWrapping;
    if (renderer && aniso > 1) tiles.anisotropy = Math.min(aniso, renderer.capabilities.getMaxAnisotropy());
    tiles.needsUpdate = true;

    const index = dataTex(THREE, m.index.data, m.index.w, m.index.h);
    const slot = dataTex(THREE, m.slot.data, m.slot.w, m.slot.h);
    const blocks = m.blocks ? dataTex(THREE, m.blocks.data, m.blocks.w, m.blocks.h) : null;

    const far = dataTex(THREE, m.far.data, m.far.w, m.far.h);
    far.colorSpace = THREE.SRGBColorSpace;
    far.magFilter = THREE.LinearFilter;
    far.minFilter = THREE.LinearMipmapLinearFilter;
    far.generateMipmaps = true;
    if (renderer && aniso > 1) far.anisotropy = Math.min(aniso, renderer.capabilities.getMaxAnisotropy());
    far.needsUpdate = true;

    this.textures = { tiles, index, slot, blocks, far };

    // Values, not empty strings: three emits '#define NAME <value>' verbatim,
    // and an empty value is fragile enough to be worth not relying on.
    const defines = {};
    if (blocks) defines.BLOCKS = 1;
    if (blend) defines.BLEND_PASS = 1;

    this.material = new THREE.ShaderMaterial({
      glslVersion: THREE.GLSL3,
      defines,
      side: THREE.DoubleSide,
      transparent: blend,
      depthWrite: !blend,
      uniforms: {
        uTiles: { value: tiles }, uIndex: { value: index }, uSlot: { value: slot },
        uFar: { value: far }, uBlocks: { value: blocks },
        uGrid: { value: new THREE.Vector2(m.grid.w, m.grid.h) },
        uSub: { value: m.sub }, uTs: { value: m.ts }, uP: { value: m.P },
        uLodMax: { value: Math.log2(m.ts) },
      },
      vertexShader: VERT,
      fragmentShader: FRAG,
    });

    // Brick quads: position in map pixels centred on the origin with +Y up
    // (the convention the AR placement already uses), aCell in cells with +Y
    // down and the origin top-left (the convention the index texture uses).
    const n = m.bricks.length;
    const pos = new Float32Array(n * 6 * 3);
    const cell = new Float32Array(n * 6 * 2);
    const cw = m.ts * m.sub;
    const W = m.mapPx.w, H = m.mapPx.h;
    m.bricks.forEach((b, i) => {
      const X0 = b.x * cw - W / 2, X1 = (b.x + b.w) * cw - W / 2;
      const Y0 = H / 2 - b.y * cw, Y1 = H / 2 - (b.y + b.h) * cw;
      const c0 = b.x, c1 = b.x + b.w, r0 = b.y, r1 = b.y + b.h;
      const p = i * 18, t = i * 12;
      const P = [X0, Y0, 0, X1, Y0, 0, X1, Y1, 0, X0, Y0, 0, X1, Y1, 0, X0, Y1, 0];
      const T = [c0, r0, c1, r0, c1, r1, c0, r0, c1, r1, c0, r1];
      for (let k = 0; k < 18; k++) pos[p + k] = P[k];
      for (let k = 0; k < 12; k++) cell[t + k] = T[k];
    });
    this.geometry = new THREE.BufferGeometry();
    this.geometry.setAttribute('position', new THREE.BufferAttribute(pos, 3));
    this.geometry.setAttribute('aCell', new THREE.BufferAttribute(cell, 2));

    this.mesh = new THREE.Mesh(this.geometry, this.material);
    this.mesh.frustumCulled = false;
    this.group = new THREE.Group();
    this.group.add(this.mesh);

    // ---- animation state: everything writes the slot LUT, nothing repaints ----
    this._tileAnims = (m.tm?.tileAnims || []).map(() => ({ acc: 0, step: 0 }));
    // Start where the model started: a tile animation's step 0 is frames[0].
    this._overrides = new Map(m.initialOverrides);
    this._cycle = { acc: 0, step: 0 };
    this._dirty = false;
  }

  get counts() {
    return { ...this.model.stats, draws: 1, tris: this.model.bricks.length * 2 };
  }

  // advance(df) takes ENGINE FRAMES, like every other clock in the 2-D stack.
  // A tile animation step or a cycle step rewrites a few texels of the slot
  // LUT — at most 14 KB — and nothing else happens.
  advance(df) {
    const anims = this.model.tm?.tileAnims || [];
    for (let i = 0; i < anims.length; i++) {
      const ta = anims[i], st = this._tileAnims[i];
      if (!st) continue;
      st.acc += df;
      const period = ta.periodFrames || 10;
      let changed = false;
      while (st.acc >= period) { st.acc -= period; st.step++; changed = true; }
      if (changed) {
        const fr = ta.frames[st.step % ta.frames.length];
        ta.tiles.forEach((t, k) => this._overrides.set(t, fr[k]));
        this._dirty = true;
      }
    }
    const cyc = this.model.cycle;
    if (cyc) {
      this._cycle.acc += df;
      while (this._cycle.acc >= cyc.periodFrames) {
        this._cycle.acc -= cyc.periodFrames;
        this._cycle.step = (this._cycle.step + 1) % cyc.steps;
        this._dirty = true;
      }
    }
    if (this._dirty) {
      this.model.writeSlots(this._overrides, this._cycle.step);
      this.textures.slot.needsUpdate = true;
      this._dirty = false;
    }
  }

  dispose() {
    this.geometry.dispose();
    this.material.dispose();
    for (const t of Object.values(this.textures)) t?.dispose();
  }
}

function dataTex(THREE, data, w, h) {
  const t = new THREE.DataTexture(data, w, h);
  t.magFilter = t.minFilter = THREE.NearestFilter;
  t.generateMipmaps = false;
  t.needsUpdate = true;
  return t;
}
