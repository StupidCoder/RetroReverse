// Local acceptance test: requires the privately supplied Fort Apocalypse TAP; no media is bundled.
import {packState,digest} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const output=document.querySelector('#result'),frame=document.querySelector('#app'),checks=[];
const assert=(v,t)=>{if(!v)throw Error(t);},sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function until(fn,label){const start=performance.now();while(!fn()){if(frame.contentWindow?.errors?.length)throw Error(frame.contentWindow.errors.join('\n'));if(performance.now()-start>30000)throw Error('Timeout: '+label);await sleep(25);}return fn();}
let report;
try{
 const release=(await (await fetch('/site/emulators/release.json')).json()).id,base='/site/emulators/releases/'+release;
 const wasm=new Uint8Array(await (await fetch(base+'/cores/c64/core.wasm')).arrayBuffer()),factory=(await import(base+'/cores/c64/core.js')).default,core=await factory({wasmBinary:wasm});
 const roms=await Promise.all(['basic','kernal','chargen'].map(async n=>new Uint8Array(await (await fetch(base+'/firmware/c64/'+n+'.rom')).arrayBuffer())));
 const put=b=>core.HEAPU8.set(b,core._rr_input()),firmware=new Uint8Array(20480);let offset=0;for(const r of roms){firmware.set(r,offset);offset+=r.length;}put(firmware);assert(core._rr_init(8192,8192,4096),'init');
 const tape=new Uint8Array(await (await fetch('/games/fort-apocalypse-c64/Fort_Apocalypse.tap')).arrayBuffer());put(tape);assert(core._rr_tape(tape.length),'tape');
 const recipe=await (await fetch('./fort-tour-boot.json')).json();
 for(const action of recipe){if(action.run)core._rr_run(action.run);if(action.key)core._rr_key(...action.key);if(action.play!==undefined){core._rr_play(action.play);break;}}
 core._rr_run(300000);
 const n=core._rr_state_save(),q=new InputQueue(50),state=await packState({format:1,platform:'c64',media:[{name:'Fort_Apocalypse.tap',size:tape.length,sha256:await digest(tape)}],firmware:await Promise.all(roms.map(digest)),core:await digest(wasm),configuration:{compatibility:true,customFirmware:false},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n));
 frame.srcdoc=`<!doctype html><html lang="en"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="${base}/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>localStorage.removeItem('rr.viewport.v1.c64');localStorage.removeItem('rr.viewport.v1.gb');window.received=[];window.errors=[];addEventListener('error',e=>errors.push(e.message+' '+e.filename+':'+e.lineno));const BaseWorker=Worker;window.Worker=class extends BaseWorker{constructor(...a){super(...a);this.addEventListener('message',e=>{received.push(e.data);if(received.length>1000)received.shift();});}};<\/script><script type="module" src="${base}/app.js"><\/script>`;
 await until(()=>frame.contentDocument?.getElementById('load'),'app mount');
 const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id),latest=t=>w.received.filter(m=>m.type===t).at(-1);
 function file(id,bytes,name){const dt=new w.DataTransfer();dt.items.add(new w.File([bytes],name));$(id).files=dt.files;$(id).dispatchEvent(new w.Event('change'));}
 file('files',tape,'Fort_Apocalypse.tap');file('statefile',state,'loading.rrstate');

 await until(()=>!$('view-code').disabled,'ready');$('workspace-nav').value='loader';$('workspace-nav').dispatchEvent(new w.Event('change'));
 const pane=()=>d.querySelector('[data-kind="storage"]'),input=()=>pane().querySelector('[aria-label="Memory offset"]'),progress=()=>pane().querySelector('.tape-position'),follow=()=>pane().querySelector('[data-follow]');
 await until(()=>progress()&&!progress().hidden,'initial tape snapshot');
 const initial=progress().value;w.received.length=0;$('run').click();
 await until(()=>progress().value>initial+1000&&w.received.filter(m=>m.type==='memory-overview').length>=4,'live tape samples during playback');
 assert(Number(input().value)>0,'follow window advances');
 checks.push('real Fort tape: multiple live snapshots, advancing head, scrolling followed window');
 follow().checked=false;follow().dispatchEvent(new w.Event('change'));input().value='0';input().closest('form').requestSubmit();
 const head=progress().value;await until(()=>progress().value>head+500,'head advances with follow off');assert(input().value==='0','manual window stays fixed');
 $('pause').click();await until(()=>!latest('state').running,'pause');await sleep(600);
 const paused=progress().value;await sleep(500);assert(progress().value===paused,'paused head stable');
 follow().checked=true;follow().dispatchEvent(new w.Event('change'));assert(Number(input().value)===Math.max(0,paused-32),'follow immediately recenters while paused');
 checks.push('manual navigation retains position; pause freezes head; Follow immediately recenters');
 assert(pane().querySelector('.pane-status').textContent.includes('%'),'visible pulse progress');
 const atlas=d.querySelector('[data-kind="atlas"] .pane-status');assert(atlas.textContent.includes('Live memory'),'shared atlas also samples');
 checks.push('tape progress and shared live atlas');
 report={result:'PASS',release,userAgent:navigator.userAgent,checks};
}catch(e){report={result:'FAIL',userAgent:navigator.userAgent,checks,error:e.stack||String(e),errors:frame.contentWindow?.errors,messages:frame.contentWindow?.received?.filter(m=>['error','message'].includes(m.type)).slice(-6)};}
output.textContent=JSON.stringify(report,null,2);document.title=report.result+' — Live tape acceptance';

if(new URLSearchParams(location.search).has('report'))await fetch('/__viewport_result__?id='+encodeURIComponent(new URLSearchParams(location.search).get('report')),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(report)});
