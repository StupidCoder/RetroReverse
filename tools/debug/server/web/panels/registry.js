// The panel registry.
//
// A panel declares the capability it needs, and the shell mounts it only if the current
// target advertises that capability. So a platform whose rasteriser cannot report pixels
// simply has no overdraw panel, rather than an empty one that looks broken — and adding a
// panel is a registration, not an edit to the shell.
//
// The registry builds the panes; the dock (dock.js) decides where they go and which one
// is visible.

import { buildDock } from './dock.js';

const panels = [];

// registerPanel adds a panel.
//
//   id       unique, also the prefix of its DOM ids ("<id>-body", "<id>-note")
//   title    the tab label, and the pane's own header when it is alone in a group
//   slot     "stage" | "side" | "bottom-left" | "bottom-right" — where it goes unless the
//            saved layout says otherwise
//   rank     where it sits among its group's tabs; the first one is the tab you land on.
//            Without it a panel would sit wherever its import happened to fall, which is
//            an ordering nobody chose.
//   requires a capability name, or "" for a panel every target can back
//   mount(body, ctx) -> void; ctx = { conn, store, ui, viewport }
export function registerPanel(p) {
  panels.push(p);
}

// byRank orders the panels within their groups. Array.prototype.sort is stable, so panels
// that declare no rank keep the order they were registered in.
const byRank = () => [...panels].sort((a, b) => (a.rank ?? 50) - (b.rank ?? 50));

let disposePrevious = null;

// mountPanels builds the panels this target's capabilities support and hands them to the
// dock. It runs again whenever the target changes, so the previous mount's subscriptions
// are disposed first — a handler left over from the last game would fire against DOM that
// is gone.
export function mountPanels(ctx) {
  if (disposePrevious) disposePrevious();
  ctx.conn.beginScope();
  ctx.store.beginScope();

  // The previous target's panes are still in the dock, and they carry the same element ids
  // as the ones about to be built — every panel names its nodes after itself. getElementById
  // returns the FIRST match in the document, so a panel mounting now would capture the OLD
  // target's node and write everything it ever renders into DOM the dock is about to throw
  // away: a savestate list nobody sees, and a viewport whose canvas is not on the page.
  //
  // So the old panes leave the document before the new ones are built. A pane belongs to
  // exactly one mount, and while a mount is in progress its ids must be the only ones.
  for (const pane of document.querySelectorAll('.pane')) pane.remove();

  // Panels look their own elements up with getElementById while mounting, so a pane has
  // to be in the document before it is mounted — even though the dock has not decided
  // where it goes yet. It is parked here first, then moved.
  const staging = document.getElementById('dock-staging');
  staging.innerHTML = '';

  const built = [];
  for (const p of byRank()) {
    if (p.requires && !ctx.store.can(p.requires)) continue;

    const pane = document.createElement('section');
    pane.className = 'pane';
    pane.dataset.panel = p.id;

    // Every panel gets a header. Its title is hidden when the panel is a tab (the tab
    // already says it), but the header stays, because panels hang controls off it — the
    // memory panel appends its address box, several write into the note span.
    const h = document.createElement('h2');
    const title = document.createElement('span');
    title.className = 'title';
    title.textContent = p.title || '';
    h.appendChild(title);
    const note = document.createElement('span');
    note.className = 'muted';
    note.id = `${p.id}-note`;
    h.appendChild(note);
    pane.appendChild(h);

    const body = document.createElement('div');
    body.className = 'pane-body';
    body.id = `${p.id}-body`;
    pane.appendChild(body);

    staging.appendChild(pane);
    p.mount(body, ctx);
    built.push({ id: p.id, pane, slot: p.slot, title: p.title || p.id });
  }

  const offConn = ctx.conn.endScope();
  const offStore = ctx.store.endScope();
  disposePrevious = () => {
    offConn();
    offStore();
  };

  buildDock(ctx, built);
  return built.map((b) => b.id);
}
