import {selectedC64Backend} from './core-backend.js';
import {createPreparedPanel} from './prepared-panel.js';
import {checkpointCache} from './checkpoint-cache.js';
import {createCodeWorkspace} from './code-workspace.js';
import {createViewportWorkspace} from './viewport-workspace.js';
import {createViewportPanels} from './viewport-panels.js';
import {createInspectionFeed} from './inspection-feed.js';
import {createMemoryWorkspace} from './memory-workspace.js';
import {mountShell} from './ui-shell.js';
import {createRenderWorkspace} from './render-workspace.js';
import {pixelCoordinates} from './inspector.js';
import {platforms} from './platforms.js';
import {dosFiles,dosKeys} from './dos-media.js';
import {amigaKeys} from './amiga-input.js';
export function mountEmulatorSession(system,{onSystem=()=>{}}={}){
let backend=system==='c64'?selectedC64Backend(location.search):'production',capabilities={};
let disposed=false,pendingSuspend=null,latestState=null,sampler=null;const lifetime=new AbortController();let paneSet;
document.body.dataset.platform=system;
const $ = id => document.getElementById(id),
      platform = system, config = platforms[platform],
      presentation = mountShell(platform),
      canvas = $('screen'), ctx = canvas.getContext('2d');
const renderTemplate=$('render-workspace').innerHTML;
let activeMedia=null;
let worker, session = 0, request = 0, selected = [], firmware = null, driveFirmware=null,
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
 memory?.present(canvas);code?.present(canvas);paneSet?.present();
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
function backendControls(){ $('files').accept=platform==='c64'&&backend==='owned'?'.tap,.d64,.g64':config.accept;if($('drive-firmware-group'))$('drive-firmware-group').hidden=backend!=='owned';}
backendControls();
$('help').textContent = config.help;
$('compat').textContent = config.compat;
$('tape').hidden = platform !== 'c64';
const send = (type, data = {}) => {const id=++request;worker?.postMessage({type,session,request:id,...data});return id;};
let memory,code,statePanel,panelLayout;
const feed=createInspectionFeed({send,platform});
const views=createViewportWorkspace({root:$('viewport-root'),navigation:$('workspace-nav'),platform,onChange:id=>{document.body.dataset.workspace=id;}});
views.register({id:'play',label:'Play',panel:$('play-workspace')});
const render=createRenderWorkspace({workspaces:views,platform,presentation,send,resume:()=>$('run').click(),playCanvas:canvas,beforeCapture:()=>memory.invalidate()});
memory=createMemoryWorkspace({root:$('memory-workspace'),views,send,transport:id=>transport(id,true),platform,onWriter:['c64','dos','3do'].includes(platform)?at=>{paneSet?.inspect(at);}:null});
if(['c64','dos','3do'].includes(platform)){
 statePanel={update:s=>paneSet?.snapshot(s),setKnowledge(){},setRunning(){},reset(){}};
}
code=createCodeWorkspace({root:$('code-workspace'),views,send,transport:id=>transport(id,true),platform,statePanel,panelLayout,onTourLayout:()=>views.tourLayout(),onInspect:at=>paneSet?.inspect(at)});
function controls(on) {
  for (const id of ['run', 'pause', 'reset', 'step', 'save'])
    $(id).disabled = !on;
  render.ready(on&&capabilities.renderCapture!==false);
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
let pendingWorker, cancelPreparation, preparedPanel, lessonShortcut;
function load(stateFile=null,preparedStart=null,requestedBackend=backend) {
  if(stateFile instanceof Event)stateFile=null;
  if (!selected.length) {
    $('status').textContent = 'Select a game image first.';
    return;
  }
  const requestedMedia={files:selected,firmware,driveFirmware,compatibility};
  paneSet?.media(selected);
  pendingWorker?.terminate();
  send('hold');
  const previousWorker=worker, previousLoaded=loaded, nextSession=session+1;
  const candidate=new Worker(new URL('./worker.js',import.meta.url),{type:'module'});pendingWorker=candidate;
  let bufferedState;
  const recover=text=>{candidate.terminate();pendingWorker=null;cancelPreparation=null;loaded=previousLoaded;if(activeMedia){selected=activeMedia.files;firmware=activeMedia.firmware;driveFirmware=activeMedia.driveFirmware;compatibility=activeMedia.compatibility;paneSet?.media(selected);paneSet?.mediaReady(selected);}controls(loaded);preparedPanel?.pending(false);$('system-select').disabled=false;if(loaded){memory.ready();code?.ready(session);paneSet?.ready(session);feed.ready();}$("pause").disabled=true;$('status').textContent=text+' Current machine retained.';preparedPanel?.message(text);};
  cancelPreparation=()=>recover('Preparation cancelled.');
  if(preparedStart){preparedPanel.pending(true);$('system-select').disabled=true;}else preparedPanel?.state({saving:true});

  candidate.onmessage=({data:m})=>{
    if(pendingWorker!==candidate||m.session!==nextSession)return;
    if(m.type==='prepared-cache-get'){checkpointCache('get',m.key).then(bytes=>{if(pendingWorker===candidate)candidate.postMessage({type:'prepared-cache-reply',session:nextSession,bytes});});return;}
    if(m.type==='prepared-cache-put'){checkpointCache('put',m.key,m.bytes);return;}
    if(m.type==='state'){bufferedState=m;return;}
    if(m.type==='ready'){
      previousWorker?.terminate();activeMedia=requestedMedia;worker=candidate;session=nextSession;pendingWorker=null;
      candidate.onmessage=handleMessage;loaded=true;cancelPreparation=null;preparedPanel?.pending(false);$('system-select').disabled=false;
      handleMessage({data:m});if(bufferedState)handleMessage({data:bufferedState});
      if(preparedStart){const r=preparedPanel.recipe(preparedStart);preparedPanel.message('Prepared start ready.');if(r.layout==='loader')views.tourLayout();else views.reveal('lesson');send(r.target.kind==='tour'?'tour-start':'experiment-prepare',{protocol:1,generation:session,id:r.target.id});}
    }else if(m.type==='error'){
      recover(m.text);
    }else if(m.type==='message')$('status').textContent=m.text;
  };
  candidate.onerror=e=>recover('State/load worker failed: '+e.message+'.');
  render.reset();memory.reset();code?.reset();paneSet?.reset();feed.reset();sources.clear();
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
    if(disposed||pendingWorker)return;
    if (m.session !== session)
      return;
    if(m.type.startsWith('drive-debug-')){paneSet?.driveResult(m);return;}
    if(m.type.startsWith('experiment-')){memory.invalidate();code?.experimentResult(m);if(m.snapshot)paneSet?.snapshot(m.snapshot);return;}
    if(m.type.startsWith('tour-')){if(m.phase==='running-to-stop')memory.invalidate();code?.tourResult(m);if(m.snapshot)paneSet?.snapshot(m.snapshot);return;}
    if(m.type==='debug-capabilities'){const count=preparedPanel?.ready(m.knowledge);if(lessonShortcut)lessonShortcut.hidden=!count;}
    if(m.type.startsWith('debug-')){if(m.type==='debug-started')memory.invalidate();code?.result(m);paneSet?.result(m);return;}
    if(m.type.startsWith('memory-')){memory.result(m);feed.result(m);return;}
    if (m.type === 'state') {
      const s = m.state;preparedPanel?.state(m);
      latestState=m;memory.state(m);code?.state(m);feed.state(m);paneSet?.state(m);$('system-select').disabled=!!(m.saving||m.memoryRecording||m.debugBusy||m.experimentOwned||m.capturing);
      if(platform==='amiga')canvas.classList.toggle('mouse-active',m.running);
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
            platform === 'dos' ? 'synthetic VGA intervals' : platform === 'ps1' || platform === 'n64' ? 'synthetic fields'
                                                     : 'display updates'}/s · ${
            ((s.steps - lastSteps) / dt / 1e6).toFixed(2)}M ${
            platform === 'c64' ? 'cycles' : 'steps'}/s`;
        if(platform==='dos')rate=rate.replace(/^[^·]+· /,'');
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
        $('reset').disabled = !!(m.saving||m.experimentOwned);
        render.ready(capabilities.renderCapture!==false&&!m.saving&&!m.memoryRecording&&!m.debugBusy&&!m.experimentOwned);
        $('run').disabled = m.running||m.capturing||m.saving||m.memoryRecording||m.debugBusy||m.experimentOwned;
        $('step').disabled = m.running||m.capturing||m.saving||m.memoryRecording||m.debugBusy||m.experimentOwned;
        $('save').disabled = m.capturing||m.saving||m.memoryRecording||m.debugBusy||m.experimentOwned;
        $('pause').disabled = !m.running&&!m.capturing&&!m.debugBusy&&!m.memoryRecording;
        $('cancelcapture').hidden=!m.capturing;
      }
    } else if(m.type==='capture-progress'){
      $('status').textContent=m.text;render.progress(m.text);$('cancelcapture').hidden=false;
      $('run').disabled=$('step').disabled=$('save').disabled=true;
    } else if(['raster-seek','raster-pixel','blit-seek','blit-pixel','seek','seek-progress','pixel','source','resource'].includes(m.type)){render.result(m);paneSet?.renderResult(m);
    } else if(m.type==='capture-cleared'){
      render.reset({starting:m.starting});paneSet?.clearCapture({starting:m.starting});
      $('capture-note').textContent=capabilities.renderCapture===false?'Rendering capture is unavailable in this development core.':platform==='dos'?'Open Render to trace RAM rendering and its copies to VGA.':'Use Capture next display in Render to record a complete interval.';$('cancelcapture').hidden=true;
    } else if(m.type==='capture'){
      render.setCapture(m);paneSet?.capture(m);showProfile(m.profile,true);lastTime=performance.now();lastSteps=m.end.steps;lastFrames=m.end.frames;lastSeconds=m.end.seconds??0;rate='';
      $('cancelcapture').hidden=true;
      $('capture-note').textContent=`Captured display ${m.start.frames}–${m.end.frames} · ${(m.elapsedMs/1000).toFixed(2)} s capture · longest call ${m.maxCall.toFixed(1)} ms · ${((m.info.bytes+(m.info.rasterBytes||0)+m.checkpointBytes)/1048576).toFixed(1)} MiB evidence/checkpoints${m.info.overflow?' · incomplete: trace limit reached':''}${m.info.renderBuffers?' · '+m.info.producerPixels.toLocaleString()+' pixels mapped to RAM'+(m.info.timedOut?' · bounded window ended before two VGA bursts':''):''}`;
    } else if(m.type==='saved'){
      if(pendingSuspend){const done=pendingSuspend;pendingSuspend=null;done.resolve({files:selected,firmware,driveFirmware,compatibility,backend,executable:$('program')?.value,stateFile:new File([m.bytes],platform+'.rrstate')});return;}
      const url=URL.createObjectURL(new Blob([m.bytes],{type:'application/octet-stream'}));
      const a=document.createElement('a');a.href=url;a.download=platform+'-'+Date.now()+'.rrstate';a.textContent='Download state';$('status').replaceChildren(document.createTextNode('State ready. Machine paused. '),a);a.click();setTimeout(()=>URL.revokeObjectURL(url),300000);
    } else if (m.type === 'slow') {
      console.warn('Long emulation slice', JSON.stringify(m));
    } else if (m.type === 'ready') {
      loaded = true;backend=m.backend||'production';capabilities=m.capabilities||{};
      backendControls();render.capabilities(capabilities);paneSet?.capabilities(capabilities);
      $('capture-note').textContent=capabilities.renderCapture===false?'Rendering capture is unavailable in this development core.':'Use Capture next display in Render to record a complete interval.';
      controls(true);paneSet?.mediaReady(selected);memory.ready();code?.ready(session);paneSet?.ready(session);feed.ready();$('game-name').textContent=selected[0]?.name||config.name;$('media-dialog').close();
      $('pause').disabled = true;
      $('status').textContent = m.text;
      lastTime = performance.now();
      send('turbo', {value : $('turbo').checked});
    } else if (m.type === 'message' || m.type === 'error') {
      $('status').textContent = m.text;
      if (m.type === 'error') {
        if(pendingSuspend){pendingSuspend.reject(Error(m.text));pendingSuspend=null;}
        render.reset();$('cancelcapture').hidden=true;
        controls(loaded);
        $('pause').disabled = true;
      }
    }
  };
  candidate.postMessage({type:'load',session:nextSession,platform,backend:requestedBackend,files:selected,firmware,driveFirmware,compatibility,stateFile,preparedStart,...(platform==='dos'?{executable:$('program').value}:{})});
}
$('load').onclick = () => {
  selected = [...$('files').files ];
  compatibility = $('compatprofile')?.checked ?? true;
  firmware = platform === 'c64'
                 ? [ 'basic', 'kernal', 'chargen' ].map(id => $(id).files[0])
                 : platform==='amiga'?[$('kickstart').files[0]]:platform==='ps2'?[$('bios').files[0]]:null;
  driveFirmware=$('drive-firmware')?.files[0]||null;load();
};
$('reset').onclick = ()=>load();
$('cancelcapture').onclick=()=>send('cancel-capture');
$('save').onclick=()=>{release();controls(false);$('status').textContent='Saving state…';send('save');};
$('statefile').onchange=()=>{const f=$('statefile').files[0];if(!f)return;if(!selected.length){selected=[...$('files').files];firmware=platform==='c64'?['basic','kernal','chargen'].map(id=>$(id).files[0]):platform==='amiga'?[$('kickstart').files[0]]:platform==='ps2'?[$('bios').files[0]]:null;compatibility=$('compatprofile')?.checked??true;driveFirmware=$('drive-firmware')?.files[0]||null;}load(f);$('statefile').value='';};
function transport(id,stay=false) {
    $('status').textContent =
        id === 'pause' ? 'Pausing at the current execution boundary…'
        : id === 'run' ? 'Running local image.'
                       : 'Advancing one display boundary…';
    if(id==='run'||id==='step'){memory.invalidate();render.reset();release();lastSteps=null;presentedCount=0;presentedAt=performance.now();}
    send(id);
    if(id==='run'&&!stay)document.querySelector('#viewport-root .viewport-game canvas')?.focus();
}
for(const id of ['run','pause','step'])$(id).onclick=()=>transport(id);
$('turbo').onchange = () => send('turbo', {value : $('turbo').checked});
$('fullscreen').onclick = () => canvas.parentElement.requestFullscreen();
$('tapeplay').onclick = () => send('tape', {down : 1});
$('tapestop').onclick = () => send('tape', {down : 0});
function input(keys = [], mouse = null) {
  if(render.isInspecting())return;
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
    ...((platform==='ds'||platform==='3ds'||platform==='dos')?{touch:inputTouch}:{}),
    buttons : buttons >>> 0,
    x : Math.max(-80, Math.min(80, x)),
    y : Math.max(-80, Math.min(80, y))
  },
        key = JSON.stringify(state);
  if (key !== lastInput || keys.length || mouse) {
    lastInput = key;
    send('input', {...state, keys, ...(mouse?{mouse}:{})});
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
  document.addEventListener(down ? 'keydown' : 'keyup', e => {
    if(![canvas,$('memory-screen'),$('code-screen')].includes(e.target)&&!e.target.closest?.('.viewport-game'))return;
    if(render.isInspecting())return;
    if (!loaded || e.repeat || e.metaKey)
      return;
    if(platform==='dos'&&dosKeys[e.code]!==undefined){e.preventDefault();const code=dosKeys[e.code];if(down)c64Keys.add(code);else c64Keys.delete(code);input([[code,+down]]);return;}
    if(platform==='amiga'&&amigaKeys[e.code]!==undefined){
      e.preventDefault();const code=amigaKeys[e.code],bit=config.keys[e.key];
      if(down)c64Keys.add(code);else c64Keys.delete(code);
      if(bit!==undefined){if(down)sources.set('key:'+e.code,{bits:[bit]});else sources.delete('key:'+e.code);}
      input([[code,+down]]);return;
    }
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
  },{signal:lifetime.signal});
function release() {
  inputTouch={...inputTouch,down:false};
  sources.clear();
  input([...c64Keys ].map(code => [code, 0]));
  c64Keys.clear();
}
canvas.onblur = release;
if($('code-screen')){$('code-screen').onblur=release;$('code-screen').onclick=()=>$('code-screen').focus();}
$('memory-screen').onblur=release;
$('memory-screen').onclick=()=>$('memory-screen').focus();
window.addEventListener('blur',release,{signal:lifetime.signal});
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    release();
    send('hold');
  }
},{signal:lifetime.signal});
const gamepadMaps = {
 dos:{0:32,1:64,8:128,9:16,12:1,13:2,14:4,15:8},
 xbox:{0:256,1:512,2:1024,3:2048,4:4096,5:8192,6:16384,7:32768,8:32,9:16,10:64,11:128,12:1,13:2,14:4,15:8},
 gba:{0:1,1:2,4:512,5:256,8:4,9:8,12:64,13:128,14:32,15:16},
 dc:{0:4,1:2,2:1024,3:512,6:65536,7:131072,9:8,12:16,13:32,14:64,15:128},
 gc:{0:256,1:512,2:1024,3:2048,4:16,5:16,6:64,7:32,9:4096,12:8,13:4,14:1,15:2},
 ps2:{0:16384,1:8192,2:32768,3:4096,4:1024,5:2048,6:256,7:512,8:1,9:8,10:2,11:4,12:16,13:64,14:128,15:32},
 amiga:{0:16,1:32,2:64,12:1,13:2,14:4,15:8},
 gb:{0:16,1:32,8:64,9:128,12:4,13:8,14:2,15:1},
 gg:{0:32,1:16,9:128,12:1,13:2,14:4,15:8},
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
  if(disposed)return;
  const pads = [...(navigator.getGamepads?.() || []) ].filter(Boolean),
        p = pads.find(p => p.mapping === 'standard');
  const allowed = loaded && !render.isInspecting() && document.hasFocus() && !document.hidden &&
                  (document.activeElement?.closest?.('.viewport-game') || document.activeElement === canvas || document.activeElement === $('memory-screen') || document.activeElement === $('code-screen') ||
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
    } else if(platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox'){
      sources.set('gamepad:'+p.index,{bits,ax:x,ay:y});
    } else {
      const directions =
          platform === 'dos' ? [4,8,1,2] : platform === 'gb' ? [2,1,4,8] : platform === 'gg' ? [4,8,1,2] : (platform === 'ds'||platform==='gba') ? [32,16,64,128] : platform === 'ps1' ? [ 128, 32, 16, 64 ]
          : (platform === 'c64'||platform==='amiga')
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

if(platform==='dos'){
 $('files').onchange=()=>{try{const entries=dosFiles([...$('files').files]);$('program').replaceChildren(...entries.filter(e=>/\.exe$/i.test(e.path)).map(e=>{const o=document.createElement('option');o.value=o.textContent=e.path;return o;}));$('status').textContent='Choose the DOS executable, then load.';}catch(e){$('status').textContent=e.message;}};

}

function bindPointerInput(target){
if(platform==='ds'||platform==='3ds'){
 let stylus=null;const screenHeight=platform==='ds'?192:240,screenWidth=platform==='ds'?256:320,left=platform==='ds'?0:40;
 const point=e=>{const p=pixelCoordinates(target.getBoundingClientRect(),target.width,target.height,e.clientX,e.clientY);return p&&p.y>=screenHeight&&p.x>=left&&p.x<left+screenWidth?{x:p.x-left,y:p.y-screenHeight}:null;};
 const pen=(e,down)=>{if(!loaded||render.isInspecting())return;const p=point(e);if(down&&!p)return;e.preventDefault();inputTouch={x:p?.x??inputTouch.x,y:p?.y??inputTouch.y,down};input();};
 target.addEventListener('pointerdown',e=>{if(!point(e)||render.isInspecting())return;stylus=e.pointerId;target.setPointerCapture(stylus);target.focus();pen(e,true);});
 target.addEventListener('pointermove',e=>{if(e.pointerId===stylus)pen(e,true);});
 for(const type of ['pointerup','pointercancel','lostpointercapture'])target.addEventListener(type,e=>{if(e.pointerId===stylus){pen(e,false);stylus=null;}});
}
if(platform==='amiga'){
 let last=null,fx=0,fy=0;
 const buttons=e=>{sources.set('amigaMouse',{bits:[...(e.buttons&1?[32]:[]),...(e.buttons&2?[64]:[])]});input();};
 target.addEventListener('pointerenter',e=>{last={x:e.clientX,y:e.clientY};});
 target.addEventListener('pointerleave',()=>{last=null;});
 target.addEventListener('pointermove',e=>{
   const previous=last;last={x:e.clientX,y:e.clientY};
   if(!loaded||render.isInspecting()||!previous)return;
   const rect=target.getBoundingClientRect(),speed=Number($('mouse-speed').value);
   fx+=(e.clientX-previous.x)*320/rect.width*speed;fy+=(e.clientY-previous.y)*256/rect.height*speed;
   const x=Math.trunc(fx),y=Math.trunc(fy);fx-=x;fy-=y;if(x||y)input([],{x,y});
 });
 target.addEventListener('pointerdown',e=>{if(!loaded||render.isInspecting())return;e.preventDefault();target.focus();target.setPointerCapture(e.pointerId);buttons(e);});
 target.addEventListener('pointerup',e=>{if(!render.isInspecting())buttons(e);});
 for(const type of ['pointercancel','lostpointercapture'])target.addEventListener(type,()=>{sources.delete('amigaMouse');input();});
 target.addEventListener('contextmenu',e=>e.preventDefault());
}
if(platform==='dos'){
 target.addEventListener('contextmenu',e=>e.preventDefault());
 const mouse=e=>{if(!loaded||render.isInspecting())return;e.preventDefault();const p=pixelCoordinates(target.getBoundingClientRect(),320,200,e.clientX,e.clientY);if(!p)return;inputTouch={x:p.x,y:p.y,down:!!e.buttons};sources.set('dosMouse',{bits:[e.buttons&1?256:0,e.buttons&2?512:0]});input();};
 target.addEventListener('pointerdown',e=>{if(render.isInspecting())return;target.focus();target.setPointerCapture(e.pointerId);mouse(e);});for(const event of ['pointermove','pointerup'])target.addEventListener(event,mouse);for(const event of ['pointercancel','lostpointercapture'])target.addEventListener(event,()=>{sources.delete('dosMouse');input();});
}

}
bindPointerInput(canvas);
const warehouse=$('panel-warehouse');$('session-options').append($('pad'),$('fullscreen'));if($('mouse-speed'))$('session-options').append($('mouse-speed').closest('label'));
const lesson=document.createElement('div');lesson.className='lesson-pane';warehouse.append(lesson);for(const node of $('code-workspace').querySelectorAll('.tour-panel'))lesson.append(node);
preparedPanel=createPreparedPanel({root:lesson,start:id=>{if(pendingWorker||!loaded||latestState?.experimentOwned||latestState?.debugBusy||latestState?.capturing||latestState?.saving||latestState?.memoryRecording)return;release();load(null,id);},cancel:()=>cancelPreparation?.()});
lessonShortcut=document.createElement('button');lessonShortcut.textContent='Open prepared lessons';lessonShortcut.hidden=true;lessonShortcut.onclick=()=>views.reveal('lesson');$('session-options').prepend(lessonShortcut);
const renderSources=$('render-auxiliary'),renderDetails=$('render-details');warehouse.append(renderSources,renderDetails);
paneSet=createViewportPanels({views,platform,send,transport:id=>transport(id,true),feed,canvas,warehouse,recording:memory,presentation,renderTemplate,legacy:{lesson,recording:$('memory-workspace'),render:$('render-workspace'),'render-sources':renderSources,'render-details':renderDetails,session:$('session-options')},bindGameInput(c){
 c.onclick=()=>c.focus();c.onblur=release;
 bindPointerInput(c);
 return()=>{c.onblur=null;c.onclick=null;};
}});
const debug=['c64','dos','3do'].includes(platform);
views.register({id:'loader',label:'Loader',panel:document.createElement('div')});
if(platform==='c64')views.register({id:'guided',label:'Guided tour',panel:document.createElement('div')});
const entries=[['game','Game'],['storage','Storage'],['atlas','Memory atlas'],['hex','Memory bytes'],['recording','Memory recording'],['render','Rendering'],['render-sources','Render sources'],['render-details','Render details'],['session','Session / controls']];
if(platform==='c64')entries.push(['drive','1541 drive']);
if(debug)entries.push(['code','Code'],['state','Game state'],['lesson','Lesson / experiment']);
views.configure(entries,(kind,id,root)=>paneSet.make(kind,id,root));
$('restore-layout').onclick=()=>views.restorePreset();
$('system-select').value=platform;$('system-select').onchange=()=>onSystem($('system-select').value);
$('inspect-storage').onclick=()=>{const files=[...$('files').files];if(!files.length){$('status').textContent='Select local media to inspect.';return;}paneSet.media(files);views.select('loader');$('media-dialog').close();$('status').textContent='Inspecting selected media. The current machine is unchanged.';};
$('open-media').onclick=()=>$('media-dialog').showModal();$('close-media').onclick=()=>$('media-dialog').close();
// State watches are independent of Code visibility; one shared 5 Hz sampler.
sampler=setInterval(()=>{if(loaded&&debug&&views.visible('state')&&!latestState?.debugBusy&&!latestState?.capturing&&!latestState?.saving&&!latestState?.memoryRecording)code?.refreshState();},200);
return {
 platform,
 restore(saved){selected=saved.files;firmware=saved.firmware;driveFirmware=saved.driveFirmware||null;compatibility=saved.compatibility;if($('program')&&saved.executable){const o=document.createElement('option');o.value=o.textContent=saved.executable;$('program').append(o);$('program').value=saved.executable;}load(saved.stateFile,null,saved.backend||'production');},
 async suspend(){if(!loaded)return null;if(latestState&&(latestState.debugBusy||latestState.capturing||latestState.saving||latestState.memoryRecording||latestState.experimentOwned))throw Error('Finish the active inspection or experiment before switching systems.');release();return new Promise((resolve,reject)=>{const timeout=setTimeout(()=>{pendingSuspend=null;reject(Error('Could not suspend the current game. It remains open.'));},30000);pendingSuspend={resolve:v=>{clearTimeout(timeout);resolve(v);},reject:e=>{clearTimeout(timeout);reject(e);}};send('save');});},
 dispose(){disposed=true;clearInterval(sampler);lifetime.abort();release();pendingWorker?.terminate();worker?.terminate();feed.dispose();views.dispose();code?.dispose();memory.setActive(false);},
 status(text){$('status').textContent=text;$('system-select').value=platform;}
};
}
