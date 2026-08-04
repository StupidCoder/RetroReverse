// engine3d.js — shared three.js plumbing for the 3D views: the stage
// (renderer/scene/camera/controls/RAF), the fly camera with virtual sticks,
// and the ObjectLibrary that turns Retro-X object documents (model3d,
// billboard3d, wireframe3d) into instantiable three.js nodes.

import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import * as SkeletonUtils from 'three/addons/utils/SkeletonUtils.js';

export { THREE };

// ---- stage --------------------------------------------------------------------

export class Stage {
  constructor(el, cam = {}) {
    this.el = el;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.domElement.classList.add('fill');
    el.appendChild(renderer.domElement);

    this.scene = new THREE.Scene();
    const fov = cam.fov || 50;
    // Orthographic when the document asks for it, and by default for the flat
    // stages (camera.mode 'pan2d'): those are 3-D polygons the game photographs
    // straight on, and a perspective frustum both skews the level's far edges
    // and slides the parallax background against the terrain as you scroll —
    // neither of which a 2-D game does. The frustum height is the one the
    // document's fov + camera distance already framed, so the opening view is
    // unchanged; only the projection is.
    const wantsOrtho = cam.projection
      ? cam.projection === 'ortho'
      : cam.mode === 'pan2d' && !!cam.pos && !!cam.target;
    if (wantsOrtho) {
      const dist = new THREE.Vector3().fromArray(cam.pos || [0, 0, 10])
        .distanceTo(new THREE.Vector3().fromArray(cam.target || [0, 0, 0])) || 10;
      this.orthoHeight = 2 * dist * Math.tan((fov * Math.PI) / 360);
      this.camera = new THREE.OrthographicCamera(-1, 1, 1, -1, cam.near ?? 0.1, cam.far ?? 100000);
    } else {
      this.camera = new THREE.PerspectiveCamera(fov, 1, cam.near || 0.1, cam.far || 100000);
    }
    if (cam.pos) this.camera.position.fromArray(cam.pos);
    this.controls = new OrbitControls(this.camera, renderer.domElement);
    this.controls.enableDamping = true;
    this.controls.dampingFactor = 0.08;
    // Every knob a mounted view is allowed to turn, as it came out of the box.
    // A stage that outlives one view (the XR shell) has to be handed to the
    // next one clean, and restoring from a snapshot beats re-listing OrbitControls'
    // defaults here where they would rot the next time three changes one.
    this._ctrlDefaults = {
      enabled: this.controls.enabled,
      autoRotate: this.controls.autoRotate,
      autoRotateSpeed: this.controls.autoRotateSpeed,
      enableRotate: this.controls.enableRotate,
      screenSpacePanning: this.controls.screenSpacePanning,
      zoomToCursor: this.controls.zoomToCursor,
      minDistance: this.controls.minDistance,
      maxDistance: this.controls.maxDistance,
      minZoom: this.controls.minZoom,
      maxZoom: this.controls.maxZoom,
      mouseButtons: { ...this.controls.mouseButtons },
      touches: { ...this.controls.touches },
    };
    if (cam.target) this.controls.target.fromArray(cam.target);
    this.controls.update();

    this.renderer = renderer;
    // XR is armed on every stage but inert until a session starts (three only
    // consults renderer.xr when isPresenting). setAnimationLoop rather than a
    // hand-rolled rAF because three swaps the window's rAF for the XR
    // session's one on sessionstart and back on sessionend, which a
    // requestAnimationFrame loop cannot do — it would keep running alongside
    // the session and render every frame twice.
    renderer.xr.enabled = true;
    this.updaters = new Set();  // fn(dt, camPosWorld, elapsed)
    this.overrideRender = null; // cutscene player takes the frame over
    this.onXRFrame = null;      // ARSession reads the head pose from the XRFrame
    this.inputEnabled = true;   // false parks FlyCam/PanInput (an XR session drives the camera)
    this._clock = new THREE.Clock();
    // Cumulative CPU milliseconds, split at the one boundary that is actually
    // a boundary: what the scene's own updaters cost, and what handing the
    // scene to three costs. Never reset here — readers (xr.js PerfMeter) take
    // deltas, so two of them cannot rob each other. Only the SUBMISSION is
    // measured: the GPU's own time lands after this frame, and pretending
    // otherwise would be a bucket whose blind spot its own sum cannot show.
    this.perf = { upd: 0, draw: 0 };
    this._camWorld = new THREE.Vector3();
    this._tick = this._tick.bind(this);
    this._ro = new ResizeObserver(() => this._resize());
    this._ro.observe(el);
    this._resize();
    renderer.setAnimationLoop(this._tick);
    this.elapsed = 0;
  }

  get canvas() { return this.renderer.domElement; }

  _tick(time, frame) {
    // three's animation loop schedules the NEXT frame after the callback
    // returns, where the old hand-rolled rAF scheduled it first — so an
    // exception in a data-driven updater would now stop the viewer dead
    // instead of costing one frame. Keep the old failure mode.
    try {
      this._frame(frame);
    } catch (e) {
      console.error('stage frame', e);
    }
  }

  _frame(frame) {
    const xr = this.renderer.xr.isPresenting;
    const dt = Math.min(0.1, this._clock.getDelta());
    this.elapsed += dt;
    if (this.overrideRender && !xr) { this.overrideRender(dt); return; }
    // OrbitControls.update() has no `enabled` guard — only its event handlers
    // do — so it must not be CALLED in a session: it rewrites camera.position
    // from its own spherical every frame, which both fights the head pose and
    // leaves its internal state poisoned for the desktop view on exit.
    if (xr) this.onXRFrame?.(frame);
    else this.controls.update();
    // World position, not camera.position: under an XR rig the camera is a
    // child and its local position is the head pose in metres. Identical to
    // camera.position for the parentless desktop camera.
    const camPos = this.camera.getWorldPosition(this._camWorld);
    const t0 = performance.now();
    for (const u of this.updaters) u(dt, camPos, this.elapsed);
    const t1 = performance.now();
    this.renderer.render(this.scene, this.camera);
    this.perf.upd += t1 - t0;
    this.perf.draw += performance.now() - t1;
  }

  _resize() {
    // WebGLRenderer.setSize() warns and no-ops while presenting; the XR
    // manager owns the drawing buffer for the session's duration and restores
    // the old size itself on exit (we re-derive the aspect there).
    if (this.renderer.xr.isPresenting || this._disposed) return;
    const w = this.el.clientWidth, h = this.el.clientHeight;
    if (!w || !h) return;
    if (this.native) {
      // Native-resolution mode (screen filter active): render at the game's
      // vertical resolution — the width follows the viewport's aspect — and
      // upscale pixelated, so the CRT's scanlines land one per game line.
      const nh = this.native.h, nw = Math.max(2, Math.round((w / h) * nh));
      this.renderer.setPixelRatio(1);
      this.renderer.setSize(nw, nh, false);
      this.canvas.style.width = '100%';
      this.canvas.style.height = '100%';
      this.canvas.style.imageRendering = 'pixelated';
    } else {
      this.renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
      this.renderer.setSize(w, h, false);
      this.canvas.style.imageRendering = '';
    }
    this.setViewport(w / h);
  }

  // setViewport re-derives the projection for an aspect ratio. The ortho
  // camera keeps its frustum HEIGHT (OrbitControls dollies it by camera.zoom)
  // and takes its width from the viewport, exactly as the perspective one
  // keeps its fov.
  setViewport(aspect) {
    if (this.camera.isOrthographicCamera) {
      const hh = this.orthoHeight / 2;
      this.camera.left = -hh * aspect;
      this.camera.right = hh * aspect;
      this.camera.top = hh;
      this.camera.bottom = -hh;
    } else {
      this.camera.aspect = aspect;
    }
    this.camera.updateProjectionMatrix();
  }

  // setNative switches the render buffer to the game's native resolution
  // (screen-filter authenticity) or back to display resolution (null).
  setNative(size) {
    this.native = size && size.h ? size : null;
    this._resize();
  }

  // frame fits the camera to an object (used when the document has no pose).
  frame(obj) {
    this.frameBox(new THREE.Box3().setFromObject(obj));
  }

  // frameBox is the same, for callers that already measured — the XR shell
  // takes the view's own contentBox, which knows to leave the skybox out.
  frameBox(box) {
    if (!box || box.isEmpty()) return;
    const s = box.getBoundingSphere(new THREE.Sphere());
    const ortho = this.camera.isOrthographicCamera;
    const dist = ortho ? s.radius * 4 : (s.radius * 1.7) / Math.sin((this.camera.fov * Math.PI) / 360);
    this.camera.position.copy(s.center).add(new THREE.Vector3(0.55, 0.42, 1).normalize().multiplyScalar(dist));
    this.controls.target.copy(s.center);
    this.camera.near = ortho ? -s.radius * 10 : Math.max(0.001, s.radius / 100);
    this.camera.far = Math.max(this.camera.far, dist + s.radius * 100);
    if (ortho) this.orthoHeight = s.radius * 2.4;
    this.setViewport((this.el.clientWidth || 1) / (this.el.clientHeight || 1));
    this.controls.update();
  }

  // cameraSnapshot/cameraRestore bracket anything that drives the camera
  // itself (an XR session; a cutscene would qualify too). An XR session
  // OVERWRITES position/quaternion/scale/fov/zoom and both projection
  // matrices every frame from the head pose and never puts them back.
  cameraSnapshot() {
    const c = this.camera;
    return {
      position: c.position.clone(),
      quaternion: c.quaternion.clone(),
      up: c.up.clone(),
      near: c.near, far: c.far, zoom: c.zoom,
      fov: c.fov, aspect: c.aspect,
      ortho: c.isOrthographicCamera ? { l: c.left, r: c.right, t: c.top, b: c.bottom } : null,
      target: this.controls.target.clone(),
      controlsEnabled: this.controls.enabled,
    };
  }

  cameraRestore(s) {
    if (!s) return;
    const c = this.camera;
    c.position.copy(s.position);
    c.quaternion.copy(s.quaternion);
    c.up.copy(s.up);
    c.near = s.near; c.far = s.far; c.zoom = s.zoom;
    if (s.fov !== undefined) c.fov = s.fov;
    if (s.aspect !== undefined) c.aspect = s.aspect;
    if (s.ortho) { c.left = s.ortho.l; c.right = s.ortho.r; c.top = s.ortho.t; c.bottom = s.ortho.b; }
    c.updateProjectionMatrix();
    this.controls.target.copy(s.target);
    this.controls.enabled = s.controlsEnabled;
    this._resize();          // the XR manager restored the buffer size, not the aspect
    this.controls.update();
  }

  // resetContent hands a LIVING stage to a new mount. The desktop shell throws
  // the whole stage away between views, so nothing ever needed this; the XR
  // shell keeps one renderer across many contents, and everything a view is
  // allowed to leave behind has to go: the scene, the per-frame updaters, a
  // cutscene's render override, and any control settings it changed.
  //
  // `keep` is the shell's own updaters — the menu and the perf meter, which
  // belong to the stage rather than to whatever is currently in it.
  resetContent(keep = null) {
    disposeScene(this.scene);
    this.scene = new THREE.Scene();
    this.updaters.clear();
    if (keep) for (const u of keep) this.updaters.add(u);
    this.overrideRender = null;
    this.onXRFrame = null;
    Object.assign(this.controls, this._ctrlDefaults, {
      mouseButtons: { ...this._ctrlDefaults.mouseButtons },
      touches: { ...this._ctrlDefaults.touches },
    });
    this.controls.target.set(0, 0, 0);
    return this.scene;
  }

  dispose() {
    // A navigation mid-session must not leave the session alive with a dead
    // renderer behind it. The session's async teardown can land after this,
    // so _disposed parks the restore path it triggers.
    this._disposed = true;
    this.renderer.xr.getSession()?.end().catch(() => {});
    this.renderer.setAnimationLoop(null);
    this._ro.disconnect();
    this.renderer.dispose();
    this.renderer.domElement.remove();
  }
}

// ---- fly camera (ported from the old Studio's flycam) ---------------------------

const COARSE = matchMedia('(pointer: coarse)').matches;
const UP = new THREE.Vector3(0, 1, 0);
const ROT = 1.0;
const PITCH_MARGIN = 0.12;
const LOOK_SENS = 0.0035; // radians per dragged pixel
const DOLLY_STEP = 0.25;  // fraction of fly speed per wheel notch

export class FlyCam {
  // speed: world units per second (Retro-X camera.fly.speed)
  constructor(stage, speed) {
    this.stage = stage;
    this.speed = speed || 100;
    this.keys = new Set();
    this.shift = false;
    // Dragging looks around in place (grab-the-world, same rotation as the
    // arrow keys) and the wheel dollies along the view direction at a
    // constant speed. OrbitControls' own orbit/zoom/pan would fight both,
    // so they are disabled while the fly cam owns the stage.
    const controls = stage.controls;
    controls.enableRotate = controls.enableZoom = controls.enablePan = false;
    const cv = stage.canvas;
    this._ptrs = new Map(); // pointerId -> last position; look only single-pointer
    this._onPtrDown = (e) => {
      this._ptrs.set(e.pointerId, { x: e.clientX, y: e.clientY });
      cv.setPointerCapture(e.pointerId);
    };
    this._onPtrMove = (e) => {
      const p = this._ptrs.get(e.pointerId);
      if (!p) return;
      const dx = e.clientX - p.x, dy = e.clientY - p.y;
      p.x = e.clientX; p.y = e.clientY;
      if (this._ptrs.size === 1) this._look(dx, dy);
    };
    this._onPtrUp = (e) => this._ptrs.delete(e.pointerId);
    this._onWheel = (e) => {
      e.preventDefault();
      if (stage.inputEnabled === false) return;
      const cam = stage.camera, target = controls.target;
      const fwd = target.clone().sub(cam.position).normalize();
      const step = -(e.deltaY / 100) * this.speed * DOLLY_STEP * (this.shift ? 2.5 : 1);
      cam.position.addScaledVector(fwd, step);
      target.addScaledVector(fwd, step);
    };
    cv.addEventListener('pointerdown', this._onPtrDown);
    cv.addEventListener('pointermove', this._onPtrMove);
    cv.addEventListener('pointerup', this._onPtrUp);
    cv.addEventListener('pointercancel', this._onPtrUp);
    cv.addEventListener('wheel', this._onWheel, { passive: false });
    this._onKeyDown = (e) => {
      this.shift = e.shiftKey;
      if (e.altKey || e.ctrlKey || e.metaKey || stage.inputEnabled === false) return;
      if (KEYS.has(e.code)) { this.keys.add(e.code); e.preventDefault(); }
    };
    this._onKeyUp = (e) => { this.shift = e.shiftKey; this.keys.delete(e.code); };
    this._onBlur = () => this.keys.clear();
    window.addEventListener('keydown', this._onKeyDown);
    window.addEventListener('keyup', this._onKeyUp);
    window.addEventListener('blur', this._onBlur);
    this.sticks = null;
    if (COARSE) this._buildSticks(stage.el);
    this._update = (dt) => this.update(dt);
    stage.updaters.add(this._update);
  }

  // _look rotates the camera in place: dx/dy in pixels, first-person
  // convention (drag right = look right, drag down = look down).
  _look(dx, dy) {
    if ((!dx && !dy) || this.stage.inputEnabled === false) return;
    const cam = this.stage.camera, target = this.stage.controls.target;
    const off = target.clone().sub(cam.position);
    if (dx) off.applyAxisAngle(UP, -dx * LOOK_SENS);
    if (dy) {
      const right = off.clone().cross(UP).normalize();
      const rot = off.clone().applyAxisAngle(right, -dy * LOOK_SENS);
      const a = rot.angleTo(UP);
      if (a > PITCH_MARGIN && a < Math.PI - PITCH_MARGIN) off.copy(rot);
    }
    target.copy(cam.position).add(off);
  }

  update(dt) {
    // While an XR session owns the camera the fly controls are parked: their
    // writes to camera.position/controls.target would be silently overwritten
    // by the head pose and would corrupt the pose restored on exit.
    if (this.stage.inputEnabled === false) { this.keys.clear(); return; }
    const k = this.keys;
    const l = this.sticks?.l || ZERO, r = this.sticks?.r || ZERO;
    const mf = (k.has('KeyW') ? 1 : 0) - (k.has('KeyS') ? 1 : 0) - l.y;
    const ms = (k.has('KeyD') ? 1 : 0) - (k.has('KeyA') ? 1 : 0) + l.x;
    const ry = (k.has('ArrowLeft') ? 1 : 0) - (k.has('ArrowRight') ? 1 : 0) - r.x;
    const rp = (k.has('ArrowUp') ? 1 : 0) - (k.has('ArrowDown') ? 1 : 0) - r.y;
    if (!mf && !ms && !ry && !rp) return;
    const cam = this.stage.camera, target = this.stage.controls.target;
    const off = target.clone().sub(cam.position);
    if (ry) off.applyAxisAngle(UP, ry * ROT * dt);
    if (rp) {
      const right = off.clone().cross(UP).normalize();
      const rot = off.clone().applyAxisAngle(right, rp * ROT * dt);
      const a = rot.angleTo(UP);
      if (a > PITCH_MARGIN && a < Math.PI - PITCH_MARGIN) off.copy(rot);
    }
    target.copy(cam.position).add(off);
    if (mf || ms) {
      const v = this.speed * dt * (this.shift ? 2.5 : 1);
      const fwd = off.clone().normalize();
      const right = fwd.clone().cross(UP).normalize();
      cam.position.addScaledVector(fwd, mf * v).addScaledVector(right, ms * v);
      target.addScaledVector(fwd, mf * v).addScaledVector(right, ms * v);
    }
  }

  _buildSticks(el) {
    if (getComputedStyle(el).position === 'static') el.style.position = 'relative';
    const mk = (side) => {
      const base = document.createElement('div');
      base.style.cssText = `position:absolute;bottom:76px;${side}:22px;width:112px;height:112px;border-radius:50%;` +
        'background:rgba(110,130,160,.16);border:1px solid rgba(150,175,210,.35);z-index:11;' +
        'touch-action:none;-webkit-tap-highlight-color:transparent;user-select:none;';
      const knob = document.createElement('div');
      knob.style.cssText = 'position:absolute;left:50%;top:50%;width:46px;height:46px;border-radius:50%;' +
        'background:rgba(170,195,230,.45);border:1px solid rgba(200,220,245,.5);transform:translate(-50%,-50%);pointer-events:none;';
      base.appendChild(knob);
      el.appendChild(base);
      const s = { base, knob, id: null, x: 0, y: 0 };
      const track = (e) => {
        const rect = base.getBoundingClientRect(), R = rect.width / 2;
        let x = (e.clientX - rect.left - R) / R, y = (e.clientY - rect.top - R) / R;
        const len = Math.hypot(x, y);
        if (len > 1) { x /= len; y /= len; }
        s.x = x; s.y = y;
        knob.style.left = `${50 + x * 32}%`;
        knob.style.top = `${50 + y * 32}%`;
      };
      base.addEventListener('pointerdown', (e) => { s.id = e.pointerId; base.setPointerCapture(e.pointerId); track(e); });
      base.addEventListener('pointermove', (e) => { if (e.pointerId === s.id) track(e); });
      const end = (e) => {
        if (e.pointerId !== s.id) return;
        s.id = null; s.x = 0; s.y = 0;
        knob.style.left = '50%'; knob.style.top = '50%';
      };
      base.addEventListener('pointerup', end);
      base.addEventListener('pointercancel', end);
      return s;
    };
    this.sticks = { l: mk('left'), r: mk('right') };
  }

  dispose() {
    window.removeEventListener('keydown', this._onKeyDown);
    window.removeEventListener('keyup', this._onKeyUp);
    window.removeEventListener('blur', this._onBlur);
    const cv = this.stage.canvas;
    cv.removeEventListener('pointerdown', this._onPtrDown);
    cv.removeEventListener('pointermove', this._onPtrMove);
    cv.removeEventListener('pointerup', this._onPtrUp);
    cv.removeEventListener('pointercancel', this._onPtrUp);
    cv.removeEventListener('wheel', this._onWheel);
    this.stage.updaters.delete(this._update);
    if (this.sticks) { this.sticks.l.base.remove(); this.sticks.r.base.remove(); }
  }
}

const KEYS = new Set(['KeyW', 'KeyA', 'KeyS', 'KeyD', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown']);
const ZERO = { x: 0, y: 0 };

export const flyHint = COARSE ? 'sticks: left move · right look' : 'WASD move · drag or arrows look · wheel dolly';

// ---- GLB + object instantiation ---------------------------------------------------

const gltfLoader = new GLTFLoader();

function loadImage(url) {
  return new Promise((res, rej) => {
    const i = new Image();
    i.onload = () => res(i);
    i.onerror = () => rej(new Error(url));
    i.src = url;
  });
}

// loadGLB fetches and parses a GLB. The optional AbortSignal cancels the
// network transfer mid-flight — three's own loaders cannot be aborted, so the
// fetch is ours and only the parse is handed to GLTFLoader. Level views pass
// their mount's signal so navigating away stops the stream (the mansion kept
// downloading all 75 room shells under the next scene otherwise).
export async function loadGLB(url, signal) {
  const res = await fetch(url, { signal });
  if (!res.ok) throw new Error(`${url}: HTTP ${res.status}`);
  const buf = await res.arrayBuffer();
  return gltfLoader.parseAsync(buf, THREE.LoaderUtils.extractUrlBase(url));
}

// applyTexFilter imposes the platform's texture sampling (manifest
// display.texFilter) on a loaded subtree: "linear" = bilinear + mipmaps
// (N64), "bilinear" = bilinear with NO mipmaps — for textures the game
// ships without mip chains (Luigi's Mansion's .mdl/.bin images: hardware
// always samples mip 0 there, and generated mipmaps banded the flashlight
// cone's 8x128 gradient into rings) — "nearest"/default = point-sampled
// magnification (PSX-style).
export function applyTexFilter(root, mode) {
  // No mode means the manifest did not state one, which means "however the
  // asset was authored" — the glTF sampler already says. It must NOT mean
  // nearest: that silently blocked Jak & Daxter's textures, whose GLBs all
  // declare LINEAR, because this path (unlike level3d's) called in
  // unconditionally and the fallback below is nearest.
  if (!mode) return;
  root.traverse((o) => {
    const m = o.material;
    if (!m) return;
    for (const mat of Array.isArray(m) ? m : [m]) {
      if (!mat.map) continue;
      filterTex(mat.map, mode);
      mat.needsUpdate = true;
    }
  });
}

function filterTex(tex, mode) {
  if (mode === 'linear') {
    tex.magFilter = THREE.LinearFilter;
    tex.minFilter = THREE.LinearMipmapLinearFilter;
    tex.generateMipmaps = true;
  } else if (mode === 'bilinear') {
    tex.magFilter = THREE.LinearFilter;
    tex.minFilter = THREE.LinearFilter;
    tex.generateMipmaps = false;
  } else {
    tex.magFilter = THREE.NearestFilter;
  }
  tex.needsUpdate = true;
}

const LOOPMAP = { loop: THREE.LoopRepeat, pingpong: THREE.LoopPingPong, once: THREE.LoopOnce, hold: THREE.LoopOnce };

// billboardCell picks which cell of a sprite sheet a billboard shows: the row
// from the camera's bearing (a Doom-style directional sprite's eight views),
// the column from its animation clock. dx/dz point from the sprite to the
// eye. Shared with billboards.js, which draws the same sprites merged — two
// spellings of this rule would be two sprites disagreeing about which way a
// creature faces.
export function billboardCell(doc, dx, dz, t) {
  const views = doc.views || 1;
  const anim = (doc.animations || [])[0] || { col: 0, framesPerView: 1, fps: 0 };
  let row = 0;
  if (views > 1) {
    const ang = Math.atan2(dx, dz) - (doc.heading || 0);
    const step = (2 * Math.PI) / views;
    row = Math.round(ang / step) % views;
    if (row < 0) row += views;
  }
  const per = anim.framesPerView || 1;
  const fps = anim.fps || 0;
  const frame = fps > 0 ? Math.floor(t * fps) % per : 0;
  return { row, col: (anim.col || 0) + frame };
}

// disposeScene frees the GPU resources reachable from a subtree.
//
// Nothing used to call anything like this: every 3-D view owned its whole
// renderer, and Stage.dispose() -> renderer.dispose() dropped the context with
// everything in it. A stage that OUTLIVES its content has no such backstop, so
// each swap has to give the memory back itself or a long session climbs until
// it dies.
//
// Two things it has to get right. Geometries and materials are SHARED between a
// proto and every clone of it (SkeletonUtils.clone and Object3D.clone(true) copy
// references, not data), so each must be disposed exactly once — hence the sets.
// And a material owns more textures than `map`: envMap, alphaMap, emissiveMap
// and friends are all real uploads, so anything on the material that looks like
// a texture goes too.
export function disposeScene(root) {
  if (!root) return { geometries: 0, materials: 0, textures: 0 };
  const geo = new Set(), mat = new Set(), tex = new Set(), skel = new Set();
  root.traverse((o) => {
    if (o.geometry) geo.add(o.geometry);
    for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) mat.add(m);
    // A skeleton owns a boneTexture — a real GPU upload that NOTHING else here
    // can see: it hangs off the mesh, not off a material. SkeletonUtils.clone
    // gives every clone its own skeleton (that is the point of it), so an
    // animated cast leaks one bone texture per skinned mesh per load. Luigi's
    // Mansion's opening, twenty skinned actors, leaked 258 textures a time
    // while its static rooms leaked none.
    if (o.isSkinnedMesh && o.skeleton) skel.add(o.skeleton);
  });
  for (const m of mat) {
    for (const v of Object.values(m)) if (v && v.isTexture) tex.add(v);
  }
  for (const g of geo) g.dispose();
  for (const m of mat) m.dispose();
  for (const t of tex) t.dispose();
  for (const s of skel) s.dispose();
  // Detach so a stale reference cannot resurrect a disposed subtree into a
  // live scene — three would silently re-upload it and the leak would be back.
  root.removeFromParent?.();
  root.clear?.();
  return { geometries: geo.size, materials: mat.size, textures: tex.size, skeletons: skel.size };
}

// ObjectLibrary caches loaded object documents + payloads per game and
// instantiates them for placements.
export class ObjectLibrary {
  // signal (optional AbortSignal) cancels in-flight model downloads; the
  // library lives and dies with one mounted view, so an aborted proto never
  // poisons a later mount's cache.
  constructor(game, signal) {
    this.game = game;
    this.signal = signal;
    this.cache = new Map(); // object id -> Promise<proto>
  }

  proto(id) {
    if (!this.cache.has(id)) this.cache.set(id, this._load(id));
    return this.cache.get(id);
  }

  // dispose frees every proto this library loaded. The scene walk cannot do it:
  // a proto's glTF scenes are not in anybody's scene graph, and the flip-book
  // frames and env cubes it holds are only reachable through here — at most one
  // flip-book frame is ever assigned to a material at a time, so the other
  // frames would be invisible to any amount of traversal.
  async dispose() {
    const protos = await Promise.all([...this.cache.values()].map((p) => p.catch(() => null)));
    this.cache.clear();
    for (const p of protos) {
      if (!p) continue;
      for (const sc of p.gltf?.scenes?.length ? p.gltf.scenes : [p.gltf?.scene]) disposeScene(sc);
      for (const frames of Object.values(p.flipTex || {})) for (const t of frames) t.dispose();
      p.tex?.dispose();
    }
  }

  async _load(id) {
    const { asset, doc, docPath } = await this.game.object(id);
    const proto = { asset, doc, docPath };
    const texFilter = this.game.display.texFilter;
    if (doc.type === 'model3d') {
      proto.gltf = await loadGLB(this.game.url(docPath, doc.model), this.signal);
      for (const sc of proto.gltf.scenes?.length ? proto.gltf.scenes : [proto.gltf.scene]) {
        applyTexFilter(sc, texFilter);
      }
      // Retro-X material extras: additive blending, and env-cube sheen fed by
      // the object's static cube faces (doc.envMap, +x,-x,+y,-y,+z,-z).
      let cube = null;
      if (doc.envMap?.length === 6) {
        const imgs = await Promise.all(doc.envMap.map((f) => loadImage(this.game.url(docPath, f))));
        cube = new THREE.CubeTexture(imgs);
        cube.colorSpace = THREE.SRGBColorSpace;
        cube.needsUpdate = true;
      }
      const mats = new Set();
      for (const sc of proto.gltf.scenes?.length ? proto.gltf.scenes : [proto.gltf.scene]) {
        sc.traverse((o) => {
          for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) {
            mats.add(m);
            // Submission order carries the depth-tie semantics (GS GEQUAL):
            // renderOrder keeps it under three's opaque sort; ties then
            // resolve later-wins via the default LessEqualDepth.
            if (Number.isFinite(m.userData?.layer)) o.renderOrder = m.userData.layer;
          }
        });
      }
      for (const m of mats) {
        if (m.userData?.blend === 'additive') {
          m.transparent = true;
          m.blending = THREE.AdditiveBlending;
          m.depthWrite = false;
          m.needsUpdate = true;
        }
        if (m.userData?.blend === 'alpha') {
          // The source hardware's translucent pass: alpha blend with depth
          // writes off (Luigi's Mansion draws its flashlight cone this way).
          m.transparent = true;
          m.depthWrite = false;
          m.needsUpdate = true;
        }
        if (cube && m.userData?.sheen) {
          // The game's combiner ADDS the reflection; the fixed 0.3 amount is
          // a declared approximation (the exact scale lives in the material
          // state setup).
          m.envMap = cube;
          m.combine = THREE.AddOperation;
          m.reflectivity = 0.3;
          m.needsUpdate = true;
        }
      }
      if (doc.flipbooks?.length) {
        proto.flipTex = {};
        for (const fb of doc.flipbooks) {
          proto.flipTex[fb.material] = await Promise.all(fb.textures.map((t) =>
            new THREE.TextureLoader().loadAsync(this.game.url(docPath, t)).then((tx) => {
              filterTex(tx, texFilter);
              tx.flipY = false;
              return tx;
            })));
        }
      }
    } else if (doc.type === 'billboard3d') {
      proto.tex = await new THREE.TextureLoader().loadAsync(this.game.url(docPath, doc.atlas.file));
      // Billboards are sprite sheets: sub-rect sampling bleeds across cells
      // with mipmaps, so they stay point-sampled regardless of texFilter.
      proto.tex.magFilter = proto.tex.minFilter = THREE.NearestFilter;
      proto.tex.generateMipmaps = false;
    }
    return proto;
  }

  // instance builds a fresh node. Returns:
  //   { node, doc, asset, mixer?, actions?, update?(dt, camPos, elapsed) }
  // opts.scene picks a named glTF scene for model3d variants (default scene 0).
  async instance(id, opts) {
    const proto = await this.proto(id);
    const { doc } = proto;
    if (doc.type === 'model3d') return this._model(proto, opts?.scene);
    if (doc.type === 'billboard3d') return this._billboard(proto);
    if (doc.type === 'wireframe3d') return this._wireframe(proto);
    throw new Error(`object ${id}: unhandled type ${doc.type}`);
  }

  _model(proto, sceneName) {
    const { doc, gltf } = proto;
    let src = gltf.scene;
    if (sceneName && gltf.scenes) src = gltf.scenes.find((s) => s.name === sceneName) || src;
    const node = doc.skinnedClone ? SkeletonUtils.clone(src) : src.clone(true);
    const inst = { node, doc, asset: proto.asset, clips: gltf.animations || [] };
    if (inst.clips.length) {
      inst.mixer = new THREE.AnimationMixer(node);
      inst.playAnim = (animId, over = {}) => {
        const meta = (doc.animations || []).find((a) => a.id === animId) || (doc.animations || [])[0];
        const clip = inst.clips.find((c) => c.name === (meta?.clip || animId)) || inst.clips[0];
        if (!clip) return null;
        inst.mixer.stopAllAction();
        const action = inst.mixer.clipAction(clip);
        const loop = over.loop || meta?.loop || 'loop';
        action.setLoop(LOOPMAP[loop] ?? THREE.LoopRepeat, Infinity);
        if (loop === 'hold' || loop === 'once') action.clampWhenFinished = true;
        action.play();
        return { action, clip, meta };
      };
    }
    const updates = [];
    // material UV animation + flipbooks, replayed on the game's frame clock
    if (doc.uvAnims?.length || doc.flipbooks?.length) {
      const mats = {};
      node.traverse((o) => {
        for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) {
          (mats[m.name] ||= []).push(m);
        }
      });
      const tickHz = this.game.display.tickHz || 60;
      let t = 0;
      const chan = (c, f) => {
        if (!c) return 0;
        if (c.const !== undefined) return c.const;
        const i = f / (c.step || 1);
        const i0 = Math.floor(i) % c.samples.length;
        const i1 = (i0 + 1) % c.samples.length;
        return c.samples[i0] + (c.samples[i1] - c.samples[i0]) * (i - Math.floor(i));
      };
      updates.push((dt) => {
        t += dt * tickHz;
        for (const ua of doc.uvAnims || []) {
          const f = t % (ua.frames || 1);
          for (const m of mats[ua.material] || []) {
            if (!m.map) continue;
            m.map.wrapS = m.map.wrapT = THREE.RepeatWrapping;
            m.map.offset.set(chan(ua.transS, f), chan(ua.transT, f));
            m.map.rotation = chan(ua.rot, f);
            const ss = ua.scaleS ? chan(ua.scaleS, f) : 1, st = ua.scaleT ? chan(ua.scaleT, f) : 1;
            if (ss !== 0 && st !== 0) m.map.repeat.set(ss || 1, st || 1);
          }
        }
        for (const fb of doc.flipbooks || []) {
          const seq = fb.sequence;
          const idx = seq[((t / (fb.step || 1)) | 0) % seq.length];
          const tex = proto.flipTex?.[fb.material]?.[idx];
          for (const m of mats[fb.material] || []) {
            if (tex && m.map !== tex) { m.map = tex; m.needsUpdate = true; }
          }
        }
      });
    }
    if (inst.mixer) updates.push((dt) => inst.mixer.update(dt));
    // Billboards go LAST, after the mixer. The parts that turn are skeleton
    // bones, and a clip writes every bone each frame — the split billboard
    // leaves carry identity tracks precisely so the viewer supplies their
    // orientation, which it cannot do if the mixer runs afterwards.
    if (doc.billboard) {
      // The game marks camera-facing BONES, not whole models, and the GLB
      // carries that per node (extras.billboard, from the .bmd flag word).
      // So a character turns only its marked part — the bob-omb's body_bill
      // swings while its feet stay put — and a one-bone sprite, which has no
      // marked node to find, turns whole.
      const parts = [];
      node.traverse((o) => { if (o.userData?.billboard) parts.push(o); });
      if (!parts.length) parts.push(node);
      const full = doc.billboard === 'camera';
      updates.push((dt, camPos) => {
        for (const p of parts) {
          if (full) {
            p.lookAt(camPos); // a sphere has no upright to keep
          } else {
            const wp = p.getWorldPosition(_wp);
            p.rotation.y = Math.atan2(camPos.x - wp.x, camPos.z - wp.z);
          }
        }
      });
    }
    if (updates.length) inst.update = (dt, camPos, t) => { for (const u of updates) u(dt, camPos, t); };
    return inst;
  }

  _billboard(proto) {
    const { doc } = proto;
    const [w, h] = doc.size || [1, 1];
    const map = proto.tex.clone();
    map.needsUpdate = true;
    const blend = doc.blend || 'opaque';
    // forceSinglePass: a transparent DoubleSide mesh is otherwise drawn TWICE
    // — three renders back faces then front faces so a self-overlapping shape
    // composites in order. A flat quad has one face, so the second pass draws
    // the same pixels again for nothing, at double the draw calls.
    const matOpts = { map, side: THREE.DoubleSide, forceSinglePass: true, toneMapped: false, alphaTest: 0.5 };
    if (blend === 'additive') Object.assign(matOpts, { transparent: true, blending: THREE.AdditiveBlending, depthWrite: false, alphaTest: 0 });
    else if (blend === 'alpha') Object.assign(matOpts, { transparent: true, depthWrite: false });
    const geo = new THREE.PlaneGeometry(w, h);
    if ((doc.anchorMode || 'center') === 'bottom') geo.translate(0, h / 2, 0);
    const node = new THREE.Mesh(geo, new THREE.MeshBasicMaterial(matOpts));
    node.frustumCulled = false;

    const img = () => map.image;
    const cw = doc.atlas.cellW, ch = doc.atlas.cellH;
    const anim = (doc.animations || [])[0] || { col: 0, framesPerView: 1, fps: 0 };
    let last = -1;
    const setCell = (row, col) => {
      const im = img();
      if (!im?.width) return;
      const k = row * 1000 + col;
      if (k === last) return;
      last = k;
      map.repeat.set(cw / im.width, ch / im.height);
      map.offset.set((col * cw) / im.width, 1 - ((row + 1) * ch) / im.height);
    };
    setCell(0, anim.col || 0);
    const inst = { node, doc, asset: proto.asset };
    inst.update = (dt, camPos, t) => {
      const dx = camPos.x - node.getWorldPosition(_wp).x;
      const dz = camPos.z - _wp.z;
      const { row, col } = billboardCell(doc, dx, dz, t);
      setCell(row, col);
      if ((doc.mode || 'camera') === 'camera') node.lookAt(camPos);
      else node.rotation.set(0, Math.atan2(dx, dz), 0);
    };
    return inst;
  }

  _wireframe(proto) {
    const { doc } = proto;
    const wf = doc.wireframe;
    const nEdges = wf.edges.length;
    const positions = new Float32Array(nEdges * 6);
    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    const node = new THREE.LineSegments(geo, new THREE.LineBasicMaterial({ color: 0xd8e6ff }));
    node.frustumCulled = false;
    const inst = { node, doc, asset: proto.asset };
    const v = wf.positions;
    // Fill every edge once up front: the camera frames the geometry before
    // the first visibility pass runs, and an empty buffer frames to nothing.
    {
      let n = 0;
      for (const [a, b] of wf.edges) {
        positions[n * 6 + 0] = v[a * 3]; positions[n * 6 + 1] = v[a * 3 + 1]; positions[n * 6 + 2] = v[a * 3 + 2];
        positions[n * 6 + 3] = v[b * 3]; positions[n * 6 + 4] = v[b * 3 + 1]; positions[n * 6 + 5] = v[b * 3 + 2];
        n++;
      }
      geo.setDrawRange(0, n * 2);
      geo.computeBoundingSphere();
    }
    // The model's own hidden-surface rule: a face is visible when its outward
    // normal points at the eye; an edge draws when either adjacent face does.
    inst.update = (dt, camPos) => {
      node.updateWorldMatrix(true, false);
      const inv = _m4.copy(node.matrixWorld).invert();
      const eye = _wp.copy(camPos).applyMatrix4(inv);
      let n = 0;
      for (const [a, b, fa, fb] of wf.edges) {
        const visA = faceVisible(wf, fa, eye);
        const visB = faceVisible(wf, fb, eye);
        if (!visA && !visB) continue;
        positions[n * 6 + 0] = v[a * 3]; positions[n * 6 + 1] = v[a * 3 + 1]; positions[n * 6 + 2] = v[a * 3 + 2];
        positions[n * 6 + 3] = v[b * 3]; positions[n * 6 + 4] = v[b * 3 + 1]; positions[n * 6 + 5] = v[b * 3 + 2];
        n++;
      }
      geo.setDrawRange(0, n * 2);
      geo.attributes.position.needsUpdate = true;
    };
    return inst;
  }
}

function faceVisible(wf, fi, eye) {
  const n = wf.faces[fi], c = wf.faceCenters[fi];
  if (!n || !c) return true;
  return n[0] * (eye.x - c[0]) + n[1] * (eye.y - c[1]) + n[2] * (eye.z - c[2]) > 0;
}

const _wp = new THREE.Vector3();
const _m4 = new THREE.Matrix4();

// Wireframe mode draws every material in one constant bright colour — wire
// lines rendered in a dark model's own texture (NFS's black Diablo) are
// unreadable. The material keeps its depth/blend/side settings (the sky dome
// relies on depthTest:false), only its texture and tint are stashed away
// while the toggle is on. The colour matches the part-isolation ghost.
export const WIREFRAME_COLOR = 0x44608a;

export function wireMaterial(mat, on) {
  if (!('wireframe' in mat)) return;
  mat.wireframe = on;
  if (!mat.color?.isColor) return; // shader materials: flag only
  if (on) {
    if (!mat.userData.wfSave) {
      mat.userData.wfSave = { map: mat.map ?? null, color: mat.color.getHex(), vertexColors: mat.vertexColors };
      mat.map = null;
      mat.color.setHex(WIREFRAME_COLOR);
      mat.vertexColors = false;
      mat.needsUpdate = true;
    }
  } else if (mat.userData.wfSave) {
    const s = mat.userData.wfSave;
    mat.map = s.map;
    mat.color.setHex(s.color);
    mat.vertexColors = s.vertexColors;
    delete mat.userData.wfSave;
    mat.needsUpdate = true;
  }
}

export function applyWireframe(root, on) {
  root.traverse((o) => {
    const m = o.material;
    if (!m) return;
    for (const mat of Array.isArray(m) ? m : [m]) wireMaterial(mat, on);
  });
}

// applyTransform puts a Retro-X placement transform onto a node.
export function applyTransform(node, pl) {
  if (pl.matrix?.length === 16) {
    node.matrixAutoUpdate = false;
    node.matrix.fromArray(pl.matrix);
    return;
  }
  if (pl.pos) node.position.fromArray(pl.pos.length === 2 ? [...pl.pos, 0] : pl.pos);
  if (pl.rot) node.rotation.set(pl.rot[0], pl.rot[1], pl.rot[2], 'XYZ');
  if (pl.scale != null) {
    const s = Array.isArray(pl.scale) ? pl.scale : [pl.scale, pl.scale, pl.scale];
    node.scale.set(s[0] ?? 1, s.length > 1 ? s[1] : s[0], s.length > 2 ? s[2] : s[0]);
  }
}
