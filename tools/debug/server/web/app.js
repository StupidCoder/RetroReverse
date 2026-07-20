// framedbg — the debugger's shell.
//
// The shell owns the connection, the shared store, and the handful of actions panels
// invoke (step a frame, select a command, play). It does not know what panels exist:
// the target says what it can do, and the registry mounts the panels those
// capabilities support. So a platform with no per-pixel provenance simply has no
// overdraw panel, and adding a tool means registering it, not editing this file.

import { Conn, KIND_IMAGE, KIND_PROV, STREAM_MAIN } from './conn.js';
import { Store } from './store.js';
import { mountPanels } from './panels/registry.js';
import { reveal, resetLayout, toggleMaximize, unmaximize } from './panels/dock.js';

// Importing a panel registers it. Order here is the order they appear in a slot.
import './panels/viewport.js';
import './panels/commands.js';
import './panels/inspect.js';
import './panels/surface.js';
import './panels/files.js';
import './panels/states.js';
import './panels/profile.js';

const $ = (id) => document.getElementById(id);

const conn = new Conn();
const store = new Store();
const ctx = { conn, store, ui: {}, viewport: null };

let wantSeq = -1; // the image request whose reply we still want; older ones are stale
let fps = { n: 0, t: 0, rate: 0 };
const pending = new Map(); // seq -> renderMsg, so the binary that follows knows its cost

// ---- actions the panels call ----

ctx.ui.stepFrame = (n) => {
  if (store.get('playing')) return;
  setBusy(true);
  status('stepping…');
  conn.send('frame.step', { n, overdraw: $('overdraw').checked });
};

// selectCommand is the hub: it selects a command, moves the scrubber to it, asks the
// emulator for the draw target as it stood after it, and lights up the pixels it drew.
ctx.ui.selectCommand = (k) => {
  const frame = store.get('frame');
  if (!frame || k < 0 || k >= frame.commands.length) return;
  store.set({ selected: k });
  // The stage may be showing a buffer of the machine rather than the frame. The command is
  // still selected — the list, the detail panel and the scrubber all follow it — but the
  // picture is not the frame's, so replaying the frame into it would be a lie.
  if (ctx.viewport.buffer) return;
  if (store.can('replay') && !$('scanout').checked) scrubTo(k);
};

// redraw puts back on the stage whatever the frame view currently means: the selected
// command's draw target, or the scanout when that is what is asked for.
ctx.ui.redraw = () => {
  scrubReset();
  const k = store.get('selected');
  if (k >= 0 && store.can('replay') && !$('scanout').checked) scrubTo(k);
  else ctx.ui.showDisplay();
};

// The scrubber walks the commands that WROTE, not the whole command stream.
//
// Most of a display processor's stream does not draw. A 3DS frame is around two hundred
// thousand PICA register writes, of which a few hundred are draws and the rest set up the
// state those draws run against — so a scrubber over all of them is a scrubber that shows
// the same picture for thousands of positions and then jumps. Over the writers it is what
// it was always meant to be: a tour of how the frame got built, one stroke at a time.
//
// The list still shows every command, and selecting one from the list still renders the
// frame exactly as it stood after it — a register write you can inspect the effect of. The
// scrubber is a route through the stream, not a filter on it.
//
// A platform whose rasteriser cannot report pixels sends no writer list at all, and there
// the scrubber walks everything, as it always did.
function scrubList() {
  const f = store.get('frame');
  const cmds = f ? f.commands.length : 0;
  const w = f && f.writers; // null: this platform does not say which commands wrote
  if (!f || !w) {
    return { n: cmds, at: (i) => i, indexOf: (k) => k, all: !w };
  }
  return {
    n: w.length,
    at: (i) => w[Math.max(0, Math.min(w.length - 1, i))] ?? -1,
    // Where the scrubber sits for a command that wrote nothing — one picked out of the
    // command list — is the last writer before it: the picture on the stage is the one that
    // command left behind, and the handle should say so rather than jump somewhere else.
    indexOf: (k) => {
      let lo = 0;
      let hi = w.length - 1;
      let at = -1;
      while (lo <= hi) {
        const mid = (lo + hi) >> 1;
        if (w[mid] <= k) {
          at = mid;
          lo = mid + 1;
        } else {
          hi = mid - 1;
        }
      }
      return at;
    },
    all: false,
  };
}
ctx.ui.scrubList = scrubList;

// stepScrub moves along the writers — the arrow keys and the scrubber's own buttons.
ctx.ui.stepScrub = (d) => {
  const s = scrubList();
  if (!s.n) return;
  const cur = s.indexOf(store.get('selected'));
  const next = Math.max(0, Math.min(s.n - 1, (cur < 0 ? 0 : cur) + d));
  ctx.ui.selectCommand(s.at(next));
};

ctx.ui.scrubEnd = (last) => {
  const s = scrubList();
  if (!s.n) return;
  ctx.ui.selectCommand(s.at(last ? s.n - 1 : 0));
};

// A drag on the scrubber is a stream of positions arriving far faster than a 3DS frame
// can be replayed, and the naive thing — send them all, draw only the newest — makes the
// picture stand still until you let go, while the emulator replays every position you
// dragged over. So the page holds itself to ONE scrub in flight and remembers only the
// latest position asked for while it waits.
//
// That is what makes the drag live: every reply is a picture the user asked for and every
// picture gets drawn, the emulator is never asked for a position the mouse has already
// left, and the last position always lands because the pending one is sent the moment the
// one before it comes back. It also lets the server's batch replay do its job — the
// positions the drag is about to reach are prefetched around each request, so most of what
// the drag asks for next is already in the cache and comes back with no replay at all.
let scrubSeq = -1; // the scrub whose reply we are waiting for, or -1
let scrubWant = -1; // a position asked for while one was in flight, or -1
let scrubAt = -1; // the position now on screen

function scrubTo(k) {
  if (scrubSeq >= 0) {
    scrubWant = k; // coalesce: only where the mouse ends up matters
    return;
  }
  if (k === scrubAt) return;
  scrubWant = -1;
  scrubSeq = conn.send('frame.scrub', { k, blank: $('blank').checked });
  wantSeq = scrubSeq;
}

// scrubDone is called when the scrub in flight has been answered — with a picture or with
// an error, because a scrub that failed must not wedge the drag.
function scrubDone(k) {
  scrubSeq = -1;
  if (k >= 0) scrubAt = k;
  if (scrubWant >= 0) {
    const next = scrubWant;
    scrubWant = -1;
    scrubTo(next);
  }
}

// A new capture, or a switch to the scanout, makes what is on screen no longer a scrub
// position — so the next selection must be sent even if it names the same command.
function scrubReset() {
  scrubSeq = -1;
  scrubWant = -1;
  scrubAt = -1;
}

ctx.ui.showDisplay = () => {
  scrubReset();
  if (ctx.viewport && ctx.viewport.drawBuffer()) return -1; // the stage is showing a buffer
  wantSeq = conn.send('frame.display');
  return wantSeq;
};

// Play free-runs the machine, streaming the scanout and capturing nothing — the way to
// reach the part of the game you actually want to look at. Pausing stops the machine and
// captures a drawn frame in full, so you land on something with commands and provenance
// to inspect, without having to step for it.
ctx.ui.setPlaying = (on) => {
  store.set({ playing: on });
  $('play').textContent = on ? '⏸ Pause' : '▶ Play';
  $('play').classList.toggle('active', on);
  setBusy(true); // the capture the pause kicks off is a step like any other

  if (on) {
    // Nothing is captured while playing, so a command list or an overlay left on
    // screen would be a lie. Clear them.
    store.set({ frame: null, selected: -1, prov: null, pick: null });
    scrubReset();
    fps = { n: 0, t: performance.now(), rate: 0 };
    conn.send('frame.play', { on: true });
  } else {
    status('capturing…');
    conn.send('frame.play', { on: false, overdraw: $('overdraw').checked });
  }
};

const status = (s) => ($('stats').textContent = s);
ctx.ui.status = status;

// reveal brings a panel's tab forward. A panel calls it when something arrives that the
// user is waiting to see — the overdraw history of the pixel they just clicked — so the
// answer lands in front of them rather than behind a tab they are not looking at.
ctx.ui.reveal = reveal;

function setBusy(b) {
  $('step').disabled = b;
  $('step1').disabled = b;
}

// ---- wiring ----

conn.onOpen = () => setBusy(false);
conn.onClose = () => {
  // The status line, not the target label: losing the server is a status, and the label is
  // hidden whenever the library's menus already name what is open.
  status('disconnected — restart framedbg and reload');
  setBusy(true);
};

conn.on('hello', (m) => {
  store.set({
    platform: m.platform,
    title: m.title,
    caps: new Set(m.caps),
    keyLegend: m.keyLegend || '',
    // The display aspect (num/den), for a target whose pixels are not square (the PS2). 0
    // means square — the page scales the display, not the image, so clicks still map.
    aspectNum: m.aspectNum || 0,
    aspectDen: m.aspectDen || 0,
    frame: null,
    selected: -1,
    prov: null,
    pick: null,
    playing: false,
    cpuRunning: false,
  });
  // The target label is for the case where nothing else names the target: framedbg opened
  // straight onto an image with -image, so there is no library and no menus. With a library
  // the two menus say the platform and the game, the tab says the image, and this would only
  // be a third copy — an expensive one, since the toolbar is the width of the window and the
  // status line is what gives up its characters to pay for it.
  $('target').textContent = `${m.platform} — ${m.title}`;
  $('target').hidden = library.length > 0;
  document.title = `framedbg — ${m.title}`;

  // The target's capabilities decide what the page is. A different game means a
  // different set of panels, so the dock is rebuilt from scratch.
  mountPanels(ctx);
  for (const id of ['play', 'step', 'step1']) {
    $(id).style.display = store.can('frames') ? '' : 'none';
  }
  $('cpu-controls').style.display = store.can('code') ? '' : 'none';

  // A hello is a machine handed over: the one before it is closed and gone, and every
  // control on the page still belongs to it. The toolbar is reset here rather than by
  // whoever opened the game, because a hello also arrives unasked — the page reconnecting,
  // or another tab opening a game on the same server.
  $('play').textContent = '▶ Play';
  $('play').classList.remove('active');
  scrubReset();
  setBusy(false); // the step buttons were disabled while the game booted
  ctx.viewport.setZoom($('zoom').value); // the new viewport starts at the zoom the toolbar shows

  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();

  // Show whatever the machine has on its screen. A freshly opened game has nothing — it is
  // at its entry point, and on the N64 the video interface is not even scanning a
  // framebuffer yet — so this quite legitimately fails. That is a cold boot, not a fault,
  // and it is answered with what to do about it.
  conn.onError(ctx.ui.showDisplay(), () => {
    status(store.can('frames') ? 'booted — ▶ Play to run it, or Step for a frame' : 'booted');
  });
});

// The library: which games have an adapter and an image on disk.
//
// Two menus rather than one, because one flat list of every game on every platform is a list
// that only gets longer — and the platform is how anyone actually navigates it ("the
// GameCube one"). So the first menu is platforms, by their human names, and choosing one
// reveals that platform's games. The game menu is hidden until then: an empty second menu
// beside the first is a control that looks broken.
let library = [];

conn.on('library', (m) => {
  library = m.games;

  // Platforms, in the order their names read, each with what is behind it. A platform whose
  // every image is missing is still shown — "you do not have those" is worth saying, and the
  // games under it say it individually.
  const platforms = [...new Set(library.map((g) => g.platform))].sort((a, b) =>
    platformName(a).localeCompare(platformName(b))
  );
  const current = library.find((g) => g.slug === m.current);

  $('platform').innerHTML =
    '<option value="">library…</option>' +
    platforms
      .map((p) => {
        const n = library.filter((g) => g.platform === p).length;
        const sel = current && current.platform === p ? ' selected' : '';
        return `<option value="${p}"${sel}>${platformName(p)} (${n})</option>`;
      })
      .join('');

  // With a game already open — another tab opened it, or this page reloaded — the menus show
  // where we are rather than making you find it again.
  fillGames(current ? current.platform : '', current ? current.slug : '');
});

// platformName is the human name the server sent for a platform tag. It comes from the games
// rather than a table here, so a new adapter names itself in one place (debug/registry.go).
function platformName(tag) {
  const g = library.find((x) => x.platform === tag);
  return (g && g.platformName) || tag;
}

// fillGames populates the second menu with one platform's games, or hides it.
function fillGames(platform, selectSlug) {
  const sel = $('game');
  if (!platform) {
    sel.hidden = true;
    sel.innerHTML = '';
    return;
  }
  const games = library
    .filter((g) => g.platform === platform)
    .sort((a, b) => a.name.localeCompare(b.name));
  sel.innerHTML =
    `<option value="">${games.length === 1 ? 'game…' : `${games.length} games…`}</option>` +
    games
      .map(
        (g) =>
          `<option value="${g.slug}"${g.slug === selectSlug ? ' selected' : ''}${g.missing ? ' disabled' : ''}>` +
          `${g.name}${g.missing ? ' — no image' : ''}</option>`
      )
      .join('');
  sel.hidden = false;
}

// Loading a state moves the machine somewhere else entirely; the captured frame belonged
// to where we were. A restored machine is between frames — it has drawn nothing yet — so,
// as with pausing, landing there means capturing a frame: the state you loaded is one you
// saved to look at something, and having to press Step to see it is a step you did not ask
// for. The capture advances the machine to the state's next drawn frame, which is where
// you wanted to be anyway.
ctx.ui.afterStateLoad = () => {
  store.set({ frame: null, selected: -1, prov: null, pick: null });
  scrubReset();
  if (store.get('playing')) return; // the stream already shows where we landed
  if (store.can('frames')) {
    ctx.ui.stepFrame(0); // its reply refreshes the registers, memory and disassembly too
    return;
  }
  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();
  if (ctx.ui.refreshDisasm) ctx.ui.refreshDisasm();
  ctx.ui.showDisplay();
};

conn.on('frame', (m) => {
  store.set({ frame: m, selected: -1, prov: null, pick: null });
  scrubReset(); // a new capture: the server's replay cache and our position both belong to the old one
  setBusy(false);
  if (m.commands.length) {
    // The finished picture: the last command that wrote. Landing on the last COMMAND would
    // put the handle at the far right of a scrubber whose positions stop well before it.
    const s = scrubList();
    ctx.ui.selectCommand(s.n ? s.at(s.n - 1) : m.commands.length - 1);
  } else {
    // A field the game drew nothing in — common during boot, and not a failure. Show
    // what is on screen rather than a blank canvas and an empty list.
    status(`field ${m.frame} · nothing drawn`);
    ctx.ui.showDisplay();
  }
  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();
  if (ctx.ui.refreshDisasm) ctx.ui.refreshDisasm();
});

conn.on('render', (m) => pending.set(m.seq, m));
conn.on('error', (m) => {
  status(`error: ${m.msg}`);
  setBusy(false);
  // A scrub that failed answers nothing, and would otherwise wedge the drag forever.
  if (m.seq === scrubSeq) scrubDone(-1);
});

// The boot note: which game is open, and how long it took to get there.
conn.on('ok', (m) => {
  if (m.op === 'target.open' && m.note) status(m.note);
});
conn.on('stopped', (m) => {
  status(`stopped: ${m.reason}${m.note ? ` (${m.note})` : ''} at ${m.pc.slice(-8)}`);
  // A halted machine ends play mode from the SERVER's side — the core has stopped for
  // good, and the runner leaves play mode on its own account. The page has to follow, but
  // it must not do it through setPlaying(false): that sends frame.play{on:false} asking
  // for the capture a pause normally lands on, and a stopped core will never produce one,
  // so the page would sit busy forever waiting for it. Reset the control locally instead
  // and leave the last picture up, which is the frame the halt happened on.
  if (store.get('playing')) {
    store.set({ playing: false });
    $('play').textContent = '▶ Play';
    $('play').classList.remove('active');
    setBusy(false);
  }
  // The CPU stopped free-running (a breakpoint, a watch, or Break). Turn off the
  // live-scanout draw path, then take one last look at where it landed: the screen,
  // the registers, and memory.
  if (store.get('cpuRunning')) {
    store.set({ cpuRunning: false });
    if (!store.can('frames')) ctx.ui.showDisplay();
  }
  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();
});

conn.onBinary((m) => {
  if (m.kind === KIND_PROV) {
    store.set({ prov: m.prov });
    return;
  }
  if (m.kind !== KIND_IMAGE || m.stream !== STREAM_MAIN) return;

  const meta = pending.get(m.seq);
  pending.delete(m.seq);

  if (meta && meta.play) {
    // A "play" frame streams both from play mode and from a free-running CPU on a
    // CPU-only target (the DOS host, which has no frames to play). Either way it is a
    // live scanout; draw it only while we are actually running, so a straggler from a
    // machine we just closed does not paint over the one we just opened.
    const live = store.get('playing') || store.get('cpuRunning');
    if (!live) return;

    // Draw it and acknowledge: the server holds itself to one unacknowledged frame, so
    // this is what paces the stream to what the page can paint.
    ctx.viewport.drawImage(m.image);
    conn.send('frame.ack');
    if (store.get('playing')) countFPS(meta);
    else status(`running · ${store.get('pc') || ''}`);
    return;
  }
  // Anything we no longer want — a display request superseded by a newer one — is stale.
  if (m.seq !== wantSeq && m.seq !== scrubSeq) return;

  ctx.viewport.drawImage(m.image);
  if (meta) showStats(meta);

  // Drawn, and only now: releasing the scrub sends the position the mouse moved to while
  // this one was rendering, which makes it the newest request.
  if (m.seq === scrubSeq) scrubDone(meta ? meta.k : -1);
});

function countFPS(meta) {
  const now = performance.now();
  fps.n++;
  if (now - fps.t >= 500) {
    fps.rate = (fps.n * 1000) / (now - fps.t);
    fps.n = 0;
    fps.t = now;
  }
  status(`playing · field ${meta.frame.toLocaleString()} · ${fps.rate.toFixed(0)} fps`);
}

function showStats(meta) {
  const frame = store.get('frame');
  const where = meta.k < 0 ? 'scanout' : `cmd ${meta.k}`;
  const how = meta.cached ? 'cached' : `${meta.renderMs.toFixed(1)} ms`;
  const n = frame ? `${frame.commands.length.toLocaleString()} cmds · ` : '';
  status(`${n}${where} · ${how} · ${(meta.bytes / 1024).toFixed(0)} KB`);
}

// ---- toolbar ----

// Picking a platform only reveals its games — it opens nothing. Opening is the second
// choice, deliberately: the platform menu is navigation, and navigating should not boot a
// machine.
$('platform').onchange = () => {
  fillGames($('platform').value, '');
  if ($('platform').value) $('game').focus();
};

$('game').onchange = () => {
  const slug = $('game').value;
  if (!slug) return;
  status(`opening ${slug}…`);
  setBusy(true);
  conn.send('target.open', { slug });
};

$('play').onclick = () => ctx.ui.setPlaying(!store.get('playing'));
$('step').onclick = () => ctx.ui.stepFrame(0);
$('step1').onclick = () => ctx.ui.stepFrame(1);
$('zoom').onchange = () => ctx.viewport.setZoom($('zoom').value);
$('scanout').onchange = () => {
  if ($('scanout').checked) ctx.ui.showDisplay();
  else ctx.ui.selectCommand(store.get('selected'));
};
// Blank changes what every position of the scrub renders, so the picture on screen is the
// one thing it cannot leave alone.
$('blank').onchange = () => ctx.ui.redraw();
$('cpu-step').onclick = () => conn.send('cpu.step', { n: 1 });
$('cpu-run').onclick = () => {
  // A CPU-only target streams its scanout while it runs; the flag is what lets the
  // render path draw those unrequested frames (see onBinary). A frame-stepping target
  // keeps its screen alive through play mode instead, so the flag is harmless there.
  store.set({ cpuRunning: true });
  conn.send('cpu.continue');
};
$('cpu-brk').onclick = () => {
  store.set({ cpuRunning: false });
  conn.send('cpu.break');
};
$('reset-layout').onclick = () => resetLayout();

document.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
  const stride = e.shiftKey ? 10 : 1; // a stride is ten WRITERS now, not ten commands
  switch (e.key) {
    case ' ':
      ctx.ui.setPlaying(!store.get('playing'));
      e.preventDefault();
      break;
    case 'n':
    case 'N':
      ctx.ui.stepFrame(0);
      break;
    case 'ArrowLeft':
      ctx.ui.stepScrub(-stride);
      e.preventDefault();
      break;
    case 'ArrowRight':
      ctx.ui.stepScrub(stride);
      e.preventDefault();
      break;
    case 'Home':
      ctx.ui.scrubEnd(false);
      break;
    case 'End':
      ctx.ui.scrubEnd(true);
      break;
    case 'f':
    case 'F':
      toggleMaximize();
      break;
    case 'Escape':
      unmaximize();
      break;
  }
});

setBusy(true);
