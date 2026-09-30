// Optional local-image integration: node memory-wasm.mjs <platform> <image>
import assert from 'node:assert/strict';
import fs from 'node:fs';
import {loadCore} from './wasm-harness.mjs';
import {createMemoryService} from '../../../site/emulators/memory-worker.js';
const [platform,path]=process.argv.slice(2),core=await loadCore(platform,path);
const file={name:path.split('/').at(-1),size:fs.statSync(path).size,arrayBuffer:async()=>{const b=fs.readFileSync(path);return b.buffer.slice(b.byteOffset,b.byteOffset+b.byteLength);}};
const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
const proof=()=>core._rr_proof?core.UTF8ToString(core._rr_proof()):JSON.stringify(status());
const service=createMemoryService({core,platform,files:[file],status});
const before=proof(),s=await service.snapshot();assert(s.regions.length);assert.equal(proof(),before);
for(const r of s.regions){const p=service.page(r.id,0);assert(p.bytes.length<=256&&p.bytes.length>0);}
assert.equal(proof(),before,'peeks changed the core');
if(platform==='gg'){assert.equal(s.regions.filter(r=>r.kind==='rom').length,16);assert(s.labels.some(l=>l.name==='Green Hills Act 1 compressed map'&&l.start===0x3430&&l.length===1926));}
if(!s.activity){
 const desc=JSON.parse(core.UTF8ToString(core._rr_inspect_regions()));
 for(const r of desc.regions){assert(r.size>0);const at=core._rr_inspect_data(r.index);assert(at>0&&at+r.size<=core.HEAPU8.length,r.id+' bounds');const p=service.page(r.id,r.size-1);assert.deepEqual(p.bytes,core.HEAPU8.slice(at+p.offset,at+p.offset+p.bytes.length));}
 const detailed=service.overview(1);assert(detailed.regions.every(r=>r.bitmap.length<=1048576),'large maps must fit canvas limits');
 await service.liveSnapshot();for(let i=0;i<3;i++)assert(core._rr_run(1000,0,0)>=0);const endpoint=proof();await service.liveSnapshot();assert.equal(proof(),endpoint,'live inspection changed execution');service.stopLive();
 const raw=JSON.parse(core.UTF8ToString(core._rr_inspect_regions()));for(const r of raw.regions){const p=service.page(r.id,0),at=core._rr_inspect_data(r.index);assert.deepEqual(p.bytes,core.HEAPU8.slice(at,at+p.bytes.length));}
}
if(s.activity){
 await service.begin();for(let i=0;i<8;i++)assert(core._rr_run(1000,0,0)>=0);
 const end=service.finish();assert(end.count>0);const control=await loadCore(platform,path);for(let i=0;i<8;i++)assert(control._rr_run(1000,0,0)>=0);assert.equal(proof(),control._rr_proof?control.UTF8ToString(control._rr_proof()):control.UTF8ToString(control._rr_status()),'recording changed execution');const live=proof();service.seek(0);service.seek(end.count);assert.equal(proof(),live,'historical replay changed live state');
 const raw=JSON.parse(core.UTF8ToString(core._rr_inspect_regions()));
 for(const r of raw.regions.filter(r=>r.kind==='ram')){const at=core._rr_inspect_data(r.index);for(let offset=0;offset<r.size;offset+=256){const page=service.page(r.id,offset);assert.deepEqual(page.bytes,core.HEAPU8.slice(at+page.offset,at+page.offset+page.bytes.length),r.id+' reconstruction at '+offset);}}
}
console.log(JSON.stringify({platform,regions:s.regions.length,labels:s.labels.length,activity:s.activity,result:'PASS'}));
