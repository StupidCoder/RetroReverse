// Integration acceptance using a user-supplied local disc, never committed media.
import {loadCore} from './wasm-harness.mjs';import assert from 'node:assert/strict';
const c=await loadCore('3do',process.argv[2]),json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const next=()=>{const frame=json('_rr_status').frame;let calls=0;while(json('_rr_status').frame===frame){assert(c._rr_run_slice(10000),c.UTF8ToString(c._rr_error()));assert(++calls<50000);}};
const save=()=>{const n=c._rr_state_save();assert(n,c.UTF8ToString(c._rr_error()));return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const restore=b=>{c.HEAPU8.set(b,c._rr_state_input(b.length));assert(c._rr_state_load(b.length),c.UTF8ToString(c._rr_error()));};
const frame=fn=>{const p=c[fn]();return c.HEAPU8.slice(p,p+320*240*4);};
let movieFrames=0,previousTime=0,tested=false,samples=[];
while(json('_rr_status').frame<600){
 next();const s=json('_rr_status');assert(s.seconds>=previousTime);previousTime=s.seconds;if(s.movieHLE)movieFrames++;
 if(s.frame%100===0)samples.push(s);
 if(movieFrames===10&&!tested){
  const b=save();assert(b.length<32*1024*1024);next();const expected=save();restore(b);next();assert.deepEqual(save(),expected);
  c._rr_capture_begin();next();c._rr_capture_end();const before=save(),final=frame('_rr_frame'),info=json('_rr_replay_begin');
  assert(info.complete);assert(info.count>0);while(!c._rr_replay_seek(info.count)){}assert.deepEqual(frame('_rr_replay_frame'),final);assert.deepEqual(save(),before);
  const pixel=JSON.parse(c.UTF8ToString(c._rr_pixel(160,120)));assert(pixel.complete);assert(pixel.contributors.some(x=>x.command.kind==='Cinepak movie HLE'));
  tested=true;
 }
}
assert(tested&&movieFrames>400);
console.log(JSON.stringify({movieFrames,saveRestore:true,replayEndMatches:true,hleProvenance:true,samples}));
