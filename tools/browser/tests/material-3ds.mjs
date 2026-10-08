import assert from 'node:assert/strict';
import {create3DSGraphics,transferWGSL,transferSupport} from '../../../site/emulators/graphics-3ds.js';
import {materialKey,specializeMaterialWGSL} from '../../../site/emulators/material-3ds.js';
const params=[8,8,15,0,0,3,0,0,0,1];for(let i=0;i<6;i++)params.push(0x0f0f0f,0x0f0f0f,0,0);params.push(0xffffffff,0);
const words=new Uint32Array(69).fill(0xffffffff);words[0]=64;
const packet={kind:5,params,input:new Uint8Array(words.buffer),before:new Uint8Array(256)};
assert.equal(transferSupport(packet),null);
const key=materialKey(packet),code=specializeMaterialWGSL(transferWGSL,packet);
// Runtime colors and the alpha reference must not enter the compiled identity.
for(const i of [6,7,8,...[13,17,21,25,29,33]]){
 const altered={...packet,params:[...params]};altered.params[i]=i===8?0xab00:0x89abcdef;
 assert.equal(materialKey(altered),key);assert.equal(specializeMaterialWGSL(transferWGSL,altered),code);
}
for(const i of [2,3,4,5,8,10,11,12,14,15,16,18,19,20,22,23,24,26,27,28,30,31,32]){
 const altered={...packet,params:[...params]};altered.params[i]^=i===3?256:1;assert.notEqual(materialKey(altered),key);
}
assert.throws(()=>specializeMaterialWGSL('',packet),/boundary/);
globalThis.GPUBufferUsage={STORAGE:1,COPY_DST:2,COPY_SRC:4,MAP_READ:8};globalThis.GPUMapMode={READ:1};globalThis.GPUShaderStage={COMPUTE:4};
function mock(){
 const compiled=[],selected=[],finished=[];let action=null,destroyed=false;
 const device={lost:new Promise(()=>{}),pushErrorScope(){},popErrorScope:async()=>null,
  createShaderModule:({code})=>({code}),createBindGroupLayout:()=>({}),createPipelineLayout:()=>({}),createBindGroup:()=>({}),
  createComputePipelineAsync(options){const pipeline={code:options.compute.module.code};compiled.push(pipeline);return compiled.length>10&&action?action(pipeline):Promise.resolve(pipeline);},
  createBuffer:({size})=>({mapAsync:async()=>{},getMappedRange:(off,n)=>new ArrayBuffer(n),unmap(){},destroy(){}}),
  createCommandEncoder:()=>({clearBuffer(){},beginComputePass:()=>({setPipeline:p=>selected.push(p),setBindGroup(){},dispatchWorkgroups(){},end(){}}),copyBufferToBuffer(){},finish:()=>({})}),
  queue:{writeBuffer(){},submit(){}},destroy(){destroyed=true;}};
 return {gpu:{requestAdapter:async()=>({features:new Set(),requestDevice:async()=>device})},compiled,selected,finished,
  defer(){action=p=>new Promise(resolve=>finished.push(()=>resolve(p)));},reject(){action=()=>Promise.reject(Error('Injected material compilation failure'));},resume(){action=null;},get destroyed(){return destroyed;}};
}
const tick=()=>new Promise(resolve=>setImmediate(resolve));
{
 const m=mock(),gpu=await create3DSGraphics({gpu:m.gpu,measureGPU:false,specializeLighting:false});
 m.defer();const original={...packet,params:[...params]};
 const running=gpu.execute(original);original.params[2]=7;assert(!(await running).materialSpecialized);
 await tick();assert.equal(m.compiled.at(-1).code,code,'Deferred source must use the immutable key snapshot');
 assert(!(await gpu.execute(original)).materialSpecialized);await tick();
 original.params[2]=3;assert(!(await gpu.execute(original)).materialSpecialized);
 assert.equal(gpu.materialCompilation().pending,2);assert.equal(gpu.materialCompilation().requested,2);
 m.finished.forEach(f=>f());await tick();assert.equal(gpu.materialCompilation().ready,2);
 const uniform={...packet,params:[...params]};uniform.params[6]=0xaabbccdd;uniform.params[8]=0x1200;
 assert((await gpu.execute(uniform)).materialSpecialized);assert.equal(m.compiled.length,12);
 m.reject();assert(!(await gpu.execute(original)).materialSpecialized);await tick();assert.equal(gpu.materialCompilation().failed,1);
 const calls=m.compiled.length;assert(!(await gpu.execute(original)).materialSpecialized);assert.equal(m.compiled.length,calls,'Failed keys must not compile on every draw');
 m.resume();for(let i=0;i<128;i++){const p={...packet,params:[...params]};p.params[2]=i>>3;p.params[8]=1|((i&7)<<4);assert((await gpu.execute(p)).supported);await tick();}
 assert.equal(gpu.materialCompilation().requested,64);assert.equal(gpu.materialCompilation().ready,63);assert.equal(gpu.materialCompilation().pending,0);
 const overflow={...packet,params:[...params]};overflow.params[3]=256;assert(!(await gpu.execute(overflow)).materialSpecialized);assert.equal(gpu.materialCompilation().requested,64);
 gpu.destroy();assert(m.destroyed);
}
{
 const m=mock(),gpu=await create3DSGraphics({gpu:m.gpu,measureGPU:false,specializeLighting:false});m.defer();
 assert((await gpu.execute(packet)).supported);await tick();assert.equal(gpu.materialCompilation().pending,1);
 gpu.destroy();m.finished.forEach(f=>f());await tick();assert.equal(gpu.materialCompilation().ready,0);assert.equal(gpu.materialCompilation().pending,0);
}
console.log('3DS material shaders: mode identity, runtime uniforms, immutable snapshots, bounded concurrency/cache, failed compilation and late disposal pass');
