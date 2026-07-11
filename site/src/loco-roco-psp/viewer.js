// LocoRoco — stage viewer. The stages are the game's own level geometry
// (Part V §6): GE triangle strips decoded from the st_*.clv files and exported
// as textured GLBs, a foreground terrain model and a background flora model
// per stage. The game is 2-D — the "3-D" is layered planes at fixed z — so the
// camera is the same map camera as the 2-D tilemap games (drag/pinch/wheel,
// arrow keys + '+'/'-' via the Studio's KeyboardCamera): an orthographic
// three.js camera driven by a shared MapCamera over a shim "world".
//
// Draw order mirrors the engine's painter algorithm: the GE draws the stage
// back-to-front with depth writes off, so every material renders transparent
// with depth testing disabled and meshes sort by their geometry's z (the
// stage data keeps its layers at distinct z planes; the terrain is frontmost).
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { MapCamera } from '../shared/camera.js';

const BASE = 'public/loco-roco-psp/';

// The PSP's screen in world units: at the game's default camera the stage's
// world units map ~1:1 to screen pixels, so this rect is the "native view"
// the camera frames by default (the same convention as the tilemap games).
const NATIVE = { w: 480, h: 272 };

export class StageViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    this.active = true;
    this.level = null;
    this.levelW = 1;
    this.levelH = 1;
    this._origin = { x0: 0, y1: 0 }; // world-space top-left of the stage

    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0xffffff, 1); // the engine clears its frame to white
    el.appendChild(renderer.domElement);
    this.renderer = renderer;

    this.scene = new THREE.Scene();
    this.group = new THREE.Group(); // the loaded stage (replaced per level)
    this.scene.add(this.group);

    // Orthographic camera looking down -z; frustum = the viewport in screen px,
    // divided by the map camera's zoom (px per world unit).
    this.camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 1, 30000);
    this.camera.position.z = 10000;

    // MapCamera drives a Pixi-style world (position = the screen point of the
    // map origin, scale = zoom). This shim records those and _applyCam maps
    // them onto the orthographic camera. Map space is y-down with its origin
    // at the stage's top-left, exactly like a tilemap.
    this.app = { screen: { width: 1, height: 1 } };
    this.world = {
      position: {
        x: 0, y: 0,
        set(x, y) { this.x = x; this.y = y; },
      },
      scale: { set() { /* zoom lives on this.cam.zoom */ } },
    };
    this.cam = new MapCamera(this, {
      bounds: () => ({ w: this.levelW, h: this.levelH }),
      wrapX: () => 0,
      onApply: () => this._applyCam(),
    });
    this.cam.wirePointer();

    this._resize = this._resize.bind(this);
    this._ro = new ResizeObserver(this._resize);
    this._ro.observe(el);
    this._resize();

    const tick = () => {
      requestAnimationFrame(tick);
      if (this.active === false) return;
      renderer.render(this.scene, this.camera);
    };
    requestAnimationFrame(tick);
  }

  async init() {
    this.meta = await fetch(BASE + 'manifest.json').then((r) => r.json());
    return this.meta;
  }

  async loadLevel(entry) {
    const level = await fetch(BASE + entry.file).then((r) => r.json());
    this.level = level;

    // drop the previous stage
    this.group.traverse((o) => {
      if (o.isMesh) {
        o.geometry.dispose();
        for (const m of Array.isArray(o.material) ? o.material : [o.material]) {
          if (m.map) m.map.dispose();
          m.dispose();
        }
      }
    });
    this.scene.remove(this.group);
    this.group = new THREE.Group();
    this.scene.add(this.group);

    const ex = level.extents;
    this.levelW = ex.max[0] - ex.min[0];
    this.levelH = ex.max[1] - ex.min[1];
    this._origin = { x0: ex.min[0], y1: ex.max[1] };

    const loader = new GLTFLoader();
    const load = (path) => new Promise((res, rej) => loader.load(BASE + path, res, undefined, rej));
    // the engine draws the whole background pass before the foreground pass,
    // then back-to-front by layer z within each; the render-order bias keeps
    // the foreground on top even where a background piece has a larger z
    const parts = [
      { path: level.sky, bias: 0 },
      { path: level.mesh?.glb, bias: 1000 },
    ].filter((p) => p.path);
    for (const { path, bias } of parts) {
      const gltf = await load(path);
      gltf.scene.traverse((o) => {
        if (!o.isMesh) return;
        o.geometry.computeBoundingBox();
        const bb = o.geometry.boundingBox;
        o.renderOrder = bias + (bb.min.z + bb.max.z) / 2;
        for (const m of Array.isArray(o.material) ? o.material : [o.material]) {
          m.transparent = true;
          m.depthTest = false;
          m.depthWrite = false;
        }
      });
      this.group.add(gltf.scene);
    }

    // the collision contours (a line GLB), a toggleable overlay above everything
    this.collisionGroup = new THREE.Group();
    this.collisionGroup.visible = this._collisionOn || false;
    this.group.add(this.collisionGroup);
    if (level.collision?.glb) {
      const gltf = await load(level.collision.glb);
      gltf.scene.traverse((o) => {
        if (!o.isLine && !o.isLineSegments && !o.isMesh) return;
        o.renderOrder = 1e6;
        for (const m of Array.isArray(o.material) ? o.material : [o.material]) {
          m.depthTest = false;
          m.depthWrite = false;
        }
      });
      this.collisionGroup.add(gltf.scene);
    }

    // frame the exported view rect — the PSP screen centred where play
    // begins — zoomable out to the whole stage and in to 4x native
    this.cam.fitView(level.view || { x: 0, y: 0, w: NATIVE.w, h: NATIVE.h }, { maxNativeFactor: 4 });
    if (this.hud) {
      this.hud.textContent = `${level.name} — ${Math.round(this.levelW)}×${Math.round(this.levelH)} world units`;
    }
    return level;
  }

  setLayer(name, on) {
    if (name === 'collision') {
      this._collisionOn = on;
      if (this.collisionGroup) this.collisionGroup.visible = on;
    }
  }

  // Map the shared camera's pan/zoom state onto the orthographic camera. The
  // map origin (0,0) is the stage's world top-left; map y grows downward.
  _applyCam() {
    const sw = this.app.screen.width, sh = this.app.screen.height;
    const z = this.cam.zoom;
    const pos = this.world.position;
    // the world point at the viewport centre
    const wcx = this._origin.x0 + (sw / 2 - pos.x) / z;
    const wcy = this._origin.y1 - (sh / 2 - pos.y) / z;
    this.camera.left = -sw / 2 / z;
    this.camera.right = sw / 2 / z;
    this.camera.top = sh / 2 / z;
    this.camera.bottom = -sh / 2 / z;
    this.camera.position.x = wcx;
    this.camera.position.y = wcy;
    this.camera.updateProjectionMatrix();
  }

  _resize() {
    const w = this.el.clientWidth || 1, h = this.el.clientHeight || 1;
    this.app.screen.width = w;
    this.app.screen.height = h;
    this.renderer.setSize(w, h, false);
    this.cam.apply();
  }
}
