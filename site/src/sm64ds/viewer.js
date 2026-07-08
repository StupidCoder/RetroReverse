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
import { installPicker } from '../shared/objinfo.js';
import { clone as cloneSkinned } from 'three/addons/utils/SkeletonUtils.js';

// Format-2 asset tree. Every GLB path inside the JSON (models/…, collision/…)
// is already root-relative to this base, so load them as BASE + path.
const BASE = 'public/super-mario-64-ds/';
// Bare stem of a "models/foo.glb" / "collision/foo.glb" path — the key our
// per-model behavior tables (PATROL, COIN_MODELS, objectInfo, …) are keyed on.
const stemOf = (p) => (p ? p.replace(/^.*\//, '').replace(/\.glb$/, '') : null);

// Clip choice: clip names end in walk/wait/run + an optional number
// (bombking_walk1, kuribo_wait). Patrolling actors lead with their walk;
// stationary ones with their idle.
const pickClip = (anims, prefer) => {
  for (const kind of prefer) {
    const hit = anims.find(a => new RegExp('_' + kind + '\\d*$').test(a.name));
    if (hit) return hit;
  }
  return anims[0];
};

// Coin spin, from the coin actor's step at $020B23A4: yaw += $C00 angle units
// per frame ($10000 = 360°), at the game's 30 fps actor tick.
const COIN_SPIN = (0xC00 / 0x10000) * 30 * 2 * Math.PI; // rad/s
const COIN_MODELS = new Set(['arc0_5', 'arc0_7']);

// Signposts (obj_tatefuda, actor 184): the traced behavior is a proximity
// dialog — the engine sets an in-range flag in the actor's +$B0 word and a
// button press starts the sign's message ($020BB060). The placement JSON
// carries each sign's decoded English text (txt), quoted in the click card.
const SIGN_MODEL = 'obj_tatefuda';

// Goomba wander AI, traced from daKrb_c's state-0 handler ($0212ABD4/$0212AE98,
// overlay 84): walk speed 2.0 world-units/frame (state table $02130248), yaw
// eases toward a target heading at $200 angle-units/frame, a 100-frame timer
// repicks — 75% turn by a random signed 16-bit angle, 25% pause — and a
// 1000-unit leash around the spawn point steers it home. Stage-GLB units are
// world/1000 and the actor tick is 30 fps.
// Bob-omb wander, traced from daBmb_c (overlay 102, $0214BE1C/$0214BEB4/
// $0214C0B8): walk 5.0 world-units/frame (set on each heading pick), turn
// $400 angle-units/frame, target heading = angle-to-home + a random signed
// 16-bit offset (home-biased, erratic), straight home beyond 1280 units,
// repick on reaching the heading or a 512-frame fallback timer.
const PATROL = {
  kuribo_model: {
    speed: (0x8000 / 4096) * 30 / 1000,          // stage units/s
    turn: (0x200 / 0x10000) * 2 * Math.PI * 30,  // rad/s
    repick: 100 / 30,                             // s
    leash: 1.0,                                   // stage units
    pauseChance: 0.25, homeBias: false, repickOnArrive: false,
  },
  bombhei: {
    speed: (0x5000 / 4096) * 30 / 1000,
    turn: (0x400 / 0x10000) * 2 * Math.PI * 30,
    repick: 512 / 30,
    leash: 1.28,
    pauseChance: 0, homeBias: true, repickOnArrive: true,
  },
};
PATROL.red_bombhei = PATROL.bombhei;

// Chain Chomp (daWanwan2_c, overlay 100): anchored to its stake by a 250-unit
// chain (the chain drawer builds from a (0,0,-250) vector, $021437D4) and
// lunges at $17000 = 23.0 world-units/frame ($02143E4C). The chain renders as
// links (ar1_1) strung from the stake to the body. Lunge cadence approximated;
// radius and speed are the traced values.
const CHOMP = {
  radius: 250 / 1000,                       // stage units
  lunge: (0x17000 / 4096) * 30 / 1000,      // stage units/s
  links: 5,
};
// Editorial per-actor behaviour notes (the shared objectInfo format: { key: {title, text} },
// keyed by the model stem), resolved by the shared picker into the InfoCard.
const objectInfo = {
  kuribo_model: { title: 'Goomba — daKrb_c, actor 202 (trio spawners: actors 200/201)', text:
    'Wander AI (overlay 84): forward speed at +$98 eases toward the per-state table at $02130248 — 2.0 world-units/frame wandering, 8.0 chasing — by $500 per frame. Yaw eases toward a target heading at $200 angle-units/frame ($10000 = 360°). A 100-frame timer repicks: the shared RNG ($0203B990) turns it by a random signed 16-bit angle three times out of four and pauses it the fourth. Wall contact reflects the heading, a 1000-unit leash around the spawn point (+$41C) steers it home, and falling out teleports it back. Chase (state 3) triggers from the profile\u2019s 100-unit sight radius. Walk clip: kuribo_walk.bca — 3 bones, 30 frames, 16-key rotation tracks.' },
  bombhei: { title: 'Bob-omb — daBmb_c, actor 206', text:
    'Wander AI (overlay 102): each heading pick ($0214BEB4) aims AT its home anchor (+$3C4) plus a random signed 16-bit offset — erratic but home-biased — and beyond 1280 units the randomness is dropped and it walks straight back. Every pick sets forward speed to $5000 (5.0 units/frame). It repicks when the yaw reaches the target or a 512-frame fallback timer (+$3E8) expires; yaw eases at $400 angle-units/frame, doubled to $800 when chasing (speed goal $10000 = 16.0 — the lit-fuse sprint). The walk-clip rate (+$35C) is speed/8, so the feet match the ground. The round body is a billboard bone (flag +$3C bit 0, \u201cbody_bill\u201d).' },
  red_bombhei: { title: 'Bob-omb Buddy — daRedBombhei_c, actor 181', text:
    'Shares the bob-omb wander mechanics (its bank sits in overlay 84): home-biased random headings, 5.0 units/frame walk, $400/frame turning, billboarded body. The buddies never arm a fuse chase.' },
  ar1_2: { title: 'Chain Chomp \u2014 daWanwan_c (actor 219) / daWanwan2_c (actor 337)', text:
    'Two variants share the model (castle-grounds archive: ar1 member 2 \u2014 a_mat_body / a_mat_eye / a_mat_mouth; member 1 is the chain link). The placed Bob-omb Battlefield chomp is daWanwan_c in the level\u2019s own overlay 14: its init spawns its STAKE \u2014 a pile, actor 27, param $11 \u2014 at its own anchor ($02112C6C: MOV r0,#27 \u2192 BL $02010E2C), keeps it at +$608 and flips the pile\u2019s chomp-stake mode byte (+$320); pound it three times and the chomp breaks the gate. daWanwan2_c (overlay 100, the star-3 and castle-grounds chomp) hops under \u2212$3C000 (\u221215.0/frame) gravity, eases its scale vector toward 1.0 (the pre-bark inflate), and lunges at $17000 = 23.0 units/frame \u2014 the fastest traced motion in the game; its chain drawer ($021437D4) strings the links from a (0, 0, \u2212250) anchor vector. The viewer\u2019s lunge cadence is approximated; radius and speed are the traced values.' },
  pile: { title: 'Pile / the Chomp\u2019s stake \u2014 actor 27', text:
    'A poundable wooden post with its own scaled-down collision mesh (pile.kcl, authored ~10\u00d7 and drawn at 0.1). Bob-omb Battlefield places four in a square \u2014 and a fifth is not placed at all: the chain chomp\u2019s init spawns it at the chomp\u2019s anchor with param $11 and marks its stake mode byte (+$320). Stomp that one three times and the chomp tears down the gate.' },
  arc0_5: { title: 'Coin — actors 288/289/290 (also item actor 276, subtypes $B/$C)', text:
    'The step at $020B2324 adds $C00 to the yaw at +$8E every frame — $10000 is a full turn, so about 1.4 revolutions per second at the 30 fps actor tick. The spin is geometry, not a texture animation: a flat quad yawing in 3D, its 16\u00d764 texture mapped at texScaleS 2.0 over mirrored-S addressing. The blob shadow joins the draw list only within 100 render units of the camera.' },
  arc0_7: { title: 'Red Coin — actor 289', text:
    'Same spin and pickup path as the yellow coin ($C00 yaw per frame at +$8E); the model is arc0 member 7, the red-paletted variant of member 5.' },
  obj_tatefuda: { title: 'Signpost — daObjTatefuda_c, actor 184', text:
    'A proximity dialog: at init ($020BC240) it snaps to the ground with a collision ray cast from y+$64000 downward, then registers an interaction cylinder. Its step ($020BBEA4) watches the engine-set flags at +$B0 — bit $4000 means the player is in range — and a button press starts its message through $020BB060. Click a signpost to read its actual in-game text (decoded from the BMG message file).' },
};

export class ModelViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    window.__smv = this; // debug handle for the headless-probe workflow
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
    this.placed = [];       // every placed object {obj, actor, model} (click for its card)
    this.mixers = [];       // skinned enemies playing their .bca walk clips
    this.patrollers = [];   // goombas wandering per their traced AI
    this.chomps = [];       // chain chomps lunging on their chains
    this.bbBones = [];      // billboard bones (bmd flag +$3C bit 0): face the camera
    this.wantObjects = true;
    // collision overlay: the level's .kcl mesh + each placed actor's own
    // collider (Part VI), reconstructed from the prism records
    this.collisionGroup = null;
    this.wantCollision = false;
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
      // The skybox rides the camera: the engine's frame draw copies the camera
      // position into the sky wrapper's matrix translation every frame
      // ($0202B940) and never rotates it.
      if (this.sky) this.sky.position.copy(camera.position);
      // Billboard trees yaw around world-up toward the camera; up stays vertical
      // (the DS renders these flat quads camera-facing).
      for (const b of this.billboards) {
        b.rotation.y = Math.atan2(camera.position.x - b.position.x, camera.position.z - b.position.z);
      }
      for (const s of this.spinners) s.rotation.y += COIN_SPIN * dt;
      for (const mx of this.mixers) mx.update(dt);
      for (const c of this.chomps) {
        c.t -= dt;
        if (c.t <= 0) {
          // pick a lunge target on the chain circle, burst toward it
          c.t = 2.5 + Math.random() * 3;
          const a = Math.random() * 2 * Math.PI;
          c.tx = c.home.x + Math.sin(a) * CHOMP.radius;
          c.tz = c.home.z + Math.cos(a) * CHOMP.radius;
          c.lunging = 0.35;
        }
        c.obj.position.y = c.home.y + c.lift;
        const dx = c.tx - c.obj.position.x, dz = c.tz - c.obj.position.z;
        const dist = Math.hypot(dx, dz);
        const sp = c.lunging > 0 ? CHOMP.lunge : CHOMP.lunge * 0.12;
        if (c.lunging > 0) c.lunging -= dt;
        if (dist > 1e-4) {
          const step = Math.min(sp * dt, dist);
          c.obj.position.x += dx / dist * step;
          c.obj.position.z += dz / dist * step;
          c.obj.rotation.y = Math.atan2(dx, dz);
        }
        // chain: string the links from the stake to the body
        for (let i = 0; i < c.links.length; i++) {
          const f = (i + 1) / (c.links.length + 1);
          c.links[i].position.set(
            c.home.x + (c.obj.position.x - c.home.x) * f,
            c.home.y + c.lift * f + Math.sin(Math.PI * f) * -0.004,
            c.home.z + (c.obj.position.z - c.home.z) * f);
        }
      }
      for (const g of this.patrollers) {
        const P = g.cfg;
        const hx = g.home.x - g.obj.position.x, hz = g.home.z - g.obj.position.z;
        g.timer -= dt;
        let d = ((g.target - g.yaw + 3 * Math.PI) % (2 * Math.PI)) - Math.PI;
        const arrived = P.repickOnArrive && Math.abs(d) < 0.02;
        if (g.timer <= 0 || arrived) {
          g.timer = P.repick;
          if (Math.random() < P.pauseChance) g.paused = true;
          else {
            g.paused = false;
            const spread = (Math.random() * 2 - 1) * Math.PI;
            g.target = P.homeBias ? Math.atan2(hx, hz) + spread : g.target + spread;
          }
          d = ((g.target - g.yaw + 3 * Math.PI) % (2 * Math.PI)) - Math.PI;
        }
        // leash: outside the home radius, head straight back
        if (hx * hx + hz * hz > P.leash * P.leash) {
          g.target = Math.atan2(hx, hz);
          d = ((g.target - g.yaw + 3 * Math.PI) % (2 * Math.PI)) - Math.PI;
        }
        const step = P.turn * dt;
        g.yaw += Math.abs(d) <= step ? d : Math.sign(d) * step;
        g.obj.rotation.y = g.yaw;
        if (!g.paused) {
          g.obj.position.x += Math.sin(g.yaw) * P.speed * dt;
          g.obj.position.z += Math.cos(g.yaw) * P.speed * dt;
        }
      }
      // Billboard bones (the bob-omb's body): after the clips AND the behavior
      // rotations settle, override the bone's world orientation to face the
      // camera — the engine does the same to its billboard-flagged bones at
      // draw time. getWorldQuaternion refreshes the ancestor matrices, so the
      // compensation composes against this frame's final pose.
      if (this.bbBones.length) {
        camera.getWorldQuaternion(this._camQ || (this._camQ = new THREE.Quaternion()));
        const pq = this._parQ || (this._parQ = new THREE.Quaternion());
        for (const b of this.bbBones) {
          b.parent.getWorldQuaternion(pq).invert();
          b.quaternion.copy(pq).multiply(this._camQ);
        }
      }
      renderer.render(scene, camera);
    };
    tick();

    // Placement interaction via the SHARED picker: a click raycasts the placed objects and shows
    // the actor id + bound model + the editorial behaviour notes (objectInfo), or the object's own
    // words (signposts). The placed set changes per level, so pass a getter; gate on the objects
    // layer. The picker composes the card title itself via `resolve`.
    this._picker = installPicker(
      { el: this.el, canvas: renderer.domElement, camera: this.three.camera },
      () => this.placed.map(c => ({ object3d: c.obj, obj: c })),
      {
        enabled: () => this.wantObjects !== false,
        resolve: (c) => {
          const ed = c.model && objectInfo[c.model];
          return {
            title: `Actor ${c.actor}` + (c.model ? ` — ${c.model}` : ' — no model bound'),
            subtitle: ed ? ed.title : undefined,
            body: ed ? ed.text : (c.model
              ? 'Placement decoded from the level overlay; model bound by the actor oracle.'
              : 'The actor oracle recorded no model load in this actor’s create/init.'),
            muted: !ed,
            // A signpost's in-game words: par1 is a message id resolved through the game's
            // $0208EEEC range table into the English BMG file (dialog font, doc Part VIII).
            quote: c.txt,
          };
        },
      });
  }

  _hideSign() {
    this._picker?.hide();
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
    // The browse list is the format-2 manifest: its levels (Levels section,
    // level-kind entries that resolve to a level envelope) followed by its
    // plain model GLBs, grouped by their section (Characters/Enemies/Objects/
    // Skyboxes) exactly as before.
    const manifest = await fetch(BASE + 'manifest.json').then(r => r.json());
    this.manifest = manifest;
    const levels = (manifest.levels || []).map(l => ({
      name: l.name,
      section: l.section || 'Levels',
      kind: 'level',
      file: l.file,          // the level envelope json (levels/<stem>.json)
      objects: l.objects,    // present => the studio shows the object/collision layer toggles
    }));
    const models = (manifest.models || []).map(m => ({
      name: m.name,
      section: m.section,
      kind: 'model',
      file: m.file,          // models/<stem>.glb, root-relative to BASE
    }));
    this.models = levels.concat(models);
    return this.models;
  }

  // The studio drives this: "objects" shows/hides the placed level objects,
  // "collision" the red .kcl overlay.
  setLayer(id, on) {
    if (id === 'objects') {
      this.wantObjects = on;
      if (this.objectsGroup) this.objectsGroup.visible = on;
    }
    if (id === 'collision') {
      this.wantCollision = on;
      if (this.collisionGroup) this.collisionGroup.visible = on;
    }
  }

  loadModel(i) {
    const m = this.models[i];
    if (!m) return;
    const gen = ++this.gen;
    if (m.kind === 'level') this._loadLevel(m, gen);
    else this._loadPlainModel(m, gen);
  }

  // Install a loaded GLB as the main scene group: keep DS textures crisp, play
  // an enemy's first clip, frame the camera, and switch fly/orbit controls.
  // Returns the group's bounding size (stage units) for downstream fitting.
  _installMain(gltf, isLevel, name) {
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
    group.traverse(o => { if (o.userData && o.userData.billboard) this.bbBones.push(o); });
    if (gltf.animations && gltf.animations.length) {
      const mx = new THREE.AnimationMixer(group);
      const clip = pickClip(gltf.animations, ['walk', 'run', 'wait']);
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
    controls.autoRotate = !isLevel;
    this.fly.setScale(size);
    this.fly.setEnabled(isLevel);

    if (this.hud) {
      this.hud.textContent = `${name} — ${tris.toLocaleString()} triangles, textures as shipped on cartridge` +
        (isLevel ? ` · ${flyHint}` : '');
    }
    return size;
  }

  // A plain model entry: one GLB, root-relative to BASE.
  _loadPlainModel(m, gen) {
    this.loader.load(BASE + m.file, (gltf) => {
      if (gen !== this.gen) return; // superseded
      this._installMain(gltf, m.section === 'Levels', m.name);
    });
  }

  // A level entry: fetch the level envelope, load its stage mesh, then its
  // object placements (which also bring in the skybox + Mario spawn from the
  // envelope). mesh.glb / sky / objects paths are all root-relative to BASE.
  async _loadLevel(m, gen) {
    let level;
    try {
      level = await fetch(BASE + m.file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !level || !level.mesh || !level.mesh.glb) return;
    this.loader.load(BASE + level.mesh.glb, (gltf) => {
      if (gen !== this.gen) return;
      const size = this._installMain(gltf, true, level.name || m.name);
      if (level.objectsFile) {
        this._loadObjects('levels/' + level.objectsFile, gen, size, level);
      }
    });
  }

  // Place the level's objects (decoded from the level overlay's object tables).
  // Placements bound to an extracted model load it once and clone per instance;
  // the rest show as small markers — their models aren't extracted yet.
  async _loadObjects(file, gen, size, level) {
    let doc;
    try {
      doc = await fetch(BASE + file).then(r => r.json());
    } catch { return; }
    if (gen !== this.gen || !doc.objects) return;

    // Prototypes are keyed by bare stem; the full "models/<stem>.glb" path
    // comes straight off each placement's `model` field.
    const protos = new Map();
    const wantProto = new Map(); // stem -> full glb path (root-relative to BASE)
    for (const o of doc.objects) if (o.model) wantProto.set(stemOf(o.model), o.model);
    if (wantProto.has('ar1_2')) wantProto.set('ar1_1', 'models/ar1_1.glb'); // chomp brings its chain
    for (const [stem, path] of wantProto) {
      if (!protos.has(stem)) {
        protos.set(stem, new Promise(res =>
          this.loader.load(BASE + path, g => {
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
      const props = o.props || {};
      const model = o.model || null;            // full "models/<stem>.glb" path
      const stem = stemOf(model);               // behavior-table key
      const pos = o.pos;
      const ry = (o.rot && o.rot[1]) || 0;      // yaw, degrees
      // Scale rides with the model: default to the engine's GLB/125 only when a
      // model is bound but the exporter emitted no explicit scale.
      const scale = props.scale != null ? props.scale : (model ? 1 / 125 : null);
      let inst;
      if (model) {
        const proto = await protos.get(stem);
        if (proto) {
          // Skinned models (the enemies) clone with their skeletons and play
          // their .bca walk cycle (decoded from the cartridge, 30 fps).
          if (proto.animations.length) {
            inst = cloneSkinned(proto.scene);
            inst.traverse(n => { if (n.userData && n.userData.billboard) this.bbBones.push(n); });
            const clip = pickClip(proto.animations,
              PATROL[stem] ? ['walk', 'run', 'wait'] : ['wait', 'walk', 'run']);
            const mx = new THREE.AnimationMixer(inst);
            mx.clipAction(clip).play();
            // desync instances so a troop doesn't march in lockstep
            mx.setTime(Math.random() * clip.duration);
            this.mixers.push(mx);
          } else {
            inst = proto.scene.clone();
          }
          if (scale != null) inst.scale.setScalar(scale);
          if (props.billboard) bills.push(inst);
          else if (ry) inst.rotation.y = ry * Math.PI / 180;
          if (COIN_MODELS.has(stem)) spinners.push(inst);
          if (stem === SIGN_MODEL) signs.push(inst);
          if (stem === 'ar1_2') { // chain chomp: anchored, with chain links
            const links = [];
            const linkProto = await protos.get('ar1_1');
            if (linkProto) {
              for (let i = 0; i < CHOMP.links; i++) {
                const l = linkProto.scene.clone();
                l.scale.setScalar(scale != null ? scale : 1 / 125);
                group.add(l);
                links.push(l);
                bills.push(l); // the link disc is a single billboard-flagged bone
              }
            }
            // The body model's pivot is its centre; the engine's gravity rests
            // it on the ground, so lift by the model's half-depth.
            const cbox = new THREE.Box3().setFromObject(proto.scene);
            const lift = -cbox.min.y * (scale != null ? scale : 1 / 125);
            this.chomps.push({
              obj: inst, links, lift, t: Math.random() * 2,
              tx: pos[0], tz: pos[2],
              home: { x: pos[0], y: pos[1], z: pos[2] },
            });
          }
          if (PATROL[stem]) {
            const yaw = ry * Math.PI / 180;
            this.patrollers.push({
              obj: inst, yaw, target: yaw, paused: false, cfg: PATROL[stem],
              timer: Math.random() * PATROL[stem].repick,
              home: { x: pos[0], z: pos[2] },
            });
          }
          placed++;
        }
      }
      if (!inst) {
        inst = new THREE.Mesh(markerGeo, markerMat);
        markers++;
      }
      inst.position.set(pos[0], pos[1], pos[2]);
      group.add(inst);
      this.placed.push({ obj: inst, actor: o.actor, model: stem, txt: props.text || null });
    }
    group.visible = this.wantObjects;
    this.three.scene.add(group);
    this.objectsGroup = group;
    this._loadCollision(doc, gen, level);

    // Mario at the level's first entrance (type-1 entry), idling (su_wait).
    // The spawn now lives on the level envelope: spawn.pos + spawn.rot (degrees).
    const spawn = level && level.spawn;
    if (spawn && spawn.pos) {
      this.loader.load(BASE + 'models/mario_model_mg.glb', g => {
        if (gen !== this.gen) return;
        g.scene.traverse(n => {
          if (n.isMesh && n.material && n.material.map) {
            n.material.map.magFilter = THREE.NearestFilter;
            n.material.map.needsUpdate = true;
          }
          if (n.isSkinnedMesh) n.frustumCulled = false;
        });
        const inst = cloneSkinned(g.scene);
        inst.scale.setScalar(1 / 125);
        inst.position.set(spawn.pos[0], spawn.pos[1], spawn.pos[2]);
        inst.rotation.y = (spawn.rot || 0) * Math.PI / 180;
        const clip = g.animations.find(a => a.name === 'su_wait') || g.animations[0];
        if (clip) {
          const mx = new THREE.AnimationMixer(inst);
          mx.clipAction(clip).play();
          this.mixers.push(mx);
        }
        group.add(inst);
      });
    }

    // Skybox (settings +$18 -> vrbox table $02075620): camera-centred, drawn
    // behind everything at the engine's NULL-scale object size (GLB/125). The
    // sky now lives on the level envelope as a full "models/<vr>.glb" path.
    if (level && level.sky) {
      this.loader.load(BASE + level.sky, g => {
        if (gen !== this.gen) return;
        g.scene.traverse(n => {
          if (n.isMesh) {
            n.frustumCulled = false;
            n.renderOrder = -1000;
            if (n.material) {
              n.material.depthWrite = false;
              n.material.depthTest = false;
              if (n.material.map) {
                n.material.map.magFilter = THREE.NearestFilter;
                n.material.map.needsUpdate = true;
              }
            }
          }
        });
        // The dome only provides direction (camera-centred, drawn behind
        // everything with depth off), so its absolute size is free — the
        // engine's GLB/125 would cross our far plane on small stages. Fit it
        // well inside the frustum instead.
        const r = new THREE.Box3().setFromObject(g.scene).getSize(new THREE.Vector3()).length() / 2;
        g.scene.scale.setScalar((this.three.camera.far * 0.45) / (r || 1));
        this.three.scene.add(g.scene);
        this.sky = g.scene;
      });
    }

    this.billboards = bills;
    this.spinners = spinners;
    this.signposts = signs;
    if (this.hud) {
      this.hud.textContent += ` · ${placed + markers} objects placed (${markers} as markers)` +
        ' · click any object for its actor ID (markers = actors the oracle bound no model to)';
    }
  }

  // Collision overlay (Part VI): the level's .kcl mesh plus each placed
  // actor's own collider (the .kcl file that actor's init loaded, per the
  // binding oracle), reconstructed from the prism records and baked red.
  // Exported in stage-GLB units: the level map sits at the origin, object
  // colliders at their placement with only the placement yaw.
  async _loadCollision(doc, gen, level) {
    // Collider GLB paths are full ("collision/<stem>.glb") and used as the map
    // key. The stage's OWN collision mesh is an optional `collision` field on
    // the level envelope; format-2 does not currently export it (see GAP note),
    // so this reads it only when present and otherwise the stage collider is
    // gracefully absent.
    const paths = new Set();
    const stageCol = level && level.collision;
    if (stageCol) paths.add(stageCol);
    for (const o of doc.objects) if (o.collision) paths.add(o.collision);
    if (!paths.size) return;
    const protos = new Map();
    await Promise.all([...paths].map(p => new Promise(res =>
      this.loader.load(BASE + p, g => {
        g.scene.traverse(n => {
          if (n.isMesh && n.material) {
            // walls are seen from both sides, and collision faces are often
            // coplanar with the render mesh — draw double-sided and pull the
            // overlay toward the camera so it wins the depth test
            n.material.side = THREE.DoubleSide;
            n.material.polygonOffset = true;
            n.material.polygonOffsetFactor = -1;
            n.material.polygonOffsetUnits = -1;
          }
        });
        protos.set(p, g.scene);
        res();
      }, undefined, () => res()))));
    if (gen !== this.gen) return;
    const group = new THREE.Group();
    if (stageCol && protos.has(stageCol)) group.add(protos.get(stageCol).clone());
    for (const o of doc.objects) {
      if (!o.collision || !protos.has(o.collision)) continue;
      const props = o.props || {};
      const pos = o.pos;
      const ry = (o.rot && o.rot[1]) || 0;
      // outer node = the placement pose; inner node = the collider's own
      // registered transform (the MtxFx43 the oracle read at +$134 — a
      // row-vector matrix, so its 3x3 transposes into three.js's column
      // convention; object .kcl are authored ~10x and scaled down here)
      const outer = new THREE.Group();
      outer.position.set(pos[0], pos[1], pos[2]);
      if (ry) outer.rotation.y = ry * Math.PI / 180;
      const inst = protos.get(o.collision).clone();
      if (props.colMtx) {
        const a = props.colMtx;
        const m = new THREE.Matrix4().set(
          a[0], a[3], a[6], a[9],
          a[1], a[4], a[7], a[10],
          a[2], a[5], a[8], a[11],
          0, 0, 0, 1);
        if (props.colScaleY) m.multiply(new THREE.Matrix4().makeScale(1, props.colScaleY, 1));
        inst.matrixAutoUpdate = false;
        inst.matrix.copy(m);
      }
      outer.add(inst);
      group.add(outer);
    }
    group.visible = this.wantCollision;
    this.three.scene.add(group);
    this.collisionGroup = group;
  }

  _disposeObjects() {
    if (this.sky) {
      this.three.scene.remove(this.sky);
      this.sky = null;
    }
    if (this.collisionGroup) {
      this.three.scene.remove(this.collisionGroup);
      this.collisionGroup.traverse(o => {
        if (o.geometry) o.geometry.dispose();
        if (o.material) o.material.dispose();
      });
      this.collisionGroup = null;
    }
    this.billboards = [];
    this.spinners = [];
    this.signposts = [];
    this.placed = [];
    this.mixers = [];
    this.patrollers = [];
    this.chomps = [];
    this.bbBones = [];
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
