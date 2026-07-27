// The `lm-room` renderer: a mansion room assembled the way the game does it —
// the room shell GLB plus its furniture, each piece a separate deduped GLB
// instanced at the {pos, rotDeg, scale} the game's own placement database
// (Map/map2.szp jmp/furnitureinfo) records for it. Rooms and furniture
// placements share one mansion-global coordinate frame, so assembly is just
// parenting each instance under a positioned holder.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

export default {
  kind: 'lm-room',
  async build({ item, base, stage }) {
    const loader = new GLTFLoader();
    const group = new THREE.Group();
    const [gltf, placements] = await Promise.all([
      loader.loadAsync(base + item.file),
      fetch(base + 'placements.json').then((r) => r.json()),
    ]);
    group.add(gltf.scene);

    const entry = (placements.rooms || []).find((r) => r.room === item.room);
    const cache = {};
    let placed = 0;
    for (const f of entry?.furniture || []) {
      if (!cache[f.model]) {
        cache[f.model] = loader.loadAsync(base + 'models/' + f.model).then((g) => g.scene, () => null);
      }
      const proto = await cache[f.model];
      if (!proto) continue;
      const holder = new THREE.Group();
      holder.position.set(f.pos[0], f.pos[1], f.pos[2]);
      holder.rotation.set(
        (f.rotDeg[0] * Math.PI) / 180,
        (f.rotDeg[1] * Math.PI) / 180,
        (f.rotDeg[2] * Math.PI) / 180,
        'ZYX',
      );
      holder.scale.set(f.scale[0], f.scale[1], f.scale[2]);
      holder.add(proto.clone(true));
      group.add(holder);
      placed++;
    }
    stage.add(group);

    // Frame the room interior: the shell's own bounds, seen from the manifest
    // camera when one is given, else from inside the shell — the rooms are
    // closed boxes, so an exterior fit would look at ceiling backfaces.
    const box = new THREE.Box3().setFromObject(gltf.scene);
    const c = box.getCenter(new THREE.Vector3());
    const s = box.getSize(new THREE.Vector3());
    const cam = item.camera || {};
    const pos = cam.pos || [c.x + s.x * 0.32, c.y + s.y * 0.18, c.z + s.z * 0.32];
    const tgt = cam.target || [c.x - s.x * 0.2, c.y - s.y * 0.1, c.z - s.z * 0.2];
    stage.camera.position.set(pos[0], pos[1], pos[2]);
    stage.controls.target.set(tgt[0], tgt[1], tgt[2]);
    stage.camera.near = 5;
    stage.camera.far = 200000;
    stage.camera.updateProjectionMatrix();
    stage.controls.update();
    if (placed) stage.hud = `${item.name} · ${placed} furniture pieces placed by the game's own database`;
    return group;
  },
};
