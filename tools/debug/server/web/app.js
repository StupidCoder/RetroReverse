// framedbg — the frame debugger's page.
//
// The loop is: step a frame (the emulator captures its command stream and per-pixel
// provenance), then scrub through that frame command by command. Every scrub position
// is a real replay in the emulator; everything about *pixels* — which command drew
// one, which pixels a command drew — is answered here from the provenance buffer.

import { Conn, KIND_IMAGE, KIND_PROV } from './conn.js';
import { Viewport } from './viewport.js';
import { CommandList } from './commands.js';
import { renderCPU, renderMem, renderOverdraw } from './panels.js';

const $ = (id) => document.getElementById(id);

const ui = {
  target: $('target'),
  play: $('play'),
  step: $('step'),
  step1: $('step1'),
  overdraw: $('overdraw'),
  scanout: $('scanout'),
  zoom: $('zoom'),
  stats: $('stats'),
  slider: $('slider'),
  prev: $('prev'),
  next: $('next'),
  scrubLabel: $('scrub-label'),
  hover: $('hover'),
  memAddr: $('mem-addr'),
};

const view = new Viewport();
const cmds = new CommandList(selectCommand);
const conn = new Conn();

let frame = null; // the last frameMsg
let selected = -1; // current command index
let wantSeq = -1; // the image request whose reply we still want; older ones are stale
let playing = false;
let fps = { n: 0, t: 0, rate: 0 };
const pending = new Map(); // seq -> renderMsg, so the binary that follows knows its own cost

conn.onOpen = () => {
  setBusy(false);
  ui.target.textContent = 'connected';
};

conn.onClose = () => {
  ui.target.textContent = 'disconnected — restart framedbg and reload';
  setBusy(true);
};

conn.onJSON = (m) => {
  switch (m.type) {
    case 'hello':
      ui.target.textContent = `${m.target} — ${m.rom}`;
      document.title = `framedbg — ${m.rom}`;
      break;

    case 'frame':
      frame = m;
      selected = -1;
      view.setProv(null);
      view.select(-1);
      cmds.setCommands(m.commands);
      ui.slider.max = Math.max(0, m.commands.length - 1);
      ui.slider.disabled = m.commands.length === 0;
      setBusy(false);
      if (m.commands.length) {
        selectCommand(m.commands.length - 1);
      } else {
        // A field the game drew nothing in — common during boot, and not a failure.
        // Show what is on screen rather than a blank canvas and an empty list.
        ui.scrubLabel.textContent = '—';
        ui.stats.textContent = `field ${m.frame} · nothing drawn`;
        ui.hover.textContent = 'no commands in this field';
        showDisplay();
      }
      conn.send('cpu');
      readMem();
      break;

    case 'render':
      pending.set(m.seq, m);
      break;

    case 'cpu':
      renderCPU(m);
      break;

    case 'mem':
      renderMem(m);
      break;

    case 'pixel':
      renderOverdraw(m, (k) => selectCommand(k));
      break;

    case 'error':
      ui.stats.textContent = `error: ${m.msg}`;
      setBusy(false);
      break;
  }
};

conn.onBinary = (m) => {
  if (m.kind === KIND_PROV) {
    view.setProv(m.prov);
    return;
  }
  if (m.kind !== KIND_IMAGE) return;

  const meta = pending.get(m.seq);
  pending.delete(m.seq);

  if (meta && meta.play) {
    // A free-running frame. Draw it and acknowledge: the server holds itself to one
    // unacknowledged frame, so this is what keeps the stream paced to what the page
    // can actually paint.
    view.drawImage(m.image);
    conn.send('ack');
    countFPS(meta);
    return;
  }

  // A scrubber drag outruns the emulator, so replies to positions the mouse has
  // already left arrive after we have asked for a newer one. Drop those.
  if (m.seq !== wantSeq) return;

  view.drawImage(m.image);
  view.redrawOverlay();
  if (meta) showStats(meta);
};

// ---- actions ----

function stepFrame(n) {
  if (playing) return;
  setBusy(true);
  ui.stats.textContent = 'stepping…';
  conn.send('step', { overdraw: ui.overdraw.checked, n });
}

// Play free-runs the machine, streaming the scanout and capturing nothing — the way
// to get to the part of the game you actually want to look at. Stopping captures the
// next field in full, so you land on a real frame with its commands and provenance.
function setPlaying(on) {
  playing = on;
  ui.play.textContent = on ? '⏸ Pause' : '▶ Play';
  ui.play.classList.toggle('active', on);
  ui.step.disabled = on;
  ui.step1.disabled = on;
  ui.slider.disabled = on || !frame || !frame.commands.length;
  ui.prev.disabled = on;
  ui.next.disabled = on;

  if (on) {
    // Nothing captured while playing, so the command list, provenance and overlay
    // would all be lies. Clear them rather than leave a stale frame on screen.
    frame = null;
    selected = -1;
    view.setProv(null);
    view.select(-1);
    view.setPick(null);
    cmds.setCommands([]);
    ui.scrubLabel.textContent = '—';
    ui.hover.textContent = 'playing — no capture';
    fps = { n: 0, t: performance.now(), rate: 0 };
    conn.send('play', { on: true });
  } else {
    ui.stats.textContent = 'capturing…';
    conn.send('play', { on: false, overdraw: ui.overdraw.checked });
  }
}

function countFPS(meta) {
  const now = performance.now();
  fps.n++;
  if (now - fps.t >= 500) {
    fps.rate = (fps.n * 1000) / (now - fps.t);
    fps.n = 0;
    fps.t = now;
  }
  ui.stats.textContent = `playing · field ${meta.frame.toLocaleString()} · ${fps.rate.toFixed(0)} fps`;
}

// selectCommand is the hub: it picks a command in the list, rewinds the scrubber to
// it, asks the emulator for the draw target as it stood after it, and lights up the
// pixels it drew.
function selectCommand(k) {
  if (!frame || k < 0 || k >= frame.commands.length) return;
  selected = k;
  cmds.select(k);
  view.select(k);
  ui.slider.value = k;
  ui.scrubLabel.textContent = `${k} / ${frame.commands.length - 1}`;
  if (!ui.scanout.checked) wantSeq = conn.send('scrub', { k });
}

function showDisplay() {
  wantSeq = conn.send('display');
}

function readMem() {
  const addr = parseInt(ui.memAddr.value, 16);
  if (!Number.isNaN(addr)) conn.send('mem', { addr, len: 256 });
}

function showStats(meta) {
  const kb = (meta.bytes / 1024).toFixed(0);
  const where = meta.k < 0 ? 'scanout' : `cmd ${meta.k}`;
  const how = meta.cached ? 'cached' : `${meta.renderMs.toFixed(1)} ms`;
  const n = frame ? `${frame.commands.length.toLocaleString()} cmds · ` : '';
  ui.stats.textContent = `${n}${where} · ${how} · ${kb} KB`;
}

function setBusy(b) {
  ui.step.disabled = b;
  ui.step1.disabled = b;
}

// ---- wiring ----

ui.play.onclick = () => setPlaying(!playing);
ui.step.onclick = () => stepFrame(0);
ui.step1.onclick = () => stepFrame(1);

ui.slider.oninput = () => selectCommand(Number(ui.slider.value));
ui.prev.onclick = () => selectCommand(selected - 1);
ui.next.onclick = () => selectCommand(selected + 1);

ui.zoom.onchange = () => view.setZoom(Number(ui.zoom.value));

ui.scanout.onchange = () => {
  if (ui.scanout.checked) showDisplay();
  else selectCommand(selected);
};

ui.memAddr.onchange = readMem;

view.onHover = (x, y) => {
  if (playing) return; // nothing is captured while playing, so there is nothing to name
  if (x < 0) {
    ui.hover.textContent = 'hover a pixel';
    return;
  }
  const k = view.provAt(x, y);
  const name = frame && k >= 0 && frame.commands[k] ? frame.commands[k].name : null;
  ui.hover.textContent = name
    ? `pixel (${x}, ${y}) ← #${k} ${name}`
    : `pixel (${x}, ${y}) — untouched`;
};

view.onPick = (x, y) => {
  if (playing) return;
  view.setPick({ x, y });
  conn.send('pixel', { x, y });
  const k = view.provAt(x, y);
  if (k >= 0) selectCommand(k);
};

document.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
  const stride = e.shiftKey ? 10 : 1;
  switch (e.key) {
    case ' ':
      setPlaying(!playing);
      e.preventDefault();
      break;
    case 'n':
    case 'N':
      stepFrame(0);
      break;
    case 'ArrowLeft':
      selectCommand(selected - stride);
      e.preventDefault();
      break;
    case 'ArrowRight':
      selectCommand(selected + stride);
      e.preventDefault();
      break;
    case 'Home':
      selectCommand(0);
      break;
    case 'End':
      if (frame) selectCommand(frame.commands.length - 1);
      break;
  }
});

setBusy(true);
view.setZoom(Number(ui.zoom.value));
