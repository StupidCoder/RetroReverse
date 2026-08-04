// xr2d.js — a 2-D view as three.js content.
//
// A sprite sheet has no scene to walk around, so the session hangs the VIEW
// ITSELF in the room: the PixiJS canvas becomes the texture of an upright quad
// floating in front of you. Everything the 2-D view does — the animation
// runner, the transport, panning — shows up on the panel, because the panel IS
// that canvas.
//
// Like xrmap.js, this used to own an AR session; the XR shell owns that now, so
// what is left is the quad. That is what lets a sprite panel and a 3-D model
// swap places without the session noticing.

export class PanelContent {
  // { canvas: () => HTMLCanvasElement, tick: (dt) => void, params }
  constructor(opts) {
    this.opts = opts;
    this._stage = null;
    this._parts = null;
    this.fit = {
      fitAxis: 'longest',
      targetSize: this._num('xrsize', 1.5), // a big screen rather than a tabletop
      distance: this._num('xrdist', 1.8),
      dropBelowEye: 0,
    };
    // The quad's face is its +Z, so the viewer belongs on the +Z side looking
    // back along −Z. Plain object: no three.js import here, and ARSession only
    // reads .x and .z.
    this.frontDir = { x: 0, y: 0, z: -1 };
    this.note = '';
  }

  _num(k, d) {
    const p = this.opts.params;
    const n = parseFloat(p?.get?.(k) ?? p?.[k]);
    return Number.isFinite(n) && n > 0 ? n : d;
  }

  async build(stage) {
    this._stage = stage;
    const { THREE } = await import('./engine3d.js');
    this.THREE = THREE;

    const src = this.opts.canvas();
    const tex = new THREE.CanvasTexture(src);
    tex.colorSpace = THREE.SRGBColorSpace;
    // No mipmaps: the source is repainted every frame, and generating a chain
    // per frame for a viewport-sized canvas is pure cost.
    tex.generateMipmaps = false;
    tex.minFilter = THREE.LinearFilter;
    tex.magFilter = THREE.LinearFilter;

    const panel = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      // Unlit and double-sided: it is a screen, not a surface in the room's
      // light, and walking behind it should not make it vanish.
      new THREE.MeshBasicMaterial({ map: tex, toneMapped: false, side: THREE.DoubleSide }),
    );
    const group = new THREE.Group();
    group.name = 'xr-panel-content';
    group.add(panel);
    stage.scene.add(group);
    this._parts = { group, panel, tex };
    this._fitAspect();
    this.note = `${src.width}×${src.height} canvas`;
    return this;
  }

  // Track the canvas's aspect — the window can be resized, and a stale aspect
  // would stretch the map.
  _fitAspect() {
    const c = this.opts.canvas();
    const { panel, tex } = this._parts;
    if (!c?.width || !c.height) return;
    if (tex.image !== c) tex.image = c;
    panel.scale.set(c.width / c.height, 1, 1);
  }

  // The pixi ticker is not reliably serviced while presenting, so the session
  // drives the view's clock AND its render: without a repaint the texture
  // uploaded below would be the same picture every frame.
  advance(dt) {
    if (!this._parts) return;
    this.opts.tick?.(dt);
    this._fitAspect();
    this._parts.tex.needsUpdate = true;
  }

  contentBox() {
    if (!this.THREE || !this._parts) return null;
    return new this.THREE.Box3().setFromObject(this._parts.group);
  }

  dispose() {
    if (this._parts) {
      const { group, panel, tex } = this._parts;
      group.removeFromParent();
      panel.geometry.dispose();
      panel.material.dispose();
      tex.dispose();
      this._parts = null;
    }
    this._stage = null;
  }
}
