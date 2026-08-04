// xrspike.js — a one-push capability probe for the tilemap renderer.
//
// The renderer that replaces the map snapshot rests on four things that a
// desktop cannot vouch for: WebGL2 GLSL3 shaders under an immersive session,
// a DataArrayTexture (sampler2DArray) with a working mip chain, texelFetch on
// an 8-bit index texture, and textureGrad with hand-supplied derivatives. If
// any of them misbehaves on the headset's driver the design has to change, and
// finding that out after building three modules would be expensive.
//
// So ?xrspike=1 puts exactly those four things in front of the user and reports
// the limits that shape the design. It draws a 4x2 grid of cells resolved
// through the same cell -> id -> slot -> array path the real renderer will use;
// each cell shows a distinct colour with a fine checkerboard, so filtering and
// mip behaviour are visible rather than inferred.
//
// What to look for in the headset:
//   * eight cells in two rows, four distinct colours, checkerboard visible up
//     close  -> GLSL3 + array texture + texelFetch + textureGrad all work
//   * BLACK  -> the array texture is mip-incomplete (generateMipmaps failed);
//     the fallback is to supply texture.mipmaps by hand
//   * stepping back should blur the checkerboard smoothly, not shimmer -> mips
//     are being selected by the supplied gradients

const TS = 16;      // texels per tile layer
const COLS = 4;     // cells across
const ROWS = 2;     // cells down

export function buildSpike(THREE, renderer, W, H) {
  const gl = renderer.getContext();
  const caps = renderer.capabilities;

  // ---- the tile array: one layer per "tile", each a flat colour + checker ----
  const colours = [[0xff, 0x40, 0x40], [0x40, 0xff, 0x60], [0x50, 0x90, 0xff], [0xff, 0xd0, 0x40]];
  const layers = colours.length;
  const data = new Uint8Array(TS * TS * 4 * layers);
  for (let l = 0; l < layers; l++) {
    const [r, g, b] = colours[l];
    for (let y = 0; y < TS; y++) {
      for (let x = 0; x < TS; x++) {
        const dark = ((x ^ y) & 1) === 1; // 1-texel checker: the finest thing mips must handle
        const o = (l * TS * TS + y * TS + x) * 4;
        data[o] = dark ? r >> 1 : r;
        data[o + 1] = dark ? g >> 1 : g;
        data[o + 2] = dark ? b >> 1 : b;
        data[o + 3] = 255;
      }
    }
  }
  const tiles = new THREE.DataArrayTexture(data, TS, TS, layers);
  tiles.format = THREE.RGBAFormat;
  tiles.type = THREE.UnsignedByteType;
  tiles.colorSpace = THREE.SRGBColorSpace;
  tiles.magFilter = THREE.NearestFilter;
  tiles.minFilter = THREE.LinearMipmapLinearFilter;
  tiles.generateMipmaps = true;
  tiles.wrapS = tiles.wrapT = THREE.ClampToEdgeWrapping;
  tiles.anisotropy = Math.min(4, caps.getMaxAnisotropy());
  tiles.needsUpdate = true;

  // ---- the index texture: one texel per cell, LE16 id in R,G ----
  const idx = new Uint8Array(COLS * ROWS * 4);
  for (let i = 0; i < COLS * ROWS; i++) {
    const id = i % layers;
    idx[i * 4] = id & 255;
    idx[i * 4 + 1] = (id >> 8) & 255;
    idx[i * 4 + 2] = 0;
    idx[i * 4 + 3] = 0;
  }
  const index = new THREE.DataTexture(idx, COLS, ROWS);
  index.magFilter = index.minFilter = THREE.NearestFilter;
  index.needsUpdate = true;

  // ---- the slot LUT: id -> layer. The real renderer animates by writing here ----
  const slotData = new Uint8Array(layers * 4);
  for (let i = 0; i < layers; i++) { slotData[i * 4] = i & 255; slotData[i * 4 + 1] = (i >> 8) & 255; }
  const slot = new THREE.DataTexture(slotData, layers, 1);
  slot.magFilter = slot.minFilter = THREE.NearestFilter;
  slot.needsUpdate = true;

  const material = new THREE.ShaderMaterial({
    glslVersion: THREE.GLSL3,
    side: THREE.DoubleSide,
    uniforms: {
      uTiles: { value: tiles }, uIndex: { value: index }, uSlot: { value: slot },
      uGrid: { value: new THREE.Vector2(COLS, ROWS) },
    },
    vertexShader: /* glsl */`
      in vec2 aCell;
      out vec2 vCell;
      void main() {
        vCell = aCell;
        gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
      }`,
    fragmentShader: /* glsl */`
      precision highp float;
      precision highp sampler2DArray;
      // Colour space, the GLSL3 way. '#include <colorspace_fragment>' assigns
      // to gl_FragColor, which does not exist under GLSL3, so call the
      // conversion directly instead — three injects linearToOutputTexel() and
      // the helpers it needs into every ShaderMaterial's prefix already, so
      // including <colorspace_pars_fragment> here is a redefinition error.
      in vec2 vCell;
      uniform sampler2DArray uTiles;
      uniform sampler2D uIndex;
      uniform sampler2D uSlot;
      uniform vec2 uGrid;
      out vec4 fragColor;
      int u16(vec4 t) { return int(t.r * 255.0 + 0.5) + 256 * int(t.g * 255.0 + 0.5); }
      void main() {
        vec2 c = clamp(floor(vCell), vec2(0.0), uGrid - 1.0);
        vec2 f = vCell - c;
        int id = u16(texelFetch(uIndex, ivec2(c), 0));
        int sl = u16(texelFetch(uSlot, ivec2(id, 0), 0));
        // Derivatives come from the CELL coordinate, which is continuous across
        // the whole quad. Taking them from the atlas coordinate instead would
        // explode at every cell boundary and select the 1x1 mip there.
        vec2 dx = dFdx(vCell), dy = dFdy(vCell);
        fragColor = textureGrad(uTiles, vec3(f, float(sl)), dx, dy);
        fragColor = linearToOutputTexel(fragColor);
      }`,
  });

  // One quad spanning the map rect, so the AR fit places it exactly as the real
  // map would be placed.
  const geo = new THREE.BufferGeometry();
  const x0 = -W / 2, x1 = W / 2, y0 = H / 2, y1 = -H / 2;
  geo.setAttribute('position', new THREE.BufferAttribute(new Float32Array([
    x0, y0, 0, x1, y0, 0, x1, y1, 0, x0, y0, 0, x1, y1, 0, x0, y1, 0,
  ]), 3));
  geo.setAttribute('aCell', new THREE.BufferAttribute(new Float32Array([
    0, 0, COLS, 0, COLS, ROWS, 0, 0, COLS, ROWS, 0, ROWS,
  ]), 2));
  const mesh = new THREE.Mesh(geo, material);
  mesh.frustumCulled = false;

  const group = new THREE.Group();
  group.add(mesh);

  // The numbers that decide the real renderer's packing.
  const maxLayers = gl.getParameter(gl.MAX_ARRAY_TEXTURE_LAYERS);
  const note = `spike · webgl2 ${!!caps.isWebGL2} · layers ${maxLayers}`
    + ` · tex ${caps.maxTextureSize} · aniso ${caps.getMaxAnisotropy()}`
    + ` · P for 3520 tiles ${pFor(3520, maxLayers)}`;

  return {
    group, note,
    dispose() {
      geo.dispose(); material.dispose();
      tiles.dispose(); index.dispose(); slot.dispose();
    },
  };
}

// pFor reports the tile-packing factor the worst level in the corpus would
// need: P tiles per layer axis so that ceil(slots / P^2) fits the driver limit.
function pFor(slots, maxLayers) {
  let p = 1;
  while (Math.ceil(slots / (p * p)) > maxLayers && p < 16) p *= 2;
  return p;
}
