import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createDebugService} from '../../../site/emulators/debug-worker.js';
import {createExecutionGate} from '../../../site/emulators/execution-gate.js';
import {decode6502,disassemble6502} from '../../../site/emulators/disassembly6502.js';
import {opcodes} from '../../../site/emulators/opcodes6502.js';
import factory from '../../../site/emulators/cores/c64/core.js';
assert.equal(Object.keys(opcodes).length,151);
const modes={imp:'',acc:'A',imm:'#$80',zp:'$80',zpx:'$80,X',zpy:'$80,Y',izx:'($80,X)',izy:'($80),Y',rel:'$0782',abs:'$1280',abx:'$1280,X',aby:'$1280,Y',ind:'($1280)'};
for(const [code,op] of Object.entries(opcodes)){
 const d=decode6502([+code,0x80,0x12],0x800);assert.equal(d.length,op.length);assert.equal(d.text,op.mnemonic+(modes[op.mode]?' '+modes[op.mode]:''));
}
assert.equal(decode6502([0x55,0x20],0).text,'EOR $20,X');
assert.equal(decode6502([0xd0,0xfe],0xffff).text,'BNE $FFFF');
assert.equal(decode6502([0x4c,null,0],0).supported,false);
assert.equal(disassemble6502([0xea,2,0xea],0).length,2);
const gate=createExecutionGate(),token=gate.acquire('seek');assert(token);assert.equal(gate.acquire('step'),null);gate.release({});assert.equal(gate.owner,'seek');gate.release(token);assert.equal(gate.owner,null);
const play=gate.acquire('run');gate.cancel(['run']);const next=gate.acquire('seek');gate.release(play);assert.equal(gate.owner,'seek');gate.release(next);
const core=await factory({wasmBinary:fs.readFileSync(new URL('../../../site/emulators/cores/c64/core.wasm',import.meta.url))});
const put=b=>core.HEAPU8.set(b,core._rr_input());
put(new Uint8Array(20480));assert(core._rr_init(8192,8192,4096));
const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;put(tape);assert(core._rr_tape(21));
const ram=new Uint8Array(65536).fill(0xea);ram.set([0xee,0,2,0x4c,0,8],0x800);ram[0x200]=0;put(ram);assert(core._rr_prepare(0x800,0));
const messages=[];let occupied=false,onYield=()=>{},paints=0;
const service=createDebugService({core,send:(type,m)=>messages.push({type,...m}),sleep:async()=>onYield(),busy:()=>occupied,paint:()=>paints++,applyInputs:()=>{},generation:1,knowledge:{status:'unknown'}});
let request=0;async function command(type,args={}){const m={type,protocol:1,generation:1,request:++request,...args};await service.request(m);return messages.filter(r=>r.request===m.request).at(-1);}
let response=await command('debug-snapshot'),s=response.snapshot;assert.equal(s.nextPC,0x800);assert.equal(s.cycle,'0');
const action=s=>({snapshotId:s.snapshotId,cycle:s.cycle,bank:s.bank});
response=await command('debug-until',{...action(s),target:0x800});assert.equal(response.reason,'target');assert.equal(response.cycles,0);
s=response.snapshot;response=await command('debug-step',action(s));assert.equal(response.reason,'instruction');assert.equal(response.snapshot.nextPC,0x803);assert.equal(response.retired,1);
s=response.snapshot;response=await command('debug-until',{...action(s),target:0x800});assert.equal(response.reason,'target');
s=response.snapshot;response=await command('debug-until',{...action(s),target:0x800,resume:'next-match'});assert.equal(response.cycles,9);
s=response.snapshot;response=await command('debug-step',{...action(s),snapshotId:0});assert.equal(response.reason,'rejected');
occupied=true;response=await command('debug-step',action(s));assert.equal(response.reason,'rejected');occupied=false;
s=(await command('debug-snapshot')).snapshot;response=await command('debug-step',{...action(s),generation:0});assert.equal(response.reason,'rejected');
onYield=()=>{assert(service.active());service.cancel();};s=(await command('debug-snapshot')).snapshot;response=await command('debug-until',{...action(s),target:0x900});assert.equal(response.reason,'cancelled');assert(response.cycles<=1000);assert(!service.active());assert.equal(messages.filter(m=>m.request===request&&m.type==='debug-result').length,1);
onYield=()=>{};s=(await command('debug-snapshot')).snapshot;response=await command('debug-until',{...action(s),target:0x900,cycleBudget:32});assert.equal(response.reason,'budget');assert.equal(response.cycles,32);
console.log('Decoder (151 official opcodes), execution gate, real WASM jobs, stale requests, current-PC resume, cancellation and budgets: PASS');
// Real WASM integration: every watch carries bytes from the same stopped CPU.
const {knowledgePackages}=await import('../../../site/emulators/knowledge-data.js');
const pkg=knowledgePackages.find(p=>p.id==='fort-apocalypse-c64');
const stateMessages=[];
const stateService=createDebugService({core,send:(type,m)=>stateMessages.push({type,...m}),sleep:async()=>{},busy:()=>false,paint:()=>{},applyInputs:()=>{},generation:1,knowledge:{data:pkg.knowledge,releaseId:pkg.releases[0].id}});
core.HEAPU8[core._rr_ram()+0x6e]=3;
const cycle=core._rr_cycle();await stateService.request({type:'debug-snapshot',protocol:1,generation:1,request:1});
const snap=stateMessages.at(-1).snapshot;
assert.equal(core._rr_cycle(),cycle);assert.equal(snap.state.cycle,snap.cycle);assert.equal(snap.state.entries.length,6);
const enemy=snap.state.entries.find(e=>e.id==='enemy-mode');assert.equal(enemy.node.value,'3');assert.equal(enemy.node.raw[0],3);assert.equal(enemy.node.enumLabel,pkg.knowledge.types['enemy-mode'].values['3']);
console.log('PASS M3 real WASM: Fort watches match backing bytes and CPU snapshot without execution');
