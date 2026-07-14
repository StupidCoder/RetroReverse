// The dock: four groups, each holding several panels and showing one.
//
// The debugger has eleven panels now, and stacking them all in a few fixed regions left
// nothing legible. So a group is a tab bar plus its panels, only one of them visible —
// and any panel can be *promoted* to the stage, the big region on the left, swapping
// places with whatever was there.
//
// The trick that makes promotion cheap: a panel's DOM node is moved, not rebuilt. Its
// listeners, its canvas contents and its subscriptions all come with it, so a promoted
// viewport keeps the frame it was showing and the pixel it had picked. Nothing in any
// panel had to change to make this work.

// The stage is the big region; the rest are tab groups. The two bottom groups are
// deliberately separate rather than one wide strip of tabs: the left one is about the
// MACHINE (its states, its CPU, its memory, its files) and the right one about the FRAME
// on the stage (what overdrew what, what the selected command was, what it all cost). Side
// by side, you can watch one while reading the other.
const GROUPS = ['stage', 'side', 'bottom-left', 'bottom-right'];
const TABBED = GROUPS.filter((g) => g !== 'stage');

// panes maps a panel id to its <section>, homes to the group it belongs to by default,
// and titles to what its tab says. All three are rebuilt whenever the target changes.
let panes = new Map();
let homes = new Map();
let titles = new Map();
let layout = null;
let platform = '';

// layoutKey remembers an arrangement per platform: the N64 and the PSX have different
// panels and you want different habits on each. The version is part of the key, so a
// layout saved against a different set of groups is not reconciled, it is simply not
// found — the old one would have been read as "the user put everything on the left".
const layoutKey = (p) => `framedbg.layout.v2.${p || 'unknown'}`;

const emptyGroups = () => Object.fromEntries(TABBED.map((g) => [g, []]));

// defaultLayout puts every panel in the group it declared, with the viewport on the stage.
function defaultLayout(ids) {
  const l = { stage: '', groups: emptyGroups(), active: {} };
  for (const id of ids) {
    const home = homes.get(id);
    if (home === 'stage' && !l.stage) l.stage = id;
    else if (home === 'stage') l.groups.side.push(id); // a second stage panel: park it
    else l.groups[home].push(id);
  }
  if (!l.stage) l.stage = ids[0] || '';
  for (const g of TABBED) l.active[g] = l.groups[g][0] || '';
  return l;
}

// reconcile makes a remembered layout fit the panels this target actually has: drop what
// it no longer offers, append what the layout never heard of. A stale layout can then
// never hide a panel, nor resurrect one that is gone.
function reconcile(saved, ids) {
  const have = new Set(ids);
  const l = {
    stage: have.has(saved.stage) ? saved.stage : '',
    groups: emptyGroups(),
    active: {},
  };
  const placed = new Set();
  if (l.stage) placed.add(l.stage);

  for (const g of TABBED) {
    for (const id of saved.groups?.[g] || []) {
      if (have.has(id) && !placed.has(id)) {
        l.groups[g].push(id);
        placed.add(id);
      }
    }
  }
  for (const id of ids) {
    if (placed.has(id)) continue;
    const home = homes.get(id);
    if (home === 'stage' && !l.stage) l.stage = id;
    else l.groups[home === 'stage' ? 'side' : home].push(id);
    placed.add(id);
  }
  // The stage must never be empty: something has to hold the big region.
  if (!l.stage) {
    for (const g of TABBED) {
      if (l.groups[g].length) {
        l.stage = l.groups[g].shift();
        break;
      }
    }
  }
  for (const g of TABBED) {
    const a = saved.active?.[g];
    l.active[g] = l.groups[g].includes(a) ? a : l.groups[g][0] || '';
  }
  return l;
}

function save() {
  try {
    localStorage.setItem(layoutKey(platform), JSON.stringify(layout));
  } catch {
    // A browser with storage disabled still gets a working dock, just a forgetful one.
  }
}

function load(ids) {
  try {
    const raw = localStorage.getItem(layoutKey(platform));
    if (raw) return reconcile(JSON.parse(raw), ids);
  } catch {
    // A corrupt layout is not worth a broken page.
  }
  return defaultLayout(ids);
}

// buildDock arranges the panes that registry.js has mounted.
export function buildDock(ctx, built) {
  panes = new Map(built.map((b) => [b.id, b.pane]));
  homes = new Map(built.map((b) => [b.id, b.slot]));
  titles = new Map(built.map((b) => [b.id, b.title]));
  platform = ctx.store.get('platform');
  maximized = null;
  layout = load(built.map((b) => b.id));
  render();
}

export function resetLayout() {
  layout = defaultLayout([...panes.keys()]);
  save();
  render();
}

// reveal brings a panel forward — its tab if it is in a group, nothing if it is already
// the stage. The overdraw panel calls this when you click a pixel, so the write history
// arrives in front of you rather than behind a tab you are not looking at.
export function reveal(id) {
  if (!layout || !panes.has(id) || layout.stage === id) return;
  for (const g of TABBED) {
    if (layout.groups[g].includes(id)) {
      layout.active[g] = id;
      save();
      render();
      return;
    }
  }
}

// promote swaps a panel with whatever is on the stage.
function promote(id) {
  const g = groupOf(id);
  if (!g || g === 'stage') return;
  const old = layout.stage;

  layout.groups[g] = layout.groups[g].filter((x) => x !== id);
  layout.stage = id;
  if (old) {
    // The displaced panel takes the promoted one's place, and becomes the group's
    // active tab — so it is visible, not buried.
    layout.groups[g].push(old);
    layout.active[g] = old;
  }
  save();
  render();
}

function groupOf(id) {
  if (layout.stage === id) return 'stage';
  for (const g of TABBED) {
    if (layout.groups[g].includes(id)) return g;
  }
  return '';
}

let maximized = null;

export function toggleMaximize(id) {
  const target = id || activeOf(focusGroup) || layout?.stage;
  if (!target) return;
  maximized = maximized === target ? null : target;
  render();
}

export function unmaximize() {
  if (maximized) {
    maximized = null;
    render();
  }
}

// focusGroup is the group whose tab was last clicked — what F maximizes if you do not
// name a panel.
let focusGroup = 'stage';

function activeOf(g) {
  if (!layout) return '';
  return g === 'stage' ? layout.stage : layout.active[g];
}

// render lays the panes out. It only moves nodes and toggles classes: no panel is rebuilt,
// so none of them loses its state.
function render() {
  if (!layout) return;

  // A maximized group covers the dock but not the toolbar — you still want Step and Play.
  // The toolbar wraps, so its height is not a constant.
  const bar = document.getElementById('toolbar');
  document.documentElement.style.setProperty('--dock-top', `${bar.offsetHeight + 6}px`);

  for (const g of GROUPS) {
    const el = document.getElementById(`slot-${g}`);
    el.innerHTML = '';

    const ids = g === 'stage' ? (layout.stage ? [layout.stage] : []) : layout.groups[g];

    // A target with few panels can leave a group with nothing in it. An empty bordered
    // box is worse than no box.
    el.classList.toggle('empty', ids.length === 0);
    el.classList.toggle('maximized', maximized != null && groupOf(maximized) === g);
    if (!ids.length) continue;

    const active = maximized && groupOf(maximized) === g ? maximized : activeOf(g);

    // A tab bar for one panel is noise: a lone panel keeps its own header and no tabs.
    if (ids.length > 1) el.appendChild(tabBar(g, ids, active));

    for (const id of ids) {
      const pane = panes.get(id);
      pane.classList.toggle('active', id === active);
      pane.classList.toggle('tabbed', ids.length > 1);
      el.appendChild(pane);
    }
  }

  // With no side group there is no second column to leave a gap for.
  document.querySelector('main').classList.toggle('no-side', !layout.groups.side.length);
}

function tabBar(g, ids, active) {
  const bar = document.createElement('div');
  bar.className = 'tabs';

  for (const id of ids) {
    const tab = document.createElement('button');
    tab.className = 'tab' + (id === active ? ' sel' : '');
    tab.title = 'double-click to maximize';

    const label = document.createElement('span');
    label.textContent = titleOf(id);
    tab.appendChild(label);

    // Promote: give this panel the big region, and send the stage's panel here.
    if (g !== 'stage') {
      const up = document.createElement('span');
      up.className = 'promote';
      up.textContent = '⬚';
      up.title = 'promote to the main panel';
      up.onclick = (e) => {
        e.stopPropagation();
        promote(id);
      };
      tab.appendChild(up);
    }

    tab.onclick = () => {
      focusGroup = g;
      if (g !== 'stage') layout.active[g] = id;
      maximized = null;
      save();
      render();
    };
    tab.ondblclick = () => {
      focusGroup = g;
      toggleMaximize(id);
    };
    bar.appendChild(tab);
  }
  return bar;
}

function titleOf(id) {
  return titles.get(id) || id;
}
