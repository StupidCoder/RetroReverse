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
 const ready=await c.wait('ready');assert(ready.backend==='owned'&&ready.capabilities.renderCapture===false,'owned capabilities');
 c.send('capture-render');await until(()=>c.messages.some(m=>m.type==='message'&&m.text.includes('not available')),'unsupported capture');
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
 assert(!$('run').disabled&&$('render-capture').disabled,'play enabled, rendering capture disabled');
 for(const selector of [...d.querySelectorAll('#viewport-root .viewport-header select')]){selector.value='render';selector.dispatchEvent(new w.Event('change'));}
 await sleep(100);assert(d.querySelectorAll('#viewport-root .render-embedded').length===2,'two independent render instances created');
 assert([...d.querySelectorAll('[id$="render-capture"]')].every(e=>e.disabled),'all rendering instances respect capabilities');
 $('step').click();await until(()=>w.received.some(m=>m.type==='state'&&m.state.cycle>cycles),'UI frame step');
 const suspended=await w.ui.suspend();assert(suspended.backend==='owned','suspended backend retained');
 const oldReady=w.received.filter(m=>m.type==='ready').length;w.ui.restore({...suspended,backend:'production'});
 await until(()=>$('status').textContent.includes('Current machine retained.'),'failed load retains session');
 assert(w.received.filter(m=>m.type==='ready').length===oldReady&&!$('run').disabled,'old machine retained after mismatched load');
 assert($('render-capture').disabled,'capabilities retained after failed load');
 const retained=await w.ui.suspend();assert(retained.backend==='owned','failed replacement retains active backend');
 w.ui.restore({files,firmware:null,compatibility:true,backend:'production',stateFile:productionFile});await until(()=>w.received.filter(m=>m.type==='ready').length>oldReady,'production replacement');assert([...d.querySelectorAll('[id$="render-capture"]')].every(e=>!e.disabled),'rendering re-enabled on successful production load');assert(!d.querySelector('#viewport-root').textContent.includes('unavailable in this development core'),'unsupported message cleared after backend replacement');
 result.textContent='PASS: actual owned worker boot, unsupported capture, stepping, recording cancellation, bound save/restore, cross-core/configuration rejection, debugger, independent render panels, system suspension and failed-load recovery.';document.title='PASS — Owned C64 browser acceptance';
}catch(e){result.textContent=e.stack||String(e);document.title='FAIL — Owned C64 browser acceptance';}
finally{for(const w of workers)w.terminate();if(previousLayout===null)localStorage.removeItem(layoutKey);else localStorage.setItem(layoutKey,previousLayout);}
