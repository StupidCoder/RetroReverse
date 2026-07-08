// Stage3D — the minimal, generic three.js plumbing shared by every 3-D viewer plugin.
// It owns a WebGLRenderer, a Scene, a PerspectiveCamera, OrbitControls, the canvas, resize
// handling and a single RAF loop; a plugin fills the scene and (optionally) takes over the
// frame for its own post-processing. Nothing here is game-specific — the visuals live in the
// plugins (see shared/renderers.js and the per-game renderer modules).
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

// A pleasant 3/4 viewing direction, used by frame() to place the camera — the same direction
// the Elite viewer opened on, so a fitted object reads in three-quarter view.
const VIEW_DIR = new THREE.Vector3(0.55, 0.42, 1).normalize();

export class Stage3D {
  constructor(el, opts = {}) {
    this.el = el;
    const { background = 0x000000, fov = 45, near = 0.1, far = 200000 } = opts;

    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    if (background != null) scene.background = new THREE.Color(background);

    const camera = new THREE.PerspectiveCamera(fov, 1, near, far);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;

    this._renderer = renderer;
    this._scene = scene;
    this._camera = camera;
    this._controls = controls;

    // Objects added since the last clear() — tracked so clear() can remove exactly them
    // (leaving anything a plugin parked directly on the scene alone).
    this._added = [];

    // Plugin hooks (reset between items by the Viewer). onFrame runs every frame before the
    // render step; when `render` is set the loop calls it INSTEAD of the default renderer.render
    // (so a plugin can own the frame for post-FX). `pixelGrid`/`hud`/`disposePlugin` let a plugin
    // publish its native pixel grid (for the global CRT filter), a HUD detail string, and a teardown.
    this.onFrame = null;
    this.render = null;
    this.pixelGrid = null;
    this.hud = null;
    this.disposePlugin = null;

    this.active = true;   // the Studio pauses hidden viewers by flipping this
    this._raf = 0;
    this._last = 0;
    this._tick = this._tick.bind(this);

    this._resize();
    this._ro = new ResizeObserver(() => this._resize());
    this._ro.observe(el);
    this.start();
  }

  get canvas() { return this._renderer.domElement; }
  get scene() { return this._scene; }
  get camera() { return this._camera; }
  get renderer() { return this._renderer; }
  get controls() { return this._controls; }

  // Add an object and remember it so clear() can take it back out again.
  add(obj) { this._scene.add(obj); this._added.push(obj); return obj; }

  // Remove everything added since the last clear() (does not dispose — a plugin owns the
  // lifetime of its GPU resources and frees them in its disposePlugin).
  clear() {
    for (const obj of this._added) this._scene.remove(obj);
    this._added = [];
  }

  // Fit the camera/controls to an object's bounding sphere from the shared 3/4 direction.
  // near/far are only ever widened, never shrunk, so a plugin's distant backdrop (e.g. a
  // far starfield) stays inside the frustum after a fit.
  frame(obj) {
    const box = new THREE.Box3().setFromObject(obj);
    if (box.isEmpty()) return;
    const sphere = box.getBoundingSphere(new THREE.Sphere());
    const r = sphere.radius || 1;
    const c = sphere.center;
    const dist = (r * 1.6) / Math.sin((this._camera.fov * Math.PI) / 360);
    this._camera.position.copy(c).addScaledVector(VIEW_DIR, dist);
    this._controls.target.copy(c);
    this._camera.near = Math.min(this._camera.near, Math.max(0.001, r / 100));
    this._camera.far = Math.max(this._camera.far, dist + r * 100);
    this._camera.updateProjectionMatrix();
    this._controls.update();
  }

  _tick(now) {
    this._raf = requestAnimationFrame(this._tick);
    if (!this.active) return;
    const dt = this._last ? Math.min(0.1, (now - this._last) / 1000) : 1 / 60;
    this._last = now;
    this._controls.update();
    if (this.onFrame) this.onFrame(this._camera.position, dt);
    if (this.render) this.render(this);
    else this._renderer.render(this._scene, this._camera);
  }

  _resize() {
    const w = this.el.clientWidth, h = this.el.clientHeight || Math.round(w * 0.62);
    if (!w) return;
    this._renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    this._renderer.setSize(w, h, false);
    this._camera.aspect = w / h;
    this._camera.updateProjectionMatrix();
  }

  start() {
    if (this._raf == null || this._raf === 0) { this._last = 0; this._raf = requestAnimationFrame(this._tick); }
  }

  stop() {
    if (this._raf) { cancelAnimationFrame(this._raf); this._raf = 0; }
  }

  // The canvas is appended to `el` in the constructor; mount()/unmount() manage that DOM link
  // (the Studio keeps a viewer mounted and just hides it, so these are mostly for completeness).
  mount() {
    if (this._renderer.domElement.parentNode !== this.el) this.el.appendChild(this._renderer.domElement);
  }

  unmount() {
    const dom = this._renderer.domElement;
    if (dom.parentNode) dom.parentNode.removeChild(dom);
  }
}
