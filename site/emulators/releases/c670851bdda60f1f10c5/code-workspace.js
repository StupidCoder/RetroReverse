import {createExperimentPanel} from './experiment-panel.js';
import {createTourPanel} from './tour-panel.js';
import {disassemble6502,hex} from './disassembly6502.js';
export function createCodeWorkspace({root,views,send,transport,platform,statePanel,panelLayout,idPrefix="code",embedded=false,lessons=true,onInspect,onTourLayout}){
 if(!['c64','dos','3do'].includes(platform))return null;
 root.innerHTML=`<div class="memory-heading"><div><h2>Code</h2><p id="code-note" role="status">Load a game to inspect its code.</p></div></div>
 <div class="memory-player"><canvas id="code-screen" tabindex="0" width="392" height="272" aria-label="Code game output"></canvas><div><div class="memory-transport"><button id="code-play">Play</button><button id="code-pause">Pause</button><button id="code-frame">Next frame</button></div><p>Click the preview for keyboard controls. Navigation never advances execution.</p><button id="code-tape-play">Play tape</button><button id="code-tape-stop">Stop tape</button></div></div>
 <div class="code-toolbar"><button id="code-normalize">Finish partial instruction</button><button id="code-step">Step instruction</button><button id="code-over">Step over</button><button id="code-out">Step out</button><button id="code-follow">Follow PC</button><button id="code-refresh">Refresh</button><button id="code-cancel" disabled>Cancel run</button></div>
 <form id="code-watch"><h3>Conditional write watchpoint</h3><label>RAM address <input name="address" value="$0200" size="6"></label> <label>Value <input name="value" value="$00" size="4"></label> <label>Mask <input name="mask" value="$00" size="4"></label> <label>Writer PC (optional) <input name="writer" size="6"></label> <button>Run until write</button><p>Hexadecimal. Stop when (written byte &amp; mask) equals value; mask $00 matches any write, including a same-value write. Stops may be mid-instruction. Step out requires a balanced JSR/RTS frame; it cannot infer arbitrary stack use.</p></form><p id="code-event" role="status"></p>
 <p id="code-registers"></p><div class="code-layout"><section><form id="code-go"><label>CPU address <input id="code-address" placeholder="$0800" size="8" spellcheck="false"></label><button>Inspect</button></form>
 <div class="code-toolbar"><button id="code-until">Run to address</button><label>At current PC <select id="code-resume"><option value="stop-if-current">Stop immediately</option><option value="next-match">Run to next visit</option></select></label></div>
 <p>Address stops use the current CPU mapping. They do not prove that a documented function is loaded. Runs stop after 10 seconds or 9,852,480 cycles.</p>
 <pre id="code-disassembly" tabindex="0" aria-label="Mapped 6510 disassembly"></pre></section><aside><h3>Identified functions</h3><p id="code-identity"></p><div id="code-functions"></div><p id="code-annotation"></p></aside></div>`;
 if(platform!=='c64'){for(const id of ['over','out','watch'])root.querySelector('#'+idPrefix+'-'+id).hidden=true;root.querySelector('#code-tape-play').hidden=true;root.querySelector('#code-tape-stop').hidden=true;root.querySelector('#code-normalize').hidden=true;root.querySelector('#code-disassembly').setAttribute('aria-label','Live x86 disassembly');root.querySelector('#code-until').parentElement.nextElementSibling.textContent='Linear-address stops use the current CPU mode. Near branch targets are shown in linear display coordinates. REP remains one instruction; oversized atomic operations are rejected.';}
 if(platform==='3do'){root.querySelector('#code-disassembly').setAttribute('aria-label','Live ARM60 disassembly');root.querySelector('#code-until').parentElement.nextElementSibling.textContent='ARM32 big-endian. Task switches and Portfolio HLE services are separate boundaries; movies cannot be instruction-stepped. Named stops bind the current task.';}
 if(idPrefix!=='code')for(const el of root.querySelectorAll('[id]'))el.id=el.id.replace(/^code-/,idPrefix+'-');
 if(embedded){root.classList.add('code-embedded');root.querySelector('.memory-player').hidden=true;const aside=root.querySelector('aside'),fold=document.createElement('details');fold.className='code-functions-fold';const summary=document.createElement('summary');summary.textContent='Identified functions';aside.before(fold);fold.append(summary,aside);const watch=root.querySelector('[id$="-watch"]'),advanced=document.createElement('details'),title=document.createElement('summary');advanced.className='code-watch-fold';title.textContent='Write watchpoint';watch.before(advanced);advanced.append(title,watch);}
 const bufferPanel=document.createElement('section');bufferPanel.className='code-buffers';root.append(bufferPanel);
 const experimentPanel=lessons&&platform==='c64'?createExperimentPanel({root,send}):null;
 const tourPanel=lessons?createTourPanel({root,send,inspect:at=>{if(onInspect)onInspect(at);else{address=at;refresh();}},layout:(show,preset)=>preset&&onTourLayout?onTourLayout(preset):panelLayout?.suggest(show)}):{ready(){},reset(){},state(){},result(){}};
 const $=id=>root.querySelector('#'+idPrefix+'-'+id);
 let enabled=false,active=false,running=false,busy=false,externalBusy=false,experimentOwned=false,snapshot=null,address=-1,pending=0,snapshotPending=false,waitingSnapshot=false,job=0,generation=0,timer=null,knowledge=null,functionId=null;
 const command=(type,args={})=>send(type,{protocol:1,generation,client:idPrefix,...args});
 function refresh(force=false){clearTimeout(timer);timer=null;if(!enabled||(!active&&!force))return;snapshotPending=true;pending=command('debug-snapshot',{address});controls();}
 function schedule(){if(!timer&&active&&enabled&&running)timer=setTimeout(refresh,200);}
 function controls(){
  const locked=!enabled||running||busy||externalBusy||snapshotPending;
  for(const id of ['play','frame','until','go','resume']){const el=$(id);if(el.tagName==='FORM'){for(const e of el.elements)e.disabled=locked;}else el.disabled=locked||experimentOwned;}
  for(const id of ['over','out'])$(id).disabled=locked||experimentOwned||!snapshot?.boundary||snapshot.interruptPending;
  for(const e of $('watch').elements)e.disabled=locked||experimentOwned;
  $('step').disabled=locked||experimentOwned||!snapshot?.boundary;$('normalize').disabled=locked||experimentOwned||!snapshot||snapshot.boundary;
  $('tape-play').disabled=$('tape-stop').disabled=!enabled||busy||externalBusy||experimentOwned;
  $('pause').disabled=!running&&!busy;$('cancel').disabled=!busy;if(embedded){$('cancel').hidden=!busy;$('normalize').hidden=!!snapshot?.boundary;}
  for(const b of $('functions').querySelectorAll('button'))b.disabled=locked;
 }
 function render(s){
  snapshot=s;statePanel?.update(s);const regs=Object.entries(s.registers).map(([k,v])=>`${k}=$${hex(v,platform!=='c64'?8:2)}`).join('  ');
  $('registers').textContent=`${regs}  Bank=$${hex(s.bank,2)}  Cycle ${s.cycle} · ${s.boundary?'Next PC $'+hex(s.nextPC):'Mid-instruction; next PC unavailable'}${s.interruptPending?' · interrupt pending':''}${s.stalled?' · RDY stall':''}`;
  if(s.architecture==='x86')$('registers').textContent=`${s.mode} · CS:IP $${hex(s.cs)}:$${hex(s.ip,s.mode==='real16'?4:8)} · CS base $${hex(s.csBase,8)} · step ${s.cycle} · ${regs} · `+s.segments.map((v,i)=>`${['ES','CS','SS','DS','FS','GS'][i]}=$${hex(v.selector)} (base $${hex(v.base,8)})`).join(' ');
  if(s.architecture==='arm60')$('registers').textContent=`ARM32 · task ${s.task} · scheduler step ${s.cycle} · retired ${s.retiredInstructions} · display ${s.frame} · field ${s.field} · ${s.hle?'Portfolio HLE boundary · ':''}${s.movieHLE?'Movie HLE active · ':''}${regs}`;
  $('disassembly').replaceChildren();
  for(const row of s.instructions||disassemble6502(s.bytes,s.address)){
   const line=document.createElement('div');line.className=row.address===s.nextPC?'code-current':'';
   line.textContent=`${row.address===s.nextPC?'▶':' '} $${hex(row.address)}  ${(row.bytes||s.bytes.slice(row.address-s.address,row.address-s.address+row.length)).map(b=>b==null?'??':hex(b,2)).join(' ').padEnd(8)}  ${row.text}  ${({0:'[RAM]',1:'[color RAM / open bus]',2:'[I/O unavailable]',3:'[ROM]'})[platform!=='c64'?(row.bytes?.some(b=>b==null)?2:0):s.mapping[(row.address-s.address)&65535]]}`;
   $('disassembly').append(line);
  }
  if(document.activeElement!==$('address'))$('address').value='$'+hex(s.address);
  if(s.architecture==='x86'){functions();bufferPanel.replaceChildren();for(const b of s.buffers||[]){const h=document.createElement('h3');h.textContent=`${b.label} · step ${b.cycle} · RAM $${hex(b.address,8)}`;const canvas=document.createElement('canvas');canvas.width=b.width;canvas.height=b.height;canvas.setAttribute('aria-label',b.label);const ctx=canvas.getContext('2d'),im=ctx.createImageData(b.width,b.height);for(let y=0;y<b.height;y++)for(let x=0;x<b.width;x++){const i=(y*b.width+x)*4,p=b.bytes[y*b.stride+x]*3;for(let c=0;c<3;c++)im.data[i+c]=(b.palette[p+c]||0)<<2;im.data[i+3]=255;}ctx.putImageData(im,0,0);bufferPanel.append(h,canvas);}if(!s.buffers?.length)bufferPanel.textContent='No verified render-buffer context at this instruction.';}
  if(platform==='3do')functions();
  controls();schedule();
 }
 function functions(){
  $('functions').replaceChildren();const k=knowledge?.data;
  $('identity').textContent=k?`${k.game.name} · exact image match. Entries are documented; runtime applicability is unverified.`:'No exact knowledge package match. Raw disassembly and address stops are available.';
  if(!k)return;if(platform==='3do')$('identity').textContent=k.game.name+(snapshot?.contextValid?' · live code and car layout guards matched.':' · documented layout unavailable.');
  if(platform==='dos')$('identity').textContent=k.game.name+' · required file hashes matched. Function navigation requires verified live modules.';
  for(const [id,f] of Object.entries(k.functions||{})){
   if(!f.releases.includes(knowledge.releaseId))continue;if(platform==='3do'&&!snapshot?.contextValid)continue;const loc=k.locations[f.entry];const resolved=platform==='dos'?snapshot?.functions?.[id]:loc?.kind==='cpu'?{address:Number(loc.address)}:null;if(!resolved){if(platform==='dos'){const p=document.createElement('p');p.textContent=f.label+' — module unavailable or ambiguous';$('functions').append(p);}continue;}
   const b=document.createElement('button');b.textContent=`$${hex(resolved.address,platform==='dos'?8:4)} · ${f.label}`;
   b.onclick=()=>{if(running||busy)return;address=resolved.address;functionId=id;const notes=Object.values(k.annotations||{}).filter(a=>a.function===id).map(a=>a.text);const evidence=f.evidence.map(e=>k.evidence[e]).map(e=>`${e.status}: ${e.description} ${e.limitations||''}`);$('annotation').textContent=[f.applicability,...evidence,...notes].join(' ');refresh();};$('functions').append(b);
  }
  controls();
 }
 function execute(type,args={}){
  if(!snapshot||running||busy||externalBusy||experimentOwned||snapshotPending)return;
  busy=true;clearTimeout(timer);controls();
  job=command(type,{functionId,snapshotId:snapshot.snapshotId,cycle:snapshot.cycle,bank:snapshot.bank,target:address<0?snapshot.address:address,resume:$('resume').value,...args});
 }
 $('go').onsubmit=e=>{e.preventDefault();const text=$('address').value.trim().replace(/^\$|^0x/i,'');if(!(platform!=='c64'?/^[0-9a-f]{1,8}$/i:/^[0-9a-f]{1,4}$/i).test(text)){$('note').textContent=platform!=='c64'?'Enter a linear RAM address in hexadecimal.':'Enter a CPU address from $0000 to $FFFF.';return;}functionId=null;address=parseInt(text,16);refresh();};
 $('follow').onclick=()=>{functionId=null;address=-1;$('annotation').textContent='';refresh();};$('refresh').onclick=refresh;
 $('step').onclick=()=>execute('debug-step');$('normalize').onclick=()=>execute('debug-normalize');$('until').onclick=()=>execute('debug-until');
 $('over').onclick=()=>execute('debug-over');$('out').onclick=()=>execute('debug-out');
 $('watch').onsubmit=e=>{e.preventDefault();const values={};for(const key of ['address','value','mask','writer']){const raw=$('watch').elements.namedItem(key).value.trim().replace(/^\$|^0x/i,'');if(key==='writer'&&!raw){values[key]=-1;continue;}if(!/^[0-9a-f]{1,4}$/i.test(raw)){$('note').textContent='Enter hexadecimal watchpoint values.';return;}values[key]=parseInt(raw,16);}execute('debug-watch',{watch:values});};
 $('cancel').onclick=()=>command('debug-cancel');
 for(const [id,type] of [['play','run'],['pause','pause'],['frame','step']])$(id).onclick=()=>transport(type);
 $('tape-play').onclick=()=>send('tape',{down:1});$('tape-stop').onclick=()=>send('tape',{down:0});
 views.register({id:idPrefix,label:'Code',panel:root,enabled:false});
 return {
  dispose(){clearTimeout(timer);active=false;enabled=false;},
  ready(session){enabled=true;generation=session;$('note').textContent='Mapped CPU bytes · snapshots do not advance execution.';views.enable(idPrefix,true);command('debug-capabilities');controls();},
  inspect(at,id=null){functionId=id;address=at;refresh();},
  refreshState(){refresh(true);},
  reset(){$('event').replaceChildren();experimentPanel?.reset();tourPanel.reset();statePanel?.reset();enabled=false;busy=false;snapshotPending=false;job=0;snapshot=null;functionId=null;address=-1;knowledge=null;bufferPanel.replaceChildren();clearTimeout(timer);views.enable(idPrefix,false);controls();},
  setActive(value){active=value;clearTimeout(timer);timer=null;if(value)refresh();},
  present(canvas){if(active){const c=$('screen');if(c.width!==canvas.width)c.width=canvas.width;if(c.height!==canvas.height)c.height=canvas.height;c.getContext('2d').drawImage(canvas,0,0);}},
  state(m){experimentPanel?.state(m);tourPanel.state(m);const was=running,wasBlocked=externalBusy;running=m.running;statePanel?.setRunning(m.running||m.debugBusy||m.capturing||m.memoryRecording||m.saving);experimentOwned=!!m.experimentOwned;externalBusy=m.capturing||m.saving||m.memoryRecording||(m.debugBusy&&!busy);controls();if(active&&((was&&!running)||(wasBlocked&&!externalBusy)))refresh();schedule();},
  experimentResult(m){experimentPanel?.result(m);functionId=null;if(m.snapshot){address=-1;render(m.snapshot);}$('note').textContent='Experiment: '+m.phase;},
  tourResult(m){functionId=null;tourPanel.result(m);$('note').textContent='Tour: '+m.phase;if(m.snapshot){functionId=null;address=-1;render(m.snapshot);}},
  result(m){
   if(m.generation!==generation)return;
   if(m.type==='debug-capabilities'){knowledge=m.knowledge;tourPanel.ready(knowledge,generation);experimentPanel?.ready(knowledge,generation);statePanel?.setKnowledge(knowledge);functions();return;}
   if(m.type==='debug-snapshot'&&m.request===pending){snapshotPending=false;if(waitingSnapshot){waitingSnapshot=false;$('note').textContent='Mapped CPU bytes · snapshots do not advance execution.';}render(m.snapshot);return;}
   if(m.type==='debug-started'&&m.request===job){$('note').textContent='Running a bounded debugger job…';return;}
   if(m.type==='debug-result'){
    if(m.request===job){job=0;busy=false;$('event').replaceChildren();if(m.event){const e=m.event;$('event').append(`Write $${hex(e.value,2)} to $${hex(e.address)} at cycle ${e.cycle}. `);const b=document.createElement('button');b.textContent='Inspect writer $'+hex(e.pc);b.onclick=()=>{functionId=null;address=e.pc;refresh();};$('event').append(b,' Shows current code bytes at the recorded writer address.');}$('note').textContent=`${m.reason}${m.text?': '+m.text:''}${m.cycles!==undefined?' · '+m.cycles+(platform==='dos'?' instruction steps':platform==='3do'?' scheduler steps':' cycles'):''}`;if(m.snapshot){functionId=null;address=-1;render(m.snapshot);}else refresh();controls();}
    else if(m.request===pending){snapshotPending=false;controls();$('note').textContent=m.text||m.reason;if(m.retryable&&active&&enabled){waitingSnapshot=true;clearTimeout(timer);timer=setTimeout(refresh,200);}else schedule();}
   }
  }
 };
}
