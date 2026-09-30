// No commercial game media. Exercise the packaged app with a synthetic call loop.
import {packState,digest} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const output=document.querySelector('#result'),frame=document.querySelector('#app'),checks=[];
const assert=(v,t)=>{if(!v)throw Error(t);},sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function until(fn,label){const start=performance.now();while(!fn()){if(performance.now()-start>30000)throw Error('Timeout: '+label);await sleep(25);}return fn();}
let report;
try{
 const release=(await (await fetch('/site/emulators/release.json')).json()).id,base='/site/emulators/releases/'+release;
 const wasm=new Uint8Array(await (await fetch(base+'/cores/c64/core.wasm')).arrayBuffer()),factory=(await import(base+'/cores/c64/core.js')).default,core=await factory({wasmBinary:wasm});
 const roms=await Promise.all(['basic','kernal','chargen'].map(async n=>new Uint8Array(await (await fetch(base+'/firmware/c64/'+n+'.rom')).arrayBuffer())));
 const put=b=>core.HEAPU8.set(b,core._rr_input()),firmware=new Uint8Array(20480);let offset=0;for(const r of roms){firmware.set(r,offset);offset+=r.length;}put(firmware);assert(core._rr_init(8192,8192,4096),'init');
 const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;put(tape);assert(core._rr_tape(21),'tape');
 const ram=new Uint8Array(65536).fill(0xea);ram.set([0x20,0,9,0x4c,0,8],0x800);ram.set([0xee,0,2,0x60],0x900);ram[0x200]=0;put(ram);assert(core._rr_prepare(0x800,0),'prepare');
 const n=core._rr_state_save(),q=new InputQueue(50),state=await packState({format:1,platform:'c64',media:[{name:'synthetic.tap',size:tape.length,sha256:await digest(tape)}],firmware:await Promise.all(roms.map(digest)),core:await digest(wasm),configuration:{compatibility:true,customFirmware:false},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n));
 frame.srcdoc=`<!doctype html><html lang="en"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="${base}/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>window.received=[];const BaseWorker=Worker;window.Worker=class extends BaseWorker{constructor(...a){super(...a);this.addEventListener('message',e=>{received.push(e.data);if(received.length>1000)received.shift();});}};<\/script><script type="module" src="${base}/app.js"><\/script>`;
 await until(()=>frame.contentDocument?.getElementById('load'),'app mount');
 const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id),latest=t=>w.received.filter(m=>m.type===t).at(-1);
 function file(id,bytes,name){const dt=new w.DataTransfer();dt.items.add(new w.File([bytes],name));$(id).files=dt.files;$(id).dispatchEvent(new w.Event('change'));}
 file('files',tape,'synthetic.tap');file('statefile',state,'synthetic.rrstate');
 await until(()=>!$('view-code').disabled,'Code ready');$('view-code').click();await until(()=>!$('code-over').disabled,'snapshot');
 assert(latest('debug-snapshot').snapshot.nextPC===0x800,'prepared PC');
 async function debug(id){const before=latest('debug-result')?.request;$(id).click();return until(()=>latest('debug-result')?.request!==before&&latest('debug-result'),'debug '+id);}
 let result=await debug('code-over');assert(result.reason==='return'&&result.snapshot.nextPC===0x803,'over returns');
 await until(()=>!$('code-step').disabled,'step enabled');await debug('code-step');await until(()=>!$('code-step').disabled,'call step enabled');await debug('code-step');
 await until(()=>!$('code-out').disabled,'out enabled');result=await debug('code-out');assert(result.reason==='return'&&result.snapshot.nextPC===0x803,'out returns');checks.push('nested native stepping through UI');
 await until(()=>!$('code-watch').elements[0].disabled,'watch enabled');const before=latest('debug-result').request;$('code-watch').requestSubmit();result=await until(()=>latest('debug-result')?.request!==before&&latest('debug-result'),'watch result');assert(result.reason==='write-watchpoint'&&result.event.pc===0x900,'actual writer');
 await until(()=>$('code-event').querySelector('button'),'writer link');const cycle=result.snapshot.cycle;$('code-event').querySelector('button').click();await until(()=>latest('debug-snapshot')?.snapshot.address===0x900,'writer code');assert(latest('debug-snapshot').snapshot.cycle===cycle,'navigation does not run');
 await until(()=>!$('code-normalize').disabled,'normalize');await debug('code-normalize');checks.push('write watchpoint, historical writer navigation, partial-instruction normalization');
 $('view-play').click();await until(()=>!$('run').disabled,'play enabled');$('run').click();await until(()=>latest('state')?.running,'running');await sleep(200);$('pause').click();await until(()=>!latest('state')?.running,'paused');checks.push('ordinary Play run/pause');
 $('view-memory').click();await until(()=>$('memory-bytes').children.length,'memory snapshot');$('memory-address').value='$0200';$('memory-go').requestSubmit();await until(()=>$('memory-address').value.includes('0200')||$('memory-address').value.includes(':200'),'memory navigation');
 $('memory-duration').value='0.1';await until(()=>!$('memory-record').disabled,'record enabled');$('memory-record').click();await until(()=>!$('memory-history').hidden&&!$('memory-record').disabled,'record complete');$('memory-address').value='$0200';$('memory-go').requestSubmit();await until(()=>$('memory-detail').querySelector('button'),'recorded writer link');$('memory-detail').querySelector('button').click();await until(()=>!$('code-workspace').hidden,'writer selects Code');checks.push('Memory atlas, bounded recording, writer links');
 $('view-render').click();await until(()=>!$('render-capture').disabled,'render ready');$('render-capture').click();await until(()=>latest('capture'),'display capture');checks.push('ordinary Render capture');
 const preReturn=latest('debug-snapshot')?.request;$('view-code').click();await until(()=>latest('debug-snapshot')?.request!==preReturn,'fresh Code snapshot after Render');checks.push('fresh Code snapshot after Render');await until(()=>!$('code-step').disabled||!$('code-normalize').disabled,'Code resumes');frame.style.width='375px';await sleep(100);
 for(const el of $('code-watch').querySelectorAll('input'))assert(el.labels.length,'watch input has label');
 $('code-follow').focus();assert(d.activeElement===$('code-follow'),'keyboard focus');assert($('code-disassembly').getAttribute('aria-label'),'disassembly name');
 assert(d.documentElement.scrollWidth<=375,'narrow layout overflow: '+JSON.stringify([...d.querySelectorAll('body *')].filter(e=>e.getBoundingClientRect().right>375&&e.getBoundingClientRect().width).map(e=>({tag:e.tagName,id:e.id,width:e.getBoundingClientRect().width,right:e.getBoundingClientRect().right})).slice(0,15)));checks.push('375px layout, form labels, focus and named disassembly');
 report={result:'PASS',release,userAgent:navigator.userAgent,checks,heapBytes:latest('state').heap};
}catch(e){report={result:'FAIL',userAgent:navigator.userAgent,checks,error:e.stack||String(e),messages:frame.contentWindow?.received?.filter(m=>['error','message'].includes(m.type)).slice(-6)};}
output.textContent=JSON.stringify(report,null,2);document.title=report.result+' — M9 browser acceptance';
// Optional test-only collector; production application has no upload endpoint.
if(new URLSearchParams(location.search).has('report'))await fetch('/__m9_result__?browser='+encodeURIComponent(new URLSearchParams(location.search).get('report')),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(report)});
