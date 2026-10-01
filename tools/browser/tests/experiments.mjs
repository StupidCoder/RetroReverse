import assert from 'node:assert/strict';
import {createC64TestCore} from './c64-test-core.mjs';
import {createExperimentService} from '../../../site/emulators/experiment-worker.js';
const core=await createC64TestCore();
const put=b=>core.HEAPU8.set(b,core._rr_input()),ram=()=>core.HEAPU8.subarray(core._rr_ram(),core._rr_ram()+65536),snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
const save=()=>{const n=core._rr_state_save();return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
const edit=(address,before,after)=>({kind:'ram',address,before:[before],after:[after],explanation:'test'});
const base={releases:['r'],guards:[{address:0x800,bytes:[0xee,0,2]}],start:{pc:0x800},setup:[{require:[],edits:[edit(0x201,0,3)],cycles:0,pc:0x800,budget:100,assertions:[{address:0x201,value:3}]}],invariants:[{pc:0x800}],patch:[edit(0x800,0xee,0xce)],duration:300,wallBudget:1000,probePC:0x800,inputs:[{cycle:0,buttons:0},{cycle:50,buttons:1},{cycle:100,buttons:0}],watches:[{id:'counter',address:0x200}],observations:[{id:'changed',when:{address:0x200,value:0}}]};
let service,response,host,anchor,request,onYield=()=>{},busy=false;
function setup(def=base){put(new Uint8Array(20480));assert(core._rr_init(8192,8192,4096));const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;put(tape);assert(core._rr_tape(21));const b=new Uint8Array(65536).fill(0xea);b.set([0xee,0,2,0x4c,0,8],0x800);b[0x200]=b[0x201]=0;put(b);assert(core._rr_prepare(0x800,0));host={pending:[{buttons:2}],sequence:7};anchor=save();request=0;
 service=createExperimentService({core,knowledge:{status:'matched',releaseId:'r',data:{experiments:{e:def}}},generation:1,send:(t,m)=>response={type:t,...m},sleep:async()=>onYield(),busy:()=>busy,paint:()=>{},snapshot,captureHost:()=>host,restoreHost:v=>host=v,clearInput:()=>{host={};core._rr_joystick(2,0);}});
}
const command=async(type,extra={})=>{await service.request({type:'experiment-'+type,protocol:1,generation:1,request:++request,id:'e',...extra});return response;};
const restored=()=>{assert.deepEqual(save(),anchor);assert.deepEqual(host,{pending:[{buttons:2}],sequence:7});assert(!service.owns());};
setup();await command('prepare');assert.equal(response.phase,'prepared');assert(service.owns());await command('original');assert(response.deterministic);assert.equal(response.branches.length,2);assert.equal(response.branches[0].stateSHA256,response.branches[1].stateSHA256);await command('modified');assert.equal(response.phase,'completed');assert.equal(new Set(response.branches.map(b=>b.id)).size,3);assert.notEqual(response.branches[0].stateSHA256,response.branches[2].stateSHA256);for(const b of response.branches)assert(b.observations.every(o=>o.branchId===b.id));await command('keep');assert.equal(response.phase,'modified-session');assert(!service.owns());assert.equal(ram()[0x800],0xce);busy=true;await command('cancel');assert.equal(response.type,'experiment-rejected');assert.equal(ram()[0x800],0xce);busy=false;await command('return');restored();
setup({...base,patch:[edit(0x201,3,9),edit(0x800,0,0xea)]});await command('prepare');await command('original');await command('modified');assert.equal(response.phase,'failed');restored();
setup({...base,setup:[{...base.setup[0],edits:[edit(0x201,0,9),edit(0x200,99,8)]}]});await command('prepare');assert.equal(response.phase,'failed');restored();
setup({...base,setup:[{...base.setup[0],cycles:100}]});onYield=()=>service.cancel();await command('prepare');onYield=()=>{};assert.equal(response.phase,'cancelled');restored();
setup();await command('prepare');onYield=()=>service.cancel();await command('original');onYield=()=>{};assert.equal(response.phase,'cancelled');restored();
setup();await command('prepare');await command('original');onYield=()=>service.cancel();await command('modified');onYield=()=>{};assert.equal(response.phase,'cancelled');restored();
setup();await command('prepare');await command('cancel');restored();
setup();await command('prepare',{generation:2});assert.equal(response.type,'experiment-rejected');restored();busy=true;await command('prepare');assert.equal(response.type,'experiment-rejected');busy=false;
setup({...base,wallBudget:1,setup:[{...base.setup[0],cycles:100}]});onYield=()=>new Promise(r=>setTimeout(r,5));await command('prepare');onYield=()=>{};assert.equal(response.phase,'failed');restored();
// Deliberate replay contamination must stop before enabling the patch.
setup();await command('prepare');onYield=()=>{if(response.phase==='running-original-2')ram()[0x300]++;};await command('original');onYield=()=>{};assert.equal(response.phase,'failed');assert.match(response.text,/not deterministic/);restored();
// A no-op repair hypothesis produces equal outcomes and is still an honest completed comparison.
setup({...base,patch:[edit(0x800,0xee,0xee)]});await command('prepare');await command('original');await command('modified');assert.equal(response.phase,'completed');assert.equal(response.branches[0].stateSHA256,response.branches[2].stateSHA256);await command('return');restored();
console.log('PASS experiments: atomic edits, identical replay, branch provenance, failed hypothesis, cancellation in every stage, stale/busy requests, rollback, nondeterminism and complete session/input restoration');
