import {createReplay} from './replay.js';
import {pixelCoordinates,createInspector} from './inspector.js';
import {platforms} from './platforms.js';
const $ = id => document.getElementById(id),
      platform = document.body.dataset.platform, config = platforms[platform],
      canvas = $('screen'), ctx = canvas.getContext('2d');
let worker, session = 0, request = 0, selected = [], firmware = null,
            compatibility = true, loaded = false;
let lastSeconds = 0, lastSteps = null, lastFrames = 0, lastTime = performance.now(), rate = '',
    lastProfile = {}, profileTime = 0;
let pendingFrame=null,presentationScheduled=false,presentedCount=0,presentedAt=performance.now(),presentedRate=0,lastCopyMs=0;
function present(m){
 if(canvas.width!==m.width)canvas.width=m.width;
 if(canvas.height!==m.height)canvas.height=m.height;
 const start=performance.now();
 ctx.putImageData(new ImageData(new Uint8ClampedArray(m.pixels),m.width,m.height),0,0);
 lastCopyMs=performance.now()-start;
 presentedCount++;const elapsed=performance.now()-presentedAt;
 if(elapsed>=1000){presentedRate=presentedCount*1000/elapsed;presentedCount=0;presentedAt=performance.now();}
}
function queuePresentation(m){
 if(!m.running){pendingFrame=null;present(m);return;}
 pendingFrame=m;
 if(presentationScheduled)return;
 presentationScheduled=true;
 requestAnimationFrame(()=>{presentationScheduled=false;const next=pendingFrame;pendingFrame=null;if(next)present(next);});
}
const sources = new Map(), c64Keys = new Set();
let lastInput = '', gamepadLabel = '';
let inputTouch={x:0,y:0,down:false};
$('files').accept = config.accept;
$('help').textContent = config.help;
$('compat').textContent = config.compat;
$('tape').hidden = platform !== 'c64';
const send = (type, data = {}) => {const id=++request;worker?.postMessage({type,session,request:id,...data});return id;};
const replay=createReplay({canvas,send});
const inspector=createInspector({platform,canvas,send,jump:step=>replay.seek(step)});
function controls(on) {
  for (const id of ['run', 'pause', 'reset', 'step', 'save'])
    $(id).disabled = !on;
}
function showProfile(p, captureWork=false) {
  if (!p || (!captureWork&&performance.now() - profileTime < 500))
    return;
  profileTime = performance.now();
  const rows = p.buckets.map(b => {
    const ms = captureWork?b.ms:Math.max(0, b.ms - (lastProfile[b.name] || 0));
    if(!captureWork)lastProfile[b.name] = b.ms;
    return {...b, ms};
  }),
        total = rows.reduce((n, b) => n + b.ms, 0);
  if (!total)
    return;
  $('profile').replaceChildren(...rows.map(b => {
    const tr = document.createElement('tr');
    for (const v
             of [b.name, b.ms.toFixed(2), (100 * b.ms / total).toFixed(1)]) {
      const td = document.createElement('td');
      td.textContent = v;
      tr.append(td);
    }
    return tr;
  }));
  $('profile-note').textContent =
      p.sampled
          ? (captureWork?'Capture work. ':'')+'Sampled chip ticks (1 in 1,021). Sample-only milliseconds and exclusive shares; timer overhead affects tiny samples.'
          : (captureWork?'Capture work only. ':'')+'Exclusive wall time since the previous update. CPU remainder includes uninstrumented devices and scheduling. Idle time and display copies excluded.';
}
let pendingWorker;
function load(stateFile=null) {
  if(stateFile instanceof Event)stateFile=null;
  if (!selected.length) {
    $('status').textContent = 'Select a game image first.';
    return;
  }
  pendingWorker?.terminate();
  send('hold');
  const previousWorker=worker, previousLoaded=loaded, nextSession=session+1;
  const candidate=new Worker(new URL('./worker.js',import.meta.url),{type:'module'});pendingWorker=candidate;
  let bufferedState;
  candidate.onmessage=({data:m})=>{
    if(pendingWorker!==candidate||m.session!==nextSession)return;
    if(m.type==='state'){bufferedState=m;return;}
    if(m.type==='ready'){
      previousWorker?.terminate();worker=candidate;session=nextSession;pendingWorker=null;
      candidate.onmessage=handleMessage;loaded=true;
      handleMessage({data:m});if(bufferedState)handleMessage({data:bufferedState});
    }else if(m.type==='error'){
      candidate.terminate();pendingWorker=null;loaded=previousLoaded;controls(loaded);$("pause").disabled=true;
      $('status').textContent=m.text+' Current machine retained.';
    }else if(m.type==='message')$('status').textContent=m.text;
  };
  candidate.onerror=e=>{candidate.terminate();pendingWorker=null;loaded=previousLoaded;controls(loaded);$("pause").disabled=true;$('status').textContent='State/load worker failed: '+e.message+'. Current machine retained.';};
  inspector.reset();replay.reset();sources.clear();
  c64Keys.clear();
  lastInput = '';
  loaded = false;
  controls(false);
  lastProfile = {};
  profileTime = 0;
  rate = '';
  lastSteps=null;lastFrames=0;pendingFrame=null;presentedCount=0;presentedAt=performance.now();presentedRate=0;
  $('profile').replaceChildren();
  $('profile-note').textContent='Run the machine to measure subsystem timings.';
  $('status').textContent = 'Loading local image…';
  function handleMessage({data : m}) {
    if(pendingWorker)return;
    if (m.session !== session)
      return;
    if (m.type === 'state') {
      const s = m.state;
      $('help').textContent =
          config.help +
          (m.inputDeferred
               ? ' Known-image profile: controls are queued until intro initialization finishes at display 300.'
               : '');
      queuePresentation(m);
      const copyMs=lastCopyMs;
      if(lastSteps===null){lastSteps=s.steps;lastFrames=s.frames;lastSeconds=s.seconds??0;lastTime=performance.now();}
      const now = performance.now(), dt = (now - lastTime) / 1000;
      if (dt > .5) {
        const ratio = platform === 'c64'
                          ? (s.steps - lastSteps) / 985248 / dt
                          : s.seconds!==undefined ? (s.seconds-lastSeconds)/dt : (s.frames - lastFrames) / config.hz / dt;
        rate = `${(ratio * 100).toFixed(0)}% ${
            platform === 'c64' ? 'PAL speed' : 'nominal display rate'} · ${
            ((s.frames - lastFrames) / dt).toFixed(1)} ${
            platform === 'ps1' || platform === 'n64' ? 'synthetic fields'
                                                     : 'display updates'}/s · ${
            ((s.steps - lastSteps) / dt / 1e6).toFixed(2)}M ${
            platform === 'c64' ? 'cycles' : 'steps'}/s`;
        lastTime = now;
        lastFrames = s.frames;
        lastSeconds = s.seconds??0;
        lastSteps = s.steps;
      }
      $('metrics').textContent = `${m.running ? 'Running' : 'Paused'} · ${
          rate||'no execution yet'} · ${presentedRate.toFixed(1)} presented frames/s · display ${s.frames.toLocaleString()} · ${
          (m.heap / 1048576).toFixed(0)} MiB WASM · longest execution slice ${
          m.maxCall.toFixed(1)} ms · canvas copy ${copyMs.toFixed(1)} ms`;
      showProfile(m.profile);
      if (loaded) {
        $('reset').disabled = !!m.saving;
        $('run').disabled = m.running||m.capturing||m.saving;
        $('step').disabled = m.running||m.capturing||m.saving;
        $('save').disabled = m.capturing||m.saving;
        $('pause').disabled = !m.running&&!m.capturing;
        $('cancelcapture').hidden=!m.capturing;
      }
    } else if(m.type==='capture-progress'){
      $('status').textContent=m.text;$('cancelcapture').hidden=false;
      $('run').disabled=$('step').disabled=$('save').disabled=true;
    } else if(m.type==='seek'||m.type==='seek-progress'){replay.result(m);
    } else if(m.type==='pixel'||m.type==='source'||m.type==='resource'){inspector.result(m);
    } else if(m.type==='capture-cleared'){
      inspector.reset();replay.reset();
      $('capture-note').textContent='Pause to record the next complete display interval.';$('cancelcapture').hidden=true;
    } else if(m.type==='capture'){
      inspector.setCapture(m);replay.setCapture(m);showProfile(m.profile,true);lastTime=performance.now();lastSteps=m.end.steps;lastFrames=m.end.frames;lastSeconds=m.end.seconds??0;rate='';
      $('cancelcapture').hidden=true;
      $('capture-note').textContent=`Captured display ${m.start.frames}–${m.end.frames} · ${(m.elapsedMs/1000).toFixed(2)} s capture · longest call ${m.maxCall.toFixed(1)} ms · ${((m.info.bytes+m.checkpointBytes)/1048576).toFixed(1)} MiB evidence/checkpoints${m.info.overflow?' · incomplete: trace limit reached':''}`;
    } else if(m.type==='saved'){
      const url=URL.createObjectURL(new Blob([m.bytes],{type:'application/octet-stream'}));
      const a=document.createElement('a');a.href=url;a.download=platform+'-'+Date.now()+'.rrstate';a.textContent='Download state';$('status').replaceChildren(document.createTextNode('State ready. Machine paused. '),a);a.click();setTimeout(()=>URL.revokeObjectURL(url),300000);
    } else if (m.type === 'slow') {
      console.warn('Long emulation slice', JSON.stringify(m));
    } else if (m.type === 'ready') {
      loaded = true;
      controls(true);
      $('pause').disabled = true;
      $('status').textContent = m.text;
      lastTime = performance.now();
      send('turbo', {value : $('turbo').checked});
    } else if (m.type === 'message' || m.type === 'error') {
      $('status').textContent = m.text;
      if (m.type === 'error') {
        controls(loaded);
        $('pause').disabled = true;
      }
    }
  };
  candidate.postMessage({type:'load',session:nextSession,platform,files:selected,firmware,compatibility,stateFile});
}
$('load').onclick = () => {
  selected = [...$('files').files ];
  compatibility = $('compatprofile')?.checked ?? true;
  firmware = platform === 'c64'
                 ? [ 'basic', 'kernal', 'chargen' ].map(id => $(id).files[0])
                 : null;
  load();
};
$('reset').onclick = ()=>load();
$('cancelcapture').onclick=()=>send('cancel-capture');
$('save').onclick=()=>{release();controls(false);$('status').textContent='Saving state…';send('save');};
$('statefile').onchange=()=>{const f=$('statefile').files[0];if(!f)return;if(!selected.length){selected=[...$('files').files];firmware=platform==='c64'?['basic','kernal','chargen'].map(id=>$(id).files[0]):null;compatibility=$('compatprofile')?.checked??true;}load(f);$('statefile').value='';};
for (const id of ['run', 'pause', 'step'])
  $(id).onclick = () => {
    $('status').textContent =
        id === 'pause' ? 'Preparing a complete-frame capture…'
        : id === 'run' ? 'Running local image.'
                       : 'Advancing one display boundary…';
    if(id==='run'||id==='step'){inspector.reset();replay.reset();release();lastSteps=null;presentedCount=0;presentedAt=performance.now();}
    send(id);
    if (id === 'run')
      canvas.focus();
  };
$('turbo').onchange = () => send('turbo', {value : $('turbo').checked});
$('fullscreen').onclick = () => canvas.requestFullscreen();
$('tapeplay').onclick = () => send('tape', {down : 1});
$('tapestop').onclick = () => send('tape', {down : 0});
function input(keys = []) {
  if(inspector.isInspecting())return;
  let buttons = 0, x = 0, y = 0;
  for (const {bits = [], ax = 0, ay = 0} of sources.values()) {
    x += ax;
    y += ay;
    for (const v of bits) {
      if (typeof v === 'number')
        buttons |= v;
      else if (v === 'left')
        x -= 80;
      else if (v === 'right')
        x += 80;
      else if (v === 'up')
        y += 80;
      else if (v === 'down')
        y -= 80;
    }
  }
  const state = {
    ...((platform==='ds'||platform==='3ds')?{touch:inputTouch}:{}),
    buttons : buttons >>> 0,
    x : Math.max(-80, Math.min(80, x)),
    y : Math.max(-80, Math.min(80, y))
  },
        key = JSON.stringify(state);
  if (key !== lastInput || keys.length) {
    lastInput = key;
    send('input', {...state, keys});
  }
}
for (const [label, bit] of config.buttons) {
  const b = document.createElement('button');
  b.textContent = label;
  b.onpointerdown = e => {
    e.preventDefault();
    b.setPointerCapture(e.pointerId);
    sources.set('touch:' + e.pointerId, {bits : [ bit ]});
    input();
  };
  b.onpointerup = b.onpointercancel = b.onlostpointercapture = e => {
    sources.delete('touch:' + e.pointerId);
    input();
  };
  b.onkeydown = e => {
    if (e.key === ' ' || e.key === 'Enter') {
      e.preventDefault();
      sources.set('button:' + label, {bits : [ bit ]});
      input();
    }
  };
  b.onkeyup = b.onblur = () => {
    sources.delete('button:' + label);
    input();
  };
  $('pad').append(b);
}
for (const down of [true, false])
  canvas.addEventListener(down ? 'keydown' : 'keyup', e => {
    if(inspector.isInspecting())return;
    if (!loaded || e.repeat || e.metaKey)
      return;
    const bit = config.keys[e.key] ?? config.keys[e.key.toLowerCase()];
    if (bit !== undefined) {
      e.preventDefault();
      if (down)
        sources.set('key:' + e.code, {bits : [ bit ]});
      else
        sources.delete('key:' + e.code);
      if (platform === 'c64' && e.key === ' ') {
        if (down)
          c64Keys.add(32);
        else
          c64Keys.delete(32);
        input([ [ 32, +down ] ]);
      } else
        input();
    } else if (platform === 'c64') {
      const special = {
        Enter : 13,
        Backspace : 1,
        Insert : 16,
        Home : e.shiftKey ? 2 : 12,
        Escape : e.shiftKey ? 7 : 3,
        PageUp : 255,
        Control : 14,
        Alt : 15
      };
      const code = special[e.key] ??
                   (/^F[1-8]$/.test(e.key) ? 240 + Number(e.key.slice(1))
                    : e.key.length === 1   ? e.key.toUpperCase().charCodeAt(0)
                                           : 0);
      if (code && code < 256) {
        e.preventDefault();
        if (down)
          c64Keys.add(code);
        else
          c64Keys.delete(code);
        input([ [ code, +down ] ]);
      }
    }
  });
function release() {
  inputTouch={...inputTouch,down:false};
  sources.clear();
  input([...c64Keys ].map(code => [code, 0]));
  c64Keys.clear();
}
canvas.onblur = release;
window.addEventListener('blur', release);
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    release();
    send('hold');
  }
});
if(platform==='ds'||platform==='3ds'){
 let stylus=null;const screenHeight=platform==='ds'?192:240,screenWidth=platform==='ds'?256:320,left=platform==='ds'?0:40;
 const point=e=>{const p=pixelCoordinates(canvas.getBoundingClientRect(),canvas.width,canvas.height,e.clientX,e.clientY);return p&&p.y>=screenHeight&&p.x>=left&&p.x<left+screenWidth?{x:p.x-left,y:p.y-screenHeight}:null;};
 const pen=(e,down)=>{if(!loaded||inspector.isInspecting())return;const p=point(e);if(down&&!p)return;e.preventDefault();inputTouch={x:p?.x??inputTouch.x,y:p?.y??inputTouch.y,down};input();};
 canvas.addEventListener('pointerdown',e=>{if(!point(e)||inspector.isInspecting())return;stylus=e.pointerId;canvas.setPointerCapture(stylus);canvas.focus();pen(e,true);});
 canvas.addEventListener('pointermove',e=>{if(e.pointerId===stylus)pen(e,true);});
 for(const type of ['pointerup','pointercancel','lostpointercapture'])canvas.addEventListener(type,e=>{if(e.pointerId===stylus){pen(e,false);stylus=null;}});
}
const gamepadMaps = {
 psp:{0:16384,1:8192,2:32768,3:4096,4:256,5:512,8:1,9:8,12:16,13:64,14:128,15:32},
 '3ds':{0:1,1:2,2:2048,3:1024,4:512,5:256,8:4,9:8,12:64,13:128,14:32,15:16},
 ds:{0:1,1:2,2:2048,3:1024,4:512,5:256,8:4,9:8,12:64,13:128,14:32,15:16},
  c64 : {0 : 16, 12 : 1, 13 : 2, 14 : 4, 15 : 8},
  ps1 : {
    0 : 16384,
    1 : 8192,
    2 : 32768,
    3 : 4096,
    4 : 1024,
    5 : 2048,
    6 : 256,
    7 : 512,
    8 : 1,
    9 : 8,
    12 : 16,
    13 : 64,
    14 : 128,
    15 : 32
  },
  n64 : {
    0 : 32768,
    1 : 16384,
    4 : 32,
    5 : 16,
    6 : 8192,
    9 : 4096,
    12 : 2048,
    13 : 1024,
    14 : 512,
    15 : 256
  },
  '3do' : {
    0 : 0x08000000,
    1 : 0x04000000,
    2 : 0x02000000,
    3 : 0x00800000,
    4 : 0x00200000,
    5 : 0x00400000,
    9 : 0x01000000,
    12 : 0x40000000,
    13 : 0x80000000,
    14 : 0x10000000,
    15 : 0x20000000
  }
};
const axis = v =>
    Math.abs(v || 0) < .18
        ? 0
        : Math.round(Math.sign(v) * (Math.abs(v) - .18) / .82 * 80);
function pollPad() {
  const pads = [...(navigator.getGamepads?.() || []) ].filter(Boolean),
        p = pads.find(p => p.mapping === 'standard');
  const allowed = loaded && !inspector.isInspecting() && document.hasFocus() && !document.hidden &&
                  (document.activeElement === canvas ||
                   $('pad').contains(document.activeElement));
  for (const key of sources.keys())
    if (key.startsWith('gamepad:'))
      sources.delete(key);
  if (p && allowed) {
    const bits = Object.entries(gamepadMaps[platform])
                     .filter(([ i ]) => p.buttons[i]?.pressed)
                     .map(([, v ]) => v),
          x = axis(p.axes[0]), y = -axis(p.axes[1]);
    if (platform === 'n64') {
      if (axis(p.axes[2]) < -30)
        bits.push(2);
      if (axis(p.axes[2]) > 30)
        bits.push(1);
      if (axis(p.axes[3]) < -30)
        bits.push(8);
      if (axis(p.axes[3]) > 30)
        bits.push(4);
      sources.set('gamepad:' + p.index, {bits, ax : x, ay : y});
    } else if(platform==='3ds'||platform==='psp'){
      sources.set('gamepad:'+p.index,{bits,ax:x,ay:y});
    } else {
      const directions =
          platform === 'ds' ? [32,16,64,128] : platform === 'ps1' ? [ 128, 32, 16, 64 ]
          : platform === 'c64'
              ? [ 4, 8, 1, 2 ]
              : [ 0x10000000, 0x20000000, 0x40000000, 0x80000000 ];
      if (x < -30)
        bits.push(directions[0]);
      if (x > 30)
        bits.push(directions[1]);
      if (y > 30)
        bits.push(directions[2]);
      if (y < -30)
        bits.push(directions[3]);
      sources.set('gamepad:' + p.index, {bits});
    }
  }
  const label =
      p ? `${p.id} · standard mapping · left stick/D-pad movement · ${
              platform === 'n64' ? 'right stick C buttons · '
                                 : ''}18% stick dead zone · ${
              allowed ? 'active' : 'click display to enable'}`
      : pads.length
          ? 'Non-standard gamepad: use keyboard/on-screen controls.'
          : 'No gamepad connected · keyboard and on-screen controls available';
  if (label !== gamepadLabel) {
    $('device').textContent = label;
    gamepadLabel = label;
  }
  if (loaded)
    input();
  requestAnimationFrame(pollPad);
}
requestAnimationFrame(pollPad);
