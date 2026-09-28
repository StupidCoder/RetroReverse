import {packState, unpackState, digest} from './state.js';
import {InputQueue} from './input.js';
import {selectMedia, sha256File} from './media.js';
import {platforms} from './platforms.js';

let profileBase = {}, core, platform, session, loaded = false, running = false,
    epoch = 0, frames = 0, turbo = false, files, maxCall = 0, runMs = 0,
    paintMs = 0, lastPaint = 0;
const inputQueue = [];
let bootOptions, firmwareIdentity=[], mediaIdentity, coreIdentity;
async function identities(){
 if(!mediaIdentity){send("message",{text:"Verifying local media identity…"});mediaIdentity=[];for(const f of files)mediaIdentity.push({name:f.name,size:f.size,sha256:await sha256File(f)});mediaIdentity.sort((a,b)=>a.name.localeCompare(b.name));}
 return {media:mediaIdentity,firmware:firmwareIdentity,core:coreIdentity,configuration:{compatibility:bootOptions.compatibility!==false,customFirmware:!!bootOptions.firmware?.some(Boolean)}};
}
function queueState(){return {...inputs,pulses:[...inputs.pulses],down:[...inputs.down],pending:inputQueue,lastButtons,lastX,lastY,inputSequence,lastInputStep};}
async function saveState(){
 cancelCapture();running=false;++epoch;const identity=await identities();
 const n=core._rr_state_save();check(n);const p=core._rr_state_data();
 const bytes=await packState({format:1,platform,...identity,input:queueState()},core.HEAPU8.slice(p,p+n));
 paint();send("saved",{bytes:bytes.buffer});
}

let capture=null,capturing=false,captureGeneration=0,captureProfileStart=null,captureRunMs=0,maxCaptureCall=0;
function finishCaptureProfile(){
 if(!captureProfileStart)return null;const p=json('_rr_profile');
 for(const b of p.buckets){const delta=b.ms-(captureProfileStart[b.name]||0);profileBase[b.name]=(profileBase[b.name]||0)+delta;b.ms=delta;}
 captureProfileStart=null;return p;
}
let inputs, lastButtons = -1, lastX = 0, lastY = 0;
let inputSequence = 0, lastInputStep = 0;
const sleep = n => new Promise(r => setTimeout(r, n));
const json = fn => JSON.parse(core.UTF8ToString(core[fn]()));
const jsonPixel=(x,y)=>JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));
const send = (type, data = {}) => postMessage({type, session, ...data});
const error = () => core.UTF8ToString(core._rr_error());
function check(ok) {
  if (!ok)
    throw Error(core.discError || error() || 'Core rejected the operation');
}
function status() {
  const s = json('_rr_status');
  s.frames = platform === 'n64'   ? Math.floor(s.steps / 750000)
             : platform === '3do' ? s.frame
             : platform === 'ps1' ? s.fields
                                  : s.frames;
  s.steps ??= s.cycle;
  s.inputSequence = inputSequence;
  s.lastInputStep = lastInputStep;
  return s;
}
function readProfile() {
  if (!core._rr_profile)
    return null;
  const p = json('_rr_profile');
  for (const b of p.buckets)
    b.ms -= profileBase[b.name] || 0;
  return p;
}
function paint() {
  const start = performance.now(), s = status();
  frames = s.frames;
  const w = platform === 'c64' ? 392 : s.width || 320,
        h = platform === 'c64' ? 272 : s.height || 240, p = core._rr_frame(),
        pixels = core.HEAPU8.slice(p, p + w * h * 4);
  paintMs += performance.now() - start;
  send('state', {
    state : s,
    width : w,
    height : h,
    pixels : pixels.buffer,
    running,
    maxCall,
    runMs,
    paintMs,
    heap : core.HEAPU8.length,
    inputDeferred : platform === '3do' && core.deferInput && s.frames < 300,
    capturing,
    profile : readProfile()
  });
}
function applyInputs() {
  const s = status();
  if (platform === '3do' && core.deferInput && s.frames < 300)
    return;
  const time = platform === 'c64'   ? s.steps / 985248
               : platform === 'n64' ? s.steps / 45000000
               : platform === 'ps1' ? s.steps / 15000000
                                    : s.frames / 30;
  for (const m of inputQueue.splice(0))
    inputs.enqueue(m, time);
  const m = inputs.drain(time);
  if (m.buttons !== lastButtons || m.x !== lastX || m.y !== lastY) {
    if (platform === 'c64')
      core._rr_joystick(2, m.buttons);
    else if (platform === 'n64')
      core._rr_pad(m.buttons, m.x, m.y);
    else
      core._rr_pad(platform === 'ps1' ? (~m.buttons) & 65535 : m.buttons);
    lastButtons = m.buttons;
    lastX = m.x;
    lastY = m.y;
  }
  if (platform === 'c64')
    for (const [code, down] of m.keys)
      core._rr_key(code, down);
  inputSequence = m.sequence;
  lastInputStep = s.steps;
}
function tick(one = false) {
  applyInputs();
  const start = performance.now();
  if (platform === 'c64')
    check(core._rr_run(10000, one ? 7 : 0, 0) >= 0);
  else if (platform === 'ps1')
    check(core._rr_run(10000, +one) >= 0);
  else if (platform === 'n64')
    check(core._rr_run(one ? Math.min(10000, 750000 - status().steps % 750000)
                           : 10000) >= 0);
  else
    check(core._rr_run_slice(10000));
  const ms = performance.now() - start;
  if(capturing){captureRunMs+=ms;maxCaptureCall=Math.max(maxCaptureCall,ms);}else{runMs += ms;maxCall = Math.max(maxCall, ms);}
  if (ms > 250)
    send('slow', {ms, state : status(), profile : readProfile()});
}
async function pump(id, one = false) {
  const first = frames, initial = status();
  let startWall = performance.now(), wasTurbo = turbo;
  const position = s => platform === 'c64' ? s.steps / 985248
                        : platform === '3do'
                            ? s.frames / 30
                            : s.frames / platforms[platform].hz;
  let startEmu = position(initial);
  while (running && id === epoch) {
    const start = performance.now();
    do {
      tick(one);
    } while (performance.now() - start < 8 && !one);
    const s = status(), now = performance.now();
    frames = s.frames;
    if (now - lastPaint > 80 || one) {
      paint();
      lastPaint = now;
    }
    if (one && frames > first)
      break;
    if (turbo !== wasTurbo ||
        (position(s) - startEmu) * 1000 - (performance.now() - startWall) <
            -100) {
      startWall = performance.now();
      startEmu = position(s);
      wasTurbo = turbo;
    }
    const lead =
        (position(s) - startEmu) * 1000 - (performance.now() - startWall);
    await sleep(!turbo && !one ? Math.max(0, Math.min(20, lead)) : 0);
    // Never accumulate unlimited real-time credit while a slow frame is
    // executing.
    if (!turbo && !one && lead > 20) {
      while (running && id === epoch && !turbo &&
             (position(s) - startEmu) * 1000 - (performance.now() - startWall) >
                 20)
        await sleep(20);
    }
  }
  if (id === epoch) {
    running = false;
    paint();
    send('message',{text:one?'Paused at the next display boundary.':'Paused.'});
  }
}
function coreState(){const n=core._rr_state_save();check(n);if(n>32*1024*1024)throw Error('Capture checkpoint exceeds the 32 MiB budget');const p=core._rr_state_data();return core.HEAPU8.slice(p,p+n);}
function cancelCapture(){
 if(capturing){core._rr_capture_end();finishCaptureProfile();capturing=false;}
 capture=null;captureGeneration++;send('capture-cleared');
}
async function captureNext(){
 running=false;const id=++epoch;cancelCapture();const generation=captureGeneration;capturing=true;captureRunMs=maxCaptureCall=0;captureProfileStart=Object.fromEntries(json('_rr_profile').buckets.map(b=>[b.name,b.ms]));
 const began=performance.now();let lastProgress=0;
 send('capture-progress',{text:'Finishing the current display interval…',generation});
 const boundary=async phase=>{
   const first=status().frames;
   while(status().frames===first){
     if(id!==epoch)throw Error('Capture cancelled');
     const start=performance.now();do{tick(true);}while(status().frames===first&&performance.now()-start<8);
     if(performance.now()-lastProgress>250){lastProgress=performance.now();send('capture-progress',{text:phase+' '+((performance.now()-began)/1000).toFixed(1)+' s',generation});}
     await sleep(0);
   }
   if(id!==epoch)throw Error('Capture cancelled');
 };
 try{
   await boundary('Finishing current interval.');
   const startState=coreState(),start=status(),input=queueState();
   core._rr_capture_begin();
   await boundary('Recording next complete interval.');
   core._rr_capture_end();const captureProfile=finishCaptureProfile();capturing=false;
   const endState=coreState(),end=status(),info=json('_rr_capture_info');
   const w=platform==='c64'?392:end.width||320,h=platform==='c64'?272:end.height||240,p=core._rr_frame();
   const pixels=core.HEAPU8.slice(p,p+w*h*4);
   const frameHash=await digest(pixels);
   if(id!==epoch)return;
   capture={id:generation,startState,endState,start,end,input,width:w,height:h,pixels,info,frameHash};
   paint();send('capture',{id:generation,start,end,width:w,height:h,info,frameHash,elapsedMs:performance.now()-began,checkpointBytes:startState.length+endState.length,profile:captureProfile,runMs:captureRunMs,maxCall:maxCaptureCall});
   send('message',{text:info.overflow?'Paused. Capture limit reached; some evidence is missing.':'Paused. A complete display interval is ready to inspect.'});
 }catch(e){if(id===epoch){if(capturing)core._rr_capture_end();finishCaptureProfile();capturing=false;capture=null;paint();send('error',{text:String(e)});}}
}
async function boot(m) {
  bootOptions=m;session=m.session;
  const restored=m.stateFile?await unpackState(m.stateFile):null;
  platform = m.platform;
  if(restored&&restored.meta.platform!==platform)throw Error("This state belongs to another console");
  inputs = new InputQueue(platforms[platform].hz);
  session = m.session;
  files = m.files;
  if (!files?.length)
    throw Error('Select a game image');
  send('message', {text : 'Loading emulator…'});
  const factory = (await import(`./cores/${platform}/core.js`)).default;
  core = await factory();
  coreIdentity=(await (await fetch('./build-manifest.json')).json())[platform+'/core.wasm'];
  const f = await selectMedia(files);
  if (platform === 'c64') {
    const custom = m.firmware?.some(Boolean);
    if (custom &&
        (!m.firmware.every(Boolean) ||
         m.firmware.some((f, i) => f.size !== [ 8192, 8192, 4096 ][i])))
      throw Error(
          'Firmware override needs 8 KiB BASIC, 8 KiB KERNAL and 4 KiB character ROMs');
    const manifest =
              custom
                  ? []
                  : await (await fetch('./firmware/c64/manifest.json')).json(),
          roms = custom ? await Promise.all(m.firmware.map(
                              async f => new Uint8Array(await f.arrayBuffer())))
                        : [];
    for (const entry of manifest) {
      const r = await fetch('./firmware/c64/' + entry.file);
      if (!r.ok)
        throw Error('Firmware download failed: ' + entry.file);
      const b = await r.arrayBuffer();
      const hash =
          [...new Uint8Array(await crypto.subtle.digest('SHA-256', b)) ]
              .map(x => x.toString(16).padStart(2, '0'))
              .join('');
      if (hash !== entry.sha256)
        throw Error('Firmware hash mismatch');
      roms.push(new Uint8Array(b));
    }
    firmwareIdentity=await Promise.all(roms.map(b=>digest(b)));
    let pos = core._rr_input();
    for (const b of roms) {
      core.HEAPU8.set(b, pos);
      pos += b.length;
    }
    check(core._rr_init(8192, 8192, 4096));
    if (f.size > 2 * 1024 * 1024)
      throw Error('TAP exceeds current 2 MiB core limit');
    core.HEAPU8.set(new Uint8Array(await f.arrayBuffer()), core._rr_input());
    check(core._rr_tape(f.size));
    core._rr_trace(0, 0, 0);
  } else if (platform === '3do') {
    if (f.size > 0xffffffff)
      throw Error('Disc exceeds 4 GiB limit');
    core.discFile = f;
    let profile = false;
    if (m.compatibility !== false && f.size === 771408960) {
      send('message', {text : 'Checking optional compatibility profile…'});
      profile =
          (await sha256File(f)) ===
          '0828e6f43e527f5a11b85b02aa5cd1d0ed93bad03c318d5b6fa735e5e3c9715c';
    }
    check(core._rr_init_config(f.size, +profile));
    core.deferInput = profile;
    core.compatProfile =
        profile ? 'Need for Speed: VBL mirror and intro input compatibility'
                : 'Generic Portfolio boot';
  } else if (platform === 'ps1') {
    if (f.size > 0xffffffff)
      throw Error('Disc exceeds 4 GiB limit');
    core.discFile = f;
    check(core._rr_init_file(f.size) > 0);
  } else {
    if (f.size > 64 * 1024 * 1024 || f.size % 4)
      throw Error('N64 image must be aligned to 4 bytes and at most 64 MiB');
    const b = await f.arrayBuffer(), p = core._rr_input(b.byteLength);
    check(p);
    core.HEAPU8.set(new Uint8Array(b), p);
    check(core._rr_init(b.byteLength));
  }
  profileBase =
      Object.fromEntries((core._rr_profile ? json('_rr_profile').buckets : [])
                             .map(b => [b.name, b.ms]));
  if(restored){
    const identity=await identities();
    for(const key of ['media','firmware','core','configuration'])
      if(JSON.stringify(identity[key])!==JSON.stringify(restored.meta[key]))throw Error('State '+key+' does not match. Reselect the original game, firmware and compatibility settings.');
    const p=core._rr_state_input(restored.payload.length);check(p);core.HEAPU8.set(restored.payload,p);check(core._rr_state_load(restored.payload.length));
    const q=restored.meta.input;
    if(!q||!Array.isArray(q.keys)||q.keys.length>4096||!Array.isArray(q.pending)||q.pending.length>4096)throw Error('Invalid saved input queue');
    Object.assign(inputs,q,{pulses:new Map(q.pulses),down:new Map(q.down)});
    inputQueue.push(...q.pending);lastButtons=q.lastButtons;lastX=q.lastX;lastY=q.lastY;inputSequence=q.inputSequence;lastInputStep=q.lastInputStep;
    // Saved events remain part of the state. Release host-held controls before
    // continuation; a newly pressed physical button creates a new input edge.
    inputs.buttons=0;inputs.x=inputs.y=0;inputs.pulses.clear();inputQueue.length=0;
    if(platform==='c64')for(let k=0;k<256;k++)core._rr_key(k,0);
    inputs.keys=[];inputs.down.clear();
  }
  loaded = true;
  paint();
  send('ready', {
    text : (restored?'State restored, paused. ':'') + f.name + ' loaded. ' + (core.compatProfile || '') + ' Press Run.'
  });
}
onmessage = async ({data : m}) => {
  try {
    if (m.type === 'load') {
      await boot(m);
      return;
    }
    if (m.session !== session || !loaded)
      return;
    if(m.type==='save'){await saveState();return;}
    if(m.type==='pixel'){if(!capture||m.capture!==capture.id)return;send('pixel',{capture:capture.id,x:m.x,y:m.y,evidence:jsonPixel(m.x,m.y),request:m.request});return;}
    if(m.type==='source'||m.type==='resource'){
      if(!capture||m.capture!==capture.id)return;
      const fn=m.type==='source'?core._rr_source:core._rr_resource;
      if(!fn)return;
      const p=m.type==='source'?fn(m.address,m.size,m.before,m.expected):fn(m.resource,m.offset);
      send(m.type,{capture:capture.id,evidence:JSON.parse(core.UTF8ToString(p)),request:m.request});return;
    }
    if(m.type==='cancel-capture'||m.type==='hold'){running=false;++epoch;cancelCapture();paint();send('message',{text:'Paused.'});return;}
    if (m.type === 'run' || m.type === 'step') {
      if (running)
        return;
      cancelCapture();running = true;
      await pump(++epoch, m.type === 'step');
    } else if (m.type === 'pause') {
      await captureNext();
    } else if (m.type === 'turbo')
      turbo = m.value;
    else if (m.type === 'input') {
      if(capturing)return;
      if (inputQueue.length >= 4096)
        throw Error('Input queue overflow');
      inputQueue.push(m);
    } else if (m.type === 'tape')
      core._rr_play(m.down);
  } catch (e) {
    running = false;
    ++epoch;
    send('error', {text : String(e)});
  }
};
