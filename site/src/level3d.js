// level3d.js — the 3D scene engine: layers (base/toggle/exclusive, camera
// attachment, paint order), room-graph streaming with area peel toggles,
// placements of any object type, level variants, routes and behaviours,
// onClick interactions, and the cutscene scripts. All of it from the level
// document; no per-game code.

import { THREE, Stage, FlyCam, ObjectLibrary, loadGLB, loadImage, applyWireframe, applyTexFilter, applyTransform, flyHint, disposeScene, fitSkyLayer } from './engine3d.js';
import { CutscenePlayer } from './cutscene.js';
import { PanInput } from './pancam.js';
import { arSupported, PerfMeter } from './xr.js';
import { BillboardBatch } from './billboards.js';
import { buildInstances, disposeInstances } from './instances.js';
import { rigFromLights, buildShadowMap, receiveShadows } from './shadowmap.js';

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

  // ctx.stage3d is the XR shell handing over ITS stage: one renderer, kept
  // alive across content swaps so the session survives them. The view then owns
  // the scene but not the stage, and everything that belongs to the stage —
  // camera input, picking, the AR button — is the shell's business instead.
  const ownStage = !ctx.stage3d;
  const stage = ctx.stage3d || new Stage(el, cam);
  // The shell's scene carries no background or fog: the AR session clears both for
  // the session's duration anyway (they would paint over passthrough), and
  // writing them here would leave the FIRST level's fog to be restored on exit.
  if (ownStage && doc.scene?.background) stage.scene.background = new THREE.Color(doc.scene.background);
  if (ownStage && doc.scene?.fog) {
    stage.scene.fog = new THREE.Fog(new THREE.Color(doc.scene.fog.color), doc.scene.fog.near, doc.scene.fog.far);
  }
  // A document that publishes the guest's own light rig has that rig baked into
  // its vertex colours already (the hardware multiplies in gamma space, which a
  // linear renderer does not reproduce), so adding lights on top would light it
  // twice. The rig is used for the shadow pass instead — see shadowmap.js.
  const rig = rigFromLights(doc.scene?.lights);
  if (!rig) {
    stage.scene.add(new THREE.AmbientLight(0xffffff, 2.2));
    const key = new THREE.DirectionalLight(0xffffff, 1.2);
    key.position.set(1, 2, 1.4);
    stage.scene.add(key);
  }

  // Camera input belongs to the stage, and in a session the head owns the
  // camera outright — a FlyCam or PanInput built against the shell's stage
  // would only attach listeners to a canvas nobody is looking at and leave
  // them there for the next content.
  let fly = null;
  if (ownStage && cam.mode === 'fly') fly = new FlyCam(stage, cam.fly?.speed);
  if (ownStage && cam.mode === 'orbit' && cam.orbit) {
    stage.controls.autoRotate = !!cam.orbit.autoRotate;
    stage.controls.autoRotateSpeed = (cam.orbit.autoRotateSpeed || 0.3) * 6;
    if (cam.orbit.minDist) stage.controls.minDistance = cam.orbit.minDist;
    if (cam.orbit.maxDist) stage.controls.maxDistance = cam.orbit.maxDist;
  }
  let panInput = null;
  if (ownStage && cam.mode === 'pan2d') {
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
  let shadowPass = null; // the static shadow map, when the document carries a rig
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
      // A layer with no file is a placement GROUP: no geometry of its own, it
      // exists so the placements naming it share one toggle.
      const gltf = ly.file ? await loadGLB(game.url(docPath, ly.file), signal) : null;
      const group = new THREE.Group();
      if (gltf) {
        if (game.display.texFilter) applyTexFilter(gltf.scene, game.display.texFilter);
        group.add(gltf.scene);
      }
      // A shadow-caster layer is the guest's own depth-shadow proxy: geometry
      // that exists to be rendered from the light and never to be seen.
      group.visible = ly.role !== 'shadow' && ly.visible !== false;
      group.userData.layer = ly; // role/attach/mode — read by contentBox() and AR mode
      if (gltf) {
        applyLayerLook(gltf.scene, ly);
        await applyLayerMaterialExtras(gltf.scene, ly, game, docPath);
        wireBillboardNodes(gltf.scene, stage);
        // Layer geometry never moves, but every auto-updating node recomposes
        // its local matrix each frame — thousands of composes on the bigger
        // courses, on the Quest's one busy thread. Freeze the tree: billboard
        // nodes already drive their own matrix and flag matrixWorldNeedsUpdate,
        // and the camera-attached sky moves via the GROUP node, which stays
        // live — a frozen child still follows a moving parent (the force
        // cascade recomputes world from the frozen local).
        gltf.scene.updateMatrixWorld(true);
        gltf.scene.traverse((o) => { o.matrixAutoUpdate = false; });
      }
      stage.scene.add(group);
      roots.push(group);
      layerNodes.set(ly.id, { group, def: ly });
      if (ly.attach === 'camera' || ly.attach === 'cameraYaw') {
        // A horizon has to be brought inside the far plane before it can be
        // seen at all — see fitSkyLayer, which also takes it out of depth.
        fitSkyLayer(group, stage.camera.far);
        // camPos is the camera's WORLD position, in metres (engine3d.js), while
        // group.position is local to a parent that in a session carries the
        // scene's metres-per-unit scale. Copying one straight into the other
        // puts the sky at 1/k of where it belongs. It never showed because AR
        // hides these layers outright — world mode is the first thing to turn
        // them back on. The billboard batch below already does it this way.
        const p = new THREE.Vector3();
        stage.updaters.add((dt, camPos) => {
          p.copy(camPos);
          group.parent?.worldToLocal(p);
          group.position.copy(p);
        });
      }
    }));
    // The shadow pass, once: the scene and its light are both static, so the
    // map is rendered a single time and every other layer samples it.
    const caster = doc.scene.layers.find((ly) => ly.role === 'shadow');
    if (rig && caster) {
      const rec = layerNodes.get(caster.id);
      const shadow = rec && buildShadowMap(stage.renderer, rec.group, rig, doc.scene.shadow);
      if (shadow) {
        for (const [id, other] of layerNodes) {
          if (id !== caster.id) receiveShadows(other.group, shadow);
        }
        shadowPass = shadow;
      }
    }
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
      // A placed object stands in the same light as the layers around it, so it
      // takes the same shadow. The pass is already built by now — it is
      // rendered from the caster layer before the placements load — and the
      // patch is idempotent, which matters because clones share materials with
      // the prototype they came from.
      receiveShadows(node, shadowPass);
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
      if (inst.update) {
        // Kept on the record: if this placement joins the billboard batch its
        // own per-frame work is what the batch replaces, and an updater still
        // in the set would be a lookAt on a mesh nobody draws.
        const upd = (dt, cp, t) => { if (node.visible) inst.update(dt, cp, t); };
        placementById.get(pl.id).upd = upd;
        stage.updaters.add(upd);
      }
      wireBehavior(stage, node, pl, routeById);
    } catch (e) {
      if (!signal.aborted) console.warn('placement failed', pl.object, e);
    }
  }));

  // ---- billboard batching -------------------------------------------------------------
  // A level's sprite cast is the draw call budget: 571 placements on the
  // Abyss's level 8, each its own mesh, material and texture. Merged into one
  // atlas and one geometry they are a single call, and the per-sprite lookAt
  // updaters go with them. The placement groups stay exactly as they were —
  // they still carry the transform, the visibility and the pick record — so
  // layers, variants, rooms and clicking are untouched by this.
  // ?nobatch=1 keeps the per-sprite meshes, which is what the batch is
  // checked AGAINST: the two paths have to render the same picture.
  let billboards = null;
  {
    const cand = [...placementById.values()].filter((r) => r.inst.doc?.type === 'billboard3d');
    if (cand.length >= 8 && !(params.get?.('nobatch') ?? params.nobatch)) {
      const batch = new BillboardBatch({ maxTextureSize: stage.renderer.capabilities.maxTextureSize });
      for (const r of cand) if (batch.add(r.node, r.inst)) r.batched = true;
      const stats = batch.size ? batch.build() : null;
      if (stats) {
        billboards = batch;
        stage.scene.add(batch.group);
        roots.push(batch.group);
        for (const r of cand) if (r.batched && r.upd) stage.updaters.delete(r.upd);
        // World space, not the camera's local: under an AR rig the camera is
        // a child of the rig and the scene carries the diorama's scale, so
        // the eye has to be expressed where the sprites live.
        const camLocal = new THREE.Vector3();
        stage.updaters.add((dt, camPos, t) => {
          camLocal.copy(camPos);
          batch.group.parent?.worldToLocal(camLocal);
          batch.update(camLocal, t);
        });
      } else {
        // Nothing was taken apart unless build() succeeded, so the per-sprite
        // path is still intact here.
        for (const r of cand) r.batched = false;
        console.warn('billboard batch not built (atlas would exceed the texture limit)');
      }
    }
  }

  // ---- static instancing --------------------------------------------------------------
  // The solid-geometry sibling of the billboard batch above (see instances.js
  // for the whole story): repeated inert scenery placements collapse to one
  // InstancedMesh per (object, primitive). The placement groups stay for
  // picking — invisible, which the raycaster ignores. Same ?nobatch=1 escape,
  // same contract: the two paths must render the same picture.
  let instanced = null;
  if (!(params.get?.('nobatch') ?? params.nobatch)) {
    instanced = buildInstances([...placementById.values()].filter((r) => !r.batched), THREE);
    if (instanced) {
      stage.scene.add(instanced.group);
      const s = instanced.stats;
      console.log(`instanced ${s.taken} placements: ${s.callsBefore} draw calls -> ${s.meshes}`);
    }
  }

  // Static placements never move, but three recomposes every auto-updating
  // matrix in the scene each frame — on a Quest that is hundreds of pointless
  // composes per frame for scenery (and for the batch's invisible pick
  // proxies, which never render at all). Freeze them; anything with an
  // updater (billboards, behaviours) keeps its live matrix.
  for (const r of placementById.values()) {
    if (r.upd || r.inst.update || r.inst.playAnim) continue;
    r.node.updateMatrixWorld(true);
    r.node.traverse((o) => { o.matrixAutoUpdate = false; });
  }

  // hoisted: the variant setter can fire while placements are still loading
  // The variant the picker last chose. World mode reads it (xrworld _tick):
  // both the top-bar select and the XR shell's radio funnel through the same
  // apply callback below, so one place holds the answer and a mode that needs
  // to react to it does not have to be wired into every picker.
  function applyVariant() {
    for (const r of placementById.values()) {
      r.node.userData.varOn = !r.pl.variants?.length || r.pl.variants.includes(activeVariant);
      // A batched record renders through its InstancedMesh; its own node is
      // the invisible pick proxy and must stay that way.
      r.node.visible = r.node.userData.varOn && !r.batched;
    }
  }
  if (variants.length) applyVariant();

  // ---- interactions ---------------------------------------------------------------------
  const raycaster = new THREE.Raycaster();
  let card = null;
  const closeCard = () => { card?.remove(); card = null; };

  let down = null, lastTap = 0;
  // Picking is a stage concern too: these listeners live on a canvas the view
  // does not own, and on a persistent stage they would pile up one pair per
  // content swap, each closing over a scene that has been disposed.
  const onDown = (e) => { down = { x: e.clientX, y: e.clientY }; };
  const onUp = (e) => {
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
  };
  if (ownStage) {
    stage.canvas.addEventListener('pointerdown', onDown);
    stage.canvas.addEventListener('pointerup', onUp);
  }

  function pick(e) {
    const r = stage.canvas.getBoundingClientRect();
    const p = new THREE.Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1);
    raycaster.setFromCamera(p, stage.camera);
    let hit = null, dist = Infinity;
    for (const h of raycaster.intersectObjects(pickables, true)) {
      let o = h.object;
      while (o && !o.userData.pick) o = o.parent;
      if (o?.visible !== false && o?.userData.pick) { hit = o; dist = h.distance; break; }
    }
    // A batched sprite has no mesh of its own left to hit, so the batch
    // answers for it — and the nearer of the two answers wins, or clicking a
    // chest would report the wall behind it.
    const b = billboards?.pick(raycaster);
    return b && b.distance < dist ? b.node : hit;
  }

  // partner=true means this call came from another placement's onClick.with,
  // so it drives its own animation but does not expand its own partners — the
  // two leaves of a door name each other, and without that guard the click
  // would bounce between them.
  function runAction(pl, node, e, partner) {
    const oc = pl.onClick;
    if (oc.action === 'text') {
      showTextCard(e, oc.title || pl.name, oc.body);
      return;
    }
    if (oc.action !== 'animate') return; // unknown actions are no-ops by spec
    // One thing in the world can be several placements. Each partner plays its
    // OWN action, because a pair need not share a clip: a double door's leaves
    // swing on mirrored ones.
    if (!partner) {
      for (const id of oc.with || []) {
        const r = placementById.get(parseInt(id, 10));
        if (r?.pl?.onClick && r.node) runAction(r.pl, r.node, e, true);
      }
    }
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
  let statusBase = '', perfNote = '';
  const showStatus = () => {
    xrStatus.hidden = false;
    xrStatus.textContent = perfNote ? `${statusBase} · ${perfNote}` : statusBase;
  };
  const setStatus = (t) => { statusBase = t; showStatus(); };

  // The frame readout — the same line the 2-D map shows in AR, for the same
  // reason: a diorama that drops frames on the part has to say so where you
  // can read it, and "it feels like 30" is not a measurement. ?perf=1 puts it
  // on the desktop HUD too, so a change can be given a control reading
  // without the headset (a desktop number is not the target, only a control).
  const meter = new PerfMeter();
  const perfHud = !!param('perf');
  stage.updaters.add((dt) => {
    const xr = stage.renderer.xr.isPresenting;
    if (!xr && !perfHud) {
      // A reading must not outlive the session that produced it.
      if (perfNote) { perfNote = ''; meter.reset(); }
      return;
    }
    const note = meter.sample(dt, stage);
    if (!note) return;
    if (xr) { perfNote = note; showStatus(); } else hud.textContent = `${hint} · ${note}`;
  });

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
  // mode: 'diorama' | 'world' | false. Which one matters, so it is named rather
  // than passed as a bag of flags — the view has mode-dependent behaviour beyond
  // the sky, and a boolean had already stopped being the truth.
  function setARScene(mode) {
    if (mode) {
      // In a DIORAMA the room is the backdrop and the game's sky would be a box
      // hanging around your carpet, so it goes. In WORLD mode you are standing
      // inside the game's world and the sky is part of it — which is exactly why
      // this was written as a save/restore and not a one-way hide.
      arSaved = new Map();
      if (mode !== 'world') {
        for (const { group, def } of layerNodes.values()) {
          if (def.role !== 'sky' && def.attach !== 'camera' && def.attach !== 'cameraYaw') continue;
          arSaved.set(group, group.visible);
          group.visible = false;
        }
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

  // Everything a session needs to place this level, in one object. When the XR
  // shell owns the session, and this is everything it needs to place the level.
  //
  // Ortho stages used to be refused a session here, but that was a fact about
  // the DESKTOP camera rather than about the level: the shell's stage is always
  // perspective, so a pan2d level reaches AR through this like any other.
  const arContent = {
    contentBox,
    frontDir: establishingDir(),
    setPresenting: setARScene,
    fit: { targetSize: num(param('xrsize'), 1.0), distance: num(param('xrdist'), 1.5) },
  };

  // ?xrdebug=1 answers "why is there no XR button?" without a console.
  if (param('xrdebug')) {
    arSupported.then((ok) => setStatus(
      ok ? 'AR: supported' : navigator.xr ? 'AR: immersive-ar not supported by this browser' : 'AR: no navigator.xr (needs a secure origin)',
    ));
  }

  window.__rx3 = { stage, placementById, layerNodes, doc, arContent, contentBox, get billboards() { return billboards; }, get player() { return player; } }; // debug

  return {
    // The variant the picker last chose, or null when the level has none. Read
    // rather than pushed: the top-bar select and the XR shell's radio both call
    // the same apply callback, so a mode that must react to a switch reads this
    // instead of being wired into every picker (xrworld _tick).
    get activeVariant() { return variants.length ? activeVariant : null; },
    unmount() {
      billboards?.dispose();
      if (instanced) { disposeInstances(instanced.group); instanced.group.removeFromParent(); }
      player?.dispose();
      fly?.dispose();
      panInput?.dispose();
      if (ownStage) {
        stage.canvas.removeEventListener('pointerdown', onDown);
        stage.canvas.removeEventListener('pointerup', onUp);
        stage.dispose(); // takes the canvas with it, so unhook first
      }
      // Give the GPU its memory back. `roots` is the precise record of what
      // this mount built — layers, streamed rooms, placements, the billboard
      // batch — so it is a better answer than walking the scene, which only
      // sees what is still attached to it. The library on top of that: its
      // protos are not in any scene graph at all.
      //
      // Both matter only on a stage that outlives the view. With its own
      // renderer, Stage.dispose() drops the whole context and none of this is
      // observable — which is exactly why it was never here.
      shadowPass?.dispose();
      for (const r of roots) disposeScene(r);
      lib.dispose();
      closeCard();
      hud.remove();
      xrStatus.remove();
      el.querySelector('.side-list')?.remove();
    },
    arContent,
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
        Sprites: billboards?.stats
          ? `${billboards.stats.sprites} batched · ${billboards.stats.draws} draw call${billboards.stats.draws > 1 ? 's' : ''}`
            + ` · ${billboards.stats.sheets} sheets in a ${billboards.stats.atlas} atlas`
          : '—',
        Variants: variants.map((v) => v.id).join(', ') || '—',
        Camera: `${cam.mode} · ${stage.camera.isOrthographicCamera ? 'orthographic' : 'perspective'}${cam.fly ? ` · ${cam.fly.speed} u/s` : ''}`,
      };
    },
  };
}

// applyLayerMaterialExtras applies the Retro-X material extras a layer's GLB
// carries — additive/alpha blending, submission-order draw layers, and
// env-cube sheen fed by the layer's own cube faces (ly.envMap, six images in
// +x,-x,+y,-y,+z,-z order) — the same semantics ObjectLibrary gives model3d
// objects, which layer GLBs used to lose (OutRun's sea sheen was the tell).
async function applyLayerMaterialExtras(root, ly, game, docPath) {
  let cube = null;
  if (ly.envMap?.length === 6) {
    const imgs = await Promise.all(ly.envMap.map((f) => loadImage(game.url(docPath, f))));
    cube = new THREE.CubeTexture(imgs);
    cube.colorSpace = THREE.SRGBColorSpace;
    cube.needsUpdate = true;
  }
  const mats = new Set();
  root.traverse((o) => {
    for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) {
      mats.add(m);
      if (!ly.renderOrder && Number.isFinite(m.userData?.layer)) o.renderOrder = m.userData.layer;
    }
  });
  for (const m of mats) {
    if (m.userData?.blend === 'additive') {
      m.transparent = true;
      m.blending = THREE.AdditiveBlending;
      m.depthWrite = false;
      m.needsUpdate = true;
    }
    if (m.userData?.blend === 'alpha') {
      m.transparent = true;
      m.depthWrite = false;
      m.needsUpdate = true;
    }
    if (cube && m.userData?.sheen) {
      m.envMap = cube;
      m.combine = THREE.AddOperation;
      m.reflectivity = 0.3;
      m.needsUpdate = true;
    }
  }
}

// wireBillboardNodes yaws nodes marked extras {"billboard":"y"} about their
// local Y each frame to face the camera — OutRun's trackside trees and signs
// ship with only their placement matrix baked (the game composes the yaw at
// enqueue time from the camera vectors; see the course markdown, Part XXIX).
function wireBillboardNodes(root, stage) {
  const bills = [];
  root.traverse((o) => { if (o.userData?.billboard === 'y') bills.push(o); });
  if (!bills.length) return;
  for (const o of bills) {
    o.userData.baked = o.matrix.clone();
    o.matrixAutoUpdate = false;
  }
  const rot = new THREE.Matrix4();
  const v = new THREE.Vector3();
  stage.updaters.add((dt, camPos) => {
    for (const o of bills) {
      v.copy(camPos);
      o.parent.worldToLocal(v);
      const e = o.userData.baked.elements;
      rot.makeRotationY(Math.atan2(v.x - e[12], v.z - e[14]));
      o.matrix.multiplyMatrices(o.userData.baked, rot);
      o.matrixWorldNeedsUpdate = true;
    }
  });
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
