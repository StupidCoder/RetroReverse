// Need for Speed (3DO) course renderer — the `nfs-course` plugin for the
// shared Viewer3D/Stage3D seam. It loads the road GLB (webexport
// course-cy1.glb: the slice geometry baked from the .trk cross-sections and
// LaunchMe's face tables), then the roadside objects: each placement in
// cy1.objects.json names a per-object GLB (models/obj-NN.glb) positioned and
// yaw-rotated in the world (webexport objects.go). The objects go in one
// group so the Studio's "Objects" layer toggle can hide them; a shared picker
// shows each object's def / type / segment card on click; and the course
// opens at the in-game race-start camera (manifest.camera — the driver's-eye
// interior pose captured from the running game) and is flown with the FlyCam.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';
import { installPicker } from '../shared/objinfo.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

export default {
  kind: 'nfs-course',
  async build({ item, base, stage }) {
    const { scene, camera, controls } = stage;

    const trackGltf = await gltfLoader().loadAsync(base + item.file);
    const track = trackGltf.scene;
    stage.add(track);

    // Roadside objects: one group so the layer toggle can hide them all. Each
    // placement loads its object GLB once (cached by url) and clones it.
    const objectsGroup = new THREE.Group();
    stage.add(objectsGroup);
    const placed = []; // { object3d, obj } for the picker
    const cache = new Map();
    const loadObj = async (url) => {
      if (!cache.has(url)) cache.set(url, gltfLoader().loadAsync(url).then((g) => g.scene));
      return (await cache.get(url)).clone(true);
    };

    let placements = [];
    if (item.objectsFile) {
      try {
        placements = (await fetch(base + item.objectsFile).then((r) => r.json())).objects || [];
      } catch { placements = []; }
    }
    await Promise.all(placements.map(async (o) => {
      const node = await loadObj(base + o.model);
      const p = o.pos || [0, 0, 0];
      node.position.set(p[0], p[1] || 0, p[2] || 0);
      if (o.rot) node.rotation.set(o.rot[0] || 0, o.rot[1] || 0, o.rot[2] || 0);
      objectsGroup.add(node);
      placed.push({ object3d: node, obj: o });
    }));

    stage.frame(track); // fits near/far to the whole course
    const size = new THREE.Box3().setFromObject(track).getSize(new THREE.Vector3()).length() || 100;

    // Horizon sky dome (models/sky-<course>.glb): the packet's two 6-cel
    // panorama rings as concentric cylinders — the opaque far ring
    // (sky + distant terrain) behind the alpha-cutout near ring (treeline).
    // Camera-centred every frame and drawn behind everything with depth
    // writes off, the same contract as the DS viewers' skyboxes; the GLB is
    // authored around a unit radius, so scale it to sit inside the far plane.
    let sky = null;
    if (item.sky) {
      try {
        const g = await gltfLoader().loadAsync(base + item.sky);
        sky = g.scene;
        let order = -1004; // primitive order is paint order: far ring, caps, near ring
        sky.traverse((n) => {
          if (n.isMesh) {
            n.frustumCulled = false;
            n.renderOrder = order++;
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
        sky.scale.setScalar(camera.far * 0.45);
        stage.add(sky);
      } catch { sky = null; }
    }

    // Open at the race-start camera (captured from the game, manifest.camera):
    // the driver's-eye interior pose at the grid looking down the track, not a
    // fitted 3/4 overview. The FlyCam continues from whatever pose it holds.
    if (item.camera) {
      const p = item.camera.pos, t = item.camera.target;
      camera.position.set(p[0], p[1], p[2]);
      controls.target.set(t[0], t[1], t[2]);
      controls.update();
    }

    // ---- fly camera ----
    controls.autoRotate = false;
    const flycam = new FlyCam(camera, controls, stage.el);
    flycam.setScale(size);
    flycam.setMoveScale(0.12); // the City stage is ~9.7 km in world units — slow the fly cam
    flycam.setEnabled(true);
    stage.fly = flycam;
    stage.hud = `${item.name} · ${flyHint}`;

    // ---- Objects / Horizon layer toggles ----
    let objectsOn = true;
    stage.setLayer = (id, on) => {
      if (id === 'objects') { objectsOn = on; objectsGroup.visible = on; }
      if (id === 'sky' && sky) sky.visible = on;
    };

    // ---- click-to-inspect: def / type / segment card ----
    const TYPES = { 1: '3D model', 4: 'billboard', 6: 'two-anchor cel' };
    const picker = installPicker(
      { el: stage.el, canvas: stage.canvas, camera },
      () => placed,
      {
        enabled: () => objectsOn,
        resolve: (o) => ({
          title: `Object def ${o.def}`,
          subtitle: `${TYPES[o.type] || `type ${o.type}`} · segment ${o.segment}`,
          body: `placement #${o.index}\nyaw ${o.rawYaw} (${(o.rawYaw * 360 / 256).toFixed(1)}°)\nrecord ${o.flags}`,
        }),
      });

    stage.onFrame = (camPos, dt) => {
      flycam.update(dt);
      if (sky) sky.position.copy(camPos); // the dome rides the camera
    };

    stage.disposePlugin = () => {
      picker.dispose();
      flycam.dispose();
      controls.autoRotate = false;
      const dispose = (root) => root.traverse((o) => {
        if (o.geometry) o.geometry.dispose();
        if (o.material) for (const m of Array.isArray(o.material) ? o.material : [o.material]) m.dispose();
      });
      dispose(track);
      dispose(objectsGroup);
      if (sky) dispose(sky);
    };

    return track;
  },
};
