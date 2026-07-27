// sprite2d.js — the runtime for Retro-X sprite2d objects (RETROX.md §6.1):
// one PNG atlas per object, row per animation, column per frame. Instances
// play durations/steps programs with the four loop modes, follow per-frame
// movement paths, and mirror through the `mirror` link or a plain h-flip.

import { Container, Rectangle, Sprite, Texture } from 'pixi.js';

export async function loadImage(url) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error(`image failed: ${url}`));
    img.src = url;
  });
}

export class SpriteObject {
  // game: data.Game; id: object asset id
  static async load(game, id) {
    const { asset, doc, docPath } = await game.object(id);
    if (doc.type !== 'sprite2d') throw new Error(`${id} is not sprite2d`);
    const img = await loadImage(game.url(docPath, doc.atlas.file));
    return new SpriteObject(asset, doc, img);
  }

  constructor(asset, doc, img) {
    this.asset = asset;
    this.doc = doc;
    this.img = img;
    const src = Texture.from(img).source;
    src.scaleMode = 'nearest';
    this.src = src;
    this.cellW = doc.atlas.cellW;
    this.cellH = doc.atlas.cellH;
    this.anims = new Map();
    for (const a of doc.animations || []) this.anims.set(a.id, a);
    this._frames = new Map(); // "row/col" -> Texture
  }

  frame(row, col) {
    const k = `${row}/${col}`;
    if (!this._frames.has(k)) {
      this._frames.set(k, new Texture({
        source: this.src,
        frame: new Rectangle(col * this.cellW, row * this.cellH, this.cellW, this.cellH),
      }));
    }
    return this._frames.get(k);
  }

  // makeInstance builds a playable display object.
  //   opts: { anim, tint ('#rrggbb'), hflip }
  // Returns { node, tick(dfFrames), setAnim(id), stop() } — tick advances the
  // program in ENGINE FRAMES (the caller owns the clock).
  makeInstance(opts = {}) {
    const inst = new SpriteInstance(this, opts);
    return inst;
  }
}

// expandProgram turns an animation into a flat [(frame, hold)] list.
function expandProgram(a) {
  if (a.steps?.length) return a.steps.map(([f, h]) => [f, Math.max(1, h)]);
  const n = a.frames || 1;
  const out = [];
  for (let i = 0; i < n; i++) out.push([i, Math.max(1, a.durations ? a.durations[i] : 1)]);
  return out;
}

export class SpriteInstance {
  constructor(obj, { anim, tint, hflip } = {}) {
    this.obj = obj;
    this.node = new Container();
    this.sprite = new Sprite();
    this.node.addChild(this.sprite);
    if (tint) this.sprite.tint = parseInt(tint.slice(1), 16);
    this.baseHflip = !!hflip;
    this.setAnim(anim || obj.doc.animations?.[0]?.id);
  }

  setAnim(id) {
    const a = this.obj.anims.get(id) || this.obj.doc.animations?.[0];
    if (!a) return;
    this.anim = a;
    this.prog = expandProgram(a);
    this.idx = 0;
    this.acc = 0;
    this.dir = 1;          // pingpong direction
    this.done = false;
    this.pathT = 0;
    const ax = a.anchor || this.obj.doc.atlas.anchor || [0, 0];
    this.sprite.pivot.set(ax[0], ax[1]);
    // A mirrored placement uses the linked left art when the object has one,
    // otherwise a horizontal flip of the base row.
    let flip = this.baseHflip;
    if (flip && a.mirror && this.obj.anims.has(a.mirror)) {
      this.anim = this.obj.anims.get(a.mirror);
      this.prog = expandProgram(this.anim);
      flip = false;
    }
    this.sprite.scale.x = flip ? -1 : 1;
    this._show();
  }

  _show() {
    const [f] = this.prog[this.idx];
    this.sprite.texture = this.obj.frame(this.anim.row || 0, f);
  }

  // tick advances by df engine frames; also applies the movement path.
  tick(df) {
    const a = this.anim;
    if (this.prog.length > 1 && !this.done) {
      this.acc += df;
      while (this.acc >= this.prog[this.idx][1]) {
        this.acc -= this.prog[this.idx][1];
        this._advance();
        if (this.done) break;
      }
      this._show();
    }
    if (a.path?.length) {
      this.pathT = (this.pathT + df) % a.path.length;
      const [dx, dy] = a.path[this.pathT | 0];
      this.sprite.position.set(dx, dy);
    }
  }

  _advance() {
    const n = this.prog.length;
    const loop = this.anim.loop || 'loop';
    if (loop === 'pingpong') {
      let next = this.idx + this.dir;
      if (next >= n || next < 0) { this.dir = -this.dir; next = this.idx + this.dir; }
      this.idx = Math.max(0, Math.min(n - 1, next));
      return;
    }
    if (this.idx + 1 >= n) {
      if (loop === 'loop') this.idx = 0;
      else if (loop === 'once') { this.idx = 0; this.done = true; }
      else this.done = true; // hold: stay on the last frame
      return;
    }
    this.idx++;
  }

  reset() {
    this.idx = 0; this.acc = 0; this.dir = 1; this.done = false; this.pathT = 0;
    this._show();
    this.sprite.position.set(0, 0);
  }

  // ---- transport support: engine-frame position within one program cycle ----

  cycleFrames() {
    return this.prog.reduce((s, [, h]) => s + h, 0);
  }

  pos() {
    let f = this.acc;
    for (let i = 0; i < this.idx; i++) f += this.prog[i][1];
    return f;
  }

  // seek scrubs to an engine-frame position in the forward pass (pingpong
  // scrubs its forward half — good enough for a slider).
  seek(f) {
    f = Math.max(0, Math.min(this.cycleFrames() - 0.001, f));
    this.done = false;
    this.dir = 1;
    this.idx = 0;
    this.acc = f;
    while (this.idx < this.prog.length - 1 && this.acc >= this.prog[this.idx][1]) {
      this.acc -= this.prog[this.idx][1];
      this.idx++;
    }
    this._show();
    if (this.anim.path?.length) {
      this.pathT = f % this.anim.path.length;
      const [dx, dy] = this.anim.path[this.pathT | 0];
      this.sprite.position.set(dx, dy);
    }
  }
}
