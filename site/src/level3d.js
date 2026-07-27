// level3d.js — the 3D scene engine: layers (base/toggle/exclusive, camera
// attachment, paint order), room-graph streaming with area peel toggles,
// placements of any object type, level variants, routes and behaviours,
// onClick interactions, and the cutscene scripts. All of it from the level
// document; no per-game code.

import { THREE, Stage, FlyCam, ObjectLibrary, loadGLB, applyWireframe, applyTexFilter, applyTransform, flyHint } from './engine3d.js';
import { CutscenePlayer } from './cutscene.js';

export async function mount(ctx, doc) {
  const { stage: el, game, asset, params } = ctx;
  const docPath = asset.file;
  const cam = doc.camera || {};

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

  const roots = []; // everything mountable, for wireframe toggle + disposal
  const dp = ctx.displayPanel;
  const hud = document.createElement('div');
  hud.className = 'hud';
  hud.textContent = flyHint;
  el.appendChild(hud);

  // ---- layers -----------------------------------------------------------------
  const layerNodes = new Map(); // id -> {group, def}
  const exclusive = new Map();  // group name -> [layer ids]
  if (doc.scene?.layers?.length) {
    dp.section('Layers');
    await Promise.all(doc.scene.layers.map(async (ly) => {
      const gltf = await loadGLB(game.url(docPath, ly.file));
      if (game.display.texFilter) applyTexFilter(gltf.scene, game.display.texFilter);
      const group = new THREE.Group();
      group.add(gltf.scene);
      group.visible = ly.visible !== false;
      applyLayerLook(gltf.scene, ly);
      stage.scene.add(group);
      roots.push(group);
      layerNodes.set(ly.id, { group, def: ly });
      if (ly.attach === 'camera' || ly.attach === 'cameraYaw') {
        stage.updaters.add(() => group.position.copy(stage.camera.position));
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
    const hudTick = () => hud.textContent = `${roomsLoaded}/${rm.list.length} rooms · ${flyHint}`;
    const worker = async () => {
      for (;;) {
        const r = queue.shift();
        if (!r) return;
        try {
          const gltf = await loadGLB(game.url(docPath, r.file));
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
  if (variants.length) ctx.setVariants(variants, activeVariant);

  // ---- placements ---------------------------------------------------------------------
  const lib = new ObjectLibrary(game);
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
      else if (inst.playAnim && inst.doc.animations?.length) inst.playAnim(inst.doc.animations[0].id);
      if (inst.update) stage.updaters.add((dt, cp, t) => { if (node.visible) inst.update(dt, cp, t); });
      wireBehavior(stage, node, pl, routeById);
    } catch (e) {
      console.warn('placement failed', pl.object, e);
    }
  }));

  const applyVariant = () => {
    for (const { pl, node } of placementById.values()) {
      node.userData.varOn = !pl.variants?.length || pl.variants.includes(activeVariant);
      node.visible = node.userData.varOn;
    }
  };
  if (variants.length) applyVariant();

  // ---- interactions ---------------------------------------------------------------------
  const raycaster = new THREE.Raycaster();
  let card = null;
  const closeCard = () => { card?.remove(); card = null; };

  let down = null, lastTap = 0;
  stage.canvas.addEventListener('pointerdown', (e) => { down = { x: e.clientX, y: e.clientY }; });
  stage.canvas.addEventListener('pointerup', (e) => {
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

  window.__rx3 = { stage, placementById, layerNodes, doc, get player() { return player; } }; // debug

  return {
    unmount() {
      player?.dispose();
      fly?.dispose();
      stage.dispose();
      closeCard();
      hud.remove();
      el.querySelector('.side-list')?.remove();
    },
    sources: () => [stage.canvas],
    setWireframe(on) { for (const r of roots) applyWireframe(r, on); },
    setVariant(id) { activeVariant = id; applyVariant(); },
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
        Camera: `${cam.mode}${cam.fly ? ` · ${cam.fly.speed} u/s` : ''}`,
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
