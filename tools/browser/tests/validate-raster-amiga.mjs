// Optional private-media test. Only metrics and writer addresses are printed.
import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
const [adf,checkpoint,nativeEnd]=process.argv.slice(2),c=await loadCore('amiga',adf);
const call=(name,...args)=>JSON.parse(c.UTF8ToString(c['_rr_'+name](...args)));
const save=()=>{const n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const restore=b=>{c.HEAPU8.set(b,c._rr_state_input(b.length));assert(c._rr_state_load(b.length),c.UTF8ToString(c._rr_error()));};
const run=n=>{const end=call('status').frames+n;while(call('status').frames<end)assert(c._rr_run(10000)>0);};
restore(fs.readFileSync(checkpoint));const initial=save();
const began=performance.now();run(120);const elapsed=performance.now()-began;
if(nativeEnd)assert.deepEqual(Buffer.from(save()),fs.readFileSync(nativeEnd));
restore(initial);run(3);const expected=save();restore(initial);
const captureStart=performance.now();assert(c._rr_capture_begin());run(3);assert(c._rr_capture_end());const captureMs=performance.now()-captureStart;
assert.deepEqual(save(),expected,'Tracing changed machine state');
const capture=call('capture_info'),info=call('raster_info'),proof=call('proof');assert.equal(capture.overflow,0);assert(info.complete);assert.equal(info.lines.length,256);
const rgba=p=>c.HEAPU8.slice(p,p+640*256*4),frame=rgba(c._rr_frame());
let pixels=0,sources=0,blitWords=0;const producerBlits=new Set(),copperAddresses=new Set(),seeks=[];
const checkSource=s=>{const h=call('source',s.address,s.size,s.before,s.value);assert(h.complete,JSON.stringify({source:s,history:h}));sources++;for(const w of h.contributors){if(w.command?.origin==='blitter')producerBlits.add(w.command.blit);if(w.command?.origin==='copper')copperAddresses.add(w.command.copperPC);}};
for(let y=0;y<256;y++){
 const begin=performance.now(),line=call('raster_seek',y);seeks.push(performance.now()-begin);assert(!line.error);
 if(y%17===0||y===255)for(let x=7;x<640;x+=79){const p=call('raster_pixel',2,x,y);assert(p.complete,JSON.stringify(p));pixels++;for(const candidate of p.candidates)for(const s of candidate.sources)checkSource(s);}
 if(y<255)assert(rgba(c._rr_raster_frame(2)).subarray((y+1)*640*4).every(v=>v===0));
}
assert.deepEqual(rgba(c._rr_raster_frame(2)),frame);
// Both frozen source layers must use the historical memory at the scrub line.
call('raster_seek',128);for(const panel of [0,1])for(const [x,y]of [[10,25],[319,127],[527,203]])for(const candidate of call('raster_pixel',panel,x,y).candidates)for(const s of candidate.sources)checkSource(s);
const chosen=[...new Set([0,info.blits.length-1,...info.blits.filter(b=>b.kind==='Cookie cut').slice(0,10).map(b=>b.index),...producerBlits].filter(i=>i>=0&&i<info.blits.length))].slice(0,30);
const blitTimings=[];let changedPlayfield=0;
for(const i of chosen){const began=performance.now(),b=call('blit_seek',i);blitTimings.push(performance.now()-began);assert(b.complete);if(!Buffer.from(rgba(c._rr_blit_frame(4))).equals(Buffer.from(rgba(c._rr_blit_frame(5)))))changedPlayfield++;
 for(const [x,y]of [[0,0],[Math.floor(b.width/2),Math.floor(b.height/2)],[b.width-1,b.height-1]]){const p=call('blit_pixel',x,y);assert(!p.error,JSON.stringify({b,p}));blitWords++;if(!(b.con1&0x19)){let d=0;for(let bit=0;bit<16;bit++){const input=(p.a>>bit&1)*4+(p.b>>bit&1)*2+(p.c>>bit&1);d|=((b.con0>>input)&1)<<bit;}assert.equal(p.d,d);}for(const s of p.sources)checkSource(s);}
}
assert.deepEqual(save(),expected,'Inspection mutated paused state');assert.deepEqual(call('proof'),proof);
const replay=call('replay_begin');while(!c._rr_replay_seek(replay.count)){}assert.deepEqual(rgba(c._rr_replay_frame()),frame);
seeks.sort((a,b)=>a-b);blitTimings.sort((a,b)=>a-b);
console.log(JSON.stringify({game:path.basename(adf),pass:true,nativeParity:!!nativeEnd,coreFPS:120000/elapsed,captureMs,capture,lines:info.lines.length,blits:info.blits.length,cookieCuts:info.blits.filter(b=>b.kind==='Cookie cut').length,pixels,sources,blitWords,producerBlits:[...producerBlits],copperAddresses:[...copperAddresses],changedPlayfield,medianLineSeekMs:seeks[128],medianBlitSeekMs:blitTimings[Math.floor(blitTimings.length/2)],heapMiB:c.HEAPU8.length/1048576,proof,checks:{captureUnchanged:true,outputPixelsMatch:true,historicalSources:true,blitterTruthTables:true,inspectionUnchanged:true,replayMatches:true}},null,2));
