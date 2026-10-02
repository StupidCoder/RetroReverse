import assert from 'node:assert/strict';
import {create3DSExecution} from '../../../site/emulators/execution-3ds.js';
import {createExecutionGate} from '../../../site/emulators/execution-gate.js';
const defer=()=>{let resolve;const promise=new Promise(r=>resolve=r);return {promise,resolve};};
const absent=create3DSExecution();
assert.equal(absent.snapshot().effective,'reference');
await assert.rejects(absent.select('experimental'),/not available/);
await assert.rejects(absent.select('__proto__'),/Unknown/);
assert.equal(absent.snapshot().busy,false);
const calls=[],events=[],gate=createExecutionGate();let wait=defer();
const modes=create3DSExecution({
  reference:{quiesce:async()=>{calls.push('quiesce reference');await wait.promise;},activate:()=>calls.push('activate reference')},
  experimental:{available:true,label:'Test reference backend',prepare:()=>calls.push('prepare'),quiesce:()=>calls.push('quiesce experimental'),activate:()=>calls.push('activate experimental')},
  onChange:s=>events.push(s)
});
const ownership=gate.acquire('execution-mode');
let pending=modes.select('experimental');await Promise.resolve();
assert.equal(modes.snapshot().effective,'reference');
assert.equal(modes.snapshot().requested,'experimental');
assert.equal(gate.acquire('save'),null);
await assert.rejects(modes.select('reference'),/pending/);
wait.resolve();await pending;gate.release(ownership);
assert.equal(modes.snapshot().effective,'experimental');
await modes.select('reference');
assert.deepEqual(calls,['prepare','quiesce reference','activate experimental','quiesce experimental','activate reference']);
wait=defer();pending=modes.select('experimental');await Promise.resolve();
modes.cancel();assert.equal(modes.snapshot().busy,true); // Keep ownership until host work settles.
wait.resolve();await pending;
assert.equal(modes.snapshot().effective,'reference');
assert.equal(modes.snapshot().busy,false);
assert.equal(calls.filter(c=>c==='activate experimental').length,1);
wait=defer();pending=modes.select('experimental');await Promise.resolve();
modes.dispose();const count=events.length;wait.resolve();await pending;
assert.equal(events.length,count);await assert.rejects(modes.select('reference'),/ended/);
const failed=create3DSExecution({experimental:{available:true,prepare(){throw Error('shader failed');}}});
await assert.rejects(failed.select('experimental'),/shader failed/);
assert.equal(failed.snapshot().effective,'reference');assert.equal(failed.snapshot().busy,false);
console.log('3DS transitions: availability, ordering, gate ownership, failure, cancellation and stale-session checks passed');
