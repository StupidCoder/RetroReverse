import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage();await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-3ds.html');
 const result=await page.evaluate(async()=>{
  const {rasterFloatWGSL}=await import('/site/emulators/raster-3ds.js');
  const adapter=await navigator.gpu.requestAdapter(),device=await adapter.requestDevice();
  const bits=new Uint32Array(1),floats=new Float32Array(bits.buffer),f=x=>{bits[0]=x;return floats[0];},u=x=>{floats[0]=x;return bits[0];};
  const pairs=[],edges=[0,0x80000000,1,2,3,0x007fffff,0x00800000,0x00800001,0x3f000000,0x3f7fffff,0x3f800000,0x3f800001,0x40000000,0x437f0000,0x7f7fffff,0x7f800000];
  for(const a of [...edges,...edges.map(x=>(x|0x80000000)>>>0)])for(const b of [...edges,...edges.map(x=>(x|0x80000000)>>>0)])pairs.push([a,b]);
  let seed=0x3d600001;const random=()=>{seed^=seed<<13;seed^=seed>>>17;seed^=seed<<5;return seed>>>0;};
  for(let i=0;i<65536;i++)pairs.push([random(),random()]);
  // Cancellation and halfway rounding, not just uniformly random exponents.
  for(let i=0;i<16384;i++){const a=random()&0x7f7fffff;pairs.push([a,(a^0x80000000)>>>0],[a,(a+1)^0x80000000],[a,0x33800000],[a,0x3f800001]);}
  const input=new Uint32Array(pairs.flat()),U=GPUBufferUsage;
  const source=device.createBuffer({size:input.byteLength,usage:U.STORAGE|U.COPY_DST}),out=device.createBuffer({size:pairs.length*12,usage:U.STORAGE|U.COPY_SRC}),read=device.createBuffer({size:pairs.length*12,usage:U.COPY_DST|U.MAP_READ});
  const module=device.createShaderModule({code:rasterFloatWGSL+`\n@group(0) @binding(0) var<storage,read> inputs:array<vec2<u32>>;@group(0) @binding(1) var<storage,read_write> outputs:array<u32>;
  @compute @workgroup_size(64) fn main(@builtin(global_invocation_id) id:vec3<u32>){let i=id.x;if(i>=arrayLength(&inputs)){return;}let a=inputs[i].x;let b=inputs[i].y;outputs[i*3u]=radd(a,b);outputs[i*3u+1u]=rmul(a,b);outputs[i*3u+2u]=rdiv(a,b);}`});
  const pipeline=await device.createComputePipelineAsync({layout:'auto',compute:{module,entryPoint:'main'}}),bind=device.createBindGroup({layout:pipeline.getBindGroupLayout(0),entries:[{binding:0,resource:{buffer:source}},{binding:1,resource:{buffer:out}}]});
  device.queue.writeBuffer(source,0,input);const encoder=device.createCommandEncoder(),pass=encoder.beginComputePass();pass.setPipeline(pipeline);pass.setBindGroup(0,bind);pass.dispatchWorkgroups(Math.ceil(pairs.length/64));pass.end();encoder.copyBufferToBuffer(out,0,read,0,pairs.length*12);device.queue.submit([encoder.finish()]);await read.mapAsync(GPUMapMode.READ);
  const actual=new Uint32Array(read.getMappedRange()),failures=[];let mismatches=0;
  for(let i=0;i<pairs.length;i++){
   const a=f(pairs[i][0]),b=f(pairs[i][1]),expected=[u(a+b),u(a*b),u(a/b)];
   for(let op=0;op<3;op++){const e=expected[op],v=actual[i*3+op];if(v!==e&&!((v&0x7fffffff)>0x7f800000&&(e&0x7fffffff)>0x7f800000)){mismatches++;if(failures.length<20)failures.push({a:pairs[i][0]>>>0,b:pairs[i][1]>>>0,op,expected:e,actual:v});}}
  }
  read.unmap();device.destroy();return {pairs:pairs.length,operations:pairs.length*3,mismatches,failures};
 });
 console.log(JSON.stringify(result));assert.equal(result.mismatches,0);if(process.argv[2])fs.writeFileSync(process.argv[2],JSON.stringify({...result,result:'PASS',browser:await browser.version()},null,2)+'\n');
}finally{await browser.close();}
