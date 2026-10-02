import {backendAssets,bindOwnedIdentity,matchingPreparedKnowledge} from './core-backend.js';
import {prepareStart} from './prepared-start.js';
import {createExperimentService} from './experiment-worker.js';
import {create3DOInspector} from './threedo-inspector.js';
import {createDOSInspector} from './dos-inspector.js';
import {identifyDOSFiles} from './dos-knowledge.js';
import {createTourService} from './tour-worker.js';
import {createDebugService} from './debug-worker.js';
import {createExecutionGate} from './execution-gate.js';
import {identifySingleImage} from './knowledge-model.js';
import {knowledgePackages} from './knowledge-data.js';
import {createMemoryService} from './memory-worker.js';
import {packState, unpackState, digest} from './state.js';
import {FrameClock} from './pacing.js';
import {InputQueue} from './input.js';
import {selectMedia, sha256File} from './media.js';
import {selectDreamcastMedia} from './dc-media.js';
import {selectDOSMedia,dosFiles} from './dos-media.js';
import {platforms} from './platforms.js';

let profileBase = {}, core, platform, session, loaded = false, running = false,
    epoch = 0, frames = 0, turbo = false, files, maxCall = 0, runMs = 0,
    paintMs = 0, lastPaint = 0;
const inputQueue = [];
const appliedKeys=new Set();
const mediaHashes=new WeakMap();
function mediaHash(file){if(!mediaHashes.has(file))mediaHashes.set(file,sha256File(file));return mediaHashes.get(file);}
let cacheReply=null;
let saving=false,driveDebugService=null,debugService=null,tourService=null,experimentService=null;
const executionGate=createExecutionGate();
function debugActive(){return !!(debugService?.active()||driveDebugService?.active());}
let memoryService=null,memoryRecording=false,memoryStop=false,memoryBusy=false;
let coreCapabilities={}, backend='production';
let bootOptions, firmwareIdentity=[], mediaIdentity, coreIdentity;
async function identities(){
 if(!mediaIdentity){send("message",{text:"Verifying local media identity…"});mediaIdentity=[];const names=platform==='dos'?dosFiles(files).map(e=>e.path):files.map(f=>f.name);for(const [i,f] of files.entries())mediaIdentity.push({name:names[i],size:f.size,sha256:await mediaHash(f)});mediaIdentity.sort((a,b)=>a.name.localeCompare(b.name));}
 return {media:mediaIdentity,firmware:firmwareIdentity,core:coreIdentity,configuration:{compatibility:bootOptions.compatibility!==false,customFirmware:!!bootOptions.firmware?.some(Boolean),...(platform==='dos'?{executable:bootOptions.executable}:{}),...(backend==='owned'?{c64Backend:backend}:{})}};
}
function queueState(){return {...inputs,appliedKeys:[...appliedKeys],pulses:[...inputs.pulses],down:[...inputs.down],pending:inputQueue,lastButtons,lastX,lastY,inputSequence,lastInputStep};}
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
  if(platform==='c64'&&core._rr_drive_status){const d=json('_rr_drive_status');if(d.enabled){s.drive=d;s.track=d.track;}}
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
  const start = performance.now(), p = core._rr_frame(), s = status();
  frames = s.frames;
  const w = platform === 'c64' ? 392 : s.width || 320,
        h = platform === 'c64' ? 272 : s.height || 240,
        pixels = core.HEAPU8.slice(p, p + w * h * 4);
  paintMs += performance.now() - start;
  send('state', {
    state : s,
    capabilities:coreCapabilities,backend,drive:platform==='c64'&&core._rr_drive_debug_snapshot&&s.drive?JSON.parse(core.UTF8ToString(core._rr_drive_debug_snapshot(-1))):null,
    width : w,
    height : h,
    pixels : pixels.buffer,
    running,
    maxCall,
    runMs,
    paintMs,
    heap : core.HEAPU8.length,
    inputDeferred : platform === '3do' && core.deferInput && s.frames < 300,
    capturing,saving,memoryRecording,debugBusy:!!(memoryBusy||debugActive()||tourService?.active()||experimentService?.active()),experimentOwned:!!experimentService?.owns(),
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
    else if (platform === 'n64'||platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox')
      core._rr_pad(m.buttons, m.x, m.y);
    else
      core._rr_pad(platform === 'ps1' ? (~m.buttons) & 65535 : m.buttons);
    lastButtons = m.buttons;
    lastX = m.x;
    lastY = m.y;
  }
  if (platform === 'ds'||platform==='3ds')core._rr_touch(m.touch.x,m.touch.y,+m.touch.down);
  if(platform==='dos')core._rr_mouse(m.touch.x,m.touch.y,((m.buttons>>8)&3));
  if (platform === 'amiga') {const mouse=inputs.mouseForFrame(s.frames);if(mouse.x||mouse.y)core._rr_mouse(mouse.x,mouse.y);}
  if (platform === 'c64'||platform==='dos'||platform==='amiga')
    for (const [code, down] of m.keys){
      core._rr_key(code, down);if(down)appliedKeys.add(code);else appliedKeys.delete(code);
    }
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
  else if(platform==='ds'||platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox'||platform==='gb'||platform==='gg'||platform==='gba'||platform==='dos'||platform==='amiga')check(core._rr_run(10000)>=0);
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
function coreState(){const n=core._rr_state_save();check(n);const limit=(platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox'||platform==='dos')?128:32;if(n>limit*1024*1024)throw Error(`Capture checkpoint exceeds the ${limit} MiB budget`);const p=core._rr_state_data();return core.HEAPU8.slice(p,p+n);}
function cancelCapture(starting=false){
 if(capturing){core._rr_capture_end();finishCaptureProfile();capturing=false;}
 capture=null;captureGeneration++;seekGeneration++;send('capture-cleared',{starting});
}
async function captureNext(){
 running=false;const id=++epoch;cancelCapture(true);const generation=captureGeneration;capturing=true;captureRunMs=maxCaptureCall=0;captureProfileStart=Object.fromEntries(json('_rr_profile').buckets.map(b=>[b.name,b.ms]));
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
   const captureStarted=core._rr_capture_begin();if(platform==='ds'||platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox'||platform==='gb'||platform==='gg'||platform==='gba'||platform==='dos'||platform==='amiga')check(captureStarted);
   const captureFields=platforms[platform].captureFields??((platform==='ps1'||platform==='ps2')?4:(platform==='gc'||platform==='dc'||platform==='amiga')?3:platform==='ds'?2:1);
   if(platform==='dos'){while(!json('_rr_capture_info').ready)await boundary('Recording RAM rendering and VGA copies.');}
   else for(let field=0;field<captureFields;field++)await boundary(captureFields>1?'Recording display and double-buffer producer context.':'Recording next complete interval.');
   const ended=core._rr_capture_end();if(platform==='ds'||platform==='3ds'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox'||platform==='gb'||platform==='gg'||platform==='gba'||platform==='dos'||platform==='amiga')check(ended);const captureProfile=finishCaptureProfile();capturing=false;
   const p=core._rr_frame(),endState=coreState(),end=status(),info=json('_rr_capture_info'),replay=json('_rr_replay_begin');
   const w=platform==='c64'?392:end.width||320,h=platform==='c64'?272:end.height||240;
   const pixels=core.HEAPU8.slice(p,p+w*h*4);
   // Validate visible output rather than assuming every memory space is watched.
   while(!core._rr_replay_seek(replay.count)){await sleep(0);if(id!==epoch)return;}
   const rp=core._rr_replay_frame();
   replay.complete=replay.complete&&pixels.every((v,i)=>core.HEAPU8[rp+i]===v);
   const tileset=platform==='gba'?tilesetSnapshot():null,vram=vramSnapshot();
   core._rr_replay_seek(0);
   const frameHash=await digest(pixels);
   if(id!==epoch)return;
   capture={id:generation,startState,endState,start,end,input,width:w,height:h,pixels,info,frameHash};
   paint();send('capture',{id:generation,tileset,vram,...(['gb','gg','c64','amiga'].includes(platform)&&core._rr_raster_info?{raster:json('_rr_raster_info')}:{}),replay,start,end,width:w,height:h,info,frameHash,elapsedMs:performance.now()-began,checkpointBytes:startState.length+endState.length,profile:captureProfile,runMs:captureRunMs,maxCall:maxCaptureCall});
   send('message',{text:info.overflow?'Paused. Capture limit reached; some evidence is missing.':platform==='dos'?'Paused. RAM rendering and VGA copies are ready to inspect.':'Paused. A complete display interval is ready to inspect.'});
 }catch(e){if(id===epoch){if(capturing)core._rr_capture_end();finishCaptureProfile();capturing=false;capture=null;paint();send('error',{text:String(e)});}}
}
function pixelEvidence(x,y){const p=jsonPixel(x,y);for(const c of p.contributors||[])c.replayStep=core._rr_replay_for_write(platform==='c64'?y:c.id);return p;}
function tilesetSnapshot(){if(!core._rr_tileset_size)return null;const size=core._rr_tileset_size();if(!size)return null;const p=core._rr_tileset_data();return p?core.HEAPU8.slice(p,p+size).buffer:null;}
function vramSnapshot(){if(!['ps1','dc'].includes(platform)||!core._rr_vram_size)return null;const size=core._rr_vram_size();if(!size)return null;const p=core._rr_vram_data();return {info:json('_rr_vram_info'),memory:core.HEAPU8.slice(p,p+size).buffer};}
async function seekReplay(m){
 if(!capture||capture.id!==m.capture)return;
 const gen=++seekGeneration,c=capture,began=performance.now();let progress=began;
 while(!core._rr_replay_seek(m.step)){
  if(performance.now()-progress>100){progress=performance.now();send('seek-progress',{capture:c.id,request:m.request});}
  await sleep(0);if(gen!==seekGeneration||capture!==c)return;
 }
 if(gen!==seekGeneration||capture!==c)return;
 if(platform==='dos'&&core._rr_replay_view)core._rr_replay_view(+!!m.reveal);
 const p=core._rr_replay_frame(),pixels=core.HEAPU8.slice(p,p+c.width*c.height*4);
 send('seek',{capture:c.id,request:m.request,pixels:pixels.buffer,tileset:platform==='gba'?tilesetSnapshot():null,vram:vramSnapshot(),info:json('_rr_replay_info'),elapsedMs:performance.now()-began});
}
function rasterSeek(m){
 if(!['gb','gg','c64','amiga'].includes(platform)||!capture||capture.id!==m.capture)return;
 if(platform==='amiga')core._rr_raster_plane(m.plane??-1);
 const info=JSON.parse(core.UTF8ToString(core._rr_raster_seek(m.line)));
 const layers=[];
 if(!info.error)for(let panel=0;panel<3;panel++){const p=core._rr_raster_frame(panel);layers.push(core.HEAPU8.slice(p,p+capture.width*capture.height*4).buffer);}
 postMessage({type:'raster-seek',session,capture:capture.id,request:m.request,info,layers,tileset:info.error?null:tilesetSnapshot()},layers);
}
async function boot(m) {
  const bootBegan=performance.now();
  bootOptions=m;session=m.session;
  const restored=m.stateFile?await unpackState(m.stateFile):null;
  platform = m.platform;
  if(restored&&restored.meta.platform!==platform)throw Error("This state belongs to another console");
  inputs = new InputQueue(platforms[platform].hz);
  if(platform==='amiga')inputs.mouseMask=96;if(platform==='dos')inputs.mouseMask=768;
  if(platform==='ds'||platform==='3ds')inputs.hold=3/platforms[platform].hz;
  session = m.session;
  files = m.files;
  if (!files?.length)
    throw Error('Select a game image');
  send('message', {text : 'Loading emulator…'});
  backend=m.backend||'production';
  const assets=backendAssets(platform,backend);backend=assets.backend;
  const manifestResponse=await fetch(assets.manifest);
  if(!manifestResponse.ok)throw Error(assets.owned?'Owned C64 manifest download failed.':'Emulator manifest download failed');
  const manifest=await manifestResponse.json();coreIdentity=manifest[assets.key];
  const response=await fetch(assets.wasm);
  if(!response.ok)throw Error('Emulator binary download failed');
  const wasmBinary=new Uint8Array(await response.arrayBuffer());
  if(await digest(wasmBinary)!==coreIdentity)throw Error('Emulator build mismatch. Rebuild the development core or reload the matching release.');
  const factory = (await import(assets.module)).default;
  core = await factory({wasmBinary});
  coreCapabilities=core._rr_capabilities?json('_rr_capabilities'):{};
  if(assets.owned&&coreCapabilities.core!=='owned-c64')throw Error('Expected the owned C64 core');
  const dcMedia=platform==='dc'?await selectDreamcastMedia(files):null;
  const dosMedia=platform==='dos'?selectDOSMedia(files,m.executable):null;
  const f = dosMedia?dosMedia.entries.find(e=>e.path===dosMedia.entry).file:dcMedia?dcMedia.file:await selectMedia(files);
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
    if(f.size>2*1024*1024)throw Error('C64 media exceeds the 2 MiB limit');
    const disk=/\.(d64|g64)$/i.test(f.name);let driveRom=null;
    if(disk){
      if(!assets.owned)throw Error('D64/G64 execution requires the owned C64 backend');
      if(!m.driveFirmware||m.driveFirmware.size!==16384)throw Error('Select the 16 KiB 1541 firmware for disk execution');
      const bytes=new Uint8Array(await m.driveFirmware.arrayBuffer());driveRom=await digest(bytes);firmwareIdentity.push(driveRom);core.HEAPU8.set(bytes,core._rr_input());check(core._rr_drive_rom(bytes.length));
    }
    core.HEAPU8.set(new Uint8Array(await f.arrayBuffer()),core._rr_input());
    check(disk?core._rr_disk(f.size,1):core._rr_tape(f.size));
    if(assets.owned)bindOwnedIdentity(core,{core:coreIdentity,firmware:firmwareIdentity.slice(0,3),driveRom,...(disk?{disk:await mediaHash(f)}:{tape:await mediaHash(f)})});
    core._rr_trace(0, 0, 0);
  } else if(platform==='dos'){
    core.gameFiles=files;const put=s=>{const b=new TextEncoder().encode(s),p=core._rr_input(b.length);check(p);core.HEAPU8.set(b,p);return b.length;};
    for(const e of dosMedia.entries)check(core._rr_file(e.index,put(e.path),e.file.size));
    check(core._rr_init(put(dosMedia.entry),+(m.compatibility!==false)));
  } else if (platform === 'amiga') {
    if(f.size!==901120)throw Error('Select a standard 880 KiB ADF disk image');
    const custom=m.firmware?.[0];let rom;
    if(custom)rom=new Uint8Array(await custom.arrayBuffer());
    else {
      const [entry]=await (await fetch('./firmware/amiga/manifest.json')).json();
      const response=await fetch('./firmware/amiga/'+entry.file);
      if(!response.ok)throw Error('Kickstart download failed');
      rom=new Uint8Array(await response.arrayBuffer());
      if(await digest(rom)!==entry.sha256)throw Error('Kickstart hash mismatch');
    }
    firmwareIdentity=[await digest(rom)];
    let p=core._rr_firmware(rom.length);check(p);core.HEAPU8.set(rom,p);
    p=core._rr_input(f.size);check(p);core.HEAPU8.set(new Uint8Array(await f.arrayBuffer()),p);check(core._rr_init(f.size));
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
  } else if (platform === 'ps1'||platform==='psp'||platform==='gc'||platform==='ps2'||platform==='dc'||platform==='xbox') {
    if (f.size > ((platform==='ps2'||platform==='xbox')?16*1024**3:0xffffffff))
      throw Error('Disc exceeds the supported size');
    core.discFile = f;
    if(dcMedia)for(const t of dcMedia.tracks)check(core._rr_disc_track(t.number,t.lba,t.offset,t.length,+t.data));
    check(core._rr_init_file(f.size) > 0);
    if(platform==='ps2'){const rom=m.firmware?.[0];if(rom){if(![4,8].includes(rom.size/1048576))throw Error('PS2 BIOS must be a 4 or 8 MiB ROM image');const b=new Uint8Array(await rom.arrayBuffer());firmwareIdentity=[await digest(b)];const p=core._rr_input(b.length);check(p);core.HEAPU8.set(b,p);check(core._rr_bios(b.length));}check(core._rr_boot());}
  } else if(platform==='ds'||platform==='3ds'||platform==='gb'||platform==='gg'||platform==='gba'){
    const limit=platform==='gba'?32:platform==='gb'?2:platform==='gg'?4:platform==='ds'?512:1024;
    if(f.size>limit*1024*1024+(platform==='gg'?512:0))throw Error(`Cartridge exceeds ${limit} MiB`);
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
    if(platform==='c64'||platform==='dos')for(let k=0;k<(platform==='dos'?128:256);k++)core._rr_key(k,0);
    if(platform==='dos'){core._rr_pad(0);core._rr_mouse(0,0,0);appliedKeys.clear();}
    if(platform==='amiga'){for(const code of q.appliedKeys||inputs.down.keys())core._rr_key(code,0);appliedKeys.clear();inputs.mousePending={x:0,y:0};inputs.mouseMask=96;inputs.mouseButtons=0;inputs.mouseButtonEvents=[];inputs.mouseButtonCursor=0;}
    inputs.keys=[];inputs.down.clear();inputs.touch={x:0,y:0,down:false};inputs.touchEvents=[];inputs.touchUntil=0;if(core._rr_touch)core._rr_touch(0,0,0);
  }
  if(m.preparedStart){
    if(platform!=='c64'||restored)throw Error('Unsupported prepared start request');
    const match=identifySingleImage(platform,{size:f.size,sha256:await mediaHash(f)}),pkg=knowledgePackages.find(p=>p.id===match.packageId),recipe=pkg?.knowledge.preparedStarts?.[m.preparedStart];
    if(match.status!=='matched'||!recipe?.releases.includes(match.releaseId))throw Error('No verified prepared start for this exact image');
    const identity=await identities(),key=await digest(new TextEncoder().encode(JSON.stringify({package:pkg.sourceSHA256,id:m.preparedStart,...identity})));
    const cached=await new Promise(resolve=>{const timer=setTimeout(()=>{cacheReply=null;resolve(null);},1500);cacheReply=bytes=>{clearTimeout(timer);resolve(bytes);};send('prepared-cache-get',{key});});
    const result=await prepareStart({core,recipe,identity,cached,sleep,progress:cycles=>send('message',{text:'Preparing '+recipe.title+' · '+(cycles/985248).toFixed(1)+' emulated seconds. You can cancel without losing the current session.'})});
    if(!result.cached)send('prepared-cache-put',{key,bytes:result.bytes});
    send('message',{text:result.cached?'Loaded verified local lesson checkpoint.':'Verified lesson checkpoint prepared and saved locally.'});
  }
  memoryService=createMemoryService({core,platform,files:[f],status});
  if(['c64','dos','3do'].includes(platform)&&core._rr_debug_snapshot){
    const identity=platform==='dos'?await identifyDOSFiles(dosMedia.entries,knowledgePackages,mediaHash,dosMedia.entry):identifySingleImage(platform,{size:f.size,sha256:await mediaHash(f)});
    const pkg=knowledgePackages.find(p=>p.id===identity.packageId);
    const knowledge=matchingPreparedKnowledge(pkg?{...identity,data:pkg.knowledge}:identity,{core:coreIdentity,firmware:firmwareIdentity});
    const adapter=platform==='dos'?createDOSInspector(core,knowledge):platform==='3do'?create3DOInspector(core,knowledge):{};
    debugService=createDebugService({core,send,sleep,paint,applyInputs,generation:session,platform,...adapter,
      knowledge,
      onStart:()=>{cancelCapture();memoryService.stopLive();},
      busy:includePlay=>!!(experimentService?.active()||tourService?.active()||driveDebugService?.active()||(executionGate.owner&&(includePlay||!['run','step'].includes(executionGate.owner)))||saving||memoryBusy||memoryRecording||capturing||(includePlay&&running))});
    if(platform==='c64'&&core._rr_drive_debug_snapshot&&json('_rr_drive_status').enabled){
      const driveCore=new Proxy(core,{get:(target,key)=>({'_rr_debug_snapshot':target._rr_drive_debug_snapshot,'_rr_debug_begin':target._rr_drive_debug_begin,'_rr_debug_run':target._rr_drive_debug_run,'_rr_cycle':target._rr_drive_cycle}[key]??target[key])});
      driveDebugService=createDebugService({core:driveCore,platform:'1541',generation:session,knowledge:{status:'unknown'},send:(type,m)=>send('drive-'+type,m),sleep,paint,applyInputs,
        busy:includePlay=>!!(debugService?.active()||experimentService?.owns()||tourService?.active()||executionGate.owner||saving||memoryBusy||memoryRecording||capturing||(includePlay&&running)),onStart:()=>{cancelCapture();memoryService.stopLive();}});
    }
    if(platform==='c64')experimentService=createExperimentService({core,knowledge,generation:session,send,sleep,paint,snapshot:()=>debugService.snapshot(),
      busy:()=>!!(running||executionGate.owner||debugActive()||tourService?.active()||saving||memoryBusy||memoryRecording||capturing),captureHost:queueState,
      clearInput:()=>{inputs=new InputQueue(50);inputQueue.length=0;appliedKeys.clear();lastButtons=lastX=lastY=0;for(let k=0;k<256;k++)core._rr_key(k,0);core._rr_joystick(1,0);core._rr_joystick(2,0);},
      restoreHost:q=>{inputs=new InputQueue(50);Object.assign(inputs,q,{pulses:new Map(q.pulses),down:new Map(q.down)});inputQueue.length=0;inputQueue.push(...q.pending);appliedKeys.clear();for(const k of q.appliedKeys)appliedKeys.add(k);lastButtons=q.lastButtons;lastX=q.lastX;lastY=q.lastY;inputSequence=q.inputSequence;lastInputStep=q.lastInputStep;},
      onStart:()=>{cancelCapture();memoryService.stopLive();tourService?.invalidate('An explicit experiment replaced this tour context.');}});
    tourService=createTourService({core,knowledge,generation:session,send,sleep,paint,...adapter,maxCheckpointBytes:platform==='dos'?128*1024*1024:platform==='3do'?32*1024*1024:4*1024*1024,
      snapshot:()=>debugService.snapshot(),busy:()=>!!(experimentService?.owns()||running||executionGate.owner||debugActive()||saving||memoryBusy||memoryRecording||capturing),
      onStart:()=>{cancelCapture();memoryService.stopLive();if(platform==='3do'&&lastButtons)core._rr_pad(0);inputs=new InputQueue(platforms[platform].hz);if(platform==='dos')inputs.mouseMask=768;inputQueue.length=0;appliedKeys.clear();lastButtons=lastX=lastY=0;if(platform!=='3do')for(let k=0;k<(platform==='dos'?128:256);k++)core._rr_key(k,0);if(platform==='dos'){core._rr_pad(0);core._rr_mouse(0,0,0);}else if(platform==='c64')core._rr_joystick(2,0);},
      onRestore:()=>{inputs=new InputQueue(platforms[platform].hz);if(platform==='dos')inputs.mouseMask=768;inputQueue.length=0;appliedKeys.clear();lastButtons=lastX=lastY=0;}
    });
  }
  loaded = true;
  paint();
  send('ready', {
    capabilities:coreCapabilities,backend,
    text : (restored?'State restored, paused. ':'') + f.name + ' loaded. ' + (assets.owned?'Owned C64 core. ':'') + (core.compatProfile || '') + ' Ready in '+((performance.now()-bootBegan)/1000).toFixed(2)+' s. Press Run.'
  });
}
onmessage = async ({data : m}) => {
  let ownership=null;
  try {
    if(m.type==='prepared-cache-reply'&&m.session===session){cacheReply?.(m.bytes);cacheReply=null;return;}
    if (m.type === 'load') {
      await boot(m);
      return;
    }
    if (m.session !== session || !loaded)
      return;
    if(m.type==='capture-render'&&coreCapabilities.renderCapture===false){send('message',{text:'Rendering capture is not available in this development core.'});return;}
    if(m.type.startsWith('experiment-')){if(experimentService)await experimentService.request(m);else send('experiment-rejected',{generation:session,request:m.request,text:'Experiments are unavailable for this core.'});return;}
    if(experimentService?.owns()){
      if(['pause','hold','cancel-capture'].includes(m.type)){experimentService.cancel();return;}
      if(m.type==='input'||m.type==='turbo')return;
      const readOnly=['debug-snapshot','debug-capabilities','memory-snapshot','memory-page','memory-detail','memory-overview'];
      if(experimentService.active()||!readOnly.includes(m.type)){send(m.type.startsWith('tour-')?'tour-rejected':m.type.startsWith('debug-')?'debug-result':m.type.startsWith('memory-')?'memory-error':'message',{generation:session,request:m.request,reason:'rejected',text:'Experiment owns the session. Return or explicitly continue the modified branch first.'});return;}
    }
    if(m.type.startsWith('tour-')){
      if(tourService)await tourService.request(m);
      else send('tour-rejected',{generation:session,request:m.request,text:'Tours are unavailable for this core.'});
      return;
    }
    if(tourService?.active()){
      if(['pause','hold','cancel-capture'].includes(m.type)){tourService.cancel();return;}
      send(m.type.startsWith('memory-')?'memory-error':m.type.startsWith('debug-')?'debug-result':'message',{generation:session,request:m.request,reason:'rejected',text:'Tour is running. Cancel it before another operation.'});return;
    }
    if(['run','step','seek','capture-render','memory-record','tape','debug-normalize','debug-step','debug-until','debug-over','debug-out','debug-watch'].includes(m.type))tourService?.invalidate();
    if(m.type==='input'&&((m.buttons??0)!==lastButtons||(m.keys||[]).some(([key,down])=>!!down!==appliedKeys.has(key))))tourService?.invalidate();
    if(m.type.startsWith('drive-debug-')){if(!driveDebugService){send('drive-debug-result',{request:m.request,generation:session,reason:'unsupported',text:'Attach a disk and 1541 firmware first.'});return;}if(!['drive-debug-snapshot','drive-debug-capabilities'].includes(m.type))tourService?.invalidate();await driveDebugService.request({...m,type:m.type.slice(6)});return;}
    if(m.type.startsWith('debug-')){
      if(!debugService){send('debug-result',{request:m.request,reason:'unsupported',text:'Instruction debugging is unavailable for this core.'});return;}
      await debugService.request(m);return;
    }
    if(debugActive()){
      if(['pause','hold','cancel-capture'].includes(m.type)){debugService?.cancel();driveDebugService?.cancel();return;}
      if(!['input','turbo'].includes(m.type)){
        send(m.type.startsWith('memory-')?'memory-error':'message',{request:m.request,text:'Debugger job is active. Cancel it before another operation.'});return;
      }
    }
    const playOwner=['run','step'].includes(executionGate.owner);
    // These existing controls explicitly pause play before taking ownership.
    if(playOwner&&['pause','hold','cancel-capture','save','capture-render','memory-snapshot','memory-record'].includes(m.type)){
      running=false;++epoch;executionGate.cancel(['run','step']);
    }
    const liveRead=playOwner&&m.type.startsWith('memory-')&&!['memory-record','memory-snapshot'].includes(m.type);
    const exclusive=['run','step','save','seek','capture-render','memory-snapshot','memory-record','memory-live-snapshot'];
    if(executionGate.owner&&!liveRead&&!(playOwner&&m.type==='tape')&&!['input','turbo','pause','hold','memory-stop','cancel-capture','cancel-seek'].includes(m.type)){
      send(m.type.startsWith('memory-')?'memory-error':'message',{request:m.request,text:`Machine is busy: ${executionGate.owner}.`});return;
    }
    if(exclusive.includes(m.type)&&!liveRead)ownership=executionGate.acquire(m.type);
    if(saving&&!['input','turbo','hold'].includes(m.type)){send('message',{text:'Finishing the state save…'});return;}
    if(m.type==='memory-stop'){memoryStop=true;return;}
    if(memoryBusy||memoryRecording){if(m.type==='pause'||m.type==='hold')memoryStop=true;if(!['input','turbo','tape'].includes(m.type))return;}
    if(m.type.startsWith('memory-')){await memoryRequest(m);return;}
    if(m.type==='raster-seek'){rasterSeek(m);return;}
    if(m.type==='blit-seek'){
      if(platform!=='amiga'||!capture||m.capture!==capture.id)return;
      const info=JSON.parse(core.UTF8ToString(core._rr_blit_seek(m.index))),layers=[];
      if(!info.error)for(let p=0;p<6;p++){const at=core._rr_blit_frame(p),n=p<4?info.width*info.height*4:640*256*4;layers.push(core.HEAPU8.slice(at,at+n).buffer);}
      postMessage({type:'blit-seek',session,capture:capture.id,request:m.request,info,layers},layers);return;
    }
    if(m.type==='blit-pixel'){
      if(platform!=='amiga'||!capture||m.capture!==capture.id)return;
      send('blit-pixel',{capture:capture.id,request:m.request,evidence:JSON.parse(core.UTF8ToString(core._rr_blit_pixel(m.x,m.y)))});return;
    }
    if(m.type==='raster-pixel'){if(!['gb','gg','c64','amiga'].includes(platform)||!capture||m.capture!==capture.id)return;const p=core._rr_raster_pixel(m.panel,m.x,m.y);send('raster-pixel',{capture:capture.id,request:m.request,evidence:JSON.parse(core.UTF8ToString(p))});return;}
    if(m.type==='seek'){await seekReplay(m);return;}
    if(m.type==='cancel-seek'){seekGeneration++;return;}
    if(m.type==='save'){await saveState();return;}
    if(m.type==='pixel'){if(!capture||m.capture!==capture.id)return;send('pixel',{capture:capture.id,x:m.x,y:m.y,evidence:pixelEvidence(m.x,m.y),request:m.request});return;}
    if(m.type==='source'||m.type==='resource'){
      if(!capture||m.capture!==capture.id)return;
      const fn=m.type==='source'?core._rr_source:core._rr_resource;
      if(!fn)return;
      const p=m.type==='source'?fn(m.address,m.size,m.before,m.expected):fn(m.resource,m.offset);
      const evidence=JSON.parse(core.UTF8ToString(p));
      if(platform==='dos')for(const c of evidence.contributors||[])c.replayStep=core._rr_replay_for_write(c.id);
      send(m.type,{capture:capture.id,evidence,request:m.request});return;
    }
    if(m.type==='cancel-capture'||m.type==='hold'){running=false;++epoch;cancelCapture();paint();send('message',{text:'Paused.'});return;}
    if (m.type === 'run' || m.type === 'step') {
      if (running)
        return;
      cancelCapture();running = true;
      await pump(++epoch, m.type === 'step');
    } else if (m.type === 'capture-render') {
      await captureNext();
    } else if (m.type === 'pause') {
      running=false;++epoch;cancelCapture();paint();send('message',{text:'Paused. Open Memory to inspect physical storage.'});
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
  } finally {if(ownership)executionGate.release(ownership);}
};

async function memoryRequest(m){
 const respond=overview=>send('memory-overview',{request:m.request,overview});
 try{
  if(m.type==='memory-snapshot'||m.type==='memory-record'){
   memoryBusy=true;running=false;++epoch;cancelCapture();paint();
   if(m.type==='memory-snapshot'){respond(await memoryService.snapshot());return;}
   memoryRecording=true;memoryStop=false;await memoryService.begin(m.fetches);memoryBusy=false;paint();
   const time=s=>platform==='c64'?s.steps/985248:s.seconds??s.frames/platforms[platform].hz;
   const start=time(status()),duration=Math.max(.01,Math.min(5,Number(m.duration)||1));
   let last=0;
   while(!memoryStop&&time(status())-start<duration&&core._rr_activity_count()<524288){
    tick();const now=performance.now();if(now-last>100){send('memory-progress',{text:`Recording ${(time(status())-start).toFixed(2)} seconds · ${core._rr_activity_count().toLocaleString()} accesses`});last=now;}await sleep(0);
   }
   respond(memoryService.finish());paint();return;
  }
  if(m.type==='memory-live-end'){memoryService.stopLive();return;}
  if(m.type==='memory-live-snapshot'){respond(await memoryService.liveSnapshot(m.scale,m.window,m.fetches));return;}
  if(m.type==='memory-overview')respond(memoryService.overview(m.scale,m.window));
  if(m.type==='memory-seek'){memoryService.seek(m.position);respond(memoryService.overview(m.scale,m.window));}
  if(m.type==='memory-page'){
   const page=memoryService.page(m.region,m.offset);
   if(page?.id===m.snapshot){if(m.region==='tape'){page.pulseStart=memoryService.pulseForOffset(page.offset);page.pulseEnd=memoryService.pulseForOffset(page.offset+page.bytes.length)+1;}send('memory-page',{request:m.request,page});}
  }
  if(m.type==='memory-map')send('memory-map',{request:m.request,region:m.region,offset:memoryService.mapOffset(m.region,m.pixel,m.scale)});
  if(m.type==='memory-pulses'){const pulses=memoryService.pulses(m.start,m.count);if(pulses?.id===m.snapshot)send('memory-pulses',{request:m.request,pulses});}
  if(m.type==='memory-detail')send('memory-detail',{request:m.request,snapshot:m.snapshot,detail:memoryService.detail(m.region,m.offset)});
 }catch(e){send('memory-error',{request:m.request,text:String(e)});}
 finally{const wasBusy=memoryBusy,recorded=memoryRecording;if(recorded)core._rr_activity_end();memoryBusy=false;memoryRecording=false;if(recorded||wasBusy)paint();}
}
