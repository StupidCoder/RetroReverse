// Memory as a picture.
//
// The fixed surfaces are the obvious ones — what the machine is scanning out, what it is
// drawing into, all of VRAM at once. The *free* surface is the one that earns its keep:
// aim it at an address and read it as a texture. Point it at where a DMA just landed and
// you can see whether what arrived is the texture the game meant to upload, which is a
// question a hex dump answers very badly.
//
// The pixel formats are the platform's, and so is the meaning of the address (the PSX's
// VRAM is a grid of 16-bit pixels, not bytes). The panel just offers what the target
// says it has.

import { registerPanel } from './registry.js';
import { KIND_IMAGE, STREAM_SURFACE } from '../conn.js';
import { esc } from '../util.js';

registerPanel({
  id: 'surface',
  title: 'Memory as an image',
  slot: 'side',
  requires: 'surfaces',
  grow: true,
  mount(body, ctx) {
    body.innerHTML = `
      <div class="surf-controls">
        <select id="surf-id"></select>
        <span id="surf-free" hidden>
          <label>at <input type="text" id="surf-addr" class="mono" size="8" value="0" spellcheck="false"></label>
          <label><input type="text" id="surf-w" class="mono" size="4" value="64"></label>
          ×
          <label><input type="text" id="surf-h" class="mono" size="4" value="64"></label>
          <select id="surf-fmt"></select>
          <label>pal <input type="text" id="surf-pal" class="mono" size="8" value="0" spellcheck="false"></label>
          <label>stride <input type="text" id="surf-stride" class="mono" size="4" value="0"></label>
        </span>
        <button id="surf-go">Draw</button>
        <label class="check"><input type="checkbox" id="surf-fit" checked> fit</label>
      </div>
      <div id="surf-stage"><canvas id="surf-canvas"></canvas></div>
      <div id="surf-note" class="muted"></div>`;

    const sel = document.getElementById('surf-id');
    const free = document.getElementById('surf-free');
    const fmt = document.getElementById('surf-fmt');
    const canvas = document.getElementById('surf-canvas');
    const note = document.getElementById('surf-note');
    const cctx = canvas.getContext('2d');

    let surfaces = [];
    let want = -1; // the request whose image we still want

    const current = () => surfaces.find((s) => s.id === sel.value);

    const hexOf = (id) => {
      const v = parseInt(document.getElementById(id).value.trim(), 16);
      return Number.isNaN(v) ? 0 : v;
    };
    const intOf = (id) => {
      const v = parseInt(document.getElementById(id).value.trim(), 10);
      return Number.isNaN(v) ? 0 : v;
    };

    const draw = () => {
      const s = current();
      if (!s) return;
      const args = { id: s.id };
      if (s.free) {
        Object.assign(args, {
          addr: hexOf('surf-addr'),
          w: intOf('surf-w'),
          h: intOf('surf-h'),
          stride: intOf('surf-stride'),
          format: fmt.value,
          palette: hexOf('surf-pal'),
        });
      }
      want = ctx.conn.send('surface.render', args);
    };

    const onPick = () => {
      const s = current();
      if (!s) return;
      free.hidden = !s.free;
      if (s.free) {
        fmt.innerHTML = s.formats.map((f) => `<option>${esc(f)}</option>`).join('');
      }
      draw();
    };

    sel.onchange = onPick;
    document.getElementById('surf-go').onclick = draw;
    for (const id of ['surf-addr', 'surf-w', 'surf-h', 'surf-fmt', 'surf-pal', 'surf-stride']) {
      document.getElementById(id).addEventListener('change', draw);
    }
    document.getElementById('surf-fit').onchange = () => fit();

    function fit() {
      const on = document.getElementById('surf-fit').checked;
      canvas.style.width = on ? '100%' : `${canvas.width}px`;
      canvas.style.height = on ? 'auto' : `${canvas.height}px`;
    }

    ctx.conn.on('surfaces', (m) => {
      surfaces = m.surfaces;
      sel.innerHTML = surfaces.map((s) => `<option value="${esc(s.id)}">${esc(s.name)}</option>`).join('');
      onPick();
    });

    ctx.conn.on('render', (m) => {
      if (m.stream !== STREAM_SURFACE) return;
      note.textContent = `${(m.bytes / 1024).toFixed(0)} KB · ${m.renderMs.toFixed(1)} ms`;
    });

    ctx.conn.onBinary((m) => {
      if (m.kind !== KIND_IMAGE || m.stream !== STREAM_SURFACE) return;
      if (m.seq !== want) return; // a stale answer to an aim we have already moved on from
      canvas.width = m.w;
      canvas.height = m.h;
      cctx.putImageData(m.image, 0, 0);
      fit();
    });

    // Redrawing after a frame step keeps the picture honest: memory has moved on.
    ctx.store.on('frame', () => draw());

    ctx.conn.send('surface.list');
  },
});
