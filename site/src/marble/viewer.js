// Marble Madness course viewer — a thin two-engine shell:
//  • Map — the playfield tilemap on the shared 2-D LevelViewer (common format).
//  • Slopes — the 3-D slope-field height mesh (slopes.js, three.js).
// A toggle switches engines; only the active canvas is shown and handles input.

import { LevelViewer } from '../shared/viewer.js';
import { SlopeView } from './slopes.js';
import config from './config.js';

export class MarbleViewer {
  constructor(viewportEl, hudEl) {
    this.el = viewportEl;
    this.hud = hudEl;
    this.mode = 'tilemap';
    this.map = new LevelViewer(viewportEl, null, {
      ...config,
      camEnabled: () => this.mode === 'tilemap',
    });
    this.slopes = new SlopeView(viewportEl, () => this.active !== false && this.mode === 'slopes');
    this.slope = null;   // the current course's slope JSON (lazy-meshed)
  }

  // The studio drives Pixi start/stop and the keyboard camera through these.
  get app() { return this.map.app; }
  get cam() { return this.map.cam; }

  async init() {
    const meta = await this.map.init();
    this.map.app.canvas.classList.add('mm-pixi');
    return meta;
  }

  async loadLevel(metaLevel) {
    this.name = metaLevel.name;
    await this.map.loadLevel(metaLevel);
    this.slope = await fetch(config.base + metaLevel.slope).then((r) => r.json());
    if (this.mode === 'slopes') this.slopes.show(this.slope);
    this._setHud();
  }

  setMode(mode) {
    this.mode = mode;
    if (mode === 'slopes') {
      this.slopes.show(this.slope);
      this.map.app.canvas.style.display = 'none';
      this.slopes.canvas.style.display = 'block';
    } else {
      if (this.slopes.canvas) this.slopes.canvas.style.display = 'none';
      this.map.app.canvas.style.display = 'block';
    }
    this._setHud();
  }

  setLayer(name, on) {
    // 'markers' (the studio toggle) and 'objects' (the generic id) both drive the
    // slope view's Track markers; everything else goes to the map viewer.
    if (name === 'markers' || name === 'objects') this.slopes.setMarkers(on);
    else this.map.setLayer(name, on);
  }

  _setHud() {
    if (!this.hud) return;
    this.hud.textContent = this.mode === 'slopes'
      ? `${this.name} · slope field · drag to rotate`
      : `${this.name} · map`;
  }
}
