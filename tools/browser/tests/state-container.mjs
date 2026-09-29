import assert from 'node:assert/strict';
import {packState,unpackState} from '../../../site/emulators/state.js';
const bytes=await packState({format:1,platform:'ps1',media:[{sha256:'example'}]},new Uint8Array([1,2,3]));
const state=await unpackState(new Blob([bytes]));assert.deepEqual([...state.payload],[1,2,3]);
for(const i of [0,8,20,bytes.length-33,bytes.length-1]){const b=bytes.slice();b[i]^=1;await assert.rejects(()=>unpackState(new Blob([b])));}
await assert.rejects(()=>unpackState(new Blob([bytes.slice(0,-1)])));
console.log('State container round trip, corruption and truncation checks passed');

for(const platform of ['gb','gg','gc','ps2']){const b=await packState({format:1,platform},new Uint8Array([4,5,6]));assert.equal((await unpackState(new Blob([b]))).meta.platform,platform);}
