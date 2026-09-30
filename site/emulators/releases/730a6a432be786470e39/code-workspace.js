import {createTourPanel} from './tour-panel.js';
import {disassemble6502,hex} from './disassembly6502.js';
export function createCodeWorkspace({root,views,send,transport,platform,statePanel,panelLayout}){
 if(!['c64','dos'].includes(platform))return null;
 root.innerHTML=`<div class="memory-heading"><div><h2>Code</h2><p id="code-note" role="status">Load a game to inspect its code.</p></div></div>
 <div class="memory-player"><canvas id="code-screen" tabindex="0" width="392" height="272" aria-label="Code game output"></canvas><div><div class="memory-transport"><button id="code-play">Play</button><button id="code-pause">Pause</button><button id="code-frame">Next frame</button></div><p>Click the preview for keyboard controls. Navigation never advances execution.</p><button id="code-tape-play">Play tape</button><button id="code-tape-stop">Stop tape</button></div></div>
 <div class="code-toolbar"><button id="code-normalize">Finish partial instruction</button><button id="code-step">Step instruction</button><button id="code-follow">Follow PC</button><button id="code-refresh">Refresh</button><button id="code-cancel" disabled>Cancel run</button></div>
 <p id="code-registers"></p><div class="code-layout"><section><form id="code-go"><label>CPU address <input id="code-address" placeholder="$0800" size="8" spellcheck="false"></label><button>Inspect</button></form>
 <div class="code-toolbar"><button id="code-until">Run to address</button><label>At current PC <select id="code-resume"><option value="stop-if-current">Stop immediately</option><option value="next-match">Run to next visit</option></select></label></div>
 <p>Address stops use the current CPU mapping. They do not prove that a documented function is loaded. Runs stop after 10 seconds or 9,852,480 cycles.</p>
 <pre id="code-disassembly" tabindex="0" aria-label="Mapped 6510 disassembly"></pre></section><aside><h3>Identified functions</h3><p id="code-identity"></p><div id="code-functions"></div><p id="code-annotation"></p></aside></div>`;
 if(platform==='dos'){root.querySelector('#code-tape-play').hidden=true;root.querySelector('#code-tape-stop').hidden=true;root.querySelector('#code-normalize').hidden=true;root.querySelector('#code-disassembly').setAttribute('aria-label','Live x86 disassembly');root.querySelector('#code-until').parentElement.nextElementSibling.textContent='Linear-address stops use the current CPU mode. Near branch targets are shown in linear display coordinates. REP remains one instruction; oversized atomic operations are rejected.';}
 const bufferPanel=document.createElement('section');bufferPanel.className='code-buffers';root.append(bufferPanel);
 const tourPanel=createTourPanel({root,send,inspect:at=>{address=at;refresh();},layout:show=>panelLayout?.suggest(show)});
 const $=id=>root.querySelector('#code-'+id);
 let enabled=false,active=false,running=false,busy=false,externalBusy=false,snapshot=null,address=-1,pending=0,snapshotPending=false,job=0,generation=0,timer=null,knowledge=null,functionId=null;
 const command=(type,args={})=>send(type,{protocol:1,generation,...args});
 function refresh(force=false){clearTimeout(timer);timer=null;if(!enabled||(!active&&!force))return;snapshotPending=true;pending=command('debug-snapshot',{address});controls();}
 function schedule(){if(!timer&&active&&enabled&&running)timer=setTimeout(refresh,200);}
 function controls(){
  const locked=!enabled||running||busy||externalBusy||snapshotPending;
  for(const id of ['play','frame','until','go','resume']){const el=$(id);if(el.tagName==='FORM'){for(const e of el.elements)e.disabled=locked;}else el.disabled=locked;}
  $('step').disabled=locked||!snapshot?.boundary;$('normalize').disabled=locked||!snapshot||snapshot.boundary;
  $('tape-play').disabled=$('tape-stop').disabled=!enabled||busy||externalBusy;
  $('pause').disabled=!running&&!busy;$('cancel').disabled=!busy;
  for(const b of $('functions').querySelectorAll('button'))b.disabled=locked;
 }
 function render(s){
  snapshot=s;statePanel?.update(s);const regs=Object.entries(s.registers).map(([k,v])=>`${k}=$${hex(v,s.architecture==='x86'?8:2)}`).join('  ');
  $('registers').textContent=`${regs}  Bank=$${hex(s.bank,2)}  Cycle ${s.cycle} · ${s.boundary?'Next PC $'+hex(s.nextPC):'Mid-instruction; next PC unavailable'}${s.interruptPending?' · interrupt pending':''}${s.stalled?' · RDY stall':''}`;
  if(s.architecture==='x86')$('registers').textContent=`${s.mode} · CS:IP $${hex(s.cs)}:$${hex(s.ip,s.mode==='real16'?4:8)} · CS base $${hex(s.csBase,8)} · step ${s.cycle} · ${regs} · `+s.segments.map((v,i)=>`${['ES','CS','SS','DS','FS','GS'][i]}=$${hex(v.selector)} (base $${hex(v.base,8)})`).join(' ');
  $('disassembly').replaceChildren();
  for(const row of s.instructions||disassemble6502(s.bytes,s.address)){
   const line=document.createElement('div');line.className=row.address===s.nextPC?'code-current':'';
   line.textContent=`${row.address===s.nextPC?'▶':' '} $${hex(row.address)}  ${(row.bytes||s.bytes.slice(row.address-s.address,row.address-s.address+row.length)).map(b=>b==null?'??':hex(b,2)).join(' ').padEnd(8)}  ${row.text}  ${({0:'[RAM]',1:'[color RAM / open bus]',2:'[I/O unavailable]',3:'[ROM]'})[s.architecture==='x86'?(row.bytes?.some(b=>b==null)?2:0):s.mapping[(row.address-s.address)&65535]]}`;
   $('disassembly').append(line);
  }
  if(document.activeElement!==$('address'))$('address').value='$'+hex(s.address);
  if(s.architecture==='x86'){functions();bufferPanel.replaceChildren();for(const b of s.buffers||[]){const h=document.createElement('h3');h.textContent=`${b.label} · step ${b.cycle} · RAM $${hex(b.address,8)}`;const canvas=document.createElement('canvas');canvas.width=b.width;canvas.height=b.height;canvas.setAttribute('aria-label',b.label);const ctx=canvas.getContext('2d'),im=ctx.createImageData(b.width,b.height);for(let y=0;y<b.height;y++)for(let x=0;x<b.width;x++){const i=(y*b.width+x)*4,p=b.bytes[y*b.stride+x]*3;for(let c=0;c<3;c++)im.data[i+c]=(b.palette[p+c]||0)<<2;im.data[i+3]=255;}ctx.putImageData(im,0,0);bufferPanel.append(h,canvas);}if(!s.buffers?.length)bufferPanel.textContent='No verified render-buffer context at this instruction.';}
  controls();schedule();
 }
 function functions(){
  $('functions').replaceChildren();const k=knowledge?.data;
  $('identity').textContent=k?`${k.game.name} · exact image match. Entries are documented; runtime applicability is unverified.`:'No exact knowledge package match. Raw disassembly and address stops are available.';
  if(!k)return;if(platform==='dos')$('identity').textContent=k.game.name+' · required file hashes matched. Function navigation requires verified live modules.';
  for(const [id,f] of Object.entries(k.functions||{})){
   if(!f.releases.includes(knowledge.releaseId))continue;const loc=k.locations[f.entry];const resolved=platform==='dos'?snapshot?.functions?.[id]:loc?.kind==='cpu'?{address:Number(loc.address)}:null;if(!resolved){if(platform==='dos'){const p=document.createElement('p');p.textContent=f.label+' — module unavailable or ambiguous';$('functions').append(p);}continue;}
   const b=document.createElement('button');b.textContent=`$${hex(resolved.address,platform==='dos'?8:4)} · ${f.label}`;
   b.onclick=()=>{if(running||busy)return;address=resolved.address;functionId=id;const notes=Object.values(k.annotations||{}).filter(a=>a.function===id).map(a=>a.text);const evidence=f.evidence.map(e=>k.evidence[e]).map(e=>`${e.status}: ${e.description} ${e.limitations||''}`);$('annotation').textContent=[f.applicability,...evidence,...notes].join(' ');refresh();};$('functions').append(b);
  }
  controls();
 }
 function execute(type){
  if(!snapshot||running||busy||externalBusy||snapshotPending)return;
  busy=true;clearTimeout(timer);controls();
  job=command(type,{functionId,snapshotId:snapshot.snapshotId,cycle:snapshot.cycle,bank:snapshot.bank,target:address<0?snapshot.address:address,resume:$('resume').value});
 }
 $('go').onsubmit=e=>{e.preventDefault();const text=$('address').value.trim().replace(/^\$|^0x/i,'');if(!(platform==='dos'?/^[0-9a-f]{1,8}$/i:/^[0-9a-f]{1,4}$/i).test(text)){$('note').textContent=platform==='dos'?'Enter a linear RAM address in hexadecimal.':'Enter a CPU address from $0000 to $FFFF.';return;}functionId=null;address=parseInt(text,16);refresh();};
 $('follow').onclick=()=>{functionId=null;address=-1;$('annotation').textContent='';refresh();};$('refresh').onclick=refresh;
 $('step').onclick=()=>execute('debug-step');$('normalize').onclick=()=>execute('debug-normalize');$('until').onclick=()=>execute('debug-until');
 $('cancel').onclick=()=>command('debug-cancel');
 for(const [id,type] of [['play','run'],['pause','pause'],['frame','step']])$(id).onclick=()=>transport(type);
 $('tape-play').onclick=()=>send('tape',{down:1});$('tape-stop').onclick=()=>send('tape',{down:0});
 views.register({id:'code',label:'Code',panel:root,enabled:false});
 return {
  ready(session){enabled=true;generation=session;$('note').textContent='Mapped CPU bytes · snapshots do not advance execution.';views.enable('code',true);command('debug-capabilities');controls();},
  inspect(at,id=null){functionId=id;address=at;refresh();},
  refreshState(){refresh(true);},
  reset(){tourPanel.reset();statePanel?.reset();enabled=false;busy=false;snapshotPending=false;job=0;snapshot=null;functionId=null;address=-1;knowledge=null;bufferPanel.replaceChildren();clearTimeout(timer);views.enable('code',false);controls();},
  setActive(value){active=value;clearTimeout(timer);timer=null;if(value)refresh();},
  present(canvas){if(active){const c=$('screen');if(c.width!==canvas.width)c.width=canvas.width;if(c.height!==canvas.height)c.height=canvas.height;c.getContext('2d').drawImage(canvas,0,0);}},
  state(m){tourPanel.state(m);const was=running;running=m.running;statePanel?.setRunning(m.running||m.debugBusy||m.capturing||m.memoryRecording||m.saving);externalBusy=m.capturing||m.saving||m.memoryRecording||(m.debugBusy&&!busy);controls();if(active&&was&&!running)refresh();schedule();},
  tourResult(m){functionId=null;tourPanel.result(m);if(m.snapshot){functionId=null;address=-1;render(m.snapshot);}},
  result(m){
   if(m.generation!==generation)return;
   if(m.type==='debug-capabilities'){knowledge=m.knowledge;tourPanel.ready(knowledge,generation);statePanel?.setKnowledge(knowledge);functions();return;}
   if(m.type==='debug-snapshot'&&m.request===pending){snapshotPending=false;render(m.snapshot);return;}
   if(m.type==='debug-started'&&m.request===job){$('note').textContent='Running a bounded debugger job…';return;}
   if(m.type==='debug-result'){
    if(m.request===job){job=0;busy=false;$('note').textContent=`${m.reason}${m.text?': '+m.text:''}${m.cycles!==undefined?' · '+m.cycles+(platform==='dos'?' instruction steps':' cycles'):''}`;if(m.snapshot){functionId=null;address=-1;render(m.snapshot);}else refresh();controls();}
    else if(m.request===pending){snapshotPending=false;controls();$('note').textContent=m.text||m.reason;schedule();}
   }
  }
 };
}
