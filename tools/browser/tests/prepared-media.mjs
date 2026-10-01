// Private media: never writes checkpoint bytes into the repository.
// node tools/browser/tests/prepared-media.mjs <game-slug> <image>
import fs from 'node:fs';import assert from 'node:assert/strict';
import factory from '../../../site/emulators/cores/c64/core.js';
import {prepareStart} from '../../../site/emulators/prepared-start.js';
import {digest} from '../../../site/emulators/state.js';
const read=p=>fs.readFileSync(new URL('../../../'+p,import.meta.url)),pkg=JSON.parse(read('games/'+process.argv[2]+'/knowledge.json')),tape=fs.readFileSync(process.argv[3]),wasm=read('site/emulators/cores/c64/core.wasm'),roms=['basic','kernal','chargen'].map(n=>read('site/emulators/firmware/c64/'+n+'.rom'));
const identity={core:await digest(wasm),firmware:await Promise.all(roms.map(digest))};
async function fresh(){const c=await factory({wasmBinary:wasm}),put=b=>c.HEAPU8.set(b,c._rr_input());put(Buffer.concat(roms));assert(c._rr_init(8192,8192,4096));put(tape);assert(c._rr_tape(tape.length));c._rr_trace(0,0,0);return c;}
const reports=[];for(const [id,recipe]of Object.entries(pkg.preparedStarts).filter(([,r])=>r.core===identity.core)){assert.equal(await digest(tape),pkg.releases[recipe.releases[0]].media[0].sha256);let first;
 for(let i=0;i<3;i++){const core=await fresh(),r=await prepareStart({core,recipe,identity,cached:i===2?first.bytes:null});assert.equal(r.cached,i===2);assert.equal(await digest(r.bytes),recipe.sha256);first??=r;reports.push({id,mode:i===2?'cache':'recipe',cycle:String(core._rr_cycle()),sha256:await digest(r.bytes)});}
}
console.log(JSON.stringify({result:'PASS',game:pkg.id,reports},null,2));
