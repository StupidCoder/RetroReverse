// Same benchmark in Node and a real browser; no media/state bytes in reports.
import {prepareStart} from '../../../site/emulators/prepared-start.js';
import {bindOwnedIdentity} from '../../../site/emulators/core-backend.js';
import {digest} from '../../../site/emulators/state.js';
export async function benchmark({factory,wasm,roms,tape,pkg,label}){
 const core=await factory({wasmBinary:wasm}),put=b=>core.HEAPU8.set(b,core._rr_input());
 const firmware=new Uint8Array(20480);let offset=0;for(const rom of roms){firmware.set(rom,offset);offset+=rom.length;}put(firmware);
 if(!core._rr_init(8192,8192,4096))throw Error('init');put(tape);if(!core._rr_tape(tape.length))throw Error('tape');
 const identity={core:await digest(wasm),firmware:await Promise.all(roms.map(digest))};
 if(core._rr_state_bind)bindOwnedIdentity(core,{...identity,tape:await digest(tape)});
 const recipe=Object.entries(pkg.preparedStarts).find(([id,r])=>id.startsWith('terrain')&&r.core===identity.core);
 if(!recipe)throw Error('No matching verified gameplay recipe');
 await prepareStart({core,recipe:recipe[1],identity});core._rr_trace(0,0,0);core._rr_run(40000,7,0);if(!core._rr_checkpoint(0))throw Error('checkpoint');
 const samples={gameplay:[],capture:[]};
 for(const mode of Object.keys(samples))for(let i=0;i<6;i++){
  if(!core._rr_restore(0))throw Error('restore');
  const begin=performance.now(),before=Number(core._rr_cycle());
  if(mode==='capture'){core._rr_capture_begin();core._rr_run(40000,7,0);core._rr_capture_end();}
  else if(core._rr_run(985248,0,0)!==985248)throw Error('short gameplay run');
  const ms=performance.now()-begin,cycles=Number(core._rr_cycle())-before;
  if(i)samples[mode].push({ms,cycles,realtime:cycles/985248/(ms/1000)});
 }
 return {label,core:identity.core,recipe:recipe[0],samples,median:Object.fromEntries(Object.entries(samples).map(([mode,s])=>[mode,[...s].sort((a,b)=>a.ms-b.ms)[2]]))};
}
