// The `lm-mansion` renderer: the whole mansion assembled — all 75 room
// shells at their true positions (the rooms are modelled in one
// mansion-global frame, so assembly is just loading them together) plus
// every furniture piece the game's placement database puts in them.
//
// Shells stream in a few at a time and appear as they arrive; furniture
// follows its room, with the piece cache shared across rooms so the eleven
// o_isu chairs download once. The giant exterior/roof shells are a layer —
// on for the building, off for the dollhouse cutaway. Flying and the
// click-to-vacuum interaction match the single-room viewer.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';
import { furnishRoom, wireVacuum } from './furnish.js';

// The shells that wrap the interiors: outer walls, roofs, the courtyard
// and boo-graveyard grounds. Identified by their bounds spanning the whole
// building.
const EXTERIOR = new Set(['room_11', 'room_16', 'room_23', 'room_37', 'room_59', 'room_60', 'room_72']);

export default {
  kind: 'lm-mansion',
  async build({ item, base, stage }) {
    const loader = new GLTFLoader();
    const group = new THREE.Group();
    const exterior = new THREE.Group();
    group.add(exterior);
    stage.add(group);

    const placements = await fetch(base + 'placements.json').then((r) => r.json());
    const rooms = placements.rooms || [];

    stage.setLayer = (id, on) => {
      if (id === 'exterior') exterior.visible = !!on;
    };

    // Camera: over the front lawn, looking into the foyer front. The
    // mansion spans x -4200..4200, z -5600..300, floors y -600..2100.
    const cam = item.camera || {};
    const pos = cam.pos || [0, 2600, 3600];
    const tgt = cam.target || [0, 300, -2000];
    stage.camera.position.set(pos[0], pos[1], pos[2]);
    stage.controls.target.set(tgt[0], tgt[1], tgt[2]);
    stage.camera.near = 10;
    stage.camera.far = 300000;
    stage.camera.updateProjectionMatrix();
    stage.controls.update();
    stage.controls.autoRotate = false;

    const flycam = new FlyCam(stage.camera, stage.controls, stage.el);
    flycam.setScale(9000);
    flycam.setMoveScale(0.45);
    flycam.setEnabled(true);
    stage.fly = flycam;

    const cache = {};
    const mixers = [];
    const clickable = [];
    let loaded = 0, placed = 0, animated = 0;
    const hud = () => {
      const bits = [];
      if (loaded < rooms.length) bits.push(`loading rooms ${loaded}/${rooms.length}`);
      else bits.push(`${rooms.length} rooms, ${placed} furniture pieces — the game's own placements`);
      if (animated) bits.push(`click to vacuum furniture open`);
      bits.push(flyHint);
      stage.hud = `${item.name} · ${bits.join(' · ')}`;
    };
    hud();

    // Stream shells with bounded concurrency; furnish each room as its
    // shell lands.
    let next = 0;
    const worker = async () => {
      for (;;) {
        const i = next++;
        if (i >= rooms.length) return;
        const entry = rooms[i];
        try {
          const gltf = await loader.loadAsync(base + 'models/' + entry.model);
          (EXTERIOR.has(entry.room) ? exterior : group).add(gltf.scene);
          const res = await furnishRoom({ loader, base, entry, group, cache, mixers, clickable });
          placed += res.placed;
          animated += res.animated;
        } catch { /* a missing room leaves a gap, not a broken build */ }
        loaded++;
        hud();
      }
    };
    const workers = [];
    for (let i = 0; i < 6; i++) workers.push(worker());
    // The build returns once the first few rooms are visible; the rest
    // keep streaming in the background.
    await Promise.race([Promise.all(workers), new Promise((r) => setTimeout(r, 1500))]);

    const disposeClicks = wireVacuum(stage, clickable);
    const prev = stage.onFrame;
    stage.onFrame = (camPos, dt) => {
      if (prev) prev(camPos, dt);
      flycam.update(dt);
      for (const m of mixers) m.update(dt || 0);
    };
    stage.disposePlugin = () => {
      flycam.dispose();
      disposeClicks();
      next = rooms.length; // stop the streamers
    };
    return group;
  },
};
