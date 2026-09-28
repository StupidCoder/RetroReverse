// Private game image validation. No media or states are written to the repository.
import {loadCore} from './wasm-harness.mjs';
import fs from 'node:fs';import assert from 'node:assert/strict';import {createHash} from 'node:crypto';
const [platform,path,statePath]=process.argv.slice(2);if(!['gb','gg'].includes(platform))throw Error('Usage: validate-handheld gb|gg ROM [native raw state]');
const c=await loadCore(platform,path),json=f=>JSON.parse(c.UTF8ToString(c[f]()));
const status=()=>json('_rr_status'),proof=()=>json('_rr_proof');
const check=n=>assert.ok(n,c.UTF8ToString(c._rr_error()));
function runFrames(n,controls=false){const end=status().frames+n;while(status().frames<end){const frame=status().frames;c._rr_pad(controls&&((frame>=180&&frame<188)||(platform==='gg'&&frame>=602&&frame<610))?128:0);check(c._rr_run(10000)>=0);}}
const frames=platform==='gb'?600:1200,start=performance.now();runFrames(frames,true);const boot=proof(),seconds=(performance.now()-start)/1000;
if(statePath){const bytes=fs.readFileSync(statePath);c.HEAPU8.set(bytes,c._rr_state_input(bytes.length));check(c._rr_state_load(bytes.length));assert.deepEqual(proof(),boot,'Native/WASM boot differs');}
const save=()=>{const n=c._rr_state_save();check(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const load=b=>{c.HEAPU8.set(b,c._rr_state_input(b.length));check(c._rr_state_load(b.length));};
const frame=()=>{const p=c._rr_frame();return c.HEAPU8.slice(p,p+160*144*4);};
const sha=b=>createHash('sha256').update(b).digest('hex');
check(c._rr_capture_begin());runFrames(1);check(c._rr_capture_end());const capture=json('_rr_capture_info');assert.equal(capture.overflow,0);const final=frame(),live=save();
const replay=json('_rr_replay_begin');const positions=[0,1,Math.floor(replay.count/3),replay.count,Math.floor(replay.count/3),replay.count],hashes=[];
for(const position of positions){while(!c._rr_replay_seek(position)){}const p=c._rr_replay_frame();hashes.push(sha(c.HEAPU8.slice(p,p+160*144*4)));}
assert.equal(hashes[2],hashes[4]);assert.equal(hashes[3],sha(final));assert.equal(hashes[5],sha(final));assert.ok(hashes.some(h=>h!==hashes[3]),'Replay never changes the display');assert.equal(sha(save()),sha(live),'Replay mutated live state');
for(const [x,y]of [[8,8],[40,72],[80,72],[120,130]]){const e=JSON.parse(c.UTF8ToString(c._rr_pixel(x,y)));assert.ok(e.complete,'Pixel reconstruction mismatch');assert.equal(e.reconstructed>>>0,new DataView(final.buffer,final.byteOffset,final.byteLength).getUint32((y*160+x)*4,true));for(const w of e.contributors.filter(w=>w.command?.hasSource)){const source=JSON.parse(c.UTF8ToString(c._rr_source(w.sourceAddress,platform==='gb'?2:4,w.sourceBefore,w.sourceValue)));assert.ok(source.complete,'Tile history mismatch');const palette=JSON.parse(c.UTF8ToString(c._rr_source(w.paletteAddress,platform==='gb'?1:2,w.paletteBefore,w.paletteValue)));assert.ok(palette.complete,'Palette history mismatch');}}
runFrames(30);const continuation=proof();load(live);runFrames(30);assert.deepEqual(proof(),continuation);
console.log(JSON.stringify({platform,frames,seconds,updatesPerSecond:frames/seconds,boot,capture,replaySteps:replay.count,stateBytes:live.length,checks:{nativeWasm:!!statePath,sourcePixelReconstruction:true,replayMatches:true,replaySeeksRepeatable:true,replayLeavesStateUntouched:true,stateContinuation:true}},null,2));
