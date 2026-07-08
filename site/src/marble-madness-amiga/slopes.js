// Marble Madness slope view — the three.js half of the course viewer: the static
// slope field (the height the marble rolls on, decoded from the Track file) as a 3-D
// height-mesh you drag to rotate, with the Track-layer markers (creature routes and
// placements). Moved verbatim from the old combined viewer; the tilemap half now runs
// on the shared 2-D LevelViewer.

import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

const HEIGHT_SCALE = 0.15;               // slope-mesh vertical exaggeration (per tile unit)

// heightRamp maps t in [0,1] to blue(low)..white(high), matching the offline
// region renderer; returns r,g,b in 0..1 for three.js vertex colours.
const RAMP = [
  [0.0, 30, 40, 120], [0.3, 40, 140, 150], [0.55, 60, 170, 80], [0.78, 220, 205, 70], [1.0, 250, 250, 250],
];
function heightRamp(t) {
  t = Math.max(0, Math.min(1, t));
  for (let i = 0; i < RAMP.length - 1; i++) {
    if (t <= RAMP[i + 1][0]) {
      const a = RAMP[i], b = RAMP[i + 1];
      const f = (t - a[0]) / (b[0] - a[0] + 1e-9);
      return [(a[1] + (b[1] - a[1]) * f) / 255, (a[2] + (b[2] - a[2]) * f) / 255, (a[3] + (b[3] - a[3]) * f) / 255];
    }
  }
  return [250 / 255, 250 / 255, 250 / 255];
}

export class SlopeView {
  // isActive() gates the render loop (the shell owns mode/visibility).
  constructor(el, isActive) {
    this.el = el;
    this.isActive = isActive;
    this.three = null;
    this.slope = null;
    this.markersOn = false;
  }

  get canvas() { return this.three && this.three.renderer.domElement; }

  setMarkers(on) {
    this.markersOn = on;
    if (this.three && this.three.markers) this.three.markers.visible = on;
  }

  show(slope) {
    this.slope = slope;
    if (!this.three) this._initThree();
    this._buildMesh();
    this._resize();
  }

  _initThree() {
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
    renderer.domElement.classList.add('mm-three');
    this.el.appendChild(renderer.domElement);
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x0a0e16);
    // Orthographic = no perspective distortion (the isometric look of the wire
    // PNGs). OrbitControls still lets you swing the angle around.
    const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0.01, 1);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.rotateSpeed = 0.9;
    this.three = { renderer, scene, camera, controls, group: null, markers: null, fill: null, frustumH: 1 };
    // Pivot where you're looking: when a gesture starts, raycast the screen
    // centre onto the surface and orbit around that point.
    const ray = new THREE.Raycaster();
    controls.addEventListener('start', () => {
      if (!this.three.fill) return;
      ray.setFromCamera(new THREE.Vector2(0, 0), this.three.camera);
      const hit = ray.intersectObject(this.three.fill, false)[0];
      if (hit) this.three.controls.target.copy(hit.point);
    });
    new ResizeObserver(() => this._resize()).observe(this.el);
    const tick = () => {
      if (this.isActive() && this.three) {
        this.three.controls.update();
        this.three.renderer.render(this.three.scene, this.three.camera);
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  }

  _resize() {
    if (!this.three) return;
    const w = this.el.clientWidth, h = this.el.clientHeight;
    if (!w || !h) return;
    this.three.renderer.setSize(w, h, false);
    const cam = this.three.camera, hh = this.three.frustumH, aspect = w / h;
    cam.left = -hh * aspect; cam.right = hh * aspect; cam.top = hh; cam.bottom = -hh;
    cam.updateProjectionMatrix();
  }

  // Build the slope as a wireframe with hidden-line removal, like the offline
  // *.wire.png: each rolling-surface tile is a vertex at (tx, height, ty); the
  // grid edges to the (+1,0) and (0,+1) neighbours are drawn as height-coloured
  // lines, and an invisible fill of the same quads writes depth so lines behind
  // the surface are hidden. Pits (no surface) leave holes.
  _buildMesh() {
    const t = this.three;
    const dispose = (g) => { if (g) { t.scene.remove(g); g.traverse((o) => { if (o.geometry) o.geometry.dispose(); if (o.material) o.material.dispose(); }); } };
    dispose(t.group); dispose(t.markers);
    const s = this.slope;
    const { w, h, heights } = s;
    const range = Math.max(1, s.hi - s.lo);
    const cx = (w - 1) / 2, cz = (h - 1) / 2;
    const present = (gx, gy) => gx >= 0 && gy >= 0 && gx < w && gy < h && heights[gy * w + gx] > 0;
    // Axis swap (tile-X -> world Z, tile-Y -> world X) so the isometric view
    // matches the offline *.wire.png instead of mirroring it.
    const wX = (gx, gy) => gy - cz;
    const wZ = (gx, gy) => gx - cx;
    const surfY = (gx, gy) => (heights[gy * w + gx] - 1) * HEIGHT_SCALE;

    // Invisible depth fill (triangulated quads) for hidden-line removal.
    const vidx = new Int32Array(w * h).fill(-1);
    const fpos = [];
    for (let gy = 0; gy < h; gy++) {
      for (let gx = 0; gx < w; gx++) {
        if (!present(gx, gy)) continue;
        vidx[gy * w + gx] = fpos.length / 3;
        fpos.push(wX(gx, gy), surfY(gx, gy), wZ(gx, gy));
      }
    }
    const idx = [];
    for (let gy = 0; gy < h - 1; gy++) {
      for (let gx = 0; gx < w - 1; gx++) {
        const a = vidx[gy * w + gx], b = vidx[gy * w + gx + 1], c = vidx[(gy + 1) * w + gx], d = vidx[(gy + 1) * w + gx + 1];
        if (a >= 0 && b >= 0 && c >= 0 && d >= 0) idx.push(a, c, b, b, c, d);
      }
    }
    const fgeom = new THREE.BufferGeometry();
    fgeom.setAttribute('position', new THREE.Float32BufferAttribute(fpos, 3));
    fgeom.setIndex(idx);
    const fill = new THREE.Mesh(fgeom, new THREE.MeshBasicMaterial({
      colorWrite: false, side: THREE.DoubleSide,
      polygonOffset: true, polygonOffsetFactor: 1, polygonOffsetUnits: 1,
    }));

    // Height-coloured grid edges.
    const epos = [], ecol = [];
    const pushV = (gx, gy) => {
      epos.push(wX(gx, gy), surfY(gx, gy), wZ(gx, gy));
      const c = heightRamp((heights[gy * w + gx] - 1) / range);
      ecol.push(c[0], c[1], c[2]);
    };
    for (let gy = 0; gy < h; gy++) {
      for (let gx = 0; gx < w; gx++) {
        if (!present(gx, gy)) continue;
        if (present(gx + 1, gy)) { pushV(gx, gy); pushV(gx + 1, gy); }
        if (present(gx, gy + 1)) { pushV(gx, gy); pushV(gx, gy + 1); }
      }
    }
    const egeom = new THREE.BufferGeometry();
    egeom.setAttribute('position', new THREE.Float32BufferAttribute(epos, 3));
    egeom.setAttribute('color', new THREE.Float32BufferAttribute(ecol, 3));
    const lines = new THREE.LineSegments(egeom, new THREE.LineBasicMaterial({ vertexColors: true }));

    const group = new THREE.Group();
    group.add(fill, lines);
    t.scene.add(group);
    t.group = group;
    t.fill = fill;

    const span = Math.max(w, h);
    t.markers = this._buildMarkers(s, { present, wX, wZ, surfY, span });
    t.markers.visible = this.markersOn;
    t.scene.add(t.markers);

    // Frame it isometrically (orthographic), down a 2:1 dimetric angle.
    const ctr = new THREE.Vector3(0, (range * HEIGHT_SCALE) / 2, 0);
    t.frustumH = span;
    const dir = new THREE.Vector3(0.632, 0.447, 0.632); // azimuth 45°, elevation ~26.6°
    t.camera.position.copy(ctr).addScaledVector(dir, span * 2);
    t.camera.near = 0.01; t.camera.far = span * 8;
    t.camera.zoom = 1;
    t.controls.target.copy(ctr);
    t.controls.minZoom = 0.4; t.controls.maxZoom = 6;
    t.controls.update();
  }

  // Build the Track-layer markers: a coloured pin per single object and a coloured
  // route polyline per creature path, depth-tested so the slopes occlude them.
  _buildMarkers(s, m) {
    const { present, wX, wZ, surfY, span } = m;
    const stem = Math.max(2, span * 0.045);
    const gx = (x) => x - s.x0, gy = (y) => y - s.y0;
    const surf = (x, y) => (present(gx(x), gy(y)) ? surfY(gx(x), gy(y)) : 0);
    const W = (x, y) => [wX(gx(x), gy(y)), surf(x, y), wZ(gx(x), gy(y))];
    const rgb = (c) => [((c >> 16) & 255) / 255, ((c >> 8) & 255) / 255, (c & 255) / 255];

    const stemPos = [], stemCol = [], headPos = [], headCol = [], routePos = [], routeCol = [];
    const pin = (x, y, c) => {
      const [px, py, pz] = W(x, y), col = rgb(c);
      stemPos.push(px, py, pz, px, py + stem, pz); stemCol.push(...col, ...col);
      headPos.push(px, py + stem, pz); headCol.push(...col);
    };
    for (const p of s.markers.points) pin(p.x, p.y, p.c);
    for (const path of s.markers.paths) {
      const col = rgb(path.c);
      for (let i = 0; i + 1 < path.pts.length; i++) {
        const a = W(path.pts[i][0], path.pts[i][1]), b = W(path.pts[i + 1][0], path.pts[i + 1][1]);
        routePos.push(...a, ...b); routeCol.push(...col, ...col);
      }
      if (path.pts.length) pin(path.pts[0][0], path.pts[0][1], path.c); // spawn pin
    }

    const g = new THREE.Group();
    const lineSeg = (pos, col) => {
      const geo = new THREE.BufferGeometry();
      geo.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3));
      geo.setAttribute('color', new THREE.Float32BufferAttribute(col, 3));
      const o = new THREE.LineSegments(geo, new THREE.LineBasicMaterial({ vertexColors: true }));
      o.renderOrder = 3;
      return o;
    };
    if (routePos.length) g.add(lineSeg(routePos, routeCol));
    if (stemPos.length) g.add(lineSeg(stemPos, stemCol));
    if (headPos.length) {
      const geo = new THREE.BufferGeometry();
      geo.setAttribute('position', new THREE.Float32BufferAttribute(headPos, 3));
      geo.setAttribute('color', new THREE.Float32BufferAttribute(headCol, 3));
      const pts = new THREE.Points(geo, new THREE.PointsMaterial({ vertexColors: true, size: 6, sizeAttenuation: false }));
      pts.renderOrder = 4;
      g.add(pts);
    }
    return g;
  }
}
