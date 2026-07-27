// The `lm-room` renderer: a mansion room assembled the way the game does it —
// the room shell GLB plus its furniture, each piece a separate deduped GLB
// instanced at the {pos, rotDeg, scale} the game's own placement database
// (Map/map2.szp jmp/furnitureinfo) records for it. Rooms and furniture
// placements share one mansion-global coordinate frame, so assembly is just
// parenting each instance under a positioned holder.
//
// A room is a level, not an object: the camera flies (WASD/arrows, sticks on
// touch) instead of auto-orbiting. And the furniture keeps the game's own
// interaction: in the game the vacuum yanks drawers and doors open — click a
// piece to play that .anm clip, freezing on the last frame; click again to
// play it backwards, closing it up.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

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
    const mixers = [];
    const clickable = [];
    let placed = 0, animated = 0;
    for (const f of entry?.furniture || []) {
      if (!cache[f.model]) {
        cache[f.model] = loader.loadAsync(base + 'models/' + f.model).then((g) => g, () => null);
      }
      const proto = await cache[f.model];
      if (!proto) continue;
      const inst = proto.scene.clone(true);
      const holder = new THREE.Group();
      holder.position.set(f.pos[0], f.pos[1], f.pos[2]);
      holder.rotation.set(
        (f.rotDeg[0] * Math.PI) / 180,
        (f.rotDeg[1] * Math.PI) / 180,
        (f.rotDeg[2] * Math.PI) / 180,
        'ZYX',
      );
      holder.scale.set(f.scale[0], f.scale[1], f.scale[2]);
      holder.add(inst);
      group.add(holder);
      placed++;

      // The interaction clip is the last one (chest_0 is the rest pose,
      // chest_1 the vacuum-yanked opening); a lone 2-frame rest clip is
      // static and gets no click.
      const clips = proto.animations || [];
      const clip = clips[clips.length - 1];
      if (clip && clip.duration > 2.5 / 30) {
        const mixer = new THREE.AnimationMixer(inst);
        const action = mixer.clipAction(clip);
        action.loop = THREE.LoopOnce;
        action.clampWhenFinished = true;
        holder.userData.toggle = () => {
          const open = holder.userData.open;
          action.timeScale = open ? -1 : 1;
          if (!open && action.time >= clip.duration) action.time = 0;
          if (open && action.time <= 0) action.time = clip.duration;
          action.paused = false;
          action.enabled = true;
          action.play();
          holder.userData.open = !open;
        };
        mixers.push(mixer);
        clickable.push(holder);
        animated++;
      }
    }
    stage.add(group);

    // Camera: start inside the shell (rooms are closed boxes), manifest
    // camera first, then fly.
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
    stage.controls.autoRotate = false;

    const flycam = new FlyCam(stage.camera, stage.controls, stage.el);
    flycam.setScale(s.length() || 1000);
    flycam.setMoveScale(0.6);
    flycam.setEnabled(true);
    stage.fly = flycam;

    // Click (not drag) a furniture piece to vacuum it open / closed.
    const canvas = stage.canvas;
    const ray = new THREE.Raycaster();
    const ndc = new THREE.Vector2();
    let downX = 0, downY = 0;
    const onDown = (e) => { downX = e.clientX; downY = e.clientY; };
    const onUp = (e) => {
      if (Math.hypot(e.clientX - downX, e.clientY - downY) > 5) return;
      const rect = canvas.getBoundingClientRect();
      ndc.set(((e.clientX - rect.left) / rect.width) * 2 - 1, -((e.clientY - rect.top) / rect.height) * 2 + 1);
      ray.setFromCamera(ndc, stage.camera);
      const hits = ray.intersectObjects(clickable, true);
      for (const h of hits) {
        let o = h.object;
        while (o && !o.userData.toggle) o = o.parent;
        if (o) { o.userData.toggle(); break; }
      }
    };
    canvas.addEventListener('pointerdown', onDown);
    canvas.addEventListener('pointerup', onUp);

    const prev = stage.onFrame;
    stage.onFrame = (camPos, dt) => {
      if (prev) prev(camPos, dt);
      flycam.update(dt);
      for (const m of mixers) m.update(dt || 0);
    };
    stage.disposePlugin = () => {
      flycam.dispose();
      canvas.removeEventListener('pointerdown', onDown);
      canvas.removeEventListener('pointerup', onUp);
    };

    const bits = [`${placed} furniture pieces placed by the game's own database`];
    if (animated) bits.push(`click to vacuum ${animated} of them open`);
    bits.push(flyHint);
    stage.hud = `${item.name} · ${bits.join(' · ')}`;
    return group;
  },
};
