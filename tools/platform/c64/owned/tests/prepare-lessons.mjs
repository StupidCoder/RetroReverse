// Regenerate only recipes and digests, never distribute game/checkpoint bytes.
// Usage: RR_C64_CORE=/absolute/core.mjs node prepare-lessons.mjs firmware-dir fort.tap elite.tap [--write]
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createC64TestCore} from '../../../../browser/tests/c64-test-core.mjs';
import {prepareStart} from '../../../../../site/emulators/prepared-start.js';
import {digest} from '../../../../../site/emulators/state.js';
const [firmwareDir,fortPath,elitePath,write]=process.argv.slice(2);
assert(firmwareDir&&fortPath&&elitePath&&process.env.RR_C64_CORE);
const roms=['basic','kernal','chargen'].map(n=>fs.readFileSync(path.join(firmwareDir,n+'.rom')));
const identity={core:await digest(fs.readFileSync(path.join(path.dirname(process.env.RR_C64_CORE),'core.wasm'))),firmware:await Promise.all(roms.map(digest))};
for(const [game,tapePath] of [['fort-apocalypse-c64',fortPath],['elite-c64',elitePath]]){
 const file=new URL('../../../../../games/'+game+'/knowledge.json',import.meta.url),pkg=JSON.parse(fs.readFileSync(file)),tape=fs.readFileSync(tapePath);
 assert.equal(await digest(tape),pkg.releases['reference-pal'].media[0].sha256);
 const core=await createC64TestCore(),put=b=>core.HEAPU8.set(b,core._rr_input());
 function fresh(){put(Buffer.concat(roms));assert(core._rr_init(8192,8192,4096));put(tape);assert(core._rr_tape(tape.length));}
 const save=()=>{const n=core._rr_state_save();assert(n>0);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
 const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
 let actions=[];
 function record(a){if(a.run&&actions.at(-1)?.run)actions.at(-1).run+=a.run;else actions.push(a);}
 function run(n){record({run:n});while(n){const slice=Math.min(n,100000);assert.equal(core._rr_run(slice,0,0),slice);n-=slice;}}
 function key(k,d){core._rr_key(k,d);record({key:[k,d]});}
 function type(text){for(const c of text){key(c.charCodeAt(0),1);run(100000);key(c.charCodeAt(0),0);run(100000);}}
 function joystick(mask){core._rr_joystick(2,mask);record({joystick:[2,mask]});}
 function until(pc,budget){record({until:pc,budget});assert(core._rr_debug_begin(2,pc,1));let r=0;while(budget&&!r){const n=Math.min(10000,budget);r=core._rr_debug_run(n);budget-=n;}assert.equal(r,4);}
 const generated=[];
 async function capture(id){const bytes=save(),recipe={...pkg.preparedStarts[id],...identity,evidence:['owned-prepared'],sha256:await digest(bytes),actions:structuredClone(actions)};generated.push({id:id+'-owned',recipe,bytes});console.log(game,id,core._rr_cycle(),recipe.sha256);}
 fresh();run(3000000);type('LOAD\r');core._rr_play(1);record({play:1});
 if(game==='fort-apocalypse-c64'){
  let loaded=false;for(let i=0;i<10000;i++){run(10000);const s=status();if(s.pulse>=48233&&!s.motor){loaded=true;break;}}assert(loaded);
  run(200000);type('RUN\r');until(0x8600,150000000);assert.equal(core._rr_cycle(),115317157);
  run(3000000);joystick(16);until(0x8cdb,1000000);joystick(0);await capture('terrain');
  run(6000000);until(0x9c52,1000000);
  // The 110th subsequent native AI invocation supplies the SID-noise phase
  // needed by the unchanged experiment's teleport and spawn assertions.
  run(2162162);assert.equal(core._rr_cycle(),126553519);await capture('enemy-ai');
 }else{until(0x378,50000000);assert.equal(core._rr_cycle(),37906392);await capture('loader-prefix');}
 for(const {id,recipe,bytes} of generated){
  fresh();const prepared=await prepareStart({core,recipe,identity});assert(!prepared.cached);assert.equal(await digest(prepared.bytes),recipe.sha256);
  core._rr_run(50,0,0);const cached=await prepareStart({core,recipe,identity,cached:bytes.buffer});assert(cached.cached);
  if(write==='--write')pkg.preparedStarts[id]=recipe;else assert.deepEqual(pkg.preparedStarts[id],recipe,'Regenerate reviewed owned recipe with --write');
 }
 if(write==='--write'){
  pkg.evidence['owned-prepared']={status:'confirmed',source:game==='fort-apocalypse-c64'?'prepared':'acceptance',description:'Owned C64 core: exact image, firmware and WASM identities; authentic keyboard boot, fresh replay and cached checkpoint hash verified.',limitations:game==='fort-apocalypse-c64'?'Terrain tour and the existing guarded 200-frame AI experiment pass. The opcode repair remains a scenario-specific hypothesis.':'Four-stop first-byte tour and 52 independently decoded vector stores pass. Later loader stages, full gameplay and object slots remain unvalidated.'};
  fs.writeFileSync(file,JSON.stringify(pkg,null,2).replace(/[\u007f-\uffff]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'))+'\n');
 }
}
console.log('PASS owned prepared lessons: verified physical-input recipes, complete state digests, fresh generation and cached restore');
