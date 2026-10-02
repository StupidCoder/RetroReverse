import {platforms} from '../../../site/emulators/platforms.js';
import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const [platform,path]=process.argv.slice(2),c=await loadCore(platform,path),json=fn=>JSON.parse(c.UTF8ToString(c[fn]()));
const save=()=>{let n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
if(process.env.NATIVE_STATE){let b=fs.readFileSync(process.env.NATIVE_STATE);c.HEAPU8.set(b,c._rr_state_input(b.length));assert(c._rr_state_load(b.length));}
const frame=()=>{let s=json('_rr_status');return platform==='c64'?s.frames:platform==='ps1'?s.fields:platform==='n64'?Math.floor(s.steps/750000):s.frame;};
const next=()=>{let n=frame(),calls=0;while(frame()===n){let s=json('_rr_status');if(platform==='3do')assert(c._rr_run_slice(10000));else if(platform==='n64')c._rr_run(Math.min(10000,750000-s.steps%750000));else c._rr_run(10000,platform==='c64'?7:1,0);if(++calls>50000)throw Error('No display');}};
next();c._rr_capture_begin();for(let i=0;i<Number(process.env.CAPTURE_FIELDS||(platforms[platform].captureFields??(platform==='ps1'?4:1)));i++)next();c._rr_capture_end();let state=save(),s=json('_rr_status'),w=platform==='c64'?392:s.width||320,h=platform==='c64'?272:s.height||240;
const bytes=fn=>{let p=c[fn]();return c.HEAPU8.slice(p,p+w*h*4);},final=bytes('_rr_frame'),info=json('_rr_replay_begin');assert.equal(info.complete,true);
const seek=n=>{let t=performance.now(),calls=0;while(!c._rr_replay_seek(n)){assert(++calls<1000);}let b=bytes('_rr_replay_frame');return {hash:crypto.createHash('sha256').update(b).digest('hex'),ms:performance.now()-t,b};};
assert.deepEqual(seek(info.count).b,final,'Replay final differs');let seen=new Map(),maxMs=0;for(let n of [0,1,Math.floor(info.count*.3),Math.floor(info.count*.7),info.count,Math.floor(info.count*.3),0,info.count]){let r=seek(n);if(seen.has(n))assert.equal(r.hash,seen.get(n),'Seek is not deterministic');seen.set(n,r.hash);maxMs=Math.max(maxMs,r.ms);}
// A static boot screen is valid; opt in with an animated in-game fixture.
if(process.env.EXPECT_REPLAY_CHANGES){
 const sampled=Array.from({length:21},(_,i)=>seek(Math.floor(info.count*i/20)).hash);
 assert(new Set(sampled).size>=3,'Replay never shows distinct intermediate rendering stages');
}
assert.deepEqual(save(),state,'Replay changed live state');
next();let continuation=save();c.HEAPU8.set(state,c._rr_state_input(state.length));assert(c._rr_state_load(state.length));next();assert.deepEqual(save(),continuation,'Continuation changed');
console.log(JSON.stringify({platform,steps:info.count,maxSeekMs:maxMs,endMatches:true,arbitrarySeek:true,liveUnchanged:true,continuationMatches:true}));
