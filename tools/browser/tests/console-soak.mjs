// Private-media regression for the per-primitive sampler lifetime leak.
// Run from a 3D checkpoint; no capture. Budget is an explicit test constraint,
// not a promise that all games fit this heap size.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {loadCore} from './wasm-harness.mjs';
const [platform,image,state,fields='600',budgetMiB='512']=process.argv.slice(2);
const c=await loadCore(platform,image),json=n=>JSON.parse(c.UTF8ToString(c['_rr_'+n]()));
const b=fs.readFileSync(state),p=c._rr_state_input(b.length);assert(p);c.HEAPU8.set(b,p);assert(c._rr_state_load(b.length));
const begin=json('status').frames;let reported=begin;const samples=[];
while(json('status').frames<begin+Number(fields)){
 assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));
 const frame=json('status').frames;
 if(frame>=reported+60){reported=frame;const sample={frame,heapBytes:c.HEAPU8.length};samples.push(sample);console.log(JSON.stringify(sample));assert(sample.heapBytes<=Number(budgetMiB)*1048576,'heap budget exceeded');}
}
console.log(JSON.stringify({pass:true,platform,fields:Number(fields),budgetMiB:Number(budgetMiB),samples,proof:json('proof')}));
