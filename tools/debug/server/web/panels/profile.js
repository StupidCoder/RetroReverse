// Where the frame's time went.
//
// The panel exists to answer one question honestly — "is it 90% rasterising?" — and to
// stop the answer being guessed. A frame's cost is shown as the machine's own buckets
// (a command list, a draw, a texture cache miss, an audio frame, a supervisor call),
// with what nobody timed reported as a remainder rather than smeared over the rest.
//
// The counters are not decoration. Milliseconds alone cannot tell a faster rasteriser
// from a frame that drew less, so the work the frame did is always shown next to the
// time it took, and fragments per millisecond — the number that actually moves when an
// optimisation lands — is spelled out at the bottom.

import { registerPanel } from './registry.js';
import { esc } from '../util.js';

// A stable colour per bucket, so the eye learns the shape of a frame and a change in it
// reads at a glance rather than being re-parsed each time.
const COLORS = [
  '#5ec8f8', // command decode
  '#f2b45c', // vertex + shader
  '#ff2ea6', // rasterise
  '#9b8cff', // texture decode
  '#59d99b', // gx transfers
  '#e0637a', // dsp
  '#6ec6b0', // svc + ipc
  '#7c8594', // the derived remainder — grey, because it is not a measurement
];

registerPanel({
  id: 'profile',
  title: 'Profile',
  slot: 'side',
  requires: 'profile',
  mount(body, ctx) {
    body.classList.add('mono');
    body.innerHTML = `<p class="muted" id="prof-empty">step a frame to see where its time went</p>
      <div id="prof-body" hidden>
        <div class="prof-bar" id="prof-bar"></div>
        <div id="prof-rows"></div>
        <div id="prof-counters"></div>
      </div>`;

    const empty = document.getElementById('prof-empty');
    const main = document.getElementById('prof-body');
    const bar = document.getElementById('prof-bar');
    const rows = document.getElementById('prof-rows');
    const counters = document.getElementById('prof-counters');
    const note = document.getElementById('profile-note');

    const show = (p) => {
      if (!p || !p.buckets || !p.buckets.length) return;
      empty.hidden = true;
      main.hidden = false;

      const total = p.totalMs || 0;
      note.textContent = `${total.toFixed(0)} ms/frame`;

      // One stacked bar: the shape of the frame, before any number is read.
      bar.innerHTML = p.buckets
        .map((b, i) => {
          const pct = total > 0 ? (100 * b.ms) / total : 0;
          if (pct <= 0) return '';
          return `<span class="prof-seg" style="width:${pct}%;background:${COLORS[i % COLORS.length]}" title="${esc(b.name)} ${b.ms.toFixed(1)} ms"></span>`;
        })
        .join('');

      rows.innerHTML = p.buckets
        .map((b, i) => {
          const pct = total > 0 ? (100 * b.ms) / total : 0;
          const count = b.count > 0 ? `×${b.count.toLocaleString()}` : '';
          return (
            `<div class="prow">` +
            `<span class="pkey" style="background:${COLORS[i % COLORS.length]}"></span>` +
            `<span class="pname">${esc(b.name)}</span>` +
            `<span class="pms">${b.ms.toFixed(1)} ms</span>` +
            `<span class="ppct">${pct.toFixed(0)}%</span>` +
            `<span class="pcount muted">${count}</span>` +
            `</div>`
          );
        })
        .join('');

      const c = {};
      for (const x of p.counters || []) c[x.name] = x.value;
      counters.innerHTML =
        (p.counters || [])
          .map(
            (x) =>
              `<div class="prow"><span class="pname muted">${esc(x.name)}</span>` +
              `<span class="grow"></span><span>${x.value.toLocaleString()}</span></div>`,
          )
          .join('') + rate(c, total);
    };

    // The rate line: the only number here that survives a change in what the frame drew.
    const rate = (c, total) => {
      const frags = (c['fragments drawn'] || 0) + (c['depth-killed'] || 0);
      if (!frags || !total) return '';
      return (
        `<div class="prow prate"><span class="pname">fragments / ms</span>` +
        `<span class="grow"></span><span>${(frags / total).toFixed(0)}</span></div>`
      );
    };

    ctx.conn.on('frame', (m) => show(m.profile));
    ctx.conn.on('render', (m) => show(m.profile)); // free-running frames keep it live
  },
});
