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
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

// Builtin `mesh3d`: load a plain GLB (base + item.file), drop its scene onto the stage, fit the
// camera to it, and hand it back. The default renderer.render(scene, camera) draws it (the plugin
// installs no stage.render), so it uses the stage's ordinary lit/unlit pipeline.
export const mesh3d = {
  kind: 'mesh3d',
  async build({ item, base, stage }) {
    const gltf = await gltfLoader().loadAsync(base + item.file);
    const obj = gltf.scene;
    stage.add(obj);
    stage.frame(obj);
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
