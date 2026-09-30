import {probeC64} from './code-m0-probe.mjs';
const out = document.getElementById('result');
const worker = new Worker(new URL('../../../site/emulators/worker.js', import.meta.url), {type:'module'});
const waiting = new Set(); let state, fatal;
function receive(m) {
  if(m.type === 'state') state = m;
  if(m.type === 'error') fatal = Error(m.text);
  for(const w of [...waiting]) {if(fatal){w.reject(fatal);waiting.delete(w);}else if(w.predicate(m)){w.resolve(m);waiting.delete(w);}}
}
worker.onmessage = e => receive(e.data);
worker.onerror = e => receive({type:'error',text:e.message});
function wait(predicate) {return new Promise((resolve,reject)=>{
 const item={predicate,resolve:m=>{clearTimeout(timer);resolve(m);},reject:e=>{clearTimeout(timer);reject(e);}};
 const timer=setTimeout(()=>{waiting.delete(item);reject(Error('Worker timeout'));},15000);
 if(fatal){clearTimeout(timer);reject(fatal);}else waiting.add(item);
});}
const send = (type,data={}) => worker.postMessage({type,session:1,...data});
const sleep = ms => new Promise(r=>setTimeout(r,ms));
const sha = async b => [...new Uint8Array(await crypto.subtle.digest('SHA-256',b))].map(x=>x.toString(16).padStart(2,'0')).join('');
try {
 const tape = new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=1;tape[20]=48;
 const ready = wait(m=>m.type==='ready');send('load',{platform:'c64',files:[new File([tape],'synthetic.tap')]});await ready;
 send('turbo',{value:true});
 const samples=[];
 for(let i=0;i<12;i++) {
   const startCycle=state.state.cycle, began=performance.now();
   const running=wait(m=>m.type==='state'&&m.running);send('run');await running;
   await sleep(150);
   const stopped=wait(m=>m.type==='message'&&m.text==='Paused. Open Memory to inspect physical storage.');
   const requested=performance.now();send('pause');await stopped;
   const ended=performance.now();
   samples.push({wallMs:ended-began,cycles:state.state.cycle-startCycle,pauseAckMs:ended-requested,
     heapBytes:state.heap,maxCoreCallMs:state.maxCall});
 }
 const workerResult={fixture:'C64 ROM reset/BASIC, synthetic one-pulse tape, turbo, no keyboard input',
   samples, limitations:'Acknowledgement latency includes messaging and framebuffer copy, not input-device or paint latency. Unpaced throughput; not Fort gameplay.'};
 worker.terminate();
 const binary=await (await fetch(new URL('../../../site/emulators/cores/c64/core.wasm',import.meta.url))).arrayBuffer();
 const factory=(await import('../../../site/emulators/cores/c64/core.js')).default;
 const core=await factory({wasmBinary:new Uint8Array(binary)});
 const result={schema:1,date:new Date().toISOString(),userAgent:navigator.userAgent,
   crossOriginIsolated,coreSHA256:await sha(binary),worker:workerResult,c64:probeC64(core),
   codeUI:'not implemented; instructionStepAndStatus is only a core-side proxy'};
 out.textContent=JSON.stringify(result,null,2);document.title='PASS — M0 baseline';
} catch(e) {out.textContent=String(e.stack||e);document.title='FAIL — M0 baseline';}
finally {worker.terminate();}
