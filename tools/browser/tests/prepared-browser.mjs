// Private local media acceptance; no uploaded image or checked-in state.
const frame=document.querySelector('#app'),output=document.querySelector('#result'),checks=[],sleep=ms=>new Promise(r=>setTimeout(r,ms));
const assert=(v,t)=>{if(!v)throw Error(t);};async function until(fn,label){const start=performance.now();while(!fn()){if(performance.now()-start>150000)throw Error('Timeout: '+label);await sleep(25);}return fn();}
let report;
try{
 const elite=new URLSearchParams(location.search).get('game')==='elite';
 const release=(await(await fetch('/site/emulators/release.json')).json()).id;
 frame.srcdoc=`<!doctype html><link rel="stylesheet" href="/site/emulators/releases/${release}/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>localStorage.removeItem('rr.viewport.v1.c64');window.received=[];const Base=Worker;window.Worker=class extends Base{constructor(...a){super(...a);this.addEventListener('message',e=>received.push(e.data));}};<\/script><script type="module" src="/site/emulators/releases/${release}/app.js"><\/script>`;
 await until(()=>frame.contentDocument?.querySelector('#load'),'mount');const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id),latest=t=>w.received.filter(m=>m.type===t).at(-1),b=a=>d.querySelector('[data-prepared="'+a+'"]');
 const files=new DataTransfer();files.items.add(new File([await(await fetch(elite?'/games/elite-c64/Elite.tap':'/games/fort-apocalypse-c64/Fort_Apocalypse.tap')).arrayBuffer()],elite?'Elite.tap':'Fort_Apocalypse.tap'));$('files').files=files.files;$('files').dispatchEvent(new w.Event('change'));$('load').click();
 await until(()=>latest('state')&&b('start')&&!b('start').disabled,'prepared lessons');const initial=latest('state').state.cycle;
 $('workspace-nav').value='code';$('workspace-nav').dispatchEvent(new w.Event('change'));
 b('clear').click();await until(()=>d.querySelector('.prepared-status').textContent.includes('cleared'),'clear cache');
 const choose=id=>{const select=d.querySelector('[aria-label="Prepared lesson"]');select.value=id;select.dispatchEvent(new w.Event('change'));};choose(elite?'loader-prefix':'terrain');b('start').click();await until(()=>!b('cancel').disabled,'preparing');b('cancel').click();await until(()=>!b('start').disabled,'cancelled');assert(latest('state').state.cycle===initial,'cancel preserves session');checks.push('cancel preserves initial session');
 if(elite){
  b('start').click();await until(()=>latest('tour-state')?.phase==='paused-at-stop','loader entry');
  for(const kind of ['code','lesson','atlas','tape'])assert(d.querySelector('#viewport-root [data-kind="'+kind+'"]'),'loader panel '+kind);
  for(let i=1;i<4;i++){const button=d.querySelector('[data-action="continue"]');await until(()=>!button.disabled,'continue enabled');button.click();await until(()=>latest('tour-state')?.index===i,'loader stop '+i);}
  assert(latest('tour-state').phase==='completed','loader complete');checks.push('Elite loader prefix completes with Code, lesson, RAM and pulses visible');
 }else{
 b('start').click();await until(()=>latest('tour-state')?.phase==='paused-at-stop','terrain first stop');assert(latest('tour-state').stop.id,'authentic tour stop');checks.push('cold preparation starts Fort terrain tour');
 const before=w.received.length;choose('terrain');b('start').click();await until(()=>w.received.slice(before).some(m=>m.type==='tour-state'&&m.phase==='paused-at-stop'),'cached first stop');assert(w.received.slice(before).some(m=>m.text==='Loaded verified local lesson checkpoint.'),'cache restored');checks.push('repeat start uses verified local cache');
 choose('enemy-ai');b('start').click();await until(()=>latest('experiment-state')?.phase==='prepared','AI setup');assert(latest('experiment-state').snapshot.nextPC===0x9cda,'native AI entry');checks.push('cold prepared experiment applies reviewed setup');
 }
 report={result:'PASS',release,checks};
}catch(e){report={result:'FAIL',checks,error:e.stack||String(e),messages:frame.contentWindow?.received?.filter(m=>['error','message','tour-state','experiment-state'].includes(m.type)).slice(-4)};}
output.textContent=JSON.stringify(report,null,2);document.title=report.result+' — Prepared lessons';const id=new URLSearchParams(location.search).get('report');if(id)await fetch('/__viewport_result__?id='+encodeURIComponent(id),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(report)});
