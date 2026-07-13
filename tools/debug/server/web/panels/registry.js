// The panel registry.
//
// A panel declares the capability it needs, and the shell mounts it only if the
// current target advertises that capability. So a platform whose rasteriser cannot
// report pixels simply has no overdraw panel, rather than an empty one that looks
// broken — and adding a panel is a registration, not an edit to the shell.

const panels = [];

// registerPanel adds a panel.
//
//   id       unique, also the DOM id of its body
//   title    the pane header
//   slot     "stage" | "side" | "bottom" — where in the dock it goes
//   requires a capability name, or "" for a panel every target can back
//   grow     take the spare vertical space in its slot
//   mount(body, ctx) -> void; ctx = { conn, store, ui }
export function registerPanel(p) {
  panels.push(p);
}

// mountPanels builds the dock for the capabilities this target has.
export function mountPanels(ctx) {
  const slots = {
    stage: document.getElementById('slot-stage'),
    side: document.getElementById('slot-side'),
    bottom: document.getElementById('slot-bottom'),
  };
  for (const el of Object.values(slots)) el.innerHTML = '';

  const mounted = [];
  for (const p of panels) {
    if (p.requires && !ctx.store.can(p.requires)) continue;

    const pane = document.createElement('section');
    pane.className = 'pane' + (p.grow ? ' grow-pane' : '');
    pane.dataset.panel = p.id;

    if (p.title) {
      const h = document.createElement('h2');
      h.textContent = p.title;
      const note = document.createElement('span');
      note.className = 'muted';
      note.id = `${p.id}-note`;
      h.appendChild(note);
      pane.appendChild(h);
    }

    const body = document.createElement('div');
    body.className = 'pane-body';
    body.id = `${p.id}-body`;
    pane.appendChild(body);

    slots[p.slot].appendChild(pane);
    p.mount(body, ctx);
    mounted.push(p.id);
  }
  return mounted;
}
