export function decodedColor(platform, value, size, format) {
  if(platform==='3ds'){
    if(format===0)return [value>>>24,value>>>16&255,value>>>8&255,255];
    if(format===1)return [value>>>16&255,value>>>8&255,value&255,255];
    const expand5=n=>(n<<3)|(n>>>2);
    if(format===2)return [expand5(value>>>11&31),((value>>>5&63)<<2)|(value>>>9&3),expand5(value&31),255];
    if(format===3)return [expand5(value>>>11&31),expand5(value>>>6&31),expand5(value>>>1&31),255];
    if(format===4)return [(value>>>12&15)*17,(value>>>8&15)*17,(value>>>4&15)*17,255];
    return null;
  }
  if(platform==='ds')return [value&255,value>>>8&255,value>>>16&255,value>>>24];
  if (platform === 'ps1')
    return [
      (value & 31) << 3, ((value >>> 5) & 31) << 3, ((value >>> 10) & 31) << 3,
      255
    ];
  const v = ((value & 255) << 8) | ((value >>> 8) & 255);
  if (platform === '3do')
    return [ v >>> 10 & 31, v >>> 5 & 31, v & 31 ]
        .map(n => (n << 3) | (n >>> 2))
        .concat(255);
  if (platform === 'n64')
    return size===2?[(v>>>11&31)<<3,(v>>>6&31)<<3,(v>>>1&31)<<3,255]:[value&255,value>>>8&255,value>>>16&255,255];
  return null;
}
export function pixelCoordinates(rect, width, height, clientX, clientY) {
  const scale = Math.min(rect.width / width, rect.height / height);
  const x = (clientX - rect.left - (rect.width - width * scale) / 2) / scale;
  const y = (clientY - rect.top - (rect.height - height * scale) / 2) / scale;
  if (x < 0 || y < 0 || x >= width || y >= height)
    return null;
  return {x : Math.floor(x), y : Math.floor(y)};
}
const hex = (n, width = 8) =>
    '0x' + (Number(n) >>> 0).toString(16).padStart(width, '0');
export function createInspector({platform, canvas, send, jump}) {
  const $ = id => document.getElementById(id);
  let capture = null, pixels = null, latest = 0, current = null,
      selected = null, rows = [];
  function reset() {
    capture = null;
    pixels = null;
    latest = 0;
    current = null;
    selected = null;
    rows = [];
    $('pixel-summary').textContent = 'Pause the emulator, then select a pixel.';
    $('contributors').replaceChildren();
    $('event-detail').replaceChildren();
    $('source-detail').replaceChildren();
    $('inspect-pixel').disabled = true;
    $('screen').classList.remove('inspectable');
  }
  function setCapture(c) {
    capture = c;
    $('inspect-pixel').disabled = false;
    $('screen').classList.add('inspectable');
    $('pixel-summary').textContent = 'Select a pixel in the captured display.';
    pixels = canvas.getContext('2d')
                 .getImageData(0, 0, canvas.width, canvas.height)
                 .data;
    $('pixel-x').max = c.width - 1;
    $('pixel-y').max = c.height - 1;
  }
  function pending(el, message) {
    el.textContent = message;
    const cancel = document.createElement('button');
    cancel.textContent = 'Cancel query';
    cancel.onclick = () => { latest = 0; el.textContent = 'Query cancelled.'; };
    el.append(cancel);
  }
  function query(x, y) {
    if (!capture || !Number.isInteger(x) || !Number.isInteger(y) || x < 0 ||
        y < 0 || x >= capture.width || y >= capture.height)
      return;
    current = {x, y};
    $('pixel-x').value = x;
    $('pixel-y').value = y;
    pending($('pixel-summary'), `Inspecting (${x}, ${y})…`);
    latest = send('pixel', {capture : capture.id, x, y});
  }
  canvas.addEventListener('click', e => {
    if (!capture)
      return;
    const p = pixelCoordinates(canvas.getBoundingClientRect(), capture.width,
                               capture.height, e.clientX, e.clientY);
    if (p)
      query(p.x, p.y);
  });
  $('inspect-pixel').onclick = () =>
      query(Number($('pixel-x').value), Number($('pixel-y').value));
  function line(parent, label, value) {
    const p = document.createElement('p');
    const strong = document.createElement('strong');
    strong.textContent = label + ': ';
    p.append(strong, document.createTextNode(String(value)));
    parent.append(p);
  }
  function sourceButton(parent, label, address, before, expected, size = 2) {
    const b = document.createElement('button');
    b.textContent = label;
    b.onclick = () => {
      const id =
          send('source',
               {capture : capture.id, address, size, before, expected});
      pending($('source-detail'), 'Reading historical source…');
      latest = id;
    };
    parent.append(b);
  }
  function snapshotButton(parent, c, offset, label) {
    const b = document.createElement('button');
    b.textContent = label;
    b.onclick = () => {
      latest =
          send('resource',
               {capture : capture.id, resource : c.sourceSnapshot, offset});
      pending($('source-detail'), 'Reading captured source bytes…');
    };
    parent.append(b);
  }
  function showEvent(c, index) {
    latest = 0;
    selected = c;
    $('source-detail').replaceChildren();
    for (const b of $('contributors').children)
      b.setAttribute('aria-pressed', String(Number(b.dataset.index) === index));
    const el = $('event-detail');
    el.replaceChildren();
    if(c.replayStep!==undefined&&jump){const b=document.createElement('button');b.textContent='Show rendering step';b.onclick=()=>jump(c.replayStep);el.append(b);}
    if (platform === 'c64') {
      line(el, 'Role', c.role);
      if (c.missing) {
        line(el, 'Evidence', 'Not observed in this capture');
        return;
      }
      line(el, 'Fetched byte', `${hex(c.address, 4)} = ${hex(c.value, 2)}`);
      line(el, 'Fetch cycle', c.fetchCycle);
      if (c.writer) {
        line(el, 'Writer',
             `PC ${hex(c.writer.pc, 4)}, cycle ${c.writer.cycle}`);
        line(el, 'Byte change',
             `${hex(c.writer.oldValue, 2)} → ${hex(c.writer.value, 2)}`);
        line(el, 'Instruction bytes',
             c.writer.code.map(n => hex(n, 2)).join(' '));
        if (c.writer.nearbyRead?.cycle)
          line(el, 'Nearby read',
               `${
                   hex(c.writer.nearbyRead.address,
                       4)} — proximity is not a proven data dependency`);
      } else
        line(el, 'Writer',
             c.space === 3 ? 'Immutable ROM'
                           : 'No writer recorded before this fetch');
      return;
    }
    line(el, 'Event', c.command?.kind || 'Memory write');
    line(el, 'Memory', `${hex(c.address)} (${c.size} bytes)`);
    line(el, c.drawn ? 'Stored bytes' : 'Rejected candidate',
         c.drawn ? `${hex(c.before)} → ${hex(c.after)}`
         : c.depthRejected ? 'Depth test'
                           : 'Alpha / transparent texel');
    line(el, c.event ? 'Submission PC' : 'Writer PC',
         hex(c.submissionPC ?? c.pc));
    line(el, 'Clock', c.clock || c.submissionClock || 0);
    if (c.event)
      line(
          el, 'Origin limit',
          'Submission does not identify the instruction that built the command data.');
    if (c.command?.pixc !== undefined)
      line(el, 'PIXC', hex(c.command.pixc));
    if (c.command?.otherModes)
      line(el, 'RDP blend/depth modes', '0x' + c.command.otherModes);
    if (c.command?.hasSource||c.command?.kind==='Cel / CCB') {
      line(el, 'Sampled texel', hex(c.texel));
      line(el, 'Texture coordinate', `${c.u}, ${c.v}`);
    }
    if (platform === 'ps1' && c.event && c.command?.hasSource) {
      line(el, 'Texture/copy source', hex(c.sourceAddress));
      sourceButton(el, 'Follow source history', c.sourceAddress, c.sourceBefore,
                   c.sourceValue);
      if (c.paletteAddress) {
        line(el, 'CLUT entry', hex(c.paletteAddress));
        sourceButton(el, 'Follow palette history', c.paletteAddress,
                     c.paletteBefore, c.paletteValue);
      }
    }
    if ((platform === 'ds' || platform === '3ds') && c.command?.hasSource) {
      line(el, 'Captured source surface', hex(c.sourceAddress));
      sourceButton(el, 'Follow source history', c.sourceAddress, c.sourceBefore, c.sourceValue, 4);
    }
    if (platform === '3do' && c.command?.kind === 'Cel / CCB') {
      line(el, 'CCB', hex(c.command.ccb));
      line(el, 'Source',
           `${hex(c.command.source)}${
               c.command.lrform ? ' · interleaved framebuffer' : ''}`);
      if (c.sourceAddress >= 0x200000 && c.sourceAddress < 0x300000)
        sourceButton(el, 'Follow framebuffer source', c.sourceAddress,
                     c.sourceBefore, c.sourceValue);
      else if (c.sourceSnapshot)
        snapshotButton(el, c, 0, 'Inspect captured source bytes');
      if (c.paletteAddress)
        line(el, 'PLUT entry',
             `${hex(c.paletteAddress)} = ${
                 hex(c.paletteValue, 4)} (captured bytes)`);
    }
    if (platform === 'n64' && c.sourceSnapshot && c.command?.hasSource) {
      line(el, 'Texture image base', hex(c.command.textureAddress));
      line(el, 'TMEM tap',
           `${
               hex(c.sourceAddress,
                   3)} · last filter tap, not the entire filtered sample`);
      snapshotButton(el, c, c.sourceAddress, 'Inspect captured TMEM');
      if (c.paletteAddress)
        snapshotButton(el, c, c.paletteAddress, 'Inspect captured TLUT');
    }
    const details = document.createElement('details'),
          summary = document.createElement('summary'),
          pre = document.createElement('pre');
    summary.textContent = 'Recorded command';
    pre.textContent = JSON.stringify(c.command, null, 2);
    details.append(summary, pre);
    el.append(details);
  }
  function result(m) {
    if (!capture || m.capture !== capture.id || m.request !== latest)
      return;
    if (m.type === 'source' || m.type === 'resource') {
      const p = m.evidence, el = $('source-detail');
      el.replaceChildren();
      if (p.error) {
        line(el, 'Source', p.error);
        return;
      }
      if (m.type === 'resource') {
        line(el, 'Immutable source snapshot',
             `Offset ${hex(p.offset)} of ${p.size} bytes`);
        line(el, 'Bytes',
             p.bytes.map(n => n.toString(16).padStart(2, '0')).join(' '));
        return;
      }
      line(el, 'Historical source', hex(p.address));
      line(el, 'At the recorded read',
           `${hex(p.reconstructed)}${
               p.complete ? ' · reconstruction matches captured bytes'
                          : ' · history incomplete'}`);
      for (const c of p.contributors || [])
        line(el, c.drawn ? 'Earlier write' : 'Rejected candidate',
             `${c.command?.kind || 'write'} at PC ${
                 hex(c.submissionPC ??
                     c.pc)} · ${hex(c.before)} → ${hex(c.after)}`);
      if (!p.contributors?.length)
        line(el, 'Ancestry', 'Source contents predate this capture');
      return;
    }
    const p = m.evidence;
    const el = $('pixel-summary');
    el.replaceChildren();
    $('event-detail').replaceChildren();
    $('source-detail').replaceChildren();
    $('contributors').replaceChildren();
    const at = (m.y * capture.width + m.x) * 4,
          rgba = [...pixels.slice(at, at + 4) ],
          color = '#' + rgba.slice(0, 3)
                            .map(n => n.toString(16).padStart(2, '0'))
                            .join('');
    const swatch = document.createElement('span');
    swatch.className = 'swatch';
    swatch.style.background = color;
    el.append(swatch, document.createTextNode(`(${m.x}, ${m.y}) · ${color}`));
    if (p.error) {
      line(el, 'Evidence', p.error);
      return;
    }
    if (platform === 'c64') {
      line(el, 'VIC decision',
           p.border ? 'Border/background'
                    : `Graphics${
                          p.spriteMask
                              ? ' + sprite candidates ' + hex(p.spriteMask, 2)
                              : ''}`);
      line(el, 'Raster',
           `${p.raster}, cycle ${p.cycle}, palette index ${p.color}`);
      const reconstructed = p.rgba === undefined ? null : [
        p.rgba & 255, (p.rgba >>> 8) & 255, (p.rgba >>> 16) & 255, 255
      ];
      line(el, 'Color check',
           reconstructed && reconstructed.every((n, i) => n === rgba[i])
               ? 'Captured VIC palette decision matches the displayed color'
               : 'Color reconstruction unavailable or mismatched');
    } else {
      line(
          el, 'Evidence',
          p.blank ? 'Modeled scanout is blank'
          : p.complete
              ? 'Recorded writes reconstruct the final stored bytes'
              : 'Incomplete history: final bytes differ or the capture limit was reached');
      if (!p.blank) {
        const reconstructed = decodedColor(platform, p.reconstructed, p.size, p.displayFormat);
        line(
            el, 'Scanout color',
            reconstructed?.every((n, i) => n === rgba[i])
                ? 'Reconstructed stored bytes match the modeled display color'
                : 'Stored-byte reconstruction does not match the displayed pixel');
      }
      if (!p.blank)
        line(el, 'Starting buffer',
             `${hex(p.initial)} · content already present when capture began`);
    }
    if (p.overflow || p.truncated)
      line(el, 'Limit',
           `${p.overflow || 0} dropped records${
               p.truncated ? ' · contributor list truncated' : ''}`);
    rows = [...(p.contributors || []) ];
    if (platform === 'c64')
      rows.sort((a, b) => (a.fetchCycle || 0) - (b.fetchCycle || 0));
    rows.forEach((c, index) => {
      const b = document.createElement('button');
      b.type = 'button';
      b.dataset.index = index;
      b.setAttribute('aria-pressed', 'false');
      b.textContent = platform === 'c64'
                          ? c.role
                          : `${c.id}. ${c.command?.kind || 'Memory write'}${
                                c.drawn ? ''
                                : c.depthRejected
                                    ? ' · depth rejected'
                                    : ' · transparent/alpha rejected'}`;
      b.onclick = () => showEvent(c, index);
      $('contributors').append(b);
    });
    if (rows.length)
      {const lastWrite=platform==='c64'?-1:rows.findLastIndex(c=>c.drawn);const index=lastWrite>=0?lastWrite:rows.length-1;showEvent(rows[index],index);}
    else
      line(el, 'Contributors',
           p.blank ? 'No active framebuffer scanout'
                   : 'No writes to this pixel in the captured interval');
  }
  reset();
  return {reset, setCapture, result, isInspecting : () => !!capture};
}
