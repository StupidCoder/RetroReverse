const output=document.querySelector('#result'),frame=document.querySelector('#app');
const assert=(ok,message)=>{if(!ok)throw Error(message);};
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function until(fn,label){const start=performance.now();while(!fn()){if(performance.now()-start>15000)throw Error('Timeout: '+label);await sleep(20);}return fn();}
try{
 const bundle=(await (await fetch('/site/emulators/release.json')).json()).id;
 frame.srcdoc=`<!doctype html><link rel="stylesheet" href="/site/emulators/releases/${bundle}/style.css"><body data-platform="c64"><div id="emulator-app"></div><script>window.sent=[];window.received=[];const BaseWorker=Worker;window.Worker=class extends BaseWorker{constructor(...a){super(...a);this.addEventListener('message',e=>received.push(e.data));}postMessage(m,...a){sent.push(m);return super.postMessage(m,...a);}};<\/script><script type="module" src="/site/emulators/releases/${bundle}/app.js"><\/script>`;
 await until(()=>frame.contentDocument?.querySelector('#load'),'app mounted');
 const w=frame.contentWindow,d=frame.contentDocument,$=id=>d.getElementById(id);
 const click=id=>{assert(!$(id).disabled,id+' enabled');$(id).click();};
 const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=1;tape[20]=48;
 const fort=new URL(location.href).searchParams.has('fort');
 const game=fort?await (await fetch('/games/fort-apocalypse-c64/Fort_Apocalypse.tap')).arrayBuffer():tape;
 const files=new DataTransfer();files.items.add(new File([game],fort?'Fort_Apocalypse.tap':'synthetic.tap'));$('files').files=files.files;click('load');
 await until(()=>!$('view-code').disabled,'Code enabled');click('view-code');
 await until(()=>$('code-registers').textContent.includes('Cycle'),'first snapshot');
 const latest=type=>w.received.filter(m=>m.type===type).at(-1);
 const first=latest('debug-snapshot').snapshot;
 if(fort){await until(()=>$('code-functions').querySelectorAll('button').length===5,'exact Fort function list');$('code-functions').querySelectorAll('button')[2].click();await until(()=>latest('debug-snapshot').snapshot.address===0xa000,'function selection');assert(latest('debug-snapshot').snapshot.cycle===first.cycle,'function navigation is read-only');assert($('code-annotation').textContent.includes('reconciled'),'bug annotation retains uncertainty');}
 click('code-follow');await sleep(100);assert(latest('debug-snapshot').snapshot.cycle===first.cycle,'navigation does not execute');
 $('code-address').value='$D000';$('code-go').requestSubmit();await until(()=>latest('debug-snapshot').snapshot.address===0xd000&&$('code-disassembly').textContent.includes('Unavailable'),'safe I/O unavailable');
 click('code-follow');await until(()=>latest('debug-snapshot').snapshot.address!==0xd000,'follow PC');
 if(!$('code-normalize').disabled){click('code-normalize');await until(()=>$('code-note').textContent.startsWith('boundary'),'normalize');}
 click('code-step');await until(()=>$('code-note').textContent.startsWith('instruction')||$('code-note').textContent.startsWith('interrupt-entry'),'instruction step');
 const stepped=latest('debug-result');assert(stepped.retired<=1,'at most one instruction');
 // Real DOM keyboard events must reach the same worker input path from Code.
 $('code-screen').focus();$('code-screen').dispatchEvent(new w.KeyboardEvent('keydown',{key:'a',code:'KeyA',bubbles:true}));$('code-screen').dispatchEvent(new w.KeyboardEvent('keyup',{key:'a',code:'KeyA',bubbles:true}));assert(w.sent.some(m=>m.type==='input'&&m.keys?.length),'Code preview keyboard input');
 click('code-play');await until(()=>latest('state')?.running,'playing');await sleep(550);
 const sample=latest('debug-snapshot').snapshot;assert(BigInt(sample.cycle)>BigInt(stepped.snapshot.cycle),'live samples advance while playing');assert($('code-step').disabled,'instruction step disabled during play');
 click('view-memory');await until(()=>$('memory-live-status').textContent.startsWith('Running'),'Memory live');click('view-code');
 const pauseStart=performance.now();click('code-pause');await until(()=>!latest('state').running&&!$('code-step').disabled||!latest('state').running&&!$('code-normalize').disabled,'pause');const pauseMs=performance.now()-pauseStart;
 await sleep(100);const paused=latest('debug-snapshot').snapshot.cycle;click('view-render');await sleep(100);assert(latest('state').state.cycle.toString()===paused,'Render navigation does not capture');click('view-play');click('view-code');await sleep(100);assert(latest('debug-snapshot').snapshot.cycle===paused,'tab switching preserves paused execution');
 $('code-address').value='$1234';$('code-go').requestSubmit();await until(()=>latest('debug-snapshot').snapshot.address===0x1234,'address selection');click('code-until');await until(()=>!$('code-cancel').disabled&&latest('debug-started'),'run started');
 const cancelStart=performance.now();click('code-cancel');await until(()=>$('code-note').textContent.startsWith('cancelled'),'cancelled');const cancelMs=performance.now()-cancelStart;
 assert(cancelMs<250,'cancel acknowledgement under 250ms');
 click('view-render');click('render-capture');await until(()=>latest('capture'),'explicit Render capture');
 click('view-code');await until(()=>!$('code-normalize').disabled||!$('code-step').disabled,'Code after capture');
 const cleared=w.received.filter(m=>m.type==='capture-cleared').length;
 click($('code-step').disabled?'code-normalize':'code-step');await until(()=>w.received.filter(m=>m.type==='capture-cleared').length>cleared,'debug invalidates old render evidence');
 await until(()=>!$('code-cancel').disabled===false,'post-capture debug completion');
 const result={schema:1,date:new Date().toISOString(),release:bundle,userAgent:navigator.userAgent,fixture:fort?'exact reference Fort tape at reset (not gameplay)':'synthetic tape',checks:['Code load','read-only navigation','safe I/O peek','explicit normalization','instruction step','keyboard input','live samples','Memory/Code/Render tab switching','explicit Render capture and debug invalidation','pause','run-to cancellation'],pauseMs,cancelMs,heap:latest('state').heap};
 output.textContent=JSON.stringify(result,null,2);document.title='PASS — M2 browser acceptance';
}catch(e){output.textContent=(e.stack||String(e))+'\n'+JSON.stringify({sent:frame.contentWindow.sent?.filter(m=>m.type.startsWith('debug-')),received:frame.contentWindow.received?.filter(m=>m.type.startsWith('debug-')).map(m=>({...m,snapshot:m.snapshot?{snapshotId:m.snapshot.snapshotId,cycle:m.snapshot.cycle}:undefined}))},null,2);document.title='FAIL — M2 browser acceptance';}
