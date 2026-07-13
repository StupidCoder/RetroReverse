// The framebuffer viewport: the image on one canvas, and everything we draw *about*
// the image (the command highlight, the picked pixel) on a second one above it.
//
// The highlight is the point of the whole exercise. The page holds the frame's
// provenance buffer — one command index per pixel — so lighting up every pixel a
// command drew is a single local pass, with no request to the emulator at all.

const HIGHLIGHT = [255, 46, 166]; // --hot

export class Viewport {
  constructor() {
    this.view = document.getElementById('view');
    this.overlay = document.getElementById('overlay');
    this.wrap = document.getElementById('canvas-wrap');
    this.vctx = this.view.getContext('2d');
    this.octx = this.overlay.getContext('2d');

    this.w = 0;
    this.h = 0;
    this.zoom = 2;
    this.prov = null;
    this.selected = -1;
    this.pick = null;

    this.onHover = () => {};
    this.onPick = () => {};

    this.overlay.addEventListener('mousemove', (e) => {
      const p = this.pointAt(e);
      if (p) this.onHover(p.x, p.y);
    });
    this.overlay.addEventListener('mouseleave', () => this.onHover(-1, -1));
    this.overlay.addEventListener('click', (e) => {
      const p = this.pointAt(e);
      if (p) this.onPick(p.x, p.y);
    });
  }

  pointAt(e) {
    const r = this.overlay.getBoundingClientRect();
    const x = Math.floor((e.clientX - r.left) / this.zoom);
    const y = Math.floor((e.clientY - r.top) / this.zoom);
    if (x < 0 || y < 0 || x >= this.w || y >= this.h) return null;
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
  }

  setZoom(z) {
    this.zoom = z;
    this.applyZoom();
  }

  applyZoom() {
    const cssW = `${this.w * this.zoom}px`;
    const cssH = `${this.h * this.zoom}px`;
    for (const c of [this.view, this.overlay]) {
      c.style.width = cssW;
      c.style.height = cssH;
    }
    this.wrap.style.width = cssW;
    this.wrap.style.height = cssH;
  }

  drawImage(imageData) {
    this.resize(imageData.width, imageData.height);
    this.vctx.putImageData(imageData, 0, 0);
  }

  setProv(prov) {
    this.prov = prov;
    this.redrawOverlay();
  }

  // provAt is the whole reason the provenance buffer is shipped to the page: the
  // answer to "which command drew this pixel?" without a round trip.
  provAt(x, y) {
    if (!this.prov || x < 0 || y < 0 || x >= this.w || y >= this.h) return -1;
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

    if (this.prov && this.selected >= 0) {
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
  }
}
