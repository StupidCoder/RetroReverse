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

    // Frame the posed skeleton, not the bind-space geometry — and only its
    // dominant cluster: a model can carry parts the demo script keeps hidden,
    // parked far from the character (E. Gadd's has some at the set origin).
    const parts = [];
    obj.traverse((o) => {
      if (!o.isMesh) return;
      o.frustumCulled = false;
      const b = new THREE.Box3();
      if (o.isSkinnedMesh && o.computeBoundingBox) {
        o.computeBoundingBox(); // skeleton-aware on SkinnedMesh
        b.copy(o.boundingBox).applyMatrix4(o.matrixWorld);
      } else {
        b.expandByObject(o);
      }
      if (!b.isEmpty()) parts.push(b);
    });
    let box = new THREE.Box3();
    if (parts.length) {
      const size = (b) => b.getSize(new THREE.Vector3()).length();
      const seed = parts.reduce((a, b) => (size(b) > size(a) ? b : a));
      const sc = seed.getCenter(new THREE.Vector3());
      const reach = Math.max(size(seed) * 2, 1);
      for (const b of parts) {
        if (b.getCenter(new THREE.Vector3()).distanceTo(sc) <= reach) box.union(b);
      }
    }
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
      // Follow the actor: some clips carry their demo blocking (E. Gadd runs
      // thousands of units), so the camera tracks the skeleton's bone centroid,
      // preserving whatever orbit offset the viewer has chosen.
      // Follow only the bones inside the framed cluster — far-parked hidden
      // parts must not drag the camera between clusters.
      const sphere = box.isEmpty() ? null : box.getBoundingSphere(new THREE.Sphere());
      const tracked = [];
      obj.traverse((o) => {
        if (!o.isBone) return;
        if (!sphere || o.getWorldPosition(new THREE.Vector3()).distanceTo(sphere.center) <= sphere.radius * 1.5) {
          tracked.push(o);
        }
      });
      const centroid = () => {
        const c = new THREE.Vector3();
        for (const b of tracked) c.add(b.getWorldPosition(new THREE.Vector3()));
        return tracked.length ? c.divideScalar(tracked.length) : c;
      };
      let last = centroid();
      const prev = stage.onFrame;
      stage.onFrame = (camPos, dt) => {
        if (prev) prev(camPos, dt);
        mixer.update(dt);
        obj.updateMatrixWorld(true);
        const now = centroid();
        const delta = now.clone().sub(last);
        last = now;
        if (delta.lengthSq() > 0) {
          stage.camera.position.add(delta);
          stage.controls.target.add(delta);
          stage.controls.update();
        }
      };
    }
    return obj;
  },
};
