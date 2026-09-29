import fs from 'node:fs';
import nodePath from 'node:path';
import {selectDreamcastMedia} from '../../../site/emulators/dc-media.js';
import {pathToFileURL} from 'node:url';
import assert from 'node:assert/strict';
export async function loadCore(platform,path){
const base=process.env.CORE_DIR?pathToFileURL(process.env.CORE_DIR.replace(/\/$/,'')+'/'):new URL('../../../site/emulators/cores/'+platform+'/',import.meta.url);
const factory=(await import(new URL('core.js',base))).default;
const core=await factory({wasmBinary:fs.readFileSync(new URL('core.wasm',base))});
let tracks;
if(platform==='dc'){const cue=fs.readFileSync(path,'utf8');const match=/(?:DATAFILE|FILE)\s+"([^"]+)"/.exec(cue);assert(match);const bin=nodePath.join(nodePath.dirname(path),match[1]);const dfd=fs.openSync(bin,'r');const local={name:match[1],size:fs.statSync(bin).size,slice:(a,b)=>({arrayBuffer:async()=>{const out=Buffer.alloc(b-a);assert.equal(fs.readSync(dfd,out,0,out.length,a),out.length);return out.buffer.slice(out.byteOffset,out.byteOffset+out.byteLength)}})};const selected=await selectDreamcastMedia([{name:nodePath.basename(path),size:Buffer.byteLength(cue),text:async()=>cue},local]);tracks=selected.tracks;fs.closeSync(dfd);path=bin;}
const random=['dc','ps1','psp','gc','ps2','3do'].includes(platform);
const fileSize=fs.statSync(path).size,fd=random?fs.openSync(path,'r'):null;
const bytes=random?null:fs.readFileSync(path);
globalThis.FileReaderSync=class{readAsArrayBuffer(b){return b.buffer.slice(b.byteOffset,b.byteOffset+b.byteLength);}};
core.discFile={size:fileSize,slice:(a,b)=>{if(!random)return bytes.subarray(a,b);const out=Buffer.alloc(Math.max(0,Math.min(b,fileSize)-a));fs.readSync(fd,out,0,out.length,a);return out;}};
const check=x=>assert.ok(x,core.UTF8ToString(core._rr_error()));
if(platform==='dc'){for(const t of tracks)check(core._rr_disc_track(t.number,t.lba,t.offset,t.length,+t.data));check(core._rr_init_file(fileSize));}
else if(platform==='amiga'){
 const rom=fs.readFileSync(process.env.KICKSTART||new URL('../../../site/emulators/firmware/amiga/kick12.rom',import.meta.url));
 let p=core._rr_firmware(rom.length);check(p);core.HEAPU8.set(rom,p);p=core._rr_input(bytes.length);check(p);core.HEAPU8.set(bytes,p);check(core._rr_init(bytes.length));
}
else if(['ps1','psp','gc','ps2'].includes(platform)){check(core._rr_init_file(fileSize)>0);if(platform==='ps2'){if(process.env.PS2_BIOS){const bios=fs.readFileSync(process.env.PS2_BIOS);const p=core._rr_input(bios.length);core.HEAPU8.set(bios,p);check(core._rr_bios(bios.length));}check(core._rr_boot());}}
else if(platform==='3do')check(core._rr_init_config(fileSize,1));
else if((platform==='n64'||platform==='ds'||platform==='3ds'||platform==='gb'||platform==='gg'||platform==='gba')){const p=core._rr_input(bytes.length);core.HEAPU8.set(bytes,p);check(core._rr_init(bytes.length));}
else{let p=core._rr_input();for(const name of ['basic','kernal','chargen']){const b=fs.readFileSync(new URL('../../../site/emulators/firmware/c64/'+name+'.rom',import.meta.url));core.HEAPU8.set(b,p);p+=b.length;}check(core._rr_init(8192,8192,4096));core.HEAPU8.set(bytes,core._rr_input());check(core._rr_tape(bytes.length));}
return core;
}
