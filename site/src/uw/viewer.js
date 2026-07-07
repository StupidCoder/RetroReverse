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
    // Unit quad shared by every billboard. Billboards face the camera around the
    // vertical axis only (cylindrical), so they stay upright — a plane we swivel
    // in Y, not a THREE.Sprite (which tilts to fully face the camera).
    this._plane = new THREE.PlaneGeometry(1, 1);

    // Click-to-identify: invisible boxes (one per object) carry the item id;
    // a click raycasts them and shows the id in a small overlay.
    this._pickBox = new THREE.BoxGeometry(1, 1, 1);
    this._pickMat = new THREE.MeshBasicMaterial({ colorWrite: false, depthWrite: false, side: THREE.DoubleSide });
    this._ray = new THREE.Raycaster();
    if (getComputedStyle(el).position === 'static') el.style.position = 'relative';
    const label = document.createElement('div');
    label.style.cssText = 'position:absolute;left:8px;bottom:8px;padding:4px 8px;' +
      'font:12px/1.4 ui-monospace,monospace;color:#cde;background:rgba(0,0,0,.62);' +
      'border:1px solid #2c3444;border-radius:4px;pointer-events:none;display:none;z-index:5;';
    el.appendChild(label);
    this._pickLabel = label;
    let downX = 0, downY = 0;
    renderer.domElement.addEventListener('pointerdown', (e) => { downX = e.clientX; downY = e.clientY; });
    renderer.domElement.addEventListener('pointerup', (e) => {
      if (Math.hypot(e.clientX - downX, e.clientY - downY) <= 5) this._pick(e); // ignore drags
    });

    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      const dt = Math.min(this._clock.getDelta(), 0.1);
      if (this.active === false) return;
      this._animT = (this._animT || 0) + dt;
      this.fly.update(dt);
      controls.update();
      this._updateCreatures();
      this._faceCamera();
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

    // Billboard objects (items, creatures): OBJECTS.GR / CRIT sprites drawn as
    // upright quads that swivel around the vertical axis to face the camera (see
    // _faceCamera). Their exported position is the base on the floor, so lift
    // each by half its height; they share the level's centring.
    const spriteGroup = new THREE.Group();
    spriteGroup.position.set(-c.x, -c.y, -c.z);
    // Textures listed in data.addTex are the ethereal creatures' index-252 "glow"
    // layers: UW brightens the framebuffer there, so draw them with additive
    // blending (and no depth write, so the glow never occludes). Every other
    // sprite — including those creatures' opaque parts (the ghost's black eyes,
    // the shadow's body) — draws normally with alpha-tested cut-outs.
    const addSet = new Set(data.addTex || []);
    const spriteMats = (data.spriteTex || []).map((png, i) => {
      const t = this._texLoader.load(png);
      t.magFilter = THREE.NearestFilter;
      t.colorSpace = THREE.SRGBColorSpace;
      if (addSet.has(i)) {
        return new THREE.MeshBasicMaterial({
          map: t, transparent: true, alphaTest: 0.05, side: THREE.DoubleSide,
          depthWrite: false, blending: THREE.AdditiveBlending,
        });
      }
      return new THREE.MeshBasicMaterial({
        map: t, transparent: true, alphaTest: 0.5, side: THREE.DoubleSide, depthWrite: true,
      });
    });
    const billboards = []; // every quad that faces the camera around Y
    for (const s of data.sprites || []) {
      const m = new THREE.Mesh(this._plane, spriteMats[s.tex]);
      m.position.set(s.pos[0], s.pos[1] + s.h / 2, s.pos[2]);
      m.scale.set(s.w, s.h, 1);
      spriteGroup.add(m);
      billboards.push(m);
    }

    // Creatures are directional (Doom-style) billboards that play their idle
    // animation. Each has eight view cycles; the render loop swaps view by the
    // camera-to-creature angle and heading (so a monster shows its back, side or
    // face as you circle it) and advances the cycle by time so it moves in place.
    const creatures = [];
    for (const c of data.creatures || []) {
      const views = c.views.map((cyc) => cyc.map((d) => ({
        mat: spriteMats[d.tex], w: d.w, h: d.h,
        addMat: d.add >= 0 ? spriteMats[d.add] : null, // index-252 glow, additive
      })));
      const spr = new THREE.Mesh(this._plane, views[0][0].mat);
      spriteGroup.add(spr);
      billboards.push(spr);
      // Ethereal creatures get a second coincident quad for the additive glow; it
      // tracks the same view/frame/transform as the body (see _setCreatureFrame).
      let add = null;
      if (c.translucent) {
        add = new THREE.Mesh(this._plane, views[0][0].addMat || views[0][0].mat);
        add.visible = false;
        spriteGroup.add(add);
        billboards.push(add);
      }
      const rec = {
        spr, add, views, heading: c.heading, base: c.pos, fps: c.fps || 1,
        phase: (creatures.length * 0.37) % 2, // desync creatures' cycles
        view: -1, frame: -1,
      };
      creatures.push(rec);
      this._setCreatureFrame(rec, 0, 0); // place until the first update
    }

    scene.add(spriteGroup);
    this._spriteGroup = spriteGroup;
    this._spriteMats = spriteMats;
    this._creatures = creatures;
    this._billboards = billboards;

    // Invisible click boxes tagged with each object's item id (doors emit two —
    // frame + leaf — with the same id). Exported Pos is the AABB centre.
    const pickGroup = new THREE.Group();
    pickGroup.position.set(-c.x, -c.y, -c.z);
    for (const p of data.picks || []) {
      const box = new THREE.Mesh(this._pickBox, this._pickMat);
      box.position.set(p.pos[0], p.pos[1], p.pos[2]);
      box.scale.set(p.size[0], p.size[1], p.size[2]);
      box.userData.id = p.id;
      pickGroup.add(box);
    }
    scene.add(pickGroup);
    this._pickGroup = pickGroup;
    this._pickLabel.style.display = 'none';

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

  // _setCreatureFrame points a creature sprite at frame `fi` of view cycle
  // `view`, rescaling and re-seating its base on the floor (frames differ in
  // size). No-op when nothing changed so material swaps stay cheap.
  _setCreatureFrame(rec, view, fi) {
    if (view === rec.view && fi === rec.frame) return;
    rec.view = view;
    rec.frame = fi;
    const cyc = rec.views[view] || rec.views[0];
    const f = cyc[fi % cyc.length] || cyc[0];
    rec.spr.material = f.mat;
    rec.spr.scale.set(f.w, f.h, 1);
    rec.spr.position.set(rec.base[0], rec.base[1] + f.h / 2, rec.base[2]);
    // Keep the additive glow quad in lockstep with the body (same frame size and
    // floor-seated position); hide it on frames that carry no index-252 pixels.
    if (rec.add) {
      if (f.addMat) {
        rec.add.material = f.addMat;
        rec.add.scale.set(f.w, f.h, 1);
        rec.add.position.set(rec.base[0], rec.base[1] + f.h / 2, rec.base[2]);
        rec.add.visible = true;
      } else {
        rec.add.visible = false;
      }
    }
  }

  // _updateCreatures picks each creature's view from the horizontal bearing
  // between it and the camera — the fine angle (0-31) of the camera→creature
  // vector is the cameraAngle and view = REMAP[(camAngle - heading*4) & 31] —
  // then advances the idle cycle by elapsed time (each creature phase-offset so
  // they don't animate in lockstep).
  _updateCreatures() {
    const list = this._creatures;
    if (!list || !list.length) return;
    const cam = this.three.camera.position;
    const wp = this._wp || (this._wp = new THREE.Vector3());
    const t = this._animT || 0;
    for (const rec of list) {
      rec.spr.getWorldPosition(wp);
      const dx = wp.x - cam.x, dz = wp.z - cam.z; // camera → creature (horizontal)
      // Match heading's fine-angle convention: heading h faces world such that
      // atan2(-fx, -fz) == h*4 steps; use the same mapping for the bearing.
      // (camAngle - heading*4), not the game formula's (heading*4 - camAngle):
      // our world's rotational sense runs opposite to UW's fine-angle count, so
      // this sign reverses the turn direction while keeping back (view 0) and
      // front (view 4) fixed — only the two sides swap (1<->7, 2<->6, 3<->5).
      const camAngle = Math.round((Math.atan2(-dx, -dz) / (2 * Math.PI)) * 32) & 31;
      const view = VIEW_REMAP[(camAngle - rec.heading * 4 + 32) & 31];
      const cyc = rec.views[view] || rec.views[0];
      const fi = Math.floor((t + rec.phase) * rec.fps) % cyc.length;
      this._setCreatureFrame(rec, view, fi);
    }
  }

  // _faceCamera swivels every billboard quad around the vertical axis so its
  // face points at the camera horizontally — cylindrical billboarding. Unlike a
  // THREE.Sprite the quad never tilts, so sprites stay upright when you look up
  // or down. The quad's default normal is +Z, so rotation.y = atan2(dx, dz)
  // aims +Z from the quad toward the camera.
  _faceCamera() {
    const list = this._billboards;
    if (!list || !list.length) return;
    const cam = this.three.camera.position;
    const wp = this._wp || (this._wp = new THREE.Vector3());
    for (const m of list) {
      m.getWorldPosition(wp);
      m.rotation.y = Math.atan2(cam.x - wp.x, cam.z - wp.z);
    }
  }

  // _pick raycasts the invisible object boxes at the clicked point and shows the
  // item id of the nearest one (the boxes share the level's centring group).
  _pick(e) {
    const grp = this._pickGroup;
    if (!grp) return;
    const rect = this.three.renderer.domElement.getBoundingClientRect();
    const ndc = this._ndc || (this._ndc = new THREE.Vector2());
    ndc.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
    ndc.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;
    this._ray.setFromCamera(ndc, this.three.camera);
    const hits = this._ray.intersectObjects(grp.children, false);
    if (hits.length) {
      this._pickLabel.textContent = `item id ${hits[0].object.userData.id}`;
      this._pickLabel.style.display = 'block';
    } else {
      this._pickLabel.style.display = 'none';
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
    this._billboards = null;
    if (this._pickGroup) {
      this.three.scene.remove(this._pickGroup);
      this._pickGroup = null;
    }
  }
}
