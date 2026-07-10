// Pilotwings 64 world renderer — the `pw-world` plugin for the shared
// Viewer3D/Stage3D seam.
//
// A world is one terrain GLB (webexport assembles it from the UVCT chunks its
// UVTR grid names) plus, optionally, two things the game keeps separate and so
// do we:
//
//   - an OBJECT SET (item.objectsFile). Every object a terrain chunk places
//     carries a 16-bit mask, and the engine draws it only where the mask meets
//     a per-scene selector. Each set is one bit of that mask, so the browse list
//     offers a world's sixteen dressings side by side instead of piling every
//     ring, balloon and target into one crowded island.
//   - the WATER plane (item.waterFile), a single flat quad far larger than the
//     island. It is loaded after the camera has been framed on the terrain, or
//     the fit would zoom out to the ocean's ±12,288 units and lose the world.
//
// Each placement carries a full 4x4 matrix rather than a position and a yaw:
// the object's own pose block gives it a scale as well as a rotation, and
// decomposing that into Euler angles would only throw the scale away.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { installPicker } from '../shared/objinfo.js';

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

export default {
  kind: 'pw-world',
  async build({ item, base, stage }) {
    const terrain = (await gltfLoader().loadAsync(base + item.file)).scene;
    stage.add(terrain);
    stage.frame(terrain); // before the water: the ocean dwarfs the island

    // ---- objects: one group, so the layer toggle hides them together ----
    const objectsGroup = new THREE.Group();
    stage.add(objectsGroup);
    const placed = []; // { object3d, obj } for the picker

    // Many placements share a model (35 palm trees, 8 rings); load each GLB once.
    const cache = new Map();
    const loadObj = async (url) => {
      if (!cache.has(url)) cache.set(url, gltfLoader().loadAsync(url).then((g) => g.scene));
      return (await cache.get(url)).clone(true);
    };

    let placements = [];
    if (item.objectsFile) {
      try {
        placements = (await fetch(base + item.objectsFile).then((r) => r.json())).objects || [];
      } catch { placements = []; }
    }
    await Promise.all(placements.map(async (o) => {
      const node = await loadObj(base + o.model);
      if (o.mat && o.mat.length === 16) {
        node.matrixAutoUpdate = false;
        node.matrix.fromArray(o.mat); // column-major, as glTF stores them
      } else if (o.pos) {
        node.position.set(o.pos[0], o.pos[1] || 0, o.pos[2] || 0);
      }
      objectsGroup.add(node);
      placed.push({ object3d: node, obj: o });
    }));

    // ---- water ----
    let water = null;
    if (item.waterFile) {
      water = (await gltfLoader().loadAsync(base + item.waterFile)).scene;
      stage.add(water);
    }

    stage.hud = `${item.name} · ${placements.length} objects`;

    let objectsOn = true;
    stage.setLayer = (id, on) => {
      if (id === 'objects') { objectsOn = on; objectsGroup.visible = on; }
      else if (id === 'water' && water) water.visible = on;
    };

    // ---- click-to-inspect ----
    const picker = installPicker(
      { el: stage.el, canvas: stage.canvas, camera: stage.camera },
      () => placed,
      {
        enabled: () => objectsOn,
        resolve: (o) => ({
          title: `Model ${String(o.uvmd).padStart(3, '0')}`,
          subtitle: `object type ${o.type} · mask ${o.mask}`,
          body: [
            `Drawn from UVMD resource ${o.uvmd} (ordinal ${o.type}) — the object's type IS the model's ordinal.`,
            `Placed by terrain chunk UVCT ${o.chunk}, in grid cell (${o.cell[0]}, ${o.cell[1]}), at`,
            `game-space (${o.pos.map((v) => v.toFixed(1)).join(', ')}).`,
            `Its mask ${o.mask} lists every scene the object appears in; this set is one bit of it.`,
          ].join(' '),
        }),
      },
    );

    stage.disposePlugin = () => { picker(); stage.setLayer = null; };
    return terrain;
  },
};
