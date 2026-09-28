import fs from 'node:fs';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {loadCore} from '../../../../browser/tests/wasm-harness.mjs';
const [rom,statePath]=process.argv.slice(2),c=await loadCore('ds',rom);
const json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const saved=fs.readFileSync(statePath);c.HEAPU8.set(saved,c._rr_state_input(saved.length));assert(c._rr_state_load(saved.length));
const next=()=>{const f=json('_rr_status').frames;while(json('_rr_status').frames===f)assert(c._rr_run(10000)>=0);};
const save=()=>{const n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const pixels=fn=>{const p=c[fn]();return c.HEAPU8.slice(p,p+256*384*4);};
const hash=b=>createHash('sha256').update(b).digest('hex');
next();assert(c._rr_capture_begin());next();next();assert(c._rr_capture_end());
const capture=json('_rr_capture_info'),state=save(),final=pixels('_rr_frame'),replay=json('_rr_replay_begin');assert.equal(capture.overflow,0);
const seek=n=>{while(!c._rr_replay_seek(n)){}return {info:json('_rr_replay_info'),bytes:pixels('_rr_replay_frame')};};
assert.deepEqual(seek(replay.count).bytes,final);
const samples=[],targets=new Set();
for(let i=0;i<=40;i++){
 const step=Math.floor(replay.count*i/40),r=seek(step),sha=hash(r.bytes);
 samples.push({step,sha,surface:r.info.surface});if(r.info.surface.startsWith('3D'))targets.add(sha);
}
assert(targets.size>1,'Scrubbing 3D commands must visibly change the active render target');
for(const s of samples.reverse())assert.equal(hash(seek(s.step).bytes),s.sha,'Backward seek changed historical output');
assert.deepEqual(seek(replay.count).bytes,final);
assert.deepEqual(save(),state,'Replay changed the live machine');
for(const [x,y] of [[40,60],[128,96],[128,288]]){
 const p=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));assert(p.complete);assert.equal(typeof p.final,'number');assert.equal(p.final,p.reconstructed);
}
console.log(JSON.stringify({capture,replaySteps:replay.count,distinct3DTargets:targets.size,finalMatches:true,backwardSeeks:true,liveUnchanged:true}));
