// The inspector panes: CPU registers, a memory hex view, and the overdraw history
// of the picked pixel.

export function renderCPU(m) {
  const el = document.getElementById('cpu');
  let rows = `<tr><td class="n">pc</td><td>${m.pc}</td>`;
  const extras = Object.entries(m.extra || {});
  rows += extras.length
    ? `<td class="n extra">${extras[0][0]}</td><td class="extra">${extras[0][1]}</td></tr>`
    : '<td></td><td></td></tr>';

  // Two register columns beside the extras, so the whole file fits without scrolling.
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
  el.innerHTML = `<table>${rows}</table>`;
}

export function renderMem(m) {
  const el = document.getElementById('mem');
  const bytes = [];
  for (let i = 0; i < m.data.length; i += 2) bytes.push(parseInt(m.data.slice(i, i + 2), 16));

  let html = '';
  for (let off = 0; off < bytes.length; off += 16) {
    const row = bytes.slice(off, off + 16);
    const hex = row.map((b) => b.toString(16).padStart(2, '0')).join(' ');
    const asc = row.map((b) => (b >= 32 && b < 127 ? String.fromCharCode(b) : '.')).join('');
    const addr = ((m.addr + off) >>> 0).toString(16).padStart(8, '0');
    html += `<div class="line"><span class="a">${addr}</span>${hex}<span class="asc">${esc(asc)}</span></div>`;
  }
  el.innerHTML = html || '<p class="muted">—</p>';
}

// renderOverdraw lists every write to the picked pixel in order — including the
// ones the rasteriser produced and then threw away on a depth or alpha test, which
// is usually the answer to "why is this pixel not the colour I expect?".
export function renderOverdraw(m, onSelectCmd) {
  const list = document.getElementById('overdraw-list');
  const picked = document.getElementById('picked');
  picked.textContent = `(${m.x}, ${m.y})`;

  if (!m.writes.length) {
    list.innerHTML =
      m.cmd < 0
        ? '<p class="muted">no command wrote this pixel.</p>'
        : `<p class="muted">last written by command ${m.cmd}. step with overdraw on to record the full history.</p>`;
    return;
  }

  const html = m.writes
    .map((w) => {
      const rgba = `rgba(${w.r},${w.g},${w.b},1)`;
      const tag = w.rejected ? '<span class="rej">rejected</span>' : 'drawn';
      const hex = [w.r, w.g, w.b, w.a].map((v) => v.toString(16).padStart(2, '0')).join('');
      return (
        `<div class="write clickable" data-cmd="${w.cmd}">` +
        `<span class="sw" style="background:${rgba}"></span>` +
        `<span class="muted">#${w.cmd}</span> ${esc(w.name)} ` +
        `<span class="muted">${hex}</span> ${tag}</div>`
      );
    })
    .join('');
  list.innerHTML = html;
  list.querySelectorAll('.write').forEach((el) => {
    el.addEventListener('click', () => onSelectCmd(Number(el.dataset.cmd)));
  });
}

function esc(s) {
  return String(s).replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c]));
}
