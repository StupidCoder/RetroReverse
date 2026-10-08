import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage();await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-3ds.html');
 const result=await page.evaluate(async()=>{
  const {rasterFloatWGSL}=await import('/site/emulators/raster-3ds.js');
  const adapter=await navigator.gpu.requestAdapter(),device=await adapter.requestDevice();
  // Every significand and exponent parity in [1,4). All normal inputs reduce
  // to one of these; the ordinary arithmetic test also covers subnormal scaling.
  const count=1<<20,U=GPUBufferUsage;
  const output=device.createBuffer({size:count*4,usage:U.STORAGE|U.COPY_SRC}),read=device.createBuffer({size:count*4,usage:U.COPY_DST|U.MAP_READ}),offset=device.createBuffer({size:4,usage:U.STORAGE|U.COPY_DST});
  try{
   const module=device.createShaderModule({code:rasterFloatWGSL+`@group(0) @binding(0) var<storage,read_write> output:array<u32>;@group(0) @binding(1) var<storage,read> offset:u32;
   @compute @workgroup_size(64) fn main(@builtin(global_invocation_id) id:vec3<u32>){output[id.x]=rsqrt(offset+id.x);}`});
   const pipeline=await device.createComputePipelineAsync({layout:'auto',compute:{module,entryPoint:'main'}}),bind=device.createBindGroup({layout:pipeline.getBindGroupLayout(0),entries:[{binding:0,resource:{buffer:output}},{binding:1,resource:{buffer:offset}}]});
   const word=new Uint32Array(1),float=new Float32Array(word.buffer);let tested=0,mismatches=0;const failures=[];
   for(let start=0x3f800000;start<0x40800000;start+=count){
    device.queue.writeBuffer(offset,0,new Uint32Array([start]));const encoder=device.createCommandEncoder(),pass=encoder.beginComputePass();pass.setPipeline(pipeline);pass.setBindGroup(0,bind);pass.dispatchWorkgroups(count/64);pass.end();encoder.copyBufferToBuffer(output,0,read,0,count*4);device.queue.submit([encoder.finish()]);await read.mapAsync(GPUMapMode.READ);
    const actual=new Uint32Array(read.getMappedRange());
    for(let i=0;i<count;i++){word[0]=start+i;float[0]=Math.sqrt(float[0]);if(actual[i]!==word[0]){mismatches++;if(failures.length<10)failures.push({input:start+i,expected:word[0],actual:actual[i]});}}
    read.unmap();tested+=count;
   }
   return {tested,mismatches,failures,adapter:{vendor:adapter.info.vendor,architecture:adapter.info.architecture}};
  }finally{device.destroy();}
 });
 assert.equal(result.mismatches,0,JSON.stringify(result));result.result='PASS';result.browser=await browser.version();if(process.argv[2])fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
}finally{await browser.close();}
