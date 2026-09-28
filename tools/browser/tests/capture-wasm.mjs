import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';import assert from 'node:assert/strict';
const [platform,path]=process.argv.slice(2),core=await loadCore(platform,path);
const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
const frameNum=()=>{const s=status();return platform==='c64'?s.frames:platform==='ps1'?s.fields:platform==='n64'?Math.floor(s.steps/750000):s.frame;};
const tick=()=>{if(platform==='3do')assert.ok(core._rr_run_slice(10000));else if(platform==='n64')assert.ok(core._rr_run(Math.min(10000,750000-status().steps%750000))>=0);else assert.ok(core._rr_run(10000,platform==='c64'?7:1,0)>=0);};
const next=()=>{let first=frameNum(),calls=0;do{tick();if(++calls>50000)throw Error('No display boundary');}while(frameNum()===first);};
const save=()=>{const n=core._rr_state_save();assert.ok(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
const restore=b=>{core.HEAPU8.set(b,core._rr_state_input(b.length));assert.equal(core._rr_state_load(b.length),1);};
if(process.env.NATIVE_STATE)restore(new Uint8Array(fs.readFileSync(process.env.NATIVE_STATE)));
for(let point=0;point<3;point++){
 next();const start=save();next();const end=save();restore(start);core._rr_capture_begin();next();core._rr_capture_end();assert.deepEqual(save(),end,'Tracing changed guest continuation');
 const info=JSON.parse(core.UTF8ToString(core._rr_capture_info()));let complete=0,missing=0;
 const dims=status();for(const [x,y] of [[0,0],[160,120],[100,100],[20,40],[Math.min(319,(dims.width||392)-1),Math.min(239,(dims.height||272)-1)]]){const p=JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));if(p.error||p.complete===false){missing++;console.error(JSON.stringify({platform,x,y,error:p.error,final:p.final,reconstructed:p.reconstructed,address:p.address}));}else complete++;}
 console.log(JSON.stringify({platform,point,frame:frameNum(),info,sampledPixels:{complete,missing}}));
}
