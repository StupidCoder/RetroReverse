// xrbrowser.js — what the in-headset panel actually says.
//
// Two tabs over one panel:
//
//   Browse    every game and every asset, so the headset never has to come off
//             to change what you are looking at.
//   Controls  whatever the mounted view offers — layers, areas, variants,
//             animation clips, wireframe, sun. The views already push these
//             through ctx.displayPanel and ctx.setVariants, so the shell hands
//             them ITS OWN implementation of that interface and the controls
//             arrive here without a single view knowing about XR.
//
// The catalogue is 3444 assets across 21 games, and one group in the mansion
// holds 451 doors on its own. A flat list would be 44 pages of pinching, so the
// asset column is a DRILL-DOWN: categories collapsed, then groups, then assets.
// Only what you opened is on screen.
//
// This module owns no three.js and no DOM — it turns state into a layout model
// and turns a clicked rect back into an intent. That is what makes it testable
// without a headset.

import { CATEGORY_LABELS } from './data.js';

const PAGE_JUMP = 10;

export class Browser {
  // { games, canShow(asset) -> bool, onPick(gameId, assetId), onAction(id) }
  constructor(opts) {
    this.games = opts.games || [];
    this.canShow = opts.canShow || (() => true);
    this.onPick = opts.onPick || (() => {});
    this.onAction = opts.onAction || (() => {});
    this.tab = 'browse';
    this.gameId = null;   // the game whose assets are LISTED (not necessarily loaded)
    this.loadedGame = null;
    this.loadedAsset = null;
    this.expanded = new Set();
    this.page = { games: 0, assets: 0, controls: 0 };
    this.controls = [];
    this.busy = false;
  }

  get game() { return this.games.find((g) => g.id === this.gameId) || null; }

  // Called when the shell has actually mounted something: open the tree to it
  // and page to it, so the panel always shows you where you are.
  setLoaded(gameId, assetId) {
    this.loadedGame = gameId;
    this.loadedAsset = assetId;
    if (this.gameId !== gameId) {
      this.gameId = gameId;
      this.expanded.clear();
      this.page.assets = 0;
    }
    const g = this.game, a = g?.asset(assetId);
    if (a) {
      this.expanded.add(`cat:${a.category}`);
      if (a.group) this.expanded.add(`grp:${a.category}|${a.group}`);
      this._pageTo(`asset:${a.id}`);
    }
    this._pageTo(`game:${gameId}`, 'games');
  }

  setControls(list) { this.controls = list || []; }

  // ---- the model ---------------------------------------------------------------

  model() {
    const g = this.game;
    const cols = this.tab === 'browse' ? this._browseCols() : this._controlCols();
    const foot = [];
    if (this.tab === 'browse') {
      const steps = this._steps();
      const i = steps.findIndex((a) => a.id === this.loadedAsset);
      foot.push({ id: 'step-', label: '◀ Prev', disabled: this.busy || i <= 0 });
      foot.push({ id: 'step+', label: 'Next ▶', disabled: this.busy || i < 0 || i >= steps.length - 1 });
    }
    foot.push({ id: 'recenter', label: 'Recentre' });
    return {
      chrome: 'panel',
      title: g ? g.man.title : 'RetroReverse',
      subtitle: this._subtitle(),
      tabs: [{ id: 'browse', label: 'Browse' }, { id: 'controls', label: `Controls${this.controls.length ? '' : ' —'}` }],
      tab: this.tab,
      closeId: 'close',
      columns: cols,
      footer: foot,
    };
  }

  _subtitle() {
    const g = this.game;
    if (!g) return '';
    const meta = [g.man.platform, g.man.year].filter(Boolean).join(' · ');
    const a = g.id === this.loadedGame ? g.asset(this.loadedAsset) : null;
    return a ? `${meta} — ${a.name || a.id}` : meta;
  }

  _browseCols() {
    return [
      {
        id: 'games',
        width: 0.33,
        header: 'Games',
        rowH: 28,
        page: this.page.games,
        items: this.games.map((g) => ({
          id: `game:${g.id}`,
          kind: 'row',
          label: g.man.title,
          sub: g.id === this.loadedGame ? '●' : '',
          selected: g.id === this.gameId,
        })),
      },
      { id: 'assets', width: 0.67, header: 'Assets', page: this.page.assets, items: this._assetItems() },
    ];
  }

  // The drill-down. A category row, then — only if it is open — its groups, then
  // — only if THEY are open — the assets. Counts go on the right so a closed row
  // still says how much is behind it.
  _assetItems() {
    const g = this.game;
    if (!g) return [];
    const items = [];
    for (const cat of g.categories()) {
      const list = g.assets(cat);
      const catKey = `cat:${cat}`;
      const catOpen = this.expanded.has(catKey);
      items.push({
        id: catKey, kind: 'row', label: `${catOpen ? '▾' : '▸'} ${CATEGORY_LABELS[cat] || cat}`,
        sub: String(list.length), selected: catOpen,
      });
      if (!catOpen) continue;

      // Preserve manifest order for both groups and their members.
      const groups = [];
      const byName = new Map();
      for (const a of list) {
        const key = a.group || '';
        if (!byName.has(key)) { byName.set(key, []); groups.push(key); }
        byName.get(key).push(a);
      }
      for (const name of groups) {
        const members = byName.get(name);
        if (!name) { for (const a of members) items.push(this._assetRow(a, 1)); continue; }
        const gk = `grp:${cat}|${name}`;
        const open = this.expanded.has(gk);
        items.push({
          id: gk, kind: 'row', label: `${open ? '▾' : '▸'} ${name}`,
          sub: String(members.length), indent: 1, selected: open,
        });
        if (open) for (const a of members) items.push(this._assetRow(a, 2));
      }
    }
    return items;
  }

  _assetRow(a, indent) {
    return {
      id: `asset:${a.id}`,
      kind: 'row',
      label: a.name || a.id,
      indent,
      selected: a.id === this.loadedAsset && this.gameId === this.loadedGame,
      // Not "greyed because broken" — greyed because this session cannot swap
      // to it without ending, which is the one thing the shell exists to avoid.
      disabled: this.busy || !this.canShow(a),
    };
  }

  _controlCols() {
    const items = this.controls.length
      ? this.controls.map((c, i) => ({
        id: c.kind === 'section' ? `sec:${i}` : `ctl:${i}`,
        kind: c.kind === 'section' ? 'group' : (c.kind === 'radio' ? 'radio' : (c.kind === 'toggle' ? 'toggle' : 'row')),
        label: c.label,
        on: !!c.on,
        sub: c.sub || '',
        selected: c.kind === 'button' && c.on,
        disabled: c.disabled,
      }))
      : [{ id: 'none', kind: 'group', label: 'this view offers no controls' }];
    // One column while it fits, two when it does not: these lists are short and
    // paging them is worse than splitting them.
    if (items.length <= 15) return [{ id: 'controls', width: 1, header: null, page: 0, items }];
    const half = Math.ceil(items.length / 2);
    return [
      { id: 'controls', width: 0.5, header: null, page: this.page.controls, items: items.slice(0, half) },
      { id: 'controls2', width: 0.5, header: null, page: this.page.controls, items: items.slice(half) },
    ];
  }

  // Prev/Next walk only what this session can actually show, in manifest order.
  _steps() {
    const g = this.games.find((x) => x.id === this.loadedGame);
    return g ? g.man.assets.filter((a) => this.canShow(a)) : [];
  }

  // ---- clicks ------------------------------------------------------------------

  // Returns true when the panel needs redrawing. Anything that leaves the
  // panel — picking an asset — goes out through onPick and the shell decides.
  click(rect) {
    if (!rect || rect.item?.disabled) return false;
    const id = rect.id;

    if (id === 'browse' || id === 'controls') { this.tab = id; return true; }
    if (id === 'close' || id === 'recenter') { this.onAction(id); return false; }

    const pg = /^(\w+):page(\+\+|\+|--|-)$/.exec(id);
    if (pg) return this._page(pg[1], pg[2]);

    if (id.startsWith('cat:') || id.startsWith('grp:')) {
      if (!this.expanded.delete(id)) this.expanded.add(id);
      this.page.assets = 0;
      return true;
    }

    if (id.startsWith('game:')) {
      const gid = id.slice(5);
      if (gid === this.gameId) return false;
      this.gameId = gid;
      this.expanded.clear();
      this.page.assets = 0;
      // Selecting a game LISTS it; it does not load anything. Loading the
      // default asset on a mis-pinch would throw away the scene you are in.
      return true;
    }

    if (id.startsWith('asset:')) {
      this.onPick(this.gameId, id.slice(6));
      return true;
    }

    if (id === 'step-' || id === 'step+') {
      const steps = this._steps();
      const i = steps.findIndex((a) => a.id === this.loadedAsset);
      const n = i + (id === 'step+' ? 1 : -1);
      if (i < 0 || n < 0 || n >= steps.length) return false;
      this.onPick(this.loadedGame, steps[n].id);
      return true;
    }

    if (id.startsWith('ctl:')) {
      const c = this.controls[Number(id.slice(4))];
      if (!c) return false;
      // The control's own set() is the authority on the new state — a radio
      // turns its siblings off, and only the recorder knows who they are.
      c.set?.(c.kind === 'toggle' ? !c.on : true);
      return true;
    }
    return false;
  }

  _page(colId, op) {
    const key = colId.startsWith('controls') ? 'controls' : colId;
    if (!(key in this.page)) return false;
    const d = op === '+' ? 1 : op === '-' ? -1 : op === '++' ? PAGE_JUMP : -PAGE_JUMP;
    const next = Math.max(0, this.page[key] + d);
    if (next === this.page[key]) return false;
    this.page[key] = next; // layout() clamps to the real page count
    return true;
  }

  // _pageTo scrolls a column so that a given row is on screen. How many rows
  // fit is layout()'s business and depends on the model, so the jump is queued
  // and resolved by applyPending once a layout has been done. Queued, not
  // stored: setLoaded moves BOTH columns, and a single slot would silently
  // drop the first of them.
  _pageTo(rowId, key = 'assets') {
    const items = key === 'games'
      ? this.games.map((g) => ({ id: `game:${g.id}` }))
      : this._assetItems();
    const i = items.findIndex((it) => it.id === rowId);
    if (i < 0) return;
    this.page[key] = 0;
    (this._pending ||= []).push({ key, index: i });
  }

  // Called by the shell after a layout, when the real rows-per-page is known.
  // Returns true if anything moved, meaning the panel must be laid out again.
  applyPending(cols) {
    const q = this._pending;
    if (!q?.length) return false;
    this._pending = null;
    let moved = false;
    for (const p of q) {
      const col = cols.find((c) => c.id === p.key);
      if (!col?.rows) continue;
      const page = Math.floor(p.index / col.rows);
      if (page !== this.page[p.key]) { this.page[p.key] = page; moved = true; }
    }
    return moved;
  }
}
