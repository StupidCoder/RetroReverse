// Bounded live module resolution. An address gets a name only while its module
// identity holds; code data and patched immediates are always decoded live.
export function resolveDOSModules(k,s,ram){
 const out={};
 for(const [id,m] of Object.entries(k?.modules||{})){
  if(s.mode!==m.mode){out[id]={status:'wrong-mode'};continue;}
  const valid=base=>base>=0&&base+m.size<=Math.min(ram.length,0xa0000)&&m.signatures.every(g=>g.bytes.every((b,i)=>ram[base+g.offset+i]===b))&&m.relocations.every(r=>(ram[base+r.offset]|ram[base+r.offset+1]<<8)===((r.relativeSegment+s.loadSegment)&65535));
  const bases=[];
  if(m.kind==='mz'){const base=(s.loadSegment+m.paragraph)*16;if(valid(base))bases.push(base);}
  else for(let base=0;base+m.size<=Math.min(ram.length,0xa0000);base+=16){if(valid(base)){bases.push(base);if(bases.length>1)break;}}
  out[id]=bases.length===1?{status:'verified',base:bases[0],segment:bases[0]/16,size:m.size}: {status:bases.length?'ambiguous':'unavailable'};
 }
 return out;
}
export function resolveDOSLocation(k,id,s,modules,depth=0){
 if(depth>=16)throw Error('Location depth limit');const l=k.locations[id];if(!l)throw Error('Unknown location');
 if(l.kind==='relative'){const a=resolveDOSLocation(k,l.base,s,modules,depth+1);return {...a,address:a.address+Number(l.offset)};}
 if(l.kind==='relocated-segment'){if(s.mode!=='real16'||modules[l.module]?.status!=='verified')throw Error('Relocation context unavailable');const r=k.modules[l.module].relocations.find(r=>r.offset===l.relocation);if(!r)throw Error('Unknown segment relocation');const segment=(s.loadSegment+r.relativeSegment)&65535;return {kind:'ram',space:l.module,module:l.module,segment,address:segment*16+Number(l.offset)};}
 if(l.kind==='module'){const m=modules[l.module];if(m?.status!=='verified')throw Error('Module '+l.module+' is '+(m?.status||'unavailable'));return {kind:'ram',space:l.module,address:m.base+Number(l.offset),module:l.module,segment:m.segment};}
 if(l.kind==='physical'&&k.spaces[l.space]?.region==='ram')return {kind:'ram',space:l.space,address:Number(l.offset)};
 throw Error('DOS requires a verified module or explicit physical RAM location');
}
export function resolveDOSCondition(c,k,s,modules){
 if(c.location){const l=resolveDOSLocation(k,c.location,s,modules);return {all:[{pc:l.address},{register:'CS',value:l.segment}]};}
 if(c.all)return {all:c.all.map(x=>resolveDOSCondition(x,k,s,modules))};
 if(c.any)return {any:c.any.map(x=>resolveDOSCondition(x,k,s,modules))};
 return c;
}
export function resolveDOSTour(t,k,s,modules){
 const address=id=>resolveDOSLocation(k,id,s,modules).address;
 const condition=c=>resolveDOSCondition(c,k,s,modules);
 const range=r=>r.location?{...r,address:address(r.location)}:r;
 const resolved={...t,start:condition(t.start),guards:t.guards.map(range),stops:t.stops.map(x=>({...x,until:condition(x.until),assertions:x.assertions.map(condition),capture:x.capture.map(range),layout:{...x.layout,address:x.until.location?address(x.until.location):x.layout.address}}))};
 for(const stop of resolved.stops)for(let i=0;i<stop.capture.length;i++)for(const b of stop.capture.slice(0,i)){const a=stop.capture[i];if(a.address<b.address+b.length&&b.address<a.address+a.length)throw Error('Resolved capture ranges overlap');}
 return resolved;
}
export async function identifyDOSFiles(entries,packages,hash,executable){
 const paths=entries.map(e=>e.path.normalize('NFC'));if(new Set(paths).size!==paths.length)return {status:'ambiguous',labels:[]};
 const matches=[];
 for(const p of packages.filter(p=>p.platform==='dos'))for(const r of p.releases){
  if(r.executable!==executable||!r.fileSet||r.fileSet.pathPolicy!=='ascii-insensitive')continue;
  let match=true;for(const member of r.fileSet.members){const e=entries.find(e=>e.path.normalize('NFC')===member.path);if(!e||e.file.size!==member.size||await hash(e.file)!==member.sha256){match=false;break;}}
  if(match)matches.push({status:'matched',packageId:p.id,releaseId:r.id,revision:p.revision,sourceSHA256:p.sourceSHA256,data:p.knowledge,labels:r.labels});
 }
 return matches.length===1?matches[0]:{status:matches.length?'ambiguous':'unknown',labels:[]};
}
