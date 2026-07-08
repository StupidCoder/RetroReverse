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
//   • objects — the NKM OBJI placements bound to the course's object GLBs
//     (Part V §2): item boxes, trees, Piantas, pipes… placed into the scene.
//     Billboard objects (flat camera-facing sprites like the Goomba) yaw around
//     world-up toward the camera each frame — their up axis stays world-aligned,
//     matching how the DS draws them.
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

// Format-2 asset tree. The manifest and the per-level envelopes live at this base;
// the level envelope's mesh.glb / sky paths are root-relative to it (models/…), so
// load them as BASE + path. Object-placement GLBs are still referenced by bare name
// inside the objects JSON (inner OBJI shape unchanged), so they load from MODELS.
const BASE = 'public/mario-kart-ds/';
const MODELS = BASE + 'models/';

// makeMover precomputes a route follower: the polyline's segments and total
// length (closing segment appended on loops), a shared plausible speed, and a
// random start offset so co-routed objects don't stack.
function makeMover(inst, route, billboard) {
  const pts = route.points;
  const segs = [];
  let total = 0;
  const n = route.loop ? pts.length : pts.length - 1;
  for (let i = 0; i < n; i++) {
    const a = pts[i], b = pts[(i + 1) % pts.length];
    const len = Math.hypot(b[0] - a[0], b[1] - a[1], b[2] - a[2]);
    segs.push({ a, b, len });
    total += len;
  }
  return {
    inst, segs, total, billboard,
    loop: !!route.loop,
    speed: 7.5, // GLB units/s (~120 world units/s); per-type speeds live in engine code
    dist: Math.random() * total,
  };
}

// trackAt samples a BTA0 component track at a (fractional) frame: a constant, or
// per-frame samples (one per `step` frames) with linear interpolation.
function trackAt(tr, frame) {
  if (!tr.samples || !tr.samples.length) return tr.const || 0;
  const pos = frame / tr.step;
  const i = Math.floor(pos);
  if (i >= tr.samples.length - 1) return tr.samples[tr.samples.length - 1];
  const f = pos - i;
  return tr.samples[i] * (1 - f) + tr.samples[i + 1] * f;
}

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
    // Tracks are explored with free-flight controls (WASD/arrows, or virtual
    // sticks on touch); karts, characters and map objects keep the orbit.
    this.fly = new FlyCam(camera, controls, el);
    this.flyWanted = false; // current model is a track (fly instead of auto-rotate)

    // per-course state
    this.gen = 0;              // load generation, to drop stale async results
    this.skybox = null;        // THREE.Object3D locked to the camera
    this.skyboxCenter = new THREE.Vector3(); // its modelled centre, pinned to the camera
    this.wantSkybox = true;    // Skybox toggle
    this.curve = null;         // CatmullRomCurve3 of the CPU drive line
    this.wantDrive = false;    // desired Drive state (survives track switches)
    this.driveOn = false;      // whether we are actually driving right now
    this.objectsGroup = null;  // placed map objects (OBJI)
    this.wantObjects = true;   // Objects toggle
    this.billboards = [];      // placed sprites to yaw toward the camera
    this.movers = [];          // placed objects following a PATH/POIT route
    this.animDefs = [];        // BTA0 texture-SRT tracks (per material name)
    this.liveAnims = [];       // { def, materials[] } bound to loaded meshes
    this.animClock = 0;        // frames (60/s) into the texture animations
    this.driveU = 0;           // position along the curve (0..1)
    this.driveDir = 1;         // travel direction (open paths ping-pong)
    // Kart-eye height above the drive line, GLB units (world/16). Constant, NOT
    // course-scaled: karts are the same size on every course, and scaling this with
    // the course put the camera through tunnel ceilings on the big tracks.
    this.eyeH = 1.4;
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
      if (this.driveOn && this.curve) {
        this._drive(dt);
      } else {
        this.fly.update(dt);
        controls.update();
      }
      // Backdrop rides the camera: keep the dome's own centre on the camera so it
      // always encloses us (parallax-free), like the DS drawing it camera-relative.
      if (this.skybox) this.skybox.position.copy(camera.position).sub(this.skyboxCenter);
      // Route followers walk their PATH polyline at constant world speed (the
      // engine's follower semantics, Part V §3), facing their travel direction.
      for (const mv of this.movers) this._advanceMover(mv, dt);
      // Billboards yaw around world-up toward the camera; their up stays vertical.
      for (const b of this.billboards) {
        b.rotation.y = Math.atan2(camera.position.x - b.position.x, camera.position.z - b.position.z);
      }
      // Texture-SRT animations (BTA0): scrolling water and boost-panel arrows.
      // The DS steps these at 60 fps; values are normalized texture space.
      if (this.liveAnims.length) {
        this.animClock += dt * 60;
        for (const la of this.liveAnims) {
          const f = this.animClock % la.def.frames;
          const s = trackAt(la.def.transS, f), t = trackAt(la.def.transT, f);
          for (const m of la.materials) m.map.offset.set(s, t);
        }
      }
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
    // The browse list is the format-2 manifest: its courses (levels[], each a
    // mesh3d entry that resolves to a level envelope) followed by its plain model
    // GLBs (karts, characters, shared objects), grouped by their section field.
    const manifest = await fetch(BASE + 'manifest.json').then(r => r.json());
    this.manifest = manifest;
    const levels = (manifest.levels || []).map(l => ({
      name: l.name,
      section: l.section,
      kind: 'mesh3d',       // marks a course: loadModel resolves the level envelope
      file: l.file,         // the level envelope json (levels/<stem>.json)
      objects: l.objects,   // present => the Studio shows the Objects toggle
    }));
    const models = (manifest.models || []).map(m => ({
      name: m.name,
      section: m.section,
      file: m.file,         // models/<x>.glb, root-relative to BASE
    }));
    this.models = levels.concat(models);
    return this.models;
  }

  // The studio drives these: "skybox" shows/hides the backdrop, "drive" enters/exits
  // the fly-along, "objects" shows/hides the placed map objects. All no-op for
  // models (karts, characters) that lack the piece.
  setLayer(id, on) {
    if (id === 'skybox') {
      this.wantSkybox = on;
      if (this.skybox) this.skybox.visible = on;
    } else if (id === 'drive') {
      this.wantDrive = on;
      this._setDrive(on);
    } else if (id === 'objects') {
      this.wantObjects = on;
      if (this.objectsGroup) this.objectsGroup.visible = on;
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
    this.animDefs = []; this.liveAnims = []; this.animClock = 0;

    if (m.kind === 'mesh3d') this._loadCourse(m, gen);
    else this._loadPlainModel(m, gen);
  }

  // Install a loaded GLB as the main scene group: dispose the old scene, keep the DS
  // textures crisp, frame the camera, and switch between fly (courses) and orbit
  // (karts/characters/objects) controls. Returns the framed size (GLB units).
  _installMain(gltf, isCourse, name) {
    const { scene, camera, controls } = this.three;
    this._disposeGroup();
    this._disposeSkybox();
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
      }
    });
    scene.add(group);
    this.three.group = group;

    // Frame the main model (the skybox, once loaded, is excluded).
    const box = new THREE.Box3().setFromObject(group);
    const c = box.getCenter(new THREE.Vector3());
    const size = box.getSize(new THREE.Vector3()).length() || 1;
    this.frame = { center: c.clone(), size };
    camera.near = size / 100;
    camera.far = size * 20;
    camera.updateProjectionMatrix();
    this._orbitFrame();

    // Tracks get fly controls (no auto-rotation); everything else keeps the orbit.
    this.flyWanted = isCourse;
    controls.autoRotate = !this.flyWanted;
    this.fly.setScale(size);
    this.fly.setEnabled(this.flyWanted && !this.driveOn);

    this.hudBase = `${name} — ${tris.toLocaleString()} triangles, textures as shipped on cartridge` +
      (this.flyWanted ? ` · ${flyHint}` : '');
    if (this.hud) this.hud.textContent = this.hudBase;

    this._bindAnims(group); // anims may already be loaded (fetch races the GLB)
    return size;
  }

  // A plain model entry (kart/character/shared object): one GLB, root-relative to BASE.
  _loadPlainModel(m, gen) {
    this.loader.load(BASE + m.file, (gltf) => {
      if (gen !== this.gen) return; // superseded by a newer selection
      this._installMain(gltf, false, m.name);
    });
  }

  // A course entry: fetch the level envelope, load its track mesh, then attach the
  // envelope's extras — the "_V" skybox, the CPU drive line (path), the OBJI object
  // placements, and any texture-SRT animations. mesh.glb / sky paths in the envelope
  // are root-relative to BASE; the path/objects/anims sidecar JSONs live under levels/.
  async _loadCourse(m, gen) {
    let level;
    try {
      level = await fetch(BASE + m.file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !level || !level.mesh || !level.mesh.glb) return;
    this.loader.load(BASE + level.mesh.glb, (gltf) => {
      if (gen !== this.gen) return; // superseded by a newer selection
      this._installMain(gltf, true, level.name || m.name);
      if (level.sky) this._loadSkybox(level.sky, gen);
      if (level.path) this._loadPath('levels/' + level.path, gen);
      if (level.objectsFile) this._loadObjects('levels/' + level.objectsFile, gen);
      if (level.anims) this._loadAnims('levels/' + level.anims, gen);
    });
  }

  // Load the course's BTA0 texture-SRT tracks and bind them to every material of
  // that name currently in the scene (track, skybox, placed objects — whichever
  // have loaded; later loads bind themselves via _bindAnims).
  async _loadAnims(file, gen) {
    let doc;
    try {
      doc = await fetch(BASE + file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !doc.anims) return;
    this.animDefs = doc.anims;
    for (const root of [this.three.group, this.skybox, this.objectsGroup]) {
      if (root) this._bindAnims(root);
    }
  }

  // Attach texture animations to a freshly loaded subtree: any mesh material
  // whose name matches an animated material joins that animation's update list.
  // Clones share material instances, so one binding animates every placement.
  _bindAnims(root) {
    if (!this.animDefs.length || !root) return;
    const byName = new Map();
    for (const def of this.animDefs) {
      let la = this.liveAnims.find(x => x.def === def);
      if (!la) { la = { def, materials: [] }; this.liveAnims.push(la); }
      byName.set(def.material, la);
    }
    root.traverse(o => {
      if (!o.isMesh || !o.material || !o.material.map) return;
      const la = byName.get(o.material.name);
      if (la && !la.materials.includes(o.material)) {
        // scrolling needs wrap; the DS materials in question are repeat-flagged,
        // but make sure a clamped sampler doesn't pin the scroll at the edge
        o.material.map.wrapS = o.material.map.wrapT = THREE.RepeatWrapping;
        o.material.map.needsUpdate = true;
        la.materials.push(o.material);
      }
    });
    this.liveAnims = this.liveAnims.filter(x => x.materials.length);
  }

  // Place the course's map objects (the NKM OBJI placements, bound to object GLBs
  // at export time). Each distinct GLB loads once; placements are cheap clones
  // sharing its geometry and materials. Billboards join the per-frame yaw list.
  async _loadObjects(file, gen) {
    let doc;
    try {
      doc = await fetch(BASE + file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !doc.objects) return;

    // one load per distinct GLB
    const protos = new Map();
    for (const o of doc.objects) {
      if (!protos.has(o.file)) {
        protos.set(o.file, new Promise(res =>
          this.loader.load(MODELS + o.file, g => {
            g.scene.traverse(n => {
              if (n.isMesh && n.material && n.material.map) {
                n.material.map.magFilter = THREE.NearestFilter;
                n.material.map.needsUpdate = true;
              }
            });
            res(g.scene);
          }, undefined, () => res(null))));
      }
    }
    await Promise.all(protos.values());
    if (gen !== this.gen) return;

    const group = new THREE.Group();
    const bills = [], movers = [];
    for (const o of doc.objects) {
      const proto = await protos.get(o.file);
      if (!proto) continue;
      const inst = proto.clone();
      inst.position.set(o.pos[0], o.pos[1], o.pos[2]);
      if (o.scale) inst.scale.set(o.scale[0], o.scale[1], o.scale[2]);
      if (o.billboard) {
        bills.push(inst); // yawed toward the camera each frame, up stays world-up
      } else if (o.rot) {
        inst.rotation.order = 'YXZ'; // Y-dominant Euler degrees from the OBJI entry
        inst.rotation.set(o.rot[0] * Math.PI / 180, o.rot[1] * Math.PI / 180, o.rot[2] * Math.PI / 180);
      }
      // Route followers: walk the PATH polyline at constant world speed. The
      // per-object speed lives in engine code we haven't traced, so a plausible
      // shared speed stands in for now.
      const route = o.route != null && doc.routes && doc.routes[o.route];
      if (route && route.points.length >= 2) {
        movers.push(makeMover(inst, route, !!o.billboard));
      }
      group.add(inst);
    }
    group.visible = this.wantObjects;
    this.three.scene.add(group);
    this.objectsGroup = group;
    this.billboards = bills;
    this.movers = movers;
    this._bindAnims(group);
    if (this.hud && this.hudBase) this.hud.textContent = `${this.hudBase} · ${doc.objects.length} objects placed`;
  }

  // Advance a route follower by dt seconds: constant speed along the polyline,
  // wrapping on looped paths and ping-ponging on open ones (the engine's
  // follower semantics), facing the direction of travel.
  _advanceMover(mv, dt) {
    mv.dist += mv.speed * dt;
    let d, dir = 1;
    if (mv.loop) {
      d = mv.dist % mv.total;
    } else {
      d = mv.dist % (2 * mv.total); // out and back
      if (d > mv.total) { d = 2 * mv.total - d; dir = -1; }
    }
    let i = 0;
    while (i < mv.segs.length - 1 && d > mv.segs[i].len) { d -= mv.segs[i].len; i++; }
    const s = mv.segs[i];
    const f = s.len > 0 ? Math.min(d / s.len, 1) : 0;
    mv.inst.position.set(
      s.a[0] + (s.b[0] - s.a[0]) * f,
      s.a[1] + (s.b[1] - s.a[1]) * f,
      s.a[2] + (s.b[2] - s.a[2]) * f,
    );
    if (!mv.billboard) {
      const dx = (s.b[0] - s.a[0]) * dir, dz = (s.b[2] - s.a[2]) * dir;
      if (dx * dx + dz * dz > 1e-8) mv.inst.rotation.y = Math.atan2(dx, dz);
    }
  }

  // Load the "_V" far model as a camera-locked backdrop: full-bright (unlit) so it
  // reads as sky/scenery, depth-test off so it never occludes the track.
  _loadSkybox(file, gen) {
    this.loader.load(BASE + file, (gltf) => {
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
      const doc = await fetch(BASE + file).then(r => r.json());
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
    const { camera, controls } = this.three;
    if (this.driveOn) {
      controls.enabled = false;
      controls.autoRotate = false;
      this.fly.setEnabled(false); // the drive line owns the camera
      this.driveU = 0; this.driveDir = 1;
      // At kart eye height the orbit near plane (size/100 ≈ several GLB units on a
      // big course) would clip the road and tunnel walls right in front of us.
      camera.near = 0.1;
      camera.updateProjectionMatrix();
    } else {
      controls.enabled = true;
      controls.autoRotate = !this.flyWanted;
      this.fly.setEnabled(this.flyWanted);
      if (this.frame) {
        camera.near = this.frame.size / 100;
        camera.updateProjectionMatrix();
        this._orbitFrame();
      }
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

  _disposeObjects() {
    this.billboards = [];
    this.movers = [];
    if (!this.objectsGroup) return;
    this.three.scene.remove(this.objectsGroup);
    // clones share geometry/materials with their prototype; disposing per node is
    // safe (three.js tolerates repeat dispose) and frees everything once
    this.objectsGroup.traverse(o => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (o.material.map) o.material.map.dispose();
        o.material.dispose();
      }
    });
    this.objectsGroup = null;
  }
}
