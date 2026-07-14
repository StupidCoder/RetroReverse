// Savestates.
//
// The point is not only to skip a boot. It is to save the exact frame where something
// is wrong and hand it to someone else — so each state shows the command line that
// resumes the game there, in the flag vocabulary that game's own oracle actually
// accepts (the oracles never agreed: -loadstate, -load, -state). Copy it, paste it into
// a shell or a Claude Code session, and you land on the same frame.

import { registerPanel } from './registry.js';
import { esc } from '../util.js';

registerPanel({
  id: 'states',
  rank: 10,
  title: 'Savestates',
  slot: 'bottom-left',
  requires: 'states',
  mount(body, ctx) {
    body.classList.add('mono');
    body.innerHTML = `
      <div class="watch-add">
        <input type="text" id="st-name" class="mono" size="16" placeholder="name" spellcheck="false">
        <button id="st-save">Save state</button>
      </div>
      <div id="st-list"><p class="muted">—</p></div>`;

    const list = document.getElementById('st-list');
    const name = document.getElementById('st-name');

    const refresh = () => ctx.conn.send('state.list');
    document.getElementById('st-save').onclick = () => {
      const n = name.value.trim();
      if (!n) {
        ctx.ui.status('a savestate needs a name');
        return;
      }
      ctx.conn.send('state.save', { name: n });
    };
    name.onkeydown = (e) => {
      if (e.key === 'Enter') document.getElementById('st-save').click();
    };

    ctx.conn.on('states', (m) => {
      if (m.note) {
        list.innerHTML = `<p class="muted">${esc(m.note)}</p>`;
        return;
      }
      if (!m.states.length) {
        list.innerHTML = `<p class="muted">none yet — they go in ${esc(m.dir)}</p>`;
        return;
      }
      list.innerHTML = m.states
        .map(
          (s) =>
            `<div class="strow" data-name="${esc(s.name)}">` +
            `<b>${esc(s.name)}</b> <span class="muted">${esc(s.when)} · ${(s.size / 1024).toFixed(0)} KB</span>` +
            `<span class="grow"></span>` +
            `<button class="tiny st-load">load</button>` +
            `<button class="tiny st-copy" title="${esc(s.resume)}">copy resume cmd</button>` +
            `<button class="tiny st-del">delete</button>` +
            `</div>`
        )
        .join('');

      list.querySelectorAll('.strow').forEach((row, i) => {
        const s = m.states[i];
        row.querySelector('.st-load').onclick = () => ctx.conn.send('state.load', { name: s.name });
        row.querySelector('.st-del').onclick = () => ctx.conn.send('state.delete', { name: s.name });
        row.querySelector('.st-copy').onclick = async () => {
          if (!s.resume) {
            ctx.ui.status('this target cannot say how to resume itself');
            return;
          }
          try {
            await navigator.clipboard.writeText(s.resume);
            ctx.ui.status(`copied: ${s.resume}`);
          } catch {
            // Clipboard access can be refused; showing the line is still useful.
            ctx.ui.status(s.resume);
          }
        };
      });
    });

    // An op that changed the state set says so with an ok; ask again.
    ctx.conn.on('ok', (m) => {
      if (m.op === 'state.save' || m.op === 'state.load') {
        ctx.ui.status(`${m.op.slice(6)}d ${m.note || ''}`);
        refresh();
        if (m.op === 'state.load') ctx.ui.afterStateLoad();
      }
    });

    refresh();
  },
});
