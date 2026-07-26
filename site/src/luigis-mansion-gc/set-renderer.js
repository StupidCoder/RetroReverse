// The `lm-set` renderer: a cutscene set piece (the forest glade, the mansion on its
// hill). These models carry their own sky dome around the playable clearing, so the
// generic bounding-sphere fit would park the camera outside the sky looking at its
// backside; instead the manifest entry supplies the establishing camera — a position
// and target inside the set, matching the cutscene's own framing — and the orbit
// controls take it from there.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

export default {
  kind: 'lm-set',
  async build({ item, base, stage }) {
    const gltf = await new GLTFLoader().loadAsync(base + item.file);
    const obj = gltf.scene;
    stage.add(obj);

    const cam = item.camera || {};
    const pos = cam.pos || [0, 500, 3000];
    const tgt = cam.target || [0, 500, 0];
    stage.camera.position.set(pos[0], pos[1], pos[2]);
    stage.controls.target.set(tgt[0], tgt[1], tgt[2]);
    // The sets are thousands of units across; keep the whole dome in the frustum.
    stage.camera.near = 5;
    stage.camera.far = 200000;
    stage.camera.updateProjectionMatrix();
    stage.controls.update();

    // When the manifest names a camera track — the shot's own .scd/.sco camera,
    // exported one sample per frame — fly it on loop until the viewer grabs the
    // controls, from then on it is an ordinary orbit scene.
    if (item.cameraTrack) {
      try {
        const doc = await fetch(base + item.cameraTrack).then((r) => r.json());
        const track = doc.track || [];
        if (track.length > 1) {
          let t = 0, flying = true;
          const stopFly = () => { flying = false; };
          stage.canvas.addEventListener('pointerdown', stopFly, { once: true });
          const prev = stage.onFrame;
          stage.onFrame = (camPos, dt) => {
            if (prev) prev(camPos, dt);
            if (!flying) return;
            t = (t + (dt || 0) * (doc.fps || 30)) % track.length;
            const f = track[Math.floor(t)];
            stage.camera.position.set(f.pos[0], f.pos[1], f.pos[2]);
            stage.controls.target.set(f.target[0], f.target[1], f.target[2]);
            stage.camera.fov = f.fov || stage.camera.fov;
            stage.camera.updateProjectionMatrix();
            stage.controls.update();
          };
          stage.hud = `${item.name} · the cutscene's own camera — drag to take over`;
        }
      } catch { /* the set still shows from its establishing camera */ }
    }
    return obj;
  },
};
