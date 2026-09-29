// Private-media validation of the shipped WASM core. Checkpoints are raw core states.
// Usage: node tools/browser/tests/validate-raster-gg.mjs ROM [checkpoint]
import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';

const [rom,checkpoint]=process.argv.slice(2);
if(!rom)throw Error('Usage: validate-raster-gg.mjs ROM [raw checkpoint]');
const c=await loadCore('gg',rom);
const json=name=>JSON.parse(c.UTF8ToString(c['_rr_'+name]()));
const run=n=>{
  const end=json('status').frames+n;
  while(json('status').frames<end)assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));
};
const save=()=>{
  const size=c._rr_state_save();assert(size>0);
  return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+size);
};
if(checkpoint){
  const b=fs.readFileSync(checkpoint);c.HEAPU8.set(b,c._rr_state_input(b.length));assert(c._rr_state_load(b.length));
}else run(600);
assert(c._rr_capture_begin());run(1);assert(c._rr_capture_end());
const info=json('raster_info'),proof=json('proof'),live=save();
assert(info.complete);assert.equal(info.lines.length,144);
const timings=[],changes=[],paletteWriters=new Set();let samples=0,sources=0;
for(const {line}of info.lines){
  const start=performance.now();
  const state=JSON.parse(c.UTF8ToString(c._rr_raster_seek(line)));
  timings.push(performance.now()-start);assert(!state.error);
  if(state.changes.length||state.colorWrites)changes.push({line,raster:state.raster,scrollX:state.scrollX,scrollY:state.scrollY,registers:state.changes,paletteWrites:state.colorWrites});
  for(let x=0;x<160;x+=13){
    const pixel=JSON.parse(c.UTF8ToString(c._rr_raster_pixel(2,x,line)));
    assert(pixel.complete);samples++;
    for(const w of pixel.candidates){
      for(const {address,size,value}of w.sources){
        if(!address)continue;
        const history=JSON.parse(c.UTF8ToString(c._rr_source(address,size,pixel.sourceBefore,value)));
        assert(history.complete,JSON.stringify({line,x,address,history}));sources++;
        if(address>=0x14000&&address<0x14040)for(const writer of history.contributors||[])paletteWriters.add(writer.pc);
      }
    }
  }
}
assert.deepEqual(json('proof'),proof);
assert.deepEqual(save(),live,'Inspection changed serialized machine state');
const rgba=pointer=>c.HEAPU8.slice(pointer,pointer+160*144*4);
assert.deepEqual(rgba(c._rr_raster_frame(2)),rgba(c._rr_frame()));
timings.sort((a,b)=>a-b);
console.log(JSON.stringify({game:path.basename(rom),runtime:'Node.js WASM',pass:true,
  lines:info.lines.length,sampledPixels:samples,sourceChecks:sources,paletteWriters:[...paletteWriters],capture:json('capture_info'),
  medianSeekMs:timings[Math.floor(timings.length/2)],maxSeekMs:Math.max(...timings),changes,proof,
  checks:{pixelReconstruction:true,historicalSources:true,finalOutputMatches:true,fullStateUnchanged:true}},null,2));
