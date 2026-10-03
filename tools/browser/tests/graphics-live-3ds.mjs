import assert from 'node:assert/strict';
import {createLive3DSGraphics} from '../../../site/emulators/graphics-live-3ds.js';
const core=()=>({_rr_graphics_enable(n){this.enabled=n;},_rr_graphics_stats:()=>'{"operations":[0,0,0,0],"accelerated":[0,0,0,0]}',UTF8ToString:s=>s});
const selfTest=()=>({supported:true,bytes:Uint8Array.from({length:4096},(_,i)=>[0x78,0x56,0x34,0x12][i%4])});
{
 const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>{throw Error('No adapter');}});
 assert(!b.experimental.available);assert.match(b.reason,/No adapter/);assert.throws(()=>b.experimental.activate());b.reference.activate();assert.equal(c.enabled,0);b.dispose();
}
{
 let resolve,calls=0,destroyed=0;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>({execute(){return ++calls===1?selfTest():new Promise(r=>resolve=r);},destroy(){destroyed++;}})});
 b.experimental.activate();const pending=c.graphicsTransfer({});await assert.rejects(c.graphicsTransfer({}),/in flight/);
 let quiet=false;const q=b.experimental.quiesce().then(()=>quiet=true);await Promise.resolve();assert(!quiet);resolve({supported:true,bytes:new Uint8Array(4)});await pending;await q;assert(quiet);b.reference.activate();assert.equal(c.enabled,0);b.dispose();assert.equal(destroyed,1);
}
{
 let calls=0,destroyed=0;const c=core(),b=await createLive3DSGraphics(c,{timeoutMs:5,createGPU:async()=>({execute:()=>++calls===1?selfTest():new Promise(()=>{}),destroy(){destroyed++;}})});
 b.experimental.activate();const result=await c.graphicsTransfer({});assert.equal(result.supported,false);assert(b.failed);assert(!b.experimental.available);assert.match(b.reason,/timed out/);assert.equal(destroyed,1);b.dispose();
}
{
 let lose;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async({onLost})=>{lose=onLost;return {execute:selfTest,destroy(){}};}});
 b.experimental.activate();lose('Lost during a frame');assert(!b.experimental.available);assert.equal((await c.graphicsTransfer({})).supported,false);b.reference.activate();assert.equal(c.enabled,0);b.dispose();
}
console.log('3DS live bridge: adapter failure, exclusive operation, quiescence, timeout and device loss pass');
{
 let resolve,destroyed=0;const c=core(),b=await createLive3DSGraphics(c,{timeoutMs:5,createGPU:()=>new Promise(r=>resolve=r)});
 assert(!b.experimental.available);assert.match(b.reason,/initialization timed out/);
 resolve({destroy(){destroyed++;}});await new Promise(r=>setTimeout(r,0));assert.equal(destroyed,1);b.dispose();
}
{
 const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>({execute:()=>({supported:true,bytes:new Uint8Array()}),destroy(){}})});
 assert(!b.experimental.available);assert.match(b.reason,/self-test failed/);b.dispose();
}
{
 let calls=0;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>({execute(){if(++calls===1)return selfTest();throw Error('Synchronous submission failed');},destroy(){}})});
 b.experimental.activate();assert.equal((await c.graphicsTransfer({})).supported,false);assert(b.failed);b.dispose();
}
console.log('3DS initialization deadline, late resource disposal, self-test length and synchronous failure pass');
{
 const {create3DSGraphics}=await import('../../../site/emulators/graphics-3ds.js');
 for(const gpu of [null,{requestAdapter:async()=>null}]){
  const b=await createLive3DSGraphics(core(),{createGPU:options=>create3DSGraphics({...options,gpu})});assert(!b.experimental.available);assert.match(b.reason,/unavailable|No WebGPU adapter/);b.dispose();
 }
 let destroyed=0;
 globalThis.GPUShaderStage={COMPUTE:4};
 const device={createBindGroupLayout:()=>({}),createPipelineLayout:()=>({}),lost:new Promise(()=>{}),pushErrorScope(){},createShaderModule:()=>({getCompilationInfo:async()=>({messages:[{lineNum:1,linePos:1,message:'Injected invalid shader'}]})}),createComputePipelineAsync:async()=>{throw Error('Compile failed');},destroy(){destroyed++;}};
 const b=await createLive3DSGraphics(core(),{createGPU:options=>create3DSGraphics({...options,gpu:{requestAdapter:async()=>({features:new Set(),requestDevice:async()=>device})}})});
 assert(!b.experimental.available);assert.match(b.reason,/Injected invalid shader/);assert.equal(destroyed,1);b.dispose();
}
console.log('3DS missing WebGPU, absent adapter and shader compilation failure keep Reference available');
{
 let calls=0;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>({info:{timestamps:true},execute:()=>++calls===1?selfTest():{supported:true,bytes:new Uint8Array(4),timing:{uploadMs:1,submitMs:2,queueAndMapMs:3,readbackMs:4,gpuMs:0}},destroy(){}})});
 b.experimental.activate();await c.graphicsTransfer({});const first=b.stats();assert.deepEqual(first.timing,{operations:1,uploadMs:1,submitMs:2,queueAndMapMs:3,readbackMs:4,gpuMs:0,gpuSamples:1});
 await c.graphicsTransfer({});assert.equal(first.timing.operations,1);assert.equal(b.stats().timing.operations,2);assert(b.stats().timestamps);b.dispose();
}
// A validation promise that resolves only once mapping starts would deadlock the
// old serial wait. Exercise both rejection orders and reuse after cleanup.
{
 const {create3DSGraphics}=await import('../../../site/emulators/graphics-3ds.js');
 globalThis.GPUBufferUsage={STORAGE:1,COPY_DST:2,COPY_SRC:4,MAP_READ:8};globalThis.GPUMapMode={READ:1};globalThis.GPUShaderStage={COMPUTE:4};
 let pops=0,resolveValidation,mode='success',unmaps=0;const variants=[];
 const device={lost:new Promise(()=>{}),pushErrorScope(){},popErrorScope(){if(++pops===1)return Promise.resolve(null);return new Promise(r=>resolveValidation=r);},
  createShaderModule:()=>({}),createBindGroupLayout:()=>({}),createPipelineLayout:()=>({}),createComputePipelineAsync:async options=>{variants.push(options.compute.constants.operation);return {};},createBindGroup:()=>({}),
  createBuffer({size}){const bytes=new ArrayBuffer(size);return {mapAsync(){assert(resolveValidation,'Validation was not requested before mapping');resolveValidation(mode==='validation-error'?{message:'Injected validation error'}:null);resolveValidation=null;return mode==='map-error'?Promise.reject(Error('Injected mapping error')):Promise.resolve();},getMappedRange(offset,length){return bytes.slice(offset,offset+length);},unmap(){unmaps++;},destroy(){}};},
  createCommandEncoder:()=>({clearBuffer(){},beginComputePass:()=>({setPipeline(){},setBindGroup(){},dispatchWorkgroups(){},end(){}}),copyBufferToBuffer(){},finish:()=>({})}),queue:{writeBuffer(){},submit(){}},destroy(){}};
 const gpu=await create3DSGraphics({measureGPU:false,gpu:{requestAdapter:async()=>({features:new Set(),requestDevice:async()=>device})}});
 assert.deepEqual(variants,[1,2,3,4,5,6,7]);
 const packet={kind:1,params:[4,0,4096,0],input:new Uint8Array(),before:new Uint8Array(4096)};
 assert((await gpu.execute(packet)).supported);assert(!gpu.busy);
 for(mode of ['validation-error','map-error']){await assert.rejects(gpu.execute(packet),/Injected/);assert(!gpu.busy);}
 mode='success';assert((await gpu.execute(packet)).supported);assert.equal(unmaps,4);gpu.destroy();
}
console.log('3DS parallel validation/readback, failure cleanup, pipeline variants and timing snapshots pass');
