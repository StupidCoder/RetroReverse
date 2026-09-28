import {loadCore} from './wasm-harness.mjs';import fs from 'node:fs';import assert from 'node:assert/strict';
const [platform,path,statePath]=process.argv.slice(2),duration=Number(process.env.SOAK_SECONDS||1800)*1000,initial=fs.readFileSync(statePath);
let c,iterations=0,captures=0,restores=0,resets=0,maxHeap=0;
const json=fn=>JSON.parse(c.UTF8ToString(c[fn]())),frame=()=>{let s=json('_rr_status');return platform==='c64'?s.frames:platform==='ps1'?s.fields:platform==='n64'?Math.floor(s.steps/750000):s.frame;};
const restore=b=>{c.HEAPU8.set(b,c._rr_state_input(b.length));assert.equal(c._rr_state_load(b.length),1);restores++;};
const save=()=>{let n=c._rr_state_save();assert(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const tick=()=>{let s=json('_rr_status');if(platform==='3do')assert(c._rr_run_slice(10000));else if(platform==='n64')c._rr_run(Math.min(10000,750000-s.steps%750000));else c._rr_run(10000,platform==='c64'?7:1,0);};
const next=()=>{let f=frame(),n=0;while(frame()===f){tick();assert(++n<50000);}};
const start=performance.now();let nextCapture=0,nextRestore=60000,nextReset=300000,lastLog=0;
c=await loadCore(platform,path);restore(initial);
while(performance.now()-start<duration){
 const now=performance.now()-start;
 if(now>=nextReset){c=null;globalThis.gc?.();c=await loadCore(platform,path);restore(initial);resets++;nextReset+=300000;}
 if(now>=nextRestore){restore(initial);nextRestore+=60000;}
 const until=performance.now()+20;do{tick();iterations++;}while(performance.now()<until);
 if(now>=nextCapture){next();c._rr_capture_begin();next();c._rr_capture_end();const b=save(),i=json('_rr_replay_begin');for(const at of [i.count,0,Math.floor(i.count/2),i.count])while(!c._rr_replay_seek(at)){};assert.deepEqual(save(),b);captures++;nextCapture+=30000;}
 maxHeap=Math.max(maxHeap,c.HEAPU8.length);
 if(now-lastLog>=60000){lastLog=now;globalThis.gc?.();console.log(JSON.stringify({platform,seconds:Math.round(now/1000),captures,restores,resets,heap:c.HEAPU8.length,rss:process.memoryUsage().rss}));}
 await new Promise(r=>setTimeout(r,20));
}
console.log(JSON.stringify({platform,complete:true,seconds:(performance.now()-start)/1000,iterations,captures,restores,resets,maxHeap,rss:process.memoryUsage().rss}));
