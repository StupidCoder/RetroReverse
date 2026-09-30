import {resolveDOSModules,resolveDOSLocation,resolveDOSTour,resolveDOSCondition} from './dos-knowledge.js';
import {matches} from './tour-worker.js';
export function createDOSInspector(core,knowledge){
 const k=knowledge?.data;let locked=null;
 const ram=()=>core.HEAPU8.subarray(core._rr_ram(),core._rr_ram()+core._rr_ram_size());
 const raw=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
 function decorate(s){
  s.contextKey=[s.mode,s.cs,s.csBase].join(':');s.modules=resolveDOSModules(k,s,ram());s.functions={};s.buffers=[];
  for(const [id,f] of Object.entries(k?.functions||{})){try{s.functions[id]=resolveDOSLocation(k,f.entry,s,s.modules);}catch{}}
  for(const [id,b] of Object.entries(k?.buffers||{})){
   if(!b.releases.includes(knowledge.releaseId))continue;
   try{if(!matches(resolveDOSCondition(b.when,k,s,s.modules),s,ram()))continue;
    const loc=resolveDOSLocation(k,b.location,s,s.modules),size=b.stride*b.height;
    if(loc.address+size>ram().length)continue;
    const regions=JSON.parse(core.UTF8ToString(core._rr_inspect_regions())).regions,pi=regions.findIndex(r=>r.id==='palette'),p=core._rr_inspect_data(pi);
    s.buffers.push({id,label:b.label,address:loc.address,width:b.width,height:b.height,stride:b.stride,bytes:Array.from(ram().slice(loc.address,loc.address+size)),palette:Array.from(core.HEAPU8.slice(p,p+768)),cycle:s.cycle});
   }catch{/* Unavailable context is not a guessed buffer. */}
  }
  return s;
 }
 function checkModules(context){const s=raw(),bytes=ram();if(s.mode!==context.mode||s.loadSegment!==context.loadSegment)throw Error('DOS execution/relocation context changed');
  for(const [id,was] of Object.entries(context.modules)){if(was.status!=='verified')throw Error('Required module is not loaded');const m=k.modules[id],base=was.base;if(!m.signatures.every(g=>g.bytes.every((b,i)=>bytes[base+g.offset+i]===b))||!m.relocations.every(r=>(bytes[base+r.offset]|bytes[base+r.offset+1]<<8)===((r.relativeSegment+s.loadSegment)&65535)))throw Error('Overlay/signature invalidated: '+id);}
 }
 return {decorate,prepareTarget(m,s){if(!m.functionId)return ()=>{};const f=k?.functions[m.functionId];if(!f)throw Error('Unknown function');const l=resolveDOSLocation(k,f.entry,s,s.modules);if(l.address!==m.target||!l.module)throw Error('Stale function location');const context={mode:s.mode,loadSegment:s.loadSegment,modules:{[l.module]:s.modules[l.module]}};checkModules(context);const check=()=>checkModules(context);check.bindTarget=()=>core._rr_debug_bind(l.segment,l.segment*16);return check;},resolveLocation:(pkg,id,s)=>resolveDOSLocation(pkg,id,s,s.modules),
  resolveTour(t){const s=decorate(raw());locked={mode:s.mode,loadSegment:s.loadSegment,modules:s.modules};return resolveDOSTour(t,k,s,s.modules);},
  validate(){if(locked)checkModules(locked);}
 };
}
