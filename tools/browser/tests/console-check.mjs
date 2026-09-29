// Private media regression: native-state import, deterministic restore, bounded
// malformed-state handling, full rendering replay and paused-machine isolation.
import fs from 'node:fs';import assert from 'node:assert/strict';import {loadCore} from './wasm-harness.mjs';
const [platform,image,state]=process.argv.slice(2),c=await loadCore(platform,image);
const json=name=>JSON.parse(c.UTF8ToString(c['_rr_'+name]()));
function load(b){let p=c._rr_state_input(b.length);assert(p);c.HEAPU8.set(b,p);return c._rr_state_load(b.length);}
function save(){let n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);}
function run(n){const end=json('status').frames+n;while(json('status').frames<end)assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));}
function frame(replay=false){let p=replay?c._rr_replay_frame():c._rr_frame(),s=json('status');return c.HEAPU8.slice(p,p+s.width*s.height*4);}
function seek(n){while(!c._rr_replay_seek(n));return frame(true);}
assert(load(fs.readFileSync(state)));const original=json('proof'),checkpoint=save();
assert.equal(load(checkpoint.subarray(0,checkpoint.length-1)),0);assert.deepEqual(json('proof'),original);
run(3);const advanced=json('proof');assert(load(checkpoint));run(3);assert.deepEqual(json('proof'),advanced,'restore deterministic');
assert(load(checkpoint));assert(c._rr_capture_begin());run(platform==='gba'?1:(platform==='gc'||platform==='dc')?3:4);assert(c._rr_capture_end());const proof=json('proof'),expected=frame(),info=json('capture_info'),r=json('replay_begin');
console.log(JSON.stringify({platform,info,replay:r,proof}));assert(info.events>0);assert(info.writes>0);assert.equal(info.overflow,0,'capture overflow');
assert.deepEqual(seek(r.count),expected,'last step equals live frame');const first=seek(0);let changed=!first.every((v,i)=>v===expected[i]);for(const fraction of [.2,.5,.8]){const p=seek(Math.floor(r.count*fraction));changed||=!p.every((v,i)=>v===expected[i]);}assert(changed,'scrubbing must change pixels');assert.deepEqual(seek(r.count),expected);
const s=json('status');let explained=0,complete=0;for(let y=16;y<s.height;y+=37)for(let x=16;x<s.width;x+=41){const p=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));if(p.contributors?.length)explained++;if(p.complete)complete++;}assert(explained>0,'pixel has actual contributors');assert(complete>0);assert.deepEqual(json('proof'),proof,'replay must not mutate guest');
assert(load(checkpoint));run(platform==='gba'?1:(platform==='gc'||platform==='dc')?3:4);assert.deepEqual(json('proof'),proof,'capture must not alter execution');console.log(JSON.stringify({pass:true,platform,explained,complete,stateBytes:checkpoint.length}));
