// Ridge Racer course renderer — the `rr-course` plugin for the shared
// Viewer3D/Stage3D seam. It loads the road GLB (webexport track.glb, sections
// only), then the roadside objects: each placement in course.objects.json
// names a per-object GLB (models/obj-NN.glb) positioned and yaw-rotated in the
// world (webexport objects.go). The objects go in one group so the Studio's
// "Objects" layer toggle can hide them, a shared picker shows each object's
// address / flag / rotation card on click, and the course is presented with
// the free-flight FlyCam like the other level viewers.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';
import { installPicker } from '../shared/objinfo.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

export default {
  kind: 'rr-course',
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

    stage.frame(track);
    const size = new THREE.Box3().setFromObject(track).getSize(new THREE.Vector3()).length() || 100;

    // ---- fly camera ----
    controls.autoRotate = false;
    const flycam = new FlyCam(camera, controls, stage.el);
    flycam.setScale(size);
    flycam.setMoveScale(1.2);
    flycam.setEnabled(true);
    stage.fly = flycam;
    stage.hud = `${item.name} · ${flyHint}`;

    // ---- Objects layer toggle ----
    let objectsOn = true;
    stage.setLayer = (id, on) => {
      if (id === 'objects') { objectsOn = on; objectsGroup.visible = on; }
    };

    // ---- click-to-inspect: address / flag / rotation card ----
    const picker = installPicker(
      { el: stage.el, canvas: stage.canvas, camera },
      () => placed,
      {
        enabled: () => objectsOn,
        resolve: (o) => ({
          title: `Object ${o.id}`,
          subtitle: `record ${o.addr}`,
          body: `flags ${o.flags}\nrotation ${o.yaw} (${(o.yaw * 360 / 4096).toFixed(1)}°)`,
        }),
      });

    stage.onFrame = (camPos, dt) => { flycam.update(dt); };

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
    };

    return track;
  },
};
