// The inspector panels: the overdraw history of the picked pixel, the CPU registers,
// a memory hex view, the disassembly around the PC, and the watch list.
//
// Each declares the capability it needs. A target that cannot report per-pixel writes
// has no overdraw panel; one that cannot disassemble has no disassembly panel. They
// are not hidden — they are never built.

import { registerPanel } from './registry.js';
import { esc, hex } from '../util.js';

// ---- overdraw ----

registerPanel({
  id: 'overdraw',
  title: 'Overdraw',
  slot: 'side',
  requires: 'frames',
  mount(body, ctx) {
    body.classList.add('mono');
    body.innerHTML = '<p class="muted">click a pixel to see every write to it, in order.</p>';

    ctx.conn.on('pixel', (m) => {
      // The answer to a click should arrive in front of you, not behind a tab.
      ctx.ui.reveal('overdraw');
      const note = document.getElementById('overdraw-note');
      note.textContent = `(${m.x}, ${m.y})`;

      if (!m.writes.length) {
        body.innerHTML =
          m.cmd < 0
            ? '<p class="muted">no command wrote this pixel.</p>'
            : `<p class="muted">last written by command ${m.cmd}. step with overdraw on to record the full history.</p>`;
        return;
      }
      // Including the writes the rasteriser produced and then threw away on a depth or
      // alpha test — usually the answer to "why is this pixel not the colour I expect?"
      body.innerHTML = m.writes
        .map((w) => {
          const tag = w.rejected ? '<span class="rej">rejected</span>' : 'drawn';
          const rgba = [w.r, w.g, w.b, w.a].map((v) => v.toString(16).padStart(2, '0')).join('');
          return (
            `<div class="write clickable" data-cmd="${w.cmd}">` +
            `<span class="sw" style="background:rgb(${w.r},${w.g},${w.b})"></span>` +
            `<span class="muted">#${w.cmd}</span> ${esc(w.name)} ` +
            `<span class="muted">${rgba}</span> ${tag}</div>`
          );
        })
        .join('');
      body.querySelectorAll('.write').forEach((el) => {
        el.addEventListener('click', () => ctx.ui.selectCommand(Number(el.dataset.cmd)));
      });
    });
  },
});

// ---- cpu ----

registerPanel({
  id: 'cpu',
  title: 'CPU',
  slot: 'bottom',
  requires: '',
  mount(body, ctx) {
    body.classList.add('mono');
    body.innerHTML = '<p class="muted">—</p>';

    ctx.conn.on('cpu', (m) => {
      ctx.store.set({ pc: m.pc });
      const extras = Object.entries(m.extra || {});
      let rows = `<tr><td class="n">pc</td><td>${m.pc}</td>`;
      rows += extras.length
        ? `<td class="n extra">${extras[0][0]}</td><td class="extra">${extras[0][1]}</td></tr>`
        : '<td></td><td></td></tr>';

      // Two register columns beside the extras, so the file fits without scrolling.
      const n = m.names.length;
      const half = Math.ceil(n / 2);
      for (let i = 0; i < half; i++) {
        const j = i + half;
        rows += '<tr>';
        rows += `<td class="n">${m.names[i]}</td><td>${m.vals[i]}</td>`;
        rows += j < n ? `<td class="n">${m.names[j]}</td><td>${m.vals[j]}</td>` : '<td></td><td></td>';
        const e = extras[i + 1];
        rows += e ? `<td class="n extra">${e[0]}</td><td class="extra">${e[1]}</td>` : '';
        rows += '</tr>';
      }
      body.innerHTML = `<table>${rows}</table>`;
    });
  },
});

// ---- memory ----

registerPanel({
  id: 'memory',
  title: 'Memory',
  slot: 'bottom',
  requires: '',
  mount(body, ctx) {
    body.classList.add('mono');
    const head = document.querySelector('[data-panel="memory"] h2');
    const addr = document.createElement('input');
    addr.type = 'text';
    addr.className = 'mono';
    addr.value = '00100000';
    addr.size = 8;
    addr.spellcheck = false;
    head.appendChild(addr);

    const read = () => {
      const a = parseInt(addr.value, 16);
      if (!Number.isNaN(a)) ctx.conn.send('mem.read', { addr: a, len: 256 });
    };
    addr.onchange = read;
    ctx.ui.readMem = read;

    ctx.conn.on('mem', (m) => {
      // The region list, when the target knows its own map, turns a hex dump into
      // something you can navigate.
      if (m.regions && !head.querySelector('select')) {
        const sel = document.createElement('select');
        sel.innerHTML =
          '<option value="">region…</option>' +
          m.regions.map((r) => `<option value="${r.lo}">${esc(r.name)}</option>`).join('');
        sel.onchange = () => {
          if (!sel.value) return;
          addr.value = sel.value;
          read();
        };
        head.appendChild(sel);
      }

      const bytes = [];
      for (let i = 0; i < m.data.length; i += 2) bytes.push(parseInt(m.data.slice(i, i + 2), 16));
      let html = '';
      for (let off = 0; off < bytes.length; off += 16) {
        const row = bytes.slice(off, off + 16);
        const hx = row.map((b) => b.toString(16).padStart(2, '0')).join(' ');
        const asc = row.map((b) => (b >= 32 && b < 127 ? String.fromCharCode(b) : '.')).join('');
        html += `<div class="line"><span class="a">${hex(m.addr + off, 8)}</span>${hx}<span class="asc">${esc(asc)}</span></div>`;
      }
      body.innerHTML = html || '<p class="muted">—</p>';
    });
  },
});

// ---- disassembly ----

registerPanel({
  id: 'disasm',
  title: 'Disassembly',
  slot: 'side',
  requires: 'disasm',
  mount(body, ctx) {
    body.classList.add('mono', 'nopad');
    body.innerHTML = '<p class="muted pad">step the CPU to see where it is.</p>';

    let bps = new Set();

    const refresh = () => ctx.conn.send('cpu.disasm', { addr: 0, n: 64 });
    ctx.ui.refreshDisasm = refresh;

    ctx.conn.on('disasm', (m) => {
      bps = new Set(m.bps || []);
      body.innerHTML = m.instr
        .map((in_) => {
          const here = in_.addr === m.pc ? ' pc' : '';
          const bp = bps.has(in_.addr) ? ' bp' : '';
          return (
            `<div class="ins${here}${bp}" data-addr="${in_.addr}">` +
            `<span class="gut">${bps.has(in_.addr) ? '●' : ''}</span>` +
            `<span class="a">${in_.addr.slice(-8)}</span>${esc(in_.text)}</div>`
          );
        })
        .join('');
      // Clicking the gutter toggles a breakpoint — the usual place to expect it.
      body.querySelectorAll('.ins').forEach((el) => {
        el.querySelector('.gut').addEventListener('click', () => {
          const a = el.dataset.addr;
          const pc = parseInt(a, 16);
          ctx.conn.send(bps.has(a) ? 'bp.clear' : 'bp.set', { pc });
          setTimeout(refresh, 30);
        });
      });
    });

    // The PC moved: follow it.
    ctx.conn.on('stopped', () => {
      refresh();
      ctx.conn.send('cpu.regs');
    });
  },
});

// ---- watches ----

registerPanel({
  id: 'watches',
  title: 'Watches',
  slot: 'bottom',
  requires: 'watch',
  mount(body, ctx) {
    body.classList.add('mono');
    body.innerHTML = `
      <div class="watch-add">
        <select id="w-kind"><option value="write">write</option><option value="read">read</option></select>
        <input type="text" id="w-lo" class="mono" size="8" placeholder="lo" spellcheck="false">
        <input type="text" id="w-hi" class="mono" size="8" placeholder="hi" spellcheck="false">
        <label class="check"><input type="checkbox" id="w-break"> break</label>
        <button id="w-add">watch</button>
      </div>
      <div id="w-list"></div>
      <div id="w-hits"></div>`;

    document.getElementById('w-add').onclick = () => {
      const lo = parseInt(document.getElementById('w-lo').value, 16);
      const hi = parseInt(document.getElementById('w-hi').value, 16) || lo + 4;
      if (Number.isNaN(lo)) return;
      ctx.conn.send('mem.watch', {
        kind: document.getElementById('w-kind').value,
        lo,
        hi,
        break: document.getElementById('w-break').checked,
      });
    };

    ctx.conn.on('watches', (m) => {
      const list = document.getElementById('w-list');
      list.innerHTML = m.watches
        .map(
          (w) =>
            `<div class="wrow"><span class="muted">#${w.id}</span> ${w.kind} ` +
            `${w.lo}..${w.hi}${w.break ? ' <span class="rej">break</span>' : ''} ` +
            `<span class="muted">${w.hits} hits</span> ` +
            `<button data-id="${w.id}" class="tiny">clear</button></div>`
        )
        .join('');
      list.querySelectorAll('button').forEach((b) => {
        b.onclick = () => ctx.conn.send('mem.unwatch', { id: Number(b.dataset.id) });
      });
    });

    // Hits arrive as events, from inside the run. Keep the most recent few: a hot
    // address would otherwise bury the pane (and the socket) in a single frame.
    const hits = [];
    ctx.conn.on('hit', (h) => {
      hits.unshift(h);
      if (hits.length > 50) hits.pop();
      document.getElementById('w-hits').innerHTML = hits
        .map(
          (x) =>
            `<div class="hit"><span class="muted">${x.kind}</span> ${x.addr} ` +
            `<span class="muted">=</span> ${x.val} ` +
            `<span class="muted">by</span> ${x.pc} ${esc(x.instr || '')}</div>`
        )
        .join('');
    });
  },
});
