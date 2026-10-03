// Verify the actual Asyncify commit boundary without a GPU: Reference supplies
// recorded results; corrupt responses must leave BOTH surfaces untouched before
// the original draw runs. This does not validate shader arithmetic or speed.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {loadCore} from './wasm-harness.mjs';
import {decodeGraphicsStream,transferSupport} from '../../../site/emulators/graphics-3ds.js';
const [media,state,out,fields='2']=process.argv.slice(2),core=await loadCore('3ds',media),checkpoint=fs.readFileSync(state);
const check=n=>assert(n,core.UTF8ToString(core._rr_error()));
const restore=()=>{core._rr_graphics_enable(0);const p=core._rr_state_input(checkpoint.length);core.HEAPU8.set(checkpoint,p);check(core._rr_state_load(checkpoint.length));core._rr_pad(0,0,0);core._rr_touch(0,0,0);};
const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
const save=()=>{const n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
const hash=bytes=>createHash('sha256').update(bytes).digest('hex');
const run=async()=>{const end=status().frames+Number(fields);while(status().frames<end)check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);};
restore();core._rr_graphics_begin();await run();const size=core._rr_graphics_end(),records=decodeGraphicsStream(core.HEAPU8.slice(core._rr_graphics_data(),core._rr_graphics_data()+size));
assert.equal(JSON.parse(core.UTF8ToString(core._rr_graphics_info())).dropped,0);const reference=save();
restore();await run();assert.equal(hash(save()),hash(reference),'Recording changed Reference continuation');
const submitted=records.filter(p=>p.before.length/(p.kind===7?2:1)>=4096&&!p.params.at(-1));
assert(submitted.some(p=>p.kind===7),'Checkpoint must contain depth draws');
for(const p of submitted)assert.equal(transferSupport(p),null);
const modes=['valid','short-result','missing-depth-counter','negative-depth-counter','excess-depth-counter','excess-combined-counters','fractional-depth-counter','nonfinite-depth-counter','rejected','throw'];
const results=[];
for(const mode of modes){
 restore();let index=0,depth=0;core._rr_graphics_enable(1);
 core.graphicsTransfer=async packet=>{
  const expected=submitted[index++];assert(expected,'Unexpected live packet');assert.equal(packet.kind,expected.kind);
  const params=expected.params.slice();if(packet.kind>=5)params[34]=0xffffffff;if(packet.kind===7)params[55]=0xffffffff;
  assert.deepEqual(packet.params,params);assert.deepEqual(packet.input,expected.input);assert.deepEqual(packet.before,expected.before);
  const result={supported:true,bytes:expected.expected,drawn:packet.kind>=5?expected.params[34]:0,depthKilled:packet.kind===7?expected.params[55]:0};
  if(packet.kind!==7)return result;depth++;
  await new Promise(r=>setTimeout(r,1)); // Suspend the real draw continuation.
  if(mode==='valid')return result;
  result.bytes=result.bytes.slice().fill(0xa5); // Any premature write is visible.
  if(mode==='short-result')result.bytes=result.bytes.subarray(0,result.bytes.length/2);
  if(mode==='missing-depth-counter')delete result.depthKilled;
  if(mode==='negative-depth-counter')result.depthKilled=-1;
  if(mode==='excess-depth-counter')result.depthKilled=packet.params[9]+1;
  if(mode==='excess-combined-counters'){result.drawn=packet.params[9];result.depthKilled=1;}
  if(mode==='fractional-depth-counter')result.depthKilled=.5;
  if(mode==='nonfinite-depth-counter')result.depthKilled=NaN;
  if(mode==='rejected')result.supported=false;
  if(mode==='throw')throw Error('Injected depth operation failure');
  return result;
 };
 await run();assert.equal(index,submitted.length);assert(depth>0);const actual=save();assert.equal(hash(actual),hash(reference),mode+' changed canonical state');
 results.push({mode,operations:index,depthOperations:depth,state:hash(actual)});
}
core._rr_graphics_enable(0);core.graphicsTransfer=null;
const report={schema:1,result:'PASS',validation:'Recorded Reference responses, not GPU execution',fields:Number(fields),checkpoint:hash(checkpoint),reference:hash(reference),records:records.length,results};
if(out)fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
