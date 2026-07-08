// Marble Madness slope view — the three.js half of the course viewer: the static
// slope field (the height the marble rolls on, decoded from the Track file) as a 3-D
// terrain you drag to rotate, with the Track-layer markers (creature routes and
// placements). The terrain is a pre-baked GLB (a solid triangulated height-band mesh
// in world coords); the markers are a sidecar JSON of world-coord pins & routes. The
// tilemap half runs on the shared 2-D LevelViewer.

import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { applyWireframe } from '../shared/wireframe.js';

// "#rrggbb" -> [r,g,b] in 0..1 for three.js vertex colours.
function hexRgb(hex) {
  const c = parseInt(hex.replace('#', ''), 16);
  return [((c >> 16) & 255) / 255, ((c >> 8) & 255) / 255, (c & 255) / 255];
}

export class SlopeView {
  // isActive() gates the render loop (the shell owns mode/visibility).
  constructor(el, isActive) {
    this.el = el;
    this.isActive = isActive;
    this.three = null;
    this.markersOn = false;
    this.wireframeOn = false;
  }

  get canvas() { return this.three && this.three.renderer.domElement; }

  setMarkers(on) {
    this.markersOn = on;
    if (this.three && this.three.markers) this.three.markers.visible = on;
  }

  setWireframe(on) {
    this.wireframeOn = on;
    if (this.three && this.three.model) applyWireframe(this.three.model, on);
  }

  // show(scene, markers): scene is a loaded GLB terrain (gltf.scene, already in the
  // world coords the offline renderer used); markers is the parsed sidecar
  // { pins:[{pos,color}], paths:[{points,color}] } in those same coords.
  show(scene, markers) {
    if (!this.three) this._initThree();
    this._buildMesh(scene, markers);
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
    this.three = { renderer, scene, camera, controls, model: null, markers: null, frustumH: 1 };
    // Pivot where you're looking: when a gesture starts, raycast the screen
    // centre onto the terrain and orbit around that point.
    const ray = new THREE.Raycaster();
    controls.addEventListener('start', () => {
      if (!this.three.model) return;
      ray.setFromCamera(new THREE.Vector2(0, 0), this.three.camera);
      const hit = ray.intersectObject(this.three.model, true)[0];
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

  // Add the pre-baked GLB terrain (a solid triangulated height-band mesh, already
  // in world coords) and frame it isometrically (orthographic) down a 2:1 dimetric
  // angle from its bounding box. The solid mesh writes depth itself, so the marker
  // routes/pins occlude correctly against it (no more invisible fill).
  _buildMesh(scene, markers) {
    const t = this.three;
    const dispose = (g) => { if (g) { t.scene.remove(g); g.traverse((o) => { if (o.geometry) o.geometry.dispose(); if (o.material) o.material.dispose(); }); } };
    dispose(t.model); dispose(t.markers);

    t.scene.add(scene);
    t.model = scene;

    const box = new THREE.Box3().setFromObject(scene);
    const ctr = box.getCenter(new THREE.Vector3());
    const size = box.getSize(new THREE.Vector3());
    const span = Math.max(size.x, size.z);

    t.markers = this._buildMarkers(markers, span);
    t.markers.visible = this.markersOn;
    t.scene.add(t.markers);

    t.frustumH = span;
    const dir = new THREE.Vector3(0.632, 0.447, 0.632); // azimuth 45°, elevation ~26.6°
    t.camera.position.copy(ctr).addScaledVector(dir, span * 2);
    t.camera.near = 0.01; t.camera.far = span * 8;
    t.camera.zoom = 1;
    t.controls.target.copy(ctr);
    t.controls.minZoom = 0.4; t.controls.maxZoom = 6;
    t.controls.update();
  }

  // Build the Track-layer markers from the sidecar (positions already in the GLB's
  // world coords): a coloured pin per placement and a coloured route polyline per
  // creature path, depth-tested so the terrain occludes them.
  _buildMarkers(markers, span) {
    const stem = Math.max(2, span * 0.045);
    const pins = (markers && markers.pins) || [];
    const paths = (markers && markers.paths) || [];

    const stemPos = [], stemCol = [], headPos = [], headCol = [], routePos = [], routeCol = [];
    const pin = (pos, col) => {
      const [px, py, pz] = pos;
      stemPos.push(px, py, pz, px, py + stem, pz); stemCol.push(...col, ...col);
      headPos.push(px, py + stem, pz); headCol.push(...col);
    };
    for (const p of pins) pin(p.pos, hexRgb(p.color));
    for (const path of paths) {
      const col = hexRgb(path.color);
      for (let i = 0; i + 1 < path.points.length; i++) {
        routePos.push(...path.points[i], ...path.points[i + 1]);
        routeCol.push(...col, ...col);
      }
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
