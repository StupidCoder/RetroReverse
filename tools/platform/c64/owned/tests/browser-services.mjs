import assert from 'node:assert/strict';
import {createC64TestCore} from '../../../../browser/tests/c64-test-core.mjs';
import {createDebugService} from '../../../../../site/emulators/debug-worker.js';
import {createMemoryService} from '../../../../../site/emulators/memory-worker.js';
const core=await createC64TestCore();
const put=b=>core.HEAPU8.set(b,core._rr_input()),json=name=>JSON.parse(core.UTF8ToString(core[name]()));
const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=tape[20]=1;
function setup(program=[0xee,0,2,0x4c,0,8]){
 put(new Uint8Array(20480));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));
 const ram=new Uint8Array(65536).fill(0xea);ram[0x200]=0;ram.set(program,0x800);put(ram);assert(core._rr_prepare(0x800,0));
}
setup();assert.equal(json('_rr_capabilities').renderCapture,true);assert.equal(typeof core._rr_capture_begin,'function');
const messages=[];let onYield=()=>{},busy=false;
const debug=createDebugService({core,send:(type,m)=>messages.push({type,...m}),sleep:async()=>onYield(),busy:()=>busy,paint:()=>{},applyInputs:()=>{},knowledge:{status:'unknown'},generation:9});
let request=0;
async function command(type,extra={}){
 const s=debug.snapshot();await debug.request({type:'debug-'+type,protocol:1,generation:9,request:++request,snapshotId:s.snapshotId,cycle:s.cycle,bank:s.bank,...extra});return messages.at(-1);
}
let result=await command('step');assert.equal(result.reason,'instruction');assert.equal(result.retired,1);assert.equal(result.snapshot.nextPC,0x803);assert.equal(result.cycles,6);
const before=core._rr_cycle();debug.snapshot(0xdc00);assert.equal(core._rr_cycle(),before);
await command('until',{target:0x800});onYield=()=>debug.request({type:'debug-cancel',protocol:1,generation:9,request:++request});
result=await command('until',{target:0x900,cycleBudget:10000});onYield=()=>{};assert.equal(result.reason,'cancelled');assert.equal(debug.active(),false);assert(result.cycles>0&&result.cycles<10000);
if(!result.snapshot.boundary){const clocks=core._rr_cycle();result=await command('step');assert.equal(result.reason,'rejected');assert.equal(core._rr_cycle(),clocks);result=await command('normalize');assert.equal(result.reason,'boundary');}
const stale=debug.snapshot();debug.snapshot();await debug.request({type:'debug-step',protocol:1,generation:9,request:++request,snapshotId:stale.snapshotId,cycle:stale.cycle,bank:stale.bank});assert.equal(messages.at(-1).reason,'rejected');
busy=true;result=await command('step');assert.equal(result.reason,'rejected');busy=false;
// Every inspector owns a snapshot ID; inspecting a second panel does not
// silently invalidate the first panel's displayed execution context.
const first=debug.snapshot(-1,'first');debug.snapshot(-1,'second');await debug.request({type:'debug-step',client:'first',protocol:1,generation:9,request:++request,snapshotId:first.snapshotId,cycle:first.cycle,bank:first.bank});assert.equal(messages.at(-1).reason,'instruction');
setup([0xa9,7,0x85,1,0xa9,0x42,0x8d,0,2,0xad,0,2,0x4c,0,8]);core._rr_play(1);
const file=new File([tape],'authored.tap'),memory=createMemoryService({core,platform:'c64',files:[file],status:()=>json('_rr_status')});
await memory.begin(true);assert.equal(core._rr_run(100,0,0),100);const end=memory.finish();assert(end.count>0);assert.equal(end.tapeCursor,1);assert.equal(memory.page('ram',0x200).bytes[0],0x42);
const clocks=core._rr_cycle(),events=end.count;assert(memory.detail('ram',0x200).events.some(e=>(e.kind&2)&&e.value===0x42));
memory.seek(0);assert.equal(memory.page('ram',0x200).bytes[0],0);memory.seek(events);assert.equal(memory.page('ram',0x200).bytes[0],0x42);assert.equal(core._rr_cycle(),clocks);
assert.equal(memory.pulses(0,16).durations[0],8);await memory.liveSnapshot();core._rr_run(100,0,0);const live=await memory.liveSnapshot();assert(live.count>0);memory.stopLive();
// A paused framebuffer is inspectable without invoking device reads or clocks.
const paused=core._rr_cycle(),p=core._rr_frame(),pixels=core.HEAPU8.slice(p,p+392*272*4);assert.equal(core._rr_cycle(),paused);assert.equal(pixels.length,392*272*4);assert.equal(pixels[3],255);
console.log('PASS owned browser services: debugger steps/cancellation/stale contexts/independent panels, RAM and tape activity, scrubbing/live memory and nonmutating display');
