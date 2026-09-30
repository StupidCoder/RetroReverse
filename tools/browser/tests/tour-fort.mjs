// Private-media acceptance. The checked-in recipe contains only input/timing,
// extracted from lesson_native.cpp's verified boot, never copyrighted RAM.
// Usage: node tools/browser/tests/tour-fort.mjs path/to/Fort_Apocalypse.tap [output.rrstate]
import fs from 'node:fs';
import assert from 'node:assert/strict';
import factory from '../../../site/emulators/cores/c64/core.js';
import {createTourService} from '../../../site/emulators/tour-worker.js';
import {packState,digest} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const read=p=>fs.readFileSync(new URL(p,import.meta.url));
const tape=fs.readFileSync(process.argv[2]),knowledge=JSON.parse(read('../../../games/fort-apocalypse-c64/knowledge.json'));
assert.equal(await digest(tape),knowledge.releases['reference-pal'].media[0].sha256);
const wasm=read('../../../site/emulators/cores/c64/core.wasm'),core=await factory({wasmBinary:wasm});
const roms=['basic','kernal','chargen'].map(n=>read('../../../site/emulators/firmware/c64/'+n+'.rom'));
const put=b=>core.HEAPU8.set(b,core._rr_input());put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));
for(const a of JSON.parse(read('./fort-tour-boot.json'))){if(a.run){for(let left=a.run;left;){const n=Math.min(left,1000000);assert(core._rr_run(n)>=0);left-=n;}}if(a.key)core._rr_key(...a.key);if(a.joystick)core._rr_joystick(...a.joystick);if(a.play!==undefined)core._rr_play(a.play);}
const snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
assert.equal(snapshot().nextPC,0x8cdb);assert.equal(core._rr_cycle(),118436151);
if(process.argv[3]){
 const n=core._rr_state_save(),q=new InputQueue(50);
 const bytes=await packState({format:1,platform:'c64',media:[{name:'Fort_Apocalypse.tap',size:tape.length,sha256:await digest(tape)}],firmware:await Promise.all(roms.map(digest)),core:await digest(wasm),configuration:{compatibility:true,customFirmware:false},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n));fs.writeFileSync(process.argv[3],bytes);
}
let response;const service=createTourService({core,knowledge:{status:'matched',releaseId:'reference-pal',data:knowledge},generation:1,send:(type,m)=>response=m,sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot});
const reports=[];
for(const [i,type] of ['tour-start','tour-continue','tour-continue'].entries()){
 await service.request({type,protocol:1,generation:1,request:i+1,id:'terrain-runs'});assert.equal(response.phase,i===2?'completed':'paused-at-stop');assert(response.evidence.complete);
 const terrain=response.evidence.changes.filter(c=>c.address>=0x503&&c.address<0x603);assert.equal(terrain.reduce((n,c)=>n+c.writes,0),[0,215,40][i]);
 reports.push({stop:response.stop.id,cycle:response.snapshot.cycle,cycles:response.evidence.cycles,terrainWrites:terrain.reduce((n,c)=>n+c.writes,0),terrainDifferences:terrain.filter(c=>c.before!==c.after).length});
}
console.log(JSON.stringify({result:'PASS',coreSHA256:await digest(wasm),reports},null,2));
