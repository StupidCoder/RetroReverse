// The shared animation runner: one ticker driving every animated subsystem of a
// level at the game's engine rate (meta.tickHz — 50 PAL, 60 GG/GB), generalized from
// the Sonic viewer's _advanceAnim. Subsystems register per level load and are all
// optional:
//
//  - tileAnims:  cycle atlas tiles in place (rebake callback per (tileId, atlasTile))
//  - cellAnims:  placement-anchored strips with per-phase hold times (Sonic $50)
//  - spriteSteps: object sprites playing (texture, hold) programs
//  - movePaths:  object sprites offset along per-frame (dx,dy) paths
//  - patrols:    pool placements shuttling over a [minDx,maxDx] span in stepPx
//                jumps, with per-facing phase art (Fort's tanks/prisoners/mines)
//  - custom fx:  anything with tick(step)/reset() (palette cycle, waterline — M4)
//
// Disabling animation resets every subsystem to step 0 — a deterministic state the
// screenshot harness (and the eye) can rely on.

export class AnimRunner {
  constructor(app, tickHz) {
    this.app = app;
    this.tickHz = tickHz || 60;
    this.enabled = true;
    this.reset();
    this._fn = () => this._tick(this.app.ticker.deltaMS);
    app.ticker.add(this._fn);
  }

  reset() {
    this.tileAnims = [];    // { anim:{tiles,frames,periodFrames}, paint(tileId, atlasTile), acc, step }
    this.cellAnims = [];    // { sprite, texs, durs, idx, acc }
    this.spriteSteps = [];  // { sprite, steps:[{tex,frames}], idx, acc }
    this.movePaths = [];    // { sprite, baseX, baseY, path, t }
    this.patrols = [];      // { group, baseX, dx, dir, dir0, minDx, maxDx, stepPx,
                            //   stepFrames, updatesPerStep, acc, upd, phase, phases }
    this.fx = [];           // { tick(frames), reset() }
  }

  _patrolShow(o, dir, phase) {
    if (!o.phases) return;
    for (const k of [1, -1]) for (const ph of o.phases[k]) ph.visible = false;
    const ph = o.phases[dir][phase % o.phases[dir].length];
    if (ph) ph.visible = true;
  }

  setEnabled(on) {
    this.enabled = on;
    if (on) return;
    // reset to the deterministic frame-0 state
    for (const a of this.tileAnims) {
      a.acc = 0; a.step = 0;
      a.anim.tiles.forEach((t, i) => a.paint(t, a.anim.frames[0] ? a.anim.frames[0][i] : t));
    }
    for (const c of this.cellAnims) { c.idx = 0; c.acc = 0; c.sprite.texture = c.texs[0]; }
    for (const o of this.spriteSteps) { o.idx = 0; o.acc = 0; o.sprite.texture = o.steps[0].tex; }
    for (const o of this.movePaths) {
      o.t = 0; o.sprite.x = o.baseX + o.path[0][0]; o.sprite.y = o.baseY + o.path[0][1];
    }
    for (const o of this.patrols) {
      o.acc = 0; o.upd = 0; o.phase = 0; o.dx = 0; o.dir = o.dir0;
      o.group.x = o.baseX;
      this._patrolShow(o, o.dir, 0);
    }
    for (const f of this.fx) f.reset();
  }

  // Deterministic advance for the screenshot harness: step exactly n engine frames.
  stepTo(n) {
    this.setEnabled(false);
    this.enabled = true;
    this._advance(n);
    this.enabled = false;
  }

  _tick(deltaMS) {
    if (!this.enabled) return;
    this._advance(deltaMS * this.tickHz / 1000);
  }

  _advance(df) {
    for (const a of this.tileAnims) {
      a.acc += df;
      const period = a.anim.periodFrames || 10;
      let changed = false;
      while (a.acc >= period) { a.acc -= period; a.step++; changed = true; }
      if (changed) {
        const fr = a.anim.frames[a.step % a.anim.frames.length];
        a.anim.tiles.forEach((t, i) => a.paint(t, fr[i]));
      }
    }
    for (const c of this.cellAnims) {
      c.acc += df;
      while (c.acc >= c.durs[c.idx]) {
        c.acc -= c.durs[c.idx];
        c.idx = (c.idx + 1) % c.texs.length;
        c.sprite.texture = c.texs[c.idx];
      }
    }
    for (const o of this.spriteSteps) {
      o.acc += df;
      while (o.acc >= o.steps[o.idx].frames) {
        o.acc -= o.steps[o.idx].frames;
        o.idx = (o.idx + 1) % o.steps.length;
        o.sprite.texture = o.steps[o.idx].tex;
      }
    }
    for (const o of this.movePaths) {
      o.t = (o.t + df) % o.path.length;
      const [dx, dy] = o.path[o.t | 0];
      o.sprite.x = o.baseX + dx;
      o.sprite.y = o.baseY + dy;
    }
    for (const o of this.patrols) {
      o.acc += df;
      while (o.acc >= o.stepFrames) {
        o.acc -= o.stepFrames;
        o.upd++;
        if (o.maxDx <= o.minDx) continue; // boxed in: stands still, no walk phases
        o.phase++;
        if (o.upd % o.updatesPerStep === 0) {
          // step stepPx in the current direction; at the span edge flip the
          // direction (and the facing) and move the other way in the same
          // update, like the engines' probe-reverse-retry loops
          let nd = o.dx + o.dir * o.stepPx;
          if (nd < o.minDx || nd > o.maxDx) {
            o.dir = -o.dir;
            nd = o.dx + o.dir * o.stepPx;
          }
          if (nd >= o.minDx && nd <= o.maxDx) o.dx = nd;
          o.group.x = o.baseX + o.dx;
        }
        this._patrolShow(o, o.dir, o.phase);
      }
    }
    for (const f of this.fx) f.tick(df);
  }

  destroy() {
    this.app.ticker.remove(this._fn);
    this.reset();
  }
}
