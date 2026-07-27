// Shared furnishing logic for the mansion viewers: instance a room's
// furniture from the game's placement database, and wire the vacuum
// interaction — click a piece to play its .anm clip (freezing open on the
// last frame), click again to run it backwards and close it up.
import * as THREE from 'three';

// furnishRoom adds one room entry's furniture instances under `group`.
// `cache` maps model path -> Promise<gltf|null> and is shared across rooms so
// repeated pieces load once. Animated instances get a mixer (pushed to
// `mixers`) and a toggle (holders pushed to `clickable`).
export async function furnishRoom({ loader, base, entry, group, cache, mixers, clickable }) {
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
  return { placed, animated };
}

// wireVacuum attaches the click-(not drag)-to-toggle handler; returns a
// disposer.
export function wireVacuum(stage, clickable) {
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
  return () => {
    canvas.removeEventListener('pointerdown', onDown);
    canvas.removeEventListener('pointerup', onUp);
  };
}
