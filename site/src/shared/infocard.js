// Shared object-info card for the level viewers (SM64DS, Ultima Underworld, …):
// a small panel pinned to the bottom-right of the viewer mount, shown when you
// click a placed object. It names the thing and lists whatever detail the viewer
// has — a description, traced behaviour notes, or the object's in-game words —
// and fades itself after a delay. Keeping one implementation means every game's
// click card looks and behaves the same.
//
// mount must be a positioned element (the studio's .mount is position:absolute).
export class InfoCard {
  constructor(mount) {
    this.mount = mount;
    this.box = null;
    this.timer = null;
  }

  // Fields (all optional except title): title (yellow header), subtitle (bold
  // secondary line), body (normal text; muted:true greys it), quote (an italic
  // quoted block, e.g. a sign's words). timeout ms until auto-hide (0 = never).
  show({ title, subtitle, body, muted, quote, timeout = 15000 }) {
    this.hide();
    const d = document.createElement('div');
    d.style.cssText = 'position:absolute;right:12px;bottom:64px;max-width:min(480px,70%);' +
      'background:rgba(10,13,18,.94);border:1px solid #3a4a5c;border-radius:8px;' +
      'padding:10px 12px;font:12px/1.55 system-ui;color:#dfe6f0;z-index:5';
    if (title) {
      const h = document.createElement('div');
      h.style.cssText = 'font-weight:600;margin-bottom:4px;color:#ffd75e';
      h.textContent = title;
      d.append(h);
    }
    if (subtitle) {
      const s = document.createElement('div');
      s.style.cssText = 'font-weight:600;margin-bottom:2px';
      s.textContent = subtitle;
      d.append(s);
    }
    if (body) {
      const b = document.createElement('div');
      if (muted) b.style.cssText = 'color:#9aa7b8';
      b.textContent = body;
      d.append(b);
    }
    if (quote) {
      const t = document.createElement('div');
      t.style.cssText = 'margin-top:8px;padding:8px 10px;background:rgba(255,255,255,.06);' +
        'border-left:3px solid #ffd75e;border-radius:4px;white-space:pre-wrap;' +
        'font-style:italic;color:#f2ecd8;max-height:180px;overflow-y:auto';
      t.textContent = quote;
      d.append(t);
    }
    this.mount.appendChild(d);
    this.box = d;
    clearTimeout(this.timer);
    if (timeout) this.timer = setTimeout(() => this.hide(), timeout);
  }

  hide() {
    if (this.box) { this.box.remove(); this.box = null; }
    clearTimeout(this.timer);
  }
}
