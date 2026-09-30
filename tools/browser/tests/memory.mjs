import assert from 'node:assert/strict';
import {applyEvents,pixelRange,parseAddress,tapIndex,summarize} from '../../../site/emulators/memory-model.js';
import {createMemoryService} from '../../../site/emulators/memory-worker.js';
const r={id:'rom-7',size:16384,base:7*16384,aliases:[16384]};
assert.equal(parseAddress('rom-7:3fff',[r],r).offset,16383);
assert.equal(parseAddress('$4001',[r],r).offset,1);
assert.equal(parseAddress('rom-7:4000',[r],r),null);
assert.equal(pixelRange(32,8,256),null);
assert.deepEqual(pixelRange(31,8,251),{start:248,end:251});
assert.deepEqual([...summarize(Uint8Array.of(0,100,200),2)],[50,200]);
const bytes=new Uint8Array(28);bytes[12]=1;bytes.set([2,0,0x34,0x12,0,3,4,5],20);
const tape=tapIndex(bytes);assert.deepEqual([...tape.offsets],[20,21,25,26,27]);assert.equal(tape.durations[1],0x1234);
const regions=[{bytes:Uint8Array.of(1,2,3,4)}],events=Uint32Array.of(1,0,0,1,0xbeef,2,2,0x8000,1,2,0,0,0,255,1,1,0x8001,0);
applyEvents(regions,events,0,2);assert.deepEqual([...regions[0].bytes],[1,239,190,4]);
// Service reconstruction never mutates the live machine and backward seek
// starts from the same immutable snapshot, including same-value writes.
const heap=new Uint8Array(256);heap.set([1,2,3,4],8);heap.set(new Uint8Array(events.buffer),32);
const core={HEAPU8:heap,UTF8ToString:x=>x,_rr_inspect_regions:()=>JSON.stringify({activity:true,regions:[{id:'ram',name:'RAM',index:0,size:4,base:0,aliases:[0]}]}),_rr_inspect_data:()=>8,_rr_activity_begin:()=>{},_rr_activity_end:()=>{},_rr_activity_count:()=>2,_rr_activity_data:()=>32,_rr_activity_dropped:()=>0};
const service=createMemoryService({core,platform:'test',files:[],status:()=>({steps:0})});
await service.begin();heap[9]=239;heap[10]=190;const end=service.finish();assert.equal(end.position,2);assert.deepEqual([...service.page('ram',0).bytes],[1,239,190,4]);
service.seek(0);assert.deepEqual([...service.page('ram',0).bytes],[1,2,3,4]);assert.equal(heap[9],239);service.seek(1);assert.equal(service.page('ram',0).bytes[1],239);service.seek(2);assert.equal(service.detail('ram',1).events[0].kind,2);
await service.liveSnapshot();heap[8]=77;const live=await service.liveSnapshot();assert.equal(live.live,true);assert.equal(live.recording,0);assert.equal(service.page('ram',0).bytes[0],77);assert.equal(live.regions[0].activityMap[1]&2,2);service.stopLive();assert.equal(service.overview().live,false);
console.log('PASS memory: address edges, aggregation, TAP indexing, forward/backward reconstruction, read-only inspection');

// Large RAM regions must remain drawable even in detailed mode.
const largeHeap=new Uint8Array(2*1024*1024+16);largeHeap[largeHeap.length-1]=123;
const largeCore={...core,HEAPU8:largeHeap,_rr_inspect_regions:()=>JSON.stringify({activity:false,regions:[{id:'ram',name:'Large RAM',kind:'ram',index:0,size:largeHeap.length-8,base:0,aliases:[0]}]}),_rr_inspect_data:()=>8};
const largeService=createMemoryService({core:largeCore,platform:'test',files:[],status:()=>({})});await largeService.snapshot();const detailed=largeService.overview(1).regions[0];assert(detailed.bitmap.length<=1048576);assert.equal(detailed.scale,4);const last=largeService.page('ram',largeHeap.length-9);assert.equal(last.bytes.at(-1),123);
console.log('PASS large-memory atlas: bounded bitmap and final-byte inspection');
