// Global screen post-process for the Studio. The per-game viewers render normally; this draws a
// full-screen overlay canvas on top of them and, each frame, captures the active viewer's canvas
// into a texture and re-renders it through a display-appropriate shader. Which shader runs is a
// per-game *profile*, chosen from the game's system:
//
//   crt - tube TV (C64 / Amiga): NTSC signal -> electron beam -> phosphor mask -> glass/glow
//   gb  - Game Boy DMG: reflective 4-shade green dot-matrix with the "hovering pixel" drop shadow
//   gg  - Game Gear: backlit colour LCD, visible pixel grid + subpixels, washed colour, backlight bloom
//
// The overlay is pointer-events:none, so all the viewer's own drag/zoom/rotate interactions pass
// straight through to it. The capture pipeline is display-agnostic (it only consumes the viewer's
// final canvas pixels), so switching profiles never touches any per-game viewer.
//
// The crt shader is a colleague's "physical model" CRT, lightly adapted: the comparison/raw paths
// are dropped (the filter is simply on or off), and the signal resolution is virtualised (a
// console-like ~240-line signal) so the NTSC/dot-crawl artefacts read correctly over content that
// is actually rendered at display resolution. The two LCD shaders quantise the (high-res) content
// into a screen-space cell grid to recreate the dot-matrix look at any zoom. LCD ghosting (slow
// pixel response) is done in canvas-2D land by fading each frame over a persistent buffer.

const VERT = `
attribute vec2 position;
varying vec2 vUv;
void main() {
  vUv = position * 0.5 + 0.5;
  vUv.y = 1.0 - vUv.y;
  gl_Position = vec4(position, 0.0, 1.0);
}`;

const FRAG_CRT = `
#extension GL_OES_standard_derivatives : enable
precision highp float;

uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform vec2 u_texResolution;
uniform float u_time;

uniform float u_noise;
uniform float u_lumaSmear;
uniform float u_iqBlur;
uniform float u_saturation;
uniform float u_phaseSpeed;

uniform float u_scanlineCount;
uniform float u_verticalScale;
uniform float u_beamFocus;
uniform float u_beamBloom;

uniform float u_maskType;
uniform float u_maskStrength;
uniform float u_tvLines;
uniform float u_curvature;
uniform float u_glow;

varying vec2 vUv;
#define PI 3.14159265359

const mat3 rgb2yiq = mat3(0.299, 0.596, 0.211, 0.587, -0.274, -0.523, 0.114, -0.322, 0.312);
const mat3 yiq2rgb = mat3(1.0, 1.0, 1.0, 0.956, -0.272, -1.106, 0.621, -0.647, 1.703);

vec3 toLinear(vec3 sRGB) { return pow(sRGB, vec3(2.4)); }
vec3 toSRGB(vec3 lin) { return pow(lin, vec3(1.0 / 2.2)); }

vec2 curve(vec2 uv) {
  uv = (uv - 0.5) * 2.0;
  uv *= 1.05;
  uv.x *= 1.0 + pow((abs(uv.y) / 5.0), 2.0) * u_curvature;
  uv.y *= 1.0 + pow((abs(uv.x) / 4.0), 2.0) * u_curvature;
  uv = (uv / 2.0) + 0.5;
  uv = uv * 0.92 + 0.04;
  return uv;
}

float gaussian(float x, float sigma) { return exp(-(x * x) / (2.0 * sigma * sigma)); }

// Domain A: NTSC signal
vec3 sampleNTSC(vec2 uv) {
  vec2 contentUV = (uv - 0.5) * vec2(1.0, 1.0 / u_verticalScale) + 0.5;
  if (contentUV.y < 0.0 || contentUV.y > 1.0) return vec3(0.0);
  vec3 col = texture2D(u_texture, contentUV).rgb;
  if (u_iqBlur <= 0.0 && u_lumaSmear <= 0.0) return col;

  vec3 yiq = rgb2yiq * col;
  float w_total = 1.0;
  vec3 yiq_acc = yiq;
  float pixelX = 1.0 / u_texResolution.x;
  float phase = u_time * 60.0 * u_phaseSpeed;

  for (float i = 1.0; i <= 3.0; i += 1.0) {
    float offset = i * pixelX;
    vec3 left = rgb2yiq * texture2D(u_texture, contentUV - vec2(offset * u_iqBlur, 0.0)).rgb;
    vec3 right = rgb2yiq * texture2D(u_texture, contentUV + vec2(offset * u_iqBlur, 0.0)).rgb;
    float lumaDecay = 2.0 / (u_lumaSmear + 0.001);
    float w_y = exp(-i * i * lumaDecay);
    float w_iq = exp(-i * i * 0.5);
    float crawl = sin(contentUV.x * u_texResolution.x * PI + phase);
    w_iq *= (1.0 + 0.3 * crawl);
    yiq_acc.x += (left.x + right.x) * w_y;
    yiq_acc.yz += (left.yz + right.yz) * w_iq;
    w_total += 2.0 * w_y;
  }

  float finalY = yiq_acc.x / w_total;
  float norm = 1.0 + 2.0 * (exp(-1.0 * 0.5) + exp(-4.0 * 0.5) + exp(-9.0 * 0.5));
  vec3 rgb = yiq2rgb * vec3(finalY, yiq_acc.y / norm, yiq_acc.z / norm);
  vec3 gray = vec3(dot(rgb, vec3(0.299, 0.587, 0.114)));
  rgb = mix(gray, rgb, u_saturation);
  return clamp(rgb, 0.0, 1.0);
}

// Domain B: electron beam
vec3 applyBeam(vec2 uv, vec3 inputLinear) {
  float L = dot(inputLinear, vec3(0.2126, 0.7152, 0.0722));
  float y_dist = fract(uv.y * u_scanlineCount) - 0.5;
  float sigma = u_beamFocus + u_beamBloom * L * 0.15;
  return inputLinear * gaussian(y_dist, sigma);
}

// Domain C: phosphor mask
vec3 calculateMaskSample(vec2 uv, float luma) {
  float x_pixel = uv.x * u_resolution.x;
  float y_pixel = uv.y * u_resolution.y;
  float P = u_resolution.x / u_tvLines * 1.5;

  float stagger = 0.0;
  if (u_maskType > 0.5) {
    float row = floor(y_pixel / P);
    if (mod(row, 2.0) > 0.5) stagger = 0.5;
  }
  float phase = (x_pixel / P) + stagger;
  float delta = fwidth(phase);
  float maskFade = 1.0 - smoothstep(0.5, 1.0, delta);
  if (maskFade < 0.01) return vec3(1.0);

  float r_mask = 0.5 + 0.5 * sin(PI * 2.0 * phase);
  float g_mask = 0.5 + 0.5 * sin(PI * 2.0 * phase + 2.0 * PI / 3.0);
  float b_mask = 0.5 + 0.5 * sin(PI * 2.0 * phase + 4.0 * PI / 3.0);

  float v_mask = 1.0;
  if (u_maskType > 0.5) {
    float y_phase = (y_pixel / P) * PI * 2.0;
    v_mask = 0.5 + 0.5 * sin(y_phase);
    v_mask = mix(1.0, v_mask, 0.5);
  }
  vec3 mask = vec3(r_mask, g_mask, b_mask) * v_mask;
  float washout = 1.0 - luma * 0.6;
  float effectiveStrength = u_maskStrength * maskFade * clamp(washout, 0.0, 1.0);
  return mix(vec3(1.0), mask, effectiveStrength);
}

vec3 applyMask(vec2 uv, vec3 inputLinear) {
  float L = dot(inputLinear, vec3(0.2126, 0.7152, 0.0722));
  vec3 accumMask = vec3(0.0);
  float pixelW = 1.0 / u_resolution.x;
  float pixelH = 1.0 / u_resolution.y;
  for (float x = -0.33; x <= 0.33; x += 0.33)
    for (float y = -0.33; y <= 0.33; y += 0.33)
      accumMask += calculateMaskSample(uv + vec2(x * pixelW, y * pixelH), L);
  return inputLinear * (accumMask / 9.0);
}

void main() {
  vec2 uv = vUv;
  vec2 curvedUV = curve(uv);
  vec3 finalCRT = vec3(0.0);

  if (curvedUV.x >= 0.0 && curvedUV.x <= 1.0 && curvedUV.y >= 0.0 && curvedUV.y <= 1.0) {
    vec3 signalVoltage = sampleNTSC(curvedUV);
    float noise = (fract(sin(dot(curvedUV * u_time, vec2(12.9898, 78.233))) * 43758.5453) - 0.5) * u_noise;
    signalVoltage = max(vec3(0.0), signalVoltage + noise);

    vec3 linearLight = toLinear(signalVoltage);
    linearLight = applyBeam(curvedUV, linearLight);
    linearLight = applyMask(curvedUV, linearLight);

    // glass halation glow
    vec3 glowAcc = vec3(0.0);
    float glowWeight = 0.0;
    float blurStride = 2.5 / u_resolution.x;
    for (float x = -2.0; x <= 2.0; x++) {
      for (float y = -2.0; y <= 2.0; y++) {
        vec2 off = vec2(x, y) * blurStride * 2.0;
        vec2 gUV = (curvedUV + off - 0.5) * vec2(1.0, 1.0 / u_verticalScale) + 0.5;
        vec3 s = vec3(0.0);
        if (gUV.y >= 0.0 && gUV.y <= 1.0) s = toLinear(texture2D(u_texture, gUV).rgb);
        float w = 1.0 / (1.0 + dot(vec2(x, y), vec2(x, y)));
        glowAcc += s * 1.5 * w;
        glowWeight += w;
      }
    }
    vec3 finalLinear = linearLight + (glowAcc / glowWeight) * u_glow;
    float vig = 16.0 * curvedUV.x * curvedUV.y * (1.0 - curvedUV.x) * (1.0 - curvedUV.y);
    finalLinear *= pow(vig, 0.2);
    finalCRT = toSRGB(finalLinear);
  }
  gl_FragColor = vec4(finalCRT, 1.0);
}`;

// Game Boy DMG: reflective 4-shade green dot-matrix. The content is quantised into a cell grid that
// tracks the game's own pixels (u_cellSize = one game pixel in overlay px, u_gridOrigin phases the
// grid to the camera pan), so gaps + shadow stay locked to the pixels as you zoom. Each cell's
// luminance is posterised onto the classic pea-green ramp, dark cells cast a soft drop shadow
// down-and-right (the "hovering pixels" effect), and thin gaps separate the cells.
const FRAG_GB = `
precision highp float;
uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform float u_cellSize;    // one game pixel, in overlay device px
uniform vec2 u_gridOrigin;   // game-pixel (0,0) position, in overlay device px
uniform float u_tint;
uniform float u_gridStrength;
uniform float u_shadowOpacity;
uniform float u_shadowOffset; // in cell units
uniform float u_contrast;
varying vec2 vUv;

// classic DMG ramp, darkest -> lightest
const vec3 G0 = vec3(0.059, 0.220, 0.059); // #0f380f
const vec3 G1 = vec3(0.188, 0.384, 0.188); // #306230
const vec3 G2 = vec3(0.545, 0.675, 0.059); // #8bac0f
const vec3 G3 = vec3(0.608, 0.737, 0.059); // #9bbc0f

float luma(vec3 c) { return dot(c, vec3(0.299, 0.587, 0.114)); }

// sample the source at the centre of the grid cell that contains fragment position p (overlay px)
vec3 sampleAt(vec2 p) {
  vec2 cell = floor((p - u_gridOrigin) / u_cellSize);
  vec2 cUv = (u_gridOrigin + (cell + 0.5) * u_cellSize) / u_resolution;
  return texture2D(u_texture, cUv).rgb;
}

vec3 greenRamp(float L) {
  if (L < 0.25) return G0;
  else if (L < 0.5) return G1;
  else if (L < 0.75) return G2;
  return G3;
}

void main() {
  vec2 p = vUv * u_resolution;
  // fade grid/shadow out once cells are too small to resolve (zoomed below native)
  float vis = smoothstep(2.0, 5.0, u_cellSize);

  vec3 src = sampleAt(p);
  float L = clamp((luma(src) - 0.5) * u_contrast + 0.5, 0.0, 1.0);
  vec3 col = mix(src, greenRamp(L), u_tint);

  // drop shadow: a dark cell up-and-left casts a shadow onto this (lighter) cell
  float nbL = luma(sampleAt(p - vec2(u_shadowOffset) * u_cellSize));
  float shadow = u_shadowOpacity * vis * (1.0 - nbL) * L;
  col *= (1.0 - shadow);

  // dot-matrix grid gaps, phased to the game-pixel grid
  vec2 f = fract((p - u_gridOrigin) / u_cellSize);
  float gx = smoothstep(0.0, 0.12, f.x) * smoothstep(0.0, 0.12, 1.0 - f.x);
  float gy = smoothstep(0.0, 0.12, f.y) * smoothstep(0.0, 0.12, 1.0 - f.y);
  col *= mix(1.0 - u_gridStrength * vis, 1.0, gx * gy);

  // faint ambient vignette (reflective panel, no backlight)
  float vig = 16.0 * vUv.x * vUv.y * (1.0 - vUv.x) * (1.0 - vUv.y);
  col *= mix(0.9, 1.0, pow(clamp(vig, 0.0, 1.0), 0.3));

  gl_FragColor = vec4(col, 1.0);
}`;

// Game Gear: backlit colour LCD. Same game-pixel-aligned cell grid as the GB profile, but colour is
// kept (slightly desaturated/gamma'd), with faint RGB subpixel stripes, grid gaps, and a soft
// backlight bloom. No curvature/shadow.
const FRAG_GG = `
precision highp float;
uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform float u_cellSize;    // one game pixel, in overlay device px
uniform vec2 u_gridOrigin;   // game-pixel (0,0) position, in overlay device px
uniform float u_gridStrength;
uniform float u_subpixel;
uniform float u_saturation;
uniform float u_glow;
varying vec2 vUv;

float luma(vec3 c) { return dot(c, vec3(0.299, 0.587, 0.114)); }

vec3 sampleAt(vec2 p) {
  vec2 cell = floor((p - u_gridOrigin) / u_cellSize);
  vec2 cUv = (u_gridOrigin + (cell + 0.5) * u_cellSize) / u_resolution;
  return texture2D(u_texture, cUv).rgb;
}

void main() {
  vec2 p = vUv * u_resolution;
  float vis = smoothstep(2.0, 5.0, u_cellSize);
  vec3 col = sampleAt(p);

  // colour response: desaturate a touch and darken mids (the washed LCD look)
  col = mix(vec3(luma(col)), col, u_saturation);
  col = pow(col, vec3(1.15));

  vec2 f = fract((p - u_gridOrigin) / u_cellSize);

  // RGB subpixel stripes: three columns per cell
  float seg = floor(f.x * 3.0);
  vec3 sp = vec3(0.6);
  if (seg < 0.5) sp.r = 1.0; else if (seg < 1.5) sp.g = 1.0; else sp.b = 1.0;
  col *= mix(vec3(1.0), sp, u_subpixel * vis);

  // grid gaps, phased to the game-pixel grid
  float gx = smoothstep(0.0, 0.1, f.x) * smoothstep(0.0, 0.1, 1.0 - f.x);
  float gy = smoothstep(0.0, 0.1, f.y) * smoothstep(0.0, 0.1, 1.0 - f.y);
  col *= mix(1.0 - u_gridStrength * vis, 1.0, gx * gy);

  // backlight bloom
  vec3 bloom = vec3(0.0);
  float stride = 4.0 / u_resolution.x;
  for (float x = -2.0; x <= 2.0; x++)
    for (float y = -2.0; y <= 2.0; y++)
      bloom += texture2D(u_texture, vUv + vec2(x, y) * stride).rgb;
  col += (bloom / 25.0) * u_glow;

  gl_FragColor = vec4(col, 1.0);
}`;

// Per-profile parameter defaults. `ghost` (>0) enables LCD motion smear via the persistent buffer;
// the crt profile has no ghost key, so it always uploads the crisp current frame.
const PROFILE_DEFAULTS = {
  crt: {
    noise: 0.12, lumaSmear: 0.7, iqBlur: 1.5, saturation: 1.1, phaseSpeed: 0.5,
    beamFocus: 0.55, beamBloom: 2.2, maskType: 1.0, maskStrength: 0.3, tvLines: 400,
    curvature: 0.7, glow: 0.15,
    // line counts (decoupled): signalLines = the vertical resolution of the simulated source
    // signal (drives the NTSC/dot-crawl artefact scale); scanLines = the number of beam scanlines.
    signalLines: 240, scanLines: 240,
  },
  // pixelsPerCell = how many native game pixels make one LCD dot (1 = one dot per pixel, authentic).
  // The actual on-screen cell size is derived per frame from the camera zoom (see _render).
  gb: { pixelsPerCell: 1, tint: 1.0, gridStrength: 0.2, shadowOpacity: 0.25, shadowOffset: 0.5, contrast: 1.15, ghost: 0.0 },
  gg: { pixelsPerCell: 1, gridStrength: 0.2, subpixel: 0.75, saturation: 0.5, glow: 0.18, ghost: 0.0 },
};

// The uniforms each profile's fragment shader declares (cached at link time, uploaded per frame).
const PROFILE_UNIFORMS = {
  crt: ['u_resolution', 'u_texResolution', 'u_time', 'u_noise', 'u_lumaSmear', 'u_iqBlur',
    'u_saturation', 'u_phaseSpeed', 'u_scanlineCount', 'u_verticalScale', 'u_beamFocus',
    'u_beamBloom', 'u_maskType', 'u_maskStrength', 'u_tvLines', 'u_curvature', 'u_glow'],
  gb: ['u_resolution', 'u_cellSize', 'u_gridOrigin', 'u_tint', 'u_gridStrength', 'u_shadowOpacity', 'u_shadowOffset', 'u_contrast'],
  gg: ['u_resolution', 'u_cellSize', 'u_gridOrigin', 'u_gridStrength', 'u_subpixel', 'u_saturation', 'u_glow'],
};

const FRAG = { crt: FRAG_CRT, gb: FRAG_GB, gg: FRAG_GG };

export class ScreenFilter {
  constructor(stageEl) {
    this.stage = stageEl;
    this.enabled = false;
    this.profile = 'crt';
    this.source = () => []; // set by the host: returns the canvases to capture (front-to-back)
    // set by the host: the active 2-D viewer's pixel grid { zoom, ox, oy, screenW } (app-screen px),
    // used to lock the LCD cell grid to the game's own pixels; null when not a map viewer.
    this.pixelGrid = () => null;
    this.paramsByProfile = {};
    for (const k of Object.keys(PROFILE_DEFAULTS)) this.paramsByProfile[k] = { ...PROFILE_DEFAULTS[k] };

    const canvas = document.createElement('canvas');
    canvas.id = 'screenfx';
    canvas.style.display = 'none';
    stageEl.appendChild(canvas);
    this.canvas = canvas;

    const gl = canvas.getContext('webgl', { alpha: false, antialias: false, depth: false });
    this.gl = gl;
    this.ok = !!gl;
    if (!gl) return;
    gl.getExtension('OES_standard_derivatives');

    const vert = compile(gl, gl.VERTEX_SHADER, VERT);
    this.programs = {};
    for (const name of Object.keys(FRAG)) {
      const prog = gl.createProgram();
      gl.attachShader(prog, vert);
      gl.attachShader(prog, compile(gl, gl.FRAGMENT_SHADER, FRAG[name]));
      gl.linkProgram(prog);
      const u = {};
      for (const n of PROFILE_UNIFORMS[name]) u[n] = gl.getUniformLocation(prog, n);
      this.programs[name] = { prog, u, pos: gl.getAttribLocation(prog, 'position') };
    }

    this.buf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, -1, 1, 1, -1, 1, 1]), gl.STATIC_DRAW);

    this.tex = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, this.tex);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

    // scratch = this frame's composited sources; persist = the decaying buffer used for LCD ghosting
    this.scratch = document.createElement('canvas');
    this.sctx = this.scratch.getContext('2d');
    this.sctx.imageSmoothingEnabled = false;
    this.persist = document.createElement('canvas');
    this.pctx = this.persist.getContext('2d');
    this.pctx.imageSmoothingEnabled = false;

    this._raf = null;
    this._loop = this._loop.bind(this);
    window.addEventListener('resize', () => this._resize());
    this._resize();
  }

  // active-profile parameter access (what the host's sliders read/write)
  get params() { return this.paramsByProfile[this.profile]; }
  set(key, value) { if (key in this.params) this.params[key] = value; }
  reset() { this.paramsByProfile[this.profile] = { ...PROFILE_DEFAULTS[this.profile] }; }

  setProfile(name) {
    if (!(name in FRAG) || name === this.profile) return;
    this.profile = name;
    this._clearPersist(); // don't smear one game's last frame into the next
  }

  setEnabled(on) {
    if (!this.ok) return;
    this.enabled = on;
    this.canvas.style.display = on ? 'block' : 'none';
    if (on) { this._resize(); this._clearPersist(); if (!this._raf) this._raf = requestAnimationFrame(this._loop); }
    else if (this._raf) { cancelAnimationFrame(this._raf); this._raf = null; }
  }

  _clearPersist() {
    if (!this.pctx) return;
    this.pctx.globalAlpha = 1;
    this.pctx.fillStyle = '#000';
    this.pctx.fillRect(0, 0, this.persist.width, this.persist.height);
  }

  _resize() {
    const ss = Math.min(2, window.devicePixelRatio || 1);
    const w = Math.max(2, Math.round(this.stage.clientWidth * ss));
    const h = Math.max(2, Math.round(this.stage.clientHeight * ss));
    this.canvas.width = w; this.canvas.height = h;
    this.scratch.width = w; this.scratch.height = h;
    this.persist.width = w; this.persist.height = h;
    this.sctx.imageSmoothingEnabled = false;
    this.pctx.imageSmoothingEnabled = false;
    this._clearPersist();
    this.gl.viewport(0, 0, w, h);
  }

  _loop(t) {
    this._render(t * 0.001);
    if (this.enabled) this._raf = requestAnimationFrame(this._loop);
  }

  _render(time) {
    const gl = this.gl, w = this.canvas.width, h = this.canvas.height;
    // composite the active viewer's visible canvas(es) into the scratch buffer.
    const sources = this.source();
    if (!sources.length) return;
    this.sctx.clearRect(0, 0, w, h);
    for (const c of sources) {
      if (!c.width || !c.height) continue;
      this.sctx.drawImage(c, 0, 0, w, h);
    }

    const p = this.params;
    // LCD ghosting: fade this frame over the persistent buffer (out = new*(1-g) + old*g).
    let upload = this.scratch;
    if (p.ghost > 0) {
      this.pctx.globalAlpha = 1 - p.ghost;
      this.pctx.drawImage(this.scratch, 0, 0, w, h);
      this.pctx.globalAlpha = 1;
      upload = this.persist;
    }
    gl.bindTexture(gl.TEXTURE_2D, this.tex);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, upload);

    const P = this.programs[this.profile];
    gl.useProgram(P.prog);
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buf);
    gl.enableVertexAttribArray(P.pos);
    gl.vertexAttribPointer(P.pos, 2, gl.FLOAT, false, 0, 0);

    if (this.profile === 'crt') {
      this._uploadCrt(P.u, w, h, time, p);
    } else {
      const g = this._cellGrid(w, p); // cell size + grid origin locked to the game's pixels
      if (this.profile === 'gb') this._uploadGb(P.u, w, h, g);
      else this._uploadGg(P.u, w, h, g);
    }

    gl.drawArrays(gl.TRIANGLES, 0, 6);
  }

  _uploadCrt(u, w, h, time, p) {
    const gl = this.gl;
    const sigH = p.signalLines, sigW = Math.round(sigH * (w / h));
    gl.uniform2f(u.u_resolution, w, h);
    gl.uniform2f(u.u_texResolution, sigW, sigH);
    gl.uniform1f(u.u_time, time);
    gl.uniform1f(u.u_noise, p.noise);
    gl.uniform1f(u.u_lumaSmear, p.lumaSmear);
    gl.uniform1f(u.u_iqBlur, p.iqBlur);
    gl.uniform1f(u.u_saturation, p.saturation);
    gl.uniform1f(u.u_phaseSpeed, p.phaseSpeed);
    gl.uniform1f(u.u_scanlineCount, p.scanLines);
    gl.uniform1f(u.u_verticalScale, 1.0);
    gl.uniform1f(u.u_beamFocus, p.beamFocus * 0.2);
    gl.uniform1f(u.u_beamBloom, p.beamBloom);
    gl.uniform1f(u.u_maskType, p.maskType);
    gl.uniform1f(u.u_maskStrength, p.maskStrength);
    gl.uniform1f(u.u_tvLines, p.tvLines);
    gl.uniform1f(u.u_curvature, p.curvature);
    gl.uniform1f(u.u_glow, p.glow);
  }

  // Size of one game pixel (and the grid origin) in overlay device px, from the active viewer's
  // camera. drawImage maps the whole viewer canvas onto the overlay, so app-screen px scale by
  // (overlay width / app-screen width); one game pixel is `zoom` app-screen px. Falls back to a
  // fixed cell when there's no map camera (shouldn't happen for the LCD profiles).
  _cellGrid(w, p) {
    const g = this.pixelGrid && this.pixelGrid();
    const mult = p.pixelsPerCell || 1;
    if (!g || !(g.screenW > 0) || !(g.zoom > 0)) return { cell: 6 * mult, ox: 0, oy: 0 };
    const ratio = w / g.screenW;
    return { cell: g.zoom * ratio * mult, ox: g.ox * ratio, oy: g.oy * ratio };
  }

  _uploadGb(u, w, h, g) {
    const gl = this.gl, p = this.params;
    gl.uniform2f(u.u_resolution, w, h);
    gl.uniform1f(u.u_cellSize, g.cell);
    gl.uniform2f(u.u_gridOrigin, g.ox, g.oy);
    gl.uniform1f(u.u_tint, p.tint);
    gl.uniform1f(u.u_gridStrength, p.gridStrength);
    gl.uniform1f(u.u_shadowOpacity, p.shadowOpacity);
    gl.uniform1f(u.u_shadowOffset, p.shadowOffset);
    gl.uniform1f(u.u_contrast, p.contrast);
  }

  _uploadGg(u, w, h, g) {
    const gl = this.gl, p = this.params;
    gl.uniform2f(u.u_resolution, w, h);
    gl.uniform1f(u.u_cellSize, g.cell);
    gl.uniform2f(u.u_gridOrigin, g.ox, g.oy);
    gl.uniform1f(u.u_gridStrength, p.gridStrength);
    gl.uniform1f(u.u_subpixel, p.subpixel);
    gl.uniform1f(u.u_saturation, p.saturation);
    gl.uniform1f(u.u_glow, p.glow);
  }
}

function compile(gl, type, src) {
  const sh = gl.createShader(type);
  gl.shaderSource(sh, src);
  gl.compileShader(sh);
  if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) console.error('screen shader:', gl.getShaderInfoLog(sh));
  return sh;
}
