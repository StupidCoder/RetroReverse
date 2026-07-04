// Mario Kart DS — NITRO model viewer. The models are the game's own NSBMD scenes
// (nodes + SBC scene bytecode + GX display lists), decoded in Go (tools/nds/nitro,
// Part IV) and exported as standard GLB with their NSBTX textures embedded — the
// texture↔palette pairing comes from each model's own material block. Textures are
// kept nearest-filtered so the DS's 32x32-texel art stays crisp.
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

export class ModelViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0x0a0d12, 1);
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

    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      if (this.active === false) return; // paused while another viewer is shown
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
    this.models = await fetch('public/mariokart/models.json').then(r => r.json());
    return this.models;
  }

  loadModel(i) {
    const m = this.models[i];
    if (!m) return;
    this.loader.load('public/mariokart/models/' + m.file, (gltf) => {
      const { scene, camera, controls } = this.three;
      if (this.three.group) {
        scene.remove(this.three.group);
        this.three.group.traverse(o => {
          if (o.geometry) o.geometry.dispose();
          if (o.material) {
            if (o.material.map) o.material.map.dispose();
            o.material.dispose();
          }
        });
      }
      const group = gltf.scene;
      let tris = 0;
      group.traverse(o => {
        if (o.isMesh) {
          tris += (o.geometry.attributes.position.count / 3) | 0;
          if (o.material && o.material.map) {
            // DS textures are tiny: keep them pixel-crisp.
            o.material.map.magFilter = THREE.NearestFilter;
            o.material.map.needsUpdate = true;
          }
        }
      });
      scene.add(group);
      this.three.group = group;

      // Frame the model.
      const box = new THREE.Box3().setFromObject(group);
      const c = box.getCenter(new THREE.Vector3());
      const size = box.getSize(new THREE.Vector3()).length() || 1;
      controls.target.copy(c);
      camera.position.set(c.x + size * 0.9, c.y + size * 0.55, c.z + size * 0.9);
      camera.near = size / 100;
      camera.far = size * 20;
      camera.updateProjectionMatrix();
      controls.update();

      if (this.hud) this.hud.textContent = `${m.name} — ${tris.toLocaleString()} triangles, textures as shipped on cartridge`;
    });
  }
}
