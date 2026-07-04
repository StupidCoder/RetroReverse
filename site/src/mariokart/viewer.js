// Mario Kart DS — NITRO model viewer. The models are the game's own NSBMD scenes
// (nodes + SBC scene bytecode + GX display lists), decoded in Go (tools/nds/nitro,
// Part IV) and exported as standard GLB with their NSBTX textures embedded — the
// texture↔palette pairing comes from each model's own material block. Textures are
// kept nearest-filtered so the DS's 32x32-texel art stays crisp.
//
// Courses carry two extras the manifest links to the entry:
//   • skybox — the track's far "_V" model. glTF has no skybox concept, so we load it
//     as an ordinary mesh, lock its origin to the camera every frame and draw it
//     behind everything (depth-test off) — exactly how the DS renders it, a backdrop
//     that turns with you but never gets closer.
//   • path — the CPU racers' drive line (NKM EPOI) as a polyline in the GLB's own
//     space. "Drive" mode flies the camera along a Catmull-Rom through it.
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

const MODELS = 'public/mariokart/models/';

export class ModelViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0x0a0d12, 1);
    renderer.autoClear = false; // we clear once, then draw skybox behind the scene
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.add(new THREE.AmbientLight(0xffffff, 1.1));
    const key = new THREE.DirectionalLight(0xffffff, 1.6);
    key.position.set(2, 4, 3);
    scene.add(key);
    const rim = new THREE.DirectionalLight(0x8899cc, 0.6);
    rim.position.set(-3, 1, -2);
    scene.add(rim);

    const camera = new THREE.PerspectiveCamera(40, 1, 0.01, 100);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 1.2;

    this.three = { renderer, scene, camera, controls, group: null };
    this.loader = new GLTFLoader();
    this.models = [];

    // per-course state
    this.gen = 0;              // load generation, to drop stale async results
    this.skybox = null;        // THREE.Object3D locked to the camera
    this.skyboxCenter = new THREE.Vector3(); // its modelled centre, pinned to the camera
    this.wantSkybox = true;    // Skybox toggle
    this.curve = null;         // CatmullRomCurve3 of the CPU drive line
    this.wantDrive = false;    // desired Drive state (survives track switches)
    this.driveOn = false;      // whether we are actually driving right now
    this.driveU = 0;           // position along the curve (0..1)
    this.driveDir = 1;         // travel direction (open paths ping-pong)
    this.eyeH = 3;             // camera height above the drive line, GLB units
    this.frame = null;         // { center, size } of the main model, for re-framing
    this._look = new THREE.Vector3();
    this._clock = new THREE.Clock();

    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      const dt = Math.min(this._clock.getDelta(), 0.1);
      if (this.active === false) return; // paused while another viewer is shown
      if (this.driveOn && this.curve) this._drive(dt);
      else controls.update();
      // Backdrop rides the camera: keep the dome's own centre on the camera so it
      // always encloses us (parallax-free), like the DS drawing it camera-relative.
      if (this.skybox) this.skybox.position.copy(camera.position).sub(this.skyboxCenter);
      renderer.clear();
      renderer.render(scene, camera);
    };
    tick();
  }

  _resize() {
    const w = this.el.clientWidth, h = this.el.clientHeight || Math.round(w * 0.62);
    if (!w) return;
    const { renderer, camera } = this.three;
    renderer.setSize(w, h, false);
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
  }

  async init() {
    this.models = await fetch('public/mariokart/models.json').then(r => r.json());
    return this.models;
  }

  // The studio drives these: "skybox" shows/hides the backdrop, "drive" enters/exits
  // the fly-along. Both no-op for models (karts, characters) that lack the piece.
  setLayer(id, on) {
    if (id === 'skybox') {
      this.wantSkybox = on;
      if (this.skybox) this.skybox.visible = on;
    } else if (id === 'drive') {
      this.wantDrive = on;
      this._setDrive(on);
    }
  }

  loadModel(i) {
    const m = this.models[i];
    if (!m) return;
    const gen = ++this.gen;
    // Stop any in-progress fly-along and drop the old line synchronously, before the
    // async load resolves — otherwise a still-running drive would follow stale data.
    this.curve = null; this.curveLoop = false;
    this._setDrive(false);

    this.loader.load(MODELS + m.file, (gltf) => {
      if (gen !== this.gen) return; // superseded by a newer selection
      const { scene, camera, controls } = this.three;
      this._disposeGroup();
      this._disposeSkybox();

      const group = gltf.scene;
      let tris = 0;
      group.traverse(o => {
        if (o.isMesh) {
          tris += (o.geometry.attributes.position.count / 3) | 0;
          if (o.material && o.material.map) {
            o.material.map.magFilter = THREE.NearestFilter; // DS textures are tiny: keep crisp
            o.material.map.needsUpdate = true;
          }
        }
      });
      scene.add(group);
      this.three.group = group;

      // Frame the main model (the skybox, once loaded, is excluded).
      const box = new THREE.Box3().setFromObject(group);
      const c = box.getCenter(new THREE.Vector3());
      const size = box.getSize(new THREE.Vector3()).length() || 1;
      this.frame = { center: c.clone(), size };
      this.eyeH = Math.min(Math.max(size * 0.01, 1), 12);
      camera.near = size / 100;
      camera.far = size * 20;
      camera.updateProjectionMatrix();
      this._orbitFrame();

      let detail = `${m.name} — ${tris.toLocaleString()} triangles, textures as shipped on cartridge`;
      if (this.hud) this.hud.textContent = detail;

      if (m.skybox) this._loadSkybox(m.skybox, gen);
      if (m.path) this._loadPath(m.path, gen);
    });
  }

  // Load the "_V" far model as a camera-locked backdrop: full-bright (unlit) so it
  // reads as sky/scenery, depth-test off so it never occludes the track.
  _loadSkybox(file, gen) {
    this.loader.load(MODELS + file, (gltf) => {
      if (gen !== this.gen) return;
      const sky = gltf.scene;
      sky.traverse(o => {
        if (!o.isMesh) return;
        const src = o.material;
        if (src && src.map) { src.map.magFilter = THREE.NearestFilter; src.map.needsUpdate = true; }
        o.material = new THREE.MeshBasicMaterial({
          map: src ? src.map : null,
          vertexColors: true,
          transparent: !!(src && src.transparent),
          alphaTest: src ? src.alphaTest : 0,
          side: THREE.DoubleSide,
          depthWrite: false,
          depthTest: false,
          fog: false,
        });
        o.renderOrder = -1000;
        o.frustumCulled = false;
      });
      this._disposeSkybox();
      // The dome is modelled around its own centre (not the origin); remember it so
      // the tick can keep that centre pinned to the camera.
      this.skyboxCenter = new THREE.Box3().setFromObject(sky).getCenter(new THREE.Vector3());
      sky.visible = this.wantSkybox;
      this.three.scene.add(sky);
      this.skybox = sky;
    });
  }

  async _loadPath(file, gen) {
    try {
      const doc = await fetch(MODELS + file).then(r => r.json());
      if (gen !== this.gen || !doc.points || doc.points.length < 2) return;
      const pts = doc.points.map(p => new THREE.Vector3(p[0], p[1], p[2]));
      this.curve = new THREE.CatmullRomCurve3(pts, !!doc.loop, 'catmullrom', 0.5);
      this.curveLoop = !!doc.loop;
      if (this.wantDrive) this._setDrive(true); // resume a fly-along carried over from the last track
    } catch { /* no path: Drive stays unavailable */ }
  }

  // ---- Drive mode: fly the camera along the CPU drive line ----
  _setDrive(on) {
    if (on === this.driveOn) return;
    this.driveOn = on && !!this.curve;
    const { controls } = this.three;
    if (this.driveOn) {
      controls.enabled = false;
      controls.autoRotate = false;
      this.driveU = 0; this.driveDir = 1;
    } else {
      controls.enabled = true;
      controls.autoRotate = true;
      if (this.frame) this._orbitFrame();
    }
  }

  _drive(dt) {
    const { camera } = this.three;
    const speed = 1 / 32; // ~one lap every 32s in curve-parameter space
    const look = 0.01;    // aim this far ahead along the line
    if (this.curveLoop) {
      this.driveU = (this.driveU + dt * speed) % 1;
    } else {
      this.driveU += dt * speed * this.driveDir;
      if (this.driveU > 1) { this.driveU = 1; this.driveDir = -1; }
      else if (this.driveU < 0) { this.driveU = 0; this.driveDir = 1; }
    }
    const u = this.driveU;
    const ahead = this.curveLoop ? (u + look) % 1 : Math.min(Math.max(u + look * this.driveDir, 0), 1);
    const p = this.curve.getPointAt(u);
    const a = this.curve.getPointAt(ahead);
    camera.position.set(p.x, p.y + this.eyeH, p.z);
    this._look.set(a.x, a.y + this.eyeH, a.z);
    camera.up.set(0, 1, 0);
    camera.lookAt(this._look);
  }

  _orbitFrame() {
    const { camera, controls } = this.three;
    const { center: c, size } = this.frame;
    controls.target.copy(c);
    camera.position.set(c.x + size * 0.9, c.y + size * 0.55, c.z + size * 0.9);
    camera.up.set(0, 1, 0);
    controls.update();
  }

  _disposeGroup() {
    if (!this.three.group) return;
    this.three.scene.remove(this.three.group);
    this.three.group.traverse(o => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (o.material.map) o.material.map.dispose();
        o.material.dispose();
      }
    });
    this.three.group = null;
  }

  _disposeSkybox() {
    if (!this.skybox) return;
    this.three.scene.remove(this.skybox);
    this.skybox.traverse(o => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (o.material.map) o.material.map.dispose();
        o.material.dispose();
      }
    });
    this.skybox = null;
  }
}
