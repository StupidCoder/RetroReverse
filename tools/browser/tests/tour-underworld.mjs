// Private media: node tools/browser/tests/tour-underworld.mjs /game/UW.EXE /native/gameplay.state [output.rrstate]
import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';
import {loadCore} from './wasm-harness.mjs';
import {createDOSInspector} from '../../../site/emulators/dos-inspector.js';
import {createDebugService} from '../../../site/emulators/debug-worker.js';
import {createTourService} from '../../../site/emulators/tour-worker.js';
import {packState,digest} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const [exe,state,out]=process.argv.slice(2),core=await loadCore('dos',exe),b=fs.readFileSync(state);
core.HEAPU8.set(b,core._rr_state_input(b.length));assert(core._rr_state_load(b.length));
// Match browser restore: release host controls before anchoring the tour.
for(let key=0;key<128;key++)core._rr_key(key,0);core._rr_pad(0);core._rr_mouse(0,0,0);
const data=JSON.parse(fs.readFileSync(new URL('../../../games/ultima-underworld-pc/knowledge.json',import.meta.url))),root=path.dirname(exe);
for(const m of data.releases['reference-installed'].fileSet.members){const p=path.join(root,m.path.toUpperCase());const bytes=fs.readFileSync(p);assert.equal(bytes.length,m.size);assert.equal(await digest(bytes),m.sha256);}
const knowledge={status:'matched',releaseId:'reference-installed',data},adapter=createDOSInspector(core,knowledge);
let response;const common={core,knowledge,generation:1,send:(type,m)=>response=m,sleep:async()=>{},busy:()=>false,paint:()=>{},...adapter};
const debug=createDebugService({...common,platform:'dos',applyInputs:()=>{}}),initial=debug.snapshot();assert.equal(initial.modules.raster.status,'verified');assert.equal(initial.modules.geometry.status,'verified');
if(out){const n=core._rr_state_save(),media=[],selected=[];async function walk(dir,prefix=''){for(const e of fs.readdirSync(dir,{withFileTypes:true})){if(e.isDirectory())await walk(path.join(dir,e.name),prefix+e.name+'/');else{const bytes=fs.readFileSync(path.join(dir,e.name));selected.push({path:prefix+e.name,size:bytes.length});media.push({name:(prefix+e.name).toLowerCase(),size:bytes.length,sha256:await digest(bytes)});}}}await walk(root);fs.writeFileSync(out+'.files.json',JSON.stringify(selected));media.sort((a,b)=>a.name.localeCompare(b.name));const q=new InputQueue(70);q.mouseMask=768;const wasm=fs.readFileSync(process.env.CORE_DIR?path.join(process.env.CORE_DIR,'core.wasm'):new URL('../../../site/emulators/cores/dos/core.wasm',import.meta.url));fs.writeFileSync(out,await packState({format:1,platform:'dos',media,firmware:[],core:await digest(wasm),configuration:{compatibility:true,customFirmware:false,executable:path.basename(exe).toLowerCase()},input:{...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:0,lastX:0,lastY:0,inputSequence:0,lastInputStep:0}},core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n)));}
const tour=createTourService({...common,snapshot:()=>debug.snapshot(),maxCheckpointBytes:128*1024*1024});const reports=[];let beforeCopy;
for(let i=0;i<data.tours.renderer.stops.length;i++){
 await tour.request({type:i?'tour-continue':'tour-start',protocol:1,generation:1,request:i+1,id:'renderer'});assert.equal(response.phase,i===3?'completed':'paused-at-stop',response.text);assert(response.evidence.complete);assert.equal(response.snapshot.mode,'real16');assert.equal(response.snapshot.state.cycle,response.snapshot.cycle);
 const p=core._rr_frame(),hash=await digest(core.HEAPU8.slice(p,p+320*200*4));if(i===2)beforeCopy=hash;if(i===3)assert([1,2,3,4].reduce((n,r)=>n+(response.evidence.storageWrites[r]||0),0)>0,'planar copy records actual VGA writes, including unchanged pixels');
 reports.push({stop:response.stop.id,step:response.snapshot.cycle,intervalSteps:response.evidence.cycles,ramWrites:response.evidence.writes,storageWrites:response.evidence.storageWrites,buffers:response.snapshot.buffers.length,displaySHA256:hash});
}
const final=response.snapshot.cycle,finalProof=core.UTF8ToString(core._rr_proof());await tour.request({type:'tour-explore',protocol:1,generation:1,request:5});core._rr_run(1);await tour.request({type:'tour-restore',protocol:1,generation:1,request:6});assert.equal(response.snapshot.cycle,final);assert.equal(core.UTF8ToString(core._rr_proof()),finalProof);
// A signature change invalidates a named breakpoint without advancing the CPU.
const snap=debug.snapshot(),ram=core._rr_ram(),a=snap.modules.geometry.base+0x614f,old=core.HEAPU8[ram+a];core.HEAPU8[ram+a]^=1;const clock=core._rr_cycle();await debug.request({type:'debug-until',protocol:1,generation:1,request:7,snapshotId:snap.snapshotId,cycle:snap.cycle,bank:snap.bank,functionId:'projection',target:snap.functions.projection.address});assert.equal(response.reason,'rejected');assert.equal(core._rr_cycle(),clock);core.HEAPU8[ram+a]=old;
console.log(JSON.stringify({result:'PASS',initialStep:initial.cycle,reports},null,2));
