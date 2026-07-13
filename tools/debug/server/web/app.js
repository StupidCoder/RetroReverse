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

// Importing a panel registers it. Order here is the order they appear in a slot.
import './panels/viewport.js';
import './panels/commands.js';
import './panels/inspect.js';
import './panels/surface.js';
import './panels/states.js';

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
  if (store.can('replay') && !$('scanout').checked) {
    wantSeq = conn.send('frame.scrub', { k });
  }
};

ctx.ui.showDisplay = () => {
  wantSeq = conn.send('frame.display');
};

// Play free-runs the machine, streaming the scanout and capturing nothing — the way to
// reach the part of the game you actually want to look at. Stopping captures the next
// field in full, so you land on a real frame with its commands and provenance.
ctx.ui.setPlaying = (on) => {
  store.set({ playing: on });
  $('play').textContent = on ? '⏸ Pause' : '▶ Play';
  $('play').classList.toggle('active', on);
  setBusy(on);

  if (on) {
    // Nothing is captured while playing, so a command list or an overlay left on
    // screen would be a lie. Clear them.
    store.set({ frame: null, selected: -1, prov: null, pick: null });
    fps = { n: 0, t: performance.now(), rate: 0 };
    conn.send('frame.play', { on: true });
  } else {
    status('capturing…');
    conn.send('frame.play', { on: false, overdraw: $('overdraw').checked });
  }
};

const status = (s) => ($('stats').textContent = s);
ctx.ui.status = status;

function setBusy(b) {
  $('step').disabled = b;
  $('step1').disabled = b;
}

// ---- wiring ----

conn.onOpen = () => setBusy(false);
conn.onClose = () => {
  $('target').textContent = 'disconnected — restart framedbg and reload';
  setBusy(true);
};

conn.on('hello', (m) => {
  store.set({
    platform: m.platform,
    title: m.title,
    caps: new Set(m.caps),
    frame: null,
    selected: -1,
    prov: null,
    pick: null,
    playing: false,
  });
  $('target').textContent = `${m.platform} — ${m.title}`;
  document.title = `framedbg — ${m.title}`;

  // The target's capabilities decide what the page is. A different game means a
  // different set of panels, so the dock is rebuilt from scratch.
  mountPanels(ctx);
  for (const id of ['play', 'step', 'step1']) {
    $(id).style.display = store.can('frames') ? '' : 'none';
  }
  $('cpu-controls').style.display = store.can('code') ? '' : 'none';

  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();
  if (!store.can('frames')) ctx.ui.showDisplay();
});

// The library: which games have an adapter and an image on disk.
conn.on('library', (m) => {
  const sel = $('game');
  sel.innerHTML =
    '<option value="">library…</option>' +
    m.games
      .map(
        (g) =>
          `<option value="${g.slug}"${g.slug === m.current ? ' selected' : ''}${g.missing ? ' disabled' : ''}>` +
          `${g.name} (${g.platform})${g.missing ? ' — no image' : ''}</option>`
      )
      .join('');
});

// Loading a state moves the machine somewhere else entirely; the captured frame
// belonged to where we were.
ctx.ui.afterStateLoad = () => {
  store.set({ frame: null, selected: -1, prov: null, pick: null });
  conn.send('cpu.regs');
  if (ctx.ui.readMem) ctx.ui.readMem();
  if (ctx.ui.refreshDisasm) ctx.ui.refreshDisasm();
  ctx.ui.showDisplay();
};

conn.on('frame', (m) => {
  store.set({ frame: m, selected: -1, prov: null, pick: null });
  setBusy(false);
  if (m.commands.length) {
    ctx.ui.selectCommand(m.commands.length - 1);
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
});
conn.on('stopped', (m) => {
  status(`stopped: ${m.reason}${m.note ? ` (${m.note})` : ''} at ${m.pc.slice(-8)}`);
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
    // A free-running frame. Draw it and acknowledge: the server holds itself to one
    // unacknowledged frame, so this is what paces the stream to what the page can paint.
    ctx.viewport.drawImage(m.image);
    conn.send('frame.ack');
    countFPS(meta);
    return;
  }
  // A scrubber drag outruns the emulator, so replies to positions the mouse has already
  // left arrive after we have asked for a newer one. Drop those.
  if (m.seq !== wantSeq) return;

  ctx.viewport.drawImage(m.image);
  if (meta) showStats(meta);
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
$('zoom').onchange = () => ctx.viewport.setZoom(Number($('zoom').value));
$('scanout').onchange = () => {
  if ($('scanout').checked) ctx.ui.showDisplay();
  else ctx.ui.selectCommand(store.get('selected'));
};
$('cpu-step').onclick = () => conn.send('cpu.step', { n: 1 });
$('cpu-run').onclick = () => conn.send('cpu.continue');
$('cpu-brk').onclick = () => conn.send('cpu.break');

document.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
  const stride = e.shiftKey ? 10 : 1;
  const selected = store.get('selected');
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
      ctx.ui.selectCommand(selected - stride);
      e.preventDefault();
      break;
    case 'ArrowRight':
      ctx.ui.selectCommand(selected + stride);
      e.preventDefault();
      break;
    case 'Home':
      ctx.ui.selectCommand(0);
      break;
    case 'End': {
      const frame = store.get('frame');
      if (frame) ctx.ui.selectCommand(frame.commands.length - 1);
      break;
    }
  }
});

setBusy(true);
