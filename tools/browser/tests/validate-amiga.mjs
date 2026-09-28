// Optional private-media validation: no game bytes or states are written into the repository.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {loadCore} from './wasm-harness.mjs';
const [adf,statePath,nativeEnd]=process.argv.slice(2),core=await loadCore('amiga',adf);
const json=n=>JSON.parse(core.UTF8ToString(core[n]()));
const restore=b=>{core.HEAPU8.set(b,core._rr_state_input(b.length));assert(core._rr_state_load(b.length),core.UTF8ToString(core._rr_error()));};
const save=()=>{const n=core._rr_state_save();assert(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
const next=()=>{let f=json('_rr_status').frames;while(json('_rr_status').frames===f)assert(core._rr_run(10000)>0);};
const initial=fs.readFileSync(statePath);restore(initial);
const begin=performance.now();for(let i=0;i<120;i++)next();const ms=performance.now()-begin,continuation=save();
if(nativeEnd)assert.deepEqual(Buffer.from(continuation),fs.readFileSync(nativeEnd),'Native / WASM state bytes differ');
restore(initial);const start=save();next();const expected=save();restore(start);assert(core._rr_capture_begin());next();assert(core._rr_capture_end());assert.deepEqual(save(),expected,'Capture changed continuation');
const info=json('_rr_capture_info');assert.equal(info.overflow,0);let sources=0,pixels=0,sprites=0;
for(let y=8;y<256;y+=31)for(let x=9;x<640;x+=61){
 const p=JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));assert.equal(p.complete,true,JSON.stringify(p));pixels++;
 for(const c of p.contributors){if(!c.command?.hasSource)continue;let pairs=[];
  if(c.command.kind==='Bitplane scanline'){
   const cmd=c.command;for(let plane=0;plane<cmd.planePointers.length;plane++){
    const hires=!!(cmd.bplcon0&0x8000),pos=(hires?c.u:Math.floor(c.u/2))-(cmd.fetchStart-0x38)*(hires?4:2)-((cmd.scroll>>(plane%2?4:0))&15),word=cmd.planeWords[plane]?.[Math.floor(pos/16)];
    if(pos>=0&&word!==undefined)pairs.push([(cmd.planePointers[plane]+Math.floor(pos/16)*2)&0x7ffff,2,c.sourceBefore,((word&255)<<8)|(word>>8)]);
   }
  }else if(c.command.kind==='Sprite scanline'){pairs.push([c.sourceAddress,4,c.sourceBefore,c.sourceValue]);sprites++;}
  pairs.push([c.paletteAddress,2,c.paletteBefore,c.paletteValue]);
  for(const args of pairs){const q=JSON.parse(core.UTF8ToString(core._rr_source(...args)));assert.equal(q.complete,true,JSON.stringify({c,args,q}));sources++;}
 }
}
const end=save(),proof=json('_rr_proof'),pixelsAt=core._rr_frame(),frame=core.HEAPU8.slice(pixelsAt,pixelsAt+640*256*4);
const replay=json('_rr_replay_begin');for(const target of [replay.count,Math.floor(replay.count/2),1,replay.count]){while(!core._rr_replay_seek(target)){}const p=core._rr_replay_frame(),bytes=core.HEAPU8.subarray(p,p+frame.length);if(target===replay.count)assert.deepEqual(bytes,frame);else if(target===1)assert.notDeepEqual(bytes,frame);}
assert.deepEqual(save(),end,'Replay changed machine');assert.deepEqual(json('_rr_proof'),proof);
restore(initial);core._rr_pad(0);for(let i=0;i<60;i++)next();const stationary=json('_rr_proof');
restore(initial);core._rr_pad(8);for(let i=0;i<60;i++)next();const moving=json('_rr_proof');assert.notEqual(moving.rgba,stationary.rgba,'Joystick movement did not change gameplay output');
console.log(JSON.stringify({controlsChangeDisplay:true,frame:json('_rr_status').frames,coreFPS:120000/ms,nativeParity:!!nativeEnd,capture:info,pixels,sources,sprites,replaySteps:replay.count,stateBytes:end.length},null,2));
