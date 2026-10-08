import assert from 'node:assert/strict';
import {pathToFileURL} from 'node:url';
import {compilePicaShader,createPicaShaders} from '../../../site/emulators/shader-3ds.js';
const {default:factory}=await import(pathToFileURL(process.argv[2]||'tools/platform/n3ds/browser/work/shader-oracle.mjs'));
const core=await factory(),ptr=Array.from({length:7},(_,i)=>core._shader_memory(i)),memory=core.testMemory;
let seed=0x3d502024,passed=0,refused=0;
const random=()=>{seed^=seed<<13;seed^=seed>>>17;seed^=seed<<5;return seed>>>0;};
const identity=(15|0x1b<<5|0x1b<<14|0x1b<<23)>>>0;
const instruction=(op,dst=0,a=0,b=1,index=0,desc=0)=>(op<<26|dst<<21|index<<19|a<<12|b<<7|desc)>>>0;
const mov=(dst,src,index=0,desc=0)=>instruction(0x13,dst,src,0,index,desc),end=0x22<<26;
const flow=(op,dst,num=0,condition=0)=>(op<<26|condition<<22|dst<<10|num)>>>0;
const u32=(slot,n)=>new Uint32Array(memory.buffer,ptr[slot],n);
const f32=(slot,n)=>new Float32Array(memory.buffer,ptr[slot],n);
function fillInputs(special=false){
 for(const [slot,n]of [[2,16*64],[5,96*4]])for(let i=0;i<n;i++){
  f32(slot,n)[i]=((random()%65536)-32768)/512;
  if(special&&i%5===0)u32(slot,n)[i]=[0,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc00123,0x7fa12345,0x4f000000,0xcf000000][random()%10];
 }
 new Uint8Array(memory.buffer,ptr[6],16).set([2,1,1,0,3,2,1,0,1,0,0,0,0,1,0,0]);
}
async function check(words,descs=[identity],{special=false,bool=random()&65535,entry=0,count=16,required=false}={}){
 const code=new Uint32Array(4096),desc=new Uint32Array(128);code.set(words);desc.set(descs);u32(0,4096).set(code);u32(1,128).set(desc);fillInputs(special);
 assert.equal(core._shader_reference(entry,count,bool),1,'Reference shader must succeed');
 const expected=u32(3,count*64).slice(),binary=compilePicaShader(code,desc,entry);
 const {instance}=await WebAssembly.instantiate(binary,{env:{memory}});
 u32(4,count*64).fill(0xdeadbeef);const ok=instance.exports.run(ptr[2],ptr[4],count,ptr[5],ptr[6],bool);
 if(ok){assert.deepEqual(u32(4,count*64),expected,`bitwise mismatch: ${words.map(v=>(v>>>0).toString(16))}`);passed++;}
 else{assert(!required,'Finite fixture unexpectedly refused');refused++;}
 return {binary,instance,code,desc};
}
// Every arithmetic operation, swizzle, write mask, source/destination alias and
// regular/inverse operand layout; execute on actual WASM Reference arithmetic.
for(const op of [0,1,2,3,8,9,10,11,12,13,14,15,18,19,24,26,27])for(let trial=0;trial<12;trial++){
 const desc=((random()&0xfffffff0)|(trial===0?15:random()&15))>>>0;
 const instructionWord=[24,26,27].includes(op)?(op<<26|16<<21|1<<19|16<<14|32<<7)>>>0:instruction(op,16,trial%2?16:32,17,trial%4);
 await check([mov(16,0,0,1),mov(17,1,0,1),instructionWord,mov(0,16,0,1),mov(1,33,1,1),mov(2,34,2,1),end],[desc,identity],{special:trial>=6,required:trial<6});
}
for(let trial=0;trial<48;trial++){
 const cmp=(0x2e<<26|(trial%8)<<24|((trial>>2)%8)<<21|32<<12|16<<7)>>>0;
 const mad=((trial%2?0x38:0x30)<<26|17<<24|16<<17|(trial%2?32<<10:16<<12)|(trial%2?16:32)<<5)>>>0;
 await check([mov(16,0,0,1),mad,mov(0,17,0,1),cmp,flow(0x28,7,1,trial%16),mov(1,1,0,1),mov(2,2,0,1),mov(3,3,0,1),end],[random(),identity],{special:trial>=24});
}
// The END inside a subroutine/IF returns only from that range. LOOP has inclusive
// counts and nested loops replace aL; uniforms and bools stay live, never baked.
const structured=[flow(0x26,12,3,0),flow(0x29,5,0,0),instruction(0,16,32,16,3),flow(0x29,4,0,1),instruction(0,16,33,16,3),mov(0,16),flow(0x27,9,2,1),mov(1,34,1),end,mov(2,35,2),end,end,mov(16,1),end,mov(16,2)];
for(let i=0;i<20;i++)await check(structured,[identity],{bool:i,required:true});
for(let i=0;i<16;i++)await check([mov(16,0),0x2e<<26|2<<24|5<<21|32<<12|16<<7,flow(0x2c,5,0,i),flow(0x2d,6,i&1,i),mov(0,1),mov(1,2),mov(2,3),end],[identity],{required:true});
// Random straight-line programs catch interactions across partial register writes.
for(let trial=0;trial<100;trial++){
 const words=Array.from({length:16},(_,r)=>mov(16+r,r,0,1)),desc=[random(),identity];
 for(let i=0;i<24;i++)words.push(instruction([0,1,2,3,8,9,10,12,13,19][random()%10],16+random()%16,random()%128,random()%32,random()%4,random()%2));
 for(let r=0;r<16;r++)words.push(mov(r,r+16,0,1));words.push(end);await check(words,desc,{special:trial>60,required:trial<=60});
}
// Boundary conversions, signed zero, subnormals, infinities and NaN payloads.
for(const op of [11,12,13,14,15,18,19])await check([instruction(op),end],[identity],{special:true});
const grown=await check([mov(0,0),end],[identity],{required:true});memory.grow(1);fillInputs();assert.equal(grown.instance.exports.run(ptr[2],ptr[4],16,ptr[5],ptr[6],0),1);assert.deepEqual(u32(4,4),u32(2,4));
for(const words of [[0x1c<<26,end],[flow(0x24,0,1),end],[flow(0x2c,0),end],[flow(0x29,3),flow(0x29,3),flow(0x29,3),0x21<<26,end]]){
 const code=new Uint32Array(4096);code.set(words);assert.throws(()=>compilePicaShader(code,new Uint32Array(128),0));
}
const manager=createPicaShaders();manager.prepare(1,grown.code,grown.desc,0,memory);assert.equal(manager.ready(1),false);
const until=async predicate=>{for(let i=0;i<100&&!predicate();i++)await new Promise(resolve=>setTimeout(resolve,10));assert(predicate(),'Asynchronous compiler did not settle');};
await until(()=>manager.ready(1));assert.equal(manager.run(1,ptr[2],ptr[4],1,ptr[5],ptr[6],0),1);manager.dispose();assert.equal(manager.run(1,ptr[2],ptr[4],1,ptr[5],ptr[6],0),0);
const immutable=createPicaShaders(),mutable=grown.code.slice();immutable.prepare(1,mutable,grown.desc,0,memory);mutable[0]=0x1c<<26;
await until(()=>immutable.ready(1));assert.equal(immutable.run(1,ptr[2],ptr[4],1,ptr[5],ptr[6],0),1);immutable.dispose();
const cancelled=createPicaShaders();cancelled.prepare(1,grown.code,grown.desc,0,memory);cancelled.dispose();
const unsupported=createPicaShaders();unsupported.prepare(1,mutable,grown.desc,0,memory);await until(()=>unsupported.snapshot().unsupported===1);unsupported.dispose();
assert.equal(cancelled.ready(1),false);assert.equal(cancelled.snapshot().ready,0);
assert(passed>200&&refused>10);console.log(JSON.stringify({test:'PICA compiled WASM vs WASM interpreter',passed,refused,flow:true,memoryGrowth:true,lifecycle:true}));
