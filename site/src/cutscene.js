// cutscene.js — the Retro-X cutscene player (RETROX.md §7): a shot sequencer
// on the script's frame clock. Each frame it poses the camera from the shot's
// baked track (pos/target/roll/fov), shows exactly the current shot's actors
// with their clips scrubbed to the shot time, and drives the declared layers.
// A transport bar offers play/pause, a scrubber and shot chapters.

import { THREE } from './engine3d.js';

export class CutscenePlayer {
  // { stage, game, el, docPath, script, placements: Map(id -> {pl, inst, node}),
  //   layers: Map(id -> {group}), onEnd }
  constructor(opts) {
    Object.assign(this, opts);
    this.fps = this.script.fps || 30;
    this.shots = this.script.shots || [];
    this.starts = [];
    let acc = 0;
    for (const s of this.shots) { this.starts.push(acc); acc += s.frames; }
    this.total = acc;
    this.t = 0;           // frames
    this.playing = false;
    this.tracks = new Map();
    this.scriptDir = this.docPathOfScript();

    // remember camera + visibility to restore on dispose
    this.saved = {
      camPos: this.stage.camera.position.clone(),
      target: this.stage.controls.target.clone(),
      fov: this.stage.camera.fov,
      near: this.stage.camera.near,
      far: this.stage.camera.far,
      up: this.stage.camera.up.clone(),
      vis: new Map([...this.placements.values()].map((p) => [p.node, p.node.visible])),
      layerVis: new Map([...this.layers.values()].map((l) => [l.group, l.group.visible])),
    };

    this._buildTransport();
    this.stage.overrideRender = (dt) => this._frame(dt);
    this._loadTracks();
  }

  docPathOfScript() {
    // The script document lives next to the level doc: refs inside it resolve
    // from ITS directory (level doc dir + script ref dir).
    const ref = this.scriptRef.file;
    const dir = this.docPath.includes('/') ? this.docPath.slice(0, this.docPath.lastIndexOf('/') + 1) : '';
    const p = dir + ref;
    return p.includes('/') ? p.slice(0, p.lastIndexOf('/') + 1) : '';
  }

  async _loadTracks() {
    await Promise.all(this.shots.map(async (s) => {
      if (s.camera?.track) { this.tracks.set(s.id, { track: s.camera.track, frames: s.frames }); return; }
      if (!s.camera?.trackFile) return;
      const doc = await this.game.doc(this.scriptDir, s.camera.trackFile);
      this.tracks.set(s.id, doc);
    }));
    this.ready = true;
  }

  play() { this.playing = true; }
  pause() { this.playing = false; }

  _frame(dt) {
    if (this.playing && this.ready) {
      this.t += dt * this.fps;
      if (this.t >= this.total) { this.t = this.total - 0.001; this.playing = false; }
    }
    const si = this._shotAt(this.t);
    const shot = this.shots[si];
    const local = this.t - this.starts[si];

    // camera from the baked track
    const tr = this.tracks.get(shot.id);
    if (tr?.track?.length) {
      const i = Math.max(0, Math.min(tr.track.length - 1, Math.floor(local)));
      const s = tr.track[i];
      const cam = this.stage.camera;
      cam.position.fromArray(s.pos);
      // roll: rotate the up vector about the view axis. Track roll is in
      // DEGREES (like fov — RETROX.md §7): Luigi's Mansion's door-knob shot
      // straightens from -12.7° to -0.7°, which read as radians was two full
      // revolutions of spin.
      const dir = new THREE.Vector3().fromArray(s.target).sub(cam.position).normalize();
      cam.up.set(0, 1, 0);
      if (s.roll) cam.up.applyAxisAngle(dir, s.roll * Math.PI / 180);
      cam.lookAt(s.target[0], s.target[1], s.target[2]);
      if (s.fov) cam.fov = s.fov;
      cam.near = shot.camera.near || tr.near || cam.near;
      cam.far = shot.camera.far || tr.far || cam.far;
      cam.updateProjectionMatrix();
    }

    // layers: when a shot declares layers, only those show (sets-as-layers);
    // when it declares none, its sets are actors and layers are untouched.
    if (shot.layers?.length) {
      for (const [id, rec] of this.layers) rec.group.visible = shot.layers.includes(id);
    }

    // actors: exactly the current shot's actors are visible, scrubbed
    const acting = new Set();
    for (const a of shot.actors || []) {
      const p = this.placements.get(a.placement);
      if (!p) continue;
      acting.add(p.node);
      p.node.visible = true;
      if (a.matrix?.length === 16) {
        p.node.matrixAutoUpdate = false;
        p.node.matrix.fromArray(a.matrix);
      }
      if (a.mirror) p.node.scale.x = -Math.abs(p.node.scale.x || 1);
      if (a.clip && p.inst.mixer) {
        if (p._csClip !== a.clip) {
          p._csHandle = p.inst.playAnim(a.clip, { loop: 'hold' });
          p._csClip = a.clip;
        }
        if (p._csHandle) {
          const time = Math.max(0, (local - (a.start || 0)) / this.fps);
          p._csHandle.action.paused = true;
          p._csHandle.action.time = Math.min(time, p._csHandle.clip.duration - 1e-4);
          p.inst.mixer.update(0);
        }
      }
    }
    // every script actor not in this shot hides; non-actor placements stay
    const allActorIds = new Set();
    for (const s of this.shots) for (const a of s.actors || []) allActorIds.add(a.placement);
    for (const id of allActorIds) {
      const p = this.placements.get(id);
      if (p && !acting.has(p.node)) p.node.visible = false;
    }

    this._syncUI(si, local);
    this.stage.renderer.render(this.stage.scene, this.stage.camera);
  }

  _shotAt(t) {
    for (let i = this.shots.length - 1; i >= 0; i--) if (t >= this.starts[i]) return i;
    return 0;
  }

  // ---- transport UI -------------------------------------------------------------

  _buildTransport() {
    const bar = document.createElement('div');
    bar.className = 'transport';
    bar.innerHTML = `
      <button class="t-play" title="Play/pause"></button>
      <input type="range" min="0" max="${Math.max(1, this.total - 1)}" value="0" step="0.5" />
      <span class="t-label"></span>
      <select class="t-shots"></select>`;
    this.el.appendChild(bar);
    this.bar = bar;
    this.playBtn = bar.querySelector('.t-play');
    this.seek = bar.querySelector('input');
    this.label = bar.querySelector('.t-label');
    const shotsSel = bar.querySelector('.t-shots');
    shotsSel.className = 'tb-select t-shots';
    this.shots.forEach((s, i) => {
      const o = document.createElement('option');
      o.value = i;
      o.textContent = s.name || s.id;
      shotsSel.appendChild(o);
    });
    shotsSel.onchange = () => { this.t = this.starts[+shotsSel.value]; };
    this.shotsSel = shotsSel;
    this.playBtn.onclick = () => { this.playing = !this.playing; if (this.t >= this.total - 1) this.t = 0; };
    this.seek.oninput = () => { this.t = +this.seek.value; };
    this._setPlayIcon(true);
  }

  _setPlayIcon(play) {
    this.playBtn.innerHTML = play
      ? '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>'
      : '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 5h4v14H6zM14 5h4v14h-4z"/></svg>';
  }

  _syncUI(si, local) {
    if (document.activeElement !== this.seek) this.seek.value = this.t;
    const s = this.shots[si];
    this.label.textContent = `${s.id} · f${Math.floor(local)}/${s.frames}`;
    if (+this.shotsSel.value !== si) this.shotsSel.value = si;
    this._setPlayIcon(!this.playing);
  }

  dispose() {
    this.stage.overrideRender = null;
    this.bar.remove();
    const cam = this.stage.camera;
    cam.position.copy(this.saved.camPos);
    cam.up.copy(this.saved.up);
    cam.fov = this.saved.fov;
    cam.near = this.saved.near;
    cam.far = this.saved.far;
    cam.updateProjectionMatrix();
    this.stage.controls.target.copy(this.saved.target);
    this.stage.controls.update();
    for (const [node, vis] of this.saved.vis) node.visible = vis;
    for (const [group, vis] of this.saved.layerVis) group.visible = vis;
    this.onEnd?.();
  }
}
