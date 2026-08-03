// xr2d.js — AR for the 2-D views.
//
// A tilemap or a sprite sheet has no scene to walk around, so the session
// hangs the VIEW ITSELF in the room: the PixiJS canvas becomes the texture of
// an upright quad floating in front of you. Everything the 2-D view does —
// panning, the animation runner, layer toggles, palette cycling — keeps
// running and shows up on the panel, because the panel is that canvas.
//
// The three.js renderer this needs is built on the first entry, not at mount:
// a 2-D view already holds a WebGL context for Pixi, browsers only allow so
// many, and most visits never press the button.

import { ARSession, arSupported } from './xr.js';

export class PanelAR {
  // { el, canvas: () => HTMLCanvasElement, params, onStatus }
  constructor(opts) {
    this.opts = opts;
    this.supported = arSupported;
    this.onChange = null;
    this._ar = null;
    this._teardown = null;
  }

  get active() { return !!this._ar?.active; }

  async enter() {
    if (!this._ar) await this._build();
    await this._ar.enter();
  }

  exit() { this._ar?.exit(); }

  dispose() {
    this._ar?.exit();
    this._teardown?.();
    this._teardown = null;
    this._ar = null;
  }

  async _build() {
    const { el, canvas, params, onStatus } = this.opts;
    const { THREE, Stage } = await import('./engine3d.js');

    // An immersive session renders into its own framebuffer, so the canvas
    // backing the WebGL context never has to be on screen — and must not be,
    // or it would cover the 2-D view it is mirroring.
    const mount = document.createElement('div');
    mount.style.cssText = 'position:absolute;left:0;top:0;width:0;height:0;overflow:hidden;pointer-events:none';
    el.appendChild(mount);
    const stage = new Stage(mount, { fov: 50, near: 0.01, far: 100 });

    const src = canvas();
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
    stage.scene.add(panel);

    // Track the canvas's aspect — the browser window can be resized, and a
    // stale aspect would stretch the map.
    const fit = () => {
      const c = canvas();
      if (!c?.width || !c.height) return;
      if (tex.image !== c) tex.image = c;
      panel.scale.set(c.width / c.height, 1, 1);
    };
    fit();
    stage.updaters.add(() => { fit(); tex.needsUpdate = true; });

    const num = (k, d) => {
      const n = parseFloat(params?.get?.(k) ?? params?.[k]);
      return Number.isFinite(n) && n > 0 ? n : d;
    };

    this._ar = new ARSession({
      stage,
      contentBox: () => new THREE.Box3().setFromObject(panel),
      fitAxis: 'longest',
      targetSize: num('xrsize', 1.5), // a big screen rather than a tabletop
      distance: num('xrdist', 1.8),
      // The quad's face is its +Z, so the viewer belongs on the +Z side
      // looking back along −Z.
      frontDir: new THREE.Vector3(0, 0, -1),
      onStatus,
    });
    // The wrapper owns the button's callback, so proxy it through.
    this._ar.onChange = (on) => this.onChange?.(on);

    this._teardown = () => {
      stage.dispose();
      panel.geometry.dispose();
      panel.material.dispose();
      tex.dispose();
      mount.remove();
    };
  }
}
