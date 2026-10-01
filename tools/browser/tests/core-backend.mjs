import assert from 'node:assert/strict';
import {backendAssets,selectedC64Backend,bindOwnedIdentity,matchingPreparedKnowledge} from '../../../site/emulators/core-backend.js';
assert.equal(backendAssets('c64').module,'./cores/c64/core.js');
assert.equal(backendAssets('c64','owned').manifest,'./cores/c64-owned/manifest.json');
assert.equal(selectedC64Backend('?c64Core=owned'),'owned');assert.equal(selectedC64Backend(''),'production');
assert.throws(()=>selectedC64Backend('?c64Core=https://example.com/core.js'));
assert.throws(()=>backendAssets('dos','owned'));
let calls=0;const core={HEAPU8:new Uint8Array(512),_rr_input:()=>16,_rr_state_bind:()=>++calls};
const identity={core:'01'.repeat(32),firmware:['02','03','04'].map(s=>s.repeat(32)),tape:'05'.repeat(32)};
bindOwnedIdentity(core,identity);assert.equal(calls,1);
assert.deepEqual([...core.HEAPU8.slice(16,244)],[...Array(32).fill(1),...Array(32).fill(2),...Array(32).fill(3),...Array(32).fill(4),...Array(32).fill(0),...Array(32).fill(5),...Array(36).fill(0)]);
assert.throws(()=>bindOwnedIdentity(core,{...identity,tape:undefined}));assert.equal(calls,1);
const recipe={core:identity.core,firmware:identity.firmware},knowledge={status:'matched',data:{functions:{entry:1},preparedStarts:{matching:recipe,old:{...recipe,core:'ff'.repeat(32)}}}};
const filtered=matchingPreparedKnowledge(knowledge,identity);
assert.deepEqual(Object.keys(filtered.data.preparedStarts),['matching']);assert.equal(filtered.data.functions,knowledge.data.functions);assert.equal(Object.keys(knowledge.data.preparedStarts).length,2);
console.log('PASS backend selection, verified identity binding and nonmutating prepared-lesson filtering');

bindOwnedIdentity(core,{...identity,tape:null,driveRom:'06'.repeat(32),disk:'07'.repeat(32)});assert.equal(core.HEAPU8[16+128],6);assert.equal(core.HEAPU8[16+192],7);assert.equal(core.HEAPU8[16+224],1);assert.throws(()=>bindOwnedIdentity(core,{...identity,tape:null,disk:'07'.repeat(32)}));
