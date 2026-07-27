// viewmedia.js — the four media category views: a music player with a track
// list, repeat/shuffle and a VU meter; an SFX soundboard; a zoomable picture
// stage; and a plain video player.

export async function mount(ctx) {
  switch (ctx.asset.category) {
    case 'music': return mountMusic(ctx);
    case 'sfx': return mountSfx(ctx);
    case 'picture': return mountPicture(ctx);
    case 'video': return mountVideo(ctx);
  }
  throw new Error(`no media view for ${ctx.asset.category}`);
}

const fmtTime = (s) => `${(s / 60) | 0}:${String((s | 0) % 60).padStart(2, '0')}`;

// ---- music -----------------------------------------------------------------------

function mountMusic(ctx) {
  const { stage, game, asset } = ctx;
  const wrap = document.createElement('div');
  wrap.className = 'media-wrap';
  const inner = document.createElement('div');
  inner.className = 'media-inner';
  wrap.appendChild(inner);
  stage.appendChild(wrap);

  const tracks = game.assets('music');
  inner.innerHTML = `<h2 style="margin:0">Music</h2><div class="dim" style="color:var(--dim)">${tracks.length} tracks · ${escape(game.man.title)}</div>`;

  const audio = new Audio();
  audio.preload = 'auto';
  let actx = null, analyser = null, freq = null;
  const ensureAnalyser = () => {
    if (actx) return;
    actx = new AudioContext();
    const src = actx.createMediaElementSource(audio);
    analyser = actx.createAnalyser();
    analyser.fftSize = 64;
    src.connect(analyser);
    analyser.connect(actx.destination);
    freq = new Uint8Array(analyser.frequencyBinCount);
  };

  let repeat = false, shuffle = false;
  let cur = -1;
  const rows = [];

  const list = document.createElement('div');
  list.className = 'tracklist';
  tracks.forEach((t, i) => {
    const row = document.createElement('div');
    row.className = 'track';
    row.innerHTML = `<span class="t-name"></span><span class="t-dur"></span>`;
    row.querySelector('.t-name').textContent = t.name;
    row.querySelector('.t-dur').textContent = t.duration ? fmtTime(t.duration) : '';
    if (t.description) row.title = t.description;
    row.onclick = () => play(i);
    list.appendChild(row);
    rows.push(row);
  });
  inner.appendChild(list);

  const bar = document.createElement('div');
  bar.className = 'player-bar';
  bar.innerHTML = `
    <button class="b-play"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg></button>
    <input type="range" min="0" max="1000" value="0" />
    <span class="t-time">0:00</span>
    <button class="b-rep" title="Repeat"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 2l4 4-4 4"/><path d="M3 11v-1a4 4 0 0 1 4-4h14"/><path d="M7 22l-4-4 4-4"/><path d="M21 13v1a4 4 0 0 1-4 4H3"/></svg></button>
    <button class="b-shuf" title="Shuffle"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 3h5v5"/><path d="M4 20L21 3"/><path d="M21 16v5h-5"/><path d="M15 15l6 6"/><path d="M4 4l5 5"/></svg></button>
    <span class="vu"></span>`;
  inner.appendChild(bar);
  const playBtn = bar.querySelector('.b-play');
  const seek = bar.querySelector('input');
  const timeEl = bar.querySelector('.t-time');
  const vu = bar.querySelector('.vu');
  for (let i = 0; i < 12; i++) vu.appendChild(document.createElement('i'));

  // Shuffle draws from a bag of every track index, refilled only when empty,
  // so no track repeats until all have played once. Manual picks count
  // toward the current bag.
  let bag = [];
  const drawShuffled = () => {
    if (!bag.length) {
      bag = tracks.map((_, i) => i);
      for (let i = bag.length - 1; i > 0; i--) {
        const j = (Math.random() * (i + 1)) | 0;
        [bag[i], bag[j]] = [bag[j], bag[i]];
      }
      // never play the track that just ended back-to-back across a refill
      if (bag.length > 1 && bag[bag.length - 1] === cur) bag.unshift(bag.pop());
    }
    return bag.pop();
  };

  const play = (i) => {
    cur = i;
    bag = bag.filter((j) => j !== i);
    rows.forEach((r, j) => r.classList.toggle('playing', j === i));
    audio.src = game.url('', tracks[i].file);
    ensureAnalyser();
    actx?.resume();
    audio.play();
    if (tracks[i].id !== asset.id) history.replaceState(null, '', `#/${game.id}/${tracks[i].id}`);
  };
  const next = () => {
    if (!tracks.length) return;
    if (shuffle) return play(drawShuffled());
    play((cur + 1) % tracks.length);
  };

  playBtn.onclick = () => {
    if (cur < 0) return play(Math.max(0, tracks.findIndex((t) => t.id === asset.id)));
    if (audio.paused) { actx?.resume(); audio.play(); } else audio.pause();
  };
  bar.querySelector('.b-rep').onclick = (e) => { repeat = !repeat; e.currentTarget.classList.toggle('on', repeat); };
  bar.querySelector('.b-shuf').onclick = (e) => { shuffle = !shuffle; e.currentTarget.classList.toggle('on', shuffle); };
  seek.oninput = () => { if (audio.duration) audio.currentTime = (seek.value / 1000) * audio.duration; };
  audio.onended = () => (repeat && !shuffle ? play(cur) : next());
  audio.onplay = () => playBtn.innerHTML = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 5h4v14H6zM14 5h4v14h-4z"/></svg>';
  audio.onpause = () => playBtn.innerHTML = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>';

  let raf = 0;
  const loop = () => {
    raf = requestAnimationFrame(loop);
    if (audio.duration && document.activeElement !== seek) {
      seek.value = (audio.currentTime / audio.duration) * 1000;
      timeEl.textContent = `${fmtTime(audio.currentTime)} / ${fmtTime(audio.duration)}`;
    }
    if (analyser) {
      analyser.getByteFrequencyData(freq);
      const bars = vu.children;
      for (let i = 0; i < bars.length; i++) {
        const v = freq[Math.floor((i / bars.length) * freq.length)] / 255;
        bars[i].style.height = `${Math.max(6, v * 100)}%`;
      }
    }
  };
  loop();

  // auto-select the routed track (no autoplay — browsers need a gesture)
  const idx = tracks.findIndex((t) => t.id === asset.id);
  if (idx >= 0) rows[idx].classList.add('playing'), cur = idx;

  return {
    unmount() {
      cancelAnimationFrame(raf);
      audio.pause();
      audio.src = '';
      actx?.close();
      wrap.remove();
    },
    stats: () => ({ Tracks: tracks.length }),
  };
}

// ---- sfx -------------------------------------------------------------------------

function mountSfx(ctx) {
  const { stage, game } = ctx;
  const wrap = document.createElement('div');
  wrap.className = 'media-wrap';
  const inner = document.createElement('div');
  inner.className = 'media-inner';
  wrap.appendChild(inner);
  stage.appendChild(wrap);
  const fx = game.assets('sfx');
  inner.innerHTML = `<h2 style="margin:0">Sound effects</h2><div style="color:var(--dim)">${fx.length} effects — tap to play</div>`;
  const board = document.createElement('div');
  board.className = 'board';
  for (const s of fx) {
    const b = document.createElement('button');
    b.textContent = s.name;
    if (s.description) b.title = s.description;
    b.onclick = () => {
      const a = new Audio(game.url('', s.file));
      a.play();
      b.classList.add('hit');
      setTimeout(() => b.classList.remove('hit'), 250);
    };
    board.appendChild(b);
  }
  inner.appendChild(board);
  return { unmount() { wrap.remove(); } };
}

// ---- picture ----------------------------------------------------------------------

function mountPicture(ctx) {
  const { stage, game, asset } = ctx;
  const box = document.createElement('div');
  box.className = 'picture-stage';
  stage.appendChild(box);
  // Drawn on a canvas (not an <img> + CSS transform) so the screen filter can
  // capture the view and phase-lock its grid to the picture's pixels.
  const cv = document.createElement('canvas');
  cv.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;';
  box.appendChild(cv);
  const c2 = cv.getContext('2d');
  const dpr = Math.min(2, devicePixelRatio || 1);

  const img = new Image();
  img.src = game.url('', asset.file);

  // Exports may bake an integer upscale (Elite's loading.png is the 320-wide
  // screen at 2x): one GAME pixel is srcScale image pixels, and the filter's
  // cell grid must follow game pixels.
  let srcScale = 1;

  let z = 1, x = 40, y = 40;
  const draw = () => {
    const w = Math.max(1, Math.round(box.clientWidth * dpr));
    const h = Math.max(1, Math.round(box.clientHeight * dpr));
    if (cv.width !== w || cv.height !== h) { cv.width = w; cv.height = h; }
    c2.setTransform(1, 0, 0, 1, 0, 0);
    c2.clearRect(0, 0, w, h);
    c2.imageSmoothingEnabled = false;
    c2.setTransform(z * dpr, 0, 0, z * dpr, x * dpr, y * dpr);
    if (img.complete && img.naturalWidth) c2.drawImage(img, 0, 0);
  };
  img.onload = () => {
    const nw = game.display.native?.w || 0;
    const ratio = nw > 0 ? img.naturalWidth / nw : 1;
    srcScale = ratio >= 1 && Number.isInteger(ratio) ? ratio : 1;
    const fit = Math.min((box.clientWidth * 0.85) / img.naturalWidth, (box.clientHeight * 0.85) / img.naturalHeight);
    z = Math.max(1, Math.floor(fit)) || fit;
    x = (box.clientWidth - img.naturalWidth * z) / 2;
    y = (box.clientHeight - img.naturalHeight * z) / 2;
    draw();
  };
  const ro = new ResizeObserver(draw);
  ro.observe(box);

  const pts = new Map();
  let pinch = 0;
  box.addEventListener('pointerdown', (e) => { pts.set(e.pointerId, { x: e.clientX, y: e.clientY }); box.setPointerCapture(e.pointerId); });
  box.addEventListener('pointermove', (e) => {
    const p = pts.get(e.pointerId);
    if (!p) return;
    const dx = e.clientX - p.x, dy = e.clientY - p.y;
    p.x = e.clientX; p.y = e.clientY;
    if (pts.size === 2) {
      const [a, b] = [...pts.values()];
      const d = Math.hypot(a.x - b.x, a.y - b.y);
      if (pinch) zoomAt((a.x + b.x) / 2, (a.y + b.y) / 2, d / pinch);
      pinch = d;
    } else { x += dx; y += dy; draw(); }
  });
  const up = (e) => { pts.delete(e.pointerId); if (pts.size < 2) pinch = 0; };
  box.addEventListener('pointerup', up);
  box.addEventListener('pointercancel', up);
  box.addEventListener('wheel', (e) => {
    e.preventDefault();
    // gentle: one wheel notch (deltaY 100) = 1.2^0.1 ≈ 1.8%
    zoomAt(e.clientX, e.clientY, Math.pow(1.2, -e.deltaY * 0.001));
  }, { passive: false });
  function zoomAt(cx, cy, f) {
    const r = box.getBoundingClientRect();
    const px = cx - r.left, py = cy - r.top;
    const nz = Math.max(0.1, Math.min(40, z * f));
    x = px - ((px - x) / z) * nz;
    y = py - ((py - y) / z) * nz;
    z = nz;
    draw();
  }

  return {
    unmount() { ro.disconnect(); box.remove(); },
    sources: () => [cv],
    pixelGrid: () => ({
      cell: z * srcScale, ox: x, oy: y,
      ref: box.clientWidth,
    }),
    stats: () => ({ File: asset.file, Size: `${asset.w || '?'} × ${asset.h || '?'} px` }),
  };
}

// ---- video ------------------------------------------------------------------------

function mountVideo(ctx) {
  const { stage, game, asset } = ctx;
  const v = document.createElement('video');
  v.src = game.url('', asset.file);
  v.controls = true;
  v.playsInline = true;
  v.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;background:#000;object-fit:contain;';
  stage.appendChild(v);
  return {
    unmount() { v.pause(); v.remove(); },
    stats: () => ({ Size: `${asset.w || '?'} × ${asset.h || '?'}`, FPS: asset.fps || '?', Duration: asset.duration ? fmtTime(asset.duration) : '?' }),
  };
}

function escape(s) { return String(s).replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c])); }
