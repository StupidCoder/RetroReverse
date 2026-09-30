// Reuse the actual M2 worker/UI interactions, then exercise shared state panels.
import './code-m2-browser.mjs';
const output=document.querySelector('#result'),frame=document.querySelector('#app');
const assert=(ok,message)=>{if(!ok)throw Error(message);};
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function until(fn,label){const start=performance.now();while(!fn()){if(performance.now()-start>15000)throw Error('Timeout: '+label);await sleep(20);}return fn();}
try{
 assert(document.title.startsWith('PASS'),'M2 regression failed: '+output.textContent);
 assert(new URL(location.href).searchParams.has('fort'),'Use ?fort with the local reference tape');
 const previous=JSON.parse(output.textContent),w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id);
 const latest=()=>w.received.filter(m=>m.type==='debug-snapshot'||m.type==='debug-result').filter(m=>m.snapshot).at(-1).snapshot;
 const panel=d.querySelector('.game-state-panel'),cycle=latest().cycle;
 await until(()=>panel.querySelectorAll('[data-state-id]').length===6,'six Fort watches');
 for(const e of latest().state.entries){assert(e.node.raw.length===1,'one-byte Fort watch');assert(e.node.value===String(e.node.raw[0]),'decoded value matches raw byte');}
 assert(latest().state.cycle===cycle,'state and CPU have same clock');
 const details=()=>panel.querySelector('[data-state-id="enemy-mode"] details');
 details().open=true;await sleep(30);
 const controls=$('code-workspace').querySelector('select[aria-label="Game state panel position"]').parentElement,checkbox=controls.querySelector('input'),position=controls.querySelector('select');
 const codeAddress=$('code-address').value;
 checkbox.checked=false;checkbox.dispatchEvent(new w.Event('change'));assert(panel.parentElement.hidden,'panel can be hidden');
 checkbox.checked=true;checkbox.dispatchEvent(new w.Event('change'));position.value='before';position.dispatchEvent(new w.Event('change'));
 assert(details().open,'expanded raw field survives layout changes');assert($('code-address').value===codeAddress,'code navigation survives layout changes');
 const memoryButton=[...details().querySelectorAll('button')].find(b=>b.textContent==='Inspect bytes in Memory');assert(!memoryButton.disabled,'paused memory link enabled');memoryButton.click();
 await until(()=>$('memory-address').value==='ram:6E','state-to-memory navigation');
 assert($('memory-workspace').contains(panel),'same panel moved to Memory');await sleep(100);assert(details().open,'disclosure survives shared snapshot refresh');
 const related=[...panel.querySelector('[data-state-id="enemy-mode"]').querySelectorAll('button')].find(b=>b.textContent.startsWith('Code:'));related.click();
 await until(()=>$('code-address').value==='$9C52','related-function navigation');assert($('code-workspace').contains(panel),'same panel returned to Code');
 assert(latest().cycle===cycle,'panel operations and links never execute');
 $('code-play').click();await until(()=>$('code-pause').disabled===false,'play');await sleep(300);
 assert([...panel.querySelectorAll('button')].every(b=>b.disabled),'navigation links disabled during execution');
 $('code-pause').click();await until(()=>$('code-play').disabled===false,'pause');
 // Unknown image must clear definitions, including the moved shared panel.
 $('view-play').click();const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=1;tape[20]=48;const dt=new DataTransfer();dt.items.add(new File([tape],'unknown.tap'));$('files').files=dt.files;$('load').click();await until(()=>!$('view-code').disabled,'unknown image ready');$('view-code').click();
 await until(()=>panel.textContent.includes('No matching live state definitions'),'unknown image generic fallback');assert(panel.querySelectorAll('[data-state-id]').length===0,'no leaked old state');
 output.textContent=JSON.stringify({schema:1,date:new Date().toISOString(),release:previous.release,userAgent:navigator.userAgent,result:'PASS',checks:['M2 regression','six Fort watches','raw/decoded/CPU clock agreement','raw disclosure persistence','panel hide/reorder','shared Code/Memory panel identity','state-to-memory link','related-function link','no navigation execution','running links disabled','unknown-image reset'],limitations:'Reference Fort tape identification and reset/ROM execution only, not gameplay-phase verification. Elite is not enabled.'},null,2);document.title='PASS — M3 browser acceptance';
}catch(e){output.textContent=e.stack||String(e);document.title='FAIL — M3 browser acceptance';}
