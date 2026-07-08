// Stunt Car Racer track renderer — the `stunt-track` plugin for the shared Viewer3D/Stage3D seam.
// The geometry is entirely the engine's own, decoded purely from the disk in Go (package track,
// Part IV) and verified against the original on our m68k core. Each track is the game's own baked
// polygon model (the race-setup bake $65BEC, reimplemented in Go and verified byte-exact by
// cmd/modeloracle): per rung, the piece-shape (x,z) vertex pairs read by $5C6C4 rotated by the
// section quadrant, placed at the section's 16x16 grid cell exactly as the game's per-frame draw
// does (one cell = $800 units) — world = cell*$800 + local. No fit, no spline, no accumulation:
// straights are straight, arcs are the real arcs, and joints meet because the data says so.
// Heights are the full-precision rail heights (their difference is the real camber); rung LINES
// are drawn only where the game's decimated model has a polygon edge. Rendered as a hidden-line
// wireframe (invisible depth fill + colour LineSegments, the Marble Madness slope-viewer technique).
//
// Each circuit is presented as a level: you fly through it with the shared FlyCam (WASD to move
// relative to the view, arrow keys to look; mouse drag still orbits), the same free-flight controls
// the other 3-D level viewers use. The geometry build and the fly-cam setup are moved VERBATIM from
// the old TrackViewer.loadLevel — the plugin is only a new shell around exactly the same visuals.
import * as THREE from 'three';
import { FlyCam, flyHint } from '../shared/flycam.js';

export default {
  kind: 'stunt-track',
  async build({ item, base, stage }) {
    // One circuit's JSON — the same per-circuit object the old TrackViewer.loadLevel consumed:
    // track = { name, sections, finishIdx, nodes, rungs }.
    const track = await fetch(base + item.file).then((r) => r.json());

    const { scene, camera, controls } = stage;
    // The engine's night-sky clear colour (was renderer.setClearColor(0x0a0d12, 1) in TrackViewer).
    scene.background = new THREE.Color(0x0a0d12);

    const group = new THREE.Group();

    // track.rungs[i][k] = [Lx,Lz,Rx,Rz,Lh,Rh,flags]: the game's own baked polygon
    // model in ABSOLUTE plan coordinates (cell*$800 + $5C6C4 local vertex), verified
    // byte-exact against the engine's bake $65BEC (cmd/modeloracle). flags: 1 = a rung
    // the game draws, 2 = hidden gap-piece end (not in the game's model), 4 = crease,
    // 8 = the finish line. Rung 0 of every section duplicates the previous section's
    // last rung (the bake skips it too), and hidden rungs are dropped exactly as the
    // engine drops them from its record strip.
    const raw = [];
    const n = track.rungs.length;
    for (let i = 0; i < n; i++) {
      const rs = track.rungs[i];
      for (let k = 0; k < rs.length; k++) {
        if (k === 0 || (rs[k][6] & 2)) continue;
        raw.push({ v: rs[k], sec: i });
      }
    }
    let minX = Infinity, maxX = -Infinity, minZ = Infinity, maxZ = -Infinity, minH = Infinity;
    for (const { v } of raw) {
      minX = Math.min(minX, v[0], v[2]); maxX = Math.max(maxX, v[0], v[2]);
      minZ = Math.min(minZ, v[1], v[3]); maxZ = Math.max(maxZ, v[1], v[3]);
      minH = Math.min(minH, v[4], v[5]);
    }
    const cx = (minX + maxX) / 2, cz = (minZ + maxZ) / 2;
    const span = Math.max(maxX - minX, maxZ - minZ) || 1;
    const S = 8 / span; // fit the plan into ~8 units

    // Fixed rail-height -> plan-unit ratio shared across all tracks so relative relief
    // is honest: the Roller Coaster and Ski Jump really do tower over the gentle
    // circuits. Heights are the engine's full-precision rail heights (one grid cell =
    // $800 plan units; 4800 height units per cell-width of rise reads well); they sit
    // on the ground via the per-track minimum.
    const HK = S * 2048 / 4800;

    // Z is negated to mirror the plan back across the X axis: the baked model's Z runs
    // opposite to ours, so without this the whole course layout comes out flipped.
    const rings = raw.map(({ v, sec }) => ({
      l: { x: (v[0] - cx) * S, z: -(v[1] - cz) * S },
      r: { x: (v[2] - cx) * S, z: -(v[3] - cz) * S },
      hl: (v[4] - minH) * HK, hr: (v[5] - minH) * HK,
      fl: v[6], sec,
    }));
    const m = rings.length;
    const V = (p, y) => new THREE.Vector3(p.x, y, p.z);

    // Segment parity. The game draws horizontal lines across the road at certain rungs
    // (the baked polygon edges, flags bit 0); those lines split the road into segments.
    // Number the segments (a new one begins at each line) so the surface and walls can
    // alternate colour per segment. segPar[k] is the parity of the segment the quad
    // starting at rung k belongs to; the first segment is parity 0.
    const segPar = new Array(m);
    for (let k = 0, s = 0; k < m; k++) {
      if (k > 0 && (rings[k].fl & 1)) s++;
      segPar[k] = s & 1;
    }

    // The road surface (the ribbon). Its colour alternates per segment between
    // (166,162,129) and (195,192,166). It also does the hidden-line removal for the
    // rail/rung lines drawn on top: it writes depth, and its polygon offset pushes it
    // back so the coplanar lines sit in front of it.
    const ROAD_A = new THREE.Color(0xa6a281); // (166,162,129) darker
    const ROAD_B = new THREE.Color(0xc3c0a6); // (195,192,166) lighter
    const fpos = [], fcol = [];
    const quad = (a, b, c, d, col) => {
      fpos.push(a.x, a.y, a.z, b.x, b.y, b.z, c.x, c.y, c.z, b.x, b.y, b.z, d.x, d.y, d.z, c.x, c.y, c.z);
      for (let i = 0; i < 6; i++) fcol.push(col.r, col.g, col.b);
    };
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m];
      quad(V(a.l, a.hl), V(a.r, a.hr), V(b.l, b.hl), V(b.r, b.hr), segPar[k] ? ROAD_B : ROAD_A);
    }
    const fgeom = new THREE.BufferGeometry();
    fgeom.setAttribute('position', new THREE.Float32BufferAttribute(fpos, 3));
    fgeom.setAttribute('color', new THREE.Float32BufferAttribute(fcol, 3));
    const fill = new THREE.Mesh(fgeom, new THREE.MeshBasicMaterial({
      vertexColors: true,
      polygonOffset: true, polygonOffsetFactor: 1, polygonOffsetUnits: 1,
      side: THREE.DoubleSide,
    }));
    group.add(fill);

    // The two side rails: lines along each edge of the road that alternate the game's
    // curb colours per rung segment. (The horizontal cross-lines that used to divide the
    // segments are gone — the alternating surface colour shows the segments now.)
    const lpos = [], lcol = [];
    const RAIL_A = new THREE.Color(0xf4f17e); // (244,241,126)
    const RAIL_B = new THREE.Color(0x280a0a); // (40,10,10)
    const seg = (p, q, c) => {
      lpos.push(p.x, p.y, p.z, q.x, q.y, q.z);
      lcol.push(c.r, c.g, c.b, c.r, c.g, c.b);
    };
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m];
      const railCol = (k % 2 === 0) ? RAIL_A : RAIL_B; // alternating curb stripes
      seg(V(a.l, a.hl), V(b.l, b.hl), railCol); // left rail
      seg(V(a.r, a.hr), V(b.r, b.hr), railCol); // right rail
    }
    const lgeom = new THREE.BufferGeometry();
    lgeom.setAttribute('position', new THREE.Float32BufferAttribute(lpos, 3));
    lgeom.setAttribute('color', new THREE.Float32BufferAttribute(lcol, 3));
    group.add(new THREE.LineSegments(lgeom, new THREE.LineBasicMaterial({ vertexColors: true })));

    // Solid side walls: the track is an elevated ribbon walled along each edge, one
    // on the left rail and one on the right (as in the game), replacing the old centre
    // support lines. Each wall is a vertical quad strip that follows its rail and drops
    // to the ground (y=0).
    // The wall colour alternates per segment in step with the road: the darker road
    // (166,162,129) gets white walls, the lighter road (195,192,166) gets red walls.
    const WALL_A = new THREE.Color(0xffffff); // (255,255,255) with the darker road
    const WALL_B = new THREE.Color(0x784238); // (120,66,56) with the lighter road
    const wpos = [], wcol = [];
    // one wall quad between rail points a (top height ha) and b (top height hb):
    // top edge follows the rail, bottom edge sits on the ground.
    const wallSeg = (a, ha, b, hb, col) => {
      wpos.push(a.x, ha, a.z, a.x, 0, a.z, b.x, hb, b.z);
      wpos.push(b.x, hb, b.z, a.x, 0, a.z, b.x, 0, b.z);
      for (let i = 0; i < 6; i++) wcol.push(col.r, col.g, col.b);
    };
    for (let k = 0; k < m; k++) {
      const a = rings[k], b = rings[(k + 1) % m];
      const wc = segPar[k] ? WALL_B : WALL_A;
      wallSeg(a.l, a.hl, b.l, b.hl, wc); // left wall
      wallSeg(a.r, a.hr, b.r, b.hr, wc); // right wall
    }
    const wgeom = new THREE.BufferGeometry();
    wgeom.setAttribute('position', new THREE.Float32BufferAttribute(wpos, 3));
    wgeom.setAttribute('color', new THREE.Float32BufferAttribute(wcol, 3));
    const walls = new THREE.Mesh(wgeom, new THREE.MeshBasicMaterial({
      vertexColors: true, side: THREE.DoubleSide,
      polygonOffset: true, polygonOffsetFactor: 1, polygonOffsetUnits: 1,
    }));
    group.add(walls);

    // Vertical strut lines on the walls: one at every rung, from the road surface down
    // to the ground, colour (40,10,10). They sit exactly on the wall plane — the wall's
    // positive polygon offset (above) pushes the wall back so these coplanar lines render
    // in front of it without z-fighting, the same trick the road uses for its rung lines.
    const spos = [];
    for (let k = 0; k < m; k += 2) { // every other rung: keep, drop, keep, ...
      const a = rings[k];
      spos.push(a.l.x, a.hl, a.l.z, a.l.x, 0, a.l.z); // left wall
      spos.push(a.r.x, a.hr, a.r.z, a.r.x, 0, a.r.z); // right wall
    }
    const sgeom = new THREE.BufferGeometry();
    sgeom.setAttribute('position', new THREE.Float32BufferAttribute(spos, 3));
    group.add(new THREE.LineSegments(sgeom, new THREE.LineBasicMaterial({ color: 0x280a0a }))); // (40,10,10)

    // Start marker: a white strip painted across the road surface at the start (the
    // first few rungs, spanning rail to rail), lifted onto the surface with a negative
    // polygon offset so it sits on top of the road.
    const START_RUNGS = 2;
    const stpos = [];
    for (let k = 0; k < START_RUNGS && k + 1 < m; k++) {
      const a = rings[k], b = rings[k + 1];
      stpos.push(a.l.x, a.hl, a.l.z, a.r.x, a.hr, a.r.z, b.l.x, b.hl, b.l.z);
      stpos.push(a.r.x, a.hr, a.r.z, b.r.x, b.hr, b.r.z, b.l.x, b.hl, b.l.z);
    }
    const stgeom = new THREE.BufferGeometry();
    stgeom.setAttribute('position', new THREE.Float32BufferAttribute(stpos, 3));
    group.add(new THREE.Mesh(stgeom, new THREE.MeshBasicMaterial({
      color: 0xffffff, side: THREE.DoubleSide,
      polygonOffset: true, polygonOffsetFactor: -1, polygonOffsetUnits: -1,
    })));

    stage.add(group);

    // Frame the whole circuit from a raised 3/4 angle so both the plan and the elevation
    // read, then hand control to the fly-cam to explore it.
    let maxH = 0;
    for (const r of rings) maxH = Math.max(maxH, r.hl, r.hr);
    const size = Math.max(8, maxH * 1.5); // world extent of the level (plan ~8 units + relief)
    controls.target.set(0, maxH * 0.35, 0);
    camera.position.set(size * 0.28, size * 0.55, size * 0.95);
    camera.near = 0.01; camera.far = 200; camera.updateProjectionMatrix();
    controls.update();

    // Free-flight camera (WASD move / arrow look), layered on the orbit controls. Bound to
    // the stage's camera/controls/element and published as stage.fly so the Studio's
    // KeyboardCamera (which checks v.fly.enabled and cedes the arrow keys to it) reaches it.
    const flycam = new FlyCam(camera, controls, stage.el);
    // Levels are explored with the free-flight controls (WASD/arrows, or the touch sticks),
    // like the other 3-D level viewers.
    flycam.setScale(size);
    flycam.setMoveScale(1.4);
    flycam.setEnabled(true);
    stage.fly = flycam;

    // Drive the fly-cam each frame; this game has no post-FX, so we do NOT set stage.render —
    // Stage3D's default renderer.render(scene, camera) draws the frame.
    stage.onFrame = (camPos, dt) => flycam.update(dt);

    // HUD detail line (as the old loadLevel set it).
    stage.hud = `${track.name} — ${track.sections} sections, ${m} rungs · ${flyHint}`;

    // Teardown: free this circuit's GPU resources and unhook the fly-cam's listeners/sticks
    // before the next item builds (the Viewer calls this before stage.clear()).
    stage.disposePlugin = () => {
      flycam.dispose();
      group.traverse((o) => {
        if (o.geometry) o.geometry.dispose();
        if (o.material) o.material.dispose();
      });
    };

    return group;
  },
};
