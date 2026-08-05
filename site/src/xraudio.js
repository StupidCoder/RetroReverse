// xraudio.js — sound for world mode.
//
// Only the background bed so far: the level's own BGM, looping while you stand
// in it. The positional ambience the preset's `sounds` block describes will land
// here too, which is why this is Web Audio rather than an `<audio>` element —
// the rest of the site plays music through HTMLAudioElement (viewmedia.js) and
// that is right for a player with a scrub bar, but wrong here for two reasons:
//
//   LOOPING. An AudioBufferSourceNode with loop = true repeats the decoded
//   buffer sample-exactly. A media element re-seeks, and several browsers put an
//   audible seam in it — which you would hear once a minute, forever.
//
//   PANNING. A drip in a corridor has to come from somewhere, and PannerNodes
//   need the graph anyway. An AudioListener follows the camera's world matrix,
//   which the rig already moves, so positional sound in a session is the
//   ordinary Web Audio thing and needs nothing from WebXR.
//
// THE ONE AWKWARD PART IS STARTING. An AudioContext begins suspended unless it
// was created inside a user gesture, and entering XR is a gesture whose
// activation is long spent by the time a level has streamed and a placement has
// solved. So a context that comes up suspended is not an error here: it is
// resumed on the first `select`, which in a headset is the first time you pinch
// anything. Nothing is lost but the seconds before you touch a control.

export class WorldAudio {
  // { session, onStatus }
  constructor(opts = {}) {
    this.session = opts.session || null;
    this.onStatus = opts.onStatus || null;
    this.ctx = null;
    this.music = null;
    this._disposed = false;
  }

  _context() {
    if (!this.ctx) {
      const AC = window.AudioContext || window.webkitAudioContext;
      if (!AC) return null;
      this.ctx = new AC();
      if (this.ctx.state === 'suspended') this._resumeOnSelect();
    }
    return this.ctx;
  }

  // A pinch is the gesture; asking again on every one costs nothing and covers
  // the case where the first was consumed by something else.
  _resumeOnSelect() {
    if (!this.session || this._hooked) return;
    this._hooked = () => {
      this.ctx?.resume().then(() => this.onStatus?.()).catch(() => {});
    };
    this.session.addEventListener('select', this._hooked);
  }

  get note() {
    if (!this.music) return '';
    if (this.ctx?.state !== 'running') return 'music: waiting for a pinch';
    return `music: ${this.music.name}`;
  }

  // playMusic loads and loops one track. `url` is fetched once; calling again
  // with the same url is a no-op, so a re-place does not restart the level's
  // theme underneath you.
  async playMusic(url, { gain = 1, name = '' } = {}) {
    if (this.music?.url === url) return;
    this.stopMusic();
    const ctx = this._context();
    if (!ctx) { this.onStatus?.(); return; }
    this.music = { url, name, node: null, gain: null };
    let buf;
    try {
      const res = await fetch(url);
      if (!res.ok) throw new Error(`${res.status}`);
      buf = await ctx.decodeAudioData(await res.arrayBuffer());
    } catch (e) {
      // The status strip is the only console in a headset.
      this.onStatus?.(`music ${name || url}: ${e.message || e}`);
      this.music = null;
      return;
    }
    if (this._disposed || this.music?.url !== url) return;

    const g = ctx.createGain();
    g.gain.value = gain;
    g.connect(ctx.destination);
    const src = ctx.createBufferSource();
    src.buffer = buf;
    src.loop = true;                 // sample-exact, for as long as you stand there
    src.connect(g);
    src.start();
    this.music.node = src;
    this.music.gain = g;
    if (ctx.state === 'suspended') this._resumeOnSelect();
    this.onStatus?.();
  }

  stopMusic() {
    if (!this.music?.node) { this.music = null; return; }
    try { this.music.node.stop(); } catch { /* already stopped */ }
    this.music.node.disconnect();
    this.music.gain?.disconnect();
    this.music = null;
  }

  dispose() {
    this._disposed = true;
    this.stopMusic();
    if (this.session && this._hooked) {
      this.session.removeEventListener('select', this._hooked);
      this._hooked = null;
    }
    // close(), not just suspend: the shell outlives every asset, and a context
    // per level switch would be a leak with a hard browser limit on it.
    this.ctx?.close().catch(() => {});
    this.ctx = null;
  }
}
