import fs from 'node:fs';
import {Session} from 'node:inspector/promises';
import assert from 'node:assert/strict';
import {loadCore} from './wasm-harness.mjs';
const [platform,image,state,frames='60']=process.argv.slice(2);
const core=await loadCore(platform,image),json=name=>JSON.parse(core.UTF8ToString(core['_rr_'+name]()));
if(state!=='-'){const b=fs.readFileSync(state);const p=core._rr_state_input(b.length);assert(p);core.HEAPU8.set(b,p);assert(core._rr_state_load(b.length));}
const session=new Session();if(process.env.RR_CPU_PROFILE){session.connect();await session.post('Profiler.enable');await session.post('Profiler.setSamplingInterval',{interval:500});await session.post('Profiler.start');}
const start=json('status'),end=start.frames+Number(frames),t=performance.now();
while(json('status').frames<end){const s=json('status');if(process.env.RR_AUTO)core._rr_pad(s.frames%120<10?(platform==='gba'?9:platform==='dc'?12:platform==='gc'?0x1100:0x4008):0,0,0);assert(core._rr_run(10000)>=0,core.UTF8ToString(core._rr_error()));}
const ms=performance.now()-t;
if(process.env.RR_CPU_PROFILE){const {profile}=await session.post('Profiler.stop');fs.writeFileSync(process.env.RR_CPU_PROFILE,JSON.stringify(profile));session.disconnect();}
console.log(JSON.stringify({platform,fields:Number(frames),ms,fps:Number(frames)*1000/ms,heapBytes:core.HEAPU8.length,proof:json('proof'),profile:json('profile')}));
if(process.env.RR_SAVE){const n=core._rr_state_save();assert(n);fs.writeFileSync(process.env.RR_SAVE,core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n));}
if(process.env.RR_PPM){const p=core._rr_frame(),s=json('status'),b=core.HEAPU8.slice(p,p+s.width*s.height*4);const out=Buffer.alloc(s.width*s.height*3);for(let i=0;i<s.width*s.height;i++){out[i*3]=b[i*4];out[i*3+1]=b[i*4+1];out[i*3+2]=b[i*4+2];}fs.writeFileSync(process.env.RR_PPM,Buffer.concat([Buffer.from(`P6\n${s.width} ${s.height}\n255\n`),out]));}
