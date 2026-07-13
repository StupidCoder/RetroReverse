// The game's own filesystem.
//
// A cartridge has none, so the N64 target simply has no such panel. A disc has one, and
// the game reads it while it runs — so this is where you go to see what the machine
// could be loading, and what a file it just streamed actually contains.
//
// Selecting a file shows the head of it as hex, which is usually enough to recognise a
// format. It is deliberately not an extractor: the extract tools do that properly.

import { registerPanel } from './registry.js';
import { esc } from '../util.js';

registerPanel({
  id: 'files',
  title: 'Disc',
  slot: 'side',
  requires: 'files',
  grow: true,
  mount(body, ctx) {
    body.classList.add('mono', 'nopad');
    body.innerHTML = `<div id="fs-list" class="pad muted">reading…</div><div id="fs-head"></div>`;
    const list = document.getElementById('fs-list');
    const head = document.getElementById('fs-head');
    const note = document.getElementById('files-note');

    let path = '/';

    const go = (p) => {
      path = p;
      ctx.conn.send('fs.list', { path: p });
    };

    ctx.conn.on('files', (m) => {
      note.textContent = m.path;
      const up =
        m.path === '/'
          ? ''
          : `<div class="frow" data-up="1"><span class="fn">../</span></div>`;
      list.innerHTML =
        up +
        m.entries
          .map(
            (e) =>
              `<div class="frow" data-path="${esc(e.path)}" data-dir="${e.dir}">` +
              `<span class="fn">${esc(e.name)}${e.dir ? '/' : ''}</span>` +
              `<span class="grow"></span>` +
              (e.dir
                ? ''
                : `<span class="muted">${e.size.toLocaleString()} B · sector ${e.offset}</span>`) +
              `</div>`
          )
          .join('');

      list.querySelectorAll('.frow').forEach((row) => {
        row.onclick = () => {
          if (row.dataset.up) {
            const parent = m.path.replace(/\/[^/]+\/?$/, '') || '/';
            go(parent);
          } else if (row.dataset.dir === 'true') {
            go(row.dataset.path);
          } else {
            ctx.conn.send('fs.read', { path: row.dataset.path });
          }
        };
      });
    });

    ctx.conn.on('file', (m) => {
      const bytes = [];
      for (let i = 0; i < m.data.length; i += 2) bytes.push(parseInt(m.data.slice(i, i + 2), 16));
      let rows = '';
      for (let off = 0; off < Math.min(bytes.length, 256); off += 16) {
        const row = bytes.slice(off, off + 16);
        const hx = row.map((b) => b.toString(16).padStart(2, '0')).join(' ');
        const asc = row.map((b) => (b >= 32 && b < 127 ? String.fromCharCode(b) : '.')).join('');
        rows += `<div class="line"><span class="a">${off.toString(16).padStart(4, '0')}</span>${hx}<span class="asc">${esc(asc)}</span></div>`;
      }
      head.innerHTML =
        `<div class="fhead"><b>${esc(m.path)}</b> <span class="muted">${m.size.toLocaleString()} bytes</span></div>` +
        rows;
    });

    go('/');
  },
});
