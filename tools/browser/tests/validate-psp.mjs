import {loadCore} from './wasm-harness.mjs';
import {decodedColor} from '../../../site/emulators/inspector.js';
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const c=await loadCore('psp',process.argv[2]),j=f=>JSON.parse(c.UTF8ToString(c[f]()));
const save=()=>{let n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const load=b=>{let p=c._rr_state_input(b.length);assert(p);c.HEAPU8.set(b,p);assert(c._rr_state_load(b.length));};
load(fs.readFileSync(process.argv[3]));console.log('loaded',j('_rr_proof'));
const next=()=>{let frame=j('_rr_status').frames,calls=0;while(j('_rr_status').frames===frame){assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));assert(++calls<100000);}};
const pixels=fn=>{let p=c[fn]();return c.HEAPU8.slice(p,p+480*272*4);};
const pre=save();let t=performance.now();for(let i=0;i<60;i++)next();console.log('benchmark',{seconds:(performance.now()-t)/1000},j('_rr_proof'),j('_rr_profile'));
load(pre);assert(c._rr_capture_begin());next();assert(c._rr_capture_end());const capture=j('_rr_capture_info');console.log('capture',capture);assert.equal(capture.overflow,0);let state=save(),final=pixels('_rr_frame');for(const [x,y] of [[0,0],[160,120],[240,136],[479,271]]){const ev=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));assert(ev.complete,'Stored bytes do not reconstruct');assert.deepEqual(decodedColor('psp',ev.reconstructed,ev.size,ev.displayFormat),Array.from(final.slice((y*480+x)*4,(y*480+x)*4+4)));}
let replay=j('_rr_replay_begin');let hashes=new Map();for(let n of [0,1,Math.floor(replay.count*.3),Math.floor(replay.count*.7),replay.count,Math.floor(replay.count*.3),0,replay.count]){while(!c._rr_replay_seek(n)){}let b=pixels('_rr_replay_frame'),h=crypto.createHash('sha256').update(b).digest('hex');if(hashes.has(n))assert.equal(hashes.get(n),h);hashes.set(n,h);if(n===replay.count)assert.deepEqual(b,final);}
assert(hashes.size>=3);assert(new Set(hashes.values()).size>=2,'Replay never changes');assert.deepEqual(save(),state);next();const continued=save();load(state);next();assert.deepEqual(save(),continued);console.log('PASS',replay,[...hashes]);
