// The display-processor command list, and the detail of the selected command.
//
// A frame runs to thousands of commands, so the list is windowed: a spacer gives the
// scrollbar its true height and only the visible rows exist in the DOM.
//
// What counts as one "command" is the platform's own granularity — an RDP command on
// N64, a GP0 primitive on PSX — so the pane takes its title from the target.

import { registerPanel } from './registry.js';
import { esc } from '../util.js';

const ROW = 18; // px, must match .row height in style.css

registerPanel({
  id: 'commands',
  title: 'GPU commands',
  slot: 'side',
  requires: 'frames',
  mount(body, ctx) {
    body.classList.add('nopad');
    body.innerHTML = `<div id="cmd-spacer"><div id="cmd-rows"></div></div>`;
    const note = document.getElementById('commands-note');
    const spacer = document.getElementById('cmd-spacer');
    const rows = document.getElementById('cmd-rows');

    let cmds = [];

    const render = () => {
      const selected = ctx.store.get('selected');
      const first = Math.max(0, Math.floor(body.scrollTop / ROW) - 2);
      const last = Math.min(cmds.length, first + Math.ceil(body.clientHeight / ROW) + 4);
      rows.style.top = `${first * ROW}px`;
      let html = '';
      for (let i = first; i < last; i++) {
        const sel = i === selected ? ' sel' : '';
        html += `<div class="row${sel}" data-i="${i}"><span class="idx">${i}</span><span class="nm">${esc(cmds[i].name)}</span></div>`;
      }
      rows.innerHTML = html;
    };

    const reveal = (k) => {
      const top = body.scrollTop;
      const y = k * ROW;
      if (y < top || y + ROW > top + body.clientHeight) {
        body.scrollTop = Math.max(0, y - body.clientHeight / 2);
      }
    };

    body.addEventListener('scroll', render);
    rows.addEventListener('click', (e) => {
      const row = e.target.closest('.row');
      if (row) ctx.ui.selectCommand(Number(row.dataset.i));
    });

    // The list only builds the rows that fit, so it needs its own height — which is zero
    // while it sits behind another tab. Re-render whenever the pane actually has a size:
    // that covers being shown for the first time, and being promoted to the stage.
    new ResizeObserver(() => {
      if (body.clientHeight) render();
    }).observe(body);

    ctx.store.on('frame', (frame) => {
      cmds = frame ? frame.commands : [];
      spacer.style.height = `${cmds.length * ROW}px`;
      note.textContent = cmds.length ? `${cmds.length.toLocaleString()} commands` : 'nothing drawn';
      body.scrollTop = 0;
      render();
      detail(null);
    });

    ctx.store.on('selected', (k) => {
      render();
      if (k >= 0 && cmds[k]) {
        reveal(k);
        detail(cmds[k], k);
      }
    });
  },
});

registerPanel({
  id: 'detail',
  title: 'Command detail',
  slot: 'side',
  requires: 'frames',
  mount(body) {
    body.classList.add('mono');
    body.innerHTML = '<p class="muted">select a command.</p>';
  },
});

function detail(c, k) {
  const el = document.getElementById('detail-body');
  if (!el) return;
  if (!c) {
    el.innerHTML = '<p class="muted">select a command.</p>';
    return;
  }
  const words = c.words.map((w) => `<p>${w}</p>`).join('');
  const dec = c.decoded ? `<p>${esc(c.decoded)}</p>` : '';
  el.innerHTML =
    `<p><span class="muted">#${k}</span> <b>${esc(c.name)}</b> ` +
    `<span class="muted">op 0x${c.op.toString(16).padStart(2, '0')}</span></p>` +
    dec +
    words;
}
