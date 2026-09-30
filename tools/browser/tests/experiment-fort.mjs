// Private image acceptance: boot recipe, explicit preparation edits, no shipped states.
import fs from 'node:fs';import assert from 'node:assert/strict';
import factory from '../../../tools/platform/c64/browser/web/core.js';
import {createExperimentService} from '../../../site/emulators/experiment-worker.js';
import {packState,digest} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const root=new URL('../../../',import.meta.url),read=p=>fs.readFileSync(new URL(p,root));
const wasm=read('tools/platform/c64/browser/web/core.wasm'),core=await factory({wasmBinary:wasm});
const tape=fs.readFileSync(process.argv[2]),data=JSON.parse(read('games/fort-apocalypse-c64/knowledge.json')),roms=['basic','kernal','chargen'].map(n=>read('site/emulators/firmware/c64/'+n+'.rom'));
assert.equal(await digest(tape),data.releases['reference-pal'].media[0].sha256);const put=b=>core.HEAPU8.set(b,core._rr_input());put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));
for(const a of JSON.parse(read('tools/browser/tests/fort-tour-boot.json'))){if(a.run)for(let left=a.run;left;){const n=Math.min(left,1000000);assert(core._rr_run(n)>=0);left-=n;}if(a.key)core._rr_key(...a.key);if(a.joystick)core._rr_joystick(...a.joystick);if(a.play!==undefined)core._rr_play(a.play);}
for(let i=0;i<6;i++)core._rr_run(1000000);assert(core._rr_debug_begin(2,0x9c52,1));let result=0;for(let i=0;i<100&&!result;i++)result=core._rr_debug_run(1000);assert.equal(result,4);
const snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1))),save=()=>{const n=core._rr_state_save();return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};const anchor=save();
if(process.argv[3]){const q=new InputQueue(50);fs.writeFileSync(process.argv[3],await packState({format:1,platform:'c64',media:[{name:'Fort_Apocalypse.tap',size:tape.length,sha256:await digest(tape)}],firmware:await Promise.all(roms.map(digest)),core:await digest(wasm),configuration:{compatibility:true,customFirmware:false},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},anchor));}
let response,host={pending:[{buttons:4}],sequence:9};
const service=createExperimentService({core,knowledge:{status:'matched',releaseId:'reference-pal',data},generation:1,send:(t,m)=>response=m,sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot,captureHost:()=>host,restoreHost:v=>host=v,clearInput:()=>{host={};core._rr_joystick(2,0);}});
let request=0;const command=async type=>{await service.request({type,protocol:1,generation:1,request:++request,id:'upward-probe'});assert(!['failed','restore-failed'].includes(response.phase),response.text);};
await command('experiment-prepare');assert.equal(response.phase,'prepared');const prepared=snapshot().cycle;
await command('experiment-original');assert.equal(response.phase,'original-complete');assert(response.deterministic);
await command('experiment-modified');assert.equal(response.phase,'completed');const branches=response.branches;
const observed=(b,id)=>b.observations.filter(o=>o.predicates.includes(id));assert(observed(branches[0],'dying').length);assert(observed(branches[0],'terrain-contact').length);assert.equal(observed(branches[2],'dying').length,0);
await command('experiment-return');assert.deepEqual(save(),anchor);assert.deepEqual(host,{pending:[{buttons:4}],sequence:9});
console.log(JSON.stringify({result:'PASS',coreSHA256:await digest(wasm),prepared,branches:branches.map(b=>({id:b.id,hash:b.stateSHA256,final:b.final,observations:b.observations.filter(o=>o.predicates.includes('terrain-contact')||o.predicates.includes('dying')).slice(0,3),samples:b.observations.length})),restored:true},null,2));
