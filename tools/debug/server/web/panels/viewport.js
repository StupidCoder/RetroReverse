// The framebuffer viewport: the image on one canvas, and everything we draw *about*
// the image — the command's pixel coverage, the picked pixel — on a second one above
// it. Plus the command scrubber, which is the reason the panel exists.
//
// The coverage overlay is the point of shipping the provenance buffer to the page: it
// holds one command index per pixel, so lighting up every pixel a command drew is a
// single local pass, with no request to the emulator at all.
//
// The same canvas is also where you TOUCH the machine, and the mode decides which. Paused,
// a click asks a question about a pixel that has already been drawn — which command drew
// it, what else wrote it — and that only means anything because there is a capture to ask
// of. Playing, there is no capture and no question; there is a machine running, and a
// click on the panel the stylus can reach is the stylus. So the two never contend: the
// pick path already bails while playing, and the touch path only exists while playing.

import { registerPanel } from './registry.js';
import { KIND_IMAGE, KIND_PROV, STREAM_MAIN } from '../conn.js';

const HIGHLIGHT = [255, 46, 166]; // --hot

registerPanel({
  id: 'viewport',
  title: 'Frame',
  slot: 'stage',
  requires: '',
  mount(body, ctx) {
    body.classList.add('stage-body');
    body.innerHTML = `
      <div id="stage-area">
        <div id="canvas-wrap">
          <canvas id="view"></canvas>
          <canvas id="overlay"></canvas>
        </div>
      </div>
      <div id="scrubber">
        <button id="prev" title="previous command (←)">◀</button>
        <input type="range" id="slider" min="0" max="0" value="0" disabled>
        <button id="next" title="next command (→)">▶</button>
        <span id="scrub-label" class="mono muted">—</span>
      </div>
      <div id="hover" class="mono muted">hover a pixel</div>`;

    const view = new Viewport(ctx);
    ctx.viewport = view; // the shell needs it to draw incoming images

    const slider = document.getElementById('slider');
    const label = document.getElementById('scrub-label');
    const hover = document.getElementById('hover');
    const canScrub = ctx.store.can('replay');

    slider.disabled = !canScrub;
    slider.oninput = () => ctx.ui.selectCommand(Number(slider.value));
    document.getElementById('prev').onclick = () => ctx.ui.selectCommand(ctx.store.get('selected') - 1);
    document.getElementById('next').onclick = () => ctx.ui.selectCommand(ctx.store.get('selected') + 1);

    ctx.store.on('frame', (frame) => {
      const n = frame ? frame.commands.length : 0;
      slider.max = Math.max(0, n - 1);
      slider.disabled = !canScrub || n === 0;
      if (!n) label.textContent = '—';
    });

    ctx.store.on('selected', (k) => {
      const frame = ctx.store.get('frame');
      if (!frame || k < 0) return;
      slider.value = k;
      label.textContent = `${k} / ${frame.commands.length - 1}`;
      view.select(k);
    });

    // The provenance buffer is a plane of the draw target, which is not always the
    // picture on screen — the 3DS renders into a padded VRAM buffer a DisplayTransfer
    // later crops. So it carries its own dimensions, and the view refuses to map a
    // pixel when they disagree with the image it is showing.
    ctx.store.on('prov', (p) => {
      const frame = ctx.store.get('frame');
      view.setProv(p, frame ? frame.w : 0, frame ? frame.h : 0);
    });
    ctx.store.on('pick', (p) => view.setPick(p));
    ctx.store.on('playing', (on) => {
      if (!on) return;
      hover.textContent = ctx.store.can('touch')
        ? 'playing — drag the touch screen to use the stylus'
        : 'playing — no capture';
    });

    view.onHover = (x, y) => {
      if (ctx.store.get('playing')) return; // nothing is captured while playing
      if (x < 0) {
        hover.textContent = 'hover a pixel';
        return;
      }
      const k = view.provAt(x, y);
      const frame = ctx.store.get('frame');
      const name = frame && k >= 0 && frame.commands[k] ? frame.commands[k].name : null;
      hover.textContent = name
        ? `pixel (${x}, ${y}) ← #${k} ${name}`
        : `pixel (${x}, ${y}) — untouched`;
    };

    view.onPick = (x, y) => {
      if (ctx.store.get('playing')) return;
      ctx.store.set({ pick: { x, y } });
      ctx.conn.send('frame.pixel', { x, y });
      const k = view.provAt(x, y);
      if (k >= 0) ctx.ui.selectCommand(k);
    };

    if (!ctx.store.can('touch')) return;

    // Where the stylus can reach is the target's answer, not ours, and on the 3DS it is
    // not fixed: the bottom screen joins the composed frame the first time the game
    // presents it, which changes the frame's size. So ask again whenever the size changes,
    // and whenever play begins.
    ctx.conn.on('touchpanels', (m) => view.setPanels(m.panels));
    view.onResize = () => ctx.conn.send('input.panels');
    view.onTouch = (panel, x, y, down) => ctx.conn.send('input.touch', { panel, x, y, down });

    ctx.store.on('playing', (on) => {
      view.setTouchable(on);
      if (on) ctx.conn.send('input.panels');
    });
    ctx.conn.send('input.panels');
  },
});

class Viewport {
  constructor(ctx) {
    this.ctx = ctx;
    this.view = document.getElementById('view');
    this.overlay = document.getElementById('overlay');
    this.wrap = document.getElementById('canvas-wrap');
    this.vctx = this.view.getContext('2d');
    this.octx = this.overlay.getContext('2d');

    this.w = 0;
    this.h = 0;
    this.zoom = 2;
    this.prov = null;
    this.provW = 0;
    this.provH = 0;
    this.selected = -1;
    this.pick = null;

    this.panels = []; // where the stylus can reach, in composed-frame coordinates
    this.touchable = false; // the machine is running, so a click is a touch
    this.pen = null; // the panel the pen is currently down on
    this.penPos = { x: 0, y: 0 }; // where it is, so pausing can lift it from there

    this.onHover = () => {};
    this.onPick = () => {};
    this.onTouch = () => {};
    this.onResize = () => {};

    this.overlay.addEventListener('mousemove', (e) => {
      const p = this.pointAt(e);
      if (!p) return;
      this.onHover(p.x, p.y);
      if (this.touchable) {
        this.overlay.style.cursor = this.panelAt(p.x, p.y) ? 'crosshair' : '';
      }
    });
    this.overlay.addEventListener('mouseleave', () => this.onHover(-1, -1));
    this.overlay.addEventListener('click', (e) => {
      const p = this.pointAt(e);
      if (p) this.onPick(p.x, p.y);
    });

    // The stylus. Pointer capture is what makes a drag behave like a stylus rather than
    // like a mouse: once the pen is down, every move belongs to the panel it went down on
    // — and, above all, the pen comes UP wherever the button is released, even if that is
    // outside the canvas or outside the window. A release that got lost would leave the
    // machine being touched for ever.
    this.overlay.addEventListener('pointerdown', (e) => {
      if (!this.touchable) return;
      const p = this.pointAt(e);
      const panel = p && this.panelAt(p.x, p.y);
      if (!panel) return;
      e.preventDefault();
      this.overlay.setPointerCapture(e.pointerId);
      this.pen = panel;
      this.emitTouch(panel, p, true);
    });
    this.overlay.addEventListener('pointermove', (e) => {
      if (!this.pen) return;
      this.emitTouch(this.pen, this.pointAt(e, true), true);
    });
    for (const ev of ['pointerup', 'pointercancel']) {
      this.overlay.addEventListener(ev, (e) => {
        if (!this.pen) return;
        this.emitTouch(this.pen, this.pointAt(e, true), false);
        this.pen = null;
      });
    }
  }

  // emitTouch reports the pen in the PANEL's own pixels, which are the touchscreen's own,
  // clamped to it: a drag that slides off the edge of the screen keeps the pen at the edge
  // rather than teleporting to wherever the mouse went.
  emitTouch(panel, p, down) {
    const x = clamp(p.x - panel.x, 0, panel.w - 1);
    const y = clamp(p.y - panel.y, 0, panel.h - 1);
    this.penPos = { x, y };
    this.onTouch(panel.id, x, y, down);
  }

  panelAt(x, y) {
    return this.panels.find((p) => x >= p.x && y >= p.y && x < p.x + p.w && y < p.y + p.h);
  }

  setPanels(panels) {
    this.panels = panels;
    this.redrawOverlay();
  }

  setTouchable(on) {
    this.touchable = on;
    // Pausing with the pen down would leave the machine touched at the moment it stopped,
    // and nothing would ever lift it. Lift it here.
    if (!on && this.pen) {
      this.onTouch(this.pen.id, this.penPos.x, this.penPos.y, false);
      this.pen = null;
    }
    if (!on) this.overlay.style.cursor = '';
    this.redrawOverlay();
  }

  // pointAt maps a mouse event to a pixel of the image. It returns null for a point
  // outside it — unless the caller says otherwise, which a stylus drag does: once the pen
  // is down, a move past the edge of the canvas is still a move of that pen, and it is
  // clamped onto the panel rather than dropped.
  pointAt(e, allowOutside = false) {
    const r = this.overlay.getBoundingClientRect();
    const x = Math.floor((e.clientX - r.left) / this.zoom);
    const y = Math.floor((e.clientY - r.top) / this.zoom);
    if (!allowOutside && (x < 0 || y < 0 || x >= this.w || y >= this.h)) return null;
    return { x, y };
  }

  resize(w, h) {
    if (w === this.w && h === this.h) return;
    this.w = w;
    this.h = h;
    for (const c of [this.view, this.overlay]) {
      c.width = w;
      c.height = h;
    }
    this.applyZoom();
    // The frame changed shape, so anything positioned within it may have moved. On the
    // 3DS this is exactly what happens the first time the game presents the bottom screen.
    this.onResize();
  }

  setZoom(z) {
    this.zoom = z;
    this.applyZoom();
  }

  applyZoom() {
    const cssW = `${this.w * this.zoom}px`;
    const cssH = `${this.h * this.zoom}px`;
    for (const c of [this.view, this.overlay]) {
      c.style.width = cssW;
      c.style.height = cssH;
    }
    this.wrap.style.width = cssW;
    this.wrap.style.height = cssH;
  }

  drawImage(imageData) {
    this.resize(imageData.width, imageData.height);
    this.vctx.putImageData(imageData, 0, 0);
    this.redrawOverlay();
  }

  setProv(prov, w, h) {
    this.prov = prov;
    this.provW = w;
    this.provH = h;
    this.redrawOverlay();
  }

  // provAligned reports whether the provenance plane and the picture on screen are
  // the same plane. When they are not, every answer provenance could give would be
  // off by the difference — a pixel attributed to a command that never touched the
  // buffer being shown. Better to say nothing than to say that.
  provAligned() {
    return !!this.prov && this.provW === this.w && this.provH === this.h;
  }

  // provAt is the whole reason the provenance buffer is shipped to the page: the
  // answer to "which command drew this pixel?" without a round trip.
  provAt(x, y) {
    if (!this.provAligned() || x < 0 || y < 0 || x >= this.w || y >= this.h) return -1;
    return this.prov[y * this.w + x];
  }

  select(k) {
    this.selected = k;
    this.redrawOverlay();
  }

  setPick(p) {
    this.pick = p;
    this.redrawOverlay();
  }

  redrawOverlay() {
    if (!this.w) return;
    this.octx.clearRect(0, 0, this.w, this.h);

    if (this.provAligned() && this.selected >= 0) {
      const img = this.octx.createImageData(this.w, this.h);
      const px = img.data;
      let hits = 0;
      for (let i = 0; i < this.prov.length; i++) {
        if (this.prov[i] !== this.selected) continue;
        const o = i * 4;
        px[o] = HIGHLIGHT[0];
        px[o + 1] = HIGHLIGHT[1];
        px[o + 2] = HIGHLIGHT[2];
        px[o + 3] = 150;
        hits++;
      }
      if (hits) this.octx.putImageData(img, 0, 0);
    }

    if (this.pick) {
      const { x, y } = this.pick;
      this.octx.strokeStyle = '#ffffff';
      this.octx.lineWidth = 1;
      this.octx.strokeRect(x - 1.5, y - 1.5, 4, 4);
    }

    // The touch zone, while the machine is running: the stylus can only reach one of the
    // two panels, and which one is not something the picture itself tells you.
    if (this.touchable) {
      this.octx.save();
      this.octx.strokeStyle = 'rgba(255, 46, 166, 0.55)';
      this.octx.lineWidth = 1;
      this.octx.setLineDash([3, 3]);
      for (const p of this.panels) {
        this.octx.strokeRect(p.x + 0.5, p.y + 0.5, p.w - 1, p.h - 1);
      }
      this.octx.restore();
    }
  }
}

function clamp(v, lo, hi) {
  return v < lo ? lo : v > hi ? hi : v;
}

export { Viewport, KIND_IMAGE, KIND_PROV, STREAM_MAIN };
