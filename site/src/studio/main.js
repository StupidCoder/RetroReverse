// Studio front-end: an immersive presentation of the extracted game assets. The active
// game's viewer (the same per-game viewer the standalone pages use) renders full-bleed into
// #stage; a floating control window picks the game and, cascading from it, the level.
//
// Each game exposes a slightly different viewer API, so a small adapter per game normalises
// "create a viewer", "list the levels (named)", and "show level i". Viewers are created lazily
// on first visit and cached (kept mounted but hidden) so switching back is instant.
//
// A global screen filter (screen.js) overlays all of them: a post-process that captures whichever
// viewer is active and re-renders it through a display-appropriate shader (CRT for the tube systems,
// a Game Boy / Game Gear LCD shader for the handhelds), picked from the game's system.

import { ScreenFilter } from './screen.js';
import { KeyboardCamera } from './camera.js';
import { INFO_TABS, infoHtml } from './info-content.js';

// The inbound deep-link state, parsed once at load: ?game=<id>&level=<slug|index> selects a
// game and one of its levels; ?objects=0/1 and ?filter=0/1 force those display flags; ?seed / ?debug
// are dev knobs. As the user navigates, syncUrl() writes this same shape back so every view is a
// copyable link. `level` accepts a stable per-level slug (preferred) or a numeric index; `asset`
// is the old index-only param name, still honoured.
const BOOT = new URLSearchParams(location.search);

// A URL-safe slug from a level's display name; used as its stable deep-link id.
const slugify = (s) => String(s).toLowerCase().normalize('NFKD')
  .replace(/[^\w]+/g, '-').replace(/^-+|-+$/g, '') || 'x';

const GAMES = [
  {
    id: 'sonic', name: 'Sonic the Hedgehog', system: 'Sega Game Gear',
    load: () => Promise.all([
      import('../shared/viewer.js'), import('../sonic/config.js'),
    ]).then(([m, c]) => class extends m.LevelViewer {
      constructor(el, hud) { super(el, hud, c.default); }
    }),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => (await v.init()).levels,
    show: (v, lvl, i) => v.loadLevel(lvl),
    // zone -> act accordion from the meta's section field ("Green Hills" -> "Act 1")
    group: (lvl) => ({
      section: lvl.section || lvl.name,
      label: lvl.name.startsWith(lvl.section) ? lvl.name.slice(lvl.section.length).trim() || lvl.name : lvl.name,
    }),
    layers: [
      { id: 'objects', label: 'Objects & enemies', default: true },
      { id: 'collision', label: 'Collision layer', default: false },
    ],
    music: async () => ['Green Hills:greenhills', 'Bridge:bridge', 'Jungle:jungle',
      'Labyrinth:labyrinth', 'Scrap Brain:scrapbrain', 'Sky Base:skybase', 'Special Stage:special']
      .map(s => { const [name, f] = s.split(':'); return { name, url: `public/sonic/music/${f}.mp3` }; }),
  },
  {
    id: 'fort', name: 'Fort Apocalypse', system: 'Commodore 64',
    load: () => Promise.all([
      import('../shared/viewer.js'), import('../fort/config.js'),
    ]).then(([m, c]) => class extends m.LevelViewer {
      constructor(el, hud) { super(el, hud, c.default); }
    }),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => (await v.init()).levels,
    show: (v, lvl, i) => v.loadLevel(lvl),
    layers: [{ id: 'objects', label: 'Objects & enemies', default: true }],
  },
  {
    id: 'turrican', name: 'Turrican', system: 'Amiga',
    load: () => Promise.all([
      import('../shared/viewer.js'), import('../turrican/config.js'),
    ]).then(([m, c]) => class extends m.LevelViewer {
      constructor(el, hud) { super(el, hud, c.default); }
    }),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => (await v.init()).levels,
    show: (v, lvl, i) => v.loadLevel(lvl),
    // world -> scene accordion from the meta's section field
    group: (lvl) => ({ section: lvl.section || lvl.name, label: lvl.name }),
    layers: [
      { id: 'objects', label: 'Objects & enemies', default: true },
      { id: 'collision', label: 'Collision layer', default: false },
    ],
    music: async () => (await fetch('public/turrican/music/manifest.json').then(r => r.json()))
      .map(m => ({ name: turricanTrackName(m), url: `public/turrican/music/${m.file}` })),
  },
  {
    id: 'marble', name: 'Marble Madness', system: 'Amiga',
    load: () => import('../marble/viewer.js').then(m => m.MarbleViewer),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => (await v.init()).levels,
    show: (v, lvl, i) => v.loadLevel(lvl),
    // per-view toggles: the scenery-overlay sprites live in the 2-D map, the
    // Track markers only in the slope view
    layers: [
      { id: 'objects', label: 'Scenery overlays', default: true, when: (m) => m.leaves?.[m.currentIdx]?.name === 'Map' },
      { id: 'markers', label: 'Markers', default: false, when: (m) => m.leaves?.[m.currentIdx]?.name === 'Slopes' },
    ],
    music: async () => (await fetch('public/marble/music/manifest.json').then(r => r.json()))
      .map(m => ({ name: m.course, url: `public/marble/music/${m.file}` })),
  },
  {
    id: 'sml', name: 'Super Mario Land', system: 'Nintendo Game Boy',
    load: () => Promise.all([
      import('../shared/viewer.js'), import('../sml/config.js'),
    ]).then(([m, c]) => class extends m.LevelViewer {
      constructor(el, hud) { super(el, hud, c.default); }
    }),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => (await v.init()).levels,
    show: (v, lvl, i) => v.loadLevel(lvl),
    layers: [
      { id: 'objects', label: 'Objects & enemies', default: true },
      { id: 'collision', label: 'Collision layer', default: false },
    ],
    // world -> level accordion from the meta's section field
    group: (lvl) => ({ section: lvl.section || lvl.name, label: lvl.name }),
    music: async () => [
      { name: 'Levels 1-1, 1-2, 3-1', url: 'public/sml/music/level-1-1.mp3' },
      { name: 'Levels 1-3, 3-2, 3-3', url: 'public/sml/music/level-1-3.mp3' },
      { name: 'Levels 2-1, 2-2 (Muda)', url: 'public/sml/music/level-2-1.mp3' },
      { name: 'Levels 4-1, 4-2 (Chai)', url: 'public/sml/music/level-4-1.mp3' },
      { name: 'Levels 2-3, 4-3 (boss)', url: 'public/sml/music/level-2-3.mp3' },
      { name: 'Bonus rooms', url: 'public/sml/music/bonus.mp3' },
    ],
  },
  {
    id: 'stuntcar', name: 'Stunt Car Racer', system: 'Amiga', render: '3d',
    load: () => import('../stuntcar/viewer.js').then(m => m.TrackViewer),
    make: (V, el, hud) => new V(el),
    list: async () => (await fetch('public/stuntcar/tracks.json').then(r => r.json())),
    show: (v, lvl, i) => v.show(lvl, i),
  },
  {
    id: 'elite', name: 'Elite', system: 'Commodore 64', render: '3d',
    load: () => import('../elite/viewer.js').then(m => m.ShipViewer),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => await v.init(), // returns the ship list
    show: (v, lvl, i) => v.loadShip(i),
    // open on the Cobra Mk III — the iconic player ship — rather than the missile
    defaultAsset: (ships) => ships.findIndex(s => s.name === 'Cobra Mk III'),
    // the docking music (The Blue Danube), rendered from the $BDDC engine through our SID emulator
    music: async () => [{ name: 'Docking — The Blue Danube', url: 'public/elite/music/docking.mp3' }],
  },
  {
    id: 'mariokart', name: 'Mario Kart DS', system: 'Nintendo DS', render: '3d',
    load: () => import('../mariokart/viewer.js').then(m => m.ModelViewer),
    make: (V, el, hud) => new V(el, hud),
    list: async (v) => await v.init(), // returns the model list from models.json
    show: (v, lvl, i) => v.loadModel(i),
    // Tracks are sections of their own: "Track" plus the course's map objects.
    // The manifest's short label names the item inside its section; the full
    // name (e.g. "Delfino Square · bridge") stays on the HUD and deep link.
    group: (lvl) => ({ section: lvl.section, label: lvl.label || lvl.name }),
    // open on Mario's B-Dasher rather than the first list item
    defaultAsset: (models) => models.findIndex(m => m.file === 'kart_MR_a.glb'),
    // course-only toggles: the "_V" backdrop (camera-locked skybox) and a fly-along
    // of the CPU racers' drive line. `when` keys off the current manifest entry so
    // they show only for tracks that ship those pieces.
    layers: [
      { id: 'skybox', label: 'Skybox', default: true, when: (m) => !!m.leaves?.[m.currentIdx]?.level?.skybox },
      { id: 'objects', label: 'Objects', default: true, when: (m) => !!m.leaves?.[m.currentIdx]?.level?.objects },
      { id: 'drive', label: 'Drive the CPU line', default: false, when: (m) => !!m.leaves?.[m.currentIdx]?.level?.path },
    ],
    // the cartridge's 76 SSEQ sequences, rendered through our SDAT sequencer+synth
    // (the retail SDAT ships no symbol block, so tracks are numbered, not named)
    music: async () => (await fetch('public/mariokart/music/tracks.json').then(r => r.json())),
  },
];

// Turrican's manifest labels worlds 0-based with hex start offsets; make them readable.
function turricanTrackName(m) {
  let l = String(m.label || 'Music').replace(/world (\d+)/i, (_, n) => 'World ' + (Number(n) + 1));
  l = l.charAt(0).toUpperCase() + l.slice(1);
  if (m.start && m.start !== '0') l += ` · $${m.start}`;
  return l;
}

const stage = document.getElementById('stage');
const hud = document.getElementById('hud');
// The HUD is two stacked lines: a caption the Studio owns (game · system — asset)
// and a detail line the active viewer owns (dimensions / zoom info). They no longer
// clobber each other, so the caption stays put while the viewer updates its detail.
const hudCaption = document.createElement('div'); hudCaption.className = 'hud-caption';
const hudDetail = document.createElement('div'); hudDetail.className = 'hud-detail';
hud.append(hudCaption, hudDetail);
const panel = document.getElementById('panel');
const gameList = document.getElementById('gameList');
const assetList = document.getElementById('assetList');
const assetLabel = document.getElementById('assetLabel');
const titlecard = document.getElementById('titlecard');
const spinner = document.getElementById('spinner');

const mounts = new Map(); // gameId -> { game, el, viewer, levels, current }
let activeId = null;
let busy = false;

// ---- the game list: systems (collapsible, one open at a time) -> games ----
const SYSTEMS = [
  { full: 'Amiga', short: 'Amiga' },
  { full: 'Commodore 64', short: 'C64' },
  { full: 'Sega Game Gear', short: 'Game Gear' },
  { full: 'Nintendo Game Boy', short: 'Game Boy' },
  { full: 'Nintendo DS', short: 'DS' },
];
const CHEVRON = '<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6"/></svg>';

function buildGameList() {
  gameList.innerHTML = '';
  for (const sys of SYSTEMS) {
    const games = GAMES.filter(g => g.system === sys.full);
    if (!games.length) continue;
    const group = document.createElement('div');
    group.className = 'sys-group';
    group.dataset.sys = sys.full;

    const header = document.createElement('button');
    header.className = 'item sys-header';
    header.innerHTML = `<span class="name">${sys.short}</span>${CHEVRON}`;
    header.addEventListener('click', () => toggleSystem(sys.full));
    group.appendChild(header);

    const sub = document.createElement('div');
    sub.className = 'sys-games';
    for (const game of games) {
      const b = document.createElement('button');
      b.className = 'item game-item';
      b.dataset.game = game.id;
      b.innerHTML = `<span class="name">${game.name}</span>`;
      b.addEventListener('click', () => selectGame(game.id));
      sub.appendChild(b);
    }
    group.appendChild(sub);
    gameList.appendChild(group);
  }
}

function openSystem(sysFull) {
  for (const group of gameList.querySelectorAll('.sys-group')) {
    group.classList.toggle('open', group.dataset.sys === sysFull);
  }
}

function toggleSystem(sysFull) {
  const group = gameList.querySelector(`.sys-group[data-sys="${sysFull}"]`);
  openSystem(group && group.classList.contains('open') ? null : sysFull);
}

function markActiveGame(id) {
  const game = GAMES.find(g => g.id === id);
  for (const b of gameList.querySelectorAll('.game-item')) b.classList.toggle('active', b.dataset.game === id);
  if (game) openSystem(game.system); // unfold the active game's system
}

// ---- asset list (per game). Most games are a flat list. Marble Madness is a two-level
//      accordion (course -> { Map, Slopes }); games with a `group(lvl,i)` adapter become a
//      world/zone -> act/scene accordion. Each leaf has { name, hud, run }. ----
function assetEntries(m) {
  const { game, viewer, levels } = m;
  if (game.id === 'marble') {
    return levels.map((course, ci) => ({
      name: course.name,
      children: [
        { name: 'Map', hud: `${course.name} · Map`, run: async () => { await game.show(viewer, course, ci); viewer.setMode('tilemap'); applyLayers(m); } },
        { name: 'Slopes', hud: `${course.name} · Slopes`, run: async () => { await game.show(viewer, course, ci); viewer.setMode('slopes'); applyLayers(m); } },
      ],
    }));
  }

  const leaf = (lvl, i, name) => ({
    name,
    hud: lvl.name || `Asset ${i + 1}`,
    level: lvl, // the manifest entry, so asset-specific layer toggles can inspect it
    run: async () => { await game.show(viewer, lvl, i); applyLayers(m); },
  });

  // grouped: fold the flat level list into section accordions via the adapter's group()
  if (game.group) {
    const groups = [];
    const byKey = new Map();
    levels.forEach((lvl, i) => {
      const { section, label } = game.group(lvl, i);
      let g = byKey.get(section);
      if (!g) { g = { name: section, children: [] }; byKey.set(section, g); groups.push(g); }
      g.children.push(leaf(lvl, i, label));
    });
    return groups;
  }

  return levels.map((lvl, i) => leaf(lvl, i, lvl.name || `Asset ${i + 1}`));
}

function addLeaf(m, leaf, parent) {
  const idx = m.leaves.length;
  // give the leaf a stable, readable deep-link slug (deduped within the game)
  const base = slugify(leaf.hud || leaf.name || `asset-${idx + 1}`);
  let slug = base, n = 2;
  while (m.slugSeen.has(slug)) slug = `${base}-${n++}`;
  m.slugSeen.add(slug);
  leaf.slug = slug;
  m.leaves.push(leaf);
  const b = document.createElement('button');
  b.className = 'item asset-item';
  b.dataset.idx = idx;
  b.innerHTML = `<span class="name">${leaf.name}</span>`;
  b.addEventListener('click', () => selectAsset(m, idx));
  parent.appendChild(b);
}

function buildAssetList(m) {
  assetLabel.style.display = ''; // reveal the Asset section once a game is chosen
  assetList.innerHTML = '';
  m.leaves = [];
  m.slugSeen = new Set();
  for (const entry of assetEntries(m)) {
    if (entry.children) {
      const group = document.createElement('div');
      group.className = 'asset-group';
      const header = document.createElement('button');
      header.className = 'item asset-header';
      header.innerHTML = `<span class="name">${entry.name}</span>${CHEVRON}`;
      header.addEventListener('click', () => toggleAssetGroup(group));
      group.appendChild(header);
      const sub = document.createElement('div');
      sub.className = 'asset-children';
      for (const leaf of entry.children) addLeaf(m, leaf, sub);
      group.appendChild(sub);
      assetList.appendChild(group);
    } else {
      addLeaf(m, entry, assetList);
    }
  }
}

function toggleAssetGroup(group) {
  const open = group.classList.contains('open');
  for (const g of assetList.querySelectorAll('.asset-group')) g.classList.toggle('open', !open && g === group);
}

function markActiveAsset(m) {
  for (const b of assetList.querySelectorAll('.asset-item')) b.classList.toggle('active', +b.dataset.idx === m.currentIdx);
  const active = assetList.querySelector(`.asset-item[data-idx="${m.currentIdx}"]`);
  const grp = active && active.closest('.asset-group');
  for (const g of assetList.querySelectorAll('.asset-group')) g.classList.toggle('open', g === grp);
  if (active) active.scrollIntoView({ block: 'nearest' });
}

// the leaf shown on a game's first visit: the adapter's defaultAsset (a fn of the level list,
// returning a level/leaf index — leaves are built in level order), else the first leaf.
function defaultLeaf(m) {
  const i = m.game.defaultAsset ? m.game.defaultAsset(m.levels) : 0;
  return (Number.isInteger(i) && i >= 0 && i < m.leaves.length) ? i : 0;
}

// run an asset (no busy management — used for the initial selection inside selectGame).
async function runAsset(m, idx) {
  const leaf = m.leaves[idx];
  if (!leaf) return;
  m.currentIdx = idx;
  m.currentName = leaf.hud;
  markActiveAsset(m);
  buildLayerToggles(m); // some toggles are asset-specific (e.g. Marble's Markers = Slopes only)
  await leaf.run();
}

async function selectAsset(m, idx) {
  if (busy || idx === m.currentIdx) return;
  setBusy(true);
  try {
    await runAsset(m, idx);
    updateHud(m);
  } catch (err) {
    console.error('studio: failed to show asset', idx, err);
  } finally {
    setBusy(false);
  }
}

function setBusy(on) {
  busy = on;
  spinner.classList.toggle('on', on);
}

// Pause a hidden viewer's render loop (and resume the shown one) so cached viewers don't
// keep rendering in the background. Pixi via the ticker; three.js loops gate on viewer.active.
function setMountActive(m, active) {
  if (!m) return;
  const v = m.viewer;
  v.active = active;
  if (v.app && typeof v.app.stop === 'function') v.app[active ? 'start' : 'stop']();
}

// ---- selection ----
async function selectGame(id) {
  if (busy || id === activeId) return;
  const game = GAMES.find(g => g.id === id);
  markActiveGame(id);
  setBusy(true);
  try {
    // hide and pause the currently mounted viewer
    if (activeId && mounts.has(activeId)) {
      const old = mounts.get(activeId);
      old.el.style.display = 'none';
      setMountActive(old, false);
    }

    let m = mounts.get(id);
    const firstMount = !m;
    if (!m) {
      const el = document.createElement('div');
      el.className = 'mount';
      el.dataset.render = game.render || '2d';
      stage.appendChild(el);
      const Viewer = await game.load();
      const viewer = game.make(Viewer, el, hudDetail);
      const levels = await game.list(viewer);
      m = { game, el, viewer, levels, currentIdx: 0, currentName: '' };
      mounts.set(id, m);
    }
    m.el.style.display = 'block';
    setMountActive(m, true);
    activeId = id;
    screen.setProfile(SYSTEM_PROFILE[game.system] || 'crt'); // pick CRT vs LCD shader by system
    buildScreenControls();
    buildLayerToggles(m);
    buildAssetList(m);
    if (firstMount) await runAsset(m, defaultLeaf(m)); // load the default asset on first visit
    else markActiveAsset(m);                           // returning to a cached viewer: keep its asset
    await loadGameMusic(game);
    updateHud(m);
    if (infoPanel.classList.contains('open')) renderInfo(); // refresh the details for the new game
    hideTitlecard();
  } catch (err) {
    console.error('studio: failed to load', id, err);
    hudCaption.innerHTML = `<b>${game.name}</b> — failed to load (${err.message})`;
  } finally {
    setBusy(false);
  }
}

function updateHud(m) {
  hudCaption.innerHTML = `<b>${m.game.name}</b> · ${m.game.system} &nbsp;—&nbsp; ${m.currentName}`;
  syncUrl();
}

// Resolve a deep-link `level` value (slug preferred, numeric index as fallback) to a leaf index,
// or -1 if it doesn't name a level of this game.
function resolveLevel(m, val) {
  if (val == null || val === '') return -1;
  const bySlug = m.leaves.findIndex(l => l.slug === val);
  if (bySlug >= 0) return bySlug;
  const n = parseInt(val, 10);
  return Number.isInteger(n) && n >= 0 && n < m.leaves.length ? n : -1;
}

// Mirror the current game/level and the objects/filter flags into the address bar so any view is a
// copyable deep link. replaceState (not push): we reflect state, we don't stack history entries.
// Only non-default flags are emitted, keeping links clean (absent objects/filter = the defaults).
function syncUrl() {
  const params = new URLSearchParams();
  const m = activeId && mounts.get(activeId);
  if (m) {
    params.set('game', m.game.id);
    const leaf = m.leaves && m.leaves[m.currentIdx];
    if (leaf && leaf.slug) params.set('level', leaf.slug);
    const objLayer = (m.game.layers || []).find(l => l.id === 'objects');
    if (objLayer) {
      const on = layerState(m).objects;
      if (on !== !!objLayer.default) params.set('objects', on ? '1' : '0');
    }
  }
  if (screen.ok && !screen.enabled) params.set('filter', '0'); // filter defaults on; only note the deviation
  for (const k of ['seed', 'debug']) { const v = BOOT.get(k); if (v != null) params.set(k, v); }
  const qs = params.toString();
  history.replaceState(null, '', qs ? '?' + qs : location.pathname);
}

function hideTitlecard() {
  if (titlecard.style.display === 'none') return;
  titlecard.classList.add('hidden');
  setTimeout(() => { titlecard.style.display = 'none'; }, 500); // fade out, then drop it for good
}

// ---- menu / panel ----
const menuBtn = document.getElementById('menuBtn');
function setMenu(open) {
  panel.classList.toggle('open', open);
  menuBtn.classList.toggle('hidden', open); // the panel grows out of the button; hide it while open
}
menuBtn.addEventListener('click', () => setMenu(true));
// Optional-chain the header wiring: this module is deferred and runs top-to-bottom,
// so a null element here (e.g. a stale cached index.html served without cache headers,
// from before the panelClose->panelBar rename) would throw and abort the rest of boot
// -- including buildGameList() far below, leaving an empty menu. Degrade instead.
document.getElementById('panelBar')?.addEventListener('click', () => setMenu(false));

// ---- info panel (technical details, tabbed) ----
// A second window, top-right, that fills the space the control window leaves. Its header is a
// fixed row of tabs (folded from the games' Markdown parts); the body scrolls. Content is keyed
// by the active game and the selected tab and re-rendered whenever either changes.
const infoBtn = document.getElementById('infoBtn');
const infoPanel = document.getElementById('info');
const infoTabsEl = document.getElementById('infoTabs');
const infoBody = document.getElementById('infoBody');
let infoTab = INFO_TABS[0].id;

function buildInfoTabs() {
  infoTabsEl.innerHTML = '';
  for (const t of INFO_TABS) {
    const b = document.createElement('button');
    b.className = 'info-tab' + (t.id === infoTab ? ' active' : '');
    b.dataset.tab = t.id;
    b.textContent = t.label;
    b.addEventListener('click', () => selectInfoTab(t.id));
    infoTabsEl.appendChild(b);
  }
}

function selectInfoTab(id) {
  infoTab = id;
  for (const b of infoTabsEl.querySelectorAll('.info-tab')) b.classList.toggle('active', b.dataset.tab === id);
  renderInfo();
}

function renderInfo() {
  const game = activeId && GAMES.find(g => g.id === activeId);
  if (!game) {
    infoBody.innerHTML = `<div class="info-doc info-empty">Pick a game from the menu to read its technical notes.</div>`;
    infoBody.scrollTop = 0;
    return;
  }
  const tab = INFO_TABS.find(t => t.id === infoTab);
  const html = infoHtml(game.id, infoTab);
  infoBody.innerHTML = `<div class="info-doc">` + (html ||
    `<div class="info-eyebrow">${game.name} · ${tab.label}</div>
     <p class="info-todo">This section hasn't been written yet.</p>`) + `</div>`;
  infoBody.scrollTop = 0;
}

function setInfo(open) {
  infoPanel.classList.toggle('open', open);
  infoBtn.classList.toggle('hidden', open); // the panel grows out of the button; hide it while open
  if (open) renderInfo();
}
infoBtn.addEventListener('click', () => setInfo(true));
// the bar also hosts the tab buttons -- clicking one of those switches tabs (its own
// listener above) rather than closing the panel; anywhere else in the bar closes it
document.getElementById('infoBar')?.addEventListener('click', (e) => {
  if (e.target.closest('.info-tab')) return;
  setInfo(false);
});
buildInfoTabs();

// ---- fullscreen ----
const fsBtn = document.getElementById('fsBtn');
const EXPAND = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V5a1 1 0 0 1 1-1h4M20 9V5a1 1 0 0 0-1-1h-4M4 15v4a1 1 0 0 0 1 1h4M20 15v4a1 1 0 0 1-1 1h-4"/></svg>';
const COLLAPSE = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 4v3a2 2 0 0 1-2 2H4M20 9h-3a2 2 0 0 1-2-2V4M4 15h3a2 2 0 0 1 2 2v3M15 20v-3a2 2 0 0 1 2-2h3"/></svg>';
fsBtn.innerHTML = EXPAND;
fsBtn.addEventListener('click', () => {
  if (document.fullscreenElement) document.exitFullscreen();
  else document.documentElement.requestFullscreen?.();
});
document.addEventListener('fullscreenchange', () => {
  const on = !!document.fullscreenElement;
  fsBtn.classList.toggle('on', on);
  fsBtn.innerHTML = on ? COLLAPSE : EXPAND;
});

// ---- per-game display layers (e.g. Fort's objects/enemies overlay) ----
// Not universal: each game adapter declares its own `layers` ([{id,label,default}]); games without
// one show no toggles. State is per game, persisted, and re-applied every time an asset (re)loads.
const displayLayers = document.getElementById('displayLayers');
const savedLayers = JSON.parse(localStorage.getItem('studio.layers') || '{}');

function layerState(m) {
  if (!m.layerState) {
    m.layerState = {};
    const saved = savedLayers[m.game.id] || {};
    for (const l of m.game.layers || []) m.layerState[l.id] = (l.id in saved) ? saved[l.id] : !!l.default;
    // a ?objects=0/1 deep link overrides the objects/enemies layer for the linked game on load
    if (m.game.id === BOOT.get('game') && BOOT.get('objects') != null && 'objects' in m.layerState) {
      m.layerState.objects = BOOT.get('objects') === '1';
    }
  }
  return m.layerState;
}

function applyLayers(m) {
  const st = layerState(m);
  // Optional-call setLayer: a viewer served stale (an older cached mariokart/viewer.js
  // against this newer main.js -- the site ships no cache headers) may predate the
  // layer support. Skip it rather than throw, which would reject the asset load and
  // leave the splash up. Its toggles just won't do anything until the cache refreshes.
  for (const l of m.game.layers || []) m.viewer.setLayer?.(l.id, st[l.id]);
}

function persistLayers(m) {
  savedLayers[m.game.id] = { ...layerState(m) };
  localStorage.setItem('studio.layers', JSON.stringify(savedLayers));
}

function buildLayerToggles(m) {
  // a layer may be asset-specific (l.when) -- only show toggles that apply to the current asset
  const layers = (m && m.game.layers || []).filter(l => !l.when || l.when(m));
  displayLayers.innerHTML = '';
  displayLayers.style.display = layers.length ? '' : 'none';
  if (!m) return;
  const st = layerState(m);
  for (const l of layers) {
    const label = document.createElement('label');
    label.className = 'switch';
    const input = document.createElement('input');
    input.type = 'checkbox';
    input.checked = st[l.id];
    const track = document.createElement('span');
    track.className = 'track';
    track.innerHTML = '<span class="knob"></span>';
    const text = document.createElement('span');
    text.className = 'switch-label';
    text.textContent = l.label;
    label.append(input, track, text);
    input.addEventListener('change', () => {
      st[l.id] = input.checked;
      m.viewer.setLayer?.(l.id, input.checked);
      persistLayers(m);
      syncUrl(); // reflect the objects/enemies (etc.) flag in the deep link
    });
    displayLayers.appendChild(label);
  }
}

// ---- Screen filter (global post-process over the active viewer; shader picked per game) ----
const screen = new ScreenFilter(stage);
screen.source = () => {
  const m = activeId && mounts.get(activeId);
  if (!m) return [];
  return [...m.el.querySelectorAll('canvas')].filter(c => getComputedStyle(c).display !== 'none');
};
// The active viewer's pixel grid { cell, ox, oy, ref } in the viewer's CSS px, so the shaders can
// lock their cell grid / scanlines / phosphor mask to the game's own pixels. A 2-D map camera gives
// one game pixel = `zoom` CSS px at the pan offset; a viewer may also expose its own pixelGrid()
// (e.g. low-res Elite's fixed C64 resolution). null when neither applies (a 3-D viewer at full res).
screen.pixelGrid = () => {
  const m = activeId && mounts.get(activeId);
  const v = m && m.viewer;
  if (!v) return null;
  const cam = v.cam;
  if (cam && cam.isMapCamera) {
    return { cell: cam.zoom, ox: cam.world.position.x, oy: cam.world.position.y, ref: cam.app.screen.width };
  }
  if (typeof v.pixelGrid === 'function') return v.pixelGrid();
  return null;
};

// Which shader profile each system uses (CRT tube vs handheld LCD).
const SYSTEM_PROFILE = {
  'Commodore 64': 'crt', 'Amiga': 'crt',
  'Sega Game Gear': 'gg', 'Nintendo Game Boy': 'gb',
};

// The sliders shown in the settings panel, per profile. The rows are (re)built in JS whenever the
// active profile changes, so each display type exposes its own knobs.
const PROFILE_CONTROLS = {
  crt: [
    { key: 'curvature', label: 'Curvature', min: 0, max: 10, step: 0.1 },
    { key: 'beamFocus', label: 'Scanline focus', min: 0.2, max: 1.0, step: 0.01 },
    { key: 'maskStrength', label: 'Mask intensity', min: 0, max: 1.0, step: 0.01 },
    { key: 'glow', label: 'Halation', min: 0, max: 1.0, step: 0.01 },
    { key: 'iqBlur', label: 'Chroma blur', min: 0, max: 5.0, step: 0.1 },
    { key: 'noise', label: 'Signal noise', min: 0, max: 0.5, step: 0.01 },
    { key: 'maskType', label: 'Mask type', min: 0, max: 1, step: 1, fmt: (v) => v < 0.5 ? 'Trinitron' : 'Shadow' },
    // Track pixels only applies to 2-D games (a map camera); hidden for the 3-D CRT games. When on,
    // the two line-count sliders are auto-driven from the on-screen pixels, so they're hidden.
    { key: 'trackPixels', label: 'Track pixels', min: 0, max: 1, step: 1, fmt: (v) => v < 0.5 ? 'Off' : 'On', rebuild: true, hidden: () => !screen.pixelGrid() },
    { key: 'signalLines', label: 'Signal lines', min: 60, max: 600, step: 2, int: true, hidden: (p) => p.trackPixels > 0.5 && !!screen.pixelGrid() },
    { key: 'scanLines', label: 'CRT scanlines', min: 60, max: 600, step: 2, int: true, hidden: (p) => p.trackPixels > 0.5 && !!screen.pixelGrid() },
  ],
  gb: [
    { key: 'pixelsPerCell', label: 'Dot size', min: 1, max: 4, step: 1, fmt: (v) => Math.round(v) + '× px' },
    { key: 'tint', label: 'Green tint', min: 0, max: 1, step: 0.01 },
    { key: 'gridStrength', label: 'Grid gap', min: 0, max: 0.8, step: 0.01 },
    { key: 'shadowOpacity', label: 'Pixel shadow', min: 0, max: 1, step: 0.01 },
    { key: 'shadowOffset', label: 'Shadow offset', min: 0, max: 1, step: 0.05, fmt: (v) => v.toFixed(2) + ' px' },
    { key: 'contrast', label: 'Contrast', min: 0.5, max: 2.5, step: 0.05 },
    { key: 'ghost', label: 'Ghosting', min: 0, max: 0.9, step: 0.01 },
  ],
  gg: [
    { key: 'pixelsPerCell', label: 'Dot size', min: 1, max: 4, step: 1, fmt: (v) => Math.round(v) + '× px' },
    { key: 'gridStrength', label: 'Grid gap', min: 0, max: 0.8, step: 0.01 },
    { key: 'subpixel', label: 'Subpixels', min: 0, max: 1, step: 0.01 },
    { key: 'saturation', label: 'Saturation', min: 0, max: 1.5, step: 0.01 },
    { key: 'glow', label: 'Backlight', min: 0, max: 0.6, step: 0.01 },
    { key: 'ghost', label: 'Ghosting', min: 0, max: 0.9, step: 0.01 },
  ],
};

const screenSliders = document.getElementById('screenSliders');
const fmtCtl = (c, v) => c.fmt ? c.fmt(v) : c.int ? String(Math.round(v)) : v.toFixed(2);

// Rebuild the slider rows for the active profile from its current param values.
function buildScreenControls() {
  const list = PROFILE_CONTROLS[screen.profile] || [];
  screenSliders.innerHTML = '';
  for (const c of list) {
    if (c.hidden && c.hidden(screen.params)) continue;
    const row = document.createElement('div'); row.className = 'ctl';
    const label = document.createElement('label');
    const val = document.createElement('span');
    label.append(document.createTextNode(c.label), val);
    const input = document.createElement('input');
    input.type = 'range'; input.min = c.min; input.max = c.max; input.step = c.step;
    input.value = screen.params[c.key];
    val.textContent = fmtCtl(c, screen.params[c.key]);
    input.addEventListener('input', () => {
      const v = parseFloat(input.value);
      screen.set(c.key, v);
      val.textContent = fmtCtl(c, v);
      persistScreen();
      if (c.rebuild) buildScreenControls(); // this control shows/hides others (e.g. Track pixels)
    });
    row.append(label, input);
    screenSliders.append(row);
  }
}

function wireScreen() {
  const saved = JSON.parse(localStorage.getItem('studio.screen') || '{}');
  if (saved.byProfile) {
    for (const prof of Object.keys(PROFILE_CONTROLS)) {
      const sp = saved.byProfile[prof];
      if (!sp) continue;
      for (const c of PROFILE_CONTROLS[prof]) {
        if (sp[c.key] !== undefined) screen.paramsByProfile[prof][c.key] = sp[c.key];
      }
    }
  }
  buildScreenControls();
  document.getElementById('screenReset').addEventListener('click', () => {
    screen.reset();
    buildScreenControls();
    persistScreen();
  });
  const toggle = document.getElementById('screenToggle');
  const controls = document.getElementById('screenControls');
  const gear = document.getElementById('screenSettings');
  toggle.addEventListener('change', () => { screen.setEnabled(toggle.checked); persistScreen(); syncUrl(); });
  gear.addEventListener('click', () => gear.classList.toggle('on', controls.classList.toggle('shown')));

  // enabled by default; only an explicit prior "off" (or ?filter=0, legacy ?crt=0) turns it off.
  const forced = BOOT.get('filter') ?? BOOT.get('crt');
  const startOn = (forced === '1' || (forced !== '0' && saved.enabled !== false)) && screen.ok;
  toggle.checked = startOn;
  if (!screen.ok) toggle.disabled = true;
  if (startOn) screen.setEnabled(true);
  // the settings panel stays folded until the gear is clicked
}

// Persist the on/off state and every profile's slider values independently.
function persistScreen() {
  const byProfile = {};
  for (const prof of Object.keys(PROFILE_CONTROLS)) {
    byProfile[prof] = {};
    for (const c of PROFILE_CONTROLS[prof]) byProfile[prof][c.key] = screen.paramsByProfile[prof][c.key];
  }
  localStorage.setItem('studio.screen', JSON.stringify({ enabled: screen.enabled, byProfile }));
}

// ---- Music player (per-game song list + transport). Switching games stops the music;
//      switching levels within a game leaves it playing. ----
const audio = new Audio();
audio.preload = 'none';
let musicTracks = [];
let playingUrl = null;
let repeat = localStorage.getItem('studio.repeat') !== '0'; // loop the current song; on by default
audio.loop = repeat;
const musicLabel = document.getElementById('musicLabel');
const musicListEl = document.getElementById('musicList');
const transport = document.getElementById('musicTransport');
const musPlay = document.getElementById('musPlay');
const musSeek = document.getElementById('musSeek');
const musTime = document.getElementById('musTime');
const PLAY_SVG = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 5v14l12-7z"/></svg>';
const PAUSE_SVG = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 5h3v14H7zM14 5h3v14h-3z"/></svg>';
musPlay.innerHTML = PLAY_SVG;

const fmtTime = (s) => { s = Math.max(0, s | 0); return Math.floor(s / 60) + ':' + String(s % 60).padStart(2, '0'); };

function renderMusicList() {
  musicListEl.innerHTML = '';
  musicTracks.forEach((t, i) => {
    const b = document.createElement('button');
    b.className = 'item' + (t.url === playingUrl ? ' active' : '');
    b.innerHTML = `<span class="name">${t.name}</span>`;
    b.addEventListener('click', () => playTrack(i));
    musicListEl.appendChild(b);
  });
}

function updateMusicUI() {
  const hasMusic = musicTracks.length > 0;
  const loaded = !!playingUrl;
  const show = hasMusic || loaded;
  musicLabel.style.display = show ? '' : 'none';
  musicListEl.style.display = hasMusic ? '' : 'none';
  transport.style.display = show ? 'flex' : 'none';
}

function playTrack(i) {
  const t = musicTracks[i];
  if (!t) return;
  if (playingUrl !== t.url) { audio.src = t.url; playingUrl = t.url; }
  audio.play().catch(() => {});
  renderMusicList();
  updateMusicUI();
}

function skip(d) {
  if (!musicTracks.length) return;
  let idx = musicTracks.findIndex(t => t.url === playingUrl);
  if (idx < 0) idx = d > 0 ? -1 : 0;
  playTrack((idx + d + musicTracks.length) % musicTracks.length);
}

musPlay.addEventListener('click', () => {
  if (!playingUrl) { if (musicTracks.length) playTrack(0); return; }
  if (audio.paused) audio.play().catch(() => {}); else audio.pause();
});
document.getElementById('musPrev').addEventListener('click', () => skip(-1));
document.getElementById('musNext').addEventListener('click', () => skip(1));
const musRepeat = document.getElementById('musRepeat');
function syncRepeat() {
  audio.loop = repeat;
  musRepeat.classList.toggle('on', repeat);
  musRepeat.title = repeat ? 'Repeat: on' : 'Repeat: off';
}
musRepeat.addEventListener('click', () => {
  repeat = !repeat;
  localStorage.setItem('studio.repeat', repeat ? '1' : '0');
  syncRepeat();
});
syncRepeat();
audio.addEventListener('play', () => { musPlay.innerHTML = PAUSE_SVG; musPlay.title = 'Pause'; });
audio.addEventListener('pause', () => { musPlay.innerHTML = PLAY_SVG; musPlay.title = 'Play'; });
audio.addEventListener('ended', () => skip(1));
audio.addEventListener('timeupdate', () => {
  if (audio.duration) musSeek.value = String(Math.round(audio.currentTime / audio.duration * 1000));
  musTime.textContent = fmtTime(audio.currentTime);
});
musSeek.addEventListener('input', () => { if (audio.duration) audio.currentTime = (musSeek.value / 1000) * audio.duration; });

function stopMusic() {
  audio.pause();
  try { audio.currentTime = 0; } catch {}
  playingUrl = null;
  musPlay.innerHTML = PLAY_SVG; musPlay.title = 'Play';
  musSeek.value = '0';
  musTime.textContent = '0:00';
}

async function loadGameMusic(game) {
  stopMusic(); // switching games stops the music
  try { musicTracks = game.music ? await game.music() : []; }
  catch (e) { console.error('studio: music load failed', game.id, e); musicTracks = []; }
  renderMusicList();
  updateMusicUI();
}

// ---- keyboard camera: cursor keys scroll (with momentum), +/- zoom the active viewer ----
new KeyboardCamera(() => {
  const m = activeId && mounts.get(activeId);
  return m ? m.viewer : null;
});

// ---- boot ----
buildGameList();
wireScreen();
displayLayers.style.display = 'none'; // per-game display toggles appear once a game is picked
assetLabel.style.display = 'none';    // the Asset + Music sections stay hidden until a game
updateMusicUI();                      // is picked (the splash is up meanwhile)
setMenu(true);                        // start with the control window open (discoverable)
// Keep the title card up until the user picks a game -- unless a ?game= deep link asks for one
// (e.g. ?game=sonic&level=green-hills-act-1), in which case load it straight. `level` takes a
// stable slug or a numeric index; `asset` is the legacy index-only alias.
const startGame = GAMES.find(g => g.id === BOOT.get('game'));
// ?debug=1 exposes the mount table for the headless screenshot driver; ?seed=N makes
// randomized object placement (Fort) reproducible (consumed by the shared layers code).
if (BOOT.get('debug')) window.__studio = { mounts, get activeId() { return activeId; } };
window.__studioSeed = BOOT.get('seed') ? parseInt(BOOT.get('seed'), 10) : null;
if (startGame) {
  selectGame(startGame.id).then(() => {
    const m = mounts.get(startGame.id);
    if (!m) return;
    const idx = resolveLevel(m, BOOT.get('level') ?? BOOT.get('asset'));
    if (idx >= 0) selectAsset(m, idx); // no-op if it's already the default-loaded leaf
  });
}
