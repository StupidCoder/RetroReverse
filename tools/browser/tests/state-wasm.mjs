import fs from 'node:fs';
import assert from 'node:assert/strict';
const platform=process.argv[2],path=process.argv[3];
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
const tick=()=>platform==='3do'?check(core._rr_run_slice(10000)):check(core._rr_run(10000,0,0)>=0);
const save=()=>{const n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
const restore=b=>{const p=core._rr_state_input(b.length);check(p);core.HEAPU8.set(b,p);check(core._rr_state_load(b.length));};
if(process.env.NATIVE_STATE){const b=new Uint8Array(fs.readFileSync(process.env.NATIVE_STATE));restore(b);assert.deepEqual(save(),b);console.log(JSON.stringify({platform,nativeCrossLoad:true,bytes:b.length}));}
for(let point=0;point<3;point++){
 for(let i=0;i<100;i++)tick();const a=save();for(let i=0;i<33;i++)tick();const expected=save();
 const start=performance.now();restore(a);const ms=performance.now()-start;for(let i=0;i<33;i++)tick();assert.deepEqual(save(),expected);
 const bad=a.slice(0,-1);const p=core._rr_state_input(bad.length);core.HEAPU8.set(bad,p);assert.equal(core._rr_state_load(bad.length),0);assert.deepEqual(save(),expected);
 console.log(JSON.stringify({platform,point,bytes:a.length,restoreMs:ms}));
}

if(process.env.STATE_FIXTURE){
 const {packState,digest}=await import('../../../site/emulators/state.js');
 const manifest=JSON.parse(fs.readFileSync(new URL('../../../site/emulators/build-manifest.json',import.meta.url)));
 const firmware=platform==='c64'?await Promise.all(['basic','kernal','chargen'].map(n=>digest(fs.readFileSync(new URL('../../../site/emulators/firmware/c64/'+n+'.rom',import.meta.url))))):[];
 const {InputQueue}=await import('../../../site/emulators/input.js');const q=new InputQueue(60);
 const input={...q,pulses:[],down:[],pending:[],lastButtons:-1,lastX:0,lastY:0,inputSequence:0,lastInputStep:0};
 const payload=save();fs.writeFileSync(process.env.STATE_FIXTURE,await packState({format:1,platform,media:[{name:path.split('/').at(-1),size:bytes.length,sha256:await digest(bytes)}],firmware,core:manifest[platform+'/core.wasm'],configuration:{compatibility:true,customFirmware:false},input},payload));
}
