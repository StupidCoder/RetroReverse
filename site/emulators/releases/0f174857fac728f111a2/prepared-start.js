import {digest} from './state.js';
// This profile only drives ordinary C64 inputs and bounded execution. No RAM edits.
export async function prepareStart({core,recipe,identity,cached,progress=()=>{},sleep=()=>Promise.resolve()}){
 if(identity.core!==recipe.core||JSON.stringify(identity.firmware)!==JSON.stringify(recipe.firmware))throw Error('This prepared start requires its verified core and firmware versions.');
 const save=()=>{const n=core._rr_state_save();if(n<=0||n>4*1024*1024)throw Error('Prepared checkpoint exceeds 4 MiB');return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
 const verify=()=>{const s=JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));if(!s.boundary||s.interruptPending||s.nextPC!==recipe.pc)throw Error('Prepared start did not reach its verified instruction boundary.');};
 if(cached instanceof ArrayBuffer&&cached.byteLength<=4*1024*1024&&await digest(cached)===recipe.sha256){const p=core._rr_state_input(cached.byteLength);if(!p)throw Error('Checkpoint allocation failed');core.HEAPU8.set(new Uint8Array(cached),p);if(!core._rr_state_load(cached.byteLength))throw Error('Cached checkpoint restore failed');verify();return {bytes:cached,cached:true};}
 let cycles=0,last=0;const began=performance.now();
 async function yieldProgress(){if(performance.now()-began>120000)throw Error('Preparation exceeded its two-minute budget');if(performance.now()-last>40){progress(cycles);last=performance.now();await sleep(0);}}
 for(const a of recipe.actions){
  if(a.run)for(let left=a.run;left;){const n=Math.min(left,100000);if(core._rr_run(n)<0)throw Error('Preparation execution failed');left-=n;cycles+=n;await yieldProgress();}
  if(a.key)core._rr_key(...a.key);
  if(a.joystick)core._rr_joystick(...a.joystick);
  if(a.play!==undefined)core._rr_play(a.play);
  if(a.until!==undefined){if(!core._rr_debug_begin(2,a.until,1))throw Error('Preparation target rejected');let r=0;for(let left=a.budget;left&&!r;){const n=Math.min(left,1000),before=core._rr_cycle();r=core._rr_debug_run(n);cycles+=core._rr_cycle()-before;left-=n;await yieldProgress();}if(r!==4)throw Error('Preparation did not reach the required code address');}
 }
 verify();const bytes=save();if(await digest(bytes)!==recipe.sha256)throw Error('Prepared state differs from the verified checkpoint. Current session retained.');return {bytes:bytes.buffer,cached:false};
}
