// No commercial game media. Exercise the packaged app with a synthetic call loop.
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
 const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;put(tape);assert(core._rr_tape(21),'tape');
 const ram=new Uint8Array(65536).fill(0xea);ram.set([0x20,0,9,0x4c,0,8],0x800);ram.set([0xee,0,2,0x60],0x900);ram[0x200]=0;put(ram);assert(core._rr_prepare(0x800,0),'prepare');
 const n=core._rr_state_save(),q=new InputQueue(50),state=await packState({format:1,platform:'c64',media:[{name:'synthetic.tap',size:tape.length,sha256:await digest(tape)}],firmware:await Promise.all(roms.map(digest)),core:await digest(wasm),configuration:{compatibility:true,customFirmware:false},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n));
 frame.srcdoc=`<!doctype html><html lang="en"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="${base}/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>localStorage.removeItem('rr.viewport.v1.c64');localStorage.removeItem('rr.viewport.v1.gb');window.received=[];window.errors=[];addEventListener('error',e=>errors.push(e.message+' '+e.filename+':'+e.lineno));const BaseWorker=Worker;window.Worker=class extends BaseWorker{constructor(...a){super(...a);this.addEventListener('message',e=>{received.push(e.data);if(received.length>1000)received.shift();});}};<\/script><script type="module" src="${base}/app.js"><\/script>`;
 await until(()=>frame.contentDocument?.getElementById('load'),'app mount');
 const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id),latest=t=>w.received.filter(m=>m.type===t).at(-1);
 function file(id,bytes,name){const dt=new w.DataTransfer();dt.items.add(new w.File([bytes],name));$(id).files=dt.files;$(id).dispatchEvent(new w.Event('change'));}
 file('files',tape,'synthetic.tap');file('statefile',state,'synthetic.rrstate');

 await until(()=>!$('view-code').disabled,'ready');$('workspace-nav').value='code';$('workspace-nav').dispatchEvent(new w.Event('change'));
 await until(()=>d.querySelectorAll('#viewport-root .viewport').length===4,'four panes');
 const panes=()=>[...d.querySelectorAll('#viewport-root .viewport')],kind=k=>panes().find(p=>p.dataset.kind===k),part=(k,suffix)=>kind(k)?.querySelector('[id$="-'+suffix+'"]');
 await until(()=>part('code','step')&&!part('code','step').disabled,'Code ready');
 const initial=latest('state').state.cycle;assert(initial===0,'layout does not execute');
 part('code','over').click();await until(()=>latest('debug-result')?.reason==='return','step over');assert(latest('debug-result').snapshot.nextPC===0x803,'return PC');
 checks.push('four-pane Code with native execution');
 // Replace State with another independently addressed Code inspector.
 let choose=kind('state').querySelector('.viewport-header select');choose.value='code';choose.dispatchEvent(new w.Event('change'));
 await until(()=>panes().filter(p=>p.dataset.kind==='code').length===2,'two code panes');await sleep(250);
 const codes=panes().filter(p=>p.dataset.kind==='code');for(let i=0;i<2;i++){codes[i].querySelector('[id$="-address"]').value=i?'0900':'0800';codes[i].querySelector('[id$="-go"]').requestSubmit();}
 await until(()=>codes[0].querySelector('[id$="-disassembly"]').textContent.includes('$0800')&&codes[1].querySelector('[id$="-disassembly"]').textContent.includes('$0900'),'independent code addresses');
 const cycle=latest('state').state.cycle;$('workspace-nav').value='loader';$('workspace-nav').dispatchEvent(new w.Event('change'));await until(()=>kind('atlas')&&kind('storage'),'loader panes');await until(()=>kind('storage').querySelector('.memory-selection').textContent.includes('Pulses'),'pulse stream');
 assert(latest('state').state.cycle===cycle,'loader layout does not execute');checks.push('independent Code instances and Game/Session/Storage/atlas Loader preset');
 const bounds=k=>kind(k).getBoundingClientRect();assert(bounds('game').left<bounds('session').left&&bounds('storage').top>bounds('game').top&&bounds('atlas').top>bounds('session').top,'Loader quadrant positions');
 // Split / merge and keyboard resizing change only presentation.
 let pane=kind('session');[...pane.querySelectorAll('.viewport-menu button')].find(b=>b.textContent.startsWith('Close')).click();pane=kind('atlas');pane.querySelector('details').open=true;[...pane.querySelectorAll('.viewport-menu button')].find(b=>b.textContent==='Split side by side').click();assert(panes().length===4,'split');
 const bar=d.querySelector('#viewport-root [role=separator]'),before=bar.getAttribute('aria-valuenow');bar.dispatchEvent(new w.KeyboardEvent('keydown',{key:'ArrowRight',bubbles:true}));assert(bar.getAttribute('aria-valuenow')!==before,'keyboard resize');
 pane=kind('hex');pane.querySelector('details').open=true;[...pane.querySelectorAll('.viewport-menu button')].find(b=>b.textContent.startsWith('Close')).click();assert(panes().length===3,'merge');
 checks.push('split, merge and accessible resize');
 $('workspace-nav').value='memory';$('workspace-nav').dispatchEvent(new w.Event('change'));await until(()=>kind('hex')?.querySelector('pre').textContent,'memory bytes');
 const memorySelect=kind('game').querySelector('.viewport-header select');memorySelect.value='hex';memorySelect.dispatchEvent(new w.Event('change'));await sleep(250);
 const hexes=panes().filter(p=>p.dataset.kind==='hex');for(let i=0;i<2;i++){hexes[i].querySelector('input').value=i?'200':'100';hexes[i].querySelector('form').requestSubmit();}
 await until(()=>hexes[0].querySelector('pre').textContent.startsWith('$00000100')&&hexes[1].querySelector('pre').textContent.startsWith('$00000200'),'independent memory offsets');checks.push('independent memory ranges');
 $('run').click();await until(()=>latest('state')?.running,'run in Memory');await sleep(250);assert(kind('atlas'),'play retains layout');$('pause').click();await until(()=>!latest('state')?.running,'pause');checks.push('global transport retains Memory layout');
 $('workspace-nav').value='render';$('workspace-nav').dispatchEvent(new w.Event('change'));await until(()=>!$('render-capture').disabled,'render');$('render-capture').click();await until(()=>latest('capture'),'capture');checks.push('Rendering preset capture and source/detail panes');
 let renderSelect=kind('render-details').querySelector('.viewport-header select');renderSelect.value='render';renderSelect.dispatchEvent(new w.Event('change'));
 await until(()=>panes().filter(p=>p.dataset.kind==='render').length===2,'independent render instances');await sleep(300);
 const renderCopies=panes().filter(p=>p.dataset.kind==='render'),sliders=renderCopies.map(p=>p.querySelector('input[type=range]'));assert(sliders.every(Boolean),'separate render timelines');
 const otherPosition=sliders[0].value;sliders[1].value=sliders[1].min;sliders[1].dispatchEvent(new w.Event('input'));await sleep(100);assert(sliders[0].value===otherPosition,'render timeline selection isolated');checks.push('independent Rendering instances');
 $('workspace-nav').value='play';$('workspace-nav').dispatchEvent(new w.Event('change'));await until(()=>d.querySelector('#viewport-root .viewport-game canvas'),'Play viewport');const screen=d.querySelector('#viewport-root .viewport-game canvas');screen.focus();assert(d.activeElement===screen,'game focus');
 const beforeSwitch=latest('state').state.cycle;$('system-select').value='gb';$('system-select').dispatchEvent(new w.Event('change'));await until(()=>$('system-select')?.value==='gb'&&d.body.dataset.platform==='gb','switch system');
 $('system-select').value='c64';$('system-select').dispatchEvent(new w.Event('change'));await until(()=>d.body.dataset.platform==='c64'&&!$('view-code').disabled,'restore C64');assert(latest('state').state.cycle===beforeSwitch,'suspended cycle restored');checks.push('system switch restores exact paused cycle');
 const systems=[...$('system-select').options].map(o=>o.value);for(const system of systems.filter(s=>s!=='c64')){$('system-select').value=system;$('system-select').dispatchEvent(new w.Event('change'));await until(()=>d.body.dataset.platform===system,'shell '+system);await sleep(30);assert(d.querySelector('#viewport-root .viewport'),'viewport shell '+system);}
 $('system-select').value='c64';$('system-select').dispatchEvent(new w.Event('change'));await until(()=>d.body.dataset.platform==='c64'&&!$('view-code').disabled,'restore after shell sweep');checks.push('all sixteen system shells mount');
 assert(new Set([...d.querySelectorAll('[id]')].map(e=>e.id)).size===d.querySelectorAll('[id]').length,'unique DOM IDs');
 frame.style.height='780px';frame.style.width='1280px';await sleep(100);assert(d.documentElement.scrollHeight<=780,'workspace fits laptop height');
 const storageFiles=new DataTransfer();storageFiles.items.add(new File([new Uint8Array(174848)],'inspection.d64'));$('files').files=storageFiles.files;const storageCycle=latest('state').state.cycle;$('inspect-storage').click();await until(()=>kind('storage')?.querySelector('.storage-status').textContent.includes('C64 D64'),'D64 inspection without boot');assert(latest('state').state.cycle===storageCycle,'storage inspection does not execute');assert($('game-name').textContent==='synthetic.tap','inspection keeps loaded game identity');checks.push('D64 inspection without replacing or executing the game');
 const otherTape=new DataTransfer();otherTape.items.add(new File([tape],'other.tap'));$('files').files=otherTape.files;$('inspect-storage').click();await until(()=>kind('storage')?.querySelector('.storage-status').textContent.includes('other.tap'),'separate selected tape');assert(kind('storage').querySelector('[aria-label="Storage view"] option[value="pulses"]').disabled,'different tape cannot borrow live head');checks.push('live tape stream stays bound to the loaded image');
 frame.style.width='375px';await until(()=>d.querySelector('.mobile-pane-picker'),'mobile pane selector');assert(panes().length===1,'one readable mobile viewport');assert(d.documentElement.scrollWidth<=375,'no horizontal overflow');checks.push('375px single-pane adaptation without changing desktop layout');
 report={result:'PASS',release,userAgent:navigator.userAgent,checks};
}catch(e){report={result:'FAIL',userAgent:navigator.userAgent,checks,error:e.stack||String(e),errors:frame.contentWindow?.errors,messages:frame.contentWindow?.received?.filter(m=>['error','message'].includes(m.type)).slice(-6)};}
output.textContent=JSON.stringify(report,null,2);document.title=report.result+' — Viewport acceptance';

if(new URLSearchParams(location.search).has('report'))await fetch('/__viewport_result__?id='+encodeURIComponent(new URLSearchParams(location.search).get('report')),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(report)});
