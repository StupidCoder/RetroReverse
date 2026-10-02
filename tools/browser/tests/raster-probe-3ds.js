// Diagnostic only: coverage is necessary but insufficient for a live draw path.
// Deliberately includes pixel-centre edges where PICA reference uses >= 0 for
// every edge, unlike fixed-function rasterization's top-left rule.
export async function rasterProbe(){
 const a=await navigator.gpu.requestAdapter(),d=await a.requestDevice(),U=GPUBufferUsage;
 const width=32,height=32,triangles=[[[.5,.5],[24.5,.5],[.5,24.5]],[[.1,.1],[30.3,.2],[.25,30.25]],[[7.5,3.5],[28.5,15.5],[4.5,27.5]]];
 const results=[],f=Math.fround,edge=(a,b,x,y)=>f(f(f(b[0]-a[0])*f(y-a[1]))-f(f(b[1]-a[1])*f(x-a[0])));
 d.pushErrorScope('validation');
 const compute=d.createShaderModule({code:`
 @group(0) @binding(0) var<storage,read> v:array<vec2<f32>>;
 @group(0) @binding(1) var<storage,read_write> out:array<u32>;
 fn edge(a:vec2<f32>,b:vec2<f32>,p:vec2<f32>)->f32{return (b.x-a.x)*(p.y-a.y)-(b.y-a.y)*(p.x-a.x);}
 @compute @workgroup_size(8,8) fn main(@builtin(global_invocation_id) id:vec3<u32>){let p=vec2<f32>(id.xy)+vec2<f32>(.5);out[id.y*32u+id.x]=select(0u,1u,edge(v[1],v[2],p)>=0&&edge(v[2],v[0],p)>=0&&edge(v[0],v[1],p)>=0);}`});
 const cp=await d.createComputePipelineAsync({layout:'auto',compute:{module:compute,entryPoint:'main'}});
 const render=d.createShaderModule({code:`
 @vertex fn vs(@location(0) p:vec2<f32>)->@builtin(position) vec4<f32>{return vec4<f32>(p.x/16.-1.,1.-p.y/16.,0.,1.);}
 @fragment fn fs()->@location(0) vec4<f32>{return vec4<f32>(1.);}`});
 const rp=await d.createRenderPipelineAsync({layout:'auto',vertex:{module:render,entryPoint:'vs',buffers:[{arrayStride:8,attributes:[{shaderLocation:0,offset:0,format:'float32x2'}]}]},fragment:{module:render,entryPoint:'fs',targets:[{format:'rgba8unorm'}]},primitive:{topology:'triangle-list'}});
 try{for(const t of triangles){
  const verts=new Float32Array(t.flat()),vb=d.createBuffer({size:24,usage:U.STORAGE|U.VERTEX|U.COPY_DST}),ob=d.createBuffer({size:4096,usage:U.STORAGE|U.COPY_SRC}),read=d.createBuffer({size:4096,usage:U.COPY_DST|U.MAP_READ}),rr=d.createBuffer({size:8192,usage:U.COPY_DST|U.MAP_READ});
  const tex=d.createTexture({size:[width,height],format:'rgba8unorm',usage:GPUTextureUsage.RENDER_ATTACHMENT|GPUTextureUsage.COPY_SRC});d.queue.writeBuffer(vb,0,verts);
  const enc=d.createCommandEncoder(),c=enc.beginComputePass();c.setPipeline(cp);c.setBindGroup(0,d.createBindGroup({layout:cp.getBindGroupLayout(0),entries:[{binding:0,resource:{buffer:vb}},{binding:1,resource:{buffer:ob}}]}));c.dispatchWorkgroups(4,4);c.end();enc.copyBufferToBuffer(ob,0,read,0,4096);
  const r=enc.beginRenderPass({colorAttachments:[{view:tex.createView(),clearValue:{r:0,g:0,b:0,a:0},loadOp:'clear',storeOp:'store'}]});r.setPipeline(rp);r.setVertexBuffer(0,vb);r.draw(3);r.end();enc.copyTextureToBuffer({texture:tex},{buffer:rr,bytesPerRow:256},[width,height]);d.queue.submit([enc.finish()]);await Promise.all([read.mapAsync(GPUMapMode.READ),rr.mapAsync(GPUMapMode.READ)]);
  const cb=new Uint32Array(read.getMappedRange()),rb=new Uint8Array(rr.getMappedRange());let computeDiff=0,rasterDiff=0;
  const tv=Array.from({length:3},(_,i)=>[verts[i*2],verts[i*2+1]]);
  for(let y=0;y<height;y++)for(let x=0;x<width;x++){const expected=+(edge(tv[1],tv[2],x+.5,y+.5)>=0&&edge(tv[2],tv[0],x+.5,y+.5)>=0&&edge(tv[0],tv[1],x+.5,y+.5)>=0);computeDiff+=cb[y*width+x]!==expected;rasterDiff+=+(rb[y*256+x*4]>0)!==expected;}
  results.push({vertices:t,computeCoverageDifferences:computeDiff,conventionalCoverageDifferences:rasterDiff});read.unmap();rr.unmap();for(const b of [vb,ob,read,rr])b.destroy();tex.destroy();
 }
 const error=await d.popErrorScope();if(error)throw Error(error.message);return results;
 }finally{d.destroy();}
}
