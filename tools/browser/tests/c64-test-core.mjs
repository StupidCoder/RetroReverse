// Reuse the same worker-protocol tests with the production or owned C64 core.
// This host shim hashes actual inputs and binds the owned codec's identities;
// it does not change execution, stepping, snapshots or checkpoint contents.
import fs from 'node:fs';
import {createHash} from 'node:crypto';
import {pathToFileURL} from 'node:url';
export async function createC64TestCore(){
 const url=process.env.RR_C64_CORE?pathToFileURL(process.env.RR_C64_CORE):new URL('../../../site/emulators/cores/c64/core.js',import.meta.url);
 const wasm=fs.readFileSync(new URL('./core.wasm',url)),factory=(await import(url.href)).default;
 const core=await factory({wasmBinary:wasm});
 if(!core._rr_state_bind)return core;
 const digest=b=>new Uint8Array(createHash('sha256').update(b).digest()),zero=new Uint8Array(32);
 const hashes=[digest(wasm),zero,zero,zero,zero,zero,zero];let configuration=0;
 const bind=()=>{const bytes=new Uint8Array(228);hashes.forEach((h,i)=>bytes.set(h,i*32));new DataView(bytes.buffer).setUint32(224,configuration,true);core.HEAPU8.set(bytes,core._rr_input());if(!core._rr_state_bind())throw Error('Identity binding failed');};
 const take=n=>core.HEAPU8.slice(core._rr_input(),core._rr_input()+Math.max(0,n));
 const init=core._rr_init;core._rr_init=(b,k,c)=>{const bytes=take(b+k+c),r=init(b,k,c);if(r){hashes[1]=digest(bytes.slice(0,b));hashes[2]=digest(bytes.slice(b,b+k));hashes[3]=digest(bytes.slice(b+k));hashes[4]=hashes[5]=hashes[6]=zero;configuration=0;bind();}return r;};
 const tape=core._rr_tape;core._rr_tape=n=>{const bytes=take(n),r=tape(n);if(r){hashes[5]=digest(bytes);bind();}return r;};
 const drive=core._rr_drive_rom;core._rr_drive_rom=n=>{const bytes=take(n),r=drive(n);if(r){hashes[4]=digest(bytes);configuration|=1;bind();}return r;};
 const disk=core._rr_disk;core._rr_disk=(n,protectedMedia)=>{const bytes=take(n),r=disk(n,protectedMedia);if(r){hashes[6]=digest(bytes);configuration=(configuration&~2)|(protectedMedia?0:2);bind();}return r;};
 return core;
}
