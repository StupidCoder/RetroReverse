import {packState, unpackState, digest} from './state.js';
import {FrameClock} from './pacing.js';
import {InputQueue} from './input.js';
import {selectMedia, sha256File} from './media.js';
import {platforms} from './platforms.js';

let profileBase = {}, core, platform, session, loaded = false, running = false,
    epoch = 0, frames = 0, turbo = false, files, maxCall = 0, runMs = 0,
    paintMs = 0, lastPaint = 0;
const inputQueue = [];
const mediaHashes=new WeakMap();
function mediaHash(file){if(!mediaHashes.has(file))mediaHashes.set(file,sha256File(file));return mediaHashes.get(file);}
let saving=false;
let bootOptions, firmwareIdentity=[], mediaIdentity, coreIdentity;
async function identities(){
 if(!mediaIdentity){send("message",{text:"Verifying local media identity…"});mediaIdentity=[];for(const f of files)mediaIdentity.push({name:f.name,size:f.size,sha256:await mediaHash(f)});mediaIdentity.sort((a,b)=>a.name.localeCompare(b.name));}
 return {media:mediaIdentity,firmware:firmwareIdentity,core:coreIdentity,configuration:{compatibility:bootOptions.compatibility!==false,customFirmware:!!bootOptions.firmware?.some(Boolean)}};
}
function queueState(){return {...inputs,pulses:[...inputs.pulses],down:[...inputs.down],pending:inputQueue,lastButtons,lastX,lastY,inputSequence,lastInputStep};}
async function saveState(){
 if(saving)return;saving=true;
 try{
 cancelCapture();running=false;++epoch;const identity=await identities();
 const n=core._rr_state_save();check(n);const p=core._rr_state_data();
 const bytes=await packState({format:1,platform,...identity,input:queueState()},core.HEAPU8.slice(p,p+n));
 send("saved",{bytes:bytes.buffer});
 }finally{saving=false;paint();}
}

let seekGeneration=0;
let capture=null,capturing=false,captureGeneration=0,captureProfileStart=null,captureRunMs=0,maxCaptureCall=0;
function finishCaptureProfile(){
 if(!captureProfileStart)return null;const p=json('_rr_profile');
 for(const b of p.buckets){const delta=b.ms-(captureProfileStart[b.name]||0);profileBase[b.name]=(profileBase[b.name]||0)+delta;b.ms=delta;}
 captureProfileStart=null;return p;
}
let inputs, lastButtons = -1, lastX = 0, lastY = 0;
let inputSequence = 0, lastInputStep = 0;
// Yield to queued input without turning every 8 ms work slice into a nested
// timer. Timed waits still use setTimeout for the actual emulation pacing.
const yieldChannel = new MessageChannel(), yieldQueue = [];
yieldChannel.port1.onmessage = () => yieldQueue.shift()?.();
const sleep = n => n > 0 ? new Promise(r => setTimeout(r, n)) :
  new Promise(r => { yieldQueue.push(r); yieldChannel.port2.postMessage(0); });
const json = fn => JSON.parse(core.UTF8ToString(core[fn]()));
const jsonPixel=(x,y)=>JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));
const send = (type, data = {}) => postMessage({type, session, ...data}, data.pixels ? [data.pixels] : []);
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
    capturing,saving,
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
                                    : s.inputSeconds ?? s.frames / platforms[platform].hz;
  for (const m of inputQueue.splice(0))
    inputs.enqueue(m, time);
  const m = inputs.drain(time);
  if (m.buttons !== lastButtons || m.x !== lastX || m.y !== lastY) {
    if (platform === 'c64')
      core._rr_joystick(2, m.buttons);
    else if (platform === 'n64'||platform==='3ds')
      core._rr_pad(m.buttons, m.x, m.y);
    else
      core._rr_pad(platform === 'ps1' ? (~m.buttons) & 65535 : m.buttons);
    lastButtons = m.buttons;
    lastX = m.x;
    lastY = m.y;
  }
  if (platform === 'ds'||platform==='3ds')core._rr_touch(m.touch.x,m.touch.y,+m.touch.down);
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
  else if(platform==='ds'||platform==='3ds')check(core._rr_run(10000)>=0);
  else
    check(core._rr_run_slice(10000));
  const ms = performance.now() - start;
  if(capturing){captureRunMs+=ms;maxCaptureCall=Math.max(maxCaptureCall,ms);}else{runMs += ms;maxCall = Math.max(maxCall, ms);}
  if (ms > 250)
    send('slow', {ms, state : status(), profile : readProfile()});
}
async function pump(id, one = false) {
  const position = s => platform === 'c64' ? s.steps / 985248
      : s.seconds ?? s.frames / platforms[platform].hz;
  const clock=new FrameClock(performance.now(),position(status()),turbo);
  let boundary=status().frames;
  while(running&&id===epoch){
    const start=performance.now();
    // Always stop at a display boundary, including fast-forward. Sampling the
    // C64's in-progress raster buffer was the source of mixed-frame tearing.
    do{tick(true);}while(status().frames===boundary&&performance.now()-start<8);
    const s=status(),complete=s.frames!==boundary,now=performance.now();
    frames=s.frames;
    if(complete){
      boundary=s.frames;
      if(!turbo||one||now-lastPaint>=1000/60){paint();lastPaint=now;}
      if(one)break;
      let delay=clock.delay(performance.now(),position(s),turbo);
      while(delay>0&&running&&id===epoch&&!turbo){
        await sleep(Math.min(delay,20));
        delay=clock.delay(performance.now(),position(s),turbo);
      }
    }
    // Yield for input/cancellation without accumulating unbounded catch-up.
    await sleep(0);
  }
  if(id===epoch){running=false;paint();send('message',{text:one?'Paused at the next display boundary.':'Paused.'});}
}
function coreState(){const n=core._rr_state_save();check(n);const limit=platform==='3ds'?128:32;if(n>limit*1024*1024)throw Error(`Capture checkpoint exceeds the ${limit} MiB budget`);const p=core._rr_state_data();return core.HEAPU8.slice(p,p+n);}
function cancelCapture(){
 if(capturing){core._rr_capture_end();finishCaptureProfile();capturing=false;}
 capture=null;captureGeneration++;seekGeneration++;send('capture-cleared');
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
   const captureStarted=core._rr_capture_begin();if(platform==='ds'||platform==='3ds')check(captureStarted);
   const captureFields=platform==='ps1'?4:platform==='3ds'?3:1;
   for(let field=0;field<captureFields;field++)await boundary(captureFields>1?'Recording display and double-buffer producer context.':'Recording next complete interval.');
   const ended=core._rr_capture_end();if(platform==='ds'||platform==='3ds')check(ended);const captureProfile=finishCaptureProfile();capturing=false;
   const endState=coreState(),end=status(),info=json('_rr_capture_info'),replay=json('_rr_replay_begin');
   const w=platform==='c64'?392:end.width||320,h=platform==='c64'?272:end.height||240,p=core._rr_frame();
   const pixels=core.HEAPU8.slice(p,p+w*h*4);
   // Validate visible output rather than assuming every memory space is watched.
   while(!core._rr_replay_seek(replay.count)){await sleep(0);if(id!==epoch)return;}
   const rp=core._rr_replay_frame();
   replay.complete=replay.complete&&pixels.every((v,i)=>core.HEAPU8[rp+i]===v);
   core._rr_replay_seek(0);
   const frameHash=await digest(pixels);
   if(id!==epoch)return;
   capture={id:generation,startState,endState,start,end,input,width:w,height:h,pixels,info,frameHash};
   paint();send('capture',{id:generation,replay,start,end,width:w,height:h,info,frameHash,elapsedMs:performance.now()-began,checkpointBytes:startState.length+endState.length,profile:captureProfile,runMs:captureRunMs,maxCall:maxCaptureCall});
   send('message',{text:info.overflow?'Paused. Capture limit reached; some evidence is missing.':'Paused. A complete display interval is ready to inspect.'});
 }catch(e){if(id===epoch){if(capturing)core._rr_capture_end();finishCaptureProfile();capturing=false;capture=null;paint();send('error',{text:String(e)});}}
}
function pixelEvidence(x,y){const p=jsonPixel(x,y);for(const c of p.contributors||[])c.replayStep=core._rr_replay_for_write(platform==='c64'?y:c.id);return p;}
async function seekReplay(m){
 if(!capture||capture.id!==m.capture)return;
 const gen=++seekGeneration,c=capture,began=performance.now();let progress=began;
 while(!core._rr_replay_seek(m.step)){
  if(performance.now()-progress>100){progress=performance.now();send('seek-progress',{capture:c.id,request:m.request});}
  await sleep(0);if(gen!==seekGeneration||capture!==c)return;
 }
 if(gen!==seekGeneration||capture!==c)return;
 const p=core._rr_replay_frame(),pixels=core.HEAPU8.slice(p,p+c.width*c.height*4);
 send('seek',{capture:c.id,request:m.request,pixels:pixels.buffer,info:json('_rr_replay_info'),elapsedMs:performance.now()-began});
}
async function boot(m) {
  const bootBegan=performance.now();
  bootOptions=m;session=m.session;
  const restored=m.stateFile?await unpackState(m.stateFile):null;
  platform = m.platform;
  if(restored&&restored.meta.platform!==platform)throw Error("This state belongs to another console");
  inputs = new InputQueue(platforms[platform].hz);
  if(platform==='ds'||platform==='3ds')inputs.hold=3/platforms[platform].hz;
  session = m.session;
  files = m.files;
  if (!files?.length)
    throw Error('Select a game image');
  send('message', {text : 'Loading emulator…'});
  const manifest=await (await fetch('./build-manifest.json')).json();
  coreIdentity=manifest[platform+'/core.wasm'];
  const response=await fetch(`./cores/${platform}/core.wasm`);
  if(!response.ok)throw Error('Emulator binary download failed');
  const wasmBinary=new Uint8Array(await response.arrayBuffer());
  if(await digest(wasmBinary)!==coreIdentity)throw Error('Emulator build mismatch. Reload the page to obtain a matching release.');
  const factory = (await import(`./cores/${platform}/core.js`)).default;
  core = await factory({wasmBinary});
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
          (await mediaHash(f)) ===
          '0828e6f43e527f5a11b85b02aa5cd1d0ed93bad03c318d5b6fa735e5e3c9715c';
    }
    check(core._rr_init_config(f.size, +profile));
    core.deferInput = false;
    core.compatProfile =
        profile ? 'Need for Speed: VBL mirror and Cinepak movie HLE'
                : 'Generic Portfolio boot';
  } else if (platform === 'ps1') {
    if (f.size > 0xffffffff)
      throw Error('Disc exceeds 4 GiB limit');
    core.discFile = f;
    check(core._rr_init_file(f.size) > 0);
  } else if(platform==='ds'||platform==='3ds'){
    const limit=platform==='ds'?512:1024;if(f.size>limit*1024*1024)throw Error(`Cartridge exceeds ${limit} MiB`);
    const b=await f.arrayBuffer(),p=core._rr_input(b.byteLength);check(p);core.HEAPU8.set(new Uint8Array(b),p);check(core._rr_init(b.byteLength));
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
    Object.assign(inputs,q,{hold:inputs.hold,pulses:new Map(q.pulses),down:new Map(q.down)});
    inputQueue.push(...q.pending);lastButtons=q.lastButtons;lastX=q.lastX;lastY=q.lastY;inputSequence=q.inputSequence;lastInputStep=q.lastInputStep;
    // Saved events remain part of the state. Release host-held controls before
    // continuation; a newly pressed physical button creates a new input edge.
    inputs.buttons=0;inputs.x=inputs.y=0;inputs.pulses.clear();inputQueue.length=0;
    if(platform==='c64')for(let k=0;k<256;k++)core._rr_key(k,0);
    inputs.keys=[];inputs.down.clear();inputs.touch={x:0,y:0,down:false};inputs.touchEvents=[];inputs.touchUntil=0;if(core._rr_touch)core._rr_touch(0,0,0);
  }
  loaded = true;
  paint();
  send('ready', {
    text : (restored?'State restored, paused. ':'') + f.name + ' loaded. ' + (core.compatProfile || '') + ' Ready in '+((performance.now()-bootBegan)/1000).toFixed(2)+' s. Press Run.'
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
    if(saving&&!['input','turbo','hold'].includes(m.type)){send('message',{text:'Finishing the state save…'});return;}
    if(m.type==='seek'){await seekReplay(m);return;}
    if(m.type==='cancel-seek'){seekGeneration++;return;}
    if(m.type==='save'){await saveState();return;}
    if(m.type==='pixel'){if(!capture||m.capture!==capture.id)return;send('pixel',{capture:capture.id,x:m.x,y:m.y,evidence:pixelEvidence(m.x,m.y),request:m.request});return;}
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
