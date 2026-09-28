import fs from 'node:fs';import assert from 'node:assert/strict';
import {loadCore} from '../../../../browser/tests/wasm-harness.mjs';
const [rom,oraclePath,count='1200']=process.argv.slice(2),c=await loadCore('ds',rom);
const json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const next=()=>{const frame=json('_rr_status').frames;let calls=0;while(json('_rr_status').frames===frame){assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));assert(++calls<1000);}};
const hash=b=>{let h=2166136261;for(const v of b)h=Math.imul(h^v,16777619)>>>0;return h;};
const regHash=r=>{const b=new Uint8Array(64),v=new DataView(b.buffer);r.forEach((x,i)=>v.setUint32(i*4,x,true));return hash(b);};
const oracle=oraclePath?fs.readFileSync(oraclePath,'utf8').trim().split('\n').map(JSON.parse):null;
let start=performance.now();for(let i=0;i<=Number(count);i++){if(i){if(process.argv.includes("--touch")){if(i===320||i===420)c._rr_touch(128,120,1);if(i===340||i===440)c._rr_touch(0,0,0);}next();}if(process.argv.includes("--every-frame")||i%30===0||i===Number(count)){const p=json('_rr_proof');if(oracle){const ref=oracle[i];assert.equal(p.arm9,regHash(ref.arm9),'ARM9 at '+i);assert.equal(p.arm7,regHash(ref.arm7),'ARM7 at '+i);assert.equal(p.steps,ref.steps,'scheduler at '+i);const ptr=c._rr_frame(),bytes=c.HEAPU8.subarray(ptr,ptr+256*384*4);assert.equal(hash(bytes.subarray(0,256*192*4)),ref.top,'top at '+i);assert.equal(hash(bytes.subarray(256*192*4)),ref.bottom,'bottom at '+i);}if(i%300===0)console.log(JSON.stringify({frame:i,elapsedMs:performance.now()-start,heap:c.HEAPU8.length}));}}
const save=()=>{const n=c._rr_state_save();assert(n,c.UTF8ToString(c._rr_error()));return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const restore=b=>{const p=c._rr_state_input(b.length);c.HEAPU8.set(b,p);assert(c._rr_state_load(b.length),c.UTF8ToString(c._rr_error()));};
const before=save();for(let i=0;i<5;i++)next();const after=json('_rr_proof');restore(before);for(let i=0;i<5;i++)next();assert.deepEqual(json('_rr_proof'),after);
assert(c._rr_capture_begin());next();assert(c._rr_capture_end());const info=json('_rr_capture_info');assert.equal(info.overflow,0);const snapshot=save(),ptr=c._rr_frame(),pixels=c.HEAPU8.slice(ptr,ptr+256*384*4),replay=json('_rr_replay_begin');
for(const at of [replay.count,0,Math.floor(replay.count/2),replay.count]){while(!c._rr_replay_seek(at)){}if(at===replay.count){const p=c._rr_replay_frame();assert.deepEqual(c.HEAPU8.slice(p,p+pixels.length),pixels);}}
assert.deepEqual(save(),snapshot,'Replay altered live state');
for(const [x,y]of [[0,0],[128,96],[128,288],[100,250]]){const pixel=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));assert.equal(pixel.complete,true);}
fs.writeFileSync('/private/tmp/ds-browser.state',snapshot);fs.writeFileSync('/private/tmp/ds-browser.rgba',pixels);
console.log(JSON.stringify({frames:Number(count),seconds:(performance.now()-start)/1000,stateBytes:before.length,stateContinuation:true,replayIsolated:true,replayFinal:true,capture:info,replaySteps:replay.count,profile:json('_rr_profile')}));
