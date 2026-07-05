// Super Mario 64 DS — model viewer. The models are the game's own ".bmd" scenes
// (its bespoke format, NOT the NITRO BMD0 container Mario Kart DS uses), decoded in
// Go (supermario64ds/extract/sm64ds) and exported as standard GLB. The low-level
// primitives are the same DS silicon as NITRO — GX display lists and the seven DS
// texture formats — so textures come out embedded and are kept nearest-filtered to
// stay pixel-crisp.
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

const MODELS = 'public/sm64ds/models/';

export class ModelViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0x0a0d12, 1);
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.add(new THREE.AmbientLight(0xffffff, 1.25));
    const key = new THREE.DirectionalLight(0xffffff, 1.5);
    key.position.set(2, 4, 3);
    scene.add(key);
    const rim = new THREE.DirectionalLight(0x8899cc, 0.6);
    rim.position.set(-3, 1, -2);
    scene.add(rim);

    const camera = new THREE.PerspectiveCamera(45, 1, 0.001, 1000);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 1.0;

    this.three = { renderer, scene, camera, controls, group: null };
    this.loader = new GLTFLoader();
    this.models = [];
    this.gen = 0;
    // Levels are explored with free-flight controls (WASD/arrows, or virtual
    // sticks on touch); objects and characters keep the slow auto-rotating orbit.
    this.fly = new FlyCam(camera, controls, el);
    this._clock = new THREE.Clock();

    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      const dt = Math.min(this._clock.getDelta(), 0.1);
      if (this.active === false) return; // paused while another viewer is shown
      this.fly.update(dt);
      controls.update();
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
    this.models = await fetch('public/sm64ds/models.json').then(r => r.json());
    return this.models;
  }

  loadModel(i) {
    const m = this.models[i];
    if (!m) return;
    const gen = ++this.gen;
    this.loader.load(MODELS + m.file, (gltf) => {
      if (gen !== this.gen) return; // superseded
      const { scene, camera, controls } = this.three;
      this._dispose();

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

      const box = new THREE.Box3().setFromObject(group);
      const c = box.getCenter(new THREE.Vector3());
      const size = box.getSize(new THREE.Vector3()).length() || 1;
      controls.target.copy(c);
      camera.position.set(c.x + size * 0.7, c.y + size * 0.5, c.z + size * 0.7);
      camera.near = size / 200;
      camera.far = size * 20;
      camera.updateProjectionMatrix();
      controls.update();

      // Levels get fly controls (no auto-rotation); everything else keeps the orbit.
      const isLevel = m.section === 'Levels';
      controls.autoRotate = !isLevel;
      this.fly.setScale(size);
      this.fly.setEnabled(isLevel);

      if (this.hud) {
        this.hud.textContent = `${m.name} — ${tris.toLocaleString()} triangles, textures as shipped on cartridge` +
          (isLevel ? ` · ${flyHint}` : '');
      }
    });
  }

  _dispose() {
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
}
