// Generic guarded ARM32 knowledge. Game addresses and signatures stay in packages.
export function create3DOInspector(core,knowledge){
 const k=knowledge?.data;
 const valid=()=>{if(!k?.liveGuards?.length)return false;const p=core._rr_ram(),b=core.HEAPU8;return k.liveGuards.every(g=>g.bytes.every((v,i)=>b[p+g.address+i]===v));};
 const validate=()=>{if(!valid())throw Error('The documented 3DO code/car layout is not present.');};
 return {decorate(s){s.contextKey=s.mode+':'+s.task;s.contextValid=valid();return s;},validate,
  resolveTour(t){validate();return t;},
  resolveLocation(pkg,id,s){if(!s.contextValid)throw Error('Live layout guards do not match');const l=pkg.locations[id];if(l?.kind==='physical'&&pkg.spaces[l.space]?.region==='ram')return {kind:'ram',address:Number(l.offset)};if(l?.kind==='cpu')return {kind:'cpu',address:Number(l.address)};throw Error('Unsupported 3DO live location');},
  prepareTarget(m,s){if(!m.functionId)return ()=>{};validate();const f=k.functions[m.functionId],l=k.locations[f?.entry];if(l?.kind!=='cpu'||Number(l.address)!==m.target)throw Error('Stale function');const check=()=>validate();check.bindTarget=()=>core._rr_debug_bind(s.task);return check;},
  eventAddress:e=>e.region===0?e.offset:e.region===1?e.offset+0x200000:null,
  activityScope:'Selected DRAM/VRAM bus writes and endpoint differences; direct HLE bulk stores are not write-traced'
 };
}
