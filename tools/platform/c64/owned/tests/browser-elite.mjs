// Optional private-media acceptance of the exported module and existing tour
// service. No ROM traps, injected game RAM or redistributed checkpoint bytes.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createC64TestCore} from '../../../../browser/tests/c64-test-core.mjs';
import {createTourService} from '../../../../../site/emulators/tour-worker.js';
import {prepareStart} from '../../../../../site/emulators/prepared-start.js';
import {digest} from '../../../../../site/emulators/state.js';
const [firmwareDir,tapePath]=process.argv.slice(2);assert(firmwareDir&&tapePath&&process.env.RR_C64_CORE);
const tape=fs.readFileSync(tapePath),roms=['basic','kernal','chargen'].map(n=>fs.readFileSync(path.join(firmwareDir,n+'.rom')));
const pkg=JSON.parse(fs.readFileSync(new URL('../../../../../games/elite-c64/knowledge.json',import.meta.url)));
assert.equal(await digest(tape),pkg.releases['reference-pal'].media[0].sha256);
const identity={core:await digest(fs.readFileSync(path.join(path.dirname(process.env.RR_C64_CORE),'core.wasm'))),firmware:await Promise.all(roms.map(digest))};
assert.deepEqual(identity.firmware,pkg.preparedStarts['loader-prefix'].firmware);
const core=await createC64TestCore(),put=b=>core.HEAPU8.set(b,core._rr_input());
function fresh(){put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));}
const raw=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1))),status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
const run=n=>{while(n){const slice=Math.min(n,100000);assert.equal(core._rr_run(slice,0,0),slice);n-=slice;}};
fresh();run(3000000);for(const code of [76,79,65,68,13]){core._rr_key(code,1);run(100000);core._rr_key(code,0);run(100000);}core._rr_play(1);
assert(core._rr_debug_begin(2,0x378,0));let result=0;for(let i=0;i<8000&&!result;i++)result=core._rr_debug_run(10000);assert.equal(result,4);
assert.equal(core._rr_cycle(),37906392);assert.equal(status().pulse,52797);
const n=core._rr_state_save();assert(n>0&&n<=4*1024*1024);const saved=core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);
// Regenerate only in memory for this build. Production knowledge packages
// remain bound to the shipped core until the complete C6 integration gate.
const recipe={...pkg.preparedStarts['loader-prefix'],...identity,sha256:await digest(saved)};
fresh();const prepared=await prepareStart({core,recipe,identity});assert.equal(prepared.cached,false);assert.equal(await digest(prepared.bytes),recipe.sha256);
run(50);const cached=await prepareStart({core,recipe,identity,cached:saved.buffer});assert(cached.cached);assert.equal(core._rr_cycle(),37906392);
let response;const tour=createTourService({core,knowledge:{status:'matched',releaseId:'reference-pal',data:pkg},generation:1,send:(type,m)=>response={type,...m},sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot:raw});
for(let i=0;i<4;i++){await tour.request({type:i?'tour-continue':'tour-start',id:'first-byte',protocol:1,generation:1,request:i+1});assert.equal(response.phase,i===3?'completed':'paused-at-stop',response.text);assert(response.evidence.complete);}
assert.equal(response.evidence.changes.find(c=>c.address===0x300).writes,1);
// Continue beyond the shipped tour, checking every original vector byte
// against a direct decoder of the initial nine-pulse framing.
const expected=[];for(let i=0;i<52;i++){let b=0;const at=0xd7a5+(i+5)*9;assert.equal(tape[at],0x5d);for(let j=1;j<=8;j++){assert([0x30,0x5d].includes(tape[at+j]));b=(b<<1)|(tape[at+j]===0x5d?1:0);}expected.push(b);}
const ram=()=>core.HEAPU8.subarray(core._rr_ram(),core._rr_ram()+65536);assert.equal(ram()[0x300],expected[0]);
for(let i=1;i<52;i++){assert(core._rr_debug_watch(0x300+i,0,0,0x3b1));result=0;for(let k=0;k<1000&&!result;k++)result=core._rr_debug_run(1000);assert.equal(result,13);assert.equal(ram()[0x300+i],expected[i]);}
console.log('PASS owned Elite browser adapter: authentic boot, independently regenerated recipe/cache, existing four-stop tour, 52 verified vector stores; later loader stages unvalidated');
