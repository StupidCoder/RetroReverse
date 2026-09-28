import fs from 'node:fs';import assert from 'node:assert/strict';
import {loadCore} from '../../../../browser/tests/wasm-harness.mjs';
const c=await loadCore('3ds',process.argv[2]);const json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const next=()=>{const f=json('_rr_status').frames;let calls=0;while(json('_rr_status').frames===f){assert(c._rr_run(50000)>=0,c.UTF8ToString(c._rr_error()));assert(++calls<100000);}};
const save=()=>{const n=c._rr_state_save();assert(n,c.UTF8ToString(c._rr_error()));return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const restore=b=>{const p=c._rr_state_input(b.length);c.HEAPU8.set(b,p);assert(c._rr_state_load(b.length),c.UTF8ToString(c._rr_error()));};
const start=performance.now();const resumed=process.argv.includes("--resume");if(resumed)restore(fs.readFileSync("/private/tmp/3ds-wasm-before.state"));for(let i=1;i<=(resumed?0:60);i++){next();if(i%10===0)console.log(JSON.stringify({frame:i,seconds:(performance.now()-start)/1000,heap:c.HEAPU8.length,proof:json('_rr_proof')}));}
const state=save();fs.writeFileSync("/private/tmp/3ds-wasm-before.state",state);for(let i=0;i<3;i++)next();const after=json('_rr_proof');restore(state);for(let i=0;i<3;i++)next();assert.deepEqual(json('_rr_proof'),after);
for(let attempt=0;attempt<12;attempt++){assert(c._rr_capture_begin());next();assert(c._rr_capture_end());if(json('_rr_capture_info').events>10)break;}const info=json('_rr_capture_info');console.log(JSON.stringify({capture:info,heap:c.HEAPU8.length}));assert.equal(info.overflow,0);const snapshot=save(),ptr=c._rr_frame(),pixels=c.HEAPU8.slice(ptr,ptr+400*480*4),replay=json('_rr_replay_begin');
for(const at of [replay.count,0,Math.floor(replay.count/2),replay.count]){while(!c._rr_replay_seek(at)){}if(at===replay.count){const p=c._rr_replay_frame();assert.deepEqual(c.HEAPU8.slice(p,p+pixels.length),pixels);}}
assert.deepEqual(save(),snapshot,'Replay altered live state');
for(const[x,y]of [[200,120],[80,280]]){const p=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));assert.equal(p.complete,true);}
fs.writeFileSync('/private/tmp/3ds-wasm.state',snapshot);fs.writeFileSync('/private/tmp/3ds-wasm.rgba',pixels);
console.log(JSON.stringify({seconds:(performance.now()-start)/1000,stateBytes:state.length,stateContinuation:true,replayFinal:true,replayIsolated:true,capture:info,replaySteps:replay.count,profile:json('_rr_profile')}));
