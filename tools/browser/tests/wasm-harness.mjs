import fs from 'node:fs';
import assert from 'node:assert/strict';
export async function loadCore(platform,path){
const base=new URL('../../../site/emulators/cores/'+platform+'/',import.meta.url);
const factory=(await import(new URL('core.js',base))).default;
const core=await factory({wasmBinary:fs.readFileSync(new URL('core.wasm',base))});
const bytes=fs.readFileSync(path);
globalThis.FileReaderSync=class{readAsArrayBuffer(b){return b.buffer.slice(b.byteOffset,b.byteOffset+b.byteLength);}};
core.discFile={slice:(a,b)=>bytes.subarray(a,b)};
const check=x=>assert.ok(x,core.UTF8ToString(core._rr_error()));
if(platform==='ps1')check(core._rr_init_file(bytes.length)>0);
else if(platform==='3do')check(core._rr_init_config(bytes.length,1));
else if(platform==='n64'){const p=core._rr_input(bytes.length);core.HEAPU8.set(bytes,p);check(core._rr_init(bytes.length));}
else{let p=core._rr_input();for(const name of ['basic','kernal','chargen']){const b=fs.readFileSync(new URL('../../../site/emulators/firmware/c64/'+name+'.rom',import.meta.url));core.HEAPU8.set(b,p);p+=b.length;}check(core._rr_init(8192,8192,4096));core.HEAPU8.set(bytes,core._rr_input());check(core._rr_tape(bytes.length));}
return core;
}
