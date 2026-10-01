// Verify only the authored loader prefix; later compatibility is explicitly gated.
import fs from 'node:fs';import assert from 'node:assert/strict';
import factory from '../../../site/emulators/cores/c64/core.js';
import {prepareStart} from '../../../site/emulators/prepared-start.js';
import {createTourService} from '../../../site/emulators/tour-worker.js';
import {digest} from '../../../site/emulators/state.js';
const read=p=>fs.readFileSync(new URL('../../../'+p,import.meta.url)),data=JSON.parse(read('games/elite-c64/knowledge.json')),tape=fs.readFileSync(process.argv[2]),wasm=read('site/emulators/cores/c64/core.wasm'),roms=['basic','kernal','chargen'].map(n=>read('site/emulators/firmware/c64/'+n+'.rom'));
assert.equal(await digest(tape),data.releases['reference-pal'].media[0].sha256);const core=await factory({wasmBinary:wasm}),put=b=>core.HEAPU8.set(b,core._rr_input());put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));
await prepareStart({core,recipe:data.preparedStarts['loader-prefix'],identity:{core:await digest(wasm),firmware:await Promise.all(roms.map(digest))}});
let response;const snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1))),service=createTourService({core,knowledge:{status:'matched',releaseId:'reference-pal',data},generation:1,send:(t,m)=>response=m,sleep:async()=>{},busy:()=>false,paint:()=>{},snapshot});
const stops=[];for(let i=0;i<4;i++){await service.request({type:i?'tour-continue':'tour-start',id:'first-byte',protocol:1,generation:1,request:i+1});assert.equal(response.phase,i===3?'completed':'paused-at-stop',response.text);assert(response.evidence.complete);stops.push({id:response.stop.id,cycle:response.snapshot.cycle,pulse:JSON.parse(core.UTF8ToString(core._rr_status())).pulse,writes:response.evidence.writes});}
const writes=response.evidence.changes.filter(c=>c.address===0x300);assert.equal(writes[0].writes,1);assert.equal(writes[0].before,0x8b);assert.equal(writes[0].after,0x8b);
// Probe the first known divergence separately; never present this as a valid tour stop.
const expected=[0x8b,0xe3,0x83,0xa4,0x7c,0xa5,0xa8],actual=[0x8b];for(let i=1;i<expected.length;i++){assert(core._rr_debug_begin(2,0x3b1,1));let r=0;for(let j=0;j<1000&&!r;j++)r=core._rr_debug_run(1000);assert.equal(r,4);actual.push(snapshot().registers.A);}
const mismatch=actual.findIndex((b,i)=>b!==expected[i]),compatibility={status:mismatch<0?'prefix-matches':'blocked',firstMismatch:mismatch<0?null:{address:0x300+mismatch,expected:expected[mismatch],actual:actual[mismatch]},note:'Expected prefix from independently extracted segment-2 vector block; later loader stages and gameplay remain unvalidated.'};
console.log(JSON.stringify({result:'PASS',compatibility,coreSHA256:await digest(wasm),stops,scope:'Initial loader prefix only; full loading remains incompatible'},null,2));
