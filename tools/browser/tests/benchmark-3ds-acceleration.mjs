// Private media and checkpoints are inputs; output contains measurements/hashes only.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {loadCore} from './wasm-harness.mjs';
import {unpackState} from '../../../site/emulators/state.js';
import {create3DSExecution} from '../../../site/emulators/execution-3ds.js';
const [media,statePath,label='scene',countArg='30',out]=process.argv.slice(2);
assert(media&&statePath,'Usage: benchmark-3ds-acceleration.mjs MEDIA STATE_OR_- LABEL [COUNT] [OUT]');
const count=Number(countArg);assert(Number.isInteger(count)&&count>0&&count<=10000);
const c=await loadCore('3ds',media),json=n=>JSON.parse(c.UTF8ToString(c[n]()));
const check=v=>assert(v,c.UTF8ToString(c._rr_error()));
const save=()=>{const n=c._rr_state_save();check(n);return c.HEAPU8.slice(c._rr_state_data(),c._rr_state_data()+n);};
const restore=b=>{const p=c._rr_state_input(b.length);check(p);c.HEAPU8.set(b,p);check(c._rr_state_load(b.length));};
let initial=statePath==='-'?save():fs.readFileSync(statePath);
if(Buffer.from(initial.subarray(0,8)).toString()==='RRSTATE1')initial=(await unpackState(new Blob([initial]))).payload;
restore(initial);initial=save();
const hash=b=>createHash('sha256').update(b).digest('hex');
function next(){const f=json('_rr_status').frames;let longest=0,calls=0;while(json('_rr_status').frames===f){const t=performance.now();check(c._rr_run(10000)>=0);longest=Math.max(longest,performance.now()-t);assert(++calls<100000);}return longest;}
const trials=[];
for(let trial=0;trial<3;trial++){
 restore(initial);for(let i=0;i<3;i++)next();restore(initial);
 const start=json('_rr_status'),before=json('_rr_profile'),times=[];let longest=0;
 for(let i=0;i<count;i++){const t=performance.now();longest=Math.max(longest,next());times.push(performance.now()-t);}
 const after=json('_rr_profile'),sorted=[...times].sort((a,b)=>a-b),total=times.reduce((a,b)=>a+b,0);
 trials.push({intervals:count,totalMs:total,intervalsPerSecond:count*1000/total,medianMs:sorted[Math.floor(sorted.length/2)],p95Ms:sorted[Math.ceil(sorted.length*.95)-1],longestCallMs:longest,
  start,end:json('_rr_status'),proof:json('_rr_proof'),stateSha256:hash(save()),heapBytes:c.HEAPU8.length,
  profile:after.buckets.map(b=>({name:b.name,ms:b.ms-(before.buckets.find(a=>a.name===b.name)?.ms||0)}))});
}
assert(trials.every(r=>r.stateSha256===trials[0].stateSha256),'Restored runs diverged');
// M1 uses a test-only second reference backend. Mode changes themselves must
// preserve every serialized field, pending event and current input latch.
const engine=create3DSExecution({experimental:{available:true,label:'Test reference backend'}});
function input(i){c._rr_pad(i%4===0?1:0,(i%3-1)*20,0);c._rr_touch(80+i,100,i%5===0?1:0);}
restore(initial);for(let i=0;i<32;i++){input(i);check(c._rr_run(10000)>=0);}const expected=save();
restore(initial);
for(let i=0;i<32;i++){
 input(i);const before=i%8===0?hash(save()):null;
 await engine.select(i%2?'reference':'experimental');
 if(before)assert.equal(hash(save()),before,'Transition mutated the machine');
 check(c._rr_run(10000)>=0);
}
assert.equal(hash(save()),hash(expected),'Switched continuation diverged');
const result={schema:1,label,backend:'browser-reference',method:'Node WASM, 3 restored trials, 3 warmup intervals, execution/status timing; hashing and serialization excluded. Display intervals are not presented images.',checkpointSha256:hash(initial),trials,transitionCheck:{switches:32,canonicalStateMatch:true,inputContinuationMatch:true}};
const text=JSON.stringify(result,null,2)+'\n';if(out)fs.writeFileSync(out,text);else console.log(text);
console.error(label+': '+trials.map(r=>r.intervalsPerSecond.toFixed(2)).join(', ')+' intervals/s; transitions match');
