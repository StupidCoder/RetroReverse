// xrlayout.js — the pure half of the in-headset panel: geometry, hit rects and
// canvas drawing, with NO three.js and no DOM beyond a 2-D context handed in.
//
// It is split out for one reason: this is the only part of the XR menu that can
// be checked without a headset, and the headset has no console. layout() is a
// pure function of (model, size) returning a flat, non-overlapping rect list;
// draw() renders from that same list, so anything visible is pointable.

// Canvas-pixel geometry. One place, so layout() and draw() cannot disagree.
export const M = {
  W: 1024,
  // Tall enough that the games column holds the whole catalogue without paging
  // — 21 games today, and the list only grows. Paging the thing you navigate
  // WITH is the one place a page button is pure friction.
  H: 800,          // the main panel; the status strip is separate
  STATUS_H: 34,
  PAD: 20,
  HEADER_H: 58,
  TABS_H: 46,
  FOOTER_H: 52,
  ROW_H: 32,
  COL_HEAD_H: 30,
  PAGER_H: 32,
  COL_GAP: 14,
  BAR_H: 64,       // the summon tab's own canvas
  BAR_W: 320,
};

export const C = {
  panel: 'rgba(17,22,32,0.94)',
  panel2: '#1c2331',
  line: '#2a3446',
  text: '#d7dfeb',
  dim: '#8a97ab',
  accent: '#6ea8ff',
  bad: '#ff8a8a',
  sel: 'rgba(110,168,255,0.18)',
};

const FONT = 'system-ui, -apple-system, "Segoe UI", sans-serif';

// ---- layout ------------------------------------------------------------------
//
// model = {
//   chrome: 'panel' | 'bar',
//   title, subtitle,
//   tabs: [{ id, label }], tab: <id>,
//   columns: [{ id, width, header, page, items: [item] }],
//   footer: [{ id, label, disabled, on }],
//   closeId,
// }
// item  = { id, kind: 'row'|'group'|'toggle'|'radio'|'button'|'blank',
//           label, sub, on, selected, disabled, indent }
//
// Returns { rects, cols, body } where rects is a flat, NON-OVERLAPPING list in
// canvas pixels. Every drawn, hittable thing is a rect; draw() renders from the
// same list, so anything you can see you can point at.
export function layout(model, size = M) {
  const { W, H, PAD } = size;
  const rects = [];
  const push = (r) => { rects.push(r); return r; };

  if (model.chrome === 'bar') {
    const items = model.footer || [];
    const n = Math.max(1, items.length);
    const w = (size.BAR_W - PAD * 2) / n;
    items.forEach((it, i) => push({
      id: it.id, kind: 'button', item: it,
      x: PAD + i * w, y: 8, w, h: size.BAR_H - 16,
    }));
    return { rects, cols: [], body: null, bar: true };
  }

  let y = 0;
  // header — title left, close right
  if (model.closeId) {
    push({
      id: model.closeId, kind: 'button', item: { label: '✕' },
      x: W - PAD - 40, y: 10, w: 40, h: 38,
    });
  }
  y += size.HEADER_H;

  // tabs
  const tabs = model.tabs || [];
  if (tabs.length) {
    let tx = PAD;
    for (const t of tabs) {
      const w = 22 + textWidth(t.label, 17);
      push({ id: t.id, kind: 'tab', item: t, x: tx, y: y + 6, w, h: size.TABS_H - 12 });
      tx += w + 8;
    }
    y += size.TABS_H;
  }

  const bodyY = y + 6;
  const bodyH = H - bodyY - size.FOOTER_H;
  const body = { x: PAD, y: bodyY, w: W - PAD * 2, h: bodyH };

  // columns share the body width by weight
  const columns = model.columns || [];
  const total = columns.reduce((s, c) => s + (c.width || 1), 0) || 1;
  const gaps = size.COL_GAP * Math.max(0, columns.length - 1);
  const cols = [];
  let cx = body.x;
  for (const col of columns) {
    const w = ((body.w - gaps) * (col.width || 1)) / total;
    const items = col.items || [];

    // Two passes: the pager only exists if the items do not fit without it,
    // and reserving its height can itself push them over. Measure, then commit.
    // A column may set its own row height: the games list is one short title
    // per row and packs tighter than an asset tree with its indents and counts.
    const rowH = col.rowH || size.ROW_H;
    const headH = col.header ? size.COL_HEAD_H : 0;
    const availAll = body.h - headH;
    const fitAll = Math.max(1, Math.floor(availAll / rowH));
    const paged = items.length > fitAll;
    const rows = paged
      ? Math.max(1, Math.floor((availAll - size.PAGER_H) / rowH))
      : fitAll;
    const pages = Math.max(1, Math.ceil(items.length / rows));
    const page = Math.min(Math.max(0, col.page || 0), pages - 1);
    const visible = items.slice(page * rows, page * rows + rows);

    let iy = body.y + headH;
    visible.forEach((it) => {
      if (it.kind !== 'blank') {
        push({ id: it.id, kind: it.kind || 'row', item: it, col: col.id, x: cx, y: iy, w, h: rowH });
      }
      iy += rowH;
    });

    if (paged) {
      // Big pinch targets, not precise ones. A hand ray is worth about half a
      // degree, so an A-Z index strip would be ~18 mm cells at arm's length —
      // pointable, but not reliably. Jumping ten pages at a time crosses the
      // deepest list in the catalogue (451 doors in the mansion, 30 pages) in
      // three pinches, with targets nobody can miss.
      const py = body.y + body.h - size.PAGER_H + 2;
      const bw = 52, ph = size.PAGER_H - 4;
      const far = pages > 5;
      const btn = (id, label, disabled, x) => push({
        id: `${col.id}:${id}`, kind: 'pager', item: { label, disabled }, col: col.id, x, y: py, w: bw, h: ph,
      });
      btn('page-', '◀', page === 0, cx);
      if (far) btn('page--', '◀◀', page === 0, cx + bw + 6);
      btn('page+', '▶', page >= pages - 1, cx + w - bw);
      if (far) btn('page++', '▶▶', page >= pages - 1, cx + w - bw * 2 - 6);
    }

    cols.push({ ...col, x: cx, y: body.y, w, h: body.h, rowH, rows, pages, page, visible, paged, headH });
    cx += w + size.COL_GAP;
  }

  // footer buttons, right-aligned
  const foot = model.footer || [];
  if (foot.length) {
    const fy = H - size.FOOTER_H + 8;
    let fx = W - PAD;
    for (let i = foot.length - 1; i >= 0; i--) {
      const it = foot[i];
      const w = 28 + textWidth(it.label, 17);
      fx -= w;
      push({ id: it.id, kind: 'button', item: it, x: fx, y: fy, w, h: size.FOOTER_H - 16 });
      fx -= 10;
    }
  }

  return { rects, cols, body, bar: false };
}

// A rough advance-width model so layout() stays pure — no canvas, no DOM, so
// it runs in node. Only used to SIZE tab and button chrome, never to clip text
// (draw() measures for real), so a few pixels of error cost nothing.
function textWidth(s, px) {
  return Math.round(String(s ?? '').length * px * 0.56);
}

export function hitAt(lay, x, y) {
  for (const r of lay.rects) {
    if (x >= r.x && x < r.x + r.w && y >= r.y && y < r.y + r.h) return r;
  }
  return null;
}

// ---- drawing -----------------------------------------------------------------

export function draw(ctx, model, lay, size = M) {
  const W = lay.bar ? size.BAR_W : size.W;
  const H = lay.bar ? size.BAR_H : size.H;
  ctx.clearRect(0, 0, W, H);
  roundRect(ctx, 1, 1, W - 2, H - 2, lay.bar ? 16 : 18);
  ctx.fillStyle = C.panel;
  ctx.fill();
  ctx.strokeStyle = C.line;
  ctx.lineWidth = 2;
  ctx.stroke();

  if (lay.bar) {
    for (const r of lay.rects) drawButton(ctx, r, 19);
    return;
  }

  ctx.textBaseline = 'middle';
  // header
  ctx.fillStyle = C.text;
  ctx.font = `600 24px ${FONT}`;
  const titleMax = size.W - size.PAD * 2 - 60;
  ctx.textAlign = 'left';
  ctx.fillText(clip(ctx, model.title || '', titleMax), size.PAD, 30);
  if (model.subtitle) {
    ctx.fillStyle = C.dim;
    ctx.font = `400 15px ${FONT}`;
    ctx.fillText(clip(ctx, model.subtitle, titleMax), size.PAD, 50);
  }

  // column headers + separators
  for (const col of lay.cols) {
    if (col.header) {
      ctx.fillStyle = C.dim;
      ctx.font = `600 13px ${FONT}`;
      ctx.textAlign = 'left';
      ctx.fillText(col.header.toUpperCase(), col.x + 6, col.y + 14);
      ctx.strokeStyle = C.line;
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(col.x, col.y + col.headH - 6);
      ctx.lineTo(col.x + col.w, col.y + col.headH - 6);
      ctx.stroke();
    }
    if (col.paged) {
      ctx.fillStyle = C.dim;
      ctx.font = `400 13px ${FONT}`;
      ctx.textAlign = 'center';
      ctx.fillText(`${col.page + 1} / ${col.pages}`, col.x + col.w / 2, col.y + col.h - 12);
    }
  }

  for (const r of lay.rects) {
    switch (r.kind) {
      case 'tab': drawTab(ctx, r, r.id === model.tab); break;
      case 'button': case 'pager': drawButton(ctx, r, 17); break;
      case 'group': drawGroup(ctx, r); break;
      case 'toggle': case 'radio': drawCheck(ctx, r); break;
      default: drawRow(ctx, r);
    }
  }
}

function drawRow(ctx, r) {
  const it = r.item;
  if (it.selected) {
    roundRect(ctx, r.x, r.y + 2, r.w, r.h - 4, 7);
    ctx.fillStyle = C.sel;
    ctx.fill();
  }
  const ind = (it.indent || 0) * 16;
  ctx.textAlign = 'left';
  ctx.font = `${it.selected ? 600 : 400} 18px ${FONT}`;
  ctx.fillStyle = it.disabled ? C.dim : (it.selected ? C.accent : C.text);
  let maxW = r.w - 16 - ind;
  if (it.sub) {
    ctx.font = `400 14px ${FONT}`;
    const sw = ctx.measureText(it.sub).width;
    ctx.fillStyle = C.dim;
    ctx.textAlign = 'right';
    ctx.fillText(it.sub, r.x + r.w - 8, r.y + r.h / 2);
    maxW -= sw + 12;
    ctx.textAlign = 'left';
    ctx.font = `${it.selected ? 600 : 400} 18px ${FONT}`;
    ctx.fillStyle = it.disabled ? C.dim : (it.selected ? C.accent : C.text);
  }
  ctx.fillText(clip(ctx, it.label, maxW), r.x + 8 + ind, r.y + r.h / 2);
}

function drawGroup(ctx, r) {
  ctx.textAlign = 'left';
  ctx.font = `600 13px ${FONT}`;
  ctx.fillStyle = C.dim;
  const ind = (r.item.indent || 0) * 16;
  ctx.fillText(clip(ctx, String(r.item.label).toUpperCase(), r.w - 16 - ind), r.x + 8 + ind, r.y + r.h / 2 + 2);
}

function drawCheck(ctx, r) {
  const it = r.item;
  const cy = r.y + r.h / 2, bx = r.x + 8, s = 18;
  if (r.kind === 'radio') {
    ctx.beginPath();
    ctx.arc(bx + s / 2, cy, s / 2 - 1, 0, Math.PI * 2);
  } else {
    roundRect(ctx, bx, cy - s / 2, s, s, 5);
  }
  ctx.fillStyle = it.on ? C.accent : 'transparent';
  ctx.strokeStyle = it.on ? C.accent : C.line;
  ctx.lineWidth = 2;
  if (it.on) ctx.fill();
  ctx.stroke();
  if (it.on) {
    ctx.strokeStyle = '#0b0e14';
    ctx.lineWidth = 2.5;
    ctx.beginPath();
    ctx.moveTo(bx + 4, cy);
    ctx.lineTo(bx + 7.5, cy + 4);
    ctx.lineTo(bx + 14, cy - 4.5);
    ctx.stroke();
  }
  ctx.textAlign = 'left';
  ctx.font = `400 18px ${FONT}`;
  ctx.fillStyle = it.disabled ? C.dim : C.text;
  ctx.fillText(clip(ctx, it.label, r.w - s - 26), bx + s + 10, cy);
}

function drawTab(ctx, r, on) {
  roundRect(ctx, r.x, r.y, r.w, r.h, 8);
  ctx.fillStyle = on ? C.panel2 : 'transparent';
  ctx.fill();
  ctx.textAlign = 'center';
  ctx.font = `${on ? 600 : 400} 17px ${FONT}`;
  ctx.fillStyle = on ? C.accent : C.dim;
  ctx.fillText(r.item.label, r.x + r.w / 2, r.y + r.h / 2);
}

function drawButton(ctx, r, px) {
  const it = r.item;
  roundRect(ctx, r.x, r.y, r.w, r.h, 8);
  ctx.fillStyle = it.on ? C.panel2 : 'rgba(255,255,255,0.04)';
  ctx.fill();
  ctx.strokeStyle = C.line;
  ctx.lineWidth = 1;
  ctx.stroke();
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.font = `500 ${px}px ${FONT}`;
  ctx.fillStyle = it.disabled ? C.line : (it.on ? C.accent : C.text);
  ctx.fillText(clip(ctx, it.label, r.w - 12), r.x + r.w / 2, r.y + r.h / 2);
}

export function drawStatus(ctx, text, size = M) {
  const W = size.W, H = size.STATUS_H;
  ctx.clearRect(0, 0, W, H);
  roundRect(ctx, 1, 1, W - 2, H - 2, 10);
  ctx.fillStyle = 'rgba(11,14,20,0.9)';
  ctx.fill();
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  ctx.font = `400 15px ui-monospace, SFMono-Regular, Menlo, monospace`;
  ctx.fillStyle = /error|fail|cannot|unavailable/i.test(text || '') ? C.bad : C.dim;
  ctx.fillText(clip(ctx, text || '', W - 24), 12, H / 2);
}

function clip(ctx, s, maxW) {
  s = String(s ?? '');
  if (ctx.measureText(s).width <= maxW) return s;
  let lo = 0, hi = s.length;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (ctx.measureText(`${s.slice(0, mid)}…`).width <= maxW) lo = mid;
    else hi = mid - 1;
  }
  return `${s.slice(0, lo)}…`;
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}
