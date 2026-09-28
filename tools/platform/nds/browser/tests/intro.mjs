// Private-media visual regression: boot -> Adventure -> first new save slot.
// Output contains game pixels/state; keep it outside the published site.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {loadCore} from '../../../../browser/tests/wasm-harness.mjs';
const [rom,out]=process.argv.slice(2);assert(rom&&out,'Usage: intro.mjs ROM OUTPUT_DIRECTORY');
fs.mkdirSync(out,{recursive:true});
const c=await loadCore('ds',rom),json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const next=()=>{const f=json('_rr_status').frames;while(json('_rr_status').frames===f)assert(c._rr_run(10000)>=0,c.UTF8ToString(c._rr_error()));};
const taps=new Map([[320,[128,120]],[420,[128,120]],[1300,[128,164]],[1500,[48,68]]]);
const releases=new Set([...taps.keys()].map(n=>n+20));
const checkpoints=new Set([1200,1400,1550,1850,2600,2800,3000,3500]);
for(let frame=1;frame<=3500;frame++){
 if(taps.has(frame))c._rr_touch(...taps.get(frame),1);
 if(releases.has(frame))c._rr_touch(0,0,0);
 next();
 if(!checkpoints.has(frame))continue;
 const p=c._rr_frame(),rgba=c.HEAPU8.subarray(p,p+256*384*4),rgb=Buffer.alloc(256*384*3);
 for(let i=0;i<256*384;i++)rgb.set(rgba.subarray(i*4,i*4+3),i*3);
 fs.writeFileSync(path.join(out,frame+'.ppm'),Buffer.concat([Buffer.from('P6\n256 384\n255\n'),rgb]));
 const n=c._rr_state_save();assert(n);fs.writeFileSync(path.join(out,frame+'.state'),c.HEAPU8.subarray(c._rr_state_data(),c._rr_state_data()+n));
 console.log(JSON.stringify({frame,...json('_rr_proof')}));
}
