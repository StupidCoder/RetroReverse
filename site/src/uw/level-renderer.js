// Ultima Underworld level renderer — the `uw-level` plugin for the shared Viewer3D/Stage3D seam.
//
// The dungeon mesh is UW's own tile geometry, reverse-engineered from LEV.ARK and reimplemented in
// Go (extract/levgeo, exported by cmd/levexport): floors, ceilings and walls, each carrying its real
// W64.TR / F32.TR texture through a per-level texture list. The exporter emits it INLINE (positions,
// UVs, material draw-groups, and each texture as a data-URI PNG) in the level JSON's `mesh` block —
// not as a GLB — so this plugin owns the mesh build instead of the builtin `mesh3d` GLB loader.
//
// The mesh-building here is the old standalone UW viewer's build, ported verbatim (BufferGeometry
// from positions/uvs, one material per texture, per-group draw ranges). What it drops is that
// viewer's bespoke creature/billboard/additive-glow code: the level's objects (creatures, doors,
// items) are now placed through the SHARED object layer (shared/renderers.js → placeObjects), which
// renders each sprite via shared/sprites3d.js and composes their per-frame updates onto stage.onFrame.
// No hand-rolled billboarding lives here anymore.
//
// You explore the level first-person with the shared FlyCam (WASD move / arrow look; mouse drag still
// orbits), published as stage.fly so the Studio's KeyboardCamera cedes the arrow keys to it — the same
// free-flight control the other 3-D level viewers use.
import * as THREE from 'three';
import { FlyCam, flyHint } from '../shared/flycam.js';
import { placeObjects } from '../shared/renderers.js';

// The Abyss's near-black clear colour and the fog that swallows distance — carried from the old
// viewer so the first-person look is preserved.
const ABYSS = 0x05060a;

export default {
  kind: 'uw-level',
  async build({ item, base, stage }) {
    const data = await fetch(base + item.file).then((r) => r.json());
    const mesh = data.mesh || data; // the geometry block (positions/uvs/groups/textures)

    const { scene, camera, controls } = stage;
    // The Abyss atmosphere: near-black background + distance fog (matches the old UW viewer).
    scene.background = new THREE.Color(ABYSS);
    scene.fog = new THREE.Fog(ABYSS, 20, 70);

    // --- the dungeon mesh (ported verbatim from the old UW viewer) ---
    // One texture (material) per entry in mesh.textures; UW textures are tiny, keep them crisp and
    // let walls tile vertically (a wall's V runs 0..height). textures are data-URI PNG strings.
    const texLoader = new THREE.TextureLoader();
    const materials = mesh.textures.map((uri) => {
      const tex = texLoader.load(uri);
      tex.wrapS = tex.wrapT = THREE.RepeatWrapping;
      tex.magFilter = THREE.NearestFilter;
      tex.minFilter = THREE.LinearMipmapLinearFilter;
      tex.colorSpace = THREE.SRGBColorSpace;
      // Default double-sided (as before); singleSided groups override the referenced material to
      // FrontSide below. Ceilings are wound normal-down, so drawing them FrontSide culls them when
      // seen from above (you can peer into rooms from a bird's-eye view) while they still cap the
      // room when you're inside looking up.
      return new THREE.MeshBasicMaterial({ map: tex, side: THREE.DoubleSide });
    });

    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.Float32BufferAttribute(mesh.positions, 3));
    geo.setAttribute('uv', new THREE.Float32BufferAttribute(mesh.uvs, 2));
    for (const g of mesh.groups) {
      geo.addGroup(g.start, g.count, g.material);
      // Honor per-group single-sidedness: a singleSided group's material is back-face culled
      // (THREE.FrontSide); every other material stays double-sided. (UW's ceilings carry their own
      // ceiling textures, so this is consistent per material index.)
      if (g.singleSided && materials[g.material]) materials[g.material].side = THREE.FrontSide;
    }
    geo.computeBoundingBox();

    // Keep the mesh in the level's own world coordinates (NOT centred): the spawn point and the
    // shared object layer both place things in these same world units, so centring the mesh would
    // shift it off the objects. The old standalone viewer centred and then re-offset every sprite by
    // hand — placeObjects has no such offset, so world coordinates keep mesh + objects aligned.
    const dungeon = new THREE.Mesh(geo, materials);
    stage.add(dungeon);

    // --- the level's objects, via the SHARED object layer ---
    // Creatures / doors / items become directional-billboard sprites (including additive translucents)
    // through placeObjects → sprites3d; their per-frame updates compose onto stage.onFrame. No bespoke
    // billboard/additive code here. Objects that are pure props (props.text writings, no sprite) carry
    // no billboard and are simply not drawn: the old viewer's click-to-inspect InfoCard (raycasting
    // invisible pick boxes tagged with each object's name/text) is DROPPED in this migration — it was
    // bespoke pick geometry outside the shared object layer. A future pass could re-add it generically.
    if (item.objects) {
      const doc = await fetch(base + item.objects).then((r) => r.json());
      await placeObjects({ objects: doc.objects || [], base, stage });
    }

    // --- first-person camera ---
    // Preserve the old viewer's first-person feel: a wider FOV and a tight near/far to match the fog.
    camera.fov = 60;
    camera.near = 0.01;
    camera.far = 500;
    const r = geo.boundingBox.getSize(new THREE.Vector3()).length() || 40;
    if (data.spawn && data.spawn.pos) {
      const [sx, sy, sz] = data.spawn.pos;
      const [dx, dy, dz] = data.spawn.dir || [1, 0, 0]; // initial look direction
      camera.position.set(sx, sy, sz);
      controls.target.set(sx + dx, sy + dy, sz + dz);
    } else {
      // No spawn exported: fall back to an angled overview of the level's centre.
      const c = geo.boundingBox.getCenter(new THREE.Vector3());
      camera.position.set(c.x, c.y + r * 0.42, c.z + r * 0.62);
      controls.target.copy(c);
    }
    camera.updateProjectionMatrix();
    controls.autoRotate = false;
    controls.update();

    // Free-flight camera (WASD move / arrow look), layered on the orbit controls and published as
    // stage.fly so the Studio's KeyboardCamera reaches it. UW's levels are vast, so movement runs at
    // quarter speed (look speed unchanged) — carried from the old viewer.
    const flycam = new FlyCam(camera, controls, stage.el);
    flycam.setScale(r);
    flycam.setMoveScale(0.25);
    flycam.setEnabled(true);
    stage.fly = flycam;

    // Drive the fly-cam each frame BEFORE any sprite updater placeObjects composed onto onFrame.
    const prev = stage.onFrame;
    stage.onFrame = (camPos, dt) => { flycam.update(dt); if (prev) prev(camPos, dt); };

    const tris = (mesh.positions.length / 9) | 0;
    stage.hud = `${tris} triangles · ${mesh.textures.length} textures · ${flyHint}`;

    // Teardown: free this level's GPU resources, drop the Abyss fog, and unhook the fly-cam's
    // listeners/sticks before the next item builds (the Viewer calls this before stage.clear()).
    stage.disposePlugin = () => {
      flycam.dispose();
      scene.fog = null;
      geo.dispose();
      for (const m of materials) { m.map?.dispose(); m.dispose(); }
    };

    return dungeon;
  },
};
