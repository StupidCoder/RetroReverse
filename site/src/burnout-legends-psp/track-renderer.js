// Burnout Legends track renderer — the `bl-track` plugin for the shared
// Viewer3D/Stage3D seam.
//
// A track ships as two GLBs, because the game keeps its world in two layers and
// swaps between them (webexport, package bgt):
//
//   item.file  the streamed world — the city you drive through, decoded from
//              streamed.dat's chain of per-cell blocks
//   item.env   static.dat — the environment it sits in: terrain, cliffs, forest
//              and the mountain backdrop
//
// They are not additive. static.dat carries a COARSE stand-in for the ground the
// streamed cells cover, and the engine draws a static group only where its cell
// is not resident — so with both layers on, that stand-in wins the depth test
// over the real road in places. The environment therefore lives in its own group
// behind a layer toggle: the streamed world is the thing you came to look at,
// and the environment is there when you want to see the valley it sits in.
//
// The track flies like the other world viewers, opening low over the road rather
// than out at the bounding sphere — a circuit is long and thin, and a fitted 3/4
// overview of one puts the camera uselessly far away.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

export default {
  kind: 'bl-track',
  async build({ item, base, stage }) {
    const { camera, controls } = stage;

    const world = (await gltfLoader().loadAsync(base + item.file)).scene;
    stage.add(world);
    stage.frame(world); // fits near/far to the whole track

    // The environment layer, loaded up front so the toggle is instant.
    const env = new THREE.Group();
    env.visible = false;
    stage.add(env);
    if (item.env) {
      env.add((await gltfLoader().loadAsync(base + item.env)).scene);
    }
    stage.setLayer = (id, on) => {
      if (id === 'environment') env.visible = !!on;
    };

    // Give the camera a frustum a city fits in. stage.frame() only ever widens
    // one — near shrinks, far grows — and a track's bounding sphere is thousands
    // of units across, so it leaves a near/far ratio in the millions. That has
    // nothing left for the depth buffer, and the world comes out as z-fighting
    // mush that reads, on a dark track, as no world at all.
    const size = new THREE.Box3().setFromObject(world).getSize(new THREE.Vector3());
    const span = Math.max(size.x, size.y, size.z) || 100;
    camera.near = Math.max(0.5, span / 4000);
    camera.far = span * 3;

    // Open on the road looking along it (manifest spawn: the game's own per-cell
    // anchors, which trace the circuit — webexport). A fitted 3/4 overview of a
    // long thin loop leaves it a speck in the middle of the screen, and a point
    // picked out of the bounding box lands inside a building as often as not.
    if (item.spawn) {
      const { pos, target } = item.spawn;
      camera.position.set(pos[0], pos[1], pos[2]);
      controls.target.set(target[0], target[1], target[2]);
    }
    camera.updateProjectionMatrix();
    controls.update();
    controls.autoRotate = false;
    const flycam = new FlyCam(camera, controls, stage.el);
    flycam.setScale(size.length() || 100);
    flycam.setMoveScale(0.12); // the tracks are large in world units, like Ridge Racer's
    flycam.setEnabled(true);
    stage.fly = flycam;

    const tris = countTris(world);
    stage.hud = `${item.name} · ${tris.toLocaleString()} triangles · ${flyHint}`;

    stage.onFrame = (camPos, dt) => flycam.update(dt);
    stage.disposePlugin = () => {
      flycam.dispose();
      controls.autoRotate = false;
      for (const root of [world, env]) {
        root.traverse((o) => {
          if (o.geometry) o.geometry.dispose();
          if (o.material) {
            for (const m of Array.isArray(o.material) ? o.material : [o.material]) m.dispose();
          }
        });
      }
    };
    return world;
  },
};

function countTris(root) {
  let n = 0;
  root.traverse((o) => {
    const g = o.geometry;
    if (!g) return;
    n += (g.index ? g.index.count : g.attributes.position?.count || 0) / 3;
  });
  return Math.round(n);
}
