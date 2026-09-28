import {decodedColor} from '../../../site/emulators/inspector.js';
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
 const info=JSON.parse(core.UTF8ToString(core._rr_capture_info()));let complete=0,missing=0,sourceChecks=0;const dims0=status(),width=platform==='c64'?392:dims0.width||320,height=platform==='c64'?272:dims0.height||240,fp=core._rr_frame(),pixels=core.HEAPU8.slice(fp,fp+width*height*4);
 const grid=process.env.PIXEL_GRID?Array.from({length:80},(_,i)=>[Math.floor((i%10+.5)*width/10),Math.floor((Math.floor(i/10)+.5)*height/8)]):[];
 const dims=status();for(const [x,y] of [...grid,[0,0],[160,120],[100,100],[20,40],[Math.min(319,(dims.width||392)-1),Math.min(239,(dims.height||272)-1)]]){const p=JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));if(p.error||p.complete===false){missing++;console.error(JSON.stringify({platform,x,y,error:p.error,final:p.final,reconstructed:p.reconstructed,address:p.address}));}else {complete++;if(!p.blank){const color=platform==='c64'?[p.rgba&255,p.rgba>>>8&255,p.rgba>>>16&255,255]:decodedColor(platform,p.reconstructed,p.size);assert.deepEqual(color,[...pixels.slice((y*width+x)*4,(y*width+x)*4+4)],'Modeled color mismatch');}
 if(platform==='ps1'||platform==='3do')for(const c of p.contributors||[]){if(!c.drawn||!(c.command?.hasSource||c.command?.lrform))continue;if(platform==='3do'&&(c.sourceAddress<0x200000||c.sourceAddress>=0x300000))continue;const q=JSON.parse(core.UTF8ToString(core._rr_source(c.sourceAddress,2,c.sourceBefore,c.sourceValue)));assert.equal(q.complete,true,'Historical source mismatch '+JSON.stringify(c));sourceChecks++;}
 }}
 console.log(JSON.stringify({platform,point,frame:frameNum(),info,sampledPixels:{complete,missing},sourceChecks}));next();
}
