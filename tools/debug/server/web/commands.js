// The command list. A frame runs to thousands of RDP commands, so the list is
// windowed: a spacer div gives the scrollbar its true height and only the visible
// rows exist in the DOM.

const ROW = 18; // px, must match .row height in style.css

export class CommandList {
  constructor(onSelect) {
    this.scroll = document.getElementById('cmd-scroll');
    this.spacer = document.getElementById('cmd-spacer');
    this.rows = document.getElementById('cmd-rows');
    this.count = document.getElementById('cmd-count');
    this.detail = document.getElementById('cmd-detail');
    this.onSelect = onSelect;

    this.cmds = [];
    this.selected = -1;

    this.scroll.addEventListener('scroll', () => this.render());
    this.rows.addEventListener('click', (e) => {
      const row = e.target.closest('.row');
      if (row) this.onSelect(Number(row.dataset.i));
    });
  }

  setCommands(cmds) {
    this.cmds = cmds;
    this.selected = -1;
    this.spacer.style.height = `${cmds.length * ROW}px`;
    this.count.textContent = `${cmds.length.toLocaleString()} commands`;
    this.scroll.scrollTop = 0;
    this.render();
    this.detail.innerHTML = '<p class="muted">select a command.</p>';
  }

  select(k, reveal = true) {
    this.selected = k;
    if (reveal) this.reveal(k);
    this.render();
    this.showDetail(k);
  }

  reveal(k) {
    const top = this.scroll.scrollTop;
    const bottom = top + this.scroll.clientHeight;
    const y = k * ROW;
    if (y < top || y + ROW > bottom) {
      this.scroll.scrollTop = Math.max(0, y - this.scroll.clientHeight / 2);
    }
  }

  render() {
    const total = this.cmds.length;
    const first = Math.max(0, Math.floor(this.scroll.scrollTop / ROW) - 2);
    const visible = Math.ceil(this.scroll.clientHeight / ROW) + 4;
    const last = Math.min(total, first + visible);

    this.rows.style.top = `${first * ROW}px`;
    let html = '';
    for (let i = first; i < last; i++) {
      const c = this.cmds[i];
      const sel = i === this.selected ? ' sel' : '';
      html += `<div class="row${sel}" data-i="${i}"><span class="idx">${i}</span><span class="nm">${esc(c.name)}</span></div>`;
    }
    this.rows.innerHTML = html;
  }

  showDetail(k) {
    const c = this.cmds[k];
    if (!c) {
      this.detail.innerHTML = '<p class="muted">select a command.</p>';
      return;
    }
    const words = c.words.map((w) => `<p>${w}</p>`).join('');
    const dec = c.decoded ? `<p>${esc(c.decoded)}</p>` : '';
    this.detail.innerHTML =
      `<p><span class="muted">#${k}</span> <b>${esc(c.name)}</b> <span class="muted">op ${hex2(c.op)}</span></p>` +
      dec + words;
  }
}

function hex2(n) {
  return '0x' + n.toString(16).padStart(2, '0');
}

function esc(s) {
  return String(s).replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c]));
}
