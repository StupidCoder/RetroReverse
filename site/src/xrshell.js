// xrshell.js — one AR session that outlives what is in it.
//
// The problem this exists to solve: app.js teardown() destroys the mounted
// view's Stage on every navigation, and the XRSession lives on that renderer's
// renderer.xr — so changing level used to mean leaving the headset. The shell
// owns a Stage of its OWN, built once and kept, and mounts the ordinary view
// modules against it by handing them ctx.stage3d. The session never notices.
//
// It is a generalisation of what xrmap.js already did for one content type: a
// hidden stage of its own, with _buildContent/_dropContent against a live
// session. This does the same for every type, by reusing the views themselves
// rather than reimplementing what they build.
//
// Three things are less obvious than they look:
//
//   SCALE. ARSession fits content by scaling stage.scene, and the menu lives in
//   that same scene — so the panel would be scaled by the fit too, differently
//   for every asset. The UI group carries the inverse (see _layoutUI).
//
//   MEMORY. A stage that outlives its content has no renderer.dispose() to fall
//   back on. Every swap must hand the GPU its memory back explicitly, which is
//   what Stage.resetContent + ObjectLibrary.dispose are for. The status line
//   reports resident geometries and textures so a leak is visible in the
//   headset, where there is no console to ask.
//
//   ENTRY. Entering is just "switch to the current asset" with an empty room to
//   start from, so entry and every later switch are one code path — and the
//   session is requested inside the pinch, before any loading, which is what
//   keeps WebXR's user-activation requirement satisfied.

import { THREE, Stage } from './engine3d.js';
import { ARSession, arSupported, PerfMeter } from './xr.js';
import { Panel, Pointer } from './xrui.js';
import { Browser } from './xrbrowser.js';
import { Grabber } from './xrgrab.js';

// What a session can swap to without ending. Everything else is listed but not
// selectable — see the note on `disabled` in xrbrowser.js.
const SHOWABLE = new Set(['level', 'object']);
const UP = new THREE.Vector3(0, 1, 0);
const ONE = new THREE.Vector3(1, 1, 1);

export class XRShell {
  // { games, mountView(ctx) -> view, params, onSwitch(gameId, assetId), onChange(active) }
  constructor(opts) {
    this.opts = opts;
    this.games = opts.games || [];
    this.supported = arSupported;
    this.onChange = null;
    this.fake = false;          // ?xrfake=1 — the whole menu, on a desktop
    this._built = false;
    this._busy = false;
    this._msg = '';
    this._perf = '';
    this._seq = 0;              // NOT undefined: ++undefined is NaN, and NaN
                                // never equals itself, so every load would look
                                // superseded and unmount itself on arrival
    this._current = null;       // { gameId, assetId, view, abort }
  }

  get active() { return !!this.ar?.active || !!this._fakeOn; }

  // ---- lifecycle ---------------------------------------------------------------

  async enter(gameId, assetId, { fake = false } = {}) {
    this.fake = fake;
    if (!this._built) await this._build();
    this.host.style.display = '';
    // The AR path reports through ARSession.onChange; the fake one has no
    // session to report for.
    if (this.fake) { this._fakeOn = true; this._fakeAttach(); this._open(true); this.onChange?.(true); } else {
      // Re-anchor every session, and stay INVISIBLE until there is a head pose
      // to anchor to. The menu's resting place before that is the reference
      // space's origin — which is a spot on the floor, and a panel served up
      // inside the carpet is worse than no panel.
      this._uiRef = null;
      this.ui.visible = false;
      this._open(true);
      await this.ar.enter();
    }
    // Enter first, load second: the room comes up with the menu in it and the
    // asset arrives behind it. See the header note on user activation.
    await this.switchTo(gameId, assetId);
  }

  exit() {
    if (this.fake) {
      this._fakeOn = false;
      this._teardownFake();
      this._drop();
      this.host.style.display = 'none';
      this.onChange?.(false);
      return;
    }
    this.ar?.exit();
  }

  async _build() {
    const host = document.createElement('div');
    // NOT under #stage: app.js teardown() does stage.innerHTML = '', which would
    // detach the presenting renderer's canvas out from under the session.
    host.className = 'xr-shell-host';
    host.style.cssText = this.fake
      ? 'position:fixed;inset:0;z-index:5'
      : 'position:fixed;left:0;top:0;width:0;height:0;overflow:hidden;pointer-events:none';
    document.body.appendChild(host);
    this.host = host;

    // Always perspective: the fit and the head pose assume it, and it is why a
    // pan2d level can reach AR at all (its desktop camera is orthographic).
    const stage = new Stage(host, { fov: 50, near: 0.02, far: 200 });
    this.stage = stage;

    this._buildUI();

    this.meter = new PerfMeter();
    this._step = (dt) => {
      if (!this.fake && !stage.renderer.xr.isPresenting) return;
      // A level streams its rooms in over seconds, each arriving in the scene
      // the moment it lands. Until a fit has been solved they would all be
      // drawn at raw world scale — a mansion the size of a street — so the
      // curtain is held down every frame, not just once after the mount.
      if (!this._shown) this._setContentVisible(false);
      // 2-D content has a clock of its own that nothing else will turn: the
      // pixi ticker is stopped for the session's duration, because window rAF
      // is not reliably serviced while presenting.
      try { this._current?.view?.arContent?.update?.(dt); } catch (e) { console.error('xr content update', e); }
      const note = this.meter.sample(dt, stage);
      if (note) { this._perf = note; this._status(); }
    };
    stage.updaters.add(this._step);
    // The shell's own updaters survive resetContent; a view's do not.
    this._ownUpdaters = new Set(stage.updaters);

    if (!this.fake) {
      this.ar = new ARSession({
        stage,
        contentBox: () => this._contentBox(),
        ready: () => !!this._current?.view,
        onStatus: (t) => { this._msg = t; this._status(); },
        onFrame: (frame, ref) => this._frame(frame, ref),
        onSelect: (src) => this._select(src),
        onSelectStart: (src) => this._selectStart(src),
        onSelectEnd: (src) => this.grab.end(src),
        onPlaced: () => this._onPlaced(),
      });
      this.ar.onChange = (on) => {
        // Leaving drops the content: the desktop view is about to rebuild the
        // same asset, and the shell holding on to a whole mansion behind it is
        // a copy nobody can see.
        if (!on) { this._teardownFake(); this._drop(); this.pointer.hide(); }
        this.onChange?.(on);
      };
    }
    // Moving and resizing the content with your hands. Built after the session,
    // because it drives it — and it never touches the rig itself, it hands
    // ARSession the same four numbers the automatic fit does.
    this.grab = new Grabber({
      ar: this.ar,
      // Only the content is grabbable: not the menu (the pointer owns that),
      // and not the camera or the lights, which have nothing to hit but would
      // be walked on every cast.
      pickables: () => this.stage.scene.children.filter(
        (o) => o !== this.ui && !o.isCamera && !o.isLight,
      ),
      onChange: () => this._status(),
    });

    this._built = true;
    window.__xr = this; // debug handle, like __rx3 / __rxo
  }

  _buildUI() {
    // One group for everything that is menu rather than content, so the fit's
    // scale can be undone in exactly one place.
    this.ui = new THREE.Group();
    this.ui.name = 'xr-ui';

    this.panel = new Panel({ status: true });
    this.tab = new Panel({ chrome: 'bar' });
    this.pointer = new Pointer();
    this.panel.group.visible = false;
    this.ui.add(this.panel.group, this.tab.group, this.pointer.group);

    this.browser = new Browser({
      games: this.games,
      canShow: (a) => SHOWABLE.has(a.category),
      onPick: (g, a) => { this.switchTo(g, a); },
      onAction: (id) => {
        if (id === 'close') this._open(false);
        // Re-place, rather than exit: "recentre" is the one thing you reach for
        // when the model has ended up behind you.
        if (id === 'recenter') this.ar?.rePlace();
      },
    });
    this.tab.setModel({ chrome: 'bar', footer: [{ id: 'open', label: '☰ Browse' }] });
    this._redraw();
  }

  _redraw() {
    const lay = this.panel.setModel(this.browser.model());
    // Paging to the loaded asset needs the rows-per-page that only a layout
    // knows, so it is resolved after one and applied on a second.
    if (this.browser.applyPending(lay.cols)) this.panel.setModel(this.browser.model());
  }

  _open(on) {
    this.panel.group.visible = on;
    this.tab.group.visible = !on;
    if (on) this._redraw();
  }

  _status() {
    const bits = [this._msg, this._ptr, this._perf].filter(Boolean);
    this.panel?.setStatus(bits.join(' · '));
  }

  // ---- content -------------------------------------------------------------------

  // switchTo is the whole point of the shell: drop what is here, mount what was
  // asked for against the SAME stage, refit. The session is not touched.
  async switchTo(gameId, assetId) {
    const game = this.games.find((g) => g.id === gameId);
    const asset = game?.asset(assetId);
    if (!game || !asset) { this._say(`no such asset: ${gameId}/${assetId}`); return; }

    this._busy = true;
    this.browser.busy = true;
    // List the game being loaded straight away. Waiting for the mount would
    // leave the asset column blank for as long as the level takes, which reads
    // as a broken menu rather than a loading one.
    if (this.browser.gameId !== gameId) {
      this.browser.gameId = gameId;
      this.browser.expanded.clear();
      this.browser.page.assets = 0;
    }
    this.browser.expanded.add(`cat:${asset.category}`);
    if (asset.group) this.browser.expanded.add(`grp:${asset.category}|${asset.group}`);
    this._say(`loading ${asset.name || asset.id}…`);
    this._redraw();

    this._drop();
    this._shown = false; // curtain down until the fit lands (see _step)

    // The shell's own controller, never ctx.signal: app.js aborts that on every
    // navigation, and a navigation is exactly what a switch looks like from
    // outside.
    const abort = new AbortController();
    const seq = ++this._seq;
    let view = null;
    try {
      view = await this.opts.mountView(this._ctx(game, asset, abort.signal));
    } catch (e) {
      // The panel is the console: there is nowhere else for this to go.
      this._busy = false;
      this.browser.busy = false;
      this._say(`failed to load ${asset.id}: ${e.message || e}`);
      this._redraw();
      return;
    }
    if (seq !== this._seq) { view?.unmount?.(); return; } // superseded mid-load

    // A 3-D view has already built itself into this stage — it was handed
    // ctx.stage3d. A 2-D one has not: it is a PixiJS view, and its three.js
    // form (a tilemap in one draw call, or its canvas on a quad) is made here,
    // on demand, against the stage the session is presenting from.
    const built = view.arContent;
    if (built?.build) {
      try {
        await built.build(this.stage);
      } catch (e) {
        this._busy = false;
        this.browser.busy = false;
        this._say(`cannot show ${asset.id} in AR: ${e.message || e}`);
        this._redraw();
        return;
      }
      if (seq !== this._seq) { view.unmount?.(); return; }
    }

    this._current = { gameId, assetId, view, abort };
    this._busy = false;
    this.browser.busy = false;
    this.browser.setLoaded(gameId, assetId);
    this.browser.setControls(this._controls(view));

    const c = view.arContent || {};
    if (this.fake) this._fakePlace();
    else this.ar.setContent({ ...c, ready: () => true });
    c.setPresenting?.(true);

    this.opts.onSwitch?.(gameId, assetId);
    this._say(`${game.man.title} — ${asset.name || asset.id}`);
    this._redraw();
  }

  // The ctx contract the view modules expect (app.js showAsset builds the same
  // shape). displayPanel and setVariants are recorded rather than rendered:
  // that is how the Controls tab gets its contents without any view knowing
  // that XR exists.
  _ctx(game, asset, signal) {
    this._rec = [];
    const rec = this._rec;
    let section = null;
    const push = (o) => { if (section) { rec.push({ kind: 'section', label: section }); section = null; } rec.push(o); };
    // visibility:hidden rather than display:none — a laid-out box still has a
    // clientWidth for pixi to size against, and WebGL renders into a hidden
    // canvas just the same, which is what the panel quad samples.
    const el = document.createElement('div');
    el.style.cssText = 'position:absolute;left:0;top:0;width:960px;height:720px;visibility:hidden';
    this.host.appendChild(el);
    this._viewEl = el;
    return {
      stage: el,
      stage3d: this.stage,
      game,
      asset,
      params: (typeof this.opts.params === 'function' ? this.opts.params() : this.opts.params) || new URLSearchParams(),
      signal,
      navigate: (assetId, params) => this.switchTo(game.id, assetId, params),
      toast: (t) => this._say(t),
      setVariants: (variants, activeId, apply) => {
        if (!variants?.length || variants.length < 2) return;
        section = 'Variant';
        for (const v of variants) {
          push({
            kind: 'radio', label: v.name || v.id, on: v.id === activeId,
            set: () => { apply(v.id); for (const r of rec) if (r.kind === 'radio' && r.group === 'variant') r.on = false; },
            group: 'variant',
          });
        }
      },
      displayPanel: {
        section: (t) => { section = t; },
        toggle: (label, on, cb) => {
          const o = { kind: 'toggle', label, on: !!on, set: (v) => { o.on = v; cb(v); } };
          push(o);
        },
        radio: (name, label, on, cb) => {
          const o = {
            kind: 'radio',
            label,
            on: !!on,
            group: name,
            set: () => { for (const r of rec) if (r.group === name) r.on = (r === o); cb(); },
          };
          push(o);
        },
      },
    };
  }

  // Everything the mounted view offers, in the order it makes sense to read it.
  _controls(view) {
    const out = [...(this._rec || [])];
    const clips = view.clips?.() || [];
    if (clips.length) {
      out.push({ kind: 'section', label: 'Animation' });
      for (const c of clips) {
        out.push({ kind: 'radio', label: c.name, on: c.on, group: 'clip', set: () => { view.playClip(c.id); this.browser.setControls(this._controls(view)); } });
      }
    }
    const look = [];
    if (view.setWireframe) {
      const o = { kind: 'toggle', label: 'Wireframe', on: false, set: (v) => { o.on = v; view.setWireframe(v); } };
      look.push(o);
    }
    if (view.setSunlight) {
      const o = { kind: 'toggle', label: 'Sunlight', on: true, set: (v) => { o.on = v; view.setSunlight(v); } };
      look.push(o);
    }
    if (look.length) { out.push({ kind: 'section', label: 'Look' }); out.push(...look); }
    return out;
  }

  // _drop returns the previous content's memory. Nothing else will: a stage
  // that outlives its content has no renderer teardown to hide behind.
  _drop() {
    const cur = this._current;
    this._current = null;
    // Whatever was being held is about to stop existing.
    this.grab?.cancel();
    if (cur) {
      cur.abort.abort();
      cur.view?.arContent?.setPresenting?.(false);
      try { cur.view?.arContent?.dispose?.(); } catch (e) { console.error('xr content dispose', e); }
      try { cur.view?.unmount?.(); } catch (e) { console.error('xr unmount', e); }
    }
    // Each mount gets its own element and it goes with the mount. A 2-D view
    // needs a real size to lay out against (pixi's resizeTo reads clientWidth),
    // which the host itself does not have — it is a 0x0 clip in a session.
    this._viewEl?.remove();
    this._viewEl = null;
    // Detach the menu FIRST. resetContent disposes everything the old scene can
    // reach, and the panel is in that scene — dropping a level would otherwise
    // take the menu's own textures with it.
    this.ui.removeFromParent();
    this.stage.resetContent(this._ownUpdaters);
    // The session's own wiring, re-asserted rather than assumed: resetContent
    // is about content, and what it does or does not clear must not be able to
    // silently cost us the frame hook.
    this.ar?.reattach();
    this.stage.scene.add(this.ui);
    this.stage.scene.scale.set(1, 1, 1);
    this._shown = true;
  }

  // The content box is the view's own — measured the way that view knows how
  // (posed vertices for a skinned model, sky and collision layers excluded for
  // a level). The menu must never be in it, or the fit would shrink the scene
  // to make room for its own panel.
  //
  // Measured with the content VISIBLE even when it is being kept hidden until
  // the fit lands: both contentBox implementations skip invisible objects
  // (level3d.js:546, viewobject.js:558) because that is how they exclude the
  // collision shell — so measuring it hidden would report an empty box and the
  // session would say "nothing to place".
  _contentBox() {
    const hidden = !this._shown;
    if (hidden) this._setContentVisible(true);
    try {
      const c = this._current?.view?.arContent;
      if (c?.contentBox) return c.contentBox();
      const box = new THREE.Box3();
      for (const o of this.stage.scene.children) if (o !== this.ui) box.expandByObject(o);
      return box;
    } finally {
      if (hidden) this._setContentVisible(false);
    }
  }

  _setContentVisible(on) {
    for (const o of this.stage.scene.children) if (o !== this.ui) o.visible = on;
  }

  _contentVisible(on) {
    this._shown = on;
    this._setContentVisible(on);
  }

  _onPlaced() {
    this._contentVisible(true);
    this._layoutUI();
  }

  // ---- the menu in the room --------------------------------------------------------

  // Put the menu where it was anchored. Two corrections, every frame:
  //
  // The anchor is held in REFERENCE space, not world space, so it stays put in
  // the ROOM. The rig is what maps one to the other, and re-placing the content
  // (the Recentre button, a pointed floor spot) moves the rig — so a world-space
  // anchor would drag the menu along with the diorama instead of leaving it
  // where you put it.
  //
  // And ARSession fits by scaling stage.scene, which the menu is in, so the
  // group carries the inverse or the panel would resize with every asset. The
  // scene's transform is a pure uniform scale — the rig carries position and
  // yaw, and it is not the scene's parent — so world = k * local exactly.
  _layoutUI() {
    if (this.fake) return; // the menu rides the camera there, not the scene
    const rig = this.ar?.rig;
    if (!this._uiRef || !rig) return;
    const k = this.stage.scene.scale.x || 1;
    const m = new THREE.Matrix4()
      .compose(this._uiRef.pos, this._uiRef.quat, ONE)
      .premultiply(rig.matrixWorld);
    const pos = new THREE.Vector3(), quat = new THREE.Quaternion(), scl = new THREE.Vector3();
    m.decompose(pos, quat, scl);
    this.ui.position.copy(pos).divideScalar(k);
    this.ui.quaternion.copy(quat);
    this.ui.scale.setScalar(1 / k);
  }

  // Park the menu where the viewer is looking, once, and leave it there: it is
  // furniture in the room, not a heads-up display that swims with the head.
  // headPos/headQuat are in reference space — see _layoutUI.
  _anchorUI(headPos, headQuat) {
    const fwd = new THREE.Vector3(0, 0, -1).applyQuaternion(headQuat);
    fwd.y = 0;
    if (fwd.lengthSq() < 1e-6) fwd.set(0, 0, -1);
    fwd.normalize();
    const pos = headPos.clone().addScaledVector(fwd, 1.15).setY(headPos.y - 0.1);
    // A quad's front is its +Z, and Matrix4.lookAt(eye, target) sets +Z to
    // (eye - target) — so the EYE argument is the head, not the panel. The
    // other way round hangs the panel facing away from you, and because the
    // material is double-sided it does not vanish, it reads mirrored.
    const quat = new THREE.Quaternion().setFromRotationMatrix(
      new THREE.Matrix4().lookAt(headPos, pos, UP),
    );
    this._uiRef = { pos, quat };
    this.ui.visible = true;
    // The tab sits just below the panel, where the panel would be if it were
    // open — so opening and closing does not move the thing you aim at.
    this.tab.group.position.set(0, -this.panel.height / 2 - 0.10, 0);
    this._layoutUI();
  }

  // ---- input ------------------------------------------------------------------------

  _frame(frame, ref) {
    const rig = this.ar?.rig;
    if (!rig) return;
    if (!this._uiRef) {
      // Reference space, straight off the pose: the anchor is a spot in the
      // room, and the rig that maps it into the world moves whenever content
      // is re-placed.
      const pose = frame.getViewerPose(ref);
      if (pose) {
        const p = pose.transform.position, o = pose.transform.orientation;
        this._anchorUI(new THREE.Vector3(p.x, p.y, p.z), new THREE.Quaternion(o.x, o.y, o.z, o.w));
      }
    }
    const rays = this._rays = this._aim(frame, ref, rig);
    // Grabbing first, THEN the menu: a drag moves the rig, and the menu is
    // anchored in the room, so it has to be re-placed against the rig the drag
    // just produced or it swims by a frame.
    this.grab.rebaseWith(rays);
    this.grab.update(rays);
    this._layoutUI();
  }

  // Point at whatever is pointing at the panel. Hands deliver a targetRaySpace
  // just as controllers do — a pinch is a select from a tracked pointer — so
  // none of this needs the hand-tracking feature or a single joint.
  _aim(frame, ref, rig) {
    const sources = this.ar?.session?.inputSources || [];
    const rays = new Map();
    let best = null, posed = 0;
    for (const src of sources) {
      if (!src.targetRaySpace) continue;
      const pose = frame.getPose(src.targetRaySpace, ref);
      if (!pose) continue;
      posed++;
      // Reference space first: that is the space grabbing works in, because it
      // is the space that does not move when the content does.
      const rm = new THREE.Matrix4().fromArray(pose.transform.matrix);
      const rq = new THREE.Quaternion().setFromRotationMatrix(rm);
      // quat as well as dir: a grab holds its anchor in the HAND'S frame, so
      // that an anchor which is not on the ray still travels with the hand
      // instead of being snapped onto it.
      rays.set(src, {
        origin: new THREE.Vector3().setFromMatrixPosition(rm),
        dir: new THREE.Vector3(0, 0, -1).applyQuaternion(rq).normalize(),
        quat: rq.clone(),
      });

      const m = rm.clone().premultiply(rig.matrixWorld);
      const origin = new THREE.Vector3().setFromMatrixPosition(m);
      const dir = new THREE.Vector3(0, 0, -1)
        .applyQuaternion(new THREE.Quaternion().setFromRotationMatrix(m)).normalize();
      const hit = this._cast(origin, dir);
      if (!best || hit) best = { origin, dir, hit, src };
      if (hit) break;
    }
    this._hover = best?.hit || null;
    // Say what the pointer can see. With no console on the headset, "is it even
    // getting a ray?" is otherwise unanswerable — and it was the exact question
    // when the frame hook went missing and there was no laser to be seen.
    const grabbing = this.grab.active ? ` · ${this.grab.mode}` : '';
    const note = `${sources.length} src/${posed} posed${best?.hit ? ' · on panel' : ''}${grabbing}`;
    if (note !== this._ptr) { this._ptr = note; this._status(); }
    if (!best) { this.pointer.hide(); return rays; }
    // Into the menu group's space: the pointer lives under it, and it carries
    // the inverse of the fit scale (see _layoutUI). A world-space beam would be
    // drawn in the wrong place, by a factor that changes with every asset.
    this.ui.updateMatrixWorld(true);
    const end = best.hit ? best.hit.point.clone() : best.origin.clone().addScaledVector(best.dir, 2.5);
    this.pointer.aim(
      this.ui.worldToLocal(best.origin.clone()),
      this.ui.worldToLocal(end),
      // Panel and tab both face +Z with no rotation of their own inside the
      // group, so the surface the dot sits on is the group's own facing.
      best.hit ? new THREE.Quaternion() : null,
    );
    this._applyHover();
    return rays;
  }

  _applyHover() {
    const h = this._hover;
    this.panel.setHover(h?.panel === this.panel ? h.rect?.id : null);
    this.tab.setHover(h?.panel === this.tab ? h.rect?.id : null);
  }

  // _cast returns which panel a ray meets and which of its rects, or null.
  _cast(origin, dir) {
    // No near/far bounds. In a session the panel is about a metre away, but on
    // the desktop rig it is parked thousands of units out so that it clears a
    // near plane sized to the level — any fixed range would be wrong for one of
    // the two, and a raycast against two known quads has no use for one.
    const ray = new THREE.Raycaster(origin, dir);
    // three's raycaster does not consult `visible` — an unanchored menu still
    // sitting at the origin would answer pinches from inside the floor.
    if (!this.ui.visible) return null;
    for (const p of [this.panel, this.tab]) {
      if (!p.group.visible) continue;
      const hits = ray.intersectObject(p.mesh, false);
      if (!hits.length) continue;
      return { panel: p, mesh: p.mesh, point: hits[0].point, rect: p.hitTest(hits[0].uv) };
    }
    return null;
  }

  _select() {
    const h = this._hover;
    if (!h) return false;
    if (h.panel === this.tab) { this._open(true); return true; }
    if (!h.rect) return true;   // a pinch on the panel's own background: eaten, not a click
    if (this.browser.click(h.rect)) this._redraw();
    return true;
  }

  // A pinch is a click if it lands on the menu and a grab otherwise. The menu
  // gets first refusal, so you can never drag the room by reaching for a row.
  _selectStart(src) {
    if (this._hover) return;              // on the menu: the click handles it
    const rays = this._rays;
    this.grab.start(src, rays?.get(src));
  }

  _say(t) { this._msg = t; this._status(); }

  // ---- ?xrfake=1 -------------------------------------------------------------------
  //
  // The same shell, the same panel, the same click routing, on a desktop with
  // the mouse for a hand. Everything except the placement maths, which is
  // ARSession's and unchanged. Without this almost none of the menu could be
  // looked at before it reached a headset that has no console.

  // _fakeAttach puts the menu in front of the camera and wires the mouse to it.
  // Called before the first load as well as after every swap, so the panel is up
  // while content streams — the same order the AR path gets for free.
  _fakeAttach() {
    const stage = this.stage;
    // Head-locked here on purpose: there is no head, and a world-anchored panel
    // on a desktop would just be a panel you have to go and find. three only
    // draws camera children when the camera is itself in the scene.
    //
    // The distance is not fixed, because the near plane is not: frameBox sets
    // it from the content's radius, and a level measured in hundreds of game
    // units pushes near past a panel parked at 1.2. Hold the RATIO instead —
    // apparent size is distance over scale — so the menu reads the same whether
    // it is hanging in front of a door handle or a mansion.
    const d = Math.max(1.2, stage.camera.near * 12);
    this._uiRef = null;
    this.ui.scale.setScalar(d / 1.2);
    this.ui.quaternion.identity();
    this.ui.position.set(0, 0, -d);
    stage.camera.add(this.ui);
    stage.scene.add(stage.camera);
    this.tab.group.position.set(0, -this.panel.height / 2 - 0.10, 0);
    if (!this._fakeHooked) this._hookFake();
  }

  _fakePlace() {
    this.stage.scene.scale.set(1, 1, 1);
    this.stage.frameBox(this._contentBox()); // sets near/far — attach reads them
    this._fakeAttach();
    this._onPlaced();
  }

  _hookFake() {
    this._fakeHooked = true;
    const stage = this.stage;
    const ndc = new THREE.Vector2();
    const cast = (e) => {
      const r = stage.canvas.getBoundingClientRect();
      ndc.set(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1);
      const ray = new THREE.Raycaster();
      ray.setFromCamera(ndc, stage.camera);
      return this._cast(ray.ray.origin.clone(), ray.ray.direction.clone());
    };
    this._fakeMove = (e) => { this._hover = cast(e); this._applyHover(); };
    this._fakeDown = (e) => { this._hover = cast(e); this._applyHover(); this._select(); };
    stage.canvas.addEventListener('pointermove', this._fakeMove);
    stage.canvas.addEventListener('pointerdown', this._fakeDown);
  }

  _teardownFake() {
    if (!this._fakeHooked) return;
    this._fakeHooked = false;
    this.stage.canvas.removeEventListener('pointermove', this._fakeMove);
    this.stage.canvas.removeEventListener('pointerdown', this._fakeDown);
  }

  // ---- teardown ---------------------------------------------------------------------

  dispose() {
    this.exit();
    this._drop();
    this.panel?.dispose();
    this.tab?.dispose();
    this.pointer?.dispose();
    this.stage?.dispose();
    this.host?.remove();
    this._built = false;
  }
}
