import fs from 'node:fs';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {loadCore} from './wasm-harness.mjs';
import {unpackState} from '../../../site/emulators/state.js';

// Private media/checkpoint paths stay outside the measured, publishable result.
const [platform,rom,statePath,count='120']=process.argv.slice(2);
const core=await loadCore(platform,rom);
const json=name=>JSON.parse(core.UTF8ToString(core[name]()));
let state=fs.readFileSync(statePath);
if(state.subarray(0,8).toString()==='RRSTATE1')state=(await unpackState(new Blob([state]))).payload;
const restore=()=>{const p=core._rr_state_input(state.length);core.HEAPU8.set(state,p);assert(core._rr_state_load(state.length),core.UTF8ToString(core._rr_error()));};
const next=()=>{let f=json('_rr_status').frames;while(json('_rr_status').frames===f)assert(core._rr_run(10000)>=0,core.UTF8ToString(core._rr_error()));};
restore();for(let i=0;i<3;i++)next();restore();
const before=json('_rr_profile'),checkpoints=[];
let seconds=0;
for(let i=1;i<=Number(count);i++){
 const t=performance.now();next();seconds+=(performance.now()-t)/1000;
 if(i%30===0||i===Number(count))checkpoints.push({frame:i,...json('_rr_proof')});
}
const after=json('_rr_profile');
const profile=after.buckets.map(b=>({...b,ms:b.ms-(before.buckets.find(a=>a.name===b.name)?.ms||0)}));
const size=core._rr_state_save();assert(size);
const stateSha256=createHash('sha256').update(core.HEAPU8.subarray(core._rr_state_data(),core._rr_state_data()+size)).digest('hex');
console.log(JSON.stringify({platform,frames:Number(count),seconds,fps:Number(count)/seconds,heap:core.HEAPU8.length,profile,checkpoints,stateSha256},null,2));
