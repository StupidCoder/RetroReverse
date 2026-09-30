import {hex} from './disassembly6502.js';
export function createStatePanel({memory,code}){
 const root=document.createElement('section');root.className='game-state-panel';root.setAttribute('aria-label','Game state');
 let snapshot=null,knowledge=null,running=false;const opened=new Set();
 const element=(tag,text)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;};
 function value(n){if(n.status==='unavailable')return 'unavailable';if(n.kind==='array')return `${n.children.length} slots`;if(n.kind==='struct')return 'record';return `${n.value??''}${n.enumLabel&&n.kind==='enum'?' · '+n.enumLabel:''}${n.status==='unknown'?' · unknown enum':''}`;}
 function draw(){
  for(const d of root.querySelectorAll('details[data-state-key]')){if(d.open)opened.add(d.dataset.stateKey);else opened.delete(d.dataset.stateKey);}
  const scroll=root.scrollTop;root.replaceChildren(element('h3','Game state'));
  root.append(element('p',snapshot?`Snapshot ${snapshot.snapshotId} · cycle ${snapshot.cycle} · ${snapshot.boundary?'instruction boundary':'mid-instruction sample'}. Raw bytes and values come from this same snapshot.`:'No state snapshot yet.'));
  root.append(element('p','Documented meanings are not a live gameplay-phase assertion. Slot incarnations count observed empty/unavailable → occupied transitions; reuse between samples may be missed.'));
  if(!snapshot?.state.entries.length){root.append(element('p','No matching live state definitions for this image.'));return;}
  function node(n,entry,key,label){
   const details=element('details'),summary=element('summary',`${label}: ${value(n)}${n.occupancy?' · '+n.occupancy:''}${n.occupancy==='occupied'?' · observed incarnation '+n.incarnation:''}`);details.dataset.stateKey=key;details.open=opened.has(key);details.append(summary);
   details.ontoggle=()=>{if(details.open)opened.add(key);else opened.delete(key);};
   const raw=element('pre',(n.raw||[]).slice(0,256).map(b=>b==null?'??':hex(b,2)).join(' '));raw.className='state-raw';details.append(raw);
   if(n.raw?.length>256)details.append(element('p','First 256 bytes shown; inspect Memory for the full range.'));
   const address=entry.location.address+n.offset,map=entry.mapping?.[n.offset],region=map===0?'ram':map===3?(address>=0xe000?'kernal':address>=0xd000?'chars':'basic'):null;
   if(region){const b=element('button','Inspect bytes in Memory');b.disabled=running;b.onclick=()=>memory(region,address-(region==='kernal'?0xe000:region==='chars'?0xd000:region==='basic'?0xa000:0));details.append(b);}
   if(entry.location.kind==='cpu'){const b=element('button','Inspect CPU address in Code');b.disabled=running;b.onclick=()=>code(address);details.append(b);}
   // Bound DOM work for large arrays: descendants mount when disclosed.
   if(n.children){let built=false;const build=()=>{if(!details.open||built)return;built=true;for(const c of n.children)details.append(node(c,entry,key+'/'+c.path,c.label||c.path));};build();details.addEventListener('toggle',build);}
   return details;
  }
  for(const entry of snapshot.state.entries){
   const section=element('section');section.dataset.stateId=entry.id;
   if(entry.error)section.append(element('h4',entry.label),element('p',entry.error));
   else section.append(node(entry.node,entry,entry.id,entry.label));
   section.append(element('p',entry.applicability));
   section.append(element('p',entry.evidence.map(e=>`${e.status}: ${e.description}${e.limitations?' '+e.limitations:''}`).join(' ')));
   for(const id of entry.relatedFunctions){const f=knowledge?.data?.functions[id],l=knowledge?.data?.locations[f?.entry];if(l?.kind!=='cpu')continue;const b=element('button','Code: '+f.label);b.disabled=running;b.onclick=()=>code(Number(l.address));section.append(b);}
   root.append(section);
  }
  root.scrollTop=scroll;
 }
 return {root,setKnowledge(k){knowledge=k;},update(s){snapshot=s;draw();},setRunning(value){if(running===value)return;running=value;draw();},reset(){snapshot=null;knowledge=null;opened.clear();root.replaceChildren();draw();}};
}
