// Serve the repository root, after building site/emulators/cores/c64-owned.
// Exercises real DOM, workers, fetches, identity binding and checkpoint transport.
import {packState,unpackState} from '../../../site/emulators/state.js';
const layoutKey='rr.viewport.v1.c64',previousLayout=localStorage.getItem(layoutKey);localStorage.removeItem(layoutKey);
const result=document.querySelector('#result'),frame=document.querySelector('#app');
const assert=(ok,text)=>{if(!ok)throw Error(text);},sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function until(fn,label){const began=performance.now();while(!fn()){if(performance.now()-began>30000)throw Error('Timeout: '+label);await sleep(20);}return fn();}
const workers=[];
function client(){const worker=new Worker('/site/emulators/worker.js',{type:'module'}),messages=[];workers.push(worker);worker.onmessage=e=>messages.push(e.data);worker.onerror=e=>messages.push({type:'error',text:e.message});let request=0;return {worker,messages,send:(type,data={})=>worker.postMessage({type,session:1,request:++request,...data}),wait:type=>until(()=>messages.find(m=>m.type===type||m.type==='error'),type).then(m=>{if(m.type==='error')throw Error(m.text);return m;})};}
try{
 const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;
 const files=[new File([tape],'authored.tap')],c=client();c.send('load',{platform:'c64',backend:'owned',files});
 const ready=await c.wait('ready');assert(ready.backend==='owned'&&ready.capabilities.renderCapture===true,'owned capabilities');
 c.send('capture-render');const ownedCapture=await c.wait('capture');assert(ownedCapture.replay.complete&&ownedCapture.raster.complete,'owned complete rendering replay');c.send('raster-seek',{capture:ownedCapture.id,line:271});const raster=await c.wait('raster-seek');assert(raster.layers.length===3&&raster.info.complete,'frozen owned layers');c.send('raster-pixel',{capture:ownedCapture.id,panel:2,x:0,y:0});assert((await c.wait('raster-pixel')).evidence.contributors.length>0,'owned pixel provenance');
 c.send('step');await until(()=>c.messages.some(m=>m.type==='state'&&m.state.cycle>0&&!m.running),'frame step');
 c.send('memory-record',{duration:5,fetches:true});await until(()=>c.messages.some(m=>m.type==='state'&&m.memoryRecording),'recording');c.send('memory-stop');
 await c.wait('memory-overview');await until(()=>c.messages.at(-1)?.type==='state'&&!c.messages.at(-1).memoryRecording,'recording stopped');
 c.send('save');const saved=await c.wait('saved'),file=new File([saved.bytes],'owned.rrstate'),decoded=await unpackState(file);
 assert(decoded.meta.configuration.c64Backend==='owned','state backend identity');
 const cycles=c.messages.filter(m=>m.type==='state').at(-1).state.cycle;
 const restored=client();restored.send('load',{platform:'c64',backend:'owned',files,stateFile:file});await restored.wait('ready');
 assert(restored.messages.find(m=>m.type==='state').state.cycle===cycles,'checkpoint cycle restored');
 restored.send('debug-snapshot',{protocol:1,generation:1});const snap=(await restored.wait('debug-snapshot')).snapshot;
 restored.send(snap.boundary?'debug-step':'debug-normalize',{protocol:1,generation:1,snapshotId:snap.snapshotId,cycle:snap.cycle,bank:snap.bank});
 const step=await restored.wait('debug-result');assert(['instruction','boundary','interrupt-entry'].includes(step.reason),'debug execution '+step.reason);
 const production=client();production.send('load',{platform:'c64',files});assert((await production.wait('ready')).backend==='production','production remains default');
 production.send('capture-render');await production.wait('capture');production.send('save');const productionFile=new File([(await production.wait('saved')).bytes],'production.rrstate');assert(!(await unpackState(productionFile)).meta.configuration.c64Backend,'production metadata stays compatible');
 const wrong=client();wrong.send('load',{platform:'c64',files,stateFile:file});const rejection=await until(()=>wrong.messages.find(m=>m.type==='error'),'cross-core rejection');assert(rejection.text.includes('State core does not match'),'cross-core identity rejected');
 const mismatch=await packState({...decoded.meta,configuration:{...decoded.meta.configuration,c64Backend:'production'}},decoded.payload);
 const bad=client();bad.send('load',{platform:'c64',backend:'owned',files,stateFile:new File([mismatch],'bad.rrstate')});
 assert((await until(()=>bad.messages.find(m=>m.type==='error'),'configuration rejection')).text.includes('State configuration does not match'),'configuration rejected');
 // Source UI deliberately opts in via its URL. No release bundle is modified.
 frame.srcdoc=`<!doctype html><link rel="stylesheet" href="/site/emulators/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>window.received=[];const BaseWorker=Worker;window.Worker=class extends BaseWorker{constructor(...a){super(...a);this.addEventListener('message',e=>received.push(e.data));}};<\/script><script type="module">import {mountEmulatorSession} from '/site/emulators/emulator-session.js';window.ui=mountEmulatorSession('c64');<\/script>`;
 await until(()=>frame.contentDocument?.querySelector('#load'),'UI mounted');
 const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id);
 // Restore carries the explicit backend, as actual system suspension does.
 w.ui.restore({files,firmware:null,compatibility:true,backend:'owned',stateFile:file});
 await until(()=>w.received.some(m=>m.type==='ready'),'UI ready');
 assert(!$('run').disabled&&!$('render-capture').disabled,'play and rendering capture enabled');
 for(const selector of [...d.querySelectorAll('#viewport-root .viewport-header select')]){selector.value='render';selector.dispatchEvent(new w.Event('change'));}
 await sleep(100);assert(d.querySelectorAll('#viewport-root .render-embedded').length===2,'two independent render instances created');
 assert([...d.querySelectorAll('[id$="render-capture"]')].every(e=>!e.disabled),'all rendering instances respect capabilities');
 $('step').click();await until(()=>w.received.some(m=>m.type==='state'&&m.state.cycle>cycles),'UI frame step');
 const suspended=await w.ui.suspend();assert(suspended.backend==='owned','suspended backend retained');
 const oldReady=w.received.filter(m=>m.type==='ready').length;w.ui.restore({...suspended,backend:'production'});
 await until(()=>$('status').textContent.includes('Current machine retained.'),'failed load retains session');
 assert(w.received.filter(m=>m.type==='ready').length===oldReady&&!$('run').disabled,'old machine retained after mismatched load');
 assert(!$('render-capture').disabled,'capabilities retained after failed load');
 const retained=await w.ui.suspend();assert(retained.backend==='owned','failed replacement retains active backend');
 // Authored drive firmware turns on the spindle and then loops. Both CPUs
 // keep running on one timeline; no private firmware or disk is needed here.
 const rom=new Uint8Array(16384);rom.fill(0xea);rom.set([0xa9,4,0x8d,2,0x1c,0x8d,0,0x1c,0xea,0x4c,8,0xc0]);rom[0x3ffc]=0;rom[0x3ffd]=0xc0;
 const driveFirmware=new File([rom],'authored-1541.rom'),diskFiles=[new File([new Uint8Array(174848)],'authored.d64')];
 const drive=client();drive.send('load',{platform:'c64',backend:'owned',files:diskFiles,driveFirmware});await drive.wait('ready');
 const g64=new Uint8Array(690),gv=new DataView(g64.buffer);g64.set(new TextEncoder().encode('GCR-1541'));g64[9]=84;gv.setUint16(10,4,true);gv.setUint32(12+34*4,684,true);gv.setUint16(684,4,true);g64.set([0,64,128,255],686);
 const rawDrive=client();rawDrive.send('load',{platform:'c64',backend:'owned',files:[new File([g64],'authored.g64')],driveFirmware});await rawDrive.wait('ready');rawDrive.send('step');await until(()=>rawDrive.messages.some(m=>m.type==='state'&&m.state.cycle>0),'G64 frame');rawDrive.send('memory-snapshot');const gm=(await rawDrive.wait('memory-overview')).overview;assert(gm.state.drive.g64&&gm.state.track===34&&gm.state.drive.bitCount===32,'G64 half-track identity and cursor');rawDrive.worker.terminate();
 async function driveCommand(type,data={}){const from=drive.messages.length;drive.send(type,{protocol:1,generation:1,...data});return until(()=>drive.messages.slice(from).find(m=>m.type==='drive-debug-result'||m.type==='drive-debug-snapshot'),type);}
 let ds=(await driveCommand('drive-debug-snapshot')).snapshot;assert(ds.bytes.length===256,'drive snapshot');
 ds=(await driveCommand('drive-debug-normalize',{snapshotId:ds.snapshotId,cycle:ds.cycle,bank:ds.bank})).snapshot;
 let stopped=await driveCommand('drive-debug-until',{snapshotId:ds.snapshotId,cycle:ds.cycle,bank:ds.bank,target:0xc008});assert(stopped.reason==='target','drive breakpoint '+JSON.stringify(stopped));ds=stopped.snapshot;
 const stepped=await driveCommand('drive-debug-step',{snapshotId:ds.snapshotId,cycle:ds.cycle,bank:ds.bank});assert(stepped.reason==='instruction'&&Number(stepped.snapshot.cycle)>Number(ds.cycle),'drive instruction step');assert(stepped.snapshot.drive.motor,'drive motor');ds=stepped.snapshot;
 const from=drive.messages.length;drive.send('drive-debug-until',{protocol:1,generation:1,snapshotId:ds.snapshotId,cycle:ds.cycle,bank:ds.bank,target:0xdead,cycleBudget:10000000});await until(()=>drive.messages.slice(from).some(m=>m.type==='drive-debug-started'),'drive job');drive.send('pause');
 assert((await until(()=>drive.messages.slice(from).find(m=>m.type==='drive-debug-result'),'drive cancellation')).reason==='cancelled','global pause cancels drive');
 drive.send('memory-snapshot');const dm=(await drive.wait('memory-overview')).overview;assert(dm.regions.some(r=>r.id==='drive-ram')&&!dm.regions.some(r=>r.id==='tape'),'disk memory inspection excludes TAP');assert(Number.isFinite(dm.state.track)&&dm.state.drive.bit>0,'live disk cursor');
 drive.send('save');const diskState=new File([(await drive.wait('saved')).bytes],'drive.rrstate'),diskDecoded=await unpackState(diskState);assert(diskDecoded.meta.firmware.length===4,'drive firmware bound');
 const beforeDrive=w.received.filter(m=>m.type==='ready').length;w.ui.restore({files:diskFiles,driveFirmware,firmware:null,compatibility:true,backend:'owned',stateFile:diskState});await until(()=>w.received.filter(m=>m.type==='ready').length>beforeDrive,'disk UI');
 const selects=[...d.querySelectorAll('#viewport-root .viewport-header select')];for(const s of selects){s.value='drive';s.dispatchEvent(new w.Event('change'));}await until(()=>d.querySelectorAll('[data-registers]').length===2,'independent drive panels');
 const panels=[...d.querySelectorAll('[data-registers]')].map(e=>e.parentElement);for(const p of panels)p.querySelector('[data-action="refresh"]').click();await until(()=>panels.every(p=>p.querySelector('[data-code]').textContent.includes('NOP')),'drive disassembly');
 const previousDriveResults=w.received.filter(m=>m.type==='drive-debug-result').length;panels[0].querySelector('[data-action="normalize"]').click();await until(()=>w.received.filter(m=>m.type==='drive-debug-result').length>previousDriveResults,'drive panel normalize');
 const diskSuspend=await w.ui.suspend();assert(diskSuspend.driveFirmware===driveFirmware,'drive firmware retained on suspension');
 const alteredROM=rom.slice();alteredROM[15]^=1;w.ui.restore({...diskSuspend,driveFirmware:new File([alteredROM],'wrong.rom')});await until(()=>$('status').textContent.includes('State firmware does not match'),'wrong drive firmware rejected');const recoveredDisk=await w.ui.suspend();assert(recoveredDisk.driveFirmware===driveFirmware&&recoveredDisk.files[0]===diskFiles[0],'failed load retains active media and firmware');
 selects[0].value='storage';selects[0].dispatchEvent(new w.Event('change'));await until(()=>d.querySelector('.storage-pane .storage-status')?.textContent.includes('C64 D64'),'disk metadata');const storageView=d.querySelector('.storage-pane [aria-label="Storage view"]');storageView.value='sectors';storageView.dispatchEvent(new w.Event('change'));await until(()=>d.querySelector('.storage-pane .disk-legend')?.textContent.includes('live head'),'disk atlas head');
 for(const s of selects){s.value='render';s.dispatchEvent(new w.Event('change'));}
 const beforeProduction=w.received.filter(m=>m.type==='ready').length;w.ui.restore({files,firmware:null,compatibility:true,backend:'production',stateFile:productionFile});await until(()=>w.received.filter(m=>m.type==='ready').length>beforeProduction,'production replacement');assert([...d.querySelectorAll('[id$="render-capture"]')].every(e=>!e.disabled),'rendering re-enabled on successful production load');assert(!d.querySelector('#viewport-root').textContent.includes('unavailable in this development core'),'unsupported message cleared after backend replacement');
 result.textContent='PASS: actual owned worker boot, complete rendering capture/replay/provenance, stepping, recording cancellation, bound save/restore, cross-core/configuration rejection, debugger, independent render/drive panels, drive stepping/breakpoints/global cancellation, live disk atlas and firmware-bound disk restore, system suspension and failed-load recovery.';document.title='PASS — Owned C64 browser acceptance';
}catch(e){result.textContent=e.stack||String(e);document.title='FAIL — Owned C64 browser acceptance';}
finally{for(const w of workers)w.terminate();if(previousLayout===null)localStorage.removeItem(layoutKey);else localStorage.setItem(layoutKey,previousLayout);}
