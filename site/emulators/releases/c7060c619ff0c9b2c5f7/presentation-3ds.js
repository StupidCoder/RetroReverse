// Rotate and decode the canonical scanout buffers directly into a GPU canvas.
// The resulting ImageBitmap carries the image to Play without an RGBA readback.
export async function create3DSPresentation(device){
 if(typeof OffscreenCanvas==='undefined')return null;
 const canvas=new OffscreenCanvas(400,480),context=canvas.getContext('webgpu');if(!context)return null;
 const format=navigator.gpu.getPreferredCanvasFormat();context.configure({device,format,alphaMode:'opaque'});
 const module=device.createShaderModule({code:`
 @group(0) @binding(0) var<storage,read> data:array<u32>;
 @group(0) @binding(1) var<storage,read> p:array<u32>;
 fn byte(i:u32)->u32{return (data[i/4u]>>((i%4u)*8u))&255u;}
 @vertex fn vs(@builtin(vertex_index) i:u32)->@builtin(position) vec4<f32>{let x=f32((i<<1u)&2u);let y=f32(i&2u);return vec4<f32>(x*2.-1.,y*2.-1.,0.,1.);}
 @fragment fn fs(@builtin(position) pos:vec4<f32>)->@location(0) vec4<f32>{
  var x=u32(pos.x);var y=u32(pos.y);let screen=y/240u;let b=screen*8u;y=y%240u;
  if(screen==1u){if(x<40u||x>=360u){return vec4<f32>(0.,0.,0.,1.);}x-=40u;}
  let w=p[b+1u];let h=p[b+2u];if(p[b+7u]==0u||x>=h||y>=w){return vec4<f32>(0.,0.,0.,1.);}
  let q=p[b]+(x*p[b+3u]+w-1u-y)*p[b+5u];let fmt=p[b+4u];var c=vec3<u32>(0u);
  if(fmt==0u){c=vec3<u32>(byte(q+3u),byte(q+2u),byte(q+1u));}
  if(fmt==1u){c=vec3<u32>(byte(q+2u),byte(q+1u),byte(q));}
  if(fmt>=2u){let v=byte(q)|(byte(q+1u)<<8u);
   if(fmt==2u){let r=(v>>11u)&31u;let g=(v>>5u)&63u;let blue=v&31u;c=vec3<u32>((r<<3u)|(r>>2u),(g<<2u)|(g>>4u),(blue<<3u)|(blue>>2u));}
   if(fmt==3u){let r=(v>>11u)&31u;let g=(v>>6u)&31u;let blue=(v>>1u)&31u;c=vec3<u32>((r<<3u)|(r>>2u),(g<<3u)|(g>>2u),(blue<<3u)|(blue>>2u));}
   if(fmt==4u){c=vec3<u32>((v>>12u)&15u,(v>>8u)&15u,(v>>4u)&15u)*17u;}
  }return vec4<f32>(vec3<f32>(c)/255.,1.);
 }`});
 const pipeline=await device.createRenderPipelineAsync({layout:'auto',vertex:{module,entryPoint:'vs'},fragment:{module,entryPoint:'fs',targets:[{format}]}});
 const U=GPUBufferUsage,params=device.createBuffer({size:64,usage:U.STORAGE|U.COPY_DST});let storage=null,capacity=0,bind=null;
 function present(core){
  const meta=JSON.parse(core.UTF8ToString(core._rr_graphics_scanout()));if(!meta.complete)return null;
  const offsets=[0,Math.ceil((meta.screens[0]?.bytes??0)/4)*4],n=Math.max(4,offsets[1]+(meta.screens[1]?.bytes??0));
  if(n>capacity){storage?.destroy();capacity=2**Math.ceil(Math.log2(n));storage=device.createBuffer({size:capacity,usage:U.STORAGE|U.COPY_DST});bind=device.createBindGroup({layout:pipeline.getBindGroupLayout(0),entries:[{binding:0,resource:{buffer:storage}},{binding:1,resource:{buffer:params}}]});}
  const words=new Uint32Array(16);
  for(let i=0;i<2;i++){
   const s=meta.screens[i];if(!s)continue;const bytes=new Uint8Array(Math.ceil(s.bytes/4)*4);bytes.set(core.HEAPU8.subarray(s.pointer,s.pointer+s.bytes));device.queue.writeBuffer(storage,offsets[i],bytes);
   words.set([offsets[i],s.width,s.height,s.stride,s.format,s.bpp,0,1],i*8);
  }
  device.queue.writeBuffer(params,0,words);const e=device.createCommandEncoder(),p=e.beginRenderPass({colorAttachments:[{view:context.getCurrentTexture().createView(),loadOp:'clear',storeOp:'store',clearValue:{r:0,g:0,b:0,a:1}}]});
  p.setPipeline(pipeline);p.setBindGroup(0,bind);p.draw(3);p.end();device.queue.submit([e.finish()]);return canvas.transferToImageBitmap();
 }
 return {present,get bytesAllocated(){return capacity+64;},destroy(){storage?.destroy();params.destroy();context.unconfigure();}};
}
