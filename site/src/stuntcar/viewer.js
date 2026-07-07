// Stunt Car Racer — track level viewer. The geometry is entirely the engine's own,
// decoded purely from the disk in Go (package track, Part IV) and verified against the
// original on our m68k core. Each track is the game's own baked polygon model (the
// race-setup bake $65BEC, reimplemented in Go and verified byte-exact by cmd/modeloracle):
// per rung, the piece-shape (x,z) vertex pairs read by $5C6C4 rotated by the section
// quadrant, placed at the section's 16x16 grid cell exactly as the game's per-frame draw
// does (one cell = $800 units) — world = cell*$800 + local. No fit, no spline, no
// accumulation: straights are straight, arcs are the real arcs, and joints meet because
// the data says so. Heights are the full-precision rail heights (their difference is the
// real camber); rung LINES are drawn only where the game's decimated model has a polygon
// edge. Rendered as a hidden-line wireframe (invisible depth fill + colour LineSegments,
// the Marble Madness slope-viewer technique).
//
// Each circuit is presented as a level: you fly through it with the shared FlyCam (WASD
// to move relative to the view, arrow keys to look; mouse drag still orbits), the same
// free-flight controls the other 3-D level viewers use.
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

export class TrackViewer {
  constructor(el, hud) {
    this.el = el;
    this.hud = hud;
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.setClearColor(0x0a0d12, 1);
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.01, 200);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;

    this.three = { renderer, scene, camera, controls, group: null };
    // Free-flight camera (WASD move / arrow look), layered on the orbit controls. The
    // Studio's KeyboardCamera checks v.fly.enabled and cedes the arrow keys to it.
    this.fly = new FlyCam(camera, controls, el);
    this._clock = new THREE.Clock();
    this._resize();
    window.addEventListener('resize', () => this._resize());
    new ResizeObserver(() => this._resize()).observe(el);

    const tick = () => {
      requestAnimationFrame(tick);
      if (this.active === false) return; // paused while another viewer is shown
      const dt = Math.min(0.05, this._clock.getDelta());
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

  // The level list: the eight decoded circuits (public/stuntcar/tracks.json). Fetched
  // and cached here so the Studio treats the viewer like every other level viewer.
  async init() {
    this.levels = await fetch('public/stuntcar/tracks.json').then(r => r.json());
    return this.levels;
  }

  // Load one track as a level. track: { name, sections, finishIdx, nodes, rungs }; id = 0..7
  loadLevel(track, id) {
    this.trackId = id;
    const t = this.three;
    if (t.group) { t.scene.remove(t.group); disposeGroup(t.group); }
    const group = new THREE.Group();

    // track.rungs[i][k] = [Lx,Lz,Rx,Rz,Lh,Rh,flags]: the game's own baked polygon
    // model in ABSOLUTE plan coordinates (cell*$800 + $5C6C4 local vertex), verified
    // byte-exact against the engine's bake $65BEC (cmd/modeloracle). flags: 1 = a rung
    // the game draws, 2 = hidden gap-piece end (not in the game's model), 4 = crease,
    // 8 = the finish line. Rung 0 of every section duplicates the previous section's
    // last rung (the bake skips it too), and hidden rungs are dropped exactly as the
    // engine drops them from its record strip.
    const raw = [];
    const n = track.rungs.length;
    for (let i = 0; i < n; i++) {
      const rs = track.rungs[i];
      for (let k = 0; k < rs.length; k++) {
        if (k === 0 || (rs[k][6] & 2)) continue;
        raw.push({ v: rs[k], sec: i });
      }
    }
    let minX = Infinity, maxX = -Infinity, minZ = Infinity, maxZ = -Infinity, minH = Infinity;
    for (const { v } of raw) {
      minX = Math.min(minX, v[0], v[2]); maxX = Math.max(maxX, v[0], v[2]);
      minZ = Math.min(minZ, v[1], v[3]); maxZ = Math.max(maxZ, v[1], v[3]);
      minH = Math.min(minH, v[4], v[5]);
    }
    const cx = (minX + maxX) / 2, cz = (minZ + maxZ) / 2;
    const span = Math.max(maxX - minX, maxZ - minZ) || 1;
    const S = 8 / span; // fit the plan into ~8 units

    // Fixed rail-height -> plan-unit ratio shared across all tracks so relative relief
    // is honest: the Roller Coaster and Ski Jump really do tower over the gentle
    // circuits. Heights are the engine's full-precision rail heights (one grid cell =
    // $800 plan units; 4800 height units per cell-width of rise reads well); they sit
    // on the ground via the per-track minimum.
    const HK = S * 2048 / 4800;

    const rings = raw.map(({ v, sec }) => ({
      l: { x: (v[0] - cx) * S, z: (v[1] - cz) * S },
      r: { x: (v[2] - cx) * S, z: (v[3] - cz) * S },
      hl: (v[4] - minH) * HK, hr: (v[5] - minH) * HK,
      fl: v[6], sec,
    }));
    const m = rings.length;
    const V = (p, y) => new THREE.Vector3(p.x, y, p.z);

    // Invisible depth fill (the ribbon surface) for hidden-line removal.
    const fpos = [];
    const quad = (a, b, c, d) => fpos.push(a.x, a.y, a.z, b.x, b.y, b.z, c.x, c.y, c.z, b.x, b.y, b.z, d.x, d.y, d.z, c.x, c.y, c.z);
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m];
      quad(V(a.l, a.hl), V(a.r, a.hr), V(b.l, b.hl), V(b.r, b.hr));
    }
    const fgeom = new THREE.BufferGeometry();
    fgeom.setAttribute('position', new THREE.Float32BufferAttribute(fpos, 3));
    const fill = new THREE.Mesh(fgeom, new THREE.MeshBasicMaterial({
      colorWrite: false, polygonOffset: true, polygonOffsetFactor: 1, polygonOffsetUnits: 1,
      side: THREE.DoubleSide,
    }));
    group.add(fill);

    // Wireframe: the two rails (coloured along the lap) + rung lines exactly where the
    // game's baked model has a polygon edge (flags bit 0); the finish rung is gold.
    const lpos = [], lcol = [];
    const col = new THREE.Color();
    const edge = (p, q, f, gold) => {
      if (gold) col.setHex(0xe5c04b);
      else col.setHSL(0.58 - 0.5 * f, 0.85, 0.55);
      lpos.push(p.x, p.y, p.z, q.x, q.y, q.z);
      lcol.push(col.r, col.g, col.b, col.r, col.g, col.b);
    };
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m], f = k / m;
      edge(V(a.l, a.hl), V(b.l, b.hl), f); // left rail
      edge(V(a.r, a.hr), V(b.r, b.hr), f); // right rail
      if (a.fl & 1) edge(V(a.l, a.hl), V(a.r, a.hr), f, (a.fl & 8) !== 0); // game rung
    }
    const lgeom = new THREE.BufferGeometry();
    lgeom.setAttribute('position', new THREE.Float32BufferAttribute(lpos, 3));
    lgeom.setAttribute('color', new THREE.Float32BufferAttribute(lcol, 3));
    group.add(new THREE.LineSegments(lgeom, new THREE.LineBasicMaterial({ vertexColors: true })));

    // Solid side walls: the track is an elevated ribbon walled along each edge, one
    // on the left rail and one on the right (as in the game), replacing the old centre
    // support lines. Each wall is a vertical quad strip that follows its rail and drops
    // to the ground (y=0). (Colour is a first pass — refined next.)
    const wpos = [];
    // one wall quad between rail points a (top height ha) and b (top height hb):
    // top edge follows the rail, bottom edge sits on the ground.
    const wallSeg = (a, ha, b, hb) => {
      wpos.push(a.x, ha, a.z, a.x, 0, a.z, b.x, hb, b.z);
      wpos.push(b.x, hb, b.z, a.x, 0, a.z, b.x, 0, b.z);
    };
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m];
      wallSeg(a.l, a.hl, b.l, b.hl); // left wall
      wallSeg(a.r, a.hr, b.r, b.hr); // right wall
    }
    const wgeom = new THREE.BufferGeometry();
    wgeom.setAttribute('position', new THREE.Float32BufferAttribute(wpos, 3));
    const walls = new THREE.Mesh(wgeom, new THREE.MeshBasicMaterial({
      color: 0x39424f, side: THREE.DoubleSide,
      polygonOffset: true, polygonOffsetFactor: 1, polygonOffsetUnits: 1,
    }));
    group.add(walls);

    // Start/finish marker (green) at ring 0.
    const r0 = rings[0];
    const sm = new THREE.Mesh(new THREE.SphereGeometry(0.12, 12, 12), new THREE.MeshBasicMaterial({ color: 0x35d07f }));
    sm.position.set((r0.l.x + r0.r.x) / 2, (r0.hl + r0.hr) / 2, (r0.l.z + r0.r.z) / 2);
    group.add(sm);

    t.scene.add(group);
    t.group = group;

    // Frame the whole circuit from a raised 3/4 angle so both the plan and the elevation
    // read, then hand control to the fly-cam to explore it.
    let maxH = 0;
    for (const r of rings) maxH = Math.max(maxH, r.hl, r.hr);
    const size = Math.max(8, maxH * 1.5); // world extent of the level (plan ~8 units + relief)
    const { camera, controls } = this.three;
    controls.target.set(0, maxH * 0.35, 0);
    camera.position.set(size * 0.28, size * 0.55, size * 0.95);
    camera.near = 0.01; camera.far = 200; camera.updateProjectionMatrix();
    controls.update();

    // Levels are explored with the free-flight controls (WASD/arrows, or the touch
    // sticks), like the other 3-D level viewers.
    this.fly.setScale(size);
    this.fly.setMoveScale(1.4);
    this.fly.setEnabled(true);

    if (this.hud) {
      this.hud.textContent = `${track.name} — ${track.sections} sections, ${m} rungs · ${flyHint}`;
    }
  }
}

function disposeGroup(g) {
  g.traverse(o => { if (o.geometry) o.geometry.dispose(); if (o.material) o.material.dispose(); });
}
