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

    // ---- flying objects (course.paths.json: the helicopter's waypoint
    // routes and the airplane's linear glide, decoded from the game's own
    // flight scripts — see webexport paths.go). Both animate on the game's
    // 30 Hz tick timeline; the airplane jumps back to its spawn when its
    // 1800-tick life ends.
    let flyers = null;
    if (item.pathsFile) {
      try {
        const paths = await fetch(base + item.pathsFile).then((r) => r.json());
        const heliBody = await loadObj(base + paths.helicopter.body);
        const heliRotor = await loadObj(base + paths.helicopter.rotor);
        const heli = new THREE.Group();
        heli.add(heliBody, heliRotor);
        const plane = await loadObj(base + paths.airplane.model);
        objectsGroup.add(heli, plane);
        placed.push(
          { object3d: heli, obj: { id: 188, addr: paths.helicopter.routes[0].addr, flags: 'dynamic, flight script', yaw: 0 } },
          { object3d: plane, obj: { id: 190, addr: '0x800739E4', flags: 'dynamic, linear glide', yaw: 0xF00 } },
        );
        // Per-route cumulative tick table: [{t0, t1, from, to, yaw0, yaw1, holdUntil}].
        // The routes end parked at the helipad for 9000 ticks (5 min) before
        // the next route starts; cap holds so the loop stays watchable.
        const routes = paths.helicopter.routes.map((r) => {
          const segs = [];
          let t = 0;
          let from = r.start, yaw = r.yaw;
          for (const k of r.keys) {
            const hold = Math.min(k.hold, 300);
            segs.push({ t0: t, t1: t + k.dur, from, to: k.pos, yaw0: yaw, yaw1: k.yaw, holdUntil: t + k.dur + hold });
            t = t + k.dur + hold;
            from = k.pos; yaw = k.yaw;
          }
          return { segs, total: t };
        });
        flyers = { paths, heli, heliRotor, plane, routes, t: 0 };
      } catch { flyers = null; }
    }
    const lerpAngle = (a, b, f) => {
      let d = (b - a) % (2 * Math.PI);
      if (d > Math.PI) d -= 2 * Math.PI;
      if (d < -Math.PI) d += 2 * Math.PI;
      return a + d * f;
    };
    const updateFlyers = (dt) => {
      if (!flyers) return;
      flyers.t += dt * flyers.paths.fps; // game ticks
      // Helicopter: fly route 0, then route 1, alternating (the script's
      // end-of-route opcode restarts with one of the two picked at random).
      const total = flyers.routes[0].total + flyers.routes[1].total;
      let ht = flyers.t % total;
      let route = flyers.routes[0];
      if (ht >= route.total) { ht -= route.total; route = flyers.routes[1]; }
      let seg = route.segs[route.segs.length - 1];
      for (const s of route.segs) { if (ht < s.holdUntil) { seg = s; break; } }
      const f = seg.t1 > seg.t0 ? Math.min(1, Math.max(0, (ht - seg.t0) / (seg.t1 - seg.t0))) : 1;
      flyers.heli.position.set(
        seg.from[0] + (seg.to[0] - seg.from[0]) * f,
        seg.from[1] + (seg.to[1] - seg.from[1]) * f,
        seg.from[2] + (seg.to[2] - seg.from[2]) * f,
      );
      flyers.heli.rotation.y = lerpAngle(seg.yaw0, seg.yaw1, f);
      flyers.heliRotor.rotation.y = (flyers.t / flyers.paths.fps) * (flyers.paths.helicopter.rotorRate || 0);
      // Airplane: linear glide, jumping back to the spawn at end of life.
      const a = flyers.paths.airplane;
      const pt = flyers.t % a.frames;
      flyers.plane.position.set(
        a.start[0] + a.delta[0] * pt,
        a.start[1] + a.delta[1] * pt,
        a.start[2] + a.delta[2] * pt,
      );
      flyers.plane.rotation.y = a.yaw;
    };

    stage.frame(track); // fits near/far to the whole course
    const size = new THREE.Box3().setFromObject(track).getSize(new THREE.Vector3()).length() || 100;

    // Open at the race-start camera (captured from the game, manifest.camera) so
    // the course reads from the grid looking down the track, not from a fitted
    // 3/4 overview. The FlyCam continues from whatever pose the camera holds.
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

    stage.onFrame = (camPos, dt) => { flycam.update(dt); updateFlyers(dt); };

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
