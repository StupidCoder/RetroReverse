// Authentic owned-core boot followed by the existing Fort tour and experiment.
// All private input and checkpoint bytes stay in memory/scratch test output.
import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';
import {createC64TestCore} from '../../../../browser/tests/c64-test-core.mjs';
import {createTourService} from '../../../../../site/emulators/tour-worker.js';
import {createExperimentService} from '../../../../../site/emulators/experiment-worker.js';
import {digest} from '../../../../../site/emulators/state.js';
const [firmwareDir,tapePath]=process.argv.slice(2);assert(firmwareDir&&tapePath&&process.env.RR_C64_CORE);
const data=JSON.parse(fs.readFileSync(new URL('../../../../../games/fort-apocalypse-c64/knowledge.json',import.meta.url))),tape=fs.readFileSync(tapePath);
assert.equal(await digest(tape),data.releases['reference-pal'].media[0].sha256);
const roms=['basic','kernal','chargen'].map(n=>fs.readFileSync(path.join(firmwareDir,n+'.rom')));assert.deepEqual(await Promise.all(roms.map(digest)),data.preparedStarts.terrain.firmware);
const core=await createC64TestCore(),put=b=>core.HEAPU8.set(b,core._rr_input()),status=()=>JSON.parse(core.UTF8ToString(core._rr_status())),snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
const run=n=>{while(n){const slice=Math.min(n,100000);assert.equal(core._rr_run(slice,0,0),slice);n-=slice;}};
const until=(pc,budget)=>{assert(core._rr_debug_begin(2,pc,1));let r=0;while(budget&&!r){const n=Math.min(10000,budget);r=core._rr_debug_run(n);budget-=n;}assert.equal(r,4,`target ${pc.toString(16)}: ${core.UTF8ToString(core._rr_error())}`);};
const type=text=>{for(const c of text){core._rr_key(c.charCodeAt(0),1);run(100000);core._rr_key(c.charCodeAt(0),0);run(100000);}};
const save=()=>{const n=core._rr_state_save();assert(n>0);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));
run(3000000);type('LOAD\r');core._rr_play(1);let loaded=false;
for(let i=0;i<10000;i++){run(10000);const s=status();if(s.pulse>=48233&&!s.motor){loaded=true;break;}}assert(loaded);
run(200000);type('RUN\r');until(0x8600,150000000);assert.equal(core._rr_cycle(),115317157);
run(3000000);core._rr_joystick(2,16);until(0x8cdb,1000000);core._rr_joystick(2,0);
const terrainAnchor=save();
let response;const knowledge={status:'matched',releaseId:'reference-pal',data};
const tour=createTourService({core,knowledge,generation:1,send:(type,m)=>response={type,...m},sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot});
for(let i=0;i<3;i++){
 await tour.request({type:i?'tour-continue':'tour-start',protocol:1,generation:1,request:i+1,id:'terrain-runs'});
 assert.equal(response.phase,i===2?'completed':'paused-at-stop',response.text);assert(response.evidence.complete);
 const writes=response.evidence.changes.filter(c=>c.address>=0x503&&c.address<0x603).reduce((n,c)=>n+c.writes,0);assert.equal(writes,[0,215,40][i]);
}
putState(terrainAnchor);
function putState(bytes){const p=core._rr_state_input(bytes.length);assert(p);core.HEAPU8.set(bytes,p);assert(core._rr_state_load(bytes.length));}
run(6000000);until(0x9c52,1000000);const anchor=save();let host={pending:[{buttons:4}],sequence:9};
const experiment=createExperimentService({core,knowledge,generation:2,send:(type,m)=>response={type,...m},sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot,captureHost:()=>host,restoreHost:v=>host=v,clearInput:()=>{host={};core._rr_joystick(2,0);}});
// The shipped experiment targets the previous core's timing/RNG context.
// Its start predicate passes here, but the native teleport produces another
// horizontal location. Preserve the exact package guards and verify rollback;
// this is a known compatibility gap, not a successful AI comparison.
const experimentAnchor=snapshot();
assert.equal(experimentAnchor.nextPC,0x9c52);
let beforeRollback=null;const load=core._rr_state_load;
core._rr_state_load=n=>{const ram=core.HEAPU8.subarray(core._rr_ram(),core._rr_ram()+65536);beforeRollback={pc:snapshot().nextPC,x:ram[105],cameraX:ram[86]};return load(n);};
await experiment.request({type:'experiment-prepare',protocol:1,generation:2,request:1,id:'upward-probe'});
assert.equal(response.phase,'failed');assert.match(response.text,/Experiment invariant failed/);
assert.deepEqual(beforeRollback,{pc:0x9c52,x:209,cameraX:188});
assert.deepEqual(save(),anchor);assert.deepEqual(host,{pending:[{buttons:4}],sequence:9});assert.equal(experiment.owns(),false);
console.log('PASS owned Fort browser adapter: authentic KERNAL/Novaload boot, terrain tour (215/40 writes), safe rejection and full rollback of incompatible AI preparation');
console.log('KNOWN GAP: shipped AI teleport preparation expects player/camera X=53/34; owned boot produces 209/188. Original/patched gameplay comparison remains unvalidated.');
