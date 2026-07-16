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
import { KIND_IMAGE, KIND_PROV, STREAM_MAIN, STREAM_SURFACE } from '../conn.js';
import { esc } from '../util.js';

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
      </div>`;

    const view = new Viewport(ctx);
    ctx.viewport = view; // the shell needs it to draw incoming images

    const slider = document.getElementById('slider');
    const label = document.getElementById('scrub-label');
    const canScrub = ctx.store.can('replay');

    // What the pixel under the cursor was drawn by goes in the panel's own header, where
    // there is already a line, rather than on a line of its own under the scrubber — the
    // stage is the one place worth spending vertical space on.
    const hover = document.getElementById('viewport-note');
    const say = (s) => (hover.textContent = s);

    // The buffer selector: the frame is the default, and everything else the machine can
    // show as a picture is one pick away — the two screens of a DS or a 3DS, the 3D core's
    // own render buffer, the PICA's draw target and its depth buffer, all of the PSX's
    // VRAM. It is the same list the "Memory as an image" panel offers, less the free
    // texture surface, which needs an address and a format and so needs that panel's boxes.
    mountBufferPicker(ctx, view, say);

    // The scrubber's positions are the commands that WROTE to memory (ctx.ui.scrubList), so
    // dragging it walks the frame being built rather than the register writes that set it
    // up. The command list is still the whole stream, and picking any command from it still
    // shows the frame as it stood after it — the handle then rests on the last write before
    // it, which is the picture you are looking at.
    slider.disabled = !canScrub;
    slider.oninput = () => ctx.ui.selectCommand(ctx.ui.scrubList().at(Number(slider.value)));
    document.getElementById('prev').onclick = () => ctx.ui.stepScrub(-1);
    document.getElementById('next').onclick = () => ctx.ui.stepScrub(1);

    ctx.store.on('frame', () => {
      const s = ctx.ui.scrubList();
      slider.max = Math.max(0, s.n - 1);
      slider.disabled = !canScrub || s.n === 0;
      if (!s.n) label.textContent = '—';
    });

    ctx.store.on('selected', (k) => {
      const frame = ctx.store.get('frame');
      if (!frame || k < 0) return;
      const s = ctx.ui.scrubList();
      const at = s.indexOf(k);
      if (at >= 0) slider.value = at;
      // What the handle is at, and what it is at OF: "#1,182 · draw 7/243" — the command,
      // then where that sits among the frame's writes. On a platform that does not report
      // pixels there are no writes to count, and it says the plain thing.
      label.textContent = s.all
        ? `${k} / ${frame.commands.length - 1}`
        : `#${k} · draw ${at + 1}/${s.n}`;
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
      say(ctx.store.can('touch') ? 'playing — drag the touch screen to use the stylus' : 'playing — no capture');
    });

    view.onHover = (x, y) => {
      if (ctx.store.get('playing')) return; // nothing is captured while playing
      if (!ctx.store.can('frames')) return; // a CPU-only target has no per-pixel provenance to hover
      if (x < 0) {
        say('');
        return;
      }
      const k = view.provAt(x, y);
      const frame = ctx.store.get('frame');
      const name = frame && k >= 0 && frame.commands[k] ? frame.commands[k].name : null;
      say(name ? `pixel (${x}, ${y}) ← #${k} ${name}` : `pixel (${x}, ${y}) — untouched`);
    };

    view.onPick = (x, y) => {
      if (ctx.store.get('playing')) return;
      if (!ctx.store.can('frames')) return; // no capture to interrogate on a CPU-only target
      ctx.store.set({ pick: { x, y } });
      ctx.conn.send('frame.pixel', { x, y });
      const k = view.provAt(x, y);
      if (k >= 0) ctx.ui.selectCommand(k);
    };

    // Keyboard input (the Keyer). When the target takes keys, the frame panel captures
    // keydown/keyup while it is focused and forwards each as input.key — click the picture
    // to focus it, then type to drive the game's menus. Focus-gating is what keeps the
    // debugger's own shortcuts (space, arrows) usable: a keystroke meant for the game is
    // stopped from bubbling to the document handler, but ONLY while the frame has focus, so
    // clicking away hands the keyboard back to the debugger.
    //
    // A press and its release are a make and a break scancode, and both must reach the
    // machine — so nothing is coalesced here, and hardware auto-repeat (e.repeat) is
    // dropped: one physical press is one make, held until the key comes up.
    if (ctx.store.can('keys')) {
      const wrap = document.getElementById('canvas-wrap');
      wrap.tabIndex = 0;
      wrap.classList.add('keyable');
      view.overlay.addEventListener('pointerdown', () => wrap.focus());
      const sendKey = (e, down) => {
        if (e.repeat) return;
        ctx.conn.send('input.key', { name: e.key, code: e.keyCode || 0, down });
        e.preventDefault();
        e.stopPropagation();
      };
      wrap.addEventListener('keydown', (e) => sendKey(e, true));
      wrap.addEventListener('keyup', (e) => sendKey(e, false));
      say('click the frame, then type to drive the game (arrows = stick, Enter = start)');
    }

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

// mountBufferPicker puts the buffer selector in the panel's header, ahead of everything
// else it says — it names what you are looking at, so it reads first.
//
// The surfaces are the target's own answer (debug.Surfacer), which is why this is one
// control rather than a per-platform list: the DS offers its two screens and the 3D core's
// render buffer, the 3DS its PICA draw target and depth buffer, the PSX all of VRAM, and
// none of that is written down here.
function mountBufferPicker(ctx, view, say) {
  if (!ctx.store.can('surfaces')) return;

  const head = document.querySelector('[data-panel="viewport"] h2');
  const sel = document.createElement('select');
  sel.id = 'vp-buffer';
  sel.title = 'which buffer of the machine to show';
  sel.innerHTML = '<option value="">Frame</option>';
  head.insertBefore(sel, document.getElementById('viewport-note'));

  ctx.conn.on('surfaces', (m) => {
    const fixed = m.surfaces.filter((s) => !s.free); // the free one needs an address and a format
    sel.innerHTML =
      '<option value="">Frame</option>' +
      fixed.map((s) => `<option value="${esc(s.id)}">${esc(s.name)}</option>`).join('');
    sel.value = view.buffer; // a re-listed surface set must not silently change what is shown
  });

  sel.onchange = () => {
    view.setBuffer(sel.value);
    say('');
    if (view.buffer) {
      view.drawBuffer();
      return;
    }
    ctx.ui.redraw(); // back to the frame: whatever the scrubber and the scanout box say
  };

  // Surface images ride their own stream, so this and the "Memory as an image" panel can
  // each have one in flight without drawing the other's.
  ctx.conn.onBinary((m) => {
    if (m.kind !== KIND_IMAGE || m.stream !== STREAM_SURFACE) return;
    if (!view.buffer || m.seq !== view.bufSeq) return;
    view.drawImage(m.image);
  });

  // While the machine free-runs there is no capture and the stream is the scanout, so a
  // buffer cannot be held on screen — and a half-drawn one would be a lie about a machine
  // that has moved on. Pausing captures a frame, and the frame redraws the buffer.
  ctx.store.on('playing', (on) => {
    sel.disabled = on;
  });

  // A step, a pause, a state load: memory has moved on and so has the buffer.
  ctx.store.on('frame', () => view.drawBuffer());

  ctx.conn.send('surface.list');
}

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
    this.zoom = 2; // the fallback if fit is turned off; the toolbar's own default is Fit
    this.fit = true; // scale the picture to the panel instead of to a whole number
    this.buffer = ''; // a buffer of the machine instead of the frame, by surface id
    this.bufSeq = -1;
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

    // Fit is a rule about the panel, so the panel changing size re-applies it: promoting
    // the viewport to the stage, maximizing it, dragging the window narrower.
    new ResizeObserver(() => {
      if (this.fit) this.applyZoom();
    }).observe(document.getElementById('stage-area'));

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

  // setZoom takes the toolbar's value: a whole-number zoom, or "fit".
  //
  // Fit is not a zoom level, it is a rule — as large as the panel allows, aspect kept — so
  // it has to be recomputed whenever either side of that changes: a new picture (a DS frame
  // grows when the game presents its second screen) or a new panel size (promoting the
  // viewport to the stage, maximizing it, resizing the window).
  setZoom(z) {
    this.fit = z === 'fit';
    if (!this.fit) this.zoom = Number(z) || 1;
    this.applyZoom();
  }

  fitScale() {
    const area = document.getElementById('stage-area');
    if (!area || !this.w || !this.h) return 1;
    const pad = 8; // the checkered surround should stay visible as a surround
    const aw = area.clientWidth - pad;
    const ah = area.clientHeight - pad;
    if (aw <= 0 || ah <= 0) return 1;
    return Math.max(0.05, Math.min(aw / this.w, ah / this.h));
  }

  applyZoom() {
    if (this.fit) this.zoom = this.fitScale();
    const cssW = `${this.w * this.zoom}px`;
    const cssH = `${this.h * this.zoom}px`;
    for (const c of [this.view, this.overlay]) {
      c.style.width = cssW;
      c.style.height = cssH;
    }
    this.wrap.style.width = cssW;
    this.wrap.style.height = cssH;
  }

  // setBuffer switches the stage between the frame and one of the machine's own buffers.
  setBuffer(id) {
    this.buffer = id;
    // Provenance is a plane of the frame's DRAW TARGET. Another buffer is another plane —
    // and one that is often the same size, which is worse than a different one, because the
    // overlay would happily light up pixels of a picture those commands never touched.
    if (id) this.setProv(null, 0, 0);
  }

  // drawBuffer asks for the selected buffer as it stands now. This is a read of live memory
  // rather than a replay, which is exactly what makes it worth having: it says where the
  // machine IS, even while the scrubber has the frame wound back into the middle.
  drawBuffer() {
    if (!this.buffer) return false;
    this.bufSeq = this.ctx.conn.send('surface.render', { id: this.buffer });
    this.ctx.conn.onError(this.bufSeq, (m) => this.ctx.ui.status(m.msg.replace(/^\w+adapter: /, '')));
    return true;
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
