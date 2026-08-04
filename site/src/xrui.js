// xrui.js — a canvas panel you can point at, for use inside an XR session.
//
// WebXR draws no DOM, so an in-headset menu has to be geometry. This is the
// same trick xr2d.js already uses to hang a PixiJS view in the room — a canvas
// as the texture of a quad — with a hit-rect list on the side so a raycast can
// be turned back into "which row is that".
//
// The layout, hit-testing and drawing live in xrlayout.js, which imports
// nothing: that is the half that can be tested off-headset. What is left here
// is the three.js side, and two decisions that keep it cheap:
//
//   hover   is a separate quad MOVED over the row, not a redraw. Pointing at a
//           list sweeps rows continuously, and re-uploading a 1024x720 texture
//           per row would be the most expensive thing in the frame.
//   status  is its own strip with its own small texture, because it is the one
//           line that changes while a level streams — and with no console on
//           the headset, it has to be able to change often.

import { THREE } from './engine3d.js';
import { M, C, layout, hitAt, draw, drawStatus } from './xrlayout.js';

export { M, layout, hitAt, draw, drawStatus } from './xrlayout.js';

// ---- the quad ----------------------------------------------------------------

function texCanvas(w, h) {
  const cv = document.createElement('canvas');
  cv.width = w;
  cv.height = h;
  const tex = new THREE.CanvasTexture(cv);
  tex.colorSpace = THREE.SRGBColorSpace;
  tex.generateMipmaps = false;
  tex.minFilter = THREE.LinearFilter;
  tex.magFilter = THREE.LinearFilter;
  return { cv, ctx: cv.getContext('2d'), tex };
}

function quad(w, h, map, order) {
  const m = new THREE.Mesh(
    new THREE.PlaneGeometry(w, h),
    // depthTest off: a menu that a wall of the diorama can swallow is a menu
    // you cannot get out of. renderOrder keeps the parts stacked correctly
    // among themselves, since with depth off it is draw order that decides.
    new THREE.MeshBasicMaterial({ map, transparent: true, toneMapped: false, side: THREE.DoubleSide, depthTest: false, depthWrite: false }),
  );
  m.renderOrder = order;
  return m;
}

// Panel is a group: the main canvas, a status strip under it, and a hover
// highlight that MOVES rather than redraws.
export class Panel {
  // { width (m), chrome: 'panel'|'bar', status: bool }
  constructor(opts = {}) {
    this.chrome = opts.chrome || 'panel';
    const bar = this.chrome === 'bar';
    this.px = { w: bar ? M.BAR_W : M.W, h: bar ? M.BAR_H : M.H };
    this.width = opts.width ?? (bar ? 0.22 : 0.92);
    this.height = (this.width * this.px.h) / this.px.w;

    this.group = new THREE.Group();
    this.group.name = bar ? 'xr-tab' : 'xr-panel';

    const main = texCanvas(this.px.w, this.px.h);
    this._main = main;
    this.mesh = quad(this.width, this.height, main.tex, 1000);
    this.group.add(this.mesh);

    if (opts.status && !bar) {
      const st = texCanvas(M.W, M.STATUS_H);
      this._status = st;
      this.statusH = (this.width * M.STATUS_H) / M.W;
      this.statusMesh = quad(this.width, this.statusH, st.tex, 1001);
      this.statusMesh.position.y = -this.height / 2 - this.statusH / 2 - 0.006;
      this.group.add(this.statusMesh);
      this._statusText = null;
    }

    this._hover = quad(1, 1, null, 1002);
    this._hover.material.map = null;
    this._hover.material.color.set(C.accent);
    this._hover.material.opacity = 0.22;
    this._hover.visible = false;
    this.group.add(this._hover);

    this.model = null;
    this.lay = { rects: [], cols: [], bar };
    this._hoverId = null;
  }

  // The model is owned by the caller; setModel redraws unconditionally, so
  // callers should only hand over a NEW model when something actually changed.
  setModel(model) {
    this.model = model;
    this.lay = layout(model, M);
    draw(this._main.ctx, model, this.lay, M);
    this._main.tex.needsUpdate = true;
    this._place(this._hoverId); // the hovered row may have moved or gone
    return this.lay;
  }

  setStatus(text) {
    if (!this._status || text === this._statusText) return;
    this._statusText = text;
    drawStatus(this._status.ctx, text, M);
    this._status.tex.needsUpdate = true;
  }

  // uv comes from a raycast against this.mesh. three's PlaneGeometry has its
  // UV origin at the BOTTOM-left; canvas y runs the other way.
  hitTest(uv) {
    if (!uv) return null;
    return hitAt(this.lay, uv.x * this.px.w, (1 - uv.y) * this.px.h);
  }

  setHover(id) {
    if (id === this._hoverId) return;
    this._hoverId = id;
    this._place(id);
  }

  _place(id) {
    const r = id && this.lay.rects.find((q) => q.id === id);
    if (!r || r.item?.disabled) { this._hover.visible = false; return; }
    this._hover.visible = true;
    this._hover.scale.set((r.w / this.px.w) * this.width, (r.h / this.px.h) * this.height, 1);
    this._hover.position.set(
      ((r.x + r.w / 2) / this.px.w - 0.5) * this.width,
      (0.5 - (r.y + r.h / 2) / this.px.h) * this.height,
      0.002,
    );
  }

  dispose() {
    for (const m of [this.mesh, this.statusMesh, this._hover]) {
      if (!m) continue;
      m.geometry.dispose();
      m.material.dispose();
    }
    this._main.tex.dispose();
    this._status?.tex.dispose();
  }
}

// ---- the pointer -------------------------------------------------------------

// A laser and a dot. Both are unlit and depth-free for the same reason the
// panel is: the ray has to be visible across the room whatever is in the way.
export class Pointer {
  constructor(opts = {}) {
    this.group = new THREE.Group();
    this.group.name = 'xr-pointer';

    const g = new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(), new THREE.Vector3(0, 0, -1)]);
    this.line = new THREE.Line(g, new THREE.LineBasicMaterial({
      color: opts.color ?? 0x6ea8ff, transparent: true, opacity: 0.6, depthTest: false, depthWrite: false,
    }));
    this.line.renderOrder = 1003;
    this.line.frustumCulled = false;
    this.group.add(this.line);

    this.dot = new THREE.Mesh(
      new THREE.CircleGeometry(0.006, 20),
      new THREE.MeshBasicMaterial({ color: 0xffffff, depthTest: false, depthWrite: false, transparent: true }),
    );
    this.dot.renderOrder = 1004;
    this.dot.frustumCulled = false;
    this.group.add(this.dot);

    this.group.visible = false;
  }

  // origin and end are in THIS GROUP'S PARENT space, not world space. The
  // pointer hangs under the menu group, which carries the inverse of the fit
  // scale, so world coordinates fed in here would put the laser somewhere else
  // entirely — and by a factor that changes with every asset.
  //
  // faceQuat orients the dot (the surface it landed on, in the same space);
  // null means the ray hit nothing and only the beam is drawn.
  aim(origin, end, faceQuat) {
    this.group.visible = true;
    const pos = this.line.geometry.attributes.position;
    pos.setXYZ(0, origin.x, origin.y, origin.z);
    pos.setXYZ(1, end.x, end.y, end.z);
    pos.needsUpdate = true;
    this.line.geometry.computeBoundingSphere();
    this.dot.visible = !!faceQuat;
    if (faceQuat) {
      // Lift the dot off the surface along the ray, not along the normal: the
      // panel is double-sided and can be read from behind.
      const back = origin.clone().sub(end).normalize().multiplyScalar(0.004);
      this.dot.position.copy(end).add(back);
      this.dot.quaternion.copy(faceQuat);
    }
  }

  hide() { this.group.visible = false; }

  dispose() {
    this.line.geometry.dispose();
    this.line.material.dispose();
    this.dot.geometry.dispose();
    this.dot.material.dispose();
  }
}
