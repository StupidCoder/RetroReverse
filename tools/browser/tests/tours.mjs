import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createTourService,matches,intervalEvidence} from '../../../site/emulators/tour-worker.js';
import factory from '../../../site/emulators/cores/c64/core.js';
const core=await factory({wasmBinary:fs.readFileSync(new URL('../../../site/emulators/cores/c64/core.wasm',import.meta.url))});
const put=b=>core.HEAPU8.set(b,core._rr_input());put(new Uint8Array(20480));assert(core._rr_init(8192,8192,4096));
const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;put(tape);assert(core._rr_tape(21));
// A genuine write then restore in one interval, deliberately no endpoint diff.
const ram=new Uint8Array(65536).fill(0xea),program=[0xa9,1,0x8d,0,2,0xa9,0,0x8d,0,2,0x4c,0,8];ram.set(program,0x800);ram[0x200]=0;
function reset(){put(ram);assert(core._rr_prepare(0x800,0));}
const snapshot=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
function stop(id,extra={}){return {id,title:id,explanation:'verified',until:{pc:0x800},assertions:[{address:0x200,value:0}],cycleBudget:1000,wallBudget:1000,hitCount:1,capture:[{address:0x200,length:1}],eventBudget:10000,layout:{address:0x800,gameState:true},...extra};}
const tour={title:'test',releases:['test'],guards:[{address:0x800,bytes:program}],start:{pc:0x800},requirements:'Start at $0800',stops:[stop('entry'),stop('roundtrip')]};
let response,messages=[],onYield=()=>{},busy=false,identity='matched',request=0;
let service,initialCycle;
function setup(t=tour){reset();initialCycle=core._rr_cycle();request=0;messages=[];service=createTourService({core,knowledge:{status:identity,releaseId:'test',data:{tours:{test:t}}},generation:7,send:(type,m)=>{response={type,...m};messages.push(response);},sleep:async()=>await onYield(),busy:()=>busy,paint:()=>{},snapshot});}
async function command(type,args={}){await service.request({type,request:++request,protocol:1,generation:7,id:'test',...args});return response;}
setup();await command('tour-start');assert.equal(response.phase,'paused-at-stop');assert.equal(response.evidence.cycles,0);
await command('tour-continue');assert.equal(response.phase,'completed');assert.equal(response.evidence.writes,2);assert.equal(response.evidence.changed,0);assert.equal(response.evidence.changes[0].writes,2);assert.equal(response.evidence.complete,true);const cycle=core._rr_cycle();
service.invalidate();assert.equal(response.phase,'diverged');core._rr_run(2);await command('tour-restore');assert.equal(response.phase,'completed');assert.equal(core._rr_cycle(),cycle);
setup();await command('tour-start');core.HEAPU8[core._rr_ram()+0x200]=9;await command('tour-continue');assert.equal(response.phase,'failed');assert.match(response.text,/diverged/);await command('tour-restore');assert.equal(response.phase,'paused-at-stop');assert.equal(core.HEAPU8[core._rr_ram()+0x200],0);await command('tour-continue');assert.equal(response.phase,'completed');
setup();await command('tour-start');onYield=()=>service.cancel();await command('tour-continue');onYield=()=>{};assert.equal(response.phase,'cancelled');assert(response.canRestore);await command('tour-restore');assert.equal(response.phase,'paused-at-stop');
setup({...tour,stops:[stop('entry'),stop('missed',{until:{pc:0x900},cycleBudget:24})]});await command('tour-start');await command('tour-continue');assert.equal(response.text,'budget');assert.equal(response.evidence.cycles,24);
setup({...tour,stops:[stop('entry'),stop('overflow',{until:{pc:0x900},eventBudget:1})]});await command('tour-start');await command('tour-continue');assert.equal(response.text,'trace-overflow');assert.equal(response.evidence.complete,false);
setup({...tour,stops:[stop('entry'),stop('wrong',{assertions:[{address:0x200,value:2}]})]});await command('tour-start');await command('tour-continue');assert.equal(response.phase,'failed');assert.match(response.text,/assertion/);
setup({...tour,stops:[stop('entry'),stop('twice',{hitCount:2})]});await command('tour-start');await command('tour-continue');assert.equal(response.evidence.writes,4);
setup();await command('tour-start',{generation:6});assert.equal(response.type,'tour-rejected');assert.equal(core._rr_cycle(),initialCycle);await command('tour-start');await command('tour-continue',{request:1});assert.equal(response.type,'tour-rejected');assert.equal(core._rr_cycle(),initialCycle);
setup();busy=true;await command('tour-start');assert.equal(response.type,'tour-rejected');busy=false;
identity='unknown';setup();await command('tour-start');assert.equal(response.phase,'failed');assert.match(response.text,/exact/);identity='matched';
setup();core.HEAPU8[core._rr_ram()+0x800]=0xea;await command('tour-start');assert.equal(response.phase,'failed');assert.match(response.text,/signature/);
setup({...tour,stops:[stop('entry'),stop('wall',{until:{pc:0x900},wallBudget:1,cycleBudget:100000})]});await command('tour-start');onYield=()=>new Promise(r=>setTimeout(r,5));await command('tour-continue');onYield=()=>{};assert.equal(response.text,'budget');
setup({...tour,stops:[stop('entry'),stop('byte',{until:{any:[{address:0x200,value:1},{pc:0x900}]},assertions:[{address:0x200,value:1}]})]});await command('tour-start');await command('tour-continue');assert.equal(response.phase,'completed');assert.equal(response.evidence.writes,1);
const state={boundary:true,nextPC:10,mapping:[0]},bytes=new Uint8Array(256);bytes[5]=0xa1;
assert(matches({all:[{pc:10},{any:[{address:5,value:0xa0,mask:0xf0},{pc:4}]}]},state,bytes));assert(!matches({pc:10},{...state,interruptPending:true},bytes));
const evidence=intervalEvidence([{address:5,length:1}],[Uint8Array.of(0xa1)],bytes,[{region:0,kind:2,offset:5}],1);assert.equal(evidence.complete,false);assert.equal(evidence.changed,0);assert.equal(evidence.writes,1);
console.log('PASS tours: real WASM stops, write/restore evidence, restore, divergence, cancellation, budgets, overflow, assertions, hit counts, stale requests, ownership, identity and signatures');
