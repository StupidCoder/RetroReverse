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

    // Floor groups: each shell lands in the group of its lowest storey, so
    // the floor toggles peel the dollhouse layer by layer.
    const floors = {
      attic: new THREE.Group(),
      first: new THREE.Group(),
      ground: new THREE.Group(),
      basement: new THREE.Group(),
    };
    for (const g of Object.values(floors)) group.add(g);
    const floorOf = (minY) => (minY < -300 ? 'basement' : minY < 300 ? 'ground' : minY < 900 ? 'first' : 'attic');

    stage.setLayer = (id, on) => {
      if (id === 'exterior') exterior.visible = !!on;
      else if (floors[id]) floors[id].visible = !!on;
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
          let dest = exterior;
          if (!EXTERIOR.has(entry.room)) {
            const box = new THREE.Box3().setFromObject(gltf.scene);
            dest = floors[floorOf(box.min.y)];
          }
          dest.add(gltf.scene);
          const res = await furnishRoom({ loader, base, entry, group: dest, cache, mixers, clickable });
          placed += res.placed;
          animated += res.animated;
        } catch { /* a missing room leaves a gap, not a broken build */ }
        loaded++;
        hud();
      }
    };
    // Doors: the DOL-side list — each record an opening between two rooms;
    // leafed ones get their door model hinged into it. The record's position
    // is the opening's centre; the leaf model is 200x300 with the hinge at
    // its local x=0 edge, so a single leaf sits at +width/2 (hinge on the
    // right jamb) and a double adds a mirrored leaf on the left. Axis 2
    // turns the plane onto x; axis 4 lays it flat (a trapdoor). Doors click
    // open and closed like the furniture (the swing clips ride in the GLB).
    const doorWatch = [];
    const placeDoors = async () => {
      for (const dr of placements.doors || []) {
        if (!dr.model) continue;
        if (!cache[dr.model]) {
          cache[dr.model] = loader.loadAsync(base + 'models/' + dr.model).then((g) => g, () => null);
        }
        const proto = await cache[dr.model];
        if (!proto) continue;
        const holder = new THREE.Group();
        holder.position.set(dr.pos[0], dr.pos[1], dr.pos[2]);
        if (dr.axis === 2) holder.rotation.y = Math.PI / 2;
        if (dr.axis === 4) holder.rotation.x = Math.PI / 2;
        // The leaf GLB is centred on its opening (the mesh hangs -200..0 off
        // a node at +100, putting the hinge node on the right edge): a
        // single leaf sits at the record position; a double's two leaves
        // shift half an opening each way, mirrored, hinges on the jambs.
        const w = (dr.axis === 2 ? dr.size[2] : dr.size[0]) || 200;
        const h = dr.size[1] || 300;
        const off = Math.max(0, (w - 200) / 2);
        const leaves = dr.double ? [[off, 1], [-off, -1]] : [[0, 1]];
        for (const [x, sx] of leaves) {
          const inst = proto.scene.clone(true);
          const leaf = new THREE.Group();
          leaf.position.set(x, -h / 2, 0);
          leaf.scale.set(sx, 1, 1);
          leaf.add(inst);
          holder.add(leaf);
          // The swing clips are a complete open-and-shut cycle (the game
          // plays them as Luigi walks through); clicking holds the leaf at
          // the swing's apex instead, and clicking again runs it back shut.
          const clips = proto.animations || [];
          const clip = clips[clips.length - 1];
          if (clip) {
            const mixer = new THREE.AnimationMixer(inst);
            const action = mixer.clipAction(clip);
            action.loop = THREE.LoopOnce;
            action.clampWhenFinished = true;
            const apex = clip.duration / 2;
            leaf.userData.toggle = () => {
              const open = leaf.userData.open;
              action.timeScale = open ? -1 : 1;
              if (!open && action.time >= apex) action.time = 0;
              if (open && action.time <= 0) action.time = apex;
              action.paused = false;
              action.enabled = true;
              action.play();
              leaf.userData.open = !open;
            };
            doorWatch.push({ action, apex, leaf });
            mixers.push(mixer);
            clickable.push(leaf);
          }
        }
        floors[floorOf(dr.pos[1] - h / 2)].add(holder);
      }
    };

    const workers = [];
    for (let i = 0; i < 6; i++) workers.push(worker());
    workers.push(placeDoors());
    // The build returns once the first few rooms are visible; the rest
    // keep streaming in the background.
    await Promise.race([Promise.all(workers), new Promise((r) => setTimeout(r, 1500))]);

    const disposeClicks = wireVacuum(stage, clickable);
    const prev = stage.onFrame;
    stage.onFrame = (camPos, dt) => {
      if (prev) prev(camPos, dt);
      flycam.update(dt);
      for (const m of mixers) m.update(dt || 0);
      // Hold opening doors at the swing's apex.
      for (const w of doorWatch) {
        if (w.leaf.userData.open && w.action.timeScale > 0 && w.action.time >= w.apex) {
          w.action.time = w.apex;
          w.action.paused = true;
        }
      }
    };
    stage.disposePlugin = () => {
      flycam.dispose();
      disposeClicks();
      next = rooms.length; // stop the streamers
    };
    return group;
  },
};
