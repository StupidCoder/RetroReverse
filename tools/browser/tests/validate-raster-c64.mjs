// Private media/checkpoint validation; no game bytes are emitted.
import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';
import assert from 'node:assert/strict';
const [image,statePath]=process.argv.slice(2);
if(!image||!statePath)throw Error('Usage: validate-raster-c64.mjs TAP raw-state');
const c=await loadCore('c64',image),json=name=>JSON.parse(c.UTF8ToString(c['_rr_'+name]()));
const bytes=fs.readFileSync(statePath);c.HEAPU8.set(bytes,c._rr_state_input(bytes.length));assert(c._rr_state_load(bytes.length));
const run=n=>{while(n){const slice=Math.min(n,1000000);assert(c._rr_run(slice,0,0)>=0);n-=slice;}};
const boundary=()=>assert(c._rr_run(20000,7,0)>0);
const save=()=>{const n=c._rr_state_save();assert(n>0);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
run(Number(process.env.ADVANCE_CYCLES||0));
if(process.env.SAVE_STATE)fs.writeFileSync(process.env.SAVE_STATE,save());
// Populate writers after loading the state, then capture one complete interval.
boundary();boundary();c._rr_capture_begin();boundary();c._rr_capture_end();
const info=json('raster_info');assert(info.complete);assert.equal(info.lines.length,272);
const status=json('status'),live=save(),timings=[],charsets=new Set(),changes=[];
let pixels=0,sources=0,scannerBytes=0;const scannerWriters=new Set();
for(const row of info.lines){
  let t=performance.now();const state=JSON.parse(c.UTF8ToString(c._rr_raster_seek(row.line)));timings.push(performance.now()-t);assert(state.complete);charsets.add(state.charsetBase);
  if(state.changeCount)changes.push({line:row.line,raster:state.raster,registers:state.changes.filter(c=>c.space===2).map(c=>({address:c.address,before:c.before,after:c.after,pc:c.pc})),graphicsWrites:state.tileWrites,screenWrites:state.mapWrites});
  for(let x=24;x<392;x+=37){
    const p=JSON.parse(c.UTF8ToString(c._rr_raster_pixel(2,x,row.line)));assert(p.complete);pixels++;
    for(const ref of p.contributors){
      if(ref.missing||ref.space!==0)continue;sources++;
      if(ref.writer){assert.equal(ref.writer.value,ref.value);assert(ref.writer.cycle<=ref.fetchCycle);}
      if(ref.address>=0x52e0&&ref.address<=0x53ff){scannerBytes++;if(ref.writer)scannerWriters.add(ref.writer.pc);}
    }
  }
}
for(const line of [0,104,160,271]){
 JSON.parse(c.UTF8ToString(c._rr_raster_seek(line)));
 for(const panel of [0,1])for(const [x,y]of [[100,70],[200,120],[250,200]]){
  const p=JSON.parse(c.UTF8ToString(c._rr_raster_pixel(panel,x,y)));assert(p.complete);
  for(const r of p.contributors)if(!r.missing&&r.space===0&&r.writer)assert.equal(r.value,r.writer.value);
 }
}
assert.deepEqual(json('status'),status);assert.deepEqual(save(),live);
const rgba=p=>c.HEAPU8.slice(p,p+392*272*4);
assert.deepEqual(rgba(c._rr_raster_frame(2)),rgba(c._rr_frame()));
timings.sort((a,b)=>a-b);
console.log(JSON.stringify({platform:'c64',runtime:'Node.js WASM',pass:true,info:{lines:info.lines.length,bytes:info.bytes},capture:json('capture_info'),pixels,sources,scannerBytes,scannerWriters:[...scannerWriters],charsetBases:[...charsets],medianSeekMs:timings[136],maxSeekMs:Math.max(...timings),checks:{fullStateUnchanged:true,finalOutputMatches:true,historicalRAMWriters:true},changes},null,2));
