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
  await app.init({ background: '#101520', resizeTo: stage, antialias: false, resolution: devicePixelRatio, autoDensity: true });
  app.canvas.classList.add('pixi', 'fill');
  stage.appendChild(app.canvas);

  const obj = new SpriteObject(asset, doc, await loadImg(game.url(asset.file, doc.atlas.file)));
  const world = new Container();
  app.stage.addChild(world);
  const inst = obj.makeInstance({});
  world.addChild(inst.node);

  const tickHz = game.display.tickHz || 60;
  app.ticker.add(() => inst.tick((app.ticker.deltaMS * tickHz) / 1000));

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
  addEventListener('resize', fit);

  // animation list
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
      btns.forEach((x) => x.classList.remove('on'));
      b.classList.add('on');
    };
    list.appendChild(b);
    btns.push(b);
  }
  btns[0]?.classList.add('on');
  stage.appendChild(list);

  const hud = document.createElement('div');
  hud.className = 'hud';
  hud.textContent = `${obj.cellW}×${obj.cellH} px cells · ${doc.animations?.length || 0} animations`;
  stage.appendChild(hud);

  return {
    unmount() {
      removeEventListener('resize', fit);
      app.destroy(true, { children: true });
      list.remove(); hud.remove();
      stage.classList.remove('render2d');
    },
    sources: () => [app.canvas],
    stats: () => ({
      Type: 'sprite2d',
      Cell: `${obj.cellW} × ${obj.cellH} px`,
      Sheet: `${obj.img.width} × ${obj.img.height} px`,
      Animations: (doc.animations || []).map((a) => `${a.id} (${a.frames}f, ${a.loop})`).join(', '),
      ...(doc.stats || {}),
    }),
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
  const { stage: el, game, asset } = ctx;
  const { THREE, Stage, ObjectLibrary, applyWireframe } = await import('./engine3d.js');

  const stage = new Stage(el, { fov: 45 });
  stage.scene.background = new THREE.Color(0x101520);
  stage.scene.add(new THREE.AmbientLight(0xffffff, 2.2));
  const key = new THREE.DirectionalLight(0xffffff, 1.2);
  key.position.set(1, 2, 1.4);
  stage.scene.add(key);
  stage.controls.autoRotate = true;
  stage.controls.autoRotateSpeed = 1.0;
  el.addEventListener('pointerdown', () => { stage.controls.autoRotate = false; }, { once: true });

  const lib = new ObjectLibrary(game);
  const inst = await lib.instance(asset.id);
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
  if (bones.length) {
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
    stage.controls.target.copy(c0);
    stage.controls.update();
    let last = c0.clone();
    stage.updaters.add(() => {
      const c = centroid();
      const d = c.clone().sub(last);
      stage.camera.position.add(d);
      stage.controls.target.add(d);
      last = c;
    });
  } else {
    stage.frame(inst.node);
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
  hud.textContent = 'drag to orbit · wheel to zoom';
  el.appendChild(hud);

  window.__rxo = { stage, inst, bones }; // debug handle

  return {
    unmount() {
      stage.dispose();
      list.remove(); hud.remove();
    },
    sources: () => [stage.canvas],
    setWireframe(on) { applyWireframe(inst.node, on); },
    stats: () => {
      let tris = 0, verts = 0, mats = new Set(), bones = 0;
      inst.node.traverse((o) => {
        if (o.isMesh) {
          const g = o.geometry;
          verts += g.attributes.position?.count || 0;
          tris += (g.index ? g.index.count : g.attributes.position?.count || 0) / 3;
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
