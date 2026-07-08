// Marble Madness course viewer — a thin two-engine shell:
//  • Courses — the playfield tilemaps on the shared 2-D LevelViewer (common format).
//  • Slopes — the 3-D slope-field terrain (GLB + markers sidecar; slopes.js, three.js).
// Courses (manifest.levels, kind "tilemap2d") and slopes (manifest.views, kind
// "marble-slope") are independent browse items; showItem() picks the engine by kind and
// shows only the active canvas.

import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
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
    this.gltf = new GLTFLoader();
    this.name = '';
  }

  // The studio drives Pixi start/stop and the keyboard camera through these.
  get app() { return this.map.app; }
  get cam() { return this.map.cam; }

  // Returns the flat browse list: the tilemap courses followed by the slope views.
  async init() {
    const manifest = await this.map.init();
    this.map.app.canvas.classList.add('mm-pixi');
    const levels = (manifest.levels || []).map((l) => ({ ...l, kind: 'tilemap2d' }));
    const slopes = (manifest.views || []).map((v) => ({ ...v, kind: 'marble-slope' }));
    return levels.concat(slopes);
  }

  async showItem(item) {
    this.name = item.name;
    if (item.kind === 'marble-slope') {
      this.mode = 'slopes';
      const [gltf, markers] = await Promise.all([
        this.gltf.loadAsync(config.base + item.file),
        fetch(config.base + item.markers).then((r) => r.json()),
      ]);
      this.slopes.show(gltf.scene, markers);
      this.map.app.canvas.style.display = 'none';
      if (this.slopes.canvas) this.slopes.canvas.style.display = 'block';
    } else {
      this.mode = 'tilemap';
      if (this.slopes.canvas) this.slopes.canvas.style.display = 'none';
      this.map.app.canvas.style.display = 'block';
      await this.map.loadLevel(item);
    }
    this._setHud();
  }

  setLayer(name, on) {
    // 'markers' drives the slope view's Track pins; 'objects' (the scenery
    // overlay sprites) and the rest go to the 2-D map viewer.
    if (name === 'markers') this.slopes.setMarkers(on);
    else this.map.setLayer(name, on);
  }

  _setHud() {
    if (!this.hud) return;
    this.hud.textContent = this.mode === 'slopes'
      ? `${this.name} · slope field · drag to rotate`
      : `${this.name} · map`;
  }
}
