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
    return obj;
  },
};
