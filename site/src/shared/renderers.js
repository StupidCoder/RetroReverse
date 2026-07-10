// The renderer registry for manifest-driven 3-D viewers. Each manifest model entry carries a
// `kind`; a renderer plugin for that kind knows how to turn the entry into scene content. A
// plugin is `{ kind, async build({ item, base, stage }) { … } }` — it fetches whatever the
// entry points at, adds objects to the stage, fits the camera (stage.frame) and, if it needs
// bespoke post-processing, takes over the frame by setting stage.render. build() returns the
// primary Object3D it installed.
//
// resolveRenderer(kind, gameRenderers) returns a LOADER (an async fn resolving to a module whose
// default export is the plugin): the game's own `gameRenderers[kind]` (a lazy `() => import(...)`)
// takes precedence, otherwise the builtin for that kind, otherwise it throws.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { makeSprite } from './sprites3d.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));
let _tex = null;
const texLoader = () => (_tex || (_tex = new THREE.TextureLoader()));

// resolveObjects returns the placement list for a level item: `item.objects` when it is an
// inline array, otherwise the `objects` array fetched from `base + item.objectsFile` (or
// `item.objects` when that is a string path). Returns [] when the item carries no object layer,
// so the plain GLB-only mesh3d path is untouched.
async function resolveObjects(item, base) {
  if (Array.isArray(item.objects)) return item.objects;
  const file = item.objectsFile || (typeof item.objects === 'string' ? item.objects : null);
  if (!file) return [];
  try {
    const doc = await fetch(base + file).then((r) => r.json());
    return doc.objects || [];
  } catch {
    return [];
  }
}

// placeObjects is the shared object-layer helper: given a decoded `objects[]` list and the asset
// `base`, it places each object into a group added to the stage and returns that group. An object
// with `model` (a "models/x.glb" path) is GLTF-loaded and placed at its pos/rot; an object with a
// `sprite` spec (see shared/sprites3d.js) becomes a directional billboard placed at its pos, and
// its per-frame update is composed onto stage.onFrame. The composition accumulates the stage's
// per-frame dt into an elapsed time and calls every sprite's update(camera.position, t) after any
// pre-existing onFrame — multiple updaters (and a plugin's own hook) compose instead of clobbering.
export async function placeObjects({ objects, base, stage }) {
  const group = new THREE.Group();
  const updaters = [];

  const jobs = (objects || []).map(async (o) => {
    const pos = o.pos || [0, 0, 0];
    if (o.model) {
      const gltf = await gltfLoader().loadAsync(base + o.model);
      const node = gltf.scene;
      node.position.set(pos[0], pos[1] || 0, pos[2] || 0);
      if (o.rot) node.rotation.set(o.rot[0] || 0, o.rot[1] || 0, o.rot[2] || 0); // radians
      group.add(node);
    } else if (o.sprite && typeof o.sprite === 'object') {
      const spec = o.sprite;
      const texture = await texLoader().loadAsync(base + spec.sheet);
      const { object3d, update } = await makeSprite(spec, texture);
      object3d.position.set(pos[0], pos[1] || 0, pos[2] || 0);
      group.add(object3d);
      updaters.push(update);
    }
  });
  await Promise.all(jobs);

  if (updaters.length) {
    const prev = stage.onFrame;
    let elapsed = 0;
    stage.onFrame = (camPos, dt) => {
      if (prev) prev(camPos, dt);
      elapsed += dt || 0;
      for (const u of updaters) u(camPos, elapsed);
    };
  }

  stage.add(group);
  return group;
}

// Builtin `mesh3d`: load a plain GLB (base + item.file), drop its scene onto the stage, fit the
// camera to it, and hand it back. If the item carries an object layer (`objects[]` inline or an
// `objectsFile`), place it too via placeObjects.
//
// If the GLB carries animations, a mixer plays ONE clip on loop (a model with several clips —
// Pilotwings pilots have up to ten — would tear if they all drove the same nodes at once),
// composed onto stage.onFrame so it stacks with any other per-frame updater. When there is more
// than one clip the stage gets a `cycleClip()` the Studio can wire to a control; the HUD names the
// current clip. The default renderer.render(scene, camera) draws everything, so it uses the stage's
// ordinary pipeline.
export const mesh3d = {
  kind: 'mesh3d',
  async build({ item, base, stage }) {
    const gltf = await gltfLoader().loadAsync(base + item.file);
    const obj = gltf.scene;
    stage.add(obj);
    stage.frame(obj);
    const clips = gltf.animations || [];
    if (clips.length) {
      const mixer = new THREE.AnimationMixer(obj);
      let cur = -1;
      const play = (i) => {
        i = ((i % clips.length) + clips.length) % clips.length;
        if (i === cur) return;
        mixer.stopAllAction();
        mixer.clipAction(clips[i]).reset().play();
        cur = i;
        stage.hud = clips.length > 1 ? `${item.name} · ${clips[i].name} (${i + 1}/${clips.length}) · click to cycle` : null;
      };
      play(0);
      const prev = stage.onFrame;
      stage.onFrame = (camPos, dt) => { if (prev) prev(camPos, dt); mixer.update(dt); };
      if (clips.length > 1) {
        stage.cycleClip = () => play(cur + 1);
        // Click (not drag) anywhere on the model to advance to the next clip.
        const canvas = stage.canvas;
        let dx = 0, dy = 0;
        const down = (e) => { dx = e.clientX; dy = e.clientY; };
        const up = (e) => { if (Math.hypot(e.clientX - dx, e.clientY - dy) <= 5) play(cur + 1); };
        canvas.addEventListener('pointerdown', down);
        canvas.addEventListener('pointerup', up);
        stage.disposePlugin = () => {
          canvas.removeEventListener('pointerdown', down);
          canvas.removeEventListener('pointerup', up);
          stage.cycleClip = null;
        };
      }
    }
    const objects = await resolveObjects(item, base);
    if (objects.length) await placeObjects({ objects, base, stage });
    return obj;
  },
};

// Loaders for the builtin kinds, in the same `() => Promise<module>` shape as a game's lazy
// `() => import('...')`, so resolveRenderer can treat both uniformly.
const BUILTINS = {
  mesh3d: async () => ({ default: mesh3d }),
};

export function resolveRenderer(kind, gameRenderers = {}) {
  const loader = (gameRenderers && gameRenderers[kind]) || BUILTINS[kind];
  if (!loader) throw new Error(`renderers: unknown renderer kind "${kind}"`);
  return loader;
}
