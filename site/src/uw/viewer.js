// Ultima Underworld — static level geometry viewer.
//
// The mesh is the dungeon's own tile geometry, reverse-engineered from LEV.ARK
// and reimplemented in Go ("Ultima Underworld (PC)/extract/levgeo") and exported
// by cmd/levexport: floors, ceilings and walls, each carrying its real W64.TR /
// F32.TR texture through the per-level texture list. This viewer just loads that
// self-contained JSON (positions, UVs, material groups, and each texture as a
// data-URI PNG) into a three.js BufferGeometry and lets you fly through it.

import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

const LEVELS = 'public/uw/';

export class LevelViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0x05060a, 1);
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.fog = new THREE.Fog(0x05060a, 20, 70); // the Abyss swallows distance
    const camera = new THREE.PerspectiveCamera(60, 1, 0.01, 500);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;

    this.three = { renderer, scene, camera, controls, group: null };
    this.gen = 0;
    this.fly = new FlyCam(camera, controls, el);
    this._clock = new THREE.Clock();
    this._texLoader = new THREE.TextureLoader();

    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      const dt = Math.min(this._clock.getDelta(), 0.1);
      if (this.active === false) return;
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
    this.levels = await fetch(LEVELS + 'levels.json').then(r => r.json());
    return this.levels;
  }

  async loadLevel(lvl) {
    const gen = ++this.gen;
    const data = await fetch(LEVELS + lvl.file).then(r => r.json());
    if (gen !== this.gen) return; // superseded

    this._dispose();
    const { scene } = this.three;

    // One texture (material) per group; UW textures are tiny, keep them crisp
    // and let walls tile vertically (a wall's V runs 0..height).
    const materials = data.textures.map((t) => {
      const tex = this._texLoader.load(t.png);
      tex.wrapS = tex.wrapT = THREE.RepeatWrapping;
      tex.magFilter = THREE.NearestFilter;
      tex.minFilter = THREE.LinearMipmapLinearFilter;
      tex.colorSpace = THREE.SRGBColorSpace;
      return new THREE.MeshBasicMaterial({ map: tex, side: THREE.DoubleSide });
    });

    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.Float32BufferAttribute(data.positions, 3));
    geo.setAttribute('uv', new THREE.Float32BufferAttribute(data.uvs, 2));
    for (const g of data.groups) geo.addGroup(g.start, g.count, g.material);
    geo.computeBoundingBox();

    const mesh = new THREE.Mesh(geo, materials);
    // Centre the level on the origin so the fly-cam starts inside it.
    const c = geo.boundingBox.getCenter(new THREE.Vector3());
    mesh.position.set(-c.x, -c.y, -c.z);
    scene.add(mesh);
    this.three.group = mesh;
    this._materials = materials;

    // Start inside the dungeon (it's now ceiling-enclosed): the exported spawn
    // is an interior point at eye height; place the camera there looking ahead.
    // Fall back to an angled overview if no spawn was exported.
    const { camera, controls } = this.three;
    const r = geo.boundingBox.getSize(new THREE.Vector3()).length() || 40;
    if (data.spawn) {
      const [sx, sy, sz] = data.spawn;
      camera.position.set(sx - c.x, sy - c.y, sz - c.z);
      controls.target.set(sx - c.x + 1, sy - c.y, sz - c.z); // look along +X
    } else {
      camera.position.set(0, r * 0.42, r * 0.62);
      controls.target.set(0, 0, 0);
    }
    controls.update();

    // Levels are explored with free-flight controls (WASD/arrows, or the touch
    // sticks), not the orbit camera used for single objects.
    controls.autoRotate = false;
    this.fly.setScale(r);
    this.fly.setMoveScale(0.25); // UW's levels are vast — quarter-speed movement (look speed unchanged)
    this.fly.setEnabled(true);

    if (this.hud?.detail) {
      this.hud.detail(`${(data.positions.length / 9) | 0} triangles · ${data.textures.length} textures · ${flyHint}`);
    }
  }

  _dispose() {
    const g = this.three.group;
    if (g) {
      this.three.scene.remove(g);
      g.geometry.dispose();
    }
    for (const m of this._materials || []) { m.map?.dispose(); m.dispose(); }
    this._materials = null;
    this.three.group = null;
  }
}
