// Local-media benchmark: deterministic driving state, input and checkpoints.
// Pass a core directory to compare the deployed baseline against a candidate.
import fs from 'node:fs';import {pathToFileURL} from 'node:url';import assert from 'node:assert/strict';
const [folder,disc,state,frames='120']=process.argv.slice(2);
const factory=(await import(pathToFileURL(folder+'/core.js'))).default;
const core=await factory({wasmBinary:fs.readFileSync(folder+'/core.wasm')});
const bytes=fs.readFileSync(disc);
globalThis.FileReaderSync=class{readAsArrayBuffer(b){return b.buffer.slice(b.byteOffset,b.byteOffset+b.byteLength);}};
core.discFile={slice:(a,b)=>bytes.subarray(a,b)};
const check=x=>assert(x,core.UTF8ToString(core._rr_error()));
const json=n=>JSON.parse(core.UTF8ToString(core[n]()));
check(core._rr_init_config(bytes.length,1));
const b=fs.readFileSync(state);core.HEAPU8.set(b,core._rr_state_input(b.length));check(core._rr_state_load(b.length));
core._rr_pad(134217728);
const before=json('_rr_status'),profileBefore=json('_rr_profile'),start=performance.now(),checkpoints=[];
let next=before.frame+30;
while(json('_rr_status').frame<before.frame+Number(frames)){
 check(core._rr_run_slice(10000));
 if(json('_rr_status').frame>=next){checkpoints.push(json('_rr_proof'));next+=30;}
}
const seconds=(performance.now()-start)/1000;
console.log(JSON.stringify({seconds,displays:Number(frames),displaysPerSecond:Number(frames)/seconds,profileBefore,profile:json('_rr_profile'),checkpoints}));
