// app.js — the Studio shell. 100% data-driven: the only registry here is
// "category → view module". Routing lives in the URL hash
// (#/<game>/<asset>?variant=..&seed=..) so the browser's back button walks
// level → object → picture naturally.

import { loadIndex, CATEGORY_LABELS } from './data.js';
import { ScreenFilter } from './screenfx.js';

const $ = (id) => document.getElementById(id);
const landing = $('landing'), stage = $('stage');

const VIEWS = {
  level: () => import('./viewlevel.js'),
  object: () => import('./viewobject.js'),
  music: () => import('./viewmedia.js'),
  sfx: () => import('./viewmedia.js'),
  picture: () => import('./viewmedia.js'),
  video: () => import('./viewmedia.js'),
};

let games = [];
let current = null;   // { game, asset, view, params }
let screenFx = null;
let filterOn = false; // the user's screen-filter choice survives navigation

// ---- routing ---------------------------------------------------------------

function parseHash() {
  const h = location.hash.replace(/^#\/?/, '');
  if (!h) return null;
  const [path, query] = h.split('?');
  const [gameId, assetId] = path.split('/').map(decodeURIComponent);
  const params = new URLSearchParams(query || '');
  return { gameId, assetId, params };
}

export function navigate(gameId, assetId, params = {}) {
  let h = `#/${encodeURIComponent(gameId)}`;
  if (assetId) h += `/${encodeURIComponent(assetId)}`;
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) if (v != null && v !== '') q.set(k, v);
  const qs = q.toString();
  if (qs) h += `?${qs}`;
  if (location.hash === h) return;
  location.hash = h; // hashchange drives route(); history entries come free
}

async function route() {
  const r = parseHash();
  if (!r) return showLanding();
  const game = games.find((g) => g.id === r.gameId);
  if (!game) return showLanding();
  let asset = r.assetId ? game.asset(r.assetId) : null;
  if (!asset) asset = defaultAsset(game);
  await showAsset(game, asset, r.params);
}

function defaultAsset(game) {
  return game.assets('level')[0] || game.man.assets[0];
}

// ---- landing ----------------------------------------------------------------

function showLanding() {
  teardown();
  stage.hidden = true;
  landing.hidden = false;
  $('gameTitle').textContent = '';
  for (const id of ['catSelect', 'assetSelect', 'variantSelect']) $(id).hidden = true;
  for (const id of ['displayBtn', 'wireBtn', 'sunBtn', 'filterBtn', 'infoBtn']) $(id).hidden = true;

  landing.innerHTML = '';
  const head = document.createElement('div');
  head.className = 'landing-head';
  head.innerHTML = `<h1>RetroReverse</h1>
    <p>Games taken apart from their original ROM and disc images and rebuilt entirely
    from the binaries — maps, objects, models, cutscenes and music, no emulation.
    Everything below is served as <a href="https://github.com/StupidCoder/RetroReverse/blob/main/RETROX.md" style="color:var(--accent)">Retro-X</a> data.</p>`;
  landing.appendChild(head);

  const grid = document.createElement('div');
  grid.className = 'cards';
  for (const g of games) {
    const c = document.createElement('div');
    c.className = 'card';
    const meta = [g.man.platform, g.man.year].filter(Boolean).join(' · ');
    const logo = g.man.logo ? `<img src="${g.url('', g.man.logo)}" alt="">` : '';
    c.innerHTML = `<h2>${logo}${esc(g.man.title)}</h2>
      <div class="meta">${esc(meta)}</div><p>${esc(g.man.description || '')}</p>`;
    c.onclick = () => navigate(g.id, defaultAsset(g).id);
    grid.appendChild(c);
  }
  landing.appendChild(grid);
  document.title = 'RetroReverse — Studio';
  requestAnimationFrame(updateTopbarFades);
}

// ---- asset views -------------------------------------------------------------

function teardown() {
  if (current?.view?.unmount) current.view.unmount();
  current = null;
  stage.innerHTML = '';
  stage.className = '';
  if (screenFx) {
    stage.appendChild(screenFx.canvas);
    screenFx.setEnabled(false);
    // Unbind the dead view: its capture closures survive here otherwise, and
    // reading a destroyed pixi Application's canvas throws in the filter loop.
    screenFx.source = () => [];
    screenFx.pixelGrid = () => null;
  }
  $('displayPanel').hidden = true;
  $('displayPanel').innerHTML = '';
}

let mountSeq = 0;
async function showAsset(game, asset, params) {
  const seq = ++mountSeq;
  teardown();
  landing.hidden = true;
  stage.hidden = false;
  document.title = `${asset.name} — ${game.man.title}`;

  buildTopbar(game, asset, params);
  spinner(true);
  try {
    const mod = await VIEWS[asset.category]();
    const view = await mod.mount({
      stage, game, asset, params,
      navigate: (assetId, p) => navigate(game.id, assetId, p),
      setVariants: (variants, activeId, apply) => buildVariantSelect(game, asset, variants, activeId, params, apply),
      displayPanel: displayPanelAPI(),
      toast,
    });
    if (seq !== mountSeq) {
      // A newer navigation superseded this mount while it was loading:
      // tear the late arrival down instead of leaking its listeners.
      view.unmount?.();
      return;
    }
    current = { game, asset, view, params };
    wireViewButtons(game, view);
  } catch (e) {
    if (seq !== mountSeq) return;
    console.error(e);
    toast(`Failed to load ${asset.name}: ${e.message}`);
  } finally {
    if (seq === mountSeq) spinner(false);
  }
}

// ---- top bar ------------------------------------------------------------------

function buildTopbar(game, asset, params) {
  $('gameTitle').textContent = game.man.title;

  const cat = $('catSelect');
  cat.hidden = false;
  cat.innerHTML = '';
  for (const c of game.categories()) {
    const o = document.createElement('option');
    o.value = c;
    o.textContent = CATEGORY_LABELS[c] || c;
    if (c === asset.category) o.selected = true;
    cat.appendChild(o);
  }
  cat.onchange = () => navigate(game.id, game.assets(cat.value)[0].id);

  // The music view is a player with its own track list, so a second track
  // dropdown in the top bar is redundant — hide it there.
  const sel = $('assetSelect');
  sel.hidden = asset.category === 'music';
  sel.innerHTML = '';
  let group = null, groupEl = null;
  for (const a of game.assets(asset.category)) {
    let parent = sel;
    if (a.group) {
      if (a.group !== group) {
        groupEl = document.createElement('optgroup');
        groupEl.label = a.group;
        sel.appendChild(groupEl);
        group = a.group;
      }
      parent = groupEl;
    }
    const o = document.createElement('option');
    o.value = a.id;
    o.textContent = a.name;
    if (a.id === asset.id) o.selected = true;
    parent.appendChild(o);
  }
  sel.onchange = () => navigate(game.id, sel.value);

  $('variantSelect').hidden = true;

  // info modal is available whenever there is something to say
  const hasInfo = !!(asset.description || game.man.docs?.length);
  $('infoBtn').hidden = false;
  $('infoBtn').onclick = () => showInfoModal(game, asset);
  $('infoBtn').classList.toggle('on', false);
  if (!hasInfo && !statsProvider) { /* keep the button: stats still show */ }

  // screen filter: offered when the game declares a profile; the on/off
  // choice is the user's and persists across levels, objects and games
  const filter = game.display.filter;
  const fb = $('filterBtn');
  fb.hidden = !filter || !screenFx?.ok;
  if (filter && screenFx?.ok) {
    screenFx.setProfile(filter);
    screenFx.setEnabled(filterOn);
    fb.classList.toggle('on', filterOn);
    fb.onclick = () => {
      filterOn = !filterOn;
      screenFx.setEnabled(filterOn);
      fb.classList.toggle('on', filterOn);
      // 3D views drop to the game's native resolution under the filter
      current?.view?.setNative?.(filterOn ? game.display.native : null);
    };
  }

  $('wireBtn').hidden = true;
  $('sunBtn').hidden = true;
  $('displayBtn').hidden = true;
  requestAnimationFrame(updateTopbarFades);
}

function buildVariantSelect(game, asset, variants, activeId, params, apply) {
  const vs = $('variantSelect');
  if (!variants || variants.length < 2) { vs.hidden = true; return; }
  vs.hidden = false;
  vs.innerHTML = '';
  for (const v of variants) {
    const o = document.createElement('option');
    o.value = v.id;
    o.textContent = v.name || v.id;
    if (v.id === activeId) o.selected = true;
    vs.appendChild(o);
  }
  vs.onchange = () => {
    const p = Object.fromEntries(new URLSearchParams(location.hash.split('?')[1] || ''));
    p.variant = vs.value;
    // Apply in place when the view supports it: a remount would reset the
    // camera, and comparing variants wants the viewpoint kept. The view
    // passes its setter directly (usable even while the mount is still
    // streaming rooms); current.view is the fallback.
    const set = apply || current?.view?.setVariant;
    if (set) {
      set(vs.value);
      const q = new URLSearchParams();
      for (const [k, v] of Object.entries(p)) if (v != null && v !== '') q.set(k, v);
      history.replaceState(null, '', `#/${encodeURIComponent(game.id)}/${encodeURIComponent(asset.id)}?${q}`);
      return;
    }
    navigate(game.id, asset.id, p);
  };
}

let statsProvider = null;
let sunOn = true; // the sun toggle defaults on and persists across asset/model switches
function wireViewButtons(game, view) {
  statsProvider = view.stats || null;

  const wb = $('wireBtn');
  if (view.setWireframe) {
    wb.hidden = false;
    let on = false;
    wb.onclick = () => { on = !on; view.setWireframe(on); wb.classList.toggle('on', on); };
  }

  const sb = $('sunBtn');
  if (view.setSunlight) {
    sb.hidden = false;
    if (sunOn) view.setSunlight(true);
    sb.classList.toggle('on', sunOn);
    sb.onclick = () => { sunOn = !sunOn; view.setSunlight(sunOn); sb.classList.toggle('on', sunOn); };
  }

  const db = $('displayBtn');
  const panel = $('displayPanel');
  if (panel.childElementCount > 0) {
    db.hidden = false;
    db.onclick = () => { panel.hidden = !panel.hidden; };
  }

  if (screenFx) {
    screenFx.source = () => (view.sources ? view.sources() : []);
    screenFx.pixelGrid = () => (view.pixelGrid ? view.pixelGrid() : null);
    // a view mounted while the filter is on starts in native-resolution mode
    if (filterOn && game.display.filter && screenFx.ok) view.setNative?.(game.display.native);
  }
}

function displayPanelAPI() {
  const panel = $('displayPanel');
  return {
    section(title) {
      const h = document.createElement('h4');
      h.textContent = title;
      panel.appendChild(h);
    },
    toggle(label, on, cb) {
      const l = document.createElement('label');
      const i = document.createElement('input');
      i.type = 'checkbox';
      i.checked = on;
      i.onchange = () => cb(i.checked);
      l.append(i, document.createTextNode(label));
      panel.appendChild(l);
      return i;
    },
    radio(name, label, on, cb) {
      const l = document.createElement('label');
      const i = document.createElement('input');
      i.type = 'radio';
      i.name = name;
      i.checked = on;
      i.onchange = () => i.checked && cb();
      l.append(i, document.createTextNode(label));
      panel.appendChild(l);
      return i;
    },
  };
}

// ---- info modal -----------------------------------------------------------------

async function showInfoModal(game, asset) {
  const body = $('modalBody');
  body.innerHTML = '<div class="modal-tabs" id="modalTabs" hidden></div><div class="modal-tab-body" id="modalTabBody"></div>';
  const tabsEl = $('modalTabs');
  const content = $('modalTabBody');

  let assetHtml = `<h2>${esc(asset.name)}</h2><div class="dim">${esc(game.man.title)} · ${esc(CATEGORY_LABELS[asset.category] || asset.category)}</div>`;
  if (asset.description) assetHtml += `<p>${esc(asset.description)}</p>`;
  if (statsProvider) {
    const stats = await statsProvider();
    if (stats && Object.keys(stats).length) {
      assetHtml += '<h3>Stats</h3><table>';
      for (const [k, v] of Object.entries(stats)) assetHtml += `<tr><td>${esc(k)}</td><td>${esc(String(v))}</td></tr>`;
      assetHtml += '</table>';
    }
  }

  // Sections: the asset itself first, then the game's doc pages.
  const sections = [{ id: 'asset', title: 'This asset', load: async () => assetHtml }];
  for (const d of game.man.docs || []) {
    sections.push({
      id: d.id, title: d.title,
      load: () => fetch(game.url('', d.file)).then((r) => r.text(), () => ''),
    });
  }

  const cache = new Map();
  const btns = new Map();
  const select = async (id) => {
    const sec = sections.find((x) => x.id === id) || sections[0];
    if (!cache.has(sec.id)) cache.set(sec.id, sec.load());
    content.innerHTML = await cache.get(sec.id);
    content.scrollTop = 0;
    for (const [bid, b] of btns) b.classList.toggle('on', bid === sec.id);
  };
  if (sections.length > 1) {
    tabsEl.hidden = false;
    for (const sec of sections) {
      const b = document.createElement('button');
      b.className = 'mtab';
      b.textContent = sec.title;
      b.onclick = () => select(sec.id);
      tabsEl.appendChild(b);
      btns.set(sec.id, b);
    }
    // same sideways-scroll affordance as the main top bar
    tabsEl.addEventListener('scroll', () => updateFades(tabsEl));
    requestAnimationFrame(() => updateFades(tabsEl));
  }
  await select('asset');
  $('modalBack').hidden = false;
}

// ---- chrome ----------------------------------------------------------------------

// The top bar scrolls sideways on narrow screens; fade classes signal
// truncated content on either edge.
function updateTopbarFades() {
  updateFades($('topbar'));
}

// updateFades sets the sideways-scroll fade cues on any horizontally
// scrolling bar (the top bar, the info modal's section tabs).
function updateFades(el) {
  const max = el.scrollWidth - el.clientWidth;
  el.classList.toggle('fade-r', max > 1 && el.scrollLeft < max - 1);
  el.classList.toggle('fade-l', max > 1 && el.scrollLeft > 1);
}

function spinner(on) {
  let s = stage.querySelector('.spinner');
  if (on && !s) { s = document.createElement('div'); s.className = 'spinner'; stage.appendChild(s); }
  if (!on && s) s.remove();
}

let toastTimer = 0;
function toast(msg) {
  const t = $('toast');
  t.textContent = msg;
  t.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.hidden = true; }, 3500);
}

function esc(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

// ---- boot ------------------------------------------------------------------------

$('homeBtn').onclick = () => { location.hash = ''; };
$('fsBtn').onclick = () => {
  if (document.fullscreenElement) document.exitFullscreen();
  else document.documentElement.requestFullscreen?.();
};
$('modalClose').onclick = () => { $('modalBack').hidden = true; };
$('modalBack').onclick = (e) => { if (e.target === $('modalBack')) $('modalBack').hidden = true; };
document.addEventListener('click', (e) => {
  const p = $('displayPanel');
  if (!p.hidden && !p.contains(e.target) && e.target !== $('displayBtn') && !$('displayBtn').contains(e.target)) {
    p.hidden = true;
  }
});
window.addEventListener('hashchange', route);
$('topbar').addEventListener('scroll', updateTopbarFades);
window.addEventListener('resize', updateTopbarFades);

(async () => {
  try {
    games = await loadIndex();
  } catch (e) {
    console.error(e);
    landing.innerHTML = `<div class="landing-head"><h1>RetroReverse</h1><p>Could not load public/index.json: ${esc(e.message)}</p></div>`;
    return;
  }
  screenFx = new ScreenFilter(stage);
  route();
})();
