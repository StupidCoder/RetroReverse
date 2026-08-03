// level3d.js — the 3D scene engine: layers (base/toggle/exclusive, camera
// attachment, paint order), room-graph streaming with area peel toggles,
// placements of any object type, level variants, routes and behaviours,
// onClick interactions, and the cutscene scripts. All of it from the level
// document; no per-game code.

import { THREE, Stage, FlyCam, ObjectLibrary, loadGLB, applyWireframe, applyTexFilter, applyTransform, flyHint } from './engine3d.js';
import { CutscenePlayer } from './cutscene.js';
import { PanInput } from './pancam.js';
import { ARSession, arSupported } from './xr.js';

export async function mount(ctx, doc) {
  const { stage: el, game, asset, params } = ctx;
  const docPath = asset.file;
  const cam = doc.camera || {};

  // ctx.signal aborts everything this mount streams (layers, room shells,
  // placement models). The app shell fires it the moment a navigation
  // supersedes this view — even mid-mount: navigating away from a
  // half-loaded mansion must stop the downloads, not let them fight the
  // next scene's.
  const signal = ctx.signal ?? new AbortController().signal;

  const stage = new Stage(el, cam);
  if (doc.scene?.background) stage.scene.background = new THREE.Color(doc.scene.background);
  if (doc.scene?.fog) {
    stage.scene.fog = new THREE.Fog(new THREE.Color(doc.scene.fog.color), doc.scene.fog.near, doc.scene.fog.far);
  }
  stage.scene.add(new THREE.AmbientLight(0xffffff, 2.2));
  const key = new THREE.DirectionalLight(0xffffff, 1.2);
  key.position.set(1, 2, 1.4);
  stage.scene.add(key);

  let fly = null;
  if (cam.mode === 'fly') fly = new FlyCam(stage, cam.fly?.speed);
  if (cam.mode === 'orbit' && cam.orbit) {
    stage.controls.autoRotate = !!cam.orbit.autoRotate;
    stage.controls.autoRotateSpeed = (cam.orbit.autoRotateSpeed || 0.3) * 6;
    if (cam.orbit.minDist) stage.controls.minDistance = cam.orbit.minDist;
    if (cam.orbit.maxDist) stage.controls.maxDistance = cam.orbit.maxDist;
  }
  let panInput = null;
  if (cam.mode === 'pan2d') {
    // Levels that are 3D geometry but 2D in spirit (Loco Roco): the camera
    // faces the plane and never rotates — drag pans, wheel/pinch zooms, and
    // (Stage) it looks through an orthographic frustum, so scrolling slides
    // the stage past the window instead of swinging it.
    const c = stage.controls;
    c.enableRotate = false;
    c.screenSpacePanning = true;
    c.zoomToCursor = true;
    c.mouseButtons = { LEFT: THREE.MOUSE.PAN, MIDDLE: THREE.MOUSE.DOLLY, RIGHT: THREE.MOUSE.PAN };
    c.touches = { ONE: THREE.TOUCH.PAN, TWO: THREE.TOUCH.DOLLY_PAN };
    // An ortho camera dollies by camera.zoom, not by distance, so its limits
    // are the zoom ones; the distance limits would silently do nothing.
    if (stage.camera.isOrthographicCamera) { c.minZoom = 0.15; c.maxZoom = 60; }
    else if (cam.near) c.minDistance = cam.near * 4;

    // The same inertial scroll the tilemap views have: the arrows accelerate
    // into a glide and coast to a stop, at a speed measured in SCREEN pixels
    // so it feels identical at every zoom level.
    const right = new THREE.Vector3(), up = new THREE.Vector3();
    const worldPerPx = () => {
      const h = stage.el.clientHeight || 1;
      if (stage.camera.isOrthographicCamera) {
        return (stage.camera.top - stage.camera.bottom) / stage.camera.zoom / h;
      }
      const d = stage.camera.position.distanceTo(c.target);
      return (2 * d * Math.tan((stage.camera.fov * Math.PI) / 360)) / h;
    };
    panInput = new PanInput({
      onPan: (dx, dy) => {
        // dx/dy scroll the CONTENT (level2d's sign convention, +y down), so
        // the camera moves the opposite way on x and along +up for +dy.
        const s = worldPerPx();
        right.setFromMatrixColumn(stage.camera.matrix, 0).multiplyScalar(-dx * s);
        up.setFromMatrixColumn(stage.camera.matrix, 1).multiplyScalar(dy * s);
        const off = right.add(up);
        stage.camera.position.add(off);
        c.target.add(off);
      },
      onZoom: (f) => {
        if (stage.camera.isOrthographicCamera) {
          stage.camera.zoom = Math.max(c.minZoom, Math.min(c.maxZoom, stage.camera.zoom * f));
          stage.camera.updateProjectionMatrix();
        } else {
          const off = stage.camera.position.clone().sub(c.target);
          const d = Math.max(c.minDistance, Math.min(c.maxDistance, off.length() / f));
          stage.camera.position.copy(c.target).add(off.setLength(d));
        }
      },
    });
    stage.updaters.add((dt) => { if (stage.inputEnabled) panInput.step(dt); });
    // Grabbing the stage arrests a glide in progress.
    stage.canvas.addEventListener('pointerdown', () => panInput.stop());
  }

  const roots = []; // everything mountable, for wireframe toggle + disposal
  const dp = ctx.displayPanel;
  const hud = document.createElement('div');
  hud.className = 'hud';
  const hint = cam.mode === 'fly' ? flyHint
    : cam.mode === 'pan2d' ? 'drag or arrows to pan · wheel/pinch to zoom'
      : 'drag to orbit · wheel to zoom';
  hud.textContent = hint;
  el.appendChild(hud);

  // ---- layers -----------------------------------------------------------------
  const layerNodes = new Map(); // id -> {group, def}
  const exclusive = new Map();  // group name -> [layer ids]
  if (doc.scene?.layers?.length) {
    dp.section('Layers');
    await Promise.all(doc.scene.layers.map(async (ly) => {
      const gltf = await loadGLB(game.url(docPath, ly.file), signal);
      if (game.display.texFilter) applyTexFilter(gltf.scene, game.display.texFilter);
      const group = new THREE.Group();
      group.add(gltf.scene);
      group.visible = ly.visible !== false;
      group.userData.layer = ly; // role/attach/mode — read by contentBox() and AR mode
      applyLayerLook(gltf.scene, ly);
      stage.scene.add(group);
      roots.push(group);
      layerNodes.set(ly.id, { group, def: ly });
      if (ly.attach === 'camera' || ly.attach === 'cameraYaw') {
        stage.updaters.add((dt, camPos) => group.position.copy(camPos));
      }
    }));
    for (const ly of doc.scene.layers) {
      const rec = layerNodes.get(ly.id);
      if (!rec) continue;
      const mode = ly.mode || 'base';
      if (mode === 'toggle') {
        dp.toggle(ly.name || ly.id, rec.group.visible, (on) => { rec.group.visible = on; });
      } else if (mode.startsWith('exclusive:')) {
        const g = mode.slice(10);
        if (!exclusive.has(g)) exclusive.set(g, []);
        exclusive.get(g).push(ly.id);
        dp.radio(`excl-${g}`, ly.name || ly.id, rec.group.visible, () => {
          for (const id of exclusive.get(g)) layerNodes.get(id).group.visible = id === ly.id;
        });
      }
    }
  }

  // ---- rooms ---------------------------------------------------------------------
  const roomNodes = new Map();      // room id -> group
  const areaRooms = new Map();      // area id -> [room groups]
  const roomPlacements = new Map(); // room id -> [placement nodes] (filled below)
  let roomsLoaded = 0;
  if (doc.scene?.rooms) {
    const rm = doc.scene.rooms;
    dp.section('Areas');
    for (const a of rm.areas || []) areaRooms.set(a.id, []);
    // Progressive streaming with a small worker pool; the level is browsable
    // while shells arrive.
    const queue = [...rm.list];
    const hudTick = () => hud.textContent = `${roomsLoaded}/${rm.list.length} rooms · ${hint}`;
    const worker = async () => {
      for (;;) {
        const r = queue.shift();
        if (!r || signal.aborted) return;
        try {
          const gltf = await loadGLB(game.url(docPath, r.file), signal);
          if (signal.aborted) return;
          if (game.display.texFilter) applyTexFilter(gltf.scene, game.display.texFilter);
          const group = new THREE.Group();
          group.add(gltf.scene);
          group.userData.room = r;
          stage.scene.add(group);
          roots.push(group);
          roomNodes.set(r.id, group);
          if (r.area && areaRooms.has(r.area)) {
            areaRooms.get(r.area).push(group);
            group.visible = areaOn.get(r.area) !== false;
          }
          for (const pl of roomPlacements.get(r.id) || []) pl.visible = group.visible;
        } catch (e) {
          if (signal.aborted) return;
          console.warn('room failed', r, e);
        }
        roomsLoaded++;
        hudTick();
      }
    };
    const areaOn = new Map();
    for (const a of rm.areas || []) {
      areaOn.set(a.id, true);
      dp.toggle(a.name || a.id, true, (on) => {
        areaOn.set(a.id, on);
        for (const g of areaRooms.get(a.id) || []) g.visible = on;
        for (const [rid, group] of roomNodes) {
          const r = group.userData.room;
          if (r.area === a.id) for (const pl of roomPlacements.get(rid) || []) pl.visible = on;
        }
      });
    }
    const workers = Array.from({ length: rm.stream ? 6 : 1 }, worker);
    Promise.all(workers).then(() => hudTick());
  }

  // ---- variants ---------------------------------------------------------------------
  const variants = doc.variants || [];
  let activeVariant = params.get?.('variant') ?? params.variant;
  if (!activeVariant && variants.length) {
    activeVariant = (variants.find((v) => v.default) || variants[0]).id;
  }
  if (variants.length) {
    ctx.setVariants(variants, activeVariant, (id) => { activeVariant = id; applyVariant(); });
  }

  // ---- placements ---------------------------------------------------------------------
  const lib = new ObjectLibrary(game, signal);
  const placementById = new Map(); // id -> { pl, inst, node }
  const pickables = [];
  const routeById = new Map((doc.routes || []).map((r) => [r.id, r]));

  await Promise.all((doc.placements || []).map(async (pl) => {
    try {
      const inst = await lib.instance(pl.object);
      const node = new THREE.Group();
      node.add(inst.node);
      applyTransform(node, pl);
      node.userData.pick = { pl, inst };
      stage.scene.add(node);
      roots.push(node);
      placementById.set(pl.id, { pl, inst, node });
      pickables.push(node);
      if (pl.room != null) {
        if (!roomPlacements.has(pl.room)) roomPlacements.set(pl.room, []);
        roomPlacements.get(pl.room).push(node);
      }
      if (pl.layer && layerNodes.has(pl.layer)) {
        const lg = layerNodes.get(pl.layer).group;
        stage.updaters.add(() => { node.visible = lg.visible && node.userData.varOn !== false; });
      }
      if (pl.variants?.length) {
        node.userData.variants = pl.variants;
      }
      if (inst.playAnim && pl.anim) inst.playAnim(pl.anim);
      else if (inst.playAnim) {
        // Autoplay only an idle (looping) clip. One-shot clips are
        // interactions — autoplaying them made every door in the mansion
        // swing open and shut once at load.
        const idle = (inst.doc.animations || []).find((a) => (a.loop || 'loop') === 'loop' || a.loop === 'pingpong');
        if (idle) inst.playAnim(idle.id);
        else {
          // One-shot-only models (the mansion's doors): the GLB's rest
          // transforms are the bind file's parking positions, not a real
          // pose — the panels sat inside the wall until first clicked.
          // Hold frame 0 of the interaction clip instead (paused).
          const first = pl.onClick?.clip || (inst.doc.animations || [])[0]?.id;
          const h = first ? inst.playAnim(first) : null;
          if (h) {
            h.action.paused = true;
            h.action.time = 0;
            inst.mixer.update(0);
          }
        }
      }
      if (inst.update) stage.updaters.add((dt, cp, t) => { if (node.visible) inst.update(dt, cp, t); });
      wireBehavior(stage, node, pl, routeById);
    } catch (e) {
      if (!signal.aborted) console.warn('placement failed', pl.object, e);
    }
  }));

  // hoisted: the variant setter can fire while placements are still loading
  function applyVariant() {
    for (const { pl, node } of placementById.values()) {
      node.userData.varOn = !pl.variants?.length || pl.variants.includes(activeVariant);
      node.visible = node.userData.varOn;
    }
  }
  if (variants.length) applyVariant();

  // ---- interactions ---------------------------------------------------------------------
  const raycaster = new THREE.Raycaster();
  let card = null;
  const closeCard = () => { card?.remove(); card = null; };

  let down = null, lastTap = 0;
  stage.canvas.addEventListener('pointerdown', (e) => { down = { x: e.clientX, y: e.clientY }; });
  stage.canvas.addEventListener('pointerup', (e) => {
    // In an XR session the canvas is not what the viewer is looking at, and a
    // dom-overlay can still deliver pointer events — an infocard opened behind
    // the immersive view is invisible and unclosable.
    if (stage.renderer.xr.isPresenting) return;
    if (!down || Math.hypot(e.clientX - down.x, e.clientY - down.y) > 6) return;
    const now = performance.now();
    const dbl = now - lastTap < 320;
    lastTap = now;
    const hit = pick(e);
    closeCard();
    if (!hit) return;
    const { pl, inst } = hit.userData.pick;
    if (dbl) return showInfo(e, pl, inst);
    if (pl.onClick) return runAction(pl, hit, e);
    showInfo(e, pl, inst);
  });

  function pick(e) {
    const r = stage.canvas.getBoundingClientRect();
    const p = new THREE.Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1);
    raycaster.setFromCamera(p, stage.camera);
    for (const h of raycaster.intersectObjects(pickables, true)) {
      let o = h.object;
      while (o && !o.userData.pick) o = o.parent;
      if (o?.visible !== false && o?.userData.pick) return o;
    }
    return null;
  }

  function runAction(pl, node, e) {
    const oc = pl.onClick;
    if (oc.action === 'text') {
      showTextCard(e, oc.title || pl.name, oc.body);
      return;
    }
    if (oc.action !== 'animate') return; // unknown actions are no-ops by spec
    let target = node.userData.pick;
    if (oc.target && oc.target !== 'self') {
      target = placementById.get(parseInt(oc.target, 10)) || target;
    }
    const inst = target.inst;
    if (!inst.playAnim) return;
    const st = (node.userData.toggleState ||= { open: false, handle: null });
    if (!st.handle) {
      st.handle = inst.playAnim(oc.clip, { loop: 'hold' });
      if (!st.handle) return;
    }
    const { action, clip } = st.handle;
    const apex = oc.holdAt != null ? clip.duration * oc.holdAt : clip.duration;
    action.paused = false;
    action.enabled = true;
    if (!st.open) {
      action.timeScale = 1;
      if (action.time >= apex || action.time === 0) action.time = 0;
      st.stopAt = apex;
    } else {
      action.timeScale = -1;
      if (action.time <= 0) action.time = apex;
      st.stopAt = 0;
    }
    action.play();
    const watch = () => {
      if ((action.timeScale > 0 && action.time >= st.stopAt) || (action.timeScale < 0 && action.time <= st.stopAt)) {
        action.time = st.stopAt;
        action.paused = true;
        stage.updaters.delete(watch);
      }
    };
    stage.updaters.add(watch);
    st.open = oc.toggle ? !st.open : st.open;
  }

  function showTextCard(e, title, body) {
    card = mkCard(e, `<span class="x">×</span><b></b><div class="dim"></div>`);
    card.querySelector('b').textContent = title || '';
    card.querySelector('.dim').textContent = body || '';
    card.querySelector('.x').onclick = closeCard;
    placeCard(card, e);
  }

  function showInfo(e, pl, inst) {
    closeCard();
    card = mkCard(e, `<span class="x">×</span><b></b><div class="dim"></div><div class="mono"></div><a>Open object →</a>`);
    card.querySelector('b').textContent = pl.name || inst.asset.name;
    card.querySelector('.dim').textContent = pl.info?.body || inst.asset.description || '';
    const bits = [];
    if (pl.pos) bits.push(`pos ${pl.pos.map((v) => Math.round(v)).join(', ')}`);
    for (const [k, v] of Object.entries(pl.props || {})) bits.push(`${k} ${JSON.stringify(v)}`);
    card.querySelector('.mono').textContent = bits.join(' · ');
    card.querySelector('a').onclick = () => ctx.navigate(inst.asset.id);
    card.querySelector('.x').onclick = closeCard;
    placeCard(card, e);
  }

  function mkCard(e, html) {
    const c = document.createElement('div');
    c.className = 'infocard';
    c.innerHTML = html;
    el.appendChild(c);
    return c;
  }

  // placeCard positions an infocard near the click but always fully inside
  // the stage: measure the real card (its height depends on the text), then
  // clamp both axes. Callers fill the card's text first.
  function placeCard(c, e) {
    const r = el.getBoundingClientRect();
    const x = e.clientX - r.left + 10, y = e.clientY - r.top + 10;
    c.style.left = `${Math.max(8, Math.min(x, r.width - c.offsetWidth - 8))}px`;
    c.style.top = `${Math.max(8, Math.min(y, r.height - c.offsetHeight - 8))}px`;
  }

  // ---- cutscene scripts --------------------------------------------------------------
  let player = null;
  const scriptBtns = [];
  if (doc.scripts?.length) {
    const list = document.createElement('div');
    list.className = 'side-list';
    for (const sr of doc.scripts) {
      const btn = document.createElement('button');
      btn.textContent = `▶ ${sr.name || sr.id}`;
      btn.onclick = async () => {
        if (player) { player.dispose(); player = null; btn.textContent = `▶ ${sr.name || sr.id}`; return; }
        const script = await game.doc(docPath, sr.file);
        player = new CutscenePlayer({
          stage, game, el,
          scriptPath: game.root ? resolvePath(docPath, sr.file) : sr.file,
          docPath, scriptRef: sr, script,
          placements: placementById,
          layers: layerNodes,
          onEnd: () => { player = null; btn.textContent = `▶ ${sr.name || sr.id}`; },
        });
        btn.textContent = '■ Stop';
        player.play();
      };
      list.appendChild(btn);
      scriptBtns.push(btn);
    }
    el.appendChild(list);
  }

  function resolvePath(from, ref) {
    const dir = from.includes('/') ? from.slice(0, from.lastIndexOf('/') + 1) : '';
    return dir + ref;
  }

  // ---- AR (diorama) mode ----------------------------------------------------------
  // The level goes into the room as a tabletop model. There is no console on a
  // headset, so the status line is the only channel a failure has.
  const param = (k) => params.get?.(k) ?? params[k];
  const num = (v, d) => { const n = parseFloat(v); return Number.isFinite(n) && n > 0 ? n : d; };

  const xrStatus = document.createElement('div');
  xrStatus.className = 'hud xr-status';
  xrStatus.hidden = true;
  el.appendChild(xrStatus);
  const setStatus = (t) => { xrStatus.hidden = false; xrStatus.textContent = t; };

  // contentBox measures what the diorama should be fitted to. The skybox has
  // to be excluded or the fit is meaningless: SM64DS's vr01.glb is a
  // ~118,000-unit dome around a 16-unit level, and Box3 does not skip
  // invisible objects, so the hidden collision shell has to go too.
  function contentBox() {
    const box = new THREE.Box3();
    for (const r of roots) {
      const ly = r.userData.layer;
      if (ly && (ly.role === 'sky' || ly.role === 'collision' || ly.attach === 'camera' || ly.attach === 'cameraYaw')) continue;
      if (!r.visible) continue;
      box.expandByObject(r);
    }
    return box;
  }

  // The direction the document's own opening shot looks along, flattened: the
  // diorama turns so you are standing where that camera stood. Bob-omb
  // Battlefield's is [0.1,12.9,-17.5] → [0.1,2.9,-0.3], i.e. looking +Z.
  function establishingDir() {
    const d = new THREE.Vector3();
    if (cam.pos && cam.target) d.fromArray(cam.target).sub(new THREE.Vector3().fromArray(cam.pos));
    d.y = 0;
    return d.lengthSq() > 1e-9 ? d.normalize() : new THREE.Vector3(0, 0, 1);
  }

  let arSaved = null;
  function setARScene(on) {
    if (on) {
      // In AR the room is the backdrop; the game's sky belongs to the game's
      // world. (A future VR mode standing inside the level at 1:1 would want
      // it back — hence save/restore rather than a one-way hide.)
      arSaved = new Map();
      for (const { group, def } of layerNodes.values()) {
        if (def.role !== 'sky' && def.attach !== 'camera' && def.attach !== 'cameraYaw') continue;
        arSaved.set(group, group.visible);
        group.visible = false;
      }
      player?.dispose();
      for (const b of scriptBtns) b.disabled = true;
      hud.hidden = true;
    } else {
      for (const [g, v] of arSaved || []) g.visible = v;
      arSaved = null;
      for (const b of scriptBtns) b.disabled = false;
      hud.hidden = false;
    }
  }

  // Ortho stages (the pan2d levels) have no place in a headset.
  const ar = stage.camera.isPerspectiveCamera
    ? new ARSession({
      stage,
      contentBox,
      targetSize: num(param('xrsize'), 1.0),
      distance: num(param('xrdist'), 1.5),
      frontDir: establishingDir(),
      onStatus: setStatus,
      onScene: setARScene,
    })
    : null;
  // ?xrdebug=1 answers "why is there no XR button?" without a console.
  if (param('xrdebug')) {
    arSupported.then((ok) => setStatus(
      ok ? 'AR: supported' : navigator.xr ? 'AR: immersive-ar not supported by this browser' : 'AR: no navigator.xr (needs a secure origin)',
    ));
  }

  window.__rx3 = { stage, placementById, layerNodes, doc, ar, contentBox, get player() { return player; } }; // debug

  return {
    unmount() {
      ar?.exit();
      player?.dispose();
      fly?.dispose();
      panInput?.dispose();
      stage.dispose();
      closeCard();
      hud.remove();
      xrStatus.remove();
      el.querySelector('.side-list')?.remove();
    },
    xr: ar,
    sources: () => [stage.canvas],
    setWireframe(on) { for (const r of roots) applyWireframe(r, on); },
    setVariant(id) { activeVariant = id; applyVariant(); },
    setNative(size) { stage.setNative(size); },
    pixelGrid: () => (stage.native
      ? { cell: el.clientHeight / stage.native.h, ox: 0, oy: 0, ref: el.clientWidth }
      : null),
    stats: () => {
      let tris = 0, meshes = 0;
      stage.scene.traverse((o) => {
        if (o.isMesh) { meshes++; const g = o.geometry; tris += (g.index ? g.index.count : g.attributes.position?.count || 0) / 3; }
      });
      return {
        Triangles: Math.round(tris).toLocaleString(),
        Meshes: meshes,
        Layers: (doc.scene?.layers || []).map((l) => l.id).join(', ') || '—',
        Rooms: doc.scene?.rooms ? `${doc.scene.rooms.list.length} (${roomsLoaded} loaded)` : '—',
        Placements: (doc.placements || []).length,
        Variants: variants.map((v) => v.id).join(', ') || '—',
        Camera: `${cam.mode} · ${stage.camera.isOrthographicCamera ? 'orthographic' : 'perspective'}${cam.fly ? ` · ${cam.fly.speed} u/s` : ''}`,
      };
    },
  };
}

function applyLayerLook(root, ly) {
  root.traverse((o) => {
    if (ly.renderOrder) o.renderOrder = ly.renderOrder;
    const m = o.material;
    if (!m) return;
    for (const mat of Array.isArray(m) ? m : [m]) {
      if (ly.transparent) mat.transparent = true;
      if (ly.depthTest === false) { mat.depthTest = false; mat.depthWrite = false; }
      if (ly.polygonOffset) {
        mat.polygonOffset = true;
        mat.polygonOffsetFactor = ly.polygonOffset;
        mat.polygonOffsetUnits = ly.polygonOffset;
      }
    }
  });
}

function wireBehavior(stage, node, pl, routeById) {
  const b = pl.behavior;
  if (pl.route && routeById.has(pl.route.id)) {
    const r = routeById.get(pl.route.id);
    const pts = r.points.map((p) => new THREE.Vector3(p[0], p[1] ?? 0, p[2] ?? 0));
    const lens = [0];
    for (let i = 1; i < pts.length; i++) lens.push(lens[i - 1] + pts[i].distanceTo(pts[i - 1]));
    const closed = r.loop && (pl.route.mode || 'loop') === 'loop';
    if (closed) lens.push(lens[lens.length - 1] + pts[0].distanceTo(pts[pts.length - 1]));
    const total = lens[lens.length - 1];
    let d = 0, dir = 1;
    stage.updaters.add((dt) => {
      d += (pl.route.speed || 1) * dt * dir;
      if (closed) d = ((d % total) + total) % total;
      else if (d > total) { d = total; dir = -1; }
      else if (d < 0) { d = 0; dir = 1; }
      // find segment
      let i = 1;
      while (i < lens.length - 1 && lens[i] < d) i++;
      const a = pts[(i - 1) % pts.length], bb = pts[i % pts.length];
      const t = (d - lens[i - 1]) / Math.max(1e-6, lens[i] - lens[i - 1]);
      const prev = node.position.clone();
      node.position.lerpVectors(a, bb, t);
      if (pl.route.face) {
        const dv = node.position.clone().sub(prev);
        if (dv.lengthSq() > 1e-8) node.rotation.y = Math.atan2(dv.x, dv.z);
      }
    });
  }
  if (!b) return;
  if (b.kind === 'spin') {
    const axis = new THREE.Vector3(...(b.axis || [0, 1, 0])).normalize();
    stage.updaters.add((dt) => node.rotateOnAxis(axis, (b.rate || 1) * dt));
  } else if (b.kind === 'flyer' && b.keys?.length) {
    const spinNode = b.spinPart ? node.getObjectByName(b.spinPart.node) : null;
    let seg = 0, t = 0;
    stage.updaters.add((dt) => {
      const tickHz = 60;
      t += dt * tickHz;
      let k = b.keys[seg % b.keys.length];
      let hold = k.hold || 0, dur = Math.max(1, k.dur || 1);
      while (t > dur + hold) { t -= dur + hold; seg++; k = b.keys[seg % b.keys.length]; hold = k.hold || 0; dur = Math.max(1, k.dur || 1); }
      const next = b.keys[(seg + 1) % b.keys.length];
      const f = Math.min(1, t / dur);
      node.position.set(
        k.pos[0] + (next.pos[0] - k.pos[0]) * f,
        k.pos[1] + (next.pos[1] - k.pos[1]) * f,
        k.pos[2] + (next.pos[2] - k.pos[2]) * f,
      );
      if (k.yaw != null && next.yaw != null) node.rotation.y = k.yaw + (next.yaw - k.yaw) * f;
      if (spinNode && b.spinPart.rate) spinNode.rotateOnAxis(new THREE.Vector3(...b.spinPart.axis).normalize(), b.spinPart.rate * dt);
    });
  }
}
