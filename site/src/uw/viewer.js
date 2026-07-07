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

// UW's 32-fine-angle → 8-view remap (DGROUP:05AC), reverse-engineered from the
// creature emit path (2DFE:0221). The game selects a directional sprite as
//   view = REMAP[(heading*4 + 0x20 - cameraAngle) & 0x1F]
// where cameraAngle is the bearing from the camera to the creature in 32 steps
// and heading is the creature's facing (0-7). View 0 is the creature's back,
// view 4 its front. We reproduce it per render from the true line-of-sight.
const VIEW_REMAP = [
  0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 4, 4,
  4, 4, 4, 5, 5, 5, 6, 6, 6, 6, 6, 7, 7, 7, 0, 0,
];

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
      this._updateCreatures();
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

    // Billboard objects (items, creatures): OBJECTS.GR sprites drawn as
    // camera-facing THREE.Sprites. Their exported position is the base on the
    // floor, so lift each by half its height; they share the level's centring.
    const spriteGroup = new THREE.Group();
    spriteGroup.position.set(-c.x, -c.y, -c.z);
    const spriteMats = (data.spriteTex || []).map((png) => {
      const t = this._texLoader.load(png);
      t.magFilter = THREE.NearestFilter;
      t.colorSpace = THREE.SRGBColorSpace;
      return new THREE.SpriteMaterial({ map: t, transparent: true, alphaTest: 0.5, depthWrite: true });
    });
    for (const s of data.sprites || []) {
      const spr = new THREE.Sprite(spriteMats[s.tex]);
      spr.position.set(s.pos[0], s.pos[1] + s.h / 2, s.pos[2]);
      spr.scale.set(s.w, s.h, 1);
      spriteGroup.add(spr);
    }

    // Creatures are directional (Doom-style) billboards: eight view frames that
    // the render loop swaps by the camera-to-creature angle and the creature's
    // heading, so a monster shows its back, side or face as you circle it.
    const creatures = [];
    for (const c of data.creatures || []) {
      const dirs = c.dirs.map((d) => ({ mat: spriteMats[d.tex], w: d.w, h: d.h }));
      const spr = new THREE.Sprite(dirs[0].mat);
      spriteGroup.add(spr);
      const rec = { spr, dirs, heading: c.heading, base: c.pos, view: -1 };
      creatures.push(rec);
      this._applyView(rec, 0); // place with view 0 until the first update
    }

    scene.add(spriteGroup);
    this._spriteGroup = spriteGroup;
    this._spriteMats = spriteMats;
    this._creatures = creatures;

    // Start inside the dungeon (it's now ceiling-enclosed): the exported spawn
    // is an interior point at eye height; place the camera there looking ahead.
    // Fall back to an angled overview if no spawn was exported.
    const { camera, controls } = this.three;
    const r = geo.boundingBox.getSize(new THREE.Vector3()).length() || 40;
    if (data.spawn) {
      const [sx, sy, sz] = data.spawn;
      const [dx, dy, dz] = data.spawnDir || [1, 0, 0]; // initial look direction
      camera.position.set(sx - c.x, sy - c.y, sz - c.z);
      controls.target.set(sx - c.x + dx, sy - c.y + dy, sz - c.z + dz);
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

  // _applyView points a creature sprite at one of its eight view frames,
  // rescaling and re-seating its base on the floor (frames differ in size).
  _applyView(rec, view) {
    if (view === rec.view) return;
    rec.view = view;
    const d = rec.dirs[view] || rec.dirs[0];
    rec.spr.material = d.mat;
    rec.spr.scale.set(d.w, d.h, 1);
    rec.spr.position.set(rec.base[0], rec.base[1] + d.h / 2, rec.base[2]);
  }

  // _updateCreatures picks each creature's view frame from the horizontal
  // bearing between it and the camera, in the game's own convention: the fine
  // angle (0-31) of the camera→creature vector is the cameraAngle, and
  //   view = REMAP[(heading*4 + 32 - cameraAngle) & 31].
  _updateCreatures() {
    const list = this._creatures;
    if (!list || !list.length) return;
    const cam = this.three.camera.position;
    const wp = this._wp || (this._wp = new THREE.Vector3());
    for (const rec of list) {
      rec.spr.getWorldPosition(wp);
      const dx = wp.x - cam.x, dz = wp.z - cam.z; // camera → creature (horizontal)
      // Match heading's fine-angle convention: heading h faces world such that
      // atan2(-fx, -fz) == h*4 steps; use the same mapping for the bearing.
      let camAngle = Math.round((Math.atan2(-dx, -dz) / (2 * Math.PI)) * 32) & 31;
      const idx = (rec.heading * 4 + 32 - camAngle) & 31;
      this._applyView(rec, VIEW_REMAP[idx]);
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
    if (this._spriteGroup) {
      this.three.scene.remove(this._spriteGroup);
      this._spriteGroup = null;
    }
    for (const m of this._spriteMats || []) { m.map?.dispose(); m.dispose(); }
    this._spriteMats = null;
    this._creatures = null;
  }
}
