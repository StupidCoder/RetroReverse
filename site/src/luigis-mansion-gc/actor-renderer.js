// The `lm-actor` renderer: an animated cutscene actor (Luigi, the gate torch).
// These are envelope-skinned GLBs whose bind-space geometry sits where the
// artist modelled it, while the animation places the skeleton somewhere else —
// so the generic fit-to-geometry framing can point the camera away from the
// posed character, and bind-space frustum culling can drop it entirely. This
// plugin plays the clip, poses the skeleton once, frames the camera on the
// skeleton-aware bounds, and disables frustum culling for the bob of the walk.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

export default {
  kind: 'lm-actor',
  async build({ item, base, stage }) {
    const gltf = await new GLTFLoader().loadAsync(base + item.file);
    const obj = gltf.scene;
    stage.add(obj);

    const clips = gltf.animations || [];
    let mixer = null;
    if (clips.length) {
      mixer = new THREE.AnimationMixer(obj);
      mixer.clipAction(clips[0]).play();
      mixer.update(0);
    }
    obj.updateMatrixWorld(true);

    // Frame the posed skeleton, not the bind-space geometry.
    const box = new THREE.Box3();
    obj.traverse((o) => {
      if (!o.isMesh) return;
      o.frustumCulled = false;
      if (o.isSkinnedMesh && o.computeBoundingBox) {
        o.computeBoundingBox(); // skeleton-aware on SkinnedMesh
        box.union(o.boundingBox.clone().applyMatrix4(o.matrixWorld));
      } else {
        box.expandByObject(o);
      }
    });
    if (!box.isEmpty()) {
      const sphere = box.getBoundingSphere(new THREE.Sphere());
      const c = sphere.center, r = sphere.radius || 1;
      const dist = (r * 1.5) / Math.sin((stage.camera.fov * Math.PI) / 360);
      stage.camera.position.set(c.x + dist * 0.6, c.y + dist * 0.25, c.z + dist * 0.76);
      stage.camera.near = Math.max(0.1, r / 50);
      stage.camera.far = Math.max(stage.camera.far, dist + r * 100);
      stage.camera.updateProjectionMatrix();
      stage.controls.target.copy(c);
      stage.controls.update();
    }

    if (mixer) {
      const prev = stage.onFrame;
      stage.onFrame = (camPos, dt) => { if (prev) prev(camPos, dt); mixer.update(dt); };
    }
    return obj;
  },
};
