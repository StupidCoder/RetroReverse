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
import { clone as cloneSkinned } from 'three/addons/utils/SkeletonUtils.js';

const MODELS = 'public/sm64ds/models/';

// Coin spin, from the coin actor's step at $020B23A4: yaw += $C00 angle units
// per frame ($10000 = 360°), at the game's 30 fps actor tick.
const COIN_SPIN = (0xC00 / 0x10000) * 30 * 2 * Math.PI; // rad/s
const COIN_MODELS = new Set(['arc0_5', 'arc0_7']);

// Signposts (obj_tatefuda, actor 184): the traced behavior is a proximity
// dialog — the engine sets an in-range flag in the actor's +$B0 word and a
// button press starts the sign's message ($020BB060). The message text lives
// in the per-language archives and isn't extracted yet, so the viewer shows
// the traced mechanics instead.
const SIGN_MODEL = 'obj_tatefuda';

// Goomba wander AI, traced from daKrb_c's state-0 handler ($0212ABD4/$0212AE98,
// overlay 84): walk speed 2.0 world-units/frame (state table $02130248), yaw
// eases toward a target heading at $200 angle-units/frame, a 100-frame timer
// repicks — 75% turn by a random signed 16-bit angle, 25% pause — and a
// 1000-unit leash around the spawn point steers it home. Stage-GLB units are
// world/1000 and the actor tick is 30 fps.
const KRB = {
  speed: (0x8000 / 4096) * 30 / 1000,           // stage units/s
  turn: (0x200 / 0x10000) * 2 * Math.PI * 30,   // rad/s
  repick: 100 / 30,                              // s
  leash: 1.0,                                    // stage units
};
const SIGN_TEXT = 'This signpost is readable in the game: stand in range (the engine flags the ' +
  'actor at +$B0) and press A — the sign starts its dialog (traced at $020BB060). ' +
  'The message text lives in the per-language archives, not yet extracted.';

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
    // level object placements (decoded from the level overlays' object tables)
    this.objectsGroup = null;
    this.billboards = [];   // flat tree quads, yawed toward the camera each frame
    this.spinners = [];     // coins: yaw at the traced $C00/frame
    this.signposts = [];    // clickable obj_tatefuda instances
    this.mixers = [];       // skinned enemies playing their .bca walk clips
    this.patrollers = [];   // goombas wandering per their traced AI
    this.wantObjects = true;
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
      // Billboard trees yaw around world-up toward the camera; up stays vertical
      // (the DS renders these flat quads camera-facing).
      for (const b of this.billboards) {
        b.rotation.y = Math.atan2(camera.position.x - b.position.x, camera.position.z - b.position.z);
      }
      for (const s of this.spinners) s.rotation.y += COIN_SPIN * dt;
      for (const mx of this.mixers) mx.update(dt);
      for (const g of this.patrollers) {
        g.timer -= dt;
        if (g.timer <= 0) {
          g.timer = KRB.repick;
          if (Math.random() < 0.25) g.paused = true;
          else { g.paused = false; g.target += (Math.random() * 2 - 1) * Math.PI; }
        }
        // leash: outside the home radius, head back
        const hx = g.home.x - g.obj.position.x, hz = g.home.z - g.obj.position.z;
        if (hx * hx + hz * hz > KRB.leash * KRB.leash) g.target = Math.atan2(hx, hz);
        // ease yaw the short way around, walk forward
        let d = ((g.target - g.yaw + 3 * Math.PI) % (2 * Math.PI)) - Math.PI;
        const step = KRB.turn * dt;
        g.yaw += Math.abs(d) <= step ? d : Math.sign(d) * step;
        g.obj.rotation.y = g.yaw;
        if (!g.paused) {
          g.obj.position.x += Math.sin(g.yaw) * KRB.speed * dt;
          g.obj.position.z += Math.cos(g.yaw) * KRB.speed * dt;
        }
      }
      renderer.render(scene, camera);
    };
    tick();

    // Signpost interaction: a click (not a drag) raycasts the placed signposts.
    this._ray = new THREE.Raycaster();
    let downAt = null;
    renderer.domElement.addEventListener('pointerdown', e => { downAt = [e.clientX, e.clientY]; });
    renderer.domElement.addEventListener('pointerup', e => {
      if (!downAt || Math.hypot(e.clientX - downAt[0], e.clientY - downAt[1]) > 5) return;
      downAt = null;
      this._clickSign(e);
    });
  }

  _clickSign(e) {
    this._hideSign();
    if (!this.signposts.length || this.wantObjects === false) return;
    const r = this.three.renderer.domElement.getBoundingClientRect();
    const p = new THREE.Vector2(
      ((e.clientX - r.left) / r.width) * 2 - 1,
      -((e.clientY - r.top) / r.height) * 2 + 1);
    this._ray.setFromCamera(p, this.three.camera);
    const hits = this._ray.intersectObjects(this.signposts, true);
    if (!hits.length) return;
    const d = document.createElement('div');
    d.style.cssText = 'position:absolute;left:12px;bottom:12px;max-width:min(440px,80%);' +
      'background:rgba(10,13,18,.92);border:1px solid #3a4a5c;border-radius:8px;' +
      'padding:10px 12px;font:12px/1.5 system-ui;color:#dfe6f0;pointer-events:none;z-index:5';
    d.textContent = SIGN_TEXT;
    // this.el is the studio's .mount (position:absolute, inset:0) — already a
    // positioning context; never touch its position or the canvas collapses.
    this.el.appendChild(d);
    this._signBox = d;
    clearTimeout(this._signTimer);
    this._signTimer = setTimeout(() => this._hideSign(), 9000);
  }

  _hideSign() {
    if (this._signBox) { this._signBox.remove(); this._signBox = null; }
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

  // The studio drives this: "objects" shows/hides the placed level objects.
  setLayer(id, on) {
    if (id === 'objects') {
      this.wantObjects = on;
      if (this.objectsGroup) this.objectsGroup.visible = on;
    }
  }

  loadModel(i) {
    const m = this.models[i];
    if (!m) return;
    const gen = ++this.gen;
    this.loader.load(MODELS + m.file, (gltf) => {
      if (gen !== this.gen) return; // superseded
      const { scene, camera, controls } = this.three;
      this._dispose();
      this._disposeObjects();

      const group = gltf.scene;
      let tris = 0;
      group.traverse(o => {
        if (o.isMesh) {
          tris += (o.geometry.attributes.position.count / 3) | 0;
          if (o.material && o.material.map) {
            o.material.map.magFilter = THREE.NearestFilter; // DS textures are tiny: keep crisp
            o.material.map.needsUpdate = true;
          }
          if (o.isSkinnedMesh) o.frustumCulled = false;
        }
      });
      // Animated models (the enemies) play their first .bca clip in the gallery.
      if (gltf.animations && gltf.animations.length) {
        const mx = new THREE.AnimationMixer(group);
        const clip = gltf.animations.find(a => a.name.endsWith('_walk')) || gltf.animations[0];
        mx.clipAction(clip).play();
        this.mixers.push(mx);
      }
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

      if (m.objects) this._loadObjects(m.objects, gen, size);
    });
  }

  // Place the level's objects (decoded from the level overlay's object tables).
  // Placements bound to an extracted model load it once and clone per instance;
  // the rest show as small markers — their models aren't extracted yet.
  async _loadObjects(file, gen, size) {
    let doc;
    try {
      doc = await fetch('public/sm64ds/' + file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !doc.objects) return;

    const protos = new Map();
    for (const o of doc.objects) {
      if (o.m && !protos.has(o.m)) {
        protos.set(o.m, new Promise(res =>
          this.loader.load(MODELS + o.m + '.glb', g => {
            g.scene.traverse(n => {
              if (n.isMesh && n.material && n.material.map) {
                n.material.map.magFilter = THREE.NearestFilter;
                n.material.map.needsUpdate = true;
              }
              if (n.isSkinnedMesh) n.frustumCulled = false; // skinned bounds don't track the pose
            });
            res({ scene: g.scene, animations: g.animations || [] });
          }, undefined, () => res(null))));
      }
    }
    await Promise.all(protos.values());
    if (gen !== this.gen) return;

    const group = new THREE.Group();
    const bills = [], spinners = [], signs = [];
    const markerGeo = new THREE.SphereGeometry(size / 260, 8, 6);
    const markerMat = new THREE.MeshBasicMaterial({ color: 0xffd75e, transparent: true, opacity: 0.75 });
    let placed = 0, markers = 0;
    for (const o of doc.objects) {
      let inst;
      if (o.m) {
        const proto = await protos.get(o.m);
        if (proto) {
          // Skinned models (the enemies) clone with their skeletons and play
          // their .bca walk cycle (decoded from the cartridge, 30 fps).
          if (proto.animations.length) {
            inst = cloneSkinned(proto.scene);
            const clip = proto.animations.find(a => a.name.endsWith('_walk') || a.name.endsWith('_run'))
              || proto.animations.find(a => a.name.endsWith('_wait')) || proto.animations[0];
            const mx = new THREE.AnimationMixer(inst);
            mx.clipAction(clip).play();
            // desync instances so a troop doesn't march in lockstep
            mx.setTime(Math.random() * clip.duration);
            this.mixers.push(mx);
          } else {
            inst = proto.scene.clone();
          }
          if (o.s) inst.scale.setScalar(o.s);
          if (o.b) bills.push(inst);
          else if (o.ry) inst.rotation.y = o.ry * Math.PI / 180;
          if (COIN_MODELS.has(o.m)) spinners.push(inst);
          if (o.m === SIGN_MODEL) signs.push(inst);
          if (o.m === 'kuribo_model') {
            const yaw = (o.ry || 0) * Math.PI / 180;
            this.patrollers.push({
              obj: inst, yaw, target: yaw, paused: false,
              timer: Math.random() * KRB.repick,
              home: { x: o.p[0], z: o.p[2] },
            });
          }
          placed++;
        }
      }
      if (!inst) {
        inst = new THREE.Mesh(markerGeo, markerMat);
        markers++;
      }
      inst.position.set(o.p[0], o.p[1], o.p[2]);
      group.add(inst);
    }
    group.visible = this.wantObjects;
    this.three.scene.add(group);
    this.objectsGroup = group;
    this.billboards = bills;
    this.spinners = spinners;
    this.signposts = signs;
    if (this.hud) {
      this.hud.textContent += ` · ${placed + markers} objects placed (${markers} as markers)` +
        (signs.length ? ' · signposts are clickable' : '');
    }
  }

  _disposeObjects() {
    this.billboards = [];
    this.spinners = [];
    this.signposts = [];
    this.mixers = [];
    this.patrollers = [];
    this._hideSign();
    if (!this.objectsGroup) return;
    this.three.scene.remove(this.objectsGroup);
    this.objectsGroup.traverse(o => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (o.material.map) o.material.map.dispose();
        o.material.dispose();
      }
    });
    this.objectsGroup = null;
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
