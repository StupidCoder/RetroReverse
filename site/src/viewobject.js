// viewobject.js — the Object/Character category view. sprite2d objects get a
// pixel stage with an animation list; 3D objects (model3d, billboard3d,
// wireframe3d) get an orbit stage with clip buttons and mesh stats.

export async function mount(ctx) {
  const doc = await ctx.game.assetDoc(ctx.asset);
  if (doc.type === 'sprite2d') return mount2D(ctx, doc);
  return mount3D(ctx, doc);
}

// ---- sprite2d ---------------------------------------------------------------

async function mount2D(ctx, doc) {
  const { stage, game, asset } = ctx;
  const { Application, Container } = await import('pixi.js');
  const { SpriteObject } = await import('./sprite2d.js');

  stage.classList.add('render2d');
  const app = new Application();
  // Reflective displays (the Game Boy's dot-matrix panel) show sprites as
  // dark ink on a PALE panel: give their objects light "paper", or the gb
  // screen filter posterizes the Studio's dark backdrop to the darkest
  // green and swallows the art.
  const paper = game.display.filter === 'gb';
  // preserveDrawingBuffer: see level2d.js — the screen-filter capture
  // must not race pixi's render loop.
  await app.init({ background: paper ? '#dfe6ea' : '#101520', resizeTo: stage, antialias: false, resolution: devicePixelRatio, autoDensity: true, preserveDrawingBuffer: true });
  app.canvas.classList.add('pixi', 'fill');
  stage.appendChild(app.canvas);

  const obj = new SpriteObject(asset, doc, await loadImg(game.url(asset.file, doc.atlas.file)));
  const world = new Container();
  app.stage.addChild(world);
  const inst = obj.makeInstance({});
  world.addChild(inst.node);

  let playing = true;
  const tickHz = game.display.tickHz || 60;
  app.ticker.add(() => {
    if (playing) inst.tick((app.ticker.deltaMS * tickHz) / 1000);
    tp?.sync(inst.pos(), inst.cycleFrames(), playing && !inst.done);
  });

  // ---- camera: auto-fit until the user pans/zooms --------------------------------
  let touched = false;
  const fit = () => {
    const z = Math.max(1, Math.floor(Math.min(
      (app.screen.width * 0.5) / obj.cellW, (app.screen.height * 0.5) / obj.cellH)));
    world.scale.set(z);
    const ax = inst.sprite.pivot.x, ay = inst.sprite.pivot.y;
    world.position.set(
      app.screen.width / 2 - (obj.cellW / 2 - ax) * z,
      app.screen.height / 2 - (obj.cellH / 2 - ay) * z);
  };
  fit();
  const onResize = () => { if (!touched) fit(); };
  addEventListener('resize', onResize);

  const cv = app.canvas;
  const zoomAt = (cx, cy, f) => {
    const r = cv.getBoundingClientRect();
    const px = cx - r.left, py = cy - r.top;
    const z0 = world.scale.x, z = Math.max(0.5, Math.min(64, z0 * f));
    world.position.x = px - (px - world.position.x) * (z / z0);
    world.position.y = py - (py - world.position.y) * (z / z0);
    world.scale.set(z);
  };
  const pts = new Map();
  let pinch = 0;
  cv.addEventListener('pointerdown', (e) => { pts.set(e.pointerId, { x: e.clientX, y: e.clientY }); cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', (e) => {
    const p = pts.get(e.pointerId);
    if (!p) return;
    const dx = e.clientX - p.x, dy = e.clientY - p.y;
    p.x = e.clientX; p.y = e.clientY;
    touched = true;
    if (pts.size === 2) {
      const [a, b] = [...pts.values()];
      const d = Math.hypot(a.x - b.x, a.y - b.y);
      if (pinch) zoomAt((a.x + b.x) / 2, (a.y + b.y) / 2, d / pinch);
      pinch = d;
    } else if (pts.size === 1) {
      world.position.x += dx;
      world.position.y += dy;
    }
  });
  const up = (e) => { pts.delete(e.pointerId); if (pts.size < 2) pinch = 0; };
  cv.addEventListener('pointerup', up);
  cv.addEventListener('pointercancel', up);
  cv.addEventListener('wheel', (e) => {
    e.preventDefault();
    touched = true;
    // gentle: one wheel notch (deltaY 100) = 1.2^0.1 ≈ 1.8%
    zoomAt(e.clientX, e.clientY, Math.pow(1.2, -e.deltaY * 0.001));
  }, { passive: false });

  // ---- animation list + transport --------------------------------------------------
  const list = document.createElement('div');
  list.className = 'side-list';
  const btns = [];
  for (const a of doc.animations || []) {
    const b = document.createElement('button');
    b.textContent = `${a.name || a.id} · ${a.frames}f ${a.loop}`;
    if (a.description) b.title = a.description;
    b.onclick = () => {
      inst.setAnim(a.id);
      inst.reset();
      playing = true;
      btns.forEach((x) => x.classList.remove('on'));
      b.classList.add('on');
    };
    list.appendChild(b);
    btns.push(b);
  }
  btns[0]?.classList.add('on');
  stage.appendChild(list);

  const animated = (doc.animations || []).some((a) => expandedFrames(a) > 1);
  const tp = animated ? makeTransport(stage, {
    onToggle() {
      playing = !playing;
      if (playing && inst.done) inst.reset();
    },
    onSeek(f) {
      playing = false;
      inst.seek(f);
    },
  }) : null;

  const hud = document.createElement('div');
  hud.className = 'hud';
  hud.textContent = `${obj.cellW}×${obj.cellH} px cells · ${doc.animations?.length || 0} animations · drag to pan · wheel/pinch to zoom`;
  stage.appendChild(hud);

  // ---- AR: the sprite hangs in the room ---------------------------------------
  const { PanelAR } = await import('./xr2d.js');
  const xrStatus = document.createElement('div');
  xrStatus.className = 'hud xr-status';
  xrStatus.hidden = true;
  stage.appendChild(xrStatus);
  const ar = new PanelAR({
    el: stage,
    canvas: () => app.canvas,
    params: ctx.params,
    onStatus: (t) => { xrStatus.hidden = false; xrStatus.textContent = t; },
  });

  window.__rxo2 = { app, world, inst, obj, ar }; // debug handle

  return {
    unmount() {
      removeEventListener('resize', onResize);
      ar.dispose();
      app.destroy(true, { children: true });
      list.remove(); hud.remove(); tp?.remove(); xrStatus.remove();
      stage.classList.remove('render2d');
    },
    xr: ar,
    sources: () => [app.canvas],
    pixelGrid: () => ({
      cell: world.scale.x, ox: world.position.x, oy: world.position.y,
      ref: app.screen.width,
    }),
    stats: () => ({
      Type: 'sprite2d',
      Cell: `${obj.cellW} × ${obj.cellH} px`,
      Sheet: `${obj.img.width} × ${obj.img.height} px`,
      Animations: (doc.animations || []).map((a) => `${a.id} (${a.frames}f, ${a.loop})`).join(', '),
      ...(doc.stats || {}),
    }),
  };
}

// ---- animation transport (frame counter + slider + play/pause) --------------------

function expandedFrames(a) {
  if (a.steps?.length) return a.steps.length;
  return a.frames || 1;
}

const PLAY_SVG = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>';
const PAUSE_SVG = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 5h4v14H6zM14 5h4v14h-4z"/></svg>';

// makeTransport builds the cutscene-style bar: play/pause, a scrubber in
// engine frames and a frame counter. The view owns the clock — it calls
// sync(frame, total, playing) every tick and receives onToggle/onSeek.
function makeTransport(el, { onToggle, onSeek }) {
  const bar = document.createElement('div');
  bar.className = 'transport';
  bar.innerHTML = `
    <button class="t-play" title="Play/pause">${PAUSE_SVG}</button>
    <input type="range" min="0" max="1" value="0" step="1" />
    <span class="t-label"></span>`;
  el.appendChild(bar);
  const btn = bar.querySelector('.t-play');
  const seek = bar.querySelector('input');
  const label = bar.querySelector('.t-label');
  btn.onclick = onToggle;
  seek.oninput = () => onSeek(+seek.value);
  let lastPlaying = null;
  return {
    sync(frame, total, playing) {
      const max = Math.max(1, Math.ceil(total) - 1);
      if (+seek.max !== max) seek.max = max;
      if (document.activeElement !== seek) seek.value = Math.min(max, frame | 0);
      label.textContent = `f ${Math.min(max, frame | 0)}/${Math.ceil(total)}`;
      if (playing !== lastPlaying) {
        btn.innerHTML = playing ? PAUSE_SVG : PLAY_SVG;
        lastPlaying = playing;
      }
    },
    remove() { bar.remove(); },
  };
}

function loadImg(url) {
  return new Promise((res, rej) => {
    const i = new Image();
    i.onload = () => res(i);
    i.onerror = () => rej(new Error(url));
    i.src = url;
  });
}

// ---- 3D objects ----------------------------------------------------------------

async function mount3D(ctx, doc) {
  const { stage: el, game, asset, params } = ctx;
  const { THREE, Stage, ObjectLibrary, wireMaterial, disposeScene } = await import('./engine3d.js');
  const { ARSession } = await import('./xr.js');

  // ctx.stage3d is the XR shell's own stage, kept alive across content swaps so
  // the session survives them (see level3d.js for the same handover). With it,
  // the camera, its input and the AR button all belong to the shell.
  const ownStage = !ctx.stage3d;
  const stage = ctx.stage3d || new Stage(el, { fov: 45 });
  let clearIsolationRef = null; // assigned by the part-isolation block below
  // A solid background would be painted over the passthrough; in a session
  // the room IS the backdrop.
  if (ownStage) stage.scene.background = new THREE.Color(0x101520);
  const ambient = new THREE.AmbientLight(0xffffff, 2.2);
  stage.scene.add(ambient);
  const key = new THREE.DirectionalLight(0xffffff, 1.2);
  key.position.set(1, 2, 1.4);
  stage.scene.add(key);
  const stopSpin = () => { stage.controls.autoRotate = false; };
  if (ownStage) {
    stage.controls.autoRotate = true;
    stage.controls.autoRotateSpeed = 1.0;
    el.addEventListener('pointerdown', stopSpin, { once: true });
  }

  const lib = new ObjectLibrary(game, ctx.signal);
  // Model variants (Retro-X `variants`): one glTF scene each in the same GLB.
  const variants = doc.variants || [];
  const variantById = (id) => variants.find((v) => v.id === id) || variants[0];
  let activeVariant = variants.length ? variantById(params?.variant)?.id : null;
  let inst = await lib.instance(asset.id, { scene: activeVariant ? variantById(activeVariant).scene : undefined });
  stage.scene.add(inst.node);
  if (inst.update) stage.updaters.add((dt, cp, t) => inst.update(dt, cp, t));

  // Skinned models can carry far-parked hidden parts, so geometry bounds lie:
  // frame (and follow) the SKELETON — pose it, fit the bone cloud, then track
  // its centroid so root-motion clips stay under the camera.
  const bones = [];
  inst.node.traverse((o) => {
    if (o.isBone) bones.push(o);
    if (o.isSkinnedMesh) o.frustumCulled = false;
  });
  if (bones.length && ownStage) {
    inst.mixer?.update(0);
    inst.node.updateMatrixWorld(true);
    const v = new THREE.Vector3();
    const centroid = () => {
      const c = new THREE.Vector3();
      for (const b of bones) c.add(b.getWorldPosition(v));
      return c.divideScalar(bones.length);
    };
    const c0 = centroid();
    let r = 1;
    for (const b of bones) r = Math.max(r, b.getWorldPosition(v).distanceTo(c0));
    const dist = (r * 2.2) / Math.sin((stage.camera.fov * Math.PI) / 360);
    stage.camera.position.copy(c0).add(new THREE.Vector3(0.55, 0.25, 1).normalize().multiplyScalar(dist));
    stage.camera.near = Math.max(0.01, r / 50);
    stage.camera.far = Math.max(stage.camera.far, dist + r * 10);
    // Without this the default near/far (0.1/100000) stay baked into the
    // projection until the first resize rebuilds it — at character scale
    // that is not enough depth precision and everything shimmers.
    stage.camera.updateProjectionMatrix();
    stage.controls.target.copy(c0);
    stage.controls.update();
    let last = c0.clone();
    stage.updaters.add(() => {
      const c = centroid();
      const d = c.clone().sub(last);
      last = c;
      // In an XR session the head owns the camera; nudging it here would be
      // overwritten anyway, and the drift into controls.target would come
      // back out in the pose restored on exit.
      if (!stage.inputEnabled) return;
      stage.camera.position.add(d);
      stage.controls.target.add(d);
    });
  } else if (ownStage) {
    stage.frame(inst.node);
  } else {
    // Skinned models still need one mixer tick before anything measures them:
    // the AR fit reads POSED vertices, and at t=0 they are still the bind pose.
    inst.mixer?.update(0);
    inst.node.updateMatrixWorld(true);
  }

  let handle = null;
  const list = document.createElement('div');
  list.className = 'side-list';
  const btns = [];
  for (const a of doc.animations || []) {
    const b = document.createElement('button');
    b.textContent = `${a.name || a.id}${a.fps ? ` · ${a.fps} fps` : ''} ${a.loop || ''}`;
    if (a.description) b.title = a.description;
    b.onclick = () => {
      handle = inst.playAnim?.(a.id);
      btns.forEach((x) => x.classList.remove('on'));
      b.classList.add('on');
    };
    list.appendChild(b);
    btns.push(b);
  }
  // transport over the active clip: frames at the clip's declared fps
  const clipFps = () => handle?.meta?.fps || 30;
  const tp = (doc.animations || []).length && inst.mixer ? makeTransport(el, {
    onToggle() {
      if (!handle) return;
      const { action, clip } = handle;
      if (action.paused || !action.isRunning()) {
        if (action.time >= clip.duration - 1e-4) action.time = 0;
        action.paused = false;
        action.play();
      } else {
        action.paused = true;
      }
    },
    onSeek(f) {
      if (!handle) return;
      const { action, clip } = handle;
      action.play();
      action.paused = true;
      action.time = Math.min(f / clipFps(), clip.duration - 1e-4);
      inst.mixer.update(0);
    },
  }) : null;
  if (tp) {
    stage.updaters.add(() => {
      if (!handle) return;
      const { action, clip } = handle;
      tp.sync(action.time * clipFps(), clip.duration * clipFps(),
        !action.paused && action.isRunning());
    });
  }
  if (doc.atlasPicture) {
    const b = document.createElement('button');
    b.textContent = '🖼 Texture sheet';
    b.onclick = () => {
      // open the sheet like a picture: a lightweight modal
      const url = game.url(asset.file, doc.atlasPicture);
      window.open(url, '_blank');
    };
    list.appendChild(b);
  }
  if (btns.length) { btns[0].classList.add('on'); handle = inst.playAnim?.(doc.animations[0].id); }
  el.appendChild(list);

  const hud = document.createElement('div');
  hud.className = 'hud';
  const multiPart = inst.node.children.length > 1 || (inst.node.children[0]?.children.length ?? 0) > 1;
  const baseHud = 'drag to orbit · wheel to zoom' + (multiPart ? ' · double-click a part to isolate' : '');
  hud.textContent = baseHud;
  el.appendChild(hud);

  // ---- part isolation: double-click / double-tap a named node ------------
  // The exported variants carry one node per logical part (door, steering
  // wheel, wheel FL, ...). Picking one keeps it solid and ghosts the rest as
  // wireframe; picking it again, empty space, or Escape restores.
  const ray = new THREE.Raycaster();
  const ghost = new THREE.MeshBasicMaterial({ wireframe: true, color: 0x44608a, transparent: true, opacity: 0.5 });
  let isolated = null;
  const clearIsolation = () => {
    if (!isolated) return;
    inst.node.traverse((o) => {
      if (o.isMesh && o.userData.savedMat !== undefined) {
        o.material = o.userData.savedMat;
        delete o.userData.savedMat;
      }
    });
    isolated = null;
    hud.textContent = baseHud;
  };
  const isolate = (root) => {
    clearIsolation();
    for (const c of inst.node.children) {
      if (c === root) continue;
      c.traverse((o) => {
        if (o.isMesh) {
          o.userData.savedMat = o.material;
          o.material = ghost;
        }
      });
    }
    isolated = root;
    hud.textContent = `${(root.name || 'part').replace(/_/g, ' ')} — double-click again to restore`;
  };
  const pickAt = (cx, cy) => {
    const r = stage.canvas.getBoundingClientRect();
    ray.setFromCamera(new THREE.Vector2(
      ((cx - r.left) / r.width) * 2 - 1,
      -((cy - r.top) / r.height) * 2 + 1), stage.camera);
    const hit = ray.intersectObject(inst.node, true).find((h) => h.object.isMesh);
    if (!hit) { clearIsolation(); return; }
    let n = hit.object;
    while (n.parent && n.parent !== inst.node) n = n.parent;
    if (n === inst.node || isolated === n) clearIsolation();
    else isolate(n);
  };
  // Isolation is driven by pointing at the canvas, which in a session nobody is
  // looking at — and these listeners would otherwise accumulate one set per
  // content swap on a stage that outlives the view.
  const onDbl = (e) => {
    // After a touch double-tap the browser synthesizes a dblclick too; without
    // this guard pickAt ran twice and the second call undid the isolation.
    if (performance.now() < suppressDbl) return;
    pickAt(e.clientX, e.clientY);
  };
  // Touch double-tap, hand-rolled. A tap must be short, still and
  // one-fingered (a drag end or a pinch finger is not a tap); two taps
  // within 600 ms and 64 px pick.
  const taps = new Map();
  let lastTap = -1e9, lastTapAt = [0, 0], suppressDbl = 0;
  const onTapDown = (e) => {
    if (e.pointerType !== 'touch') return;
    if (taps.size) for (const t of taps.values()) t.moved = true; // second finger: a pinch, not taps
    taps.set(e.pointerId, { x: e.clientX, y: e.clientY, t: performance.now(), moved: taps.size > 0 });
  };
  const onTapMove = (e) => {
    const t = taps.get(e.pointerId);
    if (t && Math.hypot(e.clientX - t.x, e.clientY - t.y) > 12) t.moved = true;
  };
  const onTapUp = (e) => {
    const t = taps.get(e.pointerId);
    if (!t) return;
    taps.delete(e.pointerId);
    const now = performance.now();
    if (t.moved || now - t.t > 350) { lastTap = -1e9; return; }
    if (now - lastTap < 600 && Math.hypot(e.clientX - lastTapAt[0], e.clientY - lastTapAt[1]) < 64) {
      pickAt(e.clientX, e.clientY);
      suppressDbl = now + 700;
      lastTap = -1e9;
    } else {
      lastTap = now;
      lastTapAt = [e.clientX, e.clientY];
    }
  };
  const onTapCancel = (e) => taps.delete(e.pointerId);
  const onKey = (e) => { if (e.key === 'Escape') clearIsolation(); };
  const isoEvents = [['dblclick', onDbl], ['pointerdown', onTapDown], ['pointermove', onTapMove],
    ['pointerup', onTapUp], ['pointercancel', onTapCancel]];
  if (ownStage) {
    for (const [n, f] of isoEvents) el.addEventListener(n, f);
    window.addEventListener('keydown', onKey);
  }
  clearIsolationRef = clearIsolation;

  // ---- sun toggle: models that carry normals can opt into real lighting ----
  // The exported materials are unlit (the platform-authentic flat look), so
  // the toggle swaps each mesh's material for a Lambert twin and adds a
  // directional "sun" diagonally from above; off restores the originals.
  let hasNormals = false;
  inst.node.traverse((o) => {
    if (o.isMesh && o.geometry.attributes.normal) hasNormals = true;
  });
  let sun = null;
  let wfOn = false;
  let sunOn = false;
  // Wireframe toggle that respects part isolation: the shared ghost material
  // stays wireframe no matter what, so the toggle only affects the isolated
  // part. Ghosted meshes' saved (real) materials are updated too, so clearing
  // the isolation restores them in the current toggle state. wireMaterial
  // renders the wires in one constant bright colour (dark models are
  // unreadable in their own), stashing each material's texture and tint.
  const applyWf = (on) => {
    inst.node.traverse((o) => {
      if (!o.isMesh) return;
      for (const m of [o.material, o.userData.savedMat].flat().filter(Boolean)) {
        if (m !== ghost) wireMaterial(m, on);
      }
    });
  };
  const setSunlight = (on) => {
    sunOn = on;
    clearIsolationRef?.();
    applyWf(false); // swap on the real materials; re-applied below
    if (on && !sun) {
      sun = new THREE.DirectionalLight(0xffffff, 1.6);
      sun.position.set(1, 2, 1.2);
    }
    inst.node.traverse((o) => {
      if (!o.isMesh || !o.geometry.attributes.normal) return;
      const swap = (m) => {
        if (on) {
          if (!m.userData.lit) {
            const lit = new THREE.MeshLambertMaterial({
              map: m.map || null, color: m.color.clone(),
              transparent: m.transparent, opacity: m.opacity,
              alphaTest: m.alphaTest, side: m.side, name: m.name,
              vertexColors: m.vertexColors, // COLOR_0-tinted models (Jak's logo) keep their tint
              envMap: m.envMap || null, combine: m.combine ?? THREE.MultiplyOperation,
              reflectivity: m.reflectivity ?? 1,
            });
            lit.blending = m.blending;
            lit.depthWrite = m.depthWrite;
            m.userData.lit = lit;
            lit.userData.unlit = m;
          }
          return m.userData.lit;
        }
        return m.userData.unlit || m;
      };
      o.material = Array.isArray(o.material) ? o.material.map(swap) : swap(o.material);
    });
    applyWf(wfOn);
    ambient.intensity = on ? 0.55 : 2.2;
    if (on) stage.scene.add(sun);
    else if (sun) stage.scene.remove(sun);
  };

  // Variant switching swaps the instance in place: the camera and toggle
  // states survive, so LOD levels can be compared from one viewpoint.
  if (variants.length > 1) {
    ctx.setVariants?.(
      variants.map((v) => ({ id: v.id, name: v.name })),
      activeVariant,
      async (id) => {
        const v = variantById(id);
        if (!v || v.id === activeVariant) return;
        clearIsolationRef?.();
        const next = await lib.instance(asset.id, { scene: v.scene });
        stage.scene.remove(inst.node);
        inst = next;
        stage.scene.add(inst.node);
        activeVariant = v.id;
        applyWf(wfOn);
        if (sunOn) setSunlight(true);
      });
  }

  // ---- AR (hold it in your hand) -------------------------------------------------
  // Same session machinery as the levels, fitted to the object's LONGEST axis
  // rather than its footprint: a lamppost should be a metre tall, not a metre
  // wide and ten metres tall.
  const xrStatus = document.createElement('div');
  xrStatus.className = 'hud xr-status';
  xrStatus.hidden = true;
  el.appendChild(xrStatus);
  const xrNum = (k, d) => {
    const n = parseFloat(params?.get?.(k) ?? params?.[k]);
    return Number.isFinite(n) && n > 0 ? n : d;
  };

  // contentBox measures what you will actually be handed. Neither shortcut
  // works for a skinned model: its geometry bounds are the BIND pose and
  // include far-parked hidden parts (the bob-omb measures 13.7 that way
  // against a real 6), while the bone cloud is only the skeleton and
  // understates the skin around it (5.9 — the model would arrive at 2.3 m
  // when asked for 1). So pose the vertices and measure those.
  const contentBox = () => {
    const box = new THREE.Box3();
    const v = new THREE.Vector3();
    let posed = false;
    inst.node.updateMatrixWorld(true);
    inst.node.traverse((o) => {
      if (!o.isSkinnedMesh || !o.visible) return;
      const pos = o.geometry.attributes.position;
      if (!pos) return;
      posed = true;
      for (let i = 0; i < pos.count; i++) {
        v.fromBufferAttribute(pos, i);
        o.applyBoneTransform(i, v);      // bind pose -> current pose, mesh-local
        box.expandByPoint(o.localToWorld(v));
      }
    });
    // Unskinned meshes (and whole unskinned models) measure straight.
    inst.node.traverse((o) => {
      if (o.isMesh && !o.isSkinnedMesh && o.visible) box.expandByObject(o);
    });
    if (box.isEmpty() && !posed) box.setFromObject(inst.node); // `inst` is reassigned by the variant switcher
    return box;
  };

  // The AR fitting policy for an object, in one object so the XR shell can take
  // it and drive its own long-lived session instead of this per-view one.
  const arContent = {
    contentBox,
    fit: {
      fitScale: (size) => Math.min(
        xrNum('xrsize', 1.0) / (size.x || 1),
        xrNum('xrsize', 1.0) / (size.z || 1),
        xrNum('xrheight', 1.8) / (size.y || 1),
      ),
      anchor: (params?.get?.('xranchor') ?? params?.xranchor) === 'eye' ? 'eye' : 'floor',
      distance: xrNum('xrdist', 1.2),
    },
    // Without an orbit camera to read, present the model's front.
    frontDir: () => new THREE.Vector3(0, 0, 1),
    setPresenting: () => {},
  };

  let arSpin = false;
  const ar = ownStage && stage.camera.isPerspectiveCamera
    ? new ARSession({
      stage,
      contentBox,
      // A play area is far taller than it is wide, so fit the FOOTPRINT to a
      // metre square and allow up to 1.8 m of height rather than squeezing the
      // longest axis — a tall model should use the headroom, not shrink to fit
      // it into the same box a wide one gets.
      fitScale: (size) => Math.min(
        xrNum('xrsize', 1.0) / (size.x || 1),
        xrNum('xrsize', 1.0) / (size.z || 1),
        xrNum('xrheight', 1.8) / (size.y || 1),
      ),
      // Standing on the floor rather than hanging at eye level: a model reads
      // as an object in the room that way, and WebXR's local-floor reference
      // space already puts y = 0 on the real floor, so it costs nothing.
      // ?xranchor=eye hangs it at eye height instead.
      anchor: (params?.get?.('xranchor') ?? params?.xranchor) === 'eye' ? 'eye' : 'floor',
      // Far enough back to look down at something on the floor; arm's length
      // would put it under your chin.
      distance: xrNum('xrdist', 1.2),
      // Present whichever side you had orbited to.
      frontDir: () => {
        const d = stage.controls.target.clone().sub(stage.camera.position);
        d.y = 0;
        return d.lengthSq() > 1e-9 ? d.normalize() : new THREE.Vector3(0, 0, 1);
      },
      onStatus: (t) => { xrStatus.hidden = false; xrStatus.textContent = t; },
      onScene: (on) => {
        // Autorotate would spin the model where it hangs (the session skips
        // controls.update, but leaving the flag set would resume the spin the
        // moment you exit, from a pose you never chose).
        if (on) { arSpin = stage.controls.autoRotate; stage.controls.autoRotate = false; }
        else stage.controls.autoRotate = arSpin;
        hud.hidden = on;
      },
    })
    : null;

  window.__rxo = { stage, inst, bones, ar, contentBox }; // debug handle

  return {
    unmount() {
      ar?.exit();
      if (ownStage) {
        window.removeEventListener('keydown', onKey);
        for (const [n, f] of isoEvents) el.removeEventListener(n, f);
        el.removeEventListener('pointerdown', stopSpin);
        stage.dispose();
      }
      // Give the GPU its memory back — see the same block in level3d.js. Only
      // observable on a stage that outlives the view; with its own renderer,
      // Stage.dispose() drops the whole context instead.
      disposeScene(inst.node);
      lib.dispose();
      list.remove(); hud.remove(); tp?.remove(); xrStatus.remove();
    },
    xr: ar,
    arContent,
    // The clip list is built as DOM buttons rather than through displayPanel,
    // so the in-headset Controls tab needs its own way in.
    clips: () => (doc.animations || []).map((a) => ({
      id: a.id, name: a.name || a.id, on: handle?.meta?.id === a.id,
    })),
    playClip(id) {
      handle = inst.playAnim?.(id);
      btns.forEach((x, i) => x.classList.toggle('on', (doc.animations || [])[i]?.id === id));
      return handle;
    },
    sources: () => [stage.canvas],
    setNative(size) { stage.setNative(size); },
    pixelGrid: () => (stage.native
      ? { cell: el.clientHeight / stage.native.h, ox: 0, oy: 0, ref: el.clientWidth }
      : null),
    setWireframe(on) { wfOn = on; applyWf(on); },
    ...(hasNormals ? { setSunlight } : {}),
    stats: () => {
      // Wireframe models report their blueprint's own terms — vertices,
      // edges, faces — not mesh-renderer notions like triangles or bones.
      if (doc.type === 'wireframe3d' && doc.wireframe) {
        const wf = doc.wireframe;
        return {
          Type: doc.type,
          Vertices: (wf.positions.length / 3).toLocaleString(),
          Edges: wf.edges.length.toLocaleString(),
          Faces: wf.faces.length.toLocaleString(),
          ...(doc.stats || {}),
        };
      }
      let tris = 0, verts = 0, mats = new Set(), bones = 0;
      const seenPos = new Set(); // primitives of one variant share a POSITION accessor
      inst.node.traverse((o) => {
        if (o.isMesh) {
          const g = o.geometry;
          const pos = g.attributes.position;
          if (pos && !seenPos.has(pos)) { seenPos.add(pos); verts += pos.count; }
          tris += (g.index ? g.index.count : pos?.count || 0) / 3;
          for (const m of Array.isArray(o.material) ? o.material : [o.material]) mats.add(m);
        }
        if (o.isBone) bones++;
      });
      return {
        Type: doc.type,
        Vertices: verts.toLocaleString(),
        Triangles: Math.round(tris).toLocaleString(),
        Materials: mats.size,
        Bones: bones || undefined,
        Clips: (doc.animations || []).map((a) => a.id).join(', ') || (inst.clips?.length ? inst.clips.map((c) => c.name).join(', ') : '—'),
        ...(doc.stats || {}),
      };
    },
  };
}
